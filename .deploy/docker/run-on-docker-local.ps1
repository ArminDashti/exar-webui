<#
.SYNOPSIS
  Deploy stack on the local Docker daemon using sibling YAML only.

.DESCRIPTION
  Sample for .deploy/docker/run-on-docker-local.ps1.
  Reads run-on-docker-local.yaml — no CLI -- flags.
  Flow: ensure Docker → build image → ensure network → optional down → compose up -d.
#>
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$DeployDir = $PSScriptRoot
$RepoRoot = (Resolve-Path (Join-Path $DeployDir '../..')).Path
$ConfigPath = Join-Path $DeployDir 'run-on-docker-local.yaml'

function Write-Step([string]$Message) {
    Write-Host ">> $Message" -ForegroundColor Cyan
}

function Write-Ok([string]$Message) {
    Write-Host "OK  $Message" -ForegroundColor Green
}

function Write-Fail([string]$Message) {
    Write-Host "ERR $Message" -ForegroundColor Red
}

function Show-Help {
    Write-Host @"
run-on-docker-local.ps1 — local Docker deploy (YAML-only)

USAGE:
  .\.deploy\docker\run-on-docker-local.ps1

CONFIG:
  Sibling file: run-on-docker-local.yaml

  stack_name          Compose project name (-p)
  image_tag           Image tag for build and compose; overrides compose when set
  compose_file        Compose path relative to .deploy/docker
  dockerfile          Dockerfile path relative to .deploy/docker
  docker_network      External Docker network
  publish_port        Optional host bind port; omit or empty = compose default
  internal_port       Container listen port; overrides compose when set
  delete_volume       yes/true/1/y/on → remove volumes before up
  delete_image        yes/true/1/y/on → remove image during teardown

NOTES:
  - No CLI -- flags. Change behavior only via YAML.
  - Non-empty override fields replace compose / Dockerfile values via env vars.
  - Requires Docker Desktop / local Docker daemon.
"@ -ForegroundColor Cyan
}

function Test-Truthy([string]$Value) {
    if ([string]::IsNullOrWhiteSpace($Value)) { return $false }
    return $Value.Trim().ToLowerInvariant() -in @('yes', 'true', '1', 'y', 'on')
}

function Read-FlatYaml([string]$Path) {
    if (-not (Test-Path -LiteralPath $Path)) {
        throw "Missing config: $Path"
    }
    $map = @{}
    foreach ($raw in Get-Content -LiteralPath $Path) {
        $line = $raw.Trim()
        if ($line -eq '' -or $line.StartsWith('#')) { continue }
        if ($line -match '^\s*-') { continue }
        if ($line -notmatch '^(?<key>[^:#]+):\s*(?<val>.*)$') { continue }
        $key = $Matches['key'].Trim()
        $val = $Matches['val'].Trim()
        if (($val.StartsWith('"') -and $val.EndsWith('"')) -or ($val.StartsWith("'") -and $val.EndsWith("'"))) {
            $val = $val.Substring(1, $val.Length - 2)
        }
        $map[$key] = $val
    }
    return $map
}

function Require-Key($Map, [string]$Key) {
    if (-not $Map.ContainsKey($Key) -or [string]::IsNullOrWhiteSpace([string]$Map[$Key])) {
        throw "YAML missing required key: $Key"
    }
    return [string]$Map[$Key]
}

function Resolve-DeployPath([string]$RelativePath) {
    $candidate = Join-Path $DeployDir $RelativePath
    return (Resolve-Path -LiteralPath $candidate).Path
}

function Clear-ComposeEnv {
    foreach ($name in @('IMAGE_TAG', 'DOCKER_NETWORK', 'INTERNAL_PORT', 'PUBLISH_PORT')) {
        Remove-Item "Env:$name" -ErrorAction SilentlyContinue
    }
}

function Ensure-Docker {
    docker version *> $null
    if ($LASTEXITCODE -ne 0) { throw 'Docker CLI is not available. Start Docker Desktop / daemon.' }
}

if ($args.Count -gt 0) {
    Write-Fail 'This script accepts no CLI arguments. Edit run-on-docker-local.yaml instead.'
    Show-Help
    exit 1
}

try {
    $cfg = Read-FlatYaml $ConfigPath
    $stackName = Require-Key $cfg 'stack_name'
    $imageTag = Require-Key $cfg 'image_tag'
    $composeFileRel = Require-Key $cfg 'compose_file'
    $dockerfileRel = Require-Key $cfg 'dockerfile'
    $network = Require-Key $cfg 'docker_network'
    $publishPort = if ($cfg.ContainsKey('publish_port')) { [string]$cfg['publish_port'] } else { $null }
    $internalPort = if ($cfg.ContainsKey('internal_port')) { [string]$cfg['internal_port'] } else { '' }
    $deleteVolume = Test-Truthy ($(if ($cfg.ContainsKey('delete_volume')) { [string]$cfg['delete_volume'] } else { 'no' }))
    $deleteImage = Test-Truthy ($(if ($cfg.ContainsKey('delete_image')) { [string]$cfg['delete_image'] } else { 'no' }))

    Ensure-Docker

    $composePath = Resolve-DeployPath $composeFileRel
    $dockerfile = Resolve-DeployPath $dockerfileRel
    $composeDir = Split-Path -Parent $composePath

    Write-Step "Stack=$stackName image=$imageTag network=$network publish_port='$publishPort' internal_port='$internalPort'"

    Write-Step "Building $imageTag (dockerfile=$dockerfile context=$RepoRoot)"
    docker build -f $dockerfile -t $imageTag $RepoRoot
    if ($LASTEXITCODE -ne 0) { throw 'docker build failed' }
    Write-Ok "Built $imageTag"

    Write-Step "Ensuring network $network"
    docker network inspect $network *> $null
    if ($LASTEXITCODE -ne 0) {
        docker network create $network
        if ($LASTEXITCODE -ne 0) { throw "Failed to create network $network" }
    }
    Write-Ok "Network $network ready"

    $downFlags = @()
    if ($deleteVolume) { $downFlags += '-v' }

    if ($deleteVolume -or $deleteImage) {
        Write-Step 'Compose down'
        Push-Location $composeDir
        try {
            $env:IMAGE_TAG = $imageTag
            $env:DOCKER_NETWORK = $network
            if (-not [string]::IsNullOrWhiteSpace($internalPort)) {
                $env:INTERNAL_PORT = $internalPort
            }
            if ($null -ne $publishPort) {
                $env:PUBLISH_PORT = $publishPort
            }
            & docker compose -p $stackName -f $composePath down @downFlags
        }
        finally {
            Clear-ComposeEnv
            Pop-Location
        }
    }

    if ($deleteImage) {
        Write-Step "Removing image $imageTag"
        docker image rm -f $imageTag 2>$null
    }

    Write-Step 'Compose up -d'
    Push-Location $composeDir
    try {
        $env:IMAGE_TAG = $imageTag
        $env:DOCKER_NETWORK = $network
        if (-not [string]::IsNullOrWhiteSpace($internalPort)) {
            $env:INTERNAL_PORT = $internalPort
        }
        if ($null -ne $publishPort) {
            $env:PUBLISH_PORT = $publishPort
        }

        & docker compose -p $stackName -f $composePath up -d
        if ($LASTEXITCODE -ne 0) { throw 'docker compose up failed' }
    }
    finally {
        Clear-ComposeEnv
        Pop-Location
    }

    Write-Ok "Stack '$stackName' is up (image=$imageTag, network=$network)"
}
catch {
    Write-Fail $_.Exception.Message
    Show-Help
    exit 1
}
