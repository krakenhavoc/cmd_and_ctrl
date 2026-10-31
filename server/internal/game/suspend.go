package game

import (
	"strconv"

	"github.com/google/uuid"
)

// suspend.go — suspend (CR 702.62), #659.
//
// Suspend is four things, and like foretell the interesting part is
// how few of them are new:
//
//  1. A SPECIAL ACTION from hand (CR 702.62a, CR 116.2f): pay the
//     suspend cost, exile the card with N time counters. The CR 116.2
//     verb in special_action.go, with the one timing row that says
//     "whenever you could begin to cast the card" — so a sorcery may
//     be suspended only at sorcery speed, and neither kind may be
//     suspended under split second (CR 702.62c).
//  2. TWO TRIGGERS THAT WORK IN EXILE (CR 702.62a). Both are #925's
//     `TriggeredAbility.Zones = {ZoneExile}`, built for exactly this
//     and for cycling's graveyard trigger.
//  3. A FREE CAST when the last counter comes off (CR 702.62a,
//     CR 608.2g) — ADR 0066's per-instance `CastPermission` with a
//     `{0}` price, `TimingFlash` (a trigger resolving in an upkeep is
//     not a main phase, and Rift Bolt is a sorcery) and the CR 107.3b
//     X lock that comes free with a cost that is not the printed one.
//  4. HASTE on the creature that results (CR 702.62a).
//
// # What "suspended" is
//
// It is having time counters on you in exile (CR 702.62b), and that
// is the whole marker. No flag on Card, no per-card special case, and
// the good consequence is the one #659 was opened for: a copy of the
// same card that reached exile some other way — Path to Exile, a mill
// into a Bojuka Bog — has no time counters, so nothing ticks and
// nothing is ever castable. The retired card-level
// `CastableZones: exile` declaration got that exactly backwards: it
// opened exile for every copy of the card at any time.
//
// # Two triggers, as printed (#990)
//
// CR 702.62a prints two triggered abilities and the engine has two:
//
//	SuspendUpkeepTrigger()        "at the beginning of your upkeep,
//	                               if this card is suspended, remove
//	                               a time counter from it"
//	SuspendLastCounterTrigger()   "when the last time counter is
//	                               removed from this card, if it's
//	                               exiled, cast it without paying its
//	                               mana cost"
//
// #659 shipped ONE, with the last-counter half folded into the
// upkeep trigger's resolution, on the argument that nothing else in
// the engine removes a time counter. That argument was about the
// catalog rather than about the rules, and it cost two things the
// second trigger buys back:
//
//   - A removal from ANYWHERE fires it. Jhoira's Timebug,
//     Clockspinning and Vampire Hexmage take counters off a
//     suspended card without an upkeep anywhere near, and each of
//     them ends the countdown early — which is the whole reason
//     those cards see play alongside suspend.
//   - It is INDEPENDENTLY COUNTERABLE. Two abilities go on the stack
//     at two different moments, so Stifle can answer the free cast
//     and leave the countdown alone, or the other way round.
//
// THEY CANNOT DOUBLE-FIRE, and the shape is what guarantees it
// rather than a guard: the upkeep trigger removes a counter and
// stops. The offer lives in the second trigger and nowhere else, so
// the natural countdown's last tick reaches it the same way
// Clockspinning does — through the `EventCounterPlaced` the removal
// emits. One removal, one event, one offer.
//
// What the second trigger watches is that event with a post-change
// count of ZERO, which is exactly "the last time counter was
// removed": the counter map deletes a key the moment it hits zero,
// so a stored count is always ≥ 1, and a post-change 0 can only be a
// removal that emptied a non-empty pile.
//
// # The free cast is a grant, not an inline cast
//
// The same simplification cascade takes, and the same reason: casting
// from inside a resolution would need the whole announce — targets,
// modes, X, the cost picker — to run under a paused resolution frame.
// So the "yes" answer stamps a permission good for the rest of the
// turn and the player casts it the ordinary way. It is declared here
// rather than left to be discovered, and it is strictly narrower than
// paper in one direction (the window ends at the turn's end rather
// than at the trigger's resolution) and wider in another (the player
// may respond to things in between). Cascade's precedent settled the
// trade for this engine.

