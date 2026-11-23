package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Second Breakfast — Instant {2}{W} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Up to two target creatures each get +2/+1 until end of turn. Create a
//	 Food token."
//
// Each still-legal target gets the pump as the spell resolves. The Food is made
// whenever the spell resolves, including with no targets chosen (it is "up to
// two"); it is not made when every target chosen has become illegal and the
// spell is countered on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf5f68b8-8ae2-4d08-b86c-4e6d3b653fec",
		Name:         "Second Breakfast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		Purpose:      game.Purpose{Tokens: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 2, 1))},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 1, Label: "Second Breakfast — +2/+1"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.CreateToken{Template: effects.FoodToken(), N: 1},
		),
	})
}
