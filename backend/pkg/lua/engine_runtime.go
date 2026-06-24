package lua

import (
	"context"

	lua "github.com/yuin/gopher-lua"

	gsEngine "github.com/tx7do/go-scripts"
	gsLua "github.com/tx7do/go-scripts/lua"

	"go-wind-admin/pkg/lua/api"
)

// applySandbox 配置 Lua 引擎的标准库白名单，仅开启安全库，禁用危险库。
//
// 安全库（白名单）：
//   - Base:      基础函数（print/pairs/error/require 等）
//   - Load:      package 库（require / module loaders，保持模块系统可用）
//   - Table:     table 库
//   - String:    string 库
//   - Math:      math 库
//   - Coroutine: coroutine 库
//
// 禁用的危险库：
//   - os:    os.execute / os.remove / os.getenv（命令注入 / 文件系统 / 环境泄漏）
//   - io:    读写任意文件
//   - debug: 反射调试（可绕过元表保护）
//
// 仅对 go-scripts/lua 引擎生效（实现了 SetOpenLibs）；其他引擎为 no-op。
// 必须在 Init 前调用。
func (e *Engine) applySandbox(eng gsEngine.Engine) {
	type openLibsSetter interface {
		SetOpenLibs(libs ...string)
	}
	if setter, ok := eng.(openLibsSetter); ok {
		setter.SetOpenLibs(
			gsLua.AllowedLibBase,
			gsLua.AllowedLibLoad,
			gsLua.AllowedLibTab,
			gsLua.AllowedLibStr,
			gsLua.AllowedLibMath,
			gsLua.AllowedLibCoroutine,
		)
		e.logger.Info("🔒 Lua sandbox enabled (allowed libs: base/load/table/string/math/coroutine)")
	}
}

// hookEngineAdapter 将编排器适配为 api.HookEngine 接口，
// 供 RuntimeHook 注入的 hook 模块使用（脚本调用 hook.register 时回调编排器）。
type hookEngineAdapter struct {
	orchestrator *Engine
}

// 确保 hookEngineAdapter 实现 api.HookEngine。
var _ api.HookEngine = (*hookEngineAdapter)(nil)

func (a *hookEngineAdapter) RegisterHook(name, description string) error {
	return a.orchestrator.RegisterHook(name, description)
}

func (a *hookEngineAdapter) AddScript(hookName string, script interface{}) error {
	return a.orchestrator.AddScript(hookName, script)
}

func (a *hookEngineAdapter) ListHooks() []string {
	return a.orchestrator.ListHooks()
}

// RegisterCallback 注册脚本侧的 Lua 回调到编排器。
// go-scripts/lua 引擎的池会自动隔离 VM，回调函数在调用 VM 上注册即可。
func (a *hookEngineAdapter) RegisterCallback(hookName string, L *lua.LState, fn *lua.LFunction) {
	a.orchestrator.registerLuaCallback(hookName, L, fn)
}

// vmManagerAdapter 将编排器适配为 api.VMManager。
// go-scripts/lua 引擎内部管理池，MarkVMDedicated 在新架构下为 no-op（池自动处理生命周期）。
type vmManagerAdapter struct{}

func (vmManagerAdapter) MarkVMDedicated(L *lua.LState) {
	// go-scripts/lua 引擎的 statePool 管理生命周期，此处无需处理。
}

// preloadAdapter 将 builder 风格的 loader（构建模块并 push）适配为
// go-scripts/lua 的 RegisterModule 所期望的 preload 风格 loader。
//
// go-scripts 的 virtualMachine.RegisterModule(name, mod) 会执行：
//
//	L.Push(NewFunction(mod)); L.Push(name); L.Call(1, 0)
//
// 即调用 mod(name)。故 mod 必须是一个「读 name 参数并注册 preload」的函数，
// 而非直接构建模块的 builder。
//
// 本适配器读取栈顶 name，调用 L.PreloadModule(name, builder)，
// builder 是原始的 api.Loader*（构建模块并 push，返回 1）。
func preloadAdapter(builder lua.LGFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		name := L.CheckString(1) // RegisterModule 传入的模块名
		L.PreloadModule(name, builder)
		return 0
	}
}

