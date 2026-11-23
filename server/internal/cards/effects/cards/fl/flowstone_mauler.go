package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowstone Mauler — Creature — Beast {4}{R}{R}, 4/5 (Stronghold,
// #5627):
//
//	"Trample
//	 {R}: This creature gets +1/-1 until end of turn."
//
// Trample is a printed keyword. The pump is Flowstone Shambler's +1/-1 at
// {R}, with the shared thisGetsUntilEndOfTurn body; a fifth activation
// leaves it 9/0 and it dies to the state-based check (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ea205569-b5a0-4c86-b825-67cd48d3b2e8",
		Name:            "Flowstone Mauler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, -1, "Flowstone Mauler — +1/-1 until end of turn"),
		}},
	})
}
