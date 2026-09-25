package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scrawling Crawler — Artifact Creature — Phyrexian Construct {3},
// 3/2 (EDHREC rank 602):
//
//	"At the beginning of your upkeep, each player draws a card.
//	 Whenever an opponent draws a card, that player loses 1 life."
//
// A colourless Howling Mine with a tax. Two ordinary triggers:
//
//   - The upkeep half is "your upkeep" — EventBeginUpkeep gated on
//     the active player being the controller — and the draws are
//     real draws in APNAP order, so the Crawler's own second ability
//     (and every other "whenever an opponent draws" watcher) sees
//     each of them.
//   - The drain half fires once per card drawn (EventDrawCard is per
//     card), for any opponent's draw from any source. The drawing
//     player rides the trigger's closure as a copied ID, the posture
//     Arcane Denial's victim takes: a value, never a pointer into
//     game state.
//
// Life loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d60c80d-9a95-4051-bbf7-02ca51ee1be2",
		Name:         "Scrawling Crawler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Scrawling Crawler — each player draws a card", func(g *game.Game, item *game.StackItem) error {
				return b05EachPlayerDraws(g, item, 1)
			}),
			{
				Watches: []game.EventKind{game.EventDrawCard},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor != uuid.Nil && ev.Actor != source.Controller
				},
				Key: "Scrawling Crawler — that player loses 1 life",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Scrawling Crawler — that player loses 1 life", nil)
					item.Params.Player = ev.Actor
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					drawer := item.Params.Player
					if g.PlayerByIDForEffect(drawer) == nil {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, drawer, -1)
				},
			},
		},
	})
}
