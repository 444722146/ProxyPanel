import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '@/views/Dashboard.vue'
import ProxyList from '@/views/ProxyList.vue'
import Logs from '@/views/Logs.vue'
import SSL from '@/views/SSL.vue'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard
  },
  {
    path: '/proxy',
    name: 'ProxyList',
    component: ProxyList
  },
  {
    path: '/logs',
    name: 'Logs',
    component: Logs
  },
  {
    path: '/ssl',
    name: 'SSL',
    component: SSL
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router