package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gruul Nodorog — Creature — Beast {4}{G}{G}, 4/4 (Guildpact):
//
//	"{R}: This creature gains menace until end of turn. (It can't be blocked
//	 except by two or more creatures.)"
//
// A layer 6 grant of menace on the creature itself until end of turn, River
// Merfolk's mountainwalk grant with another keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68fc3611-8d82-4f87-ab50-603790a27455",
		Name:         "Gruul Nodorog",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gains menace until end of turn. (It can't be blocked except by two or more creatures.)",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("menace", "Gruul Nodorog — menace until end of turn"),
		}},
	})
}
