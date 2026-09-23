# ADR 0085 — "Your life total can't change": a player-scoped replacement with a CR 611.2 duration

**Status:** Accepted · 2026-09-23 · S39 — The CR 614 replacement surface and its
paused continuations
**Issues:** [#1200](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1200)
("your life total can't change", and a replacement with a CR 611.2 duration),
under trackers [#882](https://github.com/krakenhavoc/cmd_and_ctrl/issues/882)
(the CR 614 replacement surface) and
[#883](https://github.com/krakenhavoc/cmd_and_ctrl/issues/883)
(protection, emblems, winning by effect). Filed out of
[#1197](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1197) (player
protection), whose Teferi's Protection shipped with this clause as a caveat, and
handed on by [ADR 0084](0084-phasing.md)'s closing section, which took the card's
other caveat and left this one a home.
**Numbering:** swept on 2026-09-23 with `git ls-tree origin/develop docs/decisions/`
and with `git ls-tree -r --name-only origin/<branch> docs/decisions` over all 406
remote heads. The highest number present anywhere was **0084**
(`0084-phasing.md`, on `develop`) and nothing held an `0085-*` file. 0005, 0024,
0029 and 0030 stay permanently unused per AGENTS.md §4.

**Related:** [ADR 0072](0072-protection.md) and its 2026-09-22 amendment
(`Player.Statics` — the registry this decision adds a payload to, and the
derived/granted split it copies), [ADR 0063](0063-durations-and-control.md) (the
`Duration` model, one `durationExpiredLocked`, and Decision 8's deferral of
durations on `TurnScopedReplacements`), [ADR 0013](0013-replacement-effects.md)
(the CR 614/616 pipeline, built-in replacements, `Preemptive`, and §5b's
cost-shaped life path), [ADR 0061](0061-token-creation-and-discard-are-replaceable-events.md)
(an instruction is the replaceable event, and the amount-replacement family
#1233 finished), [ADR 0066](0066-granted-cast-and-play-permissions.md)
(`CastPermission`, the other plain-data player-scoped registry with a duration),
[ADR 0084](0084-phasing.md) (Teferi's Protection's larger half, and the handoff
this ADR picks up), [ADR 0010](0010-card-effect-catalog.md) (the `Spec` slots the
proof cards declare through).

**A note on rule numbers.** In the pinned August 7, 2026 edition, **CR 119.7** is
"can't gain life" and **CR 119.8** is "can't lose life" — and CR 119.8's second
sentence is the one that makes this more than a pair of replacement effects:
*"If a cost or effect would cause a player who can't lose life to pay life, that
cost can't be paid and that effect does nothing."* **CR 119.4** is the ordinary
affordability rule (you may pay life only down to 0) and **CR 614.17b** is its
generalisation (*"if an event can't happen, a player can't choose to pay a cost
that includes that event"*). **CR 120.3a** is *"damage dealt to a player causes
that player to lose that much life"*, which is why damage needs a second consumer
rather than riding the life window. **CR 611.2b** is the duration, and
**CR 500.1** puts "your next turn" at the beginning of the untap step.

---

## Context

Teferi's Protection prints three clauses and the engine now has two of them:

> **Until your next turn, your life total can't change** and you gain protection
> from everything. All permanents you control phase out. Exile Teferi's
> Protection.

"You gain protection from everything" shipped with #1197 (ADR 0072's amendment),
which gave a **player** an ability slice for the first time. "All permanents you
control phase out" shipped with #1199 (ADR 0084). The card has carried exactly one
caveat since:

> "Your life total can't change" isn't implemented — damage is prevented by the
> protection, but life loss that isn't damage (a drain) still reaches you, and you
> can still gain life.

#1200 filed that gap as a REPLACEMENT-registry problem, and proposed the obvious
fix: put a `Duration` on `ReplacementEffect`, teach
`sweepTurnScopedReplacementsLocked` to run it through `durationExpiredLocked`
instead of clearing wholesale, and register the clause there. ADR 0063 Decision 8
had already stated the absence as deliberate — *"`TurnScopedReplacements` is still
cleared wholesale at cleanup and has no duration field; no card needs a
longer-lived replacement yet"* — and #1200 is the card that does.

ADR 0084's closing section, written while the other half of this same card was
being built, argued the opposite and this ADR agrees with it. The rest of this
document is that argument, plus the four consumers the rule turns out to have and
the two that were already written and waiting.

---

## Decision 1 — the entry is a `PlayerStatic`, not a `TurnScopedReplacements` entry

`Game.TurnScopedReplacements` holds `ReplacementEffect` values, and a
`ReplacementEffect` is **two closures** (`AppliesTo`, `Replace`). That is fine for
the Fog machinery it was built for, which lives for part of one turn and is
cleared at the cleanup step. It is not fine for this clause, and the reason is the
snapshot census, not taste:

`ContinuationCensus.TurnScopedReplacements` counts the registry, and
`snapshot_drift_test.go` classifies it `dropped` — a closure cannot be written to
disk and cannot be brought back. A game holding one is a game with **no restore
point**. Storing "until your next turn, your life total can't change" there would
mean that the one card famous for buying you a whole turn cycle is also the one
card that takes undo and persistence away from the table for that entire turn
cycle. That is a worse outcome than the caveat it replaces.

`Player.Statics` is the registry that already solved this for the *other* half of
the *same card*:

| | `TurnScopedReplacements` | `Player.Statics` |
|---|---|---|
| entry shape | two closures | plain data |
| duration | none (ADR 0063 D8) | `Duration`, CR 611.2 |
| swept by | wholesale clear at cleanup | `durationExpiredLocked` |
| clone | closure shared | value copy, fresh array |
| snapshot | `dropped`, censused | `carried` |
| scope | the game | one player |

And the last row is the one that decides it even without the census argument.
"Your life total can't change" is a statement about **a player**, in the same
grammatical position as "you have hexproof" and "you may cast spells as though
they had flash". It is not a statement about an event the game happens to be
processing. A registry keyed on the game would have to carry the player in a
closure; a registry keyed on the player already has it.

**So: `PlayerStatic` gains one more payload beside `Keyword` and `Timing`.** It is
the third thing that slice can say, and it is told apart from the other two by its
payload rather than by a discriminator field, exactly as #1195's timing statement
is (`player_statics.go`, `PlayerStatic.Timing`'s comment makes the argument at
length). One slice, one sweep, one `durationExpiredLocked` read, one clone, one
snapshot field.

```go
// LifeTotalLocked is "your life total can't change" (CR 119.7,
// CR 119.8) — the entry carries no Keyword and no Timing.
LifeTotalLocked bool `json:"lifeTotalLocked,omitempty"`
```

**Why not a `ScopedStatic`.** The same reason ADR 0072's amendment gives: a
`ScopedStatic` is adapted into a `ContinuousEffect` whose `Apply` signature is
`(*Characteristic, *Card)` — a characteristic of an OBJECT — and a player has
neither. The layer pass has nothing to do here.

**Why not a keyword token on the existing `Keyword` field.** It would be the
cheapest change in the diff and the most expensive one to live with:
`playerAbilityTokensLocked` is read by three rules (targeting, damage prevention,
attachment) and by `PlayerProtectedFromLocked`, all of which would then be
walking a token none of them can parse. `protection.go`'s closed grammar exists
precisely so that an unparseable token grants nothing; putting a non-ability in
the ability list to get it onto the wire for free is the kind of category error
that is invisible until something starts matching on prefixes. The lock is a
continuous effect that modifies the rules (CR 613.1e-shaped in spirit, though it
is applied in the CR 614 window), not an ability the player HAS.

## Decision 2 — two homes, and only one of them is stored

This is ADR 0072's amendment's split, verbatim, because the catalog needs both
halves and for the same reasons:

- **DERIVED** — a permanent on the battlefield whose printed static is "Your life
  total can't change" (**Platinum Emperion**). Declared as
  `effects.Spec.PlayerLifeTotalLocked`, read off the battlefield through
  `game.CatalogPlayerLifeTotalLocked` on every query, **written nowhere**. Two
  Emperions therefore compose, and one of them leaving cannot revoke the other's
  lock — the argument `CatalogNoMaxHandSize` (#338) and `CatalogPlayerKeywords`
  (#1197) both make at length: a "set on enter, restore on leave" design has to
  answer *restore to what?*, and gets two real cases wrong.
- **GRANTED** — a resolved spell, "until your next turn, your life total can't
  change" (**Teferi's Protection**, **Teferi's Reproach**). This one HAS to be
  stored, because the source is in exile a moment after it resolves. It carries a
  CR 611.2 `Duration` and is swept through `durationExpiredLocked`.

`CatalogAbilityKey`, not `CatalogKey`, on the derived walk: "Your life total can't
change" is a static ability of the permanent, so a Platinum Emperion that has lost
its abilities (layer 6) stops locking. That is the existing reader's posture and
this one shares its walk.

## Decision 3 — one reader, and it tests the duration itself

`playerLifeTotalCantChangeLocked(p *Player) bool` is the life half's twin of
`playerAbilityTokensLocked`: derived grants first, stored ones after, an early-out
on the first true, and **the duration tested at the READ as well as in the sweep**.

The sweep is hygiene, run at known moments (the cleanup step, the beginning of a
turn, the top of a layer recompute). The reader is the truth and has to be right
between them — a lock that ends as your next turn begins must not still be
answering during the priority round that ends the previous turn. That is the
posture `CastPermissionActiveForEffect` and `playerAbilityTokensLocked` both take,
for exactly this reason.

**"Until your next turn" ends as that player's next turn BEGINS** (CR 611.2b,
CR 500.1 — a turn begins with its untap step, and CR 502.1 puts the untapping
inside that step, so the turn has begun before anything untaps). `Duration` already
expresses that: `UntilYourNextTurn` is stamped `Player.TurnsBegun + 1` and
`durationExpiredLocked` compares against the seat-turn counter, not `Turn.Number`.
Three opponents' turns and three cleanup steps therefore do not touch it, which is
the whole point of the card.

A turn the rotation **steps over** is counted by `noteTurnBegunLocked` as it is
skipped (`rotation.go`, CR 800.4m) — so an effect that lasts "until that player's
next turn" ends at the turn that WOULD have begun, and a lock on seat A is not
prolonged for the rest of the game by seat B conceding. Nothing here is new; the
lock reads the same counter every other CR 611.2b effect in this engine reads.

## Decision 4 — the consumer is one engine-owned replacement on the life event

`lifeTotalCantChangeReplacement` (`game/life_lock.go`) joins the three built-ins
in `builtin_replacements.go`'s registry, and it is a built-in for the reason the
other three are: **it is printed in the Comprehensive Rules rather than on any
object**, and there is no source card for the pipeline to hang it on. The
permanent (or the resolved spell) that locked the total is one input to the
predicate; the *prohibition* is a rule.

```go
var lifeTotalCantChangeReplacement = ReplacementEffect{
        Watches:    []EventKind{EventChangeLife},
        Preemptive: true,
        AppliesTo:  func(ev *ReplacementEvent, g *Game, _ *Card) bool {
                return g.lifeChangeIsLockedLocked(ev)
        },
        Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
                ev.Cancel()
                return nil
        },
        Label: "Life total can't change",
}
```

**Everything downstream was already written.** This is the part of the design that
made it worth doing this way rather than any other:

- `changeLifeThroughReplacementsLocked` (`life_tail.go`) already has the branch,
  and the branch already names this card's clause in its comment: *"CR 614.10 with
  a null replacement — 'your life total can't change'. No mutation, and no
  `EventChangeLife` for a 'whenever you gain life' trigger to see, because nothing
  happened."* The continuation still runs, with zero, so
  `LoseLifeEachThenForEffect`'s running total is correct across a table where one
  seat is locked and the drain's caster gains only what the others really lost.
- **Since #482 every writer of a life total runs this window.** That is what makes
  one predicate cover a catalog `GainLife`, a drain, an "each opponent loses 3
  life", the CR 702.15b lifelink credit, the life half of an exchange, and the
  public `ChangePlayerLife` sandbox verb, without any of them learning the rule.
- `payLifeAsCostLocked` already refuses a cancelled payment, and its comment
  already cites the rule: *"CR 119.8 … a cost that involves having that player pay
  life can't be paid", and CR 614.17b … A null replacement on the payment ("your
  life total can't change") therefore refuses the cost with `ErrInvalidParam`*. It
  ends *"Nothing in the catalog makes a player unable to lose life today, so the
  second has no up-front check to back up yet; a card that adds one should add that
  check to the cost validators it reaches."* Decision 6 is that check.

**`Preemptive`, with the cost named.** It applies before every other applicable
replacement and asks no CR 616 ordering question. Three reasons, in the order they
matter:

1. **There is only one answer.** A life-amount replacement (Rhox Faithmender's
   doubler, `effects.LifeGainBecomes`; Bloodletter of Aclazotz's loss doubler)
   rewrites the delta and the lock then cancels whatever it rewrote it to. Both
   orders reach zero. Asking is a prompt with a single outcome — the argument
   `allPureCancels` (#710) and `sameModification` (#792) already make for two
   other windows.
2. **A charged shield must not spend itself on a change that cannot happen.**
   Exactly #420's argument for CR 702.16e, one event kind over.
3. **A life PAYMENT cannot pause at all** (`mustSettleNow`, CR 601.2h), so the
   ordering question could not be asked on the path that most needs the answer.

The declared cost is the same one ADR 0072 §4 names for protection: CR 616.1 would
hand the ordering choice to the affected player, and a replacement that did
something *other than* change the amount — "if you would lose life, draw a card
instead" — would get to do its other thing if it were ordered first. No printed
card in this catalog is that shape, and the branch errs toward "nothing happens",
which is what the lock says.

## Decision 5 — damage is still DEALT; the total does not move

Damage does **not** fire the life window, and that is deliberate and predates this
work: damage reduces life directly once the DAMAGE replacements have settled
(CR 120.3, `applyResolvedDamageToPlayerLocked`), and a life-change replacement must
not get a second bite at one event. So the lock needs a second consumer, and it is
four lines in that one function:

```go
if !g.playerLifeTotalCantChangeLocked(p) {
        p.ChangeLife(-ev.DamageAmount)
        g.invalidateLayersForLifeChangeLocked()
}
```

**Everything else about the damage event still happens**, and each of them is a
rule rather than an oversight:

- `EventDealDamage` fires. CR 120.3a says damage *causes* life loss; the damage is
  dealt either way, and "whenever ~ is dealt damage" triggers see it. This is
  Platinum Emperion's printed ruling.
- The **CR 903.10a commander tally** accrues. It counts combat damage DEALT by a
  commander, not life lost, so a locked player still loses to 21 commander damage.
  Platinum Emperion does not stop that and neither does this.
- **Lifelink still credits** (CR 702.15b). The credit is a life change to the
  source's controller and runs the window on ITS own seat — so a lifelinker
  hitting a locked player gains its controller life, and a locked controller gains
  none.
- **Deathtouch, prevention shields and protection** are all upstream of this line
  and untouched.
- **Poison and infect are untouched**, because they are counters and not a life
  change. ADR 0056's counters-on-players already run their own CR 614 window
  (`RepEventCounter`), which this built-in does not watch. (The damage-to-poison
  conversion itself is still unwired — `DamageResultSource.DamageToPlayer` has no
  caller — which is the "Infect, wither and toxic" seam row, not this one.)

## Decision 6 — an unpayable life cost is refused UP FRONT, not only at payment

CR 119.8 says a cost that involves paying life can't be paid by a player who can't
lose life, and CR 614.17b generalises it: *an event that can't happen can't be
chosen as part of a cost*. `payLifeAsCostLocked` is the backstop and has been since
#808. What was missing is the **gate** — and the argument for adding it is #695's,
made once already for CR 119.4: *announce refused the cast, the view showed the
offer anyway, and a Force of Will at 0 life was a button that could only fail.*

So one predicate replaces the open-coded affordability test at every validator:

```go
// CanPayLifeLocked is CR 119.4 and CR 119.8 in one answer.
func (g *Game) CanPayLifeLocked(p *Player, amount int) bool
```

| site | file | what it gates |
|---|---|---|
| activated ability, `Cost.Life` | `game/activated.go` | CR 602.2b announce |
| catalog ability, `LifeCost` | `game/mutations.go` | the same gate, other path |
| alternative cost | `game/alternative_cost.go` (×2) | the offer AND its validator |
| Phyrexian mana | `game/phyrexian_mana.go` | CR 107.4's "or 2 life" |
| entry replacement | `game/entry_choice.go` (×2) | the shocklands' "you may pay 2 life" |
| bot enumerator | `legal/abilities.go` (×2) | #544 — never offer a refused move |

`legal/cast.go`'s alternative-cost line needs nothing: it is a bot POLICY on top of
an offer `CastOffersForLocked` has already filtered through
`AlternativeCostPayableLocked`, which is one of the rows above. The view follows
for the same reason — it reads the same predicate. **Three readers, one rule**, so
a shown offer, an enumerated move and an accepted cast cannot disagree; that is the
contract `AlternativeCostPayableLocked`'s own comment already states.

`AlternativeCost.LifePayableBy` is deleted rather than kept beside the new
predicate: it was CR 119.4 alone, it had two callers, and both of them hold a
`*Game`. Two spellings of one affordability rule is the drift this decision exists
to avoid.

The entry-replacement rows deserve their own sentence, because the shape there is
a PROMPT rather than a refusal: `offerEntryLifePaymentLocked` already declines to
ask a question whose only answer is "no" (*"A player at 1 life does not get to pay
2, and asking a question whose only answer is 'no' is worse than not asking"*), and
a locked player is in exactly that position. The shockland enters tapped, which is
the un-paid branch — weaker than printed, which is the only direction a
simplification may go.

## Decision 7 — the wire says so, on the seat

`PlayerView.life_total_locked` is a bool beside `keywords`, projected on every
frame from the same effective read. Public and unredacted, for the reason
`keywords` and `emblems` are: it is a fact about the board that changes what every
player at the table may do, and a viewer who can see their Bolt fail but not why is
strictly worse off.

It is **not** folded into `keywords`. That field is documented as "the bare engine
tokens a SEAT has right now", the client parses `protection from <quality>` out of
it, and a non-ability token riding the same list would be a badge built on a
grammar the token does not belong to. A bool is honest and the badge is a case in
the same module (`client/src/lib/playerKeywordBadges.ts`), rendered in #979's badge
language beside the hexproof and protection badges the seat tile already draws.

## Decision 8 — `TurnScopedReplacements` still has no `Duration`, and that is fine

#1200 asked for two things. This ADR builds one of them and **declines the other**,
which is worth stating rather than leaving as an omission:

ADR 0063 Decision 8's deferral stands. `TurnScopedReplacements` is still cleared
wholesale at the cleanup step and still carries no duration field, because the card
that was said to need one does not: "your life total can't change" is
player-scoped, and its home is the player. A duration on the floating replacement
registry would be a **strict generalisation with no card behind it** — and the
census problem in Decision 1 means the first card that did need one would still be
better served by a plain-data entry somewhere else. When a genuinely game-scoped
replacement that outlives a turn arrives ("Fog, but for the whole turn cycle"), the
question to ask first is whether it can be plain data, not how to give the closure
registry a window.

The same goes for **"you can't lose the game"** (Platinum Angel, Phyrexian Unlife,
Gideon of the Trials). A locked life total cannot reach 0, so CR 104.3b never fires
for the locked player, and it is tempting to read that as half of "can't lose". It
is not: CR 104.3 has other clauses (drawing from an empty library, ten poison
counters, 21 commander damage), and gating those is its own seam with its own
tracker (#883, "Win the game by effect; 'can't lose' and 'can't win'"). Nothing
here touches the SBA loop.

---

## What this leaves for whoever comes next

- **Perch Protection** prints this clause verbatim alongside "Gift an extra turn"
  and is blocked on the extra-turns seam, not on this one.
- **Flare of Fortitude** prints "Until end of turn, your life total can't change,
  and permanents you control gain hexproof and indestructible" — the lock half is
  one call on this decision's grant with `DurationUntilEndOfTurn`; what it waits on
  is its alternative cost ("sacrifice a nontoken white creature") and the two
  granted keywords. Batch #299 triaged it accordingly.
- **The gain-only and loss-only prohibitions** — Sulfuric Vortex's "if a player
  would gain life, that player gains no life instead", Erebos's "your opponents
  can't gain life", Rampaging Ferocidon — are a **different home** and need nothing
  from this ADR: they are a permanent's printed static with a scope, which is
  `effects.LifeGainBecomes`'s family (`cards/effects/life_replacements.go`) with a
  null arithmetic rather than a doubling one. They are card work, not engine work.

---

## Consequences

**Good**

- Teferi's Protection is **complete** — three printed clauses, three shipped, zero
  caveats — across three ADRs and three sprints.
- The rule reaches every life change in the engine because #482 already made every
  life change one window. Nothing had to be taught about gain, loss, drains,
  lifelink, exchanges or the sandbox verb.
- Two of the four consumers (`changeLifeThroughReplacementsLocked`'s cancel branch,
  `payLifeAsCostLocked`'s CR 119.8 refusal) were already written, already
  commented with this card's clause, and had never been reachable. They are now
  covered by tests instead of being dead branches.
- Undo, clone and the snapshot are exact across a lock, because the state is a bool
  on a plain-data value in a slice those three already carry — no census entry, no
  restore-point refusal, no schema bump.
- The bot and the client are correct with no new rule of their own: the enumerator
  and the view both read `CanPayLifeLocked`, and `legal_targets` never learned
  about protection either.

**Bad / accepted**

- `Preemptive` is a declared simplification of CR 616.1, with the cost named in
  Decision 4: a replacement that does something *besides* rewriting the amount
  never gets to do it. No catalogued card is that shape.
- The lock makes the **manual life-counter drag** a no-op for that seat. That is
  the rules-correct answer and the same one the window already gives for a doubler,
  but it is a sandbox affordance quietly going away while the badge is up, which is
  why the badge is on the seat tile rather than nowhere.
- A `v6` snapshot binary reading a `v7`-era file would… not exist, because there is
  no bump: the new field zero-values correctly (an older file has no locks, and
  `false` is right), and the reverse — an older binary dropping the bool — costs one
  clause on one seat for one turn cycle, not a vanished board. That is the same call
  #1195 and #1197 made for the two payloads already on `PlayerStatic`, and it is
  worth writing down that it was a call.
- "Your life total can't change" is now a thing a player can BE, and three files
  ask about it. A fourth writer of a life total would have to be routed through
  `changeLifeThroughReplacementsLocked` to be covered — which is #482's standing
  contract, guarded by `cards/effects/life_continuation_guard_test.go`, and not a
  new obligation.

**Cards shipped:** Platinum Emperion (`full`, the derived half),
Teferi's Reproach (`full`, the granted half aimed at an opponent),
Teferi's Protection (`caveats` → **`full`**).
