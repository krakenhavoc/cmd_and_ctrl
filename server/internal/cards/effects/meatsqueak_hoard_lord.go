package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meatsqueak, Hoard Lord — {2}{G} Legendary Creature — Squirrel 2/3
// (Mystery Booster Commander, S82 GAP batch 3, #5948):
//
//	"Whenever another creature dies, create a Food token. This ability
//	 triggers only once each turn.
//	 For every seven Foods you control, Squirrels you control get +3/+3."
//
// The trigger is any other creature's death, anyone's, held to once each
// turn (TriggersOnlyOnceEachTurn). The static is the scaling tribal
// anthem: three times the number of complete sevens among the Foods
// (token or card) its controller controls, read on every recompute, and
// it reaches Meatsqueak herself, who is a Squirrel.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57994ace-3dd6-407f-b465-c70b552bd466",
		Name:         "Meatsqueak, Hoard Lord",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(On(game.EventLTB, AnotherCreatureDied,
				"Meatsqueak, Hoard Lord — create a Food token",
				Do(CreateToken{Template: FoodToken(), N: 1}))),
		},
		Static: []game.StaticAbility{
			TribalScalingAnthem(TribeFilter{Tribes: []string{"Squirrel"}, YoursOnly: true},
				func(source *game.Card, g *game.Game) int {
					foods := permanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype("Food") })
					return 3 * (foods / 7)
				}),
		},
	})
}
