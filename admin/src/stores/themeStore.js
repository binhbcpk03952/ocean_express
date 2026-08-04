import { defineStore } from 'pinia';
import { ref } from 'vue';

const STORAGE_KEY = 'ocean_theme';

export const useThemeStore = defineStore('theme', () => {
  const initial = localStorage.getItem(STORAGE_KEY)
    || (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');

  const theme = ref(initial);

  const apply = (value) => {
    document.documentElement.classList.toggle('dark', value === 'dark');
    localStorage.setItem(STORAGE_KEY, value);
  };

  apply(theme.value);

  const toggle = () => {
    theme.value = theme.value === 'dark' ? 'light' : 'dark';
    apply(theme.value);
  };

  return { theme, toggle };
});
