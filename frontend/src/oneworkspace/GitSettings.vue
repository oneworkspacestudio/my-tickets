<script setup lang="ts">
import {computed,ref,watch} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import type {OwGitSettings,OwGitSettingsWritable,OwProjectSettings} from '@/client/generated'
import {owGitSettingsQuery,useOwGitSettings,useOwGitResync} from '@/client/queries/owTickets'
const props=defineProps<{project:number,settings:OwProjectSettings}>()
const config=useQuery(computed(()=>owGitSettingsQuery(props.project)))
const save=useOwGitSettings(),resync=useOwGitResync()
const draft=ref<OwGitSettings & OwGitSettingsWritable>({project_id:0,url:'',repository:'',token:'',secret:'',rules:[]})
watch(()=>config.data.value,value=>{if(value)draft.value=JSON.parse(JSON.stringify(value))},{immediate:true})
async function submit(){if(save.isPending.value)return;await save.mutateAsync({...draft.value,project_id:props.project});draft.value.token='';draft.value.secret=''}
</script>

<template>
	<form
		class="ow-card ow-editor"
		@submit.prevent="submit"
	>
		<h2>My Git / Gitea</h2><p>Connect a repository to show commits and pull requests that mention issue keys.</p><div class="ow-form-grid">
			<label>Git server URL<input
				v-model="draft.url"
				type="url"
				required
				placeholder="https://git.example.com"
			></label><label>Repository<input
				v-model="draft.repository"
				required
				placeholder="team/project"
			></label><label>Service token<input
				v-model="draft.token"
				type="password"
				autocomplete="new-password"
				:placeholder="draft.configured ? 'Saved · leave blank to keep' : 'Repository read token'"
			></label><label>Webhook secret<input
				v-model="draft.secret"
				type="password"
				autocomplete="new-password"
				minlength="32"
				:placeholder="draft.configured ? 'Saved · leave blank to keep' : 'At least 32 characters'"
			></label>
		</div><details>
			<summary>Optional workflow automation</summary><p class="ow-caption">
				Disabled until you add a rule. Ticket permissions and required transition fields still apply.
			</p><div
				v-for="(rule,i) in draft.rules"
				:key="i"
				class="ow-config-row"
			>
				<label>Event<select v-model="rule.event"><option>push</option><option>pull_request.closed</option></select></label><label>Destination<select v-model="rule.state"><option
					v-for="state in settings.states"
					:key="state.id"
					:value="state.id"
				>{{ state.name }}</option></select></label><button
					type="button"
					aria-label="Remove automation rule"
					@click="draft.rules?.splice(i,1)"
				>
					×
				</button>
			</div><button
				type="button"
				@click="draft.rules=[...draft.rules ?? [],{event:'push',state:settings.states?.[0]?.id ?? 'backlog'}]"
			>
				Add explicit rule
			</button>
		</details><div class="ow-actions">
			<button
				class="ow-primary"
				:aria-disabled="save.isPending.value"
			>
				Save connection
			</button><button
				v-if="draft.configured"
				type="button"
				:aria-disabled="resync.isPending.value"
				@click="!resync.isPending.value && resync.mutate(project)"
			>
				{{ resync.isPending.value ? 'Checking Git…' : 'Resync recent commits' }}
			</button><span
				v-if="save.isSuccess.value"
				role="status"
			>Connection saved securely</span>
		</div><p class="ow-caption">
			Configure Gitea's JSON webhook to /tickets/api/v2/tickets/projects/{{ project }}/git/webhook. Set the secret above and a My Tickets agent token in its Authorization header. See <a
				href="https://github.com/oneworkspacestudio/my-tickets/blob/main/docs/mygit-gitea.md"
				target="_blank"
				rel="noopener"
			>setup instructions</a>.
		</p>
	</form>
</template>
