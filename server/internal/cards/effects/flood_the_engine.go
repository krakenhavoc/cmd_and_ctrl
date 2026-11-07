package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flood the Engine — Enchantment — Aura {2}{U} (Aetherdrift, #4464):
//
//	"Enchant creature or Vehicle
//	 When this Aura enters, tap enchanted permanent.
//	 Enchanted permanent loses all abilities and doesn't untap during its
//	 controller's untap step."
//
// Infinite Coursework's two statics (LoseAllAbilities, enchantedDoesntUntap)
// and Claustrophobia's entry tap (tapEnchantedCreatureOnEntry, which taps
// whatever permanent the Aura is attached to), under Aether Meltdown's
// "enchant creature or Vehicle" clause. An uncrewed Vehicle can be
// enchanted, and losing all abilities takes its crew ability with it
// (CR 613.1f), so it can no longer be crewed. The trigger reads the
// Aura's attachment as it resolves (CR 608.2h). The
// enchant clause is also the legality check the Aura is held to after it
// enters (CR 702.5a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "e16e2199-d821-49fd-94a0-30ecebd8bd4c",
		Name:                  "Flood the Engine",
		Completeness:          CompletenessFull,
		Targets:               TargetPermanent("enchant creature or Vehicle", Or(Creature(), HasSubtype("Vehicle"))),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Static:                []game.StaticAbility{LoseAllAbilities()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Flood the Engine — tap enchanted permanent", tapEnchantedCreatureOnEntry),
		},
	})
}
