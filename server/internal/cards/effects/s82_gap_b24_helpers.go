package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82GapB24OpponentDrawsExceptTheFirst is the AppliesTo shared by Notion
// Thief and Hullbreacher: "if an opponent would draw a card except the first
// one they draw in each of their draw steps". The drawer is a player still in
// the game other than the source's controller, and the draw is not the first
// one of their own draw step (a draw expanded by another replacement keeps its
// first card exempt).
func S82GapB24OpponentDrawsExceptTheFirst(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventDraw || ev.DrawPlayer == src.Controller {
		return false
	}
	if g.PlayerByIDForEffect(ev.DrawPlayer) == nil {
		return false
	}
	return !g.FirstDrawInOwnDrawStepForEffect(ev.DrawPlayer)
}

// s82GapB24TreasureInstead is Hullbreacher's draw substitute: the draw is
// cancelled and the source's controller creates a Treasure token, then the
// rest of the draw instruction goes on. `source` is read for its controller
// wherever it now is: a Hullbreacher that left in response still names the
// player who controlled it. A registered body (an on-disk key).
func S82GapB24TreasureInstead(g *game.Game, _, source uuid.UUID, done game.DrawDone) error {
	if c, ok := g.LookupCardForEffect(source); ok && c.Controller != uuid.Nil {
		if err := g.CreateTokenForEffect(c.Controller, c.TokenThroughText(TreasureToken()), 1); err != nil {
			return err
		}
	}
	return done.Run(g)
}
