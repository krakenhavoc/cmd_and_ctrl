package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2S43ChooseMode answers chooser's open mode_pick prompt with one mode.
func Cmdr2S43ChooseMode(t *testing.T, g *game.Game, chooser uuid.UUID, mode int) {
	t.Helper()
	c := ModePickChoiceFor(g, chooser)
	if c == nil {
		t.Fatal("no mode prompt is open")
	}
	if err := g.ResolveModePick(c.ID, chooser, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}
