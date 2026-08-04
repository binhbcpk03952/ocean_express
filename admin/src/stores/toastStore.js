import { defineStore } from 'pinia';
import { ref } from 'vue';

let seq = 0;

export const useToastStore = defineStore('toast', () => {
  const toasts = ref([]);

  const push = (type, message, timeout = 3500) => {
    const id = ++seq;
    toasts.value.push({ id, type, message });
    if (timeout > 0) {
      setTimeout(() => dismiss(id), timeout);
    }
    return id;
  };

  const dismiss = (id) => {
    toasts.value = toasts.value.filter((t) => t.id !== id);
  };

  const success = (msg, t) => push('success', msg, t);
  const error = (msg, t) => push('danger', msg, t);
  const info = (msg, t) => push('info', msg, t);
  const warning = (msg, t) => push('warning', msg, t);

  return { toasts, push, dismiss, success, error, info, warning };
});
