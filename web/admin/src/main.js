import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

import PrimeVue from 'primevue/config'
import Aura from '@primevue/themes/aura'
import ToastService from 'primevue/toastservice';
import ConfirmationService from 'primevue/confirmationservice';

import './assets/main.css'
import 'primeicons/primeicons.css'

const app = createApp(App)

app.use(router)
app.use(PrimeVue, {
    theme: {
        preset: Aura,
        options: {
            darkModeSelector: '.dark',
        }
    }
})
app.use(ToastService)
app.use(ConfirmationService)

app.mount('#app')
