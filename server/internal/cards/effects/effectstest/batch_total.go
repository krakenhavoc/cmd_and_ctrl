package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch_total_test.go — #3172 slice cav-d, ADR 0108 amendment
// 2026-10-13: "that much" / "that many" triggers read the TOTAL of what
// happened at the same time (CR 120.3, CR 510.2, CR 603.2c) instead of
// firing once per damage source, card or permanent.

// triggersFrom counts the triggered items waiting or on the stack from
// one source.
func TriggerAmountsFrom(g *game.Game, source uuid.UUID) []int {
	var out []int
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source && item.Trigger != nil {
			out = append(out, item.Trigger.Event.Amount)
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source && item.Trigger != nil {
			out = append(out, item.Trigger.Event.Amount)
		}
	}
	return out
}

func DealDamageEventsTo(g *game.Game, target uuid.UUID) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == game.EventDealDamage && ev.Target == target {
			out = append(out, ev)
		}
	}
	return out
}
