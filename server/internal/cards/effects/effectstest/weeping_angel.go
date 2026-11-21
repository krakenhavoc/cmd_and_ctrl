package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// weepingAngelOpponentCastsCreature has `opp` cast a vanilla creature
// with flash (so it may at any time), and lets the stack settle.
func WeepingAngelOpponentCastsCreature(t *testing.T, g *game.Game, opp *game.Player) {
	t.Helper()
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Flash Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Keywords: []string{"flash"}, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts a creature: %v", err)
	}
	PassPriorityAroundTable(t, g)
}
