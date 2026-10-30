<script setup lang="ts">
import {ref, watch} from 'vue'
import type {OwProjectSettings, OwState, OwField} from '@/client/generated'
import {useOwSaveProject} from '@/client/queries/owTickets'
const props = defineProps<{settings: OwProjectSettings}>()
const draft = ref<OwProjectSettings>({})
watch(() => props.settings, value => { draft.value = JSON.parse(JSON.stringify(value)) }, {immediate: true})
const save = useOwSaveProject()
const field = ref({id: '', name: '', type: 'text' as OwField['type'], options: '', required: false})
const state = ref({id: '', name: '', category: 'open' as OwState['category'], limit: 0})
const transition = ref({from: '', to: '', admin_only: false, required: ''})
function addField() {if (!field.value.id || !field.value.name) return; draft.value.fields = [...draft.value.fields ?? [], {...field.value, options: field.value.options.split(',').map(s => s.trim()).filter(Boolean)}]; field.value = {id: '',name:'',type:'text',options:'',required:false}}
function addState() {if (!state.value.id || !state.value.name) return; draft.value.states = [...draft.value.states ?? [], {...state.value}]; state.value = {id:'',name:'',category:'open',limit:0}}
function addTransition() {if (!transition.value.from || !transition.value.to) return; draft.value.transitions = [...draft.value.transitions ?? [], {...transition.value, required: transition.value.required.split(',').map(s => s.trim()).filter(Boolean)}]}
</script>

<template>
	<form
		class="ow-card ow-editor"
		@submit.prevent="save.mutate(draft)"
	>
		<h2>Project settings · {{ settings.key }}</h2><p class="ow-caption">
			Project keys stay permanent. Used states and fields cannot be removed until affected issues are updated.
		</p><h3>Workflow states</h3><div
			v-for="(item,i) in draft.states"
			:key="item.id"
			class="ow-config-row"
		>
			<code>{{ item.id }}</code><label class="ow-check"><input
				v-model="item.working"
				type="checkbox"
			>Active work</label><label>Name<input v-model="item.name"></label><label>Category<select v-model="item.category"><option>open</option><option>done</option><option>cancelled</option></select></label><label>Work limit<input
				v-model.number="item.limit"
				type="number"
				min="0"
			></label><button
				type="button"
				aria-label="Remove state"
				@click="draft.states?.splice(i,1)"
			>
				×
			</button>
		</div><div class="ow-form-grid">
			<label>New state ID<input
				v-model="state.id"
				placeholder="awaiting-review"
			></label><label>State name<input v-model="state.name"></label>
		</div><button
			type="button"
			@click="addState"
		>
			Add state
		</button><h3>Allowed transitions</h3><div
			v-for="(item,i) in draft.transitions"
			:key="i"
			class="ow-transition"
		>
			<span>{{ item.from }} → {{ item.to }}</span><span>{{ item.admin_only ? 'Administrators only' : 'Contributors' }}</span><small v-if="item.required?.length">Requires {{ item.required.join(', ') }}</small><button
				type="button"
				aria-label="Remove transition"
				@click="draft.transitions?.splice(i,1)"
			>
				×
			</button>
		</div><div class="ow-form-grid">
			<label>From<select v-model="transition.from"><option value="">Choose state</option><option value="*">Any state</option><option
				v-for="s in draft.states"
				:key="s.id"
				:value="s.id"
			>{{ s.name }}</option></select></label><label>To<select v-model="transition.to"><option value="">Choose state</option><option
				v-for="s in draft.states"
				:key="s.id"
				:value="s.id"
			>{{ s.name }}</option></select></label><label>Required fields<input
				v-model="transition.required"
				placeholder="assignee, description, custom-field-id"
			></label><label class="ow-check"><input
				v-model="transition.admin_only"
				type="checkbox"
			>Administrators only</label>
		</div><button
			type="button"
			@click="addTransition"
		>
			Add transition
		</button><h3>Custom fields</h3><div
			v-for="(item,i) in draft.fields"
			:key="item.id"
			class="ow-transition"
		>
			<strong>{{ item.name }}</strong><code>{{ item.id }}</code><span>{{ item.type }}{{ item.required ? ' · required' : '' }}</span><button
				type="button"
				aria-label="Remove field"
				@click="draft.fields?.splice(i,1)"
			>
				×
			</button>
		</div><div class="ow-form-grid">
			<label>Field ID<input
				v-model="field.id"
				placeholder="customer-impact"
			></label><label>Label<input v-model="field.name"></label><label>Type<select v-model="field.type"><option
				v-for="type in ['text','number','date','boolean','select','multi-select']"
				:key="type"
			>{{ type }}</option></select></label><label v-if="field.type?.includes('select')">Options, separated by commas<input v-model="field.options"></label><label class="ow-check"><input
				v-model="field.required"
				type="checkbox"
			>Required</label>
		</div><button
			type="button"
			@click="addField"
		>
			Add field
		</button><footer class="ow-actions">
			<button
				class="ow-primary"
				:disabled="save.isPending.value"
			>
				{{ save.isPending.value ? 'Saving…' : 'Save project settings' }}
			</button><span
				v-if="save.isSuccess.value"
				role="status"
			>Settings saved</span>
		</footer>
	</form>
</template>
