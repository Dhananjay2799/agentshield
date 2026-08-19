from __future__ import annotations

from dataclasses import dataclass
from typing import Sequence

import torch

FEATURE_NAMES = (
    "event_count",
    "deny_ratio",
    "high_risk_ratio",
    "action_diversity_ratio",
    "resource_diversity_ratio",
    "average_risk_score",
    "production_access_ratio",
    "sensitive_action_ratio",
)

FEATURE_COUNT = len(FEATURE_NAMES)


@dataclass(frozen=True)
class BehaviorFeatures:
    event_count: float
    deny_ratio: float
    high_risk_ratio: float
    action_diversity_ratio: float
    resource_diversity_ratio: float
    average_risk_score: float
    production_access_ratio: float
    sensitive_action_ratio: float

    def as_list(self) -> list[float]:
        return [
            float(self.event_count),
            float(self.deny_ratio),
            float(self.high_risk_ratio),
            float(self.action_diversity_ratio),
            float(self.resource_diversity_ratio),
            float(self.average_risk_score),
            float(self.production_access_ratio),
            float(self.sensitive_action_ratio),
        ]

    def as_tensor(self) -> torch.Tensor:
        return torch.tensor(
            self.as_list(),
            dtype=torch.float32,
        )


def validate_feature_vector(
    values: Sequence[float],
) -> None:
    if len(values) != FEATURE_COUNT:
        raise ValueError(
            f"expected {FEATURE_COUNT} features, got {len(values)}"
        )


def stack_feature_vectors(
    vectors: Sequence[BehaviorFeatures],
) -> torch.Tensor:
    if not vectors:
        return torch.empty(
            (0, FEATURE_COUNT),
            dtype=torch.float32,
        )

    return torch.stack(
        [vector.as_tensor() for vector in vectors]
    )


@dataclass(frozen=True)
class FeatureScaler:
    event_count_scale: float = 100.0
    risk_score_scale: float = 100.0

    def transform(
        self,
        features: BehaviorFeatures,
    ) -> torch.Tensor:
        values = features.as_list()

        normalized = [
            min(max(values[0] / self.event_count_scale, 0.0), 1.0),
            min(max(values[1], 0.0), 1.0),
            min(max(values[2], 0.0), 1.0),
            min(max(values[3], 0.0), 1.0),
            min(max(values[4], 0.0), 1.0),
            min(max(values[5] / self.risk_score_scale, 0.0), 1.0),
            min(max(values[6], 0.0), 1.0),
            min(max(values[7], 0.0), 1.0),
        ]

        return torch.tensor(
            normalized,
            dtype=torch.float32,
        )

    def transform_batch(
        self,
        features: Sequence[BehaviorFeatures],
    ) -> torch.Tensor:
        if not features:
            return torch.empty(
                (0, FEATURE_COUNT),
                dtype=torch.float32,
            )

        return torch.stack(
            [
                self.transform(feature)
                for feature in features
            ]
        )