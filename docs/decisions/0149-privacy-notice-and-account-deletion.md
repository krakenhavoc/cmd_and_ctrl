# ADR 0149 — Privacy notice and self-serve account deletion: a tombstone, not a delete; anonymised past games; a notice that cannot drift

**Status:** Accepted · 2026-10-13 · Backlog (public launch)
**Issue:** [#2983](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2983)
**Numbering:** every ADR file on `origin/develop` and on all eleven remote heads was listed; the highest number anywhere was **0148**, so this is **0149**. The claim comment on #2983 reserved it.
**Builds on:** [ADR 0051](0051-user-database.md) (users, identities, the revocation watermark), [ADR 0110](0110-remember-me.md) (settings, setups, decks), [ADR 0128](0128-playmats.md) (playmats), [ADR 0017](0017-bug-report-button.md) (bug reports and their artifacts), [ADR 0095](0095-deck-coverage-and-deck-requests.md) (deck requests), [ADR 0123](0123-monitoring-metrics-logs-dashboards-and-alerts.md) (log retention).

---

## Context

A public launch means storing strangers' data: a Discord id and name, decks, games, replays, bug reports. Two things were missing. A person had no way to read what is kept, why, for how long and with whom it is shared. And nobody had a way to take it back: the only removal was an admin's `revoke-sessions`, which ADR 0051 scopes to sessions and leaves every row.

Two constraints shape the answer. Sessions are stateless signed tokens (`auth.HMACAuthenticator`); the only thing that can revoke one is the per-user watermark in `users.sessions_invalid_before`, loaded into memory at boot. And a finished game belongs to everyone who played it: deleting one person must not delete the other three players' history.

## Decision

### 1. One route, the caller's own account

`DELETE /me/account` with `{"confirm": "delete my account"}`. It names no account: the user id comes from the session, so only the caller's own account can be deleted and an admin gets nothing new. The phrase is checked by the server as well as the client, so a stray script cannot delete an account by accident.

| Caller | Answer |
|---|---|
| no credential | 401 |
| guest seat, spectator link, admin token (no `user_id`) | 403, saying there is no account to delete |
| signed in, wrong or missing phrase | 400 |
| signed in, holds a seat at a table that has not ended | 409, naming the table |
| signed in | 204, session cookie cleared |

The route is rate limited per caller (one try every ten seconds, burst three).

### 2. A tombstone, not a delete

The `users` row is **not** deleted. It is scrubbed: display name `Deleted player`, no avatar, no last deck, no playmat, `admin_mode_at` zero, `last_seen_at` set to the deletion time, and the new `users.deleted_at` stamped (migration 0015). The id stays.

Two reasons, both load-bearing:

- **The watermark must outlive the account.** A session token issued before the deletion is revoked by `sessions_invalid_before`, which is read from this row at boot. Delete the row and a restart un-revokes every token the person ever held, for up to 30 days. The transaction stamps the watermark itself, and the lobby then calls the ordinary `RevokeAll` so this process's cache and open sockets follow at once.
- **Foreign keys.** `seats.user_id`, `games.created_by` and `invites.created_by` point at the row. The tombstone carries no identity, so leaving them pointing at it is anonymous, and the other players' history needs no rewrite.

Signing in with the same Discord account afterwards finds no identity row and mints a new, empty user. The old account is not resurrected.

### 3. What happens to each column

`internal/accountdelete` runs this in one SQL transaction, all or nothing. `accountdelete.Inventory` lists every column in the schema with its treatment; that list is the source of truth and a test holds it to the live schema (§6).

| Data | Treatment |
|---|---|
| `identities` (Discord id, name, avatar hash, encrypted refresh token) | **deleted** |
| `decks`, `user_settings`, `table_setups`, `user_playmats` rows | **deleted** |
| playmat image files, the cached Discord avatar directory | **deleted**, after the commit, best effort, logged |
| `deck_request_asks` rows keyed by the person's Discord id | **deleted** |
| `seats.guest_name` of every seat the person held | **anonymised** to `Deleted player` |
| `seats.pending_discord_id`, `games.host_discord_id` equal to the person's Discord id | **cleared** |
| `seats.deck_id` | cleared by the existing `ON DELETE SET NULL` |
| `seats.user_id`, `games.created_by`, `invites.created_by` | kept, pointing at the tombstone |
| `seats.deck_name`, the games' names, clocks, winner and outcome | kept: part of the record the other players keep |
| `users` row | **tombstone** (§2) |

A Discord seat claimed before any `users` row existed (`pending_discord_id`) is found by the person's Discord id and treated as theirs.

### 4. The active-table rule: refuse

The deletion is refused while the person holds a seat at a table that is `lobby` or `active`. The alternatives were to concede the seat or to remove it, and both change a game the other players are in the middle of, for a person who may have deleted their account by mistake. A refusal costs one message ("Leave the table or wait for it to end") and is trivially safe. The check runs twice: in the lobby, where `syncStateLocked` first catches up a table that ended over the WebSocket and has not yet been written to its row, and in the transaction, for tables that are not resident. A practice table is neither: it has no rows, is never listed, and is reaped on its own.

A join that lands between the check and the commit is possible and harmless: the person's sessions are revoked by the same request, so the seat is a ghost on a table that cannot reach it, and a host can remove it.

### 5. Past games: the live game, the files, the restart

The seats keep their place and their result. What changes is the name and the Discord metadata, in every place a past game keeps them:

- **The lobby's copy and the live room** of a table that is still resident. The room's game is changed through `Room.ApplyExternalClearingUndo` (`Game.AnonymisePlayer`), so the commit rewrites the crash dump and restore point from the anonymised game, drops the undo ring (its clones still hold the name) and reaches anyone looking at the ended table.
- **The files on disk**, whether or not the table is resident: `games/<id>.json` (last view), `replays/<id>.jsonl` (one full view per commit) and `restore/<id>.json` (engine snapshot, kept `ws.EndedTableRetention`, seven days). `ws.RoomManager.ScrubPlayerFiles` rewrites each line structurally: any object whose `id`, `seat`, `target_id` or `source_id` is one of the person's player ids loses its name, Discord id, avatar hash, display name and playmat fields. The game log in these files is structured (seat indexes and card ids), not rendered text, so a name appears only on player objects and on a few annotations that repeat it, and no free-text search is needed. A file with nothing to scrub is not rewritten; a rewrite is atomic and keeps the file mode and its trailing newline.

A restart restores the table from the scrubbed file with the same two seats. The replay still reads and still downloads.

### 6. The notice, and the rule that keeps it true

The privacy page is a public client route, `#/privacy` (`routes/Privacy.svelte`), linked from the site footer, the sign-in screen, the account menu and the table's legal notice (which opens it in a new tab so a game in progress is not left). Every statement on it is derived from the code; the PR cites the source of each, and each row of the page carries it in its source comment.

A page that was true on the day it was written is not enough. Three tests hold it:

- `accountdelete.TestInventoryCoversTheSchema` fails when a migration adds a table or column that `Inventory` has no deletion decision for.
- `lobby.TestPrivacyNoticeNumbersMatchTheServer` reads `client/src/lib/privacy.ts` (the numbers the page states) and fails when any drifts from its source constant: the ended-table retention, the bug-artifact retention and size ceiling, the sign-in and guest session lifetimes, the deck-request limit, the playmat slots, the backup schedule, the log retention.
- `lobby.TestPrivacyNoticeIsCurrentWithTheSchema` fails when a migration newer than `PRIVACY_SCHEMA_VERSION` exists. Every new migration therefore costs one person one read of the page and a one-line bump. That is the point, and it is cheap; the failure it prevents is a table of personal data the page does not mention.

The rule for contributors: **a change that stores anything new about a person updates `Inventory`, `Delete` and the page in the same change.** It is also in `AGENTS.md`.

### 7. What cannot be deleted, and why the page says so

- **GitHub issues** (bug reports, ideas, questions, deck requests) carry the filer's display name and are public. The server files them with its own token; the person cannot edit them and we cannot unpublish a public issue's history. The page says so, and says the owner will edit or remove one on request.
- **Bug-report artifacts** (screenshots, the pinned game log and replay) are not linked to an account: the manifest records the game id and the issue, never the reporter. They cannot be found by user, so they are left, and they expire on their own (`bugstore.DefaultRetention`, ninety days). The alternative, adding a reporter id to the manifest so deletion could find them, would store more personal data in order to be able to delete it, and the artifacts are short-lived and admin-only except the screenshots the reporter chose to publish.
- **Backups** (restic, 7 daily, 4 weekly, 6 monthly) roll off on their schedule; deleting an account does not rewrite history.
- **Logs** hold internal ids and remote addresses for thirty days (ADR 0123). The audit line below carries the user id and counts only.
- **This browser's own storage** (the session, the settings cache) is the person's; signing out clears the session, and their browser's site-data controls clear the rest.

### 8. Audit

One log line, `account deleted`, with our own user id and counts of what was removed (`decks`, `playmats`, `seats_anonymised`, `deck_request_asks`, `settings_rows`, `setup_rows`, `identities`, `sockets_closed`, `games_scrubbed`). It never carries a name or a Discord id, because the log outlives the account.

### 9. Contact

No contact address has been chosen, and none is invented. The page says "Contact: see the project's GitHub" and links the repository. `TODO(owner): contact` marks the constant in `client/src/lib/privacy.ts` to replace.

## Consequences

- A person can read what is kept and remove it. The other players' games survive intact, with one seat renamed.
- A deleted account's sessions stay dead across a restart, which a row delete would not have achieved.
- The page cannot silently go stale on a retention change, a new column or a new migration; the cost is a one-line bump per migration.
- The replay scrub is structural. A new field that repeats a player's name outside a player object (a new annotation, for instance) needs adding to `scrubPlayers`; `TestDeletedPlayersPastGameStillRestoresAndRenders` fails if the name survives in the files a real game writes.
- `users.deleted_at` is the only schema change. Admin views show a deleted account as "Deleted player".

## Back-out

| To undo | Do |
|---|---|
| the route | remove the `DELETE /me/account` registration in `lobby/http.go`; the client menu item then answers 404 and `DeleteAccount.svelte` can be dropped |
| the column | `ALTER TABLE users DROP COLUMN deleted_at` by hand after rolling the binary back (an older binary refuses schema 15, `ErrSchemaTooNew`); tombstoned rows then read as live accounts named "Deleted player" with no identity |
| the page | remove the route, the footer and sign-in links; the numbers and tests in `lobby/privacy_facts_test.go` go with it |

An account that was already deleted is not recoverable by any of these.
