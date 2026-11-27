package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b737_helpers.go — shared pieces of S82 OLD batch 737 (#5626).

// WithMoreCreaturesAsActivated narrows a player clause to "who controls
// more creatures than you do as you activate this ability" (Keeper of the
// Beasts, game.TargetSpec.MoreCreaturesAsActivated): judged on the
// announcement only, never on the resolution re-check.
func WithMoreCreaturesAsActivated(spec *game.TargetSpec) *game.TargetSpec {
	spec.MoreCreaturesAsActivated = true
	return spec
}
