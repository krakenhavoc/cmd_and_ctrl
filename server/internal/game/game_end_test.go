package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// game_end_test.go — ADR 0057 (#749): winning and losing the game by
// an effect, the collect-then-apply SBA loss pass, and the "can't
// lose" / "can't win" gates. The card-level tests (Platinum Angel,
// Laboratory Maniac, Felidar Sovereign, ...) live in cards/effects;
// these pin the engine's half with catalog hooks stubbed, because the
// game package cannot import the catalog.

const (
	testAngelOracle      = "test-749-platinum-angel"
	testPersecutorOracle = "test-749-abyssal-persecutor"
)

// withCatalogGameEndGates stubs the DERIVED half of the gate reader:
// a battlefield permanent keyed testAngelOracle prints Platinum
// Angel's two gates, and one keyed testPersecutorOracle prints Abyssal
// Persecutor's.
func withCatalogGameEndGates(t *testing.T) {
	t.Helper()
	prev := CatalogGameEndGates
	CatalogGameEndGates = func(key string) []GameEndGate {
		switch key {
		case testAngelOracle:
			return []GameEndGate{
				{Scope: GateYou, CantLose: true},
				{Scope: GateOpponents, CantWin: true},
			}
		case testPersecutorOracle:
			return []GameEndGate{
				{Scope: GateYou, CantWin: true},
				{Scope: GateOpponents, CantLose: true},
			}
		}
		return nil
	}
	t.Cleanup(func() { CatalogGameEndGates = prev })
}

// seedGatePermanent puts a permanent with the given oracle ID onto the
// battlefield under `owner` and returns its instance ID.
func seedGatePermanent(g *Game, owner *Player, oracle, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.OracleID = oracle
	c.TypeLine = "Artifact Creature — Angel"
	c.Power, c.Toughness = 4, 4
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// destroyGate takes a permanent off the board the way a destroy
// would, through the ordinary zone move.
func destroyGate(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}

func runChecks(g *Game) {
	g.WithWriteLock(func() { g.runStateChecksLocked() })
}

func lastEvent(g *Game, kind EventKind) (Event, bool) {
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == kind {
			return g.Events[i], true
		}
	}
	return Event{}, false
}

// seedEffectTrigger puts a permanent under `owner` whose upkeep trigger
// runs `effect`, and returns the permanent's instance ID. The effect
// sees the resolving item, as a catalog Effect does.
func seedEffectTrigger(t *testing.T, g *Game, owner *Player, oracle string, effect func(g *Game, item *StackItem) error) uuid.UUID {
	t.Helper()
	c := NewCard("Test Win Trigger", owner.ID)
	c.OracleID = oracle
	c.TypeLine = "Enchantment"
	g.Battlefield.PushTop(c)
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventBeginUpkeep},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return newTriggeredItemForTest(source, "Test Win Trigger — you win", effect)
			},
		}}
	})
	return c.InstanceID
}

// winEffect is effects.WinTheGame, spelled out in the game package.
func winEffect(g *Game, item *StackItem) error {
	if _, err := g.WinTheGameForEffect(item.Controller, item.SourceCardID); err != nil {
		return err
	}
	if g.State != StateActive {
		return ErrStopResolution
	}
	return nil
}

// loseEffect is effects.LoseTheGame on the item's own controller.
func loseEffect(g *Game, item *StackItem) error {
	lost, err := g.LoseTheGameForEffect(item.Controller, item.SourceCardID)
	if err != nil {
		return err
	}
	if g.State != StateActive || lost {
		return ErrStopResolution
	}
	return nil
}

// --- Decision 3 + 5: an effect win ------------------------------------

