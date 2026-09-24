package game

import (
	"testing"

	"github.com/google/uuid"
)

// becomes_plotted_test.go — #1382: EventBecomesPlotted, the event
// "when this card becomes plotted" watches. One emitter
// (plotExiledCardLocked), two routes into it — the plot special action
// and an "it becomes plotted" effect — and nothing else.

// plottedEvents returns every EventBecomesPlotted in the log.
func plottedEvents(g *Game) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventBecomesPlotted {
			out = append(out, ev)
		}
	}
	return out
}

// TestBecomesPlottedFromTheSpecialAction: plotting from hand emits the
// event once, naming the card, its owner as the plotter and the card
// itself as the source — after the card has reached exile and before
// the special action's own EventSpecialAction.
func TestBecomesPlottedFromTheSpecialAction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withPlotCard(t, "{1}")
	toMainPhase(t, g)
	card := seedHandCard(me, "Plotter", plotOracle, "Sorcery", "{4}{U}")
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	if err := g.PerformSpecialAction(me.ID, card.InstanceID, SpecialActionPlot, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("plot: %v", err)
	}
	evs := plottedEvents(g)
	if len(evs) != 1 {
		t.Fatalf("EventBecomesPlotted: got %d, want 1", len(evs))
	}
	ev := evs[0]
	if ev.CardID != card.InstanceID || ev.Actor != me.ID || ev.Source != card.InstanceID {
		t.Errorf("event = card %s actor %s source %s; want card %s, actor the owner, source the card itself",
			ev.CardID, ev.Actor, ev.Source, card.InstanceID)
	}
	var moveSeq, specialSeq uint64
	for _, e := range g.Events {
		switch {
		case e.Kind == EventZoneMove && e.CardID == card.InstanceID && e.NewZone == ZoneExile:
			moveSeq = e.Seq
		case e.Kind == EventSpecialAction && e.CardID == card.InstanceID:
			specialSeq = e.Seq
		}
	}
	if moveSeq == 0 || ev.Seq < moveSeq {
		t.Errorf("EventBecomesPlotted (seq %d) came before the move to exile (seq %d)", ev.Seq, moveSeq)
	}
	if specialSeq == 0 || ev.Seq > specialSeq {
		t.Errorf("EventBecomesPlotted (seq %d) came after EventSpecialAction (seq %d)", ev.Seq, specialSeq)
	}
	// Emitted after the grant: a watcher sees a card that IS plotted.
	if perm := g.CastPermissionOnCardByIDForEffect(card.InstanceID); !perm.Granted() || perm.Timing != TimingPlot {
		t.Errorf("plot permission = %+v, want the TimingPlot grant", perm)
	}
}

// TestBecomesPlottedFromAnEffect: "exile target spell. It becomes
// plotted." emits the same event through the same function, with the
// effect's source named.
func TestBecomesPlottedFromAnEffect(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := plotASpell(t, g, me, "Sorcery", "{2}{U}")
	evs := plottedEvents(g)
	if len(evs) != 1 {
		t.Fatalf("EventBecomesPlotted: got %d, want 1", len(evs))
	}
	if evs[0].CardID != id {
		t.Errorf("event card = %s, want %s", evs[0].CardID, id)
	}
	// Nothing is resolving in this fixture, so the plotter falls back
	// to the owner. (The catalog's Aven Interrupter test pins the
	// resolving controller as the plotter.)
	if evs[0].Actor != me.ID {
		t.Errorf("event actor = %s, want the owner %s", evs[0].Actor, me.ID)
	}

	source := uuid.New()
	other := NewCard("Another", me.ID)
	g.Exile.PushTop(other)
	g.WithWriteLock(func() { g.PlotExiledCardForEffect(other.InstanceID, source) })
	evs = plottedEvents(g)
	if len(evs) != 2 || evs[1].Source != source {
		t.Fatalf("the effect's source is not the event's Source: %+v", evs)
	}
}

// TestPlainExileIsNotBecomingPlotted: exiling a card is not plotting
// it, and "it becomes plotted" off a move that went elsewhere (a card
// that is not in exile) emits nothing.
func TestPlainExileIsNotBecomingPlotted(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)

	spell := NewCard("Exiled Spell", me.ID)
	spell.TypeLine = "Sorcery"
	spell.Controller = me.ID
	pushStackSpell(t, g, spell)
	g.WithWriteLock(func() {
		if err := g.ExileSpellThenForEffect(spell.InstanceID, nil); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	if !g.Exile.Contains(spell.InstanceID) {
		t.Fatal("the spell did not reach exile")
	}

	inHand := seedHandCard(me, "Still In Hand", plotOracle, "Sorcery", "{1}")
	g.WithWriteLock(func() { g.PlotExiledCardForEffect(inHand.InstanceID, uuid.Nil) })

	if evs := plottedEvents(g); len(evs) != 0 {
		t.Fatalf("EventBecomesPlotted without a plot: %+v", evs)
	}
}
