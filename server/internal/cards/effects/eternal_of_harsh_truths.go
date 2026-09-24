package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eternal of Harsh Truths — Creature — Zombie Cleric {2}{U}, 1/3
// (EDHREC rank 17907):
//
//	"Afflict 2 (Whenever this creature becomes blocked, defending
//	 player loses 2 life.)
//	 Whenever this creature attacks and isn't blocked, draw a card."
//
// A proof card for #1279, and the one that shows both halves of a
// declaration's answer: exactly one of its two triggers fires each
// combat, and which one is decided by the defending player's COMPLETED
// block declaration. Afflict is "becomes blocked" (CR 702.130a),
// EventBecomesBlocked, emitted once per blocked attacker when the
// declaration is locked in (CR 506.4, #830); "isn't blocked" is
// EventBlockersDeclared (see Swamp Mosquito). Both are emitted by the
// same completion, afflict first, so a defender who declares no
// blocks never makes the card blocked and a defender who blocks never
// lets it draw.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "5c09d646-afea-4fd0-9752-04820b81cc5b",
		Name:         "Eternal of Harsh Truths",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBecomesBlocked},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventBecomesBlocked && ev.CardID == source.InstanceID
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					defender := ev.Actor
					return game.NewTriggeredItem(source, "Eternal of Harsh Truths — afflict 2: defending player loses 2 life",
						func(g *game.Game, item *game.StackItem) error {
							if !defendingPlayerStillIn(g, defender) {
								return nil
							}
							return g.ChangePlayerLifeForEffect(item.SourceCardID, defender, -2)
						})
				},
			},
			WhenAttacksAndIsNotBlocked("Eternal of Harsh Truths — draw a card", func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
				return g.DrawNForEffect(item.Controller, 1)
			}),
		},
	})
}
