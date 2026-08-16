package middleware

import (
	"encoding/json"
	"net/http"
)

func RequireRole(
	allowedRoles ...string,
) func(http.HandlerFunc) http.HandlerFunc {
	allowed := make(map[string]struct{})

	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(
		next http.HandlerFunc,
	) http.HandlerFunc {
		return func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			principal, ok :=
				PrincipalFromContext(
					r.Context(),
				)

			if !ok {
				writeRBACError(
					w,
					http.StatusUnauthorized,
					"authenticated principal unavailable",
				)
				return
			}

			_, isAllowed :=
				allowed[principal.Role]

			if !isAllowed {
				writeRBACError(
					w,
					http.StatusForbidden,
					"insufficient permissions",
				)
				return
			}

			next(w, r)
		}
	}
}

func writeRBACError(
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
