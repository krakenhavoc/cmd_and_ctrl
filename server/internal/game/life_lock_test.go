package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// life_lock_test.go — #1200, CR 119.7 / CR 119.8 / CR 120.3a: a
// player whose LIFE TOTAL CAN'T CHANGE.
//
// Every test here is an assertion about one of the four consumers
// ADR 0085 names, written from the side a player at the table would
// notice: the drain that does not drain, the Bolt that is still dealt
// and still moves nothing, the cost that is refused before it is
// offered, and the boundary the lock ends on.
//
// The table has FOUR seats everywhere a turn boundary matters,
// because "until your next turn" is a whole rotation and a two-seat
// table cannot tell "ends as my next turn begins" from "ends as the
// next turn begins".

// withCatalogLifeTotalLock stubs the DERIVED half of the reader — a
// battlefield permanent whose printed static locks its controller's
// life total (Platinum Emperion). The game package cannot import
// cards/effects, which is why every catalog hook is a var.
func withCatalogLifeTotalLock(t *testing.T, fn func(oracleID string) bool) {
	t.Helper()
	prev := CatalogPlayerLifeTotalLocked
	CatalogPlayerLifeTotalLocked = fn
	t.Cleanup(func() { CatalogPlayerLifeTotalLocked = prev })
}

// lockLifeTotal applies Teferi's Protection's clause directly, for a
// duration the caller names. Returns the duration it was stamped with.
func lockLifeTotal(g *Game, p *Player, d Duration) Duration {
	g.WithWriteLock(func() {
		g.GrantLifeTotalLockForEffect(p.ID, "Test — your life total can't change", uuid.Nil, d)
	})
	return d
}

// lockUntilNextTurn is the shape both catalogued spells print.
func lockUntilNextTurn(g *Game, p *Player) Duration {
	var d Duration
	g.WithWriteLock(func() { d = g.UntilYourNextTurnDuration(p.ID) })
	return lockLifeTotal(g, p, d)
}

func lifeTotalLocked(g *Game, p *Player) bool {
	var out bool
	g.ReadSnapshot(func() { out = g.PlayerLifeTotalCantChangeLocked(p) })
	return out
}

// --- CR 119.7 / CR 119.8: the life window --------------------------

// TestLockedLifeTotalRefusesGainAndLoss is the headline, through the
// one entry point every catalog life change goes down. Gain and loss
// in one test because a predicate that read the SIGN of the delta
// would pass either alone.
func TestLockedLifeTotalRefusesGainAndLoss(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lockUntilNextTurn(g, me)

	startMe, startOpp := me.Life, opp.Life
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 5); err != nil {
			t.Fatalf("gain at the locked seat: %v", err)
		}
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -7); err != nil {
			t.Fatalf("loss at the locked seat: %v", err)
		}
		// The control: a lock on one player does not touch another.
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -7); err != nil {
			t.Fatalf("loss at the unlocked seat: %v", err)
		}
	})
	if me.Life != startMe {
		t.Errorf("locked seat is at %d, want %d — CR 119.7 and CR 119.8", me.Life, startMe)
	}
	if opp.Life != startOpp-7 {
		t.Errorf("unlocked seat is at %d, want %d — the built-in is cancelling changes it should not",
			opp.Life, startOpp-7)
	}
}

// TestLockedLifeChangeEmitsNoEventAndTellsTheContinuationZero is the
// half a card author has to be able to rely on. CR 614.10 with a null
// replacement means nothing happened, so a "whenever you gain life"
// trigger must see nothing — and a drain adding up "the life lost this
// way" must still be told, or it waits forever.
func TestLockedLifeChangeEmitsNoEventAndTellsTheContinuationZero(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	lockUntilNextTurn(g, me)

	before := len(g.Events)
	applied := -1
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeThenForEffect(uuid.Nil, me.ID, -4,
			func(_ *Game, n int) error { applied = n; return nil }); err != nil {
			t.Fatalf("ChangePlayerLifeThenForEffect: %v", err)
		}
	})
	if applied != 0 {
		t.Errorf("the continuation was told %d moved, want 0", applied)
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventChangeLife {
			t.Errorf("a cancelled life change emitted EventChangeLife (%+v)", ev)
		}
	}
}