// TestEffectWinEndsAFourSeatGame is test plan item 2: a Felidar-style
// upkeep win in a four-player game. The game ends with a recorded
// result, nobody is eliminated (the others did not leave the game —
// it ended), no prompt is left open, and the stop is a clean one.
func TestEffectWinEndsAFourSeatGame(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	active := g.Seats[0]
	src := seedEffectTrigger(t, g, active, "test-749-win", winEffect)
	raiseUpkeepTrigger(t, g)
	resolveTop(t, g)

	if g.State != StateEnded {
		t.Fatalf("state %s, want ended", g.State)
	}
	want := GameOutcome{Kind: OutcomeWin, Winner: active.ID, Cause: OutcomeCauseEffect, Source: src}
	if g.Outcome == nil || *g.Outcome != want {
		t.Fatalf("outcome %+v, want %+v", g.Outcome, want)
	}
	for _, p := range g.Seats {
		if p.Eliminated {
			t.Errorf("%s is eliminated; an effect win eliminates nobody (ADR 0057 Decision 5)", p.Name)
		}
	}
	if len(g.PendingChoices) != 0 || len(g.PendingTriggers) != 0 {
		t.Errorf("prompts left open over the game-over banner: %d choices, %d triggers",
			len(g.PendingChoices), len(g.PendingTriggers))
	}
	if n := countEvents(g, EventGameOver); n != 1 {
		t.Errorf("%d game_over events, want 1", n)
	}
	if n := countEvents(g, EventEffectError); n != 0 {
		t.Errorf("%d effect errors; ErrStopResolution is a clean stop", n)
	}
	if n := countEvents(g, EventPlayerEliminated); n != 0 {
		t.Errorf("%d eliminations; a win is not every opponent losing (CR 104.3h is not offered)", n)
	}
	if seat, ok := g.WinnerSeat(); !ok || seat != active.Seat {
		t.Errorf("WinnerSeat = %d, %v; want %d — the lobby's winner_seat reads the outcome", seat, ok, active.Seat)
	}
}

// TestEffectWinEndsATwoSeatGame is the same win at a two-seat table.
func TestEffectWinEndsATwoSeatGame(t *testing.T) {
	g := newActiveGame(t)
	active, other := g.Seats[0], g.Seats[1]
	seedEffectTrigger(t, g, active, "test-749-win-2", winEffect)
	raiseUpkeepTrigger(t, g)
	resolveTop(t, g)

	if g.State != StateEnded || g.Outcome == nil || g.Outcome.Winner != active.ID {
		t.Fatalf("state %s outcome %+v, want %s to have won", g.State, g.Outcome, active.Name)
	}
	if other.Eliminated {
		t.Errorf("the loser of an effect win is marked eliminated")
	}
	// The game is over: nothing more can be done.
	if err := g.PassPriority(); err == nil {
		t.Errorf("PassPriority on an ended game succeeded")
	}
}

// TestEffectWinIsIgnoredOnceTheGameHasEnded: a second "you win" in a
// game that already ended changes nothing — the first outcome stands.
func TestEffectWinIsIgnoredOnceTheGameHasEnded(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if won, err := g.WinTheGameForEffect(a.ID, uuid.Nil); !won || err != nil {
			t.Fatalf("first win: %v %v", won, err)
		}
		if won, err := g.WinTheGameForEffect(b.ID, uuid.Nil); won || err != nil {
			t.Fatalf("second win: won=%v err=%v, want false, nil", won, err)
		}
	})
	if g.Outcome.Winner != a.ID {
		t.Errorf("winner changed to %v", g.Outcome.Winner)
	}
}

// TestJaceMinusEightWinsPastAPendingEmptyDraw is test plan item 11:
// drawing past the end of the library sets the CR 704.5b flag, and an
// effect win in the same resolution still lands, because the flag is
// read only at the next check (the Jace, Wielder of Mysteries ruling).
func TestJaceMinusEightWinsPastAPendingEmptyDraw(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	for me.Library.Size() > 3 {
		if _, err := me.Library.PopTop(); err != nil {
			t.Fatal(err)
		}
	}
	g.WithWriteLock(func() {
		_ = g.DrawNForEffect(me.ID, 7)
		if !me.AttemptedEmptyDraw {
			t.Fatalf("drawing seven from three did not set the flag")
		}
		if won, err := g.WinTheGameForEffect(me.ID, uuid.Nil); !won || err != nil {
			t.Fatalf("win: %v %v", won, err)
		}
	})
	if g.Outcome == nil || g.Outcome.Winner != me.ID || me.Eliminated {
		t.Fatalf("outcome %+v eliminated %v; the effect win lands before the check", g.Outcome, me.Eliminated)
	}
}

