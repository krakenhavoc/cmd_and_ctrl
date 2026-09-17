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
func (g *Game) harvestSimultaneousExitLocked(ev Event) {
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
			g.dispatchTriggerLocked(ev, card, lki, t)
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
	ids = g.DestructibleForEffect(ids)
	return g.destroyEachStepLocked(g.simultaneousExitSnapshotLocked(ids), ids, nil, then)
}

// destroyEachStepLocked destroys the head of `ids` and continues with
// the tail from that leg's continuation, carrying the landed list
// forward by value. The empty list is the base case: the sweep is done
// and `then` gets the list.
//
// `batch` is the pre-move copy set taken before the first move; it is
// re-published on every step so a leg that runs on the far side of a
// CR 903.9 prompt is still part of the same simultaneous exit.
//
// Caller must hold g.mu in write mode.
func (g *Game) destroyEachStepLocked(
	batch []Card,
	ids, destroyed []uuid.UUID,
	then func(g *Game, destroyed []uuid.UUID) error,
) error {
	// A card that is no longer on the battlefield has nothing to
	// destroy — it left while an earlier leg was paused, or it was
	// never there. Skipped, not counted, exactly as the fire-and-forget
	// loop skips it.
	for len(ids) > 0 && findCardOnBattlefield(g, ids[0]) < 0 {
		ids = ids[1:]
	}
	if len(ids) == 0 {
		if then == nil {
			return nil
		}
		return then(g, destroyed)
	}
	next, rest := ids[0], ids[1:]
	closeBatch := g.publishSimultaneousExitLocked(batch)
	defer closeBatch()
	return g.routeBattlefieldExitThenLocked(next, func(g *Game) error {
		landed := destroyed
		if g.destroyedThisWayLocked(next) {
			// A fresh slice rather than an append in place: two runs of
			// the same continuation (an undo, then the same answer
			// again) must not see each other's entry.
			landed = append(append(make([]uuid.UUID, 0, len(destroyed)+1), destroyed...), next)
		}
		return g.destroyEachStepLocked(batch, rest, landed, then)
	})
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
// Caller must hold g.mu in write mode.
func (g *Game) destroyPermanentsLocked(ids []uuid.UUID) int {
	if len(ids) == 0 {
		return 0
	}
	defer g.beginSimultaneousExitLocked(ids)()
	destroyed := 0
	for _, id := range ids {
		if err := g.routeBattlefieldCardToOwnerGraveyardLocked(id); err != nil {
			continue
		}
		if g.destroyedThisWayLocked(id) {
			destroyed++
		}
	}
	return destroyed
}

// ExileCardsForEffect exiles every card in `ids` as one simultaneous
// event and returns how many left the battlefield. Same batch as the
// destroy sweep, for the same reason: a "whenever a permanent leaves
// the battlefield" watcher exiled by Farewell still sees the rest of
// the board go with it.
//
// Cards outside the battlefield (a graveyard sweep) are exiled too
// and counted; they simply are not part of the leaves-the-battlefield
// batch, which is what the copies in beginSimultaneousExitLocked
// already encode — it only picks up battlefield residents.
//
// Caller must hold g.mu in write mode.
func (g *Game) ExileCardsForEffect(ids []uuid.UUID) int {
	if len(ids) == 0 {
		return 0
	}
	defer g.beginSimultaneousExitLocked(ids)()
	exiled := 0
	for _, id := range ids {
		if err := g.ExileCardForEffect(id); err == nil {
			exiled++
		}
	}
	return exiled
}

// BounceCardsToHandForEffect returns every card in `ids` to its
// owner's hand as one simultaneous event and returns how many moved.
// Evacuation, an overloaded Cyclonic Rift, Whelming Wave.
//
// Caller must hold g.mu in write mode.
func (g *Game) BounceCardsToHandForEffect(ids []uuid.UUID) int {
	if len(ids) == 0 {
		return 0
	}
	defer g.beginSimultaneousExitLocked(ids)()
	bounced := 0
	for _, id := range ids {
		if err := g.BounceToHandForEffect(id); err == nil {
			bounced++
		}
	}
	return bounced
}
