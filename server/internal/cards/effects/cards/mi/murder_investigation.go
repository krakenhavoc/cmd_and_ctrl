package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Murder Investigation — Enchantment — Aura {1}{W} (Jumpstart 2022, #5948):
//
//	"Enchant creature you control
//	 When enchanted creature dies, create X 1/1 white Soldier creature tokens,
//	 where X is its power."
//
// Angelic Destiny's dies trigger (harvested while the Aura is still attached)
// and Elenda's Hierophant's token count: X is the creature's last-known power
// (CR 603.10a), and the Soldiers are the Aura controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42ca56c5-a9c4-4a2c-ae05-bce47aa6e16c",
		Name:         "Murder Investigation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Murder Investigation — create X 1/1 white Soldier creature tokens, where X is its power", effects.GapB169SoldiersEqualToItsPower),
		},
	})
}
