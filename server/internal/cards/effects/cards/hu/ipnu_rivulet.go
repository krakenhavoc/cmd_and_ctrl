package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ipnu Rivulet — Land — Desert:
//
//	"{T}: Add {C}.
//	 {T}, Pay 1 life: Add {U}.
//	 {1}{U}, {T}, Sacrifice a Desert: Target player mills four cards."
//
// The Desert shape (cmdr2B13DesertLand): the sacrificed Desert may be
// this land or another.
//
// No simplification.
func init() {
	effects.Register(effects.Cmdr2B13DesertLand("c17d799f-adc9-4c41-87cf-b243b5ea3be1", "Ipnu Rivulet", "U", effects.ActivatedAbility{
		Label:   "{1}{U}, {T}, Sacrifice a Desert: Target player mills four cards.",
		Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost(), effects.Cmdr2B13SacrificeADesert()),
		Targets: effects.TargetPlayer("target player"),
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			victim, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return effects.MillCards{Player: victim, N: 4}.Apply(ctx)
		},
	}))
}
