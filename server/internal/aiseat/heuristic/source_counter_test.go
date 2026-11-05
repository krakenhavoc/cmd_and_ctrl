package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// source_counter_test.go is #5083: Hangarback Walker's "{1}, {T}: Put a
// +1/+1 counter on this creature." declares the counter it puts on its
// own source, and with PriceSourceCounters on the bot grows the Walker
// in its own main phase with spare mana, where the flat price (0.50 less
// the 0.30 blocker) left it just under PassThreshold. An attack worth
// more than the counter, a blocker the coming combat needs, and a spell
// the mana is wanted for still come first. The switch ships off (the
// price lost in the arena, docs/bot-testing.md), so the default policy
// passes on its own turn and grows the Walker in the end step before its
// untap, as it did before.

const growWalker = "Hangarback Walker: put a +1/+1 counter on it"

// hangarback is Hangarback Walker with n +1/+1 counters and its tap row
// as the view projects it.
func hangarback(id string, controller, n int, opts ...cardOpt) protocol.CardView {
	c := creature(id, controller, "Hangarback Walker", n, n, opts...)
	c.TypeLine = "Artifact Creature — Construct"
	c.ManaCost = "{X}{X}"
	c.Counters = map[string]int{"+1/+1": n}
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Label: "{1}, {T}: Put a +1/+1 counter on this creature.", ManaCost: "{1}", TapCost: true,
		Purpose: &protocol.PurposeView{Answers: &[]string{"pump"}, SourceCounters: 1},
	}}
	return c
}

// walkerBoard is the bot's Walker with two counters and two Islands, the
// opponent's `theirs`, in `step` of seat `active`'s turn, with the
// Walker's activation and a pass on offer, plus `extra` moves.
func walkerBoard(t *testing.T, active int, step string, theirs []protocol.CardView, seats []protocol.PlayerView, extra ...legal.Move) aiseat.Input {
	t.Helper()
	cards := []protocol.CardView{hangarback(cardID(1), 0, 2), island(cardID(90), 0), island(cardID(91), 0)}
	cards = append(cards, theirs...)
	if seats == nil {
		seats = []protocol.PlayerView{newSeat(0), newSeat(1)}
	}
	v := newView(seats, withBattlefield(cards...), withTurn(6, active, step))
	moves := append([]legal.Move{passMove(0), activateMove(t, 0, cardID(1), growWalker, nil)}, extra...)
	return input(0, v, moves...)
}

// sourceCounters is today's heuristic with the counter priced, as the
// arena's `heuristic-sourcecounters` seats it.
func sourceCounters() *heuristic.Policy {
	return withConfig(func(c *heuristic.Config) { c.PriceSourceCounters = true })
}

// The issue's position: the bot's own second main phase, two spare
// Islands, nothing in hand. The counter (1.45) less the blocker (0.30)
// clears PassThreshold; with the price off (the default) the row is the
// flat 0.20 and the bot passes, as it did on its own turn in #4991's
// arena.
func TestHangarbackGrowsInItsOwnMainPhase(t *testing.T) {
	for _, step := range []string{"postcombat_main", "precombat_main"} {
		t.Run(step, func(t *testing.T) {
			// Before combat a 2/2 into the opponent's 4/4 has no attack
			// to give up, so it grows there too.
			in := walkerBoard(t, 0, step, []protocol.CardView{creature(cardID(20), 1, "Wall", 4, 4)}, nil)
			if got := chose(t, in, decide(t, sourceCounters(), in)); got != growWalker {
				t.Errorf("with the counter priced the heuristic chose %q, want the Walker grown with the spare mana", got)
			}
			if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
				t.Errorf("the default heuristic chose %q, want the pass it makes with the switch off", got)
			}
			if got := chose(t, in, decide(t, baseline(), in)); got != "Pass priority" {
				t.Errorf("the baseline chose %q, want a pass", got)
			}
		})
	}
}

