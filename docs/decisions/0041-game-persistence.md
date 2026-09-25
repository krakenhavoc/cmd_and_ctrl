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
**Amended by:** #521 (S33 sub-PR 5) — the first item of Phase 3 has
landed. Token templates carry a synthetic catalog key
(`token:treasure`, `token:food`, …) and their abilities are registered
in the catalog, so the "intrinsic token abilities" row of Decision 4's
table is empty and a Treasure on the battlefield no longer blocks every
restore point. The snapshot schema is **v5**; see the addendum at the
end.
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
Artifact layout table.
**Amended by:** [ADR 0051](0051-user-database.md) decision 4 (S34
sub-PR 3, [#607](https://github.com/krakenhavoc/cmd_and_ctrl/issues/607))
— `<dumpDir>/lobby/<id>.json` is retired. Its content is now the
`games`, `seats` and `invites` rows in the database. Invite tokens are
stored only as SHA-256 hashes. The first boot of the new binary imports
any `lobby/*.json` and renames each file to `<id>.json.imported` rather
than deleting it, so a rolled-back binary can have them back
([docs/environments.md](../environments.md)). Two behaviours below
change with it. A restored room is paired with its `games` row instead
of a file. Rows for games that did not come back are **kept**, not
pruned: a finished game has no restore point by design, and its row is
the history "my games" reads. The engine artifacts (`restore/`,
`replays/`, `games/`) are unchanged.
**Amended by:** the phase 3 amendment of 2026-09-24
([#1497](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1497)) at
the end of this file. It replaces the priority order in "Phase 3 (not
in this change)" and sets the design for the rest of phase 3: scoped
continuous effects and delayed triggers become data records, and every
closure left in game state must be rebuilt, keyed, or transient.

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
| ~~`Card.ManaAbilities` / `ActivatedAbilities` on a true token~~ | **Closed by #521** — a token template is a catalog entry under its own key, so its abilities are re-derived like a printed card's. What remains in this row is an ability stamped onto ONE instance at runtime with no entry behind it, which is rare and still censused |

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
  mysterious. **(#521: no longer true of tokens, which was the worst
  case of it. The point stands for the continuations still listed.)**
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
| `<dumpDir>/lobby/<id>.json` | `lobby.GameMeta`, mode 0600. **Retired by ADR 0051**: imported into the database at boot and renamed `<id>.json.imported` | only by the one-time importer |
| `<dumpDir>/db/cmdctrl.sqlite` | The persistent database (ADR 0051, `internal/db`), mode 0600 | **yes** (opened + migrated before `RestoreFromDisk`) |
| `<dumpDir>/db/cmdctrl.backup.sqlite` | `VACUUM INTO` copy of the above, on a timer (`CMDCTRL_DB_BACKUP_INTERVAL`), mode 0600 | no |

The room's `seq` travels with the snapshot: a restored room that
restarted its counter would hand reconnecting clients a sequence number
they had already seen.

**Superseded in part by [ADR 0044](0044-surviving-a-deploy.md) decision
5 (#523).** That covers only restarting the counter *at zero*; it does
not cover restarting it at some other value LOWER than what a connected
client already held — which is exactly what happens whenever the
restore point named above is not the game's current state (the
`ContinuationCensus` case just below). 0044 adds a restore generation
alongside `seq` so the client can tell that rewind apart from a dropped
frame, and corrects `docs/protocol.md`'s "monotonically non-decreasing"
claim to "non-decreasing within a generation" to match.

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

Priority order, by how often each blocks a restore point: ~~intrinsic
token abilities (a Treasure blocks every subsequent capture)~~ **(done,
#521)**, then `TurnScopedStatics`/`Replacements`, then ability
`StackItem.Effect`, then the resume frames.

## Addendum — token catalog keys and schema v5 (#521)

A token has no printing and therefore no oracle ID, and every catalog
hook keys on one. That is the whole reason a Treasure's four fields of
plain data had to ride on the instance as a closure-bearing
`ManaAbilityShape`, and the census counted it as an unserialisable
continuation — so the predicate was really asking *"is there an oracle
ID to look up?"* rather than *"is anything unserialisable actually
here?"*.

The fix is the shape emblems (ADR 0064) and granted ability bundles
(CR 707.9a, #665) already use: a synthetic key in a namespace of its
own — `game.TokenKey("treasure")` → `token:treasure` — filed in the
same def map cards are filed in, resolved by the same `CatalogLookup`.
A Scryfall oracle ID is a UUID and contains no colon, so the namespaces
cannot collide. `Card.TokenKey` carries it; `CatalogKey` prefers a real
oracle ID and reads it only when there is none, which is what keeps CR
707.2's token *copy* resolving to the card it copied. The census now
asks the registry whether it can return the abilities an object is
holding, so an ability stamped onto one instance at runtime is still
counted — the counter became accurate, not unreachable.

It fixed a second bug one zone over, for free: `stampActivatedAbilities`
skipped any card with no oracle ID, so **every** token's activated
abilities were dropped on the way to the client. Food, Clue, Blood and
the Lander reached the table inert.

**Schema v5.** The new `tokenKey` on a card is additive and zero-values
correctly for an old file. The bump is for the other direction, exactly
as v2's was for emblems: before this change a board holding a live
token was never written as a restore point at all, so no file a v4
binary could be handed contained one. Afterwards they are the normal
output, and a v4 binary reading one would drop the key and restore
those tokens as blank artifacts — a Treasure that no longer taps for
mana, silently. There is no per-field way to say "refuse this file if
you do not know what a token key is", so the version is it, and the
cost is the one v2 accepted: a pre-#521 binary refuses every post-#521
restore point rather than restoring one wrong.

## Amendment, 2026-09-24 — `rebuilt` is checked, and compatibility has fixtures (#522)

ADR 0044 decision 7 names two gaps in this ADR. This amendment records how
they were closed.

**A card restored with fewer abilities than captured.** Decision 2 marks a
card's ability closures `rebuilt`: restore looks them up in the catalog
again. Until this change nothing checked that it got back what was
captured. Two things are compared now, by one function,
`abilityShortfallOf`:

- the instance counts capture always wrote (`manaAbilityCount`,
  `activatedAbilityCount`), which only tokens and token copies carry;
- a new per-card record, `catalogAbilities`. For every card that had a
  catalog entry at capture, it stores the entry's mana, activated,
  triggered, static and replacement counts. Printed cards need it
  because they read their abilities from the catalog when used, so a
  vanished entry was invisible to the instance counts.

A shortfall in any slot, or an entry that no longer exists, is not a
reason to abandon the game (owner decision on #515). This is a deliberate
exception to Decision 5's abandon policy. That policy is for a file the
binary cannot read. Here the binary reads the file correctly, and the one
card it cannot automate stays playable by hand.

When a card falls short:

- the table is restored;
- the card is flagged `Card.AbilitiesLostOnRestore`, which `Unimplemented`
  reads, so the client shows the `manual` chip;
- the boot log gets one ERROR line per card, and the restore summary
  counts the affected games and cards.

The flag is carried by every later snapshot. Nothing clears it except the
card leaving the game. More abilities than captured is not a shortfall.

Both fields are additive and zero-value correctly in both directions, so
this is not a schema bump.

**Compatibility has fixtures.** Three things now enforce the rule next to
`SnapshotSchemaVersion`:

- a reflected shape file per schema version, which any non-additive
  change must bump past;
- a frozen set of generated restore points per version;
- scrubbed real restore points from cmd-dev, added by
  `cmd/snapshotscrub`.

Every CI run restores every fixture. A key in a fixture that a fresh
capture no longer carries fails the build, and that is what a rename looks
like from yesterday's file. Fixtures are never rewritten. A bump adds a
directory, and a bump that migrates a field lists the migrated paths in
the corpus test.

## Amendment, 2026-09-24 — phase 3: effects as data (#1497)

*Amendment, 2026-09-24, branch `docs/1497-adr-0041-phase3`. This is
the design step for
[#1497](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1497). The
owner approved the sprint in [ADR 0044](0044-surviving-a-deploy.md)'s
2026-09-24 amendment, decision 7. Docs only: no engine code changes
with this amendment.*

This amendment replaces the priority order in "Phase 3 (not in this
change)" above. Decisions 1 to 5 are unchanged.

### What blocks a restore point today

`CaptureSnapshot` builds a `ContinuationCensus` (`snapshot.go`). When
the census is non-empty the capture is not a restore point, and
`Room.writeRestorePointLocked` (`ws/persist.go`) keeps the previous
file without logging anything. `Room.Apply` tries a capture after
every action that advances `seq` (`ws/room.go`). **So a capture point
is the end of an action.** A capture never sees the state in the
middle of an action.

These are the census counters and what increments them, as of
`origin/develop` at `9c3d26ff`:

| Counter | What increments it | How long it lasts |
|---|---|---|
| `ScopedStatics` (wire key `turnScopedStatics`) | Every `Game.ScopedStatics` entry. `ScopedStatic.Ability` is two closures, and `ScopedStatic.Source` is a whole `Card` | The entry's `Duration`: until cleanup, until a player's next turn, while a source remains, or the rest of the game |
| `DelayedTriggerEffects` | Every `DelayedTrigger` with an `Effect`, which is all of them | Until it fires: usually the next end step, sometimes several turns |
| `TurnScopedReplacements` | Every `Game.TurnScopedReplacements` entry (Fog and the like) | Until cleanup |
| `TurnScopedBlockRules` | Every `Game.TurnScopedBlockRules` entry (Gingerbrute) | Until cleanup |
| `StackEffects` | Every stack item with an `Effect` closure: triggered and activated abilities. Spells resolve through the catalog and are not counted | While on the stack |
| `StackTargetSpecs` | An ability item's target or mode spec. The item does not record where the spec came from | While on the stack |
| `ChoiceResumeFrames` | A pending choice that holds one of its 17 resume frames or a prompt run | While the prompt is open |
| `IntrinsicAbilityCards` | An ability stamped on one instance with no catalog entry behind it | Rare since #521 |

`DelayedTrigger` is already mostly data. `Controller`, `At`, `On`,
`Cards`, `Duration`, `SourceObject` and `CreatedSeq` are all carried
today. Only `Effect`, and `AppliesTo` on an event-conditioned trigger,
are closures.

### Measurement: what the scoped statics actually are

Only the following functions register a scoped static. Everything else
calls one of them.

- In `internal/game`:
  - `animateEarthbentLandLocked`: three statics (add Creature, base P/T
    0/0, haste), indefinite and pinned.
  - `grantAmassSubtypeLocked`: add a subtype, indefinite and pinned.
  - `GainControlForEffect` and `ExchangeControlForEffect`, through
    `controlStatic` in layer 2.
  - `resolveProwess`: +1/+1 until end of turn.
  - `grantHasteForCastLocked`: suspend's haste, for as long as the
    player controls the card.
- In `cards/effects`:
  - `BoostUntilEOT`, `GrantKeywordUntilEOT`, `RestrictUntilEOT` and
    `GrantAllCreatureTypesUntilEOT`.
  - `BecomeArtifactCreature`, which every crew ability reaches through
    `CrewEffect`.
  - `StaticForDuration` and its wrapper `StaticUntilEOT`.
  - Two direct calls, in `tree_of_perdition.go` and `pupu_ufo.go`.

115 card files in `cards/effects` create a scoped static. The table
groups them by what the effect does. Some files do two things, so the
rows add up to 132.

| What it does | Layer | Card files | Written with |
|---|---|---|---|
| Modify P/T by +N/+M | 7c | 40 | `BoostUntilEOT` |
| Add keywords | 6 | 55 | `GrantKeywordUntilEOT`, directly or through three batch helpers |
| Become an artifact creature: add types and subtypes, set base P/T | 4, 7b | 16 | `CrewEffect` / `BecomeArtifactCreature` |
| Add restriction bits | 6 | 5 | `RestrictUntilEOT` |
| Set controller | 2 | 5 | `GainControl` (Act of Treason, Coercive Recruiter, Agent of Treachery, Sower of Temptation) and `ExchangeControl` (Switcheroo) |
| Be every creature type | 4 | 1 | `GrantAllCreatureTypesUntilEOT` |
| A raw `StaticAbility` | various | 10 | `StaticForDuration`, `StaticUntilEOT` or a direct call |

A bespoke closure could only be in the ten raw-ability files. None of
them is bespoke. Each one does one of five more operations:

- **Set base P/T:** Katara, Water Tribe's Hope (X/X), Mass Diminish
  (1/1, until your next turn), Sudden Spoiling (0/2), PuPu UFO (power
  only) and Tree of Perdition (toughness only, indefinite).
- **Add a subtype:** Coercive Recruiter (Pirate) and The Legend of
  Kyoshi, chapter II (Island, indefinite).
- **Remove a card type:** Enduring Curiosity ("it's not a creature",
  while its source remains).
- **Set colours:** Cerulean Wisps.
- **Remove keywords:** Shadowspear. Its affected set is not pinned at
  resolution: its `AppliesTo` reads the controller again on every layer
  pass. It is the only effect like this.

Sudden Spoiling also removes all abilities. That is the existing
`RemovesAbilities` declaration.

**Result: no scoped static in the tree is a bespoke closure.** Twelve
operations, each with a few plain parameters and applied to a pinned
set of objects, cover all 115 files and every engine site. That result
is the basis for Decision P1.

Delayed triggers are the same, on a smaller scale:

- 19 files use `ScheduleDelayedTrigger`, and between them they name 14
  distinct bodies.
- Every body is already a package-level function, as the AGENTS.md
  recipe requires.
- Three of the bodies are built by a factory from a captured argument:
  `pactPayment(cost, name)`, `arcaneDenialDraws(victim)` and
  `manaDrainRefund(mv)`.
- Four engine sites use inline closures: warp's exile, madness's
  decline, cascade's bottom, and earthbend's event-conditioned return.
  They capture a player, a card and an epoch.
- Doublecast and Galvanic Iteration use `WhenYouNextCast` with a card
  predicate.

Every captured value is a player ID, an object reference, a mana value
or a cost string.

The turn-scoped replacements come from:

- Fog's four users of `PreventAllCombatDamageThisTurn`;
- Mending Hands (`PreventNextDamage`);
- Whip of Erebos and Unearth (`ExileInsteadOfLeavingBattlefield`);
- Druid's Deliverance and Cosmic Intervention.

The turn-scoped block rules come from Gingerbrute and Departed
Deckhand (`BlockRuleUntilEOT`).

### Real-table evidence

The #524 shutdown census has been running for only a few hours.
cmd-dev's journal (`192.168.200.35`) holds 19 shutdowns across four
games, all on 2026-09-24 between 16:47 and 19:24 UTC. Every line reads
`"clean":true` and `"seq_behind":0`. Every table was idle at all 19
deploys, so the census had no blocker to see. The prod journal
(`192.168.200.33`) could not be read, because the deploy user needs a
sudo password there to read the unit's logs. The owner is adding the
deploy user to `systemd-journal` on both hosts by hand (owner decision
5).

So there is no ranking from real tables yet. The order below rests on
static reasoning: how long each blocker holds a restore point back,
weighed against how common the cards that create it are. The census
cannot rank blockers in any case. It reads one instant per deploy, and
it misses the case that matters most, a deploy in the middle of a
game. PR 1 therefore adds a tally per blocker kind (Decision P7).

### Decision P1 — a scoped continuous effect is a data record over a closed set of operations

A scoped continuous effect is no longer a `StaticAbility` with a
duration attached. It becomes a record the engine interprets:

```go
// ScopedEffect is a continuous effect created by a resolving spell or
// ability (CR 611.2). It holds no function and no pointer, so the
// snapshot carries it exactly as written.
type ScopedEffect struct {
    ID         uuid.UUID
    Affected   []AffectedObject // CR 611.2c: the set, locked at creation
    Mods       []Mod            // each applied in its own layer, all at one timestamp
    Source     ObjectRef        // the spell or ability that created it (an LKI identity)
    SourceName string           // for the label and the wire
    Controller uuid.UUID        // who "you" is for this effect, read once (CR 611.2c)
    Timestamp  int64            // CR 613.7; the two halves of an exchange share one
    Duration   Duration         // ADR 0063, unchanged
    Label      string
}

// AffectedObject is the key every CR 611.2c set in the engine already
// uses: an instance plus its battlefield-entry stamp, so a permanent
// that left and came back is a new object (CR 400.7). EnteredAt 0
// means "this instance, wherever it is". It is legal only when the
// Duration pins the object across the move: suspend's haste, which
// follows the card from the stack onto the battlefield.
type AffectedObject struct {
    ID        uuid.UUID `json:"id"`
    EnteredAt int64     `json:"enteredAt"`
}

// Mod is one operation. Kind is a string on disk, never an int, so
// renumbering constants cannot change what an old file means.
type Mod struct {
    Kind         ModKind     `json:"kind"`
    Types        []string    `json:"types,omitempty"`
    Subtypes     []string    `json:"subtypes,omitempty"`
    Colors       []string    `json:"colors,omitempty"`
    Keywords     []string    `json:"keywords,omitempty"`
    Restrictions Restriction `json:"restrictions,omitempty"`
    Power        int         `json:"power,omitempty"`
    Toughness    int         `json:"toughness,omitempty"`
    Player       uuid.UUID   `json:"player,omitempty"`
    Keys         []string    `json:"keys,omitempty"` // ADR 0093 grant bundles
}
```

The kind fixes the layer, so a record cannot put an operation in the
wrong layer. There are thirteen operations. Twelve cover everything
measured above, and the thirteenth, `grantAbilities`, is ADR 0093's.
7b has two kinds, so that makes fourteen kind strings.

| Kind | Layer | Covers |
|---|---|---|
| `setController` | 2 | Gain control, exchange of control (ADR 0063 D4/D5) |
| `addTypes` | 4 | Earthbend's Creature, crew's Artifact Creature |
| `removeTypes` | 4 | Enduring Curiosity |
| `addSubtypes` | 4 | Amass, Kyoshi II, Coercive Recruiter, crew's subtypes |
| `allCreatureTypes` | 4 | Shields of the Velis Vel (`Characteristic.AllCreatureTypes`) |
| `setColors` | 5 | Cerulean Wisps |
| `addKeywords` | 6 | Every `GrantKeywordUntilEOT`, earthbend's and suspend's haste. Goes through `AppendKeywordAbility`, so a cumulative keyword stays cumulative |
| `removeKeywords` | 6 | Shadowspear |
| `loseAllAbilities` | 6 | Sudden Spoiling. Sets `RemovesAbilities` and appends `Keywords` back (ADR 0046) |
| `addRestrictions` | 6 | Every `RestrictUntilEOT` |
| `grantAbilities` | 6 | ADR 0093's `ScopedGrant` (Decision P6) |
| `setBasePower` / `setBaseToughness` | 7b | Earthbend 0/0, crew, Mass Diminish, Katara, Sudden Spoiling, PuPu UFO (power only), Tree of Perdition (toughness only) |
| `modifyPT` | 7c | Every `BoostUntilEOT`, prowess |

**One record, several layers, one timestamp.** Earthbend's three
statics become one record with three mods. That is what CR 613.7 and
ADR 0063 D5 already required when they had three registrations share
a timestamp.

**The layer pass adapts each record.** `activeStaticAbilitiesLocked`
(`layers.go`) turns each live record into one `ContinuousEffect` per
mod. `controlStatic` already does this, and ADR 0093 D7 does it for
`ScopedGrant`. The running binary builds the adapter's closures from
the record, so they are **rebuilt**, not persisted. They are memoised
in a field the drift plan classifies `rebuilt`, so the pass does not
allocate on every recompute. (The memo landed with #1558, keyed by
record identity rather than by `layerVersion`; the implementation
notes below say why.)

**Expiry does not change.** `sweepScopedStaticsLocked` sweeps
`Game.ScopedEffects` through the same `durationExpiredLocked` as
before. The pin, the three sweep sites and ADR 0063's four kinds stay
as they are.

**The registry key, and why there is no per-card registry.** The
twelve kinds do not use a per-card registry key. The `ModKind` string
is the key, and the engine owns its namespace. This is ADR 0093 D7's
pattern, "a registry key plus plain parameters, re-derived on
restore", with the engine as the registry. A per-card registry would
mean 115 registrations of the same few bodies, which is exactly what
the clone gate exists to stop.

**A `ModKind` string is an on-disk identity**, like a token slug or a
grant key:

- It is never renamed and never reused.
- A mod whose meaning changes gets a new kind.
- The old kind keeps its old meaning for as long as
  `minRestorableSchema` still admits files that use it.

**The escape hatch is designed but not built.** The first effect that
none of the kinds can express gets a `catalog` mod, which adds a
`Key` field to `Mod`:
`Mod{Kind: "catalog", Key: "<oracle_id>|scoped:<name>", ...params}`.

- The key resolves to a builder that the card declares on its `Spec`.
- The builder is filed under that composite key in the same def map as
  `GrantKey`, `EmblemKey` and `TokenKey`.
- It is handed the record's plain parameters.

Phase 3 does not build it, because no card needs it yet. It lands with
the first card that does, and uses this naming.

**Allowed parameter types.** These apply to `Mod` and to Decision
P2's `EffectParams`.

- **Allowed:** `bool`, `int`, `int64`, `string`, `uuid.UUID`,
  `ObjectRef`, `AffectedObject`, `Duration`, the engine's string and
  bit enums (`Restriction`, `Step`, `ZoneKind`, `ModKind`), and slices
  of any of these.
- **Not allowed:** a `func`, an `interface`, any pointer, a map whose
  keys are not strings, a `Card` or `*Card`, and a `CardPredicate`.

The shape file records the type of every path, so any such type shows
up in review. A `func` field also fails JSON encoding outright.

**Restore takes nothing from the catalog for these records.** The
snapshot carries each record verbatim, and the running binary's
interpreter decides what each kind means. That is deliberate: a deploy
that fixes how `addKeywords` behaves also fixes it in every restored
game. The only check on restore is that the binary knows every
`ModKind` in the file (Decision P4).

### Decision P2 — a delayed trigger names its body by key

`DelayedTrigger.Effect` and `DelayedTrigger.AppliesTo` are replaced by
keys into one registry of **effect bodies**:

```go
type DelayedTrigger struct {
    // ... every data field it has today, unchanged ...
    Body       string       // e.g. "flicker/return-exiled-to-owners"
    Params     EffectParams // the factory's arguments, as data
    Condition  string       // event-conditioned triggers only, e.g. "earthbend/this-object-left"
    CondParams EffectParams
}

// EffectParams is the closed set of values a body may be handed. A body
// that needs a new kind of value gets a new field here, of an allowed
// type. It never gets an interface{}.
type EffectParams struct {
    Player uuid.UUID  `json:"player,omitempty"`
    Object ObjectRef  `json:"object,omitempty"`
    Amount int        `json:"amount,omitempty"`
    Cost   string     `json:"cost,omitempty"`
    Name   string     `json:"name,omitempty"`
    Filter CastFilter `json:"filter,omitempty"` // WhenYouNextCast's predicate, as data
}

type BodyFunc func(g *Game, item *StackItem, p EffectParams) error
type ConditionFunc func(ev Event, dt *DelayedTrigger, g *Game) bool
```

**Registration.** Each body is registered once, at `init`, in the file
that defines its function:

```go
var FlickerReturnExiled = game.DelayedBody(
    "flicker/return-exiled-to-owners", returnExiledCardsToOwners)
```

- `DelayedBody` panics on a duplicate key, the way `Register` does.
- It returns a `game.BodyRef`, a struct that holds only the key.
- `ScheduleDelayedTrigger.Effect` changes type from a function to a
  `game.BodyRef`. **A card file cannot pass a func literal, because it
  no longer compiles.** That is the lint.

**Keys.** A key has the form `<area>/<name>`, lowercase and
hyphenated. The `area` is the mechanic or card that owns it (`flicker`,
`warp`, `madness`, `mana-drain`), never a batch number. Keys are
on-disk identities, so the set of keys only grows:

- `server/internal/game/testdata/effect_keys.txt` is the ledger.
- `TestEveryPersistedEffectKeyResolves` fails if a ledger line does
  not resolve, or if a registered key is not in the ledger. `-update`
  appends the missing ones.
- To rename a key, register the new key and keep the old one with
  `game.EffectAlias(old, new)`. Never delete a key.

**Factories become parameters.**

| Today | After |
|---|---|
| `pactPayment("{3}{U}{U}", "Pact of Negation")` | `Body: PactPayment`, `Params{Cost, Name}` |
| `arcaneDenialDraws(victim)` | `Params{Player: victim}` |
| `manaDrainRefund(mv)` | `Params{Amount: mv}` |
| The four engine sites, which capture a player, a card and an epoch | `Params{Player, Object}` |

**Conditions.** An event-conditioned trigger (#663) names its predicate
the same way:

- `WhenYouNextCast(Or(Instant(), Sorcery()))` becomes a `CastFilter`,
  read by one registered condition.
- `CastFilter` is a small data struct beside `PermissionFilter`, with
  fields such as `InstantOrSorceryOnly`, `Types` and `NonCreature`.
- A filter `CastFilter` cannot express gets its own condition key,
  per card. That is the escape hatch.

**The fired stack item carries the key.** When a delayed trigger goes
on the stack, it makes a `StackItem` with a nil `Effect` and new
`Body` and `Params` fields that name the body. Resolution looks the
body up. So a fired delayed trigger is still data while it waits on the
stack, for example while a counterspell answers it. This is the first
half of tier 4 (Decision P5), and it costs nothing extra here.

### Decision P3 — what makes a closure acceptable

After phase 3, a function value that can be reached from `*Game` is
acceptable only if it is one of these three kinds:

1. **Rebuilt.** It is installed by `NewGame`, or the running binary
   derives it from carried data. Listeners, built-in replacements and
   Decision P1's adapter memo are rebuilt. The drift plan classifies it
   `rebuilt`, and #522's shortfall check proves the rebuild happened.
2. **Keyed.** It is looked up in a registry by a carried key. Catalog
   abilities by `CatalogKey`, token templates, grant bundles and effect
   bodies are keyed. Restore refuses a key it does not know (Decision
   P4).
3. **Transient.** It cannot be alive at a capture point. A capture
   point is the end of a `Room.Apply` (and the shutdown census, which
   reads the same state). A closure that is created and used up inside
   one engine call is transient. Examples: a `Then` continuation that
   runs before the call returns, and a local `AppliesTo` inside a
   sweep.

   A closure stored on `Game`, a `Card`, a `StackItem`, a
   `PendingChoice` or a `DelayedTrigger` is **not** transient, however
   short-lived it usually is. If it can outlive the return of `Apply`,
   it is a blocker.

Anything else increments a census counter. **The census counters are a
fixed list that only gets shorter**, and three mechanisms enforce that:

- **A ratchet over the type graph.**
  `TestClosureFieldsReachableFromGame` walks by reflection from `Game`
  through struct fields, slices, maps, pointers and named types. It
  lists every route to a function or an interface: every field of
  every reachable struct whose type reaches one, for example
  `Game.ScopedStatics`, `ScopedStatic.Ability` and
  `StaticAbility.AppliesTo`. It compares the list with
  `testdata/closure_fields.txt`.
  - Each line in that file is `<Type>.<Field> rebuilt|keyed|census:<Counter>`.
  - A new route fails the build, including a new field whose type is a
    struct that is already listed (#1558).
  - A `census:` line whose counter has been retired fails the build.
  - Tier PRs delete lines. The number of `census:` lines per counter,
    and of `transient` lines, is pinned in the test and may only fall
    (#1558).

  The drift test classifies seven types by field name. This test also
  follows what those fields hold, all the way down, which is where a
  closure can hide.
- **Retirement at compile time.** When a tier lands, its closure-taking
  API is deleted, not deprecated. That covers
  `RegisterScopedStaticForEffect`, `StaticForDuration`,
  `StaticUntilEOT`, `ScopedStatic.Ability`, and
  `ScheduleDelayedTrigger.Effect` as a function type. Once they are
  gone, nothing accepts a closure for that state, so no card can bring
  one back. The counter's struct field stays, like `UnpersistableRNG`,
  so an old census still decodes. Nothing increments it.
- **A soak check.** The catalog soak takes a capture after every
  action (Decision P7). A retired kind in the census fails the run.
  This catches a path the type graph cannot see, such as one hidden
  behind an interface.

### Decision P4 — schema: v7, once, with the reader for tiers 1 and 2

In shape, the new state is purely additive:

- a new `scopedEffects` array;
- new `body`, `params`, `condition` and `condParams` keys on a delayed
  trigger and on a stack item.

**Older files need nothing.** The census kept every live scoped static
and delayed trigger out of v6 restore points, so no v6 restore point
contains one. A v6 file therefore restores under v7 unchanged.
`corpusMigrations` stays empty, and the `v6/` and `real/` fixtures must
pass untouched.

**The other direction forces a bump**, for the same reason v2, v5 and
v6 gave. A v6 binary reading a v7 file would silently drop
`scopedEffects`. It would restore an earthbent land as a plain land,
and a warped creature with nothing left to exile it.

- **PR 1 bumps the schema to v7.** The v7 reader already understands
  **both** tiers. It carries `scopedEffects`, and it reads a delayed
  trigger's `body` even though PR 1 writes none. Delayed triggers as
  data (PR 2) is then an additive change under v7, with no second
  bump.
- **v7 refuses what it cannot interpret.** If a file contains a
  `ModKind`, body key or condition key the binary does not have,
  restore returns `ErrUnknownEffectKey`.
  - The game is abandoned and the file kept, as Decision 5 already
    does for a file that is too new.
  - Because of the ledger test, a key can only be missing after a
    rollback. So this is exactly the rollback case.
  - This refusal is what lets the vocabulary grow **within** v7. A new
    `ModKind` or body is an additive change that an older v7 binary
    refuses rather than misreads, so it needs no bump.
- **Tiers 3b and 4 bump again, to v8,** when they land. Their new
  fields are scoped replacements, block rules and ability refs on
  stack items. An earlier v7 binary would ignore those fields, and
  there is no way to make it refuse a single field.
- **Recording the new shapes.** The shape guard records `v7.txt`, and
  `TestWriteSnapshotCorpus` writes `v7/`. PR 1 adds scripted boards to
  the corpus writer so that `v7/` freezes the new shapes on the day
  they appear:
  - an earthbent land;
  - an amassed Army;
  - a Sower theft;
  - a Switcheroo exchange;
  - a suspended creature's haste;
  - a Mass Diminish.
- **One change to the corpus rule.** PR 2's shapes, a pending warp
  exile and a flicker return, would also belong in `v7/`. So the rule
  "the writer never touches an existing directory" narrows to "the
  writer never touches an existing **file**". A later PR under the same
  version may add a new, separately named fixture. It still may not
  edit, regenerate or delete any existing fixture, which is the
  invariant the rule exists for.
- **Drift test.** `ScopedEffect` joins `driftPlans` with every field
  `carried`, so `snapshot_carried_test.go` probes each one.
  `ScopedStatic.Duration` leaves `carriedNotRoundTrippable` once
  `ScopedStatic` is deleted.

### Decision P5 — migration order, longest-lived first

| Tier | Blocks a restore point for | Work | Counter retired |
|---|---|---|---|
| 1 | The rest of the game, or the life of a permanent | `ScopedEffect`. Engine sites: earthbend, amass, gain control and exchange, suspend's haste. Card sites: Tree of Perdition, Kyoshi II, Enduring Curiosity and, because it runs until your next turn, Mass Diminish | None yet: the old registry still exists |
| 2 | Several turns | Delayed triggers as data (P2): 14 card bodies, 3 factories, 4 engine sites, `WhenYouNextCast` | `DelayedTriggerEffects` |
| 3a | Until cleanup | Move the until-end-of-turn builders onto `ScopedEffect`: `BoostUntilEOT`, `GrantKeywordUntilEOT`, `RestrictUntilEOT`, `GrantAllCreatureTypesUntilEOT`, `BecomeArtifactCreature` / crew, and prowess. Their signatures do not change. Counting tier 1's `GainControl` and `ExchangeControl` users, **105 card files move without an edit**. Rewrite the six remaining raw-ability files as mods. Delete the old registry | `ScopedStatics` |
| 3b | Until cleanup | `ScopedReplacement` and `ScopedBlockRule` records, on the P1 pattern, each with its own closed set of kinds: prevent all combat damage (to everyone, or to one player), prevent the next N damage, exile instead of leaving, and the two block-rule shapes. Needs v8 | `TurnScopedReplacements`, `TurnScopedBlockRules` |
| 4 | While open | An ability's stack item carries ADR 0093 D5's stable ability `ref` (catalog key, slot, index). Restore reads `Effect`, `targetSpec` and `modeSpec` back from the catalog, as it already does for a spell. Triggers the engine owns (prowess, ward, cascade and the rest) use P2 body keys. Resume frames are out of scope (owner decision 4): `ChoiceResumeFrames` stays the one allowed counter | `StackEffects`, `StackTargetSpecs` |

**Tier 3a is cheap for its size.** Crew and prowess probably make it
the most frequent blocker on a real table, yet it changes only seven
builder bodies. It still follows tier 2 (owner decision 3), and the
order is revisited once P7's numbers are in.

#### The first slice

This amendment is the ADR PR. The code comes after it.

**PR 1: `ScopedEffect` and the effects that last all game.**

- In `internal/game`:
  - the `ScopedEffect`, `AffectedObject` and `Mod` types, and their
    kinds;
  - the layer-pass adapter with its memo, and the sweep.
- The snapshot:
  - v7, which carries `scopedEffects` and reads the P2 fields;
  - the refusal of unknown kinds;
  - the `driftPlans` and shape entries;
  - the `v7/` corpus boards named in P4.
- The guards: `TestClosureFieldsReachableFromGame` with its ratchet
  file, and the per-kind skip tally (P7).
- Engine sites migrated:
  - `animateEarthbentLandLocked` and `grantAmassSubtypeLocked`;
  - `GainControlForEffect` and `ExchangeControlForEffect`;
  - `grantHasteForCastLocked`.
- Card sites migrated: Tree of Perdition, The Legend of Kyoshi,
  Enduring Curiosity and Mass Diminish move from `StaticForDuration`
  and `RegisterScopedStaticForEffect` to `ScopedEffectFor` (P6).
- Unchanged: `effects.GainControl` and `ExchangeControl` keep their
  signatures. Act of Treason, Coercive Recruiter's steal, Agent of
  Treachery, Sower of Temptation and Switcheroo do not change a line.
- Tests:
  - every migrated card's existing test passes unchanged;
  - each migrated shape gets a test that restores mid-effect. It
    captures, restores, and asserts what a player can see: P/T, types,
    controller and haste. It also asserts that the effect still
    expires at the right moment in the restored game.

**PR 2: delayed triggers as data.** P2 in full:

- `BodyRef` and the ledger;
- `CastFilter`;
- the four engine sites;
- the fired stack item that carries its data.

It retires `DelayedTriggerEffects`. It needs no bump, because v7
already reads these fields.

PRs 3a, 3b and 4 follow in that order (owner decision 3).

### Decision P6 — what a card file writes afterwards

**Most card files change nothing.** These constructors keep their
names and signatures, and only their bodies change:

- `BoostUntilEOT`, `GrantKeywordUntilEOT`, `RestrictUntilEOT` and
  `GrantAllCreatureTypesUntilEOT`;
- `BecomeArtifactCreature` and `CrewEffect`;
- `GainControl` and `ExchangeControl`;
- `ScheduleDelayedTrigger`, apart from its `Effect` field.

**The raw-`StaticAbility` escape hatch is replaced** by one constructor
over the mod vocabulary:

```go
// "Until your next turn, creatures target player controls have base P/T 1/1."
ScopedEffectFor{
    Match:    And(Creature(), ControlledBy(victim)), // snapshotted once (CR 611.2c)
    Mods:     []game.Mod{game.SetBasePT(1, 1)},
    Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
    Label:    "Mass Diminish — base power and toughness 1/1",
}.Apply(ctx)
```

- The mod constructors live in `internal/game`, beside the kinds:
  - `SetController(p)`;
  - `AddTypes(...)`, `RemoveTypes(...)`, `AddSubtypes(...)` and
    `AllCreatureTypes()`;
  - `SetColors(...)`;
  - `AddKeywords(...)`, `RemoveKeywords(...)`,
    `LoseAllAbilities(keep...)`, `AddRestrictions(bits)` and
    `GrantAbilities(keys...)`;
  - `SetBasePT(p, t)` (which returns both 7b mods), `SetBasePower(n)`,
    `SetBaseToughness(n)` and `ModifyPT(p, t)`.
- `ScopedEffectFor.Target` pins one object, as `BoostUntilEOT.Target`
  does.
- `Match` is always resolved to a pinned set at `Apply`.

A delayed trigger is written like this:

```go
ScheduleDelayedTrigger{
    Label: "Waterbender's Restoration — return the exiled creatures",
    Cards: exiled,
    Body:  FlickerReturnExiled, // a game.BodyRef, registered once at init
}.Apply(ctx)

ScheduleDelayedTrigger{
    Label:              "Mana Drain — add {C} equal to its mana value",
    At:                 game.StepPrecombatMain,
    ControllerTurnOnly: true,
    Body:               ManaDrainRefund,
    Params:             game.EffectParams{Amount: mv},
}.Apply(ctx)
```

**The lint** is Decision P3's:

- The closure-taking APIs are deleted, so a card cannot hand a closure
  to a scoped registry or to a delayed trigger at all.
- The type-graph ratchet catches a new closure-bearing field anywhere
  else.
- Between PR 1 and PR 3a, `StaticForDuration` still exists. For that
  period a source-scan lint, written like `color_purpose_guard_test.go`,
  fails any file outside an allowlist that calls `StaticForDuration`,
  `StaticUntilEOT` or `RegisterScopedStaticForEffect`. The allowlist
  only shrinks.

**AGENTS.md** changes with each code PR, not with this amendment:

- The "Durations" paragraph and the delayed-trigger recipe will
  describe the new constructors.
- "Snapshot compatibility" gains the command for the effect-key ledger.

**ADR 0093 Decision 7.** `ScopedGrant` becomes the `grantAbilities`
mod of a `ScopedEffect`, not a second registry. It keeps everything
0093 D7 asks for: the pinned object set, the timestamp, the
`Duration`, the sweep, and `carried` in the drift plan. It gets them
through the same record and the same adapter. The record lands first
(owner decision 1): phase 3's PR 1 adds a one-line amendment to 0093
D7 pointing here, and 0093's PR 4 builds `ScopedGrant` as this mod.

### Decision P7 — measure before tiers 3 and 4

The shutdown census reads one instant per deploy. Across 19 deploys it
could not rank anything, because every table was idle. PR 1 adds two
cheap sources of evidence.

- **A skip tally per room.** `writeRestorePointLocked` counts skipped
  captures per census kind (`skipped_by_kind`). It also tracks how many
  actions the current run of skipped captures has lasted. The shutdown
  census line logs both. This adds no disk write and no log line per
  action. On the hot path it costs nothing beyond the capture that is
  already made.
- **The catalog soak as a probe.** The nightly soak plays heuristic
  seats with catalog-dealt decks through a real `ws.Room`. PR 1 makes
  it:
  - capture after every action;
  - tally census kinds and the longest stale run for each kind;
  - write both into its JSON report next to the per-card coverage.

  This is not real-table evidence. But it covers every action of every
  soak game, across the whole catalog, every night. If the real census
  stays quiet, it decides the order of tiers 3 and 4.

### What this amendment deliberately does not do

- **It does not persist the undo stack.** ADR 0044 decision 6 still
  holds.
- **It records a rules question the migration uncovers, and does not
  fix it.**
  - `RestrictUntilEOT`'s mass form pins its affected set.
  - CR 611.2c pins only effects that modify characteristics or change
    control.
  - A "can't block this turn" from a resolving spell arguably changes
    the rules rather than a characteristic, and if so it should also
    cover creatures that enter later.
  - The engine models restrictions as layer-6 bits, so pinning is
    today's behaviour. Tier 3a keeps it exactly.
- **It does not change any card's behaviour**, except Shadowspear,
  whose affected set is pinned at resolution (owner decision 2). The
  PR that migrates it calls the change out.

### Owner decisions (2026-09-24)

The owner answered all six open questions this amendment was proposed
with, taking each recommendation.

1. **The general record comes first.** Phase 3's PR 1 lands
   `ScopedEffect`. ADR 0093's `ScopedGrant` becomes its
   `grantAbilities` mod rather than a registry of its own, and PR 1
   adds the one-line amendment to 0093 D7 pointing here.
2. **Shadowspear is pinned at resolution** (CR 611.2c): a permanent
   that enters after the activation keeps hexproof and
   indestructible. The PR that migrates it calls out the behaviour
   change. No dynamic `Selector` is added to `ScopedEffect`.
3. **The order stays:** tier 2 (delayed triggers) before tier 3a
   (until end of turn). It is revisited once P7's numbers are in.
4. **Resume frames are out of scope for this sprint.**
   `ChoiceResumeFrames` is the one census counter that stays allowed
   after tier 4, and `closure_fields.txt` says so on its lines.
5. **Journal access is an owner action, not a CD step.** The owner adds
   `krkn` to `systemd-journal` on both hosts, so the #524 shutdown
   census can be read on prod as well as dev.
6. **v7 now, v8 later.** v7 covers tiers 1 and 2; v8 comes with tiers
   3b and 4. The corpus rule narrows from "the writer never touches an
   existing directory" to "the writer never touches an existing
   **file**" — a fixture is still never edited, regenerated or
   deleted. *Second half superseded 2026-09-25 by Decision P10:
   tiers 3b and 4 stay in v7 (owner answer Q1).*

### Implementation notes (PR 1, #1555)

Three things the first slice settled that the decisions above left open.

- **P4 covers durations and fields, not just mod kinds.** A duration's
  `Kind` and `Condition` are bare ints on disk, so an unknown one would
  have restored as an effect that never ends, or fallen through to a
  bare "source on battlefield" test. Restore refuses both with
  `ErrUnknownEffectKey`, and it refuses an unknown JSON field on a
  record, an affected member, a mod or its duration, which
  `encoding/json` would otherwise drop without a word. So a new field on
  any of those four types, like a new mod kind, is additive within v7:
  an older v7 binary refuses the file rather than misreading it.
  Changing what an existing kind or field means is still a new kind or a
  bump.
- **The pin sees a phased-out permanent.** `Duration.Pinned` is garbage
  collection for an object that is gone, and a phased-out permanent is
  not gone (CR 702.26d). The pin also looks in `g.PhasedOut`, so an
  indefinite effect applies again when its object phases in. A
  ForAsLongAs condition still reads the battlefield alone, because
  CR 702.26f ends a duration that tracks a phased-out permanent.
- **`grantAbilities` is not declared yet.** It lands with ADR 0093's PR 4,
  together with the grant seam it adapts into. *(Landed with #1584: see
  ADR 0093's "Amendment 2026-09-24 — what PR 4 built".)*

### Implementation notes (PR 2, tier 2: delayed triggers as data)

- **No schema bump.** v7's reader has had the `body` and `condition` keys
  since PR 1. The `params`, `condParams` and `optionalQuestion` keys are
  additive, and are omitted when zero. A PR 1 binary handed a PR 2 file
  meets a delayed trigger or a fired item with a `body` it cannot resolve,
  and refuses the file (`ErrUnknownEffectKey`), which is the rollback case
  P4 designed for.
- **The registry lives in `game/effect_bodies.go`.** Bodies and conditions
  are registered at init. The catalog registers its keys in one place,
  `cards/effects/delayed_bodies.go`, rather than one line beside each
  function, so that a reviewer adding a key sees all of them.
  - The engine registers its four bodies (warp, madness, cascade and
    earthbend's return) in `init` functions, because a package-level var
    would be an initialisation cycle through the exit primitives.
  - The registry has a lock, because tests register throwaway `test/…`
    bodies while other tests' games resolve. The ledger ignores that
    namespace.
- **`ConditionFunc` receives its params.** Its signature is
  `func(ev, dt, g, EffectParams) bool`. The params duplicate
  `dt.CondParams`, and are passed so that a condition never has to
  unpack them.
- **A delayed trigger's "you may" is a question, not a prompt.**
  `DelayedTrigger.OptionalQuestion` replaces the `Optional`
  `*TriggerOptionalPrompt`, because that prompt's `Chooser` is a closure
  and a delayed trigger always asks its controller.
- **A fired item is data too.** `StackItem.Body` and `Params` are set on
  the item a delayed trigger puts on the stack. The snapshot counts only
  an unkeyed `Effect`, and restore re-derives the `Effect` from `Body`.
- **`ContinuationCensus.DelayedTriggerEffects` is retired.**
  `ScheduleDelayedTriggerForEffect` refuses a trigger with no body. The
  counter now fires only for a hand-built trigger with nothing to do. The
  closure ratchet lists the counter in `retiredCensusCounters`.
- **Review fixes (#1568):**
  - **Unknown fields are refused wherever params live.** The P4 refusal of unknown JSON fields now covers every `params` and `condParams` on a delayed trigger or a stack item, down through `filter` and `object`, not only scoped effects.
  - **`CastFilter.Types` is a closed set.** It must be one of CR 205.2a's card types, spelled exactly, and it is matched against the card's type list as whole words, never as a substring. It is refused both where it is scheduled and at restore.
  - **The engine's `DelayedTrigger.Body` and `.Condition` are typed refs.** They are `BodyRef` and `ConditionRef`, so an unregistered key cannot be written at all. A zero ref is a programming fault (`effectKeyFault`): it panics in a test binary, and in production it is logged and the trigger is dropped. It never crashes the server.
  - **An ability copy of a keyed item is keyed too.**

### Implementation notes (#1558: hardening after the #1555 review)

- **The closure ratchet is keyed by route.** `closure_fields.txt` has a
  line for every field of every reachable struct that reaches a
  function or an interface, not only for the fields that hold one
  directly. A new field whose type is an already-listed struct is a new
  line, so it fails until it is classified. The probe from the review
  (`[]ReplacementEffect` on `TargetSpec`, `[]StaticAbility` on `Card`)
  now fails the build. The reachability walk is a fixed point, because
  a memoised walk scored a type inside a cycle wrongly.
- **Blocker classes may only shrink.** `closureClassCeilings` pins the
  number of lines in each `census:` class and in `transient`. A count
  above its ceiling fails. A count below it fails too, until the
  ceiling is lowered in the same PR.
- **The adapter memo landed.** It is keyed by record identity (the
  address of each record's `Mods`, plus its scalars), not by
  `layerVersion`. An undo stores the snapshot's version plus one, so one
  version number can name two different registries, and a
  version-keyed memo could return a record the undo removed. The memo
  is not cloned, and it is never mutated in place. At 100 records on
  one creature, one recompute went from 428 allocations and about
  26 µs to 20 allocations and about 11 µs. The legacy closure statics
  cost 121 allocations and about 19.5 µs.
- **An unstamped permanent is pinned exactly.** `PinObject` and
  `PinnedTo` record `unstamped` / `PinnedUnstamped` for a permanent
  with no entry stamp, and the pin matches it only while it is still
  unstamped. So a flicker ends the effect, as the legacy `== stamp`
  comparison did. `EnteredAt 0` with no flag is still the wildcard,
  and suspend's haste is its only writer. The flags are additive within
  v7 and omitted when false. An older v7 binary refuses a scoped
  effect that carries one (P4's unknown-field refusal). Anywhere else
  it drops the flag, which gives back the old wildcard.

### Implementation notes (tier 3a: until end of turn)

- **The closure registry is gone.** `ScopedStatic`, `Game.ScopedStatics`,
  `RegisterScopedStaticForEffect`, `registerScopedStaticLocked`,
  `StaticForDuration`, `StaticUntilEOT` and `SnapshotAffected` are
  deleted, not deprecated (P3's retirement at compile time). The two
  interim source-scan guards went with them, since nothing is left for
  them to find. `ContinuationCensus.ScopedStatics` stays as a field so
  an old census still decodes, nothing increments it, and the closure
  ratchet lists it in `retiredCensusCounters`. Its five
  `closure_fields.txt` lines are deleted and its ceiling with them:
  `StaticAbility` is no longer reachable from `Game` at all.
- **No new mod kind and no schema change.** Every migrated site is one
  of the kinds PR 1 declared. `BoostUntilEOT` is `modifyPT`,
  `GrantKeywordUntilEOT` is `addKeywords`, `RestrictUntilEOT` is
  `addRestrictions`, `GrantAllCreatureTypesUntilEOT` is
  `allCreatureTypes`, and prowess is `modifyPT`. The six raw-ability
  files became `setColors` (Cerulean Wisps), `addSubtypes` (Coercive
  Recruiter), `setBasePower` (PuPu UFO), `setBasePower` plus
  `setBaseToughness` (Katara), `loseAllAbilities` plus both 7b kinds
  (Sudden Spoiling) and `removeKeywords` (Shadowspear). The builders
  keep their names and fields, so no card calling them changed.
- **One record per effect.** Crew's type change and its base P/T, and
  Sudden Spoiling's ability loss and its 0/2, were two registrations
  with two clock reads. Each is now one record with one timestamp,
  which is what CR 613.7 means by one effect. The halves sit in
  different layers, so no board can tell the difference.
- **Shadowspear is pinned at resolution** (owner decision 2). A
  permanent that enters after the ability resolves keeps hexproof and
  indestructible, and one that changes control afterwards keeps the
  state it had. That is CR 611.2c, since abilities are characteristics
  (CR 109.3), so the card's caveat is dropped and it is declared
  `CompletenessFull`.
- **Fixtures.** `v7/until_eot_pump.json` (a real Giant Growth on a
  prowess creature, which gives two `modifyPT` records) and
  `v7/crewed_vehicle.json` (a real crew of Smuggler's Copter) are new
  files under the "never touch an existing file" rule. No existing
  fixture changed.

### Implementation notes (ADR 0093 PR 4: `grantAbilities`)

- **The thirteenth operation is declared.** `grantAbilities` (#1584) is
  a layer-6 kind that reads a new `Mod.Grants` list of catalog bundle
  keys (JSON `grants`, omitted when empty; the sketch above called it
  `keys`). The adapter gives the record's layer-6 `StaticAbility` a
  `GrantAbilities` list, so the grant is the same declaration a
  granting static makes, written in the record's timestamp slot.
- **No schema bump.** The kind and the field are additive within v7
  and recorded in the shape file. A v7 binary from before them meets an
  unknown kind and an unknown key, and refuses the file (P4).
- **The bundle is part of the key.** A restore point whose
  `grantAbilities` mod names a bundle this binary's catalog does not
  register, or names none, is refused with `ErrUnknownEffectKey`, like
  an unknown kind.
- **Fixture.** `v7/duration_grants.json` (a real Feign Death, and Fake
  Your Own Death's `modifyPT` plus `grantAbilities` record) is a new
  file. No existing fixture changed, and `closure_fields.txt` is
  unchanged.

## Amendment, 2026-09-24 — phase 3 tiers 3b and 4 (#1497)

*Amendment, 2026-09-24, branch `docs/1497-tiers-3b-4-design`. The
design step for the last two tiers of
[#1497](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1497). Docs
only: no engine code changes with this amendment. Decisions P8 to P11
were accepted on 2026-09-25; the owner's answers are at the end.*

This amendment is measured against `origin/develop` at `933cddfc`,
after tier 1 (#1555), tier 2 (#1568, #1582), the #1558 hardening
(#1601), tier 3a (#1602) and `grantAbilities` (#1603). It replaces the
3b and 4 rows of P5's table, and it proposes to replace P4's last
bullet ("tiers 3b and 4 bump again, to v8") and owner decision 6's
second half. It does not change P1 to P3, P6 or P7.

### What is left: the inventory

The closure ratchet (`testdata/closure_fields.txt`,
`closure_fields_test.go`) is the complete list, because it walks every
route from `Game` to a function or an interface. These are the classes
still on it and their pinned ceilings (`closureClassCeilings`):

| Class | Lines | Tier |
|---|---|---|
| `census:ChoiceResumeFrames` | 98 | None. Out of scope (owner decision 4) |
| `census:IntrinsicAbilityCards` | 45 | None. See "What stays census" |
| `census:TurnScopedReplacements` | 10 | 3b |
| `census:TurnScopedBlockRules` | 4 | 3b |
| `census:StackEffects` | 6 | 4 |
| `census:StackTargetSpecs` | 11 | 4 |
| `transient` | 1 | `Game.simultaneousExit`, never alive at a capture point |
| `test-only` | 1 | `Game.testReplacements` |

`DelayedTriggerEffects` and `ScopedStatics` are retired. Nothing else
bumps a census counter.

#### Tier 3b: what writes the two turn-scoped registries

Both registries are slices of closures on `Game`
(`game/game.go:717`, `game/game.go:739`), emptied at cleanup
(`game/rotation.go:86` and `:94`) and counted by the capture
(`game/snapshot.go:1492-1498`). They are read in three places: the
replacement gather (`game/replacements.go:1893`), the mana-production
replacement check (`game/produce_mana.go:300`) and the block-rule walk
(`game/block_rules.go:179`).

Every writer, and every card that reaches it:

| Writer | What it says | Cards | Census counter |
|---|---|---|---|
| `PreventAllCombatDamageThisTurn` (`cards/effects/prevention.go:70`) | Prevent all combat damage this turn | Fog, Holy Day, Tangle, Constant Mists | `TurnScopedReplacements` |
| `b38PreventAllCombatDamageToPlayerThisTurn` (`batch38_helpers.go:245`) | Prevent all combat damage that would be dealt to you this turn | Druid's Deliverance | `TurnScopedReplacements` |
| `PreventNextDamage` (`prevention.go:144`) | Prevent the next N damage to any target this turn. The charge is a captured variable that the closure mutates | Mending Hands | `TurnScopedReplacements` |
| `ExileInsteadOfLeavingBattlefield` (`exile_instead_of_leaving.go:50`) | If it would leave the battlefield, exile it instead | Whip of Erebos, Dregscape Zombie (unearth) | `TurnScopedReplacements` |
| inline (`cosmic_intervention.go:79`) | If a permanent you control would be put into a graveyard this turn, exile it instead and return it at the next end step | Cosmic Intervention | `TurnScopedReplacements` |
| `BlockRuleUntilEOT` (`cards/effects/block_rules.go:355`) | Can't be blocked this turn except by creatures with X | Gingerbrute (haste), Departed Deckhand (Spirits) | `TurnScopedBlockRules` |
| `EachOpponentCantBlockWithMoreThanN` (`block_rules.go:288`) | Each opponent can't block with more than N creatures this combat | Mirri, Weatherlight Duelist | `TurnScopedBlockRules` |

That is 12 card files out of about 2,300 registered specs. Fog-class
cards are uncommon in Commander, so these registries rarely block a
restore point, and they never block one past cleanup. The Whip is the
exception, and it is a bug:
[#1591](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1591). The
redirect expires at cleanup, so a creature whose end-step exile was
Stifled can die to the graveyard later and be whipped again.

#### Tier 4: what puts a closure on the stack

A stack item blocks a restore point when its `Effect` is a closure with
no `Body` (`game/snapshot.go:1795`), or when its target or mode spec
came off an ability declaration (`:1804`, `:1811`). A spell never
does, because restore re-derives a spell's spec from its oracle ID.
Every triggered and activated ability does, from the moment it is put
on the stack until it resolves, and so does every trigger waiting in
`Game.PendingTriggers`.

| Source of the item | Where the closure comes from | Scale in the catalog |
|---|---|---|
| A catalog activated ability | `ActivateCatalogAbility` copies the row's `Effect`, `Targets` and `Modes` onto the item (`game/activated.go:1454-1482`). Nothing is captured at activation time | 525 card files declare `Activated` |
| A catalog triggered ability built by a constructor | `OnAny` (`cards/effects/triggers_common.go:96`) and the constructors on it build `NewTriggeredItem(source, label, effect)`, and `effect` is the value the card registered | About 690 of the 1,009 files that declare `Triggered` use constructors only |
| A catalog triggered ability with a hand-written `Build` | 333 `Build` closures in 318 files. By signature, 153 read only the source, 125 read the event, 45 read the event and the board, and 10 read the board or LKI without the event. A sample shows that the source-only ones are the long form of a constructor (Archaeomancer, Esika's Chariot, Oubliette, Terastodon), and that the others capture trigger-time values: Suture Priest's victim, Terror of the Peaks' power at entry | 318 files |
| A keyword the catalog appends | `Ward`/`WardGranted` (19 files), `Cascade` and `Storm` (15), madness and suspend from `buildDef` | Constructors, so they follow the row above |
| An engine trigger with no catalog row | prowess (`game/prowess.go:78`), suspend's two (`suspend.go:165`, `:225`), madness (`madness.go:185`), the monarch's two (`monarch.go:119`, `:155`), evoke's sacrifice (`alternative_cost.go:802`), a mana-spend rider (`mana_spend_rider.go:320`) and face-down ward (`face_down.go:242`) | 9 sites |
| A reflexive trigger | `QueueReflexiveTriggerForEffect` wraps `ReflexiveTrigger.Effect` (`game/reflexive.go:164`). The recipe already requires a package-level function | 15 sites in 10 files |
| A modal ability's bullets | `ModeOption.Effect` on the ability's `ModeSpec` (`ModeDoing`) | 22 files. They sit on the declaration, so they follow the row |
| A copy of an ability | `ability_copy.go:227` copies `Effect` | Follows the original |

Two more records are affected by this tier:

- `TriggeredAbility.TargetsFrom` (#1223) builds a trigger's target
  clause from the trigger context. Its contract says it is a pure read,
  and 18 files use it.
- `Game.lastKnownStack` (`game/stack_lki.go:82`) is classified
  `dropped` by the drift test. The reason it gives is that its only
  readers are closure items. Storm's copy of a countered spell reads it.
  Once storm's trigger is data, that reason no longer holds.

The sandbox verbs `ActivateAbility` and `AnnounceTrigger`
(`game/mutations.go:3228`, `:3458`) build items with a nil `Effect`.
They are data already.

**How long a stack item blocks.** Stack items are the most frequent
blocker, but each one blocks for only a short time. An ETB trigger
blocks every capture from its announcement until it resolves. At a
four-seat table that is usually a handful of actions: one pass per
seat, plus any responses. So during a busy turn the restore point is
often a few actions stale, and a deploy then rewinds the table to just
before the trigger. There is still no measurement of this.
#1558 item 4, the soak half of P7, has not landed, and the shutdown
census has only seen idle tables. The ranking above is a static
estimate.

#### What stays census after both tiers

- **`ChoiceResumeFrames`**, by owner decision 4. A target-pick or mode
  prompt for a trigger still blocks while it is open. The item it
  produces is data once it is on the stack.
- **`IntrinsicAbilityCards`.** This counter now fires only for an
  ability stamped onto one instance with no catalog entry behind it.
  Since #521 no production path is known to write one, except a copy
  of such a card (`game/copy.go:515`, `game/spell_copy.go:529`). Its 45
  lines are routes through `Card.ActivatedAbilities` and
  `Card.ManaAbilities` into every zone. They stay, and the counter stays
  as it is.
- **`transient` and `test-only`**, unchanged.

### Decision P8 — tier 3b: replacements and block rules are `ScopedEffect` kinds

P5 proposed two new record types, `ScopedReplacement` and
`ScopedBlockRule`. This amendment proposes new kinds on the existing
`ScopedEffect` instead. There are three reasons:

- **They are the same kind of effect.** Fog's shield, the Whip's
  redirect and Gingerbrute's evasion are each a continuous effect
  created by a resolving spell or ability (CR 611.2), exactly as a
  +3/+3 is. The duration model (ADR 0063), the CR 611.2c pin and the
  CR 400.7 new-object rule apply to them as they apply to the +3/+3.
- **The precedent has already landed.** #1571 put a rules effect that
  is not a characteristic on this record: `addAttackRequirement`,
  together with `AffectedScope` for a set read live instead of locked.
  An attack requirement and a block rule are two sides of the same
  thing.
- **The schema follows from the record, not from the tier.** Anything
  new inside `scopedEffects` is already refused by every v7 binary that
  cannot read it: an unknown kind, scope, duration or field. A new
  top-level key would be dropped without a word, and that is the whole
  reason P4 expected a bump. See P10.

**A kind names its reader, not only a layer.** `modKindSpec` gains a
reader. Every existing kind has `layer`. The new kinds have
`replacement` or `blockRule`. The layer adapter skips any kind that is
not in the layer system. The replacement gather and the block-rule
walk each adapt the records of their own reader into the closures they
already consume, with the same "rebuilt from the record by the running
binary" status as the layer adapter. There is no memo, because these
records are rare and the gather runs once per event. So nothing new is
stored on `Game`.

**The kinds.** Each one covers a measured card. A kind is added only
with the first card that needs it, as P1 already requires.

| Kind | Reader | Reads | Affected | Cards |
|---|---|---|---|---|
| `preventCombatDamage` | replacement | `Player`: `uuid.Nil` is all combat damage, a player is combat damage dealt to that player | `ScopeGame` | Fog, Holy Day, Tangle, Constant Mists, Druid's Deliverance |
| `preventDamage` | replacement | `Amount`, the charge left (must be at least 1); `CombatOnly` | the pinned object, or `ScopeGame` plus `Player` for a player | Mending Hands |
| `exileInsteadOfLeaving` | replacement | nothing | the pinned object | Whip of Erebos, Dregscape Zombie (unearth) |
| `exileInsteadOfGraveyard` | replacement | `Then`: a registered delayed-trigger body scheduled at the next end step for each redirected card | `ScopeYourPermanents` | Cosmic Intervention |
| `cantBeBlockedExceptBy` | blockRule | `Keywords` and `Subtypes`, each any-of, and `Text`, the printed parameter the refusal sentence reads | the pinned attackers | Gingerbrute, Departed Deckhand |
| `limitBlockersPerDefender` | blockRule | `Amount` | `ScopeOpponentsCreatures` (from #1571) | Mirri, Weatherlight Duelist |

New `Mod` fields: `amount`, `combatOnly`, `then` and `text`. New
`AffectedScope` values:

- `ScopeGame`, for an effect that names no object;
- `ScopeYourPermanents`, the permanents the record's `Controller`
  controls, read live.

A replacement effect is not a characteristic, so CR 611.2c does not lock
its set. Cosmic Intervention also saves a permanent that entered after
it resolved, which is what the closure does today.

**Durations.** Every record is `DurationUntilEndOfTurn` except the two
`exileInsteadOfLeaving` users. Those become `IndefiniteDuration`,
pinned to the returned object. The redirect then lasts exactly as long
as that object is on the battlefield, which is the printed "if it would
leave the battlefield". **That fixes #1591**, and the PR says so. Mirri
stays until end of turn, not end of combat, as today (#753).

**The one record that changes.** A `preventDamage` shield can be
partly spent. P1's immutability contract stays, with one clarification:
a record is never mutated in place. When the shield absorbs damage, the
record at that position is replaced by a new record with the lower
`Amount`, and a spent shield is removed. A clone taken before the
damage keeps the old record. So an undo now rewinds the charge, which
fixes the gap `prevention.go` admits in its comment. The PR calls this
out as a behaviour change.

**Replacement IDs must not move.** The old turn slot was append-only
within a turn, so `turnScopedIDBase + index` was stable for as long as a
CR 616 prompt could hold it. `ScopedEffects` shrinks in the middle of a
turn: a sweep, a pin collected, a spent shield. So the gather mints a
scoped record's `ReplacementEffectID` from a per-record sequence number,
never from its position. That means a new record field `seq`, taken
from a game counter at registration. It is additive under P10, and
`ReplacementOptionMetaForEffect` looks the record up by it.

**Retired at compile time:**

- `Game.TurnScopedReplacements` and `Game.TurnScopedBlockRules`;
- `RegisterTurnScopedReplacement`, `RegisterTurnScopedBlockRuleLocked`
  and the two `Clear…Locked` sweeps;
- the `Rule func(BlockScope) game.BlockRule` field on
  `BlockRuleUntilEOT`.

`BlockRuleUntilEOT` becomes `CantBeBlockedThisTurnExceptBy{Target,
Keywords, Subtypes, Text}`. Gingerbrute and Departed Deckhand change
one call each. Every other card-side builder keeps its signature:
`PreventAllCombatDamageThisTurn` (with a new `Player` field that
replaces the batch-38 helper), `PreventNextDamage`,
`ExileInsteadOfLeavingBattlefield` and
`EachOpponentCantBlockWithMoreThanN`. Cosmic Intervention calls a new
`ExileInsteadOfGraveyardThisTurn{Then: returnExiledToOwnersBody}`.
`ContinuationCensus.TurnScopedReplacements` and `.TurnScopedBlockRules`
keep their fields so an old census decodes, and both go into
`retiredCensusCounters`.

**`ErrUnknownEffectKey` covers** each new kind and scope, the four new
`Mod` fields and the `seq` field. It also covers `then` when it names a
body this binary has not registered, checked through the same
`KnownEffectBody` a delayed trigger uses.

### Decision P9 — tier 4: an ability's stack item names its catalog row

**The reference.** Every ability item the engine builds from a catalog
row carries a reference to that row, as data:

```go
// AbilityRef names one row of one CardDef (ADR 0041 P9). It is carried
// in EffectParams, so it is subject to P4's refusal of unknown fields.
type AbilityRef struct {
    Key  string `json:"key"`  // the CatalogAbilityKey at announce: base, copy grants, layered grants, token, emblem
    Slot string `json:"slot"` // "activated" | "triggered"
    Ref  string `json:"ref"`  // ADR 0093 D5's grammar: own:<i>, grant:<bundle>:<i>:<n>
    Name string `json:"name"` // the row's declared Label (activated) or Key (triggered), checked on restore
}
```

- **The key is the composite string**, captured when the item is
  built. It is not recomputed from the source. The source may have left
  the battlefield, and a dies trigger's key comes from last-known
  information, for example a Feign Death grant. `catalogDef` already
  merges any composite key.
- **The ref indexes the catalog list before designation gating**
  (`activeOnly`) and after the keyword triggers that `TriggersForCard`
  puts in front (prowess). So a Class that levels up after the
  announcement does not shift the row. ADR 0093 D5's ref grammar
  extends to triggered rows unchanged.
- **The item names a body as well:** `Body: "catalog/activated"` or
  `"catalog/triggered"`, with the reference in `Params.Ability`. These
  are two new keys in tier 2's registry and ledger. The body is the
  refusal token. Every v7 binary before this tier refuses a file that
  names it (P10). At announce time nothing else changes: the item still
  holds the row's `Effect`, `targetSpec` and `modeSpec` directly. At
  restore, `restoreStackItem` looks the row up once and sets all three
  from it, including each `ModeDoing` bullet. A `TargetsFrom` clause is
  called again with the carried `item.Trigger`.

**Activated abilities need no card edits.** `ActivateCatalogAbility`
already copies the row verbatim. It gains the two stamps and nothing
else, and that covers all 525 files at once.

**Triggered abilities become declarative.** `TriggeredAbility` gains
`Effect func(g *Game, item *StackItem) error`. When a declaration sets
it, the engine builds the item: `NewTriggeredItem(source, Key, Effect)`,
stamped with the reference. `Build` may still exist to fill in data:
a controller override (`WhenYouLoseControlOfThis`), a payload, a label,
or `Params`. But it must leave `item.Effect` nil. If it does not, that
is an `effectKeyFault`: the test binary panics, and production logs it
and keeps the closure, which then counts as census. A value the old
`Build` captured at trigger time moves onto the item:

- the event and its object are already on `item.Trigger` (#1379), so
  Suture Priest's victim is `ctx.Trigger().Object.Controller`;
- a board read made at trigger time goes into `Params` from `Build`
  (`Player`, `Object`, `Amount`, `Cost`, `Name`).

`OnAny` and every constructor on it set `Effect` and drop `Build`.
That covers the constructor-only files, plus `Ward`, `Cascade`,
`Storm` and the `buildDef` keywords, in one small change.

**Engine triggers with no catalog row use tier 2 body keys:**

- `prowess/pump`, `suspend/tick` and `suspend/free-cast`;
- `madness/offer`, reading `Params.Cost`;
- `monarch/draw` and `monarch/crown`, reading `Params.Player`;
- `evoke/sacrifice` and `facedown/ward`;
- `mana-rider/<name>` for the four `WhenManaSpent` cards.

**Reflexive triggers** move from `ReflexiveTrigger.Effect` to
`ReflexiveTrigger.Body` with `Params`. A targeted reflexive trigger
cannot re-derive its clause from a row, so the body registration can
declare the clause beside the function:
`game.ReflexiveBody(key, fn, targets)`. Tier 2 already restores any
item that names a body, so this slice depends on nothing.

**Carried with this tier.** `Game.lastKnownStack` becomes `carried`. Its
card is snapshotted like any card, and its item like a spell item. Its
readers become data in this tier, so dropping it would make a restored
storm miss the copies of countered spells.

**The census folds.** From the first tier-4 slice onwards, an item
whose `Effect` is an unkeyed closure is counted once, in `StackEffects`,
whatever its spec. An item with a reference or a body re-derives its
spec. So `StackTargetSpecs` retires in that slice. `StackEffects`
retires when the last hand-written trigger `Build` is migrated (see the
slicing).

**Restore when the row has changed.** A deploy can edit a card, so the
row at `Ref` may be gone, or it may now be a different ability.
Restore compares `Name` with the row's declared label.

- If they match, the item is restored.
- If they do not match, it looks for the one row in the same slot with
  that name, and uses it if there is exactly one.
- If there is none, it applies the owner's answer to Q3. The
  recommended answer is the #522 posture: restore the game, leave the
  item on the stack with no automatic effect (as a sandbox-announced
  trigger is), flag the source card `AbilitiesLostOnRestore`, and log
  one ERROR line.

**The lint.** A runtime test walks the registry. Every
`TriggeredAbility` must set `Effect`, or its key must be on
`testdata/legacy_trigger_builds.txt`. That list only shrinks, the
`closureClassCeilings` way. Once it is empty,
`NewTriggeredItem(source, label, effect)` loses its `effect`
parameter, which is P3's retirement at compile time.

### Decision P10 — schema: tiers 3b and 4 stay in v7

P4 said tiers 3b and 4 must bump to v8: "An earlier v7 binary would
ignore those fields, and there is no way to make it refuse a single
field." Tiers 1 and 2 have since built exactly that. A v7 binary refuses
a file with an unknown mod kind, scope, duration, record field, mod
field, delayed-trigger or stack-item body, params field, filter or
grant bundle. It keeps the file and abandons the game, as Decision 5
already does for a file that is too new.

**The test for "additive".** A new file may hold state X. Ask what an
older binary does with X:

- If the older binary counted X in its own census, it never wrote a
  file holding X. Dropping X would restore something it knows it cannot
  restore. X must reach it through a channel it refuses, or it needs a
  bump.
- If the older binary drops X from files it writes itself, carrying X
  makes nothing worse for it. X is additive with no refusal needed.

**Tier 3b passes.** Everything new is inside `scopedEffects`: new kinds,
new scopes, new `Mod` fields and a new record field. A v7 binary from
before 3b refuses the file on any of them.

**Tier 4 passes**, because every item it makes restorable carries a
body that no earlier binary registered (`catalog/activated`,
`catalog/triggered` or an engine key). Every earlier v7 binary refuses a
stack-item body it does not know. That has been true since PR 1, and
`params` fields have been refused too since #1582.
`lastKnownStack` falls under the second rule: today's binary drops it
from its own files. **The first tier-4 slice also extends the
unknown-field refusal to the whole stack-item record.** This does not
protect earlier binaries, which is why the body carries the refusal. It
does mean that a later stack-item field is refused by every binary from
this slice onwards.

**What a bump would cost.** After a bump, a rollback refuses **every**
restore point written since the deploy, whatever it holds. With the
refusal channels, a rollback refuses only the games that hold a new
kind of effect or item. The rollback case is the only one that
differs, and the refusal channels are strictly better there.

**Migration: none either way.** No v6 or v7 file can hold a turn-scoped
replacement, a turn-scoped block rule or an ability closure, because
the census kept every one of them out. So older files restore unchanged
under both options, and `corpusMigrations` stays empty. There is no
choice to make between migrating v7 files and refusing them.

**Fixtures.** These are new files under `v7/`, following the
"never touch an existing file" rule, each made by a real card through
`TestWriteSnapshotCorpus`:

- tier 3b: `fog.json`, `mending_hands_partial.json` (a shield with charge
  left), `whip_redirect.json`, `cosmic_intervention.json`,
  `gingerbrute.json` and `mirri_limit.json`;
- tier 4: `activated_on_stack.json` (a targeted ping),
  `etb_trigger_on_stack.json`, `granted_dies_trigger.json` (Feign Death,
  keyed through LKI), `token_trigger.json`, `emblem_trigger.json`,
  `modal_trigger.json`, `reflexive_trigger.json`,
  `prowess_on_stack.json` and `storm_after_counter.json` (with
  `lastKnownStack`).

The shape file `v7.txt` records the new paths through `-update-shape`.

**If the owner prefers v8 (Q1 option B).** The same design ships under
v8. The difference is `SnapshotSchemaVersion = 8`, a frozen `v7.txt`, a
new `v8.txt` and a `v8/` directory holding the fixtures above.
`corpusMigrations` is still empty.

### Decision P11 — the ratchet when a counter retires into another class

A field reachable by several routes carries the most restrictive class
among them. When a tier deletes the route that set its class, the field
falls to its next most restrictive route. That is usually the resume
frames. For example, `ReplacementEffect.Replace` is reached through
`BuiltinReplacements` (rebuilt), through `testReplacements` and through
a paused CR 616 prompt's `replacementResume.applicable.effect`.

That moves lines into `census:ChoiceResumeFrames`, and that class's
ceiling has to rise by the number that moved. The ratchet file says a
ceiling is never raised "to make a new route pass". This is not a new
route. So the rule gains one sentence: *a retirement may move a line to
the class of its remaining route, in the same PR, and the PR lists the
lines and the new ceiling*.

Worked out for this design:

- **3b, replacements.** `Game.TurnScopedReplacements` goes. The five
  `ReplacementEffect.*` lines, the two `CopySelector.*` lines and the
  two `EntryHandReveal.*` lines move to `ChoiceResumeFrames`, which goes
  from 98 to 107. The `TurnScopedReplacements` ceiling is deleted.
- **3b, block rules.** `Game.TurnScopedBlockRules` and the three
  `BlockRule.*` lines go. No other route reaches `BlockRule`, so its
  ceiling is deleted.
- **Tier 4, first slice.** The 11 `StackTargetSpecs` lines move to
  `ChoiceResumeFrames` (through `pickTargetFrame.spec`,
  `modePickFrame.ability.Modes` and `resolving.item`), or to `keyed`
  where a stack route was the only one.
- **Tier 4, last slice.** When `StackEffects` retires, its six lines
  move in the same way. `Game.StackMeta` and `Game.PendingTriggers`
  become `keyed`. `lastKnownSpell.card` becomes
  `census:IntrinsicAbilityCards`, through `Card`.

After tier 4, the only census classes left are `ChoiceResumeFrames`,
at about 120 lines, and `IntrinsicAbilityCards`.

### Slicing

Every slice is an ordinary feature PR against `develop` that updates
AGENTS.md where a recipe changes. Sizes include tests.

| Slice | What | Depends on | Model | Size |
|---|---|---|---|---|
| **M** | #1558 item 4: the catalog soak captures after every action and reports census kinds and the longest stale run. This is P7's measurement. It is not a gate, but it tells us whether the tier-4 tail is worth hurrying | none | Sonnet | about 250 lines |
| **3b-1** | Non-layer readers, `ScopeGame`, `ScopeYourPermanents`, `seq`, the four replacement kinds, copy-on-write shields, and the nine replacement cards. Retires `TurnScopedReplacements` and fixes #1591 | none | Opus (replacement IDs, CR 616 interaction) | about 900 lines |
| **3b-2** | The two block-rule kinds, three cards. Retires `TurnScopedBlockRules` | 3b-1's reader field, or rebased onto it | Sonnet | about 350 lines |
| **4-0** | Engine triggers and reflexive triggers onto tier 2 body keys, including `ReflexiveBody` with a target clause | none (tier 2 already restores keyed items) | Sonnet | about 500 lines, 10 card files |
| **4-1** | `AbilityRef`, the `catalog/*` bodies, restore-time lookup with the name check and the Q3 policy, the stack-item unknown-field refusal, `lastKnownStack` carried, and the census fold. Activated abilities are stamped. Retires `StackTargetSpecs` | none | Opus | about 1,000 lines |
| **4-2** | Declarative `TriggeredAbility.Effect`, the engine default build, the constructors (`OnAny`, `Ward`, `Cascade`, `Storm`, `buildDef` keywords), the `TargetsFrom` audit, the legacy-`Build` allowlist and its test | 4-1 | Opus (the harvester) | about 700 lines |
| **4-3 … 4-10** | The hand-written `Build` tail, in batches of about 40 files: the 149 source-only files first, then the event readers, then the board readers. Mechanical. Each batch shrinks the allowlist | 4-2 | Sonnet, several in parallel in separate worktrees | 8 PRs, about 300 to 500 lines each |
| **4-final** | Retires `StackEffects`, deletes the `effect` parameter of `NewTriggeredItem`, moves the P11 lines | the tail | Sonnet | about 150 lines |

**What can run at once.** M, 3b-1, 4-0 and 4-1 touch different code.
They share only `closure_fields.txt`, `closureClassCeilings`, the
effect-key ledger and the `checkEffectKeys` lists. Those conflicts are
small, and whichever lands second rebases. Tail batches conflict only in
the allowlist file, whose lines are sorted, one per trigger key.

### Risks, and what stays out of scope

- **A catalog edit between write and restore** can make a stack item's
  row a different ability. The name check catches a rename or a
  reorder. It cannot catch an ability whose label stayed the same while
  its effect changed. That is the same exposure as a spell, which
  restores by oracle ID today, and it is accepted for the same reason:
  the new binary's meaning wins across a deploy (P1).
- **`TargetsFrom` is called again at restore**, not at announce. The
  restored board is the board as it was captured, and that can differ
  from the board when the trigger was announced. A clause that reads
  only the trigger context is exact. The 4-2 audit keeps any of the 18
  that read the board on the allowlist.
- **Hand-written `Build`s that set more than `Params` can say.** If a
  batch finds one, it stays on the allowlist, gets a `Params` field of an
  allowed type, or goes through the P1 `catalog` escape hatch. The
  hatch is still unbuilt, and it lands with its first card.
- **A shield or redirect under a CR 616 prompt at capture time** is
  covered by the prompt's frame, which is `ChoiceResumeFrames`. It does
  not break the record.
- **Out of scope**, as before: resume frames (owner decision 4), the
  undo stack (ADR 0044 decision 6), the P1 `catalog` escape hatch
  (until a card needs it), extra combats for Mirri's "this combat"
  (#753), and `IntrinsicAbilityCards`, which is rare and stays censused.

### Questions for the owner

These are the only choices that change the design. Everything else
above follows from the code.

**Q1. Schema: stay in v7, or bump to v8 as decided on 2026-09-24?**
- *A. Stay in v7.* Everything new rides a channel every earlier v7
  binary already refuses (P10). A rollback abandons only the games that
  hold a new effect or item, and keeps their files.
- *B. Bump to v8* (owner decision 6). The design is the same. A rollback
  across the bump refuses every restore point written since the deploy.
  It needs one more shape file and fixture directory. It protects
  nothing that A does not.
- *C. v7 for 3b, v8 for 4.*

*Recommendation: A.* The case for v8 was that a field could not be
refused on its own. Tiers 1 and 2 built exactly that refusal, and P10
shows that every tier-3b and tier-4 addition is covered by it. A has
the smaller rollback cost and one fewer version.

**Q2. How does a stack item name its ability?**
- *A. The ref alone* (`own:3`). This is cheap. But a deploy that
  reorders a card's abilities silently points the item at the wrong
  one.
- *B. The label alone.* This survives a reorder. But a wording fix
  breaks it, and two instances of a granted ability share one label.
- *C. The ref, checked against the declared label.* If they do not
  match, restore uses the single row with that label, and otherwise
  applies Q3's policy.

*Recommendation: C.* It costs one string per item and turns both
failure modes into a visible outcome instead of a silently wrong one.

**Q3. What happens to a stack ability the new binary cannot find?**
- *A. Abandon the game*, as Decision 5 does for an unreadable file.
- *B. Restore the game.* The item stays on the stack with no automatic
  effect, so the players resolve it by hand as they would a
  sandbox-announced trigger. The card is flagged
  `AbilitiesLostOnRestore`, and one ERROR line is logged.
- *C. Drop the item.*

*Recommendation: B.* It is the #522 decision applied to the stack: the
binary reads the file correctly, and one ability it cannot automate
stays playable by hand. C loses a trigger without anyone seeing it. A
costs the whole table for one card.

**Q4. When is #1497 done: after 4-2, or when `StackEffects` retires?**
- *A. When `StackEffects` retires.* The hand-written `Build` tail, about
  330 closures in 8 mechanical batches, is part of the sprint.
- *B. After 4-2.* The tail becomes background card work like the #583
  promotion backlog, and `StackEffects` stays a live counter until it
  is finished.
- *C. Allow the tail to stay censused permanently.*

*Recommendation: A.* The batches are mechanical and run in parallel.
And a trigger that still blocks is one of the most frequent blockers
there is, so leaving it blocking wastes most of tier 4's value. C would
leave a counter that never falls, with no reason written down.

**Q5. What order?**
- *A. In parallel.* M, 3b-1, 4-0 and 4-1 start together. Then 3b-2 and
  4-2. Then the tail.
- *B. 3b before 4*, as in owner decision 3.
- *C. 4 before 3b*, because tier 4 is where the frequency is.

*Recommendation: A.* 3b is twelve cards, and its only overlap with
tier 4 is the ratchet file. Waiting on it delays the high-frequency tier
for no reason. M runs alongside rather than gating anything, because
both tiers are planned in full whatever it shows.

**Behaviour changes the PRs will call out** (not questions):

- #1591: the Whip's and unearth's redirect lasts while the returned
  object remains, not until cleanup.
- An undo now rewinds a prevention shield's spent charge.
- Cosmic Intervention and Mirri behave exactly as they do today.

### Owner answers (2026-09-25)

The owner took the recommendation on all five questions.

1. **Q1 — A. Stay in v7.** Decision P10 stands and owner decision 6's
   second half ("v8 comes with tiers 3b and 4") is withdrawn. No shape
   file or fixture directory is added for a version bump. New fixtures
   for 3b and 4 are written into `v7/` beside the existing ones, under
   the "never touch an existing file" rule.
2. **Q2 — C. The ref, checked against the declared label.** A mismatch
   falls back to the single row with that label; otherwise answer 3
   applies.
3. **Q3 — B. Restore the game.** An ability the new binary cannot find
   stays on the stack with no automatic effect, to be resolved by hand.
   The card is flagged `AbilitiesLostOnRestore` and one ERROR line is
   logged.
4. **Q4 — A. #1497 closes when `StackEffects` retires.** The `Build`
   tail (slices 4-3 to 4-10) is part of the sprint.
5. **Q5 — A. In parallel.** M, 3b-1, 4-0 and 4-1 start together; then
   3b-2 and 4-2; then the tail and 4-final.

### Implementation notes (tier 4-1)

Slice 4-1 of P9 (`AbilityRef`, the catalog body, the restore lookup
with Q2 and Q3, the stack-item field refusal, `lastKnownStack` carried,
and the census fold). Where the code differs from the design above, the
code wins, and each difference is listed here.

- **Only activated abilities are stamped, so only `catalog/activated`
  is registered.** `ActivateCatalogAbility` puts
  `Body: "catalog/activated"` and `Params.Ability` on the item. It
  names the row while the source is still the object that has it,
  before any cost moves it. `catalog/triggered` is **not** registered
  in this slice. A 4-1 binary cannot rebuild a trigger's `Effect`,
  because the row has no `Effect` until 4-2. So it has to refuse a 4-2
  file that names the key, and it does, because the key is unknown.
  4-2 registers it.
- **`AbilityRef.Key` is the key whose list the ref indexes, not the
  full `CatalogAbilityKey`.** For `own:<i>` it is the object's
  `CatalogKey`, which already includes any copy grants. For
  `grant:<bundle>:<i>:<n>` it is the bundle's `GrantKey`. The composite
  key cannot be split back apart: a copy grant and a layer-6 grant are
  both `|grant:<x>`, and `own:<i>` counts only the own list. A grant
  ref whose bundle is not its key is refused.
- **A stamp is only written if the lookup works.** The announce path
  checks that the running catalog returns the same row under the ref.
  If it does not, the item is not stamped. That covers an
  instance-carried ability (`census:IntrinsicAbilityCards`) and a
  catalog stub that does not return rows. An unstamped item stays an
  unkeyed closure, and the census counts it.
- **Q3's manual item has no body and no ref.** A row that cannot be
  found leaves the item on the stack with its announcement: label,
  targets, modes, X and payment. Its `Effect`, target clause, mode
  clause, body and ref are cleared, so it resolves as a
  sandbox-announced item does. Its source card is flagged
  `AbilitiesLostOnRestore`, and that flag is carried from then on. The
  next capture is an ordinary restore point, so a later boot does not
  report the item again. The report is
  `GameSnapshot.LostStackAbilities()`. It makes the same lookup restore
  makes, and reads only the snapshot and the catalog, as
  `AbilityShortfalls` does. The boot path (`ws/persist.go`) logs one
  ERROR line for each lost item, naming the game, the card, the label
  and the ref. The summary counts these items in
  `cards_with_lost_abilities`.
- **Q2 applies inside the list the ref names.** For a moved row, the
  single row with the declared label in the same list is used, and the
  ref is rewritten to point at it. A grant ref keeps its bundle and
  instance. An empty label never uses the fallback.
- **The whole-record refusal applies to the current schema only.**
  Unknown keys on a stack item, on its `params.ability`, on a
  `lastKnownStack` entry and on that entry's item are refused
  (`ErrUnknownEffectKey`) when the file is v7. In an older file an
  unknown key can only be one that a later bump removed, and that is the
  bump's migration to handle. No fixture holds such a key. The
  existing `params` checks still apply to every schema.
- **More refusals** (`unknownAbilityRef`):
  - a `catalog/activated` item with no ref;
  - a ref with an unknown slot, including `triggered`, which 4-2 adds;
  - a ref outside the grammar;
  - a ref under a body that does not read one;
  - any `params.ability` on a delayed trigger.

  No binary writes any of these, so each one means a newer file.
- **The census fold.** An item is counted once, in `StackEffects`, in
  either of two cases:
  - its `Effect` is an unkeyed closure;
  - it holds a target or mode clause, and has neither a spell's oracle
    ID nor a catalog ref to rebuild the clause from.

  The second case includes a tier-2 keyed item that carries a clause.
  None does today. 4-0's reflexive bodies will, until restore can
  rebuild their clause. `StackTargetSpecs` is on
  `retiredCensusCounters`. The field is kept so that a census written
  by an older binary still decodes as not restorable.
- **P11, as applied.** The eleven `census:StackTargetSpecs` lines
  moved to `census:ChoiceResumeFrames`, the class of the route that is
  left to them (a resume frame reaches every one):
  - `ModeOption.Effect`, `ModeOption.Targets`, `ModeSpec.Options`;
  - `StackItem.modeSpec`, `StackItem.targetSpec`;
  - `TargetDifference.Key`;
  - `TargetSpec.AbilityOK`, `.CardOK`, `.Different`, `.PlayerOK`,
    `.Rest`.

  Its ceiling went from 98 to 109. The rule now appears in
  `closure_fields_test.go` and in the header of `closure_fields.txt`.
  `StackEffects` stays at 6, because `lastKnownStack`'s three lines are
  still reached through `StackItem.Effect`.
- **`lastKnownStack` is carried** as
  `GameSnapshot.LastKnownStack`: a list sorted by card ID, where each
  entry is a card mirror and a stack-item mirror. The item's oracle ID
  is read from the recorded card, because a countered copy is in no
  zone. The census sees each entry as it sees a live item and card, and
  a real spell holds nothing it counts. One engine fix came with it.
  `createSpellCopyLocked` now creates `StackMeta` when it is nil. A game
  restored with an empty stack has no map, and a copy made from a
  carried record can be the first item put on it.
- **Verifying P10.** A test removes `catalog/activated` from the
  registry, standing in for a binary from before this slice, and
  confirms the file is refused. Another removes `lastKnownStack` from a
  file, standing in for what an older binary writes, and confirms that
  file still restores.
- **Fixtures.** Three new files in `v7/`, each made by real cards:
  - `activated_on_stack.json`: Goblin Bombardment's ping, `own:0`;
  - `granted_activated_on_stack.json`: Squirrel Nest's granted ability,
    `grant:squirrel-nest/make-a-squirrel:0:0`;
  - `countered_spell_lki.json`: a Lightning Bolt countered by
    Counterspell.

  No existing fixture changed. The triggered fixtures P10 lists
  (`etb_trigger_on_stack`, `granted_dies_trigger`, `token_trigger`,
  `emblem_trigger`, `modal_trigger`) need 4-2's declarative `Effect`.
  `reflexive_trigger` and `prowess_on_stack` need 4-0's body keys.
  `storm_after_counter` needs storm's trigger to be data. None of these
  can be a restore point in this slice.

### Implementation notes (tier 3b-1)

What the replacement half of tier 3b settled where the code and P8
differ or P8 left it open. Everything else is as P8 says.

- **The four kinds are declared, and nothing else.** `preventCombatDamage`,
  `preventDamage`, `exileInsteadOfLeaving` and `exileInsteadOfGraveyard`
  carry `reader: readerReplacement`. The layer adapter skips them. The
  replacement gather adapts each live record in
  `game/scoped_replacements.go`. The block-rule reader, `Mod.Text` and the
  two block-rule kinds are left to 3b-2, because nothing here reads them.
- **The engine owns the writers.** There are four `*ForEffect` functions,
  one per kind, and no exported mod constructor for a replacement kind. A
  `ScopedEffectFor` cannot write one by accident. The card-side builders
  keep their names. `PreventAllCombatDamageThisTurn` gains `Player`, which
  replaces the batch-38 helper. Cosmic Intervention calls
  `ExileInsteadOfGraveyardThisTurn{Then}`.
- **Every adapted closure reads its record back by `Seq` at call time.**
  So an effect held by an open CR 616 prompt acts on the registry as it is
  when the prompt is answered. A spent shield, a record an undo removed and
  a redirect whose object has gone all answer "does not apply". None of
  them acts on a stale copy.
- **`Seq` goes only on a record with a non-layer mod.** A layer-only record
  is never named. Keeping its `Seq` at zero keeps it byte-identical to what
  an earlier v7 binary writes and reads, so a Giant Growth does not become
  a rollback refusal. The ID is `scopedReplacementIDBase + Seq×8 + mod
  index`. It sits in the range the turn-scoped registry used.
- **The counter is derived, not carried.** `Game.scopedEffectSeq` is
  cloned with the game. Restore sets it to the largest `Seq` among the
  restored records. Uniqueness needs nothing more, because no prompt
  survives a restore. So the snapshot gains no top-level key. The drift
  test classifies the field `rebuilt`.
- **A shield on a permanent also pins its duration** (`PinnedTo`). The
  record is then swept with its object, like the Whip's redirect. The
  affected set is pinned as P8 says, so a flicker ends the shield
  (CR 400.7).
- **The uncharged "prevent that damage" form is not built.** Before this
  change, `PreventNextDamage` with `Amount: 0` meant "the next damage
  event, whole". No catalogued card prints it, and P8 gives
  `preventDamage` a charge of at least 1. An `Amount` below 1 now
  registers nothing. The one test that used it
  (`TestCreepingBloodsuckerGainsNothingWhenEveryOpponentIsFogged`) uses a
  100-point shield instead.
- **Cosmic Intervention's per-card return is labelled from the record.**
  The label is `SourceName + " — return the exiled permanent"`, the string
  the closure used to hard-code. The body is `Mod.Then`, checked through
  `KnownEffectBody` at registration and at restore.
- **The Whip's and unearth's record has no `Source`.**
  `ExileInsteadOfLeavingBattlefield(g, cardID, controller, label)` keeps
  its signature, as P8 asks, and that signature names no source. The label
  carries the attribution.
- **Ratchet (P11).** The `Game.TurnScopedReplacements` line is deleted.
  Nine lines move to `census:ChoiceResumeFrames`, through
  `replacementResume.applicable.effect`:
  - `ReplacementEffect.AppliesTo`, `.Controller`, `.CopySelector`,
    `.EntryHandReveal` and `.Replace`;
  - `CopySelector.Candidates` and `.Except`;
  - `EntryHandReveal.Matches` and `.Then`.

  The `ChoiceResumeFrames` ceiling goes from 98 to 107. The
  `TurnScopedReplacements` ceiling is deleted and the counter is added to
  `retiredCensusCounters`. `ContinuationCensus.TurnScopedReplacements`
  keeps its field so an old census still decodes.
- **Two caveats were stale, and both are cleared.**
  - Whip of Erebos's "Stifle the exile" caveat was #1591. It is fixed by
    the indefinite pin (`TestWhipRedirectOutlivesACounteredExile`).
  - Dregscape Zombie's "a bounce goes to hand" caveat was already stale:
    #539 routed the bounce through the exit primitive
    (`TestUnearthedCreatureBouncedGoesToExile`).

  Both cards are `CompletenessFull`.
- **Fixtures.** Four new files are written into `v7/`: `fog.json`,
  `mending_hands_partial.json`, `whip_redirect.json` and
  `cosmic_intervention.json`. No existing fixture changed. `v7.txt` records
  `mods[].amount`, `.combatOnly`, `.then` and `seq`.

### Implementation notes (tier 3b-2)

What the block-rule half of tier 3b settled where the code and P8
differ, or P8 left it open. Everything else is as P8 says. This slice
depends on 3b-1's reader field, `Mod.Amount` and the `seq` machinery,
and rebases onto it.

- **The two kinds are declared, and nothing else new in the reader
  system.** `cantBeBlockedExceptBy` and `limitBlockersPerDefender`
  carry `reader: readerBlockRule`, a THIRD `modReader` alongside
  `readerLayer` and `readerReplacement` (P8 speaks of "the layer pass,
  or the replacement gather" as if there were only two readers; a
  block rule is its own kind of non-layer reader, not a variety of
  replacement). The layer adapter skips both, exactly as it skips the
  replacement kinds.
- **No `*ForEffect` writer functions, and no validation gap either.**
  Unlike the four replacement kinds (which each got a dedicated
  `*ForEffect` function because they have engine-decided semantics —
  the shield's minimum charge, the delayed-trigger body check), the
  two block-rule kinds are plain layer-mod-style constructors,
  `CantBeBlockedExceptByMod` and `LimitBlockersPerDefenderMod`
  (`game/scoped_block_rules.go`), registered through the EXISTING
  `RegisterScopedEffectForEffect` / `RegisterScopedRuleEffectForEffect`
  entry points — the same ones `AddAttackRequirementMod` (#1571) uses.
  Validation still happens once, centrally: `blockRuleModProblem`
  mirrors `replacementModProblem` and is called from
  `appendScopedEffectLocked` alongside it, so a bad parameter panics at
  registration whichever entry point reached it.
- **The block-rule closures are rebuilt fresh on every walk and never
  read back by Seq.** P8's language ("rebuilt from the record by the
  running binary" — the same status the replacement gather has)
  reads as implying the replacement kinds' by-Seq indirection carries
  over. It does not, and the reason is a real difference between the
  two consumers rather than an oversight: a replacement effect's
  closures can be handed to a CR 616 ordering prompt and called again
  after the registry has shrunk underneath them (a sweep, a spent
  shield), so `scopedReplacementEffect` reads the record back by `Seq`
  at call time. A block declaration never pauses on a prompt — the
  whole check (`blockRuleRefusalLocked`, `blockerBoundsLocked`,
  `blockLimitRefusalLocked`) runs synchronously inside one locked call,
  and the `BlockRule` values `forEachScopedBlockRuleLocked` hands to it
  are used and discarded before that call returns. So
  `blockRuleFromScopedMod` captures `ScopedEffect` and `Mod` BY VALUE
  (both are immutable once registered) instead of closing over `(seq,
  mod, kind)` and re-reading the registry. Every such record still
  takes a `Seq`, because `appendScopedEffectLocked` stamps one on any
  non-layer mod regardless of reader — nothing here reads it back, and
  nothing needs to.
- **`forEachScopedBlockRuleLocked` fully replaces
  `forEachTurnScopedBlockRuleLocked`,** rather than sitting beside it
  as a third walk. `blockRuleRefusalLocked`, `blockerBoundsLocked` and
  `blockLimitRefusalLocked` each call `forEachBlockRuleLocked` (the
  battlefield walk, unchanged) and then the new walk in its place —
  same two-step order P8's "walked after the battlefield" preserves.
  `CatalogBlockRules` (a permanent's own printed rule) is untouched by
  this tier; only the until-end-of-turn twin moved.
- **The engine owns the "allowed" predicate.**
  `blockRuleAllowedPredicate(keywords, subtypes)` mirrors
  `effects.HasKeyword` / `effects.OfCreatureType` (`cards/effects/targets.go`)
  exactly — a keyword any-of via `game.HasKeyword`, a subtype any-of
  via `Card.HasSubtype` guarded by `IsCreature()` — so a changeling
  still passes for a named subtype and a granted keyword still counts.
  Duplicating the predicate rather than sharing code with
  `cards/effects` is deliberate: `internal/game` cannot import
  `cards/effects` (the dependency runs the other way), and the two
  copies are held in step by the integration tests (Gingerbrute's
  haste, Departed Deckhand's Spirits and changeling case), not by a
  shared function.
- **`BlockRuleUntilEOT`'s generality is retired along with the
  struct.** P8 gives the retired shape as `CantBeBlockedThisTurnExceptBy{Target,
  Keywords, Subtypes, Text}` — four fields, no `Match` predicate and no
  `Rule func(scope) game.BlockRule` closure. The old struct's `Match`
  field and arbitrary `Rule` closure had exactly zero production
  callers (only `Target` was ever used, by Gingerbrute and Departed
  Deckhand), so nothing is lost: the new struct is `ScopedEffectFor`
  under the hood, reusing the existing Target/Match → snapshot →
  `RegisterScopedEffectForEffect` pipeline `BoostUntilEOT` and
  `GrantKeywordUntilEOT` already share, rather than inventing a second
  one. `eotAffected.appliesTo()`, which existed only for
  `BlockRuleUntilEOT`'s old closure-based `Rule`, is deleted as dead
  code — every continuous effect (layer, replacement or block-rule
  mod) now pins its affected set as data, with no closure predicate
  left in the tier 3b family.
- **`EachOpponentCantBlockWithMoreThanN.Apply` drops straight to
  `RegisterScopedRuleEffectForEffect`** with `game.ScopeOpponentsCreatures`
  and a `LimitBlockersPerDefenderMod`, the same shape
  `OpponentsCreaturesAttackIfAble` (#1571) already uses for its own
  `ScopeOpponentsCreatures` record. No new card-facing wrapper type was
  needed for this one.
- **Ratchet (P11).** `Game.TurnScopedBlockRules` and its three
  `BlockRule.{Pair,Count,Limit}` lines are DELETED outright, not moved.
  P11's worked example ("No other route reaches `BlockRule`, so its
  ceiling is deleted") holds exactly: unlike the replacement kinds'
  `ReplacementEffect`, `CopySelector` and `EntryHandReveal` types (each
  also reachable through a paused CR 616 prompt's resume frame, so
  their lines migrated to `census:ChoiceResumeFrames` when the turn-scoped
  registry that set their class went away), nothing else in the tree
  holds a `BlockRule` value across an action boundary — a block
  declaration has no resume frame to fall back to. All four lines
  (`BlockRule.Count`, `BlockRule.Limit`, `BlockRule.Pair`,
  `Game.TurnScopedBlockRules`) are removed from
  `testdata/closure_fields.txt`, and `census:TurnScopedBlockRules` is
  deleted from `closureClassCeilings` rather than lowered to 0 (a
  ceiling of 0 and no entry are equivalent to the ratchet test, but the
  ADR's own worked example says "deleted", so the code follows it).
  `TurnScopedBlockRules` joins `retiredCensusCounters`.
  `ContinuationCensus.TurnScopedBlockRules` keeps its field so an old
  census still decodes.
- **Fixtures.** Two new files are written into `v7/`: `gingerbrute.json`
  (a real `{1}` activation: one `cantBeBlockedExceptBy` record with
  `Keywords: ["haste"]`, pinned to the Gingerbrute object) and
  `mirri_limit.json` (a real attack trigger: one
  `limitBlockersPerDefender` record, `Scope: "opponentsCreatures"`).
  No existing fixture changed. `v7.txt` records the one new field,
  `mods[].text`; `Mod.Keywords` and `Mod.Subtypes` and `Duration`
  already had shape entries from tier 3a and 3b-1.
