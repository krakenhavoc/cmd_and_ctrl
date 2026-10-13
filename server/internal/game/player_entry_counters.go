package game

import "github.com/google/uuid"

// player_entry_counters.go — "each creature you control enters with an
// additional +1/+1 counter on it" as a duration effect (Arlinn, the Pack's
// Hope's +1: until your next turn, #3172).
//
// A permanent's own entry replacement (Grumgully, the Generous) lives on
// the permanent and ends when it does. This one is created by a resolving
// ability and outlives its source, so it is a PlayerStatic payload with a
// CR 611.2 duration, read by a built-in CR 614 replacement on the
// entry event: the entering card is looked up before it arrives (a token
// whose entry window is open included), and the counters ride
// ev.AddCounterAtETB, so they go through the counter pipeline and a
// Doubling Season doubles them.

// EntryCounterGrant is one kind and count of counters an entering creature
// gets.
type EntryCounterGrant struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
}

// GrantCreaturesEnterWithCountersForEffect gives `player` the grant for
// duration d. Caller must hold g.mu.
func (g *Game) GrantCreaturesEnterWithCountersForEffect(player uuid.UUID, kind string, count int, label string, source uuid.UUID, d Duration) {
	p := g.playerByIDLocked(player)
	if p == nil || kind == "" || count <= 0 {
		return
	}
	p.Statics = append(p.Statics, PlayerStatic{
		CreaturesEnterWithCounters: &EntryCounterGrant{Kind: kind, Count: count},
		Source:                     source,
		Label:                      label,
		Duration:                   d,
	})
}

// entryCountersFor is the counters a creature entering under `controller`
// gets from live grants, summed per kind in declaration order.
func (g *Game) entryCountersFor(controller uuid.UUID) []EntryCounterGrant {
	p := g.playerByIDLocked(controller)
	if p == nil {
		return nil
	}
	var out []EntryCounterGrant
	for _, s := range p.Statics {
		if s.CreaturesEnterWithCounters == nil || g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		out = append(out, *s.CreaturesEnterWithCounters)
	}
	return out
}

// playerEntryCountersReplacement is the built-in reader.
var playerEntryCountersReplacement = ReplacementEffect{
	Watches: []EventKind{EventZoneMove},
	AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
		if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield {
			return false
		}
		entering, ok := g.LookupCardForEffect(ev.CardID)
		return ok && entering.IsCreature() && len(g.entryCountersFor(entering.Controller)) > 0
	},
	Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
		entering, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return nil
		}
		for _, grant := range g.entryCountersFor(entering.Controller) {
			ev.AddCounterAtETB(grant.Kind, grant.Count)
		}
		return nil
	},
	Controller: func(ev *ReplacementEvent, g *Game, _ *Card) uuid.UUID {
		if entering, ok := g.LookupCardForEffect(ev.CardID); ok {
			return entering.Controller
		}
		return uuid.Nil
	},
	Label: "Creatures enter with additional counters",
}
