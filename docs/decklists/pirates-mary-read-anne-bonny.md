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
sprint item, and the deck says how much each one buys.

**Alternative cast paths (S29)** — 7 cards: Faithless Looting
(flashback), Impulsive Pilferer (encore), Marauding Mako (cycling),
Ragavan (dash), Cyclonic Rift (overload), Deflecting Swat (free
cast), Improvisation Capstone. The cards work without them; only the
extra mode is missing.

**Additional costs on cast** — 4 cards: Thrill of Possibility, Big
Score, Unexpected Windfall ("as an additional cost, discard a card"),
Read the Runes (discard-or-sacrifice per card drawn).
`CastSpellParams` has no cost-payment slot, so these can't be cast
correctly at all. Cheapest high-value unlock in the list: one field
plus a prompt, and it turns three dead cards live.

**Discard as an ability cost** — 3 cards: Glint-Horn's activated
half, Solphim's indestructible ability, Bag of Holding. Wants the
same card-choice plumbing `SacrificeOther` already has.

**Vehicles / crew** — 3 cards: Smuggler's Copter, Magmatic Galleon,
RMS Titanic. No `Vehicle` handling and no crew action. Notable
because the commander's trigger explicitly names Vehicle cards, so
the deck is built around a type the engine can't represent.

**Sagas / chapter counters** — 2 cards: Fable of the Mirror-Breaker,
Brass's Tunnel-Grinder.

**Impulse exile ("exile the top card, you may play it this turn")** —
5 cards: Ragavan, Breeches (both), Malcolm, Coin of Mastery. Needs an
exile-with-permission zone and a cast-from-exile path. This is the
deck's whole card-advantage engine, so a Pirates deck plays badly
without it even though every individual card "works".

**Control change** — 1 card: Coercive Recruiter ("gain control until
end of turn"). Needs a controller-change effect with end-of-turn
cleanup.

**Type/ability overwrite** — 1 card: Kitesail Larcenist (permanents
*become* Treasures and lose all abilities). Layer 4 + 6 + a new
ability-granting shape.

**Batched triggers (CR 603.1)** — cosmetic here. "Whenever one or
more X" fires once per X instead of once per batch. Same totals for
every card in this deck; it would matter for a card that reads the
batch size nonlinearly.

**Not in the local Scryfall snapshot** — Fable of the Mirror-Breaker,
Ojer Axonil, Storm the Vault. Double-faced or renamed; the importer
matches on exact face name and these need the `card_faces` path.

## Suggested order

1. Additional costs on cast — 3 cards, small change.
2. Impulse exile — 5 cards, and it's the archetype's engine.
3. Discard as an ability cost — 3 cards, reuses the sacrifice picker.
4. Vehicles — 3 cards, and the commander references the type.
