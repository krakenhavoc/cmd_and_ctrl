package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flesh Allergy — Sorcery {2}{B}{B} (Scars of Mirrodin, S82 OLD batch 471,
// #5577):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Destroy target creature. Its controller loses life equal to the number of
//	 creatures that died this turn."
//
// Rite of Consumption's sacrifice cost. The target's controller is read before
// it is destroyed; the life is counted after, from the turn's tally of
// creatures that died (any player's, tokens included), so the sacrificed
// creature and the destroyed one both count, and one that regenerated or was
// indestructible does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "bd3ce6a2-9efc-4dab-a357-9f2aa0a4c0c6",
		Name:           "Flesh Allergy",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetCreature("target creature"),
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			controller, ok := effects.ControllerOfTarget(ctx, id)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: controller, Amount: -effects.B11CreaturesDiedThisTurn(ctx.Game)}.Apply(ctx)
		},
	})
}
