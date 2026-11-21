package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func MscWlBlocker(g *game.Game, name string, p, tough int) uuid.UUID {
	return ApaPush(g, g.Seats[1].ID, g.Seats[1].ID, game.Card{
		Name: name, TypeLine: "Creature — Test", Power: p, Toughness: tough,
	})
}

// mscWlBlockedCombat attacks seat 1 with the attackers, has each blocker
// block its attacker, and walks to the combat damage step.
func MscWlBlockedCombat(t *testing.T, g *game.Game, attackers []uuid.UUID, blocks map[uuid.UUID]uuid.UUID) {
	t.Helper()
	Pr7bAttack(t, g, g.Seats[1].ID, attackers...)
	Pr7bBlock(t, g, blocks)
	Pr7bDamageStep(t, g)
}

func MscWlDamagePrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	return Pr7bPrompt(t, g, game.PendingChoiceDamageAssignment)
}

func MscWlAnswer(t *testing.T, g *game.Game, c *game.PendingChoice, toPlayer int, shares ...game.DamageAssignmentEntry) error {
	t.Helper()
	return g.ResolveDamageAssignment(c.ID, g.Seats[0].ID, shares, toPlayer)
}
