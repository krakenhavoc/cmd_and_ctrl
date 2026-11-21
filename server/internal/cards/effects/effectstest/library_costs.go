package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lcRecord is what a test ability's effect saw on its payment record.
type LcRecord struct {
	Ran       bool
	Discarded []uuid.UUID
	Exiled    []uuid.UUID
}

func LcSpec(oracle, name string, cost game.AbilityCost, rec *LcRecord) effects.Spec {
	return effects.Spec{
		OracleID: oracle, Name: name,
		Activated: []effects.ActivatedAbility{{
			Label: name + ": Note the payment.",
			Cost:  cost,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				rec.Ran = true
				rec.Discarded = ctx.Discarded()
				rec.Exiled = ctx.Exiled()
				return nil
			},
		}},
	}
}

// lcLibrary replaces `p`'s library with n named cards, bottom first, and
// returns their IDs top first.
func LcLibrary(g *game.Game, p *game.Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	g.WithWriteLock(func() {
		p.Library.Cards = nil
		for i := 0; i < n; i++ {
			id := uuid.New()
			p.Library.PushTop(game.Card{InstanceID: id, Name: "Library Card", TypeLine: "Land", Owner: p.ID, Controller: p.ID})
			ids[n-1-i] = id
		}
	})
	return ids
}
