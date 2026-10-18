# A2A CLI (`a2acli`) & Pandemonium Integration Guide

This How-To guide describes how to use [`a2acli`](https://github.com/ghchinoy/a2acli) alongside Pandemonium to inspect, stream, debug, and mock A2A (Agent-to-Agent) services in generative social simulations.

Both **Pandemonium** and **a2acli** are aligned on the **`a2a-go/v2` SDK** (`v2.6.0`, Go `1.26`), providing full wire and schema compatibility over the A2A v1.0 protocol specification.

---

## Installation

Install `a2acli` via Homebrew or `go install`:

```bash
# macOS / Linux (Homebrew)
brew tap ghchinoy/tap
brew install a2acli

# Or via Go (any platform)
go install github.com/ghchinoy/a2acli/cmd/a2acli@latest
```

---

## Architecture Overview

Pandemonium exposes and consumes A2A protocol endpoints through two core components:

1. **`ui.StartA2AServer`**: Hosts any Pandemonium `Entity` as an A2A service (JSON-RPC over HTTP) with an auto-generated `AgentCard` at `/.well-known/agent-card.json`.
2. **`entity.A2AProxyEntity`**: Delegates a Pandemonium entity's `Observe` and `Act` calls to a remote A2A service via `a2aclient` (with optional `ate-target-actor` and custom HTTP headers).

`a2acli` serves as the CLI client, test harness, and mock server for these A2A endpoints.

```text
+-------------------+                    +------------------------+
|   a2acli client   |  --- JSON-RPC ---> | Pandemonium agent-node |
| (discover / send) |  <--- Events ----  |  (ui.StartA2AServer)   |
+-------------------+                    +------------------------+

+--------------------+                   +--------------------+
| Pandemonium Engine |                   |   a2acli serve     |
| (A2AProxyEntity)   |  --- JSON-RPC --->|   (--echo mock)    |
+--------------------+                   +--------------------+
```

---

## Scenario 1: Inspecting & Debugging Pandemonium Agent Nodes

When running distributed simulation nodes (e.g. `cmd/agent-node`), `a2acli` provides AgentCard discovery, blocking task execution, live `--stream` artifact inspection, and multi-turn `--context` correlation.

### 1. Start a Pandemonium Agent Node

```bash
cd go
go run ./cmd/agent-node -port 9001 -name Bob -persona "A curious investigator"
```

### 2. Discover Agent Capabilities

Inspect the agent's published `AgentCard`, supported transports, and capabilities:

```bash
a2acli discover --service-url http://localhost:9001
```

### 3. Send a Message (Blocking vs. Live Thought Streaming)

By default, `a2acli send` blocks until the task reaches a terminal state (`COMPLETED`) and prints the result:

```bash
a2acli send "What do you observe in the room?" --service-url http://localhost:9001
```

To follow live `TaskStatusUpdateEvent` and `TaskArtifactUpdateEvent` updates (`Instructions`, `Goal`, `Memory`, `Timeline`, `Reflection`, `SituationPerception`, `SelfPerception`, `PersonBySituation`) as they stream from `cmd/agent-node`, pass **`--stream`**:

```bash
a2acli send "What do you observe in the room?" --service-url http://localhost:9001 --stream
```

### 4. Multi-Turn Session Correlation (`--context`)

Pandemonium groups turns in a simulation under a shared `ContextID`. You can simulate a multi-turn conversation thread from `a2acli` using `--context`:

```bash
a2acli send "The fireplace is dying out." --service-url http://localhost:9001 --context sim-session-1
a2acli send "What should we do next?" --service-url http://localhost:9001 --context sim-session-1 --stream
```

### 5. Non-Interactive / Scripting Mode (`--output json`)

For test scripts or agentic drivers, pass `--output json` (emits a single JSON result when blocking, or JSONL one-event-per-line when combined with `--stream`):

```bash
# Single JSON result
a2acli send "State your primary goal." --service-url http://localhost:9001 --output json

# Streaming JSONL events (one JSON object per line)
a2acli send "State your primary goal." --service-url http://localhost:9001 --stream --output json
```

---

## Scenario 2: Mock Remote Agents for Distributed Simulations

During scenario testing, you can use `a2acli serve --echo` as a zero-dependency remote agent without requiring live LLM API keys.

### 1. Start a Mock Agent

```bash
a2acli serve --echo --port 9002
```

### 2. Configure Scenario YAML

In your scenario file (e.g. `go/pkg/prefab/scenarios/distributed_painting.yaml`), set the remote agent URL to the `a2acli` endpoint:

```yaml
agents:
  - name: "Alice"
    persona: "Local bar owner"
  - name: "Bob"
    persona: "Remote patron"
    url: "http://localhost:9002/invoke"
```

### 3. Run Pandemonium Simulation

```bash
cd go
go run ./cmd/pandemonium -scenario distributed_painting.yaml
```

Pandemonium's `A2AProxyEntity` sends observation messages and action prompts to `a2acli serve`, which reflects them back as completed task responses.

---

## Scenario 3: Automated Conformance & Smoke Testing

### Running Conformance Smoke Checks

Validate that a Pandemonium agent node complies with A2A protocol specifications:

```bash
a2acli conformance --service-url http://localhost:9001
```

### CI Integration

Pandemonium includes a GitHub Actions workflow (`.github/workflows/ci.yml`) that installs `a2acli` and runs full unit and interop integration tests (`go test -v ./pkg/ui`) verifying A2A client/server round-trips.
