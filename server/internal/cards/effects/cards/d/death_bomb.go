package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death Bomb — Instant {3}{B} (Planeshift, S82 OLD b644, #5615):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Destroy target nonblack creature. It can't be regenerated. Its controller
//	 loses 2 life."
//
// The sacrifice is paid with the spell on the stack (CR 601.2h). The destruction
// ignores regeneration shields (CR 701.19c); the life loss is the target's controller
// as it last was, read before the destruction, and happens whether or not the creature
// was destroyed (an indestructible one still costs its controller 2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "e323d468-def5-4090-af5b-19481289168f",
		Name:           "Death Bomb",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			card, found := ctx.Game.LookupCardForEffect(id)
			if !found {
				return nil
			}
			if err := (effects.DestroyTarget{Target: id, CantBeRegenerated: true}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), card.Controller, -2)
		},
	})
}
