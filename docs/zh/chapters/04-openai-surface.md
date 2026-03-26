# 第 04 章：OpenAI 接口面

这一章先把“对客户端长什么样”固定下来，再去接上游。

## 这一章解决什么问题

它定义了 `POST /v1/chat/completions` 和 `GET /v1/models` 这两个 OpenAI 风格入口，包括请求校验、错误格式和返回形状。

## 为什么上一章还不够

上一章只有访问控制，还没有客户端真正能调用的 API 面。

## 新概念

- OpenAI 风格 payload
- handler 级别的请求校验
- 请求体大小限制
- 稳定的错误 envelope

## 实现

- 起点：`chapter-03-access`
- 结束 Tag：`chapter-04-openai-surface`
- 注册 `chat/completions` 和 `models`
- 对 JSON、必填字段、体积做校验
- 即使还没有上游执行，也先返回稳定的 API 错误

## 验证

```bash
cd nanocpa
go test ./internal/api/... -run 'TestOpenAI|TestChatCompletions|TestModels'
```

## 你现在得到什么

- 一个真正“长得像 OpenAI API”的下游表面
- 即便还没接上游，也已经把协议边界固定住

## 下一章

第 5 章会让 `/v1/models` 和模型校验变成配置驱动。

