<script setup lang="ts">
import type {OwFilter, OwRule, OwProjectSettings} from '@/client/generated'

const props = defineProps<{modelValue: OwFilter, settings?: OwProjectSettings}>()
const emit = defineEmits<{'update:modelValue': [filter: OwFilter]}>()

function update(rule: number, field: keyof OwRule, value: string) {
	const rules = (props.modelValue.rules ?? []).map(item => ({...item}))
	rules[rule] = {...rules[rule], [field]: value}
	emit('update:modelValue', {...props.modelValue, rules})
}
function add() {
	emit('update:modelValue', {...props.modelValue, rules: [...(props.modelValue.rules ?? []), {field: 'state', operator: 'eq', value: 'backlog'}]})
}
function remove(index: number) {
	emit('update:modelValue', {...props.modelValue, rules: props.modelValue.rules?.filter((_, i) => i !== index)})
}
</script>

<template>
	<details class="ow-filter-dropdown">
		<summary class="ow-button">
			Filters <span
				v-if="modelValue.rules?.length"
				class="ow-count"
			>{{ modelValue.rules.length }}</span>
		</summary>
		<div class="ow-filter-panel ow-card">
			<h3>Match issues</h3>
			<label>Combine conditions<select
				:value="modelValue.match || 'and'"
				@change="emit('update:modelValue', {...modelValue, match: ($event.target as HTMLSelectElement).value as 'and' | 'or'})"
			><option value="and">All conditions (AND)</option><option value="or">Any condition (OR)</option></select></label>
			<div
				v-for="(rule, index) in modelValue.rules"
				:key="index"
				class="ow-rule"
			>
				<select
					:value="rule.field"
					aria-label="Filter field"
					@change="update(index, 'field', ($event.target as HTMLSelectElement).value)"
				>
					<option
						v-for="field in ['state', 'type', 'priority', 'assignee', 'reporter', 'label', 'severity', 'component', 'fix_version', 'due_date', 'created', 'updated']"
						:key="field"
						:value="field"
					>
						{{ field.replaceAll('_', ' ') }}
					</option><option
						v-for="field in settings?.fields"
						:key="field.id"
						:value="`custom.${field.id}`"
					>
						{{ field.name }}
					</option>
				</select>
				<select
					:value="rule.operator"
					aria-label="Filter comparison"
					@change="update(index, 'operator', ($event.target as HTMLSelectElement).value)"
				>
					<option value="eq">
						is
					</option><option value="ne">
						is not
					</option><option value="contains">
						contains
					</option><option value="before">
						before / less than
					</option><option value="after">
						after / greater than
					</option><option value="empty">
						is empty
					</option>
				</select>
				<input
					v-if="rule.operator !== 'empty'"
					:value="rule.value"
					aria-label="Filter value"
					:placeholder="rule.field?.includes('date') ? 'YYYY-MM-DD' : 'Value or user ID'"
					@change="update(index, 'value', ($event.target as HTMLInputElement).value)"
				>
				<button
					type="button"
					:aria-label="`Remove condition ${index + 1}`"
					@click="remove(index)"
				>
					×
				</button>
			</div>
			<div class="ow-actions">
				<button
					type="button"
					@click="add"
				>
					+ Add condition
				</button><button
					type="button"
					@click="emit('update:modelValue', {...modelValue, rules: []})"
				>
					Clear
				</button>
			</div>
		</div>
	</details>
</template>
