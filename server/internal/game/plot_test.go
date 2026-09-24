package game

import (
	"testing"

	"github.com/google/uuid"
)

// plot_test.go — CR 702.170d, the effect half of plot (#1318): "exile
// target spell. It becomes plotted."

// plotASpell puts a spell of `owner` on the stack, exiles it through
// ExileSpellThenForEffect and plots it from the continuation — the
// Aven Interrupter shape. Returns the card's ID.
func plotASpell(t *testing.T, g *Game, owner *Player, typeLine, manaCost string) uuid.UUID {
	t.Helper()
	c := NewCard("Plotted "+typeLine, owner.ID)
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	c.Controller = owner.ID
	pushStackSpell(t, g, c)
	g.WithWriteLock(func() {
		err := g.ExileSpellThenForEffect(c.InstanceID, func(g *Game, exiled bool) error {
			if exiled {
				g.PlotExiledCardForEffect(c.InstanceID, uuid.Nil)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	if !g.Exile.Contains(c.InstanceID) {
		t.Fatal("the spell did not reach exile")
	}
	if _, ok := g.StackMeta[c.InstanceID]; ok {
		t.Fatal("the stack record outlived the exile")
	}
	return c.InstanceID
}

func plotLive(g *Game, p *Player, id uuid.UUID) (*CastPermission, bool) {
	var perm *CastPermission
	var open bool
	g.WithWriteLock(func() {
		c := exiledCardByIDLocked(g, id)
		if c == nil {
			return
		}
		perm = g.CastPermissionForLocked(p.ID, *c, ZoneExile)
		open = perm != nil && g.CastTimingOpenLocked(p.ID, *c, ZoneExile, perm)
	})
	return perm, open
}

// TestPlottedOnItsOwnersTurnWaitsForALaterTurn: "any turn AFTER the
// turn in which it became plotted" — not now, and then for free.
func TestPlottedOnItsOwnersTurnWaitsForALaterTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := plotASpell(t, g, me, "Sorcery", "{4}{U}{U}")

	if perm, _ := plotLive(g, me, id); perm != nil {
		t.Fatal("a card plotted on its owner's turn is castable that same turn")
	}
	g.WithWriteLock(func() { g.Turn.Seq++ }) // the owner's next turn
	perm, open := plotLive(g, me, id)
	if perm == nil || !open {
		t.Fatalf("plot not castable on a later main phase: perm=%v open=%v", perm, open)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast a plotted {4}{U}{U} for nothing: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Fatal("the plotted spell never reached the stack")
	}
}

// TestPlottedOnAnotherTurnOpensOnTheOwnersNextTurn: Turn.Seq changes at
// the boundary, so the owner's turn later in this round IS a later turn.
func TestPlottedOnAnotherTurnOpensOnTheOwnersNextTurn(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[1]
	toMainPhase(t, g)
	id := plotASpell(t, g, opp, "Sorcery", "{2}{W}")

	if _, open := plotLive(g, opp, id); open {
		t.Fatal("a plotted card is castable on somebody else's turn")
	}
	g.WithWriteLock(func() {
		g.Turn.Seq++
		g.Turn.ActiveSeat = 1
		g.Turn.PriorityHolder = 1
	})
	if perm, open := plotLive(g, opp, id); perm == nil || !open {
		t.Fatalf("plot closed on its owner's later turn in the same round: perm=%v open=%v", perm, open)
	}
}

// TestPlottedInstantIsCastOnlyInTheMainPhase: the plot window is the
// permission's, so a plotted INSTANT gains nothing from being one — not
// in the upkeep, and not with a spell on the stack.
func TestPlottedInstantIsCastOnlyInTheMainPhase(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := plotASpell(t, g, me, "Instant", "{U}")
	g.WithWriteLock(func() {
		g.Turn.Seq++
		g.Turn.Step = StepUpkeep
		g.Turn.Phase = PhaseBeginning
	})
	if _, open := plotLive(g, me, id); open {
		t.Error("a plotted instant is castable in the upkeep — CR 702.170d is main phase only")
	}
	g.WithWriteLock(func() {
		g.Turn.Step = StepPrecombatMain
		g.Turn.Phase = PhasePrecombatMain
	})
	if _, open := plotLive(g, me, id); !open {
		t.Error("a plotted instant is not castable in its owner's main phase")
	}
	stackSpellFor(t, g, g.Seats[1], "Something On The Stack")
	if _, open := plotLive(g, me, id); open {
		t.Error("a plotted instant is castable with a spell on the stack")
	}
}

// TestAFlashGrantDoesNotWidenThePlotWindow is what TimingPlot exists
// for, and what TimingSorcery would get wrong: Vedalken Orrery's "you
// may cast spells as though they had flash" overrides a permission's
// sorcery timing, but not the plot rule — it is the permission's own
// window, not the card's speed.
func TestAFlashGrantDoesNotWidenThePlotWindow(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := plotASpell(t, g, me, "Sorcery", "{3}")
	withCatalogCastTimings(t, "test-orrery-1318", CastTimingRule{Timing: TimingFlash, Label: "as though they had flash"})
	timingSource(g, me, "Orrery", "test-orrery-1318")
	g.WithWriteLock(func() {
		g.Turn.Seq++
		g.Turn.Step = StepUpkeep
		g.Turn.Phase = PhaseBeginning
	})
	if _, open := plotLive(g, me, id); open {
		t.Error("a flash grant opened a plotted card outside its main phase — CR 702.170d")
	}
}
