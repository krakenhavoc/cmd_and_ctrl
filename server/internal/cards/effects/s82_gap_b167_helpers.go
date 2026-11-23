package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OpponentLostLifeLastTurn reports whether an opponent of `you` lost life
// during the previous turn, whoever's it was (Feast on the Fallen). Caller
// must hold g.mu.
func OpponentLostLifeLastTurn(g *game.Game, you uuid.UUID) bool {
	for _, p := range g.Seats {
		if p != nil && p.ID != you && !p.Eliminated && g.LostLifeLastTurnForEffect(p.ID) {
			return true
		}
	}
	return false
}

// CondOpponentLostLifeLastTurn is "if an opponent lost life last turn" as an
// intervening-if Condition, read for the item's controller. It declares no
// footprint read: the fact is stamped as the turn begins and no step changes
// it.
type CondOpponentLostLifeLastTurn struct{}

func (CondOpponentLostLifeLastTurn) Holds(ctx *Context) bool {
	return OpponentLostLifeLastTurn(ctx.Game, ctx.Controller())
}

func (CondOpponentLostLifeLastTurn) Condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// OpponentLostLifeLastTurnWhen is the When for the same condition as the
// trigger fires.
func OpponentLostLifeLastTurnWhen(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return OpponentLostLifeLastTurn(g, source.Controller)
}
