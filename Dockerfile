FROM golang:1.26.1-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache git ca-certificates

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /workspace/manager .

COPY --from=builder /workspace/config/crd/bases /config/crd/bases

EXPOSE 8080 8443 8081

ENTRYPOINT ["/manager"]