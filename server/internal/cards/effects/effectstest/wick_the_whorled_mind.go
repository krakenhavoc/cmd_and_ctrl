package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wickCastRat casts a vanilla Rat from hand and settles the stack.
func WickCastRat(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rat", TypeLine: "Creature — Rat", ManaCost: "{B}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	FloatForTest(g, me, "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	PassPriorityAroundTable(t, g)
	PassPriorityAroundTable(t, g)
}
