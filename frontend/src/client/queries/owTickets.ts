import {mutationOptions, queryOptions, useMutation} from '@tanstack/vue-query'
import {
	labelsList,
	labelsCreate,
	taskLabelsCreate,
	taskLabelsDelete,
	tasksRelationsCreate,
	tasksRelationsDelete,
	ticketsAttachmentsList,
	ticketsAttachmentsUpload,
	ticketsAttachmentsDownload,
	ticketsAttachmentsState,
	ticketsGitSettingsRead,
	ticketsGitSettingsUpdate,
	ticketsGitResync,
	ticketsGitLinks,
	ticketsIssueMove,
	projectUsersList,
	projectUsersCreate,
	projectUsersDelete,
	projectUsersUpdate,
	ticketsCatalog,
	ticketsQuery,
	ticketsIssuesCreate,
	ticketsIssuesRead,
	ticketsIssuesUpdate,
	ticketsIssuesDelete,
	ticketsIssueActivity,
	ticketsIssueDuplicate,
	ticketsBulk,
	ticketsProjectCreate,
	ticketsProjectSettingsUpdate,
	ticketsSprintsList,
	ticketsSprintsCreate,
	ticketsSprintStart,
	ticketsSprintClose,
	ticketsDashboardsList,
	ticketsDashboardsCreate,
	ticketsDashboardsUpdate,
	ticketsDashboardsDelete,
	ticketsDashboardData,
	ticketsDashboardLayout,
	ticketsDashboardTemplate,
	ticketsViewsList,
	ticketsViewsCreate,
	ticketsViewsDelete,
	ticketsReport,
	taskCommentsList,
	taskCommentsCreate,
	projectsUsersSearch,
	subscriptionsCreate,
	subscriptionsDelete,
	notificationsList,
	notificationsMarkAllRead,
} from '@/client/generated'
import type {
	TaskRelationWritable,
	OwGitSettingsWritable,
	OwBulkChange,
	OwDashboard,
	OwFilter,
	OwIssue,
	OwIssueWritable,
	OwProjectSettings,
	OwSprintWritable,
	OwWidget,
} from '@/client/generated'

export const owKeys = {
	all: ['my-tickets'] as const,
	catalog: () => [...owKeys.all, 'catalog'] as const,
	issues: (filter: OwFilter, page: number) => [...owKeys.all, 'issues', filter, page] as const,
	issue: (key: string) => [...owKeys.all, 'issue', key] as const,
	activity: (key: string) => [...owKeys.all, 'activity', key] as const,
	dashboards: () => [...owKeys.all, 'dashboards'] as const,
	dashboard: (id: number) => [...owKeys.all, 'dashboard', id] as const,
	sprints: (project: number) => [...owKeys.all, 'sprints', project] as const,
	views: () => [...owKeys.all, 'views'] as const,
	comments: (id: number) => [...owKeys.all, 'comments', id] as const,
	users: (project: number, q: string) => [...owKeys.all, 'users', project, q] as const,
	report: (kind: string, filter: OwFilter) => [...owKeys.all, 'report', kind, filter] as const,
}

export function normalizeIssue(value: OwIssue) {
	if (!value.key || !value.task_id || !value.project_id) {
		throw new Error('The server returned an incomplete issue. Reload and try again.')
	}
	return {
		...value,
		key: value.key,
		task_id: value.task_id,
		project_id: value.project_id,
		summary: value.summary ?? '',
		type: value.type ?? 'Task',
		state: value.state ?? 'backlog',
		assignee_ids: value.assignee_ids ?? [],
		custom: value.custom ?? {},
		description: value.description ?? '',
	}
}

export type Ticket = ReturnType<typeof normalizeIssue>

export const owCatalogQuery = () => queryOptions({
	queryKey: owKeys.catalog(),
	queryFn: async () => {const data=(await ticketsCatalog()).data;return {...data,projects:data.projects?.filter(p=>p.settings)}},
})

export const owIssuesQuery = (filter: OwFilter, page = 1) => queryOptions({
	queryKey: owKeys.issues(filter, page),
	queryFn: async () => {
		const {data} = await ticketsQuery({body: {filter, page, per_page: 50}})
		return {...data, items: (data.items ?? []).map(normalizeIssue)}
	},
	refetchInterval: 15_000,
})

export const owIssueQuery = (key: string) => queryOptions({
	queryKey: owKeys.issue(key),
	queryFn: async () => normalizeIssue((await ticketsIssuesRead({path: {key}})).data),
	enabled: key.length > 0,
})

