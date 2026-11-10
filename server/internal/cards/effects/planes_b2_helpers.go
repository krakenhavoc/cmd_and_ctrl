package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b2_helpers.go holds the shapes the second and third batches of plane
// cards (ADR 0216 slice 4, #4870) share. Append-only.

// planesB2LoseLife is "you lose N life" for the ability's controller (the
// planar controller on a plane, CR 901.6).
type planesB2LoseLife struct {
	Player uuid.UUID
	Amount int
}

func (l planesB2LoseLife) Apply(ctx *Context) error {
	if l.Amount <= 0 {
		return nil
	}
	player := l.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), player, -l.Amount)
}

// planesB2Alien is the 2/2 white Alien creature token Doctor Who's planes make.
func planesB2Alien() game.Card { return TokenCard("2/2 white Alien") }
