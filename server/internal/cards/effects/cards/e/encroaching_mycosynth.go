package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Encroaching Mycosynth — {3}{U} Artifact (ONE):
//
//	"Nonland permanents you control are artifacts in addition to their
//	 other types. The same is true for permanent spells you control and
//	 nonland permanent cards you own that aren't on the battlefield."
//
// Three layer 4 statics (CR 613.1d), one per kind of object, since the
// "you" differs: permanents are read by their controller, permanent
// spells on the stack by their controller (CR 112.3b's caster), and cards
// in every other zone by their owner (zone_layer_statics.go, Maskwood
// Nexus's shape). Lands are left out of each.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f0bb053-4967-4289-b2b1-f6309f2eca7d",
		Name:         "Encroaching Mycosynth",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer: game.Layer4Type,
				Label: "Nonland permanents you control are artifacts in addition to their other types.",
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && effects.S82OneB18NonlandPermanentType(target)
				},
				Apply: effects.S82OneB18AddArtifactType,
			},
			{
				Layer:        game.Layer4Type,
				Label:        "Permanent spells you control are artifacts in addition to their other types.",
				AffectsZones: effects.SpellsOnTheStack(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && effects.S82OneB18NonlandPermanentType(target)
				},
				Apply: effects.S82OneB18AddArtifactType,
			},
			{
				Layer:        game.Layer4Type,
				Label:        "Nonland permanent cards you own that aren't on the battlefield are artifacts in addition to their other types.",
				AffectsZones: effects.CardsOffTheBattlefield(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Owner == source.Controller && effects.S82OneB18NonlandPermanentType(target)
				},
				Apply: effects.S82OneB18AddArtifactType,
			},
		},
	})
}
