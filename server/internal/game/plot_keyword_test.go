package game

import (
	"errors"
	"testing"
)

// plot_keyword_test.go — the plot KEYWORD (CR 702.170a, #1342): the
// CR 116.2 special action that exiles a card with plot from its
// owner's hand. The plotted card's cast window is pinned in
// plot_test.go (#1318); what is tested here is the special action's
// own window and what it leaves behind.

const plotOracle = "test-plot-oracle"

// withPlotCard declares "Plot {cost}" on plotOracle.
func withPlotCard(t *testing.T, cost string) {
	t.Helper()
	withCatalogSpecialActions(t, func(id string) []SpecialAction {
		if id != plotOracle {
			return nil
		}
		return []SpecialAction{{Kind: SpecialActionPlot, Cost: cost, Label: "Plot " + cost}}
	})
}

// TestPlotTimingIsYourMainPhaseWithAnEmptyStack is CR 702.170a's
// window: "any time you have priority during your main phase while
// the stack is empty". It is the KEYWORD's window, not the card's —
// an instant or a card with flash is plotted at sorcery speed all the
// same, which is where plot parts company with suspend.
func TestPlotTimingIsYourMainPhaseWithAnEmptyStack(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		keywords []string
		step     Step
		seat     int
		stack    bool
		want     bool
	}{
		{"your precombat main, empty stack", "Creature — Djinn", nil, StepPrecombatMain, 0, false, true},
		{"your postcombat main, empty stack", "Sorcery", nil, StepPostcombatMain, 0, false, true},
		{"an opponent's main phase", "Creature — Djinn", nil, StepPrecombatMain, 1, false, false},
		{"your main phase with a spell on the stack", "Creature — Djinn", nil, StepPrecombatMain, 0, true, false},
		{"your combat", "Creature — Djinn", nil, StepDeclareAttackers, 0, false, false},
		{"your upkeep", "Creature — Djinn", nil, StepUpkeep, 0, false, false},
		{"an instant in your upkeep", "Instant", nil, StepUpkeep, 0, false, false},
		{"a card with flash on an opponent's turn", "Creature — Djinn", []string{"flash"}, StepPrecombatMain, 1, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			g.Turn.Step = tc.step
			if tc.stack {
				stackSpellFor(t, g, g.Seats[1], "Something On The Stack")
			}
			p := g.Seats[tc.seat]
			card := seedHandCard(p, "Plotter", plotOracle, tc.typeLine, "{4}{U}")
			card.Keywords = tc.keywords
			if got := g.SpecialActionTimingOKLocked(p.ID, card, SpecialActionPlot); got != tc.want {
				t.Errorf("plot timing: got %v, want %v", got, tc.want)
			}
		})
	}
}

// TestPlotExilesFaceUpAndGrantsTheFreeCastOnALaterTurn: the whole
// keyword end to end. Pay the plot cost in your main phase; the card
// is in exile, face up, and plotted — not castable this turn
// (CR 702.170d's "any turn after"), castable for free on the next.
func TestPlotExilesFaceUpAndGrantsTheFreeCastOnALaterTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withPlotCard(t, "{3}{U}")
	toMainPhase(t, g)
	card := seedHandCard(me, "Plotter", plotOracle, "Sorcery", "{4}{U}{U}")
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"}, ManaToken{Color: "C"}, ManaToken{Color: "U"})

	if err := g.PerformSpecialAction(me.ID, card.InstanceID, SpecialActionPlot, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("plot: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool: %d tokens left, want 0 — the plot cost was not {3}{U}", len(me.ManaPool))
	}
	if handCard(me, card.InstanceID) != nil {
		t.Fatal("the plotted card is still in hand")
	}
	exiled := exiledCardByIDLocked(g, card.InstanceID)
	if exiled == nil {
		t.Fatal("the plotted card is not in exile")
	}
	if exiled.FaceDown {
		t.Error("the plotted card is face down — CR 702.170a exiles it face up")
	}
	// The zone move names the special action as its cause (#1320).
	var cause MoveCauseKind
	for _, ev := range g.Events {
		if ev.Kind == EventZoneMove && ev.CardID == card.InstanceID {
			cause = ev.Cause
		}
	}
	if cause != MoveCauseSpecialAction {
		t.Errorf("zone move cause = %q, want %q", cause, MoveCauseSpecialAction)
	}

	if _, open := plotLive(g, me, card.InstanceID); open {
		t.Fatal("a card plotted this turn is castable this turn — CR 702.170d says a LATER turn")
	}
	if err := g.CastSpell(me.ID, card.InstanceID, CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Fatal("cast the plotted card on the turn it was plotted")
	}

	g.WithWriteLock(func() { g.Turn.Number++ }) // the owner's next turn
	perm, open := plotLive(g, me, card.InstanceID)
	if perm == nil || !open {
		t.Fatalf("plotted card not castable on a later main phase: perm=%v open=%v", perm, open)
	}
	if err := g.CastSpell(me.ID, card.InstanceID, CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast the plotted {4}{U}{U} for nothing: %v", err)
	}
	if !g.Stack.Contains(card.InstanceID) {
		t.Fatal("the plotted spell never reached the stack")
	}
}

// TestPlotOutsideItsWindowIsRefusedBeforePaying: an opponent's turn,
// a spell on the stack and combat all refuse with
// ErrSpecialActionTiming, and the refusal costs nothing — the card
// stays in hand and the mana stays in the pool.
func TestPlotOutsideItsWindowIsRefusedBeforePaying(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(t *testing.T, g *Game) *Player
	}{
		{"an opponent's turn", func(t *testing.T, g *Game) *Player {
			toMainPhase(t, g)
			return g.Seats[1]
		}},
		{"a spell on the stack", func(t *testing.T, g *Game) *Player {
			toMainPhase(t, g)
			stackSpellFor(t, g, g.Seats[1], "Something On The Stack")
			return g.Seats[0]
		}},
		{"combat", func(t *testing.T, g *Game) *Player {
			g.WithWriteLock(func() {
				g.Turn.Step = StepDeclareAttackers
				g.Turn.Phase = PhaseCombat
			})
			return g.Seats[0]
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			withPlotCard(t, "{1}")
			p := tc.setup(t, g)
			card := seedHandCard(p, "Plotter", plotOracle, "Creature — Djinn", "{4}{U}")
			p.ManaPool.AddMana(ManaToken{Color: "C"})

			err := g.PerformSpecialAction(p.ID, card.InstanceID, SpecialActionPlot, SpecialActionParams{Strict: true})
			if !errors.Is(err, ErrSpecialActionTiming) {
				t.Fatalf("PerformSpecialAction: got %v, want ErrSpecialActionTiming", err)
			}
			if len(p.ManaPool) != 1 {
				t.Errorf("mana pool: got %d, want 1 — a refused plot charged for itself", len(p.ManaPool))
			}
			if handCard(p, card.InstanceID) == nil {
				t.Error("the card left the hand on a refused plot")
			}
		})
	}
}

// TestPlotIsBuilt: the keyword's performer exists, so effects.Register
// accepts a card that declares it and the view and enumerator project
// it.
func TestPlotIsBuilt(t *testing.T) {
	if !SpecialActionKindBuilt(SpecialActionPlot) {
		t.Fatal("plot has no performer")
	}
	if specialActionZone(SpecialActionPlot) != ZoneHand {
		t.Errorf("plot acts on a card in %q, want the hand (CR 702.170a)", specialActionZone(SpecialActionPlot))
	}
}
