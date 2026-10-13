package game

import (
	"errors"

	"github.com/google/uuid"
)

// ErrNotInHand is returned by RevealHandCard when the named card is
// not sitting in a player's hand. The verb is hand-only (ADR 0028's
// 2026-10-13 amendment): a card anywhere else is already public or
// already hidden by rules this verb has no business overriding.
var ErrNotInHand = errors.New("game: that card is not in a hand")

// ErrCannotShowTo is returned when the "show to <player>" target is
// not a seat that can be shown anything: not seated, eliminated, or
// the card's own owner (who already knows it).
var ErrCannotShowTo = errors.New("game: that player cannot be shown this card")

// handRevealReason is the one-line label the table-wide reveal banner
// carries for the sandbox verb.
const handRevealReason = "revealed a card from their hand"

// RevealHandCard is the sandbox verb behind the hand context menu's
// "Reveal to table" and "Show to <player>" rows (#671).
//
// With to == uuid.Nil it is a CR 701.20 reveal to the whole table:
// RevealForEffect marks every seat a knower and emits the grouped
// EventRevealCards run, so the reveal frame and banner fire. Eliminated
// seats become knowers too, as they do for any reveal.
//
// With a player in to it is a private show, which the rules have no
// action for (CR 701.20a reveals to ALL players): the card gains that
// one seat as a knower and nothing is announced on the wire except a
// log line that names the target and not the card. It uses
// EventRevealCards with Target set and RevealSeq zero: the reveal
// frame keys on RevealSeq and skips it, and the log already words that
// shape as "P1 revealed a card from their hand to P2" with no identity.
// An eliminated target is refused, so is the owner.
//
// Knowledge lasts as long as KnownBy does: the card leaving the hand
// into the library or a face-down exile replaces it, and a public zone
// makes everyone a knower anyway. Nothing moves.
func (g *Game) RevealHandCard(cardID, to uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	zone := g.findCardZoneLocked(cardID)
	if zone == nil {
		return ErrCardNotFound
	}
	if zone.Kind != ZoneHand {
		return ErrNotInHand
	}
	owner := zone.Owner
	if to == uuid.Nil {
		g.RevealForEffect(RevealSpec{
			Player: owner,
			Reason: handRevealReason,
			Cards:  []uuid.UUID{cardID},
		})
		return nil
	}
	p := g.playerByIDLocked(to)
	if p == nil || p.Eliminated || to == owner {
		return ErrCannotShowTo
	}
	for i := range zone.Cards {
		if zone.Cards[i].InstanceID != cardID {
			continue
		}
		zone.Cards[i].AddKnower(to)
		break
	}
	g.EmitEvent(Event{
		Kind:    EventRevealCards,
		Actor:   owner,
		Target:  to,
		CardID:  cardID,
		Label:   handRevealReason,
		OldZone: ZoneHand,
	})
	return nil
}
