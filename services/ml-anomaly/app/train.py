from __future__ import annotations

import csv
import json
import random
from dataclasses import asdict, dataclass
from pathlib import Path

import torch
from torch import nn
from torch.utils.data import DataLoader, TensorDataset

from app.features import (
    BehaviorFeatures,
    FeatureScaler,
)
from app.model import (
    BehaviorAutoencoder,
    reconstruction_error,
)


@dataclass(frozen=True)
class TrainingConfig:
    seed: int = 42
    epochs: int = 120
    batch_size: int = 64
    learning_rate: float = 0.001
    validation_ratio: float = 0.20
    threshold_quantile: float = 0.99


def load_dataset(
    path: Path,
) -> list[BehaviorFeatures]:
    rows: list[BehaviorFeatures] = []

    with path.open(
        "r",
        encoding="utf-8",
        newline="",
    ) as file:
        reader = csv.DictReader(file)

        for row in reader:
            rows.append(
                BehaviorFeatures(
                    event_count=float(
                        row["event_count"]
                    ),
                    deny_ratio=float(
                        row["deny_ratio"]
                    ),
                    high_risk_ratio=float(
                        row["high_risk_ratio"]
                    ),
                    action_diversity_ratio=float(
                        row["action_diversity_ratio"]
                    ),
                    resource_diversity_ratio=float(
                        row["resource_diversity_ratio"]
                    ),
                    average_risk_score=float(
                        row["average_risk_score"]
                    ),
                    production_access_ratio=float(
                        row["production_access_ratio"]
                    ),
                    sensitive_action_ratio=float(
                        row["sensitive_action_ratio"]
                    ),
                )
            )

    return rows


def prepare_tensor(
    rows: list[BehaviorFeatures],
    scaler: FeatureScaler,
) -> torch.Tensor:
    return scaler.transform_batch(
        rows
    )


def split_normal_data(
    tensor: torch.Tensor,
    validation_ratio: float,
    seed: int,
) -> tuple[torch.Tensor, torch.Tensor]:
    generator = torch.Generator()
    generator.manual_seed(seed)

    indexes = torch.randperm(
        tensor.shape[0],
        generator=generator,
    )

    validation_count = max(
        1,
        int(
            tensor.shape[0]
            * validation_ratio
        ),
    )

    validation_indexes = indexes[
        :validation_count
    ]

    training_indexes = indexes[
        validation_count:
    ]

    return (
        tensor[training_indexes],
        tensor[validation_indexes],
    )


def train_model(
    model: BehaviorAutoencoder,
    training_tensor: torch.Tensor,
    validation_tensor: torch.Tensor,
    config: TrainingConfig,
) -> list[dict[str, float]]:
    dataset = TensorDataset(
        training_tensor
    )

    loader = DataLoader(
        dataset,
        batch_size=config.batch_size,
        shuffle=True,
    )

    optimizer = torch.optim.Adam(
        model.parameters(),
        lr=config.learning_rate,
    )

    loss_function = nn.MSELoss()

    history: list[
        dict[str, float]
    ] = []

    for epoch in range(
        1,
        config.epochs + 1,
    ):
        model.train()

        training_loss = 0.0
        samples_seen = 0

        for (batch,) in loader:
            optimizer.zero_grad()

            reconstructed = model(
                batch
            )

            loss = loss_function(
                reconstructed,
                batch,
            )

            loss.backward()
            optimizer.step()

            batch_size = batch.shape[0]

            training_loss += (
                loss.item()
                * batch_size
            )

            samples_seen += batch_size

        training_loss /= max(
            1,
            samples_seen,
        )

        model.eval()

        with torch.no_grad():
            validation_errors = (
                reconstruction_error(
                    model,
                    validation_tensor,
                )
            )

            validation_loss = (
                validation_errors
                .mean()
                .item()
            )

        history.append(
            {
                "epoch": float(epoch),
                "training_loss":
                    training_loss,
                "validation_loss":
                    validation_loss,
            }
        )

        if (
            epoch == 1
            or epoch % 10 == 0
            or epoch == config.epochs
        ):
            print(
                f"epoch={epoch:03d} "
                f"train_loss={training_loss:.6f} "
                f"validation_loss={validation_loss:.6f}"
            )

    return history


def calculate_threshold(
    model: BehaviorAutoencoder,
    validation_tensor: torch.Tensor,
    quantile: float,
) -> float:
    model.eval()

    with torch.no_grad():
        errors = reconstruction_error(
            model,
            validation_tensor,
        )

    threshold = torch.quantile(
        errors,
        quantile,
    )

    return float(
        threshold.item()
    )


