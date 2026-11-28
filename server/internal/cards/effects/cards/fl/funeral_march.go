package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Funeral March — Enchantment — Aura {1}{B}{B} (Homelands, S82 OLD b825, #5635):
//
//	"Enchant creature
//	 When enchanted creature leaves the battlefield, its controller sacrifices a
//	 creature of their choice."
//
// Any departure, not only dying (CR 603.10a: the trigger looks back to the
// Aura still attached). The controller is the one the creature had as it last
// existed on the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	const label = "Funeral March — its controller sacrifices a creature of their choice"
	effects.Register(effects.Spec{
		OracleID:     "4d8afa5c-cef3-4e67-bd4c-ebd5c68b96a7",
		Name:         "Funeral March",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WhenEnchantedCreatureLeavesTheBattlefield(label, effects.ItsControllerSacrificesACreature(label))),
		},
	})
}
