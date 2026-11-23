package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mardu Hateblade — Creature — Human Warrior {W}, 1/1 (Khans of Tarkir,
// #5554):
//
//	"{B}: This creature gains deathtouch until end of turn. (Any amount
//	 of damage it deals to a creature is enough to destroy it.)"
//
// Twilight Panther's row: a layer-6 deathtouch grant pinned to the
// source until end of turn (thisGainsKeywordUntilEndOfTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acff2573-cf8a-4aee-96a3-28721a4128f6",
		Name:         "Mardu Hateblade",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("deathtouch", "Mardu Hateblade — deathtouch until end of turn"),
		}},
	})
}
