package game

import "github.com/google/uuid"

// face_up_cost_less.go — "This cost is reduced by {1} for each …" on a
// disguise or morph cost (Fugitive Codebreaker, #4452).
//
// It is part of the keyword's own cost, so it is applied to the printed
// turn-face-up cost before the cost-modifier pass: a Thalia-style tax lands
// on the reduced cost, not the other way round. Generic mana only, down to
// zero (CR 118.7), and nothing when the card prints no such clause.

// faceUpCostLessLocked is base reduced by the face-down card's own
// FaceUpCostLess, or base unchanged.
//
// Caller must hold g.mu.
func (g *Game) faceUpCostLessLocked(base ParsedCost, actor uuid.UUID, card Card) ParsedCost {
	if !card.FaceDownIsPermanent() {
		return base
	}
	alt := FaceDownCastFor(faceUpCatalogKey(card))
	if alt == nil || alt.FaceDown == nil || alt.FaceDown.FaceUpCostLess == nil || alt.FaceDown.FaceUpCostLess.Count == nil {
		return base
	}
	n := alt.FaceDown.FaceUpCostLess.Count(g, actor)
	if n <= 0 {
		return base
	}
	if n > base.Generic {
		n = base.Generic
	}
	base.Generic -= n
	return base
}
