package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b47_helpers.go — shared bodies for GAP batch 47 (#5948).

// targetOpponentGetsAnEmblem is "Target opponent gets an emblem with ..." as
// a planeswalker ability's effect: the first still-legal player target is
// given the card's emblem (Ob Nixilis Reignited's −8, Garruk, Apex Predator's
// −8). A target that is no longer legal gives nothing (CR 608.2b).
func TargetOpponentGetsAnEmblem(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return CreateEmblem{Player: t.ID}.Apply(ctx)
		}
	}
	return nil
}

// garrukApexPredatorGainLife is the −3's second sentence: the controller gains
// life equal to the creature's toughness (N), as it was before it was
// destroyed (CR 608.2h). It is not conditional on the destruction, and the
// destruction's continuation runs whatever the move did. An on-disk key.
var GarrukApexPredatorGainLife = CardTail("garruk-apex-predator/gain-life", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	return GainLife{Amount: o.Tail.N}.Apply(ctx)
})
