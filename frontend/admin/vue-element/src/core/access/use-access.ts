import { computed } from "vue";

import { preferences, updatePreferences } from "@/core/preferences";
import { useAccessStore, useAppUserStore } from "@/stores";

function useAccess() {
  const accessStore = useAccessStore();
  const userStore = useAppUserStore();
  const accessMode = computed(() => {
    return preferences.app.accessMode;
  });

  /**
   * 基于角色判断是否有权限
   * @description: Determine whether there is permission，The role is judged by the user's role
   * @param roles
   */
  function hasAccessByRoles(roles: string[]) {
    const userRoleSet = new Set(userStore.userRoles);
    const intersection = roles.filter((item) => userRoleSet.has(item));
    return intersection.length > 0;
  }

  /**
   * 基于权限码判断是否有权限
   * @description: Determine whether there is permission，The permission code is judged by the user's permission code
   * @param codes
   */
  function hasAccessByCodes(codes: string[]) {
    const userCodesSet = new Set(accessStore.accessCodes);

    // 检查用户是否拥有超级管理员权限
    if (userCodesSet.has("*:*:*")) {
      return true;
    }

    // 检查是否有精确匹配的权限码，或所需的权限码是 *:*:*
    const exactMatch = codes.filter((item) => item === "*:*:*" || userCodesSet.has(item));
    if (exactMatch.length > 0) {
      return true;
    }

    // 检查前缀匹配：如果用户拥有某个权限前缀，则拥有所有子权限
    // 例如：用户有 sys:manage_tenants，则拥有 sys:manage_tenants:add, sys:manage_tenants:edit 等
    for (const requiredCode of codes) {
      // 如果需要的权限是 *:*:*，直接通过
      if (requiredCode === "*:*:*") {
        return true;
      }
      for (const userCode of userCodesSet) {
        // 如果用户权限码是所需权限码的前缀（加冒号），则认为有权限
        if (requiredCode.startsWith(userCode + ":")) {
          return true;
        }
      }
    }

    return false;
  }

  async function toggleAccessMode() {
    updatePreferences({
      app: {
        accessMode: preferences.app.accessMode === "frontend" ? "backend" : "frontend",
      },
    });
  }

  return {
    accessMode,
    hasAccessByCodes,
    hasAccessByRoles,
    toggleAccessMode,
  };
}

export { useAccess };
