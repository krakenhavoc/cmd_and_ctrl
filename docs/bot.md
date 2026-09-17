# AI bot seats (S31)

A bot is a seat at the table with no browser attached. A goroutine on
the server watches the room, asks a policy which of its legal moves to
make, and dispatches it down exactly the path a WebSocket client's
action takes — so a bot's move is one `seq`, one snapshot, one replay
line and one undo entry, the same as yours.

This page is the player-facing guide: how to add one, what the tiers
and decks are, and what a bot will and will not do. The architecture,
and the reasoning behind every decision below, is in
[ADR 0033 — AI bot seat](decisions/0033-ai-bot-seat.md); the eval loop
and decision harness that measure and tune the model tiers — decision
logs, the position suite and the bot-vs-bot arena — are in
[ADR 0052](decisions/0052-bot-decision-harness-and-eval.md). The HTTP
surface is specified in [docs/lobby.md](lobby.md); the chat and view
fields are in [docs/protocol.md](protocol.md).

> **What is actually wired today.** All four tiers and all four
> curated decks are reachable from the lobby. `random` and `heuristic`
> need nothing but the binary; `assisted` and `strong` call a model,
> and a server with no model endpoint configured reports them
> unavailable — with a reason the picker shows — rather than offering
> a seat that would play the heuristic under a model tier's name. See
> [the model endpoint](#the-model-endpoint) for the one environment
> variable that turns them on.
>
> **What is not live yet: improvisation.** The mechanism described
> [below](#improvisation-and-why-an-undo-is-free) shipped and is
> tested. No tier implements it yet, though, so no bot improvises in a
> game today. [#686](https://github.com/krakenhavoc/cmd_and_ctrl/issues/686)
> builds it for `assisted` and `strong`.

---

## Adding a bot

**Any player already seated at an unstarted table can add one** — this
is not an admin chore. Admins can too, and a spectator cannot: a
spectator's session carries the game ID but not a seat, and the
request is refused with a 403.

In the lobby, an open seat grows an **add bot** control. It opens a
small picker with two dropdowns — tier and deck — and the seat appears
immediately, filled, with a **BOT** chip. A bot seat carries a real
deck, so it satisfies the "every seat has uploaded a deck" gate and
**Start** lights up exactly as if a human had joined and uploaded.

Removing one is the same control in reverse, on the bot's seat, and
works only while the game is unstarted. Removing a bot closes the gap
in seat numbering, which is why the API keys the delete on the seat's
**player UUID** rather than its index — an index is a value the client
would have to re-read between reading it and using it.

### Seat arithmetic

**Bots take real seats.** The table holds four, so a seated human can
add **at most three**. There is no way for a player to make a
four-bot table: making one would require a fourth bot in the seat the
requesting human is sitting in.

A four-bot table is reachable only through the admin path (`POST
/games` is already admin-only, so an admin can create a table and fill
all four seats). That is deliberate — the all-bot table is the
engine's fuzz harness, not a player flow.

One human plus one bot is a legal game (`MinPlayers` is 2) and is the
solo-practice case.

### Under the hood

| | |
|---|---|
| `GET /bot/options` | What the picker renders from: the tier list and the deck catalog. Game-independent, so the client fetches it once. |
| `POST /games/{id}/seats/bot` | `{tier, deck}` — or `{tier, format, source}` to paste a decklist instead of naming a curated one. 201 with the updated game. |
| `DELETE /games/{id}/seats/bot/{player_id}` | While unstarted. 200 with the updated game. |

The deck travels exactly as it does for a human's upload: the same
`ParseText` → `Resolve` → `Validate` pipeline, the same 422 violation
shape. A bot cannot be seated with a deck you could not have uploaded
yourself, and a curated deck that rots — a renamed card, a new ban —
fails loudly at the same place a human's would rather than silently
seating a broken library.

Full request and response shapes, and every error code, are in
[docs/lobby.md](lobby.md#get-botoptions-s31).

### Lifecycle

`Start` launches one runner goroutine per bot seat. A runner exits on
its own the moment the game leaves the active state, so the end of a
game needs no explicit stop. Deleting a game and shutting the server
down both cancel the runners and wait for them.

**A bot survives a deploy.** `is_bot`, `bot_tier` and `bot_deck` ride
both the engine snapshot and the persisted lobby metadata, and the
lobby's restore path relaunches a runner for every bot seat in a game
that came back active.

---

## The four tiers

The tier is the difficulty slider, set per bot at add time and shown
on the seat's chip. All four names are declared by the API from the
first release so the wire shape never changes as policies land.

| Tier | Plays | `MaxThink` | Needs |
|---|---|---|---|
| `random` | Picks uniformly among its legal moves. | 2s | nothing |
| `heuristic` | Scores the board and plays the best move it can see. Free and deterministic. | 2s | nothing |
| `assisted` | The heuristic, with a model consulted on the close calls. The intended default. | 2s | a model endpoint |
| `strong` | A model on every window that survives the rules filter, over a wider candidate list. | **5s** | a model endpoint |

`MinThink` is 700ms for every tier: a fast decision is held so the
table does not feel precognitive. `MaxThink` is a hard deadline, not a
target — on expiry the runner takes the fallback answer and logs the
miss. **The table never waits on a model.**

Those `MaxThink` figures are sized for a hosted model. A model running
on your own hardware is usually slower than either, so the deadline is
a deployment setting (`CMDCTRL_BOT_MAX_THINK`) and defaults to **20s**
when a local endpoint is configured. Getting this wrong is quiet
rather than loud: a model that never answers in time means the
heuristic plays every window while the chip still says `assisted`. The
server logs `bot model call TIMED OUT` per window when that is
happening, and the funnel counts it separately from an outage.

**`random` is not a joke tier.** A four-`random` table playing
unattended is the cheapest rules-engine fuzzer this project will ever
get. Its first hour of unattended play found four engine bugs, two of
which had been wrong in every human game for months. It is not an
opponent; it is a warm body.

**`heuristic` is the safety net under everything above it.** It is the
fallback whenever a model call times out, errors, or returns an
unusable answer, which is why it had to be built before the model
tiers and why it must stand alone.

### An unavailable tier is refused, not downgraded

Every declared tier is listed by `GET /bot/options`, including the
ones no policy is wired for, each carrying an `available` flag. The
picker greys out the unavailable ones; asking for one anyway is a
**422**, deliberately, and never a silent downgrade to `random`.

A bot whose chip says "strong" and which plays at random is worse than
no bot at all, because you would tune your play against a label that
is lying to you.

### The model endpoint

`assisted` and `strong` need one, and either of two kinds will do. The
server picks the local one when both are set.

| Variable | What it is |
|---|---|
| `CMDCTRL_OPENAI_ENDPOINT` | An OpenAI-compatible `/v1/chat/completions` server — Ollama, LM Studio, llama.cpp's server, vLLM. A URL (`http://192.168.1.18:11434` for a box on the LAN), or `1` for a stock Ollama on this machine. |
| `CMDCTRL_OPENAI_API_KEY` | Optional; most local servers want no key at all. |
| `CMDCTRL_OPENAI_SEND_THINK` | `0` stops the client sending the two thinking-off fields below. Only for a server that rejects one of them — see below. |
| `CMDCTRL_BOT_MODEL` | The model id to ask for. **Required for a local endpoint** — it is the name your server serves, e.g. what you `ollama pull`ed. |
| `CMDCTRL_BOT_FRONTIER_MODEL` | The model for escalated windows. Defaults to `CMDCTRL_BOT_MODEL`; one model in both slots is a supported configuration, and escalation then changes how a window is asked, not which model answers it (see [Known limitations](#known-limitations)). |
| `CMDCTRL_BOT_MAX_THINK` | The model tiers' hard deadline, as a Go duration. Defaults to 20s with a local endpoint. |
| `CMDCTRL_ANTHROPIC_API_KEY` | The hosted alternative. `CMDCTRL_ANTHROPIC_ENDPOINT` overrides the URL. |

**Thinking is turned off on the local transport, and that is a
deadline decision.** Several strong local models — qwen3 among them —
ship hybrid thinking on by default, and a thinking model inside a
2-to-20 second budget spends the budget on thinking tokens and answers
nothing. Turning it off was measured, against a real host (Ollama
0.34.0, `qwen3:14b`), to need two fields, not one: `think: false`
alone is IGNORED by `POST /v1/chat/completions` — the model kept
thinking, the reply came back empty, and every call scored as a
malformed reply, with the seat quietly playing the heuristic under the
`assisted` label. `reasoning_effort: "none"` is the field that
endpoint actually honours, so the server now sends both — `think:
false` for a server that honours it, `reasoning_effort: "none"` for
one that (like Ollama's) does not. `CMDCTRL_OPENAI_SEND_THINK=0` stops
both fields being sent, for a server that rejects one of them — at the
cost of getting the behaviour above back.

**Probe the endpoint before you trust it.** `boteval probe` sends one
request in the funnel's exact shape and prints what came back: the
endpoint and model it dialled, the prompt's byte size and a rough
token estimate, `usage.prompt_tokens` and `completion_tokens`, the
server's prefix-cache hit count, `finish_reason`, whether the reply
text was empty, whether a `reasoning` field came back, the first 300
characters of the reply, the parsed index or the parse error, and the
wall time. It ends with a verdict line for each of the two failures
that make a model seat look like a heuristic seat: **truncation** (the
server counted far fewer prompt tokens than were sent, so it silently
dropped the front of the prompt — the primer and the instructions),
and **thinking not suppressed** (a `reasoning` field came back, or
`finish_reason` was `length` with nothing usable in `content`). It
always exits 0: it is a diagnostic, and "the endpoint is down" is a
finding.

The probe sends what the seat sends — the same `OpenAIClient`, so both
thinking-off fields go with it — which means **`THINKING: suppressed`
is the expected verdict now that `reasoning_effort` ships**. A probe
that still says "not suppressed" is telling you this server honours
neither field, and that tier will play the heuristic on every window.

Point it at a Scryfall dump. Without one the static block is the deck
NAME alone, about 1.5 KB, and the truncation verdict is then measuring
a prompt an order of magnitude smaller than the one a real seat sends.

```bash
make -C server build-boteval
CMDCTRL_BOT_MODEL=qwen3:14b \
  ./server/bin/boteval probe --endpoint http://192.168.1.18:11434 \
    --dump data/scryfall/default-cards.json
```

**Running without one is still a supported deployment.** The model
tiers are complete policies with no endpoint — they play the rules
filter plus the heuristic — which is what makes every model failure
cheap, and what the model-outage drill proves: kill the endpoint
mid-game and the table plays on to a winner without stalling. What a
server with no endpoint does NOT do is offer those tiers in the
picker, because a seat playing the heuristic under the `assisted`
label tells you something false about the game you are in. A seat
already at the table when an endpoint goes away keeps playing; a new
one cannot be seated at a tier this server cannot honour.

---

## The four curated decks

Curated, not generated, and **every card in them is covered by a
build-failing test**: a card whose oracle ID has no registered effect
spec fails the build, with the back-face keys of double-faced cards
explicitly rejected so a card does not count as covered because its
back face registered. Without that test "curated" would rot the first
time a spec was refactored.

| ID | Name | Commander | Archetype |
|---|---|---|---|
| `izzet-aggro` | Raid and Ransack | Mary Read and Anne Bonny | Aggro (UR) — cheap creatures that turn artifacts and discards into damage, backed by burn and a thin counterspell suite |
| `simic-ramp` | Deep Roots | Tatyova, Benthic Druid | Ramp-stompy (UG) — mana creatures and land ramp into large green threats, drawing a card on every land drop |
| `esper-control` | The Long Answer | Hashaton, Scarab's Fist | Control (WUB) — counterspells, one-for-one removal and six board wipes, with just enough creatures to close |
| `mono-black-aristocrats` | Body Count | Syr Konrad, the Grim | Aristocrats (B) — free sacrifice outlets, drain-on-death payoffs, and a graveyard full of things to sacrifice again |

Each is exactly 100 cards: one commander and ninety-nine mainboard.
Deck IDs are wire values and will not be renamed.

All four are in the picker. The `placeholder-mono-red` stand-in that
sub-PR 4 shipped — a commander and ninety-nine Mountains — is gone
from it.

**A bot deck runs the identical pipeline your upload runs.** The
server holds the decklist as plain text — the same bytes you could
paste into the upload box — and then parses, resolves and validates it
against the same Scryfall index, so there is no second legality path
to keep in sync.

Not built, and for two different reasons. **Voltron / Equipment /
Aura** is waiting on card count rather than on engine machinery: the
attachment layer shipped, and so did the first seven attachments,
which is seven short of a deck. **Combo** is out on principle — a bot
executing a combo line is a bad experience for the table regardless of
how well the catalog supports it.

---

## Improvisation, and why an undo is free

> **Not live yet.** Everything in this section is built into the
> runner and covered by tests: the bundle, the validated announcement,
> the replay tag and the free undo. But improvising is something a
> policy has to opt into, through `aiseat.Improviser`, and none of the
> four shipped tiers does. **No bot improvises in a game today.** A bot
> picks only from its legal moves. The section describes how
> improvisation will behave once
> [#686](https://github.com/krakenhavoc/cmd_and_ctrl/issues/686) gives
> `assisted` and `strong` an implementation.

The catalog is a few hundred cards, and a bot's deck is drawn entirely
from it — but a card can be registered without every clause of it
being implemented. When the line a bot wants needs an effect the
engine cannot execute, **the bot will be allowed to do it by hand**, with
the same four sandbox verbs you have: `move_card`, `change_life`,
`add_counter`, `mark_damage`.

Three things make that safe, and all three are enforced in code rather
than left to the policy:

**It is one bundle, all or nothing.** Every verb in an improvisation
runs under a single hold of the room lock and commits as one `seq`,
one replay line and one undo entry — or, if any step fails, none of
them do. There is no reachable state where half an improvisation is
applied. The verb list is a closed allow-list: a bundle carrying
`concede` or `discard_selection` is refused outright.

**It must announce itself, before it happens.** The bot posts a chat
line naming the card, the intended effect, and the fact that it was
improvised, and the announcement is built and validated **ahead of**
the commit — a bundle that cannot name its card and its effect is
refused before anything is dispatched. The line is always the same
shape — the card, the effect, and a fixed suffix:

> *&lt;card&gt;*: *&lt;what it was meant to do&gt;* (improvised — the rules
> engine can't run this card, so I applied it by hand; any player can
> undo it)

This is **mandatory disclosure**. An unannounced improvisation is a
bot cheating, and a client that hid the announcement is what would
make it one. The same commit writes a `bot_improvisation` tag into the
replay log, so improvisations are greppable when a game goes wrong.

**Any player can undo it, and it costs them nothing.** This is the
part that will otherwise look like a bug, so it is worth stating
plainly:

- An improvisation is stamped with a nil caller rather than the bot's
  seat, so **any seated player can undo it** — not just an admin, and
  not just the bot.
- The undo is **free**. It does not spend your `UndosRemaining`, and
  it works even when you have already spent your take-backs this turn.

The budget exists to police the social cost of taking back *your own*
move. An improvisation is a bot asserting a rules interpretation the
engine could not execute; a human correcting it is doing maintenance
on a catalog gap. Charging for that would make the careful response —
read the announcement, check the card, put it back — cost more than
the lazy one, in a feature whose entire safety argument is that it is
reversible.

**The recourse is shallow, and the announcement matters more than the
undo.** Undo pops only the top entry of one room-wide stack, so an
improvisation followed by anything else is out of reach. That is a
known limit, and it is why the bundle is one entry and why the
disclosure is not optional.

One consequence worth knowing: the nil caller is a genuine power
grant. It bypasses the "you may only touch your own cards" gate — it
has to, since improvised removal reaches an opponent's board and an
improvised drain reaches their life total. Improvisation is the one
path on which a bot acts with admin authority, which is exactly why
the verb list is closed and the announcement is validated first.

---

## Show bot reasoning

**Settings → Gameplay → Show bot reasoning** (`settings.gameplay.showBotReasoning`,
off by default) surfaces the policy's own one-line justification for
each move it makes, in the bot feed on the board.

It is debug output, and it is honest debug output: a bot's reasoning
can name cards in **its own hand**. That is a disadvantage the bot
accepts, not a leak of yours — a policy is handed the same filtered
view a human at that seat receives and has no handle that could reach
another seat's hidden state.

The setting does **not** gate improvisation announcements. Those are
mandatory disclosure and are shown at every setting, to everyone.

---

## Decision log

**Off by default. Operator-only. Never served, never attached to a bug
report.** Rationale: [ADR 0052](decisions/0052-bot-decision-harness-and-eval.md).

`CMDCTRL_BOT_DECISION_LOG=<dir>` makes the server write one JSONL file
per game — `<dir>/<game-id>.decisions.jsonl` — with one line per
decision window per bot seat. A line records the turn and step, which
seat, which layer of the funnel answered, the Layer A rule or the
heuristic's whole ranking, the exact prompt the model was shown and
its raw reply, the index parsed out of it, the runner's own fallback
cause when it overruled the policy, whether the engine accepted the
move, and how long the decision took.

Two facts are recorded separately on purpose: why the runner did not
use the answer the policy returned (a timeout, an error, an
out-of-range index) and what it did *instead* (forced a pass, took the
enumerator's unconditional answer, ran out of answers, or was
cancelled mid-window). A window can be both, and collapsing them loses
the half that says the model is too slow.

It exists because nothing else can measure the bot. Play strength,
the funnel's absorption rate, whether the model is being truncated,
whether a blunder was the heuristic's fault or the model's — all of
those need the evidence of individual windows, and until now that
evidence lived for a microsecond inside one function call. Because a
record carries the seat's whole `Input`, a window can be **replayed
offline**: the rules filter and the heuristic are pure functions of
it, so a recorded decision can be re-taken years later by a tool that
never touched a game.

`CMDCTRL_BOT_DECISION_LOG_MODE` sets how much is kept:

| Mode | What it writes |
|---|---|
| `escalated` (default) | Every window. The full board view only for windows that left Layer A; the rest keep their move list and trace. On the `heuristic` tier that is about 85% of windows compacted; a policy that never runs Layer A saves nothing. |
| `all` | Every window, with the full board view. Roughly 38 KiB per window at two seats, 59 KiB at four — a 12-turn two-seat game is ~23 MiB, and a four-seat game to a winner is 100–250 MiB. |
| `model` | Only the windows that actually reached a model, with the full view. The mode for reviewing a model tier's play. |

Writing happens on one background goroutine per game, fed by a bounded
queue; a bot seat hands over its record and returns. The log never
slows the table down, and the cost of that is that it can lose lines:
if the seats outrun the disk the record is dropped and counted, the
same way the byte cap does. `Stats` says which cause.

The static half of a model prompt — the rules primer and the deck
list, several KiB, identical on every window — is written **once per
file** and carried by a `system_hash` on every later record. A reader
walking the file in order keeps the blocks it has seen by hash; the
per-decision half of the prompt is always present.

One game's file is capped at 256 MiB; past the cap records are dropped
and counted, with a single WARN. A long four-seat game can reach that
cap, so this is a real limit, not a theoretical one. Nothing rotates
these files — the operator who turns them on cleans them up. The
directory is created `0700` and each file `0600`.

**Why operator-only.** Each individual record is the seat's *own*
filtered view — the same bytes a human in that chair receives — so no
record leaks anything its seat could not see. The *file* is the
problem: it aggregates every bot seat at the table, so reading it end
to end shows several hands at once. That is fine for an operator
debugging their own server and is not something to serve over HTTP,
paste into an issue, or leave on a shared box. It is off unless you
turn it on.

Whole-game tests have the same knob under `AISEAT_DECISION_LOG=<dir>`,
which is how the position corpus gets harvested.

---

## Measuring the bot

### Position suite

A **position** is one frozen decision window with a human's answer
attached: the seat's own `GameView`, the exact list of legal moves it
was offered, and a label saying which of those moves are right, which
are specifically wrong, and why. They live one JSON file per position
under `server/internal/aiseat/suite/testdata/positions/`.

The answer is stored as **matchers, never as an index** — a move's
label, a regexp over it, a `legal.Kind`, or an action type plus a
subset of its params. The index is derived when the file loads. That
is the detail the whole suite rests on: a stored index stays correct
only until the enumerator's ordering changes, at which point every
position would silently start grading a different move while still
reporting a number. A matcher either still matches or it does not, and
one that matches nothing **fails to load**, naming the position.

Each position also carries `gate`: the policy names for which a miss
**fails `go test`**. `aiseat/suite`'s own test runs the `heuristic`
tier over the whole suite on **every CI run** — it is pure computation
over frozen JSON, so it costs milliseconds and needs no endpoint — and
a gated position is a pinned invariant ("do not chump-block at 40
life") or a pinned bug that can never come back quietly. Model tiers
run through the `boteval` binary, because they need a server to talk
to.

**Harvest, label, run.** Build the binary with
`make -C server build-boteval` and run these from `server/`:

```bash
# 1. play some games with the decision log on (any whole-game test works)
AISEAT_GAME_TESTS=1 AISEAT_DECISION_LOG=/tmp/dl \
  go test ./internal/aiseat -run TestFourHeuristicBotsPlayToAWinner

# 2. pull the interesting windows into an inbox of UNLABELLED positions
./bin/boteval suite harvest --from /tmp/dl --to /tmp/inbox --escalated
./bin/boteval suite harvest --from /tmp/dl --to /tmp/inbox --disagree --limit 20 --seed 1

# 3. look at one the way the model would, and decide what the right move is
./bin/boteval suite render --pos /tmp/inbox/<id>.json

# 4. fill in expected.accept / expected.reject, move it into the suite, run it
./bin/boteval suite run --policy heuristic --md
./bin/boteval suite run --policy assisted --max-think 20s --out report.json
```

`harvest` filters by escalation, heuristic/model disagreement, fallback
cause, layer, seat and tag, and samples deterministically under
`--seed` when `--limit` cuts, so the same logs always produce the same
inbox. Every harvested position arrives with `expected` empty; an
unlabelled position is reported as `skipped` and is never counted as
agreement.

`render` prints the exact system blocks and user delta the model would
be shown, then the move list with its real indices and `<- accept`,
`<- reject`, `<- heuristic` and `<- model@capture` markers. It rebuilds
the prompt from the frozen input rather than replaying a recorded one,
which is what lets it render a position harvested from a game no model
ever played in.

`run` reports agreement overall and per tag, plus reject-hits and the
funnel's own failure counts — malformed replies, out-of-range indices,
timeouts. Those are read off the funnel's classification rather than
recomputed, so a window where the model produced nothing usable is
counted as a model failure even though the heuristic underneath
answered it correctly. Counting that as agreement is exactly how a
broken endpoint would hide. `run` exits non-zero when a gated position
misses.

**Labelling rules of thumb.** Label only windows where the right move
is unambiguous to a competent player — make the land drop, block when
the alternative is lethal, do not chump-block at 40 life, never target
yourself with a burn spell. Write the one-line `note` that explains the
label; a position nobody can review is a position nobody will trust.
Gate a policy only once you have run the suite and seen it pass.

---

## Known limitations

Stated plainly, because most of them are design decisions rather than
bugs.

**Play quality is capped by catalog coverage, not by the policy.** At
a few hundred implemented cards, a bot is a competent player of a
deliberately small format. A better model does not move this ceiling;
more cards do. This is the honest expectation to set.

**No politics, no deal-making, no bluffing, no table talk.** Bots
speak only to disclose an improvisation and, behind the setting
above, to explain a move. Commander is a political format and a bot
does not play that half of it. A four-bot table is four players who
never negotiate.

**No learning across games, and no opponent modelling.** Bots are
stateless between games. The one played you last night remembers
nothing about it.

**No deckbuilding.** A bot never builds, tunes or swaps a deck. In the
lobby, picking one of the four curated decks is the whole of the
customisation. The API also takes a pasted decklist instead of a
curated one (`{tier, format, source}`, [above](#under-the-hood)). The
test harness uses that path, and so can anyone trying a list that is
not in the picker. A pasted list is validated exactly like a human
upload, and its uncatalogued cards are listed in the response's
`unimplemented`. It is not covered by the curated decks' build-failing
test, though. The heuristic prices a card the engine does not
implement as a card spent for nothing, with no credit for what it
would have done ([ADR 0037](decisions/0037-unimplemented-card-signal.md)).
The model tiers' prompt marks such cards too.

**On a one-model deployment, `strong` and `assisted` ask the same
model.** `CMDCTRL_BOT_FRONTIER_MODEL` defaults to `CMDCTRL_BOT_MODEL`,
so a local setup that names one model puts it in both slots. There
`strong` is not a better model. Both tiers already ask the model about
every window that survives the rules filter. `strong` sends each one as
an escalated request, over a wider candidate list. `assisted` escalates
only the windows that trip an escalation trigger. With a local endpoint,
`CMDCTRL_BOT_MAX_THINK`'s 20s default replaces both tiers' deadlines,
so the 5s-versus-2s difference goes too. Name a second, stronger model
in `CMDCTRL_BOT_FRONTIER_MODEL` if you want the gap between the tiers
to be the model as well.

**`strong` has no lookahead, and will not get one here.** The tier was
specified as a one-ply simulation over the top candidates. Simulating
a move means cloning a game; cloning a game means holding a
`*game.Game`; and a policy may not hold one — that is the
hidden-information guarantee in [ADR 0033 §3](decisions/0033-ai-bot-seat.md),
enforced by an import test that fails the build over it. The two
requirements are incompatible and the guarantee is the more important
of the pair. What `strong` actually buys is the frontier model on
every surviving window and a wider candidate list.

**No Voltron and no combo deck.** See the deck section above.

**Bots do not concede lightly.** The concede heuristic is deliberately
conservative, on the grounds that a human playing a bot generally
wants the finish.
