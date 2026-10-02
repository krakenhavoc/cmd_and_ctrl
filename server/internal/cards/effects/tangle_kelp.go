package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tangle Kelp — Enchantment — Aura {U}:
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step
//	 if it attacked during its controller's last turn."
//
// Claustrophobia's entry trigger (shared) with an untap restriction (CR
// 502.3) that reads ADR 0108 §6's record. "Its controller" is the
// creature's controller now, so an enchanted creature that changes hands
// is asked of its new controller's last turn; it counts only if this very
// object attacked then (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06ba313f-50b4-4e84-a8e0-b6dda47225f2",
		Name:         "Tangle Kelp",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		UntapStepRestrictions: []game.UntapStepRestriction{
			doesntUntapIfItAttackedDuringYourLastTurn(AttachedToSource),
		},
		Triggered: []game.TriggeredAbility{WhenThisEnters("Tangle Kelp — tap enchanted creature", tapEnchantedCreatureOnEntry)},
	})
}