// TestLockedSeatContributesNothingToADrainTotal is the composition
// case LoseLifeEachThenForEffect exists for: three opponents, one of
// them locked, and the caster gains exactly what the OTHER two lost.
func TestLockedSeatContributesNothingToADrainTotal(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b, c := g.Seats[1], g.Seats[2], g.Seats[3]
	lockUntilNextTurn(g, b)

	total := -1
	g.WithWriteLock(func() {
		if err := g.LoseLifeEachThenForEffect(uuid.Nil, []uuid.UUID{a.ID, b.ID, c.ID}, 3,
			func(_ *Game, lost int) error { total = lost; return nil }); err != nil {
			t.Fatalf("LoseLifeEachThenForEffect: %v", err)
		}
	})
	if total != 6 {
		t.Errorf("the drain totalled %d, want 6 — the locked seat lost nothing and the others lost 3 each", total)
	}
	if b.Life != 40 {
		t.Errorf("the locked seat is at %d, want its starting total", b.Life)
	}
}

// TestALifeLockPreemptsADoubler — the CR 616 half. An amount
// replacement rewrites the delta and the lock cancels whatever it
// rewrote it to, so both orders reach zero; Preemptive is what stops
// the engine asking a question with one answer (ADR 0085 Decision 4).
func TestALifeLockPreemptsADoubler(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	lockUntilNextTurn(g, me)

	doubled := false
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches: []EventKind{EventChangeLife},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventLife && ev.LifeDelta > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			doubled = true
			ev.LifeDelta *= 2
			return nil
		},
		Label: "Test Faithmender",
	})

	start := me.Life
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3); err != nil {
			t.Fatalf("gain under a doubler at a locked seat: %v", err)
		}
	})
	if me.Life != start {
		t.Errorf("locked seat gained %d under a doubler", me.Life-start)
	}
	if doubled {
		t.Error("the doubler applied before the lock; the lock is not Preemptive")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("a CR 616 ordering prompt was queued for an event with one outcome: %d", len(g.PendingChoices))
	}
}

// --- CR 120.3a: damage is still DEALT -------------------------------

// TestDamageToALockedPlayerIsStillDealtButMovesNoLife is Platinum
// Emperion's printed ruling, through both damage entry points. The
// event has to fire — "whenever ~ is dealt damage" triggers key on it
// — and only the life loss CR 120.3a would have caused is gone.
func TestDamageToALockedPlayerIsStillDealtButMovesNoLife(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lockUntilNextTurn(g, me)

	start := me.Life
	bolt := NewCard("Lightning Bolt", opp.ID)
	bolt.TypeLine = "Instant"
	bolt.Colors = []string{"R"}
	before := len(g.Events)
	g.WithWriteLock(func() {
		g.Stack.PushTop(bolt)
		if err := g.DealDamageToPlayerForEffect(bolt.InstanceID, me.ID, 3); err != nil {
			t.Fatalf("noncombat damage at the locked player: %v", err)
		}
	})
	if me.Life != start {
		t.Errorf("locked player lost %d life to noncombat damage", start-me.Life)
	}
	dealt := false
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventDealDamage && ev.Target == me.ID {
			dealt = true
		}
	}
	if !dealt {
		t.Error("the damage was not DEALT at all; CR 119.8 stops the life loss, not the damage")
	}

	// Combat, the other entry point, and the CR 903.10a tally with it.
	attacker := pushColouredCreature(g, opp, "Big Attacker", []string{"G"})
	g.WithWriteLock(func() {
		g.markCombatDamageToPlayerLocked(me.ID, attacker, 8, "")
	})
	if me.Life != start {
		t.Errorf("locked player lost %d life to combat damage", start-me.Life)
	}
}

// TestCommanderDamageStillAccruesAgainstALockedPlayer — CR 903.10a
// counts damage DEALT by a commander, not life lost, so the 21-damage
// loss is not something this clause turns off. The card does not say
// "you can't lose the game" and must not behave as though it did.
func TestCommanderDamageStillAccruesAgainstALockedPlayer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lockUntilNextTurn(g, me)

	cmdr := pushColouredCreature(g, opp, "General", []string{"W"})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cmdr {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
		g.markCombatDamageToPlayerLocked(me.ID, cmdr, 7, "")
	})
	if got := me.CommanderDamage[cmdr]; got != 7 {
		t.Errorf("commander damage tally = %d, want 7 — CR 903.10a counts damage dealt", got)
	}
}

