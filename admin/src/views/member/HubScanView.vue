<template>
  <div class="space-y-4">
    <div>
      <h1 class="text-lg font-semibold text-strong">Quét đơn nhập/xuất kho</h1>
      <p class="text-xs text-meta mt-0.5">Nhập mã vận đơn để tra cứu và xử lý tại kho</p>
    </div>

    <!-- Ô nhập mã vận đơn -->
    <form @submit.prevent="lookup" class="flex items-center gap-2">
      <div class="relative flex-1">
        <ScanLine class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-meta" />
        <input
          v-model="tracking"
          type="text"
          inputmode="text"
          autocomplete="off"
          autocapitalize="characters"
          placeholder="VD: OE-1712345678"
          class="w-full h-11 pl-9 pr-3 bg-surface border rounded-[var(--r-md)] text-sm font-mono text-strong placeholder:text-meta placeholder:font-sans outline-none focus:border-[var(--primary)] focus:ring-2 focus:ring-[var(--primary-ring)]/40 transition-shadow"
        />
      </div>
      <BaseButton type="submit" variant="primary" size="md" :loading="loading">
        <Search class="w-4 h-4" /> Tra cứu
      </BaseButton>
    </form>

    <!-- Chưa tra cứu -->
    <div v-if="!order && !notFound && !loading" class="py-16 text-center">
      <PackageSearch class="w-12 h-12 mx-auto text-meta/40 mb-3" />
      <p class="text-meta text-sm">Nhập mã vận đơn phía trên để bắt đầu.</p>
    </div>

    <!-- Không tìm thấy -->
    <div v-else-if="notFound" class="py-16 text-center">
      <PackageX class="w-12 h-12 mx-auto text-danger/50 mb-3" />
      <p class="text-body text-sm font-medium">Không tìm thấy vận đơn</p>
      <p class="text-meta text-xs mt-1">Kiểm tra lại mã <span class="font-mono">{{ lastQuery }}</span></p>
    </div>

    <!-- Kết quả -->
    <div v-else-if="order" class="space-y-4">
      <div class="rounded-[var(--r-lg)] border bg-surface p-4 shadow-e1">
        <div class="flex items-start justify-between gap-3">
          <span class="font-mono text-sm font-semibold text-strong">{{ order.tracking_number }}</span>
          <StatusBadge :status="order.status" />
        </div>
        <div class="mt-3 space-y-1.5 text-sm">
          <div class="flex items-start gap-2">
            <User class="w-4 h-4 text-meta mt-0.5 shrink-0" />
            <div>
              <span class="font-medium text-strong">{{ order.receiver_name }}</span>
              <span class="text-meta"> · {{ order.receiver_phone }}</span>
            </div>
          </div>
          <div class="flex items-start gap-2">
            <MapPin class="w-4 h-4 text-meta mt-0.5 shrink-0" />
            <span class="text-body text-[13px] leading-snug">{{ order.receiver_address_detail }}</span>
          </div>
        </div>
      </div>

      <!-- Hành động tại kho -->
      <div v-if="actions.length" class="space-y-2">
        <BaseButton
          v-for="act in actions"
          :key="act.to"
          :variant="act.variant"
          size="md"
          class="w-full"
          :loading="updating === act.to"
          @click="doAction(act)"
        >
          {{ act.label }}
        </BaseButton>
      </div>
      <div v-else class="rounded-[var(--r-md)] bg-subtle px-3 py-3 text-center text-xs text-meta">
        Đơn ở trạng thái "{{ statusConfig(order.status).label }}" — không có thao tác kho nào khả dụng.
      </div>

      <router-link :to="{ name: 'MemberOrderDetail', params: { id: order.id } }" class="block">
        <BaseButton variant="ghost" size="sm" class="w-full">Xem chi tiết hành trình</BaseButton>
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import api from '../../services/api';
import { useAuthStore } from '../../stores/authStore';
import { useToastStore } from '../../stores/toastStore';
import { statusConfig } from '../../composables/useStatus';
import { actionsFor } from '../../composables/useMemberActions';
import { ScanLine, Search, PackageSearch, PackageX, User, MapPin } from 'lucide-vue-next';
import BaseButton from '../../components/ui/BaseButton.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';

const authStore = useAuthStore();
const toast = useToastStore();

const tracking = ref('');
const lastQuery = ref('');
const loading = ref(false);
const notFound = ref(false);
const order = ref(null);
const updating = ref('');

const actions = computed(() => actionsFor(authStore.user?.role, order.value?.status));

const lookup = async () => {
  const code = tracking.value.trim();
  if (!code) return;
  loading.value = true;
  notFound.value = false;
  order.value = null;
  lastQuery.value = code;
  try {
    const res = await api.get(`/tracking/${encodeURIComponent(code)}`);
    if (res.success) {
      order.value = res.data.order;
    } else {
      notFound.value = true;
    }
  } catch (error) {
    if (error.response?.status === 404 || error.response?.status === 500) {
      notFound.value = true;
    } else {
      toast.error('Không thể tra cứu vận đơn');
    }
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const doAction = async (act) => {
  if (!order.value) return;
  updating.value = act.to;
  try {
    const res = await api.put(`/orders/${order.value.id}/status`, {
      status: act.to,
      note: '',
    });
    if (res.success) {
      toast.success('Cập nhật trạng thái thành công');
      // Tra cứu lại để phản ánh trạng thái mới
      await lookup();
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Cập nhật thất bại');
  } finally {
    updating.value = '';
  }
};
</script>
