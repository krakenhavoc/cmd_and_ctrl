package game

import "github.com/google/uuid"

// cast_from_exile_own_text.go — a card whose OWN printed text says "You
// may cast this card from exile" (Khan, Vengeful Heart). The 2026-11-05
// amendment to ADR 0066.
//
// #659 refused ZoneExile in Spec.CastableZones because every exile cast
// it had seen was granted by ANOTHER effect to one object: impulse
// exile, suspend, foretell, cascade, warp. A card-level declaration
// would have opened every copy of those cards in exile, however each
// got there, which their text never says. A card that prints the
// permission on itself is the one case where that is exactly the text:
// the ability is the card's, so it is there whenever the card is face
// up in exile, by whatever route it arrived (CR 400.7 makes it a new
// object; its text comes with it).
//
// Three limits, each from the rules:
//
//   - FACE UP only. A card exiled face down has no characteristics
//     (CR 406.3a, 708.2), so it has no text and grants itself nothing.
//     A face-down exile is castable only under a permission the effect
//     that exiled it granted (foretell, Gonti), as before.
//   - Its OWNER only. "You" on a card that has no controller is its
//     owner (CR 109.5); exile is a shared zone, so the owner is asked
//     here rather than by the pile, as a graveyard's would be.
//   - Through the card's catalog declaration, asked of every face
//     (CardCastableFromAnyFaceLocked), so the derived permission and the
//     card's own zone opening are one answer.
//
// It is DERIVED, never stored, like the sticker permission (#4307): the
// card's presence face up in exile is the duration, so nothing reaches
// Player.CastPermissions, the snapshot or Clone. It is handed out by
// castPermissionLocked, so CastSpell's exile gate, the view's
// `exile_play` stamp and the bot enumerator all read the same answer
// (ADR 0066 Decision 3), and AnyCastPermissionsForEffect's fast positive
// sees it.

// ownTextExilePermissionLocked is the permission a card's own "You may
// cast this card from exile" gives its owner, for the card as it sits
// face up in exile, or nil.
//
// Caller must hold g.mu.
func (g *Game) ownTextExilePermissionLocked(playerID uuid.UUID, card Card) *CastPermission {
	if card.Owner != playerID || !g.cardTextOpensExileLocked(card) {
		return nil
	}
	return &CastPermission{
		Player:     playerID,
		Zone:       ZoneExile,
		Scope:      ScopeCards,
		Cards:      []PermissionCardRef{{ID: card.InstanceID, Epoch: card.ObjectEpoch}},
		Duration:   IndefiniteDuration(),
		CastOnly:   true,
		Source:     card.InstanceID,
		SourceName: card.Name,
		Label:      card.Name + " — you may cast this card from exile",
	}
}

// cardTextOpensExileLocked reports whether the card's own text opens
// exile for it: face up, not a CR 722.3c prepare copy (which only the
// prepared permanent opens, ADR 0090), and declared by some face it may
// be cast as.
//
// Caller must hold g.mu.
func (g *Game) cardTextOpensExileLocked(card Card) bool {
	if card.FaceDown || card.PrepareCopy {
		return false
	}
	return g.CardCastableFromAnyFaceLocked(card, ZoneExile)
}

// anyOwnTextExileCastLocked is AnyCastPermissionsForEffect's fast
// positive for the derived permission above: exile holds a face-up card
// whose own text opens it.
//
// Caller must hold g.mu.
func (g *Game) anyOwnTextExileCastLocked() bool {
	if g.Exile == nil || CatalogCastableZones == nil {
		return false
	}
	for i := range g.Exile.Cards {
		if g.cardTextOpensExileLocked(g.Exile.Cards[i]) {
			return true
		}
	}
	return false
}
