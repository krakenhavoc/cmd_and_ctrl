# ADR 0222 — Formats beyond Commander: Standard 1v1 and booster draft

**Status:** Accepted · 2026-11-10 (Ian read the summary of this ADR and its calls and said to go ahead) · [S81](../sprints.md)
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

- A pack is generated from Scryfall data for one set: the cards the dump marks as found in boosters, in slots by rarity, at the rates Wizards publishes for that set's play booster. Where a set's exact sheet isn't public, the pack follows the published slot rates and says so on the set's draft page. **Packs are approximations of retail packs, and the page says that.**
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

## Consequences

- The catalog's "Full" now carries more weight: a draftable set is a promise that every card in its packs works unaided.
- Two-player games become common, so two-player rules and layouts get real use. The first-draw rule changes existing two-seat Commander tables too, which is correct (CR 103.8a) and is called out in the release notes.
- The lobby gains a second kind of thing to join. The front page and the Discord bot's commands need a place for it.
- Commander work is not displaced: slices are sized to run beside the card-coverage work, and the charter says which wins when they compete.

## Sprint

S81 — Formats: Standard 1v1 and booster draft, opened by the owner on 2026-11-10 with this ADR.
