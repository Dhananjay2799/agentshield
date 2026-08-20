from __future__ import annotations

from collections import deque
from dataclasses import dataclass
from statistics import mean
from threading import Lock


@dataclass(frozen=True)
class DriftSnapshot:
    sample_count: int
    mean_score_ratio: float
    anomaly_rate: float
    drift_score: float
    is_drifting: bool


class DriftDetector:
    def __init__(
        self,
        window_size: int = 100,
        minimum_samples: int = 20,
        score_ratio_threshold: float = 1.5,
        anomaly_rate_threshold: float = 0.25,
    ) -> None:
        if window_size <= 0:
            raise ValueError(
                "window_size must be positive"
            )

        if minimum_samples <= 0:
            raise ValueError(
                "minimum_samples must be positive"
            )

        if minimum_samples > window_size:
            raise ValueError(
                "minimum_samples cannot exceed window_size"
            )

        self.window_size = window_size
        self.minimum_samples = minimum_samples
        self.score_ratio_threshold = (
            score_ratio_threshold
        )
        self.anomaly_rate_threshold = (
            anomaly_rate_threshold
        )

        self._score_ratios: deque[float] = deque(
            maxlen=window_size
        )

        self._anomalies: deque[bool] = deque(
            maxlen=window_size
        )

        self._lock = Lock()

    def record(
        self,
        score_ratio: float,
        is_anomaly: bool,
    ) -> DriftSnapshot:
        with self._lock:
            self._score_ratios.append(
                max(
                    0.0,
                    float(score_ratio),
                )
            )

            self._anomalies.append(
                bool(is_anomaly)
            )

            return self._snapshot_unlocked()

    def snapshot(
        self,
    ) -> DriftSnapshot:
        with self._lock:
            return self._snapshot_unlocked()

    def _snapshot_unlocked(
        self,
    ) -> DriftSnapshot:
        sample_count = len(
            self._score_ratios
        )

        if sample_count == 0:
            return DriftSnapshot(
                sample_count=0,
                mean_score_ratio=0.0,
                anomaly_rate=0.0,
                drift_score=0.0,
                is_drifting=False,
            )

        mean_score_ratio = mean(
            self._score_ratios
        )

        anomaly_rate = (
            sum(
                1
                for value in self._anomalies
                if value
            )
            / sample_count
        )

        score_component = (
            mean_score_ratio
            / self.score_ratio_threshold
            if self.score_ratio_threshold > 0
            else 0.0
        )

        anomaly_component = (
            anomaly_rate
            / self.anomaly_rate_threshold
            if self.anomaly_rate_threshold > 0
            else 0.0
        )

        drift_score = max(
            score_component,
            anomaly_component,
        )

        is_drifting = (
            sample_count
            >= self.minimum_samples
            and drift_score >= 1.0
        )

        return DriftSnapshot(
            sample_count=
                sample_count,
            mean_score_ratio=
                mean_score_ratio,
            anomaly_rate=
                anomaly_rate,
            drift_score=
                drift_score,
            is_drifting=
                is_drifting,
        )