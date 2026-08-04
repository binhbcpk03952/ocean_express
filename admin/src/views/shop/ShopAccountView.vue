<template>
  <div class="space-y-6 max-w-2xl">
    <div>
      <h1 class="text-2xl font-semibold text-strong">Thông tin tài khoản</h1>
      <p class="text-meta text-sm mt-1">Hồ sơ đối tác và cấu hình gửi hàng</p>
    </div>

    <div v-if="loading" class="space-y-4">
      <div class="skeleton h-40 w-full rounded-[var(--r-lg)]"></div>
      <div class="skeleton h-32 w-full rounded-[var(--r-lg)]"></div>
    </div>

    <template v-else-if="shop">
      <!-- Nhắc cấu hình khu vực khi còn thiếu -->
      <div
        v-if="!shop.location_id"
        class="flex items-start gap-2 px-4 py-3 rounded-[var(--r-md)] text-sm bg-warning/10 text-[var(--warning)]"
      >
        <TriangleAlert class="w-4 h-4 mt-0.5 shrink-0" />
        <span>Bạn cần chọn <b>khu vực gửi hàng</b> bên dưới thì mới tạo được vận đơn.</span>
      </div>

      <BaseCard title="Hồ sơ đối tác">
        <div class="space-y-4">
          <div class="flex justify-between items-center text-sm border-b pb-3">
            <span class="text-meta">Email đăng nhập</span>
            <span class="text-body">{{ shop.email || '—' }}</span>
          </div>
          <div class="flex justify-between items-center text-sm border-b pb-3">
            <span class="text-meta">Trạng thái duyệt</span>
            <span
              class="inline-flex items-center gap-1.5 rounded-sm px-2.5 py-1 text-xs font-medium"
              :style="statusStyle(shop.status)"
            >
              <span class="h-1.5 w-1.5 rounded-full" style="background: currentColor"></span>
              {{ statusLabel(shop.status) }}
            </span>
          </div>

          <FormField v-model="form.name" label="Tên shop" required />
          <FormField v-model="form.phone" label="Số điện thoại" required />
          <LocationCascader
            v-model="form.location_id"
            label="Khu vực lấy hàng (Huyện/Xã)"
            hint="Để tối ưu việc tạo vận đơn, hãy chọn đúng khu vực."
          />
          <FormField v-model="form.address_detail" label="Địa chỉ lấy hàng chi tiết (Số nhà, đường)" required />
          <div>
            <div class="text-sm font-medium mb-1">Định vị lấy hàng trên bản đồ</div>
            <div class="text-xs text-meta mb-3">Tùy chọn. Nếu ghim đúng vị trí, tài xế sẽ lấy hàng nhanh chóng hơn.</div>
            <LocationPickerMap v-model="form.coordinates" />
          </div>
          <FormField
            v-model="form.webhook_url"
            label="Webhook URL (tùy chọn)"
            placeholder="https://shop.com/webhook"
            hint="Nhận thông báo mỗi khi đơn đổi trạng thái. Để trống nếu không dùng."
          />
        </div>
      </BaseCard>

      <BaseCard title="Bảo mật & Tích hợp">
        <div class="space-y-4">
          <div class="flex items-start justify-between flex-wrap gap-4">
            <div>
              <div class="text-sm font-medium text-strong">API Key</div>
              <p class="text-sm text-meta mt-1">Dùng để kết nối hệ thống của bạn với Ocean Express qua API.</p>
            </div>
            <BaseButton variant="secondary" size="sm" @click="showKeyModal = true">
              <Key class="w-4 h-4" /> Tạo lại API Key
            </BaseButton>
          </div>
        </div>
      </BaseCard>

      <div class="flex justify-end">
        <BaseButton variant="primary" :loading="isSaving" @click="submit">
          <Save class="w-4 h-4" /> Lưu thay đổi
        </BaseButton>
      </div>
    </template>

    <div v-else class="py-16 text-center">
      <PackageX class="w-12 h-12 mx-auto text-meta/40 mb-3" />
      <p class="text-meta">Không tải được thông tin tài khoản.</p>
    </div>

    <!-- Modal tạo lại API Key -->
    <BaseModal v-model="showKeyModal" title="Tạo lại API Key" :close-on-backdrop="false">
      <div v-if="!createdApiKey" class="space-y-4">
        <div class="flex items-start gap-2 px-3 py-2.5 rounded-md text-sm bg-warning/10 text-[var(--warning)]">
          <TriangleAlert class="w-4 h-4 mt-0.5 shrink-0" />
          <span>Khi tạo API Key mới, <b>Key cũ sẽ bị vô hiệu hóa ngay lập tức</b>. Các tích hợp đang dùng Key cũ sẽ ngừng hoạt động.</span>
        </div>

        <!-- Bước 1: Yêu cầu OTP -->
        <div class="space-y-2">
          <div class="text-sm font-medium text-strong">Bước 1 — Xác nhập Email</div>
          <p class="text-xs text-meta">Nhấn nút bên dưới để nhận mã OTP vào hộp thư <b>{{ shop?.email }}</b>.</p>
          <BaseButton variant="secondary" size="sm" :loading="isRequestingOTP" :disabled="otpCountdown > 0" @click="requestOTP">
            <span v-if="otpCountdown > 0">Đợi {{ otpCountdown }}s để gửi lại</span>
            <span v-else>Gửi mã OTP qua Email</span>
          </BaseButton>
          <div v-if="otpSent" class="text-xs text-[var(--success)] flex items-center gap-1">
            <Check class="w-3 h-3" /> Mã OTP đã gửi! Kiểm tra hộp thư của bạn.
          </div>
        </div>

        <!-- Bước 2: Nhập OTP + Mật khẩu -->
        <div class="space-y-3">
          <div class="text-sm font-medium text-strong">Bước 2 — Xác thực</div>
          <FormField v-model="otpCode" label="Mã OTP (6 số)" placeholder="______" maxlength="6" />
          <FormField v-model="confirmPassword" type="password" label="Mật khẩu xác nhập" placeholder="Nhập mật khẩu của bạn" />
        </div>
      </div>
      <div v-else class="space-y-4">
        <div class="flex items-start gap-2 px-3 py-2.5 rounded-md text-sm bg-success/10 text-[var(--success)]">
          <Check class="w-4 h-4 mt-0.5 shrink-0" />
          <span>API Key mới đã được tạo thành công! Lưu ý: Key sẽ <b>không hiển thị lại</b>.</span>
        </div>
        <div class="bg-subtle border rounded-[var(--r-md)] p-3 font-mono text-sm text-strong break-all select-all">
          {{ createdApiKey }}
        </div>
      </div>
      <template #footer>
        <template v-if="!createdApiKey">
          <BaseButton variant="secondary" @click="closeKeyModal">Hủy bỏ</BaseButton>
          <BaseButton variant="danger" :loading="isRegenerating" :disabled="!otpCode || !confirmPassword" @click="regenerateApiKey">Xác nhập tạo mới</BaseButton>
        </template>
        <template v-else>
          <BaseButton variant="secondary" @click="copyKey">
            <Check v-if="copied" class="w-4 h-4" /><Copy v-else class="w-4 h-4" />
            {{ copied ? 'Đã copy' : 'Copy Key' }}
          </BaseButton>
          <BaseButton variant="primary" @click="closeKeyModal">Đã lưu, đóng</BaseButton>
        </template>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import api from '../../services/api';
