package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// exhaust_mana_test.go — #1183, the MANA half of the exhaust keyword.
//
// #1181 built the record and deliberately left this out: a mana
// ability takes the other entry point (ActivateManaAbility, CR 605.3a
// — no stack, no priority, no announcement to hang a record on), so
// `ManaAbility` carried no Exhaust marker and the combination was
// unspellable. Loot, the Pathfinder prints "Exhaust — {G}, {T}: Add
// three mana of any one color" beside two ordinary exhaust abilities.
//
// What is pinned here is what the mana half has that the activated
// half does not: TWO ways to spend the ability — the hand click and
// the AUTO-TAPPER's executor — and three readers that must all refuse
// it once it is spent (the activation path, the planner, and CR
// 106.7's "could produce"). Everything the two halves share — the
// per-object key, CR 400.7, the snapshot — is asserted once more here
// because the mana path writes its own record and could have got the
// key wrong on its own.

const (
	exhaustManaLabel = "Exhaust — {T}: Add {G}{G}"
	plainManaLabel   = "{T}: Add {C}"
)

// exhaustManaCatalog installs a mana-ability catalog entry for
// `oracle`. A catalog hook rather than Card.ManaAbilities for the
// reason exhaustCatalog gives: an instance ability list is closures,
// which no restore can bring back, so the round-trip test below would
// otherwise pass for the wrong reason.
func exhaustManaCatalog(t *testing.T, oracle string, abilities ...ManaAbilityShape) {
	t.Helper()
	withCatalogHook(t, func(key string) []ManaAbilityShape {
		if key == oracle {
			return abilities
		}
		return nil
	})
}

// exhaustManaAbility is the shape Loot's first ability has, shrunk to
// two fixed green mana so a test can count the pool.
func exhaustManaAbility() ManaAbilityShape {
	return ManaAbilityShape{
		TapCost:  true,
		Produced: "{G}{G}",
		Label:    exhaustManaLabel,
		Exhaust:  true,
	}
}

// plainManaAbility is the repeatable ability that must NOT be caught
// by the exhaust gate — the per-ability half of the record, on a mana
// source.
func plainManaAbility() ManaAbilityShape {
	return ManaAbilityShape{TapCost: true, Produced: "{C}", Label: plainManaLabel}
}

// pushManaExhaustSource puts an untapped artifact keyed on `oracle` on
// the battlefield. An artifact rather than a creature so CR 302.6
// never enters into it.
func pushManaExhaustSource(g *Game, owner *Player, oracle string) uuid.UUID {
	c := NewCard("Exhaust Mana Probe", owner.ID)
	c.TypeLine = "Artifact"
	c.OracleID = oracle
	c.Controller = owner.ID
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// untapForTest puts a permanent back untapped without a turn
// boundary, so a test can prove the ability is refused for being SPENT
// rather than for being tapped.
func untapForTest(g *Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
}

func tappedForTest(g *Game, id uuid.UUID) bool {
	var tapped bool
	g.WithWriteLock(func() {
		if c := g.findCardByIDLocked(id); c != nil {
			tapped = c.Tapped
		}
	})
	return tapped
}

func producibleForTest(g *Game, id uuid.UUID) []string {
	var out []string
	g.WithWriteLock(func() {
		if c := g.findCardByIDLocked(id); c != nil {
			out = g.ProducibleManaLocked(*c)
		}
	})
	return out
}

// TestAnExhaustManaAbilityTapsOnceByHand is the headline, and the
// second half is the one that matters: the refusal taps nothing.
func TestAnExhaustManaAbilityTapsOnceByHand(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustManaCatalog(t, "probe-mana-once", exhaustManaAbility())
	src := pushManaExhaustSource(g, me, "probe-mana-once")

	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Fatalf("pool = %v, want {G}{G}", me.ManaPool)
	}
	if got := exhaustedThisGame(g, src, exhaustManaLabel); got != 1 {
		t.Fatalf("the record reads %d activations, want 1 — ActivateManaAbility did not write it", got)
	}

	// Untapped again, so the refusal below is about the record and not
	// about the tap cost.
	untapForTest(g, src)
	me.ManaPool = nil
	err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{})
	if !errors.Is(err, ErrAbilityExhausted) {
		t.Fatalf("second activation: err = %v, want ErrAbilityExhausted", err)
	}
	if tappedForTest(g, src) {
		t.Errorf("a refused activation tapped the source — the gate runs before any cost is paid")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("a refused activation added %v to the pool", me.ManaPool)
	}
	if got := exhaustedThisGame(g, src, exhaustManaLabel); got != 1 {
		t.Errorf("the refusal wrote a second record: %d", got)
	}
}

