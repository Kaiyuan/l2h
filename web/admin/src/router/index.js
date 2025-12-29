import { createRouter, createWebHistory } from 'vue-router'
import AppLayout from '@/layout/AppLayout.vue'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            component: AppLayout,
            children: [
                {
                    path: '/',
                    name: 'dashboard',
                    component: () => import('@/views/Dashboard.vue')
                },
                {
                    path: '/paths',
                    name: 'paths',
                    component: () => import('@/views/Paths.vue')
                },
                {
                    path: '/api-keys',
                    name: 'api-keys',
                    component: () => import('@/views/APIKeys.vue')
                },
                {
                    path: '/settings',
                    name: 'settings',
                    component: () => import('@/views/Settings.vue')
                }
            ]
        },
        {
            path: '/login',
            name: 'login',
            component: () => import('@/views/Login.vue')
        }
    ]
})

export default router
