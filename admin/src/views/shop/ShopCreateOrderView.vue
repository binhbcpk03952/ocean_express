<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center gap-3">
      <router-link :to="{ name: 'ShopDashboard' }">
        <BaseButton variant="secondary" size="sm"><ArrowLeft class="w-4 h-4" /> Quay lại</BaseButton>
      </router-link>
      <h1 class="text-2xl font-semibold text-strong">Tạo vận đơn</h1>
    </div>

    <!-- 2-column layout -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
      <!-- Left: Form chính (2/3) -->
      <div class="lg:col-span-2 space-y-5">
        <BaseCard title="Thông tin người nhận (Sổ địa chỉ)">
          <div class="space-y-4">
            <!-- Autocomplete SĐT/Tên từ Sổ Địa Chỉ -->
            <div class="relative">
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div class="relative">
                  <FormField
                    v-model="form.receiver_phone"
                    label="Số điện thoại người nhận"
                    required
                    placeholder="Gõ SĐT để tìm khách cũ..."
                    @input="onPhoneInput"
                  />
                  <!-- Dropdown gợi ý Sổ địa chỉ -->
                  <div
                    v-if="customerSuggestions.length > 0"
                    class="absolute z-20 top-full left-0 right-0 mt-1 bg-surface border rounded-md shadow-lg max-h-56 overflow-y-auto divide-y"
                  >
                    <div
                      v-for="c in customerSuggestions"
                      :key="c.id"
                      @click="selectCustomer(c)"
                      class="p-2.5 hover:bg-subtle cursor-pointer text-xs space-y-0.5"
                    >
                      <div class="font-semibold text-strong">{{ c.name }} — <span class="font-mono text-primary">{{ c.phone }}</span></div>
                      <div class="text-meta truncate">{{ c.address_detail }}</div>
                    </div>
                  </div>
                </div>

                <FormField v-model="form.receiver_name" label="Tên người nhận" required placeholder="Nguyễn Văn B" />
              </div>
            </div>

            <LocationCascader v-model="form.receiver_location_id" label="Khu vực nhận (Huyện/Xã)" required />
            <FormField v-model="form.receiver_address_detail" label="Địa chỉ chi tiết (Số nhà, đường)" required placeholder="Số nhà, đường, phường..." />
            <div>
              <div class="text-sm font-medium mb-1">Định vị điểm giao hàng trên bản đồ</div>
              <div class="text-xs text-meta mb-3">Tùy chọn. Nếu ghim đúng vị trí, tài xế sẽ giao hàng nhanh chóng và chính xác hơn.</div>
              <LocationPickerMap v-model="form.receiver_coordinates" />
            </div>
          </div>
        </BaseCard>

        <BaseCard title="Thông tin hàng hóa & Kích thước">
          <div class="space-y-4">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <FormField v-model="form.weight" label="Trọng lượng thực (gram)" type="number" required />
              <FormField v-model="form.cod_amount" label="Tiền thu hộ COD (VNĐ)" type="number" />
            </div>

            <!-- Cước phí thể tích (Step 7) -->
            <div>
              <label class="block text-xs font-semibold text-strong uppercase tracking-wider mb-2">
                Kích thước Đóng gói (cm) - Quy đổi thể tích
              </label>
              <div class="grid grid-cols-3 gap-3">
                <FormField v-model="form.length" label="Dài (cm)" type="number" placeholder="0" />
                <FormField v-model="form.width" label="Rộng (cm)" type="number" placeholder="0" />
                <FormField v-model="form.height" label="Cao (cm)" type="number" placeholder="0" />
              </div>
              <div v-if="volumetricWeight > 0" class="mt-2 text-xs text-meta bg-subtle p-2.5 rounded-md flex justify-between">
                <span>Trọng lượng quy đổi thề tích (V/5000): <strong class="text-strong font-mono">{{ volumetricWeight }} gram</strong></span>
                <span>Trọng lượng tính cước: <strong class="text-primary font-bold font-mono">{{ chargeableWeight }} gram</strong></span>
              </div>
            </div>
          </div>
        </BaseCard>
      </div>

      <!-- Right: Tóm tắt & Submit (1/3) -->
      <div class="space-y-4">
        <!-- Cước phí -->
        <BaseCard title="Cước phí">
          <div class="space-y-4">
            <div class="rounded-[var(--r-md)] bg-subtle border px-4 py-5 text-center">
              <div class="text-xs text-meta mb-1">Ước tính cước phí</div>
              <div class="text-2xl font-bold" style="color: var(--primary)">
                {{ estimatedFee === null ? '—' : formatMoney(estimatedFee) + 'đ' }}
              </div>
              <div v-if="chargeableWeight" class="text-[11px] text-meta mt-1">
                Theo TL tính cước: {{ chargeableWeight }}g
              </div>
            </div>
            <BaseButton variant="secondary" size="sm" class="w-full" :loading="calculating" @click="calcFee">
              <Calculator class="w-4 h-4" /> Tính thử cước
            </BaseButton>
          </div>
        </BaseCard>

        <!-- Tóm tắt -->
        <BaseCard title="Tóm tắt">
          <dl class="space-y-3 text-sm">
            <div class="flex justify-between gap-2">
              <dt class="text-meta">Người nhận</dt>
              <dd class="text-strong font-medium text-right truncate max-w-[60%]">{{ form.receiver_name || '—' }}</dd>
            </div>
            <div class="flex justify-between gap-2">
              <dt class="text-meta">SĐT</dt>
              <dd class="text-strong font-medium">{{ form.receiver_phone || '—' }}</dd>
            </div>
            <div class="flex justify-between gap-2">
              <dt class="text-meta">TL Thực / Quy đổi</dt>
              <dd class="text-strong font-medium">{{ form.weight || 0 }}g / {{ volumetricWeight }}g</dd>
            </div>
            <div class="flex justify-between gap-2">
              <dt class="text-meta">COD</dt>
              <dd class="text-strong font-medium">{{ form.cod_amount > 0 ? formatMoney(form.cod_amount) + 'đ' : 'Không' }}</dd>
            </div>
          </dl>
          <div class="mt-4 pt-4 border-t">
            <BaseButton variant="primary" size="md" class="w-full" :loading="submitting" @click="submit">
              <PackageCheck class="w-4 h-4" /> Tạo vận đơn
            </BaseButton>
          </div>
        </BaseCard>
      </div>
    </div>

    <!-- Modal thành công -->
    <BaseModal v-model="showSuccess" title="Tạo đơn thành công" :close-on-backdrop="false">
      <div class="space-y-3">
        <div class="flex items-start gap-2 px-3 py-2.5 rounded-md text-sm bg-success/10 text-[var(--success)]">
          <PackageCheck class="w-4 h-4 mt-0.5 shrink-0" />
          <span>Vận đơn đã được tạo và thông tin người nhận đã lưu vào sổ địa chỉ.</span>
        </div>
        <div class="bg-subtle border rounded-[var(--r-md)] p-3 flex items-center justify-between">
          <div>
            <div class="text-xs text-meta">Mã vận đơn</div>
            <div class="font-mono text-base font-bold text-primary">{{ createdTracking }}</div>
          </div>
          <router-link :to="{ name: 'ShopOrderDetail', params: { id: createdOrderId } }">
            <BaseButton variant="ghost" size="sm">Chi tiết →</BaseButton>
          </router-link>
        </div>
      </div>
      <template #footer>
        <BaseButton variant="primary" @click="showPrintModal = true">
          <Printer class="w-4 h-4" /> In Vận Đơn Ngay
        </BaseButton>
        <BaseButton variant="secondary" @click="createAnother">Tạo đơn khác</BaseButton>
        <BaseButton variant="secondary" @click="goToOrders">Xem danh sách</BaseButton>
      </template>
    </BaseModal>

    <!-- Shipping Label Modal -->
    <ShippingLabelModal
      v-model="showPrintModal"
      :order="createdOrderObj"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import api from '../../services/api';
