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

**Not in this change: Spawnsire of Ulamog's {20}.** "Cast any number of Eldrazi spells from among cards you own outside the game without paying their mana costs" needs a cast from the outside-the-game container: `castSourceZoneLocked`, the cast-permission lookup, the view's castable flags and the zone browser's cast button all assume a zone a `ZoneRef` reaches, and none of them reaches `ZoneOutside`. That is its own seam, filed as [#3955](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3955); Spawnsire keeps its caveat until it lands. Wish ("you may play a card you own from outside the game this turn") waits on the same seam. *Built by the amendment of 2026-10-16 at the end of this ADR.*

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

---

## Amendment (2026-10-16, #3955): a card outside the game can be cast or played

**Status:** Accepted · 2026-10-16 · S76. **Issue:** [#3955](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3955), filed by the #3858 builder, assigned to the owner. **Numbering:** an amendment, so no number was taken. **Builds on:** [ADR 0066](0066-granted-cast-and-play-permissions.md) (a permission is a statement about an object), [ADR 0099](0099-discover.md) §4 and #3270 (a free cast during a resolution, "now or never"), [ADR 0191](0191-a-sideboard-is-the-outside-the-game-pool.md) (the container).

### Context

Both pools end with the chosen card in `Player.Sideboard`, a `ZoneOutside` container, and nothing could cast from it: `castZoneFromWire` had no value for it, `castSourceZoneLocked` no case, a grant is made through the card index (which deliberately does not cover the container, CR 400.11c), the view stamped cast surfaces for zones only, the sideboard tab was read only and the enumerator walked zones. Two printed cards waited: **Spawnsire of Ulamog**, "{20}: Cast any number of Eldrazi spells from among cards you own outside the game without paying their mana costs", and **Wish**, "You may play a card you own from outside the game this turn".

### Decision

**1. A cast from outside the game is a per-instance permission over the container, claimed with an ordinary `cast_spell`.** Nothing about the cast changes shape; the container becomes one more place a `CastPermission` can name.

- `cast_spell { from_zone: "outside" }` (`castZoneFromWire`). `castSourceZoneLocked` resolves it to the **caster's own** container and refuses it while another player makes their decisions (ADR 0167 §5). It answers "which pile", never "may you".
- **Only a permission opens it.** A card can declare nothing that does: `cardOpensZoneLocked` answers false for `ZoneOutside`, and `effects.Register` panics on a `CastableZones` that names it, as it does for exile. A card-level declaration would open a whole sideboard to every copy of the card, which is Wish's permission without Wish.
- **The grant doors** are `FreeCastFromOutsideForEffect` and `PlayFromOutsideForEffect` (`game/cast_outside.go`). They look in the holder's container, because `GrantCastPermissionOverCardForEffect` cannot (it finds the card by index), and refuse a card the holder does not own ("a card you own"). The permission is an ordinary `ScopeCards` permission with `Zone: ZoneOutside`; the epoch it names is the card's, as ever (CR 400.7), so a card that is cast and comes back is a new object.
- **The sweep** finds the named card in the container (`anyNamedObjectStillThereLocked` falls back to `outsideCardZoneLocked`), so a live grant is not dropped at the next cleanup as "the card is gone".

**2. The two grants.**

| | Spawnsire's cast | Wish's right |
| --- | --- | --- |
| scope | `ScopeCards`: one permission per picked card | `ScopeStanding`: one permission over the whole container, `CastsLeft: 1` |
| chosen | when the ability resolves (the printed text says "cast any number ... from among cards you own") | **when the card is played**, not when Wish resolves (section 6) |
| price | `Cost: "{0}"` (X locked at 0, additional costs still paid, CR 118.9 / 107.3b) | the card's own cost |
| timing | `TimingFlash` (CR 608.2g: the cast is part of the resolution) | `TimingNormal` |
| a land | refused (`CastOnly`: a land is played, not cast, CR 305.1) | played, and it is the turn's land drop (it spends the right) |
| window | pass-closed: `LapseOnPass: LapseStaysOutside` | until end of turn, spent by the first card played |
| not used | the card stays outside the game | the card stays outside the game |

`LapseStaysOutside` is `LapseStaysInHand`'s sibling: nothing moves when the window shuts, the grant is just gone. Nothing leaves the container except to the stack (or the battlefield, for a land), as a cast does (CR 601.2a).

**3. The pick (Spawnsire).** `effects.PlayFromOutside` asks the owner through `QueueOutsidePickForEffect`, so both pools apply unchanged: the sideboard as a `choose_cards`, the database as an `outside_collection` whose answer mints the named cards. `OutsidePickPrompt` gained `Destination` (default `outside`); the two new destinations are `cast_free_outside` and `play_outside`. Spawnsire's pick is "any number": `Min 0`, `Max game.OutsideAnyNumber` (100), a bound on the size of an answer, not on the game. Each picked card carries its **own** grant, so casting one leaves the others open (a grant naming all of them would be spent by the first cast), and the holder's next priority pass ends every one together. This is `grantFreeHandCastsFrom`'s shape for the container, and it is the same sandbox simplification every "you may cast it" a resolution offers takes (ADR 0066, ADR 0091): the casts are ordinary `cast_spell`s taken after the pick rather than a frame inside the resolution.

**4. A land played from outside the game keeps its entry replacements.** The entry pipeline (`ReplacementEvent`, the self-replacements, a shockland's payment, Kismet) reads the entering land through `LookupCardForEffect`, which cannot see the container. `Game.outsideEntering` opens one window, for the land being played, from the replacement pass until it lands (or is refused, or its paused entry is answered); the card index consults it only when the lookup would otherwise miss (`locateCardLocked`). Cloned for undo, not snapshotted (`enteringTokens`' posture). Temple of Abandon enters tapped as a replacement (no tap event); a shockland asks and resumes.

**5. The view, the enumerator, the bot and the client.**

- The view stamps a card in the **owner's** sideboard with the same cast surface a graveyard or exile card carries (`castable_here`, `alternative_costs`, `cast_prices`, `legal_targets`, `clauses`, `exile_play`) through the one `castStampsFor`, for a card a live permission opens (`stampOutsideCasts`). It reaches the owner alone, because the sideboard does. `liveCardForView` and `castSourceOf` fall back to the read-only `LookupOutsideCardForEffect`, which exists for the surfaces that SHOW or CHECK a card and which no effect may use. The auto-tap preview reads the caller's own container card the same way and answers 404 for anyone else's.
- The enumerator walks the seat's own container (only when something is granted) and offers the cast at every price the engine accepts, labelled "... from outside the game". A collection pick that says "any number" also offers the whole suggestion list as one answer, which a pair-capped expansion would never reach.
- The heuristic prices a pick over its own sideboard or the database as the take it is for all three destinations, and finds a card it owns outside the game when it prices a cast of it. A free cast from outside is not charged a card in hand.
- The client's sideboard tab shows a button on each card the server marks `castable_here` ("cast free", "play", "cast") that hands the card to the one cast chain with `from_zone: "outside"`. Answering a `cast_free_outside` or `play_outside` pick waits for the frame carrying the grants, then goes straight into the cast chain when exactly one card is castable and opens the tab otherwise, so the player never has to find it. An "any number" database pick may take the same name more than once (a player may own two copies of an Eldrazi): the picker adds a copy per click and a row's "-" takes one back (`allowsRepeatedNames`: a ceiling above two; a pair keeps its toggle, since no set rule asks for more than two names).

**6. Wish names its card when it is played, not when it resolves.** (This replaces the first cut of this amendment, which named the card at resolution in both pools. That was weaker than printed: a Wish followed by a draw could not use what it drew, so grading Wish Full would have been a false grade.) "You may play a card you own from outside the game this turn" is a permission that names no card. `GrantOutsidePlayRight` stores `OutsidePlayRightPermission` when Wish resolves: `ScopeStanding`, `CastsLeft: 1`, until end of turn, the card's own timing and cost. A standing permission covers a card that does not exist yet, which is the point.

- **A sideboard pool** needs nothing more. Every card in the container is castable or playable under the right, so the sideboard tab shows the button on all of them, and the first card played, a land included, spends it (`consumeLimitedGrantLocked`; the land branch of `castSpellLocked` now spends a one-use permission as a spell cast does).
- **The database pool** names its card at that moment with the `outside_play` action (`actions.TypeOutsidePlay`, no params, `Game.NameOutsideCardToPlay`). It opens the ordinary `outside_collection` pick (any card in the database, `Min 0`, `Max 1`, destination `play_outside`); answering mints the card into the container, where the standing right already covers it, and the client starts the cast or the land play at once (the one granted card goes straight into the cast chain). The action is taken while holding priority, at any time the right is live: the card is then cast at whatever timing *it* allows, so a sorcery named on another player's turn is minted and waits for its window (it can be cast from the tab later this turn).
- **Cancelling the pick spends nothing.** The right is spent by a card played, not by a card named; a declined pick mints nothing, and a named card that is never played stays outside the game with the right still live until end of turn.
- **A sideboard pool is never asked to name a card** (`CanNameOutsideCardLocked` is false when `HasSideboard`; the action is refused with `ErrNoPlayPermission`), because naming a database card would add a card the deck does not own.
- **Wire:** `PlayerView.outside_play { source_name, can_name }`, owner only, present while the right is live; `can_name` is true for the database pool. The enumerator offers the action (`Kind: special_action`, `type: outside_play`) to a database-pool seat holding the right whose container is empty. The engine accepts it whenever the right is live; the empty container is bot policy, so an automated seat that named a sorcery it cannot cast yet does not name a fresh card at every priority. A human has the button regardless (sideboard tab and pile).
- The per-card `PlayFromOutsideForEffect` grant of the first cut is gone; nothing plays one card by name any more.

### Not changed

- The pool and its rules (sections 1 to 6 of the ADR above), the controlled-turn rule and the minting. A card the effect leaves outside stays in the container.
- No snapshot field: a `CastPermission` already persists `zone` and `lapseOnPass` as strings, and a permission naming the container restores like any other. `SnapshotSchemaVersion` is unchanged.

### Validation

`game` (cast, land play, lapse, end-of-turn sweep, independence of the grants, ownership, controlled-turn, no card-level opening), `protocol` (the owner's stamps and nobody else's), `legal` (every move offered is accepted, a play pays, a cast-only grant offers no land, "any number"), `aiseat/heuristic`, `lobby` (the preview), `cards/effects` (Wish and Spawnsire over both pools, Wish choosing at play time, the right spent once, a cancelled naming spending nothing, a Temple entering tapped, a shockland pausing and resuming) and the client (the tab's button, the Wish button and pile, the any-number picker with repeated names, the request that opens the cast).

## Amendment 2026-10-23: the outside-the-game picks are keyed, so a table waiting on one is a restore point (#4256)

### Context

Both branches of `QueueOutsidePickForEffect` resumed through a closure: the sideboard branch's `choose_cards` frame held the continuation, and the database branch's `searchResumeFrame` held the continuation and the filter the typed names are checked against (decision 2 above). A prompt with a frame is not written to a restore point, so a table waiting on any wish, on learn's Lesson, on Research or on Spawnsire of Ulamog's {20} could be undone but not saved and restored. #4246 (ADR 0099, amendment 2026-10-23) did this for the free-cast offers; the outside picks needed a keyed filter as well, which is why they were a separate seam.

### Decision

1. **One key names both halves.** `game.RegisterOutsidePick(key, filter, then)` (`game/outside_pick_keyed.go`) registers the continuation as a keyed choose-cards continuation (ledgered `choose <key>`) and the filter beside it under the same key. The filter is `OutsideMatchFunc`, a function of the card and one plain word, so it captures nothing. `OutsidePickPrompt.Keyed` queues the keyed form; with it, `Match`, `Validate` and `Then` must be nil (a keyed prompt that sets one is refused with an effect error, because a restore would drop it).
2. **The sideboard branch** reads the filter once, to list the candidates, and queues an ordinary keyed `choose_cards` pick (`choose_cards_keyed.go`). The continuation re-checks that its picks are still in the container (`StillOutsideForEffect`), the live-zone re-check a keyed pick does not have.
3. **The database branch** carries the key (`pickThen`), the word (`pickMatchArg`, new and additive in v9) and the carried IDs (`optionCarry`), and no frame. The answer, the human picker's search and the enumerator's legality check all find the filter through `collectionFilterLocked`: the frame's for the closure form, the registered one for the keyed form. The minted cards' IDs go to the continuation as a sideboard pick's are, through `runKeyedChooseCardsLocked`. `CardsChosen.Question` (the prompt's header) is new so a reveal can name its reason.
4. **The word is data.** A filter that reads a word of the card's text a text-changing effect may have changed (CR 612.2) takes it as `MatchArg`: Spawnsire of Ulamog's and Coax from the Blind Eternities' "Eldrazi". The restored table's filter reads the same word.
5. **The callers.** Every wish (Burning, Cunning, Glittering, Golden, Living and Death Wish, Granted, Mastermind's Acquisition), learn's Lesson half, Research, Spawnsire of Ulamog and the outside half of Karn, the Great Creator and Coax from the Blind Eternities: 14 keys, each registered in the card's file (`registerOutsideTake`, `registerOutsideTakeOfWord`, `registerOutsideShuffle`, `registerPlayFromOutside`). The cast-from-outside name is bound to the key, not read off the source when the answer comes in, as the free-cast offers' is.

### What is not keyed

Extrapolate the Impossible's pick has a set rule ("two cards with different names", `Validate`), which is a closure, and its continuation captures the stack item; it stays on the closure form. Karn's and Coax's exile pick, before the outside pick, stays a closure: its "none" asks the outside pick with the card's word, which a keyed choose-cards pick has nowhere to carry. Wish's `outside_play` action is not a prompt and is unchanged.

### Compatibility, closure counts and tests

A binary from before this change does not have these keys and refuses a file with one of the picks open (`checkEffectKeys`: `outside pick <key>` for the database branch, `choose-cards continuation <key>` for the sideboard), rather than restoring a question whose answer would run nothing. The one new snapshot field is `pendingChoices[].pickMatchArg`, additive within v9 (`-update-shape`); no fixture is edited. No field reaches a func from `Game`, so the closure counts do not move. Tests (`effects/keyed_prompts_restore_test.go`): learn's Lesson from a sideboard and from the database across a strict restore, Burning Wish (and its next step) across one, Spawnsire's database pick keeping its word across one, and an unknown key refused on each branch. **Back-out:** sending the keyed sideboard pick back to the closure prompt fails `TestLearnLessonPickIsARestorePoint`, `TestBurningWishPickIsARestorePoint` and `TestOutsidePickWithAnUnknownKeyIsRefused/sideboard`; dropping `PickMatchArg` from the database prompt fails `TestSpawnsireDatabasePickKeepsItsWordAcrossARestore`.

## Amendment 2026-10-24 (#4264): a keyed pick carries a word and a set rule; Karn's and Coax's exile half and Extrapolate the Impossible are keyed

The amendment above left three prompts on the closure form: Karn, the Great Creator's and Coax from the Blind Eternities' exile half (its "none" asks the outside pick with the spell's word, which a keyed choose-cards pick had nowhere to carry) and Extrapolate the Impossible's first pick (a set rule over the whole answer).

**A string carry.** `ChooseCardsPrompt.CarryWord` rides the prompt as `PendingChoice.PickMatchArg`, the field the database branch already writes for its filter's word, and the continuation reads it as `CardsChosen.Word`. No snapshot field is new. Karn's and Coax's exile pick is `registerExileOrOutside(key, outside, question)`, a keyed `choose_cards` over the face-up cards the player owns in exile whose continuation puts the chosen card into the hand (re-checking that it is still a face-up card they own in exile, since a keyed pick has no live-zone re-check), or on "none" asks the card's outside pick with the word. `TakeFromExileOrOutside{Exile}` names it.

**A set rule.** `game.RegisterChooseCardsThenRule(key, rule, fn)` registers the rule with the continuation, and `RegisterOutsidePickRule(key, filter, rule, then)` with an outside pick's filter, all under one key, so no ledger line is new. `checkChooseCardsPicksLocked` takes the rule from the closure frame or, for a keyed pick, from the registry, reading a card that no zone holds from the sideboards (the pick is over the outside-the-game container); `collectionFilterLocked` does the same for the database branch, so the typed names are judged by the rule after a restore. Extrapolate the Impossible's first pick is `extrapolatePick`: no filter, the rule "exactly two cards with different names", and the reveal that follows. The two prompts after it (the opponent choice and the reveal pick) are still closures.

**Compatibility.** An older binary refuses a file with one of these prompts open (`choose-cards continuation <key>`, `outside pick <key>`). No snapshot field and no closure route is added. **Tests** (`effects/keyed_prompts_restore2_test.go`, `game/keyed_confirm_rule_test.go`): Karn's exile pick across a strict restore, taking the card and, on "none", asking the database for artifacts; Coax's exile pick keeping its word and asking the database for Eldrazi; Extrapolate's pick from a sideboard and from the database refusing one card and a pair of copies after the restore, with the enumerator's window agreeing; a game-level pick keeping its rule and its word; an unknown key refused on each. **Back-out:** dropping `PickMatchArg` from the keyed choose-cards prompt fails `TestCoaxExilePickKeepsItsWordAcrossARestore`; sending the exile pick back to the closure form fails `TestKarnExilePickIsARestorePoint`; removing the keyed lookup from `collectionFilterLocked` fails `TestExtrapolateTheImpossibleDatabasePickIsARestorePoint`.

## Amendment 2026-10-24 (second, #4264): Extrapolate the Impossible's opponent and reveal picks are keyed

The amendment above keyed the first pick and left the second and third prompts as closures: the caster naming the opponent (`ChoosePlayer`) and that opponent choosing one of the revealed pair (`RevealPick`).

**Decision.** The opponent question is the engine's existing keyed option pick (`QueueChoosePlayerForEffect{ThenKey}`, `effects.OptionPickThen`) carrying the caster and the revealed pair; its answer queues a keyed reveal pick. `RevealPickPrompt{ThenKey, Carry}` (`resolution_pick.go`) queues a `reveal_pick` with no frame and no run: the continuation is a registered `choose_cards` continuation (`RegisterChooseCardsThen`), `isKeyedChooseCards` covers the kind, `checkEffectKeys` checks the key against the same registry, the answer logs the pick as the closure leg does, and `CardsChosen.From` gives the continuation the owner of the pool. Extrapolate's keys are `option-pick/extrapolate-the-impossible-opponent` and `choose-cards/extrapolate-the-impossible-take`. A chooser who has left queues nothing, and the card stays outside the game (CR 800.4a).

**Test** (`effects/keyed_asked_item_restore_test.go`): `TestExtrapolateTheImpossibleOpponentAndRevealPicksAreRestorePoints` takes a strict restore at the second prompt and again at the third, and the chosen card comes into the caster's hand on the restored table; an unknown key at either is refused. **Back-out:** removing the reveal pick from `isKeyedChooseCards` fails it.
