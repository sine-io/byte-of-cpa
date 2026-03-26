# Byte of CPA 教程

Byte of CPA 是一套面向程序员的分章节教程，目标是带你一步一步做出一个最小但真实可用的 CPA。

整套教程按里程碑展开。每一章都会说明：

- 这一章解决什么问题
- 为什么上一章还不够
- 这一章引入了什么代码和架构边界
- 应该如何验证结果

## 阅读方式

1. 先看[中文导览](chapter-guide.md)。
2. 按章节阅读，理解这一章的设计动作。
3. 如果想看这一章完成后的代码快照，可以检出对应 Git tag。
4. 运行该章节列出的验证命令。

## 章节入口

- [第 01 章 启动](chapters/01-bootstrap.md)
- [第 02 章 配置](chapters/02-config.md)
- [第 03 章 鉴权](chapters/03-access.md)
- [第 04 章 OpenAI 接口面](chapters/04-openai-surface.md)
- [第 05 章 模型注册表](chapters/05-model-registry.md)
- [第 06 章 运行时骨架](chapters/06-runtime-skeleton.md)
- [第 07 章 Claude Provider](chapters/07-claude-provider.md)
- [第 08 章 路由与加固](chapters/08-routing-and-hardening.md)

## 本地预览

```bash
. .venv/bin/activate 2>/dev/null || python -m venv .venv && . .venv/bin/activate
python -m pip install -r requirements-docs.txt
mkdocs serve
```
