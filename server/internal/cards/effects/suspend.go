package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// suspend.go — #659: suspend (CR 702.62).
//
// "Suspend 3—{R}" is two numbers and a card file should spell out
// exactly those two. Everything else belongs to the keyword and the
// engine carries it: the special action's timing window, the time
// counters, the exile-zone upkeep countdown, the free cast when the
// last counter comes off, and the haste on the creature that results.
//
// In particular the card file does NOT write the countdown trigger.
// `buildDef` grows it from the declaration (game.SuspendUpkeepTrigger),
// so three cards cannot spell it three ways and the fourth cannot
// forget it — the same bargain `Cycling` makes for its zone and its
// discard cost.

// Suspend is "Suspend N—<cost>" — CR 702.62a. `n` is the number of
// time counters and `cost` is the printed suspend cost.
//
// Give it to the card the way its oracle text reads:
//
//	SpecialActions: []game.SpecialAction{Suspend(1, "{R}")},   // Rift Bolt
//	SpecialActions: []game.SpecialAction{Suspend(3, "{0}")},   // Lotus Bloom
//
// A suspend cost of "{0}" is a real printed cost and not an omission
// — Lotus Bloom's is free — so it is passed as written rather than
// left empty.
func Suspend(n int, cost string) game.SpecialAction {
	return game.SpecialAction{
		Kind:     game.SpecialActionSuspend,
		Cost:     cost,
		Counters: n,
		Label:    game.SuspendLabel(n, cost),
	}
}
