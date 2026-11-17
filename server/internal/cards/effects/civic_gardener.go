package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Civic Gardener — Creature — Human Citizen {1}{G}, 2/2 (Streets of New
// Capenna):
//
//	"Whenever this creature attacks, untap target creature or land."
//
// Cat-Owl's attack trigger over creatures and lands; the target is chosen as
// the trigger goes on the stack (CR 603.3d), and may be the Gardener itself.
// A target gone or no longer a creature or land as it resolves is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9eb7f301-b9b6-48c3-9dad-e7df9cc15355",
		Name:         "Civic Gardener",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Civic Gardener — untap target creature or land", untapTheTarget),
				TargetPermanent("target creature or land", Or(Creature(), Land()))),
		},
	})
}
