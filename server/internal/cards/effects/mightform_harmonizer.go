package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	landfall := Landfall("Mightform Harmonizer — double the power of target creature you control until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return DoublePowerUntilEOT(ctx, t.ID, "Mightform Harmonizer — double power")
			}
			return nil
		})
	landfall.Targets = TargetCreature("target creature you control", YouControl())
	Register(Spec{
		OracleID:         "8e025cda-71c5-4d06-8436-decd408667af",
		Name:             "Mightform Harmonizer",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{2}{G}")},
		Triggered:        []game.TriggeredAbility{landfall},
	})
}
