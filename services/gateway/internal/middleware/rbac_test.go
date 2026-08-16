package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func requestWithPrincipal(
	role string,
) *http.Request {
	request :=
		httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

	principal :=
		&Principal{
			ID:   "test-principal",
			Role: role,
		}

	ctx :=
		context.WithValue(
			request.Context(),
			principalContextKey,
			principal,
		)

	return request.WithContext(ctx)
}

func TestRequireRoleAllowsAnalyst(
	t *testing.T,
) {
	handler :=
		RequireRole(
			"analyst",
			"admin",
		)(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(
					http.StatusOK,
				)
			},
		)

	recorder :=
		httptest.NewRecorder()

	handler(
		recorder,
		requestWithPrincipal(
			"analyst",
		),
	)

	if recorder.Code !=
		http.StatusOK {

		t.Fatalf(
			"expected 200, got %d",
			recorder.Code,
		)
	}
}

func TestRequireRoleAllowsAdmin(
	t *testing.T,
) {
	handler :=
		RequireRole(
			"analyst",
			"admin",
		)(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(
					http.StatusOK,
				)
			},
		)

	recorder :=
		httptest.NewRecorder()

	handler(
		recorder,
		requestWithPrincipal(
			"admin",
		),
	)

	if recorder.Code !=
		http.StatusOK {

		t.Fatalf(
			"expected 200, got %d",
			recorder.Code,
		)
	}
}

func TestRequireRoleBlocksAnalystFromAdminOnly(
	t *testing.T,
) {
	handler :=
		RequireRole(
			"admin",
		)(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(
					http.StatusOK,
				)
			},
		)

	recorder :=
		httptest.NewRecorder()

	handler(
		recorder,
		requestWithPrincipal(
			"analyst",
		),
	)

	if recorder.Code !=
		http.StatusForbidden {

		t.Fatalf(
			"expected 403, got %d",
			recorder.Code,
		)
	}
}

func TestRequireRoleAllowsService(
	t *testing.T,
) {
	handler :=
		RequireRole(
			"service",
		)(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(
					http.StatusOK,
				)
			},
		)

	recorder :=
		httptest.NewRecorder()

	handler(
		recorder,
		requestWithPrincipal(
			"service",
		),
	)

	if recorder.Code !=
		http.StatusOK {

		t.Fatalf(
			"expected 200, got %d",
			recorder.Code,
		)
	}
}

func TestRequireRoleBlocksServiceFromAdmin(
	t *testing.T,
) {
	handler :=
		RequireRole(
			"admin",
		)(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(
					http.StatusOK,
				)
			},
		)

	recorder :=
		httptest.NewRecorder()

	handler(
		recorder,
		requestWithPrincipal(
			"service",
		),
	)

	if recorder.Code !=
		http.StatusForbidden {

		t.Fatalf(
			"expected 403, got %d",
			recorder.Code,
		)
	}
}

func TestRequireRoleRejectsMissingPrincipal(
	t *testing.T,
) {
	handler :=
		RequireRole(
			"admin",
		)(
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
