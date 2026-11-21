package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mscM11Token makes a token under `owner`'s control, as an effect would.
func MscM11Token(t *testing.T, g *game.Game, owner uuid.UUID, row string) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{ID: uuid.New(), Controller: owner, SourceCardID: uuid.Nil})
		if err := (effects.CreateToken{Template: effects.TokenCard(row), N: 1, Controller: owner}).Apply(ctx); err != nil {
			t.Fatalf("token %q: %v", row, err)
		}
	})
}
