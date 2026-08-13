<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center gap-4 flex-wrap justify-between w-full">
      <div class="flex items-center gap-4 flex-wrap">
        <router-link to="/orders">
          <BaseButton variant="secondary" size="sm"><ArrowLeft class="w-4 h-4" /> Quay lại</BaseButton>
        </router-link>
        <div class="flex items-center gap-3">
          <h1 class="text-2xl font-semibold text-strong">
            <span class="font-mono">{{ order?.tracking_number || '—' }}</span>
          </h1>
          <StatusBadge v-if="order" :status="order.status" />
          <div v-if="order && isSlaBreached(order)" class="bg-red-100 text-red-600 px-2 py-1 rounded text-xs font-bold flex items-center gap-1 border border-red-200">
            <AlertCircle class="w-3.5 h-3.5" /> QUÁ HẠN SLA
          </div>
        </div>
      </div>
      <BaseButton v-if="order" variant="secondary" size="sm" @click="printLabel">
        <Printer class="w-4 h-4" /> In vận đơn
      </BaseButton>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-1 space-y-6">
        <div class="skeleton h-48 w-full rounded-[var(--r-lg)]"></div>
        <div class="skeleton h-48 w-full rounded-[var(--r-lg)]"></div>
      </div>
      <div class="lg:col-span-2 space-y-6">
        <div class="skeleton h-72 w-full rounded-[var(--r-lg)]"></div>
        <div class="skeleton h-64 w-full rounded-[var(--r-lg)]"></div>
      </div>
    </div>

    <div v-else-if="order" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Cột thông tin -->
      <div class="lg:col-span-1 space-y-6">
        <BaseCard title="Người nhận">
          <div class="space-y-3 text-sm">
            <div class="flex items-start gap-3">
              <User class="w-4 h-4 text-meta mt-0.5 shrink-0" />
              <div>
                <div class="font-medium text-strong">{{ order.receiver_name }}</div>
                <div class="text-meta">{{ order.receiver_phone }}</div>
              </div>
            </div>
            <div class="flex items-start gap-3">
              <MapPin class="w-4 h-4 text-meta mt-0.5 shrink-0" />
              <div class="text-body">
                {{ order.receiver_address_detail }}
                <span v-if="order.receiver_location_id" class="text-meta block text-xs mt-0.5">Mã KV: {{ order.receiver_location_id }}</span>
              </div>
            </div>
          </div>
        </BaseCard>

        <BaseCard title="Cước phí & Hàng hóa">
          <div class="space-y-3 text-sm">
            <div class="flex justify-between">
              <span class="text-meta flex items-center gap-2"><Weight class="w-4 h-4" /> Khối lượng</span>
              <span class="font-medium text-strong tabular">{{ order.weight }}g</span>
            </div>
            <div class="flex justify-between">
              <span class="text-meta flex items-center gap-2"><Wallet class="w-4 h-4" /> Thu hộ (COD)</span>
              <span class="font-bold text-[var(--primary)] tabular">{{ formatMoney(order.cod_amount) }}đ</span>
            </div>
            <div class="flex justify-between">
              <span class="text-meta flex items-center gap-2"><Receipt class="w-4 h-4" /> Phí ship</span>
              <span class="font-medium text-strong tabular">{{ formatMoney(order.shipping_fee) }}đ</span>
            </div>
            <div class="flex justify-between pt-3 border-t">
              <span class="text-meta flex items-center gap-2"><Clock class="w-4 h-4" /> Tạo lúc</span>
              <span class="text-body tabular">{{ formatDate(order.created_at) }}</span>
            </div>
          </div>
        </BaseCard>

        <BaseCard v-if="order.delivery_attempts > 0" title="Giao hàng thất bại" class="border-orange-200 bg-orange-50/50">
          <div class="space-y-3 text-sm">
            <div class="flex justify-between">
              <span class="text-meta">Số lần thử:</span>
              <span class="font-bold text-orange-600">{{ order.delivery_attempts }} lần</span>
            </div>
            <div class="flex justify-between flex-col gap-1" v-if="order.failure_reason">
              <span class="text-meta">Lý do gần nhất:</span>
              <span class="text-strong bg-white p-2 border border-orange-100 rounded">{{ order.failure_reason }}</span>
            </div>
          </div>
        </BaseCard>
      </div>

      <!-- Cột bản đồ + timeline -->
      <div class="lg:col-span-2 space-y-6">
        <BaseCard title="Hành trình GPS">
          <template #actions>
            <span class="text-xs text-meta">{{ logsWithGps.length }} điểm ghi nhận</span>
          </template>
          
          <MapTracking :logs="logs" :order="order" />
          
          <p v-if="logsWithGps.length === 0" class="text-xs text-meta mt-3 text-center italic">
            Chưa có tọa độ GPS nào được ghi nhận cho đơn hàng này.
          </p>
        </BaseCard>

        <BaseCard title="Chi tiết hành trình">
          <div v-if="sortedLogs.length === 0" class="py-8 text-center text-meta text-sm">
            Chưa có dữ liệu hành trình.
          </div>
          <ol v-else class="relative border-l-2 border-[var(--border)] ml-2 space-y-6">
            <li v-for="(log, idx) in sortedLogs" :key="log.id" class="relative pl-6">
              <span
                class="absolute -left-[9px] top-1 w-4 h-4 rounded-full border-2 border-[var(--bg-surface)]"
                :style="{ background: statusColor(log.status), boxShadow: idx === 0 ? `0 0 0 4px ${statusColor(log.status)}33` : 'none' }"
              ></span>
              <div class="flex items-start justify-between gap-3 flex-wrap">
                <div>
                  <StatusBadge :status="log.status" />
                  <p v-if="log.note" class="text-sm text-body mt-2">{{ log.note }}</p>
                  <div v-if="log.employee_id" class="flex items-center gap-1.5 text-xs text-meta mt-2">
                    <UserCircle class="w-3.5 h-3.5" /> NV: <span class="font-medium">{{ log.employee_name || shortId(log.employee_id) }}</span>
                  </div>
                </div>
                <span class="text-xs text-meta tabular whitespace-nowrap">{{ formatDate(log.created_at) }}</span>
              </div>
            </li>
          </ol>
        </BaseCard>
      </div>
    </div>

    <div v-else class="py-20 text-center">
      <PackageX class="w-12 h-12 mx-auto text-meta/40 mb-3" />
      <p class="text-meta">Không tìm thấy vận đơn.</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRoute } from 'vue-router';
