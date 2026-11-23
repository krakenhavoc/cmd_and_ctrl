package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elderleaf Mentor — Creature — Elf Warrior {3}{G}, 3/2 (Kaldheim, #5505):
//
//	"When this creature enters, create a 1/1 green Elf Warrior creature
//	 token."
//
// Ambassador Oak's enters trigger, word for word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3795fc31-2369-47d0-b9c9-120f44b8f114",
		Name:         "Elderleaf Mentor",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Elderleaf Mentor — create a 1/1 green Elf Warrior creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 1})),
		},
	})
}
