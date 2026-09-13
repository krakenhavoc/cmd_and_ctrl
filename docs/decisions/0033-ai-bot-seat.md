# ADR 0033 — AI bot seat: legal-move enumeration, virtual seats, tiered policies

**Status:** proposed
**Supersedes:** the architecture section of [S31](../sprints.md#s31--ai-bot-seat-heuristic-policy) (heuristic-only, gated behind S27–S30).
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
triggers and no 704.5u state-based action (S27).
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

### 7. Bot decks are curated and catalog-only, enforced by a test

`server/internal/aiseat/decks/`. Every card in every bot deck must
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
- `DELETE /games/{id}/seats/bot/{seat}` while unstarted.
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

`MinThinkMs` (default 700) holds a fast decision so the table does not
feel precognitive. `MaxThinkMs` (2000 for `assisted`, 5000 for
`strong`) is a hard context deadline; on expiry the runner takes Layer
B's answer and logs the miss. A bot that cannot decide passes. The
table never waits on a model.

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
