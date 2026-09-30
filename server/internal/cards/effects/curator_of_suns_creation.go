package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curator of Sun's Creation — Creature — Human Artificer {3}{R}, 3/3:
//
//	"Whenever you discover, discover again for the same value. This
//	 ability triggers only once each turn."
//
// "The same value" is the N of the discover that triggered it, which
// EventDiscover carries (CR 701.57b, ADR 0099). The event fires once the
// discovered card is settled — after the discovered spell is cast — so
// this trigger goes on the stack above that spell and resolves first,
// as in paper.
//
// "Only once each turn" is a check before the trigger fires, per object
// (#936): the second discover of a turn, this ability's own included,
// triggers nothing.
func init() {
	const label = "Curator of Sun's Creation — discover again for the same value"
	Register(Spec{
		OracleID:     "0e636c98-67c7-4f02-b909-5d43e276fa41",
		Name:         "Curator of Sun's Creation",
		Completeness: CompletenessFull,
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscover, AllOf(YouDiscovered, curatorFirstThisTurn(label)), label,
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					if item.Trigger != nil {
						n = item.Trigger.Event.Amount
					}
					return Discover{N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}

// curatorFirstThisTurn is "this ability triggers only once each turn".
func curatorFirstThisTurn(label string) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return source != nil && !b11TriggeredThisTurn(g, source.InstanceID, label)
	}
}
