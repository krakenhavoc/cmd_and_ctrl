# ADR 0083 — Triggered and static abilities on a token that is not a copy

**Status:** Accepted · 2026-09-23 · S46 — permanents that change what they are
**Issues:** [#1248](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1248) (this seam),
[#521](https://github.com/krakenhavoc/cmd_and_ctrl/issues/521) (token catalog keys, the half that shipped),
[#343](https://github.com/krakenhavoc/cmd_and_ctrl/issues/343) (the Avatar set, two of the caveats this clears)
**Tracker:** [#889](https://github.com/krakenhavoc/cmd_and_ctrl/issues/889) — S46, "Permanents that change what they are"
**Numbering:** swept with the AGENTS.md §4 loop on 2026-09-23 — `git fetch origin`,
then `git ls-tree --name-only <sha> docs/decisions` over all 405 remote heads plus
`origin/develop`. The highest number present anywhere is **0082**
(`0082-casting-face-down-and-turning-face-up.md`); `0083` appears on no branch.
`0005`, `0024`, `0029` and `0030` stay permanently unused per AGENTS.md §4.

**Related:** [ADR 0010](0010-card-effect-catalog.md) (the catalog and the hook-variable
seam the engine reads it through), [ADR 0041](0041-game-persistence.md) (what a
snapshot may hold, and why a closure on an instance is not it),
[ADR 0043](0043-copy-effects.md) and [ADR 0064](0064-emblems.md) (the two synthetic
key namespaces this one copies), [ADR 0061](0061-token-creation-and-discard-are-replaceable-events.md)
(the one token-creation path), [ADR 0071](0071-designations-that-switch-abilities-on.md)
(`ActiveWhen`, which rides along for free), [ADR 0078](0078-token-art.md) (token art,
still Proposed — the reason a token's text has nowhere to come from).

---

## Context

`docs/engine-seams.md` has carried this row since the 2026-09-16 audit:

> **Triggered and static abilities on non-copy tokens** — primitive: `game.Card`
> gives a token intrinsic mana and activated abilities only, and a token that
> isn't a copy has no catalog key, so a token's own trigger ("a Pest with 'when
> this dies, gain 1 life'") or static can't be written.

Half of that is out of date and half of it is exactly right, and the difference is
the whole of this ADR.

**What already shipped.** #521 gave a token a catalog identity: `game.TokenKey`
mints `"token:treasure"`, `game.CatalogKey` falls back to `Card.TokenKey` when
there is no oracle ID (`game/effect_hooks.go`), and `PrintedValues.TokenKey`
travels with a copy so CR 707.2 keeps working in both directions. Treasure, Food,
Clue, Blood, Gold, the Powerstone, the Eldrazi Spawn and Scion, the Lander, the
Replicated Ring and the Springleaf Shapeshifter are all catalog entries today.
So `TriggersForCard`, `StaticAbilitiesForCard`, `TriggersForKey` and the rest have
been *able* to answer for a token since #521 landed. They answer with nothing,
because nothing puts anything in the slot.

**What is actually missing** is the declaration, and it is missing in a way that
is not an oversight. `token_catalog.go`'s builder returns a `game.Card`, and
`game.Card` has no `Triggered` or `Static` field — deliberately, because a
`TriggeredAbility` holds closures and a closure on an instance is what #521
deleted. ADR 0041's rule is that a snapshot holding an unserialisable
continuation is not a restore point, and before #521 a single live Treasure meant
no restore point was written at all for as long as it sat on the battlefield. So
"put the trigger on the template" is not available: it would put the bug back.

The registration is also narrower than the fallback. `token_catalog.go`'s `init`
files exactly two slots, `ManaAbilities` and `Activated`, and **panics** on a
template that declares neither — which makes a trigger-only token unregistrable
even if it had somewhere to write the trigger down.

And one layer further out, nine battlefield walks in `internal/game` still ask
`c.OracleID == ""` where they mean "does this object have a catalog entry":
`block_rules.go`, `cast_timing.go`, `cast_permission.go` (twice), `land_drops.go`,
`library_top.go`, `player_statics.go`, `effect_hooks.go`, `keywords.go`, and
`characteristic.go`'s printed-keyword read. Every one of them skips every token by
construction. `activated.go` already carries the fixed idiom and says why.

Finally, the player. A token has no printing, so `scryfall_id` is empty, no art
loads (ADR 0078 is still Proposed) and the client renders the card through
`Card.svelte`'s `.name-fallback` branch: the name on a grey rectangle. A mana or
activated ability escapes that, because it reaches the player as a row in the
right-click menu. A **trigger** has no control to hang text off. A Dragon Egg
whose dies-trigger the player cannot read is not a card, it is a surprise.

**Nine cards wait on this row.** Three cannot be written at all (Reef Worm,
Nesting Dragon, Mysidian Elder); three ship today with a caveat that names this
seam in so many words (Fable of the Mirror-Breaker, Avatar Roku, Avatar Kuruk);
and three more say the same thing in their own files without being on the row
(Beledros Witherbloom, Sedgemoor Witch, Cornered by Black Mages). The audit puts
the seam at 18 cards where it is the only core blocker and 24 where it is any
blocker.

### The rules

- **CR 111.1** — a token is a marker representing a permanent that is not
  represented by a card. Its characteristics, abilities included, are the ones
  the effect that created it specified.
- **CR 111.7** — a token that has left the battlefield still exists in its new
  zone long enough for triggered abilities to see it leave.
- **CR 704.5d** — a token in a zone other than the battlefield ceases to exist,
  as a state-based action.
- **CR 603.10** — a leaves-the-battlefield trigger is judged on the permanent's
  last known information.
- **CR 707.2** — a copy acquires the copiable values of the copied object, so a
  copy of a token that prints an ability prints the same ability; a token that
  copies a printed card prints that card's.
- **CR 611.2b** — a static ability's continuous effect ends when the object
  generating it leaves the battlefield.
- **CR 113.7a** — "this creature" on an ability is the source of that ability.

---

## Decision 1 — The declaration is one `tokenTemplate`, and it has four ability slots

A token template that carries any ability is declared once, as an
`effects.tokenTemplate`:

```go
type tokenTemplate struct {
    Slug      string     // its catalog identity: game.TokenKey(Slug)
    Card      game.Card  // the printed CHARACTERISTICS, and only those
    Mana      []game.ManaAbilityShape
    Activated []game.ActivatedAbilityShape
    Triggered []game.TriggeredAbility
    Static    []game.StaticAbility
    Text      string     // the printed ability text, verbatim
}
```

`init` projects it into a `game.CardDef` and files it in the same `defs` map cards
are filed in, under `game.TokenKey(Slug)`. The abilities are written with the
**same constructors printed cards use** — `WhenThisDies`, `WheneverThisAttacks`,
`Landfall`, `WheneverYouCast` out of `triggers_common.go`, and plain
`game.StaticAbility` values. There is no token dialect.

The four slots sit beside each other rather than two on the `Card` and two
outside it, because the split would be arbitrary and a reader would have to learn
it. Nothing about a trigger makes it less "printed on the token" than a mana
ability.

**No new field on `game.Card`, and that is the decision.** The instance keeps
carrying exactly one thing — `TokenKey`, a string — so a board holding a Dragon
Egg is as serialisable as a board holding a Treasure and the #521 restore-point
guarantee extends to every new token for free.
`TestEveryTokenTemplateWithAbilitiesIsRestorable` walks the whole list and asserts
it.

### Why the registry is a list and the slug lives on the template

The obvious shape — `map[slug]builder` — does not compile. A token template may
name **another token**: Fable's Goblin Shaman creates a Treasure, Reef Worm's Fish
creates a Whale. With the slug as the map key, the constructor
(`TreasureToken()`) has to read the map to find its builder, so the
package-level map's initialiser reaches back into the map, and Go refuses the
initialization cycle:

```
tokenTemplates → printedFableGoblinShamanToken → TreasureToken
              → tokenFromCatalog → tokenTemplates
```

Moving the slug onto the template breaks it: `tokenFromCatalog` takes the
**builder** rather than a slug, reads the slug off what it builds, and never
touches the registry. The registry becomes a plain list that `init` indexes. As a
bonus there is no second spelling of the slug for a constructor to get wrong.

### The registration rules, as a function

`checkTokenTemplate` is every rule in one place, so a test can state them instead
of provoking panics at boot. A template is refused when it declares no slug, no
abilities at all (a vanilla token is a row in `tokens_table.go` — a key resolving
to an empty entry reads as "this token's abilities were removed" everywhere
downstream), no printed text (decision 6), an oracle ID as well as a token key
(`CatalogKey` would silently prefer the oracle ID and the token's own abilities
would never be found), its own `TokenKey` on the `Card`, or a mana/activated
ability left on the instance.

---

## Decision 2 — No new dispatch. The read was never the problem

Nothing in `internal/game` grows a token arm. The trigger harvest
(`TriggersForCard`), the LTB harvest (`TriggersForKey`), the layer pass's gather
(`StaticAbilitiesForCard`) and the activation path (`ActivatedAbilitiesForCard`)
all key on `CatalogKey` / `CatalogAbilityKey`, and both have answered with the
token key since #521. Widening the registration is the entire change.

Two riders fall out of that and are worth stating because they would each have
been a decision if the dispatch were new:

- **ADR 0071's `ActiveWhen` gate** applies to a token's abilities exactly as it
  does to a card's, because `activeOnly` sits inside the shared accessor.
- **CR 613.1f ability removal** applies too, and through the same
  `CatalogAbilityKey`: a token under Dress Down loses its trigger like anything
  else.

---

## Decision 3 — The gate is "has no catalog entry", never "has no oracle ID"

The nine walks named in the context are changed to compute the key once and skip
on the empty key:

```go
key := CatalogAbilityKey(*c)
if key == "" {
    continue
}
```

This is behaviour-neutral today — no token template declares a block rule, a cast
permission or a player keyword — and it is not cosmetic. It is the difference
between a slot that answers for tokens and a slot that silently does not, and the
next person to add one should not have to notice.

The empty key is the right spelling of the question in both directions:
`CatalogKey` returns `""` for an uncatalogued card **and** for a CR 708.2a
face-down permanent, so one predicate covers both and neither needs its own arm.

`TestNoBattlefieldWalkGatesOnAnEmptyOracleID` scans the package's AST and fails on
a new one, with an allowlist of five sites where an oracle ID genuinely is the
question (`CatalogKey`'s own face-key arm, two "is there a Scryfall printing
behind this" reads, two stack-item restores gated on `StackItemSpell`). A source
scan rather than nine behaviour tests, because the failure mode is a tenth walk
written next sprint with the old idiom copied from its neighbour.

---

## Decision 4 — The dies path: the token key joins the CR 603.10 identity snapshot

A token that dies really reaches its owner's graveyard and stays there until the
next state-based check removes it (CR 111.7, then CR 704.5d in
`game/token_existence.go`). The LTB harvest runs in between, off
`harvestLTB`'s `CatalogKey(source)` — so a token's dies-trigger fires from the
graveyard with no special case.

What it needed was one field. `triggerIdentityLKI` — the small identity snapshot
taken before the exit so a permanent whose identity changed on the way out still
fires the right trigger — recorded `OracleID`, `ActiveFace` and `AttachedTo`, and
`withLastKnownTriggerIdentityLocked` wrote all three back. For a token the
identity is `TokenKey`, so restoring the snapshot **blanked** it: a token that was
a copy of something when it died would come back with no key and no trigger. The
field is now recorded and restored with the others.

The ordering in `mutations.go`'s SBA loop was already right and needed no change:
the token sweep runs last, after the destruction and sacrifice passes, precisely
so a token that died in this pass has already had its triggers harvested.

---

## Decision 5 — CR 707.2 needs nothing, and that is the test

`PrintedValues.TokenKey` has been a copiable value since #521, so a copy of a
token that prints a trigger prints the same trigger, and a token that copies
Llanowar Elves prints Llanowar Elves' abilities because `CatalogKey` prefers a
real oracle ID. Both directions are asserted rather than assumed
(`TestACopyOfATokenCarriesItsTriggerAndStatic`): a rule that holds by construction
is exactly the kind that stops holding silently.

---

## Decision 6 — The token's printed text goes on the wire, because nothing else can say it

`CardDef.TokenText` carries a token template's printed ability text verbatim;
`game.TokenTextForCard` derives it per read; `CardView.token_text` ships it; and
`Card.svelte` renders it under the name in the fallback branch, with the full text
in the `title` so a clamped card is still readable.

This is `EmblemDef.Text`'s argument one object over. An emblem has no printing
either, and its text is on the wire for the same reason: there is no oracle text
to fetch and a trigger's `Label` is a log line, not card text
("Dragon Egg — create a 2/2 red Dragon" is not what the card says).

Three consequences of it being a **derived catalog read** rather than stored
state:

- Fixing a token's wording reaches a game already in progress.
- It is keyed on `CatalogKey`, not `CatalogAbilityKey`, so a token silenced by
  Dress Down still shows what it prints — the same reason the client keeps
  rendering a silenced card's oracle text.
- CR 707.2 rides along: a token copying a printed card has an oracle ID, answers
  `""` here, and renders that card's own printing.

The field is **required** of every registered template, enforced at boot. A token
that does something the player cannot read is the failure this decision exists to
prevent, and a template is the only place that knows what the token prints.

---

## Decision 7 — Two cards that print the same token share one template

Beledros Witherbloom and Sedgemoor Witch both print "a 1/1 black and green Pest
creature token with 'When this token dies, you gain 1 life'". That is one object,
so it is one template under one slug, and both cards call `PestToken()`. Cornered
by Black Mages and Mysidian Elder share the Wizard the same way.

The slug is the token's printed name, lowercased and hyphenated, and it is a
stable **on-disk** identity: a snapshot carries the key and a restore looks it
up, so renaming a slug orphans the tokens already written under the old one.
Rename one with the care a database column gets. A name that would be ambiguous
takes a qualifier from what distinguishes it — `dragon-firebending` for Avatar
Roku's, `dragon-nesting` for the one a Dragon Egg hatches into — because the slug
is an identity and not a label.

A token template is filed in `defs` and never in `registry`, so `All()` — which
the coverage census and the deck builders read — keeps counting cards and the
catalogue page does not grow eight new "cards". That was #521's decision and it
still holds.

---

## Cards

Three that could not be written at all:

| Card | What it needed |
|---|---|
| **Reef Worm** (#459) | three nested token dies-triggers — `token:fish` makes `token:whale` makes a plain 9/9 Kraken |
| **Nesting Dragon** (#397) | a token with a TRIGGER (`token:dragon-egg`) hatching a token with an ACTIVATED ability (`token:dragon-nesting`) |
| **Mysidian Elder** (#448) | a token trigger watching an event elsewhere — "whenever you cast a noncreature spell" |

Five caveats cleared:

| Card | Caveat removed |
|---|---|
| **Fable of the Mirror-Breaker** (#343) | "The Goblin Shaman token is a plain 2/2" → `caveats` to `full` |
| **Avatar Roku** (#343) | "The Dragon token doesn't have its own firebending 4" (the mana-duration caveat stays, and now names the token too) |
| **Beledros Witherbloom** | "The Pest tokens don't gain you 1 life when they die" → `caveats` to `full` |
| **Sedgemoor Witch** | the same Pest caveat (the magecraft-on-copies caveat stays) |
| **Cornered by Black Mages** | "The Wizard token is a plain 0/1" → `caveats` to `full` |

---

## Out of scope (explicit deferrals)

- **Avatar Kuruk's Spirit token** — "can't block or be blocked by non-Spirit
  creatures" is a `game.BlockRule.Pair`, and `CardDef` has no `BlockRules` slot:
  `CatalogBlockRules` exists with its card side explicitly deferred
  (`game/block_rules.go`, ADR 0045's addendum, PR 4). Its
  `OracleID == ""` skip is fixed here; the Spec field is that PR's. The card
  keeps its caveat, now pointing at the block-rule row rather than this one.
- **Double-faced tokens** (Incubator) — the same object model, tracked on the
  transform row of `docs/engine-seams.md`.
- **Token art** — ADR 0078 / #1115. `token_text` is deliberately not a
  substitute: when art lands, the text stays, because a token's printing is a
  prop and its text is rules.
- **A token template declaring `PrintedKeywords`** — the gate now lets one
  through, but every token's keywords stay plain data on the `Card`, which is
  where the deck importer and `printedCharacteristic` already read them.
- **Chocobo Racetrack (#394) and Gwaihir (#464)** — both are now one template
  away and neither needs anything from the engine; they belong to their own batch
  PRs rather than to this one.

---

## Consequences

- A token can carry any printed ability a card can, declared the same way.
- Nine catalog slots stop skipping tokens; nothing declares one on a token yet,
  so the next card to want one is a `Spec` field rather than an engine change.
- `game.Card` is unchanged. A board of ability-bearing tokens is still a restore
  point, which is #521's whole point extended rather than dented.
- One new wire field, `token_text`, additive and `omitempty`.
- The seam row moves to Closed. The audit's 18-cards-only / 24-cards-any figure
  is unblocked; five caveats clear today and three new cards ship.
- The cost: a token with an ability is now declared in a catalog file rather than
  as a row in `tokens_table.go`, so there are two places a token can live and the
  reader has to know which. The rule is the one #521 set — behaviour goes in the
  catalog, data goes in the table — and `TestEveryTokenKeyResolves` plus the
  registration rules keep each honest.

---

## Test plan

Engine (`server/internal/game/token_abilities_test.go`), catalog stubbed so each
property stands alone:

1. a token's dies-trigger fires and resolves;
2. it fires from the graveyard, and the CR 704.5d sweep runs after it;
3. a token's static applies to another creature and stops when the token leaves;
4. a copy of a token carries its trigger and its static, and a token copying a
   printed card carries neither (CR 707.2, both directions);
5. undo across a token trigger unqueues it, and the replay queues exactly one;
6. `TokenTextForCard` is a catalog read, silent for a printed card;
7. the AST scan for `OracleID == ""` gates, with its allowlist.

Cards (`server/internal/cards/effects/token_abilities_cards_test.go`), through the
real catalog: the Reef Worm chain end to end and its tokens ceasing to exist,
Nesting Dragon's Egg and the Dragon's firebreathing, Mysidian Elder's Wizard
pinging each opponent and not its controller, Fable's Goblin Shaman making a
Treasure on attack, Avatar Roku's Dragon adding four red, the Pest's life gain,
and the registration rules.

Wire (`server/internal/protocol/token_abilities_view_test.go`): `token_text` for a
trigger-only token and for one with an activated ability, and absent for a printed
card and for a vanilla token.

Client (`client/src/lib/tokenText.render.test.ts`): rendered, full text in the
`title`, printed line breaks kept, absent for a vanilla token and a printed card.