// TestPoisonCountersStillLandOnALockedPlayer — the lock watches the
// LIFE event and nothing else. A counter on a player is its own
// CR 614 window (ADR 0056) and has to be untouched, or an infect deck
// would find Platinum Emperion a blank.
func TestPoisonCountersStillLandOnALockedPlayer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	lockUntilNextTurn(g, me)

	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(me.ID, CounterPoison, 3); err != nil {
			t.Fatalf("poison at the locked player: %v", err)
		}
	})
	if got := me.Counters[CounterPoison]; got != 3 {
		t.Errorf("poison counters = %d, want 3 — the lock is not a counter prohibition", got)
	}
}

// --- CR 119.8 / CR 614.17b: the cost --------------------------------

// TestALockedPlayerCannotPayLifeAsACost is CR 119.8's second
// sentence through the payment path, and the assertion that the
// refusal costs nothing: no life moved, so there is nothing to put
// back.
func TestALockedPlayerCannotPayLifeAsACost(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	lockUntilNextTurn(g, me)

	start := me.Life
	var err error
	g.WithWriteLock(func() { err = g.PayLifeForEffect(uuid.Nil, me.ID, 2) })
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("paying 2 life under a lock: got %v, want ErrInvalidParam", err)
	}
	if me.Life != start {
		t.Errorf("the refused payment moved %d life", start-me.Life)
	}
	// Paying ZERO is still legal — Platinum Emperion's own
	// parenthetical says "except 0".
	g.WithWriteLock(func() { err = g.PayLifeForEffect(uuid.Nil, me.ID, 0) })
	if err != nil {
		t.Errorf("paying 0 life under a lock: %v", err)
	}
}

// TestActivationWithALifeCostIsRefusedUpFront is CR 614.17b at the
// VALIDATOR rather than at the payment: a cost that cannot be paid
// cannot be chosen, so the announcement is refused before anything
// else in it is paid.
func TestActivationWithALifeCostIsRefusedUpFront(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := phyrexianPodSource(g, me, AbilityCost{Life: 3})

	// Unlocked, the ability is payable — otherwise the refusal below
	// would prove nothing.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activation with a life cost and no lock: %v", err)
	}
	life := me.Life
	lockUntilNextTurn(g, me)
	err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{Strict: true})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("activation with a life cost under a lock: got %v, want ErrInvalidParam", err)
	}
	if me.Life != life {
		t.Errorf("the refused activation moved %d life", life-me.Life)
	}
}

// TestPhyrexianSymbolCannotBeClaimedUnderALock — CR 107.4's "or 2
// life" is a life payment like any other, so the claim is refused at
// announce with nothing spent.
func TestPhyrexianSymbolCannotBeClaimedUnderALock(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := phyrexianPodSource(g, me, abilityCostMana("{1}{G/P}"))
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	lockUntilNextTurn(g, me)

	life := me.Life
	err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{Strict: true, PhyrexianLife: 1})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claiming a Phyrexian symbol under a lock: got %v, want ErrInvalidParam", err)
	}
	if me.Life != life {
		t.Errorf("the refused claim moved %d life", life-me.Life)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("the refused claim spent mana: %v", me.ManaPool)
	}
}

// --- the duration ---------------------------------------------------

// TestLifeLockEndsAsYourNextTurnBegins pins the boundary against the
// one function that decides when any continuous effect is over. Four
// seats, so the lock has to sit through three opponents' turns and
// three cleanup steps and be gone the instant its own player's turn
// begins (CR 500.1).
func TestLifeLockEndsAsYourNextTurnBegins(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lockUntilNextTurn(g, me)

	if !lifeTotalLocked(g, me) {
		t.Fatal("the lock did not take at all")
	}
	for i := 0; i < 3; i++ {
		advanceOneTurn(t, g)
		if !lifeTotalLocked(g, me) {
			t.Fatalf("the lock ended after %d opponent turns; it lasts until YOUR next turn", i+1)
		}
	}
	advanceOneTurn(t, g)
	if lifeTotalLocked(g, me) {
		t.Error("the lock survived the beginning of its own player's next turn")
	}
	// The sweep, not just the reader: an expired entry the slice still
	// holds would leak into every snapshot for the rest of the game.
	if n := len(me.Statics); n != 0 {
		t.Errorf("Player.Statics still holds %d expired entries; the sweep did not run", n)
	}
	// And the total moves again.
	start := me.Life
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3); err != nil {
			t.Fatalf("loss after the lock expired: %v", err)
		}
	})
	if me.Life != start-3 {
		t.Errorf("life = %d after the lock expired, want %d", me.Life, start-3)
	}
}

