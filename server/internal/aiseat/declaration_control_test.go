package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// declaration_control_test.go — #4154: bots under a "you choose which
// creatures attack / block" effect (Master Warcraft). Bob chooses the
// declarations on Alice's turn; every seat is a bot, played lockstep.

func attackedThisTurn(g *game.Game, card uuid.UUID) bool {
	var hit bool
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventAttack && ev.CardID == card {
				hit = true
			}
		}
	})
	return hit
}

// A bot choosing a bot's declarations makes them legally and the table
// moves on: no attackers but a goaded one (which must attack), whose
// target the active player's bot then accepts; no blocks for the
// defender. Nothing either runner sends is refused, and the declarer's
// own policy never sees a move it makes for another seat.
func TestBotsUnderADeclarationControl(t *testing.T) {
	room := newRoom(t, 3, 4154)
	g := room.Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	a := g.Turn.ActiveSeat
	alice, bob, carol := g.Seats[a], g.Seats[(a+1)%3], g.Seats[(a+2)%3]
	var goaded, bear uuid.UUID
	g.WithWriteLock(func() {
		goaded = limitPush(g, alice, "Goaded Bear", "Creature — Bear", "", 2, 2)
		bear = limitPush(g, alice, "Bear", "Creature — Bear", "", 2, 2)
		limitPush(g, carol, "Wall", "Creature — Wall", "", 0, 4)
		limitPush(g, bob, "Wall", "Creature — Wall", "", 0, 4)
		if !g.BeginDeclarationControlForEffect(game.DeclarationControl{
			Declarer: bob.ID, Source: uuid.New(), SourceName: "Master Warcraft", Attackers: true, Blockers: true,
		}) {
			t.Fatal("BeginDeclarationControlForEffect refused")
		}
	})
	if err := g.SetGoaded(goaded, carol.ID); err != nil {
		t.Fatalf("SetGoaded: %v", err)
	}
	seq := turnSeq(g)
	alicePol, bobPol, carolPol := &watching{}, &watching{}, &watching{}
	ar := aiseat.NewSteppedRunner(room, alice.ID, alicePol, aiseat.Config{}, testLogger())
	br := aiseat.NewSteppedRunner(room, bob.ID, bobPol, aiseat.Config{}, testLogger())
	cr := aiseat.NewSteppedRunner(room, carol.ID, carolPol, aiseat.Config{}, testLogger())

	stepUntilTurnEnds(t, g, seq, ar, br, cr)

	if bobPol.sawMoveFor(alice.ID) || bobPol.sawMoveFor(carol.ID) {
		t.Error("the declarer's own policy was handed a move it makes for another seat")
	}
	for name, r := range map[string]*aiseat.Runner{"Alice": ar, "Bob": br, "Carol": cr} {
		if st := r.Stats(); st.Rejected != 0 {
			t.Errorf("%s's runner: %+v, want nothing rejected", name, st)
		}
	}
	if !attackedThisTurn(g, goaded) {
		t.Error("the goaded creature, which had to attack, did not")
	}
	if attackedThisTurn(g, bear) {
		t.Error("the declarer's bot attacked with a creature nothing required")
	}
}
