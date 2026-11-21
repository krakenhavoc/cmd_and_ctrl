package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowstone Thopter — Artifact Creature — Thopter {7}, 4/4
// (Nemesis, #5618):
//
//	"{1}: This creature gets +1/-1 and gains flying until end of turn."
//
// Leaping Lizard's row at +1/-1: the shared thisCreatureUntilEOT body
// pumps the Thopter and grants flying until the cleanup step (CR
// 514.2). A fourth activation leaves it 8/0, and it dies to the
// state-based check (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3169d1a-fd8d-474e-9e9f-1bd7db510e8b",
		Name:         "Flowstone Thopter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: This creature gets +1/-1 and gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}"),
			Effect:  effects.ThisCreatureUntilEOT("Flowstone Thopter — +1/-1 and flying until end of turn", 1, -1, "flying"),
		}},
	})
}
