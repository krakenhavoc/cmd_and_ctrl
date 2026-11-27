package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akki Underling — Creature — Goblin Warrior {1}{R}, 2/1 (Saviors of
// Kamigawa, #5601):
//
//	"As long as you have seven or more cards in hand, this creature gets
//	 +2/+1 and has first strike."
//
// A layer 7c +2/+1 and a layer 6 first strike, both read live against the size
// of your hand (DependsOnHandSize).
//
// No simplification.
func init() {
	const when = "As long as you have seven or more cards in hand, this creature "
	sevenCards := effects.OldB588YouHaveAtLeastCardsInHand(7)
	effects.Register(effects.Spec{
		OracleID:     "91f95d83-3637-4cf8-af6c-56bc4e4b2591",
		Name:         "Akki Underling",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.OldB588WhileYouHavePump(sevenCards, 2, 1, when+"gets +2/+1."),
			effects.OldB588WhileYouHaveKeyword(sevenCards, "first strike", when+"has first strike."),
		},
	})
}
