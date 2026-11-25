package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nahiri's Stoneblades — Instant {1}{R} (War of the Spark, #5520):
//
//	"Up to two target creatures each get +2/+0 until end of turn."
//
// Each target still legal as the spell resolves gets the pump (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "769774f9-1d8d-406f-8b0c-2d6713e7ab3e",
		Name:         "Nahiri's Stoneblades",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Nahiri's Stoneblades — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
