# Build Scripts

This directory contains scripts to build the k8s-http-fake-operator Docker image.

## Available Scripts

### Linux/macOS: `build-image.sh`
Bash script for building Docker images on Linux and macOS systems.

### Windows: `build-image.ps1`
PowerShell script for building Docker images on Windows systems.

## Usage

### Linux/macOS

```bash
# Make the script executable (first time only)
chmod +x build-image.sh

# Build with default settings
./build-image.sh

# Build with custom name and tag
./build-image.sh --name my-operator --tag v1.0.0

# Build without cache
./build-image.sh --no-cache

# Build with custom Dockerfile
./build-image.sh --dockerfile Dockerfile.custom

# Show help
./build-image.sh --help
```

### Windows PowerShell

```powershell
# Build with default settings
.\build-image.ps1

# Build with custom name and tag
.\build-image.ps1 -ImageName my-operator -ImageTag v1.0.0

# Build without cache
.\build-image.ps1 -NoCache

# Build with custom Dockerfile
.\build-image.ps1 -Dockerfile Dockerfile.custom

# Show help
.\build-image.ps1 -Help
```

## Environment Variables

You can also configure the build using environment variables:

- `IMAGE_NAME`: Docker image name (default: k8s-http-fake-operator)
- `IMAGE_TAG`: Docker image tag (default: latest)
- `DOCKERFILE`: Path to Dockerfile (default: Dockerfile)
- `BUILD_CONTEXT`: Build context directory (default: .)
- `NO_CACHE`: Build without using cache (true/false)

Example:

```bash
export IMAGE_NAME=my-operator
export IMAGE_TAG=v1.0.0
./build-image.sh
```

## Default Configuration

- **Image Name**: `k8s-http-fake-operator`
- **Image Tag**: `latest`
- **Dockerfile**: `../Dockerfile`
- **Build Context**: `..` (project root)

## After Building

Once the image is built successfully, you can:

1. **Test the image locally**:
   ```bash
   docker run -p 8080:8080 -p 8443:8443 k8s-http-fake-operator:latest
   ```

2. **Push to a registry**:
   ```bash
   docker tag k8s-http-fake-operator:latest <registry>/k8s-http-fake-operator:latest
   docker push <registry>/k8s-http-fake-operator:latest
   ```

3. **Deploy with Helm**:
   Update your `values.yaml` to use the new image:
   ```yaml
   image:
     repository: <registry>/k8s-http-fake-operator
     tag: latest
   ```
   
   Then deploy:
   ```bash
   helm install k8s-http-fake-operator ../charts/k8s-http-fake-operator
   ```

## Troubleshooting

### Docker not found
Make sure Docker is installed and running:
```bash
docker --version
docker ps
```

### Permission denied (Linux/macOS)
Make the script executable:
```bash
chmod +x build-image.sh
```

### Build fails
Check that:
- Dockerfile exists in the correct location
- All required files are present
- Docker daemon is running
- You have sufficient disk space

### PowerShell execution policy (Windows)
If you get an execution policy error, run:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```