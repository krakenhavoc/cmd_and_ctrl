package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boreal Griffin — Snow Creature — Griffin {3}{W}{W}, 3/2 (Coldsnap,
// #5986):
//
//	"Flying
//	 {S}: This creature gains first strike until end of turn. ({S} can be
//	 paid with one mana from a snow source.)"
//
// Burning Shield Askari's first-strike grant through the shared
// thisGainsKeywordUntilEndOfTurn body, paid by one mana of any colour
// from a snow source (CR 107.4h, #5986). Gained after blockers are
// declared, it still deals its damage in the first-strike step
// (CR 510.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "862ecb4e-7030-48d5-bbdf-655662059430",
		Name:            "Boreal Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{S}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
			Cost:    effects.ManaCost("{S}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Boreal Griffin — first strike until end of turn"),
		}},
	})
}
