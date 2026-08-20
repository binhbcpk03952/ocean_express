<template>
  <Teleport to="body">
    <div v-if="modelValue" class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-black/60 backdrop-blur-sm print-overlay">
      <!-- Modal Box (Ẩn khi in ấn, chỉ in phần printable-content) -->
      <div class="relative w-full max-w-4xl max-h-[90vh] bg-surface rounded-[var(--r-lg)] shadow-2xl flex flex-col overflow-hidden border border-[var(--border)] no-print">
        
        <!-- Header -->
        <div class="flex items-center justify-between px-6 py-4 border-b bg-subtle">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
              <Printer class="w-5 h-5" />
            </div>
            <div>
              <h2 class="text-lg font-bold text-strong">In Tem Vận Đơn (Shipping Label)</h2>
              <p class="text-xs text-meta">
                {{ ordersList.length > 1 ? `Đang xem ${ordersList.length} vận đơn đã chọn` : `Mã đơn: ${currentOrder?.tracking_number || '—'}` }}
              </p>
            </div>
          </div>

          <div class="flex items-center gap-2">
            <!-- Format selector -->
            <div class="flex items-center bg-surface border rounded-md p-0.5 text-xs">
              <button
                type="button"
                @click="paperSize = 'a6'"
                class="px-2.5 py-1 rounded transition-colors"
                :class="paperSize === 'a6' ? 'bg-primary text-white font-medium' : 'text-body hover:bg-subtle'"
                title="Khổ A6 (100x150mm) - Máy in nhiệt TMĐT tiêu chuẩn"
              >
                A6 (100x150mm Tem nhiệt)
              </button>
              <button
                type="button"
                @click="paperSize = 'a5'"
                class="px-2.5 py-1 rounded transition-colors"
                :class="paperSize === 'a5' ? 'bg-primary text-white font-medium' : 'text-body hover:bg-subtle'"
                title="Khổ A5 (148x210mm) - Nửa tờ A4"
              >
                A5 (Văn phòng)
              </button>
            </div>

            <button @click="close" class="p-2 text-meta hover:text-strong rounded-md hover:bg-subtle transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
        </div>

        <!-- Body / Label Preview Area -->
        <div class="flex-1 overflow-y-auto p-6 bg-slate-100 dark:bg-slate-900/60 flex flex-col items-center gap-6">
          <div
            v-for="(order, idx) in ordersList"
            :key="order.id || idx"
            class="shipping-label-card bg-white text-slate-900 shadow-md border border-slate-300 rounded-sm relative overflow-hidden select-none"
            :class="paperSize === 'a6' ? 'label-a6' : 'label-a5'"
          >
            <!-- Label Content -->
            <div class="p-4 flex flex-col h-full justify-between text-xs leading-tight font-sans">
              
              <!-- Top Row: Brand + Tracking Barcode + QR -->
              <div class="border-b-2 border-slate-900 pb-2.5 mb-2.5 flex items-start justify-between gap-3">
                <div class="flex-1">
                  <div class="flex items-center gap-1.5 mb-1">
                    <span class="text-sm font-black tracking-wider text-blue-800">OCEAN EXPRESS</span>
                    <span class="text-[9px] px-1 py-0.5 bg-blue-100 text-blue-800 font-bold rounded uppercase">Hỏa Tốc</span>
                  </div>
                  <div class="font-mono text-base font-black tracking-wide text-slate-900">{{ order.tracking_number }}</div>
                  <div class="text-[10px] text-slate-500 mt-0.5">Ngày tạo: {{ formatDateTime(order.created_at) }}</div>
                </div>

                <!-- QR Code container -->
                <div class="w-16 h-16 shrink-0 border border-slate-300 p-0.5 bg-white rounded flex items-center justify-center">
                  <img
                    :src="`https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${encodeURIComponent(order.tracking_number || '')}`"
                    :alt="order.tracking_number"
                    class="w-full h-full object-contain"
                  />
                </div>
              </div>

              <!-- Route Box: Sender & Receiver -->
              <div class="grid grid-cols-2 gap-2 border-b-2 border-slate-900 pb-2.5 mb-2.5">
                <!-- Sender -->
                <div class="pr-2 border-r border-slate-300 space-y-1">
                  <div class="text-[10px] font-bold uppercase text-slate-500">Từ (Người gửi):</div>
                  <div class="font-bold text-slate-900 truncate">{{ getSenderName(order) }}</div>
                  <div class="font-medium text-slate-700">{{ order.sender_phone || '09xxxxxxxx' }}</div>
                  <div class="text-[11px] text-slate-600 line-clamp-2">{{ order.sender_address_detail || 'Kho hàng đối tác' }}</div>
                </div>

                <!-- Receiver -->
                <div class="pl-1 space-y-1">
                  <div class="text-[10px] font-bold uppercase text-slate-500">Đến (Người nhận):</div>
                  <div class="font-bold text-slate-900 text-[13px] truncate">{{ order.receiver_name }}</div>
                  <div class="font-bold text-blue-700 text-xs">{{ order.receiver_phone }}</div>
                  <div class="text-[11px] font-medium text-slate-800 line-clamp-3">{{ order.receiver_address_detail }}</div>
                  <div v-if="order.receiver_location_id" class="text-[9px] text-slate-500 font-mono">Mã KV: {{ order.receiver_location_id }}</div>
                </div>
              </div>

              <!-- Package Specs & Financials -->
              <div class="grid grid-cols-3 gap-2 border-b-2 border-slate-900 pb-2.5 mb-2.5 text-center">
                <div class="border-r border-slate-300">
                  <div class="text-[10px] text-slate-500 font-semibold">Khối lượng</div>
                  <div class="font-bold text-sm text-slate-900">{{ order.weight || 500 }}g</div>
                </div>
                <div class="border-r border-slate-300">
                  <div class="text-[10px] text-slate-500 font-semibold">Kích thước</div>
                  <div class="font-mono text-xs text-slate-700">
                    {{ order.length || 0 }}x{{ order.width || 0 }}x{{ order.height || 0 }}cm
                  </div>
                </div>
                <div>
                  <div class="text-[10px] text-slate-500 font-semibold">Cước vận chuyển</div>
                  <div class="font-bold text-xs text-slate-900">{{ formatMoney(order.shipping_fee) }}đ</div>
                </div>
              </div>

              <!-- COD Block: Big & Bold -->
              <div class="border-2 border-slate-900 rounded p-2 mb-2.5 bg-slate-50 flex items-center justify-between">
                <div>
                  <div class="text-[10px] uppercase font-bold text-slate-600">Tiền Thu Hộ (COD)</div>
                  <div class="text-[10px] text-slate-500">Chỉ thu khi giao thành công</div>
                </div>
                <div class="text-lg font-black text-red-600 font-mono">
                  {{ order.cod_amount > 0 ? formatMoney(order.cod_amount) + ' VNĐ' : 'KHÔNG THU TIỀN' }}
                </div>
              </div>

              <!-- Instructions & Note -->
              <div class="border-b border-slate-300 pb-2 mb-2 text-[10px] text-slate-700">
                <span class="font-bold">Ghi chú:</span> Cho xem hàng, không thử. Quý khách vui lòng quay video khi bóc hàng để được bảo vệ quyền lợi.
              </div>

              <!-- Signatures -->
              <div class="grid grid-cols-2 gap-4 text-center pt-1 text-[11px]">
                <div>
                  <div class="font-bold text-slate-800">Chữ ký người nhận</div>
                  <div class="h-10 border-b border-dashed border-slate-400"></div>
                  <div class="text-[9px] text-slate-400 mt-0.5">(Ký và ghi rõ họ tên)</div>
                </div>
                <div>
                  <div class="font-bold text-slate-800">Chữ ký bưu tá</div>
                  <div class="h-10 border-b border-dashed border-slate-400"></div>
                  <div class="text-[9px] text-slate-400 mt-0.5">(Xác nhận đã giao đủ hàng)</div>
                </div>
              </div>

            </div>
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="flex items-center justify-between px-6 py-4 border-t bg-surface">
          <div class="text-xs text-meta">
            Mẹo: Dùng nút <strong class="text-strong">"In Ngay"</strong> để in trực tiếp qua hộp thoại máy in mà không lo bị trình duyệt chặn pop-up.
          </div>

          <div class="flex items-center gap-3">
            <button
              type="button"
              @click="downloadPDF"
              :disabled="downloading"
              class="px-4 py-2 bg-subtle hover:bg-subtle-hover text-strong text-xs font-semibold rounded-md border transition-colors flex items-center gap-2 disabled:opacity-50"
            >
              <Download class="w-4 h-4" />
              <span>{{ downloading ? 'Đang tạo PDF...' : 'Tải File PDF' }}</span>
            </button>

            <button
              type="button"
              @click="printDirect"
              class="px-5 py-2 bg-primary hover:bg-primary-hover text-white text-xs font-bold rounded-md shadow-md transition-colors flex items-center gap-2"
            >
              <Printer class="w-4 h-4" />
              <span>In Ngay (Print)</span>
            </button>
          </div>
        </div>

      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed } from 'vue';
