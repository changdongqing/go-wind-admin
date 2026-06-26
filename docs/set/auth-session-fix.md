# 登录状态维持问题修复方案

> 适用对象：**开发人员**
>
> 本文档针对「登录持续时间短」和「操作中也出现登录状态失效」两个问题，给出完整的根因分析和修复方案。
>
> ⚠️ 本文档仅描述方案，不涉及实际代码修改。

> 📌 **实施状态（2026-06-26 更新）**
> - ✅ **已实施**：Go 后端（B1/B2/B3）+ React 前端（B3 配套 / F1 / F2），已通过 `go vet`、`tsc --noEmit`、`vite build` 验证。
> - 📌 **未实施**：Vue 前端同步改动，待后续处理。
> - 🔧 **评审修正**（相比原始方案）：
>   1. **C1/C2（阻断）**：B3 前端原方案从 `userInfo.userId` 取值，但 `UserInfo` 类型无 `userId` 字段，且 `userInfo` 不持久化（F5 后为 null）会导致刷新失败。改为统一从 AT payload 解码 `uid`/`jti`（`decodeAccessTokenIdentity`）。
>   2. **C3**：P1 中死代码路径修正为 `backend/pkg/jwt/user_token_payload.go`。
>   3. **C4**：F2 由「硬编码 `/refresh-token`」改为 `shouldSkipAuth` 谓词注入，符合现有「业务层 → 基础设施层」架构。

---

## 一、问题现象

| 现象 | 描述 |
|------|------|
| 🔴 登录持续时间短 | 登录后约 15 分钟即被踢出，需要重新登录 |
| 🔴 操作中也会失效 | 用户正在操作页面（没有闲置），也会突然跳转到登录页 |

---

## 二、根因分析

### 认证架构概览

```
┌──────────┐    Login     ┌──────────┐   JWT+Redis    ┌───────┐
│  前端    │ ──────────→  │  后端    │ ──────────────→ │ Redis │
│ (React)  │ ←────────── │  (Go)    │                 │       │
└──────────┘  AT + RT     └──────────┘                 └───────┘

AT = Access Token  (JWT, 15min)     ← 🚨 太短！
RT = Refresh Token (随机串, 7天)
```

### 问题清单

| # | 问题 | 严重度 | 说明 |
|---|------|--------|------|
| P1 | Access Token 只有 15 分钟 | 🔴 高 | 任何刷新失败都会导致登出，刷新压力极大 |
| P2 | RefreshToken API 在 auth 中间件保护下 | 🔴 高 | AT 过期后刷新请求本身无法通过认证，逻辑死锁 |
| P3 | 页面刷新后定时器丢失 | 🔴 高 | F5 后只能靠被动 401 刷新，若此时 AT 已过期则直接登出 |
| P4 | 并发 401 处理风险 | 🟡 中 | 多请求同时 401 时，刷新请求自身若 401 则直接 forceLogout |
| P5 | 无活动续期机制 | 🟡 中 | 用户持续操作也不会延长 AT 寿命 |

### P1：Access Token 只有 15 分钟

**位置**：`backend/app/admin/service/internal/data/authenticator.go:24`

```go
const (
    DefaultAccessTokenExpires  = time.Minute * 15   // 🚨 实际生效值：15 分钟
    DefaultRefreshTokenExpires = time.Hour * 24 * 7
)
```

而 `backend/pkg/jwt/user_token_payload.go:33` 中定义的 `DefaultTokenExpiration = 2 * time.Hour` **从未被引用**，属于死代码。

**影响**：每 15 分钟必须刷新一次 AT，刷新频率高，任何网络抖动、服务短暂不可用都会导致刷新失败→登出。

### P2：RefreshToken API 在 auth 中间件保护下（核心矛盾）

**位置**：`backend/app/admin/service/internal/server/rest_server.go:54-62`

白名单中**不包含** `OperationAuthenticationServiceRefreshToken`：

```go
rpc.AddWhiteList(
    adminV1.OperationAuthenticationServiceLogin,
    adminV1.OperationAuthenticationServiceGenerateCaptcha,
    adminV1.OperationAuthenticationServiceVerifyCaptcha,
    // ❌ 缺少 OperationAuthenticationServiceRefreshToken
)
```

**位置**：`backend/app/admin/service/internal/service/authentication_service.go:464-479`

`RefreshToken` 方法从 `auth.FromContext(ctx)` 获取操作人信息——而这些信息是由 auth 中间件从 AT 中解析后注入的：

