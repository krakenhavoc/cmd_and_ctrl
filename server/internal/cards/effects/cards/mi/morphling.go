package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morphling — Creature — Shapeshifter {3}{U}{U} 3/3 (Urza's Saga, #5624):
//
//	"{U}: Untap this creature.
//	 {U}: This creature gains flying until end of turn.
//	 {U}: This creature gains shroud until end of turn. (It can't be the target
//	 of spells or abilities.)
//	 {1}: This creature gets +1/-1 until end of turn.
//	 {1}: This creature gets -1/+1 until end of turn."
//
// Five separate activated abilities, in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19f233b2-8be6-4172-abf9-b132fc3393d9",
		Name:         "Morphling",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}: Untap this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{U}"),
				Effect:  effects.B31UntapSelf,
			},
			{
				Label:   "{U}: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{U}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Morphling — flying until end of turn"),
			},
			{
				Label:   "{U}: This creature gains shroud until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    effects.ManaCost("{U}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("shroud", "Morphling — shroud until end of turn"),
			},
			{
				Label:   "{1}: This creature gets +1/-1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, -1, "Morphling — +1/-1 until end of turn"),
			},
			{
				Label:   "{1}: This creature gets -1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(-1, 1, "Morphling — -1/+1 until end of turn"),
			},
		},
	})
}
