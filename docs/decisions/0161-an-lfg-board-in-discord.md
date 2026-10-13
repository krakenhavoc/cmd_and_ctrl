# ADR 0161 — An LFG board in Discord

**Status:** Proposed · 2026-10-13 · Backlog
**Issue:** [#3319](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3319)
**Owner decisions:** the two answers of 2026-10-13 are quoted under [Owner decisions](#owner-decisions-2026-10-13). They are binding. This ADR also makes calls the answers did not cover; they are listed under [Calls made here](#calls-made-here) so the owner can overturn any of them in review, before slice 2 lands.
**Numbering:** every ADR file on all 12 remote heads was listed after `git fetch --all --prune`. The highest number anywhere was **0160**, so this is **0161**. ADR 0162 (DM notifications) and ADR 0163 (Linked Roles) are drafted in the same PR.
**Builds on:** [ADR 0051](0051-user-database.md) decision 5 (the server's bot-token REST calls), [ADR 0075](0075-table-settings-and-host-controls.md) §2.1 (the host) and §2.6 (moderation for a public table), [ADR 0149](0149-privacy-notice-and-account-deletion.md) (the privacy notice), [ADR 0156](0156-a-front-door-site-shell-front-page-and-a-discord-gate.md) (the community server, `GET /community`), [ADR 0159](0159-per-table-discord-voice-rooms.md) §4 (the reconciler pattern this reuses).

---

## Context

The hardest part of a game of Commander is finding three other people. The CMD & CTRL Discord (ADR 0156) is where the players are, and its onboarding already hands out an **LFG** role to people who want to hear about open tables. Today a host who wants players copies the invite link into a channel by hand.

What already exists:

- **A table's invite link is how anyone joins it.** `/#/games/{id}/join?t={invite_token}`. ADR 0075 §2.6 already treats "a public post of the invite" as the public-table case, and gave hosts the tools for it: the name and chat filter, remove-and-block, and reports.
- **The server makes bot-token REST calls itself** (ADR 0051 decision 5, `discord.Bot`; ADR 0159 `discord.Voice`), from the server process, without the gateway bot. Posting, editing and deleting a channel message are the same kind of call.
- **ADR 0159's voice manager** is a reconciler: lobby hooks mark a table dirty, one worker compares what the table wants with what exists in Discord and makes the calls, never holding `l.mu` across a call, and a boot pass adopts or deletes what survived a restart.
- **`/c2-games`** lists active and lobby tables in reply to a slash command. It is a pull; this is a push.

## Decision

### 1. The host opts a table in

Owner answer 1.

- **Create:** `POST /games` takes `"lfg": true`. The create form shows **Post to the CMD & CTRL Discord (#lfg)**, off by default, only when the feature is on (§7).
- **Later:** `PUT /games/{id}/lfg` with `{"enabled": true|false}`, from table settings. Host only (ADR 0075 §2.1), plus the server admin.
- **Which tables:** a table in the **lobby** state with at least one empty seat, not a practice table, whose host is **signed in with Discord**. A guest host gets the checkbox disabled with "Sign in with Discord to post to #lfg"; the post names the host, and a guest has nothing to name.
- Like ADR 0159's voice choice, this is lobby metadata, not a `game.TableSettings` key, and never enters the engine snapshot.

### 2. The post

One message per table in the configured #lfg channel, posted by the bot:

- **Content line:** the LFG role ping (§3), or nothing.
- **Embed:**
  - title: the table name (already through the ADR 0075 §2.6 name filter);
  - **Host:** a Discord mention of the host (`<@id>`), which renders as their server nickname;
  - **Seats:** `2 of 4`, the count only (see [Calls made here](#calls-made-here) on names);
  - **Voice room:** yes or no (ADR 0159);
  - the table's settings that change how the game plays, when not the default: starting life, and the house-rule toggles shown in table settings;
  - footer: when the table was created.
- **Components:** one **link button**, **Join table**, to the invite URL. A link button needs no interaction handling, so the gateway bot is not involved and the server needs no interactions endpoint.

### 3. The ping

Owner answer 2.

- The post that **creates** the message pings the LFG role once, with `allowed_mentions` naming only that role. Edits never ping; Discord does not re-notify on an edit anyway.
- **Cooldown:** at most one ping per host per **30 minutes**. A post inside the cooldown is made without the ping and says nothing about it.
- `CMDCTRL_DISCORD_LFG_ROLE_ID` unset means posts never ping.

### 4. The post follows the table

The manager edits or removes the message as the table changes. Edits are coalesced to at most one per post per **10 seconds**, under Discord's per-channel rate limit.

| Table event | The post |
|---|---|
| a seat filled or freed | edited: seat count |
| all seats filled | edited: **Full**, the button disabled; a freed seat restores it |
| invite rotated | edited: the button's new URL (the old link stops working, so the post must not keep it) |
| renamed, settings changed, voice toggled | edited |
| the game **starts** | **deleted** |
| host turns LFG off, table archived or deleted, `closeUnresidentRows` at boot | **deleted** |

A block (ADR 0075 §2.6) still applies to anyone who clicks Join: the post is only a copy of the link.

### 5. How the server talks to Discord

- **`discord.Channel`** (or methods on `discord.Voice`'s client, whichever slice 2 finds cleaner): create a message, edit a message, delete a message. Same token, same no-token-in-logs rule, same error mapping (`ErrVoiceForbidden`'s twin for a 403, 404 is "already gone").
- **An LFG manager** per lobby with ADR 0159 §4's shape: `notify(gameID)` from the lobby hooks (seat claim, seat leave, rename, settings, invite rotation, state change, the toggle), one worker goroutine, no `l.mu` across a call. It reads what the table wants (`Lobby.lfgTable`), compares it with the message it knows about, and posts, edits or deletes.
- **Failures:** a 429 waits its `retry_after` (capped) and retries; a 5xx or timeout retries with backoff up to 3 times and logs at WARN. A 403 logs once naming the channel and marks the table's post `unavailable`; the table plays exactly as it would without the post. The host's toggle tries again.

### 6. Durability

- **Migration** (the next free number when slice 2 lands): `games.lfg INTEGER NOT NULL DEFAULT 0` and `games.lfg_message_id TEXT`. Written through `persistGameLocked`. Account deletion's inventory (ADR 0149) gains nothing: neither column holds personal data.
- **Boot reconcile:** for each `games` row with a `lfg_message_id`, keep and re-render the message if the table is resident, in the lobby and still opted in; otherwise delete it and clear the column. A 404 on edit clears the id and posts afresh, without a ping. The manager never lists or touches messages it did not post: #lfg is a shared channel.

### 7. Configuration

- **New:** `CMDCTRL_DISCORD_LFG_CHANNEL_ID` (the #lfg channel) and `CMDCTRL_DISCORD_LFG_ROLE_ID` (the role to ping; optional).
- **On** when the channel, `CMDCTRL_DISCORD_COMMUNITY_GUILD_ID` and `CMDCTRL_DISCORD_BOT_TOKEN` are set. A channel without the guild, a role without the channel, or a value that is not a snowflake **fails the boot** naming the variable, as `GateFromEnv` and `VoiceFromEnv` do. The channel without the bot token leaves the feature off, said once at boot (the token is production only, ADR 0159 §7).
- `GET /community` gains `"lfg": true|false`. Off: no checkbox, `PUT /games/{id}/lfg` answers 503 naming what is missing.
- **CD** writes both from Actions variables of the same names in the community step, always, empty included, with the same checks before anything is written.

### 8. Privacy

The notice gains a sentence: when a host posts a table to #lfg, the post shows the table's name, the host's Discord account and the seat count to every member of the CMD & CTRL server, and is deleted when the game starts or the post is withdrawn. No new personal data is stored.

## Consequences

- **A host fills a table in one click,** and the server's LFG members hear about it once.
- **The invite is visible to the whole server.** That is the public-table case ADR 0075 §2.6 was written for; the host's tools there are the answer to a bad actor, and rotating the invite updates the post.
- **The post is a mirror, never a source of truth.** Clicking Join on a stale post (a missed edit) leads to the table's own join page, which says if the table is full or started.
- **Production only,** like voice rooms and DM invites: one Discord application, one #lfg channel.

### Owner setup (once, not code)

1. Create **#lfg** (or use an existing channel). On that channel, give the `cmd_and_ctrl` role **View Channel, Send Messages, Embed Links, Read Message History**. *Embed Links is not in the bot's original invite.* Learned from ADR 0159: permissions are computed in the channel's context, so set them on the channel itself when it is private or its category restricts the bot.
2. Make the **LFG** role mentionable (Server Settings → Roles → LFG → **Allow anyone to @mention this role**), or give the bot **Mention @everyone, @here, and All Roles** on #lfg. Without one of them the ping renders as text and notifies nobody.
3. Set the Actions variables `CMDCTRL_DISCORD_LFG_CHANNEL_ID` and `CMDCTRL_DISCORD_LFG_ROLE_ID` (prod environment).

## Owner decisions, 2026-10-13

1. **How a table is posted:** the host opts in, with a checkbox at create and in table settings. Recommended option, chosen.
2. **When the role is pinged:** once per post, never on edits, with a 30-minute per-host cooldown. Recommended option, chosen.

## Calls made here

- **Seat count, not seated names.** A post is visible to the whole server, and players who sat down did not choose to be advertised. The host chose; the host is named.
- **Delete on start** rather than edit to "started". #lfg stays a list of tables you can join.
- **Signed-in hosts only.**
- **10-second edit coalescing** and **30-minute ping cooldown** are constants, not settings.

## Later

[#3321](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3321) (scheduled games as Discord Scheduled Events) builds on this: at a scheduled table's start time its empty seats are posted here.

## Slices

| # | Slice | Model | Needs |
|---|---|---|---|
| 1 | This ADR | — | — |
| 2 | **Server:** config and boot checks (§7), the message calls (§5) with httptest coverage, the LFG manager (§4, §5), migration and boot reconcile (§6), `lfg` on create and `PUT /games/{id}/lfg`, `GET /community`'s flag; docs/lobby.md; CD variables | Opus | 1 |
| 3 | **Client:** the create checkbox, the table-settings toggle with the post's status ("Posted in #lfg", "unavailable"), the guest-host hint; privacy paragraph; screenshots at 1440 px and 820 px | Sonnet | 2 |
