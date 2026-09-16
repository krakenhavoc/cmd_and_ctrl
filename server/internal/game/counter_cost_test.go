package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// counter_cost_test.go — #625: AbilityCost.RemoveCounters, the
// "remove N counters" cost component, in its three printed shapes.
//
// The card that asked for it is Heart of Kiran ("remove a loyalty
// counter from a planeswalker you control rather than pay Heart of
// Kiran's crew cost"), and most of the rules pinned here are that
// card's: the walker is chosen as a cost rather than targeted, the
// removal is not a loyalty activation, a walker paid down to zero dies
// to the SBA, and the whole thing is instant speed. The catalog half
// (the real card files) is in cards/effects/counter_cost_cards_test.go.

// pwCostSpec is "a planeswalker" as a cost predicate — the in-package
// stand-in for effects.RemoveCountersFrom's TargetPermanent.
func pwCostSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "permanent",
		Label: "a planeswalker you control",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsPlaneswalker()
		},
		Min: 1, Max: 1,
	}
}

func creatureCostSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "permanent",
		Label: "a creature you control",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
}

// pushCounterCostSource puts an artifact on the battlefield whose one
// ability costs `cost` and, on resolution, stamps a marker counter on
// itself so a test can see the effect ran.
func pushCounterCostSource(g *Game, owner *Player, cost AbilityCost, targets *TargetSpec) uuid.UUID {
	c := NewCard("Counter Cost Source", owner.ID)
	c.TypeLine = "Artifact — Vehicle"
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label:   "remove counters: mark",
		Cost:    cost,
		Targets: targets,
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// pushPermanentWithCounters seats a permanent of the given type line
// with the given counters.
func pushPermanentWithCounters(g *Game, controller *Player, name, typeLine string, counters map[string]int) uuid.UUID {
	c := NewCard(name, controller.ID)
	c.TypeLine = typeLine
	c.Controller = controller.ID
	if strings.Contains(typeLine, "Creature") {
		c.Power, c.Toughness = 2, 2
	}
	c.Counters = counters
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func heartShapedCost() AbilityCost {
	return AbilityCost{RemoveCounters: &CounterRemovalCost{Counter: CounterLoyalty, N: 1, From: pwCostSpec()}}
}

func passBothForTest(g *Game) {
	_ = g.PassPriority()
	_ = g.PassPriority()
}

// The headline: the named planeswalker loses one loyalty counter at
// announce, the ability goes on the stack, and it resolves.
func TestCounterCostRemovesFromTheNamedPlaneswalker(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, heartShapedCost(), nil)
	a := pushLoyaltyWalker(g, me, 4, 1)
	b := pushLoyaltyWalker(g, me, 3, 1)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := loyaltyOf(g, b); got != 2 {
		t.Errorf("the named walker: loyalty %d, want 2", got)
	}
	if got := loyaltyOf(g, a); got != 4 {
		t.Errorf("the other walker was charged: loyalty %d, want 4", got)
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("StackMeta has %d items, want 1 (the cost is paid at announce, the effect waits)", len(g.StackMeta))
	}
	if counterOf(g, src, "effect-ran") != 0 {
		t.Error("the effect ran at announce")
	}
	passBothForTest(g)
	if counterOf(g, src, "effect-ran") != 1 {
		t.Error("the ability did not resolve")
	}
}

// Every way the other-permanent form is unpayable, and that none of
// them pays anything.
func TestCounterCostRejectsWhatCannotPay(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCounterCostSource(g, me, heartShapedCost(), nil)
	theirs := pushLoyaltyWalker(g, opp, 3, 1)
	notAWalker := pushPermanentWithCounters(g, me, "Loyal Artifact", "Artifact", map[string]int{CounterLoyalty: 3})
	empty := pushLoyaltyWalker(g, me, 0, 1)

	g.mu.RLock()
	opts := g.CounterCostOptionsForEffect(me.ID, src, heartShapedCost().RemoveCounters)
	g.mu.RUnlock()
	if len(opts) != 0 {
		t.Errorf("options with no payable walker of mine: %+v, want none", opts)
	}

	cases := []struct {
		name string
		ids  []uuid.UUID
		want error
	}{
		{"nothing named", nil, ErrInvalidParam},
		{"two named", []uuid.UUID{theirs, empty}, ErrInvalidParam},
		{"a card that does not exist", []uuid.UUID{uuid.New()}, ErrCardNotFound},
		{"an opponent's planeswalker", []uuid.UUID{theirs}, ErrCardCallerMismatch},
		{"a non-planeswalker holding loyalty counters", []uuid.UUID{notAWalker}, ErrIllegalTarget},
		{"a planeswalker with no loyalty", []uuid.UUID{empty}, ErrInsufficientCounters},
	}
	for _, tc := range cases {
		err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: tc.ids})
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
	if got := loyaltyOf(g, theirs); got != 3 {
		t.Errorf("the opponent's walker lost loyalty: %d", got)
	}
	if got := counterOf(g, notAWalker, CounterLoyalty); got != 3 {
		t.Errorf("the artifact lost counters: %d", got)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("a rejected activation put %d items on the stack", len(g.StackMeta))
	}
}

