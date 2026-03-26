# 第 07 章：Claude Provider

这一章第一次让系统真正完成一条从下游到上游再返回下游的闭环。

## 这一章解决什么问题

前面已经有了 OpenAI 风格表面、模型注册表和运行时骨架，但请求还不能真正走到一个 provider。这里用 Claude 作为第一个 concrete adapter 把整条链打通。

## 为什么上一章还不够

运行时骨架只有接口，没有实际 provider 逻辑。没有 concrete executor，请求仍然只会停在稳定的 `502 api_error`。

## 新概念

- OpenAI -> Claude 请求翻译
- Claude -> OpenAI 响应翻译
- Claude executor
- manager generic，provider concrete

## 实现

- 起点：`chapter-06-runtime-skeleton`
- 结束 Tag：`chapter-07-claude-provider`
- 把 OpenAI chat 请求翻译成 Claude messages 请求
- 用 Claude executor 调上游 `/v1/messages`
- 把 Claude 响应再翻回 OpenAI chat completion
- 显式拒绝本章不支持的能力：stream、`n != 1`、tools、tool_choice

## 验证

```bash
cd nanocpa
go test ./internal/translator ./internal/runtime/executor ./internal/api/...
```

## 你现在得到什么

- 第一个真正可用的 provider 适配器
- 一个被显式收窄并严格验证的 Claude contract

## 下一章

第 8 章会把多上游轮询和最后一层 hardening 补齐。

