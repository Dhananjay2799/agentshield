package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RateLimitKeyFunc func(
	r *http.Request,
) string

type bucket struct {
	Tokens     float64
	LastRefill time.Time
	LastSeen   time.Time
}

type RateLimiter struct {
	mu sync.Mutex

	buckets map[string]*bucket

	ratePerSecond float64
	burst         float64

	keyFunc RateLimitKeyFunc
}

func NewRateLimiter(
	requestsPerMinute int,
	burst int,
	keyFunc RateLimitKeyFunc,
) *RateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60
	}

	if burst <= 0 {
		burst = 1
	}

	return &RateLimiter{
		buckets: make(
			map[string]*bucket,
		),

		ratePerSecond: float64(requestsPerMinute) / 60.0,

		burst: float64(burst),

		keyFunc: keyFunc,
	}
}

func (l *RateLimiter) Middleware(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		key := l.keyFunc(r)

		if key == "" {
			key = "anonymous"
		}

		allowed, retryAfter :=
			l.allow(key)

		if !allowed {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.Header().Set(
				"Retry-After",
				strconv.Itoa(
					int(
						retryAfter.Seconds(),
					)+1,
				),
			)

			w.WriteHeader(
				http.StatusTooManyRequests,
			)

			_ = json.NewEncoder(
				w,
			).Encode(
				map[string]any{
					"error": "rate limit exceeded",

					"retry_after_seconds": int(
						retryAfter.Seconds(),
					) + 1,
				},
			)

			return
		}

		next(w, r)
	}
}

func (l *RateLimiter) allow(
	key string,
) (bool, time.Duration) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	current, exists :=
		l.buckets[key]

	if !exists {
		l.buckets[key] =
			&bucket{
				Tokens: l.burst - 1,

				LastRefill: now,

				LastSeen: now,
			}

		return true, 0
	}

	elapsed :=
		now.Sub(
			current.LastRefill,
		).Seconds()

	current.Tokens +=
		elapsed *
			l.ratePerSecond

	if current.Tokens >
		l.burst {

		current.Tokens =
			l.burst
	}

	current.LastRefill = now
	current.LastSeen = now

	if current.Tokens >= 1 {
		current.Tokens--
		return true, 0
	}

	missing :=
		1 - current.Tokens

	retrySeconds :=
		missing /
			l.ratePerSecond

	return false,
		time.Duration(
			retrySeconds *
				float64(time.Second),
		)
}

func (l *RateLimiter) Cleanup(
	maxIdle time.Duration,
) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	for key, current := range l.buckets {

		if now.Sub(
			current.LastSeen,
		) > maxIdle {

			delete(
				l.buckets,
				key,
			)
		}
	}
}

func PrincipalRateLimitKey(
	r *http.Request,
) string {
	principal, ok :=
		PrincipalFromContext(
			r.Context(),
		)

	if !ok {
		return ""
	}

	return "principal:" +
		principal.ID
}

func SessionRateLimitKey(
	r *http.Request,
) string {
	session, ok :=
		SessionFromContext(
			r.Context(),
		)

	if !ok {
		return ""
	}

	return "session:" +
		session.ID
}

func IPRateLimitKey(
	r *http.Request,
) string {
	return "ip:" +
		r.RemoteAddr
}
