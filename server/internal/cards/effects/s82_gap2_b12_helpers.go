package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap2_b12_helpers.go — shared pieces of S82 GAP2 batch 12 (#6100).

// WithMoreCreaturesThanTheChooser narrows a player clause the triggering
// player chooses (OfThatPlayersChoice) to "who controls more creatures than
// they do and is their opponent", they being that chooser (Oath of Druids,
// game.TargetSpec.MoreCreaturesThanChooser).
func WithMoreCreaturesThanTheChooser(spec *game.TargetSpec) *game.TargetSpec {
	spec.MoreCreaturesThanChooser = true
	return spec
}

// WithEqualToughness is the set rule "with equal toughness": every pick of
// the clause has the same toughness after the layers (V.A.T.S.,
// game.TargetShareToughness).
func WithEqualToughness() *game.TargetSameness {
	return &game.TargetSameness{
		Label: "have equal toughness",
		Key:   game.TargetShareToughness,
	}
}
