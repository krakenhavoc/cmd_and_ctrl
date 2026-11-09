package game

import (
	"slices"

	"github.com/google/uuid"
)

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
// (OffersASpell on the permission it stamps), and the land branch plays
// a modal DFC's land back face (CR 712.12) when no face is a spell that
// fits. A card that gives the player more than one answer — a Town (land
// // sorcery Adventure) with a land drop left, a modal DFC with a spell
// face that fits and a land face, a pathway (land // land) — is asked
// one combined prompt instead: the spell offer's resolution_cast and
// capped permission, under MayCastKeywordPlayOrFreeOrHand, which also
// takes a land play of each face that can be played now
// (playLandForResolutionChoiceLocked). One answer is chosen, and
// declining puts the card into the hand. Outside an open resolution,
// which no catalog card reaches, the single offers below stand.
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
	// LeaveOnTop is the clause that does not send the card to the hand
	// when it is not played: "reveal the top card of your library. You
	// may play it without paying its mana cost" (Quicksilver Sea, #4870).
	// Declining, a card that does not fit, and a land with no land drop
	// left all leave it where it is, on top of the library. The prompts
	// say so instead of "Put it into your hand".
	LeaveOnTop bool
}

// declineLabel is the words of the decline answer.
func (o FreePlayFromTop) declineLabel() string {
	if o.LeaveOnTop {
		return "Leave it on top"
	}
	return "Put it into your hand"
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
	spellPerm := g.freeCastFromTopPermissionLocked(o)
	// #4697: more than one answer (a land face and a spell face that
	// fits, or two land faces) is one prompt that offers each of them.
	if lands := g.freePlayLandFacesLocked(o, c); len(lands) > 0 &&
		(len(lands) > 1 || spellPerm.OffersASpell(c)) &&
		g.resolutionOpen && g.playerInGameLocked(o.Player) {
		return g.offerFreePlayOrCastLocked(o, id)
	}
	// A land the player can play as the card is, its front face, is the
	// land offer (a plain land; a Town whose Adventure does not fit).
	if c.IsLand() && o.Land && g.CanPlayLandDuringResolutionForEffect(o.Player, id) {
		then := freePlayLandOrHand
		if o.LeaveOnTop {
			then = freePlayLandOrStay
		}
		return g.offerFreePlayLandLocked(o, id, c.Name, then)
	}
	// #4697: a modal DFC whose back face is a land is played as that face
	// (CR 712.12). Offered alone when no face is a spell that fits; with
	// one that fits, the combined offer above asks both.
	if o.Land && !spellPerm.OffersASpell(c) && c.FaceCount() > 1 &&
		g.CanPlayLandFaceDuringResolutionForEffect(o.Player, id, 1) {
		back := c
		back.SetFace(1)
		then := freePlayBackLandOrHand
		if o.LeaveOnTop {
			then = freePlayBackLandOrStay
		}
		return g.offerFreePlayLandLocked(o, id, back.Name, then)
	}
	// #4697: "a spell with mana value N or less" is about the spell, so
	// the card is offered when some face a cast may choose is a spell
	// that fits (CR 601.3e, 712.11c, 715.3a), asked of the permission
	// the offer is about to stamp (CastPermission.OffersASpell, as
	// #4509's free casts ask it): a modal DFC whose back face fits, a
	// creature card whose Adventure fits, a Town's sorcery Adventure when
	// the land is not played. The permission caps the face being cast,
	// so the cast can only be a face that fits.
	if !spellPerm.OffersASpell(c) || !g.playerInGameLocked(o.Player) {
		return g.putLibraryCardIntoHandLocked(o, id)
	}
	if !g.grantFreeCastFromTopLocked(o, id) {
		return g.putLibraryCardIntoHandLocked(o, id)
	}
	g.queueFreeCastFromTopOfferLocked(o, id, MayCastKeywordFreeOrHand, "Cast it free")
	return nil
}

// MayCastKeywordPlayOrFreeOrHand is the resolution_cast keyword of the
// combined offer (#4697): "you may play a land or cast a spell with mana
// value N or less from the top of your library without paying its mana
// cost. If you don't, put that card into your hand", over a card that
// gives the player more than one of those answers. Its cast is
// free_or_hand's, capped on the face cast (CR 601.3e, 712.11c, 715.3a).
// It also takes a land play of each land face that can be played now
// (CR 305.2b, 712.12), sent as a cast_spell naming the prompt and the
// face. An on-disk value, never renamed.
const MayCastKeywordPlayOrFreeOrHand = "play_or_free_or_hand"

