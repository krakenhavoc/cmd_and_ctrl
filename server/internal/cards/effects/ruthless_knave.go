package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "00225c40-27ff-410f-a591-5c788c6a2bd6",
		Name:         "Ruthless Knave",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{B}, Sacrifice a creature: Create two Treasure tokens.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Tokens: 2},
				Cost:    Plus(ManaCost("{2}{B}"), SacrificeACreature()),
				Effect:  Do(CreateToken{Template: TreasureToken(), N: 2}),
			},
			{
				Label:   "Sacrifice three Treasures: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    SacrificeN(3, "three Treasures", HasSubtype("Treasure")),
				Effect:  Do(DrawCards{N: 1}),
			},
		},
	})
}