// ADR 0020 §3: validate everything, then pay everything. A bad target
// on the same activation must leave the counter where it is, and a bad
// counter source must leave the tap unpaid.
func TestCounterCostIsValidatedBeforeAnythingIsPaid(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	onlyWalkers := pwCostSpec()
	cost := heartShapedCost()
	cost.Tap = true
	src := pushCounterCostSource(g, me, cost, onlyWalkers)
	walker := pushLoyaltyWalker(g, me, 3, 1)
	bear := pushPermanentWithCounters(g, me, "Bear", "Creature — Bear", nil)

	err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{walker},
		Targets:          []TargetRef{{Kind: TargetCard, ID: bear}},
	})
	if !errors.Is(err, ErrIllegalTarget) {
		t.Fatalf("illegal target: err = %v, want ErrIllegalTarget", err)
	}
	if got := loyaltyOf(g, walker); got != 3 {
		t.Errorf("a rejected activation removed a counter: loyalty %d, want 3", got)
	}

	err = g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{uuid.New()},
		Targets:          []TargetRef{{Kind: TargetCard, ID: walker}},
	})
	if !errors.Is(err, ErrCardNotFound) {
		t.Fatalf("bad counter source: err = %v, want ErrCardNotFound", err)
	}
	if c := findBattlefieldCard(g, src); c == nil || c.Tapped {
		t.Error("a rejected activation tapped the source")
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("a rejected activation put %d items on the stack", len(g.StackMeta))
	}
}

// CR 704.5i: a walker paid down to 0 is put into its owner's graveyard
// by the state-based check at the end of the activation — with the
// ability still on the stack, which resolves anyway.
func TestCounterCostCanKillTheWalkerThroughTheSBA(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, heartShapedCost(), nil)
	walker := pushLoyaltyWalker(g, me, 1, 1)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{walker}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(walker) {
		t.Error("a 0-loyalty planeswalker survived the 704.5i SBA")
	}
	if !me.Graveyard.Contains(walker) {
		t.Error("the walker did not reach its owner's graveyard")
	}
	passBothForTest(g)
	if counterOf(g, src, "effect-ran") != 1 {
		t.Error("the ability did not resolve after its cost killed the walker")
	}
}

// Removing a loyalty counter to pay another permanent's cost is not
// activating a loyalty ability (CR 606): the once-per-turn flag stays
// clear and the walker's own loyalty ability is still available.
func TestCounterCostIsNotALoyaltyActivation(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, heartShapedCost(), nil)
	walker := pushLoyaltyWalker(g, me, 3, 1)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{walker}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.LoyaltyActivatedThisTurn[walker] {
		t.Fatal("the counter cost burned the walker's once-per-turn loyalty activation")
	}
	passBothForTest(g)
	if err := g.ActivateCatalogAbility(me.ID, walker, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("the walker's own +1 after paying a counter cost: %v", err)
	}
	if got := loyaltyOf(g, walker); got != 3 {
		t.Errorf("loyalty %d, want 3 (3 − 1 for the cost + 1 for the ability)", got)
	}
}

// Crew is instant speed, and so is its counter-removal alternative: on
// the opponent's turn, in their upkeep, with nothing about sorcery
// timing open.
func TestCounterCostIsInstantSpeed(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, heartShapedCost(), nil)
	walker := pushLoyaltyWalker(g, me, 3, 1)
	advanceToStepOfSeat(t, g, 1, StepUpkeep)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{walker}}); err != nil {
		t.Fatalf("activate on the opponent's upkeep: %v", err)
	}
	if got := loyaltyOf(g, walker); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

// Paying a cost is not an effect (CR 614.1): a replacement that would
// double every counter change does not touch the removal.
func TestCounterCostIsNotReplaceable(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, heartShapedCost(), nil)
	walker := pushLoyaltyWalker(g, me, 5, 1)

	g.mu.Lock()
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta *= 2
			return nil
		},
		Label: "double every counter change",
	})
	g.mu.Unlock()

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{walker}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := loyaltyOf(g, walker); got != 4 {
		t.Errorf("loyalty %d, want 4 — exactly the printed one counter comes off", got)
	}
}

