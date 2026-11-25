package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gideon's Resolve — {4}{W} Enchantment (Amonkhet, S82 old batch 245):
//
//	"When this enchantment enters, you may search your library and/or graveyard
//	 for a card named Gideon, Martial Paragon, reveal it, and put it into your
//	 hand. If you search your library this way, shuffle.
//	 Creatures you control get +1/+1."
//
// The enters trigger is the War of the Spark family's shape
// (MayFetchNamedFromLibraryOrGraveyard); the static is Glorious Anthem's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "71c5718f-580d-4993-8492-81b7a557d10c",
		Name:         "Gideon's Resolve",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered:    []game.TriggeredAbility{effects.MayFetchNamedFromLibraryOrGraveyard("Gideon's Resolve", "Gideon, Martial Paragon")},
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				Label:    "Creatures you control get +1/+1.",
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
		},
	})
}
