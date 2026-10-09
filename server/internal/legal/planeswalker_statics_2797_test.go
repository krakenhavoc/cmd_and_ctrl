package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// planeswalker_statics_2797_test.go — #2797 asked of the readers that are
// not the engine: the enumerator a bot picks from, and the wire view the
// client draws. Every move offered here is dispatched against a clone
// (dispatchAll), which is the #544 parity check.

const (
	oracleKioraSaltAndSand = "e70bb7f8-8098-49c5-929c-68f63cd821c5"
	oracleUrSphinx         = "4a3fdb8e-4699-4bd9-84e6-3cc7fea0e1ef"
)

// timeRaveler seats Teferi, Time Raveler with `loyalty` counters under `name`
// (distinct names, so two of them do not meet the legend rule).
func timeRaveler(g *game.Game, p *game.Player, name string, loyalty int) uuid.UUID {
	return battlefieldCard(g, p, game.Card{
		Name: name, TypeLine: "Legendary Planeswalker — Teferi",
		OracleID: oracleTimeRavelerTiming, Counters: map[string]int{"loyalty": loyalty},
	})
}

// kiora seats Kiora of Salt and Sand and announces her arrival, so the layer
// pass (which reads the battlefield's events) sees the grantor.
func kiora(g *game.Game, p *game.Player) uuid.UUID {
	id := battlefieldCard(g, p, game.Card{
		Name: "Kiora of Salt and Sand", TypeLine: "Legendary Creature — Merfolk Noble",
		OracleID: oracleKioraSaltAndSand, Power: 2, Toughness: 4,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	return id
}

func hasActivation(moves []legal.Move, source uuid.UUID, labelPrefix string) bool {
	for _, m := range activationsOf(moves, source) {
		if strings.HasPrefix(m.Label, labelPrefix) || strings.Contains(m.Label, labelPrefix) {
			return true
		}
	}
	return false
}

// A loyalty ability granted to every planeswalker you control is on the
// list for each walker that can pay it (CR 606.6) and only for that walker,
// and it shares the walker's one activation a turn (CR 606.3).
func TestEnumeratorOffersTheClassGrantedLoyaltyAbility(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	rich := timeRaveler(g, active, "Teferi, Time Raveler", 9)
	poor := timeRaveler(g, active, "Teferi, Hidden Twin", 7)
	advanceTo(t, g, game.StepPrecombatMain)

	if hasActivation(legal.EnumerateFor(g, active.ID), rich, "−8") {
		t.Fatal("setup: a −8 is offered with no grantor on the table")
	}
	kiora(g, active)

	moves := legal.EnumerateFor(g, active.ID)
	if !hasActivation(moves, rich, "−8") {
		t.Fatalf("Kiora's granted −8 is not offered on a 9-loyalty walker: %v", labels(moves))
	}
	if hasActivation(moves, poor, "−8") {
		t.Errorf("the granted −8 is offered on a 7-loyalty walker: %v", labels(activationsOf(moves, poor)))
	}
	dispatchAll(t, g, active.ID, activationsOf(moves, rich))

	// Once the walker has used a loyalty ability this turn, nothing on it is
	// offered — printed rows and the granted one alike.
	g.WithWriteLock(func() {
		if g.LoyaltyActivatedThisTurn == nil {
			g.LoyaltyActivatedThisTurn = map[uuid.UUID]bool{}
		}
		g.LoyaltyActivatedThisTurn[rich] = true
	})
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), rich); len(acts) != 0 {
		t.Errorf("the walker already activated this turn and %d loyalty moves are still offered: %v", len(acts), labels(acts))
	}
}

// The wire view draws the granted row on the walker, with the grantor's name
// and the loyalty cost, for the client's ordinary ability menu.
func TestWireViewShowsTheClassGrantedLoyaltyRow(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	walker := timeRaveler(g, active, "Teferi, Time Raveler", 9)
	kiora(g, active)
	advanceTo(t, g, game.StepPrecombatMain)

	v := protocol.ViewOfGameFor(g, active.ID.String())
	var found bool
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID != walker.String() {
			continue
		}
		for _, row := range c.ActivatedAbilities {
			if !strings.Contains(row.Label, "Leviathan") {
				continue
			}
			found = true
			if row.GrantedBy == nil || row.GrantedBy.Name != "Kiora of Salt and Sand" {
				t.Errorf("granted_by = %+v, want Kiora of Salt and Sand", row.GrantedBy)
			}
			if row.LoyaltyCost == nil || *row.LoyaltyCost != -8 {
				t.Errorf("loyalty_cost = %v, want -8", row.LoyaltyCost)
			}
			if row.TimingClosed {
				t.Error("the row is timing_closed in my main phase")
			}
		}
	}
	if !found {
		t.Fatal("the walker's wire view has no row for Kiora's granted −8")
	}
}

