<script setup lang="ts">
import {ref,watch} from 'vue'
import type {OwDashboard,OwProjectInfo,OwSprint} from '@/client/generated'
import {useOwSaveDashboard} from '@/client/queries/owTickets'
const props=defineProps<{dashboard:OwDashboard,projects:OwProjectInfo[],sprints:OwSprint[],owner:boolean}>()
const draft=ref<OwDashboard>({filter:{}}),save=useOwSaveDashboard()
watch(()=>props.dashboard,value=>{draft.value=JSON.parse(JSON.stringify(value));draft.value.filter??={}},{immediate:true})
const shares=ref('')
async function submit(){if(save.isPending.value)return;const ids=shares.value.split(',').map(s=>Number(s.trim())).filter(n=>n>0);await save.mutateAsync({...draft.value,shares:shares.value.trim()?ids:draft.value.shares})}
</script>

<template>
	<details class="ow-menu">
		<summary class="ow-button">
			Sharing & filters
		</summary><form
			class="ow-card"
			@submit.prevent="submit"
		>
			<label>Name<input
				v-model="draft.name"
				required
			></label><label>Visibility<select v-model="draft.scope"><option>private</option><option>project</option><option v-if="owner">global</option></select></label><label>Project<select v-model.number="draft.project_id"><option :value="0">All accessible</option><option
				v-for="p in projects"
				:key="p.project?.id"
				:value="p.project?.id"
			>{{ p.project?.title }}</option></select></label><label>Sprint<select v-model.number="draft.filter!.sprint_id"><option :value="0">Any sprint</option><option
				v-for="s in sprints"
				:key="s.id"
				:value="s.id"
			>{{ s.name }}</option></select></label><label>Additional user IDs<input
				v-model="shares"
				:placeholder="draft.shares?.join(', ') || 'Optional, separated by commas'"
			></label><p class="ow-caption">
				Every viewer only sees their permitted projects and issues.
			</p><button :aria-disabled="save.isPending.value">
				Save dashboard
			</button>
		</form>
	</details>
</template>
