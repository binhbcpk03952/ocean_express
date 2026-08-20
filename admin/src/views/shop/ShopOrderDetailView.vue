<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center gap-4 flex-wrap justify-between w-full">
      <div class="flex items-center gap-4 flex-wrap">
        <router-link :to="{ name: 'ShopOrders' }">
          <BaseButton variant="secondary" size="sm">
            <ArrowLeft class="w-4 h-4" /> Quay lại
          </BaseButton>
        </router-link>
        <div class="flex items-center gap-3">
          <h1 class="text-2xl font-bold text-strong">
            <span class="font-mono">{{ order?.tracking_number || '—' }}</span>
          </h1>
          <StatusBadge v-if="order" :status="order.status" />
          <div
            v-if="order && isSlaBreached(order)"
            class="bg-red-100 text-red-600 px-2 py-1 rounded text-xs font-bold flex items-center gap-1 border border-red-200"
          >
            <AlertCircle class="w-3.5 h-3.5" /> QUÁ HẠN SLA
          </div>
        </div>
      </div>

      <div class="flex items-center gap-2" v-if="order">
        <BaseButton variant="secondary" size="sm" @click="copyCustomerLink">
          <Share2 class="w-4 h-4 text-primary" /> Link gửi khách
        </BaseButton>
        <BaseButton variant="primary" size="sm" @click="openPrintModal">
          <Printer class="w-4 h-4" /> In vận đơn
        </BaseButton>
      </div>
    </div>

    <!-- Loading Skeleton -->
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

    <!-- Order Data -->
    <div v-else-if="order" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      
      <!-- Cột trái: Thông tin Người nhận, Hàng hóa, Cước phí (1/3) -->
      <div class="lg:col-span-1 space-y-6">
        
        <!-- Người nhận -->
        <BaseCard title="Thông tin người nhận">
          <div class="space-y-3 text-sm">
            <div class="flex items-start gap-3">
              <User class="w-4 h-4 text-meta mt-0.5 shrink-0" />
              <div>
                <div class="font-bold text-strong text-base">{{ order.receiver_name }}</div>
                <a :href="`tel:${order.receiver_phone}`" class="text-primary font-mono text-xs font-semibold">{{ order.receiver_phone }}</a>
              </div>
            </div>
            <div class="flex items-start gap-3">
              <MapPin class="w-4 h-4 text-meta mt-0.5 shrink-0" />
              <div class="text-body text-xs">
                {{ order.receiver_address_detail }}
                <span v-if="order.receiver_location_id" class="text-meta block text-[11px] font-mono mt-0.5">Mã KV: {{ order.receiver_location_id }}</span>
              </div>
            </div>
          </div>
        </BaseCard>

        <!-- Hàng hóa & Cước phí -->
        <BaseCard title="Hàng hóa & Cước phí">
          <div class="space-y-3 text-xs">
            <div class="flex justify-between">
              <span class="text-meta flex items-center gap-1.5"><Weight class="w-4 h-4" /> Khối lượng thực</span>
              <span class="font-bold text-strong tabular font-mono">{{ order.weight || 500 }}g</span>
            </div>

            <div v-if="order.length && order.width && order.height" class="flex justify-between">
              <span class="text-meta flex items-center gap-1.5"><Box class="w-4 h-4" /> Kích thước (DxRxC)</span>
              <span class="font-medium text-strong font-mono">{{ order.length }} x {{ order.width }} x {{ order.height }} cm</span>
            </div>

            <div class="flex justify-between">
              <span class="text-meta flex items-center gap-1.5"><Wallet class="w-4 h-4" /> Thu hộ (COD)</span>
              <span class="font-black text-sm text-red-600 tabular font-mono">{{ formatMoney(order.cod_amount) }}đ</span>
            </div>

            <div class="flex justify-between">
              <span class="text-meta flex items-center gap-1.5"><Receipt class="w-4 h-4" /> Phí giao hàng</span>
              <span class="font-semibold text-strong tabular font-mono">{{ formatMoney(order.shipping_fee) }}đ</span>
            </div>

            <div class="flex justify-between pt-3 border-t">
              <span class="text-meta flex items-center gap-1.5"><Clock class="w-4 h-4" /> Ngày tạo</span>
              <span class="text-body tabular">{{ formatDate(order.created_at) }}</span>
            </div>

            <div v-if="order.estimated_delivery_time" class="flex justify-between">
              <span class="text-meta flex items-center gap-1.5"><Calendar class="w-4 h-4" /> Dự kiến giao</span>
              <span class="font-semibold text-primary tabular">{{ formatDate(order.estimated_delivery_time) }}</span>
            </div>
          </div>
        </BaseCard>

        <!-- Lịch sử giao thất bại (nếu có) -->
        <BaseCard v-if="order.delivery_attempts > 0" title="Giao hàng thất bại" class="border-orange-200 bg-orange-50/40">
          <div class="space-y-2 text-xs">
            <div class="flex justify-between">
              <span class="text-meta">Số lần thử phát:</span>
              <span class="font-bold text-orange-600">{{ order.delivery_attempts }} lần</span>
            </div>
            <div v-if="order.failure_reason" class="space-y-1">
              <span class="text-meta">Lý do gần nhất:</span>
              <div class="bg-white p-2 rounded border border-orange-200 text-strong">{{ order.failure_reason }}</div>
            </div>
          </div>
        </BaseCard>

      </div>

      <!-- Cột phải: Map Tracking + Timeline (2/3) -->
      <div class="lg:col-span-2 space-y-6">
        
        <!-- Bản đồ GPS Map Tracking -->
        <BaseCard title="Hành trình GPS (Bản đồ)">
          <template #actions>
            <span class="text-xs text-meta">{{ logsWithGps.length }} điểm ghi nhận</span>
          </template>
          
          <MapTracking :logs="logs" :order="order" />
          
          <p v-if="logsWithGps.length === 0" class="text-xs text-meta mt-3 text-center italic">
            Chưa có tọa độ GPS nào được ghi nhận cho đơn hàng này.
          </p>
        </BaseCard>

        <!-- Chi tiết hành trình Timeline -->
        <BaseCard title="Nhật ký hành trình">
          <div v-if="sortedLogs.length === 0" class="py-8 text-center text-meta text-xs">
            Chưa có dữ liệu hành trình.
          </div>
          <ol v-else class="relative border-l-2 border-[var(--border)] ml-2 space-y-6">
            <li v-for="(log, idx) in sortedLogs" :key="log.id || idx" class="relative pl-6">
              <span
                class="absolute -left-[9px] top-1 w-4 h-4 rounded-full border-2 border-[var(--bg-surface)]"
                :style="{ background: statusColor(log.status), boxShadow: idx === 0 ? `0 0 0 4px ${statusColor(log.status)}33` : 'none' }"
              ></span>
              <div class="flex items-start justify-between gap-3 flex-wrap">
                <div>
                  <StatusBadge :status="log.status" />
                  <p v-if="log.note" class="text-xs text-body mt-2">{{ log.note }}</p>
                  <div v-if="log.employee_name" class="text-[11px] text-meta mt-1 font-medium">
                    Nhân viên phụ trách: {{ log.employee_name }}
                  </div>
                </div>
                <span class="text-xs text-meta tabular font-mono whitespace-nowrap">{{ formatDate(log.created_at) }}</span>
              </div>
            </li>
          </ol>
        </BaseCard>

      </div>

    </div>

    <!-- Not Found -->
    <div v-else class="py-20 text-center">
      <PackageX class="w-12 h-12 mx-auto text-meta/40 mb-3" />
      <p class="text-meta">Không tìm thấy thông tin vận đơn.</p>
    </div>

    <!-- Printable Label Modal -->
    <ShippingLabelModal
      v-model="showPrintModal"
      :order="order"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRoute } from 'vue-router';
