package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// counter_cost_rest_test.go — #789: the rest of the counter-cost
// seam, on top of #625's three shapes.
//
// Four things arrive here and each has a card behind it:
//
//	counter costs on MANA abilities  Vivid Creek, Ramos, Mage-Ring
//	a VARIABLE count                 Mage-Ring Network
//	a removal SPLIT across permanents Iron Spider, Hopeful Initiate
//	a cost that ADDS a counter       Devoted Druid
//
// The catalog half — the real card files — is in
// cards/effects/counter_cost_rest_cards_test.go. What is pinned here
// is the engine's contract: one component with two owners, validated
// as a set, paid at announce, never replaceable, and never planned by
// the auto-tapper when it cannot both decide and afford it.

// artifactCostSpec is "artifacts you control" as a cost predicate —
// Iron Spider's clause, without the targeting gate.
func artifactCostSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "permanent",
		Label: "artifacts you control",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsArtifact()
		},
		Min: 1, Max: 1,
	}
}

// pushManaSource seats a land whose one mana ability has `cost`,
// producing `produced`.
func pushManaSource(g *Game, owner *Player, ab ManaAbilityShape, counters map[string]int) uuid.UUID {
	c := NewCard("Test Mana Source", owner.ID)
	c.TypeLine = "Land"
	c.Controller = owner.ID
	c.Counters = counters
	c.ManaAbilities = []ManaAbilityShape{ab}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// --- counter costs on mana abilities -------------------------------

// The headline for the mana half: Vivid Creek's "{T}, Remove a charge
// counter from this land: Add one mana of any color". The counter
// comes off, the land taps, and the mana arrives.
func TestManaAbilityCounterCostIsPaid(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	land := pushManaSource(g, me, ManaAbilityShape{
		TapCost:        true,
		RemoveCounters: &CounterRemovalCost{Counter: "charge", N: 1},
		Produced:       "{U}",
		Label:          "Add one mana of any color",
	}, map[string]int{"charge": 2})

	if err := g.ActivateManaAbility(me.ID, land, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterOf(g, land, "charge"); got != 1 {
		t.Errorf("charge counters %d, want 1", got)
	}
	if len(me.ManaPool) != 1 {
		t.Fatalf("pool has %d tokens, want 1", len(me.ManaPool))
	}
	if !findBattlefieldCard(g, land).Tapped {
		t.Error("the land did not tap")
	}
}

// The other half of the rule: a Vivid land out of charge counters
// cannot fire its second ability at all, and nothing is spent trying.
func TestManaAbilityCounterCostRefusedWithNothingToRemove(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	land := pushManaSource(g, me, ManaAbilityShape{
		TapCost:        true,
		RemoveCounters: &CounterRemovalCost{Counter: "charge", N: 1},
		Produced:       "{U}",
	}, nil)

	err := g.ActivateManaAbility(me.ID, land, 0, ManaAbilityParams{})
	if !errors.Is(err, ErrInsufficientCounters) {
		t.Fatalf("activate: %v, want ErrInsufficientCounters", err)
	}
	if findBattlefieldCard(g, land).Tapped {
		t.Error("the land tapped for a cost it could not pay")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool has %d tokens, want 0", len(me.ManaPool))
	}
}

// Ramos, Dragon Engine: "Remove five +1/+1 counters from Ramos: Add
// {W}{W}{U}{U}{B}{B}{R}{R}{G}{G}." No tap, five counters, ten mana.
func TestManaAbilityCounterCostWithoutATap(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	ramos := pushManaSource(g, me, ManaAbilityShape{
		RemoveCounters: &CounterRemovalCost{Counter: "+1/+1", N: 5},
		Produced:       "{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}",
	}, map[string]int{"+1/+1": 6})

	if err := g.ActivateManaAbility(me.ID, ramos, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterOf(g, ramos, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	if len(me.ManaPool) != 10 {
		t.Fatalf("pool has %d tokens, want 10", len(me.ManaPool))
	}
	if findBattlefieldCard(g, ramos).Tapped {
		t.Error("an ability with no {T} tapped its source")
	}
}

// A counter cost on a mana ability is paid with the SAME component
// the CR 602 path uses, so a permanent that cannot pay refuses the
// same way — here, four counters against a printed five.
func TestManaAbilityCounterCostShortOfThePrintedCount(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	ramos := pushManaSource(g, me, ManaAbilityShape{
		RemoveCounters: &CounterRemovalCost{Counter: "+1/+1", N: 5},
		Produced:       "{W}",
	}, map[string]int{"+1/+1": 4})

	if err := g.ActivateManaAbility(me.ID, ramos, 0, ManaAbilityParams{}); !errors.Is(err, ErrInsufficientCounters) {
		t.Fatalf("activate: %v, want ErrInsufficientCounters", err)
	}
	if got := counterOf(g, ramos, "+1/+1"); got != 4 {
		t.Errorf("counters were spent on a refused activation: %d, want 4", got)
	}
}

// --- the auto-tapper -----------------------------------------------

// The rule the issue asks for by name: the planner never books a
// Vivid land it cannot charge.
func TestAutoTapperSkipsAnUnpayableCounterCostLand(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	vivid := ManaAbilityShape{
		TapCost:        true,
		RemoveCounters: &CounterRemovalCost{Counter: "charge", N: 1},
		Produced:       "{U}",
	}
	charged := pushManaSource(g, me, vivid, map[string]int{"charge": 1})
	empty := pushManaSource(g, me, vivid, nil)

	cost, err := ParseCost("{U}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok {
		t.Fatalf("no plan; the charged land should have funded {U}")
	}
	if len(plan) != 1 || plan[0] != charged {
		t.Fatalf("plan = %v, want just the charged land %v", plan, charged)
	}
	if plan[0] == empty {
		t.Error("the planner booked a land with no charge counter")
	}
}

// And with NO land that can pay, the plan fails rather than tapping
// a land for mana it cannot mint.
func TestAutoTapperWontPlanAnyUnpayableCounterCostLand(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pushManaSource(g, me, ManaAbilityShape{
		TapCost:        true,
		RemoveCounters: &CounterRemovalCost{Counter: "charge", N: 1},
		Produced:       "{U}",
	}, nil)

	cost, _ := ParseCost("{U}")
	if _, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Fatal("planned a land whose counter cost cannot be paid")
	}
}

// The executor pays what the planner booked: the plan taps the land
// AND takes the counter, so the two halves of the auto-tapper agree
// with the hand-clicked activation.
func TestAutoTapExecutorPaysTheCounter(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	land := pushManaSource(g, me, ManaAbilityShape{
		TapCost:        true,
		RemoveCounters: &CounterRemovalCost{Counter: "charge", N: 1},
		Produced:       "{U}",
	}, map[string]int{"charge": 2})

	cost, _ := ParseCost("{U}")
	g.mu.Lock()
	plan, ok := g.autoTapLocked(me.ID, cost, 0, nil)
	if ok {
		g.materializePlanLocked(me, plan, cost)
	}
	g.mu.Unlock()
	if !ok {
		t.Fatal("no plan")
	}
	if got := counterOf(g, land, "charge"); got != 1 {
		t.Errorf("charge counters %d, want 1 — the executor did not pay", got)
	}
	if !findBattlefieldCard(g, land).Tapped {
		t.Error("the land did not tap")
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool has %d tokens, want 1", len(me.ManaPool))
	}
}

// A counter cost the planner cannot DECIDE — a variable count, an
// any-kind pick, a removal from another permanent — is not planned
// even when it could be paid. The player clicks it.
func TestAutoTapperWontDecideACounterCost(t *testing.T) {
	cases := []struct {
		name string
		rc   *CounterRemovalCost
	}{
		{"variable", &CounterRemovalCost{Counter: "storage", N: 0, Variable: true}},
		{"any kind", &CounterRemovalCost{Counter: "", N: 1}},
		{"from another permanent", &CounterRemovalCost{Counter: "charge", N: 1, From: artifactCostSpec()}},
		{"among several", &CounterRemovalCost{Counter: "charge", N: 2, From: artifactCostSpec(), Among: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			pushManaSource(g, me, ManaAbilityShape{
				TapCost:        true,
				RemoveCounters: tc.rc,
				Produced:       "{U}",
			}, map[string]int{"charge": 5, "storage": 5})

			cost, _ := ParseCost("{U}")
			if _, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
				t.Fatal("the planner answered a question only the player can")
			}
		})
	}
}

// A cost that ADDS a counter is a resource the player never agreed to
// spend, like a life cost: the planner leaves it alone.
func TestAutoTapperWontAddACounter(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pushManaSource(g, me, ManaAbilityShape{
		TapCost:    true,
		AddCounter: &CounterAddCost{Counter: "-1/-1", N: 1},
		Produced:   "{G}",
	}, nil)

	cost, _ := ParseCost("{G}")
	if _, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Fatal("the planner spent a counter the player never agreed to")
	}
}

// --- a variable count ----------------------------------------------

// Mage-Ring Network: "Remove any number of storage counters from this
// land: Add {C} for each storage counter removed this way." The count
// the activator announces reaches the produced-mana computation
// through the one paid-cost record.
func TestVariableCounterCostReachesTheProducedMana(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	land := pushManaSource(g, me, ManaAbilityShape{
		TapCost:        true,
		RemoveCounters: &CounterRemovalCost{Counter: "storage", N: 0, Variable: true},
		ProducedForPaid: func(_ *Game, _, _ uuid.UUID, paid PaidCost) string {
			if paid.CountersRemoved <= 0 {
				return ""
			}
			return "{C" + itoaForTest(paid.CountersRemoved) + "}"
		},
	}, map[string]int{"storage": 4})

	if err := g.ActivateManaAbility(me.ID, land, 0, ManaAbilityParams{CounterCounts: []int{3}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterOf(g, land, "storage"); got != 1 {
		t.Errorf("storage counters %d, want 1", got)
	}
	if len(me.ManaPool) != 3 {
		t.Fatalf("pool has %d tokens, want 3 — the paid count did not reach the effect", len(me.ManaPool))
	}
}

// The announced count is bounded by what the permanent actually
// holds: naming more than it has refuses the whole activation.
func TestVariableCounterCostCannotOverdraw(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	land := pushManaSource(g, me, ManaAbilityShape{
		TapCost:        true,
		RemoveCounters: &CounterRemovalCost{Counter: "storage", N: 0, Variable: true},
		Produced:       "{C}",
	}, map[string]int{"storage": 2})

	err := g.ActivateManaAbility(me.ID, land, 0, ManaAbilityParams{CounterCounts: []int{3}})
	if !errors.Is(err, ErrInsufficientCounters) {
		t.Fatalf("activate: %v, want ErrInsufficientCounters", err)
	}
	if got := counterOf(g, land, "storage"); got != 2 {
		t.Errorf("counters were spent on a refused activation: %d, want 2", got)
	}
}

// On an ACTIVATED ability the announced count rides the same record
// and is read back at resolution — the counters are off the board by
// then, so nothing else could tell the effect how many there were.
func TestVariableCounterCostOnAnActivatedAbilityReachesResolution(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	var sawCount int
	c := NewCard("Variable Source", me.ID)
	c.TypeLine = "Artifact"
	c.Controller = me.ID
	c.Counters = map[string]int{"storage": 5}
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "remove X storage counters",
		Cost:  AbilityCost{RemoveCounters: &CounterRemovalCost{Counter: "storage", N: 1, Variable: true}},
		Effect: func(_ *Game, item *StackItem) error {
			sawCount = item.Paid.CountersRemoved
			return nil
		},
	}}
	g.Battlefield.PushTop(c)

	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{
		CounterCounts: []int{4},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterOf(g, c.InstanceID, "storage"); got != 1 {
		t.Errorf("storage counters %d, want 1", got)
	}
	passBothForTest(g)
	if sawCount != 4 {
		t.Errorf("the effect saw %d counters removed, want 4", sawCount)
	}
}

// A variable cost with a printed floor refuses an announcement under
// it, the way MinX refuses an X below "X can't be 0".
func TestVariableCounterCostHonoursItsFloor(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, AbilityCost{
		RemoveCounters: &CounterRemovalCost{Counter: "storage", N: 2, Variable: true},
	}, nil)
	setCountersForTest(g, src, map[string]int{"storage": 5})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		CounterCounts: []int{1},
	}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("activate under the floor: %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		CounterCounts: []int{2},
	}); err != nil {
		t.Fatalf("activate at the floor: %v", err)
	}
}

