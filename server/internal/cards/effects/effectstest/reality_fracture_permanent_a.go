package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func RfPermAExpectLegal(t *testing.T, g *game.Game, caster uuid.UUID, spec *game.TargetSpec, want map[uuid.UUID]bool) {
	t.Helper()
	if spec == nil {
		t.Fatal("no target clause")
	}
	var cards []uuid.UUID
	g.ReadSnapshot(func() {
		cards = g.LegalTargetsForEffect(game.SourceChooser(caster), spec).Cards
	})
	in := map[uuid.UUID]bool{}
	for _, id := range cards {
		in[id] = true
	}
	for id, legal := range want {
		if in[id] != legal {
			t.Errorf("target %v legal = %v, want %v", id, in[id], legal)
		}
	}
}
