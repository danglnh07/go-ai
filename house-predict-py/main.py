import argparse
import sys
import traceback
from config import CONFIG
from house_analysis.data_processing import (
    load_data,
    prepare_model_data,
    preprocess_data,
)
from house_analysis.exceptions import DataProcessingError, ModelOperationError
from house_analysis.logging_config import logger
from house_analysis.model import evaluate_model, load_model, save_model, train_model
from house_analysis.visualization import (
    create_2d_visualization,
    create_3d_visualization,
    create_visualization_data,
    print_result,
)


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Multi linear regression analysys on house pricing"
    )

    parser.add_argument(
        "-i",
        "--input",
        type=str,
        default=CONFIG["default_csv"],
        help=f"path to CSV data file (default: {CONFIG['default_csv']})",
    )

    parser.add_argument(
        "--no-plot",
        action="store_true",
        help="not creating visualization when running",
    )

    parser.add_argument(
        "--save-model",
        action="store_true",
        help="save trained model to file system",
    )

    parser.add_argument(
        "--model-path",
        type=str,
        default=CONFIG["default_model_path"],
        help=f"path to load/save model (default: {CONFIG['default_model_path']})",
    )

    parser.add_argument(
        "--load-model",
        action="store_true",
        help="load a previously trained model",
    )

    parser.add_argument(
        "--metadata-path",
        type=str,
        default=CONFIG["default_metadata_path"],
        help=f"path to load/save model's metadata that has been pretrained (default: {CONFIG['default_metadata_path']})",
    )

    parser.add_argument(
        "--predict-only",
        action="store_true",
        help="only make prediction based on pretrained model",
    )

    return parser.parse_args()


def main() -> int:
    try:
        # Parse command line arguments
        args = parse_arguments()

        # Choose action based on command line arguments
        if args.load_model:
            # Load model and metadata from file
            model, scaler, _ = load_model(args.model_path, args.metadata_path)
            logger.info("model load success")

        # If not load pretrained model, then we assume that we want to train a new one
        else:
            # Prepare data
            df = load_data(args.input)
            processed_df = preprocess_data(df)
            data = prepare_model_data(processed_df)

            # Train model
            model, scaler = train_model(data)

            # Evaludate model
            result = evaluate_model(data, model, scaler)

            # Print the result into the terminal
            print_result(data, result)

            # If provide save option, save model
            if args.save_model:
                save_model(result, args.model_path, args.metadata_path)

            # If want to create visualization
            if not args.no_plot:
                vis_data = create_visualization_data(data, result)

                # Create 2D visualization
                output_2d = CONFIG["output_image"]
                logger.info(f"create 2D visualization at: {output_2d}")
                create_2d_visualization(data, result, vis_data, output_2d)

                # Create 3D visualization
                output_3d = CONFIG["output_image_3d"]
                logger.info(f"create 3D visualization at: {output_3d}")
                create_3d_visualization(data, vis_data, output_3d)

        return 0
    except DataProcessingError as e:
        err = f"failed to process data: {str(e)}"
        logger.error(err)
        traceback.print_exc()
        return 1
    except ModelOperationError as e:
        err = f"model failed to run: {str(e)}"
        logger.error(err)
        traceback.print_exc()
        return 1
    except Exception as e:
        err = f"unexpected error: {str(e)}"
        logger.error(err)
        traceback.print_exc()
        return 1


if __name__ == "__main__":
    sys.exit(main())
