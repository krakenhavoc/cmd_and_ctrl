package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b738_helpers.go — shared pieces of S82 OLD batch 738 (#5626).

// WithMoreLifeThanTheChooser narrows a player clause the triggering player
// chooses (OfThatPlayersChoice) to "who has more life than they do and is
// their opponent", they being that chooser (Oath of Mages,
// game.TargetSpec.MoreLifeThanChooser).
func WithMoreLifeThanTheChooser(spec *game.TargetSpec) *game.TargetSpec {
	spec.MoreLifeThanChooser = true
	return spec
}

// WithMoreLifeAsActivated narrows a player clause to "who has more life than
// you do as you activate this ability" (Keeper of the Flame,
// game.TargetSpec.MoreLifeAsActivated): judged on the announcement only,
// never on the resolution re-check.
func WithMoreLifeAsActivated(spec *game.TargetSpec) *game.TargetSpec {
	spec.MoreLifeAsActivated = true
	return spec
}
