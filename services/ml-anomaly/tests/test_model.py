import torch

from app.features import FEATURE_COUNT
from app.model import (
    BehaviorAutoencoder,
    reconstruction_error,
)


def test_autoencoder_output_shape():
    model = BehaviorAutoencoder()

    batch = torch.rand(
        4,
        FEATURE_COUNT,
    )

    output = model(batch)

    assert output.shape == (
        4,
        FEATURE_COUNT,
    )


def test_autoencoder_output_is_bounded():
    model = BehaviorAutoencoder()

    batch = torch.rand(
        4,
        FEATURE_COUNT,
    )

    output = model(batch)

    assert torch.all(output >= 0)
    assert torch.all(output <= 1)


def test_reconstruction_error_returns_one_score_per_sample():
    model = BehaviorAutoencoder()

    batch = torch.rand(
        5,
        FEATURE_COUNT,
    )

    errors = reconstruction_error(
        model,
        batch,
    )

    assert errors.shape == (5,)

    assert torch.all(
        errors >= 0
    )


def test_model_uses_three_dimensional_latent_space():
    model = BehaviorAutoencoder(
        latent_dim=3
    )

    batch = torch.rand(
        2,
        FEATURE_COUNT,
    )

    encoded = model.encoder(batch)

    assert encoded.shape == (
        2,
        3,
    )