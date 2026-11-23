package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soulknife Spy — Creature — Elf Rogue {2}{U}, 3/2 (Adventures in the Forgotten Realms):
//
//	Whenever this creature deals combat damage to a player, draw a card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "813c0654-7e3b-435b-b7a5-148f0c36a4be",
		Name:         "Soulknife Spy",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Soulknife Spy — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
