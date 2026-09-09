# Decklist support — Aang is so Flashy

Source: [Archidekt](https://archidekt.com/decks/21873490/aang_is_so_flashy) —
Azorius flash + blink, built around *airbend* and end-step flicker.

This is a **triage of one real deck against the catalog**, in the same
shape as [the Pirates list](pirates-mary-read-anne-bonny.md): a deck is
a better forcing function than a card count, because it says which
gaps actually stop a game from being played.

Counts below were recomputed from the deck's actual 77 entries resolved
against the Scryfall dump, not from the deck page. **9 entries are
lands** (including Island ×14 and Plains ×14 as two entries), leaving
**68 nonland cards**. Of those, **13 are in the catalog** — four from
earlier sprints plus the nine in batch 1 — and **55 are blocked**.

## Card index

The dump was refreshed **2026-09-09**; all 77 entries now resolve and the
deck imports.

Before that refresh two cards were genuinely absent, both from Marvel
Super Heroes (2026-06-26): **The Mighty Thor, Jane Foster** and **The
Mind Stone**. Note the second is *not* the ordinary Mind Stone already in
the catalog — it is a Legendary Artifact — Infinity Stone, a different
card with the same-ish name.

Worth knowing for next time: `data/scryfall/last-refresh` is not a
reliable age for the data. It read `2026-04-17` while the dump itself
contained sets released through `2026-07-17`. Check a known-recent card,
not the stamp.

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
Aang (all three), Appa Steadfast Guardian, Monk Gyatso, Avatar's Wrath.
"Exile it. While it's exiled, its owner may cast it for {2} rather than
its mana cost." Needs exile that carries a durable per-card permission
plus an alternative cost, and casting from the exile zone. **This is the
same machinery as the Pirates list's impulse exile** — together the
single highest-value unlock across both decks. Note the deck is *named*
for it. Appa also pays it off directly: "whenever you cast a spell from
exile, create a 1/1 Ally".

**Flicker — exile and return, immediately or at the next end step (9
cards).** Thassa Deep-Dwelling, Y'shtola Rhul, Waterbender's
Restoration, Phelia, Cosmic Intervention, Skycoach Conductor // All
Aboard, Sword of Hearth and Home, and both newly-resolvable Marvel
cards — The Mighty Thor and The Mind Stone. `ExileTarget` exists but
nothing returns a card to the battlefield, and most of these return "at
the beginning of the next end step" — a delayed trigger, which has no
shape in `game.TriggeredAbility` at all. Two pieces: a return-from-exile
primitive and a delayed-trigger slot. (Sun Titan and Enduring Curiosity
are adjacent but different — graveyard recursion and a returns-as-an-
enchantment clause respectively.)

**Multi-face cards (7 cards).** Sea Gate Restoration // Sea Gate,
Reborn and Sink into Stupor // Soporific Springs (modal DFCs — spells
with land backs, so they do not count toward the 9 land entries above),
Aang Swift Savior // Aang and La (transform), The Legend of Kuruk //
Avatar Kuruk, Virtue of Knowledge // Vantress Visions (adventure),
Skycoach Conductor // All Aboard (prepare), Fortune Teller's Talent
(class). `game.Card` has one name, one type line, one mana cost. This
also affects **deck import and validation**, not just resolution, so it
is bigger than its card count suggests.

**Alternative and additional cast costs (7 cards).** Evoke
(Slithermuse), cleave (Wash Away), spree (Three Steps Ahead), foretell
(Cosmic Intervention, Ranar), plot (Aven Interrupter), warp
(Anticausal Vestige). Same family as the Pirates list's group — the
base cards mostly work, only the extra mode is missing. Note S21
sub-PR 5 has since landed `game.AdditionalCost`, which is the hook the
waterbend group below would extend.

**Tap permanents as a cost — convoke and waterbend (7 cards).** Clever
Concealment and The Wandering Rescuer (convoke); Aang Swift Savior,
Katara Water Tribe's Hope, The Legend of Kuruk, The Unagi of Kyoshi
Island, Waterbender's Restoration (waterbend {X}, which is convoke with
a different name and no colour clause). One cost component serves both.

**Attack triggers (4 cards).** Katara Waterbending Master, Phelia, Sun
Titan's attack half, and The Mighty Thor. There is no `EventAttack` in
`events.go` — the event plumbing is the PR, not the cards. Cheapest item
on this list by a wide margin.

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

**Keywords not in the canonical list (5 cards).** Hexproof (Lotus
Field, Stoic Sphinx, The Wandering Rescuer) and indestructible (Thassa,
The Mind Stone, The Seriema) are both absent from the twelve tokens
`PrintedKeywords` accepts. Cheap to add for the static half; the
combat/SBA behaviour behind them is the real work.

**One-offs.** Herald of Eternal Dawn (can't-lose replacement), Mandate
of Peace (end the combat phase), Clever Concealment (phasing), Rabble
Rousing (hideaway), The Seriema (station), Misleading Signpost
(re-select an attacker's target), Hullbreaker Horror (can't be
countered), Thassa (devotion), Enduring Curiosity (returns as a
non-creature enchantment), Ty Lee (permanent doesn't-untap lockdown),
Meticulous Archive (surveil), The Mind Stone ("harness", a new
once-activated state gate).

## Lands

28 basics and Command Tower work today.

**Azorius Chancery is a clean batch-2 candidate**: enters-tapped
replacement, an ETB that returns a land you control to hand, and a
`{W}{U}` mana ability — all existing machinery.

**Demolition Field is a partial**: the three-part activated cost
(`{2}`, `{T}`, sacrifice) has a precedent in Mind Stone, but the
ability makes the *opponent* optionally search their library, and there
is no shape for prompting another player mid-resolution. Implementable
only if that half is declared as a simplification.

**Lotus Field is not a candidate** (corrected — an earlier draft listed
it as one). Two independent blockers: `hexproof` is not among the
twelve canonical `PrintedKeywords` tokens, and "{T}: Add three mana of
any one color" cannot be expressed in the pipe syntax — three
`{W|U|B|R|G}` slots would let the controller pick a *different* colour
per slot, which is a strictly stronger card.

Floodfarm Verge needs a conditional mana ability; Hallowed Fountain and
Meticulous Archive need the "pay life as it enters, or it enters
tapped" / surveil ETB clauses.

## Suggested order

1. **Cast-from-exile** — airbend here plus the Pirates list's impulse exile. The largest unlock across both decks, and this deck is named for it.
2. **Attack triggers** — one event kind, 4 cards here and more elsewhere.
3. **Flicker + delayed triggers** — 9 cards, and the delayed-trigger slot is reusable far beyond this deck.
4. **Tap-permanents-as-cost** — convoke and waterbend together, 7 cards, and `game.AdditionalCost` from S21 sub-PR 5 is the hook.
5. **Multi-face cards** — 7 cards, but it touches import, protocol and client, so it is a sprint of its own.
