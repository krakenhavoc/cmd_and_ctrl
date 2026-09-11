package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coat of Arms — Artifact for {5}:
//
//	"Each creature gets +1/+1 for each other creature on the
//	 battlefield that shares at least one creature type with it."
//
// EACH creature, not each creature you control — Coat of Arms is
// symmetrical and famously loses games. The bonus is counted per
// creature, so a board of three Goblins gives each of them +2/+2, and
// a Goblin Warrior next to a Goblin Shaman and an Elf Warrior gets
// +2/+2 (one shared Goblin, one shared Warrior, two creatures).
//
// This is quadratic by construction: the Apply walks the battlefield
// once per affected creature, so a board of N creatures costs N²
// shared-type tests per recompute pass. At Commander board sizes that
// is a few thousand map lookups behind a cached resolution, which is
// cheap enough not to optimise and expensive enough to mention.
//
// CHANGELINGS ARE THE TRAP, and SharesCreatureType is where it is
// handled rather than here. A changeling shares a type with every
// creature that has one, INCLUDING another changeling — but not with
// a creature printed with no creature type at all, because there is
// no type for them to share. Counting "does the other creature have
// any subtype" instead would wrongly include Dryad Arbor's Forest;
// counting a raw subtype intersection would wrongly have two Forests
// share a creature type. Both of those are in the shared helper's
// test, not in this file, because the next card to ask the question
// must get the same answer.
func init() {
	Register(Spec{
		OracleID: "5f7f133e-58ea-41ab-b1be-be4b400fac4c",
		Name:     "Coat of Arms",
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature()
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				n := sharedCreatureTypeCount(g, target)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}

// sharedCreatureTypeCount counts the OTHER creatures on the
// battlefield that share at least one creature type with `target`.
//
// Iterates g.Battlefield.Cards directly rather than through
// BattlefieldCardsForEffect: this runs inside a layer recompute pass,
// where the caller may hold only the read lock, and the copy that
// helper makes would allocate a full battlefield slice once per
// creature per pass.
//
// Reading the LIVE cards is also what makes the count correct. Each
// candidate's effective characteristic is the one the earlier layers
// have already written for this pass, so a creature Maskwood Nexus
// made every type counts as sharing, and a creature an Adaptive
// Automaton typed counts too — layer 4 ran before layer 7.
func sharedCreatureTypeCount(g *game.Game, target *game.Card) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		other := &g.Battlefield.Cards[i]
		if other.InstanceID == target.InstanceID || !other.IsCreature() {
			continue
		}
		if game.SharesCreatureType(target, other) {
			n++
		}
	}
	return n
}
