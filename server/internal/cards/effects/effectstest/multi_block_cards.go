package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mbCapacity is game.BlockCapacity of a battlefield card, layers fresh.
func MbCapacity(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c := E2Card(t, g, id)
	return game.BlockCapacity(&c)
}

// mbBlocked is the attackers `id` blocks.
func MbBlocked(t *testing.T, g *game.Game, id uuid.UUID) []uuid.UUID {
	t.Helper()
	c := E2Card(t, g, id)
	return c.BlockedAttackers()
}

// mbAttack has `me` attack `opp` with n fresh 2/2s and moves to
// declare blockers.
func MbAttack(t *testing.T, g *game.Game, me, opp *game.Player, n int) []uuid.UUID {
	t.Helper()
	var atks []uuid.UUID
	for i := 0; i < n; i++ {
		atks = append(atks, PushVanillaCreature(g, me.ID, "Bear", 2, 2))
	}
	DeclareAttack(t, g, opp.ID, atks...)
	AdvanceTo(t, g, game.StepDeclareBlockers)
	return atks
}

// mbBlock declares `blocker` on every attacker in one declaration.
func MbBlock(g *game.Game, blocker uuid.UUID, atks ...uuid.UUID) error {
	decls := make([]game.BlockDeclaration, len(atks))
	for i, a := range atks {
		decls[i] = game.BlockDeclaration{Blocker: blocker, Attacker: a}
	}
	return g.DeclareBlockers(decls)
}
