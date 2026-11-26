package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yew Spirit — Creature — Spirit Treefolk {4}{G}, 3/3:
//
//	"{2}{G}{G}: This creature gets +X/+X until end of turn, where X is its
//	 power."
//
// X is read once, at resolution (CR 611.2c): activated twice in a row, the
// second activation sees the power the first one left.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0333cb81-5f44-4f31-9ee7-725e4f6af7b3",
		Name:         "Yew Spirit",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}{G}: This creature gets +X/+X until end of turn, where X is its power.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{G}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				x := c.PowerForComparison()
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Toughness: x, Label: "Yew Spirit — +X/+X"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
