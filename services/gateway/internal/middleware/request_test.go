package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequireJSONAllowsApplicationJSON(
	t *testing.T,
) {
	handler :=
		RequireJSON(
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
			http.MethodPost,
			"/",
			strings.NewReader(
				`{"ok":true}`,
			),
		)

	request.Header.Set(
		"Content-Type",
		"application/json",
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

func TestRequireJSONAllowsCharset(
	t *testing.T,
) {
	handler :=
		RequireJSON(
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
			http.MethodPost,
			"/",
			strings.NewReader(
				`{"ok":true}`,
			),
		)

	request.Header.Set(
		"Content-Type",
		"application/json; charset=utf-8",
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

func TestRequireJSONRejectsMissingContentType(
	t *testing.T,
) {
	handler :=
		RequireJSON(
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
			http.MethodPost,
			"/",
			strings.NewReader(
				`{"ok":true}`,
			),
		)

	recorder :=
		httptest.NewRecorder()

	handler(
		recorder,
		request,
	)

	if recorder.Code !=
		http.StatusUnsupportedMediaType {

		t.Fatalf(
			"expected 415, got %d",
			recorder.Code,
		)
	}
}

func TestRequireJSONRejectsWrongContentType(
	t *testing.T,
) {
	handler :=
		RequireJSON(
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
			http.MethodPost,
			"/",
			strings.NewReader(
				`{"ok":true}`,
			),
		)

	request.Header.Set(
		"Content-Type",
		"text/plain",
	)

	recorder :=
		httptest.NewRecorder()

	handler(
		recorder,
		request,
	)

	if recorder.Code !=
		http.StatusUnsupportedMediaType {

		t.Fatalf(
			"expected 415, got %d",
			recorder.Code,
		)
	}
}

func TestLimitBodyAllowsSmallRequest(
	t *testing.T,
) {
	handler :=
		LimitBody(
			1024,
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				buffer :=
					make(
						[]byte,
						32,
					)

				_, _ =
					r.Body.Read(
						buffer,
					)

				w.WriteHeader(
					http.StatusOK,
				)
			},
		)

	request :=
		httptest.NewRequest(
			http.MethodPost,
			"/",
			strings.NewReader(
				`{"small":"payload"}`,
			),
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
