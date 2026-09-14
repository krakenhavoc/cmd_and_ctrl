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
// Things this listener INTENTIONALLY does NOT bump on:
//   - Turn advance. Handled, but not here: S25 (#77) put the bump in
//     `onTurnAdvanceLocked` (game.go) exactly as this note used to
//     prescribe, because there is no event for a turn change to
//     listen to. Zurgo Helmsmasher's "during your turn, ~ has
//     indestructible" is the forcing function — a static whose
//     predicate reads the turn rather than the battlefield.
//     (Turn-scoped "until end of turn" effects do NOT depend on that
//     bump: `ClearExpiredTurnScopedStaticsLocked` bumps the version
//     itself when it sweeps.)
//   - Control changes. Nothing needed as of S24: Mind Control's
//     layer-2 control change is a continuous effect whose only input
//     is the attachment, and attachment bumps already via the
//     EventAttach / EventUnattach kinds below.
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
			// S18 sub-PR 2: every battlefield entry is stamped
			// unconditionally. This flag is a bare "entered this
			// turn" marker, NOT the verdict — both the creature
			// test (CR 302.6) and the haste bypass are evaluated
			// at read time by HasSummoningSickness, so a haste
			// creature attacks immediately, an artifact or land
			// is never sick at all, and a permanent that becomes
			// a creature later this turn is judged on the types
			// it has when the question is asked. Stamping here
			// unconditionally is what makes all three fall out
			// with no clear logic on this path.
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
//
// S16.5: the same walk also ENDS a copy effect (CR 400.7 — the
// permanent that left became a new object, and the copy applied to
// the permanent). A Clone that dies is a card named Clone in its
// owner's graveyard; without this it would be a second Llanowar
// Elves there, castable for {G}, and a second clone of it later
// would copy the wrong card.
func clearEffectiveCacheLocked(g *Game, cardID uuid.UUID) {
	if c := findCardInNonBattlefieldZoneLocked(g, cardID); c != nil {
		c.effective = nil
		c.restorePrintedSelf()
	}
}

// findCardInNonBattlefieldZoneLocked returns a pointer to the named
// card in whichever non-battlefield zone currently holds it, or nil.
// Split out of the leave path so the two things that happen there —
// dropping the layer cache and undoing a copy effect — read as two
// statements rather than ten copies of a zone walk.
func findCardInNonBattlefieldZoneLocked(g *Game, cardID uuid.UUID) *Card {
	for _, p := range g.Seats {
		for _, z := range []*Zone{p.Hand, p.Graveyard, p.Library, p.Command} {
			if z == nil {
				continue
			}
			for i := range z.Cards {
				if z.Cards[i].InstanceID == cardID {
					return &z.Cards[i]
				}
			}
		}
	}
	for _, z := range []*Zone{g.Exile, g.Stack} {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			if z.Cards[i].InstanceID == cardID {
				return &z.Cards[i]
			}
		}
	}
	return nil
}

// timeNowUnixNano is a thin indirection so tests can stub
// monotonic time without depending on time package internals.
// Production use of the real clock is fine; if test parallelism
// ever produces same-nanosecond entries, the stable sort in the
// layer engine resolves the tie deterministically.
var timeNowUnixNano = func() int64 { return time.Now().UnixNano() }
