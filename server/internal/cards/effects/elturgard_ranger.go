package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elturgard Ranger — Creature — Human Elf Ranger {4}{G}, 4/1 (Adventures
// in the Forgotten Realms):
//
//	"Reach
//	 When this creature enters, create a 2/2 green Wolf creature token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "27083627-a5d0-4fc6-9e66-fa44803ef080",
		Name:            "Elturgard Ranger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Elturgard Ranger — create a 2/2 green Wolf creature token",
				Do(CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1})),
		},
	})
}
