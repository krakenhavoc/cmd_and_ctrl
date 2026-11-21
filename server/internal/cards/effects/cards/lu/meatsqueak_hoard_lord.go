package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "57994ace-3dd6-407f-b465-c70b552bd466",
		Name:         "Meatsqueak, Hoard Lord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.On(game.EventLTB, effects.AnotherCreatureDied,
				"Meatsqueak, Hoard Lord — create a Food token",
				effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1}))),
		},
		Static: []game.StaticAbility{
			effects.TribalScalingAnthem(effects.TribeFilter{Tribes: []string{"Squirrel"}, YoursOnly: true},
				func(source *game.Card, g *game.Game) int {
					foods := effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype("Food") })
					return 3 * (foods / 7)
				}),
		},
	})
}
