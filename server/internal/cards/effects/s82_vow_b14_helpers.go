package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82VowB14YouDiscarded is "you discarded a card this turn".
func S82VowB14YouDiscarded(g *game.Game, player uuid.UUID) bool {
	return g.TurnTallyFor(player).CardsDiscarded >= 1
}

// s82VowB14YouGainedAtLeast is "you gained N or more life this turn".
func S82VowB14YouGainedAtLeast(g *game.Game, player uuid.UUID, n int) bool {
	return B15LifeGainedThisTurn(g, player) >= n
}

// s82VowB14YourEndStepIf is the condition of "At the beginning of your end
// step, if <cond>": your end step began and cond holds (CR 603.4).
func S82VowB14YourEndStepIf(cond func(g *game.Game, source *game.Card) bool) func(game.Event, *game.Card, game.Characteristic, *game.Game) bool {
	return func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return ByYou(ev, source, lki, g) && cond(g, source)
	}
}

// s82VowB14TransformIf is a trigger body that transforms the source only if
// cond still holds as it resolves (CR 603.4).
func S82VowB14TransformIf(cond func(g *game.Game, source *game.Card) bool) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok || !OnBattlefield(g, item.SourceCardID) || !cond(g, &c) {
			return nil
		}
		return TransformThis{}.Apply(NewContext(g, item))
	}
}
