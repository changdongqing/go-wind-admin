# 一、界面布局方案

> 配套：`README.md`（总览）、`02-cli-connection.md`（连接方案）、`03-implementation-plan.md`（实施计划）。
>
> 本章只设计「长什么样、怎么交互」，不涉及 CLI 连接细节。

---

## 1. 设计原则

1. **不打断主业务**：Agent 是「辅助」，默认不挤占业务页面布局。主交付形态是**右侧抽屉**，业务页面始终可见。
2. **复用现有框架**：主题（亮/暗，`ThemeProvider`）、国际化（`useI18n`）、布局（`MainLayout` 的 header widget 位）、偏好设置（`PreferencesPanel`）全部沿用，不引入新设计系统。
3. **组件语义化**：能用 Ant Design X 原子组件表达的，绝不自绘。映射关系见第 3 节。
4. **三种形态可切**：抽屉（轻）→ 分栏（中）→ 全屏（重），用户按 session 深度自选，配置持久化进 `PreferencesPanel`。

---

## 2. 三种布局形态

### 2.1 形态 A：抽屉式（默认，轻量问答）

```
┌──────────────────────────────────────────────────────────────┐
│ Header  [☰] 面包屑 …    [🔍][🌙][🔔][✨Agent][👤]   ← widget 位 │
├────────────┬─────────────────────────────────────────────────┤
│            │                                  ┌─────────────┐ │
│  Sider     │       业务页面（始终可见）         │  AI Drawer  │ │
│  Menu      │                                  │  (右抽屉)    │ │
│            │                                  │             │ │
│            │                                  │  会话切换 ▾  │ │
│            │                                  │ ─────────── │ │
│            │                                  │  消息流      │ │
│            │ │ │ │ │ (BubbleList) │ │
│            │                                  │ ─────────── │ │
│            │                                  │  [Sender]   │ │
│            │                                  └─────────────┘ │
├────────────┴─────────────────────────────────┘ ← 宽度可拖拽    │
│                       TabsBar / Footer                          │
└──────────────────────────────────────────────────────────────┘
```

- **触发**：Header 右侧 widget 区新增 `✨`（sparkles）图标，点击切换显隐。复用 `preferences.widget` 配置位（新增 `aiAssistant: true`），与现有 `themeToggle` / `notification` 等并列；快捷键复用 `shortcutKeys` 体系（新增 `globalAgent`）。
- **位置**：右侧滑出（`placement="right"`），`z-index` 高于业务内容、低于 Modal。
- **宽度**：默认 `480px`，区间 `[380, 720]`，支持拖拽左边缘改宽，宽度持久化。
- **遮罩**：**不遮罩**（`mask={false}`），保持业务页面可交互 —— 这是「不打断主业务」的关键。
- **移动端**（`isMobile`）：退化为全屏抽屉（`width: 100%` + 遮罩）。

### 2.2 形态 B：分栏式（中度，边干边问）

```
┌──────────────────────────────────────────────────────────────┐
│ Header                                                         │
├────────────┬──────────────────────────┬───────────────────────┤
│  Sider     │   业务页面（左栏）         │   AI 面板（右栏）       │
│  Menu      │   可缩放                  │   常驻、与左栏并排      │
│            │ ◀── 拖拽分隔条 ──▶        │                       │
└────────────┴──────────────────────────┴───────────────────────┘
```

- **触发**：抽屉内的「钉住 / 分栏」按钮，或 `PreferencesPanel` 设默认形态。
- **实现**：业务内容区与 AI 区用可拖拽分隔条（参考 `react-resizable-panels` 或自实现），AI 栏宽度持久化。
- **语义**：抽屉是「浮层」，分栏是「布局变更」。分栏态下 AI 区拥有完整高度、可滚动，适合长对话 + 边看业务边操作。

### 2.3 形态 C：全屏对话（重度，独立 Agent 工作台）

```
┌────────────┬─────────────────────────────────────────────────┐
│ 会话列表    │  当前会话                                        │
│ (Conversations) │                                            │
│            │  BubbleList（满高，滚动）                         │
│ + 新建      │                                                 │
│ • 会话 1    │  ThoughtChain 工具调用展开                       │
│ • 会话 2 ▣ │                                                 │
│ • 会话 3    │                                                 │
│            │  ──────────────────────────────────────────── │
│ [设置]      │  Sender（多行 + 附件 + 快捷指令 Prompts）         │
└────────────┴─────────────────────────────────────────────────┘
```

