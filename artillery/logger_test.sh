#!/bin/bash

# Load environment variables from .env file
if [ -f ../.env ]; then
    export $(grep -v '^#' ../.env | xargs)
fi

# Check if ARTILLERY_API_KEY is set
if [ -z "$ARTILLERY_API_KEY" ]; then
    echo "Error: ARTILLERY_API_KEY is not set in the .env file."
    exit 1
fi

# Run Artillery test
artillery run booking-test.yml \
    --record \
    --key "$ARTILLERY_API_KEY"
