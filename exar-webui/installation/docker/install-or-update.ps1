#Requires -Version 5.1
$ErrorActionPreference = 'Stop'

$ImageName = if ($env:IMAGE_NAME) { $env:IMAGE_NAME } else { 'exar-web:latest' }
$ContainerName = if ($env:CONTAINER_NAME) { $env:CONTAINER_NAME } else { 'exar-web' }
$AppPort = if ($env:APP_PORT) { $env:APP_PORT } else { '8080' }
$DataDir = if ($env:DATA_DIR) { $env:DATA_DIR } else { Join-Path $env:LOCALAPPDATA 'exar-web\data' }

$ProjectRoot = Resolve-Path (Join-Path $PSScriptRoot '..\..')
$Dockerfile = Join-Path $ProjectRoot 'dockerfile'

function Write-Log([string]$Message) {
    Write-Host "[docker-install] $Message"
}

function Write-Err([string]$Message) {
    Write-Error "[docker-install] $Message"
}

function Need-Command([string]$Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        Write-Err "missing required command: $Name"
        exit 1
    }
}

Need-Command docker

try {
    docker info *> $null
    if ($LASTEXITCODE -ne 0) {
        Write-Err 'docker daemon is not running'
        exit 1
    }
}
catch {
    Write-Err 'docker is not accessible on this host'
    exit 1
}

if (-not (Test-Path -LiteralPath $Dockerfile)) {
    Write-Err "dockerfile not found: $Dockerfile"
    exit 1
}

New-Item -ItemType Directory -Force -Path $DataDir | Out-Null

Write-Log "Building Docker image $ImageName (multi-stage) ..."
docker build -f $Dockerfile -t $ImageName $ProjectRoot
if ($LASTEXITCODE -ne 0) {
    Write-Err 'docker build failed'
    exit $LASTEXITCODE
}

$existing = docker ps -a --format '{{.Names}}' | Where-Object { $_ -eq $ContainerName }
if ($existing) {
    Write-Log "Container $ContainerName exists. Replacing with updated image ..."
    docker rm -f $ContainerName | Out-Null
}
else {
    Write-Log "Container $ContainerName does not exist. Installing new container ..."
}

docker run -d `
    --name $ContainerName `
    --restart unless-stopped `
    -p "${AppPort}:8080" `
    -v "${DataDir}:/app/data" `
    $ImageName | Out-Null

if ($LASTEXITCODE -ne 0) {
    Write-Err 'docker run failed'
    exit $LASTEXITCODE
}

Write-Log "Done. Running container: $ContainerName on http://localhost:$AppPort"
exit 0
