# ADR 0017 — In-app bug reports as GitHub issues

**Status:** Implemented · 2026-07-08 · Branch `feat/bug-report-button`
**Revised:** 2026-09-08 · §4 amended, §6–§7 added · Branch
`feat/bugreport-logs-attachments`
**Revised:** 2026-09-14 · §8 added (report kinds → labels) · Branch
`feat/inapp-report-kinds`
**Revised:** 2026-09-16 · §9 added (credentials redacted before a
report is published, #721; blocks #517) · Branch
`fix/redact-tokens-721`

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
  distinguishable from hand-written ones at a glance. (§8 keeps the
  prefix and makes the label depend on the report's kind.)
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

---

## 8. Report kinds (added 2026-09-14)

The button files bugs. It is also the only place a player can say
anything to the tracker, so it collects feature requests and questions
as bug reports, which then have to be retagged by hand. The fix is a
**kind** on the report, applied as a GitHub label at filing time.

### 8.1 An explicit picker, not inferred classification

The kind is a radio group at the top of the modal, not something
derived from the prose. Classification would mean a model call — or a
keyword heuristic — in front of somebody who is mid-game and annoyed,
and it would sometimes be wrong in a way the reporter can't see or
correct. A three-way pick is one tap, is never wrong, and adds nothing
to the latency of the thing it gates. The picker is *first* in the
form because it changes the wording below it and what the report
attaches.

### 8.2 Three kinds, mapped to labels that already exist

| kind | label | the player's framing |
|---|---|---|
| `bug` | `bug` | something's broken |
| `idea` | `enhancement` | something's missing |
| `question` | `question` | I don't understand something |

`bug` and `enhancement` are the two the feature obviously needs.
`question` earns the third slot because the alternative is worse:
without it, "why did my commander go to the graveyard?" is filed as a
bug, and a rules question that turns out to be correct behaviour
closes as `invalid` — which is a discouraging thing to do to the one
person who bothered to report anything. It maps to the existing
`question` label; nothing new is invented, and the mapping is
**server-side**. The client names a kind, never a label: a
client-supplied label string would let any session with a report
button attach `good first issue` — or make GitHub create labels — on
the tracker.

An omitted kind means `bug`. Every client built before the field
existed sends nothing, and those are bug reports; a 400 there would
break the button for an open tab mid-game, which is exactly when it
gets used. A non-empty *unknown* kind is a 400, matching this
endpoint's existing refusal of unknown JSON fields — a kind quietly
rewritten to `bug` is a mislabelled issue nobody learns about.

### 8.3 Artifacts follow the kind

§6 and §7 attach screenshots, the client log, a pinned replay and a
pinned game log. Collecting all four for "the stack panel should be
wider" is waste — tens of MiB of replay, retained 90 days, that
nobody will ever open.

| kind | screenshots | client log | pinned replay | pinned game log |
|---|---|---|---|---|
| `bug` | yes | yes | yes | yes |
| `idea` | yes | no | no | no |
| `question` | yes | yes | no | yes |

A question sits between the two on purpose: "why did that die?" is
answered from what **happened** — the public game log, a few KiB —
and from what the reporter's client did. It is not answered
frame-by-frame from hidden state, which is what the replay is for.

The split §7 established is untouched by any of this. Whatever a kind
does attach keeps its posture: screenshots public because Camo has to
reach them, pins admin-only with no game-has-ended relaxation, and no
`GameView` content in an issue body ever. The gating happens *before*
the artifact is read, so an idea never causes the server to copy a
replay at all. The modal states what the chosen kind will send, and
the log stays behind its own opt-out and "show" toggle for the kinds
that carry it — §7's consent rule does not get weaker because there
are now three doors into the same form.

### 8.4 The title prefix stays `[in-app] `

Rejected: `[in-app bug] ` / `[in-app idea] `. Every issue this feature
has filed carries the current string, and it is what people filter on;
widening it orphans that convention on ~20 open issues, and it makes
the kind a second source of truth in a field triage cannot fix with a
click — retag an issue and the title still says `idea`. The label is
the kind. The issue body also carries a `Kind:` row, which costs one
line and means a stripped (or never-applied) label doesn't erase what
the reporter said they were filing.

### 8.5 A label may never cost the report

A label that doesn't exist, or a token that may not apply it, must not
turn a report into a 502. When GitHub **rejects** the create (403 or
422 — the issue was definitively not created), the server re-files it
unlabelled and logs at error level; the 201 then carries no `label`,
so the modal doesn't claim one. An unlabelled issue costs one click in
triage. A report that evaporated because someone renamed a label costs
the report, and nobody ever finds out.

Delivery failures are deliberately *not* retried. A timeout or a 5xx
may mean GitHub created the issue and lost the response; a duplicate
issue is worse than the 502 the reporter can act on. That distinction
is why `github.CreateIssue` now returns a typed `*APIError` carrying
the status — the decision is made from the status, not by grepping an
error string.

---

## 9. Secrets are redacted before a report is published (added 2026-09-16)

§7 called the client log safe to inline because it holds "only what
that browser already had". That browser also had its session token.
`GameClient` logged `connected to <ws URL>`, and the WebSocket URL
carries `?token=` (browsers cannot set headers on an upgrade), so
every report with a connect line published the reporter's session
token to everyone who can read the repo (#721). Tokens die with the
server process today; **durable sessions (#517) would keep a pasted
token valid until it expires, so #517 is blocked on this.**

§7 is amended: the log is inlined **after credentials are redacted**,
and so is every other client- or player-supplied string in the issue.

- **One rule set, two copies.** `client/src/lib/redact.ts` and
  `server/internal/util/redact` replace the value (never the key) of
  any `*token` / `*secret` / `*password` / `*ticket` key=value pair,
  the invite and reclaim `?t=` as a query parameter, JSON
  `"token":"…"` fields, and `Bearer <credential>`; literal, URL-encoded
  (`%3D`) and double-encoded (`%253D`) forms. The value becomes `REDACTED`, so triage
  still sees that a token was sent. Keep the two in step.
- **Client, at the source.** `ws.ts` logs `connectLogLine(url)`, which
  keeps host, path, `game` and `player` and redacts the token.
- **Client, as a backstop.** Every protocol-log entry, every captured
  console/JS error, the merged report log (`collectBugLog`) and the
  submitted title, description and context are redacted. A log line
  added later cannot reintroduce the leak through the report.
- **Server, for every other client.** `renderBugIssueBody` redacts its
  input (`redactBugIssue`) before rendering and before clipping, so a
  clip can never cut a key off and leave its value. That input includes
  every reporter display name the footer can fall back to — seat name,
  Discord global name, Discord username, all self-chosen — and the
  rendered name is clipped like the other footer values. The handler
  also redacts the issue title and the client-supplied game ID written
  to the bugstore manifest. An old tab or a curl gets the same
  treatment as the current client.

**What is not redacted, and why.** The pinned replay and game log (§7)
are server-generated projections of game state: they carry no session
or invite tokens (chat, the one player-typed text on the wire, bypasses
`Room.Apply` and never reaches the replay), so their format is
unchanged. Screenshots (§6) are images and cannot be filtered: a
reporter who captures an address bar holding an invite link publishes
it. The server never logs a request URL or query string, so
`?token=` does not reach the server log either.

Guards: `ws.redact.test.ts` drives a real `GameClient` connect and
fails on any `token=` in the log that is not `REDACTED`;
`bugreport_redact_test.go` fails if `renderBugIssueBody`, or the whole
handler, lets a token from the client log, description, context,
reporter name or title through, or into the stored manifest. It seeds
every string field of the issue, principal, context and log entry by
reflection, so a field added later is covered unless it is explicitly
exempted as server-issued.
