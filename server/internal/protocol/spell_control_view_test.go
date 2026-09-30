package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spell_control_view_test.go — ADR 0104 on the wire: a stolen spell's
// stack item carries its live controller and, while they differ, the
// player it was taken from; and the log says it out loud.

func TestStolenSpellCarriesItsDefaultController(t *testing.T) {
	g := buildActiveGame(t)
	thief, caster := g.Seats[0], g.Seats[1]
	id := pushStackSpell(g, "Lightning Bolt", "bolt", game.PaidCost{}, game.StackItemSpell)

	if v := stackItemView(t, g, id); v.DefaultController != "" {
		t.Fatalf("an untouched spell sends default_controller %q", v.DefaultController)
	}
	g.WithWriteLock(func() {
		if !g.GainControlOfSpellForEffect(uuid.New(), id, thief.ID, "test — steal") {
			t.Fatal("the steal was refused")
		}
	})
	v := stackItemView(t, g, id)
	if v.Controller != thief.ID.String() {
		t.Errorf("controller = %s, want the thief %s", v.Controller, thief.ID)
	}
	if v.DefaultController != caster.ID.String() {
		t.Errorf("default_controller = %q, want the caster %s", v.DefaultController, caster.ID)
	}
}

func TestLogNarratesASpellControlChange(t *testing.T) {
	g := buildActiveGame(t)
	thief := g.Seats[0]
	id := pushStackSpell(g, "Lightning Bolt", "bolt", game.PaidCost{}, game.StackItemSpell)
	g.WithWriteLock(func() {
		g.GainControlOfSpellForEffect(uuid.New(), id, thief.ID, "test — steal")
	})
	entry := findLog(t, ViewOfGame(g).Log, LogControl)
	if want := "P1 gained control of Lightning Bolt from P2"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	assertNoUUID(t, entry.Text)
}
