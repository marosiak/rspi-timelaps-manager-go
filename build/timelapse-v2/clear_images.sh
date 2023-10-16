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

# Calculate the total size of files older than 10 minutes and save it to a temp file
find "$OUTPUT_DIR" -type f -mmin +10 -exec du -ch {} + > temp_size.txt

# Sum all 'total' sizes from the temp file
TOTAL_SIZE=$(awk '/total$/ {gsub(/[A-Za-z]/, "", $1); total += $1} END {print total "G"}' temp_size.txt)

# Remove the temp file
rm temp_size.txt

# Display the total size of files to be deleted
echo "Total size of files to be deleted: $TOTAL_SIZE"

# Ask for confirmation with the directory path displayed
read -p "Are you sure you want to delete files from the directory $OUTPUT_DIR? (yes/no): " CONFIRM
if [[ $CONFIRM != "yes" ]] && [[ $CONFIRM != "y" ]]; then
    echo "Operation aborted."
    exit 1
fi

# Delete all but the 10 newest files in the OUTPUT_DIR directory
find "$OUTPUT_DIR" -type f -printf "%T@ %p\n" | sort -n | cut -d ' ' -f 2- | head -n -10 | xargs -d '\n' rm -f

echo "Operation completed successfully!"