```go
func (s *AuthenticationService) RefreshToken(...) {
    operator, err := auth.FromContext(ctx)  // 🚨 依赖 AT 已通过认证
    ...
    req.UserId = trans.Ptr(operator.GetUserId())
    req.Jti = operator.Jti
    return s.doGrantTypeRefreshToken(ctx, req)
}
```

**矛盾链条**：

```
AT 过期 → 前端发 RefreshToken 请求 → 请求携带过期的 AT
→ auth 中间件校验 AT → 校验失败返回 401
→ 前端 401 拦截器发现是 /refresh-token 请求的 401 → 直接 forceLogout
→ 用户被踢出 💥
```

**这是「操作中也会失效」的根本原因**：AT 15 分钟一到，刷新请求自身就无法通过认证，用户必然被登出。

### P3：页面刷新后定时器丢失

**位置**：`frontend/admin/react/src/stores/auth.ts:141`

`startRefreshTimer()` 只在登录成功时调用一次：

```typescript
// 登录成功后
startRefreshTimer();
```

页面 F5 刷新后，Zustand persist 从 localStorage 恢复了 token 数据，但 `startRefreshTimer()` 不会重新执行。此后只能依赖 401 被动拦截来刷新。

**位置**：`frontend/admin/react/src/hooks/useTokenRefresh.ts`

定时器模块级变量 `refreshTimer` 在页面刷新后重置为 `null`，没有任何恢复机制。

### P4：并发 401 处理风险

**位置**：`frontend/admin/react/src/core/transport/rest/preset-interceptors.ts:38-45`

当 `/refresh-token` 请求自身返回 401 时，直接 forceLogout：

```typescript
if (config.url?.includes('/refresh-token')) {
    await doReAuthenticate();  // → forceLogout()
    throw ...;
}
```

这个逻辑本身是合理的（防死锁），但在 P2 存在的情况下，它会成为「必然触发」的路径。

### P5：无活动续期机制

当前 AT 的过期时间在创建时就固定了（`time.Now().Add(expires)`），不会因为用户持续操作而延长。即使定时刷新成功，也只是「换了一张新 AT」，不会续期旧 AT。这意味着如果定时刷新因任何原因延迟（如浏览器标签页休眠），旧 AT 在延迟期间过期，用户就会遇到 401。

---

## 三、修复方案

### 修改总览

| 优先级 | 编号 | 修改项 | 端 | 风险 |
|--------|------|--------|-----|------|
| ⭐⭐⭐ | B2 | RefreshToken API 加入白名单 | 后端 | 中 |
| ⭐⭐⭐ | B3 | RefreshToken 方法改为从请求体获取用户信息 | 后端 | 中 |
| ⭐⭐ | B1 | Access Token 过期时间调整为 2 小时 | 后端 | 低 |
| ⭐⭐ | F1 | 页面刷新后恢复定时刷新定时器 | 前端 | 低 |
| ⭐ | F2 | RefreshToken 请求不携带过期 AT | 前端 | 低 |

> 🔔 **B2 + B3 必须一起修改**，否则 RefreshToken API 加入白名单后，服务端无法获取用户信息。

---

### B1：Access Token 过期时间调整为 2 小时

**问题**：15 分钟太短，刷新压力大，任何刷新失败都会导致登出。

**修改文件**：`backend/app/admin/service/internal/data/authenticator.go`

**修改前**（第 23-24 行）：

```go
const (
    // DefaultAccessTokenExpires  默认访问令牌过期时间
    DefaultAccessTokenExpires = time.Minute * 15

    // DefaultRefreshTokenExpires 默认刷新令牌过期时间
    DefaultRefreshTokenExpires = time.Hour * 24 * 7
)
```

**修改后**：

```go
const (
    // DefaultAccessTokenExpires  默认访问令牌过期时间：2 小时
    DefaultAccessTokenExpires = time.Hour * 2

    // DefaultRefreshTokenExpires 默认刷新令牌过期时间：7 天
    DefaultRefreshTokenExpires = time.Hour * 24 * 7
)
```

**影响范围**：
- 所有新创建的 Access Token 有效期变为 2 小时
- 已发放的旧 AT 不受影响（TTL 已写入 Redis）
- 前端 `DEFAULT_ACCESS_EXPIRES_IN` 常量（`auth.ts:60`）当前值为 `7200`（2小时），与后端修改后一致，无需改动

