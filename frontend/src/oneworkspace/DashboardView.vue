<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import type {OwDashboardData, OwWidget} from '@/client/generated'
import {useOwSaveDashboard, useOwSaveLayout} from '@/client/queries/owTickets'
import WidgetChart from './WidgetChart.vue'

type Widget = OwWidget
const props = defineProps<{data: OwDashboardData, canEdit: boolean, kinds: string[]}>()
const layout = ref<Widget[]>([])
const grid = ref<HTMLElement>()
const dragging = ref('')
const editing = ref(false)
const kind = ref('by-status')
const title = ref('')
const layoutMutation = useOwSaveLayout()
const saveMutation = useOwSaveDashboard()
let saving = false
const widgets = computed(() => [...layout.value].sort((a, b) => (a.y ?? 0) - (b.y ?? 0) || (a.x ?? 0) - (b.x ?? 0)))
watch(() => props.data.dashboard?.widgets, value => {if (!saving) layout.value = (value ?? []).map(w => ({...w}))}, {immediate: true})

function overlaps(a: Widget, b: Widget) {
	return (a.x ?? 0) < (b.x ?? 0) + (b.w ?? 1) && (a.x ?? 0) + (a.w ?? 1) > (b.x ?? 0) && (a.y ?? 0) < (b.y ?? 0) + (b.h ?? 1) && (a.y ?? 0) + (a.h ?? 1) > (b.y ?? 0)
}

function pack(moved: Widget) {
	const placed = [moved]
	for (const original of widgets.value.filter(w => w.id !== moved.id)) {
		const item = {...original}
		while (placed.some(p => overlaps(p, item))) item.y = (item.y ?? 0) + 1
		placed.push(item)
	}
	layout.value = placed
}

let queued = false
async function persist() {
	queued = true
	if (!props.data.dashboard?.id || saving) return
	saving = true
	try {
		while (queued) {
			queued = false
			await layoutMutation.mutateAsync({id: props.data.dashboard.id, widgets: JSON.parse(JSON.stringify(layout.value))})
		}
	} finally {saving = false}
}
const printDashboard = () => window.print()

function move(widget: Widget, dx: number, dy: number) {
	const next = {...widget, x: Math.max(0, Math.min(12 - (widget.w ?? 6), (widget.x ?? 0) + dx)), y: Math.max(0, (widget.y ?? 0) + dy)}
	pack(next)
	void persist()
}

function drop(event: DragEvent) {
	if (!grid.value || !dragging.value) return
	const widget = layout.value.find(w => w.id === dragging.value)
	if (!widget) return
	const rect = grid.value.getBoundingClientRect()
	pack({...widget, x: Math.max(0, Math.min(12 - (widget.w ?? 6), Math.floor((event.clientX - rect.left) / (rect.width / 12)))), y: Math.max(0, Math.floor((event.clientY - rect.top) / 80))})
	dragging.value = ''
	void persist()
}

function resized(event: PointerEvent, widget: Widget) {
	if (!editing.value || !grid.value) return
	const card = (event.currentTarget as HTMLElement)
	const unit = (grid.value.clientWidth + 16) / 12
	const w = Math.max(2, Math.min(12 - (widget.x ?? 0), Math.round((card.offsetWidth + 16) / unit)))
	const h = Math.max(2, Math.min(12, Math.round((card.offsetHeight + 16) / 80)))
	if (w === widget.w && h === widget.h) return
	card.style.width = ''
	card.style.height = ''
	pack({...widget, w, h})
	void persist()
}

async function sharedSave(next: Widget[]) {
	if (!props.data.dashboard?.id || saveMutation.isPending.value) return
	try {await saveMutation.mutateAsync({...props.data.dashboard, widgets: next})} catch {return}
}

function add() {
	const y = Math.max(0, ...layout.value.map(w => (w.y ?? 0) + (w.h ?? 3)))
	void sharedSave([...layout.value, {id: crypto.randomUUID(), kind: kind.value, title: title.value || kind.value.replaceAll('-', ' '), x: 0, y, w: 6, h: 3, filter: {}}])
}

