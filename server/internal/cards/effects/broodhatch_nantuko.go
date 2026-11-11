package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Broodhatch Nantuko — Creature — Insect Druid {1}{G}, 1/1:
//
//	"Whenever this creature is dealt damage, you may create that many
//	 1/1 green Insect creature tokens.
//	 Morph {2}{G}"
//
// Hexhaven Invigorator's reading of "is dealt damage": the trigger waits
// for the damage to finish and counts the TOTAL the Nantuko was dealt
// at the same time (CR 120.3, CR 510.2). "You may" is a real decline.
func init() {
	const label = "Broodhatch Nantuko — create that many 1/1 green Insect creature tokens"
	Register(Spec{
		OracleID:         "b0a54050-3493-4ff6-8eca-e43b92c1f3d3",
		Name:             "Broodhatch Nantuko",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{2}{G}")},
		Triggered: []game.TriggeredAbility{Optional(game.TriggeredAbility{
			Watches:    []game.EventKind{game.EventDealDamage},
			Key:        label,
			AtBatchEnd: ThatMuchDamage(),
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Params.Amount <= 0 {
					return nil
				}
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 green Insect"), N: item.Params.Amount}.Apply(NewContext(g, item))
			},
		}, "Broodhatch Nantuko — create that many Insect tokens?")},
	})
}
