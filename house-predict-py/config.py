from typing import Any

CONFIG: dict[str, Any] = {
    # Model configurations
    "test_size": 0.2,  # Keep 20% of data as test
    "random_state": 42,  # Random state
    "outlier_threshold": 3,  # How many std away from mean that we consider, outliers
    "required_cols": [
        "square_footage",
        "bedrooms",
        "price_thousands",
    ],  # Required columns must exists in CSV
    "feature_cols": ["square_footage", "bedrooms"],
    "target_col": "price_thousands",
    # Visualization configurations
    "figure_size": (10, 6),
    "train_data_point_color": "red",
    "test_data_point_color": "green",
    "line_color": "blue",
    "plane_color": "pink",
    "line_width": 1,
    "alpha": 0.7,
    "mesh_grid_size": 50,
    # File/resource configurations
    "default_csv": "resources/house_data_multi_linear_regression.csv",  # Path to the CSV data
    "output_image": "resources/housing_regression.png",  # Path to the output image
    "output_image_3d": "resources/housing_regression_3d.png",  # Path to the 3D output image
    "default_model_path": "resources/housing_model.pkl",  # Store training data to avoid training everytime app run
    "default_metadata_path": "resources/housing_model_metadata.json",
}
