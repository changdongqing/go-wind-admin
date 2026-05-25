import { useState, useEffect, type RefObject } from 'react';

/**
 * 动态计算 ProTable scroll.y，使表格体 + 分页器恰好填满容器剩余空间。
 *
 * 测量策略（不穿透 antd 内部 DOM）：
 *   1. 容器高度 = container.clientHeight
 *   2. 逐个测量各组件的 offsetHeight + marginTop + marginBottom
 *   3. scroll.y = 容器高度 - 所有组件高度之和 - 安全边距
 *
 * @param containerRef - .page-container-content 容器 div 的 ref
 * @param options.buffer - 安全边距（像素），默认 4
 * @param options.minHeight - scroll.y 最小值，默认 100
 */
export function useProTableScrollY(
  containerRef: RefObject<HTMLElement | null>,
  options: { buffer?: number; minHeight?: number } = {},
): string {
  const { buffer = 4, minHeight = 100 } = options;

  // 初始值必须是像素值，触发 antd 创建 .ant-table-body
  const [scrollY, setScrollY] = useState<string>(() => {
    const estimate = window.innerHeight - 380;
    return `${Math.max(estimate, minHeight)}px`;
  });

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    let frameId = 0;

    const measure = () => {
      const containerHeight = container.clientHeight;
      if (containerHeight <= 0) return;

      // 逐个测量各组件高度（含 margin）
      let usedHeight = 0;

      const addElement = (selector: string) => {
        const el = container.querySelector(selector) as HTMLElement | null;
        if (!el) return;
        const style = getComputedStyle(el);
        usedHeight +=
          el.offsetHeight +
          (parseFloat(style.marginTop) || 0) +
          (parseFloat(style.marginBottom) || 0);
      };

      // 搜索表单
      addElement('.ant-pro-table-search');
      // 工具栏
      addElement('.ant-pro-table-list-toolbar');
      // 表头
      addElement('.ant-table-thead');
      // 分页器
      addElement('.ant-pagination');

      const result = Math.max(containerHeight - usedHeight - buffer, minHeight);
      const newValue = `${result}px`;

      // 避免 setState 循环：只在值变化时更新
      setScrollY((prev) => {
        if (prev === newValue) return prev;
        return newValue;
      });
    };

    // 防抖：用 raf 合并快速变化
    const scheduleMeasure = () => {
      cancelAnimationFrame(frameId);
      frameId = requestAnimationFrame(measure);
    };

    // 多次重试，确保 ProTable 异步渲染完成
    const timers = [
      setTimeout(measure, 50),
      setTimeout(measure, 200),
      setTimeout(measure, 500),
      setTimeout(measure, 1500),
    ];

    const resizeObserver = new ResizeObserver(scheduleMeasure);
    resizeObserver.observe(container);

    const mutationObserver = new MutationObserver(scheduleMeasure);
    mutationObserver.observe(container, { childList: true, subtree: true });

    return () => {
      cancelAnimationFrame(frameId);
      timers.forEach(clearTimeout);
      resizeObserver.disconnect();
      mutationObserver.disconnect();
    };
  }, [containerRef, buffer, minHeight]);

  return scrollY;
}
