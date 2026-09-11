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
//   - Nothing, as of S24. Both entries this list used to carry —
//     aura attachment and Mind Control's layer-2 control change —
//     landed in S24. Attachment got its own EventAttach /
//     EventUnattach kinds below; the control change needed no event
//     of its own, because it is a continuous effect whose only input
//     is the attachment that already bumps.
//
// EventETB and EventLTB are also covered by EventZoneMove for every
// CARD (every zone change emits both), so the listener doesn't
// double-bump on those. A TOKEN is the exception: CreateTokenForEffect
// puts it on the battlefield from nowhere and emits EventTokenCreated
// + EventETB with no ZoneMove at all, so the listener also handles
// EventTokenCreated as an entry — bump and stamp — or a Goblin made
// under Glorious Anthem stays 1/1 until something unrelated
// invalidates the cache, and its CR 613 timestamp is never set.
// (Found by the roadmap's batch 01: Storm-Kiln Artist counting the
// Treasure its own trigger made.)

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
	case EventTokenCreated:
		g.layerVersion.Add(1)
		stampBattlefieldEntryLocked(g, ev.CardID)
	case EventCounterPlaced:
		g.layerVersion.Add(1)
	case EventAttach, EventUnattach:
		// S24: attachment is an AppliesTo input for every
		// "equipped creature" / "enchanted creature" static, and
		// CR 613.7d gives the attachment a fresh timestamp when it
		// lands. Without this bump the cached resolution survives
		// the equip and the sword grants nothing until some
		// unrelated event invalidates.
		g.layerVersion.Add(1)
	case EventTapCard, EventUntapCard:
		// Tap state is an AppliesTo input, not just a display flag:
		// The Wandering Rescuer grants hexproof to "other TAPPED
		// creatures you control", so a creature that taps or untaps
		// changes which permanents its static covers. Without this
		// bump the cached resolution survives the tap and the grant
		// appears or disappears only when some unrelated event
		// happens to invalidate — which is how the S22 card looked
		// half-working even once the keyword table honoured it.
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
			// S18 sub-PR 2: every battlefield entry earns
			// summoning sickness unconditionally. Haste bypass
			// is evaluated at read time by HasSummoningSickness
			// so a haste creature is attackable immediately this
			// same turn without extra clear logic here.
			g.Battlefield.Cards[i].SummonedThisTurn = true
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
