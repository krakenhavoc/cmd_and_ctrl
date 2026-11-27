package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b558_helpers.go — S82 set OLD, batch 558 (Coldsnap, cumulative
// upkeep).

// AgeCountersAsItLeft is the number of age counters on an ability's source:
// the count the source had as it left the battlefield when the ability is a
// trigger about that move (the trigger's last-known object, CR 603.10a), else
// live while it is on the battlefield, else last-known (CR 608.2h). "For each
// age counter on it" on a dies trigger (Jötun Owl Keeper, Arctic Nishoba) and
// "you lose 2 life for each age counter on it" (Krovikan Whispers).
func AgeCountersAsItLeft(g *game.Game, item *game.StackItem) int {
	if item.Trigger != nil && item.Trigger.Event.Kind == game.EventLTB && item.Trigger.Object != nil && item.Trigger.Object.ID == item.SourceCardID {
		return item.Trigger.Object.Counters[game.CounterAge]
	}
	return TrekB5CountersOnThis(g, item.SourceCardID, game.CounterAge)
}

// DivideAgeCountersOnSource is "X damage divided as you choose among any
// number of target creatures, where X is the number of age counters on it"
// (Magmatic Core): the total is read as the trigger is put on the stack.
var DivideAgeCountersOnSource = game.RegisterDivideAmount("age-counters-on-source",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		return game.DivideSpec{Total: B39CountersOn(g, a.Source, game.CounterAge)}
	})
