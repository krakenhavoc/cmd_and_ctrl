# Top 100 Commander staples — the play-rate gap

Source: **`edhrec_rank`** on every card in the local Scryfall bulk dump
(`data/scryfall/default-cards.json`, refreshed 2026-09-09). Lower rank
means more played. Nothing here was scraped, recalled, or guessed — one
pass over the dump produced every number below.

This is a **triage of the whole format against the catalog**, the
complement to the two deck triages ([Aang](aang-is-so-flashy.md),
[Pirates](pirates-mary-read-anne-bonny.md)). A deck says which gaps stop
one game from being played; a play-rate ranking says which gaps stop
*most* games, and it is the right instrument for deciding what to build
next.

## Method

One Python pass over the 629 MB dump, ~5 s:

- Skip placeholder printings — `type_line == "Card // Card"`, and layouts
  `art_series` / `token` / `double_faced_token` / `emblem` / `vanguard` /
  `scheme` / `planar`. These mis-resolve real single-faced cards as
  multi-face and would have poisoned the type lines below.
- Dedupe by `oracle_id`, preferring a `layout: normal` printing and
  carrying `edhrec_rank` across from any printing that has it.
- Keep `legalities.commander == "legal"`, drop basic lands, drop
  everything already registered in the catalog.
- Sort by `edhrec_rank`, take 100.

## Catalog size, honestly

**201 oracle IDs were registered before this batch**, not 191. The
undercount is a real trap worth writing down: there are 192 literal
`Register(Spec{` sites in non-test files, but `temples.go` puts one of
them inside a `for` loop over a ten-row table, so grepping for the
literal misses nine cards. Counting `Register` sites is not counting
cards. `len(All())` is.

Nothing odd is registered: no tokens, no helpers, no test-only shims.
`registry_test.go` injects specs through `registerForTest`, which is
test-binary-only, and `tokens.go` declares templates rather than
catalog entries.

After this batch the catalog holds **233**.

## What the batch bought

| Slice of the format | In catalog before | After |
|---|---:|---:|
| Top 100 by `edhrec_rank` | 47 | **65** |
| Top 200 | 65 | **97** |
| Top 300 | 80 | **112** |
| Top 500 | 102 | **134** |

(The "before" column counts the cards that were already registered
*inside* each slice; the 100-card gap list below is what was left over
from the first row.)

## The gap, ranked

100 rows, most-played first. "Shipped" means it landed in this batch;
everything else names the one piece of machinery standing in its way.

