package game

import (
	"testing"

	"github.com/google/uuid"
)

// damage_continuation_test.go is the #807 regression suite: what a
// paused damage event owes its CALLER.
//
// Every damage event runs the CR 614 window, so every damage event can
// pause on a CR 616 ordering prompt when two DIFFERENT damage
// replacements apply to it. The caller that noticed was Creeping
// Bloodsucker — "deals 1 damage to each opponent. You gain life equal
// to the damage dealt this way" — which read each opponent's life total
// back on the line after damaging them, before the prompt was answered,
// and gained nothing for that opponent. damage_tail.go carries the
// continuation instead, and this file pins what has to be true of it:
// it runs with the REPLACED amount, it runs after a pause, it runs with
// zero when nothing landed, and the batch form adds up across several
// targets including a paused one.
//
// This is life_continuation_test.go's sibling, on purpose and almost
// line for line — the point of #807 is that the two idioms are one
// idiom.
//
// The two-replacement board that makes a damage event prompt is shared
// with damage_tail_test.go's scenario runner in spirit but built here
// per test, because these tests care which TARGET pauses.

// preventOneDamageTo and doubleDamageTo are the two DIFFERENT effects
// it takes to make a damage event prompt: identical ones collapse
// without asking (#800) and pure cancels do too (#710). Gated on a
// target so a batch can have one leg pause and another not.
func preventOneDamageTo(only uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventDealDamage},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventDamage && (only == uuid.Nil || ev.DamageTarget == only)
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.DamageAmount--
			return nil
		},
		Label: "Prevent 1",
	}
}

func doubleDamageTo(only uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventDealDamage},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventDamage && (only == uuid.Nil || ev.DamageTarget == only)
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.DamageAmount *= 2
			return nil
		},
		Label: "Double",
	}
}

// --- (a) one target, with a continuation -----------------------------

// TestDamageContinuationRunsWithTheReplacedAmount — the unpaused floor.
// `then` is told what the window settled on, not what the caller asked
// for: 3 damage under a doubler is 6 dealt, and "you gain life equal to
// the damage dealt this way" gains 6.
func TestDamageContinuationRunsWithTheReplacedAmount(t *testing.T) {
	g := newActiveGame(t)
	var got []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(doubleDamageTo(uuid.Nil))
		err := g.DealDamageToPlayerThenForEffect(uuid.Nil, g.Seats[1].ID, 3,
			func(_ *Game, dealt int) error {
				got = append(got, dealt)
				return nil
			})
		if err != nil {
			t.Fatalf("DealDamageToPlayerThenForEffect: %v", err)
		}
	})
	if len(got) != 1 || got[0] != 6 {
		t.Fatalf("continuation saw %v, want [6] — the REPLACED amount, once", got)
	}
	if g.Seats[1].Life != StartingLife-6 {
		t.Errorf("life = %d, want %d", g.Seats[1].Life, StartingLife-6)
	}
}

// TestDamageContinuationRunsAfterACR616Pause — the bug. Two different
// damage replacements apply, the affected player is asked to order
// them, and the damage does not land until they answer. The
// continuation has to wait with it and then be told the settled amount;
// before #807 the caller was told nothing at all and carried on with a
// life total that had not moved yet.
func TestDamageContinuationRunsAfterACR616Pause(t *testing.T) {
	g := newActiveGame(t)
	var got []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(preventOneDamageTo(uuid.Nil))
		g.RegisterReplacementForTest(doubleDamageTo(uuid.Nil))
		err := g.DealDamageToPlayerThenForEffect(uuid.Nil, g.Seats[1].ID, 3,
			func(_ *Game, dealt int) error {
				got = append(got, dealt)
				return nil
			})
		if err != nil {
			t.Fatalf("DealDamageToPlayerThenForEffect: %v", err)
		}
	})
	if len(got) != 0 {
		t.Fatalf("continuation ran %v before the prompt was answered — that is the read-back bug with extra steps", got)
	}
	if g.Seats[1].Life != StartingLife {
		t.Fatalf("life moved to %d before the prompt was answered", g.Seats[1].Life)
	}

	answerOnlyPrompt(t, g)

	// Gather order is registration order, so the answer as offered is
	// prevent-one then double: (3-1)*2 = 4.
	if len(got) != 1 || got[0] != 4 {
		t.Fatalf("continuation saw %v, want [4] — (3-1)*2 after the ordering was answered", got)
	}
	if g.Seats[1].Life != StartingLife-4 {
		t.Errorf("life = %d, want %d", g.Seats[1].Life, StartingLife-4)
	}
}

