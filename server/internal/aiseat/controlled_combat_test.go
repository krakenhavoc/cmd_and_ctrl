package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// controlled_combat_test.go — #3999: bots and a combat phase another
// player plays (Secret of Bloodbending). Alice controls Bob during Bob's
// next combat phase; every seat is a bot, played lockstep.

// A bot controlling a bot's combat phase: the controller passes for it
// through the phase with the passive policy, its own policy never sees
// Bob's moves, and nothing either runner sends is refused. Bob decides
// for himself the rest of his turn.
func TestBotControllingABotsCombatPhase(t *testing.T) {
	room := newRoom(t, 3, 3999)
	g := room.Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	a := g.Turn.ActiveSeat
	alice, bob, carol := g.Seats[a].ID, g.Seats[(a+1)%3].ID, g.Seats[(a+2)%3].ID
	g.WithWriteLock(func() {
		if !g.BeginCombatControlForEffect(game.CombatControl{Controller: alice, Controlled: bob, Source: uuid.New(), SourceName: "Secret of Bloodbending"}) {
			t.Fatal("BeginCombatControlForEffect refused")
		}
	})
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	seq := turnSeq(g)
	alicePol, bobPol, carolPol := &watching{}, &watching{}, &watching{}
	ar := aiseat.NewSteppedRunner(room, alice, alicePol, aiseat.Config{}, testLogger())
	br := aiseat.NewSteppedRunner(room, bob, bobPol, aiseat.Config{}, testLogger())
	cr := aiseat.NewSteppedRunner(room, carol, carolPol, aiseat.Config{}, testLogger())

	stepUntilTurnEnds(t, g, seq, ar, br, cr)

	if alicePol.sawMoveFor(bob) {
		t.Error("the controller's own policy was handed a move for Bob")
	}
	if st := ar.Stats(); st.Applied == 0 || st.Rejected != 0 {
		t.Errorf("Alice's runner: %+v, want moves applied for Bob in his combat and none rejected", st)
	}
	if st := br.Stats(); st.Applied == 0 || st.Rejected != 0 {
		t.Errorf("Bob's runner: %+v, want his own moves outside combat and none rejected", st)
	}
	if g.ControllerOfPlayer(bob) != bob {
		t.Error("Bob is still controlled after his turn")
	}
	var left int
	g.ReadSnapshot(func() { left = len(g.PlayerControls) })
	if left != 0 {
		t.Errorf("%d control records left after the combat phase", left)
	}
}
