<script setup>
import { ref, onMounted } from 'vue';
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import Button from 'primevue/button';
import Dialog from 'primevue/dialog';
import InputText from 'primevue/inputtext';
import InputNumber from 'primevue/inputnumber';
import { useToast } from 'primevue/usetoast';
import { useConfirm } from 'primevue/useconfirm';
import axios from 'axios';

const toast = useToast();
const confirm = useConfirm();

const apiKeys = ref([]);
const loading = ref(false);
const dialogVisible = ref(false);
const saving = ref(false);
const generatedKey = ref('');

const form = ref({
    name: '',
    expires_in_days: 0
});

const loadKeys = async () => {
    loading.value = true;
    try {
        const res = await axios.get('/api/api-keys');
        apiKeys.value = res.data || [];
    } catch (e) {
        toast.add({ severity: 'error', summary: 'Error', detail: '无法加载 API Keys', life: 3000 });
    } finally {
        loading.value = false;
    }
};

const openAddDialog = () => {
    form.value = { name: '', expires_in_days: 30 };
    generatedKey.value = '';
    dialogVisible.value = true;
};

const generateKey = async () => {
    if (!form.value.name) {
        toast.add({ severity: 'warn', summary: 'Validation', detail: '名称不能为空', life: 3000 });
        return;
    }
    saving.value = true;
    try {
        const res = await axios.post('/api/api-keys', form.value);
        generatedKey.value = res.data.key;
        toast.add({ severity: 'success', summary: 'Success', detail: 'Key 生成成功', life: 3000 });
        loadKeys();
    } catch (e) {
        toast.add({ severity: 'error', summary: 'Error', detail: '生成失败', life: 3000 });
    } finally {
        saving.value = false;
    }
};

const deleteKey = (id) => {
    confirm.require({
        message: '确定要删除这个 API Key 吗?',
        header: '确认删除',
        icon: 'pi pi-exclamation-triangle',
        accept: async () => {
            try {
                await axios.delete(`/api/api-keys/${id}`);
                toast.add({ severity: 'success', summary: 'Success', detail: '删除成功', life: 3000 });
                loadKeys();
            } catch (e) {
                toast.add({ severity: 'error', summary: 'Error', detail: '删除失败', life: 3000 });
            }
        }
    });
};

const copyKey = () => {
    navigator.clipboard.writeText(generatedKey.value);
    toast.add({ severity: 'info', summary: 'Copied', detail: '已复制到剪贴板', life: 2000 });
};

onMounted(() => {
    loadKeys();
});
</script>

<template>
    <div class="card">
        <div class="flex justify-between items-center mb-4">
            <h1 class="text-2xl font-bold">API Key 管理</h1>
            <Button label="生成新 Key" icon="pi pi-plus" @click="openAddDialog" />
        </div>

        <DataTable :value="apiKeys" :loading="loading" stripedRows>
            <Column field="id" header="ID" sortable></Column>
            <Column field="name" header="名称" sortable></Column>
            <Column field="key" header="Key (部分)" sortable>
                <template #body="slotProps">
                    <span class="font-mono text-gray-500">{{ slotProps.data.key.substring(0, 8) }}...</span>
                </template>
            </Column>
            <Column field="expires_at" header="过期时间" sortable>
                <template #body="slotProps">
                    {{ slotProps.data.expires_at ? new Date(slotProps.data.expires_at).toLocaleString() : '永久有效' }}
                </template>
            </Column>
            <Column field="usage_count" header="使用次数" sortable></Column>
            <Column header="操作">
                <template #body="slotProps">
                    <Button icon="pi pi-trash" severity="danger" text rounded @click="deleteKey(slotProps.data.id)" />
                </template>
            </Column>
            <template #empty>暂无 API Key</template>
        </DataTable>

        <Dialog v-model:visible="dialogVisible" header="生成 API Key" modal :style="{ width: '500px' }">
            <div v-if="!generatedKey" class="flex flex-column gap-4">
                <div class="flex flex-column gap-2">
                    <label for="name">名称 / 描述</label>
                    <InputText id="name" v-model="form.name" placeholder="e.g. My Worker Server" />
                </div>
                <div class="flex flex-column gap-2">
                    <label for="expires">有效期 (天)</label>
                    <InputNumber id="expires" v-model="form.expires_in_days" suffix=" 天" />
                    <small class="text-gray-500">设置为 0 表示永久有效</small>
                </div>
            </div>
            
            <div v-else class="flex flex-column gap-4">
                <div class="p-3 bg-green-50 rounded border border-green-200">
                    <p class="text-green-700 font-bold mb-2">生成成功! 请立即复制保存，稍微将不再显示。</p>
                    <div class="flex gap-2">
                        <InputText :value="generatedKey" readonly class="w-full font-mono bg-white" />
                        <Button icon="pi pi-copy" @click="copyKey" />
                    </div>
                </div>
            </div>

            <template #footer>
                <Button v-if="!generatedKey" label="取消" text @click="dialogVisible = false" />
                <Button v-if="!generatedKey" label="生成" @click="generateKey" :loading="saving" />
                <Button v-if="generatedKey" label="完成" @click="dialogVisible = false" />
            </template>
        </Dialog>
    </div>
</template>
