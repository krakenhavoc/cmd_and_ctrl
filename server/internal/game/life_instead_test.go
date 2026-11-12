package game

import (
	"testing"

	"github.com/google/uuid"
)

// life_instead_test.go covers the engine half of "If your life total
// would be reduced to 0 or less, instead ... your life total becomes 3.
// Then ... you lose the game" (#5678, life_instead.go) with a test
// replacement, so every landing site is exercised without a card: the
// life tail, the cancelled branch, the damage result and the CR 616
// resume. The card's own tests are in cards/effects.

// becomesThreeInstead is the clause for `only`, without the transform:
// the event becomes "life total becomes 3", and with thenLose the event
// carries the loss of the game.
func becomesThreeInstead(only uuid.UUID, thenLose bool) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventChangeLife},
		AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
			return ev.LifePlayer == only && g.LifeWouldBeReducedToZeroOrLessForEffect(ev)
		},
		Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
			g.LifeTotalBecomesInsteadForEffect(ev, 3)
			if thenLose && !g.ReplacementIsPreviewForEffect(ev) {
				ev.LifeThenLoses = uuid.New()
			}
			return nil
		},
		Label: "life total becomes 3 instead",
	}
}

// lifeEventsFor are the EventChangeLife amounts `player` has seen.
func lifeEventsFor(g *Game, player uuid.UUID) []int {
	var out []int
	for _, ev := range g.Events {
		if ev.Kind == EventChangeLife && ev.Target == player {
			out = append(out, ev.Amount)
		}
	}
	return out
}

func TestLifeWouldBeReducedToZeroOrLess(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	p := g.Seats[0]
	for _, tc := range []struct {
		life, delta int
		want        bool
	}{
		{10, -9, false},
		{10, -10, true},
		{10, -15, true},
		{10, 5, false},
		{10, 0, false},
		{-2, -1, true}, // already below, and reduced further
		{-2, 1, false}, // a gain is not a reduction
	} {
		p.Life = tc.life
		ev := &ReplacementEvent{Kind: RepEventLife, LifePlayer: p.ID, LifeDelta: tc.delta}
		if got := g.LifeWouldBeReducedToZeroOrLessForEffect(ev); got != tc.want {
			t.Errorf("life %d, delta %d: got %v, want %v", tc.life, tc.delta, got, tc.want)
		}
	}
	if g.LifeWouldBeReducedToZeroOrLessForEffect(&ReplacementEvent{Kind: RepEventDraw, LifePlayer: p.ID, LifeDelta: -99}) {
		t.Error("an event that is not a life event matched")
	}
}

// A loss rewritten into "becomes 3" loses the difference, and the
// caller's continuation is told what really moved (CR 119.5).
func TestLifeBecomesThreeInsteadLosesTheDifference(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	p := g.Seats[1]
	p.Life = 10
	var told []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(becomesThreeInstead(p.ID, false))
		if err := g.ChangePlayerLifeKeyedForEffect(uuid.Nil, p.ID, -25, recordTotalsTail(&told)); err != nil {
			t.Fatalf("ChangePlayerLifeKeyedForEffect: %v", err)
		}
	})
	if p.Life != 3 || p.Eliminated {
		t.Fatalf("life %d (eliminated %v), want 3 and still in the game", p.Life, p.Eliminated)
	}
	if len(told) != 1 || told[0] != -7 {
		t.Errorf("the continuation was told %v, want [-7]", told)
	}
	g.RunStateChecksForTest()
	if p.Eliminated {
		t.Error("a player at 3 life lost to the state-based check")
	}
}

