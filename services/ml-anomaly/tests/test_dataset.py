from pathlib import Path

from app.dataset import (
    generate_anomalous_behavior,
    generate_normal_behavior,
    write_dataset,
)


def test_generate_normal_behavior_count():
    rows = generate_normal_behavior(
        100,
        seed=1,
    )

    assert len(rows) == 100


def test_generate_anomalous_behavior_count():
    rows = generate_anomalous_behavior(
        50,
        seed=1,
    )

    assert len(rows) == 50


def test_normal_behavior_is_lower_risk():
    normal = generate_normal_behavior(
        500,
        seed=10,
    )

    anomalous = generate_anomalous_behavior(
        500,
        seed=10,
    )

    normal_risk = sum(
        row.average_risk_score
        for row in normal
    ) / len(normal)

    anomaly_risk = sum(
        row.average_risk_score
        for row in anomalous
    ) / len(anomalous)

    assert normal_risk < anomaly_risk


def test_anomalous_behavior_has_more_denials():
    normal = generate_normal_behavior(
        500,
        seed=5,
    )

    anomalous = generate_anomalous_behavior(
        500,
        seed=5,
    )

    normal_denials = sum(
        row.deny_ratio
        for row in normal
    ) / len(normal)

    anomaly_denials = sum(
        row.deny_ratio
        for row in anomalous
    ) / len(anomalous)

    assert normal_denials < anomaly_denials


def test_write_dataset(tmp_path: Path):
    rows = generate_normal_behavior(
        10,
        seed=2,
    )

    output = (
        tmp_path
        / "normal.csv"
    )

    write_dataset(
        output,
        rows,
        "normal",
    )

    assert output.exists()

    lines = output.read_text(
        encoding="utf-8"
    ).splitlines()

    assert len(lines) == 11

    assert lines[0].endswith(
        ",label"
    )