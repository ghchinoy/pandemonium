# Pandemonium: Generative Social Simulation in Go

Pandemonium is a high-performance, concurrent Go implementation of the **Concordia** generative social simulation library. It allows researchers and developers to build "digital pocket universes" where LLM-powered agents interact in grounded environments.

---

## 🚀 Key Features

- **Concurrent Agent Lifecycle**: Uses goroutines to parallelize cognitive "phases" (Memory retrieval, Goal checking, etc.), ensuring high performance as agent complexity grows.
- **Entity-Component Architecture**: Pure composition-based design. Agents and Game Masters are "Entities" built from modular "Components."
- **Scenario Templates**: Define world facts, premises, and agent personas in portable YAML files.
- **Multi-Layered Observability**:
    - **TUI**: Reactive terminal interface with real-time narrative logging and agent status view.
    - **Web**: Decoupled Lit-based frontend for real-time simulation "X-Ray" via WebSockets.
    - **JSONL**: Persistent event logging for post-simulation analysis.
- **Grounding & Memory**: Built-in associative memory and immutable "World Facts" to prevent narrative drift.

![Pandemonium Architecture](docs/pandemonium_architecture.webp)

---

## 🛠 Project Structure

```text
go/
├── cmd/
│   ├── pandemonium/    # Simulation orchestrator CLI entry point
│   └── agent-node/     # Standalone A2A v2 agent microservice
└── pkg/
    ├── types/          # Core interfaces, EventBus, and Scenario config
    ├── entity/         # BaseEntity and remote A2AProxyEntity
    ├── component/      # Reusable building blocks (Memory, Timeline, Constant, Dynamic)
    ├── engine/         # Simulation loops (Sequential)
    ├── llm/            # LLM & Embedding providers (Gemini / Vertex AI)
    ├── vector/         # Vector storage (Gonum MemoryStore, FileStore, MarkdownStore)
    ├── store/          # Operational state stores (LocalStore, FirestoreStore)
    ├── prefab/         # Agent recipes, Scenario Loader, and embedded scenarios/
    └── ui/             # Bubble Tea TUI, WebSocket Server, and A2A Server
web/                    # Decoupled Lit/Vite Web X-Ray Frontend
docs/                   # Architecture guides, specifications, and tutorials/
```

---

## 🏃 Getting Started

1. **Configure Environment**:
   ```bash
   cp go/.env.example go/.env
   # Optional: set GOOGLE_CLOUD_PROJECT or GOOGLE_GENAI_API_KEY
   # (If left unset, Pandemonium runs in deterministic offline mock mode)
   ```

2. **Build Web Frontend** (optional, for Web UI support):
   ```bash
   cd web
   npm install
   npm run build
   cd ..
   ```

3. **Run a Simulation**:
   ```bash
   cd go
   go run ./cmd/pandemonium -log sim.jsonl
   ```

4. **CLI Flags**:
   - `-scenario`: Path to a YAML scenario file or built-in template name (default: `pub.yaml`).
   - `-log`: Path to save the simulation trace in JSONL format.
   - `-web`: Path to the web frontend build directory (default: `../web/dist`).

5. **Observe**:
   - **Terminal**: Interactive TUI (Press `q` to exit).
   - **Browser**: Open `http://localhost:8080` for the Web X-Ray view.

---

## 📋 Task Tracking

This project uses **bd (beads)** for issue tracking.
- Run `bd list` to see the current backlog.
- Refer to `AGENTS.md` for AI-agent workflow instructions.

---

## 📖 Documentation

- **Tutorials (Hands-On & Validation Suite)**:
  - [Tutorials & Validation Track Overview](docs/tutorials/README.md)
  - [01: Running Your First Local Simulation](docs/tutorials/01-local-simulation.md)
  - [02: Hosting & Inspecting a Standalone A2A Agent Node](docs/tutorials/02-standalone-a2a-agent.md)
  - [03: Orchestrating a Distributed Simulation over A2A](docs/tutorials/03-distributed-simulation.md)
  - [04: Custom Scenarios, Turn Strategies & Persistent Memory](docs/tutorials/04-custom-scenarios-and-memory.md)
- **Explanation & Architecture**:
  - [Project Primer](docs/PRIMER.md) - The TRPG analogy and high-level concepts.
  - [Go Adaptation Guide](docs/GO_ADAPTATION.md) - Technical implementation and roadmap.
  - [Research Insights](docs/RESEARCH_INSIGHTS.md) - Deep dive into foundational papers.
- **Reference & How-To Guides**:
  - [Scenario Specification](docs/SCENARIO_SPEC.md) - YAML schema for scenarios, agents, and storage.
  - [A2A CLI Integration Guide](docs/A2ACLI_INTEGRATION.md) - Inspecting and mocking A2A v2 agents with `a2acli`.
  - [Audio Story Recreation Guide](docs/stories/RECREATION_GUIDE.md) - Synthesizing narrated audio from simulation logs.



## Acknowledgements

This work is an adaptation of Google DeepMind's Concordia

<!-- disableFinding(SNIPPET_INVALID_LANGUAGE) -->
```bibtex
@article{vezhnevets2023generative,
  title={Generative agent-based modeling with actions grounded in physical,
  social, or digital space using Concordia},
  author={Vezhnevets, Alexander Sasha and Agapiou, John P and Aharon, Avia and
  Ziv, Ron and Matyas, Jayd and Du{\'e}{\~n}ez-Guzm{\'a}n, Edgar A and
  Cunningham, William A and Osindero, Simon and Karmon, Danny and
  Leibo, Joel Z},
  journal={arXiv preprint arXiv:2312.03664},
  year={2023}
}
```

## Disclaimer

This is not an officially supported Google product.