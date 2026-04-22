package game

import (
	"time"

	"github.com/google/uuid"
)

// layer_listener.go ships the built-in `Listener` that drives the S16
// continuous-effect engine's invalidation. Auto-registered by
// NewGame so every game starts with the layer system listening for
// the events that change which static abilities are active or what
// they apply to.
//
// Today (sub-PR 2) the listener bumps `g.layerVersion` on:
//   - EventZoneMove where either OldZone or NewZone is the
//     battlefield — covers ETB and LTB universally without having
//     to audit every emission site individually.
//   - EventCounterPlaced — counters change layer 7d inputs
//     (CurrentPower / CurrentToughness delegation) and Tarmogoyf-
//     style CDA inputs (graveyard-counter changes etc.).
//
// Things this listener INTENTIONALLY does NOT bump on yet:
//   - Step / phase advance. "Until end of turn" continuous effects
//     are out of scope for S16; when they arrive in a later sprint,
//     advance the version inside the step-advance helper directly
//     (no event for it today; cleanest is a direct call rather than
//     a new event kind for one consumer).
//   - Control changes. Mind Control / aura attach is deferred to
//     S17. When the first "creatures you control" predicate that
//     can flip mid-game lands, add an EventControlChanged kind +
//     bump here.
//
// EventETB and EventLTB are also covered by EventZoneMove (every
// zone change emits both), so the listener doesn't double-bump on
// those — the ZoneMove path is the single source of truth.

// layerVersionBump is the `Listener` that auto-installs into every
// game to invalidate the layer engine's cached resolution when
// events occur that could change static-ability active-set or
// applies-to inputs.
type layerVersionBump struct{}

// OnEvent fires synchronously inside the EmitEvent path under the
// game's write lock. Bumps the layer version on relevant events,
// and on battlefield zone moves additionally maintains per-card
// engine state: stamps EnteredBattlefieldAt on entry (drives CR
// 613 timestamp ordering) and clears the cached effective
// characteristic on exit (so the next recompute re-derives from
// printed without leaking stale state from a prior battlefield
// stay).
func (layerVersionBump) OnEvent(g *Game, ev Event) {
	switch ev.Kind {
	case EventZoneMove:
		if ev.OldZone == ZoneBattlefield || ev.NewZone == ZoneBattlefield {
			g.layerVersion.Add(1)
		}
		if ev.NewZone == ZoneBattlefield {
			stampBattlefieldEntryLocked(g, ev.CardID)
		}
		if ev.OldZone == ZoneBattlefield {
			clearEffectiveCacheLocked(g, ev.CardID)
		}
	case EventCounterPlaced:
		g.layerVersion.Add(1)
	}
}

// stampBattlefieldEntryLocked sets EnteredBattlefieldAt on the
// named card if it currently lives on the battlefield. Caller must
// hold g.mu (the EmitEvent path always does). Stamps with
// monotonic time.Now() in nanoseconds — uniqueness across same-
// instant entries is not guaranteed, but the layer-engine sort is
// stable so ties resolve to insertion order, which is good enough
// for sandbox semantics.
func stampBattlefieldEntryLocked(g *Game, cardID uuid.UUID) {
	if g.Battlefield == nil {
		return
	}
	now := timeNowUnixNano()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			g.Battlefield.Cards[i].EnteredBattlefieldAt = now
			// New entry => stale effective; let the next recompute
			// rebuild from the fresh printed baseline.
			g.Battlefield.Cards[i].effective = nil
			return
		}
	}
}

// clearEffectiveCacheLocked is symmetric to the entry stamp: when
// a card leaves the battlefield, its effective cache is no longer
// reachable through the battlefield iteration. The card slice has
// already moved the value into a different zone, so we walk every
// zone once defensively to clear any straggling effective on the
// in-flight moved card. Cheap — most Card values aren't on the
// battlefield and the zones are small.
func clearEffectiveCacheLocked(g *Game, cardID uuid.UUID) {
	for _, p := range g.Seats {
		for i := range p.Hand.Cards {
			if p.Hand.Cards[i].InstanceID == cardID {
				p.Hand.Cards[i].effective = nil
				return
			}
		}
		for i := range p.Graveyard.Cards {
			if p.Graveyard.Cards[i].InstanceID == cardID {
				p.Graveyard.Cards[i].effective = nil
				return
			}
		}
		for i := range p.Library.Cards {
			if p.Library.Cards[i].InstanceID == cardID {
				p.Library.Cards[i].effective = nil
				return
			}
		}
		for i := range p.Command.Cards {
			if p.Command.Cards[i].InstanceID == cardID {
				p.Command.Cards[i].effective = nil
				return
			}
		}
	}
	if g.Exile != nil {
		for i := range g.Exile.Cards {
			if g.Exile.Cards[i].InstanceID == cardID {
				g.Exile.Cards[i].effective = nil
				return
			}
		}
	}
}

// timeNowUnixNano is a thin indirection so tests can stub
// monotonic time without depending on time package internals.
// Production use of the real clock is fine; if test parallelism
// ever produces same-nanosecond entries, the stable sort in the
// layer engine resolves the tie deterministically.
var timeNowUnixNano = func() int64 { return time.Now().UnixNano() }
