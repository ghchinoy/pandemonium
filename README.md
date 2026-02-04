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

![Pandemonium Architecture](docs/pandemonium_architecture.png)

---

## 🛠 Project Structure

```text
go/
├── cmd/pandemonium/    # CLI entry point
├── pkg/
│   ├── types/          # Shared interfaces and Event types
│   ├── entity/         # BaseEntity and lifecycle logic
│   ├── component/      # Reusable building blocks (Memory, Constant)
│   ├── engine/         # Simulation loops (Sequential) and EventBus
│   ├── llm/            # LLM providers (Gemini/Vertex AI)
│   ├── vector/         # Vector storage and retrieval
├── prefab/         # Agent recipes and Scenario Loader
	└── ui/             # TUI and Web Server implementation
web/                    # Decoupled Lit-based Web Frontend
go/pkg/prefab/scenarios/ # Built-in simulation YAML templates
```

```

---

## 🏃 Getting Started

1. **Configure Environment**:
   ```bash
   cp go/.env.example go/.env
   # Add your GOOGLE_CLOUD_PROJECT and optional GOOGLE_GENAI_API_KEY
   ```

2. **Run a Simulation**:
   ```bash
   cd go
   go run cmd/pandemonium/main.go -log sim.jsonl
   ```

3. **CLI Flags**:
   - `-scenario`: Path to a YAML scenario file (default: `scenarios/pub.yaml`).
   - `-log`: Path to save the simulation trace in JSONL format.
   - `-web`: Path to the web frontend directory (default: `../web`).

4. **Observe**:
   - **Terminal**: Interactive TUI (Press 'q' to exit).
   - **Browser**: Open `http://localhost:8080` for the Web X-Ray view.

---

## 📋 Task Tracking

This project uses **bd (beads)** for issue tracking.
- Run `bd list` to see the current backlog.
- Refer to `GEMINI.md` for AI-agent workflow instructions.

---

## 📖 Documentation

- [Project Primer](docs/PRIMER.md) - The TRPG analogy and high-level concepts.
- [Go Adaptation Guide](docs/GO_ADAPTATION.md) - Technical implementation and roadmap.
- [Research Insights](docs/RESEARCH_INSIGHTS.md) - Deep dive into foundational papers.



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