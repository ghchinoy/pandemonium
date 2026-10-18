# Tutorial 04: Custom Scenarios, Turn Strategies & Persistent Memory

In this tutorial, you will author a custom scenario YAML file that exercises:
- **Deterministic Turn Rotation** (`strategy: "rotation"`)
- **Formative Memories** (`formative_memories` seeded at startup)
- **Persistent Vector & Audit Storage** (`file` JSONL store and `markdown` audit log)

---

## Step 1: Create a Custom Scenario YAML

Create `/tmp/library_scenario.yaml`:

```bash
cat << 'EOF' > /tmp/library_scenario.yaml
name: "The Midnight Archive"
description: "Testing rotation strategy, formative memories, and persistent stores."
premise: "Archivist Clara and Historian Dev are cataloging a newly discovered vault of maps."
world_facts:
  - "The archive vault is underground and lit by brass lamps."
  - "A locked cedar chest sits on the central table."
agents:
  - name: "Clara"
    persona: "A meticulous archivist who values preservation above all."
    storage:
      type: "file"
      path: "/tmp/pandemonium_clara_memory.jsonl"
    formative_memories:
      - "Clara holds the brass key to the upper library."
      - "Clara previously restored a 17th-century nautical chart."
  - name: "Dev"
    persona: "An eager historian looking for clues about a lost expedition."
    storage:
      type: "markdown"
      path: "/tmp/pandemonium_dev_diary.md"
    formative_memories:
      - "Dev's mentor disappeared searching for the Northern Passage."
settings:
  min_steps: 2
  max_steps: 4
  strategy: "rotation"
EOF
```

---

## Step 2: Run the Custom Scenario

From the `go/` directory, run Pandemonium pointing `-scenario` at `/tmp/library_scenario.yaml`:

```bash
cd go
rm -f /tmp/pandemonium_clara_memory.jsonl /tmp/pandemonium_dev_diary.md
go run ./cmd/pandemonium \
  -scenario /tmp/library_scenario.yaml \
  -log /tmp/pandemonium_library.jsonl \
  -web ../web/dist
```

Press **`q`** once the simulation reaches `Simulation Complete (end)`.

---

## Step 3: Verify Turn Rotation and Persistent Memory Files

### 1. Verify Strict Turn Rotation (`Clara` → `Dev` → `Clara` → `Dev`)
Inspect the `action` events in `/tmp/pandemonium_library.jsonl`:

```bash
grep '"type":"action"' /tmp/pandemonium_library.jsonl
```

**Expected Output**: Actors strictly alternate (`Clara`, then `Dev`, then `Clara`, etc.) rather than waiting for the GM to pick the next actor.

### 2. Verify Clara's JSONL Vector Store (`/tmp/pandemonium_clara_memory.jsonl`)
Inspect Clara's persisted JSONL memory file:

```bash
head -n 4 /tmp/pandemonium_clara_memory.jsonl
```

**Expected Output**:
- The first two lines are Clara's `formative_memories` with their embedding vectors.
- Subsequent lines record the scenario premise, GM observations, and Clara's own actions (`"I decided to: ..."`).

### 3. Verify Dev's Markdown Audit Store (`/tmp/pandemonium_dev_diary.md`)
Inspect Dev's human-readable Markdown memory file:

```bash
cat /tmp/pandemonium_dev_diary.md
```

**Expected Output**: A Markdown bullet list (`- ...`) containing Dev's formative memory, incoming observations, and actions.

---

## Step 4: Verify Simulation Checkpointing & Resume (`-checkpoint` / `-resume`)

Pandemonium supports snapshotting simulation state (step progress, turn-rotation index, `TimelineMemory`, and `MemoryStore`) to JSON after each step and resuming from any checkpoint:

```bash
# 1. Run and save a checkpoint after each step
go run ./cmd/pandemonium \
  -scenario /tmp/library_scenario.yaml \
  -checkpoint /tmp/library_checkpoint.json \
  -log /tmp/pandemonium_library_cp.jsonl \
  -web ../web/dist

# 2. Inspect the saved checkpoint JSON
jq '{completed_steps, last_actor, last_rotation_index}' /tmp/library_checkpoint.json

# 3. Resume from the checkpoint
go run ./cmd/pandemonium \
  -scenario /tmp/library_scenario.yaml \
  -resume /tmp/library_checkpoint.json \
  -web ../web/dist
```

**What to Verify**:
- `/tmp/library_checkpoint.json` contains `completed_steps`, `last_actor`, `last_rotation_index`, and the `entities` map with each entity's `timeline` and `memories`.
- When resumed with `-resume`, `Sequential.Run` skips Step 0 initialization and continues from `completed_steps`.
