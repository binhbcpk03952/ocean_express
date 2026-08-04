<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold text-strong">Ví & Đối soát</h1>
      <p class="text-meta text-sm mt-1">Tiền thu hộ (COD) trừ cước vận chuyển, chờ đối soát chi trả</p>
    </div>

    <!-- Số dư khả dụng -->
    <div class="rounded-[var(--r-lg)] border bg-surface p-5 shadow-e1">
      <div class="flex items-center gap-2 text-meta text-sm mb-1">
        <Wallet class="w-4 h-4" /> Số dư khả dụng (chưa đối soát)
      </div>
      <div v-if="loading" class="skeleton h-9 w-40 mt-1"></div>
      <div v-else class="text-3xl font-bold tabular" :class="balance >= 0 ? 'text-[var(--st-delivered-fg)]' : 'text-[var(--st-returned-fg)]'">
        {{ formatMoney(balance) }}đ
      </div>
      <p class="text-xs text-meta mt-2">
        Là tổng tiền COD đã thu trừ cước của các đơn giao thành công, chưa nằm trong phiên chi trả nào.
      </p>
    </div>

    <!-- Lịch sử phiên chi trả -->
    <div>
      <h2 class="text-sm font-semibold text-strong mb-3">Lịch sử đối soát</h2>
      <div v-if="loading" class="space-y-2">
        <div v-for="i in 2" :key="i" class="skeleton h-14 w-full rounded-[var(--r-md)]"></div>
      </div>
      <div v-else-if="settlements.length === 0" class="rounded-[var(--r-md)] bg-subtle px-4 py-6 text-center text-sm text-meta">
        Chưa có phiên đối soát nào.
      </div>
      <div v-else class="space-y-2">
        <div
          v-for="s in settlements"
          :key="s.id"
          class="flex items-center justify-between rounded-[var(--r-md)] border bg-surface px-4 py-3"
        >
          <div>
            <div class="font-semibold text-strong tabular">{{ formatMoney(s.total_amount) }}đ</div>
            <div class="text-xs text-meta">Chốt: {{ formatDate(s.created_at) }}</div>
          </div>
          <span
            class="inline-flex items-center gap-1.5 rounded-sm px-2.5 py-1 text-xs font-medium"
            :style="s.status === 'paid'
              ? { color: 'var(--st-delivered-fg)', background: 'var(--st-delivered-bg)' }
              : { color: 'var(--st-ready-fg)', background: 'var(--st-ready-bg)' }"
          >
            <span class="h-1.5 w-1.5 rounded-full" style="background: currentColor"></span>
            {{ s.status === 'paid' ? 'Đã chi trả' : 'Chờ chi trả' }}
          </span>
        </div>
      </div>
    </div>

    <!-- Lịch sử bút toán -->
    <div>
      <h2 class="text-sm font-semibold text-strong mb-3">Lịch sử giao dịch ví</h2>
      <div v-if="loading" class="space-y-2">
        <div v-for="i in 4" :key="i" class="skeleton h-12 w-full rounded-[var(--r-md)]"></div>
      </div>
      <div v-else-if="transactions.length === 0" class="rounded-[var(--r-md)] bg-subtle px-4 py-6 text-center text-sm text-meta">
        Chưa có giao dịch nào. Giao dịch được ghi khi đơn giao thành công.
      </div>
      <div v-else class="overflow-hidden rounded-[var(--r-md)] border">
        <div
          v-for="tx in transactions"
          :key="tx.id"
          class="flex items-center justify-between border-b bg-surface px-4 py-3 last:border-b-0"
        >
          <div class="min-w-0">
            <div class="text-sm text-body truncate">{{ tx.note }}</div>
            <div class="text-xs text-meta">
              {{ formatDate(tx.created_at) }}
              <span v-if="tx.settlement_id" class="ml-1">· đã đối soát</span>
            </div>
          </div>
          <div class="font-semibold tabular shrink-0 ml-3" :class="tx.amount >= 0 ? 'text-[var(--st-delivered-fg)]' : 'text-[var(--st-returned-fg)]'">
            {{ tx.amount >= 0 ? '+' : '' }}{{ formatMoney(tx.amount) }}đ
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import api from '../../services/api';
import { useToastStore } from '../../stores/toastStore';
import { Wallet } from 'lucide-vue-next';

const toast = useToastStore();

const loading = ref(true);
const balance = ref(0);
const transactions = ref([]);
const settlements = ref([]);

const fetchData = async () => {
  loading.value = true;
  try {
    const [wallet, sett] = await Promise.all([
      api.get('/shop/wallet'),
      api.get('/shop/settlements'),
    ]);
    if (wallet.success) {
      balance.value = wallet.data.available_balance || 0;
      transactions.value = wallet.data.transactions || [];
    }
    if (sett.success) settlements.value = sett.data || [];
  } catch (error) {
    toast.error('Không thể tải thông tin ví');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

onMounted(fetchData);

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
const formatDate = (s) => {
  if (!s) return '';
  return new Date(s).toLocaleString('vi-VN', { hour: '2-digit', minute: '2-digit', day: '2-digit', month: '2-digit', year: 'numeric' });
};
</script>
