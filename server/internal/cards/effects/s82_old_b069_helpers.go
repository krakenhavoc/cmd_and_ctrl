package effects

import (
	"sort"

	"github.com/google/uuid"
)

// s82_old_b069_helpers.go — shared body for Kaldheim batch 069.

// OldB069RemoveCounters is "remove X counters from target permanent" as an
// effect (Thornmantle Striker): the controller picks which kinds of counter
// and how many of each, X in all, or every counter the permanent has when it
// has fewer. A bot removes -1/-1 counters from its own permanent and
// everything else from another's. An on-disk key: never renamed, never reused.
var OldB069RemoveCounters = RegisterCounterChoice("number/old-b069-remove-counters",
	func(ctx *Context, target uuid.UUID, kind string, most int) int {
		mine := false
		if c, ok := ctx.Game.LookupCardForEffect(target); ok {
			mine = c.Controller == ctx.Controller()
		}
		if (kind == "-1/-1") == mine {
			return most
		}
		return 0
	},
	func(ctx *Context, target uuid.UUID, _ []uuid.UUID, picked map[string]int) error {
		kinds := make([]string, 0, len(picked))
		for kind := range picked {
			kinds = append(kinds, kind)
		}
		sort.Strings(kinds)
		for _, kind := range kinds {
			if err := (AddCounter{Target: target, Kind: kind, N: -picked[kind]}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