// buildRuntimeHook 构造一个 RuntimeHook，在 VM 创建后注入全部业务 API 模块与执行上下文。
//
// 该 hook 通过 go-scripts 的 AddRuntimeHook 注册，引擎保证：
//   - Init 后、首次 Load*/Execute* 前执行
//   - 池中 VM 复用时重放，并清除上一个引擎实例注入的业务全局变量（池隔离）
func (e *Engine) buildRuntimeHook() gsEngine.RuntimeHook {
	return func(_ context.Context) error {
		eng := e.scriptEngine

		// 注入 8 个业务模块（经 preloadAdapter 适配为 go-scripts 期望的 preload 风格）
		if err := eng.RegisterModule("kratos_logger", preloadAdapter(api.LoaderLogger(e.logger))); err != nil {
			return err
		}
		if err := eng.RegisterModule("kratos_crypto", preloadAdapter(api.LoaderCrypto(e.logger))); err != nil {
			return err
		}
		if err := eng.RegisterModule("kratos_util", preloadAdapter(api.LoaderUtil(e.logger))); err != nil {
			return err
		}

		if e.rdb != nil {
			if err := eng.RegisterModule("kratos_cache", preloadAdapter(api.LoaderCache(e.rdb, e.logger))); err != nil {
				return err
			}
		}
		if e.eventbusManager != nil {
			if err := eng.RegisterModule("kratos_eventbus", preloadAdapter(api.LoaderEventBus(e.eventbusManager, e.logger))); err != nil {
				return err
			}
		}
		if e.ossClient != nil {
			if err := eng.RegisterModule("kratos_oss", preloadAdapter(api.LoaderOSS(e.ossClient, e.logger))); err != nil {
				return err
			}
		}

		// hook 模块（脚本自注册 hook 回调）
		hookAdapter := &hookEngineAdapter{orchestrator: e}
		if err := eng.RegisterModule("kratos_hook", preloadAdapter(api.LoaderHook(hookAdapter, e.logger))); err != nil {
			return err
		}

		// task 模块
		if err := eng.RegisterModule("task", preloadAdapter(api.LoaderTask(vmManagerAdapter{}, e.logger))); err != nil {
			return err
		}

		// 注册执行上下文访问函数（脚本通过 __get_ctx / __set_ctx / __stop 访问当前 Context）
		e.registerContextFunctions(eng)

		return nil
	}
}

// registerContextFunctions 注册执行上下文的全局便捷函数。
// 脚本可调用 __get_ctx() / __set_ctx(k,v) / __stop(reason) 操作当前执行上下文。
func (e *Engine) registerContextFunctions(eng gsEngine.Engine) {
	var (
		getCtx lua.LGFunction = func(L *lua.LState) int {
			ctx := e.execCtx.current
			if ctx == nil {
				L.Push(L.NewTable())
				return 1
			}
			L.Push(e.contextToLuaTable(L, ctx))
			return 1
		}
		setCtx lua.LGFunction = func(L *lua.LState) int {
			ctx := e.execCtx.current
			if ctx != nil {
				key := L.CheckString(1)
				val := L.Get(2)
				ctx.Data[key] = apiToGoValue(val)
			}
			return 0
		}
		stopCtx lua.LGFunction = func(L *lua.LState) int {
			ctx := e.execCtx.current
			if ctx != nil {
				ctx.Stopped = true
				ctx.StopReason = L.OptString(1, "stopped by script")
			}
			return 0
		}
	)

	_ = eng.RegisterFunction("__get_ctx", getCtx)
	_ = eng.RegisterFunction("__set_ctx", setCtx)
	_ = eng.RegisterFunction("__stop", stopCtx)
}
