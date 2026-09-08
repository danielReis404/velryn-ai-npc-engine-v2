# Velryn — AI Agent Engine

**Velryn** is an experimental persistent-world simulation built with Go, exploring how AI-driven NPCs could behave in a dynamic game world.

Instead of relying entirely on predefined dialogue trees and scripted behaviors, NPCs can perceive their surroundings, maintain memories, consider their goals and relationships, and use LLMs to make decisions when a situation requires contextual reasoning.

The project is inspired by the concept of highly autonomous NPCs explored in *Shangri-La Frontier* and was created to experiment with a possible future of AI-driven game worlds.

> **The world was never meant to be understood.**

---

## ✨ What is Velryn?

Velryn is not intended to be a complete game. It is an **AI and backend engineering experiment** focused on exploring the architecture required to combine:

* Autonomous NPC behavior
* LLM-based decision making
* Persistent long-term memory
* Semantic memory retrieval
* World simulation
* NPC relationships and goals
* Deterministic game mechanics
* AI agent reasoning
* Multiple LLM providers
* Real-time world observation

The central idea is to keep the game world deterministic while allowing AI agents to reason about situations that benefit from contextual decision making.

A simplified representation of the architecture is:

```text
                    VELRYN WORLD
                         │
                         ▼
                  NPC Perception
                         │
                         ▼
                  World Context
                         │
              ┌──────────┴──────────┐
              │                     │
       Routine / Rules        AI Reasoning
              │                     │
              │              Memory Retrieval
              │                     │
              │                  LLM Call
              │                     │
              └──────────┬──────────┘
                         ▼
                  Structured Action
                         │
                         ▼
                  World Simulation
                         │
                         ▼
                   New World State
```

The LLM does **not** directly control the game world.

It produces a structured decision that is interpreted, validated, and executed by the Go engine.

---

## 🧠 AI-Driven NPCs

Each NPC has its own state and characteristics, including:

* Personality
* Background
* Goals
* Routines
* Relationships
* Trust and fear
* Memories
* Inventory
* Economy
* Combat behavior
* Current location
* Perception of nearby entities

NPCs do not need to call an LLM for every action.

Routine and deterministic behaviors are handled directly by the engine. AI reasoning is reserved for situations where contextual decision making can add value.

For example:

```text
NPC is walking toward a destination
        │
        ▼
Deterministic routine
        │
        ▼
No AI call required
```

But when something meaningful happens:

```text
NPC encounters another character
        │
        ▼
Relevant situation detected
        │
        ▼
Retrieve relevant memories
        │
        ▼
Build contextual prompt
        │
        ▼
LLM reasoning
        │
        ▼
Structured action
        │
        ▼
Go engine executes the action
```

This separation helps reduce unnecessary LLM calls while keeping the world responsive and predictable.

---

## 🧠 Persistent NPC Memory

NPC experiences can be stored as long-term memories.

When a relevant event occurs, the system generates an embedding for the memory and stores it in PostgreSQL using `pgvector`.

Later, when an NPC encounters a relevant situation, the engine can search for semantically similar memories and provide the most relevant ones as context for the LLM.

```text
NPC Experience
      │
      ▼
Memory Creation
      │
      ▼
Gemini Embedding
      │
      ▼
PostgreSQL + pgvector
      │
      ▼
Semantic Similarity Search
      │
      ▼
Relevant Memories
      │
      ▼
LLM Context
```

The complete world lore is also maintained separately from the condensed representation used by the engine. This avoids sending the entire world history to the LLM on every request and keeps context usage under control.

See [`docs/lore.md`](docs/lore.md) for the complete world lore.

---

## ⚙️ Key Features

### Autonomous NPC decision making

NPCs can use LLM reasoning to determine actions based on:

* Current surroundings
* Previous experiences
* Relationships
* Personality
* Goals
* Current world state

### Long-term memory

