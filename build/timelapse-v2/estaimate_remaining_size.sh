#!/bin/bash

# Load the .env file
source .env

# Check if OUTPUT_DIR and DELAY are defined
if [[ -z $OUTPUT_DIR ]] || [[ -z $DELAY ]]; then
    echo "Error: OUTPUT_DIR or DELAY is not defined in .env file."
    exit 1
fi

# Make sure OUTPUT_DIR exists
if [[ ! -d $OUTPUT_DIR ]]; then
    echo "Error: Directory $OUTPUT_DIR doesn't exist."
    exit 1
fi

# Calculate the average size of the 20 latest files and inflate it by 20%
avg_size_bytes=$(find "$OUTPUT_DIR" -type f -printf "%s\n" | sort -rn | head -n 20 | awk '{total += $1; count++} END {print int((total/count)*1.2)}')
avg_size_MB=$(awk -v bytes="$avg_size_bytes" 'BEGIN {print bytes / (1024*1024)}' | awk '{printf "%.2f MB", $1}')

# Get available space in bytes
available_space=$(df -B1 "$OUTPUT_DIR" | awk 'NR==2 {print $4}')

# Reserve 3GB
reserve_space=$((3 * 1024 * 1024 * 1024))

# Calculate files to fit, handle large numbers
files_to_fit=$(awk -v available="$available_space" -v reserve="$reserve_space" -v avg="$avg_size_bytes" 'BEGIN {print int((available - reserve) / avg)}')

# Convert DELAY to minutes
delay_mins=$(echo $DELAY | sed 's/m//')

# Calculate total time available in minutes
total_mins=$((files_to_fit * delay_mins))

# Calculate time in different units
w=$((total_mins / (60 * 24 * 7)))
d=$(( (total_mins % (60 * 24 * 7)) / (60 * 24) ))
h=$(( (total_mins % (60 * 24)) / 60 ))
m=$((total_mins % 60))

# Generate the time string
time_str=""
[[ $w -gt 0 ]] && time_str+="$w weeks "
[[ $d -gt 0 ]] && time_str+="$d days "
[[ $h -gt 0 ]] && time_str+="$h hours "
[[ $m -gt 0 ]] && time_str+="$m minutes"

# Print results
echo "Average size of last 20 files (increased by 20%): $avg_size_MB"
echo "Available space: $((available_space / (1024*1024))) MB"
echo "Files that can fit with 3GB reserved: $files_to_fit"
echo -n "Time for which space is available: "
[[ -n "$time_str" ]] && echo "$time_str" || echo "No time left"
