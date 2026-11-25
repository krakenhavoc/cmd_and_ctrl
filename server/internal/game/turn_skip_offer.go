package game

import (
	"slices"

	"github.com/google/uuid"
)

// turn_skip_offer.go — "If you would begin your turn while this artifact
// is tapped, you may skip that turn instead. If you do, untap this
// artifact" (Time Vault, #6226; CR 614.10, 614.10a, 500.11).
//
// An optional replacement on a turn beginning. The rotation seam
// (chooseNextTurnLocked) is where a turn is chosen, and since ADR 0188
// Decision 12 it is a place that can pause: the old turn has closed,
// Game.TurnChangePending is set, nothing has begun. A TurnSkipOffer is
// plain data on the card's definition; when the turn about to begin —
// normal or extra — is its controller's and the offer is live, the turn
// change pauses on a keyed confirm (confirm/turn-skip-offer) addressed
// to that player.
//
//   - Yes: the player takes on one "skip your next turn" charge (the ADR
//     0188 debt, ModSkipTurns) from the source, which the turn about to
//     begin spends at once through the existing walk, and the source is
//     untapped when the offer says so ("If you do, untap this
//     artifact"). A skipped turn never begins (CR 500.11).
//   - No: the turn change resumes with that source declined, so the next
//     live offer (a second Time Vault) is asked, and the turn begins once
//     none is left.
//
// The confirm carries the source and the sources already declined for
// this turn, so the table is a restore point. An older binary refuses
// the file (unknown continuation key).

// TurnSkipOffer declares "If you would begin your turn [while this
// permanent is tapped], you may skip that turn instead. [If you do,
// untap this permanent.]" Plain data on CardDef.TurnSkipOffers.
type TurnSkipOffer struct {
	// WhileTapped is "while this <permanent> is tapped".
	WhileTapped bool
	// UntapSource is "If you do, untap this <permanent>".
	UntapSource bool
	// Label names the clause.
	Label string
}

// CatalogTurnSkipOffers returns the offers declared by an ability key.
var CatalogTurnSkipOffers func(oracleID string) []TurnSkipOffer

var turnSkipOfferThen ConfirmThen

func init() {
	turnSkipOfferThen = RegisterConfirmThen("confirm/turn-skip-offer", turnSkipOfferAnswered)
}

// liveTurnSkipOfferLocked is the first permanent `player` controls whose
// offer is live and not in `declined`, and its offer.
//
// Caller holds g.mu.
func (g *Game) liveTurnSkipOfferLocked(player uuid.UUID, declined []uuid.UUID) (uuid.UUID, TurnSkipOffer, bool) {
	if g.Battlefield == nil || CatalogTurnSkipOffers == nil {
		return uuid.Nil, TurnSkipOffer{}, false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != player || slices.Contains(declined, c.InstanceID) {
			continue
		}
		key := catalogAbilityKeyOf(c)
		if key == "" {
			continue
		}
		for _, o := range CatalogTurnSkipOffers(key) {
			if o.WhileTapped && !c.Tapped {
				continue
			}
			return c.InstanceID, o, true
		}
	}
	return uuid.Nil, TurnSkipOffer{}, false
}

// offerTurnSkipLocked pauses the turn change to ask `seat`'s player a
// live offer, reporting whether it did. The caller has put the
// candidate turn back where the resume will find it.
//
// Caller must hold g.mu (write).
func (g *Game) offerTurnSkipLocked(seat int, declined []uuid.UUID) bool {
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil || g.Seats[seat].Eliminated {
		return false
	}
	player := g.Seats[seat].ID
	src, offer, ok := g.liveTurnSkipOfferLocked(player, declined)
	if !ok {
		return false
	}
	label := offer.Label
	if label == "" {
		if c := findBattlefieldCard(g, src); c != nil {
			label = c.Name
		}
	}
	carry := append([]uuid.UUID{src}, declined...)
	id, err := g.QueueConfirmKeyedForEffect(ConfirmPrompt{
		Chooser:      player,
		Source:       src,
		Question:     "You would begin your turn. Skip that turn instead? (" + label + ")",
		AcceptLabel:  "Skip my turn",
		DeclineLabel: "Take my turn",
	}, turnSkipOfferThen, carry)
	if err != nil || id == uuid.Nil {
		return false
	}
	g.TurnChangePending = true
	g.Turn.PriorityHolder = NoPriority
	return true
}

func turnSkipOfferAnswered(g *Game, r ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	src := r.Carry[0]
	declined := append([]uuid.UUID(nil), r.Carry[1:]...)
	if r.Accepted {
		declined = nil
		if c := findBattlefieldCard(g, src); c != nil && c.Controller == r.Chooser {
			var offer TurnSkipOffer
			if key := catalogAbilityKeyOf(c); key != "" && CatalogTurnSkipOffers != nil {
				if offers := CatalogTurnSkipOffers(key); len(offers) > 0 {
					offer = offers[0]
				}
			}
			label := offer.Label
			if label == "" {
				label = c.Name
			}
			g.SkipNextTurnsForEffect(r.Chooser, src, 1, label)
			if offer.UntapSource {
				g.untapPermanentByIDLocked(src)
			}
		}
	} else {
		declined = append(declined, src)
	}
	if !g.TurnChangePending {
		return nil
	}
	g.TurnChangePending = false
	if g.chooseNextTurnExceptLocked(declined) {
		return nil
	}
	g.runStepEntryHooksLocked()
	g.runStateChecksLocked()
	return nil
}