// --- Decision 3: an effect loss is immediate --------------------------

// TestEffectLossOfAnotherPlayerIsImmediate is test plan item 3's first
// half: the loser leaves at once, before any state-based check, and
// the game goes on without them.
func TestEffectLossOfAnotherPlayerIsImmediate(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	loser := g.Seats[2]
	g.WithWriteLock(func() {
		lost, err := g.LoseTheGameForEffect(loser.ID, uuid.Nil)
		if !lost || err != nil {
			t.Fatalf("LoseTheGameForEffect: %v %v", lost, err)
		}
		if !loser.Eliminated {
			t.Fatalf("the loser is still in the game before any check — an effect loss is not an SBA (CR 104.3e)")
		}
	})
	if g.State != StateActive {
		t.Fatalf("three seats, one lost: the game should go on")
	}
	ev, ok := lastEvent(g, EventPlayerEliminated)
	if !ok || ev.Label != string(LossEffect) {
		t.Errorf("elimination event %+v, want Label %q", ev, LossEffect)
	}
}

// TestEffectLossDownToOnePlayerEndsTheGame is test plan item 4.
func TestEffectLossDownToOnePlayerEndsTheGame(t *testing.T) {
	g := newActiveGame(t)
	active, other := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if _, err := g.LoseTheGameForEffect(other.ID, uuid.Nil); err != nil {
			t.Fatal(err)
		}
	})
	want := GameOutcome{Kind: OutcomeWin, Winner: active.ID, Cause: OutcomeCauseLastStanding}
	if g.State != StateEnded || g.Outcome == nil || *g.Outcome != want {
		t.Fatalf("state %s outcome %+v, want %+v", g.State, g.Outcome, want)
	}
}

// TestActiveSeatEffectLossDefersTheRotation is test plan item 3's
// second half: the active player's own trigger makes them lose. They
// leave at once, but the turn does not move on inside the resolution
// (ADR 0059 Decision 6); the next SBA pass does it, once.
func TestActiveSeatEffectLossDefersTheRotation(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	active := g.Seats[0]
	seedEffectTrigger(t, g, active, "test-749-lose", loseEffect)
	raiseUpkeepTrigger(t, g)
	turnBefore := g.Turn
	began := countEvents(g, EventTurnBegan)

	resolveTop(t, g)

	if !active.Eliminated {
		t.Fatalf("the active player has not left")
	}
	if g.Turn.ActiveSeat != turnBefore.ActiveSeat || g.Turn.Seq != turnBefore.Seq {
		t.Fatalf("the turn moved inside the resolution: %+v -> %+v", turnBefore, g.Turn)
	}
	if countEvents(g, EventTurnBegan) != began {
		t.Fatalf("a turn began inside the resolution")
	}
	if !g.ActiveSeatLeftPending {
		t.Fatalf("ActiveSeatLeftPending not set")
	}
	if n := countEvents(g, EventEffectError); n != 0 {
		t.Errorf("%d effect errors; the stop is clean", n)
	}

	// The flag survives Clone and a snapshot round trip.
	if !g.Clone().ActiveSeatLeftPending {
		t.Errorf("Clone dropped ActiveSeatLeftPending")
	}
	snap := g.CaptureSnapshot()
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var back GameSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if !back.ActiveSeatLeftPending {
		t.Errorf("the snapshot dropped ActiveSeatLeftPending")
	}

	runChecks(g)
	if g.ActiveSeatLeftPending {
		t.Errorf("the SBA pass did not consume the flag")
	}
	if g.Turn.ActiveSeat != 1 {
		t.Errorf("active seat %d after the pass, want 1 — the next seat's turn", g.Turn.ActiveSeat)
	}
	if got := countEvents(g, EventTurnBegan) - began; got != 1 {
		t.Errorf("%d turns began, want exactly 1", got)
	}
}

