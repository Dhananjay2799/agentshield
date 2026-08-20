from app.drift import DriftDetector


def test_empty_detector_is_not_drifting():
    detector = DriftDetector()

    snapshot = detector.snapshot()

    assert snapshot.sample_count == 0
    assert snapshot.is_drifting is False


def test_detector_requires_minimum_samples():
    detector = DriftDetector(
        window_size=20,
        minimum_samples=5,
        score_ratio_threshold=1.5,
        anomaly_rate_threshold=0.25,
    )

    for _ in range(4):
        snapshot = detector.record(
            score_ratio=5.0,
            is_anomaly=True,
        )

    assert snapshot.sample_count == 4
    assert snapshot.is_drifting is False


def test_normal_behavior_does_not_trigger_drift():
    detector = DriftDetector(
        window_size=20,
        minimum_samples=5,
    )

    for _ in range(10):
        snapshot = detector.record(
            score_ratio=0.25,
            is_anomaly=False,
        )

    assert snapshot.sample_count == 10
    assert snapshot.is_drifting is False
    assert snapshot.drift_score < 1.0


def test_high_reconstruction_ratios_trigger_drift():
    detector = DriftDetector(
        window_size=20,
        minimum_samples=5,
        score_ratio_threshold=1.5,
        anomaly_rate_threshold=0.50,
    )

    for _ in range(5):
        snapshot = detector.record(
            score_ratio=2.0,
            is_anomaly=False,
        )

    assert snapshot.is_drifting is True
    assert snapshot.drift_score >= 1.0


def test_high_anomaly_rate_triggers_drift():
    detector = DriftDetector(
        window_size=20,
        minimum_samples=5,
        score_ratio_threshold=10.0,
        anomaly_rate_threshold=0.25,
    )

    observations = [
        True,
        True,
        False,
        False,
        True,
    ]

    for value in observations:
        snapshot = detector.record(
            score_ratio=0.2,
            is_anomaly=value,
        )

    assert snapshot.anomaly_rate == 0.6
    assert snapshot.is_drifting is True


def test_window_discards_old_samples():
    detector = DriftDetector(
        window_size=5,
        minimum_samples=5,
    )

    for _ in range(5):
        detector.record(
            score_ratio=5.0,
            is_anomaly=True,
        )

    for _ in range(5):
        snapshot = detector.record(
            score_ratio=0.1,
            is_anomaly=False,
        )

    assert snapshot.sample_count == 5
    assert snapshot.anomaly_rate == 0.0
    assert snapshot.mean_score_ratio == 0.1
    assert snapshot.is_drifting is False