// SuspendFreeCastLabel is what the granted cast is called in the log
// and on the wire.
const SuspendFreeCastLabel = "Suspend — cast it without paying its mana cost"

// suspendLocked is the suspend special action's performer, run by
// PerformSpecialAction once the suspend cost is paid (CR 702.62a).
//
// Caller must hold g.mu (write).
func (g *Game) suspendLocked(p *Player, cardID uuid.UUID, sa SpecialAction) error {
	n := sa.Counters
	if n <= 0 {
		// effects.Register refuses this at boot; a hand-built
		// declaration that slipped through exiles nothing rather than
		// exiling a card that can never come back.
		return ErrSpecialActionNotOffered
	}
	owner := p.ID
	return g.routeAllKeyedLocked(zoneRoute{
		Dst:   ZoneExile,
		Actor: owner,
		// #1320: a special action (CR 116.2f), not a spell or ability.
		Cause: MoveCause{Kind: MoveCauseSpecialAction, Controller: owner},
	}, []uuid.UUID{cardID}, KeyedTail{Ref: suspendExiled, Player: owner, N: n})
}

// suspendExiled is the suspend exile's continuation as data (#4361 S11a),
// so a CR 616 prompt on the exile is a restore point. Player is the owner
// and N the number of time counters. An on-disk key.
var suspendExiled = RegisterTail("engine/suspend-exiled", func(g *Game, o TailOutcome) error {
	if len(o.IDs) != 1 {
		// A commander took CR 903.9's offer: it left the hand but
		// not to exile, so it was never suspended.
		return nil
	}
	// Face UP (CR 702.62b says nothing about hiding it), and the
	// counters are the suspended state. Through the counter
	// primitive so the placement is one event the log and any
	// future counter payoff can see.
	// #3942: named, because a commander's CR 903.9 answer can run
	// this after the special action has returned.
	return g.AddCounterByForEffect(o.Tail.Player, o.IDs[0], CounterTime, o.Tail.N)
})

// CardIsSuspended reports whether this card object is suspended
// (CR 702.62b): it is in exile with one or more time counters on it.
//
// The whole marker, and deliberately not a flag: a copy of the same
// card exiled by anything else has no counters, so it is not
// suspended and nothing about suspend applies to it.
func CardIsSuspended(c Card) bool {
	return c.Counters[CounterTime] > 0
}

// SuspendUpkeepTrigger builds the first of the two exile-zone
// triggers every suspended card carries (CR 702.62a): "at the
// beginning of your upkeep, if this card is suspended, remove a time
// counter from it". The catalog attaches it from the card's suspend
// declaration rather than the card file writing it out, so the
// countdown cannot be spelled differently on two cards.
//
// It watches `EventBeginUpkeep` from ZoneExile — #925's zone
// dimension, which exists for exactly this — and only fires while the
// card really is suspended, so a Path to Exile'd copy of the same
// card sitting in the same exile zone never ticks.
//
// "You" is the card's OWNER while it sits in exile (CR 108.4), and
// the zone harvest hands the predicate a source whose Controller is
// its Owner, so the ordinary "is this your upkeep" comparison reads
// the way the card prints it.
func SuspendUpkeepTrigger() TriggeredAbility {
	return TriggeredAbility{
		Zones:   []ZoneKind{ZoneExile},
		Watches: []EventKind{EventBeginUpkeep},
		Key:     "Suspend — remove a time counter",
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.Actor == source.Controller && CardIsSuspended(*source)
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			// ADR 0041 P9 (#1497, tier 4): suspend has no catalog row,
			// so this item is keyed directly.
			return NewKeyedTriggeredItem(source, "Suspend — remove a time counter", suspendTickBody, EffectParams{})
		},
	}
}