// --- a removal split across permanents -----------------------------

// Iron Spider, Stark Upgrade: "Remove two +1/+1 counters from among
// artifacts you control." Any split that totals two is legal.
func TestAmongCounterCostSplitsAcrossPermanents(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, AbilityCost{
		RemoveCounters: &CounterRemovalCost{Counter: "+1/+1", N: 2, From: artifactCostSpec(), Among: true},
	}, nil)
	a := pushPermanentWithCounters(g, me, "Artifact A", "Artifact", map[string]int{"+1/+1": 1})
	b := pushPermanentWithCounters(g, me, "Artifact B", "Artifact", map[string]int{"+1/+1": 3})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{a, b},
		CounterCounts:    []int{1, 1},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterOf(g, a, "+1/+1"); got != 0 {
		t.Errorf("A has %d counters, want 0", got)
	}
	if got := counterOf(g, b, "+1/+1"); got != 2 {
		t.Errorf("B has %d counters, want 2", got)
	}
}

// The whole payment is validated as a SET (the crew shape): a total
// that misses N, a permanent named twice, a permanent that does not
// match the clause and a permanent short of its own share all refuse
// the activation with nothing removed.
func TestAmongCounterCostIsValidatedAsASet(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, AbilityCost{
		RemoveCounters: &CounterRemovalCost{Counter: "+1/+1", N: 3, From: artifactCostSpec(), Among: true},
	}, nil)
	a := pushPermanentWithCounters(g, me, "Artifact A", "Artifact", map[string]int{"+1/+1": 2})
	b := pushPermanentWithCounters(g, me, "Artifact B", "Artifact", map[string]int{"+1/+1": 2})
	notArtifact := pushPermanentWithCounters(g, me, "Bear", "Creature — Bear", map[string]int{"+1/+1": 3})

	cases := []struct {
		name  string
		ids   []uuid.UUID
		count []int
		want  error
	}{
		{"short of the printed total", []uuid.UUID{a, b}, []int{1, 1}, ErrInvalidParam},
		{"over the printed total", []uuid.UUID{a, b}, []int{2, 2}, ErrInvalidParam},
		{"one permanent named twice", []uuid.UUID{a, a}, []int{2, 1}, ErrInvalidParam},
		{"a part that removes nothing", []uuid.UUID{a, b}, []int{2, 0}, ErrInvalidParam},
		{"a permanent that does not match", []uuid.UUID{notArtifact}, []int{3}, ErrIllegalTarget},
		{"a permanent short of its share", []uuid.UUID{a}, []int{3}, ErrInsufficientCounters},
		{"counts that do not line up with the ids", []uuid.UUID{a, b}, []int{3}, ErrInvalidParam},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
				CounterSourceIDs: tc.ids,
				CounterCounts:    tc.count,
			})
			if !errors.Is(err, tc.want) {
				t.Fatalf("activate: %v, want %v", err, tc.want)
			}
			if counterOf(g, a, "+1/+1")+counterOf(g, b, "+1/+1") != 4 {
				t.Fatal("a refused payment removed counters")
			}
		})
	}
}

