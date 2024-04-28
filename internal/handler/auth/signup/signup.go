package signup

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
    op = "Handler. Signup"
)

type Request struct {
	GUID string `json:"guid"`
}

type Response struct {
	IsSuccess bool `json:"is_success"`
}

type AuthSignUpper interface {
	SignUp(ctx context.Context, t models.GetTokens)  error
}

func New(log *slog.Logger, authSignUpper AuthSignUpper) http.HandlerFunc{
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

        if err := authSignUpper.SignUp(r.Context(), t); err!= nil {
			msg := "internal error"
			if errors.Is(err, auth.ErrUserAlreadyExists){
				msg = "user already exists"
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
			IsSuccess: true,
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
	var t models.GetTokens
    if err := json.NewDecoder(r.Body).Decode(&t); err!= nil {
        return models.GetTokens{}, err
    }
    if t.GUID == "" || !(utils.IsGuid(t.GUID)) {
        return models.GetTokens{}, handler.ErrInvalidCredentials
    }
    return t, nil
}