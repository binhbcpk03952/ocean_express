<template>
  <div class="min-h-screen flex items-center justify-center bg-base p-6 sm:p-12">
    <div class="w-full max-w-md">
      <!-- Logo -->
      <div class="flex items-center gap-3 mb-8 justify-center">
        <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-gradient-to-br from-sky-400 to-teal-500">
          <Store class="w-5 h-5 text-white" />
        </div>
        <span class="text-lg font-bold text-strong tracking-wide">OCEAN EXPRESS</span>
      </div>

      <div class="bg-surface border rounded-[var(--r-lg)] shadow-e2 p-6 sm:p-8">
        <!-- Thành công -->
        <div v-if="done" class="text-center py-6">
          <div class="w-14 h-14 rounded-full bg-[var(--st-delivered-bg)] flex items-center justify-center mx-auto mb-4">
            <CircleCheck class="w-8 h-8" :style="{ color: 'var(--st-delivered-fg)' }" />
          </div>
          <h2 class="text-xl font-semibold text-strong mb-2">Đăng ký thành công!</h2>
          <p class="text-meta text-sm mb-6">
            Tài khoản của bạn đã được <b>kích hoạt ngay lập tức</b>. Bạn có thể đăng nhập và sử dụng hệ thống ngay bây giờ.
          </p>
          <router-link :to="{ name: 'ShopLogin' }">
            <BaseButton variant="primary" block>Đăng nhập ngay</BaseButton>
          </router-link>
        </div>

        <!-- Form -->
        <template v-else>
          <div class="mb-6">
            <h2 class="text-2xl font-semibold text-strong mb-1">Đăng ký đối tác</h2>
            <p class="text-meta text-sm">Tạo tài khoản Shop để gửi hàng qua Ocean Express</p>
          </div>

          <form @submit.prevent="handleRegister" class="space-y-4">
            <FormField v-model="form.name" label="Tên shop" placeholder="VD: Cửa hàng ABC" required />
            <FormField v-model="form.phone" label="Số điện thoại" placeholder="VD: 0987654321" required />
            <FormField v-model="form.email" label="Email đăng nhập" type="email" placeholder="shop@example.com" required />
            <FormField v-model="form.password" label="Mật khẩu" type="password" placeholder="••••••••" required />
            <LocationCascader v-model="form.location_id" label="Khu vực gửi hàng" required hint="Bắt buộc — dùng để tính cước khi bạn tạo đơn." />
            <FormField v-model="form.address_detail" label="Địa chỉ lấy hàng" placeholder="Số nhà, đường, phường..." required />

            <div v-if="error" class="flex items-start gap-2 px-3 py-2.5 rounded-md text-sm bg-danger/10 text-danger animate-fade-in">
              <AlertCircle class="w-4 h-4 mt-0.5 shrink-0" />
              <span>{{ error }}</span>
            </div>

            <BaseButton type="submit" variant="primary" size="lg" :loading="loading" block>
              Đăng ký
            </BaseButton>
          </form>

          <div class="mt-6 text-center text-xs text-meta">
            Đã có tài khoản?
            <router-link :to="{ name: 'ShopLogin' }" class="text-[var(--primary)] font-medium hover:underline">Đăng nhập</router-link>
          </div>
        </template>
      </div>

      <p class="text-center text-meta text-xs mt-6">
        Bạn là tài xế?
        <router-link :to="{ name: 'RegisterShipper' }" class="text-[var(--primary)] hover:underline">Đăng ký làm tài xế</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import api from '../../services/api';
import { Store, AlertCircle, CircleCheck } from 'lucide-vue-next';
import FormField from '../../components/ui/FormField.vue';
import LocationCascader from '../../components/ui/LocationCascader.vue';
import BaseButton from '../../components/ui/BaseButton.vue';

const form = ref({ name: '', phone: '', email: '', password: '', location_id: '', address_detail: '' });
const locations = ref([]);
const loading = ref(false);
const error = ref('');
const done = ref(false);

const fetchLocations = async () => {
  // location data is now handled inside LocationCascader
};

onMounted(fetchLocations);

const handleRegister = async () => {
  if (!form.value.name || !form.value.phone || !form.value.email || !form.value.password || !form.value.address_detail) {
    error.value = 'Vui lòng điền đầy đủ thông tin bắt buộc';
    return;
  }
  if (!form.value.location_id) {
    error.value = 'Vui lòng chọn khu vực gửi hàng';
    return;
  }
  error.value = '';
  loading.value = true;
  try {
    const payload = { ...form.value };
    const res = await api.post('/shops/register', payload);
    if (res.success) done.value = true;
  } catch (err) {
    error.value = err.response?.data?.error?.message || err.response?.data?.error || 'Đăng ký thất bại';
  } finally {
    loading.value = false;
  }
};
</script>
