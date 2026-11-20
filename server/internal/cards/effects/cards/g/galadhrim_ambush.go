package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galadhrim Ambush — Instant {3}{G}:
//
//	"Create X 1/1 green Elf Warrior creature tokens, where X is the
//	 number of attacking creatures.
//	 Prevent all combat damage that would be dealt this turn by non-Elf
//	 creatures."
//
// X counts every attacking creature, whoever it attacks and whoever
// controls it (the ruling), as the spell resolves (CR 608.2h). The
// shield is #2026's negation over a subtype, read as each creature would
// deal combat damage (CR 609.7b).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b22d21f-e19a-40df-840b-a2b7f11f79c2",
		Name:         "Galadhrim Ambush",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			// The shield waits for a prompt the Elves' creation asks (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				n := len(effects.AttackingCreatures(ctx.Game))
				if n <= 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: n}.Apply(ctx)
			}),
			effects.CombatShieldAgainstCreatures(effects.ExceptSubtypes("Elf")),
		),
	})
}
