param(
    [string]$Profile = $(if ($env:AWS_PROFILE) { $env:AWS_PROFILE } else { "" }),
    [string]$Region = $(if ($env:AWS_REGION) { $env:AWS_REGION } else { "us-east-1" }),
    [string]$StackName = "serviceNest",
    [string]$ApiUrl = ""
)

$ErrorActionPreference = "Stop"

function Get-HttpStatus {
    param(
        [string]$Url,
        [string]$Method = "GET",
        [string]$Body = ""
    )

    $curlArgs = @("-s", "-o", "NUL", "-w", "%{http_code}", "-X", $Method)
    if ($Body) {
        $curlArgs += @("-H", "Content-Type: application/json", "-d", $Body)
    }
    $curlArgs += $Url

    $status = & curl.exe @curlArgs
    if ($LASTEXITCODE -ne 0) {
        throw "curl failed for $Url"
    }
    return [int]$status
}

function Get-HttpBody {
    param(
        [string]$Url,
        [string]$Method = "GET",
        [string]$Body = ""
    )

    $curlArgs = @("-s", "-X", $Method)
    if ($Body) {
        $curlArgs += @("-H", "Content-Type: application/json", "-d", $Body)
    }
    $curlArgs += $Url

    $content = & curl.exe @curlArgs
    if ($LASTEXITCODE -ne 0) {
        throw "curl failed for $Url"
    }
    return $content
}

if ([string]::IsNullOrWhiteSpace($ApiUrl)) {
    $awsArgs = @(
        "cloudformation", "describe-stacks",
        "--stack-name", $StackName,
        "--region", $Region,
        "--query", "Stacks[0].Outputs[?OutputKey=='ServiceNestApiUrl'].OutputValue",
        "--output", "text",
        "--no-cli-pager"
    )
    if ($Profile) {
        $awsArgs += @("--profile", $Profile)
    }

    $ApiUrl = (& aws @awsArgs).Trim()
}

if ([string]::IsNullOrWhiteSpace($ApiUrl)) {
    Write-Error "Could not resolve ServiceNestApiUrl from stack '$StackName'."
    exit 1
}

$base = $ApiUrl.TrimEnd("/")
Write-Host "Smoke testing $base"

$healthStatus = Get-HttpStatus -Url "$base/health"
if ($healthStatus -ne 200) {
    Write-Error "GET /health failed with status $healthStatus"
    exit 1
}
Write-Host "GET /health -> $healthStatus"

$readyStatus = Get-HttpStatus -Url "$base/health/ready"
if ($readyStatus -ne 200) {
    Write-Error "GET /health/ready failed with status $readyStatus"
    exit 1
}
Write-Host "GET /health/ready -> $readyStatus"

$loginStatus = Get-HttpStatus -Url "$base/login" -Method "POST" -Body "{}"
if ($loginStatus -lt 400 -or $loginStatus -ge 500) {
    Write-Error "POST /login expected 4xx, got $loginStatus"
    exit 1
}

$loginBody = Get-HttpBody -Url "$base/login" -Method "POST" -Body "{}"
if ($loginBody -notmatch '"Fail"') {
    Write-Error "POST /login response missing standard error envelope"
    exit 1
}
Write-Host "POST /login -> $loginStatus (error envelope ok)"

Write-Host "Smoke bundle passed for $base"
