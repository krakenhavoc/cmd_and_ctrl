package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_direction_view_test.go — ADR 0206 (#4211) on the wire. A
// direction rule in force reaches every viewer as one `attack_directions`
// row naming the direction, what said so, and who each player may
// attack, so the table can show it before anyone tries to attack.
func TestAttackDirectionRulesReachTheView(t *testing.T) {
	g := busyTable(t, 0)
	if v := ViewOfGameFor(g, g.Seats[0].ID.String()); len(v.AttackDirections) != 0 {
		t.Fatalf("no rule in force, but the view lists %+v", v.AttackDirections)
	}

	caster := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		g.GrantAttackDirectionForEffect(game.AttackDirectionRight, "Teyo, Geometric Tactician", uuid.Nil, g.UntilYourNextTurnDuration(caster.ID))
	})

	for _, viewer := range []string{g.Seats[0].ID.String(), g.Seats[2].ID.String(), ""} {
		v := ViewOfGameFor(g, viewer)
		if len(v.AttackDirections) != 1 {
			t.Fatalf("viewer %q: %d rows, want 1: %+v", viewer, len(v.AttackDirections), v.AttackDirections)
		}
		row := v.AttackDirections[0]
		if row.Direction != "right" || row.Source != "Teyo, Geometric Tactician" {
			t.Errorf("viewer %q: row = %+v", viewer, row)
		}
		for i, p := range g.Seats {
			if want := g.Seats[(i+3)%4].ID.String(); row.Nearest[p.ID.String()] != want {
				t.Errorf("viewer %q: seat %d may attack %q, want %q", viewer, i, row.Nearest[p.ID.String()], want)
			}
		}
	}

	// A rule that has ended is not listed.
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Statics = nil
		}
	})
	if v := ViewOfGameFor(g, ""); len(v.AttackDirections) != 0 {
		t.Errorf("rule ended, but the view lists %+v", v.AttackDirections)
	}
}
