<template>
  <el-dropdown trigger="click" @command="handleLanguageChange">
    <div class="i-svg:language" :class="size" />
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="item in languageStore.availableLanguages"
          :key="item.code"
          :disabled="languageStore.currentLanguage === item.code"
          :command="item.code"
        >
          {{ item.name }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { useLanguageStore } from "@/stores";

defineProps({
  size: {
    type: String,
    required: false,
  },
});

const languageStore = useLanguageStore();
const { locale, t } = useI18n();

/**
 * 处理语言切换
 *
 * @param lang  语言（zh-cn、en-US）
 */
async function handleLanguageChange(lang: SupportedLanguagesType) {
  await languageStore.setLanguage(lang);
  locale.value = lang;

  ElMessage.success(t("common.langSelect.message.success"));
}
</script>
