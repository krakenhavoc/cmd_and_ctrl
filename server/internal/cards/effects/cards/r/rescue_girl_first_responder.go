package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rescue Girl, First Responder — Legendary Creature — Human Cleric
// {2}{W}, 1/3:
//
//	"Flying
//	 {T}: Return another target permanent you control to its owner's
//	 hand. Activate only during your turn."
//
// The target is chosen at activation from your other permanents, and
// the activation is refused outside your own turn (any step). The
// bounce goes through BounceToHand, so a commander asks its owner about
// the command zone. The tap makes it subject to summoning sickness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "df6fc0b4-333b-4f6c-804c-ba6c8e29c88e",
		Name:            "Rescue Girl, First Responder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Return another target permanent you control to its owner's hand. Activate only during your turn",
			Cost:      effects.TapCost(),
			Targets:   effects.Another(effects.TargetPermanent("another target permanent you control", effects.YouControl())),
			Condition: effects.DuringYourTurn(),
			Effect:    effects.BounceChosenTarget,
		}},
	})
}
