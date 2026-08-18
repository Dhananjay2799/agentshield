package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type AlertmanagerPayload struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
	Status            string            `json:"status"`
	Receiver          string            `json:"receiver"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Alerts            []struct {
		Status       string            `json:"status"`
		Labels       map[string]string `json:"labels"`
		Annotations  map[string]string `json:"annotations"`
		StartsAt     time.Time         `json:"startsAt"`
		EndsAt       time.Time         `json:"endsAt"`
		GeneratorURL string            `json:"generatorURL"`
		Fingerprint  string            `json:"fingerprint"`
	} `json:"alerts"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(
			map[string]string{
				"service": "agentshield-alert-receiver",
				"status":  "healthy",
			},
		)
	})

	mux.HandleFunc("POST /v1/alerts", func(w http.ResponseWriter, r *http.Request) {
		var payload AlertmanagerPayload

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(
				w,
				`{"error":"invalid alert payload"}`,
				http.StatusBadRequest,
			)
			return
		}

		log.Printf(
			"ALERT RECEIVED receiver=%s status=%s alerts=%d group=%v",
			payload.Receiver,
			payload.Status,
			len(payload.Alerts),
			payload.GroupLabels,
		)

		for _, alert := range payload.Alerts {
			log.Printf(
				"alert=%s severity=%s service=%s status=%s summary=%s",
				alert.Labels["alertname"],
				alert.Labels["severity"],
				alert.Labels["service"],
				alert.Status,
				alert.Annotations["summary"],
			)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(
			map[string]string{
				"status": "received",
			},
		)
	})

	server := &http.Server{
		Addr:              ":8084",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println(
		"AgentShield Alert Receiver starting on http://localhost:8084",
	)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
