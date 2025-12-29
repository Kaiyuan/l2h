<script setup>
import { ref, onMounted } from 'vue';
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import Button from 'primevue/button';
import Dialog from 'primevue/dialog';
import InputText from 'primevue/inputtext';
import InputNumber from 'primevue/inputnumber';
import Password from 'primevue/password';
import { useToast } from 'primevue/usetoast';
import { useConfirm } from 'primevue/useconfirm';
import axios from 'axios';

const toast = useToast();
const confirm = useConfirm();

const paths = ref([]);
const loading = ref(false);
const dialogVisible = ref(false);
const saving = ref(false);

const form = ref({
    path: '',
    server_b_port: 55055,
    password: ''
});

const loadPaths = async () => {
    loading.value = true;
    try {
        const res = await axios.get('/api/paths');
        paths.value = res.data || [];
    } catch (e) {
        toast.add({ severity: 'error', summary: 'Error', detail: '无法加载路径列表', life: 3000 });
    } finally {
        loading.value = false;
    }
};

const openAddDialog = () => {
    form.value = { path: '', server_b_port: 55055, password: '' };
    dialogVisible.value = true;
};

const savePath = async () => {
    if (!form.value.path) {
        toast.add({ severity: 'warn', summary: 'Validation', detail: '路径不能为空', life: 3000 });
        return;
    }
    saving.value = true;
    try {
        await axios.post('/api/paths', form.value);
        toast.add({ severity: 'success', summary: 'Success', detail: '路径添加成功', life: 3000 });
        dialogVisible.value = false;
        loadPaths();
    } catch (e) {
        toast.add({ severity: 'error', summary: 'Error', detail: '添加失败: ' + (e.response?.data?.message || e.message), life: 3000 });
    } finally {
        saving.value = false;
    }
};

const deletePath = (id) => {
    confirm.require({
        message: '确定要删除这个路径吗?',
        header: '确认删除',
        icon: 'pi pi-exclamation-triangle',
        accept: async () => {
            try {
                await axios.delete(`/api/paths/${id}`);
                toast.add({ severity: 'success', summary: 'Success', detail: '删除成功', life: 3000 });
                loadPaths();
            } catch (e) {
                toast.add({ severity: 'error', summary: 'Error', detail: '删除失败', life: 3000 });
            }
        }
    });
};

onMounted(() => {
    loadPaths();
});
</script>

<template>
    <div class="card">
        <div class="flex justify-between items-center mb-4">
            <h1 class="text-2xl font-bold">路径管理</h1>
            <Button label="添加路径" icon="pi pi-plus" @click="openAddDialog" />
        </div>

        <DataTable :value="paths" :loading="loading" stripedRows>
            <Column field="id" header="ID" sortable></Column>
            <Column field="path" header="路径" sortable></Column>
            <Column field="server_b_port" header="Server B 端口" sortable></Column>
            <Column header="密码保护">
                <template #body="slotProps">
                    <span v-if="slotProps.data.password" class="text-green-500 font-bold">是</span>
                    <span v-else class="text-gray-400">否</span>
                </template>
            </Column>
            <Column header="操作">
                <template #body="slotProps">
                    <Button icon="pi pi-trash" severity="danger" text rounded @click="deletePath(slotProps.data.id)" />
                </template>
            </Column>
            <template #empty>暂无数据</template>
        </DataTable>

        <Dialog v-model:visible="dialogVisible" header="添加新路径" modal :style="{ width: '400px' }">
            <div class="flex flex-column gap-4">
                <div class="flex flex-column gap-2">
                    <label for="path">路径 (URL Path)</label>
                    <InputText id="path" v-model="form.path" placeholder="e.g. my-service" />
                </div>
                <div class="flex flex-column gap-2">
                    <label for="port">Server B 端口</label>
                    <InputNumber id="port" v-model="form.server_b_port" :useGrouping="false" />
                </div>
                <div class="flex flex-column gap-2">
                    <label for="password">访问密码 (可选)</label>
                    <Password id="password" v-model="form.password" :feedback="false" toggleMask />
                </div>
            </div>
            <template #footer>
                <Button label="取消" text @click="dialogVisible = false" />
                <Button label="保存" @click="savePath" :loading="saving" />
            </template>
        </Dialog>
    </div>
</template>
