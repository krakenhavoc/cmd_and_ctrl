package game

import (
	"testing"

	"github.com/google/uuid"
)

// life_cost_count_test.go — #1594, ADR 0020 Decision 47: the engine
// half of a computed life cost. The card half (War Room, Murderous
// Betrayal, Lurking Evil) is in effects/computed_life_cost_test.go.

// testLifeFromCounters is a count the test board controls: the number
// of "life-tax" counters on the source.
var testLifeFromCounters = LifeCount("test/life-from-source-counters", func(g *Game, _ uuid.UUID, source uuid.UUID) int {
	if c := findBattlefieldCard(g, source); c != nil {
		return c.Counters["life-tax"]
	}
	return 0
})

func pushLifeCountSource(g *Game, owner *Player, cost AbilityCost) uuid.UUID {
	c := NewCard("Life Count Source", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "pay the count: mark",
		Cost:  cost,
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// The printed Life and the count are summed, read once, and charged.
func TestComputedLifeCostIsReadAtAnnounceAndCharged(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushLifeCountSource(g, me, AbilityCost{Life: 1, LifeFrom: testLifeFromCounters})
	findBattlefieldCard(g, src).Counters = map[string]int{"life-tax": 3}

	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := life - me.Life; got != 4 {
		t.Errorf("paid %d life, want 1 printed + 3 counted", got)
	}
	// The count moving after the announcement does not re-price the
	// activation on the stack.
	findBattlefieldCard(g, src).Counters["life-tax"] = 10
	for _, meta := range g.StackMeta {
		if meta != nil && meta.SourceCardID == src && meta.Paid.LifePaid != 4 {
			t.Errorf("LifePaid = %d, want 4", meta.Paid.LifePaid)
		}
	}
}

// An unregistered key is "cannot be paid", never "free": the engine
// refuses the activation with nothing spent.
func TestUnregisteredLifeCountIsRefusedNotFree(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushLifeCountSource(g, me, AbilityCost{LifeFrom: LifeCostCount{key: "test/never-registered"}})
	if _, ok := g.AbilityLifeCostLocked(me.ID, src, findBattlefieldCard(g, src).ActivatedAbilities[0].Cost); ok {
		t.Fatal("an unregistered count priced as ok")
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err == nil {
		t.Fatal("an ability whose life cost cannot be priced must be refused")
	}
	if me.Life != life || len(g.Stack.Cards) != 0 {
		t.Errorf("a refused activation spent something: life %d → %d, stack %d", life, me.Life, len(g.Stack.Cards))
	}
}

func TestLifeCountRegistrationRefusesMistakes(t *testing.T) {
	for name, register := range map[string]func(){
		"empty key": func() { LifeCount("", func(*Game, uuid.UUID, uuid.UUID) int { return 0 }) },
		"nil func":  func() { LifeCount("test/nil-func", nil) },
		"duplicate": func() {
			LifeCount("test/life-from-source-counters", func(*Game, uuid.UUID, uuid.UUID) int { return 0 })
		},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: LifeCount did not panic", name)
				}
			}()
			register()
		})
	}
}

// A negative count reads as 0 rather than refunding life.
func TestNegativeLifeCountReadsAsZero(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushLifeCountSource(g, me, AbilityCost{LifeFrom: testLifeFromCounters})
	findBattlefieldCard(g, src).Counters = map[string]int{"life-tax": -5}
	if life, ok := g.AbilityLifeCostLocked(me.ID, src, AbilityCost{LifeFrom: testLifeFromCounters}); !ok || life != 0 {
		t.Errorf("negative count priced %d (ok %v), want 0", life, ok)
	}
}
