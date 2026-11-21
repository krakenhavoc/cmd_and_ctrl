package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peema Aether-Seer — Creature — Elf Druid {3}{G}, 3/2:
//
//	"When this creature enters, you get an amount of {E} (energy
//	 counters) equal to the greatest power among creatures you control.
//	 Pay {E}{E}{E}: Target creature blocks this turn if able."
//
// The amount is counted as the trigger resolves, layered power, and the
// Seer counts itself ("creatures you control", not "other"). Being
// counted at resolution, it declares no Purpose amount (ADR 0126 §6).
// The block requirement is a scoped CR 509.1c requirement until end of
// turn on the target (BlockRequirementUntilEOT).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a00f56e-e558-4239-bc75-fb0af7639042",
		Name:         "Peema Aether-Seer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Peema Aether-Seer — you get {E} equal to the greatest power among creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GetEnergy{N: effects.B42GreatestPowerControlledBy(g, item.Controller)}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay {E}{E}{E}: Target creature blocks this turn if able.",
			Cost:    effects.PayEnergy(3),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BlockRequirementUntilEOT{
					Target: effects.FirstLegalBattlefieldTarget(ctx),
					Kind:   game.BlockRequirementBlocks,
					Label:  "Peema Aether-Seer — blocks this turn if able",
				}.Apply(ctx)
			},
		}},
	})
}
