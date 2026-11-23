package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gene Pollinator — Artifact Creature — Robot Insect {G}, 1/2 (Edge of
// Eternities, #4461):
//
//	"{T}, Tap an untapped permanent you control: Add one mana of any
//	 color."
//
// A mana ability (CR 605.1a) with Baylen, the Haymaker's tap-others cost
// component beside the {T} symbol. The Pollinator is tapped by {T}, so
// the other permanent is a different one (TapAnotherUntapped's
// ExcludeSource); any untapped permanent you control pays it, a land or
// a creature that arrived this turn included, since tapping it is not
// the {T} symbol (CR 302.6). The Pollinator itself needs to have been
// under your control since your turn began to pay {T}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a356d04-2e86-435b-9766-27b93b468a61",
		Name:         "Gene Pollinator",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, TapOthers: effects.TapAnotherUntapped("an untapped permanent you control").TapOthers},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
