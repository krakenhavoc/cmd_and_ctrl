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
//   - EventPlayerCounterPlaced — a poison / energy count is an
//     AppliesTo input ("corrupted") and a layer 7 input (Vishgraz)
//     (CurrentPower / CurrentToughness delegation) and Tarmogoyf-
//     style CDA inputs (graveyard-counter changes etc.).
//
// And one CONDITIONAL bump, added for #74:
//   - ANY event whose OldZone / NewZone crosses a HAND boundary — a
//     draw, a discard, a cast, "put it into your hand" — but only
//     while a permanent declaring StaticAbility.DependsOnHandSize is
//     on the battlefield. Psychosis Crawler is the card; see
//     handSizeStaticIsLiveLocked for why the condition is the whole
//     design and not an optimisation.
//
// Things this listener INTENTIONALLY does NOT bump on:
//   - Turn advance. Handled, but not here: S25 (#77) put the bump in
//     `onTurnBeganLocked` (rotation.go) exactly as this note used to
//     prescribe, because there is no event for a turn change to
//     listen to. Zurgo Helmsmasher's "during your turn, ~ has
//     indestructible" is the forcing function — a static whose
//     predicate reads the turn rather than the battlefield.
//     (Turn-scoped "until end of turn" effects do NOT depend on that
//     bump: `ClearEndOfTurnScopedStaticsLocked` bumps the version
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
	// The one CONDITIONAL bump, and the one keyed on the ZONES an
	// event names rather than on its kind. A draw is EventDrawCard, a
	// discard is EventDiscardCard and "put it into your hand" is
	// EventZoneMove; all three carry OldZone / NewZone, and the only
	// question being asked is whether a hand just changed size. A
	// kind-based switch here would have to list every current spelling
	// of "a card crossed a hand boundary" and would silently miss the
	// next one — which is exactly how the Crawler was wrong for a
	// sprint.
	//
	// Battlefield moves are excluded because the switch below bumps
	// for them already, unconditionally.
	if (ev.OldZone == ZoneHand) != (ev.NewZone == ZoneHand) &&
		ev.OldZone != ZoneBattlefield && ev.NewZone != ZoneBattlefield &&
		handSizeStaticIsLiveLocked(g) {
		g.layerVersion.Add(1)
	}
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
	case EventPlayerCounterPlaced:
		// ADR 0056 Decision 5, and the one bump on this list with NO
		// condition on it by decision rather than by omission. A
		// "corrupted" static ("as long as an opponent has three or more
		// poison counters", Skrelv's Hive) and a poison-count P/T
		// (Vishgraz) are layer inputs that nothing else invalidates:
		// before this event existed a player counter was a bare map
		// write, so the resolution went stale until an unrelated
		// permanent happened to move.
		//
		// A gate in the shape of handSizeStaticIsLiveLocked below was
		// considered and rejected: player counters change a handful of
		// times a game, so it would save nothing measurable, and it is
		// exactly the kind of gate that was wrong about Psychosis
		// Crawler for a sprint.
		g.layerVersion.Add(1)
	case EventControlChanged:
		// #990: who controls a permanent is an AppliesTo input for
		// every "creatures you control" static and for every
		// ForAsLongAs duration keyed on control (suspend's haste,
		// CR 702.62e). The pass that MOVES control cannot see its own
		// answer — materialiseControlLocked writes Card.Controller
		// after the layer walk has already run against the old one —
		// so without this bump the stale resolution survives until
		// some unrelated event invalidates it, and a creature keeps
		// the grant it just lost.
		//
		// Safe at this point in the pass for the reason
		// emitControlChangesLocked gives: the deltas are emitted
		// AFTER lastResolvedVersion is stored, so this schedules the
		// next pass rather than re-entering the current one. That
		// second pass produces no further delta and so no further
		// bump.
		g.layerVersion.Add(1)
	case EventClassLevel, EventCaseSolved:
		// ADR 0071: a designation switches printed statics on and off,
		// so a level-up or a solve changes which continuous effects
		// are in play. Charge-counter thresholds need no arm of their
		// own — EventCounterPlaced above is emitted for removals too,
		// which is exactly the pair a live "{N+}" gate needs.
		//
		// Without this the level-3 anthem would appear only when some
		// unrelated permanent happened to move, which is the same
		// staleness the chosen-creature-type bump fixes in
		// creature_type_choice.go.
		g.layerVersion.Add(1)
	case EventAttach, EventUnattach:
		// S24: attachment is an AppliesTo input for every
		// "equipped creature" / "enchanted creature" static, and
		// CR 613.7d gives the attachment a fresh timestamp when it
		// lands. Without this bump the cached resolution survives
		// the equip and the sword grants nothing until some
		// unrelated event invalidates.
		g.layerVersion.Add(1)
	case EventChangeLife:
		// The life-total twin of the hand-size bump above, and
		// conditional for the same reason: a life total is read by
		// exactly one static shape in the catalog (Aettir and
		// Priwen's base P/T), and life changes at every table in
		// every combat. With no such permanent in play this is the
		// no-op the two irrelevant-event guards assert.
		if lifeTotalStaticIsLiveLocked(g) {
			g.layerVersion.Add(1)
		}
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

