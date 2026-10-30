<script setup lang="ts">
import {computed, nextTick, ref, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useQuery} from '@tanstack/vue-query'
import type {OwFilter} from '@/client/generated'
import type {Ticket} from '@/client/queries/owTickets'
import {owCatalogQuery, owIssuesQuery, owSprintsQuery, owUsersQuery, owDashboardsQuery, owDashboardQuery, owViewsQuery, owNotificationsQuery, owReportQuery, useOwCreateProject, useOwCreateIssue, useOwSaveIssue, useOwBulk, useOwCreateSprint, useOwSprintLifecycle, useOwTemplate, useOwSaveView, useOwDeleteView, useOwReadNotifications} from '@/client/queries/owTickets'
import type {Bootstrap} from './main'
import type {NativeSession} from '@/client/native'
import {base, notice} from './main'
import DashboardView from './DashboardView.vue'
import IssueDetail from './IssueDetail.vue'
import IssueFields from './IssueFields.vue'
import FilterBuilder from './FilterBuilder.vue'
import ProjectSettings from './ProjectSettings.vue'
import GitSettings from './GitSettings.vue'
import ProjectAccess from './ProjectAccess.vue'
import DashboardSettings from './DashboardSettings.vue'
import WidgetChart from './WidgetChart.vue'
import {displayDate, dateValue} from './safeHtml'
const props = defineProps<{bootstrap: Bootstrap, session: NativeSession}>()
const router = useRouter(), route = useRoute()
const catalog = useQuery(owCatalogQuery()), dashboards = useQuery(owDashboardsQuery()), views = useQuery(owViewsQuery())
const notifications = useQuery(owNotificationsQuery()), markRead = useOwReadNotifications()
const section = ref('Issues'), filter = ref<OwFilter>({sort: 'updated', descending: true, sprint_id: 0}), page = ref(1), projectID = ref(0), dashboardID = ref(0)
const project = computed(() => catalog.data.value?.projects?.find(p => p.project?.id === projectID.value))
const settings = computed(() => project.value?.settings)
const canEdit = computed(() => (project.value?.permission ?? 0) >= 1)
const issues = useQuery(computed(() => owIssuesQuery({...filter.value, project_id: projectID.value, sort: ['Board','Backlog'].includes(section.value) ? 'position' : filter.value.sort}, page.value)))
const allEpics = useQuery(computed(() => owIssuesQuery({project_id: projectID.value, rules: [{field:'type',operator:'eq',value:'Epic'}]})))
const sprints = useQuery(computed(() => owSprintsQuery(projectID.value)))
const users = useQuery(computed(() => owUsersQuery(projectID.value, '')))
const dashboard = useQuery(computed(() => owDashboardQuery(dashboardID.value)))
const reportKind = ref('created-resolved'), report = useQuery(computed(() => owReportQuery(section.value === 'Reports' ? reportKind.value : '', {...filter.value, project_id: projectID.value})))
const createProject = useOwCreateProject(), createIssue = useOwCreateIssue(), saveIssue = useOwSaveIssue(), bulk = useOwBulk(), createSprint = useOwCreateSprint(), lifecycle = useOwSprintLifecycle(), template = useOwTemplate(), saveView = useOwSaveView(), deleteView = useOwDeleteView()
const dialog = ref<HTMLDialogElement>(), dialogType = ref('issue'), fields = ref<InstanceType<typeof IssueFields>>()
const title = ref(''), key = ref(''), sprintGoal = ref(''), sprintStart = ref(''), sprintEnd = ref(''), viewName = ref(''), selected = ref<string[]>([]), bulkAction = ref('state'), bulkValue = ref('selected'), dragging = ref('')
const issueKey = computed(() => route.path.startsWith('/browse/') ? decodeURIComponent(route.path.slice(8)) : '')
const draft = ref<Ticket>({key: '', task_id: 0, project_id: 0, summary: '', description: '', type: 'Task', state: 'backlog', assignee_ids: [], custom: {}, priority: 2, points: 0, estimate: 0, sprint_id: 0, epic_id: 0})
watch(projectID, () => {page.value = 1; selected.value = []; filter.value = {...filter.value, sprint_id: 0}})
watch(() => filter.value, () => {page.value = 1}, {deep: true})
watch(() => dashboards.data.value, value => {if (!dashboardID.value && value?.length) dashboardID.value = value[0].id!})
function openIssue(value: string) {void router.push(`/browse/${encodeURIComponent(value)}`)}
function navigate(value: string) {section.value = value; void router.push('/')}
async function beginCreate() {
	if (busy.value) return
	const eligible=catalog.data.value?.projects?.find(p=>p.settings && (p.permission ?? 0)>=1)
	if (!projectID.value && eligible) {projectID.value=eligible.project!.id!;await nextTick()}
	show(settings.value ? 'issue' : 'project')
}
function show(type: string) {
	if (busy.value) return
	dialogType.value = type; title.value = ''; key.value = ''
	if (type === 'issue') draft.value = {key:'',task_id:0,project_id:projectID.value,summary:'',description:'',type:'Task',severity:'',state:settings.value?.states?.[0]?.id ?? 'backlog',assignee_ids:[],custom:{},priority:2,points:0,estimate:0,sprint_id:0,epic_id:0}
	dialog.value?.showModal()
}
async function submit() {
	if (busy.value) return
	if (dialogType.value === 'project') {const p = await createProject.mutateAsync({title:title.value,key:key.value.toUpperCase()}); projectID.value = p.project_id!}
	if (dialogType.value === 'issue' && fields.value) {const i = await createIssue.mutateAsync(fields.value.getDraft()); dialog.value?.close(); openIssue(i.key); return}
	if (dialogType.value === 'sprint') await createSprint.mutateAsync({project_id:projectID.value,name:title.value,goal:sprintGoal.value,start:dateValue(sprintStart.value),end:dateValue(sprintEnd.value)})
	dialog.value?.close()
}
async function moveIssue(value: string, state: string) {const issue = issues.data.value?.items.find(i => i.key === value); if (!issue || saveIssue.isPending.value) return; await saveIssue.mutateAsync({...issue,state}); dragging.value = ''}
async function reorder(issue: Ticket, index: number, direction: number) {const next = issues.data.value?.items[index + direction]; if (!next) return; const adjacent = issues.data.value?.items[index + 2 * direction]; const position = adjacent ? ((next.position ?? 0) + (adjacent.position ?? 0)) / 2 : (next.position ?? 0) + direction * 1024; await saveIssue.mutateAsync({...issue,position})}
async function applyBulk() {if (!selected.value.length || bulk.isPending.value) return; await bulk.mutateAsync({keys:selected.value,action:bulkAction.value as 'state'|'priority'|'sprint'|'archive'|'duplicate',value:bulkValue.value}); selected.value = []}
async function addTemplate(name: string) {const d = await template.mutateAsync({name:name as 'Personal',project:projectID.value}); dashboardID.value = d.id!; section.value = 'Dashboards'}
async function logout() {await fetch(`${base}/auth/logout`, {method:'POST', credentials:'same-origin',headers:{'X-CSRF-Token':props.session.csrf}}); window.location.replace(`${base}/`)}
function apps() {window.dispatchEvent(new CustomEvent('workspace:open-apps'))}
const busy = computed(() => createProject.isPending.value || createIssue.isPending.value || createSprint.isPending.value)
</script>