import api from '../../services/api';
import { useToastStore } from '../../stores/toastStore';
import { statusColor } from '../../composables/useStatus';
import {
  ArrowLeft, User, MapPin, Weight, Wallet, Receipt, Clock, Box, Calendar,
  PackageX, Printer, AlertCircle, Share2
} from 'lucide-vue-next';
import BaseCard from '../../components/ui/BaseCard.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import MapTracking from '../../components/MapTracking.vue';
import ShippingLabelModal from '../../components/ShippingLabelModal.vue';

const route = useRoute();
const toast = useToastStore();
const orderId = route.params.id;

const loading = ref(true);
const order = ref(null);
const logs = ref([]);
const showPrintModal = ref(false);

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

const logsWithGps = computed(() => {
  let count = logs.value.filter((l) => l.latitude && l.longitude).length;
  if (order.value && order.value.sender_latitude && order.value.sender_longitude) count++;
  if (order.value && order.value.receiver_latitude && order.value.receiver_longitude) count++;
  return new Array(count);
});

const openPrintModal = () => {
  showPrintModal.value = true;
};

const copyCustomerLink = () => {
  if (!order.value?.tracking_number) return;
  const url = window.location.origin + '/tracking?code=' + encodeURIComponent(order.value.tracking_number);
  navigator.clipboard.writeText(url);
  toast.success('Đã sao chép link tra cứu đơn hàng gửi khách');
};

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);

const formatDate = (s) => {
  if (!s) return '';
  return new Date(s).toLocaleString('vi-VN', { hour: '2-digit', minute: '2-digit', day: '2-digit', month: '2-digit', year: 'numeric' });
};

const isSlaBreached = (o) => {
  if (!o.sla_deadline) return false;
  if (['delivered', 'returned'].includes(o.status)) return false;
  return new Date(o.sla_deadline) < new Date();
};
</script>
