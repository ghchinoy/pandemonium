# Tutorial 01: Running Your First Local Simulation

In this tutorial, you will build the Web X-Ray frontend, run the built-in **Snowed-in Pub** (`pub.yaml`) simulation, and validate all three observability layers: the **Bubble Tea TUI**, the **Web X-Ray Inspector**, and the **JSONL Event Logger**.

---

## Step 1: Build the Web X-Ray Frontend

From the repository root, build the Lit/Vite web bundle so the Go server can serve it at `http://localhost:8080`:

```bash
cd web
npm install
npm run build
cd ..
```

**Checkpoint**: Verify that `web/dist/index.html` exists.

---

## Step 2: Launch the Simulation Orchestrator

Run the orchestrator from the `go/` directory with the `-log` flag enabled:

```bash
cd go
go run ./cmd/pandemonium -scenario pub.yaml -log /tmp/pandemonium_pub.jsonl -web ../web/dist
```

> **Tip**: Leave `GOOGLE_CLOUD_PROJECT` and `GOOGLE_GENAI_API_KEY` unset to run in deterministic offline mock mode, or export either variable to use live Gemini models.

---

## Step 3: Verify the Terminal UI (TUI)

While the simulation is running in your terminal, verify the following in the Bubble Tea interface:

1. **Header & Viewport**:
   - The top banner displays `PANDEMONIUM SOCIAL SIMULATION`.
   - The narrative log begins with `✨ Simulation initializing... (start)` and `--- Setup: Setting the Scene ---`.
2. **Agents Sidebar**:
   - The right-hand `AGENTS` panel lists `● Alice`, `● Bob`, and `● Charlie`.
   - As each turn progresses, the active agent's status transitions through `Observing...` → `Thinking...` → `Acting` → `Ready`.
3. **Turn Progression**:
   - Each step logs `[Actor] (observation)`, `[Actor] (thought)`, `[Actor] (action)`, and `[World] (resolution)`.

---

## Step 4: Verify the Web X-Ray Dashboard

While the TUI is active (before pressing `q`), open **`http://localhost:8080`** in your browser:

1. **Agent Cards**: Verify cards appear for `Alice`, `Bob`, and `Charlie`.
2. **Simulation Log**: Verify real-time WebSocket events stream into the central feed.
3. **Trace Inspector**: Click an agent or thought event to inspect the underlying component contexts gathered during `PRE_ACT` (`Instructions`, `Memory`, `Timeline`, `SelfPerception`, `SituationPerception`).

---

## Step 5: Exit and Verify the JSONL Event Trace

1. Press **`q`** (or `Ctrl+C`) in the terminal to exit the TUI cleanly.
2. Inspect the generated JSONL log file:

```bash
head -n 6 /tmp/pandemonium_pub.jsonl
```

**Expected Output**: Each line is a valid JSON object representing a `types.Event`:

```json
{"type":"observation","actor":"Alice","message":"Registered","data":{"isRemote":false}}
{"type":"observation","actor":"Bob","message":"Registered","data":{"isRemote":false}}
{"type":"observation","actor":"Charlie","message":"Registered","data":{"isRemote":false}}
{"type":"sim_start","message":"Simulation initializing...","data":"Alice, Bob, and Charlie are stuck in a snowed-in pub called 'The Rusty Anchor'. There is only one cookie left."}
{"type":"step_start","message":"Setup: Setting the Scene","data":0}
```

---

## Next Step

Continue to [Tutorial 02: Hosting & Inspecting a Standalone A2A Agent Node](02-standalone-a2a-agent.md).
