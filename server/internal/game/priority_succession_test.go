package game

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// priority_succession_test.go — #2275, CR 117.4.
//
// "If all players pass in succession (that is, if all players pass
// without taking any actions in between passing), the spell or ability
// on top of the stack resolves or, if the stack is empty, the phase or
// step ends."
//
// The engine used to read "priority came back round to the ACTIVE
// seat" as "everyone passed". That is only the same thing when the
// round began with the active player; a non-active caster keeps
// priority (CR 117.3c), and their spell resolved on their own pass. In
// prod game 497de7d2 a Stroke of Genius cast in its opponent's draw
// step resolved through a Counterspell and a Mana Drain. These tests
// pin the succession from every side: who gets priority next, when the
// top resolves, when the step ends, and what restarts the count.

// successionGameAt returns a started game with `seats` seats whose
// cursor stands in `active`'s `step`, with the active player holding
// priority and nothing on the stack.
func successionGameAt(t *testing.T, seats, active int, step Step) *Game {
	t.Helper()
	g := newActiveGameWithSeats(t, seats)
	for range seats {
		if g.Turn.ActiveSeat == active {
			break
		}
		if err := g.EndTurnNowForTest(); err != nil {
			t.Fatalf("PassTurn: %v", err)
		}
	}
	if g.Turn.ActiveSeat != active {
		t.Fatalf("setup: active seat %d, want %d", g.Turn.ActiveSeat, active)
	}
	advanceTo(t, g, step)
	if g.Turn.PriorityHolder != active || g.stackHasItemsLocked() {
		t.Fatalf("setup: holder %d with stack %v, want %d with an empty stack",
			g.Turn.PriorityHolder, g.stackHasItemsLocked(), active)
	}
	return g
}

// castFreeInstant puts a no-cost instant in `seat`'s hand and casts it.
// The seat must hold priority.
func castFreeInstant(t *testing.T, g *Game, seat int, name string) uuid.UUID {
	t.Helper()
	if g.Turn.PriorityHolder != seat {
		t.Fatalf("seat %d to cast %s, but seat %d holds priority", seat, name, g.Turn.PriorityHolder)
	}
	p := g.Seats[seat]
	id := handCardFor(p, name, "Instant")
	if err := g.CastSpell(p.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("seat %d casts %s: %v", seat, name, err)
	}
	if g.Turn.PriorityHolder != seat {
		t.Fatalf("CR 117.3c: seat %d cast %s and should keep priority; seat %d holds it", seat, name, g.Turn.PriorityHolder)
	}
	return id
}

// passAs passes priority for `seat`, failing if somebody else holds it.
func passAs(t *testing.T, g *Game, seat int) {
	t.Helper()
	if g.Turn.PriorityHolder != seat {
		t.Fatalf("seat %d to pass, but seat %d holds priority (passed so far: %v)",
			seat, g.Turn.PriorityHolder, g.Turn.PassedInSuccession.Seats())
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("seat %d passes: %v", seat, err)
	}
}

// onStack reports whether a spell or ability with this id is on the
// stack.
func onStack(g *Game, id uuid.UUID) bool {
	_, ok := g.StackMeta[id]
	return ok
}

// The issue, exactly: two seats, the active player passes in their own
// draw step, the other casts an instant and passes. The active player
// must receive priority with the spell on the stack, and the spell
// resolves only once they pass.
func TestNonActiveCastGivesTheActivePlayerPriorityBeforeItResolves(t *testing.T) {
	g := successionGameAt(t, 2, 1, StepDraw)

	passAs(t, g, 1) // the active player passes: correct, priority to seat 0
	stroke := castFreeInstant(t, g, 0, "Stroke of Genius")
	passAs(t, g, 0) // the caster passes

	if !onStack(g, stroke) {
		t.Fatal("the spell resolved on its caster's pass; the active player never held priority over it (CR 117.4)")
	}
	if g.Turn.PriorityHolder != 1 {
		t.Fatalf("after the caster's pass seat %d holds priority, want the active seat 1", g.Turn.PriorityHolder)
	}
	if g.Turn.Step != StepDraw {
		t.Fatalf("the step moved to %s with a spell on the stack", g.Turn.Step)
	}

	passAs(t, g, 1)
	if onStack(g, stroke) {
		t.Fatal("both seats have passed in succession and the spell is still on the stack")
	}
	if g.Turn.PriorityHolder != 1 || g.Turn.Step != StepDraw {
		t.Errorf("after the resolution: holder %d in %s, want the active seat 1 in draw (CR 117.3b)",
			g.Turn.PriorityHolder, g.Turn.Step)
	}
}

