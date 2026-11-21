package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// socB4Spell puts an instant or sorcery with a real mana cost into the
// active seat's hand, walks to a main phase and casts it.
func SocB4Spell(t *testing.T, g *game.Game, name, typeLine, cost string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
	AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}
