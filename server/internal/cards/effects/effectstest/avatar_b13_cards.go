package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avB13Cast puts a card with its printed cost and body in the active
// player's hand, walks to a main phase and casts it.
func AvB13Cast(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = active.ID, active.ID
	active.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return c.InstanceID, g.CastSpell(active.ID, c.InstanceID, params)
}

func AvB13Restrictions(g *game.Game, id uuid.UUID) game.Restriction {
	var r game.Restriction
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				r = c.Effective().Restrictions
			}
		}
	})
	return r
}
