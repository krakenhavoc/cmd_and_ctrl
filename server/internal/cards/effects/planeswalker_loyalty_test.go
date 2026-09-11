package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planeswalker_loyalty_test.go — issues #329 / #334. Before this,
// no catalog card declared a loyalty ability and none could: the
// cost had nowhere to live. These tests are the end-to-end proof
// that a loyalty ability is now an ordinary CR 602 activation.

const (
	teferiTimeRavelerOracle = "ae7604bb-4818-45a3-960c-cf3d83f15964"
	wanderingEmperorOracle  = "0c7f18d5-36cb-4bc6-a358-443b97666215"
)

// pushCatalogWalker seats a catalog planeswalker with `loyalty`
// loyalty counters under the active seat's control.
func pushCatalogWalker(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Legendary Planeswalker — " + name,
		OracleID:   oracle,
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{game.CounterLoyalty: loyalty},
	})
	return id
}

func toMain(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func loyaltyCount(g *game.Game, id uuid.UUID) int {
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterLoyalty]
}

// The catalog declares loyalty abilities at all — the thing that was
// impossible before AbilityCost grew a Loyalty component.
func TestLoyaltyAbilitiesAreWired(t *testing.T) {
	for _, tc := range []struct {
		oracle string
		name   string
		want   []int
	}{
		{teferiTimeRavelerOracle, "Teferi, Time Raveler", []int{1, -3}},
		{wanderingEmperorOracle, "The Wandering Emperor", []int{1, -1, -2}},
	} {
		abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})
		if len(abilities) != len(tc.want) {
			t.Fatalf("%s: %d abilities, want %d", tc.name, len(abilities), len(tc.want))
		}
		for i, want := range tc.want {
			cost := abilities[i].Cost.Loyalty
			if cost == nil {
				t.Errorf("%s ability %d has no loyalty cost", tc.name, i)
				continue
			}
			if *cost != want {
				t.Errorf("%s ability %d: loyalty cost %d, want %d", tc.name, i, *cost, want)
			}
		}
	}
}

// #334's card, end to end: −3 pays three loyalty, bounces the
// target and draws.
func TestTeferiMinusThreeBouncesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Time Raveler", teferiTimeRavelerOracle, 4)
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Victim")
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate -3: %v", err)
	}
	// CR 601/602: the cost is paid at announce, the effect waits.
	if got := loyaltyCount(g, teferi); got != 1 {
		t.Errorf("loyalty after -3 on a 4-loyalty Teferi: got %d, want 1", got)
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("the bounce happened at announce; it belongs at resolution")
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("the -3 did not bounce its target")
	}
	if !opp.Hand.Contains(victim) {
		t.Error("the bounced creature did not reach its OWNER's hand")
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand size %d -> %d, want +1 from the draw", handBefore, got)
	}
}

// "Up to one target" means the −3 is activatable with nothing
// picked — an empty board still draws a card. This is the whole
// reason the clause carries Min 0.
func TestTeferiMinusThreeTargetsNothingAndStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Time Raveler", teferiTimeRavelerOracle, 4)
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate -3 with no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand size %d -> %d, want +1 from the draw", handBefore, got)
	}
}

// CR 606.3 on a real card: Teferi at 2 loyalty cannot −3.
func TestTeferiCannotMinusThreeBelowCost(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Time Raveler", teferiTimeRavelerOracle, 2)

	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{}); err != game.ErrInsufficientLoyalty {
		t.Errorf("-3 at 2 loyalty: got %v, want ErrInsufficientLoyalty", err)
	}
	// The +1 is still available: a refused activation doesn't spend
	// the turn's window.
	if err := g.ActivateCatalogAbility(me.ID, teferi, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1 after a refused -3: %v", err)
	}
	if got := loyaltyCount(g, teferi); got != 3 {
		t.Errorf("loyalty after +1: got %d, want 3", got)
	}
}

// CR 606.5 on a real card, across both of Teferi's abilities: one
// activation per planeswalker per turn, not one per ability.
func TestTeferiOneLoyaltyAbilityPerTurn(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Time Raveler", teferiTimeRavelerOracle, 4)

	if err := g.ActivateCatalogAbility(me.ID, teferi, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{}); err != game.ErrLoyaltyAlreadyActivated {
		t.Errorf("-3 after +1 on the same turn: got %v, want ErrLoyaltyAlreadyActivated", err)
	}
}

// The Emperor's +1 makes a token; the loyalty cost is paid up.
func TestWanderingEmperorPlusOneMakesASamurai(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	emperor := pushCatalogWalker(g, me.ID, "The Wandering Emperor", wanderingEmperorOracle, 3)

	if err := g.ActivateCatalogAbility(me.ID, emperor, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate +1: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := loyaltyCount(g, emperor); got != 4 {
		t.Errorf("loyalty after +1: got %d, want 4", got)
	}
	samurai := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Samurai" && c.Controller == me.ID {
			samurai++
			if !game.HasKeyword(&c, "vigilance") {
				t.Error("the Samurai token has no vigilance")
			}
		}
	}
	if samurai != 1 {
		t.Errorf("got %d Samurai tokens, want 1", samurai)
	}
}

// The Emperor's −1 exiles a TAPPED creature, and only a tapped one:
// an untapped creature is not a legal target (CR 115.4).
func TestWanderingEmperorMinusOneExilesOnlyTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	emperor := pushCatalogWalker(g, me.ID, "The Wandering Emperor", wanderingEmperorOracle, 3)
	untapped := pushCreatureToBattlefieldForTest(g, opp.ID, "Untapped")
	tapped := pushCreatureToBattlefieldForTest(g, opp.ID, "Tapped")
	if err := g.TapCard(tapped, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}

	if err := g.ActivateCatalogAbility(me.ID, emperor, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: untapped}},
	}); err != game.ErrIllegalTarget {
		t.Errorf("-1 at an untapped creature: got %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, emperor, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tapped}},
	}); err != nil {
		t.Fatalf("-1 at a tapped creature: %v", err)
	}
	if got := loyaltyCount(g, emperor); got != 2 {
		t.Errorf("loyalty after -1: got %d, want 2", got)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature was not exiled")
	}
}

// The Emperor's −2 is two turn-scoped statics: +2/+1 in layer 7c
// and lifelink in layer 6.
func TestWanderingEmperorMinusTwoPumpsAndGrantsLifelink(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	emperor := pushCatalogWalker(g, me.ID, "The Wandering Emperor", wanderingEmperorOracle, 3)
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "Bear")

	if err := g.ActivateCatalogAbility(me.ID, emperor, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate -2: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := loyaltyCount(g, emperor); got != 1 {
		t.Errorf("loyalty after -2: got %d, want 1", got)
	}
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("P/T after +2/+1 on a 2/2: got %d/%d, want 4/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the -2 did not grant lifelink")
	}
}
