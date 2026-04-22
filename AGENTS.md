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
    └── decisions/       # ADRs (0001 WS library, 0002 client framework, 0003 auth+lobby)
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

3. **Pick a `TargetMode`.** If the card has a target, declare it via
   `Spec.TargetMode` so the client pops the targeting UI before
   `cast_spell`. Valid values: `"any"`, `"player"`, `"creature"`,
   `"stack_spell"`, `"card_in_graveyard"`. Empty means no prompt.

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

Sacrifice-cost abilities (Lotus Petal) parse but reject at activation
in S15 — `ManaAbilityCost{Tap: true, Sacrifice: true}` will fail with
`ErrInvalidParam`. Keep adding them to specs; the activation gate
opens in a later sprint.

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

### When NOT to add a catalog entry

- **Non-mana, non-static activated abilities** (planeswalker +1/-1
  loyalty costs, equip, cycling, etc.) land with S19's activated-
  ability pipeline. Don't invent a shape; wait. (Mana abilities and
  static abilities are the exceptions — see the recipes above.)
- **Triggered abilities on non-ETB events** (die-to-graveyard, attack
  triggers, "whenever you cast a spell") land with S19's listener pipeline.
  Don't use `OnETB` as a workaround.
- **Replacement effects** ("enters tapped", "if would die, exile instead",
  "draw 2 instead of 1") land with S17. Cultivate / Path's land-fetch
  enters **untapped** in S14 — document the deferral in the card file.
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