**风险**：⬇️ 低。AT 有效期变长，安全性略有降低，但有 RT 机制兜底 + Redis 黑名单机制可主动吊销。

---

### B2：RefreshToken API 加入白名单

**问题**：RefreshToken API 在 auth 中间件保护下，AT 过期后刷新请求无法通过认证，形成死锁。

**修改文件**：`backend/app/admin/service/internal/server/rest_server.go`

**修改前**（第 54-62 行）：

```go
// add white list for authentication.
rpc.AddWhiteList(
    adminV1.OperationAuthenticationServiceLogin,
    adminV1.OperationAuthenticationServiceGenerateCaptcha,
    adminV1.OperationAuthenticationServiceVerifyCaptcha,
)
```

**修改后**：

```go
// add white list for authentication.
rpc.AddWhiteList(
    adminV1.OperationAuthenticationServiceLogin,
    adminV1.OperationAuthenticationServiceGenerateCaptcha,
    adminV1.OperationAuthenticationServiceVerifyCaptcha,
    adminV1.OperationAuthenticationServiceRefreshToken, // 刷新令牌接口免认证，否则 AT 过期后无法刷新
)
```

**影响范围**：
- `POST /admin/v1/refresh-token` 不再需要有效的 AT 即可访问
- 必须同步修改 B3，否则 `RefreshToken` 方法中 `auth.FromContext(ctx)` 会取不到值

**风险**：🟡 中。需确保 B3 的修改后，RefreshToken 方法仅依赖请求体中的信息（`user_id`、`refresh_token`）来验证身份，而非依赖 auth 中间件注入的上下文。同时需确保 RT 本身的验证足够安全（当前已有 Redis 原子验证 + 吊销机制，安全性足够）。

---

### B3：RefreshToken 方法改为从请求体获取用户信息

**问题**：B2 加入白名单后，`auth.FromContext(ctx)` 将返回错误，需要改为从请求参数获取用户信息。

**修改文件**：`backend/app/admin/service/internal/service/authentication_service.go`

**修改前**（第 464-479 行）：

```go
// RefreshToken 刷新令牌
func (s *AuthenticationService) RefreshToken(ctx context.Context, req *authenticationV1.LoginRequest) (*authenticationV1.LoginResponse, error) {
    // 校验授权类型
    if req.GetGrantType() != authenticationV1.GrantType_refresh_token {
        return nil, authenticationV1.ErrorInvalidGrantType("invalid grant type")
    }

    operator, err := auth.FromContext(ctx)  // 🚨 依赖 auth 中间件注入的上下文
    if err != nil {
        return nil, err
    }

    req.ClientType = trans.Ptr(authenticationV1.ClientType_admin)
    req.UserId = trans.Ptr(operator.GetUserId())
    req.Jti = operator.Jti

    return s.doGrantTypeRefreshToken(ctx, req)
}
```

**修改后**：

```go
// RefreshToken 刷新令牌
func (s *AuthenticationService) RefreshToken(ctx context.Context, req *authenticationV1.LoginRequest) (*authenticationV1.LoginResponse, error) {
    // 校验授权类型
    if req.GetGrantType() != authenticationV1.GrantType_refresh_token {
        return nil, authenticationV1.ErrorInvalidGrantType("invalid grant type")
    }

    // 校验必要参数：refresh_token 和 user_id 必须由客户端提供
    if req.GetRefreshToken() == "" {
        return nil, authenticationV1.ErrorBadRequest("refresh_token is required")
    }
    if req.GetUserId() == 0 {
        return nil, authenticationV1.ErrorBadRequest("user_id is required")
    }

    // 设置客户端类型
    if req.GetClientType() == authenticationV1.ClientType(0) {
        req.ClientType = trans.Ptr(authenticationV1.ClientType_admin)
    }

    // 使用请求体中的 user_id 和 jti，不再依赖 auth 中间件注入的上下文
    return s.doGrantTypeRefreshToken(ctx, req)
}
```

**同时需要修改 `doGrantTypeRefreshToken` 方法**（第 390-441 行），去除对 `auth.FromContext` 的依赖：

**修改前**：

