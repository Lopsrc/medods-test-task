package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	// "os/user"

	"test-task/internal/config"
	"test-task/internal/models"
	"test-task/internal/storage"
	tokenmanager "test-task/pkg/auth"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInternalError 	  = errors.New("internal error")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthRepository interface{
	Update(ctx context.Context, user models.User) error
	Insert(ctx context.Context, GUID string) error
	Get(ctx context.Context, GUID string) ([]byte, error)
	GetAll(ctx context.Context) ([]models.User, error)
}

type Auth struct {
	r 			 AuthRepository
	log 		 *slog.Logger
	tokenManager tokenmanager.Manager
	tokenTTL 	 config.AuthConfig
}

func New(log *slog.Logger, r AuthRepository, tokenManager tokenmanager.Manager, ttl config.AuthConfig) *Auth {
	return &Auth{
        r: r,
        log: log,
		tokenManager: tokenManager,
		tokenTTL: ttl,
    }
}
// SignIn creates a tokens and writes it to the database.
// If user is not exist, returns error.
func (a *Auth) SignIn(ctx context.Context, t models.GetTokens) (models.Tokens, error) {
	op := "Auth. SignIn"

	tokens, err := a.createSession(t.GUID)
	if err != nil {
		a.log.Error(fmt.Sprintf("%s: %v", op, err))
		return models.Tokens{}, ErrInternalError
	}
	
	refreshHash, err := bcrypt.GenerateFromPassword([]byte(tokens.RefreshToken), bcrypt.DefaultCost)
	if err!= nil {
		a.log.Error(fmt.Sprintf("%s: %v", op, err))
        return models.Tokens{}, ErrInternalError
    }

	if err := a.r.Update(ctx, models.User{
		GUID: t.GUID,
        RefreshTokenHash: refreshHash,
        ExpiresAt: time.Now().Add(a.tokenTTL.RefreshTokenTTL),
	}); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			a.log.Error("User not found")
			return models.Tokens{}, ErrUserNotFound
		}
		a.log.Error(fmt.Sprintf("%s: %v", op, err))
		return models.Tokens{}, ErrInternalError
	}
    return tokens, nil
}
// SignUp creates a user with the given GUID.
// If user already exist, returns error.
func (a *Auth) SignUp(ctx context.Context, t models.GetTokens) error{
	op := "Auth. SignUp"

	_, err := a.r.Get(ctx, t.GUID)
	if err == nil {
		a.log.Error(fmt.Sprintf("%s: %v", op, ErrUserAlreadyExists))
		return ErrUserAlreadyExists
	}
	
	if err := a.r.Insert(ctx, t.GUID); err != nil {
		a.log.Error(fmt.Sprintf("%s: %v", op, err))
        return ErrInternalError
    }
	return nil
}
// SignUp updates refresh and access tokens.
// Get all users from the database.
// If user is not found, returns error. (Compare hashes with refresh tokens)
// Otherwise, update tokens and write the refresh token to the database, after return pair tokens.
func (a *Auth) Refresh(ctx context.Context, t models.UpdateTokens) (models.Tokens, error) {
	op := "Auth. Refresh"

	users, err := a.r.GetAll(ctx)
	if err != nil {
		a.log.Error(fmt.Sprintf("%s: %v", op, err))
        return models.Tokens{}, ErrInternalError	
	}

	user, err := a.compareHashesAndToken(users, t.RefreshToken)
	if err != nil {
		a.log.Error(fmt.Sprintf("%s: %v", op, err))
		return models.Tokens{}, err
	}

	if err := a.tokenManager.ValidateRefreshToken(user.ExpiresAt); err != nil {
		a.log.Error(fmt.Sprintf("%s: %v", op, err))
        return models.Tokens{}, err
    }

	tokens, err := a.createSession(user.GUID)
	if err != nil {
		a.log.Error(fmt.Sprintf("%s: %v", op, err))
		return models.Tokens{}, ErrInternalError
	}

	refreshHash, err := bcrypt.GenerateFromPassword([]byte(tokens.RefreshToken), bcrypt.DefaultCost)
	if err!= nil {
		a.log.Error(fmt.Sprintf("%s: %v", op, err))
        return models.Tokens{}, ErrInternalError
    }

	if err := a.r.Update(ctx, models.User{
		GUID: user.GUID,
        RefreshTokenHash: refreshHash,
        ExpiresAt: time.Now().Add(a.tokenTTL.RefreshTokenTTL),
	}); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
            a.log.Error("User not found")
            return models.Tokens{}, ErrUserNotFound
        }
		a.log.Error(fmt.Sprintf("%s: %v", op, err))
        return models.Tokens{}, ErrInternalError
    }
	
	return tokens, nil
}
// createSession creates a new tokens and returns it.
func (a *Auth) createSession(guid string) (t models.Tokens, err error) {
	t.AccessToken, err = a.tokenManager.NewJWT(guid, a.tokenTTL.AccessTokenTTL)
	if err != nil {
		return 
	}

	t.RefreshToken, err = a.tokenManager.NewRefreshToken()
	if err != nil {
		return 
	}
	return 
}
// compareHashAndTokens compares hashes from all users with the refresh token.
func (a *Auth) compareHashesAndToken(users []models.User, refreshToken string) (models.User, error) {
	for _, user := range users {
		if err := bcrypt.CompareHashAndPassword(user.RefreshTokenHash, []byte(refreshToken)); err == nil {
            return user, nil
        }
	}
	return models.User{}, ErrInvalidCredentials
}