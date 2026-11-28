package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crazed Armodon — Creature — Elephant {2}{G}{G}, 3/3 (Tempest, S82 OLD b756,
// #5628):
//
//	"{G}: This creature gets +3/+0 and gains trample until end of turn.
//	 Destroy this creature at the beginning of the next end step. Activate
//	 only once each turn."
//
// The once-a-turn limit is Roterothopter's AtMostNActivationsPerTurn (announcements
// this turn, CR 602.1b). The destruction is a delayed trigger (CR 603.7) that
// destroys the Armodon as it is then, if it is still the same object
// (OldB756ThisGetsAndGainsThenDestroyAtEndStep).
//
// No simplification.
func init() {
	const label = "{G}: This creature gets +3/+0 and gains trample until end of turn. Destroy this creature at the beginning of the next end step. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "ce7d672a-7cb9-40e5-a8ad-dd3ce0f8ecd3",
		Name:         "Crazed Armodon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 3, Keywords: []string{"trample"}}},
			Cost:      effects.ManaCost("{G}"),
			Condition: effects.AtMostNActivationsPerTurn(label, 1),
			Effect:    effects.OldB756ThisGetsAndGainsThenDestroyAtEndStep("Crazed Armodon", 3, "trample"),
		}},
	})
}
