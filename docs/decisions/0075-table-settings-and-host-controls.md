# ADR 0075 — Table settings and host controls: who runs a table, what they can change, and spawning on a live game

**Status:** Accepted · 2026-09-19 · S35 — Playtest stabilisation, round 2
**Issues:** [#1032](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1032) (tracker); relates to
[#614](https://github.com/krakenhavoc/cmd_and_ctrl/issues/614) (`/cc-end`, whose
open "host or admin" design point this settles) and
[#607](https://github.com/krakenhavoc/cmd_and_ctrl/issues/607) (S34 user rows)
**Numbering:** on 2026-09-19, after `git fetch --all --prune`, `docs/decisions/`
was listed on every remote branch (321 refs) and every local branch. The highest
number present anywhere was `0074-triggered-mana-abilities.md`, so this ADR takes
**0075**. 0005, 0024, 0029 and 0030 stay permanently unused per AGENTS.md §4.
**Builds on:** [ADR 0003](0003-auth-and-lobby.md) (roles),
[ADR 0023](0023-develop-environment.md) (the dev-only card spawner and the rule
that dev features are gated server-side),
[ADR 0033](0033-ai-bot-seat.md) (bot pacing),
[ADR 0051](0051-user-database.md) (`created_by`, which replaces this ADR's host
field once S34 lands)
**Amends:** [ADR 0023](0023-develop-environment.md). Spawning stops being
dev-only; a production table can allow it through a visible table setting.

---

## 1. Context

A table has no owner and no settings surface. What exists today:

- **Undo budget.** `Game.UndoLimit` (default `DefaultUndoLimit = 1` per player per
  turn, refreshed on the active seat's untap). `set_undo_limit` sets it, but
  **any seated player** may send it (`mutations.go` `SetUndoLimit`: "Sandbox: any
  seated player or admin may call it") and the client has no control for it.
  There is also a trap: `Game.Start` rewrites `UndoLimit <= 0` to the default
  (`game.go`, the `if g.UndoLimit <= 0` block), so "no undos" set before start
  quietly becomes one. A snapshot restore of a limit of 0 has the same problem.
- **Undo scope** is fixed: a seated player can undo only their own entries, and
  the admin (`Caller == uuid.Nil`) can undo anything without spending budget
  (`ws/room.go`).
- **Starting life** (`game.StartingLife = 40`) and **commander damage**
  (`game.CommanderDamageLethal = 21`) are package constants.
- **Bot pacing** is a compile-time default (`aiseat` `defaultMinThink` 700ms,
  `MaxThink` 2s / 5s for `strong`).
- **Spawning** is `POST /games/{id}/dev/spawn`, wrapped in `requireDevFeature`,
  so it returns 404 in production. On a dev box anyone at the table may call it.
  It goes through `Lobby.SpawnCards` → `applyLocked` → `room.ApplyExternal`, so it
  is broadcast. It is **not** narrated to the table.
- **"Admin"** means only `auth.RoleAdmin`, the holder of the shared server token.
  `POST /games` is `RoleAdmin`-only, so every game is created by the operator
  (directly or through the Discord bot's `/cc-invite`). Nothing records *which
  player* the table belongs to.

The owner wants a game's admin to control the table's settings, starting with
undos, and to be able to spawn tokens and cards on a live table. That covers
cards the engine cannot create yet and repairs after a misplay. The owner
decided the following on 2026-09-19:

| Question | Decision |
|---|---|
| Who controls a table | **The host plus the server admin.** |
| Spawning in production | **A table setting, off by default. When it is on, the host and admin may spawn, and every spawn is announced in the game log.** |
| v1 settings | **Undo rules, starting life, commander damage threshold, bot pacing** (plus the spawn switch itself) |
| When settings can change | **Any time, in the lobby or mid-game. Every change is logged.** |

## 2. Decision

### 2.1 The host

`GameMeta` gains `HostPlayerID uuid.UUID` (`json:"host_player_id,omitempty"`).

- **Set at creation** when the creator names one: `POST /games` accepts an
  optional `host_discord_id`. `/cc-invite` passes the invoking Discord user, so
  whoever runs `/cc-invite` hosts that table. When that Discord identity claims
  a seat, `HostPlayerID` is bound to that seat's player ID.
- **Otherwise it is the first human seat to join.** A bot seat is never host.
- **Transferable**: `POST /games/{id}/host {player_id}` is available to the host or
  the admin. It is also the escape hatch if a host leaves (ADR 0060). When the
  host seat is eliminated or leaves, hosting passes automatically to the
  next human seat in turn order, so there is always someone who can change the
  settings.
- **"Host or admin"** is one predicate, `lobby.CanManageTable(principal, meta)`:
  `RoleAdmin`, or `RolePlayer` whose `PlayerID == meta.HostPlayerID` and whose
  `GameID` is this game. Every route and action in this ADR uses it. #614's bot
  command uses the same predicate by comparing the invoking Discord user with the
  host seat's `discord_id`.
- **S34 migration.** When ADR 0051's `games.created_by` exists, the creator's
  user row seeds `HostPlayerID` at claim time. The field stays because hosting
  is a seat role that can be transferred, while `created_by` is historical.

The host is visible to everyone: `SeatInfo` and the game view carry `is_host`,
and the client shows a crown on that seat.

### 2.2 Settings: one struct, owned by the engine

```go
// game/settings.go
type TableSettings struct {
    UndoLimit        int       // per player per turn; UndoUnlimited = -1
    UndoScope        UndoScope // "own" (today) | "host_any": the host may undo anyone's entry
    StartingLife     int       // 1..999, default 40
    CommanderDamage  int       // 1..99, default 21
    BotPace          BotPace   // "fast" | "normal" | "slow"
    AllowSpawn       bool      // default false
}
```

- It lives on `Game.Settings` and replaces the `UndoLimit` field. It is cloned by
  `Clone`, carried by the snapshot, and restored as written. **No `<= 0` →
  default rewrite.** A zero value in an old snapshot is migrated once, by the
  snapshot schema version rather than by the value, so a deliberate 0 survives
  a restore.
- Engine readers change from constants to settings:
  `StartingLife` → `g.Settings.StartingLife` at `Start` (seat creation), and
  `CommanderDamageLethal` → `g.Settings.CommanderDamage` in the SBA check. The
  constants remain as the defaults.
- `UndoUnlimited` means the budget is never debited, and the client shows "∞".
- Bot pace is a named preset rather than raw milliseconds, because nobody wants
  to tune `MinThink` at the table. The `aiseat` runner reads
  `g.Settings.BotPace` on each decision (`fast` = 0/2s, `normal` = today's
  defaults, `slow` = 2s/8s). `strong` keeps its longer hard deadline in every
  preset.

The game view exposes `settings` to every viewer, and `is_host` to the host.
Settings are public on purpose: spawning and undo rules are things the other
players should be able to see.

### 2.3 Changing settings

- **One mutation**, `Game.UpdateSettings(actor uuid.UUID, patch SettingsPatch)`,
  validates ranges, applies only the fields that are present, and emits
  `EventSettingsChanged` with the old and new values. The game log narrates it:
  *"Luke (host) set undos to 3 per turn."*, *"Luke (host) allowed spawning."*
- **Lobby phase:** `PATCH /games/{id}/settings` (host or admin).
- **Mid-game:** WebSocket action `set_table_settings` (same payload, same
  predicate). `set_undo_limit` remains as a deprecated alias that calls the same
  mutation, and it is now **host-or-admin only**. That is a behaviour change,
  and the purpose of this ADR. `docs/protocol.md` changes with it.
- **Settings changes are not undoable.** The change is not pushed onto the undo
  stack, and an undo does not roll a setting back. Otherwise, lowering the undo
  limit could be undone with the undo it was meant to stop. `Game.Settings` is
  therefore copied forward across a `RestoreFrom` rather than restored from the
  pre-state.
- **Timing semantics per field:**
  - *Undo limit:* takes effect immediately and refreshes every seat's
    `UndosRemaining` to the new value, as today.
  - *Starting life:* changes after `Start` are **rejected**. The setting has
    already been applied, and rewriting life totals mid-game would be a
    different, sneakier feature. The client disables the control once the
    game is active. Use the existing life controls to adjust totals.
  - *Commander damage:* takes effect at the next state-based action check. If
    lowering it puts a player over the new threshold, that player loses at that
    check. The confirm dialog says so when it applies.
  - *Bot pace, spawn switch:* immediate.

### 2.4 Spawning on a live table

- **New route** `POST /games/{id}/spawn`. It requires `CanManageTable` **and**
  `Settings.AllowSpawn`. Otherwise it returns 403 with a message saying which
  one failed. It is not wrapped in `requireDev`. The dev route
  `POST /games/{id}/dev/spawn` keeps its current dev-only semantics (anyone at
  the table, no setting required), so the preview environment's tooling does
  not change.
- **What can be spawned:** any indexed Scryfall card (the existing
  `resolveDevCard` path) **and any token template** from `tokens_table.go` and
  the behaviour tokens in `tokens.go` (Treasure, Food, Clue, …). Tokens go
  through the same `CreateToken` primitive the catalog uses, so a spawned
  Treasure works like a real one. Tokens are why the owner wants this in
  production. The dev spawner cannot make them today, because a token has no
  Scryfall printing.
- **Zones** as the dev spawner allows, except that tokens may only go to the
  battlefield: a token in any other zone ceases to exist at the next
  state-based action check (`game/token_existence.go`), so spawning one there does nothing.
- **The mutation is renamed** `SpawnCardsForDev` → `SpawnCards` and gains an
  `actor` argument. The "nothing in this file authorizes anything" contract
  stays: authorization happens at the HTTP edge, which is now either
  `requireDevFeature` or `CanManageTable` plus `AllowSpawn`.
- **Announced:** each spawn emits `EventSpawned {actor, controller, zone, name,
  count}`. The log says *"Luke (host) spawned 2 × Treasure onto Ana's
  battlefield."* A spawn into a hidden zone names the zone but not the card
  (*"…spawned a card into Ana's hand"*), and the visibility rules already in
  `SpawnCardsForDev` still decide who sees the card itself.
- **Undoable.** A production spawn pushes an undo entry attributed to the
  spawner, so a mistaken spawn can be undone with the ordinary undo. It is a
  `FreeUndo` entry, so fixing a typo does not cost the host their per-turn
  undo.
- **Triggers fire.** A battlefield spawn emits `EventETB`, as it does today. That
  is the point of the feature. It also means a spawn can put abilities on the
  stack.

### 2.5 Client

- **Lobby:** a "Table settings" panel on the game's lobby page. The host and
  admin can edit it; everyone else sees it read-only.
- **In game:** a settings entry in the game menu that opens the same panel
  (starting life is disabled). The host gets a "Spawn" entry, available only
  while `AllowSpawn` is on. It reuses `CardSpawner.svelte` with a Tokens tab
  added, and moves it out of `components/dev/`.
- **Everyone** sees a small "spawning on" badge in the table header while the
  switch is on, and the undo counter shows the configured limit.

## 3. Consequences

- A table has one visible owner and one settings panel, and #614 can be
  written.
- `set_undo_limit` is no longer available to every seat. No client ever sent it,
  so the only callers are tests and `gamecli` scripts, which get updated.
- Spawning in production is a deliberate exception to ADR 0023's "spawning is a
  cheat on a live table". The table opts in, the setting is visible to
  everyone, and every use is in the log. The dev-feature gate still protects
  the *dev* route, and the `requireDev` rule in AGENTS.md is unchanged for dev
  tooling.
- `Game.Settings` is new snapshot state, so it needs a snapshot schema bump and
  a migration for older snapshots (S33 rules).
- Bots never host and cannot change settings. The legal-move enumerator does
  not offer `set_table_settings` or spawn.

## 4. Alternatives considered

- **Server admin only.** This needs no new host concept, but players could never
  run their own table, and the Discord flow already knows who asked for the
  game.
- **Any seated player (today).** Anyone can change the undo limit in their own
  favour mid-game without asking the table.
- **Spawning always allowed for the host, not logged.** The most flexible
  option, but the other players cannot tell a spawned Treasure from a real one.
- **Settings in `GameMeta` (lobby) rather than on `Game`.** The engine reads
  three of the five fields at rules time, and snapshots and undo would have
  to agree across two stores.

## 5. Implementation plan (sub-PRs)

1. **Engine settings.** Add `game.TableSettings`, `UpdateSettings`,
   `EventSettingsChanged`, readers for starting life and commander damage,
   `UndoUnlimited`, and the snapshot schema and migration. Remove the `<= 0`
   rewrite. Tests: a deliberate 0 survives Start and restore, a settings change
   survives an undo, and a lowered commander damage threshold kills at the next
   SBA check.
2. **Host.** Add `HostPlayerID`, `host_discord_id` on create, first-human-seat
   fallback, transfer and auto-pass, `CanManageTable`, and `is_host` on the
   wire. `/cc-invite` passes the invoker.
3. **Settings surfaces.** Add `PATCH /games/{id}/settings`, the
   `set_table_settings` action, and the gated `set_undo_limit` alias. Update
   `docs/protocol.md` and `docs/lobby.md`, and add the log narration.
4. **Production spawn.** Add `POST /games/{id}/spawn`, token templates,
   `EventSpawned`, the log line, and the free undo entry. Rename the mutation.
5. **Client.** Add the settings panel (lobby and in game), the host crown, the
   spawner move and Tokens tab, the spawning badge, and ∞ undo.
6. **Bot pace.** The `aiseat` runner reads the preset. Update `docs/bot.md`.

Sub-PRs 1 and 2 are independent. 3 depends on both. 4 depends on 2 and on 1's
`AllowSpawn`. 5 and 6 come last.

## 6. Open questions

- Should `UndoScope: host_any` exist in v1, or is "admin can undo anything"
  (today) enough? It is included because "undo rules" was chosen, and it is
  easy to drop.
- Settings **templates** (a "house rules" preset reused across tables) wait
  for S34, which is where a per-user store appears.
