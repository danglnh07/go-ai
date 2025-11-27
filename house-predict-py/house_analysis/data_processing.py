import os
from dataclasses import dataclass
import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split  # type: ignore
from sklearn.linear_model import LinearRegression  # type: ignore
from sklearn.preprocessing import StandardScaler  # type: ignore

from config import CONFIG
from house_analysis.logging_config import logger
from house_analysis.exceptions import DataProcessingError, ModelOperationError


@dataclass
class ModelData:
    X_trained: np.ndarray
    X_test: np.ndarray
    y_trained: np.ndarray
    y_test: np.ndarray


# Load data from CSV
def load_data(filepath: str) -> pd.DataFrame:
    # Check if file exists
    if not os.path.isfile(filepath):
        err = f"file does not exists: {filepath}"
        logger.error(err)
        raise DataProcessingError(err)

    # Try open file
    try:
        logger.info(f"start loading data from {filepath}")
        df: pd.DataFrame = pd.read_csv(filepath)

        # Validation: check if all the required columns exits in the CSV file
        missing_cols = set(CONFIG["required_cols"]).difference(df.columns)
        if missing_cols:
            err = f"columns missing, required: {', '.join(missing_cols)}"
            logger.error(err)
            raise DataProcessingError(err)

        return df
    except Exception as e:
        err = f"error loading data: {str(e)}"
        logger.error(err)
        raise DataProcessingError(err)


# Preprocess data
def preprocess_data(df: pd.DataFrame) -> pd.DataFrame:
    logger.info("start preprocessing data")

    # Copy data frame, so that even when we failed midway, it wouldn't corrupt the original data frame
    processed_df: pd.DataFrame = df.copy()

    # Ensure data is numeric
    for col in CONFIG["required_cols"]:
        processed_df[col] = pd.to_numeric(processed_df[col], errors="coerce")

    # Handle missing data.
    # Calling to isna().any() will check for each column, is there any missing values
    # The second any() will check that, for all columns, is there any missing values
    if processed_df[CONFIG["required_cols"]].isna().any().any():
        logger.warning("missing values, dropping row")
        processed_df = processed_df.dropna(subset=CONFIG["required_cols"])

    # Filter outliers: take any values that is too small or too large
    for col in CONFIG["required_cols"]:
        # Calculate mean
        mean: float = processed_df[col].mean()

        # Calculate standard deviation (std)
        standard_deviation: float = processed_df[col].std()

        # Get threshold from config.
        threshold = CONFIG["outlier_threshold"]

        # Define the lower and upper bound of the valid data
        lower_bound: float = mean - threshold * standard_deviation
        upper_bound: float = mean + threshold * standard_deviation

        # Get the outlier set and drop them out of the data frame
        outliers = (processed_df[col] < lower_bound) | (processed_df[col] > upper_bound)
        if outliers.any():
            logger.warning(f"found {outliers.sum()} outliers, drop data")
            processed_df = processed_df[~outliers]

    return processed_df


# Prepare data for model training
def prepare_model_data(df: pd.DataFrame) -> ModelData:
    # Get X (features) and y (prediction).
    X = df[CONFIG["feature_cols"]].to_numpy()
    y = df[CONFIG["target_col"]].to_numpy()

    # Splitting the data set into trained set and test set
    X_trained: np.ndarray
    X_test: np.ndarray
    y_trained: np.ndarray
    y_test: np.ndarray
    X_trained, X_test, y_trained, y_test = train_test_split(
        X,
        y,
        test_size=CONFIG["test_size"],
        random_state=CONFIG["random_state"],
    )

    logger.info(
        f"splitting data set: {len(X_trained)} training samples, {len(X_test)} test samples"
    )

    return ModelData(
        X_trained=X_trained,
        X_test=X_test,
        y_trained=y_trained,
        y_test=y_test,
    )


def make_predictions(
    df: pd.DataFrame,
    model: LinearRegression,
    scaler: StandardScaler,
) -> np.ndarray:
    try:
        # Scale the feature columns
        X_scaled = scaler.transform(df[CONFIG["required_cols"]].to_numpy())

        # Predict using the scaled features
        return model.predict(X_scaled)
    except Exception as e:
        err = f"failed to make prediction: {str(e)}"
        logger.error(err)
        raise ModelOperationError(err)