// Choosing a permanent to pay a cost does not target it, so shroud —
// the keyword that stops even its controller targeting it — is no
// obstacle. The same rule as sacrifice and crew.
func TestCounterCostIgnoresShroud(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, heartShapedCost(), nil)
	walker := pushLoyaltyWalker(g, me, 2, 1)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == walker {
			g.Battlefield.Cards[i].Keywords = []string{"shroud"}
		}
	}

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{walker}}); err != nil {
		t.Fatalf("a shrouded walker of mine pays: %v", err)
	}
	g.mu.RLock()
	c := findBattlefieldCard(g, walker)
	targetable := c != nil && CanBeTargetedBy(c, ZoneBattlefield, me.ID)
	g.mu.RUnlock()
	if targetable {
		t.Fatal("test setup: the walker should be untargetable, or this test proves nothing")
	}
}

// The self form: no ID needed (or the source's own), the printed kind
// optional, and the source must hold enough.
func TestCounterCostSelfForm(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	cost := AbilityCost{RemoveCounters: &CounterRemovalCost{Counter: "gold", N: 1}}
	src := pushCounterCostSource(g, me, cost, nil)
	other := pushPermanentWithCounters(g, me, "Other Hoard", "Artifact", map[string]int{"gold": 5})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); !errors.Is(err, ErrInsufficientCounters) {
		t.Errorf("no gold counters: err = %v, want ErrInsufficientCounters", err)
	}
	g.mu.Lock()
	_ = g.applyCounterLocked(src, "gold", 2)
	g.mu.Unlock()

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{other}}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("naming a different permanent for a self cost: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterKind: "charge"}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("naming a different kind than printed: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("self form with nothing named: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{src}, CounterKind: "gold"}); err != nil {
		t.Fatalf("self form naming itself and its printed kind: %v", err)
	}
	if got := counterOf(g, src, "gold"); got != 0 {
		t.Errorf("gold after two activations: %d, want 0", got)
	}
	if got := counterOf(g, other, "gold"); got != 5 {
		t.Errorf("the other permanent paid: %d", got)
	}
}

// The any-kind form: the kind is named at announce and must be on the
// named permanent.
func TestCounterCostAnyKindForm(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	cost := AbilityCost{RemoveCounters: &CounterRemovalCost{N: 1, From: creatureCostSpec()}}
	src := pushCounterCostSource(g, me, cost, nil)
	bear := pushPermanentWithCounters(g, me, "Bear", "Creature — Bear", map[string]int{CounterPlusOne: 1, CounterStun: 2})
	bare := pushPermanentWithCounters(g, me, "Bare Bear", "Creature — Bear", nil)

	g.mu.RLock()
	opts := g.CounterCostOptionsForEffect(me.ID, src, cost.RemoveCounters)
	g.mu.RUnlock()
	if len(opts) != 1 || opts[0].CardID != bear {
		t.Fatalf("options: %+v, want only the bear", opts)
	}
	if len(opts[0].Kinds) != 2 || opts[0].Kinds[0].Kind != CounterStun || opts[0].Kinds[1].Kind != CounterPlusOne {
		t.Errorf("kinds: %+v, want stun (2) then +1/+1 (1)", opts[0].Kinds)
	}

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{bear}}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("any kind with no kind named: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{bear}, CounterKind: CounterMinusOne}); !errors.Is(err, ErrInsufficientCounters) {
		t.Errorf("a kind the bear does not have: err = %v, want ErrInsufficientCounters", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{bare}, CounterKind: CounterPlusOne}); !errors.Is(err, ErrInsufficientCounters) {
		t.Errorf("a creature with no counters: err = %v, want ErrInsufficientCounters", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{bear}, CounterKind: CounterStun}); err != nil {
		t.Fatalf("any kind naming stun: %v", err)
	}
	if got, plus := counterOf(g, bear, CounterStun), counterOf(g, bear, CounterPlusOne); got != 1 || plus != 1 {
		t.Errorf("after removing a stun counter: stun %d (want 1), +1/+1 %d (want 1)", got, plus)
	}
}

// A payload that names a counter source or kind for an ability with no
// counter component is a confused client, and is refused rather than
// ignored — the crew_ids / sacrifice_ids rule.
func TestCounterParamsAreRejectedWithoutACounterCost(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, AbilityCost{}, nil)
	walker := pushLoyaltyWalker(g, me, 3, 1)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{walker}}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("counter_source_ids on a free ability: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterKind: CounterLoyalty}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("counter_kind on a free ability: err = %v, want ErrInvalidParam", err)
	}
	if got := loyaltyOf(g, walker); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

