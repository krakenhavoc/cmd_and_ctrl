# ADR 0075 — Table settings and host controls: who runs a table, what they can change, and spawning on a live game

**Status:** Accepted · 2026-09-19 · S35 — Playtest stabilisation, round 2
**Issues:** [#1032](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1032) (tracker); relates to
[#614](https://github.com/krakenhavoc/cmd_and_ctrl/issues/614) (`/c2-end`, whose
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
**Amended by:** the [2026-10-13 amendment](#amendment--2026-10-13-moderation-for-a-public-table-2984)
at the end of this file (#2984): the name and chat filter, remove-and-block,
and report a player.

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
  (directly or through the Discord bot's `/c2-invite`). Nothing records *which
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
  optional `host_discord_id`. `/c2-invite` passes the invoking Discord user, so
  whoever runs `/c2-invite` hosts that table. When that Discord identity claims
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
   wire. `/c2-invite` passes the invoker.
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

---

## Amendment — 2026-10-13: moderation for a public table (#2984)

**Issue:** [#2984](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2984).
**Numbering:** no new ADR. Seats, the host and who may manage a table are this
ADR's; removing a seat is [ADR 0060](0060-leaving-the-game.md)'s
2026-10-12 amendment (#2785, #3093). A block is one more thing the host can do
to a seat, so it belongs here as §2.6. ADR 0149 (the privacy page) is amended
only in the sense that it now lists these tables.
**Builds on:** §2.1 (the host), [ADR 0051](0051-user-database.md) (users and
seats), [ADR 0149](0149-privacy-notice-and-account-deletion.md) (what is stored
and how it is deleted), [ADR 0124](0124-admin-views-accounts-games-and-who-is-on-now.md)
(the admin views).

### 2.6 Moderation

A public post means strangers share tables. Three small tools, none of which
changes the lobby state machine (a removal and a join are the transitions they
already were; a block is a refusal on the join).

1. **A name and chat filter.** `internal/moderation`: a hand-kept denylist
   (`denylist.txt`, which owns the rules for adding to it) and a normaliser that
   folds case, diacritics, look-alike letters, leetspeak, zero-width characters,
   stretched letters, and letters spaced out one or two at a time. Entries are
   either **whole-word** (`=word`, most of them: `ass` as a substring would flag
   "classic") or **substring** (only long distinctive words, at least five
   letters). It is a slur filter, not a profanity filter: swearing and trash
   talk are table banter.
   - **Names** are refused, not rewritten: a guest name, the Discord display
     name shown on a seat (global name first, the username as the fallback,
     the join refused if both are listed) and an explicit table name. The error
     is a 400 with the stable `code` `name_not_allowed`, and says to pick
     another name without saying which word.
   - **Chat** is **masked**, not rejected. A false positive costs one word
     instead of a message, and a rejection would teach the sender which spelling
     to try next. The sender sees the masked line too, so what the table read is
     what a report holds. Bot lines are not filtered.
   - A **test holds the list clear of Magic**: every card name in the Scryfall
     dump (skipped without `CMDCTRL_SCRYFALL_DUMP`, like the other dump tests)
     and every catalog name (always), each tested bare and inside a sentence,
     and each face of a split card.
2. **Remove and block.** `DELETE /games/{id}/seats/{player_id}?block=true`.
   - **Who** is unchanged from #2785 and #3093, by the owner's decision of
     2026-10-12: on an unstarted table the host, the creator or an admin; on a
     started game an admin alone. `?block=true` adds no authority.
   - **What a block keys on** decides what it can hold against, so the limits are
     in the open. A signed-in seat is blocked by its account and its Discord id;
     a guest seat by its folded name (`moderation.Key`). A guest has no account,
     no stable id, and no address the server keeps, so a different name gets
     past a name block; so does a blocked signed-in person who signs out and
     joins as a guest under a new name. A name block is applied to guest joins
     only, so two people called "Alex" on Discord are not one person. We did not
     key on an IP address: it would mean storing one, it punishes everyone behind
     a shared address, and it is trivially changed by the person it is aimed at.
   - **The wall is the invite.** Rotating it was always the way to keep a person
     out; a block answers the narrower, common case without changing the link
     for everybody. The Remove control says so beside the checkbox. This also
     answers #3093's worry that a removed person who still holds the link can sit
     down again.
   - **Scope and storage:** per table, in `table_blocks` (migration 0016,
     deleted with the game), read back at restore, so a restart does not lift
     it. There is **no unblock**: a block is one host's act at one table, and a
     mistake costs a new invite or a new table, never a ban from the site. The
     refusal is a 403 with the stable `code` `blocked_from_table`, given only
     after the invite is checked. A spectator invite is a separate credential and
     is not blocked; rotate it too.
3. **Report a player.** `POST /games/{id}/reports`, from the ⋯ menu.
   - A seated player reports another **human** seat (not themselves, not a bot,
     not a spectator, who has no seat to name). Nothing is sent to anyone and
     nothing is done to the reported player: it is a tip to the operator, listed
     on the admin **Reports** tab (`GET /admin/reports`, admin only) beside
     Live now, Games and Accounts, with Dismiss.
   - **Rate limit** per reporter: a burst of three, then one every five minutes,
     counted before the request is validated, so a script cannot probe for free.
   - **A recent chat excerpt.** Chat is otherwise not stored at all (§ on chat in
     `docs/protocol.md`). Each room keeps the last 20 human lines in memory only,
     lost at a restart, and a report copies the last 10 at the moment it is
     filed. Nothing else about chat changes.
   - **Storage and retention:** `player_reports` (migration 0016). Removed 90
     days after it is filed, past 500 rows (oldest first), or on Dismiss; swept
     every six hours and on each filing and read. It outlives its table.
     Deleting an account removes every report the person filed or is the subject
     of; a report that merely quotes their chat line keeps the line until it
     expires, which the privacy page says.

**Privacy.** The page now states chat's in-memory window, what a report holds,
and what a block keeps (`accountdelete.Inventory` lists both tables;
`PRIVACY_SCHEMA_VERSION` is 16).

**Alternatives considered.**

- *Reject a chat line that matches.* Rejected: it blocks the message over a
  possible false positive, and it tells the sender what to dodge.
- *Block by IP address or a browser token.* Rejected for guests above; a browser
  token would be a new thing to store about a person for a control a determined
  person defeats by opening a private window.
- *A site-wide ban list.* Out of scope: the server has no account-level
  moderation state, and ADR 0149 just spent a release deleting accounts, not
  adding to what they hold.
- *Let the host kick from a started game.* The owner decided 2026-10-12 that
  removal from a started game is admin-only (#3093), and this does not widen it.

**Consequences.** Three new surfaces and one migration. The filter is
maintained by hand and goes stale slowly; its test is what lets someone add a
word without reading thirty thousand card names. A guest who is blocked and
renames gets back in, and the docs and the control say so. Spectators cannot be
reported or blocked.

**Out of scope, left for follow-ups:** unblocking; blocking a spectator; a
host-visible list of who is blocked; reporting a spectator or a chat line
directly from the chat panel; notifying the operator (the list is polled by
looking at it); per-table or per-account mutes.
