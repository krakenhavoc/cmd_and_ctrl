package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// erGraveyardCard puts a card in p's graveyard.
func ErGraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

// erReturn reanimates a graveyard card through the engine's own path.
func ErReturn(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(id, game.ZoneBattlefield); err != nil {
			t.Errorf("return: %v", err)
		}
	})
}

func ErOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := BattlefieldCard(g, id)
	return ok
}
