<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-strong flex items-center gap-2">
          <ScanLine class="w-6 h-6 text-primary" /> Quét mã vạch xuất/nhập kho
        </h1>
        <p class="text-xs text-meta mt-0.5">Dành cho nhân viên Hub. Tự động xử lý trạng thái khi bắn mã vạch.</p>
      </div>
      <div class="flex items-center gap-2">
        <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700 border border-emerald-200">
          <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
          Sẵn sàng quét
        </span>
      </div>
    </div>

    <!-- Ô nhập mã vạch chính (Always Focused) -->
    <div class="bg-surface border rounded-[var(--r-lg)] p-5 shadow-e1">
      <form @submit.prevent="handleScanSubmit" class="space-y-3">
        <label class="block text-xs font-semibold text-strong uppercase tracking-wider">
          Mã vận đơn (Quét hoặc gõ và bấm Enter)
        </label>
        <div class="relative flex items-center">
          <Barcode class="w-5 h-5 absolute left-3.5 text-meta" />
          <input
            ref="scanInputRef"
            v-model="tracking"
            type="text"
            autocomplete="off"
            autofocus
            placeholder="Bắn mã vạch vào đây (Ví dụ: OE-1712345678)..."
            class="w-full h-12 pl-11 pr-24 bg-subtle border rounded-[var(--r-md)] text-base font-mono font-bold text-strong placeholder:font-sans placeholder:text-meta outline-none focus:border-[var(--primary)] focus:ring-2 focus:ring-[var(--primary-ring)]/40 transition-all"
            @blur="keepFocus"
          />
          <BaseButton
            type="submit"
            variant="primary"
            size="md"
            class="absolute right-1.5"
            :loading="processing"
          >
            Quét đơn
          </BaseButton>
        </div>
      </form>
    </div>

    <!-- Kết quả quét gần nhất -->
    <div v-if="lastScan" class="rounded-[var(--r-lg)] border p-4 transition-all" :class="lastScan.success ? 'bg-emerald-50/70 border-emerald-200' : 'bg-rose-50/70 border-rose-200'">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <CheckCircle2 v-if="lastScan.success" class="w-6 h-6 text-emerald-600 shrink-0" />
          <XCircle v-else class="w-6 h-6 text-rose-600 shrink-0" />
          <div>
            <div class="font-mono text-base font-bold text-strong">{{ lastScan.trackingNumber }}</div>
            <div class="text-xs mt-0.5" :class="lastScan.success ? 'text-emerald-800 font-medium' : 'text-rose-800 font-medium'">
              {{ lastScan.message }}
            </div>
          </div>
        </div>
        <span class="text-xs text-meta font-mono">{{ lastScan.time }}</span>
      </div>
    </div>

    <!-- Danh sách lịch sử quét trong phiên -->
    <BaseCard title="Nhật ký quét trong phiên">
      <div v-if="scanLogs.length === 0" class="py-10 text-center text-meta text-xs">
        Chưa có lượt quét nào trong phiên làm việc này.
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="bg-subtle text-meta text-[11px] uppercase tracking-wide">
              <th class="px-4 py-2.5 font-medium">Thời gian</th>
              <th class="px-4 py-2.5 font-medium">Mã vận đơn</th>
              <th class="px-4 py-2.5 font-medium">Trạng thái mới</th>
              <th class="px-4 py-2.5 font-medium">Kết quả</th>
            </tr>
          </thead>
          <tbody class="divide-y">
            <tr v-for="(log, idx) in scanLogs" :key="idx" class="hover:bg-subtle/50">
              <td class="px-4 py-2.5 font-mono text-xs text-meta">{{ log.time }}</td>
              <td class="px-4 py-2.5 font-mono font-medium text-strong">{{ log.trackingNumber }}</td>
              <td class="px-4 py-2.5"><StatusBadge :status="log.newStatus || 'N/A'" /></td>
              <td class="px-4 py-2.5 text-xs">
                <span v-if="log.success" class="text-emerald-600 font-medium flex items-center gap-1">
                  <CheckCircle2 class="w-3.5 h-3.5" /> Thành công
                </span>
                <span v-else class="text-rose-600 font-medium flex items-center gap-1">
                  <XCircle class="w-3.5 h-3.5" /> {{ log.error }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </BaseCard>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue';
import api from '../../services/api';
import { useAuthStore } from '../../stores/authStore';
import { useToastStore } from '../../stores/toastStore';
import { actionsFor } from '../../composables/useMemberActions';
import { ScanLine, Barcode, CheckCircle2, XCircle } from 'lucide-vue-next';
import BaseButton from '../../components/ui/BaseButton.vue';
import BaseCard from '../../components/ui/BaseCard.vue';
import StatusBadge from '../../components/ui/StatusBadge.vue';

const authStore = useAuthStore();
const toast = useToastStore();

const scanInputRef = ref(null);
const tracking = ref('');
const processing = ref(false);
const lastScan = ref(null);
const scanLogs = ref([]);

const keepFocus = () => {
  setTimeout(() => {
    scanInputRef.value?.focus();
  }, 100);
};

onMounted(() => {
  nextTick(() => {
    scanInputRef.value?.focus();
  });
});

const playAudioBeep = (type = 'success') => {
  try {
    const ctx = new (window.AudioContext || window.webkitAudioContext)();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = 'sine';
    osc.frequency.setValueAtTime(type === 'success' ? 880 : 300, ctx.currentTime);
    gain.gain.setValueAtTime(0.3, ctx.currentTime);
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.start();
    osc.stop(ctx.currentTime + (type === 'success' ? 0.2 : 0.45));
  } catch (e) {
    console.error('Audio beep failed:', e);
  }
};

const handleScanSubmit = async () => {
  const code = tracking.value.trim();
  if (!code) return;

  tracking.value = '';
  processing.value = true;

  const timeStr = new Date().toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit', second: '2-digit' });

  try {
    // 1. Tra cứu vận đơn
    const searchRes = await api.get(`/tracking/${encodeURIComponent(code)}`);
    if (!searchRes.success || !searchRes.data?.order) {
      throw new Error('Không tìm thấy vận đơn');
    }

    const order = searchRes.data.order;
    const actions = actionsFor(authStore.user?.role || 'hub_staff', order.status);

    if (!actions || actions.length === 0) {
      throw new Error(`Đơn ở trạng thái ${order.status} không có thao tác kho phù hợp`);
    }

    // Chọn hành động chính (mặc định action đầu tiên: VD picked_up -> hub_inbound)
    const targetAction = actions[0];

    // 2. Chuyển trạng thái đơn kho
    const updateRes = await api.put(`/orders/${order.id}/status`, {
      status: targetAction.to,
      note: 'Quét tự động tại kho'
    });

    if (updateRes.success) {
      playAudioBeep('success');
      lastScan.value = {
        success: true,
        trackingNumber: order.tracking_number,
        message: `Đã chuyển trạng thái -> ${targetAction.label}`,
        time: timeStr
      };
      scanLogs.value.unshift({
        time: timeStr,
        trackingNumber: order.tracking_number,
        newStatus: targetAction.to,
        success: true
      });
      toast.success(`Quét đơn ${order.tracking_number} thành công`);
    } else {
      throw new Error('Cập nhật trạng thái thất bại');
    }
  } catch (err) {
    playAudioBeep('error');
    const errMsg = err.response?.data?.error || err.message || 'Lỗi quét đơn';
    lastScan.value = {
      success: false,
      trackingNumber: code,
      message: errMsg,
      time: timeStr
    };
    scanLogs.value.unshift({
      time: timeStr,
      trackingNumber: code,
      newStatus: null,
      success: false,
      error: errMsg
    });
    toast.error(`Lỗi quét đơn ${code}: ${errMsg}`);
  } finally {
    processing.value = false;
    keepFocus();
  }
};
</script>
