package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drake Hatchling — Creature — Drake {2}{U}, 1/3 (Mercadian Masques, S82 OLD,
// #5619):
//
//	"Flying
//	 {U}: This creature gets +1/+0 until end of turn. Activate only once each
//	 turn."
//
// Flying is printed data. The once-a-turn limit is Beetleform Mage's
// (CR 602.5b), counted per object, and the bonus ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	const label = "{U}: This creature gets +1/+0 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "bd55bcc4-f0c5-4c09-8b3b-2e25e9fbd9ee",
		Name:            "Drake Hatchling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:      effects.ManaCost("{U}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(1, 0, "Drake Hatchling — +1/+0 until end of turn"),
		}},
	})
}
