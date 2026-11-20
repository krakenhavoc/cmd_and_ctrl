package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pickTargetRefs answers the newest pick_target owed by `chooser`.
func PickTargetRefs(t *testing.T, g *game.Game, chooser uuid.UUID, dist map[uuid.UUID]int, refs ...game.TargetRef) {
	t.Helper()
	p := LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	var err error
	if dist != nil {
		err = g.ResolvePickTargetsDivided(p.ID, chooser, refs, dist)
	} else {
		err = g.ResolvePickTargets(p.ID, chooser, refs)
	}
	if err != nil {
		t.Fatalf("answer targets: %v", err)
	}
}
