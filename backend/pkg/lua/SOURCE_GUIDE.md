# 脚本引擎与多来源加载指南

本文档说明 `pkg/lua` 升级后的**多语言脚本引擎架构**与**多来源加载 / 热更新**能力。

## 架构概览

升级后的 `pkg/lua` 采用 **Hook 编排器 + 脚本引擎** 双层架构：

```
┌───────────────────────────────────────────────┐
│  Engine (Hook 编排器, 语言无关)                  │
│  ├─ Hook 注册 / 优先级 / 链式执行 / 回调管理       │
│  ├─ 执行上下文 (Context)                         │
│  └─ 持有 gsEngine.Engine (脚本引擎接口)           │
│       │                                         │
│       ├── LuaEngine   ← 当前实现 (安全沙箱)       │
│       ├── JSEngine    ← 未来 (goja)             │
│       └── PythonEngine ← 未来 (gpython)         │
└───────────────────────────────────────────────┘
```

**关键点**：直接采用 [go-scripts/lua](https://github.com/tx7do/go-scripts/tree/main/lua) 子模块作为脚本引擎实现。
通过 `gsEngine.NewScriptEngine(LuaType)` 按类型创建，业务 API 与 hook.register 经 **RuntimeHook** 注入。
未来添加 JS/Python 只需 `NewScriptEngine(JavaScriptType)` 等，编排器零改动。

### 为何现在能直接用 go-scripts/lua？

go-scripts v0.0.7 新增了 **RuntimeHook** 机制，解决了之前阻碍直接采用的两个核心问题：

| 能力 | 机制 | 说明 |
|------|------|------|
| ✅ **业务 API 注入** | `AddRuntimeHook` | VM 创建后、Load/Execute 前运行，可注入 Redis/EventBus/OSS/Crypto 等 8 个业务模块 |
| ✅ **池隔离** | `recordBusinessGlobal` + `ClearGlobals` | 业务全局变量被追踪，归还池前清除，引擎实例间不泄漏 |
| ✅ **Hook 自注册** | RuntimeHook 注册 `hook.register` Go 函数 | 脚本可调用 `hook.register(name, fn)` 将回调交还 Go 侧 |

**沙箱已启用**（go-scripts/lua v0.0.8+）：通过 `SetOpenLibs` 仅开启安全标准库，
禁用 `os`/`io`/`debug`/`load` 等危险库（命令注入 / 文件系统 / 反射绕过防护）。
安全库白名单：base / load / table / string / math / coroutine。详见 `applySandbox`。

| 库 | 状态 | 原因 |
|----|------|------|
| base | ✅ 开启 | 基础函数（print/pairs/error/require） |
| load (package) | ✅ 开启 | require / module loaders，模块系统必需 |
| table | ✅ 开启 | 表操作 |
| string | ✅ 开启 | 字符串处理 |
| math | ✅ 开启 | 数学运算 |
| coroutine | ✅ 开启 | 协程 |
| **os** | ❌ 禁用 | `os.execute` 命令注入 / `os.remove` 文件删除 / `os.getenv` 环境泄漏 |
| **io** | ❌ 禁用 | 读写任意文件 |
| **debug** | ❌ 禁用 | 反射调试，可绕过元表保护 |

### 架构关键点

- **引擎单 VM**：go-scripts/lua 引擎生命周期内持有单个 VM（statePool.Borrow），回调引用在 Close 前持续有效
- **LoadString vs ExecuteString**：go-scripts 的 `LoadString` 仅编译不执行；本项目 `LoadScriptString` 内部用 `ExecuteString` 以确保 `hook.register` 立即触发
- **模块注册语义**：`engine.RegisterModule(name, loader)` 调用 loader 时传入 name 参数，loader 须执行 `PreloadModule`（见 `preloadAdapter`）

## 多语言切换

通过 `ScriptEngineFactory` 注入引擎实现：

```go
import gsEngine "github.com/tx7do/go-scripts"

// 默认使用 go-scripts/lua（经 init() 自动注册工厂）
engine := lua.NewEngine(lua.DefaultConfig(), logger)

// 自定义引擎工厂（切换语言）
lua.SetEngineFactory(func(config *lua.Config, logger log.Logger) (gsEngine.Engine, error) {
    return gsEngine.NewScriptEngine(config.EngineType) // JavaScriptType 等
})

// 或单次指定
config := lua.DefaultConfig()
config.EngineType = gsEngine.LuaType // 当前支持 lua，未来扩展 javascript 等
```

### 脚本编写约定（适配 go-scripts）

执行上下文通过 RuntimeHook 注入的全局函数访问（非参数）：

```lua
local log = require "kratos_logger"
local hook = require "kratos_hook"

-- 入口函数：execute() 无参数，上下文经 __get_ctx() 获取
function execute()
    local ctx = __get_ctx()
    local input = ctx.get("input")
    ctx.set("result", "processed: " .. input)
    -- ctx.stop("reason")  -- 中止 hook 链
    return true  -- 返回 false 也表示中止
end
```

> 注：hook.register 注册的回调函数仍以 `function(ctx)` 形式接收上下文参数（回调路径直接传递 ctx 表）。

## 脚本来源 (Source)

脚本不再只能从文件系统加载，支持多种来源，统一实现 `source.Reader` 接口：

| 来源 | 类型 | 热更新 | 适用场景 |
|------|------|--------|---------|
| **FileSource** | 本地文件 | ✅ mtime 轮询 | 开发调试 |
| **DBSource** | 数据库 | ✅ 轮询/hash | 生产（管理后台 UI 维护脚本） |
| **MemSource** | 内存 | ✅ channel | 单测、动态推送 |
| go-scripts 其他 | S3/etcd/Redis/HTTP | ✅ 各自机制 | 分布式部署 |

### 文件来源

```go
engine := lua.NewEngine(config, logger)

// 从目录自动加载（向后兼容）
engine.LoadScriptsFromDir(ctx, "./scripts")

// 或绑定 FileSource（支持搜索路径 + 热更新）
src := lua.NewFileSource("./scripts", "./scripts/hooks")
engine.SetSource(src)

// 按 key 加载
engine.LoadScript(ctx, "on_login.lua")

// 启用热更新（文件变更自动重新加载）
engine.WatchScript(ctx, "on_login.lua")
```

### 数据库来源（核心收益）

脚本存储在数据库表中，通过管理后台 UI 增删改，无需重启服务：

```go
// 在 app 层 Wire 装配时注入真正的 DB 查询函数（避免 pkg → data 循环依赖）
dbSource := lua.NewDBSource(
    func(ctx context.Context, key string) (string, error) {
        // 从数据库按 name/id 查询脚本源码
        return scriptRepo.GetSourceByName(ctx, key)
    },
    lua.WithScriptHasher(func(ctx context.Context, key string) (string, error) {
        // 高效变更检测（返回 updated_at 或版本号）
        return scriptRepo.GetUpdatedAt(ctx, key)
    }),
    lua.WithPollInterval(5*time.Second), // 热更新轮询间隔
)
defer dbSource.Close()

engine.SetSource(dbSource)

// 加载并监听
engine.LoadScript(ctx, "user_registered")
engine.WatchScript(ctx, "user_registered") // DB 变更自动重载
```

**DBSource 特性**：
- 内存缓存：避免每次 Load 都查库
- `Invalidate(key)` / `InvalidateAll()`：主动失效缓存
- 热更新：`hasher` 优先（高效），否则比对源码字符串

### 缓存层（远程源推荐）

对远程源（DB/S3/etcd）可用 go-scripts 的 `CachedSource` 包裹，减少 IO：

```go
import gsSource "github.com/tx7do/go-scripts/source"

cached, _ := gsSource.NewCachedSource(dbSource, gsSource.WithTTL(5*time.Minute))
engine.SetSource(cached)
```

## 引擎接口能力

go-scripts/lua 引擎实现完整的 `Engine` 接口（7 个能力子接口）：

| 能力 | 方法 | 说明 |
|------|------|------|
| 生命周期 | `Init` / `Close` / `IsInitialized` | 引擎初始化与释放 |
| 脚本加载 | `Load` / `LoadMulti` / `LoadString` | 从 Source 或内联加载（仅编译） |
| 脚本执行 | `Execute` / `ExecuteFromKey` / `ExecuteString` | 执行并返回结果 |
| 全局访问 | `RegisterGlobal` / `GetGlobal` | 全局变量读写 |
| 函数注册 | `RegisterFunction` / `CallFunction` | 宿主函数注册与调用 |
| 模块注册 | `RegisterModule` | Lua 模块（preload 风格 loader） |
| 热更新 | `StartWatch` / `StopWatch` | 脚本变更自动重载 |
| **RuntimeHook** | `AddRuntimeHook` | **业务 API / hook.register / 执行上下文注入** |

## 向后兼容

编排器 API 保持兼容：
- `NewEngine(config, logger)` — 创建编排器（默认 go-scripts/lua）
- `ExecuteHook(ctx, hookName, execCtx)` — Hook 执行
- `AddScript` / `RegisterHook` / `RegisterCallback` — Hook 管理
- `LoadScriptsFromDir` / `LoadScriptFile` / `LoadScriptString` — 脚本加载
- `SetRedis` / `SetEventBus` / `SetOSS` — 业务依赖注入

**脚本 API 变更**：`execute()` 不再接收 ctx 参数，改用 `__get_ctx()` 获取上下文（适配 go-scripts 单 VM 架构）。

## 文件结构

```
pkg/lua/
├── engine.go              # Hook 编排器（语言无关，持有 gsEngine.Engine）
├── engine_runtime.go      # RuntimeHook：注入 8 个业务 API + hook.register + 执行上下文
├── source_file.go         # FileSource（包装 go-scripts FileSource）
├── source_db.go           # DBSource（数据库 + 热更新）
├── errors.go              # 占位（引擎错误由 go-scripts 提供）
├── script.go / context.go # 数据结构
├── hook/registry.go       # Hook 系统
├── api/                   # 8 个业务 API 模块（各含 Loader* 函数供 RuntimeHook 注入）
└── internal/convert/      # Lua ↔ Go 转换
```

