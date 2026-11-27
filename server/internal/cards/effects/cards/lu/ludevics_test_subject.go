package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ludevic's Test Subject // Ludevic's Abomination — Creature — Lizard Egg
// {1}{U}, 0/3 // Creature — Lizard Horror, 13/13 (Innistrad, S82 older sets
// batch 449, #5572):
//
//	"Defender
//	 {1}{U}: Put a hatchling counter on this creature. Then if there are
//	 five or more hatchling counters on it, remove all of them and
//	 transform it."
//	// "Trample"
//
// Two Specs, the back face's key being "<oracle id>#1". The counters are
// counted after the new one lands, so the fifth hatches it.
//
// No simplification.
func init() {
	const oracle = "53175678-1f5c-4889-9086-66381b3ddc3a"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Ludevic's Test Subject",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Put a hatchling counter on this creature. Then if there are five or more hatchling counters on it, remove all of them and transform it.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect:  effects.S82OldB449HatchlingHatch,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Ludevic's Abomination",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
