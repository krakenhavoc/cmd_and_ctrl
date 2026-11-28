package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infuse — Instant {2}{U} (Ice Age, S82 OLD, #5636):
//
//	"Untap target artifact, creature, or land.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6035303e-22b3-4b64-8160-1f46ae3dad55",
		Name:         "Infuse",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
		Purpose:      game.Purpose{Answers: game.AnswerValue, Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if id := effects.FirstLegalBattlefieldTarget(ctx); id != uuid.Nil {
				if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawACardAtTheNextUpkeep("Infuse — draw a card").Apply(ctx)
		},
	})
}
