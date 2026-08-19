from __future__ import annotations

import csv
import random
from pathlib import Path

from app.features import BehaviorFeatures


def clamp(
    value: float,
    minimum: float = 0.0,
    maximum: float = 1.0,
) -> float:
    return max(
        minimum,
        min(value, maximum),
    )


def generate_normal_behavior(
    count: int,
    seed: int = 42,
) -> list[BehaviorFeatures]:
    rng = random.Random(seed)

    rows: list[BehaviorFeatures] = []

    for _ in range(count):
        event_count = max(
            1.0,
            rng.gauss(8.0, 2.0),
        )

        deny_ratio = clamp(
            rng.gauss(0.05, 0.03)
        )

        high_risk_ratio = clamp(
            rng.gauss(0.03, 0.02)
        )

        action_diversity_ratio = clamp(
            rng.gauss(0.30, 0.08)
        )

        resource_diversity_ratio = clamp(
            rng.gauss(0.25, 0.08)
        )

        average_risk_score = max(
            0.0,
            min(
                rng.gauss(18.0, 6.0),
                100.0,
            ),
        )

        production_access_ratio = clamp(
            rng.gauss(0.15, 0.08)
        )

        sensitive_action_ratio = clamp(
            rng.gauss(0.02, 0.02)
        )

        rows.append(
            BehaviorFeatures(
                event_count=event_count,
                deny_ratio=deny_ratio,
                high_risk_ratio=high_risk_ratio,
                action_diversity_ratio=action_diversity_ratio,
                resource_diversity_ratio=resource_diversity_ratio,
                average_risk_score=average_risk_score,
                production_access_ratio=production_access_ratio,
                sensitive_action_ratio=sensitive_action_ratio,
            )
        )

    return rows


def generate_anomalous_behavior(
    count: int,
    seed: int = 1337,
) -> list[BehaviorFeatures]:
    rng = random.Random(seed)

    rows: list[BehaviorFeatures] = []

    for _ in range(count):
        event_count = max(
            10.0,
            rng.gauss(35.0, 10.0),
        )

        deny_ratio = clamp(
            rng.gauss(0.75, 0.15)
        )

        high_risk_ratio = clamp(
            rng.gauss(0.70, 0.15)
        )

        action_diversity_ratio = clamp(
            rng.gauss(0.80, 0.12)
        )

        resource_diversity_ratio = clamp(
            rng.gauss(0.80, 0.12)
        )

        average_risk_score = max(
            0.0,
            min(
                rng.gauss(80.0, 10.0),
                100.0,
            ),
        )

        production_access_ratio = clamp(
            rng.gauss(0.90, 0.08)
        )

        sensitive_action_ratio = clamp(
            rng.gauss(0.70, 0.15)
        )

        rows.append(
            BehaviorFeatures(
                event_count=event_count,
                deny_ratio=deny_ratio,
                high_risk_ratio=high_risk_ratio,
                action_diversity_ratio=action_diversity_ratio,
                resource_diversity_ratio=resource_diversity_ratio,
                average_risk_score=average_risk_score,
                production_access_ratio=production_access_ratio,
                sensitive_action_ratio=sensitive_action_ratio,
            )
        )

    return rows


def write_dataset(
    path: Path,
    rows: list[BehaviorFeatures],
    label: str,
) -> None:
    path.parent.mkdir(
        parents=True,
        exist_ok=True,
    )

    with path.open(
        "w",
        newline="",
        encoding="utf-8",
    ) as file:
        writer = csv.writer(file)

        writer.writerow(
            [
                "event_count",
                "deny_ratio",
                "high_risk_ratio",
                "action_diversity_ratio",
                "resource_diversity_ratio",
                "average_risk_score",
                "production_access_ratio",
                "sensitive_action_ratio",
                "label",
            ]
        )

        for row in rows:
            writer.writerow(
                [
                    *row.as_list(),
                    label,
                ]
            )