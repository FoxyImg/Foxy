import './assets/main.css'

import FloatingVue from 'floating-vue';
import 'floating-vue/dist/style.css';

import { createApp } from 'vue'
import App from './App.vue'

createApp(App).use(FloatingVue).mount('#app')
