package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
)

func decodeJSONBody(
	w http.ResponseWriter,
	r *http.Request,
	dst any,
) bool {
	err := json.NewDecoder(
		r.Body,
	).Decode(dst)

	if err == nil {
		return true
	}

	var maxBytesError *http.MaxBytesError

	if errors.As(
		err,
		&maxBytesError,
	) {
		writeJSON(
			w,
			http.StatusRequestEntityTooLarge,
			map[string]string{
				"error": "request body exceeds maximum size",
			},
		)

		return false
	}

	writeJSON(
		w,
		http.StatusBadRequest,
		map[string]string{
			"error": "invalid JSON body",
		},
	)

	return false
}