// One permanent may pay the whole among cost — "from among" is a
// clause about where the counters may come from, not a requirement to
// spread them.
func TestAmongCounterCostFromOnePermanent(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, AbilityCost{
		RemoveCounters: &CounterRemovalCost{Counter: "+1/+1", N: 2, From: artifactCostSpec(), Among: true},
	}, nil)
	a := pushPermanentWithCounters(g, me, "Artifact A", "Artifact", map[string]int{"+1/+1": 4})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{a},
		CounterCounts:    []int{2},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterOf(g, a, "+1/+1"); got != 2 {
		t.Errorf("A has %d counters, want 2", got)
	}
}

// --- a cost that adds a counter ------------------------------------

// Devoted Druid: "Put a -1/-1 counter on this creature: Untap this
// creature." The counter goes on at announce, before the effect runs.
func TestAddCounterCostIsPaidAtAnnounce(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	var sawAtResolution int
	c := NewCard("Devoted Test Druid", me.ID)
	c.TypeLine = "Creature — Elf Druid"
	c.Power, c.Toughness = 0, 2
	c.Controller = me.ID
	c.Tapped = true
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "untap",
		Cost:  AbilityCost{AddCounter: &CounterAddCost{Counter: "-1/-1", N: 1}},
		Effect: func(g *Game, item *StackItem) error {
			sawAtResolution = counterOf(g, item.SourceCardID, "-1/-1")
			if t := findBattlefieldCard(g, item.SourceCardID); t != nil {
				t.Tapped = false
			}
			return nil
		},
	}}
	g.Battlefield.PushTop(c)

	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterOf(g, c.InstanceID, "-1/-1"); got != 1 {
		t.Fatalf("-1/-1 counters %d at announce, want 1 (CR 118.3: the cost is paid now)", got)
	}
	if !findBattlefieldCard(g, c.InstanceID).Tapped {
		t.Error("the effect ran at announce")
	}
	passBothForTest(g)
	if sawAtResolution != 1 {
		t.Errorf("the effect saw %d counters, want 1", sawAtResolution)
	}
	if findBattlefieldCard(g, c.InstanceID).Tapped {
		t.Error("the ability did not untap its source")
	}
}

