<script setup lang="ts">
import {computed, ref} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {owIssueQuery, owActivityQuery, owCommentsQuery, owUsersQuery, owSprintsQuery, owIssuesQuery, useOwSaveIssue, useOwAddComment, useOwArchiveIssue, useOwDuplicateIssue, useOwWatch, owGitLinksQuery, useOwMove} from '@/client/queries/owTickets'
import type {OwProjectInfo} from '@/client/generated'
import IssueFields from './IssueFields.vue'
import IssueAttachments from './IssueAttachments.vue'
import IssueLinks from './IssueLinks.vue'
import RichText from './RichText.vue'
import {safeHtml, displayDate} from './safeHtml'

const props = defineProps<{issueKey: string, projects: OwProjectInfo[]}>()
const emit = defineEmits<{close: [], open: [key: string]}>()
const issue = useQuery(computed(() => owIssueQuery(props.issueKey)))
const activity = useQuery(computed(() => owActivityQuery(props.issueKey)))
const comments = useQuery(computed(() => owCommentsQuery(issue.data.value?.task_id ?? 0)))
const project = computed(() => props.projects.find(p => p.project?.id === issue.data.value?.project_id))
const users = useQuery(computed(() => owUsersQuery(issue.data.value?.project_id ?? 0, '')))
const sprints = useQuery(computed(() => owSprintsQuery(issue.data.value?.project_id ?? 0)))
const epics = useQuery(computed(() => owIssuesQuery({project_id: issue.data.value?.project_id, rules: [{field: 'type', operator: 'eq', value: 'Epic'}]})))
const save = useOwSaveIssue(), addComment = useOwAddComment(), archive = useOwArchiveIssue(), duplicate = useOwDuplicateIssue(), watchIssue = useOwWatch()
const gitLinks = useQuery(computed(() => owGitLinksQuery(props.issueKey)))
const move = useOwMove(), destination = ref(0)
async function moveProject() {if(!destination.value || move.isPending.value) return;const result=await move.mutateAsync({key:props.issueKey,project:destination.value});emit('open',result.key)}
const fields = ref<InstanceType<typeof IssueFields>>()
const editing = ref(false), comment = ref(''), tab = ref('Comments')
const canEdit = computed(() => (project.value?.permission ?? 0) >= 1)
async function submit() {if (fields.value) {await save.mutateAsync(fields.value.getDraft()); editing.value = false}}
async function postComment() {if (!comment.value.trim() || addComment.isPending.value) return; await addComment.mutateAsync({task: issue.data.value!.task_id, comment: comment.value}); comment.value = ''}
async function copyIssue() {const value = await duplicate.mutateAsync(props.issueKey); emit('open', value.key)}
async function archiveIssue() {await archive.mutateAsync(props.issueKey); emit('close')}
</script>

