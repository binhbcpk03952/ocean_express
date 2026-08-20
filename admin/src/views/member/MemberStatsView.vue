<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-strong">Báo cáo hiệu suất cá nhân</h1>
    </div>

    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="i in 4" :key="i" class="skeleton h-32 rounded-[var(--r-lg)]"></div>
    </div>

    <div v-else-if="stats" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <template v-if="isHubStaff">
        <BaseCard class="p-4 flex flex-col justify-center items-center text-center">
          <ScanLine class="w-8 h-8 text-[var(--primary)] mb-3" />
          <div class="text-3xl font-bold text-strong">{{ stats.total_scans || 0 }}</div>
          <div class="text-meta mt-1">Lượt quét mã (In/Out)</div>
        </BaseCard>
      </template>

      <template v-else>
        <BaseCard class="p-4 flex flex-col justify-center items-center text-center border-l-4 border-l-[var(--success)]">
          <CheckCircle class="w-8 h-8 text-[var(--success)] mb-3" />
          <div class="text-3xl font-bold text-strong">{{ stats.total_delivered || 0 }}</div>
          <div class="text-meta mt-1">Đơn giao thành công</div>
        </BaseCard>

        <BaseCard class="p-4 flex flex-col justify-center items-center text-center border-l-4 border-l-[var(--danger)]">
          <XCircle class="w-8 h-8 text-[var(--danger)] mb-3" />
          <div class="text-3xl font-bold text-strong">{{ stats.total_failed || 0 }}</div>
          <div class="text-meta mt-1">Đơn hoàn/thất bại</div>
        </BaseCard>

        <BaseCard class="p-4 flex flex-col justify-center items-center text-center border-l-4 border-l-[var(--warning)]">
          <Wallet class="w-8 h-8 text-[var(--warning)] mb-3" />
          <div class="text-3xl font-bold text-strong">{{ formatMoney(stats.total_cod_holding) }}</div>
          <div class="text-meta mt-1">Tiền mặt COD đang giữ</div>
        </BaseCard>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useToastStore } from '../../stores/toastStore';
import { useAuthStore } from '../../stores/authStore';
import api from '../../services/api';
import BaseCard from '../../components/ui/BaseCard.vue';
import { ScanLine, CheckCircle, XCircle, Wallet } from 'lucide-vue-next';

const toast = useToastStore();
const authStore = useAuthStore();

const loading = ref(true);
const stats = ref(null);

const isHubStaff = computed(() => authStore.user?.role === 'hub_staff');

const formatMoney = (val) => new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(val || 0);

const fetchStats = async () => {
  loading.value = true;
  try {
    const res = await api.get('/stats/member/me');
    if (res.success) {
      stats.value = res.data;
    }
  } catch (error) {
    toast.error('Lỗi khi lấy dữ liệu thống kê');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  fetchStats();
});
</script>
