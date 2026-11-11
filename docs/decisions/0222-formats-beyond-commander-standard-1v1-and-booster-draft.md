# ADR 0222 — Formats beyond Commander: Standard 1v1 and booster draft

**Status:** Accepted · 2026-11-10 (Ian read the summary of this ADR and its calls and said to go ahead) · [S81](../sprints.md)
**Amended 2026-11-11:** packs follow the draft product of the set's era (owner correction). §4, [#5437](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5437).
**Issue:** [#5431](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5431) (the formats epic)
**Owner decisions:** Ian, 2026-11-10, recorded under [Owner decisions](#owner-decisions). They are **decided**. Everything under [Calls made here](#calls-made-here) is this ADR's and can be changed in review.
**Numbering:** checked on 2026-11-10 after `git fetch --all --prune`, over `docs/decisions/` on all 129 remote heads: 0221 is the highest number anywhere. This is **0222**.
**Builds on:**
- [PLAN.md](../../PLAN.md) §3, which this ADR's PR amends: other formats move from "out of scope" to a secondary area.
- [#2352](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2352): Luke's plan for a format on the table and best-of-one Standard. It becomes slice 1 here, unchanged in substance.
- [ADR 0191](0191-a-sideboard-is-the-outside-the-game-pool.md): a sideboard already rides into the game as the outside-the-game pool.
- [ADR 0033](0033-ai-bot-seat.md) and [ADR 0126](0126-bots-that-play-their-decks-the-heuristic-prices-what-a-card-does.md): the bot seat, and the Purpose declarations a draft bot can price a card with.
- [ADR 0148](0148-capacity-limits-for-a-public-launch.md) and [ADR 0149](0149-privacy-notice-and-account-deletion.md): the capacity caps and the rules for storing anything about a person.
- Discussion [#4472](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/4472), which set out the options.

---

## Context

The charter kept C2 to Commander. Ian wants draft and two-player constructed as well, and on 2026-11-10 decided to add them as a secondary area.

What changed since #4472 was written is the catalog. It measured Standard at 28% and said a set draft "needs a whole set implemented". On 2026-11-10 the 5,163 Standard-legal cards (unique, basics excluded) were 99% built: 4,629 on `develop` and the rest in two roll-ups that merged the same day. A set can now be drafted, and real Standard decks played, on cards alone.

What the code already gives us (#2352 lists it with file names): two-player tables (`MinPlayers = 2`), `TableSettings.StartingLife`, Commander rules that do nothing when a deck has no commander, Scryfall's `legalities` map, and sideboards parsed and carried into the game.

What it lacks: any notion of a format, a deck check other than Commander's, a match of more than one game, packs, a draft, and a bot that picks cards or builds a deck.

One naming note: "pod" is taken. [ADR 0220](0220-pods-groups-of-players-for-shared-stats.md) uses it for a group of players who share stats. The eight seats of a draft are called a **draft** here, never a pod.

## Owner decisions

1. **Commander stays primary.** Other formats are a secondary area. The charter says so.
2. **Constructed 1v1 starts with best-of-one Standard**, because the cards are there. Best-of-three with sideboards follows.
3. **Draft means drafting individual sets**, with retail-style packs, as well as cube. Cube can wait unless it turns out to be cheap.
4. **Two ways to pick:** pick one and pass, and pick two and pass.
5. **Drafted decks play one against one.**
6. **Bots fill the draft.** A draft can start with one person and bots in every other seat, so someone can learn to draft and build alone. The bots should try to build a real deck.
7. **Custom packs come later** (powered packs for a cube and the like). Retail packs for now.

## Decision

### 1. A format on the table

- `TableSettings.Format` names the format: `commander` (the default, and what every existing table and snapshot reads as when the field is absent), `standard`, `limited`. It is chosen when the table is made and fixed once the game starts. The field is additive in the snapshot.
- A format is data: starting life, seat cap, deck rules, and whether a commander is expected. Pioneer, Modern and Pauper later are new rows, not new code.
- `MaxPlayers` stays the global bound. A format's seat cap is at most that.
- **The starting player of a two-player game skips their first draw** (CR 103.8a). This is a rule of two-player games, not of a format, so it keys on the number of players. A three- or four-player game is unchanged (CR 103.8c). #2352 cites this as 103.7a; in the pinned edition it is 103.8a.

### 2. Deck checks

- `deck.Validate` takes the format.
- **Commander:** unchanged.
- **Standard:** at least 60 cards and at most four of any card but basic lands (CR 100.2a), every card `legalities.standard == legal`, no commander, a sideboard of at most 15 with the four-card limit counted across deck and sideboard (CR 100.4a).
- **Limited:** at least 40 cards, built only from the player's drafted pool plus any number of basic lands (CR 100.2b, 905.1d). The rest of the pool is the sideboard (CR 100.4b).
- A saved deck records the format it was built for. Existing decks are Commander decks.
- **A deck check also says what the engine can't play.** A constructed table accepts a deck with uncatalogued cards the way a Commander table does today (the card plays, and resolves by hand), and the coverage check warns. Limited never has the problem (section 4).

### 3. Matches

- A **match** is one or three games between the same two players at the same table. Best-of-one is a match of one game, so every 1v1 table has a match and there is one code path.
- Between games the table returns to a short sideboarding screen: each player moves cards between deck and sideboard, the deck check runs again, and both ready up. A timer ends it.
- The loser of the previous game chooses who plays first (CR 103.1).
- The match score is kept on the table and shown on the board. A match counts as one finished table for the charter's "games reach an end" goal; the stats record keeps each game.

### 4. Packs

- A pack is generated from Scryfall data for one set: the cards the dump marks as found in boosters, in slots by rarity, to the draftable booster of that set's own era, chosen from the set's release date: a **play booster** (14 cards) for a set released on or after 2024-02-09 (Murders at Karlov Manor); a **draft booster** (15 cards: 10 commons, 3 uncommons, 1 rare that is a mythic rare about 1 time in 8, 1 basic land) for a set from 2008-10-03 (Shards of Alara) up to that date; and a **draft booster from before mythic rares** (15 cards: 11 commons, 3 uncommons, 1 rare) for an older set. These dates and rates were written from memory and are to be checked against a source; a set whose booster differed is corrected per set, as data. Set boosters and collector boosters were sold beside draft boosters and were never draft products, so they are not modelled. Where a set's exact sheet isn't public, the pack follows the published slot rates and says so on the set's draft page. **Packs are approximations of retail packs, and the page says that.**
- **A set is draftable only when every card its packs can contain is in the catalog as Full.** The check is computed from the catalog, never a hand-kept list, and a test fails if a draftable set loses a card. A draft never hands someone a card that doesn't work.
- Foil, showcase and other treatments are not modelled. One printing of each card.
- Nothing about a pack is scarce or owned. A pack exists for the length of a draft and is thrown away.

### 5. The draft

- A **draft** is its own object in the lobby, not a game: 2 to 8 seats, a set (or later a card list), and a pick mode.
- **Pick one:** three packs each. Packs pass left, right, left (CR 905.1a–b).
- **Pick two:** each pick takes two cards. It is offered at 4 seats, the size it was designed for, and allowed at any size.
- Any seat not taken by a person when the host starts is a bot. One person and seven bots is a normal draft.
- A pick timer keeps a draft moving. When it runs out, the seat takes the bot's pick for that pack. A person who disconnects is picked for the same way and can come back.
- A player sees their current pack and their own picks, nothing else (CR 905.1c). Picks are hidden information: the server never sends another seat's picks or pack, and the draft has a per-seat view like the game's.
- After the last pick each person builds a 40-card deck from their pool with free basic lands, on a timer, with a land-count suggestion.

### 6. Playing the drafted decks

- Drafted decks play 1v1 matches at `limited` tables the draft opens.
- People in the draft are paired with each other first. With an odd number, or with one person, the others are bots: each bot plays the deck it drafted.
- First version: **three rounds, paired at random without repeats**, best-of-one by default and best-of-three as a setting. No standings beyond the draft's own win-loss table. No prizes, ever.
- A person may stop after any round. The draft ends when its rounds do or when every person has left.

### 7. Draft bots

There are two jobs, and they differ in cost.

- **Picking.** The bot rates each card in the pack and takes the best, then leans toward the colours it has taken. The rating is computed from what the catalog already knows: mana value, type, power and toughness, keywords, rarity, and the Purpose a card declares (removal, card draw, a sweep) from ADR 0126. After a few picks it commits to two colours and weighs the curve and creature count it still needs. This is a contained piece of work with a clear measure, and it gives a bot that drafts a coherent two-colour deck.
- **Building.** From its pool the bot takes its two best colours, the best 22 or 23 cards by the same rating with a curve check, and 17 or 18 lands split by colour weight. This is small.
- **Playing** needs nothing new: the heuristic bot already plays whatever deck it holds.
- **What this does not give** is a bot that drafts like a strong player: reading which colours are open from what gets passed, or knowing that one common is far better than its stats suggest. That needs per-card ratings from real draft results for each set. It is a later, optional layer over the same seam (a ratings file per set), and whether outside data may be used is an owner and licence question, not decided here.
- The measure: a bot's drafted deck is two colours with a sane curve in at least 19 of 20 seeded drafts, and beats a deck of random picks from the same packs well over half the time in the arena.

### 8. Cube and custom packs

- A cube draft is section 5 with a different pack source: a saved card list, shuffled into packs of 15. Once the draft exists this is cheap, so it is the last slice, not a separate project.
- A cube list is checked like a deck: every card must be Full, and the check names the ones that aren't.
- Custom pack recipes (powered packs and so on) are a later ADR.

### 9. Limits, storage and moderation

- A draft counts against the table cap of ADR 0148 as one table while people are picking, and each match it opens counts as a table. Bot-versus-bot matches inside a draft are not played out at a table; they are resolved by the arena runner or skipped (see Calls made here).
- **What is stored about a person:** a finished draft's set, date, the person's picks and their match results, so they can look back at a draft. This is new personal data. The slice that stores it updates `accountdelete.Inventory` and `Delete`, the privacy page and `PRIVACY_SCHEMA_VERSION` in the same PR (ADR 0149). A guest's draft is kept only while it runs.
- Draft and table names go through the name filter. Reports work from a match table as they do today.
- A visitor can draft against bots with no sign-in, as they can play a practice game today. No card data sits behind a gate.

## Slices

Each is one issue and one or more PRs. 1 and 3 can start at once; the rest are in order.

1. **A format on the table, and best-of-one Standard** (#2352): sections 1 and 2 (Commander and Standard), the two-player first draw, the client's format picker, hiding commander UI when no seat has a commander, and Standard bot decks.
2. **Matches:** section 3, best-of-three with sideboarding.
3. **Packs:** section 4, with the draftable-set check and a page listing which sets can be drafted.
4. **The draft:** section 5, pick one and pick two, with a random-pick placeholder bot.
5. **Draft bots:** section 7.
6. **Limited decks and rounds:** the limited deck check, the build screen, and section 6.
7. **Cube:** section 8.

Slices 1 and 2 together are roughly a sprint. Slices 3 to 6 are roughly two. They are S81's checklist: #2352, #5436, #5437, #5438, #5439, #5440 and #5441, in slice order.

## Calls made here

Open to change in review.

- **The format lives on `TableSettings`**, not `GameMeta`: it already carries starting life through snapshots, saved setups and the create flow.
- **Best-of-one is a match of one**, so there is one path for 1v1.
- **A set is draftable only at 100% Full** of its pack contents. The alternative, drafting a set with a few manual cards, was rejected: a new drafter can't be expected to resolve a card by hand.
- **Pick two is allowed at any seat count**, offered by default at 4.
- **Three random rounds**, not Swiss or single elimination. With bots filling seats, standings mean little.
- **Bot-versus-bot matches are not shown.** Their result is decided by the arena runner when capacity allows and by a coin weighted on deck rating when it doesn't. A person never waits on one.
- **A draft keeps a person's picks** after it ends. If the owner would rather store nothing, the look-back page goes and section 9's privacy work with it.
- **Outside draft data for bot ratings is not used** until the owner decides it may be.

### Matches as built (slice 2, #5436)

Section 3 says what a match is. These are the calls its build made, each open to change in review like the ones above.

- **Where the match lives.** `Game.Match`, on the engine, for every table of a format whose row says `Matches` (Standard). A Commander table has none, and nothing of the match code runs for it. The match length is a table setting, `best_of` (1 or 3), fixed at Start.
- **The match is the table.** One `games` row, one room, one invite and one finished table. The match id is the table's id, and a game of it is that id plus its number. The lobby sees the table end once, when the match does, so the "games reach an end" goal and each seat's result count a match once, with the match's winner. `match_games` (migration 0037) keeps each game: who won it by seat, why it ended, who went first and how many turns it took. It is read from the final view, so a replay backfill writes it too.
- **Between games is a state of the engine.** `StateBetweenGames`: the finished game stays on the board, every door that asks for an active game refuses, and the sideboarding window is open. `CurrentState` and the wire report it as `active`, because to the lobby, the bot runners and the table list the table is still in play; `match.sideboarding` on the view says where it is. The alternative, rebuilding the board the moment a game ends, would have hidden the final board and run inside whatever ended the game.
- **The next game begins the way a restart does** ([ADR 0201](0201-restarting-the-game.md) §2): a fresh `NewGame` adopted in place, so every game-scoped field starts clean. Each seat's deck is seated from a pool of pristine cards kept from registration, with new instance ids, so nothing a game did to a card reaches the next. The pool is kept as the snapshot's own card records, so the match adds no route from `Game` to a closure and the ADR 0041 ratchet does not move.
- **One window, one timer, three minutes.** Sideboarding and the play-or-draw choice share it. The choice is the loser's (CR 103.1), defaults to playing first, and is shown to the chooser alone until the game starts, so the opponent sideboards without knowing it. After a drawn game the player who chose for that game chooses again, as CR 103.1 says.
- **Drawn games.** A drawn game counts for nobody. The match is won by the first player to win two games, and after three games by the player with more wins, or is drawn. This is the reading of "one or three games"; a tournament would play on until someone had two wins.
- **The deck check between games** is the format's `DeckRules` row: the deck's size, the sideboard's size and the four-card limit counted across both (CR 100.2a, 100.4a). Card legality is not asked again, because no card joins or leaves a pool. A deck that fails is refused and the last legal one stands.
- **A companion is revealed only by the deck it was registered with.** Its deck-building condition was checked against that list. A deck changed between games keeps the card as a plain sideboard card: weaker than the rules allow, never stronger.
- **A deck returns to its registered list when the match ends** by never having left it anywhere else: sideboarding changes the match's pool, not the saved deck and not the list a rematch seats.
- **Conceding.** `concede` concedes the game in play. `concede_match` concedes the match, during a game or between two. A seat an admin removes, or one the table moves on without ([ADR 0177](0177-away-players-and-moving-the-table-on.md)), concedes the match: that player is not coming back for the next game.
- **Undo stops at a game's end.** The room empties its undo history when a game of a match ends and when the next begins, so nobody undoes back into sideboarding having seen an opening hand. This also means the action that ended a game of a match cannot be taken back, by anyone.
- **Bots** do not sideboard and are always ready, and a bot that chooses plays first. An agent seat is treated the same, because its client has no sideboarding verb yet.
- **A restart during sideboarding** restores into the window and gives the players at least a minute to come back.

Not built here: per-game detail in the statistics record (the casts, damage and turn tables describe the match's last game, because the log begins again with each game; a per-game record needs a game number on seven tables and a `gamestats.Version` bump), a sideboarding verb for the MCP seat, and bot sideboarding.

### The draft as built (slice 4, #5438)

Section 5 says what a draft is. These are the calls its build made, each open to change in review like the ones above.

- **Where the draft lives.** `server/internal/draft` is the draft and its rules: it takes the time as an argument, holds no lock and starts no goroutine, so a seeded draft replays in a test. Its `Room` adds the lock, the pick timer and the sockets. The lobby keeps live drafts in a map of its own (`lobby/drafts.go`), beside its tables and not among them: a draft has no `games` row, no room in the room manager, no snapshot and no engine. It is in memory only, and a restart or a deploy ends every draft. **Nothing about a person is stored**; section 9's privacy work belongs to the slice that first persists a finished draft (slice 6).
- **Everyone picks, then every pack passes.** Each seat picks from the pack it holds, and the packs pass together once the last seat has picked. The alternative, a queue of packs in front of each seat so a fast drafter never waits, was left out: one shared step is what makes "who is still picking" and one timer mean something.
- **Seats and direction.** The player to a seat's left is the next higher seat number, wrapping round, so packs go up the seat numbers in rounds one and three and down in round two (CR 905.1b). People sit in the order they joined and bots fill the rest; seats are not shuffled.
- **Pick two** opens three packs a seat, as pick one does. A pack with an odd number of cards ends on a pick of one.
- **The seat token is a session of a new role, `drafter`, and is not the cookie.** It is minted by the server's one authenticator and returned in the body; the client keeps it in `localStorage` for that draft. So a signed-in person stays signed in while they draft, a visitor needs no cookie, and coming back to a draft is opening its page. Because page script holds it, it is worth nothing elsewhere: `auth.Middleware` refuses the role on every route that does not name it, and the game socket refuses it. It carries the user of the sign-in it was minted from, so revoking that sign-in revokes the seat.
- **A visitor's draft is solo.** Someone with no sign-in gets a draft with no invite and bots in every other seat, the way a visitor gets a practice table and not a table of their own. A draft other people join is made by someone signed in, who may hold one unfinished draft at a time. A guest joins such a draft by its invite link and a name, as a guest joins a table.
- **The pick timer** is one length for every pick, 60 seconds by default and 15 to 300 by the create request. A timer that shortens as the pack does is left for later.
- **A seat that is away is picked for ten seconds into the pick**, not at the end of the timer. Section 5 says a person who disconnects is picked for the same way; they are, by the same placeholder, only sooner, so one closed laptop does not cost seven people a minute on every pick. A seat that comes back has the whole timer again, and finds its seat, its picks and the pack it holds.
- **Leaving.** In the lobby a seat is freed and the next person hosts. Once picking has begun the seat becomes a bot for the rest of the draft and its token is spent.
- **Bots pick at once**, behind `draft.Picker`: given the pack, the seat's picks so far and the pick mode, return the pick. The placeholder picks at random, as a function of the draft's seed and where the pick falls, so a seeded draft replays whatever the timing of its seats. Slice 5 implements that interface and nothing else.
- **Hidden information is in the shape of the view, not in a filter.** `draft.View` holds cards in three fields, all the viewer's own; every other seat is a `PlayerView`, which has no card field. There is no unfiltered view to redact. One test walks every view of a whole draft, another every byte sent to three people through a whole draft over real sockets, and each fails on a card the seat never held or drafted.
- **The load cap.** A draft counts as one table plus a human seat for each person, from its creation until its last pick. Its bots count nothing: a bot seat's weight stands for a bot playing a game, and a draft bot picks a card. A finished draft counts nothing, like an ended table.
- **Idle drafts** close on the tables' sweep and by its three settings (section 9, ADR 0211): a lobby after the lobby window, a draft that is picking or finished after the bots window with nobody connected and after the game window whoever is. A visitor's draft still in its lobby goes after 15 minutes with its page closed: it costs nothing to make and holds a table's weight.
- **Draft card images have their own public route**, `GET /draft/image/{id}`. `/cards/{id}/image` needs a session cookie, which a visitor's draft does not have, and `/catalog/image` serves one representative printing of a card where a pack holds the set's own. It is scoped as `/catalog/image` is: only the printings a draftable set's packs can contain.
- **Every draft route lives under `/draft`**, beside `GET /draft/sets`, the path the proxy already sends to the server.

Not built here: the admin views do not list drafts or their sockets; the Discord bot has no draft command; there is no report or remove-a-seat on a draft (a draft has no chat, and shows other people as a name that has passed the name filter); a draft cannot be watched by a spectator; and a deploy ends a draft in progress with no warning to its seats.

### Limited decks and rounds as built (slice 6, #5440)

Sections 2 (limited), 6 and 9 say what follows a draft's picks. These are the calls its build made, each open to change in review like the ones above.

- **`limited` is a row, and a draft-only one.** 20 life, two seats, no commander, matches. Nobody has a limited deck without a card pool to check it against, so the row is `DraftOnly`: `POST /games` and a settings patch refuse it, and a table gets it only through `Game.OpenLimited`, which a draft's match calls. The deck check is `deck.ValidateLimited(deck, pool)`, not `deck.Validate`, because a list alone does not carry its pool.
- **The sideboard has no size limit at a limited table** (CR 100.4b: it is the rest of the pool). The engine's cap of fifteen (ADR 0191 §5) and the between-games check now read the format's row; every other format keeps the cap.
- **The draft's `state` stays `done` after the last pick**, which is what it says, and a new `phase` (`building`, `rounds`, `finished`) says what follows. The alternative, new states between `picking` and `done`, would have changed what `done` means for the draft bots being built beside this slice.
- **The build.** A person submits a deck by naming their pool cards and a count of each basic land; they may submit again until the rounds begin, which is when the last person has submitted or the timer (15 minutes by default, 2 to 30) runs out. A person who has not submitted then gets a deck from the `DeckBuilder`, as every bot seat does when the picks end. `DeckBuilder` is the seam for slice 5's bot; its default is deliberately plain (the pool's two most common colours, its 23 cheapest playable cards, 17 basic lands split by colour), and a builder whose deck fails the limited check is overruled by it.
- **The land-count suggestion is the client's**, computed from the deck being built: 17 lands for 23 other cards, in proportion otherwise, split by the colours of the deck's mana costs.
- **Pairing.** Each round is chosen at random, from the draft's seed, among the ways to pair the seats that are best by these preferences in order: no seat meets a seat it has played; people meet people; with an odd number of seats a bot sits the round out; nobody sits out twice. **A draft of two seats has one possible pairing, so its rounds are rematches**: the alternative was a draft of two playing one round.
- **A sit-out is a bye and counts for nothing** in the win-loss table.
- **Bot-versus-bot matches are a seeded coin toss**, even, decided as the round begins. Section 9 and the call above say the arena runner when capacity allows and a coin weighted on deck rating when it doesn't; neither a deck rating nor a hosted arena runner exists yet (slice 5 brings the rating), so this is the plain coin, and the results are marked as tossed for in the draft's table. **This is one of the ADR's calls built smaller than written.**
- **A match's table is made the way the practice table is**: started, with both seats at it, no invite, not listed, never written to the database, and gone with its draft. So a draft's matches are not in "My games", write no statistics record and do not count toward the "games reach an end" goal. What is kept of a match is its result, in the draft. The alternative, an ordinary stored table, would have put a half-anonymous two-seat table with no way to rejoin it in every drafter's history.
- **Sitting down.** `POST /draft/{id}/match` mints the table seat's session from the draft seat's and sets it as the cookie, as a join does. The draft seat token is not spent, so the way back to the draft is its page.
- **A match that stalls** (no commit for 20 minutes, or three hours old) is recorded as a draw and its table closed, so one person walking away cannot hold a round open. When it was the draft's only match in play and nobody is on the draft's page either, everyone has walked away, and the draft is closed instead of seating a bot to wait for them through two more rounds. There is no away-player offer (ADR 0177) at a match table. A person who leaves the draft in the middle of a match concedes it.
- **The load cap.** Through the build the draft is one table and its people, as while picking. In the rounds it counts nothing itself and each open match table counts as a table with its seats. A round's tables are opened even at the cap: the draft was admitted when it was made, and turning a drafter away from the match they drafted for is worse than a brief overshoot. While open they count, so new tables and drafts are refused.
- **Basic lands a deck adds** are the drafted set's own printing when it has one, else any printing.
- **Between the games of a best-of-three** the pool is the registered one, as in Standard: the deck, and the rest of the drafted pool as the sideboard. Basic lands beyond the ones the deck registered are not offered there.
- **What is stored (section 9).** Migration 0038, `draft_history` and `draft_history_matches`: for a signed-in seat, the draft's name, set and date, the cards picked and which of them the deck played, the basic lands added, and each round's result. An opponent who was signed in is stored as a pointer to their account and shown by the name it has now; no other person's name is stored. A record is written when the draft finishes, when the person leaves after the picks, and when the server closes the draft. A guest's seat writes nothing. Account deletion deletes the person's rows and clears the pointer wherever another person's record named them. `GET /me/drafts` and `GET /me/drafts/{id}` read a person's own.
- **The entry point.** Signed-in Home's start cards (ADR 0156 §6: "Practice game", "Bring a deck" and, with an invite, "No pod yet?") gain a "Draft a set" card, which opens the create form at `#/draft`. Like the others it shows while the person has no table. "My drafts" (`#/drafts`) is a link on the create form and on a finished draft, not a new entry in the header: ADR 0156 §3 fixes the header's groups, and a draft is not yet a place a person goes every visit. Whether the header should name Draft is left to the owner.
- **The way back from a match.** The game page's leave control and its game-over button go to the draft, not the lobby, at a table the draft page sent the player to, and that table offers no rematch: the draft pairs the next round.

Not built here: a weighted coin or an arena result for bot matches; a way to add basic lands while sideboarding; the away-player offer and the report button at a match table are whatever the table already gives (a report works as at any table, the away offer does not appear); a person's finished drafts cannot be deleted one at a time; the admin views list a draft's match tables as practice tables and do not list drafts.

## Consequences

- The catalog's "Full" now carries more weight: a draftable set is a promise that every card in its packs works unaided.
- Two-player games become common, so two-player rules and layouts get real use. The first-draw rule changes existing two-seat Commander tables too, which is correct (CR 103.8a) and is called out in the release notes.
- The lobby gains a second kind of thing to join. The front page and the Discord bot's commands need a place for it.
- Commander work is not displaced: slices are sized to run beside the card-coverage work, and the charter says which wins when they compete.

## Sprint

S81 — Formats: Standard 1v1 and booster draft, opened by the owner on 2026-11-10 with this ADR.