NPC experiences can persist beyond a single interaction through embedding-based semantic retrieval.

### Multiple LLM providers, decided outside Go

NPC and world decisions are no longer made by an `llm/` package inside this
repository — that logic moved into a sibling service, **[`agent-service/`](../agent-service)**,
a Python/LangGraph/FastAPI app the engine calls over HTTP for every decision
(`/decide`, `/decide-batch`, `/world-decide`). It runs its own fallback chain:

```text
Groq → Gemini → Cerebras → Mistral → local Ollama (last resort)
```

If a provider fails or is rate-limited, it's put in cooldown and the next
one in the chain is tried — see `agent-service/app/llm/router.py`.

The engine still talks to an LLM provider directly in exactly one place:
Gemini, for NPC memory embeddings (see below) — that's unrelated to NPC
decision-making and stays in Go because it's a simple, synchronous call with
no fallback chain or tool use involved.

### Gemini embeddings

Google Gemini is used as the embedding provider for NPC memories.

### Deterministic combat

Combat does not require an LLM for every action.

Combat behavior such as attacks, fleeing, revival, trust, and fear can be handled directly by the engine.

This keeps frequent game mechanics predictable and avoids unnecessary model calls.

### Batch decision making

NPCs sharing the same scene can be grouped so that related decisions can be processed together when appropriate — resolved via `agent-service`'s `/decide-batch`, which runs every member of the group concurrently instead of one call at a time.

### Dirty-state detection

NPC perception uses a fingerprint-based mechanism to determine whether the relevant environment has actually changed enough to justify new reasoning.

### Real-time observation

The simulation exposes the current world state through HTTP/SSE.

The simulation loop runs while there are active viewers observing the world.

When there are no active viewers, the simulation pauses and avoids unnecessary AI and database activity.

### Persistent world state

NPCs, memories, items, relationships, and other relevant state are persisted through PostgreSQL.

### Human player

One NPC can be flagged `IsPlayerControlled` and driven by a real person instead of the AI or the simulation's own routines — see `POST /api/npcs/{name}/act` in `server/api.go`. The engine loop leaves that NPC idle (`WAIT`) until an action arrives from the API, applying it immediately instead of waiting for the next tick (`ApplyExternalAction` in `engine/public_api.go`).

---

# 🏗️ Architecture

This `engine/` folder is one of three services in the **Velryn monorepo** — see
the [root README](../README.md) for how they fit together. Inside `engine/`
itself, the simulation is organized around clear responsibilities rather than
placing the entire thing inside a single file.

```text
engine/                    (this folder — the Go simulation + HTTP/SSE server)
│
├── main.go
│
├── engine/
│   ├── world.go
│   ├── loop.go
│   ├── turn.go
│   ├── perception.go
│   ├── prompt.go
│   ├── actions.go
│   ├── combat.go
│   ├── movement.go
│   ├── geometry.go
│   ├── memory.go
│   ├── economy.go
│   ├── dialogue.go
│   ├── spawn.go
│   ├── lore.go
│   ├── agent_client.go        # HTTP client → agent-service /decide, /decide-batch
│   ├── world_agent_client.go  # HTTP client → agent-service /world-decide
│   ├── public_api.go          # PerceiveNPC / ApplyExternalAction (used by server/api.go)
│   ├── types.go               # NPCPerception + related DTOs sent to agent-service
│   ├── api_types.go           # WorldSnapshot / WorldEvent DTOs
│   ├── intent_source.go       # IntentSource enum (ai, instinct, scripted, external, ...)
│   └── world_event_types.go   # WorldEventKind / Weather / TimeOfDay enums
│
├── models/
│   ├── npc.go
│   ├── bestiary.go
│   └── action.go              # ActionVerb enum (MOVE, TALK, ATTACK, ...)
│
├── database/
│   └── PostgreSQL persistence
│
├── embeddings/
│   └── Gemini embedding client
│
├── server/
│   ├── server.go               # HTTP/SSE server
│   ├── api.go                  # /api/npcs/{name}/act — player actions
│   └── static/                 # frontend
│
├── docs/
│   └── lore.md
│
└── supabase/
    └── migrations/
```