def evaluate_model(
    model: BehaviorAutoencoder,
    normal_tensor: torch.Tensor,
    anomaly_tensor: torch.Tensor,
    threshold: float,
) -> dict[str, float]:
    model.eval()

    with torch.no_grad():
        normal_errors = (
            reconstruction_error(
                model,
                normal_tensor,
            )
        )

        anomaly_errors = (
            reconstruction_error(
                model,
                anomaly_tensor,
            )
        )

    normal_predictions = (
        normal_errors > threshold
    )

    anomaly_predictions = (
        anomaly_errors > threshold
    )

    false_positives = int(
        normal_predictions.sum().item()
    )

    true_negatives = (
        normal_tensor.shape[0]
        - false_positives
    )

    true_positives = int(
        anomaly_predictions.sum().item()
    )

    false_negatives = (
        anomaly_tensor.shape[0]
        - true_positives
    )

    precision_denominator = (
        true_positives
        + false_positives
    )

    recall_denominator = (
        true_positives
        + false_negatives
    )

    precision = (
        true_positives
        / precision_denominator
        if precision_denominator
        else 0.0
    )

    recall = (
        true_positives
        / recall_denominator
        if recall_denominator
        else 0.0
    )

    f1 = (
        2
        * precision
        * recall
        / (
            precision
            + recall
        )
        if precision + recall
        else 0.0
    )

    false_positive_rate = (
        false_positives
        / (
            false_positives
            + true_negatives
        )
        if false_positives
        + true_negatives
        else 0.0
    )

    return {
        "threshold": threshold,
        "normal_mean_error":
            float(
                normal_errors
                .mean()
                .item()
            ),
        "anomaly_mean_error":
            float(
                anomaly_errors
                .mean()
                .item()
            ),
        "true_positives":
            float(true_positives),
        "false_positives":
            float(false_positives),
        "true_negatives":
            float(true_negatives),
        "false_negatives":
            float(false_negatives),
        "precision":
            precision,
        "recall":
            recall,
        "f1":
            f1,
        "false_positive_rate":
            false_positive_rate,
    }


def main() -> None:
    config = TrainingConfig()

    random.seed(
        config.seed
    )

    torch.manual_seed(
        config.seed
    )

    data_dir = Path(
        "data"
    )

    model_dir = Path(
        "models"
    )

    model_dir.mkdir(
        parents=True,
        exist_ok=True,
    )

    normal_rows = load_dataset(
        data_dir / "normal.csv"
    )

    anomaly_rows = load_dataset(
        data_dir / "anomalies.csv"
    )

    scaler = FeatureScaler()

    normal_tensor = prepare_tensor(
        normal_rows,
        scaler,
    )

    anomaly_tensor = prepare_tensor(
        anomaly_rows,
        scaler,
    )

    (
        training_tensor,
        validation_tensor,
    ) = split_normal_data(
        normal_tensor,
        config.validation_ratio,
        config.seed,
    )

    print(
        f"normal_samples={normal_tensor.shape[0]}"
    )

    print(
        f"anomaly_samples={anomaly_tensor.shape[0]}"
    )

    print(
        f"training_samples={training_tensor.shape[0]}"
    )

    print(
        f"validation_samples={validation_tensor.shape[0]}"
    )

    model = BehaviorAutoencoder()

    history = train_model(
        model,
        training_tensor,
        validation_tensor,
        config,
    )

    threshold = calculate_threshold(
        model,
        validation_tensor,
        config.threshold_quantile,
    )

    metrics = evaluate_model(
        model,
        validation_tensor,
        anomaly_tensor,
        threshold,
    )

    checkpoint_path = (
        model_dir
        / "behavior_autoencoder.pt"
    )

    torch.save(
        {
            "model_state_dict":
                model.state_dict(),
            "input_dim":
                model.input_dim,
            "latent_dim":
                model.latent_dim,
            "threshold":
                threshold,
            "feature_names": [
                "event_count",
                "deny_ratio",
                "high_risk_ratio",
                "action_diversity_ratio",
                "resource_diversity_ratio",
                "average_risk_score",
                "production_access_ratio",
                "sensitive_action_ratio",
            ],
            "scaler": {
                "event_count_scale":
                    scaler.event_count_scale,
                "risk_score_scale":
                    scaler.risk_score_scale,
            },
            "training_config":
                asdict(config),
        },
        checkpoint_path,
    )

    metrics_path = (
        model_dir
        / "training_metrics.json"
    )

    metrics_path.write_text(
        json.dumps(
            {
                "config":
                    asdict(config),
                "metrics":
                    metrics,
                "last_epoch":
                    history[-1],
            },
            indent=2,
        ),
        encoding="utf-8",
    )

    print("")
    print(
        f"model_saved={checkpoint_path}"
    )

    print(
        f"threshold={threshold:.8f}"
    )

    print(
        f"normal_mean_error="
        f"{metrics['normal_mean_error']:.8f}"
    )

    print(
        f"anomaly_mean_error="
        f"{metrics['anomaly_mean_error']:.8f}"
    )

    print(
        f"precision="
        f"{metrics['precision']:.4f}"
    )

    print(
        f"recall="
        f"{metrics['recall']:.4f}"
    )

    print(
        f"f1="
        f"{metrics['f1']:.4f}"
    )

    print(
        f"false_positive_rate="
        f"{metrics['false_positive_rate']:.4f}"
    )


if __name__ == "__main__":
    main()