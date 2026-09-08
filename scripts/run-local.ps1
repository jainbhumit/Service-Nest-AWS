# Starts DynamoDB Local, initializes the table, and runs the local API server.
$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

if (-not (Test-Path ".env")) {
    Write-Host "No .env file found. Copy .env.example to .env and adjust values."
    if (Test-Path ".env.example") {
        Copy-Item ".env.example" ".env"
        Write-Host "Created .env from .env.example - update JWT_SECRET before prod use."
    }
}

Write-Host "Starting Docker Compose ..."
docker compose up -d

Write-Host "Initializing DynamoDB Local table ..."
& "$PSScriptRoot\local-init.ps1"

Write-Host "Starting local API server ..."
Set-Location (Join-Path $Root "service-nest")
go run ./cmd/local
