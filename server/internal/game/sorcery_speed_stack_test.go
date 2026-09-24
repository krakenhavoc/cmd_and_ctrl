package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// sorcery_speed_stack_test.go — #1352, CR 307.1 / CR 117.1a / CR
// 405.1: sorcery timing needs an EMPTY stack, and an activated or
// triggered ability on the stack makes it not empty exactly as a
// spell does. Abilities have no card in g.Stack — they live only in
// g.StackMeta — so a gate that looked at the stack zone alone let a
// sorcery, an equip and a plot through over a pending upkeep trigger.
//
// Every test here puts a TRIGGER on the stack (not a spell: the
// spell case was always refused and is pinned elsewhere) and asks
// one caller of SorcerySpeedOpenLocked for its answer. The
// enumerator's half is internal/legal/sorcery_speed_stack_test.go.

// stackTriggerForTest puts one triggered ability controlled by
// `owner` on the stack, through the same APNAP drain a real trigger
// takes, and returns its id. Nothing about it is a card: g.Stack
// stays empty, which is the whole point.
func stackTriggerForTest(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	item := &StackItem{
		ID:           uuid.New(),
		Kind:         StackItemTriggered,
		Controller:   owner.ID,
		Owner:        owner.ID,
		SourceCardID: uuid.New(),
		Label:        "Test upkeep trigger — nothing happens",
		Effect:       func(*Game, *StackItem) error { return nil },
	}
	g.WithWriteLock(func() {
		g.PendingTriggers = append(g.PendingTriggers, item)
		g.drainPendingTriggersAPNAPLocked()
	})
	if g.StackMeta[item.ID] == nil {
		t.Fatal("setup: the trigger did not reach the stack")
	}
	if g.Stack != nil && len(g.Stack.Cards) != 0 {
		t.Fatalf("setup: %d cards in the stack zone, want none — the trigger is the only object", len(g.Stack.Cards))
	}
	return item.ID
}

// resolveWholeStackForTest passes priority until nothing is on the
// stack, or gives up.
func resolveWholeStackForTest(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		var busy bool
		g.ReadSnapshot(func() { busy = g.stackHasItemsLocked() })
		if !busy {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

func sorceryWindowOpen(g *Game, p *Player) bool {
	var out bool
	g.ReadSnapshot(func() { out = g.SorcerySpeedOpenLocked(p.ID) })
	return out
}

// The gate itself: a trigger shuts it, and it reopens once the
// trigger has resolved. Whose trigger it is does not matter.
func TestSorceryWindowIsShutByATriggerOnTheStack(t *testing.T) {
	for _, whose := range []int{0, 1} {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		me := g.Seats[0]
		if !sorceryWindowOpen(g, me) {
			t.Fatal("setup: the active player's main phase with an empty stack is not open")
		}
		stackTriggerForTest(t, g, g.Seats[whose])
		if sorceryWindowOpen(g, me) {
			t.Errorf("seat %d's trigger on the stack: the sorcery window is open, want shut (CR 307.1)", whose)
		}
		resolveWholeStackForTest(t, g)
		if !sorceryWindowOpen(g, me) {
			t.Errorf("seat %d's trigger resolved: the sorcery window is still shut", whose)
		}
	}
}

// CastSpell through CastTimingOpenLocked: a sorcery and a creature
// are refused over a trigger, an instant is not, and a land play —
// CR 305.1's own "when the stack is empty" — is refused too.
func TestATriggerOnTheStackBlocksASorcery(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	ritual := timingHandCard(me, "Test Ritual", "Sorcery")
	bear := timingHandCard(me, "Test Bear", "Creature — Bear")
	bolt := timingHandCard(me, "Test Bolt", "Instant")
	land := timingHandCard(me, "Test Forest", "Basic Land — Forest")
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})

	stackTriggerForTest(t, g, g.Seats[1])

	if err := g.CastSpell(me.ID, ritual, CastSpellParams{}); !errors.Is(err, ErrSorcerySpeedRequired) {
		t.Errorf("sorcery over a trigger: err = %v, want ErrSorcerySpeedRequired", err)
	}
	if err := g.CastSpell(me.ID, bear, CastSpellParams{}); !errors.Is(err, ErrSorcerySpeedRequired) {
		t.Errorf("creature over a trigger: err = %v, want ErrSorcerySpeedRequired", err)
	}
	if err := g.CastSpell(me.ID, land, CastSpellParams{}); !errors.Is(err, ErrSorcerySpeedRequired) {
		t.Errorf("land play over a trigger: err = %v, want ErrSorcerySpeedRequired (CR 305.1)", err)
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("mana pool = %d, want 2 — a refused cast paid for itself", len(me.ManaPool))
	}
	if err := g.CastSpell(me.ID, bolt, CastSpellParams{}); err != nil {
		t.Errorf("an instant over a trigger is instant-speed (CR 117.1a): %v", err)
	}

	resolveWholeStackForTest(t, g)
	if err := g.CastSpell(me.ID, ritual, CastSpellParams{}); err != nil {
		t.Errorf("the sorcery with the stack empty again: %v", err)
	}
}

