package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tandem Takedown — Instant {1}{G}{G} (March of the Machine):
//
//	"Up to two target creatures you control each get +1/+0 until end of
//	 turn. They each deal damage equal to their power to another target
//	 creature, planeswalker, or battle."
//
// Allies at Last's shape. The boost lands first, so each creature's power
// includes it, and the damage is one simultaneous instance from each
// creature that is still legal. With no creature chosen the spell does
// nothing, and the victim must differ from the chosen creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8b0edb8-6a69-4760-9bcd-969c6c898a16",
		Name:         "Tandem Takedown",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2),
			effects.Distinct(effects.S82MomB12CreaturePlaneswalkerOrBattle("another target creature, planeswalker, or battle")),
		),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var sources []game.TargetRef
			for _, t := range ctx.ClauseTargets(0) {
				if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
					sources = append(sources, t)
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: "Tandem Takedown — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			return effects.EachDealsDamageEqualToItsPower(ctx, sources, victim.ID)
		},
	})
}
