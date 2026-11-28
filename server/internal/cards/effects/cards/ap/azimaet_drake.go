package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Azimaet Drake — Creature — Drake {2}{U}, 1/3 (Mirage, S82 OLD b797, #5633):
//
//	"Flying
//	 {U}: This creature gets +1/+0 until end of turn. Activate only once each
//	 turn."
//
// Beetleform Mage's once-per-turn gate: the announce tally is per object and
// per label, so an activation on the stack, countered or fizzled still uses it
// up (CR 602.5b).
//
// No simplification.
func init() {
	const label = "{U}: This creature gets +1/+0 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "a82d9c9b-ec17-4f51-a6bf-5984ca4433e5",
		Name:            "Azimaet Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:      effects.ManaCost("{U}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(1, 0, "Azimaet Drake — +1/+0 until end of turn"),
		}},
	})
}
