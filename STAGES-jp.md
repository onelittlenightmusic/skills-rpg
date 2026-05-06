# Stages

## このゲームで学ぶこと

このゲームは **MCP・Skill・Want という3つの介入手段** を段階的に体験させることを目的とする。

```
MCP（直接操作）→ Skill（自動化）→ Want（意図の定義・スケール）
```

プレイヤーは「you」として世界を観察・移動し、AIエージェント「chap」に介入を依頼する。
chapへの依頼手段を徐々に高度化することで、エージェントシステムの本質を体で覚える。

### 学習の流れ

| フェーズ | ステージ | 学ぶこと |
|---|---|---|
| MCP直接操作 | stage1 | chapにMCPツールで直接アクションを依頼する |
| 手動試行錯誤 | stage2 | 鍵を1本ずつ手動で試す。繰り返しの限界を感じる |
| スキルによる自動化 | stage3 | `/rpg-try-keys` スキルに繰り返しを任せる |
| デバイス操作（activate） | stage4 | 前提条件（requires_device）を満たしてから開錠する |
| デバイス操作（deactivate） | stage5 | 障害（blocked_by_device）を除去してから開錠する |
| 順序のある複合操作 | stage6 | deactivate → activate → open の正しい順番を学ぶ |
| wantによるスケール | stage7 | 繰り返しパターンをwantとして定義・デプロイする |
| wantの組み合わせ | stage8 | 複数のwantを連携させ、MCP不要で世界に介入する |

### 核心メッセージ

- **スキルは道具**、**wantはその道具を使う「意図の定義」**
- 意図を定義すれば繰り返しはスケールする
- MCPがなくても、wantとスキルの組み合わせで世界に介入できる

---

全8ステージの概要。詳細は `stages/<id>.yaml` を参照。

---

## stage1 — The Locked Room

**学習テーマ**: MCP経由でchapに操作を依頼する基本

| 項目 | 内容 |
|---|---|
| 扉 | door1 |
| デバイス | なし |
| 鍵 | なし（chapがopen能力を持つ） |
| クリア条件 | `escaped_room1`（room2へ移動） |
| 次ステージ | stage2 |

**ゴール手順**:
1. 観察（look_around）
2. youがdoor1を開けようとする → rejected
3. chapにdoor1を開けさせる（MCP / Skill / Want 経路）
4. room2へ移動

---

## stage2 — The Vault Door

**学習テーマ**: 鍵を1本ずつ手動で試す試行錯誤

| 項目 | 内容 |
|---|---|
| 扉 | vault_door（key: key_gold） |
| デバイス | なし |
| 鍵 | key_bronze, key_silver, key_gold（chapが保持） |
| クリア条件 | `entered_vault`（vault_roomへ移動） |
| 次ステージ | stage3 |

**ゴール手順**:
1. 観察
2. key_bronze で試す → rejected
3. key_silver で試す → rejected
4. key_gold で開錠
5. vault_roomへ移動

---

## stage3 — The Forgotten Lab

**学習テーマ**: `/rpg-try-keys` スキルで鍵探しを自動化

| 項目 | 内容 |
|---|---|
| 扉 | lab_door（key: key_copper） |
| デバイス | なし |
| 鍵 | key_iron, key_copper, key_obsidian（chapが保持） |
| クリア条件 | `entered_lab`（lab_roomへ移動） |
| 次ステージ | stage4 |

**ゴール手順**:
1. 観察
2. 手動で1回試す（tried_key_manually）
3. `/rpg-try-keys {"target":"lab_door"}` で自動開錠
4. lab_roomへ移動

---

## stage4 — Dark Lab

**学習テーマ**: `activate` アクションと `requires_device`

| 項目 | 内容 |
|---|---|
| 扉 | power_door（key: key_lab, requires_device: generator） |
| デバイス | generator（初期: OFF） |
| 鍵 | key_lab, key_wrong_a（chapが保持） |
| クリア条件 | `entered_lab`（lab_innerへ移動） |
| 次ステージ | stage5 |

**ゴール手順**:
1. 観察
2. `activate generator`（発電機起動）
3. `/rpg-try-keys {"target":"power_door"}` で開錠
4. lab_innerへ移動

---

## stage5 — Alarm Room

**学習テーマ**: `deactivate` アクションと `blocked_by_device`

| 項目 | 内容 |
|---|---|
| 扉 | alarm_door（key: key_escape, blocked_by_device: alarm） |
| デバイス | alarm（初期: ON） |
| 鍵 | key_escape, key_dummy（chapが保持） |
| クリア条件 | `escaped_alarm_room`（safe_corridorへ移動） |
| 次ステージ | stage6 |

**ゴール手順**:
1. 観察
2. `deactivate alarm`（アラーム停止）
3. `/rpg-try-keys {"target":"alarm_door"}` で開錠
4. safe_corridorへ移動

---

## stage6 — Control Room

**学習テーマ**: deactivate → activate の順番、`rpg-switch` スキル

| 項目 | 内容 |
|---|---|
| 扉 | vault_door（key: key_vault, requires_device: main_generator） |
| デバイス | alarm_system（初期: ON）、main_generator（初期: OFF, blocked_by_device: alarm_system） |
| 鍵 | key_vault, key_wrong_1, key_wrong_2（chapが保持） |
| クリア条件 | `escaped_control_room`（exit_vaultへ移動） |
| 次ステージ | stage7 |
| 例ファイル | `examples/stage6-switch.yaml` |

**ゴール手順**:
1. 観察
2. `deactivate alarm_system`
3. `activate main_generator`
4. `/rpg-try-keys {"target":"vault_door"}`
5. exit_vaultへ移動

---

## stage7 — The Want Factory

**学習テーマ**: want のデプロイによるスケール（繰り返しの自動化）

| 項目 | 内容 |
|---|---|
| 扉 | hall_door_01〜10（各鍵: key_azure〜key_vermilion） |
| デバイス | なし |
| 鍵 | 10本（chapが保持） |
| クリア条件 | `escaped_corridor`（exit_roomへ移動） |
| 次ステージ | stage8 |
| 例ファイル | `examples/stage7-open-all.yaml` |

**ゴール手順**:
1. 観察
2. `/rpg-try-keys {"target":"hall_door_01"}` で手動開錠
3. `/rpg-try-keys {"target":"hall_door_02"}` で手動開錠
4. `mywant apply examples/stage7-open-all.yaml` で残り8枚を自動開錠
5. exit_roomへ移動

---

## stage8 — Lights Out

**学習テーマ**: want の組み合わせによる自律的な世界への介入

| 項目 | 内容 |
|---|---|
| 扉 | exit_door（key: key_final, requires_device: generator） |
| デバイス | generator（初期: OFF） |
| 鍵 | key_final, key_decoy_a, key_decoy_b（chapが保持） |
| クリア条件 | `escaped_darkness`（exit_hallへ移動） |
| 次ステージ | なし（エンディング） |
| 例ファイル | `examples/stage8-generator.yaml`, `examples/stage8-monitor.yaml`, `examples/stage8-try-keys.yaml` |

**ゴール手順**:
1. 観察
2. `mywant apply examples/stage8-generator.yaml`（発電機起動want）
3. `mywant apply examples/stage8-monitor.yaml`（監視want）
4. `mywant apply examples/stage8-try-keys.yaml`（鍵試行want）
5. exit_hallへ移動
