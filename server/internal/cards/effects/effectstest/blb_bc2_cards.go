package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blbBC2Tokens makes n 1/1 Rabbit tokens for owner and returns their IDs.
func BlbBC2Tokens(t *testing.T, g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	t.Helper()
	before := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		before[c.InstanceID] = true
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(owner, effects.TokenCard("1/1 white Rabbit"), n) })
	var ids []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if !before[c.InstanceID] && effects.IsToken(c) && c.Controller == owner {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) != n {
		t.Fatalf("made %d tokens, want %d", len(ids), n)
	}
	return ids
}
