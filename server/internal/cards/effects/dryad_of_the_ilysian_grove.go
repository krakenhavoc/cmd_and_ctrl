package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dryad of the Ilysian Grove — Enchantment Creature — Nymph Dryad
// {2}{G}, 2/4 (EDHREC rank 305):
//
//	"You may play an additional land on each of your turns.
//	 Lands you control are every basic land type in addition to
//	 their other types."
//
// Two primitives already built for other cards, composed:
//
//   - The land-drop half is Exploration's field, `Spec.AdditionalLandPlays`
//     (see exploration.go) — one line, no engine change.
//   - The type-granting half is Urborg, Tomb of Yawgmoth's Layer-4
//     static (urborg_tomb_of_yawgmoth.go) narrowed from "each land" to
//     "lands YOU control", and widened from one basic type to all
//     five. Same idempotent append, same CR 613.8a dependency on any
//     other effect that makes something a land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "bdbde5d0-f5e4-44da-b27c-b4ad6f374cc9",
		Name:                "Dryad of the Ilysian Grove",
		Completeness:        CompletenessFull,
		AdditionalLandPlays: 1,
		Static: []game.StaticAbility{
			{
				Layer: game.Layer4Type,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.IsLand() && target.Controller == source.Controller
				},
				Apply: dryadOfTheIlysianGroveGrantAllBasicTypes,
			},
		},
	})
}

// dryadOfTheIlysianGroveGrantAllBasicTypes appends every basic land
// type not already present, the same idempotent append Urborg's
// single-type version uses.
func dryadOfTheIlysianGroveGrantAllBasicTypes(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
	for _, want := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		found := false
		for _, st := range c.Subtypes {
			if st == want {
				found = true
				break
			}
		}
		if !found {
			c.Subtypes = append(c.Subtypes, want)
		}
	}
}
