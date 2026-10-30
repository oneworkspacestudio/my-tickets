<script setup lang="ts">
import {ref} from 'vue'
import type {Bootstrap} from './main'
import {base, friendly} from './main'

const props = defineProps<{bootstrap: Bootstrap}>()
const email = ref('')
const password = ref('')
const pending = ref(false)
const message = ref('')
const setupToken = window.location.hash.startsWith('#setup/') ? window.location.hash.slice(7) : ''

async function submit() {
	if (pending.value || props.bootstrap.needsSetup && !setupToken) return
	pending.value = true
	message.value = ''
	try {
		const response = await fetch(`${base}/api/${props.bootstrap.needsSetup ? 'setup' : 'login'}`, {
			method: 'POST',
			credentials: 'same-origin',
			headers: {'Content-Type': 'application/json', 'X-Setup-Token': setupToken},
			body: JSON.stringify({email: email.value, password: password.value}),
		})
		if (!response.ok) throw await response.json()
		password.value = ''
		window.location.replace(`${base}/`)
	} catch (error) {
		message.value = friendly(error)
	} finally {
		pending.value = false
	}
}
</script>

<template>
	<main class="ow-signin">
		<form
			class="ow-card ow-signin-card"
			@submit.prevent="submit"
		>
			<img
				:src="`${base}/tickets.svg`"
				width="64"
				height="64"
				alt=""
			>
			<p class="ow-eyebrow">
				One Workspace
			</p>
			<h1>{{ bootstrap.needsSetup ? 'Set up My Tickets' : 'Welcome back' }}</h1>
			<p>Projects, issues and insights in one place.</p>
			<p
				v-if="message"
				class="ow-error"
				role="alert"
			>
				{{ message }}
			</p>
			<p
				v-if="bootstrap.needsSetup && !setupToken"
				class="ow-info"
			>
				Open the one-time setup link printed by My Tickets on this PC.
			</p>
			<label>Email address<input
				v-model="email"
				type="email"
				autocomplete="username"
				required
			></label>
			<label>Password<input
				v-model="password"
				type="password"
				:autocomplete="bootstrap.needsSetup ? 'new-password' : 'current-password'"
				minlength="8"
				required
			></label>
			<button
				class="ow-primary"
				:aria-disabled="pending || bootstrap.needsSetup && !setupToken"
				:aria-busy="pending"
			>
				{{ pending ? 'Connecting…' : bootstrap.needsSetup ? 'Create private workspace' : 'Sign in' }}
			</button>
			<p class="ow-caption">
				Managed installations connect automatically through MyAccounts.
			</p>
		</form>
	</main>
</template>
