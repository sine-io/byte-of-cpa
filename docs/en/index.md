# Byte of CPA

Byte of CPA is a tutorial-first project for programmers who want to build a minimal CPA step by step.

The tutorial is organized as concrete chapter milestones. Each chapter explains:

- what problem that milestone solves
- why the previous milestone is not enough
- what code and architectural boundary are introduced
- how to verify the result

## How To Read This Site

1. Start with the [Chapter Guide](chapter-guide.md).
2. Read a chapter to understand the design move.
3. Check out that chapter's tag in Git if you want to inspect the exact snapshot.
4. Run the verification commands for that chapter.

## Milestones

- [Chapter 01 Bootstrap](chapters/01-bootstrap.md)
- [Chapter 02 Config](chapters/02-config.md)
- [Chapter 03 Access](chapters/03-access.md)
- [Chapter 04 OpenAI Surface](chapters/04-openai-surface.md)
- [Chapter 05 Model Registry](chapters/05-model-registry.md)
- [Chapter 06 Runtime Skeleton](chapters/06-runtime-skeleton.md)
- [Chapter 07 Claude Provider](chapters/07-claude-provider.md)
- [Chapter 08 Routing and Hardening](chapters/08-routing-and-hardening.md)

## Local Preview

```bash
. .venv/bin/activate 2>/dev/null || python3 -m venv .venv && . .venv/bin/activate
python -m pip install -r requirements-docs.txt
mkdocs serve
```