import { Printer, X, Download } from 'lucide-vue-next';
import { useAuthStore } from '../stores/authStore';
import { useToastStore } from '../stores/toastStore';
import { usePdfPrint } from '../composables/usePdfPrint';

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  order: {
    type: Object,
    default: null
  },
  orders: {
    type: Array,
    default: () => []
  }
});

const emit = defineEmits(['update:modelValue']);

const authStore = useAuthStore();
const toast = useToastStore();
const { downloadOrderPDF, downloadBatchPDF } = usePdfPrint();

const paperSize = ref('a6'); // 'a6' hoặc 'a5'
const downloading = ref(false);

const ordersList = computed(() => {
  if (props.orders && props.orders.length > 0) return props.orders;
  if (props.order) return [props.order];
  return [];
});

const currentOrder = computed(() => ordersList.value[0] || null);

const close = () => {
  emit('update:modelValue', false);
};

const getSenderName = (order) => {
  if (order.sender_name) return order.sender_name;
  if (authStore.user?.name) return authStore.user.name;
  return 'Ocean Express Shop';
};

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);

const formatDateTime = (dateString) => {
  if (!dateString) return new Date().toLocaleDateString('vi-VN');
  const d = new Date(dateString);
  if (isNaN(d.getTime())) return dateString;
  const h = d.getHours().toString().padStart(2, '0');
  const m = d.getMinutes().toString().padStart(2, '0');
  const D = d.getDate().toString().padStart(2, '0');
  const M = (d.getMonth() + 1).toString().padStart(2, '0');
  const Y = d.getFullYear();
  return `${h}:${m} ${D}/${M}/${Y}`;
};

