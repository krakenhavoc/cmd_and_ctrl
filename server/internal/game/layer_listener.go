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
// And two CONDITIONAL bumps, the first added for #74 and the second
// for #1117 in the same mould:
//   - ANY event whose OldZone / NewZone crosses a HAND boundary — a
//     draw, a discard, a cast, "put it into your hand" — but only
//     while a permanent declaring StaticAbility.DependsOnHandSize is
//     on the battlefield. Psychosis Crawler is the card; see
//     handSizeStaticIsLiveLocked for why the condition is the whole
//     design and not an optimisation.
//   - A LIFE TOTAL that just changed, while something declares
//     DependsOnLifeTotal. Serra Ascendant is the card. That one is
//     NOT driven from this switch alone — see
//     invalidateLayersForLifeChangeLocked below, and the note on the
//     EventChangeLife arm — because the damage path emits its event
//     on one side of the write on one route and the other side on
//     the other.
//
// A third CONDITIONAL bump, added for #1218 in the same mould as the
// hand and life ones above:
//   - EventAttack, while something declares
//     StaticAbility.DependsOnAttackingStatus. Ohran Frostfang is the
//     card. Attacking status has two other exits this listener cannot
//     reach from an event switch — a control change pulling a
//     permanent out of combat (CR 506.4) and combat ending (CR
//     511.3) — because neither emits an event naming every affected
//     card; see invalidateLayersForAttackChangeLocked, called
//     directly from removeFromCombatLocked and clearCombatLocked.
//
// A fourth CONDITIONAL bump, added for #1325 in the same mould:
//   - EventCast, while something declares
//     StaticAbility.DependsOnSpellsCast. Stoic Sphinx's "hexproof as
//     long as you haven't cast a spell this turn" is the card. The
//     tally the condition reads (Game.SpellsCastThisTurn) is bumped
//     BEFORE EventCast fires (mutations.go), so the listener sees the
//     count this spell just added. The turn-advance exit that resets
//     the tally back to zero needs no arm here — S25's unconditional
//     bump in onTurnBeganLocked already covers every turn change.
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
	// #1117's graveyard half, and the second bump keyed on the ZONES
	// rather than the kind — a mill is an EventZoneMove, a discard is
	// an EventDiscardCard, a spell finishing is neither on every
	// route, and the only question being asked is whether some
	// graveyard just changed. The XOR catches a card LEAVING a
	// graveyard too (a graveyard exiled, a card reanimated out of
	// one), which shrinks a Lhurgoyf exactly as an arrival grows it.
	//
	// Battlefield moves are excluded because the switch below bumps
	// for them already — which is why this gap was hard to see at
	// all: a creature DYING refreshed every graveyard count on the
	// board, and only the mill, the discard and the resolving spell
	// did not.
	//
	// # Why this one is NOT gated on a declared flag
	//
	// The hand-size bump above is, and this deliberately breaks the
	// symmetry. Two reasons, both about the population rather than
	// about the cost:
	//
	// A hand-size CDA is ONE card (Psychosis Crawler). A static that
	// reads a graveyard is a whole family that was already in the
	// catalog before this bump existed — Tarmogoyf, Lord of
	// Extinction, Nighthowler, Nighthawk Scavenger, Consuming
	// Aberration, Jarad, Multani, Wight of the Reliquary, Elvish
	// Reclaimer, The Warring Triad's layer-4 clause — and several of
	// them are built by SHARED helpers, so an opt-in field would have
	// to be threaded through functions other cards call. Worse, there
	// is no test that can catch a Lhurgoyf that forgets to declare
	// it: a missing flag is a card that is silently one mill behind,
	// which is precisely the failure this issue is about. An
	// unconditional bump cannot be forgotten.
	//
	// And the frequency argument that justifies the hand gate does
	// not transfer. A hand changes on every draw, every cast and
	// every land drop; a graveyard changes when a spell resolves, a
	// card is milled or something is discarded, which is a smaller
	// number and is already the same order as the battlefield moves
	// this switch bumps for unconditionally.
	//
	// Measured, at a 40-permanent board:
	// BenchmarkGraveyardArrivalOnABoard is 6.5us/op against the
	// pre-existing BenchmarkHandMoveWithNoHandSizeCDA's 7.9us — the
	// bump itself is one atomic add and does not show. What it really
	// costs is one extra layer recompute at the NEXT read, and only
	// when a read falls between this and the next bump:
	// BenchmarkReadSnapshotStale, 46us, against a fresh read's 11ns.
	// A turn cycle with twenty graveyard arrivals is therefore under
	// a millisecond, in exchange for ten catalog cards that were
	// wrong and could not have declared anything.
	if (ev.OldZone == ZoneGraveyard) != (ev.NewZone == ZoneGraveyard) &&
		ev.OldZone != ZoneBattlefield && ev.NewZone != ZoneBattlefield {
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
	case EventTransform:
		// ADR 0079 / CR 712.18: a transform is the one mutation that
		// changes a permanent's PRINTED characteristics wholesale —
		// name, type line, colours, base P/T, printed keywords and the
		// whole catalog entry move at once — without the permanent
		// going anywhere. So it is the one invalidation input that no
		// zone move, counter, attach or tap can stand in for, and
		// without this arm a transformed Storm the Vault would keep
		// reporting "Legendary Enchantment" until something unrelated
		// happened to bump the version.
		//
		// The card's own effective cache is nilled at the mutation
		// site (TransformPermanentForEffect) rather than here, for the
		// window between the two: EmitEvent dispatches synchronously,
		// but a listener earlier in the slice than this one would
		// otherwise read the stale characteristic off the card it was
		// just told about.
		g.layerVersion.Add(1)
	case EventPhaseOut, EventPhaseIn:
		// ADR 0084 / CR 702.26: what is on the battlefield has just
		// changed, which is the same invalidation input a zone move
		// is — every "creatures you control get +1/+1", every
		// AppliesTo and every ForAsLongAs condition has a new answer.
		// The EventZoneMove arm above cannot stand in for it, because
		// phasing deliberately emits no zone move (CR 702.26d).
		//
		// The card's own effective cache is nilled at the mutation
		// site (phaseOutLocked / phaseInLocked) rather than here, for
		// the window the transform arm explains — and because by the
		// time this runs the card is in a different slice than the one
		// a cache-clearing helper would look in.
		g.layerVersion.Add(1)
	case EventTurnedFaceUp:
		// ADR 0082 / CR 708.6: the twin of the transform arm above,
		// and for the same reason. A permanent turning face up
		// changes its PRINTED characteristics wholesale — the CR
		// 708.2 body (a nameless 2/2 with no text) is replaced by the
		// real card at layer 0 — without the permanent going
		// anywhere, so no zone move, counter, attach or tap
		// invalidation stands in for it.
		//
		// The card's own effective cache is nilled at the mutation
		// site (turnFaceUpLocked) rather than here, for the window
		// between the two, exactly as the transform arm explains.
		g.layerVersion.Add(1)
	case EventTurnedFaceDown:
		// ADR 0082 amendment / CR 708.2a, #1209: the same
		// invalidation input as the arm above, travelling the other
		// way. The real card at layer 0 is replaced by the nameless
		// 2/2, so every anthem, every AppliesTo and every "creatures
		// you control" count has a new answer while the permanent has
		// not moved an inch.
		//
		// Nilled at the mutation site (TurnFaceDownForEffect) for the
		// window the transform arm explains, and nilled there for
		// every card in an Ixidron-sized batch BEFORE the first event
		// goes out, so no listener reads a half-turned board.
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
		// conditional for the same reason: a life total is read by a
		// handful of static shapes in the catalog (Aettir and
		// Priwen's base P/T, Serra Ascendant's threshold), and life
		// changes at every table in every combat. With no such
		// permanent in play this is the no-op the two
		// irrelevant-event guards assert.
		//
		// #1117: this arm is the BELT to
		// invalidateLayersForLifeChangeLocked's braces. That helper is
		// called at each of the three places a player's Life field is
		// actually written, which is what gets the ORDERING right on
		// the damage path, and this arm catches any future emitter of
		// EventChangeLife that forgets to call it. A double bump costs
		// one atomic add and still produces exactly one recompute at
		// the next read, so the overlap is free; a missed bump is a
		// card that lies about its own power.
		if lifeTotalStaticIsLiveLocked(g) {
			g.layerVersion.Add(1)
		}
	case EventAttack:
		// #1218: the declare-attackers half of attacking-status
		// invalidation. EventAttack fires once per creature at the
		// declaration's LOCK-IN (commitAttackDeclarationLocked), by
		// which point Card.AttackingTarget is already stamped for
		// every attacker in the batch, so a static reading "is this
		// creature attacking" sees the fresh answer once this bump
		// runs. The other two exits (removed from combat, combat
		// ends) have no event of their own; see
		// invalidateLayersForAttackChangeLocked.
		if attackingStatusStaticIsLiveLocked(g) {
			g.layerVersion.Add(1)
		}
	case EventCast:
		// #1325: a static reading "you haven't cast a spell this
		// turn" (Stoic Sphinx) has a new answer the instant a spell
		// is cast, and nothing else on this listener's switch fires
		// for a cast that doesn't also cross a hand boundary in a way
		// the hand-size gate above would catch — casting from the
		// graveyard (flashback) or exile (foretell, impulse) crosses
		// no hand boundary at all, and even a hand cast is caught
		// there only while a HAND-size static happens to be live,
		// which is a different card's gate. Game.SpellsCastThisTurn
		// (the tally this condition reads) is bumped before EventCast
		// is emitted, so the count is already current here.
		if spellsCastStaticIsLiveLocked(g) {
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
//
// Serra Ascendant is the card that makes it matter: in Commander its
// clause is ON at the opening hand and switches off the first time
// its controller takes eleven damage, so the invalidation is not a
// corner case, it is most of what the card does.
func lifeTotalStaticIsLiveLocked(g *Game) bool {
	return staticOnBattlefieldLocked(g, func(ab StaticAbility) bool { return ab.DependsOnLifeTotal })
}

// attackingStatusStaticIsLiveLocked is handSizeStaticIsLiveLocked for
// attacking status. Ohran Frostfang's "attacking creatures you
// control have deathtouch" is the card that makes it matter.
func attackingStatusStaticIsLiveLocked(g *Game) bool {
	return staticOnBattlefieldLocked(g, func(ab StaticAbility) bool { return ab.DependsOnAttackingStatus })
}

// spellsCastStaticIsLiveLocked is handSizeStaticIsLiveLocked for the
// per-turn spells-cast count. Stoic Sphinx's "hexproof as long as you
// haven't cast a spell this turn" is the card that makes it matter.
func spellsCastStaticIsLiveLocked(g *Game) bool {
	return staticOnBattlefieldLocked(g, func(ab StaticAbility) bool { return ab.DependsOnSpellsCast })
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

// invalidateLayersForLifeChangeLocked drops the cached layer
// resolution when a life total has just changed and something on the
// battlefield is reading one. #1117.
//
// # Why this is not simply an arm of the listener switch
//
// Because a life total changes on two routes with opposite event
// ordering, and only one of them emits a life event at all.
// ChangePlayerLifeForEffect's tail writes the total and THEN emits
// EventChangeLife, so a listener arm is correctly placed. Damage to a
// player writes the total through the same Player.ChangeLife but
// emits EventDealDamage instead — before the write on the non-combat
// route, after it on the combat one (applyResolvedDamageToPlayerLocked
// keeps both orders deliberately, because listeners' tests pin what
// they see). A listener bumping on EventDealDamage would therefore
// invalidate BEFORE the life moved on the commonest route of all, and
// any read in between — the trigger harvester asks for effective
// characteristics — would recache the stale answer with nothing left
// to invalidate it. Serra Ascendant would shrink one Lightning Bolt
// late, forever.
//
// So the bump goes where the write is: immediately after each
// Player.ChangeLife on a game path. Three call sites, each one line,
// each impossible to get out of order.
//
// Caller must hold g.mu.
func (g *Game) invalidateLayersForLifeChangeLocked() {
	if lifeTotalStaticIsLiveLocked(g) {
		g.layerVersion.Add(1)
	}
}

// invalidateLayersForAttackChangeLocked drops the cached layer
// resolution when a permanent's ATTACKING status just changed off the
// declare-attackers event — a control change removing it from combat
// (CR 506.4, removeFromCombatLocked) or combat ending (CR 511.3,
// clearCombatLocked) — and something on the battlefield is reading
// that status. #1218, the other two exits the EventAttack arm above
// cannot reach: neither of these mutates through an event that names
// every affected card, so each call site invalidates directly at the
// write, in the same mould #1117's invalidateLayersForLifeChangeLocked
// does for a life total.
//
// Caller must hold g.mu.
func (g *Game) invalidateLayersForAttackChangeLocked() {
	if attackingStatusStaticIsLiveLocked(g) {
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