// TestActiveSeatEffectLossInATwoSeatGameEndsAtTheSweep: the same loss
// at a two-seat table ends the game at the next pass, and no next turn
// begins.
func TestActiveSeatEffectLossInATwoSeatGameEndsAtTheSweep(t *testing.T) {
	g := newActiveGame(t)
	active, other := g.Seats[0], g.Seats[1]
	seedEffectTrigger(t, g, active, "test-749-lose-2", loseEffect)
	raiseUpkeepTrigger(t, g)
	began := countEvents(g, EventTurnBegan)
	resolveTop(t, g)
	if g.State != StateActive {
		t.Fatalf("the game ended inside the resolution; the active seat's departure is deferred")
	}
	runChecks(g)
	want := GameOutcome{Kind: OutcomeWin, Winner: other.ID, Cause: OutcomeCauseLastStanding}
	if g.State != StateEnded || g.Outcome == nil || *g.Outcome != want {
		t.Fatalf("state %s outcome %+v, want %+v", g.State, g.Outcome, want)
	}
	if countEvents(g, EventTurnBegan) != began {
		t.Errorf("a turn began after the game ended")
	}
}

// --- Decision 2: the SBA loss pass ------------------------------------

// TestSimultaneousLossesLeaveTheLastSeatStanding is test plan item 13.
func TestSimultaneousLossesLeaveTheLastSeatStanding(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	g.Seats[0].Life = 0
	g.Seats[1].Life = -3
	runChecks(g)
	want := GameOutcome{Kind: OutcomeWin, Winner: g.Seats[2].ID, Cause: OutcomeCauseLastStanding}
	if g.Outcome == nil || *g.Outcome != want {
		t.Fatalf("outcome %+v, want %+v", g.Outcome, want)
	}
	if n := countEvents(g, EventPlayerEliminated); n != 2 {
		t.Errorf("%d eliminations, want 2", n)
	}
}

// TestEveryoneLosingAtOnceIsADraw is CR 104.4a, and CR 104.3f's order:
// the last seat standing in the pass that makes it lose is a loser.
func TestEveryoneLosingAtOnceIsADraw(t *testing.T) {
	g := newActiveGame(t)
	g.Seats[0].Life = 0
	g.Seats[1].Life = 0
	runChecks(g)
	want := GameOutcome{Kind: OutcomeDraw, Cause: OutcomeCauseAllLost}
	if g.State != StateEnded || g.Outcome == nil || *g.Outcome != want {
		t.Fatalf("state %s outcome %+v, want %+v", g.State, g.Outcome, want)
	}
	if _, ok := g.WinnerSeat(); ok {
		t.Errorf("a draw has a winner seat")
	}
	ev, _ := lastEvent(g, EventGameOver)
	if ev.Actor != uuid.Nil || ev.Label != OutcomeCauseAllLost {
		t.Errorf("game_over event %+v", ev)
	}
}

// TestStateLossesNameTheirCause: each SBA loss labels its elimination.
func TestStateLossesNameTheirCause(t *testing.T) {
	cases := []struct {
		name  string
		set   func(p *Player)
		cause LossCause
	}{
		{"life", func(p *Player) { p.Life = 0 }, LossLife},
		{"empty draw", func(p *Player) { p.AttemptedEmptyDraw = true }, LossEmptyDraw},
		{"poison", func(p *Player) { p.Poison = PoisonLethal }, LossPoison},
		{"commander damage", func(p *Player) { p.CommanderDamage = map[uuid.UUID]int{uuid.New(): 21} }, LossCommanderDamage},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGameWithSeats(t, 3)
			p := g.Seats[1]
			tc.set(p)
			runChecks(g)
			ev, ok := lastEvent(g, EventPlayerEliminated)
			if !ok || ev.Actor != p.ID || ev.Label != string(tc.cause) {
				t.Fatalf("elimination %+v, want %s for %s", ev, tc.cause, p.Name)
			}
		})
	}
}

