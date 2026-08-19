from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any

import torch

from app.features import (
    BehaviorFeatures,
    FeatureScaler,
)
from app.model import (
    BehaviorAutoencoder,
    reconstruction_error,
)


@dataclass(frozen=True)
class AnomalyPrediction:
    reconstruction_error: float
    threshold: float
    is_anomaly: bool
    score_ratio: float


class InferenceEngine:
    def __init__(
        self,
        model_path: Path,
    ) -> None:
        checkpoint: dict[str, Any] = torch.load(
            model_path,
            map_location="cpu",
            weights_only=False,
        )

        self.threshold = float(
            checkpoint["threshold"]
        )

        self.feature_names = tuple(
            checkpoint["feature_names"]
        )

        scaler_config = checkpoint[
            "scaler"
        ]

        self.scaler = FeatureScaler(
            event_count_scale=float(
                scaler_config[
                    "event_count_scale"
                ]
            ),
            risk_score_scale=float(
                scaler_config[
                    "risk_score_scale"
                ]
            ),
        )

        self.model = BehaviorAutoencoder(
            input_dim=int(
                checkpoint["input_dim"]
            ),
            latent_dim=int(
                checkpoint["latent_dim"]
            ),
        )

        self.model.load_state_dict(
            checkpoint[
                "model_state_dict"
            ]
        )

        self.model.eval()

    def predict(
        self,
        features: BehaviorFeatures,
    ) -> AnomalyPrediction:
        tensor = (
            self.scaler
            .transform(features)
            .unsqueeze(0)
        )

        with torch.no_grad():
            error = reconstruction_error(
                self.model,
                tensor,
            )[0]

        error_value = float(
            error.item()
        )

        score_ratio = (
            error_value
            / self.threshold
            if self.threshold > 0
            else 0.0
        )

        return AnomalyPrediction(
            reconstruction_error=
                error_value,
            threshold=
                self.threshold,
            is_anomaly=
                error_value
                > self.threshold,
            score_ratio=
                score_ratio,
        )