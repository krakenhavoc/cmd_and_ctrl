package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Simic Ragworm — Creature — Worm {3}{G}, 3/3 (Dissension, #5598):
//
//	"{U}: Untap this creature."
//
// Blizzard Elemental's untap row for {U}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d8c4172-eb14-47a4-993a-6d68906a26c5",
		Name:         "Simic Ragworm",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			UntapSelf: true,
			Label:     "{U}: Untap this creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{U}"),
			Effect:    effects.B31UntapSelf,
		}},
	})
}
