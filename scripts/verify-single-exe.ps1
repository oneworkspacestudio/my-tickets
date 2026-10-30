# Requires PowerShell 7. Run only against disposable test data.
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$Executable,
    [Parameter(Mandatory)][string]$EvidenceDirectory,
    [ValidateSet('DeveloperHost', 'CleanWindowsVM')][string]$EnvironmentKind = 'DeveloperHost',
    [ValidateRange(1025,65535)][int]$Port = 19456
)
$ErrorActionPreference = 'Stop'
if (-not $IsWindows) { throw 'This proof must run on Windows.' }
$sourceExe = (Resolve-Path -LiteralPath $Executable).Path
$evidence = [IO.Path]::GetFullPath($EvidenceDirectory)
if (Test-Path -LiteralPath $evidence) { throw 'Choose a new evidence directory for each run.' }
if (Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue) {
    throw 'The chosen fixture port is in use; choose another port.'
}
New-Item -ItemType Directory -Path $evidence | Out-Null
$fixture = Join-Path $evidence 'fixture'
New-Item -ItemType Directory -Path $fixture | Out-Null
$fixtureExe = Join-Path $fixture 'MyTickets-baseline.exe'
Copy-Item -LiteralPath $sourceExe -Destination $fixtureExe
if (@(Get-ChildItem -LiteralPath $fixture -Force).Count -ne 1) { throw 'Fixture was not EXE-only.' }
$base = 'http://127.0.0.1:' + $Port
$settings = @{
    VIKUNJA_SERVICE_ROOTPATH = $fixture
    VIKUNJA_DATABASE_TYPE = 'sqlite'
    VIKUNJA_DATABASE_PATH = (Join-Path $fixture 'data/mytickets.db')
    VIKUNJA_FILES_BASEPATH = (Join-Path $fixture 'data/attachments')
    VIKUNJA_LOG_PATH = (Join-Path $fixture 'data/logs')
    VIKUNJA_SERVICE_INTERFACE = ('127.0.0.1:' + $Port)
    VIKUNJA_SERVICE_PUBLICURL = ($base + '/')
    VIKUNJA_SERVICE_ENABLEREGISTRATION = 'false'
    VIKUNJA_MAILER_ENABLED = 'false'
    VIKUNJA_REDIS_ENABLED = 'false'
    VIKUNJA_LOG_LEVEL = 'WARN'
    PATH = ($env:WINDIR + '/System32;' + $env:WINDIR)
}
# Refuse inherited app overrides: the proof must not contact any owner installation.
$inherited = @(Get-ChildItem Env: | Where-Object Name -Like 'VIKUNJA_*')
if ($inherited.Count) { throw 'Run from a shell without VIKUNJA_* overrides.' }
$oldEnvironment = @{}
$process = $null
$password = [Guid]::NewGuid().ToString('N') + 'Aa!7'
$proof = [ordered]@{
    status = 'FAIL'
    artifact = 'Source-built upstream baseline; not completed My Tickets product'
    upstreamCommit = '5619c4a9f66ddfc13f7bbba24ada07c3e82ae7a0'
    sha256 = (Get-FileHash -LiteralPath $fixtureExe -Algorithm SHA256).Hash.ToLowerInvariant()
    environmentKind = $EnvironmentKind
    # The script records the operator's environment declaration, not VM attestation.
    environmentDeclarationByOperator = $true
    windowsVersion = [Environment]::OSVersion.VersionString
    initialFileCount = 1
    runtimePath = 'Windows system directories only'
    embeddedUI = $false
    embeddedScriptServed = $false
    sqliteWrites = $false
    forcedTerminationPersistence = $false
    anonymousTaskDenied = $false
    nonSystemRuntimeModules = @()
    runtimeModuleAudit = 'PENDING'
    singleExecutableAcceptance = 'UNVERIFIED'
    browserInteractionTested = $false
    customMyTicketsFeaturesTested = $false
}
function Start-FixtureServer {
    $child = Start-Process -FilePath $fixtureExe -ArgumentList 'web' -WorkingDirectory $fixture -WindowStyle Hidden -RedirectStandardOutput (Join-Path $evidence 'server.stdout.log') -RedirectStandardError (Join-Path $evidence 'server.stderr.log') -PassThru
    try {
        $deadline = [DateTime]::UtcNow.AddSeconds(90)
        do {
            if ($child.HasExited) { throw 'Fixture server exited. Inspect the server logs.' }
            try {
                $health = Invoke-RestMethod ($base + '/api/v2/health') -NoProxy -TimeoutSec 10
                if ($health.status -eq 'OK') { return $child }
            } catch {
                if ($child.HasExited) { throw 'Fixture server exited during readiness.' }
            }
            Start-Sleep -Milliseconds 250
        } while ([DateTime]::UtcNow -lt $deadline)
        throw 'Fixture healthcheck timed out.'
    } catch {
        if (-not $child.HasExited) { $child.Kill(); $child.WaitForExit() }
        throw
    }
}
try {
    foreach ($name in $settings.Keys) {
        $oldEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
        [Environment]::SetEnvironmentVariable($name, $settings[$name], 'Process')
    }
    New-Item -ItemType Directory -Path (Join-Path $fixture 'data') | Out-Null
    Push-Location $fixture
    try {
        & $fixtureExe user create --username singleexetest --email single-exe@example.test --password $password *> (Join-Path $evidence 'user-create.log')
        if ($LASTEXITCODE) { throw 'Synthetic user creation failed.' }
    } finally { Pop-Location }
    $process = Start-FixtureServer
    $root = Invoke-WebRequest ($base + '/') -NoProxy -TimeoutSec 30
    if ($root.StatusCode -ne 200 -or $root.Content -notmatch '<html') { throw 'Embedded UI was not served.' }
    $proof.embeddedUI = $true
    $asset = [regex]::Match($root.Content, 'src="([^\"]+\.js)"').Groups[1].Value
    if (-not $asset) { throw 'Embedded UI contains no script.' }
    $assetUri = [Uri]::new([Uri]($base + '/'), $asset)
    if ($assetUri.Authority -ne ([Uri]$base).Authority) { throw 'Essential script points to a remote server.' }
    if ((Invoke-WebRequest $assetUri -NoProxy -TimeoutSec 30).StatusCode -ne 200) { throw 'Embedded JS missing.' }
    $proof.embeddedScriptServed = $true
    $windowsRoot = [IO.Path]::GetFullPath($env:WINDIR).TrimEnd('\') + '\'
    $unexpectedModules = @($process.Modules | Where-Object {
        $_.FileName -ine $fixtureExe -and -not $_.FileName.StartsWith($windowsRoot, [StringComparison]::OrdinalIgnoreCase)
    } | ForEach-Object { [IO.Path]::GetFileName($_.FileName) })
    $proof.nonSystemRuntimeModules = $unexpectedModules
    $proof.runtimeModuleAudit = 'PASS'
    if ($unexpectedModules.Count) {
        $proof.runtimeModuleAudit = 'REVIEW_REQUIRED'
        if ($EnvironmentKind -eq 'CleanWindowsVM') {
            throw 'Non-system runtime modules were loaded. Inspect packaging before accepting the clean-machine proof.'
        }
    }
    $login = Invoke-RestMethod ($base + '/api/v2/login') -Method Post -ContentType application/json -Body (@{ username = 'singleexetest'; password = $password } | ConvertTo-Json) -NoProxy
    $headers = @{ Authorization = 'Bearer ' + $login.token }
    $project = Invoke-RestMethod ($base + '/api/v2/projects') -Method Post -ContentType application/json -Headers $headers -Body '{"title":"Single executable proof fixture"}' -NoProxy
    $task = Invoke-RestMethod ($base + '/api/v2/projects/' + $project.id + '/tasks') -Method Post -ContentType application/json -Headers $headers -Body '{"title":"Persists through forced termination"}' -NoProxy
    if (-not $task.id) { throw 'Task was not saved.' }
    $proof.sqliteWrites = $true
    $anonymous = Invoke-WebRequest ($base + '/api/v2/tasks/' + $task.id) -NoProxy -SkipHttpErrorCheck
    if ($anonymous.StatusCode -notin @(401,403)) { throw 'Anonymous access was not rejected.' }
    $proof.anonymousTaskDenied = $true
    $process.Kill(); $process.WaitForExit(); $process = $null
    $process = Start-FixtureServer
    $login = Invoke-RestMethod ($base + '/api/v2/login') -Method Post -ContentType application/json -Body (@{ username = 'singleexetest'; password = $password } | ConvertTo-Json) -NoProxy
    $saved = Invoke-RestMethod ($base + '/api/v2/tasks/' + $task.id) -Headers @{ Authorization = 'Bearer ' + $login.token } -NoProxy
    if ($saved.title -cne 'Persists through forced termination') { throw 'Saved task was lost.' }
    $proof.forcedTerminationPersistence = $true
    $proof.status = 'PASS'
    if ($EnvironmentKind -eq 'CleanWindowsVM' -and $proof.runtimeModuleAudit -eq 'PASS') {
        $proof.singleExecutableAcceptance = 'PASS'
    }
} finally {
    if ($process -and -not $process.HasExited) { $process.Kill(); $process.WaitForExit() }
    $password = $null; $login = $null; $headers = $null
    foreach ($name in $oldEnvironment.Keys) { [Environment]::SetEnvironmentVariable($name, $oldEnvironment[$name], 'Process') }
    $proof | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath (Join-Path $evidence 'result.json')
}
$proof | ConvertTo-Json -Depth 4
