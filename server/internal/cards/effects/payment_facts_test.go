package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// payment_facts_test.go — ADR 0109 §8 and §9 on the activation path:
// a discard recorded on an activated ability (#1862) and a target
// bounded by the counters its cost removed (#1842), against abilities
// registered for the test. The card halves are in each card's own
// test file.

const (
	pfDiscardReaderOracle = "test-pf-discard-reader"
	pfCyclerOracle        = "test-pf-cycler"
	pfCounterThiefOracle  = "test-pf-counter-thief"
)

// onlyStackItemOf is the single item on the stack, or a failure.
func onlyStackItemOf(t *testing.T, g *game.Game) *game.StackItem {
	t.Helper()
	if len(g.StackMeta) != 1 {
		t.Fatalf("stack holds %d items, want 1", len(g.StackMeta))
	}
	for _, it := range g.StackMeta {
		return it
	}
	return nil
}

// CR 400.7j: "If the cost of a spell or ability causes an object to
// move to a public zone, that spell or ability's effects can find that
// object." The activated-ability payer records the discard on the
// item, and the effect reads it back.
func TestAnAbilityRecordsTheCardItsCostDiscarded(t *testing.T) {
	var seen []uuid.UUID
	registerForTest(t, Spec{
		OracleID: pfDiscardReaderOracle, Name: "Discard Reader",
		Activated: []ActivatedAbility{{
			Label: "Discard a card: Note the discarded card.",
			Cost:  DiscardACard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				seen = NewContext(g, item).Discarded()
				return nil
			},
		}},
	})
	g, me, _ := exileCostTable(t)
	src := pushCatalogPermanent(g, me.ID, "Discard Reader", "Enchantment", pfDiscardReaderOracle, false)
	land := pushCatalogHandCard(me, "Spare Land", "Land", "")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := onlyStackItemOf(t, g).Paid.Discarded; len(got) != 1 || got[0] != land {
		t.Fatalf("the item's Paid.Discarded = %v, want [%v]", got, land)
	}
	passPriorityAroundTable(t, g)
	if len(seen) != 1 || seen[0] != land {
		t.Fatalf("the effect read %v, want [%v]", seen, land)
	}
	if !me.Graveyard.Contains(land) {
		t.Error("the discarded card is not in the graveyard")
	}
}

// Cycling's "Discard this card" is a discarded card too, and CR 400.7j
// does not tell them apart.
func TestCyclingRecordsTheCycledCard(t *testing.T) {
	var seen []uuid.UUID
	registerForTest(t, Spec{
		OracleID: pfCyclerOracle, Name: "Recorded Cycler",
		Activated: []ActivatedAbility{{
			Label:   "Cycling {0}",
			Cost:    DiscardThis(),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Cycling: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				seen = NewContext(g, item).Discarded()
				return nil
			},
		}},
	})
	g, me, _ := exileCostTable(t)
	cycler := pushCatalogHandCard(me, "Recorded Cycler", "Creature — Test", pfCyclerOracle)
	if err := g.ActivateCatalogAbility(me.ID, cycler, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(seen) != 1 || seen[0] != cycler {
		t.Fatalf("the effect read %v, want the cycled card [%v]", seen, cycler)
	}
}

// counterThiefSpec is Simic Manipulator's ability without the card:
// "{T}, Remove one or more +1/+1 counters from this creature: Gain
// control of target creature with power less than or equal to the
// number of +1/+1 counters removed this way."
func counterThiefSpec() Spec {
	return Spec{
		OracleID: pfCounterThiefOracle, Name: "Counter Thief",
		Activated: []ActivatedAbility{{
			Label: "{T}, Remove one or more +1/+1 counters from this creature: Gain control of target creature with power less than or equal to the number of +1/+1 counters removed this way.",
			Cost:  Plus(TapCost(), RemoveCountersXFromThis("+1/+1", 1)),
			Targets: TargetCreature("target creature with power less than or equal to the number of +1/+1 counters removed this way").
				WithPowerAtMostX().BoundByTheCountersRemoved(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, ref := range ctx.LegalTargets() {
					if ref.Kind == game.TargetCard {
						return GainControl{Target: ref.ID, Duration: game.IndefiniteDuration()}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	}
}

// pfSetPT gives a battlefield card a printed power and toughness.
func pfSetPT(g *game.Game, id uuid.UUID, p, tough int) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = p, tough
			}
		}
	})
	g.BumpLayerVersionForTest()
}

// pfAddCounters puts n counters of `kind` on a battlefield card.
func pfAddCounters(g *game.Game, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID == id {
				if c.Counters == nil {
					c.Counters = map[string]int{}
				}
				c.Counters[kind] += n
			}
		}
	})
	g.BumpLayerVersionForTest()
}

func pfController(g *game.Game, id uuid.UUID) uuid.UUID {
	var out uuid.UUID
	g.WithWriteLock(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			out = c.Controller
		}
	})
	return out
}

