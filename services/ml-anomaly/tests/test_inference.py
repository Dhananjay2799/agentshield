from pathlib import Path

from app.features import (
    BehaviorFeatures,
)
from app.inference import (
    InferenceEngine,
)


MODEL_PATH = Path(
    "models/behavior_autoencoder.pt"
)


def test_inference_loads_model():
    engine = InferenceEngine(
        MODEL_PATH
    )

    assert engine.threshold > 0
    assert len(
        engine.feature_names
    ) == 8


def test_normal_behavior_is_not_anomalous():
    engine = InferenceEngine(
        MODEL_PATH
    )

    prediction = engine.predict(
        BehaviorFeatures(
            event_count=8,
            deny_ratio=0.03,
            high_risk_ratio=0.02,
            action_diversity_ratio=0.30,
            resource_diversity_ratio=0.25,
            average_risk_score=18,
            production_access_ratio=0.15,
            sensitive_action_ratio=0.01,
        )
    )

    assert (
        prediction
        .reconstruction_error
        >= 0
    )

    assert not prediction.is_anomaly


def test_malicious_behavior_is_anomalous():
    engine = InferenceEngine(
        MODEL_PATH
    )

    prediction = engine.predict(
        BehaviorFeatures(
            event_count=35,
            deny_ratio=0.90,
            high_risk_ratio=0.85,
            action_diversity_ratio=0.90,
            resource_diversity_ratio=0.90,
            average_risk_score=90,
            production_access_ratio=1.0,
            sensitive_action_ratio=0.80,
        )
    )

    assert prediction.is_anomaly

    assert (
        prediction
        .reconstruction_error
        > prediction.threshold
    )

    assert prediction.score_ratio > 1


def test_prediction_contains_threshold():
    engine = InferenceEngine(
        MODEL_PATH
    )

    prediction = engine.predict(
        BehaviorFeatures(
            event_count=8,
            deny_ratio=0.05,
            high_risk_ratio=0.03,
            action_diversity_ratio=0.30,
            resource_diversity_ratio=0.25,
            average_risk_score=20,
            production_access_ratio=0.10,
            sensitive_action_ratio=0.02,
        )
    )

    assert prediction.threshold == (
        engine.threshold
    )