<template>
	<section
		class="ow-detail"
		aria-labelledby="issue-heading"
	>
		<header class="ow-detail-header">
			<button @click="emit('close')">
				← All issues
			</button><span class="ow-key">{{ issueKey }}</span><button
				aria-label="Close issue"
				@click="emit('close')"
			>
				×
			</button>
		</header>
		<p
			v-if="issue.isPending.value"
			role="status"
		>
			Loading issue…
		</p>
		<p
			v-else-if="issue.isError.value"
			class="ow-error"
			role="alert"
		>
			This issue could not be opened. Check your access and try again.
		</p>
		<template v-else-if="issue.data.value && project?.settings">
			<div class="ow-heading">
				<div>
					<p class="ow-eyebrow">
						{{ project.project?.title }} · {{ issue.data.value.type }}
					</p><h1 id="issue-heading">
						{{ issue.data.value.summary }}
					</h1>
				</div><span class="ow-badge">{{ project.settings.states?.find(s => s.id === issue.data.value?.state)?.name }}</span>
			</div>
			<div class="ow-actions">
				<button
					v-if="canEdit"
					class="ow-primary"
					@click="editing = !editing"
				>
					{{ editing ? 'Cancel editing' : 'Edit issue' }}
				</button><button
					:aria-pressed="!!issue.data.value.task?.subscription"
					@click="watchIssue.mutate({id: issue.data.value!.task_id, watch: !issue.data.value?.task?.subscription})"
				>
					{{ issue.data.value.task?.subscription ? 'Watching' : 'Watch issue' }}
				</button><button
					v-if="canEdit"
					@click="copyIssue"
				>
					Duplicate
				</button><details
					v-if="canEdit"
					class="ow-menu"
				>
					<summary class="ow-button">
						More
					</summary><div class="ow-card">
						<button @click="archiveIssue">
							Archive issue
						</button><label>Move to project<select v-model.number="destination"><option :value="0">Choose destination</option><option
							v-for="p in projects.filter(p => (p.permission ?? 0) >= 1 && p.project?.id !== issue.data.value?.project_id)"
							:key="p.project?.id"
							:value="p.project?.id"
						>{{ p.project?.title }}</option></select></label><button
							:aria-disabled="!destination || move.isPending.value"
							@click="moveProject"
						>
							Move issue
						</button>
					</div>
				</details>
			</div>
			<form
				v-if="editing"
				class="ow-card ow-editor"
				@submit.prevent="submit"
			>
				<IssueFields
					ref="fields"
					:initial="issue.data.value"
					:settings="project.settings"
					:sprints="sprints.data.value ?? []"
					:users="users.data.value ?? []"
					:epics="epics.data.value?.items ?? []"
				/><footer class="ow-actions">
					<button
						class="ow-primary"
						:disabled="save.isPending.value"
					>
						{{ save.isPending.value ? 'Saving…' : 'Save changes' }}
					</button><button
						type="button"
						@click="editing = false"
					>
						Cancel
					</button>
				</footer>
			</form>
			<div
				v-else
				class="ow-detail-columns"
			>
				<article class="ow-card">
					<h2>Description</h2><div
						v-if="issue.data.value.description"
						class="ow-rich-content"
						v-html="safeHtml(issue.data.value.description)"
					/><p
						v-else
						class="ow-caption"
					>
						No description yet.
					</p><h3 v-if="issue.data.value.task?.labels?.length">
						Labels
					</h3><span
						v-for="label in issue.data.value.task?.labels"
						:key="label.id"
						class="ow-badge"
					>{{ label.title }}</span>
				</article><aside class="ow-card">
					<h2>Details</h2><dl><dt>Assignees</dt><dd>{{ users.data.value?.filter(u => issue.data.value?.assignee_ids.includes(u.id!)).map(u => u.name || u.username).join(', ') || 'Unassigned' }}</dd><dt>Priority</dt><dd>{{ ['None','Low','Normal','High','Urgent','Immediate'][issue.data.value.priority ?? 0] }}</dd><dt>Due</dt><dd>{{ displayDate(issue.data.value.due_date) }}</dd><dt>Sprint</dt><dd>{{ sprints.data.value?.find(s => s.id === issue.data.value?.sprint_id)?.name || 'Backlog' }}</dd><dt>Story points</dt><dd>{{ issue.data.value.points || '—' }}</dd><dt>Estimate</dt><dd>{{ issue.data.value.estimate || '—' }} hours</dd><dt>Component</dt><dd>{{ issue.data.value.component || '—' }}</dd><dt>Target version</dt><dd>{{ issue.data.value.fix_version || '—' }}</dd><dt>Created</dt><dd>{{ displayDate(issue.data.value.created) }}</dd><dt>Updated</dt><dd>{{ displayDate(issue.data.value.updated) }}</dd></dl>
				</aside>
			</div>
			<IssueLinks
				:issue="issue.data.value"
				:can-edit="canEdit"
				@open="emit('open',$event)"
			/>
			<IssueAttachments
				:issue-key="issueKey"
				:can-edit="canEdit"
			/><section class="ow-card ow-git-links">
				<h2>Development</h2><a
					v-for="link in gitLinks.data.value"
					:key="link.id"
					:href="link.url"
					target="_blank"
					rel="noopener noreferrer"
				>{{ link.kind }} · {{ link.title }}</a><p
					v-if="!gitLinks.data.value?.length"
					class="ow-caption"
				>
					Mention {{ issueKey }} in a connected repository's commit or pull request to link development work.
				</p>
			</section><section class="ow-card">
				<div class="ow-tabs">
					<button
						v-for="item in ['Comments','History']"
						:key="item"
						:aria-current="tab === item ? 'page' : undefined"
						@click="tab = item"
					>
						{{ item }}
					</button>
				</div><template v-if="tab === 'Comments'">
					<article
						v-for="item in comments.data.value"
						:key="item.id"
						class="ow-comment"
					>
						<header><strong>{{ item.author?.name || item.author?.username }}</strong><time>{{ displayDate(item.created) }}</time></header><div
							class="ow-rich-content"
							v-html="safeHtml(item.comment ?? '')"
						/>
					</article><p
						v-if="!comments.data.value?.length"
						class="ow-caption"
					>
						Start the conversation. Choose a teammate from the comment toolbar to notify them.
					</p><form
						v-if="canEdit"
						@submit.prevent="postComment"
					>
						<label>Add a comment<RichText
							v-model="comment"
							label="Comment"
							:people="users.data.value ?? []"
						/></label><button
							class="ow-primary"
							:disabled="addComment.isPending.value || !comment.trim()"
						>
							Post comment
						</button>
					</form>
				</template><ol
					v-else
					class="ow-timeline"
				>
					<li
						v-for="entry in activity.data.value"
						:key="entry.id"
					>
						<strong>{{ entry.detail }}</strong><small>{{ displayDate(entry.at) }} · user {{ entry.actor_id }}</small>
					</li>
				</ol>
			</section>
		</template>
	</section>
</template>
