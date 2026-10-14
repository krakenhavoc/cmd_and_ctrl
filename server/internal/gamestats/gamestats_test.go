package gamestats

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// A four-seat game, built by hand so every definition in ADR 0164 §3 has
// a line that tests it. Seat 0 (Ana) wins; seat 3 (Dee) concedes on
// turn 4; seat 1 (Bo) dies on turn 6; seat 2 (Cy) is still standing.

const (
	ana, bo, cy, dee = "p-ana", "p-bo", "p-cy", "p-dee"
)

func card(id, name, owner string, opts ...func(*protocol.CardView)) protocol.CardView {
	c := protocol.CardView{InstanceID: id, Name: name, Owner: owner, Controller: owner, ScryfallID: "sf-" + id}
	for _, o := range opts {
		o(&c)
	}
	return c
}

func commander(c *protocol.CardView) { c.IsCommander = true }

func zone(cards ...protocol.CardView) protocol.ZoneView {
	return protocol.ZoneView{Count: len(cards), Cards: cards}
}

func seatPtr(n int) *int { return &n }

func fixture() Input {
	view := protocol.GameView{
		StartingSeat: 1,
		Battlefield: zone(
			card("c-ana-cmdr", "Atraxa, Praetors' Voice", ana, commander),
			card("c-bolt", "Lightning Bolt", ana),
		),
		Seats: []protocol.PlayerView{
			{ID: ana, Seat: 0, Life: 31, MulligansTaken: 1, CommanderDamage: map[string]int{}},
			{ID: bo, Seat: 1, Life: 0, Eliminated: true, CommanderDamage: map[string]int{"c-ana-cmdr": 14},
				Command: zone(card("c-bo-cmdr", "Krenko, Mob Boss", bo, commander))},
			{ID: cy, Seat: 2, Life: 12, CommanderDamage: map[string]int{"c-ana-cmdr": 5},
				Graveyard: zone(card("c-sol", "Sol Ring", cy), card("c-sol2", "Sol Ring", cy))},
			{ID: dee, Seat: 3, Life: 40, Eliminated: true, CommanderDamage: map[string]int{}},
		},
	}
	log := []protocol.LogEvent{
		{Kind: protocol.LogStep, Turn: 1, Round: 1, Seat: 1},
		{Kind: protocol.LogDraw, Turn: 1, Seat: 1, Amount: 7},
		{Kind: protocol.LogCast, Turn: 1, Seat: 2, CardID: "c-sol"},
		{Kind: protocol.LogCast, Turn: 3, Seat: 2, CardID: "c-sol2"},
		{Kind: protocol.LogCast, Turn: 2, Seat: 0, CardID: "c-ana-cmdr"},
		// A cast whose card the view no longer holds still counts as a
		// spell, but has no card line.
		{Kind: protocol.LogCast, Turn: 2, Seat: 1, CardID: "gone"},
		{Kind: protocol.LogToken, Turn: 2, Seat: 1, CardID: "t1"},
		{Kind: protocol.LogToken, Turn: 2, Seat: 1, CardID: "t2"},
		{Kind: protocol.LogEliminated, Turn: 4, Round: 2, Seat: 3, Amount: 1, Cause: "concede"},
		{Kind: protocol.LogAttack, Turn: 5, Seat: 0, CardID: "c-ana-cmdr", TargetSeat: seatPtr(1)},
		{Kind: protocol.LogDamage, Turn: 5, Seat: 0, CardID: "c-ana-cmdr", Amount: 14, Combat: true, TargetSeat: seatPtr(1)},
		{Kind: protocol.LogLife, Turn: 5, Seat: 1, Amount: -14},
		// No seat on the entry: the card's controller is responsible.
		{Kind: protocol.LogDamage, Turn: 5, Seat: protocol.NoSeat, CardID: "c-bolt", Amount: 3, TargetSeat: seatPtr(2)},
		{Kind: protocol.LogLife, Turn: 5, Seat: 2, Amount: -3},
		// Damage to a permanent, and to yourself, is not damage to players.
		{Kind: protocol.LogDamage, Turn: 5, Seat: 0, CardID: "c-bolt", Amount: 3, Target: "some-creature"},
		{Kind: protocol.LogDamage, Turn: 5, Seat: 2, CardID: "c-sol", Amount: 1, TargetSeat: seatPtr(2)},
		{Kind: protocol.LogLife, Turn: 5, Seat: 0, Amount: 6},
		{Kind: protocol.LogLife, Turn: 6, Seat: 1, Amount: -26},
		{Kind: protocol.LogEliminated, Turn: 6, Round: 3, Seat: 1, Cause: "life"},
		{Kind: protocol.LogGameOver, Turn: 6, Round: 3, Seat: 0},
	}
	return Input{
		GameID:      uuid.MustParse("6f1c1a52-1b44-4c55-9a1c-3c5e4f8d2a10"),
		View:        view,
		Log:         log,
		Outcome:     "win",
		WinnerSeat:  seatPtr(0),
		Source:      SourceLive,
		LogComplete: true,
		Now:         time.Date(2026, 10, 13, 20, 0, 0, 0, time.UTC),
	}
}

