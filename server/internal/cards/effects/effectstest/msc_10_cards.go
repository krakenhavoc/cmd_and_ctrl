package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc10Cast puts the named spell in `who`'s hand, floats `mana` and casts
// it under the strict mana gate, so a cast that costs more than the pool
// holds is refused.
func Msc10Cast(t *testing.T, g *game.Game, who *game.Player, name, typeLine, cost, oracle, mana string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	id := HandCardFull(who, name, typeLine, cost, oracle, nil)
	if mana != "" {
		if err := g.AddManaForEffect(who.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	params.Strict = true
	return id, g.CastSpell(who.ID, id, params)
}

// msc10CastCommander casts a free commander creature for `me`.
func Msc10CastCommander(t *testing.T, g *game.Game, me *game.Player, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2,
		IsCommander: true, Owner: me.ID, Controller: me.ID,
	})
	AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
