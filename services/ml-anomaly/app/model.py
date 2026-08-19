from __future__ import annotations

import torch
from torch import nn

from app.features import FEATURE_COUNT


class BehaviorAutoencoder(nn.Module):
    def __init__(
        self,
        input_dim: int = FEATURE_COUNT,
        latent_dim: int = 3,
    ) -> None:
        super().__init__()

        self.input_dim = input_dim
        self.latent_dim = latent_dim

        self.encoder = nn.Sequential(
            nn.Linear(input_dim, 6),
            nn.ReLU(),
            nn.Linear(6, latent_dim),
            nn.ReLU(),
        )

        self.decoder = nn.Sequential(
            nn.Linear(latent_dim, 6),
            nn.ReLU(),
            nn.Linear(6, input_dim),
            nn.Sigmoid(),
        )

    def forward(
        self,
        x: torch.Tensor,
    ) -> torch.Tensor:
        encoded = self.encoder(x)
        return self.decoder(encoded)


def reconstruction_error(
    model: BehaviorAutoencoder,
    features: torch.Tensor,
) -> torch.Tensor:
    reconstruction = model(features)

    return torch.mean(
        torch.square(
            features - reconstruction
        ),
        dim=-1,
    )