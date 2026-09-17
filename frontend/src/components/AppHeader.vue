<template>
  <header class="app-header">
    <div>
      <p class="eyebrow">Skill Swap Campus</p>
      <h1>{{ APP_NAME }}</h1>
      <p>{{ APP_TAGLINE }}</p>
    </div>
    <el-badge :value="unread" class="message-badge">
      <el-button type="primary">站内消息</el-button>
    </el-badge>
  </header>
</template>

<script setup lang="ts">
import { APP_NAME, APP_TAGLINE } from '../constants/app.constants';

defineProps<{ unread: number }>();
</script>
