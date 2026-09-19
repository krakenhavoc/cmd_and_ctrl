package game

import "github.com/google/uuid"

// simultaneous.go — S23: making a board wipe look like ONE event to
// the cards watching it (CR 700.4, CR 603.10).
//
// The problem this file solves. Destruction in this engine is a
// per-card call: routeBattlefieldCardToOwnerGraveyardLocked moves one
// permanent and emits its own EventLTB, and the trigger harvester
// answers each event by walking g.Battlefield. That is correct for a
// Doom Blade and wrong for a Wrath of God, because a wrath destroys
// every creature SIMULTANEOUSLY. The observable difference is not
// subtle:
//
//	Blood Artist + three other creatures, all wiped.
//	  Sequential:   the Artist is processed first, leaves the
//	                battlefield, and the harvester's battlefield walk
//	                never finds it again — the three later deaths
//	                drain nobody. One trigger, and only if the Artist
//	                happened to be last.
//	  Simultaneous: all four creatures die at once, the Artist sees
//	                every one of them including itself (CR 603.10
//	                last-known information), and the table loses four
//	                life.
//
// Four is the right answer and everyone who has played the deck
// knows it, so this is a rules bug with a play-pattern cost rather
// than a technicality.
//
// The fix, and why it is small. Making the MOVES simultaneous would
// mean restructuring the replacement pipeline (the CR 903.9 commander
// prompt is asynchronous — see routeBattlefieldCardToOwnerGraveyardLocked)
// for no gain, because nothing in the engine can observe the moves
// themselves. What IS observed is the harvest, and the harvest is
// wrong for exactly one reason: a watcher that already left the
// battlefield is no longer in the zone the harvester scans.
//
// So the batch is published, not deferred. Before the first move,
// every card in the batch is copied; while the batch is open, the
// harvester scans those copies alongside the live battlefield,
// skipping any that are still on it (harvestFromZone already has
// those) and the card whose own death is being reported
// (harvestLTB already has that one). The moves stay sequential; the
// observation becomes simultaneous, which is the half that has
// rules consequences.
//
// Deliberately NOT addressed here:
//   - Totem armor is absent.
//
// Regeneration USED to be on that list too, with the same shape of
// note indestructible below carried: "not modelled anywhere in the
// engine, so they can't be regenerated is still cosmetic". #667
// shipped CR 701.19, and the batch DID have to change for it — see
// doomedPermanent and routeAllLandedPerLegLocked below. One
// state-based-action pass is one event (CR 704.3) and its doomed set
// mixes two rules that destroy with three that merely put a permanent
// into a graveyard, so the legs now leave together by different
// routes rather than together by one.
//
// Indestructible USED to be on that list, and the entry outlived its
// truth: S25 (#380) shipped CR 702.12 hours after this file landed,
// and this note kept telling readers the engine had no such concept
// for four sprints while every board wipe killed an Avacyn
// (#470 / #446). DestroyPermanentsForEffect now filters the set
// through DestructibleForEffect before the batch opens.
//   - The CR 903.9 commander replacement still fires per card, and
//     still fires ASYNCHRONOUSLY: a commander in the batch queues its
//     owner's yes/no prompt and does not move until they answer. That
//     is the pre-existing shape and the batch preserves it — the
//     commander is counted as destroyed either way, because CR 903.9
//     replaces the zone change, not the destruction.
//
// #815 changed one thing about that last note and left the rest
// standing. "Counted either way" was being applied to the prompt as
// well as to the answer: a leg that had merely PAUSED counted as a
// destruction before anybody had said anything, and so did one the
// window had CANCELLED outright. The count now comes from the landed
// outcome (destroyedThisWayLocked), and the caller that needs it waits
// for the answer through DestroyPermanentsThenForEffect. A commander
// that takes the offer is still destroyed; a creature that an "exile
// it instead" replacement removed is not (CR 701.7a).
//
// #866 finished that: exile and bounce were still counting the attempt
// rather than the arrival, and Settle the Wreckage was searching for a
// basic land per creature whose owner had merely been ASKED about the
// command zone. Rather than a third and fourth copy of the sequencing,
// the three verbs now share ONE batch body (below), parameterised by
// the zoneRoute each leg takes and by what "landed" means for it:
// CR 701.7a plus the command-zone carry-over for a destruction,
// CR 400.7's arrived-object for everything else.

// beginSimultaneousExitLocked publishes `ids` as one simultaneous
// battlefield exit and returns the closer. The copies are taken
// BEFORE any move so that a watcher processed early in the batch is
// still visible to the harvester when its neighbours die.
//
// Cards in `ids` that are not on the battlefield are skipped. The
// returned func must be called (defer it) — leaving a batch open
// would leak dead cards into every later harvest.
//
// Caller must hold g.mu in write mode.
func (g *Game) beginSimultaneousExitLocked(ids []uuid.UUID) func() {
	return g.publishSimultaneousExitLocked(g.simultaneousExitSnapshotLocked(ids))
}

// simultaneousExitSnapshotLocked takes the pre-move copies the batch
// is made of. Split out from beginSimultaneousExitLocked because a
// batch that can PAUSE has to carry its copies across the pause: the
// destroy-with-a-continuation path (#815) takes them once, before the
// first move, and re-publishes the SAME copies from each leg's
// continuation, so a wipe whose commander stops to answer CR 903.9
// still looks like one event to every dies-trigger in it. Re-taking
// them on the far side of the prompt would find the creatures that
// already left missing, which is the whole thing this file exists to
// prevent.
//
// Caller must hold g.mu in write mode.
func (g *Game) simultaneousExitSnapshotLocked(ids []uuid.UUID) []Card {
	batch := make([]Card, 0, len(ids))
	for _, id := range ids {
		if i := findCardOnBattlefield(g, id); i >= 0 {
			batch = append(batch, g.Battlefield.Cards[i])
		}
	}
	return batch
}

