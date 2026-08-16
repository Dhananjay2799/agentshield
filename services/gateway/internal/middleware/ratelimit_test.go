package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiterAllowsBurst(t *testing.T) {
	limiter := NewRateLimiter(
		60,
		3,
		func(r *http.Request) string {
			return "test"
		},
	)

	handler := limiter.Middleware(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		},
	)

	for i := 0; i < 3; i++ {
		request := httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

		recorder := httptest.NewRecorder()

		handler(
			recorder,
			request,
		)

		if recorder.Code != http.StatusOK {
			t.Fatalf(
				"request %d expected 200, got %d",
				i+1,
				recorder.Code,
			)
		}
	}
}

func TestRateLimiterRejectsBeyondBurst(t *testing.T) {
	limiter := NewRateLimiter(
		60,
		2,
		func(r *http.Request) string {
			return "test"
		},
	)

	handler := limiter.Middleware(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		},
	)

	for i := 0; i < 2; i++ {
		request := httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

		recorder := httptest.NewRecorder()

		handler(
			recorder,
			request,
		)

		if recorder.Code != http.StatusOK {
			t.Fatalf(
				"expected burst request %d to succeed, got %d",
				i+1,
				recorder.Code,
			)
		}
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler(
		recorder,
		request,
	)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected 429, got %d",
			recorder.Code,
		)
	}

	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal(
			"expected Retry-After header",
		)
	}
}

func TestRateLimiterSeparatesKeys(t *testing.T) {
	limiter := NewRateLimiter(
		60,
		1,
		func(r *http.Request) string {
			return r.Header.Get("X-Test-Key")
		},
	)

	handler := limiter.Middleware(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		},
	)

	requestOne := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	requestOne.Header.Set(
		"X-Test-Key",
		"principal-a",
	)

	recorderOne := httptest.NewRecorder()

	handler(
		recorderOne,
		requestOne,
	)

	if recorderOne.Code != http.StatusOK {
		t.Fatalf(
			"expected first principal to receive 200, got %d",
			recorderOne.Code,
		)
	}

	requestTwo := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	requestTwo.Header.Set(
		"X-Test-Key",
		"principal-b",
	)

	recorderTwo := httptest.NewRecorder()

	handler(
		recorderTwo,
		requestTwo,
	)

	if recorderTwo.Code != http.StatusOK {
		t.Fatalf(
			"expected independent principal to receive 200, got %d",
			recorderTwo.Code,
		)
	}

	requestThree := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	requestThree.Header.Set(
		"X-Test-Key",
		"principal-a",
	)

	recorderThree := httptest.NewRecorder()

	handler(
		recorderThree,
		requestThree,
	)

	if recorderThree.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected repeated principal to receive 429, got %d",
			recorderThree.Code,
		)
	}
}

func TestRateLimiterFallbackKey(t *testing.T) {
	limiter := NewRateLimiter(
		60,
		1,
		func(r *http.Request) string {
			return ""
		},
	)

	handler := limiter.Middleware(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		},
	)

	firstRequest := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	firstRecorder := httptest.NewRecorder()

	handler(
		firstRecorder,
		firstRequest,
	)

	if firstRecorder.Code != http.StatusOK {
		t.Fatalf(
			"expected first anonymous request 200, got %d",
			firstRecorder.Code,
		)
	}

	secondRequest := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	secondRecorder := httptest.NewRecorder()

	handler(
		secondRecorder,
		secondRequest,
	)

	if secondRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected second anonymous request 429, got %d",
			secondRecorder.Code,
		)
	}
}
