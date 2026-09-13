package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Shire — Legendary Land (EDHREC rank 1165):
//
//	"The Shire enters tapped unless you control a legendary creature.
//	 {T}: Add {G}.
//	 {1}{G}, {T}, Tap an untapped creature you control: Create a Food
//	 token."
//
// Mines of Moria's green sibling: the same "unless you control a
// legendary creature" entry, read post-layer so an animated legendary
// artifact counts, and a plain {G}.
//
// DECLARED SIMPLIFICATION: the Food ability is not implemented. "Tap
// an untapped creature you control" is a cost component the engine
// cannot express — AbilityCost carries tap-this, sacrifice, mana,
// life, loyalty and crew, and nothing that taps ANOTHER permanent
// (the Springleaf Drum / Relic of Legends gap) — and shipping the
// ability with that cost omitted would make the card STRONGER than
// printed (#259). The land is still recognisably itself without it
// (a conditional-untapped Forest for a legends deck), which is the
// Mines of Moria posture, and the gap runs the weaker way. The
// ability lands when a tap-another-creature cost component does.
func init() {
	Register(Spec{
		OracleID:     "9abf9a0e-8e7d-406b-a01d-d4870b30134e",
		Name:         "The Shire",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Food-making ability isn't implemented — the land can only be played and tapped for {G}."},
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(b08ControlsLegendaryCreature)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