// The option walk orders the most counters first, which is the order
// the client lists and the enumerator spends its budget in.
func TestCounterCostOptionsAreMostCountersFirst(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushCounterCostSource(g, me, heartShapedCost(), nil)
	small := pushLoyaltyWalker(g, me, 1, 1)
	big := pushLoyaltyWalker(g, me, 6, 1)
	mid := pushLoyaltyWalker(g, me, 3, 1)

	g.mu.RLock()
	opts := g.CounterCostOptionsForEffect(me.ID, src, heartShapedCost().RemoveCounters)
	g.mu.RUnlock()
	want := []uuid.UUID{big, mid, small}
	if len(opts) != len(want) {
		t.Fatalf("options: %+v, want three", opts)
	}
	for i, id := range want {
		if opts[i].CardID != id {
			t.Errorf("option %d = %v, want %v", i, opts[i].CardID, id)
		}
	}
}

// The #683 tests below reach the toughness state check through the
// counter-removal cost. pushZeroZero and pushStarCreature, and the
// same rules reached through develop's own roads (the add_counter
// action, an effect, the 704.5q cancel), are in zero_toughness_test.go.

// A printed 0/0 that pays a cost with its LAST +1/+1 counter has a real
// toughness of 0 and is put into its owner's graveyard (CR 704.5f) —
// not skipped as a placeholder because it now has no counters. Before
// Card.LostLastCounter it stayed on the battlefield: Fain, the Broker
// spending a Hangarback Walker's counter left an unkillable 0/0.
func TestCounterCostSpendingAZeroZerosLastCounterKillsIt(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	cost := AbilityCost{RemoveCounters: &CounterRemovalCost{N: 1, From: creatureCostSpec()}}
	src := pushCounterCostSource(g, me, cost, nil)
	walker := pushZeroZero(g, me, map[string]int{CounterPlusOne: 1})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{walker}, CounterKind: CounterPlusOne}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(walker) {
		t.Fatal("a 0/0 that spent its last +1/+1 counter stayed on the battlefield")
	}
	if !me.Graveyard.Contains(walker) {
		t.Error("the 0/0 did not reach its owner's graveyard")
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == walker && c.LostLastCounter {
			t.Error("the flag belongs to the permanent; it must not follow the card to the graveyard")
		}
	}
	passBothForTest(g)
	if counterOf(g, src, "effect-ran") != 1 {
		t.Error("the ability did not resolve after its cost killed the 0/0")
	}
}

// The self form, Mikaeus-shaped: the 0/0 source pays with its own last
// counter and dies with its ability on the stack, which still resolves.
func TestCounterCostSelfFormOnAZeroZeroKillsTheSource(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushZeroZero(g, me, map[string]int{CounterPlusOne: 1})
	bear := pushPermanentWithCounters(g, me, "Bear", "Creature — Bear", nil)
	g.mu.Lock()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == src {
			g.Battlefield.Cards[i].ActivatedAbilities = []ActivatedAbilityShape{{
				Label: "remove a +1/+1 counter: pump the bear",
				Cost:  AbilityCost{RemoveCounters: &CounterRemovalCost{Counter: CounterPlusOne, N: 1}},
				Effect: func(g *Game, _ *StackItem) error {
					return g.applyCounterLocked(bear, CounterPlusOne, 1)
				},
			}}
		}
	}
	g.mu.Unlock()

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(src) {
		t.Fatal("a 0/0 that paid its own cost with its last counter stayed on the battlefield")
	}
	passBothForTest(g)
	if counterOf(g, bear, CounterPlusOne) != 1 {
		t.Error("the ability did not resolve after its source died paying for it")
	}
}

// Paying a cost does not end the placeholder convention by itself: a
// Toughness 0 creature that never had a counter is still skipped, and
// so is a 0/0 that still has counters left after paying.
func TestCounterCostLeavesPlaceholdersAndZeroZerosWithCountersLeft(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	cost := AbilityCost{RemoveCounters: &CounterRemovalCost{Counter: CounterPlusOne, N: 1, From: creatureCostSpec()}}
	src := pushCounterCostSource(g, me, cost, nil)
	placeholder := pushZeroZero(g, me, nil)
	big := pushZeroZero(g, me, map[string]int{CounterPlusOne: 2})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{big}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Battlefield.Contains(placeholder) {
		t.Error("a Toughness 0 creature that never had counters was killed — the placeholder skip regressed")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("a 0/0 with a +1/+1 counter left died")
	}
}

// A `*` creature's 0 toughness is the import stand-in, so paying a
// cost with its last counter does not end the placeholder skip
// (Card.VariableToughness): it stays on the battlefield.
func TestCounterCostOnAVariableToughnessCreatureLeavesItInPlace(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	cost := AbilityCost{RemoveCounters: &CounterRemovalCost{N: 1, From: creatureCostSpec()}}
	src := pushCounterCostSource(g, me, cost, nil)
	star := pushStarCreature(g, me, map[string]int{CounterPlusOne: 1})
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{star}, CounterKind: CounterPlusOne}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Battlefield.Contains(star) {
		t.Error("a `*` creature died paying a cost with its last counter")
	}
}