// publishSimultaneousExitLocked installs a batch of pre-move copies as
// the open simultaneous exit and returns the closer. The returned func
// must be called (defer it).
//
// Caller must hold g.mu in write mode.
func (g *Game) publishSimultaneousExitLocked(batch []Card) func() {
	if len(batch) == 0 {
		return func() {}
	}
	prev := g.simultaneousExit
	g.simultaneousExit = batch
	return func() { g.simultaneousExit = prev }
}

// harvestSimultaneousExitLocked is the second half of the trigger
// harvest while a batch is open: the watchers that have ALREADY left
// the battlefield as part of this same event.
//
// Two exclusions keep it from double-firing:
//
//   - ev.CardID — the card whose death is being reported. Its own
//     triggers are harvestLTB's job, which finds it in its
//     destination zone and pairs it with the real LKI snapshot.
//   - anything still on the battlefield — harvestFromZone walks the
//     live zone first and has already seen it.
//
// The characteristics handed to AppliesTo / Build are the ones the
// copy carried when the batch opened, which is precisely CR 603.10's
// last-known information for a permanent that is no longer there.
//
// Caller must hold g.mu in write mode.
func (g *Game) harvestSimultaneousExitLocked(pass *harvestPass) {
	ev := pass.ev
	if len(g.simultaneousExit) == 0 || CatalogTriggers == nil {
		return
	}
	// Snapshot the batch header: dispatching a trigger can open a
	// nested batch (a dies-trigger that sacrifices), and ranging over
	// a slice the callee may replace is how that becomes a bug.
	batch := g.simultaneousExit
	for i := range batch {
		card := batch[i]
		if card.InstanceID == ev.CardID {
			continue
		}
		if findCardOnBattlefield(g, card.InstanceID) >= 0 {
			continue
		}
		// TriggersForCard, and it answers off the copy the batch
		// captured while the card was still on the battlefield — so
		// a creature wiped while under a Kenrith's Transformation
		// has no dies-trigger here either, and a Case that was not
		// solved when the wipe hit has no solved dies-trigger, for
		// the same CR 603.10 reason harvestLTB reads its snapshot.
		triggers := TriggersForCard(card)
		if len(triggers) == 0 {
			continue
		}
		lki := card.Effective()
		for _, t := range triggers {
			// #925: same rule the battlefield walk and harvestLTB
			// apply — this pass is about permanents that were on the
			// battlefield when the batch opened, so an ability that
			// declared another zone is not one of them.
			if !TriggerWatchesFromZone(t, ZoneBattlefield) {
				continue
			}
			if !triggerWatches(t.Watches, ev.Kind) {
				continue
			}
			if t.AppliesTo != nil && !t.AppliesTo(ev, &card, lki, g) {
				continue
			}
			g.harvestMatchLocked(pass, card, lki, t, false)
		}
	}
}

// DestroyPermanentsForEffect destroys every permanent in `ids` as one
// simultaneous event and returns how many of them were destroyed.
//
// The FIRE-AND-FORGET form. The count it returns is the one taken
// while the call is still on the stack, so it cannot include a leg
// that PAUSED on the CR 903.9 prompt — a card that reads "for each
// creature destroyed this way" must use
// DestroyPermanentsThenForEffect, which waits (#815).
//
// This is the mass-destruction entry point. A single-target destroy
// stays on DestroyPermanentForEffect; routing one card through here
// would be correct but would pay for a batch nothing observes.
//
// S30 (#470 / #446): indestructible permanents are dropped from the
// set BEFORE the batch opens, so a wipe neither destroys them nor
// announces their death to the dies-triggers watching. See
// DestructibleForEffect in indestructible.go for why the filter is
// here rather than in destroyPermanentsLocked, which the SBA sweep
// shares and whose zero-counter branches indestructible must not
// save.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) DestroyPermanentsForEffect(ids []uuid.UUID, opts ...DestroyOptions) int {
	return g.destroyPermanentsLocked(g.DestructibleForEffect(ids), firstDestroyOptions(opts))
}

// DestroyPermanentsThenForEffect destroys every permanent in `ids` as
// one simultaneous event and hands `then` the ones that were actually
// DESTROYED — "for each creature destroyed this way" (Fumigate's life
// gain, Deadly Tempest's life loss, Bane of Progress's counters, Blood
// Money's treasures).
//
// It is a continuation rather than a return value for the reason
// LoseLifeEachThenForEffect and DealDamageEachThenForEffect are
// (ADR 0013 §5b, §5c): any leg can pause. A commander caught in the
// wipe stops to answer CR 903.9, and the number is not knowable until
// it does. The legs are therefore destroyed IN SEQUENCE, each from the
// previous one's continuation, with the landed list carried forward by
// value — which is what makes an undo across the prompt replay
// identically, and what lets the count be true rather than optimistic.
//
// The observable cost is the one the drain and the discard batch
// already pay: a paused leg delays the rest of the sweep until the
// prompt is answered. What it does NOT cost is the simultaneity —
// the pre-move copies are taken once, before the first move, and
// re-published from every leg, so a Blood Artist still sees the whole
// board die with it.
//
// Indestructible is filtered first, exactly as the fire-and-forget
// form does, so a survivor is neither destroyed nor counted.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) DestroyPermanentsThenForEffect(ids []uuid.UUID, then func(g *Game, destroyed []uuid.UUID) error, opts ...DestroyOptions) error {
	return g.routeAllThenLocked(destroyRouteWith(firstDestroyOptions(opts)), g.DestructibleForEffect(ids), then)
}

