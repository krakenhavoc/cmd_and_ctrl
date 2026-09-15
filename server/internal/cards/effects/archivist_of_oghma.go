package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archivist of Oghma — "Flash. Whenever an opponent searches their
// library, you gain 1 life and draw a card."
//
// EventSearchLibrary is emitted by the search path and, separately,
// by ShuffleLibrary with Label "shuffle". Only a real search should
// trigger, so the shuffle flavour is filtered out — otherwise every
// opponent's fetch-shuffle would double-trigger.
func init() {
	Register(Spec{
		OracleID:        "08b13e1f-27ca-40a8-b5ed-88ac933d24bf",
		Name:            "Archivist of Oghma",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			On(game.EventSearchLibrary, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				if ev.Label == "shuffle" {
					return false
				}
				return ev.Actor != source.Controller && ev.Actor != ZeroUUID
			}, "Archivist of Oghma — gain 1 life and draw", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GainLife{Player: item.Controller, Amount: 1}).Apply(ctx); err != nil {
					return err
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		},
	})
}
