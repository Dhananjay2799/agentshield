from pathlib import Path

from app.model_metadata import (
    ModelMetadata,
    combined_dataset_hash,
    file_sha256,
    load_metadata,
    save_metadata,
)


def test_file_sha256_is_stable(
    tmp_path: Path,
):
    path = tmp_path / "sample.txt"

    path.write_text(
        "agentshield",
        encoding="utf-8",
    )

    first = file_sha256(path)
    second = file_sha256(path)

    assert first == second
    assert len(first) == 64


def test_combined_dataset_hash_changes(
    tmp_path: Path,
):
    normal = tmp_path / "normal.csv"
    anomaly = tmp_path / "anomalies.csv"

    normal.write_text(
        "normal-v1",
        encoding="utf-8",
    )

    anomaly.write_text(
        "anomaly-v1",
        encoding="utf-8",
    )

    first = combined_dataset_hash(
        [
            normal,
            anomaly,
        ]
    )

    anomaly.write_text(
        "anomaly-v2",
        encoding="utf-8",
    )

    second = combined_dataset_hash(
        [
            normal,
            anomaly,
        ]
    )

    assert first != second


def test_metadata_round_trip(
    tmp_path: Path,
):
    metadata = ModelMetadata(
        model_name=
            "behavior-autoencoder",
        model_version=
            "v1",
        model_type=
            "pytorch-autoencoder",
        feature_count=
            8,
        threshold=
            0.0069,
        training_samples=
            4000,
        validation_samples=
            1000,
        anomaly_test_samples=
            1000,
        training_data_sha256=
            "abc123",
    )

    path = (
        tmp_path
        / "metadata.json"
    )

    save_metadata(
        metadata,
        path,
    )

    restored = load_metadata(
        path
    )

    assert restored == metadata