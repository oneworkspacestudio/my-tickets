import {createApp, ref} from 'vue'
import {createRouter, createWebHistory} from 'vue-router'
import {MutationCache, QueryCache, QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {configureNativeClient, type NativeSession} from '@/client/native'
import TicketsApp from './TicketsApp.vue'
import SignIn from './SignIn.vue'
import './tickets.css'

export const notice = ref('')
export const base = document.querySelector<HTMLMetaElement>('meta[name="workspace-base"]')?.content ?? ''

export type Bootstrap = {
	name: string
	id: string
	version: string
	oidc: boolean
	needsSetup: boolean
	workspace: string
	base: string
}

export function friendly(error: unknown) {
	if (error instanceof Error) return error.message
	if (error && typeof error === 'object') {
		const body = error as {detail?: string, error?: string, message?: string}
		return body.detail ?? body.error ?? body.message ?? 'Could not finish. Please try again.'
	}
	return 'Could not finish. Please try again.'
}

async function start() {
	// These are the SDK's documented bootstrap/session calls, not resource-data requests.
	const initial = await fetch(`${base}/api/bootstrap`, {credentials: 'same-origin'})
	if (!initial.ok) throw new Error('My Tickets is not ready. Retry in a moment.')
	const bootstrap = await initial.json() as Bootstrap
	const response = await fetch(`${base}/api/me`, {credentials: 'same-origin'})
	if (response.status === 401) {
		if (bootstrap.oidc) {
			window.location.replace(`${base}/auth/login?return=${encodeURIComponent(window.location.pathname + window.location.search)}`)
			return
		}
		createApp(SignIn, {bootstrap}).mount('#app')
		return
	}
	if (!response.ok) throw new Error('Your session could not be checked. Please reload.')
	const session = await response.json() as NativeSession
	configureNativeClient(base, session)
	const report = (error: unknown) => {
		if ((error as {name?: string})?.name !== 'AbortError') notice.value = friendly(error)
	}
	const client = new QueryClient({
		queryCache: new QueryCache({onError: report}),
		mutationCache: new MutationCache({onError: report}),
		defaultOptions: {queries: {retry: 1, staleTime: 10_000}, mutations: {retry: false}},
	})
	const router = createRouter({
		history: createWebHistory(`${base}/`),
		routes: [{path: '/:pathMatch(.*)*', component: {template: '<span />'}}],
	})
	const app = createApp(TicketsApp, {bootstrap, session})
	app.config.errorHandler = report
	app.use(router)
	app.use(VueQueryPlugin, {queryClient: client})
	await router.isReady()
	app.mount('#app')
	window.addEventListener('workspace-account-changed', () => {
		void client.cancelQueries()
		client.clear()
		app.unmount()
		window.location.replace(bootstrap.oidc ? `${base}/auth/login?return=${encodeURIComponent(window.location.pathname + window.location.search)}` : `${base}/`)
	})
}

void start().catch(error => {
	document.querySelector('#app')!.textContent = friendly(error)
})
