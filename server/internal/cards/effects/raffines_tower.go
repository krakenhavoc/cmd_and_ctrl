package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raffine's Tower — Land — Plains Island Swamp:
//
//	"({T}: Add {W}, {U}, or {B}.)"
//	"This land enters tapped."
//	"Cycling {3}"
//
// The Esper tri-land. Enters-tapped is a real CR 614
// self-replacement and the three-colour pipe gives one picker rather
// than three menu entries.
//
// Declared sandbox simplification: NO CYCLING. Cycling is
// "{3}, Discard this card: Draw a card", an activated ability with a
// DISCARD cost, activated FROM HAND. Two things are missing and both
// are engine-side: `game.AbilityCost` has no discard component
// (Spec.AdditionalCost's DiscardCost is a spell-cast cost, not an
// ability cost), and the CR 602 activation path only offers
// abilities on battlefield permanents. Shipping without it makes the
// Tower strictly worse than printed, which is the safe direction.
func init() {
	Register(Spec{
		OracleID:     "6e9ef5ef-6aed-4d3e-a59b-9e3dc8740b1b",
		Name:         "Raffine's Tower",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Cycling isn't implemented — the land can only be played, not cycled."},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B}",
			Label:    "Add {W}, {U}, or {B}",
		}},
	})
}
