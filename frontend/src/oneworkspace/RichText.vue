<script setup lang="ts">
import {watch} from 'vue'
import {EditorContent, useEditor} from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import {Node} from '@tiptap/core'
import type {User} from '@/client/generated'

const props = withDefaults(defineProps<{modelValue: string, label: string, people?: User[]}>(), {people: () => []})
const emit = defineEmits<{'update:modelValue': [value: string]}>()
const mentionUser = Node.create({
	name: 'mentionUser', group: 'inline', inline: true, atom: true,
	addAttributes() {return {id: {default: '',parseHTML: element => element.getAttribute('data-id')},label: {default:'',parseHTML: element => element.getAttribute('data-label')}}},
	parseHTML() {return [{tag:'mention-user'}]},
	renderHTML({HTMLAttributes}) {return ['mention-user',{'data-id':HTMLAttributes.id,'data-label':HTMLAttributes.label},`@${HTMLAttributes.label || HTMLAttributes.id}`]},
})
const editor = useEditor({
	extensions: [StarterKit, mentionUser],
	content: '',
	editorProps: {attributes: () => ({'aria-label': props.label, role: 'textbox', 'aria-multiline': 'true'})},
	onUpdate: ({editor}) => emit('update:modelValue', editor.getHTML()),
})
watch([() => props.modelValue, editor], ([value]) => {
	if (editor.value && editor.value.getHTML() !== value) editor.value.commands.setContent(value, {emitUpdate: false})
}, {immediate: true})
function mention(event: Event) {
	const select=event.target as HTMLSelectElement
	const person=props.people.find(p=>p.username===select.value)
	if(person) editor.value?.chain().focus().insertContent([{type:'mentionUser',attrs:{id:person.username,label:person.name || person.username}},{type:'text',text:' '}]).run()
	select.value=''
}
</script>

<template>
	<div class="ow-editor">
		<div
			class="ow-editor-tools"
			role="toolbar"
			:aria-label="`${label} formatting`"
		>
			<button
				type="button"
				aria-label="Bold"
				:aria-pressed="editor?.isActive('bold')"
				@click="editor?.chain().focus().toggleBold().run()"
			>
				<strong>B</strong>
			</button>
			<button
				type="button"
				aria-label="Italic"
				:aria-pressed="editor?.isActive('italic')"
				@click="editor?.chain().focus().toggleItalic().run()"
			>
				<em>I</em>
			</button>
			<button
				type="button"
				aria-label="Heading"
				@click="editor?.chain().focus().toggleHeading({level: 2}).run()"
			>
				H2
			</button>
			<button
				type="button"
				aria-label="Bullet list"
				@click="editor?.chain().focus().toggleBulletList().run()"
			>
				• List
			</button>
			<button
				type="button"
				aria-label="Code block"
				@click="editor?.chain().focus().toggleCodeBlock().run()"
			>
				&lt;/&gt;
			</button>
			<select
				v-if="people.length"
				aria-label="Mention a teammate"
				@change="mention"
			>
				<option value="">
					@ Mention
				</option><option
					v-for="person in people"
					:key="person.id"
					:value="person.username"
				>
					{{ person.name || person.username }}
				</option>
			</select>
		</div>
		<EditorContent :editor="editor" />
	</div>
</template>
