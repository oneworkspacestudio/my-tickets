<script setup lang="ts">
import {computed} from 'vue'
import type {OwWidgetData} from '@/client/generated'
import {displayDate} from './safeHtml'

const props = defineProps<{data: OwWidgetData, chart?: string}>()
const points = computed(() => props.data.points ?? [])
const metric = computed(() => ['open', 'completed', 'overdue', 'unassigned', 'cycle-time', 'sprint-progress'].includes(props.data.kind ?? ''))
const style = computed(() => props.chart || (metric.value ? 'metric' : ['created-resolved', 'throughput', 'burndown', 'velocity', 'cumulative-flow'].includes(props.data.kind ?? '') ? 'line' : props.data.kind === 'by-status' ? 'donut' : 'bar'))
const max = computed(() => Math.max(1, ...points.value.flatMap(p => [p.value ?? 0, p.secondary ?? 0])))
const total = computed(() => points.value.reduce((n, p) => n + (p.value ?? 0), 0))
const colors = ['#0b57d0', '#6750c8', '#008675', '#b06000', '#c5221f', '#64748b']

function line(secondary = false) {
	return points.value.map((p, i) => `${12 + i / Math.max(1, points.value.length - 1) * 296},${126 - ((secondary ? p.secondary : p.value) ?? 0) / max.value * 112}`).join(' ')
}

function arc(index: number) {
	const start = points.value.slice(0, index).reduce((n, p) => n + (p.value ?? 0), 0)
	return {stroke: colors[index % colors.length], 'stroke-dasharray': `${((points.value[index]?.value ?? 0) / (total.value || 1)) * 100} 100`, 'stroke-dashoffset': -(start / (total.value || 1)) * 100}
}
</script>

<template>
	<div
		v-if="data.kind === 'recent'"
		class="ow-recent"
	>
		<a
			v-for="issue in data.issues"
			:key="issue.key"
			:href="`browse/${issue.key}`"
		><span class="ow-key">{{ issue.key }}</span> {{ issue.summary }}<small>{{ displayDate(issue.updated) }}</small></a>
		<p
			v-if="!data.issues?.length"
			class="ow-caption"
		>
			No matching issues yet.
		</p>
	</div>
	<div
		v-else-if="data.kind === 'activity'"
		class="ow-activity-list"
	>
		<p
			v-for="activity in data.activities"
			:key="activity.id"
		>
			<span class="ow-activity-dot" /><span>{{ activity.detail }}<small>{{ displayDate(activity.at) }}</small></span>
		</p>
		<p
			v-if="!data.activities?.length"
			class="ow-caption"
		>
			Activity will appear as work progresses.
		</p>
	</div>
	<div
		v-else-if="style === 'metric'"
		class="ow-metric"
	>
		<strong>{{ Number(data.value ?? 0).toLocaleString(undefined, {maximumFractionDigits: 1}) }}</strong><span v-if="data.kind === 'cycle-time'">hours average cycle time</span><span v-if="data.kind === 'sprint-progress'">of {{ data.total ?? 0 }} matching issues complete</span><progress
			v-if="data.kind === 'sprint-progress'"
			:value="data.value"
			:max="data.total || 1"
			aria-label="Sprint completion"
		/>
	</div>
	<template v-else>
		<svg
			v-if="style === 'line' && points.length"
			viewBox="0 0 320 145"
			role="img"
			:aria-label="`${data.title || data.kind}, calculated from issue history`"
			class="ow-chart"
		>
			<line
				x1="12"
				y1="126"
				x2="308"
				y2="126"
				stroke="currentColor"
				opacity=".2"
			/>
			<polyline
				:points="line()"
				fill="none"
				stroke="#0b57d0"
				stroke-width="3"
			/>
			<polyline
				v-if="points.some(p => p.secondary)"
				:points="line(true)"
				fill="none"
				stroke="#008675"
				stroke-width="3"
				stroke-dasharray="6 3"
			/>
		</svg>
		<div
			v-else-if="style === 'donut' && total"
			class="ow-donut"
		>
			<svg
				viewBox="0 0 100 100"
				role="img"
				:aria-label="`${data.title || data.kind}: ${total} matching issues`"
			><circle
				cx="50"
				cy="50"
				r="35"
				fill="none"
				stroke="var(--ow-border)"
				stroke-width="13"
			/><circle
				v-for="(point, i) in points"
				:key="point.label"
				cx="50"
				cy="50"
				r="35"
				pathLength="100"
				fill="none"
				stroke-width="13"
				transform="rotate(-90 50 50)"
				v-bind="arc(i)"
			/></svg><strong>{{ total }}</strong>
		</div>
		<div
			v-else-if="style !== 'table'"
			class="ow-bars"
		>
			<div
				v-for="(point, i) in points.slice(0, 12)"
				:key="point.label"
			>
				<span>{{ point.label }}</span><div><i :style="{width: `${(point.value ?? 0) / max * 100}%`, background: colors[i % colors.length]}" /></div><b>{{ Number(point.value ?? 0).toLocaleString() }}</b>
			</div>
		</div>
		<p
			v-if="!points.length"
			class="ow-caption"
		>
			No matching data. Check this widget's project and sprint filters.
		</p>
		<details
			v-if="points.length"
			:open="style === 'table'"
			class="ow-chart-data"
		>
			<summary>View data</summary><table>
				<thead>
					<tr>
						<th>Group / date</th><th>Value</th><th v-if="points.some(p => p.secondary)">
							Second series
						</th>
					</tr>
				</thead><tbody>
					<tr
						v-for="point in points"
						:key="point.label"
					>
						<td>{{ point.label }}</td><td>{{ point.value }}</td><td v-if="points.some(p => p.secondary)">
							{{ point.secondary }}
						</td>
					</tr>
				</tbody>
			</table>
		</details>
	</template>
</template>