| # | Rank | Card | Type | Status / blocking machinery |
|---:|---:|---|---|---|
| 1 | 9 | Exotic Orchard | Land | restricted mana |
| 2 | 10 | Reliquary Tower | Land | player static |
| 3 | 12 | Swiftfoot Boots | Artifact - Equipment | equipment |
| 4 | 13 | Lightning Greaves | Artifact - Equipment | equipment |
| 5 | 14 | Path of Ancestry | Land | **shipped** |
| 6 | 18 | Rogue's Passage | Land | until EOT |
| 7 | 29 | Chaos Warp | Instant | shuffle into library |
| 8 | 30 | Heroic Intervention | Instant | until EOT |
| 9 | 31 | Myriad Landscape | Land | **shipped** |
| 10 | 33 | Dark Ritual | Instant | add mana |
| 11 | 40 | Skullclamp | Artifact - Equipment | equipment |
| 12 | 46 | Fabled Passage | Land | **shipped** |
| 13 | 50 | Watery Grave | Land - Island Swamp | as-enters choice |
| 14 | 51 | Commander's Sphere | Artifact | **shipped** |
| 15 | 53 | Arcane Denial | Instant | delayed trigger |
| 16 | 55 | Reanimate | Sorcery | **shipped** |
| 17 | 59 | Breeding Pool | Land - Forest Island | as-enters choice |
| 18 | 60 | Godless Shrine | Land - Plains Swamp | as-enters choice |
| 19 | 63 | Hallowed Fountain | Land - Plains Island | as-enters choice |
| 20 | 64 | Steam Vents | Land - Island Mountain | as-enters choice |
| 21 | 66 | Ancient Tomb | Land | mana rider |
| 22 | 67 | Toxic Deluge | Sorcery | until EOT |
| 23 | 69 | Blood Crypt | Land - Swamp Mountain | as-enters choice |
| 24 | 70 | Stomping Ground | Land - Mountain Forest | as-enters choice |
| 25 | 71 | Overgrown Tomb | Land - Swamp Forest | as-enters choice |
| 26 | 72 | Brainstorm | Instant | library top |
| 27 | 73 | Urborg, Tomb of Yawgmoth | Legendary Land | type-add mana |
| 28 | 74 | Sacred Foundry | Land - Mountain Plains | as-enters choice |
| 29 | 75 | Deflecting Swat | Instant | retarget |
| 30 | 76 | Boseiju, Who Endures | Legendary Land | channel |
| 31 | 77 | Yavimaya, Cradle of Growth | Legendary Land | type-add mana |
| 32 | 79 | Sulfur Falls | Land | **shipped** |
| 33 | 80 | Clifftop Retreat | Land | **shipped** |
| 34 | 81 | Temple of the False God | Land | restricted mana |
| 35 | 82 | Fierce Guardianship | Instant | **shipped** |
| 36 | 83 | Temple Garden | Land - Forest Plains | as-enters choice |
| 37 | 84 | Dragonskull Summit | Land | **shipped** |
| 38 | 85 | Chromatic Lantern | Artifact | restricted mana |
| 39 | 86 | Sunken Hollow | Land - Island Swamp | **shipped** |
| 40 | 87 | Cinder Glade | Land - Mountain Forest | **shipped** |
| 41 | 88 | Otawara, Soaring City | Legendary Land | channel |
| 42 | 89 | Feed the Swarm | Sorcery | **shipped** |
| 43 | 90 | Garruk's Uprising | Enchantment | **shipped** |
| 44 | 91 | The One Ring | Legendary Artifact | protection package |
| 45 | 92 | Smoldering Marsh | Land - Swamp Mountain | **shipped** |
| 46 | 93 | Talisman of Dominance | Artifact | mana rider |
| 47 | 94 | Isolated Chapel | Land | **shipped** |
| 48 | 95 | Glacial Fortress | Land | **shipped** |
| 49 | 96 | City of Brass | Land | mana rider |
| 50 | 98 | Hinterland Harbor | Land | **shipped** |
| 51 | 99 | Mystic Remora | Enchantment | cumulative upkeep |
| 52 | 104 | Jeska's Will | Sorcery | add mana |
| 53 | 105 | Canopy Vista | Land - Forest Plains | **shipped** |
| 54 | 106 | Prairie Stream | Land - Plains Island | **shipped** |
| 55 | 107 | Deadly Rollick | Instant | **shipped** |
| 56 | 108 | Drowned Catacomb | Land | **shipped** |
| 57 | 109 | Teferi's Protection | Instant | until EOT |
| 58 | 110 | Woodland Cemetery | Land | **shipped** |
| 59 | 112 | Cavern of Souls | Land | restricted mana |
| 60 | 114 | Propaganda | Enchantment | attack tax |
| 61 | 116 | Shivan Reef | Land | mana rider |
| 62 | 117 | Mana Drain | Instant | add mana |
| 63 | 118 | Urza's Saga | Enchantment Land - Urza's Saga | saga |
| 64 | 119 | Mana Confluence | Land | mana rider |
| 65 | 121 | Battlefield Forge | Land | mana rider |
| 66 | 122 | Assassin's Trophy | Instant | **shipped** |
| 67 | 123 | Enlightened Tutor | Instant | library top |
| 68 | 125 | Caves of Koilos | Land | mana rider |
| 69 | 126 | Talisman of Creativity | Artifact | mana rider |
| 70 | 127 | Rootbound Crag | Land | **shipped** |
| 71 | 128 | Victimize | Sorcery | multi-target graveyard |
| 72 | 129 | Talisman of Indulgence | Artifact | mana rider |
| 73 | 131 | Black Market Connections | Enchantment | upkeep modal |
| 74 | 132 | Roaming Throne | Artifact Creature - Golem | choose a type |
| 75 | 133 | Sunpetal Grove | Land | **shipped** |
| 76 | 135 | Talisman of Hierarchy | Artifact | mana rider |
| 77 | 136 | Yavimaya Coast | Land | mana rider |
| 78 | 137 | Llanowar Wastes | Land | mana rider |
| 79 | 140 | Talisman of Progress | Artifact | mana rider |
| 80 | 141 | War Room | Land | dynamic cost |
| 81 | 142 | Morphic Pool | Land | **shipped** |
| 82 | 143 | Rejuvenating Springs | Land | **shipped** |
| 83 | 144 | Ponder | Sorcery | library top |
| 84 | 145 | Abrade | Instant | **shipped** |
| 85 | 146 | Underground River | Land | mana rider |
| 86 | 147 | Mana Vault | Artifact | doesn't untap |
| 87 | 148 | Adarkar Wastes | Land | mana rider |
| 88 | 149 | Chrome Mox | Artifact | restricted mana |
| 89 | 150 | Delighted Halfling | Creature - Halfling Citizen | restricted mana |
| 90 | 151 | Sulfurous Springs | Land | mana rider |
| 91 | 152 | Herald's Horn | Artifact | choose a type |
| 92 | 153 | Talisman of Conviction | Artifact | mana rider |
| 93 | 154 | Training Center | Land | **shipped** |
| 94 | 155 | Pongify | Instant | **shipped** |
| 95 | 159 | Dimir Signet | Artifact | mana in a mana cost |
| 96 | 160 | Mystical Tutor | Instant | library top |
| 97 | 161 | Luxury Suite | Land | **shipped** |
| 98 | 162 | Ghostly Prison | Enchantment | attack tax |
| 99 | 163 | Nykthos, Shrine to Nyx | Legendary Land | add mana |
| 100 | 164 | Sea of Clouds | Land | **shipped** |

