package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viashino Slaughtermaster — Creature — Lizard Warrior {1}{R}, 1/1
// (Conflux, #5587):
//
//	"Double strike
//	 {B}{G}: This creature gets +1/+1 until end of turn. Activate only
//	 once each turn."
//
// Plated Rootwalla's once-each-turn pump (OncePerTurnActivation, CR
// 602.5b) at +1/+1, on a double striker.
//
// No simplification.
func init() {
	const pump = "{B}{G}: This creature gets +1/+1 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:        "34bdf2f1-0cc9-4588-9a3e-5cfe5cde815d",
		Name:            "Viashino Slaughtermaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Activated: []ActivatedAbility{{
			Label:     pump,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:      ManaCost("{B}{G}"),
			Condition: OncePerTurnActivation(pump),
			Effect:    thisGetsUntilEndOfTurn(1, 1, "Viashino Slaughtermaster — +1/+1 until end of turn"),
		}},
	})
}
