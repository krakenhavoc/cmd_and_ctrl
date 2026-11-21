package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soulreaper of Mogis — Enchantment Creature — Minotaur Shaman {2}{B}, 2/3 (Theros Beyond Death, #5514):
//
//	"{2}{B}, Sacrifice a creature: Draw a card."
//
// Grim Backwoods' sacrifice outlet without the tap: any creature its
// controller controls, itself included, pays the cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "508ce3a9-9050-4494-9adf-7f9cd6cc2f9f",
		Name:         "Soulreaper of Mogis",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice a creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeACreature()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