// freePlayLandFacesLocked is the faces of c, the top card, that o.Player
// could play as a land now (CR 712.12: a modal DFC is played as one of
// its land faces; a Town's land is its card face). None when the clause
// allows no land. Caller must hold g.mu.
func (g *Game) freePlayLandFacesLocked(o FreePlayFromTop, c Card) []int {
	if !o.Land {
		return nil
	}
	var faces []int
	for _, face := range c.CastableFaces() {
		if g.CanPlayLandFaceDuringResolutionForEffect(o.Player, c.InstanceID, face) {
			faces = append(faces, face)
		}
	}
	return faces
}

// offerFreePlayOrCastLocked is the combined offer: the spell offer's
// capped permission over the top card, and its resolution_cast under
// MayCastKeywordPlayOrFreeOrHand, which also takes the land plays.
//
// Caller must hold g.mu (write), inside an open resolution.
func (g *Game) offerFreePlayOrCastLocked(o FreePlayFromTop, id uuid.UUID) error {
	if !g.grantFreeCastFromTopLocked(o, id) {
		return g.putLibraryCardIntoHandLocked(o, id)
	}
	if g.queueFreeCastFromTopOfferLocked(o, id, MayCastKeywordPlayOrFreeOrHand, "Play or cast it") == uuid.Nil {
		// Not reached (the caller checked): no prompt to hold the land
		// play, so the offer is withdrawn, which sends the card to the
		// hand.
		g.withdrawCardCopyOffersLocked(o.Player, []uuid.UUID{id}, nil)
	}
	return nil
}

// queueFreeCastFromTopOfferLocked raises the resolution_cast over the
// top card under keyword. Answers its id, or uuid.Nil when no resolution
// is open (the permission alone is then the offer).
//
// Caller must hold g.mu (write).
func (g *Game) queueFreeCastFromTopOfferLocked(o FreePlayFromTop, id uuid.UUID, keyword, accept string) uuid.UUID {
	reason := o.SourceName
	if reason == "" {
		reason = "Cast it free"
	}
	return g.queueResolutionCastOfferLocked(resolutionCastOffer{
		Chooser:      o.Player,
		Source:       o.Source,
		Reason:       reason,
		Candidates:   []uuid.UUID{id},
		Keyword:      keyword,
		AcceptLabel:  accept,
		DeclineLabel: o.declineLabel(),
	})
}

// playLandForResolutionChoiceLocked is the combined offer's land play
// (#4697): a cast_spell through a MayCastKeywordPlayOrFreeOrHand prompt
// whose face is a land plays the card as that face during the
// resolution, with that face up and spending the land drop (CR 305.2b,
// 712.12), as the land offer's "Play it" does. It reports handled false
// for anything else (a spell face, another prompt), which
// castForResolutionChoiceLocked casts as before. A face that cannot be
// played now is refused with the prompt left open. Once the land is
// played the offer is spent and the prompt answered; a play the rules
// end up ignoring is a decline, so the card goes into the hand.
//
// Caller must hold g.mu (write).
func (g *Game) playLandForResolutionChoiceLocked(choiceID, caster, cardID uuid.UUID, face int) (bool, error) {
	i := g.choiceIndexLocked(choiceID)
	if i < 0 || g.PendingChoices[i].MayCastKeyword != MayCastKeywordPlayOrFreeOrHand {
		return false, nil
	}
	card, _, ok := g.ResolutionCastCandidateLocked(caster, cardID)
	if !ok || !slices.Contains(card.CastableFaces(), face) {
		return false, nil
	}
	card.SetFace(face)
	if !card.IsLand() {
		return false, nil
	}
	if !g.CanPlayLandFaceDuringResolutionForEffect(caster, cardID, face) {
		return true, ErrLandDropUnavailable
	}
	c := g.PendingChoices[i]
	played, err := g.PlayLandFaceDuringResolutionForEffect(caster, cardID, face)
	if err != nil {
		return true, err
	}
	if !played {
		// "If you don't, put that card into your hand": withdrawing the
		// offer lapses its permission to the hand.
		if j := g.choiceIndexLocked(choiceID); j >= 0 {
			g.dequeueChoiceLocked(j)
		}
		g.finishResolutionCastLocked(c, c.DiscardOptions)
		g.runStateChecksLocked()
		return true, nil
	}
	// The play is made (its entry may still wait on a prompt, the card
	// where it was), so the permission goes without its lapse to the
	// hand, and the prompt is answered.
	g.dropCardPermissionsLocked(caster, cardID)
	g.settleResolutionCastLocked(choiceID, cardID)
	g.runStateChecksLocked()
	return true, nil
}

