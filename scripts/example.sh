#!/bin/bash

echo '{"message": "Hello from script!", "timestamp": '$(date +%s)', "env_var": "'$CUSTOM_VAR'"}'
echo '{"Content-Type": "application/json"}'
echo '200'