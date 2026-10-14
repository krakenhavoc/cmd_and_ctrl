# ADR 0172 — Companion: a card outside the game

**Status:** Proposed (draft for the owner's review) · 2026-10-14 · S58 — Deck requests, October batch.
**Issue:** [#3401](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3401) (the `companion` seam). Requested card: Jegantha, the Wellspring ([#2034](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2034), shipped with the companion caveat in #2170).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-14. I ran `git fetch --all --prune` and listed every file `docs/decisions/` has held on every remote head (`git log --remotes --name-only -- docs/decisions/`; 12 refs, then 15 on the re-check before pushing), and read the files of the open PRs (#3421, #3420, #3418, #3356). `develop` ends at 0171 and no branch or open PR holds a higher number, so this ADR takes **0172**.
**Builds on:** [ADR 0062](0062-abilities-and-special-actions-from-the-hand.md) Decision 4 (one `special_action` verb), [ADR 0082](0082-casting-face-down-and-turning-face-up.md) decision 6 (the per-kind zone table), [ADR 0064](0064-emblems.md) (a second container a `ZoneRef` can't reach), [ADR 0041](0041-game-persistence.md) and [ADR 0044](0044-surviving-a-deploy.md) (restore points and the snapshot rules), [ADR 0095](0095-deck-coverage-and-deck-requests.md) (deck coverage), [ADR 0110](0110-remember-me.md) and [ADR 0112](0112-signed-in-home-player-mode-and-one-decks-page.md) (saved decks as source text), [ADR 0144](0144-choose-a-background-and-the-partner-pairings.md) (the last deck-format change to the command zone), [ADR 0033](0033-ai-bot-seat.md) and [ADR 0122](0122-an-agent-at-the-table-a-local-mcp-seat.md) (the enumerator, bots and the MCP seat), and [ADR 0167](0167-controlling-another-players-turn-emrakul-the-promised-end.md) §5 (a "card from outside the game" offers a controller nothing hidden).

---

## Context

Jegantha, the Wellspring ({4}{R/G}, Legendary Creature — Elemental Elk, 5/5):

> Companion — No card in your starting deck has more than one of the same mana symbol in its mana cost. (If this card is your chosen companion, you may put it into your hand from outside the game for {3} as a sorcery.)
> {T}: Add {W}{U}{B}{R}{G}. This mana can't be spent to pay generic mana costs.

The mana ability shipped in #2170 (`ManaRestrictNoGeneric`). The companion line is the card's caveat: "Companion isn't supported — Jegantha can't be put into your hand from outside the game, and the deck checker doesn't enforce its deck-building condition" (`cards/effects/jegantha_the_wellspring.go`). Nine more companions would follow once the seam exists: Lurrus, Yorion, Kaheera, Gyruda, Zirda, Obosh, Umori, Keruga and Lutri. None is catalogued yet.

### The rules

Checked against `MagicCompRules 20260925.txt`. Rulings are Scryfall's for Jegantha and Lurrus (2020-04-17, 2020-06-01, 2023-04-14), read on 2026-10-14.

- **CR 702.139a:** "Companion is a keyword ability that functions outside the game. It's written as "Companion—[Condition]." Before the game begins, you may reveal one card you own from outside the game with a companion ability whose condition is fulfilled by your starting deck. (See rule 103.2b.) Once during the game, any time you have priority and the stack is empty, but only during a main phase of your turn, you may pay {3} and put that card into your hand. This is a special action that doesn't use the stack (see rule 116.2g)."
- **CR 702.139b:** "If a companion ability refers to your starting deck, it refers to your deck after you've set aside any sideboard cards. In a Commander game, this is also before you've set aside your commander." So the commander is part of the deck the condition is checked against.
- **CR 702.139c:** once the card is in your hand, "it remains in the game until the game ends". **CR 702.139d:** "Cards can enter Commander games from outside the game via the companion special action."
- **CR 116.2g:** "A player who has chosen a companion may pay {3} to put that card from outside the game into their hand. This is a special action. A player can take this action any time they have priority and the stack is empty during a main phase of their turn, but only if they haven't done so yet this game." CR 116.3: the player receives priority afterwards.
- **CR 103.2b:** after the starting player is determined, a player may reveal "no more than one" companion, "only if their deck fulfills the condition". "The revealed card remains outside the game." **CR 103.2c** (the commander goes to the command zone) comes *after* 103.2b.
- **CR 400.11:** "Outside the game is not a zone." **400.11a:** sideboard cards are outside the game. **400.11c:** cards outside the game can't be affected by spells or abilities, except CDAs and effects that bring them in.
- **CR 903.5a** (exactly 100 cards), **903.5b** (singleton), **903.5c** (colour identity), **903.5e** ("Commander games do not use sideboards").
- **Characteristics for the condition:** **CR 202.3e** ({X} is 0 off the stack), **CR 709.4b** (a split card's mana cost is the halves combined, and "an effect that refers specifically to the symbols in a split card's mana cost sees the separate symbols"), **CR 712.8a** (a double-faced card outside the game has only its front face), **CR 715.4** (an adventurer card has only its normal characteristics).
- **The 2020 errata** (Scryfall ruling 2020-06-01): "Instead of casting companions from outside the game: Once per game, any time you could cast a sorcery …, you can pay {3} to put your companion … into your hand. This is a special action, not an activated ability. It happens immediately and can't be responded to. It can't be countered or stopped by cards like Phyrexian Revoker."
- **Rulings:**
  - "Your companion begins the game outside the game. In tournament play, this means your sideboard. In casual play, it's simply a card you own that's not in your starting deck."
  - "You may have one companion in the Commander variant. Your deck, including your commander, must meet its companion requirement. Your companion is not one of your one hundred cards."
  - "The companion ability has no effect if the card is in your starting deck and creates no restriction on putting a card with a companion ability into your starting deck."
  - "If more than one player wishes to reveal a companion, the starting player does so first, and players proceed in turn order. Once a player has chosen not to reveal a companion, that player can't change their mind."
  - "Once you put your companion into your hand, it behaves like any other card you've brought into the game. For example, if it's countered or destroyed, it's put into your graveyard, remaining in the game."
  - "The companion's other abilities apply only if the creature is on the battlefield."
  - Jegantha: "If any one card has the same symbol twice, such as {X}{X}{R} or {(r/g)}{(r/g)}, the companion condition isn't satisfied." Generic mana is "numeric mana symbols … and also {X}".

**A commander can't be its own companion.** The commander is part of the starting deck in Commander (702.139b) and is still in the deck when companions are revealed (103.2b runs before 103.2c), so it is not a card "from outside the game". A commander that prints companion (Jegantha as a commander) is legal: its companion ability simply does nothing (the "no effect if the card is in your starting deck" ruling).

### What exists (develop at 153cfd1b6)

**Deck lists** (`server/internal/deck`).

- `deck.List` has `Commanders`, `Mainboard` and `Sideboard []cards.Card`. `Sideboard` is parsed and never reaches a game. `deck.Entry` has `IsCommander` and `IsSideboard`.
- `ParseMoxfield` (`moxfield.go`) reads a `companions` board and refuses any card in it with `ErrUnsupportedMechanic`. The server fetch is blocked by Moxfield anyway (`moxfieldBlockedMessage` in `lobby/deckcheck.go`, ADR 0095 amendment 2026-09-25), so people paste Moxfield's plain-text export.
- `ParseArchidekt` (`archidekt.go`) sorts on the first recognised category: `Commander`, `Sideboard`, `Maybeboard`, else mainboard.
- `ParseText` (`text.go`) knows `Commander:`, `Deck:`/`Mainboard:`, `Sideboard:` headers (with or without a colon, with or without `// `), the `SB:` prefix and the `*CMDR*` suffix. `List.Text()` writes `Commander:`/`Mainboard:`/`Sideboard:` back out, which is how a deck fetched from a link is saved (ADR 0110 owner decision 7).
- `Resolve` refuses a commander whose oracle text contains `Companion — ` (`mentionsUnsupportedMechanic`), with `UnsupportedMechanicError` / `CodeUnsupportedMechanic`. A companion in the main deck imports as an ordinary card.
- `Validate` (`validate.go`) collects every `Violation`. The lobby treats `CodeSideboardUnsupported` and `CodeUnsupportedLayout` as warnings and every other code as fatal (`lobby/http.go`, the deck-upload path). There is no hook for a per-card condition over the whole deck.
- Saved decks (`decklibrary`) keep the **source text**, not resolved cards, and re-run parse → resolve → validate when seated. A saved row also stores `Commanders []string` for the decks page.
- Deck coverage (`deckcoverage.BuildWith`) resolves the entries, buckets every non-sideboard card, and reports `deck.Validate`'s violations without blocking. `Report.Copies` totals the deck's size, 100 for a legal deck.

**The engine** (`server/internal/game`).

- Zones are `ZoneKind` strings (`zone.go`): library, hand, battlefield, graveyard, exile, command, stack, plus `ZonePhasedOut`. `Player` holds `Library`, `Hand`, `Graveyard`, `Command` and `Emblems *Zone`. There is nothing outside the game. ADR 0167 §5 says so: "The engine has no sideboard or outside-the-game zone".
- **Emblems are the precedent.** They are a second `*Zone` on `Player` rather than more cards in `Command`, because "every reader of Command.Cards means 'commander card'": the cast enumerator offers them, the tax counts them, the admin move verb moves them, `ReplaceDeck` truncates them, and the CR 704.5d token sweep would delete one (`player.go`). `zoneFromRefLocked` (`mutations.go`) can't resolve a `ZoneRef` to `Emblems`, so nothing that moves cards by reference can reach it. It is cloned (`clone.go`), snapshotted as `playerSnapshot.Emblems` (`emblems,omitempty`) and walked by the census.
- `AddPlayer` and `ReplaceDeck` (`game.go`) take a `[]Card` and route `IsCommander` cards to `Command` and the rest to `Library`. `deck.List.ToGameCards()` builds that slice.
- **Special actions** (`special_action.go`). One verb for five kinds: foretell, suspend, plot, turn face up and unlock. Each kind differs in three tables: `specialActionZone` (where the card is), `SpecialActionTimingOKLocked` (when) and `specialActionPerformer` (what it does). `PerformSpecialAction` checks the offer and the window before paying, pays through `payAbilityManaCostLocked` with a zero spend context (a special action is neither a cast nor an activation), emits `EventSpecialAction`, calls `noteActionTakenLocked` (CR 117.3c) and runs state checks. `TurnFaceUpOffer` and `UnlockOffers` *derive* an offer from the card rather than a catalog declaration. Plot and unlock already use `SorcerySpeedOpenLocked`, which is exactly 116.2g's window. (The file's header comment cites turn face up as "CR 116.2g"; in the September 2026 rules that is 116.2b, and 116.2g is companion. The implementation PR fixes the citation.)
- **The wire.** `actions.TypeSpecialAction` takes `{card_id, kind, strict, auto_tap, cost, door}` (`actions/actions.go`). `legal/special_actions.go` walks the hand, the top of the library (grants) and the battlefield, asks the same timing table, and prices with `SpecialActionSpendContext`. `protocol.CardView.SpecialActions` carries the offers to the owner only (`redactCardForViewer` clears them for everyone else). The legal-action lookup lists `special_actions` kinds per card (`protocol/legal_actions.go`).
- **Bots and the MCP seat.** The heuristic prices any special action as the flat `SpecialActionValue` (1.00), except unlock (`aiseat/heuristic/moves.go`). The MCP seat shows legal moves by label, and `aiseat/boardtext` renders "your command zone: …" and the rest of a seat's board.
- **Snapshots.** `SnapshotSchemaVersion = 8`. Within a version a change is additive only: "a new key whose zero value is right for every older file AND whose absence a binary from before it can live with" (`snapshot.go`). New snapshot fields are recorded in `internal/game/testdata/snapshot_shape/v8.txt` with `-update-shape`. ADR 0166 added `playerControls` additively in v8 this way.
- **The client.** Your own commanders sit in the castable strip beside your hand (`ExileStrip.svelte`, #2202), and every other seat's command zone is drawn by `CommandStrip.svelte`. Special-action rows come from `special_actions` on a card (`specialActionItems`, `contextMenu.logic.ts`).

## Options considered

### Option A — the companion as a command-zone-like card

Put the companion in `Player.Command` with a flag (`Card.IsCompanion`), show it with the commanders, and offer the {3} special action from the command zone. No new container, and no new snapshot key beyond the flag.

- **For:** smallest diff. `CommandStrip` and `ExileStrip` already draw the command zone, and a card there already has an instance ID, a view and a snapshot.
- **Against:** every reader of `Command.Cards` means "commander card". That is the reason emblems got their own slice. The cast enumerator would offer to *cast* the companion from the command zone. Commander tax, commander damage, CR 903.9 returns, the admin move verb, `ReplaceDeck` and the deck checker's commander counting would each need an `IsCompanion` exception. A missed exception makes the card stronger than printed (castable for its mana cost, never paying {3}), which is the wrong failure mode for this project. It is also wrong in the rules: the companion is not in the command zone (CR 400.11). ADR 0167 §5 asks for a "card from outside the game" to be handled as its own thing.

### Option B — a holding area outside the game, with the full deck-format and validation work (recommended)

A per-player container for the revealed companion that no `ZoneRef` reaches, built the way emblems are. Add a companion slot to the deck format and every importer. The server checks the condition per companion. A sixth special-action kind moves the card into the hand.

- **For:** each existing reader of the command zone, hand and library stays correct without being touched. The special action reuses the one verb, its timing table, its payment, its enumerator walk and its client row. The card is a normal `game.Card` with an instance ID, so it snapshots, clones and renders like any other card.
- **Against:** more surface. It touches the deck format, three parsers, the validator, a new zone kind, the snapshot shape, one view field, the enumerator walk, the bot and the client. Each condition is per-card code.

### Option C — lift the import refusal and do nothing else

Accept a companion line and either drop it with a warning or put the card in the library. That is what a sandbox would do by hand today. Rejected: it either loses the card or makes the deck 101 cards and changes its draws. It also settles nothing that Option B doesn't.

| | A: command-zone-like | B: outside the game (recommended) | C: lift refusal only |
|---|---|---|---|
| New snapshot key | `Card.isCompanion` | `Player.companion` zone (+ card flag) | none |
| Commander readers to audit | every one | none | none |
| Failure mode if a reader is missed | stronger than printed | n/a | n/a |
| Deck-condition check | same work as B | per companion | none |
| Matches CR 400.11 | no | yes | no |

**Recommendation: Option B.** It costs one more container than A and saves an audit of every command-zone reader. Validation is the same work in either option.

## Decision (proposed)

### 1. The deck format: one companion slot

- `deck.List` gains `Companion *cards.Card` (`json:"companion,omitempty"`), and `deck.Entry` gains `IsCompanion bool`. A list holds at most one companion (CR 103.2b).
- **Pasted text** (`ParseText`). Add a `Companion` section header in the same forms as the others: `Companion`, `Companion:` and `// Companion`. That is the header MTG Arena exports use. No suffix marker is invented. `List.Text()` writes a `Companion:` section after `Commander:`, so a fetched deck saves and reseats with its companion.
- **Moxfield JSON** (`ParseMoxfield`). The `companions` board becomes `IsCompanion` entries instead of `ErrUnsupportedMechanic`. The fetch is still blocked, so in practice this is the pasted JSON path and the fixture tests.
- **Moxfield plain text.** This is the route people actually use. How Moxfield's "Copy plain text" export labels a companion (its own section, the sideboard, or nothing) has not been checked against a real export. The implementation PR adds a real export as `deck/testdata` and parses whatever it shows. See owner question 4.
- **Archidekt** (`ParseArchidekt`). A `Companion` category (case-insensitive, first recognised category wins, as with `Commander`) maps to `IsCompanion`. Whether Archidekt users label a companion that way also needs a real public deck in `deck/testdata` before merge.
- **A companion in the sideboard.** Commander has no sideboard (CR 903.5e), and builders commonly put the companion there. The proposal is that when no row is marked companion and the sideboard holds exactly one card, and that card prints a companion ability, it is read as the companion and a warning says so (`companion_from_sideboard`). In every other case the sideboard keeps today's `sideboard_not_supported_in_commander` warning. See owner question 3.
- **`Resolve`** stops refusing a commander that prints companion. That card is in the starting deck and its companion ability does nothing (ruling). `mentionsUnsupportedMechanic` is kept for companions whose condition isn't implemented yet (§2).
- **Saved decks** need no migration, because a saved deck is source text and the text dialect gains the section. The decks page shows the companion beside the commanders. That can be derived when the row is written, as `Commanders` is, or with an additive column. The implementation picks one, and if it adds a column, it goes in a migration of its own.

### 2. Validation: the server checks the condition

`deck.Validate` gains a companion block, run only when `List.Companion` is set:

| Code | Fatal | When |
|---|---|---|
| `not_a_companion` | yes | the card has no companion ability |
| `too_many_companions` | yes | more than one companion row |
| `companion_is_commander` | yes | the companion row names a commander of the same list (see owner question 5) |
| `companion_condition_not_met` | yes | the condition fails. The message names the condition, and `Card` names the first offender. The violation list carries up to 10 offenders, one row each |
| `companion_not_legal` | yes | the card is not legal in Commander (Scryfall `legalities.commander`, as for every other card) |
| `companion_color_identity` | yes | outside the commander's colour identity (see owner question 2) |
| `unsupported_mechanic` | yes | a companion whose condition is not implemented yet (today's code, now only for those cards) |

The 100-card count, singleton and the colour identity of the 100 are unchanged. The companion is not one of the 100 cards (ruling), so it is excluded from all three.

**The condition registry.** Conditions are code in `internal/deck` (new `companion.go`), keyed by oracle ID, one `func(starting []cards.Card) []string` per companion. The function returns the offending card names. "Starting deck" is `Commanders` plus `Mainboard` (702.139b). The registry is the list of supported companions: a companion with no entry is `unsupported_mechanic`. It reads only Scryfall data (`mana_cost`, `cmc`, `type_line`, the faces), the same data the colour-identity check reads. It does not read the effects catalog: `internal/deck` can't import it (see the `deckcoverage` package comment), and a deck condition is about printed characteristics, not engine support.

**What a failure does.** At seat time a fatal violation refuses the deck, as any other fatal violation does today, and the client highlights the row. The player can fix the list or delete the companion line. The deck checker (ADR 0095) reports the violations without blocking, as it does now. The bot's built-in decks have no companions. See owner question 1 on whether to check conditions at all.

**Deck coverage.** `deckcoverage.Report` gains `companion *Card` (additive, ADR 0095 §1 style). It is bucketed like any card but kept out of `Cards`, `Counts` and `Copies`, so `Copies` still totals 100 for a legal deck. The deck request checklist lists it under its own heading.

### 3. Seating

`List.ToGameCards()` appends the companion with a new `game.Card.IsCompanion` routing flag. `AddPlayer` and `ReplaceDeck` route an `IsCompanion` card into `Player.Companion` (§4) and refuse a second one (`ErrTooManyCompanions`), the way they route `IsCommander`. `ReplaceDeck` resets `Companion` along with `Library` and `Command`. The lobby's `SetDeck`, practice games (`lobby/practice_http.go`) and saved setups (`lobby/setups.go`) need no change because they pass the slice through. The upload response (`lobby/http.go`) gains `companion string` beside `commanders`, and the seat's deck label shows it.

`IsCompanion` stays on the card after it reaches the hand. It is harmless there and useful: the log, the game statistics (ADR 0165) and the client's "companion" badge read it. It never grants anything.

### 4. Where the companion lives: `Player.Companion`

- `Player` gains `Companion *Zone`, holding zero or one card. It has a new zone kind, `ZoneOutside ZoneKind = "outside"`, documented in `zone.go` as "not a zone (CR 400.11): the container for a revealed companion". As with `Emblems`, `zoneFromRefLocked` has no case for it, so `move_card`, every `MoveCardByID*` path, every effect and the admin move verb can't reach it (CR 400.11c). A test pins that.
- Being in this container *is* the companion state. The card leaves only through the special action (§6), so an empty container means either no companion or one already taken. That gives "only if they haven't done so yet this game" (116.2g) with no separate flag. Even if the card later goes back to the hand or the library, it does not come back here (702.139c).
- **Clone and undo.** `clonePlayer` copies it as `cloneZone(p.Companion)`, beside `Emblems`.
- **Snapshot.** `playerSnapshot` gains `Companion *zoneSnapshot` with `json:"companion,omitempty"`, restored with `restoreZone(p.Companion, ZoneOutside)` and walked by the census beside `Emblems`. A table with a companion outside the game is an ordinary restore point. **This is additive within v8:**
  - *Older file, new binary.* The key is absent, which means no companion. That is right for every file written before this change.
  - *New file, older binary (a rollback).* The older binary drops the key, and the player loses the option to take their companion: weaker than printed, never stronger. A companion already in the hand, graveyard or battlefield is an ordinary card and survives. `Card.IsCompanion` is an omitempty bool whose absence means no badge.

  The shape is recorded in `testdata/snapshot_shape/v8.txt` (`-update-shape`) in the same change. A corpus board with a companion waiting is added to the v8 set (the writer adds a file for a new board without a bump), and `TestSnapshotCorpusBoardsStillMatchTheirFixtures` guards it from then on. No fixture is edited. The closure ratchet sees nothing new, because the zone and the card are plain data. See owner question 6 if the owner wants a bump anyway.

### 5. The reveal

When the game starts (`start`, after the opening roll has settled the starting player and before libraries are shuffled and hands drawn: CR 103.2b before 103.3), each seat with a companion emits `EventCompanionRevealed` in seat order from the starting player, as the ruling orders it. Its log line is "Alice reveals Jegantha, the Wellspring as her companion". Declaring a companion in the deck list *is* the choice to reveal it, so there is no pre-game prompt. See owner question 7.

The companion is public: it is revealed and remains revealed. `PlayerView` gains `companion *CardView` (`json:"companion,omitempty"`), unredacted for every viewer, with the same posture as `emblems`. `special_actions` on it are filled for the owner only, as `redactCardForViewer` already does for every card. Under an ADR 0167 controlled turn the controller sees what everyone sees. The card isn't hidden information, so 723.4's "outside the game" sentence has nothing to withhold.

### 6. The special action: `SpecialActionCompanion`

A sixth kind, `SpecialActionCompanion SpecialActionKind = "companion"` (CR 116.2g, 702.139a), added in the three tables and nowhere else, as ADR 0062 Decision 4 and ADR 0082 decision 6 intend:

- **Zone:** `specialActionZone` returns `ZoneOutside`, and `specialActionCardLocked` finds the card in `p.Companion`. A container holds only its owner's companion, so finding it there is the ownership check.
- **Offer:** derived, like `TurnFaceUpOffer`. `CompanionOffer(c Card) *SpecialAction` returns `{Kind: companion, Cost: "{3}", Label: "Put into your hand (companion)"}` for any card in the container. No card file declares it, so a companion's catalog entry needs nothing for the action. The card's own abilities don't apply outside the game (ruling).
- **Timing:** `SpecialActionTimingOKLocked` returns `SorcerySpeedOpenLocked(playerID)`, the same row as plot and unlock. Split second never applies because the window needs an empty stack. The "once per game" part is the container being empty (§4).
- **Cost and payment:** {3} through the existing path, with a **zero spend context**. Mana restricted to casting can't pay it, and Jegantha's own `ManaRestrictNoGeneric` mana can't pay it either, because {3} is generic. Phyrexian Revoker and other "can't activate" effects don't apply, because this is not an activated ability (2020 errata ruling). No printed card modifies this cost today. If one appears, `CostModifier.SpecialActions` is already the hook.
- **Performer:** `companionLocked` removes the card from `p.Companion` and puts it into `p.Hand` as a new object (CR 400.7). It emits the usual zone-change event and the existing `EventSpecialAction`, then `noteActionTakenLocked` and state checks run as for every kind. Every seated player is added as a knower of the card (`AddKnowersAll`), because everyone saw it revealed, so opponents' views show it face up in the hand until it changes zones.
- **Priority:** the player keeps priority (CR 116.3) and may cast it at once if it's legal.
- **Controlled turns:** under ADR 0167, `special_action` is already one of the routed verbs. The controller can take the controlled player's companion action, paid from the controlled player's mana (CR 723.5a). No change is needed.

### 7. Wire, legal moves, bots and the MCP seat

- **Protocol.** No new action type. `special_action {card_id, kind: "companion", strict, auto_tap}` is the whole wire surface. `card_id` is the companion's instance ID, which the owner reads from `PlayerView.companion`. `docs/protocol.md` gets the new kind and the new view field. The log adds a `companion_revealed` kind. A special action already logs as `LogSpecialAction`.
- **Legal moves.** `specialActionMoves` walks `p.Companion.Cards` with `game.ZoneOutside`, the same `specialActionMovesForCard` call as the hand. Timing and affordability come from the engine's own table and pricing, so the bot is never offered a move the engine refuses (#544). The move label is "Pay {3}: put Jegantha, the Wellspring into your hand (companion)". The legal-action lookup lists it under the companion's card ID, which the client's click-to-act reads.
- **Bot heuristic.** The flat `SpecialActionValue` is not enough: the action is effectively "pay {3}: draw a known card". `valueOfCompanion` prices it as the heuristic's card-draw value plus a share of the card's cast value. It is preferred when the bot has three mana it would otherwise float in its main phase and is never chosen over a castable spell of greater value. Exact weights are for the implementation PR, tuned against the bot suite (`aiseat/suite`).
- **MCP seat.** `boardtext` renders a line per seat, "companion (outside the game): Jegantha, the Wellspring", and "taken" once it has been. The action shows in `legal_moves` with the label above. `get_state` includes `companion` through the view. No new MCP tool is needed.

### 8. The client

- **Your companion** sits at the start of the castable strip beside your hand (`ExileStrip`), with a "Companion" caption and a {3} price badge, in the style the strip already uses for foretold cards. A click, or the context-menu row from `specialActionItems`, sends the special action. Outside its window the card is greyed like any card waiting on its window. Once taken, it leaves the strip.
- **Other seats' companions** are shown in `CommandStrip` after the commanders, captioned "Companion" and without a tax badge, so every revealed companion is visible where that seat's commander is.
- **The lobby and decks page.** The deck panel shows "Companion: Jegantha, the Wellspring" under the commanders. A fatal companion violation highlights the companion row like any other violation.
- **Labels and hints.** New strings go through `client/src/lib/labels.ts` (ADR 0125 §2), so `docs/labels.md` regenerates. A first-use hint for the strip is optional.

### 9. Card scope

**Jegantha first.** Its condition: for each card in the starting deck (commanders plus mainboard), take the printed mana cost's symbols and fail if any symbol occurs twice. The details:

- Symbols are compared exactly as printed. `{R/G}{R/G}` fails, and so does `{X}{X}{R}`. `{R}{R/G}` passes because those are different symbols.
- A generic numeral can repeat only in a split card. A split card's cost is both halves combined, but the symbols are seen separately (CR 709.4b), so Fire // Ice's `{1}{R}` + `{1}{U}` has `{1}` twice and **fails**.
- A double-faced card is its front face (CR 712.8a), and an adventurer card is its normal characteristics (CR 715.4). A card with no mana cost (a land) passes.

The tests cover each of these against the real Scryfall dump. After that, Jegantha's caveat is removed and its card file comment rewritten. The roadmap gets a `companion` seam entry for #3401, closed by the implementation PR. Its `Waiting` list is the nine companions below.

**The other nine**, as the registry would express them. Each is one function in the same file and is added with its card:

| Companion | Condition over the starting deck | Expression | Note |
|---|---|---|---|
| Lurrus of the Dream-Den | each permanent card has mana value 2 or less | type line has a permanent type → `cmc ≤ 2` | X = 0 (202.3e) |
| Kaheera, the Orphanguard | each creature card is a Cat, Elemental, Nightmare, Dinosaur or Beast | creature type line → subtype in set | changelings pass (CR 702.73a, a CDA that works everywhere) |
| Gyruda, Doom of Depths | only cards with even mana values | `cmc % 2 == 0` | lands are 0, so they pass |
| Obosh, the Preypiercer | only cards with odd mana values, and land cards | nonland → `cmc % 2 == 1` | |
| Keruga, the Macrosage | only cards with mana value 3 or greater, and land cards | nonland → `cmc ≥ 3` | |
| Umori, the Collector | each nonland card shares a card type | intersection of nonland cards' card types is non-empty | |
| Lutri, the Spellchaser | each nonland card has a different name | always true in a singleton deck | the check is free, so register it |
| Zirda, the Dawnwaker | each permanent card has an activated ability | needs "has an activated ability" from oracle text (a `cost: effect` line, keyword abilities that are activated, such as equip and cycling) | the hardest to get right; its own issue |
| Yorion, Sky Nomad | at least twenty cards more than the minimum deck size | Commander's minimum is exactly 100 (903.5a), so never satisfiable | register as always failing with a clear message |

## Implementation slices

1. **Deck format and validation** (`internal/deck`, `deckcoverage`, lobby upload response). Add the companion slot, the parsers, `Text()`, the condition registry with Jegantha, the violations and the coverage field. Real Moxfield-text and Archidekt fixtures go in `deck/testdata`. The deck still seats without the companion in the game until slice 2. This slice could merge on its own, with a non-fatal `companion_not_in_game_yet` warning, but the recommendation is to land 1–3 together so a deck never says it has a companion the table can't use.
2. **The engine:** `ZoneOutside`, `Player.Companion`, seating, clone, snapshot shape and corpus board, the reveal event, `SpecialActionCompanion` with its three table rows, and the view field. Fix the stale CR 116.2g citation in `special_action.go`.
3. **Moves and play surfaces:** the enumerator walk, the bot pricing, boardtext, the client strips, lobby labels and `docs/protocol.md`. Remove Jegantha's caveat, close the roadmap seam and close #3401.

## Compatibility

- **Snapshots:** additive within v8 (§4). New shape lines and a new corpus board, no edited fixture, no bump.
- **Wire:** additive. A new `kind` value on an existing action, two optional view fields, one log kind and one upload-response field. Older clients ignore the fields, and the server never offers the action to a seat that can't see it.
- **Saved decks:** none need rewriting. A saved deck with no `Companion:` section behaves as before. A saved Moxfield-JSON deck with a `companions` board, which today fails at seat time, starts working.
- **Behaviour change:** a commander that prints companion (Jegantha as a commander) is no longer refused at import.

## Validation

- **`internal/deck`:** each parser's companion path, from real fixtures. The `Text()` round trip. Every violation code. Jegantha's condition on each edge case in §9 (hybrid twice, {X}{X}, split card, DFC, adventurer, land, the commander counted). A companion is excluded from the 100 and from singleton. A companion commander imports.
- **`internal/game`:** seating routes the companion. `ReplaceDeck` resets it. `zoneFromRefLocked` can't reach it. The reveal events come in starting-player order. The special action's window (each step and phase, the opponent's turn, a non-empty stack), its cost (restricted mana refused, Jegantha's mana refused, auto-tap), once only, the knowers, priority kept, and the card behaving as a normal card afterwards (cast, countered to the graveyard, never back to the container). Clone and undo, snapshot round trip, shape and corpus. A controlled turn (ADR 0167) taking the action with the controlled player's mana.
- **`internal/legal` and `aiseat`:** the move is offered exactly when the engine accepts it. Bot pricing goes through the suite harvest (`aiseat/suite/harvest.go` already counts `special_action`).
- **Client:** the strip and `CommandStrip` render tests, the context-menu row and the label contract.

## Consequences

- Companion becomes a deck-format feature with a real check, and the nine other companions become card work: one condition function and a card file each.
- `ZoneOutside` is the first container for "outside the game". It holds only a revealed companion. It is **not** a general sideboard, and nothing here makes Extrapolate the Impossible, Spawnsire of Ulamog's {20} ability or wishes work. Those need a decision about what a casual player's "outside the game" is (their whole collection), which this ADR does not take.
- The deck checker can now give a precise reason for refusing a companion deck, where today it gives a blanket `unsupported_mechanic`.

## Open owner questions

1. **Validate the condition at all?** The recommendation is yes: the server checks the condition, failure is fatal at seating and reported by the deck checker. The alternative is to trust the player, warn only, and seat the companion anyway. That is sandbox-friendly, but it lets any deck use any companion.
2. **Colour identity.** CR 903.5c covers cards *in* the deck, and the companion isn't one. The Commander format's own guidance has companions follow the commander's colour identity, but this ADR could not confirm that page's current text. Enforce `companion_color_identity`? The recommendation is yes.
3. **A lone companion in the sideboard.** Read it as the companion with a warning (the proposal), or require an explicit `Companion` section or board?
4. **Moxfield plain-text export.** Can you paste a real Moxfield plain-text export of a companion deck (and an Archidekt link, if handy)? The parsers should match what the exports actually emit, not a guess.
5. **The same card as commander and companion.** The commander card itself can never be the companion (103.2b before 103.2c). A second copy of the same name as the companion is arguably legal, since singleton covers only the 100. Refuse it (`companion_is_commander`, the proposal, which also keeps a pasted list unambiguous) or allow it?
6. **Snapshot version.** Keep the companion additive in v8 (the proposal: on a rollback the companion option is lost, weaker than printed), or bump to v9 so an older binary refuses the file?
7. **No pre-game reveal prompt.** Declaring a companion in the list is the choice to reveal it. Is that enough, or do you want a pre-game "reveal your companion?" prompt that lets a player sit it out for one game without editing the deck?
8. **Bots.** Should any built-in bot deck (`internal/decks`) get a companion, so the bot path runs in normal games, or is a test deck enough?
9. **Scope of the first PR.** Ship slices 1–3 together (the proposal), or let slice 1 merge alone with the `companion_not_in_game_yet` warning?
