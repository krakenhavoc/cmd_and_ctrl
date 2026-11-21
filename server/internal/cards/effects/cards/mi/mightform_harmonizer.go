package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mightform Harmonizer — Creature — Insect Druid {2}{G}{G}, 4/4 (Edge
// of Eternities):
//
//	"Landfall — Whenever a land you control enters, double the power of
//	 target creature you control until end of turn.
//	 Warp {2}{G}"
//
// Landfall targets a creature you control (the Harmonizer itself is a
// legal target). Doubling reads the power as the trigger resolves, and a
// negative power doubles downward (CR 701.10b-c, DoublePowerUntilEOT).
// Warp is the engine's keyword (CR 702.185).
//
// No simplification.
func init() {
	landfall := effects.Landfall("Mightform Harmonizer — double the power of target creature you control until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.DoublePowerUntilEOT(ctx, t.ID, "Mightform Harmonizer — double power")
			}
			return nil
		})
	landfall.Targets = effects.TargetCreature("target creature you control", effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:         "8e025cda-71c5-4d06-8436-decd408667af",
		Name:             "Mightform Harmonizer",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{2}{G}")},
		Triggered:        []game.TriggeredAbility{landfall},
	})
}
