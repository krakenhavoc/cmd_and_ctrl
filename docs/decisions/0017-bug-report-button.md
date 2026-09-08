# ADR 0017 — In-app bug reports as GitHub issues

**Status:** Implemented · 2026-07-08 · Branch `feat/bug-report-button`
**Revised:** 2026-09-08 · §4 amended, §6–§7 added · Branch
`feat/bugreport-logs-attachments`

## Context

Playtest bugs currently travel by Discord message ("the cast dialog
ate my X again") and evaporate. By the time anyone looks, the game
is over, nobody remembers the turn, and the replay that could prove
what happened is sitting unexamined on the VPS. The project already
tracks all work as GitHub issues, so the tracker is where reports
should land — with enough context attached that the operator can
pull the right replay without an interview.

The constraint that shapes everything else: filing a GitHub issue
requires a token with write access to the repo, and the repo is
private. Whatever design we pick must keep that token off the
browser.

## Decisions

### 1. Server-proxied filing; the token never reaches the client

The browser talks only to `POST /bugreport`. The server holds a
fine-grained PAT (`CMDCTRL_GITHUB_TOKEN`, **Issues: write** on the
one repo, nothing broader) and renders + files the issue itself.
The token lives in the environment, is held in memory only, and
never appears in logs, error strings, or responses (`github.Client`
truncates upstream error bodies and never echoes headers).

A leaked token from this design can file issues on one repo —
annoying, revocable, nothing more. Compare the bot's
`CMDCTRL_ADMIN_TOKEN` (full admin), which is why the PAT gets the
same separate-env-file treatment on the VPS.

### 2. A `BugReporter` seam in `lobby`, nil = disabled

`lobby` declares the one-method interface it needs; `*github.Client`
satisfies it; `main.go` wires them. Identical shape to `GameEvictor`
and ADR 0016's `StateBroadcaster`: no import cycle, and handler
tests inject a recording fake instead of standing up GitHub.

Nil follows the house disabled-feature convention: `POST /bugreport`
returns 503, and `GET /bugreport/config` reports `enabled:false` so
the client never renders the button (mirrors
`/auth/discord/config`). An unconfigured deploy shows no dead UI.

### 3. A purpose-built `internal/github` package, not an SDK

The server needs exactly one GitHub call. A third-party client would
import a dependency tree for one POST. The package follows the
S06.5 outbound-HTTP posture: stdlib only, 10 s timeout, descriptive
User-Agent, capped response reads.

### 4. Context is coordinates, never state

Reports attach: game ID, turn number, phase/step, seq watermark,
connection status — and the server adds reporter identity, server
time, and User-Agent. Deliberately excluded: hands, libraries,
decklists, any `GameView` content. Embedding state in an issue would
route hidden information around the wire visibility filter (the
S04 invariant), and it's redundant — the issue links
`GET /games/{id}/replay`, which is the authoritative record anyway.

Client-supplied context strings are clipped to 64 bytes and
newline-stripped server-side, so a crafted value can't forge extra
metadata rows in the issue body.

### 5. Tight rate limit, outermost

Every accepted report is an outbound write to the shared tracker, so
the limiter wraps *outside* auth: burst 3, then ~1 report / 30 s per
client IP. One stuck retry loop can't wallpaper the tracker. The
limiter honours `CMDCTRL_DEV_RELAX_RATE_LIMITS` like every other
lobby bucket.

## Alternatives considered

- **Prefilled `github.com/.../issues/new?title=…` link.** No server
  work, but the repo is private — most of the playgroup can't open
  issues at all — and auto-context would be limited to what fits in
  a URL.
- **Discord webhook.** Matches where the playgroup lives, but
  reports become chat scroll again — no dedupe, no labels, no closed
  state. The tracker is the system of record.
- **Filing through the bot binary.** The bot is Discord-scoped and
  holds the *admin* token; giving it GitHub credentials too widens
  the blast radius of a bot compromise. The lobby already has
  sessions, rate limiting, and JSON hygiene.

## Consequences

- The PAT is provisioned by the CD job, not by hand: the deploy
  upserts it into `/etc/cmd_and_ctrl/env` from the
  `CMDCTRL_GITHUB_TOKEN` Actions secret via
  `sudo scripts/set-server-env.sh` (stdin transport, atomic replace,
  owner/mode preserved). That path — not the `server.env` name ADR
  0004 §6 sketched — is what the HomeLab cloud-init template
  actually writes and what the systemd units read via
  `EnvironmentFile=`; the file is `root:cmdctrl` `0640` and the
  deploy user's cloud-init sudoers grant covers the write. Rotation
  is "update the secret, rerun the deploy". Note the self-hosted
  runner already holds SSH access to the host, so the secret adds no
  new trust boundary beyond the PAT itself.
- In-app issues arrive titled `[in-app] …` with the `bug` label —
  distinguishable from hand-written ones at a glance.
- The operator's loop is: read issue → `GET /games/{id}/replay` with
  the ID in the metadata block → reproduce.
- Reports require a session (any role — spectators hit bugs too);
  drive-by spam needs an invite token first, and the rate limit caps
  what a hijacked session can file.
- If GitHub is down, reports fail with a 502 and the modal says try
  again / tell the admin — reports are not queued server-side. At
  this scale, a lost report during a GitHub outage is acceptable;
  a persistence queue is not worth its failure modes.

---

## 6. Screenshots (added 2026-09-08)

Reports carry up to 4 images, ≤4 MiB each and ≤10 MiB per report.

**The constraint:** GitHub's REST API has no attachment-upload
endpoint. The web uploader that produces
`user-images.githubusercontent.com` URLs is not exposed as an API, and
a private repo's `raw.githubusercontent.com` needs a token — so an
image committed to the repo would not render either. The only shape
that renders in an issue is a plain markdown image pointing at a URL
GitHub's image proxy (Camo) can fetch anonymously.

So the server hosts them: `internal/bugstore` writes to
`$CMDCTRL_DATA_DIR/bugreports/<report-id>/att-N.<ext>` and the issue
body embeds `<CMDCTRL_PUBLIC_BASE_URL>/bugreport/att/<report-id>/att-N.png`.

That forces **one unauthenticated route**, which is the largest new
attack surface in this feature and gets the compensating controls to
match:

- **The report ID is the capability.** A v4 uuid, 122 bits, appearing
  only inside a private-repo issue body. There is no listing endpoint,
  and a wrong id is indistinguishable from a pruned one (both 404), so
  the route can't be used to probe which reports exist.
- **Magic bytes, not the declared type.** `http.DetectContentType`
  must return one of png/jpeg/gif/webp, on write *and* on read. A part
  labelled `image/png` carrying HTML is refused — that is precisely the
  trick a same-origin unauthenticated media route has to survive.
- **SVG is refused outright.** It is a script carrier; serving
  attacker-supplied SVG from our own origin is stored XSS.
- **Inert responses.** `nosniff`, an explicit allowlisted
  `Content-Type`, `Content-Disposition: inline`, and
  `default-src 'none'; sandbox` — so even a file that somehow got past
  the sniffing does nothing when navigated to directly.
- **Path shape is validated, not sanitised.** The id must match the
  uuid pattern and the name must match `att-N.<ext>`; nothing else
  reaches `filepath.Join`.
- **90-day retention and a 512 MiB ceiling**, both enforced at boot and
  after each report, oldest-first. Retention alone is not a disk
  guarantee: §5's limiter permits ~1 report / 30 s per IP, which inside
  a 90-day window is tens of GiB. The data dir also holds a ~630 MiB
  Scryfall dump and the replay logs, and filling it takes the game
  server down — much worse than losing an old report's screenshots.

`CMDCTRL_PUBLIC_BASE_URL` has no default: guessing the origin wrong
produces issues full of broken images, which is worse than a deploy
where the modal simply doesn't offer upload. Unset (or no data dir)
leaves `attachments:false` in `/bugreport/config` and text reports
working exactly as before.

**Rejected:** committing images to a branch (pollutes history, and
private-repo raw URLs don't render anyway); base64 data URIs (GitHub
markdown doesn't render them); a Discord webhook as the image host
(moves the artifact away from the tracker again).

## 7. Logs: §4 amended (added 2026-09-08)

§4 said "context is coordinates, never state", and pointed the
operator at `GET /games/{id}/replay`. In practice that reference was
close to useless: the replay is gated until the game ends for
non-admins, the file keeps being appended to after the report is
filed, and game eviction deletes it outright — so by the time anyone
read the issue, the log it named was usually gone or unrecognisable.

§4's *reason* was sound and is preserved. Two distinct things were
conflated under "state":

**The client log — inlined.** `GameClient.log` already ringed the last
200 protocol events (actions sent, snapshot seq/turn/step, socket
drops, reconnect backoff, server error frames); this revision raises it
to a named `WS_LOG_LIMIT`, adds JS-level capture
(`lib/clientErrors.ts`: uncaught errors, unhandled rejections,
`console.error`/`warn`), and merges the two into one timeline. It is
inlined in the issue inside a collapsed `<details>` block, because by
construction it contains **only what that browser already had** — no
`GameView` content, and no information the reporter could not already
see. The interleaving is the diagnostic value: "action sent, then a
TypeError, then no snapshot" is a diagnosis; the same three facts in
separate places are three facts.

**The replay — pinned, never inlined.** On report the server copies
the game's replay JSONL to `bugreports/<report-id>/replay.jsonl` —
a snapshot, so it survives both further play and eviction. The issue
carries only the report ID; reading it takes
`GET /bugreport/<id>/replay`, **admin only, with no
game-has-ended relaxation** (a pinned replay has no live game to
reason about, so the strict rule is the only safe one). Over 32 MiB
the tail is kept and the issue says so, because an operator's first
assumption is that a replay starts at turn one.

That split is what keeps §4's invariant true where it matters: a
reporter who is also a repo collaborator gains nothing by filing a
mid-game bug. They see their own client's log — which they already had
— and a report ID they cannot dereference.

Two smaller consequences:

- Pinning is entitlement-checked (admins anywhere, everyone else only
  for the game their session is bound to), mirroring
  `downloadReplay`. Not because the contents would leak — they are
  admin-gated — but so a session in game A can't make the server copy
  game B's log.
- Log lines carry chat text, which other players type. They land
  inside a fenced code block, so backticks are replaced and control
  characters stripped: an unescaped fence would end the block early
  and let the rest render as markdown, which is how a "log line"
  forges a metadata row.

**On consent.** Both automatic attachments are opt-out and both are
shown in full in the modal before sending — the log behind a "show"
toggle. A report that quietly ships a log is a report people stop
filing the day they notice.
