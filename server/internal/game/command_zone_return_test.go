package game

import (
	"testing"

	"github.com/google/uuid"
)

// cmdrB19CommanderOnBattlefield puts a commander from its owner's command
// zone onto the battlefield, the way an effect does.
func cmdrB19CommanderOnBattlefield(t *testing.T, g *Game, p *Player) uuid.UUID {
	t.Helper()
	cmd := pushCommandZoneCommander(p, "Yuriko")
	var err error
	g.WithWriteLock(func() {
		_, err = g.PutFromCommandZoneOntoBattlefieldForEffect(cmd, ZoneEntryOptions{Controller: p.ID})
	})
	if err != nil {
		t.Fatalf("PutFromCommandZoneOntoBattlefieldForEffect: %v", err)
	}
	return cmd
}

// A commander on the battlefield returns to the command zone through the
// routed exit, keeping its identity and asking its owner nothing.
func TestReturnToCommandZoneForEffectSendsACommanderHome(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	id := cmdrB19CommanderOnBattlefield(t, g, p)
	var err error
	g.WithWriteLock(func() { err = g.ReturnToCommandZoneForEffect(id) })
	if err != nil {
		t.Fatalf("ReturnToCommandZoneForEffect: %v", err)
	}
	if g.Battlefield.Contains(id) {
		t.Fatal("the commander is still on the battlefield")
	}
	if !p.Command.Contains(id) {
		t.Fatal("the commander is not in its owner's command zone")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("nothing should be asked: %+v", g.PendingChoices)
	}
}

// A card that is not on the battlefield is left where it is.
func TestReturnToCommandZoneForEffectIgnoresACardOffTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	id := cmdrB19CommanderOnBattlefield(t, g, p)
	for i := 0; i < 2; i++ {
		var err error
		g.WithWriteLock(func() { err = g.ReturnToCommandZoneForEffect(id) })
		if err != nil {
			t.Fatal(err)
		}
	}
	if !p.Command.Contains(id) {
		t.Error("the commander should still be in the command zone")
	}
}
