# ADR 0156 — A front door: one site shell, a front page, the C2 Discord, and sign-up gated by a Discord role

**Status:** Accepted · 2026-10-13 · S71
**Issue:** [#3177](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3177)
**Numbering:** every ADR file on `origin/develop` and on every remote head was listed; the highest number anywhere was **0155**, so this is **0156**. The claim comment on #3177 reserved it.
**Builds on:** [ADR 0004](0004-discord-identity.md) (Discord identity), [ADR 0051](0051-user-database.md) (users, identities, the revocation watermark, sealed refresh tokens), [ADR 0076](0076-tutorial.md) (the practice table), [ADR 0110](0110-remember-me.md) §1 (session lifetimes and renewal), [ADR 0125](0125-a-walkthrough-that-keeps-up.md) (labels and first-use hints), [ADR 0148](0148-capacity-limits-for-a-public-launch.md) (capacity caps).
**Amends:** [ADR 0092](0092-public-roadmap-and-site-portal.md) decision 5 and [ADR 0112](0112-signed-in-home-player-mode-and-one-decks-page.md) §1 item 5 (`#/home` is retired; the front page and the header are the portal), [ADR 0125](0125-a-walkthrough-that-keeps-up.md) §3 (on site pages a hint no longer floats over the page's controls), and the comment on `discordScopes` in `server/internal/discord/oauth.go` (`guilds.members.read` is asked for when the gate is on).

---

## Context

The non-game pages grew one feature at a time, and it shows. Screenshots of develop at `d14ddbf6` (2026-10-13), desktop and 820 px tablet:

- **Navigation is flat and doubled.** The header lists Lobby, My games, Decks, Catalog, Roadmap, Home and Admin as equals. `#/home` is a thirteen-tile site map that repeats the header, four of whose tiles say "sign in first".
- **The front door is two pages.** A signed-out visitor lands on `#/login`, which has its own centred layout, no header, and four loose links stacked over a card fan. `#/home` is the other half.
- **No page frame.** Each route builds its own: prose widths of 60, 62, 64, 68 and 72ch, five breakpoints (480, 560, 600, 700, 760 px), titles in four styles ("Tables" with no eyebrow, "Card catalogue" with one, "live now" and "home" lowercase, About's orange monospace section heads). There is no shared card or link-button primitive.
- **Tips cover the work.** The first-use hint popover sits over the page's main control on most first visits: Lobby's create button, the catalog filters, the deck check.
- **The roadmap at tablet width is about 44,000 px tall.**
- **The C2 Discord server is nearly invisible.** Its invite is a client constant (`client/src/lib/about.ts`) rendered once, as one word in About's Feedback list.
- **Anyone with a Discord account gets an account.** The OAuth scope is `identify`; `discordCallback` upserts a `users` row for every caller. The only lever is an admin's `revoke-sessions`, and the person can sign straight back in.

The owner wants the site tidied, the Discord server made the place a player goes to join, and sign-up to run through it: a player asks on the Discord and is let in.

## Decision

### 1. The gate: a Discord role, checked at sign-in and at renewal

Owner answers 1–3 and 7.

1. **Config.** Three new server variables:
   - `CMDCTRL_DISCORD_COMMUNITY_GUILD_ID`: the C2 server's snowflake.
   - `CMDCTRL_DISCORD_PLAYER_ROLE_ID`: the role that admits a person. On the C2 server this is **Elder Dragon Highlander** (owner, 2026-10-13; see "The C2 server's roles" below).
   - `CMDCTRL_DISCORD_INVITE_URL`: the public invite link (`https://discord.gg/…`). It replaces the client constant, so the link can change without a release.

   **Amendment (owner, 2026-10-13): an exemption list.** `CMDCTRL_DISCORD_GATE_EXEMPT_USER_IDS` is a comma-separated list of Discord user ids that sign in and renew without the role and without a member check, the same way an admin on the allowlist does. It keeps the accounts that existed before the gate went live from being locked out while they're given the role. A malformed entry fails the boot and the deploy, and CD writes the variable on every deploy, empty included.

   **The gate is on when the guild and the role are both set, and off when both are unset.** One without the other fails the boot naming the missing variable, as a malformed snowflake does. With the gate off, sign-in behaves exactly as today; that is every dev loop and test, and any environment the owner leaves open. The invite URL is independent: set, it is shown; unset, every Discord placement is hidden. CD writes all three from Actions variables of the same names (an environment-level value overrides the repo-level one), always, empty included, like the admin allowlist: clearing a variable turns the gate off on the next deploy.

2. **Scope.** With the gate on, the authorize request asks for `identify guilds.members.read`. Discord's consent screen then reads that the site "will read your member info (nickname, avatar, roles…) in servers you're in"; that line is the cost of the gate and the privacy notice (ADR 0149) says so. With the gate off the scope stays `identify`.

3. **The check.** `GET /users/@me/guilds/{guild}/member` with the user's access token. Admitted when it answers 200 and `roles` contains the role. A 404 (not a member) or a member without the role is refused. Any other answer (5xx, timeout, a 429 after one retry) is an **outage**, handled per item 5.

4. **At sign-in** (`discordCallback`, both the login-page and the invite flows):
   - **Admins are exempt.** A Discord id on `CMDCTRL_DISCORD_ADMIN_USER_IDS` is admitted without the call, so the owner can never lock themselves out.
   - **Admitted:** unchanged from today.
   - **Refused, login-page flow:** no session is issued and **no `users` row is created or updated**. A person the site has turned away leaves no data behind (ADR 0149). The callback redirects to `#/request-access`.
   - **Refused, invite flow:** the host's link vouches for this one table (owner answer 2). The person is seated as a **guest** named from their Discord display name, with no `UserID`, exactly as if they had typed that name. They get the table and nothing that needs an account.
   - **Outage:** refused, with `#/request-access?retry=1`, which says Discord did not answer and offers to try again. A sign-in is the one moment the person is present to retry, so it fails closed.

5. **At renewal** (`POST /me/session`, ADR 0110 §1 item 5: a signed-in session more than half spent). Before re-issuing, the server re-checks:
   - It exchanges the sealed refresh token (`identities.refresh_token`, ADR 0051 decision 5) for a new access token and stores the rotated refresh token.
   - **Admitted:** renewed as today.
   - **Refused:** not renewed, and the user's `sessions_invalid_before` is moved to now, so every session that person holds ends (ADR 0051 decision 6). Removing the role on Discord therefore locks a person out within half a session lifetime, about fifteen days at most for someone who is playing. An admin's `revoke-sessions` is still the immediate lever.
   - **No refresh token** (one never stored because `CMDCTRL_IDENTITY_KEY` was unset, or Discord rejects it): not renewed. The session runs to its expiry and the next sign-in is the check.
   - **Outage:** renewed anyway, and logged at WARN. A Discord blip must not sign every active player out, and the session's own expiry still bounds it.
   - Existing accounts are checked by exactly this path: nothing is grandfathered (owner answer 2).

6. **What the gate covers.** Only accounts: signing in, and therefore creating tables, My games, the deck library, synced settings and the catalog. Guests seated by a host's invite, spectator links, seat reclaim and the public pages are unchanged.

   **Amendment (2026-10-13, S70 [#2985](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2985)): the catalog is not gated.** Scryfall's guidelines forbid putting its data behind a chat-server membership, so `GET /catalog`, its image route and `#/catalog` are public ([ADR 0042](0042-card-catalog-page.md)'s amendment). The header shows Catalog signed out too.

7. **Admin view.** Accounts (ADR 0124) shows, per person, when the role was last confirmed and the result of the last check ("admitted", "refused: not a member", "refused: no role", "unchecked"). A new nullable `users.access_checked_at` and `users.access_result` hold it (one additive migration). A refused sign-in writes nothing, per item 4, so the view only ever shows people who once had an account.

#### The C2 server's roles

Set up by the owner on 2026-10-13. Discord holds them, not this repo; recorded here because the copy follows them.

| Role | Who has it | What it means here |
|---|---|---|
| **Planeswalker** | everyone who joins through the invite link | a member of the community; **no account**. The gate refuses them as "member without the role". They can still play a guest practice game (§2) and sit at a host's table by invite link. |
| **Elder Dragon Highlander** | given by a mod, on request | **may sign in**: create an account, host tables, keep decks. This is `CMDCTRL_DISCORD_PLAYER_ROLE_ID`. |

The request runs through Discord's onboarding. One question is "request account access"; answering yes adds the member to the access-request channel, where a mod grants Elder Dragon Highlander. Taking the role away ends that person's account access at their next renewal (item 5).

#### Amendment (S77, 2026-10-16, [#3325](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3325)): approve from Discord, and a mod log

Mods had to grant Elder Dragon Highlander by hand in Server Settings, and what happened in the app (reports, tables ended, people removed) was visible only in the admin view. This amendment adds two things and changes the bot's place in the role list.

**1. Approve and Decline buttons on access requests.** With `CMDCTRL_DISCORD_ACCESS_REQUEST_CHANNEL_ID` (the access-request channel, #account-requests on the C2 server) and `CMDCTRL_DISCORD_PLAYER_ROLE_ID` set in the gateway bot's environment:

- **The request.** The first message a member without the player role posts in that channel gets a reply from the bot: an *Access request* naming the member, with **Approve** and **Decline**. The bot reads who posted and their roles, never the text, so it needs only the unprivileged Guild Messages intent. Bots, webhooks, system messages, members who already hold the role and people on the bot's admin lists get nothing. At most one request per member per 24 hours, remembered in memory (a restart may post a second one). Nothing is stored.
- **Who may press.** A member who could grant the role by hand: Manage Roles or Administrator in that channel, as Discord computes it for the interaction, or someone on `CMDCTRL_DISCORD_ADMIN_USER_IDS` / `CMDCTRL_DISCORD_ADMIN_ROLE_IDS`. Anyone else gets an ephemeral refusal. The buttons give no one a power they did not already have.
- **Approve** grants the role (`PUT /guilds/{guild}/members/{user}/roles/{role}`, with an audit-log reason naming the mod) and edits the request to say who approved it and that the member can sign in. **Decline** changes nothing on Discord and edits the request to say who declined it. Either way the buttons are removed. The buttons are stateless (`c2-access:approve:<user id>`), so a request posted before a restart still works.
- **The gate is unchanged.** Approve only grants the role that item 3 already checks; a member is let in at their next sign-in exactly as if a mod had given the role by hand.

**2. The mod log.** `CMDCTRL_DISCORD_MOD_LOG_CHANNEL_ID` names a private channel that only moderators can read. The server posts to it with the bot token (the same REST path as ADR 0174's results and ADR 0196's bug forum), and the gateway bot adds a line when a request is approved or declined:

| Entry | Says |
|---|---|
| Player report | the table's name, the reported seat's (or spectator's) name, the reason |
| Table ended / restored / deleted | the table's name, and which admin |
| Seat removed / Spectator removed | the table, the name the seat had there, whether it was blocked, and which admin. Only an admin's removals: a host managing their own table is not a mod action. |
| Sessions revoked | the account's display name, how many open connections closed, and which admin |
| Access request approved / declined | the member and the mod, as Discord mentions (an embed never notifies) |

Each server entry links into the admin view (Reports, the table, the account) when `CMDCTRL_PUBLIC_BASE_URL` is set. "Which admin" is an allowlisted admin's display name, or "the admin token" for the shared credential, which is also what the bot's `/c2-end` uses.

**Privacy (ADR 0149).** An entry carries only what an admin already reads in the admin view, and less: **never the reporter's name, the chat excerpt, an account id or a Discord id** in the server's entries. Nothing new is stored by the server, so `accountdelete.Inventory` and the privacy schema version do not change. What is new is that a copy reaches Discord and stays until a moderator deletes it, unlike the report itself, which is removed after 90 days. The privacy page says so in the Player reports entry and in what Discord receives.

Posting never blocks or fails the request that caused it: one goroutine per entry, at most eight at once (past that an entry is dropped with a WARN), up to three attempts while Discord rate-limits or does not answer, then a WARN. Unset, or no bot token: nothing is posted. A malformed id stops the boot of the server and of the bot.

**3. The bot's role (reverses earlier advice).** ADR 0004 and `deploy/README.md` authorized the bot with no permissions, and ADR 0159 §7 asked for its role to be kept below the moderators'. To grant a role, Discord requires the bot to hold **Manage Roles** and its own highest role to sit **above the role it grants**. So, on the C2 server:

- the bot's role (`cmd_and_ctrl`) holds Manage Roles (voice rooms already gave it this server-wide), and
- it is dragged **above Elder Dragon Highlander** in Server Settings → Roles, while staying below the moderators' roles. Discord never lets it grant or remove a role at or above its own, so it can still not touch a moderator.
- In #account-requests it needs View Channel, Send Messages, Embed Links and Read Message History; in #mod-log, View Channel, Send Messages and Embed Links. #mod-log denies View Channel to @everyone.

These are owner actions in Discord; no code or deploy does them. Until they are done, Approve answers the mod with an ephemeral message saying exactly this.

**Calls made here** (the owner can overturn any in review):

- **The trigger is the member's first message in the channel**, not a new onboarding role or a "Request access" button. It matches today's copy ("say hi in #account-requests"), needs no privileged intent and no extra role, and asks nothing new of the member.
- **"A mod" is Discord's own Manage Roles,** plus the bot's admin lists, rather than a new variable.
- **The reporter is not named in #mod-log.** Discord keeps the message for as long as a moderator leaves it there, and the person who reported is the one most at risk if it leaks; the admin Reports view has the name.
- **Ban sync is not built.** Revoking sessions when a member is banned or loses the role, at once rather than at renewal, needs a gateway listener this ADR rejected under Consequences. It waits for an owner decision on #3325. (Built by the 2026-10-29 amendment below.)

#### Amendment (S77, 2026-10-29, [#3325](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3325)): ban sync

**Owner decision (2026-10-29): option 1.** A Discord ban, the player role taken away, or a member leaving or being kicked ends that person's CMD & CTRL sessions at once, through the gateway. This **reverses the rejection of a gateway listener** under Consequences ("A bot listening for role removals was considered and rejected"). Revocation at renewal (item 5) stays as the backstop.

- **Events.** The gateway bot listens, in the community server (`CMDCTRL_DISCORD_COMMUNITY_GUILD_ID`) only, for `GUILD_BAN_ADD`, for `GUILD_MEMBER_UPDATE` where the member no longer holds the player role, and for `GUILD_MEMBER_REMOVE`. An update from a member the bot's cache shows did not hold the role before it either (a nickname change) is ignored. A remove is held 5 seconds so that a ban's own event, which Discord may send after it, is acted on first.
- **The bot asks; the server decides.** The bot is a separate process, so it calls a new admin route, `POST /admin/discord-members/{discord_id}/revoke-sessions` with `{"reason": "banned" | "role_removed" | "left"}`, with its admin session, as its commands already call back in. The server looks the account up by Discord id and revokes through the same code as `POST /admin/users/{id}/revoke-sessions` (the watermark, then the open sockets). It leaves alone allowlisted admins and `CMDCTRL_DISCORD_GATE_EXEMPT_USER_IDS`, exactly as the gate does; Discord users with no account (quietly: no log line, no mod log); and everything while the gate is off.
- **Idempotent.** The refusal is recorded as the account's gate answer (`not_member` for a ban or a leave, `no_role` for the role), so the admin view shows it. An account whose last answer is already a refusal is left alone, and a sign-in records `admitted`, so a ban and the member-remove that comes with it revoke and log once, and a person let back in is caught again next time. The route is serialized, so two events for one person cannot both pass that check.
- **Mod log.** A revocation posts "Sessions revoked by ban sync": the account's display name, why (banned, the role taken away, left or kicked) and the connections closed, with the admin view link. No Discord id or account id, as for the other entries.
- **Off by default: `CMDCTRL_DISCORD_BAN_SYNC=1`** in the bot's environment, with the guild and `CMDCTRL_DISCORD_PLAYER_ROLE_ID`. Role changes and leaves need the **privileged Server Members intent**; asking for it before it is switched on in the Developer Portal makes Discord close the gateway with 4014, and every command would go with it. So the bot asks for it only with the flag on, and if Discord refuses it anyway, the bot logs an ERROR naming the portal switch and connects again without it: bans keep syncing (the unprivileged Guild Moderation intent) and everything else works. Bans alone are behind the same flag, so turning the feature on is one decision. CD writes the flag and the guild to `bot.env` from Actions variables, always, empty included.
- **Owner setup.** Switch on Server Members Intent in the Developer Portal; give the bot's role **View Audit Log**, without which Discord does not send it `GUILD_BAN_ADD` (a ban then still ends sessions through its member-remove, logged as "left"); then set `CMDCTRL_DISCORD_BAN_SYNC=1` and deploy `main`.

**Privacy (ADR 0149).** Nothing new is stored: the event is acted on and forgotten, and the only write is the gate answer the renewal check already records. The bot now reads membership changes in the C2 server, which the privacy page says.

**Calls made here** (the owner can overturn any in review):

- **The server, not the bot, holds the rules.** The exemption list and the user database live in the server, so the bot sends every qualifying event and the server answers `revoked`, `no_account`, `exempt`, `gate_off` or `already_out`.
- **A leave revokes too.** A member who leaves loses the gate's "member" condition, and a kick arrives only as a remove.
- **The fallback happens at connect time.** If the intent is switched off while the bot runs, discordgo's own reconnect keeps failing until the bot restarts, which then falls back. Turn the flag off, or restart the bot, after switching the intent off.

### 2. A guest practice game

Owner answer 4. Today the practice table needs a session, so a stranger must sign in with Discord first. Under the gate that would shut out everyone not yet approved, and it already falls short of S70's exit criterion 3.

- **`POST /games/practice/guest`**, public, no session. It mints a guest session (name "Player", no `UserID`, the `CMDCTRL_SESSION_TTL` lifetime) and opens a practice table for it in one step, with the same decks and bot as `POST /games/practice`.
- **Limits:** one live practice table per guest session; a per-IP limit of 3 per 10 minutes; ADR 0148's table and connection caps apply and a hit cap shows the friendly message. Practice tables are unlisted and never persisted (ADR 0076), so nothing is kept.
- The front page, the gate page and the Help menu offer **Try a practice game**. A guest still cannot create a real table.

### 3. One site shell

Owner answer 5.

1. **A page frame.** A new `Page.svelte` gives every non-game route the same container and head: an optional eyebrow, a title, a lede, and an actions slot on the right. It also sets a width (`wide` 1200 px for tables and grids, `prose` 72ch for reading). Titles are sentence case, in the display face, at one size. One set of breakpoints replaces the five: tablet (≥ 820 px) is the design target (the tablet-minimum rule); below it the frame and header wrap but get no separate layout.
2. **Shared primitives** go into `app.css`: `.card`, `.btn` with `primary`, `ghost` and `discord` variants, and `.link-arrow`. A page keeps one-off CSS only for what is truly its own.
3. **The header, grouped.**
   - Signed in: the wordmark, then **Tables** and **My games** | **Decks**, **Catalog** and **Roadmap**. On the right: the Discord icon, Help, and the account menu. **Admin** shows in admin mode, as its own accent-tinted group after Roadmap in the header nav and as an item in the account menu (owner, 2026-10-13).
   - Signed out: the wordmark, then **Decks**, **Roadmap** and **About**. On the right: the Discord icon, Help, and **Sign in**.
   - "Home" leaves the header. The route stays `#/lobby`, and only the link's text becomes "Tables", through `labels.ts` (ADR 0125 §2).
4. **The footer** is one row: Discord, About, Privacy, GitHub, then the Fan Content line. The legal disclosure (ADR 0149) is unchanged.
5. **Tips on site pages render inside the frame**, as a dismissible line under the lede, rather than as a popover over the controls. The one-per-visit rule, "Hide tips" and the hint registry are unchanged. In-game hints are unchanged.

### 4. The front page

Owner answer 5, mockup A.

- **`#/`** for a signed-out visitor becomes the front page. `#/login` stays as an alias, since links, the e2e suite and the OAuth error path use it, and `#/home` redirects to `#/`. `#/admin`'s token form is unchanged.
- **The hero:** an eyebrow, a one-line pitch and a short lede, then two buttons side by side: **Sign in with Discord** and **Join our Discord server**. Beside them is a three-step panel: *Join the C2 Discord*, *Request account access* (answer yes in onboarding and a mod gives the Elder Dragon Highlander role), *Sign in and sit down*. With the gate off, the panel shows only step 3 and a line inviting people to the server.
- **Below:** a row of four cards. **Have an invite?** has the code box, today's join behaviour unchanged. **Check a deck**, **Roadmap** and **About** link out. **Try a practice game** (§2) sits under the hero buttons.
- **A signed-in person** still lands on Tables (ADR 0112 §1).

### 5. The gate page

`#/request-access`, public.

- It explains that accounts are for members of the C2 Discord with the Elder Dragon Highlander role, in steps that follow the server's onboarding:
  1. **Join**, with the invite button. Everyone who joins is a Planeswalker.
  2. **Answer yes to "request account access"** in the server's onboarding. A member who already joined and answered no changes the answer in the server's *Channels & Roles*.
  3. **Ask in the access-request channel** that the answer opens. A mod gives the Elder Dragon Highlander role.
  4. **Check again**, which restarts Discord sign-in. Discord remembers the consent, so this is a redirect, not a prompt.

  The role and channel names are copy in the client, not config: renaming them on Discord means a copy change here.
- It offers **Try a practice game**, and says that an invite link from a host still works.
- With `?retry=1` it leads with "Discord didn't answer; try again".

### 6. The Discord placements

Owner answer 6. All of them hide when `CMDCTRL_DISCORD_INVITE_URL` is unset.

- **The front page** (§4) and **the gate page** (§5).
- **The header:** an icon link on every non-game page.
- **The footer:** the first link.
- **The lobby's empty state:** its three cards become Practice game, Bring a deck, and **No pod yet? Find a pod**, which links the invite.

The client reads the invite and the gate state from **`GET /community`**: public, `Cache-Control: public, max-age=300`, returning `{"invite_url": "…", "gate": true}`. It returns no ids.

### 7. Page passes

Each page moves into the frame. Its content is unchanged except:

- **Roadmap:** each status section is collapsed to its heading and counts until opened. "Up next" stays open.
- **About:** section heads use the shared style, and the Discord link moves from the Feedback list to a line of its own.
- **Admin, Catalog, Decks, My games, Join, Reclaim and Privacy** take the frame's head. Their bodies are unchanged.

## Consequences

- **Consent shows more.** With the gate on, Discord's consent screen shows the member-info line. The privacy notice must name the scope, the roles read and the two new columns, in the same PR as the gate.
- **Getting in means asking.** A new player has to join the server and be given the role by a mod. The soft launch's "one Commander Discord" becomes this one, and the S70 post points people at the server. The guest practice game is how a stranger tries the site before asking.
- **Revocation is bounded but not instant.** Taking the role away ends access at the person's next renewal, at most half a session lifetime away, or at once through `revoke-sessions`. A bot listening for role removals was considered and rejected: it runs on prod only, and a second path to revocation is a second thing to test.
- **Label renames break literal selectors.** "Lobby" becoming "Tables" in the header goes through `labels.ts`. The older e2e specs that select on the literal are updated in the same PR, and the nightly is run on the branch.
- **`#/home` is gone.** This ends ADR 0092's two-portal experiment: the front page serves visitors and the header serves everyone else.
- **Revocation is instant with ban sync on (2026-10-29 amendment).** The owner reversed the rejection of a gateway listener above: with `CMDCTRL_DISCORD_BAN_SYNC=1`, a ban, the role taken away, or a leave in the C2 server ends the person's sessions at once. Renewal stays the backstop when the flag is off, when the Server Members intent is refused, or when the bot is down.

## Slices

| # | Slice | Model | Needs |
|---|---|---|---|
| 1 | This ADR | — | — |
| 2 | **Server: the gate** (§1): config and boot checks; scope; member check; callback branches; renewal re-check with the refresh-token exchange; migration and admin fields; `GET /community`; CD "Sync server env (community)"; privacy notice; AGENTS.md env vars; docs/lobby.md | Opus | 1 |
| 3 | **Server: guest practice** (§2): the route, limits, tests; docs/lobby.md | Sonnet | 1 |
| 4 | **Client: the shell** (§3): `Page.svelte`, primitives, header groups, footer, tips in the frame, `GET /community` store; every route moved into the frame with no body changes; e2e selector updates | Opus | 2's `GET /community` (or a stub) |
| 5 | **Client: front door** (§4–6): the front page, `#/request-access`, `#/home` redirect, Try a practice game, lobby empty state, the Discord placements | Opus | 3, 4 |
| 6 | **Client: page passes** (§7) | Sonnet | 4 |

Every client slice attaches before and after screenshots at 1440 px and 820 px to its PR, and runs the nightly e2e on its branch (the label contract).
