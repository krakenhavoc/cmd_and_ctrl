package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ferocious Pup — Creature — Wolf {2}{G}, 0/1 (Core Set 2020, #5518):
//
//	"When this creature enters, create a 2/2 green Wolf creature token."
//
// Elturgard Ranger's enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a1112750-8d38-49e0-ab1c-77110b2bbc4d",
		Name:         "Ferocious Pup",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ferocious Pup — create a 2/2 green Wolf creature token",
				Do(CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1})),
		},
	})
}
