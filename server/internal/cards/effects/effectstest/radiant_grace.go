package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func RtaBear(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return PushTribalCreature(g, owner, name, "Creature — Bear", 2, 2)
}

// rtaWaitForTarget passes priority until chooser is asked for a target.
func RtaWaitForTarget(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 6 && LatestPickTarget(g, chooser) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	p := LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no target prompt for the dies trigger: %+v", g.PendingChoices)
	}
	return p
}

func RtaDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(id) })
	if err != nil {
		t.Fatalf("DestroyPermanentForEffect: %v", err)
	}
}
