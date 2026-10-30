# ADR 0214 — CMD & CTRL as a Discord Activity: the table inside the voice channel

**Status:** Proposed · 2026-10-30 · S77 — Discord and social: voice, friends, mod tools and events
**Issue:** [#3312](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3312)
**Owner decisions:** none yet. The owner asked for this ADR on 2026-10-13 and wants product questions put to them, not settled here. They are listed under [Open questions for the owner](#open-questions-for-the-owner), each with a recommendation. The calls this ADR makes on its own are under [Calls made here](#calls-made-here). No production code until the owner accepts it.
**Numbering:** every remote head's `docs/decisions/` was listed on 2026-10-30 after `git fetch --all --prune` (the heads are named in the PR). `develop` ends at 0211; 0212 is taken on `origin/docs/lobby-character-select` (#4656) and 0213 on `origin/docs/3313-friends-presence` (#4662), so this is **0214**.
**Builds on:**
- [ADR 0004](0004-discord-identity.md) (Discord identity and the bot) and [ADR 0051](0051-user-database.md) decisions 2, 3, 5 and 6 (users, sessions, the stored refresh token, revocation).
- [ADR 0110](0110-remember-me.md) §1 (session lifetimes and renewal) and [ADR 0156](0156-a-front-door-site-shell-front-page-and-a-discord-gate.md) §1 with its amendments (the sign-up gate, the invite-flow guest, ban sync).
- [ADR 0159](0159-per-table-discord-voice-rooms.md) (per-table voice rooms).
- [ADR 0042](0042-card-catalog-page.md)'s amendment and [ADR 0092](0092-public-roadmap-and-site-portal.md) decision 1 (where card art may be served, and to whom).
- [ADR 0148](0148-capacity-limits-for-a-public-launch.md) §6 (per-IP buckets), [ADR 0149](0149-privacy-notice-and-account-deletion.md) (privacy), [ADR 0075](0075-table-settings-and-host-controls.md) §2.1 and §2.6 (the host, moderation), [ADR 0164](0164-rematch-the-host-offers-the-same-game-again.md) (rematch), [ADR 0177](0177-away-players-and-moving-the-table-on.md) (away seats), [ADR 0211](0211-closing-idle-tables.md) (idle tables), [ADR 0031](0031-progressive-web-app.md) (the service worker).
**Supersedes:** #612 (the Rich Presence spike), closed by the owner on 2026-10-15 in favour of this ADR.

---

## Context

Commander is a talking game, and the C2 Discord is where the players are. ADR 0159 gave a table its own voice channel; the owner then asked (2026-10-13) for the tightest version of that: the game itself running **inside** the voice channel as a Discord Activity, with voice, video and screen sharing native around it. The website stays exactly as it is. The Activity is a second way in.

This ADR settles sign-in, networking, joining, layout, Rich Presence, and cost and risk, as the issue asks, and plans the slices.

### What an Activity is

Everything here is from Discord's documentation, read on 2026-10-30 (the source of the docs site at commit `e3ea589033` of `discord/discord-api-docs`) and from Discord's developer help centre. The full list is under [Sources](#sources). What could not be confirmed is listed under [What the docs do not settle](#what-the-docs-do-not-settle).

- **A web app in an iframe** inside Discord on desktop, web and mobile, talking to the Discord client over `postMessage` through the Embedded App SDK (`@discord/embedded-app-sdk`, MIT, v2.5.0). The SDK is meant for a single-page app ([How Activities work](https://docs.discord.com/developers/activities/how-activities-work)).
- **Served through Discord's proxy.** The iframe loads `https://<application_id>.discordsays.com/`, and every request goes through Discord's proxy, which "hides the users' IP addresses, your application's IP addresses" and blocks known-bad URLs. The proxy runs on **Cloudflare Workers**. It carries WebSockets; WebRTC and WebTransport are not supported ([Networking](https://docs.discord.com/developers/activities/development-guides/networking)).
- **URL mappings** in the Developer Portal say where requests go: the `/` mapping is the app itself, and further prefixes map to other hosts. A target is a host and path without a protocol, and must be a directory ([Local development, URL mapping](https://docs.discord.com/developers/activities/development-guides/local-development#url-mapping)).
- **A content security policy** sandboxes the iframe: a request to anything that is not the app's own proxy fails with `blocked:csp`, apart from a short list of Discord API and CDN paths. External links must go through the SDK's `openExternalLink`, which asks the user first ([User actions](https://docs.discord.com/developers/activities/development-guides/user-actions)).
- **Cookies** work through the proxy only with the domain matching `<application_id>.discordsays.com` and `SameSite=None; Partitioned` ([Networking, using cookies](https://docs.discord.com/developers/activities/development-guides/networking#using-cookies)).
- **Sign-in** is the SDK's `authorize` command, which returns an OAuth2 code (PKCE `S256` is supported), exchanged for an access token **by the app's server** with the client secret, and then `authenticate` with that token. Scopes outside those granted make SDK commands fail; adding a scope may show an OAuth modal ([Embedded App SDK](https://docs.discord.com/developers/developer-tools/embedded-app-sdk), [Building an Activity](https://docs.discord.com/developers/activities/building-an-activity)).
- **Instances.** Everyone who joins an Activity in one place gets the same `instanceId`; when the last participant leaves, that instance is over and is never reused. The SDK lists and pushes the instance's **connected participants**. Client data is not to be trusted: the server confirms an instance with `GET /applications/{app}/activity-instances/{instance}` and the **app's bot token**, which returns the instance's location (`guild_id`, `channel_id`) and its `users`. Optional signed proxy headers (`X-Signature-Ed25519`, `X-Signature-Timestamp`, `X-Discord-Proxy-Payload`) prove a request came through the proxy ([Multiplayer experience](https://docs.discord.com/developers/activities/development-guides/multiplayer-experience)).
- **Rich Presence** comes free: a user in an Activity shows the app on their profile, with "Ask to Join" for people who can join. `setActivity` with the `rpc.activities.write` scope sets the details, state, party size and timestamps ([Rich Presence with the Embedded App SDK](https://docs.discord.com/developers/rich-presence/using-with-the-embedded-app-sdk)).
- **Layout.** Three modes: focused, picture-in-picture and grid. Orientation can be locked, with separate portal defaults for phones and tablets. Mobile safe areas come as `--discord-safe-area-inset-*` CSS variables, and thermal state is published on iOS and Android ([Layout](https://docs.discord.com/developers/activities/development-guides/layout), [Mobile](https://docs.discord.com/developers/activities/development-guides/mobile)). The SDK's `platform` is only `desktop` or `mobile`, so it cannot tell a tablet from a phone (SDK source, `src/Discord.ts`).
- **Platforms** are switched on per platform in the portal: web and desktop by default, iOS and Android each a checkbox ([Mobile](https://docs.discord.com/developers/activities/development-guides/mobile)).
- **Launching.** Enabling Activities creates a global **Entry Point** command ("Launch") that opens the Activity from the App Launcher. A bot can also launch it by answering an interaction with `LAUNCH_ACTIVITY` (type 12), which only works for the same application ([How Activities work](https://docs.discord.com/developers/activities/how-activities-work)).
- **Who can launch it.** An **unverified** Activity is "only visible to you, your development team, and app testers" and "limited to servers with fewer than 25 members" (and DMs). A team holds up to 100 members and an app can have up to 50 testers, all of whom must be the owner's Discord friends. A **verified** Activity is "visible to everyone on Discord" and "playable in any server, regardless of size". Verification needs the team owner's identity checked through Stripe, and **once an app is verified, its Public Bot setting cannot be turned off**. Being listed in the App Directory and App Launcher (**Discovery**) is a separate opt-in after verification ([What are Verified and Unverified Activities?](https://support-dev.discord.com/hc/en-us/articles/26576097154199), [How Can Users Discover and Play My Activity?](https://support-dev.discord.com/hc/en-us/articles/21204493235991), [How Do I Get My App Verified?](https://support-dev.discord.com/hc/en-us/articles/23926564536471), [Enabling Discovery](https://docs.discord.com/developers/discovery/enabling-discovery)).
- **Privileged intents are not affected by verification.** The gateway page still says a verified app needs its privileged intents approved, but the help centre's notice of 2026-10-26 says App Verification and intent review "are now separate", and an app reaching fewer than 10,000 users keeps the self-serve toggles ([Changes to Privileged Intent Access](https://support-dev.discord.com/hc/en-us/articles/40281523410967)). The newer notice is taken as current, so verifying the C2 app does not put ban sync's Server Members intent (ADR 0156, 2026-10-29) or #3313's Presence intent at risk.
- **Testing.** The SDK ships a `DiscordSDKMock` for tests (SDK source, `src/mock.ts`). Logs from the iframe reach Discord's own console and debug logs.

### What exists on our side

Each item was checked in the code at `origin/develop` (`2edf3b5dd`).

- **One Discord application** serves sign-in and the bot: "The app ID equals the sign-in client ID" ([docs/environments.md](../environments.md#secrets)). The client ID and secret are the same secrets in the `prod` and `dev` environments. The bot and the server's bot token run on **production only**.
- **Sign-in is a top-level redirect.** `discord.Config.AuthorizeURL` and `ExchangeCode` (`server/internal/discord/oauth.go`) drive a redirect to `discord.com/oauth2/authorize` and back to `/auth/discord/callback`. The exchange always sends `redirect_uri`. With the gate on, the scope is `identify guilds.members.read`.
- **The gate** (`server/internal/lobby/access_gate.go`) checks the member with the person's own access token (`GET /users/@me/guilds/{guild}/member`). A refused person on an invite link is seated as a guest named from their Discord display name, with "no user and no Discord identity" (`seatRefusedAsGuest`). Renewal exchanges the sealed refresh token and asks again (`renewalGate`).
- **Sessions** are HMAC-signed principals (`server/internal/auth`). They travel in the `cmdctrl_session` cookie (`HttpOnly`, `Secure`, `SameSite=Lax`, `lobby/http.go` `setSessionCookie`), as `Authorization: Bearer`, or as `?token=` (`auth.CredentialFromRequest`). The client keeps the token in `localStorage["cmdctrl.session"]` and sends the bearer header on every API call (`client/src/lib/session.ts` `authFetch`). An `identified` session cannot open a WebSocket (`lobby/ws_authorizer.go`).
- **The client uses relative URLs throughout.** API calls are `/games…`, and the table dials `location.host + "/ws"` with `?token=` (`client/src/routes/Game.svelte`, `client/src/lib/gameURL.ts`). Behind a `/` mapping, all of it would resolve to `<app_id>.discordsays.com` and through the proxy, unchanged.
- **Card images are cookie-borne.** `/cards/{id}/image` sits behind `auth.Middleware` (`server/cmd/server/main.go`), and `cardImageURL` builds `/cards/<id>/image?size=…` with no token (`client/src/lib/cardImage.ts`), so an `<img>` relies on the session cookie. The response is `Cache-Control: public, max-age=604800, immutable` (`server/internal/cards/http.go`). Avatars carry `?token=` (`api.ts` `avatarURL`). The general image route is gated on purpose: unauthenticated it would be "a general-purpose image proxy for ~35k cards" (ADR 0042).
- **The shell forbids framing.** Caddy's app-shell headers set `frame-ancestors 'none'` and `X-Frame-Options: DENY`, with `connect-src 'self'` (`deploy/Caddyfile`). Caddy serves plain HTTP behind cloudflared and the Cloudflare edge.
- **The WebSocket checks Origin.** `Hub.checkOrigin` allows the same host or a host listed in `CMDCTRL_ALLOWED_ORIGINS` (`server/internal/ws/hub.go`).
- **Rate limits are keyed by client IP** from `CF-Connecting-IP` (`server/internal/util/clientip`): table creation (ADR 0148 §6), guest practice (3 per 10 minutes per IP, ADR 0156 §2), deck checks.
- **Bot Fight Mode** is on for `labxp.io` and challenges datacenter traffic: it 403s GitHub's runners, and no WAF rule can skip it on the free plan ([docs/environments.md](../environments.md#uptime-watcher)). The owner keeps it on (2026-10-05).
- **External links** open with `target="_blank"` or `window.open` in 22 non-test client files: Scryfall, Moxfield and Archidekt, shop links, the Discord invite, GitHub, the legal pages.
- **Voice rooms** (ADR 0159) are built: `lobby/voice_manager.go`, `games.voice_channel_id` (migration `0020_voice_rooms.sql`).
- **A table seats four** (`game.MaxPlayers = 4`).
- **The service worker** is registered in `client/src/lib/pwa.ts`.
- **Frames** are about 70 KB of JSON uncompressed (ADR 0205), and the hub pings every 54 s (`pingPeriod`).

## Decision

### 1. Build it, in phases, behind a switch that fails closed

The Activity is the same client and the same server, reached through Discord's proxy. It is not a second app: one bundle, one API, one table. Activity mode is a runtime switch in the client: the page was loaded with the SDK's `frame_id` and `instance_id` query parameters, on the Activity host (§3). The SDK is loaded only in that mode, as its own chunk, so the website's bundle does not grow.

The server side is **on only when fully configured** (§9), as ADR 0159 §7 does. Until then, every Activity route answers 503 naming what is missing, and the website is unchanged. The develop preview is configured first, with its own Discord app (§2). Production is configured only after Phase 0's findings are in, Phase 2 is built and the owner has verified the app (§8.2). Phase 0's throwaway page and the E2E stand-ins are dev-only routes behind `requireDev`, as AGENTS.md requires.

### 2. Which Discord application

**Recommended: A1, the existing application on production, and a second, unverified application for the develop preview.** This is open question 1.

| | **A1. The existing app** (recommended) | **A2. A separate Activity app** |
|---|---|---|
| Consent | one: a player who signed in on the website has already granted `identify` and `guilds.members.read`; the Activity adds `rpc.activities.write` | a second consent screen, under the second app's name |
| Tokens | the same client, so the refresh token from an Activity sign-in can be stored, and ADR 0156's renewal check keeps working | a different client: its refresh token cannot be used for the website's renewal, so it is not stored |
| Launching from Discord | the existing bot can answer an interaction with `LAUNCH_ACTIVITY`, for example a **Play here** button in a voice room's chat (Phase 4) | the bot is a different app and cannot. Launching is through the App Launcher only, unless the Activity app gets its own interactions endpoint, which ADR 0004 §1 avoided |
| Instance check | the bot token the server already holds on production | a second bot token, for REST calls only |
| Verification | the C2 app becomes verified, and **its bot becomes public for good**: anyone with Manage Server can add the C2 bot to their own server. The guild allow-list (ADR 0004 §3) keeps its commands out of those servers, but the bot would still sit in them, so under A1 **the bot leaves any server not on its allow-list**, which amends ADR 0004 §3's "we do NOT auto-leave" | only the Activity app is verified and public. The C2 bot is unchanged |
| Privileged intents | the self-serve toggles hold while the app reaches fewer than 10,000 users (see Context). Strangers' servers count toward that, which is the other reason for leaving them; past it, ban sync's Server Members intent would need Discord's review within 90 days | unaffected: the Activity app asks for no intents |
| In the C2 server | one app in the server's integrations | two |

Either way the **develop preview needs its own application.** An app has one set of URL mappings, so its `/` can point at one host only. The dev app stays unverified: its testers are the owner's team and app testers, in a test server of fewer than 25 members. The dev box holds the dev app's client secret and bot token. It uses that bot token for REST calls only and never opens a gateway, so AGENTS.md's rule that the bot never runs from `develop` (two processes answering one application's commands) is not touched.

The server reads the Activity's credentials from their own variables (§9) whichever option is chosen. Under A1 on production they hold the same values as sign-in. The code does not care which option the owner picks.

### 3. Networking

#### 3.1 A dedicated hostname per environment

**Recommended: N1.** The Activity's `/` mapping targets its own hostname, for example `cmd-activity.labxp.io` and `cmd-dev-activity.labxp.io`. Both reach the same tunnel, Caddy and server as today. The name is one level under `labxp.io`, like the existing hosts, so the zone's existing edge certificate covers it.

- **Caddy** gets a second site block for that host, importing the same routes as the main one from a shared snippet, so the `@api` matcher stays single-sourced (the Caddyfile's own comment explains why it must). The only difference is the shell's framing headers: `frame-ancestors https://discord.com https://*.discord.com https://*.discordsays.com`, and no `X-Frame-Options`. The exact ancestor list is confirmed in Phase 0 (U4).
- **The server** knows Activity traffic by the request's `Host` (`CMDCTRL_ACTIVITY_HOST`), which only Caddy and the tunnel set. It uses that for the cookie attributes (§3.3) and the rate-limit keys (§3.5).
- **Rejected, N2: the same hostnames, with framing allowed to Discord site-wide.** This needs no new hostname, but the server cannot tell Activity requests from website ones (a same-origin GET carries no `Origin`), so cookies and rate limits could not differ by entry point.

The proxy's iframe origin is `<app_id>.discordsays.com`, so that host goes into `CMDCTRL_ALLOWED_ORIGINS` for the WebSocket's Origin check. The `connect-src 'self'` and `img-src 'self'` rules already fit, because inside the Activity "self" is the proxy.

#### 3.2 The game WebSocket

The table's socket is `wss://<app_id>.discordsays.com/ws?game=…&token=…`, through the `/` mapping, with no code change. Discord says WebSockets are supported. Not documented, and measured in Phase 0 (U14): message size and idle limits through the Worker, and whether permessage-deflate (ADR 0205) is negotiated end to end. Our frames are about 70 KB, and the hub pings every 54 s. The reconnect ladder (`ws.ts`) and the dead-session probe are unchanged.

#### 3.3 Card images and the session cookie

`<img src="/cards/…/image">` needs the cookie, and inside Discord's iframe a `SameSite=Lax` cookie is not sent. So:

- **On the Activity host, the session cookie is `SameSite=None; Secure; Partitioned`**, host-only. That is Discord's documented pattern, and Go's `http.Cookie` has `Partitioned` (the server is on Go 1.27). The website's cookie is unchanged.
- **Fallback if Phase 0 finds partitioned cookies unreliable on a platform** (U5, iOS especially): an image-scoped token on card image URLs, valid only for `/cards/*/image` and only for the session's life. The cost is a browser cache that does not survive a new session. The route stays gated either way. Making `/cards/{id}/image` public is rejected, for ADR 0042's reason.
- **On the Activity host, card images are served `Cache-Control: private`.** The proxy's own site is "publicly reachable" in any browser. If Discord's proxy honours `public` (U6), a gated image cached there could be fetched by anyone at `<app_id>.discordsays.com/cards/…`, which would make it the public image proxy ADR 0042 refused.
- **Scryfall.** Card data stays public on the website (the catalogue), so the Activity puts no sign-in or chat-server gate in front of card data (PLAN.md §3). It is one more client of the same pages. The artist credit and the no-cropping rule apply unchanged.

#### 3.4 What breaks under Discord's CSP, and the fixes

| What | Why it breaks | Fix |
|---|---|---|
| Every external link (22 files) | sandboxed iframe | one `openExternal(url)` helper that calls `openExternalLink` in Activity mode and opens a new tab otherwise; a source-scan test fails on a raw `window.open` or `target="_blank"` elsewhere |
| "Sign in with Discord" | it navigates the frame to `discord.com` | hidden in Activity mode; the SDK signs in (§4) |
| The service worker | partitioned storage inside a third-party frame, and a cached shell inside Discord would be one more thing to go stale | not registered in Activity mode (`pwa.ts`) |
| Fonts, images, API, `/ws` | none: all same-origin through `/` | nothing |
| Discord avatars | none: the server's avatar cache serves them | nothing |

No other URL mappings are needed. Everything the client loads is our own.

#### 3.5 Client addresses and rate limits

The proxy hides users' IPs, so every Activity request reaches us from Discord's Cloudflare egress. Left alone, the per-IP buckets (table creation, guest practice, deck checks) would put every Activity player in one bucket, and one busy evening would lock everyone out.

On the Activity host, the limiters key on the **principal** (the account, or for a guest the session) instead of the IP. A route that is per-IP and unauthenticated on the website requires an Activity sign-in first on the Activity host. That way the Activity host cannot be used, from an ordinary browser, as a way round the per-IP limits. If Phase 0 shows that the signed proxy payload names the Discord user (U2), the pre-sign-in route (`POST /auth/discord/activity`) keys on that. Otherwise it gets its own global bucket, sized for a launch-night burst. Logs record the principal for Activity traffic, not an IP.

#### 3.6 Bot Fight Mode

Discord's proxy is a Cloudflare Worker fetching our hostname, and Bot Fight Mode challenges datacenter traffic. Discord does not document how its proxy fares against a zone with Bot Fight Mode on, and a forum report says Workers subrequests have been challenged by it. This is **Phase 0's first test** (U1). A challenged request would break the Activity outright: an HTML challenge page where the shell or JSON should be.

This ADR does not propose any skip, allow rule or bypass for Bot Fight Mode on `labxp.io`; the owner keeps it on. If the proxy is challenged, the Activity cannot ship as designed, and the work stops at Phase 0 for an owner decision (open question 13).

### 4. Sign-in

**Recommended: S1, the SDK's code mapped onto our own sessions.**

1. **Handshake.** `new DiscordSDK(clientId)` (the client ID comes from `GET /community` (ADR 0156 §6), which gains `activity_client_id` when the feature is on), then `await sdk.ready()`.
2. **Authorize.** `sdk.commands.authorize` with `response_type: "code"`, `prompt: "none"`, a PKCE `S256` challenge, and the scopes `identify`, `rpc.activities.write`, plus `guilds.members.read` while the gate is on. The first launch shows Discord's consent modal; after that it is silent.
3. **Exchange, on our server.** `POST /auth/discord/activity {code, code_verifier, instance_id}`, on the Activity host only. The server:
   - exchanges the code with the Activity app's client secret, **without** `redirect_uri` (the authorize request had none; Discord's sample omits it, U11), and fetches `/users/@me`;
   - **confirms the instance** with `GET /applications/{app}/activity-instances/{instance_id}` and the bot token: it exists, the caller is in its `users`, and its location is a guild channel in an allowed guild (§6, open question 3). Anything else is a 403, and nothing is written;
   - runs **the gate unchanged**: `signInGate`'s rules, admins and the exemption list included.
4. **Admitted:** the `users` and `identities` rows are upserted as at a website sign-in, the refresh token is stored under A1, the access check is recorded, and the server mints an **`identified` session** carrying the `UserID`. If the caller already holds a seat at the table bound to this instance (§5), it also mints that seat's `player` session.
5. **Refused** (a C2 member without the Elder Dragon Highlander role): no account, no row, nothing stored (ADR 0149). The server mints a short **instance ticket**, signed, bound to `(instance_id, discord_id, display name)` and good for 10 minutes. It can do one thing: take a guest seat or a spectator place at the table bound to that instance (§5). This is ADR 0156 §1 item 4's invite rule, with the instance in the place of the host's link: the people in the voice channel vouch for this table, and for nothing else.
6. **Authenticate.** The answer carries the Discord **access token**, which the client hands to `sdk.commands.authenticate` so that `setActivity` and the other SDK commands work. The website never gives the browser a Discord token. The Activity has to, and the token lives in that tab's memory only. Its scopes are the three above.
7. **Lifetime.** An Activity sign-in session lasts `CMDCTRL_SESSION_TTL` (12 h), not 30 days. It still carries the `UserID`, so it can be revoked, and ban sync ends it at once. It is not renewed through `POST /me/session`; the client signs in again silently, at launch and at half-life. Each new sign-in is a fresh gate check with a fresh access token. localStorage is used as on the website, but nothing depends on it surviving (U5).

**Rejected:**

- **S2, reusing the website session.** The Activity runs on another origin with partitioned storage, so the website's token is not there, and a top-level OAuth redirect cannot run inside the frame.
- **S3, the signed proxy payload as the identity.** Its fields are not documented (U2). It gives no `guilds.members.read`, so the gate could not run. It is kept only as an extra check (§3.5).

### 5. Joining a game

#### 5.1 One table per instance

**Recommended: J1, the instance is the table.** This is open question 5.

- **Binding.** The server keeps `instance → table` in memory, and on the `games` row as `activity_channel_id`, so a table survives a restart knowing its channel. An instance has at most one table, and a table at most one channel.
- **Launch.** The Activity opens to the table bound to its instance, or, if the instance is in an ADR 0159 voice room, to that room's table (`games.voice_channel_id` matches the instance's `channel_id`).
- **No table yet.** A person with an account sees **Start a table here**: `POST /games` with `activity_instance_id`, the same form and limits as on the website. They become its creator and host (ADR 0075 §2.1). Someone without an account sees "Someone with a CMD & CTRL account can start a table here", with **Try a practice game** (the guest practice table, ADR 0156 §2).
- **Seats by click, spectators by default.** Everyone in the instance sees the table's lobby. **Take a seat** claims a free seat with no invite link (`POST /games/{id}/activity/seat`): the server confirms the caller is in the bound instance (the instances API, cached for 30 s per instance). An account holder sits as themselves. A refused member sits as a guest, through the instance ticket. Anyone who has not taken a seat watches with a spectator session: the ordinary spectator view, every hand hidden. Discord's guidance is to let people who drop in watch until they can join without disrupting the game, and that is what this does.
- **Everything else is the table as it is today.** Bots, deck selection, host controls, remove and block (ADR 0075 §2.6), away seats (ADR 0177) and idle closing (ADR 0211) apply unchanged. A blocked person in the instance gets the removed notice.
- **Website players can join the same table** by its invite link, and play from the browser (open question 7). The table is a normal table, and the Activity is one way to sit at it.
- **When the instance ends** (the last participant leaves, and the instances API answers 404 on the sweep), the binding is dropped and the table carries on under the normal rules. A **rematch** (ADR 0164) moves the binding to the new table, so the next game opens in the same Activity.
- **A guest's seat across a reload.** A guest's seat is tied to a session with no user. If a guest closes the Activity and opens it again in the same instance, the server matches their Discord id to the seat **in memory only**, for the life of the instance, and never writes it down (open question 12). The website's guest has no such path, and needs none.

**Rejected:**

- **J2, the instance as a window onto the normal Tables page.** Each person picks the table themselves. It is simpler, but the room is not the table. It is Phase 1's behaviour (§10), on the way to J1.
- **J3, several tables per instance.** Discord gives one instance per channel at a time (U8), and two pods would normally take two channels. The extra screen is not worth it for the beta.

#### 5.2 One voice channel per table, either way

The issue asks whether it is one room per table either way. **Yes.**

- A table bound to an Activity instance uses the channel the Activity was launched in as its voice. **The voice manager never creates an ADR 0159 room for it.** If the host had voice rooms switched on, the switch shows "Playing in a Discord voice channel" and is disabled.
- An Activity launched **in** an ADR 0159 room opens that room's table. The room's permission overwrites still decide who can be in it, so only the seated players (and server staff) are there.
- The game view's `voice` object (ADR 0159 §6) gains the status `activity`. A website player at the same table sees **Join the voice channel**, with the channel's link, when they are a member of that server.

### 6. Where it can be launched

**Recommended: the C2 server's voice channels only, for the beta** (open question 3).

- The server refuses a sign-in whose instance location is not a guild channel in `CMDCTRL_ACTIVITY_GUILD_IDS`, which defaults to the community guild. The Entry Point command is registered for guild installs and guild contexts only, so it is not offered in DMs.
- After verification the Activity is playable in any server that adds the app, so this check is the only thing keeping it to the C2 community. ADR 0148's caps assume that community.
- Discord's Activities also run in **text channels**. The guild check alone allows them; refusing them takes one more call, `GET /channels/{id}` with the bot token, to read the channel's type. The premise here is voice. Whether to refuse a text-channel launch is part of open question 3.
- **Discovery is off** (open question 4). There is no App Directory listing and no App Launcher listing beyond the server the app is installed in. The beta is for the C2 community (PLAN.md §1), and being listed would invite load the caps are not set for.

### 7. Layout

- **Focused mode** is the table as it is. Discord's frame is smaller than the window, and the docs give no size (U7). If the frame is under the tablet minimum (820 CSS px wide, ADR 0156 §3), the table shows a one-line hint ("Make Discord's window larger, or pop the Activity out") rather than a cramped board. Phase 0 measures it at 1366×768, 1920×1080 and 2560×1440.
- **Picture-in-picture and grid** are too small to play in. Subscribed through `subscribeToLayoutModeUpdatesCompat`, they show a status card instead of the board: whose turn it is, whether the game is waiting on you, life totals. Interactive PiP is switched off on desktop (`setConfig`, `use_interactive_pip: false`).
- **Tablets** (iPad, Android tablets): enabled with iOS and Android in Phase 3. The portal's tablet default is landscape; the safe-area variables pad the frame; at `THERMAL_STATE_UPDATE` of serious or worse, the signature animations (ADR 0209) are skipped.
- **Phones.** The table is never played below tablet size (PLAN.md §3). The portal switches on iOS and Android as a whole, phones included, and the SDK cannot tell a phone from a tablet. A viewport whose shorter side is under 600 CSS px is a phone. **Recommended: a phone gets a spectator view of the status card and the log, and a "Join from a tablet or computer to take a seat" line** (open question 8). The alternatives are no mobile at all (which loses tablets) or a phone table layout, which PLAN.md rules out.
- **Discord's own conventions** go into every Activity screen, from its quality guidelines: ESC closes our dialogs, and TAB must still reach Discord's own controls. Our keyboard shortcuts are checked against Discord's in Phase 1.

### 8. Rich Presence, and cost and risk

#### 8.1 Rich Presence

The default presence is free: the app's icon, and "Ask to Join" for people who can join. Each Activity client also calls `setActivity` for its own user, when its state changes and at most once every 15 seconds (Discord's own limit is not documented, U10):

- `details`: "In the lobby", "Playing Commander" or "Watching a game";
- `state`: "Turn 7", or "Your turn" when the game waits on them;
- `party.size`: `[seats filled, 4]`; `timestamps.start`: the game's start.
- **Never** the table's name, which is text a user typed, nor hidden information. Whether to show commanders is open question 9.

This answers #612 for **Activity players only**. A browser tab cannot set Discord presence, so website players get none. #3313's own presence (in a game, on the site; drafted as ADR 0213 in #4662) counts an Activity socket like any other.

#### 8.2 Discord's side: verification, and who can play

- **Without verification, the Activity cannot reach the C2 server's players.** An unverified app launches only for its team and testers, in servers of fewer than 25 members. Verification is the team owner's identity check through Stripe, plus the portal's checklist. No fee is mentioned anywhere (U12).
- **Verification is permanent for the bot setting:** the app's Public Bot switch cannot be turned off afterwards. Under A1 that is the C2 bot (§2).
- **Discovery** is a further opt-in, recommended off (§6).
- **Fan Content Policy:** the app has no SKUs and no purchases, ever (PLAN.md §3). Discord's monetization is left off.

#### 8.3 Risks

| # | Risk | Effect | Handling |
|---|---|---|---|
| R1 | Bot Fight Mode challenges the proxy | the Activity cannot load | Phase 0's first test; an owner decision if it fails (§3.6) |
| R2 | Partitioned cookies are unreliable in some client | no card art in the Activity | the image-scoped token fallback (§3.3) |
| R3 | The proxy caches `public` responses | gated art reachable at `discordsays.com` | `private` on the Activity host (§3.3) |
| R4 | Every Activity user shares one IP | per-IP limits lock everyone out, or let anyone round them | principal keys (§3.5) |
| R5 | Discord's frame is under the tablet minimum | a cramped board | the hint, measured in Phase 0 (§7) |
| R6 | A Discord client update changes the SDK or the proxy | the Activity breaks while the website keeps working | `INVALID_COMMAND` handled as Discord advises; the SDK pinned and bumped deliberately; the manual check (§8.4) after a bump |
| R7 | Two entry points drift | a feature works on one and not the other | §8.4 |
| R8 | The Activity takes C2 beyond its community | load past ADR 0148's caps | the guild check, Discovery off (§6) |

#### 8.4 Keeping two entry points in sync, and testing

- **One bundle, one server, one table.** Activity code lives in `client/src/lib/activity/` (handshake, sign-in, presence, layout mode, participants, links, invite dialog) and in the server's Activity routes. The rest of the code does not branch on it, apart from the link helper and the service-worker skip.
- **Caddy shares one route snippet** between the two site blocks, so a new API path cannot reach one host and miss the other.
- **The link test** fails on a raw external link outside the helper (§3.4).
- **Unit tests** use the SDK's `DiscordSDKMock`. **Server tests** fake Discord's token, user, member and activity-instance endpoints with `httptest`, through the existing `DiscordHTTPClient` seam.
- **E2E.** On the dev environment only (`requireDev`), `?activity_mock=1` swaps in `DiscordSDKMock`, and a dev-only route stands in for the Discord exchange. A Playwright project then drives Activity mode: sign in, start a table, take a seat, spectate, PiP status, the phone guard. It runs as a branch E2E on Activity PRs, which are client-interaction changes.
- **A manual matrix on real Discord** (Windows, macOS and Linux desktop; web in Chrome, Firefox and Safari; iPad; an Android tablet; a phone), written down in docs/environments.md. It is run before production is switched on and after every SDK bump. Discord cannot be automated.

### 9. Configuration and deploy

**New server variables.** The feature is on only when all of the required ones are set, and a malformed one fails the boot, as in ADR 0159 §7.

| Variable | What |
|---|---|
| `CMDCTRL_ACTIVITY_HOST` | the Activity hostname Caddy forwards (`cmd-activity.labxp.io`) |
| `CMDCTRL_DISCORD_ACTIVITY_CLIENT_ID` / `_CLIENT_SECRET` | the app the Activity signs in through. Under A1 on production, the sign-in values; on dev, the dev app |
| `CMDCTRL_DISCORD_ACTIVITY_BOT_TOKEN` | for the instances API. Under A1 on production, `CMDCTRL_DISCORD_BOT_TOKEN`; on dev, the dev app's |
| `CMDCTRL_DISCORD_ACTIVITY_PUBLIC_KEY` | optional: verifies the signed proxy headers (stdlib `crypto/ed25519`) |
| `CMDCTRL_ACTIVITY_GUILD_IDS` | where it may launch; defaults to `CMDCTRL_DISCORD_COMMUNITY_GUILD_ID` |
| `CMDCTRL_ALLOWED_ORIGINS` | gains `<app_id>.discordsays.com` (an existing variable) |

**CD** writes them, empty included, as the community variables are. **Deploy** adds the hostnames: the develop tunnel's ingress and DNS in HomeLab's Terraform, and the production tunnel's hostname by hand (it was made by hand; [docs/environments.md](../environments.md#secrets)). Caddy's second site block ships from this repo, like the rest of the Caddyfile.

**Discord app setup (owner, in the Developer Portal, not code):** enable Activities; set the URL mapping `/` → the Activity hostname; choose the supported platforms (web and desktop in Phase 1; iOS and Android in Phase 3); set the orientation defaults (tablets landscape); keep the default Entry Point command ("Launch", which Discord opens itself), and limit its `integration_types` and `contexts` to guild installs and guild channels with the global-command edit endpoint (§6); upload the cover art, background and app tile, and the Rich Presence art (1024×1024); add testers to the dev app; verify the production app; leave Discovery and monetization off; install the app in the C2 server so it shows in the server's App Launcher (U9).

**Privacy (ADR 0149).** The notice gains a paragraph: playing inside Discord sends all game traffic through Discord's proxy (so Discord carries it and we never see the player's IP); the Activity asks for `rpc.activities.write` and sets the player's Discord presence as §8.1 describes; a table played in a voice channel records that channel's id (`games.activity_channel_id`, beside ADR 0159's `voice_channel_id`). A guest's Discord id is held in memory only, for the life of the instance. The new column is about a table, not a person, but it is listed with `voice_channel_id`, and `PRIVACY_SCHEMA_VERSION` is bumped with its migration.

### 10. Phases

Each phase lists what it changes. Each ends with a working, shippable state, and production is switched on only after Phase 2, with the owner.

| Phase | What a player gets | Code | Protocol and API | Auth | Deploy and Discord |
|---|---|---|---|---|---|
| **0. Spike** (time-boxed, about a day, dev only) | nothing yet: the findings, as a comment on #3312 and an amendment here | a throwaway Activity page behind `requireDev`: the SDK handshake, authorize and exchange, a socket echo, an image | none | a dev-only exchange that logs the headers it sees | the dev app, a test server under 25 members, the dev Activity hostname, the `/` mapping |
| **1. The smallest Activity:** the site inside Discord, signed in | launch in a voice channel, sign in silently, land on Tables, start or join a table as on the website, play it; links open through Discord; basic presence | `lib/activity/` (boot, sign-in, presence, links); the link helper and its test; the SW skip; the frame-size hint; the server's Activity routes, cookie attributes, rate-limit keys | `POST /auth/discord/activity`; `GET /community` gains `activity_client_id` | §4 in full, gate included | Caddy site block and snippet; the variables and CD sync; privacy paragraph; the manual matrix (desktop and web) |
| **2. The room is the table** | everyone in the channel sees one table, takes a seat by click or watches; members without the role sit as guests; ADR 0159 rooms open their table; rich presence with turn and party | the instance binding and sweep; seat and spectate by instance; the instance ticket; the rematch move; the voice manager's `activity` status; client lobby and presence; Caddy's `@api` matcher gains `/activity` | `POST /games` takes `activity_instance_id`; `GET /activity/table`; `POST /games/{id}/activity/seat` and `/activity/spectate`; `voice.status: "activity"` in docs/protocol.md; docs/lobby.md | the ticket (§4 item 5) | migration `games.activity_channel_id`; under A1, the bot's auto-leave lands first; **owner verifies the production app**; switched on in production with the owner |
| **3. Layout and tablets** | PiP and grid status card; iPad and Android tablets play; phones watch | layout-mode subscription, status card, safe areas, thermal state, the phone guard | none | none | iOS and Android switched on; orientation defaults; the manual matrix (tablets, a phone) |
| **4. Discord-native extras** | a **Play here** button in a voice room's chat and on `/c2-invite` (A1 only); the invite dialog; sharing the end-of-game board (the existing board screenshot through `openShareMomentDialog`); deep links | bot: a component that answers `LAUNCH_ACTIVITY`; client: invite and share | the Discord attachment upload for share moments | none | none |

Phase 2 is the first phase worth turning on in production. Phase 1 alone makes four people pick the same table by hand.

**Not planned:** speaking indicators on seats. They need `rpc.voice.read`, which the docs list as a scope but which this ADR could not confirm is open to every app. Prompting for hardware acceleration is not planned either: it restarts Discord.

### 11. Slices

Each targets `develop` and references #3312. Branch E2E on every client slice (Activity work is client interaction).

| # | Slice | Model | Needs |
|---|---|---|---|
| 1 | This ADR | — | — |
| 2 | **Phase 0 spike**: dev app, hostname, the throwaway page, the findings comment and ADR amendment (U1–U14) | Opus | 1, owner setup of the dev app |
| 3 | **Server: Activity sign-in** (§4): the exchange without `redirect_uri`, the instances API client in `internal/discord`, the gate path, the session lifetime, the cookie attributes and `private` images on the Activity host, principal rate-limit keys, config and boot checks, `GET /community`, httptest coverage, docs/env.md and docs/lobby.md | Opus | 2 |
| 4 | **Client: Activity mode** (§3.4, §4, §7 hint): `lib/activity/`, the link helper and test, the SW skip, sign-in UI, basic presence, the mock switch and the E2E project | Opus | 3 |
| 5 | **Deploy**: Caddy snippet and site block, CD sync, HomeLab hostname (an owner or HomeLab PR), privacy paragraph, the manual matrix in docs/environments.md | Sonnet | 3 |
| 6 | **Server: the instance is the table** (§5): binding, migration, sweep, seat and spectate routes, the ticket, the rematch move, the voice manager's `activity` status, docs/protocol.md | Opus | 3 |
| 7 | **Client: the room is the table** (§5, §8.1): the instance lobby, seat and watch, guest copy, presence with turn and party, the website's Join the voice channel | Opus | 4, 6 |
| 8 | **Layout and tablets** (§7): the status card, safe areas, thermal state, the phone guard; screenshots at 1440 px, 820 px and a phone | Sonnet | 7 |
| 9 | **Discord-native extras** (Phase 4): `LAUNCH_ACTIVITY` buttons, the invite dialog, share moments | Sonnet | 7, A1 |
| 10 | **Bot: leave unlisted servers** (A1 only, §2): on `GuildCreate` and at boot, leave a guild that is neither on `CMDCTRL_DISCORD_GUILD_IDS` nor the community guild, with a log line; ADR 0004 §3 amended. Must be deployed before the app is verified | Sonnet | 1 |

## What the docs do not settle

Phase 0 answers these, apart from U12 and U13, which the owner can read off the Developer Portal's App Verification tab. Each one is marked where it matters above.

- **U1.** Whether Bot Fight Mode on `labxp.io` challenges Discord's proxy. No Discord or Cloudflare document covers it; one forum report says Workers subrequests have been challenged.
- **U2.** Which headers reach us through the proxy (`Host`, `Origin`, `CF-Connecting-IP`, `X-Forwarded-For`), what `X-Discord-Proxy-Payload`'s "user context" holds, and whether proxy authentication has to be switched on.
- **U3.** Whether the `/` mapping carries every path, `/ws` included, without a `/.proxy/` prefix. Discord's lifecycle sample fetches `/.proxy/api/token`; its networking guide uses plain paths.
- **U4.** Whether the proxy passes our framing headers through as they are, and which ancestor origins the iframe has on web, desktop and mobile.
- **U5.** Whether partitioned cookies and localStorage work and persist inside the iframe on each platform, iOS above all.
- **U6.** Whether the proxy caches responses we mark `public`.
- **U7.** The focused-mode frame size on desktop, and whether a desktop Activity can be popped out into its own window.
- **U8.** That a channel has one instance of an app at a time. This is inferred from the instance id's shape (`i-<launch>-gc-<guild>-<channel>`) and from "Join Application", not stated.
- **U9.** How a verified Activity that is not in Discovery is offered to the members of a server it is installed in.
- **U10.** Any rate limit on `setActivity`.
- **U11.** That the token exchange works without `redirect_uri`. Discord's sample omits it; our exchange always sends one.
- **U12.** Whether verification costs anything, and whether Discord reviews an app's content (Magic card art under the Fan Content Policy) for verification.
- **U13.** The conflict between the gateway page and the help centre's 2026-10-26 notice on privileged intents after verification. The ADR follows the newer notice.
- **U14.** WebSocket limits through the Worker: message size, idle timeout, and permessage-deflate.

## Open questions for the owner

Each has a recommendation, the option the ADR is written for.

1. **Which Discord application?** **A1, the existing app** (one consent, one identity, the bot can launch the Activity; the bot becomes public for good once verified, so it must leave servers not on its allow-list), or **A2, a separate Activity app** (the bot untouched; two consents; no launch buttons). Recommended: **A1**, with the bot leaving unlisted servers. The develop preview gets its own unverified app either way.
2. **Will you verify the production app?** It needs your identity checked through Stripe, and it is the only way the Activity reaches C2 server members. An unverified one launches only for your developer team (up to 100) and app testers (up to 50), all your Discord friends, in servers of fewer than 25 members. Recommended: **yes, before Phase 2 goes to production.**
3. **Where may it be launched?** **The C2 server's channels only**, any server that adds the app, or DMs and group DMs too. And within the C2 server: **voice channels only**, or text channels as well? Recommended: **C2 server, voice channels only**, for the beta.
4. **Discovery (App Directory and App Launcher listing)?** Recommended: **off** for the beta.
5. **Joining:** **one table per instance, take a seat by click, everyone else watches** (J1), or seat people automatically in the order they arrive, or a list of tables per instance? Recommended: **J1 with click to sit.**
6. **C2 members without the Elder Dragon Highlander role in the channel:** **seated as guests, vouched for by the instance** (as ADR 0156 does for a host's invite link), or allowed to watch only? Recommended: **guests**.
7. **Mixed tables:** may website players join an Activity table by its invite link? Recommended: **yes**.
8. **Phones in the Activity:** **watch only, with "Join from a tablet or computer"**; no mobile at all (which also loses tablets); or a phone table layout (against PLAN.md §3)? Recommended: **watch only**.
9. **Rich Presence content:** the generic lobby, game, turn and party line of §8.1, or also the commanders being played (visible to anyone who can see the player's profile)? Recommended: **generic only**.
10. **Spectators in voice:** in an open channel, people who are not seated hear the table and watch the spectator view (no hands). Acceptable, or should an Activity table be watchable only by its players? Recommended: **acceptable**. It is the drop-in behaviour Discord designs for, and the spectator view hides every hand.
11. **Session lifetime:** Activity sign-ins last 12 h and are refreshed silently at each launch, not 30 days. Recommended: **12 h**.
12. **A guest's seat across a reload:** hold a guest's Discord id in memory, never stored, for the instance's life, so they can reopen the Activity and keep their seat? Recommended: **yes**.
13. **If Phase 0 finds that Bot Fight Mode challenges Discord's proxy,** the Activity cannot ship on `labxp.io` as it is. The ADR stops there for your decision and does not recommend weakening the zone's protection.
14. **Testers for the dev app:** who are they? Up to 50 app testers, each your Discord friend, in a test server of under 25 members.

## Calls made here

The owner can overturn any of these in review.

- **Same bundle, runtime switch**, with the SDK in its own lazy chunk. A second build would be a second thing to keep in step.
- **A dedicated Activity hostname** (N1), so the server can tell the two entry points apart without trusting a header a browser can set.
- **The gate is unchanged**, and the instance stands in for the invite link for a refused member, exactly as far as ADR 0156's invite rule goes and no further: one table, as a guest, no account.
- **The access token reaches the Activity tab** because `authenticate` needs it. It is in memory only, with three scopes.
- **`Cache-Control: private` for images on the Activity host**, and principal-keyed rate limits there.
- **No ADR 0159 room for an Activity table**, and an Activity in a 0159 room opens that room's table.
- **Presence never shows the table's name.**
- **No hardware-acceleration prompt.**
- **A phone is a viewport whose shorter side is under 600 CSS px.** The SDK reports only `desktop` or `mobile`.
- **Timings:** presence at most once every 15 s, the instance ticket good for 10 minutes, the instances API cached for 30 s per instance.
- **Production is switched on with Phase 2**, not Phase 1.
- **Max participants is left unset in the portal**, because spectators are welcome. Setting it to 4 would tell people a fifth person cannot join.

## Consequences

- **A pod in the C2 server can play without leaving Discord:** launch, sign in once, take a seat. Voice, video and screen share are Discord's.
- **There are two ways in to keep working.** That costs the link helper, the Caddy snippet, the E2E project and a manual pass on real Discord clients after every SDK bump.
- **Discord sits in the middle of every Activity request.** It carries our traffic, sets the frame, and can change the SDK or the proxy. The website is unaffected when it does.
- **The C2 app becomes verified, and its bot public** (A1). Strangers can add the bot to their servers; it leaves them at once.
- **Player IPs disappear from Activity traffic.** That is good for privacy, but abuse handling for Activity players rests on accounts and Discord ids, not addresses.
- **New surface:** one hostname per environment, six variables, one column, five routes, a game-view status, an SDK dependency, and a Discord app per environment.

## Back-out

Unset `CMDCTRL_ACTIVITY_HOST` (through the variable, then a deploy). Every Activity route then answers 503 and the website is unchanged. Switching Activities off in the Developer Portal removes the Activity from Discord. Tables started in an Activity stay ordinary tables; `games.activity_channel_id` is nullable and unused with the feature off.

## Sources

Read on 2026-10-30. The docs site pages are built from `discord/discord-api-docs` at commit `e3ea589033`. The help-centre articles were read through the help centre's public article API, as dated.

**Discord developer documentation**

- Activities overview: <https://docs.discord.com/developers/activities/overview>
- How Activities work: <https://docs.discord.com/developers/activities/how-activities-work>
- Building an Activity: <https://docs.discord.com/developers/activities/building-an-activity>
- Networking (the proxy, Cloudflare Workers, WebSockets, cookies): <https://docs.discord.com/developers/activities/development-guides/networking>
- Local development (URL mappings, the CSP and its exceptions): <https://docs.discord.com/developers/activities/development-guides/local-development>
- Multiplayer experience (instances, participants, the activity-instances API, proxy headers): <https://docs.discord.com/developers/activities/development-guides/multiplayer-experience>
- Mobile: <https://docs.discord.com/developers/activities/development-guides/mobile>
- Layout: <https://docs.discord.com/developers/activities/development-guides/layout>
- User actions (external links, invite dialog, share moments, Entry Point command): <https://docs.discord.com/developers/activities/development-guides/user-actions>
- Production readiness: <https://docs.discord.com/developers/activities/development-guides/production-readiness>
- Growth and referrals (activity links, `custom_id`): <https://docs.discord.com/developers/activities/development-guides/growth-and-referrals>
- Design patterns: <https://docs.discord.com/developers/activities/design-patterns>
- Embedded App SDK reference: <https://docs.discord.com/developers/developer-tools/embedded-app-sdk>
- Rich Presence with the Embedded App SDK: <https://docs.discord.com/developers/rich-presence/using-with-the-embedded-app-sdk>
- Enabling Discovery: <https://docs.discord.com/developers/discovery/enabling-discovery>
- Gateway, privileged intents: <https://docs.discord.com/developers/events/gateway#privileged-intents>

**Discord developer help centre**

- What are Verified and Unverified Activities? (updated 2026-10-28): <https://support-dev.discord.com/hc/en-us/articles/26576097154199>
- How Can Users Discover and Play My Activity? (2026-10-24): <https://support-dev.discord.com/hc/en-us/articles/21204493235991>
- Unverified Activity Safety (2026-10-26): <https://support-dev.discord.com/hc/en-us/articles/21692628851351>
- How Do I Get My App Verified? (2026-10-29): <https://support-dev.discord.com/hc/en-us/articles/23926564536471>
- Changes to Privileged Intent Access for Discord Apps (2026-10-26): <https://support-dev.discord.com/hc/en-us/articles/40281523410967>

**The SDK**

- `discord/embedded-app-sdk` v2.5.0, `src/Discord.ts` (the query parameters, `platform`) and `src/mock.ts` (`DiscordSDKMock`): <https://github.com/discord/embedded-app-sdk>
