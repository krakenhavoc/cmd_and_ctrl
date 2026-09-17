# ADR 0057 — Winning and losing the game by effect, and "can't lose" / "can't win"

**Status:** Accepted · 2026-09-17 · unscheduled (card-coverage audit, wave 2) · tracked on [#749](https://github.com/krakenhavoc/cmd_and_ctrl/issues/749). The owner's answers to the open questions are recorded in [Decided (2026-09-17)](#decided-2026-09-17).
**Numbering:** 0052 is reserved for the emblems ADR
([#623](https://github.com/krakenhavoc/cmd_and_ctrl/issues/623)); 0055 is
the loop breaker on `develop`; 0056, 0058 and 0059 are being drafted in
parallel for #748, #751 and #753. On 2026-09-17, after `git fetch origin`,
every remote branch (`git for-each-ref refs/remotes/origin`) was listed
with `git ls-tree docs/decisions/`. No branch has a `0057-*` file. The
highest number present is 0055.
**Amends:** [ADR 0051](0051-user-database.md) Decision 4 (the `games`
row gains `outcome` next to `winner_seat`; Decision 7 here).
**Fixes, as sub-PR 1:** [#767](https://github.com/krakenhavoc/cmd_and_ctrl/issues/767)
(a mill that runs a library out loses the game), because it sets the same
flag this ADR gates.
**Coordinates with:** [#769](https://github.com/krakenhavoc/cmd_and_ctrl/issues/769)
(what leaving the game does to a player's objects, CR 800.4a),
[#808](https://github.com/krakenhavoc/cmd_and_ctrl/issues/808) (a player
who leaves mid-drain), [ADR 0059](0059-turn-machinery.md) Decision 6
(the one rotation seam, which fixes
[#766](https://github.com/krakenhavoc/cmd_and_ctrl/issues/766): the turn
rotation when the active player leaves; see Decision 3 here for the
effect-loss constraint both ADRs share), [ADR 0056](0056-infect-wither-toxic.md)
(the poison clock, and the "as though its source had infect" item that
Phyrexian Unlife waits on), and
[#623](https://github.com/krakenhavoc/cmd_and_ctrl/issues/623) (emblems,
for Gideon of the Trials' "can't lose" emblem).
**Builds on:** [ADR 0033](0033-ai-bot-seat.md) (bots read only
`GameView`), [ADR 0035](0035-until-end-of-turn-effects.md) (the turn-scoped
registry pattern), [ADR 0041](0041-game-persistence.md) (snapshots and
the drift test).

Line references are to `origin/develop` at `684f2786`. The issue's
references were to `ad9413f` and most of them have moved; the ones below
were re-read on `684f2786`.

## Context

### How a game ends today

There is one way out of a game: elimination.

- `eliminatePlayerLocked` (`server/internal/game/mutations.go:2479-2501`)
  sets `Eliminated`, clears `LosesAtNextSBA` (`:2484`), runs
  `cleanupStackForEliminatedLocked` (`:2525`) and
  `advancePastEliminatedLocked`, emits `EventPlayerEliminated`, and sets
  `StateEnded` when `survivors <= 1` (`:2493-2500`).
- It has two callers: `Concede` (`mutations.go:4929`, which first emits
  `EventConcede` at `:4942`) and the player-loss loop in
  `stateBasedActionsLocked` (`mutations.go:2278-2302`).
- There is no winner anywhere in the engine. `Concede`'s comment says so
  ("no separate Winner field is stored (the surviving seat is the
  implicit winner)", `mutations.go:4916-4919`). The client derives the
  winner as the one non-eliminated seat of an ended game
  (`client/src/routes/Game.svelte:469-480`), plays the win or loss sound
  from it (`:529-536`) and prints "Game ended — no survivors" when there
  is none (`:1325-1335`). `docs/protocol.md:231` documents the same
  "implicit winner". ADR 0051's `games.winner_seat` is "NULL until the
  engine reports one" (`0051-user-database.md:214`), and nothing on
  `develop` writes it: the user database is not built yet.
- `Game.End()` (`game.go:640-644`) sets `StateEnded` directly. Nothing
  outside tests calls it.

### The loss checks, and what feeds them

The SBA loop at `mutations.go:2278-2302` eliminates a player when any of
these holds, with no gate of any kind:

| check | rule | how it is set |
|---|---|---|
| `p.Life <= 0` | CR 704.5a / 104.3b | life changes |
| `p.LosesAtNextSBA` | CR 704.5b / 104.3c / 121.4 | see below |
| `p.IsDeadByCommanderDamage()` | CR 704.6c / 903.10a | combat damage from a commander (`damage_tail.go:482-487`) |
| poison `>= PoisonLethal` | CR 704.5c / 104.3d | poison counters |

`LosesAtNextSBA` (`player.go:163-169`) has three writers:

1. **An empty-library draw** (`actuallyDrawCardLocked`,
   `mutations.go:173-175`). Correct: this is CR 704.5b.
2. **`LoseTheGameForEffect`** (`effect_api.go:804-825`), the only
   engine verb for "you lose the game". Its comment says CR 104.3
   requires the loss to wait for the next SBA check. Its one catalog
   caller is the Pact cycle's decline branch (`pactPayment`,
   `cards/effects/pact_of_negation.go:66-79`), whose comment repeats the
   claim (`:33-35`).
3. **`MillToZoneForEffect`** (`effect_api.go:920-926`), when a mill or a
   top-of-library exile runs the library out. This is #767. A probe on
   `684f2786` (a three-seat game, `MillNForEffect` for 93 cards against a
   92-card library) left the player eliminated after
   `runStateChecksLocked`.

### What the rules say

From the Aug 7, 2026 Comprehensive Rules:

- **CR 104.1.** "A game ends immediately when a player wins, when the game
  is a draw, or when the game is restarted."
- **CR 104.2a.** A player still in the game wins when all their opponents
  have left. "This happens immediately and overrides all effects that
  would preclude that player from winning the game."
- **CR 104.2b.** "An effect may state that a player wins the game."
- **CR 104.3a.** A player can concede at any time, leaves the game
  immediately, and loses.
- **CR 104.3b, c, d, j.** Life, the empty-library draw, poison and
  commander damage each say "the next time a player would receive
  priority. (This is a state-based action.)"
- **CR 104.3e.** "An effect may state that a player loses the game." This
  sentence has **no** state-based-action clause.
- **CR 104.3f.** "If a player would both win and lose the game
  simultaneously, that player loses the game."
- **CR 104.3h.** Under the limited range of influence option, an effect
  win makes the player's opponents within range lose instead. The engine
  has no range of influence option.
- **CR 104.4a.** "If all the players remaining in a game lose
  simultaneously, the game is a draw."
- **CR 104.5.** A player who loses leaves the game; CR 800.4 covers what
  leaving does.
- **CR 614.17 and 614.17a.** "Can't" effects are not replacement effects.
  They have to exist before the event and can't reach back in time.
- **CR 614.17c.** An event that can't happen can only be replaced by a
  self-replacement effect.
- **CR 701.17b.** A player mills as many cards as possible. Only a draw
  loses (CR 704.5b).
- **CR 704.3.** All applicable state-based actions are performed
  simultaneously, as a single event.
- **CR 810.8a.** In Two-Headed Giant, "can't win" and "can't lose" apply
  to the team. The engine has no team format (`game.go:1019-1020`).

The Scryfall rulings settle the cases the CR leaves open:

- **Platinum Angel** (2009-10-01): "No game effect can cause you to lose
  the game or cause any opponent to win the game while you control
  Platinum Angel. It doesn't matter whether you have 0 or less life,
  you're forced to draw a card while your library is empty, you have ten
  or more poison counters, you're dealt combat damage by Phage the
  Untouchable, …" A concession still loses (2004-12-01), and "Effects
  that say the game is a draw … are not affected".
- **Abyssal Persecutor** (2017-11-17): if the Persecutor leaves while an
  opponent is at 0 or less life, "that opponent will lose the game as a
  state-based action. No player can respond".
- **Phyrexian Unlife** (2011-06-01): it stops only the life loss, and "If
  you're at 0 or less life and Phyrexian Unlife leaves the battlefield,
  you'll lose the game".
- **Laboratory Maniac** (2021-03-19) and **Jace, Wielder of Mysteries**
  (2019-05-03): "If for some reason you can't win the game … you won't
  lose for having tried to draw a card from a library with no cards in
  it. The draw was still replaced."
- **Jace, Wielder of Mysteries** (2019-05-03), the −8: "you'll draw as
  many cards as you can and then win the game before state-based actions
  would cause you to lose the game for trying to draw from an empty
  library." An effect win happens during resolution, not at the next SBA
  check.
- **Everybody Lives!** (2023-10-13): "Players can always concede from the
  game, even if they can't otherwise lose the game."
- **Lich's Mastery** (2018-04-27): "While you can't lose the game, your
  opponents can still win the game if an effect says so."
- **Angel's Grace, Abyssal Persecutor, Herald of Eternal Dawn:** a player
  who can't lose still can't pay more life than they have. The engine
  already refuses that payment (`life_tail.go:251`, `activated.go:499`,
  `alternative_cost.go:379`, `mutations.go:3448`).

### So the current comment is wrong

`LoseTheGameForEffect`'s deferral is not what CR 104.3 says. Only the four
state-based losses wait for priority. An effect loss (104.3e), like an
effect win (104.2b), happens when the effect says so. The deferral is
observable. In multiplayer, the rest of the resolution still treats the
loser as a player in the game. It also cannot be gated correctly. A
"can't lose" that exists when the effect resolves but is gone by the next
SBA check would not stop the loss, which is backwards (CR 614.17a).

### The cards

The audit on #749 counts 23 cards where this is the only core blocker and
29 where it is one of several. The batch issues name 30 cards (#296–#463).
Their oracle text, taken from the Scryfall dump, falls into these shapes:

| shape | cards |
|---|---|
| upkeep "if …, you win the game" (intervening if, CR 603.4) | Felidar Sovereign, Test of Endurance, Revel in Riches, Simic Ascendancy, Helix Pinnacle, Liliana's Contract, Triskaidekaphile, Knuckles the Echidna, Mechanized Production, Call the Spirit Dragons |
| draw replacement: "win the game instead" | Laboratory Maniac, Jace, Wielder of Mysteries (static and −8) |
| resolution check | Thassa's Oracle, Approach of the Second Sun, Maze's End |
| attack trigger | Twenty-Toed Toad |
| activated | Halo Fountain |
| "whenever / when a player loses the game, … you win" | Ramses, Assassin Lord; Zenos yae Galvus // Shinryu |
| state trigger (CR 603.8) | Darksteel Reactor |
| static "you can't lose and your opponents can't win" | Platinum Angel, Herald of Eternal Dawn |
| granted static of that shape | Cloudsteel Kirin, The Book of Exalted Deeds |
| "this turn" gate from a spell | Angel's Grace, Everybody Lives! |
| "you don't lose for having 0 or less life" | Pact Weapon |
| "would lose the game" replacement | The Golden Throne |
| effect loss only (misattributed: already possible) | Strixhaven Stadium, Summon: Primal Odin |

Cards outside the audit that have the same shapes, and that the design has
to cover: Abyssal Persecutor ("You can't win the game and your opponents
can't lose the game"), Lich's Mastery ("You can't lose the game" and "When
Lich's Mastery leaves the battlefield, you lose the game"), Phyrexian Unlife
and Transcendence ("You don't lose the game for having 0 or less life"),
and Gideon of the Trials' emblem ("As long as you control a Gideon
planeswalker, you can't lose the game and your opponents can't win the
game").

## Decision 1 — One loss path, with a cause; concede is never gated

Every loss names why it happened:

```go
// game/game_end.go
type LossCause string

const (
    LossLife            LossCause = "life"             // CR 704.5a / 104.3b
    LossEmptyDraw       LossCause = "empty_draw"       // CR 704.5b / 104.3c
    LossPoison          LossCause = "poison"           // CR 704.5c / 104.3d
    LossCommanderDamage LossCause = "commander_damage" // CR 704.6c / 903.10a
    LossEffect          LossCause = "effect"           // CR 104.3e
    LossConcede         LossCause = "concede"          // CR 104.3a
)
```

`eliminatePlayerLocked` is split into three functions with one job each:

1. **`loseGameLocked(p *Player, cause LossCause, source uuid.UUID) bool`**
   checks the gate (Decision 4) unless `cause == LossConcede`. If the
   player can't lose, it returns false and nothing happens. Otherwise it
   calls `leaveGameLocked`. It does **not** check whether the game is
   over. Callers do that once per batch (Decision 2 and Decision 3).
2. **`leaveGameLocked(p *Player, cause LossCause)`** is today's elimination
   body: `Eliminated`, stack and prompt cleanup, and
   `EventPlayerEliminated`, which gains `Label = string(cause)`. #769
   grows this function with CR 800.4a's object removal. This ADR adds no
   object handling to it.
3. **`checkGameOverLocked()`** replaces the `survivors <= 1` tail. See
   Decision 5.

The turn-cursor rotation (`advancePastEliminatedLocked`) runs once after
a batch, not once per player inside the SBA loop, and only if the game
did not end in that batch. What that rotation does is
[ADR 0059](0059-turn-machinery.md) Decision 6's (#766): when the active
seat has left, it ends the rest of the turn and begins the next one.
Because beginning a turn runs step entry hooks, the rotation never runs
inside a resolving callback (Decision 3).

`Concede` calls `leaveGameLocked(p, LossConcede)` and then
`checkGameOverLocked()`. It never reads a gate: CR 104.3a, and the Platinum
Angel and Everybody Lives! rulings. `EventConcede` stays for existing
listeners, but the log projects only `EventPlayerEliminated`. That fixes
a double line found while verifying this ADR: on `684f2786`, one concede
logs both "P1 conceded" and "P1 was eliminated"
(`protocol/log.go:458-465` maps both events to `LogEliminated`).

## Decision 2 — The SBA loss pass is collect-then-apply, and the empty-draw flag is cleared on every pass

CR 704.3 performs every state-based action at once. The loop at
`mutations.go:2278-2302` eliminates players one at a time, and each
elimination rotates the turn cursor and re-counts survivors mid-loop.
It gives the right answer today only because nothing reads the winner.
Once the engine records one (Decision 5), a player counted as the last
survivor could be eliminated a line later in the same pass.

The new pass:

```go
var losers []lossEntry // {player, cause}
for _, p := range g.Seats {
    if p.Eliminated { continue }
    drew := p.AttemptedEmptyDraw
    p.AttemptedEmptyDraw = false // CR 704.5b: "since the last time SBAs were checked"
    switch {
    case p.Life <= 0 && g.canLoseLocked(p, LossLife):
        losers = append(losers, lossEntry{p, LossLife})
    case drew && g.canLoseLocked(p, LossEmptyDraw):
        losers = append(losers, lossEntry{p, LossEmptyDraw})
    case poisonOf(p) >= PoisonLethal && g.canLoseLocked(p, LossPoison):
        losers = append(losers, lossEntry{p, LossPoison})
    case p.IsDeadByCommanderDamage() && g.canLoseLocked(p, LossCommanderDamage):
        losers = append(losers, lossEntry{p, LossCommanderDamage})
    }
}
for _, l := range losers { g.leaveGameLocked(l.p, l.cause) }
moveOn := false
if len(losers) > 0 || g.ActiveSeatLeftPending { // Decision 3's deferred departure
    g.ActiveSeatLeftPending = false
    moveOn = !g.checkGameOverLocked()
    fired = true
}
// ... the permanent SBAs (704.5f-j, battles, sagas, legend rule) ...
// Return moveOn to the settling loop. Accumulate it across repeated SBA
// passes; rotate once they are quiet, before draining waiting triggers.
```

- **The rotation waits for repeated SBA passes to settle.** Ending
  the active player's turn runs the cleanup sweep, which removes marked
  damage and deathtouch marks. The destruction SBAs read those marks,
  including on later checks: a lord dying on one pass can make another
  creature's damage lethal on the next. `runStateChecksLocked` therefore
  repeats the checks before rotating, then checks the new turn's board
  before draining the waiting triggers (#834 review; pinned by
  `TestActiveSeatSBALossStillDestroysMarkedCreatures` and
  `TestActiveSeatSBALossSettlesChainedLethalDamageBeforeCleanup`).

- **Every gate is read at the check**, against the battlefield as it is
  then. A player at 0 life whose Platinum Angel dies loses at the next
  check, with no window to respond, which is the Abyssal Persecutor
  ruling. Life, poison and commander damage are conditions the check
  reads again every time, so nothing has to be stored for them.
- **The empty-draw flag is cleared on every pass, gated or not.** CR
  704.5b looks only at draws "since the last time state-based actions
  were checked". A flag left set under a Platinum Angel would kill the
  player the moment the Angel left, long after the draw. Today the flag
  is cleared only on elimination (`mutations.go:2484`). That was
  equivalent while nothing could stop the loss, and it is wrong once
  something can.
- **The first matching cause is the one recorded.** The order is life,
  empty draw, poison, commander damage, which is the CR 704.5 order. A
  player who is both at 0 life and at 10 poison while a Phyrexian Unlife
  (life only) is out falls through to poison and loses to it.
- **The flag is renamed `AttemptedEmptyDraw`**, and the JSON tag
  `losesAtNextSba` is kept (`snapshot.go:229`), so no schema bump is
  needed. The old name invited #767: it described a consequence, not the
  one fact CR 704.5b reads. `snapshot_drift_test.go:247` and
  `clone.go` follow the rename. After sub-PR 1 and Decision 3 the only
  writer is `actuallyDrawCardLocked`.

## Decision 3 — Effect wins and effect losses are immediate

CR 104.3e and 104.2b have no state-based-action clause, CR 104.1 ends the
game "immediately", and the Jace ruling has an effect win land before the
SBA check. So:

```go
// LoseTheGameForEffect: CR 104.3e. Reports whether the player lost.
// A player who can't lose (Decision 4) does not, and nothing is
// remembered for later (CR 614.17a).
func (g *Game) LoseTheGameForEffect(player, source uuid.UUID) (lost bool, err error)

// WinTheGameForEffect: CR 104.2b. Reports whether the game ended with
// this player as the winner. A player who can't win does not, and
// nothing is remembered.
func (g *Game) WinTheGameForEffect(player, source uuid.UUID) (won bool, err error)
```

- **`LoseTheGameForEffect`** calls `loseGameLocked(p, LossEffect,
  source)` right away. The player is `Eliminated`, and their stack items
  and prompts are cleaned up (`leaveGameLocked`), before the call
  returns. It stops writing the flag. Its only caller today,
  `pactPayment`, passes the Pact as `source`. The new signature exists
  for the source, which the log names.
- **The rotation and the game-over check depend on who left.**
  - **The loser is not the active seat.** `checkGameOverLocked()` and,
    if the game goes on, the rotation run right away. The rotation's
    only work here is its priority-holder branch.
  - **The loser is the active seat.** Both are **deferred**.
    `LoseTheGameForEffect` sets `Game.ActiveSeatLeftPending`, and the
    next SBA loss pass consumes it (Decision 2): the resolution
    bookend's sweep around `resolveTopOfStackLocked`, or the sweep at a
    later prompt answer, whichever comes first. The pass runs
    `checkGameOverLocked()` and, if the game goes on, the rotation.
    Only the stack and prompt cleanup and `ErrStopResolution` happen at
    once. The reason is [ADR 0059](0059-turn-machinery.md) Decision 6:
    with the active seat gone, the rotation ends the turn and begins the
    next one (untap, per-turn resets, step entry hooks, upkeep
    triggers), and none of that may run inside a callback that is still
    resolving. Until the pass, the rest of the resolution sees a turn
    whose active player has left, which is CR 800.4j for the length of
    one resolution. The flag is plain data: `Clone` copies it, the
    snapshot carries it (`activeSeatLeftPending`, omitempty), and the
    drift test marks it `carried`.
  - The shapes that reach the deferred branch: Final Fortune and Last
    Chance's end-step loss (ADR 0059's first wave), the Pact cycle's
    decline (always in its controller's upkeep), and Lich's Mastery's
    leaves-the-battlefield trigger on its controller's turn.
- **`WinTheGameForEffect`** checks `canWinLocked(p)` (Decision 4). If the
  player can't win, it emits `EventWinPrevented{Actor, Source}` and
  returns false. Otherwise it calls `endGameLocked` with
  `GameOutcome{Kind: "win", Winner: p, Cause: "effect", Source: source}`
  (Decision 5).
- **An empty-draw flag pending in the same resolution does not stop an
  effect win.** The flag is only read at the next check. That is the Jace
  −8 ruling. CR 104.3f is about a win and a loss that are simultaneous,
  and a flag that has not yet been checked is not a loss.
- **A draw replaced by "win instead" stays replaced when the win is
  prevented.** Laboratory Maniac and Jace cancel the draw in their
  `RepEventDraw` replacement and then call `WinTheGameForEffect`. If the
  win is prevented, the draw is still cancelled, so no flag is set and
  the player does not lose (the 2019 and 2021 rulings). This needs no
  engine code, only the order in the card helper.

### What happens to the rest of the resolution

- **The game ended (a win, or a loss that left one player or none).**
  For an effect win or a non-active player's loss this is at once. For
  the active seat's loss it is at the next SBA pass (above), and the
  resolution has already stopped if that seat controlled it. Every later
  mutation in that resolution is moot. `endGameLocked`
  clears `PendingChoices`, `PendingTriggers` and `DiscardPending`, so no
  prompt is left open over the game-over banner. The public mutators and
  choice resolvers already refuse a game that is not active
  (`ErrGameNotActive`; for example `pending_choice.go:744`,
  `chained_choice.go:314`, `SpendUndo` at `mutations.go:5583`).
- **The loser controls the resolving spell or ability, and the game
  goes on** (multiplayer: Pact of Negation's decline, Lich's Mastery's
  leaves-the-battlefield trigger). CR 800.4a says that player's objects
  on the stack leave the game, and the resolving object is one of them.
  So the resolution stops.
- **Anyone else loses** (Strixhaven Stadium's "that player loses the
  game"). The resolution continues. The departed player is already out
  of `Context.Opponents()` (`cards/effects/context.go:77-87`) and of
  every other seat walk that skips `Eliminated`.

How it stops: the catalog helpers `effects.WinTheGame{Player}` and
`effects.LoseTheGame{Player}` return a new sentinel,
`game.ErrStopResolution`, when the game ended or when the loser is
`ctx.Controller()`. The sites that run a catalog callback and turn its
error into `EventEffectError` treat that sentinel as a clean stop, not
an error. They are `fireEffectResolverLocked` (`effect_hooks.go:437-447`),
`resolveTopAbilityLocked` (`mutations.go:1844-1851`), the chained-choice
branches (`chained_choice.go:342`, `:394`) and the pending-choice
continuations (`pending_choice.go:1007`, `:1209`, `:1268`). They share one
helper, `reportEffectErrorLocked(ev, err)`. The resolving card still
routes out of the stack as it does today. Where it goes when its owner
has left is #769's to decide.

**#808 must land first.** Today a player can leave mid-resolution only by
conceding. After this decision, any effect loss inside a prompt
continuation can do it. #808's two fixes (a dropped life-change
continuation still runs its tail; an eliminated player is skipped in the
drain legs) then cover a path cards reach on their own.

## Decision 4 — "Can't lose" and "can't win" are derived gates, read where they are checked

```go
// game/game_end.go
type GateScope uint8

const (
    GateYou       GateScope = iota + 1 // the source's controller
    GateOpponents                      // each opponent of the controller
    GateEachPlayer
)

type GameEndGate struct {
    Scope    GateScope
    CantLose bool
    CantWin  bool
    // Causes narrows CantLose. Nil means every cause except
    // LossConcede. Phyrexian Unlife is []LossCause{LossLife}.
    Causes []LossCause
    // While is a static's condition, read at every check. Nil means
    // always. Gideon of the Trials' emblem ("as long as you control a
    // Gideon planeswalker") is the shape.
    While func(g *Game, source Card) bool
}
```

**Sources, in the order they are read:**

1. **Battlefield statics.** `Spec.GameEndGates []game.GameEndGate` is
   wired through a new `CatalogGameEndGates func(key string) []GameEndGate`
   hook in `carddef.go`, next to `CatalogNoMaxHandSize` (`:206-209`).
   It is read with `CatalogAbilityKey(c)`, so a Platinum Angel that has
   lost all its abilities stops working. "You" is the card's current
   `Controller`.
2. **Turn-scoped gates** from a resolved spell (Angel's Grace, Everybody
   Lives!). They live in `Game.TurnScopedGameEndGates []ScopedGameEndGate`,
   a copy of the `TurnScopedReplacements` / `ScopedStatic` registry shape
   (`turn_scoped_statics.go`):

   ```go
   type ScopedGameEndGate struct {
       Gate       GameEndGate // While is always nil here
       You        uuid.UUID   // the spell's controller at resolution
       SourceID   uuid.UUID
       SourceName string
       Seq        int         // Turn.Seq it was registered on (ADR 0059 Decision 1)
   }
   func (g *Game) RegisterTurnScopedGameEndGateForEffect(gate GameEndGate, you, sourceID uuid.UUID)
   ```

   `ClearExpiredTurnScopedGameEndGatesLocked` runs in the cleanup body
   next to `ClearExpiredTurnScopedStaticsLocked` (`game.go:1077`). "This
   turn" ends in cleanup (CR 514.2). The entries are plain data, so
   `Clone` copies the slice, the snapshot carries it, and the drift test
   marks the field `carried`. Registration refuses a gate with a
   non-nil `While`.

   [ADR 0059](0059-turn-machinery.md) Decision 6 moves the cleanup body
   into `sweepTurnEndLocked` and names this sweep in it, so the gates
   also end when the active player leaves or `PassTurn` skips cleanup.

   **The stamp is `Seq`, named as ADR 0059 names turn identity.** If this
   sub-PR lands before ADR 0059's sub-PR 1, it is stamped from
   `Turn.Number` under the name `Seq`, and ADR 0059 sub-PR 1 re-stamps it
   with its other readers (its Decision 2 table) and converts it when an
   older snapshot is restored (its Decision 10 list).

   **#755 is expected to fold this registry in.** #755's "until your next
   turn" durations replace `ScopedStatic`'s expiry stamp (ADR 0059
   Decision 2), and this registry is the same shape. When #755 lands,
   its duration replaces `Seq` here too, and the sweep becomes that
   duration's check. Until then the registry only knows "this turn".

3. **Emblems**, when #623 lands. The reader is one loop, and #623 adds
   a third source to it.

**Built with its first card.** Owner policy is that every new seam path
ships with at least one real card. No first-wave card (question 3)
narrows `Causes`, sets `While`, or uses `GateEachPlayer`. Sub-PR 3 builds
whole gates only (every cause except concede, scopes `GateYou` and
`GateOpponents`), and the type keeps the shape above so the first card
that needs each part adds it: `Causes` with Phyrexian Unlife (after ADR
0056's "as though its source had infect" item), Pact Weapon (#660) or
Transcendence; `While` with Gideon of the Trials' emblem (#623);
`GateEachPlayer` with Everybody Lives!.

**Queries.** `canLoseLocked(p, cause) bool` and `canWinLocked(p) bool`
walk the three sources. A gate applies to `p` when its scope includes
`p`. For a loss it must also name `cause` (or have nil `Causes`), and its
`While` must be nil or true. Concede never reaches `canLoseLocked`.
Exported read-only forms, `CantLoseCauses(playerID) []LossCause` and
`CantWin(playerID) bool`, feed the view (Decision 7) and bots.

**Why derived rather than stored.** It is the same argument as
`CatalogNoMaxHandSize` (`effect_hooks.go:368-396`). Two Platinum Angels
where one leaves, an Angel whose abilities are removed, and an Angel that
changes controller all come out right with no bookkeeping. Undo needs
nothing beyond the turn-scoped slice. The walk is linear in the
battlefield and runs once per player per SBA pass, which is the same cost
the hand-size check pays once per cleanup.

**Why a registry and not player flags for Angel's Grace.** #749 suggested
flags. "Your opponents can't win this turn" is relative to the caster, so
a flag would need one entry per opponent, written at resolution and
wrong for a player who joins no one's side later. The registry stores the
gate once with its `You`, and the reader resolves it. The sweep is also
already written.

**Granted gates** (Cloudsteel Kirin's "Equipped creature has …", The Book
of Exalted Deeds' "It gains …") need a way to grant a non-keyword static
to a permanent. The engine has no such way (`GrantToAttached` grants
keywords), so these cards stay blocked. The reader gains a fourth source
when that seam exists. This ADR does not build it.

**CR 614.17c ordering, for later.** A "would lose the game" replacement
(The Golden Throne) is out of scope (Decision 8). When it is built, the
gate is checked first, because an event that can't happen can't be
replaced.

## Decision 5 — The outcome is engine state: who won, or a draw, and why

```go
type GameOutcome struct {
    Kind   string    // "win" | "draw"
    Winner uuid.UUID // Kind == "win"
    Cause  string    // "last_standing" | "effect" | "all_lost"
    Source uuid.UUID // Cause == "effect": the object whose effect won
}

// Game.Outcome is nil while the game is active, and nil for a game
// ended by End() with no result (an abandoned table).
Outcome *GameOutcome
```

**`endGameLocked(o GameOutcome)`** sets `State = StateEnded` and
`Outcome`, clears the prompt queues (Decision 3), and emits
`EventGameOver{Actor: Winner, Source, Label: Cause}`. It is the only
writer of `StateEnded` apart from `Game.End()`.

**`checkGameOverLocked() (ended bool)`** runs after every batch of
departures: an SBA pass, an effect loss (at once, or at the next pass
when the active seat left mid-resolution; Decision 3), or a concede. It
reports whether the game ended, and the rotation runs only when it
didn't.

1. If no player is left, the game is a draw: `{Kind: "draw", Cause:
   "all_lost"}` (CR 104.4a).
2. If exactly one player is left, that player wins: `{Kind: "win",
   Cause: "last_standing"}`. **No gate is read here.** CR 104.2a
   overrides "can't win", so Abyssal Persecutor's controller wins when
   the opponent it kept alive at −5 life concedes.
3. Otherwise the game goes on.

**CR 104.3f, stated order.** Losses in one batch are applied before the
last-standing check. So a player who would be the last one standing in
the same pass that makes them lose is a loser. That gives a draw when
everyone left loses at once, and never a win. No engine path produces an
effect win and a loss that are simultaneous. An effect win happens during
resolution, and state-based losses happen only at a check. If a future
card does produce both at once, the loss is applied first.

**Multiplayer.** A win ends the game for everyone (CR 104.1). The other
players are **not** marked `Eliminated`, and no `EventPlayerEliminated`
fires for them: they did not leave the game, the game ended. "Whenever a
player loses the game" triggers (Ramses) have nothing to fire into. The
issue's reading that "every opponent loses at once" is CR 104.3h, the
range of influence option, which the engine does not offer.

**Clone and snapshot.** `Outcome` is carried by `cloneLocked` /
`RestoreFrom` and by the snapshot (`outcome`, omitempty, so no schema
bump). The drift test gains the row. A restore point is removed once the
game ends (`ws/room.go:536-541`), so the snapshot field matters for
fixtures and forensics rather than restarts. **Undo can't reach back
past the end** (decided, question 1): once `State` is `StateEnded`, the
room refuses `undo` from every caller except the admin.

## Decision 6 — Bots and the legal enumerator

- **Enumerator.** Nothing new. Winning and losing add no prompts and no
  moves.
- **The view is enough.** Bots read only `GameView` (ADR 0033 §3), so the
  gate reaches them through the seat fields in Decision 7.
- **Heuristic** (`aiseat/heuristic`):
  - **Each clock is read against its own cause.** Life reads `"life"`,
    [ADR 0056](0056-infect-wither-toxic.md)'s poison clock (its
    Decision 7) reads `"poison"`, and commander damage reads
    `"commander_damage"`. A seat whose `cant_lose` names a clock's cause
    can't be finished on that clock, and every reader below skips or
    clamps that clock only.
  - `ShouldConcede` (`concede.go:66`) returns false while the bot's own
    seat has `cant_lose` including `"life"`, `"poison"` or
    `"commander_damage"` for the clock that makes it look hopeless.
    Today `hopeless` is judged on life alone. A bot at −10 behind a
    Lich's Mastery is not hopeless, and conceding is the one loss its
    gate can't stop.
  - `lethalPush` (`combat.go:197`) and the finish-life targeting
    (`moves.go:321`) skip a seat whose `cant_lose` includes `"life"`,
    because damage can't finish it. ADR 0056's poison push skips a seat
    with `"poison"`. The heuristic has no commander-damage push today.
    One added later skips `"commander_damage"`.
  - `lifeDanger` (`score.go:379-385`) is clamped at 0 life for a seat
    with `"life"`, and ADR 0056's `PoisonDanger` is clamped at 9 poison
    for a seat with `"poison"`. Otherwise the squared penalty keeps
    growing on a player the penalty can't hurt, and the bot plays as if
    it were losing.
- **No new threat weighting.** A gate source is not scored as a priority
  removal target in this ADR. The catalog soak (#601) reports stalls
  rather than failing on them, so a table stuck behind a Platinum Angel
  shows up there, and weighting can follow if it does.
- **Model tier.** `aiseat/model/prompt.go:169` adds "can't lose" and
  "can't win" to the seat line.

## Decision 7 — The wire, the log and ADR 0051

**`GameView.outcome`** (omitempty):

```json
{ "kind": "win", "winner": "<player id>", "winner_seat": 2,
  "cause": "effect", "source": "<card instance id>", "source_name": "Felidar Sovereign" }
```

A draw has `kind: "draw"` and no winner. `source_name` is resolved the
way the log resolves card names, with the same `known_by` treatment. A
winning source is public in every shape in the Context table.

**`PlayerView`** gains, all omitempty:

- `cant_lose: string[]`, the causes that currently can't make this
  player lose. All five non-concede causes when fully gated.
- `cant_win: bool`.
- `end_gates: [{ source?, source_name, cant_lose?, cant_win?, this_turn? }]`,
  the gates that name this player, for the seat tooltip and the model
  prompt.

These are derived from public state (battlefield statics and resolved
spells), so they are not redacted.

**Log.**

- `LogEliminated` gains `cause`. Its text says why: "Alice lost the game
  (0 or less life)", "… (drew from an empty library)", "… (10 poison
  counters)", "… (21 commander damage)", "… (Pact of Negation)", "Alice
  conceded". `Amount` keeps its concede bit for older clients. This is
  the only place an elimination's reason is written. ADR 0056 adds no
  text to it, and the poison cause is `"poison"` (`LossPoison`) in both
  ADRs.
- A new `LogGameOver` ("game_over") entry: "Alice won the game (Felidar
  Sovereign)", "Alice won the game: every opponent has left", "The game
  is a draw".
- A new `LogWinPrevented` ("win_prevented") entry: "Alice would have won
  the game (Laboratory Maniac), but can't (Platinum Angel)". It is logged
  once per prevented win.
- **A loss stopped by a gate at an SBA check logs nothing.** Otherwise it
  would log on every pass while a player sits at 0 life. The seat's
  `cant_lose` shows the state instead.

Final wording is settled in sub-PR 2's review. `docs/protocol.md` is
updated: the `concede` row (`:231`) loses "implicit winner", and GameView,
PlayerView and the log kinds gain the new fields. `client/src/lib/protocol.ts`
mirrors the types.

**The client** reads the winner from `view.outcome`. The survivors
derivation (`Game.svelte:480`) stays only as the fallback for a view
with `state: "ended"` and no `outcome`: JSONL replays written before this
change, and `Game.End()`. The win and loss sounds follow `outcome`. A
spectator hears neither, as today.

**What the board shows** (decided, question 2, option (b)):

- **The log** as above.
- **A cause on the game-over banner**, from `outcome`: "Alice wins the
  game — Felidar Sovereign" for an effect win, "Alice wins the game" for
  the last player standing, "The game is a draw" for `all_lost`.
- **A "can't lose" / "can't win" badge on a gated seat**, from
  `cant_lose` and `cant_win`. Its tooltip names the sources from
  `end_gates`.
- **No reveal-strip cue** for a prevented win (option (c) was not
  chosen). The `win_prevented` log line covers it. Owner policy: no
  reveal-strip cues for these events.

**ADR 0051.** The `games` row gains `outcome TEXT` (`win` | `draw` |
NULL) next to `winner_seat`. A NULL `winner_seat` otherwise can't tell a
draw from an abandoned table. Both are written when the room sees
`StateEnded`, from `Game.Outcome`. `winner_seat` is the winner's seat
index. A table archived or deleted while active keeps both NULL. ADR 0051
is not implemented yet, so this is a column in a schema that has not
shipped, not a migration.

## Decision 8 — Out of scope, stated

- **"Would lose the game" replacements** (The Golden Throne). A new
  replacement kind; Decision 4 fixes its order relative to the gate.
- **Draw effects** (CR 104.4c, Divine Intervention) and the **CR 104.4b
  loop draw**. ADR 0055 hands loops back to the players.
- **Teams** (Two-Headed Giant, CR 810.8a; Emperor, CR 104.2d) and the
  **range of influence option** (CR 104.3h). The engine has neither.
- **Granting a non-keyword static** (Cloudsteel Kirin, The Book of
  Exalted Deeds). See Decision 4.
- **Emblem gates** (Gideon of the Trials). Waits on #623.
- **What leaving the game does to objects** (CR 800.4a) is #769.
  **The turn after the active player leaves** is
  [ADR 0059](0059-turn-machinery.md) Decision 6 (#766). This ADR only
  fixes when the rotation may run (Decision 3).
- **Restarting the game** (CR 104.6, Karn Liberated).
- **Angel's Grace's second sentence** ("damage that would reduce your
  life total to less than 1 reduces it to 1 instead") is a
  damage-result replacement that keeps lifelink whole (the 2021-03-19
  ruling). That is a separate seam. Angel's Grace ships without it,
  declared incomplete (question 3).

## Consequences

- **An effect loss happens when the effect says so, not at the next
  check.** This is observable for the Pact cycle today: in a multiplayer
  game the decline removes the player at once, and the resolution stops
  (Decision 3). The comments at `effect_api.go:804-816` and
  `pact_of_negation.go:33-35` are rewritten.
- **A mill never loses the game** (sub-PR 1, #767). The card-side clamps
  `b31MillAtMost` (`batch31_helpers.go:154-168`), `b35MillAtMostCollect`
  (`batch35_helpers.go:222-243`) and Tasha's Hideous Laughter's bound
  (`batch27_helpers.go:340-345`) become plain calls, and their "the engine
  flags the loss" comments go. So do the comments at `effect_api.go:827-831`,
  `:854-857` and `:920-923`, `primitives.go:113`,
  `glimpse_the_unthinkable.go:7` and `nephalia_drownyard.go:15-18`. The
  engine has no mill *cost* component, so CR 701.17b's "can't pay"
  clause has no engine caller. Two catalog cards print a mill cost:
  Millikin (`{T}, Mill a card`) allows the activation only while the
  library isn't empty, and The Warring Triad declares the gap. The mill
  code comment says so.
- **The engine names a winner.** The client, `docs/protocol.md` and ADR
  0051 stop inferring one. A four-player game won by Felidar Sovereign
  ends with three seats still standing, and the banner names the right
  player.
- **One concede is one log line.**
- **The end is final.** Once a game has ended, only the admin can undo.
  A player who passed priority into an effect win can't take it back.
- **An active player who loses by effect mid-resolution leaves at once,
  but the turn moves on at the next SBA pass.** The rest of that
  resolution runs on a turn with no active player.
- **A player can sit at 0 or less life indefinitely** behind a gate. The
  engine's life, poison and commander damage values are unbounded
  integers already, and the life-payment refusals already exist.
- **Bot games can last longer.** A bot that can't lose and can't win
  either (Everybody Lives!, only for a turn) or a Platinum Angel nobody
  removes can hold a soak game until its turn cap. The soak reports
  that as a stall.
- **`docs/engine-seams.md`**: the "Win the game by effect" row moves to
  Closed when sub-PR 3 lands.
- **`AGENTS.md`**: the Spec field list gains `GameEndGates`, and the
  catalog guidance gains one rule: a card that wins or loses the game
  calls `effects.WinTheGame` / `effects.LoseTheGame` and returns their
  error. It never calls the rotation or the game-over check itself.

## Alternatives considered

- **Keep effect losses on the SBA flag.** Rejected. CR 104.3e has no
  SBA clause, the Jace ruling puts an effect win before the check, and a
  deferred loss reads the gate at the wrong time (CR 614.17a).
- **Make effect wins wait for the SBA check too, for symmetry.**
  Rejected for the same ruling. A Jace −8 that draws the library out
  would lose to its own draws before it could win.
- **Model a win as "every opponent loses".** Rejected. That is CR 104.3h
  (range of influence). It would fire "whenever a player loses the game"
  triggers into a finished game and print one elimination per opponent.
- **"Can't lose" as a "would lose the game" replacement effect.** Rejected.
  CR 614.17 says "can't" is not a replacement. The difference shows when
  The Golden Throne and a Platinum Angel are both out: the Throne must
  not be exiled.
- **Store gate state on `Player`, written on enter and cleared on leave.**
  Rejected for the reasons `CatalogNoMaxHandSize` gives: two sources, a
  controller change and ability removal all need bookkeeping that a
  derived read doesn't.
- **Player flags for Angel's Grace.** Rejected for the turn-scoped
  registry (Decision 4).
- **Keep "last seat standing" on the client and add only a draw flag.**
  Rejected. An effect win leaves several seats standing, so the client
  can't infer the winner.
- **Gate inside the existing one-at-a-time SBA loop.** Rejected. Once a
  winner is recorded, eliminating mid-loop can name a winner who loses a
  line later in the same pass (CR 704.3, CR 104.3f).
- **Continue the resolution after its controller loses.** Rejected under
  CR 800.4a: the resolving object belongs to a player who has left.

## PR split

**Sub-PR 1 — #767: a mill never loses the game.** Can land now, before
anything else here.
- Remove the flag write from `MillToZoneForEffect`, rename the flag to
  `AttemptedEmptyDraw` (the JSON tag stays), and fix every comment listed
  in Consequences.
- Reduce `b31MillAtMost`, `b35MillAtMostCollect` and Tasha's bound to
  plain calls. The existing "a mill never loses" assertions
  (`batch32_test.go:345`, `batch35_test.go:212,380,443`) must still pass
  without the clamps.
- Tests: #767's list.

**Sub-PR 2 — engine: loss causes, collect-then-apply, immediate effect
losses and wins, outcome, wire. After #808.**
- `game_end.go`: `LossCause`, `loseGameLocked`, `leaveGameLocked`,
  `checkGameOverLocked`, `endGameLocked`, `GameOutcome`,
  `WinTheGameForEffect`, the new `LoseTheGameForEffect` signature,
  `ErrStopResolution` and `reportEffectErrorLocked`.
- The SBA loss pass (Decision 2), with the flag cleared on every pass.
  `canLoseLocked` / `canWinLocked` return true until sub-PR 3.
- `EventGameOver`, `EventWinPrevented`, the `EventPlayerEliminated`
  label, and the concede log fix.
- `GameView.outcome`, the log kinds and their text, `docs/protocol.md`,
  `protocol.ts`, and the client's winner from `outcome`.
- Clone, snapshot and drift test for `Outcome`.
- `effects.WinTheGame` / `effects.LoseTheGame`, and `pactPayment` moved
  onto `LoseTheGame`.
- `ActiveSeatLeftPending`: set by an effect loss of the active seat,
  consumed by the SBA loss pass; clone, snapshot and drift-test row.
- The end is final (question 1): the room refuses `undo` once the game
  has ended, for every caller except the admin.

**Sub-PR 3 — gates.**
- `GameEndGate` (whole gates only; see "Built with its first card" in
  Decision 4), `Spec.GameEndGates`, the `CatalogGameEndGates` hook, the
  turn-scoped registry and its cleanup sweep, clone, snapshot and drift
  test.
- `PlayerView.cant_lose` / `cant_win` / `end_gates`, and the model
  prompt line.
- Heuristic changes (Decision 6).
- `docs/engine-seams.md`: move the row to Closed.

**Sub-PR 4 — client** (question 2, option (b)): the cause on the
game-over banner, and the "can't lose" / "can't win" seat badge with its
`end_gates` tooltip. No reveal-strip cue.

**Card PRs** (question 3, option (c)): every card this seam alone
blocks, plus Angel's Grace declared incomplete. The list is in
[Decided (2026-09-17)](#decided-2026-09-17). Each path the engine PRs
build has a real card in the wave: effect wins (Felidar Sovereign and the
upkeep cycle), the draw-replacement win (Laboratory Maniac, Jace), an
effect loss of another player (Strixhaven Stadium), battlefield gates
(Platinum Angel, Herald of Eternal Dawn), "can't win" (Abyssal
Persecutor) and the turn-scoped registry (Angel's Grace). Each card
follows AGENTS.md (`Spec` slots, completeness declared, caveats weaker
than printed and never stronger). A card found to be blocked by
something else drops out of the wave rather than shipping stronger than
printed.

## Test plan

Engine (`internal/game`):

1. **#767.** A bounded mill past the library, an unbounded `Until` mill
   that never matches, and a `ZoneExile` run past the end each leave the
   library empty, the flag false, and the player in the game after
   `runStateChecksLocked`. The same player's next draw still loses
   (CR 704.5b).
2. **Felidar-style effect win, four seats.** `WinTheGameForEffect` for
   seat 2 during resolution: `State` is ended, `Outcome` is `{win, seat 2,
   effect, source}`, no seat is `Eliminated`, `PendingChoices` and
   `PendingTriggers` are empty, and one `EventGameOver` is emitted.
3. **Effect loss is immediate.** Three seats. `LoseTheGameForEffect` on a
   non-controller: `Eliminated` is true before any state check, the
   resolution continues, and `Context.Opponents()` no longer includes
   them. When the loser is the controller, the helper returns
   `ErrStopResolution`, the rest of the effect doesn't run, and no
   `EventEffectError` is emitted.

   **By the active seat, the rotation is deferred.** Three seats, the
   active player's own spell calls `LoseTheGame` on them: they are
   `Eliminated` and their stack items are gone at once, but the cursor
   still names their turn, no untap or upkeep trigger has run, and
   `ActiveSeatLeftPending` is set when the callback returns. The
   resolution bookend's sweep rotates once to the next seat's turn and
   clears the flag. In a two-seat game the same loss ends the game at
   that sweep with `last_standing`, and no next turn begins. The flag
   survives `Clone` and a snapshot round trip.
4. **Effect loss down to one player.** A two-seat `LoseTheGameForEffect`
   on the seat that is not active: the other seat wins at once with
   `Cause: "last_standing"`. (The active-seat case is in test 3.)
5. **Platinum Angel at 0 life.** Several state checks: the player stays
   in, and nothing is logged. The Angel is destroyed, and the player
   loses at the next check with `LossLife`.
6. **Empty-library draw under the Angel, then the Angel removed.** The
   player does not lose (CR 704.5b), and the flag is false after the
   first check.
7. **Poison and commander damage under the Angel.** No loss. (The
   `Causes: {LossLife}` case, where 0 life does not lose but 10 poison
   does, is tested by the first card that narrows `Causes`; see
   Decision 4.)
8. **Concede under the Angel** loses, logs one line, and ends a two-seat
   game with the opponent winning.
9. **Ability removal.** An Angel that has lost all abilities gates
   nothing.
10. **Laboratory Maniac against an opponent's Platinum Angel.** The draw
    is replaced, `EventWinPrevented` is emitted, the game goes on, and
    the Maniac's controller does not lose at the next check.
11. **Jace −8 ordering.** Draw seven from a three-card library (the flag
    is set), then `WinTheGameForEffect`: the controller wins.
12. **Abyssal Persecutor.** The controller can't win by effect, and an
    opponent at −5 stays in. When that opponent concedes, the controller
    wins (CR 104.2a overrides).
13. **Simultaneous losses.** Three seats, two at 0 life: one SBA pass
    eliminates both, and seat 3 wins (`last_standing`). Two seats, both
    at 0: a draw (`all_lost`) with no winner. The turn cursor rotates
    once per pass.
14. **Angel's Grace turn gate** (the real card, in its card PR; a
    registry unit test in sub-PR 3). Registered on a turn: the caster
    can't lose and the opponents can't win until cleanup. At the next
    turn's upkeep the gate is gone. It survives `Clone` / `RestoreFrom`
    and a snapshot round trip.
15. **Drift test:** `Outcome` and `TurnScopedGameEndGates` are `carried`,
    and the flag row is renamed.

Room (`internal/ws`):

16. **The end is final** (question 1). After a concede or an effect win
    ends the game, a player's `Undo` returns an error and the state stays
    ended. The admin's undo still succeeds. A probe on `684f2786` shows
    that today every undo succeeds and the game is active again.

Protocol:

17. `outcome` in the view for a win, a draw and `End()` (absent). The
    log has one `eliminated` line per concede, a `game_over` line, and a
    `win_prevented` line. `PlayerView.cant_lose` lists the causes.

Bots:

18. `ShouldConcede` is false for a hopeless-looking seat with
    `cant_lose` including `"life"`. `lethalPush` skips such a seat.
19. **Catalog soak** with the first-wave cards: no crash, and a stall
    under a gate is reported, not failed.

Client:

20. The winner and sounds come from `outcome`. An ended view with no
    `outcome` falls back to the survivors. The banner's cause text for a
    win by effect, a last-standing win and a draw, and the seat badge's
    label from `cant_lose` / `cant_win`, are pure helpers unit-tested
    with vitest. The banner and badge are checked by hand until #689.

## Decided (2026-09-17)

The owner answered all three open questions on 2026-09-17. The options
are kept as they were proposed. The chosen one is marked.

1. **Can a player undo past the end of the game?** When this was asked,
   they could. On
   `684f2786`, a room where one of two seats concedes and then calls
   `Undo` gets a nil error and an active game again. By then the restore
   point has already been deleted (`ws/room.go:536-541`), and bot
   runners stop when the game is not active (`aiseat/runner.go:236`).
   With effect wins, whoever took the last action (often the loser, by
   passing priority) owns the undo entry that ended the game.
   - (a) **The end is final.** The room refuses `undo` once the game has
     ended, for every caller except the admin. **Chosen.**
   - (b) Undo keeps working within the budget, and the room rewrites the
     restore point and relaunches bot runners after an undo that
     reactivates the game.

   **Decision: (a)**, as recommended. A game that has announced a
   winner, played the win sound and been written to the database should
   not come back. (b) is real work to support a take-back nobody asks
   for once the banner is up. The admin keeps undo for mistakes.
   Applied in Decision 5, sub-PR 2 and test 16.

2. **How much does the board show about wins, losses and gates?**
   - (a) The log only. The banner stays "Alice wins the game." and
     nothing marks a gated seat.
   - (b) The log, plus a cause on the game-over banner ("Alice wins the
     game — Felidar Sovereign", "The game is a draw"), plus a small
     "can't lose" / "can't win" badge on a gated seat. The badge's
     tooltip names the sources from `end_gates`. **Chosen.**
   - (c) (b), plus a reveal-strip cue when a win is prevented.

   **Decision: (b)**, as recommended. A player at −8 who is still in the
   game is confusing unless the seat says why, and a four-player game
   that ends with three players standing needs the banner to say how. A
   prevented win is rare, and the log line covers it. No reveal-strip
   cue. Applied in Decision 7 and sub-PR 4.

3. **Which cards ship in the first wave?**
   - (a) A reference set that exercises every path: Platinum Angel,
     Felidar Sovereign, Laboratory Maniac, Abyssal Persecutor, and
     Strixhaven Stadium (effect loss).
   - (b) Every card this seam alone blocks. From the audit: Felidar
     Sovereign, Test of Endurance, Revel in Riches, Simic Ascendancy,
     Helix Pinnacle, Liliana's Contract, Triskaidekaphile, Knuckles the
     Echidna, Mechanized Production, Call the Spirit Dragons, Laboratory
     Maniac, Jace, Wielder of Mysteries, Thassa's Oracle, Platinum Angel
     and Herald of Eternal Dawn. Also Strixhaven Stadium and Summon:
     Primal Odin, which the audit misattributed, and Abyssal Persecutor,
     which the audit did not list. Twenty-Toed Toad
     ships only if the card PR finds a way to set a numeric maximum hand
     size, and otherwise with that as a declared caveat.
   - (c) (b), plus Angel's Grace declared incomplete: it has "can't lose
     / can't win this turn" and split second, but not the life floor
     (Decision 8). **Chosen.**

   Excluded either way, because they are blocked elsewhere: Approach of
   the Second Sun (a whole-game record of spells cast by name), Maze's
   End (a return-to-hand cost), Halo Fountain (an untap cost), Darksteel
   Reactor (CR 603.8 state triggers), Cloudsteel Kirin and The Book of
   Exalted Deeds (granting a static), Ramses (attack history), Zenos
   (transform, #343), Pact Weapon (#660), Everybody Lives! (player
   hexproof and "can't lose life"), The Golden Throne (Decision 8), Lich's
   Mastery (its exile-for-each-life prompt), Phyrexian Unlife (damage "as
   though its source had infect", out of scope in
   [ADR 0056](0056-infect-wither-toxic.md) Decision 8) and Gideon of the
   Trials (#623).

   **Decision: (c)**, as recommended. Angel's Grace is the only card in
   the list that uses the turn-scoped gate registry. If it doesn't ship,
   that code ships with no real card behind it (the same argument ADR
   0054 made for Fiery Gambit). Its missing clause makes it weaker than
   printed, and AGENTS.md allows that when it is declared. Each card in
   (b) is checked for other blockers in its own PR, and one that turns
   out to be blocked drops out rather than shipping stronger than
   printed. This is also the owner's cross-cutting policy: every new
   seam path ships with at least one real card, even with a declared
   weaker caveat. Applied in Decision 4 ("Built with its first card")
   and the card PRs.
