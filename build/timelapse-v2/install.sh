#!/bin/bash
# Path definitions
LOCAL_DIR="$(pwd)" 
BINARY_DIR="$LOCAL_DIR/bin"
BINARY_NAME="timelapse_v2"
SERVICE_NAME="timelapse.service"
SYSTEM_BIN_DIR="/usr/local/bin"
SYSTEM_SERVICE_DIR="/etc/systemd/system"
ENV_FILE=".env"
PHOTOS_DIR="$LOCAL_DIR/photos/"
WEB_CLIENT_DIR="$LOCAL_DIR/web_client"

# Check if the binary exists in the local directory
if [[ ! -f "$BINARY_DIR/$BINARY_NAME" ]]; then
    echo "Error: Binary file $BINARY_NAME not found in directory $LOCAL_DIR."
    exit 1
fi

# Check if the .env file exists
if [[ ! -f "$LOCAL_DIR/$ENV_FILE" ]]; then
    echo "Error: File $ENV_FILE not found in directory $LOCAL_DIR."
    exit 1
fi

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root."
   exit 1
fi

# Check if service is running and stop it
SERVICE_PID=$(pgrep -f "$BINARY_NAME")
if [ ! -z "$SERVICE_PID" ]; then
  echo "Stopping running service..."
  sudo systemctl stop $SERVICE_NAME
fi

# Kill the process if it's still running after stopping the service
if [ ! -z "$SERVICE_PID" ]; then
  echo "Killing running process..."
  sudo kill -9 $SERVICE_PID
fi

# Create photos directory if it doesn't exist
if [[ ! -d "$PHOTOS_DIR" ]]; then
    mkdir "$PHOTOS_DIR"
fi

# Update the .env file to set the photos directory and web client directory
awk -v photos_dir="$PHOTOS_DIR" -v web_client_dir="$WEB_CLIENT_DIR" '
BEGIN { output_updated = 0; web_updated = 0 }
/OUTPUT_DIR/ {
    if ($0 ~ /WILL_BE_OVERRITEN_BY_INSTALL_SCRIPT/ || $2 == "") {
        $0 = "OUTPUT_DIR=" photos_dir
        output_updated = 1
    }
}
/WEB_INTERFACE_FILES_PATH/ {
    if ($2 == "") {
        $0 = "WEB_INTERFACE_FILES_PATH=" web_client_dir
        web_updated = 1
    }
}
{ print }
END {
    if (!output_updated) {
        print "OUTPUT_DIR=" photos_dir
    }
    if (!web_updated) {
        print "WEB_INTERFACE_FILES_PATH=" web_client_dir
    }
}' "$LOCAL_DIR/$ENV_FILE" > "$LOCAL_DIR/$ENV_FILE.tmp" && mv "$LOCAL_DIR/$ENV_FILE.tmp" "$LOCAL_DIR/$ENV_FILE"

# Copy the binary to the system directory
sudo cp "$BINARY_DIR/$BINARY_NAME" "$SYSTEM_BIN_DIR/"
sudo chmod +x "$SYSTEM_BIN_DIR/$BINARY_NAME"

# Create .service file for systemd
cat <<EOL | sudo tee "$SYSTEM_SERVICE_DIR/$SERVICE_NAME" > /dev/null
[Unit]
Description=Timelapse Service
After=network.target

[Service]
EnvironmentFile=$LOCAL_DIR/$ENV_FILE
ExecStart=$SYSTEM_BIN_DIR/$BINARY_NAME
User=root
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOL

# Reload changes and start the service
sudo systemctl daemon-reload
sudo systemctl enable $SERVICE_NAME
sudo systemctl restart $SERVICE_NAME

echo "Operation completed successfully!"
