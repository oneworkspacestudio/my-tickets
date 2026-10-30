<script setup lang="ts">
import {computed, ref} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {owLabelsQuery, owIssuesQuery, useOwLabel, useOwRelation} from '@/client/queries/owTickets'
import type {Ticket} from '@/client/queries/owTickets'
import type {TaskRelationWritable} from '@/client/generated'

const props=defineProps<{issue:Ticket,canEdit:boolean}>()
const emit=defineEmits<{open:[key:string]}>()
const labels=useQuery(owLabelsQuery())
const search=ref('')
const candidates=useQuery(computed(()=>owIssuesQuery({q:search.value})))
const label=ref('')
const labelID=ref(0)
const other=ref(0)
const kind=ref<TaskRelationWritable['relation_kind']>('related')
const saveLabel=useOwLabel()
const saveRelation=useOwRelation()
async function addLabel(){
	if(saveLabel.isPending.value || (!labelID.value && !label.value.trim()))return
	await saveLabel.mutateAsync({task:props.issue.task_id,id:labelID.value || undefined,title:label.value.trim()})
	label.value='';labelID.value=0
}
async function addRelation(){
	if(saveRelation.isPending.value || !other.value)return
	await saveRelation.mutateAsync({task:props.issue.task_id,other:other.value,kind:kind.value})
	other.value=0
}
function keyOf(task:{identifier?:string,id?:number}){return task.identifier || `Issue ${task.id}`}
</script>

<template>
	<section class="ow-card">
		<h2>Labels and linked issues</h2>
		<div class="ow-actions">
			<span
				v-for="item in issue.task?.labels"
				:key="item.id"
				class="ow-badge"
			>
				{{ item.title }}
				<button
					v-if="canEdit"
					:aria-label="`Remove label ${item.title}`"
					@click="saveLabel.mutate({task:issue.task_id,id:item.id,remove:true})"
				>×</button>
			</span>
		</div>
		<form
			v-if="canEdit"
			class="ow-actions"
			@submit.prevent="addLabel"
		>
			<label>Existing label<select v-model.number="labelID"><option :value="0">Choose a label</option><option
				v-for="item in labels.data.value"
				:key="item.id"
				:value="item.id"
			>{{ item.title }}</option></select></label>
			<label>Or create label<input
				v-model="label"
				maxlength="250"
			></label>
			<button :aria-disabled="saveLabel.isPending.value || (!labelID && !label.trim())">
				Add label
			</button>
		</form>
		<ul>
			<template
				v-for="(tasks,relation) in issue.task?.related_tasks"
				:key="relation"
			>
				<li
					v-for="task in tasks"
					:key="task.id"
				>
					<span>{{ relation }} · </span><button
						class="ow-link"
						@click="emit('open',keyOf(task))"
					>
						{{ keyOf(task) }} {{ task.title }}
					</button>
					<button
						v-if="canEdit"
						:aria-label="`Remove linked issue ${keyOf(task)}`"
						@click="saveRelation.mutate({task:issue.task_id,other:task.id!,kind:relation as TaskRelationWritable['relation_kind'],remove:true})"
					>
						×
					</button>
				</li>
			</template>
		</ul>
		<form
			v-if="canEdit"
			class="ow-actions"
			@submit.prevent="addRelation"
		>
			<label>Find issue<input
				v-model="search"
				type="search"
				placeholder="Issue key or summary"
			></label>
			<label>Relationship<select v-model="kind"><option value="related">Relates to</option><option value="blocking">Blocks</option><option value="blocked">Is blocked by</option><option value="duplicateof">Duplicates</option><option value="subtask">Has subtask</option><option value="parenttask">Has parent</option></select></label>
			<label>Issue<select v-model.number="other"><option :value="0">Choose an issue</option><option
				v-for="item in candidates.data.value?.items.filter(i=>i.task_id!==issue.task_id)"
				:key="item.task_id"
				:value="item.task_id"
			>{{ item.key }} · {{ item.summary }}</option></select></label>
			<button :aria-disabled="saveRelation.isPending.value || !other">
				Link issue
			</button>
		</form>
	</section>
</template>
