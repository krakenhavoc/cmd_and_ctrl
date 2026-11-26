package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b259_helpers.go — shared pieces of Kaladesh batch 259 (#5539).

// S82OldB259DiscardUnlessArtifactRef is "then, if you control no artifacts,
// discard N cards" (Tezzeret's Ambition), the body a draw hands on once its
// last card is drawn. An on-disk key: never renamed, never reused.
var S82OldB259DiscardUnlessArtifactRef = game.RegisterDrawThen("s82-old-b259/discard-if-no-artifacts", func(g *game.Game, d game.DrawThen) error {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == d.Player && c.IsArtifact() {
			return nil
		}
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: d.Player, Source: d.Source, N: d.N})
	return nil
})
