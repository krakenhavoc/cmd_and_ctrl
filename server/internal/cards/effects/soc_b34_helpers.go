package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// soc_b34_helpers.go — helpers for Secrets of Strixhaven batch 34 (#4456).

// socB34StunCountersThreeMinusX is "this creature enters with a number of
// stun counters on it equal to three minus X" (Slumbering Trudge): a CR
// 614.1c clause over the announced X, floored at zero (a negative amount puts
// none on). A permanent that was not cast has an X of 0 and enters with three.
func socB34StunCountersThreeMinusX() game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind: game.CounterStun,
		Count: func(cast game.CastCounts) int {
			if cast.X >= 3 {
				return 0
			}
			return 3 - cast.X
		},
	}
}

// socB34EntersTappedIfXAtMost is "if X is N or less, this creature enters
// tapped": a self-replacement (CR 614.1c) that reads the X the permanent's
// own spell was cast with (EntryCastCountsForEffect, CR 107.3m). It is
// SelfEntersTapped's shape with the condition read off the cast; an
// unconditional SelfEntersTapped would be wrong for X above N.
func socB34EntersTappedIfXAtMost(n int) game.ReplacementEffect {
	return game.ReplacementEffect{
		RowLabel:  "Enters tapped if X is 2 or less.",
		EntryOnly: true,
		Watches:   []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID &&
				g.EntryCastCountsForEffect(ev).X <= n
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = true
			return nil
		},
	}
}
