package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// assertOpponentsLost checks each opponent lost exactly `n` since `before`.
func AssertOpponentsLost(t *testing.T, g *game.Game, me uuid.UUID, before map[uuid.UUID]int, n int) {
	t.Helper()
	for id, life := range OpponentLives(g, me) {
		if got := before[id] - life; got != n {
			t.Errorf("opponent %s lost %d, want %d", id, got, n)
		}
	}
}

func ChandraTorchGrant(me *game.Player, card uuid.UUID) *game.CastPermission {
	for i := range me.CastPermissions {
		p := &me.CastPermissions[i]
		for _, ref := range p.Cards {
			if ref.ID == card {
				return p
			}
		}
	}
	return nil
}
