package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Encase in Ice — Enchantment — Aura {1}{U} (Dragons of Tarkir, S82 old batch 332):
//
//	"Flash
//	 Enchant red or green creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Bitter Chill's tap-and-lock, without its death trigger, with flash and an enchant
// clause that names two colours. The clause is checked as the Aura is cast and
// again as state-based actions run (CR 704.5m), so a creature that stops being
// red or green puts the Aura into the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "a61277d6-a27d-45d1-ba2a-7afe1b8f5ce1",
		Name:                  "Encase in Ice",
		Completeness:          effects.CompletenessFull,
		PrintedKeywords:       []string{"flash"},
		Targets:               effects.TargetCreature("enchant red or green creature", effects.Or(effects.OfColor("R"), effects.OfColor("G"))),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Encase in Ice — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
		},
	})
}
