import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',
    port: 18629,
    proxy: {
      '/api': 'http://localhost:19629',
    },
  },
  preview: { host: '0.0.0.0', port: 18629 },
});
