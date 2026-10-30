import DOMPurify from 'dompurify'

export function safeHtml(value: string | undefined) {
	return DOMPurify.sanitize(value ?? '', {
		ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'u', 's', 'h1', 'h2', 'h3', 'ul', 'ol', 'li', 'blockquote', 'pre', 'code', 'a', 'hr', 'span'],
		ALLOWED_ATTR: ['href', 'title'],
	})
}

export function dateInput(value: string | undefined) {
	if (!value || value.startsWith('0001-')) return ''
	const date = new Date(value)
	if (Number.isNaN(date.getTime())) return ''
	return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

export function dateValue(value: string) {
	return value ? new Date(`${value}T23:59:59`).toISOString() : '0001-01-01T00:00:00Z'
}

export function displayDate(value: string | undefined) {
	if (!value || value.startsWith('0001-')) return '—'
	return new Date(value).toLocaleDateString(undefined, {day: 'numeric', month: 'short', year: 'numeric'})
}
