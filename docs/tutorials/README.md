# Pandemonium Tutorials & Validation Track

These step-by-step tutorials follow the [Diátaxis](https://diataxis.fr/tutorials/) learning-oriented pattern while doubling as a **reproducible manual validation suite** for the Pandemonium application.

---

## Prerequisite: Automated Test & Benchmark Verification

Before running the interactive tutorials, you can verify that all packages, race checks, A2A protocol interop tests, and vector search benchmarks pass:

```bash
cd go
go test -v -race ./...
go test -bench=. -benchtime=100ms ./pkg/vector/...
```

### Offline Mock Mode vs. Live Gemini Mode

Pandemonium supports two execution modes out of the box:
- **Deterministic Mock Mode (Zero-Config)**: If `GOOGLE_CLOUD_PROJECT` and `GOOGLE_GENAI_API_KEY` are **unset**, `llm.GeminiModel` and `llm.GeminiEmbedder` automatically run in deterministic local mock mode. Every tutorial below works offline without API keys.
- **Live Gemini Mode**: Set `GOOGLE_CLOUD_PROJECT` (for Vertex AI) or `GOOGLE_GENAI_API_KEY` (for the Gemini API) in `go/.env` or your shell environment to run with live `gemini-3-flash-preview` and `gemini-embedding-001` calls.

---

## Tutorial Track

| # | Tutorial | What You Will Build & Validate |
| :--- | :--- | :--- |
| **01** | [Running Your First Local Simulation](01-local-simulation.md) | Run the built-in *Snowed-in Pub* scenario (`pub.yaml`), observe the Bubble Tea TUI, inspect thoughts in the Lit Web X-Ray dashboard (`:8080`), and verify the JSONL event trace. |
| **02** | [Hosting & Inspecting a Standalone A2A Agent Node](02-standalone-a2a-agent.md) | Launch `cmd/agent-node` as an A2A v2 microservice, discover its `AgentCard` with `a2acli`, and stream cognitive reasoning artifacts over JSON-RPC. |
| **03** | [Orchestrating a Distributed Simulation over A2A](03-distributed-simulation.md) | Connect a local orchestrator (`cmd/pandemonium`) to a remote agent node (`cmd/agent-node`) using `distributed_pub.yaml` and verify cross-process trace tunneling. |
| **04** | [Custom Scenarios, Turn Strategies & Persistent Memory](04-custom-scenarios-and-memory.md) | Author a custom scenario YAML using `rotation` turn-taking, `formative_memories`, persistent `file` (JSONL) and `markdown` memory stores, and JSON checkpointing (`-checkpoint` / `-resume`). |
| **05** | [Advanced Simulation Engines (`Simultaneous` & `Asynchronous`)](05-advanced-engines.md) | Run multi-actor concurrent rounds (`engine: "simultaneous"`) and independent per-entity goroutine loops (`engine: "asynchronous"`). |

---

## Related Documentation (Diátaxis Map)

- **Explanation**: [Project Primer](../PRIMER.md) · [Go Adaptation Guide](../GO_ADAPTATION.md) · [Research Insights](../RESEARCH_INSIGHTS.md)
- **Reference**: [Scenario Specification](../SCENARIO_SPEC.md) · [A2A CLI Integration Guide](../A2ACLI_INTEGRATION.md)
- **How-To Guides**: [Audio Story Recreation Guide](../stories/RECREATION_GUIDE.md)
