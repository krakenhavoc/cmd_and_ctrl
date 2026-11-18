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

	// EnchantmentOnly narrows the grant to enchantment creatures
	// (Storyweave, #5493). Additive within a snapshot version (#522).
	EnchantmentOnly bool `json:"enchantmentOnly,omitempty"`
	// Once makes it "the next time one or more creatures enter": the
	// first entry that takes it spends it, and so does every other card
	// of that same simultaneous entry (CR 614.1, 603.2c). FiredSeq is the
	// event-log position at that first use. No event is emitted between
	// the replacement windows of one simultaneous entry, so a later window
	// at the same position is the same entry; a later position is a new one.
	Once     bool   `json:"once,omitempty"`
	Fired    bool   `json:"fired,omitempty"`
	FiredSeq uint64 `json:"firedSeq,omitempty"`
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
func (g *Game) entryCountersFor(controller uuid.UUID, entering Card) []EntryCounterGrant {
	p := g.playerByIDLocked(controller)
	if p == nil {
		return nil
	}
	var out []EntryCounterGrant
	for _, s := range p.Statics {
		if s.CreaturesEnterWithCounters == nil || g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		grant := *s.CreaturesEnterWithCounters
		if grant.EnchantmentOnly && !entering.IsEnchantment() {
			continue
		}
		if grant.Once && grant.Fired && grant.FiredSeq != g.eventSeq {
			continue
		}
		out = append(out, grant)
	}
	return out
}

// spendOnceEntryCountersLocked marks every live "next time" grant that
// just applied to an entering creature as used at the current log position.
func (g *Game) spendOnceEntryCountersLocked(controller uuid.UUID, entering Card) {
	p := g.playerByIDLocked(controller)
	if p == nil {
		return
	}
	for i := range p.Statics {
		grant := p.Statics[i].CreaturesEnterWithCounters
		if grant == nil || !grant.Once || g.durationExpiredLocked(p.Statics[i].Duration, false) {
			continue
		}
		if grant.EnchantmentOnly && !entering.IsEnchantment() {
			continue
		}
		if grant.Fired && grant.FiredSeq != g.eventSeq {
			continue
		}
		// A fresh copy: clones share the pointer (clonePlayerStatics is shallow).
		spent := *grant
		spent.Fired, spent.FiredSeq = true, g.eventSeq
		p.Statics[i].CreaturesEnterWithCounters = &spent
	}
}

// GrantNextEntryCountersForEffect gives `player` "the next time one or more
// <enchantment> creatures you control enter this turn, each enters with N
// additional <kind> counters" for duration d (Storyweave, #5493).
func (g *Game) GrantNextEntryCountersForEffect(player uuid.UUID, kind string, count int, enchantmentOnly bool, label string, source uuid.UUID, d Duration) {
	p := g.playerByIDLocked(player)
	if p == nil || kind == "" || count <= 0 {
		return
	}
	p.Statics = append(p.Statics, PlayerStatic{
		CreaturesEnterWithCounters: &EntryCounterGrant{Kind: kind, Count: count, EnchantmentOnly: enchantmentOnly, Once: true},
		Source:                     source,
		Label:                      label,
		Duration:                   d,
	})
}

// playerEntryCountersReplacement is the built-in reader.
var playerEntryCountersReplacement = ReplacementEffect{
	Watches: []EventKind{EventZoneMove},
	AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
		if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield {
			return false
		}
		entering, ok := g.LookupCardForEffect(ev.CardID)
		return ok && entering.IsCreature() && len(g.entryCountersFor(entering.Controller, entering)) > 0
	},
	Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
		entering, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return nil
		}
		for _, grant := range g.entryCountersFor(entering.Controller, entering) {
			ev.AddCounterAtETB(grant.Kind, grant.Count)
		}
		g.spendOnceEntryCountersLocked(entering.Controller, entering)
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
