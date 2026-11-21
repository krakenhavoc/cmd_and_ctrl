package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "e02164ba-34f8-4a5f-a05b-dd3ef3f8ceae",
		Name:         "Sivvi's Valor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{
			effects.TapInstead(1, "an untapped creature you control", effects.ControlsA("Plains"), effects.Creature()),
		},
		OnResolve: effects.RedirectSpell(effects.RedirectDamage{Protect: effects.ShieldTheTarget, To: effects.RedirectToYou}),
	})
}
