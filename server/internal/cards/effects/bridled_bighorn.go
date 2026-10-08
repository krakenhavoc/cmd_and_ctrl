package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bridled Bighorn — Creature — Sheep Mount {3}{W}:
//
//	"Vigilance
//	 Whenever this creature attacks while saddled, create a 1/1 white
//	 Sheep creature token.
//	 Saddle 2"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6a1ad848-6fda-46dd-aad8-2f6a5c71a696",
		Name:            "Bridled Bighorn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []ActivatedAbility{Saddle(2)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Bridled Bighorn — create a 1/1 white Sheep", func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Sheep"), N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
