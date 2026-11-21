package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmtB3MakeToken creates one token for owner from the token table.
func TmtB3MakeToken(g *game.Game, owner uuid.UUID, key string) {
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(owner, effects.TokenCard(key), 1) })
}
