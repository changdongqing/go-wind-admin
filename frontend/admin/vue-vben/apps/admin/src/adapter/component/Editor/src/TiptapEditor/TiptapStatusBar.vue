<script setup lang="ts">
defineProps<{
  isDark: boolean;
  statusInfo: { chars: number; cursor: string; words: number };
}>();
</script>

<template>
  <div
    class="pointer-events-none absolute bottom-3 right-4 flex select-none items-center gap-2 text-xs"
    :class="isDark ? 'text-[#626773]' : 'text-[#94a3b8]'"
  >
    <span>{{ statusInfo.words }} {{ $t('common.editor.words') }}</span>
    <span
      :class="isDark ? 'text-[var(--tte-toolbar-border)]' : 'text-[#cbd5e1]'"
      >|</span
    >
    <span>{{ statusInfo.chars }} {{ $t('common.editor.chars') }}</span>
    <span
      :class="isDark ? 'text-[var(--tte-toolbar-border)]' : 'text-[#cbd5e1]'"
      >|</span
    >
    <span>Ln {{ statusInfo.cursor }}</span>
    <span
      class="rounded px-2 py-0.5 text-[11px]"
      :class="
        isDark
          ? 'bg-[rgba(255,255,255,0.06)] text-[#626773]'
          : 'bg-[#f1f5f9] text-[#94a3b8]'
      "
    >
      MD
    </span>
  </div>
</template>
