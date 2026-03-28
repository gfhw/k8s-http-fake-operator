#!/bin/bash

DELAY_SECONDS=${1:-5}

echo "Sleeping for $DELAY_SECONDS seconds..."
sleep $DELAY_SECONDS

echo '{"message": "Delayed response", "delay": "'$DELAY_SECONDS'", "timestamp": '$(date +%s)'}'
echo '{"Content-Type": "application/json"}'
echo '200'