// TestDamageContinuationRunsWithZeroWhenTheDamageIsPrevented — a Fog.
// Nothing landed, so "you gain life equal to the damage dealt this way"
// gains nothing; the continuation is still told, because a batch adding
// up several opponents would otherwise wait forever on the one that
// took nothing.
func TestDamageContinuationRunsWithZeroWhenTheDamageIsPrevented(t *testing.T) {
	g := newActiveGame(t)
	var got []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches:   []EventKind{EventDealDamage},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventDamage },
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Cancel()
				return nil
			},
			Label: "prevent all damage",
		})
		err := g.DealDamageToPlayerThenForEffect(uuid.Nil, g.Seats[1].ID, 4,
			func(_ *Game, dealt int) error {
				got = append(got, dealt)
				return nil
			})
		if err != nil {
			t.Fatalf("DealDamageToPlayerThenForEffect: %v", err)
		}
	})
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("continuation saw %v, want [0]", got)
	}
	if g.Seats[1].Life != StartingLife {
		t.Errorf("life = %d, want %d — the damage was prevented", g.Seats[1].Life, StartingLife)
	}
}

// TestDamageToPermanentContinuationRunsWithTheReplacedAmount — the same
// contract on the other target kind, so a card that fights and then
// says "gain that much life" gets the same answer a ping does.
func TestDamageToPermanentContinuationRunsWithTheReplacedAmount(t *testing.T) {
	g := newActiveGame(t)
	id := uuid.New()
	var got []int
	g.WithWriteLock(func() {
		pushTestCreature(g, id, g.Seats[1], 4, 9)
		g.RegisterReplacementForTest(doubleDamageTo(id))
		err := g.DealDamageToCreatureThenForEffect(uuid.Nil, id, 3,
			func(_ *Game, dealt int) error {
				got = append(got, dealt)
				return nil
			})
		if err != nil {
			t.Fatalf("DealDamageToCreatureThenForEffect: %v", err)
		}
	})
	if len(got) != 1 || got[0] != 6 {
		t.Fatalf("continuation saw %v, want [6]", got)
	}
	if c := findBattlefieldCard(g, id); c == nil || c.DamageMarked != 6 {
		t.Errorf("DamageMarked = %v, want 6", c)
	}
}

// --- (b) the batch form ----------------------------------------------

// TestDealDamageEachThenSumsTheDealtAmounts — the unpaused floor for
// the batch. One opponent is under a damage doubler and the other is
// not, so "the damage dealt this way" is 1+2 = 3 rather than 1×2.
func TestDealDamageEachThenSumsTheDealtAmounts(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	a, b := g.Seats[1].ID, g.Seats[2].ID
	var total []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(doubleDamageTo(b))
		err := g.DealDamageEachThenForEffect(uuid.Nil, []uuid.UUID{a, b}, 1,
			func(_ *Game, dealt int) error {
				total = append(total, dealt)
				return nil
			})
		if err != nil {
			t.Fatalf("DealDamageEachThenForEffect: %v", err)
		}
	})
	if len(total) != 1 || total[0] != 3 {
		t.Fatalf("total dealt = %v, want [3] — 1 to one opponent and 2 to the doubled one", total)
	}
	if g.Seats[1].Life != StartingLife-1 || g.Seats[2].Life != StartingLife-2 {
		t.Errorf("lives = %d/%d, want %d/%d",
			g.Seats[1].Life, g.Seats[2].Life, StartingLife-1, StartingLife-2)
	}
	if g.Seats[0].Life != StartingLife {
		t.Errorf("the dealer's own life moved to %d — the batch damages the listed targets only", g.Seats[0].Life)
	}
}

