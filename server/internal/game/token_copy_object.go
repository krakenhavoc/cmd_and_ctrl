package game

import "github.com/google/uuid"

// TokenCopyTemplateForObjectForEffect copies one battlefield object's
// copiable values (CR 707.2), using last-known information once that
// object has departed (CR 608.2h). A card that has returned with a new
// object epoch is not the object the resolving trigger named.
//
// Caller must hold g.mu in write mode.
func (g *Game) TokenCopyTemplateForObjectForEffect(ref ObjectRef) (Card, bool) {
	info, ok := g.PermanentForEffect(ref)
	if !ok || info.Copiable == nil {
		return Card{}, false
	}
	values := info.Copiable.Clone()
	values.MakeToken()
	var token Card
	token.setPrintedValues(values)
	return token, true
}

// TokenCopySourceForEffect is the card `cardID` as a token copy of it
// sees it (CR 707.2): a value copy whose printed fields are the values
// the copy takes. It is the read behind effects.TokenCopyTemplate, so
// every "create a token that's a copy of …" reads the same thing.
//
//   - On the battlefield, the permanent's copiable values: its printed
//     values as other copy effects modified them (a Clone copying a
//     dragon is the dragon), a face-down permanent's CR 708.2 body, a
//     Room's whole card, and a flip card's unflipped top half, since
//     flipped is a status and is not copied (CR 707.3).
//   - Off the battlefield, when the card has not moved since it left
//     (its object epoch is the one after the departed permanent's),
//     the effect is about the permanent that left, and it reads that
//     permanent's last-known copiable values (CR 608.2h). A Clone that
//     was copying a dragon and died in response is copied as the
//     dragon, not as the Clone card in the graveyard.
//   - Anywhere else, the card as it is in its zone: a card exiled from a
//     graveyard by an embalm cost or The Scarab God, a discarded card,
//     a card that died earlier and has moved on since. Each of those is
//     a new object the effect names on its own (CR 400.7).
//
// The card-carried ability slices stay those of the card that was
// found: they are the token's only source of abilities when the copied
// identity has no catalog key (a copy of a token copying a token). A
// face-down body has no text, so it carries none.
//
// False when no card has that ID. Caller must hold g.mu in write mode.
func (g *Game) TokenCopySourceForEffect(cardID uuid.UUID) (Card, bool) {
	src, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return Card{}, false
	}
	var values *PrintedValues
	if live := findBattlefieldCard(g, cardID); live != nil {
		if info, ok := g.PermanentForEffect(ObjectRef{ID: cardID, Epoch: live.ObjectEpoch}); ok {
			values = info.Copiable
		}
		if live = findBattlefieldCard(g, cardID); live != nil {
			src = *live
		}
	} else if recs := g.lastKnownPermanents[cardID]; len(recs) > 0 {
		if rec := recs[len(recs)-1]; rec.Copiable != nil && src.ObjectEpoch == rec.Epoch+1 {
			values = rec.Copiable
		}
	}
	if values == nil {
		return src, true
	}
	faceDown := src.FaceDownIsPermanent()
	src.ClearFaceDown()
	src.Flipped = false
	src.setPrintedValues(*values)
	if faceDown {
		src.ManaAbilities = nil
		src.ActivatedAbilities = nil
	}
	return src, true
}
