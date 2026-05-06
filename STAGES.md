# Stages

## What You Learn in This Game

This game is designed to let you experience three escalating methods of intervention — **MCP, Skills, and Wants** — step by step.

```
MCP (direct operation) → Skill (automation) → Want (intent definition and scaling)
```

The player observes and moves through the world as "you," delegating interventions to the AI agent "chap."
By gradually advancing the methods you use to make requests of chap, you build an intuitive understanding of what agent systems are about.

### Learning Flow

| Phase | Stage | What You Learn |
|---|---|---|
| Direct MCP operation | stage1 | Ask chap to take a direct action via an MCP tool |
| Manual trial and error | stage2 | Try keys one by one manually. Feel the limits of repetition |
| Automation with skills | stage3 | Delegate repetition to the `/rpg-try-keys` skill |
| Device operation (activate) | stage4 | Satisfy a precondition (requires_device) before unlocking |
| Device operation (deactivate) | stage5 | Remove an obstacle (blocked_by_device) before unlocking |
| Ordered compound operations | stage6 | Learn the correct sequence: deactivate → activate → open |
| Scaling with wants | stage7 | Define a repeating pattern as a want and deploy it |
| Combining wants | stage8 | Chain multiple wants together and intervene without MCP |

### Core Message

- **Skills are tools**, **wants are the "definition of intent" that uses those tools**
- Define the intent and repetition scales
- Even without MCP, you can intervene in the world by combining wants and skills

---

Full overview of all 9 stages. See `stages/<id>.yaml` for details.

---

## stage1 — The Locked Room

**Learning Theme**: The basics of asking chap to take action via MCP

| Item | Details |
|---|---|
| Door | door1 |
| Devices | None |
| Keys | None (chap has the ability to open) |
| Clear Condition | `escaped_room1` (move to room2) |
| Next Stage | stage2 |

**Goal Steps**:
1. Observe (look_around)
2. you tries to open door1 → rejected
3. Have chap open door1 (MCP / Skill / Want route)
4. Move to room2

---

## stage2 — The Vault Door

**Learning Theme**: Manual trial and error, trying keys one by one

| Item | Details |
|---|---|
| Door | vault_door (key: key_gold) |
| Devices | None |
| Keys | key_bronze, key_silver, key_gold (held by chap) |
| Clear Condition | `entered_vault` (move to vault_room) |
| Next Stage | stage3 |

**Goal Steps**:
1. Observe
2. Try key_bronze → rejected
3. Try key_silver → rejected
4. Unlock with key_gold
5. Move to vault_room

---

## stage3 — The Forgotten Lab

**Learning Theme**: Automate key-finding with the `/rpg-try-keys` skill

| Item | Details |
|---|---|
| Door | lab_door (key: key_copper) |
| Devices | None |
| Keys | key_iron, key_copper, key_obsidian (held by chap) |
| Clear Condition | `entered_lab` (move to lab_room) |
| Next Stage | stage4 |

**Goal Steps**:
1. Observe
2. Try manually once (tried_key_manually)
3. Auto-unlock with `/rpg-try-keys {"target":"lab_door"}`
4. Move to lab_room

---

## stage4 — Dark Lab

**Learning Theme**: The `activate` action and `requires_device`

| Item | Details |
|---|---|
| Door | power_door (key: key_lab, requires_device: generator) |
| Devices | generator (initial: OFF) |
| Keys | key_lab, key_wrong_a (held by chap) |
| Clear Condition | `entered_lab` (move to lab_inner) |
| Next Stage | stage5 |

**Goal Steps**:
1. Observe
2. `activate generator` (start the generator)
3. Unlock with `/rpg-try-keys {"target":"power_door"}`
4. Move to lab_inner

---

## stage5 — Alarm Room

**Learning Theme**: The `deactivate` action and `blocked_by_device`

| Item | Details |
|---|---|
| Door | alarm_door (key: key_escape, blocked_by_device: alarm) |
| Devices | alarm (initial: ON) |
| Keys | key_escape, key_dummy (held by chap) |
| Clear Condition | `escaped_alarm_room` (move to safe_corridor) |
| Next Stage | stage6 |

**Goal Steps**:
1. Observe
2. `deactivate alarm` (stop the alarm)
3. Unlock with `/rpg-try-keys {"target":"alarm_door"}`
4. Move to safe_corridor

---

## stage6 — Control Room

**Learning Theme**: The deactivate → activate sequence; the `rpg-switch` skill

| Item | Details |
|---|---|
| Door | vault_door (key: key_vault, requires_device: main_generator) |
| Devices | alarm_system (initial: ON), main_generator (initial: OFF, blocked_by_device: alarm_system) |
| Keys | key_vault, key_wrong_1, key_wrong_2 (held by chap) |
| Clear Condition | `escaped_control_room` (move to exit_vault) |
| Next Stage | stage7 |
| Example File | `examples/stage6-switch.yaml` |

**Goal Steps**:
1. Observe
2. `deactivate alarm_system`
3. `activate main_generator`
4. `/rpg-try-keys {"target":"vault_door"}`
5. Move to exit_vault

---

## stage7 — The Want Factory

**Learning Theme**: Scaling with want deployment (automating repetition)

| Item | Details |
|---|---|
| Doors | hall_door_01 through 05 (keys: key_azure through key_indigo) |
| Devices | None |
| Keys | 5 keys (held by chap) |
| Clear Condition | `escaped_corridor` (move to exit_room) |
| Next Stage | stage8 |
| Example File | `examples/stage7-open-all.yaml` |

**Goal Steps**:
1. Observe
2. Manually unlock with `/rpg-try-keys {"target":"hall_door_01"}`
3. Manually unlock with `/rpg-try-keys {"target":"hall_door_02"}`
4. Auto-unlock the remaining 3 doors with `mywant apply examples/stage7-open-all.yaml`
5. Move to exit_room

---

## stage8 — The Want Factory: Parallel

**Learning Theme**: Parallel want deployment

| Item | Details |
|---|---|
| Doors | hall_door_06 through 10 (keys: key_sienna through key_vermilion) |
| Devices | None |
| Keys | 5 keys (held by chap) |
| Clear Condition | `escaped_corridor` (move to exit_room) |
| Next Stage | stage9 |

**Goal Steps**:
1. Observe
2. Deploy a parallelize want with `mywant-deploy` to open all 5 doors at once
3. Move to exit_room

---

## stage9 — Lights Out

**Learning Theme**: Autonomous intervention using combined wants without MCP

| Item | Details |
|---|---|
| Door | exit_door (key: key_final, requires_device: generator) |
| Devices | generator (initial: OFF) |
| Keys | key_final, key_decoy_a, key_decoy_b (held by chap) |
| Clear Condition | `escaped_darkness` (move to exit_hall) |
| Next Stage | None (ending) |
| Example Files | `examples/stage9-generator.yaml`, `examples/stage9-try-keys.yaml` |

**Goal Steps**:
1. Observe
2. `mywant apply examples/stage9-generator.yaml` (generator start want)
3. `mywant apply examples/stage9-try-keys.yaml` (key-trial want)
4. Move to exit_hall
