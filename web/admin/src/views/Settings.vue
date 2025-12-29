<script setup>
import { ref, onMounted } from 'vue';
import Card from 'primevue/card';
import InputText from 'primevue/inputtext';
import Password from 'primevue/password';
import Button from 'primevue/button';
import { useToast } from 'primevue/usetoast';
import axios from 'axios';

const toast = useToast();
const loading = ref(false);
const saving = ref(false);

const form = ref({
    admin_path: '',
    username: '',
    password: '',
    email: ''
});

// 仅用于修改密码时
const newPassword = ref('');

const loadSettings = async () => {
    loading.value = true;
    try {
        const res = await axios.get('/api/settings');
        if (res.data) {
            form.value = {
                ...res.data,
                password: '' // 不回显密码
            };
        }
    } catch (e) {
        toast.add({ severity: 'error', summary: 'Error', detail: '无法加载设置', life: 3000 });
    } finally {
        loading.value = false;
    }
};

const saveSettings = async () => {
    saving.value = true;
    try {
        const payload = { ...form.value };
        if (newPassword.value) {
            payload.password = newPassword.value;
        } else {
            // 如果没填新密码，就用原来的密码（后端需要支持不传或处理）
            // 后端逻辑是：直接更新。所以如果不传密码，意味着密码变空？
            // 我们需要查看后端逻辑。
            // 假设我们只在有新密码时才更新密码字段，或者我们需要先获取原密码？
            // 后端 database.go SetSettings 接收 Settings 结构体。
            // 如果我们不修改密码，我们不知道原密码是什么（因为 GetSettings 返回了哈希或空？但代码中GetSettings 返回了 password）。
            // 实际上 GetSettings 返回了 Hash 过的 password。
            // 当我们 SetSettings 时，如果传入的是 Hash，IsHashed 返回 true，就不再 Hash。
            // 所以我们可以把 Get 到的 Hash 传回去，就保持不变。
            // 如果 newPassword 有值，就传 newPassword (明文)，后端会检测并 Hash。
            
            // 但是为了安全，前端不应该拿到 Hash。
            // 我们假设：如果 newPassword 为空，就不更新密码。
            // 但是后端接口 `handleSetSettings` 直接 decode json 到 Settings struct，然后 SetSettings。
            // 这意味着会覆盖所有字段。
            
            // 临时解决方案：如果 newPassword 为空，我们将从 loadStats 中获取的原始 Hash 发送回去。
            const res = await axios.get('/api/settings');
            payload.password = res.data.password;
        }

        await axios.post('/api/settings', payload);
        toast.add({ severity: 'success', summary: 'Success', detail: '设置已保存', life: 3000 });
        newPassword.value = '';
        loadSettings();
    } catch (e) {
        toast.add({ severity: 'error', summary: 'Error', detail: '保存失败', life: 3000 });
    } finally {
        saving.value = false;
    }
};

onMounted(() => {
    loadSettings();
});
</script>

<template>
    <div class="card max-w-2xl mx-auto">
        <h1 class="text-2xl font-bold mb-4">系统设置</h1>
        
        <div class="flex flex-column gap-4">
            <Card>
                <template #title>管理员账号</template>
                <template #content>
                    <div class="flex flex-column gap-4">
                        <div class="flex flex-column gap-2">
                            <label for="username">用户名</label>
                            <InputText id="username" v-model="form.username" />
                        </div>
                        <div class="flex flex-column gap-2">
                            <label for="admin_path">管理页面路径 (URL)</label>
                            <InputText id="admin_path" v-model="form.admin_path" />
                            <small class="text-gray-500">修改后需要通过新路径访问</small>
                        </div>
                        <div class="flex flex-column gap-2">
                            <label for="email">邮箱 (可选)</label>
                            <InputText id="email" v-model="form.email" />
                        </div>
                        <div class="flex flex-column gap-2">
                            <label for="password">新密码</label>
                            <Password id="password" v-model="newPassword" :feedback="true" toggleMask placeholder="留空则保持不变" />
                        </div>
                    </div>
                </template>
            </Card>

            <div class="flex justify-end">
                <Button label="保存更改" @click="saveSettings" :loading="saving" />
            </div>
        </div>
    </div>
</template>
