# ADR 0163 — Discord Linked Roles: CMD & CTRL stats on a member's profile

**Status:** Proposed · 2026-10-13 · Backlog
**Issue:** [#3323](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3323)
**Owner decisions:** the two answers of 2026-10-13 are quoted under [Owner decisions](#owner-decisions-2026-10-13). They are binding. Calls this ADR makes on its own are listed under [Calls made here](#calls-made-here) for the owner to overturn in review, before slice 2 lands.
**Numbering:** checked with ADR 0161's sweep (12 remote heads, highest 0160). ADRs 0161 and 0162 are drafted in the same PR; this is **0163**.
**Builds on:** [ADR 0004](0004-discord-identity.md) (Discord sign-in), [ADR 0051](0051-user-database.md) (users, identities, the sealed refresh token), [ADR 0057](0057-win-and-lose-by-effect.md) decision 7 (`games.outcome`), [ADR 0149](0149-privacy-notice-and-account-deletion.md) (the notice and the deletion inventory), [ADR 0156](0156-a-front-door-site-shell-front-page-and-a-discord-gate.md) (the gate and its use of the refresh token).

---

## Context

Discord's **Linked Roles** let an application publish a few verified numbers about a person, shown on their Discord profile as a connection, and let a server's owner make roles that require them ("Played 10+ games"). The person links the application from Discord: a server role's **Links** requirement, or **User Settings → Connections**, sends them to the application's *Linked Roles Verification URL*, the application runs OAuth with the `role_connections.write` scope, and then writes the person's numbers with their own access token.

For CMD & CTRL this puts the app on its players' profiles in the server where new players arrive, and gives the owner roles that reward playing, with no moderator work.

What already exists:

- **Discord OAuth** (ADR 0004): `/auth/discord/start` and the callback, with a `StateEntry` that already branches by purpose: claim a seat, link Discord to a guest seat, or the plain sign-in.
- **A sealed refresh token per identity** (`identities.refresh_token`, sealed under `CMDCTRL_IDENTITY_KEY`). The sign-up gate (ADR 0156) uses it to re-check the role at renewal.
- **The numbers:** `games.outcome` and `winner_seat` (ADR 0057 decision 7), `seats.user_id`, `users.created_at`.

## Decision

### 1. Stats only

Owner answer 1. Linked Roles shows stats and lets the owner build roles from them. **It grants nothing in CMD & CTRL.** Elder Dragon Highlander stays a moderator's grant and the sign-up gate (ADR 0156) is unchanged. Linking needs an account; it never creates one.

### 2. The three fields

Owner answer 2. Registered once as the application's role-connection metadata (`PUT /applications/{app}/role-connections/metadata`, bot token), at boot when the feature is on. It is idempotent, so every boot re-sends the same list.

| key | name | type | value |
|---|---|---|---|
| `games_played` | Games played | integer ≥ | finished games the account sat in to the end or was eliminated from |
| `games_won` | Games won | integer ≥ | of those, games whose `outcome` is `win` and whose `winner_seat` is theirs |
| `joined` | Days since joining | datetime ≥ (days) | `users.created_at`, as ISO 8601; Discord turns it into "joined at least N days ago" |

**Counted:** non-practice games with `state = ended` and `outcome` of `win` or `draw` where the account held a seat. Abandoned tables (`outcome` null) and practice games do not count. Bot and agent seats have no account and count for nobody.

The owner sets roles in Discord (Server Settings → Roles → a role → **Links** → Add requirement → CMD & CTRL → for example *Games played ≥ 10*). No code knows about the roles.

### 3. Linking

- **Verification URL:** `https://cmd.labxp.io/auth/discord/linked-role`, set once in the Developer Portal (General Information → Linked Roles Verification URL).
- That route starts the existing OAuth flow with a fourth `StateEntry` purpose, **role connection**, and the scopes `identify role_connections.write`. It comes back to the **existing callback and redirect URI**, so the portal needs no new redirect.
- **The callback, for this purpose:** exchange the code, read `/users/@me`, and find the identity by Discord id.
  - **No account:** a page that says "Sign in to CMD & CTRL first, then link again", with the sign-in button. Nothing is stored.
  - **An account:** store this grant's refresh token (§4), push the numbers (§5), and show "Linked. You can go back to Discord."
- It mints **no session**: linking is not signing in.

### 4. A separate grant

A refresh token carries the scopes of the authorization that issued it. The sign-in grant in `identities.refresh_token` has `identify` (and `guilds.members.read` with the gate on), and **every sign-in replaces it**, so it cannot hold `role_connections.write`.

- **Migration:** `identities.role_connection_token BLOB`, sealed with the same `Sealer` under `CMDCTRL_IDENTITY_KEY`. No key configured: the feature is off (said once at boot), because the token could not be kept.
- Each push exchanges it for a fresh access token and stores the rotated refresh token.
- **`invalid_grant`** (the person removed the connection in Discord, or deauthorized the app): clear the column. They are unlinked until they link again.

### 5. When the numbers are pushed

`PUT /users/@me/applications/{app}/role-connection` with the person's access token: `platform_name` "CMD & CTRL", `platform_username` their display name, and the three values.

- **On link.**
- **After a game ends,** for every linked account that sat in it, from a small queue so the end of a game never waits on Discord. The values are recomputed from the database, never incremented, so a missed push is corrected by the next one.
- **No sweep.** `joined` is a date, so Discord's "days since" moves on its own; the counts only change when a game ends.
- **Failures:** a 429 waits `retry_after` (capped) once; a 5xx or timeout is logged at WARN and left for the next game.

### 6. Configuration

On when all of these are set: `CMDCTRL_DISCORD_APP_ID`, `CMDCTRL_DISCORD_BOT_TOKEN` (to register the metadata), the OAuth client secret (the existing sign-in config) and `CMDCTRL_IDENTITY_KEY`. There is no new variable: the owner turns it on by setting the Verification URL in the portal. Without the bot token (every environment but production) the route answers a page saying linking is unavailable. Production only, like voice rooms.

### 7. Privacy and deletion

- **The notice** gains a paragraph: if you link CMD & CTRL on Discord, your games played, games won and the date you joined are shown on your Discord profile and can be used by the server's roles. Unlink at any time in Discord (User Settings → Connections).
- **Account deletion** (ADR 0149): before the identity row goes, push empty metadata so the profile stops showing the numbers (best effort, as a failure must not block a deletion), then the column goes with the row. The inventory lists `identities.role_connection_token`.

## Consequences

- **CMD & CTRL shows up on players' profiles** in the server where new players arrive, and the owner can make roles that reward playing.
- **Two grants per linked person:** sign-in and role connection. Each is used for one thing, and deleting either does not break the other.
- **The numbers can be a game behind** after a failed push, until the next game.
- **No self-serve access.** Owner answer 1 keeps Elder Dragon Highlander a moderator's grant. Linking needs an account, which needs the role.

### Owner setup (once, not code)

1. Developer Portal → the application → **General Information** → **Linked Roles Verification URL** = `https://cmd.labxp.io/auth/discord/linked-role`.
2. After the deploy that registers the metadata, add requirements to the roles you want in Server Settings → Roles → *role* → **Links**.

## Owner decisions, 2026-10-13

1. **Relation to account access:** stats only; Elder Dragon Highlander stays a moderator's grant. Recommended option, chosen.
2. **Fields:** games played, games won, days since joining. "Has an account" was offered and not chosen.

## Calls made here

- **The existing callback and redirect URI,** with a fourth state purpose, instead of a second redirect URI to register.
- **A separate sealed column** for the role-connection grant (§4).
- **Abandoned tables do not count** as played.
- **Push after each game,** with no periodic sweep.

## Slices

| # | Slice | Model | Needs |
|---|---|---|---|
| 1 | This ADR | — | — |
| 2 | **Server:** migration and sealed column, metadata registration at boot, the `linked-role` route and the callback branch, the stats query, the push queue, `invalid_grant` handling, account-deletion step and inventory; httptest coverage of every Discord call; docs/lobby.md | Opus | 1 |
| 3 | **Client:** the linked and not-signed-in pages, and a "Show my stats on Discord" line in account settings with a link to the verification URL; privacy paragraph | Sonnet | 2 |
