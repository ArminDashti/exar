<#
.SYNOPSIS
  Deploy stack to a remote host over SSH using sibling YAML only.

.DESCRIPTION
  Remote deploy script for .armin/docker-scripts/run-on-docker-server.yaml.
  Reads run-on-docker-server.yaml — no CLI -- flags.
  Flow when build_image_on is local: build locally → docker save → SCP → remote docker load → sync files → remote compose up -d.
  Flow when build_image_on is server: sync repo to remote → remote docker build → remote compose up -d.
#>
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$DeployDir = $PSScriptRoot
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $DeployDir '../..'))
$ConfigPath = Join-Path $DeployDir 'run-on-docker-server.yaml'

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
run-on-docker-server.ps1 — remote Docker deploy (YAML-only)

USAGE:
  .\.armin\docker-scripts\run-on-docker-server.ps1

CONFIG:
  Sibling file: run-on-docker-server.yaml

  stack_name          Compose project name (-p)
  image_tag           Image tag for build and compose; overrides compose when set
  compose_file        Compose path relative to .armin/docker-scripts
  dockerfile          Dockerfile path relative to .armin/docker-scripts
  docker_network      External Docker network on remote
  publish_port        Optional host bind port; omit or empty = no host bind
  internal_port       Container listen port; overrides compose when set
  delete_volume       yes/true/1/y/on → remove volumes before up
  delete_image        yes/true/1/y/on → remove image during teardown
  build_image_on      local = build here and upload; server = build on remote
  ssh                 "ssh <alias>", "ssh <alias> -p <port>", "ssh -p <port> <alias>", or "host@user@password"
  volume_dir          Absolute remote directory for project + compose files

NOTES:
  - No CLI -- flags. Change behavior only via YAML.
  - Sets IMAGE_TAG / DOCKER_NETWORK / INTERNAL_PORT / PUBLISH_PORT for compose.
  - Alias mode uses ~/.ssh/config (no ssh_key field); optional -p overrides Port.
  - Rejects placeholder ssh values at runtime.
  - Never prints the password segment of host@user@password.
  - build_image_on=local requires Docker on this machine.
  - build_image_on=server syncs the repo to volume_dir and builds there.
"@ -ForegroundColor Cyan
}

function Test-Truthy([string]$Value) {
    if ([string]::IsNullOrWhiteSpace($Value)) { return $false }
    return $Value.Trim().ToLowerInvariant() -in @('yes', 'true', '1', 'y', 'on')
}

