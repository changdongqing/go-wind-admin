# Ant Design X + 本地 CLI Agent 集成方案（总览）

> 目标读者：开发团队 / 技术决策者
>
> 状态：**方案设计阶段（仅文档，未开发）**。本文档是后续编码实现的蓝图。
>
> 📌 **v2 评审修订（对比 open-design）**：2026-06 评审新增 4 条修订——交互式权限假设、本地 BYOK 传输、Windows argv 防护、事件泵与生命周期解耦，已落入 `02`/`03` 对应章节（标 🔄），溯源与对比见 `04-comparison-with-opendesign.md`。
>
> 关联文档：
> - `docs/desktop/desktop-client-plan.md` —— 桌面客户端（Electron）整体方案，本方案是其上的「能力扩展」
> - `docs/caozuo/desktop-optional-config.md` —— 桌面端 IPC 能力启停约定，本方案新增的 IPC 能力遵循同一约定
> - `frontend/admin/react/CLAUDE.md` —— React 前端数据层与组件规范

---

## 一、背景与目标

### 1.1 背景

GoWind Admin 的桌面客户端已用 Electron 落地（P0–P5，Windows NSIS 安装包已验证产出）。桌面端最大的价值之一，是突破了浏览器沙箱，**可以调度本地进程与本地文件系统**。这条能力一直没有被消费到一个高价值场景上。

与此同时，本地编码 Agent CLI（Claude Code、OpenCode 等）已成为开发者日常工作流的核心。它们能力强大，但交互形态是命令行 / TUI，在企业后台管理这种 GUI 场景里难以直接使用。

### 1.2 目标

