# ADR 0213 — Friends, recently played with, and live presence

**Status:** Proposed · 2026-10-30 · S77 — Discord and social: voice, friends, mod tools and events
**Issue:** [#3313](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3313)
**Owner decisions:** none yet. The brief is the owner's request of 2026-10-13, quoted in the issue body. Every product choice below is a recommendation. Each one is listed under [Open questions](#open-questions-for-the-owner), recommended option first. The smaller calls this ADR makes on its own are listed under [Calls made here](#calls-made-here). Building starts only after the owner answers.
**Numbering:** checked on 2026-10-30. `develop` ends at 0211. `git ls-tree` over every remote head's `docs/decisions/` shows one file above it: `docs/lobby-character-select` (PR #4656) holds 0212 for the lobby card redesign. This was drafted as 0212 and renumbered before it was pushed. The heads checked are listed in the PR. This is **0213**.
**Builds on:**
- [ADR 0051](0051-user-database.md): decision 5 (DM invites sent with the bot token) and decision 8 (tablemates are a query, not a table).
- [ADR 0110](0110-remember-me.md) §5: the last setup, which remembers the last table's tablemates.
- [ADR 0149](0149-privacy-notice-and-account-deletion.md): the privacy notice, `accountdelete.Inventory` and `Delete`.
- [ADR 0156](0156-a-front-door-site-shell-front-page-and-a-discord-gate.md): the community server, the gate, the mod log, and ban sync's bot-to-server admin route.
- [ADR 0158](0158-signed-in-pages-tables-with-history-decks-and-the-catalog.md): Tables with Now and History, and its 2026-10-14 amendment (Invite by DM in the card toolbar).
- [ADR 0159](0159-per-table-discord-voice-rooms.md) (voice rooms), [ADR 0161](0161-an-lfg-board-in-discord.md) (#lfg), [ADR 0162](0162-discord-dm-notifications.md) (DM notifications and the `visibility` frame).
- [ADR 0148](0148-capacity-limits-for-a-public-launch.md) (capacity caps) and [ADR 0124](0124-admin-views-accounts-games-and-who-is-on-now.md) §3.4 (Live now).

**Amends, if accepted:**
- ADR 0051 decision 8 said "no friend requests, no acceptance, no `friendships` table" until "the group is not eight people who already know each other". The public beta is that group. The query stays, and friends are "one table on top of this", as decision 8 foresaw.
- ADR 0158 §1 item 3: the History filter **With friends** is renamed (§9).

---

## Context

On 2026-10-13 the owner asked for a **friends** list and a **recently played with** list. Each person on them should show where they are now:

- in a CMD & CTRL game, with a link to watch or join when that is allowed;
- on CMD & CTRL but not at a table;
- online on Discord but somewhere else;
- offline.

### What exists

- **"Recently played with" is already built, as tablemates.** `GET /me/tablemates` (`lobby/tablemates.go`) is ADR 0051 decision 8's self-join of `seats`. It lists every account the caller has shared a table with, ordered by the most recent shared table (`games.created_at`).
  - Guests and bots hold seats with no `user_id`, so they never appear. Agent seats join as guests by rule (ADR 0122), so they never appear either. A stand-in bot (ADR 0177) keeps the person's `user_id`, so that person still counts.
  - Its only consumer is `TablematePicker.svelte`. It is the **Invite by DM** panel in a Now card's toolbar. It asks the server to DM the table's invite link through the bot (`POST /games/{id}/invites/dm`).
  - When a table's seats change, its `seats` rows are replaced wholesale (`SQLStore.ReplaceSeats`). A seat row is therefore the seat's last occupant, and someone who sat in a lobby and left is not remembered.
- **The server knows who is connected to each table.** The hub keeps every socket's game, seat, user and spectator flag. Since ADR 0162 it also keeps whether the socket's tab is hidden (`ws.LiveSocket`, `Hub.LiveSockets`, `Hub.SeatAway` in `ws/presence.go`). Admins already read this in Live now.
- **It does not know who is on the site but away from a table.** The only WebSocket is the per-game `/ws`. The Tables page has none: it polls `GET /games` every 5 s while a Now card exists, and again on focus (`Lobby.svelte`, `POLL_MS`). `users.last_seen_at` is written only at sign-in (`users/metrics.go`).
- **Which tables people sit at is already visible to every session.** `GET /games` gives any session every active, unarchived table: its name, state, turn and seat names (`listGames`). `redactMetaFor` strips only the invite tokens and the creator. `/c2-games` lists the same tables in Discord. Spectators' names are shown only to that table's players (the privacy page).
- **The bot reads no presence.** It always asks Discord for Guilds. It adds Guild Messages, Message Content, Guild Moderation and Server Members only when the feature that needs one is on (`bot.Config.Intents`). It has no HTTP listener of its own: it calls the server's admin routes with its admin session, as ban sync does (`POST /admin/discord-members/{discord_id}/revoke-sessions`).
- **The server sends DMs with the bot token** (`discord.Bot.SendDM`), on production only.
- **Blocks are per table** (`table_blocks`, ADR 0075 §2.6). They are the host's remove-and-block. One person cannot block another.

### Gaps found while reading

1. **Deleted accounts stay in tablemates.** The query joins `users` with no `deleted_at IS NULL` filter. A tombstone (ADR 0149 §2) is therefore offered in the picker as "Deleted player", and inviting it answers 422 because it has no identity.
2. **The DM invite route accepts any account id.** `dmRecipient` takes any `user_id`, not only a tablemate's. Today account ids leave the server only through `/me/tablemates`, so the gap is latent. Friends and presence would hand out more of them.
3. **The privacy page leaves out the tablemate picker and its DMs.** It does not say that your name and avatar appear in the invite picker of everyone you have sat with, or that those people can have the bot DM you an invite. "Games you played in" and "The other people at your table" cover the table, not the list.
4. **"With friends" already means something else.** ADR 0158's History filter means "another human was at the table" (`lib/tablesPage.ts`).
5. **Avatar paths carry the Discord id** (`/avatars/<snowflake>/<hash>.png`, `users/users.go`). So "a snowflake is never on the wire" (docs/lobby.md, tablemates) holds for the id fields only. Every seat's avatar already works this way. This ADR adds no new exposure and does not change it.

### What Discord allows

- **A web app cannot read a user's Discord friends.** The OAuth2 scope `relationships.read` "allows your app to access a user's Discord Friends list, their pending requests, and blocked users". The scope table says it is part of the Social SDK, needs an access request, and falls under the Social SDK Terms ([OAuth2 scopes](https://docs.discord.com/developers/topics/oauth2#shared-resources-oauth2-scopes)).
  - The Social SDK ships for Windows, macOS, Linux, Android, iOS and consoles, as standalone C++, Unity and Unreal. It lists no web platform ([platform compatibility](https://docs.discord.com/developers/discord-social-sdk/core-concepts/platform-compatibility)).
  - The Embedded App SDK, which a Discord Activity (#3312) would use, has `getRelationships()`, behind the same `relationships.read` and the same access request ([Embedded App SDK](https://docs.discord.com/developers/developer-tools/embedded-app-sdk)).
  - So Discord friends are out of reach for the site, and an Activity could read them only after Discord approves.
  - `activities.read` ("Now Playing/Recently Played") is "not currently available for apps" (same scope table).
- **Presence reaches a bot over the gateway only, and only for members of a server the bot is in.**
  - `PRESENCE_UPDATE` needs the `GUILD_PRESENCES` intent ([Presence Update](https://docs.discord.com/developers/events/gateway-events#presence-update)). Guild Create carries the presences of non-offline members ([Guild Create](https://docs.discord.com/developers/events/gateway-events#guild-create)). The guild member object that REST returns has no status.
  - The intent is privileged. "Apps with fewer than 10,000 users can access privileged intents by enabling them in the Developer Portal". Verification, and with it approval of the intent, is required for apps in 100 or more servers ([privileged intents](https://docs.discord.com/developers/events/gateway#privileged-intents)). C2's application is far below both.
- **What a presence holds.** The status is `online`, `idle`, `dnd` or `offline`. There is no "invisible", so someone who is invisible on Discord reads as offline. Activities carry a type (Playing, Streaming, Listening, Watching, Custom, Competing) and a name, plus free-text `details` and `state`. A custom status's text is its `state`.
- **Discord's Developer Policy** limits API Data to what the app's stated function needs, and bars profiling users or their relationships ([Developer Policy](https://support-dev.discord.com/hc/en-us/articles/8563934450327)). A presence history is not something to keep.

## Decision

### 1. Recently played with: the tablemates query, cleaned up

**Recommended:** keep it derived, with no new table, as ADR 0051 decision 8 has it. The rejected alternative is a `recent_players` table written at game end. It would duplicate `seats`, and it would be one more thing to delete.

Changes to `GET /me/tablemates`:

- **Exclude tombstones** (`u.deleted_at IS NULL`), which fixes gap 1. Once §3 lands, also exclude anyone either person has blocked.
- **Which tables count stays the same:** any table both people sat at, including a lobby that never started. Decision 8's reasoning holds: a lobby yesterday is a better suggestion than a game a year ago. The last-occupant rule means a passer-by is not kept. Practice tables have no rows, so they never count.
- **A cap:** the 50 most recent people.
- **Each entry gains two fields:**
  - `games_together`: the count of shared tables;
  - `last_table`: the name of the most recent one. It is already in both people's History.
- **Later fields:** `friend` and `request` (incoming or outgoing) once §2 lands, and `presence` once §5 lands, where the person allows it.
- **Names:** the route keeps its name. The UI says **Played with**.

**Inviting from it** (slice 2):

- A Played-with row has an **Invite** button. It offers your Now tables that are in the lobby with an empty seat, plus **New table**, which creates a table from your last setup and then sends the invite.
- Delivery is today's DM invite. Where there is no bot token (every environment but production), the row offers **Copy invite**, as the card does today.

**Narrowing the DM route** (gap 2): with `user_id`, `POST /games/{id}/invites/dm` requires the target to be one of the caller's tablemates or friends, and neither person to have blocked the other. The admin's `discord_id` path, which `/c2-invite-dm` uses, is unchanged.

### 2. Friends: mutual, by request

| | **Mutual: request and accept** | One-way follow | Private favourites plus mutual friends |
|---|---|---|---|
| Consent to share presence | built in: accepting is the consent | none: the followed person never agreed | mutual friends only |
| New data | pairs, pending requests | follows | both |
| Abuse surface | requests, which are rate limited and blockable | silent following | requests |
| Familiar from | Discord, Steam | social feeds | (none) |

**Recommended: mutual.** Seeing someone's live status should take their yes. A request is the cheapest way to ask, and accepting is the yes.

- **Who you can ask:** someone in your Played with, meaning you have shared a table. Slice 7 adds two more ways: a friend link, and `/c2-friend @user` in the C2 Discord. There is **no search by name**, so nobody can list the site's accounts.
- **Where you ask from:**
  - a Played-with row;
  - the friends drawer;
  - a seat at your own table. The client sends `{game_id, player_id}` and the server resolves the account, because the game view carries no account ids and should not start to.
- **A request carries no text,** so friends add no new free-text surface to moderate.
- **Answering a request:**
  - **Accept** or **Ignore**. Ignore is silent: the request leaves your list, and the sender still sees it as pending until it expires.
  - Requests expire after **30 days**.
  - Two people who ask each other become friends at once.
- **Unfriending is silent,** from either side.
- **Limits:**
  - 200 friends;
  - 20 outgoing requests pending at once;
  - 10 new requests per hour per sender, counted in memory.
- **A setting:** "Who can send me friend requests": *people I've played with* (default) or *nobody*.
- **Who can't be a friend:** guests, bots and agent seats, which have no account, and the admin token, which is a credential, not a person.
- **Gate refusals and bans:** losing the gate (ADR 0156) or a ban leaves friendships in place. The person cannot sign in, so they show as offline. Deleting the account removes them (§10).

### 3. Blocking one person

`user_blocks` is a personal list. A block:

- ends any friendship and any pending request between the two, in either direction;
- removes each person from the other's Played with, friends and presence;
- refuses friend requests, DM invites and in-app invites between them, in either direction. The refusal reads like any other refusal ("can't send this to that person"), so the blocked person is not told. When the blocker sends, the message says they blocked this person.

**Open question:** should a personal block also keep the blocked person out of tables the blocker hosts? The list would then act as a standing `table_blocks` entry on every table the blocker hosts, checked at join and at the invite link. **Recommended: yes.** At someone else's table, that host's tools apply (ADR 0075 §2.6).

**Moderation hooks:**

- **No new report kind.** A request has no text. Abuse at a table is reported from the table, as it is today.
- **No #mod-log entries.** Friending and blocking are personal actions, and the mod log already leaves out a host managing their own table (ADR 0156's 2026-10-16 amendment).
- **No friend graph in the admin views** (a call; see below). Admins keep Live now, which already shows who is connected.
- **Ban sync is unchanged.** Signing someone out everywhere closes their sockets and stops their heartbeats, so they show as offline within one heartbeat window (§6).

### 4. What a presence says

The first row that applies wins:

| State | Source | Shown to a friend |
|---|---|---|
| `in_game` | a seat socket, not hidden, at a lobby or active table | "At «FNM» · lobby 2/4", or "In a game · «FNM» · turn 7" |
| `away` | the person's seat sockets at a table, every one hidden | "At «FNM» · away" |
| `watching` | a spectator socket with a user | "Watching «FNM»" |
| `practice` | a seat at a practice table | "Playing a practice game", with no link |
| `on_site` | a heartbeat (§6) within 60 s from a visible tab | "On CMD & CTRL" |
| `idle` | a heartbeat within the last 3 minutes from a hidden tab only | "On CMD & CTRL · idle" |
| `offline` | none of the above | "Offline" |

- **The Discord status (§7) is a second line,** shown alongside any C2 state: "On Discord · idle", or "Playing Balatro".
- **Admins:** an allowlisted admin in admin mode who watches a table as an admin shows as `on_site`, never `watching`, so moderators watching are not broadcast. The shared admin token is not a person.
- **Stand-ins:** a seat a stand-in bot plays is not the person being there. If they are disconnected, their seat does not count.
- **Never in a presence:** the voice room (ADR 0159 keeps who is connected to the table's seated players), the person's deck, anything hidden at the table, or who else is spectating.

### 5. Who sees it

| Who sees it | |
|---|---|
| **Friends only** | **Recommended default.** Accepting a request was the consent. |
| Friends and people I've played with | Offered as a setting. Shows your status to everyone you have shared a table with, #lfg strangers included. |
| Nobody | Offered as a setting. You read as offline. |
| Everyone signed in | Rejected. Seat names are already public through `GET /games`, but "online now" for everyone is a tracking feed. |

- **The setting:** "Who can see what I'm doing on CMD & CTRL": *Friends* (default) · *Friends and people I've played with* · *Nobody*.
- **Discord status has its own switch** (§7), **off** by default.
- **No reciprocity rule:** hiding yourself does not hide others from you. Hiding is a right, not a trade.
- **Played-with rows show presence** only for people who chose the second level.

### 6. Getting presence to the browser: polled (recommended) or pushed

**Recommended: poll, and let the poll be the heartbeat.**

- `GET /me/friends?visible=1|0` returns one body: friends, requests, Played with and presence.
- The header's Friends button (§9) calls it:
  - every 20 s while the tab is visible;
  - at once on focus;
  - every 2 minutes while the tab is hidden, with `visible=0`.
- Each call records the caller's heartbeat **in memory only**: user id, time and whether the tab was visible. A visible beat counts for 60 s, and a hidden one for 3 minutes, which covers the 2-minute hidden poll.
- Presence is computed per request from four sources:
  - a copy of the hub's `LiveSockets`, taken at most once a second and shared between requests;
  - the heartbeats;
  - the Discord map (§7);
  - the lobby's table metadata.
- No request holds two of those locks at once, the same rule `GET /admin/live` follows.

**Cost at ADR 0148's target of 120 players:** 120 players polling every 20 s is 6 requests a second. Each one is a few indexed SQLite reads (at most 200 friends and 50 Played-with entries) and map lookups, plus one copy of at most 400 sockets a second. That is small next to the Tables page's existing 5-second poll.

**The pushed alternative is a site socket**, `/ws/site`, one per signed-in tab.

- **What it buys:** presence, requests and invites arrive at once, and the Tables page's poll could ride the same socket.
- **What it costs:**
  - **A second socket lifecycle:** auth, origin checks, reconnects, and the deploy hand-over of ADR 0195.
  - **Connections against the 400-socket cap:** each tab is one. These sockets hold no seat, so they are the first refused when the server is full, and they would need a budget of their own.
  - **Fan-out:** each change goes to every connected friend allowed to see it. At 120 players, about 2 changes per player per minute and about 10 friends online each, that is about 40 frames a second. The CPU for that is cheap; the cost is running a second kind of socket.
  - **Debouncing:** tab visibility must be debounced, or every alt-tab becomes a broadcast.

A friend entry has the same wire shape either way. Moving from poll to push later is server and client plumbing only. **Recommended:** poll now, and move to push when a site socket is wanted for something else, or when measured poll load says so.

### 7. Discord status: read by the bot, held in memory

- **Who has one:** members of the C2 Discord server. With the gate on, that is every account except the gate's exemption list.
- **The bot:**
  - `CMDCTRL_DISCORD_PRESENCE=1` in the bot's environment (with `CMDCTRL_DISCORD_COMMUNITY_GUILD_ID`) adds `GUILD_PRESENCES`.
  - If Discord refuses the intent with close 4014, the bot falls back the way ban sync does: it logs an ERROR naming the portal switch and connects without the intent.
  - It keeps the latest presence of each community-server member in memory, built from Guild Create and Presence Update. It ignores every other server it is in.
- **Whom the bot forwards:**
  - Every 5 minutes, it reads `GET /admin/discord-presence/subjects`: the Discord ids of accounts that turned Discord status on.
  - It forwards those members only, with `POST /admin/discord-presence`. The body is a batch of `{discord_id, status, playing}`: changes every 15 s, and the whole set every 5 minutes.
  - Members who have not opted in never leave the bot process.
- **What is forwarded:**
  - The status.
  - For a **Playing** activity (type 0) only, its name, cut to 64 characters.
  - Never a custom status's text, Listening (Spotify), Streaming, or any `details` or `state`.
  - Open question: show the status only, with no "Playing".
- **The server** keeps a map in memory. An entry expires after 15 minutes without a refresh. It is never written to the database, never logged (the log carries counts only) and never put in a metric.
- **Production only.** The bot never runs from `develop`.
- **The client knows** from `GET /community`, which gains `"discord_presence": true|false`.
- **Note for #3312:** a member using C2's Discord Activity shows an activity with C2's application id, which the bot could read as "In CMD & CTRL (Discord)". That belongs to the Activity ADR, not here.

### 8. Joining, watching and inviting

- **Invite a friend.** The Now card's **Invite by DM** becomes **Invite a friend**. Its picker lists friends, online first, then Played with, each with a presence chip. It offers two deliveries:
  - the **Discord DM** (today's route; production only);
  - the **in-app invite**, added in slice 5 and described next.
- **The in-app invite:** `POST /games/{id}/invites/friend` with `{user_id}`.
  - **Who may send:** the same people who may send a DM invite.
  - **Who may receive:** a friend or a Played-with entry, not blocked.
  - **What it is:** a grant held **in memory**: game, invitee, inviter. It ends when the table starts or fills, or after 30 minutes.
  - **What the invitee sees:** a badge on the next poll, and **Join** in the drawer. Join calls `POST /games/{id}/join` with `"via": "friend"`, authorised by the grant instead of the invite token.
  - **What it fixes:** the grant needs no invite plaintext, so it still works after a restart. The DM path answers 409 then.
  - **What it costs:** a restart loses grants still pending.
- **Joining from presence.** A friend's presence offers **Join** for an empty seat at a lobby table in three cases:
  - **The table is posted to #lfg** (ADR 0161). Its link is already public in the C2 Discord, so Join asks for nothing new.
  - **The host has turned on "Friends of the players can join"** (slice 5; lobby metadata, like the voice and #lfg choices; off by default).
  - **Otherwise it offers Ask to join.** This is an in-app request to your friend, held 10 minutes in memory. They answer with one click, which sends an in-app invite, or they ignore it.
- **Watching from presence:**
  - With the host setting **"Friends of the players can watch"** (off by default), **Watch** gives the friend a spectator session.
  - Otherwise it offers **Ask to watch**, answered with an in-app spectator grant.
  - A spectator sees only what any spectator sees.
- **These stay as they are:** Copy invite, Post to #lfg and the rematch offer. Friends are for people you know; #lfg is for strangers.
- **No "tell all my friends" broadcast.** It is the spam the #lfg ping cooldown exists to prevent.
- **Voice rooms:** a friend who joins is a seated player, and the voice manager grants them as it does today.
- **DM notifications:** Settings → Notifications (ADR 0162) gains a fourth switch, **Friend requests and invites**. It is off by default, and uses the same away rule and per-person rate window.

### 9. Where it lives

| | |
|---|---|
| **A. A Friends button in the header and a drawer** | **Recommended.** The button shows a people icon and the online count, on every signed-in page that is not a game. The drawer has Requests, Friends (online first) and Played with. A `#/friends` page manages requests sent, blocks and the two settings. |
| B. A rail on the Tables page | Cheaper, but missing from Decks, Catalog and Stats, so those pages send no heartbeat. |
| C. A nav item and a page | Clean, but presence is always a click away. |

- **At the table:** no drawer. The table is busy, and tablet is the minimum screen. Friends appear in the invite picker while the table is in the lobby.
- **At 820 px**, the drawer covers the page.
- **History filters:** **With friends** becomes **With people** (another human was at the table), and a new **With friends** means a current friend was there.
- **Labels:** every new label goes through `labels.ts`.

### 10. Privacy and deletion (ADR 0149)

What is stored, for how long, and what deleting an account does to it:

| Data | Where | Kept | Deleting the account |
|---|---|---|---|
| Played with | derived from `seats`; nothing new | as long as the games | the tombstone is filtered out (§1); nothing more to delete |
| Friendships | `friendships(user_a, user_b, created_at)`, one row per pair, `user_a < user_b` | until either person removes it | every row with the person on either side is **deleted** |
| Friend requests | `friend_requests(from_user, to_user, created_at, ignored_at)` | until answered or cancelled, or **30 days** | either side: **deleted** |
| Blocks | `user_blocks(user_id, blocked_user_id, created_at)` | until unblocked | either side: **deleted** |
| Social settings | `social_prefs(user_id, presence, requests, updated_at)`; slice 6 adds `discord_status` | until changed | **deleted** |
| The host's friends-join and friends-watch choices | `games.friends_join`, `games.friends_watch` | with the table | not personal |
| Heartbeats | memory | 60 s (visible) or 3 min (hidden) | dropped |
| Discord status | memory | 15 min without a refresh | dropped, by Discord id |
| In-app invites, asks to join or watch | memory | 30 min at most | dropped |
| Request rate counters | memory | 1 h | dropped |

- **`accountdelete.Inventory`** lists every column of the four tables as **Deleted**, with a note.
- **`Delete`** removes the person's rows from each table, on either side, in its one transaction: `WHERE user_a = ? OR user_b = ?`, and the like.
- **After the commit,** the memory entries are dropped (`presence.Forget(userID, discordID)`).
- **The audit line** gains the counts `friendships`, `friend_requests`, `user_blocks` and `social_prefs_rows`.
- **Tests:** `TestInventoryCoversTheSchema` holds the inventory to the schema, and `TestDeleteTreatsEveryInventoriedColumn` seeds rows that name the person on both sides.
- **A new account starts clean.** A deleted person who signs in again is a new account (ADR 0149 §2). Blocks they made go, and blocks others made against them go too, as `table_blocks` entries do today.

**The privacy page** (`routes/Privacy.svelte`, numbers in `lib/privacy.ts`) changes in the slice that ships each part:

- **Slice 2: a new "People you've played with" entry under "Who it is shared with".** It also fixes gap 3. The text:
  - "Your display name and avatar appear in the Played with list of each signed-in person you have shared a table with, with the name of your latest table together and how many you have shared."
  - "They can have the bot send you a table's invite link as a Discord direct message."
  - "Deleting your account removes you from these lists."
- **Slice 3: a new "Friends and blocks" entry under "What is stored, and why".** It covers your friends, pending requests (expiring after 30 days), your blocks and your two settings: what each lets other people see, and that each lasts until you change it or delete your account.
  - New numbers: `friendRequestDays` and `friendCap`, pinned by `TestPrivacyNoticeNumbersMatchTheServer`.
  - `PRIVACY_SCHEMA_VERSION` takes the migration's number.
- **Slice 4: "What your friends see", under "Who it is shared with".** It lists the states of §4, who sees them under your setting, and that the heartbeat is kept in memory for at most three minutes and never written down.
- **Slice 6: a paragraph in the Discord row.** "If you turn on Discord status, the bot reads your status in the CMD & CTRL Discord server, and the name of a game you are playing, and shows it to the people your setting allows. It is kept in memory for at most fifteen minutes and never written down. Discord already shows your status to that server's members."
  - New number: `discordPresenceMinutes`.

### 11. Configuration

- **Slices 2 to 5** have no new variables. Friends need accounts, so they are on wherever there is a database. The Discord parts (DM invites, the DM switch) stay off without the bot token, as today.
- **Slice 6** adds `CMDCTRL_DISCORD_PRESENCE=1`, read by both the bot and the server. CD writes it from an Actions variable, always, empty included, like ban sync's flag.
- **Owner setup (once):** switch on **Presence Intent** under the application's Bot → Privileged Gateway Intents in the Developer Portal, then deploy `main`.

## Calls made here

Each of these is overturnable in review:

- **Limits:** a Played-with cap of 50; 200 friends; 20 outgoing requests; 10 new requests an hour; requests expire after 30 days.
- **Timings:** a 20 s visible poll and a 2-minute hidden poll; heartbeats last 60 s visible and 3 minutes hidden; Discord status entries last 15 minutes.
- **Ignoring a request is silent.** There is no "declined" record.
- **No search by name, and no text in requests.**
- **No reciprocity rule** for hiding presence.
- **Admins watching a table show as on the site, not as watching.**
- **A practice game shows with no link.**
- **No broadcast of a table to all friends.**
- **No friend graph in the admin views.**
- **"With friends" in History becomes "With people",** and the name goes to the real thing.
- **The DM invite route is narrowed** to tablemates and friends.
- **Played with keeps counting lobbies that never started.**

## Consequences

- **Slice 2 changes no schema.** The first useful slice, Played with plus invite, is a query fix, two fields, a narrower route and a privacy line.
- **Friends are this database's first many-to-many between people.** Every row names two people, and deletion clears both sides.
- **Presence is never written down.** A restart forgets it, and the first poll after the restart rebuilds the C2 part. The Discord part rebuilds within 5 minutes, at the bot's next full set.
- **Discord status comes from the bot process.** If the bot is down, the second line disappears and nothing else breaks.
- **Hosts get two new table settings** (friends join, friends watch). Both are off by default, so a table is invite-only unless its host opens it.
- **No more server traffic than the Tables page already makes.** The header poll is one more request every 20 s per signed-in tab.

## Back-out

| To undo | Do |
|---|---|
| Slice 2 | revert the query and route changes. Nothing is stored. |
| Slices 3 and 5 | drop the routes. The tables can be dropped by hand after rolling the binary back (an older binary refuses the newer schema, `ErrSchemaTooNew`); see docs/environments.md. |
| Slice 4 | remove the poll's presence fields. Nothing is stored. |
| Slice 6 | unset `CMDCTRL_DISCORD_PRESENCE`. The bot stops asking for the intent and the server forgets the map within 15 minutes. |

## Slices

| # | Slice | Schema | Routes and wire | Client | Privacy | Model | Needs |
|---|---|---|---|---|---|---|---|
| 1 | This ADR | — | — | — | — | — | — |
| 2 | **Played with, and inviting from it** | none | `GET /me/tablemates`: drop tombstones, cap at 50, `games_together`, `last_table`; narrow `POST /games/{id}/invites/dm`; docs/lobby.md | the drawer with Played with only (fetched when it opens, no polling); Invite → pick a table or New table; Copy invite where there is no bot | "People you've played with" | Sonnet | 1 |
| 3 | **Friends and blocks** | `friendships`, `friend_requests`, `user_blocks`, `social_prefs` | `GET /me/friends`; `POST /me/friends/requests` (`user_id`, or `game_id` + `player_id`); accept, ignore, cancel; `DELETE /me/friends/{user_id}`; `GET`/`PUT`/`DELETE /me/blocks/{user_id}`; `GET`/`PUT /me/social`; blocks applied to tablemates and DM invites | drawer sections, `#/friends`, Add friend on rows and seats, the History filter rename, labels and e2e | "Friends and blocks", the numbers, the schema bump, `Inventory` and `Delete` | Opus | 2 |
| 4 | **Presence on CMD & CTRL** | none (memory) | `GET /me/friends?visible=`: heartbeats and `presence`; the shared socket copy; Join for tables posted to #lfg; Ask to join and Ask to watch | the header button with its count; the poll schedule; chips; Join and Ask | "What your friends see" | Opus | 3 |
| 5 | **In-app invites, and friends join or watch** | `games.friends_join`, `games.friends_watch`; `notification_prefs.friends` | `POST /games/{id}/invites/friend`; join and spectate by grant; the host settings; the fourth DM kind | Invite a friend; the host toggles; the Settings switch | the notifications paragraph; the schema bump | Opus | 4 |
| 6 | **Discord status** | `social_prefs.discord_status` | the bot's intent, its 4014 fallback and its forwarding; `GET /admin/discord-presence/subjects` and `POST /admin/discord-presence`; the `GET /community` flag; CD | the switch; the Discord line | the Discord paragraph and its number | Opus | 4 |
| 7 | *Optional:* **adding someone you haven't played with** | `friend_codes(user_id, code_hash, created_at)` | a friend link that can be rotated; `/c2-friend @user` through the admin route | Copy my friend link | one sentence | Sonnet | 3 |

Slices 5 and 6 are independent of each other. Every client slice attaches screenshots at 1440 px and 820 px.

## Open questions for the owner

Each question lists the recommended option first.

1. **The friend model:** (a) mutual, request and accept; (b) one-way follow; (c) private favourites plus mutual friends.
2. **Who sees presence by default:** (a) friends only, with "people I've played with" and "nobody" as settings; (b) friends and people I've played with by default; (c) everyone signed in.
3. **Discord status:** (a) opt-in (off by default), the status plus a "Playing" game name; (b) opt-in, the status only; (c) on by default; (d) not built.
4. **Joining from presence:** (a) Join for tables posted to #lfg, a host setting "Friends of the players can join" (off by default), and Ask to join otherwise; (b) Ask to join only; (c) Join whenever a friend's table has an empty seat.
5. **Watching from presence:** (a) a host setting "Friends of the players can watch" (off by default), and Ask to watch otherwise; (b) Ask to watch only; (c) no watching from presence.
6. **How far a personal block reaches:** (a) lists, requests and invites, and a seat at every table the blocker hosts; (b) lists, requests and invites only.
7. **Poll or push:** (a) poll every 20 s now, push later if a site socket is wanted for other things; (b) build the site socket now.
8. **Where it lives:** (a) a header button and drawer, plus a `#/friends` page; (b) a rail on the Tables page; (c) a nav item and page.
9. **Limits:** (a) 200 friends, 30-day request expiry, Played with capped at 50; (b) other numbers.
10. **A DM for friend requests and invites:** (a) a fourth switch in Settings → Notifications, off by default; (b) in-app only.
11. **Adding someone you have not played with:** (a) a friend link and `/c2-friend @user` (slice 7); (b) the friend link only; (c) not at all: you can only friend someone you have shared a table with.
12. **Admin views:** (a) no friend graph; (b) friend and block counts per account in Accounts.
