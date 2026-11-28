package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snowblind — Enchantment — Aura {3}{G} (Ice Age, S82 OLD b828, #5636):
//
//	"Enchant creature
//	 Enchanted creature gets -X/-Y. If that creature is attacking, X is the
//	 number of snow lands defending player controls. Otherwise, X is the
//	 number of snow lands its controller controls. Y is equal to X or to
//	 enchanted creature's toughness minus 1, whichever is smaller."
//
// One layer 7c modify (OldB828SnowblindStatic) recomputed live: X reads the
// defending player while the creature is attacking and its controller
// otherwise; Y is the smaller of X and its toughness minus one, so the
// creature is never killed by the toughness half.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f8bfa3d-7490-4259-a403-d13585ed2f66",
		Name:         "Snowblind",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.OldB828SnowblindStatic()},
	})
}