### Engine

The `engine` package contains the core simulation logic:

* World state
* Game loop
* NPC turns
* Perception
* Prompt construction
* Action resolution
* Movement
* Combat
* Memory
* Economy
* Dialogue
* Spawning
* World lore

The original implementation had most of the simulation logic concentrated in a single large file. It was subsequently separated by responsibility to make the codebase easier to navigate, maintain, and evolve.

---

## 🔄 NPC Decision Flow

A simplified NPC turn looks like this:

```text
                     ┌──────────────┐
                     │   NPC Turn   │
                     └──────┬───────┘
                            │
                            ▼
                    Perceive Environment
                            │
                            ▼
                     Has World Changed?
                       /           \
                     No             Yes
                     │               │
                     ▼               ▼
                 Skip AI       Classify Situation
                                     │
                                     ▼
                              Relevant Memories
                                     │
                                     ▼
                               Build Context
                                     │
                                     ▼
                                LLM Request
                                     │
                                     ▼
                            Structured Intent
                                     │
                                     ▼
                              Validate Intent
                                     │
                                     ▼
                              Execute Action
                                     │
                                     ▼
                              Update World
```

The important architectural principle is that **LLMs are used for reasoning, not for basic simulation**.

The Go engine remains responsible for enforcing the rules of the world.

---

# 🌍 The World of Velryn

Velryn is a persistent fantasy world built around exploration, discovery, and emergent interactions.

The world contains:

* Ancient civilizations
* Ruins
* Wildlands
* Independent cities
* Guilds
* Unique creatures
* Dynamic ecosystems
* Ancient artifacts known as Echoes
* Unexplored territories
* Mysteries surrounding the disappearance of the Architects

Rather than traditional dungeon-based progression, the world itself acts as the unexplored frontier.

The player's knowledge of the world becomes part of progression.

The complete world lore is available in [`docs/lore.md`](docs/lore.md).

---

# 🛠️ Technology Stack (this `engine/` service)

| Area                    | Technology               |
| ----------------------- | ------------------------ |
| Language                | Go                       |
| NPC/world decisions     | HTTP calls to `agent-service` (see root README) |
| Embeddings              | Google Gemini            |
| Database                | PostgreSQL               |
| Vector search           | pgvector                 |
| Backend                 | Go HTTP server           |
| Real-time communication | Server-Sent Events (SSE) |
| Database platform       | Supabase                 |
| Configuration           | Environment variables    |

See [`agent-service/README.md`](../agent-service/README.md) for its stack (Python, FastAPI, LangGraph, LangChain, Groq/Gemini/Cerebras/Mistral/Ollama).

---

# 🚀 Running Locally

