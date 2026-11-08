package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seam Rip — Enchantment {W} (Edge of Eternities, #4461):
//
//	"When this enchantment enters, exile target nonland permanent an
//	 opponent controls with mana value 2 or less until this enchantment
//	 leaves the battlefield."
//
// Glass Casket's entry trigger (exileChosenTargetUntilThisLeaves) over a
// nonland permanent an opponent controls with mana value 2 or less
// (CR 202.3), read at targeting and again as the trigger resolves
// (CR 608.2b). The card returns under its owner's control when Seam Rip
// leaves (CR 610.3c); if Seam Rip has already left when the trigger
// resolves, nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dfd5939e-a71a-4e16-9b93-752162be4ccd",
		Name:         "Seam Rip",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Seam Rip — exile target nonland permanent an opponent controls with mana value 2 or less until this enchantment leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Seam Rip — the exiled card returns when Seam Rip leaves the battlefield")),
				TargetPermanent("target nonland permanent an opponent controls with mana value 2 or less", Nonland(), OpponentControls(), ManaValueLE(2))),
		},
	})
}
