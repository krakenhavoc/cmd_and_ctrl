package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ddCombatDamage records `n` combat damage from `src` to `victim` the
// way the combat damage step emits it.
func DdCombatDamage(g *game.Game, actor, src, victim uuid.UUID, n int) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: actor, Source: src, Target: victim, Amount: n, Combat: true})
	})
}

func DdFlicker(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	var back uuid.UUID
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatal(err)
		}
		var err error
		if back, err = g.ReturnFromExileToBattlefieldForEffect(id, uuid.Nil, false); err != nil {
			t.Fatal(err)
		}
	})
	return back
}