// At exactly 3 the rewritten event moves no life: nothing lands and no
// EventChangeLife is emitted. Below 3 it is a gain, which a player who
// can't gain life does not get.
func TestLifeBecomesThreeInsteadAtAndBelowThree(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	p := g.Seats[1]
	g.WithWriteLock(func() { g.RegisterReplacementForTest(becomesThreeInstead(p.ID, false)) })

	p.Life = 3
	var told []int
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeKeyedForEffect(uuid.Nil, p.ID, -5, recordTotalsTail(&told)); err != nil {
			t.Fatalf("lose 5 at 3: %v", err)
		}
	})
	if p.Life != 3 || len(lifeEventsFor(g, p.ID)) != 0 {
		t.Errorf("at 3: life %d, life events %v; want 3 and none", p.Life, lifeEventsFor(g, p.ID))
	}
	if len(told) != 1 || told[0] != 0 {
		t.Errorf("the continuation was told %v, want [0]", told)
	}

	p.Life = 2
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, -5); err != nil {
			t.Fatalf("lose 5 at 2: %v", err)
		}
	})
	if got := lifeEventsFor(g, p.ID); p.Life != 3 || len(got) != 1 || got[0] != 1 {
		t.Errorf("at 2: life %d, life events %v; want 3 and one gain of 1", p.Life, got)
	}

	p.Life = 2
	g.WithWriteLock(func() {
		g.PlayerCantGainLifeForEffect(uuid.Nil, p.ID, g.UntilEndOfTurnDuration(), "can't gain life")
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, -5); err != nil {
			t.Fatalf("lose 5 at 2 with no life gain: %v", err)
		}
	})
	if p.Life != 2 {
		t.Errorf("a player who can't gain life went from 2 to %d", p.Life)
	}
}

// "Then ... you lose the game": the life lands first, then the player
// leaves at once (CR 104.3e), and the game-over check waits for the next
// state-based pass.
func TestLifeThenLosesAfterTheLifeLands(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	p := g.Seats[1]
	p.Life = 10
	var told []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(becomesThreeInstead(p.ID, true))
		if err := g.ChangePlayerLifeKeyedForEffect(uuid.Nil, p.ID, -10, recordTotalsTail(&told)); err != nil {
			t.Fatalf("ChangePlayerLifeKeyedForEffect: %v", err)
		}
	})
	if p.Life != 3 {
		t.Errorf("life %d, want 3: the total becomes 3 before the loss", p.Life)
	}
	if !p.Eliminated || p.LossCause != LossEffect {
		t.Fatalf("eliminated %v by %q, want an effect loss", p.Eliminated, p.LossCause)
	}
	if len(told) != 1 || told[0] != -7 {
		t.Errorf("the continuation was told %v, want [-7]", told)
	}
	if !g.ActiveSeatLeftPending {
		t.Error("the departure's game-over check was not left for the next state-based pass")
	}
	g.RunStateChecksForTest()
	if g.ActiveSeatLeftPending || g.State != StateActive {
		t.Errorf("after the pass: pending %v, state %q; want settled and two players playing on", g.ActiveSeatLeftPending, g.State)
	}
}

// The last opponent losing this way ends the game at the next pass.
func TestLifeThenLosesEndsATwoPlayerGame(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	p := g.Seats[1]
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(becomesThreeInstead(p.ID, true))
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, -p.Life)
	})
	g.RunStateChecksForTest()
	if g.State != StateEnded || g.Outcome == nil || g.Outcome.Winner != g.Seats[0].ID {
		t.Fatalf("state %q, outcome %+v; want seat 0 the winner", g.State, g.Outcome)
	}
}

// The life damage costs runs the same window (ADR 0169): a loss rewritten
// into a gain is gained, the damage is still dealt whole, and "then you
// lose" lands after the damage.
func TestDamageLifeLossRewrittenIntoAGainAndALoss(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	p := g.Seats[1]
	g.WithWriteLock(func() { g.RegisterReplacementForTest(becomesThreeInstead(p.ID, false)) })

	p.Life = 2
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, p.ID, 5); err != nil {
			t.Fatalf("5 damage at 2: %v", err)
		}
	})
	if got := lifeEventsFor(g, p.ID); p.Life != 3 || len(got) != 1 || got[0] != 1 {
		t.Errorf("5 damage at 2: life %d, life events %v; want 3 and one gain of 1", p.Life, got)
	}
	var dmg *Event
	for i := range g.Events {
		if g.Events[i].Kind == EventDealDamage && g.Events[i].Target == p.ID {
			dmg = &g.Events[i]
		}
	}
	if dmg == nil || dmg.Amount != 5 || dmg.DamageLifeLoss() != 0 {
		t.Fatalf("damage event %+v, want 5 dealt and no life lost", dmg)
	}

	p.Life = 10
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, p.ID, 12); err != nil {
			t.Fatalf("12 damage at 10: %v", err)
		}
	})
	if p.Life != 3 || p.Eliminated {
		t.Fatalf("12 damage at 10: life %d, eliminated %v; want 3 and alive", p.Life, p.Eliminated)
	}
	for i := range g.Events {
		if g.Events[i].Kind == EventDealDamage && g.Events[i].Target == p.ID {
			dmg = &g.Events[i]
		}
	}
	if dmg.Amount != 12 || dmg.DamageLifeLoss() != 7 {
		t.Errorf("damage event: %d dealt, %d life lost; want 12 and 7", dmg.Amount, dmg.DamageLifeLoss())
	}

	q := g.Seats[2]
	q.Life = 4
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(becomesThreeInstead(q.ID, true))
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, q.ID, 9); err != nil {
			t.Fatalf("9 damage at 4: %v", err)
		}
	})
	if q.Life != 3 || !q.Eliminated || q.LossCause != LossEffect {
		t.Errorf("9 damage at 4 with the loss: life %d, eliminated %v by %q; want 3 and an effect loss", q.Life, q.Eliminated, q.LossCause)
	}
}

