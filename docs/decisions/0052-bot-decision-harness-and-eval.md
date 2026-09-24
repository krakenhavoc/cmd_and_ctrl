# ADR 0052 — Bot decision harness and eval loop for the local model

**Status:** Proposed · 2026-09-17 · Sprint S31 (follow-up) · Tracking issue [#837](https://github.com/krakenhavoc/cmd_and_ctrl/issues/837)
**Builds on:** [ADR 0033](0033-ai-bot-seat.md) — §3's type gate (a policy
never holds a `*game.Game`), §5's three-layer funnel, §6's tiers, and
§10's `MinThink`/`MaxThink` pacing. Nothing here changes those. This
ADR is about the fact that on the local transport the funnel's Layer C
is currently doing nothing at all, about proving that rather than
assuming it, and about making it do something.
**Related:** [#505](https://github.com/krakenhavoc/cmd_and_ctrl/issues/505)
(latency percentiles), [#735](https://github.com/krakenhavoc/cmd_and_ctrl/issues/735)
(per-game spend — ADR 0033 §5's last open exit criterion),
[Discussion #637](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/637)
(AI-assisted nightly, option A "measure first"),
[#798](https://github.com/krakenhavoc/cmd_and_ctrl/issues/798) and
[#780](https://github.com/krakenhavoc/cmd_and_ctrl/issues/780)
(heuristic choice bugs), [#686](https://github.com/krakenhavoc/cmd_and_ctrl/issues/686)
(Improviser — out of scope here),
[#839](https://github.com/krakenhavoc/cmd_and_ctrl/pull/839) (the
`reasoning_effort` fix, shipped ahead of this ADR's PR sequence).

## Context

`assisted` and `strong` shipped in S31 and were built, start to
finish, against `model/fake.go`. Every test that exercises Layer C
asserts on a fake client's canned reply. That was the right way to
build the plumbing — prompt assembly, profile selection, deadline
arithmetic, index validation and the fallback under every failure are
all proven — but it means one thing has never been true of this
feature: **nobody has looked at a real prompt, a real reply, or a real
game played by a real model.**

The intended deployment is a local one. `CMDCTRL_OPENAI_ENDPOINT`
points at Ollama on the homelab box, running `qwen3:14b` on a 16 GB
card. ADR 0033's rejected alternative "local model on the homelab"
was reversed by #514 for a self-hosted server, and the objection it
recorded — that a *deployed* game server should not depend on a home
network — still applies to production but not to this work, which is
about the dev box.

Two things about that path were suspect at the transport layer, and
either one on its own would make the `assisted` tier a heuristic seat
wearing a model tier's name. Both were probed by hand against the live
host on 2026-09-17 — Ollama 0.34.0, `qwen3:14b` Q4_K_M — before this
ADR was written, and they did not come back the same:

- **Thinking is not suppressed. Confirmed, and it is the live
  failure.** `model/openai.go` sends Ollama's native `think: false` on
  `/v1/chat/completions`, and that endpoint ignores it. The model
  reasons anyway, the trace arrives in `message.reasoning`, `content`
  comes back **empty**, and `finish_reason` is `length` — at
  `max_tokens: 64` the trace consumes the entire reply budget.
  `parseAnswer` has nothing to parse, so the window is charged
  `FallbackMalformed` and the runner takes Layer B's move. **Every
  `assisted` window on the live server today is the heuristic playing
  under the `assisted` label** — precisely the misstatement #514
  refused to ship at the picker, arriving anyway through the
  transport, invisibly, because a fallback is a supported outcome and
  nothing counts how often it is the only outcome.
  The fix was probed in the same session and works on this version:
  `reasoning_effort: "none"` suppresses the reasoning, and combined
  with `response_format: json_schema` (`analysis` then `index`) the
  reply came back valid JSON, `finish_reason` `stop`, 36 completion
  tokens, no reasoning field at all. That fix did not wait for this
  ADR's sequence — it shipped as
  [#839](https://github.com/krakenhavoc/cmd_and_ctrl/pull/839), which
  is why the baseline in decision 1 is a measurement of the model
  rather than of Layer B.
- **Silent truncation is a configuration hazard, not the live
  failure.** Ollama drops the *front* of an over-long prompt without
  an error and without a flag on the response, and its stock default
  on a 16 GB card is a 4k window. Our static system block — rules
  primer plus the bot's own decklist with oracle text — measures
  3.4–4.3k tokens across the four curated decks (decision 6 has the
  per-deck figures), and a whole call lands at 5–7k once the board
  delta is added. That does not fit a stock 4k window, and the half
  that gets dropped is the front: the primer and the instructions. On
  *this* box it is not currently biting:
  `/api/ps` reports the model loaded at `context_length` 16384 (the
  model's own ceiling is 40960), so the operator had already raised
  it. That is a host setting no part of this repo can see, assert or
  change — the OpenAI-compatible endpoint has no per-request field for
  `num_ctx` — so a fresh machine, a reinstall or a colleague's box is
  one `ollama serve` away from making it the live failure. A
  client-side guard and an operator note are the whole of the defence
  available.

Neither could have been settled by reading code: the behaviour is on
the far side of an HTTP boundary, and the repo had no way to look at
that boundary. The numbers above came from curl. **That is why the
probe becomes a command** (decision 5) rather than a paragraph in this
ADR — the next model, the next Ollama release and the next endpoint
will each raise the same question, and hand-running curl is not a
method.

One more thing the probe turned up that the design should use: the
`/v1` endpoint reports `usage.prompt_tokens_details.cached_tokens`, so
prefix-cache hits are measurable on the local transport. Our `Usage`
records cache-write and cache-read as zero today, and decision 6's
identical system prefix is a bet that only that field can settle.

Beyond the transport, what the model is *given* is thin for a 14B
model. There is no oracle text for anything outside the bot's own
decklist, so opponents' boards, stack items and pending-choice options
arrive as bare names. `GameView.Log` has existed on the wire since ADR
0033 §4 with 200 public events and the prompt does not render one of
them. Combat windows — 60–70% of everything that escalates, per ADR
0033 §5's own measurement — get no framing at all: no per-defender
life, no "this dies to that", nothing. Pending choices render as an
index list. There is no structured output, no temperature, and one
budget for every window regardless of what the window is worth.

And there is no measurement of any of it. The only strength number in
the repo is heuristic-versus-random. There is no arena, no decision
log carrying prompt and reply text, no labelled position suite.

**The owner's decisions (2026-09-17):** build both halves, eval first;
the target is `qwen3:14b` on the same box; real reasoning time per
decision is acceptable, but it must be **adjustable at the table**, so
a player who is bored can shorten it without an admin or a restart.
Success is defined as **win rate against the heuristic**, **agreement
with a labelled position suite**, and **fewer obvious blunders**,
which means decision logs a human can read.

## Decisions

### 1. Eval first, and "better" is three numbers

No lever is touched until there is something to measure it with. The
first three PRs build the decision log, the position suite and the
arena, and produce a baseline for `assisted` (qwen3:14b) against
`heuristic` before a single prompt byte changes. Every PR after that
carries the same report block in its description, so a change that
makes the bot worse is visible in the pull request that made it worse
rather than three sprints later.

"Better" is three measurements and nothing else:

1. **Win rate against the heuristic**, from a headless bot-vs-bot
   arena, with a **Wilson 95% interval** and the null rate printed
   beside it (25% for one seat in four, 50% heads-up). Ten games at
   one seat in four has an interval roughly ±25 points wide, which is
   the honest reading: the arena is a blunder detector and a
   regression alarm long before it is a strength meter, and the
   interval is printed so nobody reads a 4-of-10 as a result.
2. **Agreement with a labelled position suite**, overall and per tag.
   This is the fast signal — seconds for the heuristic, a minute or
   two for the model — and it is the one that says *where* a change
   helped. A lever that lifts `block` agreement 15 points and drops
   `choice` agreement 10 is a trade, not a win, and only the per-tag
   breakdown shows it.
3. **Reviewable decision logs.** Not a number. The definition of
   "fewer obvious blunders" that the owner gave is a human reading
   what the bot was shown and what it answered, which is impossible
   today because the prompt and the reply are locals inside
   `model.Policy.Decide` and are discarded when it returns.

Per-game spend (#735) and decision latency percentiles (#505) fall
out of the same instrumentation and close as side effects.

### 2. A decision trace, a runner observer, and a per-game decision log

The prompt and the reply are function locals. A wrapper policy cannot
see them and neither can the runner, so the hook goes on the policy
as an **optional extension interface**, the same pattern ADR 0033's
`Conceder` already uses:

```go
type Trace struct {
    Layer        string // "rules" | "heuristic" | "model" | "random"
    Rule         string // Layer A's absorption reason, when it fired
    HeuristicIndex int
    Candidates   []Candidate
    Escalations  []string
    Model        string
    Prompt       *Prompt // {System []string; User string}
    Reply        string
    ParsedIndex  *int
    Fallback     string
    TimedOut     bool
    ModelLatency time.Duration
    Usage        TokenUsage
}

type Tracer interface {
    DecideTraced(ctx context.Context, in Input) (Decision, Trace, error)
}
```

`model.Policy`, `rules.Filter` and `heuristic.Policy` implement it;
`Decide` becomes a thin wrapper over `decideTraced` so the two can
never disagree, and a compile-time assertion keeps each one honest.
Random policies are labelled by the runner rather than made to
implement the interface.

The runner grows `Config.Observer`, which receives **one event per
window**, emitted after the `room.Apply` result is known. That
placement matters: the event carries not only what the policy decided
but what the runner did with it — the dispatched index and label, the
runner's own fallback cause (`timeout`, `policy-error`,
`out-of-range`, `decline-pass`, `forced-pass`, `forced-always-legal`),
and whether the engine accepted it. A policy that answers well and a
runner that rejects the answer look identical from inside the policy
and completely different from here.

The writer is `aiseat/decisionlog`: one `<gameID>.decisions.jsonl` per
game, mode `0600` in a `0700` directory, buffered, mutex-serialised
across the four seat goroutines, capped at 256 MiB per game. Hitting
the cap logs **one** WARN — a cap that fires on every subsequent
record would bury the game it is describing — and every dropped record
is counted in `Stats`, so a short log is never mistaken for a quiet
game. `Manager` never imports
it — two one-method interfaces in `aiseat` keep the dependency
pointing the right way. Three modes: `escalated` (the default; the
full `Input` for windows that left Layer A, a compact record for the
rest — roughly 15–20 MiB per game), `all` (~80 MiB), and `model`.

Records are replayable offline, and there is a gated test that proves
it: re-running each record through `rules.Resolve` and
`heuristic.New().Decide` must reproduce the recorded verdict and pick.
A log that cannot be replayed is a log nobody will trust a month from
now.

**The hidden-information rule, stated once so it is not rediscovered
later.** Each record's `Input.View` is the seat's own filtered view —
the same bytes ADR 0033 §3's golden test pins to what a human client
at that seat receives. So an individual record leaks nothing that
seat could not see, and that is a real property worth having. **The
file is a different object.** It aggregates every bot seat in the
game, so reading one is reading four hands at once. Therefore:

- the decision log is **operator-only**;
- it is **never served over HTTP** — no lobby route, no admin route,
  no download link beside the replay;
- it is **never attached to a bug report** (ADR 0017's redaction
  pipeline is about credentials and does not help here — the problem
  is not secrets in the file, it is the file);
- it is **off by default**, enabled by `CMDCTRL_BOT_DECISION_LOG=<dir>`
  in production and `AISEAT_DECISION_LOG=<dir>` in gated tests.

This rule goes in `docs/bot.md` as well as here, because the person
who trips over it will be reading that page, not this one.

### 3. A position suite whose answers are labels, never indices

A suite position is one JSON file under
`suite/testdata/positions/<id>.json` holding a **frozen
`aiseat.Input`** — view, seat and the exact move list — plus `tags`
(`mulligan land cast removal combat attack block choice target
stack-response concede`), a `note` explaining the position in a
sentence, `source` provenance (game, seat, seq, log path, reviewer,
date), `at_capture` (what the heuristic and the model picked when it
was harvested), and `expected`.

`expected` is **matchers, not indices**: `accept` and `reject` sets
keyed by `label`, `label_re`, `kind`, or `type` plus a params subset,
with a `decline_ok` flag for windows where passing is defensible. The
index is derived at load time by matching against the frozen moves.

This is the load-bearing detail of the whole suite. A stored index is
correct until the day the enumerator's ordering changes, at which
point every position in the suite silently starts grading a different
move and the suite reports agreement numbers that mean nothing. A
matcher either still matches or it does not, and a matcher that
matches nothing **fails `Load` loudly** rather than being skipped.
The suite has to survive enumerator churn to be worth building, and
`legal`'s ordering is not a stable interface — ADR 0033 §1's own
closeout update records that candidate order past the expansion cap is
currently an accident of enumeration.

Each position also carries a `gate` field: the policy names for which
a miss **fails `go test`**. That is what makes this more than a model
eval. `suite_test.go` runs **ungated on every CI run** against the
heuristic tier — it is pure computation over frozen inputs, so it
costs milliseconds and needs no endpoint — and a gated position is a
pinned heuristic bug or a pinned heuristic invariant. The choice bugs
in #798 and #780 are exactly the shape that belongs here: reproduce
once, label once, and the regression can never come back quietly.
Model runs go through the `boteval` binary, because they need a
server to talk to.

Two supporting verbs make labelling affordable:

- **`harvest`** pulls candidate windows out of decision logs into an
  inbox — filtered by escalation, heuristic/model disagreement,
  fallback cause, layer, seat or tag — with `expected` left empty and
  tags guessed from the move kinds. An unlabelled position is
  `skipped` by `Run`, never counted as agreement.
- **`render`** prints the exact system and user text the model would
  see for a position, with the move list annotated
  `<- accept / reject / heuristic / model@capture`. This is the
  labelling screen, and it is also the fastest way to answer "what did
  the model actually get?" during the transport work.

### 4. The arena lives outside `aiseat/`, and rotates seats

The bot-vs-bot arena is `server/internal/botarena` plus a
`server/cmd/boteval` binary, and it is **not** under `aiseat/` for a
reason that is not stylistic: `heuristic/imports_test.go` bans
`internal/game` from every subpackage of `aiseat/` **including their
`_test.go` files** — it checks `Imports`, `TestImports` and
`XTestImports` — with only the `aiseat` root and `decks` exempt. An
arena has to build games. Meanwhile the `aiseat` root cannot import
`tiers`, so the arena cannot live there either. Putting it outside
keeps ADR 0033 §3's type gate exactly as strict as it is today, which
is worth more than the convenience of a shorter import path. The
decision log and the suite handle only `aiseat.Input`, `legal.Move`
and `protocol.GameView`, so they stay inside.

`Play` is `catalog_soak_test.go`'s `playCatalogGame` with the stall
returned rather than reported, one `rules.Meter` shared across the
seats, `deckprofile.Build` for model seats, and replays off by default
(320 MiB each, and the decision log is the artifact we actually want).
`BattleDeck` moves verbatim out of `heuristic_game_test.go`.

**Seat order rotates per game** — spec *k* sits at position
`(k+i) % n` in game *i*. Turn order in a four-player game is a large
effect and the decks are not mirror matches; without rotation a ten
game run measures the seat and the deck at least as much as the
policy, and the win rate has no defensible reading. Rotation is also
the cheapest possible fix, which is why it is not optional and not a
flag.

The report a run produces is fixed, and it is the same block the suite
prints and every PR pastes:

- **policy table** — games, wins, win %, Wilson 95% CI, the null rate,
  draws and stalls, turns p50;
- **funnel** — windows, Layer A/B/C share, escalation reasons, calls,
  timeouts, malformed, out-of-range, no-budget, tokens, prompt bytes
  p50;
- **latency** — whole-decision p50/p99/p999/max per policy and
  model-call p50/p99/max (this is #505 part 1, from a 1024-entry ring
  on the runner);
- **wall clock per game**, and where the logs landed;
- **provenance** — model id and quant, endpoint, `max_think`, the git
  SHA of `model/prompt.go`, seats, decks, rotation, games, seed.

The provenance line is not bookkeeping. Every number here is from one
model at one quantisation behind one prompt; a report that does not
say which is not comparable to the next one, and the prompt SHA is the
field that will be wrong most often.

Budget: one `assisted` seat is roughly 60 model calls per game at 3–15
seconds each, so 5–15 minutes per game, and four model seats serialise
on one GPU. Ten games is an afternoon, not a CI job, and the arena is
never wired into CI.

**Amendment (2026-09-24, #1503): a lockstep schedule, opt-in.** A seed
here fixed the deal and the policies' randomness, never the game: every
seat is a runner goroutine, and which seat acts first after a commit is
the scheduler's choice, so a rerun forks at the first contested
window. #1409 fixed the heuristic gate through a test-only door;
`aiseat` now exports that door as a small production API —
`NewStepped` builds a runner with no goroutine and no subscription, and
`Runner.Step` runs one wake of its ordinary act-loop on the caller's
goroutine (it panics on a runner built by `Start`, whose own goroutine
already steps it). `botarena.Config.Lockstep` / `boteval arena
--lockstep` steps the seats in chair order, round after round, so a seed
replays move for move; the stall there is exact (a round with no
commit) rather than timed, and the runner's wall-clock holds (MinThink,
BlockGrace) are off because nobody else can act while a seat holds. The
concurrent schedule stays the default, because it is the schedule a live
table runs and discussion #1390 keeps concurrent liveness and
deterministic quality as separate questions. A model seat is only as
reproducible as its endpoint.

### 5. `probe` becomes a command, and the transport fix that could not wait

The hand-run probe that produced the Context numbers becomes
`boteval probe`, shipped in PR 1: **one real request in the funnel's
exact shape** at the configured endpoint, printing

- `usage.prompt_tokens` against the client-side estimate — a server
  count far below the estimate is truncation, caught directly;
- `usage.prompt_tokens_details.cached_tokens`, so prefix reuse is
  visible;
- `finish_reason`;
- whether a reasoning field came back at all
  (`message.reasoning` / `message.reasoning_content`);
- whether `content` was empty;
- wall time.

It is run against every new endpoint, model or Ollama version before
anything else, and it answers "is the transport doing what we think"
in under a minute. That question has now been answered wrong once, for
an unknown number of weeks; the cost of asking it properly is one
HTTP request.

**`reasoning_effort: "none"` alongside `think: false` did not wait for
this ADR's sequence, and should not have.** It shipped on its own as
[#839](https://github.com/krakenhavoc/cmd_and_ctrl/pull/839), ahead of
the eval work, because it is a one-field bug fix for a live server
whose `assisted` tier was not calling a model at all. Both fields stay
under the existing `CMDCTRL_OPENAI_SEND_THINK` escape hatch, because
an operator pointing at something that is not Ollama needs a way to
turn our provider-specific fields off. #839 also lands
`Response.Reasoning`, so the trace is decoded and recorded rather than
discarded.

**What that changes about the PR 3 baseline, and it is the whole
point of putting it first:** the baseline is taken with thinking
correctly suppressed, so it measures **the model as designed today** —
a real qwen3:14b call against the thin prompt this ADR is about to
rebuild — and not the heuristic wearing the `assisted` label. Every
prompt lever from PR 5 onward is measured against a model that was
genuinely answering. Had the fix landed after the baseline, the whole
comparison would have been a rounding error against Layer B.

PR 4 keeps the rest of the transport work:

- **Temperature and seed plumbing.** `Request` gains `Temperature`,
  `Seed` and `Schema`; the OpenAI client sends all three, the
  Anthropic client sends temperature only. Local default temperature
  0, because a bot that plays a different move from the same position
  makes the suite meaningless.
- **Per-window recording** of `StopReason`, `Usage` and the reasoning
  field #839 already decodes, onto both `DecisionRecord` and `Trace`.
  **Never parse the reasoning for an index.** A reasoning trace is
  diagnostic text; a truncated one can contain a number that looks
  like an answer, and taking it would turn a clean malformed-reply
  fallback into a confident wrong move.
- **The context guard**, `Config.ContextTokens` — **default 0, meaning
  no pre-call shrinking.** This is the one place the plan's first
  instinct was wrong. A 4096 default would shrink every prompt on the
  only host we have, which runs at 16384, by a factor of four; the
  arena would then measure a shrunk prompt and attribute the result to
  the model. So the pre-call shrink is opt-in: the operator sets
  `CMDCTRL_BOT_CONTEXT_TOKENS` to match their `OLLAMA_CONTEXT_LENGTH`
  when they want it. When it is set, an estimate over budget drops
  material in a fixed, recorded order — recent events, then half the
  glossary, then opponents' graveyards, then the decklist — so that
  when truncation is unavoidable *we* choose what goes, from the back,
  instead of the model server choosing from the front.
- **The post-call detector always runs**, configured or not: a server
  `prompt_tokens` well below the client-side estimate, or a reply that
  ignores the requested format, logs a rate-limited warning and
  increments `Stats.PromptTruncations`. Detection is free and costs
  nothing when it never fires; shrinking is not, so only one of the
  two is on by default.
- The client cannot set `num_ctx` on `/v1/chat/completions`. So the
  real fix is operator-side: `OLLAMA_CONTEXT_LENGTH=8192` or more on
  the Ollama host, or a Modelfile `num_ctx`. That instruction belongs
  in `docs/bot.md` next to the endpoint variable, and the client-side
  detector above exists precisely because the instruction will
  sometimes not have been followed — as the current dev box shows, the
  setting that saves us is invisible from here and was never written
  down.

### 6. The prompt rebuilt for a 14B model

A frontier model can fill in a thin prompt from what it already knows
about Magic. A 14B model cannot, and the difference is most of the gap
between the two deployments.

- **Oracle text for what is on the table.** `OracleLookup` is injected
  into `model.Config` through `tiers.Options` and `tiers.FactoryOptions`
  and built in `botFactory` from the card index — the same injection
  pattern `DeckProfileFunc` already uses, so the type gate is not
  bent. Per-seat memo map. Face cards prefer the wire's active face.
  Everything goes through `cardName`, so a redacted card can never get
  oracle text attached to it.
- **A per-window `CARDS` glossary**, relevance-ordered and deduped,
  capped at 40 entries and 300 characters each: sources and targets of
  the moves actually shown, then the seat's own pending-choice options,
  then the stack, then opponents' non-land permanents by power, then
  the seat's own permanents with abilities, then hand. Basics and
  tokens are skipped. It renders after the board and before the moves.
- **The decklist comes off for the local profile.** `Config.Decklist`
  is `full | names | off`, local default **`off`**. The glossary
  carries what is in play, which is what the decision is about. The
  decklist block is the biggest thing we send and it was measured
  rather than guessed — rendered exactly as `prompt.go` renders it,
  from the Scryfall dump, at 3.5 characters per token (3 chars/token
  in parentheses):

  | Deck | Distinct cards | Static block |
  |---|---|---|
  | izzet-aggro | 88 | ≈ 4.3k (5.0k) tokens |
  | simic-ramp | 88 | ≈ 3.9k (4.5k) |
  | esper-control | 92 | ≈ 3.6k (4.2k) |
  | mono-black-aristocrats | 82 | ≈ 3.4k (4.0k) |

  Add a 1–2k board delta and a single call is **5–7k tokens**. That
  fits this box's 16k window and would be silently truncated at
  Ollama's stock 4k default — so the decklist is most of the §5 hazard
  on its own. It is also the prefill that gets **re-paid every time
  the single KV slot switches between bot seats**, because each seat's
  decklist is different, which is the cost that does not go away even
  on a correctly configured host. `names` is the middle option held in
  reserve — one line per distinct card, `name — type line`, no cost
  and no oracle text, about 700 tokens — to be measured if `off` turns
  out to cost the bot too much awareness of what is still in its
  library. Anthropic keeps `full`, where it is prompt-cached and
  effectively free.
- **An identical system prefix across every seat.** With the decklist
  gone, the static block is the primer alone and is byte-identical for
  all four seats, which is exactly what a shared KV cache wants. The
  deck name and archetype plan move to the first line of the user
  delta.
- **Window framing.** `windowKind(in)` classifies the window as
  attack / block / response / choice:*kind* / main / mulligan, and
  each gets its own section: attack renders per-defender life,
  untapped potential blockers and a per-creature "dies to X /
  unblockable at Y" built from the heuristic's own `couldBlock` and
  `kills` predicates; block renders each attacker, unblocked damage
  against my life, and what each of my blockers does; response renders
  the top of the stack with **"targets YOU"** in caps; choice renders
  the full `PendingChoice` shape — kind, reason, counts, accept and
  decline labels, life and pay costs, damage assignment — with oracle
  text on the options; main renders untapped mana by colour and
  whether the land drop is spent; mulligan renders the hand with
  oracle text, the land count and mulligans taken.
- **Grouped moves with real indices.** The flat list is grouped by
  (kind, source): a header carrying the card and its cost, children
  carrying the **real index** and only the part of the label that
  differs. Groups are capped, targets within a group are capped at six
  by heuristic rank, and a group that was cut says "N more not
  listed". Every line still carries exactly one `N: ` index so
  `fake.go`'s `indexFromPrompt` keeps working, and the fallback marker
  becomes an exported constant shared by `prompt.go` and `fake.go`
  instead of two string literals that have to agree.
- **`Config.HintMode`** — `none | marker | reason | scores`, default
  `reason`: the heuristic's pick is marked, with its one-phrase reason,
  at the **end** of the line and with no numeric scores. The hint is
  useful and it is also an anchor, and we can measure which it is:
  agreement-with-fallback per window kind on the suite, and `reason`
  against `none` in the arena. Above about 90% agreement on escalated
  windows the model is copying the hint and the tier is decorative.
- **Recent events.** The last 24 entries of `GameView.Log` — already
  rendered and redacted per viewer by ADR 0033 §4 — with `step`
  entries collapsed into one header per turn and actor, `draw` dropped,
  and the seat's own name substituted for "you". A snapshot cannot say
  that the player on the left wiped the board last turn, and ADR 0033
  §4 argued at length that this is most of what a Commander player is
  reasoning about. The log shipped; the prompt never read it.

### 7. Structured output with a bounded scratchpad; native thinking stays off

The reply is requested as `response_format: json_schema` with the
properties in emission order: `analysis`, a string capped at 400
characters, then `index`, an integer. The routine profile's schema is
`index` only. A server that rejects `response_format` gets one retry
with `json_object` and is then remembered as unsupported for the life
of the process; `parseAnswer` stays as the tolerant fallback under all
three cases.

Emission order is the entire point. The model writes a short amount of
reasoning *before* the index, inside a grammar that bounds how much,
and the grammar guarantees the index is there when it stops.

**Native thinking stays off**, and this is a deliberate choice rather
than an oversight:

- it is unbounded, so it cannot be fitted into a 4k context budget or
  a table-latency budget;
- it is not constrained by the response grammar, so no schema can cap
  it;
- and a thinking trace that hits the token cap produces **no content
  at all** — the failure mode is a total loss of the answer, not a
  shorter one. A 400-character scratchpad that overruns still leaves
  a parseable index most of the time.

`ModelProfile` gains `Scratchpad`, with `MaxTokens` 48 routine and 160
escalated. `Thinking: "adaptive"` stays on the shelf as a later
experiment, to be tried once there is a suite to measure it with.

A `finish_reason` of `length`, or a parse failure with budget
remaining, triggers **one retry with the index-only routine profile on
the same prompt bytes** — the prefix is hot in the KV cache, so the
retry is cheap — counted in `Stats.Retries`, and falling back with
`Fallback = "length"` if it fails again.

The `analysis` text is logged on every answer. It is the bot
explaining itself in its own words on every window, which is free
labelled data for the suite and the fastest way to see a
misunderstanding that a chosen index alone would hide.

### 8. Two budgets per tier, and a think budget adjustable at the table

One budget for every window is wrong in both directions: it is too
long for a window the funnel barely cared about and too short for the
one it escalated. `ModelProfile.Budget` splits it — 8s routine,
`maxThink − Reserve` escalated — and the effective budget is
`min(deadline − Reserve, profile.Budget)`. When the remaining time is
below the escalated budget but above the routine one, an escalated
window **downgrades to the routine profile instead of timing out**. A
narrower answer beats no answer, and no answer means the heuristic's
move under the model tier's name, which is the failure ADR 0033 §10
and #514 both went out of their way to avoid.

The per-game think budget is adjustable **from the table, mid-game, by
any seated player** — not admin-only, not restart-only. The owner's
requirement is a player who is bored of waiting, and that player is
sitting at the table with no admin session and no intention of
restarting a server.

The data flow, end to end:

- **lobby →** `addBotRequest.ThinkSeconds` at add time, and
  `PATCH /games/{id}/bots` with `{"think_seconds": n}` afterwards,
  gated by the existing `botSeatAuthorised` check and deliberately
  carrying **no game-state check**, so it works mid-game;
  `botOptionsResponse` grows `Think{Min, Max, Default, Presets}` so
  the client renders the real bounds rather than hard-coding them.
- **lobby state →** `GameMeta.BotThinkSeconds`, persisted like every
  other meta field, `aiseat.SeatSpec.Think`, and an optional
  `BotThinkHost` interface on the host — the pattern `BotTierReasons`
  already establishes.
- **Manager →** a `botGame.think atomic.Int64`, shared with every
  runner in that game as `Config.ThinkOverride`. Atomic because the
  write comes from an HTTP handler and the reads come from four seat
  goroutines mid-game; nothing else needs synchronising because
  nothing else changes.
- **Runner →** `decide()` uses `min(override, cfg.MaxThink)` as the
  context deadline. **The funnel needs no change at all**: the call
  budget already follows the context deadline, so the entire feature
  lands above `model.Policy`.

The bounds are derived from the ceiling rather than hard-coded, so
that the control behaves sensibly on a deployment this ADR is not
about:

- **Ceiling** is the tier's `MaxThink` after `CMDCTRL_BOT_MAX_THINK` —
  the operator's number, 20s by default on the local transport. The
  table can lower, never raise.
- **Floor** is `min(CMDCTRL_BOT_THINK_FLOOR, ceiling)`, the variable
  defaulting to **4s**. Not zero, and not ADR 0033 §10's 2s tier
  default: at 2s a 14B model on this hardware cannot finish, so every
  window times out and "fast" silently means "heuristic seat wearing
  the `assisted` label". A floor that lies about what the player is
  playing against is worse than a slow bot.
- **The control is offered only when `ceiling > floor`.** On an
  Anthropic deployment, where ADR 0033 §10's 2s and 5s still stand,
  floor and ceiling collapse together, `botOptionsResponse` reports no
  think options, the lobby select and the Settings row do not render,
  and nothing about that deployment changes. There is no useful range
  to expose when the whole budget is shorter than the floor, and a
  disabled slider is worse than no slider.
- **Presets are computed from the ceiling**, not written down:
  **fast = `max(floor, ceiling/3)`**, **normal = `max(floor,
  ceiling×0.6)`**, **deep = ceiling**. With the local defaults that is
  6s / 12s / 20s, which is where the numbers in `CMDCTRL_BOT_THINK_DEFAULT`
  come from; on any other ceiling the three presets stay meaningfully
  apart instead of piling up against a hard-coded 6. Presets rather
  than a slider because the meaningful question is "is this too slow",
  not "is 9 better than 11".

**A profile budget never exceeds the remaining deadline** — that is
what the `min` above says, and it is intended rather than incidental.
A table that sets think to 6s shortens the routine calls too: routine
asks for 8s, gets 6s minus `Reserve`, and answers in the time it has.
The alternative, holding routine at its configured 8s inside a 6s
deadline, would time out every routine window while the player
believed they had merely asked for a faster bot.

The client gets `setBotThink` in `api.ts`, `think` on `BotOptions` and
`GameMeta`, a third select in the lobby picker, and a preset row in
Settings beside "show bot reasoning", shown when the current game has
bot seats.

### 9. Escalation on a single-model deployment carries a plan forward

`CMDCTRL_BOT_FRONTIER_MODEL` defaults to `CMDCTRL_BOT_MODEL`, so on
this deployment escalation does not buy a better model — ADR 0033 §6's
closeout update already records that. What it buys is the scratchpad,
the larger budget and a wider candidate list, and those are worth
spending on the window that decides the turn and wasteful on the third
chump-block.

`Config.CombatEscalation` is `every | first | none`, default **`first`**:
escalate the **first declaration window of a combat step**, then carry
that window's `analysis` forward on the same turn and step's later
windows as

```
YOUR PLAN FROM A MOMENT AGO: "…"
```

at routine cost. This is the cheapest multi-window memory available to
us — it needs no session state, no conversation history and no second
call — and it addresses the specific waste ADR 0033 §5's closeout
update named: combat fires the escalation trigger on every declaration
and bots declare a lot of combat.

The one unconditional exception: **a block window where unblocked
damage is at or above the seat's life total always escalates**,
whatever the mode says. That is the window where a cheap answer costs
the game.

### 10. The config surface, and what does not change

Every lever above is a `model.Config` or `ModelProfile` field with a
local default, so the suite and the arena can A/B any of them without
a rebuild. `model.LocalConfig()` joins `DefaultConfig` and
`StrongConfig`, and `tiers.New` selects it when the OpenAI transport
won at boot.

- `model.Config`: `Oracle`, `Decklist`, `MaxGlossary`,
  `MaxOracleChars`, `LogEvents`, `ContextTokens`, `HintMode`,
  `CombatEscalation`, `RetryOnLength`, `Local`.
- `ModelProfile`: `Budget`, `Temperature`, `Seed`, `Scratchpad`,
  `Format`.
- New environment variables, all resolved in `loadConfig`:
  `CMDCTRL_BOT_THINK_DEFAULT`, `CMDCTRL_BOT_THINK_FLOOR`,
  `CMDCTRL_BOT_ROUTINE_BUDGET`, `CMDCTRL_BOT_CONTEXT_TOKENS`,
  `CMDCTRL_BOT_DECKLIST`, `CMDCTRL_BOT_HINTS`,
  `CMDCTRL_BOT_SCRATCHPAD`, `CMDCTRL_BOT_JSON_SCHEMA`,
  `CMDCTRL_BOT_TEMPERATURE`, `CMDCTRL_BOT_LOG_EVENTS`,
  `CMDCTRL_BOT_COMBAT_ESCALATION`, `CMDCTRL_BOT_DECISION_LOG`,
  `CMDCTRL_BOT_DECISION_LOG_MODE`.
- Existing variables are unchanged. The boot log prints the resolved
  local profile once, because a bot behaving oddly is a question about
  which knobs were set, and the answer should be in the log the
  operator already has.

**The Anthropic transport keeps today's behaviour exactly.** It stays
on the `{"index","why"}` reply shape, with `Scratchpad=false` and no
`Format`, temperature the only new field it sends, and the decklist
`full` and prompt-cached. `TestStaticBlockIsByteIdenticalAcrossDecisions`
still holds, because the glossary and the event log are both in the
per-decision delta, not the static block. None of the work below is
allowed to regress the hosted path in exchange for the local one; two
profiles is the cost of having two very different deployments.

## Measurements

*Empty on purpose. PR 3 fills this section with the baseline run and
each later PR appends its own row; an ADR that describes an eval loop
and carries no numbers from it is a plan, not a record.*

Every entry records: model id and quantisation · endpoint ·
`max_think` · git SHA of `model/prompt.go` · seats and decks · games ·
seed · win rate with Wilson 95% CI · Layer A absorption, escalation
rate, timeout rate, malformed rate · decision p50/p99 · tokens per
game, with `cached_tokens` where the endpoint reports it · wall clock
per game · suite agreement overall and per tag.

## PR sequence

One lever per PR. Every PR from 4 onward pastes the suite and arena
report blocks into its description and names the `model/prompt.go` SHA
it measured.

| # | Scope |
|---|-------|
| 0 | This ADR and the tracking issue |
| 0.5 | **Shipped as [#839](https://github.com/krakenhavoc/cmd_and_ctrl/pull/839)** — `reasoning_effort: "none"` on the `/v1` endpoint and `Response.Reasoning`. Out of sequence on purpose: it is a live bug, and it is what makes the PR 3 baseline a measurement of the model |
| 1 | Trace + runner observer + decision log + latency percentiles (#505 part 1) + `boteval probe`; also the stale "ADR 0024" citations in `legal/legal.go` and `aiseat/policy.go` (from #501) |
| 2 | Position suite + `boteval suite run\|harvest\|render`, ~20 seed positions hand-labelled, ≥5 gated for the heuristic |
| 3 | Arena + `boteval arena` + the baseline numbers, into this ADR and `docs/bot.md`; folds Discussion #637 option A. Measures the model as it stands today, with #839 in |
| 4 | Temperature/seed plumbing, per-window stop-reason and usage recording, opt-in context guard, always-on truncation detector (§5) |
| 5 | Oracle glossary, hand with oracle text, decklist off, shared prefix (§6) — the largest expected single gain |
| 6 | JSON schema, bounded scratchpad, length retry (§7) |
| 7 | Combat framing, grouped moves, `HintMode` (§6) — combat is most of what escalates |
| 8 | Recent events from `GameView.Log` (§6) |
| 9 | Two budgets and the table think control, server and client (§8) |
| 10 | `CombatEscalation: first` and plan carry-over (§9) |

Then tune `HintMode`, `Decklist: names` and `Thinking: "adaptive"`
against the suite, as the data says.

## Consequences

- **The table gets slower on the local transport, on purpose.** A 12s
  default against ADR 0033 §10's 2s tier default is a different game
  to sit at. The table control in §8 is what makes that acceptable,
  and it is the reason that decision is in this ADR rather than
  deferred: raising the budget without shipping the lever to lower it
  would be a regression for anyone who does not want to wait.
- **Decision logs are large — larger than this ADR first estimated.**
  Measured in PR 1 (#842): a four-player board view is most of a
  record, so one window costs 38–59 KiB and a four-seat game writes
  **114–236 MiB** in the default `escalated` mode, not the 15–20 MiB
  guessed at here. `escalated` also only compacts windows that Layer A
  absorbed, so it saves nothing for a policy that does not run Layer A
  (the bare `heuristic.New()` the whole-game tests seat, as opposed to
  the `heuristic` *tier*, which is `rules.Filter` over it). The 256 MiB
  per-game cap is therefore a real limit rather than a theoretical one:
  a long four-seat game can reach it, and past it records are dropped
  and counted with one WARN. They remain off by default, and nothing
  rotates them — the operator who turns them on is responsible for
  cleaning up, and `docs/bot.md` says so.
- **The Ollama host needs configuring, and the server cannot do it.**
  `OLLAMA_CONTEXT_LENGTH=8192` or more (or a Modelfile `num_ctx`) has
  to be set on the model host; `/v1/chat/completions` has no
  per-request field for it. §5's post-call detector keeps a
  misconfigured host from failing *silently*, but it cannot make it
  correct, and the pre-call shrink is opt-in precisely so that it
  cannot quietly degrade a host that was configured properly.
- **A new binary.** `server/cmd/boteval` is an operator tool, built by
  a Makefile target, never deployed and never in CI. `botarena` is the
  first package that builds games outside a test, which is what keeps
  ADR 0033 §3's type gate intact rather than eroded.
- **Two prompt profiles to maintain.** Local and hosted diverge on the
  decklist, the reply schema, the scratchpad and the budgets. That is
  a real maintenance cost, accepted because a 14B model on a 16 GB
  card and a frontier model behind an API want genuinely different
  prompts, and pretending otherwise is what produced the current
  situation.
- **CI gains one ungated test and no model dependency.**
  `suite_test.go` runs against the heuristic in milliseconds on every
  push. The arena never runs in CI; it is an afternoon on the dev box.
- **#505 and #735 close as side effects**, and ADR 0033 §5's open exit
  criterion — per-game spend measured rather than estimated — is
  finally answerable, against a local endpoint where the spend is
  measured in seconds and watts rather than dollars.

## Rejected alternatives

**Send every window to the model.** ADR 0033 §5 rejected this on cost
and it stays rejected on a different ground here: at 3–15 seconds a
call on this hardware, a four-bot table would take hours per game, and
Layer A absorbs ~90% of windows precisely because ~90% of windows are
not decisions. Nothing in the measured absorption data suggests the
rules layer is throwing away interesting choices.

**Turn native thinking on with a token cap.** The obvious response to
"thinking is probably not suppressed" is to embrace it. Rejected for
the three reasons in §7, of which the third is decisive: a thinking
trace cut off by the cap yields **no content**, so the failure is
total rather than graceful. The bounded `analysis` field gets most of
the benefit inside a grammar that can actually bound it. Revisit when
there is a suite to measure it against — which is what PR 2 builds.
This is not a prediction, either: the probe produced exactly that
failure, `content` empty and `finish_reason` `length`, and it is what
the live server has been doing on every `assisted` window.

**Propose-and-validate instead of the closed list.** Re-rejected for
ADR 0033's reason, and more strongly for a 14B model: the closed list
is the structural defence against invented actions, and a smaller
model invents more, not less. Every decision above adds information
*about* the listed moves; none of them lets the model emit one.

**Keep the full decklist in the local prompt.** It is the single
largest block we send, it is ~4k tokens of prefill on a 4k context, it
is mostly cards not currently in play, and four seats sharing a GPU
re-prefill it against one KV slot. The glossary carries what the
decision is actually about. The hosted profile keeps it because there
it is cached and costs nothing.

**A fixed think deadline, set by the operator only.** Simplest thing
that could work, and it fails the stated requirement: the person who
wants the bot to hurry up is a player mid-game, not an operator with
shell access. A restart to change a bot's pace is not a feature.

**Storing move indices as the suite's expected answers.** Half the
code and a fraction of the matcher machinery. It also produces a suite
that goes quietly wrong the first time enumeration order shifts,
grading a different move than the reviewer labelled while continuing
to report a number. A regression gate that can silently grade the
wrong thing is worse than no gate, because it is trusted.
