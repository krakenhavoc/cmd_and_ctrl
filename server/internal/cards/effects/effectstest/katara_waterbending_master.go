package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func KataraCastInstant(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: "Impulse", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	PassPriorityAroundTable(t, g)
}
