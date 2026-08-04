<template>
  <div class="min-h-screen flex bg-base">
    <div class="flex-1 flex flex-col items-center justify-center p-6 sm:p-12">
      <div class="w-full max-w-sm text-center mb-8">
        <h2 class="text-2xl font-semibold text-strong mb-2">Đặt lại mật khẩu</h2>
        <p class="text-meta text-sm">Vui lòng nhập mã xác nhận (OTP) và mật khẩu mới</p>
      </div>

      <div class="w-full max-w-sm bg-surface p-6 sm:p-8 rounded-[var(--r-lg)] shadow-e2">
        <form v-if="!successMessage" @submit.prevent="handleReset" class="space-y-5">
          <FormField
            v-model="otp"
            label="Mã xác nhận (OTP)"
            type="text"
            placeholder="123456"
            :icon="Key"
          />
          <FormField
            v-model="newPassword"
            label="Mật khẩu mới"
            type="password"
            placeholder="••••••••"
            :icon="Lock"
          />
          <FormField
            v-model="confirmPassword"
            label="Xác nhận mật khẩu"
            type="password"
            placeholder="••••••••"
            :icon="Lock"
          />

          <div v-if="error" class="flex items-start gap-2 px-3 py-2.5 rounded-md text-sm bg-danger/10 text-danger animate-fade-in">
            <AlertCircle class="w-4 h-4 mt-0.5 shrink-0" />
            <span>{{ error }}</span>
          </div>

          <BaseButton type="submit" variant="primary" size="lg" :loading="loading" block>
            Đặt lại mật khẩu
          </BaseButton>
        </form>
        
        <div v-else class="text-center">
          <div class="w-12 h-12 rounded-full bg-success/10 text-success flex items-center justify-center mx-auto mb-4">
            <CheckCircle class="w-6 h-6" />
          </div>
          <p class="text-strong font-medium mb-2">{{ successMessage }}</p>
          <p class="text-meta text-sm mb-6">Bạn đã có thể đăng nhập bằng mật khẩu mới.</p>
          
          <router-link :to="{ name: 'Login' }">
            <BaseButton variant="primary" block>
              Về trang đăng nhập
            </BaseButton>
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRoute } from 'vue-router';
import api from '../services/api';
import { Key, Lock, AlertCircle, CheckCircle } from 'lucide-vue-next';
import FormField from '../components/ui/FormField.vue';
import BaseButton from '../components/ui/BaseButton.vue';

const route = useRoute();
const identifier = ref(route.query.identifier || '');
const otp = ref('');
const newPassword = ref('');
const confirmPassword = ref('');

const loading = ref(false);
const error = ref('');
const successMessage = ref('');

const handleReset = async () => {
  if (!otp.value || !newPassword.value) {
    error.value = 'Vui lòng điền đủ thông tin';
    return;
  }
  if (newPassword.value !== confirmPassword.value) {
    error.value = 'Mật khẩu xác nhận không khớp';
    return;
  }
  
  loading.value = true;
  error.value = '';
  
  try {
    const res = await api.post('/auth/reset-password', {
      identifier: identifier.value,
      otp: otp.value,
      new_password: newPassword.value
    });
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
