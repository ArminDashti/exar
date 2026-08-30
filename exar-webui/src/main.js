import { createApp } from 'vue'
import App from './App.vue'
import { router } from './router'
import { registerSW } from 'virtual:pwa-register'
import './style.css'

registerSW({ immediate: true })

createApp(App).use(router).mount('#app')
