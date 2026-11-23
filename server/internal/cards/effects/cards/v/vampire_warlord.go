package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire Warlord — Creature — Vampire Warrior {4}{B}, 4/2 (Magic 2014, #5563):
//
//	"Sacrifice another creature: Regenerate this creature. (The next time this creature would be destroyed this turn, instead tap it, remove it from combat, and heal all damage on it.)"
//
// Eater of Hope's regeneration for a sacrifice, with no mana: each activation
// is one regeneration shield (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "393586d3-1f1a-4f3a-96b8-ed1a8fdf00c6",
		Name:         "Vampire Warlord",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.SacrificeAnotherN(1, "another creature", effects.Creature()),
			Effect:  effects.RegenerateThisSource,
		}},
	})
}
