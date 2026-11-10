package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diversion Specialist — Creature — Human Warrior {3}{R}, 4/3 (Duskmourn,
// #4466):
//
//	"Menace
//	 {1}, Sacrifice another creature or enchantment: Exile the top card of
//	 your library. You may play it this turn."
//
// The sacrifice is a cost (b23SacrificeAnotherCreatureOrEnchantment); the
// effect is the impulse exile of Breeches, Eager Pillager's third mode
// (exileTopCardYouMayPlayThisTurn): the card is played, not cast, so a land
// may be played from exile on a main phase with an empty stack, paying all
// costs and following all timing rules.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7130d063-fc1b-408b-97f2-14cb6b55d9cd",
		Name:            "Diversion Specialist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature or enchantment: Exile the top card of your library. You may play it this turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), b23SacrificeAnotherCreatureOrEnchantment()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return exileTopCardYouMayPlayThisTurn(item, NewContext(g, item), 0)
			},
		}},
	})
}