import { useToastStore } from '../../stores/toastStore';
import { PackageX, TriangleAlert, Save, Key, Check, Copy } from 'lucide-vue-next';
import BaseCard from '../../components/ui/BaseCard.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import BaseModal from '../../components/ui/BaseModal.vue';
import FormField from '../../components/ui/FormField.vue';
import LocationCascader from '../../components/ui/LocationCascader.vue';
import LocationPickerMap from '../../components/LocationPickerMap.vue';

const toast = useToastStore();

const shop = ref(null);
const locations = ref([]);
const loading = ref(true);
const isSaving = ref(false);

const form = ref({
  name: '',
  phone: '',
  webhook_url: '',
  location_id: '',
  address_detail: '',
  coordinates: null
});

const showKeyModal = ref(false);
const isRegenerating = ref(false);
const isRequestingOTP = ref(false);
const otpCode = ref('');
const confirmPassword = ref('');
const createdApiKey = ref('');
const copied = ref(false);
const otpSent = ref(false);
const otpCountdown = ref(0);
let countdownTimer = null;

const fetchMe = async () => {
  loading.value = true;
  try {
    const [meRes] = await Promise.all([
      api.get('/shops/me')
    ]);
    if (meRes.success) {
      shop.value = meRes.data;
      form.value = {
        name: meRes.data.name || '',
        phone: meRes.data.phone || '',
        location_id: meRes.data.location_id || '',
        address_detail: meRes.data.address_detail || '',
        webhook_url: meRes.data.webhook_url || '',
        coordinates: meRes.data.latitude && meRes.data.longitude 
          ? { lat: meRes.data.latitude, lng: meRes.data.longitude }
          : null
      };
    }
  } catch (error) {
    toast.error('Không thể tải thông tin tài khoản');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

onMounted(fetchMe);

const submit = async () => {
  if (!form.value.name || !form.value.address_detail || !form.value.phone) {
    toast.warning('Vui lòng điền tên shop, số điện thoại và địa chỉ lấy hàng');
    return;
  }
  isSaving.value = true;
  try {
    const payload = {
      name: form.value.name,
      phone: form.value.phone,
      webhook_url: form.value.webhook_url,
      location_id: form.value.location_id || null,
      address_detail: form.value.address_detail,
      latitude: form.value.coordinates?.lat || null,
      longitude: form.value.coordinates?.lng || null,
    };
    const res = await api.put('/shops/me', payload);
    if (res.success) {
      shop.value = res.data;
      toast.success('Đã lưu thông tin tài khoản');
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Lưu thất bại');
  } finally {
    isSaving.value = false;
  }
};

const statusLabel = (s) => ({ pending: 'Chờ duyệt', approved: 'Đã duyệt', rejected: 'Bị từ chối' }[s] || s);
const statusStyle = (s) => {
  if (s === 'approved') return { color: 'var(--st-delivered-fg)', background: 'var(--st-delivered-bg)' };
  if (s === 'rejected') return { color: 'var(--st-returned-fg)', background: 'var(--st-returned-bg)' };
  return { color: 'var(--st-ready-fg)', background: 'var(--st-ready-bg)' };
};

const closeKeyModal = () => {
  showKeyModal.value = false;
  confirmPassword.value = '';
  otpCode.value = '';
  createdApiKey.value = '';
  copied.value = false;
  otpSent.value = false;
  otpCountdown.value = 0;
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; }
};

const requestOTP = async () => {
  isRequestingOTP.value = true;
  try {
    const res = await api.post('/shops/me/api-key/request-otp', {});
    if (res.success) {
      otpSent.value = true;
      otpCountdown.value = 60;
      countdownTimer = setInterval(() => {
        otpCountdown.value--;
        if (otpCountdown.value <= 0) { clearInterval(countdownTimer); countdownTimer = null; }
      }, 1000);
      toast.success('Mã OTP đã gửi vào email!');
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Gửi OTP thất bại');
  } finally {
    isRequestingOTP.value = false;
  }
};

const regenerateApiKey = async () => {
  if (!otpCode.value) {
    toast.warning('Vui lòng nhập mã OTP đã được gửi tới email');
    return;
  }
  if (!confirmPassword.value) {
    toast.warning('Vui lòng nhập mật khẩu xác nhập');
    return;
  }
  isRegenerating.value = true;
  try {
    const res = await api.post('/shops/me/api-key', { password: confirmPassword.value, otp: otpCode.value });
    if (res.success && res.data.api_key) {
      createdApiKey.value = res.data.api_key;
    }
  } catch (error) {
    toast.error(error.response?.data?.error?.message || error.response?.data?.error || 'Xác thực thất bại');
  } finally {
    isRegenerating.value = false;
  }
};

const copyKey = async () => {
  try {
    await navigator.clipboard.writeText(createdApiKey.value);
    copied.value = true;
  } catch {
    // ignore
  }
};
</script>