// Four seats, every active seat, every caster: the round restarts from
// the caster and visits every other seat in turn order — the active
// player and the seats between them included — before the spell
// resolves on the pass of the seat just before the caster.
func TestSuccessionRestartsFromTheCasterAtFourSeats(t *testing.T) {
	const seats = 4
	for active := range seats {
		for offset := range seats {
			caster := (active + offset) % seats
			t.Run(fmt.Sprintf("active %d caster %d", active, caster), func(t *testing.T) {
				g := successionGameAt(t, seats, active, StepUpkeep)
				for i := range offset {
					passAs(t, g, (active+i)%seats)
				}
				spell := castFreeInstant(t, g, caster, "Test Instant")
				passAs(t, g, caster)
				for i := 1; i < seats; i++ {
					seat := (caster + i) % seats
					if !onStack(g, spell) {
						t.Fatalf("the spell resolved before seat %d passed over it", seat)
					}
					passAs(t, g, seat)
				}
				if onStack(g, spell) {
					t.Fatal("every seat passed in succession and the spell is still on the stack")
				}
				if g.Turn.PriorityHolder != active || g.Turn.Step != StepUpkeep {
					t.Errorf("after the resolution: holder %d in %s, want the active seat %d in upkeep",
						g.Turn.PriorityHolder, g.Turn.Step, active)
				}
				if g.Turn.PassedInSuccession != 0 {
					t.Errorf("a resolution begins a new succession; still recorded: %v", g.Turn.PassedInSuccession.Seats())
				}
			})
		}
	}
}

// A response chain: only the top object resolves when everyone has
// passed, then the ACTIVE player receives priority (CR 117.3b) and the
// succession starts again from them — the spell underneath needs a
// full round of its own.
func TestResponseResolvesOnlyTheTopThenTheActivePlayerStartsANewRound(t *testing.T) {
	g := successionGameAt(t, 4, 0, StepUpkeep)

	first := castFreeInstant(t, g, 0, "First")
	passAs(t, g, 0)
	response := castFreeInstant(t, g, 1, "Response")
	passAs(t, g, 1)
	passAs(t, g, 2)
	passAs(t, g, 3)
	if !onStack(g, response) {
		t.Fatal("the response resolved before the active player passed over it")
	}
	passAs(t, g, 0)
	if onStack(g, response) {
		t.Fatal("every seat passed over the response and it did not resolve")
	}
	if !onStack(g, first) {
		t.Fatal("the spell underneath resolved on the same passes as the response (CR 117.4 resolves only the top)")
	}
	if g.Turn.PriorityHolder != 0 || g.Turn.PassedInSuccession != 0 {
		t.Fatalf("after the response resolved: holder %d with passes %v, want the active seat 0 and a new round",
			g.Turn.PriorityHolder, g.Turn.PassedInSuccession.Seats())
	}
	passAs(t, g, 0)
	passAs(t, g, 1)
	passAs(t, g, 2)
	if !onStack(g, first) {
		t.Fatal("the first spell resolved before seat 3 passed over it in the new round")
	}
	passAs(t, g, 3)
	if onStack(g, first) {
		t.Fatal("every seat passed over the first spell and it did not resolve")
	}
}

