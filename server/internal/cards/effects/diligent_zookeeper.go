package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diligent Zookeeper — Creature — Human Citizen Ally {3}{G}, 4/4
// (Avatar: The Last Airbender, #4459):
//
//	"Each non-Human creature you control gets +1/+1 for each of its
//	 creature types, to a maximum of 10."
//
// A layer 7c modify (CR 613.4c) over every creature its controller
// controls that isn't a Human, read live each recompute. The count is of
// the creature's CREATURE types after layer 4 (CR 205.3m: a land type
// such as an earthbent Forest's is not one), each counted once; a
// creature that is every creature type (changeling, CR 702.73a) has more
// than ten and gets the maximum, +10/+10 — but it is also a Human, so
// it is excluded, and only a type-changing effect that leaves it every
// type but not a Human could reach that branch. A creature with no
// creature type gets nothing. "Human" is a word of the card's text
// (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a84a8a4-c8b0-45ce-af6f-605fcffc395b",
		Name:         "Diligent Zookeeper",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Each non-Human creature you control gets +1/+1 for each of its creature types, to a maximum of 10.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					!target.HasSubtype(source.TextSubtype("Human"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				n := diligentZookeeperCreatureTypes(c)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}

// diligentZookeeperCreatureTypes is the number of distinct creature
// types on a characteristic as the layer pass has it, to a maximum of 10.
func diligentZookeeperCreatureTypes(c *game.Characteristic) int {
	const most = 10
	if c.AllCreatureTypes {
		return most
	}
	seen := map[string]bool{}
	for _, s := range c.Subtypes {
		if t, ok := game.CanonicalCreatureType(s); ok {
			seen[t] = true
		}
	}
	return min(len(seen), most)
}
