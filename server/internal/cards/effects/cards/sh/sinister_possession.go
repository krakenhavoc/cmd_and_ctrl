package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sinister Possession — Enchantment — Aura {B} (Dragon's Maze, S82 OLD b405,
// #5564):
//
//	"Enchant creature
//	 Whenever enchanted creature attacks or blocks, its controller loses 2
//	 life."
//
// Contaminated Bond's trigger for 2 (S82OldB405EnchantedAttacksOrBlocksLoseLife):
// the loser is the enchanted creature's controller, not the Aura's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ee431f3-64ec-4015-8f5a-9c2ac6004ebd",
		Name:         "Sinister Possession",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB405EnchantedAttacksOrBlocksLoseLife("Sinister Possession — the enchanted creature's controller loses 2 life", 2),
		},
	})
}
