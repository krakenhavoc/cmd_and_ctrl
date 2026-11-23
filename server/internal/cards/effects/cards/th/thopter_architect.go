package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thopter Architect — Creature — Human Artificer {3}{W}, 2/3 (The Brothers'
// War, #5482):
//
//	"Whenever an artifact you control enters, target creature gains flying
//	 until end of turn."
//
// Any creature is a legal target, not only yours; it is chosen as the
// trigger goes on the stack and checked again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3194fe3d-57d3-4dab-ab39-3778b75150fa",
		Name:         "Thopter Architect",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.On(game.EventETB, effects.ArtifactEnteredUnderYourControl,
					"Thopter Architect — target creature gains flying until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Thopter Architect"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying"))),
		},
	})
}
