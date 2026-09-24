package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// explosive_derailment_test.go — Spree proof card #2 (CR 702.172a,
// ADR 0065's 2026-09-23 amendment): two bullets, each independently
// targeted at the same price, proving a mode's cost is not tied to
// which bullet it is.

const explosiveDerailmentOracle = "b23dc81d-01bb-4bf0-9932-5d32a6b22cf7"

func TestExplosiveDerailmentDamagesACreatureAlone(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castModal(t, g, "Explosive Derailment", "Instant", explosiveDerailmentOracle,
		[]int{0}, []game.TargetRef{modeRef(game.TargetCard, bear, 0, 0)})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("4 damage should have killed the 2/2")
	}
}

func TestExplosiveDerailmentDestroysAnArtifactAlone(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)

	castModal(t, g, "Explosive Derailment", "Instant", explosiveDerailmentOracle,
		[]int{1}, []game.TargetRef{modeRef(game.TargetCard, rock, 0, 0)})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact bullet should have destroyed the target")
	}
}

// Both bullets, two different targets, each answering its own clause
// — the artifact is never offered to the damage bullet or vice versa.
func TestExplosiveDerailmentBothBulletsHitDifferentTargets(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)

	castModal(t, g, "Explosive Derailment", "Instant", explosiveDerailmentOracle,
		[]int{0, 1},
		[]game.TargetRef{
			modeRef(game.TargetCard, bear, 0, 0),
			modeRef(game.TargetCard, rock, 1, 0),
		})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("the damage bullet should have killed the bear")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the destroy bullet should have destroyed the rock")
	}
}
