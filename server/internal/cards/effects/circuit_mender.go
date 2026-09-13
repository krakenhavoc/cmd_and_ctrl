package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Circuit Mender — Artifact Creature — Insect {3}, 2/3 (EDHREC rank
// 2204):
//
//	"When this creature enters, you gain 2 life.
//	 When this creature leaves the battlefield, draw a card."
//
// The colourless value body. Two triggers: a mandatory ETB lifegain
// and a leaves-the-battlefield draw that fires on ANY exit — dying,
// exile, bounce, a flicker — which is the Slithermuse shape, not the
// dies-only cardDied gate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1665ca9f-176d-40f1-a4e9-42da4f1236e9",
		Name:         "Circuit Mender",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Circuit Mender — gain 2 life",
						func(g *game.Game, item *game.StackItem) error {
							return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Circuit Mender — draw a card",
						func(g *game.Game, item *game.StackItem) error {
							return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
