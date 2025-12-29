<script setup>
import { RouterView } from 'vue-router'
import AppSidebar from './AppSidebar.vue'
import AppTopbar from './AppTopbar.vue'
</script>

<template>
    <div class="layout-wrapper">
        <app-sidebar></app-sidebar>
        <div class="layout-main-container">
            <app-topbar></app-topbar>
            <div class="layout-main">
                <router-view></router-view>
            </div>
            <div class="layout-footer">
                <span class="font-medium ml-2">L2H Server Admin</span>
            </div>
        </div>
    </div>
</template>

<style scoped>
.layout-wrapper {
    display: flex;
    min-height: 100vh;
}

.layout-main-container {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
    flex: 1;
    margin-left: 250px;
    background-color: var(--p-surface-ground);
}

.layout-main {
    flex: 1;
    padding: 2rem;
}

.layout-footer {
    padding: 1rem 2rem;
    border-top: 1px solid var(--p-surface-border);
    background-color: var(--p-surface-0);
}
</style>
