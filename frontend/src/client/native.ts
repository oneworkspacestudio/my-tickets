import {client} from '@/client/generated/client.gen'
import {normalizeProblemBody} from '@/client/problem'

export type NativeUser = {
	id: string
	issuer: string
	subject: string
	name: string
	email: string
	role: string
}

export type NativeSession = {
	user: NativeUser
	csrf: string
	expires: string
}

let account = ''
let epoch = 0
const requests = new WeakMap<Request, number>()
window.addEventListener('workspace-account-changed', () => { epoch++ })

// The native SDK owns the HttpOnly cookie. No browser bearer token is created or persisted.
export function configureNativeClient(base: string, session: NativeSession) {
	account = session.user.id
	epoch++
	const root = new URL(`${base}/api/v2/`, window.location.origin)
	client.setConfig({
		baseUrl: root.toString().replace(/\/$/, ''),
		credentials: 'same-origin',
		throwOnError: true,
	})
	client.interceptors.request.clear()
	client.interceptors.response.clear()
	client.interceptors.error.clear()
	client.interceptors.request.use(request => {
		const url = new URL(request.url)
		if (url.origin !== root.origin || !url.pathname.startsWith(root.pathname)) {
			throw new DOMException('The application request address changed', 'AbortError')
		}
		const headers = new Headers(request.headers)
		headers.delete('Authorization')
		headers.set('X-Workspace-Expected-Account', account)
		if (!['GET', 'HEAD'].includes(request.method)) {
			headers.set('X-CSRF-Token', session.csrf)
		}
		const guarded = new Request(request, {headers})
		requests.set(guarded, epoch)
		return guarded
	})
	client.interceptors.response.use((response, request) => {
		if (requests.get(request) !== epoch) throw new DOMException('Your session changed', 'AbortError')
		if (response.status === 401) window.dispatchEvent(new Event('workspace-account-changed'))
		const current = response.headers.get('X-Workspace-Account')
		if (current && current !== account) {
			window.dispatchEvent(new Event('workspace-account-changed'))
			throw new DOMException('The signed-in account changed', 'AbortError')
		}
		return response
	})
	client.interceptors.error.use(normalizeProblemBody)
}
