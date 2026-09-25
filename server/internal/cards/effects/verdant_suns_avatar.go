package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Verdant Sun's Avatar — Creature — Dinosaur Avatar {5}{G}{G}, 5/5
// (EDHREC rank 2169):
//
//	"Whenever this creature or another creature you control enters,
//	 you gain life equal to that creature's toughness."
//
// The lifegain Dinosaur. One trigger for both halves — "this
// creature or another creature you control" is every creature
// entering under the controller's control, the Avatar included, so
// it gains 5 for itself. "That creature's toughness" is read as the
// trigger resolves — current toughness, counters and anthems
// included — falling back to the toughness it had when it entered if
// it has since left (CR 608.2h's last-known value), the Righteous
// Valkyrie body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ee71432-a44d-495e-86ea-d495d891c331",
		Name:         "Verdant Sun's Avatar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			Key:     "Verdant Sun's Avatar — gain life equal to its toughness",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			},
			Build:  b20GainLifeEqualToToughnessBuild("Verdant Sun's Avatar — gain life equal to its toughness"),
			Effect: b20GainLifeEqualToToughnessEffect,
		}},
	})
}
