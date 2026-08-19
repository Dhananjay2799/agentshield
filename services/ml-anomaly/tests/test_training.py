import torch

from app.model import (
    BehaviorAutoencoder,
)
from app.train import (
    TrainingConfig,
    calculate_threshold,
    split_normal_data,
    train_model,
)


def test_split_normal_data_preserves_count():
    tensor = torch.rand(
        100,
        8,
    )

    train_tensor, validation_tensor = (
        split_normal_data(
            tensor,
            validation_ratio=0.20,
            seed=42,
        )
    )

    assert train_tensor.shape[0] == 80
    assert validation_tensor.shape[0] == 20


def test_calculate_threshold_is_positive():
    model = BehaviorAutoencoder()

    tensor = torch.rand(
        100,
        8,
    )

    threshold = calculate_threshold(
        model,
        tensor,
        0.99,
    )

    assert threshold >= 0


def test_training_changes_model_parameters():
    torch.manual_seed(42)

    model = BehaviorAutoencoder()

    before = [
        parameter.detach().clone()
        for parameter
        in model.parameters()
    ]

    training_tensor = torch.rand(
        100,
        8,
    )

    validation_tensor = torch.rand(
        20,
        8,
    )

    config = TrainingConfig(
        epochs=2,
        batch_size=16,
    )

    train_model(
        model,
        training_tensor,
        validation_tensor,
        config,
    )

    after = list(
        model.parameters()
    )

    assert any(
        not torch.equal(
            old,
            new,
        )
        for old, new
        in zip(
            before,
            after,
        )
    )