<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-end justify-between flex-wrap gap-4">
      <div>
        <h1 class="text-2xl font-semibold text-strong">Tổng quan</h1>
        <p class="text-meta text-sm mt-1">Tình hình vận đơn của {{ authStore.user?.name || 'shop' }}</p>
      </div>
      <router-link :to="{ name: 'ShopCreateOrder' }">
        <BaseButton variant="primary" size="md"><PackagePlus class="w-4 h-4" /> Tạo đơn mới</BaseButton>
      </router-link>
    </div>

    <!-- Stat tiles -->
    <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
      <div class="rounded-[var(--r-lg)] border bg-surface p-5 shadow-e1">
        <div class="flex items-center justify-between mb-3">
          <div class="text-xs text-meta font-medium uppercase tracking-wide">Tổng đơn</div>
          <Package class="w-4 h-4 text-meta/50" />
        </div>
        <div class="text-3xl font-bold text-strong tabular">{{ loading ? '—' : dashboardStats?.total_orders || 0 }}</div>
      </div>
      <div v-for="tile in tiles" :key="tile.key" class="rounded-[var(--r-lg)] border bg-surface p-5 shadow-e1">
        <div class="flex items-center justify-between mb-3">
          <div class="text-xs text-meta font-medium uppercase tracking-wide">{{ tile.label }}</div>
          <component :is="tile.icon" class="w-4 h-4 text-meta/50" />
        </div>
        <div class="text-3xl font-bold tabular" :style="{ color: tile.fg }">{{ loading ? '—' : tile.count }}</div>
      </div>
    </div>

    <!-- Main content grid -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Đơn gần đây (2/3) -->
      <div class="lg:col-span-2">
        <BaseCard title="Đơn gần đây" body-class="">
          <template #actions>
            <router-link :to="{ name: 'ShopOrders' }" class="text-xs text-[var(--primary)] hover:underline">Xem tất cả</router-link>
          </template>
          <div v-if="loading" class="divide-y">
            <div v-for="i in 5" :key="i" class="flex items-center justify-between px-5 py-4 gap-3">
              <div class="space-y-1.5 flex-1">
                <div class="skeleton h-3.5 w-40 rounded"></div>
                <div class="skeleton h-3 w-56 rounded"></div>
              </div>
              <div class="skeleton h-5 w-16 rounded-full"></div>
            </div>
          </div>
          <div v-else-if="recentOrders.length === 0" class="py-16 text-center">
            <Package class="w-10 h-10 mx-auto text-meta/40 mb-3" />
            <p class="text-meta text-sm">Chưa có đơn nào.</p>
            <router-link :to="{ name: 'ShopCreateOrder' }" class="inline-block mt-3 text-xs text-[var(--primary)] hover:underline">Tạo đơn đầu tiên →</router-link>
          </div>
          <div v-else class="divide-y">
            <router-link
              v-for="order in recentOrders"
              :key="order.id"
              :to="{ name: 'ShopOrders' }"
              class="flex items-center justify-between gap-3 px-5 py-4 transition-colors hover:bg-subtle"
            >
              <div class="min-w-0 flex-1">
                <div class="font-mono text-[13px] font-semibold text-strong">{{ order.tracking_number }}</div>
                <div class="text-xs text-meta mt-0.5 truncate">{{ order.receiver_name }} · {{ order.receiver_address_detail }}</div>
              </div>
              <StatusBadge :status="order.status" />
            </router-link>
          </div>
        </BaseCard>
      </div>

      <!-- Quick actions & Chart (1/3) -->
      <div class="space-y-4 lg:col-span-1">
        <BaseCard title="Phân bổ vận đơn" class="h-64">
          <div v-if="loading" class="flex items-center justify-center h-full">
            <div class="skeleton w-32 h-32 rounded-full"></div>
          </div>
          <div v-else-if="!dashboardStats?.total_orders" class="flex items-center justify-center h-full text-sm text-meta">
            Chưa có vận đơn
          </div>
          <div v-else class="h-full pb-4">
             <Doughnut :data="chartData" :options="chartOptions" />
          </div>
        </BaseCard>

        <BaseCard title="Thao tác nhanh">
          <div class="space-y-2">
            <router-link :to="{ name: 'ShopCreateOrder' }" class="flex items-center gap-3 p-3 rounded-[var(--r-md)] border hover:bg-subtle transition-colors">
              <div class="w-9 h-9 rounded-[var(--r-md)] flex items-center justify-center shrink-0" style="background: var(--primary-soft)">
                <PackagePlus class="w-4 h-4" style="color: var(--primary)" />
              </div>
              <div class="min-w-0">
                <div class="text-sm font-medium text-strong">Tạo vận đơn mới</div>
                <div class="text-xs text-meta">Gửi hàng ngay</div>
              </div>
            </router-link>
            <router-link :to="{ name: 'ShopOrders' }" class="flex items-center gap-3 p-3 rounded-[var(--r-md)] border hover:bg-subtle transition-colors">
              <div class="w-9 h-9 rounded-[var(--r-md)] flex items-center justify-center shrink-0 bg-subtle">
                <ClipboardList class="w-4 h-4 text-meta" />
              </div>
              <div class="min-w-0">
                <div class="text-sm font-medium text-strong">Danh sách đơn</div>
                <div class="text-xs text-meta">Tra cứu & theo dõi</div>
              </div>
            </router-link>
            <router-link :to="{ name: 'ShopWallet' }" class="flex items-center gap-3 p-3 rounded-[var(--r-md)] border hover:bg-subtle transition-colors">
              <div class="w-9 h-9 rounded-[var(--r-md)] flex items-center justify-center shrink-0 bg-subtle">
                <Wallet class="w-4 h-4 text-meta" />
              </div>
              <div class="min-w-0">
                <div class="text-sm font-medium text-strong">Ví & Đối soát</div>
                <div class="text-xs text-meta">Số dư & lịch sử</div>
              </div>
            </router-link>
            <router-link :to="{ name: 'ShopAccount' }" class="flex items-center gap-3 p-3 rounded-[var(--r-md)] border hover:bg-subtle transition-colors">
              <div class="w-9 h-9 rounded-[var(--r-md)] flex items-center justify-center shrink-0 bg-subtle">
                <UserCircle class="w-4 h-4 text-meta" />
              </div>
              <div class="min-w-0">
                <div class="text-sm font-medium text-strong">Tài khoản & API</div>
                <div class="text-xs text-meta">Hồ sơ & API Key</div>
              </div>
            </router-link>
          </div>
        </BaseCard>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import api from '../../services/api';
