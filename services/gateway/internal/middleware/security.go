package middleware

import "net/http"

// SecurityHeaders adds defensive HTTP response headers
// to AgentShield Gateway responses.
func SecurityHeaders(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.Header().Set(
				"X-Content-Type-Options",
				"nosniff",
			)

			w.Header().Set(
				"X-Frame-Options",
				"DENY",
			)

			w.Header().Set(
				"Referrer-Policy",
				"no-referrer",
			)

			w.Header().Set(
				"Content-Security-Policy",
				"default-src 'none'; frame-ancestors 'none'",
			)

			w.Header().Set(
				"Permissions-Policy",
				"camera=(), microphone=(), geolocation=()",
			)

			w.Header().Set(
				"Cache-Control",
				"no-store",
			)

			next.ServeHTTP(
				w,
				r,
			)
		},
	)
}