```go
func (s *AuthenticationService) doGrantTypeRefreshToken(ctx context.Context, req *authenticationV1.LoginRequest) (*authenticationV1.LoginResponse, error) {
    // 获取操作人信息
    operator, err := auth.FromContext(ctx)  // 🚨 依赖 auth 中间件
    if err != nil {
        return nil, err
    }

    // 获取用户信息
    user, err := s.userRepo.Get(ctx, &identityV1.GetUserRequest{
        QueryBy: &identityV1.GetUserRequest_Id{
            Id: operator.UserId,  // 🚨 来自 auth 中间件
        },
    })
    ...
    // 验证刷新令牌
    if err = s.authenticator.VerifyRefreshToken(ctx, req.GetClientType(), req.GetUserId(), operator.GetJti(), req.GetRefreshToken()); err != nil {
        ...
    }
    ...
}
```

**修改后**：

```go
func (s *AuthenticationService) doGrantTypeRefreshToken(ctx context.Context, req *authenticationV1.LoginRequest) (*authenticationV1.LoginResponse, error) {
    // 重置上下文（绕过隐私保护中间件）
    ctx = s.resetContextForLogin(ctx)

    // 获取用户信息（使用请求体中的 user_id）
    user, err := s.userRepo.Get(ctx, &identityV1.GetUserRequest{
        QueryBy: &identityV1.GetUserRequest_Id{
            Id: req.GetUserId(),  // ✅ 来自请求体
        },
    })
    if err != nil {
        s.log.Errorf("get user by id [%d] failed [%s]", req.GetUserId(), err.Error())
        return nil, err
    }

    tokenPayload := &authenticationV1.UserTokenPayload{
        UserId:   user.GetId(),
        TenantId: user.TenantId,
        Username: user.Username,
        ClientId: req.ClientId,
        DeviceId: req.DeviceId,
    }

    // 解析用户权限信息
    err = s.resolveUserAuthority(ctx, user, tokenPayload)
    if err != nil {
        s.log.Errorf("resolve user [%d] authority failed [%s]", user.GetId(), err.Error())
        return nil, err
    }

    // 验证刷新令牌（使用请求体中的 user_id 和 jti）
    if err = s.authenticator.VerifyRefreshToken(ctx, req.GetClientType(), req.GetUserId(), req.GetJti(), req.GetRefreshToken()); err != nil {
        s.log.Errorf("verify refresh token failed for user [%d]: [%s]", req.GetUserId(), err)
        return nil, authenticationV1.ErrorIncorrectRefreshToken("invalid refresh token")
    }

    // 生成令牌
    accessToken, refreshToken, err := s.authenticator.CreateUserToken(ctx, req.GetClientType(), tokenPayload)
    if err != nil {
        return nil, err
    }

    return &authenticationV1.LoginResponse{
        TokenType:        authenticationV1.TokenType_bearer,
        AccessToken:      accessToken,
        RefreshToken:     trans.Ptr(refreshToken),
        ExpiresIn:        int64(s.authenticator.GetAccessTokenExpires(req.GetClientType()).Seconds()),
        RefreshExpiresIn: trans.Ptr(int64(s.authenticator.GetRefreshTokenExpires(req.GetClientType()).Seconds())),
    }, nil
}
```

**前端需同步修改**：RefreshToken 请求必须携带 `user_id` 和 `jti` 参数。

**修改文件**：`frontend/admin/react/src/api/hooks/auth.ts`（第 109-117 行）

**修改前**：

```typescript
export const refreshTokenMutation = queryClient.getMutationCache().build(queryClient, {
  mutationKey: ['refreshToken'],
  mutationFn: (token: string) =>
    apiClient.authenticationService.RefreshToken({
      grant_type: 'refresh_token',
      refresh_token: token ?? '',
    }),
  retry: 0,
});
```

**修改后**：

```typescript
export const refreshTokenMutation = queryClient.getMutationCache().build(queryClient, {
  mutationKey: ['refreshToken'],
  mutationFn: (params: { refreshToken: string; userId: number; jti: string }) =>
    apiClient.authenticationService.RefreshToken({
      grant_type: 'refresh_token',
      refresh_token: params.refreshToken ?? '',
      user_id: params.userId,
      jti: params.jti,
    }),
  retry: 0,
});
```

**同步修改 store 中的调用**：`frontend/admin/react/src/stores/auth.ts`（第 211-241 行）

> ⚠️ **重要修正**：userId 和 jti **必须从当前 Access Token 的 payload 解码得到，不能从 `userInfo` 取**。原因：`userInfo` 不被持久化（见 `auth.ts` 的 `partialize`，仅持久化 token 相关字段），页面 F5 刷新后 `userInfo` 为 `null`，若依赖它会导致刷新失败 → 被踢出。而 AT 的 payload 即使已过期仍可正常解码（不验签），从中取出 `uid`（user_id）和 `jti` 即可。详见下方 `decodeAccessTokenIdentity`。