// An empty stack ends the step only when every seat has passed in
// succession. A mana ability activated by the priority holder is an
// action (CR 117.3c), so the seats that passed before it pass again.
func TestEmptyStackStepEndsOnlyAfterEverySeatPassesInSuccession(t *testing.T) {
	g := successionGameAt(t, 4, 0, StepUpkeep)
	forest := pushForest(g, g.Seats[2])

	passAs(t, g, 0)
	passAs(t, g, 1)
	if err := g.ActivateManaAbility(g.Seats[2].ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("seat 2 taps its Forest: %v", err)
	}
	passAs(t, g, 2)
	passAs(t, g, 3)
	if g.Turn.Step != StepUpkeep {
		t.Fatalf("the step ended on seat 3's pass, but seats 0 and 1 passed before seat 2 acted")
	}
	passAs(t, g, 0)
	if g.Turn.Step != StepUpkeep {
		t.Fatal("the step ended before seat 1 passed again")
	}
	passAs(t, g, 1)
	if g.Turn.Step != StepDraw {
		t.Fatalf("every seat has passed in succession with an empty stack; still in %s", g.Turn.Step)
	}
	if g.Turn.PriorityHolder != 0 || g.Turn.PassedInSuccession != 0 {
		t.Errorf("the new step: holder %d with passes %v, want the active seat 0 and none",
			g.Turn.PriorityHolder, g.Turn.PassedInSuccession.Seats())
	}
}

// Mana tapped by a player who does NOT hold priority — to answer a
// "pay {1}" prompt, say — is not an action in the CR 117.4 sense, and
// the passes before it stand.
func TestManaTappedWithoutPriorityLeavesTheSuccessionAlone(t *testing.T) {
	g := successionGameAt(t, 4, 0, StepUpkeep)
	forest := pushForest(g, g.Seats[3])

	passAs(t, g, 0)
	passAs(t, g, 1)
	if err := g.ActivateManaAbility(g.Seats[3].ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("seat 3 taps its Forest: %v", err)
	}
	passAs(t, g, 2)
	passAs(t, g, 3)
	if g.Turn.Step != StepDraw {
		t.Fatalf("seat 3 tapped mana without priority, which restarted the round; still in %s", g.Turn.Step)
	}
}

// A land play is a special action (CR 116.2a) and so an action: the
// active player who gets priority again in their own main phase and
// plays a land keeps priority, and the seat that passed before it
// passes again.
func TestLandPlayRestartsTheSuccession(t *testing.T) {
	g := successionGameAt(t, 2, 0, StepPrecombatMain)
	forest := pushForest(g, g.Seats[1])

	passAs(t, g, 0)
	if err := g.ActivateManaAbility(g.Seats[1].ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("seat 1 taps its Forest: %v", err)
	}
	passAs(t, g, 1)
	if g.Turn.Step != StepPrecombatMain || g.Turn.PriorityHolder != 0 {
		t.Fatalf("seat 0 passed before seat 1 acted, so it gets priority again: %s, holder %d", g.Turn.Step, g.Turn.PriorityHolder)
	}
	land := handCardFor(g.Seats[0], "Mountain", "Basic Land — Mountain")
	if err := g.CastSpell(g.Seats[0].ID, land, CastSpellParams{}); err != nil {
		t.Fatalf("seat 0 plays a land: %v", err)
	}
	passAs(t, g, 0)
	if g.Turn.Step != StepPrecombatMain {
		t.Fatal("the step ended on the land player's pass; seat 1 passed before the land was played")
	}
	passAs(t, g, 1)
	if g.Turn.Step == StepPrecombatMain {
		t.Fatal("both seats passed in succession and the main phase did not end")
	}
}

// A special action that uses no stack (CR 116.2) is an action all the
// same: an instant suspended in the active player's upkeep restarts the
// round, so the active player — who passed first — gets priority again
// before the step ends.
func TestSpecialActionRestartsTheSuccession(t *testing.T) {
	g := successionGameAt(t, 2, 0, StepUpkeep)
	withSuspendCard(t, 2, "")

	passAs(t, g, 0)
	suspendIt(t, g, g.Seats[1], "Suspended Instant", "Instant", "{1}{U}")
	passAs(t, g, 1)
	if g.Turn.Step != StepUpkeep || g.Turn.PriorityHolder != 0 {
		t.Fatalf("the active player passed before the suspend and must get priority again: %s, holder %d",
			g.Turn.Step, g.Turn.PriorityHolder)
	}
	passAs(t, g, 0)
	if g.Turn.Step != StepDraw {
		t.Fatalf("both seats passed in succession after the suspend; still in %s", g.Turn.Step)
	}
}

