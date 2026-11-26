package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merrow Grimeblotter — Creature — Merfolk Wizard {3}{U/B} 2/2:
//
//	"{1}{U/B}, {Q}: Target creature gets -2/-0 until end of turn. ({Q} is the
//	 untap symbol.)"
//
// {Q} is the untap-self cost (Umbral Mantle's): the Grimeblotter must be
// tapped to activate and summoning sickness does not apply to it (CR 302.6).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e9a63d2-b48b-4740-95ba-7966ada95baf",
		Name:         "Merrow Grimeblotter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U/B}, {Q}: Target creature gets -2/-0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    game.AbilityCost{Mana: "{1}{U/B}", UntapSelf: true},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Merrow Grimeblotter — -2/-0 until end of turn", -2, 0),
		}},
	})
}
