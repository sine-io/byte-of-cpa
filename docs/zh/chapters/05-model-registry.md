# 第 05 章：模型注册表

这一章把“模型名”从字符串提升成路由输入。

## 这一章解决什么问题

`/v1/models` 不能再返回占位符，chat 请求也不能随便接受任意 model。系统需要一个真实的模型注册表，来源就是配置里的 provider/model 快照。

## 为什么上一章还不够

上一章虽然有了 OpenAI 风格接口，但 handler 还不知道哪些 model 真正存在。

## 新概念

- Model Registry
- 配置驱动的 `/v1/models`
- 在运行时之前先挡掉不支持的 model

## 实现

- 起点：`chapter-04-openai-surface`
- 结束 Tag：`chapter-05-model-registry`
- 从 `config.Providers` 构造 registry
- `/v1/models` 返回真实模型快照
- `chat/completions` 先挡掉不支持的 model

## 验证

```bash
cd nanocpa
go test ./internal/registry ./internal/api/...
```

## 你现在得到什么

- 配置里的模型快照真正进入了 API 行为
- `/v1/models` 和请求校验都开始依赖真实配置

## 下一章

第 6 章会补上 provider-agnostic 的运行时骨架。