// destroyedThisWayLocked reports whether a settled destruction of
// cardID counts as a DESTRUCTION — the question "for each creature
// destroyed this way" is asking.
//
// CR 701.7a: "To destroy a permanent, move it from the battlefield to
// its owner's graveyard." So the answer is read off the LANDED
// outcome, not off the attempt, and there are three of them:
//
//   - the permanent is in a graveyard. Destroyed.
//   - the permanent is in the command zone. Destroyed: CR 903.9
//     replaces where the card goes, not whether it was destroyed, and
//     the engine has counted a commander either way since S23 (see the
//     note at the top of this file). Declared rather than derived —
//     this is the one place to change it if that ever flips.
//   - anything else. NOT destroyed. A permanent still on the
//     battlefield had its destruction cancelled outright
//     (indestructible granted mid-window, "it isn't destroyed"), and
//     one that a replacement sent somewhere else — exile, hand,
//     library — was never put into a graveyard, so by CR 701.7a it was
//     not destroyed however thoroughly it left.
//
// Read off the live board rather than off the event, because that is
// the reading that is still true after an undo rewinds into an open
// CR 903.9 prompt and the answer is replayed.
//
// Caller must hold g.mu.
func (g *Game) destroyedThisWayLocked(cardID uuid.UUID) bool {
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		return false
	}
	return z.Kind == ZoneGraveyard || z.Kind == ZoneCommand
}

// sacrificedThisWayLocked reports whether a settled sacrifice of
// cardID counts as a SACRIFICE — the question "for each permanent
// sacrificed this way" (God-Eternal Bontu's draw) is asking.
//
// CR 701.17a: "To sacrifice a permanent, its controller moves it from
// the battlefield to its owner's graveyard." The sacrifice is that
// MOVE OFF the battlefield, and nothing replaces the sacrifice itself
// — a replacement rewrites where the permanent goes. So the answer is
// "is it still on the battlefield":
//
//   - anywhere else. Sacrificed. A graveyard is the printed
//     destination; the command zone is CR 903.9 taking the offer; exile
//     is Rest in Peace or Liesa, and a hand or a library is some other
//     rewrite. All of them replaced the destination of a sacrifice
//     that had already happened.
//   - gone from the game entirely (a sacrificed TOKEN, CR 111.8).
//     Sacrificed: it left the battlefield because its controller
//     sacrificed it.
//   - still on the battlefield. NOT sacrificed — the CR 614 window
//     cancelled the move outright, or its prompt was abandoned
//     (ADR 0013 §5j), and nothing ever left.
//
// This is where sacrifice PARTS COMPANY with destroy, and the two
// rules are why. CR 701.7a defines a destruction BY the graveyard, so
// a permanent a replacement sent to exile was not destroyed however
// thoroughly it left (destroyedThisWayLocked, #863) — the command zone
// is destroy's one declared carry-over. CR 701.17a names the same
// destination but the keyword action is the controller's move off the
// battlefield, and Korvold triggers on a sacrifice whose card an
// "exile it instead" replacement took. Two rules, two functions, one
// board read each; routeLegLandedLocked picks between them.
//
// Read off the live board rather than off the settled event, for the
// reason destroyedThisWayLocked gives: it is the reading that is still
// true after an undo rewinds into an open prompt.
//
// Caller must hold g.mu.
func (g *Game) sacrificedThisWayLocked(cardID uuid.UUID) bool {
	return findCardOnBattlefield(g, cardID) < 0
}

// landedInZoneLocked reports whether cardID is now in the zone its
// route asked for — the CR 400.7 reading of "exiled this way",
// "returned this way", "put into a graveyard this way": the object
// that ARRIVED is the one the effect moved.
//
// So a leg a replacement sent somewhere else did not go this way,
// however thoroughly it left. A commander that took CR 903.9's offer
// went to the command zone rather than to exile, and "for each card
// exiled this way" does not see it. A leg the CR 614 window cancelled,
// and one whose prompt was abandoned (#865), never moved at all.
//
// Destroy is the one route that does NOT use this, because CR 701.7a
// defines destruction by the move to a graveyard and the engine
// declares the command zone a destruction as well (ADR 0013 §5i).
// destroyedThisWayLocked holds that rule; routeLegLandedLocked picks
// between the two.
//
// Read off the live board rather than off the settled event, for the
// reason destroyedThisWayLocked gives: it is the reading that is still
// true after an undo rewinds into an open prompt.
//
// Caller must hold g.mu.
func (g *Game) landedInZoneLocked(cardID uuid.UUID, dst ZoneKind, dstOwner uuid.UUID) bool {
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		return false
	}
	want, _, err := g.routeDestinationLocked(cardID, dst, dstOwner)
	return err == nil && want != nil && want == z
}

// --- the one batch-sequencing body -------------------------------
//
// #866. #815 gave the destroy sweep a continuation that reports what
// actually LANDED, and left exile and bounce counting a paused or
// cancelled leg the way destroy used to — Settle the Wreckage searched
// for a basic land per creature whose owner had merely been ASKED
// about the command zone. Writing that twice more would have been
// three copies of one sequencing loop, so the loop moved here and the
// three verbs became what genuinely differs between them: the route
// each leg takes, and what "landed" means for it.
//
// Both halves of a verb are built from the same zoneRoute template.
// The template is a DESCRIPTION, never a route that is run as-is: each
// leg gets its own copy with CardID, simultaneousExit and then filled
// in (routeLegLocked).

