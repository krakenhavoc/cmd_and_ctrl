package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b182_helpers.go — shared pieces of GAP batch 182 (#5948, Starter
// Commander Decks).

// GapB182RegisterTappedDual registers a land that "enters tapped" and
// taps for either of two colours (the Guildgates' shape), one Spec for the
// five lands of this batch that print exactly those two lines.
func GapB182RegisterTappedDual(oracleID, name, a, b string) {
	Register(Spec{
		OracleID:      oracleID,
		Name:          name,
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{DualManaAbility(a, b)},
	})
}
