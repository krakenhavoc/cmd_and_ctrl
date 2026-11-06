package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// resolution_mana_test.go is #4980: Soulbright Seeker's "{R}: Target
// creature you control gains trample until end of turn. If this is the
// third time this ability has resolved this turn, add {R}{R}{R}{R}."
// declares the mana and which resolution adds it, and the view stamps
// the row with the engine's count of its resolutions this turn. With
// PriceResolutionMana on (the default), an activation on a reachable
// run to the third resolution, in the bot's own main phase, is priced by
// the extra mana the hand can then spend. Where the mana does not come
// back (the run cannot be paid, or an activation already on the stack
// is the third) or cannot be used (nothing more to cast, another
// player's turn), the row keeps the trample grant's price, which on a
// creature that already has trample is below a pass.

const (
	seekerLabel = "{R}: Target creature you control gains trample until end of turn; the third time it resolves this turn, add {R}{R}{R}{R}"
	runSeeker   = "Soulbright Seeker: trample"
)

// seeker is Soulbright Seeker with its row as the view projects it,
// having resolved `resolved` times this turn. It has trample once the
// row has resolved.
func seeker(id string, controller, resolved int) protocol.CardView {
	var opts []cardOpt
	if resolved > 0 {
		opts = append(opts, keywords("trample"))
	}
	c := creature(id, controller, "Soulbright Seeker", 2, 1, opts...)
	c.TypeLine = "Creature — Elemental Sorcerer"
	c.ManaCost = "{R}"
	entries := []protocol.TargetPurposeView{{Slot: 0, UntilEndOfTurn: "benefit",
		Pump: &protocol.PumpView{Keywords: []string{"trample"}}}}
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Label: seekerLabel, ManaCost: "{R}", ResolvedThisTurn: resolved,
		Purpose: &protocol.PurposeView{
			Targets:        &entries,
			ResolutionMana: &protocol.ResolutionManaView{Resolution: 3, Mana: "{R}{R}{R}{R}"},
		},
	}}
	return c
}

// seekerItem is an activation of the Seeker's row on the stack.
func seekerItem(n int, src string) protocol.StackItemView {
	return protocol.StackItemView{ID: "seeker-item-" + cardID(n), Kind: "activated", Controller: seatID(0).String(),
		Owner: seatID(0).String(), SourceCardID: src, Label: seekerLabel,
		Targets: []protocol.TargetRefView{{Kind: "card", ID: src}}}
}

// fourDrop is a {3}{R} 4/4 in hand, the cast the Seeker's mana pays for.
func fourDrop(id string) protocol.CardView {
	c := creature(id, 0, "Ogre", 4, 4)
	c.ManaCost = "{3}{R}"
	return c
}

// seekerBoard is the bot's Seeker, resolved `resolved` times, with
// `mountains` untapped Mountains and `hand`, in `step` of seat `active`'s
// turn, facing an opponent's 4/4 the Seeker's trample does not get past.
// A pass and the Seeker's activation aimed at itself are on offer.
func seekerBoard(t *testing.T, resolved, mountains, active int, step string, hand []protocol.CardView, items ...protocol.StackItemView) aiseat.Input {
	t.Helper()
	cards := []protocol.CardView{seeker(cardID(1), 0, resolved), creature(cardID(20), 1, "Wall", 4, 4)}
	for i := 0; i < mountains; i++ {
		cards = append(cards, land(cardID(90+i), 0))
	}
	seats := []protocol.PlayerView{newSeat(0, withHand(hand...)), newSeat(1)}
	v := newView(seats, withBattlefield(cards...), withTurn(6, active, step), withStackItems(items...))
	return input(0, v, passMove(0), activateMove(t, 0, cardID(1), runSeeker, nil, cardTarget(cardID(1))))
}

// noResolutionMana is today's heuristic with the price off, as the
// arena's `heuristic-noresolutionmana` seats it.
func noResolutionMana() *heuristic.Policy {
	return withConfig(func(c *heuristic.Config) { c.PriceResolutionMana = false })
}

// The issue's position: the row has resolved twice, one Mountain is
// left and the hand holds a four-drop. The third activation costs {R}
// and returns {R}{R}{R}{R}, which casts the Ogre. With the price off it
// is a third trample grant on a creature that has it, below a pass.
func TestSeekersThirdActivationIsPricedByTheManaItReturns(t *testing.T) {
	in := seekerBoard(t, 2, 1, 0, "precombat_main", []protocol.CardView{fourDrop(cardID(30))})
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != runSeeker {
		t.Errorf("the default heuristic chose %q, want the third activation for the four-drop's mana", got)
	}
	if got := chose(t, in, decide(t, noResolutionMana(), in)); got != "Pass priority" {
		t.Errorf("with the price off the heuristic chose %q, want a pass (the grant repeats)", got)
	}
}

