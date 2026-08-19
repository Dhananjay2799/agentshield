from fastapi.testclient import TestClient

from app.api import app


client = TestClient(app)


def test_health():
    response = client.get(
        "/health"
    )

    assert response.status_code == 200

    payload = response.json()

    assert payload["status"] == "healthy"


def test_ready():
    response = client.get(
        "/ready"
    )

    assert response.status_code == 200

    payload = response.json()

    assert payload["model_loaded"] is True
    assert payload["feature_count"] == 8
    assert payload["threshold"] > 0


def test_predict_normal_behavior():
    response = client.post(
        "/predict",
        json={
            "event_count": 8,
            "deny_ratio": 0.03,
            "high_risk_ratio": 0.02,
            "action_diversity_ratio": 0.30,
            "resource_diversity_ratio": 0.25,
            "average_risk_score": 18,
            "production_access_ratio": 0.15,
            "sensitive_action_ratio": 0.01,
        },
    )

    assert response.status_code == 200

    payload = response.json()

    assert payload["is_anomaly"] is False
    assert payload["threshold"] > 0


def test_predict_malicious_behavior():
    response = client.post(
        "/predict",
        json={
            "event_count": 35,
            "deny_ratio": 0.90,
            "high_risk_ratio": 0.85,
            "action_diversity_ratio": 0.90,
            "resource_diversity_ratio": 0.90,
            "average_risk_score": 90,
            "production_access_ratio": 1.0,
            "sensitive_action_ratio": 0.80,
        },
    )

    assert response.status_code == 200

    payload = response.json()

    assert payload["is_anomaly"] is True

    assert (
        payload["reconstruction_error"]
        > payload["threshold"]
    )


def test_predict_rejects_invalid_ratio():
    response = client.post(
        "/predict",
        json={
            "event_count": 8,
            "deny_ratio": 5,
            "high_risk_ratio": 0,
            "action_diversity_ratio": 0.3,
            "resource_diversity_ratio": 0.3,
            "average_risk_score": 10,
            "production_access_ratio": 0,
            "sensitive_action_ratio": 0,
        },
    )

    assert response.status_code == 422


def test_metrics_endpoint():
    client.get(
        "/health"
    )

    response = client.get(
        "/metrics"
    )

    assert response.status_code == 200

    assert (
        "agentshield_ml_predictions_total"
        in response.text
    )