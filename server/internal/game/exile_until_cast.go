package game

import "github.com/google/uuid"

// exile_until_cast.go — #3213: the "exile cards from the top of your
// library until you exile a card that <matches>. You may cast it
// without paying its mana cost. Put the exiled cards on the bottom of
// your library in a random order" shape, with the hit test supplied by
// the caller. Cascade (a mana-value test) and discover (a bounded one)
// are the keyword forms; Tibalt's Trickery is the first card whose test
// is a name. ADR 0099 amendment 2026-10-13.
//
// It is cascade's body on the same pieces: the exile walk, a {0}
// flash-timed grant over the hit, the resolution_cast prompt that holds
// the resolution open while the chooser casts it (CR 608.2g; ADR 0066's
// 2026-10-23 amendment, #4250), and the random-order bottom. The prompt
// carries the cards the walk passed over (PendingChoice.CastPile) and
// puts them on the bottom when it is answered, with the hit when it was
// declined and the shape bottoms it. Outside an open resolution the
// grant alone is the window, closing on the chooser's next pass. It differs from cascade
// only in having no mana-value cap and in letting the CALLER say whose
// library is walked and who chooses, which for Tibalt's Trickery is the
// countered spell's controller and not the caster.

// MayCastKeywordExileUntil is the PendingChoice.MayCastKeyword the
// shape asks under. It is not a keyword ability, so the client words it
// as "exile until" rather than borrowing cascade's rule number.
const MayCastKeywordExileUntil = "exile_until"

// ExileUntilCastLabel names the grant on the card.
const ExileUntilCastLabel = "Cast it without paying its mana cost"

// ExileUntilCastFreeForEffect walks `player`'s library to the first
// card `hit` accepts, exiling each card on the way, then offers `player`
// the hit's free cast during this resolution. Casting it bottoms the
// rest of the pile in a random order; declining bottoms the hit with the
// pile. A walk that finds nothing bottoms everything. Every "nothing
// happened" case returns nil (CR 608.2c).
//
// `hit` is called during the walk only; it is never stored, so it may
// close over values read before the call (a name, a count) but not a
// *Card. Caller must hold g.mu.
func (g *Game) ExileUntilCastFreeForEffect(player, source uuid.UUID, hit func(Card) bool) error {
	return g.exileUntilCastFreeLocked(player, source, hit, false)
}

// ExileUntilCastFreeKeepingForEffect is ExileUntilCastFreeForEffect for
// the cards that read "each player may cast the nonland card they
// exiled" and then leave an uncast card where it is (#4205, Guff
// Rewrites History): the rest of the pile still goes to the bottom in a
// random order, but a hit the player declines, or doesn't cast before
// passing priority (LapseStaysInExile), stays in exile. Cascade does not
// use it; its hit goes to the bottom (CR 702.85a). A walk that finds
// nothing bottoms everything, as the base form does.
//
// Caller must hold g.mu.
func (g *Game) ExileUntilCastFreeKeepingForEffect(player, source uuid.UUID, hit func(Card) bool) error {
	return g.exileUntilCastFreeLocked(player, source, hit, true)
}

func (g *Game) exileUntilCastFreeLocked(player, source uuid.UUID, hit func(Card) bool, keepUncast bool) error {
	pile, found, err := g.exileUntilLocked(player, hit)
	if err != nil {
		return err
	}
	if found == uuid.Nil {
		return g.PutOnBottomInRandomOrderForEffect(player, ZoneExile, pile)
	}
	declineLabel, lapse := "Put it on the bottom", LapseToLibraryBottom
	if keepUncast {
		declineLabel, lapse = "Leave it in exile", LapseStaysInExile
	}
	uncast := func() error {
		if keepUncast {
			return g.PutOnBottomInRandomOrderForEffect(player, ZoneExile, pile)
		}
		return g.PutOnBottomInRandomOrderForEffect(player, ZoneExile, append(pile, found))
	}
	if !g.playerInGameLocked(player) {
		return uncast()
	}
	g.grantExileUntilCastLocked(player, source, found, lapse)
	if _, _, ok := g.ResolutionCastCandidateLocked(player, found); !ok {
		return uncast()
	}
	reason := "Cast it free"
	if c, _ := g.findCardAndZoneLocked(source); c != nil && c.Name != "" {
		reason = c.Name
	}
	if g.queueResolutionCastOfferLocked(resolutionCastOffer{
		Chooser:      player,
		Source:       source,
		Reason:       reason,
		Candidates:   []uuid.UUID{found},
		Keyword:      MayCastKeywordExileUntil,
		AcceptLabel:  "Cast it free",
		DeclineLabel: declineLabel,
		Pile:         pile,
	}) != uuid.Nil {
		return nil
	}
	// Outside a resolution: the permission is the window, and the
	// cards passed over go to the bottom now.
	return g.PutOnBottomInRandomOrderForEffect(player, ZoneExile, pile)
}

// grantExileUntilCastLocked is grantFreeCastLocked without the mana
// value cap: {0} (so X is 0), flash timing (CR 608.2g), cast-only, and a
// window that the resolution_cast prompt's answer closes (or, outside a
// resolution, the chooser's next pass), with an uncast card going where
// `lapse` says: the bottom of the library for cascade's shape,
// staying in exile for the keeping one.
//
// Caller must hold g.mu.
func (g *Game) grantExileUntilCastLocked(player, source, cardID uuid.UUID, lapse PermissionLapse) {
	c, ok := g.cardInZoneLocked(g.Exile, cardID)
	if !ok {
		return
	}
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:      player,
		Zone:        ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      TimingFlash,
		CastOnly:    true,
		LapseOnPass: lapse,
		Source:      source,
		Label:       ExileUntilCastLabel,
	}, []Card{c})
}