import { useToastStore } from '../../stores/toastStore';
import { ArrowLeft, Calculator, PackageCheck, Printer } from 'lucide-vue-next';
import BaseCard from '../../components/ui/BaseCard.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import BaseModal from '../../components/ui/BaseModal.vue';
import FormField from '../../components/ui/FormField.vue';
import LocationCascader from '../../components/ui/LocationCascader.vue';
import LocationPickerMap from '../../components/LocationPickerMap.vue';
import ShippingLabelModal from '../../components/ShippingLabelModal.vue';

const router = useRouter();
const toast = useToastStore();

const emptyForm = () => ({
  receiver_name: '',
  receiver_phone: '',
  receiver_location_id: '',
  receiver_address_detail: '',
  receiver_coordinates: null,
  weight: 500,
  length: 0,
  width: 0,
  height: 0,
  cod_amount: 0,
});

const form = ref(emptyForm());
const estimatedFee = ref(null);
const calculating = ref(false);
const submitting = ref(false);
const showSuccess = ref(false);
const createdTracking = ref('');
const createdOrderId = ref('');
const createdOrderObj = ref(null);
const showPrintModal = ref(false);

const customerSuggestions = ref([]);
let searchTimeout = null;

// Tính trọng lượng quy đổi thể tích (Volumetric weight)
const volumetricWeight = computed(() => {
  const l = Number(form.value.length) || 0;
  const w = Number(form.value.width) || 0;
  const h = Number(form.value.height) || 0;
  if (l > 0 && w > 0 && h > 0) {
    return Math.round((l * w * h) / 5);
  }
  return 0;
});

