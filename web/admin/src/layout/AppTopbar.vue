<script setup>
import { useRouter } from 'vue-router';
import Button from 'primevue/button';

const router = useRouter();

const logout = () => {
    // 这里可以清除 cookie 或调用登出 API
    document.cookie = 'l2h_auth=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
    router.push('/login');
};
</script>

<template>
    <div class="layout-topbar">
        <div class="layout-topbar-actions">
            <Button icon="pi pi-sign-out" class="p-button-text" @click="logout" aria-label="Logout" />
        </div>
    </div>
</template>

<style scoped>
.layout-topbar {
    height: 4rem;
    padding: 0 2rem;
    background-color: var(--p-surface-0);
    border-bottom: 1px solid var(--p-surface-border);
    display: flex;
    align-items: center;
    justify-content: flex-end;
}
</style>
