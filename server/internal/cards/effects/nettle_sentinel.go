package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nettle Sentinel — Creature — Elf Warrior {G}, 2/2 (Eventide, #5590):
//
//	"This creature doesn't untap during your untap step.
//	 Whenever you cast a green spell, you may untap this creature."
//
// Brass Man's untap-step restriction and the optional cast-trigger untap
// (eveB1UntapOnCast).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "5c367b7d-3cb4-45db-b836-dad953efce15",
		Name:                  "Nettle Sentinel",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Triggered:             []game.TriggeredAbility{eveB1UntapOnCast("Nettle Sentinel", "G")},
	})
}