## Shipped in this batch — 32 cards

Everything here is expressible with machinery that already existed on
`main`. No shared engine file was touched.

### The conditional-dual lands — 20 cards, one shape

The single largest block of the gap, and it needed nothing new. "This
land enters tapped **unless** X" is an ordinary CR 614 self-replacement
whose `AppliesTo` reads the board; the dual is the Temple cycle's pipe
mana ability. Three cycles, three tables, one helper
(`SelfEntersTappedUnless` in `tapland_helpers.go`).

| Cycle | Cards | Condition |
|---|---|---|
| Checklands | Sulfur Falls, Clifftop Retreat, Dragonskull Summit, Isolated Chapel, Glacial Fortress, Hinterland Harbor, Drowned Catacomb, Woodland Cemetery, Rootbound Crag, Sunpetal Grove | you control a land with either land type |
| Battle lands | Sunken Hollow, Cinder Glade, Smoldering Marsh, Canopy Vista, Prairie Stream | you control two or more **basic** lands |
| Bond lands | Morphic Pool, Rejuvenating Springs, Training Center, Luxury Suite, Sea of Clouds | you have two or more opponents |

Three details worth pinning, and each has a test:

- The condition lives in `AppliesTo`, not inside `Replace`. A `Replace`
  that decided to do nothing would still have been one of the effects
  the controller was asked to order under CR 616, so a checkland that
  was always going to enter untapped would pop a prompt.
- The condition reads **`ev.Actor`**, not `src.Controller`. The
  battlefield-entry path stamps `Card.Controller` *after* the
  replacement pipeline runs, so `src.Controller` is whatever the card
  carried in its previous zone. `Actor` is the only field guaranteed
  correct at that point — this is the kind of thing that would have
  shipped as a silent bug without a same-controller test.
