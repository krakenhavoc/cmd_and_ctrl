package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gitaxian Raptor — Creature — Phyrexian Bird {2}{U}, 1/4 (Phyrexia: All
// Will Be One):
//
//	"Flying
//	 This creature enters with three oil counters on it.
//	 Remove an oil counter from this creature: This creature gets +1/-1
//	 until end of turn."
//
// The toughness loss is real: a fourth +1/-1 on an undamaged Raptor is
// impossible (it has only three counters), but damage plus pumps can still
// kill it as a state-based action.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6096b221-7b6e-4528-8b5d-8ad0fcef3fba",
		Name:            "Gitaxian Raptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{b10EntersWithCounters(oilCounter, 3, "Gitaxian Raptor: enters with three oil counters on it")},
		Activated: []ActivatedAbility{{
			Label:   "Remove an oil counter from this creature: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    RemoveCountersFromThis(oilCounter, 1),
			Effect:  thisGetsUntilEndOfTurn(1, -1, "Gitaxian Raptor — +1/-1 until end of turn"),
		}},
	})
}
