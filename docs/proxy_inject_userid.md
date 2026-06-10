# InjectFromId Feature Design

## Summary

Add a per-account toggle `inject_userid_in_proxy` stored in Account's `extra` JSONB field. When enabled, the feature reads the `X-Proxy-User-Id` header from incoming requests and injects it into the proxy URL's username field. This allows the downstream proxy server to identify each caller.

## Requirements

- The downstream caller provides `X-Proxy-User-Id` via request header
- Toggle lives on Account's `extra` field (not Proxy entity)
- If `X-Proxy-User-Id` is missing, skip injection — use proxy URL as-is
- All proxy protocols supported: SOCKS5, SOCKS5h, HTTP, HTTPS

## Data Model

The setting is stored in Account's existing `extra` JSONB field:

```json
{
  "inject_userid_in_proxy": true
}
```

No schema migration required since `extra` is already a JSONB field.

## URL Injection Logic

Method `InjectFromIdURL(injectEnabled bool, fromId string) string` on the Proxy service object.

Behavior:
- If `injectEnabled == false` or `fromId == ""`, return `p.URL()` unchanged
- If proxy has username + password: inject `username@fromId` as the URL user, keep password
- If proxy has username only (no password): inject `username@fromId` as the URL user
- If proxy has no username: return `p.URL()` unchanged (nothing to inject into)

Transformation table:

| Stored fields | inject_userid_in_proxy | X-Proxy-User-Id | Result |
|---|---|---|---|
| socks5, user=alice, pass=secret | true | 42 | `socks5://alice@42:secret@host:port` |
| socks5, user=alice, pass="" | true | 42 | `socks5://alice@42@host:port` |
| socks5, user="", pass=secret | true | 42 | `socks5://host:port` (no injection) |
| socks5, user="", pass="" | true | 42 | `socks5://host:port` (no injection) |
| *(any protocol)* | false | *(any)* | `p.URL()` (unchanged) |
| *(any protocol)* | true | *(empty)* | `p.URL()` (unchanged) |

Same pattern applies to socks5h, http, https protocols.

## Relay Integration

Wherever the proxy URL is used, call `account.Proxy.NewURL(c, account)` instead of `account.Proxy.URL()`.

This method internally:
1. Checks if account has `inject_userid_in_proxy` enabled via `account.IsInjectUserIdInProxyEnabled()`
2. Reads `X-Proxy-User-Id` from the incoming request header: `c.GetHeader("X-Proxy-User-Id")`
3. Calls `InjectFromIdURL(injectEnabled, fromId)` with the values

Example:
```go
proxyURL = account.Proxy.NewURL(c, account)
```

The shared client pool keys by the full modified URL, so each unique FromId value gets its own cached client. This is a deliberate tradeoff: simplicity over optimal connection reuse. The pool fragmentation is acceptable when FromId values are reused frequently (same caller making many requests).

## Frontend

Add a boolean toggle "Inject X-Proxy-User-Id into Proxy" on the Account edit form (near the proxy selector). Maps to `account.extra.inject_userid_in_proxy`.

## Files Modified

- `backend/internal/service/account.go` — add `IsInjectUserIdInProxyEnabled()` helper method
- `backend/internal/service/proxy.go` — add `InjectFromIdURL()` and `NewURL()` methods
- Relay layer files — use `account.Proxy.NewURL(c, account)` instead of `account.Proxy.URL()`
- `frontend/src/components/account/EditAccountModal.vue` — add toggle
- `frontend/src/i18n/locales/en.ts` and `zh.ts` — add translations