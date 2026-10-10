package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saheeli's Lattice // Mastercraft Raptor — a transforming artifact
// (#2709, ADR 0137 and its 2026-10-10 amendment):
//
//	Saheeli's Lattice — Artifact {1}{R}
//	  "When this artifact enters, you may discard a card. If you do,
//	   draw two cards.
//	   Craft with one or more Dinosaurs {4}{R}"
//	Mastercraft Raptor — Artifact Creature — Dinosaur, */4
//	  "Mastercraft Raptor's power is equal to the total power of the
//	   exiled cards used to craft it."
//
// The enters trigger is Witch's Mark's optional discard, with the draw
// as its continuation, so "if you do" is a card really discarded. The
// craft is an open count of Dinosaurs, from the battlefield and the
// graveyard mixed (CraftWithOneOrMoreSubtype). The Raptor's power is a
// characteristic-defining ability read off CR 702.167c's link: each
// material still in exile, at its power there — a `*` card's own
// characteristic-defining ability applied, per the ruling — so a
// material that leaves exile stops counting.
//
// No simplification.
const saheelisLatticeOracleID = "ef8df503-6160-4e40-a689-49e97fd56cd0"

func init() {
	const label = "Saheeli's Lattice — you may discard a card; if you do, draw two cards"
	Register(Spec{
		OracleID:     saheelisLatticeOracleID,
		Name:         "Saheeli's Lattice",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return mayDiscardACardToDrawTwo(label)(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with one or more Dinosaurs {4}{R}", "{4}{R}", CraftWithOneOrMoreSubtype("Dinosaur")),
		},
	})

	Register(Spec{
		OracleID:     saheelisLatticeOracleID + "#1",
		Name:         "Mastercraft Raptor",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = craftMaterialsTotalPower(g, source)
			},
		}},
	})
}
