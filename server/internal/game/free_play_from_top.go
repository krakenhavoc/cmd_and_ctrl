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
// #4697: the spell branch judges the face being cast, not the front face
// (OffersASpell on the permission it stamps). The land branch still reads
// the front face: a modal DFC whose back face is a land is not offered
// as a land play, and a Town (land // sorcery Adventure) whose land can
// be played is offered as the land play only.
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
	// The land offer is about the card as it is in the library, its
	// front face: a modal DFC whose BACK face is a land is not offered
	// as a land play (#4697 left that half open: playing a back face
	// during a resolution needs a face on the land play and its prompt).
	if c.IsLand() && o.Land && g.CanPlayLandDuringResolutionForEffect(o.Player, id) {
		return g.QueueMayCastKeyedForEffect(MayCastPrompt{
			Chooser:      o.Player,
			Source:       o.Source,
			Card:         id,
			Question:     o.SourceName + " — play " + c.Name + "?",
			Keyword:      MayCastKeywordPlayOrHand,
			AcceptLabel:  "Play it",
			DeclineLabel: "Put it into your hand",
			Private:      !o.Reveal,
		}, freePlayLandOrHand, nil)
	}
	// #4697: "a spell with mana value N or less" is about the spell, so
	// the card is offered when some face a cast may choose is a spell
	// that fits (CR 601.3e, 712.11c, 715.3a), asked of the permission
	// the offer is about to stamp (CastPermission.OffersASpell, as
	// #4509's free casts ask it): a modal DFC whose back face fits, a
	// creature card whose Adventure fits, a Town's sorcery Adventure when
	// the land is not played. The permission caps the face being cast,
	// so the cast can only be a face that fits.
	if !g.freeCastFromTopPermissionLocked(o).OffersASpell(c) || !g.playerInGameLocked(o.Player) {
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

// freePlayLandOrHand is the land offer's answer (#4361): accepting plays
// the land from the top of the library, and declining, or a play the
// rules end up ignoring, puts it into its owner's hand. Everything it
// reads is on the prompt: the player (the chooser), the source and the
// card. An on-disk key: never renamed, never reused.
var freePlayLandOrHand = RegisterMayCastThen("may-cast/free-play-land-or-hand", func(g *Game, r MayCastAnswer) error {
	o := FreePlayFromTop{Player: r.Chooser, Source: r.Source}
	if r.Accepted {
		played, err := g.PlayLandDuringResolutionForEffect(r.Chooser, r.Card)
		if err != nil {
			return err
		}
		if played {
			return nil
		}
	}
	return g.putLibraryCardIntoHandLocked(o, r.Card)
})

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
	g.GrantCastPermissionToCardsForEffect(*g.freeCastFromTopPermissionLocked(o), []Card{c})
	return true
}

// freeCastFromTopPermissionLocked is the permission the spell offer
// stamps, before it names its card: OfferFreePlayFromTopOrHandForEffect
// asks it whether the card is a spell it could cast (OffersASpell), so
// the offer and the cast it opens judge the same faces by the same cap.
//
// Caller must hold g.mu.
func (g *Game) freeCastFromTopPermissionLocked(o FreePlayFromTop) *CastPermission {
	limit := o.MaxManaValue
	return &CastPermission{
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
	}
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