// The run to the third resolution is priced from wherever it stands
// when the rest of it can be paid: from none with three Mountains (three
// {R} for four), from one with two, in either main phase.
func TestSeekerRunsToItsThirdResolution(t *testing.T) {
	for _, step := range []string{"precombat_main", "postcombat_main"} {
		for _, c := range []struct{ resolved, mountains int }{{0, 3}, {1, 2}, {2, 1}} {
			in := seekerBoard(t, c.resolved, c.mountains, 0, step, []protocol.CardView{fourDrop(cardID(30))})
			if got := chose(t, in, decide(t, heuristic.New(), in)); got != runSeeker {
				t.Errorf("%s, resolved %d with %d Mountains: chose %q, want the Seeker activated", step, c.resolved, c.mountains, got)
			}
		}
	}
}

// Only where the mana comes back: with too few Mountains for the rest of
// the run, nothing is returned, and the row keeps the grant's price.
func TestSeekerIsNotRunWhenTheRunCannotBePaid(t *testing.T) {
	for _, c := range []struct{ resolved, mountains int }{{0, 2}, {1, 1}} {
		in := seekerBoard(t, c.resolved, c.mountains, 0, "precombat_main", []protocol.CardView{fourDrop(cardID(30))})
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
			t.Errorf("resolved %d with %d Mountains: chose %q, want a pass (the run cannot reach the third resolution)", c.resolved, c.mountains, got)
		}
	}
}

// Only where the mana can be used: with nothing in hand, or only a card
// the returned mana still cannot pay for, the four red mana would empty
// unspent at the end of the step.
func TestSeekerIsNotRunForManaTheHandCannotSpend(t *testing.T) {
	five := creature(cardID(31), 0, "Giant", 5, 5)
	five.ManaCost = "{4}{R}"
	for name, hand := range map[string][]protocol.CardView{"empty hand": nil, "a five-drop": {five}} {
		in := seekerBoard(t, 2, 1, 0, "precombat_main", hand)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
			t.Errorf("%s: chose %q, want a pass", name, got)
		}
	}
}

// The count is resolutions, not activations: an activation still on the
// stack has not resolved. Resolved once with one on the stack, the one
// on offer is the third and is priced (in response, against the instant
// bar). Resolved twice with one on the stack, that one is the third, and
// another adds nothing.
func TestSeekerCountsResolutionsNotActivations(t *testing.T) {
	hand := []protocol.CardView{fourDrop(cardID(30))}
	in := seekerBoard(t, 1, 1, 0, "precombat_main", hand, seekerItem(1, cardID(1)))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != runSeeker {
		t.Errorf("resolved once, one on the stack: chose %q, want the activation that is the third resolution", got)
	}
	in = seekerBoard(t, 2, 1, 0, "precombat_main", hand, seekerItem(1, cardID(1)))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("resolved twice, one on the stack: chose %q, want a pass (the one on the stack is the third)", got)
	}
}

// Another player's turn is not a window for it: the mana empties at the
// end of the step, and nothing the bot casts at sorcery speed can use it.
func TestSeekerIsNotRunOnAnotherPlayersTurn(t *testing.T) {
	for _, step := range []string{"precombat_main", "end"} {
		in := seekerBoard(t, 2, 1, 1, step, []protocol.CardView{fourDrop(cardID(30))})
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
			t.Errorf("%s of the opponent's turn: chose %q, want a pass", step, got)
		}
	}
}

// The run makes mana rather than spending what the casts leave, so it is
// not held behind a cast as a mana sink is: with a two-drop castable now
// and the four-drop castable only after the run, the run comes first,
// and three Mountains cast the four-drop rather than the two-drop.
func TestSeekerRunIsNotAManaSink(t *testing.T) {
	two := creature(cardID(31), 0, "Bear", 2, 2)
	two.ManaCost = "{1}{R}"
	in := seekerBoard(t, 0, 3, 0, "precombat_main", []protocol.CardView{fourDrop(cardID(30)), two})
	in.Moves = append(in.Moves, castMove(t, 0, cardID(31), "Cast Bear"))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != runSeeker {
		t.Errorf("chose %q, want the Seeker run toward the four-drop", got)
	}
}

// The switch ships on in the default and off in the baseline.
func TestPriceResolutionManaDefaults(t *testing.T) {
	if !heuristic.DefaultConfig().PriceResolutionMana {
		t.Error("DefaultConfig: PriceResolutionMana is off")
	}
	if heuristic.BaselineConfig().PriceResolutionMana {
		t.Error("BaselineConfig: PriceResolutionMana is on")
	}
}