export const owActivityQuery = (key: string) => queryOptions({
	queryKey: owKeys.activity(key),
	queryFn: async () => (await ticketsIssueActivity({path: {key}})).data,
	enabled: key.length > 0,
})

export const owDashboardsQuery = () => queryOptions({
	queryKey: owKeys.dashboards(),
	queryFn: async () => (await ticketsDashboardsList({query: {per_page: 200}})).data.items ?? [],
})

export const owDashboardQuery = (id: number) => queryOptions({
	queryKey: owKeys.dashboard(id),
	queryFn: async () => (await ticketsDashboardData({path: {id}})).data,
	enabled: id > 0,
	refetchInterval: 15_000,
})

export const owSprintsQuery = (project: number) => queryOptions({
	queryKey: owKeys.sprints(project),
	queryFn: async () => (await ticketsSprintsList({query: {project_id: project, per_page: 200}})).data.items ?? [],
})

export const owViewsQuery = () => queryOptions({
	queryKey: owKeys.views(),
	queryFn: async () => (await ticketsViewsList({query: {per_page: 200}})).data.items ?? [],
})

export const owCommentsQuery = (task: number) => queryOptions({
	queryKey: owKeys.comments(task),
	queryFn: async () => (await taskCommentsList({path: {task}, query: {per_page: 100}})).data.items ?? [],
	enabled: task > 0,
})

export const owUsersQuery = (project: number, q: string) => queryOptions({
	queryKey: owKeys.users(project, q),
	queryFn: async () => (await projectsUsersSearch({path: {project}, query: {q}})).data.items ?? [],
	enabled: project > 0,
})

export const owReportQuery = (kind: string, filter: OwFilter) => queryOptions({
	queryKey: owKeys.report(kind, filter),
	queryFn: async () => (await ticketsReport({path: {kind}, body: filter})).data,
	enabled: kind.length > 0,
})

// Invalidation stays in the mutation lifecycle. Server-side SDK sessions fence identity changes.
const write = <TInput, TResult>(fn: (input: TInput) => Promise<TResult>) => mutationOptions({
	mutationFn: fn,
	onSettled: async (_data, _error, _input, _context, {client}) => {
		await client.invalidateQueries({queryKey: owKeys.all})
	},
})

export const useOwCreateIssue = () => useMutation(write(async (body: OwIssueWritable) =>
	normalizeIssue((await ticketsIssuesCreate({body})).data)))

export const useOwSaveIssue = () => useMutation(write(async (issue: Ticket) =>
	normalizeIssue((await ticketsIssuesUpdate({path: {key: issue.key}, body: {...issue,expected_revision:issue.revision,expected_updated:issue.updated}})).data)))

export const useOwArchiveIssue = () => useMutation(write(async (key: string) => ticketsIssuesDelete({path: {key}})))
export const useOwDuplicateIssue = () => useMutation(write(async (key: string) =>
	normalizeIssue((await ticketsIssueDuplicate({path: {key}})).data)))
export const useOwBulk = () => useMutation(write(async (body: OwBulkChange) => (await ticketsBulk({body})).data))

export const useOwCreateProject = () => useMutation(write(async (body: {title: string, key: string}) =>
	(await ticketsProjectCreate({body})).data))
export const useOwSaveProject = () => useMutation(write(async (body: OwProjectSettings) =>
	(await ticketsProjectSettingsUpdate({path: {id: body.project_id!}, body})).data))

export const useOwCreateSprint = () => useMutation(write(async (body: OwSprintWritable) =>
	(await ticketsSprintsCreate({body})).data))
export const useOwSprintLifecycle = () => useMutation(write(async (input: {id: number, action: 'start' | 'close', next: number}) =>
	(await (input.action === 'start' ? ticketsSprintStart : ticketsSprintClose)({path: {id: input.id}, body: {next_sprint_id: input.next}})).data))

export const useOwTemplate = () => useMutation(write(async (input: {name: 'Personal' | 'Project Overview' | 'Engineering' | 'Sprint Overview', project: number}) =>
	(await ticketsDashboardTemplate({path: {name: input.name}, body: {project_id: input.project}})).data))
export const useOwCreateDashboard = () => useMutation(write(async (body: OwDashboard) =>
	(await ticketsDashboardsCreate({body})).data))
export const useOwSaveDashboard = () => useMutation(write(async (body: OwDashboard) =>
	(await ticketsDashboardsUpdate({path: {id: body.id!}, body})).data))
export const useOwDeleteDashboard = () => useMutation(write(async (id: number) => ticketsDashboardsDelete({path: {id}})))
export const useOwSaveLayout = () => useMutation(write(async (input: {id: number, widgets: OwWidget[]}) =>
	ticketsDashboardLayout({path: {id: input.id}, body: {widgets: input.widgets}})))

