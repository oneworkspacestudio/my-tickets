<script setup lang="ts">
import {computed,ref} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {owAttachmentsQuery,useOwUpload,useOwAttachmentState,downloadOwAttachment} from '@/client/queries/owTickets'
const props=defineProps<{issueKey:string,canEdit:boolean}>()
const rows=useQuery(computed(()=>owAttachmentsQuery(props.issueKey)))
const upload=useOwUpload(),state=useOwAttachmentState()
const message=ref('')
async function selected(event:Event){const input=event.target as HTMLInputElement;const file=input.files?.[0];if(!file||upload.isPending.value)return;await upload.mutateAsync({key:props.issueKey,file});input.value='';message.value='Attachment saved and verified.'}
</script>

<template>
	<section class="ow-card ow-attachments">
		<h2>Attachments</h2><p class="ow-caption">
			Files stay in the protected My Tickets source in Drive. Project access applies to every download. Up to 64 MB per file.
		</p><label v-if="canEdit">Attach a file<input
			type="file"
			:aria-disabled="upload.isPending.value"
			@change="selected"
		></label><p
			v-if="upload.isPending.value"
			role="status"
		>
			Uploading and checking the file…
		</p><p role="status">
			{{ message }}
		</p><ul>
			<li
				v-for="file in rows.data.value"
				:key="file.id"
			>
				<div>
					<button
						v-if="!file.deleted"
						class="ow-text-button"
						@click="downloadOwAttachment(file.id!,file.name!)"
					>
						▧ {{ file.name }}
					</button><span v-else>{{ file.name }} · deleted</span><small>{{ ((file.size ?? 0)/1024).toFixed(1) }} KB · SHA256 {{ file.sha256?.slice(0,12) }}…</small>
				</div><button
					v-if="canEdit"
					:aria-disabled="state.isPending.value"
					@click="!state.isPending.value && state.mutate({id:file.id!,deleted:!file.deleted})"
				>
					{{ file.deleted ? 'Restore' : 'Delete' }}
				</button>
			</li>
		</ul><p
			v-if="!rows.data.value?.length"
			class="ow-caption"
		>
			No attachments yet.
		</p>
	</section>
</template>
