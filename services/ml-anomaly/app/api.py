from __future__ import annotations

from pathlib import Path
from time import perf_counter

from fastapi import FastAPI
from pydantic import BaseModel, Field
from prometheus_client import (
    CONTENT_TYPE_LATEST,
    Counter,
    Histogram,
    generate_latest,
)
from starlette.responses import Response

from app.features import BehaviorFeatures
from app.inference import InferenceEngine


MODEL_PATH = Path(
    "models/behavior_autoencoder.pt"
)

engine = InferenceEngine(
    MODEL_PATH
)

app = FastAPI(
    title="AgentShield ML Anomaly Service",
    version="1.0.0",
)


PREDICTIONS = Counter(
    "agentshield_ml_predictions_total",
    "Total ML anomaly predictions.",
)

ANOMALIES = Counter(
    "agentshield_ml_anomalies_total",
    "Total ML predictions classified as anomalous.",
)

PREDICTION_ERRORS = Counter(
    "agentshield_ml_prediction_errors_total",
    "Total ML prediction failures.",
)

PREDICTION_LATENCY = Histogram(
    "agentshield_ml_prediction_latency_seconds",
    "ML anomaly prediction latency in seconds.",
)


class PredictionRequest(BaseModel):
    event_count: float = Field(
        ge=0,
    )

    deny_ratio: float = Field(
        ge=0,
        le=1,
    )

    high_risk_ratio: float = Field(
        ge=0,
        le=1,
    )

    action_diversity_ratio: float = Field(
        ge=0,
        le=1,
    )

    resource_diversity_ratio: float = Field(
        ge=0,
        le=1,
    )

    average_risk_score: float = Field(
        ge=0,
        le=100,
    )

    production_access_ratio: float = Field(
        ge=0,
        le=1,
    )

    sensitive_action_ratio: float = Field(
        ge=0,
        le=1,
    )


class PredictionResponse(BaseModel):
    reconstruction_error: float
    threshold: float
    is_anomaly: bool
    score_ratio: float
    model: str


@app.get(
    "/health"
)
def health() -> dict[str, str]:
    return {
        "service":
            "agentshield-ml-anomaly",
        "status":
            "healthy",
    }


@app.get(
    "/ready"
)
def ready() -> dict[str, object]:
    return {
        "service":
            "agentshield-ml-anomaly",
        "status":
            "ready",
        "model_loaded":
            True,
        "threshold":
            engine.threshold,
        "feature_count":
            len(engine.feature_names),
    }


@app.post(
    "/predict",
    response_model=PredictionResponse,
)
def predict(
    request: PredictionRequest,
) -> PredictionResponse:
    started = perf_counter()

    try:
        prediction = engine.predict(
            BehaviorFeatures(
                event_count=
                    request.event_count,
                deny_ratio=
                    request.deny_ratio,
                high_risk_ratio=
                    request.high_risk_ratio,
                action_diversity_ratio=
                    request.action_diversity_ratio,
                resource_diversity_ratio=
                    request.resource_diversity_ratio,
                average_risk_score=
                    request.average_risk_score,
                production_access_ratio=
                    request.production_access_ratio,
                sensitive_action_ratio=
                    request.sensitive_action_ratio,
            )
        )

        PREDICTIONS.inc()

        if prediction.is_anomaly:
            ANOMALIES.inc()

        return PredictionResponse(
            reconstruction_error=
                prediction.reconstruction_error,
            threshold=
                prediction.threshold,
            is_anomaly=
                prediction.is_anomaly,
            score_ratio=
                prediction.score_ratio,
            model=
                "behavior-autoencoder-v1",
        )

    except Exception:
        PREDICTION_ERRORS.inc()
        raise

    finally:
        PREDICTION_LATENCY.observe(
            perf_counter()
            - started
        )


@app.get(
    "/metrics"
)
def metrics() -> Response:
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST,
    )