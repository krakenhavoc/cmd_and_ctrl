package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Head of the Homestead — Creature — Rabbit Citizen {3}{G/W}{G/W}, 3/2
// (Bloomburrow, #4467):
//
//	"When this creature enters, create two 1/1 white Rabbit creature
//	 tokens."
//
// Beetleback Chief's entry trigger with Hop to It's Rabbits: three bodies
// for five mana, the hybrid cost read off the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "467d6acb-9a3c-433f-836d-54e0e670b7d0",
		Name:         "Head of the Homestead",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Head of the Homestead — create two 1/1 white Rabbits",
				Do(CreateToken{Template: TokenCard("1/1 white Rabbit"), N: 2})),
		},
	})
}
