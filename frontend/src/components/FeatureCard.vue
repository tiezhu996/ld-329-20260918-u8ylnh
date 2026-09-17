<template>
  <article class="feature-card">
    <div class="feature-card__title">
      <strong>{{ title }}</strong>
      <slot name="tag" />
    </div>
    <p>{{ description }}</p>
    <slot />
  </article>
</template>

<script setup lang="ts">
defineProps<{ title: string; description: string }>();
</script>
