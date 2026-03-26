# 第 01 章：启动

第 1 章只做一件事：把项目变成一个能安全启动、能监听端口的服务进程。

## 这一章解决什么问题

在任何代理能力出现之前，系统首先得是一个可靠的网络服务。它需要能读取启动配置、构造监听地址，并设置安全的 HTTP 超时。

## 为什么上一章还不够

这是起点，没有上一章。这里的目标不是代理请求，而是拿到一个稳定的服务外壳。

## 新概念

- 服务构造函数
- 启动配置
- 监听地址与超时

## 实现

- 起点：基线提交 `62f02a2`
- 结束 Tag：`chapter-01-bootstrap`
- `main.go` 读取配置路径并启动服务
- `config.go` 只保留 `host` 和 `port`
- `server.go` 只负责 HTTP Server 的构造和运行

## 验证

```bash
cd nanocpa
go test ./internal/api -run 'TestServer_'
```

可选：

```bash
cd nanocpa
go test ./internal/config -run 'TestLoad_'
```

## 你现在得到什么

- 一个可启动的 Go 服务
- 一个最小配置面
- 一个安全的 HTTP Server 基线

## 下一章

第 2 章会把最小配置面扩展成真正的数据边界。