- **触发**：抽屉/分栏标题栏的「全屏打开」按钮，或独立路由 `/ai/chat`。
- **定位**：独立路由页面（走 `src/router/modules/` 自动导入 + `meta.authority` 权限码 `sys:agent:view`），不复用 `MainLayout` 的业务内容区，但**仍包在 `MainLayout` 外壳**内（保留 Sider/Header/Tabs），仅内容区替换为「会话列表 + 对话区」两栏。
- **价值**：长任务、文件 diff 审阅、多会话切换时，全屏空间最充裕；会话列表常驻左侧。

### 2.4 形态切换关系

```
       抽屉(A)  ──「钉住」──▶  分栏(B)  ──「全屏」──▶  全屏路由(C)
          ◀──「浮起」──          ◀──「收起」──
```

三种形态**共享同一套对话组件**（`<AgentConversation />`），差异仅在「容器/布局壳」，避免三份实现。

---

## 3. Ant Design X 组件映射

| UI 区域 | Ant Design X 组件 | 说明 |
|---------|-------------------|------|
| 空态首屏 | `Welcome` | 引导语 + `Prompts` 快捷指令（如「总结当前模块」「生成 CRUD」） |
| 消息流 | `BubbleList`（v2）/ `Bubble` | 用户/助手/系统三类气泡，助手气泡支持 markdown |
| 工具调用 / 思考过程 | `ThoughtChain` | 嵌入助手气泡内或独立折叠块，展示「读文件 → 改代码 → 运行」等步骤 |
| 输入框 | `Sender` | 多行、回车发送 / Shift+回车换行、`SubmitType`（发送/中断） |
| 输入引导 | `Prompts` | Sender 上方/下方快捷指令；`Suggestion` 输入联想 |
| 附件 | `Attachments` | 文件拖拽上传（传给 CLI 作为上下文） |
| 会话列表 | `Conversations` | 形态 C 左栏 / 抽屉顶部下拉 |
| 中断/重发 | Bubble `footer` actions | 复制 / 重试 / 重新生成 |

### 3.1 消息气泡设计

| 角色 | 对齐 | 渲染要点 |
|------|------|---------|
| **用户** | 右 | 纯文本 / markdown，主色边框 |
| **助手** | 左 | markdown（代码块高亮复用项目已有 `highlight.js`/`lowlight`）；流式增量渲染；footer：复制、重试 |
| **助手·工具调用** | 左（嵌套） | `ThoughtChain` 节点：工具名 + 输入摘要（可展开看完整 JSON）+ 状态（进行中✅/成功✅/失败⚠️）+ 结果摘要 |
| **系统/错误** | 居中 | 弱化样式（muted），如「CLI 未安装」「会话已中断」「权限被拒」 |

> **流式渲染**：助手文本逐 token 追加，由 `useXChat` 的 `onUpdate` 驱动 `Bubble` 内容增量更新；代码块用 `marked`（已在依赖中）+ `highlight.js` 渲染，与项目 `Editor` 组件保持一致的代码主题。

### 3.2 ThoughtChain（工具调用可视化）

```
 ThoughtChain（一次助手回复内可能多步）
 ├─ 🟢 Read  src/api/hooks/user.ts          ✅  120 行
 │     ▸ 展开看返回片段
 ├─ 🟡 Edit src/api/hooks/user.ts           ⏳ 进行中
 └─ ⚪ Bash pnpm lint                        ⏸ 等待权限（见权限 UX）
```

每个工具调用映射为一个 `ThoughtChain` 节点；点击节点展开「输入参数 / 输出结果」两段。结果超长则折叠 + 「查看全部」抽屉。

---

## 4. 权限确认 UX（交互式工具调用）

> 🔄 **v2 评审修订（对比 open-design）—— 本节 UX 依赖 P0 验证结果** ⚠️
> 下述「权限请求卡片 + stdin 回写」交互流，前提是 Claude Code headless `stream-json` 模式支持「emit 权限请求并阻塞等 stdin 应答」。该语义**未经验证**（`--permission-mode` 无此模式，open-design 全线 `bypassPermissions` 回避了它，详见 `02-cli-connection.md` §4.1、`04-comparison-with-opendesign.md`）。P0 实测后：
> - 若**成立** → 按本节设计落地交互卡；
> - 若**不成立** → 改走「预批准只读白名单 + 高危会话级开闸」或「沙箱副本 + bypass + patch 审查」闭环，届时本节卡片降级为「白名单/沙箱配置面板」。