// --- Decision 4: the gates -------------------------------------------

// TestPlatinumAngelHoldsAtZeroLifeUntilItLeaves is test plan item 5:
// under the Angel a player at 0 life stays in through several checks
// and nothing is logged; when the Angel dies they lose at the next
// check, to life, with no window.
func TestPlatinumAngelHoldsAtZeroLifeUntilItLeaves(t *testing.T) {
	withCatalogGameEndGates(t)
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[1]
	angel := seedGatePermanent(g, me, testAngelOracle, "Test Platinum Angel")
	me.Life = -5
	for i := 0; i < 3; i++ {
		runChecks(g)
	}
	if me.Eliminated {
		t.Fatalf("eliminated at -5 life under their own Platinum Angel")
	}
	if n := countEvents(g, EventPlayerEliminated); n != 0 {
		t.Errorf("a stopped loss logged %d eliminations", n)
	}
	var causes []LossCause
	g.ReadSnapshot(func() { causes = g.CantLoseCausesForEffect(me) })
	if len(causes) != len(GatableLossCauses) {
		t.Errorf("cant_lose %v, want every gatable cause", causes)
	}

	destroyGate(t, g, angel)
	runChecks(g)
	if !me.Eliminated {
		t.Fatalf("still in the game at -5 life after the Angel died")
	}
	ev, _ := lastEvent(g, EventPlayerEliminated)
	if ev.Label != string(LossLife) {
		t.Errorf("lost to %q, want life", ev.Label)
	}
}

// TestPlatinumAngelCoversOnlyItsController: an opponent at 0 life
// still loses, because "you can't lose" is about the Angel's
// controller.
func TestPlatinumAngelCoversOnlyItsController(t *testing.T) {
	withCatalogGameEndGates(t)
	g := newActiveGameWithSeats(t, 3)
	me, opp := g.Seats[1], g.Seats[2]
	seedGatePermanent(g, me, testAngelOracle, "Test Platinum Angel")
	me.Life, opp.Life = 0, 0
	runChecks(g)
	if me.Eliminated {
		t.Errorf("the Angel's controller lost")
	}
	if !opp.Eliminated {
		t.Errorf("the opponent at 0 life did not lose")
	}
}

// TestEmptyDrawUnderTheAngelIsForgotten is test plan item 6: CR
// 704.5b reads only draws since the last check, so the flag is cleared
// by the check the Angel stopped, and removing the Angel later does
// not kill the player for an old draw.
func TestEmptyDrawUnderTheAngelIsForgotten(t *testing.T) {
	withCatalogGameEndGates(t)
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[1]
	angel := seedGatePermanent(g, me, testAngelOracle, "Test Platinum Angel")
	me.AttemptedEmptyDraw = true
	runChecks(g)
	if me.Eliminated {
		t.Fatalf("lost to an empty draw under the Angel")
	}
	if me.AttemptedEmptyDraw {
		t.Fatalf("the flag survived the check (CR 704.5b: since the last check)")
	}
	destroyGate(t, g, angel)
	runChecks(g)
	if me.Eliminated {
		t.Errorf("lost to an old draw once the Angel left")
	}
}

// TestAngelStopsPoisonCommanderDamageAndEffectLosses is test plan
// item 7, plus the effect loss: every gatable cause, stopped.
func TestAngelStopsPoisonCommanderDamageAndEffectLosses(t *testing.T) {
	withCatalogGameEndGates(t)
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[1]
	seedGatePermanent(g, me, testAngelOracle, "Test Platinum Angel")
	me.Poison = PoisonLethal + 2
	me.CommanderDamage = map[uuid.UUID]int{uuid.New(): 30}
	runChecks(g)
	g.WithWriteLock(func() {
		if lost, err := g.LoseTheGameForEffect(me.ID, uuid.Nil); lost || err != nil {
			t.Errorf("effect loss under the Angel: lost=%v err=%v", lost, err)
		}
	})
	if me.Eliminated {
		t.Fatalf("lost under the Angel")
	}
}

