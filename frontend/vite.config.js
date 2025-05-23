import {defineConfig} from 'vite'
import {dirname, resolve} from 'node:path'
import {fileURLToPath} from 'node:url'

import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

const __dirname = dirname(fileURLToPath(import.meta.url))

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [
        vue(),
        tailwindcss(),
    ],
    resolve: {
        alias: {
            '@': resolve(__dirname, 'src'),
            'wailsjs': resolve(__dirname, 'wailsjs'),
        },
    },

})
