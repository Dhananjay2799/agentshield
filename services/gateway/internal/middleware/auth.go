package middleware

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

type Principal struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

type authContextKey string

const principalContextKey authContextKey = "agentshield-principal"

type APIKeyAuth struct {
	AdminKey            string
	AnalystKey          string
	CredentialBrokerKey string
}

func NewAPIKeyAuth(
	adminKey string,
	analystKey string,
	credentialBrokerKey string,
) *APIKeyAuth {
	return &APIKeyAuth{
		AdminKey:            adminKey,
		AnalystKey:          analystKey,
		CredentialBrokerKey: credentialBrokerKey,
	}
}

func (a *APIKeyAuth) Required(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		authorization := strings.TrimSpace(
			r.Header.Get("Authorization"),
		)

		if authorization == "" {
			writeAuthError(
				w,
				http.StatusUnauthorized,
				"missing authorization credentials",
			)
			return
		}

		const prefix = "Bearer "

		if !strings.HasPrefix(
			authorization,
			prefix,
		) {
			writeAuthError(
				w,
				http.StatusUnauthorized,
				"authorization header must use Bearer authentication",
			)
			return
		}

		token := strings.TrimSpace(
			strings.TrimPrefix(
				authorization,
				prefix,
			),
		)

		if token == "" {
			writeAuthError(
				w,
				http.StatusUnauthorized,
				"missing bearer token",
			)
			return
		}

		principal, ok :=
			a.authenticate(token)

		if !ok {
			writeAuthError(
				w,
				http.StatusUnauthorized,
				"invalid authorization credentials",
			)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			principalContextKey,
			principal,
		)

		next(
			w,
			r.WithContext(ctx),
		)
	}
}

func (a *APIKeyAuth) authenticate(
	token string,
) (*Principal, bool) {
	if secureCompare(
		token,
		a.AdminKey,
	) {
		return &Principal{
			ID:   "security-admin",
			Role: "admin",
		}, true
	}

	if secureCompare(
		token,
		a.AnalystKey,
	) {
		return &Principal{
			ID:   "soc-analyst",
			Role: "analyst",
		}, true
	}

	if secureCompare(
		token,
		a.CredentialBrokerKey,
	) {
		return &Principal{
			ID:   "credential-broker",
			Role: "service",
		}, true
	}

	return nil, false
}

func PrincipalFromContext(
	ctx context.Context,
) (*Principal, bool) {
	principal, ok :=
		ctx.Value(
			principalContextKey,
		).(*Principal)

	return principal, ok
}

func secureCompare(
	provided string,
	expected string,
) bool {
	if expected == "" {
		return false
	}

	if len(provided) !=
		len(expected) {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(provided),
		[]byte(expected),
	) == 1
}

func writeAuthError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(
		w,
	).Encode(
		map[string]string{
			"error": message,
		},
	)
}
