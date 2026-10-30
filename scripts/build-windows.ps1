param(
 [string]$Go='go',
 [string]$Pnpm='pnpm',
 [string]$FrontendDirectory='',
 [string]$Version='0.1.0-preview.1',
 [Parameter(Mandatory)][string]$Out,
 [switch]$PreparedFrontend
)
$ErrorActionPreference='Stop'
$root=Split-Path $PSScriptRoot -Parent
if(!$FrontendDirectory){$FrontendDirectory=Join-Path $root 'frontend'}
if(!$PreparedFrontend){
 Push-Location $FrontendDirectory
 try {
  & $Pnpm install --frozen-lockfile
  if($LASTEXITCODE){throw 'Pinned frontend dependencies failed'}
  & $Pnpm run typecheck
  if($LASTEXITCODE){throw 'Frontend typecheck failed'}
  & $Pnpm run build
  if($LASTEXITCODE){throw 'Complete embedded frontend build failed'}
  & $Pnpm exec vite build --config vite.tickets.config.ts
  if($LASTEXITCODE){throw 'My Tickets UI build failed'}
 } finally {Pop-Location}
 if([IO.Path]::GetFullPath($FrontendDirectory) -ne [IO.Path]::GetFullPath((Join-Path $root 'frontend'))){
  Copy-Item (Join-Path $FrontendDirectory 'dist/*') (Join-Path $root 'frontend/dist/') -Recurse -Force
 }
}
if(!(Test-Path (Join-Path $root 'frontend/dist/mytickets.html'))){throw 'Complete frontend is missing'}
$commit=& git -c "safe.directory=$($root.Replace('\','/'))" -C $root rev-parse HEAD
if($LASTEXITCODE){throw 'Source commit unavailable'}
$built=if($env:SOURCE_DATE_EPOCH){[DateTimeOffset]::FromUnixTimeSeconds([long]$env:SOURCE_DATE_EPOCH).UtcDateTime.ToString('yyyy-MM-ddTHH:mm:ssZ')}else{[DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ssZ')}
New-Item -ItemType Directory -Force $Out|Out-Null
$oldWork=$env:GOWORK
try {
 $env:GOWORK='off';$env:CGO_ENABLED='1';$env:GOOS='windows';$env:GOARCH='amd64'
 & $Go -C $root run github.com/akavel/rsrc@v0.10.2 -ico frontend/public/tickets.ico -o rsrc_windows_amd64.syso
 if($LASTEXITCODE){throw 'Windows icon resource failed'}
 $flags="-s -w -extldflags=-static -X code.vikunja.io/api/pkg/version.Version=2.7.1 -X code.vikunja.io/api/pkg/oneworkspace.Version=$Version -X code.vikunja.io/api/pkg/oneworkspace.Commit=$commit -X code.vikunja.io/api/pkg/oneworkspace.BuildTime=$built"
 & $Go -C $root build -buildvcs=false -trimpath -tags osusergo -ldflags $flags -o (Join-Path $Out 'MyTickets.exe') .
 if($LASTEXITCODE){throw 'Single Windows executable build failed'}
 Get-FileHash (Join-Path $Out 'MyTickets.exe') -Algorithm SHA256
}finally{$env:GOWORK=$oldWork}
