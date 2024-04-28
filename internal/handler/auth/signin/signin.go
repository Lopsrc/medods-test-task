package signin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"test-task/internal/handler"
	"test-task/internal/models"
	"test-task/internal/service/auth"
	"test-task/pkg/utils"
)

const (
    op = "Handler. SignIn"
)

type Request struct {
    GUID string `json:"guid"`
}

type Response struct {
	AccessToken string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}

type AuthSignIner interface {
	SignIn(ctx context.Context, t models.GetTokens) (models.Tokens, error)
}

func New(log *slog.Logger, authSignIner AuthSignIner) http.HandlerFunc{
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

        token, err := authSignIner.SignIn(r.Context(), t)
        if err!= nil {
            msg := "internal error"
            if errors.Is(err, auth.ErrUserNotFound){
                msg = "user not found"
				log.Error(fmt.Sprintf("%s: %v", op, err))
				http.Error(w, msg, http.StatusBadRequest)
                return
			}
            log.Error(fmt.Sprintf("%s: %v", op, err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

func handleRequest(r *http.Request) (models.GetTokens, error){
	var req Request
    if err := json.NewDecoder(r.Body).Decode(&req); err!= nil {
        return models.GetTokens{}, err
    }
    if req.GUID == "" || !(utils.IsGuid(req.GUID)) {
        return models.GetTokens{}, handler.ErrInvalidCredentials
    }
    return models.GetTokens{
        GUID: req.GUID,
    }, nil
}