// The loss rides the event across a CR 616 ordering prompt: applied
// first, it is still owed when the prompt over the two effects left is
// answered.
func TestLifeThenLosesIsCarriedAcrossAnOrderingPrompt(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	p := g.Seats[1]
	p.Life = 10
	instead := becomesThreeInstead(p.ID, true)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(instead)
		g.RegisterReplacementForTest(plusOneLifeReplacement(p.ID))
		g.RegisterReplacementForTest(doubleLifeReplacement(p.ID))
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, -10); err != nil {
			t.Fatalf("ChangePlayerLifeForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the CR 616 ordering prompt", len(g.PendingChoices))
	}
	c := g.PendingChoices[0]
	if len(c.ReplacementEffectIDs) != 3 {
		t.Fatalf("the prompt orders %d effects, want 3", len(c.ReplacementEffectIDs))
	}
	if p.Eliminated || p.Life != 10 {
		t.Fatalf("before the answer: life %d, eliminated %v", p.Life, p.Eliminated)
	}
	// The order offered is the order registered: "becomes 3" first, so
	// the loss of 7 is then one more and doubled, 16, and the player is
	// at -6 when the loss of the game lands.
	if err := g.ResolveReplacementOrder(c.ID, c.Chooser, c.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if !p.Eliminated || p.LossCause != LossEffect {
		t.Fatalf("after the answer: eliminated %v by %q, want an effect loss", p.Eliminated, p.LossCause)
	}
	if p.Life != -6 {
		t.Errorf("life %d, want -6: becomes 3, then one more, then doubled", p.Life)
	}
}

// A player who can't lose the game keeps the 3 life and stays.
func TestLifeThenLosesRespectsCantLose(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	p := g.Seats[1]
	p.Life = 5
	g.WithWriteLock(func() {
		g.GrantGameEndGateForEffect(p.ID, GameEndGate{Scope: GateYou, CantLose: true}, "can't lose", uuid.Nil, g.UntilEndOfTurnDuration())
		g.RegisterReplacementForTest(becomesThreeInstead(p.ID, true))
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, -8)
	})
	g.RunStateChecksForTest()
	if p.Eliminated || p.Life != 3 || g.State != StateActive {
		t.Errorf("life %d, eliminated %v, state %q; want 3, in the game, playing on", p.Life, p.Eliminated, g.State)
	}
}

// The read-only preview (the auto-tapper's question) rewrites the event
// and nothing else.
func TestLifePreviewOfAGainTurnedIntoALethalLoss(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	p := g.Seats[1]
	p.Life = 5
	toLoss := ReplacementEffect{
		Watches: []EventKind{EventChangeLife},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventLife && ev.LifePlayer == p.ID && ev.LifeDelta > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.LifeDelta = -ev.LifeDelta
			return nil
		},
		Label: "loses that much instead",
	}
	var got int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(toLoss)
		g.RegisterReplacementForTest(becomesThreeInstead(p.ID, true))
		got = g.lifeChangePreviewLocked(p.ID, uuid.Nil, 6)
	})
	if got != -2 {
		t.Errorf("preview of gaining 6 at 5: %d, want -2 (the total becomes 3)", got)
	}
	if p.Eliminated || p.Life != 5 {
		t.Errorf("the preview touched the game: life %d, eliminated %v", p.Life, p.Eliminated)
	}
}
