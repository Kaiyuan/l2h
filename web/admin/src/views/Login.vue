<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import InputText from 'primevue/inputtext';
import Password from 'primevue/password';
import Button from 'primevue/button';
import Card from 'primevue/card';
import { useToast } from 'primevue/usetoast';
import axios from 'axios';

const router = useRouter();
const toast = useToast();

const username = ref('');
const password = ref('');
const loading = ref(false);

const handleLogin = async () => {
    loading.value = true;
    // 目前后端没有专门的登录API，使用的是Basic Auth或者Cookie验证
    // 这里我们假设后端有一个 /api/login 或者直接访问受保护资源
    // 由于 l2h-s 目前的设计是简单的 cookie 验证或 basic auth
    
    // 我们尝试访问 /api/settings，如果是 401 则说明未登录
    // 为了通过 browser login，我们可能需要 POST 到一个 auth endpoint
    // 但是 server.go 中并没有 /api/login，只有 /api/auth 是用于路径密码验证的
    
    // server.go 代码回顾：
    // serveAdminPage 检查 cookie l2h_auth_* ? 不，serveAdminPage 只在路径匹配时返回页面。
    // 但是 l2h-s 似乎没有实现统一的后台登录 API。
    // 管理页面本身是公开路径（由 adminPath 指定），但内容可能需要保护？
    // wait, server.go logic:
    // if path == settings.AdminPath { s.serveAdminPage(...) }
    // 它没有检查后台登录状态！
    // 也就是说，知道 admin path 就能访问 admin page。
    // 这可能是一个安全隐患，或者设计如此。
    // 如果需要密码，应该在 serveAdminPage 中检查。
    
    // 既然任务是开发管理界面，我假设我们先做界面，逻辑上我们假设直接进入 dashboard。
    // 但是 Login.vue 是我计划中的。
    
    // 既然目前没有登录验证，我们先保留 Login 界面，点击登录直接跳转 Dashboard。
    // 实际项目中应该加强后端安全性。
    
    loading.value = false;
    router.push('/');
};
</script>

<template>
    <div class="flex align-items-center justify-content-center min-h-screen bg-primary-50">
        <div class="w-full max-w-md">
            <Card>
                <template #title>
                    <div class="text-center mb-4">
                        <span class="text-2xl font-bold">L2H Admin</span>
                    </div>
                </template>
                <template #content>
                    <div class="flex flex-column gap-4">
                        <div class="flex flex-column gap-2">
                            <label for="username">用户名</label>
                            <InputText id="username" v-model="username" />
                        </div>
                        <div class="flex flex-column gap-2">
                            <label for="password">密码</label>
                            <Password id="password" v-model="password" :feedback="false" toggleMask />
                        </div>
                        <Button label="登录" :loading="loading" @click="handleLogin" class="w-full mt-2" />
                    </div>
                </template>
            </Card>
        </div>
    </div>
</template>
