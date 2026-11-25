package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b247_helpers.go — shared pieces of Amonkhet batch 247 (#5535).

// WhenFirstTargetedEachTurnCounterIt is "Whenever this creature becomes the
// target of a spell or ability for the first time each turn, counter that spell
// or ability" (Kira, Great Glass-Spinner; Glyph Keeper): it counts the triggers
// the source has had this turn under `label` (the "first time" of CR 603.4's
// family), and counters the stack item that targeted it.
func WhenFirstTargetedEachTurnCounterIt(label string) game.TriggeredAbility {
	return On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return Self(ev, source, lki, g) && g.TriggeredThisTurn(source.InstanceID, label) < 1
	}, label, func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		return CounterTarget{StackID: item.Trigger.Event.StackItemID}.Apply(NewContext(g, item))
	})
}