// suspendTickBody registers suspendTick under "suspend/tick" (ADR 0041
// P9, #1497, tier 4). Registered in init, not by a var initialiser: the
// trigger reaches suspendTick, which reaches the counter code, which reads
// the triggers of a card (triggersOf derives these two for a card an effect
// gave suspend), and an initialiser would close that loop (#4455).
var suspendTickBody BodyRef

func init() { suspendTickBody = SimpleDelayedBody("suspend/tick", suspendTick) }

// suspendTick is the upkeep trigger's resolution, and the whole of it:
// remove one time counter (CR 702.62a).
//
// It does NOT offer the cast. Taking the last counter off emits the
// `EventCounterPlaced` that SuspendLastCounterTrigger watches, so the
// offer arrives by the same road an external removal takes — see the
// file header. Before #990 the offer was made inline from here, which
// is what made a Clockspinning removal a dead end.
//
// Everything is re-read at resolution rather than captured at
// announce, because CR 608.2 resolves against the game as it is: the
// card may have left exile, or somebody may have removed its last
// counter in response.
func suspendTick(g *Game, item *StackItem) error {
	cardID := item.SourceCardID
	c := exiledCardByIDLocked(g, cardID)
	if c == nil || !CardIsSuspended(*c) {
		// CR 608.2: it is not there any more, or it is not suspended
		// any more. The trigger does as much as it can, which is
		// nothing.
		return nil
	}
	return g.AddCounterForEffect(cardID, CounterTime, -1)
}

// SuspendLastCounterTrigger builds the second exile-zone trigger
// (CR 702.62a): "when the last time counter is removed from this
// card, if it's exiled, cast it without paying its mana cost". Like
// its sibling the catalog attaches it from the suspend declaration.
//
// It watches the COUNTER EVENT rather than the upkeep, which is the
// point of #990: the upkeep removal, Jhoira's Timebug, Clockspinning
// and Vampire Hexmage all reach it the same way, and none of them
// needs to know suspend exists.
//
// `Amount` on EventCounterPlaced is the post-change count, so zero is
// "the pile is now empty" — and only a removal can produce it,
// because applyCounterLocked deletes a counter key the moment it
// stops being positive and an addition of zero emits nothing. There
// is deliberately no `CardIsSuspended` guard: by the time this fires
// the card is not suspended any more, which is the condition.
//
// The "if it's exiled" clause is the harvest itself — the zone walk
// only offers this ability to a card sitting in exile (CR 113.6) —
// and it is re-checked at resolution (CR 603.4) by
// offerSuspendedCastLocked.
func SuspendLastCounterTrigger() TriggeredAbility {
	return TriggeredAbility{
		Zones:   []ZoneKind{ZoneExile},
		Watches: []EventKind{EventCounterPlaced},
		Key:     SuspendFreeCastLabel,
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.Target == source.InstanceID && ev.Label == CounterTime && ev.Amount == 0
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			// ADR 0041 P9 (#1497, tier 4): suspend has no catalog row,
			// so this item is keyed directly.
			return NewKeyedTriggeredItem(source, SuspendFreeCastLabel, suspendFreeCastBody, EffectParams{})
		},
	}
}

// suspendFreeCastBody registers suspendLastCounterRemoved under
// "suspend/free-cast" (ADR 0041 P9, #1497, tier 4), in init for the reason
// suspendTickBody is.
var suspendFreeCastBody BodyRef

func init() {
	suspendFreeCastBody = SimpleDelayedBody("suspend/free-cast", suspendLastCounterRemoved)
}

// suspendLastCounterRemoved is the second trigger's resolution: the
// CR 702.62a offer, made against the game as it is now (CR 608.2).
// A card that left exile while the trigger was on the stack is not
// offered anything, which is the intervening "if it's exiled" read at
// resolution.
func suspendLastCounterRemoved(g *Game, item *StackItem) error {
	return g.offerSuspendedCastLocked(item.Controller, item.SourceCardID)
}

