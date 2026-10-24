package game

import "github.com/google/uuid"

// free_play_from_top.go — #3801: "look at the top card of your library.
// You may play a land from the top of your library or cast a spell with
// mana value N or less from the top of your library without paying its
// mana cost. If you don't, put that card into your hand" (Ziatora's
// Envoy), and the revealed spell-only form (Rashmi, Eternities Crafter).
//
// ADR 0099's 2026-10-15 amendment. It is cast_free_or_hand.go's offer
// with the card still in the LIBRARY instead of in exile, plus a land
// branch:
//
//   - a nonland card whose mana value fits is cast during the resolution
//     (CR 608.2g; ADR 0066's 2026-10-23 amendment, #4250): a {0},
//     flash-timed, cast-only, mana-value-capped permission over that
//     card in the library, and a resolution_cast prompt over it under
//     MayCastKeywordFreeOrHand, which holds the resolution open while
//     the player casts it from the top of the library or declines.
//     Declining lapses the permission with the card sent to its owner's
//     hand (LapseToHand; lapseCastPermissionLocked reads the zone the
//     grant opens). Outside an open resolution the permission alone is
//     the window, closing on the holder's next pass the same way;
//   - a land the player can play right now (CR 305.2b, 305.3) gets a
//     may_cast prompt under MayCastKeywordPlayOrHand. Accepting plays it
//     through PlayLandDuringResolutionForEffect, which works from any
//     zone, so a library card needs nothing new there. Declining, or a
//     play the rules end up ignoring, puts the card into the hand;
//   - anything else (a land with no land drop left, a spell over the
//     cap, a land when the card does not offer the play) goes straight
//     to the hand with no question: the player could not have played it,
//     so "if you don't" is already true.
//
// The look is private (CR 701.20): only the looker is marked a knower of
// the card. The land's may_cast is marked private so its words, which
// name the card, reach nobody else; the spell's resolution_cast names
// only the source, and its card is dropped from every other viewer's
// copy of the prompt by the knower filter, as any hidden card on a
// prompt is. A reveal (Rashmi) marks every seat and both are public.

// MayCastKeywordPlayOrHand is the PendingChoice.MayCastKeyword a land
// offer asks under: accepting plays the land at once, so the client has
// no cast to open.
const MayCastKeywordPlayOrHand = "play_or_hand"

// FreePlayFromTop is the offer: whose library, what the clause allows,
// and whether the look is a reveal.
type FreePlayFromTop struct {
	// Player is "you": whose library it is, who looks, and who may
	// play or cast the card.
	Player uuid.UUID
	// Source is the object whose ability this is.
	Source uuid.UUID
	// SourceName heads the question and labels the grant.
	SourceName string
	// MaxManaValue caps the spell that may be cast, inclusive. Below
	// zero means no spell fits (Rashmi after a cast of mana value 0).
	MaxManaValue int
	// Land is whether the clause also lets the player play a land
	// from the top of the library.
	Land bool
	// Reveal is "reveal the top card" rather than "look at": every seat
	// learns the card, and the prompt is public.
	Reveal bool
}

// OfferFreePlayFromTopOrHandForEffect looks at (or reveals) the top
// card of o.Player's library and runs the offer above. An empty library
// does nothing.
//
// Caller must hold g.mu (write).
func (g *Game) OfferFreePlayFromTopOrHandForEffect(o FreePlayFromTop) error {
	p := g.playerByIDLocked(o.Player)
	if p == nil || p.Library == nil || p.Library.Size() == 0 {
		return nil
	}
	var ids []uuid.UUID
	if o.Reveal {
		ids = g.RevealTopOfLibraryForEffect(o.Player, o.Source, 1, o.SourceName+" — the top card of the library")
	} else {
		ids = g.LookAtTopOfLibraryForEffect(o.Player, 1)
	}
	if len(ids) == 0 {
		return nil
	}
	id := ids[0]
	c, ok := g.cardInZoneLocked(p.Library, id)
	if !ok {
		return nil
	}
	if c.IsLand() {
		if !o.Land || !g.CanPlayLandDuringResolutionForEffect(o.Player, id) {
			return g.putLibraryCardIntoHandLocked(o, id)
		}
		return g.QueueMayCastPromptForEffect(MayCastPrompt{
			Chooser:      o.Player,
			Source:       o.Source,
			Card:         id,
			Question:     o.SourceName + " — play " + c.Name + "?",
			Keyword:      MayCastKeywordPlayOrHand,
			AcceptLabel:  "Play it",
			DeclineLabel: "Put it into your hand",
			Private:      !o.Reveal,
			OnAccept: func(g *Game) error {
				played, err := g.PlayLandDuringResolutionForEffect(o.Player, id)
				if err != nil {
					return err
				}
				if !played {
					return g.putLibraryCardIntoHandLocked(o, id)
				}
				return nil
			},
			OnDecline: func(g *Game) error {
				return g.putLibraryCardIntoHandLocked(o, id)
			},
		})
	}
	mv, readable := g.ManaValueForEffect(c)
	if !readable || mv > o.MaxManaValue || !g.playerInGameLocked(o.Player) {
		return g.putLibraryCardIntoHandLocked(o, id)
	}
	if !g.grantFreeCastFromTopLocked(o, id) {
		return g.putLibraryCardIntoHandLocked(o, id)
	}
	reason := o.SourceName
	if reason == "" {
		reason = "Cast it free"
	}
	g.queueResolutionCastOfferLocked(resolutionCastOffer{
		Chooser:      o.Player,
		Source:       o.Source,
		Reason:       reason,
		Candidates:   []uuid.UUID{id},
		Keyword:      MayCastKeywordFreeOrHand,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put it into your hand",
	})
	return nil
}

// grantFreeCastFromTopLocked stamps the free cast over the card where
// it sits in the library. Reports false when it has left the library
// since the question.
//
// Caller must hold g.mu (write).
func (g *Game) grantFreeCastFromTopLocked(o FreePlayFromTop, cardID uuid.UUID) bool {
	p := g.playerByIDLocked(o.Player)
	if p == nil {
		return false
	}
	c, ok := g.cardInZoneLocked(p.Library, cardID)
	if !ok {
		return false
	}
	limit := o.MaxManaValue
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:            o.Player,
		Zone:              ZoneLibrary,
		Duration:          g.UntilEndOfTurnDuration(),
		Cost:              "{0}",
		Timing:            TimingFlash,
		CastOnly:          true,
		MaxSpellManaValue: &limit,
		LapseOnPass:       LapseToHand,
		Source:            o.Source,
		SourceName:        o.SourceName,
		Label:             o.SourceName + " — cast it without paying its mana cost",
	}, []Card{c})
	return true
}

// putLibraryCardIntoHandLocked is "put that card into your hand": a move
// out of the library to its owner's hand, not a draw. A card that is no
// longer in the library is left alone (CR 608.2c).
//
// Caller must hold g.mu (write).
func (g *Game) putLibraryCardIntoHandLocked(o FreePlayFromTop, cardID uuid.UUID) error {
	p := g.playerByIDLocked(o.Player)
	if p == nil {
		return nil
	}
	c, ok := g.cardInZoneLocked(p.Library, cardID)
	if !ok {
		return nil
	}
	_, err := g.routeCardToZoneLocked(zoneRoute{
		CardID:   cardID,
		Dst:      ZoneHand,
		DstOwner: c.Owner,
		Actor:    o.Player,
		Source:   o.Source,
	})
	return err
}
