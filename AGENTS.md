# AGENTS.md — working on cmd_and_ctrl

This file is for AI coding agents (Claude Code, Codex, Cursor, etc.) operating in
this repository. Humans should read [PLAN.md](PLAN.md) first for the project
vision and phased roadmap; this file is about *how* to work, not *what* to build.

---

## 1. What this project is

A private, personal 4-player Magic: The Gathering Commander sandbox
with a Go game server and a TypeScript client. Personal-use only — not
a product. See [PLAN.md](PLAN.md) for scope, stack, and the Option B
("sandbox first, rules grafted in incrementally from S13+") decision.

Three hard problems, ranked: **game state and multiplayer sync** (Go
server, authoritative state, WebSocket broadcast), **UX polish** (the
whole point — Commander-specific affordances nothing else has), and —
long horizon — **incremental rules enforcement** (B→C track, starting
in S13+ after the sandbox is shipped).

---

## 2. Ground rules

1. **Read [PLAN.md](PLAN.md) before making architectural suggestions.** The
   central decision (reuse XMage vs build from scratch) is already made. Do not
   relitigate it without new information.
2. **This is a hobby project at ~10 hours/week.** Optimise for momentum and
   clarity, not enterprise rigour. No microservices, no k8s, no premature
   abstractions. One VPS. One database. One deployable per service.
3. **Personal use only changes the calculus.** Legal risk is low (Cockatrice and
   XMage exist). Scale is ≤8 users. Don't build for hypothetical public launch.
4. **UX polish is the entire point.** "Make it feel good" is core product
   work and deserves real care. Sandbox features (manual resolution,
   manual priority) are a deliberate design choice, not a shortcut.
