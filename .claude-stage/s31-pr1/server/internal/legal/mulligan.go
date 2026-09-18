package legal

import "fmt"

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
