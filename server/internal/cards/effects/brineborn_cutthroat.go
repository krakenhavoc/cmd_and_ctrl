package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brineborn Cutthroat — "Flash. Whenever you cast a spell during an
// opponent's turn, put a +1/+1 counter on this creature."
//
// The flash payoff: every instant-speed play on someone else's turn
// grows it. "During an opponent's turn" is read off the turn cursor
// at trigger time — the active seat's player is not the caster.
func init() {
	Register(Spec{
		OracleID:        "916cb70f-3b06-48ed-972d-75f805aa0892",
		Name:            "Brineborn Cutthroat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				return !isActivePlayer(g, source.Controller)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Brineborn Cutthroat — +1/+1 counter",
					func(g *game.Game, item *game.StackItem) error {
						return AddCounter{
							Target: item.SourceCardID,
							Kind:   "+1/+1",
							N:      1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
