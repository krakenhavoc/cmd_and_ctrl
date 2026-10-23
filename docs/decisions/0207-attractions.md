# ADR 0207 — Attractions

**Status:** Accepted · 2026-10-23 · S76 — Cards: incubate, face-down and outside the game.
**Issue:** [#2121](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2121). Related: [#2138](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2138) (stickers and ticket counters, the other Unfinity seam, built next), [#2078](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2078) (the ADR 0116 pool that found Down for Repairs).
**Numbering:** checked on 2026-10-23. `develop` ends at 0206, and `git log --remotes --name-only -- docs/decisions` over every one of the 87 remote heads shows no file above 0206. No open issue's claim comment reserves 0207; the claim comment on #2121 reserves it.
**Builds on:** [ADR 0064](0064-emblems.md) (a second container in the command zone that no `ZoneRef` reaches), [ADR 0172](0172-companion-a-card-outside-the-game.md) and [ADR 0191](0191-a-sideboard-is-the-outside-the-game-pool.md) (a deck section that seats somewhere other than the library, and a section that breaks a rule being said and not seated), [ADR 0192](0192-a-deck-names-its-printing.md) §2 (the printing a deck names), [ADR 0116](0116-a-card-filter-on-the-revealed-hand-pick.md) (Down for Repairs' first half), [ADR 0041](0041-game-persistence.md) (restore points, the additive snapshot rule and the closure ratchet), [ADR 0054](0054-dice-rolls-and-coin-flips.md) (dice and keyed random streams).

---

## Context

### Should C2 support them at all?

Yes. Attractions are an Unfinity mechanic, and Unfinity is split. Its acorn-stamped cards are not legal in any eternal format; its other cards are. Scryfall's `legalities.commander` on the 2026-10-05 dump: 46 Attraction-related cards are `legal` (every Attraction without the acorn, and every card that opens, visits or destroys one, Down for Repairs among them) and 31 are `not_legal` (the acorn Attractions and openers). No ADR excludes Un-set cards; ADR 0078 and ADR 0152 only skip the silver-bordered and acorn printings where a card is not legal. So C2 supports the legal ones and refuses the rest the way it refuses any card that is not legal in Commander.

### The rules

Checked against `MagicCompRules 20260925.txt`.

- **717.1.** Attraction is an artifact subtype seen only on nontraditional cards with an Astrotorium back and a column of numbers, some "lit up". Cards with one English name may light different numbers.
- **717.2.** Attraction cards do not begin the game in a deck and do not count toward deck size; a player who plays with them begins with a supplementary Attraction deck in the command zone, shuffled before the game begins (103.3a). **717.2a:** in constructed play at least ten, each with a different English name.
- **717.4 / 703.4g.** As a player's precombat main phase begins, immediately after the Saga lore counters, a player who controls an Attraction rolls to visit their Attractions. A turn-based action; it does not use the stack.
- **701.51.** Open an Attraction: only while playing with an Attraction deck (a); move its top card off it, face up, onto the battlefield under your control (b); "whenever you open" triggers only if it entered (c).
- **701.52a.** Roll to visit: roll a d6; every Attraction you control with that number lit up is visited and its visit ability triggers.
- **702.159a.** "Visit — [Effect]" means "Whenever you roll to visit your Attractions, if the result is equal to a number that is lit up on this Attraction, [effect]."
- **717.6.** A card with an Astrotorium back that would be put into a zone other than the battlefield, exile or the command zone from anywhere goes to the command zone instead; may apply more than once to an event; an exception to 614.5. **717.6a:** those cards form one face-up pile per owner, the "junkyard", which is not its own zone.
- **727.2.** Restarting the game involves nontraditional cards in the new game too.

### What existed (develop at 1b1fea5ef)

Nothing. There was a registry row (`attractions`, missing), and Down for Repairs was waiting on it. Dice already existed (`RollDiceForEffect`, `EventRollDie`), as did the Saga turn-based action at the start of the precombat main phase, the entry batch every "put onto the battlefield" door uses, and a dependent target clause that may name a player as its anchor (`ControlledByTargetsController`, Ravager of the Fells).

## Options considered

**Where the Attraction deck and the junkyard live.**

- **A. In `Player.Command`.** Rejected for ADR 0064's reason: every reader of `Command.Cards` means "commander card" (the cast enumerator, the tax, the admin move verb, `ReplaceDeck`).
- **B. A new `ZoneKind`.** Rejected: rules-wise both piles are the command zone, and events and logs should say `command`.
- **C. Two containers of kind `ZoneCommand` beside `Command` and `Emblems`, outside the card index (chosen).** Nothing that resolves a zone by reference or looks a card up by ID reaches them, so no effect, move verb or admin tool can take a card out except the open, which is enforced by the container rather than by a check somebody has to remember.

**How a deck declares its Attractions.**

- **A. A deck-row flag set by a section header**, the way the sideboard is. Rejected as the only route: a pasted list with no heading would put Attractions in the library, where CR 717.2 says they never are.
- **B. By type line (chosen).** `Resolve` puts every card whose printed type line has the Attraction subtype into `List.Attractions`, wherever the source listed it. The text importer recognises `Attractions:` / `Attraction deck` headings, Moxfield's `attractions` board is read, and `List.Text()` writes the section back, but none of them is needed for the card to reach the right pile.

## Decision

### 1. The containers

`Player.AttractionDeck` (face down, ordered, top last) and `Player.Junkyard` (face up, arrival order) are `*Zone` of kind `ZoneCommand`, nil until a card arrives. Readers ask `AttractionDeckCards()` / `JunkyardCards()`. Neither is in the card index or behind a `ZoneRef`. `leave_game` sweeps both with the departing player's cards.

### 2. Seating and the deck list

- `deck.PrintsAttraction` (the printed subtype) routes a card to `List.Attractions` in `Resolve`. It is not one of the 100: count, singleton and colour identity never see it. Each card is still held to the format's legality, so an acorn Attraction is refused (`not_legal_in_format`).
- **CR 717.2a.** An Attraction deck with fewer than ten cards, or two with one name, is warned (`attraction_deck_illegal`, non-fatal, listed beside the sideboard warnings in `lobby`) and `List.SeatedAttractions()` seats none of it: ADR 0191 §5's posture, so the server never chooses which Attractions to keep, and a deck is never refused over its supplementary deck. Commander has no rule of its own for the Attraction deck, so the constructed rule is the one applied.
- `toGameCard` stamps `Card.AstrotoriumBack` (the physical back) on every Attraction card, and `seatDeckCard` puts an Astrotorium-backed card into the Attraction deck, known to nobody. `dealAndBeginLocked` shuffles it on a stream of its own (`attraction_shuffle`), so a deck with Attractions shuffles its library exactly as it would without them.
- A restart (CR 727.2) puts every Astrotorium-backed card its owner still has (battlefield, junkyard, deck, anywhere) back into the Attraction deck, which the new game's deal shuffles.

### 3. The lights are printing data the rules read

`cards.Card.AttractionLights` is Scryfall's `attraction_lights`, and `toGameCard` copies it to `game.Card.AttractionLights`. ADR 0192 §2 says a deck's chosen printing "picks a picture and nothing else"; for an Attraction it also picks the lit numbers, because the printings differ (Balloon Stand's four printings light {2,6}, {3,6}, {4,6} and {5,6}). This is that ADR's one exception. A deck that names no printing gets the index's default printing and its lights, exactly as it gets that printing's picture.

### 4. Open an Attraction (CR 701.51)

`Game.OpenAttractionsThenForEffect(player, source, n, then)` opens n times, one at a time. Each open puts the top card through the ordinary entry batch (`ZoneEntryOptions{Controller: player}`), so every CR 614 entry window, "enters" trigger and layer stamp applies, then announces `EventAttractionOpened` (CR 701.51c) only if the card entered. The card is outside the index, so for the span of its entry the index sees that one card (`Game.attractionEntering`, the shape of #3955's `outsideEntering`), closed by the entry's continuation whether it landed, was redirected or was refused. An empty or missing deck opens nothing and is not an error (701.51a). `then` is told the cards opened.

### 5. Roll to visit (CR 701.52, 702.159, 717.4)

`Game.RollToVisitForEffect(player, source)` rolls one d6 through `RollDiceForEffect` (so it is a die roll: `EventRollDie`, "whenever you roll a 1"), then announces `EventRollToVisit` (Amount the result), then `EventVisitAttraction` for each Attraction the player controls with that number lit, collected before anything is announced. The turn-based action is `rollToVisitForActiveSeatLocked`, called right after `advanceSagasForActiveSeatLocked` as the precombat main phase begins, for an active player who controls one or more Attractions.

A visit ability is an ordinary catalog trigger on `EventVisitAttraction` naming its own object (`effects.Visit`). The intervening "if" is the event itself: it is only emitted for a lit Attraction, and the lights are printed, so the condition cannot change before the ability resolves.

### 6. The junkyard (CR 717.6)

`astrotoriumJunkyardReplacement` is a mandatory built-in replacement on exits (`EventZoneMove`, `EventDiscardCard`): an Astrotorium-backed card (`HasAstrotoriumBack`, never a token) headed anywhere but the battlefield, exile or the command zone is sent to the command zone, owned by its owner. The two landings that map `ZoneCommand` to a pile (`routeDestinationLocked`, `executeBattlefieldLeaveLocked`) put such a card in its owner's junkyard (`commandPileFor`), never among the commanders. Command-zone cards are public, so a junkyard card is face up to the table. "May apply more than once" needs nothing extra: a built-in has no once-per-event identity.

Where another replacement also applies (Rest in Peace's exile), the ordinary CR 616 ordering applies with the owner choosing, which is the rules-correct outcome.

### 7. Persistence, undo and the closure ratchet

- **Snapshot (additive within v9):** `seats[].attractionDeck` and `seats[].junkyard` (written only while they hold a card), and the card's `astrotoriumBack` and `attractionLights`. An older file has none of them, which is right for every game before this. An older binary drops them: its Attractions are ordinary artifacts that never visit, and nothing more can be opened. Weaker than printed, never stronger. Recorded with `-update-shape`; no fixture edited, no version bump.
- **Clone** copies both containers and `attractionEntering`; the snapshot does not carry `attractionEntering`, for `outsideEntering`'s reason (non-empty between actions only under a paused entry, whose resume frame the census counts).
- **Closure ratchet:** `Player.AttractionDeck` and `Player.Junkyard` are two new routes to cards, classified `census:IntrinsicAbilityCards` like `Player.Companion` and `Player.Sideboard`, and the ceiling goes from 51 to 53 with the reason written beside it. They are containers of ordinary cards that the census already walks, not a new kind of blocker.

### 8. The wire

`PlayerView.attraction_deck_count` (public, like a library's size; the cards are on nobody's wire) and `PlayerView.junkyard` (public `CardView`s), and `CardView.attraction_lights` (redacted with the rest of a hidden card's identity). The three new events are deliberate log silences: the open is the zone-move line, the roll is the die-roll line, and a visit is its trigger's line.

### 9. Cards

The vocabulary (`cards/effects/attractions.go`): `OpenAttractions{N}`, `OpenAnAttraction`, `WhenThisEntersOpenAnAttraction`, `RollToVisitYourAttractions`, `Visit`, `Attraction()`. Sixteen cards ship, all `full`: Down for Repairs (the reveal-and-discard, then "up to one target Attraction that player controls", whose destruction waits for the pick, CR 608.2c), Deadbeat Attendant, Petting Zookeeper, Rad Rascal, Ride Guide, Seasoned Buttoneer, Step Right Up, Line Cutter, Coming Attraction, Draconian Gate-Bot, Information Booth, Fortune Teller, Concession Stand, Roller Coaster, Kiddie Coaster and Spinny Ride.

## Consequences

- **Left open for #2138 (stickers and tickets).** A sticker is put on a permanent and a ticket counter is a player counter; neither touches the containers above. The obvious attachment points: tickets on `Player.Counters` (as energy and poison are), and a sticker as data on the card it is put on. A ticket spent to put a sticker on an Attraction does not change its lights.
- **Still open, card work on these events, not seams:** "Attractions you've visited this turn" (Soul Swindler, Squirrel Squatters, Black Hole, Storybook Ride) can count `EventVisitAttraction` in the turn's events; "whenever you open an Attraction" (The Most Dangerous Gamer) watches `EventAttractionOpened`; "claim the prize" (Pick-a-Beeble, CR 702.159b) and Monitor Monitor's reroll need their own small pieces.
- **The client** shows an opened Attraction as the artifact it is (its image shows the lights). `attraction_deck_count` and `junkyard` are on the wire for a seat panel to show; no client surface is added here.
