<template>
  <div class="min-h-screen flex items-center justify-center bg-base p-6 sm:p-12">
    <div class="w-full max-w-md">
      <!-- Logo -->
      <div class="flex items-center gap-3 mb-8 justify-center">
        <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-gradient-to-br from-sky-400 to-teal-500">
          <Truck class="w-5 h-5 text-white" />
        </div>
        <span class="text-lg font-bold text-strong tracking-wide">OCEAN EXPRESS</span>
      </div>

      <div class="bg-surface border rounded-[var(--r-lg)] shadow-e2 p-6 sm:p-8">
        <!-- Đã đăng ký xong -->
        <div v-if="done" class="text-center py-6">
          <div class="w-14 h-14 rounded-full bg-[var(--st-delivered-bg)] flex items-center justify-center mx-auto mb-4">
            <CircleCheck class="w-8 h-8" style="color: var(--st-delivered-fg)" />
          </div>
          <h2 class="text-xl font-semibold text-strong mb-2">Đăng ký thành công</h2>
          <p class="text-meta text-sm mb-6">
            Tài khoản của bạn đang chờ Admin duyệt. Bạn sẽ đăng nhập được sau khi được phê duyệt.
          </p>
          <router-link :to="{ name: 'Login' }">
            <BaseButton variant="primary" size="md" class="w-full">Về trang đăng nhập</BaseButton>
          </router-link>
        </div>

        <!-- Form đăng ký -->
        <template v-else>
          <div class="mb-6">
            <h2 class="text-2xl font-semibold text-strong mb-1">Đăng ký tài xế</h2>
            <p class="text-meta text-sm">Tạo tài khoản shipper và chờ Admin duyệt</p>
          </div>

          <form @submit.prevent="handleRegister" class="space-y-4">
            <FormField v-model="form.name" label="Họ tên" required placeholder="Nguyễn Văn A" :icon="User" />
            <FormField v-model="form.phone" label="Số điện thoại" required placeholder="09..." :icon="Phone" />
            <FormField v-model="form.email" label="Email" placeholder="abc@email.com" :icon="User" />
            <FormField v-model="form.password" label="Mật khẩu" type="password" required placeholder="••••••••" :icon="Lock" />

            <FormSelect v-model="form.role" label="Loại tài xế" required>
              <option value="first_mile_driver">Tài xế lấy hàng (First-mile)</option>
              <option value="last_mile_driver">Tài xế giao hàng (Last-mile)</option>
            </FormSelect>

            <FormSelect v-model="form.hub_id" label="Bưu cục trực thuộc" required hint="Chọn bưu cục bạn sẽ làm việc.">
              <option value="">-- Chọn bưu cục --</option>
              <option v-for="h in hubs" :key="h.id" :value="h.id">{{ h.name }}</option>
            </FormSelect>

            <div v-if="error" class="flex items-start gap-2 px-3 py-2.5 rounded-md text-sm bg-danger/10 text-danger animate-fade-in">
              <AlertCircle class="w-4 h-4 mt-0.5 shrink-0" />
              <span>{{ error }}</span>
            </div>

            <BaseButton type="submit" variant="primary" size="md" :loading="loading" class="w-full">
              Đăng ký
            </BaseButton>
          </form>

          <p class="text-center text-meta text-xs mt-6">
            Đã có tài khoản?
            <router-link :to="{ name: 'Login' }" class="text-[var(--primary)] font-medium hover:underline">Đăng nhập</router-link>
          </p>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import api from '../../services/api';
import { useToastStore } from '../../stores/toastStore';
import { User, Phone, Lock, AlertCircle, Truck, CircleCheck } from 'lucide-vue-next';
import FormField from '../../components/ui/FormField.vue';
import FormSelect from '../../components/ui/FormSelect.vue';
import BaseButton from '../../components/ui/BaseButton.vue';

const toast = useToastStore();

const hubs = ref([]);
const form = ref({ name: '', phone: '', email: '', password: '', role: 'first_mile_driver', hub_id: '' });
const error = ref('');
const loading = ref(false);
const done = ref(false);

// Danh sách bưu cục để tài xế tự chọn (GET /hubs công khai — không cần auth).
const fetchHubs = async () => {
  try {
    const res = await api.get('/hubs');
    if (res.success) hubs.value = res.data || [];
  } catch (err) {
    console.error(err);
  }
};

onMounted(fetchHubs);

const handleRegister = async () => {
  error.value = '';
  if (!form.value.name || !form.value.phone || !form.value.password) {
    error.value = 'Vui lòng điền đầy đủ họ tên, số điện thoại và mật khẩu';
    return;
  }
  if (!form.value.hub_id) {
    error.value = 'Vui lòng chọn bưu cục trực thuộc';
    return;
  }
  loading.value = true;
  try {
    const res = await api.post('/employees/register', { ...form.value });
    if (res.success) {
      done.value = true;
      toast.success('Đăng ký thành công, chờ Admin duyệt');
    }
  } catch (err) {
    error.value = err.response?.data?.error?.message || err.response?.data?.error || 'Đăng ký thất bại';
  } finally {
    loading.value = false;
  }
};
</script>
