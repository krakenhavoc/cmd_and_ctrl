package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coat of Arms — Artifact, {5} (EDHREC rank 940):
//
//	"Each creature gets +1/+1 for each other creature on the
//	 battlefield that shares at least one creature type with it.
//	 (For example, if two Goblin Warriors and a Goblin Shaman are on
//	 the battlefield, each gets +2/+2.)"
//
// The symmetrical tribal anthem — EVERY creature at the table,
// yours and theirs, gets the bonus, which is the whole risk of the
// card. One layer 7c static whose AppliesTo is "is a creature" and
// whose Apply counts, for the creature being computed, the other
// creatures that share a creature type with it. Both reads go
// through the effective subtypes, and that is safe inside a layer
// recompute because subtypes are settled at layer 4 and this applies
// at 7c — a Layer-4 type grant (Conspiracy, a Maskwood Nexus) counts
// exactly as a printed type does.
//
// "Creature type" is honoured rather than "subtype": the basic land
// types a Dryad Arbor carries and the artifact / enchantment
// subtypes an animated permanent keeps (Vehicle, Equipment, Treasure)
// are not creature types and are not counted.
//
// Two Coats double the bonus, as printed — each is its own static.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f7f133e-58ea-41ab-b1be-be4b400fac4c",
		Name:         "Coat of Arms",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature()
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				n := b08OtherCreaturesSharingAType(g, target)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