function Test-Placeholder([string]$Value) {
    if ([string]::IsNullOrWhiteSpace($Value)) { return $true }
    return $Value -match '<[^>]+>'
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
        if ($val -match '^(?<q>["'']).*?\k<q>(?<rest>\s+#.*)?$') {
            if ($Matches['rest']) {
                $val = $val.Substring(0, $val.Length - $Matches['rest'].Length).Trim()
            }
        }
        elseif ($val -match '^(?<body>.*?)\s+#') {
            $val = $Matches['body'].Trim()
        }
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

function Get-RepoRelativePath([string]$AbsolutePath) {
    $root = $RepoRoot.TrimEnd('\', '/')
    $path = $AbsolutePath.TrimEnd('\', '/')
    if (-not $path.StartsWith($root, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Path is outside repo root: $AbsolutePath"
    }
    $relative = $path.Substring($root.Length).TrimStart('\', '/')
    return ($relative -replace '\\', '/')
}

function Build-ComposeEnvPrefix([hashtable]$Cfg, [string]$PublishPort) {
    $pairs = New-Object System.Collections.Generic.List[string]

    $mapping = @{
        image_tag      = 'IMAGE_TAG'
        docker_network = 'DOCKER_NETWORK'
        internal_port  = 'INTERNAL_PORT'
    }
    foreach ($entry in $mapping.GetEnumerator()) {
        if (-not $Cfg.ContainsKey($entry.Key)) { continue }
        $value = [string]$Cfg[$entry.Key]
        if ([string]::IsNullOrWhiteSpace($value)) { continue }
        $escaped = $value.Replace("'", "'\\''")
        [void]$pairs.Add("$($entry.Value)='$escaped'")
    }
    if (-not [string]::IsNullOrWhiteSpace($PublishPort)) {
        $escapedPublish = $PublishPort.Replace("'", "'\\''")
        [void]$pairs.Add("PUBLISH_PORT='$escapedPublish'")
    }
    return ($pairs -join ' ') + ' '
}

function Ensure-Docker {
    docker version *> $null
    if ($LASTEXITCODE -ne 0) { throw 'Docker CLI is not available. Start Docker Desktop / daemon.' }
}

function Parse-SshTarget([string]$SshValue) {
    $value = $SshValue.Trim()
    if ($value -match '^(?i)ssh(?:\s+-p\s*(?<port1>\d+))?\s+(?<alias>\S+)(?:\s+-p\s*(?<port2>\d+))?$') {
        $alias = $Matches['alias']
        $port = $null
        if ($Matches['port1']) { $port = $Matches['port1'] }
        if ($Matches['port2']) { $port = $Matches['port2'] }
        $logTarget = if ($port) { "ssh $alias -p $port" } else { "ssh $alias" }
        return @{
            Mode      = 'alias'
            Alias     = $alias
            Port      = $port
            LogTarget = $logTarget
        }
    }

    $parts = $value.Split('@')
    if ($parts.Count -eq 3) {
        $hostName = $parts[0]
        $userName = $parts[1]
        $password = $parts[2]
        if ((Test-Placeholder $hostName) -or (Test-Placeholder $userName) -or [string]::IsNullOrWhiteSpace($password)) {
            throw 'ssh password mode still has placeholders. Fill host@user@password in YAML.'
        }
        return @{
            Mode      = 'password'
            Host      = $hostName
            User      = $userName
            Password  = $password
            Port      = $null
            LogTarget = "$userName@$hostName"
        }
    }

    throw 'ssh must be "ssh <alias>", "ssh <alias> -p <port>", "ssh -p <port> <alias>", or "host@user@password".'
}

function Invoke-Remote {
    param($Target, [string]$RemoteCommand)

    if ($Target.Mode -eq 'alias') {
        $sshArgs = @('-o', 'BatchMode=yes')
        if ($Target.Port) { $sshArgs += @('-p', $Target.Port) }
        $sshArgs += @($Target.Alias, $RemoteCommand)
        & ssh @sshArgs
        if ($LASTEXITCODE -ne 0) { throw "Remote command failed on $($Target.LogTarget)" }
        return
    }

    if (-not (Get-Command sshpass -ErrorAction SilentlyContinue)) {
        throw 'Password mode requires sshpass on PATH (or switch YAML to ssh alias mode).'
    }
    $env:SSHPASS = $Target.Password
    try {
        & sshpass -e ssh -o StrictHostKeyChecking=accept-new "$($Target.User)@$($Target.Host)" $RemoteCommand
        if ($LASTEXITCODE -ne 0) { throw "Remote command failed on $($Target.LogTarget)" }
    }
    finally {
        Remove-Item Env:SSHPASS -ErrorAction SilentlyContinue
    }
}

function Copy-ToRemote {
    param($Target, [string]$LocalPath, [string]$RemotePath)

    if ($Target.Mode -eq 'alias') {
        $scpArgs = @('-o', 'BatchMode=yes')
        if ($Target.Port) { $scpArgs += @('-P', $Target.Port) }
        $scpArgs += @($LocalPath, "$($Target.Alias):$RemotePath")
        & scp @scpArgs
        if ($LASTEXITCODE -ne 0) { throw "SCP failed to $($Target.LogTarget):$RemotePath" }
        return
    }

    if (-not (Get-Command sshpass -ErrorAction SilentlyContinue)) {
        throw 'Password mode requires sshpass on PATH (or switch YAML to ssh alias mode).'
    }
    $env:SSHPASS = $Target.Password
    try {
        & sshpass -e scp -o StrictHostKeyChecking=accept-new $LocalPath "$($Target.User)@$($Target.Host):$RemotePath"
        if ($LASTEXITCODE -ne 0) { throw "SCP failed to $($Target.LogTarget):$RemotePath" }
    }
    finally {
        Remove-Item Env:SSHPASS -ErrorAction SilentlyContinue
    }
}

function Copy-DirToRemote {
    param($Target, [string]$LocalDir, [string]$RemoteDir)

    if ($Target.Mode -eq 'alias') {
        $scpArgs = @('-r', '-o', 'BatchMode=yes')
        if ($Target.Port) { $scpArgs += @('-P', $Target.Port) }
        $scpArgs += @("$LocalDir/.", "$($Target.Alias):$RemoteDir/")
        & scp @scpArgs
        if ($LASTEXITCODE -ne 0) { throw "SCP directory failed to $($Target.LogTarget):$RemoteDir" }
        return
    }

    if (-not (Get-Command sshpass -ErrorAction SilentlyContinue)) {
        throw 'Password mode requires sshpass on PATH (or switch YAML to ssh alias mode).'
    }
    $env:SSHPASS = $Target.Password
    try {
        & sshpass -e scp -r -o StrictHostKeyChecking=accept-new "$LocalDir/." "$($Target.User)@$($Target.Host):$RemoteDir/"
        if ($LASTEXITCODE -ne 0) { throw "SCP directory failed to $($Target.LogTarget):$RemoteDir" }
    }
    finally {
        Remove-Item Env:SSHPASS -ErrorAction SilentlyContinue
    }
}

if ($args.Count -gt 0) {
    Write-Fail 'This script accepts no CLI arguments. Edit run-on-docker-server.yaml instead.'
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
    $publishPort = if ($cfg.ContainsKey('publish_port')) { [string]$cfg['publish_port'] } else { '' }
    $internalPort = if ($cfg.ContainsKey('internal_port')) { [string]$cfg['internal_port'] } else { '' }
    $deleteVolume = Test-Truthy ($(if ($cfg.ContainsKey('delete_volume')) { [string]$cfg['delete_volume'] } else { 'no' }))
    $deleteImage = Test-Truthy ($(if ($cfg.ContainsKey('delete_image')) { [string]$cfg['delete_image'] } else { 'no' }))
    $buildImageOn = if ($cfg.ContainsKey('build_image_on')) { [string]$cfg['build_image_on'] } else { 'local' }
    $buildImageOn = $buildImageOn.Trim().ToLowerInvariant()
    $sshValue = Require-Key $cfg 'ssh'
    $volumeDir = Require-Key $cfg 'volume_dir'

    if ($buildImageOn -notin @('local', 'server')) {
        throw "build_image_on must be 'local' or 'server'."
    }
    if ($buildImageOn -eq 'local') {
        Ensure-Docker
    }
    if (Test-Placeholder $sshValue) {
        throw 'ssh still has placeholders. Fill run-on-docker-server.yaml before server deploy.'
    }
    if (Test-Placeholder $volumeDir) {
        throw 'volume_dir still has placeholders. Fill a real absolute remote path.'
    }

    $composePath = Resolve-DeployPath $composeFileRel
    $dockerfile = Resolve-DeployPath $dockerfileRel
    $composeFileName = Split-Path -Leaf $composePath
    $publishComposeName = 'docker-compose.publish.yml'
    $publishComposePath = Join-Path $RepoRoot $publishComposeName
    $usePublishOverlay = -not [string]::IsNullOrWhiteSpace($publishPort)
    if ($usePublishOverlay -and -not (Test-Path -LiteralPath $publishComposePath)) {
        throw "publish_port is set but overlay missing: $publishComposePath"
    }
    $remoteDockerfile = Get-RepoRelativePath $dockerfile
    $remoteCompose = "$volumeDir/$composeFileName"
    $remotePublishCompose = "$volumeDir/$publishComposeName"
    $remoteComposeFileArgs = "-f '$remoteCompose'"
    if ($usePublishOverlay) {
        $remoteComposeFileArgs = "$remoteComposeFileArgs -f '$remotePublishCompose'"
    }

    $target = Parse-SshTarget -SshValue $sshValue
    Write-Step "Remote target: $($target.LogTarget)"
    Write-Step "Stack=$stackName image=$imageTag build_image_on=$buildImageOn volume_dir=$volumeDir publish_port='$publishPort' internal_port='$internalPort'"

    Write-Step "Ensuring remote volume dir $volumeDir"
    Invoke-Remote -Target $target -RemoteCommand "mkdir -p '$volumeDir'"

    # Sync compose/config before teardown so `compose down` has a file to read.
    if ($buildImageOn -eq 'server') {
        Write-Step "Syncing repo to $volumeDir for remote build"
        Copy-DirToRemote -Target $target -LocalDir $RepoRoot -RemoteDir $volumeDir
        Write-Ok 'Repo synced to remote'
    }
    else {
        $syncItems = New-Object System.Collections.Generic.List[string]
        [void]$syncItems.Add($composeFileName)
        if ($usePublishOverlay -or (Test-Path -LiteralPath $publishComposePath)) {
            [void]$syncItems.Add($publishComposeName)
        }
        [void]$syncItems.Add('run-on-docker-server.yaml')
        foreach ($item in $syncItems) {
            $localItem = if ($item -eq $composeFileName) {
                $composePath
            }
            elseif ($item -eq $publishComposeName) {
                $publishComposePath
            }
            else {
                Join-Path $DeployDir $item
            }
            if (-not (Test-Path -LiteralPath $localItem)) { throw "Sync source not found: $localItem" }
            $remoteItem = "$volumeDir/$item"
            Write-Step "Sync $item"
            Copy-ToRemote -Target $target -LocalPath $localItem -RemotePath $remoteItem
        }
    }

    # Teardown BEFORE loading/building the new image so delete_image does not
    # wipe the image we just uploaded (local) or are about to build (server).
    $downFlags = if ($deleteVolume) { '-v' } else { '' }

    # Prefer both compose files on down when overlay exists remotely (avoids orphan publish bindings).
    $remoteDownFileArgs = "-f '$remoteCompose'"
    if (Test-Path -LiteralPath $publishComposePath) {
        $remoteDownFileArgs = "$remoteDownFileArgs -f '$remotePublishCompose'"
    }

    if ($deleteVolume -or $deleteImage) {
        Write-Step 'Remote compose down'
        Invoke-Remote -Target $target -RemoteCommand "docker compose -p '$stackName' $remoteDownFileArgs --project-directory '$volumeDir' down $downFlags || true"
    }

    if ($deleteImage) {
        Write-Step "Removing remote image $imageTag"
        Invoke-Remote -Target $target -RemoteCommand "docker image rm -f '$imageTag' || true"
    }

    if ($buildImageOn -eq 'local') {
        Write-Step "Building $imageTag locally (dockerfile=$dockerfile context=$RepoRoot)"
        docker build -f $dockerfile -t $imageTag $RepoRoot
        if ($LASTEXITCODE -ne 0) { throw 'docker build failed' }
        Write-Ok "Built $imageTag"

        $tarName = ($imageTag -replace '[:/]', '_') + '.tar'
        $tarPath = Join-Path $env:TEMP $tarName
        Write-Step "Saving image to $tarPath"
        docker save -o $tarPath $imageTag
        if ($LASTEXITCODE -ne 0) { throw 'docker save failed' }

        $remoteTar = "/tmp/$tarName"
        Write-Step "Uploading image to $($target.LogTarget)"
        Copy-ToRemote -Target $target -LocalPath $tarPath -RemotePath $remoteTar
        Invoke-Remote -Target $target -RemoteCommand "docker load -i $remoteTar && rm -f $remoteTar"
        Write-Ok 'Image loaded on remote'
        Remove-Item -LiteralPath $tarPath -Force -ErrorAction SilentlyContinue
    }
    else {
        Write-Step "Building $imageTag on remote (dockerfile=$remoteDockerfile context=$volumeDir)"
        Invoke-Remote -Target $target -RemoteCommand "docker build -f '$volumeDir/$remoteDockerfile' -t '$imageTag' '$volumeDir'"
        Write-Ok "Built $imageTag on remote"
    }

    Write-Step "Ensuring remote network $network"
    Invoke-Remote -Target $target -RemoteCommand "docker network inspect '$network' >/dev/null 2>&1 || docker network create '$network'"

    $envPrefix = Build-ComposeEnvPrefix @{
        image_tag      = $imageTag
        docker_network = $network
        internal_port  = $internalPort
    } $publishPort

    # Local builds ship a preloaded image; forbid remote rebuild (no Dockerfile synced).
    $upExtra = if ($buildImageOn -eq 'local') { ' --no-build' } else { '' }

    Write-Step 'Remote compose up -d'
    Invoke-Remote -Target $target -RemoteCommand "${envPrefix}docker compose -p '$stackName' $remoteComposeFileArgs --project-directory '$volumeDir' up -d$upExtra"
    Write-Ok "Stack deployed at $volumeDir on $($target.LogTarget)"
}
catch {
    Write-Fail $_.Exception.Message
    Show-Help
    exit 1
}