// handSizeStaticIsLiveLocked reports whether any permanent on the
// battlefield declares a static ability marked DependsOnHandSize.
//
// # Why the bump this gates is conditional
//
// A hand-size CDA — "power and toughness are each equal to the number
// of cards in your hand" (Psychosis Crawler) — is the one static in
// the catalog whose input is not on the battlefield, not a counter,
// not tap state and not the turn. Every other invalidation input is
// rare; a hand change is the single most frequent thing that happens
// in a game of Magic. Putting hand moves on the unconditional bump
// list makes the recompute run after every draw, discard, cast and
// land drop at every table in the world, whether or not a card that
// cares is anywhere in play — which is why the two guards
// (TestLayerVersionDoesNotBumpOnZoneMoveOutsideBattlefield,
// TestLayerVersionDoesNotBumpOnIrrelevantEvent) exist and why they
// still pass: with no such permanent on the battlefield this is
// exactly the no-op they assert.
//
// So the question the listener asks is not "did a hand change?" but
// "did a hand change while something was reading it?", and the answer
// is a battlefield walk with a catalog lookup per permanent. That is
// the cost this pays on hand moves at a table that does have one, and
// it is bounded by the board rather than by the hand.
//
// The walk is deliberately not replaced by an incrementally
// maintained counter. A permanent's catalog key can change without
// any zone move at all — a Clone that copies Psychosis Crawler starts
// depending on hand size while sitting still — so a counter
// maintained on entry and exit would be wrong in exactly the case
// that is hardest to notice.
//
// Caller must hold g.mu (the EmitEvent path always does).
func handSizeStaticIsLiveLocked(g *Game) bool {
	return staticOnBattlefieldLocked(g, func(ab StaticAbility) bool { return ab.DependsOnHandSize })
}

// lifeTotalStaticIsLiveLocked is handSizeStaticIsLiveLocked for a
// life total. Everything the long note above says about why the bump
// is conditional, why the walk is not replaced by a maintained
// counter, and what it costs applies here word for word.
func lifeTotalStaticIsLiveLocked(g *Game) bool {
	return staticOnBattlefieldLocked(g, func(ab StaticAbility) bool { return ab.DependsOnLifeTotal })
}

// staticOnBattlefieldLocked reports whether any permanent on the
// battlefield declares a static ability the predicate accepts. The
// shared walk under the two invalidation hints, so a third hint is a
// one-line function rather than a third copy of the loop.
//
// Caller must hold g.mu.
func staticOnBattlefieldLocked(g *Game, want func(StaticAbility) bool) bool {
	if g.Battlefield == nil || CatalogStaticAbilities == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		for _, ab := range StaticAbilitiesForCard(g.Battlefield.Cards[i]) {
			if want(ab) {
				return true
			}
		}
	}
	return false
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
