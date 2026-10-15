# ADR 0196 — Bug reports mirrored to a Discord forum thread

**Status:** Accepted · 2026-10-15 · Backlog
**Issue:** [#3327](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3327)
**Numbering:** the highest ADR on `develop` was **0192**; open branches hold **0193** (`docs/adr-demonstrate-a-loop`), **0194** (`fix/…` snapshot-only-what-someone-can-undo) and **0195** (`docs/adr-blue-green-deploys`), so this is **0196**. Every remote head was listed (`cost-ledger`, `docs/*`, `feat/*`, `fix/*`, `issue-3016-reuse-trigger-order`, `perf/clone-3136`, `test/3744-mount-copy-saddlers`, `wip/836-one-click-default`, `main`, `develop`); no open issue holds a reservation for 0196.
**Builds on:** [ADR 0017](0017-bug-report-button.md) (the in-app report, its throttle §10 and its redaction §9), [ADR 0174](0174-a-game-results-feed-in-discord.md) and [ADR 0161](0161-an-lfg-board-in-discord.md) (the server posts to a community-server channel with the bot token), [ADR 0023](0023-develop-environment.md) (the bot binary is never deployed from `develop`).

---

## Context

An in-app report becomes a GitHub issue. Most players have no GitHub account, so they cannot see what became of a report or answer a maintainer's question. They are in the CMD & CTRL Discord server, and most are signed in to the app with Discord.

The issue asks for a thread per report in a forum channel (`#bug-reports`): the report text, the screenshot and a link to the issue, the reporter mentioned, a note and a close when the issue closes, and nothing posted for private tables or sensitive text. It also leaves two things open that this ADR decides: which process posts, and how a thread learns the issue closed.

## Decision

### 1. Opt-in by one variable

`CMDCTRL_DISCORD_BUG_FORUM_ID`, read by `discord.BugForumFromEnv`:

- Unset: **off**. Nothing is posted anywhere, and no report behaves differently.
- Not a snowflake: the **boot fails**, naming the variable.
- Set without `CMDCTRL_DISCORD_BOT_TOKEN`: off, said once at boot (the token is production only, ADR 0159 §7).
- Set without a bug-reporting credential (no GitHub App and no PAT): off, said once at boot. The mirror needs the same client to read issue state.

It needs no guild id. The forum channel knows its own guild, and the close sweep reads it from there, which also lets the sweep refuse an id that names something that is not a forum.

### 2. The server posts, not the Discord bot binary

The server holds the report, the bot token (`discord.BotFromEnv`), the issue number and the GitHub client. The `cmd/bot` binary is a separate deploy that holds none of them and **must never be deployed from `develop`** (AGENTS.md §4: two bot processes on one application answer every slash command twice), so a feature that lived there could not be exercised on the dev environment at all. The #lfg board (ADR 0161), the voice rooms (ADR 0159) and the results feed (ADR 0174) already post this way, with `botREST`; this is one more file in `server/internal/discord/` (`bugforum.go`) and one in `server/internal/lobby/` (`bugforum.go`). The bot binary is unchanged.

### 3. Which reports are mirrored

A report is mirrored only when **all** of these hold:

- **It was filed as an issue.** A report that only added a +1 to an existing issue, and one that was queued for an admin (ADR 0017 §10), opens no thread: there is no new issue to follow. A queued report is not filed yet.
- **It does not come from a private table.** Every table is invite-only, and a report carries the reporter's words, screenshots of the board and their name; a Discord forum shows all of it to everyone in the community server who was never invited. So a report whose context names a game, or whose session is bound to one, is mirrored only if that table is **public**, which here means the host posted it to the #lfg board (`GameMeta.LFG`) and it is not the tutorial's practice table. A report from outside any table (the lobby, the deck pages) has no table to protect and is mirrored. Anything that makes the table's publicness unknowable counts as private: a game id that does not parse, a table the lobby no longer holds, no lobby to ask.
- **Its text is the text the issue carries.** The title and description go through `redact.Secrets` (ADR 0017 §9) exactly as the issue's do, so a session token or a reclaim ticket typed or pasted into a report never reaches Discord either. Nothing else from the report is sent: not the client log, the user agent, the game id, the turn or the pinned replay and game log.

The throttle (ADR 0017 §10) is the rate limit: a thread is attempted only on the `filed` route, which the per-account window and the daily cap already bound. The mirror adds a second wall of its own: at most four threads being opened at once, the rest dropped with a log line rather than queued.

### 4. What a thread says

One forum post per report, created with `POST /channels/{forum}/threads`:

- **Name:** `#<issue number> <title>`, one line, cut to Discord's 100 characters. The number leads because the close sweep reads it back (§5).
- **First message:** an embed titled `GitHub issue #N` that links to the issue, with the reporter's description as its text, the kind (`bug`, `idea`, `question`), the issue URL, and the screenshots. The first screenshot is the embed's image and every one is listed as a link. Screenshot links are the same absolute, publicly reachable URLs the issue already uses (ADR 0017 §6); with attachments off there are none.
- **Content line:** `<@reporter> thanks for the report. A maintainer may reply in this thread.` The reporter is mentioned only when the session has a Discord id; a guest gets the second sentence alone. This is how a question a maintainer asks in the thread reaches the reporter: Discord notifies the mention, and the forum notifies a follower of the post.

**Markup and mentions.** The description is reporter-typed, so every character Discord reads as markup or a mention is backslash-escaped (`escapeText`, the helper the results feed's `plainText` is built on, keeping line breaks), and `allowed_mentions` is `{"parse": [], "users": ["<reporter id>"]}`, which means nothing in the content or the description can ping anyone except that one person: an `@everyone`, a role mention, or a `<@123>` in the report is text. The reporter id must itself be a snowflake or it is dropped. Escaping `:` and `[` also leaves a pasted link inert rather than clickable. Embed image and link URLs are accepted only if they begin `http://` or `https://`.

### 5. Closing the thread: the server polls GitHub, nothing is persisted

A GitHub webhook would need a public route on the server, a shared secret and a payload verifier, for a feature that runs a handful of times a week. The Discord bot binary cannot poll for the reasons in §2. So the server polls.

`BugForum.Sweep` runs 30 seconds after boot and then every 5 minutes. It:

1. asks Discord for the guild's active threads (`GET /guilds/{id}/threads/active`) and keeps only those whose parent is the forum, **whose owner is the bot itself**, and whose name starts `#<number>`;
2. reads each one's issue (`GetIssue`), at most 50 per sweep;
3. for an issue that is **closed**, or **gone** (a 404 or 410 from GitHub, which is a deleted spam report), posts one note (`GitHub issue #N is closed, so this thread is closed too. Thanks for reporting it.`, or `… was removed, so this thread is closed.`) and archives the thread.

Why this shape:

- **No state to keep.** The thread's name carries the issue number and Discord remembers the threads, so a restart loses nothing and nothing needs a migration, a file or a column. The cost is that renaming a thread away from `#<number> …` (or a moderator deleting it) detaches it from the sweep, which is the correct way for a maintainer to say "leave this one alone".
- **Only the bot's own posts are ever touched.** The owner check means a person's forum post titled `#5 …` cannot be archived by closing issue 5.
- **Archive, not lock or delete.** Anyone who posts in an archived thread reopens it, so a reporter can still answer late. The sweep remembers, in memory, each thread it has closed and leaves it alone while its issue stays closed, so a conversation on a closed issue is not closed over and over; if the issue is reopened and closed again, that is a new closing and gets a new note. A thread whose archive failed is retried without posting the note twice. The memory is one small entry per closed report and is lost on restart, after which a thread reopened in the meantime is noted and archived once more.
- **Cost.** A sweep is three Discord reads and at most 50 GitHub reads (authenticated, 5,000 an hour): well inside both limits for a forum with a handful of open posts. The active-thread listing has no paging and includes every thread in the server, so the call reads up to 2 MiB rather than the 8 KiB every other call reads.
- **A failure waits for the next sweep.** A 429 from Discord ends the sweep; a GitHub error leaves that thread for next time.

### 6. Failure never reaches the reporter

The mirror runs after the issue exists and in the background (`Mirror` returns at once). A Discord outage, a 403 for a missing permission or a wrong forum id costs a missing thread and one WARN line. A 429 waits its `retry_after` (capped at 30 s) and a 5xx or timeout 2 s, three tries in all, the results feed's count; a 403 or 404 is tried once. A restart between the issue and the thread loses the thread, which is the cost of having no outbox (ADR 0174 §5).

### 7. Configuration and secrets

- `CMDCTRL_DISCORD_BUG_FORUM_ID` is a Discord snowflake, not a secret. The bot token is the only secret involved; it is the one `discord.BotFromEnv` already reads, it appears only in the `Authorization` header (`botREST`), and no error, log line or response carries it.
- **CD** writes the variable from the Actions variable of the same name in the "Sync server env (community)" step, always, empty included, after checking it is a snowflake, and the step's summary says `bug forum set|unset`. `docs/environments.md` and `docs/env.md` document it.

### 8. Privacy

The privacy notice (ADR 0149) gains a sentence in the Discord entry: a report filed outside a table, or at a table the host posted to #lfg, is also posted in the community server's bug forum with what the reporter wrote, their screenshot links and the issue link, mentions them if they signed in with Discord, is closed with the issue, and stays until a moderator deletes it; deleting an account does not remove a post already made. No new personal data is stored by this server.

## Consequences

- **A player with no GitHub account can follow a report and answer a question,** for reports from the lobby, the deck pages and public tables.
- **Reports from ordinary invite-only tables are not mirrored.** That is most in-game reports, deliberately: the thread would publish a private table to the community. If the owner wants those too, the lever is a per-report "share this in Discord" checkbox in the report modal, not a looser server rule; it is not built here.
- **The close lags the issue by up to five minutes,** a renamed thread is never closed, and a restart can close a reopened thread once more.
- **The close is a poll.** If the mirror is ever the only reason the server calls GitHub on a timer, a webhook is the next step; at this volume it is not worth a public route.

### Owner setup (once, not code)

1. Create a **Forum** channel (`#bug-reports`). Do not turn on **Require tags**: the bot creates posts without any. On that channel give the `cmd_and_ctrl` role **View Channel, Send Messages, Create Posts, Send Messages in Posts, Embed Links, Manage Threads**, set on the channel itself (ADR 0159's lesson: permissions are computed in the channel's context). It does not need **Mention Everyone**: the mention is one user.
2. Set the Actions variable `CMDCTRL_DISCORD_BUG_FORUM_ID` (prod environment).

## Slices

| # | Slice | Needs |
|---|---|---|
| 1 | **Server:** this ADR, `discord.BugForum`, `lobby.BugForum` (create at filing, the close sweep), `Config.BugForum`, `CMDCTRL_DISCORD_BUG_FORUM_ID`, the CD variable, the docs and the privacy sentence | — |
