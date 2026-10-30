# ADR 0215 — Scheduled games: Discord Scheduled Events with RSVP seats

**Status:** Proposed · 2026-10-30 · S77 — Discord and social: voice, friends, mod tools and events
**Issue:** [#3321](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3321)
**Owner decisions:** none yet. The product questions are under [Owner decisions](#owner-decisions), each with the recommended option first. The calls this ADR makes on its own are under [Calls made here](#calls-made-here), for the owner to overturn in review. No production code before the owner has answered.
**Numbering:** checked on 2026-10-30 after `git fetch --all --prune`. `develop` ends at 0214 (ADR 0214 merged in #4665 while this was drafted), and `git ls-tree` over every remote head's `docs/decisions/` shows nothing above it. The heads checked are listed in the PR. This is **0215**.
**Builds on:**
- [ADR 0004](0004-discord-identity.md) (the bot, its install with `permissions=0`) and [ADR 0051](0051-user-database.md) decision 5 (the server's own bot-token REST calls).
- [ADR 0075](0075-table-settings-and-host-controls.md) §2.1 and §2.6 (the host, the name filter, remove-and-block).
- [ADR 0110](0110-remember-me.md) §5 (the last setup, the open-table cap and the create rate limit).
- [ADR 0148](0148-capacity-limits-for-a-public-launch.md) (the load cap) and [ADR 0149](0149-privacy-notice-and-account-deletion.md) (the privacy notice, `accountdelete.Inventory` and `Delete`).
- [ADR 0156](0156-a-front-door-site-shell-front-page-and-a-discord-gate.md) §1 and its amendments (the community server, the sign-up gate, the invite-flow guest, ban sync).
- [ADR 0159](0159-per-table-discord-voice-rooms.md) (voice rooms), [ADR 0161](0161-an-lfg-board-in-discord.md) (#lfg, which names this issue under "Later"), [ADR 0162](0162-discord-dm-notifications.md) (DM notifications).
- [ADR 0211](0211-closing-idle-tables.md) (closing idle tables), [ADR 0213](0213-friends-recently-played-with-and-live-presence.md) (friends, personal blocks), [ADR 0214](0214-cmd-and-ctrl-as-a-discord-activity.md) question 6 (a refused member seated as a guest).

**Amends, if accepted:** ADR 0211 gains one rule for a scheduled table that nobody comes to (§7). ADR 0162 gains a fourth kind of DM, beside the one ADR 0213 adds (§9).

---

## Context

Many Commander groups play at a set time: a weekly pod night, or a game arranged in chat for Friday. Today that is arranged outside C2. Someone names a time in a channel, people say they're in, and at the time someone makes a table and pastes its link.

Discord has **Scheduled Events** for this. An event shows in the server's event list and has an **Interested** button. The owner's idea (the issue body, 2026-10-13):

- a host schedules a table for a date and time, and the bot creates a server event that links to it;
- **Interested** reserves a seat for the first N people, and the rest join a waitlist;
- Discord sends its own reminder;
- at the start time the table opens, and an #lfg post goes out for any empty seats;
- cancelling or rescheduling in the app updates the event.

The issue leaves three questions open: no-shows, how long a reservation is held, and time zones.

### What exists

- **A table exists from the moment it is created.** `POST /games` makes a lobby table with a room in memory. Any signed-in person may create one, with three limits: three open lobby tables per creator (`MaxOpenTablesPerCreator`, `lobby/player_tables.go`), one create per 30 s with a burst of 3, and ADR 0148's load cap. The engine knows three states, `lobby`, `active` and `ended` (`GameMeta.State`). Nothing in it means "later".
- **The creator sits down at once** (#3494, `seatCreator`). A table holds at most 4 players and starts with at least 2 (`game.MaxPlayers`, `game.MinPlayers`). Seats are filled as people join; there is no reserved seat.
- **A named host can wait for a seat.** `GameMeta.HostDiscordID` is "a named host still waiting to sit down". `/c2-invite` uses it. The #lfg board already reads it as the host (`lfgHostLocked`).
- **`/c2-invite` creates through the admin token.** That path is uncapped and records no creator, because a server credential is not a person (`createGameWith`, `lobby/http.go`).
- **The server calls Discord's REST API itself**, with the bot token, over plain `net/http` (`discord` package, `botAPIBase`). DMs, voice rooms (ADR 0159), the #lfg board (ADR 0161), the mod log and the results feed all work this way. The two managers are reconcilers. Lobby hooks mark a table dirty, and one worker compares what the table wants with what Discord holds. Neither holds `l.mu` across a call. A boot pass adopts or removes what survived a restart (`voice_manager.go`, `lfg_manager.go`).
- **The bot process holds the gateway.** It asks for each intent only when the feature that needs it is on (`bot.Config.Intents`). It has no HTTP listener. It reports events to the server through admin routes with its admin session, as ban sync does (`POST /admin/discord-members/{discord_id}/revoke-sessions`). It uses `bwmarrin/discordgo` v0.29.0, which has `IntentGuildScheduledEvents`, the `GuildScheduledEventUserAdd` and `GuildScheduledEventUserRemove` events and `Session.GuildScheduledEventUsers`.
- **The #lfg post counts seats as `len(Players)` of `game.MaxPlayers`** (`lfgTable`). It is posted for a lobby table with an empty seat and a host on Discord. It is removed when the game starts.
- **A voice room is made when its first seated Discord member arrives** (`voiceMember`, ADR 0159 §2). It is removed after 2 hours with no seat change while nobody is in it (`voiceIdleAfter`).
- **DM notifications** are three switches, all off by default: Your turn, Table ready and Table filled (`notify.Kind`, migration 0022). ADR 0213 adds a fourth, Friend requests and invites.
- **A member without the player role** (the Planeswalker role, ADR 0156) who signs in through a table's invite link is seated as a guest named after their Discord display name (`seatRefusedAsGuest`, `lobby/access_gate.go`). ADR 0214 question 6 extends the same rule to an Activity instance.
- **Idle tables** (ADR 0211, accepted, not built yet). It closes a lobby table after 24 hours with no activity. No table is closed within 1 hour of becoming resident. Migration 0029 is reserved for it. The newest migration on `develop` is 0028.

### What Discord allows

Read on 2026-10-30. Sources are under [Sources](#sources). Points the docs leave open are under [What the docs do not settle](#what-the-docs-do-not-settle).

- **Three kinds of event** ([entity types](https://docs.discord.com/developers/resources/guild-scheduled-event#guild-scheduled-event-object-guild-scheduled-event-entity-types)):
  - `STAGE_INSTANCE` (1) and `VOICE` (2) need a `channel_id`.
  - `EXTERNAL` (3) has no channel. It needs `entity_metadata.location` (1–100 characters) and a `scheduled_end_time`.
  - A name is 1–100 characters and a description 1–1000.
- **Who can see an event.** The only privacy level is `GUILD_ONLY` (2), "only accessible to guild members". For a `VOICE` event, seeing it requires `VIEW_CHANNEL` in the event's channel. An `EXTERNAL` event needs "no permissions" ([Guild Scheduled Event, Permissions Requirements](https://docs.discord.com/developers/resources/guild-scheduled-event)).
- **Status runs** `SCHEDULED` → `ACTIVE` → `COMPLETED`, or `SCHEDULED` → `CANCELED`. Once completed or cancelled, the status "can no longer be updated". Discord changes the status itself in some cases:
  - "An external event will automatically begin at its scheduled start time" and "automatically end at its scheduled end time".
  - A voice or stage event ends minutes after its channel empties.
  - An event that never began is cancelled "after a few hours" ([Guild Scheduled Event, Status Update Automation](https://docs.discord.com/developers/resources/guild-scheduled-event)).
- **Permissions** ([bit flags](https://docs.discord.com/developers/topics/permissions#permissions-bitwise-permission-flags)):
  - `CREATE_EVENTS` (`1 << 44`) "Allows for creating scheduled events, and editing and deleting those created by the current user".
  - `MANAGE_EVENTS` (`1 << 33`) is for "editing and deleting scheduled events created by all users".
  - An external event needs `CREATE_EVENTS` at guild level. A voice event also needs `VIEW_CHANNEL` and `CONNECT` in its channel.
  - The change log of 2025-11-20 says that "Starting on February 23, 2026, bots will need the `CREATE_EVENTS` permission to create scheduled events" and that "`MANAGE_EVENTS` alone will no longer be sufficient" ([change log](https://docs.discord.com/developers/change-log#guild-expressions-and-events-permissions-now-available-to-developers)).
- **Interested users.** The docs say "subscribed". There are two ways to learn who is subscribed:
  - **The gateway.** `GUILD_SCHEDULED_EVENT_USER_ADD` and `_USER_REMOVE` carry `guild_scheduled_event_id`, `user_id` and `guild_id` ([gateway events](https://docs.discord.com/developers/events/gateway-events#guild-scheduled-event-user-add)). They come with the `GUILD_SCHEDULED_EVENTS` intent (`1 << 16`), which is **not privileged** ([intents](https://docs.discord.com/developers/events/gateway#list-of-intents)).
  - **REST.** `GET …/scheduled-events/{id}/users` returns up to 100 per page with `before`/`after` cursors. "Users will always be returned in ascending order by `user_id`", so the order people subscribed in cannot be read back.
  - **Nothing adds or removes a subscriber.** No endpoint subscribes a user, so a bot cannot mark someone Interested.
- **Limits.**
  - "A guild can have a maximum of 100 events with `SCHEDULED` or `ACTIVE` status at any time."
  - No event-specific rate limit is listed. The global limit is 50 requests a second, and per-route buckets are keyed by guild ([rate limits](https://docs.discord.com/developers/topics/rate-limits)).
- **Recurrence.** An event may carry a `recurrence_rule`, "a subset of the behaviors defined in the iCalendar RFC". Its frequencies are daily, weekly (with an interval of 2 for every other week), monthly on the nth weekday, and yearly. The docs give no occurrence or exception field on a subscriber or on the users endpoint, so they don't say whether Interested applies to one occurrence or to the series.
- **Times in messages.** `<t:TIMESTAMP:STYLE>` "display[s] the given timestamp in the user's timezone and locale" ([message formatting](https://docs.discord.com/developers/reference#message-formatting)). Events carry ISO 8601 times. A bot cannot read a user's time zone.

### Gaps found while reading

1. **The issue names the wrong permission, and re-authorizing isn't needed.** Since 2026-02-23 a bot needs **Create Events**, not Manage Events, to create an event. The bot was installed with `permissions=0`. On the C2 server its role is given permissions by hand (deploy/README.md, Discord scopes; ADR 0159's owner setup). So the owner turns on **Create Events** for the `cmd_and_ctrl` role. No new invite URL is needed.
2. **A voice event in a table's room would hide the game from everyone it should reach.** Table rooms deny View Channel to `@everyone` (ADR 0159 §3), and a voice event is visible only to members who can view its channel. Only the seated players could see the event, and those are exactly the people who don't need it. The room also doesn't exist until someone sits down. So the event must be `EXTERNAL` (§4).
3. **RSVP order can only come from the gateway.** The REST list is sorted by user id. "First N" therefore needs the bot to forward each Interested as it arrives. REST can only fill gaps afterwards (§5).
4. **A schedule from Discord would be uncapped.** `/c2-invite` creates with the admin token, which records no creator and skips the open-table cap. A `/c2-schedule` built the same way would need the server to cap the named person (§12).

## Decision

### 1. A scheduled table is a plan until its doors open

| | **A plan that becomes a table at doors-open** | A lobby table with a start time | A new engine state |
|---|---|---|---|
| Before the start | a `scheduled_tables` row: no room, no cost against the cap, no restore point | a room in memory for days, counted by the load cap and the three-open-tables cap | a fourth `game.State`, carried by snapshots and every state switch |
| ADR 0211 | nothing to exempt: the 24-hour clock starts at doors-open | needs an exemption until the start | needs an exemption |
| Voice, #lfg | unchanged, since the table is an ordinary lobby once open | both need a rule against acting early (the host's seat would make the room days ahead) | both need a rule |
| Restart | the row survives; doors open at boot if they are due | the restore pass brings it back like any table | snapshot schema bump |

**Recommended: a plan.** A scheduled table is a row until its **doors open, 15 minutes before the start**. Then the server creates an ordinary lobby table from it, through the same path as `POST /games`, and holds its seats for the people who reserved them (§3).

- **What the plan holds:** the name (through the ADR 0075 §2.6 filter, 80 characters); the start; the expected length (2, 3 or 4 hours, default 3); the host's IANA time zone; the number of players (2 to 4, default 4, the host included); and the same create options as `POST /games`: voice room, #lfg, results, and the host's last setup. It also holds "fill empty seats with bots when the grace ends" (§3).
- **At doors-open:**
  - The table is created with `created_by` = the host and `HostDiscordID` = the host's Discord id. The host is therefore the named host waiting to sit down, as with `/c2-invite`, and the #lfg board already reads that.
  - The host is **not** seated automatically. They are not necessarily at their computer. They hold seat 1 like everyone else holds theirs.
  - Opening is exempt from the open-table cap and the create rate limit, because the plan was counted when it was made (§12). It is **not** exempt from ADR 0148's load cap. If the server is full, the open is retried every minute until the grace ends (§3). After that the plan is marked `missed`, and the people holding seats are told (§9).
- **A restart** opens any plan whose doors-open has passed. A plan more than an hour past its start is marked `missed` instead.

### 2. Who schedules, from where

- **Who:** anyone who may create a table, which means a signed-in account that passes the gate. Not a guest, an agent seat or the admin token on its own.
- **From the site:** **Schedule a table** next to **New table**. The form takes a date and time, shown in the browser's zone (`Intl.DateTimeFormat().resolvedOptions().timeZone`), with a zone picker. It also takes the players, the length, the create options and, in slice 8, **Repeat weekly** or **every other week**.
- **From Discord:** `/c2-schedule` (slice 7) has the options `name`, `date`, `time`, `zone`, `players` and `length`. `date` autocompletes the next 30 days. `zone` autocompletes IANA names, and its default is the zone of the person's last scheduled table. The bot calls the server's schedule route with its admin session and `host_discord_id`. The server refuses unless that Discord id belongs to an account that passes the gate. It caps that account exactly as if the person had used the site (gap 4). `/c2-schedule list` shows your upcoming tables, and `/c2-schedule cancel` cancels one (autocomplete).
- **Where scheduled tables show:**
  - the Tables page, in a **Scheduled** section above Now;
  - `/c2-games`, under the live tables;
  - the server's Discord event list (§4).

### 3. RSVP seats, the waitlist and no-shows

**How an RSVP maps to a seat. Recommended: the first N, then a waitlist.**

- N is the number of players minus the host.
- An RSVP made while a seat is free takes the seat. Otherwise it joins the waitlist, in the order the RSVPs arrived.
- When someone leaves before doors-open, the first person on the waitlist moves up at once.
- The host may **remove** a reservation. That person can't reserve this occurrence again. A removal is per occurrence. ADR 0213's personal block (its slice 3, not built yet) is standing: once it exists, a person the host has blocked is refused silently ("This table can't take your reservation").
- Rejected alternatives:
  - **host approval of each RSVP:** it turns a pod into an application form;
  - **no waitlist:** a dropout then leaves a hole that only #lfg can fill.

**Where an RSVP comes from. Recommended: both places, one RSVP per person.**

- **Interested** on the Discord event (§5), or **Reserve a seat** on the site. A person is keyed by Discord id, which every account has.
- The **latest action wins**:
  - clicking Interested after Leave on the site re-reserves, at the back of the order;
  - removing Interested after reserving on the site leaves.
- Discord's own count of interested people can therefore differ from the seats. The event's description carries the real count (§4).
- The site can't mark anyone Interested, because no endpoint does. A person who reserved only on the site therefore gets no reminder from Discord. They get ours, if they opt in (§9).

**Holds.** At doors-open each person with a seat holds it until **10 minutes after the start** (the grace).

- **Taking the seat:**
  - **Take your seat** on the scheduled table's page, with no invite token. The hold is the authority, keyed by user or Discord id.
  - A member without the player role signs in with Discord, and the invite flow seats them as a guest (Q5).
- **For anyone else**, a held seat counts as taken:
  - a join through the invite link takes a free seat that nobody holds, or is refused with "The other seats are reserved until 19:40";
  - the #lfg post counts held seats as taken (`lfgTable` adds the holds to `Seated`).
- **The host can start early.** If seats are still held, the host confirms "2 reserved seats are still empty. Start without them?", and the unclaimed holds are dropped.
- **Leaving after doors-open** releases the hold at once. The seat goes down the waitlist as below.

**When the grace ends:**

1. Each unclaimed hold passes to the next person on the waitlist, who holds it for **10 more minutes**.
2. A seat nobody on the waitlist can take is **open**. The invite link and the #lfg post (if the host chose #lfg) can fill it.
3. If the host chose **fill with bots**, any seat still empty when the last waitlist hold ends gets a bot from the host's last setup (the default bot tier otherwise).

### 4. The Discord event

- **Entity type `EXTERNAL`** (gap 2). Every member can see it, and Discord starts and ends it on time by itself.
- **Name:** the table's name.
- **Location:** the scheduled table's page, `https://cmd.labxp.io/#/scheduled/{id}` (69 characters).
- **Description:** written by the server, with no free text from the host:
  > Commander · 4 players · hosted by «display name»
  > Seats: 1 of 3 open · 2 on the waitlist
  > Click Interested to reserve a seat. The first 3 get seats and the rest join the waitlist in order. Seats are held until 10 minutes after the start.
  > https://cmd.labxp.io/#/scheduled/{id}
- **Start and end:** the start, and the start plus the expected length. `privacy_level` is 2.
- **Kept in step by a schedule manager,** the third reconciler after voice and #lfg, with the same shape and failure rules:
  - a 429 waits;
  - a 5xx or timeout retries 3 times;
  - a 403 marks the plan's event `unavailable`, logged once. The scheduled table works without it, site RSVPs only.
  - Description edits are coalesced to one per event per 30 s.

| Plan or table event | The Discord event |
|---|---|
| scheduled | created (`POST /guilds/{guild}/scheduled-events`) |
| renamed, moved, seat counts changed | `PATCH`: name, times, description |
| cancelled before the start | `PATCH status=CANCELED` |
| the table's game ends, or it is archived or idle-closed after the start | `PATCH status=COMPLETED` if still active; otherwise Discord's own end time completes it |
| archived before the start | `PATCH status=CANCELED` |

- **The event id moves to the table** at doors-open (`games.discord_event_id`), so the game's end can complete the event after the plan's row is gone (§10).
- **Boot reconcile:** list the guild's events, `GET /guilds/{guild}/scheduled-events`. An event whose creator is the bot and that no live plan or table claims is cancelled. Events the bot didn't create are never touched.
- **Changes made in Discord.** A moderator with Manage Events can edit or delete our event. A **delete or cancel** cancels the plan, with a #mod-log line ("Scheduled table cancelled in Discord"). Any **other edit** is put back at the next reconcile, because the site is the source of truth (a call).

### 5. Interested, from the bot

- **The bot** adds `GUILD_SCHEDULED_EVENTS` when `CMDCTRL_DISCORD_SCHEDULED_EVENTS=1` and the community guild is set. The intent isn't privileged, so no portal switch is needed and there is no 4014 risk.
- **What it forwards.** For each `USER_ADD` and `USER_REMOVE` in the community guild, it calls `POST /admin/scheduled-events/{event_id}/rsvps` with `{discord_id, interested, at}`. `at` is the time the event reached the bot.
  - It calls three times while the server doesn't answer, as ban sync does.
  - An event the server doesn't own answers 404 and costs one loopback call, with no log line.
  - `UPDATE` and `DELETE` for our events go to `POST /admin/scheduled-events/{event_id}/discord-change`.
- **Order** is the order the server receives RSVPs from both sources, by `at`.
- **Reconcile over REST,** because the docs promise nothing about events missed across a failed resume. It runs every 5 minutes for each upcoming event, once at doors-open, and at boot. It pages through `GET …/users?after=` and compares the list with the Discord-sourced RSVPs.
  - Someone missing from our list is added **at the reconcile's time**, after everyone already in order. Ties are broken by user id, the only order REST has.
  - Someone gone from Discord leaves, unless a later site action of theirs is newer.
  - At most 50 upcoming events of a few pages each, every 5 minutes, is a few requests a minute.
- **Bans and leaves.** When ban sync (ADR 0156) reports a ban, a lost player role or a leave, that person's RSVPs are dropped too. A lost role doesn't drop the RSVP if Q5's answer lets members without the role hold seats as guests.
- **Production only.** The bot never runs from `develop`, and the server's bot token is provisioned on production only (ADR 0159 §7).

### 6. Voice room and #lfg

- **Voice room (ADR 0159):** if the host chose one, it is made at doors-open by the existing rule: the first seated Discord member, so the first person to take a held seat. Never at scheduling time: a room made days early would be removed by the 2-hour idle rule, and would take one of the 45 room slots meanwhile.
- **#lfg (ADR 0161):** if the host chose it, the post goes up **at doors-open** for seats that are neither taken nor held. It is edited when holds lapse at the end of the grace. A table whose seats are all held at doors-open posts nothing until a seat frees. Nothing goes to #lfg before doors-open: the Discord event is the announcement (Q8). The 30-minute ping cooldown applies as it does now.

### 7. Idle closing (ADR 0211)

- A plan is not a table, so the sweep never sees it, and a scheduled table can't be closed as idle before its start.
- At doors-open it is created like any table. Its 24-hour lobby clock and its 1-hour residency grace start then.
- **The amendment: a table nobody came to.** A scheduled table with **no seated person 1 hour after its start** is closed as `idle_lobby`, through ADR 0211 §4's close, and its event is completed. Without this rule it would hold load for 24 hours. In code it is one more branch in the sweep's deadline. Whichever of this work and ADR 0211's sub-PR 3 lands second adds it.

### 8. Time zones and display

- **Stored:** the start as a UTC instant, plus the host's IANA zone. A series needs the zone (slice 8): "Fridays at 19:30 New York" must stay 19:30 across a daylight-saving change.
- **Shown in the viewer's own zone everywhere:**
  - the site formats with the browser's zone and locale, and adds the host's time when the zones differ ("Fri 7:30 PM · 4:30 PM for the host, Los Angeles");
  - Discord shows event times in each member's zone itself;
  - DMs and bot replies use `<t:unix:F>` and `<t:unix:R>`, which Discord renders in the reader's zone.
- **No zone is stored on the account.** `/c2-schedule` defaults to the zone of the person's last plan, and that disappears with the plan (§10).

### 9. Reminders

- **Discord's own.** Discord may remind people who clicked Interested, but the developer docs don't say whether it does. The help-centre article couldn't be read (D1). The design doesn't depend on it.
- **Ours.** A new switch in Settings → Notifications, **Scheduled games**, off by default like the other kinds (ADR 0162). It is offered as a one-line nudge on Reserve. It sends:
  - "You have a seat at «FNM», <t:…:F>." (when reserved);
  - "You're off the waitlist for «FNM»." (when promoted, before or after doors-open);
  - "«FNM» opens now. Your seat is held until <t:…:t>. <link>" (at doors-open);
  - "«FNM» moved to <t:…:F>." and "«FNM» was cancelled.".
- The DM is sent whether or not the person is away, because a plan has no table to be away from. ADR 0162's limit of one DM per person per table per 10 minutes applies, keyed by the plan.
- People with no account have no settings, so they get no DM from us.

### 10. Cancelling, moving and repeating

- **Cancel:** the host, an admin, or a Discord delete (§4). People holding a seat or on the waitlist are told (§9). The plan stays listed as "Cancelled" for 24 hours and is then deleted.
- **Move:**
  - **When:** before doors-open, at most **3 times** per occurrence, and only to a time at least 30 minutes ahead.
  - **RSVPs are kept,** and everyone on the list is told. Anyone who can't make the new time leaves.
  - **Fewer players** moves the people whose seats are cut to the front of the waitlist, latest RSVP first.
- **After doors-open** it is a table. The host or an admin closes it with the existing archive, which completes the event (§4).
- **Recurring games** (slice 8). Recommended: **our own series, one Discord event per occurrence**.
  - A `scheduled_series` row holds the weekday, the local time, the zone and the interval (every week or every other week).
  - The next occurrence's plan and event are created when the current one opens. A series therefore shows one upcoming event, and its RSVPs belong to one game.
  - Rejected: Discord's `recurrence_rule`. The docs give no way to tell which occurrence an Interested is for, or to cancel one occurrence (D3).
  - **Regulars** (Q11): on a series, the site offers **Every week** beside Reserve. A standing RSVP is put in at the front of each new occurrence's order, before anyone's Interested, and kept until it is withdrawn. Two no-shows in a row drop it, and its holder is told.
- **Ending a series** cancels its upcoming occurrence.

### 11. Who sees what

- **The scheduled table's page** (`GET /scheduled/{id}`) is public by id, because it is the event's link and a member without an account must be able to open it. It shows the name, the host's display name, the time, the players, and the counts of open seats and the waitlist.
- **Names:**
  - The host, and the people on the list, see the names in order: seats, then the waitlist.
  - Nobody else sees names on the site. That is ADR 0161's rule: the host chose to advertise, the others did not.
  - The Discord Interested list is Discord's, shown to the server's members by Discord.
- **Where names come from:** an account's display name, or, for a person with no account, their server nickname read from Discord when the page is opened. It is kept in memory for 5 minutes and never written down.
- **The Scheduled list** (`GET /scheduled`) is for signed-in people. It shows upcoming plans with their counts, and your own plans and reservations first.

### 12. Abuse limits

| Limit | Value | Why |
|---|---|---|
| Upcoming plans per host | **3** (a series counts once: it has one upcoming occurrence) | the same as open lobby tables |
| How far ahead | at most **30 days**, at least **30 minutes** | no calendar squatting; doors-open must be in the future |
| Moves | **3** per occurrence | a time that keeps moving is spam to everyone on the list |
| Creates and moves | the create bucket: 1 per 30 s, burst 3 | reuses `limit` in `createGameWith` |
| Upcoming plans server-wide | **50** | Discord allows 100 scheduled or active events per guild, and the server's own events share them. Past it: "The calendar is full" |
| Free text | none: the name only, filtered | no new surface to moderate |

- **Moderation:**
  - an admin can cancel any plan (`DELETE /scheduled/{id}`, with a #mod-log line);
  - a mod can delete the event in Discord (§4);
  - a reported table is handled as today once it is open.
- **`/c2-schedule`** is held to the same limits, as the named account (gap 4).

## Privacy and deletion (ADR 0149)

| Data | Where | Kept | Deleting the account |
|---|---|---|---|
| Plans | `scheduled_tables(id, series_id, host_user_id, name, starts_at, length_minutes, zone, players, options, status, moves, discord_event_id, game_id, created_at, updated_at)` | until **24 hours after the start**, or 24 hours after a cancel | the host's upcoming plans are **cancelled** (Q14), their events cancelled after the commit, and the rows **deleted** |
| RSVPs | `scheduled_rsvps(schedule_id, discord_id, user_id, source, state, rsvp_at, updated_at)`. `user_id` is NULL for a person with no account | with the plan: 24 hours after the start | the person's rows, by user id **and** Discord id, **deleted** |
| Series | `scheduled_series(id, host_user_id, name, weekday, local_time, zone, every_weeks, options, created_at, ended_at)` | until the host ends it | **deleted**, which ends it |
| Standing RSVPs | `series_regulars(series_id, user_id, missed, created_at)` | until withdrawn, two no-shows, or the series ends | **deleted** |
| The table's event id | `games.discord_event_id` | with the table | not personal |
| Nicknames of people with no account | memory | 5 minutes | not stored |
| The notification switch | `notification_prefs.scheduled` | as the other switches | deleted with the row, as today |

- **People with no account.** A member who clicks Interested on a C2 event without an account has their Discord id and RSVP time stored. Few rows today are keyed by the Discord id of someone who may never have signed in: deck requests made with `/c2-deck-req` (`deck_request_asks`), a named host (`games.host_discord_id`) and a pending seat (`seats.pending_discord_id`). This adds a fourth, and the first one a single click creates.
  - It is kept only as long as the plan's 24-hour window.
  - Removing Interested, leaving the server, or a ban deletes it sooner.
  - The privacy page says so in a sentence anyone in the server can read before clicking.
- **Logs** carry plan ids and counts, never a Discord id. Metrics carry counts.
- **`accountdelete.Inventory`** lists every new column. `Delete` removes the person's rows from each table, and cancels their hosted plans, in its transaction. The Discord cancels run after the commit, best effort and logged, as the avatar files are. The audit line gains `scheduled_tables`, `scheduled_rsvps`, `scheduled_series` and `series_regulars`.
- **Tests:** `TestInventoryCoversTheSchema` and `TestDeleteRemovesOrAnonymisesEveryTable` cover the new tables, with rows that name the person by user id and by Discord id. `PRIVACY_SCHEMA_VERSION` takes each migration's number.
- **The privacy page** (`routes/Privacy.svelte`, numbers in `lib/privacy.ts`), in the slice that ships each part:
  - **Slice 2, "Scheduled tables", under "What is stored, and why":** "When you schedule a table or reserve a seat, we keep the plan and the list of people with seats or on the waitlist, in the order they reserved, until 24 hours after the start. The host and the people on the list see the names on it; everyone else sees how many seats are open." New number: `scheduledKeepHours`.
  - **Slice 5, a paragraph in the Discord row:** "If you click Interested on a CMD & CTRL game in the Discord server, we keep your Discord id and when you clicked, to hold your seat, even if you have no account. It is deleted when you remove Interested, leave the server, or 24 hours after the game's start. The Interested list itself is Discord's, and every member of the server can see it."
  - **Slice 8:** standing RSVPs and series, one sentence each.

## Configuration and the owner's setup

- **Slices 2 and 3** have no new variable. Scheduled tables need accounts, so they are on wherever there is a database, `develop` included. That is where the RSVP and hold logic is tested.
- **Slice 4** adds `CMDCTRL_DISCORD_SCHEDULED_EVENTS=1`, read by the server and, from slice 5, by the bot.
  - It is on when it, `CMDCTRL_DISCORD_COMMUNITY_GUILD_ID` and the bot token are all set.
  - The flag without the guild fails the boot, naming the variable, as `VoiceFromEnv` does. Without the token the feature is off, said once at boot.
  - `GET /community` gains `"scheduled_events": true|false`.
  - CD writes the flag from an Actions variable, always, empty included, as it writes ban sync's.
- **Owner setup (once):**
  1. C2 server → Server Settings → Roles → `cmd_and_ctrl`: turn on **Create Events**. Manage Events isn't needed: the bot edits only its own events. Nothing is re-authorized (gap 1).
  2. Set the Actions variable `CMDCTRL_DISCORD_SCHEDULED_EVENTS` to `1`, then deploy `main`.
  3. Check `journalctl -u cmd-and-ctrl-bot` for "scheduled events on", and schedule one test table. The token exists only on production, so this is the first live test. Cancel the test at once: an external event is visible to every member.

## Calls made here

Each of these is overturnable in review:

- **`EXTERNAL` events**, with the scheduled table's page as the location (gap 2).
- **Doors open 15 minutes before the start.** Holds last until **10 minutes after** it. A promoted waitlist hold lasts **10 more minutes**.
- **The host holds seat 1**, and is not seated automatically at doors-open.
- **Seat order is by arrival at the server.** A REST reconcile appends what it finds after everyone already in order, with ties by user id.
- **The site is the source of truth.** Discord-side edits are put back, except a delete or cancel, which cancels the plan.
- **No free text** in an event or a plan.
- **Description edits** at most one per event per 30 s. **Reconcile** every 5 minutes.
- **A table nobody came to closes 1 hour after its start** (§7).
- **Opening skips the open-table cap and the create rate limit**, but not the load cap.
- **Retention:** a plan and its RSVPs are kept 24 hours after the start.
- **The host's removal is per occurrence.** ADR 0213's personal block applies silently.
- **Names on the site** go to the host and the people on the list only.
- **No zone is stored on the account.**
- **Recurring games use our own series**, not Discord's recurrence rule.

## Consequences

- **A pod can book a game ahead,** in Discord or on the site, and C2 keeps the seats for the people who said yes.
- **Nothing exists in the engine until 15 minutes before the start.** Plans cost no load, survive restarts as rows, and need no exemption from idle closing.
- **Scheduled tables are a C2 feature, with the Discord event as a mirror,** as the #lfg post is. They work on `develop` and without the bot. In the Discord parts, only the Interested intake needs the bot process.
- **Discord's "interested" count is not the seat count.** Site RSVPs don't appear there. The description carries the real numbers.
- **Some data is stored about people with no account,** briefly (privacy section).
- **New surface:** four tables and a column; `/scheduled` routes and two admin routes; one bot intent and one slash command; one schedule manager; a fourth DM kind; one variable; one role permission for the owner.

## Back-out

| To undo | Do |
|---|---|
| Slice 4 or 5 (Discord) | unset `CMDCTRL_DISCORD_SCHEDULED_EVENTS` and deploy. The bot stops asking for the intent and the server stops calling Discord. Events already made are left for Discord to end on time, or a mod deletes them. Plans keep working with site RSVPs. |
| Slices 2 and 3 | drop the routes and the scheduler. Tables already opened are ordinary tables. The tables can be dropped by hand after rolling the binary back (`ErrSchemaTooNew`; docs/environments.md). |
| Slice 8 | stop creating occurrences. The series rows are inert. |

## Slices

| # | Slice | Schema | Server and wire | Client | Privacy | Model | Needs |
|---|---|---|---|---|---|---|---|
| 1 | This ADR | — | — | — | — | — | — |
| 2 | **Scheduled tables with site RSVPs** | `scheduled_tables`, `scheduled_rsvps` | `POST`/`GET /scheduled`, `GET`/`PATCH`/`DELETE /scheduled/{id}`, `PUT`/`DELETE /scheduled/{id}/rsvp`, `DELETE /scheduled/{id}/rsvps/{discord_id}` (host), `POST /scheduled/{id}/claim`; `Lobby.RunSchedule` (doors-open, holds, grace, waitlist, bots, missed, retention); seat holds in join, `lfgTable` and Start's confirm; caps; ADR 0211's no-show rule if 0211 has landed; docs/lobby.md | — | "Scheduled tables", `scheduledKeepHours`, `Inventory`, `Delete`, the schema bump | Opus | 1 |
| 3 | **The site** | — | — | Schedule a table, the Scheduled section, the scheduled table's page (Reserve, Leave, Take your seat, the host's list and Remove), zones, labels and e2e; screenshots at 1440 px and 820 px | — | Sonnet | 2 |
| 4 | **The Discord event** | `games.discord_event_id` | `discord.Events` (create, modify, list, users) with httptest coverage; the schedule manager and its boot reconcile; config and boot checks; `GET /community`; CD; docs/env.md, docs/discord-bot.md (owner setup) | the event's status on the host's page | — | Sonnet | 2 |
| 5 | **Interested is an RSVP** | — | the bot's intent and forwarding; `POST /admin/scheduled-events/{event_id}/rsvps` and `…/discord-change`; the REST reconcile; Planeswalker holds through the invite flow (Q5); ban sync drops RSVPs | "Reserved in Discord" on rows | the Discord paragraph | Opus | 4 |
| 6 | **Reminders** | `notification_prefs.scheduled` | the fourth DM kind, its messages with `<t:>` | the switch; the nudge on Reserve | the notifications line; the schema bump | Sonnet | 2 |
| 7 | **`/c2-schedule`** | — | create, list and cancel, with autocomplete; the server's named-host cap | — | — | Sonnet | 4 |
| 8 | **Series and regulars** | `scheduled_series`, `series_regulars` | create the next occurrence at doors-open; DST by zone; standing RSVPs, two-miss drop | Repeat on the form; Every week | two sentences; `Inventory`, `Delete` | Opus | 4 |
| 9 | *Optional:* **friends** | — | hold a seat for a friend at scheduling (an in-app invite to confirm, ADR 0213 §8), and Invite a friend on a plan | the picker on the form and page | — | Sonnet | 3, ADR 0213 slice 3 |

Slices 4 and 6 are independent of each other. Slice 2 is the smallest useful one: a table you can book, share by link and fill in order. It needs nothing from Discord, so it can be proven on `develop`. Slice 5 changes the bot and runs on production only, so it ships with a manual check (owner setup step 3). Client slices attach screenshots at 1440 px and 820 px.

## What the docs do not settle

- **D1.** Whether Discord notifies Interested users when an event starts, or before. The developer docs don't say, and the help-centre article ([Scheduled Events](https://support.discord.com/hc/en-us/articles/4409494125719-Scheduled-Events)) answered 403 to every fetch. §9 doesn't depend on it.
- **D2.** Whether a link in an external event's location or description is clickable. If it isn't, the description still carries it as text. Slice 4 checks it on production.
- **D3.** How Interested relates to the occurrences of a recurring event. No occurrence or exception field appears in the docs. §10 avoids the question.
- **D4.** Whether `USER_ADD` and `USER_REMOVE` are delivered reliably. Nothing is stated beyond "the Gateway will send the missed events in order" after a successful resume. §5's reconcile covers the rest.
- **D5.** What members see of a `CANCELED` event, and whether Discord tells its subscribers. Slice 4 checks it. Our own DM covers people who opted in.
- **D6.** Whether a user mention renders in an event description. §4 uses the host's display name instead.
- **D7.** Whether an active external event's end time can be pushed back. Not used: a game that runs long outlives its event, which is harmless.
- **D8.** The recurrence section numbers `WEEKLY` and `DAILY` inconsistently (the text says 0 and 1, the table and examples 2 and 3). Not used.

## Owner decisions

Open questions, each with the recommended option first, (a). The ADR is written for (a) throughout.

1. **Before the start:** (a) a plan that becomes an ordinary lobby table when doors open, 15 minutes before; (b) a lobby table from the moment it is scheduled, with a start time; (c) a new engine state.
2. **Who may schedule:** (a) anyone who can create a table (an account that passes the gate), within §12's limits; (b) only accounts that have finished a game; (c) mods only.
3. **From where:** (a) the site first, then `/c2-schedule` (slice 7); (b) the site only; (c) Discord only.
4. **How RSVPs become seats:** (a) the first N, then a waitlist in order with automatic promotion, and the host can remove a reservation; (b) host approval of each; (c) first N with no waitlist.
5. **Members without the player role** (Planeswalker) who click Interested: (a) they hold a seat like anyone, and take it as a guest through the Discord sign-in, as with a host's invite link (ADR 0156) and an Activity (ADR 0214 question 6); (b) their Interested is ignored, and the description says an account is needed; (c) waitlisted until they have the role. Guests with no Discord at all can't RSVP under any option. Seats that open after the grace are open to them by invite link, as today.
6. **Where an RSVP comes from:** (a) Interested in Discord or Reserve on the site, one RSVP per person, latest action wins; (b) Interested only, and the site links to the event; (c) the site only, with the event as an announcement.
7. **No-shows:** (a) holds last until 10 minutes after the start, then pass down the waitlist (10 minutes each), then open to the link and #lfg, with an optional host setting to fill with bots; (b) holds last until the host starts; (c) no hold past the start.
8. **#lfg:** (a) at doors-open, for seats neither taken nor held, updated as holds lapse; nothing before doors-open; (b) also a post when the table is scheduled; (c) only after the grace.
9. **Voice room:** (a) at doors-open, by ADR 0159's first-seat rule; (b) at scheduling time; (c) at the start.
10. **Reminders:** (a) a fourth DM switch, Scheduled games, off by default and offered on Reserve, plus whatever Discord sends Interested users; (b) the same messages, on by default for anyone who reserves on the site; (c) Discord's own only.
11. **Recurring games:** (a) our own weekly or every-other-week series, one Discord event per occurrence, with an **Every week** standing RSVP for regulars (two no-shows drop it); (b) the same without standing RSVPs; (c) Discord's recurrence rule on one event; (d) no recurrence.
12. **Moving a game:** (a) RSVPs are kept and everyone is told, at most 3 moves, never after doors-open; (b) a move clears the list.
13. **Limits:** (a) 3 upcoming plans per host, 30 days ahead at most, 30 minutes at least, 50 server-wide; (b) other numbers.
14. **Deleting an account that hosts upcoming plans:** (a) the deletion cancels them, the confirmation lists them, and the people on their lists are told; (b) refuse until the host cancels them, as for a table in progress (ADR 0149 §4).
15. **Names on the site:** (a) the host and the people on the list see who holds which seat, and everyone else sees counts; (b) every signed-in person sees the names, as every server member can on Discord.
16. **Friends** (slice 9): (a) the host may hold seats for friends at scheduling, each confirmed by an in-app invite, ahead of the Interested order; (b) invite a friend to the page only, and they reserve like anyone; (c) neither.

## Sources

Read on 2026-10-30. The `discord.com/developers/docs/...` addresses redirect to these.

- [Guild Scheduled Event](https://docs.discord.com/developers/resources/guild-scheduled-event): the object, entity types, field requirements by type, privacy level, status and its automation, permissions requirements, the create, modify, delete, list and users endpoints, the 100-event cap, the recurrence rule.
- [Gateway events](https://docs.discord.com/developers/events/gateway-events): Guild Scheduled Event Create, Update, Delete, User Add and User Remove; Guild Create's `guild_scheduled_events`.
- [Gateway](https://docs.discord.com/developers/events/gateway): the list of intents, privileged intents, resuming.
- [Permissions](https://docs.discord.com/developers/topics/permissions#permissions-bitwise-permission-flags): `CREATE_EVENTS`, `MANAGE_EVENTS`.
- [Change log](https://docs.discord.com/developers/change-log): 2025-11-20, "Guild Expressions and Events Permissions now available to developers"; 2025-11-24, the February 2026 permission changes.
- [Rate limits](https://docs.discord.com/developers/topics/rate-limits) and [OAuth2](https://docs.discord.com/developers/topics/oauth2) (bot authorization: no events scope; permissions are an integer).
- [Message formatting](https://docs.discord.com/developers/reference#message-formatting): timestamp styles.
- Not read (403): [Discord help centre, Scheduled Events](https://support.discord.com/hc/en-us/articles/4409494125719-Scheduled-Events).
