# 第 06 章：运行时骨架

这一章加入第一层真正 provider-agnostic 的运行时边界。

## 这一章解决什么问题

Model registry 已经能回答“哪个 provider 声称支持这个 model”，但系统还没有稳定的运行时接缝去承接“选哪个 auth、交给哪个 executor 执行”。

## 为什么上一章还不够

如果没有运行时骨架，下一章加 Claude 时就会把 provider 细节重新塞回 handler 或 server 启动逻辑里。

## 新概念

- `Auth`
- `Selector`
- `Executor`
- `Manager`

## 实现

- 起点：`chapter-05-model-registry`
- 结束 Tag：`chapter-06-runtime-skeleton`
- 引入 generic manager
- 让 server 从配置构造运行时 auth 并注册
- handler 通过稳定接口调用运行时，而不是知道具体 provider

## 验证

```bash
cd nanocpa
go test ./internal/auth ./internal/registry
go test ./internal/api/...
go test ./internal/runtime/executor/...
```

## 你现在得到什么

- 一个通用的运行时骨架
- handler 到 runtime 的稳定接缝
- 为下一章 concrete provider 做好的插槽

## 下一章

第 7 章会把第一个 concrete provider: Claude 接进来。

