# Tutorial 02: Hosting & Inspecting a Standalone A2A Agent Node

In this tutorial, you will run a Pandemonium entity as a standalone **A2A v2** microservice using `cmd/agent-node` and interact with it directly using [`a2acli`](https://github.com/ghchinoy/a2acli) (or `curl`).

---

## Step 1: Start the Standalone Agent Node

Open **Terminal 1** and start an agent node named `Bob` on port `9001`:

```bash
cd go
go run ./cmd/agent-node \
  -port 9001 \
  -name "Bob" \
  -persona "A skeptical journalist investigating the winter storm."
```

**Expected Output in Terminal 1**:

```text
Starting Pandemonium Agent Node [Bob]...
[A2A] Hosting agent Bob at http://localhost:9001
```

---

## Step 2: Verify the Published AgentCard

Open **Terminal 2** and inspect the agent's `/.well-known/agent-card.json` endpoint.

### Using `a2acli`:
```bash
a2acli discover --service-url http://localhost:9001 --output json
```

### Or using `curl`:
```bash
curl -s http://localhost:9001/.well-known/agent-card.json | jq .
```

**Checkpoint**: Verify that:
- `name` is `"Bob"`.
- `description` is `"Pandemonium Agent: Bob"`.
- `capabilities.streaming` is `true`.
- `supportedInterfaces` lists `http://localhost:9001/invoke` with protocol `JSONRPC`.

---

## Step 3: Send an Action Prompt & Stream Reasoning Artifacts

In **Terminal 2**, send a message to `Bob` with `--stream` to follow the live reasoning artifact events (`JSONL` in `--output json` mode):

```bash
a2acli send "What do you notice about the pub around you?" \
  --service-url http://localhost:9001 \
  --stream \
  --output json
```

**What to Verify**:
1. **In Terminal 1 (`agent-node`)**:
   - Logs `[Bob Thinking] Bob is considering their action...` as `BaseEntity.Act` executes its parallel `PRE_ACT` components.
2. **In Terminal 2 (`a2acli`)**:
   - Receives `TaskArtifactUpdateEvent` payloads containing Bob's internal component contexts (`Instructions`, `Memory`, `Timeline`, `Reflection`, `SituationPerception`, `SelfPerception`, `PersonBySituation`).
   - Completes with a final `TaskStatusUpdateEvent` (`COMPLETED`) containing Bob's action message.

---

## Next Step

Keep `cmd/agent-node` running in Terminal 1 and continue to [Tutorial 03: Orchestrating a Distributed Simulation over A2A](03-distributed-simulation.md).
