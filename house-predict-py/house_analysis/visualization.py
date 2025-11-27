import pandas as pd
import numpy as np

import matplotlib.pyplot as plt
from typing import Dict, Any

from config import CONFIG
from house_analysis.data_processing import ModelData
from house_analysis.model import ModelResult, get_model_formula


def print_result(data: ModelData, result: ModelResult) -> None:
    # Get multi-linear regression formula
    intercept, coefficients = get_model_formula(result)
    print(
        f"model formula: price = {intercept:.4f} + {coefficients[0]:.4f} x square_footage + {coefficients[1]:.4f} x bedroom"
    )

    # Print r2
    print(f"r2 result: {result.train_r2:.4f} (trained) | {result.test_r2:.4f} (test)")

    # Print RMSE
    print(
        f"rmse result: {result.train_rmse:.4f} (trained) | {result.test_rmse:.4f} (test)"
    )

    # Print some sample data set (both trained and test)
    trained_sample = pd.DataFrame(
        {
            "Square footage": data.X_trained[
                :, 0
            ],  # Get all row of column 0 (square_footage)
            "Bedrooms": data.X_trained[:, 1],  # Get all row of column 1 (bedrooms)
            "Actual price ($K)": data.y_trained,
            "Predicted price ($K)": np.round(result.train_predictions, 2),
        }
    )

    test_sample = pd.DataFrame(
        {
            "Square footage": data.X_test[:, 0],
            "Bedrooms": data.X_test[:, 1],
            "Actual price ($K)": data.y_test,
            "Predicted price ($K)": np.round(result.test_predictions, 2),
        }
    )

    print("training sample data")
    print(f"{trained_sample.head(5).to_string(index=False)}")

    print("test data sample")
    print(f"{test_sample.head(5).to_string(index=False)}")


def create_visualization_data(data: ModelData, result: ModelResult) -> Dict[str, Any]:
    # Combine both train and test sample to get the value range for features
    X_combined = np.vstack((data.X_trained, data.X_test))

    # For square_footage column
    x_min, x_max = X_combined[:, 0].min(), X_combined[:, 0].max()

    # For bedrooms column
    y_min, y_max = X_combined[:, 1].min(), X_combined[:, 1].max()

    # Create the feature range with step = 100
    feature_ranges = [np.linspace(x_min, x_max, 100), np.linspace(y_min, y_max, 100)]

    # Calculate feature mean for regression line/plane
    # Since we have 2 features here, we basically will create 2 plots, where 1 feature is variable (vary) and 1 is constant
    feature_mean = X_combined.mean(axis=0)

    # Get formula for displaying
    intercept, coefficients = get_model_formula(result)
    formula_text = f"price = {intercept:.4f} + {coefficients[0]:.4f} x square_footage + {coefficients[1]:.4f} x bedroom"

    # Create mesh grid for 3D visualization
    x_range = np.linspace(x_min, x_max, CONFIG["mesh_grid_size"])
    y_range = np.linspace(y_min, y_max, CONFIG["mesh_grid_size"])
    xx, yy = np.meshgrid(x_range, y_range)

    # Prepare grid point for predictions
    grid_points = np.c_[xx.ravel(), yy.ravel()]

    # Scale grid point using the same scaler
    grid_points_scaled = result.scaler.transform(grid_points)

    # Make predictions
    z_pred = result.model.predict(grid_points_scaled)
    zz = z_pred.reshape(xx.shape)

    return {
        "feature_ranges": feature_ranges,
        "feature_mean": feature_mean,
        "formula_text": formula_text,
        "xx": xx,
        "yy": yy,
        "zz": zz,
    }


def create_2d_visualization(
    model_data: ModelData,
    model_result: ModelResult,
    vis_data: Dict[str, Any],
    output: str,
) -> None:
    # Create a figure with 2 side-by-side plots (1 row, 2 cols), one for each feature
    _, axes = plt.subplots(1, 2, figsize=CONFIG["figure_size"])

    # Create plot for each feature
    for i, feature in enumerate(CONFIG["feature_cols"]):
        ax = axes[i]

        # Plot training data point
        ax.scatter(
            model_data.X_trained[:, i],  # X coordinates
            model_data.y_trained,  # Y coordinates
            color=CONFIG["train_data_point_color"],
            alpha=CONFIG["alpha"],
            label="Training data",
        )

        # Plot test data points
        ax.scatter(
            model_data.X_test[:, i],
            model_data.y_test,
            color=CONFIG["test_data_point_color"],
            alpha=CONFIG["alpha"],
            label="Test data",
        )

        # Add regresion line
        feature_ranges = vis_data["feature_ranges"]
        feature_mean = vis_data["feature_mean"]
        if i == 0:  # If feature is square_footage
            line_x = np.c_[
                feature_ranges[0],
                np.full(feature_ranges[0].shape, feature_mean[1]),
            ]
        else:  # If feature is bedroom
            line_x = np.c_[
                np.full(feature_ranges[1].shape, feature_mean[0]),
                feature_ranges[1],
            ]

        # Scale the line point and predict price
        line_x_scaled = model_result.scaler.transform(line_x)
        line_y = model_result.model.predict(line_x_scaled)

        # Plot the regression line
        ax.plot(
            feature_ranges[i],
            line_y,
            color=CONFIG["line_color"],
            linewidth=CONFIG["line_width"],
            label="Regression line",
        )

        # Add labels and title
        ax.set_xlabel(feature.replace("_", " ").title())
        ax.set_ylabel("Price ($K)")
        ax.set_title(
            f"Price vs {feature.replace('_', ' ').title()} with regression line"
        )
        ax.legend()
        ax.grid(True, alpha=CONFIG["alpha"])

    # Add overall title
    plt.suptitle("Multi linear regression: housing price vs features")
    plt.tight_layout(rect=(0.0, 0.0, 1.0, 0.95))
    plt.savefig(output)
    plt.close()


def create_3d_visualization(
    model_data: ModelData,
    vis_data: Dict[str, Any],
    output: str,
) -> None:
    fig = plt.figure(figsize=CONFIG["figure_size"])
    ax = fig.add_subplot(111, projection="3d")

    # Set initial view angle
    ax.view_init(elev=30, azim=45)

    ax.scatter(
        model_data.X_trained[:, 0],  # X coordinates
        model_data.X_trained[:, 1],
        model_data.y_trained,  # type: ignore
        color=CONFIG["train_data_point_color"],
        alpha=CONFIG["alpha"],
        label="Training data",
    )

    # Plot test data points
    ax.scatter(
        model_data.X_test[:, 0],
        model_data.X_test[:, 1],
        model_data.y_test,  # type: ignore
        color=CONFIG["test_data_point_color"],
        alpha=CONFIG["alpha"],
        label="Test data",
    )

    # Plot the regression plane
    ax.plot_surface(
        vis_data["xx"],
        vis_data["yy"],
        vis_data["zz"],
        alpha=CONFIG["alpha"],
        color=CONFIG["plane_color"],
        rstride=2,
        cstride=2,
    )

    # Add titles and label
    ax.set_xlabel("Square footage")
    ax.set_ylabel("Bedrrooms")
    ax.set_zlabel("Price ($K)")
    ax.set_title("Multi linear regression 3D visualization: Price vs House features")
    ax.legend()

    # Add formula as text
    plt.figtext(0.1, 0.01, vis_data["formula_text"], fontsize=12)

    # Save plot
    plt.savefig(output, bbox_inches="tight")
    plt.close()
