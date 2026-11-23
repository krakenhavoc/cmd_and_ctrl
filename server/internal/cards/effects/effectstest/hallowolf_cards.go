package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func HwHasKeyword(c game.Card, kw string) bool {
	for _, a := range c.Effective().Abilities {
		if a == kw {
			return true
		}
	}
	return false
}

// hwDeclare declares the creatures as one attack declaration (CR 508.1),
// so a trigger that reads the whole declaration sees all of it.
func HwDeclare(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	var decls []game.AttackDeclaration
	for _, id := range attackers {
		decls = append(decls, game.AttackDeclaration{Attacker: id, Target: defender})
	}
	if _, err := g.DeclareAttackers(decls); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
}

func HwCount(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			n++
		}
	}
	return n
}
