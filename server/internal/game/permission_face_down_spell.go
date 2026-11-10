package game

// permission_face_down_spell.go — a cast permission judges a spell cast
// face down as the object it is (#4452, Assemble the Players).
//
// A face-down spell is a 2/2 creature spell with no name, mana cost or
// types beyond creature (CR 708.4, CR 708.2a), so "a creature spell with
// power 2 or less" admits a 5-power disguise card cast face down (the card's
// ruling). The permission lookups read the card as it sits in its zone,
// whose printed numbers are the real ones, so the lookups that have a claim
// in hand judge this stand-in instead whenever the claim is the card's own
// face-down offer (morph, megamorph, disguise). Only the copy a lookup
// judges is changed; the card in its zone is not.

// faceDownSpellStandIn is the card as a face-down cast of it would be
// judged: a copy stamped with the face-down state of the offer `claim`
// names, or the card unchanged when the claim is not a face-down offer of
// it (or it is already face down).
func faceDownSpellStandIn(card Card, claim string) Card {
	if claim == "" || card.FaceDown {
		return card
	}
	alt := AlternativeCostByKey(castOfferKey(card), AltCostKeyword(claim))
	if alt == nil || alt.FaceDown == nil {
		return card
	}
	card.SetFaceDown(alt.FaceDown.Kind)
	return card
}
