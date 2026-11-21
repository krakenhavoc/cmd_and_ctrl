package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wolfsbane, Highland Hero — Legendary Creature — Mutant Werewolf Hero
// {1}{G}, 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Trample
//	 {2}{G}: Wolfsbane gets +2/+2 until end of turn. Activate only once
//	 each turn."
//
// Trample rides PrintedKeywords. The limit counts activations (CR 602.1b)
// as Barbarian Bully's does, and the pump is the shared self pump.
//
// No simplification.
func init() {
	const label = "{2}{G}: Wolfsbane gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "23d3ff11-bf19-4315-b53b-3e4f20c6a89b",
		Name:            "Wolfsbane, Highland Hero",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{2}{G}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Wolfsbane — +2/+2 until end of turn"),
		}},
	})
}