// dropCardPermissionsLocked removes holder's ScopeCards permissions that
// name cardID, with no lapse. A fresh slice, for the reason
// lapsePassClosedGrantsLocked gives. Caller must hold g.mu (write).
func (g *Game) dropCardPermissionsLocked(holder, cardID uuid.UUID) {
	p := g.playerByIDLocked(holder)
	if p == nil || len(p.CastPermissions) == 0 {
		return
	}
	var kept []CastPermission
	for _, perm := range p.CastPermissions {
		if perm.Scope == ScopeCards && slices.ContainsFunc(perm.Cards, func(r PermissionCardRef) bool { return r.ID == cardID }) {
			continue
		}
		kept = append(kept, perm)
	}
	p.CastPermissions = kept
}

// freePlayLandOrHand is the land offer's answer (#4361): accepting plays
// the land from the top of the library, and declining, or a play the
// rules end up ignoring, puts it into its owner's hand. Everything it
// reads is on the prompt: the player (the chooser), the source and the
// card. An on-disk key: never renamed, never reused.
var freePlayLandOrHand = RegisterMayCastThen("may-cast/free-play-land-or-hand", func(g *Game, r MayCastAnswer) error {
	return g.freePlayLandFaceOrHandLocked(r, -1)
})

// freePlayBackLandOrHand is freePlayLandOrHand for a modal DFC whose
// back face is the land on offer (#4697): accepting plays the card as
// its back face (CR 712.12). An on-disk key: never renamed, never
// reused.
var freePlayBackLandOrHand = RegisterMayCastThen("may-cast/free-play-back-land-or-hand", func(g *Game, r MayCastAnswer) error {
	return g.freePlayLandFaceOrHandLocked(r, 1)
})

// freePlayLandOrStay and freePlayBackLandOrStay are the land answers of a
// LeaveOnTop clause: accepting plays the land, and declining, or a play the
// rules end up ignoring, leaves it on top of the library. An on-disk key:
// never renamed, never reused.
var freePlayLandOrStay = RegisterMayCastThen("may-cast/free-play-land-or-stay", func(g *Game, r MayCastAnswer) error {
	return g.freePlayLandFaceOrStayLocked(r, -1)
})

var freePlayBackLandOrStay = RegisterMayCastThen("may-cast/free-play-back-land-or-stay", func(g *Game, r MayCastAnswer) error {
	return g.freePlayLandFaceOrStayLocked(r, 1)
})

func (g *Game) freePlayLandFaceOrStayLocked(r MayCastAnswer, face int) error {
	if !r.Accepted {
		return nil
	}
	_, err := g.PlayLandFaceDuringResolutionForEffect(r.Chooser, r.Card, face)
	return err
}

// freePlayLandFaceOrHandLocked is both land answers: play the card as
// `face` (below zero: as it is) on an accept, and put it into its
// owner's hand on a decline or a play the rules end up ignoring.
//
// Caller must hold g.mu (write).
func (g *Game) freePlayLandFaceOrHandLocked(r MayCastAnswer, face int) error {
	o := FreePlayFromTop{Player: r.Chooser, Source: r.Source}
	if r.Accepted {
		played, err := g.PlayLandFaceDuringResolutionForEffect(r.Chooser, r.Card, face)
		if err != nil {
			return err
		}
		if played {
			return nil
		}
	}
	return g.putLibraryCardIntoHandLocked(o, r.Card)
}

// offerFreePlayLandLocked queues the land offer's may_cast over the top
// card, named `name` (the face it would be played as), answered by
// `then`.
//
// Caller must hold g.mu (write).
func (g *Game) offerFreePlayLandLocked(o FreePlayFromTop, id uuid.UUID, name string, then MayCastThen) error {
	return g.QueueMayCastKeyedForEffect(MayCastPrompt{
		Chooser:      o.Player,
		Source:       o.Source,
		Card:         id,
		Question:     o.SourceName + " — play " + name + "?",
		Keyword:      MayCastKeywordPlayOrHand,
		AcceptLabel:  "Play it",
		DeclineLabel: o.declineLabel(),
		Private:      !o.Reveal,
	}, then, nil)
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
		LapseOnPass:       o.lapse(),
		Source:            o.Source,
		SourceName:        o.SourceName,
		Label:             o.SourceName + " — cast it without paying its mana cost",
	}
}

// lapse is where the card goes when its permission lapses uncast: the
// hand, or nowhere (it stays on top) for LeaveOnTop.
func (o FreePlayFromTop) lapse() PermissionLapse {
	if o.LeaveOnTop {
		return ""
	}
	return LapseToHand
}

// putLibraryCardIntoHandLocked is "put that card into your hand": a move
// out of the library to its owner's hand, not a draw. A card that is no
// longer in the library is left alone (CR 608.2c).
//
// Caller must hold g.mu (write).
func (g *Game) putLibraryCardIntoHandLocked(o FreePlayFromTop, cardID uuid.UUID) error {
	if o.LeaveOnTop {
		return nil
	}
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
