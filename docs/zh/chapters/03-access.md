# 第 03 章：鉴权

这一章给服务补上一个最小但明确的访问控制层。

## 这一章解决什么问题

下游调用者是否有权访问这个代理，和代理本身如何调用上游 provider，是两件不同的事。这里先解决前者。

## 为什么上一章还不够

有了配置并不代表服务是安全的。没有访问控制，任何人都可以直接打进来。

## 新概念

- Bearer API Key 解析
- 中间件里的访问控制
- 稳定的 401 JSON 错误体

## 实现

- 起点：`chapter-02-config`
- 结束 Tag：`chapter-03-access`
- 只接受 `Authorization: Bearer <key>`
- key 做精确匹配
- 未授权统一返回 `401` 和稳定错误体

## 验证

```bash
cd nanocpa
go test ./internal/access
go test ./internal/api -run 'Test.*Unauthorized|Test.*Middleware'
```

## 你现在得到什么

- 一个能在进入业务逻辑前拦住未授权请求的服务边界

## 下一章

第 4 章会把下游 API 表面先做出来。

