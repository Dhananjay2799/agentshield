from __future__ import annotations

import hashlib
import json
from dataclasses import asdict, dataclass
from pathlib import Path


@dataclass(frozen=True)
class ModelMetadata:
    model_name: str
    model_version: str
    model_type: str
    feature_count: int
    threshold: float
    training_samples: int
    validation_samples: int
    anomaly_test_samples: int
    training_data_sha256: str


def file_sha256(
    path: Path,
) -> str:
    digest = hashlib.sha256()

    with path.open("rb") as file:
        for chunk in iter(
            lambda: file.read(1024 * 1024),
            b"",
        ):
            digest.update(chunk)

    return digest.hexdigest()


def combined_dataset_hash(
    paths: list[Path],
) -> str:
    digest = hashlib.sha256()

    for path in sorted(
        paths,
        key=lambda value: str(value),
    ):
        digest.update(
            str(path.name).encode(
                "utf-8"
            )
        )

        digest.update(
            file_sha256(path).encode(
                "utf-8"
            )
        )

    return digest.hexdigest()


def save_metadata(
    metadata: ModelMetadata,
    path: Path,
) -> None:
    path.parent.mkdir(
        parents=True,
        exist_ok=True,
    )

    path.write_text(
        json.dumps(
            asdict(metadata),
            indent=2,
        ),
        encoding="utf-8",
    )


def load_metadata(
    path: Path,
) -> ModelMetadata:
    payload = json.loads(
        path.read_text(
            encoding="utf-8"
        )
    )

    return ModelMetadata(
        **payload
    )