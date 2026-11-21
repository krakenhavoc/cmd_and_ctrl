package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seraphic Steed — Creature — Unicorn Mount {G}{W}:
//
//	"First strike, lifelink
//	 Whenever this creature attacks while saddled, create a 3/3 white
//	 Angel creature token with flying.
//	 Saddle 4"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "51622e89-aa5e-4dd6-a078-eb45e8ebb434",
		Name:            "Seraphic Steed",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "lifelink"},
		Activated:       []effects.ActivatedAbility{effects.Saddle(4)},
		Triggered: []game.TriggeredAbility{
			effects.AttacksWhileSaddled("Seraphic Steed — create a 3/3 Angel with flying", func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("3/3 white Angel with flying"), N: 1}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