// A trigger put on the stack between passes is a new object on top,
// and the seats that passed before it existed pass again before it
// resolves.
func TestTriggerBetweenPassesRestartsTheSuccession(t *testing.T) {
	g := successionGameAt(t, 2, 0, StepUpkeep)
	spell := castFreeInstant(t, g, 0, "Test Instant")
	passAs(t, g, 0)

	resolved := false
	var trigger uuid.UUID
	g.WithWriteLock(func() {
		source := NewCard("Trigger Source", g.Seats[1].ID)
		source.TypeLine = "Enchantment"
		g.Battlefield.PushTop(source)
		item := newTriggeredItemForTest(&source, "Test trigger", func(*Game, *StackItem) error {
			resolved = true
			return nil
		})
		trigger = item.ID
		g.PendingTriggers = append(g.PendingTriggers, item)
		g.runStateChecksLocked()
	})
	if !onStack(g, trigger) {
		t.Fatal("setup: the trigger is not on the stack")
	}

	passAs(t, g, 1)
	if resolved {
		t.Fatal("the trigger resolved on one pass; seat 0 passed before it was on the stack")
	}
	if g.Turn.PriorityHolder != 0 {
		t.Fatalf("seat %d holds priority, want seat 0 with the trigger on top", g.Turn.PriorityHolder)
	}
	passAs(t, g, 0)
	if !resolved || onStack(g, trigger) {
		t.Fatal("both seats passed over the trigger and it did not resolve")
	}
	if !onStack(g, spell) {
		t.Fatal("the spell underneath resolved with the trigger")
	}
	passAs(t, g, 0)
	passAs(t, g, 1)
	if onStack(g, spell) {
		t.Fatal("both seats passed over the spell and it did not resolve")
	}
}

// A seat that has left the game is not waited for (CR 800.4a), and a
// departure restarts the round for the seats still playing.
func TestConcededSeatIsNotWaitedFor(t *testing.T) {
	t.Run("a seat not holding priority", func(t *testing.T) {
		g := successionGameAt(t, 4, 0, StepUpkeep)
		spell := castFreeInstant(t, g, 0, "Test Instant")
		passAs(t, g, 0)
		passAs(t, g, 1)
		if err := g.Concede(g.Seats[3].ID); err != nil {
			t.Fatal(err)
		}
		passAs(t, g, 2)
		passAs(t, g, 0)
		if !onStack(g, spell) {
			t.Fatal("the spell resolved before seat 1 passed again after the departure")
		}
		passAs(t, g, 1)
		if onStack(g, spell) {
			t.Fatal("every seat still in the game has passed and the spell is still on the stack")
		}
	})
	t.Run("the seat holding priority", func(t *testing.T) {
		g := successionGameAt(t, 4, 0, StepUpkeep)
		spell := castFreeInstant(t, g, 0, "Test Instant")
		passAs(t, g, 0)
		passAs(t, g, 1)
		if err := g.Concede(g.Seats[2].ID); err != nil {
			t.Fatal(err)
		}
		passAs(t, g, 0)
		passAs(t, g, 1)
		if !onStack(g, spell) {
			t.Fatal("the spell resolved before seat 3 passed over it")
		}
		passAs(t, g, 3)
		if onStack(g, spell) {
			t.Fatal("every seat still in the game has passed and the spell is still on the stack")
		}
	})
}