// Trọng lượng tính cước = Max(Trọng lượng thực, Quy đổi thể tích)
const chargeableWeight = computed(() => {
  const actWeight = Number(form.value.weight) || 0;
  return Math.max(actWeight, volumetricWeight.value);
});

// Tim kiem khach hang trong so dia chi
const onPhoneInput = () => {
  clearTimeout(searchTimeout);
  const q = form.value.receiver_phone.trim();
  if (q.length < 2) {
    customerSuggestions.value = [];
    return;
  }
  searchTimeout = setTimeout(async () => {
    try {
      const res = await api.get('/shop/customers', { params: { q } });
      if (res.success) {
        customerSuggestions.value = res.data || [];
      }
    } catch (e) {
      console.error(e);
    }
  }, 300);
};

const selectCustomer = (c) => {
  form.value.receiver_name = c.name;
  form.value.receiver_phone = c.phone;
  if (c.location_id) form.value.receiver_location_id = c.location_id;
  if (c.address_detail) form.value.receiver_address_detail = c.address_detail;
  if (c.latitude && c.longitude) {
    form.value.receiver_coordinates = { lat: Number(c.latitude), lng: Number(c.longitude) };
  }
  customerSuggestions.value = [];
};

const calcFee = async () => {
  if (!form.value.receiver_location_id || !form.value.weight) {
    toast.warning('Chọn khu vực nhận và nhập khối lượng để tính cước');
    return;
  }
  calculating.value = true;
  try {
    const res = await api.post('/shop/rates/calculate', {
      receiver_location_id: form.value.receiver_location_id,
      weight: Number(form.value.weight),
      length: Number(form.value.length) || 0,
      width: Number(form.value.width) || 0,
      height: Number(form.value.height) || 0,
    });
    if (res.success) estimatedFee.value = res.data.fee;
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Không tính được cước');
  } finally {
    calculating.value = false;
  }
};

const submit = async () => {
  const f = form.value;
  if (!f.receiver_name || !f.receiver_phone || !f.receiver_location_id || !f.receiver_address_detail || !f.weight) {
    toast.warning('Vui lòng điền đầy đủ thông tin bắt buộc');
    return;
  }
  submitting.value = true;
  try {
    // 1. Lưu/cập nhật thông tin khách hàng vào Sổ địa chỉ (Step 6)
    api.post('/shop/customers', {
      name: f.receiver_name,
      phone: f.receiver_phone,
      location_id: f.receiver_location_id,
      address_detail: f.receiver_address_detail,
      latitude: f.receiver_coordinates?.lat || null,
      longitude: f.receiver_coordinates?.lng || null,
    }).catch(e => console.warn('Lỗi lưu sổ địa chỉ:', e));

    // 2. Tạo vận đơn (Step 7: với length, width, height)
    const res = await api.post('/shop/orders', {
      receiver_name: f.receiver_name,
      receiver_phone: f.receiver_phone,
      receiver_location_id: f.receiver_location_id,
      receiver_address_detail: f.receiver_address_detail,
      receiver_latitude: f.receiver_coordinates?.lat || null,
      receiver_longitude: f.receiver_coordinates?.lng || null,
      weight: Number(f.weight),
      length: Number(f.length) || 0,
      width: Number(f.width) || 0,
      height: Number(f.height) || 0,
      cod_amount: Number(f.cod_amount) || 0,
    });
    if (res.success) {
      createdTracking.value = res.data.tracking_number;
      createdOrderId.value = res.data.id;
      createdOrderObj.value = res.data;
      showSuccess.value = true;
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Tạo đơn thất bại');
  } finally {
    submitting.value = false;
  }
};

const createAnother = () => {
  showSuccess.value = false;
  form.value = emptyForm();
  estimatedFee.value = null;
  customerSuggestions.value = [];
};

const goToOrders = () => {
  router.push({ name: 'ShopOrders' });
};

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
</script>