// destroyRoute, exileRoute and bounceRoute are the three templates.
//
// destroyRoute carries no destination: ViaBattlefieldLeave says the
// physical move belongs to executeBattlefieldLeaveLocked, whose event
// names its own default ("its owner's graveyard", or exile when the
// owner has left the table). See zoneRoute.ViaBattlefieldLeave.
//
// battlefieldExitRoute is the same exit WITHOUT the destruction flag,
// and it is the one every other way off the battlefield takes: a
// sacrifice (CR 701.21a), the legend rule (CR 704.5j), an illegally
// attached Aura (CR 704.5m), a creature at zero toughness (CR 704.5f),
// a planeswalker at zero loyalty (CR 704.5i), a battle at zero defense
// (CR 704.5v). None of those is a destruction, so none of them may be
// regenerated — and the flag is the only thing that says so, because
// all of them end in the same graveyard through the same primitive
// (#667, regeneration.go).
var (
	destroyRoute         = zoneRoute{ViaBattlefieldLeave: true, Destruction: true}
	battlefieldExitRoute = zoneRoute{ViaBattlefieldLeave: true}
	exileRoute           = zoneRoute{Dst: ZoneExile}
	bounceRoute          = zoneRoute{Dst: ZoneHand}
)

// destroyRouteWith is the destroy template carrying the rider a
// particular destruction printed — "it can't be regenerated" and
// nothing else, today (CR 701.19c).
func destroyRouteWith(opts DestroyOptions) zoneRoute {
	r := destroyRoute
	r.CantBeRegenerated = opts.CantBeRegenerated
	return r
}

// sacrificeRoute is the SEVENTH template (#910): the exit a SACRIFICE
// takes, built on battlefieldExitRoute rather than on destroyRoute
// because a sacrifice is not a destruction (CR 701.17a — "sacrificing
// a permanent doesn't destroy it, so regeneration and other effects
// that replace destruction can't affect it"). Inheriting the
// undestructive route is what makes that true by construction rather
// than by a comment: no Destruction flag, so the CR 701.19 shield
// never sees it (#667).
//
// It differs from the exits around it in exactly two declared ways.
//
//   - it ANNOUNCES. EventSacrifice is emitted while the permanent is
//     still on the battlefield, before the window opens over its move,
//     so a "whenever you sacrifice" payoff can read what it is losing
//     (sacrifice.go). No other battlefield exit announces anything.
//   - "this way" is a different rule, sacrificedThisWayLocked's.
//
// A function rather than a package var because the announcement names
// the card that asked for the sacrifice, which belongs to the caller.
func sacrificeRoute(source uuid.UUID) zoneRoute {
	r := battlefieldExitRoute
	r.Sacrifice = true
	r.Source = source
	return r
}

// millRoute is the fourth template (#893) and the one that cannot be a
// package var, because two things about a mill are decided by the
// caller rather than by the verb:
//
//   - the destination. CR 701.17a's mill is library -> graveyard, and
//     the same helper expresses "exile the top N cards of your
//     library", which is not a mill at all. Mill is set only for the
//     graveyard, so EventMill fires for a mill and an ordinary zone
//     move fires for the exile — the distinction executeZoneRouteLocked
//     already makes off the SETTLED destination, so a commander that
//     took CR 903.9's offer emits neither.
//   - the Actor, which is the player whose library is being read. It
//     is stamped on the events, and it is not the card's owner: an
//     opponent's Glimpse the Unthinkable mills YOUR library.
//
// The destination is where the cards were ASKED to go, which is what
// landedInZoneLocked measures "milled this way" against (CR 400.7).
func millRoute(player uuid.UUID, dest ZoneKind) zoneRoute {
	return zoneRoute{Dst: dest, Actor: player, Mill: dest == ZoneGraveyard}
}

// searchRoute is the SIXTH template (#931): a library search that
// takes the cards it finds into a hand or a graveyard — "search your
// library for a card and put it into your graveyard" (Entomb, Buried
// Alive, Unmarked Grave, Vile Entomber, Goblin Engineer, Final
// Parting's first half).
//
// Before this the take was a raw MoveCard, so the one thing a
// graveyard arrival owes — the CR 614 window, "if a card would be put
// into a graveyard from anywhere, exile it instead" — never opened for
// it, and neither did CR 903.9 for a tutored commander. Rest in Peace
// could not be written because of it (ADR 0061 §7).
//
// Like millRoute it is a function rather than a package var, for the
// same two reasons: the destination belongs to the CARD ("into your
// graveyard" / "into your hand") and the Actor is the SEARCHER, who is
// always the library's owner but not always the card's — Assassin's
// Trophy makes the victim search their own library.
//
// It is NOT a mill: CR 701.17a defines a mill from the TOP of a
// library by count, and no mill payoff may see an Entomb. That is the
// one thing it does not share with millRoute, and it is why the two
// are separate templates rather than one with a flag. A battlefield
// destination never reaches here — an entry is not an exit, and
// searchEnterBattlefieldLocked owns it.
func searchRoute(player uuid.UUID, dest ZoneKind) zoneRoute {
	return zoneRoute{Dst: dest, Actor: player}
}

