package refresh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"test-task/internal/service/auth"
	"test-task/internal/handler"
	"test-task/internal/models"
)

const (
	op = "Handler. Refresh"
)

type Request struct {
	RefreshToken string `json:"refresh_token"`
}

type Response struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthRefresher interface {
    Refresh(ctx context.Context, t models.UpdateTokens) (models.Tokens, error)
}

func New(log *slog.Logger, authRefresher AuthRefresher) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info(fmt.Sprintf("%s: start", op))
        
		t, err := handleRequest(r)
		if err!= nil {
			msg := "internal error"
			if errors.Is(err, handler.ErrInvalidCredentials){
				msg = "invalid credentials"
				log.Error(fmt.Sprintf("%s: %v", op, err))
                http.Error(w, msg, http.StatusBadRequest)
                return
			}
            log.Error(fmt.Sprintf("%s: %v", op, err))
			http.Error(w, msg, http.StatusInternalServerError)
            return
        }
        token, err := authRefresher.Refresh(r.Context(), t)
        if err!= nil {
			msg := "internal error"
            if errors.Is(err, auth.ErrUserNotFound){
				msg = "user not found"
				log.Error(fmt.Sprintf("%s: %v", op, err))
				http.Error(w, msg, http.StatusBadRequest)
				return
			}else if errors.Is(err, auth.ErrInvalidCredentials){
				msg = "invalid credentials"
				log.Error(fmt.Sprintf("%s: %v", op, err))
                http.Error(w, msg, http.StatusBadRequest)
				return
			}
            log.Error(fmt.Sprintf("%s: %v", op, err))
			http.Error(w, msg, http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(Response{
			AccessToken: token.AccessToken,
			RefreshToken: token.RefreshToken,
		}); err != nil {
			msg := "internal error"
			log.Error(fmt.Sprintf("%s: %v", op, err))
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		log.Info(fmt.Sprintf("%s: successfully", op))
    }
}

func handleRequest(r *http.Request) (models.UpdateTokens, error){
	var req Request
    if err := json.NewDecoder(r.Body).Decode(&req); err!= nil {
        return models.UpdateTokens{}, err
    }
    if req.RefreshToken == ""{
        return models.UpdateTokens{}, handler.ErrInvalidCredentials
    }
    return models.UpdateTokens{
		RefreshToken: req.RefreshToken,
	}, nil
}