部分 CLI（Claude Code 交互流）在执行敏感工具（写文件、跑命令）前会**请求权限**。UI 设计：

1. 消息流中出现一条「权限请求卡片」（特殊 Bubble）：说明工具 + 目标 + 影响范围。
2. 卡片内按钮：`允许本次` / `允许本会话该工具` / `拒绝`。
3. 用户选择 → 经 Adapter 回写 stdin（`--input-format stream-json`，详见 `02-cli-connection.md`）。
4. 非交互模式（`--allowedTools` 预批准 / `--permission-mode` 自动）下不出现卡片，`ThoughtChain` 直接走「进行中→完成」。

权限策略可在 `PreferencesPanel` 新增「AI Agent」分区配置默认权限模式（本次/会话/始终询问），与桌面端既有偏好面板风格一致。

---

## 5. 主题、国际化与无障碍

### 5.1 主题适配

- **亮/暗模式**：复用 `ThemeProvider`，Ant Design X 组件继承 `ConfigProvider` 的 `algorithm`（`defaultAlgorithm` / `darkAlgorithm`）。气泡背景、代码块主题随 `effectiveMode` 切换。
- **主题色**：助手气泡 / 链接 / 代码高亮跟随 `theme.colorPrimary`，与全站一致。
- **圆角**：跟随 `theme.radius`。
- 注意：Agent 面板默认 `背景 = 容器背景`，在亮色下用 `#ffffff`/`#fafafa`，暗色下用 `#141414`，避免大面积纯黑对比过强（呼应 `MainLayout` 现有 `isDark ? '#141414' : '#ffffff'` 的取值）。

### 5.2 国际化

- 新增命名空间 `_modules/agent.json`（zh-CN / en-US 各一份），覆盖：空态文案、权限按钮、错误提示、会话操作（新建/重命名/删除）、形态切换、设置项。
- 硬编码文本一律走 `useI18n('agent')`，遵循 `frontend/admin/react/CLAUDE.md` 国际化规范。

### 5.3 无障碍 & 键盘

- `Sender` 回车发送、Shift+回车换行、`Esc` 中断生成、`Cmd/Ctrl+K` 唤起 Agent（复用 `shortcutKeys`）。
- 工具调用节点、权限卡片均为键盘可达（Tab + Enter）。

---

## 6. 响应式与降级

| 场景 | 表现 |
|------|------|
| 桌面（≥1280） | 抽屉默认 480px；可切分栏 / 全屏 |
| 平板（768–1279） | 抽屉 380px 起；分栏禁用，仅抽屉/全屏 |
| 移动（<768，`isMobile`） | 抽屉全屏化 + 遮罩；会话列表收进顶部下拉 |
| B/S 浏览器（非桌面） | Agent 入口默认隐藏；若后端提供远程 Agent，则走 `RemoteAgentAdapter`，UI 完全相同（适配层屏蔽来源） |

---

## 7. 与现有布局的集成点（改动面预览）

> 详细改动清单见 `03-implementation-plan.md`，这里只标注 UI 集成触点。

| 集成点 | 文件 | 改动 |
|--------|------|------|
| Header widget 区 | `layouts/MainLayout/components/HeaderContent.tsx` + `widgetConfig` | 新增 `aiAssistant` widget + ✨ 图标按钮 |
| 偏好设置 | `core/preferences/`（types/default/panel） | `widget.aiAssistant`、`agent.*`（默认形态/默认 CLI/权限模式） |
| 抽屉/分栏容器 | `layouts/MainLayout/index.tsx` | 在内容区右侧挂 `<AgentDock />`（抽屉/分栏可切） |
| 全屏路由 | `router/modules/agent.tsx`（新增，自动导入） | `/ai/chat` → `<AgentWorkbench />` |
| 共享对话组件 | `pages/.../agent/AgentConversation.tsx`（新增） | 三形态复用 |
| 国际化 | `locales/{zh-CN,en-US}/_modules/agent.json` | 新增 |

设计上保证：**B/S 与桌面同源、不构建分叉**，桌面专有能力（spawn CLI）通过 `window.desktop?.agent` 可选链兜底，B/S 下优雅降级 —— 与 `desktop-client-plan.md` 第四章原则一致。
