package game

import "github.com/google/uuid"

// route_batch_keyed.go — #4253 slice 9: a route batch with the rest of the
// batch as data, so a leg paused on a CR 616 prompt is a restore point.
//
// routeAllThenLocked routes a batch one leg at a time, each from the
// previous leg's continuation (zoneRoute.then), so a leg paused on a CR
// 616 ordering prompt or a "may" replacement's question holds the rest of
// the batch behind it. That continuation was a closure over the remaining
// cards, the landed list and the caller's `then`, and the leg's route also
// carried the batch's last-known information (zoneRoute.simultaneousExit,
// CR 603.10) as live Card values. Neither could be written to a restore
// point, so the frame was counted and the table was not one.
//
// The keyed form is slice 3's face-down exile batch (faceDownLegs)
// generalised to every route verb: the run rides the route of the leg now
// moving (zoneRoute.batchLegs) and holds the leg, the cards still to
// route, the cards that have landed and the caller's continuation as a
// registered tail. The route of the next leg is the current leg's route
// with the card cleared (the template a batch was started with, plus the
// cause routeCardToZoneLocked captured on the first leg, which is the same
// for every leg). The last-known information is the route's own
// simultaneousExit, written through the card mirror beside the run
// (zoneRouteSnapshot.SimultaneousExit), and a frame whose copies the
// mirror cannot write whole stays off the disk.

// routeBatchRun is the rest of a keyed route batch, riding the route of
// the leg now moving (Leg). Never written after it is built, so an undo
// snapshot may share it: each step builds a new one.
type routeBatchRun struct {
	Leg    uuid.UUID   `json:"leg"`
	Rest   []uuid.UUID `json:"rest,omitempty"`
	Landed []uuid.UUID `json:"landed,omitempty"`
	// Then is the caller's continuation, told the cards that landed where
	// the route asked (CR 400.7), in batch order, as TailOutcome.IDs.
	Then *keyedTailSnapshot `json:"then,omitempty"`
	// Planned and AllMoved are a mill's "those cards" (CR 701.17c,
	// MillToZoneKeyedMovedForEffect): when AllMoved is set the tail is
	// told every planned card that left its library, wherever it went,
	// instead of the landed list.
	Planned  []uuid.UUID `json:"planned,omitempty"`
	AllMoved bool        `json:"allMoved,omitempty"`
	// Mill is the "mills a card, then repeats this process until …" run
	// one repetition of this batch belongs to (millUntilRunKeyed). Nil
	// for every other batch.
	Mill *millUntilRun `json:"mill,omitempty"`
}

func cloneRouteBatchRun(r *routeBatchRun) *routeBatchRun {
	if r == nil {
		return nil
	}
	out := *r
	out.Rest = copyUUIDs(r.Rest)
	out.Landed = copyUUIDs(r.Landed)
	out.Planned = copyUUIDs(r.Planned)
	if r.Then != nil {
		t := snapshotKeyedTail(restoreKeyedTail(r.Then))
		out.Then = t
	}
	out.Mill = cloneMillUntilRun(r.Mill)
	return &out
}

// routeAllKeyedLocked is routeAllThenLocked with the continuation as data:
// `then` is a registered tail (RegisterTail), told the cards that landed
// where the route asked as TailOutcome.IDs. A leg's prompt is then a
// restore point.
//
// Caller must hold g.mu in write mode.
func (g *Game) routeAllKeyedLocked(r zoneRoute, ids []uuid.UUID, then KeyedTail) error {
	return g.routeAllRunLocked(r, ids, &routeBatchRun{Then: snapshotKeyedTail(then)})
}

// routeAllRunLocked starts a keyed batch over ids with run's continuation.
//
// Caller must hold g.mu in write mode.
func (g *Game) routeAllRunLocked(r zoneRoute, ids []uuid.UUID, run *routeBatchRun) error {
	start := cloneRouteBatchRun(run)
	start.Rest = copyUUIDs(ids)
	return g.routeBatchStepKeyedLocked(r, g.simultaneousExitSnapshotLocked(ids), start)
}

// routeBatchStepKeyedLocked routes the head of run.Rest, with the rest of
// the batch riding the leg's route, or finishes the batch.
//
// Caller must hold g.mu in write mode.
func (g *Game) routeBatchStepKeyedLocked(tmpl zoneRoute, batch []Card, run *routeBatchRun) error {
	ids := run.Rest
	// A leg with nothing to do is skipped and not counted, as in
	// routeEachStepLocked.
	for len(ids) > 0 && g.routeLegNothingToDoLocked(tmpl, ids[0]) {
		ids = ids[1:]
	}
	if len(ids) == 0 {
		return g.finishRouteBatchKeyedLocked(run, batch)
	}
	leg := cloneRouteBatchRun(run)
	leg.Leg = ids[0]
	leg.Rest = copyUUIDs(ids[1:])
	tmpl.batchLegs = leg
	return g.routeLegLocked(tmpl, ids[0], batch, nil)
}

