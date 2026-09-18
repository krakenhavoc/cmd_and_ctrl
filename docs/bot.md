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

### What a bot gives up

**A prompt that asks a bot to name cards in its own hand is scored by
what the answer KEEPS**, so the bot pitches the cards it values least
rather than the first ones it happens to be holding. Every effect
discard arrives as that one prompt — Mind Rot's forced two, a loot's
discard after the draw, a rummage's before it — and the valuation is
the same one the cleanup-step discard to hand size uses, so "the worst
card in hand" means one thing wherever a bot has to give a card up.
Keeping is never worth less than nothing, so a prompt that says "up
to" is answered with as few cards as it will accept.

The rule is about giving cards up, and it reads every hand prompt that
way: asked to name a card from hand for something *good* — "you may
put a land from your hand onto the battlefield" — a bot names its
worst one, or, where the prompt allows it, declines. A prompt over
anything but the bot's own hand (the battlefield, a library, an
opponent's hand) carries nothing on the wire that says what naming a
card costs, so the bot takes the first answer offered.

### What a bot answers when somebody else's card asks (#796 / #568)

Two prompt shapes reach a bot seat from a spell or ability it does not
control, and neither is a card-from-hand pick, so the section above
does not price them.

**A free yes/no at resolution (`confirm`, #796).** `MayChoice` is the
"you may [do X]. If you do, [Y]" a resolving effect asks — Eden's
sacrifice, Mask of Memory's optional draw, Combustible Gearhulk's
question to its target. It rides the existing `confirm` kind, so the
policy is the one `confirm` already has: both branches are offered, the
accept branch carries whatever life the card charges as
`LegalMoveView.cost.life`, and the DECLINE is the kind's always-legal
answer. A bot therefore never takes a life payment it cannot see the
price of, which is the #547 rule pointed at a prompt instead of an
activated ability.

**An option pick (`option_pick`, #568).** "Choose one of the
following", addressed to any seat: Torment of Hailfire's three-way
question, and the pile a Fact or Fiction chooser takes. Every branch is
enumerated, in the card's printed order, and each carries its declared
life cost — so "lose 3 life" and "discard a card" are priced
differently and a bot at 3 life is not handed a way to kill itself for
free. **The first option is the always-legal one**, by the kind's own
contract: an effect builds its option list out of what this seat can
actually do (CR 608.2), and the branch it puts first is one that never
fails ("lose 3 life", which needs no permanent and no card in hand).
A policy with nothing better to say takes it, which terminates.

**A pile split** needs no policy of its own. Its first half is an
ordinary `choose_cards` prompt over public, revealed cards, so the
existing card-set enumeration offers the subsets — including the empty
pile, which is this prompt's always-legal answer — and its second half
is the option pick above, two branches, each labelled with its pile's
size. A heuristic that takes the first offer splits and then takes
pile one; that is a weak split rather than an illegal one, and it
terminates, which is the bar this list exists to clear.

The general rule behind all three: a prompt from somebody else's card
carries nothing on the wire that says what an answer is WORTH beyond
its declared cost, so a bot prices what it can see and takes the first
offer otherwise — the same posture the paragraph above takes for a
prompt over anything but the bot's own hand.

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

Nothing in the product measures play strength, and nothing can: a
table is one game, and one game of Commander tells you almost nothing
about a policy. The measurement lives in a separate local binary,
`boteval` (`make -C server build-boteval`), because the runs that
matter need a model endpoint, a Scryfall dump and minutes of wall
clock, none of which CI has.

### Arena

`boteval arena` plays N headless bot-vs-bot games and prints the
report block [ADR 0052](decisions/0052-bot-decision-harness-and-eval.md)
asks every bot PR to carry.

```
boteval arena --seats assisted,heuristic,heuristic,heuristic \
              --decks izzet-aggro,simic-ramp,esper-control,mono-black-aristocrats \
              --games 10 --seed 1 --rotate \
              --model qwen3:14b --endpoint http://192.168.1.18:11434/v1 \
              --max-think 20s --dump ../data/scryfall/default-cards.json \
              --out ./arena-out --decision-log
```

There is no server, no websocket and no client: the arena builds a
`game.Game`, wraps it in the same `ws.Room` the server uses, and
starts one ordinary `aiseat.Runner` per chair. Every seat sees exactly
the filtered `aiseat.Input` it would see at a real table.

| Flag | What it does |
|---|---|
| `--seats` | one tier per chair, comma-separated. 2–4 chairs. |
| `--decks` | one curated deck id per chair, or none at all — a partial list is refused. No `--decks` deals a synthetic 65-card red deck that needs no Scryfall dump. |
| `--names` | one tally name per chair. Use it when every chair is the same tier and the thing being compared is the deck or the configuration. |
| `--games`, `--seed` | game *i* uses `seed+i`, so a run is exactly reproducible and two policies can be compared on the same deals. |
| `--rotate` | moves each contestant one chair along per game (contestant *k* sits at position `(k+i) mod n`). **Use it.** Turn order in Commander is worth real percentage points; without rotation you are measuring the chair. |
| `--turn-budget`, `--wall`, `--stall` | when to stop a game that will not end (default 60 turns, 30 minutes, `3×max-think+15s` with no committed move). |
| `--max-think`, `--model`, `--frontier-model`, `--endpoint` | the model tiers' deadline and transport. With an endpoint set and no `--max-think`, the deadline defaults to **20s**, the same local default `cmd/server` applies and for the same reason. |
| `--out` | artifacts directory. Each run gets its own `<out>/<RFC3339 start>/`. |
| `--decision-log`, `--decision-log-mode` | per-game decision logs under `<out>/decisions`. **Operator-only** — see the section above. |
| `--replays` | per-game replay JSONL under `<out>/replays`. Off by default: a four-seat replay is ~320 MiB — and this flag hands the directory to `ws.Room`, which also writes `games/<id>.json` (the full authoritative state, rewritten on **every committed move**) and `restore/<id>.json` while a game is live. Budget for all three. |
| `--block-grace` | how long an attacking bot holds its pass in declare-blockers while a defender still has a legal block. Defaults to production's value. `0` turns it off, which is faster and declares systematically fewer blocks than a real table — see below. |
| `--md`, `--json` | what to print. Markdown by default. |

Env fallbacks match the server's: `CMDCTRL_OPENAI_ENDPOINT`,
`CMDCTRL_OPENAI_API_KEY`, `CMDCTRL_BOT_MODEL`,
`CMDCTRL_BOT_FRONTIER_MODEL`, `CMDCTRL_BOT_MAX_THINK`,
`CMDCTRL_SCRYFALL_DUMP`.

**A model tier with no endpoint is refused, not downgraded.** An
`assisted` seat with no client plays Layer A + B and still calls
itself `assisted`, so the run would report a heuristic's win rate
under a model's name. That is the one misconfiguration that corrupts
the measurement instead of breaking it, so the arena will not start.

**A stall is reported, not fatal.** A table that stops committing
moves still played twenty turns of real decisions, and the dump of
every seat's legal moves at the moment it froze is the evidence that
says why. The game is marked `stalled`, counted, and the run
continues.

**`games.jsonl` is operator-only, for the same reason the decision log
is.** A stall dump enumerates every seat's legal moves, and a cast
move is enumerated from that seat's hand — so the file names cards in
hands the reader was never entitled to see. It is written `0600`
inside a `0700` run directory, and it is exactly the file someone
would reach for when reporting a stall. Quote the structural head of a
dump (step, pending kinds, per-seat life and move counts) in a bug
report; do not attach the file. `summary.md` and `summary.json` carry
no dump and are safe to paste.

**Turn order only cancels over a whole number of rotations.**
`--rotate` seats contestant *k* in chair `(k+i) mod n` for game *i*,
which balances turn order exactly when `--games` is a multiple of the
seat count. It is not, the arena prints a warning naming the
imbalance and does **not** silently change `--games` — and the run's
chair histogram goes into the report either way, so a finished run is
auditable on the point. Turn order in Commander is worth real
percentage points, the same order as the differences being measured,
so prefer 12 games over 10 on a four-seat table.

**`--block-grace` is a fidelity knob, not a speed knob.** Without it
an attacking bot re-steps the moment it commits, and the step can end
before a slower seat has declared a block; combat is where policy
differences actually show, so a run with it off understates every
difference. It defaults to production's value and should stay there
for any number that goes into an ADR.

#### Wall clock

| Table | Per game |
|---|---|
| 2 heuristic seats, synthetic deck | ~0.2 s |
| 4 heuristic seats, curated decks | ~3.5 s |
| 4 random seats | seconds |
| 1 assisted seat (local 14B, 20s think) | 5–15 minutes |
| 4 assisted seats on one GPU | they serialise; budget 4× |

A ten-game baseline with one `assisted` seat is an hour or two, so
`games.jsonl` is written **as each game ends** and Ctrl-C stops
between games rather than killing the run.

#### Reading the report

`summary.md` is the block that gets pasted into a PR; `summary.json`
is the same numbers for a tool; `games.jsonl` has one full result per
game.

- **Play** — seat-games, decided seat-games, wins, win %, a Wilson
  95% interval, and the **null rate** (1/seats). The interval and the
  null are the whole point: "assisted won 30% of a four-seat table" is
  not a result, because the null is 25% and ten games cannot tell them
  apart. The `beats null` column is `yes` only when the entire
  interval clears the null. Expect it to say `no` for a long time.

  **Win % is over *decided* seat-games**, not all of them. The null is
  P(win | somebody won) — on a four-seat table exactly one chair takes
  a decided game, so the four rates sum to 1 — and counting undecided
  games in the denominator would scale every policy down while leaving
  the null where it is. On a run where 40% of games hit the turn
  budget, every policy's ceiling would be 0.60 and a policy winning
  45% of the games that ended would report 27% and `beats null: no`.
  `seat-games`, `draws` and `stalled seat-games` stay in the table so
  the undecided share is visible rather than buried.

  One more caveat the footnote repeats: two chairs of the same policy
  contribute two seat-games to one game and at most one of them can
  win, so those trials are negatively correlated. The interval treats
  them as independent, which makes it **conservative** — it will not
  manufacture a `beats null` — but `seat-games` is not a count of
  independent trials.
- **Funnel** — windows by layer, escalations, model calls, timeouts,
  fallback reasons, tokens, median prompt size. A model tier whose
  every window fell back to Layer B has the heuristic's win rate and a
  completely different explanation; this is where that shows up.
- **Latency** — decision p50/p99/p999/max and model-call p50/p99/max
  per policy. Read the tail, not the median: a seat whose p999 is nine
  seconds hangs the table twice a game and its mean says nothing about
  it.
- **Games** — one line each, so a stall or a runaway game is findable.

With `--decision-log` on, the header line also carries the writer's
own tally. The writer is asynchronous and **drops rather than blocks a
bot seat**, so a run can end with a corpus that has holes in it; the
report says `nothing dropped`, or names the cause (`queue` — the seats
outran the writer, `cap` — the 256 MiB per-game limit, `error`). A
position suite harvested from a run with drops is missing windows.

Record the model id and quantisation with every run (`--note` puts a
line in the report); the build's git revision is stamped
automatically.

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

## Enumerating a modal announcement (#764)

A modal cast or activation is a product: every legal selection of
modes, times every legal set of targets for each clause of each
chosen mode. That product is unbounded in principle and is bounded in
practice by `legal.Options.MaxExpansionPerSource` (default 12, [ADR
0033](decisions/0033-ai-bot-seat.md) §1). The policy for spending
that budget is stated here because it decides what a bot is even
allowed to consider, and [ADR 0065
§6](decisions/0065-modal-and-multi-target-clauses.md) is where it was
decided:

- **Prefer the modes that have legal targets.** An option whose
  clause cannot be filled from the current board is dropped before
  any combination is built, so the budget is never spent on a
  selection the engine would refuse at announce. This is the same
  `ChoosableModeOptions` walk the `mode_pick` prompt uses, so the
  enumerator and the prompt offer the same bullets.
- **All-one-mode first for a repeatable spec.** With CR 700.2d in
  play (Mystic Confluence's "you may choose the same mode more than
  once") the selections that take one bullet `Max` times are emitted
  before the mixed multisets. When only one bullet is legal, "that
  bullet three times" is the only selection there is, and it must not
  be crowded out by mixtures the seat cannot take.
- **Modes outermost.** The budget is spent mode-selection first, so
  every selection gets at least one target set before any selection
  gets a second. Without that, one charm's first bullet with twelve
  targets would be the whole move list and the other three bullets
  would never be offered.

A `mode_pick` prompt is enumerated the same way: `choiceMoves` offers
every legal multiset of the bullets the prompt carries, capped by the
same budget, and labels each move with the bullets rather than their
indexes so the decision log reads.

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

**A bot is never offered an {X} spell or ability at X=0 when X is the
whole of what it does.** Soothsaying's "{X}: Look at the top X cards of
your library" is free at X=0, does nothing, and is back on the list the
moment it resolves — so a table of bots took it 79,519 times in five
minutes and never got past turn 18
([#810](https://github.com/krakenhavoc/cmd_and_ctrl/issues/810)). The
enumerator's rule is one line of policy: the smallest X it will
announce for such a cost is 1, and where X=1 cannot be paid for the
move is not offered at all. A card with a fixed rider — The Goose
Mother is a 2/2 flier before X buys anything — is still offered at X=0,
and then only when nothing larger is affordable, because the enumerator
always takes the largest X the seat can pay. Which cards are which is
the catalog's declaration (`Spec.XMatters`), not a guess. CR 602.2b
still makes X=0 a legal announcement and the engine still accepts one;
this is about what is worth putting in front of a player.

**A bot takes one CR 726 shortcut per loop per turn, then stops.** When
the loop breaker fires ([ADR 0055](decisions/0055-loop-breaker.md)) the
repeating ability's controller is asked how many more times it should
resolve. `internal/legal` offers a bot seat 10 first — and a bot takes
the first offer — then 100, then "stop here"; on the turn's **second**
ask for the same loop it offers nothing but stop. So a bot-only table
that finds a real loop runs the threshold's worth of iterations, then
ten more, and comes to rest with the notice naming the ability. The
guarantee is in the enumerator rather than in a policy on purpose: a
policy that can rank "100 more" top can rank it top every time, and a
random one eventually will. A human at the same prompt types any
number up to 1000 into the client's field.

**A loop the bot is feeding itself gets "stop" on the first ask, and
the bot stops activating.** An activation loop is the other shape of
runaway: nothing repeats on its own, the seat just keeps taking the
same free ability. "Resolve it ten more times" is no kind of shortcut
past a crank somebody has to keep turning, so for a loop whose
repeating ability is an activated ability of the chooser's own
permanent the enumerator offers only "stop here" — and while the notice
stands, a bot runner holds on an activation of that permanent exactly
as it holds on a pass. The table comes to rest at the threshold with
the notice naming the ability, which is [ADR 0055
§5](decisions/0055-loop-breaker.md)'s outcome for a bot-only table.