在 GoWind Admin（**React 前端 + Electron 桌面壳**）中，集成 [Ant Design X](https://x.ant.design/) 提供的 AI 交互组件，构建一个**图形化的本地 Agent 对话台**，通过桌面端 IPC 拉起本地已安装的 Agent CLI（优先 Claude Code、OpenCode），把它们的流式输出（文本 / 工具调用 / 思考过程）渲染成结构化的对话界面，并支持多轮对话、工具权限确认、会话管理。

效果对标 Cursor / Cline / Continue 这类「在 GUI 里驱动本地 Agent」的产品形态。

### 1.3 非目标（本期不做）

- 不自研模型、不自建推理后端；只做「GUI ↔ 本地 CLI」的桥接。
- 不替换或包装 CLI 自身的能力（不实现 MCP 客户端、不接管权限模型），只做忠实透传与可视化。
- 不在 B/S（纯浏览器）形态下强行接本地 CLI —— 浏览器无法 spawn 进程，B/S 形态降级为「远程 Agent」（后端 SSE）或禁用，详见第四章。

---

## 二、技术选型

### 2.1 UI 层：Ant Design X（`@ant-design/x`）

选定 Ant Design X，理由：

1. **与现有技术栈同源**：本项目 UI 框架是 Ant Design v6，Ant Design X 是官方 AI 场景组件库，主题 / ConfigProvider / 国际化天然打通，无需引入第二套设计系统。
2. **覆盖完整对话语义**：`Bubble` / `BubbleList`（消息流）、`Sender`（输入）、`Conversations`（会话列表）、`ThoughtChain`（工具调用/思考链）、`Prompts`（引导）、`Welcome`（空态）、`Attachments`（附件）、`Suggestion`（补全）—— 一套原子组件即可拼出完整 Agent 对话台。
3. **运行时工具就绪**：`useXChat`（会话状态 Hook）、`XStream` / `XRequest`（流式适配，v2 已迁移至 `@ant-design/x-sdk`）能把任意异步流（包括我们的 IPC 事件流）接进对话状态机。

> ⚠️ **版本注意（v2 破坏性变更）**：Ant Design X 2.0 中，模型调度 Hook `useXAgent` **已被移除**，运行时工具迁移到 `@ant-design/x-sdk`，`useXChat` 升级为唯一的会话管理 Hook，`Bubble` / `Sender` / `ThoughtChain` 有 UI 升级。本方案以 **v2** 为基准设计；落地前在 P0 阶段确认 `@ant-design/x@2.x` 与 `antd@6` / `react@19` 的兼容矩阵（详见 `03-implementation-plan.md` 依赖清单）。

### 2.2 CLI 接入层：Electron 主进程 IPC（主推）

选定「桌面主进程 spawn CLI + IPC 流式回推」作为本地 Agent 的接入主链路，理由：

1. **复用已有架构**：桌面端已建立 `registerXxxIpc` 模式（见 `electron/main/ipc/*.ts`），新增一个 `registerAgentIpc` 即可，零新架构。
2. **能力可达**：主进程是 Node 环境，`child_process.spawn` 可拉起任意本地可执行文件并拿到 stdout / stderr / stdin 全双工流 —— 这是浏览器拿不到的能力，正是桌面端的核心价值。
3. **安全可控**：主进程层可做命令白名单、参数校验、工作目录沙箱、凭证注入，所有「危险操作」集中在一处把关，渲染进程只拿到结构化事件。

### 2.3 备选：本地 Bridge Server（B/S 降级）

纯浏览器（B/S）无法 spawn 进程。降级方案是跑一个本地小服务（独立可执行或复用桌面主进程的 Node 能力）暴露 SSE/WebSocket。**本期仅设计、不实现**，列为 P5，详见 `02-cli-connection.md` 第六章。

---

## 三、总体架构

```
┌──────────────────────────── Electron 渲染进程（React）────────────────────────────┐
│                                                                                      │
│   Ant Design X 组件层                                                                │
│   ┌──────────┐ ┌──────────┐ ┌───────────────┐ ┌────────────┐ ┌────────────┐         │
│   │ Welcome  │ │  Sender  │ │ Bubble/List   │ │ThoughtChain│ │Conversations│        │
│   └────┬─────┘ └────┬─────┘ └───────┬───────┘ └─────┬──────┘ └─────┬──────┘         │
│        │            │               │                │              │               │
│        └────────────┴───────┬───────┴────────────────┴──────────────┘               │
│                              ▼                                                        │
│                    useXChat + XStream（@ant-design/x-sdk）                            │
│                              │                                                        │
│                              ▼                                                        │
│   Agent Adapter 抽象层（渲染侧）                                                       │
│   ┌──────────────────────┐   ┌──────────────────────┐                                 │
│   │ DesktopAgentAdapter  │   │  RemoteAgentAdapter  │ (P5, B/S 降级)                  │
│   │ (window.desktop.agent)│   │  (后端 SSE)          │                                 │
│   └──────────┬───────────┘   └──────────────────────┘                                 │
└──────────────┼───────────────────────────────────────────────────────────────────────┘
               │  contextBridge（preload 白名单）
═══════════════╪═════════════════════════════════════════════════════════════════════════
               ▼  ipcRenderer ↔ ipcMain
┌──────────────────────────── Electron 主进程（Node）──────────────────────────────────┐
│                                                                                      │
│   registerAgentIpc（electron/main/ipc/agent.ts，新增）                                │
│                                                                                      │
│   ┌──────────────── AgentProvider 适配器（主进程侧）─────────────────┐                │
│   │                                                                  │                │
│   │  ClaudeCodeProvider          OpenCodeProvider                   │                │
│   │  spawn('claude', [...])      spawn('opencode', ['serve',..])    │                │
│   │  解析 NDJSON(stream-json)     HTTP + SSE                        │                │
│   │  stdin ← 权限/多轮响应       REST 收发                          │                │
│   │                                                                  │                │
│   └────────────────────────────┬─────────────────────────────────────┘                │
│                                ▼                                                      │
│   归一化事件 → webContents.send('agent:event', ...) 流式回推渲染进程                   │
│                                                                                      │
│   安全：命令/参数白名单 · 工作目录沙箱 · 凭证经 env 注入（不经命令行）· 拒绝 shell:true │
└──────────────────────────────────────────────────────────────────────────────────────┘
               │  child_process.spawn（无 shell）
               ▼
        本机已安装的 Agent CLI（claude / opencode / …）
```

**三段式总结**：
1. **渲染层**：Ant Design X 组件 + `useXChat`/`XStream`，只认「归一化事件流」，不感知具体 CLI。
2. **适配层**：`AgentAdapter` 接口屏蔽「桌面 IPC」与「远程 SSE」两种来源；主进程内 `AgentProvider` 屏蔽不同 CLI 的协议差异。
3. **主进程层**：负责 spawn、流解析、生命周期、安全管控，是唯一接触本机进程与凭证的地方。

---

## 四、运行形态与降级策略

| 形态 | 本地 Agent 能力 | 接入链路 | 优先级 |
|------|----------------|---------|--------|
| **桌面端**（Electron，已落地） | ✅ 完整（spawn 本地 CLI） | 渲染 ↔ IPC ↔ 主进程 spawn | **本期主交付** |
| **B/S 纯浏览器** | ⚠️ 降级 | 仅 `RemoteAgentAdapter`（后端 SSE）或禁用入口 | P5，本期不开发 |

判断方式与桌面端既有约定一致：渲染进程用 `window.desktop?.isDesktop` 可选链判定，B/S 下 `window.desktop` 不存在，AI 入口自动隐藏或切到远程适配器（见 `desktop-client-plan.md` 第四章「运行时检测，不构建分叉」原则）。

---

## 五、文档导航

| 文档 | 内容 |
|------|------|
| **`01-ui-layout.md`** | 界面布局方案：三种形态（抽屉/分栏/全屏）、Ant Design X 组件映射、气泡与思考链设计、主题与国际化、响应式 |
| **`02-cli-connection.md`** | 连接 CLI 方案：Electron IPC 主链路、Claude Code NDJSON 接入、OpenCode SSE 接入、AgentProvider 抽象、事件契约、进程生命周期、安全模型 |
| **`03-implementation-plan.md`** | 实施计划：分阶段（P0–P6）、文件清单与改动点、依赖清单、风险与回退、验收标准 |
| **`04-comparison-with-opendesign.md`** | 与 open-design 代理本地 CLI 方式的深度对比：机制异同、逐轴对比表、两种方式优缺点、对本方案 4 条修订的溯源 |

---

## 六、关键设计决策一览

| 决策点 | 结论 | 依据 |
|--------|------|------|
| UI 库 | Ant Design X v2 | 与 antd v6 同源，组件语义完整 |
| 会话 Hook | `useXChat`（非 `useXAgent`，v2 已移除） | v2 破坏性变更 |
| 流式接入 | `XStream` 接 IPC 事件流 → `useXChat` | 官方运行时工具 |
| 本地接入主链路 | Electron 主进程 `spawn` + IPC 回推 | 复用桌面端架构，能力可达 |
| CLI 抽象 | 渲染层 `AgentAdapter` + 主进程 `AgentProvider` 双层适配 | 屏蔽来源差异与协议差异 |
| 进程模型 | `child_process.spawn`（**禁止** `shell:true`，禁止 `exec`） | 防注入 + 拿全双工流 |
| 凭证传递 | 经 env 注入子进程，**不**进命令行 | 进程列表可见性 / 审计 |
| 权限模型（v2 修订） | 交互式权限为「未验证假设」，P0 实测；兜底走沙箱+bypass 或预批准白名单 | 对比 open-design 全线 `bypassPermissions`；见 `02` §4.1、`04` |
| 传输解耦（v2 修订） | 新增「本地 BYOK SSE proxy」第三传输，解耦「传输」与「形态」 | 对比 open-design 同 daemon 既 spawn 又 BYOK；见 `03` §1、`04` |
| Windows argv（v2 修订） | 长 prompt 走 stdin / 临时文件，避开 CreateProcess 32KB 上限 | 对比 open-design 每 adapter 的 `ENAMETOOLONG` 回退；GoWind Windows 优先 |
| 事件泵解耦（v2 修订） | stdout 持续 pump + 看门狗 + SIGTERM 超时强杀，不绑进程退出 | 吸取 open-design issue #3146 |
| 工作目录 | 用户经文件对话框选择项目路径，作为 cwd 沙箱 | 复用既有 `dialog` 能力 |
| B/S 形态 | 降级为远程 Agent（P5），不在浏览器 spawn | 沙箱限制 |
| 默认 CLI | Claude Code 优先，OpenCode 次之 | 成熟度 / 流式协议清晰度 |
