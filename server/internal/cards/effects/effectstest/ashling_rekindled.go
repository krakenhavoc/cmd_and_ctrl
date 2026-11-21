package effectstest

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ashlingPickColor answers the mana trigger's colour pick with the
// option for `color`.
func AshlingPickColor(t *testing.T, g *game.Game, chooser uuid.UUID, color string) {
	t.Helper()
	c := LatestChoiceOfKindFor(g, game.PendingChoiceMana, chooser)
	if c == nil {
		t.Fatal("no colour pick for Ashling, Rimebound's mana")
	}
	for _, opt := range c.ColorOptions {
		if strings.HasPrefix(opt, color) {
			if err := g.ResolveManaChoice(c.ID, chooser, opt); err != nil {
				t.Fatalf("ResolveManaChoice %s: %v", opt, err)
			}
			return
		}
	}
	t.Fatalf("colour pick offers %v, no %s", c.ColorOptions, color)
}
