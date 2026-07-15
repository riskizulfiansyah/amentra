# Amentra

[![Go backend](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/riski/ai-chat/gh-pages/coverage/backend.json)](https://github.com/riski/ai-chat/actions/workflows/ci.yml)
[![Frontend](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/riski/ai-chat/gh-pages/coverage/frontend.json)](https://github.com/riski/ai-chat/actions/workflows/ci.yml)

> **Intelligence that adapts to your context**

*Amentra* — from *Amenti* (ancient Egyptian), a hidden realm of knowledge. An invisible intelligence layer that connects users to the right knowledge, in the right context.

| Part | Meaning |
|---|---|
| **Ament** | hidden / behind-the-scenes |
| **-ra** | energy / system / activation |

Scoped multi-app AI platform — monorepo with Go backend + frontend web component.

## Structure

```
backend/       — Go API server (see backend/README)
frontend/      — SPA frontend (see frontend/README)
```

## System Architecture

```mermaid
graph LR
    subgraph Frontend["Frontend (Browser)"]
        direction TB
        W["Amentra Widget"]
        CACHE["Response Cache<br/>n-gram fuzzy · 30min TTL"]
    end
    subgraph Backend["Backend (Go)"]
        API["HTTP Server<br/>internal/server"]
        SVC["Chat Service<br/>internal/chat"]
        LLM["LLM Client<br/>internal/llm"]
    end
    LLM_API["LLM API<br/>OpenAI-compatible"]

    W --> CACHE
    CACHE -->|"miss →"| API
    CACHE -->|"hit → instant reply"| W
    API -->|"HTTP / SSE"| SVC
    SVC --> LLM
    LLM --> LLM_API
```


```
  Frontend                Backend (Go)              External
 ┌──────────────┐       ┌──────────────────┐      ┌──────────┐
 │ @amentra/    │       │ HTTP Server      │      │  LLM API │
 │  react  vue  │       │ internal/server  │      │ (OpenAI- │
 │    ↘   ↙     │       └──────┬───────────┘      │ compat)  │
 │  Amentra     │──HTTP─→      │                   └────▲─────┘
 │  Widget      │   /SSE       │                        │
 │              │              ↓                        │
 └──────────────┘       ┌──────────────┐                │
                        │ Chat Service │──OpenAI────────┘
                        │ internal/    │
                        │   chat       │
                        └──────────────┘
```


## Quick Start

```bash
cd backend && make dev
```

## Backend

Config-driven, stateless, streaming chat API backed by any OpenAI-compatible LLM. See [backend/README.md](backend/README.md).

## Frontend

[Amentra Widget](frontend/README.md) — embeddable Lit web component with React and Vue wrappers.
