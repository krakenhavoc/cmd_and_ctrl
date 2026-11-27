package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Poison the Well — Sorcery {2}{B/R}{B/R} (Shadowmoor, #5591):
//
//	"Destroy target land. Poison the Well deals 2 damage to that land's
//	 controller."
//
// The land's controller is read before it is destroyed (CR 608.2h), and the
// damage does not depend on the destruction: an indestructible land's
// controller is still dealt 2. The damage runs once the destruction has
// settled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a47ae84f-dcbc-4a60-9266-c1a02f92e7f8",
		Name:         "Poison the Well",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{id}, effects.WithCarry(ctx.Tail(poisonTheWellDamage), c.Controller))
		},
	})
}

// poisonTheWellDamage is the second sentence: 2 damage to the land's last
// controller. An on-disk key: never renamed, never reused.
var poisonTheWellDamage = effects.CardTail("card/poison-the-well-damage", func(ctx *effects.Context, o game.TailOutcome) error {
	player := effects.TailCarry0(o)
	if player == uuid.Nil {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: player, Amount: 2}.Apply(ctx)
})
