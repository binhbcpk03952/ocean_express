<template>
  <div class="relative" ref="bellRef">
    <button @click="toggleDropdown" class="relative p-2 text-gray-500 hover:text-gray-700 transition-colors focus:outline-none">
      <Bell class="w-6 h-6" />
      <span v-if="unreadCount > 0" class="absolute top-1 right-1 flex items-center justify-center w-4 h-4 text-xs font-bold text-white bg-red-500 rounded-full">
        {{ unreadCount > 99 ? '99+' : unreadCount }}
      </span>
    </button>

    <div v-if="isOpen" class="absolute right-0 mt-2 w-80 bg-white rounded-lg shadow-xl overflow-hidden z-50 border border-gray-100">
      <div class="flex items-center justify-between px-4 py-3 bg-gray-50 border-b border-gray-100">
        <h3 class="text-sm font-semibold text-gray-700">Thông báo</h3>
        <button v-if="unreadCount > 0" @click="markAllAsRead" class="text-xs text-blue-600 hover:text-blue-800">
          Đánh dấu tất cả đã đọc
        </button>
      </div>

      <div class="max-h-96 overflow-y-auto">
        <div v-if="loading && notifications.length === 0" class="flex justify-center p-4">
          <Loader2 class="w-5 h-5 animate-spin text-gray-400" />
        </div>
        
        <div v-else-if="notifications.length === 0" class="p-8 text-center text-gray-500">
          <BellOff class="w-8 h-8 mx-auto mb-2 text-gray-300" />
          <p class="text-sm">Chưa có thông báo nào</p>
        </div>

        <div v-else>
          <div v-for="notif in notifications" :key="notif.id" 
               @click="handleRead(notif)"
               class="p-4 border-b border-gray-50 hover:bg-gray-50 cursor-pointer transition-colors"
               :class="{'bg-blue-50/50': !notif.is_read}">
            <div class="flex items-start">
              <div class="flex-shrink-0 mt-1">
                <div class="w-2 h-2 rounded-full mt-1.5" :class="notif.is_read ? 'bg-transparent' : 'bg-blue-500'"></div>
              </div>
              <div class="ml-3 w-full">
                <p class="text-sm font-medium text-gray-800">{{ notif.title }}</p>
                <p class="text-xs text-gray-600 mt-1 line-clamp-2">{{ notif.message }}</p>
                <p class="text-xs text-gray-400 mt-2">{{ formatDate(notif.created_at) }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { Bell, BellOff, Loader2 } from 'lucide-vue-next';
import { storeToRefs } from 'pinia';
import { useNotificationStore } from '../stores/useNotifications';

const store = useNotificationStore();
const { notifications, unreadCount, loading } = storeToRefs(store);

const isOpen = ref(false);
const bellRef = ref(null);

const toggleDropdown = async () => {
  isOpen.value = !isOpen.value;
  if (isOpen.value) {
    await store.fetchNotifications();
  }
};

const handleRead = async (notif) => {
  if (!notif.is_read) {
    await store.markAsRead(notif.id);
  }
};

const markAllAsRead = async () => {
  await store.markAllAsRead();
};

const formatDate = (dateString) => {
  const date = new Date(dateString);
  return date.toLocaleString('vi-VN', {
    hour: '2-digit', minute:'2-digit',
    day: '2-digit', month: '2-digit', year: 'numeric'
  });
};

const closeDropdown = (e) => {
  if (isOpen.value && bellRef.value && !bellRef.value.contains(e.target)) {
    isOpen.value = false;
  }
};

onMounted(() => {
  store.fetchUnreadCount();
  document.addEventListener('click', closeDropdown);
});

onUnmounted(() => {
  document.removeEventListener('click', closeDropdown);
});
</script>
