package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Molten Rain — Sorcery {1}{R}{R} (Mirrodin, S82 OLD, #5606):
//
//	"Destroy target land. If that land was nonbasic, Molten Rain deals 2
//	 damage to the land's controller."
//
// Stone Rain with Price of Progress's rider. The land's controller and
// whether it was basic are read before it is destroyed (a land that is
// indestructible or regenerates still counts as the nonbasic land that was
// targeted), and the damage is dealt to its controller, not its owner. A
// basic land costs its controller nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9774695-b2f1-4014-8504-04b2339be568",
		Name:         "Molten Rain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			controller, nonbasic := uuid.Nil, false
			if c, found := ctx.Game.LookupCardForEffect(id); found {
				controller, nonbasic = c.Controller, !c.HasSupertype("Basic")
			}
			if !nonbasic {
				return effects.DestroyTarget{Target: id}.Apply(ctx)
			}
			// The damage is the spell's next sentence: a keyed tail after the
			// destruction, dealt whether or not the land died.
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{id},
				effects.WithCarry(ctx.Tail(moltenRainDamage), controller))
		},
	})
}

// moltenRainDamage is "Molten Rain deals 2 damage to the land's controller".
// An on-disk key: never renamed, never reused.
var moltenRainDamage = effects.CardTail("then/s82-old-b603-molten-rain-damage", func(ctx *effects.Context, o game.TailOutcome) error {
	controller := effects.TailCarry0(o)
	if controller == uuid.Nil {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: 2}.Apply(ctx)
})
