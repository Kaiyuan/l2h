<script setup>
import { ref, onMounted } from 'vue';
import Card from 'primevue/card';
import axios from 'axios';

const stats = ref({
    pathsCount: 0,
    apiKeysCount: 0,
    serverVersion: '1.0.0'
});

const loadStats = async () => {
    try {
        const [pathsRes, keysRes] = await Promise.all([
            axios.get('/api/paths'),
            axios.get('/api/api-keys')
        ]);
        stats.value.pathsCount = pathsRes.data ? pathsRes.data.length : 0;
        stats.value.apiKeysCount = keysRes.data ? keysRes.data.length : 0;
    } catch (e) {
        console.error('Failed to load stats', e);
    }
};

onMounted(() => {
    loadStats();
});
</script>

<template>
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card>
            <template #title>路径数量</template>
            <template #content>
                <div class="text-4xl font-bold text-primary">{{ stats.pathsCount }}</div>
            </template>
        </Card>
        <Card>
            <template #title>API Key 数量</template>
            <template #content>
                <div class="text-4xl font-bold text-primary">{{ stats.apiKeysCount }}</div>
            </template>
        </Card>
        <Card>
            <template #title>系统版本</template>
            <template #content>
                <div class="text-4xl font-bold text-gray-500">{{ stats.serverVersion }}</div>
            </template>
        </Card>
    </div>
</template>
