# ADR 0174 — A game results feed in Discord

**Status:** Accepted · 2026-10-14 · Backlog
**Issue:** [#3322](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3322)
**Owner decisions:** the defaults of 2026-10-14 quoted under [Owner decisions](#owner-decisions-2026-10-14). They are binding. The calls the issue left open and this ADR made are listed under [Calls made here](#calls-made-here).
**Numbering:** the highest ADR on `develop` was **0173**; 0172 is claimed by an open PR, so this is **0174**. Migration **0024** follows 0023 (game statistics).
**Builds on:** [ADR 0161](0161-an-lfg-board-in-discord.md) (the #lfg board: the same opt-in, the same bot-token REST calls, the same config rules), [ADR 0165](0165-game-statistics-a-record-derived-from-every-finished-game.md) (the record this post is computed from, and the moment it is computed), [ADR 0149](0149-privacy-notice-and-account-deletion.md) (the privacy notice), [ADR 0156](0156-a-front-door-site-shell-front-page-and-a-discord-gate.md) (the community server, `GET /community`).

---

## Context

Finished games are the server's best conversation starter, and a post in a results channel shows the app to people who have not tried it. Everything the post needs already exists:

- **The game statistics record** (ADR 0165) is computed once per stored game, off the lobby lock, the first time the lobby sees it ended (`recordStatsLocked`). It has the turn count, each seat's place, result, elimination turn, concession and commander names.
- **The server calls Discord with the bot token** (ADR 0051 decision 5; ADR 0159 voice rooms; ADR 0161 `discord.LFG`). Posting a channel message with an embed is a call the #lfg board already makes.
- **The host opt-in has a pattern** (ADR 0161 §1): a boolean on the table, `PUT /games/{id}/lfg`, a create-form checkbox and a Table settings section, hidden when the server has no channel.

What it adds is a post that names the seats at a table, which is what makes it an opt-in decision and a privacy line.

## Decision

### 1. The host opts a table in

Owner decision 1. Per table, off by default, mirroring ADR 0161 §1.

- **Create:** `POST /games` takes `"results": true`. The create form shows a checkbox, only when the feature is on (§6).
- **Later:** `PUT /games/{id}/results` with `{"enabled": true|false}`, from Table settings. Host (ADR 0075 §2.1) and the server admin.
- **Which tables:** any table whose game has not ended and that is not archived or the practice table. A guest host may opt in: unlike an #lfg post, this one names seats, not the host's Discord account.
- Lobby metadata, like the voice and #lfg choices: not a `game.TableSettings` key, never in the engine snapshot. `GameMeta.results` is public, so every seat at the table can see that the result of their game will be posted.

### 2. The server posts

Owner decision 2. The server, not the gateway bot process, posts, with `discord.Results` in `server/internal/discord/results.go`. It reuses `botREST` and the voice error mapping that `discord.LFG` uses: the token is only in the `Authorization` header, a 403 becomes `ErrResultsForbidden`, a 404 `ErrResultsChannelNotFound`, and Discord's own error text is never forwarded.

### 3. When a game is not posted

Owner decision 3. Never:

- a **practice table**;
- a table whose host did **not opt in**;
- a game whose **concession came before turn `resultsMinTurns` = 3**: it is posted only when its last turn number is at least 3, or when nobody conceded;
- a game that ended **with no outcome** (closed by its host or an admin before it was played out): there is no result to post. This is a call made here (see below).

### 4. What the post says

Owner decisions 4 and 5. One embed, in the configured channel:

- **Title:** the table's name.
- **Description:** `**Winner:** <seat>`, or `**Draw** between <seats>`.
- **One field per seat, in placing order:** `1st · Alice`, the seat's commander name(s) joined with ` / ` (`No commander` when it has none), and, for a seat that left, `Conceded on turn N` or `Eliminated on turn N`. The seat name is the one the table shows. A seat played by an AI agent is suffixed `(AI agent)` and a bot seat `(bot)`.
- **Turns** and **Duration** fields: the game's last turn number, and from the game starting (the table opening, when no start is on record) to its end.
- Footer `Game ended`, timestamp the end.

**No mentions.** There is no content line, `allowed_mentions` is `{"parse": []}`, and every name is escaped as text (`plainText` in `results.go`: markup, `<`, `>`, `@`, `#`, `:` get a backslash, newlines become spaces), so a seat called `<@123>` or `@everyone` is only text.

**No card images.** Neither the engine nor the Discord package has a helper that produces a Scryfall image URL from a commander name, and `gamestats.Seat.Commanders` holds names only. Names only. The issue's "notable moment from the game log" is skipped.

An example body is in `server/internal/discord/results_test.go` (`TestResultsPostCardAndNoMentions`).

### 5. When and how it is posted

Owner decision 7. The post is hooked where the statistics record is: `Lobby.recordStatsLocked`, which `syncStateLocked` calls once per game on its first sight of ended, after the `games` row has its outcome. Under `l.mu` it captures what the table knows (name, opt-in, start and end times, outcome) into a `resultsJob`; on the same goroutine, off the lock, it runs `gamestats.Compute` once, writes the record if there is a store, and then builds and sends the post from the same `Compute` output and the final game view (which carries each seat's name and whether it is an agent or a bot). The lobby lock is never held across a Discord call. `WaitGameStats` waits for both.

- A practice table already skips the whole path.
- A server with no stats store but with the feed still posts: the two share the computation, not the dependency.
- **Failures:** a 429 waits its `retry_after` (capped at 30 s) and a 5xx or timeout waits 2 s, three tries in all, the #lfg manager's count (`lfgMaxAttempts`). A 403 or 404 will fail the same way again and is tried once. A dropped post is logged at WARN and never tried again. A restart between the game ending and the call loses the post; the game is long over and it is the cost of having no outbox.
- **Exactly once:** the post is made on the first sight of ended, which the lobby takes once per game, and a game restored after a restart is already ended in its row and is never seen ending again.

### 6. Configuration

Owner decision 6. `CMDCTRL_DISCORD_RESULTS_CHANNEL_ID`, read by `discord.ResultsFromEnv` with `LFGFromEnv`'s rules:

- Unset: the feature is **off**; nothing is posted; `GET /community` says `"results": false`; `PUT /games/{id}/results` answers 503 naming the variables.
- Set without `CMDCTRL_DISCORD_COMMUNITY_GUILD_ID`, or either not a snowflake: the **boot fails**, naming the variable.
- Set with the guild but without the bot token: off, said once at boot (the token is production only, ADR 0159 §7).
- `GET /community` gains `"results": true|false`.
- **CD** writes it from the Actions variable of the same name in the "Sync server env (community)" step, always, empty included, after the same checks before anything is written (a snowflake; the guild is set), and its summary line says `results channel set|unset`. `docs/environments.md` documents it beside the #lfg variables.

### 7. Durability

Migration 0024: `games.results INTEGER NOT NULL DEFAULT 0 CHECK (results IN (0, 1))`, written through `persistGameLocked` and restored with the table. No message id is stored: the post is made once and never edited or deleted. The column holds no personal data, so account deletion's inventory lists it as `NotPersonal`.

### 8. Privacy

Owner decision 8. The post publishes seat names to every member of the CMD & CTRL Discord server, which is why it is the host's opt-in and why the opt-in is visible to the seats. The privacy notice gains a row: when a host turns on the results post, the table's name, each seat's name, its commander(s) and the result are posted in the server's results channel when the game ends, and stay there until a moderator removes them; a signed-in player's account deletion does not remove a post already made. What a seat is called is whatever the person set at the table, as it is on the board. No new personal data is stored by this server.

## Consequences

- **A host who wants the table's result shared ticks one box;** a table that did not is never posted, so the channel is only what people chose to show.
- **Players at a table can see the choice,** but cannot refuse it except by leaving: a seat name is public to anyone who joins a public table anyway, and the host already decides the table's visibility (ADR 0075 §2.6). The notice says so.
- **Deleting an account does not unpost.** The post is a Discord message, and the server does not keep its id, so it cannot find it again; a moderator deletes it in Discord. The notice says so.
- **No catch-up.** Games that ended before the feature, or while the channel was unset, are not posted later.
- **Production only,** like voice rooms and #lfg: one Discord application, one channel.

### Owner setup (once, not code)

1. Create **#game-results** (or use an existing channel). On that channel give the `cmd_and_ctrl` role **View Channel, Send Messages, Embed Links**, and for a private channel or a restricted category set them on the channel itself (ADR 0159's lesson: permissions are computed in the channel's context). It does not need **Mention Everyone**: the post pings no one.
2. Set the Actions variable `CMDCTRL_DISCORD_RESULTS_CHANNEL_ID` (prod environment).

## Owner decisions, 2026-10-14

1. **Opt-in:** per table, by the host, off by default, mirroring the #lfg board.
2. **Who posts:** the server, with the existing bot token and REST client.
3. **Never posted:** practice tables, concessions before turn 3, tables without the opt-in.
4. **Names:** each seat by its seat name as the table shows it, guests and agents included; no mentions.
5. **Content:** table name, winner or draw, commanders and placing, turns, duration; no notable moment; card images only if a helper existed.
6. **Config:** `CMDCTRL_DISCORD_RESULTS_CHANNEL_ID`; unset is off; set without the guild fails the boot; wired through CD like the #lfg channel.
7. **Data source:** the game statistics record, hooked where it is recorded.
8. **Privacy:** a line in the privacy notice and, for a new column, the deletion inventory.

## Calls made here

- **`resultsMinTurns = 3`** is a constant, not a setting. A concession counts when any seat conceded and the game's last turn number is below it. A short game nobody conceded (there is none in practice) is still posted.
- **No outcome, no post.** A game closed by its host or an admin, with neither a winner nor a draw, has nothing to report, whatever its turn count.
- **Guest hosts may opt in,** unlike #lfg: the post carries no Discord account.
- **No opt-in after the end.** A table that has ended cannot be opted in, since its post has already been decided; opting out is always allowed.
- **No manager.** The #lfg board needs a reconciler because its post follows the table; this post is one call at one moment, so it runs on the statistics goroutine.
- **Retries** are the #lfg manager's three tries, not an outbox.
- **A server with no stats store still posts,** because the feed depends on the final game view, not on a database record.

## Slices

| # | Slice | Needs |
|---|---|---|
| 1 | **Server:** this ADR, migration 0024, `discord.Results`, the post at the stats hook, `results` on create and `PUT /games/{id}/results`, `GET /community`'s flag, docs, CD variable | — |
| 2 | **Client:** `parseCommunity`, `GameMeta.results`, the create-form checkbox, the Table settings section, the privacy row | 1 |