// The post-block window (#1501): the last declaration hands the active
// player priority with a fresh succession, and a defender's instant cast
// in that window does not resolve until the active player passes over
// it.
func TestDefenderInstantAfterBlocksGivesTheActivePlayerPriority(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if g.Turn.PriorityHolder != 0 || g.Turn.PassedInSuccession != 0 {
		t.Fatalf("after the declaration: holder %d with passes %v, want the active seat and none",
			g.Turn.PriorityHolder, g.Turn.PassedInSuccession.Seats())
	}
	passAs(t, g, 0)
	trick := castFreeInstant(t, g, 1, "Combat Trick")
	passAs(t, g, 1)
	if !onStack(g, trick) || g.Turn.PriorityHolder != 0 {
		t.Fatalf("the defender's instant resolved without the active player holding priority over it (holder %d)", g.Turn.PriorityHolder)
	}
	passAs(t, g, 0)
	if onStack(g, trick) || g.Turn.Step != StepDeclareBlockers {
		t.Fatalf("after both passes: trick on stack %v in %s, want resolved in declare_blockers", onStack(g, trick), g.Turn.Step)
	}
	passAs(t, g, 0)
	passAs(t, g, 1)
	if g.Turn.Step == StepDeclareBlockers {
		t.Fatal("both seats passed with an empty stack and declare_blockers did not end")
	}
}

// Undo restores a partial succession exactly: the clone the room keeps
// as an undo point carries who has passed, so the seats that had not
// passed yet still must.
func TestUndoRestoresAPartialSuccession(t *testing.T) {
	g := successionGameAt(t, 4, 0, StepUpkeep)
	spell := castFreeInstant(t, g, 0, "Test Instant")
	passAs(t, g, 0)
	passAs(t, g, 1)
	point := g.Clone()

	passAs(t, g, 2)
	passAs(t, g, 3)
	if onStack(g, spell) {
		t.Fatal("setup: the spell should have resolved")
	}

	g.RestoreFrom(point)
	if g.Turn.PriorityHolder != 2 || g.Turn.PassedInSuccession != SeatSet(0).With(0).With(1) {
		t.Fatalf("restored: holder %d with passes %v, want seat 2 with seats 0 and 1 passed",
			g.Turn.PriorityHolder, g.Turn.PassedInSuccession.Seats())
	}
	passAs(t, g, 2)
	if !onStack(g, spell) {
		t.Fatal("after the undo the spell resolved before seat 3 passed")
	}
	passAs(t, g, 3)
	if onStack(g, spell) {
		t.Fatal("after the undo every seat passed and the spell did not resolve")
	}
}

// The succession survives a snapshot round trip, and a restore point
// written before the field reads as "nobody has passed yet" — every
// seat gets one more chance to act, never one fewer.
func TestSnapshotCarriesThePartialSuccession(t *testing.T) {
	g := successionGameAt(t, 4, 0, StepUpkeep)
	spell := castFreeInstant(t, g, 0, "Test Instant")
	passAs(t, g, 0)
	passAs(t, g, 1)

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	restore := func(t *testing.T, raw []byte) *Game {
		t.Helper()
		var snap GameSnapshot
		if err := json.Unmarshal(raw, &snap); err != nil {
			t.Fatal(err)
		}
		out, err := snap.Restore()
		if err != nil {
			t.Fatal(err)
		}
		return out
	}

	t.Run("round trip", func(t *testing.T) {
		r := restore(t, raw)
		if r.Turn.PassedInSuccession != g.Turn.PassedInSuccession {
			t.Fatalf("restored passes %v, want %v", r.Turn.PassedInSuccession.Seats(), g.Turn.PassedInSuccession.Seats())
		}
		passAs(t, r, 2)
		passAs(t, r, 3)
		if onStack(r, spell) {
			t.Fatal("every seat passed across the restore and the spell did not resolve")
		}
	})
	t.Run("a restore point without the field", func(t *testing.T) {
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		turn, ok := doc["turn"].(map[string]any)
		if !ok {
			t.Fatalf("snapshot has no turn object: %v", doc["turn"])
		}
		if _, ok := turn["PassedInSuccession"]; !ok {
			t.Fatal("the snapshot does not carry PassedInSuccession under turn; the test's premise is wrong")
		}
		delete(turn, "PassedInSuccession")
		old, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		r := restore(t, old)
		passAs(t, r, 2)
		passAs(t, r, 3)
		if !onStack(r, spell) {
			t.Fatal("an older restore point lost seats 0 and 1's chance to pass again")
		}
		passAs(t, r, 0)
		passAs(t, r, 1)
		if onStack(r, spell) {
			t.Fatal("every seat passed in succession and the spell did not resolve")
		}
	})
}
