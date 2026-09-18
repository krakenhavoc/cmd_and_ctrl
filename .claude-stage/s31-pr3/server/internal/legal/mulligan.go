package legal

import (
	"fmt"

	"github.com/google/uuid"
)

type mulliganParams struct {
	HandSize int `json:"hand_size"`
}

// mulliganMoves enumerates the opening-hand window: keep, or
// mulligan. The engine's mulligan is sandbox-shaped (the player
// names the new hand size); we offer the multiplayer London rule
// (CR 103.5b): the first mulligan is free (seven again), each
// further one is a card fewer.
func (e *enumerator) mulliganMoves() {
	p := e.p
	if p.HandKept {
		return
	}
	e.add(Move{
		Type:   TypeKeepHand,
		Player: e.seat,
		Kind:   KindMulligan,
		Label:  "Keep hand",
	})
	next := 7 - p.MulligansTaken
	if next < 0 {
		next = 0
	}
	if next > 0 {
		e.add(Move{
			Type:   TypeMulligan,
			Player: e.seat,
			Kind:   KindMulligan,
			Label:  fmt.Sprintf("Mulligan to %d", next),
			Params: mustJSON(mulliganParams{HandSize: next}),
		})
	}
}

type discardSelectionParams struct {
	CardIDs []string `json:"card_ids"`
}

// cleanupDiscardMoves enumerates the cleanup-step discard owed by this
// seat (Game.DiscardPending), one move per hand subset of the owed
// size, capped. Returns true when a discard is owed.
func (e *enumerator) cleanupDiscardMoves() bool {
	n, owed := e.g.DiscardPending[e.seat]
	if !owed || n <= 0 || e.p.Hand == nil {
		return owed && n > 0
	}
	pool := make([]uuid.UUID, 0, len(e.p.Hand.Cards))
	for _, c := range e.p.Hand.Cards {
		pool = append(pool, c.InstanceID)
	}
	for _, set := range combinations(pool, n, n, e.opts.MaxExpansionPerSource) {
		label := "Discard to hand size:"
		for _, id := range set {
			label += " " + cardName(e.g, id)
		}
		e.add(Move{
			Type:   TypeDiscardSelection,
			Player: e.seat,
			Kind:   KindChoice,
			Label:  label,
			Params: mustJSON(discardSelectionParams{CardIDs: idStrings(set)}),
		})
	}
	return true
}
