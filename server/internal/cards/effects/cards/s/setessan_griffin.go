package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Setessan Griffin — Creature — Griffin {4}{W}, 3/2 (Theros, S82 OLD b389, #5561):
//
//	"Flying
//	 {2}{G}{G}: This creature gets +2/+2 until end of turn. Activate only
//	 once each turn."
//
// Basking Rootwalla's once-a-turn pump (OncePerTurnActivation, CR 602.5b)
// at {2}{G}{G} and +2/+2, beside printed flying.
//
// No simplification.
func init() {
	const label = "{2}{G}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "fbab033e-3a8f-4c08-a445-46ae34c283eb",
		Name:            "Setessan Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{2}{G}{G}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Setessan Griffin — +2/+2 until end of turn"),
		}},
	})
}
