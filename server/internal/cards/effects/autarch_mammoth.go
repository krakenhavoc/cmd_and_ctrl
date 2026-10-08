package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Autarch Mammoth — Creature — Elephant Mount {4}{G}{G}:
//
//	"When this creature enters and whenever it attacks while saddled,
//	 create a 3/3 green Elephant creature token.
//	 Saddle 5"
//
// One printed ability with two conditions, so two declarations sharing
// one body.
//
// No simplification.
func init() {
	const label = "Autarch Mammoth — create a 3/3 green Elephant"
	makeElephant := func(g *game.Game, item *game.StackItem) error {
		return CreateToken{Controller: item.Controller, Template: TokenCard("3/3 green Elephant"), N: 1}.Apply(NewContext(g, item))
	}
	Register(Spec{
		OracleID:     "eb50b5f2-4840-4f90-9c85-40fc581be9c7",
		Name:         "Autarch Mammoth",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(5)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, makeElephant),
			AttacksWhileSaddled(label, makeElephant),
		},
	})
}
