$ErrorActionPreference='Stop'
$fixture=Join-Path $env:TICKETS_TEST_ROOT ('ui-'+[Guid]::NewGuid().ToString('N'))
if(!$env:TICKETS_TEST_EXE -or !$env:TICKETS_TEST_ROOT -or !$env:TICKETS_TEST_EVIDENCE){throw 'Set isolated test paths first'}
if((Get-Volume -DriveLetter ([IO.Path]::GetPathRoot($fixture).Substring(0,1))).FileSystem -ne 'NTFS'){throw 'UI fixture must be NTFS'}
New-Item -ItemType Directory -Force $fixture | Out-Null
Get-ChildItem Env:WORKSPACE_* | ForEach-Object {[Environment]::SetEnvironmentVariable($_.Name,$null,'Process')}
$env:TICKETS_TEST_PASSWORD=[Guid]::NewGuid().ToString('N')+'Aa!8'
$env:WORKSPACE_BOOTSTRAP_PASSWORD=$env:TICKETS_TEST_PASSWORD
$env:TICKETS_TEST_URL='http://localhost:19458'
$env:TICKETS_BACKUP_PASSWORD=[Guid]::NewGuid().ToString('N')+'Synthetic!Backup'
$env:TICKETS_BACKUP_FILE=$fixture+'/recovery.owbackup'
if(Get-NetTCPConnection -LocalPort 19458 -State Listen -ErrorAction SilentlyContinue){throw 'UI test port occupied'}
$process=$null
try{
 $launch=@('--data',($fixture+'/data'),'--listen','127.0.0.1:19458','--origin',$env:TICKETS_TEST_URL,'--no-browser','--bootstrap-email','native-owner@example.test')
 $process=Start-Process -FilePath $env:TICKETS_TEST_EXE -ArgumentList $launch -WorkingDirectory $fixture -WindowStyle Hidden -PassThru -RedirectStandardOutput ($fixture+'/stdout.log') -RedirectStandardError ($fixture+'/stderr.log')
 $ready=$false
 for($i=0;$i -lt 90;$i++){if($process.HasExited){throw ('Native UI server exited; inspect '+$fixture+'/stderr.log')};try{$response=Invoke-RestMethod ($env:TICKETS_TEST_URL+'/readyz') -NoProxy -TimeoutSec 3;if($response.status -eq 'ready' -or $response.status -eq 'ok'){$ready=$true;break}}catch{};Start-Sleep -Milliseconds 500}
 if(!$ready){throw 'UI server readiness timed out'}
 node (Join-Path $PSScriptRoot 'test-native-ui.mjs')
 if($LASTEXITCODE){throw 'Native Chrome UI acceptance failed'}
 node (Join-Path $PSScriptRoot 'test-native-api.mjs')
 if($LASTEXITCODE){throw 'Native API and concurrency acceptance failed'}
 $process.Kill();$process.WaitForExit();$process=$null
 $env:WORKSPACE_BACKUP_PASSWORD=$env:TICKETS_BACKUP_PASSWORD
 & $env:TICKETS_TEST_EXE --data ($fixture+'/restored') --restore $env:TICKETS_BACKUP_FILE --no-browser *> ($fixture+'/restore.log')
 if($LASTEXITCODE){throw 'Native encrypted restore failed'}
 $launch=@('--data',($fixture+'/restored'),'--listen','127.0.0.1:19458','--origin',$env:TICKETS_TEST_URL,'--no-browser')
 $process=Start-Process -FilePath $env:TICKETS_TEST_EXE -ArgumentList $launch -WorkingDirectory $fixture -WindowStyle Hidden -PassThru -RedirectStandardOutput ($fixture+'/restored-stdout.log') -RedirectStandardError ($fixture+'/restored-stderr.log')
 for($i=0;$i -lt 90;$i++){try{$response=Invoke-RestMethod ($env:TICKETS_TEST_URL+'/readyz') -NoProxy -TimeoutSec 3;if($response.status -eq 'ready' -or $response.status -eq 'ok'){break}}catch{};Start-Sleep -Milliseconds 500}
 node (Join-Path $PSScriptRoot 'test-native-restored.mjs')
 if($LASTEXITCODE){throw 'Restored application persistence failed'}
}finally{
 if($process -and !$process.HasExited){$process.Kill();$process.WaitForExit()}
 $env:WORKSPACE_BOOTSTRAP_PASSWORD=$null;$env:TICKETS_TEST_PASSWORD=$null
 $env:WORKSPACE_BACKUP_PASSWORD=$null;$env:TICKETS_BACKUP_PASSWORD=$null
}
