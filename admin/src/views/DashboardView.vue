<template>
  <div class="space-y-6">
    <!-- Page header -->
    <div class="flex items-end justify-between flex-wrap gap-4">
      <div>
        <h1 class="text-2xl font-semibold text-strong">Tổng quan vận hành</h1>
        <p class="text-meta text-sm mt-1">Giám sát dòng chảy hàng hóa toàn hệ thống</p>
      </div>
      <BaseButton variant="secondary" size="md" :loading="loading" @click="fetchStats">
        <RefreshCw class="w-4 h-4" /> Làm mới
      </BaseButton>
    </div>

    <!-- Stat Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-5">
      <StatCard
        label="Tổng vận đơn"
        :value="loading ? null : formatNumber(stats.total_orders)"
        :icon="Package"
        tone="info"
      />
      <StatCard
        label="Đang xử lý"
        :value="loading ? null : formatNumber(stats.in_progress_count)"
        :icon="Truck"
        tone="warning"
      />
      <StatCard
        label="Giao thành công"
        :value="loading ? null : formatNumber(stats.delivered_count)"
        :icon="CheckCircle2"
        tone="success"
      />
      <StatCard
        label="Tỷ lệ thành công"
        :value="loading ? null : stats.success_rate.toFixed(1) + '%'"
        :icon="TrendingUp"
        tone="success"
      />
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-5 gap-6">
      <!-- Phân bổ theo trạng thái -->
      <BaseCard class="lg:col-span-3" title="Phân bổ theo trạng thái">
        <div v-if="loading" class="space-y-4 py-1">
          <div v-for="i in 6" :key="i" class="skeleton h-8 w-full"></div>
        </div>
        <div v-else-if="statusRows.length === 0" class="py-10 text-center text-meta text-sm">
          Chưa có vận đơn nào trong hệ thống.
        </div>
        <div v-else class="h-[300px] flex items-center justify-center p-4">
          <Doughnut :data="chartData" :options="chartOptions" />
        </div>
      </BaseCard>

      <!-- Bưu cục ùn tắc -->
      <BaseCard class="lg:col-span-2" title="Bưu cục ùn tắc">
        <template #actions>
          <span class="text-xs text-meta">Top {{ stats.congested_hubs.length }}</span>
        </template>
        <div v-if="loading" class="space-y-3 py-1">
          <div v-for="i in 4" :key="i" class="skeleton h-10 w-full"></div>
        </div>
        <div v-else-if="stats.congested_hubs.length === 0" class="py-10 text-center">
          <PackageCheck class="w-10 h-10 mx-auto text-meta/50 mb-2" />
          <p class="text-meta text-sm">Không có bưu cục nào đang giữ đơn.</p>
        </div>
        <ul v-else class="space-y-2">
          <li
            v-for="(hub, idx) in stats.congested_hubs"
            :key="hub.hub_id"
            class="flex items-center justify-between p-3 rounded-[var(--r-md)] bg-subtle"
          >
            <div class="flex items-center gap-3 min-w-0">
              <span
                class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0"
                :class="idx === 0 ? 'bg-danger text-white' : 'bg-[var(--border)] text-body'"
              >{{ idx + 1 }}</span>
              <span class="text-body font-medium truncate">{{ hub.hub_name }}</span>
            </div>
            <span class="px-2.5 py-1 rounded-full text-xs font-semibold tabular shrink-0"
              :style="{ color: 'var(--st-returned-fg)', background: 'var(--st-returned-bg)' }">
              {{ formatNumber(hub.count) }} đơn
            </span>
          </li>
        </ul>
      </BaseCard>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import api from '../services/api';
import { Package, Truck, CheckCircle2, TrendingUp, RefreshCw, PackageCheck } from 'lucide-vue-next';
import { STATUS_META, statusColor } from '../composables/useStatus';
import StatCard from '../components/ui/StatCard.vue';
import BaseCard from '../components/ui/BaseCard.vue';
import BaseButton from '../components/ui/BaseButton.vue';
import { useToastStore } from '../stores/toastStore';
import { Doughnut } from 'vue-chartjs';
import { Chart as ChartJS, Title, Tooltip, Legend, ArcElement, CategoryScale } from 'chart.js';

ChartJS.register(Title, Tooltip, Legend, ArcElement, CategoryScale);

const toast = useToastStore();
const loading = ref(true);
const stats = ref({
  total_orders: 0,
  status_counts: {},
  delivered_count: 0,
  returned_count: 0,
  in_progress_count: 0,
  success_rate: 0,
  congested_hubs: [],
});

const fetchStats = async () => {
  loading.value = true;
  try {
    const res = await api.get('/stats/dashboard');
    if (res.success) {
      stats.value = {
        ...stats.value,
        ...res.data,
        status_counts: res.data.status_counts || {},
        congested_hubs: res.data.congested_hubs || [],
      };
    }
  } catch (error) {
    toast.error('Không thể tải số liệu dashboard');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

onMounted(fetchStats);

const statusRows = computed(() =>
  Object.entries(stats.value.status_counts)
    .map(([status, count]) => ({ status, count }))
    .sort((a, b) => b.count - a.count)
);

const chartData = computed(() => {
  return {
    labels: statusRows.value.map(r => formatStatus(r.status)),
    datasets: [{
      data: statusRows.value.map(r => r.count),
      backgroundColor: statusRows.value.map(r => statusColor(r.status)),
    }]
  }
});

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'right' }
  }
};

const formatNumber = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
const formatStatus = (status) => STATUS_META[status]?.label || status;
</script>
