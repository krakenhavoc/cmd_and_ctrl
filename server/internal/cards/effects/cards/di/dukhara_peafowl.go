package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dukhara Peafowl — Artifact Creature — Bird {4}, 2/4 (Kaladesh, #5539):
//
//	"{U}: This creature gains flying until end of turn."
//
// Femeref Knight's self grant: a layer-6 flying pinned to this object
// until the turn ends (CR 613.1f, CR 514.2) through the shared
// thisGainsKeywordUntilEndOfTurn body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "abc5ac3e-f196-49df-8a9d-16df3114ffad",
		Name:         "Dukhara Peafowl",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Dukhara Peafowl — flying until end of turn"),
		}},
	})
}