// The per-ABILITY half on a mana source: spending the exhaust ability
// must not lock the repeatable one beside it, and the auto-tapper must
// still plan that one.
func TestAnExhaustManaAbilityDoesNotLockTheOtherManaAbility(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustManaCatalog(t, "probe-mana-mixed", exhaustManaAbility(), plainManaAbility())
	src := pushManaExhaustSource(g, me, "probe-mana-mixed")

	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("the exhaust ability: %v", err)
	}
	untapForTest(g, src)
	me.ManaPool = nil

	// Twice, because "repeatable" is the whole assertion.
	for i := 0; i < 2; i++ {
		if err := g.ActivateManaAbility(me.ID, src, 1, ManaAbilityParams{}); err != nil {
			t.Fatalf("the plain ability, activation %d: %v", i+1, err)
		}
		untapForTest(g, src)
	}
	if got := exhaustedThisGame(g, src, plainManaLabel); got != 2 {
		t.Errorf("the plain ability's record reads %d, want 2", got)
	}
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); !errors.Is(err, ErrAbilityExhausted) {
		t.Errorf("the exhaust ability is not still spent: err = %v", err)
	}

	// And the picker the planner and the executor share skips the
	// spent ability and lands on the one behind it, which is what a
	// per-SOURCE check outside the picker would have got wrong.
	g.mu.Lock()
	picked := g.autoTapAbilityFor(src, ManaAbilitiesForCard(*g.findCardByIDLocked(src)))
	g.mu.Unlock()
	if picked == nil || picked.Label != plainManaLabel {
		t.Errorf("autoTapAbilityFor picked %+v, want the repeatable ability behind the spent one", picked)
	}
}

// The auto-tapper's half, and the reason #1183 is its own issue: the
// planner may spend an exhaust ability while paying for something
// else, so the executor has to write the record — and neither half may
// touch the ability again afterwards.
func TestTheAutoTapperSpendsAnExhaustManaAbilityExactlyOnce(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustManaCatalog(t, "probe-mana-plan", exhaustManaAbility())
	src := pushManaExhaustSource(g, me, "probe-mana-plan")

	// Non-vacuity: the source really is plannable before it is spent.
	cost := costFor(t, "{G}{G}")
	g.WithWriteLock(func() {
		plan, ok := g.autoTapLocked(me.ID, cost, 0, nil)
		if !ok || len(plan) != 1 || plan[0].CardID != src {
			t.Fatalf("plan = %+v (ok=%v), want the exhaust source", plan, ok)
		}
		g.materializePlanLocked(me, plan, cost)
	})
	if len(me.ManaPool) != 2 {
		t.Fatalf("pool = %v, want the planned {G}{G}", me.ManaPool)
	}
	if got := exhaustedThisGame(g, src, exhaustManaLabel); got != 1 {
		t.Fatalf("the executor did not write the record: %d activations, want 1 — a plan that "+
			"spent an exhaust ability without recording it hands the ability back", got)
	}

	// Untapped again, and now it is not a mana source at all: not
	// planned, not enumerated, and not producible.
	untapForTest(g, src)
	me.ManaPool = nil
	if _, ok := g.AutoTapForCost(me.ID, costFor(t, "{G}"), 0); ok {
		t.Errorf("the planner still counts a spent exhaust ability as a {G} source")
	}
	if got := producibleForTest(g, src); len(got) != 0 {
		t.Errorf("ProducibleManaLocked = %v for a spent exhaust ability, want nothing — a cast "+
			"priced on it could not be paid (CR 106.7, #1183)", got)
	}
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); !errors.Is(err, ErrAbilityExhausted) {
		t.Errorf("after the plan spent it: err = %v, want ErrAbilityExhausted", err)
	}
	if tappedForTest(g, src) {
		t.Errorf("the refused hand activation tapped the source")
	}
}

// The EXECUTOR's own refusal, asked directly. A plan can arrive stale
// — built before the ability was spent in response — and
// materializePlanLocked re-asks every gate before it taps, exactly as
// it re-asks the condition and the counter cost.
func TestAStalePlanDoesNotTapASpentExhaustManaAbility(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustManaCatalog(t, "probe-mana-stale", exhaustManaAbility())
	src := pushManaExhaustSource(g, me, "probe-mana-stale")

	cost := costFor(t, "{G}{G}")
	var stale tapPlan
	g.WithWriteLock(func() {
		var ok bool
		stale, ok = g.autoTapLocked(me.ID, cost, 0, nil)
		if !ok {
			t.Fatalf("no plan for {G}{G}")
		}
	})

	// Spent by hand in between, and untapped again so the executor's
	// own "is it tapped" guard cannot be what refuses it.
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("hand activation: %v", err)
	}
	untapForTest(g, src)
	me.ManaPool = nil

	g.WithWriteLock(func() { g.materializePlanLocked(me, stale, cost) })
	if tappedForTest(g, src) {
		t.Errorf("the stale plan tapped a spent exhaust ability")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the stale plan minted %v off a spent exhaust ability", me.ManaPool)
	}
	if got := exhaustedThisGame(g, src, exhaustManaLabel); got != 1 {
		t.Errorf("the record reads %d, want 1 — the executor wrote a second use", got)
	}
}

