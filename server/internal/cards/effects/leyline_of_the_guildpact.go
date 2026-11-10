package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline of the Guildpact — Enchantment {G/W}{G/U}{B/G}{R/G} (Murders at
// Karlov Manor, #4469):
//
//	"If this card is in your opening hand, you may begin the game with it
//	 on the battlefield.
//	 Each nonland permanent you control is all colors.
//	 Lands you control are every basic land type in addition to their
//	 other types."
//
// The opening-hand action is the other Leylines' (BeginTheGameOnTheBattlefield,
// ADR 0133). The colour line is a layer 5 set of every colour on each nonland
// permanent its controller controls, the Leyline itself included (CR 105.2c,
// 613.1e); the land line is Dryad of the Ilysian Grove's layer 4 add of all five
// basic land types, which brings each type's mana ability with it (CR 305.7)
// while the land keeps its name and its other subtypes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b2e6a51-b9e6-4848-b7d7-98f02f2b897f",
		Name:         "Leyline of the Guildpact",
		Completeness: CompletenessFull,
		OpeningHand:  BeginTheGameOnTheBattlefield(),
		Static: []game.StaticAbility{
			{
				Label: "Each nonland permanent you control is all colors.",
				Layer: game.Layer5Color,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return !target.IsLand() && target.Controller == source.Controller
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Colors = []string{"W", "U", "B", "R", "G"}
				},
			},
			{
				Label: "Lands you control are every basic land type in addition to their other types.",
				Layer: game.Layer4Type,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsLand() && target.Controller == source.Controller
				},
				Apply: dryadOfTheIlysianGroveGrantAllBasicTypes,
			},
		},
	})
}
