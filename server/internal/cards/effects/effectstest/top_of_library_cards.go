package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func PushTopSource(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		OracleID: oracle, Owner: owner, Controller: owner,
	})
}

func RemoveFromBattlefield(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		cards := g.Battlefield.Cards[:0]
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				cards = append(cards, c)
			}
		}
		g.Battlefield.Cards = cards
	})
}

func TopVisibleTo(g *game.Game, owner, viewer uuid.UUID) bool {
	var ok bool
	g.ReadSnapshot(func() { ok = g.LibraryTopVisibleToLocked(owner, viewer) })
	return ok
}