// TestLifeLockSurvivesASkippedTurn — CR 800.4m and CR 611.2b
// together. A seat the rotation STEPS OVER (its player has left) has
// its never-taken turn counted for ITS own durations, and must not
// end anybody else's: the lock belongs to seat 0 and is keyed on seat
// 0's seat-turn counter, so a seat between here and there vanishing
// changes when the lock ends only by making it arrive sooner.
func TestLifeLockSurvivesASkippedTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	skipped := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	lockUntilNextTurn(g, me)

	advanceOneTurn(t, g) // the seat immediately after me
	if !lifeTotalLocked(g, me) {
		t.Fatal("the lock ended on the first opponent's turn")
	}
	// The next seat leaves the game; the rotation steps over it and
	// counts its never-taken turn (rotation.go, CR 800.4m).
	g.WithWriteLock(func() { skipped.Eliminated = true })
	advanceOneTurn(t, g)
	if !lifeTotalLocked(g, me) {
		t.Fatal("a SKIPPED seat's turn ended a lock that is keyed on another seat's counter")
	}
	if skipped.TurnsBegun == 0 {
		t.Fatal("the skipped seat's never-taken turn was not counted; the test is vacuous")
	}
	advanceOneTurn(t, g) // back to me
	if lifeTotalLocked(g, me) {
		t.Error("the lock survived the beginning of its own player's next turn")
	}
}

// TestExpiredLifeLockIsRefusedBeforeTheSweepRuns is why the reader
// tests the duration as well as the sweep: the sweep runs at known
// moments and the reader has to be right BETWEEN them. Stamped into
// the past by hand, which is the only way to observe the window.
func TestExpiredLifeLockIsRefusedBeforeTheSweepRuns(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	lockLifeTotal(g, me, Duration{Kind: UntilYourNextTurn, Player: me.ID, ExpiresAtTurnsBegun: 1})
	if len(me.Statics) != 1 {
		t.Fatal("the stale lock was not stored; the test is vacuous")
	}
	if lifeTotalLocked(g, me) {
		t.Error("an expired lock still answers; the reader is trusting the sweep")
	}
	start := me.Life
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2); err != nil {
			t.Fatalf("loss under an expired lock: %v", err)
		}
	})
	if me.Life != start-2 {
		t.Errorf("an expired lock still cancelled a life change (life %d, want %d)", me.Life, start-2)
	}
}

// TestAnUnstampedLifeLockIsUntilEndOfTurn — the grant's default. A
// card file that forgot its window gets the narrowest real one rather
// than a permanent lock, the direction every other default in the
// catalog errs in.
func TestAnUnstampedLifeLockIsUntilEndOfTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lockLifeTotal(g, me, Duration{})
	if !lifeTotalLocked(g, me) {
		t.Fatal("the lock did not take")
	}
	if got := me.Statics[0].Duration.Kind; got != UntilEndOfTurn {
		t.Fatalf("an unstamped lock came back as %v, want until end of turn", got)
	}
	advanceOneTurn(t, g)
	if lifeTotalLocked(g, me) {
		t.Error("an until-end-of-turn lock outlived its turn")
	}
}

// --- the derived half ------------------------------------------------

