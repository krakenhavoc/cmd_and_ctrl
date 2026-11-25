package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conifer Wurm — Snow Creature — Wurm {4}{G}, 4/4 (Modern Horizons):
//
//	"Trample
//	 {3}{G}: This creature gets +X/+X until end of turn, where X is the number
//	 of snow permanents you control."
//
// Trample is a printed keyword. X is the count of snow permanents you control
// as the ability resolves (CR 608.2c), the Wurm itself among them since it is
// snow. It applies only to this permanent while it is still the same object
// (CR 400.7).
//
// No simplification.
func init() {
	const name = "Conifer Wurm"
	effects.Register(effects.Spec{
		OracleID:        "8f780497-23bb-44c3-982c-e2b8a7a29999",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: This creature gets +X/+X until end of turn, where X is the number of snow permanents you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				x := effects.CountControlled(g, ctx.Controller(), func(c game.Card) bool { return c.HasSupertype("snow") })
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Toughness: x, Label: name + " — +X/+X"}.Apply(ctx)
			},
		}},
	})
}
