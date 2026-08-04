<template>
  <div class="flex min-h-screen flex-col bg-base text-body">
    <!-- Top bar -->
    <header class="sticky top-0 z-10 flex h-16 items-center justify-between border-b bg-surface/95 px-4 sm:px-8 backdrop-blur-md">
      <div class="flex items-center gap-8">
        <!-- Brand -->
        <div class="flex items-center gap-3 min-w-0">
          <div
            class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg shadow-sm"
            style="background: var(--brand-gradient)"
          >
            <svg class="h-5 w-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 16.5 12 21l9-4.5" /><path d="M3 12l9 4.5L21 12" /><path d="M12 3 3 7.5 12 12l9-4.5L12 3z" />
            </svg>
          </div>
          <div class="min-w-0 leading-tight hidden sm:block">
            <p class="truncate text-sm font-semibold text-strong">{{ authStore.user?.name || 'Nhân viên' }}</p>
            <p class="truncate text-xs text-meta">{{ roleLabel }}</p>
          </div>
        </div>

        <!-- Desktop Nav -->
        <nav class="hidden md:flex items-center gap-2">
          <router-link
            v-for="tab in tabs"
            :key="tab.to"
            :to="tab.to"
            class="flex items-center gap-2 px-3.5 py-2 rounded-md text-sm font-medium transition-colors"
            :class="isActive(tab) ? 'bg-[var(--primary)] text-white shadow-sm' : 'text-meta hover:bg-subtle hover:text-body'"
          >
            <component :is="tab.icon" class="h-4 w-4" />
            <span>{{ tab.label }}</span>
          </router-link>
        </nav>
      </div>

      <!-- Right actions -->
      <div class="flex items-center gap-2">
        <ThemeToggle />
        <button class="rounded-md p-2 text-meta hover:bg-danger/10 hover:text-danger transition-colors" @click="handleLogout" aria-label="Đăng xuất">
          <LogOut class="h-5 w-5" />
        </button>
      </div>
    </header>

    <!-- Content -->
    <main class="flex-1 w-full max-w-7xl mx-auto p-4 sm:p-6 lg:p-8 pb-24 md:pb-8">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>

    <!-- Bottom tab bar (Mobile only) -->
    <nav class="fixed inset-x-0 bottom-0 z-20 flex h-[68px] items-stretch border-t bg-surface/95 backdrop-blur-md md:hidden pb-safe">
      <router-link
        v-for="tab in tabs"
        :key="tab.to"
        :to="tab.to"
        class="flex flex-1 flex-col items-center justify-center gap-1 text-[11px] font-medium transition-colors"
        :class="isActive(tab) ? 'text-[var(--primary)]' : 'text-meta'"
      >
        <component :is="tab.icon" class="h-[22px] w-[22px]" :class="isActive(tab) ? 'scale-110 transition-transform' : ''" />
        <span>{{ tab.label }}</span>
      </router-link>
    </nav>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '../stores/authStore';
import ThemeToggle from './ui/ThemeToggle.vue';
import { ListChecks, ScanLine, LogOut, BarChart2 } from 'lucide-vue-next';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const roleLabels = {
  hub_staff: 'Nhân viên kho',
  first_mile_driver: 'Tài xế lấy hàng',
  last_mile_driver: 'Tài xế giao hàng',
};
const roleLabel = computed(() => roleLabels[authStore.user?.role] || authStore.user?.role || '');

const tabs = computed(() => {
  const base = [
    { to: '/m', name: 'MemberTasks', label: 'Nhiệm vụ', icon: ListChecks },
    { to: '/m/stats', name: 'MemberStats', label: 'Thống kê', icon: BarChart2 }
  ];
  if (authStore.user?.role === 'hub_staff') {
    base.splice(1, 0, { to: '/m/scan', name: 'HubScan', label: 'Quét đơn', icon: ScanLine });
  }
  return base;
});

const isActive = (tab) => {
  if (tab.name === 'MemberTasks') {
    // Tab nhiệm vụ active cả khi đang xem chi tiết đơn
    return route.name === 'MemberTasks' || route.name === 'MemberOrderDetail';
  }
  return route.name === tab.name;
};

const handleLogout = () => {
  authStore.logout();
  router.push({ name: 'Login' });
};
</script>

<style scoped>
.page-enter-active, .page-leave-active { transition: opacity 0.15s ease, transform 0.15s ease; }
.page-enter-from { opacity: 0; transform: translateY(6px); }
.page-leave-to { opacity: 0; transform: translateY(-6px); }
</style>
