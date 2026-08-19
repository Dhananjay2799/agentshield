import pytest
import torch

from app.features import (
    FEATURE_COUNT,
    FEATURE_NAMES,
    BehaviorFeatures,
    FeatureScaler,
    stack_feature_vectors,
    validate_feature_vector,
)


def test_feature_contract_has_eight_features():
    assert FEATURE_COUNT == 8
    assert len(FEATURE_NAMES) == 8


def test_behavior_features_preserve_order():
    features = BehaviorFeatures(
        event_count=10,
        deny_ratio=0.2,
        high_risk_ratio=0.1,
        action_diversity_ratio=0.3,
        resource_diversity_ratio=0.4,
        average_risk_score=25,
        production_access_ratio=0.5,
        sensitive_action_ratio=0.1,
    )

    assert features.as_list() == [
        10.0,
        0.2,
        0.1,
        0.3,
        0.4,
        25.0,
        0.5,
        0.1,
    ]


def test_scaler_normalizes_features():
    scaler = FeatureScaler()

    features = BehaviorFeatures(
        event_count=50,
        deny_ratio=0.2,
        high_risk_ratio=0.3,
        action_diversity_ratio=0.4,
        resource_diversity_ratio=0.5,
        average_risk_score=80,
        production_access_ratio=0.6,
        sensitive_action_ratio=0.7,
    )

    result = scaler.transform(features)

    expected = torch.tensor(
        [
            0.5,
            0.2,
            0.3,
            0.4,
            0.5,
            0.8,
            0.6,
            0.7,
        ],
        dtype=torch.float32,
    )

    assert torch.allclose(
        result,
        expected,
    )


def test_scaler_clamps_features():
    scaler = FeatureScaler()

    features = BehaviorFeatures(
        event_count=500,
        deny_ratio=2.0,
        high_risk_ratio=-1.0,
        action_diversity_ratio=1.5,
        resource_diversity_ratio=0.5,
        average_risk_score=200,
        production_access_ratio=-5,
        sensitive_action_ratio=3,
    )

    result = scaler.transform(features)

    assert torch.all(result >= 0)
    assert torch.all(result <= 1)


def test_stack_feature_vectors():
    vectors = [
        BehaviorFeatures(
            event_count=1,
            deny_ratio=0,
            high_risk_ratio=0,
            action_diversity_ratio=1,
            resource_diversity_ratio=1,
            average_risk_score=10,
            production_access_ratio=0,
            sensitive_action_ratio=0,
        ),
        BehaviorFeatures(
            event_count=2,
            deny_ratio=0.5,
            high_risk_ratio=0.5,
            action_diversity_ratio=1,
            resource_diversity_ratio=1,
            average_risk_score=50,
            production_access_ratio=1,
            sensitive_action_ratio=0.5,
        ),
    ]

    result = stack_feature_vectors(vectors)

    assert result.shape == (2, 8)


def test_validate_feature_vector_rejects_wrong_size():
    with pytest.raises(ValueError):
        validate_feature_vector(
            [1.0, 2.0]
        )