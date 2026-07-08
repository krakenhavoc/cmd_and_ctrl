# ADR 0017 — In-app bug reports as GitHub issues

**Status:** Implemented · 2026-07-08 · Branch `feat/bug-report-button`

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