// tuckRoute is the fifth template (#783): "put it into its owner's
// library", at the position the tucking effect printed.
//
// The position rides the route rather than being applied by the caller
// after the move, for the reason zoneRoute.Depth gives: a commander
// tucked to the bottom, or third from the top, whose owner is asked
// about the command zone and declines still lands where the card said,
// because nothing has moved until the prompt is answered.
//
// The destination owner is deliberately left nil — every printed tuck
// is "its OWNER's library" (CR 903.9's destinations all are), and
// routeDestinationLocked resolves that per card, which is what lets one
// batch tuck four attackers into four different libraries.
func tuckRoute(opts TuckOptions) zoneRoute {
	return zoneRoute{Dst: ZoneLibrary, ToBottom: opts.ToBottom, Depth: opts.Depth}
}

// routeAllThenLocked routes every card in `ids` as one simultaneous
// exit and hands `then` the ones that LANDED where `r` asked.
//
// It is a continuation rather than a return value because any leg can
// pause on the CR 903.9 prompt, so the answer is not knowable on the
// line after the sweep (ADR 0013 §5b, §5c, §5i). The legs go IN
// SEQUENCE, each from the previous one's continuation, with the landed
// list carried forward BY VALUE — the property that makes an undo
// across the prompt replay identically.
//
// The pre-move copies are taken ONCE, before the first move, and
// re-published from every leg, so a Blood Artist still sees the whole
// board leave with it on either side of a prompt.
//
// Caller must hold g.mu in write mode.
func (g *Game) routeAllThenLocked(r zoneRoute, ids []uuid.UUID, then func(g *Game, landed []uuid.UUID) error) error {
	return g.routeEachStepLocked(r, g.simultaneousExitSnapshotLocked(ids), ids, nil, then)
}

// routeEachStepLocked routes the head of `ids` and continues with the
// tail from that leg's continuation. The empty list is the base case:
// the batch is done and `then` gets the landed list.
//
// Caller must hold g.mu in write mode.
func (g *Game) routeEachStepLocked(
	r zoneRoute,
	batch []Card,
	ids, landed []uuid.UUID,
	then func(g *Game, landed []uuid.UUID) error,
) error {
	// A leg with nothing to do is skipped rather than routed, and is
	// not counted: it left while an earlier leg was paused, it was
	// never there, or it is already where the route would put it. It
	// did not go THIS way, so it must not be paid out as if it had.
	for len(ids) > 0 && g.routeLegNothingToDoLocked(r, ids[0]) {
		ids = ids[1:]
	}
	if len(ids) == 0 {
		if then == nil {
			return nil
		}
		return then(g, landed)
	}
	next, rest := ids[0], ids[1:]
	return g.routeLegLocked(r, next, batch, func(g *Game) error {
		out := landed
		if g.routeLegLandedLocked(r, next) {
			// A fresh slice rather than an append in place: two runs of
			// the same continuation (an undo, then the same answer
			// again) must not see each other's entry.
			out = append(append(make([]uuid.UUID, 0, len(landed)+1), landed...), next)
		}
		return g.routeEachStepLocked(r, batch, rest, out, then)
	})
}

// routeAllLocked is the FIRE-AND-FORGET form of the same batch: every
// leg is routed on this line and the count is how many of them had
// LANDED by the time the call returns.
//
// Which is every leg that settled. A leg that paused on the CR 903.9
// prompt cannot be in it — nothing has moved yet — which is exactly
// why a card that reads the number uses the `Then` form instead. The
// batch is published for the whole loop here rather than re-published
// per leg, because nothing in it can pause the loop.
//
// Caller must hold g.mu in write mode.
func (g *Game) routeAllLocked(r zoneRoute, ids []uuid.UUID) int {
	return len(g.routeAllLandedLocked(r, ids))
}

// routeAllLandedLocked is that body, reporting WHICH legs landed
// rather than how many. The count above is its length.
//
// #893: the mill is the caller that needs the list from the
// fire-and-forget form, because MillToZoneForEffect has returned the
// cards that moved since long before any of this existed and callers
// read them. The rule is the same one the count reports — a leg that
// landed where the route asked, and nothing else — so it is the same
// loop, not a second one.
//
// Caller must hold g.mu in write mode.
func (g *Game) routeAllLandedLocked(r zoneRoute, ids []uuid.UUID) []uuid.UUID {
	return g.routeAllLandedPerLegLocked(ids, func(uuid.UUID) zoneRoute { return r })
}

// routeAllLandedPerLegLocked is that body with the route chosen PER
// LEG rather than once for the batch.
//
// One caller needs it and it is the state-based-action sweep. A
// single SBA pass is one event (CR 704.3) and its doomed set mixes
// two rules that DESTROY a permanent (CR 704.5g lethal damage,
// CR 704.5h deathtouch) with three that merely put it into a
// graveyard (CR 704.5f zero toughness, CR 704.5i zero loyalty,
// CR 704.5v zero defense). They have to leave together — a Blood
// Artist must see the whole pass — and they have to leave by
// different routes, because a regeneration shield may replace the
// first two and must not touch the last three (#667).
//
// Everything else about the batch is unchanged: the pre-move copies
// are published once for the whole loop, and nothing here can pause.
//
// Caller must hold g.mu in write mode.
func (g *Game) routeAllLandedPerLegLocked(ids []uuid.UUID, routeFor func(uuid.UUID) zoneRoute) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	defer g.beginSimultaneousExitLocked(ids)()
	landed := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		r := routeFor(id)
		if g.routeLegNothingToDoLocked(r, id) {
			continue
		}
		if err := g.routeLegLocked(r, id, nil, nil); err != nil {
			continue
		}
		if g.routeLegLandedLocked(r, id) {
			landed = append(landed, id)
		}
	}
	return landed
}

