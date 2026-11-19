package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ambassador Oak — Creature — Treefolk Warrior {3}{G}, 3/3 (Foundations
// Jumpstart, #5929):
//
//	"When this creature enters, create a 1/1 green Elf Warrior creature
//	 token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4045c24e-098f-4582-8c50-576a94c7b705",
		Name:         "Ambassador Oak",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ambassador Oak — create a 1/1 green Elf Warrior creature token",
				Do(CreateToken{Template: TokenCard("1/1 green Elf Warrior"), N: 1})),
		},
	})
}