// TestDealDamageEachThenWaitsForAPausedLeg is the issue's engine-level
// repro: one leg pauses on a CR 616 prompt and the other does not. The
// total has to be the sum of what BOTH targets really took, which is
// the number the pre-#807 code could not produce — it read the paused
// opponent's life total back before the prompt was answered and counted
// zero.
func TestDealDamageEachThenWaitsForAPausedLeg(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	a, b := g.Seats[1].ID, g.Seats[2].ID
	var total []int
	g.WithWriteLock(func() {
		// Two different replacements on `a` only: that leg prompts.
		g.RegisterReplacementForTest(preventOneDamageTo(a))
		g.RegisterReplacementForTest(doubleDamageTo(a))
		err := g.DealDamageEachThenForEffect(uuid.Nil, []uuid.UUID{a, b}, 3,
			func(_ *Game, dealt int) error {
				total = append(total, dealt)
				return nil
			})
		if err != nil {
			t.Fatalf("DealDamageEachThenForEffect: %v", err)
		}
	})
	if len(total) != 0 {
		t.Fatalf("the batch reported %v before its paused leg landed", total)
	}
	if g.Seats[2].Life != StartingLife {
		t.Fatalf("the second opponent took damage at %d before the first leg's prompt was answered — "+
			"the batch is sequenced through the continuation", g.Seats[2].Life)
	}

	answerOnlyPrompt(t, g)

	// a: 3 → 2 → 4. b: 3, untouched. Total dealt = 7.
	if len(total) != 1 || total[0] != 7 {
		t.Fatalf("total dealt = %v, want [7] — 4 to the paused opponent plus 3 to the other", total)
	}
	if g.Seats[1].Life != StartingLife-4 {
		t.Errorf("paused opponent life = %d, want %d", g.Seats[1].Life, StartingLife-4)
	}
	if g.Seats[2].Life != StartingLife-3 {
		t.Errorf("unpaused opponent life = %d, want %d", g.Seats[2].Life, StartingLife-3)
	}
}

// TestDealDamageEachThenSkipsTargetsThatAreGone — a batch over a seat
// that has left, and over an instance ID that is on no zone at all,
// reports what the rest took rather than stalling or failing on them.
func TestDealDamageEachThenSkipsTargetsThatAreGone(t *testing.T) {
	g := newActiveGame(t)
	gone := uuid.New()
	var total []int
	g.WithWriteLock(func() {
		err := g.DealDamageEachThenForEffect(uuid.Nil, []uuid.UUID{gone, g.Seats[1].ID, gone}, 2,
			func(_ *Game, dealt int) error {
				total = append(total, dealt)
				return nil
			})
		if err != nil {
			t.Fatalf("DealDamageEachThenForEffect: %v", err)
		}
	})
	if len(total) != 1 || total[0] != 2 {
		t.Fatalf("total dealt = %v, want [2]", total)
	}
}

