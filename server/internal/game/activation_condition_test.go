package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// activation_condition_test.go — #743, ADR 0020's activation-condition
// addendum: ActivatedAbilityShape.Condition, the CR 602.1b "Activate
// only if …" gate on a non-mana activated ability.
//
// What is pinned here is the engine half: the gate runs before
// anything is announced or paid, it has its own error, it sits beside
// (not inside) the sorcery-speed check, and it is never consulted
// again at resolution. The catalog half — the helpers and the cards
// that use them — is in cards/effects/activation_conditions_test.go.

// pushConditionSource puts an artifact on the battlefield whose one
// ability costs every component a failed gate must leave unpaid —
// {1}, {T}, "Sacrifice a creature", and a charge counter off itself —
// and whose effect stamps a marker counter so a test can see it ran.
func pushConditionSource(g *Game, owner *Player, sorcery bool, cond func(*Game, uuid.UUID, uuid.UUID) bool) uuid.UUID {
	c := NewCard("Condition Source", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.Counters = map[string]int{"charge": 2}
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "{1}, {T}, Remove a charge counter, Sacrifice a creature: mark. Activate only if …",
		Cost: AbilityCost{
			Mana:           "{1}",
			Tap:            true,
			SacrificeOther: creatureCostSpec(),
			RemoveCounters: &CounterRemovalCost{Counter: "charge", N: 1},
		},
		SorcerySpeed: sorcery,
		Condition:    cond,
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func conditionFixed(ok bool) func(*Game, uuid.UUID, uuid.UUID) bool {
	return func(*Game, uuid.UUID, uuid.UUID) bool { return ok }
}

func tappedOf(g *Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	return false
}

// A false condition refuses with ErrConditionNotMet and every cost
// component stays unpaid: the pool, the tap, the sacrifice, the
// counters, and no stack item or trigger.
func TestActivationConditionFalsePaysNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushConditionSource(g, me, false, conditionFixed(false))
	fodder := pushPermanentWithCounters(g, me, "Fodder", "Creature — Bear", nil)
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	eventsBefore := len(g.Events)

	err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if !errors.Is(err, ErrConditionNotMet) {
		t.Fatalf("err = %v, want ErrConditionNotMet", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %v, want the {C} still floating", me.ManaPool)
	}
	if tappedOf(g, src) {
		t.Error("the source tapped")
	}
	if !g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice was paid")
	}
	if got := counterOf(g, src, "charge"); got != 2 {
		t.Errorf("charge counters = %d, want 2", got)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("StackMeta has %d items, want none", len(g.StackMeta))
	}
	for _, ev := range g.Events[eventsBefore:] {
		if ev.Kind == EventTrigger || ev.Kind == EventSacrifice {
			t.Errorf("event %s emitted by a refused activation", ev.Kind)
		}
	}

	// The same activation with the gate open goes through and pays.
	src2 := pushConditionSource(g, me, false, conditionFixed(true))
	if err := g.ActivateCatalogAbility(me.ID, src2, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate with the condition met: %v", err)
	}
	if len(me.ManaPool) != 0 || !tappedOf(g, src2) || g.Battlefield.Contains(fodder) || counterOf(g, src2, "charge") != 1 {
		t.Error("a met condition should pay every component")
	}
	passBothForTest(g)
	if counterOf(g, src2, "effect-ran") != 1 {
		t.Error("the ability did not resolve")
	}
}

// The condition and sorcery speed are separate gates with separate
// errors. Each failing alone refuses with its own; both failing
// refuses with the timing error, which is checked first.
func TestActivationConditionAndSorcerySpeedAreSeparate(t *testing.T) {
	cases := []struct {
		name      string
		mainPhase bool
		condition bool
		want      error
	}{
		{"condition fails, timing open", true, false, ErrConditionNotMet},
		{"timing shut, condition holds", false, true, ErrSorcerySpeedRequired},
		{"both fail", false, false, ErrSorcerySpeedRequired},
		{"both hold", true, true, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			step := StepUpkeep
			if tc.mainPhase {
				step = StepPrecombatMain
			}
			advanceTo(t, g, step)
			me := g.Seats[0]
			src := pushConditionSource(g, me, true, conditionFixed(tc.condition))
			fodder := pushPermanentWithCounters(g, me, "Fodder", "Creature — Bear", nil)
			me.ManaPool.AddMana(ManaToken{Color: "C"})
			err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
			if tc.want == nil {
				if err != nil {
					t.Fatalf("activate: %v", err)
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

// CR 602.1b: the instruction is not part of the effect, so a
// condition that stops holding after activation does not stop the
// ability resolving.
func TestActivationConditionIsNotRecheckedAtResolution(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	open := true
	calls := 0
	src := pushConditionSource(g, me, false, func(*Game, uuid.UUID, uuid.UUID) bool {
		calls++
		return open
	})
	fodder := pushPermanentWithCounters(g, me, "Fodder", "Creature — Bear", nil)
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	open = false
	atActivation := calls
	passBothForTest(g)
	if counterOf(g, src, "effect-ran") != 1 {
		t.Error("the ability should resolve although its condition no longer holds")
	}
	if calls != atActivation {
		t.Errorf("the condition was evaluated %d more time(s) during resolution", calls-atActivation)
	}
}

// The closure receives the activating player and the source's
// instance ID.
func TestActivationConditionReceivesControllerAndSource(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	var gotController, gotSource uuid.UUID
	src := pushConditionSource(g, me, false, func(_ *Game, controller, source uuid.UUID) bool {
		gotController, gotSource = controller, source
		return false
	})
	_ = g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{})
	if gotController != me.ID || gotSource != src {
		t.Errorf("condition saw (%v, %v), want (%v, %v)", gotController, gotSource, me.ID, src)
	}
}
