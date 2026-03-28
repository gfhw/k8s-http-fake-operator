# Build script for k8s-http-fake-operator Docker image (Windows PowerShell)
# This script automates the Docker image build process

param(
    [string]$ImageName = "k8s-http-fake-operator",
    [string]$ImageTag = "latest",
    [string]$Dockerfile = "Dockerfile",
    [string]$BuildContext = ".",
    [switch]$NoCache = $false,
    [switch]$Help = $false
)

# Functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Show-Usage {
    Write-Host "Usage: .\build-image.ps1 [OPTIONS]" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Build Docker image for k8s-http-fake-operator" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "OPTIONS:" -ForegroundColor Cyan
    Write-Host "    -ImageName IMAGE_NAME      Docker image name (default: k8s-http-fake-operator)"
    Write-Host "    -ImageTag IMAGE_TAG        Docker image tag (default: latest)"
    Write-Host "    -Dockerfile DOCKERFILE     Path to Dockerfile (default: Dockerfile)"
    Write-Host "    -BuildContext BUILD_CONTEXT Build context directory (default: .)"
    Write-Host "    -NoCache                   Build without using cache"
    Write-Host "    -Help                      Show this help message"
    Write-Host ""
    Write-Host "EXAMPLES:" -ForegroundColor Cyan
    Write-Host "    # Build with default settings"
    Write-Host "    .\build-image.ps1"
    Write-Host ""
    Write-Host "    # Build with custom name and tag"
    Write-Host "    .\build-image.ps1 -ImageName my-operator -ImageTag v1.0.0"
    Write-Host ""
    Write-Host "    # Build without cache"
    Write-Host "    .\build-image.ps1 -NoCache"
    Write-Host ""
    Write-Host "    # Build with custom Dockerfile"
    Write-Host "    .\build-image.ps1 -Dockerfile Dockerfile.custom"
}

# Check for help flag
if ($Help) {
    Show-Usage
    exit 0
}

# Main build process
try {
    Write-Info "Starting Docker image build process..."
    Write-Info "Image name: $ImageName"
    Write-Info "Image tag: $ImageTag"
    Write-Info "Dockerfile: $Dockerfile"
    Write-Info "Build context: $BuildContext"
    Write-Info "No cache: $NoCache"

    # Check if Docker is installed
    $dockerCmd = Get-Command docker -ErrorAction SilentlyContinue
    if (-not $dockerCmd) {
        Write-Error "Docker is not installed or not in PATH"
        exit 1
    }

    # Check if Dockerfile exists
    $dockerfilePath = Join-Path $BuildContext $Dockerfile
    if (-not (Test-Path $dockerfilePath)) {
        Write-Error "Dockerfile not found: $dockerfilePath"
        exit 1
    }

    # Build Docker image
    Write-Info "Building Docker image..."
    
    $buildArgs = @("build")
    
    if ($NoCache) {
        $buildArgs += "--no-cache"
    }
    
    $buildArgs += "-t"
    $buildArgs += "$ImageName`:$ImageTag"
    $buildArgs += "-f"
    $buildArgs += $dockerfilePath
    $buildArgs += $BuildContext
    
    $buildCmd = "docker " + ($buildArgs -join " ")
    Write-Info "Executing: $buildCmd"
    
    $result = & docker @buildArgs
    
    if ($LASTEXITCODE -eq 0) {
        Write-Info "Docker image built successfully!"
        Write-Info "Image: $ImageName`:$ImageTag"
        
        # Show image size
        $imageInfo = docker images $ImageName`:$ImageTag --format "{{.Size}}"
        Write-Info "Image size: $imageInfo"
        
        # List the image
        docker images $ImageName`:$ImageTag
        
        Write-Info "You can now push the image to your registry:"
        Write-Info "  docker tag $ImageName`:$ImageTag <registry>/$ImageName`:$ImageTag"
        Write-Info "  docker push <registry>/$ImageName`:$ImageTag"
    } else {
        Write-Error "Docker image build failed!"
        exit 1
    }
} catch {
    Write-Error "An error occurred: $_"
    exit 1
}