// TestTwoDerivedLocksComposeAndOneLeavingDoesNotUnlock is the whole
// argument for deriving rather than writing: two Platinum Emperions,
// one dies, the total is still locked. A "set on enter, restore on
// leave" design gets exactly this case wrong.
func TestTwoDerivedLocksComposeAndOneLeavingDoesNotUnlock(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const emperion = "test-life-lock-emperion"
	withCatalogLifeTotalLock(t, func(id string) bool { return id == emperion })

	first := pushLeyline(t, g, me, emperion)
	pushLeyline(t, g, me, emperion)
	if !lifeTotalLocked(g, me) {
		t.Fatal("two Emperions and no lock")
	}
	if lifeTotalLocked(g, opp) {
		t.Error("the Emperion locked a seat that does not control it")
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, first); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if !lifeTotalLocked(g, me) {
		t.Error("one Emperion leaving unlocked a total the other is still locking")
	}
}

// TestADerivedLockGoesAwayWithItsPermanent — nothing is stored, so
// nothing has to be unwound.
func TestADerivedLockGoesAwayWithItsPermanent(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	const emperion = "test-life-lock-emperion-solo"
	withCatalogLifeTotalLock(t, func(id string) bool { return id == emperion })
	id := pushLeyline(t, g, me, emperion)
	if !lifeTotalLocked(g, me) {
		t.Fatal("no lock with the Emperion out")
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, id); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if lifeTotalLocked(g, me) {
		t.Error("the lock outlived the permanent that granted it")
	}
	if n := len(me.Statics); n != 0 {
		t.Errorf("the derived half wrote %d entries to Player.Statics; it must write none", n)
	}
}

// --- undo, clone and the snapshot ------------------------------------

// TestLifeLockSurvivesCloneUndoAndSnapshot — the granted half is
// state, so it has to rewind and it has to persist. Three assertions
// in one test because the three mechanisms share one contract: the
// entry is plain data, copied by value into a fresh backing array.
func TestLifeLockSurvivesCloneUndoAndSnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	window := lockUntilNextTurn(g, me)

	var before *Game
	g.ReadSnapshot(func() { before = g.cloneLocked() })

	// The clone carries it, with the duration verbatim.
	got := before.Seats[0].Statics
	if len(got) != 1 || !got[0].LifeTotalLocked {
		t.Fatalf("the clone holds %+v, want one life-total lock", got)
	}
	if got[0].Duration != window {
		t.Errorf("the clone's duration is %+v, want %+v", got[0].Duration, window)
	}

	// The snapshot carries it too, and the game stays restorable —
	// which is the whole argument for not putting this in
	// TurnScopedReplacements (ADR 0085 Decision 1).
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Errorf("a locked life total cost the table its restore point: %+v", snap.Continuations)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if !lifeTotalLocked(restored, restored.Seats[0]) {
		t.Error("the lock did not survive a snapshot round-trip")
	}

	// And an undo across a SECOND grant rewinds to exactly one.
	g.WithWriteLock(func() {
		g.GrantLifeTotalLockForEffect(me.ID, "second lock", uuid.Nil, IndefiniteDuration())
	})
	if n := len(me.Statics); n != 2 {
		t.Fatalf("Player.Statics has %d entries after the second grant, want 2", n)
	}
	if n := len(before.Seats[0].Statics); n != 1 {
		t.Fatalf("the clone grew to %d entries; the backing array is shared", n)
	}
	g.RestoreFrom(before)
	if n := len(g.Seats[0].Statics); n != 1 {
		t.Errorf("after the undo Player.Statics holds %d entries, want 1", n)
	}
	if !lifeTotalLocked(g, g.Seats[0]) {
		t.Error("the undo took the ORIGINAL lock away too")
	}
}

// TestUndoOfTheGrantItselfUnlocksTheTotal is the other direction: a
// lock applied after the snapshot must be gone, and the life total
// must move again once it is.
func TestUndoOfTheGrantItselfUnlocksTheTotal(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]

	var before *Game
	g.ReadSnapshot(func() { before = g.cloneLocked() })
	lockUntilNextTurn(g, me)
	if !lifeTotalLocked(g, me) {
		t.Fatal("the lock did not take")
	}

	g.RestoreFrom(before)
	if lifeTotalLocked(g, g.Seats[0]) {
		t.Fatal("the lock survived an undo of the grant that made it")
	}
	start := g.Seats[0].Life
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[0].ID, -3); err != nil {
			t.Fatalf("loss after the undo: %v", err)
		}
	})
	if g.Seats[0].Life != start-3 {
		t.Errorf("life = %d after the undo, want %d", g.Seats[0].Life, start-3)
	}
}