// The end step before the bot's untap, where the tap costs nothing, is
// still a window it grows in, as it was before (the 83 activations of
// #4991's arena).
func TestHangarbackStillGrowsBeforeItsUntap(t *testing.T) {
	in := walkerBoard(t, 1, "end", nil, nil)
	for name, p := range map[string]*heuristic.Policy{"default": heuristic.New(), "priced": sourceCounters()} {
		if got := chose(t, in, decide(t, p, in)); got != growWalker {
			t.Errorf("%s: the heuristic chose %q in the end step before its untap, want the Walker grown", name, got)
		}
	}
}

// Before combat on its own turn, an attack worth more than the counter
// is kept: the opponent is at 2 life with no blockers, so the 2/2 swings
// for the game.
func TestHangarbackAttacksRatherThanGrowsWhenTheAttackIsWorthMore(t *testing.T) {
	seats := []protocol.PlayerView{newSeat(0), newSeat(1, withLife(2))}
	in := walkerBoard(t, 0, "precombat_main", nil, seats)
	if got := chose(t, in, decide(t, sourceCounters(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q before combat, want the Walker kept untapped for a lethal attack", got)
	}
}

// On an opponent's turn, with their creature coming at the bot, the
// Walker is a blocker: the counter is not priced before blocks, and the
// bot does not tap it. One counter (1.45) is under InstantThreshold
// anyway; a row that declared two (2.90) would not be, which is what the
// window rule is for, so both are held here.
func TestHangarbackBlocksRatherThanGrows(t *testing.T) {
	for _, n := range []int{1, 2} {
		for _, step := range []string{"precombat_main", "begin_combat", "declare_attackers"} {
			t.Run(step, func(t *testing.T) {
				var opts []cardOpt
				if step == "declare_attackers" {
					opts = append(opts, attacking(0))
				}
				theirs := []protocol.CardView{creature(cardID(20), 1, "Ogre", 3, 3, opts...)}
				in := walkerBoard(t, 1, step, theirs, []protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2)})
				in.View.Battlefield.Cards[0].ActivatedAbilities[0].Purpose.SourceCounters = n
				if got := chose(t, in, decide(t, sourceCounters(), in)); got != "Pass priority" {
					t.Errorf("%d counters: the heuristic chose %q on the opponent's turn before blocks, want the Walker kept to block", n, got)
				}
			})
		}
	}
}

// With a spell in hand that wants the same mana, the spell comes first:
// a 3/3 flier for the bot's three Islands beats a counter for one.
func TestHangarbackGrowsOnlyWithSpareMana(t *testing.T) {
	drake := creature(cardID(30), 0, "Drake", 3, 3, keywords("flying"))
	drake.ManaCost = "{2}{U}"
	drake.TypeLine = "Creature — Drake"
	seats := []protocol.PlayerView{newSeat(0, withHand(drake)), newSeat(1)}
	cards := []protocol.CardView{hangarback(cardID(1), 0, 2), island(cardID(90), 0), island(cardID(91), 0), island(cardID(92), 0)}
	v := newView(seats, withBattlefield(cards...), withTurn(6, 0, "postcombat_main"))
	in := input(0, v, passMove(0), activateMove(t, 0, cardID(1), growWalker, nil), castMove(t, 0, cardID(30), "Cast Drake"))
	if got := chose(t, in, decide(t, sourceCounters(), in)); got != "Cast Drake" {
		t.Errorf("the heuristic chose %q, want the Drake cast with the mana", got)
	}
}

// The switch ships off, in the default and in the baseline.
func TestPriceSourceCountersDefaults(t *testing.T) {
	if heuristic.DefaultConfig().PriceSourceCounters {
		t.Error("DefaultConfig: PriceSourceCounters is on")
	}
	if heuristic.BaselineConfig().PriceSourceCounters {
		t.Error("BaselineConfig: PriceSourceCounters is on")
	}
}