// TestDealDamageEachThenMixesPlayersAndPermanents — the batch routes
// each target as the kind of thing it is, the way the DealDamage
// primitive does, so "deals 2 damage to each opponent and each creature
// they control" is one call and one total.
func TestDealDamageEachThenMixesPlayersAndPermanents(t *testing.T) {
	g := newActiveGame(t)
	creature := uuid.New()
	var total []int
	g.WithWriteLock(func() {
		pushTestCreature(g, creature, g.Seats[1], 2, 9)
		err := g.DealDamageEachThenForEffect(uuid.Nil, []uuid.UUID{g.Seats[1].ID, creature}, 2,
			func(_ *Game, dealt int) error {
				total = append(total, dealt)
				return nil
			})
		if err != nil {
			t.Fatalf("DealDamageEachThenForEffect: %v", err)
		}
	})
	if len(total) != 1 || total[0] != 4 {
		t.Fatalf("total dealt = %v, want [4]", total)
	}
	if g.Seats[1].Life != StartingLife-2 {
		t.Errorf("player life = %d, want %d", g.Seats[1].Life, StartingLife-2)
	}
	if c := findBattlefieldCard(g, creature); c == nil || c.DamageMarked != 2 {
		t.Errorf("creature DamageMarked = %v, want 2", c)
	}
}

// TestFireAndForgetDamageStillWorks — the synchronous forms are the
// ones every other card in the catalog uses, and #807 must not have
// changed what they do. DealDamageToPlayerForEffect is now
// DealDamageToPlayerThenForEffect with a nil continuation; a nil
// continuation is a no-op, not a crash.
func TestFireAndForgetDamageStillWorks(t *testing.T) {
	g := newActiveGame(t)
	creature := uuid.New()
	g.WithWriteLock(func() {
		pushTestCreature(g, creature, g.Seats[1], 2, 9)
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, g.Seats[1].ID, 3); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, creature, 2); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
		// Zero damage is not an event and is not an error, with or
		// without a seat to aim it at (CR 120.8).
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, uuid.New(), 0); err != nil {
			t.Fatalf("zero damage at an absent player: %v", err)
		}
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, uuid.New(), 3); err != ErrPlayerNotFound {
			t.Fatalf("damage at an absent player = %v, want ErrPlayerNotFound", err)
		}
	})
	if g.Seats[1].Life != StartingLife-3 {
		t.Errorf("life = %d, want %d", g.Seats[1].Life, StartingLife-3)
	}
	if c := findBattlefieldCard(g, creature); c == nil || c.DamageMarked != 2 {
		t.Errorf("DamageMarked = %v, want 2", c)
	}
}

// TestUndoneDamageAnswerRunsTheContinuationAgain — the clone contract.
// runDamageTailLocked clears `then` on the event's damageTail, so an
// undo snapshot that SHARED that tail would come back with the
// continuation already consumed: the replay would land the damage and
// silently skip the rest of the card. cloneReplacementResume gives the
// snapshot its own copy; this is the test that says so.
func TestUndoneDamageAnswerRunsTheContinuationAgain(t *testing.T) {
	g := newActiveGame(t)
	var got []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(preventOneDamageTo(uuid.Nil))
		g.RegisterReplacementForTest(doubleDamageTo(uuid.Nil))
		err := g.DealDamageToPlayerThenForEffect(uuid.Nil, g.Seats[1].ID, 3,
			func(_ *Game, dealt int) error {
				got = append(got, dealt)
				return nil
			})
		if err != nil {
			t.Fatalf("DealDamageToPlayerThenForEffect: %v", err)
		}
	})

	// The snapshot an undo would restore, taken while the prompt is
	// still open.
	var snapshot *Game
	g.WithWriteLock(func() { snapshot = g.cloneLocked() })

	answerOnlyPrompt(t, g)
	if len(got) != 1 || got[0] != 4 {
		t.Fatalf("continuation saw %v on the first answer, want [4]", got)
	}

	// Undo: answer the restored game's prompt instead. It must run the
	// continuation again, with the same amount.
	got = nil
	answerOnlyPrompt(t, snapshot)
	if len(got) != 1 || got[0] != 4 {
		t.Fatalf("continuation saw %v after the undo, want [4] — the snapshot shared a consumed tail", got)
	}
	if snapshot.Seats[1].Life != StartingLife-4 {
		t.Errorf("restored life = %d, want %d", snapshot.Seats[1].Life, StartingLife-4)
	}
}
