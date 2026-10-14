# ADR 0191 — A sideboard is the outside-the-game pool

**Status:** Accepted · 2026-10-14 (the owner's decision, quoted below, was given the same day this was written) · S58 — Deck requests, October batch (the Reality Fracture finish, tracker [#2795](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2795)).
**Issue:** [#3506](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3506). Related: [#2832](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2832) (the engine gaps behind the set's skipped cards), [#2352](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2352) (game formats, starting with best-of-one Standard).
**Numbering:** checked on 2026-10-14. `develop` ends at 0183 and no remote head holds a higher number. The heads checked, from `git ls-remote --heads origin`, were `cost-ledger`, `develop`, `docs/issue-audit`, `feat/3343-autopass-reason-undo`, `feat/750-conditional-block-restrictions`, `feat/playmats`, `fix/2950-darkness-replacement-continuation`, `fix/3159-3160-attachments`, `fix/3489-max-completion-tokens`, `fix/caddy-reload-admin-off`, `main` and `wip/836-one-click-default`, plus `git log --remotes --name-only -- docs/decisions` over every ref. The issue's claim comment says "ADR 0191 reserved", and no open issue's claim names it.
**Builds on:** [ADR 0172](0172-companion-a-card-outside-the-game.md) (the `ZoneOutside` container this extends, §4, and the question it left open, Consequences), [ADR 0167](0167-controlling-another-players-turn-emrakul-the-promised-end.md) §5 (a "card from outside the game" must offer a controller nothing), [ADR 0064](0064-emblems.md) (a container a `ZoneRef` cannot reach), [ADR 0041](0041-game-persistence.md) and [ADR 0044](0044-surviving-a-deploy.md) (restore points and the additive-snapshot rule), [ADR 0095](0095-deck-coverage-and-deck-requests.md) and [ADR 0110](0110-remember-me.md) (deck coverage, and saved decks as source text), [ADR 0033](0033-ai-bot-seat.md) and [ADR 0122](0122-an-agent-at-the-table-a-local-mcp-seat.md) (the enumerator, bots and the MCP seat).

---

## Context

### The owner's decision

Ian, 2026-10-14: "Add a sideboard. While commander does not allow sideboards, we plan to add other 1v1 formats like standard and the like that do."

ADR 0172 built `ZoneOutside` for one card, a revealed companion, and said so in its Consequences: it "is **not** a general sideboard, and nothing here makes Extrapolate the Impossible, Spawnsire of Ulamog's {20} ability or wishes work. Those need a decision about what a casual player's 'outside the game' is (their whole collection), which this ADR does not take." This is that decision: **a deck's sideboard is that player's outside-the-game pool.** It is not the player's whole collection, and the engine never reads a card index to invent a card.

### The rules

Checked against `MagicCompRules 20260925.txt`, and the Scryfall rulings read on 2026-10-14.

- **CR 400.11:** "Outside the game is not a zone." **400.11a:** sideboard cards are outside the game. **400.11c:** cards outside the game can't be affected by spells or abilities, "except for effects that bring them in". There is nothing for an effect to move out of it except by bringing it in.
- **CR 903.5e:** "Commander games do not use sideboards." This is a deck-construction rule about the 100 cards. It says nothing about a casual player's pool of cards they own and could bring in.
- **CR 723.4** (via ADR 0167 §5): the controller of another player's turn sees what that player can see about objects in the game, but "information about cards outside the game stays with the controlled player."
- **Scryfall, Extrapolate the Impossible (2026-08-21):** "The other card stays outside the game." And: "In a casual game, a card you choose from outside the game comes from your personal collection. In a tournament event, a card you choose from outside the game must come from your sideboard. In Limited events, your sideboard also includes an arbitrary number of basic lands. You may look at your sideboard at any time."
- **Scryfall, Burning Wish (2009, 2004):** in an unsanctioned event you may choose any card from your collection; exiled, phased-out and ante cards can't be acquired.

The last ruling gives the owner's reading its rules footing: the sideboard is what "outside the game" means once there is a deck list, and the player may look at it whenever they like.

### What existed (develop at 8db7ae33c)

- `deck.List.Sideboard []cards.Card` was **parsed and dropped**: all three importers fill it (`ParseText`'s `Sideboard:` header, `SB:` prefix and `// Sideboard`; Moxfield's `sideboard` board; Archidekt's `Sideboard` category), `List.Text()` writes a `Sideboard:` section back, and `Validate` warned `sideboard_not_supported_in_commander` ("they will be ignored"). `ToGameCards()` never included them.
- `game.Player` had `Companion *Zone` of kind `ZoneOutside` (ADR 0172 §4) and nothing else outside the game. `zoneFromRefLocked` and the card index do not cover `ZoneOutside`, so no move verb, effect or admin tool reaches a card there.
- `AddPlayer` and `ReplaceDeck` routed commanders to the command zone and the companion to `Player.Companion` through `seatDeckCard`.
- Saved decks keep **source text**; they re-run parse, resolve and validate when seated, so a section in the text is all a saved deck needs.
- Three prompt shapes already existed that this needs and none had to be invented: `choose_cards` (a pick over cards, with a set-level `Validate`, the enumerator's `ChooseCardsPickLegalLocked`, and a zone re-check on submit), `reveal_pick` ("an opponent chooses from cards you revealed") and the choose-a-player option pick (the engine skips it at two seats).
- #2352 (formats) already plans "a sideboard of at most 15" for Standard and lists best-of-three sideboarding as deferred. Its decklist importers already parse the section.

## Options considered

### Option A — a free pick from the card index

"A card you own from outside the game" as a prompt over any card in Scryfall, the way a paper casual game reads "your personal collection". Rejected by the owner's decision, and it would be stronger than printed in every deck without a sideboard (a Burning Wish that fetches any sorcery in the game), which is the wrong failure mode for this project (AGENTS.md §7: "Stronger than printed is the wrong direction").

### Option B — a second `ZoneOutside` container, `Player.Sideboard` (chosen)

`ZoneOutside` stays the one kind for "outside the game". Each player gets a second `*Zone` of that kind beside `Player.Companion`, built the same way, for the same reason and with the same persistence. The companion's container, its special action and its offer are **not touched**.

- **For:** it is the concept ADR 0172 already accepted, not a parallel one. Every reader of the hand, library, command zone and companion stays correct without an audit, because nothing that moves or finds a card by reference can see it. It snapshots, clones and renders like any other `game.Card`.
- **Against:** one more container, one more field in the census, the snapshot shape and the view.

### Option C — one container for both the companion and the sideboard

Put the sideboard cards in `Player.Companion` and read "the companion" as the card with `IsCompanion`. Rejected: ADR 0172 §6 derives the {3} offer from "a card is in the container" (`CompanionOffer` says every card in it offers the special action) and reads "the container emptied" as "the once-per-game action is spent" (§4). A shared container would offer {3} on every sideboard card, or need an `IsCompanion` exception in each reader. That changes accepted decisions of Luke's ADR rather than extending them.

| | A: card index | B: second container (chosen) | C: shared with the companion |
|---|---|---|---|
| Matches the owner's decision | no | yes | yes |
| Changes ADR 0172's accepted decisions | no | no | yes (§4, §6) |
| Stronger than printed in a deck with no sideboard | yes | no | no |
| New snapshot key | none | `seats[].sideboard` | none, but every companion reader changes |

## Decision

Option B. The seven sections below are the whole design, then the cards, then what #2352 reuses.

### 1. The container

- `game.Player` gains `Sideboard *Zone` of kind `ZoneOutside`, documented in `player.go` and `zone.go` as a second container of the outside-the-game kind. `ZoneOutside` stays "not a zone (CR 400.11)". `zoneFromRefLocked` has no case for it, the card index does not cover it, and `LookupCardForEffect` does not see its cards. That is CR 400.11c enforced by the container, not by a check somebody has to remember (ADR 0064's reasoning for emblems).
- It is `nil` when a deck has no sideboard, never an empty container: every reader asks `Player.SideboardCards()`, which answers nil and empty the same way. An emptied container reads as "nothing outside", which is what it is.
- **The only doors in** are effects that bring a card in, three calls in `game/sideboard.go` and nothing else: `OutsideCandidatesForEffect` reads it, `RevealOutsideForEffect` shows a card to the table, `BringFromOutsideForEffect` moves a card into the owner's hand and emits an `EventZoneMove` from `ZoneOutside` to `ZoneHand`. The first two never move anything. A card that leaves never comes back (the owner's later bounce, shuffle or discard of it is an ordinary card in an ordinary zone).
- `LookupOutsideCardForEffect` is the one read-only lookup, for the surfaces that show or check a card a prompt names (the view, the enumerator, a prompt's set-level `Validate`). It is deliberately a different method from `LookupCardForEffect`, so a card file cannot reach the container by accident.
- **The companion's container is unchanged.** ADR 0172's decisions stand as accepted: `Player.Companion` still holds the revealed companion, the {3} offer is still derived from being in it, a sideboard card is never offered it, and a sideboard card is never an `IsCompanion` card. A test pins that.
- `Card.IsSideboard` is the seating stamp, the way `IsCompanion` is: it routes the card at `AddPlayer` / `ReplaceDeck` / `ToGameCards()` and stays on the card after it reaches a hand, where it grants nothing.

### 2. Commander tables

CR 903.5e says Commander uses no sideboard. The owner wants one anyway, so **in Commander the sideboard is the casual outside-the-game pool, read only by effects that bring a card in.** It is not part of the 100 (the count, singleton and colour-identity checks never see it), it is not a deck-legality input (cards in it are not checked against the commander's colour identity or banlist, the way a companion's are, because nothing but a wish can ever put one in play and a wish is a casual convenience, not a deck-building rule), and nothing about it changes the deck's draws.

The existing `sideboard_not_supported_in_commander` warning keeps its **code** (clients, docs and tests key off it) and its **non-fatal** treatment, and its message now says what the sideboard is used for:

> deck has N sideboard cards; Commander doesn't use a sideboard, so they stay outside the game and only a card that brings one in from outside the game (a wish, Extrapolate the Impossible) can use them

The old text ("they will be ignored") was false the moment this shipped.

### 3. Seating, the deck list and the importers

- `List.ToGameCards()` appends the sideboard after the companion, each card with `IsSideboard` set. `Resolve`, the three importers and `List.Text()` are unchanged: they already parsed, resolved and wrote the section, and a saved deck keeps its **source text**, so the `Sideboard:` section survives a save and a reload with no migration. A deck fetched from a link is saved through `List.Text()`, and a test round-trips it through `Text`, `ParseText` and `Resolve`.
- `AddPlayer` and `ReplaceDeck` route an `IsSideboard` card into `Player.Sideboard` through `seatDeckCard` (after the commander and companion arms, so a card that is both a companion and a sideboard row is a companion). `ReplaceDeck` resets `Sideboard` along with `Library`, `Command` and `Companion`. The deck source ordinals (`setDeckSourceOrdinalLocked`) include sideboard cards, and the lobby-departure path re-stamps them as seats compact. `removeObjectsOwnedByLocked` sweeps a departing player's sideboard with the rest of their cards.
- The owner is made a **knower** of each sideboard card at seating (`seatDeckCard`), and nobody else is. That is what lets the existing per-viewer filter show them to the owner and no one else, and it is how the Scryfall ruling's "you may look at your sideboard at any time" falls out.
- `lobby.uploadDeckResponse` gains `sideboard int` (omitempty) beside `commanders` and `companion`; `card_count` never counts it, and the client's "uploaded" line says "sideboard: N outside the game".
- **A companion in the sideboard (ADR 0172 owner decision 3), now that the sideboard means something.** A sideboard of exactly one card that prints companion, with no explicit `Companion` row, is still read as the companion: it leaves the sideboard, `companion_from_sideboard` still warns, and nothing changes for any list that worked before. **Every other sideboard stays a sideboard.** A sideboard of two or more cards keeps all of them, companion-keyword cards included, as ordinary cards outside the game; a companion has to be declared (`Companion:` section, the `companions` board, the `Companion` category) to be one. An explicit companion wins, and a sideboard copy of the same card is still dropped as one card, not two.

### 4. Persistence, undo and `Clone`

- **`Clone`:** `clonePlayer` copies it with `cloneZone(p.Sideboard)`, beside `Companion`. Undo restores a clone through `RestoreFrom`, so bringing a card in and undoing puts it back (a test).
- **Snapshot (AGENTS.md §5, additive within v8):** `playerSnapshot.Sideboard *zoneSnapshot` with `json:"sideboard,omitempty"`, written only while the container holds a card (`snapshotCompanionZone`, the companion's helper, is reused), restored with `restoreZone(…, ZoneOutside)`, and walked by the continuation census and the ability-ref lookup beside `Companion`. The card snapshot carries `Card.IsSideboard` (`json:"isSideboard,omitempty"`).
  - *Older file, new binary.* The key is absent, which means no sideboard. That is right for every file written before a sideboard could be seated.
  - *New file, older binary (a rollback).* The older binary drops the key, and the player loses the cards an effect could have brought in: weaker than printed, never stronger. A card already in a hand, graveyard or battlefield is an ordinary card and survives; `isSideboard` is an omitempty bool whose absence means nothing.
  - The shape is recorded in `testdata/snapshot_shape/v8.txt` (`-update-shape`) in the same change; no fixture is edited and there is **no version bump**. The drift test classifies `Player.Sideboard` and `Card.IsSideboard` as carried.
- **The closure ratchet.** `Player.Sideboard` is a new route from `Game` to cards that hold closures, so it is classified `census:IntrinsicAbilityCards` like `Player.Companion`, and the class ceiling goes from 50 to 51 with the reason written beside it. It is a container of ordinary cards, not a new kind of blocker.
- **Restorability while a prompt is open.** A pick over the sideboard is a `choose_cards` whose continuation is a closure, so, like every such prompt, it holds the restore window shut until it is answered (ADR 0041, the `ChoiceResumeFrames` census). Nothing new, and the same few seconds every tutor already costs.

### 5. The size cap: 15

A deck's sideboard holds **at most 15 cards**, `game.MaxSideboardSize` = `deck.MaxSideboard`. Over the cap is **not an error**:

- `deck.Validate` reports `sideboard_too_large` as a **non-fatal warning** ("deck has N sideboard cards, over the limit of 15; the sideboard is ignored until it has 15 or fewer"), and `List.SeatedSideboard()` hands the game **no sideboard at all** for such a list. The deck imports, validates and seats exactly as it did before this ADR, with an empty sideboard. The server never picks which cards of an oversized sideboard to keep.
- `AddPlayer` / `ReplaceDeck` keep a guard and refuse a deck that still carries more than the cap with `ErrSideboardTooLarge`, so a caller that bypasses `ToGameCards()` cannot seat an unbounded pool.

Why not fatal (the coordinator's decision, 2026-10-14): saved decks are source text re-validated at seating (ADR 0172 notes this), so a fatal cap would stop an existing Commander deck with a large sideboard or maybeboard section from importing or seating, which is a regression. Why 15: it is the number #2352 plans for Standard, a cap is cheap to raise and expensive to lower once decks are saved against it, and it bounds what rides the wire and the snapshot for a player who pastes a whole collection into the section. The cost is that such a deck silently has no sideboard unless the warning is read; the warning says so.

### 6. The prompt, and the cards that use it

An effect asks for a card with `QueueOutsidePickForEffect`: **a `choose_cards` over the owner's sideboard, not a new prompt kind.** Every rule about the payload, the bounds, the candidate and duplicate checks, the set-level `Validate` (#1017), the empty answer (`choose nothing` is the always-legal answer when the floor is zero) and the enumerator's reach is already written once for that kind. What differs is three small things, and each is a few lines in an existing place:

- the pick declares `choose_destination: "outside"` (`ChooseFromOutside`), which the heuristic and the client read, and the frame's zone re-check is pointed at the container (`pickStillInPickZoneLocked` asks `pickStillOutsideLocked`, so a card that left under an open prompt is pruned and an answer naming it is refused, with no #544 wedge);
- the card lookups behind the prompt (the view's options, the set-level `Validate`, the enumerator's names) fall back to `LookupOutsideCardForEffect`;
- the prompt is **not queued at all** when nothing matches (a clause that finds no card is not an error, CR 608.2c) or when the pool is closed (§8).

`effects.TakeFromOutside` is the one-call shape every wish uses ("You may reveal a *kind* card you own from outside the game and put it into your hand"; `Required` for the unrevealed, mandatory "Put a card you own from outside the game into your hand"). A card that needs more composes the three game-level calls.

**Extrapolate the Impossible** is the card that needed the most, and it is built from existing pieces: the caster's `choose_cards` over the sideboard with `Min 0, Max 2` and a `Validate` that wants exactly two cards with different names (so a single card, a pair of copies or three cards is refused, and the enumerator offers only the pairs the engine accepts); `RevealOutsideForEffect` marks both known to every seat and announces them as one grouped reveal; the caster names an opponent with the choose-a-player pick the engine already skips at two seats (as Fact or Fiction does for "an opponent"); and a `reveal_pick` asks that opponent to choose one. The chosen card goes to the caster's hand; the other stays outside, and stays known to the table, which saw it. A sideboard without two differently named cards asks nothing. This is exactly what is printed, so it ships **Full**.

Proof cards that bring a card in and need nothing else new: **Burning Wish** (a sorcery card), **Living Wish** (a creature or land card) and **Mastermind's Acquisition** (a modal card whose second mode is the mandatory, unrevealed pick and whose first is Demonic Tutor's search). All three are **Full**. The wishes exile themselves as their last instruction (`ExileCardForEffect`, #489); Mastermind's Acquisition does not.

### 7. Visibility and the wire

- `PlayerView` gains `sideboard []CardView` (omitempty) and `sideboard_count int` (omitempty). **`sideboard` is owner-only:** `FilterViewFor` keeps it for the seat's own viewer, through the same knower redaction as every other card list, and drops it for every other seat, every spectator, the admin view and the controller of the owner's turn. A test serialises each of those views and searches the JSON for the card names and instance IDs. **`sideboard_count` is public**, like a hand's size, and answers "a count at most" for spectators.
- **A card revealed from it is public from then on.** `RevealOutsideForEffect` adds every seated player to the card's knowers (the same set a hand reveal writes) and emits one grouped `EventRevealCards` per call with `OldZone: ZoneOutside`; the reveal frame resolves its names from the sideboard while the card is still outside (`resolveRevealCards`), and an opponent asked to choose among the revealed cards sees both through the existing `redactChoiceCards` (they are knowers). It does **not** open the rest of the sideboard to anyone. The log projection treats `ZoneOutside` as a hidden zone, so a move from outside to a hand logs no instance ID (the reveal line already names the card).
- A prompt's options for its own owner go through `viewOfPendingChoices` (the owner's candidates) and `filterPendingChoices` (every other seat gets the open prompt and no candidates, no bounds), the way a hand pick already does.
- `docs/protocol.md` documents the two fields, the `outside` destination and the `from: "outside"` reveal.

### 8. Controlled turns (ADR 0167 §5)

ADR 0167 §5 left a standing instruction: "When a 'card from outside the game' prompt is added, it must offer the controller nothing." The pool is closed while another player makes its owner's decisions: `OutsideCandidatesForEffect` returns nothing when `ControllerOfPlayerLocked(owner) != owner`, so the effect finds no card to bring in, no prompt goes up, and the controller's view of the owner's seat carries no sideboard. This is weaker than printed and never stronger. CR 723.5 lets the controller make the controlled player's choices, and CR 723.4 keeps the information from them, which together say a controller cannot choose a card they may not see; the engine's reading is that nothing is offered. It affects only a wish cast during an Emrakul or Word of Command turn, and it is flagged for the owner below.

### 9. Bots, the enumerator and the MCP seat

Any new prompt must be answerable (#544, #499), and this one adds no new kind, so the existing coverage reaches it:

- **The enumerator** (`legal.EnumerateFor`) already answers `choose_cards` with the empty set first when the floor is zero and every set `ChooseCardsPickLegalLocked` accepts, so "a pair with the same name" is never offered and a mandatory pick has no "choose nothing". `findCardAnywhere` and `cardName` read the sideboard, so the owner's labels name the cards and nobody else's list does. The opponent's half is an ordinary `reveal_pick`.
- **The heuristic** prices a `choose_cards` that declares `outside` and whose candidates are all in the bot's own sideboard as the cards it **takes** (`valueTakenFromOutside`: each named card's value plus the library-take floor), so a bot takes the best card rather than none, and a "reveal two" picks a pair. A prompt whose destination says `outside` over cards that are not the bot's sideboard keeps its old price.
- **The MCP seat** reads the board text, which shows the seat's own sideboard (`your sideboard (outside the game): …`) and each opponent's count, and the move list carries the pick by label. No new tool.
- The built-in bot decks have no sideboard, so in practice the prompt reaches a bot only with a deck someone built for it. The tests drive it with a hand-built one.

### 10. The client

A minimal, read-only affordance, reusing the zone browser: a **sideboard pile** in the player rail (`PileBar`, the fourth tile beside library, graveyard and exile) on your own seat when it holds cards, face down and captioned "outside the game", that opens `ZoneBrowserModal` on a fourth `sideboard` tab, shown only on your own seat. `BrowsableZone` gains `sideboard`; `cardsForZone` reads `seat.sideboard`; `canManageZone` is **false for everyone, the owner included**, so no move button renders and `buildMovePayload` returns null (CR 400.11c, and there is no verb that could carry one). The prompt modal says what the cards are when `choose_destination` is `outside`. Every other seat shows nothing: the server sends it only a count.

## What #2352 reuses

#2352 builds formats and a deck check per format. It should reuse, and not rebuild, this:

- **The container and the seating.** A Standard deck's 15-card sideboard is `List.Sideboard`, seated by `ToGameCards()` exactly as above. `deck.MaxSideboard` is the number the format validator already needs; a format table would hold its own value and `game.MaxSideboardSize` stays the engine's hard ceiling.
- **The warning.** In a format with a sideboard, `sideboard_not_supported_in_commander` does not apply, and the format-aware validator should not emit it. Its message here is Commander's, and is the only place the "stays outside the game" text lives.
- **The owner-only view and the read-only browser.** They are the same for any format.
- **What this deliberately does not do, and #2352 owns:** formats, format legality, a sideboard-size rule per format ("at least 60, at most 15"), best-of-three and swapping cards between games. The container is built so that swapping would move a card between `Library` and `Sideboard` between games with no change to how an effect reads it. Nothing here reads or enforces "this deck is Standard".

## Compatibility

- **Snapshots:** additive within v8 (§4). One new key, one new omitempty card key, new shape lines, no edited fixture, no bump.
- **Wire:** additive. Two optional `PlayerView` fields, one new `choose_destination` value, one new `from` value on a reveal frame, one upload-response field, one new violation code. Older clients ignore all of it, and the server sends no sideboard to a seat that may not see it.
- **Saved decks:** none need rewriting; the section was already in their source text. A saved deck with a sideboard over 15 cards still seats, with an empty sideboard and a warning (§5).
- **Behaviour changes:** a Commander list's sideboard now reaches the game (it was dropped). The warning's text changed and its code did not. A sideboard over 15 is a warning and is not carried into the game. A lone companion card in the sideboard is read as the companion exactly as before.

## Validation

- **`internal/deck`:** the sideboard is kept into the game by each importer (text with each header form and the `SB:` prefix, Moxfield JSON, the real Moxfield text export, Archidekt), `ToGameCards()` stamps and seats it, a `Text()` save-and-reload round trip keeps it, the warning text, the cap at 15 and 16 (a warning, and an empty sideboard in the game), and the companion-from-sideboard rule against a shared sideboard.
- **`internal/game`:** seating, order and knowers; the container is unreachable by `findCardZoneLocked`, `LookupCardForEffect`, `zoneFromRefLocked`, the admin move verb and the bounce, exile and graveyard effects; the cap through `AddPlayer` and `ReplaceDeck`; `ReplaceDeck` and departure; clone, undo and the snapshot round trip, and a file without the key; candidates for the owner only and closed to a controlled player; the bring-in event; the reveal marks knowers and emits one run; the pick's validation, prune and empty answer.
- **`internal/protocol`:** the view for the owner, an opponent, a spectator and the admin with a JSON search for leaks, the controller of a controlled turn, the owner-only prompt options, and the public reveal.
- **`internal/legal`, `aiseat`:** every move the enumerator offers is accepted by the dispatcher; a mandatory pick has an answer; pairs with the same name are not offered; the opponent's pick is enumerated; no other seat's list names a sideboard card; the heuristic takes the best card and the best pair and leaves a non-sideboard pick on its old price; the board text.
- **Cards:** each of the four, including declining, nothing to fetch, the type filters, the mandatory mode, the two-seat and four-seat Extrapolate, the controlled turn, and the oracle fixtures.
- **Client:** the zone logic, the pile (own seat only, face down, count), the browser tab and its read-only note.

## Consequences

- "A card you own from outside the game" works for the decks that carry a sideboard. Extrapolate the Impossible and the first three wishes ship; the rest of the family is card work on a landed seam.
- A player can finally see what they pasted under `Sideboard:`: the pile is the evidence the section was read.
- `ZoneOutside` now has two containers and one meaning. The companion's special action is still its own, and a future card that reads "from outside the game" has one place to look.

## Not built, and filed for the owner

The cards below also read "outside the game" and are not shipped here. Each needs something this ADR does not add; they are listed in the PR body and in one follow-up issue.

- **Wishes with a clause of their own:** Cunning Wish (an instant), Glittering Wish (a multicolored card), Golden Wish (an artifact or enchantment) and Death Wish (a card, and you lose half your life) are `TakeFromOutside` with a different `Match`; Fae of Wishes // Granted is the same on an Adventure face; Wish ("you may play a card you own from outside the game this turn") needs a play permission over a card that is not in a hand, and Coax from the Blind Eternities adds "or choose a face-up Eldrazi card in exile".
- **Karn, the Great Creator:** a loyalty ability that does the same, plus "or choose a face-up artifact card you own in exile".
- **Spawnsire of Ulamog:** the {20} ability casts spells from outside the game, which is a cast path from a container no `ZoneRef` reaches.
- **Research // Development:** shuffles up to four cards from outside the game into the library, a move the other way.
- **Learn** (the Lessons): "you may reveal a Lesson card you own from outside the game and put it into your hand, or discard a card to draw", a `TakeFromOutside` with a branch.
- **Boomerang Basics and other Lessons** already ship as plain sorceries.
- **Deck coverage** does not list the sideboard's cards, so the deck checker cannot yet say a sideboard card is unsupported.

## Decisions on the judgement calls (2026-10-14)

1. **The cap is not fatal** (§5). An over-cap sideboard warns and is not carried into the game, so no existing deck stops importing or seating.
2. **A controlled player's pool is closed** (§8), and the cards stay **Full**. This is an engine-wide policy for "outside the game" under another player's control (ADR 0167 §5), not a per-card caveat: every card that reads the container behaves the same way, so none declares it.
3. **The sideboard is not checked against the commander's colour identity or the banlist.** That is fine for the casual Commander pool, and the check is deferred to #2352's format validation, which will know which format a deck is for.
4. **The owner may see the sideboard at any time**, including in the zone browser during an opponent's turn, which is what the Scryfall ruling says and not a leak.
