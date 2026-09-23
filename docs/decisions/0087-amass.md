# ADR 0087 — Amass: find-or-create an Army, and a keyword that carries a subtype

**Status:** Accepted · 2026-09-23 · S46 — permanents that change what they are
**Issues:** [#1236](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1236) (this seam),
[#295](https://github.com/krakenhavoc/cmd_and_ctrl/issues/295) (Orcish Bowmasters, the card it was filed for)
**Tracker:** [#889](https://github.com/krakenhavoc/cmd_and_ctrl/issues/889) — S46, "Permanents that change what they are"
**Numbering:** swept with the AGENTS.md §4 loop on 2026-09-23 — `git fetch origin`,
then `git ls-tree --name-only <ref> docs/decisions` over all 406 remote heads plus
`origin/develop`. The highest number present anywhere is **0086**
(`0086-storm-and-the-turns-cast-order.md`, on `feat/storm`); `0087` appears on no
branch and no open issue reserves it. `0005`, `0024`, `0029` and `0030` stay
permanently unused per AGENTS.md §4.

**Related:** [ADR 0013 §5s](0013-replacement-effects.md) (the CR 614 window on a
keyword action with a count), [ADR 0081](0081-earthbend-and-object-keyed-delayed-triggers.md)
(earthbend — the other counted keyword action whose count is a number of
counters, and the layer-4 shape this one copies),
[ADR 0061](0061-token-creation-and-discard-are-replaceable-events.md) (the one
token-creation path, which the find-or-create branch goes through),
[ADR 0083](0083-token-abilities.md) (token templates with abilities — and why
the Army token is **not** one),
[ADR 0060](0060-leaving-the-game.md) / [#568](https://github.com/krakenhavoc/cmd_and_ctrl/issues/568)
(resolution-time prompts with a continuation),
[ADR 0063](0063-durations-and-control.md) (`PinnedTo`, the duration the subtype
grant takes).

---

## Context

`docs/engine-seams.md` never carried a row for this one. The roadmap's own
detector groups it under "Keyword actions with no primitive (proliferate,
surveil, explore, connive, amass…)", and by September 2026 amass was the last
unclosed member of that bucket: proliferate shipped with #381, scry and surveil
with the CR 614 window in #976, explore and connive have primitives.

#1236 filed it against **Orcish Bowmasters** (EDHREC rank 257, roadmap batch 02 /
#295), and filed it correctly: *"Registered without this clause would ship
stronger than printed by omission is not possible here — dropping 'amass' loses
the whole army-building half of the card, so the card stays unregistered rather
than shipping a card that does something the oracle text doesn't say."* Seventy-six
printed cards amass. The next two by play rate are Dreadhorde Invasion (1476) and
Barad-dûr (1755).

### The rules number is 701.47, not 701.35 or 701.51

#1236's title and body say CR 701.35; the roadmap detector says the same. In the
pinned edition (`MagicCompRules 20260819`) **amass is CR 701.47**. 701.35 is the
number it had before the Alchemy and Attraction keyword actions were inserted
ahead of it, and 701.51 — the other candidate that gets quoted — is *Open an
Attraction*. Corrected here and in the code for the reason
[ADR 0086](0086-storm-and-the-turns-cast-order.md) corrected storm's: a wrong
rules number in a file comment is a wrong citation in every file that copies it.

### The rule

> **701.47a.** To amass [subtype] N means "If you don't control an Army creature,
> create a 0/0 black [subtype] Army creature token. Choose an Army creature you
> control. Put N +1/+1 counters on that creature. If it isn't a [subtype], it
> becomes a [subtype] in addition to its other types."
>
> **701.47b.** A player "amassed" after the process described in rule 701.47a is
> complete, even if some or all of those actions were impossible.
>
> **701.47c.** The phrases "the Army you amassed" and "the amassed Army" refer to
> the creature you chose, whether or not it received counters.
>
> **701.47d.** Some older cards were printed with amass N without including a
> subtype. Those cards have received errata in the Oracle card reference so that
> they read "amass Zombies N."

Four sentences, and #1236's triage of them is right: the engine had no shape for
the first (a token creation that is conditional on the board), no shape for the
second when there is more than one candidate (a resolution-time pick among your
own permanents that is not a targeted announce), and no way for anything to see
the third as an amass rather than as a bare counter add. What #1236 did not
notice, because the reminder text on the WAR printings does not say it, is the
**fourth**: an amass changes what a permanent *is*, which is what puts this on
tracker #889 rather than on the token tracker.

---

## Decision 1 — Amass is a counted keyword action on the CR 614 window

`KeywordActionAmass`, a count on `RepEventKeywordAction`, and one arm of
`applyResolvedKeywordActionLocked`. The same three pieces proliferate, scry,
surveil and earthbend have.

The count is printed and the subtype is printed, and nothing else about the verb
moves across seventy-six cards. That is exactly the shape a CR 614 window is for.
Opening one costs nothing today — nothing printed replaces an amass — and is the
difference between a seam and a rewrite the day one does: *"whenever you would
amass, amass twice that much instead"* becomes a card-side `ReplacementEffect`
narrowing on `KeywordAction == KeywordActionAmass`, with no engine change at all.

It is also #1236's third bullet, answered where it belongs. The issue asked for
the counter placement to be *"CR 701.47c's own event, distinct from a plain
`AddCounter`"*. The placement itself stays a plain `AddCounterByForEffect`, which
is what keeps Hardened Scales, Doubling Season and Corpsejack Menace working; what
carries the amass identity is the **instruction**, one level up, because a
"double the counters from amass" doubles the amass and not each counter.

**Amass acts at a count of zero** (`KeywordAction.actsAtZeroCount`), the second
action in the family to do so and for earthbend's reason: the count is a number of
COUNTERS and the other three sentences do not depend on it. *"If you're instructed
to amass 0, you'll create an Army token if you don't control one, but you won't
put any counters on it"* (War of the Spark release notes). Summons of Saruman off
an empty graveyard and Shagrat, Loot Bearer with no Equipment attached both amass
zero, and both leave an Army on the battlefield — which then dies to CR 704.5f,
which is why `ArmyToken`'s 0/0 is `PrintedPTKnown`.

---

## Decision 2 — The Army token is DERIVED from the keyword, not tabled and not an ADR 0083 template

`effects.ArmyToken(subtype)` builds `game.Card{Name: "<subtype> Army", TypeLine:
"Token Creature — <subtype> Army", 0/0, black}` from the subtype. It is neither a
row in `tokens_table.go` nor a template in `token_catalog.go`, and both exclusions
are deliberate.

**Not ADR 0083's catalog.** That file is for a token that carries an ability of
its own, and `checkTokenTemplate` *refuses* a template that declares none. The
Army token is vanilla — the counters and the subtype are put on it by the verb,
not printed on it — so it has nothing to register and needs no catalog key. This
is the reason this is a new ADR rather than a dated amendment to 0083: 0083's
subject is abilities on tokens, and the object at the centre of this one has
none.

**Not `tokens_table.go` either**, and this is the interesting half. That table is
keyed by printed text and `TokenCard` **panics** on a key it does not have. Four
species have a printed Army token card today (Zombie, Orc, Goblin, Sliver) and
four more are named by amass cards outside tournament-legal sets (Ooze, Rat, Bird,
Fan); the next set can name a ninth. A card that amassed a species with no row
would panic at the moment it resolved, on a live table, for a reason nothing in
the card file hints at. More to the point, the characteristics are not card data at all:
CR 701.47a *derives* them from the keyword's parameter, and a table row per
species is a second copy of a rule the engine already has.

The cost of deriving is that nothing pins the derivation against reality, so the
derivation is pinned explicitly:
`TestRealDumpEveryTokenTemplateMatchesAPrintedToken` — the #1127 guard, gated on
`CMDCTRL_SCRYFALL_DUMP` — now walks `ArmyToken` for all four printed species
alongside the table's own rows, and all four match a printed token card exactly
(0/0, black, `Token Creature — <subtype> Army`).

The template travels to `internal/game` as an argument, because that package owns
no token templates and is not about to start.

---

## Decision 3 — Find-or-create is a board query, then the ordinary token pipeline

```
armies := armiesControlledLocked(actor)
if len(armies) == 0 {
        CreateTokensThenForEffect(… one ArmyToken …, func { re-read the board })
}
choose(armies)
```

Two things about this are load-bearing.

**The creation goes through ADR 0061's one path**, so the CR 701.7b window opens
and a Doubling Season really does make **two** Armies — at which point decision 4's
prompt is a real question, which is the ruling. `CreateTokensThenForEffect` rather
than `CreateTokensForEffect`, because the creation can pause on a CR 616 ordering
prompt and the returned slice is empty when it does; the rest of the amass is the
continuation, not the next line.

**The continuation re-reads the board** instead of trusting the IDs it was handed.
A creation replaced away entirely leaves none and a doubled one leaves two, and
"choose an Army creature you control" is a question about the board in either
case.

`IsArmy` is `IsCreature() && HasSubtype("Army")` and both halves matter: an Army
that has stopped being a creature (Song of the Dryads, Humility) is not amassable,
and a creature that *became* an Army in layer 4 is. Both read through the
effective characteristic, so changeling answers too.

---

## Decision 4 — The pick is a battlefield `ChooseCardsPrompt`, and it is only asked when there is a choice

CR 701.47a's "choose an Army creature you control" is a resolution-time pick among
the controller's own permanents. `QueueChooseCardsForEffect` with
`Zone: ZoneBattlefield`, `Min: 1`, `Max: 1` is that shape and already exists —
`effects.SacrificeChoice` uses the same door, and the choose-cards kind is the one
prompt kind in the engine that carries a **continuation**, which this needs because
the counters and the subtype come after the answer.

**Zero Armies asks nothing** (CR 701.47b: the player amassed anyway). **One Army
asks nothing**, and that is a decision rather than an optimisation: a forced choice
is not a decision, Orcish Bowmasters amasses on every opponent draw, and a modal
the player cannot answer wrongly seventy-six cards' worth of times is a worse table
than a silent one. `QueueChooseCardsForEffect`'s own comment argues the other way
for a chain whose next link is queued by the resolver; amass queues its own
continuation either way, so the argument does not reach it.

**The enumerator needed nothing.** `legal/choices.go` has enumerated
`PendingChoiceChooseCards` since #544, bounds and all, so a bot seat owing this
prompt is offered one move per Army and every one of them dispatches.
`legal/amass_choice_test.go` pins that, plus the two things a floor of one implies:
no move is marked `AlwaysLegal` (a floor above zero has no unconditional answer),
and nothing that is not one of the seat's own Armies is offered.

---

## Decision 5 — The subtype is a layer-4 grant pinned to the object, registered only when it is missing

*"If it isn't a [subtype], it becomes a [subtype] in addition to its other types."*
A continuous effect with no stated duration, adding one subtype at layer 4
(CR 613.1d) — `animateEarthbentLandLocked`'s layer-4 arm one subtype over, and
the reason this seam is on tracker #889 and not on the token one.

**Added, never set.** An Orc Army amassed as Zombies is an Orc Zombie Army, and
both lords see it.

**Indefinite, pinned to `{instance, battlefield-entry stamp}`.** CR 611.2a says
"until the game ends"; CR 400.7 says the Army that dies and comes back is a new
object and the effect named the old one. `Game.PinnedTo(IndefiniteDuration(),
army)` is both at once, and `durationExpiredLocked` drops the entry at the next
sweep rather than leaving a husk per amass in the snapshot census.

**The "if it isn't" guard is the rule and the cost control.** Orcish Bowmasters
amasses on EVERY opponent draw; a game with two wheels in it would otherwise leave
a hundred identical layer-4 effects in the registry for the rest of the game. The
guard reads the EFFECTIVE subtype, so `grantAmassSubtypeLocked` recomputes the
layer cache before it asks — a no-op on every amass that changed nothing, and the
difference between one registration and one per trigger.

---

## Decision 6 — "The Army you amassed" is a second continuation, and it carries the Army

CR 701.47c names the chosen creature for the rest of the sentence, and five
printed cards use the name: Widespread Brutality, Grishnákh, Brash Instigator,
Foray of Orcs, Surrounded by Orcs and Goblin Plate Mail.

`keywordActionTail` therefore grows `amassed func(g *Game, army uuid.UUID) error`
**beside** `then func(g *Game) error` rather than reusing it. Scry's continuation
has no argument to give, and re-deriving the Army afterwards is not possible: with
two Armies out, the one that grew is the one the player picked at the prompt, and
the board alone does not say which.

It runs **exactly once**, cleared through the tail pointer the way `then` is, and
it runs on the abandoned path too — with `uuid.Nil`, because CR 701.47b says the
player amassed even when nothing could happen. A caller sequencing work behind the
verb has to be told even when the answer is "none", which is the call #808 made for
the life tail, #853 for the route tail and #762 for the token tail.

---

## Decision 7 — The subtype goes before the counters, and CR 701.47a says after

`AddCounterByForEffect` can pause: a window with two different counter
replacements in it queues the CR 616 ordering prompt and returns nil with the
placement owed to the resume. So everything the amass still owes is registered
**before** the counters, or a paused Doubling Season prompt would leave an Army
that never got its creature type. It is `applyEarthbendLocked`'s ordering argument,
applied to a different verb.

The swap is unobservable. Counters are not subtypes, so "if it isn't a [subtype]"
reads the same on either side of them, and no player receives priority in the
middle of one resolution for a state-based action to notice (CR 704.3).

**The declared cost, and it is earthbend's too:** a continuation sequenced behind
the amass runs from the counter call's return, so on that same CR 616 pause it
would run before the counters land — Widespread Brutality's Army would deal damage
equal to its pre-amass power. Two different counter-count replacements on one board
is the whole population of that bug, and closing it means a continuation on the
counter API, which is a change to a primitive twenty callers share. Declared here
rather than discovered later;
`game/amass.go: finishAmassLocked` carries the same note.

---

## Cards

Four proof cards. Three are on a roadmap batch's skip list with "keyword actions"
as the recorded blocker; the fourth is here because it is the only printed card
that exercises decision 6 and the read side.

| Card | Batch | What it proves |
|---|---|---|
| **Orcish Bowmasters** (EDHREC 257) | #295 | the row's named card — one ability with two trigger conditions, amass off a trigger, twice in a turn, onto the same Army |
| **Dreadhorde Invasion** (1476) | #306 | find-or-create across turns: the first upkeep mints the Army, the second grows it |
| **Eternal Skylord** (4666) | #451 | decision 5 reaching another card's static — "Zombie tokens you control have flying" finds the amassed Army with no help |
| **Widespread Brutality** (7598) | — | CR 701.47c ("the Army you amassed") and the read side ("each non-Army creature") |

Orcish Bowmasters ships `caveats` with **Notion Thief's caveat, word for word**:
"except the first one they draw in each of their draw steps" is read as "except any
draw during their own draw step", because the engine keeps no per-draw-step tally
(`drawnInOwnDrawStep`, the second half of the "Draw-replacement count" seam row).
Weaker than printed for the Bowmasters' controller, never stronger.

Widespread Brutality ships `caveats` for simultaneity: one source's damage to many
creatures is dealt in battlefield order, which every other catalog card of that
family also does.

---

## Out of scope (explicit deferrals)

- **A continuation that survives a paused counter placement** — decision 7's
  declared cost, shared with earthbend. It is a `then` on
  `AddCounterByForEffect`, not an amass change.
- **The other seventy-two amass cards.** Nothing engine-side blocks them; they are
  batch PRs. The nearest by play rate are Barad-dûr (#309, also waiting on the
  mana pipeline and a per-turn death tally), Lazotep Plating (#388, protection /
  prevention), Gleaming Overseer (#389, protection / prevention), Sauron, the Dark
  Lord (#390, the Ring) and Saruman, the White Hand (#391, cost modification) —
  each waiting on its own row, not on this one.
- **"Whenever an Army you control attacks / deals combat damage"** — the read side
  ships (`effects.Army()`, `effects.ArmiesControlledBy`,
  `game.ArmiesControlledForEffect`, `game.IsArmy`) and the first card to put it in
  a TARGET clause has not been registered yet. March from the Black Gate and
  Sauron, the Dark Lord are the two that will.
- **A replacement that rewrites the amass SUBTYPE** rather than its count. Nothing
  printed does, and `keywordActionTail.armySubtype` is deliberately not on the
  `ReplacementEvent` where a card could reach it.
- **Azog, Moria's Ruin's "its controller amasses Goblins X"** — an amass taken by
  a player who is not the effect's controller. `Amass.Controller` is the field for
  it and nothing else is missing; the card waits on its own batch.

---

## Consequences

- One keyword action constant, one switch arm, one new engine file, and a card
  file that says `Do(Amass{Subtype: "Orc", N: 1})` and nothing else. A printed
  "Amass Slivers 2" needs no code at all.
- The Army token is data nobody has to maintain: a ninth printed species works on
  the day the card that amasses it is registered.
- `game.Card` is unchanged; `keywordActionTail` grows three fields, one of which
  is a `Card` and is therefore `cloneCard`-ed in `cloneReplacementResume` for the
  reason its two slices already were.
- One new prompt SITE, no new prompt kind and no wire change — so the client and
  the bot both answer it with what they already had. It inherits that kind's
  known cost: `chooseCardsResume` is a `dropped` continuation frame counted in
  `ContinuationCensus.ChoiceResumeFrames`, so a table sitting on an unanswered
  Army pick is not a restore point for as long as it sits there. That is
  Ponder's cost and `SacrificeChoice`'s cost, not a new one, and decision 4's
  "only asked when there is a choice" is what keeps it rare: the overwhelmingly
  common amass — nought or one Army — queues nothing at all.
- The "Keyword actions with no primitive" bucket has no unclosed member left.
- The cost: amass is the first verb in the engine that can pause **twice** in one
  instruction (the CR 614 count window, then the Army pick), which makes
  `finishAmassLocked` the only place the action's second half exists. A future
  reader adding a fifth part to the verb has to add it there and not to
  `applyResolvedKeywordActionLocked`'s arm.

---

## Test plan

Engine (`server/internal/game/amass_test.go`), no catalog involved:

1. no Army → a 0/0 black [subtype] Army token with N counters on it;
2. an Army already out → the same Army grows, no second token;
3. an opponent's Army and a non-creature Army are both invisible;
4. amass 0 makes the Army, places nothing, and the 0/0 dies to CR 704.5f;
5. one Army asks nothing; two Armies queue the prompt, place nothing until it is
   answered, and put every counter on the one that was picked;
6. the counters go through the CR 614 placement window (Hardened Scales);
7. the subtype is ADDED, is registered once however many times you amass it, and
   is dropped from the registry when the Army leaves;
8. the count rides the keyword-action window;
9. a CR 614.10 null replacement leaves no Army and still runs the continuation,
   with `uuid.Nil`; and the continuation runs exactly once;
10. `ArmiesControlledForEffect` is the read side.

Cards (`server/internal/cards/effects/amass_cards_test.go`), through the real
catalog: `ArmyToken` against CR 701.47a's words for four species and the CR 701.47d
default; the Bowmasters' ETB half and its draw half growing one Army to 3/3; its
draw-step exemption; Dreadhorde Invasion across two upkeeps and its 6-power
lifelink clause; the Skylord's Army flying; an Orc Army becoming an Orc Zombie
Army; and Widespread Brutality's sweep sparing every Army and killing the Bear.

Enumerator (`server/internal/legal/amass_choice_test.go`): every Army offered and
nothing else, every move dispatches, nothing marked always-legal, and no move at
all when there was no choice to make.

Real dump (`tokens_realdump_manual_test.go`, `CMDCTRL_SCRYFALL_DUMP`):
`ArmyToken` for Zombie, Orc, Goblin and Sliver each match a printed token card.
