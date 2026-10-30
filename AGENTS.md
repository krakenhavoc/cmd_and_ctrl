# AGENTS.md — working on cmd_and_ctrl

This file is for AI coding agents (Claude Code, Codex, Cursor, etc.) operating in
this repository. Humans should read [PLAN.md](PLAN.md), the charter, first. This
file holds the rules and the links; the reference is in `docs/`, indexed in §3.

---

## 1. What this project is

A free, unofficial 4-player Magic: The Gathering Commander table with
a Go game server, an automation-first rules engine, bots and a TypeScript
client, in public beta for the C2 Discord community. See [PLAN.md](PLAN.md),
the charter, for who it is for, scope, principles and the beta goals.

Three hard problems, ranked: **game state and multiplayer sync** (Go
server, authoritative state, WebSocket broadcast), **UX polish** (the
whole point — Commander-specific affordances nothing else has), and —
long horizon — **incremental rules enforcement** (B→C track, starting
in S13+ after the sandbox is shipped).

**Naming.** The project is *cmd & ctrl* (the repository is `cmd_and_ctrl`,
the site's title is *CMD & CTRL*). Its one abbreviation is **C2**: use it
wherever a short name is wanted (prose, issues, PRs, chat, the Discord
bot's `/c2-…` commands, the installed app's `short_name`). Don't coin
others (CC, C&C, CMD&CTRL). Machine identifiers keep their existing
`cmdctrl` / `CMDCTRL_` spelling (env vars, Docker volumes, metric and
service names); don't rename those to match.

---

## 2. Ground rules

1. **Read [PLAN.md](PLAN.md) before making architectural suggestions.** The
   central decision (reuse XMage vs build from scratch) is already made. Do not
   relitigate it without new information.
2. **This is a hobby project, not a company.** Optimise for momentum and
   clarity, not enterprise rigour. No microservices, no k8s, no premature
   abstractions. One VPS. One database. One deployable per service.
3. **Build for the beta in [PLAN.md](PLAN.md), not beyond it.** C2 is a free,
   non-commercial public beta for a small community: tens of players on one VM,
   held to the capacity caps of [ADR 0148](docs/decisions/0148-capacity-limits-for-a-public-launch.md);
   grow it on measured load, not on a hypothetical. It is fan content within the
   Wizards Fan Content Policy: never charge for anything, keep guest play, and
   keep card data out from behind any sign-in or Discord gate (PLAN.md §3).
   Strangers' data and safety matter: privacy ([ADR 0149](docs/decisions/0149-privacy-notice-and-account-deletion.md)),
   moderation and limits are product work. Don't build for a scale or audience
   the charter doesn't name.
4. **UX polish is the entire point.** "Make it feel good" is core product
   work and deserves real care. Automation comes first; the sandbox actions
   (manual resolution, manual priority) stay as the fallback for whatever
   the engine doesn't know yet.
5. **We do not use XMage.** This was evaluated and rejected in S01 — see
   [PLAN.md, History](PLAN.md#history). The entire backend is an original Go
   game server; the rules engine is grown incrementally on top of it. No Java
   anywhere in the stack.

---

## 3. Repo layout and reference

```
cmd_and_ctrl/
├── PLAN.md        # the charter: what C2 is, who the beta is for, scope, goals
├── AGENTS.md      # this file
├── DESIGN.md      # the UI design brief
├── Makefile       # top-level dev/test/lint targets
├── server/        # Go game server (authoritative state, WebSocket + HTTP API)
├── client/        # TypeScript + Svelte 5 + Vite
├── scripts/       # CD helpers, cron and backup jobs, one-off tools
├── deploy/        # monitoring alert rules and Grafana dashboards
├── data/          # runtime state (gitignored)
└── docs/          # reference, runbooks and ADRs
```

When you create a new top-level directory, add it here.

Where things are:

- [docs/repo-layout.md](docs/repo-layout.md): every package under `server/` and `client/`.
- [docs/commands.md](docs/commands.md): make targets, server and client commands, scheduled jobs.
- [docs/env.md](docs/env.md): every environment variable the server, the bot and the bot seat read.
- [docs/lobby.md](docs/lobby.md#endpoint-index): the HTTP endpoints. [docs/protocol.md](docs/protocol.md) is the WebSocket wire format.
- [docs/bot.md](docs/bot.md) and [docs/bot-testing.md](docs/bot-testing.md): the AI bot seat and its test gates. [docs/discord-bot.md](docs/discord-bot.md): the Discord bot.
- [docs/snapshots.md](docs/snapshots.md): restore points and the fixture corpus.
- [docs/rules-citations.md](docs/rules-citations.md): citing the Comprehensive Rules.
- [docs/decisions/README.md](docs/decisions/README.md): the ADR index and how to number a new one.
- [docs/sprints.md](docs/sprints.md): the sprint plan. [docs/environments.md](docs/environments.md): environments and runbooks.

---

## 4. Sprint and tracking discipline

### Branches and environments

`develop` is the integration branch; `main` is the release branch. Feature PRs
target **`develop`**, which auto-deploys to `https://cmd-dev.labxp.io`. Always pass
`--base develop` to `gh pr create`: `main` is still the default branch, and a
feature PR into `main` fails the required `promotion-guard` check. Promotion is a
`develop` → `main` PR. Matrix and runbook: [docs/environments.md](docs/environments.md);
[ADR 0023](docs/decisions/0023-develop-environment.md).

Two rules that are easy to get wrong:

- **Dev-only features are gated server-side.** `CMDCTRL_ENV=dev` is the only
  switch, and `appenv.LoadFeatures` refuses to enable one outside a dev
  deployment. Wrap every dev-only route in `requireDev`. A client-side `if` is not
  gating it: spawning cards and acting as another seat are cheats on a live table.
- **Never deploy or restart the Discord bot from `develop`.** Two bot processes on
  one application answer every slash command twice. CD enforces it.

Work is organised into 2-week sprints, tracked on the
[project board](https://github.com/users/krakenhavoc/projects/5), the
[milestones](https://github.com/krakenhavoc/cmd_and_ctrl/milestones) and the
[sprint plan](docs/sprints.md). Each sprint has one tracking issue whose body is its
checklist; the plan lists them.

**Every commit and pull request must reference the sprint and issue it relates
to.** This is non-negotiable. If a PR does not belong to the active sprint, say so
and justify it.

### Picking an ADR number

Two branches can both take the next free ADR number, and git merges the different
filenames without a conflict. Before you add one, check every remote branch's
`docs/decisions/` (the command is in [docs/decisions/README.md](docs/decisions/README.md)),
take the first number none of them uses, and name the branches you checked in the
PR body. Numbers are never reused. `TestADRNumbersAreUniqueAndMatchTheirHeading`
(`server/internal/docsguard`) fails CI on a duplicate or on an H1 that disagrees
with its filename. If yours collides, renumber your file, not the other branch's.
Current range: **0001–0215**.

### Commit message format

```
<type>(<scope>): <subject>

<body>

Sprint: S<NN> — <sprint name>
Issue: #<issue_number>
```

`<type>` is one of `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `spike`.
`<scope>` is the top-level directory touched (`server`, `client`, ...). A change to
documentation only has no scope: `docs: <subject>`, never `docs(docs): …`. A PR
title follows the same rule, since it becomes the squash commit.

### Pull request format

A PR description has `## Summary` (1–3 bullets), `## Sprint` (S<NN> — name),
`## Issues` and `## Test plan` (a checklist), in that order. In `## Issues`, put
each closing directive on its own line (`Closes #<n>`, `Fixes`, `Resolves`).
Feature PRs target `develop`, so GitHub ignores closing keywords there;
`main-promotion-issue-close.yml` reads those lines after promotion. Prose such as
"does not close" is ignored on purpose.

### Storing anything about a person (ADR 0149)

A change that stores something new about a person, or changes how long it is kept,
updates in the same PR: `accountdelete.Inventory` and `Delete`, the privacy page
(`client/src/routes/Privacy.svelte`) and the numbers in `client/src/lib/privacy.ts`. Tests in
`accountdelete` and `lobby` enforce it; every migration newer than
`PRIVACY_SCHEMA_VERSION` needs that one-line bump.

### Model tiers

Every open issue has one `tier:*` label: the weakest model that can take it to a
merged PR unsupervised (1 < 2 < 3). Tier by the hardest part; pick higher when unsure.

| Label | Takes it to a merged PR |
|---|---|
| `tier:1-mechanical` | One area of the code, with a recipe or a sibling to copy: a card whose seams all exist, a docs fix, a copy or label change, a flake with an obvious cause. The acceptance test is clear from the issue text. Any capable model. |
| `tier:2-standard` | Several files, but it follows an existing ADR or pattern: a bug that has a repro, a small extension to an existing seam, a client feature built from existing components, a heuristic tweak with an arena or suite measurement. A mid-tier model or above. |
| `tier:3-design` | Needs an ADR or an owner decision not yet recorded, a new engine seam, CR edge cases (layers, replacement ordering, copies, loops), cross-cutting engine + protocol + client work, a snapshot-schema change, bot strategy design, a diagnosis with no repro, or an epic. The strongest model only. |

An agent told to stay at or below a tier lists candidates with
`gh issue list --label tier:1-mechanical` (plus `tier:2-standard` when allowed),
skipping `in-progress` and `question`. If the work is harder than its label, it
stops, comments why, and relabels. Re-tier an issue when its scope changes.

---

## 5. Commands and rules

Commands, environment variables, endpoints and test knobs are in `docs/`
(see §3). The rules that the code and its tests depend on are here.

### Go without a local toolchain

The host has no Go. Run it only through `scripts/go-docker.sh`, from the worktree
root: `scripts/go-docker.sh test ./...`, `lint`, `vet`, `go <args>`. Redirect its
output to a file and read that; piping through `tail` or `head` hides the exit
status. Never hand-write a `docker run` for Go, and never create a cache volume
beyond `cmdctrl-gomod`, `cmdctrl-gobuild` and `cmdctrl-golangci` (#2035).
Usage and env pass-through: [docs/commands.md](docs/commands.md#commands).

### Rules the code depends on

- **Snapshots are additive within a version (#522).** Changes inside one
  `SnapshotSchemaVersion` are additive only, and fixtures are never edited. A
  rename, removal or retype is a bump. Never regenerate or delete a fixture to make
  a test pass. [docs/snapshots.md](docs/snapshots.md).
- **Stack items name their catalog row.** A restore never drops a stack item or
  abandons the game, and an ability's identity is its catalog ref (#1497,
  ADR 0041 P9). Effect keys (a `ScopedEffect` mod kind, a stack-item effect key)
  are on-disk identities: never renamed, never reused. [docs/snapshots.md](docs/snapshots.md).
- **The closure ratchet.** A new route from `Game` to a func or an interface is
  classified in `server/internal/game/testdata/closure_fields.txt`. The `census:`
  and `transient` counts may only fall (`TestClosureFieldsReachableFromGame`).
- **The bot seat.** `aiseat.Start` subscribes before it returns (#938), so never
  move `room.Subscribe()` onto the runner goroutine. A test that waits on a bot
  table polls a monotone predicate, a `>=` on a counter that only climbs, a latch,
  or a history, never `== k` or a state the table merely passes through (#634).
  The arena lives in `server/internal/botarena/`, outside `aiseat/`, because
  `heuristic/imports_test.go` bans `internal/game` from every `aiseat/` package. The
  detail is in [docs/bot.md](docs/bot.md#engineering-notes).
- **Labels are a contract.** Render a UI name as `L.<key>` from
  `client/src/lib/labels.ts`, never as a literal, and rename it in the registry.
  A rename updates the e2e specs and any `*.hint.ts` in the same PR.
  List: [docs/labels.md](docs/labels.md).
- **Design skills** (`.claude/skills/`): `web-design-guidelines` for UI review; `design-taste-frontend` for signed-out pages only. See each `C2.md`.
- **UI work starts from [DESIGN.md](DESIGN.md).**

---

## 6. When you're unsure

1. Re-read [PLAN.md](PLAN.md): scope (section 3) and principles (section 5).
2. Check the current sprint in [docs/sprints.md](docs/sprints.md) — the scope
   there is authoritative for "what should I be working on right now".
3. Look for an ADR in [docs/decisions/](docs/decisions/) that settles it. If
   none does, surface the question to the user rather than guessing.
4. Prefer a spike (time-boxed, throwaway) over speculative architecture.

### Rules citations

`CR NNN.Nx` in code, tests and docs means the **Magic: The Gathering Comprehensive
Rules effective September 25, 2026** (the TXT download is `MagicCompRules 20260925.txt`).
Check a number against that text before you write it. Do not cite from memory:
rule numbers move between editions. Moving the pin to a newer edition, and the
search that finds every citation, are in [docs/rules-citations.md](docs/rules-citations.md).

---

## 7. Adding a catalog card (S14+)

The full guide moved to [docs/adding-cards.md](docs/adding-cards.md) (#1747).
**Read it before adding or changing any catalog card or engine seam.** It is
long on purpose; open the subsection you need, not the whole file.

Rules that must not be missed:

- **One file per card**, `server/internal/cards/effects/<snake_name>.go`, one
  `init()` each. The catalog is opt-in per Scryfall `oracle_id` (never
  `scryfall_id`).
- **Declare `Completeness` truthfully.** The zero value is `Unreviewed` and is a
  legal thing to ship; never stamp `Full` to tidy it. `Caveats` are written for
  a player, one sentence, no engine vocabulary.
- **Generate only your cards' oracle fixtures** (`-update-oracle -oracle-ids=...`);
  never regenerate them all in a card PR.
- **Stronger than printed is the wrong direction.** If the engine can't express
  a cost or clause, leave the card out rather than ship it without.

---

## 8. Things to explicitly *not* do

- Do not add payment of any kind: no paid tier, paid feature or paywall
  (PLAN.md §3). Do not add a second database, a new auth provider or a new
  hosted service without an ADR. SQLite and Discord sign-in are the stack
  (see [PLAN.md §4](PLAN.md#4-tech-stack)).
- Do not attempt a "full rules engine" sprint. Rules enforcement grows
  incrementally, one mechanic at a time, driven by the decks people bring.
- Do not reintroduce a dependency on XMage or any JVM tooling. That was
  evaluated and rejected in S01.
- Do not generate card art, card text, or anything else outside the Wizards
  Fan Content Policy; card data and images come from Scryfall.
- Do not create commits or PRs that do not reference a sprint (see section 4).
