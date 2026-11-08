package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elfsworn Giant — Creature — Giant {3}{G}{G}, 5/3 (Foundations, #4465):
//
//	"Reach
//	 Landfall — Whenever a land you control enters, create a 1/1 green
//	 Elf Warrior creature token."
//
// Zendikar's Roil's landfall token on a reach body: one token for each
// land that enters under your control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f6abdf34-cfdc-465c-b067-cdebb659406e",
		Name:            "Elfsworn Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Elfsworn Giant — create a 1/1 green Elf Warrior (landfall)", game.LandfallPayoff{Tokens: 1},
				Do(CreateToken{Template: TokenCard("1/1 green Elf Warrior"), N: 1})),
		},
	})
}
