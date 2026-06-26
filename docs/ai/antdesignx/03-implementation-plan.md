# 三、实施计划

> 配套：`README.md`（总览）、`01-ui-layout.md`（界面）、`02-cli-connection.md`（连接）。
>
> 本章给出落地路径、改动清单、依赖、风险与验收。**本文档仅为蓝图，对应代码尚未开发。**

---

## 1. 分阶段路线（P0 → P6）

| 阶段 | 目标 | 关键产出 | 验收（DoD） |
|------|------|---------|------------|
| **P0 打通** | 证明 IPC spawn → 流式回推链路；引入 Ant Design X | `registerAgentIpc` + `ClaudeCodeProvider` 最小实现；渲染层裸 `Sender` 收到流式文本 | 点「发送」→ 看到 `claude` 文本逐字出现 |
| **P1 MVP 对话** | 单 Provider（Claude Code）完整对话 | `BubbleList` + `Sender` + `useXChat`；抽屉形态 A；`tool.*` 渲染 `ThoughtChain` | 多轮对话、工具调用折叠展示、中断可用 |
| **P2 会话与多 Provider** | 会话管理 + OpenCode | `Conversations`；`OpenCodeProvider`；Provider 切换；会话持久化/恢复（`--resume`） | 切 Provider 不报错；重启后历史会话可见 |
| **P3 权限与附件** | 交互式权限 + 文件上下文 | 权限请求卡片 + stdin 回写；`Attachments`；工作目录选择 | 敏感工具弹卡询问；附件作为上下文送出 |
| **P4 形态完整 + 体验** | 分栏 / 全屏 + 主题 + i18n | 形态 B/C；`PreferencesPanel`「AI Agent」分区；亮/暗 + zh/en；快捷键 | 三形态可切；暗色代码块正常；全英文可跑 |
| **P5 B/S 远程 Agent**（可选） | 浏览器降级 | `RemoteAgentAdapter` + 后端 SSE 端点（需后端配合） | B/S 下 Agent 入口可用（远程模型） |
| **P6 打磨** | 健壮性 | 断流重连、错误兜底、超时、孤儿进程清理审计、性能（长会话虚拟滚动） | 异常注入全过；长会话不卡 |

> 🔄 **v2 评审修订（对比 open-design）—— 解耦「传输」与「形态」，补本地 BYOK 传输**：
> 原方案把「desktop↔本地 CLI spawn」与「browser↔远程 agent」绑成两轴。open-design 更优：同一 daemon 既 spawn CLI 又跑 OpenAI 兼容 BYOK proxy，两路吐同一种事件流，于是「没装 CLI」与「浏览器形态」是两个独立问题。本方案据此新增**第三种传输：本地 BYOK SSE proxy**（主进程内挂极小 OpenAI 兼容端点，`127.0.0.1` only + 鉴权，凭证经 `safeStorage`，复用 `02-cli-connection.md` §8 安全模型）：
> - 没装任何 CLI 时，渲染层 `LocalByokAdapter` 走它，吐同样的 `AgentEvent`，体验不降级；
> - B/S 浏览器形态可复用同一管线（同机），P5「远程后端 agent」降级为「第三种传输」而非唯一兜底。
>
> 排期：**P3 之后、P5 之前**插入（spawn 主链路稳定再做），不阻断 P0–P2；未做前「无 CLI」仍走入口灰显 + 安装提示。文件见 §3.1 `byok-proxy.ts`、§3.2 `LocalByokAdapter`。

> 顺序原则：**P0 先证伪技术风险**（spawn + 流式 + Ant Design X 兼容性 + **交互式权限语义**），再迭代 UI。P0 不通过不进 P1。

> 🔄 **v2 评审修订 —— P0 新增两个必验项**：
> 1. **交互式权限 over stdio**：Claude Code `-p --input-format stream-json` 下，遇到需权限工具是「emit 请求并阻塞等 stdin 应答」还是「直接拒绝返回 `tool_result`」？结论决定 P3 权限 UX 走交互卡 or 预批准白名单/沙箱（见 `02-cli-connection.md` §4.1）。**go/no-go 关卡之一。**
> 2. **Windows argv 上限**：长 prompt 拼 `-p` 是否触发 `ENAMETOOLONG`，stdin / 临时文件回退是否生效（GoWind Windows 优先，必测）。

---

## 2. 依赖清单

### 2.1 新增前端依赖

