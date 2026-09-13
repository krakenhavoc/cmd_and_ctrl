package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mindcrank — Artifact, {2} (EDHREC rank 961):
//
//	"Whenever an opponent loses life, that player mills that many
//	 cards. (Damage causes loss of life.)"
//
// The mill deck's two-drop, and half of the Duskmantle Guildmage
// loop. One trigger on the two event kinds that carry a life loss —
// a negative life change and damage to a player, which the batch 04
// helper already tells apart so nothing is counted twice — with the
// player and the amount captured in Build by value, so the mill lands
// on whoever lost the life rather than on a target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c73c1d91-0163-49c6-832a-b9327e7a2c9b",
		Name:         "Mindcrank",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b04OpponentLostLife(ev, source.Controller, g)
				return ok
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				lost, _ := b04OpponentLostLife(ev, source.Controller, g)
				victim := ev.Target
				return game.NewTriggeredItem(source, "Mindcrank — that player mills that many cards",
					func(g *game.Game, item *game.StackItem) error {
						if g.PlayerByIDForEffect(victim) == nil {
							return nil
						}
						return MillCards{Player: victim, N: lost}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