// TestConcedeUnderTheAngelLoses is test plan item 8: a concession is
// never gated (CR 104.3a), and in a two-seat game the opponent wins.
func TestConcedeUnderTheAngelLoses(t *testing.T) {
	withCatalogGameEndGates(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedGatePermanent(g, me, testAngelOracle, "Test Platinum Angel")
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if !me.Eliminated {
		t.Fatalf("a player who can't lose could not concede")
	}
	ev, _ := lastEvent(g, EventPlayerEliminated)
	if ev.Label != string(LossConcede) {
		t.Errorf("elimination label %q, want concede", ev.Label)
	}
	want := GameOutcome{Kind: OutcomeWin, Winner: opp.ID, Cause: OutcomeCauseLastStanding}
	if g.Outcome == nil || *g.Outcome != want {
		t.Errorf("outcome %+v, want %+v", g.Outcome, want)
	}
}

// TestAngelWithoutAbilitiesGatesNothing is test plan item 9: the gate
// is read through CatalogAbilityKey, so an Angel that has lost all its
// abilities (CR 613.1f) stops nothing.
func TestAngelWithoutAbilitiesGatesNothing(t *testing.T) {
	withCatalogGameEndGates(t)
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[1]
	angel := seedGatePermanent(g, me, testAngelOracle, "Test Platinum Angel")
	c := findCard(g, angel)
	eff := c.printedCharacteristic()
	eff.AbilitiesRemoved = true
	c.effective = &eff
	var can bool
	g.ReadSnapshot(func() { can = g.canLoseLocked(me, LossLife) })
	if !can {
		t.Fatalf("an Angel with no abilities still says its controller can't lose")
	}
}

// TestAngelStopsAnOpponentsEffectWin: "your opponents can't win the
// game" — the win is prevented, logged once, and the game goes on.
func TestAngelStopsAnOpponentsEffectWin(t *testing.T) {
	withCatalogGameEndGates(t)
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	angel := seedGatePermanent(g, me, testAngelOracle, "Test Platinum Angel")
	src := uuid.New()
	g.WithWriteLock(func() {
		if won, err := g.WinTheGameForEffect(opp.ID, src); won || err != nil {
			t.Fatalf("win under an opponent's Angel: won=%v err=%v", won, err)
		}
		// The Angel's own controller CAN win.
		if !g.canWinLocked(me) {
			t.Errorf("the Angel stops its own controller winning")
		}
	})
	if g.State != StateActive {
		t.Fatalf("the game ended")
	}
	ev, ok := lastEvent(g, EventWinPrevented)
	if !ok || ev.Actor != opp.ID || ev.Source != src || ev.Target != angel {
		t.Errorf("win_prevented event %+v", ev)
	}
	if n := countEvents(g, EventWinPrevented); n != 1 {
		t.Errorf("%d win_prevented events, want 1", n)
	}
}

// TestAbyssalPersecutorAndCR104_2a is test plan item 12: the
// Persecutor's controller can't win by effect and keeps an opponent at
// -5 in the game; when that opponent concedes, the controller wins
// anyway, because the last player standing overrides "can't win".
func TestAbyssalPersecutorAndCR104_2a(t *testing.T) {
	withCatalogGameEndGates(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedGatePermanent(g, me, testPersecutorOracle, "Test Abyssal Persecutor")
	opp.Life = -5
	runChecks(g)
	if opp.Eliminated {
		t.Fatalf("the Persecutor's opponent lost at -5")
	}
	g.WithWriteLock(func() {
		if won, _ := g.WinTheGameForEffect(me.ID, uuid.Nil); won {
			t.Fatalf("the Persecutor's controller won by effect")
		}
	})
	if err := g.Concede(opp.ID); err != nil {
		t.Fatal(err)
	}
	want := GameOutcome{Kind: OutcomeWin, Winner: me.ID, Cause: OutcomeCauseLastStanding}
	if g.Outcome == nil || *g.Outcome != want {
		t.Fatalf("outcome %+v, want %+v (CR 104.2a overrides can't win)", g.Outcome, want)
	}
}

// TestGrantedGateLastsTheTurn is test plan item 14's registry half:
// Angel's Grace's two gates, granted until end of turn, protect the
// caster and stop the opponents winning, survive Clone and a snapshot
// round trip, and are gone the next turn.
func TestGrantedGateLastsTheTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := uuid.New()
	g.WithWriteLock(func() {
		d := g.UntilEndOfTurnDuration()
		g.GrantGameEndGateForEffect(me.ID, GameEndGate{Scope: GateYou, CantLose: true}, "Angel's Grace", src, d)
		g.GrantGameEndGateForEffect(me.ID, GameEndGate{Scope: GateOpponents, CantWin: true}, "Angel's Grace", src, d)
		// A While on a granted gate is refused.
		g.GrantGameEndGateForEffect(me.ID, GameEndGate{Scope: GateYou, CantLose: true,
			While: func(*Game, Card) bool { return true }}, "bad", src, d)
	})
	if n := len(me.Statics); n != 2 {
		t.Fatalf("%d statics, want 2 (a While is refused)", n)
	}
	check := func(g *Game, label string) {
		t.Helper()
		me, opp := g.Seats[0], g.Seats[1]
		g.ReadSnapshot(func() {
			if g.canLoseLocked(me, LossLife) {
				t.Errorf("%s: the caster can lose", label)
			}
			if !g.canLoseLocked(opp, LossLife) {
				t.Errorf("%s: the opponent can't lose — the grant is the caster's", label)
			}
			if g.canWinLocked(opp) {
				t.Errorf("%s: the opponent can win", label)
			}
			gates := g.GameEndGatesForEffect(me)
			if len(gates) != 1 || !gates[0].ThisTurn || gates[0].SourceName != "Angel's Grace" {
				t.Errorf("%s: end gates %+v", label, gates)
			}
		})
	}
	check(g, "live")
	check(g.Clone(), "clone")
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	check(restored, "snapshot")

	me.Life = 0
	runChecks(g)
	if me.Eliminated {
		t.Fatalf("lost at 0 life this turn under the grant")
	}
	g.WithWriteLock(func() {
		g.clearCombatLocked()
		g.sweepTurnEndLocked()
		g.beginNextTurnLocked()
		g.runStepEntryHooksLocked()
	})
	g.ReadSnapshot(func() {
		if !g.canWinLocked(opp) {
			t.Errorf("the grant outlived its turn")
		}
	})
	runChecks(g)
	if !me.Eliminated {
		t.Errorf("still in at 0 life the turn after the grant")
	}
}

// TestOutcomeSurvivesCloneAndSnapshot is test plan item 15's value
// half (the drift test holds the field rows).
func TestOutcomeSurvivesCloneAndSnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	src := uuid.New()
	g.WithWriteLock(func() { _, _ = g.WinTheGameForEffect(g.Seats[3].ID, src) })
	want := *g.Outcome
	if c := g.Clone(); c.Outcome == nil || *c.Outcome != want || c.Outcome == g.Outcome {
		t.Errorf("clone outcome %+v (shared pointer: %v)", c.Outcome, c.Outcome == g.Outcome)
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if restored.Outcome == nil || *restored.Outcome != want {
		t.Errorf("restored outcome %+v, want %+v", restored.Outcome, want)
	}
}

// TestEndRecordsNoOutcome: an admin closing a table is not a result.
func TestEndRecordsNoOutcome(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.End()
	if g.Outcome != nil {
		t.Errorf("End() recorded %+v", g.Outcome)
	}
	if countEvents(g, EventGameOver) != 0 {
		t.Errorf("End() emitted game_over")
	}
}