// offerSuspendedCastLocked is CR 702.62a's third ability: "when the
// last time counter is removed from this card, if it's exiled, you may
// play it without paying its mana cost if able. If you don't, it remains
// exiled."
//
// The cast happens during this resolution (CR 608.2g; ADR 0066's
// 2026-10-23 amendment, #4250): the free cast is stamped on the card and
// a resolution_cast prompt over it holds the resolution open while its
// owner casts it, with its targets and modes chosen then, or declines.
// The decline withdraws the permission and nothing else, so the card
// stays in exile with no counters, which is the printed outcome. Outside
// an open resolution the permission alone is the window, for the rest of
// the turn as before.
//
// Caller must hold g.mu (write).
func (g *Game) offerSuspendedCastLocked(chooser, cardID uuid.UUID) error {
	c := exiledCardByIDLocked(g, cardID)
	if c == nil || !g.playerInGameLocked(chooser) {
		return nil
	}
	g.grantSuspendedFreeCastLocked(chooser, cardID)
	g.queueResolutionCastOfferLocked(resolutionCastOffer{
		Chooser:      chooser,
		Source:       cardID,
		Reason:       "Suspend",
		Candidates:   []uuid.UUID{cardID},
		Keyword:      MayCastKeywordSuspend,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it in exile",
	})
	return nil
}

// grantSuspendedFreeCastLocked stamps the free cast on the one exiled
// object whose last time counter just came off.
//
// Three fields carry the rules:
//
//   - `Cost: "{0}"` is "without paying its mana cost" (CR 702.62a).
//     Spelled as a zero cost rather than left empty, for the reason
//     cascade's grant spells it: empty means "pay the printed cost",
//     which for Lotus Bloom and Ancestral Vision is CR 118.6's
//     unpayable one. It is also what locks X at 0 (CR 107.3b) through
//     CastCostFor, with no second rule anywhere.
//   - `Timing: TimingFlash`, because the trigger resolves in an
//     UPKEEP and Rift Bolt is a sorcery. CR 608.2g lets the cast
//     happen during that resolution whatever the card's own timing
//     says, and the grant is how this engine spells that.
//   - `GrantsHaste`, CR 702.62a.
//
// Caller must hold g.mu (write).
func (g *Game) grantSuspendedFreeCastLocked(player, cardID uuid.UUID) {
	c := exiledCardByIDLocked(g, cardID)
	if c == nil {
		return
	}
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:      player,
		Zone:        ZoneExile,
		Cost:        "{0}",
		Timing:      TimingFlash,
		CastOnly:    true,
		GrantsHaste: true,
		// CR 702.62a: the cast happens as the trigger resolves, so
		// the window is this turn and no longer (#945). Under the
		// resolution_cast prompt (#4250) it is narrower still: the
		// prompt's answer ends it. Stamped
		// explicitly rather than left zero because the trigger fires
		// in an UPKEEP, and a reader should be able to see which turn
		// the grant names without tracing the write path.
		Duration: g.UntilEndOfTurnDuration(),
		Label:    SuspendFreeCastLabel,
	}, []Card{*c})
}

// grantHasteForCastLocked is CR 702.62a — "if the resulting spell is
// a permanent spell, that permanent gains haste" — applied at the
// moment the granted cast is announced.
//
// A layer-6 keyword grant scoped to the ONE object the permission
// opened, and to its controller: losing control of the creature ends
// it, which is the clause CR 702.62a actually writes ("until that
// player loses control of it").
//
// #990 RETIRED THE UNTIL-END-OF-TURN SIMPLIFICATION #659 shipped.
// The duration is CR 611.2b's "for as long as", through ADR 0063's
// one model: `UntilYouLoseControlOfDuration`, whose condition is
// false the moment the named player stops controlling the named
// object, and whose GRACE PERIOD — the spell is still on the stack
// when this runs — is what let the grant move onto the model at all
// (see duration.go's WhileYouControlSourceOnceItLands).
//
// Two things follow, and both are the rule rather than a bonus:
//
//   - The haste OUTLIVES THE TURN. A suspended creature stolen on a
//     later turn and given back is summoning sick for its original
//     controller, because the grant ended the first time control
//     left them and a ForAsLongAs duration that has gone false is
//     swept rather than re-evaluated back to life.
//   - A card that never becomes a permanent — Rift Bolt, Ancestral
//     Vision — drops its grant the moment the spell leaves the
//     stack, instead of sitting in the registry until cleanup.
//
// Caller must hold g.mu (write).
func (g *Game) grantHasteForCastLocked(player, cardID uuid.UUID) {
	// A data record (ADR 0041 phase 3, #1497). EnteredAt 0 because the
	// card is still a spell on the stack and gets its entry stamp only
	// as it lands; Controller because the haste applies only while that
	// player controls it. The duration ends the record when control
	// goes; the member test keeps the pass exact until the sweep does.
	g.RegisterScopedEffectForEffect(cardID,
		[]AffectedObject{{ID: cardID, Controller: player}},
		[]Mod{AddKeywordsMod("haste")},
		UntilYouLoseControlOfDuration(cardID, player), "Suspend — haste (CR 702.62a)")
}

