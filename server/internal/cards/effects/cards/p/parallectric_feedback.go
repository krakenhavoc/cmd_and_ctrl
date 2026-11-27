package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Parallectric Feedback — Instant {3}{R} (Guildpact):
//
//	"Parallectric Feedback deals damage to target spell's controller equal
//	 to that spell's mana value."
//
// The spell is the one target; with it illegal on resolution (countered or
// gone) the whole spell does nothing (CR 608.2b). Its mana value counts an {X}
// as the value chosen while it is on the stack (CR 202.3e), and its controller
// is read as it now is (CR 108.4: a spell whose control changed hits the new
// controller). The damage is dealt by Parallectric Feedback, so it is a red
// source, not the targeted spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a01071c5-aca5-4faa-8a8a-30b50b642f42",
		Name:         "Parallectric Feedback",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			spell := targets[0].ID
			controller, ok := ctx.Game.LastKnownSpellControllerForEffect(spell)
			if !ok || controller == uuid.Nil {
				return nil
			}
			mv := effects.SpellManaValueForEffect(ctx.Game, spell)
			if mv <= 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: mv}.Apply(ctx)
		})),
	})
}
