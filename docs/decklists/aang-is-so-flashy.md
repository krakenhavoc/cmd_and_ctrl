# Decklist support — Aang is so Flashy

Source: [Archidekt](https://archidekt.com/decks/21873490/aang_is_so_flashy) —
Azorius flash + blink, built around *airbend* and end-step flicker.

This is a **triage of one real deck against the catalog**, in the same
shape as [the Pirates list](pirates-mary-read-anne-bonny.md): a deck is
a better forcing function than a card count, because it says which
gaps actually stop a game from being played.

Headline: of 75 nonland cards, **6 were already in the catalog and 9
land here**. The remaining 60 are blocked, and they are blocked on
unusually few things — three gaps account for 20 of them.

## Blocking finding: the card index predates the deck

`data/scryfall/default-cards.json` was last refreshed **2026-04-17**.
*The Mighty Thor, Jane Foster* is not in it, so this deck cannot be
imported at all until `scripts/scryfall-refresh.sh` runs — the upload
fails validation on an unknown card rather than degrading. Every other
card in the list resolves, including the whole Avatar set.

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

Already in the catalog from earlier sprints: Sol Ring, Arcane Signet,
Mind Stone, Command Tower, Solemn Simulacrum, The Wandering Emperor.

### Sandbox simplifications declared in this batch

- **Fellwar Stone** offers the full five-colour pipe rather than the
  intersection of what opponents' lands could produce. Narrowing needs
  a per-activation scan; the engine has that hook only for commander
  identity (Arcane Signet).
- **Peregrine Drake** untaps up to five *tapped lands its controller
  controls*, in battlefield order. "Up to five lands" is a free pick in
  paper; there is no trigger-side multi-pick UI. Strictly conservative
  — it can never untap an opponent's land.
- **Deputy of Acquittals** cannot express "**another** target creature":
  `TargetSpec` is declared statically at `init()` and `NotSelf` needs an
  `InstanceID` that does not exist yet. The picker offers every creature
  you control; the effect declines to bounce itself, so the printed
  restriction still holds at resolution.
- **Slithermuse** auto-picks the opponent holding the most cards (always
  the maximising choice) and has no evoke.

### One engine finding from this batch

`containsFoldASCII` folds **only the haystack**, so a capitalised needle
silently matches nothing. `IsBasicLandExcept("")` is the sharper trap:
an empty needle makes the check true for every card, so the negation
matches nothing at all — a fetch predicate written that way fails
silently rather than loudly. Every existing caller happens to pass a
lowercase literal, so nothing is broken today; the contract is now
documented on `IsBasicLandWithSubtype`, which exists because "a basic
Plains card" is expressible by neither `IsBasicLand` (any basic) nor
`IsLandWithSubtype` (admits nonbasics with the type).

## Blocked, by machinery needed

Grouped by what would unblock them, ordered by how many cards each buys.

**Airbend — exile with a cast-from-exile permission (6 cards).**
Aang (all three), Appa, Monk Gyatso, Avatar's Wrath. "Exile it. While
it's exiled, its owner may cast it for {2} rather than its mana cost."
Needs exile that carries a durable per-card permission plus an
alternative cost, and casting from the exile zone. **This is the same
machinery as the Pirates list's impulse exile (5 cards)** — together
11 cards across two real decks, which makes it the highest-value
unlock in the repo. Note the deck is *named* for it.

**Flicker — exile and return, immediately or at the next end step (7
cards).** Thassa Deep-Dwelling, Y'shtola Rhul, Waterbender's
Restoration, Phelia, Cosmic Intervention, All Aboard, Sword of Hearth
and Home. `ExileTarget` exists but nothing returns a card to the
battlefield, and most of these return "at the beginning of the next
end step" — a delayed trigger, which has no shape in
`game.TriggeredAbility` at all. Two pieces: a return-from-exile
primitive and a delayed-trigger slot.

**Multi-face cards (7 cards).** Sea Gate Restoration, Sink into Stupor
(modal DFC lands — both count toward the deck's land base), Aang Swift
Savior, The Legend of Kuruk (transform), Virtue of Knowledge
(adventure), Skycoach Conductor (prepare), Fortune Teller's Talent
(class). `game.Card` has one name, one type line, one mana cost. This
also affects **deck import and validation**, not just resolution, so it
is bigger than its card count suggests.

**Alternative and additional cast costs (7 cards).** Evoke
(Slithermuse), cleave (Wash Away), spree (Three Steps Ahead), foretell
(Cosmic Intervention, Ranar), plot (Aven Interrupter), warp
(Anticausal Vestige). Same family as the Pirates list's S29 group —
the base cards mostly work, only the extra mode is missing.

**Tap permanents as a cost — convoke and waterbend (6 cards).** Clever
Concealment and The Wandering Rescuer (convoke); Katara Water Tribe's
Hope, The Unagi, Avatar Kuruk, Waterbender's Restoration (waterbend
{X}, which is convoke with a different name and no colour clause).
One cost component serves both.

**Attack triggers (4 cards).** Sun Titan's attack half, Phelia, Katara
Waterbending Master, Rabble Rousing. There is no `EventAttack` in
`events.go` — the event plumbing is the PR, not the cards. Cheapest
item on this list by a wide margin.

**Copy effects (3 cards).** Clone, Three Steps Ahead, Skycoach
Conductor.

**Cost reduction (4 cards).** Pearl Medallion, Oketra's Monument, The
Wind Crystal, Fortune Teller's Talent level 3. Already scheduled for
S28 (cost engine).

**ETB-trigger doubling (2 cards).** Panharmonicon, Virtue of Knowledge.
Needs a trigger-count replacement, distinct from the CR 614 pipeline,
which replaces events rather than triggers.

**Planeswalker loyalty abilities (2 cards).** Both Teferis.
`AbilityCost` has tap / sacrifice / mana / life but no loyalty
component, and loyalty is sorcery-speed and once per turn.

**Experience counters (2 cards).** Aang Airbending Master, Katara
Waterbending Master. Counters live on cards; there is no player-scoped
counter store.

**One-offs.** Herald of Eternal Dawn (can't-lose replacement), Mandate
of Peace (end the combat phase), Clever Concealment (phasing), Rabble
Rousing (hideaway), The Seriema (station), Misleading Signpost
(re-select an attacker's target), Hullbreaker Horror (can't be
countered), Thassa (devotion), Enduring Curiosity (returns as a
non-creature enchantment), Ty Lee (permanent doesn't-untap lockdown),
Meticulous Archive (surveil).

## Lands

28 basics and Command Tower work today. **Azorius Chancery, Demolition
Field and Lotus Field are batch-2 candidates** — all three are
expressible now (enters-tapped replacement, an activated ability with a
three-part cost, ETB sacrifice). Floodfarm Verge needs a conditional
mana ability; Hallowed Fountain, Sea Gate Reborn and Soporific Springs
need the "pay life as it enters, or it enters tapped" ETB choice.

## Suggested order

1. **Refresh the Scryfall dump.** Nothing else matters until the deck imports.
2. **Cast-from-exile** — airbend + impulse exile, 11 cards across two decks.
3. **Attack triggers** — one event kind, 4 cards here and more elsewhere.
4. **Flicker + delayed triggers** — 7 cards, and the delayed-trigger slot is reusable far beyond this deck.
5. **Tap-permanents-as-cost** — convoke and waterbend together, 6 cards.
6. **Multi-face cards** — 7 cards, but it touches import, protocol and client, so it is a sprint of its own.
