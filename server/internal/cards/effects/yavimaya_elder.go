package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yavimaya Elder — Creature — Human Druid {1}{G}{G}, 2/1 (EDHREC
// rank 3563):
//
//	"When this creature dies, you may search your library for up to
//	 two basic land cards, reveal them, put them into your hand, then
//	 shuffle.
//	 {2}, Sacrifice this creature: Draw a card."
//
// The three-for-one. The dies trigger is optional ("you may"),
// answered when it fires; the search prompt is where the controller
// picks up to two basics, revealed, to hand. The activated ability
// pays the sacrifice as a cost, so the dies trigger lands on the
// stack ABOVE the ability and resolves first: lands, then the card —
// which is how the Elder is played on paper. Basic is read off the
// supertype, so a Snow-Covered Forest is offered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fd4c452-07f2-492c-9c78-d1c6362d9eec",
		Name:         "Yavimaya Elder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Yavimaya Elder — search your library for up to two basic land cards?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Yavimaya Elder — search for up to two basic lands", b34SearchUpToTwoBasicsToHand)
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, Sacrifice this creature: Draw a card.",
			Cost:  Plus(ManaCost("{2}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
