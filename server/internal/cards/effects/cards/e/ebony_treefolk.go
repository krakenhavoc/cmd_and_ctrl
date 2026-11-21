package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ebony Treefolk — Creature — Treefolk {1}{B}{G}, 3/3 (Apocalypse,
// #5614):
//
//	"{B}{G}: This creature gets +1/+1 until end of turn."
//
// Evernight Shade's self pump at a two-colour price, with no limit on
// activations. The boost ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04683bd7-b100-4fc8-8165-316c5508c255",
		Name:         "Ebony Treefolk",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{G}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{B}{G}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Ebony Treefolk — +1/+1 until end of turn"),
		}},
	})
}