// SuspendLabel is the printed keyword line — "Suspend 3—{R}" — built
// once so the wire, the menu row and the log cannot disagree.
func SuspendLabel(n int, cost string) string {
	return "Suspend " + strconv.Itoa(n) + "—" + cost
}

// grantedSuspendZones adds exile to the zones an upkeep or a counter
// event walks (#4455): a card an effect gave suspend carries the two
// exile triggers without the catalog declaring them, so the index cannot be
// the only reason the exile zone is read for these two events. Any other
// event, or a walk that already names exile, is returned as it came.
func grantedSuspendZones(kind EventKind, zones []ZoneKind) []ZoneKind {
	if kind != EventBeginUpkeep && kind != EventCounterPlaced {
		return zones
	}
	for _, z := range zones {
		if z == ZoneExile {
			return zones
		}
	}
	return append(append([]ZoneKind(nil), zones...), ZoneExile)
}

// CardHasSuspend reports whether the card has suspend (CR 702.62a): it
// prints it, or an effect gave it (Card.SuspendGranted). The question
// "if it doesn't have suspend" asks.
func CardHasSuspend(c Card) bool {
	if c.SuspendGranted {
		return true
	}
	for _, sa := range SpecialActionsFor(CatalogKey(c)) {
		if sa.Kind == SpecialActionSuspend {
			return true
		}
	}
	return false
}

// GainSuspendForEffect is "Put N time counters on that card. If it doesn't
// have suspend, it gains suspend." (Kang Prime, Doom's Time Platform, CR
// 702.62a), over a card that is in exile: the counters go on first, then a
// card without suspend is marked as having been given it, which is what
// puts the two exile triggers on it (triggersOf). A card that prints
// suspend, or was already given it, gets only the counters, so it never
// has two instances. A card that is no longer in exile is left alone
// (CR 608.2b): there is nothing to count down.
//
// Caller must hold g.mu (write).
func (g *Game) GainSuspendForEffect(player, cardID uuid.UUID, n int) error {
	c := exiledCardByIDLocked(g, cardID)
	if c == nil || n <= 0 {
		return nil
	}
	if err := g.AddCounterByForEffect(player, cardID, CounterTime, n); err != nil {
		return err
	}
	// Re-read: the counter placement may have rewritten the zone's slice.
	c = exiledCardByIDLocked(g, cardID)
	if c == nil || CardHasSuspend(*c) {
		return nil
	}
	c.SuspendGranted = true
	return nil
}

// GainSuspendExileTail is the keyed continuation that finishes an exile
// whose cards "gain suspend": Tail.Player is the controller and N the
// number of time counters. Every card the exile reached is counted down
// from here, so a CR 616 prompt on the exile is a restore point. An
// on-disk key.
var GainSuspendExileTail = RegisterTail("engine/gain-suspend-exiled", func(g *Game, o TailOutcome) error {
	for _, id := range o.IDs {
		if err := g.GainSuspendForEffect(o.Tail.Player, id, o.Tail.N); err != nil {
			return err
		}
	}
	return nil
})