// ActivateCatalogAbility through ActivationTimingOpenLocked: an
// "activate only as a sorcery" ability — equip's shape — and a
// loyalty ability are refused over a trigger, before anything is
// paid, and go through once it has resolved.
func TestATriggerOnTheStackBlocksASorcerySpeedActivation(t *testing.T) {
	for _, tc := range []struct {
		name string
		cost AbilityCost
		sorc bool
	}{
		{"equip-style: activate only as a sorcery", AbilityCost{Mana: "{1}"}, true},
		{"a loyalty ability (CR 606.3)", AbilityCost{Loyalty: loyaltyN(1)}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			src := NewCard("Test Sorcery-Speed Source", me.ID)
			src.TypeLine = "Artifact — Equipment"
			src.Controller = me.ID
			src.Counters = map[string]int{CounterLoyalty: 3}
			src.ActivatedAbilities = []ActivatedAbilityShape{{
				Label:        tc.name,
				Cost:         tc.cost,
				SorcerySpeed: tc.sorc,
				Effect:       func(*Game, *StackItem) error { return nil },
			}}
			g.Battlefield.PushTop(src)
			me.ManaPool.AddMana(ManaToken{Color: "C"})

			trig := stackTriggerForTest(t, g, me)
			err := g.ActivateCatalogAbility(me.ID, src.InstanceID, 0, ActivateAbilityParams{})
			if !errors.Is(err, ErrSorcerySpeedRequired) {
				t.Fatalf("activation over a trigger: err = %v, want ErrSorcerySpeedRequired", err)
			}
			if len(g.StackMeta) != 1 || g.StackMeta[trig] == nil {
				t.Errorf("StackMeta has %d items, want only the trigger", len(g.StackMeta))
			}
			if len(me.ManaPool) != 1 || counterOf(g, src.InstanceID, CounterLoyalty) != 3 {
				t.Error("a refused activation paid its cost")
			}

			resolveWholeStackForTest(t, g)
			if err := g.ActivateCatalogAbility(me.ID, src.InstanceID, 0, ActivateAbilityParams{}); err != nil {
				t.Errorf("activation with the stack empty again: %v", err)
			}
		})
	}
}

// The sandbox loyalty verb asks the same read.
func TestATriggerOnTheStackBlocksTheManualLoyaltyVerb(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 4, 1)
	stackTriggerForTest(t, g, me)
	if err := g.ActivateLoyalty(me.ID, pw, "+1", 1); !errors.Is(err, ErrSorcerySpeedRequired) {
		t.Errorf("ActivateLoyalty over a trigger: err = %v, want ErrSorcerySpeedRequired", err)
	}
}

// Suspend's window is a sorcery's own casting window for a sorcery
// (CR 702.62c), so a trigger shuts it; an instant stays suspendable.
func TestATriggerOnTheStackBlocksSuspendingASorcery(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	sorcery := seedHandCard(me, "Rift Bolt", "oracle-suspend", "Sorcery", "{2}{R}")
	instant := seedHandCard(me, "Test Instant", "oracle-suspend-instant", "Instant", "{2}{R}")
	stackTriggerForTest(t, g, g.Seats[1])
	var sorcOK, instOK bool
	g.ReadSnapshot(func() {
		sorcOK = g.SpecialActionTimingOKLocked(me.ID, sorcery, SpecialActionSuspend)
		instOK = g.SpecialActionTimingOKLocked(me.ID, instant, SpecialActionSuspend)
	})
	if sorcOK {
		t.Error("suspending a sorcery over a trigger was allowed")
	}
	if !instOK {
		t.Error("suspending an instant over a trigger was refused — its window is instant speed")
	}
}
