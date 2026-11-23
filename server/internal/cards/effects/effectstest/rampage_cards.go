package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rampageBlock has seat 0's `attacker` attack seat 1, blocked by each of
// `blockers`, locks the blocks in and resolves what triggered, stopping
// in the declare blockers step.
func RampageBlock(t *testing.T, g *game.Game, attacker uuid.UUID, blockers ...uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	AdvanceTo(t, g, game.StepDeclareBlockers)
	// One declaration, so Gorilla Berserkers' three-blocker minimum is
	// judged on the whole set.
	decls := make([]game.BlockDeclaration, 0, len(blockers))
	for _, b := range blockers {
		decls = append(decls, game.BlockDeclaration{Blocker: b, Attacker: attacker})
	}
	if err := g.DeclareBlockers(decls); err != nil {
		t.Fatal(err)
	}
	FinishBlockDeclarations(t, g)
	PassPriorityAroundTable(t, g)
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("combat left the declare blockers step (now %v)", g.Turn.Step)
	}
}

// rampageBlockers is n 1/1s with reach for seat 1, so they can block
// Chromium too.
func RampageBlockers(g *game.Game, n int) []uuid.UUID {
	var out []uuid.UUID
	for range n {
		c := ApaCreature("Archer", "", 1, 1)
		c.Keywords = []string{"reach"}
		out = append(out, ApaPush(g, g.Seats[1].ID, g.Seats[1].ID, c))
	}
	return out
}
