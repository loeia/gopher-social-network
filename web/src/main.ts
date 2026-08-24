import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

import App from '@/App.vue'
import router from './router'

export const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus)

app.mount('#app')
