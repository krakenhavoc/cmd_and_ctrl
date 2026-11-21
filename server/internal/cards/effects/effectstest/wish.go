package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nameWished takes the outside_play action and names `name`.
func NameWished(t *testing.T, g *game.Game, me *game.Player, name string) {
	t.Helper()
	if err := g.NameOutsideCardToPlay(me.ID); err != nil {
		t.Fatalf("outside_play: %v", err)
	}
	c := LatestCollectionPickFor(g, me.ID)
	if c == nil {
		t.Fatalf("no collection pick: %+v", g.PendingChoices)
	}
	if c.ChooseDestination != game.ChoosePlayFromOutside || c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("destination %q bounds %d..%d, want a may-name-one for a play", c.ChooseDestination, c.ChooseMin, c.ChooseMax)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{name}); err != nil {
		t.Fatalf("name %s: %v", name, err)
	}
}
