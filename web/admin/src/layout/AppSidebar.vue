<script setup>
import { ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';

const router = useRouter();
const route = useRoute();

const menu = ref([
    { label: '仪表盘', icon: 'pi pi-home', to: '/' },
    { label: '路径管理', icon: 'pi pi-link', to: '/paths' },
    { label: 'API 密钥', icon: 'pi pi-key', to: '/api-keys' },
    { label: '系统设置', icon: 'pi pi-cog', to: '/settings' }
]);
</script>

<template>
    <div class="layout-sidebar">
        <div class="sidebar-header">
            <span class="text-xl font-bold text-primary">L2H Admin</span>
        </div>
        <ul class="layout-menu">
            <li v-for="item in menu" :key="item.to" class="layout-menuitem">
                <router-link :to="item.to" class="layout-menuitem-root-text" :class="{ 'active': route.path === item.to }">
                    <i :class="item.icon" class="mr-2"></i>
                    <span>{{ item.label }}</span>
                </router-link>
            </li>
        </ul>
    </div>
</template>

<style scoped>
.layout-sidebar {
    position: fixed;
    width: 250px;
    height: 100vh;
    top: 0;
    left: 0;
    background-color: var(--p-surface-0);
    border-right: 1px solid var(--p-surface-border);
    display: flex;
    flex-direction: column;
    z-index: 999;
}

.sidebar-header {
    height: 4rem;
    display: flex;
    align-items: center;
    padding: 0 1.5rem;
    border-bottom: 1px solid var(--p-surface-border);
}

.layout-menu {
    list-style: none;
    margin: 0;
    padding: 1rem;
    flex: 1;
    overflow-y: auto;
}

.layout-menuitem-root-text {
    display: flex;
    align-items: center;
    padding: 0.75rem 1rem;
    border-radius: var(--p-border-radius);
    color: var(--p-text-color);
    transition: background-color 0.2s;
    text-decoration: none;
    margin-bottom: 0.5rem;
}

.layout-menuitem-root-text:hover {
    background-color: var(--p-surface-hover);
}

.layout-menuitem-root-text.active {
    background-color: var(--p-primary-50);
    color: var(--p-primary-600);
}
</style>