// runRouteBatchLegLocked counts the leg that just settled and routes the
// next one: the keyed form of routeEachStepLocked's leg closure. r is the
// settled leg's route, and the next leg's route is it with the card
// cleared.
//
// Caller must hold g.mu in write mode.
func (g *Game) runRouteBatchLegLocked(r *zoneRoute, run *routeBatchRun) error {
	tmpl := *r
	tmpl.CardID = uuid.Nil
	tmpl.then = nil
	tmpl.batchLegs = nil
	tmpl.simultaneousExit = nil
	tmpl.simultaneousEntry = nil
	next := cloneRouteBatchRun(run)
	next.Leg = uuid.Nil
	if g.routeLegLandedLocked(tmpl, run.Leg) {
		next.Landed = append(next.Landed, run.Leg)
	}
	return g.routeBatchStepKeyedLocked(tmpl, r.simultaneousExit, next)
}

// finishRouteBatchKeyedLocked runs a finished batch's continuation: the
// caller's tail, or the next repetition of the mill run it belongs to,
// which reads the batch's pre-move copies.
//
// Caller must hold g.mu in write mode.
func (g *Game) finishRouteBatchKeyedLocked(run *routeBatchRun, batch []Card) error {
	landed := run.Landed
	if run.AllMoved {
		landed = g.leftLibraryLocked(run.Planned)
	}
	if run.Mill != nil {
		return g.continueMillUntilRunKeyedLocked(run.Mill, landed)
	}
	// The batch is open around the tail, as it is around the closure
	// form's (executeZoneRouteLocked registers the publish first), so a
	// tail that reads the cards that left (SimultaneousExitCardsForEffect)
	// reads them from any terminal outcome of the last leg.
	defer g.publishSimultaneousExitLocked(batch)()
	return g.runKeyedTailLocked(restoreKeyedTail(run.Then), landed, 0)
}

// routeBatchLKIIsData reports whether a keyed batch's last-known copies
// can be written whole by the card mirror.
func routeBatchLKIIsData(batch []Card) bool {
	for _, c := range batch {
		var scratch ContinuationCensus
		_ = snapshotCard(c, &scratch)
		if !scratch.Empty() {
			return false
		}
	}
	return true
}

// unknownRouteBatchKeys is checkEffectKeys' half for a keyed batch.
func unknownRouteBatchKeys(r *routeBatchRun) []string {
	if r == nil {
		return nil
	}
	out := unknownKeyedTail(r.Then)
	return append(out, unknownMillUntilKeys(r.Mill)...)
}

// --- the verbs ---------------------------------------------------------

// ExileCardsKeyedForEffect is ExileCardsThenForEffect with the
// continuation as data (#4253 slice 9): `then` is told the cards that
// reached exile, in batch order, as TailOutcome.IDs. A zero tail is "no
// continuation". A leg's prompt is a restore point.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileCardsKeyedForEffect(ids []uuid.UUID, then KeyedTail) error {
	return g.routeAllKeyedLocked(exileRoute, ids, then)
}

// ExileCardKeyedForEffect is the single-card form: `then` is told the card
// when it reached exile, nothing when it did not.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileCardKeyedForEffect(cardID uuid.UUID, then KeyedTail) error {
	return g.ExileCardsKeyedForEffect([]uuid.UUID{cardID}, then)
}

// ExileLinkedKeyedForEffect is ExileLinkedThenForEffect with the
// continuation as data.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileLinkedKeyedForEffect(source ObjectRef, ids []uuid.UUID, then KeyedTail) error {
	r := exileRoute
	r.ExiledWith = PermissionCardRef(source)
	return g.routeAllKeyedLocked(r, ids, then)
}

