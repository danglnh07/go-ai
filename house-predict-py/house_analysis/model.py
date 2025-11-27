import json
import os
import pickle
from dataclasses import dataclass
from typing import Any, Dict, List, Tuple

import numpy as np
from sklearn.linear_model import LinearRegression  # type: ignore
from sklearn.metrics import r2_score, mean_squared_error  # type: ignore
from sklearn.preprocessing import StandardScaler  # type: ignore

from config import CONFIG
from house_analysis.logging_config import logger
from house_analysis.exceptions import ModelOperationError
from house_analysis.data_processing import ModelData


@dataclass
class ModelResult:
    model: LinearRegression
    scaler: StandardScaler
    train_predictions: np.ndarray
    test_predictions: np.ndarray
    train_r2: float
    test_r2: float
    train_rmse: float
    test_rmse: float


def train_model(data: ModelData) -> Tuple[LinearRegression, StandardScaler]:
    # Scale the feature set
    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(data.X_trained)

    # Train model with trained data set
    model = LinearRegression()
    model.fit(X_scaled, data.y_trained)

    return model, scaler


def evaluate_model(
    data: ModelData,
    model: LinearRegression,
    scaler: StandardScaler,
) -> ModelResult:
    # Transform the feature sets
    X_trained_scaled = scaler.transform(data.X_trained)
    X_test_scaled = scaler.transform(data.X_test)

    # Predict
    trained_predictions = model.predict(X_trained_scaled)
    test_predictions = model.predict(X_test_scaled)

    # Calculate R2 (coefficient of determination)
    train_r2 = r2_score(data.y_trained, trained_predictions)
    test_r2 = r2_score(data.y_test, test_predictions)

    # Calculate RMSE (root square mean error)
    train_rmse = np.sqrt(mean_squared_error(data.y_trained, trained_predictions))
    test_rmse = np.sqrt(mean_squared_error(data.y_test, test_predictions))

    return ModelResult(
        model,
        scaler,
        trained_predictions,
        test_predictions,
        train_r2,
        test_r2,
        train_rmse,
        test_rmse,
    )


def save_model(result: ModelResult, model_path: str, metadata_path: str) -> None:
    try:
        # Create directories if not exists
        model_dir = os.path.dirname(model_path)
        if model_dir and not os.path.exists(model_dir):
            os.makedirs(model_dir)

        metadata_dir = os.path.dirname(metadata_path)
        if metadata_dir and not os.path.exists(metadata_dir):
            os.makedirs(metadata_dir)

        # Save model and scaler
        with open(model_path, "wb") as f:  # open file with write-binary permission
            model_components = {
                "model": result.model,
                "scaler": result.scaler,
            }
            pickle.dump(model_components, f)

        # Save metadata
        intercept, coefficients = get_model_formula(result)

        metadata = {
            # "coefficients": coefficients,
            "coefficients": [float(c) for c in coefficients],
            "intercept": float(intercept),
            "feature": CONFIG["feature_cols"],
            "target": CONFIG["target_col"],
            "train_r2": float(result.train_r2),
            "test_r2": float(result.test_r2),
            "train_rmse": float(result.train_rmse),
            "test_rmse": float(result.test_rmse),
        }

        with open(metadata_path, "w") as f:
            json.dump(metadata, f, indent=4)

    except Exception as e:
        err = f"error saving model: {str(e)}"
        logger.error(err)
        raise ModelOperationError(err)


# Get multi linear regression formula
def get_model_formula(result: ModelResult) -> Tuple[float, List[float]]:
    # Since this is multi linear regression, its formula should be:
    # y = a1x1 + a2x2 + ... + anxn + b
    # So coefficient would be a list, while intercept, can be sum up into a single number
    model = result.model
    scaler = result.scaler

    # Assert that scale_ and mean_ are not None (they're set after fit_transform)
    assert scaler.scale_ is not None, "Scaler must be fitted"
    assert scaler.mean_ is not None, "Scaler must be fitted"

    # Calculate coefficients
    coefficients: List[float] = []
    for i in range(len(model.coef_)):
        # We need to divide to the scaler since we normalize the feature set when training
        coefficients.append(model.coef_[i] / scaler.scale_[i])

    # Calculate intercept
    intercept = float(model.intercept_) - sum(
        model.coef_[i] * scaler.mean_ / scaler.scale_
    )

    return intercept, coefficients


def load_model(
    model_path: str,
    metadata_path: str,
) -> Tuple[LinearRegression, StandardScaler, Dict[str, Any]]:
    try:
        # Load model data
        if not os.path.isfile(model_path):
            err = f"model path must be a file and exist: {model_path}"
            logger.error(err)
            raise ModelOperationError(err)

        with open(model_path, "rb") as f:  # read file as binary permission
            model_components = pickle.load(f)

        model = model_components["model"]
        scaler = model_components["scaler"]

        # Load metedata
        metadata = {}

        if not os.path.isfile(metadata_path):
            err = f"metadata path must be a file and exists: {metadata_path}"
            logger.error(err)
            raise ModelOperationError(err)

        with open(metadata_path, "r") as f:
            metadata = json.load(f)

        return model, scaler, metadata
    except Exception as e:
        err = f"failed to load model: {str(e)}"
        logger.error(err)
        raise ModelOperationError(err)
