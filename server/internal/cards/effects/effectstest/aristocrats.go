package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aristocrats_test.go — S21 sub-PR 3: the payoffs, and the sprint's
// exit criteria end to end.

const (
	BloodArtistOracle       = "310f141c-7f37-4729-aed6-dd9c09db448d"
	ZulaportCutthroatOracle = "76b003e0-15af-4f22-bdf2-1ade5430964a"
	MayhemDevilOracle       = "4709f11c-aef8-45ac-b2bf-e640c568dfac"
	MidnightReaperOracle    = "e8c7566d-7cc0-48af-a986-83223ec7e06c"
)

// pickPlayer answers the latest pick_target prompt with a player.
func PickPlayer(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	p := LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: target}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}
