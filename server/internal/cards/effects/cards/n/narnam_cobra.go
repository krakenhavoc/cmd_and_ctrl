package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Narnam Cobra — Artifact Creature — Snake {2}, 2/1 (Kaladesh, #5539):
//
//	"{G}: This creature gains deathtouch until end of turn. (Any amount
//	 of damage it deals to a creature is enough to destroy it.)"
//
// Femeref Knight's self grant: a layer-6 deathtouch pinned to this object
// until the turn ends (CR 613.1f, CR 514.2) through the shared
// thisGainsKeywordUntilEndOfTurn body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ac0d584-997b-4d11-aecf-f1494eebc94c",
		Name:         "Narnam Cobra",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{G}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("deathtouch", "Narnam Cobra — deathtouch until end of turn"),
		}},
	})
}