// The bound is the number of counters the activation REMOVES, known
// before targets are chosen (CR 602.2b, 601.2b–c): a creature with
// more power than that is refused, with nothing paid.
func TestATargetBoundedByTheCountersRemoved(t *testing.T) {
	registerForTest(t, counterThiefSpec())
	g, me, opp := exileCostTable(t)
	thief := pushCatalogPermanent(g, me.ID, "Counter Thief", "Creature — Mutant Wizard", pfCounterThiefOracle, false)
	pfSetPT(g, thief, 0, 1)
	pfAddCounters(g, thief, "+1/+1", 3)
	two := pushCatalogPermanent(g, opp.ID, "Two Power", "Creature — Test", "", false)
	pfSetPT(g, two, 2, 2)
	four := pushCatalogPermanent(g, opp.ID, "Four Power", "Creature — Test", "", false)
	pfSetPT(g, four, 4, 4)

	err := g.ActivateCatalogAbility(me.ID, thief, 0, game.ActivateAbilityParams{
		CounterCounts: []int{1}, Targets: cardRefs(two),
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("one counter removed for a power-2 creature: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, thief, 0, game.ActivateAbilityParams{
		CounterCounts: []int{3}, Targets: cardRefs(four),
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("three counters removed for a power-4 creature: err = %v, want ErrIllegalTarget", err)
	}
	var counters int
	g.WithWriteLock(func() {
		c, _ := g.LookupCardForEffect(thief)
		counters = c.Counters["+1/+1"]
	})
	if counters != 3 {
		t.Fatalf("a refused activation removed counters: %d left", counters)
	}
	if err := g.ActivateCatalogAbility(me.ID, thief, 0, game.ActivateAbilityParams{
		CounterCounts: []int{2}, Targets: cardRefs(two),
	}); err != nil {
		t.Fatalf("two counters removed for a power-2 creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if pfController(g, two) != me.ID {
		t.Error("the power-2 creature was not taken")
	}
}

// CR 608.2b: the bound is checked again at resolution, from the count
// the item carries. A creature pumped past it in response is an
// illegal target, and the ability does nothing.
func TestTheCountersRemovedBoundIsRecheckedAtResolution(t *testing.T) {
	registerForTest(t, counterThiefSpec())
	g, me, opp := exileCostTable(t)
	thief := pushCatalogPermanent(g, me.ID, "Counter Thief", "Creature — Mutant Wizard", pfCounterThiefOracle, false)
	pfSetPT(g, thief, 0, 1)
	pfAddCounters(g, thief, "+1/+1", 2)
	two := pushCatalogPermanent(g, opp.ID, "Two Power", "Creature — Test", "", false)
	pfSetPT(g, two, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, thief, 0, game.ActivateAbilityParams{
		CounterCounts: []int{2}, Targets: cardRefs(two),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := onlyStackItemOf(t, g).Paid.CountersRemoved; got != 2 {
		t.Fatalf("the item records %d counters removed, want 2", got)
	}
	pfAddCounters(g, two, "+1/+1", 1)
	passPriorityAroundTable(t, g)
	if pfController(g, two) != opp.ID {
		t.Error("a creature pumped to power 3 was taken by a two-counter activation")
	}
}

// effects.Register refuses a counters-removed bound on an ability whose
// cost removes no variable number of counters, and on a spell.
func TestRegisterRefusesAMisplacedCountersRemovedBound(t *testing.T) {
	cases := map[string]Spec{
		"fixed removal": {
			OracleID: "test-pf-bad-fixed", Name: "Bad Fixed",
			Activated: []ActivatedAbility{{
				Label: "Remove two counters: steal", Cost: RemoveCountersFromThis("+1/+1", 2),
				Targets: TargetCreature("x").WithPowerAtMostX().BoundByTheCountersRemoved(),
				Effect:  func(*game.Game, *game.StackItem) error { return nil },
			}},
		},
		"spell": {
			OracleID: "test-pf-bad-spell", Name: "Bad Spell",
			Targets: TargetCreature("x").WithPowerAtMostX().BoundByTheCountersRemoved(),
		},
		"no statistic": {
			OracleID: "test-pf-bad-nostat", Name: "Bad No Stat",
			Activated: []ActivatedAbility{{
				Label: "Remove X counters: steal", Cost: RemoveCountersXFromThis("+1/+1", 1),
				Targets: TargetCreature("x").BoundByTheCountersRemoved(),
				Effect:  func(*game.Game, *game.StackItem) error { return nil },
			}},
		},
		"two statistics": {
			OracleID: "test-pf-bad-twostat", Name: "Bad Two Stats",
			Activated: []ActivatedAbility{{
				Label: "{X}: steal", Cost: ManaCost("{X}"),
				Targets: TargetCreature("x").WithPowerAtMostX().WithToughnessAtMostX(),
				Effect:  func(*game.Game, *game.StackItem) error { return nil },
			}},
		},
		"power bound without X": {
			OracleID: "test-pf-bad-nox", Name: "Bad No X",
			Activated: []ActivatedAbility{{
				Label: "{1}: steal", Cost: ManaCost("{1}"),
				Targets: TargetCreature("x").WithPowerAtMostX(),
				Effect:  func(*game.Game, *game.StackItem) error { return nil },
			}},
		},
	}
	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("Register accepted it")
				}
				delete(registry, spec.OracleID)
			}()
			Register(spec)
		})
	}
}
