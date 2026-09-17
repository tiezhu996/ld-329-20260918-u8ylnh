<template>
  <div ref="chartRef" class="radar-chart" />
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import * as echarts from 'echarts';

const props = defineProps<{ radar: Record<string, number> }>();
const chartRef = ref<HTMLDivElement | null>(null);

function renderChart() {
  if (!chartRef.value) return;
  const chart = echarts.init(chartRef.value);
  const names = Object.keys(props.radar);
  chart.setOption({
    radar: { indicator: names.map((name) => ({ name, max: 100 })) },
    series: [{ type: 'radar', data: [{ value: names.map((name) => props.radar[name]), name: '技能熟练度' }], areaStyle: {} }],
  });
}

onMounted(renderChart);
watch(() => props.radar, renderChart);
</script>
