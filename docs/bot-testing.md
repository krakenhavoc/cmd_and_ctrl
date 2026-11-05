# AI bot seat: the harness and its test knobs

The commands and the whole-game test gates for the bot seat. The seat itself is in [bot.md](bot.md).

## boteval commands

- `boteval suite run [--dir DIR] [--policy heuristic|assisted|strong] [--max-think 20s] [--think] [--max-tokens N] [--note text] [--parallel N] [--out report.json] [--md]` — asks a policy every labelled position in `server/internal/aiseat/suite/testdata/positions/` and reports agreement overall and per tag, plus reject-hits, malformed replies, out-of-range indices and timeouts. Exits non-zero on a gated miss. The `heuristic` run is also an ordinary Go test and runs on every CI run.
- `boteval suite harvest --from 'dir/*.decisions.jsonl' --to inbox/ [--escalated] [--disagree] [--fallback a,b] [--layer A,B,C] [--seat 0,1] [--tag block,attack] [--limit N] [--seed N]` — pulls candidate windows out of decision logs into an inbox of UNLABELLED positions. Deterministic under `--seed`.
- `boteval suite render --pos path/to/position.json [--deck ID]` — prints the exact prompt a model would see for one position and the move list with `<- accept / reject / heuristic / model@capture` markers. This is the labelling screen. See [docs/bot.md](bot.md#position-suite).
- `boteval arena --seats a,b,c,d [--decks …] --games N --rotate --out DIR [--think] [--max-tokens N]` — headless bot-vs-bot games with the report block ADR 0052 asks every bot PR to carry: win rate with a **Wilson 95% interval** against the table's null rate (1/seats), per policy and per policy-and-deck, ADR 0126's Cards section (each card's offers and uses, the `never` list, the A2/A3 canaries), the funnel's layer/escalation/timeout counters, and decision + model-call latency tails. Rotation seats contestant `k` at position `(k+i)%n` in game `i`, so turn order cancels. `heuristic-baseline` seats the heuristic frozen before S66 (`heuristic.BaselineConfig`, an arena name only). A model tier with **no endpoint is refused, not downgraded** (an `assisted` seat with no client plays the heuristic under a model tier's name). A stall is **reported, not fatal**. Artifacts land in `<out>/<RFC3339 start>/`: `summary.md`, `summary.json`, `games.jsonl` (streamed per game), `decisions/`, `replays/`. Wall clock: ~0.2 s per two-seat heuristic game, ~3.5 s per four-seat curated-deck game, 5–15 min per game with one local-model seat. See [docs/bot.md](bot.md#arena).

## Burnished Hart's land search by the turn (#5084, S75)

Burnished Hart's "{3}, Sacrifice Burnished Hart: Search your library for
up to two basic land cards" priced its two lands (2.00) plus a ramp
premium from the hand's deficit alone, against the Hart's board value
(about 3.30) and, a second time, the flat 0.30 for an untapped creature.
`PriceRampCurve` ([bot.md](bot.md#a-land-search-that-sacrifices-its-source-by-the-turn-5084))
prices such a row's lands up to the mana a deck on curve wants two turns
ahead while the round is below seven, and drops the double charge. It
**ships off**: the arena below found no gain. `heuristic-rampcurve` seats
today's heuristic with it on. The position tests
(`heuristic/ramp_curve_test.go`) and the real-engine regression
(`aiseat/ramp_curve_bot_test.go`) hold both sides: with the curve on, the
Hart is cracked in rounds 3 to 5 on curve and in round 6 a land behind,
with a castable hand (+1.10 against −0.20 to −1.20 off; the real card
+0.70), and two basic Islands enter tapped; it is kept from round 7, on
a stalled board facing two 7/7s, and a land or more behind. The default
keeps the Hart in every one of those positions. Ramp spells, rows that
declare no lands and a hand that already asks for the mana are priced as
before.

The 2026-11-05 UTC runs used develop `f01023882` plus #5084 and the
synthetic `--decks thopter-battle` at all four seats, rotation and
lockstep. Every game ended with a winner and none stalled. First,
curated-arena.sh's shape (16 games, seeds 1000–1015) with all four seats
on one policy:

| Policy | Hart activations (seat-games) | Hart windows offered | Wins |
|---|---|---|---|
| curve on | 10 (9 of 43) | 3,872 | 16 / 64 |
| curve off (today) | 9 (8 of 44) | 3,941 | 16 / 64 |

Fifteen of the sixteen games had the same length either way. Decision
logs of four of those games with the curve on show why: in 108 windows
of the Hart's controller's own main phase with an empty stack, 96 were
from round 7 on, when the curve has closed, and in the 12 earlier ones
the bot cast what it held, played a land or passed in round 6 with six
sources, where the curve adds one mana and the Hart prices −0.30.

Then head-to-head: two seats with the curve against two without, seeds
2000–2047 (48 games). The four-seat null rate is 25%.

| Policy | Wins / 96 seat-games | Hart activations (seat-games) |
|---|---|---|
| curve on | 21, 21.9% (14.8–31.1%) | 10 (10 of 65) |
| curve off | 27, 28.1% (20.1–37.8%) | 14 (14 of 66) |

Two seats of one policy share a game, so the intervals are conservative.
The intervals overlap and the curve seats made fewer Hart activations,
not more, so the gap is the games the change re-dealt, not the curve. On
this deck the Hart arrives late and its {3} competes with Drakes and
Ogres the turn plan prices higher, so the price is not what keeps it on
the battlefield. A ramp deck with big spells in hand, or the curated
decks, is the run to make before turning the switch on. The labelled
position suite (42/42), the heuristic, suite, botarena, boteval and
aiseat package tests and docsguard passed. With the switch off,
`DefaultConfig` plays exactly as before. Reports stayed on the authoring
machine.

## Mana sinks and graveyard exiles (#4559, S75)

`heuristic-nomanasink` is today's heuristic with `SinkManaLast` and
`PriceGraveyardExile` off. Compare it with `heuristic` on the same seeds
and rotated chairs; the curated `boros-voltron` and `mono-white-weenie`
decks hold Lion Sash. Decision logs show a sink taken last with the
reason suffix `mana sink last`; count Sash activations made while a cast
that cleared the bar was on offer, and the exiled card's owner and type.

The unit tests (`heuristic/mana_sink_test.go`) cast a cheap creature
before the Sash's exile, take the exile once nothing else clears its
bar, order exile picks as an opponent's creature, their land, the bot's
own creature, their instant, and cast a cheap instant before the sink in
the end step before the bot's turn, and hold the sink in the bot's own
upkeep for the main phase. The ungated whole-engine test
(`aiseat/mana_sink_bot_test.go`) gives the bot the real Lion Sash, two
Plains and Mother of Runes in hand: backed out, it exiles its own
Baneslayer Angel twice (priced +8.09 as recursion) and never casts
Mother; with the change it casts Mother, then exiles the opponent's Sun
Titan with the other {W} and the Sash gets its counter.

No arena comparison was run with the change: the curated decks need a
Scryfall dump the authoring machine did not have, and the synthetic
battle decks hold no graveyard-exile sink. The labelled position suite
kept its agreement.

## A counter on its own source (#5083, S75)

Hangarback Walker's "{1}, {T}: Put a +1/+1 counter on this creature."
declared only that it answers (`pump`), so it was priced at the flat
`ActivateBase` (0.50) less the blocker its tap costs (0.30): 0.20 on the
bot's own turn, under `PassThreshold` (0.25). The row now declares
`source_counters: 1`, and `PriceSourceCounters` prices each counter at
`Weights.Power` plus `Weights.Toughness` (1.45; [bot.md](bot.md#a-counter-on-its-own-source-5083)).
It **ships off**: the arena below found the price a loss.
`heuristic-sourcecounters` seats today's heuristic with it on. The
position tests (`heuristic/source_counter_test.go`) and the real-engine
regression (`aiseat/source_counter_bot_test.go`) hold both sides: with
the price on, the Walker grows in the bot's own main phase with two
spare Islands, keeps its attack when the attack is lethal, keeps its
block on an opponent's turn, and gives way to a castable 3/3 flier; with
it off the bot passes on its own turn. Both grow it in the end step
before the bot's untap.

The 2026-11-05 UTC runs used develop `f01023882` plus #5083 and the
synthetic `--decks thopter-battle` at all four seats, rotation and
lockstep. Every game ended with a winner and none stalled.

First, curated-arena.sh's shape (16 games, seeds 1000–1015) with all four
seats on one policy, and decision logs:

| Policy | Walker activations | When | Mean turns |
|---|---|---|---|
| price on | 185 | own precombat main 177, own postcombat main 8 | 15.1 |
| price off (today) | 67 | all in the end step before the bot's untap ("leftover mana") | 14.3 |

With the price on, 126 of the 177 precombat activations were made with no
cast on offer. The Walker was never grown on an opponent's turn, so it
was tapped through each of them.

Then one head-to-head: two seats with the price on against two without,
seeds 2000–2047 (48 games). The four-seat null rate is 25%.

| Policy | Wins / 96 seat-games | 95% CI |
|---|---|---|
| price on (`heuristic-sourcecounters`) | 17, 17.7% | 11.4–26.5% |
| price off (today's `heuristic`) | 31, 32.3% | 23.8–42.2% |

Two seats of one policy share a game, so these trials are not
independent and the intervals are conservative. The seats that grew the
Walker before combat did worse. Grown on its own turn, the Walker stays
tapped through every opponent's turn; grown in the end step before the
bot's untap, it gets the same counter from the same mana and blocks all
round. The blocker price (0.30) is too small for a round of blocks
against a counter one window away, so the flat price's timing was the
better play. A price that waits for that end step, or one that counts
the blocks the opponents' turns would ask for, is the next thing to
measure before turning the switch on. The labelled position suite and
the package tests passed. Reports and decision logs stayed on the
authoring machine.

## Riders: scry, learn, discover, amass and the Ring (#4978, S75)

`heuristic-noriders` is today's heuristic with `PriceRiders` off: every
declared `scry`, `learn`, `discover`, `amass` and `ring_tempts` is
priced at nothing again. The position tests
(`heuristic/riders_test.go`) hold each rider's weight, the back-out, and
the issue's acceptance case: Preordain over a plain cantrip of the same
cost, and Claim the Precious over the same removal without its
tempting, with the proxy kept. `TestRiderPurposesMatchThePrintedCard`
(`cards/coverage`) holds the catalog's declarations to the card text.

The 2026-11-05 UTC runs used the #4978 branch at `734108a46` (on that
morning's develop, before its rebase), the standard pool (mono-red-aggro, simic-ramp,
esper-control, mono-black-aristocrats), seeds 2000–2047, `--rotate
--lockstep`, two seats of each policy. Rotation moves each chair's deck
with it, so a second run swapped the policies between the chairs, and
each policy played each deck 48 times. All 96 games ended with a winner
and none stalled.

| Deck | heuristic: wins / 48 | heuristic-noriders: wins / 48 |
|---|---|---|
| mono-red-aggro (Fiery Inscription) | 20, 41.7% (28.8–55.7%) | 20, 41.7% (28.8–55.7%) |
| simic-ramp (Preordain) | 6, 12.5% (5.9–24.7%) | 6, 12.5% (5.9–24.7%) |
| esper-control (Read the Bones) | 7, 14.6% (7.2–27.2%) | 8, 16.7% (8.7–29.6%) |
| mono-black-aristocrats (Read the Bones) | 14, 29.2% (18.2–43.2%) | 15, 31.2% (19.9–45.3%) |
| all four | 47 / 192, 24.5% (18.9–31.0%) | 49 / 192, 25.5% (19.9–32.1%) |

The four-seat null rate is 25%. The riders on these decks are small
(one scry 2 spell per deck, one tempting), and the win rate did not
move beyond noise in either direction. What moved is when the cards are
cast: priced, Read the Bones was taken in 7 of the 21 aristocrats windows
it was offered in (without the price, 7 of 48), and Preordain in 10 of 44
(11 of 58), each used in the same share of games. Fiery Inscription was
used in 6 of 6 games with the price and 5 of 6 without. Charming
Prince's bullet is on mono-white-weenie, which these runs did not seat;
its position test covers it. The labelled position suite kept its
agreement. Reports are local under `data/arena-4978/` (gitignored).

## Spare tokens for Sai's draw (#4991, S75)

Sai, Master Thopterist's "{1}{U}, Sacrifice two artifacts: Draw a card"
was never taken in the issue's 16 arena games of the curated
`mono-blue-artifacts` deck. Its row declared no draw, so it was priced
at the flat `ActivateBase` (0.50), and each sacrificed Thopter cost its
board value (2.95 for a 1/1 flier): about −5.70 against a pass. The row
now declares `draws: 1` (1.20). `PriceSpareTokens` prices a spare token
at the blocker it is ([bot.md](bot.md#sacrifices)), and it **ships off**:
the arena below found the trade a loss. `heuristic-sparetokens` seats
today's heuristic with it on. The position tests
(`heuristic/spare_token_test.go`) and the real-engine regression
(`aiseat/spare_token_bot_test.go`) hold both sides: with the price on,
four Thopters after combat trade two for a card and keep two; three
Thopters, Thopters before they attack, nontoken artifacts and a copy
with a counter are kept; with it off the bot passes.

The curated deck was not run: it needs a Scryfall dump, and the
authoring sandbox could not reach Scryfall. The 2026-11-05 UTC runs used
develop `a2035a940` plus #4991 and the synthetic `--decks thopter-battle`
at all four seats (Sai, eight Ornithopters, six Memnites, four Hangarback
Walkers, four Burnished Harts, four Mind Stones and vanilla blue
creatures), rotation and lockstep. Every game ended with a winner and
none stalled. They were played while the price was still on by default,
with the off seats named `heuristic-nosparetokens`; those seats play
exactly as today's `heuristic`, and the on seats as `heuristic-sparetokens`.

First, curated-arena.sh's shape (16 games, seeds 1000–1015) with all four
seats on one policy, and decision logs:

| Policy | Sai activations (seat-games) | When | Paid with |
|---|---|---|---|
| price on | 42 (21 of 64) | 30 after combat, 3 before it, 9 after blockers | two Thopters 34, a Thopter or Memnite with an Ornithopter or Memnite 8 |
| price off | 5 (5 of 64) | all 5 after blockers | two Thopters 1, the other 4 a Memnite or Thopter with an Ornithopter |

With the price on, Sai was never activated outside combat while its
controller held three Thopters (1,908 windows) and was activated in 24 of
242 windows with four. Hangarback Walker's "{1}, {T}" was used 83 and 67
times and Burnished Hart's sacrifice 8 and 9 times, so neither is never
used on this deck. Mind Stone's draw was never taken in either run.

Then head-to-head: two seats with the price on against two without,
seeds 2000–2047, 3000–3047 and 4000–4047 (48 games each). The four-seat
null rate is 25%. `SpareTokenKeep` was varied in temporary builds.

| Keep | Seeds | Price on: wins / 96 | Price off: wins / 96 |
|---|---|---|---|
| 2 | 2000 | 16, 16.7% (10.5–25.4%) | 32, 33.3% (24.7–43.2%) |
| 2 | 3000 | 17, 17.7% (11.4–26.5%) | 31, 32.3% (23.8–42.2%) |
| 2 | 4000 | 25, 26.0% (18.3–35.6%) | 23, 24.0% (16.5–33.4%) |
| 2 | all | 58 / 288, 20.1% (15.9–25.1%) | 86 / 288, 29.9% (24.9–35.4%) |
| 3 | 2000 | 24, 25.0% (17.4–34.5%) | 24, 25.0% (17.4–34.5%) |
| 3 | 3000 | 17, 17.7% (11.4–26.5%) | 31, 32.3% (23.8–42.2%) |
| 3 | 4000 | 23, 24.0% (16.5–33.4%) | 25, 26.0% (18.3–35.6%) |
| 3 | all | 64 / 288, 22.2% (17.8–27.4%) | 80 / 288, 27.8% (22.9–33.2%) |
| 4 | 2000 | 17, 17.7% (11.4–26.5%) | 31, 32.3% (23.8–42.2%) |

Two seats of one policy share a game, so these trials are not
independent and the intervals are conservative. At every keep the seats
that cashed in Thopters did no better, and at the default keep of two
they did worse. On this deck the draws are vanilla creatures and lands,
and the Thopters are its evasion, so the board price was right. The
curated deck draws better cards, and is the run to make before turning
the switch on. The labelled position suite, the gated heuristic games
(`AISEAT_GAME_TESTS=1`, the four- and two-seat heuristic tests) and the
package tests passed. Reports and decision logs stayed on the authoring
machine.

## Spare mana rocks for Mind Stone's draw (#5085, S75)

Mind Stone's "{1}, {T}, Sacrifice Mind Stone: Draw a card" was offered
about 6,000 times in #4991's 16 games and never taken. Its row already
declared `draws: 1`; the price was the cause. The card is
`Weights.Hand` (1.20) and the stone its board value as an untapped mana
source (`ManaSource`, 1.00), so the activation netted 0.20, under
`PassThreshold` (0.25) in every window. `PriceSpareManaSources` prices a
spare rock at `SpareManaSourceCost` (0.50; [bot.md](bot.md#sacrifices)),
and it **ships on**. `heuristic-nosparemana` seats today's heuristic
with it off. The position tests (`heuristic/spare_mana_test.go`) and the
real-engine regression (`aiseat/mind_stone_bot_test.go`) hold both
sides: with seven lands and an empty hand the stone is cashed in (0.70);
with three or four lands, or a hand card that still wants the stone's
mana, it is kept; a mana creature is never spare; with the price off the
bot passes.

The 2026-11-05 UTC runs used develop `f01023882` plus #5085 and the
synthetic `--decks thopter-battle` at all four seats, rotation and
lockstep. Every game ended with a winner and none stalled.

First, curated-arena.sh's shape (16 games, seeds 1000–1015) with all
four seats on one policy:

| Policy | Mind Stone activations | Offered windows | Seat-games used |
|---|---|---|---|
| price on (`heuristic`) | 44 | 1,168 | 32 of 41 |
| price off (`heuristic-nosparemana`) | 0 | 5,771 | 0 of 41 |

With the price on, the decision logs put every activation at +0.70 in
the bot's own main phase (31 before combat, 13 after), from game turn 5
to 18 (median 8), with six to eleven mana sources counting the stone and
at most three cards in hand. Hangarback Walker's activation (56 against
67) and Burnished Hart's (5 against 9) moved little; Sai's went from 5
to 8.

Then head-to-head: two seats with the price on against two without,
seeds 2000–2047 (48 games). The four-seat null rate is 25%.

| Policy | Wins / 96 | Mind Stone activations |
|---|---|---|
| price on | 23, 24.0% (16.5–33.4%) | 95 |
| price off | 25, 26.0% (18.3–35.6%) | 0 |

Two seats of one policy share a game, so these trials are not
independent and the intervals are conservative. Two wins in 96 is well
inside the noise: the trade neither wins nor loses measurably on a deck
whose draws are vanilla creatures and lands, and it ships on because a
deck that draws better cards gets more for each stone. Back it out with
`PriceSpareManaSources` if a curated-deck run says otherwise. The
labelled position suite (42 positions, 100% agreement), the gated
heuristic games (`AISEAT_GAME_TESTS=1`, the four- and two-seat heuristic
tests) and the package tests passed. Reports and decision logs stayed on
the authoring machine.

## What #4312 left (#4324, S75)

**Declared removal in response windows (measured, unchanged).** #4312
took the mana-value proxy off declared removal, so Dismember is priced as
the removal less the card. The 2026-11-04 UTC run used integration base
`43de8af88`, `--decks tricks-battle` at all four seats (three Dismember
each), seeds 1000–1047, rotation, lockstep and decision logs: 48 games,
all to a winner, none stalled. Dismember was offered in 126 seat-games and
cast in 116; 49 casts were in response windows (median price +2.53) and
89 in the caster's own main phase. Of the 440 response windows where the
bot passed with Dismember priced above zero but at or below
`InstantThreshold`, the target was a 1/1 Mother of Runes (+0.32) or a
2/2 Bear for 4 life (+1.45; the deck has no black mana, so every cast
pays Phyrexian life) in all but two, a 3/3 in the bot's own draw step,
and none of those targets was attacking. Those
holds cost no combat answer, and the bot cast Dismember on Bears and
Mothers in its own main phase instead (19 of the 89 main-phase casts). So the response bar stays as it is;
`TestDeclaredRemovalClearsTheResponseBar` holds a 2/2 kill above it on
an opponent's turn and leaves a 1/1 for the end step's leftover bar.

**Haste for a {T} ability.** `PriceHasteTapRows` (on by default, off in
`BaselineConfig`) is the back-out; [bot.md](bot.md#purposes-what-a-spell-or-an-ability-does)
says what it prices. No curated or synthetic deck holds a haste grant,
so the evidence is `heuristic/haste_tap_test.go`, and the arena is the
check that nothing else moved. With the change, the same tricks-battle
run and the std pool (mono-red-aggro, simic-ramp, esper-control,
mono-black-aristocrats; seeds 1000–1047, rotation, lockstep, all four
seats `heuristic`) were identical to the base game for game (turns, each
seat's final life, decisions and card offers and uses in all 96 games),
with no stalls. Std wins per deck in both: esper-control 12/48, 25.0%
(14.9–38.8%); mono-black-aristocrats 10, 20.8% (11.7–34.3%);
mono-red-aggro 16, 33.3% (21.7–47.5%); simic-ramp 10, 20.8%
(11.7–34.3%). The labelled suite stayed at 42/42. Reports, decision logs
and the scanner stayed on the authoring machine.

## Modal trigger bullets (#4674, S79)

`heuristic-nomodepicks` is today's heuristic with `PriceModePicks` off: every
`mode_pick` answer scores 0 again, so an "up to" prompt is answered with none.
The fixed policy prices each answer by its bullets' purposes and best legal
targets (`mode_bullets`, [bot.md](bot.md#enumerating-a-modal-announcement-764)).
The cards it changes (Ertai Resurrected, Hullbreaker Horror, Ravager Wurm, The
Ruinous Wrecking Crew) are in no curated deck, so the evidence is the
position tests (`heuristic/mode_pick_test.go`) and the real-engine
regression (`aiseat/mode_pick_bot_test.go`: Ertai destroys an opponent's
creature with the price on and chooses none with it off, spares its own,
and the Crew at X = 2 takes the rummage and the drain).

The arena is the check that nothing else moved. The 2026-11-03 UTC run used
develop `e1b6012f` plus #4674, the four pool2 decks, seeds 2000–2023,
rotation and lockstep, all four seats `heuristic` and then all four
`heuristic-nomodepicks` (24 games each). All 48 ended with a winner and none
stalled. The two runs were identical game for game (same winner and length
in all 24), and each recorded the same two `mode_pick` windows, Charming
Prince's "choose one", answered with its first bullet: its blink clause can
pick the bot's own creature and declares no direction, so it is not priced.
The labelled position suite and the gated heuristic games passed. Reports
are local under `data/arena-4674/` (gitignored).

## Entry-counter creature bodies (#4322, S75)

The heuristic prices a creature's prospective body using reliable fixed
and X own-entry counters projected as `purpose.entry_counters`. Shared
constructors supply the metadata; neither the view nor the policy executes
a replacement callback or parses names, oracle text or row labels. Casts
use the announced X. Fixed counts also inform off-board card/fuel value;
actual battlefield evaluation never adds entry counters again.

`PriceEntryCounters` is on by default and off in the frozen baseline.
`heuristic-noentrycounters` disables this price alone for future arena
comparisons. Only +1/+1 and -1/-1 counters adjust P/T; other counter kinds,
conditional or board/payment-dependent counts and external replacements
(including counter doublers) remain unpredicted. Krasis's separate half-X
cast payout is unchanged. Metadata remains hidden with a hidden card.

Focused tests compare Hydra with a vanilla 4/4 trampler, require Mikaeus
and Krasis prices to rise with X, and preserve the zero-X and disabled
prices. The real-engine regression gives the bot five Forests and Hydra
and Burnished Hart in hand: it chooses Hart with the price off, Hydra with
it on, and the Hydra enters with four counters and survives. Projection
tests ensure catalog callbacks are never executed while viewing a card.

The 2026-10-29 UTC comparison used develop `12ad35d5c` before and after
#4322, all four seats using the same policy, seeds 2000–2095, rotation and
lockstep. Each run played 96 pool2 games with the decks below. All 192
ended with a winner, without stalls or dropped decision records. The
labelled position suite retained its agreement.

| Card | Before games cast / offered | Fixed games cast / offered | Before offer windows | Fixed offer windows |
|---|---:|---:|---:|---:|
| Kalonian Hydra | 5/15 | 11/15 | 77 | 41 |
| Mikaeus, the Lunarch | 11/17 | 17/17 | 170 | 38 |
| Hydroid Krasis | 7/10 | 7/10 | 27 | 24 |

An offered game has at least one legal cast in a recorded window. Window
counts are repeated observations within games, not independent trials.
Kalonian Hydra is absent from the fixed run's `never` list; it was already
absent on this newer baseline, unlike the original report's older sample.
The controlled choice test establishes the scoring regression directly.

| Deck | Before wins / 96 (Wilson 95%) | Fixed wins / 96 (Wilson 95%) |
|---|---|---|
| boros-voltron | 42, 43.8% (34.3–53.7%) | 43, 44.8% (35.2–54.7%) |
| mono-white-weenie | 34, 35.4% (26.6–45.4%) | 34, 35.4% (26.6–45.4%) |
| mono-green-stompy | 12, 12.5% (7.3–20.6%) | 11, 11.5% (6.5–19.4%) |
| wubrg-dragons | 8, 8.3% (4.3–15.6%) | 8, 8.3% (4.3–15.6%) |

The four-seat null rate is 25%. Stompy lost one win and dragons held;
these intervals do not establish a strength gain or prove noninferiority.
The evidence supports the card-pricing behavior change. Reports, fixed
binaries and the streaming scanner remain local under `data/arena-4322/`
(gitignored). Full decision logs stay private.

## Own-target removal returns (#4319, S75)

A removal target's declared `returns` now prices an own permanent as a
loss plus the return, rather than a helpful pick. Such an own-target cast
also loses its generic mana-value proxy and still pays for the card spent.
An own commander pays its tax-and-tempo loss, capped at its board value.
Opposing-target prices are unchanged, and `NetRemoval=false` restores the
old behavior. Own-target pricing requires `PriceTargetPurposes`.

Token bodies receive credit when replacing an ordinary creature. They
receive no credit against a land, mana rock, planeswalker or commander:
the existing body weights alone would otherwise still prefer turning
Sol Ring into a 3/3. This conservative rule preserves those resources;
it does not model rescuing a doomed permanent or future death triggers.
A profitable ordinary creature conversion remains allowed, including a
1/1 mana creature. Position tests cover these choices and prefer removing
an opposing 6/6. An ungated whole-engine regression uses the real Generous
Gift, Sol Ring, enumerator and view, and reproduces the own-goal with
`NetRemoval` off.

The 2026-10-29 UTC evaluation used develop `fd4b401d3` plus #4319,
all four seats using the fixed heuristic, rotation and lockstep. Each pool
played 48 games: std seeds 1000–1047 and pool2 seeds 2000–2047, with the
decks listed below. All 96 ended with a winner, without stalls or dropped
decision records. The full bot package tests, labelled suite, real-engine
regression and `go vet` passed.

| Pool | Generous Gift casts | Beast Within casts | Own-target casts | Own targets |
|---|---:|---:|---:|---|
| std | 6 | 8 | 3 | 1/1 Human, Saproling and Soldier tokens |
| pool2 | 20 | 8 | 1 | 1/1 Human token |

No applied cast of either spell hit an own land, rock, planeswalker,
commander or larger creature. The four remaining own-target casts are
profitable token-body conversions, consistent with allowing the issue's
Elvish Mystic example. The issue's literal zero-own-target count was not
reached: four profitable conversions remain by design.

| Pool | Deck | Wins / 48 (Wilson 95%) |
|---|---|---|
| std | esper-control | 13, 27.1% (16.6–41.0%) |
| std | mono-black-aristocrats | 10, 20.8% (11.7–34.3%) |
| std | mono-red-aggro | 16, 33.3% (21.7–47.5%) |
| std | simic-ramp | 9, 18.8% (10.2–31.9%) |
| pool2 | boros-voltron | 22, 45.8% (32.6–59.7%) |
| pool2 | mono-green-stompy | 5, 10.4% (4.5–22.2%) |
| pool2 | mono-white-weenie | 18, 37.5% (25.2–51.6%) |
| pool2 | wubrg-dragons | 3, 6.2% (2.1–16.8%) |

The four-seat null rate is 25%. This is a behavior validation with wide
intervals and no paired back-out comparison; it does not establish a
strength gain. Reports and the streaming scanner are local under
`data/arena-4319/`; final runs are in `final/`. Full decision logs stay
private. Earlier provisional runs are excluded from these counts.

## Land-drop sequencing (#4317, S75)

`heuristic-nolandorder` is today's heuristic with `OrderLandDrops` off.
The fixed policy ranks a bounce land behind alternatives on an empty
board and favours a known untapped land over a known tapped one when
it unlocks an ordinary printed mana cost. Entry facts come from
structured catalog metadata, not names or oracle text. Unknown entries
and competing untapped lands alone do not justify a bonus. The real
engine regression plays Mountain before Path of Ancestry and casts
Firebrand Archer that turn.

The 2026-10-27 UTC comparison used develop `bb785912a` plus #4317,
the std pool (mono-red-aggro, simic-ramp, esper-control and
mono-black-aristocrats), seeds 1000–1047, rotation and lockstep.
All four seats used either `heuristic` or `heuristic-nolandorder`
(48 games each). All 96 games ended with a winner without stalls or
dropped decision records. The back-out control is reused from the
initial comparison: its disabled policy did not change when the
active rule was narrowed. The labelled suite remained 42/42.

The local scanner checks ordinary printed mana costs with coloured
repeatable sources. It excludes X and additional costs, and skips
windows with filter/restricted/conditional mana it cannot model. It is a
sequencing estimate, not proof of future legal casts under all targets,
taxes, discounts or entry replacements. Applied land drops and the
offered alternatives come from the seat's decision window.

| Measurement | Back-out | Fixed |
|---|---:|---:|
| Applied land drops | 1462 | 1437 |
| Empty-board bounce drops, all | 13 | 7 |
| Empty-board bounce with a non-bounce alternative | 6 | 0 |
| Tapped drops with a known untapped alternative | 87 | 68 |
| Ordinary mana windows audited | 80 | 57 |
| Tapped choice strands an ordinary mana cost | 30 | 2 |
| Complex mana windows skipped | 7 | 11 |

The seven remaining fixed empty-board bounces had no non-bounce land
alternative. Ordering keeps the only land playable; one-land bounce
hands remain a mulligan-policy issue. Mulligans were unchanged at 53.
Opening missed drops (own turns 2–4) fell overall from **59/576 to
58/576**, with no deck's count increasing. These changed games are not
independent per-decision trials.

| Deck | Back-out wins (Wilson 95%) | Fixed wins (Wilson 95%) | Back-out missed drops | Fixed missed drops |
|---|---|---|---:|---:|
| esper-control | 8/48, 16.7% (8.7–29.6%) | 10/48, 20.8% (11.7–34.3%) | 11/144 | 11/144 |
| mono-black-aristocrats | 8/48, 16.7% (8.7–29.6%) | 11/48, 22.9% (13.3–36.5%) | 23/144 | 22/144 |
| mono-red-aggro | 24/48, 50.0% (36.4–63.6%) | 17/48, 35.4% (23.4–49.6%) | 14/144 | 14/144 |
| simic-ramp | 8/48, 16.7% (8.7–29.6%) | 10/48, 20.8% (11.7–34.3%) | 11/144 | 11/144 |

The four-seat null is 25%. Aggro's point estimate fell when all seats
used the fix, prompting the mixed-policy comparison below. The wide
intervals do not establish a strength gain or rule out a deck-specific
regression.

### Mixed-table strength check

An additional 96 games used the same decks and seeds with two fixed
seats and two back-out seats. The first 48 used
`heuristic,heuristic-nolandorder,heuristic,heuristic-nolandorder`;
the second 48 swapped both policies across every deck. Rotation and
lockstep remained on, giving each policy 48 games on each deck.
All 96 ended with a winner without stalls.

| Deck | Back-out wins (Wilson 95%) | Fixed wins (Wilson 95%) |
|---|---|---|
| esper-control | 11/48, 22.9% (13.3–36.5%) | 9/48, 18.8% (10.2–31.9%) |
| mono-black-aristocrats | 9/48, 18.8% (10.2–31.9%) | 12/48, 25.0% (14.9–38.8%) |
| mono-red-aggro | 17/48, 35.4% (23.4–49.6%) | 19/48, 39.6% (27.0–53.7%) |
| simic-ramp | 11/48, 22.9% (13.3–36.5%) | 8/48, 16.7% (8.7–29.6%) |

Both policies won 48 tables, or 25.0% of seat-games (Wilson 95%
19.4–31.6%), against the 25% four-seat null. Aggro's mixed-table count
rose 17 → 19, while simic-ramp and esper-control fell. This supports
the targeted behavior fix without establishing a strength gain or
proving noninferiority for every deck.

The narrower scope is deliberate: a preliminary rule also reordered
unknown entries and competing untapped lands and lost 42–54 in mixed
tables. It is not shipped. Its reports remain under `data/arena-4317/`;
the final reports, scanner and compressed decision logs are under
`data/arena-4317/final/` (gitignored). Full decision logs remain private.

## Late attackers and optional blocks (#4316, S75)

`heuristic-nooutnumber` is today's heuristic with `PriceOutnumberLoss`
off. It restores the old free-damage price for a late attacker even when
a legal single blocker kills it without dying. The fixed policy retains
the usual best-block loss price in that case. Exact lethal pushes, race
and attrition plans still take their existing swings.

The 2026-10-26 comparison used develop `7f87ca32c` plus #4316, the
four pool2 decks, seeds 2000–2047, rotation and lockstep. All four seats
used either `heuristic` or `heuristic-nooutnumber` (48 games each).
All 96 games ended with a winner and none stalled. The labelled suite
retained 42/42 agreement without errors or timeouts.

Applied declarations of named, non-token attackers were joined to the
next matching public attack event. Deaths count battlefield-to-graveyard
moves in either combat damage step; a combat without a kill is one in
which no declared blocker died. Bounces, exile and removal before damage
are not counted as combat deaths. Public log sequences were deduplicated
across seat views. All declarations matched an attack event. Final
combats without a subsequent step in the decision logs are censored,
including lethal attacks: 181 back-out and 182 fixed declarations.
The table uses only the observed, completed combats; those censored
attacks are not silently counted as survivors.

| Attack reason | Back-out observed | Died | Without a kill | Fixed observed | Died | Without a kill |
|---|---:|---:|---:|---:|---:|---:|
| undefended | 295 | 0 | 0 | 318 | 0 | 0 |
| blockers can't profit | 188 | 14 | 6 | 205 | 16 | 7 |
| attrition | 240 | 22 | 1 | 222 | 21 | 0 |
| two-turn race | 280 | 50 | 9 | 281 | 53 | 9 |
| all-in for the kill | 285 | 31 | 31 | 300 | 30 | 26 |
| outnumbers their blockers | 101 | 48 | 34 | 50 | 10 | 0 |
| blocked at a loss | 44 | 18 | 2 | 45 | 20 | 3 |
| gang-blocked at a loss | 6 | 2 | 0 | 6 | 1 | 0 |
| lethal, unblockable | 1 | 0 | 0 | 1 | 0 | 0 |

For the targeted outnumber attacks, deaths fell from **48/101 (47.5%)
to 10/50 (20.0%)**, and deaths in a combat without a kill from **34 to
0**. Across all observed reasons, deaths without a kill fell from 83
to 45. These counts measure the behavior fix; changed attacks produce
different later positions, so they are not independent per-creature trials.

| Deck | Back-out wins / 48 (Wilson 95%) | Fixed wins / 48 (Wilson 95%) |
|---|---|---|
| boros-voltron | 9, 18.8% (10.2–31.9%) | 12, 25.0% (14.9–38.8%) |
| mono-white-weenie | 27, 56.2% (42.3–69.3%) | 26, 54.2% (40.3–67.4%) |
| mono-green-stompy | 6, 12.5% (5.9–24.7%) | 6, 12.5% (5.9–24.7%) |
| wubrg-dragons | 6, 12.5% (5.9–24.7%) | 4, 8.3% (3.3–19.6%) |

The four-seat null is 25%. The intervals overlap broadly; this does not
establish a strength gain or rule out a deck-specific regression. The
lower dragons win count prompted the additional mixed-table comparison
below. Full Cards reports, the scanner and compressed decision logs stay
local under `data/arena-4316/` (gitignored).

### Mixed-table strength check

An additional 96 games used the same pool2 decks and seeds, with two
fixed seats and two back-out seats at each table. The first 48 used
`heuristic,heuristic-nooutnumber,heuristic,heuristic-nooutnumber`;
the second 48 swapped both policies across the four decks. Rotation
and lockstep remained on, giving each policy 48 games on each deck.
All 96 games ended with a winner without stalls. No decision logs were
needed for this strength check; the behavioral audit above is separate.

| Deck | Back-out wins (Wilson 95%) | Fixed wins (Wilson 95%) |
|---|---|---|
| boros-voltron | 7/48, 14.6% (7.2–27.2%) | 13/48, 27.1% (16.6–41.0%) |
| mono-green-stompy | 7/48, 14.6% (7.2–27.2%) | 5/48, 10.4% (4.5–22.2%) |
| mono-white-weenie | 25/48, 52.1% (38.3–65.5%) | 29/48, 60.4% (46.3–73.0%) |
| wubrg-dragons | 5/48, 10.4% (4.5–22.2%) | 5/48, 10.4% (4.5–22.2%) |

The fixed seats won 52 tables and the back-out seats 44. Per seat,
that is 27.1% (Wilson 95% 21.3–33.8%) versus 22.9% (17.5–29.4%),
against the 25% four-seat null. Dragons' point estimate did not fall
in this mixed check, and voltron improved. These wide, overlapping
intervals still do not establish a strength gain or prove noninferiority.
The mixed results support shipping the behavior fix while retaining
the arena back-out for broader measurements.

## Defensive answers with nothing to answer (#4318, S75)

`heuristic-noidleanswers` is today's heuristic with `PriceAnswerOnlyRows`
off. It restores only the previous prices of answer-only defensive rows,
while retaining response saves and value-activation pricing. Compare it
with `heuristic` on the same seeds and rotated chairs; the pool2 decks
include Selfless Spirit and Ranger-Captain of Eos. Count their activations
with an empty stack from decision logs, and check the Cards table for
responses still taken. The whole-engine tests keep both creatures through
an idle postcombat main phase and sacrifice Selfless Spirit to save two
creatures from a real Wrath of God.

The 2026-10-26 comparison used develop `587c02cdf` plus #4318, the
four pool2 decks, seeds 2000–2047, rotation and lockstep. All four seats
used either `heuristic` or `heuristic-noidleanswers` (48 games each).
All 96 games ended with a winner and none stalled. Empty-stack
sacrifices fell from **14 to 0**: Selfless Spirit 8 → 0 and Ranger-Captain
6 → 0. Thirteen old sacrifices were in postcombat main and one was in
another empty-stack window. Activations with a nonempty stack fell from
4 to 0; those counts alone do not say whether an answer was useful.
The separate whole-engine Wrath regression confirms a useful response
still works. The labelled suite retained 42/42 agreement with no errors
or timeouts.

The Cards table's activation-use games, among games where offered:

| Deck / card | Back-out | Fixed |
|---|---:|---:|
| boros-voltron / Selfless Spirit | 6 / 7 | 0 / 7 |
| mono-white-weenie / Selfless Spirit | 5 / 6 | 0 / 6 |
| mono-white-weenie / Ranger-Captain of Eos | 7 / 9 | 0 / 9 |

| Deck | Back-out wins / 48 (Wilson 95%) | Fixed wins / 48 (Wilson 95%) |
|---|---|---|
| boros-voltron | 9, 18.8% (10.2–31.9%) | 10, 20.8% (11.7–34.3%) |
| mono-white-weenie | 27, 56.2% (42.3–69.3%) | 26, 54.2% (40.3–67.4%) |
| mono-green-stompy | 7, 14.6% (7.2–27.2%) | 6, 12.5% (5.9–24.7%) |
| wubrg-dragons | 5, 10.4% (4.5–22.2%) | 6, 12.5% (5.9–24.7%) |

The four-seat null is 25%. These wide intervals do not establish a
strength gain. Reports, the streaming counter and compressed decision
logs are local under `data/arena-4318/` (gitignored).

## Reconfigure comparison (#4321, S75)

`heuristic-noreconfigure` is today's heuristic with `PriceReconfigure`
turned off. Compare it with `heuristic` using the same seeds and rotated
chairs. The curated `boros-voltron` and `mono-white-weenie` decks include
Rabbit Battery and Lion Sash; synthetic battle decks do not exercise this
change. Decision logs expose the row's `reconfigure` flag and the source's
`attached_to` relation, so activations onto the current host can be counted
without guessing from labels. The position tests cover better and worse
hosts, first attachment and the back-out. The ungated whole-engine test
plays Rabbit Battery's turn through the real enumerator and policy.

The 2026-10-26 evaluation used base `01645dea7` plus #4321, the four
decks listed below, seeds 2000–2047, rotation and lockstep, with all four seats
using either `heuristic` or `heuristic-noreconfigure` (48 games each).
All 96 games ended with a winner and none stalled. Attachments onto the
current host fell from **27 to 0** (Rabbit Battery 21 → 0, Lion Sash
6 → 0); first attachments remained in use (8 → 10). Neither run chose
an unattach. The 42-position labelled suite retained 100% agreement,
with no errors, timeouts or rejected choices.

| Deck | Back-out wins / 48 (Wilson 95%) | Fixed wins / 48 (Wilson 95%) |
|---|---|---|
| boros-voltron | 9, 18.8% (10.2–31.9%) | 7, 14.6% (7.2–27.2%) |
| mono-white-weenie | 27, 56.2% (42.3–69.3%) | 28, 58.3% (44.3–71.2%) |
| mono-green-stompy | 7, 14.6% (7.2–27.2%) | 7, 14.6% (7.2–27.2%) |
| wubrg-dragons | 5, 10.4% (4.5–22.2%) | 6, 12.5% (5.9–24.7%) |

The four-seat null rate is 25%. These runs establish that the repeated
attachment stops; their wide intervals do not establish a strength gain.
Local reports, the streaming counter and decision logs are under
`data/arena-4321/` (gitignored). Full decision logs stay local.

## Lifegain payoffs on a lifelink grant (#4979, S75)

`PriceLifeGainPayoffs` adds Heliod's and Ajani's Pridemate's declared
`life_gain_payoff` to a lifelink grant on the bot's creature in combat
made with spare mana ([bot.md](bot.md#purposes-what-a-spell-or-an-ability-does)). There is no
arena variant: compare one tree built twice, with the switch on and off
in `DefaultConfig()`. The position tests
(`heuristic/lifegain_payoff_test.go`) hold the price (a counter over a
plain grant, two with a Pridemate out, nothing for an opponent's payoff
or before attacks) and one decision: Heliod's lifelink beats a +1/+0 on
the same attacker only with the payoff. The whole-engine tests in
`aiseat/heliod_grant_bot_test.go` still hold the grant for combat and for
the second main phase's cast, and make it once.

The 2026-11-05 runs used develop `f01023882` plus #4979, rotation and
lockstep, every seat `heuristic`: 48 four-seat games on
mono-white-weenie, boros-voltron, mono-green-stompy and simic-ramp, and
96 two-seat games of the mono-white-weenie mirror, both at `--seed 4979`.
Every game ended on the same turn with the same winner with the switch on
and off, and Heliod's activation was taken in the same windows (3 of 140,
and 17 of 582): on these decks the change reprices the grant and changes
no decision. The mirror's one stall is the same game in both builds.

## Triggered mana in the turn plan (#4770, S75)

The turn plan's mana model read each source's `produced`, which CR 106.7
keeps triggered mana out of, so a Swamp under Crypt Ghast counted as one
{B} and the bot under-planned turns with two spells. The view now names
that mana on each mana row as `trigger_adds` ([protocol.md](protocol.md)),
from the auto-tap planner's own judge, and `PlanTriggerMana` counts it
([bot.md](bot.md#the-mana-model)). It **ships off**: the arena below
found no gain. `heuristic-triggermana` seats today's heuristic with it
on. The view tests (`game/trigger_mana_view_test.go`,
`protocol/trigger_mana_view_test.go`) and the model tests
(`heuristic/plan_trigger_mana_internal_test.go`) hold both sides: with
the switch on, two Swamps under Crypt Ghast and a Forest pay
{B}{B}{B}{B}{G}, and with it off they make three mana.

The 2026-11-05 UTC runs used develop `ee606a09b` plus #4770 and the
synthetic `--decks ghast-battle` at all four seats (six Crypt Ghasts,
four Zendikar Resurgents, BattleDeck's vanilla creatures recoloured
black and green, twelve Swamps and twelve Forests), two seats with the
switch on against two without, rotation and lockstep:

```
boteval arena --seats heuristic-triggermana,heuristic-triggermana,heuristic,heuristic \
  --decks ghast-battle,ghast-battle,ghast-battle,ghast-battle \
  --games 48 --seed 2000 --rotate --lockstep --turn-budget 60 --wall 10m --out DIR --md
```

and the same with `--seed 3000`. Every game ended with a winner and none
stalled. The four-seat null rate is 25%.

| Seeds | Switch on: wins / 96 | Switch off: wins / 96 |
|---|---|---|
| 2000–2047 | 19, 19.8% (13.1–28.9%) | 29, 30.2% (21.9–40.0%) |
| 3000–3047 | 26, 27.1% (19.2–36.7%) | 22, 22.9% (15.6–32.3%) |
| all | 45 / 192, 23.4% (18.0–29.9%) | 51 / 192, 26.6% (20.8–33.2%) |

Two seats of one policy share a game, so these trials are not
independent and the intervals are conservative. The two batches point
opposite ways and the intervals overlap.

The arena's Turn mana section shows what the switch changes:

| Policy | Own turns | Planned windows | Plan misses | Mean unspent | Decision p99 |
|---|---|---|---|---|---|
| switch on | 2,057 | 460 | 31 of 446 checked (7.0%) | 1.35 | 3.35 ms, 2.39 ms |
| switch off | 2,094 | 233 | 6 of 224 checked (2.7%) | 1.23 | 2.95 ms, 2.95 ms |

With the switch on, the bot plans two or more spells about twice as
often, and no turn in either policy ended with stranded mana and a cast
on offer. But its plan misses rose to 7.0%, above ADR 0136 §8's 5% bar
(P6). Decision logs of four games with the switch on at all four seats
suggest one likely cause: the bot paid Crypt Ghast's extort {W/B} 89 times,
and the plan does not budget for extort, so mana it planned for the
second spell can go to the first spell's extort. Pricing extort into
the plan, or a deck without it, is the run to make before turning the
switch on. The view field works as specified: lands under Crypt Ghast and
Zendikar Resurgent carried `{B}`, `{G}`, `{B}{B}` and `{B}{B}{B}`, which
the plans counted. The labelled position suite, the gated heuristic games
(`AISEAT_GAME_TESTS=1`, the four- and two-seat heuristic tests) and the
package tests passed. With the switch off, `DefaultConfig` plays exactly
as before. Reports and decision logs stayed on the authoring machine.

## Test knobs

- **The whole-game tests are gated off by default** and the package owns the longest tests in the tree (CI runs `go test` with a 30m timeout because of them; `server-test` is sharded across five runners by `scripts/ci-test-shard.sh`, the same flags on each, and `aiseat` is in the `rest` shard, #2766):
  - `AISEAT_GAME_TESTS=1` — the master gate. Without it every whole-game test in `internal/aiseat` skips. The nightly `bot-games` job runs `go test ./internal/aiseat/... -race -timeout 30m -skip 'TestFourRandomBotsPlayToAWinner|TestFourHeuristicBotsPlayToAWinner|TestTwoHeuristicBotsPlayToAWinner'`. The catalog soak is its own job, `catalog-soak` (#1456; it needs `CMDCTRL_SCRYFALL_DUMP`, fetched through the day-keyed Scryfall cache; like every test job it runs on a GitHub-hosted runner, because the self-hosted ones hold the deploy credentials, #3335). The skipped tests run without `-race` in their own jobs, side by side (#2661): the random-table run is `bot-random`, four shards of five games, and the heuristic gates are `bot-soak-4p` and `bot-soak-2p` (`.github/workflows/e2e-nightly.yml`).
  - `AISEAT_HEURISTIC_GAMES=N` / `AISEAT_H2H_GAMES=N` — widen the four-heuristic and heuristic-vs-random samples (defaults 3 and small, for CI). The nightly `bot-soak-4p` and `bot-soak-2p` jobs set `AISEAT_HEURISTIC_GAMES=20`, which is S31 exit criterion 2 (#685); the seeds are fixed at 101..120, so each plays the same twenty tables every night. Neither sets `AISEAT_H2H_GAMES`. Both tests play **lockstep** since #1409 (`playLockstepGame` in `heuristic_game_test.go`: the seats' ordinary act-loop called in seat order on one goroutine, through `aiseat.NewStepped` / `Runner.Step` since #1503), so a seed replays move for move — and `boteval arena --lockstep` plays the same schedule; the soak keeps the production one-goroutine-per-seat schedule, because concurrent liveness is what it tests (discussion #1390).
  - `AISEAT_HEURISTIC_SCHEDULE=concurrent` — plays the heuristic gate the old way, one runner goroutine per seat. For comparing the two schedules; the nightly does not set it.
  - `AISEAT_RANDOM_GAMES=N` / `AISEAT_RANDOM_SEED=<uint64>` — `TestFourRandomBotsPlayToAWinner`, S31's four-`random`-bots test: N consecutive games to a winner, default 3, from base seed 31000. The nightly runs it without `-race` as the `bot-random` job: four shards at `AISEAT_RANDOM_GAMES=5` with `AISEAT_RANDOM_SEED` 31000, 31005, 31010 and 31015, which are the twenty seeds 31000..31019 a single run of twenty plays (#2661).
  - `AISEAT_REPLAY_DIR=<path>` — where that test writes each game's replay JSONL. It sets the location only, not whether replays are kept. Unset uses a temp dir. Either way a passing game's replay is deleted and a failing one kept, since one game is hundreds of MiB. The nightly points it at the workspace and uploads it on failure.
  - `AISEAT_FUNNEL_GAMES=N` — widen the Layer A absorption / model-funnel whole-game run.
  - `AISEAT_SOAK_GAMES=N` — independent of the master gate; runs `TestRandomBotSoak` for N games. `AISEAT_SOAK_POLICY=random|heuristic|mixed` picks what fills the seats (default `random`), and `AISEAT_SOAK_SEED=<uint64>` pins the base seed (default: the clock). Game *i* uses `base+i`, and the run logs the base seed and the replay command before the first game, so a green artifact still says which hundred tables were played. A stall prints its own seed too. The nightly `bot-soak-4p` job sets `AISEAT_SOAK_GAMES=100` — S31 exit criterion 4 (#685) — with `AISEAT_SOAK_SEED` derived from the UTC date times 1000, so consecutive nights fuzz disjoint tables rather than overlapping in 99 of 100 games; `bot-soak-2p` plays 100 two-seat games from that seed plus 500. Both take their seeds and the commit they test from the `bot-soak-seed` job.
  - `CMDCTRL_PRINTED_CACHE_CHECKS=0` — set it when **profiling or benchmarking** a test binary (the 3-game soak at `AISEAT_SOAK_SEED=20260923000`, `GOMAXPROCS=2` is the usual recipe). In every test binary each hit of `game`'s printed-characteristic cache is rebuilt and compared ([#1498](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1498), `game/printed_cache.go`), which is the whole suite doubling as the cache's invalidation test — and exactly the cost the cache saves, so a profile taken with the check on shows it (`checkPrintedEntry`). It cannot turn the check on in production.
  - `AISEAT_DEBUG=1` — per-move log in the runner tests.
  - `AISEAT_LIVE_MODEL_TESTS=1` — **the only test in the tree that talks to a model over a network** (`aiseat/model/improvise_live_test.go`, #686): it sends one real improvisation call and checks the bundle against the rail the runner would put it through. Needs a transport too (`CMDCTRL_ANTHROPIC_API_KEY`, or `CMDCTRL_OPENAI_ENDPOINT` plus `CMDCTRL_BOT_MODEL`) and skips without one. **No workflow sets it and none should**: CI has no key, and a test whose result depends on a model's judgement is not a gate. Everything else in that package runs against `model.FakeClient`.
  - `AISEAT_DECISION_LOG=<dir>` — writes a per-game decision log for every game played through `playGame`/`playGameIn` (the same writer the server uses, `aiseat/decisionlog`). Unset (the default, and CI) costs nothing: the runner builds no event without an observer. `AISEAT_DECISION_LOG_MODE` takes the same three values as the server variable. This is how the position corpus gets harvested, and what `TestDecisionLogReplaysOffline` uses to prove a recorded window re-decides identically offline. **Mind the disk**: a four-seat game writes a few thousand windows and a four-player board view is ~50 KiB, so one game is 100–250 MiB. `escalated` only saves anything for policies that run Layer A (the `heuristic` TIER, `rules.NewFilter(heuristic.New(), …)`); the bare `heuristic.New()` the whole-game tests seat reports every window as Layer B and keeps them all in full.
  - `AISEAT_CATALOG_GAMES=N` / `AISEAT_CATALOG_SEED=<uint64>` — the catalog soak (`TestCatalogSoak`, #601): N four-bot games on decks dealt from the catalog itself rather than from the hand-written vanilla decks the other whole-game tests use. The seed defaults to one derived from the UTC date, so each night deals new decks and coverage accumulates; the log prints the base seed, and every failure prints the seed to replay. Needs `CMDCTRL_SCRYFALL_DUMP` as well (a `Spec` carries an oracle ID and a name, not a type line or a mana cost) and skips without it. It **fails on any `EventEffectError`** — a card whose primitive threw mid-resolution, which the engine logs and survives, so nothing else in the tree goes red over it.
  - `AISEAT_CATALOG_POLICY=random|heuristic|mixed` (#1456) — what fills the catalog soak's seats, modelled on `AISEAT_SOAK_POLICY` (the random soak, below): `random` (default, unchanged from before this variable existed) is the widest exploration of the catalog; `heuristic` seats `heuristic.New()` at every seat, which builds board states — blockers up, equipment attached, auras out — that `random` rarely sets up and most catalog cards actually need to be reached at all; `mixed` alternates by seat index. An unrecognised value fails the test outright rather than falling back silently, because a catalog soak run is expensive enough that a misspelled flag deserves a loud failure. The nightly's `catalog-soak` job sets `heuristic`.
  - `AISEAT_CATALOG_REPORT=<path>` — where the catalog soak writes its report JSON: `{"cards": [...], "census": {...}}`. `cards` is the per-card table (cast / resolved / entered / triggered / errored, per oracle ID) it always wrote. `census` is ADR 0041 P7's measurement (#1497, #1558 item 4) — the catalog soak captures a restore point (`game.Game.CaptureSnapshot`, the same call `ws.Room.writeRestorePointLocked` makes) after every observed action of every game in the run, and tallies what it found: `actions` and `captured` (how many of those captures were restorable), `by_kind` (blocked captures per census kind, summed over the whole run — `game.ContinuationCensus.Kinds`'s keys), and `longest_run_by_kind` (the longest run of CONSECUTIVE actions any one kind stayed the blocker — the number the shutdown census can't give, since it only ever reads one instant). It is not a gate: the shutdown census (#524) reads one instant per deploy and every deploy so far has caught an idle table, so this is the stand-in evidence P7 asks for to order phase 3's remaining tiers. Capture is sampled at the poll rate the stall detector already uses, so a burst of actions inside one poll tick coalesces into one capture — an undercount that can only make a stale run look shorter than it was, never longer. The nightly uploads the whole file as an artifact on every run, green included: the useful half of `cards` is the list of cards no bot game reached, and the useful half of `census` is whether tier 4 (stack effects and target specs, by a wide margin in every run so far) is worth hurrying.
  - `AISEAT_STALL` / `AISEAT_WALLCLOCK` — Go durations. Since [#685](https://github.com/krakenhavoc/cmd_and_ctrl/issues/685) **every bot-table budget in the package reads them**, so there is one pair of knobs for a loaded runner and no literal left to edit:
    - `TestFourRandomBotsPlay` — stall detector and wall clock, defaults `15s` / `300s`.
    - `TestCatalogSoak` — the same, same defaults.
    - `TestRandomBotSoak` — per-game stall detector and wall clock, defaults `3s` / `60s`.
    - `playGame` / `playGameIn`, which is every other whole-game test (four-heuristic, random-to-a-winner, head-to-head, latency, funnel, model, life-cost, decision-log replay) — stall detector, default `5s`.
    - `TestManagerPlaysALobbySeatedTable` — `AISEAT_WALLCLOCK` as its wait for the table to finish, default `300s`. Its wait for the runners to EXIT is no longer a budget at all: that is `waitForRunner` now (#1048).

    **`AISEAT_WALLCLOCK` is a floor on the `playGame` / `playGameIn` wall clock, not a replacement.** Each caller passes its own — 120s for a 50-turn heuristic table, 25 turns for the decision-log replay — and those numbers say something about the game being played, so the variable may raise them and never cuts them. Everywhere else it is the value. Unset (the default, and PR CI) changes nothing anywhere.

    Raise them on a loaded runner rather than editing a test. The nightly `bot-soak-4p` and `bot-soak-2p` jobs set `AISEAT_STALL=30s` and `AISEAT_WALLCLOCK=300s`, and both are `workflow_dispatch` inputs.
  - `CMDCTRL_SCRYFALL_DUMP=<path>` — gates the manual bot-deck test that validates the four curated decks against the real Scryfall dump. Also used by other packages.