5. **We do not use XMage.** This was evaluated and rejected in S01 — see
   [PLAN.md §2.2](PLAN.md#22-why-not-option-a-the-s01-discovery). The entire
   backend is an original Go game server; the rules engine (eventually,
   S13+) is grown incrementally on top of it. No Java anywhere in the stack.

---

## 3. Repo layout (evolving)

```
cmd_and_ctrl/
├── PLAN.md              # vision, roadmap, open decisions
├── AGENTS.md            # this file
├── Makefile             # top-level dev/test/lint targets
├── .devcontainer/       # Go + Node dev environment
├── server/              # Go game server (authoritative state, WebSocket + HTTP API)
│   ├── cmd/
│   │   ├── server/      # main package — serves :8080 with lobby + hub + cards routes
│   │   └── gamecli/     # dev WebSocket client for driving a game via v0 actions
│   ├── internal/
│   │   ├── game/        # authoritative domain: Game, Player, Zone, Card, Turn, mutations
│   │   ├── protocol/    # v0 wire format types + ViewOfGame + FilterViewFor
│   │   ├── actions/     # action type enum + Dispatch(Game, Action) router
│   │   ├── ws/          # gorilla/websocket hub, Room, RoomManager, per-viewer broadcast
│   │   ├── auth/        # pluggable Authenticator interface + MemoryAuthenticator + HTTP middleware
│   │   ├── lobby/       # GameMeta registry, invite flow, lobby HTTP handler, WSAuthorizer, deck upload
│   │   ├── cards/       # Scryfall index (streaming load) + disk-backed image cache + /cards routes
│   │   ├── bugstore/    # bug-report artifacts: reporter screenshots (public, Camo-reachable) + pinned replays (admin-only)
│   │   └── deck/        # decklist parsers (Moxfield, plain text) + Commander validation
│   ├── Makefile
│   └── .golangci.yml
├── client/              # TypeScript + Svelte 5 + Vite (PixiJS arrives in S05)
│   ├── src/
│   │   ├── App.svelte   # router shell
│   │   ├── main.ts
│   │   ├── app.css
│   │   ├── lib/         # protocol types, WebSocket client, session, api, hash router
│   │   └── routes/      # Login / Lobby / Join / Game views
│   ├── index.html
│   ├── vite.config.ts
│   ├── svelte.config.js
│   ├── tsconfig.json
│   ├── eslint.config.js
│   └── package.json
├── scripts/             # scryfall-refresh.sh (weekly cron) + one-off tools
├── data/                # runtime state (gitignored): Scryfall cache, snapshots, images
└── docs/
    ├── protocol.md      # v0 wire format spec
    ├── lobby.md         # lobby HTTP API reference
    ├── sprints.md       # sprint plan
    └── decisions/       # ADRs (0001 WS library … 0018 triggers on the stack)
```

When you create a new top-level directory, add it here.

---

## 4. Sprint and tracking discipline

Work is organised into 2-week sprints tracked in:

- **Project board:** https://github.com/users/krakenhavoc/projects/5
- **Milestones:** https://github.com/krakenhavoc/cmd_and_ctrl/milestones
- **Sprint plan with task checklists:** [docs/sprints.md](docs/sprints.md)

Each sprint has one tracking issue (`#1` through `#12`) whose body is the
checklist of sub-tasks. Commits reference the sprint issue number.

**Every commit and pull request must reference the sprint and issue it relates
to.** This is non-negotiable — it's how we keep a part-time, multi-month project
coherent.

### Commit message format

```
<type>(<scope>): <subject>

<body>

Sprint: S<NN> — <sprint name>
Issue: #<issue_number>
```

Example:

```
feat(server): add websocket ping/pong round-trip

Minimal gorilla/websocket hub accepting JSON frames on /ws. Replies to
v0 ping frames with pong carrying server_time. No game state yet.

Sprint: S01 — Go server + client scaffold
Issue: #1
```

`<type>` is one of: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `spike`.
`<scope>` is the top-level dir being touched (`server`, `client`, `docs`, etc.).

### Pull request format

PR descriptions must include:

```
## Summary
<1-3 bullets>

## Sprint
S<NN> — <sprint name>

## Issues
Closes #<n>, relates to #<n>

## Test plan
<bulleted checklist>
```

If a PR does not belong to the active sprint, say so explicitly and justify it.

---

## 5. Commands you'll actually run

*(Populate as the project takes shape. Empty sections are fine — don't invent.)*

### Dev environment

The devcontainer installs Go, Node, and the GitHub CLI. Ports 3000, 5173, and
8080 are forwarded. (The Java/Maven features remain installed for now but are
unused — they can be removed in a later cleanup PR.)

### Top-level
- `make help` — list targets
- `make server-dev` — run the Go server on :8080
- `make client-dev` — run the Vite dev server on :5173
- `make test` — server tests + client typecheck
- `make lint` — `go vet` + client ESLint + Prettier check

### Server (Go, `server/`)
- `make -C server dev` — run locally on :8080 (seeds a 4-player demo game)
- `make -C server test` — `go test -race -cover ./...`
- `make -C server vet` — `go vet ./...`
- `make -C server fmt` — `gofmt -s -w .`
- `make -C server build` — produces `server/bin/cmd_and_ctrl-server`
- `cd server && go run ./cmd/gamecli -addr ws://localhost:8080/ws` — drive the demo game from a terminal; reads action JSON on stdin or via `-script path.json`
- Endpoints: `GET /healthz`, `GET /ws` (protocol v0, see [docs/protocol.md](docs/protocol.md)), `POST /admin/login`, `/games*` lobby routes (see [docs/lobby.md](docs/lobby.md)), `/cards/*` image + metadata routes
- Env vars:
  - `CMDCTRL_ADDR` — listen addr (default `:8080`)
  - `CMDCTRL_DATA_DIR` — data root (default `./data`; empty string disables disk writes + card cache)
  - `CMDCTRL_ADMIN_TOKEN` — **required**. Shared admin secret for `POST /admin/login`. At least 16 characters.
  - `CMDCTRL_SESSION_TTL` — session lifetime as a Go duration (default `12h`)
  - `CMDCTRL_ALLOWED_ORIGINS` — comma-separated hostnames (or full URLs) permitted as cross-origin WebSocket callers. Same-origin is always allowed; unset = same-origin only.
  - `CMDCTRL_SEED_DEMO=1` — seed the S03 4-player demo game at startup for the gamecli dev loop
  - `CMDCTRL_DISCORD_CLIENT_ID` / `CMDCTRL_DISCORD_CLIENT_SECRET` / `CMDCTRL_DISCORD_REDIRECT_URI` — S12.5 OAuth credentials. Unset disables the Discord sign-in button (manual name entry still works).
  - `CMDCTRL_SECURE_COOKIES` — truthy sets the `Secure` attribute on the session cookie. Enable in any TLS deployment; leave unset for plain-HTTP local dev (a `Secure` cookie is never sent over http and would silently break login).
  - `CMDCTRL_TRUST_FORWARDED` — truthy keys the lobby rate limiter off the leftmost `X-Forwarded-For` hop instead of the socket `RemoteAddr`. Enable **only** when the server sits behind a trusted reverse proxy that sets the header; otherwise clients can spoof it to dodge limits.
  - `CMDCTRL_DEV_RELAX_RATE_LIMITS` — truthy effectively disables the lobby rate limiters. For the e2e suite and local load-y dev loops only; logs a loud warning on boot. **Never set in production.**
  - `CMDCTRL_GITHUB_TOKEN` — enables the in-app "report a bug" button (`POST /bugreport` files a GitHub issue). Use a fine-grained PAT with **Issues: write** on the one repo, nothing broader. Unset disables the feature and the client hides the button. See [ADR 0017](docs/decisions/0017-bug-report-button.md). **Provisioned by CI/CD**: the deploy job upserts it into `/etc/cmd_and_ctrl/env` (the env file the HomeLab cloud-init template writes and the systemd units read) from the `CMDCTRL_GITHUB_TOKEN` Actions secret, via `sudo scripts/set-server-env.sh`; to rotate, update the secret and rerun the deploy — no host access needed.
  - `CMDCTRL_GITHUB_REPO` — `owner/name` slug issues are filed against (default `krakenhavoc/cmd_and_ctrl`).
  - `CMDCTRL_PUBLIC_BASE_URL` — the origin this server is reachable at from the public internet (e.g. `https://cmd.labxp.io`). Needed for bug-report **screenshots**: GitHub renders an issue image by fetching it through its Camo proxy, so the URL in the issue body has to be absolute and publicly resolvable. Falls back to `CMDCTRL_CLIENT_BASE_URL`; with neither set (or no `CMDCTRL_DATA_DIR`) attachments are disabled, `/bugreport/config` reports `attachments:false`, the modal hides its file picker, and text reports keep working. **No default on purpose** — a wrong origin produces issues full of broken images, which is worse than a deploy that doesn't offer upload. Provisioned by CI/CD from the `CMDCTRL_PUBLIC_BASE_URL` repo variable. See [ADR 0017 §6](docs/decisions/0017-bug-report-button.md).
- Cron: `scripts/scryfall-refresh.sh` — weekly refresh of the Scryfall default-cards dump (suggested cron: `0 5 * * 0`)

### Discord bot (Go, `server/cmd/bot/`)
- Separate binary from the game server; runs as `cmd-and-ctrl-bot.service` on the prod VPS. See [docs/decisions/0004-discord-identity.md](docs/decisions/0004-discord-identity.md).
- `make -C server build-bot` — produces `server/bin/cmd_and_ctrl-bot`
- Commands: `/cc-invite [name]` (channel-visible invite URL) and `/cc-games` (ephemeral list).
- Env vars (bot binary reads these; server binary does not):
  - `CMDCTRL_DISCORD_BOT_TOKEN` — **required**. Discord Developer Portal → Bot → Reset Token.
  - `CMDCTRL_DISCORD_APP_ID` — **required**. Application ID from the same portal.
  - `CMDCTRL_DISCORD_GUILD_IDS` — **required**. Comma-separated guild snowflakes; commands register only on these guilds and the bot rejects interactions from any other.
  - `CMDCTRL_ADMIN_TOKEN` — **required**. Same shared secret the server uses; the bot hits `POST /admin/login` + `POST /games` + `GET /games` over loopback.
  - `CMDCTRL_SERVER_BASE_URL` — default `http://127.0.0.1:8080`. Where the bot calls the admin API.
  - `CMDCTRL_CLIENT_BASE_URL` — default `https://cmd.labxp.io`. Used to compose the invite URL posted back to Discord.
- Unset `CMDCTRL_DISCORD_BOT_TOKEN` disables the bot (binary exits 0 after logging `bot disabled`). Convenient for dev stacks without a registered Discord app.
- Store bot secrets in a dedicated env file (`/etc/cmd_and_ctrl/bot.env`, mode `0640`) rather than the server's env file — ADR 0004 §6 explains why.

### Client (TypeScript + Svelte 5 + Vite, `client/`)
- `cd client && npm install` — first-time setup
- `npm run dev` — Vite dev server on :5173, proxies `/ws` to the Go server
- `npm run build` — type-check + production build into `client/dist/`
- `npm run check` — `svelte-check` typecheck only
- `npm run lint` — ESLint + Prettier check
- `npm run format` — auto-format

---

## 6. When you're unsure

1. Re-read [PLAN.md](PLAN.md) section 2 (architectural decision) and section 6
   (roadmap).
2. Check the current sprint in [docs/sprints.md](docs/sprints.md) — the scope
   there is authoritative for "what should I be working on right now".
3. Look for an open decision in [PLAN.md](PLAN.md) section 7. If the question is
   listed there, surface it to the user rather than guessing.
4. Prefer a spike (time-boxed, throwaway) over speculative architecture.

---

## 7. Adding a catalog card (S14+)

The catalog at [server/internal/cards/effects/](server/internal/cards/effects/) is
opt-in per Scryfall `oracle_id`. A card that isn't in the catalog keeps its
pre-S14 manual sandbox behaviour; a card that is in the catalog resolves
automatically at the right stack boundary. Each card is one `init()` in its
own file — one-file-per-card keeps `git blame` clean and the merge-conflict
surface tiny.

### Recipe

1. **Find the oracle ID.** The Scryfall bulk dump at
   `data/scryfall/default-cards.json` (local only; refreshed weekly via the
   scheduled workflow) has every printing. One-liner:
   ```bash
   python3 -c "import json; d=json.load(open('data/scryfall/default-cards.json')); \
     print(next(c['oracle_id'] for c in d if c['name']=='CARD NAME'))"
   ```
   `oracle_id` (NOT `scryfall_id`) is the catalog key — stable across
   printings.

2. **Pick a primitive composition.** See
   [server/internal/cards/effects/primitives.go](server/internal/cards/effects/primitives.go)
   for the 15 primitives. Most cards are 1-2 primitives sequenced. A
   card that can't be expressed with existing primitives either needs a new
   primitive (add it to `primitives.go`) or a new `*ForEffect` helper on
   `*Game` (under a lock caller already holds — follow the existing naming
   in [server/internal/game/effect_api.go](server/internal/game/effect_api.go)).

3. **Declare the target (S20).** If the card has a target, build a
   `Spec.Targets` from the constructors in
   [targets.go](server/internal/cards/effects/targets.go) so it reads
   like the oracle text:
   ```go
   Targets: TargetAny(),                                           // Lightning Bolt
   Targets: TargetCreature("target nonblack creature", NonBlack()), // Doom Blade
   Targets: TargetSpell("target noncreature spell", Noncreature()), // Negate
   Targets: TargetPlayer("target opponent", Opponent()),
   Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
   Targets: TargetCardInGraveyard("target card in your graveyard", YouOwn()),
   ```
   Predicates compose with `And` / `Or` / `Not`; add missing ones to
   `targets.go`, not to the card file. Multi-target clauses set the
   count on the same spec — `TargetCreature("two target nonartifact
   creatures", Not(Artifact())).WithCount(2, 2)`, `.WithCount(0, 2)`
   for "up to two", `.WithCount(1, 0)` for "any number" — and their
   `OnResolve` iterates `ctx.LegalTargets()` (or indexes
   `item.Targets` with `ctx.IsTargetLegal` per slot when the order
   matters, as in Arc Trail) so a target that left in response is
   skipped rather than erroring. The engine computes the legal
   set for the client's picker on every snapshot, rejects an illegal
   pick at announce (`ErrIllegalTarget`, CR 601.2c), and re-runs the
   same predicate at resolution (CR 608.2b). Colour predicates read
   `Card.Colors` (Scryfall's computed colours; mana-cost fallback for
   fixtures). See [ADR 0019](docs/decisions/0019-structured-targeting.md).
   The legacy `TargetMode` string is derived from `Targets.Mode` —
   set it directly only for a card you deliberately leave on the
   free-form picker. Empty means no prompt.

   **Modal cards ("Choose one —", S20 sub-PR 4)** declare
   `Spec.Modes` instead of `Spec.Targets`, with the target clause on
   the option that has one:
   ```go
   Modes: ChooseOne(
       Mode("Exile target player's graveyard.", TargetPlayer("target player")),
       Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
       Mode("Each creature deals 1 damage to its controller."),
   ),
   // "Choose two —": ChooseN("Choose two", 2, 2, Mode(…), Mode(…), …)
   ```
   `OnResolve` is a run of `if ctx.HasMode(i) { … }` blocks in
   printed order (CR 700.2c). The engine validates the choice at
   announce and applies the chosen option's target clause exactly as
   it would a card-level one; the client shows a mode picker before
   targeting. Limit: one targeted option per cast — `Register`
   panics on a `Max > 1` card with two targeted options (per-mode
   target slots ride with multi-target).

4. **Write the card file.** One file per card at
   `server/internal/cards/effects/<snake_name>.go`:
   ```go
   package effects

   import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

   // Card Name — "full oracle text quoted."
   //
   // S14 sandbox simplifications (if any — enters-tapped, auto-pick,
   // target-at-ETB deferrals, etc.). Be specific about WHAT is
   // deferred and to which future sprint.
   func init() {
       Register(Spec{
           OracleID:   "<uuid from step 1>",
           Name:       "Card Name",
           TargetMode: "<player / creature / ...>",
           OnResolve: func(item *game.StackItem, ctx *Context) error {
               // primitive composition here
               return nil
           },
       })
   }
   ```
   For permanents with an ETB trigger, populate `OnETB` instead of
   (or alongside) `OnResolve`. For planeswalkers, set
   `StartingLoyalty` — the ETB hook stamps loyalty counters
   automatically.

5. **Add a test case** in
   [cards_test.go](server/internal/cards/effects/cards_test.go). Use
   `newCatalogGame(t)` + `castCatalogSpell(t, g, name, typeLine, oracleID, targets)`
   + `passPriorityAroundTable(t, g)` and assert the resulting state. For
   OnETB tests, the ETB fires inline during resolution — no extra setup
   needed. For library tutors, seed needles via `pushLibraryCardForTest`
   (which uses `PushBottom` so the "first match" sandbox pick is
   deterministic).

6. **Verify.** `cd server && go test ./internal/cards/effects/...` and
   `gofmt -l internal/cards/effects/` should both be clean. The
   `TestNonCatalogSpellStaysSandbox` canary should still pass — it's the
   opt-in invariant.

7. **Manual smoke-test.** `CMDCTRL_DEV_SKIP_DECK_VALIDATION=1 make server-dev`
   plus a small deck (`make dev-skip-validation` target on the top-level
   Makefile) so the library is small enough to find your card quickly.
   Cast it, verify the AUTO badge renders, verify the effect resolves.

### Adding a mana ability (S15+)

Mana abilities live on the same `Spec{}` struct via the optional
`ManaAbilities []ManaAbility` field. Used by Sol Ring, Arcane Signet,
Birds of Paradise today; basic lands fall back to a synthetic shape
the engine derives from `TypeLine` (no spec needed).

```go
func init() {
    Register(Spec{
        OracleID: "<uuid>",
        Name:     "Sol Ring",
        ManaAbilities: []ManaAbility{
            {
                Cost:     ManaAbilityCost{Tap: true},
                Produced: "{C}{C}",
                Label:    "Add {C}{C}",
            },
        },
    })
}
```

`Produced` is parsed by `game.ParseProducedMana`. Single-color slots
drop straight into the controller's pool when the ability fires;
multi-option slots use **pipe syntax** and queue a `mana_pick`
PendingChoice for the controller to resolve:

- `"{C}{C}"` — Sol Ring: two colorless slots.
- `"{W|U|B|R|G}"` — Birds of Paradise: one any-color slot, picker.
- `"{W|U|B|R|G}"` + `commanderIdentityFor` filter — Arcane Signet:
  the engine narrows the pipe set against the controller's commander
  identity at activation time.

Mana abilities can carry cost components beyond `{T}`:

| Cost | Field | Card |
| --- | --- | --- |
| `{T}` | `ManaAbilityCost{Tap: true}` | Sol Ring |
| Sacrifice this | `ManaAbilityCost{Sacrifice: true}` | Lotus Petal, Treasure |
| Sacrifice another permanent | `ManaAbilityCost{SacrificeOther: SacrificeACreature().SacrificeOther}` | Ashnod's Altar, Phyrexian Altar |

`SacrificeOther` takes a `*game.TargetSpec`, the same shape the CR 602
activated abilities use — build it with the `SacrificeACreature()` /
`SacrificeAPermanent()` helpers and take their `.SacrificeOther` field
rather than writing a spec by hand. The engine filters the candidate
set to the controller's own permanents (CR 701.17b), stamps it onto
`ManaAbilityView.SacrificeOptions`, and the client reuses
`SacrificeCostModal` to pick one. The chosen card comes back in the
`activate_mana_ability` payload as `sacrifice_ids`, and
`ManaAbilityParams.SacrificeIDs` carries it into the engine.

`ActivateManaAbility` validates every component before paying any of
them, so an illegal sacrifice choice leaves the source untapped. Mana
lands in the pool first and the dies-triggers go on the stack after
(CR 605.3a — a mana ability doesn't use the stack, but the sacrifice
still triggers), which is what makes Ashnod's Altar + a drain outlet
work.

Summoning sickness applies to any mana ability with a tap cost on a
creature source (CR 302.1) — Birds of Paradise, Palladium Myr. The
engine enforces it inside `ActivateManaAbility`; specs don't declare
it.

For non-mana, non-static activated abilities (planeswalker +1/-1,
equip, cycling, etc.), wait — see the deferral list below.

### Adding a static ability (S16+)

Static abilities (anthems, type-changers, keyword grants, CDAs) live
on `Spec.Static []game.StaticAbility`. The layer engine recomputes
from scratch on every relevant event (battlefield zone change,
counter change); the wire-side `power` / `toughness` / `type_line` /
`abilities` fields reflect the post-layer effective characteristics.

```go
import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
    Register(Spec{
        OracleID: "<uuid>",
        Name:     "Glorious Anthem",
        Static: []game.StaticAbility{
            {
                Layer:    game.Layer7PT,
                SubLayer: game.SubLayer7C_Modify,
                AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
                    return target.IsCreature() && target.Controller == source.Controller
                },
                Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
                    c.Power++
                    c.Toughness++
                },
            },
        },
    })
}
```

**Layer / SubLayer choices** (CR 613):

| What you're doing | Layer | SubLayer |
|---|---|---|
| Add a creature type / artifact / enchantment | `Layer4Type` | (ignored) |
| Grant a keyword (flying, trample, etc.) | `Layer6Ability` | (ignored) |
| Set P/T to a specific value (Tarmogoyf-style CDA) | `Layer7PT` | `SubLayer7A_CDA` |
| Modify P/T (+1/+1 anthem) | `Layer7PT` | `SubLayer7C_Modify` |
| +1/+1 / -1/-1 counter math | (don't — counter math stays in `CurrentPower`) | — |

**`AppliesTo` patterns:**
- "Creatures you control" — `target.IsCreature() && target.Controller == source.Controller`
- "OTHER X you control" — add `target.InstanceID != source.InstanceID`
- Has subtype X — read `target.Effective().Subtypes` (so type-add effects compose)
- Self-only (CDA) — `target.InstanceID == source.InstanceID`

**`Apply` patterns:**
- Anthem +1/+1 — `c.Power++; c.Toughness++`
- Type-add — append to `c.Types` after checking idempotency
- Keyword grant — append to `c.Abilities` after checking duplicate
- CDA P/T — `c.Power = computed; c.Toughness = computed + 1`

**Tests** — see [anthem_test.go](server/internal/cards/effects/anthem_test.go) and [tarmogoyf_test.go](server/internal/cards/effects/tarmogoyf_test.go) for the layer-aware pattern. Use `pushBattlefieldCardWithTimestamp` (fires `EventZoneMove` so the listener stamps `EnteredBattlefieldAt` + bumps `layerVersion`); read effective characteristics via `effectivePower` / `effectiveToughness` / `effectiveTypes` / `effectiveAbilities` helpers.

**Don't bypass the printed/effective split:** if an effect needs to read another card's characteristic, use `target.Effective()` not `target.Power` / `target.TypeLine`. Reading printed values inside `AppliesTo` or `Apply` is a layer-ordering bug waiting to happen.

### Adding a replacement effect (S17+)

Replacement effects ("enters tapped", "if that would place counters,
place twice that many instead", "if a player would draw a card, that
player mills instead") live on the same `Spec{}` struct via the
optional `Replacements []game.ReplacementEffect` field. Used today by
Doubling Season, Hardened Scales, Kismet, Stasis, Hangarback Walker,
Fog, Library of Leng.

Unlike static abilities, replacements fire **before** the event
happens — the pipeline constructs a `game.ReplacementEvent`, the
engine offers each applicable replacement a chance to mutate or
cancel it, then the underlying mutation runs (or is skipped, if
canceled). CR 616.1 iterative apply-loop, CR 614.5 once-per-event
tracking, and CR 616 affected-player-chooses-order are enforced
centrally.

```go
import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
    Register(Spec{
        OracleID: "<uuid>",
        Name:     "Doubling Season",
        Replacements: []game.ReplacementEffect{
            {
                Watches: []game.EventKind{game.EventCounterPlaced},
                AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
                    if ev.Kind != game.RepEventCounter { return false }
                    target, ok := g.LookupCardForEffect(ev.CounterTarget)
                    if !ok { return false }
                    return target.Controller == src.Controller
                },
                Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
                    ev.CounterDelta *= 2
                    return nil
                },
                Controller: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) uuid.UUID {
                    return src.Controller
                },
                Label: "Doubling Season: double counters",
            },
        },
    })
}
```

**Kind picker** (CR 614):

| What the replacement watches | `ReplacementEventKind` | Relevant fields |
|---|---|---|
| Counter placement (+1/+1, loyalty, …) | `RepEventCounter` | `CounterTarget`, `CounterName`, `CounterDelta` |
| Zone motion (ETB, LTB, draw-as-move) | `RepEventMove` | `CardID`, `OldZone`, `NewZone`, `NewZoneOwner`, `EntersTapped`, `EntersWithCounters` |
| Card draw | `RepEventDraw` | `DrawPlayer` |
| Life total change | `RepEventLife` | `LifePlayer`, `LifeDelta` |
| Damage (combat and direct) | `RepEventDamage` | `DamageSource`, `DamageTarget`, `DamageAmount`, `IsCombatDamage` |
| Step entry (skip-step) | `RepEventStepTransition` | `StepTransitionStep`, `StepTransitionSeat` |

**`AppliesTo` patterns:**
- "Counters go on a creature you control" — `target.Controller == src.Controller && target.IsCreature()`
- "When a permanent enters the battlefield" — `ev.Kind == RepEventMove && ev.NewZone == ZoneBattlefield`
- Self-replacement (Hangarback's X counters on own ETB) — `ev.CardID == src.InstanceID`
- Opponents only (Kismet) — `controllerOf(ev.CardID) != src.Controller`

**`Replace` patterns:**
- Counter multiplier — `ev.CounterDelta *= 2` (Doubling Season)
- Counter addition — `ev.CounterDelta += 1` (Hardened Scales)
- Cancel — `ev.Cancel()` (Fog, Stasis)
- Redirect move — `ev.NewZone = ZoneBottomOfLibrary` (Library of Leng)
- Enters-tapped — `ev.EntersTapped = true` (Kismet)
- Enters-with-counters — `ev.AddCounterAtETB("+1/+1", n)` (Hangarback Walker)

**Tests** — see `server/internal/cards/effects/doubling_season_test.go` for the CR 616 ordering pattern (Doubling Season + Hardened Scales → the affected player picks order → `[HS, DS]` yields 4 counters, `[DS, HS]` yields 3). Use `pushBattlefieldCardWithTimestamp` to get the source on the battlefield + the listener to stamp `EnteredBattlefieldAt`; trigger the event with the public mutation (`AddCounter`, `DrawCard`, etc.) and assert on the resulting state plus any queued `PendingChoice`.

**Don't use the replacement pipeline when a primitive flag suffices.** "This card does X to a land it fetches" (Cultivate, Path to Exile, Solemn Simulacrum) is a self-contained card behavior, not a general replacement. Declare `TappedOnEntry: true` on the `SearchLibrary` primitive rather than a full `ReplacementEffect`. The generic pipeline is for effects that watch *other* cards' events.

### Adding a combat-keyword card (S18+)

Creatures with printed combat keywords (flying, reach, deathtouch,
lifelink, trample, vigilance, first strike, double strike, menace,
defender, haste, flash) declare those keywords through a single
`Spec.PrintedKeywords []string` slot. The engine auto-generates a
Layer 6 `StaticAbility` at catalog load time that appends each
keyword to the card's own `Characteristic.Abilities`, so
battlefield-side consumers see the same surface Lord of Atlantis
uses for granted keywords. The same list feeds off-battlefield
reads (flash gating on a card in hand).

```go
import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
    Register(Spec{
        OracleID:        "<uuid>",
        Name:            "Serra Angel",
        PrintedKeywords: []string{"flying", "vigilance"},
    })
}
```

Keywords are bare strings, case-sensitive, matching the
canonicalised forms the engine expects. Canonical tokens:

| Token | Keyword |
|---|---|
| `"flying"` | Flying (CR 702.9) |
| `"reach"` | Reach (CR 702.17) |
| `"first strike"` | First strike (CR 702.7) |
| `"double strike"` | Double strike (CR 702.4) |
| `"deathtouch"` | Deathtouch (CR 702.2) |
| `"lifelink"` | Lifelink (CR 702.15) |
| `"trample"` | Trample (CR 702.19) |
| `"vigilance"` | Vigilance (CR 702.20) |
| `"menace"` | Menace (CR 702.110) |
| `"defender"` | Defender (CR 702.3) |
| `"haste"` | Haste (CR 702.10) |
| `"flash"` | Flash (CR 702.8) |

**Layer-granted keywords still use `Spec.Static`.** Lord of Atlantis
grants `"flying"` to *other* Merfolk via a conditional Layer 6
`StaticAbility` — that pattern stays. `PrintedKeywords` is only for
the card's own printed keywords.

**Tests** — assert `Effective().Abilities` contains the keyword
strings after the card is pushed to the battlefield. See
`server/internal/cards/effects/serra_angel_test.go` for the template.
Combat behaviour (flying block restriction, trample overflow, etc.)
is tested in `server/internal/game/combat_test.go` against
manufactured battlefield state — card-level tests just verify the
keyword strings are exposed.

**Keyword behaviour is engine-side, not catalog-side.** You do not
write flying/trample/deathtouch logic in the card file. The combat
engine reads `HasKeyword(card, "flying")` and routes accordingly.
Card files declare the strings; the engine does the rest.

### Adding a triggered ability (S19+)

Triggered abilities ("when ~ enters", "when ~ dies", "at the
beginning of your upkeep") live on `Spec.Triggered
[]game.TriggeredAbility`. A per-game harvester listens to the event
log, runs `AppliesTo` for each watched event kind, and calls `Build`
to put an item on the stack. The item resolves — and its `Effect`
runs — only when every player has passed priority in succession,
so opponents can respond (counter the ability, remove the target,
sacrifice in response). See [ADR 0018](docs/decisions/0018-triggers-on-the-stack.md).

```go
import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
    Register(Spec{
        OracleID: "<uuid>",
        Name:     "Mulldrifter",
        Triggered: []game.TriggeredAbility{{
            Watches: []game.EventKind{game.EventETB},
            AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
                return ev.CardID == source.InstanceID
            },
            Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
                return game.NewTriggeredItem(source, "Mulldrifter — draw two cards",
                    func(g *game.Game, item *game.StackItem) error {
                        return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
                    })
            },
        }},
    })
}
```

**Adding an activated ability (S21+):** put it in
`Spec.Activated`, one entry per printed ability, with the cost built
from the constructors in
[activated.go](server/internal/cards/effects/activated.go):

```go
Activated: []ActivatedAbility{{
    Label:   "Sacrifice a creature: deal 1 damage to any target",
    Cost:    SacrificeACreature(),          // or TapCost(), SacrificeThis(),
    Targets: TargetAny(),                   // ManaCost("{1}{B}"), PayLife(2),
    Effect: func(g *game.Game, item *game.StackItem) error {
        // Same contract as a triggered ability's item: never
        // capture a *Card; read the source via NewContext(g, item).
    },
}},
```

Compose multi-part costs with `Plus(ManaCost("{2}"), TapCost())`.
The engine validates every component before paying any of them, and
pays at announce — so a sacrifice cost's dies-triggers land on the
stack above the ability and resolve first. Mana abilities do NOT go
here (they skip the stack, CR 605.3a); they stay in `ManaAbilities`.
See [ADR 0020](docs/decisions/0020-activated-abilities.md).

**Adding an additional cost to cast (S21 sub-PR 5):** "As an
additional cost to cast this spell, discard a card" goes in
`Spec.AdditionalCost`, not in `OnResolve`:

```go
AdditionalCost: DiscardCost(1),   // Thrill of Possibility, Big Score
```

The distinction is observable, which is why it's modelled: the cost
is paid to CAST the spell, so the discard happens with the spell
already on the stack and a discard payoff (Mary Read and Anne Bonny,
Marauding Mako) triggers ABOVE it and resolves first. It is also
paid whether or not the spell resolves — countering it doesn't give
the card back. Fold the discard into `OnResolve` and both of those
go wrong. The caster picks the cards in a client prompt that opens
before the X / mode / target prompts; they ride `cast_spell` as
`discard_ids` and the engine validates them at announce (the spell
itself is never a legal pick — CR 601.2a already moved it to the
stack). See [ADR 0021](docs/decisions/0021-additional-costs.md).

**Event picker:**

| Trigger text | `Watches` | `AppliesTo` |
|---|---|---|
| "When ~ enters the battlefield" | `EventETB` | `ev.CardID == source.InstanceID` |
| "When ~ dies" | `EventLTB` | `cardDied(ev, source)` (graveyard-only; bounce / exile don't count) |
| "Whenever you sacrifice a permanent" | `EventSacrifice` | `ev.Actor == source.Controller` — fires while the permanent is still on the battlefield, before its `EventLTB` |
| "Whenever a player sacrifices a permanent" | `EventSacrifice` | `ev.CardID != uuid.Nil` (Mayhem Devil) — any player, any permanent type |
| "Whenever another creature dies" | `EventLTB` | `diedCreature(ev, g)` — resolves the dying card post-move; add `dead.Controller == source.Controller` for "you control", `!IsToken(dead)` for "nontoken" |
| "At the beginning of your upkeep" | `EventBeginUpkeep` | `ev.Actor == source.Controller` |
| "Whenever you cast a creature spell" | `EventCast` | `ev.Actor == source.Controller` + `g.LookupCardForEffect(ev.CardID)` for the spell's type |
| "Whenever an opponent casts their first noncreature spell each turn" | `EventCast` | `g.CastTallyFor(ev.Actor).Noncreature == 1` (tally is bumped before the event fires) |
| "Whenever an opponent draws a card" | `EventDrawCard` | `ev.Actor != uuid.Nil && ev.Actor != source.Controller` — fires once per card |
| "Whenever ~ deals combat damage to a player" | `EventDealDamage` | `ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)` |
| "Whenever a creature you control deals combat damage to a player" | `EventDealDamage` | `combatDamageToPlayerBy(ev, source.Controller, g)` — checks `ev.Combat`, player target, creature source |
| "…its controller may draw" (Edric) | `EventDealDamage` | `ev.Actor` is the dealing creature's controller; use it for both `OptionalPrompt.Chooser` and the draw |

**The two rules that matter:**

1. **`Build` builds; it does not resolve.** Do the work inside the
   `Effect` closure passed to `NewTriggeredItem`, never in `Build`
   itself. Applying the effect in `Build` skips the stack and denies
   every player their response window. The only game reads `Build`
   should do are the ones that pick targets.
2. **The `Effect` closure reads everything off `item` and the `g` it
   receives.** Don't capture `source *game.Card` (a pointer into a
   zone slice) or the `*game.Game` from `Build`'s arguments — undo
   restores a cloned game and the closure has to resolve against
   that one. `item.Controller`, `item.SourceCardID`, `item.Targets`
   carry what you need.

**Targeted triggers** declare the clause on the ability, exactly
like a spell's `Spec.Targets`:
```go
Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
```
The engine does the rest (S20): it computes the legal set when the
trigger fires and removes the trigger if the set is empty (CR
603.3d — no prompt at all), asks "you may" if there is one, then
queues a `pick_target` prompt the controller answers by clicking
the board. The chosen ref arrives in `item.Targets[0]` — your
`Effect` reads it from there (see `destroyChosenTargetTrigger` in
[acidic_slime.go](server/internal/cards/effects/acidic_slime.go)) —
and resolution re-checks it (CR 608.2b). Never pick a target
inside `Build`.

**"You may" triggers** set `OptionalPrompt: &game.TriggerOptionalPrompt{Question: "..."}`.
The harvester queues a yes/no `PendingChoice` instead of calling
`Build`; on "Yes", `Build` runs (via the target pick first, if the
trigger is targeted) and the item goes onto the stack.

**"Unless that player pays {N}"** (Rhystic Study, Smothering Tithe,
Esper Sentinel) is a `PayUnless` primitive the trigger's `Effect`
applies: it queues a `pay_unless` prompt for the taxed player and
returns; the "unless" consequence runs later as `OnDecline` when
they answer "Don't pay" — or "Pay" without the mana in pool +
untapped sources. Capture the payer's ID in `Build` (it's
`ev.Actor` for cast / draw events) and read the controller off the
`Context` inside `OnDecline`. See
[rhystic_study.go](server/internal/cards/effects/rhystic_study.go).

**Dies triggers** get the CR 603.10 last-known-information
characteristics as the third `Build` argument — the card is already
in the graveyard when `Build` runs, so read power / toughness /
types from `sourceLKI`, not `source`.

**Tests** — `castCatalogSpell` + `passPriorityAroundTable` settles
the spell *and* the trigger it queues (the helper waits for
`Game.Stack`, `StackMeta`, and `PendingTriggers` to all empty).
Assert the trigger is on the stack with `triggerOnStack(g, cardID)`
before the second pass if the timing is the point of the test. For
optional triggers, `answerLatestTriggerPrompt` then
`passPriorityAroundTable` again. Upkeep triggers: `advanceToUpkeepOf`
then `passPriorityAroundTable`. See the S19 sections of
[cards_test.go](server/internal/cards/effects/cards_test.go).

### When NOT to add a catalog entry

- **Non-mana, non-static activated abilities** (planeswalker +1/-1
  loyalty costs, equip, cycling, etc.) land with a later activated-
  ability pipeline. Don't invent a shape; wait. (Mana abilities and
  static abilities are the exceptions — see the recipes above.)
- **Triggers on events the engine doesn't emit yet** (attack
  declarations, "whenever a creature enters under an opponent's
  control", landfall-with-a-target) — check
  [events.go](server/internal/game/events.go) for an `EventKind`
  first. If there isn't one, the event plumbing is the PR, not the
  card.
- **Cost-replacement effects** (Trinisphere, Thalia, Spellshift, Kambal)
  touch the S15 cost engine rather than the S17 event pipeline. They
  land with S28.
- **Aura-attachment + control-change** (Mind Control) — requires
  aura-attaching state the engine doesn't model. Lands with S24.
- **Cards that need a pick-from-zone UI** the client doesn't have yet —
  e.g. "target card in any graveyard" (Regrowth / Eternal Witness) works
  today only because S14 sandbox auto-picks the top of the controller's
  graveyard. If the picker UX is load-bearing for the card, wait for S20
  smart-cast.

### When in doubt

[ADR 0010](docs/decisions/0010-card-effect-catalog.md) captures every
architectural decision and sandbox simplification the catalog was built
around. Start there.

---

## 8. Things to explicitly *not* do

- Do not add CI/CD beyond basic lint + test until there's code to protect.
- Do not introduce a database, auth provider, or payment anything without
  discussion. Shared password is fine for now (see [PLAN.md §4](PLAN.md#4-tech-stack)).
- Do not attempt a "full rules engine" sprint. Rules enforcement grows
  incrementally in the S13+ B→C track, one mechanic at a time, driven by
  real games. See [PLAN.md §2.3](PLAN.md#23-why-not-option-c-straight-custom-rules-engine).
- Do not reintroduce a dependency on XMage or any JVM tooling. That was
  evaluated and rejected in S01.
- Do not generate card art, card text, or anything else that would pull this
  project out of "private, personal use" territory.
- Do not create commits or PRs that do not reference a sprint (see section 4).
