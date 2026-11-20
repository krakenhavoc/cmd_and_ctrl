package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prismatic Geoscope — Artifact {5} (Commander 2016, #5537):
//
//	"This artifact enters tapped.
//	 Domain — {T}: Add X mana in any combination of colors, where X is
//	 the number of basic land types among lands you control."
//
// Axebane Guardian's any-combination mana ability with DomainCount as X:
// the basic land types among the lands its controller controls as it is
// activated (a dual with two types counts both). With no basic land type
// it adds nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be838f57-4e74-44ad-aee1-506e63b0ff37",
		Name:         "Prismatic Geoscope",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedAnyCombinationOfColors(prismaticGeoscopeX),
			Label:          "Domain — {T}: Add X mana in any combination of colors, where X is the number of basic land types among lands you control",
		}},
	})
}

func prismaticGeoscopeX(g *game.Game, controller, _ uuid.UUID) int {
	return DomainCount(g, controller)
}