// routeLegLocked routes ONE card of a batch, with `then` as the
// route's continuation and `batch` as the simultaneous-exit copies to
// re-publish around the move.
//
// The destroy / sacrifice / SBA exit keeps its own mover for the
// reasons zone_route.go lists, so it is the one leg that does not go
// through routeCardToZoneLocked; ViaBattlefieldLeave is the flag that
// says so, and it is already the flag the resume path reads.
//
// Caller must hold g.mu in write mode.
func (g *Game) routeLegLocked(r zoneRoute, id uuid.UUID, batch []Card, then func(g *Game) error) error {
	if r.ViaBattlefieldLeave {
		if r.Sacrifice {
			// Before the window, while the permanent is still there:
			// CR 701.17a's announcement is not part of the move, and a
			// "whenever you sacrifice" payoff reads characteristics
			// that are gone a line later (sacrifice.go).
			if err := g.announceSacrificeLocked(id, r.Source); err != nil {
				return err
			}
		}
		return g.routeBattlefieldExitInBatchThenLocked(id, r, batch, then)
	}
	r.CardID = id
	r.simultaneousExit = batch
	r.then = then
	_, err := g.routeCardToZoneLocked(r)
	return err
}

// routeLegNothingToDoLocked reports that a leg has no move to make, so
// the batch should step over it without opening a replacement window.
//
// Skipping rather than routing matters beyond tidiness: a card that is
// not where the route expects it would otherwise open a CR 614 window
// for a move that cannot happen, and a commander among them would be
// asked a question about it.
//
// Caller must hold g.mu.
func (g *Game) routeLegNothingToDoLocked(r zoneRoute, id uuid.UUID) bool {
	if r.ViaBattlefieldLeave {
		return findCardOnBattlefield(g, id) < 0
	}
	src := g.findCardZoneLocked(id)
	if src == nil {
		return true
	}
	dst, _, err := g.routeDestinationLocked(id, r.Dst, r.DstOwner)
	return err != nil || dst == nil || dst == src
}

// routeLegLandedLocked answers "did this leg go THIS way", the
// question every "for each … this way" clause is asking. One rule per
// verb, each declared in one place: CR 701.7a plus the command-zone
// carry-over for a destruction, CR 400.7's arrived-object for every
// other destination.
//
// Caller must hold g.mu.
func (g *Game) routeLegLandedLocked(r zoneRoute, id uuid.UUID) bool {
	if r.ViaBattlefieldLeave {
		if r.Sacrifice {
			return g.sacrificedThisWayLocked(id)
		}
		return g.destroyedThisWayLocked(id)
	}
	return g.landedInZoneLocked(id, r.Dst, r.DstOwner)
}

// destroyPermanentsLocked is the shared implementation behind the
// fire-and-forget entry point above and the state-based-action sweep
// in mutations.go, which has exactly the same simultaneity
// requirement: every creature that dies to one Pyroclasm or one Toxic
// Deluge dies at the same time as the others.
//
// It does NOT filter indestructible: both callers hand it a set that
// has already been narrowed by the rule that applies to them, and
// they are different rules. The effect path drops every
// indestructible permanent (CR 702.12b); the SBA path drops them only
// from the two damage-driven branches, because CR 704.5f / 704.5i /
// 704.5v put a permanent into a graveyard rather than destroying it
// and indestructible is no help there.
//
// #815: the count is taken from the LANDED outcome
// (destroyedThisWayLocked) rather than from "the call returned no
// error", which counted a destruction the CR 614 window had cancelled
// and a leg that had merely paused. The SBA sweep never read it as a
// destruction count — its question is "did this pass do anything" —
// and it no longer reads it at all.
//
// #866: the loop itself is the shared fire-and-forget body below, with
// the destroy route's template. What used to be three copies of it —
// here, in the exile sweep and in the bounce sweep — is one.
//
// Caller must hold g.mu in write mode.
func (g *Game) destroyPermanentsLocked(ids []uuid.UUID, opts DestroyOptions) int {
	return g.routeAllLocked(destroyRouteWith(opts), ids)
}

// doomedPermanent is one entry of the state-based-action sweep's
// doomed set: the permanent, and whether the rule that doomed it
// DESTROYS it (CR 704.5g lethal damage, CR 704.5h deathtouch) or
// merely puts it into a graveyard (CR 704.5f, CR 704.5i, CR 704.5v).
//
// The distinction was invisible before regeneration, because all five
// end in the same graveyard by the same route. It is visible now: a
// regeneration shield replaces the first two and does nothing about
// the last three, and a 2/2 under two -1/-1 counters with a shield on
// it dies (#667).
type doomedPermanent struct {
	id          uuid.UUID
	destruction bool
}

// sweepDoomedPermanentsLocked performs one state-based-action pass's
// collected doomed set as ONE simultaneous event (CR 704.3), each
// permanent by the route its own rule asks for.
//
// The SBA's half of what destroyPermanentsLocked used to do for both
// callers. It is a separate function rather than a flag on that one
// because the two callers are answering different questions: the
// effect path destroys a set it has already narrowed with one rule
// (CR 702.12b indestructible), and this path leaves with five.
//
// Caller must hold g.mu in write mode.
func (g *Game) sweepDoomedPermanentsLocked(doomed []doomedPermanent) {
	if len(doomed) == 0 {
		return
	}
	ids := make([]uuid.UUID, 0, len(doomed))
	destruction := make(map[uuid.UUID]bool, len(doomed))
	for _, d := range doomed {
		ids = append(ids, d.id)
		destruction[d.id] = d.destruction
	}
	g.routeAllLandedPerLegLocked(ids, func(id uuid.UUID) zoneRoute {
		if destruction[id] {
			return destroyRoute
		}
		return battlefieldExitRoute
	})
}

