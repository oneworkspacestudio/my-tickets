import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import {fileURLToPath, URL} from 'node:url'

export default defineConfig({
	base: './',
	plugins: [vue()],
	resolve: {alias: {'@': fileURLToPath(new URL('./src', import.meta.url))}},
	build: {
		emptyOutDir: false,
		copyPublicDir: false,
		rolldownOptions: {input: fileURLToPath(new URL('./mytickets.html', import.meta.url))},
	},
})
