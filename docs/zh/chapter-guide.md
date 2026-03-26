# 中文导览

这份导览把教程拆成 8 个可检出的里程碑。每一章都对应一个明确的开始参考点、结束 tag 和验证命令。

## 章节顺序

| 章节 | 起点 | 结束 Tag | 验证命令 |
| --- | --- | --- | --- |
| 第 01 章 启动 | 基线提交 `62f02a2` | `chapter-01-bootstrap` | `cd nanocpa && go test ./internal/api -run 'TestServer_'` |
| 第 02 章 配置 | `chapter-01-bootstrap` | `chapter-02-config` | `cd nanocpa && go test ./internal/config -run 'TestLoad|TestValidate'`，以及 `cd nanocpa && go test ./internal/config` |
| 第 03 章 鉴权 | `chapter-02-config` | `chapter-03-access` | `cd nanocpa && go test ./internal/access`，以及 `cd nanocpa && go test ./internal/api -run 'Test.*Unauthorized|Test.*Middleware'` |
| 第 04 章 OpenAI 接口面 | `chapter-03-access` | `chapter-04-openai-surface` | `cd nanocpa && go test ./internal/api/... -run 'TestOpenAI|TestChatCompletions|TestModels'` |
| 第 05 章 模型注册表 | `chapter-04-openai-surface` | `chapter-05-model-registry` | `cd nanocpa && go test ./internal/registry ./internal/api/...` |
| 第 06 章 运行时骨架 | `chapter-05-model-registry` | `chapter-06-runtime-skeleton` | `cd nanocpa && go test ./internal/auth ./internal/registry`，`cd nanocpa && go test ./internal/api/...`，`cd nanocpa && go test ./internal/runtime/executor/...` |
| 第 07 章 Claude Provider | `chapter-06-runtime-skeleton` | `chapter-07-claude-provider` | `cd nanocpa && go test ./internal/translator ./internal/runtime/executor ./internal/api/...` |
| 第 08 章 路由与加固 | `chapter-07-claude-provider` | `chapter-08-routing-and-hardening` | `cd nanocpa && go test ./internal/auth ./internal/api -run 'TestManager_|TestServer_'`，以及 `cd nanocpa && go test ./...` |

## 推荐流程

1. 从本章的起点开始。
2. 阅读章节文档，理解这一章解决的问题和引入的边界。
3. 检出本章结束 Tag，查看完成后的快照。
4. 运行验证命令。
5. 继续下一章。

