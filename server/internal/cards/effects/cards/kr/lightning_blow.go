package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Blow — Instant {1}{W} (Ice Age, S82 OLD, #5636):
//
//	"Target creature gains first strike until end of turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike"))
	purpose.Draws = 1
	effects.Register(effects.Spec{
		OracleID:     "6b7bbdb7-bdd5-4eb1-912e-f2f810772ac8",
		Name:         "Lightning Blow",
		Completeness: effects.CompletenessFull,
		Purpose:      purpose,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.S82OldB726TargetGainsKeywordUntilEOT("Lightning Blow", "first strike")(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Lightning Blow — draw a card").Apply(ctx)
		},
	})
}
