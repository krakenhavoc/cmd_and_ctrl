package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- helpers --------------------------------------------------------------------

// avB16Cast puts `c` into the active seat's hand, walks to a main phase
// and casts it.
func AvB16Cast(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, params)
}

// avB16CastCreature casts a creature card and resolves it, leaving its
// enters trigger (if any) pending.
func AvB16CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id, err := AvB16Cast(t, g, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle, Power: power, Toughness: toughness}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}
