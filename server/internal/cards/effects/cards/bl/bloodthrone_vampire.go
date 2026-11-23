package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodthrone Vampire — Creature — Vampire {1}{B}, 1/1 (Rise of the Eldrazi,
// #5581):
//
//	"Sacrifice a creature: This creature gets +2/+2 until end of turn."
//
// Phyrexian Ghoul's sacrifice outlet. It can sacrifice itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa76cc50-5be0-4844-b3a4-fb33fca3cdd3",
		Name:         "Bloodthrone Vampire",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.SacrificeACreature(),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Bloodthrone Vampire — +2/+2 until end of turn"),
		}},
	})
}
