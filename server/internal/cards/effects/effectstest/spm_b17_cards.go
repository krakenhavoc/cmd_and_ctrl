package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spmB17Top puts a card on top of p's library, keeping the rest.
func SpmB17Top(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}

// pickOptionWords lists the open option prompt's words for `chooser`.
func PickOptionWords(t *testing.T, g *game.Game, chooser uuid.UUID) []string {
	t.Helper()
	c := LatestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatal("no option prompt open")
	}
	return OptionWords(c)
}
