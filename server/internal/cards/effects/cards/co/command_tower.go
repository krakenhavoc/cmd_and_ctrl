package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Command Tower — Land:
//
//	"{T}: Add one mana of any color in your commander's color
//	identity."
//
// The single most-played card in Commander, and free to express:
// the identity narrowing already exists for Arcane Signet, so this
// is the same pipe set with the same NarrowToCommanderIdentity
// filter, on a land.
//
// A land with a catalog entry declaring ManaAbilities bypasses the
// synthetic basic-land ability the engine derives from TypeLine —
// which is what we want here, since "Land" alone would otherwise
// produce nothing.
//
// CR 903.4f (#844): with no commander, or a commander whose colour
// identity is colourless (Kozilek), the quality is undefined or empty
// and this land adds no mana at all. The engine handles that centrally
// — the ability is simply not offered — so nothing about it is written
// per card here.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0895c9b7-ae7d-4bb3-af17-3b75deb50a25",
		Name:         "Command Tower",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color in your commander's color identity",
			// The printed text asks for the narrowing (manaPickOptionsFor).
			NarrowToCommanderIdentity: true,
		}},
	})
}
