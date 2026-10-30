<script setup lang="ts">
import {computed,ref} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {owMembersQuery,useOwMember} from '@/client/queries/owTickets'
const props=defineProps<{project:number}>()
const members=useQuery(computed(()=>owMembersQuery(props.project))),save=useOwMember()
const username=ref(''),permission=ref(1)
async function add(){if(save.isPending.value)return;const current=username.value;await save.mutateAsync({project:props.project,username:current,permission:permission.value,action:'add'});if(username.value===current)username.value=''}
</script>

<template>
	<section class="ow-card ow-editor">
		<h2>People & access</h2><p class="ow-caption">
			Use the person's My Tickets username. They appear after their first MyAccounts sign-in. Administrators can manage the project; contributors can change issues; viewers can read.
		</p><form
			class="ow-form-grid"
			@submit.prevent="add"
		>
			<label>Username<input
				v-model="username"
				required
			></label><label>Role<select v-model.number="permission"><option :value="0">Viewer</option><option :value="1">Contributor</option><option :value="2">Administrator</option></select></label><button
				class="ow-primary"
				:aria-disabled="save.isPending.value"
			>
				Add member
			</button>
		</form><div
			v-for="member in members.data.value"
			:key="member.id"
			class="ow-config-row"
		>
			<span>{{ member.name || member.username }} <small>{{ member.username }}</small></span><select
				:value="member.permission"
				:aria-label="`Role for ${member.username}`"
				@change="save.mutate({project,username:member.username!,permission:Number(($event.target as HTMLSelectElement).value),action:'update'})"
			>
				<option :value="0">
					Viewer
				</option><option :value="1">
					Contributor
				</option><option :value="2">
					Administrator
				</option>
			</select><button
				:aria-disabled="save.isPending.value"
				@click="!save.isPending.value && save.mutate({project,username:member.username!,permission:0,action:'remove'})"
			>
				Remove
			</button>
		</div>
	</section>
</template>
