package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elderleaf Mentor — Creature — Elf Warrior {3}{G}, 3/2 (Kaldheim, #5505):
//
//	"When this creature enters, create a 1/1 green Elf Warrior creature
//	 token."
//
// Ambassador Oak's enters trigger, word for word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3795fc31-2369-47d0-b9c9-120f44b8f114",
		Name:         "Elderleaf Mentor",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Elderleaf Mentor — create a 1/1 green Elf Warrior creature token",
				Do(CreateToken{Template: TokenCard("1/1 green Elf Warrior"), N: 1})),
		},
	})
}
