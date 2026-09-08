param(
    [string]$Profile = $(if ($env:AWS_PROFILE) { $env:AWS_PROFILE } else { "" }),
    [string]$Region = $(if ($env:AWS_REGION) { $env:AWS_REGION } else { "us-east-1" }),
    [string]$StackName = "serviceNest",
    [switch]$Guided,
    [switch]$NoConfirm
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

function Import-DotEnv {
    param([string]$Path)
    if (-not (Test-Path $Path)) {
        return
    }
    Get-Content $Path | ForEach-Object {
        $line = $_.Trim()
        if ($line -eq "" -or $line.StartsWith("#")) {
            return
        }
        $idx = $line.IndexOf("=")
        if ($idx -lt 1) {
            return
        }
        $key = $line.Substring(0, $idx).Trim()
        $value = $line.Substring($idx + 1).Trim()
        if ($value.StartsWith('"') -and $value.EndsWith('"')) {
            $value = $value.Substring(1, $value.Length - 2)
        }
        if (-not [string]::IsNullOrWhiteSpace($key) -and -not (Test-Path "env:$key")) {
            Set-Item -Path "env:$key" -Value $value
        }
    }
}

Import-DotEnv (Join-Path $Root ".env")

$jwtSecret = $env:JWT_SECRET
if ([string]::IsNullOrWhiteSpace($jwtSecret)) {
    $jwtSecret = $env:SECRET
}
$smtpPassword = $env:SMTP_APP_PASSWORD
if ([string]::IsNullOrWhiteSpace($smtpPassword)) {
    $smtpPassword = $env:APP_PASSWORD
}
$smtpFrom = $env:SMTP_FROM

if ([string]::IsNullOrWhiteSpace($jwtSecret)) {
    Write-Error "JWT_SECRET (or SECRET) must be set in the environment or .env file."
    exit 1
}
if ([string]::IsNullOrWhiteSpace($smtpPassword)) {
    Write-Error "SMTP_APP_PASSWORD (or APP_PASSWORD) must be set in the environment or .env file."
    exit 1
}
if ([string]::IsNullOrWhiteSpace($smtpFrom)) {
    Write-Error "SMTP_FROM must be set in the environment or .env file."
    exit 1
}

Write-Warning "This deploy updates the PROD stack '$StackName' in $Region."
Write-Host "Running sam build ..."
sam build
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

$overrideArgs = @(
    "JwtSecret=$jwtSecret",
    "SmtpAppPassword=$smtpPassword",
    "SmtpFrom=$smtpFrom"
)

$samArgs = @("deploy", "--stack-name", $StackName, "--region", $Region, "--capabilities", "CAPABILITY_IAM", "--resolve-s3", "--parameter-overrides") + $overrideArgs

if ($Profile) {
    $samArgs += @("--profile", $Profile)
}
if ($Guided) {
    $samArgs += "--guided"
} elseif ($NoConfirm -or ($env:CONFIRM_CHANGESET -eq "false")) {
    $samArgs += "--no-confirm-changeset"
}

Write-Host "Running sam deploy ..."
& sam @samArgs
exit $LASTEXITCODE
