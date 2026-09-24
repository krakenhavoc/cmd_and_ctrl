package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seedborn Muse — Creature — Spirit {3}{G}{G}, 2/4 (EDHREC rank 285):
//
//	"Untap all permanents you control during each other player's
//	 untap step."
//
// The card the untap-step seam was named after in the coverage
// roadmap, and the reason it could not be written until now: the
// engine's untap step was a bare loop over the active seat's
// permanents with no hook a card could add to, and the only thing
// that could see the step at all was a replacement effect that had
// to CANCEL it to observe it.
//
// It is not a trigger and is not implemented as one. Nothing is
// announced, nothing goes on the stack and there is nothing to
// respond to — the untap step grants no priority (CR 502.4). The
// clause widens CR 502.3's "the active player determines which
// permanents they control untap", and Spec.UntapStep is exactly that
// widening. See game/untap.go.
//
// "All permanents", not "all creatures and lands": the Muse untaps
// everything, which is why she is the four-mana engine behind every
// untap-for-value deck in the format.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "463865bc-087e-477b-9e86-84e77f1ad931",
		Name:         "Seedborn Muse",
		Completeness: CompletenessFull,
		UntapStep: []game.UntapStepPermission{
			untapDuringEachOtherPlayersUntapStep(
				"Seedborn Muse — untap all permanents you control", nil),
		},
	})
}
