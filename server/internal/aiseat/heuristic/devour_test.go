package heuristic_test

import (
	"fmt"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// devour_test.go — #2419: a devour prompt is a "may" with any-number
// bounds, and the bot eats the creatures worth less than the counters
// they buy.

const devourChoiceID = "choice-devour"

func token(id string, power, tough int) protocol.CardView {
	c := creature(id, 0, "Goblin Token", power, tough)
	c.IsToken = true
	return c
}

// devourSubsets is the enumerator's answer set for a devour over
// `fodder`: every subset, the empty one included, each labelled by the
// ids it names.
func devourSubsets(t *testing.T, fodder []protocol.CardView) []legal.Move {
	t.Helper()
	var moves []legal.Move
	for mask := 0; mask < 1<<len(fodder); mask++ {
		ids := []string{}
		for i := range fodder {
			if mask&(1<<i) != 0 {
				ids = append(ids, fodder[i].InstanceID)
			}
		}
		moves = append(moves, choiceMove(t, 0, devourChoiceID, fmt.Sprintf("devour %v", ids), map[string]any{"card_ids": ids}))
	}
	return moves
}

func devourLabel(cards ...protocol.CardView) string {
	ids := []string{}
	for _, c := range cards {
		ids = append(ids, c.InstanceID)
	}
	return fmt.Sprintf("devour %v", ids)
}

func devourDecision(t *testing.T, n, draw, life int, fodder ...protocol.CardView) string {
	t.Helper()
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(fodder...),
		withChoice(protocol.PendingChoiceView{
			ID:         devourChoiceID,
			Kind:       "entry_sacrifice",
			Chooser:    seatID(0).String(),
			Reason:     "devour",
			Options:    fodder,
			ChooseMin:  0,
			ChooseMax:  len(fodder),
			Devour:     n,
			DevourDraw: draw,
			DevourLife: life,
		}))
	in := input(0, v, devourSubsets(t, fodder)...)
	return chose(t, in, decide(t, heuristic.New(), in))
}

func TestDevourEatsTheTokens(t *testing.T) {
	a, b := token(cardID(1), 1, 1), token(cardID(2), 1, 1)
	if got, want := devourDecision(t, 3, 0, 0, a, b), devourLabel(a, b); got != want {
		t.Fatalf("chose %q, want %q: three counters each beat a 1/1 token", got, want)
	}
}

func TestDevourKeepsAValuableCreature(t *testing.T) {
	tok := token(cardID(1), 1, 1)
	big := creature(cardID(2), 0, "Craw Wurm", 6, 4)
	if got, want := devourDecision(t, 3, 0, 0, tok, big), devourLabel(tok); got != want {
		t.Fatalf("chose %q, want %q: eat the token, keep the 6/4", got, want)
	}
	if got, want := devourDecision(t, 1, 0, 0, big), devourLabel(); got != want {
		t.Fatalf("chose %q, want %q", got, want)
	}
}

func TestDevourNeverEatsTheCommander(t *testing.T) {
	cmd := creature(cardID(1), 0, "Commander", 1, 1, commander())
	if got, want := devourDecision(t, 3, 0, 0, cmd), devourLabel(); got != want {
		t.Fatalf("chose %q, want to keep the commander", got)
	}
}

func TestDevourEatsMoreWhenTheCountersAreWorthMore(t *testing.T) {
	bear := creature(cardID(1), 0, "Grizzly Bears", 2, 2)
	if got, want := devourDecision(t, 1, 0, 0, bear), devourLabel(); got != want {
		t.Fatalf("devour 1: chose %q, a 2/2 is not worth one counter", got)
	}
	if got, want := devourDecision(t, 3, 0, 0, bear), devourLabel(bear); got != want {
		t.Fatalf("devour 3: chose %q, want %q", got, want)
	}
}

func TestDevourCountsTheCreaturesOwnReader(t *testing.T) {
	// A nontoken 1/1: eaten for Skullmulcher's card (or Marrow
	// Chomper's life) per creature, not for its single counter.
	elf := creature(cardID(1), 0, "Llanowar Elves", 1, 1)
	if got, want := devourDecision(t, 1, 0, 0, elf), devourLabel(); got != want {
		t.Fatalf("without a reader: chose %q, want to keep the Elf", got)
	}
	if got, want := devourDecision(t, 1, 1, 0, elf), devourLabel(elf); got != want {
		t.Fatalf("Skullmulcher: chose %q, want %q", got, want)
	}
	if got, want := devourDecision(t, 2, 0, 2, elf), devourLabel(elf); got != want {
		t.Fatalf("Marrow Chomper: chose %q, want %q", got, want)
	}
}

// The fixed-count sacrifice lands share the kind and keep their price.
func TestEntrySacrificeWithoutDevourStillPicksTheCheapest(t *testing.T) {
	a, b := land(cardID(1), 0), creature(cardID(2), 0, "Craw Wurm", 6, 4)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(a, b),
		withChoice(protocol.PendingChoiceView{
			ID: devourChoiceID, Kind: "entry_sacrifice", Chooser: seatID(0).String(),
			Options: []protocol.CardView{a, b}, ChooseMin: 1, ChooseMax: 1,
		}))
	in := input(0, v,
		choiceMove(t, 0, devourChoiceID, "sac land", map[string]any{"card_ids": []string{a.InstanceID}}),
		choiceMove(t, 0, devourChoiceID, "sac wurm", map[string]any{"card_ids": []string{b.InstanceID}}),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "sac land" {
		t.Fatalf("chose %q", got)
	}
}
