<template>
  <ECharts :options="chartOptions" />
</template>

<script lang="ts" setup>
import ECharts from "@/components/ECharts/index.vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const chartOptions = computed(() => ({
  grid: {
    bottom: 0,
    containLabel: true,
    left: "1%",
    right: "1%",
    top: "2%",
  },
  series: [
    {
      barMaxWidth: 80,
      data: [3000, 2000, 3333, 5000, 3200, 4200, 3200, 2100, 3000, 5100, 6000, 3200, 4800],
      type: "bar",
    },
  ],
  tooltip: {
    axisPointer: {
      lineStyle: {
        width: 1,
      },
    },
    trigger: "axis",
  },
  xAxis: {
    data: Array.from({ length: 12 }).map(
      (_item, index) => `${index + 1}${t("pages.dashboard.month")}`
    ),
    type: "category",
  },
  yAxis: {
    max: 8000,
    splitNumber: 4,
    type: "value",
  },
}));
</script>
