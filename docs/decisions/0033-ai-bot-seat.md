# ADR 0033 — AI bot seat: legal-move enumeration, virtual seats, tiered policies

**Status:** Accepted · 2026-09-16 (proposed 2026-09-11 in [#286](https://github.com/krakenhavoc/cmd_and_ctrl/pull/286)) · Sprint S31 · Issue [#89](https://github.com/krakenhavoc/cmd_and_ctrl/issues/89)
**Amended:** 2026-09-16 · S31 closeout: accepted as built. §6's no-endpoint tiers are refused, not downgraded (#514), the §7 deck path is corrected, §10's field names are corrected, and the loop guard is descoped (see the amendment at the end) · 2026-09-19 · §1's threat ordering is built as an injected hook and the "choose target" move is refused (#687), and the enumerator casts from every zone at every payable price (#673)
**Amended:** 2026-09-19 · #986: a colour prompt's answers are ordered by the card's declared `ColorPurpose`, by one function in `legal` that the enumerator and the wire projection both call (see the amendment at the end)
**Amended:** 2026-09-19 · #1013: a bot chooses WHICH cards an alternative cost eats — the payments are priced through a second injected hook and the cap rises to three (see the amendment at the end)
**Amended:** 2026-09-19 · #1060 / #735: an optional Policy extension survives being wrapped — a wrapper declares what it wraps and the lookup walks the chain — and a game's model spend is one record, split into deciding and improvising (see the amendment at the end)
**Supersedes:** the architecture section of [S31](../sprints.md#s31--ai-bot-seat-legal-move-enumeration--tiered-policy) (heuristic-only, gated behind S27–S30).
**Depends on:** [ADR 0010](0010-card-effect-catalog.md) (effect specs), [ADR 0019](0019-structured-targeting.md) (target specs), [ADR 0020](0020-activated-abilities.md).

## Context

We want bot seats: any player at an unstarted table can add up to four
of them, and the bot plays the game. S31 already specs this as a
rule-based `HeuristicPolicy` and explicitly defers LLM policies to
"a future sprint" on cost, latency and hallucination grounds. This ADR
pulls the sprint forward, keeps S31's two load-bearing ideas (virtual
seat, swappable policy) and changes the rest.

Three facts about the codebase as it stands drive every decision below.

**1. There is no server-side legal-move enumerator — but most of the
raw material is already there.** Nothing in `actions`, `game` or
`protocol` answers *"what may this seat do right now?"*, and S31's
`Policy.DecideAction(view, legal []Action)` signature assumes such a
function exists. It does not.

What *does* exist, already computed server-side and already stamped
onto the filtered view: `CardView.LegalTargets`
(`protocol.stampLegalTargets`), `CardView.Modes` with per-option legal
sets, `CardView.AdditionalCost` including `SacrificeOptions`,
`CardView.ActivatedAbilities` (with tap/mana/life/sacrifice costs and
`SorcerySpeed`), `CardView.ManaAbilities`, `SummoningSick`,
`Game.AutoTapForCost` as an affordability planner, and
`GameView.PendingChoices` — which already ships a **closed option list
per choice**, i.e. the `choice` move kind is effectively enumerated
today.

The genuinely missing pieces are narrower than they first look: the
timing and priority gate, the land-drop rule, the cross-product
expansion of source × target × mode into discrete moves, and combat.
That is a real sprint of work, not half the project.

**2. The engine is a sandbox with rules grafted on, not a rules
engine.** Roughly 290 cards carry effect specs (289 registered as of
#270 on 2026-09-11; `grep -c 'Register(Spec{'` undercounts — several
cycles register in loops). There is no attachment layer,
so Equipment and Auras are inert (S24). Sagas do not exist beyond a `lore` counter type with no chapter
triggers and no 704.5s state-based action (S27).
Mass-removal primitives are S23. A bot cannot be better than the
engine it plays inside, and the catalog — not the model — is the
binding constraint on play quality.

**3. `server/internal/bot/` is already the Discord bot.** The new
package cannot be called `bot`.

## Decisions

### 1. A server-side legal-move enumerator, shared with the client

New package `server/internal/legal`:

```go
// Move is one fully-specified thing a seat may do right now.
// Params is the exact protocol.ActionPayload params that perform it.
type Move struct {
    Type    actions.Type
    Params  json.RawMessage
    Label   string   // "Cast Lightning Bolt targeting Kess"
    Group   MoveKind // land / cast / activate / combat / choice / pass
    Source  string   // instance ID, when a card is involved
}

func EnumerateFor(g *game.Game, seat uuid.UUID) []Move
```

`EnumerateFor` runs under the room lock and returns every legal move
for one seat, with targets and modes already expanded into concrete
alternatives. A cast with three legal targets is three `Move`s, not
one move plus a promise.

This is the keystone. It gives the bot a **closed list to choose
from** — the model never emits an action, it selects an index — which
is the mechanism that makes an LLM safe here rather than a source of
plausible-looking illegal plays. It also lets `timing.ts` become a lookup instead of a
reimplementation of the rules in a second language — though that
payoff should not be oversold. Of the four predicates in `timing.ts`,
`canActivateLoyalty` and `canPassPriority` have no callers at all,
and `canActivateAbility` is three checks used only by the auto-pass
heuristic. Only `canCastFromHand` is load-bearing, and even it already
reads server-stamped fields for targets, modes and additional costs.
What is genuinely reimplemented in TypeScript is the **timing layer**:
sorcery-speed windows, land-drop gating, split-second, flash, priority
derivation, and type-line classification by string matching.

That timing layer is exactly where the "greyed-out in the UI but
rejected by the server" bugs come from, and one is live today:
`sorcery_speed` ships on the wire and the client never reads it, so a
sorcery-speed activated ability is offered at instant speed and then
refused by the server.

**Enumeration is capped and staged.** Combat is combinatorial — every
attack subset across four opponents is not enumerable. Attack and
block declarations enumerate as *per-creature* moves the policy
composes, not as whole assignments. Cast moves cap target expansion at
`MaxTargetExpansion` (default 12) per source, ordered by the existing
threat heuristics, with the remainder reachable through an explicit
"choose target" move.

> **Update, 2026-09-16 (S31 closeout).** The cap shipped as
> `legal.Options.MaxExpansionPerSource` (default 12, spent per source
> across the whole crossed product), not `MaxTargetExpansion`. **The
> ordering and the "choose target" move were not built.** Candidates
> arrive in `LegalTargetsForEffect` order, and whatever falls past the
> cap is dropped. Threat ranking lives one layer up, in
> `aiseat/heuristic/rank.go`, so it can only rank the moves that
> survived the cap. The owner decided to build the ordering, tracked in
> [#687](https://github.com/krakenhavoc/cmd_and_ctrl/issues/687). Until
> it lands, which targets a bot is offered past the cap depends on
> enumeration order.
>
> **Update, 2026-09-19 (#687).** Built, as an ordering hook
> (`legal.Options.OrderTargets`) the policy supplies rather than a
> scorer inside `legal`, and the "choose target" move is refused rather
> than deferred. See the amendment at the end of this file.

A corollary learned the hard way in #544, and applied since to every
new cost component: **a variable in a cost must not become an arity of
that cross product.** `{X}` — on a spell, or since the S22 work on an
activated ability — collapses to a single value (the largest
affordable, at or above the cost's printed floor) and consumes none of
the expansion budget. Enumerating a range instead would spend the
whole per-source budget on near-identical moves pointed at the first
target and never offer the second, which is precisely the shape #544
found in the library-search walk.

The same corollary covers a **rule about a set of cards** on a prompt:
a search's `SearchLibrarySpec.Validate`, and since #624 a choose-cards
prompt's `ChooseCardsPrompt.Validate`. The enumerator doesn't
re-derive the rule. It asks the engine's own acceptance check
(`SearchPickLegalLocked`, `ChooseCardsPickLegalLocked`) before it offers
a set. It also spreads the per-source budget across set sizes
(`filteredCombinations`). The spreading matters as much as the check.
If singles filled the budget first, a "one creature card, or two cards"
prompt over a hand with no creature would list nothing at all, and a
seat that owes a choice with an empty list is stuck.

### 2. Bot = in-process virtual seat

New package `server/internal/aiseat`. One goroutine per bot seat,
living beside `ws.Room`:

- `Room` grows an observer list, notified **after** `r.mu` is
  released. Notification is an edge-trigger on a capacity-1
  non-blocking channel; the runner then re-reads `Room.Snapshot()`.
  No payload ordering, no backpressure, no chance of a bot deadlocking
  the room by acting from inside the lock it is being notified under.
- On wake the runner asks: do I hold priority, do I owe a
  `PendingChoice`, am I in a mulligan window, am I being asked to
  declare blockers? If none, sleep.
- Decisions emit through `Room.Apply(seatID, fn)` calling
  `actions.Dispatch` — the same path a WebSocket frame takes. Zero
  protocol forking, undo history and replay logging come free.

Rejected: a separate service playing over WebSocket. It makes
hidden-information cheating structurally impossible, which is
genuinely better, but costs a seat-token issuance path, reconnect
handling and a second deployable on the VPS. We take the type gate in
decision 3 instead and keep the option open — the policy interface is
network-shaped on purpose.

### 3. `Policy` takes a view and a move list, never a `*game.Game`

```go
type Input struct {
    View    protocol.GameView // filtered exactly as the seat sees it
    Seat    uuid.UUID
    Moves   []legal.Move      // the closed list
    Oracle  OracleLookup      // oracle text for names visible in View
    Log     []protocol.LogEvent // public game log — see decision 4
}

type Decision struct {
    Index  int    // into Input.Moves
    Say    string // optional chat line
    Reason string // logged always, surfaced behind a setting
}

type Policy interface {
    Name() string
    Decide(ctx context.Context, in Input) (Decision, error)
}
```

The hidden-information guarantee is a **type gate**, enforced by an
import test: package `aiseat/policy` and everything under it must not
import `internal/game`. A policy physically cannot read another
seat's hand because it has no handle that exposes one. A golden test
asserts `Input.View` is byte-identical to what a human client at that
seat receives.

`Decision.Index` — not `Decision.Move` — is deliberate. A policy
cannot return a move that was not offered.

> **Correction, 2026-09-17.** `Input.Oracle` and `Input.Log` above
> never shipped. `aiseat.Input` is `{View, Seat, Moves}`. The public
> game log decision 4 argues for did land, but on the view, as
> `GameView.Log` — a policy reads it there and nothing has to hand it
> a second copy. Oracle text reaches a policy the other way round,
> through `model.Config.Oracle` injected by `tiers.Options` at seat
> time, which keeps the lookup off the interface every policy has to
> satisfy and out of the type gate's way. Neither field is coming
> back. See [ADR 0052 decision 6](0052-bot-decision-harness-and-eval.md).

### 4. A public game log is a prerequisite, not a nice-to-have

**There is no player-facing, per-event, semantic game log.** Three
things exist and none of them is one:

- `Room.appendReplayLocked` writes a JSONL replay per game, served at
  `GET /games/{id}/replay` with a download link in the lobby — but it
  is a stream of **full snapshots**, so "what happened" has to be
  diffed back out of it.
- The client holds a 200-entry ring buffer of raw protocol frames
  (`ws.ts`), which `bugReport.ts` already folds into bug reports.
- `PlayerView.LifeHistory`, `CommanderCasts`, `MulligansTaken`.

None of that reaches `GameView` as events, and none of it is
something a policy — or a player — can read as a history.

A policy handed only a snapshot cannot know that the seat to its left
wiped the board last turn, that this creature has attacked it twice, or
that an opponent held up two blue mana and did nothing. Those are not
refinements — they are most of what a Commander player is reasoning
about, and their absence caps play quality far below what the model is
capable of.

So `protocol.LogEvent` and a bounded public log on `GameView` land as
part of this work. It is independently worth having: a game log is a
standing gap in the client, and bug reports get materially better when
one ships alongside the replay.

Scope it narrowly — zone changes, casts, resolutions, combat
declarations, life changes, step boundaries — public information only,
filtered through the same visibility rules as everything else, capped
at a few hundred entries.

### 5. Three-layer decision funnel

The cost model only works if most windows never reach a model.

**Layer A — rules (free, sub-millisecond).** Resolves any window where
the choice is forced or trivial: one legal move, or only `pass`
available; an opponent's step with an empty stack and nothing
castable at instant speed; untap/upkeep/draw with no options. In a
four-player game this is the overwhelming majority of priority
windows.

Note that **S13.6's smart auto-pass is client-side** (`priority.ts`,
`settings.ts`) — it is a convenience that suppresses prompts in one
browser, not a server behaviour. The server hands a bot every window a
human would be offered, so Layer A has to do the entire job itself
rather than inheriting any of it.

**Layer B — heuristic (free, deterministic).** S31's scoring function,
scoring each `Move` by `Δscore` on a cloned game. This is also the
`heuristic` tier on its own, and the fallback whenever a model call
times out, errors, or returns an unusable answer. Building it is not
optional: it is the safety net under the whole feature.

**Layer C — model.** Cheap/fast model for routine decisions, frontier
model on escalation. Escalation triggers:

- something on the stack targets the bot or its permanents
- declare attackers, declare blockers
- removal or a counterspell is available and an opponent's board
  crosses a threat threshold
- top-two heuristic candidates within ε (the heuristic is admitting
  it does not know)
- modal spells, X costs, `pick_target` choices with >4 candidates

The static half of the prompt — rules primer, the bot's own decklist
with oracle text, archetype plan — is prompt-cached and stable for the
whole game. The per-decision delta is the board state and the move
list.

**Order-of-magnitude estimate**, to be replaced with measurements from
the first bot-vs-bot run: ~30–50 real decisions per bot per game after
Layer A, ~20% escalating. Three bots ≈ 100–150 cheap calls and ~30
frontier calls per game. That is cents per game, and it is the number
the design should be checked against — if Layer A is not absorbing
>80% of windows, the funnel is broken and should be fixed before
reaching for a cheaper model.

> **Update, 2026-09-11 (sub-PR 7 as built).** Both numbers are now
> measured rather than estimated, from four-bot games against the real
> engine (`aiseat/funnel_game_test.go`, `aiseat/model_game_test.go`).
>
> **Layer A absorbs ~90%** of priority windows — 4,753 of 5,207 over
> the two games the gate runs by default, and 7,036 of 7,810 over
> three — comfortably past the 80% floor. The distribution is
> the interesting part: `mana-only` alone is 51.4% (the enumerator
> offers a mana ability for every untapped land in every window where
> the seat holds priority, and floating mana is never the play because
> casts auto-tap) and `forced` is around 38% (the pass-only window,
> which is most of what a four-player table consists of). The land
> rule is half a percent. Three rules, and two of them do all the work.
>
> **Escalation is ~60–70% of surviving windows, not ~20%.** The
> estimate above was wrong by a factor of three, and it is worth
> recording why rather than quietly retuning: the combat trigger fires
> on every attack and every block declaration, and bots declare a lot
> of combat. The trigger list in this section is the specified one and
> the implementation follows it; whether a frontier call is worth it
> on the third chump-block of a turn is a tuning question that wants
> real play data, and every threshold lives in `model.Config`.
>
> **What is not measured is per-game spend.** That needs a live
> endpoint, and there is no API key in CI. The plumbing is proven
> offline against a fake — prompt assembly, cheap/frontier selection,
> the deadline arithmetic, index validation, and the fallback under
> every failure — and the funnel's shape is what the numbers above
> describe. Exit criterion 4's "per-game model spend is measured and
> recorded, not estimated" remains open until a key exists.
>
> **Update, 2026-09-19 (#735).** There is now somewhere for that
> number to appear. A game reports ONE spend record when its bots are
> done — per seat, split into deciding and improvising — on the
> runner's `Stats`, as the game's admin log line, and in the decision
> log. The MEASUREMENT still waits on a keyed endpoint; the path no
> longer does. See the #1060 / #735 amendment at the foot of this ADR.

### 6. Tiers are the difficulty slider

| Tier | Layers | Cost | For |
|------|--------|------|-----|
| `random` | uniform over `Moves` | none | engine fuzzing, bug-hunt harness |
| `heuristic` | A + B | none | filling seats, fast games |
| `assisted` | A + B + C | low | the default |
| `strong` | A + C, wider candidates, 1-ply sim on top-K | high | solo practice |

Set per bot at add time. `random` is not a joke tier — a four-`random`
table playing unattended is the cheapest rules-engine fuzzer this
project will ever get, and it exists from the first PR.

> **Update, 2026-09-11 (sub-PR 7 as built).** The table shipped as
> written with one exception: **`strong`'s "1-ply sim on top-K" is not
> implemented, and cannot be under decision 3.** Simulating a move
> means cloning a game; cloning a game means holding a `*game.Game`;
> a policy may not hold one, and the import test fails the build over
> it. This is the identical collision sub-PR 6 hit with "Δscore on a
> cloned game" and it resolves the same way — the hidden-information
> guarantee is worth more than the lookahead. `strong` is therefore
> "A + C, wider candidates, frontier model on every surviving window,
> 5s MaxThink", and the row should be read that way.
>
> One thing the table does not say and should: **every tier works
> without a model endpoint.** `assisted` and `strong` with a nil
> Client keep their names and play on Layer A + B, because a VPS with
> no API key has to be a working deployment rather than a broken one.

> **Update, 2026-09-16 (S31 closeout).** The paragraph above is still
> true of the *policy*, and it is what the model-outage drill relies on.
> It is no longer true of the *picker*. Since
> [#514](https://github.com/krakenhavoc/cmd_and_ctrl/pull/514), a
> server with no model endpoint reports `assisted` and `strong` as
> `available:false`, with a reason, and a request to seat one is a
> 422. A seat playing the heuristic under a model tier's name would
> misstate the game the player is in. A seat that is already running
> when its endpoint goes away keeps playing on Layer A + B. #514 also
> added an OpenAI-compatible transport for a self-hosted model
> (`CMDCTRL_OPENAI_ENDPOINT`) beside the Anthropic one. A deployment
> may put one model in both funnel slots, since
> `CMDCTRL_BOT_FRONTIER_MODEL` defaults to `CMDCTRL_BOT_MODEL`. On such
> a deployment `strong` and `assisted` ask the same model. `strong`
> differs only in sending every surviving window as an escalated
> request, over a wider candidate list.

### 7. Bot decks are curated and catalog-only, enforced by a test

`server/internal/decks/` (drafted as `server/internal/aiseat/decks/`;
[#553](https://github.com/krakenhavoc/cmd_and_ctrl/pull/553) moved it
when human seats began picking from the same decks). Every card in every bot deck must
resolve to a registered `effects.Spec`. A test walks the bot decks
against `effects.All()` and fails the build on any gap — without it,
"curated" rots the first time a spec is refactored.

**This constrains archetypes hard, and it should be said plainly.** Of
S31's six proposed decks, Voltron and any Equipment or Aura strategy
are unbuildable until S24 lands the attachment layer; Aristocrats
needs more of S21/S23 than exists; Combo is a bad idea regardless.
Aggro, ramp-stompy, and a thin spell-based control deck are what
today's catalog actually supports. Ship three good decks, not six thin ones.

> **Update, 2026-09-11 (sub-PR 5 as built).** The count was four, not
> three. The principle above held; the catalog moved underneath it.
> **Aristocrats became buildable** — S21's sacrifice outlets and death
> payoffs and S23's board wipes both landed, so mono-black aristocrats
> shipped alongside aggro, ramp-stompy and control. **Voltron and
> Equipment/Aura are still out, but no longer for the reason given
> here**: the attachment layer shipped (#374), and so did the first
> seven attachments (#379, #380). Seven is not a Voltron deck. That
> archetype is now waiting on card count, not on engine machinery, and
> needs no decision to unblock it — just more cards.
>
> One thing this section did not anticipate is worth recording, because
> it cost real cards: **registered is not implemented.** A spec can be
> in the registry with its load-bearing clause declared unmodelled, and
> the coverage test cannot see the difference. Gemcutter Buccaneer and
> Teferi, Time Raveler both pass the test and were both cut by hand.
> The test is a floor; reading the card file is still the job.

Any-deck support — where the bot improvises from oracle text — is a
later tier, gated behind decision 8 being proven in practice.

### 8. Improvisation is allowed, announced, and tagged

When the chosen line needs an effect the catalog cannot execute, the
bot may use the sandbox verbs (`move_card`, `change_life`,
`add_counter`, `mark_damage`) exactly as a human at this table does.
Every improvised change:

- is emitted as one bundle, not a drip of individual mutations
- posts a chat line naming the card, the intended effect, and that it
  was improvised
- is tagged in the replay log so bot improvisations are greppable
  when a game goes wrong

**What this requires of undo** — and the first draft of this ADR got
it wrong, so it is worth stating precisely. `Room.Undo`'s caller gate
is `caller != uuid.Nil && top.caller != uuid.Nil && top.caller !=
caller`, so entries stamped `uuid.Nil` are already undoable by any
seated player. The runner therefore stamps improvisation bundles as
`uuid.Nil` rather than as the bot's seat, and no change to the gate is
needed.

Two real constraints remain:

- A human undoing a nil-stamped entry spends **their own**
  `UndosRemaining` and hits `ErrNoUndosRemaining` at zero. Cleaning up
  after a bot should probably not cost a player their own undo budget;
  that needs a decision.
- `Undo` pops only the **top** entry of one room-wide stack. A bot
  action followed by anything else puts the improvisation out of reach.
  The safety valve is real but shallow — which is another reason the
  improvisation must be one bundle, and why the chat announcement
  matters more than the undo does.

> **Update, 2026-09-11 (sub-PR 8 as built).** Both open points above
> are now settled, and one thing this section did not anticipate
> turned out to matter more than either of them.
>
> **The undo budget: it is free.** `ws.undoEntry` grows a `freeUndo`
> flag, set on improvisation bundles and nothing else, and `Room.Undo`
> skips both the budget peek and the `SpendUndo` debit for a flagged
> entry. The reasoning: the budget exists to police the social cost of
> taking back **your own** move, and it refreshes once per turn. An
> improvisation is a bot asserting a rules interpretation the engine
> could not execute; a human correcting it is doing maintenance on a
> catalog gap, not rewinding their own play. Charging for it would
> make the careful response (read the announcement, check the card,
> put it back) cost more than the lazy one (let it stand) — exactly
> backwards for a feature whose entire safety argument is that it is
> reversible. It also means a player who has already spent their
> take-backs this turn is not stuck with a bot's mistake, which is the
> case where the recourse matters most. The nil stamp already made the
> undo *permitted*; this makes it *affordable*.
>
> **The shallow stack: unchanged, and load-bearing.** `Undo` still
> pops only the top entry, so an improvisation followed by anything
> else is out of reach. Nothing here fixes that, and the section is
> right that the announcement matters more than the undo does. What
> the implementation adds is that the bundle is now genuinely **one**
> entry — `Room.ApplyBundle` runs every verb under a single hold of
> the room lock and rolls the whole thing back from the pre-bundle
> clone if any step fails, so there is no reachable state where half
> an improvisation is applied. `Apply` could not promise that: it
> stashes a pre-mutation clone and drops it on error, which is correct
> for one dispatch and wrong for four.
>
> **What this section understated: the caller-nil stamp is a power
> grant, not just an undo convenience.** `uuid.Nil` bypasses
> `requireCardController` and `ErrPlayerCallerMismatch`, and it has to
> — an improvised removal spell reaches an opponent's board and an
> improvised drain reaches their life total, both of which the seated
> gates exist to refuse. So improvisation is the one path on which a
> bot acts with admin authority. The counterweights are enforced in
> code rather than left to the policy: the verb list is an allow-list
> of the four named here (a bundle containing `concede` or
> `discard_selection` is refused), a bundle that cannot name its card
> and its effect is refused before anything is dispatched, and the
> announcement is built and validated ahead of the commit rather than
> after it.
>
> **The chat line has nowhere to land yet.** S08.5 removed the chat
> UI; the transport stayed live, so the server side of the
> announcement works and the client stores the frame, but nothing
> rendered it. An announcement nobody can see is not an announcement,
> so sub-PR 8 ships a small read-only `BotFeed` in the board's
> attention strip that renders improvisation disclosures always and
> reasoning behind the setting. It is deliberately minimal and is
> subsumed by a real chat panel whenever one returns.

### 9. Lobby integration

- `POST /games/{id}/seats/bot` with `{tier, deck}` — allowed for any
  seated player at an unstarted table, and for admin. Not admin-only:
  the request was "add to any unstarted table."
- `DELETE /games/{id}/seats/bot/{seat}` while unstarted. Shipped in
  sub-PR 4 keyed by the seat's **player UUID** rather than its index:
  removing a bot renumbers every seat behind it, so an index is a
  value the client would have to re-read between reading it and using
  it. The path is `/seats/bot/{player_id}`.
- Bot seats get real decks, so `Lobby.Start`'s `DeckUploaded` gate
  passes without special-casing.
- **Seat arithmetic, stated plainly:** bots occupy real seats and
  `MaxPlayers = 4`, so a seated human can add at most **three** bots —
  a four-bot table has no seated human to have asked for it. Four bots
  is therefore reachable only through the admin path (`POST /games` is
  already admin-only), which is fine: the all-bot table is the test
  harness, not a player-facing flow. One human plus one bot is a legal
  game (`MinPlayers = 2`) and is the solo-practice case.
- `PlayerView` grows `is_bot`, `bot_tier`, `bot_deck`.

### 10. Pacing and hard timeouts

`MinThink` (default 700ms) holds a fast decision so the table does not
feel precognitive. `MaxThink` (2s for `assisted`, 5s for `strong`) is a
hard context deadline; on expiry the runner takes Layer B's answer and
logs the miss. A bot that cannot decide passes. The table never waits
on a model.

> **Update, 2026-09-16 (S31 closeout).** The first draft of this
> section named the fields `MinThinkMs` / `MaxThinkMs`, as millisecond
> integers. They shipped as `MinThink` / `MaxThink`, `time.Duration`
> fields on `aiseat.Config`, and the text above now uses those names.
> A deployment can widen the model tiers' `MaxThink` with
> `CMDCTRL_BOT_MAX_THINK`, but never shorten it. It defaults to 20s
> when a self-hosted endpoint is configured, because a model that
> overruns the deadline plays the heuristic's move under the tier's
> name.

## Consequences

- The enumerator lands before anything visibly bot-shaped exists, which
  is a motivation problem more than a scheduling one. It is smaller
  than it first appears — targets, modes, costs, abilities and pending
  choices are already computed and stamped on the view; timing, land
  drops, cross-product expansion and combat are what is missing — and
  it pays for itself independently by closing the "greyed-out in the UI
  but rejected by the server" bug class.
- The heuristic policy is not optional. It is the fallback for every
  model failure, so the "LLM bot" cannot ship without the "heuristic
  bot" underneath it.
- Play quality is capped by catalog coverage, not by the model. At a
  few hundred cards a bot will be a competent player of a deliberately small
  format. That is the honest expectation to set.
- Model calls put an outbound API dependency and a key on the VPS for
  the first time in a *gameplay* path. Discord, GitHub and Scryfall
  are already outbound but none of them block a game. Layer B's
  fallback is what keeps a model outage from freezing a table.
- Bot-vs-bot games will surface engine bugs nobody hits in human play.
  Capture a replay for every automated run.

## Rejected alternatives

**Propose-and-validate instead of enumeration.** Bot proposes, dispatch
rejects, bot retries. Cheapest to build and the engine stays the
authority — but an LLM burns calls on rejected moves, and losing the
closed list means losing the one structural defence against invented
actions. Rejected.

**Heuristic-only, S31 as written.** Deterministic, free, testable. It
is also the option the roadmap gates behind S27–S30 for good reasons,
and at a few hundred cards a pure heuristic has little structured signal to
reason over. The hybrid gets a playable bot sooner and degrades to
exactly this when the model is unavailable.

**Local model on the homelab.** Zero marginal cost and no key on the
VPS, but it makes a deployed game server depend on a home network, and
MTG reasoning is where small models are weakest. Revisit if the API
bill ever becomes a real number.

**LLM on every window.** No rules layer, model decides everything.
Simplest to reason about, and roughly an order of magnitude more
expensive and slower for decisions that are not decisions.

## Amendment (2026-09-16): S31 closeout

S31 shipped every sub-PR this ADR describes, and the ADR is accepted
as built, with the inline updates above. The owner's decisions are
recorded on [#89](https://github.com/krakenhavoc/cmd_and_ctrl/issues/89).
[docs/sprints.md § S31](../sprints.md#s31--ai-bot-seat-legal-move-enumeration--tiered-policy)
has the item-by-item reconciliation. Four points fall outside the
sections above:

- **The loop guard is descoped.** S31's original list asked for a
  runner rule: the same non-pass move twice in one priority window
  forces a pass. It was never built. The runner's only guard is keyed
  on rejection (`MaxConsecutiveRejects`), not repetition. The literal
  rule would block legal repeats, such as activating the same pump
  ability twice in one window, and the random, heuristic
  and catalog soaks show no livelock. Loop breaking is tracked in
  [#628](https://github.com/krakenhavoc/cmd_and_ctrl/issues/628).
- **§8's improvisation path is live in the runner but unused in
  play.** The bundle, the validated announcement, the replay tag and
  the free undo all shipped and are tested. No shipped tier implements
  `aiseat.Improviser`, though, so no bot improvises yet. The owner
  decided to build one for the model tiers, tracked in
  [#686](https://github.com/krakenhavoc/cmd_and_ctrl/issues/686).
  **Closed 2026-09-19:** `*model.Policy` implements the hook, so
  `assisted` and `strong` improvise. See the #686 amendment at the
  foot of this ADR for the windows, the prompt, the budget and the
  failure posture.
- **§5's per-game model spend is still unmeasured.** That update's
  last paragraph still stands. Absorption is measured at ~90%, but
  measuring spend needs a game against a hosted endpoint. Split out
  as [#735](https://github.com/krakenhavoc/cmd_and_ctrl/issues/735).
  **Half-closed 2026-09-19:** the measurement is built and reported
  once per game; the number still needs a key. See the #1060 / #735
  amendment at the foot of this ADR.
- **Rejected alternative "Local model on the homelab".** #514 made a
  self-hosted OpenAI-compatible endpoint an operator option beside the
  hosted one. The transport is a deployment setting, and nothing in
  the funnel depends on which one is used. The objection recorded
  there still applies to any deployed server pointed at a model on a
  home network.

## Amendment (2026-09-18, #810, #619): the enumerator's X rule

§1's contract is that every enumerated move is one `actions.Dispatch`
accepts. Two bugs from the catalog soak sit either side of it: one where a
move was legal, accepted, and still not worth offering, and one where the
contract was simply broken.

**1. A zero-effect X=0 move is not a move (#810).** CR 601.2b / 602.2b make
X a number the caster announces and CR 107.3 leaves 0 legal wherever the
printed text sets no floor, so the engine accepts Soothsaying's "{X}: Look
at the top X cards of your library" at X=0 and is right to. It costs
nothing, does nothing, and is back on the move list the moment it resolves —
CR 732.2a's repeatable sequence of optional actions, which no player is ever
made to keep repeating. A table of bots applied 79,519 actions in five
minutes on turn 18 without the turn advancing.

The rule is **one function**, `enumeratedXFloor` in
`internal/legal/x.go`, shared by casts and activations because the mistake
was identical on both sides: the smallest X the enumerator will announce for
a cost carrying an {X} slot is 1 when the whole effect scales with X, and
where that cannot be paid the move is not offered at all — the treatment
Helm of Obedience's printed "X can't be 0" already got. Nothing about
legality changes, and no card is special-cased.

*"The whole effect scales with X" is declared, not guessed.*
`effects.Spec.XMatters`, read through `game.XMattersFor`, is the catalog's
one-bit answer to a question nothing else can derive. A card with a fixed
RIDER — The Goose Mother's 2/2 flying body, Springleaf Parade's mana static
— leaves it unset and keeps its X=0 offer, which is then made exactly when
nothing larger is affordable, because the search takes the largest payable
X. A source scan in `effects/x_matters_guard_test.go` fails the build on a
Spec that reads `ctx.X()` without declaring, with an allowlist for riders.
Thirty-four of the catalog's forty-two X cards declare; eight are riders.

*Two gaps, both "the enumerator is not choosing X here".* An X announced by
a cost that is not the mana cost is still announced as 0, because nothing in
`internal/legal` prices such a cost: Toxic Deluge is offered at X=0 and
sweeps for -0/-0. (Waterbender's Restoration is the same shape and lands on
the other side of it — its clause counts FROM X, so rule 2 below declines to
enumerate it at all.) And `CountFromX` on an activated ability has no engine
support to mirror.

**2. An X-defined target count ties X to the targets (#619).** Crackle with
Power's target count *is* X, and the enumerator offered one target with X=0
— a cast `cast_spell` refuses outright, which is a soundness break, not a
quality one.

An announcement is a list of target STEPS since #937
(`game.AnnouncedClauses`), so the rule is written over those: a step whose
clause has `CountFromX` is opened to 1..(largest payable X) rather than to
its printed count, and the X the move announces is read back off the set —
how many refs carry that step's (Mode, Slot). Two X-counted steps that
disagree are not a set any announcement could cover, so they are dropped.
The expansion is bounded by `MaxExpansionPerSource` like every other. It is
the mirror image of the engine's own `resolveStepCountsFromX`, which pins
the same clauses to an X already announced — the difference between
validating an announcement and building one.

The floor is 1, not 0: X=0 means zero targets and a spell that does nothing,
so rule 1 covers it and no separate case is needed. A `CountFromX` clause
whose X comes from a cost this package cannot price (Waterbender's
Restoration's waterbend) is not enumerated at all, because the only
announcement it could make is the one the engine refuses.

**The bot's threat ordering (§1) is still not built** (#687). Unchanged.
_(Built on 2026-09-19; see the last amendment in this file.)_

## Amendment (2026-09-18, #957): a pay-life X is priced like a mana X

The first of the two gaps above is closed for the "pay X life" half.
X is ONE announced number (CR 601.2b) and the mana cost is not the only
thing that can charge for it: Toxic Deluge prints `{2}{B}` with no `{X}`
anywhere in it and announces X by paying X life
(`game.AdditionalCost.PayLifeX`). The affordable-X search reads `{X}`
slots, so it answered 0 for the card and always would — the enumerator
offered the cast at X=0 and the board was swept for -0/-0, a zero-effect
move of rule 1's exact shape arriving through a different seam.

**One more line in the one X rule**, `xCeilingFromCost` in
`internal/legal/x.go`, and it is keyed on the COST COMPONENT rather than
on the card: any future "as an additional cost to cast this spell, pay X
life" is priced by it the day the card is registered, with no catalog
entry and no per-card branch.

- The **floor** is `enumeratedXFloor`'s, unchanged. Toxic Deluge declares
  `XMatters`, so it is 1, and a seat that cannot reach 1 is offered no
  cast at all — the Soothsaying treatment, one cost component over.
- The **ceiling** is `life - 1`. CR 119.4 permits paying exactly your life
  total and the engine accepts that (`xValue > p.Life` is the announce
  check); this package will not OFFER it, on rule 1's own footing of what
  is worth putting in front of a player. A sweep that kills its own caster
  is not the pricing to hand a bot as the card's only offer.
- `Options.MaxX` caps it like every other X search, and the announcement
  is still exactly one number, so X never enters an expansion cross
  product. (The issue said `MaxExpansionPerSource`; that one caps the
  cross product, which an X that is a single value never joins.)

A cost with BOTH an `{X}` in its mana cost and a pay-X-life would take the
smaller of the two ceilings — `announcedX` is written that way — though no
such card is registered. Waterbender's Restoration's waterbend
`TapPermanentsCost` is still unpriced and still not enumerated, for the
reason rule 2 gives.

## Amendment (2026-09-19, #673 / #687): the enumerator's zones, its prices, and §1's threat ordering

Two of §1's promises come good at once, and they share a paragraph
because they share a function: `legal.castMoves` and the cap it spends.

### §1's threat ordering is BUILT, and it is a hook

The 2026-09-16 update above recorded the ordering as not built and
named #687. It is built now, as an **injected hook** rather than as a
scorer inside `internal/legal`:

```go
// legal
type TargetCandidate struct { ID uuid.UUID; Player bool }
type TargetOrder func(c TargetCandidate) float64
type Options struct { …; OrderTargets TargetOrder }

// aiseat — an optional Policy extension
type TargetOrderer interface { TargetOrder(in Input) legal.TargetOrder }
```

The issue offered two shapes: move a neutral scorer down into `legal`,
or take an ordering through `Options`. The hook wins for a layering
reason and a correctness one. **Layering:** `aiseat/heuristic` imports
`legal`, so `legal` can never import it back; a scorer moved down would
have to be a second one, and two scorers that are supposed to agree
about a board eventually will not. **Correctness:** the heuristic's
scoring reads `protocol.GameView`, the seat's own FILTERED projection,
which is where the hidden-information guarantee (§3) lives. A scorer
inside `legal` would read the authoritative game, and the first fact it
wanted that the view redacts would be a leak.

So `legal` states the rule and asks for an order; the policy supplies
one built from the same `Weights` it prices every other decision with —
`Threat` for a seat, `boardValue` for a permanent. `Options.OrderTargets`
nil is the whole of the old behaviour, which is what every non-bot
caller (the view's move stamp included) gets.

**Ordering is by importance, not by desirability.** The hook does not
know which spell is being cast, so the biggest objects on the table
survive the cap whoever controls them, and choosing among the survivors
stays the policy's job. That is the right split: the hook's only power
is to stop a target being dropped, and dropping the board's biggest
permanent is wrong for every spell.

**The "choose target" move for the remainder is REFUSED, not deferred.**
§1 offered it as the alternative to a documented guarantee. It cannot be
built without breaking the package's contract: every `legal.Move` is a
wire `ActionPayload` that `actions.Dispatch` accepts as-is, and "open a
target picker" is not an action the dispatcher has — it is a client
affordance. A move kind the dispatcher refuses would put a permanently
rejected entry in every bot's move list, which is the #544 stall the
contract exists to prevent. The guarantee instead: **the top
`MaxExpansionPerSource` candidates by the seat's own ordering are
enough**, because the ordering is the seat's own scorer and a target it
ranks below twelve others is a target it would not have chosen.

### Casting from a zone that is not the hand (#673)

The enumerator walks all five cast surfaces (hand, command, graveyard,
exile, the top of the library) and offers **one move per price the
engine would accept**, through one new engine function:

```go
func (g *Game) CastOffersForLocked(playerID uuid.UUID, card Card,
        zone ZoneKind, grant *CastPermission) []*AlternativeCost
```

A nil entry is the printed mana cost, present only when
`validateCastPathLocked` allows a claim of nothing; the rest are the
card's own zone-bound offers and the one a grant synthesises, in
announce precedence, filtered by the same path gate the cast uses and by
`AlternativeCostPayableLocked`, the #695 predicate the view's offer
stamp and `CastSpell`'s validator also read: the offer's Condition, CR
119.4's life, CR 601.2b's card component — mana deliberately not
asked, because CR 601.2g lets the caster tap afterwards.

The card component is the combination search the issue is about, and it
is **capped at one payment** (`legal.maxEnumeratedCostPayments`). That is
this section's own corollary applied to a new kind of variable: a
variable in a COST must not become an arity of the target cross product.
Escape-five over a twenty-card graveyard is 15,504 payments, all of them
the same spell with the same targets, and the policy cannot tell them
apart because it does not price a card in a graveyard at all. The policy
is written down beside `maxEnumeratedRepeats` in `docs/bot.md`.

> **Update, 2026-09-19 (#1013).** The cap is THREE, the payments are
> priced rather than indistinguishable, and the extra ones spend no
> target budget. See the amendment below.

The faces of the cast come from the same pair of engine functions:
`Card.CastableFaces` (CR 715.3's adventure choice since #719, and a
modal DFC's two halves), NARROWED by `CastPermission.Faces` when a
grant names any — CR 715.4's "cast the creature from exile", a
defeated Siege's back. That is `faceForCastLocked`'s own rule, so the
enumerator cannot offer a half the announce path refuses, and madness
(#657) needed no enumerator code at all: it is an exile grant with a
key and `TimingFlash`, which this walk already reads.

Deliberately still not enumerated, each for a stated reason: two
different optional costs at once (no card offers it), and two
card-shaped sacrifice clauses on one cast (same).

## Amendment (2026-09-19, #986): the enumerator orders a colour prompt by its purpose, and so does the wire

§1 says the enumerator produces "a stable, ordered move list", and
the #673 / #687 amendment above added `Options.OrderTargets` so that a
seat's own policy could decide which candidate TARGETS survive the
expansion cap. A colour prompt's answers needed the opposite shape and
this records why.

**The gap.** `enumerator.colorAnswers` ordered every `choose_color`
prompt (CR 105.4) the same way — by how many permanents the enumerating
seat controls of each colour, most first — whatever the prompt was for.
#780 had already given every such prompt the card's own
`game.ColorPurpose` and PR #985 had taught the `heuristic` policy to
switch on it, so nothing played badly at the `heuristic` tier. Three
things still read the raw order: the `random` tier, the `decideChoice`
tie-break (which takes the first offered answer, and is the path a
purpose-less prompt lands on), and the human's colour buttons, which
render `color_options` in the order the server sends. All three got the
Coldsteel Heart answer for a Wash Out.

**The decision.** One exported ordering function in `legal`,
`OrderColorOptionsLocked` (`server/internal/legal/color_order.go`),
with one arm per purpose, called by BOTH readers: `colorAnswers` for
the move list and the `choose_color` projection in
`protocol.ViewOfGame` for the wire. So a bot's first offered answer and
a human's first button are the same colour, and neither side holds a
ranking rule of its own.

- `harm` ranks by what the OTHER seats lose net of what the chooser
  loses; `filter` by what the other seats have, with the chooser's own
  board not a term at all; `protect` by the greatest POWER among the
  creatures other seats control of that colour; `mana`, `benefit` and a
  prompt that declares nothing keep the pre-#986 rule exactly.
- Battlefield counts only. §3's hidden-information boundary is why: the
  order rides the wire to every viewer, so every term in it has to be
  something every viewer can already count.
- An ORDERING, never a filter. CR 105.4 makes all five colours legal
  and a narrowed printed list ("a color other than blue") stays exactly
  as narrow as the card made it. The sort is stable, so ties keep
  WUBRG.

**Why a function and not an `Options` hook, unlike #687.** Ranking a
board by what a spell is worth against it is a POLICY question, and
`legal` may not import `aiseat`, so target ordering had to be injected
by the seat. A colour prompt's order is a reading of the card's own
printed text against public counts; it must be identical for a bot and
for a human, because both read it off the same prompt; and `legal` is
the layer both already go through. A hook would have left the human's
buttons unordered — which is half of what #986 was filed for.

**What it is not.** It is not a second scorer. The `heuristic` still
prices each answer with its own `Weights` (`colorChoiceValue`), and
this decides only what order it sees them in — and therefore what it
does when it scores two of them the same. The arms are deliberately the
same questions that function asks, answered with the crudest public
proxy there is.

## Amendment (2026-09-19, #1013): a bot chooses WHICH cards an alternative cost eats

*Amends §1's corollary and the 2026-09-19 (#673 / #687) amendment above.
Closes [#1013](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1013).*

**1. The cap was never the constraint; the missing evaluation was.**
`maxEnumeratedCostPayments = 1` was justified by §1's corollary — a
variable in a cost must not become an arity of the target cross product
— and the corollary is right. But the sentence that made it
UNAVOIDABLE was the next one: "the policy cannot tell them apart
because it does not price a card in a graveyard at all". Widening the
search without an evaluation would have spent the budget ranking
payments by index. So the order of work was fixed, and the issue said
so: price the cards first, then widen the search.

**2. The price: `aiseat.CostFuelPricer`, and `TargetOrderer`'s shape.**
An optional Policy extension, asked once per decision, threaded into
the enumerator as `legal.Options.OrderCostFuel`. A policy with no
opinion gets `AltCostCandidatesLocked`'s zone order and the enumeration
is byte-identical to what it was, which is what every non-bot caller
gets.

It is a SECOND hook rather than a second meaning for `OrderTargets`,
and `legal.CostFuelOrder` is a second type despite the identical
signature, because the two ask **opposite** questions about different
objects: a target order ranks the board by importance and the
enumerator keeps the top of it; a fuel price ranks a seat's own cards
by what it would LOSE and the enumerator spends the bottom. One
interface with one method would let a policy answer one with the other,
and neither mistake — pitching your best card, targeting your worst —
would fail to compile.

**3. One function, two readers.** `heuristic.fuelValue`
(`aiseat/heuristic/fuel.go`) is read by the enumerator to decide which
payment to OFFER and by `valueOfCast` to price the payment it was
offered. That is the whole reason the ordering is injected from up here
rather than written in `legal`: a second scorer would eventually
disagree with the first about what an escape costs, and the bot would
choose a payment it then priced as a mistake. It also closes #673's
declared gap — a non-hand cast still costs no card in HAND, and what it
really spends is now priced rather than free.

The three cases, and the tunings, are tabulated in `docs/bot.md`. The
one judgement worth recording here is the split inside a graveyard: a
card the seat can still CAST from there (escape, flashback, a granted
impulse) is a discounted card in hand, and an idle one is a floor —
`FuelFloor` for a land, `FuelIdle` for anything else. Neither floor is
zero, because a card nobody can use today is still one tomorrow's delve
might want, and zero would make every unreadable card the first thing a
cost ate. "Is this a cast surface" is read off the view's own stamps
(`castable_here`, the offer list) rather than re-derived, because §3
says a policy may not import `internal/game` and a second reader of the
cast gate is exactly what that rule is for.

**4. The corollary is enforced, not restated.** The extra payments are
offered in a SECOND PASS, out of whatever `MaxExpansionPerSource` the
target walk did not use, against the first announcement it made — the
same spell with the same targets at a different price, which is what
they all are. An Uro with no targets has eleven unspent and gets its
alternatives; a removal spell with an escape cost over a wide board
spends its budget on targets and gets none. So a cost variable still
cannot displace a target, which is the corollary's actual content, and
the cap can rise without weakening it.

`maxEnumeratedCostPayments` is THREE. A policy gets a small choice
rather than a decree, the number does not depend on the size of the
graveyard, and the enumeration stays deterministic and stable — the
sort is stable, so equal prices keep zone order and two enumerations of
one board agree.

**Deliberately still not done.** The fuel price does not know what the
SPELL being cast is, the same limit `TargetOrder` declares: it cannot
prefer to pitch a card the spell would rather have in the graveyard,
and it cannot see a graveyard synergy the wire does not carry (a
Crucible on the board makes a land in the graveyard worth keeping, and
`FuelFloor` does not know). Ranking by what the seat LOSES is the whole
of its power, and losing least is right for every cost.

## Amendment (2026-09-19, #938): §2's subscription is taken by `Start`, not by the loop

§2 says the runner is notified through an edge-triggered, capacity-1
channel and re-reads the room on wake. It did not say **when** the seat
joins that observer list, and the answer was "whenever the runner
goroutine is first scheduled" — which nothing orders, because the
caller does not schedule it.

`aiseat.Start` now calls `room.Subscribe()` itself, before it returns,
and hands the channel to the loop. "Start returned" therefore means
"this seat is listening from now on", which is the only ordering a
caller can establish at all.

The gap it closes is a wake, not a decision. The runner's first look is
at the live game rather than at a notification payload, so it still sees
the *effect* of a commit it was never told about — §2's "no payload
ordering" is what makes that true, and it is why this was latent for a
sprint. What a lost wake costs is the seat's next turn to act: a runner
that wants nothing from a window returns and parks until the next
commit, so a commit that landed inside the window is one it will never
be woken for. At a table where the last seat is bot-seated while another
seat's move is in flight, that bot can park for good.

No change to the notification model, the channel, or `Room`. The
edge-trigger, the non-blocking send and the "re-read, trust no payload"
rule are all exactly as decided. The only thing that moved is which
goroutine registers.

## Amendment (2026-09-19, #686): §8's improviser is built, for the model tiers

§8 specified improvisation and S31 sub-PR 8 built the path — the
bundle, the validated announcement, the replay tag, the free undo.
What it did not build was anybody to walk it: no shipped tier
implemented `aiseat.Improviser`, so §8 described a feature no game
ever reached. #686 supplies the missing half. Here is what it decides,
because none of it follows from §8 on its own.

### The window: a spell of this seat's own that the engine did not run

**Improvisation fires once per card instance, at the moment that
instance leaves the stack**, and only when all of the following hold:

1. the seat itself cast it (`StackItemView.Controller` is this seat),
2. the card was on the stack as a CARD — it is in
   `GameView.Stack.Cards` — which is what confines this to spells and
   excludes an ability whose source sits on the battlefield,
3. its `CardView.Unimplemented` bit is set: the card prints rules the
   engine will not carry out (`game.Unimplemented`, i.e.
   `NeedsEffect && !IsAutoCard`, the same signal the deck-upload
   summary and the stack's `manual` chip are built on),
4. the seat's own deck profile has oracle text for it, and
5. the seat owes no pending choice.

That is the shape the issue calls "a castable card with no catalog
spec". It is deliberately expressed **after** the cast rather than
instead of it, and that is the load-bearing decision in this
amendment:

> **The bot pays for the card through the engine, exactly like a
> human.** The four sandbox verbs cannot tap a land. An improvisation
> that moved the card from hand to graveyard itself and then applied
> the effect would be a free spell — the one thing worse at this table
> than a bot that quietly skips a card. So the ordinary funnel casts
> the card through the ordinary move list, the engine charges the
> mana, the spell resolves into silence, and the improviser then
> supplies the text the engine could not run. This is the sequence a
> human on this server already performs by hand, and improvisation is
> defined as doing what a human does.

**Windows that deliberately do NOT trigger it**, each for its own
reason:

- **An unimplemented permanent's ongoing, triggered or activated
  abilities.** There is no wire signal for "this trigger should have
  fired", the timing is unknowable from a filtered view, and the
  condition recurs — so the cap on this would be a rate limit rather
  than a rule. One improvisation per card instance, at resolution, is
  a bound the table can check against the card.
- **Any card an opponent controls.** A bot improvises its OWN spell's
  text. An improvisation reaching an opponent's board is a consequence
  of the bot's card (that is why §8 grants the nil caller); it is
  never the bot correcting somebody else's.
- **A window in which the seat owes a pending choice** (#1000's
  `OwedInStep`, #1016's `GuardsStackItem`, and every other
  `PendingChoices` entry addressed to this seat). Those prompts halt
  the table and are answered through the move list. Improvisation is
  never a way out of a prompt.
- **`random` and `heuristic`.** Neither has a model, and a rule-based
  policy cannot read oracle text. They keep §8's old behaviour, which
  is that nothing improvises.
- **A seat with no model transport.** `assisted` and `strong` are
  already reported unavailable without one (`Factory.TierStatus`); a
  seat that loses its transport mid-game falls back exactly as Layer C
  does — no improvisation, no announcement.

### The prompt: the card's oracle text and the seat's own filtered view

The static, cache-marked half is a primer naming the four verbs with
their exact wire params, the twelve-step cap, and the answer shape.
The per-call half is:

- the card — name, mana cost, type line, **oracle text** — read from
  the seat's `DeckProfile`, which is configuration the seat was
  constructed with, not something read off the board;
- where the card is now (its owner's graveyard, or the battlefield);
- the board as `aiseat.Input.View` renders it, **annotated with the
  instance IDs and player IDs the bundle is allowed to name**.

It is the same hidden-information guarantee as Layer C and for the
same structural reason: `aiseat/model` may not import `internal/game`,
and every byte of this prompt comes from the seat's own filtered
`protocol.GameView` plus its own decklist. Nothing in this package can
reach an opponent's hand, so nothing it sends to a model can leak one.
The IDs are new in a prompt and are not new information: every one of
them is already on the wire to this seat.

### The verb list stays closed, and stays enforced in one place

The model is told `move_card`, `change_life`, `add_counter`,
`mark_damage` and nothing else, and a bundle naming anything else is
**not filtered out in the model package**. It is handed up and refused
by `Improvisation.Validate` — the rail §8 already built and the tests
already pin. One enforcement point, on the path every improviser must
cross, is worth more than a second copy of the allow-list in the one
policy that happens to exist today.

What the model package does check before handing a bundle up is only
what it can check better than the rail: that the reply was JSON, that
it had a step list, and that the step list was not empty. An **empty
step list is a first-class answer** — it is how the model says "this
spell did not actually resolve" (it was countered, its target is gone)
— and it produces no bundle, no announcement and no change.

### Model, budget, and a hard cap per game

- **Model.** The tier's FRONTIER profile, not the routine one. Writing
  a bundle from oracle text is the hardest thing a seat is ever asked
  to do and the rarest; spending the cheap model on it to save a
  fraction of a cent on a handful of calls per game is a false
  economy. `strong` raises its effort exactly as it raises the
  frontier model's.
- **Budget.** The same arithmetic as a decision, because it is the
  same table waiting: the runner now imposes `Config.MaxThink` on the
  `Improvise` call (it previously imposed nothing, because nothing
  implemented the interface), and the funnel's `Reserve` /
  `MinBudget` / `MaxCall` clamp inside it. A call that overruns is a
  dropped improvisation, never a late one. `MaxTokens` is larger than
  a decision's — a bundle is a paragraph, an index is a number.
- **A hard per-game cap.** `MaxImprovCalls`, default 8, counted per
  seat for the life of the seat, on CALLS rather than on applied
  bundles: the cap exists to bound spend, and a call that failed cost
  the same as one that worked. Together with one-shot-per-instance it
  is a closed bound on what improvisation can cost a game, which is
  what [#735](https://github.com/krakenhavoc/cmd_and_ctrl/issues/735)
  needs to measure against. `Improvise: false`
  (`CMDCTRL_BOT_IMPROVISE=0`) turns the whole thing off.

### Failure posture: dropped, and the table is told it was dropped

A bundle that fails `Improvisation.Validate` is dropped — nothing
dispatched, nothing committed — and **the table gets a
`bot_improvisation` chat line saying so**, naming the card and saying
that nothing was changed. That is new. §8's refusal was silent, and
silence is the failure this whole section exists to avoid: the table
cannot tell a bot that chose not to act from a bot that could not, and
"could not, and here is the card it was about" is exactly the
information a player needs to apply the card by hand themselves.

Two refusals stay silent, and both are deliberate:

- **A bundle that cannot name its card.** There is no truthful line to
  post. `TestUnannouncedImprovisationIsRefused` pins it.
- **A bundle that `Room.ApplyBundle` rolled back.** It passed
  validation and the engine refused a step, so the board is exactly as
  the policy found it and the announcement would be about a bundle
  that never existed. `TestImprovisationBundleIsAtomic` pins it.

Everything upstream of a bundle — a model outage, a timeout, a reply
that is not JSON, an empty step list — is likewise silent, because
nothing was attempted against the board. It is counted
(`Stats.ByImprov`) and logged.

### One interface change

`Improviser.Improvise` now takes a `context.Context`. It could not
stay without one: it is called on the runner's goroutine, before the
decision, and the only production implementation makes a network call.
§10 says the table never waits on a bot, and a hook with no deadline
cannot honour that. No production type implemented the old signature,
so nothing but one test fixture moved.

## Amendment (2026-09-19, #1060, #735): optional hooks survive a wrapper, and a game's spend is one record

*Amends §1's two ordering hooks, §5's cost argument, and §6's tier
table. Nothing here changes what a tier IS; it changes what a tier
built by the factory can be asked for, and what a finished game
reports about itself.*

### An optional Policy extension has to survive being wrapped (#1060)

§1's threat ordering (#687) and its fuel pricing (#1013) were both
built, both tested, and **dead on every seat the lobby could create**,
for a month. So was nothing else, but only by luck: the same hole was
one wrapper away from swallowing `Conceder` and `Improviser` too.

The mechanism of the defect is worth recording, because it is a
property of the design rather than a typo. A `Policy` is two required
methods and a growing set of OPTIONAL interfaces — `Conceder`,
`Tracer`, `TargetOrderer`, `CostFuelPricer`, `Improviser`, and now
`Spender` — which the runner and the enumerator find by type
assertion. A type assertion that answers false is indistinguishable
from a policy with no opinion: no error, no log line, no failing test,
and the feature silently does not happen. Every shipped tier is a
policy inside a WRAPPER (§6: `heuristic` is a `rules.Filter` around
the heuristic, `assisted` and `strong` are a `model.Policy` over it),
and a wrapper forwards whichever optional interfaces its author
remembered. The pin tests asserted against a bare `heuristic.New()` —
a policy no seat is ever given — so they passed throughout.

**Decision: a wrapper declares what it WRAPS, once, and the lookup
walks the chain.**

```go
func (f *Filter) Unwrap() aiseat.Policy { return f.Inner }
```

`aiseat.Capability[T]` walks that chain outward-in and returns the
outermost implementer, and every assertion site in the runner, the
enumeration path and the position suite goes through it. Outward-in is
the whole of the semantics and is right in both directions: a wrapper
that implements an extension itself means to OVERRIDE it (`rules.Filter`
is a `Tracer`, and Layer A's own verdict is the one that must reach the
decision log, or §5's absorption rate comes out of the file wrong), and
a wrapper that does not implement it means to be transparent.

**The alternative was a capability struct the factory fills in once**,
and it was rejected for one reason: it has to be edited every time an
optional interface is added, which is precisely the omission being
fixed. A chain covers the interface written next year with no edit to
any wrapper.

The assembly point still carries the guarantee, because that is where
a new layer gets added: `tiers/tiers.go` holds a compile-time
assertion that every wrapper it assembles is an `aiseat.Unwrapper`, so
adding one that hides its inner policy is a build failure. The runtime
half is a table test over `tiers.All()` that builds each tier THROUGH
THE FACTORY and asks it everything the runner will ask it — the test
whose absence was the actual bug.

`DecisionObserver` was checked and is not affected: it is a `Config`
field the Manager sets, never an assertion on a policy.

### A game's model spend is one record, read when the bots are done (#735)

§5's cost argument is an estimate — "~30–50 real decisions per bot per
game, ~20% escalating, cents per game" — and its escalation half has
already been measured wrong by a factor of three. S31's exit criterion
4 asks for the other half to be measured rather than estimated, and it
never was.

The numbers existed and were unreachable: `model.Policy` has counted
its own tokens since sub-PR 7 and its improvisation tokens since #686,
but those counters live inside one policy object that only the seat's
own runner holds, the improvisation half never crosses a decision
window (the improvise path is deliberately unobserved — §8), and
nothing anywhere added four seats together.

**Decision: one `GameSpend` per game, built when every runner has
exited, on three surfaces.** `aiseat.Runner.Stats().Spend` for one
seat while it plays; one INFO line — the game's admin summary — when
the table's bots are done; and one `kind:"spend"` record written into
the decision log just before it closes, for a tool. All three are the
same projection, so there is one definition of what a game cost.

Three things this decides that do not follow from §5 on their own:

- **Deciding and improvising are counted apart and never averaged.**
  §8's amendment bounds improvisation separately on purpose
  (`MaxImprovCalls`, 8 per seat per game) so that it can be reasoned
  about on its own; a decision call is an integer against a cached
  prefix and an improvisation call is a paragraph with an order of
  magnitude more `MaxTokens`. `Spend.Total()` adds them for the one
  question — the bill — that wants them added.
- **An attempted call is spend, however it ended.** Timeouts,
  malformed replies and out-of-range answers are all counted, because
  they were all billed. A measurement that counted only the useful
  calls would flatter exactly the deployment that most needs the
  warning: a self-hosted model missing its deadline on every window.
- **A table that spent nothing still gets its line.** A record that
  appeared only when there was a bill could not be told from one that
  failed to be written, and "the free tiers are free" is the other
  half of §5's argument.

**The number is still pending a keyed run.** Every deployment so far
runs a local model, which has no bill, and CI has no key, so the
figure that replaces §5's estimate cannot be taken here. What is
settled is the path: measured against `model.FakeClient` over whole
four-seat games, every call attributed to one seat and one purpose and
counted exactly once. `docs/bot.md` § "Per-game model spend" says how
to take the measurement and what to record with it. Exit criterion 4
stays open on
[#735](https://github.com/krakenhavoc/cmd_and_ctrl/issues/735) until
somebody runs it against a key.

## Amendment (2026-09-19, #685): S31's bot-table criteria are nightly jobs, and their budgets are one pair of knobs

Two of S31's exit criteria — four `heuristic` bots to a winner within
50 turns over 20 consecutive runs, and zero engine-rejected actions
across a 100-game randomised run — were recorded in `docs/sprints.md`
as met. Neither was running anywhere. The nightly played the heuristic
test's default 3 seeds, and no workflow had ever set
`AISEAT_SOAK_GAMES`, so `TestRandomBotSoak` had executed in CI exactly
zero times. Both numbers were true on the afternoon somebody typed
them and unverified every day since.

They are now the `bot-soak` job in `.github/workflows/e2e-nightly.yml`,
on the same nightly schedule and the same self-hosted runner as
`bot-games`, and both tests are `-skip`ped out of that job's `-race`
step so nothing is played twice.

**No `-race` on this job**, on the argument §5's measurement already
rests on and the random-table step already records: what these 120
tables buy is the *engine* states they reach, and the runner-goroutine
concurrency the detector would watch is the same code every whole-game
test in `bot-games` drives under `-race`, on the same rooms, with the
same observers. `-race` costs roughly 8x here, which is the difference
between a 15-minute job and a two-hour one on a runner shared with CI.

**No network and no model key.** The seats are `heuristic.New()` and
`aiseat.NewRandomPolicy`. §5's funnel, the model tiers and the
improviser are all out of scope for this job by construction, which is
what lets it be a standing gate at all: §5's update already records
that there is no key in CI.

**Every game is seeded and every seed is printed, including on a green
night.** The heuristic gate's seeds are fixed in the test (101..120),
so it plays the same twenty tables every night and a regression there
is unambiguous. The soak's base seed is derived from the UTC date, the
way `TestCatalogSoak`'s is: a fixed soak seed would fuzz the same
hundred tables for ever and stop finding anything after the first
green night.

**The budgets.** Every bot-table test in `internal/aiseat` used to
carry a literal stall detector and a literal wall clock, tuned on an
idle machine. That shape failed three PRs that had touched nothing
near the bot seat (#418, #431, #434) and then the nightly itself
(#600). `AISEAT_STALL` and `AISEAT_WALLCLOCK` now reach all of them —
`playGame` / `playGameIn`, and `TestRandomBotSoak` — alongside the
three tests that already read them. `AISEAT_WALLCLOCK` is a **floor**
on the harness's per-test wall clock rather than a replacement,
because those numbers describe the game being played (120s for a
50-turn table) and an environment variable may raise patience and must
never cut it. A loaded runner should make this job slow, not red.

**A red night opens or comments on one standing issue** (`bug`,
`tech-debt`), scheduled runs only. One issue per night trains everyone
to close them unread; no issue at all is a gate nobody is told about.
