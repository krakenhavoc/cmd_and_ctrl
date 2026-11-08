package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stop Cold — Enchantment — Aura {3}{U} (Outlaws of Thunder Junction,
// #4468):
//
//	"Flash
//	 Enchant artifact or creature
//	 When this Aura enters, tap enchanted permanent.
//	 Enchanted permanent loses all abilities and doesn't untap during its
//	 controller's untap step."
//
// Flood the Engine under Petrify's "enchant artifact or creature" clause,
// with flash: the entry tap (tapEnchantedCreatureOnEntry, which taps
// whatever the Aura is attached to as the trigger resolves, CR 608.2h)
// and the two statics (LoseAllAbilities, enchantedDoesntUntap), which end
// when the Aura leaves. A noncreature artifact loses its abilities too,
// mana abilities included. The enchant clause is also the legality check
// the Aura is held to after it enters (CR 702.5a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "e0206af2-8034-4351-a797-a1d996f080b5",
		Name:                  "Stop Cold",
		Completeness:          CompletenessFull,
		PrintedKeywords:       []string{"flash"},
		Targets:               TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Static:                []game.StaticAbility{LoseAllAbilities()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Stop Cold — tap enchanted permanent", tapEnchantedCreatureOnEntry),
		},
	})
}
