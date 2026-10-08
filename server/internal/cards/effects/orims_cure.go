package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "61eacaab-625b-499c-8602-f37ad5dda138",
		Name:         "Orim's Cure",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		AlternativeCosts: []game.AlternativeCost{
			TapInstead(1, "an untapped creature you control", ControlsA("Plains"), Creature()),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return PreventNextDamage{
				Target: item.Targets[0].ID,
				Amount: 4,
				Label:  "Orim's Cure: prevent the next 4 damage",
			}.Apply(ctx)
		},
	})
}
