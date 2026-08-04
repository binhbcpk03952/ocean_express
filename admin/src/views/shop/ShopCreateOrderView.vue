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
        <BaseCard title="Thông tin người nhận">
          <div class="space-y-4">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <FormField v-model="form.receiver_name" label="Tên người nhận" required placeholder="Nguyễn Văn B" />
              <FormField v-model="form.receiver_phone" label="Số điện thoại" required placeholder="09..." />
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

        <BaseCard title="Thông tin hàng hóa">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <FormField v-model="form.weight" label="Khối lượng (gram)" type="number" required />
            <FormField v-model="form.cod_amount" label="Tiền thu hộ COD (VNĐ)" type="number" />
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
              <dt class="text-meta">Khối lượng</dt>
              <dd class="text-strong font-medium">{{ form.weight ? form.weight + ' gram' : '—' }}</dd>
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
          <span>Vận đơn đã được tạo và đang chờ tài xế tới lấy.</span>
        </div>
        <div class="bg-subtle border rounded-[var(--r-md)] p-3">
          <div class="text-xs text-meta">Mã vận đơn</div>
          <div class="font-mono text-base font-semibold text-strong">{{ createdTracking }}</div>
        </div>
      </div>
      <template #footer>
        <BaseButton variant="secondary" @click="createAnother">Tạo đơn khác</BaseButton>
        <BaseButton variant="primary" @click="goToOrders">Xem danh sách đơn</BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import api from '../../services/api';
import { useToastStore } from '../../stores/toastStore';
import { ArrowLeft, Calculator, PackageCheck } from 'lucide-vue-next';
import BaseCard from '../../components/ui/BaseCard.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import BaseModal from '../../components/ui/BaseModal.vue';
import FormField from '../../components/ui/FormField.vue';
import LocationCascader from '../../components/ui/LocationCascader.vue';
import LocationPickerMap from '../../components/LocationPickerMap.vue';

const router = useRouter();
const toast = useToastStore();

const emptyForm = () => ({
  receiver_name: '',
  receiver_phone: '',
  receiver_location_id: '',
  receiver_address_detail: '',
  receiver_coordinates: null,
  weight: 500,
  cod_amount: 0,
});

const form = ref(emptyForm());
const estimatedFee = ref(null);
const calculating = ref(false);
const submitting = ref(false);
const showSuccess = ref(false);
const createdTracking = ref('');

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
    const res = await api.post('/shop/orders', {
      receiver_name: f.receiver_name,
      receiver_phone: f.receiver_phone,
      receiver_location_id: f.receiver_location_id,
      receiver_address_detail: f.receiver_address_detail,
      receiver_latitude: f.receiver_coordinates?.lat || null,
      receiver_longitude: f.receiver_coordinates?.lng || null,
      weight: Number(f.weight),
      cod_amount: Number(f.cod_amount) || 0,
    });
    if (res.success) {
      createdTracking.value = res.data.tracking_number;
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
};

const goToOrders = () => {
  router.push({ name: 'ShopOrders' });
};

const formatMoney = (val) => new Intl.NumberFormat('vi-VN').format(val || 0);
</script>