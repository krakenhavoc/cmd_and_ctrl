package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chrome Host Seedshark — Creature — Phyrexian Shark {2}{U}, 2/4 (March
// of the Machine):
//
//	"Flying
//	 Whenever you cast a noncreature spell, incubate X, where X is that
//	 spell's mana value."
//
// The trigger goes on the stack above the spell and resolves first, so X
// is the spell's mana value with its announced X while it is still on
// the stack (incubateThatSpellsManaValue).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "837c5d0e-0a33-4f9d-ae25-68a687bdf640",
		Name:            "Chrome Host Seedshark",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(),
				"Chrome Host Seedshark — incubate X, where X is that spell's mana value",
				incubateThatSpellsManaValue),
		},
	})
}
