# ADR 0110 — Remember me: durable sign-in, account settings, admins and saved setups

**Status:** Proposed · 2026-10-02 · S55 — Remember me: durable sign-in, account settings, admins and saved setups (tracker [#1950](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1950))
**Owner decisions:** 2026-10-02, five, recorded on #1950 and quoted under [Context](#owner-decisions-2026-10-02). They are binding. This ADR designs them and asks only what they leave open: see [Questions for the owner](#questions-for-the-owner).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-02. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 35 remote heads: `origin/develop`, `origin/main` and 33 chore, docs, feat, fix, repro and wip branches. The highest number on any of them is 0109 (`0109-rule-gates-land-types-mana-and-cost-components.md`, on `origin/develop`). No branch has 0110, so this ADR takes **0110**.
**Amends:** [ADR 0051](0051-user-database.md) sub-PR 7's rule that "seat sessions keep `CMDCTRL_SESSION_TTL`, including those minted from an identity session" (§1 below reverses it), and the comment on `discord.Config.AuthorizeURL` that leaves `prompt=none` out on purpose (§2). It also extends [ADR 0051](0051-user-database.md) decision 7 (decks) and decision 8 (tablemates).
**Builds on:** [ADR 0004](0004-discord-identity.md) (Discord identity, the bot's admin session and allow-lists), [ADR 0044](0044-surviving-a-deploy.md) decision 3 (HMAC sessions, advisory `Revoke`), [ADR 0051](0051-user-database.md) decisions 3, 5, 6, 7 and 8, [ADR 0075](0075-table-settings-and-host-controls.md) (the table host and table settings), [ADR 0076](0076-tutorial.md) §2.2 (the practice table's forced settings and session swap), [ADR 0095](0095-deck-coverage-and-deck-requests.md) (the deck coverage report).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

The owner reported on 2026-10-02 that every visit to the site goes through Discord's authorization again. The tracker's diagnosis was right in outline. The audit below, on `origin/develop` at `8e354261`, confirms it and finds three more things the tracker did not name:

- **A signed-in player who has joined one table cannot paste a code for the next one.** `joinByCode` (`server/internal/lobby/http.go`) answers 409 to any session that is not `RoleIdentified`, and the client has nowhere to type a code anyway: `App.svelte` sends a `player` session away from `#/login` to `#/lobby`, and the lobby has no code box. Today the 12-hour expiry hides this, because the player is signed out and signs in again as `identified`. Once seat sessions last 30 days (§1), it becomes a dead end. The fix belongs in the sign-in PR.
- **Only the admin can create a table on the site.** `POST /games` is `auth.Middleware(c.Auth, auth.RoleAdmin)`, and `Lobby.svelte` shows the create form only to `role === "admin"`. A signed-in player creates a table only through the bot's `/c2-invite`. Owner decision 5's "create flow" therefore does not exist for players yet ([Q2](#questions-for-the-owner)).
- **Two tracker details were off.** Migration `0007` is taken (`0007_game_outcome.sql`), so S55's schema is `0008`. And tablemates already appear in one flow: `TablematePicker` in the lobby, for a signed-in player seated at the table (`canInviteTablemates`, `client/src/lib/tablemates.ts`). What is missing is the create flow, and the creator who is not seated.

### Owner decisions, 2026-10-02

Quoted from #1950:

1. **Stay signed in.** A session minted for a signed-in user keeps the identity lifetime and stays revocable (ADR 0051 decision 6), rather than dropping to 12h. This is a bug fix.
2. **Discord consent.** Send `prompt=none`, so a repeat sign-in with the same `identify` scope skips Discord's screen.
3. **Admin.** A server-side Discord ID allowlist gives a signed-in user admin rights. The shared token keeps working.
4. **Settings.** A signed-in user's settings sync to their account. Guests keep browser-only settings.
5. **Remember players.** All of these: who I play with (tablemates in the invite and create flows); my table setup (the last seat layout: bots, tiers, decks, table settings); my name and last deck; my saved decks, each showing how many of its cards are playable (deck coverage, ADR 0095).

---

## 1. Staying signed in

### What exists

There are ten session mints. Every one calls `c.Auth.Issue` and then `setSessionCookie`, so **every one replaces whatever the `cmdctrl_session` cookie held**, the 30-day identity session included. The client keeps exactly one session too (`session.ts`, `localStorage["cmdctrl.session"]`), so the new session also replaces the client's copy.

| Mint | Role | `UserID` | TTL today |
|---|---|---|---|
| `discordCallback`, login-page branch (`lobby/discord_oauth.go`) | identified | yes | `IdentityTTL` (30 days) |
| `discordCallback`, invite branch | player | yes | `SessionTTL` (12 h) |
| `finishDiscordLink` (`discord_oauth.go`) | player | yes | `SessionTTL` |
| `joinGame` (`lobby/http.go`) | player | when the caller is signed in (`signedInIdentity`) | `SessionTTL` |
| `joinByCode` (`http.go`) | player | when the caller is `identified` | `SessionTTL` |
| `myGameSession`, `POST /me/games/{id}/session` (`lobby/mygames_http.go`) | player | yes | `SessionTTL` |
| `createPractice`, `POST /games/practice` (`lobby/practice_http.go`) | player | when the caller has one | `SessionTTL` |
| `spectateGame` (`http.go`) | spectator | **never**, even for a signed-in caller | `SessionTTL` |
| `redeemSeatReclaim` (`http.go`) | player | never; it copies the seat's `DiscordID` but no user | `SessionTTL` |
| `adminLogin` (`http.go`) | admin | never | `SessionTTL` |

So a signed-in person who joins, spectates, practises or reclaims a seat holds a 12-hour session from then on, and nothing ever gives them their 30-day one back. `leavePractice` is the one exception: it restores the cookie from the token the client saved before the swap (ADR 0076 §2.2).

**What happens at expiry.** The cookie's `Expires` is the token's expiry, so the browser drops it at the same moment. In the client, `scheduleExpiry` fires `expireSession`, which clears the session and sets "Your session expired — please sign in again", and the router sends the user to `#/login`. A 401 from any `authFetch` does the same. There is no fallback to an older or longer session.

**How the credential travels.** `auth.CredentialFromRequest` reads the cookie first, then `Authorization: Bearer`, then `?token=`. `authFetch` sends the bearer token, but the cookie wins whenever it is present. A client that holds a different token from the cookie is therefore not believed. This is why `leavePractice` has to re-set the cookie on the server.

**Revocation.** `auth.WithRevocation` refuses any principal with a non-zero `UserID` issued at or before the user's `sessions_invalid_before`, on every HTTP request and every WebSocket upgrade, and `Hub.EvictUserSessions` closes open sockets. A session without a `UserID` cannot be revoked, only expired. That is why decision 1 ties the long lifetime to `UserID`.

### Decision

1. **One rule for every mint:** a session that carries a `UserID` gets the identity lifetime. A session without one keeps `SessionTTL`. The rule lives in one helper, `issueFor(c, p, source)`, which every mint calls instead of `c.Auth.Issue`. Its TTL:
   - **No source session** (the three Discord callback mints: `discordCallback`'s two branches and `finishDiscordLink`): `IdentityTTL`.
   - **A source session** (every mint that was handed a signed-in session: `joinGame`, `joinByCode`, `myGameSession`, `createPractice`, `spectateGame`): the source's own `ExpiresAt`. A seat session minted from an identity session with 3 days left also has 3 days left. Joining a table therefore never extends a sign-in on its own. Only a new Discord sign-in, or the renewal in item 5, does that. The new session is still a fresh token with a fresh `IssuedAt`, so the revocation watermark treats it like any other.
   - **No `UserID`:** `SessionTTL`, as today. Guest, admin-token and reclaim-ticket sessions are unchanged.
2. **A signed-in spectator stays signed in.** `spectateGame` reads the optional session the way `joinGame` does (`signedInIdentity`) and carries its `UserID` and Discord fields onto the `RoleSpectator` principal. The WS binding already passes `p.UserID`, so eviction reaches it. `signedInUser` and `signedInIdentity` (`mygames_http.go`) accept a spectator session that carries a `UserID`, so `/me/*` keeps working while someone watches a table.
3. **A signed-in seat can join the next table by code.** `joinByCode` accepts a `RolePlayer` or `RoleSpectator` session with a `UserID` as the person, exactly as `joinGame` already does through `signedInIdentity`. It keeps refusing a guest seat session (409, as today), because a guest has no identity to carry over. In the client, the login page's invite card is offered to any session with a `user_id`, not only to `identified` ones. `App.svelte`'s bounce from `#/login` exempts those sessions, and the lobby header gets a "join with a code" link to it.
4. **A reclaim ticket keeps its own rule.** `redeemSeatReclaim` carries a user only when the request also carries a valid signed-in session whose `UserID` owns that seat. In that case it is the same thing as `POST /me/games/{id}/session`, and it gets the same rule. Otherwise the ticket is the whole credential, and the session it mints has no user, as today.
5. **`POST /me/session`, and renewal** ([Q1](#questions-for-the-owner)). A new route sets the cookie to the caller's own credential and returns that session. It requires a `UserID` (403 otherwise). A revoked token fails `Validate` before the handler runs, so it cannot reinstall or renew itself. Under Q1 (a) it also re-issues the principal with a fresh `IdentityTTL` when the session is more than half spent, and the client calls it at load and then hourly. Under Q1 (b) it never extends anything and exists only for item 6. Either way it never touches a session that cannot be revoked.
6. **The client keeps the person's session aside.** When the client installs a session without a `user_id` over one that has one (the admin token, or a ticket for a seat that is not yours), it moves the old one to `localStorage["cmdctrl.identity"]` first. When the session without a user expires, the client reinstalls the saved one, if it is still good, through `POST /me/session` sent with the saved token as its bearer. The browser has already dropped the expired cookie, so the bearer is the credential the server reads. Then the user is still signed in rather than sent to `#/login`. Signing out or signing out everywhere clears the saved copy. The practice table's own restore (`practiceTable.ts`) is unchanged and runs first.
7. **`myDecks` answers 403, not 401,** to a session with no user, like the other `/me/*` routes (`signedInUser`, #1154). `authFetch` treats any 401 as an expired session and clears it. The client gates the call today, so this is latent, but a long-lived session makes a stray call more likely.

### What stays

- A player session stays bound to one `(GameID, PlayerID)`. "Collapse `player` into `identified`" (ADR 0051 *Deferred*) is still the cleaner end state and is still deferred: it rewires `WSAuthorizer`, the bot seats and reclaim. A signed-in person moves between their tables with `POST /me/games/{id}/session`, as today.
- The `Principal` and the HMAC claims gain no field. A spectator with a `UserID` uses fields that already exist.

### Migration / snapshot impact

None. No schema change, and nothing in a game snapshot. Sessions minted before the deploy keep their 12-hour expiry and are replaced as they are used.

---

## 2. Discord's consent screen

### What exists

`discord.Config.AuthorizeURL` (`server/internal/discord/oauth.go`) builds the authorize URL with `response_type=code`, the client ID, the redirect URI, `scope=identify`, the state and the PKCE challenge. Its comment says why `prompt` is missing: "requesting consent every time makes the 'scope of access' transparent to users who might not remember granting it." Decision 2 reverses that comment. `discordStart` (both flows) and `discordLink` call `AuthorizeURL`. When Discord sends `error=…` back, `discordCallback` answers 400 with the reason.

### What Discord does

- **Verified** in Discord's OAuth2 documentation (`docs.discord.com/developers/topics/oauth2`, read on 2026-10-02): "If a user has previously authorized your application with the requested scopes and prompt is set to `consent`, it will request them to reapprove their authorization. If set to `none`, it will skip the authorization screen and redirect them back to your redirect URI without requesting their authorization." `consent` is what Discord does today when the parameter is left out.
- **Not verified:** what `prompt=none` does for a user who has never authorized the app, or who authorized fewer scopes than requested. Discord's page does not say. A community reference (`docs.discord.food`) says `none` "requires previous authorization with the requested scopes" and names no error. Discourse users who tried `prompt=none` report that Discord still flashes its page and then redirects without a click. I could not test it with a real Discord account from here.

### Decision

1. `AuthorizeURL` takes a `Prompt` argument, and both sign-in flows send `prompt=none`. The scope stays `identify`.
2. **The callback copes with both possible behaviours.** The state entry records which prompt it used. If a `prompt=none` round comes back with any `error`, the callback does not show it. It parks a new state entry for the same flow (the same game and invite, or the same link), marked `consent`, and redirects to Discord with `prompt=consent`. An `error` on a `consent` round is a real refusal and is shown as it is today. So the retry happens at most once, and a first-time user sees Discord's screen once, whichever of the two behaviours Discord has.
3. **"Not you?"** With `prompt=none`, Discord silently uses whichever Discord account the browser is signed in to. The login page's signed-in card gains "Sign in with a different Discord account", which starts the flow with `prompt=consent` (`GET /auth/discord/start?prompt=consent`). Discord's consent screen has its own account switcher.
4. **The link flow** (`GET /auth/discord/link`) sends `prompt=consent`. It attaches an account to a seat, so the user should see which account it is.
5. **Scopes change.** If a later ADR adds a scope (`guilds`, ADR 0051 decision 5), the first sign-in after it needs consent again. Item 2 covers that: Discord refuses or prompts, and either way the user consents once.

### Migration / snapshot impact

None.

---

## 3. Admins

### What exists

- **The only admin credential is the shared token.** `adminLogin` mints `RoleAdmin` with a fresh `AdminID`, no `UserID` and `SessionTTL`. Its rate limit is the join bucket.
- **What admin can do.** Ten routes are wrapped in `auth.Middleware(c.Auth, auth.RoleAdmin)`: create, delete, archive and unarchive a game, mint a reclaim ticket, `GET /games/{id}/creator`, the pinned bug replay and game log, `POST /admin/users/{id}/revoke-sessions`, and the bot stats route (`botstats.go`). About twenty handler checks compare `p.Role == auth.RoleAdmin` directly:
  - Any game: start it, set any seat's deck, read its replay, attach a bug report to it.
  - `CanManageTable` (`host.go`), `CanRotateInvites` (`creator.go`) and `canInviteDM` (`invite_dm.go`). Only an admin may name a raw Discord snowflake in a DM invite.
  - The deck-request paths (`deckcheck.go`): the bot's larger rate bucket, and filing on a member's behalf.
  - Practice (`practice_http.go`) and `callerKey`'s one admin bucket.
  - In `WSAuthorizer.AuthorizeUpgrade`, an admin binds to **any** game, optionally as any seat (`?player=`), not read-only, with `Binding.Admin` set for the host gates. This is the moderator escape hatch.
- **The Discord allowlist exists only on the bot.** `CMDCTRL_DISCORD_ADMIN_USER_IDS` and `CMDCTRL_DISCORD_ADMIN_ROLE_IDS` are read by `server/internal/bot/config.go` for `/c2-end`. They are written into `bot.env` by the CD step "Sync bot env" (`.github/workflows/ci-cd.yml`), on production only.
- **A CD gap.** That step writes an admin variable only when it is set. If the Actions variable is deleted, the old value stays in `bot.env`. You can remove an admin only by setting the variable to a shorter list, never by clearing it. `scripts/set-server-env.sh` can write an empty value. The step just never asks it to.

### Decision

1. **Admin is a capability checked on every request, not a role baked into a token.** `lobby.Config` gains `Admins`, a set of Discord snowflakes parsed at boot. One function decides:

   ```go
   func (c Config) isAdmin(p auth.Principal) bool {
       return p.Role == auth.RoleAdmin ||
           (p.UserID != uuid.Nil && p.DiscordID != "" && c.Admins.Has(p.DiscordID))
   }
   ```

   The `UserID` requirement is load-bearing. `redeemSeatReclaim` copies a seat's `DiscordID` onto a session with no user, so without that check a reclaim ticket for an admin's seat would confer admin. With it, only a session minted from a real Discord sign-in qualifies, whatever its role (identified, player or spectator).
2. **Every admin check goes through one of two functions, and the difference is written down.** Most of the `RoleAdmin` checks ask "is this an operator?" Those become `c.isAdmin(p)`. A few ask "is this the shared server credential?" because they exist for the bot calling on other people's behalf, and those stay on the token alone, through `isServerCredential(p)` (`p.Role == auth.RoleAdmin`):
   - the bot's own rate bucket on `POST /deck-coverage` (`deckCoverageLimit`) and `callerKey`'s single admin bucket;
   - filing a deck request on behalf of a named member (`deckcheck.go`'s `requester`);
   - naming a raw Discord snowflake in a DM invite (`invite_dm.go`).

   A signed-in admin is a person. They keep their own per-user bucket, file requests as themselves, and invite by user ID. The ten middleware lines become `requireAdmin(c, h)`: the auth middleware with any role, then `isAdmin`, else 403. The exported helpers that have no `Config` take the answer as an argument: `CanManageTable(p, meta, admin bool)`, `CanRotateInvites` and `canInviteDM`. A guard test in `internal/lobby` fails if `auth.RoleAdmin` is compared anywhere outside those two functions and `adminLogin`, so a later route has to choose one.
3. **Removing an ID takes effect at the next request.** Nothing about admin is stored in a token, so a 30-day session never outlives its admin rights. The allowlist is read from the environment, and the environment changes only on a deploy, which restarts the server and drops every socket. So a removal takes effect when the deploy lands, for HTTP and WebSockets alike. Revoking the person's sessions as well is the existing admin route.
4. **The client learns it from `/me`.** `GET /me` adds a computed `admin: bool` beside the principal. It is not part of the token. The client's admin checks (`Lobby.svelte`'s `isAdmin` and `canManageBots`, `Game.svelte`'s `isAdmin`) move to one helper, `isAdmin(session)`, that reads it. The client fetches `/me` once per installed session.
5. **WebSockets** ([Q4](#questions-for-the-owner)). Under the recommended answer, `WSAuthorizer` gives an admin capability the same reach as the shared token:
   - A player or spectator session at its own game binds as today, with `Binding.Admin` set.
   - Any user-bearing session that names a different `?game=` takes the admin branch: any game, optionally as any seat.
   - An `identified` session that is admin takes the admin branch instead of being refused.
   - A non-admin is unchanged.
6. **The shared token stays.** It is how the bot calls the server over loopback, and how an operator gets in with no Discord. Nothing about it changes. `createGame` already stamps `games.created_by` from the caller's `UserID`, so a table an allowlisted user creates is attributed to them, and one the token creates is not.
7. **Audit.** Every admin-gated route logs `admin_user_id` (or `admin_id` for the token) with the action, at Info.
8. **Configuration** ([Q3](#questions-for-the-owner)). Under the recommended answer, the server reads the same variable as the bot, `CMDCTRL_DISCORD_ADMIN_USER_IDS`: comma-separated snowflakes, empty or unset meaning no allowlisted admins. A malformed entry fails the boot, because a typo would otherwise silently deny someone. The boot log states the count, never the IDs.
   - **CD** gains "Sync server env (admin allowlist)" on **both** hosts. Each GitHub environment (`dev`, `prod`) has its own value. The step **always** writes the key, empty included, so clearing the variable removes every allowlisted admin.
   - The bot's step gets the same always-write fix for both of its admin keys.
   - Role IDs stay bot-only. The server cannot see guild roles without a member lookup per request, and owner decision 3 names user IDs.

### Migration / snapshot impact

None. No schema change, no token change, nothing in a snapshot.

---

## 4. Settings on the account

### What exists

`client/src/lib/settings.ts` is one versioned object (`SETTINGS_VERSION = 15`) in `localStorage["cmdctrl.settings.v1"]`. It has an export and import (`exportSettings` / `importSettings`, both run through `migrate`) and a fingerprint for bug reports. `migrate` merges a stored blob group by group over `defaultSettings()`. Unknown keys inside a known group survive, an unknown group is dropped, and `__version` is restamped to the running client's version. The fields:

| Group | Fields | Nature |
|---|---|---|
| `audio` | `muted`, `masterVolume`, `effectsVolume`, `musicVolume` | per device: speakers versus headphones |
| `animations` | `enabled`, `speed`, seven per-effect toggles | `enabled` follows the OS reduced-motion signal at runtime (the `matchMedia` listener), so it is per device. The rest are per person. |
| `display` | `cardSize`, `handLayout`, `tableLayout`, `opponentDetail`, `expandActivePlayer`, `expandStyle` | per device: they depend on screen size (#956) |
| `display` | `theme`, `stackStyle`, `hoverDelayMs`, `showOpponentHandCount` | per person |
| `gameplay` | every field (stops, autopass, strict mana, response categories, bluffs, `adminOverrides`, `showBotReasoning`, `highlightLegalActions`) | per person: this is how they play |
| `shortcuts` | `enabled`, `bindings` (overrides only) | per person |
| `accessibility` | `reduceMotion` (follows the OS signal), `textScale` | per device |
| `accessibility` | `colorblindPalette`, `alwaysShowFocus` | per person |

**The practice table** (`practiceTable.ts`, ADR 0076 §2.2) forces four fields: `strictMana` off, `autoPassPriority` off, `tableLayout` quadrant and `cardSize` medium. It captures the player's values in a `localStorage` record first and writes them back on every exit path, including the next page load after a crash. It does this through `settings.update`, so a naive sync would upload the forced values to the account and spread them to every other device.

### Decision

1. **Storage.** Migration `0008` adds:

   ```sql
   CREATE TABLE user_settings (
       user_id    TEXT PRIMARY KEY REFERENCES users(id),
       version    INTEGER NOT NULL,  -- the client's SETTINGS_VERSION that wrote it
       revision   INTEGER NOT NULL,  -- bumped on every write, for If-Match
       body       TEXT NOT NULL,     -- JSON object: the synced subset
       updated_at INTEGER NOT NULL   -- Unix ms
   );
   ```

   The server does not interpret the body beyond its shape. The client's `migrate` stays the one schema validator, because a server copy of the schema would drift from it.
2. **API.**
   - `GET /me/settings` returns `{version, revision, settings, updated_at}`, or `{revision: 0}` when there is no row.
   - `PUT /me/settings` takes `{version, settings}` with `If-Match: <revision>`. It answers 412 with the current copy when the revision has moved, so two tabs or devices never silently overwrite each other.
   - Both use `signedInUser`, so a guest, the admin token and a no-database deployment get 403 (never 401, #1154), and the client treats that as "browser-only".
   - **Validation:** the body is a JSON object of at most 32 KiB and depth 4. `version` is an integer from 1 to 1000. **A PUT whose `version` is below the stored one is refused with 409**, "a newer version of the site saved these settings; reload". Without that rule, a stale tab from a cached service worker would drop the groups it does not know and stamp an older version over a newer client's copy. A client that reads a copy newer than its own `SETTINGS_VERSION` applies it through `migrate` and then stops writing until the page reloads.
   - **Rate limit:** a per-caller bucket of 1 per second with a burst of 5. The client debounces writes by one second.
3. **What syncs** ([Q5](#questions-for-the-owner)). Under the recommended answer, the per-person rows of the table above sync and the per-device rows never leave the browser. The split is one exported list in `settings.ts` (`SYNCED_FIELDS`), with a test that every field of `Settings` is classified. A new setting cannot ship unclassified.
4. **Merging at sign-in** ([Q6](#questions-for-the-owner)). When a browser first installs a session with a `user_id`:
   - If there is no account copy, the browser's per-person values are uploaded as the first copy.
   - If there is one, the recommended answer applies it over the browser's per-person values and shows a toast, "Using your account's settings · Keep this browser's instead", for the rest of the visit. Taking that option uploads the browser's values.
   - The device's own fields are untouched either way.
   - Signing out leaves the browser holding the last values it had. It does not revert.
5. **The practice table still restores.** While a practice record exists, the sync layer uploads the settings *as they will be after the restore* (`withSettings(current, record.saved)`), never the forced values. A copy that arrives from the account during practice updates `record.saved` for the forced fields and the live settings for the rest. So the account never sees the tutorial's values, a crash mid-tutorial cannot leave them on the account, and the restore keeps working exactly as ADR 0076 specifies. `tableLayout` and `cardSize` are per device under Q5 (a) and are not synced at all.
6. **Guests** keep `localStorage` only, unchanged.

### Migration / snapshot impact

Migration `0008` (shared with §5 and §6, see [Shared machinery](#shared-machinery)). No snapshot impact: settings are client state and never enter a game.

---

## 5. Players and setups

### What exists

- **Tablemates.** `GET /me/tablemates` (`lobby/tablemates.go`) is ADR 0051 decision 8's self-join of `seats`, newest shared table first, with display names and avatars. The lobby shows `TablematePicker` (DM invite, `POST /games/{id}/invites/dm`) only on a lobby-state table to a signed-in player seated at it (`canInviteTablemates`). That function's comment says the creator cannot be recognised because `games.created_by` is not on the wire. That is now stale: `redactMetaFor` sets `is_creator` per viewer (#1098), and `api.ts` already has the field.
- **Creating a table** is admin-only on the site (see [Context](#context)): a name and nothing else (`createGameRequest`: `name`, `host_discord_id`). Bots are added after creation, one at a time (`POST /games/{id}/seats/bot`: tier, curated deck ID or decklist, name), by an admin or a seated player. Table settings (`game.TableSettings`: undo limit and scope, starting life, commander damage, bot pace, allow spawn) are changed by the host or admin through `PATCH /games/{id}/settings`.
- **Nothing remembers a setup.** Every table starts from `DefaultTableSettings()` and an empty seat list.
- **Name and last deck.** `Join.svelte` starts with an empty name field every time. A signed-in person is not asked: the seat takes the Discord display name (ADR 0051 sub-PR 4). Nothing remembers the last deck. The lobby's deck panel offers the library (`YourDecksPicker`), the pre-built decks and a paste box, with nothing preselected.

### Decision

1. **The last setup is captured when a table starts.** In `Lobby.Start`, the server writes one row per person to migration `0008`'s `table_setups`:

   ```sql
   CREATE TABLE table_setups (
       user_id    TEXT PRIMARY KEY REFERENCES users(id),
       body       TEXT NOT NULL,     -- JSON, built by the server, never by a client
       game_id    TEXT,              -- the table it was captured from; not a foreign key
       updated_at INTEGER NOT NULL
   );
   ```

   - **Whose row.** The game's creator, if it has one. Otherwise the person who pressed start, if they have a `UserID`.
   - **The body** is the table settings as a complete `SettingsPatch`, every bot seat as `{tier, deck_id, name}` in seat order, and the user IDs of the other signed-in humans who sat there.
   - **Why at start.** That is when a layout is final. A lobby that never started is not a setup anyone chose.
   - **No validation of client JSON.** The server writes the body from the game it starts, so there is none to validate.
2. **Applying a setup.** `POST /games/{id}/setup` with `{"from": "last"}`, host or admin (`CanManageTable`), on a lobby-state table only.
   - It applies the settings through the same `UpdateSettings` path as the PATCH, then adds the bots through the same pipeline as `addBot`, up to the free seats.
   - A bot whose tier is unavailable on this server, or whose curated deck no longer exists, is **skipped and named** in the response (`skipped: [{name, reason}]`). It is never silently downgraded, as with `addBot`'s 422 today.
   - `POST /games` accepts the same `"setup": "last"` field to do this at creation.
   - It rides the deck rate bucket, since it seats decks.
3. **Tablemates in the create flow.**
   - Right after creation, the creator sees the `TablematePicker` for the new table. The people from the last setup come first, marked "at your last table", and the rest follow in recency order.
   - `canInviteTablemates` also admits the creator (`is_creator`), which the server's `canInviteDM` already allows.
   - The DM route, its limits and its message are unchanged.
4. **Who may create** ([Q2](#questions-for-the-owner)). Under the recommended answer, `POST /games` opens to any signed-in person: `signedInUser`, or `isAdmin`. It gets a per-caller bucket of 1 per 30 seconds with a burst of 3, and a cap of 3 lobby-state tables per creator. A fourth is a 409 that names the open ones. The lobby shows the create form to any signed-in person.
5. **Last deck.** Migration `0008` adds `users.last_deck TEXT`, JSON `{"kind": "library" | "prebuilt", "id": "…"}`. A pre-built pick is not a library row, so `seats.deck_id` cannot record it.
   - It is written whenever a signed-in person seats a deck: an upload saved to the library, a library deck, or a pre-built deck.
   - The lobby's deck panel preselects it. It is never seated automatically, because a seated deck is visible to the table and a stale choice should cost a click, not a mulligan.
   - A library deck that has since been deleted, or a pre-built one that has gone, is simply not preselected.
   - Guests get the same preselection from `localStorage["cmdctrl.lastDeck"]`, pre-built decks only.
6. **Name** ([Q8](#questions-for-the-owner)). Under the recommended answer, a signed-in seat keeps the Discord display name, as ADR 0051 sub-PR 4 made it. "Remember my name" is for guests: `Join.svelte` and the login page's code box pre-fill the last name a guest typed, from `localStorage["cmdctrl.guestName"]`.

### Migration / snapshot impact

Migration `0008`. **No snapshot impact.** Applying a setup goes through `UpdateSettings` and the bot-seat pipeline, which already produce restore points of a known shape. A setup is never stored in a game, and a game never reads one.

---

## 6. Saved decks and their coverage

### What exists

- **The library** is migration `0005`'s `decks` table (owner, name, source format `moxfield` or `text`, the pasted source, commanders, card count, timestamps).
  - `saveToLibrary` (`lobby/http.go`) writes it on every upload by a signed-in person. The update rule is same owner and same name.
  - `GET /me/decks` lists id, name, commanders, card count and `updated_at`.
  - `POST /games/{id}/decks/{deck_id}` seats one.
  - `YourDecksPicker` shows the list in the lobby.
- **Missing:**
  - Deleting a deck, renaming one, and any page outside a table.
  - Any coverage information.
  - **A link import is not saved at all.** `saveToLibrary` skips `format: "url"`, because the source is a link rather than a list. Links are how most decks arrive.
- **The coverage report** is `deckcoverage.Build(idx, Deck)`, ADR 0095's five buckets (`manual`, `unreviewed`, `caveats`, `automated`, `no_effect`), counted by distinct card, plus unknown names and validation notes. It calls `catalog.Build(idx)` on every call. `catalog.Build` is deliberately uncached ("a stale cache would be a worse bug than that pass is a cost"). Re-parsing a stored hundred-line list is microseconds (ADR 0051 decision 7).
- **A second coverage shape.** The pre-built deck picker uses `decks.Coverage` (full, caveats, unreviewed, unregistered, basics), with the headline "N of M nonbasic cards as printed" (`prebuiltDecks.ts`). It is built from the same `catalog.Build` verdicts but grouped differently.

### Decision

1. **Coverage is computed when it is read, never stored.** `GET /me/decks` adds `coverage` to each deck: the five bucket counts, the unknown count and `as_printed` (see item 2). `GET /me/decks/{id}/coverage` returns the full ADR 0095 report for one deck, so the library can offer "Request these cards" through the existing `POST /deck-requests`.
   - **Why not store it.** The verdicts change with every deploy that touches a card, and with every Scryfall reload. A stored count would be wrong after the next deploy and nothing would notice. That is exactly the drift ADR 0051 decision 7 refused for the card list itself.
   - **Cost.** Computing on read costs one `catalog.Build` per request plus one parse per deck, so the list route builds the verdict map once per request and reuses it across that request's decks. There is no cross-request cache: by `catalog.Build`'s own reasoning, a cache is a staleness bug waiting for a key someone forgets.
   - **Limits.** A per-caller bucket of 1 per second with a burst of 5 bounds the work. A library is capped at 200 decks per person, and the 201st save is refused with a message.
2. **"Playable" means "plays as printed".** The headline is "N of M cards play as printed", where N counts `automated` plus `no_effect` and M counts every resolved distinct card. Below it, `caveats` is shown as "simplified", `unreviewed` as "not checked yet" and `manual` as "you resolve these by hand". This is the strict reading. It matches the pre-built picker's "as printed" and the owner's standing preference for the rules-faithful number, and a caveated card does not do everything its text says. The library uses ADR 0095's `deckcoverage` report, which owner decision 5 names. Moving the pre-built picker onto the same report is [out of scope](#out-of-scope).
3. **Delete and rename.**
   - `DELETE /me/decks/{id}` removes the row. In the same transaction it sets `seats.deck_id` to NULL wherever it pointed there, because `0005` declares the foreign key without `ON DELETE`. The seat's `deck_name` keeps the label for history.
   - `PATCH /me/decks/{id}` with `{name}` renames. A name the owner already uses is a 409, because the upsert rule keys on the name.
   - Both are owner-only. Someone else's deck answers 404, never 403, so an ID reveals nothing.
4. **Link imports** ([Q7](#questions-for-the-owner)). Under the recommended answer, a `url` upload by a signed-in person saves the list it fetched, in the source format the fetcher returned (Moxfield JSON or text), and migration `0008` adds `decks.source_url TEXT` for the link. Re-seating it re-parses the stored list and never calls the network. Fetching the deck again from its link is a later, explicit button.
5. **A "My decks" page**, `#/decks`, signed-in only. It shows the library with each deck's coverage line, rename, delete, the full report and the request button. The lobby's `YourDecksPicker` shows the same coverage line.

### Migration / snapshot impact

Migration `0008` (`decks.source_url`, under Q7 (a)). No snapshot impact: a seated deck is the resolved card list it always was.

---

## Shared machinery

- **One migration, landed first.** Migration `0008_remember_me.sql` holds all of S55's schema: `user_settings`, `table_setups`, `users.last_deck`, and `decks.source_url` under Q7 (a). It lands in one small PR with its store code and tests, before the features that use it. The feature PRs can then run in parallel without fighting over the next migration number.
  - Every column is additive.
  - The two new tables reference `users(id)`.
  - `table_setups.game_id` is deliberately not a foreign key, so deleting a game does not delete someone's setup.
  - Times are Unix milliseconds, like `0002` to `0007`.
- **One caller rule.** Every new `/me/*` route uses `signedInUser`, which §1 widens to a signed-in spectator. That gives one 403 for "not a person", and no new 401 that `authFetch` would read as a sign-out.
- **No new route prefixes.** Everything new sits under `/me/*`, `/games/*` or `/auth/*`, which `deploy/Caddyfile`'s `@api` matcher, `client/vite.config.ts` and the service worker's `API_PATH` already carry.
- **Privacy.** Each new route returns only the caller's own rows. A setup lists the user IDs of the caller's own tablemates, the same class of data `GET /me/tablemates` already returns to the same caller. The admin allowlist is never served: `/me` says only whether the caller is admin. `POST /me/session` returns a token, which the client's redaction (`redact.ts`) and the server's (`util/redact`) already treat as a credential.
- **No deployment without a database changes.** With no database, no principal carries a `UserID`. Every new `/me/*` route then answers 403, the client stays browser-only, and the admin allowlist grants nothing (`isAdmin` needs a `UserID`). That is a supported state, and the boot log says so once when the allowlist is non-empty.

### Environment variables

These are the env var edits the implementation PRs make to AGENTS.md §5. The ADR names them here and changes no docs.

- `CMDCTRL_DISCORD_ADMIN_USER_IDS` (Q3 (a)), or a new `CMDCTRL_ADMIN_DISCORD_IDS` (Q3 (b)): the server's admin allowlist. It is documented in the server's env list, and the bot's entry notes that the server reads it too. CD provisions it on both hosts and always writes it.
- `CMDCTRL_SESSION_TTL`: the doc changes to "guest, admin-token and reclaim-ticket sessions".
- `CMDCTRL_IDENTITY_TTL`: the doc changes to "every session that carries a user", and states the inheritance rule (§1 item 1) and renewal (§1 item 5).

No other variable is added.

---

## Delivery

Each PR carries its tests, its `docs/lobby.md` route docs and its AGENTS.md lines.

| PR | What | Needs | Parallel with |
|---|---|---|---|
| 1 | **The sign-in fix (bug, first and small).** §1 items 1 to 4 and 7: `issueFor` and the TTL rule on every mint, the signed-in spectator, `joinByCode` and the login page accepting a signed-in seat, the reclaim ticket's matching-user case, and `myDecks` 403. §2 in full: `prompt=none`, the one-shot `consent` retry, "a different Discord account", and the link flow on `consent`. The ADR 0051 sub-PR 7 pointer line. The `docs/sprints.md` S55 section and index row. | — | — |
| 2 | Renewal and the saved identity (§1 items 5 and 6): `POST /me/session` and the client's renew and reinstall. | PR 1 (same mint files); Q1 | 3, 4 |
| 3 | Admins (§3): `Config.Admins`, `isAdmin`, `requireAdmin`, the guard test, `/me`'s `admin`, `WSAuthorizer` (Q4), the client's `isAdmin(session)`, the CD step on both hosts, and the bot step's always-write fix. | PR 1; Q3, Q4 | 2, 4 |
| 4 | Schema: migration `0008` and its store packages (`usersettings`, `tablesetups`, the `users.last_deck` and `decks.source_url` accessors), with migration and store tests. No routes. | — (Q7 decides one column) | 1, 2, 3 |
| 5 | Settings sync (§4): the `/me/settings` routes, `SYNCED_FIELDS` and its classification test, the client's sync layer and sign-in merge, and the practice-table rule with tests for a forced field never being uploaded and a crash mid-tutorial. | PR 4; Q5, Q6 | 6, 7 |
| 6 | Saved decks (§6): coverage on read, the per-deck report, delete, rename, link imports (Q7), the 200-deck cap, `#/decks` and the picker's coverage line. | PR 4 | 5, 7 |
| 7 | Players and setups (§5): capture at start, `POST /games/{id}/setup` and `POST /games`'s `setup`, opening `POST /games` (Q2), tablemates in the create flow and `canInviteTablemates`'s creator case, last deck and guest name pre-fill (Q8). | PR 3 (`isAdmin`), PR 4; Q2, Q8 | 5, 6 |

PR 1 goes first and alone: it is the bug the owner reported, it needs no answer to any question below, and it changes no schema. PR 4 has no dependency and can start at once beside PR 1. PRs 2 and 3 start when PR 1 merges, since both edit the mint and auth paths it touches. PRs 5, 6 and 7 start when PR 4 merges, and PR 7 also waits for PR 3. PR 6 and PR 7 both edit the lobby's deck panel (the coverage line and the last-deck preselection), so whichever merges second rebases. That conflict is mechanical.

Exit criteria 2 ("still signed in after 12 hours, and a repeat sign-in skips Discord's screen") is met by PR 1 alone, and is checked on cmd-dev with a real Discord account once PR 1 deploys, because Discord's unverified `prompt=none` behaviour (§2) can only be seen there.

---

## Consequences

- A signed-in person stays signed in across tables, spectating, practice and deploys for the identity lifetime, and indefinitely under Q1 (a). Every one of those sessions is revocable.
- A repeat sign-in is one click with no Discord screen. A first sign-in, and the first after a scope change, still shows Discord's screen once.
- The owner, and anyone on the allowlist, is an admin without the shared token, from any session. Taking someone off the list takes effect at the next deploy. The token still works, and the bot still uses it.
- Clearing the bot's admin variables now actually clears them.
- Settings follow a person across devices. Screen-shaped settings stay with the screen.
- A table can be set up like the last one in one click. The creator sees their tablemates the moment a table exists.
- The deck library shows how much of each deck plays as printed, always against the running build, and gains delete, rename and link imports.
- A long-lived session is a longer-lived credential in `localStorage`. That is the trade decision 1 accepts, bounded by revocation (ADR 0051 decision 6), and unchanged in kind from the 30-day identity session that already exists.

## Out of scope

- **Collapsing `player` into `identified`** (ADR 0051 *Deferred*). A browser still holds one seat at a time.
- **Several named setups.** One "last setup" per person, as decision 5 says. Named presets would be one more table on top of `table_setups`.
- **Guild-role admins on the server.** They would need a Discord member lookup per check.
- **An account-wide sign-in history, or a list of active sessions** beyond "sign out everywhere".
- **Refreshing a library deck from its link**, and any deck editor (ADR 0051 decision 7: Moxfield and a text box are the editors).
- **Moving the pre-built deck picker onto ADR 0095's buckets.** The two coverage shapes agree on the verdicts and differ in grouping. Unifying them changes the picker's wording and is its own small change.
- **Syncing a guest's settings,** or carrying a guest's browser copy into an account on sign-up beyond Q6's merge.

---

## Questions for the owner

Each question is a product choice the five owner decisions leave open. (a) is the recommendation each time.

1. **Renewal (§1 item 5).** Decision 1 keeps the identity lifetime, which is 30 days from the Discord sign-in. A regular player would still sign in again about once a month.
   - **(a) Recommended:** renew on use. A signed-in session more than half spent is re-issued for a fresh 30 days (`POST /me/session`), so someone who plays at least every two weeks is never asked again. Revocation (sign out everywhere, the admin route) stays the way to end it, and the renewal route refuses any session without a user.
   - (b) No renewal. 30 days after the Discord sign-in the player signs in again: one click, and with `prompt=none` no Discord screen.
2. **Who may create a table on the site (§5 item 4).** Today only the admin can, and a player creates a table only through the bot's `/c2-invite`. Decision 5's "create flow" and "my table setup" assume someone other than the admin creates tables.
   - **(a) Recommended:** any signed-in person may create a table, limited to 3 open lobby tables each and 1 creation per 30 seconds. They are its creator, so they can rotate its invites, DM tablemates and `/c2-end` it, and their setup is remembered.
   - (b) Admins only: the shared token and the allowlist from decision 3. Setups and the create-flow tablemates then serve admins only, and players keep using `/c2-invite`.
3. **The allowlist's source (§3 item 8).** The bot already has a Discord user-ID allowlist for `/c2-end`, `CMDCTRL_DISCORD_ADMIN_USER_IDS`.
   - **(a) Recommended:** one list. The server reads the same Actions variable as the bot, and CD writes it to both hosts (each GitHub environment can hold its own value). An admin on Discord is an admin on the site, and removing someone removes them from both.
   - (b) A separate server variable, `CMDCTRL_ADMIN_DISCORD_IDS`, so who may `/c2-end` from Discord and who is a site admin can differ.
4. **A signed-in admin at a live table (§3 item 5).** The shared token can open any game's WebSocket and act as any seat. This is the moderator escape hatch, and it lets the holder see and play another player's hand.
   - **(a) Recommended:** full parity. An allowlisted person can do over the WebSocket exactly what the token can, from their own session, and every such binding is logged with their user ID. Decision 3 says "admin rights", and an admin who must fetch the token to fix a stuck table has not been given them.
   - (b) HTTP only. An allowlisted person gets every admin route, and is admin at a table they sit at (host powers). Binding to another game or acting as another seat stays with the shared token.
5. **Which settings follow the account (§4 item 3).**
   - **(a) Recommended:** the split in §4's table. These stay with the device: volumes and mute, card size, hand and table layout, opponent detail, the two expansion settings, text scale, and the two settings that follow the OS reduced-motion signal. Everything else syncs: every gameplay setting, shortcuts, theme, stack style, hover delay, the remaining animation toggles, colour-blind palette and focus outlines.
   - (b) Everything syncs. A phone and a desktop then share card size and layout, which #956 showed depend on the screen.
6. **Signing in on a browser that has its own settings (§4 item 4).** The account already has a copy, and this browser was changed while signed out.
   - **(a) Recommended:** the account's copy wins, and a toast for the rest of the visit offers "keep this browser's settings instead", which uploads them. It is predictable: the same person sees the same settings everywhere, and nothing is lost without a choice.
   - (b) Whichever copy was changed most recently wins, with no prompt. This needs a change time on every browser copy and can surprise someone whose old laptop happens to be newest.
7. **Decks imported from a link (§6 item 4).** Today a link import seats the deck but does not save it, because the library stores the list itself (ADR 0051 decision 7) and a link is not a list.
   - **(a) Recommended:** save the list as it was fetched, with the link beside it (`decks.source_url`). Re-seating it never calls the network, and the deck is in the library like a pasted one. Fetching it again from the link is a later, explicit button.
   - (b) Keep today's rule. Only pasted decks are saved, and a link deck must be imported again each time.
8. **"My name" for a signed-in person (§5 item 6).** A signed-in seat already takes the Discord display name, and the server ignores a typed name so that the label Discord vouched for cannot be changed.
   - **(a) Recommended:** keep that. "Remember my name" pre-fills the name a guest last typed, in their browser.
   - (b) An account-level seat name. A signed-in person may set a name to sit under in place of their Discord display name, stored on the user and shown on every seat they take. The Discord name is still shown on hover.
