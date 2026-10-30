<script setup lang="ts">
import {ref, watch} from 'vue'
import type {OwProjectSettings, OwSprint, User} from '@/client/generated'
import type {Ticket} from '@/client/queries/owTickets'
import RichText from './RichText.vue'
import {dateInput, dateValue} from './safeHtml'

const props = defineProps<{
	initial: Ticket
	settings: OwProjectSettings
	sprints: OwSprint[]
	users: User[]
	epics: Ticket[]
}>()
const draft = ref<Ticket>(null!)
watch(() => [props.initial.task_id,props.initial.project_id], () => { draft.value = JSON.parse(JSON.stringify(props.initial)) }, {immediate: true})
defineExpose({getDraft: () => JSON.parse(JSON.stringify(draft.value))})
</script>

<template>
	<label class="ow-wide">Summary<input
		v-model="draft.summary"
		maxlength="1000"
		required
		placeholder="What needs doing?"
	></label>
	<div class="ow-form-grid">
		<label>Issue type<select
			v-model="draft.type"
			aria-label="Issue type"
		><option
			v-for="kind in settings.types"
			:key="kind"
		>{{ kind }}</option></select></label>
		<label>Status<select
			v-model="draft.state"
			aria-label="Status"
		><option
			v-for="state in settings.states"
			:key="state.id"
			:value="state.id"
		>{{ state.name }}</option></select></label>
		<label>Priority<select
			v-model.number="draft.priority"
			aria-label="Priority"
		><option
			v-for="(label, value) in ['None', 'Low', 'Normal', 'High', 'Urgent', 'Immediate']"
			:key="value"
			:value="value"
		>{{ label }}</option></select></label>
		<label>Due date<input
			type="date"
			:value="dateInput(draft.due_date)"
			@input="draft.due_date = dateValue(($event.target as HTMLInputElement).value)"
		></label>
		<label>Assignees<select
			v-model="draft.assignee_ids"
			multiple
			aria-label="Assign one or more project members"
		><option
			v-for="person in users"
			:key="person.id"
			:value="person.id"
		>{{ person.name || person.username }}</option></select></label>
		<label>Sprint<select
			v-model.number="draft.sprint_id"
			aria-label="Sprint"
		><option :value="0">Backlog</option><option
			v-for="sprint in sprints.filter(s => s.state !== 'closed')"
			:key="sprint.id"
			:value="sprint.id"
		>{{ sprint.name }}</option></select></label>
		<label>Epic<select
			v-model.number="draft.epic_id"
			aria-label="Epic"
		><option :value="0">No epic</option><option
			v-for="epic in epics.filter(e => e.task_id !== draft.task_id)"
			:key="epic.key"
			:value="epic.task_id"
		>{{ epic.key }} · {{ epic.summary }}</option></select></label>
		<label>Severity<select
			v-model="draft.severity"
			aria-label="Severity"
		><option value="">Not specified</option><option>Low</option><option>Moderate</option><option>Major</option><option>Critical</option></select></label>
		<label>Story points<input
			v-model.number="draft.points"
			type="number"
			min="0"
			step="0.5"
		></label>
		<label>Estimate (hours)<input
			v-model.number="draft.estimate"
			type="number"
			min="0"
			step="0.25"
		></label>
		<label>Component<input
			v-model="draft.component"
			maxlength="250"
		></label>
		<label>Environment<input v-model="draft.environment"></label>
		<label>Affected version<input
			v-model="draft.affected_version"
			maxlength="120"
		></label>
		<label>Fix version / milestone<input
			v-model="draft.fix_version"
			maxlength="120"
		></label>
	</div>
	<label class="ow-wide">Description<RichText
		v-model="draft.description"
		label="Issue description"
		:people="users"
	/></label>
	<div
		v-if="settings.fields?.length"
		class="ow-form-grid"
	>
		<label
			v-for="field in settings.fields"
			:key="field.id"
		>{{ field.name }}<span v-if="field.required"> *</span>
			<select
				v-if="field.type === 'boolean'"
				v-model="draft.custom[field.id!]"
				:required="field.required"
			><option :value="undefined">Choose answer</option><option :value="true">Yes</option><option :value="false">No</option></select>
			<select
				v-else-if="field.type === 'select' || field.type === 'multi-select'"
				v-model="draft.custom[field.id!]"
				:multiple="field.type === 'multi-select'"
				:required="field.required"
			><option
				v-for="option in field.options"
				:key="option"
			>{{ option }}</option></select>
			<input
				v-else
				v-model="draft.custom[field.id!]"
				:type="field.type === 'number' ? 'number' : field.type === 'date' ? 'date' : 'text'"
				:required="field.required"
				@change="field.type === 'number' && (draft.custom[field.id!] = Number(($event.target as HTMLInputElement).value))"
			>
		</label>
	</div>
</template>
