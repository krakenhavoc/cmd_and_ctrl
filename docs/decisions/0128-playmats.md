# ADR 0128 — Playmats

**Status:** Accepted (owner answers 2026-10-07) · 2026-10-07 · No numbered sprint: table UX outside a sprint
**Issues:** [#2503](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2503) (this change).
**Owner decisions:** three answers of 2026-10-07, quoted under [Owner decisions](#owner-decisions-2026-10-07). They are binding.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-07, over every remote head (`git log --remotes --name-only -- docs/decisions`). The highest number on any of them is 0127 (`0127-answering-repeated-prompts-for-you.md`). This ADR takes **0128**. The same sweep over `server/internal/db/migrations` found 0010 as the highest migration, so this change adds **0011**.
**Builds on:** [ADR 0051](0051-user-database.md) (users and the database), [ADR 0110](0110-remember-me.md) §4 (settings on the account), [ADR 0017](0017-bug-report-button.md) §6 (image uploads, whose checks this reuses), [ADR 0041](0041-game-persistence.md) (the snapshot fields a new `Player` field needs).

---

## Context

A player wants to put their own image, a playmat, behind their part of the board, and wants everyone at the table to see it.

Nothing like it exists. Avatars are not uploads: the server caches each player's Discord picture (`internal/discord/avatar_cache.go`). A signed-in person's settings are on their account (ADR 0110 §4), but they are private: `GET /me/settings` serves them only to their owner. Nothing about a person reaches the other players except their name and avatar, which ride `game.Player` into `PlayerView`.

The one upload the server accepts is a bug report's screenshot (`internal/bugstore`). It sniffs the type from the bytes, allows PNG, JPEG, GIF and WebP, refuses SVG, and caps the size.

### Owner decisions (2026-10-07)

1. **Upload.** A playmat is an image file the player uploads, stored on our server and served from our origin. Not a pasted URL: every viewer's browser would fetch it from the third party (their IP and referrer go with it), and the link could change under a live table.
2. **Saved to the account.** A signed-in person sets it once and it follows them to every table. (Guests cannot keep one.)
3. **Dimmed and cropped to fill by default, with an adjustable wash.** "Standard is the dimmed, cover-fit, but allow an adjustment for tint/translucence/dark wash."

## Decision

### 1. One row and one file per person

Migration 0011 adds `user_playmats (user_id, file, wash, updated_at)`. `file` is the image's name; `wash` is the dark wash in percent.

The image lives at `$CMDCTRL_DATA_DIR/playmats/<32 hex>.<ext>` (`internal/playmats.Files`). Each upload gets a **new random name**, and the old file is deleted once the new row is written. A name is never reused, so the image can be cached for good (`immutable`) and a replaced playmat can never show through a stale cache.

An upload is checked as a bug report's screenshot is: the type is sniffed from the bytes, only PNG, JPEG, GIF and WebP are kept, and the cap is 4 MiB. SVG is refused, because it is a document that can run script and these images reach every player. Nothing is decoded or resized: the browser scales the image, and the cap bounds the cost.

### 2. The wash belongs to the playmat

The wash is part of the owner's playmat, so every viewer sees it at the strength the owner chose. Its range is 30 to 90 percent, default 60. The floor is what keeps cards, labels and counters readable over busy art, which matters more on someone else's board than on your own.

A viewer may still turn **other players'** playmats off (`display.showPlaymats`, synced). Their own always shows.

### 3. The playmat rides the seat

The image is a fact about a person; the table needs it on a seat. So it is copied onto `game.Player` (`PlaymatPath`, `PlaymatWash`) and projected as `PlayerView.playmat {path, wash}`, public and unredacted, as the Discord avatar is.

- **When a person sits down** (`join`, `LinkSeat`, `CreatePractice`), the lobby reads their playmat before it takes its lock and sets it in the same commit that seats them, so the first broadcast carries it. A reclaim by a guest clears a playmat the seat's previous owner left.
- **When a person changes it** (`PUT`, `PATCH`, `DELETE /me/playmat`), `Lobby.ApplyPlaymat` sets it on every seat they hold at a table that is not archived, each as an ordinary state broadcast.
- A failed read is treated as no playmat. It is cosmetic and must never keep anyone out of a seat.

The two fields are additive snapshot fields (recorded in `snapshot_shape/v7.txt`, classified `carried`), so a restored table keeps each seat's playmat with no bump. A table restored after its owner replaced the image points at a deleted file until the owner next sits down or changes it; the client then draws the plain surface, because the image is the bottom layer over the surface colour.

### 4. Routes

`GET`, `PUT`, `PATCH` and `DELETE /me/playmat` for the owner (signed in, like the rest of `/me/*`), and `GET /playmats/{file}` for any session, under the avatar rate limit, with `?token=` as for avatars. The full contract is in [docs/lobby.md](../lobby.md#meplaymat-and-get-playmatsfile-adr-0128). `/playmats/*` is added to the Caddyfile's `@api` matcher, the Vite proxy and the service worker's API deny list, all three of which must name every server path.

### 5. The board

`PlayerPanel` sets `has-playmat`, `--playmat` and `--playmat-wash` on the seat's panel, and `app.css` draws, top to bottom: the seat colour from the corner, a flat wash of the page background at `--playmat-wash`, the image cropped to cover, and the surface colour. The client builds the CSS `url()` only from a path matching exactly what the server writes (`/playmats/<32 hex>.<ext>`), since it is interpolated into CSS.

The summary read-out (`SeatSummary`) does not draw a playmat: it is a dense panel, not a board.

## Consequences

- Each signed-in person stores at most one image, up to 4 MiB, and the old one is deleted on replace or remove. The images are in the off-site backup (it skips only caches).
- Every viewer of a table loads each seat's image once per name; the `immutable` cache makes later frames free.
- A guest cannot set a playmat. If guests ask, a per-game playmat on the seat (not the account) can be added without changing anything here.
- Moderation is out of scope: this is a private table among people who invite each other (PLAN.md §2). A player who dislikes a playmat can turn other players' playmats off.

## Delivery

One PR: the migration and `internal/playmats`, the `game.Player` and `PlayerView` fields, the lobby wiring and routes, the proxy and service-worker entries, and the client (Settings › Display › Playmat, the board's layer, `display.showPlaymats`).
