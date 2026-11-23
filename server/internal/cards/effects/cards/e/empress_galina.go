package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Empress Galina — Legendary Creature — Merfolk Noble {3}{U}{U}, 3/3:
//
//	"{U}{U}, {T}: Gain control of target legendary permanent. (This effect
//	 lasts indefinitely.)"
//
// Agent of Treachery's indefinite GainControl behind a tap ability, with
// Minamo School at Water's Edge's legendary-permanent target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea2614bc-00ad-4859-98f5-689f33357fe6",
		Name:         "Empress Galina",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{U}, {T}: Gain control of target legendary permanent.",
			Cost:    effects.Plus(effects.ManaCost("{U}{U}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target legendary permanent", effects.B05Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GainControl{
					Target:     id,
					Controller: item.Controller,
					Duration:   game.IndefiniteDuration(),
					Label:      "Empress Galina — gain control (indefinitely)",
				}.Apply(ctx)
			},
		}},
	})
}