import { useAuthStore } from '../../stores/authStore';
import { useToastStore } from '../../stores/toastStore';
import { statusConfig } from '../../composables/useStatus';
import { PackagePlus, Package, ClipboardList, Wallet, UserCircle, Truck, CheckCircle2, RotateCcw } from 'lucide-vue-next';
import BaseCard from '../../components/ui/BaseCard.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import { Doughnut } from 'vue-chartjs';
import { Chart as ChartJS, Title, Tooltip, Legend, ArcElement, CategoryScale } from 'chart.js';
import { STATUS_META, statusColor } from '../../composables/useStatus';

ChartJS.register(Title, Tooltip, Legend, ArcElement, CategoryScale);

const authStore = useAuthStore();
const toast = useToastStore();

const dashboardStats = ref(null);
const recentOrders = ref([]);
const loading = ref(true);

const fetchStats = async () => {
  loading.value = true;
  try {
    const res = await api.get('/stats/shop/me');
    if (res.success) dashboardStats.value = res.data;
    
    // Fetch recent orders manually
    const ordersRes = await api.get('/orders', { params: { limit: 8, page: 1 } });
    if (ordersRes.success) {
       recentOrders.value = ordersRes.data || [];
    }
  } catch (error) {
    toast.error('Không thể tải tổng quan');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

onMounted(fetchStats);

const tiles = computed(() => {
  const counts = dashboardStats.value?.status_counts || {};
  const delivering = ['picked_up', 'hub_inbound', 'in_transit', 'hub_outbound', 'delivering'].reduce((sum, s) => sum + (counts[s] || 0), 0);
  return [
    { key: 'delivering', label: 'Đang xử lý', count: delivering, fg: statusConfig('delivering').fg, icon: Truck },
    { key: 'delivered', label: 'Đã giao', count: counts['delivered'] || 0, fg: statusConfig('delivered').fg, icon: CheckCircle2 },
    { key: 'returned', label: 'Hoàn', count: counts['returned'] || 0, fg: statusConfig('returned').fg, icon: RotateCcw },
  ];
});

const chartData = computed(() => {
  const counts = dashboardStats.value?.status_counts || {};
  const statusRows = Object.entries(counts)
    .map(([status, count]) => ({ status, count }))
    .sort((a, b) => b.count - a.count);
    
  return {
    labels: statusRows.map(r => STATUS_META[r.status]?.label || r.status),
    datasets: [{
      data: statusRows.map(r => r.count),
      backgroundColor: statusRows.map(r => statusColor(r.status)),
    }]
  }
});

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'bottom' }
  }
};
</script>