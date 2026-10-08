package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sivvi's Valor — Instant {2}{W}:
//
//	"If you control a Plains, you may tap an untapped creature you
//	 control rather than pay this spell's mana cost.
//	 All damage that would be dealt to target creature this turn is
//	 dealt to you instead."
//
// ADR 0135 §1 (#2030): the tap alternative cost (CR 118.9), offered only
// while you control a Plains, and ADR 0108 §9's redirection for the rest
// of the turn from the target to the spell's controller. A target that
// has left redirects nothing (CR 608.2b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "e02164ba-34f8-4a5f-a05b-dd3ef3f8ceae",
		Name:         "Sivvi's Valor",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{
			TapInstead(1, "an untapped creature you control", ControlsA("Plains"), Creature()),
		},
		OnResolve: redirectSpell(RedirectDamage{Protect: ShieldTheTarget, To: RedirectToYou}),
	})
}
