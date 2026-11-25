package game

import (
	"github.com/google/uuid"
)

// face_up_reveal.go — #6249: a morph cost that is not mana.
//
// "Morph—Reveal a green card in your hand" (Temur Charger and its four
// Khans of Tarkir siblings). CR 702.37e: to turn the permanent face up,
// show all players what its morph cost would be, PAY THAT COST, then
// turn it face up. Here the cost is revealing one card of the named
// colour from the controller's hand (CR 701.20a); the card stays where
// it is (CR 701.20b).
//
// The shape is the turn_face_up offer's {X} (#5249), end to end:
//
//   - the declaration carries it as data (FaceDownCast.FaceUpReveal, a
//     colour letter), and the offer copies it (SpecialAction.RevealColor)
//     the way it copies megamorph's counter;
//   - the offer is made only while the controller's hand holds a card
//     that can pay it (SpecialActionsOfferedLocked through
//     faceUpRevealPayableLocked), so the enumerator, the view and the
//     engine agree that a hand with no green card cannot turn Temur
//     Charger up;
//   - the payload names the card (SpecialActionParams.RevealID, wire
//     reveal_id), refused for any other offer and for a card that is not
//     a matching card in the actor's hand;
//   - the view lists the cards that may be named (FaceUpRevealOptions
//     ForEffect) for the client's picker, and the enumerator names the
//     first.

// turnFaceUpRevealLabel is the menu row and the log line for a reveal
// cost: "Turn face up — reveal a green card".
func turnFaceUpRevealLabel(color string) string {
	return "Turn face up — reveal a " + ColorName(color) + " card"
}

// FaceUpRevealOptionsForEffect is the cards in `actor`'s hand that can
// pay this offer's reveal cost, in hand order. Nil when the offer has no
// reveal cost, or when the actor has no hand.
//
// Caller must hold g.mu (read or write). Read-only.
func (g *Game) FaceUpRevealOptionsForEffect(actor uuid.UUID, sa SpecialAction) []uuid.UUID {
	if sa.RevealColor == "" {
		return nil
	}
	p := g.playerByIDLocked(actor)
	if p == nil || p.Hand == nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.HasColor(sa.RevealColor) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// faceUpRevealPayableLocked reports whether `actor` could pay this
// offer's reveal cost right now. True for every offer with none.
func (g *Game) faceUpRevealPayableLocked(actor uuid.UUID, sa SpecialAction) bool {
	return sa.RevealColor == "" || len(g.FaceUpRevealOptionsForEffect(actor, sa)) > 0
}

// checkFaceUpRevealLocked validates the card named to pay the offer's
// reveal cost: one is required when the offer has a reveal cost, and it
// must be one of FaceUpRevealOptionsForEffect; one is refused otherwise.
func (g *Game) checkFaceUpRevealLocked(p *Player, sa SpecialAction, revealID uuid.UUID) error {
	if sa.RevealColor == "" {
		if revealID != uuid.Nil {
			return ErrInvalidParam
		}
		return nil
	}
	if revealID == uuid.Nil {
		return ErrInvalidParam
	}
	for _, id := range g.FaceUpRevealOptionsForEffect(p.ID, sa) {
		if id == revealID {
			return nil
		}
	}
	return ErrInvalidParam
}

// payFaceUpRevealLocked shows the table the named card (CR 701.20a).
// A no-op for an offer with no reveal cost. Validated beforehand by
// checkFaceUpRevealLocked.
func (g *Game) payFaceUpRevealLocked(p *Player, sourceID uuid.UUID, sa SpecialAction, revealID uuid.UUID) {
	if sa.RevealColor == "" || revealID == uuid.Nil {
		return
	}
	g.RevealForEffect(RevealSpec{
		Player: p.ID,
		Source: sourceID,
		Reason: "morph — reveal a " + ColorName(sa.RevealColor) + " card from hand",
		Cards:  []uuid.UUID{revealID},
	})
}