// CR 121.1: paying a cost is not an effect, so a counter DOUBLER does
// nothing to an add-a-counter cost. Doubling Season must not make
// Devoted Druid's untapper cost two counters.
func TestAddCounterCostIsNotReplaceable(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, AbilityCost{
		AddCounter: &CounterAddCost{Counter: "-1/-1", N: 1},
	}, nil)
	g.mu.Lock()
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter && ev.CounterDelta > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta *= 2
			return nil
		},
		Label: "Doubling-Season-style",
	})
	g.mu.Unlock()

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterOf(g, src, "-1/-1"); got != 1 {
		t.Errorf("-1/-1 counters %d, want 1 — a cost is not a replaceable event", got)
	}
}

// CR 118.3's other direction: a permanent that cannot have the
// counter put on it cannot pay, and the refusal happens before
// anything else is spent. The engine models no counter prohibition
// yet, so the one reachable case is a source that is no longer on the
// battlefield under the payer's control — which is what
// canPlaceCounterLocked answers, and where a Solemnity-style static
// will plug in.
func TestAddCounterCostRefusedWhenItCannotBePlaced(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]
	src := pushCounterCostSource(g, me, AbilityCost{
		AddCounter: &CounterAddCost{Counter: "-1/-1", N: 1},
	}, nil)

	g.mu.Lock()
	ok := g.canPlaceCounterLocked(me.ID, src, &CounterAddCost{Counter: "-1/-1", N: 1})
	gone := g.canPlaceCounterLocked(me.ID, uuid.New(), &CounterAddCost{Counter: "-1/-1", N: 1})
	theirs := g.canPlaceCounterLocked(them.ID, src, &CounterAddCost{Counter: "-1/-1", N: 1})
	g.mu.Unlock()
	if !ok {
		t.Error("a permanent on the battlefield cannot pay its own add-a-counter cost")
	}
	if gone {
		t.Error("a permanent that is not there can pay an add-a-counter cost")
	}
	if theirs {
		t.Error("a permanent somebody else controls can pay your add-a-counter cost")
	}
}

