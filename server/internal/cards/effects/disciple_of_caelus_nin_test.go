package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const discipleOfCaelusNinOracle = "2e5440ef-6887-4c79-98f5-49db82406d71"

func TestDiscipleOfCaelusNinIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(discipleOfCaelusNinOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registered %v, spec %+v", ok, spec)
	}
}

func phasedOutIDs(g *game.Game) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	g.ReadSnapshot(func() {
		for _, c := range g.PhasedOut.Cards {
			out[c.InstanceID] = true
		}
	})
	return out
}

// Each player, starting with the controller, chooses up to five of their
// own permanents; everything else but the Disciple phases out, and then
// nothing phases in while the Disciple is on the battlefield.
func TestDiscipleOfCaelusNinPhasesOutTheUnchosen(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	var mine, theirs []uuid.UUID
	for i := 0; i < 7; i++ {
		mine = append(mine, b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2))
	}
	for i := 0; i < 2; i++ {
		theirs = append(theirs, b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2))
	}
	disciple := castHoldCreatureNoResolve(t, g, "Disciple of Caelus Nin", "Creature — Human Wizard", discipleOfCaelusNinOracle, 3, 4)
	passPriorityAroundTable(t, g)

	asked := 0
	for i := 0; i < 8; i++ {
		var c *game.PendingChoice
		g.ReadSnapshot(func() {
			for _, pc := range g.PendingChoices {
				if pc != nil && pc.Kind == game.PendingChoiceOwnPermanents {
					c = pc
					return
				}
			}
		})
		if c == nil {
			break
		}
		var pick []uuid.UUID
		switch c.Chooser {
		case me.ID:
			pick = mine[:5]
		case opp.ID:
			pick = theirs[:1]
		}
		if err := g.ResolveOwnPermanents(c.ID, c.Chooser, pick); err != nil {
			t.Fatalf("answering: %v", err)
		}
		asked++
		passPriorityAroundTable(t, g)
	}
	if asked < 2 {
		t.Fatalf("only %d players were asked", asked)
	}
	out := phasedOutIDs(g)
	for _, id := range append(append([]uuid.UUID(nil), mine[:5]...), theirs[0], disciple) {
		if out[id] || !g.Battlefield.Contains(id) {
			t.Errorf("a chosen permanent (or the Disciple) phased out")
		}
	}
	for _, id := range []uuid.UUID{mine[5], mine[6], theirs[1]} {
		if !out[id] || g.Battlefield.Contains(id) {
			t.Errorf("an unchosen permanent did not phase out")
		}
	}
	// Permanents can't phase in: the controller's untap step passes and
	// they stay out.
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+len(g.Seats)-0)%len(g.Seats))
	for range len(g.Seats) {
		advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	}
	out = phasedOutIDs(g)
	if !out[mine[5]] || !out[theirs[1]] {
		t.Error("a permanent phased in while the Disciple was on the battlefield")
	}
	// Once it is gone they phase in at the next untap step.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(disciple); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	for range len(g.Seats) {
		advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	}
	out = phasedOutIDs(g)
	if out[mine[5]] || out[theirs[1]] {
		t.Error("the permanents did not phase in after the Disciple left")
	}
}
