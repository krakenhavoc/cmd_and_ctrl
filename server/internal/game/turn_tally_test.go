package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_tally_test.go — #586. The tally is bumped by the same events
// everything else reads, resets on a new turn, and survives Clone /
// RestoreFrom and the snapshot round trip (the drift test covers the
// field's classification; TestTurnTallySurvivesClone covers the
// deep copy).

func emit(g *Game, evs ...Event) {
	g.WithWriteLock(func() {
		for _, ev := range evs {
			g.EmitEvent(ev)
		}
	})
}

func TestTurnTallyCountsPlayerEvents(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me, 3); err != nil {
			t.Fatal(err)
		}
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me, -2); err != nil {
			t.Fatal(err)
		}
	})
	emit(g,
		Event{Kind: EventDealDamage, Actor: me, Source: uuid.New(), Target: opp, Amount: 4, Combat: true},
		Event{Kind: EventDealDamage, Actor: opp, Source: uuid.New(), Target: me, Amount: 1},
		Event{Kind: EventDrawCard, Actor: me, CardID: uuid.New()},
		Event{Kind: EventDrawCard, Actor: me, CardID: uuid.New()},
		Event{Kind: EventTokenCreated, Actor: me, CardID: uuid.New()},
		Event{Kind: EventSacrifice, Actor: opp, CardID: uuid.New()},
		Event{Kind: EventAttack, Actor: me, CardID: uuid.New(), Target: opp},
		// Actor-less events (admin, SBA) are not anyone's.
		Event{Kind: EventDrawCard, CardID: uuid.New()},
	)

	mine, theirs := g.TurnTallyFor(me), g.TurnTallyFor(opp)
	if mine.LifeGained != 3 || mine.LifeLost != 3 {
		t.Errorf("me: gained %d lost %d, want 3 and 3 (2 life loss + 1 damage)", mine.LifeGained, mine.LifeLost)
	}
	if theirs.LifeLost != 4 || theirs.LifeGained != 0 {
		t.Errorf("opp: gained %d lost %d, want 0 and 4", theirs.LifeGained, theirs.LifeLost)
	}
	if mine.CombatDamageToPlayers != 4 || theirs.CombatDamageToPlayers != 0 {
		t.Errorf("combat damage to players: me %d opp %d, want 4 and 0 (non-combat damage does not count)", mine.CombatDamageToPlayers, theirs.CombatDamageToPlayers)
	}
	if mine.CardsDrawn != 2 || mine.TokensCreated != 1 || mine.AttacksDeclared != 1 || theirs.PermanentsSacrificed != 1 {
		t.Errorf("counts: %+v / %+v", mine, theirs)
	}
	if g.TurnTallyFor(uuid.Nil) != (PlayerTurnTally{}) {
		t.Error("an actor-less event was counted against the nil player")
	}
}

func TestTurnTallyCountsCreatureDeathsAndLandfall(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	bear := uuid.New()
	land := uuid.New()
	rock := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp, Controller: opp})
		g.Battlefield.PushTop(Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me, Controller: me})
		g.Battlefield.PushTop(Card{InstanceID: rock, Name: "Sol Ring", TypeLine: "Artifact", Owner: me, Controller: me})
	})
	emit(g,
		Event{Kind: EventETB, CardID: land},
		Event{Kind: EventETB, CardID: rock},
	)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
		if err := g.SacrificePermanentForEffect(rock); err != nil {
			t.Fatal(err)
		}
	})

	if got := g.TurnTallyFor(me).LandsEntered; got != 1 {
		t.Errorf("lands entered for me: %d, want 1 (the artifact is not a land)", got)
	}
	if g.TurnTally.CreaturesDied != 1 {
		t.Errorf("creatures died: %d, want 1 (the sacrificed artifact is not a creature)", g.TurnTally.CreaturesDied)
	}
	if got := g.TurnTallyFor(opp); got.CreaturesDied != 1 || got.PermanentsSacrificed != 1 {
		t.Errorf("opp: %+v, want one creature died and one permanent sacrificed", got)
	}
	if got := g.TurnTallyFor(me); got.CreaturesDied != 0 || got.PermanentsSacrificed != 1 {
		t.Errorf("me: %+v, want no creature died and one permanent sacrificed", got)
	}
}

func TestTurnTallyResolutionsAndTriggersPerAbility(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	src := uuid.New()
	emit(g,
		Event{Kind: EventResolve, Source: src, Label: "Ping"},
		Event{Kind: EventResolve, Source: src, Label: "Ping"},
		Event{Kind: EventResolve, Source: src, Label: "Drain"},
		Event{Kind: EventTrigger, Source: src, Label: "Ping"},
		Event{Kind: EventResolve, Source: uuid.New(), Label: "Ping"},
	)
	if got := g.ResolvedThisTurn(src, "Ping"); got != 2 {
		t.Errorf("Ping resolved %d, want 2", got)
	}
	if got := g.ResolvedThisTurn(src, ""); got != 3 {
		t.Errorf("all abilities of the source resolved %d, want 3", got)
	}
	if got := g.TriggeredThisTurn(src, "Ping"); got != 1 {
		t.Errorf("Ping triggered %d, want 1", got)
	}
	if got := g.TriggeredThisTurn(src, "Drain"); got != 0 {
		t.Errorf("Drain triggered %d, want 0", got)
	}
}

func TestTurnTallyResetsOnANewTurnAndBoundsEventsThisTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	emit(g, Event{Kind: EventDrawCard, Actor: me, CardID: uuid.New()})
	if g.TurnTallyFor(me).CardsDrawn != 1 {
		t.Fatal("setup: the draw was not counted")
	}
	before := len(g.Events)
	seat := g.Turn.ActiveSeat
	for i := 0; i < 40 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == seat {
		t.Fatal("setup: the turn never passed")
	}
	if got := g.TurnTallyFor(me); got != (PlayerTurnTally{}) && got.CardsDrawn == 1 {
		t.Errorf("tally survived the turn change: %+v", got)
	}
	if g.TurnTally.FirstEvent <= before {
		t.Errorf("FirstEvent %d should point past the previous turn's events (%d)", g.TurnTally.FirstEvent, before)
	}
	n := len(g.EventsThisTurn())
	emit(g, Event{Kind: EventTokenCreated, Actor: me, CardID: uuid.New()}, Event{Kind: EventTokenCreated, Actor: me, CardID: uuid.New()})
	if got := len(g.EventsThisTurn()); got != n+2 {
		t.Errorf("EventsThisTurn grew by %d, want 2", got-n)
	}
	for _, ev := range g.EventsThisTurn() {
		if ev.Seq <= uint64(before) && ev.Kind == EventDrawCard && ev.Actor == me {
			t.Error("EventsThisTurn reaches back into the previous turn")
		}
	}
}

func TestTurnTallySurvivesClone(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	src := uuid.New()
	emit(g, Event{Kind: EventDrawCard, Actor: me, CardID: uuid.New()}, Event{Kind: EventResolve, Source: src, Label: "X"})

	snap := g.Clone()
	emit(g, Event{Kind: EventDrawCard, Actor: me, CardID: uuid.New()}, Event{Kind: EventResolve, Source: src, Label: "X"})
	if snap.TurnTallyFor(me).CardsDrawn != 1 || snap.ResolvedThisTurn(src, "X") != 1 {
		t.Error("the clone shares its tally maps with the live game")
	}
	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if g.TurnTallyFor(me).CardsDrawn != 1 || g.ResolvedThisTurn(src, "X") != 1 {
		t.Errorf("restore did not bring the tally back: %+v", g.TurnTallyFor(me))
	}
}