// --- the paid-cost record ------------------------------------------

// Every counter component writes the ONE record on the stack item —
// which is what the ADR means by "design it once".
func TestCounterCostsAreRecordedOnTheStackItem(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, AbilityCost{
		RemoveCounters: &CounterRemovalCost{Counter: "charge", N: 2},
		AddCounter:     &CounterAddCost{Counter: "-1/-1", N: 1},
		Life:           3,
	}, nil)
	setCountersForTest(g, src, map[string]int{"charge": 3})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	var item *StackItem
	for _, it := range g.StackMeta {
		item = it
	}
	if item == nil {
		t.Fatal("nothing on the stack")
	}
	if item.Paid.CountersRemoved != 2 {
		t.Errorf("Paid.CountersRemoved = %d, want 2", item.Paid.CountersRemoved)
	}
	if item.Paid.CountersAdded != 1 {
		t.Errorf("Paid.CountersAdded = %d, want 1", item.Paid.CountersAdded)
	}
	if item.Paid.LifePaid != 3 {
		t.Errorf("Paid.LifePaid = %d, want 3", item.Paid.LifePaid)
	}
}

// itoaForTest keeps the produced-string helper above readable without
// pulling strconv into the test's import block for one call.
func itoaForTest(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// setCountersForTest stamps counters onto a battlefield card without
// going through the replacement pipeline.
func setCountersForTest(g *Game, id uuid.UUID, counters map[string]int) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Counters = counters
		}
	}
}
