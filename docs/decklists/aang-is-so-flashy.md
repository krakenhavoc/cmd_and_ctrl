# Decklist support — Aang is so Flashy

Source: [Archidekt](https://archidekt.com/decks/21873490/aang_is_so_flashy) —
Azorius flash + blink, built around *airbend* and end-step flicker.

This is a **triage of one real deck against the catalog**, in the same
shape as [the Pirates list](pirates-mary-read-anne-bonny.md): a deck is
a better forcing function than a card count, because it says which
gaps actually stop a game from being played.

> **Superseded 2026-09-23.** The live per-card checklist is
> [#1306](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1306), re-triaged
> against `develop` @ `ecaab344`: every card, its bucket (full / caveated /
> buildable / blocked) and the seam issue each blocked card waits on. The
> text below is the 2026-09-11 triage, kept as a record; its numbers and
> several of its blockers are out of date.

**Re-triaged 2026-09-11** against `main` @ `0f82505` — re-run after
[#267](https://github.com/krakenhavoc/cmd_and_ctrl/pull/267),
[#268](https://github.com/krakenhavoc/cmd_and_ctrl/pull/268),
[#269](https://github.com/krakenhavoc/cmd_and_ctrl/pull/269),
[#271](https://github.com/krakenhavoc/cmd_and_ctrl/pull/271) and
[#272](https://github.com/krakenhavoc/cmd_and_ctrl/pull/272) landed. Every number below
was recomputed, not carried forward: the deck's 77 entries were resolved
in one pass over `data/scryfall/default-cards.json` and joined against
the oracle IDs the catalog actually registers. Several claims in the
previous revision were false by the time anyone read them — they are
called out inline as **corrections** rather than quietly deleted, because
the same mistakes are easy to make again.

## The numbers

| | count |
|---|---|
| Deck entries | 77 |
| Land entries | 9 (Island ×14 and Plains ×14 are two of them) |
| Nonland entries | 68 |
| Nonland **in the catalog** | **23** |
| Nonland **blocked** | **45** |
| Land entries playable today | 7 of 9 |
| Catalog size (whole registry) | **289** registered specs |

**Counting the catalog honestly.** `grep -c 'OracleID:'` across the
non-test files in `server/internal/cards/effects/` returns **234**, and
that number is wrong — cards register from *tables* in `for` loops
(`temples.go`, `check_lands.go`, `shocklands.go`, `verges.go`,
`surveil_lands.go`, `battle_lands.go`, `bond_lands.go`,
`fastlands.go`, `slowlands.go`, the fetchlands…), so one literal
`OracleID: t.oracleID` stands for ten cards. The registry is the only
honest source: `len(effects.All())` is **289**
([registry.go:80](../../server/internal/cards/effects/registry.go)). An
earlier pass reported 191 for what was then 201 for exactly this reason.

## Card index

The dump was refreshed **2026-09-09**; all 77 entries resolve and the
deck imports.

Before that refresh two cards were genuinely absent, both from Marvel
Super Heroes (2026-06-26): **The Mighty Thor, Jane Foster** and **The
Mind Stone**. Note the second is *not* the ordinary Mind Stone already in
the catalog — it is a Legendary Artifact — Infinity Stone, a different
card with the same-ish name. Both resolve now, to a single oracle ID
each.

Worth knowing for next time: `data/scryfall/last-refresh` is not a
reliable age for the data. It read `2026-04-17` while the dump itself
contained sets released through `2026-07-17`. Check a known-recent card,
not the stamp.

### The art-series trap — read this before calling anything multi-face

**Correction.** The previous revision listed **Appa, Steadfast Guardian**
and **Phelia, Exuberant Shepherd** as multi-face cards. Both are
`layout: normal`, single-faced, no `card_faces` at all.

The error came from resolving a name against a Scryfall **art-series
placeholder printing**. Scryfall ships an art card for many recent sets,
in a set whose code is the real set's code with an `a` prefixed, and
those printings look like this:

```
name      "Appa, Steadfast Guardian // Appa, Steadfast Guardian"
set       atla   (Avatar: the Last Airbender Art Series)
layout    art_series
type_line "Card // Card"
id        235c508b-2c8e-4721-bea4-ca5b80b233b1
```

```
name      "Phelia, Exuberant Shepherd // Phelia, Exuberant Shepherd"
set       amh3   (Modern Horizons 3 Art Series)
layout    art_series
type_line "Card // Card"
id        194205cf-1b48-4df2-81fa-0e3fe6fa2cc9
```

The name is the real card's name **doubled across a `//`**, so a prefix
match, a fuzzy match, or a "does the name contain ` // `" test all say
"multi-face" about a perfectly ordinary creature. The current dump holds
**3,249** such placeholder printings (2,650 of them `art_series` with
`type_line == "Card // Card"`), covering 2,663 distinct names.

The filter that works: **drop any printing whose `type_line` is
`"Card // Card"` or `"Card"`** (that also catches `front_card` and
`double_faced_token` placeholders), and read `layout` off a surviving
printing rather than off the name. Do it in the resolve pass, once, or
it will bite again — it has now cost two separate triage passes.

## Done (batch 1)

| Card | What it exercises |
|---|---|
| Authority of the Consuls | enters-tapped replacement **and** an opponent-ETB trigger on one card |
| Peregrine Drake | ETB untap — the blink deck's mana battery |
| Brineborn Cutthroat | first "during an opponent's turn" cast trigger |
| Archivist of Oghma | first `EventSearchLibrary` consumer (shuffle flavour filtered out) |
| Loyal Warhound | intervening-if (CR 603.4), re-checked at resolution |
| Slithermuse | leaves-the-battlefield (not dies) + hand-size comparison |
| Deputy of Acquittals | optional targeted ETB bounce |
| Thought Vessel | mana rock |
| Fellwar Stone | mana rock (colour set not narrowed — see below) |

Already in the catalog from earlier sprints: **Sol Ring, Arcane Signet,
Solemn Simulacrum, The Wandering Emperor**, plus **Command Tower** on the
land side.

## Done (batch 2)

| Card | What it exercises |
|---|---|
| Azorius Chancery | first bounce land — enters-tapped + a mandatory ETB bounce that can legally take itself |
| Aetherize | reads combat **state** (`Card.AttackingTarget`) rather than waiting on an event |

## Done (S22 engine sprints — nine more nonland, three more lands)

Landed after the previous revision of this doc was written. The last
four rows landed while this correction pass was being written, which is
why the earlier revision of this file described three of them as "in
flight".

| Card | Landed in | What it exercises |
|---|---|---|
| Sun Titan | [#254](https://github.com/krakenhavoc/cmd_and_ctrl/pull/254) | "enters **or** attacks" as ONE ability watching two event kinds; graveyard→battlefield recursion |
| Wash Away | [#257](https://github.com/krakenhavoc/cmd_and_ctrl/pull/257) | cleave — the first alternative cost that *widens* a target clause; `StackItem.CastFromZone` |
| Waterbender's Restoration | [#255](https://github.com/krakenhavoc/cmd_and_ctrl/pull/255), costed in [#271](https://github.com/krakenhavoc/cmd_and_ctrl/pull/271) | delayed blink — exile now, return at the next end step; **waterbend {X} is now paid**, and `TargetSpec.CountFromX` ties the target count to the announced X |
| Cosmic Intervention | [#255](https://github.com/krakenhavoc/cmd_and_ctrl/pull/255) | a turn-scoped replacement that redirects to exile + schedules the return (ships without foretell) |
| Y'shtola Rhul | [#255](https://github.com/krakenhavoc/cmd_and_ctrl/pull/255) | immediate flicker on an end-step trigger |
| Hallowed Fountain | [#258](https://github.com/krakenhavoc/cmd_and_ctrl/pull/258), fixed in [#268](https://github.com/krakenhavoc/cmd_and_ctrl/pull/268) | shockland — now a **real self-replacement with the choice inside it** (`EntryLifeCost`): the pipeline stops and asks before anything moves, so there is no tapped window and no untap event |
| Floodfarm Verge | [#258](https://github.com/krakenhavoc/cmd_and_ctrl/pull/258) | verge — **only the unconditional {W} half** ships |
| Meticulous Archive | [#258](https://github.com/krakenhavoc/cmd_and_ctrl/pull/258) | surveil land — enters tapped + duals; **surveil is not implemented** |
| Aang, the Last Airbender | [#269](https://github.com/krakenhavoc/cmd_and_ctrl/pull/269) | first airbend — `ExileWithPermission` + an unbounded grant (`WhileExiled`) carrying a `CostOverride` |
| Appa, Steadfast Guardian | [#269](https://github.com/krakenhavoc/cmd_and_ctrl/pull/269) | airbends **any number** of targets, and "whenever you cast a spell **from exile**" — `EventCast` now carries `OldZone` |
| Monk Gyatso | [#269](https://github.com/krakenhavoc/cmd_and_ctrl/pull/269) | first `EventBecomesTarget` consumer (CR 115.3) — the trigger lands *above* the spell that targeted, so the removal fizzles |
| The Wandering Rescuer | [#271](https://github.com/krakenhavoc/cmd_and_ctrl/pull/271) | first convoke — `Spec.TapCost`, creatures tapped at cast time pay {1} or one mana of their colour |

### Corrections to batch 2's engine findings

**"A catalog replacement effect cannot fire on its own source's entry"
— no longer true, and the machinery it needed exists.**
`gatherActiveReplacementsLocked` now has a third gathering block,
consulted only for a card that is **not** already on the battlefield,
which looks up the ENTERING card's own replacements and passes the card
itself as `source`
([replacements.go:471–505](../../server/internal/game/replacements.go),
guard at `:485`). `SelfEntersTapped()` is built on it, and the whole
conditional-dual-land cycle — Temples, checklands, fastlands, slowlands,
shocklands, battle lands, bond lands, surveil lands — depends on it. The
Worn Powerstone "enter untapped, then tap" workaround is retired; the
pattern in AGENTS.md §7 is reachable and exercised.

**"Nothing returns a card to the battlefield" — false in both
directions.** `ReturnFromGraveyard{Dest: game.ZoneBattlefield}`
([primitives.go:331](../../server/internal/cards/effects/primitives.go))
routes to `ReturnFromGraveyardForEffect`
([effect_api.go:578](../../server/internal/game/effect_api.go), with
`ReturnFromGraveyardUnderControlForEffect` at `:605` for "under your
control"), and `ReturnFromExile` / `Flicker` / `ScheduleDelayedTrigger`
([primitives.go:160, :185, :213](../../server/internal/cards/effects/primitives.go))
cover the exile side. Sun Titan shipped in
[#254](https://github.com/krakenhavoc/cmd_and_ctrl/pull/254) on exactly
this. Its oracle text is also **"mana value 3 or less"**, not 2 — the
previous revision had the old number.

### Sandbox simplifications declared for this deck

- ~~**Waterbender's Restoration ships uncosted**~~ — **resolved.**
  [#259](https://github.com/krakenhavoc/cmd_and_ctrl/issues/259) is
  closed by [#271](https://github.com/krakenhavoc/cmd_and_ctrl/pull/271).
  For one sprint this was the only card in the catalog knowingly
  *stronger* than paper: with no tap-permanents-as-a-cost component and
  no {X} in the printed mana cost there was nothing to charge, so a
  two-mana mass blink shipped. It is costed now — `TapCost:
  Waterbend("{X}")`, X announced at cast time, artifacts and creatures
  tapping for {1} apiece (or mana, or both), and `CountFromX`
  ([targets.go:65–77](../../server/internal/game/targets.go)) tying the
  target count to that same X, so blinking three creatures costs {U}{U}
  plus three. No simplifications remain on the card. Worth keeping the
  history: the gap was found by writing this doc, not by a test.
- **Cosmic Intervention ships without foretell**, for the same reason
  overload was missing from Vandalblast and Cyclonic Rift before
  [#257](https://github.com/krakenhavoc/cmd_and_ctrl/pull/257): foretell
  is not one of the three alternative costs that exist.
- **Meticulous Archive does not surveil**; the enters-tapped half and
  the duals are real
  ([surveil_lands.go:16](../../server/internal/cards/effects/surveil_lands.go)).
  Surveil needs a new `PendingChoice` kind — scry's structure with
  "bottom of library" replaced by "graveyard".
- **Floodfarm Verge ships mono-{W}**: its conditional second mana
  ability has no shape (`ManaAbilityCost` has no condition slot), so the
  card is strictly *worse* than printed
  ([verges.go](../../server/internal/cards/effects/verges.go)).
- ~~**Hallowed Fountain enters tapped, then offers to untap for 2
  life**~~ — **superseded** by
  [#268](https://github.com/krakenhavoc/cmd_and_ctrl/pull/268), which
  made the whole cycle a real CR 614 self-replacement with the decision
  inside it (`EntryLifeCost` → the apply-loop stops and asks before
  anything moves). Paying means the replacement never fires and the land
  enters untapped; declining, or being unable to pay (CR 119.4), fires it
  and the land *enters* tapped. No tapped window, no untap event, no
  priority pass. One declared limit remains, and it is the fetch case
  below.
- **Fellwar Stone** offers the full five-colour pipe rather than the
  intersection of what opponents' lands could produce.
- **Peregrine Drake** untaps up to five *tapped lands its controller
  controls*, in battlefield order — strictly conservative.
- **Deputy of Acquittals** cannot express "**another** target creature":
  `TargetSpec` is declared statically at `init()`. The picker offers
  every creature you control; the effect declines to bounce itself.
- **Slithermuse** auto-picks the opponent holding the most cards. Its
  evoke is now real ([#257](https://github.com/krakenhavoc/cmd_and_ctrl/pull/257)).

### Fetched and reanimated permanents — #263, fixed, with one declared gap

**[#263](https://github.com/krakenhavoc/cmd_and_ctrl/issues/263) is
closed** by [#272](https://github.com/krakenhavoc/cmd_and_ctrl/pull/272).
A permanent reaching the battlefield via a search / fetch used to skip
every entry replacement — a fetched fastland arrived **untapped**, and a
creature an opponent fetched walked past Authority of the Consuls. Both
paths now run `applyReplacementsLocked` before the card leaves its zone,
and both now fire `fireETBHookLocked`.

Two things about the fix are worth carrying forward:

- **The reanimation path had the same hole and was not in the issue.**
  `ReturnFromGraveyardUnderControlForEffect` neither ran the pipeline nor
  fired the ETB hook, so a reanimated shockland ignored its enters-tapped
  clause and a reanimated Solemn Simulacrum fetched nothing
  ([effect_api.go:640–722](../../server/internal/game/effect_api.go)). It
  was found by fixing the neighbour, which is an argument for reading the
  sibling entry sites whenever one of them turns out to be wrong.
- **A declared gap remains, and it points the safe way.** The search
  entry site is deliberately **not** `entryResumable`
  ([replacements.go:140–160](../../server/internal/game/replacements.go)),
  so a **fetched** shockland enters **tapped with no payment offered** —
  weaker than printed, never stronger, and a strict improvement on the
  old "untapped for free". The reasoning is recorded on
  `searchEnterBattlefieldLocked`
  ([effect_api.go:1034–1056](../../server/internal/game/effect_api.go)):
  `executeEntryToBattlefieldLocked` can finish the *move* but knows
  nothing about the search that started it, so resuming generically would
  skip the library shuffle — and a missing shuffle silently leaks library
  order, which is worse than a missing prompt. A faithful version needs
  the search's own continuation to survive the entry prompt.

  The gap is executable rather than merely written down:
  `TestFetchedShocklandEntersTappedWithNoPaymentOffered`
  (`search_chooser_test.go`) pins it, and **flips** to "a prompt is
  queued and the land enters untapped if you pay" when the continuation
  lands. In this deck the case is live — Solemn Simulacrum and Loyal
  Warhound both fetch, and Hallowed Fountain is the shockland.

  **Closed 2026-09-18 ([#478](https://github.com/krakenhavoc/cmd_and_ctrl/issues/478)).**
  The search's own duties ride across the pause on
  `ReplacementEvent.entryTail`, so the search, exile-return and
  reanimation entries are `entryResumable` and a fetched shockland is
  asked. The test flipped as predicted and is now
  `TestFetchedShocklandOffersItsPaymentAndTheSearchWaits`.

### One engine finding worth keeping

`containsFoldASCII` folds **only the haystack**, so a capitalised needle
silently matches nothing. `IsBasicLandExcept("")` is the sharper trap:
an empty needle makes the check true for every card, so the negation
matches nothing at all — a fetch predicate written that way fails
silently rather than loudly. Every existing caller happens to pass a
lowercase literal, so nothing is broken today; the contract is
documented on `IsBasicLandWithSubtype`.

## Blocked, by machinery needed

**45 of the 68 nonland entries.** Groups **overlap** — Aang, Swift
Savior alone is in four of them — so the group sizes below do not sum to
45 and are not meant to. Each group says what would unblock it and what
would still be in the way afterwards.

### Writable today — no engine gap found (4)

These four came up clean against the current primitives. They are not
"done"; they still need a card file, a test and a review. But nothing in
the engine is missing for them, which was not true a week ago:

- **Phelia, Exuberant Shepherd** — attack trigger (`EventAttack`),
  exile up to one other target nonland permanent, delayed return at the
  next end step, +1/+1 counter if it came back under your control. Every
  piece shipped in [#254](https://github.com/krakenhavoc/cmd_and_ctrl/pull/254)
  and [#255](https://github.com/krakenhavoc/cmd_and_ctrl/pull/255).
  **Not a multi-face card** — see the art-series trap above.
- **The Mighty Thor, Jane Foster** — attack trigger + exile-and-return
  **tapped** (`ReturnFromExile{Tapped: true}`), plus "whenever an
  Equipment you control enters", which is an ETB trigger with a type-line
  check. (The deck plays no Equipment other than Sword of Hearth and
  Home, so the second clause is mostly cosmetic here.)
- **Vega, the Watcher** — "whenever you cast a spell from anywhere other
  than your hand, draw a card." **Correction, twice over:** the original
  revision called this blocked because `EventCast` carried no source
  zone; the first correction pointed at `StackItem.CastFromZone` as the
  workaround. Neither is the current answer —
  [#269](https://github.com/krakenhavoc/cmd_and_ctrl/pull/269) put the
  zone **on the event**, reusing the existing `OldZone` / `NewZone`
  fields (`OldZone` is where it was cast from, `NewZone` is always
  `ZoneStack` —
  [events.go:36–45](../../server/internal/game/events.go)). Appa reads
  exactly this. So the trigger is a plain `AppliesTo` on
  `ev.OldZone != game.ZoneHand`, with no stack-item lookup at all.
- **Wan Shi Tong, Librarian** — `{X}{U}{U}`, X +1/+1 counters, "draw half
  X rounded down", and an opponent-searches-their-library trigger.
  X costs are S20 sub-PR 3; the search trigger is Archivist of Oghma's.

### Airbend — shipped; 3 of the 6 cards are in the catalog

**This section said "doing all three unlocks exactly one of the six".
That was wrong within a day:**
[#269](https://github.com/krakenhavoc/cmd_and_ctrl/pull/269) shipped
**Aang, the Last Airbender, Appa Steadfast Guardian and Monk Gyatso**.

The three missing pieces the previous revision named were all built, and
two of them are on `ExilePlayPermission` rather than beside it
([exile_play.go](../../server/internal/game/exile_play.go)):

1. **`CostOverride`** — a mana cost carried by the *grant*, paid instead
   of the card's printed cost. Deliberately **not** an
   `AlternativeCost`: that one is a property of the CARD, looked up by
   oracle ID and offered to anyone casting it, while this is a property
   of a single exiled INSTANCE — two copies of the same card in exile can
   carry different grants, one airbent and one impulse-exiled.
2. **`WhileExiled`** — the unbounded grant. `Active()` short-circuits
   `turn <= UntilTurn` when it is set, and the cleanup sweep skips those
   permissions.
3. **`ExileWithPermission`**
   ([airbend.go](../../server/internal/cards/effects/airbend.go)) — the
   battlefield-facing sibling of `ExileTopWithPermission`. The difference
   that matters is who gets the grant: impulse exile hands the card to
   the player who exiled it; airbend hands it back to its **owner**
   (`GrantTo` zero means the owner, which is not the same as "the
   controller of the effect").

**Two reusable events came with it**, and both are worth more than the
cards that motivated them:

- **`EventCast` now carries the source zone** (`OldZone` / `NewZone`) —
  Appa's "whenever you cast a spell from exile", and anything else that
  cares where a spell came from.
- **`EventBecomesTarget`** (CR 115.3,
  [events.go:196–225](../../server/internal/game/events.go)) — fires once
  per target **slot** at announce, from four sites (cast, two activated
  paths, and the triggered-ability target pick). Announce rather than
  resolution is the whole point: it still fires for a spell that is later
  countered, and the triggers it produces go on the stack ABOVE that
  spell, which is what makes Monk Gyatso work — his trigger resolves
  first, exiles the creature, and the removal fizzles for want of a
  target. The previous revision said "there is no targeting event of any
  kind"; there is now, and "becomes the target of" is a common Commander
  clause well beyond this deck.

**Two declared simplifications, both weaker than printed:**

- **Airbend's {2} is charged, not offered.** Printed airbend says the
  owner *may* cast it for {2} **rather than** its mana cost — both prices
  are legal, so a {W} creature is cheaper at its printed cost. The engine
  charges {2} unconditionally, so the cheaper option is unavailable on a
  card printed below {2}. An offer of two prices narrowed to one of them:
  never stronger.
- **Aang, the Last Airbender's Lesson clause is omitted.** "Aang gains
  lifelink until end of turn" needs a continuous effect with a turn
  duration and the layer engine has nowhere to put one (below). Granting
  it permanently would be stronger than printed, so the clause is left
  out entirely.

**Still blocked, and none of them on airbend itself:**

- ~~**Aang, Swift Savior** — transform DFC, waterbend {8}~~ — **updated
  2026-09-21.** The card was already in the catalog (its ETB airbend
  targets a creature or spell, same as every other airbend user here);
  what was missing was the transform half. ADR 0079's verb closed it:
  Waterbend {8} now transforms Aang through `TransformThis`, with a flat
  {8} mana cost standing in for the printed tap-artifacts-and-creatures
  discount (declared caveat — see `aang_swift_savior.go`), and the back
  face's attack trigger ships in full (#343). **Updated 2026-09-23**: the
  flat {8} is gone — #1310 made waterbend an activated-ability cost
  (`WaterbendCost("{8}")`, ADR 0020 amendment), so artifacts and creatures
  can be tapped to help pay and the card is `full`.
- **Avatar's Wrath** — mass airbend, self-exile on resolution, and a
  continuous "opponents can't cast spells from anywhere other than their
  hands" restriction that would have to outlive its source.
- **Aang, Airbending Master** — experience counters (small, below) and
  "leave the battlefield *without dying*", which `EventLTB` can express
  today via `ev.NewZone != ZoneGraveyard`
  ([events.go:122–127](../../server/internal/game/events.go)).

### Continuous effects with a duration — "until end of turn" (3 blocked, 1 already shipped without it)

**New group, and it is not small.** Every continuous effect in the engine
is sourced from a permanent on the battlefield:
`activeStaticAbilitiesLocked` walks `g.Battlefield` and looks up
`CatalogStaticAbilities(src.OracleID)`
([layers.go:150](../../server/internal/game/layers.go)). There is no
floating effect, no duration field, and nothing that expires at end of
turn. Grep for "until end of turn" in `server/internal/` and the only
hits are comments and the *replacement* pipeline's turn-scoped slot.

Blocked here: **Ambrosia Whiteheart** (landfall +1/+0 UEOT — its ETB
bounce works today), **Katara, Water Tribe's Hope** (base P/T X/X UEOT,
on top of the activated-ability waterbend seam), **The Wind Crystal**
(mass flying + lifelink UEOT).

And one card that is **already in the catalog with the clause left out**:
**Aang, the Last Airbender** ships without "whenever you cast a Lesson
spell, Aang gains lifelink until end of turn", because granting it
permanently would be stronger than printed. That is the sharpest argument
for this item — it is no longer a gap that keeps cards out, it is a gap
that makes a shipped card wrong. It also blocks two of The Wandering
Emperor's three loyalty abilities, and it is the reason a pump spell has
never appeared in the catalog.

### Alternative cast costs the S22 slot does not cover (5)

[#257](https://github.com/krakenhavoc/cmd_and_ctrl/pull/257) shipped
**overload, evoke and cleave** — `Overload()`, `Evoke()`, `Cleave()` at
[alternative_cost.go:31, :49, :68](../../server/internal/cards/effects/alternative_cost.go),
with `ClearsTargets`, `SacrificeOnEntry` and a replacement `Targets`
clause. That retro-fixed Vandalblast, Cyclonic Rift and Slithermuse, and
it is what made Wash Away castable as printed.

Still unexpressed, all of them "pay something else, somewhere else":
**foretell** (Ranar the Ever-Watchful; Cosmic Intervention ships without
it), **plot** (Aven Interrupter), **spree** (Three Steps Ahead),
**warp** (Anticausal Vestige), **prepare** (Skycoach Conductor). Four of
the five also cast from a zone other than the hand, which is S29's
subject rather than S28's.

### Tap permanents as a cost — convoke and waterbend (shipped; 4 left, each on something else)

[#271](https://github.com/krakenhavoc/cmd_and_ctrl/pull/271) built the
shared component — one `Spec.TapCost` serves both, because waterbend is
convoke with a different name and no colour clause — and shipped **The
Wandering Rescuer** (convoke) and the costing of **Waterbender's
Restoration** (waterbend {X}), closing
[#259](https://github.com/krakenhavoc/cmd_and_ctrl/issues/259).

The Wandering Rescuer carries one declared simplification: **its hexproak
grant is inert.** The static ability really does append `"hexproof"` to
every other tapped creature you control, and nothing in the engine reads
it (below) — so the creatures convoked to cast it are targetable when
paper says they would not be. Weaker than printed, and written that way
deliberately so the card starts working untouched the day hexproof lands
in the targeting gate.

The remaining four each have a **second, independent** blocker, so none
of them is waiting on the cost component any more:

- ~~**Clever Concealment** — phasing. No implementation anywhere.~~
  **Updated 2026-09-23**: phasing is built (#1199,
  [ADR 0084](../decisions/0084-phasing.md)) and the card ships `full` —
  any number of target nonland permanents you control phase out, with
  everything attached to them, and come back at your next untap step.
- ~~**Aang, Swift Savior** and **The Legend of Kuruk** — multi-face.~~
  **Updated 2026-09-21**: multi-face transform is no longer the blocker
  (ADR 0079). Both ship — Aang's Waterbend {8} pays a flat mana cost
  instead of the printed tap discount (still this row's gap, just no
  longer blocking the whole card), and Kuruk's front face needs no
  waterbend at all (its own Saga chapters don't print the keyword); the
  back face's Exhaust — Waterbend {20} extra-turn ability is the one
  piece still unshipped, and it is blocked twice over — see
  `avatar_kuruk.go` and the "Extra turns primitive" row in
  `docs/engine-seams.md`.
- ~~**Katara, Water Tribe's Hope**~~ — **updated 2026-09-23**: ships
  `full` via #1310's activated-ability waterbend (`WaterbendCost("{X}")`
  with `MinX(1)`, base X/X in layer 7b). The original note: its waterbend
  is on an **activated ability**, which is a different seam: `Spec.TapCost` prices a *spell*,
  and `AbilityCost` still has only tap-this / sacrifice-self /
  sacrifice-other / mana / life
  ([activated.go:35–60](../../server/internal/game/activated.go)). Plus
  base P/T until end of turn.
- ~~**The Unagi of Kyoshi Island**~~ — **updated 2026-09-23**: ships
  `full` via #1311 (`WardWaterbend("{4}")`: the ward's pay-or-counter
  prompt lets the opponent tap their artifacts and creatures to help). The
  original note: ward, which is a cost paid by the *opponent*, not by the
  controller.

### Multi-face cards (7)

Sea Gate Restoration // Sea Gate, Reborn and Sink into Stupor //
Soporific Springs (`modal_dfc` — spells with land backs, so they do not
count toward the 9 land entries), Aang Swift Savior // Aang and La
(`transform`), The Legend of Kuruk // Avatar Kuruk (`transform`, and a
Saga), Virtue of Knowledge // Vantress Visions (`adventure`), Skycoach
Conductor // All Aboard (`prepare`), Fortune Teller's Talent (`class`).
Layouts read off real printings, not placeholders.

`game.Card` has one name, one type line, one mana cost. This affects
**deck import and validation**, not just resolution — and it is already
visible to players:
[#265](https://github.com/krakenhavoc/cmd_and_ctrl/issues/265) is an
in-app bug report about *this deck's* Sea Gate Restoration entering as a
land with no face prompt and no "pay 3 life" choice.

### Copy effects (3)

**Clone** needs "enters as a copy of a creature", which is CR 613 layer
1 — explicitly deferred to S16.5 in the layer engine's own header
([layers.go:20](../../server/internal/game/layers.go)). Note that the
**token**-copy keystone *does* exist now (`CreateTokenCopy`,
[token_copy.go:65](../../server/internal/cards/effects/token_copy.go),
[#258](https://github.com/krakenhavoc/cmd_and_ctrl/pull/258)) — it
carries the copied card's oracle ID onto the token so every catalog hook
follows. That is what Three Steps Ahead's `+ {3}` mode wants, so **Three
Steps Ahead is blocked on spree, not on copying**. Skycoach Conductor and
Vantress Visions copy a *spell* and an *ability* respectively, neither of
which has a shape.

### Cost reduction (4)

Pearl Medallion, Oketra's Monument, The Wind Crystal, Fortune Teller's
Talent level 3. Nothing in `server/internal/game/` modifies a cost;
scheduled for S28.

### ETB-trigger doubling (2)

Panharmonicon, Virtue of Knowledge. Needs a trigger-count replacement,
distinct from the CR 614 pipeline, which replaces events rather than
triggers.

### Planeswalker loyalty abilities (2)

Both Teferis. **Refinement of the old claim:** the once-per-turn and
sorcery-speed gates *are* enforced — `ActivateLoyalty`
([mutations.go:1397](../../server/internal/game/mutations.go)) applies
the announced delta and sets `LoyaltyActivatedThisTurn`. What is missing
is any catalog hook for the ability's *effect*: `AbilityCost` has tap /
sacrifice-self / sacrifice-other / mana / life and no loyalty component
([activated.go:35–60](../../server/internal/game/activated.go)), so a
loyalty ability is a manual click that moves a counter and does nothing
else. That is exactly how The Wandering Emperor ships today
([wandering_emperor.go](../../server/internal/cards/effects/wandering_emperor.go)).
Teferi, Who Slows the Sunset also needs an emblem.

**Both Teferis are worse off than "no abilities", and it is a live bug.**
Starting loyalty is stamped only from the catalog's
`Spec.StartingLoyalty` — the ETB hook is what applies it
([effect_hooks.go:34–39](../../server/internal/game/effect_hooks.go)) —
so a planeswalker with **no catalog entry** enters with zero loyalty
counters and dies to the CR 704.5i state-based action immediately. That
is [#274](https://github.com/krakenhavoc/cmd_and_ctrl/issues/274), an
in-app report of exactly that ("after casting Teferi… the planeswalker
goes straight to the graveyard"). Neither Teferi in this deck is
registered, so both currently hit the graveyard on resolution. *A sibling
is working on it on `fix/planeswalkers-274`, along with
`docs/decisions/0032-planeswalkers.md`* — read that ADR rather than this
paragraph once it lands, and note that registering a planeswalker with
nothing but `StartingLoyalty` (the Wandering Emperor pattern) is
currently also the workaround for the bug.

### Experience counters (2)

Aang Airbending Master, Katara Waterbending Master. **Correction:** the
previous revision said "counters live on cards; there is no
player-scoped counter store." There is: `Player.Counters map[string]int`
([player.go:165](../../server/internal/game/player.go)) since S13.2, and
`CounterExperience` is a registered counter name
([counter_types.go:64](../../server/internal/game/counter_types.go)).
What is actually missing is one accessor: `AddPlayerCounter` takes the
write lock ([mutations.go:4421](../../server/internal/game/mutations.go))
and effects run *under* it, so this needs an unlocked `…ForEffect`
sibling. That is a small PR, not a subsystem.

With that one accessor, **Katara, Waterbending Master** becomes writable
— its other two clauses are an attack trigger and Brineborn Cutthroat's
cast-during-an-opponent's-turn trigger, both of which exist.

### Keywords the engine does not enforce (8 cards)

`HasKeyword` is a string match over `Effective().Abilities`
([keywords.go:46](../../server/internal/game/keywords.go)), and its
canonical list is twelve combat tokens. Beyond that list nothing reads
the keyword at all:

- **hexproof** — Lotus Field (land), Stoic Sphinx, and **The Wandering
  Rescuer, which is now in the catalog with the grant declared and
  inert**. Nothing in `targets.go` consults the keyword; S20 explicitly
  deferred hexproof / shroud / protection as target-legality modifiers.
  This is the one gap on the list with a card already shipped against it,
  so closing it makes a catalog card better without touching the card.
- **indestructible** — Thassa, The Mind Stone, The Seriema. Only
  mentioned in comments explaining that *sacrifice* ignores it.
- **ward** — The Unagi of Kyoshi Island. No implementation.
- **prowess** — Ty Lee, Chi Blocker. No implementation. *(2026-09-24: shipped in #706 as an enforced keyword; Ty Lee is full.)*

The static half (printing the word) is cheap; the behaviour is the work.

### One-offs

(Herald of Eternal Dawn left this list with #749: "you can't lose the
game and your opponents can't win the game" is a gate the engine reads
at every loss and win, [ADR 0057](../decisions/0057-win-and-lose-by-effect.md),
and the card is in the catalog.) Mandate of
Peace (end the combat phase + a cast restriction), Rabble Rousing (hideaway, plus
"whenever you attack with **one or more** creatures", which over-fires
against per-creature `EventAttack` — the "one or more" batching gap
`events.go` documents at `:162`), The Seriema (station), Misleading
Signpost (re-select an attacker's target), Hullbreaker Horror (can't be
countered, plus returning a **spell** from the stack to hand — no
primitive does that), Venser, Shaper Savant (same spell-bounce gap; the
permanent half works today), Thassa, Deep-Dwelling (devotion — no
implementation; its end-step flicker and its `{3}{U}` tap ability are
both writable now), Enduring Curiosity (the graveyard return exists; "it
returns as an enchantment, not a creature" needs a continuous effect
conditioned on *how* the permanent got there), Ty Lee, Chi Blocker
(prowess, #706; its "doesn't untap for as long as you control this"
lockdown shipped with #1313, and prowess with #706 — the card is full),
Meticulous Archive (in the catalog, but surveil is unimplemented), The
Mind Stone (harness — a once-activated state gate), Deep Gnome
Terramancer ("lands enter under an opponent's control **without being
played**" + once each turn), PuPu UFO (put a land from hand onto the
battlefield; and Towns), Faerie Mastermind ("an opponent's **second**
draw each turn" needs a per-turn draw tally — `CastTally`
[game.go:569](../../server/internal/game/game.go) is the precedent and
counts casts only), Aven Interrupter (exiling a spell off the stack, plus
a cost increase).

### The seven entries the previous triage never listed

Worth recording because the omission is how a "68 nonland" total and a
list of 61 cards coexisted for a week: **Ambrosia Whiteheart, Deep Gnome
Terramancer, Faerie Mastermind, PuPu UFO, Vega the Watcher, Venser
Shaper Savant, Wan Shi Tong Librarian** appear in the deck and appeared
nowhere in the doc. Two of them (Vega, Wan Shi Tong) are writable today.
The fix is structural: resolve the deck programmatically and list every
entry, rather than writing the groups by hand from memory.

## Lands

**5 of the 9 land entries are registered** — Azorius Chancery, Command
Tower, Floodfarm Verge, Hallowed Fountain and Meticulous Archive (the
last three in [#258](https://github.com/krakenhavoc/cmd_and_ctrl/pull/258);
Hallowed Fountain's entry clause became a real self-replacement in
[#268](https://github.com/krakenhavoc/cmd_and_ctrl/pull/268), and the
verge and surveil simplifications above still stand). The two basics
entries (28 cards) need no registration: a card with a basic-land subtype
gets a synthetic mana ability server-side. That leaves **2 blocked**:

**Demolition Field** — the three-part activated cost (`{2}`, `{T}`,
sacrifice) is expressible (`Plus(ManaCost("{2}"), TapCost())` +
`SacrificeThis()`), but the ability makes the *opponent* optionally
search their library, and there is no shape for prompting another player
mid-resolution. Implementable only if that half is declared as a
simplification.

**Lotus Field** — two independent blockers: `hexproof` is not enforced,
and "{T}: Add three mana of any one color" cannot be expressed in the
pipe syntax — three `{W|U|B|R|G}` slots would let the controller pick a
*different* colour per slot, which is a strictly stronger card.

## Suggested order

Re-sequenced again after #267 / #268 / #269 / #271 / #272. **Everything
the last two revisions put at the top has shipped** — alternative costs,
attack triggers, flicker + delayed triggers, airbend, tap-permanents-
as-a-cost, and the search/reanimation entry pipeline with its chooser.
What is left is mostly *card-shaped* work plus four engine gaps, and the
order below is by how much each buys:

1. **Write the four cards that need no engine work** — Phelia, The Mighty
   Thor, Vega, Wan Shi Tong. Still true after this week's merges, and
   Vega got easier: `EventCast` now carries `OldZone`, so its trigger is
   a one-line `AppliesTo`. Phelia is a headline card of the deck. No
   dependency on anyone else.
2. **Two one-accessor additions**, each unlocking a card outright: an
   unlocked player-counter accessor (→ Katara, Waterbending Master, whose
   other two clauses both exist) and a per-turn draw tally alongside
   `CastTally` (→ Faerie Mastermind).
3. **Durations on continuous effects** ("until end of turn"). Now the
   biggest single item: 4 cards here, the omitted Lesson clause on a card
   *already in the catalog* (Aang, the Last Airbender), two of The
   Wandering Emperor's three loyalty abilities, and every pump spell the
   catalog has never attempted.
4. **Honour hexproof and indestructible** — 6 cards, and one of them
   (The Wandering Rescuer) is already in the catalog shipping with an
   inert grant, so this makes a live card better without editing it.
5. **Planeswalkers** — [#274](https://github.com/krakenhavoc/cmd_and_ctrl/issues/274)
   first (a non-catalog planeswalker dies on arrival, which is worse than
   a missing ability), then loyalty-ability effects for the two Teferis.
   *In flight on `fix/planeswalkers-274`, with ADR 0032.*
6. **Multi-face cards** — 7 cards, an open in-app bug
   ([#265](https://github.com/krakenhavoc/cmd_and_ctrl/issues/265)), and
   it touches import, protocol and client. A sprint of its own.
7. **Copy effects / CR 613 layer 1** (Clone), **cost reduction (S28)**
   and **the remaining alt-cast costs — foretell, plot, spree, warp,
   prepare (S29)**. Already scheduled; 1, 4 and 5 cards respectively.

Two smaller follow-ups with a card each already waiting on them: finish
the fetched-entry prompt (the `entryResumable` gap above — the test flips
when it lands) and the waterbend-on-an-activated-ability seam
(→ Katara, Water Tribe's Hope, together with item 3).

One small addition worth folding into whichever PR is nearby: a
spell-bounce primitive for returning a spell from the stack to its
owner's hand (Venser, Shaper Savant and Hullbreaker Horror, and nothing
else here). The targeting event that used to sit beside it in this
paragraph shipped as `EventBecomesTarget` in
[#269](https://github.com/krakenhavoc/cmd_and_ctrl/pull/269).