// PutCardsIntoGraveyardKeyedForEffect puts every card in `ids` into its
// owner's graveyard as one simultaneous exit and tells `then` (a registered
// tail) the ones that actually reached the graveyard — Valakut Exploration's
// "put them into their owner's graveyard, then this enchantment deals that
// much damage" (#1218) needs the count AFTER the move lands, not before: a
// leg can pause on a replacement's prompt exactly as an exile or a destroy
// leg can (ADR 0013 §5t), so a caller that read back a pre-move tally would
// pay out for a move the window has not answered yet.
// PutIntoGraveyardForEffect (random_bottom.go) is the same exit with no
// continuation. (The closure form, PutCardsIntoGraveyardThenForEffect, was
// deleted in #4253's cleanup once nothing called it.)
//
// Refuses any id still on the battlefield (ErrInvalidParam), for the same
// reason PutIntoGraveyardForEffect does: leaving the battlefield for a
// graveyard is a destroy, a sacrifice or a state-based action, each with its
// own path and its own event, and none of them is this.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) PutCardsIntoGraveyardKeyedForEffect(ids []uuid.UUID, then KeyedTail) error {
	for _, id := range ids {
		if z := g.findCardZoneLocked(id); z != nil && z.Kind == ZoneBattlefield {
			return ErrInvalidParam
		}
	}
	return g.routeAllKeyedLocked(graveyardRoute, ids, then)
}

// DestroyPermanentsKeyedForEffect destroys every permanent in `ids` as one
// simultaneous event and tells `then` (a registered tail) the ones that
// were actually DESTROYED, as TailOutcome.IDs — "for each creature
// destroyed this way" (Fumigate's life gain, Deadly Tempest's life loss,
// Bane of Progress's counters, Blood Money's treasures).
//
// It is a continuation rather than a return value because any leg can
// pause on a replacement's prompt and the number is not knowable until it
// is answered: the legs are destroyed in sequence, the rest of the batch
// riding each leg as data, so a paused leg's prompt is a restore point.
// The pre-move copies are taken once, before the first move, so a Blood
// Artist still sees the whole board die with it. Indestructible is
// filtered first, exactly as the fire-and-forget form does. (The closure
// form, DestroyPermanentsThenForEffect, was deleted in #4253's last part
// once only tests called it.)
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) DestroyPermanentsKeyedForEffect(ids []uuid.UUID, then KeyedTail, opts ...DestroyOptions) error {
	return g.routeAllKeyedLocked(destroyRouteWith(firstDestroyOptions(opts)), g.hostsBeforeAttachmentsLocked(g.DestructibleForEffect(ids)), then)
}

// SacrificeAllKeyedForEffect is SacrificeAllThenForEffect with the
// continuation as data.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) SacrificeAllKeyedForEffect(source uuid.UUID, ids []uuid.UUID, then KeyedTail) error {
	return g.routeAllKeyedLocked(sacrificeRoute(source), g.sacrificeableLocked(g.unprotectedFromSacrificeLocked(ids)), then)
}

// BounceCardsToHandKeyedForEffect is BounceCardsToHandThenForEffect with
// the continuation as data. It is also the single-card "return it, THEN …"
// (#993; a batch of one, since #4253's cleanup deleted the closure wrapper
// BounceToHandThenForEffect): a hand is a CR 903.9 destination, so every
// bounce can pause, and the tail is told the card only when it reached a
// hand — not when the CR 614 window cancelled the move, a replacement sent
// it elsewhere, or a commander took CR 903.9's offer. The tail runs on
// every one of those outcomes.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) BounceCardsToHandKeyedForEffect(ids []uuid.UUID, then KeyedTail) error {
	return g.routeAllKeyedLocked(bounceRoute, ids, then)
}

// TuckCardsToLibraryKeyedForEffect is TuckCardsToLibraryThenForEffect with
// the continuation as data.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) TuckCardsToLibraryKeyedForEffect(ids []uuid.UUID, opts TuckOptions, then KeyedTail) error {
	return g.routeAllKeyedLocked(tuckRoute(opts), ids, then)
}

// SimultaneousExitCardsForEffect returns the pre-move copies (CR 603.10)
// of the cards in ids that belong to the batch now leaving the battlefield,
// in ids order: what a keyed batch's tail reads where the closure form
// read the copies it had captured ("for each creature destroyed this way,
// its controller …"). A card that was not on the battlefield has none.
//
// Caller must hold g.mu.
func (g *Game) SimultaneousExitCardsForEffect(ids []uuid.UUID) []Card {
	var out []Card
	for _, id := range ids {
		if c, ok := g.simultaneousExitCardLocked(id); ok {
			out = append(out, c)
		}
	}
	return out
}

// BodyRefByKey returns the registered effect body under key, for a keyed
// tail that names one in its data (a delayed trigger's body to schedule
// after an exile). False when no body is registered under it.
func BodyRefByKey(key string) (BodyRef, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	if _, ok := effectBodies[key]; !ok {
		return BodyRef{}, false
	}
	return BodyRef{key: key}, true
}