// ExileCardsForEffect exiles every card in `ids` as one simultaneous
// event and returns how many of them reached exile. Same batch as the
// destroy sweep, for the same reason: a "whenever a permanent leaves
// the battlefield" watcher exiled by Farewell still sees the rest of
// the board go with it.
//
// Cards outside the battlefield (a graveyard sweep) are exiled too
// and counted; they simply are not part of the leaves-the-battlefield
// batch, which is what the copies in beginSimultaneousExitLocked
// already encode — it only picks up battlefield residents.
//
// The FIRE-AND-FORGET form, and #866 made its count mean what the
// destroy sweep's has meant since #815: the LANDED outcome, not "the
// call returned no error". A leg the CR 614 window cancelled is not
// in it, nor is a commander CR 903.9 sent to the command zone instead
// — that card left, but not to exile. A leg that merely PAUSED cannot
// be in it either, which is exactly why a card that reads the number
// uses ExileCardsThenForEffect, which waits.
//
// Caller must hold g.mu in write mode.
func (g *Game) ExileCardsForEffect(ids []uuid.UUID) int {
	return g.routeAllLocked(exileRoute, ids)
}

// ExileCardsThenForEffect exiles every card in `ids` as one
// simultaneous event and hands `then` the ones that actually reached
// EXILE — "for each card exiled this way" (Settle the Wreckage's basic
// lands, Oona-shaped payoffs).
//
// The `Then` half of the same pair DestroyPermanentsThenForEffect is
// the destroy side of (ADR 0013 §5i, §5k), built on the same body for
// the same reason: any leg can pause on the CR 903.9 prompt, so the
// number is not knowable on the line after the sweep. The legs are
// exiled IN SEQUENCE, each from the previous one's continuation, with
// the landed list carried forward by value.
//
// CR 400.7 decides what "this way" means: the object that ARRIVED in
// exile is the one the effect exiled. A leg a replacement sent
// somewhere else — a commander taking CR 903.9's offer, Stone of
// Erech's graveyard rewrite — left, but not to exile, so it is not in
// the list. That is the one place this differs from the destroy side,
// which counts the command zone by a declared carry-over (§5i).
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileCardsThenForEffect(ids []uuid.UUID, then func(g *Game, exiled []uuid.UUID) error) error {
	return g.routeAllThenLocked(exileRoute, ids, then)
}

// ExileCardThenForEffect is the SINGLE-CARD form: exile one card and
// tell `then` whether it actually reached exile.
//
// #870. A one-card read-back is the same bug as a batch one and it
// arrived by the same route: ExileCardForEffect returns nil when the
// leg merely PAUSED on the CR 903.9 prompt, so "exile it with a hit
// counter on it" put the counter on a card still in its owner's
// graveyard. The answer is not a second exile path with its own notion
// of what landed — that is how exile and destroy drifted apart in the
// first place (#815, #866) — so this is a WRAPPER over the batch,
// which already sequences the pause, carries the landed list across it
// and replays identically under an undo. A batch of one publishes a
// one-card simultaneous exit, which no watcher can observe: the
// harvest skips the card whose own move is being reported.
//
// `exiled` is CR 400.7's reading, the batch's: true when the card is
// in exile, false when the window cancelled the move, when a
// replacement sent it somewhere else, and when a commander took
// CR 903.9's offer — it left, but not to exile.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileCardThenForEffect(cardID uuid.UUID, then func(g *Game, exiled bool) error) error {
	return g.ExileCardsThenForEffect([]uuid.UUID{cardID}, func(g *Game, landed []uuid.UUID) error {
		if then == nil {
			return nil
		}
		return then(g, len(landed) == 1)
	})
}

// SacrificeAllThenForEffect sacrifices every permanent in `ids` as one
// simultaneous exit and hands `then` the ones that were actually
// SACRIFICED — "then draw a card for each permanent sacrificed this
// way" (God-Eternal Bontu), and every "sacrifice all X, then for
// each …" that could not be written before.
//
// #910. Destroy, exile, bounce, tuck and mill have all had this since
// #815 / #866 / #893; sacrifice had only the per-card call, so Living
// Death's second pass fired and forgot and any follow-up ran on the
// next line with a commander's CR 903.9 prompt still open.
//
// Sacrifice is not replaceable (CR 701.17a: sacrificing doesn't
// destroy, so nothing that replaces destruction touches it, and no
// replacement stops the sacrifice itself) — but the MOVE it makes is
// an ordinary zone change, so a sacrificed commander opens the CR 903.9
// window and a leg can PAUSE exactly like a destroy leg. That is the
// whole reason this is a continuation rather than a return value.
//
// `sacrificed` is sacrificedThisWayLocked's answer: the permanents that
// really left the battlefield. A commander that took the command zone
// IS in it — it was sacrificed, and only where the card went was
// replaced — and so is one an "exile it instead" replacement took.
//
// `source` is the card that asked, stamped on each EventSacrifice.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) SacrificeAllThenForEffect(source uuid.UUID, ids []uuid.UUID, then func(g *Game, sacrificed []uuid.UUID) error) error {
	return g.routeAllThenLocked(sacrificeRoute(source), ids, then)
}

