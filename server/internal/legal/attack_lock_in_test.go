package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// TestNoAttacksOnceTheDeclarationIsLockedIn — #3621. The declaration
// locks in at the first state-check after it is staged, and the active
// seat still holds priority then (its attack triggers are on the
// stack). The enumerator used to keep offering it attacks for every
// untapped creature, and the engine would have accepted them late.
func TestNoAttacksOnceTheDeclarationIsLockedIn(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	sent := battlefieldCard(g, active, creature("Sent", "{G}", 2, 2))
	battlefieldCard(g, active, creature("Kept Home", "{G}", 2, 2))
	advanceTo(t, g, game.StepDeclareAttackers)

	if err := g.DeclareAttacker(sent, def.ID); err != nil {
		t.Fatal(err)
	}
	// Staged, not locked in: the sandbox's click-by-click declaration
	// still gets the rest of the board offered.
	if countKind(legal.EnumerateFor(g, active.ID), legal.KindAttack) == 0 {
		t.Fatalf("a staged declaration should still be offered attacks")
	}

	g.RunStateChecksForTest()

	if !holdsPriority(g, active.ID) {
		t.Fatalf("the active seat should still hold priority after the lock-in")
	}
	if moves := legal.EnumerateFor(g, active.ID); countKind(moves, legal.KindAttack) != 0 {
		t.Fatalf("a locked-in declaration is over (CR 508.1); got %v", labels(moves))
	}
}
