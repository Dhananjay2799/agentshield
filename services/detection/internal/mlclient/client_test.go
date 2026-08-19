package mlclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPredictReturnsPrediction(t *testing.T) {
	server :=
		httptest.NewServer(
			http.HandlerFunc(
				func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					if r.Method != http.MethodPost {
						t.Fatalf(
							"expected POST, got %s",
							r.Method,
						)
					}

					if r.URL.Path != "/predict" {
						t.Fatalf(
							"expected /predict, got %s",
							r.URL.Path,
						)
					}

					var request PredictionRequest

					if err :=
						json.NewDecoder(
							r.Body,
						).Decode(
							&request,
						); err != nil {
						t.Fatalf(
							"decode request: %v",
							err,
						)
					}

					if request.EventCount != 35 {
						t.Fatalf(
							"expected event count 35, got %f",
							request.EventCount,
						)
					}

					w.Header().Set(
						"Content-Type",
						"application/json",
					)

					_ = json.NewEncoder(
						w,
					).Encode(
						PredictionResponse{
							ReconstructionError: 0.1357,
							Threshold:           0.0069,
							IsAnomaly:           true,
							ScoreRatio:          19.53,
							Model:               "behavior-autoencoder-v1",
						},
					)
				},
			),
		)

	defer server.Close()

	client :=
		New(
			server.URL,
			time.Second,
		)

	prediction, err :=
		client.Predict(
			context.Background(),
			PredictionRequest{
				EventCount:             35,
				DenyRatio:              0.90,
				HighRiskRatio:          0.85,
				ActionDiversityRatio:   0.90,
				ResourceDiversityRatio: 0.90,
				AverageRiskScore:       90,
				ProductionAccessRatio:  1.0,
				SensitiveActionRatio:   0.80,
			},
		)

	if err != nil {
		t.Fatalf(
			"Predict returned error: %v",
			err,
		)
	}

	if !prediction.IsAnomaly {
		t.Fatal(
			"expected anomalous prediction",
		)
	}

	if prediction.Model !=
		"behavior-autoencoder-v1" {
		t.Fatalf(
			"unexpected model: %s",
			prediction.Model,
		)
	}
}

func TestPredictReturnsErrorForServerFailure(
	t *testing.T,
) {
	server :=
		httptest.NewServer(
			http.HandlerFunc(
				func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					http.Error(
						w,
						"model unavailable",
						http.StatusServiceUnavailable,
					)
				},
			),
		)

	defer server.Close()

	client :=
		New(
			server.URL,
			time.Second,
		)

	_, err :=
		client.Predict(
			context.Background(),
			PredictionRequest{},
		)

	if err == nil {
		t.Fatal(
			"expected error from failed ML service",
		)
	}
}

func TestPredictHonorsTimeout(
	t *testing.T,
) {
	server :=
		httptest.NewServer(
			http.HandlerFunc(
				func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					time.Sleep(
						100 * time.Millisecond,
					)

					w.WriteHeader(
						http.StatusOK,
					)
				},
			),
		)

	defer server.Close()

	client :=
		New(
			server.URL,
			10*time.Millisecond,
		)

	_, err :=
		client.Predict(
			context.Background(),
			PredictionRequest{},
		)

	if err == nil {
		t.Fatal(
			"expected timeout error",
		)
	}
}
