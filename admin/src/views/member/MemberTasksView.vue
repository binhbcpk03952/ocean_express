<template>
  <div class="space-y-4">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-lg font-semibold text-strong">Nhiệm vụ của tôi</h1>
        <p class="text-xs text-meta mt-0.5">{{ orders.length }} đơn đang xử lý</p>
      </div>
      <div class="flex items-center gap-2">
        <BaseButton
          v-if="['first_mile_driver', 'last_mile_driver'].includes(authStore.user?.role)"
          variant="outline" size="sm" @click="optimizeRoute" :loading="optimizing"
        >
          <Map class="w-4 h-4 mr-1" /> Tối ưu lộ trình
        </BaseButton>
        <BaseButton
          v-if="['first_mile_driver', 'last_mile_driver'].includes(authStore.user?.role)"
          variant="primary" size="sm" @click="submitCOD" :loading="submittingCod"
        >
          <Wallet class="w-4 h-4 mr-1" /> Nộp COD
        </BaseButton>
        <button
          class="rounded-md p-2 text-meta hover:bg-subtle transition-colors ml-1"
          :class="{ 'animate-spin': loading }"
          @click="fetchOrders"
          aria-label="Làm mới"
        >
          <RefreshCw class="h-5 w-5" />
        </button>
      </div>
    </div>

    <!-- Filter chips -->
    <div class="flex items-center gap-2 overflow-x-auto pb-1 -mx-1 px-1">
      <button
        @click="statusFilter = ''"
        class="px-3 h-8 rounded-full text-xs font-medium whitespace-nowrap transition-colors border shrink-0"
        :class="statusFilter === '' ? 'bg-primary text-white border-transparent' : 'bg-surface text-body border-[var(--border)]'"
      >Tất cả</button>
      <button
        v-for="s in availableStatuses"
        :key="s"
        @click="statusFilter = s"
        class="px-3 h-8 rounded-full text-xs font-medium whitespace-nowrap transition-all border shrink-0"
        :style="statusFilter === s ? { color: statusConfig(s).fg, background: statusConfig(s).bg, borderColor: statusConfig(s).fg } : {}"
        :class="statusFilter === s ? 'font-semibold' : 'bg-surface text-body border-[var(--border)]'"
      >{{ statusConfig(s).label }}</button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <div v-for="i in 4" :key="i" class="skeleton h-28 w-full rounded-[var(--r-lg)]"></div>
    </div>

    <!-- Empty -->
    <div v-else-if="filteredOrders.length === 0" class="py-16 text-center">
      <PackageCheck class="w-12 h-12 mx-auto text-meta/40 mb-3" />
      <p class="text-meta text-sm">{{ orders.length === 0 ? 'Chưa có nhiệm vụ nào.' : 'Không có đơn khớp bộ lọc.' }}</p>
    </div>

    <!-- Task cards -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 md:gap-4 lg:gap-5">
      <router-link
        v-for="order in filteredOrders"
        :key="order.id"
        :to="{ name: 'MemberOrderDetail', params: { id: order.id } }"
        class="block flex-col rounded-[var(--r-lg)] border bg-surface p-4 shadow-e1 transition-all hover:-translate-y-0.5 hover:shadow-e2 active:scale-[0.99] group"
      >
        <div class="flex items-start justify-between gap-3">
          <span class="font-mono text-sm font-semibold text-[var(--primary)] group-hover:underline">{{ order.tracking_number }}</span>
          <StatusBadge :status="order.status" />
        </div>
        <div class="mt-4 flex items-start gap-2.5 text-sm">
          <User class="w-4 h-4 text-meta mt-0.5 shrink-0" />
          <div class="min-w-0">
            <div class="font-medium text-strong truncate">{{ order.receiver_name }}</div>
            <div class="text-meta text-[13px]">{{ order.receiver_phone }}</div>
          </div>
        </div>
        <div class="mt-2.5 flex items-start gap-2.5 text-sm">
          <MapPin class="w-4 h-4 text-meta mt-0.5 shrink-0" />
          <div class="text-body text-[13px] leading-relaxed line-clamp-2">{{ order.receiver_address_detail }}</div>
        </div>
        <div class="mt-4 flex items-center justify-between border-t pt-3">
          <span class="text-xs text-meta uppercase font-medium tracking-wider">COD</span>
          <span class="font-bold text-[var(--primary)] tabular">{{ formatMoney(order.cod_amount) }}đ</span>
        </div>
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import api from '../../services/api';
import { useToastStore } from '../../stores/toastStore';
import { useAuthStore } from '../../stores/authStore';
import { statusConfig } from '../../composables/useStatus';
import { RefreshCw, User, MapPin, PackageCheck, Map, Wallet } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import BaseButton from '../../components/ui/BaseButton.vue';

const toast = useToastStore();
const authStore = useAuthStore();

const orders = ref([]);
const loading = ref(true);
const statusFilter = ref('');
const statusConfigObj = statusConfig;

const optimizing = ref(false);
const submittingCod = ref(false);

const fetchOrders = async () => {
  loading.value = true;
  try {
    const res = await api.get('/orders');
    if (res.success) orders.value = res.data || [];
  } catch (error) {
    toast.error('Không thể tải danh sách nhiệm vụ');
    console.error(error);
	} finally {
		loading.value = false;
	}
};

const optimizeRoute = async () => {
  if (orders.value.length < 2) {
    toast.info('Cần ít nhất 2 đơn để tối ưu');
    return;
  }
  optimizing.value = true;
  try {
    const locations = orders.value.map(o => ({
      id: o.id,
      lat: o.receiver_latitude || 10.762622,
      lng: o.receiver_longitude || 106.660172
    }));
    
    const res = await api.post('/routes/optimize', { locations });
    if (res.success) {
      const optimizedIds = res.data.map(l => l.id);
      orders.value.sort((a, b) => optimizedIds.indexOf(a.id) - optimizedIds.indexOf(b.id));
      toast.success('Đã tối ưu lộ trình');
    }
  } catch (err) {
    toast.error('Lỗi khi tối ưu lộ trình');
  } finally {
    optimizing.value = false;
  }
};

const submitCOD = async () => {
  submittingCod.value = true;
  try {
    const res = await api.post('/orders/submit-cod');
    if (res.success) {
      toast.success(`Nộp thành công ${new Intl.NumberFormat('vi-VN').format(res.data.amount_submitted)}đ`);
      fetchOrders();
    }
  } catch (err) {
    toast.error('Lỗi nộp COD');
  } finally {
    submittingCod.value = false;
  }
};

onMounted(fetchOrders);

// Chỉ hiện chip cho các trạng thái thực sự có trong danh sách.
const availableStatuses = computed(() => {
  const set = new Set(orders.value.map((o) => o.status));
  return [...set];
});

const filteredOrders = computed(() => {
  if (!statusFilter.value) return orders.value;
  return orders.value.filter((o) => o.status === statusFilter.value);
});

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
</script>
