# ADR 0198 — Outside the game is the card database when a deck has no sideboard

**Status:** Accepted · 2026-10-16 (the owner's decision, quoted below, was given on 2026-10-15) · S76 — Cards: incubate, face-down and outside the game (tracker [#3616](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3616)).
**Issue:** [#3858](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3858).
**Numbering:** checked on 2026-10-16. `develop` ends at 0197 and no remote head holds a higher number. The heads checked, from `git ls-remote --heads origin`, were `cost-ledger`, `develop`, `docs/issue-audit`, `feat/3861-planeswalker-caveats`, `feat/750-conditional-block-restrictions`, `feat/playmats`, `fix/2947-resolution-copy-casts`, `fix/2950-darkness-replacement-continuation`, `fix/3490-rebound-e2e`, `fix/3860-targeting-caveats`, `fix/3934-playmat-battlefield-size`, `fix/caddy-reload-admin-off`, `main`, `screenshots/3934-playmat` and `wip/836-one-click-default`. No open issue's claim comment names 0198. The issue's claim comment says "ADR 0198 reserved".
**Amends:** [ADR 0191](0191-a-sideboard-is-the-outside-the-game-pool.md), whose Option A ("a free pick from the card index") the owner rejected on 2026-10-14 and now accepts for decks with no sideboard. Everything else in ADR 0191 stands.
**Builds on:** [ADR 0172](0172-companion-a-card-outside-the-game.md) (`ZoneOutside`), [ADR 0167](0167-controlling-another-players-turn-emrakul-the-promised-end.md) §5 (a controller is offered nothing from outside the game), [ADR 0041](0041-game-persistence.md) (the closure ratchet and restore points), [ADR 0033](0033-ai-bot-seat.md) and [ADR 0122](0122-an-agent-at-the-table-a-local-mcp-seat.md) (the enumerator, bots and the MCP seat), [ADR 0149](0149-privacy-notice-and-account-deletion.md) (what is stored about a person).

---

## Context

### The owner's decision

Ian, 2026-10-15, on #3858, both halves:

1. A deck may carry an optional sideboard or wishboard. The importers read it, saved decks store it, it is never part of the library and never counted for deck validation beyond a sane size cap. When the deck has one, "a card you own from outside the game" is chosen from it (CR 400.11a). A card fetched from it leaves the sideboard.
2. Without a sideboard (plain Commander), "outside the game" offers any card from the card database that matches the filter. A minted card is a real card owned by that player, imported like any deck card so the catalog hooks apply.

Half 1 shipped as ADR 0191 (#3526, #3685). This ADR is half 2.

### The rules

- **CR 400.11 / 400.11a / 400.11c** (checked for ADR 0191): outside the game is not a zone; a sideboard is outside the game; cards outside the game can't be affected except by effects that bring them in.
- **CR 903.5e:** Commander games do not use sideboards.
- **Scryfall, Extrapolate the Impossible (2026-08-21):** "In a casual game, a card you choose from outside the game comes from your personal collection. In a tournament event, a card you choose from outside the game must come from your sideboard."

ADR 0191 read the sanctioned half: the sideboard is the pool. A Commander deck has no sideboard by rule, so for most tables every wish did nothing. The casual half of the same ruling is the one a Commander table actually plays by, and C2 has no collection to read, but it has the card database, which is a superset of any collection.

### What ADR 0191 said against it, and why that no longer decides it

ADR 0191 Option A was rejected for two reasons: the owner's decision, and "it would be stronger than printed in every deck without a sideboard (a Burning Wish that fetches any sorcery in the game)". The first is reversed. The second is a misreading of what "printed" is: in a casual game the printed card *does* fetch any sorcery the player owns, and a player who owns every card is a legal player. The database is the collection of a player with a large collection. It is not stronger than the card; it is the card played at a casual table, which is the table C2 is.

## Decision

### 1. Which pool a player uses

`game.Player.HasSideboard` records whether the player's deck **seated** a sideboard (`seatDeckCard` sets it; `ReplaceDeck` clears it). It is a fact about the deck, not the container:

- **`HasSideboard` true:** the sideboard is the pool, exactly as in ADR 0191. A sideboard emptied by wishes is still the pool, and an empty pool brings nothing in. The database is never consulted.
- **`HasSideboard` false:** the pool is the card database, `game.OutsideCollection`.

A deck whose sideboard was over the 15-card cap seats none (ADR 0191 §5) and so uses the database, which is what its warning already tells the player ("the sideboard is ignored").

`game.Game.QueueOutsidePickForEffect` is still the one door every card uses. It branches on the flag, so every card already built on it (the eight wishes, Mastermind's Acquisition, Extrapolate the Impossible) gains the second half without a change to its file. The controlled-turn rule (ADR 0167 §5) closes both pools the same way.

### 2. The source, and minting

`game.OutsideCollectionSource` is the database as a collection, in the shape of every other hook into `internal/cards` (`TokenArtResolver`, `EffectResolver`): an interface in `game`, installed by `main.go` from `deck.NewCollection(cardIdx)`. Nil means no index, and a player with no sideboard then has nothing outside the game (ADR 0191's behaviour, and every game test's).

- **Search** (name match, best first), **Named** (one card by exact name) and **Suggest** (a walk in suggestion order: cards with a catalog entry first, then Commander-legal, then the most reprinted, then by name). The reprint count is a cheap proxy for "widely played" that the dump already carries.
- Every card it returns is built by `deck.toGameCard`, the function a deck import uses. A taken card has the same oracle ID, printing and faces a deck list would have seated, so every catalog hook keyed by its oracle ID applies.
- Only collectible cards: playable printings (`cards.Card.Playable`) whose type line is not a plane, phenomenon, scheme, conspiracy, vanguard, dungeon or emblem.

On an accepted answer the game **mints** each named card (`mintFromCollectionLocked`): a fresh instance, owned and controlled by the chooser, known only to them, pushed into their outside-the-game container (`Player.Sideboard`, created if absent, `HasSideboard` left false). The effect's continuation then runs exactly as a sideboard pick's does, over the new instance IDs: `RevealOutsideForEffect`, `BringFromOutsideForEffect`, the same events. Nothing downstream can tell which pool a card came from, which is the point. A card the effect leaves outside (Extrapolate the Impossible's unchosen card) stays in the container, outside the game and known to whoever saw it, as it would on a table; the owner sees it in their sideboard pile.

### 3. The prompt and the wire

A new pending-choice kind, **`outside_collection`**, because the candidate set is not a set of objects: a `choose_cards` answers with instance IDs, and no card in the database has one until it is minted.

- **Data on the choice:** `ChooseMin`, `ChooseMax`, `ChooseDestination` (`outside`) and `CollectionSuggestions` (up to 8 card names the filter accepts, in suggestion order). Plain data, snapshotted additively within v9.
- **The continuation** rides the existing `searchResumeFrame` (a library search's `Pred`, `Validate`, `Then`). Reusing a frame type adds no route to `testdata/closure_fields.txt`, so the ratchet does not move. Like every prompt with a continuation, an open one holds the restore window shut until it is answered.
- **The answer** is `resolve_choice` with `{choice_id, card_names: [...]}`, routed by kind (the empty list declines, when the floor is zero). `ResolveOutsideCollection` checks the count against the bounds, resolves each name through `Named`, runs the filter on the template's printed characteristics, then the set rule, and only then dequeues and mints. One copy of that check (`collectionTemplatesLocked`) serves the resolver and the enumerator (`CollectionPickLegalLocked`).
- **The view:** `PendingChoiceView.collection_options` is the suggestions as `CardView`s with no `instance_id`, plus `choose_min` / `choose_max` / `choose_destination`. All of it goes to the chooser only; every other seat sees that the prompt is open and whose it is.
- **The search:** `GET /games/{id}/outside/cards?choice=<id>&q=<text>` (`lobby.OutsideCardsRoute`). It answers only the seat that owes the named open pick, never a spectator or the admin, and only with cards that pick's filter accepts (`Game.SearchOutsideCollection`, at most 40 rows). The rows are the dev spawner's row shape (`id`, `name`, `type_line`, `mana_cost`, `set`), so the client reuses the spawner's search list. It is not a general card search: outside an open wish it returns 404.
- **The client** renders the prompt as a dock sheet in `ChoicePromptModal` (logic in `client/src/lib/outsideCollection.ts`): the question, a search box (debounced, the in-flight request aborted by the next, the card spawner's search pattern) over the route above, and a list of cards with their cost and type line, the suggestions until the player searches. Clicking a row picks it (a second click unpicks; a one-card pick replaces), the dock's primary button is "Take <name>", and "Take nothing" declines when `choose_min` is zero.

### 4. The bot, the enumerator and the MCP seat

- **The enumerator** offers "take nothing" (AlwaysLegal) when the floor is zero, and each suggestion (and each pair of suggestions, for a two-card pick) that `CollectionPickLegalLocked` accepts, capped at `MaxExpansionPerSource`. A mandatory pick always has an answer, because the prompt is not queued when nothing in the database matches. Labels name database cards, which hide nothing.
- **The heuristic** prices an answer as the cards it brings into hand: each named card's `cardValue` plus the library-take floor, read from `collection_options`. Declining scores zero, so a bot takes the best suggested card rather than none, and the suggestion order (automated cards first) means the cards it can take are ones the engine plays for it. That is "picks the best matching card, or skips": it skips only when the effect itself offers nothing.
- **The MCP seat** reads the moves by label, as it reads every other prompt. No new tool.

### 5. What is stored, and privacy

- **A sideboard** (ADR 0191) is part of the deck's **source text** in `decks`, which account deletion already deletes (`accountdelete.Inventory`: `decks`, every deck the person saved). Its cards in a game are seat state, carried in restore points and replays exactly as a hand or library is. Its contents are hidden from every seat but the owner's until a card is revealed (ADR 0191 §7); this ADR does not change that.
- **A collection pick** stores nothing about a person. The suggestions are card names; a minted card is an ordinary card in the game's state. No new column, no new table, so no inventory change and no privacy-page change.
- The search route logs nothing beyond the ordinary request log.

### 6. Snapshots

Additive within v9: `seats[].hasSideboard` (omitempty) and `pendingChoices[].collectionSuggestions`. A file without `hasSideboard` restores it from the `sideboard` key's presence, which is exact for every file whose sideboard still held a card. The one transitional case is a game in flight across the deploy whose deck's sideboard had been emptied: it restores as a database pool. It affects only that game and only its wishes.

## The cards

Every card already on `TakeFromOutside` gains the database without a change: Burning Wish, Living Wish, Cunning Wish, Golden Wish, Death Wish, Glittering Wish, Fae of Wishes // Granted and Mastermind's Acquisition. Extrapolate the Impossible skips its "two different names in the sideboard" pre-check when the pool is the database; its prompt's set rule keeps the answer to exactly two.

New in this change, each Full:

- **Coax from the Blind Eternities:** "You may reveal an Eldrazi card you own from outside the game or choose a face-up Eldrazi card you own in exile. Put that card into your hand." The player is first offered their face-up Eldrazi cards in exile, with "none" meaning "look outside the game instead"; declining that leads to the outside pick. The two branches are the same choice asked in order.
- **Karn, the Great Creator:** the static (`OpponentsSourcesCantActivate` over artifacts), the +1 (until your next turn, up to one target noncreature artifact becomes an artifact creature with base power and toughness equal to its mana value) and the −2, which is Coax's shape for artifacts.

**Not in this change: Spawnsire of Ulamog's {20}.** "Cast any number of Eldrazi spells from among cards you own outside the game without paying their mana costs" needs a cast from the outside-the-game container: `castSourceZoneLocked`, the cast-permission lookup, the view's castable flags and the zone browser's cast button all assume a zone a `ZoneRef` reaches, and none of them reaches `ZoneOutside`. That is its own seam, filed as [#3955](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3955); Spawnsire keeps its caveat until it lands. Wish ("you may play a card you own from outside the game this turn") waits on the same seam.

## Compatibility

- **Wire:** additive. One new pending-choice kind, one new view field, one new answer key, one new read-only route. Older clients see an unknown kind; the server never sends one to a seat that is not its chooser with anything in it.
- **Behaviour:** a deck with no sideboard now brings cards in from the database where it brought nothing. A deck with a sideboard is unchanged.

## Validation

- **`internal/game`:** the pool choice (seated sideboard, none, an emptied sideboard, `ReplaceDeck`), the prompt's data, minting (owner, knowers, container, the flag untouched), the filter and set rule at submit, the bounds, decline, a closed pool at answer time, the search (chooser only, filtered), clone and the snapshot round trip.
- **`internal/deck`:** the collection's search, named lookup, collectibility and suggestion order.
- **`internal/protocol`:** the chooser's view and every other seat's.
- **`internal/legal`, `aiseat`:** every offered answer is accepted; a mandatory pick has one; the heuristic takes the best suggestion.
- **`internal/lobby`:** the route's gates (spectator, another seat, a closed pick) and its filtering.
- **Cards:** each wish in a deck with no sideboard, Extrapolate's pair, and the new cards.
- **Client:** the picker.