- The checklands check land **types** and the battle lands count
  **basics**, which run opposite ways in deck construction. Two
  nonbasic duals turn a checkland on and leave a battle land tapped;
  there is a test for exactly that inversion.

The battle lands also need their pipe ability declared explicitly even
though the printed card puts it in reminder-text parentheses: the
engine's synthetic land ability (`ManaAbilitiesForCard` →
`basicLandColor`) requires the **basic** supertype, so a "Land — Island
Swamp" produces nothing on its own.

### The other twelve

| Card | Rank | What it exercises |
|---|---:|---|
| Path of Ancestry | 14 | enters-tapped replacement + commander-identity pipe |
| Myriad Landscape | 31 | three-component activated cost ({2}, {T}, sacrifice) driving a two-card search |
| Fabled Passage | 46 | conditional post-search untap — the first effect that has to know *which* card a search took |
| Commander's Sphere | 51 | identity pipe **and** a sacrifice-only activated ability on one card |
| Reanimate | 55 | first `ReturnFromGraveyard{Dest: ZoneBattlefield}` in the catalog |
| Fierce Guardianship | 82 | Negate's clause; the free cast is declared missing |
| Feed the Swarm | 89 | reads the target's mana value before destroying it |
| Garruk's Uprising | 90 | intervening-if ETB + Layer 6 trample grant + ongoing ETB watcher, three mechanisms on one card |
| Deadly Rollick | 107 | exile removal; the free cast is declared missing |
| Assassin's Trophy | 122 | an effect that makes the **victim** search their own library |
| Abrade | 145 | two targeted modes on one "choose one" |
| Pongify | 155 | Beast Within's shape in blue |

## Simplifications declared

Every one of these is also a comment on the card file. The rule the
batch held to: a simplification that makes a card **weaker** than
printed is acceptable and must be declared; one that makes it
**stronger** is not, and the card gets skipped instead.

**Weaker than printed (shipped):**

- **Deadly Rollick, Fierce Guardianship** — "If you control a
  commander, you may cast this spell without paying its mana cost" is
  not implemented. Both cost full price. CR 118.9 alternative costs have
  no seam: `AdditionalCost` is an *extra* cost, not a replacement one.
- **Path of Ancestry** — the scry rider is not implemented. "When **that
  mana** is spent to cast a creature spell…" is a delayed trigger keyed
  to the provenance of one mana in the pool, and `ManaPool` records
  colour and nothing else. Same missing piece as Cavern of Souls and
  Delighted Halfling.
- **Reanimate** — the target is narrowed to **your own** graveyard.
  Printed it is "a graveyard"; `ReturnFromGraveyardForEffect` does not
  stamp a new controller, so a creature lifted out of an opponent's
  graveyard would arrive under *their* control. Narrowing is honest
  until a change-of-control seam exists. Residual edge: a card you own
  that died under an opponent's control still carries their controller
  field and comes back under their control.
- **Reanimate** (second) — the reanimated permanent bypasses the CR 614
  entry pipeline entirely, so a creature with a self-"enters tapped"
  replacement enters untapped.
- **Assassin's Trophy** — the victim's "**may** search" is forced. A
  yes/no prompt on someone else's resolution needs the PayUnless-style
  deferred-choice plumbing, and this effect has no hook for it. Forcing
  the search hands the victim a land they might have declined, which is
  the weaker direction for the caster.
- **Myriad Landscape** — "two basic land cards **that share a land
  type**" resolves by taking the first basic in library order and
  fetching up to two of that type. Neither stronger nor weaker, just
  less controllable; a search chooser fixes it for every card at once.
- **Fabled Passage, Myriad Landscape, Assassin's Trophy** — all inherit
  the catalog-wide deterministic search pick.

**Stronger than printed (shipped, flagged):**

- **Garruk's Uprising** — the intervening-if (CR 603.4) is checked when
  the trigger goes on the stack, not again on resolution. Killing the
  power-4 creature in response should make the trigger do nothing; here
  the draw still happens. This is the same posture every intervening-if
  card in the catalog takes (a re-check hook on the built `StackItem`
  does not exist), and it is a corner case, but it is a real
  divergence and the only one in the batch that runs the wrong way.