// CR 400.7, on the mana path: a flicker is a new object and gets its
// exhaust ability back; a permanent that never left keeps it spent.
// The record is keyed by (object, label) and nothing here is coded —
// this asserts the mana path builds the same key the CR 602 path does.
func TestAFlickerRefreshesAnExhaustManaAbility(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustManaCatalog(t, "probe-mana-flicker", exhaustManaAbility())
	src := pushManaExhaustSource(g, me, "probe-mana-flicker")

	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	me.ManaPool = nil
	passBothForTest(g)

	epochBefore := objectEpochOf(g, src)
	bounceAndReturn(t, g, src, me.ID)
	if objectEpochOf(g, src) == epochBefore {
		t.Fatal("setup: leaving and returning did not make a new object")
	}
	if got := exhaustedThisGame(g, src, exhaustManaLabel); got != 0 {
		t.Fatalf("the returning object reads %d activations, want 0 (CR 400.7)", got)
	}
	if got := producibleForTest(g, src); len(got) == 0 {
		t.Errorf("the returning object could produce nothing; its exhaust ability is available again")
	}
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Errorf("the returning object cannot use its exhaust mana ability: %v", err)
	}
}

// "Have I used this yet" is game state a player can lose a game over,
// so it rewinds with Clone / RestoreFrom and survives a snapshot round
// trip. Asserted again on the mana path because this path writes its
// own record — a write to a map the snapshot did not carry would fail
// here and nowhere else.
func TestAnExhaustManaAbilitySurvivesCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustManaCatalog(t, "probe-mana-snapshot", exhaustManaAbility())
	src := pushManaExhaustSource(g, me, "probe-mana-snapshot")

	before := g.Clone()
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	me.ManaPool = nil
	if got := exhaustedThisGame(g, src, exhaustManaLabel); got != 1 {
		t.Fatalf("before the undo: %d, want 1", got)
	}

	// The snapshot round trip first, off the live (spent) game.
	passBothForTest(g)
	_, restored := roundTrip(t, g)
	if got := exhaustedThisGame(restored, src, exhaustManaLabel); got != 1 {
		t.Errorf("the restored game reads %d activations, want 1", got)
	}
	restoredMe := restored.Seats[0]
	if err := restored.ActivateManaAbility(restoredMe.ID, src, 0, ManaAbilityParams{}); !errors.Is(err, ErrAbilityExhausted) {
		t.Errorf("the restored game: err = %v, want ErrAbilityExhausted", err)
	}

	// And the undo, which must put the ability back.
	g.RestoreFrom(before)
	if got := exhaustedThisGame(g, src, exhaustManaLabel); got != 0 {
		t.Errorf("after undoing the activation: %d, want 0", got)
	}
	if got := exhaustedThisGame(before, src, exhaustManaLabel); got != 0 {
		t.Errorf("the snapshot itself reads %d, want 0 — the maps must be copied, not shared", got)
	}
}

// A NON-exhaust mana ability is untouched by all of it: every source
// on every board still taps as often as it untaps, and CR 106.7 still
// answers for it. The guard against a gate that accidentally applies
// to the ~40 catalog mana abilities that print no keyword.
func TestAPlainManaAbilityIsNeverExhausted(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustManaCatalog(t, "probe-mana-plain", plainManaAbility())
	src := pushManaExhaustSource(g, me, "probe-mana-plain")

	for i := 0; i < 3; i++ {
		if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		untapForTest(g, src)
	}
	if len(me.ManaPool) != 3 {
		t.Errorf("pool = %v, want three {C}", me.ManaPool)
	}
	if got := exhaustedThisGame(g, src, plainManaLabel); got != 3 {
		t.Errorf("the record reads %d, want 3 — every activation is counted, exhaust or not", got)
	}
	if got := producibleForTest(g, src); len(got) != 1 || got[0] != "C" {
		t.Errorf("ProducibleManaLocked = %v, want [C] — a plain ability is never spent", got)
	}
}
