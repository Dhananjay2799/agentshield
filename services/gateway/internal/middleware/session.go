package middleware

import (
	"context"
	"net/http"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

type contextKey string

const sessionContextKey contextKey = "agentshield-session"

func SessionRequired(repo *repository.SessionRepository, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-AgentShield-Session")

		if sessionID == "" {
			http.Error(w, `{"error":"missing AgentShield session"}`, http.StatusUnauthorized)
			return
		}

		session, err := repo.ValidateActive(r.Context(), sessionID)
		if err != nil {
			http.Error(w, `{"error":"invalid or expired AgentShield session"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), sessionContextKey, session)

		next(w, r.WithContext(ctx))
	}
}

func SessionFromContext(ctx context.Context) (*models.AgentSession, bool) {
	session, ok := ctx.Value(sessionContextKey).(*models.AgentSession)
	return session, ok
}
