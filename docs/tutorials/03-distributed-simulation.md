# Tutorial 03: Orchestrating a Distributed Simulation over A2A

In this tutorial, you will run a **hybrid distributed simulation** (`distributed_pub.yaml`) where **Alice** runs in-process inside the orchestrator while **Bob** runs in a separate OS process over the **A2A v2** protocol.

![Distributed A2A Flow](../pandemonium_a2a_flow.webp)

---

## Step 1: Ensure the Remote Agent Node is Running

In **Terminal 1**, make sure `Bob` is running on port `9001`:

```bash
cd go
go run ./cmd/agent-node \
  -port 9001 \
  -name "Bob" \
  -persona "A visiting journalist investigating the storm."
```

---

## Step 2: Inspect the Distributed Scenario Template

The built-in scenario `go/pkg/prefab/scenarios/distributed_pub.yaml` configures `Alice` with a local `persona` and `Bob` with a remote A2A `url`:

```yaml
name: "Distributed Pub Scenario"
description: "A hybrid simulation where Alice is local and Bob is a remote A2A agent."
premise: "Alice and Bob are discussing the storm in 'The Rusty Anchor'. Bob is connected via a remote A2A node."
world_facts:
  - "The location is 'The Rusty Anchor' pub."
  - "Alice is the local baker."
  - "Bob is a visiting journalist."
agents:
  - name: "Alice"
    persona: "An optimistic baker."
  - name: "Bob"
    url: "http://localhost:9001/invoke"
settings:
  min_steps: 2
  max_steps: 5
```

---

## Step 3: Run the Orchestrator Against the Remote Node

In **Terminal 2**, start the Pandemonium orchestrator with `distributed_pub.yaml`:

```bash
cd go
go run ./cmd/pandemonium \
  -scenario distributed_pub.yaml \
  -log /tmp/pandemonium_distributed.jsonl \
  -web ../web/dist
```

---

## Step 4: Verify Distributed Handshake & Trace Tunneling

1. **AgentCard Resolution at Startup**:
   - `entity.NewA2AProxyEntity` resolves `http://localhost:9001/.well-known/agent-card.json` via `agentcard.NewResolver` before starting Step 0.
   - Try stopping Terminal 1 and re-running Step 3: verify that Pandemonium fails fast with a clear error (`resolving agent card for "Bob" at "http://localhost:9001": ...`).
2. **Remote Indicator in TUI**:
   - In Terminal 2's `AGENTS` sidebar, verify that `Alice` appears as `● Alice` while `Bob` is badged as a remote cloud agent: **`● Bob ☁️`**.
3. **Cross-Process Observation & Thought Tunneling**:
   - During Step 0 (`Setup: Setting the Scene`), `Bob` receives the premise via an A2A `SendMessage` request (`metadata.type = "observation"`).
   - When `Bob` is chosen to act, Terminal 1 prints `[Bob Thinking] Bob is considering their action...`, streams his reasoning components back as `TaskArtifactUpdateEvent`s, and `A2AProxyEntity` republishes them onto the orchestrator's local `EventBus`.

---

## Next Step

Continue to [Tutorial 04: Custom Scenarios, Turn Strategies & Persistent Memory](04-custom-scenarios-and-memory.md).
