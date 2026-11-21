package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orim's Cure — Instant {1}{W}:
//
//	"If you control a Plains, you may tap an untapped creature you
//	 control rather than pay this spell's mana cost.
//	 Prevent the next 4 damage that would be dealt to any target this
//	 turn."
//
// ADR 0135 §1 (#2030): the tap alternative cost (CR 118.9), offered only
// while you control a Plains, and Mending Hands' charged shield. The
// creature is tapped as a cost with the spell already on the stack; it
// may have arrived this turn (CR 302.6), and it may be the creature the
// shield protects.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61eacaab-625b-499c-8602-f37ad5dda138",
		Name:         "Orim's Cure",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		AlternativeCosts: []game.AlternativeCost{
			effects.TapInstead(1, "an untapped creature you control", effects.ControlsA("Plains"), effects.Creature()),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.PreventNextDamage{
				Target: item.Targets[0].ID,
				Amount: 4,
				Label:  "Orim's Cure: prevent the next 4 damage",
			}.Apply(ctx)
		},
	})
}
