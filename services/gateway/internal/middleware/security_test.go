package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(
					http.StatusOK,
				)
			},
		),
	)

	request :=
		httptest.NewRequest(
			http.MethodGet,
			"/health",
			nil,
		)

	recorder :=
		httptest.NewRecorder()

	handler.ServeHTTP(
		recorder,
		request,
	)

	expected := map[string]string{
		"X-Content-Type-Options": "nosniff",

		"X-Frame-Options": "DENY",

		"Referrer-Policy": "no-referrer",

		"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'",

		"Permissions-Policy": "camera=(), microphone=(), geolocation=()",

		"Cache-Control": "no-store",
	}

	for header, expectedValue := range expected {

		actual :=
			recorder.Header().Get(
				header,
			)

		if actual != expectedValue {
			t.Fatalf(
				"%s expected %q, got %q",
				header,
				expectedValue,
				actual,
			)
		}
	}
}
