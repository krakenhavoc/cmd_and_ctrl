package game

import "github.com/google/uuid"

// named_card_copy.go — "Create a copy of the card with the chosen name. You
// may cast the copy." (Garth One-Eye, #6155; CR 707.13, CR 707.12).
//
// CR 707.13: the copy is of a card defined by NAME, not of an object: the
// player "uses the Oracle card reference to determine the characteristics
// of the copy and creates the copy outside of the game". The engine's
// Oracle reference is the printed-record registry conjure already keeps
// (ConjurableTemplate, filed from the catalog's conjurable_cards.go), so the
// copy is built from that record's copiable values and is marked
// Card.CardCopy, "not a card", exactly as a card_copy.go copy is.
//
// "Outside of the game" is the caster's own outside container,
// Player.Sideboard (ZoneOutside, CR 400.11): no ZoneRef reaches it and the
// card index does not cover it, so nothing can move, target or find the
// copy there (CR 400.11c). CR 707.12 then casts it from where it was
// created, while the ability is still resolving: the offer is the
// card_copy.go offer (a ScopeCards permission over the copy, here opening
// ZoneOutside) with the resolution_cast prompt over it, which is how the
// caster casts it before the resolution finishes (CR 608.2g). The cast is
// the ordinary cast from outside the game (cast_outside.go), so it pays
// the copy's own costs, chooses its targets and is a spell that was cast.
//
// A copy nobody casts ceases to exist (CR 707.10a, 704.5e): declining the
// prompt withdraws the offer, the grant lapses leaving the copy where it is
// (LapseStaysOutside), and the card-copy sweep removes a CardCopy no
// permission names from the outside container as it does from exile.

// CopyNamedCardAndOfferCastForEffect creates, outside the game, a copy of
// the card named `name` for offer.Player, who owns it, and offers that
// player its cast on the offer's terms. offer.Cost "" is the copy's own
// mana cost ("You still pay its costs"). It answers the copy's ID, or
// ErrNotConjurable for a name with no printed record and ErrPlayerNotFound
// for a player not in the game.
//
// Caller must hold g.mu in write mode.
func (g *Game) CopyNamedCardAndOfferCastForEffect(offer CardCopyOffer, name string) (uuid.UUID, error) {
	p := g.playerByIDLocked(offer.Player)
	if p == nil || p.Eliminated {
		return uuid.Nil, ErrPlayerNotFound
	}
	template, ok := ConjurableTemplate(name)
	if !ok {
		return uuid.Nil, ErrNotConjurable
	}
	cp := g.newCardCopyLocked(template, p.ID)
	cp.AddKnower(p.ID)
	if p.Sideboard == nil {
		p.Sideboard = newZone(ZoneOutside, p.ID)
	}
	p.Sideboard.PushTop(cp)
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:      p.ID,
		Zone:        ZoneOutside,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        offer.Cost,
		Timing:      offer.Timing,
		CastOnly:    true,
		LapseOnPass: LapseStaysOutside,
		FollowUp:    offer.FollowUp,
		Source:      offer.Source,
		SourceName:  offer.SourceName,
		Label:       offer.SourceName + " — cast a copy of " + cp.Name,
	}, []Card{cp})
	if offer.DuringResolution {
		g.queueResolutionCastLocked(offer, []uuid.UUID{cp.InstanceID})
	}
	return cp.InstanceID, nil
}

// outsideCardCopyCandidateLocked is ResolutionCastCandidateLocked's answer
// for a card in `chooser`'s own outside container: only a copy of a card
// (CR 707.13), which nothing but a resolution's own offer puts there.
//
// Caller must hold g.mu.
func (g *Game) outsideCardCopyCandidateLocked(chooser, id uuid.UUID) (Card, bool) {
	p := g.playerByIDLocked(chooser)
	if p == nil || p.Sideboard == nil {
		return Card{}, false
	}
	for _, c := range p.Sideboard.Cards {
		if c.InstanceID == id && c.CardCopy {
			return c, true
		}
	}
	return Card{}, false
}
