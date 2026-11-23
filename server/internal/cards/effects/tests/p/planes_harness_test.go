package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planeswalkToPlane makes the active player planeswalk (CR 901.11) with
// `oracle` on top of the planar deck, so it is the plane they planeswalk to
// and its "when you planeswalk to" ability triggers.
func planeswalkToPlane(t *testing.T, g *game.Game, oracle string) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat].ID
	g.WithWriteLock(func() {
		if up, ok := effectstest.FaceUpPlane(g, oracle); ok {
			// Walk away first, so it is in the deck to be walked to.
			if err := g.PlaneswalkForEffect(me); err != nil {
				t.Fatalf("PlaneswalkForEffect: %v", err)
			}
			_ = up
		}
		c, err := g.PlanarDeck.Remove(effectstest.PlaneID(g, oracle))
		if err != nil {
			t.Fatalf("plane %s is not in the planar deck: %v", oracle, err)
		}
		g.PlanarDeck.PushTop(c)
		if err := g.PlaneswalkForEffect(me); err != nil {
			t.Fatalf("PlaneswalkForEffect: %v", err)
		}
	})
	if _, ok := effectstest.FaceUpPlane(g, oracle); !ok {
		t.Fatalf("plane %s is not face up after the planeswalk", oracle)
	}
}

// rollChaos finishes a planar die roll by `roller` showing the chaos face
// (CR 901.9b), as the dice door reports one: an EventRollDie with no number
// and the face.
func rollChaos(t *testing.T, g *game.Game, roller uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventRollDie, Actor: roller, Sides: 6, PlanarFace: game.PlanarFaceChaos})
	})
	// The special action returns to the active player's priority (CR
	// 901.9b), which is when the chaos ability goes on the stack.
	g.RunStateChecksForTest()
}

// rollBlank is a blank face (CR 901.9a): nothing happens.
func rollBlank(t *testing.T, g *game.Game, roller uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventRollDie, Actor: roller, Sides: 6, PlanarFace: game.PlanarFaceBlank})
	})
}

// requirePlaneRegisteredFull is the registration check every plane test
// file starts with: registered under its oracle ID, CompletenessFull, no
// caveats.
func requirePlaneRegisteredFull(t *testing.T, planes ...effectstest.PlaneFixture) {
	t.Helper()
	for _, p := range planes {
		spec, ok := effects.Lookup(p.Oracle)
		if !ok {
			t.Errorf("%s is not registered", p.Name)
			continue
		}
		if spec.Name != p.Name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", p.Name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}
