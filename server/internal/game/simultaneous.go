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
//   - Regeneration is not modelled anywhere in the engine
//     (keywords.go's canonical set is closed and does not contain
//     it), so "they can't be regenerated" is still cosmetic. A batch
//     changes nothing about that.
//   - Totem armor, likewise absent.
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
		// CatalogAbilityKey, and it answers off the copy the batch
		// captured while the card was still on the battlefield — so
		// a creature wiped while under a Kenrith's Transformation
		// has no dies-trigger here either, for the same CR 603.10
		// reason harvestLTB reads its snapshot.
		oracle := CatalogAbilityKey(card)
		if oracle == "" {
			continue
		}
		triggers := CatalogTriggers(oracle)
		if len(triggers) == 0 {
			continue
		}
		lki := card.Effective()
		for _, t := range triggers {
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
func (g *Game) DestroyPermanentsForEffect(ids []uuid.UUID) int {
	return g.destroyPermanentsLocked(g.DestructibleForEffect(ids))
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
func (g *Game) DestroyPermanentsThenForEffect(ids []uuid.UUID, then func(g *Game, destroyed []uuid.UUID) error) error {
	return g.routeAllThenLocked(destroyRoute, g.DestructibleForEffect(ids), then)
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
var (
	destroyRoute = zoneRoute{ViaBattlefieldLeave: true}
	exileRoute   = zoneRoute{Dst: ZoneExile}
	bounceRoute  = zoneRoute{Dst: ZoneHand}
)

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
	if len(ids) == 0 {
		return nil
	}
	defer g.beginSimultaneousExitLocked(ids)()
	landed := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
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
		return g.routeBattlefieldExitInBatchThenLocked(id, batch, then)
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
func (g *Game) destroyPermanentsLocked(ids []uuid.UUID) int {
	return g.routeAllLocked(destroyRoute, ids)
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