**修改前**：

```typescript
refreshToken: async () => {
    const { refreshTokenValue: refreshVal } = get();
    if (!refreshVal) {
        get().forceLogout();
        return '';
    }
    try {
        const response = await refreshTokenMutation.execute(refreshVal);
        ...
    }
}
```

**修改后**：

```typescript
refreshToken: async () => {
    const { refreshTokenValue: refreshVal, accessToken } = get();
    if (!refreshVal) {
        get().forceLogout();
        return '';
    }

    // 身份信息（user_id、jti）从当前 AT 的 payload 解码得到。
    // AT 虽可能已过期，但 payload 仍可解码；后端刷新流程已免认证，不依赖 AT 有效性。
    // 不从 userInfo 取——userInfo 不持久化，F5 刷新后为 null。
    const identity = decodeAccessTokenIdentity(accessToken);
    if (!identity) {
        console.warn('Refresh token aborted: cannot decode identity from access token');
        get().forceLogout();
        return '';
    }

    try {
        const response = await refreshTokenMutation.execute({
            refreshToken: refreshVal,
            userId: identity.userId,
            jti: identity.jti,
        });
        ...
    }
}
```

需要新增一个工具函数 `decodeAccessTokenIdentity`，从 JWT 中同时解析 `uid`（user_id）和 `jti`（不验证签名，只解码 payload）：

```typescript
/**
 * 从 JWT（Access Token）的 payload 中解码用户身份信息（user_id、jti）。
 * 仅解码 payload，不验证签名——刷新场景下 AT 可能已过期，但其 payload 仍可解码。
 * 关键：不依赖 userInfo（不持久化，F5 后为 null）。
 */
function decodeAccessTokenIdentity(token: string | null): { userId: number; jti: string } | null {
    if (!token) return null;
    try {
        const parts = token.split('.');
        if (parts.length !== 3) return null;
        const json = decodeURIComponent(
            atob(parts[1])
                .split('')
                .map((c) => '%' + c.charCodeAt(0).toString(16).padStart(2, '0'))
                .join(''),
        );
        const payload = JSON.parse(json);
        const userId = Number(payload.uid);
        const jti = typeof payload.jti === 'string' ? payload.jti : '';
        if (!Number.isFinite(userId) || userId <= 0 || !jti) return null;
        return { userId, jti };
    } catch {
        return null;
    }
}
```

> 说明：claim 字段名 `uid` / `jti` 与后端 `backend/pkg/jwt/user_token_payload.go` 的 `ClaimFieldUserID = "uid"` 及 JWT 标准字段 `jti` 对应。

**Vue 前端同步修改**：`frontend/admin/vue-element/src/composables/use-token-refresh.ts` 和 `frontend/admin/vue-element/src/api/composables/auth.ts` 同理需在 RefreshToken 请求中补充 `user_id` 和 `jti` 参数。

> 📌 **本次实现范围说明**：Vue 端改动**本次未实施**，待后续同步。本次仅完成 **Go 后端 + React 前端**。Vue 端待办：① `refreshToken()` 中同样需从 AT payload 解码 `user_id`/`jti`（或从 `useAccessStore` 持久化的身份信息中取）；② 对应的 `onRehydrateStorage`/定时器恢复逻辑。在 Vue 端改动落地前，Vue 端用户仍会遇到原始问题。

**影响范围**：
- 后端 RefreshToken API 不再依赖 auth 中间件，可独立工作
- 前端需在 RefreshToken 请求中补充 `user_id` 和 `jti` 字段
- `doGrantTypeRefreshToken` 需添加 `resetContextForLogin` 调用以绕过隐私保护中间件

**风险**：🟡 中。安全性依赖 RT 本身的验证（Redis 原子验证 + 吊销），而非 AT 认证。这符合 OAuth2 规范中 refresh_token 刷新的标准做法。

---

### F1：页面刷新后恢复定时刷新定时器

**问题**：F5 刷新后 `startRefreshTimer()` 不会重新执行，只能依赖 401 被动刷新。

**修改文件**：`frontend/admin/react/src/stores/auth.ts`

**方案**：在 Zustand store 的 `persist` 回调 `onRehydrateStorage` 中检测已有 token，自动启动定时器。

