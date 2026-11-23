package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire Aristocrat — Creature — Vampire Rogue Noble {2}{B}, 2/2 (Magic 2010, #5585):
//
//	"Sacrifice a creature: This creature gets +2/+2 until end of turn."
//
// Bloodthrone Vampire's sacrifice outlet: the creature is sacrificed as the
// cost, and the pump lasts until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c14739b0-5c19-4873-b027-314225a10f7b",
		Name:         "Vampire Aristocrat",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.SacrificeACreature(),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Vampire Aristocrat — +2/+2 until end of turn"),
		}},
	})
}
