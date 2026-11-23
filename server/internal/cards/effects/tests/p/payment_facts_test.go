package testsp

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// CR 400.7j: "If the cost of a spell or ability causes an object to
// move to a public zone, that spell or ability's effects can find that
// object." The activated-ability payer records the discard on the
// item, and the effect reads it back.
func TestAnAbilityRecordsTheCardItsCostDiscarded(t *testing.T) {
	var seen []uuid.UUID
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: pfDiscardReaderOracle, Name: "Discard Reader",
		Activated: []effects.ActivatedAbility{{
			Label: "Discard a card: Note the discarded card.",
			Cost:  effects.DiscardACard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				seen = effects.NewContext(g, item).Discarded()
				return nil
			},
		}},
	})
	g, me, _ := effectstest.ExileCostTable(t)
	src := effectstest.PushCatalogPermanent(g, me.ID, "Discard Reader", "Enchantment", pfDiscardReaderOracle, false)
	land := effectstest.PushCatalogHandCard(me, "Spare Land", "Land", "")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.OnlyStackItemOf(t, g).Paid.Discarded; len(got) != 1 || got[0] != land {
		t.Fatalf("the item's Paid.Discarded = %v, want [%v]", got, land)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: pfCyclerOracle, Name: "Recorded Cycler",
		Activated: []effects.ActivatedAbility{{
			Label:   "Cycling {0}",
			Cost:    effects.DiscardThis(),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Cycling: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				seen = effects.NewContext(g, item).Discarded()
				return nil
			},
		}},
	})
	g, me, _ := effectstest.ExileCostTable(t)
	cycler := effectstest.PushCatalogHandCard(me, "Recorded Cycler", "Creature — Test", pfCyclerOracle)
	if err := g.ActivateCatalogAbility(me.ID, cycler, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(seen) != 1 || seen[0] != cycler {
		t.Fatalf("the effect read %v, want the cycled card [%v]", seen, cycler)
	}
}

