package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const cliffstomperOracle = "20f3bc44-c7fa-4d5a-adb9-2dd1d2d66ffe"

func pushMountains(g *game.Game, controller uuid.UUID, n int) {
	for range n {
		effectstest.PushCatalogPermanent(g, controller, "Mountain", "Basic Land — Mountain", "", false)
	}
}

func TestSardianCliffstomperIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(cliffstomperOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registered %v, spec %+v", ok, spec)
	}
}

// 0/4 with fewer than four Mountains; +X/+0 (X = Mountains) with four or
// more, on its controller's turn only.
func TestSardianCliffstomperScalesWithMountainsOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	src := effectstest.PushCatalogPermanent(g, me.ID, "Sardian Cliffstomper", "Creature — Minotaur Barbarian", cliffstomperOracle, true)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, src)
		c.Power, c.Toughness = 0, 4
	})
	pushMountains(g, me.ID, 3)
	if p, tt := effectstest.PtOf(t, g, src); p != 0 || tt != 4 {
		t.Fatalf("with three Mountains %d/%d, want 0/4", p, tt)
	}
	pushMountains(g, me.ID, 1)
	if p, tt := effectstest.PtOf(t, g, src); p != 4 || tt != 4 {
		t.Fatalf("with four Mountains %d/%d, want 4/4", p, tt)
	}
	pushMountains(g, me.ID, 2)
	if p, tt := effectstest.PtOf(t, g, src); p != 6 || tt != 4 {
		t.Fatalf("with six Mountains %d/%d, want 6/4", p, tt)
	}
	// An opponent's Mountains are not yours.
	pushMountains(g, g.Seats[1].ID, 3)
	if p, _ := effectstest.PtOf(t, g, src); p != 6 {
		t.Fatalf("opponent Mountains counted: power %d, want 6", p)
	}
	// Not your turn: no bonus.
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p, tt := effectstest.PtOf(t, g, src); p != 0 || tt != 4 {
		t.Fatalf("on an opponent's turn %d/%d, want 0/4", p, tt)
	}
}
