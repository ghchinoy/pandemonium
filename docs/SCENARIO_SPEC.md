# Pandemonium Scenario Specification

Pandemonium simulations are defined using portable YAML templates. These files describe the "Pocket Universe" rules, participants, and engine constraints.

---

## 📄 File Structure

```yaml
name: "Scenario Name"
description: "A high-level summary of the social dynamic."
premise: "The core conflict or setting."

world_facts:
  - "Immutable fact 1"
  - "Immutable fact 2"

agents:
  - name: "Alice"
    persona: "Who is this agent?"
    goal: "Optional primary objective."
    # url: "http://remote-bob:9001/invoke" # If remote

settings:
  engine: "sequential" # or "simultaneous", "asynchronous"
  min_steps: 4
  max_steps: 10
  strategy: "sequential" # or "rotation", "random"
```

---

## 🧩 Key Sections

### 1. Metadata (`name`, `description`)
Identifies the scenario in the **Command Center** UI.

### 2. The Premise (`premise`)
The "Flavor Text" that sets the scene. This is sent to every agent during the **Scenario Setup Phase (Step 0)**.

### 3. World Facts (`world_facts`)
A list of immutable truths maintained by the Game Master. These are used to prevent "Fact Drift" (hallucinations) during long-running simulations.

### 4. Agent Configuration (`agents`)
Describes the participants. 
- **Local Agents**: Require a `persona` (and optional `goal`). They use the built-in "Three Key Questions" prefab (`SituationPerception`, `SelfPerception`, `PersonBySituation`, plus `ReflectionComponent`).
- **`formative_memories`**: A list of strings representing pre-existing knowledge or background (e.g., "I have lived here 10 years"). These are injected into the agent's memory store at boot.
- **Remote Agents**: Use the `url` field. Pandemonium will fetch their **Agent Card** via A2A and orchestrate them over the network.
    - **`ate_target_actor`** *(optional)*: Sets the `ate-target-actor: <atespace>/<task>` HTTP header for routing through AX's `atenet-router` (which also auto-resumes suspended AX tasks on demand).
    - **`headers`** *(optional)*: A map of custom HTTP headers attached to both `AgentCard` resolution and JSON-RPC A2A requests.

### 5. Engine Settings (`settings`)
- **`engine`**:
    - `sequential`: Turn-based loop where one agent acts per step (Default).
    - `simultaneous`: All active players observe and act concurrently in goroutines within each step before aggregated GM resolution.
    - `asynchronous`: Each player runs an independent `Observe -> Act` goroutine loop at its own pace against a synchronized GM.
- **`min_steps`**: The number of turns before the GM is allowed to check for termination.
- **`max_steps`**: The hard cut-off for the simulation (or per-player iteration cap in `asynchronous` mode).
- **`strategy`** (used by `sequential` engine):
    - `sequential`: GM decides who acts next (Narrative).
    - `rotation`: Strict round-robin (Fair).
    - `random`: Random actor selection (Chaos).

---

## 💾 Persistence Strategy (`storage`)

Agents can utilize different storage engines to balance performance and auditability.

```yaml
agents:
  - name: "Alice"
    persona: "Optimistic Baker"
    storage:
      type: "markdown"
      path: "logs/alice_diary.md"
```

### Supported Types:
- **`memory`**: Pure in-memory (Default). Fast, but lost on restart.
- **`file`**: JSONL-backed. Persistent associative recall.
- **`markdown`**: Human-readable list. Best for "audit trails" and long-context models.

---
