package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pt_counters_1664_test.go — #1664: every P/T counter kind changes
// power and toughness (CR 122.1a). One test per proof card, one
// refusal for Contagion, and the replacement checks: Hardened Scales
// ("+1/+1 counters") leaves a +1/+2 alone, Doubling Season ("one or
// more counters") doubles it.

const (
	ContagionOracle        = "433e86ce-df04-4182-bb63-ad2607d67dbd"
	WallOfRootsOracle      = "3a21a6ae-b2f2-4f0c-acfd-5f3e8d63fd2f"
	DwarvenArmorerOracle   = "5bbd27b1-0afd-4d98-a73c-c348c8f08625"
	ArmorThrullOracle      = "35453c9e-e1ae-4fe9-926d-75724deb0555"
	LightningSerpentOracle = "d07b4bb6-0c8d-44ea-a5b3-eeb6e36f3631"
)

// ptOf is a battlefield creature's current power and toughness.
func PtOf(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	var p, tt int
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				p, tt, found = c.PowerForComparison(), c.CurrentToughness(), true
			}
		}
	})
	if !found {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return p, tt
}
