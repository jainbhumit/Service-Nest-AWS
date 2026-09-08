# Creates the servicenest DynamoDB table in DynamoDB Local.
param(
    [string]$Endpoint = $(if ($env:DYNAMODB_ENDPOINT) { $env:DYNAMODB_ENDPOINT } else { "http://localhost:8001" }),
    [string]$TableName = $(if ($env:DYNAMODB_TABLE) { $env:DYNAMODB_TABLE } else { "servicenest" }),
    [string]$Region = $(if ($env:AWS_REGION) { $env:AWS_REGION } else { "us-east-1" })
)

$ErrorActionPreference = "Stop"

Write-Host "Waiting for DynamoDB Local at $Endpoint ..."
$ready = $false
for ($i = 0; $i -lt 30; $i++) {
    aws dynamodb list-tables `
        --endpoint-url $Endpoint `
        --region $Region `
        --no-cli-pager `
        2>$null | Out-Null
    if ($LASTEXITCODE -eq 0) {
        $ready = $true
        break
    }
    Start-Sleep -Seconds 1
}

if (-not $ready) {
    Write-Error "DynamoDB Local is not reachable at $Endpoint. Run 'docker compose up -d' first."
    exit 1
}

$existing = aws dynamodb list-tables `
    --endpoint-url $Endpoint `
    --region $Region `
    --no-cli-pager `
    --output text 2>$null

if ($existing -match $TableName) {
    Write-Host "Table '$TableName' already exists."
    exit 0
}

Write-Host "Creating table '$TableName' ..."
aws dynamodb create-table `
    --endpoint-url $Endpoint `
    --region $Region `
    --table-name $TableName `
    --attribute-definitions AttributeName=PK,AttributeType=S AttributeName=SK,AttributeType=S `
    --key-schema AttributeName=PK,KeyType=HASH AttributeName=SK,KeyType=RANGE `
    --billing-mode PAY_PER_REQUEST `
    --no-cli-pager

Write-Host "Table '$TableName' created."
