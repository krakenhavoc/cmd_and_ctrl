package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shelob's Ambush — Instant {B} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Target creature gets +1/+2 and gains deathtouch until end of turn.
//	 Create a Food token."
//
// The pump and the keyword are one effect on the target as it resolves
// (CR 611.2c). The Food is made whatever happens to the target only if the
// spell resolves at all: with its one target illegal the spell is countered on
// resolution (CR 608.2b) and no Food is made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2816923-9daf-4277-adf7-20427d14ba4a",
		Name:         "Shelob's Ambush",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Tokens: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 2, "deathtouch"))},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 2, Label: "Shelob's Ambush — +1/+2"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"deathtouch"}, Label: "Shelob's Ambush — deathtouch"}.Apply(ctx)
			}),
			effects.CreateToken{Template: effects.FoodToken(), N: 1},
		),
	})
}
