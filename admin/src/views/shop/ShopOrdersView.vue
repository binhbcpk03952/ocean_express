<template>
  <div class="space-y-6">
    <div class="flex items-end justify-between flex-wrap gap-4">
      <div>
        <h1 class="text-2xl font-semibold text-strong">Vận đơn của tôi</h1>
        <p class="text-meta text-sm mt-1">Theo dõi trạng thái giao hàng</p>
      </div>
      <div class="flex items-center gap-2">
        <BaseButton variant="secondary" size="md" :loading="loading" @click="fetchOrders">
          <RefreshCw class="w-4 h-4" /> Làm mới
        </BaseButton>
        <router-link :to="{ name: 'ShopCreateOrder' }">
          <BaseButton variant="primary" size="md"><Plus class="w-4 h-4" /> Tạo đơn</BaseButton>
        </router-link>
      </div>
    </div>

    <BaseCard body-class="">
      <!-- Toolbar -->
      <div class="p-4 border-b flex flex-wrap items-center gap-3">
        <div class="relative flex-1 min-w-[220px]">
          <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-meta" />
          <input
            v-model="search"
            type="text"
            placeholder="Tìm mã VĐ, người nhận, SĐT..."
            class="w-full h-10 pl-9 pr-3 bg-surface border rounded-[var(--r-md)] text-sm text-strong placeholder:text-meta outline-none focus:border-[var(--primary)] focus:ring-2 focus:ring-[var(--primary-ring)]/40 transition-shadow"
          />
        </div>
        <div class="flex items-center gap-2 overflow-x-auto">
          <button
            @click="statusFilter = ''"
            class="px-3 h-8 rounded-full text-xs font-medium whitespace-nowrap transition-colors border"
            :class="statusFilter === '' ? 'bg-primary text-white border-transparent' : 'bg-surface text-body border-[var(--border)] hover:bg-subtle'"
          >Tất cả</button>
          <button
            v-for="s in STATUS_ORDER"
            :key="s"
            @click="statusFilter = s"
            class="px-3 h-8 rounded-full text-xs font-medium whitespace-nowrap transition-all border"
            :style="statusFilter === s ? { color: statusConfig(s).fg, background: statusConfig(s).bg, borderColor: statusConfig(s).fg } : {}"
            :class="statusFilter === s ? 'font-semibold' : 'bg-surface text-body border-[var(--border)] hover:bg-subtle'"
          >{{ statusConfig(s).label }}</button>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="bg-subtle text-meta text-[12px] uppercase tracking-wide">
              <th class="px-5 py-3 font-medium">Mã vận đơn</th>
              <th class="px-5 py-3 font-medium">Người nhận</th>
              <th class="px-5 py-3 font-medium">COD / Phí</th>
              <th class="px-5 py-3 font-medium">Trạng thái</th>
              <th class="px-5 py-3 font-medium text-right">Thao tác</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="loading">
              <tr v-for="i in 6" :key="i" class="border-t">
                <td v-for="c in 5" :key="c" class="px-5 py-4"><div class="skeleton h-4 w-full"></div></td>
              </tr>
            </template>
            <tr v-else-if="filteredOrders.length === 0">
              <td colspan="5" class="px-5 py-16 text-center">
                <PackageX class="w-10 h-10 mx-auto text-meta/40 mb-3" />
                <p class="text-meta text-sm">{{ orders.length === 0 ? 'Chưa có vận đơn nào. Tạo đơn đầu tiên nhé.' : 'Không tìm thấy đơn khớp bộ lọc.' }}</p>
              </td>
            </tr>
            <tr v-else v-for="order in filteredOrders" :key="order.id" class="border-t hover:bg-subtle transition-colors">
              <td class="px-5 py-4">
                <span class="font-mono text-[13px] font-medium text-strong">{{ order.tracking_number }}</span>
              </td>
              <td class="px-5 py-4">
                <div class="font-medium text-strong">{{ order.receiver_name }}</div>
                <div class="text-meta text-xs">{{ order.receiver_phone }}</div>
              </td>
              <td class="px-5 py-4">
                <div class="font-semibold text-strong tabular">{{ formatMoney(order.cod_amount) }}đ</div>
                <div class="text-meta text-xs tabular">Phí: {{ formatMoney(order.shipping_fee) }}đ</div>
              </td>
              <td class="px-5 py-4"><StatusBadge :status="order.status" /></td>
              <td class="px-5 py-4 text-right">
                <button
                  @click="printOrderPDF(order.id)"
                  class="inline-flex items-center gap-1 px-3 py-1.5 bg-subtle hover:bg-subtle-hover text-strong text-xs font-medium rounded-md border border-[var(--border)] transition-colors"
                  title="In vận đơn PDF"
                >
                  <Printer class="w-3.5 h-3.5 text-primary" />
                  <span>In VĐ</span>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination UI -->
      <div v-if="totalPages > 1" class="px-5 py-3 border-t bg-gray-50 flex items-center justify-between text-xs text-meta">
        <div>
          Trang <span class="font-medium text-strong">{{ page }}</span> / <span class="font-medium text-strong">{{ totalPages }}</span>
        </div>
        <div class="flex gap-2">
          <button @click="changePage(page - 1)" :disabled="page <= 1" class="px-3 py-1.5 rounded border bg-white disabled:opacity-50">
            Trước
          </button>
          <button @click="changePage(page + 1)" :disabled="page >= totalPages" class="px-3 py-1.5 rounded border bg-white disabled:opacity-50">
            Sau
          </button>
        </div>
      </div>
    </BaseCard>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import api from '../../services/api';
import { useToastStore } from '../../stores/toastStore';
import { STATUS_ORDER, statusConfig } from '../../composables/useStatus';
import { RefreshCw, Search, Plus, PackageX, Printer } from 'lucide-vue-next';
import { usePdfPrint } from '../../composables/usePdfPrint';
import BaseCard from '../../components/ui/BaseCard.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';

const toast = useToastStore();
const { printOrderPDF } = usePdfPrint();

const orders = ref([]);
const loading = ref(true);
const search = ref('');
const statusFilter = ref('');
const page = ref(1);
const limit = ref(10);
const totalPages = ref(1);

const fetchOrders = async () => {
  loading.value = true;
  try {
    const res = await api.get('/orders', { params: { page: page.value, limit: limit.value } });
    if (res.success) {
      if (res.data && res.meta) {
         orders.value = res.data;
         totalPages.value = res.meta.total_pages || 1;
      } else {
         orders.value = res.data || [];
         totalPages.value = 1;
      }
    }
  } catch (error) {
    toast.error('Không thể tải danh sách vận đơn');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const changePage = (newPage) => {
  if (newPage < 1 || newPage > totalPages.value) return;
  page.value = newPage;
  fetchOrders();
};

onMounted(fetchOrders);

const filteredOrders = computed(() => {
  const q = search.value.trim().toLowerCase();
  return orders.value.filter((o) => {
    if (statusFilter.value && o.status !== statusFilter.value) return false;
    if (!q) return true;
    return (
      (o.tracking_number || '').toLowerCase().includes(q) ||
      (o.receiver_name || '').toLowerCase().includes(q) ||
      (o.receiver_phone || '').includes(q)
    );
  });
});

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
</script>
