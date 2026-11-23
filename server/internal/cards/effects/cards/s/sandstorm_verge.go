package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandstorm Verge — Land — Desert (Outlaws of Thunder Junction, #4468):
//
//	"{T}: Add {C}.
//	 {3}, {T}: Target creature can't block this turn. Activate only as a
//	 sorcery."
//
// The colourless mana ability, and Brambleback Brute's sorcery-speed
// "target creature can't block this turn" with {3} and {T} as the
// cost. With the target gone by resolution the ability does nothing.
// Desert is on the printed type line.
//
// No simplification.
func init() {
	const label = "Sandstorm Verge — target creature can't block this turn"
	effects.Register(effects.Spec{
		OracleID:     "de417a82-8f03-4d7e-aee7-48f7d7eba61a",
		Name:         "Sandstorm Verge",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}, {T}: Target creature can't block this turn. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets:      effects.TargetCreature("target creature"),
			Purpose:      effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			SorcerySpeed: true,
			Effect:       effects.TargetCreatureCantBlockThisTurn(label),
		}},
	})
}
