package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bubble Snare — Enchantment — Aura {U} (Zendikar Rising, S82 OLD #5508):
//
//	"Kicker {2}{U} (You may pay an additional {2}{U} as you cast this
//	 spell.)
//	 Enchant creature
//	 When this Aura enters, if it was kicked, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// The tap is Bitter Chill's, gated on the kicker the entering Aura carries;
// the untap hold is the enchanted creature's, for as long as the Aura stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "be315fa6-afb5-4ab6-a143-7bbf37046d15",
		Name:                  "Bubble Snare",
		Completeness:          effects.CompletenessFull,
		OptionalCosts:         []game.AdditionalCost{effects.Kicker("{2}{U}")},
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Bubble Snare — kicked: tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
		},
	})
}