| 包 | 用途 | 备注 |
|----|------|------|
| `@ant-design/x` | AI 对话组件 | **v2**；落地前确认与 `antd@6` / `react@19` 兼容（v2 有破坏性变更，见 README §2.1） |
| `@ant-design/x-sdk` | `useXChat` / `XStream` 运行时工具 | v2 起运行时工具独立成包 |
| `react-resizable-panels`（P4 候选） | 分栏拖拽分隔条 | 也可自实现，权衡体积后定 |

### 2.2 主进程侧

- 复用 Node 内置 `child_process`、Electron `safeStorage`、`dialog` —— **零新增原生依赖**（与 `desktop-client-plan.md`「外设生态」一致的克制原则）。
- `which`（解析 CLI 可执行路径）—— 轻量纯 JS，可选。

### 2.3 CLI（用户机器，非项目依赖）

- `claude`（Claude Code CLI）—— 默认 Provider，用户自装。
- `opencode`（OpenCode CLI）—— 次选，P2 起。

> 本方案不打包、不分发这些 CLI；`probe()` 仅探测并提示用户安装。

### 2.4 兼容性矩阵（P0 必须先验）

| 维度 | 验证点 |
|------|--------|
| `@ant-design/x@2` × `antd@6` | 组件渲染、`ConfigProvider` 主题透传、`algorithm` 联动 |
| `@ant-design/x@2` × `react@19` | 无 peer 警告、并发特性可用 |
| `useXChat` 接 IPC 异步流 | `XStream` 消费自定义 AsyncIterable 的可行性 |
| `claude --output-format stream-json` | 本机版本下 NDJSON 字段实测（`type` 枚举以实测为准） |
| **Claude Code headless 权限语义**（v2 修订） | `-p --input-format stream-json` 下需权限工具：emit 请求等 stdin 应答 vs 直接拒绝？结论定 P3 权限 UX（见 `02-cli-connection.md` §4.1） |
| **Windows argv 上限**（v2 修订） | 长 prompt 拼 `-p` 是否 `ENAMETOOLONG`；stdin / 临时文件回退是否生效 |

---

## 3. 文件清单与改动点

> 遵循项目「前端只用 React」「禁止改 generated/」约定；遵循 Conventional Commits。

### 3.1 桌面端（Electron 主进程 / preload）

| 文件 | 类型 | 内容 |
|------|------|------|
| `desktop/electron/main/ipc/agent.ts` | 新增 | `registerAgentIpc` + `AgentProvider` 接口 + `ClaudeCodeProvider`/`OpenCodeProvider` |
| `desktop/electron/main/agent/byok-proxy.ts` | 新增（v2 修订） | 本地 OpenAI 兼容 SSE proxy（`127.0.0.1` only + 鉴权），`LocalByokAdapter` 的服务端；详见 §1 v2 修订 |
| `desktop/electron/main/ipc/index.ts` | 改 | `registerAllIpc()` 加 `registerAgentIpc()` |
| `desktop/electron/main/index.ts` | 改 | `before-quit` 钩子清理活跃 CLI 子进程（防孤儿） |
| `desktop/electron/preload/index.ts` | 改 | `desktop.agent.*` 白名单方法（见连接方案 §10） |
| `react/types/desktop.d.ts` | 改 | 同步 `DesktopBridge.agent` 类型 |

### 3.2 React 前端

| 文件 | 类型 | 内容 |
|------|------|------|
| `src/core/agent/adapter.ts` | 新增 | `AgentAdapter` 接口 + `DesktopAgentAdapter` + `LocalByokAdapter`（v2 修订）+ P5 `RemoteAgentAdapter` |
| `src/core/agent/events.ts` | 新增 | 归一化事件 `AgentEvent` 类型（与主进程共享 schema） |
| `src/stores/agent.ts` | 新增 | `useAgentStore`（会话/消息/连接态，持久化） |
| `src/pages/.../agent/AgentConversation.tsx` | 新增 | 三形态共享对话组件（BubbleList+Sender+ThoughtChain） |
| `src/pages/.../agent/AgentDock.tsx` | 新增 | 抽屉(A)/分栏(B) 容器 |
| `src/pages/.../agent/AgentWorkbench.tsx` | 新增 | 全屏(C) 会话工作台 |
| `src/router/modules/agent.tsx` | 新增 | `/ai/chat` 路由（自动导入），`meta.authority: ['sys:agent:view']` |
| `src/layouts/MainLayout/index.tsx` | 改 | 挂载 `<AgentDock />` |
| `src/layouts/MainLayout/components/HeaderContent.tsx` | 改 | widget 区 `✨` Agent 按钮 |
| `src/core/preferences/`（types/default/panel） | 改 | `widget.aiAssistant`、`agent.*`（默认形态/默认 Provider/权限模式） |
| `src/locales/{zh-CN,en-US}/_modules/agent.json` | 新增 | 国际化 |