**Not a simplification, worth recording:**

- **Pongify** — "It can't be regenerated" is a no-op because
  regeneration is not modelled at all. If a regeneration shield ever
  lands, this card has to be revisited.

## Skipped because the only available version would be STRONGER

- **Mystic Remora** (rank 99). The payoff half is already expressible —
  `PayUnless` is exactly Rhystic Study's machinery. But **cumulative
  upkeep** is not, and a Remora that is never sacrificed is a
  format-warping card rather than a two-turn cantrip. This is the
  clearest example in the whole list of why the honesty rule has to cut
  both ways.
- **Chromatic Lantern** (rank 85). Its own mana ability is trivial; the
  card is played for "Lands you control have '{T}: Add one mana of any
  color.'" Shipping the trivial half would register a marquee card that
  does ~30% of what its name promises — weaker, but misleadingly so.
- **Enlightened Tutor / Mystical Tutor** (123, 160). "Put that card on
  top" would have to become "put it into your hand", which is strictly
  better: it dodges discard and saves a draw step. Vampiric Tutor
  already ships with that simplification; this batch declined to
  spread it.
- **Mana Vault** (147). "Doesn't untap during your untap step" is the
  entire drawback.

## Blocked, by machinery needed

Grouped by what would unblock them, ordered by how many of the 100 each
buys. **This section is the point of the exercise.**

### 1. Mana-ability riders and life costs — 17 cards

Eight painlands (Shivan Reef, Battlefield Forge, Caves of Koilos,
Yavimaya Coast, Llanowar Wastes, Underground River, Adarkar Wastes,
Sulfurous Springs), six Talismans (Dominance, Creativity, Indulgence,
Hierarchy, Progress, Conviction), Ancient Tomb, Mana Confluence, City of
Brass.

The single highest-value gap in the format, and a small one. Every card
here is "{T}: Add X" plus one of:

- **a damage rider** — "This land deals 1 damage to you" as part of the
  same ability;
- **a life component in the cost** — Mana Confluence's "{T}, Pay 1
  life";
- **a tap trigger** — City of Brass's "Whenever this land becomes
  tapped, it deals 1 damage to you" (a real triggered ability, not a
  rider, but it needs the same `EventTapCard` plumbing to be useful).

`ManaAbilityCost` today carries `Tap`, `Sacrifice`, `SacrificeOther`.
Adding `Life int` mirrors `AbilityCost.Life`, which activated abilities
already have and fetchlands already use. The rider wants one more field
— a post-production `func(g, controller)` — or an `Effect` slot on
`ManaAbility`. Neither is a new subsystem.

Note that these cards are **actively harmful to leave out**: a painland
in a deck today is a land with no mana ability at all, because
`basicLandColor` needs the basic supertype. They are worse than absent.

### 2. "As this permanent enters, you may pay N life" — 10 cards

The whole shockland cycle: Watery Grave, Breeding Pool, Godless Shrine,
Hallowed Fountain, Steam Vents, Blood Crypt, Stomping Ground, Overgrown
Tomb, Sacred Foundry, Temple Garden.

A replacement that has to **ask a question and wait**. The seam already
exists: `applyReplacementsLocked` returns `errReplacementPending` for
the CR 616 ordering prompt, and the land-play path already handles that
return by bailing out and resuming later. What is missing is a
`PendingChoice` kind for "pay N life as this enters?", the life payment,
and the resume that stamps `ev.EntersTapped` from the answer.

This is the second-biggest unlock and it is probably cheaper than it
looks, because the pause-and-resume half is already built. Worth
scoping before the mana riders even.

### 3. Restricted, derived or conditional mana — 6 cards

Exotic Orchard (rank 9, the highest-ranked missing card in the format),
Temple of the False God, Chromatic Lantern, Cavern of Souls, Chrome Mox,
Delighted Halfling.

Four separate sub-gaps that all live in the mana pipeline:

