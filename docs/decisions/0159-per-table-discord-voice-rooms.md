# ADR 0159 — Per-table Discord voice rooms

**Status:** Accepted · 2026-10-13 · Backlog
**Issue:** [#3223](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3223)
**Numbering:** every ADR file on `origin/develop` and on every remote head was listed. The highest number anywhere was **0158**, so this is **0159**.
**Builds on:** [ADR 0004](0004-discord-identity.md) (Discord identity), [ADR 0051](0051-user-database.md) decision 5 (the server's bot-token REST calls, `server/internal/discord/dm.go`), [ADR 0075](0075-table-settings-and-host-controls.md) §2.1 (the table host), [ADR 0148](0148-capacity-limits-for-a-public-launch.md) (capacity caps), [ADR 0149](0149-privacy-notice-and-account-deletion.md) (the privacy notice), [ADR 0156](0156-a-front-door-site-shell-front-page-and-a-discord-gate.md) §1 (the community server, `CMDCTRL_DISCORD_COMMUNITY_GUILD_ID`) and §6 (`GET /community`).

---

## Context

Commander is a talking game, and today a pod that wants voice has to arrange it outside C2: someone makes a Discord call or a private channel by hand and shares it. The CMD & CTRL Discord server (ADR 0156) now exists, and every player with an account signed in through Discord, so the server already knows each seat's Discord user id (`SeatInfo.DiscordID`).

The owner wants a table to be able to get its own voice channel in that server: private to the table's players, created for them, and gone after the game.

What already exists:

- **The server makes bot-token REST calls itself.** DM invites (`discord.Bot`, `dm.go`) open a channel and post a message with `CMDCTRL_DISCORD_BOT_TOKEN`, from the server process, without the gateway bot. Creating, editing and deleting a channel are the same kind of call.
- **The lobby sees every lifecycle edge.** `syncStateLocked` records start and end, `watchEnd` rides `Room.Subscribe` until a game ends, and `persistGameLocked` writes the `games` row.
- **Tables are never closed automatically.** A lobby or active table lives until it ends or an admin archives or deletes it. Anything tied to a table needs its own rule for the table that is simply left.

Discord's limits that matter: a category holds at most **50 channels**; a channel name is 1–100 characters; a permission overwrite names a role or a user id; deleting a voice channel disconnects everyone in it and deletes its text chat.

## Decision

### 1. Opt-in per table, switched by the host

Owner answers, 2026-10-13.

- **Create:** `POST /games` takes `"voice_room": true`. The create-table form shows a **Create a voice room** checkbox, off by default, only when the feature is available (item 7).
- **Switch later:** `PUT /games/{id}/voice` with `{"enabled": true|false}`. **Host only** (ADR 0075 §2.1), plus the server admin. Turning it off deletes the room at once (§4). Allowed in lobby and active states, refused once the game has ended.
- The choice is lobby metadata, **not** a `game.TableSettings` key: it is not a rule of the game, it must not enter the engine snapshot, and it needs no snapshot schema bump.

### 2. When the room exists

- **Created at the first Discord seat.** When voice is on and the first seat with a `DiscordID` is filled (in lobby state, so the pod can talk while choosing decks). If voice is switched on later, it is created then, if such a seat exists.
- **Bot seats, guests and spectators** never count and are never given access. A guest has no Discord id, so there is nothing to grant. Spectators are excluded in this first version.
- **Deleted:**
  - **10 minutes after the game ends** (owner answer), for the post-game chat.
  - **At once** when the host switches voice off, or the table is archived or deleted, or `closeUnresidentRows` closes its row at boot.
  - **When idle:** no game action and no seat change for **2 hours**, and **no seated player connected to the room** (owner answer). Activity is the room's sequence number (`Room.Seq`) moving for any reason other than the voice manager's own pushes, or a seat change; a periodic sweep compares it. Connection is checked per seated Discord id with `GET /guilds/{guild}/voice-states/{user}`. The table keeps voice on and shows `idle`; `voice/refresh` from any seated player, the host's toggle or a seat change creates a fresh room.
- **No archive.** Discord cannot archive a voice channel; its text chat goes with it. The game log and replay remain the record of the game.

### 3. Who gets in

The room is created in the configured category with these overwrites:

| Target | Allow | Deny |
|---|---|---|
| `@everyone` | — | View Channel, Connect |
| each seated Discord user who is a **member** of the community server | View Channel, Connect, Speak, Video, Use Voice Activity, Send Messages, Read Message History | — |
| the bot's own user | View Channel, Connect | — |

Managing the room (Manage Channels, Manage Permissions, Move Members) comes from the **bot's role, server-wide**, not from an overwrite: Discord refuses an overwrite that grants Manage Permissions to anyone but an administrator. The bot's overwrite only keeps the room visible to it despite the `@everyone` deny.

- **Membership first.** An overwrite does not invite anyone to the server. Before granting, the manager checks `GET /guilds/{guild}/members/{user}` (a REST call; no privileged gateway intent). A 404 is "not in the server": no overwrite, and the table tells that player to join (§6).
- **Seat changes.** A Discord seat filled → check and grant. A seat left, removed or reclaimed by someone else → delete that user's overwrite, and if `voice-states` shows them connected **to this room**, disconnect them (`PATCH /guilds/{guild}/members/{user}` with `channel_id: null`). The room id is compared first, so a player is never pulled out of a different channel.
- **Joined the server later.** `POST /games/{id}/voice/refresh` (seated players only, rate-limited per caller) re-runs the member check for the caller, and reopens an `idle` room. The table's "I've joined — check again" and "Reopen voice room" buttons call it.
- **Server staff.** Discord lets the server's owner and administrators see every channel, private or not. The privacy notice says so (§8).

### 4. How the server talks to Discord

- **A new `discord.Voice`** beside `discord.Bot`, using the same token, the same `post`-style helper and the same rule that the token never reaches a log or an error. Calls: create channel, edit and delete an overwrite, delete channel, list the guild's channels, get a member, get a member's voice state, move a member out of voice.
- **A per-lobby voice manager** owns all of it. Lifecycle hooks enqueue work (`ensure`, `grant`, `revoke`, `delete`) keyed by game id; one worker goroutine performs it, **never holding `l.mu` across a Discord call**. A game start, a seat claim or an end is never blocked or slowed by Discord.
- **Failures.** A 429 waits its `retry_after` (capped) and retries; a 5xx or timeout retries with backoff up to 3 times, then marks the table's room `unavailable` and logs at WARN. The table plays exactly as it would without voice. The next lifecycle event, or the host's toggle, tries again.
- **Names.** The channel is named after the table: the lobby's validated table name, trimmed to 100 characters. A rename of the table does not rename the room in this version.
- **Cap.** At most **45 live rooms**, under the category's 50-channel limit with room for the owner's own channels. Past it, `ensure` marks the table `full` and the table says so. This is a voice cap only; ADR 0148's table caps are unchanged.

### 5. Durability

- **Migration** (the next free number when slice 2 lands): `games.voice_room INTEGER NOT NULL DEFAULT 0` (the opt-in) and `games.voice_channel_id TEXT` (the live room, or NULL). Both written through `persistGameLocked`.
- **Boot reconcile**, after the restore pass: list the category's channels once.
  - A channel that belongs to a resident game with voice on is kept; its overwrites are re-synced to the current seats.
  - Any other channel **in the category** is deleted. The category belongs to C2: the owner must not keep their own channels in it (the ops note says so).
  - A resident game with voice on and no surviving channel gets one if it has a Discord seat, per §2.
- An ended game restored inside its 10-minute window has its room deleted at the window's original end, computed from `ended_at`.

### 6. What the table shows

At the table the client reads the WebSocket game view, not the lobby meta, so the voice state rides the view the way the host and the playmats do (`ws/host.go`, `ws/playmat.go`): the lobby sets it on the room (`Room.SetVoice`), every capture stamps it, and nothing enters the engine or its snapshot. When the voice manager changes it, the lobby pushes a fresh view (`ApplyExternal` and `BroadcastState`), as account deletion does. `protocol.FilterViewFor` reduces it to what one viewer may see:

```json
{"enabled": true, "status": "ready", "you": "granted", "url": "https://discord.com/channels/<guild>/<channel>"}
```

The lobby meta carries only `voice_room: true|false`, so a table list can show which tables have voice.

- `status`: `off`, `pending` (voice on, no room yet), `ready`, `idle` (removed for idleness; any seated player can reopen it with `voice/refresh`), `full`, `unavailable`.
- `you`, for this viewer: `granted`, `not_in_server`, `no_discord` (a guest), `spectator`.
- `url` only when `you` is `granted`. The channel id is not a secret, since Discord enforces access, but the link is only useful to the people it lets in.

The client:

- **Join voice** in the table header when `granted`. It opens the URL (the Discord app or web client).
- **Join the Discord to use voice** with the community invite (`GET /community`) when `not_in_server`, and **I've joined, check again**.
- **Sign in with Discord to use voice** for a guest. Nothing for a spectator.
- The host's toggle sits with the table's other host controls; `full` and `unavailable` show a one-line reason.

### 7. Configuration

- **New:** `CMDCTRL_DISCORD_VOICE_CATEGORY_ID`, the category in the community server that rooms are created in.
- **On** when it, `CMDCTRL_DISCORD_COMMUNITY_GUILD_ID` and `CMDCTRL_DISCORD_BOT_TOKEN` are all set.
  - The category set without the community guild, or a value that is not a snowflake, **fails the boot** naming the variable, as in `GateFromEnv`: rooms in a guild nobody named are a misconfiguration.
  - The category set without the bot token leaves the feature **off**, said once at boot. The token is provisioned on production only, so a box that inherits a repo-level category variable must still boot.
  - Off: no checkbox, no `voice` object, every route unchanged, and `PUT /games/{id}/voice` answers 503 naming what is missing.
- `GET /community` gains `"voice": true|false`, so the client knows whether to show the checkbox.
- **CD** writes the category from an Actions variable of the same name. The bot token is provisioned on **production only**, so voice rooms are production-only, for the same reason DM invites are: one Discord application, one set of rooms.

### 8. Privacy

The notice (ADR 0149) gains one short paragraph: when a host turns on a voice room, CMD & CTRL asks Discord to create a private channel named after the table and lets in the seated players who are server members, by their Discord ids, which the site already holds. The channel and its text chat are deleted after the game. The server's owner and administrators can see private channels. No new data is stored beyond the two `games` columns.

## Consequences

- **A pod gets voice in one click,** but only for players who signed in with Discord **and** joined the server. Guests and non-members see why and what to do.
- **Discord's own moderation applies inside the room.** The host's removal of a seat also removes that player from the room; anything else (muting, reporting) is Discord's.
- **The text chat in a room is not kept.** Players who want something kept should put it in the server's public channels.
- **The category is C2's.** A channel the owner puts in it by hand is deleted at the next boot.
- **Voice rooms depend on the bot being in the community server** with the permissions in §3. That is an owner step, not code (below).

### Owner setup (once, not code)

1. Invite the C2 Discord application's bot to the CMD & CTRL server (Developer Portal → OAuth2 → URL Generator, scope `bot`).
2. Give the bot's role, **server-wide** (Server Settings → Roles): View Channels, Manage Channels, Manage Roles, Move Members, Connect, Speak, Video, Use Voice Activity, Send Messages and Read Message History. Discord lets a bot grant only permissions it holds, and a channel overwrite cannot give it Manage Permissions (§3). Keep the bot's role below the moderators' roles.
3. Create a category, for example **Table voice**, at the bottom of the channel list, and keep nothing else in it.
4. Set the Actions variable `CMDCTRL_DISCORD_VOICE_CATEGORY_ID` to the category's id, and `CMDCTRL_DISCORD_COMMUNITY_GUILD_ID` if slice 2 of ADR 0156 has not already.

## Slices

| # | Slice | Model | Needs |
|---|---|---|---|
| 1 | This ADR | — | — |
| 2 | **Server: the Discord calls** (§4 `discord.Voice`), config and boot checks (§7), `GET /community`'s `voice` flag; httptest coverage of every call and error mapping; AGENTS.md env vars | Sonnet | 1 |
| 3 | **Server: the voice manager** (§1, §2, §3, §5): migration, create flag and `PUT /games/{id}/voice`, the queue and worker, lifecycle hooks, idle rule, cap, boot reconcile, `voice/refresh`, the meta's `voice` object; privacy notice paragraph; docs/lobby.md; CD variable sync | Opus | 2 |
| 4 | **Client** (§6): the create checkbox, the host toggle, Join voice and the not-in-server and guest states; screenshots at 1440 px and 820 px | Sonnet | 3 |
