package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unwinding Clock — Artifact {4} (EDHREC rank 547):
//
//	"Untap all artifacts you control during each other player's
//	 untap step."
//
// Seedborn Muse in an artifact deck, and the second card the
// roadmap's untap-step seam was holding. Same permission, narrowed
// to artifacts — including the Clock itself, which never taps, and
// including a permanent that is an artifact only because something
// else made it one: the predicate reads the effective type line, so
// March of the Machines' animated Sol Ring is an artifact creature
// and untaps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "153fac93-5d2b-4348-a468-a5eef6a12da3",
		Name:         "Unwinding Clock",
		Completeness: CompletenessFull,
		UntapStep: []game.UntapStepPermission{
			untapDuringEachOtherPlayersUntapStep(
				"Unwinding Clock — untap all artifacts you control",
				func(c game.Card) bool { return c.IsArtifact() }),
		},
	})
}
