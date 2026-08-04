<template>
  <div class="min-h-screen flex bg-base">
    <!-- Brand panel (trái) -->
    <div class="hidden lg:flex lg:w-1/2 relative overflow-hidden bg-navy-900">
      <!-- Lớp gradient trang trí -->
      <div
        class="absolute inset-0"
        style="background: radial-gradient(1200px 600px at 20% 10%, rgba(6,182,212,0.25), transparent 55%), radial-gradient(900px 500px at 90% 90%, rgba(14,165,233,0.18), transparent 50%);"
      ></div>

      <!-- Lưới điểm mờ -->
      <div
        class="absolute inset-0 opacity-[0.15]"
        style="background-image: radial-gradient(rgba(255,255,255,0.4) 1px, transparent 1px); background-size: 28px 28px;"
      ></div>

      <div class="relative z-10 flex flex-col justify-between p-12 text-white w-full">
        <div class="flex items-center gap-3">
          <div class="w-11 h-11 rounded-xl flex items-center justify-center bg-gradient-to-br from-sky-400 to-teal-500 shadow-lg">
            <svg class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 16.5 12 21l9-4.5" /><path d="M3 12l9 4.5L21 12" /><path d="M12 3 3 7.5 12 12l9-4.5L12 3z" />
            </svg>
          </div>
          <div class="leading-tight">
            <div class="text-lg font-bold tracking-wide">OCEAN EXPRESS</div>
            <div class="text-xs text-sky-200/80">Logistics Platform</div>
          </div>
        </div>

        <div class="max-w-md">
          <h1 class="text-4xl font-bold leading-tight mb-4">
            Điều phối vận chuyển,<br />trong tầm kiểm soát.
          </h1>
          <p class="text-sky-100/70 text-base leading-relaxed">
            Quản lý vận đơn, bưu cục, tài xế và dòng chảy hàng hóa toàn quốc trên một nền tảng thống nhất.
          </p>
        </div>

        <div class="flex items-center gap-8 text-sky-100/60 text-sm">
          <div><span class="block text-2xl font-bold text-white tabular">4</span> phân hệ vận hành</div>
          <div><span class="block text-2xl font-bold text-white tabular">24/7</span> theo dõi realtime</div>
          <div><span class="block text-2xl font-bold text-white tabular">GPS</span> lộ trình giao nhận</div>
        </div>
      </div>
    </div>

    <!-- Form panel (phải) -->
    <div class="flex-1 flex items-center justify-center p-6 sm:p-12">
      <div class="w-full max-w-sm">
        <!-- Logo mobile -->
        <div class="lg:hidden flex items-center gap-3 mb-10 justify-center">
          <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-gradient-to-br from-sky-400 to-teal-500">
            <svg class="w-5 h-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 16.5 12 21l9-4.5" /><path d="M3 12l9 4.5L21 12" /><path d="M12 3 3 7.5 12 12l9-4.5L12 3z" />
            </svg>
          </div>
          <span class="text-lg font-bold text-strong tracking-wide">OCEAN EXPRESS</span>
        </div>

        <div class="mb-8">
          <h2 class="text-2xl font-semibold text-strong mb-1">Đăng nhập</h2>
          <p class="text-meta text-sm">Truy cập hệ thống quản trị vận hành</p>
        </div>

        <form @submit.prevent="handleLogin" class="space-y-5">
          <FormField
            v-model="identifier"
            label="Số điện thoại hoặc Email"
            placeholder="09... hoặc email@..."
            :icon="User"
            autocomplete="username"
          />
          <div>
            <FormField
              v-model="password"
              label="Mật khẩu"
              type="password"
              placeholder="••••••••"
              :icon="Lock"
              autocomplete="current-password"
            />
            <div class="mt-1.5 flex justify-end">
              <router-link :to="{ name: 'ForgotPassword' }" class="text-xs text-[var(--primary)] hover:underline">Quên mật khẩu?</router-link>
            </div>
          </div>

          <div v-if="error" class="flex items-start gap-2 px-3 py-2.5 rounded-md text-sm bg-danger/10 text-danger animate-fade-in">
            <AlertCircle class="w-4 h-4 mt-0.5 shrink-0" />
            <span>{{ error }}</span>
          </div>

          <BaseButton type="submit" variant="primary" size="lg" :loading="loading" block>
            Đăng nhập
          </BaseButton>
        </form>

        <div class="mt-6 flex items-center justify-between text-xs">
          <router-link :to="{ name: 'RegisterShipper' }" class="text-[var(--primary)] font-medium hover:underline">
            Đăng ký tài xế
          </router-link>
          <router-link :to="{ name: 'ShopLogin' }" class="text-meta hover:text-body">
            Cổng đối tác (Shop)
          </router-link>
        </div>

        <p class="text-center text-meta text-xs mt-8">
          © 2026 Ocean Express — Core Logistics System
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/authStore';
import { Phone, Lock, AlertCircle, User } from 'lucide-vue-next';
import FormField from '../components/ui/FormField.vue';
import BaseButton from '../components/ui/BaseButton.vue';

const router = useRouter();
const authStore = useAuthStore();

const identifier = ref('');
const password = ref('');
const error = ref('');
const loading = ref(false);

const handleLogin = async () => {
  error.value = '';
  loading.value = true;
  try {
    if (!identifier.value || !password.value) {
      error.value = 'Vui lòng nhập đầy đủ thông tin';
      loading.value = false;
      return;
    }
    const success = await authStore.login(identifier.value, password.value);
    if (success) {
      router.push(authStore.homeRoute());
    } else {
      error.value = 'Đăng nhập thất bại';
    }
  } catch (err) {
    error.value = err.response?.data?.error?.message || err.response?.data?.error || 'Có lỗi xảy ra, vui lòng thử lại';
  } finally {
    loading.value = false;
  }
};
</script>
