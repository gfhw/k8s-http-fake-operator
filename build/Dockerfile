FROM alpine:latest

RUN apk --no-cache add ca-certificates bash

WORKDIR /root/

# Debug: list all files in build context root
RUN echo "=== Listing / ===" && ls -la /
RUN echo "=== Listing current dir ===" && ls -la .
RUN echo "=== Listing build/ ===" && ls -la build/ 2>/dev/null || echo "build/ not found"
RUN echo "=== Find manager file ===" && find / -name "manager" -type f 2>/dev/null || echo "manager not found"

COPY build/manager .

RUN chmod +x /manager

RUN mkdir -p /config

# Copy the entrypoint script
COPY scripts/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

EXPOSE 8080 8443 8081

# Run the entrypoint script which will keep the container running
CMD ["/entrypoint.sh"]
