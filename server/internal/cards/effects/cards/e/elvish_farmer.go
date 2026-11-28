package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Farmer — Creature — Elf {1}{G}, 0/2 (Fallen Empires, S82 OLD b853,
// #5637):
//
//	"At the beginning of your upkeep, put a spore counter on this creature.
//	 Remove three spore counters from this creature: Create a 1/1 green
//	 Saproling creature token.
//	 Sacrifice a Saproling: You gain 2 life."
//
// Thallid's spore engine, and a sacrifice cost that is paid as the ability is
// activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91fe1075-e170-4575-ae28-8d692f051637",
		Name:         "Elvish Farmer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB853SporeUpkeep("Elvish Farmer")},
		Activated: []effects.ActivatedAbility{
			effects.S82OldB853SaprolingRow(),
			{
				Label:   "Sacrifice a Saproling: You gain 2 life.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    effects.SacrificeN(1, "a Saproling", effects.HasSubtype("Saproling")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Amount: 2}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
