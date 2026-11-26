package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boon of Emrakul — Enchantment — Aura {2}{B} (Eldritch Moon, S82 OLD, #5541):
//
//	"Enchant creature
//	 Enchanted creature gets +3/-3."
//
// Dead Weight's shape: a layer 7c change of the host. A host left with 0 or
// less toughness dies to the state-based action (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20d30d4d-cf38-41fe-a2ce-8327f7c87c01",
		Name:         "Boon of Emrakul",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpEnchantedCreature(3, -3)},
	})
}
