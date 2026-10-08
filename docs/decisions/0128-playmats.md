# ADR 0128 — Playmats: an image on your account that shows behind your battlefield

**Status:** Accepted · 2026-10-07 · outside a numbered sprint (S66 is bot play; this is table UX, like #2483 and #2486). The owner asked for it on 2026-10-07; the decisions below were set in the request brief and are recorded with their reasons.
**Issues:** [#2495](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2495) (this change); [#2515](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2515) (§11, three saved playmats and the best-size fit).
**Owner request:** 2026-10-07: "Adding a playmat to the view, so you can upload or link an image to your account and the battlefield will have the image, like a playmat in person on paper."
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-07. I ran `git fetch origin` and listed `docs/decisions/` on `origin/develop` and on every live remote head (`main`, `docs/1961-auto-answer-adr`, `docs/issue-audit`, `feat/750-conditional-block-restrictions`, `feat/opponent-top-row`, `fix/2449-reequip-loop`, `fix/caddy-reload-admin-off`, `wip/836-one-click-default`). The highest number is 0127, on PR #2490's branch `docs/1961-auto-answer-adr`. `origin/develop` stops at 0126. No open issue's claim comment reserves a higher one. This ADR takes **0128**. The same sweep covered the migration number: the highest on any head is `0010_seat_agent.sql`, so the new migration is **0011**.
**Builds on:** [ADR 0051](0051-user-database.md) (the user database; a playmat belongs to a `users` row), [ADR 0110](0110-remember-me.md) §4 (`/me/*` and the per-person and per-device settings split), [ADR 0017](0017-bug-report-button.md) §6 (the uuid-keyed, inert-headers store this one copies), [ADR 0075](0075-table-settings-and-host-controls.md) (the room stamps what is not engine state), [ADR 0124](0124-admin-views-accounts-games-and-who-is-on-now.md) (the admin account view).

---

## Context

A paper Commander table is not a bare table. Each player brings a mat: a picture under their cards, which says whose area is whose and is half the reason anyone owns one. Our table draws four panels on a flat background. The request is the same object, in the browser: a person picks an image once, and every player at the table sees it behind that person's battlefield.

The feature is small to describe and has three places where a careless version is dangerous.

1. **A link is an outbound request from our server, on a stranger's behalf.** "Paste a URL" means the server fetches an address a user chose. That is the classic server-side request forgery (SSRF) shape: `http://169.254.169.254/` reaches the cloud metadata service, `http://127.0.0.1:8080/admin/…` reaches this very server, and a hostname whose DNS answer changes between the check and the connection gets round any check made on the name.
2. **A link is also a request from every other player's browser, if we let it be.** If the board put the pasted URL in an `<img src>`, every person at the table would contact a host the owner chose: their IP address is disclosed to it, the image changes or disappears when the owner edits the host, an `http` link is mixed content, and the host sees who is playing when.
3. **The image is untrusted bytes shown to other people.** A phone photo of a real playmat carries the GPS position of the room it was taken in in its EXIF block. A file can be a different type than its name says, an SVG is a script, and a 100 KB PNG can declare itself 100 000 by 100 000 pixels and cost 40 GB to decode.

## Decisions

### 1. Who can have a playmat

Only a **signed-in account backed by the database** (a `users` row, [ADR 0051](0051-user-database.md)). One playmat per account (**up to three saved, one active, since §11**).

A guest has no account to hang it on, the admin token is a credential and not a person, and a server with no database has no `users` table. All three get **403** from every `/me/playmat` route, the same way `GET /me/settings` refuses them ([ADR 0110](0110-remember-me.md) §4, #1154: never 401, which would sign the browser out), and the client hides the control. A server with a database but no data directory has a user to own a playmat and nowhere to put the file: that is `{"enabled": false}` on `GET` and **503** on a write, as bug-report attachments report themselves disabled without `CMDCTRL_DATA_DIR`.

An agent seat is a guest ([ADR 0122](0122-an-agent-at-the-table-a-local-mcp-seat.md) §7) and a bot has no account, so neither has one. The room's stamp skips both even if one were set, so the rule does not depend on the lobby being right.

### 2. Everyone sees your playmat behind your battlefield; each device chooses how many

The mat belongs to the seat's owner and is drawn behind **that seat's** battlefield area on every viewer's board, as it would be on paper. It is not drawn on anyone else's area. The server puts it on the wire, so it is public to the table, like the seat's name.

Not everyone wants other people's art. A player may find it distracting, or be on a phone, or a low-power laptop with four large images to decode. So there is a per-device setting, `display.playmats`:

| Value | Meaning |
|---|---|
| `all` (default) | every seat's playmat |
| `mine` | only your own |
| `off` | none |

It is **per device**, in the `SYNCED_FIELDS` map's `"device"` scope ([ADR 0110](0110-remember-me.md) §4, owner answer 5), because what it answers is about this screen, not about the person: the same person wants `off` on their phone and `all` on their desk. This adds a line to the per-device list that `settingsSyncFields.test.ts` pins as owner answer 5's, with this ADR as the reason. It does not bump `SETTINGS_VERSION`: the shallow merge fills it for any older blob and `migrate` replaces an unknown value with the default, as `tableLayout` did when it gained a value (#2336). It is shown to everyone, a guest included, because a guest has a screen too, and it applies to the stored value live (no reload).

### 3. Upload or link, but the server always stores the bytes

A link is **fetched once by the server and stored exactly like an upload**. The pasted URL is not stored, is not put on the wire and is never shown to another player. The only URL a client ever receives is our own `/playmats/<uuid>`, and the client refuses to load anything else (`isPlaymatPath`).

This is the decision the others hang on, so the reasons are in full. Hotlinking, by which I mean storing the link and letting each browser load it, would:

- **Leak the other players' IP addresses** to a host the owner picked, and tell it when each of them is playing. At a four-person table that is three people who agreed to play a game and not to contact a stranger's server.
- **Break when the link rots.** A mat that goes 404 mid-game leaves a table of broken-image boxes, and a host that serves something different tomorrow changes what everyone sees without anyone having changed anything.
- **Allow mixed content and tracking.** An `http` image on an `https` page is blocked or downgrades the page, and a URL with a per-viewer query string is a tracking pixel.
- **Skip every check below.** The decode, the size caps and the EXIF strip only happen to bytes we have.

Fetching server-side makes the link a one-time import, the same as a file picker. The cost is that the server makes an outbound request, which is the next decision.

**The fetch is an SSRF guard** (`server/internal/playmat/fetch.go`):

- **`https` only.** `http` is refused, and so is a redirect to `http`. An `http` image is mixed content on our `https` site and an unauthenticated one on the way to the server, and a person who has an `http`-only image can upload the file. The cost of this rule is one error message.
- **The address is checked where it is dialled, on the resolved IP.** The `http.Transport`'s `DialContext` resolves the host itself, checks **every** address it got, refuses the host if any one is not allowed, and dials the IP it checked (not the name again). The `net.Dialer.Control` hook checks the connected address a second time at connect. A name that resolves to `127.0.0.1`, or that answers with a public address to a check and a private one to the connection (DNS rebinding), cannot get through, because the check is the connection and no second lookup is made.
- **Refused ranges.** Loopback; RFC 1918 private; link-local, which includes `169.254.169.254`; carrier-grade NAT `100.64.0.0/10`; multicast; unspecified; IPv6 unique-local `fc00::/7` and link-local `fe80::/10`; `0.0.0.0/8`; the reserved `240.0.0.0/4` (which holds the broadcast address); and the documentation and benchmarking ranges. An IPv4 address written as IPv6 (`::ffff:127.0.0.1`), through NAT64 (`64:ff9b::7f00:1`) or through 6to4 (`2002:7f00:1::`) is judged by the IPv4 address inside it, so the IPv6 spelling is not a way round. A decimal or hex IPv4 (`2130706433`, `0x7f.1`) is parsed to the address it is before the check.
- **Every redirect is re-checked**, because each is a new request through the same dialer. At most **3**. `Referer` is removed on a redirect, because `net/http` adds the previous URL and that would tell the next host which link was pasted.
- **10 seconds** for the whole request, dial to last byte, and a `LimitReader` cap on the body at 10 MB (one byte past it tells exactly-the-cap from over). A declared `Content-Length` over the cap is refused before reading.
- **No proxy from the environment, no cookie jar, no credentials**, and a URL with a user name or password is refused.
- **The error says what to do, not what the guard saw.** A refusal reads "could not fetch that link: …" with a reason the person can act on. It never echoes the resolved address or the transport's error, because that would turn the route into a probe of this server's network.

The tests use an `httptest` TLS server and an injectable resolver and allow-list, so none reach the internet, and the refusals run against the production policy.

### 4. Validate by decoding, never by what a client says

The `Content-Type`, the part's file name and the URL's extension are all ignored. The bytes are decoded.

1. `image.DecodeConfig` reads the format and dimensions **from the header only**. Anything but PNG, JPEG or WebP is refused (**415**). That is a content test, so a GIF or an SVG is refused however it is named, and an HTML file named `mat.png` is refused.
2. The pixel count is checked **before** any pixel buffer exists: over **40 megapixels** is refused (**413**). 40 MP is above every phone camera and below a bomb. The multiplication is in `int64`.
3. Only then a full decode, which is the real "is this an image" test.
4. The image is **re-encoded**: scaled so the long edge is at most **2560 px** (a battlefield area on a 4K screen is under 2000 px wide, and every other player downloads this file), flattened onto white (JPEG has no alpha), and written as a **JPEG at quality 85**. The standard library encoder writes only the segments it writes itself, so **EXIF** (the GPS of a phone photo), XMP, ICC profiles and anything appended after the image are gone. The EXIF **orientation** is read and applied to the pixels first, so a portrait photo is not stored on its side.
5. The upload cap is **10 MB**, enforced on the request reader as well as the decoder's input.

WebP is accepted through `golang.org/x/image/webp` (Go-team maintained; it is a new dependency of `server/go.mod`), and `golang.org/x/image/draw` does the scaling. Decoding is bounded to **2 at a time** (`ErrBusy`, **503** with `Retry-After`): a 40 MP image is 160 MB of RGBA, and four people uploading at once on a small VPS should be a short queue and not a restart.

### 5. Storage: the bugstore pattern, plus a pointer on the user

The files are `$CMDCTRL_DATA_DIR/playmats/<uuid>.jpg`, written to a temp name and renamed so a reader never sees half an image at a URL that has been handed out, as bugstore does ([ADR 0017](0017-bug-report-button.md) §6). The uuid is a v4 and a **fresh one for every image**, so it is a capability (122 bits) and a replacement changes the URL, which is what lets the response be `immutable`. A path segment that is not exactly the lowercase uuid shape never reaches `filepath.Join`.

The pointer is **migration 0011**: `ALTER TABLE users ADD COLUMN playmat_id TEXT`, nullable, no default, no backfill, like 0004, 0007 and 0010. It is a column on `users` and not a new table because the relation is one to one and the playmat has nothing else to say about itself (the dimensions are read from the stored file's header when asked).

Replacing or removing a playmat **deletes the old file**, after the row has changed, and the row change reads the previous id and writes the new one in one transaction so two simultaneous replacements cannot both leak a file. A pointer that outlives its file (a restored database, a wiped volume) reads as "no playmat", not as a dead link. With no data directory (or no database) the whole feature reports itself disabled.

A user row is not deleted by any route today, so there is no cleanup on account removal. If one is added it must call `Service.Remove`.

### 6. Routes

Documented in [docs/lobby.md](../lobby.md#playmats-adr-0128) and the `AGENTS.md` endpoint list.

| Route | Who |
|---|---|
| `GET /me/playmat` | a signed-in person: `{enabled, url?, width?, height?}` |
| `PUT /me/playmat` | a signed-in person: multipart upload, part `file`, streamed with no temp file |
| `POST /me/playmat/link` | a signed-in person: `{"url": "https://…"}` |
| `DELETE /me/playmat` | a signed-in person |
| `GET /playmats/{id}` | any session |
| `DELETE /admin/users/{id}/playmat` | admin (decision 9) |

**`GET /playmats/{id}` requires a session**, like `/avatars` and `/cards`, and not a public uuid route like bugstore's. Bugstore's is public because GitHub's image proxy has to fetch it and cannot sign in; nothing has that need here, only our own client does, and an image of someone's living room is worth a session check. The uuid is still a capability for the *table*: any session may fetch any id, because every player must see every other player's mat. There is no per-table check (a viewer can't be asked "are you at a table with this person" cheaply, and the answer would change mid-game) and there is no route that lists or enumerates ids. An `<img>` cannot set an `Authorization` header, so the client appends `?token=`, as `avatarURL` does (`client/src/lib/api.ts`: a `Secure`-flag mismatch or a cleared cookie with a live `localStorage` session makes the cookie alone unreliable). It is a new top-level prefix, so it is in `deploy/Caddyfile`'s `@api` matcher, `client/vite.config.ts` and the service worker's `API_PATH`; without all three it works in `develop` and 404s in production.

The response is `image/jpeg`, fixed (never sniffed, never read from anything a client sent), with `X-Content-Type-Options: nosniff`, `Content-Security-Policy: default-src 'none'; sandbox`, `Referrer-Policy: no-referrer`, `Cross-Origin-Resource-Policy: same-origin`, and `Cache-Control: private, max-age=31536000, immutable`. `private` because a shared cache must not keep a session-gated image. A malformed id and an unknown id are the same **404**.

**Rate limits** use the lobby's existing limiters: a per-IP bucket on the two writes that decode (1 a second, a burst of 10) and a per-person bucket on all three writes (a burst of 5, then one every 6 seconds), because `POST …/link` is the one route that makes this server request something, and an IP bucket alone lets one proxy address share a table's allowance (`perCallerLimit`). Reads ride the avatar bucket.

### 7. Wire

`PlayerView` gains `playmat_url` (`omitempty`), the same-origin `/playmats/<uuid>` of the seat's owner's mat, **identical for every viewer** (it is public, like the seat's name; `FilterViewFor` passes it unchanged).

It is **not engine state**. It is stamped the way `is_host` is ([ADR 0075](0075-table-settings-and-host-controls.md) §2.1): the lobby tells the `Room` what a seat shows (`Room.SetPlaymat`, guarded by a mutex of its own so it can be set from inside an `ApplyExternal` function) and the room writes it onto every captured view. That is why:

- **A change made mid-game reaches the table on the next snapshot**, without a game action. The playmat routes call `Lobby.PlaymatChanged`, which sets the new URL on every table the person holds a seat at and broadcasts that commit through the same path a join uses. There is no new message type.
- It is **not in the snapshot schema**, the undo stack or the replay's game state, so undoing a turn cannot undo a person's choice of picture, `SnapshotSchemaVersion` does not move, and `game.Player` is untouched. (The replay and crash dump hold the captured view, which carries the URL; that is a path on our own server.)
- It is set when a signed-in person sits down (`join`), when a seat is linked to an account (`LinkSeat`, and `LinkPendingSeats` at sign-in), when the practice table seats its human, **and when a restored room is paired with its metadata** (`loadEntry`): an account's mat is read from the account again after a deploy.
- **A bot, an agent and a guest never carry one**: the lobby only binds for a seat with a `UserID`, and `stampPlaymatsLocked` skips `IsBot` and `IsAgent` regardless.
- Ended and archived tables are skipped by `PlaymatChanged`: nobody is looking at them live, and an opened one is stamped from the account afresh.

### 8. Client

- **Settings.** A "Playmat" tab: the per-device choice (`all` / `mine` / `off`), and, when the server says the account can have one, a preview, an **Upload an image** button, a link field with **Use this image**, **Remove**, and the error message the server returned. The preview and board fall back to nothing, not a broken-image box, when an image will not load. The account half is hidden on a 403 and on `enabled: false`, and replaced by a one-line "Sign in with Discord" for a visitor with no session at all. A file over 10 MB is refused before it is sent.
- **Board.** `PlayerPanel`, the one component that draws a player's battlefield area, draws its seat's mat as the panel's **first child**: an absolutely positioned layer at `z-index: -1` (the panel is already a stacking context through `container-type: size`; `isolation: isolate` says so where it is relied on), `pointer-events: none`, `aria-hidden`, an `<img>` with `object-fit: cover` and `object-position: center` under a **scrim** that is `color-mix(in srgb, var(--bg) 58%, transparent)`. The scrim is the page's own background colour, so it follows the skin: a dark skin dims the art toward dark and the light skin toward light, and the cards, chips and text above it keep their contrast either way. It takes no grid space, so it does not move anything Luke's layout rules place.
- **Turned with the area.** An across-table opponent's board is "yours turned 180° about the table's centre" (#2438, #2483, #2488). Its mat is turned with it, so the art faces its owner, as on paper: `.panel.opponent.flipped .playmat img { transform: rotate(180deg) }`, inside the same `min-width: 600px` condition as the layout's turn.
- **An `<img>` and not a CSS background**, because a failed load can be detected: `onerror` drops that URL's layer for the rest of the mount (keyed by the wire URL, not by the `?token=` form, so a renewed session does not retry a failed mat). Nothing is shown.
- The board **only loads** a wire value that is exactly `/playmats/` and a lowercase uuid. A hostile or buggy server value cannot make a browser contact a third party. That is the client half of decision 3.
- The change to `PlayerPanel.svelte` is the one `{#if}` block and one CSS block, so it rebases cleanly over the layout work in flight there.

### 9. Moderation

A playmat is user-chosen art on a table other people sit at. `DELETE /admin/users/{id}/playmat` removes anyone's, behind `requireAdmin` (so admin mode off gets a non-admin's 403 and the player-mode census, `TestEveryAdminCallSiteIsInTheSameAnswerTable`, covers it). It does what the person's own `DELETE` does, and the table sees it go on the next snapshot. It removes an image; it does not ban the feature, and the person can upload another. **The admin account view ([ADR 0124](0124-admin-views-accounts-games-and-who-is-on-now.md)) has no button for it.** Making one is not cheap: the view is read-only by decision, its account body would need a `playmat` flag and a confirm dialog naming the person, and each is a render test; it is filed as a follow-up. Until then the route is `curl`-able with the admin token.

### 10. Amendment (2026-10-07, owner decision): the owner sets the wash

Owner decision: the playmat is "dimmed, cover-fit" by default, "but allow an adjustment for tint/translucence/dark wash". So the scrim's strength in §8 is no longer fixed at 58%: it is the **owner's** choice, and every viewer draws that seat's mat at it.

- **Stored on the account.** Migration 0012 adds `users.playmat_wash` (nullable `INTEGER`, no default, no backfill). NULL is the default, **58**, the scrim this ADR shipped with, so nobody's table changes until they move the slider. `playmat.Service` reads it through the same per-user cache as `playmat_id` (`Wash`) and writes it with `SetWash`, which takes **30 to 90**: the floor keeps cards readable over busy art on someone else's board, the ceiling still shows the image. A stored value outside the range (a hand edit) reads as the default.
- **Set by `PATCH /me/playmat`** with `{"wash": n}`, image or not: it is a preference kept for the next upload. Every `/me/playmat` answer carries `wash` while the feature is enabled. Its rate bucket is one a second with a burst of 5, apart from uploads, because a slider writes often.
- **On the wire as `PlayerView.playmat_wash`**, stamped by the room beside `playmat_url` and only where a URL is (`Room.SetPlaymatWash`, `stampPlaymatsLocked`), so like the URL it is not engine state. The lobby binds it wherever `bindPlaymat` binds the URL, through a small `PlaymatWashSource` interface the service implements, and `PlaymatChanged` re-reads it, so a `PATCH` reaches every live seat the person holds the way an upload does.
- **Drawn** as `color-mix(in srgb, var(--bg) var(--playmat-wash, 58%), transparent)`: the same theme-aware scrim, with `--playmat-wash` set on the seat's `.playmat` layer from the wire. The Settings tab gains a **Darken** slider under the preview; the preview wears the same scrim, follows the slider at once, and the value is saved once the slider has rested (400 ms), so a drag is one write.
- **Unchanged:** who sees a playmat (the per-device `all` / `mine` / `off` choice is the viewer's), the image pipeline and every route above.

### 11. Amendment (2026-10-07, owner request): three saved playmats, and a best-size fit

**Issue:** [#2515](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2515). **Owner request:** 2026-10-07: "Allow up to 3 playmats to be saved and used." and "If the image is not the best dimensions, suggest the best image dimensions, with a button to accept or deny the preferred size. If they choose the ideal one, resize the image to the best size for playing."

**Numbering:** no new ADR; this is a dated section of 0128, as §10 is. The migration number was swept as the first paragraph above did: `git fetch origin`, then `server/internal/db/migrations` on `origin/develop` and on every remote head (93 branches); the highest on any head is `0012_playmat_wash.sql`, and no open issue's claim comment reserves a higher one. This change takes **0013**.

This section **supersedes** what the sections above say about "one playmat per account" (the start of §1, the pointer in §5, the route table in §6, the Settings tab in §8, and the single admin route in §9). The rest of them stands: the SSRF guard (§3), validation by decoding (§4), the file store, the wire field, the serving route and the wash.

#### 11.1 Three slots, at most one active

An account keeps **up to three saved playmats, in slots 1 to 3**, and **at most one is active**: the one the table shows. "None" is allowed: a person can keep saved mats and show none.

- **Storage.** Migration 0013 adds `user_playmats(user_id, slot, playmat_id, width, height, created_at)`, primary key `(user_id, slot)`, `slot` checked to 1 to 3, one image in one slot (a unique index on `playmat_id`). `width` and `height` are written when an image is stored and are `NULL` on a backfilled row, where the service reads the file's header, as v1 did. `users.playmat_id` **stays and now means the active pointer**, so the room's stamp, the lobby's binds and the per-user cache are the code they were: `Service.URL` reads the active pointer.
- **The invariant**, kept by every write and tested: *a non-NULL `users.playmat_id` equals the `playmat_id` of exactly one `user_playmats` row of the same user.* SQLite cannot add a foreign key to an existing column, so it is the service's, not the schema's. Every write takes the user's row first (`UPDATE users SET playmat_id = playmat_id`), reads and writes the slot row and the pointer in one transaction, and updates the cache after the commit, so two simultaneous writes cannot leave a pointer at a row that is gone or leak a file. The migration **backfills** every non-NULL `users.playmat_id` into slot 1, and clears an empty-string pointer (nothing writes one), so an unmatched pointer is impossible from the moment the migration runs. `TestMigration0013BackfillsEveryExistingPlaymatIntoSlotOne` and `checkInvariant` (asked of the database after every write test) pin both.
- **Files.** Switching the active mat never touches a file. Removing a slot deletes its file, and clears the active pointer if that slot was active. Replacing a slot (upload, link or fit) deletes the old file, and moves the active pointer if that slot was active. The old file is deleted after the row changes, as in §5.
- **A saved row whose file is gone** (a restored database, a wiped volume) is left out of the list, not shown as a dead link, as §5 reads a dead pointer as "no playmat". Replacing or removing its slot still works.
- **The wash stays per account** (`users.playmat_wash`, §10), unchanged. **A per-mat wash is a possible follow-up** (a nullable `wash` on `user_playmats`, falling back to the account's); it is not built, because one slider for "whichever mat I am using" was the owner's decision in §10 and nobody has asked to split it.

#### 11.2 Routes

| Route | What |
|---|---|
| `GET /me/playmats` | the occupied slots (url, width, height, `fits`, and the `suggestion` when it does not fit), the active slot, the wash, and the constants (`max_slots`, `ideal_width`, `ideal_height`) |
| `PUT /me/playmats/{slot}` | upload into a slot |
| `POST /me/playmats/{slot}/link` | fetch a link into a slot |
| `POST /me/playmats/{slot}/fit` | `{"x", "y"}`: crop to the best size (11.4) |
| `DELETE /me/playmats/{slot}` | remove a slot's playmat |
| `PUT /me/playmats/active` | `{"slot": n \| null}` |
| `PATCH /me/playmats` | the wash, as §10 had it at `/me/playmat` |
| `DELETE /admin/users/{id}/playmats/{slot}`, `DELETE /admin/users/{id}/playmat` | admin moderation (11.5) |

Every `/me/playmats` answer is the same body (the list), and an upload, link or fit adds `slot`, the slot it wrote, so the client opens the best-size prompt on that slot's `suggestion`. A slot outside 1 to 3, or not a whole number, is **400** on every route that takes one.

- **The v1 singular routes are removed.** `grep` of the server, `cmd/` and the MCP seat found no caller of `/me/playmat` but the Settings tab, which moves with the server in this change, so keeping them would be dead surface. Their tests were rewritten for the plural routes, and `TestTheV1SingularPlaymatRoutesAreGone` pins the removal. **`PATCH` moved with them**, to `PATCH /me/playmats`: a route family with one verb left on the singular path would have been the odd one out.
- **Activation.** *Uploading into an empty slot while no mat is active makes it active.* Any other write leaves the active mat alone, except that replacing or fitting the active slot points the active pointer at the new image (the table must not keep showing a file that was just deleted). `PUT /me/playmats/active` with a slot that holds nothing is **404**; `null` shows none.
- **Rate limits.** Upload, link and fit use the buckets §6 gave upload and link: per IP (1 a second, burst 10) outside auth for the three that decode or fetch, per person (burst 5, then one every 6 s) inside it, shared with remove. Activate shares the wash's cheaper bucket (1 a second, burst 5), as it is a pointer and not an image.
- **The room.** Activating, removing or fitting the active mat, or replacing the active slot, must reach every live seat the person holds. The routes compare `Service.URL` before and after the write and call `Lobby.PlaymatChanged` when it moved, the push v1 used; no new message type. Saving into an inactive slot moves nothing the table can see and so costs no broadcast.

#### 11.3 What fits

Measured first, not assumed. The board draws a seat's battlefield area as `.panel.seat-panel`, and the playmat layer fills it (`inset: 0`, §8). I measured those boxes in Luke's current layout (#2485, #2488, #2492) with Playwright driving a Chromium-based browser against a built server and the Vite client, a four-seat table (one human, three bots) after the mulligan, at four viewport sizes:

| Viewport | Across-table seats (three) | Ratio | Own seat | Ratio |
|---|---|---|---|---|
| 1920 x 1080 | 624 x 350 | 1.784 | 1889 x 650 | 2.906 |
| 2560 x 1440 | 838 x 476 | 1.760 | 2529 x 884 | 2.861 |
| 1600 x 900 | 518 x 287 | 1.804 | 1569 x 533 | 2.944 |
| 1366 x 768 | 440 x 241 | 1.826 | 1335 x 447 | 2.986 |

The paper shape, **24 x 14 in = 12:7 = 1.714**, is within 2.7% (2560), 4.1% (1920), 5.2% (1600) and 6.5% (1366) of the three across-table panels. **The own-seat panel is 2.86 to 2.99:1**, about 70% wider than 12:7: it is the docked, full-width panel with the hand fan over its bottom edge.

**The chosen target is 2400 x 1400, 12:7, in the constants `IdealWidth` and `IdealHeight` (`internal/playmat/fit.go`).** This keeps the brief's default **although the own-seat area is beyond its 15% rule** (which says to use the measured shape then). The reasons, in order of weight:

1. **A mat is seen four times.** Every viewer sees a person's mat on the three across-table panels at 1.76 to 1.83, and only that person sees it on the docked own panel. The paper shape serves the three; a 3:1 target would crop the art to its middle 59% in each of them.
2. **The own panel is not a mat-shaped area.** It is a wide strip with the hand over it; `object-fit: cover` already shows its centred band, which is what a person looking at their own board wants to see through the cards.
3. **The prompt's text is "the shape of a paper playmat".** A person with a photo of a real mat (24 x 14 in) would be told their image is the wrong shape and asked to crop it to a banner.

The cost is that the owner sees the middle 59% of the art's height behind their own cards. If the owner wants the own view exact, change the pair and the table above together; a per-viewer `object-position` for the own panel is a cheaper follow-up that costs the art nothing.

**"Fits"** means the aspect ratio within **5%** of 12:7 (`AspectTolerance`) **and** a long edge of at least **1600 px** (`MinLongEdge`). An image that fits gets no prompt.

**The suggestion** (`Suggest`) is offered only when the shape is off, because a crop of a right-shaped image changes nothing: *the crop rectangle* is the largest 12:7 rectangle that fits inside the stored image, centred, in the stored image's pixels; *the target* is the ideal size when the crop is at least 2400 px wide, and the crop's own size when it is smaller. **The server never upscales.** A small image is cropped to the shape at its own resolution, and `smaller` is true so the prompt can say it may look soft at the table. A right-shaped image that is small has `fits: false` and no `suggestion`: the fix is a bigger image, and the Settings tab says nothing about it.

#### 11.4 The fit

`POST /me/playmats/{slot}/fit` takes `{"x", "y"}`, the crop window's top-left corner in the stored image's pixels. The window's *size* is the server's, so a client names only where it sits along the axis being cropped: the server validates that the rectangle lies inside the image (**400** otherwise), crops the stored image, scales the crop down to the target with the same Catmull-Rom scaler (never up), re-encodes through the same JPEG path (quality 85), writes **a new file under a new uuid** (so the immutable cache of §6 never serves the old image), swaps the slot's pointer and the active pointer if that slot was active, and deletes the old file. The swap is conditional on the slot still holding the image the crop was made from, so a fit that loses a race with a replacement drops its own file and answers **409** (ask again) instead of overwriting the newer upload. Decoding is bounded by the same two-at-a-time limit as uploads (`ErrBusy`, **503**).

- **Fitting a mat that already fits is a 409**, not a silent success, so a client that offers the action where it is not needed finds out. An empty slot is a 404.
- **A fit works from the stored image**, which is already normalised to a long edge of at most 2560 px (§4). That is enough for a landscape source. **A portrait source has lost resolution before the fit sees it:** a 3000 x 4000 phone photo is stored at 1920 x 2560, and the 12:7 crop of that is 1920 x 1120, below the ideal, so the result is the smaller, softer kind and says so. We do not keep originals: they are disk the account would hold twice, and the original is where EXIF (the GPS of a phone photo, §4) lives, and the design promise is that it never reaches disk.
- Each fit re-encodes a JPEG, so it costs one generation of JPEG loss on top of the upload's. Accepting the fit once is the intended path; the Settings tab does not offer it again on a mat that fits.

#### 11.5 Client, and the admin account view

- **Settings, Playmat tab** (`PlaymatSettings.svelte`): three slot cards, each with a thumbnail or an empty "Add a playmat" state, a **Using** badge on the active one, and **Use / Stop using**, **Replace** (upload or link, in a panel for that slot), **Fit to best size** (only when the image has a suggestion) and **Remove** (after a confirmation). With all three slots full, **Add a playmat** asks which one to replace. The **Darken** slider and the preview work on the active mat. The per-device `all / mine / off` choice is as before.
- **The best-size prompt** (`PlaymatFitPrompt.svelte`) opens right after an upload or a link whose result has a suggestion: the stored image with the kept area outlined and the rest dimmed; "This image is W x H. Playmats look best at 2400 x 1400 (the shape of a paper playmat)."; and **Fit to best size** and **Keep as is**. The kept window starts centred and can be **dragged along the axis being cropped** (the other axis has no room), or moved with the arrow keys once it has focus (Shift for bigger steps, Home and End for the ends); **Enter** fits and **Escape** keeps the image as it is. The window is a labelled slider with a position in words. *Keep as is* leaves the image exactly as stored, shown at cover-fit on the table, and the slot keeps its small **Fit to best size** action, so the decision is not one-shot.
- **Admin account view** (adapting ADR 0124's amendment and #2509): `GET /admin/users/{id}` serves `playmats: [{slot, url, active?}]` in place of `playmat_url`, and the view lists the person's saved mats as thumbnails, each with **Remove** and the same confirmation naming the person. **The admin route is both:** `DELETE /admin/users/{id}/playmats/{slot}` removes one (what the thumbnails call), and `DELETE /admin/users/{id}/playmat` stays as "remove all", so the route §9 documented, and anything that `curl`s it, keeps working. Both are `requireAdmin` and in the player-mode census.

#### 11.6 Not done

- A per-mat wash (11.1).
- Keeping the original upload, so a fit could work from more pixels than the stored 2560 px (11.4).
- A per-viewer `object-position` for the own-seat panel (11.3).
- Reordering slots. A slot is a place, not a rank; replace to move.

## Alternatives considered

- **Hotlink the pasted URL.** Rejected for the four reasons in decision 3.
- **Store the playmat in the settings blob (`PUT /me/settings`).** That blob is capped at 32 KiB, is read by the client's migrate and not by the server, and is per-person *preferences*, not a file other people must be able to fetch.
- **Put the URL on `game.Player` like `DiscordAvatarHash`.** It would need a snapshot field (and a schema version question), clone, a persist path and a mid-game mutation with undo semantics, to carry a value that is an account's, changes outside the game and must not be undoable. The room already has the pattern for exactly that (`is_host`).
- **A public, unauthenticated `/playmats/{id}` like bugstore's.** Nothing needs it, and a session check costs nothing.
- **Accept `http` links too.** Considered, since the SSRF guard does not depend on the scheme. Refused: it adds a network path in cleartext for a single convenience that a file upload already covers.
- **Keep the original bytes and strip metadata.** Writing an EXIF/XMP/ICC stripper is the parsing job where bugs live. Re-encoding through `image/jpeg` drops everything by construction and also fixes the size.
- **Animated WebP / GIF.** Not accepted. A moving background behind a game is a distraction, and decoding a GIF is another surface.
- **A playmat per game.** One per account, as with a physical mat: the point is that it is yours.

## Consequences

- A new package, `server/internal/playmat`, a migration, one dependency (`golang.org/x/image`), six routes, one wire field, one per-device setting, one Settings tab and one board layer.
- A signed-in person's art is stored on this server's disk: up to a few hundred KB each (2560 px at q85), one per account, deleted on replace and remove. A box with 200 accounts holds well under 100 MB.
- A table's first load fetches up to four extra images. They are `immutable`, so each is fetched once per device until it changes, and `off` / `mine` is one click away.
- The link fetch is the largest new attack surface. It is one function with one policy, `AddrAllowed`, a test table of the refused ranges, and tests against the real dialer for the refusals.
- Existing tables, the engine, the bots, the snapshot schema and every card are untouched. A client that does not know `playmat_url` ignores it.

## Delivery

One PR, [#2495](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2495): the package and its tests, migration 0011 and its test, the routes and their tests, the room stamp and its tests, the lobby binds (join, claim, reclaim, practice, restore) and their tests, the client (settings field, API calls, Settings tab, board layer, service worker, proxy) and its tests, and the docs.

## Not done

- A "remove playmat" button in the admin account view (decision 9). Follow-up issue.
- Cleanup of a playmat file when an account row is deleted. No such route exists today (decision 5).
- Cropping or repositioning the image in the browser. `cover`, centred, is the mat; an image that is the wrong shape is the owner's to edit before they upload it.
