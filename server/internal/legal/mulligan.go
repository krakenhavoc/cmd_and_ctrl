package legal

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

type mulliganParams struct {
	HandSize int `json:"hand_size"`
}

// mulliganMoves enumerates the opening-hand window: keep, or
// mulligan. The London mulligan (CR 103.5, ADR 0168): a mulligan
// draws seven, and a keep puts the counted mulligans' worth of cards
// on the bottom, the first mulligan of a multiplayer game free
// (CR 103.5c). The labels say what each costs. The mulligan's
// hand_size is the opening hand it leads to (seven less what it would
// bottom), which is what the bot's keep rules reason about; the engine
// ignores it inside the window.
func (e *enumerator) mulliganMoves() {
	p := e.p
	if p.HandKept {
		return
	}
	// CR 103.5: decisions go in turn order, starting player first. A
	// seat that is not the one deciding has nothing to offer; the
	// engine refuses its keep and mulligan with ErrNotYourMulligan
	// (#544: the enumerator and the engine answer the same question).
	if e.g.MulliganDeciderLocked() != p.Seat {
		return
	}
	keep := "Keep hand"
	if n := e.g.MulliganBottomCountLocked(p); n > 0 {
		keep = fmt.Sprintf("Keep hand (put %d on the bottom)", n)
	}
	e.add(Move{
		Type:   TypeKeepHand,
		Player: e.seat,
		Kind:   KindMulligan,
		Label:  keep,
	})
	if !e.g.MulliganAllowedLocked(p) {
		return
	}
	bottom := e.g.MulliganNextBottomLocked(p)
	label := "Mulligan (draw 7, free)"
	if bottom > 0 {
		label = fmt.Sprintf("Mulligan (draw 7, then put %d on the bottom)", bottom)
	}
	e.add(Move{
		Type:   TypeMulligan,
		Player: e.seat,
		Kind:   KindMulligan,
		Label:  label,
		Params: mustJSON(mulliganParams{HandSize: game.OpeningHandSize - bottom}),
	})
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
	// ADR 0122 §6.2: a decision with no pending-choice ID, named for
	// the cut report and the Choice filter by CleanupDiscardChoice.
	if !e.enter(scope{choice: CleanupDiscardChoice}) {
		return true
	}
	pool := make([]uuid.UUID, 0, len(e.p.Hand.Cards))
	for _, c := range e.p.Hand.Cards {
		pool = append(pool, c.InstanceID)
	}
	for _, set := range e.combos(pool, n, n, e.opts.MaxExpansionPerSource, CapPerSource) {
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
