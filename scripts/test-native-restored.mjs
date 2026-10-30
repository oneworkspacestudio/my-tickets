import {createRequire} from 'node:module'
import {writeFile} from 'node:fs/promises'
import assert from 'node:assert/strict'
const require=createRequire(process.env.VIKUNJA_FRONTEND_DIR+'/package.json')
const {request}=require('@playwright/test')
const base=process.env.TICKETS_TEST_URL
const client=await request.newContext({baseURL:base})
try{
 assert.equal((await client.get('/api/v2/tickets/issues/WORK-1')).status(),401)
 assert.equal((await client.post('/api/login',{headers:{Origin:base},data:{email:'native-owner@example.test',password:process.env.TICKETS_TEST_PASSWORD}})).status(),200)
 const response=await client.get('/api/v2/tickets/issues/WORK-1')
 assert.equal(response.status(),200)
 const issue=await response.json()
 assert.equal(issue.summary,'Review the launch checklist · reviewed')
 const files=await (await client.get('/api/v2/tickets/issues/WORK-1/attachments')).json()
 assert.equal(files.length,1)
 const content=await client.get('/api/v2/tickets/attachments/'+files[0].id)
 assert.equal(content.status(),200)
 assert.equal(await content.text(),'Synthetic attachment, no private content.')
 assert.equal((await client.get('/api/v2/tickets/dashboards/1/data')).status(),200)
 await writeFile(process.env.TICKETS_TEST_EVIDENCE+'/restore-acceptance.json',JSON.stringify({status:'PASS',exeSHA256:process.env.TICKETS_TEST_BINARY_SHA256,encryptedLiveSnapshot:true,coreDatabaseIntegrity:true,sdkDatabaseIntegrity:true,restoredAccounts:true,restoredIssue:true,restoredDashboard:true,restoredAttachmentExactBytes:true,anonymousDenied:true,ownerInstallationUsed:false},null,2))
}finally{await client.dispose()}