This `engine/` server needs `agent-service` reachable to make any NPC
decisions — running `go run main.go` alone will start the world, but every
NPC will fall back to routine/instinct behavior (no LLM calls) until
`agent-service` is up. Full local setup, all three services, is documented
once in the [root README](../README.md#running-locally-all-three-services) —
this section only covers `engine/` on its own.

## Requirements

Before running the project, make sure you have:

* Go installed (see `go.mod` for the version)
* PostgreSQL/Supabase database (optional — the engine runs in-memory-only without it)
* Gemini API key for embeddings (optional — only needed for semantic memory recall)
* `agent-service` running locally (see [`agent-service/README.md`](../agent-service/README.md)) — required for AI-driven NPC decisions

### 1. Clone the repository

```bash
git clone <repository-url>
cd velryn/engine
```

### 2. Configure environment variables

```bash
cp .env.example .env
```

Then fill in `AGENT_SERVICE_URL` (defaults to `http://localhost:8000`, which
matches `agent-service`'s default port), and optionally `GEMINI_API_KEY` /
`SUPABASE_DB_URL` if you want persistent memory. Never commit `.env` itself.

### 3. Install dependencies

```bash
go mod tidy
```

### 4. Apply database migrations (optional, only if using Postgres)

The database schema is versioned under:

```text
supabase/migrations/
```

Apply the migrations in chronological order, or use the Supabase CLI to push the schema.

### 5. Run the engine

```bash
go run main.go
```

The server starts on:

```text
http://localhost:8080
```

Open the application in a browser to observe the simulation.

The world simulation runs while there are active viewers. Closing the application pauses the simulation and prevents unnecessary AI calls.

---

# 🧪 Testing

```bash
go test ./...
```

Tests cover the deterministic, dependency-free parts of the engine:
geometry/adjacency (`engine/geometry_test.go`), the perception fingerprint
used for dirty-state detection (`engine/perception_test.go`), trust/fear
bookkeeping (`engine/combat_test.go`), shop/inventory lookups
(`engine/economy_test.go`), `resolveIntent`'s MOVE handling
(`engine/actions_test.go`), and the `ActionVerb` JSON contract
(`models/action_test.go`). They construct a real `World` via `NewWorld(...)`
rather than mocking it — the constructor has no external dependencies (no
DB, no HTTP), so this is cheap and exercises real code paths.

What isn't covered here: anything that needs a live `agent-service` or
database connection (those are integration-level, not unit-level), and the
SSE/HTTP server itself.

---

# 📊 Design Considerations

One of the main goals of Velryn is to explore the practical constraints of AI-driven simulations.

LLM-powered NPCs introduce several challenges:

### Cost

Calling an LLM for every NPC on every simulation tick would quickly become expensive.

The engine therefore uses deterministic behavior, dirty-state detection, batching, and conditional reasoning to reduce unnecessary calls.

### Latency

LLM responses are significantly slower than local game logic.

The architecture therefore keeps the simulation and game rules inside the Go engine while treating LLM calls as an external reasoning process.

### Context

Providing too much information to an LLM increases token usage and can reduce the relevance of the context.

Velryn therefore retrieves relevant memories instead of sending an NPC's entire history and maintains a condensed representation of the world's lore.

### Reliability

External AI providers can fail or become temporarily unavailable.

Multiple providers and fallback handling allow the system to continue operating when possible.

---

# 🎯 Project Goals

Velryn is primarily an exploration of the intersection between:

**Game Simulation + Backend Engineering + Generative AI + Agentic Systems**

The project aims to investigate questions such as:

* How autonomous can NPCs realistically become?
* How should long-term memory work for game characters?
* When should an LLM be involved in a simulation?
* What should remain deterministic?
* How can AI costs be controlled?
* How can agent decisions be validated before affecting the world?
* How should persistent AI agents interact with a changing environment?
* What architectural patterns are necessary for large numbers of AI-driven entities?

The project is intentionally experimental.

The goal is not to claim that fully autonomous AI NPCs are solved, but to explore the engineering challenges involved in building them.

---

# 🔮 Future Development

Potential directions for future versions include:

* LangChain/LangGraph-based agent orchestration
* More sophisticated agent state management
* Tool-based agent interactions
* Improved memory retrieval
* More complex NPC relationships
* Dynamic world events
* Larger-scale NPC simulations
* Improved observability
* Better AI decision evaluation
* Separation between deterministic world simulation and AI agent services
* Further optimization of LLM usage and cost

---

# 📚 Documentation

* [World Lore](docs/lore.md)
* Architecture documentation — coming as the project evolves

---

# ⚠️ Disclaimer

Velryn is a personal experimental project created to explore AI-driven game simulations, autonomous agents, memory systems, and backend architecture.

It is not intended to represent a production-ready game engine.

API usage and model behavior depend on external providers and their respective availability, limits, and pricing.
