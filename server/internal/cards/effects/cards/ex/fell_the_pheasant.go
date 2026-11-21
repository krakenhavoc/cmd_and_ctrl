package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fell the Pheasant — Instant {1}{G} (Throne of Eldraine, #5516):
//
//	"Fell the Pheasant deals 5 damage to target creature with flying. Create a Food token."
//
// The damage, then the Food as a second step (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c78c8f7d-45fd-46c0-9ed4-4d0e85aada2e",
		Name:         "Fell the Pheasant",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SourceDealsDamageToEachLegalTarget(5)(ctx.Game, ctx.Item)
			}),
			effects.CreateToken{Template: effects.FoodToken(), N: 1},
		),
	})
}
