import type { ExtendedVxeGridApi, VxeGridProps } from "./types";

import { defineComponent, h, onBeforeUnmount } from "vue";

import { useSelector } from "@tanstack/vue-store";

import { VxeGridApi } from "./api";
import VxeGrid from "./use-vxe-grid.vue";

export function useVxeGrid(options: VxeGridProps) {
  // const IS_REACTIVE = isReactive(options);
  const api = new VxeGridApi(options);
  const extendedApi: ExtendedVxeGridApi = api as ExtendedVxeGridApi;
  extendedApi.useStore = (selector) => {
    return useSelector(api.store, selector);
  };

  const Grid = defineComponent(
    (props: VxeGridProps, { attrs, slots }) => {
      onBeforeUnmount(() => {
        api.unmount();
      });
      api.setState({ ...props, ...attrs } as Partial<VxeGridProps>);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      return () => h(VxeGrid as any, { ...props, ...attrs, api: extendedApi } as any, slots as any) as any;
    },
    {
      name: "VxeGrid",
      inheritAttrs: false,
    },
  );
  // Add reactivity support
  // if (IS_REACTIVE) {
  //   watch(
  //     () => options,
  //     () => {
  //       api.setState(options);
  //     },
  //     { immediate: true },
  //   );
  // }

  return [Grid, extendedApi] as const;
}

export type UseVxeGrid = typeof useVxeGrid;
