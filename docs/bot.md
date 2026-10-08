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

A seat you run yourself, played by an AI agent through MCP rather than
by the server, is a different thing: see [docs/mcp-seat.md](mcp-seat.md).

> **What is actually wired today.** All four tiers and all four
> curated decks are reachable from the lobby. `random` and `heuristic`
> need nothing but the binary; `assisted` and `strong` call a model,
> and a server with no model endpoint configured reports them
> unavailable — with a reason the picker shows — rather than offering
> a seat that would play the heuristic under a model tier's name. See
> [the model endpoint](#the-model-endpoint) for the one environment
> variable that turns them on.
>
> **Improvisation is live on the model tiers.** An `assisted` or
> `strong` bot that casts a card the rules engine cannot run will
> apply the card's text **by hand**, announce it in chat, and leave it
> undoable by any player for free — see
> [below](#improvisation-and-why-an-undo-is-free) for exactly when,
> and for the one environment variable that turns it off. `random` and
> `heuristic` never improvise: neither has a model, and reading a
> card's oracle text is the whole job.

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

`Start` **subscribes the seat to the room before it returns**, on the
caller's goroutine rather than on the runner's. That is what makes
"the bot is seated" a fact you can order other things against: nobody
schedules the runner goroutine, so a subscription taken inside it left
a window — as wide as a loaded machine cares to make it — in which a
commit reached every other seat at the table and not this one. The
runner would still see the *effect* of that commit, because its first
look is at the live game rather than at a payload; what it lost was the
WAKE. A seat that wants nothing from a window parks until the next
commit, so a commit that landed in the window is one it would never be
woken for, and the last bot seated at a busy table could sit out the
rest of the game (#938).

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

**The table's own pace setting overrides both, per decision.** ADR
0075's host controls add a `bot_pace` table setting —
`fast` / `normal` / `slow` — and the runner re-reads it before every
single decision a bot makes, not once when the seat was added, so a
host changing it mid-game is live on the bot's very next move:

| Table pace | `MinThink` | `MaxThink` | Stack hold |
|---|---|---|---|
| `fast` | 0 | 2s | 0 |
| `normal` (default) | 700ms | 2s | 2s |
| `slow` | 2s | 8s | 3s |

The **stack hold** ([ADR 0119](decisions/0119-a-stack-you-can-follow.md)
§2) is how long another seat's spell or ability must have been on top
of the stack before the bot passes on it, so the people at the table
can read it. It is measured from the commit at which the bot first saw
the item, and it overlaps `MinThink` rather than adding to it. A bot
passes on its own item at once (CR 117.3c), and if the top of the stack
changes while it waits it decides again. Every seat measures from about
the same moment, so a table of bots holds once, not once per seat. A
seat built with a hand-made `Config` (tests) or stepped in lockstep
(`boteval arena --lockstep`) never holds.

`strong`'s longer 5s deadline is never shortened by the table pace —
a `fast` table still gives a model-backed seat its full budget, and
`MaxThink` is always the larger of the table's preset and the tier's
own deadline. Only `slow` (8s) lengthens it further. A game whose
settings were never given a pace (an old snapshot, or a raw `Game`
built outside the normal lobby path) behaves exactly as before this
setting existed.

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
worst one, or, where the prompt allows it, declines.

**A prompt that asks a bot to name cards in its own library is scored
by what the answer TAKES (#1831)** — the opposite sign. Every such
prompt is a look at the top of the library that moves the named cards
somewhere the bot wants them: into its hand (Horn of the Mark, Explore
the Vastlands), onto the battlefield, or under a hideaway land. So the
bot takes the most valuable card it may, by the same valuation the
library search uses; "any number" takes them all, and a "you may" is
never declined while there is a card to take. Which sets are legal
(Explore the Vastlands' "a land and/or an instant or sorcery") is the
enumerator's business, so the bot only ranks the answers it is offered.

A prompt over anything else (the battlefield, a graveyard, an
opponent's hand or library) carries nothing on the wire that says what
naming a card costs, so the bot takes the first answer offered.

### Automatic answers are for people (ADR 0127)

A person can tell the server to answer a repeated prompt for them —
"never pay for Rhystic Study", "always draw two off Consecrated Sphinx"
([ADR 0127](decisions/0127-answering-repeated-prompts-for-you.md),
`set_auto_answers`). Bots do not use it. A bot answers its own prompts
through the enumerator, as it always has: the enumerator never offers
`set_auto_answers`, the server refuses rules for a bot seat, and the room
checks that a seat is human before it answers anything for it. A
person's automatic answer reaches a bot as an ordinary state change and
an `auto_answer` log line.

### What a bot answers when somebody else's card asks (#796 / #568 / #929)

Three prompt shapes reach a bot seat from a spell or ability it does
not control, and none of them is a card-from-hand pick, so the section
above does not price them.

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

**A choose-a-player prompt (`option_pick` again, #929).** "Choose a
player" / "choose an opponent" — Gluntch, Skullwinder, Slithermuse —
is an option pick whose branches are SEATS, one option per eligible
player, labelled with that player's name. Nothing is hidden and
nothing costs life, so there is no price to read; what decides the
answer is the ORDER, and the order is the policy: **the eligible seats
are offered most life first**, ties by seat, so the always-legal first
offer a policy takes with nothing better to say is the player with the
highest life total. `Game.QueueChoosePlayerForEffect` does the
ordering, which is why there is no second copy of this judgement in
`internal/aiseat`.

It is a legality-and-termination policy, not a strength one, and the
weakness is worth naming: the best pick is frequently not the
highest-life seat (Slithermuse wants the opponent with the fullest
hand; Skullwinder wants the one with the worst graveyard). A card
whose clause makes that gap matter should change the ordering the
prompt is queued with, not teach a policy to special-case the card.

**A pile split** needs no policy of its own. Its first half is an
ordinary `choose_cards` prompt over public, revealed cards, so the
existing card-set enumeration offers the subsets — including the empty
pile, which is this prompt's always-legal answer — and its second half
is the option pick above, two branches, each labelled with its pile's
size. A heuristic that takes the first offer splits and then takes
pile one; that is a weak split rather than an illegal one, and it
terminates, which is the bar this list exists to clear.

The general rule behind all four: a prompt from somebody else's card
carries nothing on the wire that says what an answer is WORTH beyond
its declared cost, so a bot prices what it can see and takes the first
offer otherwise — the same posture the paragraph above takes for a
prompt over anything but the bot's own hand.

**Which permanents untap (`untap_choice`, #826, CR 502.3).** The one
card-set pick whose sign is never in doubt. Under a Winter Orb or a
Static Orb the untap step stops and asks the active player which of
their permanents untap, and a permanent the bot names is a permanent it
gets back — so the heuristic scores an answer as the total
`permanentValue` of the cards it names and takes the best, the same
valuation the sacrifice prompt uses with the opposite sign. That
settles the count as well as the choice: untapping is never worth less
than nothing, so a bot under a cap takes a full legal set rather than a
short one, and among full sets the most valuable.

It is a greedy pick over legal sets, not a plan — the bot does not know
which colours its hand will want three spells from now, and the wire
carries nothing that would say. What it does guarantee is that the
answer is always legal and the table always moves: the enumerator
offers only sets the engine accepts (the cap solver runs inside
`ChooseCardsPickLegalLocked`), and this is the one prompt that can open
in a step where nobody holds priority, so a seat with no answer here
would stop the game outright rather than merely stall its own turn.

**Riot and unleash (`entry_riot`, and `optional_replacement` with
`entry_keyword: "unleash"`, ADR 0109 §10).** Riot takes haste when the
creature enters during the bot's own turn before attackers are declared
and could attack (it has no defender), and the +1/+1 counter otherwise.
Unleash always takes the counter, unless it is an opponent's turn and the
entering creature would be the bot's only untapped creature, which a
counter would stop from blocking (CR 702.98a). See
`aiseat/heuristic/riot.go`.

**Sending a commander home (`commander_return`, [ADR 0115](decisions/0115-commanders-die.md),
CR 903.9a).** When the bot's commander has been put into a graveyard or
exile, the state-based action asks its owner whether it goes to the
command zone. The bot says yes, unless the prompt's `playable_from_zone`
is set, meaning it could cast the commander from where it is (escape,
flashback, an adventure in exile). Then it says no and leaves the card
there (owner decision 2). The model prompt names the card and says when
it is castable where it is. Since ADR 0115 PR 3 this is how a destroyed,
sacrificed, countered, discarded, milled or exiled commander is asked;
the `optional_replacement` above is left for a commander headed for a
hand or a library (CR 903.9b).

**The other "may" replacements (`optional_replacement`, `entry_pay_life`,
`copy_target`, #2390).** Until #2390 the heuristic said yes to every
optional replacement, so once dredge (#2388) used that prompt a bot with
a dredge card in its graveyard dredged on every draw. Each one now has an
answer of its own (`aiseat/heuristic/replacement.go`, `dredge.go`):

- **Dredge** (the prompt carries `dredge`, the N it mills, and the card in
  `source`). The yes is worth the card in hand, plus `DredgePlaySoon` if
  the bot could play it this turn or next, plus N milled cards at the mean
  fuel value of the bot's graveyard (so a graveyard it casts from makes
  milling worth more), less what the card was worth in the graveyard. The
  no is the draw, priced at the mean value of the cards the bot has already
  seen from its library, so a bot short of lands values its draw higher. A
  dredge that would leave fewer than `DredgeLibraryFloor` (10) cards in the
  library is declined, whatever it returns. The Necrobloom's grant names no
  card; the bot prices it as its best land card in the graveyard.
- **A commander headed for a hand** (CR 903.9b; the prompt sets
  `playable_from_zone`): no. In hand it is cast without the commander tax,
  which is owner decision 2 of ADR 0115 applied to the one zone a commander
  can always be cast from. Headed for a library, or asked before a cost is
  paid: yes.
- **Library of Leng and Moonlit Meditation**: yes. Leng's prompt does not
  name the discarded card (it is in a hand), so the bot cannot price it, and
  keeping a card beats losing it. Moonlit Meditation enchants a permanent
  the bot chose, and the copies are the Aura's only text.
- **Unleash**: unchanged, see riot and unleash above.
- **A shockland's "pay 2 life"** (`entry_pay_life`): the pay move now carries
  its life as `cost.life`. The bot pays only when the untapped land lets it
  cast more before the land would untap anyway (spells and its commander on
  its own turn, instants and flash otherwise), and only when the mana is
  worth more than the life at `LifeCostValue`'s price, which rises steeply
  near death. At 2 life it never pays 2.
- **"Enter as a copy of"** (`copy_target`): the most valuable permanent on
  offer, as the bot's own. It used to take the first answer, which declines,
  so a Clone entered as a 0/0 and died. A legendary permanent the bot
  already controls is passed over (the legend rule keeps one), which is
  weaker than Spark Double but never a dead copy.

Mox Diamond, the reveal lands, the sacrifice lands and devour have prompt
kinds of their own (`entry_discard_from_hand`, `entry_reveal_from_hand`,
`entry_sacrifice`), priced in `choices.go`, and never reached the old
default.

**Devour (`entry_sacrifice` with `devour`, CR 702.82a, #2419).** "You may
sacrifice any number of creatures; it enters with N +1/+1 counters for
each" is the one sacrifice prompt whose empty answer is legal, so the
bot prices each creature on its own (`heuristic/devour.go`): N counters
at `Weights.Power + Weights.Toughness` each, plus what the creature's own
"for each creature it devoured" ability pays per creature (a card at
`drawValue` for Skullmulcher, `Weights.Life` per point for Marrow
Chomper), minus what the creature is worth on the board, a premium for a
nontoken (`DevourPermanent`), and a large one for its own commander
(`DevourCommander`). A set is the sum of its creatures, so the bot eats
every creature worth less than its counters and stops there; tokens go
first, and a 1/1 token is only eaten for N of two or more. The
enumerator orders the pool cheapest first, so the best set is always
among the offered prefixes. The fixed-count sacrifice lands carry no
`devour` and keep the cheapest-set rule.

**Whom to give it to (`entry_controller`, ADR 0102, CR 614.12a).** "This
enters under the control of an opponent of your choice" is offered as
one answer per opponent. The heuristic reads the prompt's
`control_purpose`: a harmful permanent (Captive Audience, Xantcha) goes
to the opponent with the highest `SeatEval.Strength`, a helpful one
(Pendant of Prosperity) to the lowest. It does not use `Threat`, which
ranks a seat by how close it is to dying — "your life total becomes 4"
costs a seat at 5 life almost nothing.

**Whether to reveal (`entry_reveal_from_hand`, #1198, CR 614.1c).** The
one card-set pick whose answer is FREE, and the reason it is a branch
of its own rather than a second spelling of `choose_cards`. "As this
land enters, you may reveal an Island or Swamp card from your hand. If
you don't, it enters tapped" names a card and takes nothing: revealing
is not a zone change, so the card is still in hand afterwards, and what
it buys is an untapped land this turn. Both of the valuations this
package already has would get it backwards — the `choose_cards` branch
scores an answer by what it KEEPS, because a card named there is a card
given up, and the cost-fuel pricer (#1028) prices a card EATEN by a
cost. Through either, "reveal nothing" wins and every reveal-land in
the deck comes down tapped forever.

So the branch is flat and unambiguous: naming any card beats naming
none. The count settles itself (every printed member of the family
reveals at most one), and the enumerator offers the decline FIRST, so
this preference is exactly what breaks a tie the "take the first offer"
default would otherwise decide the wrong way.

What the bot gives up is information — the table learns one card in its
hand — and that is deliberately not priced. Against an untapped land on
curve it is the trade a human takes almost every time, and pricing it
would need an opponent model this policy does not have. A bluff (holding
a matching card and declining anyway, to keep the hand hidden) is a real
play and is out of reach for the same reason.

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
| `CMDCTRL_BOT_THINK` | `1` lets the model think before it picks a move. An experiment, off by default — see [Letting the model think](#letting-the-model-think-2196). |
| `CMDCTRL_BOT_MAX_TOKENS` | The decision calls' reply budget in tokens. Unset keeps 128 routine / 256 frontier, or 8000 with `CMDCTRL_BOT_THINK` on. |
| `CMDCTRL_BOT_IMPROVISE` | `0` turns [improvisation](#improvisation-and-why-an-undo-is-free) off. On by default for the model tiers. |
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

### Letting the model think (#2196)

`CMDCTRL_BOT_THINK=1` (or `--think` on `boteval probe`, `suite run`
and `arena`) turns thinking back ON for the decision calls, as an
experiment. It is off by default and nothing about a default
deployment changes. With it on:

- the client stops sending `think: false` and `reasoning_effort:
  "none"`, and sends `think: true`;
- the reply budget rises from 128 / 256 tokens to
  `CMDCTRL_BOT_MAX_TOKENS` (`--max-tokens`), 8000 when unset, because
  the model writes its thinking out of the same budget;
- the deadline, when `CMDCTRL_BOT_MAX_THINK` (`--max-think`) is unset,
  is 120s rather than the local 20s. `CMDCTRL_BOT_MAX_THINK` still
  wins when it is set, and it can only widen a tier's deadline.

Measured on Ollama 0.35.1 with `qwen3.6:35b-a3b`: thinking works WITH
the `json_schema` constraint. The thinking arrives in
`message.reasoning`, unconstrained, and `content` is the
schema-constrained answer, written after it. The answer is read from
`content` only; the thinking is never parsed, and its length is
recorded (`Trace.ReasoningChars`, `Stats.ReasoningChars`, the suite's
`reasoning` line, the arena's note under the funnel table). A server
that sends the thinking inline as a leading `<think>…</think>` block
has that block moved out of the answer first.

The failure ADR 0052 §7 names is unchanged: **thinking that reaches
the token cap leaves `content` empty**, `finish_reason` is `length`,
and the window scores as malformed and plays the heuristic's move.
`boteval probe --think` says `THINKING: on, and it RAN OUT OF BUDGET`
when that happens. Raise `--max-tokens`, and the deadline with it: at
about 80 tokens a second on a 16 GB card, 8000 tokens of thinking is
about 100 seconds.

What it bought, measured 2026-10-04 on the 22-position suite with the
dump (`qwen3.6:35b-a3b`, Ollama 0.35.1, 16k context): agreement rose
from a mean of 54% over ten runs with thinking off (range 45–64%) to
62% at 4000 tokens / 60s and 70% at 8000 tokens / 120s (three runs
each, range 55–73% and 68–73%), and reject-hits fell from 8.7 a run to
6.0 and 5.7. The median model call went from about 1s to about 22s,
the median reply from 30 tokens to about 1,700, and 8000 tokens left
no reply cut off (4000 cut one). It is still worse than the heuristic
on mulligans: three of the six are missed on most thinking runs. And
it is too slow to play: one two-seat arena game against the heuristic
had not finished after 17 minutes. Use it to measure, not at a table.

**Probe the endpoint before you trust it.** `boteval probe` sends one
request in the funnel's exact shape and prints what came back: the
endpoint and model it dialled, the prompt's byte size and a rough
token estimate, `usage.prompt_tokens` and `completion_tokens`, the
server's prefix-cache hit count, `finish_reason`, whether the reply
text was empty, whether a `reasoning` field came back, the first 300
characters of the reply, the parsed index and copied label or the
parse error, whether the reply was constrained by a `json_schema`, the
move list it was shown with the answer marked, and the wall time. It
ends with a verdict line for each of the three failures that make a
model seat look like a heuristic seat: **truncation** (the server
counted far fewer prompt tokens than were sent, so it silently dropped
the front of the prompt — the primer and the instructions), **thinking
not suppressed** (a `reasoning` field came back, or `finish_reason` was
`length` with nothing usable in `content`), and the **answer** — which
move the reply comes to, read exactly as a seat reads it, or `INDEX:
out of range` when it names a number that is not on the list and no
label that picks one out (#2196). It always exits 0: it is a
diagnostic, and "the endpoint is down" is a finding.

**The reply names its move twice, and on a local endpoint it cannot
name one that is not listed (#2196).** The model answers `{"index": n,
"move": "<that entry's text>", "why": "…"}`. On the OpenAI-compatible
transport the request carries `response_format: json_schema` with
`index` an enum of exactly the listed indices and `move` an enum of
their labels, so a server that applies the schema as a grammar (Ollama
does) cannot write a number past the end of the list — which was half
of qwen3.6:35b-a3b's replies on the probe's window before this. A
server that 400s on `response_format` is asked again without it, once,
and is not sent it again for the life of the process. Whatever the
transport, the funnel then resolves the reply against the list it
showed: the number when the label agrees or is absent; the label when
the number is off the list or names a different move and the label
names exactly one listed entry; the number again when the label names
none (or names two identical lines); and the heuristic's move when
neither names a move. Neither can select a move that was not shown.
How each used reply resolved is counted in `Stats.ByPick` and the
arena's `picks` column (`index`, `index-unlabelled`, `label-rescued`,
`label-corrected`, `label-mismatch`), next to `out-of-range` under
`fallbacks`.

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

## Which colour a bot names (#780)

Every `choose_color` prompt (CR 105.4) offers five legal answers and, on
its own, says nothing about which one the card wants. So a bot with no
other information named the colour its hand needed most — right for
Coldsteel Heart, right by accident for Selective Obliteration, and a
self-inflicted board wipe for Wash Out, where a mono-green bot named
green and bounced its own permanents while leaving the opposition's
alone.

The prompt now carries the card's own declaration of what it will do
with the answer (`color_purpose` on the wire, `game.ColorPurpose` in the
catalog), and the policy is one function switching on it —
`colorChoiceValue` in `aiseat/heuristic/choices.go`. There are no
per-card branches anywhere in it.

| Purpose | Cards | What the bot names |
|---|---|---|
| `mana` | Coldsteel Heart, the Thriving lands, the Gates, and the one-colour lands (Uncharted Haven, Crossroads Village, Mirage Mesa, Valgavoth's Lair) | the colour its **hand** asks for most — the pre-#780 rule, unchanged |
| `benefit` | Heraldic Banner, Selective Obliteration | the same, with the colour it has most of on its **own board** breaking ties. Selective Obliteration's chosen colour is the one that *survives*, so "name what you want to keep" is the same question the anthem asks |
| `harm` | Wash Out | the colour that costs the **opposition** most net of what it costs the bot, priced with the same `permanentValue` the board evaluation uses |
| `filter` | Oona, Queen of the Fae | the colour most common among the opponents' **known** cards — their battlefield, their graveyards and the stack. Libraries and hands are counts on the wire and stay that way ([ADR 0033 §3](decisions/0033-ai-bot-seat.md)), so this is a proxy, and on a board that has shown nothing every colour ties and the enumerator's order decides |
| `protect` | Mother of Runes (Story Circle and the Circles of Protection are unwritten) | the colour of the **biggest threat** pointed this way: a declared attacker first (doubled, because protection is bought a beat before it is needed), then any creature, then any other permanent, then a spell on the stack |

**An undeclared prompt keeps the old rule.** The purpose is a string,
its zero value is "nothing was said", and the default arm is the
hand-need scoring exactly as it was — so a card nobody has annotated
plays no worse than it did yesterday. What stops one *staying*
unannotated is a catalog guard rather than a runtime check:
`cards/effects/color_purpose_guard_test.go` scans the catalog's source
and fails the build for a `choose_color` prompt whose first argument is
not one of the declared constants, in the style of #806's and #810's
lints.

**The ENUMERATOR orders by the purpose too (#986).** The table above is
the policy's scoring; it is not the only reader. Until #986 the
enumerator offered every colour prompt's answers in one order — "what
this seat has most of on the battlefield" — whatever the prompt was
for, which meant the `random` tier, a heuristic that scored two colours
the same, and any future "take the first offered answer" all got the
Coldsteel Heart answer for a Wash Out. `legal.OrderColorOptionsLocked`
(`server/internal/legal/color_order.go`) now ranks them by the same
purpose, in battlefield counts only:

| Purpose | Ranked by |
|---|---|
| `mana`, `benefit`, undeclared | permanents the chooser controls of that colour — the pre-#986 rule, unchanged |
| `harm` | permanents the OTHER seats control of that colour, minus the chooser's own |
| `filter` | permanents the other seats control of that colour; the chooser's own board is not a term |
| `protect` | the greatest POWER among creatures other seats control of that colour — one 8/8 outranks four 1/1s |

It is an ORDERING and never a filter: every answer the engine accepts
is still offered, so the policy still sees all five and still re-ranks
them. It is not a second scorer either — it is battlefield counts, not
`Weights`, and it is deliberately the crudest public proxy for each of
the questions `colorChoiceValue` asks properly. What it fixes is the
seats that have no policy: the `random` tier, a tie-broken heuristic,
and the human, whose `color_options` buttons come off the same call so
that the first button and the first enumerated move are always the same
colour.

This is a FUNCTION and not an `Options` hook, which is the difference
from #687's `OrderTargets`. Ranking a board by what a spell is worth
against it is a policy question and `legal` may not import `aiseat`, so
the target ordering is injected by the seat. A colour prompt's order is
a reading of the card's own printed text against public counts, it has
to be identical for a bot and for a human because both read it off the
same prompt, and `legal` is the layer both already go through.

## How the heuristic prices a card

[ADR 0126](decisions/0126-bots-that-play-their-decks.md) (S66) rebuilt
this. Before it, every non-creature permanent was worth exactly what a
card in hand is worth (`Weights.Permanent` 1.20 against `Weights.Hand`
1.20), a small creature or a cheap spell was worth less, and the
heuristic never cast about a third of each curated deck: every mana
rock, every enchantment engine, the tutors, the loots and the
sacrifice outlets.

The heuristic still reads no oracle text ([ADR 0033
§3](decisions/0033-ai-bot-seat.md)). Every price below comes from a
field the seat's own view carries: `mana_abilities`, `ability_rows`,
`activated_abilities`, `additional_cost`, and the catalog-declared
`purpose` ([docs/protocol.md](protocol.md)).

### The bar a move has to clear

A cast is priced as what the card is worth once it resolves, less
`Weights.Hand` (1.20) for the card leaving the hand (`valueOfCast`,
`resolvedValue`). The bot takes the best move that scores above its bar:

| Where | Bar |
|---|---|
| Its own main phase, empty stack | `PassThreshold`, 0.25 |
| Anywhere else | `InstantThreshold`, 1.50: holding an instant is a real option |
| One of the two leftover windows (below), for a move that costs only mana and taps | `LeftoverThreshold`, 0.00 |
| Any window, for a move whose only non-mana cost is sacrificing permanents that are dying anyway | `LeftoverThreshold`, 0.00 |

### Mana sources

**On the battlefield** (`permanentValue`), a non-land mana source is
worth `ManaSource` (1.00) for its first mana and `ManaPerExtra` (1.00)
for each further mana its best repeatable ability makes. Repeatable
means a tap ability with no sacrifice or exile-itself cost. The amount
is the number of symbols in `produced`, net of the ability's own mana
cost: Sol Ring makes two and is worth 2.00, and a Signet's "{1}, {T}:
Add two" makes one. A land is priced as before. A mana creature is
priced as a creature.

**At cast time only**, a ramp premium (`rampPremium`): `RampPerMana`
(1.00) times the smaller of the mana the source makes and the bot's
mana deficit.

```
want    = the largest mana value among the other cards in hand and the
          commander (with its tax), capped at RampWantCap (7)
sources = the mana the bot's own repeatable sources make, tapped or not
deficit = max(0, want − sources)
```

The premium is large while the bot cannot cast what it holds, and it
falls to nothing as the bot catches up, with no turn counter:

| Card, situation | Price |
|---|---:|
| Sol Ring, turn 1, a five-drop in hand | 2.00 + 2 × 1.00 − 1.20 = **+2.80** |
| Arcane Signet, deficit 3 | 1.00 + 1.00 − 1.20 = **+0.80** |
| Arcane Signet, deficit 0 | 1.00 − 1.20 = **−0.20**: not cast |

Rituals, Lotus Petal, Treasures and the Altars are not priced as mana
sources. Their mana is one-shot, and what it is worth is the spell it
lets the bot cast this turn, which needs a plan the heuristic does not
make.

### Other permanents, by role

A permanent that is not a creature, a planeswalker, a land or a mana
source (an enchantment or artifact engine, an Equipment, a Vehicle, an
Altar) is worth:

```
max(Permanent 1.20, PermanentPerMana 0.50 × mana value) + rowUtility
```

`rowUtility` adds `RowTriggered` (0.60) for each triggered row in
`ability_rows`, `RowStatic` (0.50) for each static row and
`RowActivated` (0.40) for each activated row, counting at most `RowCap`
(3) rows. A creature gets the same `rowUtility` on top of its body,
before the tapped, summoning-sick and restriction multipliers.
Keywords and mana abilities are not rows, so a flier is not paid twice
and a mana elf is priced as a mana source.

Rhystic Study (mana value 3, one triggered row) is 1.50 + 0.60 − 1.20 =
+0.90. Viscera Seer (a 1/1 with one activated row) is
0.90 × (1.45 + 0.40) − 1.20 = +0.47.

The rows say that a card does something, not how much. A drawback row
(Sulfuric Vortex's damage to its own controller) counts as a plus. A
card the engine does not implement has no rows, and keeps the flat
`Permanent` and [ADR 0037](decisions/0037-unimplemented-card-signal.md)'s
penalty.

This is the same `permanentValue` the board evaluation uses, so an
opponent's Rhystic Study is now worth removing, and a Blood Artist is
worth more than a vanilla 1/1 when the bot chooses what to sacrifice.
`CombatValue`, which the combat planner compares attackers and blockers
with, is still the body alone.

### The two leftover windows

Mana empties between steps, and a tapped permanent untaps in its
controller's untap step. So in two windows, with the stack empty, a
move that spends only mana and taps costs nothing the bot would
otherwise keep, and it needs to clear only `LeftoverThreshold`
(`LeftoverWindows`, `windows.go`):

- **The bot's own second main phase**, for sorcery-speed moves that tap
  no creature. An instant, a flash spell or an instant-speed ability
  waits for the next window instead, keeping the mana up meanwhile.
- **The end step of the seat whose turn comes just before the bot's.**
  This is the end-of-turn Entomb, Vampiric Tutor or loot.

A land sacrifice counts as mana here when the same cast's declared
`purpose.lands` more than replaces it
([#2469](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2469),
`landSacrificeIsNetMana`). Harrow sacrifices a land and puts two onto the
battlefield untapped, so it ends with a land more than it began with and
spends nothing the bot would keep; it gets `LeftoverThreshold` like a
move that costs mana and taps. A land sacrifice with no `lands` purpose
behind it, one that does not replace what it sacrifices, and a sacrifice
of anything but the bot's own land keep the normal bar. It does not
change the price: once the ramp premium has closed, Harrow is priced
below passing and is still not cast.

Two prices go with the windows:

- `SpellFloor` (1.30): an untargeted instant or sorcery the engine
  implements is worth at least a card that replaces itself and does a
  little more. That is 0.10 above the card it costs, so it clears
  `LeftoverThreshold` and stays under `PassThreshold`: cheap spells fill
  the leftover windows and do not crowd out development. A targeted
  spell is priced by its targets instead.
- `TapByTiming`: tapping one of the bot's untapped creatures to pay a
  cost is free in the end step before its turn. In the bot's own first
  main phase it costs the blocker and the attack it gives up. Elsewhere
  it costs a flat 0.30 for the blocker.

### Purposes: what a spell or an ability does

The catalog declares a `purpose` on a card, a mode, an alternative cost
or an activated row (`effects.Spec.Purpose`, [ADR 0126
§6](decisions/0126-bots-that-play-their-decks.md)). It is declared by
hand, like `Completeness`. In S66, every card in the four curated decks
in the priced classes declares one, and so does every catalog board
wipe. A card with no purpose is priced as before: an instant or sorcery
at `SpellPerMana` (0.60) times its mana value, with `SpellFloor` under
it.

With `PricePurposes` on, a purpose replaces that proxy (`purposeValue`):

```
Hand × (draws + TutorWeight × tutors + SelfMillWeight × self_mill_tutor)
− DiscardWeight × discards
+ (ManaSource + the ramp premium) × lands
+ TokenWeight × tokens
+ Weights.Energy × energy
```

`TutorWeight` is 1.00. `SelfMillWeight` is 0.50, because Entomb finds a
card for a graveyard plan the policy cannot see. `DiscardWeight` is
0.60 and `TokenWeight` 0.50. `SpellFloor` stays under the result.

On a permanent, the purpose is its enters effect, added to its body:
Wood Elves' land, Mulldrifter's two cards. On the bot's own activated
row, it replaces `ActivateBase`, so a loot is priced by the card it
draws and a land sacrifice by the land it fetches. A row with no
purpose keeps the flat `ActivateBase` (0.50).

### Board wipes

A purpose's `sweep` names what it removes (`matches`), how (`destroy`,
`exile`, `bounce`, `damage`, `minus`, `sacrifice`) and, for damage and
−N/−N, how much. With `PriceSweeps` on, a wipe is worth the change it
makes to the bot's own score (`sweepValue`):

```
ScoreEval(with the swept permanents removed) − ScoreEval(now) − Hand
```

`ScoreEval` weighs the opposition by `OpponentMean` and `OpponentMax`,
so a wipe into the table's leader is worth a lot, and a wipe into the
bot's own winning board is worth less than nothing.

- A destroy sweep spares an indestructible permanent.
- Damage and −N/−N spare a creature whose toughness is above the amount.
- A bounce takes half a permanent's value, because it comes back, and
  all of a token's.
- A sweep marked `partial` takes half of each permanent it matches,
  because the view does not say which ones it spares.
- `opponents_only` spares the bot's own permanents.

A wipe gets no `SpellFloor`.

### Discards and discard payoffs

With `DiscardCostByCard` on, a card discarded to pay a spell's cost
costs what that card is worth to the bot (`cardValue`, the price the
cleanup discard already uses), not a flat 1.20. A spare land late in
the game costs about 0.30. The bot's last land in hand, while it has
fewer than `LandsWanted` (5) sources, costs `LastLandDiscard` (1.00)
more. The enumerator offers one payment per combination of cards, so
the bot pays with the cheapest. Unexpected Windfall discarding a spare
land is cast. Discarding the only land in hand early, it is not.

A triggered row can declare a `discard_payoff`. Mary Read and Anne
Bonny makes a Treasure for each Island, Pirate or Vehicle discarded,
and Marauding Mako takes a +1/+1 counter for any card. With
`PriceDiscardPayoffs` on, every discard of the bot's own card, whether
it pays a cost, answers a loot or happens at cleanup, costs `cardValue`
less what its payoffs pay:

```
TokenWeight × tokens
+ (Weights.Power + Weights.Toughness) × counters
+ DamageToOpponent × damage_each_opponent × live opponents
```

So with Mary Read out, the bot loots away the Island rather than the
Mountain.

### Sacrifices

With `SacrificeDyingAnyway` on, a permanent sacrificed to pay a cost
costs its value times the chance the bot would have kept it
(`sacrifice.go`):

- A target of an opponent's spell or ability on the stack is kept with
  a chance of 1 − `RemovalConfidence` (0.80).
- A permanent that a declared sweep on the stack would remove is not
  kept.
- After blockers are declared, a creature in a combat it loses without
  taking anything with it is not kept.

A move whose only non-mana cost is sacrificing such permanents clears
`LeftoverThreshold` in any window, so "sacrifice it in response" is the
bot's play.

Each `death_payoff` row on a permanent the bot controls (Blood Artist,
Zulaport Cutthroat, Bastion of Remembrance) takes `DeathPayoff` (0.60)
off the cost of every creature it sacrifices. A payoff's own row does
not count towards its own sacrifice. Sacrificing a commander or a big
creature for the payoffs is still not worth it.

The Altars' mana abilities sacrifice a creature, so the bot does not
activate them for floating mana. The auto-tapper uses them when a cast
needs the mana, as it always did.

### The old prices, kept runnable

Every term above is a `Config` or `Weights` field whose zero value is
the heuristic as it was before S66. `heuristic.BaselineConfig()` zeroes
all of them, and the arena's `heuristic-baseline` contestant plays it.
So two `heuristic` seats against two `heuristic-baseline` seats measure
exactly what S66 changed. `TestBaselineConfigRanksTheSuiteAsBefore`
holds the baseline to the rankings the policy gave every suite position
before S66. The measured effect is in ADR 0126's
[Measurements](decisions/0126-bots-that-play-their-decks.md#measurements).

## An attached permanent is priced once, by its role (#727)

An Equipment's +2/+2 arrives on the wire as its host's `power` and
`toughness` — the projection is post-layer — so a board evaluation that
also charges `w.Permanent` for the Equipment has paid for the same
+2/+2 twice. [ADR 0036](decisions/0036-attachments.md) named that
double count when attachments shipped and asked for "score it zero",
and zero is too broad: a Pacifism scored at zero is a removal spell the
bot can see no reason to cast.

So an attached permanent is classified ONCE, by **what it is doing**,
and priced by that role. The classification is
`heuristic.AttachmentRole`, and it reads the HOST rather than the card:
a policy may not hold a `*game.Game` ([ADR 0033
§3](decisions/0033-ai-bot-seat.md)) and so cannot open a `CardDef` and
inspect its `StaticAbility` list — but it does not need to, because
every one of those statics has already run and left its mark on the
host's projection. No card name appears anywhere in the split, which is
what makes it work for the next Aura nobody has written yet.

| Role | How it is recognised | What it is worth on its own line |
|---|---|---|
| **buff** — Equipment, `+N/+N` Aura | the default: attached to a permanent, no control change, no restriction | `AttachedEquipment` (0.60) for an Equipment or Fortification, which survives its host and can be moved; `AttachedAura` (0.10) for an Aura, which cannot. **The boost itself is on the host.** |
| **restriction** — Pacifism, Arrest, Faith's Fetters | the host carries a neutralising `restrictions` token (`cant_attack`, `cant_block`, `cant_activate`, `cant_activate_mana`) and is controlled by somebody else | **the host's neutralised value** — exactly what `CreatureValue`'s restriction discount took off the host's controller, credited back to the seat that cast the Aura |
| **control** — Control Magic, Mind Control | the host's `controller` is the attachment's controller while its `owner` is somebody else | **nothing.** Layer 2 has already moved the creature onto this seat's ledger; the ordinary creature pass counts it there |
| **curse** — Curse of Opulence | `attached_to.kind` is `player` | its own permanent, unchanged — there is no host permanent for the value to ride |

`cant_be_blocked` is deliberately not a neutralising restriction. It is
carried on the attacker but restricts the DEFENDER, so a Whispersilk
Cloak or an Aether Tunnel makes its host *better*, and discounting the
host for it would have the sign backwards. The combat planner already
reads it where it belongs — `couldBlock` in `combat.go` will not pair a
blocker against it — so the attack and block plans see it even though
the board evaluation charges nothing for it either way.

**The restriction discount is the other half, and it is what makes
removal Auras castable.** `CreatureValue` used to ignore
`CardView.Restrictions` entirely, so a pacified 5/5 was worth exactly
what it was worth the turn before and the only thing Pacifism changed
about the bot's score was the 1.20 the Aura cost it as a permanent —
casting removal made the bot's own evaluation go *down*. Each
restriction is now a multiplier (`CantAttack` 0.45, `CantBlock` 0.70,
`CantActivate` 0.80) and they compose, so Pacifism leaves about 31% of
a creature and Arrest about 25%. The debit lands on the host's
controller and the matching credit on the Aura's, so the ledger
balances: a removal Aura is worth exactly the creature it is holding
down, no more.

The two soft edges are deliberate and both are cheap. A buff Aura on a
creature this seat stole with something *else* reads as control and is
priced at 0 rather than 0.10. A second restriction Aura on an
already-pacified creature claims the same neutralised value as the
first — a board where two seats have spent two cards answering one
creature, and over-rating their answers is not a decision anyone is
worried about.

## The four curated decks

Curated, not generated, and **every card in them is covered by a
build-failing test**: a card whose oracle ID has no registered effect
spec fails the build, with the back-face keys of double-faced cards
explicitly rejected so a card does not count as covered because its
back face registered. Without that test "curated" would rot the first
time a spec was refactored.

| ID | Name | Commander | Archetype |
|---|---|---|---|
| `izzet-aggro` | Raid and Ransack | Mary Read and Anne Bonny | Aggro (UR) — Pirates and Islands: loot and rummage away the cards you don't need, and every Island, Pirate or Vehicle you discard becomes a Treasure |
| `simic-ramp` | Deep Roots | Tatyova, Benthic Druid | Ramp-stompy (UG) — landfall: extra land drops every turn, a card off each with Tatyova, and landfall payoffs that turn the lands into mana, tokens and threats |
| `esper-control` | The Long Answer | Y'shtola, Night's Blessed | Control (WUB) — big spells: every noncreature spell of mana value 3 or more drains each opponent, so the answers also win the game |
| `mono-black-aristocrats` | Body Count | Syr Konrad, the Grim | Aristocrats (B) — death triggers: sacrifice creatures for value, drain the table each time one dies, then reanimate the best of them |

The decks are "upgraded precons" ([#2436](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2436)):
fetchlands and dual lands, but few tutors and little fast mana (Sol
Ring everywhere, Ancient Tomb only in mono-black), and each one
teaches the single mechanic in its row above.
Every non-basic card is graded `full` in the catalog, apart from a
short list of minor caveats named in `internal/decks/full_test.go`
(Path of Ancestry, Delighted Halfling, Cultivate, Tireless Provisioner,
the two partner Pirates, Ragavan and the two Artillerists), and that
test fails the build when a card falls below it. A saved table setup
names a deck by ID, so it seats the current list.

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

> **Which tiers.** `assisted` and `strong` only. `random` and
> `heuristic` never improvise — neither has a model, and reading a
> card's oracle text is the whole job — and a model tier with no
> endpoint configured does not either. Turn it off everywhere with
> `CMDCTRL_BOT_IMPROVISE=0`; a bot then casts an uncatalogued card,
> the card does nothing, and it is yours to apply by hand, which is
> what every tier did before
> [#686](https://github.com/krakenhavoc/cmd_and_ctrl/issues/686).

The catalog is several thousand cards, and a bot's deck is drawn entirely
from it — but a card can be registered without every clause of it
being implemented. When a bot casts a card the engine cannot run,
**the bot applies the text by hand**, with the same four sandbox verbs
you have: `move_card`, `change_life`, `add_counter`, `mark_damage`.

### When it happens, exactly

**After the spell resolves, never instead of casting it.** The bot
casts the card through the ordinary move list, the engine charges the
mana, the spell resolves and does nothing, and *then* the bot applies
the text. That order is the point: the four sandbox verbs cannot tap a
land, so a bot that moved the card out of its own hand and applied the
effect would have cast it for free.

It fires **once per card**, at the moment that card leaves the stack,
and only when all of these hold:

- the bot cast it itself,
- it was a *spell* on the stack, not an ability,
- the card is one the engine will not run (the same `manual` mark the
  deck-upload summary and the stack overlay show you),
- the bot's own decklist has the card's oracle text, and
- the bot does not owe a prompt — a bot answering "pay {2} or
  sacrifice" answers it, it does not wander off.

Nothing else triggers it. In particular an **unimplemented permanent's
triggered or activated abilities are never improvised**: there is no
way to tell from the board that a trigger should have fired, and the
condition would recur every turn. A bot also never improvises somebody
*else's* card.

**It is capped.** At most **8 improvisation model calls per bot seat
per game**, and at most one per card. A bot that hits the cap says so
in the server log and leaves the rest of its uncatalogued cards alone.
Between the two limits, the worst case is a known number rather than
"however many cards it draws".

**It can decline, and often will.** The model is told that answering
with nothing is a correct answer, and it is the right one whenever the
card cannot be done faithfully with four verbs. "Search your library"
is the common case: a bot cannot see any library's contents, so it
cannot pick a card out of one, and the honest answer is to leave that
clause out and say so.

**If the attempt is refused, the table is told that too.** A bundle
the server will not apply — a verb outside the four, a card that does
not exist, anything that fails validation — is dropped whole, and a
chat line names the card and says nothing was changed. That line is
the useful half: the card is sitting in a graveyard having done
nothing, and now you know to do it by hand yourself.

### What makes it safe

Three things, and all three are enforced in code rather than left to
the policy — which matters more here than anywhere else in the bot,
because the thing writing the bundle is a language model:

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

### What it cannot see, and the one case to watch

The model writing the bundle is handed **the seat's own filtered view
and nothing else** — the same bytes a human in that seat receives —
plus the card's oracle text out of the bot's own decklist. It cannot
see your hand, it cannot see any library, and it has no handle that
could reach either.

The case worth knowing about: **a countered spell also leaves the
stack**, and from a filtered view that looks much like one that
resolved. The model is shown the board and told to answer with nothing
if the spell did not really resolve, which catches the obvious cases —
but if a bot ever improvises a card you countered, that is the
mechanism, and the recourse is the one every improvisation has: read
the line, undo it, free.

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
its raw reply, the index and label parsed out of it and the move they
resolved to (`parsed_index`, `parsed_move`, `model_index`, `pick`), the runner's own fallback
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

A decision log also carries **one extra line at the end of each
game**: the spend record, `{"kind":"spend", ...}`. Every other line is
a decision window and has no `kind` at all. `decisionlog.Scan` hands
out the windows and skips the rest, so a reader that was counting
decisions counts the same number it always did; `decisionlog.ScanAll`
is the way in when you want the spend line.

---

## Per-game model spend (#735)

**Where to read it:** the server logs one line per bot game, at INFO,
when the table's last bot seat exits.

```
bot model spend for the game game=6a1f… seats=4 tiers="assisted x2, heuristic x2"
  calls=228 decision_calls=226 improv_calls=2
  input_tokens=87743 output_tokens=2736 cache_read_tokens=0 cache_write_tokens=0
  cached_prompt_tokens=0 model_time=41.2s
  per_seat="assisted 113c/1i 43811in/1368out; assisted 113c/1i 43932in/1368out; heuristic 0c/0i 0in/0out; heuristic 0c/0i 0in/0out"
```

It is emitted for **every** bot game, including the ones that spent
nothing. "This table cost nothing" is a measurement, and a line that
appeared only when there was a bill could not be told from a line that
failed to be written.

Three surfaces, same numbers:

| Surface | What it is for |
|---|---|
| The log line above | The operator's answer to "what did last night cost". One line per game, no configuration. |
| `aiseat.Runner.Stats().Spend` | One seat, live, while it plays. It is in `summary.json` for every `boteval arena` run, and in the arena's per-policy totals. |
| The decision log's `kind:"spend"` record | The machine-readable per-game copy: every seat, split by purpose, written just before the file closes. Needs `CMDCTRL_BOT_DECISION_LOG`. |

**Deciding and improvising are counted apart and never averaged
together.** They are two different calls with two different caps: a
decision asks for one integer against a prompt-cached prefix, and
[improvisation](#improvisation-and-why-an-undo-is-free) asks for a
whole bundle written from oracle text, with `MaxTokens` an order of
magnitude larger and a hard cap of eight calls per seat per game. A
single tokens-per-call number over the two would describe neither.
`Spend.Total()` adds them when what you want is the bill.

**A call that failed is still spend.** Timeouts, malformed replies and
out-of-range answers are all counted, because they were all billed.
That is the number that makes a too-slow self-hosted model visible:
read it next to `ModelTimeouts` in the funnel stats.

**What the tokens are.** `input_tokens` / `output_tokens` are the
provider's own usage fields as the transport reported them, not an
estimate — a provider that reports none leaves them at zero, which is
a measurement that could not be taken rather than a call that was
free. `cache_read_tokens` and `cache_write_tokens` are Anthropic's
explicit prompt-cache breakpoints; `cached_prompt_tokens` is a local
server's own prefix-cache hit count (Ollama's
`prompt_tokens_details.cached_tokens`), which is an optimisation
nobody is charged for and is deliberately not added to the other two.

### S31 exit criterion 4

> "Per-game model spend is measured and recorded, not estimated."

**The measurement path is built and proven end to end against the fake
client; the NUMBER is pending a keyed run.** Every deployment so far
has run a local model (Ollama), which reports tokens but has no bill,
and there is no API key in CI — so the figure that belongs in [ADR
0033](decisions/0033-ai-bot-seat.md) §5 in place of its
order-of-magnitude estimate cannot be taken here. What is settled is
that there is now one place to read it from, for any game, with no
extra flags.

To take it, point a seat at a keyed endpoint and play:

```bash
CMDCTRL_ANTHROPIC_API_KEY=sk-… \
  ./server/bin/boteval arena --seats assisted,heuristic,heuristic,heuristic \
    --decks izzet-aggro,simic-ramp,esper-control,mono-black-aristocrats \
    --games 4 --seed 1 --rotate --out ./arena-out
```

and read `spend` off each seat in `summary.json`, or run a server with
`CMDCTRL_BOT_DECISION_LOG` set and read the `kind:"spend"` line of
each game file. Record the model ids with the numbers: a spend figure
without the model that produced it is not a measurement of anything.
Four-bot and one-human-plus-three-bots tables are both wanted — a
human at the table changes how many windows a bot sees.

The reference point to check the result against is ADR 0033 §5's
estimate of "cents per game", and the two facts that are already
measured: Layer A absorbs ~90% of windows, and ~60–70% of the
survivors escalate to the frontier model rather than the ~20% the ADR
guessed.

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
starts one ordinary `aiseat.Runner` per chair — one goroutine each,
unless `--lockstep` asks for one goroutine in all (below). Every seat
sees exactly the filtered `aiseat.Input` it would see at a real table.

| Flag | What it does |
|---|---|
| `--seats` | one contestant per chair, comma-separated. 2–4 chairs. A contestant is a tier, `heuristic-baseline` or `heuristic-noexert`. `heuristic-baseline` is the heuristic frozen as it priced cards before S66 ([ADR 0126](decisions/0126-bots-that-play-their-decks.md) §1, `heuristic.BaselineConfig`). `heuristic-noexert` is today's heuristic with [ADR 0130](decisions/0130-exert.md) §9's exert pricing off, to measure that pricing alone. Arena names only; the lobby and `GET /bot/options` never offer them. |
| `--decks` | one curated deck id per chair, or none at all — a partial list is refused. No `--decks` deals a synthetic 65-card red deck that needs no Scryfall dump, and `exert-battle` (also synthetic) is that deck in red and white with fifteen exert cards, for [ADR 0130](decisions/0130-exert.md) §9's measurement, and `monolith-battle` is that deck with six Basalt Monoliths and four Grim Monoliths, for #2500's. |
| `--names` | one tally name per chair. Use it when every chair is the same tier and the thing being compared is the deck or the configuration. |
| `--games`, `--seed` | game *i* uses `seed+i`, so two policies can be compared on the same deals. By default the seats run one goroutine each, so the seed fixes the deal and the policies' randomness, not the interleaving — a rerun is the same deals, not always the same games (#1409). Add `--lockstep` for the same games. |
| `--lockstep` | plays each game on one goroutine, seat by seat, so the same `--seed` replays the same games move for move (#1503). Off by default — see "Lockstep runs" below for what it changes. |
| `--rotate` | moves each contestant one chair along per game (contestant *k* sits at position `(k+i) mod n`). **Use it.** Turn order in Commander is worth real percentage points; without rotation you are measuring the chair. |
| `--turn-budget`, `--wall`, `--stall` | when to stop a game that will not end (default 60 turns, 30 minutes, `3×max-think+15s` with no committed move). |
| `--max-think`, `--model`, `--frontier-model`, `--endpoint` | the model tiers' deadline and transport. With an endpoint set and no `--max-think`, the deadline defaults to **20s**, the same local default `cmd/server` applies and for the same reason (**120s** with `--think`). |
| `--think`, `--max-tokens` | [let the model think](#letting-the-model-think-2196), and set the decision calls' reply budget (8000 with `--think` when unset). Both are recorded on the report's model line. |
| `--out` | artifacts directory. Each run gets its own `<out>/<RFC3339 start>/`. |
| `--decision-log`, `--decision-log-mode` | per-game decision logs under `<out>/decisions`. **Operator-only** — see the section above. |
| `--replays` | per-game replay JSONL under `<out>/replays`. Off by default: a four-seat replay is ~320 MiB — and this flag hands the directory to `ws.Room`, which also writes `games/<id>.json` (the full authoritative state, rewritten on **every committed move**) and `restore/<id>.json` while a game is live. Budget for all three. |
| `--block-grace` | how long an attacking bot holds its pass in declare-blockers while a defender is still declaring blockers with a legal block to make. Defaults to production's value. `0` turns it off — see below. |
| `--md`, `--json` | what to print. Markdown by default. |

Env fallbacks match the server's: `CMDCTRL_OPENAI_ENDPOINT`,
`CMDCTRL_OPENAI_API_KEY`, `CMDCTRL_BOT_MODEL`,
`CMDCTRL_BOT_FRONTIER_MODEL`, `CMDCTRL_BOT_MAX_THINK`,
`CMDCTRL_BOT_THINK`, `CMDCTRL_BOT_MAX_TOKENS`, `CMDCTRL_SCRYFALL_DUMP`.

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

**`--block-grace` used to be a fidelity knob; since #1279 it is a
speed knob.** Before #1279 an attacking bot that passed at once could
let the step end before a slower seat had declared a block, so a run
with it off understated every combat difference. The engine now
completes each defender's declaration explicitly — their own pass, a
`finish_blocks`, or having no legal block — and hands the attacker
priority back after the last one ([ADR 0045](decisions/0045-combat-restrictions.md)
Decision 38), so no block is lost either way; the grace only saves the
extra round of passes. It still defaults to production's value, so a
run matches what a live table does. Since #1501 it almost never has
anything to hold for: nobody holds priority while a defender is
declaring (CR 509.1), so an attacking bot has no pass to hold until the
declarations are in. A declaring bot defender finishes with the
`finish_blocks` move, which the enumerator offers as always-legal, so
a defender whose policy wants no block still ends its declaration
rather than sleeping on it.

#### Lockstep runs (#1503)

By default an arena seat is a runner goroutine, exactly as at a live
table, and after every commit the Go scheduler decides which seat acts
first. That is where a seeded game forks: a defender's block lands
before or after the attacker's pass depending on which goroutine woke
first. `--lockstep` takes the goroutines away. The arena steps each
seat in chair order, one wake's act-loop at a time
(`aiseat.NewStepped` / `Runner.Step`), round after round — the same
enumeration, policy, observer, forced answers and dispatch a woken
runner runs; only *when* a seat acts is the arena's. The same seed then
replays the same game, move for move, which is what you want when a
run turns up a strange game and you need to watch it again with a
decision log on.

What else it changes, all of it following from "nobody else can act
while a seat holds":

- **No pacing and no block grace.** `--block-grace` is ignored: the
  defenders it would wait for cannot act until the attacker's step
  returns, so it would cost its full length on every combat and change
  nothing.
- **Stalls are exact, not timed.** A full round in which no seat
  committed anything is a table nothing will move again, so the game is
  marked stalled then and there; `--stall` is ignored. The dump says
  `no move in a full lockstep round`.
- **Model seats are only as reproducible as their endpoint.** Lockstep
  removes the scheduler, not the model's own sampling, and a decision
  that overruns `--max-think` falls back — the one place wall clock
  still reaches a lockstep game.
- **It is not how a live table runs.** Use it to reproduce and to
  compare two policies move for move; keep the default for a run whose
  numbers should reflect production's scheduling, races included. The
  report's games line says `schedule lockstep` or `schedule concurrent`,
  and `summary.json` carries `lockstep`, so two reports can be told
  apart on the one point that decides whether their seeds are
  comparable.

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
- **Play by contestant** — the same columns again, one row per
  policy *and deck* ([ADR 0126](decisions/0126-bots-that-play-their-decks.md)
  §1). Four `heuristic` seats on the four curated decks are one Play
  row, which wins 25% by construction, and four rows here, which is
  where the deck balance shows. Name the chairs with `--names` or not;
  the deck column says which is which either way.
- **Cards** — for each contestant, every non-land card of its own it
  was offered: a `cast` or `activate` move naming it. Per card: the
  windows it was offered in, the windows it was taken in, and the
  seat-games in which it was offered and in which it was used at least
  once. A window counts a card once however many ways it is offered
  (two targets, two abilities). A card offered in five or more windows
  and never taken is `never`, and the summary table lists each
  contestant's `never` cards. Below it, ADR 0126's acceptance-bar cards:
  every mana rock and dork (A2: a repeatable tap mana ability, so not
  Lotus Petal or an Altar), and the six A3 canaries (Sol Ring, Rhystic
  Study, Mary Read and Anne Bonny's loot, Entomb, Harrow, Viscera Seer),
  each with the share of games it was used in and the bar it is held
  to. A canary nobody was dealt says `not offered`. The full per-card
  tables are folded under `<details>`. The tally is the runner's
  observer, so it needs no decision log; `summary.json` carries it as
  `cards` and `canaries`, and each game's per-seat counts are in
  `games.jsonl`.
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
./bin/boteval suite run --policy assisted --think --max-tokens 8000 --max-think 120s --md
```

`--think` and `--max-tokens` are the
[thinking experiment](#letting-the-model-think-2196); the report's
header records them (and `--note`), and a thinking run adds the median
reply length and how much the model thought.

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

## Enumerating an optional additional cost (#664)

A card with kicker, multikicker or buyback is not one cast, it is
several: an unkicked Burst Lightning and a kicked one are different
moves at different prices with different effects, and a bot only ever
offered the cheap one could never kick anything. Each announced set
walks the whole modes x targets x payments expansion on its own, so
the policy for choosing the sets is what keeps that product finite
([ADR 0073](decisions/0073-optional-additional-costs-and-the-cast-gate.md)
§9):

- **Decline everything, or pay exactly ONE of the offered costs — or,
  on a card with two kicker costs, both.** "Kicker {R} and/or {W}"
  (Thornscape Battlemage, Archangel of Wrath, CR 702.33b) prints lines
  that need both paid, so the pair is one extra announcement (#2153).
  Any other pair of different optional costs (a kicker beside a
  buyback) is a power-set search whose every member needs its own
  affordability probe, and a bot does not take that line; it is never
  offered one it cannot pay for.
- **A repeatable cost is offered up to THREE times.** Multikicker is
  unbounded in paper, but an announcement has to be finite and a
  decision loop has to terminate. Three is a policy number, not a
  rule, and it lives in `legal.maxEnumeratedRepeats`.
- **Priced through the engine's own helper.** `game.AdditionalCostMana`
  (which replaced `AddOptionalCostMana` in ADR 0100 sub-PR 3) adds the
  claimed costs' mana at CR 601.2f, and both the cast path and the
  enumerator reach it through `PriceCast` — so a kicked line is never
  advertised at the unkicked price, which is the same #544 discipline
  the cost modifiers follow.
- **An either/or additional cost is one move per payable branch**
  ([ADR 0100](decisions/0100-delve-either-or-and-variable-sacrifice-costs.md)
  §6). "Sacrifice an artifact or discard a card" is at most three
  announcements, each carrying its `cost_branch` and priced with it, so
  Lightning Axe's "pay {5}" move costs `{5}{R}` and its discard move
  `{R}`. A branch the engine's `AdditionalCostBranchPayableLocked` says
  the seat cannot pay is not offered, and a card with no payable branch
  is not offered at all (CR 601.2h). Each branch's cards are paid by the
  ordinary payment search — the same discard, sacrifice and blight
  payments a mandatory cost gets. One policy on top of the rule: a
  "pay N life" branch is not offered to a seat at N life or less, the
  line the alternative cost's life already refuses (CR 119.4 would allow
  it, and the next state-based check would kill the bot). The life a
  branch pays rides `MoveCost.Life`, and the move label names the branch
  ("Cast Lightning Axe (Pay {5})").
- **A variable sacrifice on a cast is a short ladder of counts**
  ([ADR 0100](decisions/0100-delve-either-or-and-variable-sacrifice-costs.md)
  §6). "Sacrifice any number of creatures" (Vicious Betrayal, Torgaar)
  is offered at zero and then at up to THREE positive counts, smallest
  first, each a prefix of the seat's fuel order
  (`legal.castVariableSacrificePayments`, `maxEnumeratedVariableCounts`),
  so the counts nest. Each count is priced on its own through the one
  pricer with `CostQuery.Sacrificing` set, so Torgaar paid with three
  creatures is offered at `{B}{B}` and a count the seat cannot afford is
  not offered at all. "Sacrifice X lands" (Devastating Summons) is the
  same ladder with the count as the move's `x_value`, starting at one
  when the card's X is the whole of its effect (#810). "Sacrifice X
  creatures. Destroy X target creatures" (Eliminate the Competition)
  lets the TARGETS fix X, and X fixes the one payment: the first X of
  the fuel order. The move label names what is sacrificed
  ("... (sacrificing Bear, Bear)").
- **The move label names the kick** ("Cast Burst Lightning (Kicker
  {4})", "... (Multikicker {G} x3)"), so the kicked and unkicked casts
  of one card are distinguishable in the move list and in a bot-eval
  trace.
- **A gift is one move per opponent** (#1267, [ADR
  0089](decisions/0089-gift.md)). Promising a gift names the opponent
  who receives it, so every set that promises one is expanded once per
  entry of `game.GiftOpponentsLocked` — the list the engine validates
  against — and each is targeted under the PROMISED clause (Long River's
  Pull counters any spell once promised). The unpromised cast is always
  offered beside them. The label names the recipient ("Cast Dawn's
  Truce (Gift a card) → Seat2"). Which opponent to feed is the policy's
  choice; the enumerator only makes sure it has one.

A NON-MANA optional cost (Constant Mists' "Buyback—Sacrifice a land")
is expanded through the same payment search the mandatory sacrifice
cost uses. A cast owing TWO card-shaped sacrifice clauses at once — a
mandatory one and a kicker's — is not enumerated, for the reason
escape's exile-from-graveyard cost is not: the two pools have to be
searched together into one flat list, and no card asks for it.

## Enumerating a cast from a zone that is not the hand (#673)

A bot casts from every zone the engine would accept a cast from —
hand, the command zone, its own graveyard, exile and the top of its
own library — at every price the engine would charge. Two questions,
and the enumerator asks the engine both rather than answering either
itself:

- **May this card be cast from here?** The card's own declaration
  (`game.CastableZonesFor` — flashback, escape, Gravecrawler) or a
  granted `game.CastPermission` ([ADR 0066](decisions/0066-granted-cast-and-play-permissions.md)
  — Snapcaster, cascade, impulse exile, warp's recast, a foretold,
  suspended or madness-discarded card). The graveyard is walked on
  every decision, because
  a card's own flashback opens it with nothing granted; exile and the
  library are walked only when something has granted a permission,
  because nothing else can open them.
- **At what price?** `game.CastOffersForLocked` is the one list: a nil
  entry for the printed mana cost when the cast path allows a claim of
  nothing, plus every alternative cost claimable from that zone, in
  announce precedence and already filtered through
  `AlternativeCostPayableLocked`, the predicate the client's picker is
  filtered with too. One move per entry, so a flashed-back Faithless
  Looting and a
  hard-cast one are separate moves at separate prices — the same rule
  the optional costs follow.

A third question the same walk answers, and the reason an adventure
needed no code here: WHICH FACE. `Card.CastableFaces` offers both
halves of an adventure (CR 715.3) and of a modal DFC, and a grant that
names faces NARROWS the choice to exactly those (CR 715.4's "cast the
creature from exile") — `faceForCastLocked`'s own rule, so a half the
enumerator offers is a half the announce path accepts.

A zone the card PRICES must be paid for. A Faithless Looting in the
graveyard is never offered at the {R} in its corner, because
`validateCastPathLocked` would refuse that cast and the enumerator
must not offer a move the engine refuses.

The move label names the price and, when the price eats cards, what it
ate: `Cast Voracious Typhon from graveyard (Escape—{5}{G}{G}, Exile
four other cards from your graveyard, exiling Fuel 0, Fuel 1, …)`. The
life half of a pitch rides `Move.Cost.Life`, so a policy reading only
the wire payload does not price Force of Will as free.

### The card component of an alternative cost

Escape's "exile N other cards from your graveyard", Force of Will's
pitched blue card and Daze's returned Island are all one mechanism: a
COST paid in cards, named on the wire as `alt_cost_ids`. The
candidates come from `game.AltCostCandidatesLocked`, which shares its
per-card predicate with the announce validator, so a payment the
enumerator builds is a payment `CastSpell` accepts.

**Up to THREE payments are enumerated per offer** —
`legal.maxEnumeratedCostPayments`, beside `maxEnumeratedRepeats` and
for the same reason. It was ONE until #1013, and the cap was never the
real constraint: the missing EVALUATION was. [ADR
0033](decisions/0033-ai-bot-seat.md) §1's corollary is that a variable
in a cost must not become an arity of the target cross product —
escape-five over a twenty-card graveyard is 15,504 payments, each
needing its own price — and the payments were indistinguishable to a
policy that priced the battlefield and the seats and valued a card in a
graveyard at nothing. It picked between them by index, so an Uro
escaping over a graveyard holding a second Uro, a Snapcaster target and
three lands ate whichever three were oldest.

Two things changed, and the corollary still holds.

**The payments are priced.** `aiseat.CostFuelPricer` is
`TargetOrderer`'s twin: an optional Policy extension the runner asks
for once per decision, threaded into the enumerator as
`legal.Options.OrderCostFuel`. It is a second hook rather than a second
meaning for the first because the two ask OPPOSITE questions — a
target order ranks the board by importance and the enumerator keeps the
TOP of it; a fuel price ranks a seat's own cards by what it would lose
and the enumerator spends the BOTTOM. A policy that answered one with
the other would pitch its best card every time. A policy with no
opinion gets `AltCostCandidatesLocked`'s zone order, byte-identical to
what it got before.

The heuristic's answer is `heuristic.fuelValue`
(`aiseat/heuristic/fuel.go`), in three cases:

| where the card is | what it is worth |
| --- | --- |
| in hand, or the command zone | `Weights.Hand` (the number `Evaluate` already charges per card in hand) plus what it would do if it resolved |
| a graveyard or exile, and the seat can still CAST it (escape, flashback, a granted impulse — read off the view's own `castable_here` / offer stamps) | `Config.FuelIdle` plus `Config.FuelRecast` × the same card-in-hand price: a real card, at a discount, because it needs its own cost and its own window |
| a graveyard or exile, and it is idle | `Config.FuelFloor` for a LAND, `Config.FuelIdle` for anything else. That gap is the whole of "eat the lands, not the spells" |
| on the battlefield (Daze's Island) | the same `boardValue` the rest of the evaluation uses |

Neither floor is zero: a card nobody can use today is still one
tomorrow's delve or flashback might want, and zero would make every
unreadable card the first thing a cost ate.

**Delve** (CR 702.66, [ADR 0100](decisions/0100-delve-either-or-and-variable-sacrifice-costs.md)
owner decision 3) is priced out of the same fuel. A delve cast is
offered TWO payments, both drawn from the graveyard in the seat's
`OrderCostFuel` order: the FEWEST cards that make the cast affordable
at the largest X the seat can reach (the ordinary move, emitted per
target set — the empty payment when the pool already pays), and the
FULL budget, `min(CastPrice.DelveBudget, graveyard size)` cards,
offered once against the first announcement out of the leftover
budget, like an alternative-cost payment. Murktide Regent and
Soulflayer want the second. `valueOfCast` charges every delved card its
fuel price, exactly as it charges escape's exiles, so the policy
compares the two on what each spends. A seat that controls Teval,
Arbiter of Virtue ("Spells you cast have delve", ADR 0100 sub-PR 2)
gets the same two payments on every cast: the enumerator asks the one
`DelveForLocked`, which reads the grant off the battlefield.

**And the extra payments spend no target budget.** They are offered in
a second pass, out of whatever `MaxExpansionPerSource` the target walk
did not use, against the FIRST announcement it made — the same spell
with the same targets at a different price, which is what they all
are. An Uro with no targets has eleven unspent and gets its
alternatives; a removal spell with an escape cost over a wide board
spends its budget on targets and gets none. The corollary is enforced
rather than restated.

ONE function answers both halves, and that is deliberate: the
enumerator reads it to decide which payment to OFFER, and
`valueOfCast` reads it to price the payment it was offered. A second
scorer for the ordering would eventually disagree with the first about
what an escape costs, and the bot would take a payment it then priced
as a mistake. It also closes #673's declared gap — a non-hand cast
still costs no card in HAND, and what it really spends is now the
alternative cost's own price rather than nothing.

### A cost whose COUNT the activator announces (#1213 / #1421)

"Sacrifice one or more artifacts" (Radiant Lotus), "Sacrifice X
Treasures" (Grim Hireling), and "Tap X untapped Foods" (Apothecary
White) are the third variable in a cost, after the card-shaped payment
above and multikicker's repeat. Unlike either, what varies is a NUMBER
the activator names at announce, and every value of it is the same
ability with the same target at a different size.

**Up to THREE counts are enumerated per ability** —
`legal.maxEnumeratedVariableCounts`, beside `maxEnumeratedCostPayments`
and `maxEnumeratedRepeats` and for the same reason: [ADR
0033](decisions/0033-ai-bot-seat.md) §1's corollary that a variable in
a cost must not become an arity of the target/mode cross product. An
open count over a ten-artifact board is ten counts, and a budget spent
ten ways there would never reach a second target.

The counts offered are the SMALLEST ones — the clause's floor and the
two above it. Small is the conservative direction for a cost: it spends
the least board, and a policy that wants more can take the biggest of
the three. Inside each payment the permanents are already ordered
cheapest-to-keep first, by the same `Options.OrderCostFuel` hook the
card-shaped payments use, then by
`game.SacrificePaymentOrderForEffect`'s policy-neutral tie-break
(tokens, then lower mana value, then the ability's own source last). So
the counts NEST: a bot asked to sacrifice three eats the same two it
would have eaten to sacrifice two.

For a `CountFromX` sacrifice or tap-others clause the announced X IS
the count, so the move's `x_value` is the size of the payment it
carries rather than a value solved from the mana cost. The tap-others
enumerator uses the same three-count ceiling and nested, cheapest-first
payments. `effects.Register` refuses a cost that also puts `{X}` in its
mana component, or that has two CountFromX components, because one
announced number cannot pay two independently chosen counts. It also
refuses CountFromX on a mana ability, which has no X announcement.

The RETURN-to-hand component (`return_ids`) has no count to vary: every
printed clause returns exactly one permanent, and the enumerator offers
one move per candidate, cheapest-to-keep first, out of the ordinary
`MaxExpansionPerSource` budget.

### A waterbend payment (#1310 / #1311)

"Waterbend {8}: Transform Aang" and "Ward—Waterbend {4}" let the payer
tap untapped artifacts and creatures, each paying {1} of the generic
(CR 701.67a). A seat with no mana but a board of creatures can therefore
pay, and an enumerator that asked only "can the pool pay" would never
offer the move. `legal/waterbend.go` offers **one** payment, as crew
does: the permanents in the order the seat misses least — free ones
first (an artifact that is neither a creature nor a mana source, taken
whenever the budget allows), then creatures that make no mana (a
blocker spent each), then mana sources (tapping one for waterbend pays
what tapping it for mana would) — and the smallest prefix of that order
that makes the rest affordable. For "Waterbend {X}" it offers the
largest X the taps and the mana reach together, never below the printed
floor (Katara's "X can't be 0"). The heuristic prices each tapped
creature as a spent blocker, exactly as it prices a crew tap.

## Ordering target expansion by threat (#687)

`legal.Options.MaxExpansionPerSource` is spent in candidate order, so
before this the answer to "which twelve of an opponent's twenty
permanents may the bot point a removal spell at" was whatever order
`LegalTargetsForEffect` happened to walk the battlefield in — and the
table leader's best creature could simply be absent from the move
list, at which point no policy could pick it. [ADR
0033](decisions/0033-ai-bot-seat.md) §1 promised an ordering from the
start and it was not built until now.

**It is a hook, not a scorer in `legal`.** Ranking a board is a policy
question and `legal` may not import `aiseat` (which imports it), so
the enumerator takes `Options.OrderTargets` and the seat that wants an
order supplies one: `aiseat.TargetOrderer`, an optional Policy
extension the runner asks for once per decision. A policy with no
opinion pays nothing and gets exactly the enumeration it got before.

The heuristic implements it with the SAME functions it scores
everything else with — `Weights.Threat` for a seat (the one the attack
rotation ranks opponents with) and `Weights.boardValue` for a
permanent (the one `Evaluate` adds up, so #727's attachment roles and
every restriction discount apply). There is no second scorer to drift
from the first.

Two things it deliberately does not do:

- **It does not know what the spell is**, so it cannot prefer an
  opponent's creature for a Murder and your own for a Giant Growth.
  The ordering is by IMPORTANCE — the biggest objects on the table
  survive the cap, whoever controls them — and choosing among the
  survivors stays `Decide`'s job, where the spell is known. Its only
  power is to stop a target being dropped, and dropping the board's
  biggest permanent is wrong for every spell.
- **It prices nothing off the battlefield.** A spell on the stack and
  a card in a graveyard score zero, which keeps them in the engine's
  own candidate order. Those clauses have a handful of candidates and
  the cap does not bite on them. (A card in a graveyard is priced by
  the FUEL hook above, and deliberately not by this one: what a card
  is worth to SPEND and what it is worth to TARGET are different
  questions, which is why #1013 added a second hook rather than
  widening this one.)

The sort is stable, so equal scores keep the engine's order and two
enumerations of one board produce the same move list — which is what
makes a decision log replayable.

### A wrapper forwards what it wraps (#1060)

Both hooks are **optional Policy extensions**, found by type assertion
on the policy the runner holds — and an assertion that answers false
is indistinguishable from a policy with no opinion. No error, no log
line, no failing test: the feature simply does not happen.

That is exactly what happened. Every shipped tier is the heuristic
inside a wrapper (`rules.Filter` for `heuristic`, `model.Policy` for
`assisted` and `strong`), neither wrapper forwarded either hook, and
both orderings were dead on every seat the lobby could create — for a
month, while the pin tests went on passing against a bare
`heuristic.New()` that no seat is ever given.

The fix is one mechanism rather than two forwarding methods. A wrapper
declares **what it wraps**, once:

```go
func (f *Filter) Unwrap() aiseat.Policy { return f.Inner }
```

and every lookup goes through `aiseat.Capability[T]`, which walks that
chain outward-in and takes the **outermost** implementer. Outward-in
is the whole of the semantics, and it is right in both directions: a
wrapper that implements an extension itself means to override it
(`rules.Filter` is a `Tracer`, and Layer A's own verdict is the one
that must be reported), and one that does not means to be transparent.
An optional interface added next year is forwarded by every wrapper
without anybody editing one.

`tiers/tiers.go` holds a compile-time assertion that each layer it
assembles is an `aiseat.Unwrapper`, so adding a wrapper to a tier
without teaching it the chain is a build failure rather than a feature
that quietly stops happening. The runtime half is
`TestEveryShippedTierForwardsItsOptionalHooks`, which builds every
tier through the factory and asks it everything the runner and the
enumerator will ask it.

## Proliferating (#2525)

A proliferate (CR 701.34a) asks the proliferating seat which permanents
and players with counters get another counter of each kind. The
enumerator offers the empty answer (always legal), the engine's
suggested set as ONE whole move, first, and then the subsets its walk
reaches. The suggested set is the beneficial pick: every permanent or
player of the seat's own that a counter helps, and every opponent's one
it hurts (a -1/-1 or stun counter, poison, rad), where a single unwanted
kind rules a permanent out. It is offered whole on purpose: on a board
wider than the expansion cap the walk never builds the large set, and a
bot that could only name small ones would strand the rest.

The heuristic scores an answer by its overlap with the suggestion (+1 a
member, -1.5 anything else), so the exact suggestion wins, a strict
subset ranks below it, and with nothing suggested it chooses nothing.
The same table of harmful counters serves the client's pre-selection and
the bot, so they cannot drift. The model tiers get the same move list and
the same labels (`proliferate <name> …`, seats by name).

## Choosing a Ring-bearer (ADR 0114 §7)

When the Ring tempts a bot seat that controls two or more creatures, it
owes a `ring_bearer` prompt (one creature is chosen by the server). The
enumerator offers one move per creature, labelled `choose your
Ring-bearer: choose Ring-bearer <name>`, and the prompt always has an
answer: it is only asked with candidates, and the submit re-check
refuses only a creature that has gone, leaving the rest open.

The heuristic gives the Ring to the creature that attacks hardest: the
highest power, then one that can attack this turn (untapped, not
summoning sick), then the current Ring-bearer, which costs nothing to
keep. A creature the Ring's "is legendary" would put into the legend
rule against a same-named legendary permanent the seat already controls
goes last. The same price orders the pool before the enumerator's cap
(`legal.Options.OrderRingBearer`, the `aiseat.RingBearerOrderer` hook,
forwarded through every wrapper like the two above), so on a wide board
the best candidates are the ones offered.

The model tiers see the Ring in the board text: one `emblem:` line per
emblem, the Ring with its count ("The Ring (tempted 3 times): …"), and
a `Ring-bearer` tag on the creature. Legendary and the evasion reach the
evaluator through the effective type line and the block refusals, with
nothing new. The Ring's other three lines (ADR 0114 PR 3) are ordinary
triggers the seat controls: the loot's discard is the usual discard
prompt, and the end-of-combat sacrifice and the life loss ask nothing.

## Never offered a banned cast (#760)

The announce-time cast gate (`game.CastGateLocked`) is called once per
candidate cast in `castMovesForCard`, and it is the SAME function
`CastSpell` refuses with. A bot under a Rule of Law that had already
cast a spell would otherwise be offered the cast, refused, and offer
it to itself again on the next decision — the stall that shared
predicates exist to prevent. The view's `cant_cast` stamp is the third
reader of the same answer, so the human client greys exactly what the
bot is not offered.

## Another player's "any player may activate" ability (#1793)

Some activated abilities say "Any player may activate this ability"
(Xantcha, Sleeper Agent, Feral Hydra, Excavation; CR 602.2). The
enumerator offers such a row to every seat with priority and the mana
to pay, on whoever's permanent it is, through the same
`game.MayActivate` the engine refuses on (ADR 0106 §1). The move's
label names the permanent's controller — "Xantcha, Sleeper Agent
(controlled by Alice): {3}: …" — so a model tier can read whose
permanent it is reaching across to.

The heuristic takes such a move on ANOTHER player's permanent only when
the catalog row declares a purpose for the activator (owner decision 2,
ADR 0106 §1 decision 8). The purpose rides the wire row as
`purpose: {draws, controller_loses_life}`, and the policy prices it as
the cards it draws plus the life the controller loses, the second only
while that controller is a live opponent whose life can matter. A row
with no purpose scores below passing, so a bot never pumps an
opponent's Flailing Ogre or grows an opponent's Feral Hydra with its own
mana. Its own permanents' rows are scored like any other ability.

## Never offered an attack it cannot pay for (#1063)

The same shape, one step over. Propaganda, Ghostly Prison and Sphere of
Safety make attacking their controller cost mana at CR 508.1a
([ADR 0080](decisions/0080-attack-taxes.md)), and an attack tax is the
first thing that can make a DECLARATION fail for want of it.

`legal/combat.go`'s attack arm prices every candidate through
`game.PriceAttackDeclarationForEffect` — the SAME function the
declaration verbs charge with — and drops the move when neither the
seat's mana pool nor the auto-tapper can cover it. A move that is
offered on the strength of the tapper carries `auto_tap` in its params,
because `Move.Params` is exactly the payload that performs the move.

Two things follow that are worth stating:

- **The policy weighs the price, it does not merely tolerate it.**
  `MoveCost.mana` carries the cost string onto the wire (a policy may
  not import `internal/game`, ADR 0033 §3, and an attack has no printed
  cost to read off the CardView), and `attackValue` subtracts its mana
  value times `Config.AttackTaxPenalty` — 0.30 by default, so one point
  of tax is worth about one point of power getting through. A bot with
  two lands and a 1/1 passes rather than spending its turn for two
  damage. `LethalBonus` still dwarfs the term, so a lethal swing
  happens at any price the seat can pay.
- **Per-creature enumeration composes with a per-declaration charge.**
  The enumerator offers one move per (attacker, target) pair and the
  engine charges one declaration verb call at a time, so three separate
  attacks under Propaganda pay {2} three times — the same {6} one bulk
  `declare_attackers` pays. After each declaration the seat's mana has
  shrunk and the next enumeration prices the next attack against what
  is left, so a bot cannot strand itself mid-swing.

The human client reads the same price from
`turn.attack_targets[].tax` and labels its attack controls with it, so
what the bot is refused and what a person is warned about come from one
number.

## Cashing an edge two turns out (#1409)

The heuristic plans an attack one creature at a time, and two checks
used to decide it: `lethalPush` ("does everything I have kill them
now?") and `attackValue` ("is this one attack a good trade?"). Between
them sat the board that stalled the S31 gate: two seats under 7 life
behind full, even boards, one side a Drake up. All-in puts one Drake
through for 3, which is not lethal, and every single attack is blocked
at a loss, so neither seat declares the first attacker and the game
waits for a library to run out.

`heuristic/race.go` adds the two-turn race, three numbers computed
with the same blocker matching `lethalPush` uses:

- **now**: what the swing connects for after the defender blocks to
  survive, losing as few creatures as it can;
- **next**: what the bot's survivors connect for next turn into what
  the defender kept;
- **crack-back**: what every opponent's creatures connect for on the
  turn in between, all of them swinging at the bot, into what it kept
  home.

The bot commits when now + next is at least the defender's life and
the crack-back is less than its own. The swing is the smallest one that
does it, tried evasive-first, and every creature outside it stays home,
because that reserve is what the crack-back check counted on. The
decision's reason says so in those numbers: `attack: two-turn race — 3
now + 3 next turn ≥ their 6 life; crack-back 0 < my 6`.

It is deliberately pessimistic wherever the estimate could make the bot
suicidal: no blocked attacker of its own is home for the crack-back, the
defender counts as both attacking with everything and keeping
everything home, a trampler is held only by a blocker that absorbs all
of it, and menace counts as unblockable. A blocked attacker counts
toward next turn only when nothing the defender could put in front of
it kills it (#1527), spare blockers ganging up on it included (#1548,
below). A perfect mirror has no race — nothing is left over for next
turn — and the bot does not invent one.

### Gang blocks (#1548)

A block move names one pair, so two Ogres on a Wurm are declared one at
a time. The block planner scores each one as the whole group in front of
the attacker, less what the group without it already did:

- **the kill**: does the group kill the attacker now, when it did not
  before? Combat damage is played in its two steps: a first striker
  picks off the blockers it can before they deal regular damage, and
  spends that damage on the ones that would finish it; a deathtoucher
  needs one point per blocker; a double striker hits twice.
- **the losses**: the attacker divides its damage as badly for the
  defender as it can, so the group loses the most value whose lethal
  damage adds up to the attacker's power. A second Ogre on a chumped
  Wurm costs nothing, since the Wurm's 7 kills one Ogre either way.
- **the damage saved**: a trampler puts over what its blockers do not
  absorb, so a second chump on a Wurm saves its toughness. The same
  overflow counts toward whether the bot is desperate enough to chump.

The race predicts the same defender. Once the defender would survive,
the model adds spare blockers to any block that does not kill its
attacker when the kill is worth what they cost, and the joiners soak up
a trampler's overflow as they go. A blocked attacker of the bot's lives
into next turn only when its blockers plus every spare blocker that
could join them cannot kill it.

### The attrition horizon (#1548)

Gang blocks take away a lot of two-turn kills. The full seed-1409
Wurm-edge mirror (two Wurms, six Ogres, five Bears and a 3/3 commander
against one Wurm and the rest the same, both players on 6) has none
left: whatever the bot sends, a gang-blocking defender holds the two
turns to 2 damage. When no race exists, `heuristic/attrition.go` plays
the race's own model forward. An exchange is acceptable when the
crack-back on the turn in between is less than the bot's life, it makes
progress the defender cannot trade away (damage that connects however
it blocks, or a creature of theirs that stays dead; a commander goes
back to the command zone, so trading into one is not progress), no
more of the bot's creatures die than the defender's, and the bot's
creatures are still worth more than the defender's afterwards. After
an acceptable exchange the board is rebuilt, with both sides untapped
and the defender's creatures home to block, and asked again: a lethal
all-in or a two-turn race ends the line, otherwise one more exchange,
up to three. The bot sends the first exchange of the first line that
ends in a race, and keeps everything else home, exactly as a race does:
`attack: attrition — 1 of mine for 1 of theirs and 0 sure now, a race
within 1 exchange(s); crack-back 0 < my 6`. In the Wurm mirror that is
one Wurm into theirs. Nothing is remembered between decisions, and the
search stops after 100 swings evaluated per decision.

### The free answer (#2310)

The race's **now** is what the defender takes when it blocks only to
survive. A defender has another cheap answer: block with everything
whose block costs it nothing, and take the rest. A block is free when
the blocker lives, or when the blocker is the defender's commander.
CR 903.9a sends a dying commander to the command zone, and the
defender casts it again before the bot's next attack. This is the
definition of a free block that the attrition horizon already used.

The race now has to win against both answers. It computes now and next
for each and decides on the pair that leaves the defender more life.
Under the free answer, an attacker of the bot's that a free block kills
is dead for next turn, and a commander that died blocking is back to
block. The crack-back is the larger of the two. Under the free answer
the defender loses none of the creatures its chumps would have cost it,
so they can all swing back. The recast commander cannot: it was cast
that turn, so it can attack only if it has haste.

Without this the full seed-1409 Wurm mirror looped. Once the Ogres had
traded off, the race sent a lone 3/3 commander: `3 now + 4 next turn ≥
their 6 life`. The defender blocked it with its own commander, both
were cast again, and the same race came back for about a dozen turns.
The loop ended only when the commander tax outran one player's lands.
Against the free answer that swing is 0 now and 2 next, so it is not a
race. The bot sends the smallest swing that is: its commander and two
Bears, `4 now + 2 next turn`.

Two refinements were measured and left out, because both made the
heuristic gate's games longer. One counted a commander as free only
while the defender could pay the tax to cast it again (CR 903.8). The
other let a trampler's overflow past a free chump through. Each makes
some smaller swing qualify, and the race and the attrition horizon
send the smallest swing that qualifies.

### Trample overflow on the attacking side (#1504)

`lethalPush` and the race's now and next count what a blocked trampler
puts over its blockers. The defender assigns each blocker lethal damage
(its toughness less the damage already on it, or 1 from a deathtouch
source) and the rest goes to the player, exactly as the engine and the
bot's own damage assignment do. So a Bear in front of a 7/7 trampler
still lets 5 through, and a one-Wurm edge is cashed the way a one-Drake
edge is. Before, a chumped Wurm counted as a Wurm that did nothing.

The count is a lower bound, never an overestimate. Which blockers go on
which trampler is a partition problem, so `heuristic/trample.go`
relaxes it to a matching: each trampler has slots, and a blocker in
slot j is credited with no more than it could absorb there. The bound
takes the larger of that relaxation and the plain blocker matching, and
a board with no trampler computes only the matching, so it is unchanged.
A first-strike blocker counts as holding a trampler in full, because it
may kill it before it deals any damage. `TestUnblockedPowerNeverOverestimates`
checks the bound against every possible set of blocks on 3000 random
small boards. On 2936 of them it is exact.

The gate that measures this plays **lockstep** (next section), so a
change to the race term moves the same forty games every time it runs.

## Replaying a heuristic gate seed (#1409)

`TestFourHeuristicBotsPlayToAWinner`, `TestTwoHeuristicBotsPlayToAWinner`
and `TestHeuristicBeatsRandomHeadToHead` run their seats on one
goroutine, in seat order: each seat's ordinary act-loop (the runner's
own `step`, reached through `aiseat.NewStepped` / `Runner.Step` since
#1503 — `lockstep_export_test.go` keeps #1409's names as one-line
delegations), round after round. A seed is then a whole game rather than a deal, and a
red night replays move for move. The soak (`TestRandomBotSoak`) keeps
one goroutine per seat on purpose: it tests concurrent liveness, and the
races lockstep removes are exactly what it exists to find (discussion
#1390). `AISEAT_HEURISTIC_SCHEDULE=concurrent` plays the gate the old
way, for comparing the two.

The arena (`boteval arena`) starts one goroutine per seat by default,
so its `--seed` fixes the deals and the policies' randomness but not the
interleaving. `boteval arena --lockstep` plays the gate's schedule, so a
seed is a whole game there too (#1503; see "Lockstep runs" above).

## Known limitations

Stated plainly, because most of them are design decisions rather than
bugs.

**Play quality is capped by catalog coverage, not by the policy.** At
several thousand catalogued cards, a bot is a competent player of a
deliberately small format. A better model does not move this ceiling;
more cards do. This is the honest expectation to set.

**The heuristic plays most of its deck, but not all of it.** S66 taught
it to price what a card does ([above](#how-the-heuristic-prices-a-card)).
At the S66 exit, in 64 four-deck games, each curated deck had at most
four non-land cards offered in five or more windows and never used:

- mono-black-aristocrats: Ashnod's Altar, Phyrexian Altar, Blood Artist,
  and Burnished Hart's activation;
- izzet-aggro: Lotus Petal, and Glint-Horn Buccaneer's and Professional
  Face-Breaker's activations;
- simic-ramp: Tarmogoyf, and Sakura-Tribe Elder's sacrifice;
- esper-control: Commander's Sphere.

The reasons are in the pricing:

- The Altars' mana abilities sacrifice a creature, so they are neither
  repeatable mana sources nor activated for floating mana. Lotus Petal's
  mana is one-shot, so it is not a mana source either.
- Blood Artist's 0/1 body and one triggered row are worth less than the
  card. A death payoff is priced while it is on the battlefield, when
  the bot chooses what to sacrifice, and not when it is cast.
- A one-mana rock drawn after the bot can already cast everything it
  holds is priced below zero on purpose. While the bot is short of mana,
  the rock often loses its window to the land drop or a bigger spell,
  and by the next window the deficit has closed. The exit run measured
  rocks and dorks cast in 210 of the 311 games in which they were
  offered while the bot was short, 68%.
- Harrow sacrifices a land as an additional cost. That is not a
  mana-and-taps cost, so it keeps the normal bar in the leftover
  windows, and two lands for one land and a card rarely clears it late
  in a game. It was cast in 8 of the 27 games it was offered in.
  Since #2469 a land sacrifice that the cast's `purpose.lands` more than
  replaces gets the leftover bar (see "The two leftover windows").

**A non-tap activated ability with no declared purpose is a flat
+0.50.** That is `ActivateBase`, and S66 left it alone (ADR 0126 owner
decision 6). Syr Konrad's `{1}{B}` mill is the visible case.
Mono-black-aristocrats still wins about 60% of four-deck games, but an
experiment that lowered `ActivateBase` cut Konrad's mills by a third
and left black's win rate where it was. The deck, not this price, is
what [#2436](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2436)
rebalances.

**No plan for the turn.** The heuristic prices one move at a time. It
does not cast a cantrip first to see what it draws before deploying
([#2458](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2458)), it
does not treat a cycling card as a cheap discard or its own draw step
as a spend window
([#2457](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2457)), and
it does not cast a ritual or crack a Treasure for a specific spell.

**Purposes are declared for the curated decks and for board wipes
only.** Any other instant or sorcery is priced by its mana value, with
`SpellFloor` under it, so it is cast when mana is spare and not for
what it does. A permanent's ability rows say that it does something,
not how much, so a drawback row counts as a plus. The combat planner
compares bodies only (`CombatValue`), so a Blood Artist blocks like a
vanilla 0/1.

**A few moves are refused on a busy table.** Each seat decides on the
board it saw, and another seat can commit in between. Two refusals in
the arena reports are races of that kind and cost nothing: the runner
decides again on the new board.

- **A pass while a prompt is unanswered.** Rhystic Study's "pay {1}?"
  does not stop the table ([ADR 0018](decisions/0018-triggers-on-the-stack.md)
  §6), so the other seats keep passing. When the payer declines, the
  Rhystic Study player's "draw a card?" opens, and that prompt does stop
  the table. A pass decided just before it opened is refused. Sun
  Titan's and other triggers' prompts race the same way.
- **An attack after the step moved on.** This one is a real defect,
  [#2462](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2462): the
  enumerator offers the active seat attacks after it has passed
  priority in declare attackers, and the runner takes one as its only
  move. Usually the attack lands, on a creature the policy had kept
  home; when the last seat passes first, the engine refuses it.

A cast refused for "insufficient mana" is also a defect, not a race:
the auto-tapper counts a bounce land's two mana as two of one colour
([#2461](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2461)).

**No politics, no deal-making, no bluffing, no table talk.** Bots
speak only to disclose an improvisation and, behind the setting
above, to explain a move. Commander is a political format and a bot
does not play that half of it. A four-bot table is four players who
never negotiate.

**No learning across games, and no opponent modelling.** Bots are
stateless between games. The one played you last night remembers
nothing about it.

**Improvisation does one clause at a time, and only a spell's.** It
fires when a spell the bot cast resolves into silence, and never for
an unimplemented permanent's triggers or activated abilities — so a
bot running an enchantment the engine will not honour plays as if the
enchantment were a blank piece of card for the rest of the game. That
is deliberate: a trigger that should have fired leaves no trace on the
board for anything to notice, and the alternative is a rate limit
pretending to be a rule.

**A card in a hidden or unordered zone has no value to the
evaluation.** `score.go` prices the battlefield and the seats; a card
in a graveyard, in exile or on top of a library is worth nothing to
it. Two consequences worth stating, because both look like bugs: the
bot cannot choose WHICH cards an escape cost eats (see
[above](#the-card-component-of-an-alternative-cost)), and a
graveyard cast whose payoff is not a permanent — Lingering Souls'
two tokens, made by a sorcery's resolution the scorer does not model
— is priced at roughly nothing and taken only when nothing else is
on offer. The move is enumerated either way; what is missing is the
number, not the option.

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

**A bot does not move an Equipment between its own creatures for
nothing.** Every equip used to be priced like a pump on the bot's own
creature, whatever the Equipment was already on, so a free one
(Lightning Greaves' Equip {0}) beat passing from either side. A
heuristic seat moved its Greaves back and forth until the CR 732
breaker named it, and the runner's hold on that permanent then stopped
the table
([#2449](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2449)).
Now an equip row says it is one on the wire (`equip`), and the
heuristic prices a move between two of its own creatures below passing
unless the new host, as it stands, is worth more than the current host
is with the Equipment on it. Re-equipping the creature it is already on
does nothing (CR 701.3b) and is never taken. The first equip of an
unattached Equipment keeps its old price. Comparing the new host
without the Equipment against the old one with it is what makes the
Equipment come to rest: the move back can never also look better.
Choosing the best host, such as moving Greaves onto a summoning-sick
creature for haste, is still out of scope ([ADR
0126](decisions/0126-bots-that-play-their-decks.md)).

**A bot does not untap a Monolith for no net mana.** Basalt Monolith
taps for {C}{C}{C} and untaps for {3}, and priced like any activation of
the bot's own, "{3}: Untap Basalt Monolith" beat passing every time: tap
for three, pay three, again. A heuristic seat repeated it until the CR
732 breaker named it and the runner's hold parked the seat
([#2500](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2500), the
same shape as #2449). A row whose whole effect is untapping its own
source says so on the wire (`untap_self`, declared by hand on the card
file), and the heuristic prices it below passing unless the source's
best repeatable mana ability makes more than the untap costs. Basalt
nets 0 and Grim Monolith (`{4}`) nets -1, so neither is taken. This is
#2493's "no gain, no move" for an activated row that nets nothing; it
does not model an untap that enables a spell this window, because the
mana that spell would spend is the mana the untap costs.

**An X paid in LIFE is priced the same way, one short of the life
total.** Toxic Deluge is `{2}{B}` with no `{X}` in it — its X is
announced by paying X life as an additional cost — so the affordable-X
search saw no `{X}` slot, announced 0, and offered a board wipe that
swept for -0/-0
([#957](https://github.com/krakenhavoc/cmd_and_ctrl/issues/957)). The
same floor applies (1, because the card declares `XMatters`), the
ceiling is the seat's life total minus one — CR 119.4 forbids paying
more life than you have, and a sweep a bot does not survive is not the
one offer to hand it — and `MaxX` caps it like any other X. A bot at 13
life is offered Toxic Deluge at X=12; a bot at 1 life is not offered it
at all. The rule is keyed on the cost component, so the next card that
prints "pay X life" is priced without a line of its own.

**Phyrexian symbols can be paid with life, on spells and abilities
alike.** A `{B/P}` is one black mana or 2 life (CR 107.4f), and the
caster announces how many symbols the life buys as `phyrexian_life`.
Abilities have had this since
[#917](https://github.com/krakenhavoc/cmd_and_ctrl/issues/917). Casts
priced only the mana until
[#1677](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1677), so a
bot with one Swamp was never offered Dismember at any life total. Each
cast now gets at most two payments. The first is all mana when the
seat can pay it, and otherwise the fewest symbols paid with life that
make the cast affordable. The second pays every symbol with life, and
is offered once per card out of the budget the targets left over. The
counts in between are the same trade paid partly, so they are not
offered. Every count is bounded by the engine's own life predicate
(CR 119.4, and CR 119.8's locked life total). The move label says
`paying 4 life instead of mana`, and `Move.Cost` carries the life
twice: in `life`, with the rest of what the move charges, and in
`phyrexian_life`, the part that buys nothing extra.

That second field is how the heuristic gets it right. Other life costs
earn a small payoff (`LifePayoff`), because what a card charges is the
only evidence of what it does. Phyrexian life gets no payoff: it buys
the same spell the mana would. So wherever the seat can pay mana, the
mana payment scores higher. The heuristic also never pays Phyrexian
life that would leave it below 10 (`phyrexianLifeFloor`, the default
`DangerLife`). The random tier picks among the legal moves like any
other.

The same payment is offered for a `{B}` under K'rrik, Son of Yawgmoth
([ADR 0131](decisions/0131-krrik-pay-life-for-black-mana.md)), and for
the two payments that are not casts: a mana ability's own mana cost (the
move carries `phyrexian_life`) and a mana `pay_unless` such as ward
(the "pay" answer carries it). The enumerator offers the fewest symbols
paid with life that make the move payable, after mana. The heuristic
prices a life-paid ward like a Phyrexian cast: declined below the floor
of 10, and below the all-mana answer above it.

**"Discard your hand" is priced by the hand it throws away.** Lion's
Eye Diamond, Diamond Lion, Null Brooch and Slate of Ancestry
([#1600](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1600)) pay
with every card in hand, so the move names none of them: the engine
refuses `discard_ids` for the clause. `Move.Cost.Hand` carries the
count instead (the activator's hand, less the source if it is in that
hand), and the label says `discarding your hand (N cards)`. The
heuristic charges `Weights.Hand` per card, the price a cast already
pays for leaving the hand, so Slate of Ancestry with seven cards in
hand ranks well below the same activation with an empty one, and an
empty hand costs nothing. Lion's Eye Diamond and Diamond Lion are only
offered while the seat could cast an instant: holding priority, owing
no prompt, with no prompt stopping the table. The auto-tapper never
cracks one to pay for a cast.

**Energy is priced at a flat amount per counter** ([ADR
0129](decisions/0129-energy-getting-and-paying-it.md) §7, owner decision
5). An activation that pays energy is only offered when the seat has it,
through the predicate the engine refuses with, and `Move.Cost.Energy`
names the counters it removes: "Pay N {E}", or N + X for "Pay X {E}",
whose X the enumerator bounds by the seat's energy. The heuristic charges
`Weights.Energy` (0.30, a quarter of a card in hand) per counter spent,
and credits the same per counter a declared `energy` purpose gives, so a
sink is used when what it buys is worth more than its counters. The
baseline prices energy at nothing. The model tiers and the MCP seat read
each seat's energy and other player counters on its board-text line
("4 energy, 2 poison").

**A spell whose target count is X is offered with X equal to the
number of targets it picks.** Crackle with Power deals five times X
damage to each of up to X targets, so the count and the announcement
are one decision, not two — the enumerator used to make them
separately and offer one target at X=0, which the engine refused
outright
([#619](https://github.com/krakenhavoc/cmd_and_ctrl/issues/619)). A bot
now sees Crackle at one target for X=1, two for X=2, and so on as far
as its mana reaches, and never sees the cast that bounces.

**A bot takes one CR 732 shortcut per loop per turn, then stops.** When
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

**A real loop stops a bot-only table outright; a batch of triggers
does not.** When a trigger or activation loop's shortcuts run out (the
second ask offers only "stop here"), autopass stays suspended and no
bot seat has a move that keeps the loop going. The room's commit
sequence stops with the notice naming the ability and count still in
the game state. The table does not finish, and it does not spin forever
([ADR 0055](decisions/0055-loop-breaker.md) §5).

A batch is different. Three of ADR 0126's measurement tables stopped
at the breaker after a wrath killed about thirty creatures with
mono-black-aristocrats' Syr Konrad and Exquisite Blood out: each Konrad
trigger pinged three opponents, each ping triggered Exquisite Blood, and
Blood's count passed 25 with nobody deciding anything
([#2450](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2450)).
That is a finite number of triggers from one event (CR 603.2c), not a
loop. Since ADR 0055's amendment of 2026-10-07 the engine measures the
work left on the stack as each triggered ability begins to resolve, and
a new low restarts every run: a batch keeps making new lows and runs to
the end, and a loop, which replaces what it resolves, trips the breaker
as before. It is engine-wide, so a human's autopass plays through a
batch too. (Sanguine Bond with Exquisite Blood never tripped it: each
Bond trigger asks for a target, and answering is a decision.)

**A loop that is ending the game plays out.** Exquisite Blood with
Marauding Blight-Priest is a real mandatory loop, and it grows the
stack rather than draining it, but every iteration takes each opponent
to a new low. A life total makes only so many lows before its player
loses, so the engine counts a new lowest life total this turn, a new
highest poison count this turn, or a player leaving the game as
progress, and progress restarts the runs (ADR 0055's amendment of
2026-10-07, option B). The loop steps on until its controller wins. A
player who can't lose the game that way (ADR 0057) makes no progress
that way. A loop that makes no progress, such as gaining life forever,
still trips the breaker, and still stops a bot-only table.
