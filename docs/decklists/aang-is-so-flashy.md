# Decklist support — Aang is so Flashy

Source: [Archidekt](https://archidekt.com/decks/21873490/aang_is_so_flashy) —
Azorius flash + blink, built around *airbend* and end-step flicker.

This is a **triage of one real deck against the catalog**, in the same
shape as [the Pirates list](pirates-mary-read-anne-bonny.md): a deck is
a better forcing function than a card count, because it says which
gaps actually stop a game from being played.

**Re-triaged 2026-09-11** against `main` @ `5cc2a87`. Every number below
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
| Nonland **in the catalog** | **19** |
| Nonland **blocked** | **49** |
| Land entries playable today | 7 of 9 |
| Catalog size (whole registry) | **261** registered specs |

**Counting the catalog honestly.** `grep -c 'OracleID:'` across the
non-test files in `server/internal/cards/effects/` returns **225**, and
that number is wrong — cards register from *tables* in `for` loops
(`temples.go`, `check_lands.go`, `shocklands.go`, `verges.go`,
`surveil_lands.go`, `battle_lands.go`, `bond_lands.go`,
`fastlands.go`, `slowlands.go`, the fetchlands…), so one literal
`OracleID: t.oracleID` stands for ten cards. The registry is the only
honest source: `len(effects.All())` is **261**
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

## Done (S22 engine sprints — five more nonland, three more lands)

Landed after the previous revision of this doc was written:

| Card | Landed in | What it exercises |
|---|---|---|
| Sun Titan | [#254](https://github.com/krakenhavoc/cmd_and_ctrl/pull/254) | "enters **or** attacks" as ONE ability watching two event kinds; graveyard→battlefield recursion |
| Wash Away | [#257](https://github.com/krakenhavoc/cmd_and_ctrl/pull/257) | cleave — the first alternative cost that *widens* a target clause; `StackItem.CastFromZone` |
| Waterbender's Restoration | [#255](https://github.com/krakenhavoc/cmd_and_ctrl/pull/255) | delayed blink — exile now, return at the next end step (ships uncosted, see below) |
| Cosmic Intervention | [#255](https://github.com/krakenhavoc/cmd_and_ctrl/pull/255) | a turn-scoped replacement that redirects to exile + schedules the return (ships without foretell) |
| Y'shtola Rhul | [#255](https://github.com/krakenhavoc/cmd_and_ctrl/pull/255) | immediate flicker on an end-step trigger |
| Hallowed Fountain | [#258](https://github.com/krakenhavoc/cmd_and_ctrl/pull/258) | shockland — enters tapped + an optional "pay 2 life to untap it" trigger |
| Floodfarm Verge | [#258](https://github.com/krakenhavoc/cmd_and_ctrl/pull/258) | verge — **only the unconditional {W} half** ships |
| Meticulous Archive | [#258](https://github.com/krakenhavoc/cmd_and_ctrl/pull/258) | surveil land — enters tapped + duals; **surveil is not implemented** |

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

- **Waterbender's Restoration ships uncosted** — waterbend {X} is free
  ([#259](https://github.com/krakenhavoc/cmd_and_ctrl/issues/259)). It is
  **stronger than printed**: you exile and return X creatures for {U}{U}
  with no tapping at all. `AbilityCost` has tap-this / sacrifice / mana /
  life and no "tap other permanents" component
  ([activated.go:35–60](../../server/internal/game/activated.go)). *A
  sibling agent is fixing this now on `feat/tap-as-cost`* — expect this
  entry to move to the convoke/waterbend group and out of
  simplifications.
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
- **Hallowed Fountain enters tapped, then offers to untap for 2 life.**
  "As this land enters, you may pay 2 life" is a replacement with a
  choice in it and the replacement pipeline is synchronous, so the
  choice survives as an optional ETB trigger instead
  ([shocklands.go:18–40](../../server/internal/cards/effects/shocklands.go)).
- **Fellwar Stone** offers the full five-colour pipe rather than the
  intersection of what opponents' lands could produce.
- **Peregrine Drake** untaps up to five *tapped lands its controller
  controls*, in battlefield order — strictly conservative.
- **Deputy of Acquittals** cannot express "**another** target creature":
  `TargetSpec` is declared statically at `init()`. The picker offers
  every creature you control; the effect declines to bounce itself.
- **Slithermuse** auto-picks the opponent holding the most cards. Its
  evoke is now real ([#257](https://github.com/krakenhavoc/cmd_and_ctrl/pull/257)).

### Known live bug that touches this deck

**[#263](https://github.com/krakenhavoc/cmd_and_ctrl/issues/263) —
search-to-battlefield bypasses the CR 614 replacement pipeline.** Any
permanent that reaches the battlefield via a *search / fetch* effect
skips every entry replacement that would otherwise apply to it, because
`SearchLibraryForEffectWithOptions` never calls
`applyReplacementsLocked`. Twenty catalog files declare `Replacements:`
and all of them are affected; a fetched fastland enters **untapped**.
In this deck it is reachable through Solemn Simulacrum and Loyal
Warhound. *A sibling agent is fixing it right now on
`fix/search-entry-pipeline`* — treat it as known-and-being-fixed, not as
a standing limitation, and re-read the issue before writing anything
that depends on the search path.

### One engine finding worth keeping

`containsFoldASCII` folds **only the haystack**, so a capitalised needle
silently matches nothing. `IsBasicLandExcept("")` is the sharper trap:
an empty needle makes the check true for every card, so the negation
matches nothing at all — a fetch predicate written that way fails
silently rather than loudly. Every existing caller happens to pass a
lowercase literal, so nothing is broken today; the contract is
documented on `IsBasicLandWithSubtype`.

## Blocked, by machinery needed

**49 of the 68 nonland entries.** Groups **overlap** — Aang, Swift
Savior alone is in four of them — so the group sizes below do not sum to
49 and are not meant to. Each group says what would unblock it and what
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
  than your hand, draw a card." **Correction:** the previous revision
  said this shape was blocked because `EventCast` carries no source zone.
  The zone is stamped on the stack item at announce
  (`StackItem.CastFromZone`,
  [stack.go:150](../../server/internal/game/stack.go), written at
  [mutations.go:661](../../server/internal/game/mutations.go)), the
  `EventCast` emit happens *after* that at `mutations.go:709`, and
  `g.StackItemForEffect(ev.CardID)`
  ([effect_api.go:37](../../server/internal/game/effect_api.go)) reads it
  back — `StackMeta` is keyed by the spell's own instance ID. Wash Away
  already targets on this fact via the `CastFromOwnersHand()` predicate.
- **Wan Shi Tong, Librarian** — `{X}{U}{U}`, X +1/+1 counters, "draw half
  X rounded down", and an opponent-searches-their-library trigger.
  X costs are S20 sub-PR 3; the search trigger is Archivist of Oghma's.

### Airbend — 6 cards, three shared pieces still missing

Aang (all three), Appa Steadfast Guardian, Monk Gyatso, Avatar's Wrath.
"Exile it. While it's exiled, its owner may cast it for {2} rather than
its mana cost."

What already exists: `Card.ExilePlay` is a durable per-card permission
naming a holder, `castSourceZoneLocked` accepts `exile`, the grant
survives `Clone`, and **the exile half of "exile a target permanent" now
exists too** (`ExileTarget`, `Flicker` —
[primitives.go:138, :185](../../server/internal/cards/effects/primitives.go)).

What is still missing:

1. **An instance-scoped alternative cost.** S22's alternative costs
   ([alternative_cost.go](../../server/internal/cards/effects/alternative_cost.go))
   are declared on a card's own `Spec` and claimed by key at cast time
   (`AlternativeCostByKey(oracleID, key)`). Airbend prices a cost onto
   *whatever card it exiled*, which that lookup cannot express —
   `ExilePlayPermission` has `CastOnly` and `AnyColor` and no cost
   override at all
   ([exile_play.go:24–47](../../server/internal/game/exile_play.go)).
   This is the expensive piece, and the S22 work did **not** deliver it.
2. **A permission with no expiry.** `Active()` gates on
   `turn <= UntilTurn` ([exile_play.go:51](../../server/internal/game/exile_play.go));
   airbend is "while it's exiled", unbounded. Small.
3. **A grant on exile from the battlefield.** The only granting API is
   `ExileTopWithPermissionForEffect`
   ([exile_play.go:85](../../server/internal/game/exile_play.go)), which
   takes cards off a library's top. Now that `ExileTarget` exists this is
   plumbing rather than design.

**Doing all three unlocks exactly one of the six** — Aang, the Last
Airbender, and even that needs one more thing (below). The others each
carry an independent blocker:

- **Monk Gyatso** — "whenever another creature you control becomes the
  target of a spell or ability". There is still **no targeting event of
  any kind** in `events.go`. Blocked regardless of airbend.
- **Appa, Steadfast Guardian** — **correction:** its second clause
  ("whenever you cast a spell from exile") is **no longer blocked**; see
  Vega above. Appa is blocked on airbend alone.
- **Aang, Swift Savior** — transform DFC, waterbend {8}, and it airbends
  a *spell* off the stack rather than a permanent.
- **Avatar's Wrath** — mass airbend, self-exile on resolution, and a
  continuous "opponents can't cast spells from anywhere other than their
  hands" restriction that would have to outlive its source.
- **Aang, Airbending Master** — experience counters (small, below) and
  "leave the battlefield *without dying*", which `EventLTB` can express
  today via `ev.NewZone != ZoneGraveyard`
  ([events.go:122–127](../../server/internal/game/events.go)).
- **Aang, the Last Airbender** — "gains lifelink **until end of turn**",
  which nothing in the layer engine can represent (below).

### Continuous effects with a duration — "until end of turn" (4)

**New group, and it is not small.** Every continuous effect in the engine
is sourced from a permanent on the battlefield:
`activeStaticAbilitiesLocked` walks `g.Battlefield` and looks up
`CatalogStaticAbilities(src.OracleID)`
([layers.go:150](../../server/internal/game/layers.go)). There is no
floating effect, no duration field, and nothing that expires at end of
turn. Grep for "until end of turn" in `server/internal/` and the only
hits are comments and the *replacement* pipeline's turn-scoped slot.

Affected here: **Aang, the Last Airbender** (lifelink UEOT), **Ambrosia
Whiteheart** (landfall +1/+0 UEOT — its ETB bounce works today), **Katara,
Water Tribe's Hope** (base P/T X/X UEOT), **The Wind Crystal** (mass
flying + lifelink UEOT). It also blocks two of The Wandering Emperor's
three loyalty abilities, and it is the reason a pump spell has never
appeared in the catalog.

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

### Tap permanents as a cost — convoke and waterbend (6)

Clever Concealment and The Wandering Rescuer (convoke); Aang Swift
Savior, Katara Water Tribe's Hope, The Legend of Kuruk, The Unagi of
Kyoshi Island (waterbend). One cost component serves both — waterbend is
convoke with a different name and no colour clause. Waterbender's
Restoration already shipped *without* it, which is
[#259](https://github.com/krakenhavoc/cmd_and_ctrl/issues/259); *the fix
is in flight on `feat/tap-as-cost`*, and when it lands this group is the
natural follow-on.

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

- **hexproof** — Lotus Field (land), Stoic Sphinx, The Wandering Rescuer.
  Nothing in `targets.go` consults it; S20 explicitly deferred
  hexproof / shroud / protection as target-legality modifiers.
- **indestructible** — Thassa, The Mind Stone, The Seriema. Only
  mentioned in comments explaining that *sacrifice* ignores it.
- **ward** — The Unagi of Kyoshi Island. No implementation.
- **prowess** — Ty Lee, Chi Blocker. No implementation.

The static half (printing the word) is cheap; the behaviour is the work.

### One-offs

Herald of Eternal Dawn (can't-lose / can't-win replacement), Mandate of
Peace (end the combat phase + a cast restriction), Clever Concealment
(phasing — no implementation anywhere), Rabble Rousing (hideaway, plus
"whenever you attack with **one or more** creatures", which over-fires
against per-creature `EventAttack` — the CR 603.1 batching gap
`events.go` documents at `:162`), The Seriema (station), Misleading
Signpost (re-select an attacker's target), Hullbreaker Horror (can't be
countered, plus returning a **spell** from the stack to hand — no
primitive does that), Venser, Shaper Savant (same spell-bounce gap; the
permanent half works today), Thassa, Deep-Dwelling (devotion — no
implementation; its end-step flicker and its `{3}{U}` tap ability are
both writable now), Enduring Curiosity (the graveyard return exists; "it
returns as an enchantment, not a creature" needs a continuous effect
conditioned on *how* the permanent got there), Ty Lee, Chi Blocker
(prowess + a "doesn't untap for as long as you control this" lockdown),
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
last three in [#258](https://github.com/krakenhavoc/cmd_and_ctrl/pull/258),
each with the simplification noted above). The two basics entries (28
cards) need no registration: a card with a basic-land subtype gets a
synthetic mana ability server-side. That leaves **2 blocked**:

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

Re-sequenced after the S22 merges. Cast-from-exile, attack triggers,
flicker + delayed triggers and the first tranche of alternative costs are
all **done**; what is left is different work than the last revision
assumed.

1. **Write the four cards that are already writable** — Phelia, The
   Mighty Thor, Vega, Wan Shi Tong. No engine work at all, and Phelia is
   one of the deck's headline cards. Do this first because it is the
   only item here with no dependency on anyone else.
2. **Two one-accessor additions**, each of which unlocks a card
   outright: an unlocked player-counter accessor (→ Katara, Waterbending
   Master) and a per-turn draw tally alongside `CastTally` (→ Faerie
   Mastermind).
3. **Tap permanents as a cost** — convoke + waterbend, 6 cards, and it
   closes [#259](https://github.com/krakenhavoc/cmd_and_ctrl/issues/259)
   rather than leaving a card in the catalog stronger than printed.
   *In flight on `feat/tap-as-cost`.*
4. **Durations on continuous effects** ("until end of turn") — 4 cards
   here, two of The Wandering Emperor's three loyalty abilities, and
   every pump spell that has never been attempted. The most reusable item
   on the list.
5. **Airbend** — the instance-scoped alternative cost, the no-expiry
   permission, and the grant-on-exile-from-battlefield. Unlocks Aang, the
   Last Airbender (with item 4); the other five each need something
   further.
6. **Multi-face cards** — 7 cards, an open in-app bug
   ([#265](https://github.com/krakenhavoc/cmd_and_ctrl/issues/265)), and
   it touches import, protocol and client. A sprint of its own.
7. **Cost reduction (S28)** and **the remaining alt-cast costs (S29)** —
   4 and 5 cards respectively, already scheduled.

Two small additions worth folding into whichever PR is nearby: a
targeting event (Monk Gyatso, and nothing else in this deck — but
"becomes the target of" is a common Commander clause), and a
spell-bounce primitive for returning a spell from the stack to its
owner's hand (Venser, Hullbreaker Horror).
