<template>
  <div class="min-h-screen flex bg-base">
    <div class="flex-1 flex flex-col items-center justify-center p-6 sm:p-12">
      <div class="w-full max-w-sm text-center mb-8">
        <h2 class="text-2xl font-semibold text-strong mb-2">Quên mật khẩu?</h2>
        <p class="text-meta text-sm">Nhập email hoặc số điện thoại để đặt lại mật khẩu</p>
      </div>

      <div class="w-full max-w-sm bg-surface p-6 sm:p-8 rounded-[var(--r-lg)] shadow-e2">
        <form v-if="!successMessage" @submit.prevent="handleForgot" class="space-y-5">
          <FormField
            v-model="identifier"
            label="Số điện thoại hoặc Email"
            type="text"
            placeholder="09... hoặc email@example.com"
            :icon="User"
          />

          <div v-if="error" class="flex items-start gap-2 px-3 py-2.5 rounded-md text-sm bg-danger/10 text-danger animate-fade-in">
            <AlertCircle class="w-4 h-4 mt-0.5 shrink-0" />
            <span>{{ error }}</span>
          </div>

          <BaseButton type="submit" variant="primary" size="lg" :loading="loading" block>
            Gửi yêu cầu
          </BaseButton>
        </form>
        
        <div v-else class="text-center">
          <div class="w-12 h-12 rounded-full bg-success/10 text-success flex items-center justify-center mx-auto mb-4">
            <CheckCircle class="w-6 h-6" />
          </div>
          <p class="text-strong font-medium mb-2">{{ successMessage }}</p>
          <p class="text-meta text-sm mb-6">Vui lòng kiểm tra tin nhắn hoặc email để nhận mã OTP.</p>
          
          <router-link :to="{ name: 'ResetPassword', query: { identifier } }">
            <BaseButton variant="primary" block>
              Nhập mã xác nhận
            </BaseButton>
          </router-link>
        </div>

        <div class="mt-6 text-center text-sm">
          <router-link :to="{ name: 'Login' }" class="text-meta hover:text-[var(--primary)] transition-colors">
            ← Quay lại đăng nhập
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import api from '../services/api';
import { User, AlertCircle, CheckCircle } from 'lucide-vue-next';
import FormField from '../components/ui/FormField.vue';
import BaseButton from '../components/ui/BaseButton.vue';

const identifier = ref('');
const loading = ref(false);
const error = ref('');
const successMessage = ref('');

const handleForgot = async () => {
  if (!identifier.value) {
    error.value = 'Vui lòng nhập số điện thoại hoặc email';
    return;
  }
  
  loading.value = true;
  error.value = '';
  
  try {
    const res = await api.post('/auth/forgot-password', { identifier: identifier.value });
    if (res.success) {
      successMessage.value = res.message;
    }
  } catch (err) {
    error.value = err.response?.data?.error || 'Có lỗi xảy ra, vui lòng thử lại';
  } finally {
    loading.value = false;
  }
};
</script>
