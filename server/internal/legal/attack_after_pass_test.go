package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// attack_after_pass_test.go — #2462. The active seat's pass in declare
// attackers ends its declaration (#1571's checkpoint). The enumerator
// used to keep offering it attacks while the other seats passed, and a
// bot holding no priority took one as its only legal move: accepted if
// it landed before the step moved on, refused as "action not legal in
// current step" if the last seat passed first.

// TestNoAttacksAfterTheActiveSeatPasses walks one declare-attackers
// step: attacks are offered while the active seat holds priority, and
// not once it has passed, at any seat's turn to hold priority.
func TestNoAttacksAfterTheActiveSeatPasses(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	sent := battlefieldCard(g, active, creature("Sent", "{G}", 2, 2))
	battlefieldCard(g, active, creature("Kept Home", "{G}", 2, 2))
	advanceTo(t, g, game.StepDeclareAttackers)

	if !holdsPriority(g, active.ID) {
		t.Fatalf("the active seat should hold priority as declare attackers begins")
	}
	moves := legal.EnumerateFor(g, active.ID)
	if countKind(moves, legal.KindAttack) == 0 {
		t.Fatalf("the active seat holding priority should be offered attacks: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)

	// Declare one attacker, keep the other home, and pass.
	if err := g.DeclareAttacker(sent, def.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}

	// Every other seat now holds priority in turn. At each of those
	// moments the active seat has no attack to take — its list is empty,
	// so a bot runner woken by a commit has nothing to be forced into.
	for passes := 0; g.Turn.Step == game.StepDeclareAttackers; passes++ {
		if passes >= len(g.Seats) {
			t.Fatalf("declare attackers never ended after %d passes", passes)
		}
		if holdsPriority(g, active.ID) {
			t.Fatalf("priority came back to the active seat before the step ended")
		}
		if moves := legal.EnumerateFor(g, active.ID); len(moves) != 0 {
			t.Fatalf("the active seat has passed; want no moves while seat %d holds priority, got %v",
				g.Turn.PriorityHolder, labels(moves))
		}
		holder := g.Seats[g.Turn.PriorityHolder]
		hm := legal.EnumerateFor(g, holder.ID)
		if countKind(hm, legal.KindAttack) != 0 {
			t.Fatalf("a non-active seat was offered an attack: %v", labels(hm))
		}
		if !hasLabel(hm, "Pass priority") {
			t.Fatalf("the seat holding priority should be offered a pass: %v", labels(hm))
		}
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
}
