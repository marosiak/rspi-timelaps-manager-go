#!/bin/bash

# Load the .env file
source .env

# Check if the OUTPUT_DIR variable is defined
if [[ -z $OUTPUT_DIR ]]; then
    echo "Error: OUTPUT_DIR variable not defined in the .env file."
    exit 1
fi

# Check if the OUTPUT_DIR directory exists
if [[ ! -d $OUTPUT_DIR ]]; then
    echo "Error: Directory $OUTPUT_DIR does not exist."
    exit 1
fi

# Ask for confirmation with the directory path displayed
read -p "Are you sure you want to delete files from the directory $OUTPUT_DIR? (yes/no): " CONFIRM
if [[ $CONFIRM != "yes" ]]; then
    echo "Operation aborted."
    exit 1
fi

# Delete files older than 10 minutes from the OUTPUT_DIR directory
find "$OUTPUT_DIR" -type f -mmin +10 -exec rm -f {} +

echo "Operation completed successfully!"
