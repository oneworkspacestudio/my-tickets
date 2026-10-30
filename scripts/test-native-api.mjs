import {createRequire} from 'node:module'
import {writeFile} from 'node:fs/promises'
import assert from 'node:assert/strict'
import {randomBytes} from 'node:crypto'
const require=createRequire(process.env.VIKUNJA_FRONTEND_DIR+'/package.json')
const {request}=require('@playwright/test')
const base=process.env.TICKETS_TEST_URL,password=process.env.TICKETS_TEST_PASSWORD,out=process.env.TICKETS_TEST_EVIDENCE
if(!base?.startsWith('http://localhost:')||!password||!out)throw Error('Isolated fixture required')
const owner=await request.newContext({baseURL:base}),peer=await request.newContext({baseURL:base}),agent=await request.newContext({baseURL:base})
async function ok(response,status=200){assert.equal(response.status(),status,`Unexpected status ${response.status()}`);return response}
async function body(response,status=200){return (await ok(response,status)).json()}
const memberPassword=randomBytes(24).toString('hex')
try{
 await ok(await owner.post('/api/login',{headers:{Origin:base},data:{email:'native-owner@example.test',password}}))
 const me=await body(await owner.get('/api/me'))
 const headers={Origin:base,'X-CSRF-Token':me.csrf,'X-Workspace-Expected-Account':me.user.id}
 const catalog=await body(await owner.get('/api/v2/tickets/catalog')),project=catalog.projects.find(p=>p.settings?.key==='WORK').project.id
 const privateProject=await body(await owner.post('/api/v2/tickets/projects',{headers,data:{title:'Private authorization fixture',key:'HIDDEN'}}),201)
 const hidden=await body(await owner.post('/api/v2/tickets/issues',{headers,data:{project_id:privateProject.project_id,summary:'Hidden synthetic issue',assignee_ids:[catalog.user.id]}}),201)
 await ok(await owner.post('/api/local-users',{headers,data:{email:'ticket-viewer@example.test',password:memberPassword}}),201)
 await ok(await peer.post('/api/login',{headers:{Origin:base},data:{email:'ticket-viewer@example.test',password:memberPassword}}))
 const peerMe=await body(await peer.get('/api/me'))
 const peerHeaders={Origin:base,'X-CSRF-Token':peerMe.csrf,'X-Workspace-Expected-Account':peerMe.user.id}
 const peerCatalog=await body(await peer.get('/api/v2/tickets/catalog'))
 await ok(await peer.get('/api/v2/tickets/issues/WORK-1'),403)
 await ok(await owner.post(`/api/v2/projects/${project}/users`,{headers,data:{username:peerCatalog.user.username,permission:0}}),201)
 const visible=await body(await peer.get('/api/v2/tickets/issues/WORK-1'))
 await ok(await peer.put('/api/v2/tickets/issues/WORK-1',{headers:peerHeaders,data:{...visible,summary:'Denied viewer edit'}}),403)
 await ok(await peer.get(`/api/v2/tickets/issues/${hidden.key}`),403)
 const wrongAccount=await peer.post('/api/v2/tickets/query',{headers:{...peerHeaders,'X-Workspace-Expected-Account':me.user.id},data:{filter:{}}})
 await ok(wrongAccount,409)
 await ok(await owner.put(`/api/v2/projects/${project}/users/${peerCatalog.user.username}`,{headers,data:{permission:1}}))
 const token=await body(await peer.post('/api/v2/tokens',{headers:peerHeaders,data:{title:'Synthetic scoped acceptance agent',expires_at:new Date(Date.now()+3600000).toISOString(),permissions:{tickets_issues:['read_one','update'],mcp:['access']}}}),201)
 assert.ok(token.token?.startsWith('tk_'))
 const agentHeaders={Authorization:'Bearer '+token.token}
 const agentIssue=await body(await agent.get('/api/v2/tickets/issues/WORK-1',{headers:agentHeaders}))
 await ok(await agent.put('/api/v2/tickets/issues/WORK-1',{headers:agentHeaders,data:{...agentIssue,summary:'Review the launch checklist'}}))
 await ok(await agent.get('/api/v2/tickets/issues/'+hidden.key,{headers:agentHeaders}),403)
 await ok(await agent.post('/api/v2/tickets/issues',{headers:agentHeaders,data:{project_id:project,summary:'Denied token scope'}}),401)
 const mcp=await agent.post('/api/v2/mcp',{headers:{...agentHeaders,Accept:'application/json, text/event-stream'},data:{jsonrpc:'2.0',id:1,method:'initialize',params:{protocolVersion:'2025-03-26',capabilities:{},clientInfo:{name:'synthetic-acceptance',version:'1'}}}})
 await ok(mcp)
 // Native requests serialize the core transaction, preserving project-local counters.
 const created=await Promise.all(Array.from({length:16},(_,n)=>owner.post('/api/v2/tickets/issues',{headers,data:{project_id:project,summary:'Concurrent synthetic '+n}}).then(r=>body(r,201))))
 assert.equal(new Set(created.map(i=>i.key)).size,16)
 const old=await body(await owner.get('/api/v2/tickets/issues/WORK-1'))
 await ok(await owner.put('/api/v2/tickets/issues/WORK-1',{headers,data:{...old,summary:'Review the launch checklist · reviewed'}}))
 await ok(await owner.put('/api/v2/tickets/issues/WORK-1',{headers,data:{...old,expected_revision:old.revision,expected_updated:old.updated,summary:'Stale edit must be rejected'}}),409)
 const attachment=(await body(await owner.get('/api/v2/tickets/issues/WORK-1/attachments')))[0]
 assert.ok(attachment)
 await ok(await peer.get('/api/v2/tickets/attachments/'+attachment.id))
 await ok(await owner.put('/api/v2/tickets/attachments/'+attachment.id+'/state',{headers,data:{deleted:true}}),204)
 await ok(await peer.get('/api/v2/tickets/attachments/'+attachment.id),404)
 await ok(await owner.put('/api/v2/tickets/attachments/'+attachment.id+'/state',{headers,data:{deleted:false}}),204)
 const backup=await ok(await owner.post('/api/backups',{headers,data:{password:process.env.TICKETS_BACKUP_PASSWORD}}))
 await writeFile(process.env.TICKETS_BACKUP_FILE,await backup.body())
 await ok(await owner.post('/auth/logout',{headers:{Origin:base}}))
 await ok(await owner.get('/api/v2/tickets/issues/WORK-1'),401)
 await writeFile(out+'/api-acceptance.json',JSON.stringify({status:'PASS',ownerInstallationUsed:false,viewerRead:true,viewerWriteDenied:true,hiddenProjectDenied:true,accountFence:true,scopedAgentReadWrite:true,agentScopeCreateDenied:true,agentHiddenProjectDenied:true,mcpInitialize:true,concurrentCreates:16,uniqueKeys:16,staleEditDenied:true,attachmentDeleteRestore:true,nativeLogoutRevoked:true},null,2))
}finally{await owner.dispose();await peer.dispose();await agent.dispose()}
