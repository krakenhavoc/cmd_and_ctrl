package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fungal Plots — Enchantment {1}{G}:
//
//	"{1}{G}, Exile a creature card from your graveyard: Create a 1/1
//	 green Saproling creature token.
//	 Sacrifice two Saprolings: You gain 2 life and draw a card."
//
// Both are ordinary activated abilities: the exile and the two
// sacrifices are costs (CR 118.12), so a table without the cards or the
// Saprolings cannot activate them. The two Saprolings may be any
// Saprolings you control, including ones other cards made.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "0a85402a-e3be-4ebb-8001-9d9a20dde7e8",
		Name:         "Fungal Plots",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}{G}, Exile a creature card from your graveyard: Create a 1/1 green Saproling creature token.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}{G}"), ExileFromGraveyard(1, "a creature card", isCreatureCard)),
				Effect:  Do(CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1}),
			},
			{
				Label:   "Sacrifice two Saprolings: You gain 2 life and draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 1},
				Cost:    SacrificeN(2, "two Saprolings", HasSubtype("Saproling")),
				Effect:  Do(GainLife{Amount: 2}, DrawCards{N: 1}),
			},
		},
	})
}