// counterThiefSpec is Simic Manipulator's ability without the card:
// "{T}, Remove one or more +1/+1 counters from this creature: Gain
// control of target creature with power less than or equal to the
// number of +1/+1 counters removed this way."
func counterThiefSpec() effects.Spec {
	return effects.Spec{
		OracleID: pfCounterThiefOracle, Name: "Counter Thief",
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Remove one or more +1/+1 counters from this creature: Gain control of target creature with power less than or equal to the number of +1/+1 counters removed this way.",
			Cost:  effects.Plus(effects.TapCost(), effects.RemoveCountersXFromThis("+1/+1", 1)),
			Targets: effects.TargetCreature("target creature with power less than or equal to the number of +1/+1 counters removed this way").
				WithPowerAtMostX().BoundByTheCountersRemoved(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, ref := range ctx.LegalTargets() {
					if ref.Kind == game.TargetCard {
						return effects.GainControl{Target: ref.ID, Duration: game.IndefiniteDuration()}.Apply(ctx)
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

// The bound is the number of counters the activation REMOVES, known
// before targets are chosen (CR 602.2b, 601.2b–c): a creature with
// more power than that is refused, with nothing paid.
func TestATargetBoundedByTheCountersRemoved(t *testing.T) {
	effectstest.RegisterForTest(t, counterThiefSpec())
	g, me, opp := effectstest.ExileCostTable(t)
	thief := effectstest.PushCatalogPermanent(g, me.ID, "Counter Thief", "Creature — Mutant Wizard", pfCounterThiefOracle, false)
	pfSetPT(g, thief, 0, 1)
	effectstest.PfAddCounters(g, thief, "+1/+1", 3)
	two := effectstest.PushCatalogPermanent(g, opp.ID, "Two Power", "Creature — Test", "", false)
	pfSetPT(g, two, 2, 2)
	four := effectstest.PushCatalogPermanent(g, opp.ID, "Four Power", "Creature — Test", "", false)
	pfSetPT(g, four, 4, 4)

	err := g.ActivateCatalogAbility(me.ID, thief, 0, game.ActivateAbilityParams{
		CounterCounts: []int{1}, Targets: effectstest.CardRefs(two),
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("one counter removed for a power-2 creature: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, thief, 0, game.ActivateAbilityParams{
		CounterCounts: []int{3}, Targets: effectstest.CardRefs(four),
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
		CounterCounts: []int{2}, Targets: effectstest.CardRefs(two),
	}); err != nil {
		t.Fatalf("two counters removed for a power-2 creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PfController(g, two) != me.ID {
		t.Error("the power-2 creature was not taken")
	}
}

// CR 608.2b: the bound is checked again at resolution, from the count
// the item carries. A creature pumped past it in response is an
// illegal target, and the ability does nothing.
func TestTheCountersRemovedBoundIsRecheckedAtResolution(t *testing.T) {
	effectstest.RegisterForTest(t, counterThiefSpec())
	g, me, opp := effectstest.ExileCostTable(t)
	thief := effectstest.PushCatalogPermanent(g, me.ID, "Counter Thief", "Creature — Mutant Wizard", pfCounterThiefOracle, false)
	pfSetPT(g, thief, 0, 1)
	effectstest.PfAddCounters(g, thief, "+1/+1", 2)
	two := effectstest.PushCatalogPermanent(g, opp.ID, "Two Power", "Creature — Test", "", false)
	pfSetPT(g, two, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, thief, 0, game.ActivateAbilityParams{
		CounterCounts: []int{2}, Targets: effectstest.CardRefs(two),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.OnlyStackItemOf(t, g).Paid.CountersRemoved; got != 2 {
		t.Fatalf("the item records %d counters removed, want 2", got)
	}
	effectstest.PfAddCounters(g, two, "+1/+1", 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PfController(g, two) != opp.ID {
		t.Error("a creature pumped to power 3 was taken by a two-counter activation")
	}
}

// effects.Register refuses a counters-removed bound on an ability whose
// cost removes no variable number of counters, and on a spell.
func TestRegisterRefusesAMisplacedCountersRemovedBound(t *testing.T) {
	cases := map[string]effects.Spec{
		"fixed removal": {
			OracleID: "test-pf-bad-fixed", Name: "Bad Fixed",
			Activated: []effects.ActivatedAbility{{
				Label: "Remove two counters: steal", Cost: effects.RemoveCountersFromThis("+1/+1", 2),
				Targets: effects.TargetCreature("x").WithPowerAtMostX().BoundByTheCountersRemoved(),
				Effect:  func(*game.Game, *game.StackItem) error { return nil },
			}},
		},
		"spell": {
			OracleID: "test-pf-bad-spell", Name: "Bad Spell",
			Targets: effects.TargetCreature("x").WithPowerAtMostX().BoundByTheCountersRemoved(),
		},
		"no statistic": {
			OracleID: "test-pf-bad-nostat", Name: "Bad No Stat",
			Activated: []effects.ActivatedAbility{{
				Label: "Remove X counters: steal", Cost: effects.RemoveCountersXFromThis("+1/+1", 1),
				Targets: effects.TargetCreature("x").BoundByTheCountersRemoved(),
				Effect:  func(*game.Game, *game.StackItem) error { return nil },
			}},
		},
		"two statistics": {
			OracleID: "test-pf-bad-twostat", Name: "Bad Two Stats",
			Activated: []effects.ActivatedAbility{{
				Label: "{X}: steal", Cost: effects.ManaCost("{X}"),
				Targets: effects.TargetCreature("x").WithPowerAtMostX().WithToughnessAtMostX(),
				Effect:  func(*game.Game, *game.StackItem) error { return nil },
			}},
		},
		"power bound without X": {
			OracleID: "test-pf-bad-nox", Name: "Bad No X",
			Activated: []effects.ActivatedAbility{{
				Label: "{1}: steal", Cost: effects.ManaCost("{1}"),
				Targets: effects.TargetCreature("x").WithPowerAtMostX(),
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
				delete(effects.Registry, spec.OracleID)
			}()
			effects.Register(spec)
		})
	}
}
