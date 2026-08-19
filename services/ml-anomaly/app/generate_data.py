from pathlib import Path

from app.dataset import (
    generate_anomalous_behavior,
    generate_normal_behavior,
    write_dataset,
)


def main() -> None:
    data_dir = Path(
        "data"
    )

    normal_rows = (
        generate_normal_behavior(
            5000,
            seed=42,
        )
    )

    anomaly_rows = (
        generate_anomalous_behavior(
            1000,
            seed=1337,
        )
    )

    write_dataset(
        data_dir / "normal.csv",
        normal_rows,
        "normal",
    )

    write_dataset(
        data_dir / "anomalies.csv",
        anomaly_rows,
        "anomaly",
    )

    print(
        "Generated "
        f"{len(normal_rows)} normal "
        f"and {len(anomaly_rows)} anomaly samples."
    )


if __name__ == "__main__":
    main()