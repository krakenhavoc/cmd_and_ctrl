# Decklist support — Women of the Sea (Mary Read and Anne Bonny)

Source: [Moxfield](https://moxfield.com/decks/Yjlzz2_Zs0CruZU_vw8m-g) —
Izzet Pirate tribal, looting into Treasure into artifact payoffs.

This is a **triage of one real deck against the catalog**, used to
decide what the engine builds next. A deck is a better forcing
function than a card count: it says which gaps actually stop a game
from being played, rather than which cards happen to be easy.

## Done (batch 1)

| Card | What it exercises |
|---|---|
| Reckless Fireweaver | artifact-ETB trigger, damage to each opponent |
| Ingenious Artillerist | same, at a different rate |
| Marauding Mako | discard trigger → +1/+1 counters |
| Glint-Horn Buccaneer | discard trigger → table ping (activated half deferred) |
| Corsair Captain | ETB Treasure **and** a Layer 7c lord in one card |
| Impulsive Pilferer | dies → Treasure |
| Angrath's Marauders | first damage-**amount** replacement in the catalog |
| Faithless Looting | draw-then-discard in printed order |
| Mary Read and Anne Bonny | the commander: tap-to-loot + typed-discard → tapped Treasure |

Already in the catalog from earlier sprints: Sol Ring, Counterspell.

## Done (batch 2 — additional costs on cast)

| Card | What it exercises |
|---|---|
| Thrill of Possibility | the discard is a **cast cost**, so the payoffs trigger above the spell |
| Big Score | same cost, plus two Treasures for the artifact payoffs above |
| Unexpected Windfall | the same card at a different price; the deck runs both |

Read the Runes stays blocked: its cost is per-card-drawn and offers a
choice between discarding and sacrificing, which is a different shape
from a fixed "discard a card".

## Done (batch 3 — impulse exile)

| Card | What it exercises |
|---|---|
| Ragavan, Nimble Pilferer | the first card played from a zone that isn't yours; Treasure + cast-only steal |
| Breeches, Brazen Plunderer | the same, with "play" instead of "cast" and the any-colour mana relaxation |

**The triage overcounted this group.** It listed five cards; only
three are impulse exile, and one of those needs machinery this batch
didn't build:

- **Malcolm, Alluring Scoundrel** is not impulse exile. It loots on
  combat damage and, at four chorus counters, lets you cast the
  *discarded* card for free — cast-from-graveyard with an alternative
  cost, which is a different mechanic (and closer to S29's
  alternative cast paths).
- **Coin of Mastery** is not impulse exile either — it's an
  enters-with-counters replacement plus a Treasure ability. Both
  halves are already expressible; it was mis-filed.
- **Breeches, Eager Pillager** is impulse exile in one of three
  modes, but its trigger is *modal with a once-per-turn-per-mode
  ledger* ("choose one that hasn't been chosen this turn"). That's
  new machinery on `TriggeredAbility`, not on exile. Still blocked.

## Done (batch 4 — everything that needed no engine work)

Fourteen cards written against machinery that already existed, once
batches 2 and 3 had landed. No engine, wire or client changes.

| Card | What it exercises |
|---|---|
| Captain Storm, Cosmium Raider | artifact-ETB watcher **plus** a target clause — first pairing |
| Imperial Recruiter | `SearchLibrary` with a power ceiling |
| Magmakin Artillerist | discard → damage to each opponent |
| Malcolm, Keen-Eyed Navigator | Pirate combat damage → a Treasure per damaged opponent |
| Scrounging Skyray | Marauding Mako with evasion |
| Weftstalker Ardent | Reckless Fireweaver widened to creatures, with "another" |
| Gemcutter Buccaneer | Pirate-ETB → tapped Treasure (Equipment half declared) |
| Solphim, Mayhem Dominus | the damage doubler, narrowed to noncombat and to opponents |
| Gamble | tutor + random discard |
| Windfall | every hand pitched before anyone draws |
| An Offer You Can't Refuse | the Treasures go to the **countered** spell's controller |
| Decaying Time Loop | count taken before the draw |
| Pull from Tomorrow | X draw, then a queued discard |
| Frantic Search | loot 2, untap up to three lands |

That is **31 of the 65 nonland cards** in the catalog.

### A correction to batch 1

**Quicksmith Genius was never actually written.** It appears in this
deck and was listed as shipped; it is not in the registry. It is also
genuinely blocked, which is presumably why: "whenever an artifact you
control enters, you may discard a card. If you do, draw a card"
needs the draw to happen *after* the player has chosen what to
discard. `DiscardChoiceForEffect` queues the choice and returns, so
the draw would land first — turning a rummage into a loot, which is a
strictly better card. Wants a continuation on the discard prompt, the
same shape `PayUnless.OnDecline` already has.

### Two engine findings from this batch

1. **Damage to a player from a spell or ability skipped the CR 614
   replacement pipeline entirely.** `DealDamageToPlayerForEffect`
   emitted its event and changed life directly. Combat damage and
   damage marked on creatures were both routed; this was the hole.
   No damage doubler or prevention shield could ever see a Lightning
   Bolt. Fixed here, because Angrath's Marauders is unimplementable
   without it — and it means Fog-style prevention now has a shot at
   direct damage too.
2. **A layer predicate must not call `Effective()` on its target.**
   The target's characteristics are mid-rebuild when `AppliesTo`
   runs, so Corsair Captain's Pirate check reads the printed type
   line. Cost: a creature *turned into* a Pirate isn't counted by the
   lord. That wants the layer engine to expose a partially-applied
   view, which it doesn't have.

## Blocked, by machinery needed

Grouped by what would unblock them — each group is a candidate
sprint item, and the deck says how much each one buys. Re-derived
against the full 65-card list after batch 4, which turned up several
groups the first pass missed.

**Alternative cast paths (S29)** — 7 cards: Faithless Looting
(flashback), Impulsive Pilferer (encore), Marauding Mako and
Scrounging Skyray (cycling), Ragavan (dash), Weftstalker Ardent
(warp), Decaying Time Loop (retrace), Cyclonic Rift (overload),
Deflecting Swat (free cast), Improvisation Capstone. The cards all
work without them; only the extra mode is missing.

**Additional costs on cast** — 1 card left. Three shipped in batch 2
([ADR 0021](../decisions/0021-additional-costs.md)). **Read the
Runes** remains: its cost is X-many payments, each a choice between
discarding a card and sacrificing a permanent, which wants a
per-payment prompt rather than a fixed count.

**Impulse exile** — shipped in batch 3
([ADR 0022](../decisions/0022-impulse-exile.md)) for Ragavan and
Breeches, Brazen Plunderer. Breeches, Eager Pillager remains, blocked
on modal triggers below.

**Modal triggered abilities with a once-per-turn ledger** — 2 cards:
Breeches, Eager Pillager and Monument to Endurance. Both read "choose
one that hasn't been chosen this turn". `ModeSpec` exists for spells;
this needs it on `TriggeredAbility` plus a per-turn record of which
options a source has already used. Two cards makes it worth doing,
and it's the last impulse-exile holdout.

**Discard as an ability cost** — 3 cards: Glint-Horn's activated
half, Solphim's indestructible ability, Bag of Holding. Wants the
same card-choice plumbing `SacrificeOther` already has, and
`AdditionalCost.DiscardCards` is the shape to copy onto
`AbilityCost`. Also closes the Blood token's declared gap from S21
sub-PR 4.

**Vehicles / crew** — 4 cards: Smuggler's Copter, Magmatic Galleon,
RMS Titanic, Jackdaw. No `Vehicle` type handling and no crew action.
Notable because the commander's trigger explicitly names Vehicle
cards, so the deck is built around a type the engine can't represent.
The Indomitable is a fifth, and also wants cast-from-graveyard.

**Until-end-of-turn pumps (S25)** — 2 cards: Captain Lannery Storm's
sacrifice payoff and Captain Howler's +2/+0. Both otherwise trivial.

**Attack triggers** — 2 cards: Captain Lannery Storm, Breeches Eager
Pillager. There is no attack event at all — `DeclareAttacker` mutates
state without emitting one — so "whenever ~ attacks" is unreachable.
Cheap to add and it unblocks a very common template.

**End-step triggers** — 1 card: Unstoppable Plan. `EventStepTransition`
exists but is an engine-internal sentinel for the replacement
pipeline, not a trigger source; upkeep has its own `EventBeginUpkeep`
and end step has no equivalent.

**Token-creation replacements** — 1 card: Academy Manufactor ("if you
would create a Clue, Food, or Treasure token, instead create one of
each"). The CR 614 pipeline has no token-creation event kind. Worth
noting that Doubling Season already wants the same hook.

**Per-turn discard tally** — 1 card: Change of Fortune ("draw a card
for each card you've discarded this turn"). Same shape as
`Game.SpellsCastThisTurn`, which already exists for cast counting.

**Sagas / chapter counters** — 2 cards: Fable of the Mirror-Breaker,
Brass's Tunnel-Grinder.

**Class enchantments with levels** — 1 card: Cool but Rude.

**Coin flips and spell copying** — 2 cards: Breeches, the Blastmaker
and Echocasting Symposium (which also wants Paradigm).

**Discover** — 1 card: Hit the Mother Lode.

**Per-mode target slots** — 2 cards: Mishra's Command and Prismari
Command. Both are "choose two" with more than one targeted option,
which `Register` panics on today (S20 sub-PR 4's declared limit).

**Restricted counterspells** — 1 card: Siren Stormtamer, which
counters only a spell "that targets you or a creature you control".
`TargetSpell` predicates see the candidate card, not what that card
is targeting, so the restriction isn't expressible. Implementing it
without the restriction would be a strictly stronger card.

**Control change** — 1 card: Coercive Recruiter. Needs a
controller-change effect with end-of-turn cleanup.

**Type/ability overwrite** — 1 card: Kitesail Larcenist (permanents
*become* Treasures and lose all abilities). The ability-removal shape
this called for now exists — S24 shipped `LoseAllAbilities` and made
layer 6 authoritative over the `Catalog*` hooks, for Darksteel
Mutation and friends. What is still missing here is the rest of the
card: a per-opponent "choose up to one" prompt, and a static whose
duration is "for as long as you control this creature" rather than
the lifetime of a permanent or the end of a turn.

**Graveyard provenance** — 1 card: Ghost of Ramirez DePietro, which
cares whether a card "was discarded or put there from a library this
turn". Nothing records how a card reached a graveyard.

**Mana-source provenance** — 1 card: Coin of Mastery, whose counters
scale with "mana from an artifact source spent to cast it".
`ManaToken.Source` exists, so this is closer than it looks, but
nothing tracks which tokens paid which spell.

**Becomes-the-target triggers + evasion restrictions** — 1 card:
Departed Deckhand.

**Batched triggers (CR 603.1)** — cosmetic for most of this deck.
"Whenever one or more X" fires once per X instead of once per batch.
Same totals everywhere here **except** Malcolm and Breeches, where
two Pirates hitting the *same* opponent produces two payouts instead
of one.

**Not in the local Scryfall snapshot** — Ojer Axonil and Storm the
Vault, both double-faced; the importer matches on exact face name and
these need the `card_faces` path. (Fable of the Mirror-Breaker is
also DFC.) Note that the dump's `last-refresh` stamp understates the
data's real coverage — check the sets, not the stamp.

## Suggested order

1. ~~Additional costs on cast~~ — **done** (batch 2).
2. ~~Impulse exile~~ — **done** (batch 3).
3. **Attack triggers** — an event the engine simply doesn't emit.
   Cheapest item on the list, unblocks 2 cards here and a very
   common template everywhere else.
4. **Modal triggered abilities with a once-per-turn ledger** —
   2 cards, and it's the last impulse-exile holdout.
5. **Discard as an ability cost** — 3 cards, reuses the sacrifice
   picker, and closes the Blood token's declared gap.
6. **Vehicles / crew** — 4–5 cards, and the commander references the
   type.
7. **Until-end-of-turn pumps** — 2 cards, and S25 wants it anyway.