// SacrificeThenForEffect is the SINGLE-CARD form: sacrifice one
// permanent and tell `then` whether it really left the battlefield.
//
// ExileCardThenForEffect's twin (#870), and a WRAPPER over the batch
// for the reason that one gives: a one-card read-back is the same
// problem as a batch one — the one leg is the leg that can pause — and
// a second sacrifice path with its own notion of what happened is how
// two verbs drift. "Sacrifice a creature. If you do, …" is this call.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) SacrificeThenForEffect(source, cardID uuid.UUID, then func(g *Game, sacrificed bool) error) error {
	return g.SacrificeAllThenForEffect(source, []uuid.UUID{cardID}, func(g *Game, landed []uuid.UUID) error {
		if then == nil {
			return nil
		}
		return then(g, len(landed) == 1)
	})
}

// SacrificeAllForEffect is the FIRE-AND-FORGET batch: "each player
// sacrifices all permanents they control that are one or more colors"
// (All Is Dust), with nothing waiting on the answer.
//
// What it buys over a loop of single sacrifices is the one thing a
// loop cannot have: the permanents leave as one SIMULTANEOUS exit, so
// a Blood Artist swept by the same spell sees every death including
// its own (CR 603.10 last-known information). That was All Is Dust's
// declared caveat until this existed.
//
// The count is how many of the legs that SETTLED were sacrificed. A
// leg that paused on the CR 903.9 prompt cannot be in it — nothing has
// left yet — which is why a card that READS the number uses
// SacrificeAllThenForEffect.
//
// Caller must hold g.mu in write mode.
func (g *Game) SacrificeAllForEffect(source uuid.UUID, ids []uuid.UUID) int {
	return g.routeAllLocked(sacrificeRoute(source), ids)
}

// BounceCardsToHandForEffect returns every card in `ids` to its
// owner's hand as one simultaneous event and returns how many of them
// reached a hand. Evacuation, an overloaded Cyclonic Rift, Whelming
// Wave.
//
// The FIRE-AND-FORGET form, counting the landed outcome for the
// reasons ExileCardsForEffect gives; a card that reads the number uses
// BounceCardsToHandThenForEffect.
//
// Caller must hold g.mu in write mode.
func (g *Game) BounceCardsToHandForEffect(ids []uuid.UUID) int {
	return g.routeAllLocked(bounceRoute, ids)
}

// BounceCardsToHandThenForEffect returns every card in `ids` to its
// owner's hand as one simultaneous event and hands `then` the ones
// that actually reached a HAND — "for each permanent returned this
// way".
//
// ExileCardsThenForEffect's twin, and the same CR 400.7 reading: a
// commander that took CR 903.9's offer went to the command zone, not
// to a hand, so it was not returned this way. An owner who has left
// the table has no hand to return to, and their card is skipped rather
// than counted.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) BounceCardsToHandThenForEffect(ids []uuid.UUID, then func(g *Game, bounced []uuid.UUID) error) error {
	return g.routeAllThenLocked(bounceRoute, ids, then)
}

// TuckCardsToLibraryThenForEffect puts every card in `ids` into its
// OWNER's library as one simultaneous exit and hands `then` the ones
// that actually reached a LIBRARY — Aetherspouts' "put all attacking
// creatures on top or bottom of their owners' libraries", after which
// each owner orders only the cards that arrived.
//
// ExileCardsThenForEffect's twin (ADR 0013 §5k, §5n), on the same body,
// with the same CR 400.7 reading of "this way": a commander that took
// CR 903.9's offer went to the command zone, not to a library, so it is
// not in the list and nothing downstream may count it.
//
// #783 is the reason the `Then` half exists at all. A library is a
// CR 903.9 destination, so EVERY tuck can pause — and until this
// existed the fire-and-forget form was the only one, so a card that had
// more to do after the tuck (Chaos Warp's shuffle and reveal,
// Aetherspouts' scry, Sylvan Library's next question) did it on the
// next line, with the card still on the battlefield and the question
// still open.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) TuckCardsToLibraryThenForEffect(ids []uuid.UUID, opts TuckOptions, then func(g *Game, tucked []uuid.UUID) error) error {
	return g.routeAllThenLocked(tuckRoute(opts), ids, then)
}

// TuckToLibraryThenForEffect is the SINGLE-CARD form: tuck one card and
// tell `then` whether it actually reached a library.
//
// ExileCardThenForEffect's twin (#870), and a WRAPPER over the batch
// for the reason that one gives: a one-card read-back is the same bug
// as a batch one, and a second tuck path with its own notion of what
// landed is how two verbs drift. A batch of one publishes a one-card
// simultaneous exit, which no watcher can observe.
//
// `tucked` is false when the CR 614 window cancelled the move, when a
// replacement sent the card somewhere else, and when a commander took
// CR 903.9's offer — it left, but not to a library. The continuation
// runs on every one of those outcomes, because a caller that is waiting
// has to be told even when the answer is "nothing happened".
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) TuckToLibraryThenForEffect(cardID uuid.UUID, opts TuckOptions, then func(g *Game, tucked bool) error) error {
	return g.TuckCardsToLibraryThenForEffect([]uuid.UUID{cardID}, opts, func(g *Game, landed []uuid.UUID) error {
		if then == nil {
			return nil
		}
		return then(g, len(landed) == 1)
	})
}
