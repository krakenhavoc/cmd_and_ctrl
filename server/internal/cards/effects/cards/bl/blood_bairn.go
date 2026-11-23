package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Bairn — Creature — Vampire {2}{B}, 2/2 (Magic 2014, #5563):
//
//	"Sacrifice another creature: This creature gets +2/+2 until end of turn."
//
// Sepulcher Ghoul's sacrifice pump without the once-a-turn limit: it can be
// activated as often as there are other creatures to sacrifice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33bea2f0-c5aa-44b5-8714-5f50f7f2bb05",
		Name:         "Blood Bairn",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump},
			Cost:    effects.SacrificeAnotherN(1, "another creature", effects.Creature()),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Blood Bairn — +2/+2"),
		}},
	})
}