// A stored "until end of turn, loyalty abilities at instant speed" statement
// (Jace's Machinations) reaches the list and the wire the same way the
// derived statements do: both read ActivationTimingOpenLocked.
func TestEnumeratorAndViewHonourAStoredInstantSpeedStatement(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	walker := timeRaveler(g, active, "Teferi, Time Raveler", 4)
	advanceTo(t, g, game.StepEnd)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), walker); len(acts) != 0 {
		t.Fatalf("a loyalty ability is offered in the end step with no statement: %v", labels(acts))
	}
	rows := loyaltyRowsClosed(t, g, active.ID, walker)
	if len(rows) == 0 {
		t.Fatal("setup: the walker has no ability rows on the wire")
	}
	for i, closed := range rows {
		if !closed {
			t.Errorf("row %d: timing_closed absent with no statement", i)
		}
	}

	g.WithWriteLock(func() {
		g.GrantLoyaltyActivationTimingForEffect(active.ID, game.TimingFlash, "Teferi", "Jace's Machinations", uuid.Nil, game.Duration{})
	})
	acts := activationsOf(legal.EnumerateFor(g, active.ID), walker)
	if len(acts) == 0 {
		t.Fatal("the stored statement did not put the loyalty abilities on the list")
	}
	dispatchAll(t, g, active.ID, acts)
	for i, closed := range loyaltyRowsClosed(t, g, active.ID, walker) {
		if closed {
			t.Errorf("row %d: timing_closed still set under the statement", i)
		}
	}

	// A walker of another subtype is not covered by a statement about Teferis.
	other := battlefieldCard(g, active, game.Card{
		Name: "Test Chandra", TypeLine: "Legendary Planeswalker — Chandra",
		OracleID: oracleTimeRavelerTiming, Counters: map[string]int{"loyalty": 4},
	})
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), other); len(acts) != 0 {
		t.Errorf("a Chandra's loyalty abilities were offered under a statement about Teferis: %v", labels(acts))
	}
}

func castsFor(moves []legal.Move, src uuid.UUID) int {
	n := 0
	for _, m := range moves {
		if m.Kind == legal.KindCast && m.Source == src {
			n++
		}
	}
	return n
}

// Eminence prices the cast in the enumerator the way the engine does: a
// {3}{U}{U} Sphinx is offered off four lands when The Ur-Sphinx sits in the
// command zone, and not without it.
func TestEnumeratorPricesEminenceFromTheCommandZone(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	sphinx := handCard(active, game.Card{Name: "Test Sphinx", TypeLine: "Creature — Sphinx", ManaCost: "{3}{U}{U}", Power: 3, Toughness: 3})
	for i := 0; i < 4; i++ {
		battlefieldCard(g, active, basic("Island", "Island"))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	if n := castsFor(legal.EnumerateFor(g, active.ID), sphinx); n != 0 {
		t.Fatalf("a five-mana Sphinx is offered off four lands: %d casts", n)
	}
	active.Command.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "The Ur-Sphinx", TypeLine: "Legendary Creature — Sphinx Avatar",
		ManaCost: "{6}{W}{U}{B}", OracleID: oracleUrSphinx, IsCommander: true,
		Owner: active.ID, Controller: active.ID,
	})
	moves := legal.EnumerateFor(g, active.ID)
	if n := castsFor(moves, sphinx); n != 1 {
		t.Fatalf("with The Ur-Sphinx in the command zone the Sphinx has %d casts, want 1: %v", n, labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)

	// The discount is the owner's: an opponent's Sphinx is not helped.
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if n := castsFor(legal.EnumerateFor(g, opp.ID), sphinx); n != 0 {
		t.Errorf("a card in my hand is castable by %s: %d casts", opp.Name, n)
	}
}
