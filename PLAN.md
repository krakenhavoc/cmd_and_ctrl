# C2 — charter

A living charter for *cmd & ctrl* (**C2**) as it is now: what it is, who the
public beta is for, what is in and out of scope, and how we will know the beta
is working. It replaces the day-one plan of 2026-04-10 (see [History](#history)).
Change it in a PR, like code, when the owner changes direction.

---

## 1. What C2 is

C2 is a browser table for four-player **Magic: The Gathering Commander**. It
is free, unofficial fan content and open source.

- **An automation-first rules engine.** An original Go server holds the
  authoritative game. Cards in the catalog resolve on their own: stack,
  priority, combat, triggers, layers, replacement effects. A card the engine
  doesn't know yet still plays. It resolves with no effect and the table
  settles it by hand, so a game never stalls on a missing card.
- **Bots.** An empty seat can be filled with a bot that plays its deck: the
  heuristic tiers by default, model-backed tiers where a budget allows. That
  means one person can play a full game.
- **A Discord community.** The C2 Discord server is where people find each
  other, ask for access and report problems. The `/c2-…` bot commands open
  tables and check decks.

## 2. Who the public beta is for

**Commander players who want a four-player table online, with or without a
pod of their own.** They come in by one of three doors:

1. **Members.** They join the C2 Discord and are given the *Elder Dragon
   Highlander* role, which lets them sign in with Discord, host tables, keep
   decks and play with others from the server
   ([ADR 0156](docs/decisions/0156-a-front-door-site-shell-front-page-and-a-discord-gate.md) §1).
2. **Invited guests.** A member sends them a table's invite link. They sit at
   that one table with no account.
3. **Visitors.** Anyone can play a practice game against bots from the front
   page, with no sign-in.

The beta is small on purpose: a soft launch to 20–50 players (S70), sized to
one VM capped at about 30 tables
([ADR 0148](docs/decisions/0148-capacity-limits-for-a-public-launch.md)).

## 3. Scope

**In scope:** four-player Commander (two- and three-seat tables are allowed);
the rules engine and card catalog, which grow with the decks people bring;
bots; deck import and coverage checks; Commander affordances (command zone,
commander damage, politics, monarch and initiative); accounts, saved decks
and settings; the public roadmap, About, privacy and legal pages; moderation;
operating one public host safely: capacity caps, monitoring, backups.

**Out of scope, and staying out:**

- **XMage, or any JVM component.** Rejected in S01 ([History](#history)).
- **Payments of any kind.** Nothing is sold. Hosted model bots are budget-capped
  and switchable, never paywalled
  ([ADR 0033](docs/decisions/0033-ai-bot-seat.md)).
- **Generated card art or card text.** Card data and images come from
  Scryfall, credited, within the Wizards **Fan Content Policy** (the S70 terms
  check on [#2985](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2985)).
- **Gating card data.** Scryfall's guidelines allow no paywall and no sign-in or
  chat-server gate in front of card data, and no cropping of the artist credit.
  Guest play stays, and so do the legal notices and the artist credit.
- **Phone-width table layouts.** A tablet is the minimum screen.
- **Other formats** (1v1 constructed, limited, cubes) and native app stores.
- **Scale beyond one VM** before the beta goals call for it.

## 4. Tech stack

A Go server (WebSocket + JSON, server-authoritative), a Svelte 5 + TypeScript
client, SQLite (one file, nightly off-site backup), Scryfall bulk data,
Discord sign-in, and Prometheus and Grafana for monitoring. One VM per
environment: `develop` deploys to cmd-dev, `main` to prod
([docs/environments.md](docs/environments.md)).

## 5. Principles

1. **Automation first, the sandbox as the fallback.** The engine does what it
   can and the table can always take over. A manual action is never removed
   while it is the only way through.
2. **Rules fidelity.** Where the engine automates, it follows the
   Comprehensive Rules (pinned edition, cited by number). When a choice is
   between a faithful answer and a smaller one, take the faithful one. A
   card the engine can't express exactly is left out, never shipped
   stronger than printed.
3. **UX polish is the point.** A feature that works but feels bad isn't done.
4. **Honest coverage.** The catalog, roadmap and deck checks say what is
   automated, partial and manual, and tests keep them honest.
5. **Small and boring operations.** One VM, one database, one deployable per
   service, every limit an environment variable with a safe default.

## 6. Beta goals (four weeks)

The owner set these on 2026-10-13. Each one can be computed
from data the server already keeps: the `games` and `seats` tables (practice
tables are never stored, so they're excluded from all three), the
`cmdctrl_*` metrics, and the `player-report` label on GitHub issues.

| Goal | Target | How it is measured |
|---|---|---|
| **People come back** | ≥ 15 signed-in players a week by week 4, and ≥ 40% of each week's players also played in an earlier beta week | Distinct `seats.user_id` at tables with `games.started_at` in the week; "returning" means also seated in an earlier week. `cmdctrl_users_played{window="7d"}` is the live view of the first number |
| **Games reach an end** | ≥ 60% of started tables end with a result | Tables started in the window with `games.outcome IN ('win','draw')`, over all tables started. The live view is `cmdctrl_games_ended_total{outcome}` against `cmdctrl_games_started_total` |
| **Tables hold up** | ≤ 1 in-app bug report per 10 started tables, with each one triaged within 3 days | Issues labelled `player-report` and `bug` opened in the window, over `cmdctrl_games_started_total` for the window. Reports from cmd-dev land in the same repo, so leave those out |

The beta's four weeks start when the soft-launch post goes up. At the end the
owner reviews the three numbers and decides to widen, hold or fix. That
decision and the next targets go back in this section.

## 7. How sprints relate to the goals

Sprints stay as they are ([docs/sprints.md](docs/sprints.md)): one milestone,
one tracking issue, and exit criteria. During the beta, a new sprint names in
its goal line the beta goal it moves, or says it is upkeep: catalog growth,
hygiene, operations. Player reports and Discord feedback come first in
triage. When a goal misses, the next sprint is planned around it.

## 8. Where the details are

- [AGENTS.md](AGENTS.md): how to work in this repo (branches, commit and PR
  format, tiers, commands, environment variables).
- [docs/sprints.md](docs/sprints.md): every sprint, its scope and status.
- [docs/decisions/](docs/decisions/): the ADRs. Read the relevant ones before
  proposing architecture.
- [docs/environments.md](docs/environments.md): hosts, deploys, secrets,
  backups and capacity.
- [docs/adding-cards.md](docs/adding-cards.md) and
  [docs/engine-seams.md](docs/engine-seams.md): growing the catalog and the
  engine.
- [docs/protocol.md](docs/protocol.md), [docs/lobby.md](docs/lobby.md) and
  [docs/bot.md](docs/bot.md): the wire format, the HTTP API and the bots.
- [docs/monitoring.md](docs/monitoring.md): metrics, dashboards and alerts.
- The public pages: About (`#/about`), the roadmap
  ([ADR 0092](docs/decisions/0092-public-roadmap-and-site-portal.md)) and the
  privacy notice ([ADR 0149](docs/decisions/0149-privacy-notice-and-account-deletion.md)).

## History

C2 started on 2026-04-10 as a private sandbox for one playgroup: at most
eight users, a shared password and no database. The plan was to build a
polished manual table first (Option B) and grow rules enforcement on top of
it from S13 (Option C as the destination). Reusing XMage as the backend
(Option A) was rejected in S01, because its wire protocol is JBoss Remoting
with Java serialization and would have tied the stack to the JVM. A
from-scratch engine as the *starting* point was rejected too, because there
would have been nothing to play for years. The full reasoning, including the
competitive landscape and the original phased roadmap, is in the
[original plan](https://github.com/krakenhavoc/cmd_and_ctrl/blob/ddd802370bb42ee5952d0da21d26f8a5e1607308/PLAN.md)
(see its [§2.2](https://github.com/krakenhavoc/cmd_and_ctrl/blob/ddd802370bb42ee5952d0da21d26f8a5e1607308/PLAN.md#22-why-not-option-a-the-s01-discovery),
[§2.3](https://github.com/krakenhavoc/cmd_and_ctrl/blob/ddd802370bb42ee5952d0da21d26f8a5e1607308/PLAN.md#23-why-not-option-c-straight-custom-rules-engine)
and [§6](https://github.com/krakenhavoc/cmd_and_ctrl/blob/ddd802370bb42ee5952d0da21d26f8a5e1607308/PLAN.md#6-phased-roadmap)).
Since then the engine has taken over most of the rules. Bots, a user
database, Discord sign-in, monitoring and a public beta (S70, S71) followed.
