# ADR 0054 — Dice rolls and coin flips from effects

**Status:** Accepted · 2026-09-17 · unscheduled (card-coverage audit, wave 1) · tracked on [#744](https://github.com/krakenhavoc/cmd_and_ctrl/issues/744)
**Numbering:** 0052 is reserved for the emblems ADR
([#623](https://github.com/krakenhavoc/cmd_and_ctrl/issues/623)) and is
skipped here, as ADR 0053 also skips it. On 2026-09-17 every remote
branch (191 of them, after `git fetch origin '+refs/heads/*:refs/remotes/origin/*'`)
was checked with the AGENTS.md §4 loop. No branch has a
`docs/decisions/0052-*`, `0054-*` or higher file. The highest is 0053.
**Amends:** [ADR 0041](0041-game-persistence.md) Decision 3 (what the
snapshot carries for the RNG) and the undo contract in
`server/internal/game/clone.go` ("Undo therefore does NOT rewind the
random stream").
**Builds on:** [ADR 0041](0041-game-persistence.md) (snapshots),
[ADR 0044](0044-surviving-a-deploy.md) (restore points, fixtures),
[ADR 0033](0033-ai-bot-seat.md) (the legal-move enumerator, bot tiers,
the public log), and the chained-choice queue
(`server/internal/game/chained_choice.go`).
**Sibling:** [#745](https://github.com/krakenhavoc/cmd_and_ctrl/issues/745)
builds the public random-order bottom. This ADR uses it and does not
redesign it ([Decision 8](#decision-8--the-random-order-bottom-is-745s-and-draws-from-a-stream)).
**Decided by the owner (2026-09-17):**

1. **Undo rewinds randomness.** The RNG state is part of the undo clone
   and the snapshot, so undo-then-redo gives the same result. This
   applies to the random uses that already exist too (shuffles, random
   discard, random-order bottom), so undo can't be used to fish for a
   different outcome ([Decision 4](#decision-4--undo-rewinds-randomness-owner)).
2. **A won/lost coin flip prompts for a call.** The flipping player calls
   heads or tails (CR 705.2) through a pending choice. A bot calls at
   random. A flip nobody wins or loses gets no prompt
   ([Decision 6](#decision-6--a-wonlost-flip-prompts-for-a-call-owner)).

3. **A roll or flip shows a log line and a reveal-strip cue**, with no
   animation (sub-PR 4).
4. **One heads/tails call covers every coin of one instruction** (Yusri);
   each coin still draws its own won/lost result. "Flip until you lose"
   still prompts once per flip.
5. **The first card wave includes Fiery Gambit, Game of Chaos and Yusri**
   alongside the cards the seam fully unblocks.
6. **Game of Chaos: the controller always flips.** The other player only
   decides whether to flip again, after a flip the controller lost
   ([Decision 6](#decision-6--a-wonlost-flip-prompts-for-a-call-owner)).

Items 3–5 were answered on 2026-09-17; the options considered are kept in
[Owner decisions, second round](#owner-decisions-second-round). Item 6
was confirmed by the owner on 2026-09-17, as Decision 6 already read.

Line references are to `origin/develop` at `bcac391`.

## Context

### The RNG exists, but no card effect can reach it

The game has one seeded, persisted random source:

- `Game.rng` (`server/internal/game/game.go:279`), a `*rand.Rand`, and
  `Game.rngState` (`:302`), the `*rand.PCG` behind it. `start` sets both
  (`game.go:581-582`). `Start(nil)`, which production uses
  (`lobby/lobby.go:701`, `cmd/server/main.go:662`), mints a PCG seeded
  from crypto/rand. A test that passes its own `*rand.Rand` gets a source
  whose position can't be read back.
- `snapshotRNG` (`snapshot.go:666`) writes the PCG's marshalled state.
  A caller-supplied source is recorded as `"external"`, and it sets
  `ContinuationCensus.UnpersistableRNG` (`:472-474`), which makes the
  game not restorable (`:492`). `restoreRNG` (`:1106`) continues the
  stream or mints a fresh one.

Card effects use the RNG in only a few places, and none of them can roll or flip:

| call site | what it does |
|---|---|
| `game.go:597` | opening library shuffle |
| `mutations.go:5040` (`Mulligan`), `:5129` (`ShuffleLibrary`) | mulligan and sandbox shuffles |
| `effect_api.go:1523` (`finishSearchLocked`) | shuffle after a search |
| `effect_api.go:1887` (`ShuffleLibraryForEffect`) | "shuffle your library" |
| `effect_api.go:329` (`DiscardRandomForEffect`) | `g.rng.IntN` at `:342`, **index 0 when `g.rng` is nil** (`:341`) |
| `cascade.go:247` (`bottomInRandomOrderLocked`) | private, exile only, falls back to the global `math/rand/v2` source when `g.rng` is nil (`:253-257`) |

No `RollDie`, `FlipCoin`, `EventRollDie` or `EventFlipCoin` exists
anywhere in `server/`. The only workaround is
`b27PseudoRandomIndex` (`cards/effects/batch27_helpers.go:253`), which
hashes the newest event's sequence number. Urza's Bauble uses it
(`batch27_helpers.go:310`) and has a caveat because of it
(`urzas_bauble.go:22-27`, caveat at `:33`). `DiscardCards`
(`cards/effects/primitives.go:87-106`) also sends every discard through
the random discard, which is a declared simplification.

### Undo does not rewind the RNG today

`cloneLocked` shares the RNG pointers with the clone (`clone.go:41-50`),
and `RestoreFrom` copies them back (`clone.go:517-518`). The comment
calls this "deliberate and long-standing". `Room.apply` clones before
every action (`ws/room.go:166`, pushed at `:170`), `applyBundle` does
the same (`:238`, `:253`), and `Room.undo` (`:363`) restores that clone.
The only limits on undo are the caller gate and the per-turn
`UndosRemaining` budget (default 1).

The result is that undoing a shuffle or a random discard and doing it
again already gives a **different** result. With dice, that becomes "undo
until you roll a 20". The owner decided to rewind instead (Decision 4).

A single stream copied into the clone is not enough to stop fishing,
though. The player can undo, spend one random draw somewhere else, and
then redo:

> Hoarding Ogre attacks and rolls a 3. Undo the attack. Tap Vexing
> Puzzlebox for mana (a d20 roll, as a mana ability, at will). Attack
> again. The Ogre now rolls the next number in the stream.

Cracking a fetchland (a shuffle) does the same thing in a
Commander deck. Decisions 2 and 4 are built to close this.

### What the rules ask for

- **CR 705.1-705.3** (flipping a coin). CR 705.2: an effect that cares
  only about heads or tails has no winner or loser, and for every other
  flip "the player that flips the coin calls 'heads' or 'tails'". CR 705.3:
  an effect can fix a flip's result. Krark's Thumb ("flip two coins
  and ignore one") is the nearest card.
- **CR 706.1-706.8** (rolling a die). A dN has N equally likely results
  from 1 to N (706.1a). The *natural result* is the number before
  modifiers (706.2). Rerolls come before +/- modifiers (706.2b).
  Results tables (706.3) and "roll again" (706.3c). Doubles (706.5,
  Celebr-8000). An ignored roll never happened, so it triggers nothing
  (706.6, Pixie Guide, Barbarian Class). The planar die triggers
  "whenever you roll" but has no numeric result (706.7). Stored results
  (706.8, Centaur of Attention).
- **"At random" and "in a random order"** have no numbered section. They
  show up in the keyword definitions (discover 701.57a, hideaway
  702.75a, cascade 702.85a) and in card text.

### The cards

The audit on #744 counts 51 cards where this is the only core blocker
and 62 where it is one of the core blockers. The oracle text in the
Scryfall dump falls into these shapes:

| shape | cards |
|---|---|
| roll a d20, use the number | Ancient Copper / Gold / Silver Dragon, Ancient Brass Dragon (also #636) |
| roll into a results table | Hoarding Ogre, Wand of Wonder, Delina, Wild Mage ("you may roll again") |
| roll several dice, per-die effect | Clown Car (X d6), Celebr-8000 (2d6, doubles) |
| roll several dice, choose one result | Reckless Endeavor (2d12) |
| roll plus a modifier | Wyll's Reversal (706.2) |
| "whenever you roll one or more dice" | Vexing Puzzlebox, Barbarian Class level 2 |
| won/lost flip | Mana Crypt, Krark, the Thumbless, Stitch in Time, Goblin Archaeologist, Frenetic Efreet, The Gold Saucer, Chaotic Strike, Mogg Assassin |
| several flips | Yusri, Fortune's Flame (1-5 coins, each won or lost) |
| flip until you lose | Fiery Gambit ("or choose to stop flipping") |
| flip again, decided each time by the last flip's winner | Game of Chaos (life stakes double each flip) |
| heads/tails only (no call) | Ral Zarek's −7 ("for each coin that comes up heads") |
| "whenever you win a coin flip" | Chance Encounter |
| pick at random | Deadbridge Chant (graveyard), Exalted Flamer of Tzeentch (graveyard, filtered), Urza's Bauble (hand) |
| random-order bottom | Icon of Ancestry, Monumental Henge, Planar Genesis (these also wait on #745) |
| roll modifiers, ignore, stored results | Pixie Guide, Barbarian Class level 1, Krark's Thumb, Centaur of Attention (**out of scope**, Decision 10) |

## Decision 1 — One engine-owned randomness API; nothing else touches the source

Every random draw in `internal/game` goes through one unexported
accessor:

```go
// rng.go
type rngStream struct {
    kind   string    // "shuffle", "roll", "flip", "pick", "random_order"
    player uuid.UUID // whose library / who rolls / who discards
    source uuid.UUID // the object whose effect draws; uuid.Nil if none
}

// randForLocked returns a *rand.Rand for ONE random operation (one
// shuffle, one roll of N dice, one pick). Caller must hold g.mu.
func (g *Game) randForLocked(s rngStream) *rand.Rand
```

`Game.rng` and `Game.rngState` are removed. Every one of the seven call sites in the
table above moves to `randForLocked`. That includes cascade's fallback to the
global source and the random discard's "index 0 when nil" branch. No
file in `internal/game` calls into `math/rand/v2` afterwards, except
`rng.go` and the body of `Zone.Shuffle` (see "No nil source" below).
`TestNoDirectRandomSource` (a test in the package that parses its
non-test files) keeps it that way.

The seven existing sites use these streams. None of them has a source
object, and "player" is the player whose library, hand or cards it is
(sub-PR 1 keys that player by seat; see the
[addendum](#addendum-2026-09-17-sub-pr-1-readings-pr-775)):

| call site | stream `(kind, player, source)` |
|---|---|
| opening library shuffle (`game.go`) | `("shuffle", library owner, uuid.Nil)` |
| `Mulligan` (`mutations.go`) | `("shuffle", library owner, uuid.Nil)` |
| `ShuffleLibrary`, the sandbox shuffle (`mutations.go`) | `("shuffle", library owner, uuid.Nil)` |
| `finishSearchLocked`, the shuffle after a search (`effect_api.go`) | `("shuffle", library owner, uuid.Nil)` |
| `ShuffleLibraryForEffect` (`effect_api.go`) | `("shuffle", library owner, uuid.Nil)` |
| `DiscardRandomForEffect` (`effect_api.go`) | `("pick", discarding player, uuid.Nil)`: one pick of n cards |
| `bottomInRandomOrderLocked` (`cascade.go`), later #745's `PutOnBottomInRandomOrderForEffect` | `("random_order", actor: the player the effect belongs to, uuid.Nil)` |

All five shuffles share one stream per player. Each shuffle is one
operation on it, so the counter tells them apart.

Card code never sees a `*rand.Rand`. It calls the public `…ForEffect`
methods in Decision 5.

**No nil source.** A game that never called `Start` (many unit tests)
mints a crypto-random key the first time something draws. Draws stay
random, rewindable and persistable in every game. The
`math/rand/v2` global-source fallback goes away, and so does the
random discard's silent "always the first card".

**The one exemption is `Zone.Shuffle`'s own nil branch.** `Zone.Shuffle(r *rand.Rand)`
keeps its fallback to the global source when `r` is nil, for zone-level
unit tests that shuffle a bare `Zone`. `TestNoDirectRandomSource` allows
that one method body. It also fails on any call `Shuffle(nil)` in a
non-test file of `internal/game`, so no game path can reach the
fallback. Naming the `rand.Rand`, `rand.PCG` and `rand.Source` types in
a signature is allowed, and nothing but `rng.go` may import
`crypto/rand`.

## Decision 2 — Keyed streams: a secret key plus per-stream counters

The state is a **32-byte secret key** and a **map of counters**, not a
single PCG position:

```go
rngKey      [32]byte          // minted at Start from crypto/rand; never changes
rngCounters map[string]uint64 // stream key -> draws taken this turn
rngTurn     int               // the Turn.Number rngCounters belong to
```

`randForLocked(s)`:

1. If `g.Turn.Number != g.rngTurn`, clear `rngCounters` and set
   `rngTurn`. The turn number is part of every stream key (step 2),
   so earlier turns' counters are never read again.
2. `seed := HMAC-SHA256(rngKey, "cmdctrl/rng/v1" ‖ len(kind) ‖ kind ‖ player[16] ‖ source[16] ‖ turn[8] ‖ counter[8])`.
3. Increment the counter, and return `rand.New(rand.NewChaCha8(seed))`.
   ChaCha8 is in `math/rand/v2` from Go 1.22, the module's floor.

Sub-PR 1 builds "player" as the seat and "turn" as a per-turn index
rather than `Turn.Number`. The
[addendum](#addendum-2026-09-17-sub-pr-1-readings-pr-775) records why,
and records how sub-PR 2 builds "source".

**Why keyed streams rather than one copied PCG.** This implements the
owner's goal for Decision 4, "undo can't be used to fish", and
the Context shows that one copied stream doesn't reach it:

- **Spending a draw elsewhere no longer moves your roll.** The Ogre's
  roll comes from `("roll", you, Ogre, turn)`, and the Puzzlebox's rolls
  and your fetchland's shuffle come from other streams. After an undo,
  the Ogre rolls the same number whatever you did in between.
- **Choosing a different line can still give a different result, as it
  would in paper.** Rolling for a *different* card (a second die-rolling
  permanent or spell) uses a different source, so it gets a different
  roll. To fish that way you need a second card that does the same
  random thing, and you have to spend it.
- **The exception is the random discard.** `DiscardRandomForEffect` has
  no source parameter, so it draws from `("pick", player, uuid.Nil)`
  (Decision 5), and every random discard by one player in one turn
  shares that stream. If you undo Gamble and cast Burning Inquiry
  instead, your discard is the same draw on the same stream: with the
  same hand, the same card. This gives less room to fish than a
  per-source stream, not more. It stays that way until a source is
  added to the random discard, which is outside this ADR.
- **Turn scoping limits what an undo teaches you.** After undoing a roll, the
  player knows what that source's stream will give next, *this turn*.
  At the next turn the key changes and that knowledge is gone.
- **Observed outputs don't reveal future draws.** A PCG's state can in
  principle be recovered from enough outputs, and die rolls are public
  outputs. Each draw here is seeded by an HMAC of a secret key.
- **Tests become persistable and rewindable.** `Start(r *rand.Rand)` and
  `StartWithSource(pcg)` keep their signatures and derive the key from
  four `Uint64()` calls on the source the caller passed, written
  little-endian into the 32 bytes. (`math/rand/v2`'s `*rand.Rand` has
  no `Read`.) A seeded test
  stays deterministic, and there is no `"external"` source any more that
  can't be persisted or rewound. `SetRNGKeyForTest([32]byte)` follows
  the existing `…ForTest` pattern (`layers.go:685`) for tests that never
  call `Start`.

Stream keys are strings in a map rather than a struct key, so the
snapshot encodes them with no custom marshalling. The map holds a
handful of entries per turn.

## Decision 3 — Clone and snapshot carry the key and the counters

- **Clone** (`cloneLocked`) copies `rngKey` (an array, so a value
  copy), deep-copies `rngCounters` and copies `rngTurn`. `RestoreFrom`
  assigns all three. The file header and the comment at `clone.go:41-50`
  are rewritten: undo now **does** rewind randomness.
- **Snapshot.** `rngSnapshot` gains a kind, `"keyed"`, with `key []byte`
  (32 bytes), `counters map[string]uint64` and `turn int`.
  `snapshotRNG` always writes it after `Start`. `UnpersistableRNG`
  can no longer be set and stays only so older census JSON still decodes.
- **Restoring older files.** A `"pcg"` or `"external"` file restores
  with a freshly minted key and empty counters. This loses nothing a
  player could notice. The RNG holds no hidden
  information that has already been decided: library orders are stored in the zones themselves, so a new key
  changes only *future* shuffles and rolls. It is the same answer
  `restoreRNG` already gives `"external"`.
- **No `SnapshotSchemaVersion` bump.** An old file on the new binary
  restores correctly, as just described. A new file on an older binary (a
  rollback) reaches `restoreRNG`'s default branch and leaves `g.rng`
  nil. Shuffles then use the global source, and a random discard takes
  the first card until the next deploy. That degrades randomness during
  a rollback window, but it does not corrupt state, and a bump would
  instead abandon every live game on rollback (ADR 0041 Decision 5).
- **Drift test.** `snapshot_drift_test.go:104,110` swaps the `rng` /
  `rngState` rows for `rngKey`, `rngCounters` and `rngTurn`, all
  `carried`.
- **The key stays on the server.** It is in the engine `GameSnapshot`
  on the server's disk, as the PCG state is today. It is **not** in
  `GameView`, so it stays out of the crash dump and the JSONL replay
  (both are `protocol.SnapshotPayload{Game: view}`, `ws/room.go:489-497`),
  bug reports and every client frame. Anyone who had the key could
  predict every future roll.

## Decision 4 — Undo rewinds randomness (owner)

Given Decisions 2 and 3, rewinding needs no code in `ws`.
`Room.undo` restores the pre-action clone, the clone carries the
counters, and the next identical action gets the identical draw. A
bundle rollback (`rollbackLocked`, `ws/room.go:265`) rewinds the same way.

This applies to **every** random use, not only the new ones:

| undo, then redo the same thing | today | after this ADR |
|---|---|---|
| a search that shuffles, `ShuffleLibraryForEffect`, a mulligan | new library order | same order |
| `DiscardRandomForEffect` (Gamble, Burning Inquiry) | possibly a different card | same card |
| cascade's random bottom / #745's `PutOnBottomInRandomOrderForEffect` | new order | same order |
| a die roll, a random pick | (didn't exist) | same result |
| a won/lost coin flip, even when the call is changed | (didn't exist) | same won/lost (Decision 6) |

**What rewind costs, written down as #744 asked:**

- **A player can learn a future value.** A player who undoes a roll knows
  what that source will roll next this turn, and can decide not to
  trigger it again. That is information paper Magic never gives. It is
  limited to that stream and that turn (Decision 2), and to the undo
  budget (`UndoLimit`, default 1). The budget is only a soft limit,
  because `set_undo_limit` is a sandbox verb any seated player can send.
- **The roll may sit in someone else's undo entry.** A roll made while
  the stack resolves belongs to whoever passed last, often the
  opponent. Under rewind, that player's undo and redo reproduce the roll
  exactly, so the entry's owner can't use it to change your result. Fencing undo past a roll would remove
  it, and the owner chose rewind over that.
- **A flip chain with "stop" lets undo turn a lost flip into a stop.**
  Fiery Gambit, or Game of Chaos when its controller decides: the
  flipper answers the call, sees the flip is lost, undoes the answer
  (it is its own `Apply`, so its own undo entry), and answers "stop"
  instead, which keeps the wins so far. The rewind doesn't prevent
  this. Calling again reproduces the same loss, but "stop" is a
  different answer, and in paper you must choose to stop *before* you
  see the next flip. Drawing the won/lost bit when the prompt is queued
  would not close it either, because the undone prompt would hold the
  same bit. Only fencing undo past a flip would close it, and the owner
  chose rewind over fencing. The cost is bounded like every other undo:
  one peek per undo, within `UndoLimit`. With the default budget of 1,
  a player can save at most one losing flip per turn. For example, a
  Fiery Gambit with two wins whose third flip loses can become a stop at
  two wins.
- **"Undo the fetch to reshuffle" stops working.** It was never
  a feature, but players may have relied on it.
- **The admin undo** (caller `uuid.Nil`, no budget) rewinds too. An
  admin is trusted.

## Decision 5 — The effect API

All in `internal/game/effect_api_random.go`. Each method takes the lock
already held and draws through `randForLocked`.

```go
// RandomDraw names the stream: who is doing it and which object's
// effect it is. Source is uuid.Nil when there is no object.
type RandomDraw struct{ Player, Source uuid.UUID }

// RollDiceForEffect rolls n dice with `sides` sides (CR 706.1a) and
// returns the natural results in roll order (CR 706.2). Emits one
// EventRollDie per die, grouped by BatchSeq. n <= 0 or sides <= 0 is
// ErrInvalidParam.
func (g *Game) RollDiceForEffect(d RandomDraw, sides, n int) ([]int, error)

// FlipCoinsForEffect flips n coins that nobody wins or loses
// (CR 705.2 first sentence: Ral Zarek's "comes up heads"). Returns
// true for heads. Emits one EventFlipCoin per coin. No prompt.
func (g *Game) FlipCoinsForEffect(d RandomDraw, n int) ([]bool, error)

// FlipCoinForEffect is a won/lost flip. It queues the call prompt
// (Decision 6) and runs `then` with the results once the flipper
// answers. See Decision 6 for AllowStop and Coins.
func (g *Game) FlipCoinForEffect(spec CoinFlipSpec) uuid.UUID

// ChooseAtRandomForEffect returns k distinct IDs from ids, chosen
// uniformly at random, in random order (k >= len(ids) returns all
// of them in a random order). The IDs can be cards in any zone,
// permanents or players. The caller decides what they mean. No event:
// no card triggers on "chosen at random", and the move that follows
// logs itself.
func (g *Game) ChooseAtRandomForEffect(d RandomDraw, ids []uuid.UUID, k int) []uuid.UUID
```

- **The random discard reuses the pick.** `DiscardRandomForEffect`
  keeps its signature and chooses through `ChooseAtRandomForEffect` on
  the stream `("pick", discarding player, uuid.Nil)`. It has no source
  parameter today, and adding one is outside this ADR. Everything it
  already does (one `EventDiscardCard` per card, the #650 CR 614 gap) is
  unchanged.
- **`b27PseudoRandomIndex` is deleted.** `b27LookAtRandomCardInHand`
  calls `ChooseAtRandomForEffect(RandomDraw{activator, source}, handIDs, 1)`.
  Urza's Bauble loses its caveat and becomes `CompletenessFull` if
  nothing else is missing, and the census is regenerated.
- **Several dice, then the modifiers, then the events.**
  `RollDiceForEffect` draws, runs a (for now empty) CR 706.2 / 706.6
  modifier pass, and emits events only after that. When Pixie Guide lands,
  the extra die is drawn and ignored *before* any event exists, which
  706.6 requires ("no abilities trigger because of the ignored roll").
  Reordering after events had been emitted would be the harder change.
- **Catalog helpers** (`cards/effects/dice.go`):
  - `RollD(n int)` returns one result.
  - `RollTable{Sides, Rows []RollRow{Min, Max int; Effect}}` covers
    CR 706.3: `Max: 0` means "N+", and a row can set `RollAgain: true`
    for 706.3c, which asks through a Confirm prompt.
  - `ChooseOneResult` covers Reckless Endeavor. Its two results become
    a `ConfirmPrompt` whose two labels spell out each assignment.
  - Doubles (706.5) is a helper over the returned slice.

## Decision 6 — A won/lost flip prompts for a call (owner)

**The prompt.** It is a new kind, `PendingChoiceCoinCall = "coin_call"`,
queued by `FlipCoinForEffect`:

```go
type CoinFlipSpec struct {
    Flipper   uuid.UUID
    Source    uuid.UUID
    Question  string // the card's sentence, for the prompt header
    Coins     int    // coins flipped under this one call (owner decision 4)
    AllowStop bool   // Fiery Gambit's "or choose to stop flipping"
    MaxUsefulWins int // bot hint: wins after which stopping is right; 0 = none
    Then      func(g *Game, r CoinFlipResult) error
}
type CoinFlipResult struct {
    Call    string // "heads" | "tails"; "" when the flipper stopped
    Faces   []bool // true = heads
    Won     []bool
    Stopped bool
}
```

- **Wire.** `resolve_choice` gains `call: "heads" | "tails" | "stop"`,
  dispatched by kind next to the confirm case (`actions/actions.go:1242`).
  `PendingChoiceView` carries `kind: "coin_call"`, `reason`, and
  `allow_stop` and `coins` when set. It also carries `max_useful_wins`
  and `wins` (the wins so far in the chain) when set, because the
  heuristic reads only `protocol.GameView` and can't see `CoinFlipSpec`
  (ADR 0033 §3; Decision 7).
- **"Flip until you lose"** (Fiery Gambit) is a chain.
  `Then` queues the next `FlipCoinForEffect` while the flipper keeps
  winning. It uses the chained-choice mechanism, not a new subsystem.
  "Choose to stop" is the prompt's third answer, so each flip costs one
  prompt, not a call plus a "continue?" confirm. The **first** prompt
  has `AllowStop: false`: the Scryfall rulings (2004-12-01) say "After
  each flip, you choose whether to continue flipping", so the first
  flip is not optional. They also say that if the target creature is
  illegal on resolution, "you don't even flip a coin".
- **Game of Chaos is not "flip until you lose".** Oracle text: "Flip a
  coin. If you win the flip, you gain 1 life and target opponent loses
  1 life, and you decide whether to flip again. If you lose the flip,
  you lose 1 life and that opponent gains 1 life, and that player
  decides whether to flip again. Double the life stakes with each flip."
  Its one Scryfall ruling (2004-10-04) says only that doubling the stakes
  "means to double the amount of life lost or gained" (1, 2, 4, …).
  No ruling says who flips the next coin, so this ADR reads it from the
  text and CR 705.2 ("Only the player who flips the coin wins or loses
  the flip"). "If you win the flip" is only meaningful when "you", the
  controller, flip. So **the controller is the `Flipper` of every flip,
  and only the decision to flip again moves to the opponent**. This
  reading is **owner-decided (2026-09-17)**:
  - After a won flip, the controller decides. The next flip's
    `coin_call` goes to the controller with `AllowStop: true`, so that
    flip costs one prompt, as for Fiery Gambit.
  - After a lost flip, the target opponent decides. That player is not
    the flipper, so the decision can't be the "stop" answer on the
    controller's prompt. It is a `PendingChoiceConfirm` to the opponent
    ("flip again?"). If they accept, a `coin_call` goes to the
    controller with `AllowStop: false`. That flip costs two prompts.
  - The first flip is not optional (`AllowStop: false`).

  **For the card PR, check:** (1) look again for newer rulings or a
  Gatherer ruling on who flips after a loss. The flipper is
  owner-decided, so a ruling that says the opponent flips is raised
  with the owner before `Flipper` or the life changes are touched; (2) the chain has no natural end, because state-based
  actions aren't checked while the spell resolves (CR 704.3), so life
  totals can go below 0 mid-chain. The stake must not overflow `int`,
  and the bot hint must stop the chain (for example, stop once the
  stake is at least the losing player's life), so that a bot controller
  and a bot opponent can't flip forever; (3) the opponent's
  "flip again?" confirm needs a bot preference of its own, because the
  heuristic prefers the accept branch of a confirm
  (`heuristic/choices.go:160`).

**The result is drawn as won/lost, and the face follows from the call.**
This is a technical choice, and it is what makes the two owner decisions
work together. When the call arrives, the engine draws **one fair bit, won or
lost**, from `("flip", flipper, source)`. The face shown is the call if
won and the other side if lost.

Drawing the face and comparing it to the call instead would break
Decision 4. The answer to the prompt is its own `Apply`, so its own undo
entry. With the face drawn independently, **undoing a lost call and
calling the other side wins every time**, and the rewind would make
that fishing certain rather than likely. With the won/lost bit drawn, an
undone and changed call keeps the same won/lost result and only the face
changes.

Nothing about the odds changes. For either call, the face is
heads or tails with probability ½ and independent of the call, and a win
is ½ (CR 705.1 asks for exactly this). The bit is drawn **when the
call is answered**, not when the prompt is queued, so no decided result
sits in state while the prompt is open. A flip with no call
(`FlipCoinsForEffect`) draws the face directly.

**Events.** `EventRollDie` (`Actor`, `Source`, `Amount` = natural
result, new `Sides int`) and `EventFlipCoin` (`Actor`, `Source`,
`Label` = `"heads"` / `"tails"`, new `Call string`, new `Won bool`),
each with a new `BatchSeq uint64`. `BatchSeq` groups the per-die or per-coin events
of one instruction, and holds the `Seq` of the first one, following the
`RevealSeq` precedent (`events.go:521-535`).

- "Whenever you roll one or more dice" (Vexing Puzzlebox) uses the
  existing `OncePerBatch` flag (`triggers.go:184-194`) and reads the
  batch through `BatchSeq`.
- "Whenever you win a coin flip" (Chance Encounter) matches on
  `EventFlipCoin` with `Won` and fires once per won flip.
- All new fields are `omitempty`, so there is no snapshot schema bump.

**Public log.** New kinds `LogRoll = "roll"` and `LogFlip = "flip"`, one
entry per batch, collapsed by `BatchSeq` the way reveals collapse
(`log.go` `revealAt`). `LogEvent` gains `sides`, `results []int`,
`faces []string`, `call` and `wins`, all `omitempty`. Text is rendered
on the server, for example "Alice rolled a d20 for Hoarding Ogre: 14",
"Alice rolled 3d6 for Clown Car: 2, 5, 6", "Alice called heads for Mana
Crypt: tails, lost". Final wording is settled in sub-PR 2's review.
Rolls and flips are public information, so the log needs no redaction
beyond the source card's existing `known_by` treatment.
`docs/protocol.md` gets both kinds, and `client/src/lib/protocol.ts`
mirrors the types. The board also shows a reveal-strip cue for each roll or flip
(owner decision 3, sub-PR 4).

## Decision 7 — Bots and the legal enumerator

- **Enumerator** (`legal/choices.go`, next to the confirm case at
  `:265`). A `coin_call` offers "call heads" and "call tails", plus "stop
  flipping" when `allow_stop` is set. Heads is the `AlwaysLegal` answer
  (#544). `ResolveCoinCall` never refuses a well-formed call, so every
  offered move is accepted (`dispatchAll`, #499).
- **Where a bot's random call comes from.** Neither Layer A nor the
  heuristic has a random source. `rules.Resolve` (`aiseat/rules/rules.go:104`)
  is a pure function of `aiseat.Input` that allocates nothing, and the
  heuristic scores are deterministic. So both take the call from **one
  bit of the pending choice's `id`**: heads if the low bit of the UUID's
  last byte is 0, tails otherwise. One shared helper in `aiseat`
  computes it. The engine mints that `id` with `uuid.New()`
  (`pending_choice.go`), which is a v4 UUID from crypto/rand, so the bit
  is uniform. It is not drawn from the game's key, and it is decided
  before the won/lost bit is drawn. A bot's call is therefore random
  (owner decision 2) without adding state to either layer. The cost is
  that the *face* shown for a bot's flip can differ between two runs of
  a seeded test game, because the choice id is not seeded. The won/lost
  result doesn't differ, because it never depends on the call
  (Decision 6).
- **Bots call at random (owner).** The heuristic (`aiseat/heuristic/choices.go`)
  gains `choiceCoinCall` and scores the call from the id bit above
  highest. "Stop" is taken once `wins` in the view reaches
  `max_useful_wins` (Fiery Gambit: 3, after which another flip can only
  lose). Zero means the heuristic keeps flipping.
  Layer A **absorbs** a `coin_call` window that doesn't offer "stop",
  with the same id-bit call, so a model tier never spends a call on a
  choice that carries no information. A window that offers "stop" is a
  real decision, so Layer A escalates it. Because both layers compute
  the call from the same bit, they agree on every window Layer A
  absorbs, and the Layer-A/heuristic agreement test
  (`TestLayerAAbsorbsMostWindowsAndAgreesWithTheHeuristic`) needs no
  exemption.
- **The random tier** (`NewRandomPolicy`) picks among the offered
  answers uniformly, as it does for every window, "stop" included.
- **Bots never draw from the game's key.** A bot's randomness is its
  own (`NewRandomPolicy`'s PCG, `aiseat/policy.go:95`). If a bot drew
  from game streams, its thinking would move the table's results.
- **Rolls and picks happen inside resolution** and add no windows. A bot
  sees them only as log entries, which the model prompt already renders.

## Decision 8 — The random-order bottom is #745's, and draws from a stream

[#745](https://github.com/krakenhavoc/cmd_and_ctrl/issues/745) builds
`PutOnBottomInRandomOrderForEffect(actor, from, ids)`: only the cards
in `ids` still in a zone of kind `from` move (PR #814), each through
`routeCardToZoneLocked` (`zone_route.go:132`, `ToBottom`), with cascade
switched to call it. This ADR changes one thing about it:
**the permutation is drawn from `randForLocked(rngStream{"random_order", actor, uuid.Nil})`**.

- If #745 merges first, sub-PR 1 changes that one draw together with the other
  call sites in the Decision 1 table.
- If sub-PR 1 merges first, #745 calls `randForLocked` directly.

The signature, the zone handling, the CR 903.9 commander offer and the
knower clearing are all #745's, and none of them change here.

*Corrected 2026-09-17:* the signature as merged is
`PutOnBottomInRandomOrderForEffect(actor uuid.UUID, from ZoneKind, ids []uuid.UUID) error`
(`game/random_bottom.go`), after
[PR #814](https://github.com/krakenhavoc/cmd_and_ctrl/pull/814).
`actor` is the player the effect belongs to: it is stamped on the zone
move events and keys the stream, `("random_order", actor, uuid.Nil)`.
Only cards **still in a zone of kind `from`** move (`ZoneExile` for
cascade's pile, `ZoneLibrary` for "the rest" of a reveal or a look). A
card that has since left that kind of zone is a new object (CR 400.7)
and is skipped. The first draft of this decision said `(player, ids)`
and "any source zone". The shuffle is still one draw over `ids` as
given, taken before any card is looked up, so the skip never changes how
much the stream advances.

## Decision 9 — Determinism and replay

- **Tests.** The same key and the same action sequence give the same
  results. `SetRNGKeyForTest` or `StartWithSource` with a fixed seed
  pins them. A known-answer test pins the HMAC-ChaCha derivation itself
  to fixed outputs for a fixed key, because a silent change to the derivation
  would change what every restored game rolls next.
- **Replay.** ADR 0041 Decision 1 rejected action-log replay, and nothing
  here brings it back. The JSONL replay is a log of views and carries no key. If
  action replay ever comes back, keyed streams are what it would need:
  results depend on the key, the stream and the count, not on
  everything else that drew before them.
- **The catalog soak** (`aiseat/catalog_soak_test.go:189-199`) seeds
  its games already. Its runs stay reproducible, and they will produce
  different games than before this change (see Consequences).

## Decision 10 — Out of scope, stated

- **Roll modifiers and rerolls** (CR 706.2, 706.2b): Wyll's Reversal's
  "+ greatest power" is a card-side addition to the natural result
  that the card applies itself. A general modifier pipeline is not built.
- **Ignoring a roll** (CR 706.6): Pixie Guide, Barbarian Class level 1.
  Decision 5 orders the API so it can be added without moving events.
- **Fixed flip results and extra flips** (CR 705.3): Krark's Thumb.
- **Stored results** (CR 706.8): Centaur of Attention.
- **The planar die** (CR 706.7, 901): no Planechase.
- **Reflexive "when you do"** (Ancient Brass Dragon) is #636.
- **Changing a spell's targets and copying abilities** (Wyll's Reversal,
  Krark, the Thumbless's copy) belong to their own seams.

## Consequences

- **Every random operation is rewindable, persistable and deterministic
  under a fixed key**, including the ones that already shipped. The
  `UnpersistableRNG` census case can no longer occur, so a seeded test game becomes a
  restore point.
- **Seeded test outcomes change once.** Every seeded shuffle draws from
  a different derivation. A test that pins a specific library order or
  hand from a seed has to be re-pinned in sub-PR 1, and that PR's
  description lists each one. Tests that relied on the random discard
  taking the first card when no RNG was set (`effect_api.go:341`) must
  set a key or assert membership instead.
- **The random discard stops being "index 0" in harness games** (see
  above). This fixes a test-only artifact. It is not a behaviour change
  in production.
- **The undo cost is written down in Decision 4.** A player can learn one stream's
  next value for a turn, and "undo to reshuffle" is gone.
- **An undone and changed coin call shows the other face with the same
  won/lost result** (Decision 6). A player only sees this if they undo.
- **In a flip chain that allows "stop", undo can turn a lost flip into a
  stop** (Decision 4): answer, see the loss, undo, answer "stop" and
  keep the wins. This affects Fiery Gambit and Game of Chaos, and it is
  limited to one peek per undo within the undo budget.
- **One HMAC-SHA256 plus a ChaCha8 setup per random operation**:
  microseconds, and a game has a few dozen random operations per turn.
- **Krark, the Thumbless prompts on every instant and sorcery its
  controller casts**, because each cast is a won/lost flip. That is the
  price of the owner's prompt decision on the one in-scope card that flips
  often. The one-call-per-instruction decision doesn't change it (one coin per trigger).
- **A coin call blocks the table** like any pending choice, and while it
  is open no restore point is written (the continuation is censused as
  `ChoiceResumeFrames`, `chained_choice.go` header). The time it holds
  that open is one click.
- **`docs/engine-seams.md`**: the "Effect-reachable RNG / die roll" row
  moves to Closed when sub-PR 2 lands. The dice and coin half of it
  closes with sub-PR 3.

## Alternatives considered

- **(a) Keep today's behaviour: undo re-rolls, limited only by the budget.**
  Rejected by the owner (Decision 4).
- **(c) A roll or flip can't be undone past (the stack is fenced).**
  Rejected by the owner (Decision 4). It would also have needed a
  new undo-stack concept, because nothing today blocks undo after hidden
  information comes out.
- **Rewind with one PCG copied by value into the clone.** This is the smallest
  change, `pcg := *g.rngState`. Rejected under Decision 2 because spending
  a draw elsewhere moves the next result, which makes fishing an undo plus a
  fetchland. Tests with a caller-supplied `*rand.Rand` still couldn't
  be rewound, and observed outputs reveal the PCG's state.
- **Keying streams by event sequence number.** Every action emits events,
  so any cheap action (tapping a land) would move the result. Rejected.
- **Keying streams without the source.** Vexing Puzzlebox's at-will
  mana-ability roll would move Hoarding Ogre's roll. Rejected.
- **Keying streams without the turn.** An undone roll's next value would
  stay known for the rest of the game. Rejected.
- **Drawing the coin's face and comparing it to the call.** Under rewind,
  undoing a lost call and changing it always wins. Rejected (Decision 6).
- **Reusing `PendingChoiceConfirm` with Heads/Tails labels.** The heuristic
  prefers the accept branch (`heuristic/choices.go:160`), so the bot
  would always call heads, and the model tier would escalate a choice that
  carries no information. It also can't carry a third "stop" answer. Rejected for
  a dedicated kind (Decision 6).
- **Auto-calling as a declared simplification.** The odds are
  identical, but the owner chose the prompt (Decision 6).
- **Keeping the global `math/rand/v2` fallback for games that never started.**
  It can't be persisted or rewound, and it is the source ADR 0041 moved
  away from. Rejected for lazy key minting (Decision 1).
- **An event for random picks.** No card triggers on "chosen at random",
  and the move that follows is already logged. Rejected (Decision 5).

## PR split

**Sub-PR 1 — engine: keyed streams, rewind, persistence. No card changes.**
- `internal/game/rng.go`: key, counters, `randForLocked`,
  `SetRNGKeyForTest`, key derivation in `Start` / `StartWithSource`.
- Move the seven call sites in the Context table (and #745's draw, if it
  has merged). Remove `Game.rng` / `rngState`. Add `TestNoDirectRandomSource`.
- `cloneLocked` / `RestoreFrom`, and rewrite the clone.go comments.
- `rngSnapshot` `"keyed"`, restoring older `"pcg"` / `"external"` files, and the drift
  test rows.
- Re-pin any seeded test whose outcome changes, and list them in the PR.
- Checks: `go test ./internal/game/... ./internal/ws/... ./internal/aiseat/... ./internal/legal/...`,
  then `go test ./...` and `make lint`.

**Sub-PR 2 — effect API, events, log.**
- `RollDiceForEffect`, `FlipCoinsForEffect`, `ChooseAtRandomForEffect`,
  and `DiscardRandomForEffect` moved onto the pick.
- The per-game object ordinal for the stream's source half, carried by
  clone and snapshot ([addendum](#addendum-2026-09-17-sub-pr-1-readings-pr-775)).
- `EventRollDie`, `EventFlipCoin`, `BatchSeq`, `LogRoll`, `LogFlip`,
  `docs/protocol.md`, `protocol.ts` types.
- `cards/effects/dice.go` helpers. Delete `b27PseudoRandomIndex`,
  un-caveat Urza's Bauble, and regenerate the census.
- `docs/engine-seams.md`: move the row to Closed.

**Sub-PR 3 — the coin call.**
- `PendingChoiceCoinCall`, `CoinFlipSpec`, `FlipCoinForEffect`,
  `ResolveCoinCall`, `resolve_choice.call`, `PendingChoiceView`.
- Enumerator case, heuristic case, Layer A rule, gamecli.
- `ChoicePromptModal.svelte`: a coin-call rendering with Heads, Tails
  and, when allowed, Stop.

**Sub-PR 4 — client roll and flip cue** on the reveal strip (`RevealBanner`, `reveals.ts` rules), built from `roll`/`flip` log entries with no new `GameView` field. No animation.

**Card PRs** — the first wave is owner decision 5: the fully unblocked cards, Urza's Bauble's caveat removal, and Fiery Gambit, Game of Chaos and Yusri. Each card follows
AGENTS.md (`Spec` slots, completeness, caveats that are weaker than
printed and never stronger).

## Test plan

Engine (`internal/game`):

1. **Known answer.** A fixed key and stream give pinned outputs from
   `randForLocked`, and a pinned d20 sequence from `RollDiceForEffect`.
2. **Same key, same actions, same results** for a shuffle, a random
   discard, a roll, a pick and a flip.
3. **Stream independence.** Drawing on `("roll", p, puzzlebox)`
   doesn't change the next value of `("roll", p, ogre)`. Drawing on
   `("shuffle", p)` doesn't change either of them.
4. **Turn scoping.** The same stream gives a different value on the
   next turn, and the counters map is cleared when the turn changes.
5. **Rewind** (`Clone` → draw → `RestoreFrom` → draw gives equal values)
   for each of: opening shuffle, `ShuffleLibraryForEffect`, a search
   shuffle, `Mulligan`, `DiscardRandomForEffect`, cascade's (or #745's)
   random bottom, `RollDiceForEffect`, `ChooseAtRandomForEffect`,
   `FlipCoinsForEffect`, and the answer to a coin call.
6. **The Context's fishing scenario.** Roll for source A, restore,
   draw on source B, roll for A again: same value.
7. **Coin call under rewind.** Answer heads and record won/lost, restore,
   answer tails: same won/lost, opposite face.
8. **Fairness, deterministic.** Under a fixed key, 20,000 d20s land in
   [1, 20] with every face present and within a fixed tolerance.
   10,000 called flips are won 50% ± 2% for each call, and uncalled faces
   come up heads 50% ± 2%. Because the key is fixed, the result can't flake.
9. **Snapshot.** Capture → encode → decode → restore, and the next draws
   on several streams equal the live game's. The round-trip property
   holds. `UnpersistableRNG` is false for a game started with
   `Start(rand.New(...))`.
10. **Older files.** A `"pcg"` and an `"external"` `rngSnapshot` restore
    to a keyed game that is restorable and draws. The committed fixture
    `testdata/snapshot_pre683.json` still restores. Once #522's corpus
    exists, a keyed restore point is added to it.
11. **Events and log.** Multi-die batches share `BatchSeq`. `LogRoll` /
    `LogFlip` collapse per batch and render text. A Vexing Puzzlebox-style
    `OncePerBatch` trigger fires once for a 3-die roll. A "whenever you
    win a coin flip" trigger fires only on won flips.
12. **`TestNoDirectRandomSource`**: no non-test file in `internal/game`
    calls into `math/rand` or `math/rand/v2` outside `rng.go`, except
    inside the body of `Zone.Shuffle` (its nil fallback, kept for
    zone-level unit tests; Decision 1). No non-test file calls
    `Shuffle(nil)`. Only `rng.go` imports `crypto/rand`. Naming the
    `rand.Rand`, `rand.PCG` and `rand.Source` types is allowed.

Room (`internal/ws`):

13. Apply an action that rolls, `Undo`, and Apply it again: the same
    `EventRollDie` values. A failed bundle rolls back the counters.

Enumerator and bots:

14. `legal`: a `coin_call` offers heads and tails (plus stop when allowed),
    exactly one answer is `AlwaysLegal`, and `dispatchAll` accepts every
    move (#499, #544).
15. `aiseat`: the heuristic answers a `coin_call`, calling from the
    choice id's bit and stopping at `max_useful_wins`. Layer A absorbs a
    `coin_call` without "stop", makes the same call as the heuristic,
    and escalates one with "stop". The agreement test passes with no
    exemption. The random tier answers it. Over many minted ids, the
    id-bit call comes out heads about half the time.
16. **Catalog soak** (#601) with the first-wave cards added: no stalls on
    a coin call.

Client:

17. `ChoicePromptModal` coin-call rendering: checked by hand until
    [#689](https://github.com/krakenhavoc/cmd_and_ctrl/issues/689), with
    any pure helper unit-tested with vitest.

## Owner decisions, second round

Answered 2026-09-17. The chosen option is marked **(chosen)**; the recommendation text is kept for the record.

1. **How does the board show a roll or a flip?**
   - (a) Log line only.
   - (b) **(chosen)** Log line plus a short cue on the existing attention strip
     (`RevealBanner`), for example "Alice rolled a d20 for Hoarding Ogre: 14".
     It would reuse `reveals.ts`'s rules (cue once per `seq`, time out,
     prime on reconnect) and be built from `log` entries, with no new
     `GameView` field.
   - (c) (b) plus an animated die or coin.

   **Recommendation: (b).** A d20 that decides how many Treasures
   appear is a moment the table should see. The log panel is easy to
   miss, and the reveal strip already solves reconnects and dropped
   frames. An animation adds client work, and reduced-motion handling
   (ADR 0053 Decision 2) for little gain.

2. **"Flip N coins" with a call: one prompt per coin, or one call for
   all N?** Yusri, Fortune's Flame flips up to five coins.
   Ral Zarek's −7 and other heads-only flips are not affected, because
   they are never called.
   - (a) One prompt per coin (five prompts for Yusri).
   - (b) **(chosen)** One call covers every coin of one instruction. Each coin still
     gets its own independent won/lost draw.

   **Recommendation: (b).** The call has no effect on the odds
   (Decision 6), so five clicks add friction and no decision. "Flip until
   you lose" (Fiery Gambit) is a series of separate instructions and
   still prompts once per flip, which is where a player's choice to stop
   actually lives.

3. **Which cards ship in the first wave?**
   - (a) Only the cards this seam fully unblocks: Ancient Copper /
     Gold / Silver Dragon, Hoarding Ogre, Deadbridge Chant, Exalted Flamer
     of Tzeentch, The Gold Saucer, Vexing Puzzlebox, Clown Car, Reckless
     Endeavor, Goblin Archaeologist, plus Urza's
     Bauble's caveat removal.
   - (b) **(chosen)** (a) plus the flip-until-lose cards (Fiery Gambit, Game of Chaos)
     and Yusri, which exercise the chained prompt.
   - (c) Seam only, with cards left to the batch issues.

   Excluded either way: Mana Crypt (banned in Commander); Stitch in Time
   and Ral Zarek (the engine has no extra turns); Chance Encounter (no
   "you win the game" effect); Frenetic Efreet (no phasing). Out of
   scope as stated in Decision 10: Pixie Guide, Barbarian Class,
   Krark's Thumb, Centaur of Attention, the planar die.

   **Recommendation: (b).** Game of Chaos and Fiery Gambit are the only
   cards that exercise `AllowStop` and a chain whose flipper changes. If
   they don't ship with the seam, that code ships untested by any real
   card, and its first real user finds the bugs.

   *Note added 2026-09-17:* in Game of Chaos the player who **decides**
   whether to flip again changes, but the flipper does not
   ([Decision 6](#decision-6--a-wonlost-flip-prompts-for-a-call-owner)).
   The owner confirmed this on 2026-09-17: the controller always flips,
   and the other player only decides whether to flip again.

## Addendum (2026-09-17): sub-PR 1 readings (PR #775)

Sub-PR 1 ([PR #775](https://github.com/krakenhavoc/cmd_and_ctrl/pull/775),
`feat/keyed-rng-744`) builds two parts of the Decision 2 stream key
differently from how Decision 2 wrote them. Both follow Decision 2's
and Decision 9's own goals. They are recorded here so this ADR matches
the code.

### The player half is the seat, not the player's UUID

`AddPlayer` mints a random UUID for each player on every run. If the
stream key used that UUID, a seeded test would shuffle differently on
each run, which breaks "a seeded test stays deterministic" (Decision 2)
and "its runs stay reproducible" (Decision 9). So a seated player is
keyed by **seat index**. Seats can't change once a game starts, because
`RemovePlayer` only works in the lobby. The `player[16]` bytes of the
seed are `"cmdctrl-seat"` followed by the seat number as a big-endian
`u32`. Byte 6 of that is `'l'` (`0x6c`), so it can never equal a v4
UUID, which has `0x4_` there. (The code comment in PR #775's `rng.go`
says `'r'`, which is byte 5; the property holds either way.) The counter map's label is `seat<N>`. A UUID that
isn't seated, or `uuid.Nil`, is keyed by its own bytes.

### Turn scoping uses a per-turn index: `Turn.Number*MaxPlayers + ActiveSeat`

In this engine `Turn.Number` only goes up when play wraps back to seat 0
(`turn.go` `advance`), so it counts **rounds**, not turns. If streams
were scoped to it, what a player learned from an undo would stay valid
through every other seat's turn in that round. Decision 2 says that
knowledge is gone "at the next turn". So `rngTurn`, and the `turn[8]`
bytes of the seed, hold `Turn.Number*MaxPlayers + Turn.ActiveSeat`,
which changes on every turn. Before the first round it is
`0..MaxPlayers-1`, and the opening shuffle runs at index 0.

### Decided for sub-PR 2: the source half is a stable per-game ordinal

**The problem.** Sub-PR 1 passes `uuid.Nil` as the source at every
site. Sub-PR 2 is the first to pass a real source (Hoarding Ogre, Urza's
Bauble). If the source half were the object's instance UUID, it would
behave like the player UUID above. `NewCard` (`card.go`), token creation
(`effect_api.go`) and spell copies (`spell_copy.go`) all mint IDs with
`uuid.New()`. An object keeps its ID through zone changes, clone,
`RestoreFrom` and snapshot restore, so rewind and restore would still
work within one run. Across two runs of the same seeded test, though,
every roll, flip and pick with a source would differ.

**The decision.** The source half of the key is a **per-game object
ordinal** that is the same on every run of the same game:

- **A card from a deck** gets class 0 and the value `seat<<16 | i`,
  where `i` is the card's index in the deck slice passed to `AddPlayer`.
- **An object created during the game** (a token, a spell copy, a
  sandbox card) gets class 1 and the value of a per-game creation
  counter, incremented when the object is created.
- **The ordinal is assigned when the object is created**, not when it
  first draws. If it were assigned on first draw, the order of draws
  would decide the ordinals, and "undo, roll for the Puzzlebox first,
  then roll for the Ogre" would move the Ogre's stream, which is the
  fishing Decision 2 closes.
- **The map from instance ID to ordinal, and the creation counter**, are
  engine state. Clone and snapshot carry them, and the drift test marks
  them `carried`. Sub-PR 2 decides where they are stored (a map on
  `Game` or a field on `Card`).
- **The seed's `source[16]` bytes** are `"cmdctrl-obj"` (11 bytes; byte
  6 is `'l'`, `0x6c`, so it can't equal a v4 UUID), then the class byte, then the
  value as a big-endian `u32`. The counter label is `obj<class>:<value>`.
- **`uuid.Nil` keeps its current bytes and label**, so the counters that
  sub-PR 1 persists and its known-answer test stay valid. An ID with no
  ordinal (a harness card added without `AddPlayer`) is keyed by its own
  bytes, like an unseated player.

**What it costs.** Creating tokens in a different order after an undo
changes a later token's ordinal, and so its stream. That is choosing a
different line with real game actions, like rolling for a different
card (Decision 2). It never applies to cards from a deck. Sub-PR 2 adds
a test that runs the same seeded game twice and gets the same
`RollDiceForEffect` results for a sourced roll, next to Test plan item 2.


## Implementation checkpoint (2026-09-17): effect APIs and first cards

The #744 implementation supplies sub-PRs 2–4 together: random-effect
APIs, public batch events/logs, source ordinals, coin-call continuations,
legal enumeration, bot policy and client prompts/cues. Open coin
continuations clone for undo but, like other closure-backed choices,
make a durable snapshot non-restorable; the continuation census records
that explicitly. Stable snapshots carry RNG counters and source ordinals.

`WheneverYouRollDice` selects the first event of each `BatchSeq`, then
sums that batch at trigger resolution. It deliberately does not use an
in-flight `OncePerBatch` guard: a second dice instruction must trigger
again even while the first ability is still waiting on the stack.

Eight first cards exercise the seam: Ancient Copper Dragon, Ancient Gold
Dragon, Hoarding Ogre, Reckless Endeavor, Vexing Puzzlebox, The Gold
Saucer, Deadbridge Chant and Exalted Flamer of Tzeentch. Urza's Bauble
also moves from its old approximation to the keyed random-pick API.

This checkpoint does not close #744's additional card wave. Goblin
Archaeologist and Fiery Gambit still need compositions and tests; Game
of Chaos needs chained life-change/continuation ordering; Yusri needs a
temporary free-cast permission; Clown Car needs its cast X retained for
its ETB trigger; Ancient Silver Dragon needs a lasting hand-size grant.
Wyll's Reversal also needs spell retargeting. None is declared fully
automated by this change.