<template>
	<div class="ow-shell">
		<header class="ow-topbar">
			<a
				:href="`${base}/`"
				class="ow-brand"
			><img
				:src="`${base}/tickets.svg`"
				width="38"
				height="38"
				alt=""
			><span>My Tickets<small>One Workspace</small></span></a><div class="ow-global-search">
				<span aria-hidden="true">⌕</span><input
					v-model="filter.q"
					aria-label="Search issues"
					placeholder="Search issues by key, title or description"
					@input="navigate('Issues')"
				>
			</div><details class="ow-menu">
				<summary
					class="ow-button"
					aria-label="Notifications"
				>
					♧ <span
						v-if="notifications.data.value?.some(n => !n.read_at || n.read_at.startsWith('0001-'))"
						class="ow-dot"
					/>
				</summary><div class="ow-card ow-notifications">
					<h3>Notifications</h3><button @click="markRead.mutate()">
						Mark all read
					</button><p
						v-for="item in notifications.data.value"
						:key="item.id"
					>
						{{ item.name }}<small>{{ displayDate(item.created) }}</small>
					</p><p v-if="!notifications.data.value?.length">
						You're all caught up.
					</p>
				</div>
			</details><button
				data-workspace-apps-trigger
				aria-label="Workspace apps"
				@click="apps"
			>
				▦
			</button><details class="ow-menu">
				<summary
					class="ow-avatar"
					:aria-label="session.user.name || session.user.email"
				>
					{{ (session.user.name || session.user.email).slice(0,1).toUpperCase() }}
				</summary><div class="ow-card">
					<strong>{{ session.user.name }}</strong><small>{{ session.user.email }}</small><small>Username: {{ catalog.data.value?.user?.username }}</small><button @click="logout">
						Sign out
					</button>
				</div>
			</details>
		</header>
		<aside class="ow-sidebar">
			<button
				class="ow-primary ow-create"
				@click="beginCreate()"
			>
				＋ {{ catalog.data.value?.projects?.some(p => (p.permission ?? 0) >= 1) ? 'Create issue' : 'Create project' }}
			</button><nav aria-label="My Tickets">
				<button
					v-for="(item,i) in ['Issues','Board','Backlog','Sprints','Roadmap','Dashboards','Reports','Project settings']"
					:key="item"
					:aria-current="section === item && !issueKey ? 'page' : undefined"
					@click="navigate(item)"
				>
					<span aria-hidden="true">{{ ['☷','▥','≡','◷','◇','▦','▤','⚙'][i] }}</span>{{ item }}
				</button>
			</nav><hr><label>Project<select v-model.number="projectID"><option :value="0">All projects</option><option
				v-for="p in catalog.data.value?.projects"
				:key="p.project?.id"
				:value="p.project?.id"
			>{{ p.settings?.key }} · {{ p.project?.title }}</option></select></label><button
				class="ow-text-button"
				@click="show('project')"
			>
				＋ New project
			</button><div v-if="views.data.value?.length">
				<h3 class="ow-eyebrow">
					Saved views
				</h3><div
					v-for="view in views.data.value"
					:key="view.id"
					class="ow-saved"
				>
					<button @click="filter = {...view.filter}; projectID = view.filter?.project_id ?? 0; navigate('Issues')">
						{{ view.name }}
					</button><button
						v-if="view.owner_id === catalog.data.value?.user?.id"
						aria-label="Delete saved view"
						@click="deleteView.mutate(view.id!)"
					>
						×
					</button>
				</div>
			</div><footer>
				<a
					href="https://github.com/oneworkspacestudio/my-tickets"
					target="_blank"
					rel="noopener"
				>Source & licences</a><small>{{ bootstrap.version }}</small>
			</footer>
		</aside>
		<main class="ow-main">
			<div
				v-if="notice"
				class="ow-error ow-notice"
				role="alert"
			>
				<span>{{ notice }}</span><button
					aria-label="Dismiss message"
					@click="notice = ''"
				>
					×
				</button>
			</div><p
				v-if="catalog.isPending.value"
				role="status"
			>
				Opening your workspace…
			</p><IssueDetail
				v-else-if="issueKey"
				:key="issueKey"
				:issue-key="issueKey"
				:projects="catalog.data.value?.projects ?? []"
				@close="router.push('/')"
				@open="openIssue"
			/>
			<template v-else>
				<div class="ow-heading">
					<div>
						<p class="ow-eyebrow">
							{{ project?.project?.title || 'Across your projects' }}
						</p><h1>{{ section }}</h1>
					</div><button
						v-if="canEdit && settings && ['Issues','Board','Backlog'].includes(section)"
						class="ow-primary"
						@click="show('issue')"
					>
						＋ Create issue
					</button>
				</div>
				<div
					v-if="!catalog.data.value?.projects?.length"
					class="ow-empty ow-card"
				>
					<img
						:src="`${base}/tickets.svg`"
						width="80"
						height="80"
						alt=""
					><h2>A clear place for your next project</h2><p>Create a project, plan work and track progress together.</p><button
						class="ow-primary"
						@click="show('project')"
					>
						Create your first project
					</button>
				</div>
				<template v-else-if="['Issues','Board','Backlog','Roadmap'].includes(section)">
					<div class="ow-filterbar">
						<FilterBuilder
							v-model="filter"
							:settings="settings ?? undefined"
						/><label class="ow-inline-label">Sprint<select v-model.number="filter.sprint_id"><option :value="0">Any sprint</option><option
							v-for="s in sprints.data.value"
							:key="s.id"
							:value="s.id"
						>{{ s.name }}</option></select></label><label class="ow-check"><input
							v-model="filter.archived"
							type="checkbox"
						>Include archived</label><label class="ow-inline-label">Sort<select v-model="filter.sort"><option
							v-for="sort in ['key','summary','priority','due_date','created','updated']"
							:key="sort"
							:value="sort"
						>{{ sort.replaceAll('_',' ') }}</option></select></label><details class="ow-menu">
							<summary class="ow-button">
								Save view
							</summary><form
								class="ow-card"
								@submit.prevent="saveView.mutate({name:viewName,shared:false,filter:{...filter,project_id:projectID}})"
							>
								<label>Name<input
									v-model="viewName"
									required
									maxlength="250"
								></label><button :disabled="saveView.isPending.value">
									Save personal view
								</button>
							</form>
						</details><span class="ow-caption">{{ issues.data.value?.total ?? 0 }} matching issues</span>
					</div><p
						v-if="issues.isPending.value"
						role="status"
					>
						Loading issues…
					</p><div
						v-if="selected.length"
						class="ow-bulk ow-actions"
					>
						<strong>{{ selected.length }} selected</strong><select
							v-model="bulkAction"
							aria-label="Bulk action"
						>
							<option
								v-for="action in ['state','priority','sprint','archive','duplicate']"
								:key="action"
							>
								{{ action }}
							</option>
						</select><input
							v-if="!['archive','duplicate'].includes(bulkAction)"
							v-model="bulkValue"
							aria-label="Bulk action value"
							placeholder="State, priority or sprint ID"
						><button
							:disabled="bulk.isPending.value"
							@click="applyBulk"
						>
							Apply to selected
						</button><button @click="selected = []">
							Clear
						</button>
					</div>
					<div
						v-if="section === 'Board' && settings"
						class="ow-board"
					>
						<section
							v-for="state in settings.states"
							:key="state.id"
							class="ow-column"
							@dragover.prevent
							@drop.prevent="moveIssue(dragging, state.id!)"
						>
							<header><h2>{{ state.name }}</h2><span class="ow-count">{{ issues.data.value?.items.filter(i => i.state === state.id).length }}{{ state.limit ? ` / ${state.limit}` : '' }}</span></header><article
								v-for="issue in issues.data.value?.items.filter(i => i.state === state.id)"
								:key="issue.key"
								class="ow-card ow-issue-card"
								:draggable="canEdit"
								@dragstart="dragging = issue.key"
							>
								<span class="ow-type">{{ issue.type }}</span><a
									:href="`${base}/browse/${issue.key}`"
									@click.prevent="openIssue(issue.key)"
								>{{ issue.summary }}</a><footer><span class="ow-key">{{ issue.key }}</span><span>{{ issue.points || '' }}</span></footer><select
									v-if="canEdit"
									:value="issue.state"
									aria-label="Move issue to status"
									@change="moveIssue(issue.key, ($event.target as HTMLSelectElement).value)"
								>
									<option
										v-for="s in settings.states"
										:key="s.id"
										:value="s.id"
									>
										{{ s.name }}
									</option>
								</select>
							</article>
						</section>
					</div>
					<div
						v-else-if="section === 'Roadmap'"
						class="ow-roadmap"
					>
						<article
							v-for="epic in allEpics.data.value?.items"
							:key="epic.key"
							class="ow-card"
						>
							<header>
								<span class="ow-badge">Epic</span><a
									:href="`${base}/browse/${epic.key}`"
									@click.prevent="openIssue(epic.key)"
								>{{ epic.key }} · {{ epic.summary }}</a>
							</header><p>Target: {{ epic.fix_version || 'Not scheduled' }} · {{ displayDate(epic.due_date) }}</p><a
								v-for="child in issues.data.value?.items.filter(i => i.epic_id === epic.task_id)"
								:key="child.key"
								:href="`${base}/browse/${child.key}`"
								@click.prevent="openIssue(child.key)"
							>{{ child.key }} {{ child.summary }}<span class="ow-badge">{{ child.state }}</span></a>
						</article><p v-if="!allEpics.data.value?.items.length">
							Create an Epic and link issues to plan a larger goal.
						</p>
					</div>
					<div
						v-else
						class="ow-table-wrap ow-card"
					>
						<table class="ow-issue-table">
							<thead>
								<tr>
									<th v-if="canEdit">
										<input
											type="checkbox"
											aria-label="Select all displayed issues"
											:checked="!!issues.data.value?.items.length && selected.length === issues.data.value.items.length"
											@change="selected = ($event.target as HTMLInputElement).checked ? issues.data.value?.items.map(i => i.key) ?? [] : []"
										>
									</th><th>Issue</th><th>Status</th><th>Priority</th><th>Assignee</th><th>Due</th><th v-if="section === 'Backlog'">
										Order
									</th>
								</tr>
							</thead><tbody>
								<tr
									v-for="(issue,index) in issues.data.value?.items"
									:key="issue.key"
								>
									<td v-if="canEdit">
										<input
											v-model="selected"
											type="checkbox"
											:value="issue.key"
											:aria-label="`Select ${issue.key}`"
										>
									</td><td>
										<span class="ow-type">{{ issue.type }} <span v-if="issue.archived">· Archived</span></span><a
											:href="`${base}/browse/${issue.key}`"
											@click.prevent="openIssue(issue.key)"
										><span class="ow-key">{{ issue.key }}</span>{{ issue.summary }}</a>
									</td><td><span class="ow-badge">{{ settings?.states?.find(s => s.id === issue.state)?.name || issue.state }}</span></td><td>{{ ['—','Low','Normal','High','Urgent','Immediate'][issue.priority ?? 0] }}</td><td>{{ users.data.value?.filter(u => issue.assignee_ids.includes(u.id!)).map(u => u.name || u.username).join(', ') || (issue.assignee_ids.length ? `${issue.assignee_ids.length} assigned` : 'Unassigned') }}</td><td>{{ displayDate(issue.due_date) }}</td><td
										v-if="section === 'Backlog'"
										class="ow-actions"
									>
										<button
											aria-label="Move issue up"
											:disabled="index === 0 || !canEdit"
											@click="reorder(issue,index,-1)"
										>
											↑
										</button><button
											aria-label="Move issue down"
											:disabled="index === issues.data.value!.items.length - 1 || !canEdit"
											@click="reorder(issue,index,1)"
										>
											↓
										</button>
									</td>
								</tr>
							</tbody>
						</table><div
							v-if="!issues.isPending.value && !issues.data.value?.items.length"
							class="ow-empty"
						>
							<h2>No matching issues</h2><p>Try clearing filters, or create an issue to get started.</p><button @click="filter = {sort:'updated',descending:true}">
								Clear filters
							</button>
						</div>
					</div><footer class="ow-pagination">
						<button
							:disabled="page === 1"
							@click="page--"
						>
							Previous
						</button><span>Page {{ page }} · 50 per page</span><button
							:disabled="page * 50 >= (issues.data.value?.total ?? 0)"
							@click="page++"
						>
							Next
						</button>
					</footer>
				</template>
				<template v-else-if="section === 'Dashboards'">
					<div class="ow-actions">
						<label class="ow-inline-label">Dashboard<select v-model.number="dashboardID"><option :value="0">Choose dashboard</option><option
							v-for="d in dashboards.data.value"
							:key="d.id"
							:value="d.id"
						>{{ d.name }}</option></select></label><details class="ow-menu">
							<summary class="ow-button">
								＋ From template
							</summary><div class="ow-card">
								<button
									v-for="name in catalog.data.value?.templates"
									:key="name"
									@click="addTemplate(name)"
								>
									{{ name }}
								</button>
							</div>
						</details><DashboardSettings
							v-if="dashboard.data.value?.dashboard?.owner_id === catalog.data.value?.user?.id"
							:dashboard="dashboard.data.value!.dashboard!"
							:projects="catalog.data.value?.projects ?? []"
							:sprints="sprints.data.value ?? []"
							:owner="session.user.role === 'owner'"
						/>
					</div><DashboardView
						v-if="dashboard.data.value"
						:data="dashboard.data.value"
						:can-edit="dashboard.data.value.dashboard?.owner_id === catalog.data.value?.user?.id"
						:kinds="catalog.data.value?.widget_kinds ?? []"
					/><div
						v-else
						class="ow-empty ow-card"
					>
						<h2>Your work, at a glance</h2><p>Start with a dashboard template, then arrange widgets for your team.</p><button
							class="ow-primary"
							@click="addTemplate('Personal')"
						>
							Create personal dashboard
						</button>
					</div>
				</template>
				<template v-else-if="section === 'Sprints'">
					<p v-if="!projectID">
						Choose a project to plan a sprint.
					</p><button
						v-if="canEdit"
						class="ow-primary"
						@click="show('sprint')"
					>
						＋ Create sprint
					</button><article
						v-for="s in sprints.data.value"
						:key="s.id"
						class="ow-card ow-sprint"
					>
						<header><h2>{{ s.name }}</h2><span class="ow-badge">{{ s.state }}</span></header><p>{{ s.goal }}</p><p>{{ displayDate(s.start) }} — {{ displayDate(s.end) }}</p><p v-if="s.state === 'closed'">
							{{ s.completed?.length }} completed from {{ s.committed?.length }} starting issues.
						</p><div
							v-if="canEdit"
							class="ow-actions"
						>
							<button
								v-if="s.state === 'planned'"
								@click="lifecycle.mutate({id:s.id!,action:'start',next:0})"
							>
								Start sprint
							</button><button
								v-if="s.state === 'active'"
								@click="lifecycle.mutate({id:s.id!,action:'close',next:0})"
							>
								Close sprint · unfinished work returns to backlog
							</button>
						</div>
					</article>
				</template>
				<template v-else-if="section === 'Reports'">
					<label class="ow-inline-label">Report<select v-model="reportKind"><option
						v-for="kind in catalog.data.value?.widget_kinds"
						:key="kind"
						:value="kind"
					>{{ kind.replaceAll('-',' ') }}</option></select></label><article
						v-if="report.data.value"
						class="ow-card ow-report"
					>
						<h2>{{ reportKind.replaceAll('-',' ') }}</h2><WidgetChart :data="report.data.value" /><p class="ow-caption">
							Calculated from your permitted issues. Use dashboard exports to download chart data.
						</p>
					</article>
				</template>
				<div
					v-else-if="section === 'Project settings' && settings && project?.permission === 2"
					:key="projectID"
					class="ow-settings-stack"
				>
					<ProjectSettings :settings="settings" /><ProjectAccess :project="projectID" /><GitSettings
						:project="projectID"
						:settings="settings"
					/>
				</div><p v-else-if="section === 'Project settings'">
					Choose a project you administer to configure its workflow and fields.
				</p>
			</template>
		</main>
		<dialog
			ref="dialog"
			class="ow-dialog"
			:aria-label="dialogType === 'issue' ? 'Create issue' : dialogType === 'project' ? 'Create project' : 'Create sprint'"
			@click="($event.target === dialog) && dialog.close()"
		>
			<form @submit.prevent="submit">
				<header>
					<h2>{{ dialogType === 'issue' ? 'Create issue' : dialogType === 'project' ? 'Create project' : 'Create sprint' }}</h2><button
						type="button"
						aria-label="Close dialog"
						@click="dialog?.close()"
					>
						×
					</button>
				</header><div class="ow-dialog-body">
					<IssueFields
						v-if="dialogType === 'issue' && settings"
						ref="fields"
						:initial="draft"
						:settings="settings"
						:sprints="sprints.data.value ?? []"
						:users="users.data.value ?? []"
						:epics="allEpics.data.value?.items ?? []"
					/><template v-else>
						<label>Name<input
							v-model="title"
							required
							maxlength="250"
						></label><label v-if="dialogType === 'project'">Project key<input
							v-model="key"
							required
							pattern="[a-zA-Z][a-zA-Z0-9]{1,9}"
							maxlength="10"
							placeholder="WORK"
						><small>2–10 letters and numbers. Permanent IDs, such as WORK-1.</small></label><template v-else>
							<label>Sprint goal<textarea v-model="sprintGoal" /></label><div class="ow-form-grid">
								<label>Start<input
									v-model="sprintStart"
									type="date"
									required
								></label><label>End<input
									v-model="sprintEnd"
									type="date"
									required
								></label>
							</div>
						</template>
					</template>
				</div><footer>
					<button
						type="button"
						@click="dialog?.close()"
					>
						Cancel
					</button><button
						class="ow-primary"
						:aria-disabled="busy"
					>
						{{ busy ? 'Creating…' : 'Create' }}
					</button>
				</footer>
			</form>
		</dialog>
	</div>
</template>