**修改位置**：`persist()` 配置中添加 `onRehydrateStorage` 回调

**修改前**：

```typescript
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // ... 状态和动作定义
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshTokenValue: state.refreshTokenValue,
        accessTokenExpireAt: state.accessTokenExpireAt,
        refreshTokenExpireAt: state.refreshTokenExpireAt,
      }),
    },
  ),
);
```

**修改后**：

```typescript
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // ... 状态和动作定义
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshTokenValue: state.refreshTokenValue,
        accessTokenExpireAt: state.accessTokenExpireAt,
        refreshTokenExpireAt: state.refreshTokenExpireAt,
      }),
      // 页面刷新后，Zustand 从 localStorage 恢复数据后触发此回调
      onRehydrateStorage: () => {
        return (state, error) => {
          if (error) {
            console.error('Auth store rehydration failed:', error);
            return;
          }
          if (state?.accessToken && state?.refreshTokenValue) {
            // AT 未过期 → 启动定时刷新
            if (state.accessTokenExpireAt && state.accessTokenExpireAt > Date.now()) {
              console.log('🔁 Rehydrated: starting refresh timer');
              startRefreshTimer();
            } else if (state.refreshTokenExpireAt && state.refreshTokenExpireAt > Date.now()) {
              // AT 已过期但 RT 未过期 → 立即刷新
              console.log('🔁 AT expired but RT valid: refreshing immediately');
              state.refreshToken().then(() => {
                startRefreshTimer();
              }).catch(() => {
                state.forceLogout();
              });
            } else {
              // RT 也过期 → 清除状态
              console.warn('🔁 Both tokens expired on rehydration');
              state.forceLogout();
            }
          }
        };
      },
    },
  ),
);
```

**影响范围**：仅在页面刷新时自动恢复定时器，不影响正常登录/登出流程。

**风险**：⬇️ 低。纯前端改动，不涉及 API 变更。

---

### F2：RefreshToken 请求不携带过期 Access Token

**问题**：RefreshToken 请求会经过 `useTokenInterceptor` 注入可能已过期的 AT，虽然后端白名单化后不再校验 AT，但携带过期 AT 在语义上不正确，且可能导致某些网关/WAF 拦截。

**修改文件**：`frontend/admin/react/src/core/transport/rest/types.ts`、`request-client.ts`、`bootstrap.ts`

**方案**：遵循现有「业务层 → 基础设施层」的回调注入架构，在 `RequestClientCallbacks` 中增加可选谓词 `shouldSkipAuth?(url)`，由 `bootstrap.ts` 注入判断逻辑（哪些 URL 免认证），`request-client.ts` 的 token 拦截器据此跳过 Authorization 注入。不在 transport 层硬编码业务路径。

**修改 1 — `types.ts`**：`RequestClientCallbacks` 新增字段

```typescript
/**
 * 是否跳过 Authorization header 注入（免认证请求）。
 * 用于刷新令牌等接口：此时 AT 可能已过期，注入过期的 AT 既无意义又可能被网关/WAF 拦截。
 */
shouldSkipAuth?: (url: string) => boolean;
```

**修改 2 — `request-client.ts`**（`useTokenInterceptor`）：

```typescript
private useTokenInterceptor(callbacks: RequestClientCallbacks) {
    this.addRequestInterceptor({
      fulfilled: (config) => {
        // 免认证请求（如刷新令牌）不注入 Authorization header
        if (callbacks.shouldSkipAuth?.(config.url ?? '')) {
          return config as never;
        }
        if (callbacks.getToken) {
          const token = callbacks.getToken();
          config.headers.Authorization = this.formatToken(token);
        }
        return config as never;
      },
    });
}
```

**修改 3 — `bootstrap.ts`**（`RequestClient.init` 注入谓词）：

```typescript
RequestClient.init(baseURL, {
    // ...其它回调
    // 刷新令牌接口已置于后端白名单免认证，且此时 AT 可能已过期，故跳过 Authorization 注入。
    shouldSkipAuth: (url) => url?.includes('/refresh-token') ?? false,
});
```

**影响范围**：仅影响被 `shouldSkipAuth` 命中的请求，其他请求不受影响。

**风险**：⬇️ 低。纯前端改动，且 B2 修改后后端不再校验刷新请求的 AT。

---

## 四、修改优先级与依赖关系

