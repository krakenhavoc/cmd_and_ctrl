package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emergent Growth — Sorcery, {3}{G}:
//
//	"Target creature gets +5/+5 until end of turn and must be blocked
//	 this turn if able."
//
// #1684 leftover: BoostUntilEOT plus the Irresistible Prey shape
// (BlockRequirementUntilEOT, must-be-blocked). No new machinery.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2afd64dd-94b3-4bc6-8411-465b3e07f3c9",
		Name:         "Emergent Growth",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 5, 5)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{
					Target:    t.ID,
					Power:     5,
					Toughness: 5,
					Label:     "Emergent Growth — +5/+5",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.BlockRequirementUntilEOT{
					Target: t.ID,
					Kind:   game.BlockRequirementMustBeBlocked,
					Label:  "Emergent Growth — must be blocked this turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
