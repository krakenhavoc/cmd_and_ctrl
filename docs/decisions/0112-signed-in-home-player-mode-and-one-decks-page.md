# ADR 0112 — Signed-in home, player mode, and one decks page

**Status:** Accepted · 2026-10-02 · S57 — Signed-in home, player mode, and one decks page (tracker [#1992](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1992))
**Owner decisions:** 2026-10-02, the three on #1992, quoted under [Context](#owner-decisions-2026-10-02). They are binding. The owner then answered this ADR's four questions the same day, each with the recommended option (a): see [Owner decisions, 2026-10-02 (questions)](#owner-decisions-2026-10-02-questions) at the end. The sections and the Delivery plan below are written as decided. The options not chosen are kept as considered options under [Questions for the owner (answered)](#questions-for-the-owner-answered).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-02. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 37 remote heads: `origin/develop`, `origin/main` and 35 chore, docs, feat, fix, repro and wip branches. The highest number on any of them is 0111 (`0111-action-dock.md`). No branch has 0112, and no open PR adds an ADR, so this ADR takes **0112**.
**Amends:** [ADR 0110](0110-remember-me.md) §1 item 3 (the code box moves from the login page to the lobby, and `App.svelte`'s bounce loses its exemption), §3 items 1, 4 and 5 (an allowlisted person is an admin only in admin mode; `/me` reports both; WebSocket parity holds in admin mode only), §6 item 5 (`#/decks` becomes public and takes in the deck check). [ADR 0095](0095-deck-coverage-and-deck-requests.md) §5 (`#/deck-check` becomes an alias of `#/decks`). [ADR 0092](0092-public-roadmap-and-site-portal.md) decision 5: for a signed-in person the header is the portal, and `#/home` stays as the public site map.
**Builds on:** [ADR 0044](0044-surviving-a-deploy.md) decision 3 (HMAC sessions, advisory `Revoke`), [ADR 0051](0051-user-database.md) decisions 3, 6 and 7 (sessions, revocation, the deck library), [ADR 0095](0095-deck-coverage-and-deck-requests.md) (the coverage report and deck requests), [ADR 0110](0110-remember-me.md) (durable sign-in, the admin allowlist, saved decks).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

The owner reported this on prod on 2026-10-02, after the S55/S56 promotion (#1984):

> I need a button to switch between admin and normal login I think. And users need a menu as this is what it looks like when I sign in as a user with discord but not joined a table yet

> and my decks is empty

The tracker's diagnosis holds on `origin/develop` at `5705ae80`:

1. **A signed-in person with no seat has no menu.** `App.svelte` keeps a session that `canJoinByCode` accepts on `#/login`, because ADR 0110 §1 item 3 put the invite-code box there. The login page has no `SiteHeader` (#1386 left it out on purpose), so from there the only ways on are "See my games", "Site map" and the catalog link.
2. **Admins have no switch.** A Discord account on `CMDCTRL_DISCORD_ADMIN_USER_IDS` is an admin on every request (ADR 0110 §3). The login page still shows that admin the shared-token form, and nothing lets them sit at a table as an ordinary player.
3. **"My decks" is empty.** `saveToLibrary` (`lobby/http.go`) saves only an upload by a caller with a `UserID`, and never a pre-built pick. Decks seated through the admin token, a guest seat or the pre-built picker were never stored, and a seat keeps only `deck_id` and `deck_name`, so those decks cannot be recovered. The deck check (`#/deck-check`, ADR 0095) and the library (`#/decks`, ADR 0110 §6) are separate pages.

The audit for this ADR found three more things:

- **An allowlisted admin who seats someone else's deck saves it to their own library.** `saveToLibrary` keys on the caller, and `uploadDeck` lets an admin set any seat's deck. `recordLastDeck` already refuses this case ("an admin installing somebody else's deck is not choosing a deck for themselves"). `saveToLibrary` does not.
- **A library deck can ask for its missing cards only if it came from a link.** `MyDecks.svelte` sends a link deck to `#/deck-check?url=…`. A pasted deck has no request button, because `POST /deck-requests` takes only a link or a list, and `GET /me/decks` does not return the list.
- **Nothing saves a deck outside a table.** The library is written only by `uploadDeck`. The deck check never writes it.

### Owner decisions, 2026-10-02

Quoted from #1992:

1. **The Lobby is the signed-in home.** Once signed in, Login sends you to the Lobby, which has the header. The join-with-a-code box moves onto the Lobby. Login is only for signed-out visitors.
2. **Admin switch: server-side "play as player".** An Admin chip in the header turns player mode on and off. In player mode the server itself drops the session's admin rights. The token box disappears for Discord admins.
3. **One decks page.** My decks and Deck check fuse into one page. You link your deck, see how many cards are implemented, request the missing ones, and save it, all in one flow.

---

## 1. The signed-in home

### What exists

- **Routing** (`App.svelte`). A signed-out visitor on any route that is not public goes to `#/login`. A session on `#/login` goes to `#/lobby`, unless `canJoinByCode(s)` holds: an `identified` session, or a player or spectator session with a `user_id`. Those stay, because the code box is there.
- **OAuth completion** (`oauthComplete`). The invite flow goes to `#/games/{id}`. The login-page flow (an identity-only session) goes back to `#/login`.
- **The login page** has three cards: Discord sign-in, the code box, and the admin token. For a signed-in person, the code card also carries "Back to your table", "See my games", "Sign in with a different Discord account", "sign out" and "sign out everywhere".
- **The lobby** has the `SiteHeader` and, under it, its own command bar: a role chip, "my games", "join with a code" (a link back to `#/login`), settings, "log out" and "log out everywhere". `visibleGames` shows a seated guest or player their own table and the tables they created. Every other session sees every table in the room. The empty state says "Create a table, then send the invite link to your pod" when the session may create one, and "You're not seated at a table yet — ask for an invite link" otherwise.
- **The header** (`SiteHeader.svelte`) links Lobby, My games, My decks, Catalog, Roadmap, Deck check and Home. It shows the person's name and a "Sign out" button. Its wordmark goes to `#/home`.

### Decision

1. **Login is for signed-out visitors only.** Every session on `#/login` goes to `#/lobby`. `canJoinByCode` stops being a routing exemption and stays only as "this session may join by code".

   | Session on `#/login` | Today | After |
   |---|---|---|
   | none | stays | stays |
   | `identified` (with or without a user database) | stays | `#/lobby` |
   | player or spectator with a `user_id` | stays | `#/lobby` |
   | guest player or spectator (no user) | `#/lobby` | `#/lobby` |
   | admin token | `#/lobby` | `#/lobby` |

   An expired session is cleared before the router runs, so the visitor is signed out and stays on `#/login`, where the expiry notice is shown, as today.
2. **The ways in that do not go through Login are unchanged.** `join`, `reclaim` and `oauthComplete` stay public routes.
   - An invite link (`#/games/{id}/join?t=…`) and a spectator link (`…&spectator=1`) open the Join page with whatever session the browser holds, as today. A signed-in spectator heading to a new table follows the link, or pastes it into the lobby's box (item 3), which opens the same Join page.
   - A reclaim link opens the Reclaim page, as today.
   - OAuth completion: the invite flow still goes to `#/games/{id}`. The login-page flow goes to `#/lobby`, or to the return route the decks page saved (§3 item 7).
3. **The code box moves to the top of the Lobby**, beside the create form, as a "Join a table" card. On a narrow screen the two stack, join first.
   - **A signed-in person** (`canJoinByCode`): a code or a link. A code posts `POST /join`, which seats them as their Discord identity (ADR 0110 §1 item 3). A link navigates to its Join page.
   - **A guest seat or guest spectator:** a link only. A bare code shows "This browser is seated as a guest. Open the invite link instead, or link Discord from your table's menu." The server already refuses a guest's code with 409, and the box says so before the request rather than after.
   - **The admin token:** no box. The token is not a person and cannot take a seat by code (409 today).
   - Home's "Join with an invite" card links `#/lobby` for any session, and `#/login` when signed out.
4. **The account controls move to the header.** `SiteHeader`'s account area becomes a menu button showing the name. Its menu holds:
   - the Admin chip, for an allowlisted person (§2 item 9);
   - Settings;
   - "Sign in with a different Discord account", when Discord sign-in is configured and the session has a `user_id`;
   - "Sign out", and "Sign out everywhere" when `canSignOutEverywhere` holds.

   The lobby's own command bar goes. Its page title stays. The login page loses the whole signed-in card, because no signed-in session reaches it any more.
5. **The header's wordmark goes to `#/lobby` for a session and to `#/home` without one.** Its links become Lobby, My games, Decks, Catalog, Roadmap and Home (§3 item 2 merges the two deck links). `#/home` stays the public site map. ADR 0092 decision 5 asked the owner to pick between the header and `#/home`. Owner decision 1 picks the header for a signed-in person, and `#/home` remains for visitors.
6. **What the Lobby lists for a signed-in person** ([owner answer 3](#owner-decisions-2026-10-02-questions)). A signed-in person who is not an admin in admin mode sees **their tables**. These are the session's own table, every table they created (`is_creator`), and every open table where they hold a seat: the `GET /me/games` entries that have a `rejoin` path. A table they hold a seat at through a different session shows an "Open" button, which calls that `rejoin` (`POST /me/games/{id}/session`) and then opens the table. An admin in admin mode, and the admin token, see every table, as today. A guest keeps today's view. This filter is presentation: `GET /games` is unchanged, and it never serves an invite to a caller who may not have it (`redactMetaFor`).
7. **A signed-in person with no tables** sees the Join and Create cards at the top (S55 PR 7, #1975, let every signed-in player create a table), and under them one empty-state card: "You're not at a table yet. Paste an invite code above, or create a table and invite your pod." It has three links: "Practice against bots" (`#/practice`), "Check a deck" (`#/decks`) and "My games" (past tables).

### Migration / snapshot impact

None. Client only.

---

## 2. Player mode

### What exists

- **Two predicates decide every admin question** (`lobby/admins.go`, ADR 0110 §3). `isServerCredential(p)` is the shared-token session (`p.Role == auth.RoleAdmin`). `isAdmin(p)`, through `isAdminPrincipal(admins, p)`, is the token, or a session with a `UserID` and a `DiscordID` that is on the allowlist. `TestRoleAdminIsComparedOnlyInTheAdminPredicates` fails CI if `auth.RoleAdmin` is compared anywhere else.
- **Nothing about admin is in a token.** The allowlist is read on every request, `GET /me` reports `admin`, and the client keeps it on the session (`lib/admin.ts`). `POST /me/session` renews a token without looking at admin.
- **A WebSocket's admin bit is fixed at upgrade.** `WSAuthorizer.AuthorizeUpgrade` sets `ws.Binding.Admin`, and the hub copies it to `Client.admin` for the connection's lifetime. Four readers depend on it:
  - `actions.overrideCaller`: an admin seated at a table drives any card from the admin context menu;
  - `Room.CanManageTable`: the host gates;
  - `viewerIDForFilter`: a seatless admin binding gets the unfiltered view, and a seatless non-admin gets the spectator view;
  - the hub's undo path: only an admin may undo after the game has ended (ADR 0057).
- **Evicting a person's sockets** exists for revocation: `Hub.EvictUserSessions(userID, before)` closes them with code 1000, which the client treats as terminal (ADR 0044 decision 2). Any other code makes the client reconnect.
- **The token form** is on the login page as a secondary card, and leads on `#/admin` (`<Login admin />`).

### Where the mode lives

Three places were considered:

- **A claim in the session token.** Rejected. `Principal` and the HMAC claims would gain a field. Every mint (`issueFor`) and the renewal (`POST /me/session`) would have to copy it. Switching would mean issuing a new token, and under HMAC the old one stays valid until it expires (`Revoke` is advisory, ADR 0044 decision 3), so a copied pre-switch token would keep admin rights. Each tab and device would also hold its own answer. And ADR 0110 §3 item 1 made admin "a capability checked on every request, not a role baked into a token" precisely so that nothing outlives a change.
- **Per session, on the server.** Rejected. HMAC sessions have no server-side row to hang it on.
- **Per user, on the server.** Chosen. The switch belongs to the person: every tab and device follows it at the next request, nothing in a token changes, and renewal is untouched.

### Decision

1. **Admin mode is per-user server state, and it is off by default.** Migration `0009` adds `users.admin_mode_at INTEGER NOT NULL DEFAULT 0`: the Unix millisecond time admin mode was switched on, or 0 for off. Every allowlisted person therefore starts in **player mode**, the owner included, the first time this deploys. Admin rights are something the person turns on, and the chip shows which mode they are in. That is the least-privilege default.
   - A new `users.AdminModes` keeps the rows with `admin_mode_at > 0` in memory, loaded at boot. It is written through on every switch: the database first, then the map. This mirrors `users.Revocations`.
   - **If the boot load fails,** the server starts with an empty map and logs an ERROR. Every allowlisted person is then in player mode until they switch again. This is deliberately not `newRevocations`' exit: there, the safe failure is to refuse to run, and here, the safe failure is fewer admins.
   - With no user database there are no users, so no allowlisted admins and no mode. That is a supported state, as in ADR 0110 §3.
   - **Admin mode lasts 12 hours, like `sudo`** ([owner answer 1](#owner-decisions-2026-10-02-questions)). It ends `AdminModeTTL` (12 hours, the default length of a non-person session, `CMDCTRL_SESSION_TTL`) after `admin_mode_at`. `AdminModes.On(userID, now)` is false from that moment, so HTTP sees the lapse at the next request with no write needed. A sweeper runs once a minute: for every lapsed row it sets `admin_mode_at` to 0, drops it from the map, logs `admin mode lapsed` with `admin_user_id`, and calls `Hub.RebindUserSessions(userID, false)` (item 5), so the admin menu goes away at the table too. A lapse found while the server was down is swept on the first tick after boot.
2. **One predicate, one more term.**

   ```go
   func isAdminPrincipal(a *Admins, p auth.Principal) bool {
       if isServerCredential(p) {
           return true
       }
       return isAllowlisted(a, p) && a.Modes.On(p.UserID, now)
   }

   // isAllowlisted: on the list, whatever the mode. Asked by the switch
   // route and by /me's admin_allowed, and by nothing else.
   func isAllowlisted(a *Admins, p auth.Principal) bool {
       return p.UserID != uuid.Nil && p.DiscordID != "" && a.List.Has(p.DiscordID)
   }
   ```

   `Admins` bundles the allowlist and the modes. `lobby.Config` and `WSAuthorizer` share one instance, as they share `*AdminList` today. The guard test gains a second rule: `AdminList.Has` and `AdminModes.On` are called only inside `admins.go`. A later route then cannot test the allowlist directly and skip the mode.
3. **The switch: `PUT /me/admin-mode`**, with body `{"on": true|false}`.
   - **Caller:** `signedInUser`, and then `isAllowlisted`. Anything else is a 403: "not an admin". The token gets 403 too, because it has no mode to switch.
   - **Effect:** it writes `admin_mode_at` (now, or 0), updates the map, and then rebinds the person's sockets (item 5).
   - **Response:** `{admin, admin_mode, admin_mode_ends_at}`. `admin_mode_ends_at` is the Unix millisecond time admin mode lapses, present only while it is on.
   - **Limit:** a per-caller bucket of 1 per 2 seconds with a burst of 5. Switching on while already on restarts the 12 hours, which is how an admin extends it. Switching off while already off is a no-op that answers 200.
   - **Audit:** every switch is logged at Info as `admin mode on` or `admin mode off`, with `admin_user_id`.
4. **HTTP: the next request sees it.** In player mode, `isAdmin(p)` is false for that person, so they get exactly what a non-admin signed-in person gets. The only extra thing they may do is switch back. [What player mode drops](#what-player-mode-drops) lists every call site.
5. **WebSockets: a switch closes the sockets whose admin bit is now wrong.**
   - **The bit is never flipped on a live connection.** A seatless admin binding is not read-only (`adminBinding`), so a socket flipped to `admin=false` in place would be a seatless, writable, non-admin connection. Its `Caller` would be `uuid.Nil`, which `requirePriorityHolder` and `requireCardController` read as "bypass". That would fail open.
   - **Instead,** `Hub.RebindUserSessions(userID, admin)` closes every socket of that user whose `admin` differs from the new answer, with close code **4001**, reason `admin mode changed`. Any code other than 1000 is non-terminal on the client, so the client reconnects, and `AuthorizeUpgrade` runs again with the new answer.
   - **A reconnect into player mode** binds the person's own seat or spectator session, without `Binding.Admin`. A binding that only an admin may hold (another game, another seat, the seatless view, an `identified` session) is refused at the upgrade. The browser cannot see an upgrade's 403, so the client does not rely on it. On a 4001 close it asks `GET /me` first. If it is no longer an admin and its binding was not its own seat, it goes to the Lobby with "Player mode is on; this table isn't yours", instead of reconnecting.
   - **Switching back on** closes the person's ordinary sockets the same way, so they come back with `Binding.Admin`.
6. **Tokens and renewal are untouched.** No claim is added, so `issueFor`, `POST /me/session` and the HMAC format do not change. A token minted before a switch obeys the switch.
7. **Signing out everywhere also ends admin mode.** `users.Revocations.RevokeAll`, behind `POST /logout/everywhere` and `POST /admin/users/{id}/revoke-sessions`, sets `admin_mode_at` to 0 in the same transaction. A person who believes a session leaked gets player mode back along with the revocation.
8. **The shared token is unaffected, and its form leaves the login page.**
   - The token session has no user, so it has no mode. It is an admin on every request, as ADR 0110 §3 item 6 says, and the bot keeps calling with it.
   - **The login page drops the token card,** and adds no link to replace it. Players never need the token (the page already says so), and owner decision 2 removes the box for Discord admins.
   - **`#/admin` becomes the token's only page,** and it is documented in `docs/lobby.md` for operators:
     - signed out: the token form;
     - an allowlisted person, in either mode: `#/lobby`, where the chip is;
     - the token session itself: `#/lobby`;
     - any other session (a non-admin person, a guest): the token form, with a note that it replaces this browser's session. A session with a user is set aside and restored when the token's session ends (ADR 0110 §1 item 6).
9. **The client.**
   - **`GET /me`** adds `admin_allowed` (on the allowlist, persons only), `admin_mode` and `admin_mode_ends_at`. `admin` stays the effective answer. `isAdmin(session)` keeps reading `admin`, so every admin control already follows the mode.
   - **The Admin chip** sits in the header's account area (§1 item 4), for a session with `admin_allowed`. It is a toggle button (`aria-pressed`). "Admin" on the accent colour means admin mode is on, and the chip shows when it ends ("Admin · until 23:40"); "Player" in muted outline means player mode. The tooltip names the other mode. A click sends `PUT /me/admin-mode` and updates the session from the answer. A refused switch leaves the chip as it was and shows the server's message.
   - **The table has no header** (ADR 0092), so the in-game menu (ADR 0111 PR 7) gets "Switch to admin mode" or "Switch to player mode" for the same sessions.
   - **The client asks `/me` again** after a switch, on a 4001 close, when a hidden tab becomes visible, and when `admin_mode_ends_at` passes. That way a second tab or device catches up, and the server stays the gate: a stale `true` only shows a control whose click is refused (ADR 0110 §3 item 4).
   - **The admin token** shows a static "Admin token" badge in place of the chip. To play as a person, sign out of the token, which restores the saved Discord session if there is one.
10. **The Discord bot is unchanged.** `/c2-end` keeps its own allowlist check (`server/internal/bot/config.go`) and calls the server with the token. Player mode is a mode of the site. A Discord command is a deliberate operator act in another app, and the bot cannot see a site session.
11. **Player mode is not a defence against a stolen session.** Whoever holds an allowlisted person's session can switch admin mode back on. Revocation (item 7, ADR 0051 decision 6) is the defence against theft. Player mode is how an admin plays fairly, and how they avoid using admin powers by accident. Asking for a fresh Discord sign-in before admin mode turns on is [out of scope](#out-of-scope).

### What player mode drops

This is every admin call site in the server's non-test Go code on `origin/develop` at `5705ae80`, from `grep -rnE 'c\.isAdmin\(|requireAdmin\(c|isServerCredential\(p|isAdminPrincipal\(a'`. In player mode, each `isAdmin` and `requireAdmin` site answers as it does for a non-admin signed-in person.

**Routes behind `requireAdmin` (403 in player mode):**

| Route | Handler | What it does |
|---|---|---|
| `DELETE /games/{id}` | `deleteGame` | delete any table |
| `POST /games/{id}/archive` | `archiveGame` | archive any table |
| `DELETE /games/{id}/archive` | `unarchiveGame` | unarchive any table |
| `POST /games/{id}/seats/{player}/reclaim` | `mintSeatReclaim` | mint a reclaim ticket for any seat |
| `GET /games/{id}/creator` | `gameCreator` | "is this Discord ID the creator" (the bot's question) |
| `GET /bugreport/{id}/replay` | `bugPinnedReplay` | the pinned unfiltered replay |
| `GET /bugreport/{id}/gamelog` | `bugPinnedGameLog` | the pinned game log |
| `POST /admin/users/{id}/revoke-sessions` | `adminRevokeUserSessions` | revoke another person's sessions |
| `GET /games/{id}/bot/stats` | `botStats` (`BotStatsHandler`) | per-bot latency and token readout |

**Handler checks through `c.isAdmin(p)` (falls back to the non-admin rule):**

| Site | Function | Admin right dropped |
|---|---|---|
| `spawn.go:339` | `requireTableManager` | spawn cards at a table you do not host (dev only) |
| `http.go:891` | `createGameWith` | the new table's invites in the reply (the creator keeps them through `is_creator`) |
| `http.go:916`, `:930` | `transferHost` | hand any table's host seat on; see every invite in the reply |
| `http.go:963` | `updateTableSettings` | change any table's settings |
| `http.go:1245` | `listGames` | every table's invites in the list |
| `http.go:1272` | `getGame` | any table's invites |
| `http.go:1529` | `rotateInvite` | rotate any table's invites |
| `http.go:2217` | `downloadReplay` | any game's replay, and before it has ended |
| `http.go:2334` | `startGame` | start any table |
| `http.go:2631` | `uploadDeck` | set any seat's deck |
| `http.go:3212` | `botSeatAuthorised` | add or remove bots at a table you do not sit at |
| `http.go:3405` | `me` | `admin` reads false; `admin_allowed` stays true |
| `bugreport.go:555` | `bugGameLog` | attach another game's log to a bug report |
| `bugreport.go:589` | `bugReplaySource` | attach another game's replay to a bug report |
| `setups.go:232` | `canApplySetup` | apply your setup to a table you neither host nor created |
| `setups.go:281` | `applySetupRoute` | invites in the reply |
| `invite_dm.go:98` | `inviteDM` | DM the invite of a table you neither sit at nor created |
| `admins.go:167` | `requireAdmin` | the nine routes above |

**WebSocket (`ws_authorizer.go:85`, `isAdminPrincipal`):** in player mode the person gets `bindingFor`, the ordinary binding. Dropped:
- binding to any game;
- binding as any seat (`?player=`);
- the seatless unfiltered view;
- an `identified` session binding at all;
- `Binding.Admin` on their own seat, and with it the card overrides from the admin context menu (`overrideCaller`), the host gates (`Room.CanManageTable`) and undo after the game ends.

**Kept on purpose.** The eight `isServerCredential` sites ask "is this the shared token", which a person never is, in either mode:
- `deckCoverageLimit` (`deckcheck.go:119`), `deckRequesterFor` (`:509`) and `renderDeckRequestIssue`'s bot flag (`:627`);
- `callerKey` (`http.go:703`) and `createGameWith`'s `server` (`:845`);
- `practiceOwner` (`practice_http.go:224`);
- `inviteDM`'s raw-snowflake rule (`invite_dm.go:114`);
- `AuthorizeUpgrade`'s `!isServerCredential(p)` before `ownBinding` (`ws_authorizer.go:86`).

A person in player mode also keeps everything any signed-in person has, including creator and host rights at their own tables. The only right player mode adds is the switch (`PUT /me/admin-mode`) and the `admin_allowed` flag on `/me`. A table-driven test asks every route above as a non-admin, as an admin in player mode and as an admin in admin mode, and asserts that the first two answers are identical.

**The client follows `admin`.** The `isAdmin(session)` readers in `Lobby.svelte` (every table, the archive list, delete, reclaim, replay links, bot management), `Game.svelte` (the admin context menu, `canManageTable`, `canSpawn`, the undo budget) and `gameURL.ts` all read the effective `admin`, so they need no change of their own.

### Migration / snapshot impact

Migration `0009_admin_mode.sql`: `ALTER TABLE users ADD COLUMN admin_mode_at INTEGER NOT NULL DEFAULT 0`. It is additive, and times are Unix milliseconds, as in `0002` to `0008`. No snapshot impact: admin is a connection property, never game state.

---

## 3. One decks page

### What exists

- **`#/deck-check`** (`DeckCheck.svelte`, ADR 0095 §5) is public. It takes a link or a pasted list, shows the five-bucket report from `POST /deck-coverage`, and offers "Request these cards" (`POST /deck-requests`) to a session with a Discord user. Everyone else sees a sign-in prompt. The bot's `/c2-deck-check` reply links `#/deck-check?url=<link>`, or `#/deck-check` for a pasted list (`server/internal/bot/deck.go`, `fullReportURL`).
- **`#/decks`** (`MyDecks.svelte`, ADR 0110 §6) is the library, for a signed-in person only. It shows each deck's coverage line, the full report (`GET /me/decks/{id}/coverage`), rename (`PATCH`) and delete (`DELETE`). It has "Request missing cards" only for a link deck, by sending it to `#/deck-check?url=…`.
- **Limits.**
  - `POST /deck-coverage`: a per-IP bucket of 1 per 10 seconds with a burst of 3. The bot's token has its own bucket. A report is cached for ten minutes, by deck key (`deckReportCache`). The cache holds the report, not the fetched list.
  - `POST /deck-requests`: the IP bucket, then three asks per requester per rolling 24 hours, counted in the database.
  - `/me/decks*`: one shared bucket of 1 per second with a burst of 5.
  - The library holds 200 decks per person (`decklibrary.MaxDecks`). A new name past that is `ErrLibraryFull`. Updating a deck with the same owner and name is never refused.

### Decision

1. **One page, `#/decks`, public.** From top to bottom:
   1. **Check a deck:** the link and paste tabs, as on `#/deck-check` today.
   2. **The report:** "N of M cards play as printed" (ADR 0110 §6 item 2), with ADR 0095's five buckets underneath. The check page and the library now use the same wording.
   3. **Actions** under the report: **Request missing cards** (shown when the report has `manual` or `unreviewed` cards), and **Save to my decks**, with a name field filled with the deck's name or its first commander.
   4. **Your decks:** the library, as `MyDecks.svelte` shows it today, with rename, delete, the report and a request button on every deck (item 5).
   5. **Pre-built decks** ([owner answer 4](#owner-decisions-2026-10-02-questions)): the curated decks from `GET /decks`, each with its "as printed" line. They are read-only, and never copied into the library.
2. **Routes and links.**
   - `#/decks` is the page. `#/decks?url=<link>` fills in the link and runs the check, as `#/deck-check?url=` does today.
   - `#/deck-check` and `#/deck-check?url=<link>` parse to the same route, permanently, so every link the bot has already posted keeps working. A router test pins both. The bot does not change.
   - `App.svelte` lists the route as public. The header and Home each show one "Decks" link in place of "My decks" and "Deck check".
3. **Who can do what.**

   | Visitor | Check | Request | Save | Library |
   |---|---|---|---|---|
   | signed out | yes | "Sign in with Discord to request these cards" | "Sign in with Discord to save this deck" | hidden; one line says it needs Discord |
   | guest seat or spectator (no user) | yes | "Link Discord from your table's menu" | same | same |
   | signed-in person, either mode | yes | yes | yes | yes |
   | admin token | yes | the bot's path only; the page hides it | hidden (no user) | hidden |

   A guest is not offered "sign in": a Discord sign-in from the login flow replaces the browser's session, and a guest seat cannot be got back without a reclaim ticket. Their way to an account is "Link Discord" at their table (ADR 0051).
4. **Saving: `POST /me/decks`**, with body `{"url": …}` or `{"text": …}`, and an optional `"name"`.
   - **Caller:** `signedInUser`, else 403. It rides the existing `/me/decks*` bucket.
   - **A pasted list** is parsed (`parseList`) and saved as `text`.
   - **A link** is saved as the list the server fetched, with the canonical link in `decks.source_url`, exactly as `saveToLibrary` saves a link import (ADR 0110 owner answer 7). The deck-check cache keeps the fetched entries and name beside the report, so a save after a check fetches nothing. On a cache miss the handler spends one token from the public `POST /deck-coverage` bucket for the caller's IP before it fetches, so saving cannot get around the fetch limit. A Moxfield link fails with the same `paste_list` hint as the check.
   - **Name:** the body's `name`, trimmed, at most 100 characters (the rename limit). Without one, the fetched deck's name, then the first commander, then "Untitled deck" (`libraryFallbackName`).
   - **The library's update rule is unchanged** (same owner and same name updates in place), but it is no longer silent. The response is `{deck, replaced}`, where `deck` is the `GET /me/decks` entry with its coverage. The page already holds the library, so when the name matches a saved deck the button reads "Replace ‹name›" before anything is sent.
   - **The cap:** a new name past 200 decks is a 409 that names the cap: "Your deck library is full (200 decks). Delete one below to save this one." The report stays on screen. Replacing a deck is never refused.
5. **Requesting.** `POST /deck-requests` keeps its route, its limits and its outcomes, and gains a third input, `{"deck_id": …}`: one of the caller's own library decks. An ID that is not theirs is a 404, as on the other `/me/decks` routes.
   - The server re-parses the stored list and keys it the way the original would have been keyed: a deck with a `source_url` keys as that link (`moxfield:<id>`, `archidekt:<id>`), so it joins the same issue as a request made from the link, and any other deck keys as its list (`list:<hash>`).
   - The token's `requester` path does not accept `deck_id`, because the bot has no library.
   - Every library deck gets the button, not only link decks.
6. **Save and request are separate buttons,** and neither does the other's job. A request never saves, a save never files, and each has its own limit: three asks per 24 hours, and the 200-deck cap. A check by itself writes nothing, and saving is always the explicit button ([owner answer 2](#owner-decisions-2026-10-02-questions)).
7. **Signing in from the page brings you back to it.** Before a signed-out visitor follows "Sign in with Discord", the page stores the route in `sessionStorage["cmdctrl.afterSignIn"]`, and a pasted list in `sessionStorage["cmdctrl.afterSignIn.text"]`. `oauthComplete`'s login-page branch navigates there instead of `#/lobby` and clears both keys.
   - The value is accepted only if it parses as the `decks` route, so a stored value can never send the browser anywhere else.
   - It is client-only. The server's OAuth state carries nothing new.
8. **A deck seated for someone else is not saved to your library.** `saveToLibrary` takes the same guard as `recordLastDeck`: it saves only when the caller is seating their own seat (`p.GameID == id && p.PlayerID == body.PlayerID`). An admin setting someone else's deck seats it and saves nothing. The person whose seat it is can save it from the decks page.
9. **Past decks are not recovered.** A seat stores `deck_id` and `deck_name`, never the list, so a deck that was never saved cannot be rebuilt, and nothing is backfilled. The library's empty state says so: "No saved decks yet. Decks you save here, or import at a table while signed in, are kept." Owner decision 2 already removes the cause for the owner: a Discord admin plays from their own session (in either mode), so their uploads save.

### Rate limits, together

| Action | Limit |
|---|---|
| Check (`POST /deck-coverage`) | per IP, 1 per 10 s, burst 3; a cached deck costs a token but fetches nothing |
| Save (`POST /me/decks`) | the `/me/decks*` bucket (1/s, burst 5); one deck-coverage token only if it must fetch; 200 decks |
| Request (`POST /deck-requests`) | per IP, then 3 per requester per 24 h (database) |
| Library reads, rename, delete | the `/me/decks*` bucket |

### Migration / snapshot impact

None. The library and request tables are unchanged (`0005`, `0006`, `0008`). No snapshot impact.

---

## Shared machinery

- **Schema:** migration `0009` (§2 item 1) is the sprint's only schema change. It lands in the player-mode PR.
- **Routes:** two new, `PUT /me/admin-mode` and `POST /me/decks`, and one widened, `POST /deck-requests` with `deck_id`. Both new routes are under `/me/*`, which `deploy/Caddyfile`'s `@api` matcher, `client/vite.config.ts` and the service worker's `API_PATH` already carry. No new prefix.
- **Caller rule:** both new routes use `signedInUser`, so anything that is not a person gets one 403, never a 401 that `authFetch` would read as a sign-out (#1154).
- **Privacy:** `/me` tells a person only about themselves. The allowlist is still never served. `admin_allowed` answers only for the caller.
- **Environment variables:** none added. `CMDCTRL_DISCORD_ADMIN_USER_IDS` keeps its meaning. The AGENTS.md entry gains one sentence: being on the list makes a person *able* to switch admin mode on, and it starts off.

---

## Delivery

Each PR carries its own tests, its `docs/lobby.md` route docs and its AGENTS.md §5 lines.

| PR | What | Needs | Parallel with |
|---|---|---|---|
| 1 | **Server: player mode** (§2). Migration `0009` and `users.AdminModes` (boot load that fails closed, write-through); `Admins`, `isAllowlisted`, the mode term in `isAdminPrincipal`, and the guard test's second rule; `PUT /me/admin-mode` with its bucket and audit line; `/me`'s `admin_allowed`, `admin_mode` and `admin_mode_ends_at`; the 12-hour lapse and its once-a-minute sweeper (owner answer 1); `Hub.RebindUserSessions` with close 4001; `RevokeAll` clearing the mode; the same-answer test across every admin route. Docs: the two `docs/lobby.md` entries, the AGENTS.md endpoint and `CMDCTRL_DISCORD_ADMIN_USER_IDS` lines, and the S57 section and index row in `docs/sprints.md`. | — | 2, 3 |
| 2 | **Server: decks** (§3 items 4, 5 and 8). `POST /me/decks`, with the deck-check cache keeping entries and the fetch-token rule; `deck_id` on `POST /deck-requests`; `saveToLibrary`'s own-seat guard. Docs: `docs/lobby.md` and the AGENTS.md endpoint line. | — | 1, 3 |
| 3 | **Client: the signed-in home** (§1). `App.svelte` routing and the identity branch of `oauthComplete`; the login page reduced to Discord sign-in and the code box, without the token card; `#/admin`'s rules for the sessions it can tell apart without `/me`; the header's account menu and wordmark; the lobby's Join card, "your tables" (owner answer 3), the empty state and the removed command bar; Home's cards. Updates the e2e specs that use the login page. | — | 1, 2 |
| 4 | **Client: the Admin chip** (§2 items 8 and 9). The chip and the in-game menu item; `/me`'s new fields in `lib/admin.ts`, with re-asking on switch, on 4001 and on visibility; the 4001 handling in `ws.ts`; `#/admin` sending an allowlisted person to the lobby; the "Admin token" badge. | 1, 3 | 5 |
| 5 | **Client: one decks page** (§3). `#/decks` public, with check, report, request, save and library; the `#/deck-check` alias and its router test; the request button on every library deck; the sign-in return route; the pre-built section (owner answer 4); the explicit save button (owner answer 2); one "Decks" link in the header and on Home. | 2 | 4 |

PRs 1, 2 and 3 start at once: they share no files. PR 4 waits for PR 1 (the route and `/me`'s fields) and PR 3 (the header's account menu). PR 5 waits for PR 2. PR 5 also edits the header's links and Home's cards, which PR 3 edits, so whichever of them merges second rebases. That conflict is mechanical. No PR waits on an owner answer any more.

Exit check, on cmd-dev once PR 4 deploys, with a real allowlisted Discord account: sign in and land on the Lobby with the header; start in player mode; open a table and see no admin menu; switch to admin mode at the table and see the socket reconnect with the admin menu; switch back. Then, once PR 5 deploys: check a pasted deck while signed out, sign in, come back to the same deck, save it and request its missing cards.

---

## Consequences

- A signed-in person always has the header and its menu, and lands on the Lobby. The code box is where they already are.
- **Every allowlisted admin, the owner included, is in player mode after the deploy,** until they click the chip, and admin mode then lasts 12 hours at a time. An admin who plays at a table sees what the other players see unless they choose otherwise, and the chip always says which mode they are in.
- A switch reaches every tab and device of the person at the next request, and every open socket within a reconnect. No token changes.
- The shared token works exactly as before. The bot is unchanged.
- One page checks, requests and saves a deck, and every library deck can ask for its missing cards. The bot's links keep working.
- An admin seating someone else's deck stops filling their own library with it.
- Decks that were never saved stay lost, and the page says so rather than implying otherwise.

## Out of scope

- **Asking for a fresh Discord sign-in before admin mode turns on** (step-up authentication). It would make admin mode a defence against a stolen session, which §2 item 11 says it is not. It needs a recent-sign-in time in the session, which is a token change.
- **Player mode for the bot's `/c2-end`** (§2 item 10).
- **Server-side filtering of `GET /games`** for non-admins. §1 item 6's filter is presentation, and `redactMetaFor` already keeps the invites.
- **Refreshing a library deck from its link**, and any deck editor (ADR 0110 *Out of scope*).
- **Moving the pre-built picker onto ADR 0095's buckets** (ADR 0110 *Out of scope*). The decks page shows the pre-built decks' existing "as printed" line.

---

## Questions for the owner (answered)

These are the questions as asked. Each is a product choice the three decisions on #1992 left open, and (a) was the recommendation each time. The owner's answers follow.

1. **How long admin mode lasts (§2 items 1 and 3).** Admin mode is off by default, and the chip turns it on.
   - **(a) Recommended:** it switches itself off 12 hours after it was turned on (the length of a non-person session, `CMDCTRL_SESSION_TTL`'s default), like `sudo`. The chip shows when it ends. A sweeper once a minute ends lapsed modes and closes those sockets with 4001, so the admin menu goes away at the table too. An admin who forgets to switch back is a player again by the next game night.
   - (b) It stays on until the person switches it off, or signs out everywhere.
2. **Saving a checked deck (§3 item 6).**
   - **(a) Recommended:** an explicit **Save to my decks** button. A check writes nothing, so a deck you only looked at (a friend's list, a deck from a forum) never fills your library or uses up the 200.
   - (b) Every check by a signed-in person saves the deck automatically, and the library gains a way to tell checked decks from decks you have played.
3. **What the Lobby lists for a signed-in person who is not in admin mode (§1 item 6).**
   - **(a) Recommended:** their tables only: the tables they sit at, watch or created, including a seat held through another session, with an "Open" button. An admin in admin mode, and the token, see every table. Other pods' table names and players stop showing to people who are not at them.
   - (b) Every table in the room, as today. A person with no seat sees other pods' tables they cannot join.
4. **Pre-built decks on the decks page (§3 item 1).** The library holds decks you brought. A pre-built pick at a table is not a library row (ADR 0110 §5 item 5), which is part of why "My decks" looked empty.
   - **(a) Recommended:** the decks page lists the pre-built decks in their own read-only section, with each one's "as printed" line from `GET /decks`. They are never copied, so a curated deck that is edited later never leaves a stale copy behind.
   - (b) Seating a pre-built deck saves a copy of its list to your library, like an upload. The copy then drifts from the curated deck as it is edited.
   - (c) Neither. Pre-built decks stay in the lobby's picker only.

---

## Owner decisions, 2026-10-02 (questions)

The owner answered the four questions on 2026-10-02. Every answer was the recommended option (a).

1. **How long admin mode lasts.** It switches itself off 12 hours after it is turned on, like `sudo`. The chip shows when it ends, and a once-a-minute sweeper ends lapsed modes and closes those sockets with 4001, so the admin menu goes away at the table too (§2 items 1, 3 and 9; Delivery PRs 1 and 4).
2. **Saving a checked deck.** Checking a deck never saves it. "Save to my decks" is a separate, explicit button (§3 items 1 and 6; Delivery PR 5).
3. **What the Lobby lists.** When admin mode is off, the Lobby lists only the person's own tables: the ones they sit at, watch or created, with "Open" for a seat held through another session. Admin mode and the token see every table (§1 item 6; Delivery PR 3).
4. **Pre-built decks on the decks page.** They get their own read-only section, with each one's "as printed" line, and are never copied into the library (§3 item 1; Delivery PR 5).
