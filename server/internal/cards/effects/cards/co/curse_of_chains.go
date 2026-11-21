package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Chains — Enchantment — Aura {1}{W/U} (Ravnica: Clue Edition):
//
//	"Enchant creature
//	 At the beginning of each upkeep, tap enchanted creature."
//
// The trigger belongs to the Aura and fires in every player's upkeep. The
// creature is found through the attachment as the ability resolves (Freed from
// the Real's tapAttachedHost), so an Aura that has fallen off taps nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be951382-5173-4fb6-9dda-2c5f7fddc863",
		Name:         "Curse of Chains",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Curse of Chains — tap enchanted creature", effects.TapAttachedHost),
		},
	})
}
