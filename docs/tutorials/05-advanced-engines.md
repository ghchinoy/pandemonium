# Tutorial 05: Advanced Simulation Engines (`Simultaneous` & `Asynchronous`)

In addition to the default turn-based `sequential` engine, Pandemonium implements two concurrent orchestration engines ported from Concordia v2.4:

1. **`simultaneous`**: All active players receive observations and execute `Act()` concurrently in goroutines within each step; the Game Master resolves all player actions together into a single environmental outcome.
2. **`asynchronous`**: Each player runs its own independent `Observe -> Act -> GM Resolve` loop concurrently in a dedicated goroutine at its own pace against a synchronized Game Master.

---

## Step 1: Run a Simultaneous Simulation

Create `/tmp/simultaneous_auction.yaml` with `engine: "simultaneous"`:

```bash
cat << 'EOF' > /tmp/simultaneous_auction.yaml
name: "The Sealed-Bid Auction"
description: "All bidders submit their actions simultaneously in each round."
premise: "Alice, Bob, and Charlie are bidding on a rare astrolabe in an auction hall."
world_facts:
  - "Bids in each round are revealed simultaneously."
  - "The astrolabe starts at 100 gold coins."
agents:
  - name: "Alice"
    persona: "A museum curator determined to acquire the astrolabe for public display."
    goal: "Win the astrolabe without exceeding 250 gold."
  - name: "Bob"
    persona: "A wealthy private collector who dislikes losing."
    goal: "Outbid rivals aggressively."
  - name: "Charlie"
    persona: "A cautious antiquarian looking for a bargain."
    goal: "Only bid if the price stays below 150 gold."
settings:
  engine: "simultaneous"
  min_steps: 2
  max_steps: 3
EOF
```

Run the simulation:

```bash
cd go
go run ./cmd/pandemonium \
  -scenario /tmp/simultaneous_auction.yaml \
  -log /tmp/simultaneous_auction.jsonl \
  -web ../web/dist
```

**What to Verify**:
- In each step (`Step 1`, `Step 2`, `Step 3`), **all three agents** (`Alice`, `Bob`, and `Charlie`) receive observations and emit `action` events before a single aggregated `[World] (resolution)` event is emitted.

---

## Step 2: Run an Asynchronous Simulation

Create `/tmp/async_town.yaml` with `engine: "asynchronous"`:

```bash
cat << 'EOF' > /tmp/async_town.yaml
name: "The Busy Newsroom"
description: "Reporters work independently at their own pace."
premise: "Alice and Bob are reporters in a bustling newsroom covering a breaking story."
world_facts:
  - "The printing press deadline is in 30 minutes."
agents:
  - name: "Alice"
    persona: "An investigative reporter verifying sources."
    goal: "Confirm two independent sources before filing."
  - name: "Bob"
    persona: "A fast-moving copy editor drafting headlines."
    goal: "Get the front-page headline ready."
settings:
  engine: "asynchronous"
  min_steps: 2
  max_steps: 2
EOF
```

Run the simulation:

```bash
go run ./cmd/pandemonium \
  -scenario /tmp/async_town.yaml \
  -log /tmp/async_town.jsonl \
  -web ../web/dist
```

**What to Verify**:
- `Alice` and `Bob` execute their `Observe -> Act -> Resolve` loops concurrently in separate goroutines (`Async Step N (Alice)`, `Async Step M (Bob)`), synchronized safely at the Game Master boundary.

---

## Next Step

Continue to [Tutorial 06: Running AX-Sandboxed Simulations over A2A v2](06-ax-sandboxed-simulation.md).
