package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seraphic Steed — Creature — Unicorn Mount {G}{W}:
//
//	"First strike, lifelink
//	 Whenever this creature attacks while saddled, create a 3/3 white
//	 Angel creature token with flying.
//	 Saddle 4"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "51622e89-aa5e-4dd6-a078-eb45e8ebb434",
		Name:            "Seraphic Steed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "lifelink"},
		Activated:       []ActivatedAbility{Saddle(4)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Seraphic Steed — create a 3/3 Angel with flying", func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("3/3 white Angel with flying"), N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
