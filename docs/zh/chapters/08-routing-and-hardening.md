# 第 08 章：路由与加固

这一章把教程推到第一版终点：它不再只是一个“能转发 Claude 的适配器”，而是一个真正有路由价值的最小 CPA。

## 这一章解决什么问题

它让多个上游实例可以服务同一个 model，并把下游错误边界收成稳定可预期的样子。

## 为什么上一章还不够

上一章虽然已经能通过 Claude 端到端完成一次请求，但它更像单连接桥接，不像一个真正的代理层。

## 新概念

- 按模型维度的 round-robin
- 每个模型独立维护游标
- 稳定的 upstream/runtime 错误归一化
- 端到端路由测试

## 实现

- 起点：`chapter-07-claude-provider`
- 结束 Tag：`chapter-08-routing-and-hardening`
- 对同一 model 的多个 auth 做 deterministic round-robin
- disabled/cooldown auth 会被跳过
- unsupported model 仍然是 `400 invalid_request_error`
- 已配置但当前无可用 auth 的模型会落到稳定的 `502 api_error`

## 验证

```bash
cd nanocpa
go test ./internal/auth ./internal/api -run 'TestManager_|TestServer_'
go test ./...
```

## 你现在得到什么

- 一个真正可讲清楚、也能跑起来的最小 CPA
- 配置驱动的模型快照
- provider-agnostic 的 runtime
- 第一个 concrete provider
- 多上游路由和稳定错误边界

## 教程为什么停在这里

这就是第一版教程要交付的最小完整闭环。再往后就会开始进入平台工程层面，而不是一个紧凑的学习项目。

仍然刻意没有做的能力：

- 加权或延迟感知路由
- cooldown 调度与健康管理
- 持久化运行时状态
- 更复杂的 provider retry / circuit breaker
- metrics、tracing、dashboard
- streaming、tools、更广的 provider 覆盖