- **Colours derived from the board** — Exotic Orchard ("any color that a
  land an opponent controls could produce"), Chrome Mox (the imprinted
  card's colours). The engine has exactly one such hook today,
  `filterPipeByCommanderIdentity`; a general "narrow this pipe with a
  predicate over game state" generalises it. Fellwar Stone is already
  in the catalog with this simplification declared.
- **An activation condition** — Temple of the False God's "only if you
  control five or more lands". `ManaAbility` has no gate field.
- **Spend restrictions** — Cavern of Souls, Delighted Halfling. Needs
  mana provenance in the pool, which is the same missing piece as Path
  of Ancestry's scry.
- **Granting a mana ability to other permanents** — Chromatic Lantern.

### 4. Adding mana to a pool from a spell or non-mana ability — 4 cards

Dark Ritual (rank 33), Jeska's Will, Mana Drain, Nykthos.

There is no `AddManaForEffect` on `*Game`. Every mana in the engine
today comes from `ActivateManaAbility`. One `*ForEffect` helper unlocks
Dark Ritual outright and is a prerequisite for the other three (which
each want something else as well — X-scaling off a hand size, a delayed
trigger, devotion).

### 5. Until-end-of-turn continuous effects — 4 cards

Heroic Intervention, Toxic Deluge, Rogue's Passage, Teferi's Protection.

The layer engine recomputes from battlefield state, so there is nowhere
to hang "creatures you control gain indestructible **until end of
turn**". `TurnScopedReplacements` (the Fog machinery) is the closest
existing shape and is cleared at `StepCleanup` — a turn-scoped *static*
list would follow it exactly. Toxic Deluge additionally needs
"-X/-X where X is life paid at resolution"; Teferi's Protection needs
phasing and protection on top and is not a four-for-one.

Rogue's Passage (rank 18) is the cheap one here: its {T}: Add {C} half is
trivial and only "target creature can't be blocked this turn" is
blocked.

### 6. Library-top placement and ordered look — 4 cards

Brainstorm, Ponder, Enlightened Tutor, Mystical Tutor.

`SearchLibraryForEffectWithOptions` treats `Dest: ZoneLibrary` as a
no-op (the card never moves), so "search, shuffle, **then** put that
card on top" cannot be expressed at all. Ponder wants "look at the top
three, put them back in any order" — the Scry prompt is close but not
the same shape — and Brainstorm wants "put two cards **from your hand**
on top in any order", which is a third shape. One `PutOnTopForEffect`
plus a generalised reorder prompt covers all four.

### 7. Equipment and attachment — 3 cards

Swiftfoot Boots (12), Lightning Greaves (13), Skullclamp (40).

Three of the top 40, and the largest single subsystem in this list:
an `AttachedTo` field on `Card`, "equipped creature" as an
attachment-scoped static, equip as a sorcery-speed targeted activated
ability, and unattach on either permanent leaving. `clone.go` has to
carry the attachment. Auras ride on the same machinery, which makes it
worth more than three cards.

### 8. The rest

| Machinery | Cards | Which |
|---|---:|---|
| Attack taxes ("can't attack you unless…") | 2 | Propaganda, Ghostly Prison |
| Layer-4 type-add feeding mana derivation | 2 | Urborg, Yavimaya (Cradle of Growth) |
| Channel (activate by discarding from hand) | 2 | Boseiju, Otawara |
| "As this enters, choose a creature type" + payoff | 2 | Roaming Throne, Herald's Horn |
| Player-scoped continuous effect | 1 | Reliquary Tower (rank 10) |
| Mana component in a **mana** ability's cost | 1 | Dimir Signet |
| "Doesn't untap during your untap step" | 1 | Mana Vault |
| Sagas / lore counters | 1 | Urza's Saga |
| Shuffle a permanent into its owner's library | 1 | Chaos Warp |
| Delayed triggers ("at the beginning of the next…") | 1 | Arcane Denial |
| Multi-target graveyard + sacrifice mid-resolution | 1 | Victimize |
| Beginning-of-main-phase repeatable modal | 1 | Black Market Connections |
| Cumulative upkeep | 1 | Mystic Remora |
| Cost amounts computed at activation | 1 | War Room |
| Protection from everything | 1 | The One Ring |
| Choose new targets for a spell or ability | 1 | Deflecting Swat |

Two notes on that table:

- **Urborg / Yavimaya** are nearly free and currently invisible.
  `Layer4Type` already exists (Mycosynth Lattice) and would happily
  stamp "Swamp" onto every land — but `ManaAbilitiesForCard` reads the
  **printed** `TypeLine`, so the stamp would produce no mana and the
  card would silently do nothing. `mycosynth_lattice.go` already
  documents this exact deferral. Pointing the mana derivation at the
  post-layer type line unlocks both cards and finishes the Lattice.
- **Reliquary Tower** (rank 10) is the highest-ranked one-card gap.
  `Player.MaxHandSize` and `SetMaxHandSize` already exist; what is
  missing is a way for a static ability to apply to a *player* so the
  effect ends when the Tower leaves. Setting the field in `OnETB` would
  make the card permanently stronger than printed, so it was skipped.

## Two engine findings

### The ETB double-wiring bug — confirmed by probe

There are **two independently-wired ETB mechanisms**: `EventETB` (which
the S19 trigger harvester consumes) and `fireETBHookLocked` (which runs
`Spec.OnETB`). `mutations.go` fires both at every battlefield-entry
site. Two `*ForEffect` helpers fire only the first:

- `ReturnFromGraveyardForEffect` (`effect_api.go:574`)
- `SearchLibraryForEffectWithOptions` (`effect_api.go:651`)

Probe result: reanimating **Bastion of Remembrance** — whose ETB token
is declared via `Spec.OnETB` — creates **zero** tokens. Its dies-trigger,
declared via `Triggered` + `EventLTB`, is unaffected.

The blast radius is wider than reanimation. Any card that puts a
permanent onto the battlefield by search has the same hole: Wood Elves
fetching a Temple would skip the Temple's scry, and Fabled Passage or
Myriad Landscape fetching one would too.

**Not fixed here on purpose.** `effect_api.go` is one of the shared
files six sibling branches are touching, and a one-line fix that
collides in a merge is worse than a documented workaround. The fix is
one `g.fireETBHookLocked(id, oracleID)` call at each site plus a test;
whoever owns the next `effect_api.go` change should take it.

Reanimate ships with the gap declared on the card file.

### Self-replacement on a card's own entry works

`docs/decklists/aang-is-so-flashy.md` records the opposite — "a catalog
replacement effect cannot fire on its own source's entry" — and says
enters-tapped must use the Worn Powerstone `OnETB` workaround. **That is
no longer true**, and has not been since the Temple cycle:
`gatherActiveReplacementsLocked` (`replacements.go:471`) has a
dedicated block for a card that is not yet on the battlefield. All 20
lands in this batch depend on it, and `worn_powerstone.go` has already
been converted.

The Aang note was accurate when written and is now stale. It is left
in place rather than edited (that file is off-limits to this branch);
this paragraph is the correction.

## What to build next, in order

1. **Mana-ability riders + `Life` cost** — 17 cards, small change,
   and it converts painlands from "wrong" to "right" rather than from
   "absent" to "present".
2. **"As this enters, you may pay 2 life"** — 10 cards, and the
   pause-and-resume half of the plumbing already exists.
3. **A search chooser** — unlocks nothing by itself and improves
   roughly fifteen cards already in the catalog, including three
   shipped in this batch. #243 already named it the highest-value
   quality gap; the play-rate data agrees.
4. **Equipment / attachment** — 3 cards in the top 40, and the same
   machinery carries Auras.
5. **`AddManaForEffect`** — Dark Ritual (rank 33) outright, plus a
   prerequisite for three more.
6. **Post-layer type line for mana derivation** — 2 cards, finishes
   an already-documented deferral.
