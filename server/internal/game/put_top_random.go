package game

import "github.com/google/uuid"

// put_top_random.go — "shuffle that pile and put it back on top of your
// library" (Triumph of Saint Katherine, S82 GAP batch 196, #5948): the
// top-of-library twin of PutOnBottomInRandomOrderForEffect.

// PutOnTopInRandomOrderForEffect puts the cards `ids`, all in zone `from`, on
// top of their owners' libraries in a random order drawn from the actor's
// random-order stream (the same stream the bottom twin uses, so a rewind
// reproduces the order). Each card is a real zone change through the exit path
// (the CR 614 window opens and a zone move is emitted), so it forgets what it
// was in `from`: an exiled card loses its exile grant and its knowers
// (CR 400.7). Cards no longer in `from` are skipped, a repeated ID is moved
// once, and an error on one card does not stop the rest; the first is
// returned.
//
// `actor` is the player the effect belongs to, stamped on the zone-move
// events.
//
// Caller must hold g.mu.
func (g *Game) PutOnTopInRandomOrderForEffect(actor uuid.UUID, from ZoneKind, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	order := make([]uuid.UUID, 0, len(ids))
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			order = append(order, id)
		}
	}
	swap := func(i, j int) { order[i], order[j] = order[j], order[i] }
	g.randForLocked(rngStream{kind: rngStreamRandomOrder, player: actor}).Shuffle(len(order), swap)
	var firstErr error
	for _, id := range order {
		src := g.findCardZoneLocked(id)
		if src == nil || src.Kind != from {
			continue
		}
		if _, err := g.routeCardToZoneLocked(zoneRoute{
			CardID: id,
			Dst:    ZoneLibrary,
			Actor:  actor,
		}); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
