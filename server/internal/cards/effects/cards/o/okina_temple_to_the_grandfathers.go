package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Okina, Temple to the Grandfathers — Legendary Land (Champions of
// Kamigawa, #5603):
//
//	"{T}: Add {G}.
//	 {G}, {T}: Target legendary creature gets +1/+1 until end of turn."
//
// Shizo, Death's Storehouse's shape with Sword of the Chosen's body: the
// bonus ends in the cleanup step (CR 514.2), and a target that stops
// being legendary in response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ad69bfb-2e51-4fe9-8d2f-7d071a4f1c69",
		Name:         "Okina, Temple to the Grandfathers",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: Target legendary creature gets +1/+1 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)), game.AnswerPump),
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target legendary creature", effects.Legendary()),
			Effect:  effects.BoostTargetUntilEOT("Okina, Temple to the Grandfathers — +1/+1", 1, 1),
		}},
	})
}
