<template>
  <div class="min-h-screen bg-subtle flex flex-col items-center py-12 px-4 sm:px-6 lg:px-8">
    <div class="w-full max-w-md bg-surface rounded-[var(--r-lg)] shadow-e2 p-8 mb-6 text-center">
      <div class="w-12 h-12 bg-primary/10 text-primary rounded-full flex items-center justify-center mx-auto mb-4">
        <Package class="w-6 h-6" />
      </div>
      <h1 class="text-2xl font-bold text-strong mb-2">Tra cứu vận đơn</h1>
      <p class="text-meta text-sm mb-6">Nhập mã vận đơn của bạn để xem hành trình đơn hàng</p>

      <form @submit.prevent="searchTracking" class="space-y-4">
        <div>
          <input
            v-model="trackingNumber"
            type="text"
            required
            class="w-full px-4 py-3 bg-subtle border rounded-[var(--r-md)] text-sm text-strong outline-none focus:border-[var(--primary)] focus:ring-2 focus:ring-[var(--primary-ring)]/40 transition-shadow uppercase"
            placeholder="VD: PKG-123456789"
          />
        </div>
        <button
          type="submit"
          :disabled="loading || !trackingNumber"
          class="w-full h-11 bg-primary text-white rounded-[var(--r-md)] text-sm font-semibold flex items-center justify-center gap-2 hover:bg-primary-hover disabled:opacity-50 transition-colors"
        >
          <Search v-if="!loading" class="w-4 h-4" />
          <Loader2 v-else class="w-4 h-4 animate-spin" />
          Tra cứu
        </button>
      </form>
    </div>

    <!-- Kết quả -->
    <div v-if="order" class="w-full max-w-2xl bg-surface rounded-[var(--r-lg)] shadow-e2 p-6">
      <div class="flex items-center justify-between border-b pb-4 mb-4">
        <div>
          <h2 class="text-lg font-bold text-strong">{{ order.tracking_number }}</h2>
          <p class="text-sm text-meta mt-1">Người nhận: {{ maskName(order.receiver_name) }} - {{ maskPhone(order.receiver_phone) }}</p>
        </div>
        <StatusBadge :status="order.status" />
      </div>

      <div class="relative pl-6 space-y-6">
        <div class="absolute top-0 bottom-0 left-[11px] w-0.5 bg-border"></div>
        <div v-for="(log, idx) in sortedLogs" :key="log.id" class="relative">
          <!-- Timeline dot -->
          <div class="absolute -left-[30px] w-3.5 h-3.5 rounded-full border-2 border-surface bg-[var(--primary)]" :class="idx !== 0 ? 'bg-meta border-border' : ''"></div>
          
          <div class="bg-subtle p-3 rounded-[var(--r-md)] border text-sm">
            <div class="font-semibold text-strong mb-1">{{ statusConfig(log.status).label }}</div>
            <div v-if="log.note" class="text-meta text-xs mb-1.5">{{ log.note }}</div>
            <div class="flex items-center gap-2 text-xs text-meta">
              <Clock class="w-3.5 h-3.5" /> {{ formatDate(log.created_at) }}
            </div>
          </div>
        </div>
        <div v-if="sortedLogs.length === 0" class="text-sm text-meta">
          Chưa có nhật ký hành trình.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import api from '../services/api';
import { useToastStore } from '../stores/toastStore';
import { Package, Search, Loader2, Clock } from 'lucide-vue-next';
import { statusConfig } from '../composables/useStatus';
import StatusBadge from '../components/ui/StatusBadge.vue';

const route = useRoute();
const router = useRouter();
const toast = useToastStore();

const trackingNumber = ref(route.query.code || '');
const loading = ref(false);
const order = ref(null);

const maskName = (name) => {
  if (!name) return '';
  const parts = name.split(' ');
  if (parts.length === 1) return name[0] + '***';
  return parts[0] + ' ' + parts[parts.length - 1][0] + '***';
};

const maskPhone = (phone) => {
  if (!phone || phone.length < 10) return phone;
  return phone.substring(0, 3) + '****' + phone.substring(phone.length - 3);
};

const formatDate = (dateString) => {
  if (!dateString) return '';
  const d = new Date(dateString);
  if (isNaN(d.getTime())) return dateString;
  const h = d.getHours().toString().padStart(2, '0');
  const m = d.getMinutes().toString().padStart(2, '0');
  const D = d.getDate().toString().padStart(2, '0');
  const M = (d.getMonth() + 1).toString().padStart(2, '0');
  const Y = d.getFullYear();
  return `${h}:${m} ${D}/${M}/${Y}`;
};

const sortedLogs = computed(() => {
  if (!order.value?.tracking_logs) return [];
  return [...order.value.tracking_logs].sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
});

const searchTracking = async () => {
  if (!trackingNumber.value) return;
  
  // Cập nhật URL
  router.replace({ query: { code: trackingNumber.value } });
  
  loading.value = true;
  order.value = null;
  
  try {
    const res = await api.get(`/public/tracking/${trackingNumber.value}`);
    if (res.success) {
      order.value = res.data;
    }
  } catch (error) {
    if (error.response?.status === 404) {
      toast.error('Không tìm thấy vận đơn ' + trackingNumber.value);
    } else {
      toast.error('Có lỗi xảy ra khi tra cứu');
    }
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  if (trackingNumber.value) {
    searchTracking();
  }
});
</script>
