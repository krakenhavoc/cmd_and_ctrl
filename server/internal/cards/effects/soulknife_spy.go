package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soulknife Spy — Creature — Elf Rogue {2}{U}, 3/2 (Adventures in the Forgotten Realms):
//
//	Whenever this creature deals combat damage to a player, draw a card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "813c0654-7e3b-435b-b7a5-148f0c36a4be",
		Name:         "Soulknife Spy",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Soulknife Spy — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
