# 第 02 章：配置

这一章把“能启动”扩展成“能由数据驱动地启动”。

## 这一章解决什么问题

配置是部署数据和运行时行为之间的边界。`host`、`port`、下游 `api_keys`、上游 `providers` 都应该来自 YAML，而不是硬编码在 Go 代码里。

## 为什么上一章还不够

第 1 章只有最小启动配置。没有严格验证的话，错误配置会一路漏到运行时，最后以更难排查的方式失败。

## 新概念

- YAML 配置加载
- 配置归一化与 fail-fast 校验
- 把 provider 声明为数据，而不是代码分支

## 实现

- 起点：`chapter-01-bootstrap`
- 结束 Tag：`chapter-02-config`
- 扩展配置字段：`host`、`port`、`api_keys`、`providers[]`
- provider 需要 `id`、`provider`、`api_key`、`base_url`、`models`
- 启动前就拒绝不完整或不合法的配置

## 验证

```bash
cd nanocpa
go test ./internal/config -run 'TestLoad|TestValidate'
go test ./internal/config
```

## 你现在得到什么

- 配置成为系统的真实边界
- 错误配置会尽早失败
- Provider 快照可以独立于运行时代码演化

## 下一章

第 3 章会把“谁能调用这个服务”这件事补上。