export const useOwSaveView = () => useMutation(write(async (body: {name: string, shared: boolean, filter: OwFilter}) =>
	(await ticketsViewsCreate({body})).data))
export const useOwDeleteView = () => useMutation(write(async (id: number) => ticketsViewsDelete({path: {id}})))
export const useOwAddComment = () => useMutation(write(async (input: {task: number, comment: string}) =>
	(await taskCommentsCreate({path: {task: input.task}, body: {comment: input.comment}})).data))

export const owNotificationsQuery = () => queryOptions({queryKey: [...owKeys.all, 'notifications'], queryFn: async () => (await notificationsList()).data.items ?? [], refetchInterval: 30_000})
export const useOwReadNotifications = () => useMutation(write(async () => notificationsMarkAllRead()))
export const useOwWatch = () => useMutation(write(async (input: {id: number, watch: boolean}) => (input.watch ? subscriptionsCreate : subscriptionsDelete)({path: {entity: 'task', entityID: input.id}})))

export const owAttachmentsQuery = (key: string) => queryOptions({queryKey:[...owKeys.issue(key),'attachments'],queryFn:async () => (await ticketsAttachmentsList({path:{key}})).data,enabled:!!key})
export const useOwUpload = () => useMutation(write(async (input:{key:string,file:File}) => (await ticketsAttachmentsUpload({path:{key:input.key},body:{file:input.file}})).data))
export const useOwAttachmentState = () => useMutation(write(async (input:{id:string,deleted:boolean}) => ticketsAttachmentsState({path:{id:input.id},body:{deleted:input.deleted}})))
export async function downloadOwAttachment(id:string,name:string){const result=await ticketsAttachmentsDownload({path:{id},parseAs:'blob'});const url=URL.createObjectURL(result.data as Blob);const link=document.createElement('a');link.href=url;link.download=name;link.click();setTimeout(()=>URL.revokeObjectURL(url),10_000)}
export const owGitSettingsQuery = (id:number) => queryOptions({queryKey:[...owKeys.all,'git-settings',id],queryFn:async()=>(await ticketsGitSettingsRead({path:{id}})).data,enabled:id>0})
export const owGitLinksQuery = (key:string) => queryOptions({queryKey:[...owKeys.issue(key),'git'],queryFn:async()=>(await ticketsGitLinks({path:{key}})).data,enabled:!!key})
export const useOwGitSettings = () => useMutation(write(async (body:OwGitSettingsWritable)=>(await ticketsGitSettingsUpdate({path:{id:body.project_id!},body})).data))
export const useOwGitResync = () => useMutation(write(async (id:number)=>(await ticketsGitResync({path:{id}})).data))
export const useOwMove = () => useMutation(write(async (input:{key:string,project:number})=>normalizeIssue((await ticketsIssueMove({path:{key:input.key},body:{project_id:input.project}})).data)))
export const owMembersQuery = (project:number) => queryOptions({queryKey:[...owKeys.all,'members',project],queryFn:async()=>(await projectUsersList({path:{project}})).data.items ?? [],enabled:project>0})
export const owLabelsQuery = () => queryOptions({queryKey:[...owKeys.all,'labels'],queryFn:async()=>(await labelsList()).data.items ?? []})
export const useOwLabel = () => useMutation(write(async (input:{task:number,id?:number,title?:string,remove?:boolean})=>{
	if(input.remove && input.id) return taskLabelsDelete({path:{task:input.task,label:input.id}})
	const label=input.id ?? (await labelsCreate({body:{title:input.title,hex_color:'3157D5'}})).data.id!
	return taskLabelsCreate({path:{task:input.task},body:{label_id:label}})
}))
export const useOwRelation = () => useMutation(write(async (input:{task:number,other:number,kind:TaskRelationWritable['relation_kind'],remove?:boolean})=>{
	if(input.remove) return tasksRelationsDelete({path:{task:input.task,otherTask:input.other,relationKind:input.kind!}})
	return tasksRelationsCreate({path:{task:input.task},body:{other_task_id:input.other,relation_kind:input.kind}})
}))
export const useOwMember = () => useMutation(write(async (input:{project:number,username:string,permission:number,action:'add'|'update'|'remove'})=>{
	if(input.action==='add') return projectUsersCreate({path:{project:input.project},body:{username:input.username,permission:input.permission}})
	if(input.action==='remove') return projectUsersDelete({path:{project:input.project,user:input.username}})
	return projectUsersUpdate({path:{project:input.project,user:input.username},body:{permission:input.permission}})
}))
