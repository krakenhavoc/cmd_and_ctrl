# ADR 0041 — Persisting games across a deploy

**Status:** Accepted · 2026-09-11 · Phases 1–2 of 3
**Superseded in part by:** [ADR 0044](0044-surviving-a-deploy.md) — the
claim below that *"the client already auto-reconnects and resyncs by
`seq`, so nothing on the client needs to change"* is false on both
halves. The client treats the shutdown close code as terminal and
never redials, and there is no seq-based resync; the session dies with
the process, and the "re-authenticate through your invite link"
recovery does not exist for a started game. Every decision in this ADR
stands; 0044 builds the return path they assumed was already there.
**Amended by:** [ADR 0054](0054-dice-rolls-and-coin-flips.md) Decision 3
(#744) — Decision 3 below no longer describes the code. The RNG is a
32-byte secret key plus per-turn stream counters, not a PCG position;
the snapshot writes `rngKind: "keyed"`, a caller-supplied `*rand.Rand`
seeds the key and is persistable like any other game, and
`"pcg"` / `"external"` files restore with a fresh key.
**Amended by:** [ADR 0051](0051-user-database.md) decision 1 (S34
sub-PR 1, [#607](https://github.com/krakenhavoc/cmd_and_ctrl/issues/607))
— `<dataDir>/db/` is a new artifact directory alongside the ones below,
holding the persistent SQLite store and its backup copy; see the
Artifact layout table. ADR 0051 also supersedes `<dumpDir>/lobby/<id>.json`
itself once its sub-PR 3 lands (`games` / `seats` / `invites` become
rows), which is not yet true of the code this ADR describes.

## Context

Every deploy destroys every live game. `RoomManager`'s own doc comment
has said so since S04: *"if the server restarts, in-memory games are
lost and clients get a 'game not found' on reconnect."* Since the
project ships several times a day, the practical effect is that nobody
can play while anyone is working.

The ask, in the owner's words: *"get game persistence so we can
publish most changes without interrupting other than refreshes most of
the time."* **A refresh is acceptable. Connection continuity is not the
requirement — surviving a process restart is.** That matters, because
the client already auto-reconnects and resyncs by `seq`, so nothing on
the client needs to change.

Three things already existed and were nearly, but not quite, the
answer:

- **`clone.go`** is 414 lines of hand-written deep copy over
  `game.Game`, and `RestoreFrom` puts a clone back. It is a serialiser
  and deserialiser in everything but format, and the undo stack
  exercises both on every action.
- **A write-only crash-recovery path.** Every `Apply` atomically writes
  `<dumpDir>/games/<id>.json` plus an append-only replay JSONL.
  **Nothing has ever read either back.** The only `Load` at boot is the
  card index.
- But that dump is a **`protocol.GameView`** — a redacted projection
  built for a browser. It cannot reconstruct a `Game`, and was never
  meant to.

## Decision 1 — State persistence, not action-log replay

Action-log replay looks ideal here: actions arrive as JSON through one
chokepoint (`room.Apply`), and the RNG could be seeded.

**It breaks on precisely the case that motivates the work.** Replay
re-derives state by *running the code*. The entire point of persisting
is to deploy **changed** code; the moment card logic changes, the
replay diverges from the game the players were actually in. Replay is
sound for crash recovery on an identical binary and useless across a
version boundary.

So: an explicit `game.GameSnapshot` (schema-versioned), written whole
and read back whole.

## Decision 2 — An explicit snapshot type, and a drift test

`encoding/json` cannot see `Game`'s eleven unexported fields and
refuses its func-typed ones, so tagging the domain struct was never an
option. `snapshot.go` mirrors the domain types instead, using
`clone.go` as its checklist.

Two hand-written deep copies of the same types is a rot risk: adding a
field to `Game` or `Card` compiles fine and silently fails to survive
either an undo or a deploy, and stays invisible until a player loses a
counter.

`snapshot_drift_test.go` closes that. Every field on `Game`, `Card`,
`Player`, `Zone`, `StackItem`, `DelayedTrigger` and `PendingChoice`
must be classified as **carried**, **rebuilt** or **dropped (with a
reason)**. An unclassified field fails CI with a message naming the
field and the three options. A `dropped` field whose reason does not
name a `ContinuationCensus` counter — or declare itself outside game
state — also fails, so the `dropped` bucket cannot become a place to
quietly lose things.

The round-trip test is a property, not a checklist:

```
capture(restore(decode(encode(capture(g))))) == capture(g)
```

New snapshot fields are covered the moment they are added.

### Amendment, 2026-09-19: `carried` is a property, not a row in a table

*Amendment, 2026-09-19, branch
`chore/993-1005-exit-lint-and-carried-enforcement`. Closes
[#1005](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1005), filed
while landing #980.*

The two sentences above — "an unclassified field fails CI" and "new
snapshot fields are covered the moment they are added" — were both true
and together they read as a guarantee neither of them makes.

**What was actually covered.** The drift test walks field NAMES. It
never reads a value, so a row saying `carried` was a promise: `dropped`
was the only disposition this ADR held to anything, through the census
check. And the round-trip property compares
`capture → JSON → restore → capture`, which is structurally blind to a
field missing from BOTH projections — absent equals absent, on both
sides. Delete a field from `snapshotCard` **and** `restoreCard` and
every snapshot test passed. That was live: `NamedTribe`, `ChosenColor`
and `ChosenPlayer`, the CR 614.12-family stored answers, were provably
uncovered until #980 added them to the round trip's fixture by hand —
and adding them by hand is the part that does not scale, because the
fixture is the other half of the blind spot.

**Decision: for every field classified `carried`, write a distinctive
value onto a fixture, run the real capture → JSON → restore, and read
the value back off the RESTORED GAME.** `snapshot_carried_test.go`, 213
fields across the seven probed types. It is reflection over the drift
plan rather than a generated table, so a `carried` row added tomorrow is
enforced the moment it lands with no test edit — the same property the
round trip claims, made true for the direction it cannot see.

Four things make it hold rather than merely run:

- **the value generator is plan-aware.** Filling a `rebuilt` field —
  `Card.ManaAbilities` is two closures the catalog re-supplies by oracle
  ID — would make the test assert the opposite of what that row says.
- **unexported fields are reached through `unsafe`.** Half of `Game`'s
  carried rows are unexported (the RNG key, the event sequence, the
  combat lock-ins) and they are exactly the fields no other test can
  see. A test may do this; production code may not.
- **the failure names the projection.** The value is looked for in the
  CAPTURED json: in the file means `restoreX` dropped it, not in the
  file means `snapshotX` did.
- **both escape hatches are checked.** `carriedFixture` (a type the
  generator cannot invent a plausible value for) and
  `carriedNotRoundTrippable` (the two rows where capture → restore is
  deliberately not an identity: `Game.layerVersion`, which restore
  advances by one to force a layer recompute, and
  `ScopedStatic.Duration`, which `Clone` carries and the snapshot does
  not). An entry that no longer names a live field fails CI, so neither
  list can outlive the code it describes.

`driftPlans` is now one list walked by all three tests, and
`TestEveryPlannedTypeHasACarriedProbe` fails when a domain TYPE is added
to the plan with no probe — so this survives somebody adding a type, not
just a field. The header comments in `snapshot_test.go` and
`snapshot_drift_test.go` say what each test does and does not cover.

## Decision 3 — The RNG is owned, seeded from crypto/rand, and carried

`Game.rng` is a `*math/rand/v2.Rand`, and **production always passed
`nil`**, so every shuffle drew from the process-global source. That was
unpredictable, but it was also *unrecordable*: `rand.Rand` exposes no
accessor for its `Source`, so there was no stream position to write
down. A restart could only start a new stream.

`Start(nil)` now mints a `*rand.PCG` seeded from `crypto/rand`, keeps
it in a new `rngState` field, and wraps it. `PCG` implements
`encoding.BinaryMarshaler`, so the game's exact position in its random
stream round-trips through a snapshot: a restored game continues the
library order the players were actually headed for. The seed never
leaves the server, so it stays unguessable.

A caller-supplied `*rand.Rand` (what tests pass) still cannot be read
back. That is recorded honestly as `rngKind: "external"` and counted in
the census, rather than silently reseeded — a silent reseed is how you
get a restore that quietly re-deals a library. `StartWithSource` is the
shape that gets determinism *and* persistence.

## Decision 4 — Continuations are censused, and the restore point holds at the last clean boundary

Parts of a live game are Go function pointers, not data:

| Where | What |
|---|---|
| `StackItem.Effect` | what an ability does on resolution |
| `StackItem.targetSpec` | the clause its targets were legal under |
| `DelayedTrigger.Effect` | "at the beginning of the next end step…" |
| `PendingChoice`'s six resume frames | a paused game holds the rest of the effect as a continuation |
| `ScopedStatic.Ability` | Giant Growth's +3/+3 |
| `TurnScopedReplacements` | Fog |
| `Card.ManaAbilities` / `ActivatedAbilities` on a true token | Treasure, Food, Clue, Blood *are* their ability |

Some are re-derivable, because the closure was looked up from the card
catalog by oracle ID in the first place and the catalog is **code the
new binary already has**. Restore re-runs those lookups: a spell's
target spec, and a token *copy*'s abilities (CR 707.2 copies the oracle
identity, so a copy has a printing to look up — a true token does not).
Listeners and built-in replacements are likewise rebuilt by `NewGame`,
which means a new binary's listener set wins, which is what you want
across a deploy.

The rest are genuinely lost in this phase.

**The rule:** `CaptureSnapshot` takes a `ContinuationCensus` of every
closure it could not represent. A snapshot with a non-empty census is
**not a restore point**, and the writer skips it, leaving the previous
file in place. So what sits on disk is always the most recent state
that can be rebuilt *exactly*, and a restart rewinds the table to that
point rather than resurrecting it subtly wrong.

This is the sanctioned "roll back to the last clean boundary"
simplification, generalised from `PendingChoice` to every continuation
and enforced by a predicate instead of a comment. Silently dropping a
Fog is worse than a rewind the players can see.

Two honest consequences:

- The rewind distance is **not** always short. A live token on the
  battlefield — a Treasure — makes every subsequent state uncleanable
  until it is spent. That is the strongest argument for prioritising
  phase 3, and the census labels make it diagnosable rather than
  mysterious.
- As continuations become data-driven, each census counter drops to a
  permanent zero and the rewind distance shrinks to nothing. No
  redesign is needed to get there.

## Decision 5 — Version skew: **abandon**, and keep the file

When the state shape changes, restore must migrate or refuse. The
schema version is checked first and never guessed at: a file from a
*newer* server (a rollback, or a mixed-version fleet) and a file older
than `minRestorableSchema` are both refused.

The policy for a game that cannot be restored is **abandon** — logged
loudly at `ERROR` with the game ID, the game does not come back, and
**the file is deliberately left on disk** so rolling the binary back
returns the game.

The alternatives were considered and rejected:

- **Block the deploy.** Makes shipping a card fix hostage to whatever
  one table happens to be holding. Converts "one game rewinds" into
  "nobody ships" — the exact failure this work exists to end.
- **Freeze read-only.** Needs a whole second lifecycle state and a UI
  for it, or it is a table that looks live and silently refuses every
  action.

Abandonment is the honest one: the players see "game not found", which
is *exactly what they see today on every restart*. The worst case of
this feature is the current behaviour.

A bad restore point is never fatal to a boot — one corrupt file costs
one game, not the server.

## What also had to be persisted: the invite

`GameMeta` — display name, seat list, and **both invite tokens** —
lived only in `Lobby.games` and was never written anywhere.

This matters more than it looks. `auth.MemoryAuthenticator` is explicit
that *"session is lost on server restart — users re-auth on
reconnect."* So after a deploy every player re-authenticates through
their invite link. Without the token on disk, a restored game is a
table nobody can open — indistinguishable, from the player's seat, from
having lost the game.

`<dumpDir>/lobby/<id>.json` holds it, written `0600` because it
contains secrets. A restored engine snapshot with no metadata behind it
is dropped and unregistered rather than left half-visible.

**Session persistence is out of scope and is the remaining gap between
"refresh" and "re-open the invite link."** Swapping
`MemoryAuthenticator` for an HMAC implementation is a one-line change
in `main.go` — the `Authenticator` interface is the only seam the rest
of the server sees — and is the obvious next increment.

## Artifact layout

| Path | Contents | Read at boot |
|---|---|---|
| `<dumpDir>/games/<id>.json` | `protocol.SnapshotPayload` — forensics | no (unchanged) |
| `<dumpDir>/replays/<id>.jsonl` | append-only view history | no (unchanged) |
| `<dumpDir>/restore/<id>.json` | `game.GameSnapshot` + room `seq` | **yes** |
| `<dumpDir>/lobby/<id>.json` | `lobby.GameMeta`, mode 0600 | **yes** |
| `<dumpDir>/db/cmdctrl.sqlite` | The persistent database (ADR 0051, `internal/db`), mode 0600 | **yes** (opened + migrated before `RestoreFromDisk`) |
| `<dumpDir>/db/cmdctrl.backup.sqlite` | `VACUUM INTO` copy of the above, on a timer (`CMDCTRL_DB_BACKUP_INTERVAL`), mode 0600 | no |

The room's `seq` travels with the snapshot: a restored room that
restarted its counter would hand reconnecting clients a sequence number
they had already seen.

Both new artifacts are removed when a game ends and when it is deleted,
so no boot rebuilds a dead table.

## Phase 3 (not in this change)

Make continuations data-driven so the census empties out. The codebase
is already drifting that way — the S19/S20/S22 resume frames
(`triggerResumeFrame`, `pickTargetFrame`, `searchResumeFrame`) are
named structs carrying mostly data with a single `Build`/`Then` closure
left at the bottom. Naming those closures and looking them up from a
registry at restore is the same move `EffectResolver` already makes for
spells.

Priority order, by how often each blocks a restore point: intrinsic
token abilities (a Treasure blocks every subsequent capture), then
`TurnScopedStatics`/`Replacements`, then ability `StackItem.Effect`,
then the resume frames.
