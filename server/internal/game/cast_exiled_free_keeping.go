package game

import "github.com/google/uuid"

// cast_exiled_free_keeping.go — "exile it … then you may cast it without
// paying its mana cost" for a card already in exile, with a decline that
// leaves it there (Transcendent Dragon: "counter target spell. If that
// spell is countered this way, exile it instead of putting it into its
// owner's graveyard, then you may cast it without paying its mana cost").
//
// It is ExileUntilCastFreeKeepingForEffect's last half on its own: a
// {0}, flash-timed, cast-only grant over the one card, lapsing to
// "stays in exile" (LapseStaysInExile), and a resolution_cast prompt that
// holds the resolution open while the chooser casts it (CR 608.2g, ADR
// 0066's 2026-10-23 amendment). The prompt is the exile-until family's
// (MayCastKeywordExileUntil), so the client needs nothing new.

// CastExiledFreeKeepingForEffect offers `player` a free cast of the exiled
// card `cardID`. A card that is no longer in exile, or that cannot be cast
// (a land), is left alone (CR 608.2c). Declining, or passing priority
// without casting, leaves the card in exile.
//
// Caller must hold g.mu.
func (g *Game) CastExiledFreeKeepingForEffect(player, source, cardID uuid.UUID) error {
	if _, ok := g.cardInZoneLocked(g.Exile, cardID); !ok || !g.playerInGameLocked(player) {
		return nil
	}
	g.grantExileUntilCastLocked(player, source, cardID, LapseStaysInExile)
	if _, _, ok := g.ResolutionCastCandidateLocked(player, cardID); !ok {
		return nil
	}
	reason := "Cast it free"
	if c, _ := g.findCardAndZoneLocked(source); c != nil && c.Name != "" {
		reason = c.Name
	}
	g.queueResolutionCastOfferLocked(resolutionCastOffer{
		Chooser:      player,
		Source:       source,
		Reason:       reason,
		Candidates:   []uuid.UUID{cardID},
		Keyword:      MayCastKeywordExileUntil,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it in exile",
	})
	return nil
}
