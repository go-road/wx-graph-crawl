import { createWebHistory, createRouter } from 'vue-router'

const routes = [
    {
        path: '/',
        name: 'index-page',
        redirect: '/home',
    },
    {
        path: '/home',
        name: 'home',
        component: () => import('@/views/home/Index.vue'),
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes,
})

export default router