package heuristic_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #4316: the first, safe attacker does not commit the defender's block.
// Walk both declarations: the flier goes, then the Bear stays home. The
// back-out reproduces the old attack into a block that kills only the Bear.
func TestOutnumberDoesNotReserveBlocksForEarlierAttackers(t *testing.T) {
	for _, blocks := range []int{1, 2} {
		t.Run(fmt.Sprintf("%d blockers", blocks), func(t *testing.T) {
			cards := []protocol.CardView{
				creature(cardID(10), 0, "Flier", 4, 4, keywords("flying")),
				creature(cardID(11), 0, "Bear", 2, 2),
			}
			if blocks == 2 {
				cards = append(cards, creature(cardID(12), 0, "Other Bear", 2, 2))
			}
			for i := range blocks {
				cards = append(cards, creature(cardID(20+i), 1, "Giant", 4, 4))
			}
			v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(cards...), withTurn(8, 0, "declare_attackers"))
			moves := []legal.Move{passMove(0), attackMove(t, 0, cardID(10), 1), attackMove(t, 0, cardID(11), 1)}
			if blocks == 2 {
				moves = append(moves, attackMove(t, 0, cardID(12), 1))
			}
			p := heuristic.New()
			in := input(0, v, moves...)
			d := decide(t, p, in)
			if got := chose(t, in, d); got != moves[1].Label {
				t.Fatalf("first declaration = %s (%s), want the flier", got, d.Reason)
			}
			v.Battlefield.Cards[0].Tapped = true
			v.Battlefield.Cards[0].AttackingTarget = v.Seats[1].ID
			moves = append(moves[:1], moves[2:]...)
			in = input(0, v, moves...)
			if ok, got := attacksIn(t, p, in); ok {
				t.Fatalf("late attacker should stay home: %s", got)
			}
			if blocks == 1 {
				off := withConfig(func(c *heuristic.Config) { c.PriceOutnumberLoss = false })
				if ok, got := attacksIn(t, off, in); !ok || !strings.Contains(got, "outnumbers") {
					t.Fatalf("back-out should send the Bear: %s", got)
				}
			}
		})
	}
}

// Raw declared power used to call the Bear "lethal, unblockable" at
// six life even though the Giant blocks it and only the flier's four
// damage connects. At four life the exact lethal push still goes.
func TestOutnumberLossDoesNotUseRawPowerForLethal(t *testing.T) {
	for _, life := range []int{6, 4} {
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1, withLife(life))},
			withBattlefield(
				creature(cardID(10), 0, "Flier", 4, 4, keywords("flying"), tapped(), attacking(1)),
				creature(cardID(11), 0, "Bear", 2, 2),
				creature(cardID(20), 1, "Giant", 4, 4),
			), withTurn(8, 0, "declare_attackers"))
		in := input(0, v, passMove(0), attackMove(t, 0, cardID(11), 1))
		ok, got := attacksIn(t, heuristic.New(), in)
		if ok != (life == 4) {
			t.Fatalf("defender life %d: attack = %v, %s", life, ok, got)
		}
		if life == 4 && !strings.Contains(got, "all-in for the kill") {
			t.Fatalf("exact lethal push should override the loss: %s", got)
		}
	}
}

func TestOutnumberStillSendsASurvivingAttacker(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			creature(cardID(10), 0, "Flier", 4, 4, keywords("flying"), tapped(), attacking(1)),
			creature(cardID(11), 0, "Giant", 5, 5),
			creature(cardID(20), 1, "Bear", 2, 2),
		), withTurn(8, 0, "declare_attackers"))
	in := input(0, v, passMove(0), attackMove(t, 0, cardID(11), 1))
	if ok, got := attacksIn(t, heuristic.New(), in); !ok {
		t.Fatalf("a surviving attacker still goes: %s", got)
	}
	if heuristic.BaselineConfig().PriceOutnumberLoss {
		t.Fatal("baseline must preserve the old outnumber price")
	}
}

func TestTwoDeclaredAttackersDoNotSpendTwoBlockers(t *testing.T) {
	// Keep the defender outside the attrition horizon: this isolates
	// the per-attacker price from a committed longer-term swing.
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1, withLife(100))},
		withBattlefield(
			creature(cardID(10), 0, "Flier", 4, 4, keywords("flying"), tapped(), attacking(1)),
			creature(cardID(12), 0, "Other flier", 4, 4, keywords("flying"), tapped(), attacking(1)),
			creature(cardID(11), 0, "Bear", 2, 2),
			creature(cardID(20), 1, "Giant", 4, 4),
			creature(cardID(21), 1, "Other Giant", 4, 4),
		), withTurn(8, 0, "declare_attackers"))
	in := input(0, v, passMove(0), attackMove(t, 0, cardID(11), 1))
	if ok, got := attacksIn(t, heuristic.New(), in); ok {
		t.Fatalf("either Giant still kills the Bear: %s", got)
	}
	off := withConfig(func(c *heuristic.Config) { c.PriceOutnumberLoss = false })
	if ok, got := attacksIn(t, off, in); !ok || !strings.Contains(got, "outnumbers") {
		t.Fatalf("back-out should send the third attacker: %s", got)
	}
}
