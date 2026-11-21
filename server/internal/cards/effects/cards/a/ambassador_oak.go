package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ambassador Oak — Creature — Treefolk Warrior {3}{G}, 3/3 (Foundations
// Jumpstart, #5929):
//
//	"When this creature enters, create a 1/1 green Elf Warrior creature
//	 token."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4045c24e-098f-4582-8c50-576a94c7b705",
		Name:         "Ambassador Oak",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ambassador Oak — create a 1/1 green Elf Warrior creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 1})),
		},
	})
}