### 3.3 文档

| 文件 | 类型 | 内容 |
|------|------|------|
| `docs/caozuo/desktop-optional-config.md` | 改 | 第四章能力清单补「AI Agent（本地 CLI）」一项（启停约定） |
| `docs/ai/antdesignx/*` | 本套 | 已完成（本目录） |

> 新增权限码：`sys:agent:view`（访问）、`sys:agent:manage`（改默认 Provider/权限设置）。超管 `*:*:*` 自动放行（遵循 `frontend_authority.md`）。

---

## 4. 风险与缓解

| 风险 | 等级 | 缓解 |
|------|------|------|
| `@ant-design/x@2` 与 antd6/react19 不兼容 | 🔴 高 | P0 第一件事验证兼容矩阵；不兼容则评估锁 v1 或等修，**这是 go/no-go 关卡** |
| **交互式权限 over stdio 不成立**（v2 修订） | 🔴 高 | P0 实测；不成立则走「沙箱+bypass」或「预批准白名单」（见 `02-cli-connection.md` §4.1、`04-comparison-with-opendesign.md`） |
| Claude Code NDJSON 字段随版本变 | 🟡 中 | 归一化层做容错（未知 `type` 转 `log`，不崩）；以本机实测版本为准 |
| 孤儿 CLI 进程 | 🟡 中 | `before-quit` + 窗口关闭钩子统一清理；启动时登记活跃句柄 |
| 凭证泄露（命令行/日志） | 🟡 中 | env 注入 + safeStorage；stderr 脱敏；禁 shell |
| 长会话性能（消息多/大 diff） | 🟢 低 | BubbleList 虚拟滚动；diff 折叠；超长结果懒加载 |
| 用户未装 CLI | 🟢 低 | `probe()` 探测 + 入口灰显 + 安装提示 |
| B/S 期望落空（纯浏览器无本地能力） | 🟢 低 | 文档明确降级；入口按 `isDesktop` 隐藏/切远程 |
| OpenCode SDK 文档滞后 | 🟢 低 | 直连 HTTP+SSE，不依赖 SDK |

---

## 5. 回退策略

- **能力可关**：Agent IPC 走 `docs/caozuo/desktop-optional-config.md` 的「能力启停」约定，默认可在偏好设置一键关闭，关闭后 `desktop.agent` 不注册、入口隐藏，等同未集成。
- **渐进发布**：P0–P1 可只在开发模式（`isDev`）暴露，稳定后再放开。
- **不影响 B/S 与既有业务**：所有改动经可选链/运行时检测兜底，Agent 子系统异常不应阻断业务页面（Agent 组件挂载失败用 Error Boundary 兜底，记录但不崩主界面）。

---

## 6. 验收标准（端到端）

1. 桌面端启动，Header 出现 `✨` Agent 按钮；点击展开右侧抽屉，显示 `Welcome` + `Prompts`。
2. 本机已装 `claude` 时，输入「在 `src/api/hooks/` 下找所有 List hook」→ 看到流式文本回复，`ThoughtChain` 展示 `Read`/`Grep` 等工具调用步骤与结果摘要。
3. 中途点「中断」→ 文本停止、出现「已中断」系统消息，无残留 `claude` 进程。
4. 切到 OpenCode Provider，同样能收发；历史会话在 `Conversations` 列表可见，重启应用后仍在。
5. 触发一次写操作 → **按 P0 权限探测结论**：若交互式权限成立，出现权限请求卡片 → 选「拒绝」→ CLI 收到拒绝、流程正确中止；若不成立走预批准白名单/沙箱，则验证高危工具被挡在白名单外、目录外写入被拒（见 `02-cli-connection.md` §4.1）。
6. 切换亮/暗主题、中/英文，Agent 面板样式与文案正确。
7. 抽屉 → 分栏 → 全屏三态切换，对话内容连续不丢。
8. 关闭窗口/退出应用，任务管理器中无 `claude`/`opencode` 孤儿进程。
9. B/S（纯浏览器）下 Agent 入口默认隐藏（或走远程适配器，若 P5 已交付）。
10. `probe()` 在未装 CLI 时入口灰显并给出安装提示。
