package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battlefield Percher — Creature — Bird {3}{B}{B}, 2/2
// (Nemesis, #5618):
//
//	"Flying
//	 This creature can block only creatures with flying.
//	 {1}{B}: This creature gets +1/+1 until end of turn."
//
// Belbe's Percher with a pump: the block rule is nemBlocksOnlyFlyers,
// and the {1}{B} row is the shared thisGetsUntilEndOfTurn body at
// +1/+1, ending in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2836c144-7268-411a-b651-351cc4ef8fe9",
		Name:            "Battlefield Percher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      effects.NemBlocksOnlyFlyers(effects.OnSelf()),
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{1}{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Battlefield Percher — +1/+1 until end of turn"),
		}},
	})
}
