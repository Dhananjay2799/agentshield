package middleware

import (
	"encoding/json"
	"mime"
	"net/http"
	"strings"
)

func LimitBody(
	maxBytes int64,
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		r.Body = http.MaxBytesReader(
			w,
			r.Body,
			maxBytes,
		)

		next(w, r)
	}
}

func RequireJSON(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		contentType :=
			strings.TrimSpace(
				r.Header.Get(
					"Content-Type",
				),
			)

		if contentType == "" {
			writeRequestError(
				w,
				http.StatusUnsupportedMediaType,
				"Content-Type application/json is required",
			)
			return
		}

		mediaType, _, err :=
			mime.ParseMediaType(
				contentType,
			)

		if err != nil ||
			mediaType != "application/json" {

			writeRequestError(
				w,
				http.StatusUnsupportedMediaType,
				"Content-Type application/json is required",
			)
			return
		}

		next(w, r)
	}
}

func writeRequestError(
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