```
B2 (白名单) ─────→ B3 (服务端方法) ────→ F2 (前端跳过AT注入)
      │                    │
      │                    └──→ 前端 RefreshToken 请求补充 user_id/jti
      │
      └──── B2 和 B3 必须同时修改，否则白名单化后服务端取不到用户信息

B1 (AT过期时间) ──→ 独立修改，无依赖
F1 (定时器恢复) ──→ 独立修改，无依赖
```

**推荐修改顺序**：

1. ✅ **第一批（核心）**：B2 + B3 + 前端 RefreshToken 参数补全 + F2 — 解决核心死锁问题
2. ✅ **第二批（体验）**：B1 + F1 — 延长 AT 有效期 + 恢复定时器

---

## 五、验证清单

修改完成后，按以下场景逐项验证：

| # | 场景 | 预期结果 | 验证方法 |
|---|------|---------|---------|
| V1 | 正常登录 | 登录成功，AT 有效期 2 小时 | 检查 Redis 中 AT key 的 TTL |
| V2 | AT 过期后刷新 | 自动刷新成功，不跳转登录页 | 等待 AT 过期（或调短测试），操作页面 |
| V3 | 页面刷新（AT 未过期） | 刷新后保持登录状态，定时器恢复 | F5 刷新页面 |
| V4 | 页面刷新（AT 已过期，RT 未过期） | 自动用 RT 刷新，恢复登录状态 | 等 AT 过期后 F5 |
| V5 | 持续操作不中断 | 长时间操作不出现 401 | 连续操作超过 2 小时 |
| V6 | RT 过期后 | 跳转登录页，提示重新登录 | 等 RT 过期（或调短测试） |
| V7 | 主动登出 | 清除所有 token，跳转登录页 | 点击登出按钮 |
| V8 | 并发请求时 AT 过期 | 只刷新一次，其他请求排队等待 | 模拟多请求并发 |

---

## 六、安全考量

| 关注点 | 说明 |
|--------|------|
| RefreshToken API 白名单化是否安全？ | ✅ 安全。RT 本身是高熵随机字符串，验证依赖 Redis 中的 RT 值比对 + 原子吊销（Lua 脚本），而非 AT 认证。这是 OAuth2 标准做法。 |
| AT 延长到 2 小时是否安全？ | ✅ 可接受。AT 有 JWT 签名保护 + Redis 在库校验 + 黑名单机制，可随时吊销。且 RT 7 天过期提供了兜底。 |
| RT 传输是否需加密？ | ✅ 已通过 HTTPS 传输。RT 仅在登录响应和刷新请求中出现，不存储在 Cookie 中。 |
| RT 被盗用风险？ | 🟡 低风险。RT 每次使用后即被吊销并换发新 RT（rotation），盗用者只有一次机会。且 Redis 中 RT 与 user_id + jti 绑定，无法跨用户使用。 |

---

## 七、涉及文件汇总

### 后端

| 文件 | 修改项 |
|------|--------|
| `backend/app/admin/service/internal/data/authenticator.go` | B1：AT 过期时间 |
| `backend/app/admin/service/internal/server/rest_server.go` | B2：白名单 |
| `backend/app/admin/service/internal/service/authentication_service.go` | B3：RefreshToken 方法 |

### 前端（React）

| 文件 | 修改项 |
|------|--------|
| `frontend/admin/react/src/stores/auth.ts` | F1：定时器恢复（onRehydrateStorage）+ B3 配套：`refreshToken()` 从 AT 解码 user_id/jti + 新增 `decodeAccessTokenIdentity` |
| `frontend/admin/react/src/api/hooks/auth.ts` | B3 配套：`refreshTokenMutation` 入参变更（补 user_id/jti） |
| `frontend/admin/react/src/core/transport/rest/types.ts` | F2：`RequestClientCallbacks` 新增 `shouldSkipAuth` 谓词 |
| `frontend/admin/react/src/core/transport/rest/request-client.ts` | F2：token 拦截器调用 `shouldSkipAuth` |
| `frontend/admin/react/src/bootstrap.ts` | F2：注入 `shouldSkipAuth` 判断逻辑 |

### 前端（Vue，📌 本次未实施，待后续同步）

| 文件 | 修改项 |
|------|--------|
| `frontend/admin/vue-element/src/composables/use-token-refresh.ts` | B3 配套 + F1 |
| `frontend/admin/vue-element/src/api/composables/auth.ts` | B3 配套 |
