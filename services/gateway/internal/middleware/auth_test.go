package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyAuthAdmin(t *testing.T) {
	auth := NewAPIKeyAuth(
		"admin-secret",
		"analyst-secret",
		"broker-secret",
	)

	handler := auth.Required(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			principal, ok :=
				PrincipalFromContext(
					r.Context(),
				)

			if !ok {
				t.Fatal(
					"principal missing from context",
				)
			}

			if principal.ID !=
				"security-admin" {
				t.Fatalf(
					"unexpected principal ID: %s",
					principal.ID,
				)
			}

			if principal.Role !=
				"admin" {
				t.Fatalf(
					"unexpected role: %s",
					principal.Role,
				)
			}

			w.WriteHeader(
				http.StatusOK,
			)
		},
	)

	request :=
		httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

	request.Header.Set(
		"Authorization",
		"Bearer admin-secret",
	)

	recorder :=
		httptest.NewRecorder()

	handler(
		recorder,
		request,
	)

	if recorder.Code !=
		http.StatusOK {

		t.Fatalf(
			"expected 200, got %d",
			recorder.Code,
		)
	}
}

func TestAPIKeyAuthAnalyst(t *testing.T) {
	auth := NewAPIKeyAuth(
		"admin-secret",
		"analyst-secret",
		"broker-secret",
	)

	handler := auth.Required(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			principal, ok :=
				PrincipalFromContext(
					r.Context(),
				)

			if !ok {
				t.Fatal(
					"principal missing from context",
				)
			}

			if principal.Role !=
				"analyst" {

				t.Fatalf(
					"expected analyst role, got %s",
					principal.Role,
				)
			}

			w.WriteHeader(
				http.StatusOK,
			)
		},
	)

	request :=
		httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

	request.Header.Set(
		"Authorization",
		"Bearer analyst-secret",
	)

	recorder :=
		httptest.NewRecorder()

	handler(
		recorder,
		request,
	)

	if recorder.Code !=
		http.StatusOK {

		t.Fatalf(
			"expected 200, got %d",
			recorder.Code,
		)
	}
}

func TestAPIKeyAuthService(t *testing.T) {
	auth := NewAPIKeyAuth(
		"admin-secret",
		"analyst-secret",
		"broker-secret",
	)

	handler := auth.Required(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			principal, ok :=
				PrincipalFromContext(
					r.Context(),
				)

			if !ok {
				t.Fatal(
					"principal missing from context",
				)
			}

			if principal.ID !=
				"credential-broker" {

				t.Fatalf(
					"unexpected service identity: %s",
					principal.ID,
				)
			}

			if principal.Role !=
				"service" {

				t.Fatalf(
					"expected service role, got %s",
					principal.Role,
				)
			}

			w.WriteHeader(
				http.StatusOK,
			)
		},
	)

	request :=
		httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

	request.Header.Set(
		"Authorization",
		"Bearer broker-secret",
	)

	recorder :=
		httptest.NewRecorder()

	handler(
		recorder,
		request,
	)

	if recorder.Code !=
		http.StatusOK {

		t.Fatalf(
			"expected 200, got %d",
			recorder.Code,
		)
	}
}

func TestAPIKeyAuthRejectsMissingCredential(
	t *testing.T,
) {
	auth := NewAPIKeyAuth(
		"admin-secret",
		"analyst-secret",
		"broker-secret",
	)

	handler := auth.Required(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(
				http.StatusOK,
			)
		},
	)

	request :=
		httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

	recorder :=
		httptest.NewRecorder()

	handler(
		recorder,
		request,
	)

	if recorder.Code !=
		http.StatusUnauthorized {

		t.Fatalf(
			"expected 401, got %d",
			recorder.Code,
		)
	}
}

func TestAPIKeyAuthRejectsInvalidCredential(
	t *testing.T,
) {
	auth := NewAPIKeyAuth(
		"admin-secret",
		"analyst-secret",
		"broker-secret",
	)

	handler := auth.Required(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(
				http.StatusOK,
			)
		},
	)

	request :=
		httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

	request.Header.Set(
		"Authorization",
		"Bearer wrong-secret",
	)

	recorder :=
		httptest.NewRecorder()

	handler(
		recorder,
		request,
	)

	if recorder.Code !=
		http.StatusUnauthorized {

		t.Fatalf(
			"expected 401, got %d",
			recorder.Code,
		)
	}
}
