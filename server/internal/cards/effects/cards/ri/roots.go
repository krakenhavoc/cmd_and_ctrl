package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roots — Enchantment — Aura {3}{G} (Eternal Masters, S82 OLD, #5542):
//
//	"Enchant creature without flying
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Bitter Chill's two halves around a narrower enchant clause: the enters trigger
// taps the host and the untap-step restriction keeps it tapped. The enchant
// restriction is the Aura spell's target (CR 303.4a) and the state-based action
// that puts an illegally attached Aura into the graveyard (CR 704.5m): a host
// that gains flying later takes the Aura with it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "0dcee3a3-6cb1-45ce-958c-1f57ced80be4",
		Name:                  "Roots",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(effects.Not(effects.HasKeyword("flying"))),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Roots — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
		},
	})
}
