# ADR 0056 — Infect, wither and toxic: damage results as -1/-1 counters and poison, inside the one damage tail

**Status:** Proposed · 2026-09-17 · unscheduled (card-coverage audit, wave 2) · tracked on [#748](https://github.com/krakenhavoc/cmd_and_ctrl/issues/748)
**Numbering:** 0052 is reserved for the emblems ADR
([#623](https://github.com/krakenhavoc/cmd_and_ctrl/issues/623)), and 0055
is the loop breaker already on `develop`. 0056 to 0059 were handed out
together for #748, #749, #751 and #753; this one is 0056. On 2026-09-17
every remote branch (209 refs after `git fetch origin`) was checked with the
AGENTS.md §4 loop. None has a `docs/decisions/0056-*` file, and the highest
number on any branch is 0055.
**Builds on:** the single damage tail
(`server/internal/game/damage_tail.go`, PRs #709 and #729 for issues #694
and #711, and #807's continuation), [ADR 0014](0014-combat-keywords.md)
(the combat keyword layer), [ADR 0038](0038-protection-style-keywords.md)
(the closed keyword table), [ADR 0053](0053-combat-damage-beats.md)
Decision 1 (`DamageAssignmentFrame.CombatStep`), and the S13.2 player
counters (`Player.Counters`, the CR 704.5c SBA).
**Related:** [#749](https://github.com/krakenhavoc/cmd_and_ctrl/issues/749)
("can't lose the game" gates, which will wrap the poison SBA without changing
it), [#667](https://github.com/krakenhavoc/cmd_and_ctrl/issues/667)
(regeneration: Skithiryx, Toxic Nim, Cinderbones),
[#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662) (protection:
Phyrexian Crusader).

Line references are to `origin/develop` at `684f2786`. The CR is the
Aug 7 2026 edition. Rulings quoted below come from Scryfall's rulings API,
fetched 2026-09-17.

## Context

### Damage has exactly two results today

Every damage entry point fills in a `damageTail` and lands through
`applyResolvedDamageLocked` (`damage_tail.go:358`). The CR 616 resume
(`pending_choice.go:1286-1322`) calls the same function. #748's body was
right about this, but its line numbers are from `ad9413f` and have moved:

- `applyResolvedDamageToPlayerLocked` (`damage_tail.go:467`) always calls
  `p.ChangeLife(-ev.DamageAmount)` (`:483` on the combat branch, `:492` on
  the noncombat branch). CR 120.3b says damage from a source with infect
  gives poison counters instead.
- `applyResolvedDamageToPermanentLocked` (`:506`) calls
  `applyDamageToPermanentLocked` (`permanent_damage.go:59`), and that
  function always adds to `DamageMarked` on a creature (`:79`). CR 120.3d
  says a source with wither or infect puts -1/-1 counters on the creature
  instead.
- Nothing adds toxic's poison counters (CR 120.3g, 702.164c).

The tail snapshots its source when the event is created, not when the
damage lands. That is the #694 rule: a CR 616 prompt can be answered after
the source has died. Three builders take the snapshot:

| builder | used by | reads the source from |
|---|---|---|
| `combatDamageTailLocked` (`:199`) | `markCombatDamageOnCardLocked` (`mutations.go:4768`), `markCombatDamageToPlayerLocked` (`:4868`) | the battlefield only |
| `effectDamageTailLocked` (`:250`) | `DealDamageToPlayerThenForEffect` (`effect_api.go:582`), `DealDamageToCreatureThenForEffect` (`:666`) | the battlefield only |
| `damageTailFromFrame` (`:279`) | the two CR 510.1c prompt resumes (`mutations.go:4823`, `:4854`) | `DamageAssignmentFrame` (`pending_choice.go:632-698`) |

The tail carries `deathtouch`, `lifelinkTo` and `commanderSource`. It has no
field for infect, wither or toxic, and the frame has no cached field for
them either.

### Keywords

`canonicalKeywords` (`keywords.go:55`) is closed: a keyword joins it in the
change that enforces it. It has no `infect`, `wither` or `toxic` entry, so
the deck importer drops them (`deck/deck.go` `printedKeywords`), and the
coverage scan counts an "Infect" line as unhandled rules
(`coverage.go` `keywordOnlyLine`).

Toxic is the first **numbered, cumulative** keyword. Two consequences:

- Scryfall's `keywords` array says `"Toxic"`, with no number. The N is
  only in the oracle line (`Toxic 1`, `Toxic 4`).
- CR 702.164b sums every instance: "a creature's total toxic value is the
  sum of all N values of toxic abilities that creature has". Every other
  keyword in the table is redundant when repeated (CR 702.80d, 702.90f,
  702.15f). The catalog's grant helpers dedupe on apply: `GrantToAttached`
  (`cards/effects/attachments.go`) and seven other sites call
  `keywordSliceContains` (`cards/effects/wire.go:41`). A Rat that prints
  toxic 1 and gets Karumonix's "other Rats you control have toxic 1" would
  lose the second instance.

### Poison exists, but is not an event and cannot be replaced

- `CounterPoison`, `PoisonLethal = 10` (`counter_types.go:57`, `:74`) and
  the CR 704.5c SBA (`mutations.go:2288-2302`) all exist.
  `Player.Poison` is a legacy mirror of `Player.Counters["poison"]`.
- `AddPlayerCounterForEffect` (`proliferate.go:109`) writes the map
  directly. Its comment (`:103-106`) says it emits no event, and that
  player counters have no replacement pipeline because `RepEventCounter`
  targets cards only (`replacements.go:215-222`: `CounterTarget`,
  `CounterName`, `CounterDelta`).
- Because no event fires, nothing bumps `layerVersion`. The layer listener
  bumps on `EventCounterPlaced` (`layer_listener.go:103`), and that event is
  emitted only for card counters (`mutations.go:5235` `applyCounterLocked`).
  A Corrupted static ("as long as an opponent has three or more poison
  counters", Skrelv's Hive) or Vishgraz's "+1/+1 for each poison counter
  your opponents have" goes stale.

### Counter events do not say who put the counters

CR 120.3b and 120.3d say the **source's controller** puts the counters or
gives them. Several catalog cards care who that is, and none of them can
read it off the event today:

- **Vorinclex, Monstrous Raider** keys its halving clause on the target's
  controller and ships with a caveat about it. Once infect exists, an
  opponent's infect creature hitting Vorinclex's controller's creature
  would be **doubled** (the target is "mine") when the rules say it should
  be halved (an opponent is putting the counters).
- **Lae'zel, Vlaakith's Champion** reads the placer with
  `b13ResolutionInProgressBy` (`cards/effects/batch13_helpers.go:196`),
  which returns `uuid.Nil` outside a resolution. Lae'zel counts `uuid.Nil`
  as "you". So an opponent's combat infect damage would get Lae'zel's +1.
  That is stronger than printed.
- **Nest of Scarabs** reads "you put" with `b11ResolvingController`
  (`batch11_helpers.go:141`): the actor of the most recent `EventResolve`,
  with no boundary. Combat wither damage would credit whoever last resolved
  anything.
- **Doubling Season** says "if **an effect** would put". Magic judges agree
  that combat damage is a turn-based action and not an effect, so wither
  and infect combat damage is not doubled, while a wither spell or a fight
  is ([Magic Judges forum](https://apps.magicjudges.org/forum/topic/23694/),
  [a second thread](https://apps.magicjudges.org/forum/topic/25757/),
  [Magic Judge Q&A](https://magicjudge.tumblr.com/post/122507302349/are-counters-from-wither-or-infect-doubled-by)).
  The catalog entry doubles every counter event.

Today none of this is reachable, because no damage puts counters. This ADR
makes it reachable, so it has to fix it.

### What the rules and rulings ask for

- **CR 120.4** processes damage in order: (b) prevention and damage
  replacements, with the triggers on "dealt"; (c) the damage is "processed
  into its results, as modified by replacement effects that interact with
  those results (such as life loss or counters)"; (d) the damage event.
  Infect and wither are not replacement effects. They choose which result
  the damage has. The results are then replaceable in their own right.
- **Vizier of Remedies** ruling: "If a creature you control is dealt damage
  by a source with wither or infect, that much damage is dealt, but one
  fewer -1/-1 counter is put on your creature." So counter replacements
  apply to the result.
- **Toxic** rulings (Pestilent Syphoner, Karumonix, Skrelv's Hive): toxic
  "doesn't change the amount of combat damage"; "Any other effects of that
  damage, such as lifelink, still apply"; toxic does nothing on combat
  damage to a creature or planeswalker, or on noncombat damage; the counter
  count doesn't change when a damage replacement changes the damage;
  "replacement effects that apply to the number of counters put on a player
  can modify the counters placed this way", naming Vorinclex; and "multiple
  instances of toxic are cumulative".
- **Infect** rulings (Tainted Strike, Blightsteel Colossus, Grafted
  Exoskeleton): infect applies to any damage, not just combat damage.
  "Damage from a source with infect is damage in all respects": lifelink
  still gains life, damage can be prevented or redirected, and damage
  triggers still trigger. Prevented damage gives no counters. Damage to a
  planeswalker removes loyalty as usual. -1/-1 counters stay after the turn
  ends and after regeneration.
- **Solemnity** ruling: infect damage "has no effect on creatures or
  players ... The damage is still dealt for purposes of effects that care
  about damage, such as lifelink." So when the result is replaced away, the
  damage still counts as dealt.
- **Melira, Sylvok Outcast** ruling: lifelink and "whenever deals damage"
  abilities still see the damage when Melira stops the counters.
- **CR 702.80b / 702.90d**: last-known information is used when the source
  has left its zone. **CR 702.80c / 702.90e**: the keywords work from any
  zone (Puncture Blast is an instant with wither).
- **CR 903.10a**: 21 combat damage from one commander. Infect damage is
  still combat damage.

### The cards

The audit counts 25 cards where this is the only core blocker, and 27 where
it is one of several. The Scryfall dump, one pass per English oracle card:

| shape | count | examples |
|---|---|---|
| prints infect | 47 | Glistener Elf, Plague Myr, Blighted Agent, Ichor Rats, Blightsteel Colossus, Skithiryx |
| prints wither | 27 | Boggart Ram-Gang, Necroskitter, Puncture Blast (instant), Hateflayer |
| prints toxic N | 45 | Pestilent Syphoner, Bloated Contaminator, Tyrranax Rex (toxic 4), Ixhel (toxic 2) |
| grants one of the three | 48 | Tainted Strike, Triumph of the Hordes, Phyresis, Grafted Exoskeleton, Karumonix, Massacre Girl, Known Killer, Skrelv's Hive |
| reads a poison count ("corrupted" and similar) | 28 | Vishgraz, Skrelv's Hive, Contaminant Grafter, Phyrexian Swarmlord |
| replaces counters on players or -1/-1 counters | — | Vorinclex, Solemnity, Melira, Vizier of Remedies, Halving Season |

**30 Commander-legal creatures print nothing except enforced keywords and
infect, wither or toxic N** (reminder text stripped): Bilious Skulldweller,
Blackcleave Goblin, Blightwidow, Boggart Ram-Gang, Branchblight Stalker,
Contagious Nim, Cystbearer, Flensermite, Glistener Elf, Harvest Gwyllion,
Ichorspit Basilisk, Jawbone Duelist, Juvenile Gloomwidow, Lost Leonin,
Oona's Gatewarden, Pestilent Syphoner, Phyrexian Digester, Plague Stinger,
Priests of Norn, Razor Swine, Rustrazor Butcher, Scourge Servant, Scuzzback
Scrapper, Sheoldred's Headcleaver, Shriek Raptor, Sickle Ripper, Smoldering
Butcher, Tine Shrike, Tyrranax Atrocity, Wildslayer Elves. Once the tokens
join the table (Decision 1), the importer stamps them and these cards need
no catalog entry.

## Decision 1 — Tokens: `infect`, `wither`, and a numbered `toxic N`

- `"infect"` and `"wither"` join `canonicalKeywords`, and are named
  constants (`game.KeywordInfect`, `game.KeywordWither`), as changeling is.
- Toxic is written **`"toxic N"`**, with N a positive decimal integer, as one
  string in `Characteristic.Abilities`. `CanonicalKeyword` accepts it by
  pattern (`toxic` followed by one space and a positive integer) and returns
  it lowercased and normalised (`"Toxic 01"` becomes `"toxic 1"`). A bare
  `"toxic"` is **not** canonical: it names no amount, and accepting it would
  let the importer's `"Toxic"` array entry through with nothing to count.
- `game.ToxicTotal(c *Card) int` sums the N of **every** `toxic N` entry in
  the card's effective abilities, duplicates included (CR 702.164b). It is
  the only reader. `HasKeyword(c, "toxic")` is not how toxic is read.
- **Repeated tokens are not deduped for cumulative keywords.** A new
  `game.AppendKeywordAbility(abilities []string, kw string) []string`
  appends a cumulative token every time and a redundant token only once.
  The eight catalog dedupe sites that use `keywordSliceContains` move onto
  it in the same PR. A grant of `toxic 1` to a card that prints `toxic 1`
  then gives a total of 2, and a second haste grant still makes one badge.
- **Importer:** `printedKeywords` accepts `infect` and `wither` through the
  normal path. When the Scryfall array contains `"Toxic"`, the importer
  reads the numbered token off the oracle keyword lines
  (`keywordLinesOf`, which is already the confirm step used for hexproof)
  and stamps each `toxic N` it finds there. If no line has a number, the
  importer stamps nothing, which is weaker and never stronger. The coverage
  scan's `keywordOnlyLine` accepts `Toxic 1` through the same
  `CanonicalKeyword`.
- **Removal:** "loses infect" (Melira's third ability) is an ordinary
  layer-6 removal of the token. Removing toxic removes every `toxic N`
  entry, because CR 702.164 has no "loses toxic 1" wording and none is
  printed.

Why a string token rather than a counted field (`Card.Toxic int`): grants
already travel through layer 6 as strings, with timestamps, dependency order
and removal, and a separate int would need its own layer path to get any of
that. Karumonix, Skrelv and Vishgraz's Mites all grant toxic through statics
or token templates that already emit `Abilities` strings.

## Decision 2 — The tail snapshots three more facts, plus the placer, from one reader

`damageTail` gains:

```go
infect     bool      // CR 702.90: player → poison; creature → -1/-1 counters
wither     bool      // CR 702.80: creature → -1/-1 counters
toxicTotal int       // CR 702.164c: extra poison on combat damage to a player
controller uuid.UUID // the source's controller: who puts/gives the result counters (CR 120.3b/d)
```

`controller` is separate from `actor` on purpose. `actor` is what
`EventDealDamage` is attributed to, and the noncombat paths leave it unset
(`damage_tail.go:239-242`). That is a log decision, and this ADR doesn't
change it. The placer is a rules fact, and every path needs it.

**One reader for every keyword the tail snapshots.**
`damageSourceTraitsLocked(sourceID) sourceTraits` replaces the inline
`HasKeyword` calls in both live builders and returns deathtouch, lifelink,
infect, wither, the toxic total, the controller and the commander flag. It
looks for the source in this order:

1. **Battlefield**: effective characteristics, as today.
2. **Stack**: the spell's own card, read with `HasKeyword`'s off-battlefield
   fallback (printed keywords). This is CR 702.80c and 702.90e ("from any
   zone"), and it is what makes Puncture Blast work. Deathtouch and lifelink
   get the same read (CR 702.2d, 702.15d), and today no catalog spell prints
   either one.
3. **`lastKnownBattlefield`**, when an entry is present (it holds
   `Effective()` from just before the exit, `triggers.go:497`).
4. Otherwise **nothing**: no keywords and no controller. This is the same
   limit deathtouch and lifelink have today. The engine has no durable
   post-departure LKI (engine-seams.md row "Post-departure LKI"). So "when
   this dies, it deals 2 damage" from a creature with infect deals ordinary
   damage. A card with that shape ships with a caveat or waits.
   *In scope:* the stack read. *Out of scope:* durable LKI.

A source that is a stack **ability** reads its source permanent, which is
what "this creature deals damage" means. An ability whose source has left
falls under step 3 or step 4.

**The frame.** `DamageAssignmentFrame` gains `SourceInfect bool`,
`SourceWither bool` and `SourceToxic int` (all `omitempty`, server-side
only, not projected onto `DamageAssignmentView`). It fills them from the
same reader when `queueDamageAssignmentPromptLocked` (`mutations.go:4732`)
queues the prompt. `damageTailFromFrame` copies them. A frame restored from
a snapshot written before this change has zeros and resumes as ordinary
damage, the same fallback ADR 0053 chose for `CombatStep`. There is no
snapshot schema bump, because the frame is embedded by value.

**Game-level "as though it had wither".** Everlasting Torment ("All damage
is dealt as though its source had wither") would OR into `wither` in the
reader. It is named here so the hook has one obvious home. It is not built.

## Decision 3 — Damage to a creature: -1/-1 counters through the counter window, marked damage otherwise

`applyDamageToPermanentLocked` stops taking a bare `deathtouch bool` and
takes the tail. The creature clause becomes:

```
if isCreature:
    if t.infect || t.wither:
        put amount "-1/-1" counters on it through the counter window,
        placed by t.controller, combat = t.combat          (CR 120.3d)
    else:
        DamageMarked += amount                              (CR 120.3e)
    if t.deathtouch:
        MarkedLethalByDeathtouch = true                     (CR 702.2b)
```

The planeswalker clause (loyalty off, CR 120.3c) and the battle clause
(defense off, CR 120.3h) do not change. The clauses stay additive, so an
animated planeswalker with infect damage gets -1/-1 counters **and** loses
loyalty.

- **Deathtouch composes unchanged.** It is about damage *dealt*, not marked,
  so a wither and deathtouch source sets the flag even when a Solemnity
  stops every counter.
- **The counters are a result, not a replacement.** They go through
  `RepEventCounter`. Vizier of Remedies, Winding Constrictor, Vorinclex and
  Solemnity all see them, as the Vizier ruling requires. They do not go
  through `applyCounterLocked` directly the way loyalty removal does
  (`permanent_damage.go:30-36`), because loyalty removal is a removal and
  this is a placement.
- **What the continuation is told.** `dealt` stays the damage amount. The
  counters that actually landed don't matter to "the damage dealt this way",
  and the Solemnity ruling says the damage was still dealt. Lifelink credits
  the damage amount for the same reason.
- **Lethal-damage assignment** (CR 510.1c, `mutations.go:4666`) does not
  change. It reads `CurrentToughness() - DamageMarked`, and -1/-1 counters
  already lower `CurrentToughness`. Power is snapshotted per combat damage
  step (`mutations.go` `assignAndDealCombatDamageLocked`), so first-strike
  infect counters shrink the other creature's regular-step damage, which the
  Phyrexian Crusader ruling asks for.
- **The manual sandbox mark** (`damageTailManualMark`) doesn't change. It is
  the "fix the board by hand" verb and never had a source.

### The second window does not nest inside the first

#748 asked how a nested pause resumes. It doesn't need to. When the tail
runs, the damage event's CR 614 window has already **settled**, and so has
its own CR 616 prompt, if it had one. The counter placement is a **new,
separate** replacement event, which is the CR 120.4b → 120.4c order. It gets
a fresh event ID and its own window. If two different counter replacements
apply (Winding Constrictor and Vizier of Remedies on the same creature), it
pauses on its own prompt and lands from its own resume. The damage tail
doesn't wait: it emits `EventDealDamage`, credits lifelink and runs `then`
with the damage amount, none of which depends on how many counters land.

No resume frame nests, no damage state rides the counter event, and the
damage resume doesn't change. Two things do have to be true, and PR 1 makes
them true:

1. **The counter resume sweeps SBAs and tolerates a vanished target**, the
   way the life resume and the damage resume do (`pending_choice.go:1252-1322`).
   Today the `RepEventCounter` case (`:1248-1249`) calls
   `applyCounterLocked` and returns its error. When the target has left, the
   answer fails after the prompt is dequeued, which is the #694 bug shape,
   and a -1/-1 placement that pauses is not swept at the answer. The fix
   matches the other two cases: log `EventEffectError`, return nil, and call
   `runStateChecksLocked()`.
2. **During the combat damage step, the pause does not hold the step.**
   Combat damage paths already ignore `paused` (`mutations.go:4798`). The
   counters land when the affected player answers. Their SBA sweep happens
   at that answer, and triggers see `EventCounterPlaced` then.

## Decision 4 — Damage to a player: poison instead of life, plus toxic, in one counter event

`applyResolvedDamageToPlayerLocked` becomes:

```
poison := 0
if t.infect:  poison += amount          // CR 120.3b, instead of life loss
else:         p.ChangeLife(-amount)     // CR 120.3a (same slot as today on each branch)
if t.combat && t.toxicTotal > 0:
              poison += t.toxicTotal    // CR 120.3g, in addition
if t.combat && t.commanderSource != Nil:
              RecordCommanderDamage(...) // CR 903.10a — unchanged, infect or not
emitDealDamageLocked                     // unchanged position on each branch
if poison > 0:
              give `poison` poison counters through the counter window,
              placed by t.controller, combat = t.combat
creditLifelinkLocked(amount)             // CR 120.3f, unchanged
```

- **One placement, not two**, when a source has both infect and toxic. CR
  120.4c processes a damage event into its results as one step. The Winding
  Constrictor ruling treats "counters of multiple kinds at the same time" as
  one placement per kind. With Vorinclex's halving, one event of 3 gives 1,
  and two events of 2 and 1 would give 1 + 0. With a "+1" effect, one event
  gets +1, not +2.
- **Toxic needs damage that was actually dealt.** The function already
  returns before this point when `ev.DamageAmount <= 0` (`damage_tail.go:468`),
  so prevented damage gives no toxic counters. A damage doubler or halver
  doesn't change the toxic count, because `toxicTotal` is not multiplied by
  the amount.
- **Toxic is combat damage to a player only.** The permanent tail never
  reads `toxicTotal`, so combat damage to a planeswalker or battle gives no
  poison, as the rulings say.
- **Double strike** gives two damage events, so two toxic placements and
  two infect placements. That is correct: each combat damage step deals
  combat damage.
- **Infect damage doesn't change life**, so no `EventChangeLife` fires and
  "whenever a player loses life" doesn't trigger. `EventDealDamage` still
  fires, and so do lifelink and CR 903.10a. A commander with infect runs
  both clocks.
- **The 10-poison SBA doesn't change.** The tail never runs SBAs (file
  comment, `damage_tail.go:48-54`). The combat step sweeps after the step,
  the effect paths sweep at the resolution bookend, and the counter resume
  sweeps at its answer (Decision 3). When #749 lands its "can't lose" gate,
  the gate wraps the same check.

## Decision 5 — Counters on players go through the replacement window, and every counter event says who put it

`ReplacementEvent`'s counter half gains three fields:

```go
CounterPlayer           uuid.UUID // set instead of CounterTarget for a player counter
CounterPlacer           uuid.UUID // who puts / gives the counters; Nil = unknown
CounterFromCombatDamage bool      // the placement is a result of combat damage (not "an effect")
```

**Why one event kind for both targets, not a new `RepEventPlayerCounter`.**
The cards that replace counters on players (Vorinclex, Lae'zel, Halving
Season, Solemnity, Melira) are the same cards that replace counters on
permanents. With one kind, each of them keeps one replacement entry and one
`Watches` value. Every existing counter replacement in the catalog (Doubling
Season, Hardened Scales, Winding Constrictor, Vorinclex, Primal Vigor,
Corpsejack Menace, Branching Evolution, Conclave Mentor, Lae'zel, the
batch-28 helper) starts its `AppliesTo` with
`g.LookupCardForEffect(ev.CounterTarget)`, and that lookup fails for
`uuid.Nil`. So none of them can apply to a player event by accident. A test
(test plan 12) iterates every registered counter replacement against a
player-targeted event, so a future card that forgets the check fails CI.

- `AddPlayerCounterForEffect` builds a `RepEventCounter` with
  `CounterPlayer` set and runs `applyReplacementsLocked`. `applyPlayerCounterLocked`
  is the landed body. It clamps at zero, mirrors the legacy `Poison` and
  `Energy` ints, and emits the new event below. `affectedPlayerForEvent`
  (`pending_choice.go:1066`) returns `CounterPlayer` for such an event (CR
  616.1: the affected player chooses). A new
  `AddPlayerCounterByForEffect(placer, player, kind, n)` sets the placer.
  The old signature passes `uuid.Nil`.
- `ProliferateForEffect` takes the proliferating player and passes it as
  the placer for both permanents and players.
- **The sandbox verbs `SetPoison` and `AddPlayerCounter` skip the window**
  (`set_poison` is set semantics, and a replacement can't apply to "set the
  total to 7"), but they emit the event and bump the layer version. The
  manual verb for card counters (`AddCounter`) keeps its window, as it does
  today.
- **`EventPlayerCounterPlaced`** (`"player_counter_placed"`) is the new
  event kind: `Target` = player, `Label` = kind, `Amount` = **the signed
  delta that landed**, `Actor` = placer, `Source` = the damage source or
  effect source. It is a separate kind from `EventCounterPlaced` because
  card-counter trigger helpers compare `ev.Target` against card IDs and walk
  the log by `Label` (`batch33_helpers.go:108`, `batch12_helpers.go:191`),
  and a player UUID with label `"poison"` must never reach them. It carries
  the delta, not the total, because the total is on `Player.Counters`, and
  the card event's post-change total is exactly what forced three helpers
  to walk the log backwards to find the delta.
- **The layer listener bumps on `EventPlayerCounterPlaced`**, with no
  condition. Player counter changes are rare. A condition like the
  hand-size one (`layer_listener.go` `handSizeStaticIsLiveLocked`) would
  save nothing measurable, and it is exactly the kind of gate that went
  stale for a sprint on Psychosis Crawler. This covers Corrupted statics and
  poison-count P/T. It doesn't touch the engine-seams row "Layer
  invalidation on hand / life / attack / graveyard state", which is about
  life totals and other zones.
- **`EventCounterPlaced` gains `Actor` (the placer) and `Source`** when they
  are known. `applyCounterLocked` takes them from the event. Unknown stays
  `uuid.Nil`, as it is today.

**Catalog readers move onto the placer in the same PR** (PR 1), because
PR 2 makes the wrong answers reachable:

| card / helper | today | after |
|---|---|---|
| Doubling Season | doubles every counter event | skips `CounterFromCombatDamage` ("an effect", per the judge rulings above); still doubles noncombat wither/infect, fights and proliferate |
| Vorinclex, Monstrous Raider | keys "you put" on the target's controller (caveat) | reads `CounterPlacer` when it is set and falls back to the target's controller when it is Nil; applies to `CounterPlayer` events too; the caveat narrows to "placements with no known placer" |
| Lae'zel, Vlaakith's Champion | `b13ResolutionInProgressBy`, Nil counts as "you"; no player counters (caveat) | `CounterPlacer` first; applies to counters on its controller; the caveat about player counters is removed |
| Nest of Scarabs, and the `b11ResolvingController` / `b13ResolutionInProgressBy` readers in batch11, batch12, batch13, batch27 and batch33 and in Exemplar of Light | last resolver | `ev.Actor` first, and their current heuristic only when it is Nil |
| Winding Constrictor, Hardened Scales, Primal Vigor, and the other passive "would be put" cards | — | unchanged. No "effect" or "you" in their text, so combat-damage counters count, as the Vizier ruling implies |

**Not solved, declared.** The engine emits one damage event per source per
target. A counter replacement that CR 120.4c applies once to *simultaneous*
damage from several sources ("Vizier of Remedies causes only one counter
fewer ... Its effect doesn't apply separately for each creature dealing
damage") would apply once per source instead. No catalog card does this
today. A future Vizier of Remedies entry must carry that caveat or wait for
a batched counter window (the `OncePerBatch` idea from #594, on the
replacement side).

## Decision 6 — The client shows the keywords, and poison stays where it is (display: owner)

Engine side (fixed): `PlayerView.poison` and `counters` already carry the
value (`docs/protocol.md` PlayerView). `PlayerIdentity.svelte` already has a
poison stepper (`:313-331`) and a poison marker when the count is above
zero (`:368-370`), and `POISON_LETHAL` mirrors the server's
(`counterTypes.ts:27`). `KEYWORD_ICONS` (`keywordIcons.ts`) gets entries for
`infect` and `wither`. `toxic N` tokens fall back to the three-letter badge
until the owner picks a treatment. `docs/protocol.md` documents the new
tokens and `player_counter_placed`.

What a player *sees* when poison or -1/-1 counters happen is a product
decision. It is open question 2 below. PR 3 is written only after it is
answered.

## Decision 7 — Bots value poison

`aiseat/heuristic` reads life only: `lethalPush` compares `through >= def.Life`
(`combat.go:237`), and `attackValue` and `decideBlock` price damage with
`MarginalLife` (`combat.go:246`, `:282`, `:352`). PR 4:

- `SeatEval` gains `Poison`, and `Weights` gains `Poison` and `PoisonDanger`,
  which is quadratic above a threshold, the same shape as `LifeDanger`.
- Damage from an attacker is priced by what it does: an infect attacker's
  damage to a player is poison progress (`PoisonLethal - poison`) and not
  life. A toxic attacker adds `ToxicTotal` poison on top of the life. The
  view reads the tokens from `CardView.abilities`. `lethalPush` checks both
  clocks.
- Blocking an infect attacker values the permanent -1/-1 counters on the
  blocker. It is not marked damage that goes away at cleanup.
- The model tier's prompt (`aiseat/model/prompt.go`) states poison counts.
- **Catalog soak** (`aiseat/catalog_soak_test.go`) adds the first-wave cards
  and the 30 keyword-only creatures to its pool. As with every soak, a stall
  is reported, not failed.

No legal-move enumerator changes. None of this adds a move or a prompt kind.

## Decision 8 — Out of scope, stated

- Durable post-departure LKI for damage sources (Decision 2, step 4).
- Batched counter replacements across simultaneous damage (Decision 5).
- Game-level "as though it had wither" (Everlasting Torment) and "damage
  becomes -1/-1 counters" damage replacements (Soul-Scar Mage). Both are
  CR 614 damage-window effects, not keywords.
- Two-Headed Giant shared poison (CR 810.10). The engine has no team model.
- Regeneration (#667). It blocks Skithiryx, Toxic Nim and Cinderbones.
  Protection (#662) blocks Phyrexian Crusader.
- "Proliferate" as a player choice. It is still the deterministic
  beneficial pick (`cards/effects/proliferate.go`), which already adds
  poison to opponents and not to its controller.

## Consequences

- **30 Commander-legal creatures become fully automatic by import**, with no
  catalog file. The coverage census changes. The card PR regenerates it.
- **Doubling Season stops doubling combat-damage counters.** Nothing can put
  combat-damage counters today, so no existing game changes. It is still a
  behaviour a player could notice, and it matches the rulings.
- **Two caveats narrow and one goes away.** Vorinclex's halving caveat
  narrows to placements with no known placer. Lae'zel's "counters put on
  you aren't increased" caveat is removed, because player counters now go
  through the window.
- **Player counters gain a replacement window and an event.** A few more
  allocations per proliferate or poison placement, and one extra layer
  recompute when a player counter changes.
- **A commander with infect threatens two losses at once** (21 commander
  damage and 10 poison), as it does in paper.
- **A paused -1/-1 placement lands after the combat step sweep.** The
  creature that should die dies when the ordering prompt is answered, not
  in the step's own sweep. That is one prompt later than paper, but the
  board is right once the prompt closes, and the same thing is already true
  for any paused event in a combat step.
- **Counter-resume hardening** (Decision 3) also fixes the existing
  `RepEventCounter` resume, which failed the answer when the target had
  left.
- **`docs/engine-seams.md`**: the "Infect, wither and toxic" row moves to
  Closed when PR 2 lands. The Vorinclex and Lae'zel notes update in PR 1.
- **`counter_types.go:53`** cites CR 122.1d for poison. The pinned CR puts
  poison at 122.1f. PR 1 corrects the citation.
- **The stale sprints pointer is already fixed.** `docs/sprints.md:745` no
  longer sends infect to S18 #68: `bcac391a` re-pointed it to #748, and the
  roadmap footnote (`card-coverage-roadmap.md:935-938`) says the same. PR 2
  changes that line to "shipped" with its PR number and updates the roadmap
  footnote and the §3 keyword table in AGENTS.md.

## Alternatives considered

- **Infect and wither as CR 614 replacement effects on the damage event**
  ("instead of dealing damage, put counters"). Rejected. The rules say the
  damage **is dealt** and only its result changes (CR 120.3, 120.4c), and
  lifelink, "whenever deals damage" triggers and prevention all depend on
  that. It would also put the keyword into the CR 616 ordering against
  prevention shields, where a player could "order" infect after a Fog.
- **Write the -1/-1 counters and poison directly** (`applyCounterLocked`, the
  way loyalty removal works). Rejected. The Vizier of Remedies and toxic
  rulings both say counter replacements modify these results. It would be
  simpler, and wrong in the direction of ignoring Solemnity and Melira.
- **A nested resume that stores the damage tail inside the counter event.**
  Rejected (Decision 3). The damage event has settled, the rest of its tail
  does not depend on the counters, and a nested frame would add a
  continuation to census and clone for nothing.
- **Two poison placements (infect, then toxic).** Rejected (Decision 4):
  that changes the result under halving and "+1" replacements.
- **A separate `RepEventPlayerCounter` kind.** It would guarantee that
  existing card predicates never see a player event. Rejected for one kind
  plus a CI test (Decision 5), because every card that cares about counters
  on players also cares about counters on permanents.
- **Reuse `EventCounterPlaced` for players.** Rejected. Card-counter trigger
  helpers compare `Target` against card IDs and walk the log by label.
- **`Card.Toxic int` or a counted grant.** Rejected (Decision 1). A string
  token rides layer 6 with timestamps and removal for free.
- **Repeat a bare `"toxic"` token N times.** Rejected. The badge row can't
  render it, and "toxic 2" printed once looks the same as toxic 1 granted
  twice, which only matters for display, but it matters there.
- **Keep deathtouch and lifelink battlefield-only while reading the stack for
  infect and wither.** Rejected. One reader for every source keyword is the
  #711 lesson: two readers drift.

## PR split

**PR 1 — counters on players through the window; who put the counters.**
No keywords yet.
- `ReplacementEvent.CounterPlayer`, `CounterPlacer`, `CounterFromCombatDamage`.
  `AddPlayerCounterForEffect` through the window, and
  `AddPlayerCounterByForEffect`. `applyPlayerCounterLocked`.
  `affectedPlayerForEvent`. The proliferate placer.
- `EventPlayerCounterPlaced`, the layer-listener bump, the sandbox verbs
  emitting it, and `Actor`/`Source` on `EventCounterPlaced`.
- The counter-resume hardening (SBA sweep, vanished target).
- Catalog: Doubling Season, Vorinclex and Lae'zel (caveats updated), and the
  resolving-controller readers preferring `ev.Actor`.
- `counter_types.go` citation. engine-seams.md notes. `docs/protocol.md`
  (`player_counter_placed`).

**PR 2 — infect, wither and toxic in the damage tail.**
- The tokens, `ToxicTotal`, `AppendKeywordAbility` and the eight dedupe
  sites, the importer and the coverage scan.
- `damageSourceTraitsLocked`, the new tail fields, the frame fields, and the
  creature and player branches (Decisions 3 and 4).
- AGENTS.md keyword table, `docs/protocol.md` tokens, `KEYWORD_ICONS` for
  infect and wither, the engine-seams row moved to Closed, `docs/sprints.md`
  and roadmap pointers. Census regenerated.

**PR 3 — client display of poison and the damage results** (after open
question 2).

**PR 4 — bots** (Decision 7) and the catalog soak pool.

**PR 5 onward — cards**, in the wave the owner picks (open question 1). Each
card follows AGENTS.md: completeness declared, caveats weaker than printed
and never stronger, and oracle text checked against the dump.

Checks for every engine PR: `go test ./internal/game/... ./internal/cards/... ./internal/deck/... ./internal/aiseat/... ./internal/legal/...`,
then `go test ./...` and `make lint`.

## Test plan

Engine (`internal/game`), PR 2 unless marked:

1. **Infect → creature:** a 2-power infect attacker blocked by a 2/2 puts two
   -1/-1 counters on it. `DamageMarked` stays 0, and the SBA destroys the
   blocker. The counters are still there after cleanup.
2. **Wither → creature** by a noncombat effect (`DealDamageToCreatureForEffect`
   from a wither source), and **wither → player** is plain life loss.
3. **Infect → player:** poison += amount, life doesn't change, and no
   `EventChangeLife`. `EventDealDamage` still fires with `Combat` set.
4. **Infect or wither → planeswalker:** loyalty comes off, and no -1/-1
   counters. **Animated planeswalker** (creature and planeswalker): both.
5. **Toxic:** a 2/2 toxic 1 unblocked means 2 life lost and 1 poison. Toxic
   against a planeswalker or blocker gives no poison. Noncombat damage from
   a toxic creature gives no poison. **Granted twice:** printed toxic 1 plus
   a granted toxic 1 gives 2 poison. **Doubler:** a damage doubler changes
   the life loss but not the poison. **Prevented:** no poison.
6. **Infect and toxic together:** one `EventPlayerCounterPlaced` with the
   sum. Under a halving replacement, floor(sum/2).
7. **Double strike with infect:** two placements, one per step, with the
   ADR 0053 step tags on both damage events. First-strike infect counters
   shrink the blocker's regular-step damage.
8. **Infect granted until end of turn** (the Tainted Strike shape): infect
   before the turn ends, ordinary damage after.
9. **Commander with infect:** poison and a CR 903.10a tally from the same
   hit.
10. **Reaching 10 poison** from combat damage eliminates at the step's
    sweep, and from a spell's infect damage at the resolution bookend.
11. **Lifelink and deathtouch compose:** infect and lifelink gains the damage
    amount (Flensermite). Wither and deathtouch set the lethal flag. Under a
    counter-cancelling replacement (a Solemnity-shaped test effect), lifelink
    still gains and deathtouch still destroys.
12. *(PR 1)* **Player counters through the window:** a test replacement
    doubles poison. Every registered catalog counter replacement returns
    false from `AppliesTo` on a `CounterPlayer` event, unless it is listed
    as opting in (Vorinclex, Lae'zel).
13. *(PR 1)* **Placer:** an opponent's combat infect damage on Vorinclex's
    controller's creature is halved, and on the opponent's own target it is
    doubled. Lae'zel doesn't add to an opponent's placement. Doubling Season
    doesn't double combat-damage counters but does double a wither spell's.
14. **Pause in the middle of the tail:** a creature with two different
    counter replacements (Winding Constrictor and a test "one fewer") takes
    infect combat damage. The damage event completes (lifelink credited,
    `then` run with the damage amount), a CR 616 prompt is queued for the
    counter event, and answering it lands the counters and sweeps. A
    creature brought to 0 toughness is destroyed at the answer.
15. *(PR 1)* **Counter resume with a vanished target** logs and returns nil.
16. **Frame resume:** a multi-blocker infect attacker's assignment prompt,
    answered after the attacker has died to first-strike damage, still puts
    counters and gives poison. A frame decoded from JSON without the new
    fields resumes as ordinary damage.
17. **Stack source:** a wither instant (the Puncture Blast shape) puts
    counters. A source that left the battlefield with no LKI entry deals
    ordinary damage (pins the declared limit).
18. *(PR 1)* **Layer bump:** a static reading an opponent's poison count
    updates the same action that gives the poison.
19. **Proliferate** adds a poison counter to a poisoned opponent through
    the window, with the proliferating player as placer.

Importer and coverage (`internal/deck`, `internal/game`):

20. Glistener Elf imports with `infect`. Tyrranax Rex imports with
    `toxic 4` (plus trample and haste). A Scryfall `"Toxic"` with no
    numbered line stamps nothing. `CanonicalKeyword("toxic")` is false, and
    `CanonicalKeyword("Toxic 2")` is `"toxic 2", true`. The coverage scan
    treats `Flying\nToxic 1` as keyword-only.
21. `AppendKeywordAbility` dedupes `haste` and does not dedupe `toxic 1`.

Bots (`internal/aiseat`), PR 4:

22. `lethalPush` goes for a lethal poison swing against a high-life seat,
    and blocks change when an infect attacker would finish a 9-poison seat.
    The soak runs with the new pool and reports no new stalls.

Client, PR 3: vitest for any pure display helper. The rest is checked by
hand until #689.

## Open questions for the owner

### 1. First card wave

Three options, which build on each other:

- **(a) Engine only.** The 30 keyword-only creatures become automatic by
  import (listed in Context), and no catalog cards ship.
- **(b) (a) plus the sole-blocker catalog cards** whose other text is
  already expressible: Tainted Strike, Triumph of the Hordes, Phyresis,
  Corrupted Conscience, Grafted Exoskeleton, Blighted Agent, Plague Myr,
  Ichor Rats, Viral Drake, Blightbelly Rat, Karumonix, the Rat King,
  Bloated Contaminator, Massacre Girl, Known Killer, and Puncture Blast.
  Each card's PR confirms its other text is expressible, and a card that
  isn't moves out with the reason.
- **(c) (b) plus the poison-count readers** that PR 1's layer bump enables:
  Vishgraz, the Doomhive, Skrelv's Hive, Phyrexian Swarmlord, Septic Rats
  and Contaminant Grafter.

**Recommendation: (c), in two card PRs.** (b) first, then the
poison-readers. Corrupted cards are why PR 1 exists, and without one of
them in the wave nothing in production uses the bump.

### 2. How poison and -1/-1 results show to players

Today the poison chip and stepper exist, the public log has no line for any
counter, a damage line says "dealt 2 damage" whatever the result was, and
elimination doesn't give a reason.

- **(a) Minimal.** Keyword badges only (infect and wither icons, toxic on
  the three-letter fallback). Nothing else changes.
- **(b) Log and chip.** Damage lines name the result ("dealt 2 damage (as
  -1/-1 counters)", "(as poison)"). A new `poison` log entry ("Alice got
  3 poison counters (7/10)"). The elimination line says "10 poison
  counters". The poison chip shows `N/10` and switches to a danger style
  from 7. Toxic shows one badge with the creature's total ("TOX 3").
- **(c) (b) plus a reveal-strip cue** when a player gains poison, the ADR
  0053 and 0054 pattern.

**Recommendation: (b).** Poison is a second life total with a much shorter
clock. A player needs to see how close to 10 they are, and why they lost,
without opening the stepper. (c) adds motion for an event that happens once
per combat at most.
