FROM alpine:latest

RUN apk --no-cache add ca-certificates bash

WORKDIR /root/

COPY build/manager .

RUN chmod +x /manager

RUN mkdir -p /config

# Copy the entrypoint script
COPY scripts/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

EXPOSE 8080 8443 8081

# Run the entrypoint script which will keep the container running
CMD ["/entrypoint.sh"]