import api from '../services/api';
import { useToastStore } from '../stores/toastStore';
import { useAuthStore } from '../stores/authStore';
import { statusColor } from '../composables/useStatus';
import {
  ArrowLeft, User, MapPin, Weight, Wallet, Receipt, Clock, UserCircle, PackageX, Printer, AlertCircle
} from 'lucide-vue-next';
import BaseCard from '../components/ui/BaseCard.vue';
import BaseButton from '../components/ui/BaseButton.vue';
import StatusBadge from '../components/ui/StatusBadge.vue';
import MapTracking from '../components/MapTracking.vue';

const route = useRoute();
const toast = useToastStore();
const authStore = useAuthStore();
const orderId = route.params.id;

const loading = ref(true);
const order = ref(null);
const logs = ref([]);

const fetchData = async () => {
  loading.value = true;
  try {
    const res = await api.get(`/orders/${orderId}`);
    if (res.success) {
      order.value = res.data.order;
      logs.value = res.data.logs || [];
    }
  } catch (error) {
    toast.error('Không thể tải chi tiết đơn hàng');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

onMounted(fetchData);

const sortedLogs = computed(() =>
  [...logs.value].reverse()
);

const printLabel = async () => {
  try {
    const res = await api.get(`/orders/${orderId}/label`, { responseType: 'blob' });
    const url = window.URL.createObjectURL(res);
    const win = window.open(url, '_blank');
    win.onload = () => { window.URL.revokeObjectURL(url); };
  } catch (err) {
    toast.error('Lỗi khi lấy tem vận đơn');
    console.error(err);
  }
};
const logsWithGps = computed(() => {
  let count = logs.value.filter((l) => l.latitude && l.longitude).length;
  if (order.value && order.value.sender_latitude && order.value.sender_longitude) count++;
  if (order.value && order.value.receiver_latitude && order.value.receiver_longitude) count++;
  // We just return an array of this length so the template gets the right length
  return new Array(count);
});
const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
const formatDate = (s) => {
  if (!s) return '';
  return new Date(s).toLocaleString('vi-VN', { hour: '2-digit', minute: '2-digit', day: '2-digit', month: '2-digit', year: 'numeric' });
};
const shortId = (id) => (id ? id.slice(0, 8) : '');

const isSlaBreached = (o) => {
  if (!o.sla_deadline) return false;
  if (['delivered', 'returned'].includes(o.status)) return false;
  return new Date(o.sla_deadline) < new Date();
};
</script>

<style scoped>
/* Order detail specific styles if needed */
</style>
