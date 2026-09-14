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
	batch := make([]Card, 0, len(ids))
	for _, id := range ids {
		if i := findCardOnBattlefield(g, id); i >= 0 {
			batch = append(batch, g.Battlefield.Cards[i])
		}
	}
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
		oracle := CatalogKey(card)
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
// simultaneous event and returns how many actually left the
// battlefield (or, for a commander, had their zone change replaced).
// The count is what "for each creature destroyed this way" clauses
// read — Fumigate's life gain, Deadly Tempest's life loss, Bane of
// Progress's counters.
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

// destroyPermanentsLocked is the shared implementation behind the
// effect-facing entry point above and the state-based-action sweep in
// mutations.go, which has exactly the same simultaneity requirement:
// every creature that dies to one Pyroclasm or one Toxic Deluge dies
// at the same time as the others.
//
// It does NOT filter indestructible: both callers hand it a set that
// has already been narrowed by the rule that applies to them, and
// they are different rules. The effect path drops every
// indestructible permanent (CR 702.12b); the SBA path drops them only
// from the two damage-driven branches, because CR 704.5f / 704.5i /
// 704.5p put a permanent into a graveyard rather than destroying it
// and indestructible is no help there.
//
// Caller must hold g.mu in write mode.
func (g *Game) destroyPermanentsLocked(ids []uuid.UUID) int {
	if len(ids) == 0 {
		return 0
	}
	defer g.beginSimultaneousExitLocked(ids)()
	destroyed := 0
	for _, id := range ids {
		if err := g.routeBattlefieldCardToOwnerGraveyardLocked(id); err == nil {
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
