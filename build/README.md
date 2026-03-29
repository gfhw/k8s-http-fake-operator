# Build Scripts

This directory contains scripts to build the k8s-http-fake-operator Docker image.

## Prerequisites

Before building, ensure you have:

1. **Docker** installed and running
2. **Go** installed (for building the binary)
3. **ctr** command available (for importing to containerd/K3s)

## Building the Binary

The `manager` binary needs to be cross-compiled for Linux amd64 before building the Docker image.

### On Windows (PowerShell):

```powershell
# Cross-compile for Linux amd64
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o build/manager cmd/main.go
```

### On Linux/macOS:

```bash
# Cross-compile for Linux amd64
GOOS=linux GOARCH=amd64 go build -o build/manager cmd/main.go
```

The binary should be placed in the `build/` directory as `manager`.

## Available Scripts

### Linux/macOS: `build-image.sh`

Bash script for building Docker images and automatically importing to containerd.

#### Usage

```bash
# Make the script executable (first time only)
chmod +x build-image.sh

# Build with default settings (uses existing manager binary)
./build-image.sh

# Build with custom name and tag
./build-image.sh --name my-operator --tag v1.0.0

# Build without cache
./build-image.sh --no-cache

# Build but don't import to containerd
./build-image.sh --no-import

# Show help
./build-image.sh --help
```

#### What the Script Does

1. **Checks prerequisites**: Verifies Docker and the `manager` binary exist
2. **Builds Docker image**: Creates the image with your binary
3. **Saves image**: Exports the image to a compressed tar file
4. **Imports to containerd**: Automatically imports the image to containerd (for K3s/K8s)

#### Environment Variables

- `IMAGE_NAME`: Docker image name (default: k8s-http-fake-operator)
- `IMAGE_TAG`: Docker image tag (default: latest)
- `DOCKERFILE`: Path to Dockerfile (default: build/Dockerfile)
- `BUILD_CONTEXT`: Build context directory (default: ..)
- `NO_CACHE`: Build without using cache (true/false)
- `IMPORT_TO_CONTAINERD`: Import image to containerd after build (default: true)

Example:

```bash
export IMAGE_NAME=my-operator
export IMAGE_TAG=v1.0.0
./build-image.sh
```

### Windows PowerShell: `build-image.ps1`

PowerShell script for building Docker images on Windows.

```powershell
# Build with default settings
.\build-image.ps1

# Build with custom name and tag
.\build-image.ps1 -ImageName my-operator -ImageTag v1.0.0

# Build without cache
.\build-image.ps1 -NoCache

# Show help
.\build-image.ps1 -Help
```

## Complete Build Workflow

### For K3s/Kubernetes Deployment

1. **Build the binary** (on Windows):
   ```powershell
   cd D:\files\operator\k8s-http-fake-operator
   $env:GOOS="linux"; $env:GOARCH="amd64"; go build -o build/manager cmd/main.go
   ```

2. **Build and import the image** (on Linux with K3s):
   ```bash
   cd /home/hwk/file/k8s-http-fake-operator/build
   ./build-image.sh
   ```

3. **Deploy with Helm**:
   ```bash
   cd /home/hwk/file/k8s-http-fake-operator
   helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator
   ```

## Output Files

After building, the script generates:

- **Docker image**: `k8s-http-fake-operator:latest` (in Docker)
- **Compressed tar**: `k8s-http-fake-operator-latest.tar.gz` (in build/)
- **containerd image**: Available in `k8s.io` namespace (if imported)

## Troubleshooting

### ImagePullBackOff Error

If you see `ErrImagePull` or `ImagePullBackOff`, the image may not be in containerd:

```bash
# Check if image exists in containerd
sudo ctr -n k8s.io images list | grep k8s-http-fake-operator

# If not found, manually import
sudo ctr -n k8s.io images import k8s-http-fake-operator-latest.tar.gz
```

### Permission Denied on Scripts

```bash
chmod +x build-image.sh
```

### Binary Not Found

Ensure you've built the `manager` binary for Linux amd64:

```bash
# Check if binary exists
ls -la build/manager

# File should show: ELF 64-bit LSB executable, x86-64
file build/manager
```

## After Building

Once the image is built and imported successfully:

1. **Verify the image**:
   ```bash
   # In Docker
   docker images | grep k8s-http-fake-operator
   
   # In containerd
   sudo ctr -n k8s.io images list | grep k8s-http-fake-operator
   ```

2. **Deploy to Kubernetes**:
   ```bash
   helm install k8s-http-fake-operator ./charts/k8s-http-fake-operator
   ```

3. **Check deployment status**:
   ```bash
   kubectl get pods
   kubectl logs -l app.kubernetes.io/name=k8s-http-fake-operator
   ```
