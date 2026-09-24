package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thunderbreak Regent — Creature — Dragon {2}{R}{R}, 4/4 (EDHREC rank
// 3028):
//
//	"Flying
//	 Whenever a Dragon you control becomes the target of a spell or
//	 ability an opponent controls, this creature deals 3 damage to
//	 that player."
//
// The Dragon tax. Flying rides PrintedKeywords. The trigger watches
// EventBecomesTarget — emitted once per target SLOT (CR 115.3) at
// announce, with the targeting player in Actor — for a Dragon the
// controller controls, the Regent itself included, targeted by an
// opponent (b28DragonYouControlTargetedByOpponent). It goes on the
// stack ABOVE the spell that targeted and resolves first: the Regent
// deals 3 to that player, red noncombat damage from a creature. A
// Regent that has left the battlefield by then still deals it —
// the damage path does not need the source in play.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bb04f927-0348-4a14-9e74-381f18083f6a",
		Name:            "Thunderbreak Regent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBecomesTarget},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b28DragonYouControlTargetedByOpponent(ev, source, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				player := ev.Actor
				return game.NewTriggeredItem(source, "Thunderbreak Regent — 3 damage to the player who targeted your Dragon",
					func(g *game.Game, item *game.StackItem) error {
						return DealDamage{Source: item.SourceCardID, Target: player, Amount: 3}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
