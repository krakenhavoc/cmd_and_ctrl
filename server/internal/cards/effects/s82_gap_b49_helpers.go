package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB49CountersPlacedOnThis is "whenever one or more +1/+1 counters are put
// on this creature": one trigger per placement event that actually added a
// counter, not one per counter (the shape Evolution Witness uses).
func gapB49CountersPlacedOnThis(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Target == source.InstanceID && b33CountersPlacedDelta(ev, game.CounterPlusOne, g) > 0
}

// gapB49OtherArtifactsYouControl counts the artifacts the activator controls
// other than the ability's own source (Etherium Pteramander).
func gapB49OtherArtifactsYouControl(q game.CostQuery) int {
	return countBattlefield(q, func(c game.Card) bool {
		return c.Controller == q.Controller && c.InstanceID != q.Source.InstanceID &&
			Artifact()(q.Game, q.Controller, c)
	})
}

// gapB49DrawnThreeOrMoreThisTurn is "if you've drawn three or more cards this
// turn" for a cost modifier, read at CR 601.2f.
func gapB49DrawnThreeOrMoreThisTurn(q game.CostQuery) bool {
	return q.Game != nil && len(q.Game.CardsDrawnThisTurnFor(q.Controller)) >= 3
}
