package game

import "github.com/google/uuid"

// cast_free_or_hand.go — #3245: "you may cast the exiled card without
// paying its mana cost if that spell's mana value is N or less. Put the
// exiled card into your hand if it wasn't cast this way" (Beseech the
// Mirror, bargained).
//
// Discover's two branches on a card already in exile, with the caller's
// cap and no EventDiscover: a {0} flash-timed cast-only grant capped at
// the mana value, and a resolution_cast prompt over the card that holds
// the resolution open while the chooser casts it (CR 608.2g: the cast
// happens as part of the resolution; ADR 0066's 2026-10-23 amendment,
// #4250), and a hand for the card the chooser declines (LapseToHand).
// Outside an open resolution the grant alone is the window, closing on
// the chooser's next pass the same way. A card that can never be cast
// this way (a land, or a mana value over the cap) goes straight to its
// owner's hand with no question.
//
// #4697: "that spell's mana value" is the spell's, so which cards can be
// cast this way is asked face by face (CR 601.3e, 712.11c, 715.3a) of
// the permission the offer is about to stamp (CastPermission.OffersASpell,
// as #4509's free casts ask it): a modal DFC whose back face fits the cap,
// a creature card whose Adventure fits it, and a Town's sorcery Adventure
// are all offered. The permission caps the face being cast, so the cast
// itself can only be the face that fits.

// MayCastKeywordFreeOrHand is the PendingChoice.MayCastKeyword the offer
// asks under. The client words it like discover and opens the free cast
// on accept (mayCast.ts, freeCastRequest.ts).
const MayCastKeywordFreeOrHand = "free_or_hand"

// FreeCastOrHandLabel names the grant on the card.
const FreeCastOrHandLabel = "Cast it without paying its mana cost"

// CastExiledFreeOrToHandForEffect offers `player` a free cast of the
// exiled card `cardID` when its mana value is `maxMV` or less, and puts
// it into its owner's hand otherwise, on a decline, or when the grant
// lapses. A card that is no longer in exile is left alone (CR 608.2c).
//
// Caller must hold g.mu.
func (g *Game) CastExiledFreeOrToHandForEffect(player, source, cardID uuid.UUID, maxMV int) error {
	c, ok := g.cardInZoneLocked(g.Exile, cardID)
	if !ok {
		return nil
	}
	if perm := g.freeCastOrHandPermissionLocked(player, source, maxMV); !perm.OffersASpell(c) || !g.playerInGameLocked(player) {
		return g.putExiledIntoHandLocked(player, source, cardID)
	}
	g.grantFreeCastOrHandLocked(player, source, cardID, maxMV)
	reason := "Cast it free"
	if src, _ := g.findCardAndZoneLocked(source); src != nil && src.Name != "" {
		reason = src.Name
	}
	g.queueResolutionCastOfferLocked(resolutionCastOffer{
		Chooser:      player,
		Source:       source,
		Reason:       reason,
		Candidates:   []uuid.UUID{cardID},
		Keyword:      MayCastKeywordFreeOrHand,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put it into your hand",
	})
	return nil
}

// grantFreeCastOrHandLocked is discover's grant without the discover
// record: {0}, flash timing, cast-only, capped at maxMV, lapsing to the
// hand when the prompt is declined (or, outside a resolution, on the
// chooser's next pass).
//
// Caller must hold g.mu.
func (g *Game) grantFreeCastOrHandLocked(player, source, cardID uuid.UUID, maxMV int) {
	c, ok := g.cardInZoneLocked(g.Exile, cardID)
	if !ok {
		return
	}
	perm := g.freeCastOrHandPermissionLocked(player, source, maxMV)
	g.GrantCastPermissionToCardsForEffect(*perm, []Card{c})
}

// freeCastOrHandPermissionLocked is the permission the offer stamps,
// before it names its card: CastExiledFreeOrToHandForEffect asks it which
// cards it could make a spell of (OffersASpell), so the offer and the
// cast it opens judge the same faces by the same cap.
//
// Caller must hold g.mu.
func (g *Game) freeCastOrHandPermissionLocked(player, source uuid.UUID, maxMV int) *CastPermission {
	limit := maxMV
	return &CastPermission{
		Player:            player,
		Zone:              ZoneExile,
		Duration:          g.UntilEndOfTurnDuration(),
		Cost:              "{0}",
		Timing:            TimingFlash,
		CastOnly:          true,
		MaxSpellManaValue: &limit,
		LapseOnPass:       LapseToHand,
		Source:            source,
		Label:             FreeCastOrHandLabel,
	}
}

// putExiledIntoHandLocked moves an exiled card to its owner's hand
// through the one exit primitive. Caller must hold g.mu.
func (g *Game) putExiledIntoHandLocked(player, source, cardID uuid.UUID) error {
	c, ok := g.cardInZoneLocked(g.Exile, cardID)
	if !ok {
		return nil
	}
	_, err := g.routeCardToZoneLocked(zoneRoute{
		CardID:   cardID,
		Dst:      ZoneHand,
		DstOwner: c.Owner,
		Actor:    player,
		Source:   source,
	})
	return err
}