func TestComputeTheGame(t *testing.T) {
	g := Compute(fixture())
	if g.Version != Version || g.Source != SourceLive || !g.LogComplete {
		t.Errorf("header: %+v", g)
	}
	if g.Turns != 6 || g.Rounds != 3 {
		t.Errorf("turns %d rounds %d, want 6 and 3", g.Turns, g.Rounds)
	}
	if g.StartingSeat == nil || *g.StartingSeat != 1 {
		t.Errorf("starting seat %v, want 1", g.StartingSeat)
	}
	if len(g.Seats) != 4 {
		t.Fatalf("%d seats, want 4", len(g.Seats))
	}
	for i, s := range g.Seats {
		if s.Seat != i {
			t.Errorf("seat %d is at %d: seats must be in seat order", s.Seat, i)
		}
	}
}

func TestComputeResultsAndPlaces(t *testing.T) {
	g := Compute(fixture())
	want := []struct {
		result   Result
		place    int
		elimTurn int // 0: not eliminated
		conceded bool
	}{
		{ResultWin, 1, 0, false},
		{ResultLoss, 3, 6, false}, // second out of four
		{ResultLoss, 2, 0, false}, // still standing at another's win
		{ResultLoss, 4, 4, true},  // first out
	}
	for i, w := range want {
		s := g.Seats[i]
		if s.Result != w.result || s.Place == nil || *s.Place != w.place || s.Conceded != w.conceded {
			t.Errorf("seat %d: result %s place %v conceded %v; want %s %d %v", i, s.Result, s.Place, s.Conceded, w.result, w.place, w.conceded)
		}
		switch {
		case w.elimTurn == 0 && s.EliminatedTurn != nil:
			t.Errorf("seat %d eliminated on %d, want standing", i, *s.EliminatedTurn)
		case w.elimTurn != 0 && (s.EliminatedTurn == nil || *s.EliminatedTurn != w.elimTurn):
			t.Errorf("seat %d eliminated on %v, want %d", i, s.EliminatedTurn, w.elimTurn)
		}
	}
}

