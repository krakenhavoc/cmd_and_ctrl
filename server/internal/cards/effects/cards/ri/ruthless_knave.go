package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruthless Knave — Creature — Orc Pirate {2}{B}, 3/2 (Foundations
// Jumpstart, S82 BNS):
//
//	"{2}{B}, Sacrifice a creature: Create two Treasure tokens.
//	 Sacrifice three Treasures: Draw a card."
//
// Both sacrifices are costs paid as the ability is activated (the creature
// may be Knave itself; the three Treasures are three separate permanents,
// Samwise Gamgee's SacrificeN). The draw is the controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00225c40-27ff-410f-a591-5c788c6a2bd6",
		Name:         "Ruthless Knave",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{B}, Sacrifice a creature: Create two Treasure tokens.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Tokens: 2},
				Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeACreature()),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 2}),
			},
			{
				Label:   "Sacrifice three Treasures: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.SacrificeN(3, "three Treasures", effects.HasSubtype("Treasure")),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
		},
	})
}
