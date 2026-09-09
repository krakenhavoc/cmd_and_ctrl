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
**68 nonland cards**. Of those, **14 are in the catalog** — four from
earlier sprints, nine in batch 1, and Aetherize in batch 2 — and **54
are blocked**. On the land side, Command Tower and Azorius Chancery are
in, and the 28 basics work.

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

## Done (batch 2)

| Card | What it exercises |
|---|---|
| Azorius Chancery | first bounce land — enters-tapped + a mandatory ETB bounce that can legally take itself |
| Aetherize | reads combat **state** (`Card.AttackingTarget`) rather than waiting on an event |

Building Chancery turned up an engine limitation worth recording: **a
catalog replacement effect cannot fire on its own source's entry.**
`gatherActiveReplacementsLocked` gathers catalog replacements by walking
`g.Battlefield.Cards`, and the entering card is not on the battlefield
yet when the pipeline runs. Confirmed by probe. So enters-tapped uses
the Worn Powerstone pattern (`OnETB` taps as it lands), and the
"self-replacement — `ev.CardID == src.InstanceID`" pattern in
AGENTS.md §7 is unreachable for catalog cards as the engine stands —
the Hangarback Walker it cites is not in the catalog, so nothing
exercised it.

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

**Airbend — 6 cards, but no longer one gap (updated after #232).**
Aang (all three), Appa Steadfast Guardian, Monk Gyatso, Avatar's Wrath.
"Exile it. While it's exiled, its owner may cast it for {2} rather than
its mana cost."

S21 sub-PR 6 (#232) landed the hard half: `Card.ExilePlay` is a durable
per-card permission naming a holder, `castSourceZoneLocked` accepts
`exile`, and the grant survives `Clone` for undo. Airbend can reuse all
of that — it grants to the card's *owner* rather than to the exiler,
which the `Player` field already allows.

**Three shared pieces are still missing:**

1. **An alternative cost.** "{2} rather than its mana cost" has no
   expression. #230's `AdditionalCost` is an *extra* cost, not a
   replacement for the mana cost. This is the expensive one — and it is
   *not* airbend-specific (see the cost group below), which is why the
   suggested order now puts it first.
2. **A permission with no expiry.** `ExilePlayPermission.UntilTurn`
   gates on `turn <= UntilTurn`; airbend is "while it's exiled", i.e.
   unbounded. Small.
3. **Exile-a-target-permanent-with-permission.** #232's primitive
   exiles the top N of a *library*. Airbend exiles a targeted permanent
   from the battlefield. Moderate, and it reuses the hard part above.

**Doing all three unlocks exactly one of these six cards** — Aang, the
Last Airbender, whose only other clause is "whenever you cast a Lesson
spell" (a subtype check). Every other airbend card carries an
independent blocker, so this group should not be counted as six:

- **Monk Gyatso** — "whenever another creature you control becomes the
  target of a spell or ability". There is no targeting event of any
  kind in `events.go`. Blocked regardless of airbend.
- **Appa, Steadfast Guardian** — "whenever you cast a spell *from
  exile*". `EventCast` is emitted with only Kind / Actor / Source /
  CardID; it carries no source zone, so this cannot be written even
  though `Event` has `OldZone` for other kinds. One field.
- **Aang, Swift Savior** — a transform DFC (multi-face, below), and it
  airbends a *spell* off the stack rather than a permanent, plus
  waterbend {8}.
- **Avatar's Wrath** — mass airbend, self-exile on resolution, and a
  continuous "opponents can't cast spells from anywhere other than
  their hands" restriction.
- **Aang, Airbending Master** — experience counters (no player-scoped
  counter store) and "leave the battlefield *without dying*".

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

**Alternative cast costs — the actual keystone (7 cards here, more
elsewhere).** Evoke (Slithermuse), cleave (Wash Away), spree (Three
Steps Ahead), foretell (Cosmic Intervention, Ranar), plot (Aven
Interrupter), warp (Anticausal Vestige). Every one of these is "pay
something *instead of* the mana cost", and none of it exists: #230's
`game.AdditionalCost` is an extra cost paid *alongside*. The same gap
already has three cards in the catalog shipping with their headline
mode missing — overload on Vandalblast and Cyclonic Rift, evoke on
Slithermuse — and it is piece 1 of the three airbend needs above. That
overlap is why it now leads the suggested order: one mechanism, this
group plus airbend plus three already-shipped cards.

**Tap permanents as a cost — convoke and waterbend (7 cards).** Clever
Concealment and The Wandering Rescuer (convoke); Aang Swift Savior,
Katara Water Tribe's Hope, The Legend of Kuruk, The Unagi of Kyoshi
Island, Waterbender's Restoration (waterbend {X}, which is convoke with
a different name and no colour clause). One cost component serves both.

**Attack triggers (4 cards).** Katara Waterbending Master, Phelia, Sun
Titan's attack half, and The Mighty Thor. There is no `EventAttack` in
`events.go` — the event plumbing is the PR, not the cards. Cheapest item
on this list by a wide margin.

Note the distinction batch 2 drew: combat **state** already exists
(`DeclareAttacker` stamps `Card.AttackingTarget`, `ClearCombat` wipes
it), so a *spell* that reads it at resolution needs nothing new — which
is why Aetherize shipped. It is only *triggers* that want the event.

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

28 basics, Command Tower and **Azorius Chancery** (batch 2) work today.

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

Re-sequenced after #232 (impulse exile) landed the cast-from-exile half
of what airbend needs. Cast-from-exile was #1 in the previous revision;
it is now partly done, and the analysis above showed the remaining
airbend work is mostly *not* airbend-specific.

1. **Alternative cast costs.** "Pay X instead of the mana cost." Serves the 7-card cost group, is piece 1 of the 3 airbend still needs, and retro-fixes overload on Vandalblast / Cyclonic Rift and evoke on Slithermuse — three cards already in the catalog shipping without their headline mode.
2. **Attack triggers** — one new `EventKind`, 4 cards here and more elsewhere. Still the cheapest item by a wide margin.
3. **Finish airbend** — the no-expiry permission and an exile-a-target-permanent primitive, on top of (1). Unlocks Aang, the Last Airbender outright; the other five each need something further, listed above.
4. **Flicker + delayed triggers** — 9 cards, and the delayed-trigger slot is reusable far beyond this deck.
5. **Tap-permanents-as-cost** — convoke and waterbend together, 7 cards.
6. **Multi-face cards** — 7 cards, but it touches import, protocol and client, so it is a sprint of its own.

Two one-field additions worth folding into whichever PR is nearby:
a source zone on `EventCast` (unblocks Appa's "cast a spell from
exile"), and a targeting event (Monk Gyatso, and nothing else in this
deck — but "becomes the target of" is a common Commander clause).