func TestComputeCounts(t *testing.T) {
	g := Compute(fixture())
	a, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]

	if a.FinalLife != 31 || a.LifeGained != 6 || a.LifeLost != 0 || a.Mulligans != 1 {
		t.Errorf("Ana life: %+v", a)
	}
	if b.LifeLost != 40 || b.FinalLife != 0 {
		t.Errorf("Bo lost %d (final %d), want 40 and 0", b.LifeLost, b.FinalLife)
	}
	if a.DamageToPlayers != 17 || a.CombatDamageToPlayers != 14 {
		t.Errorf("Ana damage %d combat %d, want 17 and 14", a.DamageToPlayers, a.CombatDamageToPlayers)
	}
	if c.DamageToPlayers != 0 {
		t.Errorf("Cy dealt %d to players, want 0: damage to yourself does not count", c.DamageToPlayers)
	}
	if a.CommanderDamageDealt != 19 || b.CommanderDamageTaken != 14 || c.CommanderDamageTaken != 5 {
		t.Errorf("commander damage: Ana dealt %d, Bo took %d, Cy took %d; want 19, 14, 5",
			a.CommanderDamageDealt, b.CommanderDamageTaken, c.CommanderDamageTaken)
	}
	if a.Attacks != 1 || b.TokensCreated != 2 || b.CardsDrawn != 7 {
		t.Errorf("attacks %d tokens %d draws %d, want 1 2 7", a.Attacks, b.TokensCreated, b.CardsDrawn)
	}
	if a.SpellsCast != 1 || b.SpellsCast != 1 || c.SpellsCast != 2 || d.SpellsCast != 0 {
		t.Errorf("spells %d %d %d %d, want 1 1 2 0", a.SpellsCast, b.SpellsCast, c.SpellsCast, d.SpellsCast)
	}
	if !reflect.DeepEqual(a.Commanders, []string{"Atraxa, Praetors' Voice"}) ||
		!reflect.DeepEqual(b.Commanders, []string{"Krenko, Mob Boss"}) ||
		!reflect.DeepEqual(c.Commanders, []string{}) {
		t.Errorf("commanders: %v %v %v", a.Commanders, b.Commanders, c.Commanders)
	}
}

func TestComputeCards(t *testing.T) {
	g := Compute(fixture())
	if got, want := g.Seats[2].Cards, []Card{{Name: "Sol Ring", ScryfallID: "sf-c-sol", Casts: 2, FirstCastTurn: 1}}; !reflect.DeepEqual(got, want) {
		t.Errorf("Cy's cards %+v, want %+v", got, want)
	}
	if got := g.Seats[1].Cards; len(got) != 0 {
		t.Errorf("Bo's unresolvable cast made a card line: %+v", got)
	}
	if got := g.Seats[0].Cards; len(got) != 1 || got[0].Name != "Atraxa, Praetors' Voice" || got[0].FirstCastTurn != 2 {
		t.Errorf("Ana's cards %+v", got)
	}
}

func TestComputeADraw(t *testing.T) {
	in := fixture()
	in.Outcome, in.WinnerSeat = "draw", nil
	g := Compute(in)
	for i, w := range []struct {
		result Result
		place  int
	}{{ResultDraw, 1}, {ResultLoss, 3}, {ResultDraw, 1}, {ResultLoss, 4}} {
		s := g.Seats[i]
		if s.Result != w.result || s.Place == nil || *s.Place != w.place {
			t.Errorf("seat %d: %s %v, want %s %d", i, s.Result, s.Place, w.result, w.place)
		}
	}
}

func TestComputeAClosedTableHasNoResultOrPlace(t *testing.T) {
	in := fixture()
	in.Outcome, in.WinnerSeat = "", nil
	for _, s := range Compute(in).Seats {
		if s.Result != ResultNone || s.Place != nil {
			t.Errorf("seat %d: %s %v, want none and no place", s.Seat, s.Result, s.Place)
		}
	}
}

func TestComputeAGameThatNeverStarted(t *testing.T) {
	in := fixture()
	in.Log, in.Outcome, in.WinnerSeat = nil, "", nil
	g := Compute(in)
	if g.Turns != 0 || g.StartingSeat != nil {
		t.Errorf("turns %d starting %v, want 0 and nil", g.Turns, g.StartingSeat)
	}
}

func TestComputeAnEliminationTheLogLost(t *testing.T) {
	// A replay merge with a gap: Dee's line is missing, but the view says
	// she is out. She left first, at an unknown turn.
	in := fixture()
	in.LogComplete = false
	var kept []protocol.LogEvent
	for _, e := range in.Log {
		if !(e.Kind == protocol.LogEliminated && e.Seat == 3) {
			kept = append(kept, e)
		}
	}
	in.Log = kept
	g := Compute(in)
	d := g.Seats[3]
	if d.EliminatedTurn != nil || d.Place == nil || *d.Place != 4 || d.Result != ResultLoss {
		t.Errorf("Dee: turn %v place %v result %s; want unknown turn, 4th, loss", d.EliminatedTurn, d.Place, d.Result)
	}
	if g.LogComplete {
		t.Error("LogComplete should carry through as false")
	}
}