function exportData() {
	const rows = [['Widget', 'Group', 'Value', 'Second series']]
	for (const widget of props.data.widgets ?? []) {
		rows.push([widget.title ?? widget.kind ?? '', 'Total', String(widget.value ?? 0), String(widget.total ?? '')])
		for (const point of widget.points ?? []) rows.push([widget.title ?? widget.kind ?? '', point.label ?? '', String(point.value ?? 0), String(point.secondary ?? 0)])
	}
	const csv = rows.map(row => row.map(value => `"${(/^[=+@\-\t\r]/.test(value) ? `'${value}` : value).replaceAll('"', '""')}"`).join(',')).join('\r\n')
	const href = URL.createObjectURL(new Blob([csv], {type: 'text/csv;charset=utf-8'}))
	const link = document.createElement('a')
	link.href = href
	link.download = 'my-tickets-dashboard.csv'
	link.click()
	URL.revokeObjectURL(href)
}
</script>

<template>
	<div class="ow-dashboard-tools ow-actions">
		<button
			:aria-pressed="editing"
			@click="editing = !editing"
		>
			{{ editing ? 'Finish arranging' : 'Arrange widgets' }}
		</button>
		<button @click="exportData">
			Export CSV
		</button><button @click="printDashboard">
			Print
		</button>
		<span
			v-if="layoutMutation.isPending.value || saveMutation.isPending.value"
			role="status"
		>Saving dashboard…</span>
		<span
			v-else
			class="ow-caption"
		>Calculated in {{ data.timezone }} · refreshed every 15 seconds</span>
	</div>
	<form
		v-if="canEdit && editing"
		class="ow-card ow-add-widget"
		@submit.prevent="add"
	>
		<label>Widget<select v-model="kind"><option
			v-for="item in kinds"
			:key="item"
			:value="item"
		>{{ item.replaceAll('-', ' ') }}</option></select></label><label>Title<input
			v-model="title"
			placeholder="Use default title"
		></label><button
			class="ow-primary"
			:aria-disabled="saveMutation.isPending.value"
		>
			+ Add widget
		</button>
	</form>
	<div
		ref="grid"
		class="ow-dashboard-grid"
		@dragover.prevent
		@drop.prevent="drop"
	>
		<article
			v-for="widget in widgets"
			:key="widget.id"
			class="ow-card ow-widget"
			:class="{'ow-resizable': editing}"
			:style="{gridColumn: `${(widget.x ?? 0) + 1} / span ${widget.w ?? 6}`, gridRow: `${(widget.y ?? 0) + 1} / span ${widget.h ?? 3}`}"
			@pointerup="resized($event, widget)"
		>
			<header
				:draggable="editing"
				@dragstart="dragging = widget.id ?? ''"
			>
				<h3>{{ widget.title || widget.kind }}</h3><span
					v-if="editing"
					aria-hidden="true"
				>⠿</span>
			</header>
			<WidgetChart
				v-if="data.widgets?.find(w => w.id === widget.id)"
				:data="data.widgets.find(w => w.id === widget.id)!"
				:chart="widget.chart"
			/>
			<div
				v-if="editing"
				class="ow-widget-controls"
			>
				<button
					:aria-label="`Move ${widget.title} left`"
					@click="move(widget, -1, 0)"
				>
					←
				</button><button
					:aria-label="`Move ${widget.title} up`"
					@click="move(widget, 0, -1)"
				>
					↑
				</button><button
					:aria-label="`Move ${widget.title} down`"
					@click="move(widget, 0, 1)"
				>
					↓
				</button><button
					:aria-label="`Move ${widget.title} right`"
					@click="move(widget, 1, 0)"
				>
					→
				</button><button
					:aria-label="`Widen ${widget.title}`"
					@click="pack({...widget, w: Math.min(12 - (widget.x ?? 0), (widget.w ?? 6) + 1)}); persist()"
				>
					＋
				</button><button
					:aria-label="`Narrow ${widget.title}`"
					@click="pack({...widget, w: Math.max(2, (widget.w ?? 6) - 1)}); persist()"
				>
					−
				</button><button
					v-if="canEdit"
					:aria-label="`Remove ${widget.title}`"
					@click="sharedSave(layout.filter(w => w.id !== widget.id))"
				>
					×
				</button><select
					v-if="canEdit"
					:value="widget.chart || ''"
					aria-label="Chart presentation"
					@change="sharedSave(layout.map(w => w.id === widget.id ? {...w, chart: ($event.target as HTMLSelectElement).value as OwWidget['chart']} : w))"
				>
					<option value="">
						Default chart
					</option><option>metric</option><option>line</option><option>bar</option><option>donut</option><option>table</option>
				</select>
			</div>
		</article>
	</div>
</template>
