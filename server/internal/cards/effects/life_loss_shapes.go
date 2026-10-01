package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_loss_shapes.go — printed life-loss sentences more than one card
// shares word for word (ADR 0107 §5's "can't gain life" batch, #1880,
// is where most of them arrived).
//
// Append-only.

// eachOpponentLosesTheLifeTheyLostThisTurn is "each opponent loses life
// equal to the life that player lost this turn" (Wound Reflection,
// Archfiend of Despair). The amounts are read off the per-turn tally
// (b18LifeLostThisTurn) for every opponent BEFORE any is applied, so
// one opponent's loss to this ability never inflates another's.
func eachOpponentLosesTheLifeTheyLostThisTurn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opponents := ctx.Opponents()
	amounts := make([]int, len(opponents))
	for i, opp := range opponents {
		amounts[i] = b18LifeLostThisTurn(g, opp)
	}
	for i, opp := range opponents {
		if amounts[i] <= 0 {
			continue
		}
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, opp, -amounts[i]); err != nil {
			return err
		}
	}
	return nil
}

// playerLosesHalfTheirLife is "<that player> loses half their life,
// rounded up" (Havoc Festival, Grievous Wound): the ceiling of half the
// life total as it is at resolution, so 21 loses 11. A player at or
// below zero loses nothing more.
func playerLosesHalfTheirLife(g *game.Game, item *game.StackItem, player uuid.UUID) error {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Eliminated || p.Life <= 0 {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, player, -((p.Life + 1) / 2))
}