const printDirect = () => {
  window.print();
};

const downloadPDF = async () => {
  if (ordersList.value.length === 0) return;
  downloading.value = true;
  try {
    if (ordersList.value.length === 1) {
      await downloadOrderPDF(ordersList.value[0].id || ordersList.value[0].tracking_number);
    } else {
      const ids = ordersList.value.map(o => o.id || o.tracking_number);
      await downloadBatchPDF(ids);
    }
  } catch (err) {
    toast.error('Lỗi khi tải file PDF: ' + err.message);
  } finally {
    downloading.value = false;
  }
};
</script>

<style scoped>
/* Preview Styling on screen */
.label-a6 {
  width: 380px;
  height: 570px;
}
.label-a5 {
  width: 480px;
  height: 680px;
}

/* -------------------------------------------------------------
   PRINT CSS: Khi người dùng bấm window.print()
   Chỉ hiển thị các tem in, mỗi tem nằm trọn vẹn trên 1 trang in,
   loại bỏ thanh header, footer, nền web, modal backdrop.
---------------------------------------------------------------- */
@media print {
  body * {
    visibility: hidden !important;
  }

  .print-overlay, .print-overlay * {
    visibility: visible !important;
  }

  .no-print {
    display: none !important;
  }

  .print-overlay {
    position: absolute !important;
    left: 0 !important;
    top: 0 !important;
    width: 100% !important;
    height: auto !important;
    background: none !important;
    padding: 0 !important;
    margin: 0 !important;
    display: block !important;
  }

  .shipping-label-card {
    visibility: visible !important;
    box-shadow: none !important;
    border: 1.5px solid #000 !important;
    page-break-after: always !important;
    break-after: page !important;
    margin: 0 auto 20px auto !important;
    width: 100% !important;
    max-width: 100mm !important;
    height: 150mm !important;
    box-sizing: border-box !important;
  }
}
</style>
