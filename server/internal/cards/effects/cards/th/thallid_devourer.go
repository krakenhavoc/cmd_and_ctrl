package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thallid Devourer — Creature — Fungus {1}{G}{G}, 2/2 (Fallen Empires, S82 OLD
// b853, #5637):
//
//	"At the beginning of your upkeep, put a spore counter on this creature.
//	 Remove three spore counters from this creature: Create a 1/1 green
//	 Saproling creature token.
//	 Sacrifice a Saproling: This creature gets +1/+2 until end of turn."
//
// Thallid's spore engine and Nemata's Saproling sacrifice pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7326fd9-7648-4db3-b506-61a7de8a4094",
		Name:         "Thallid Devourer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB853SporeUpkeep("Thallid Devourer")},
		Activated: []effects.ActivatedAbility{
			effects.S82OldB853SaprolingRow(),
			{
				Label:   "Sacrifice a Saproling: This creature gets +1/+2 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
				Cost:    effects.SacrificeN(1, "a Saproling", effects.HasSubtype("Saproling")),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 2, "Thallid Devourer — +1/+2 until end of turn"),
			},
		},
	})
}
