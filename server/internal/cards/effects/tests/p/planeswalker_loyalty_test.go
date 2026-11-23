package testsp

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A walker seated by the helper has the printed subtype, not name words.
func TestPushCatalogWalkerUsesTheRealSubtype(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for name, want := range map[string]string{
		"Jace, the Mind Sculptor":  "Jace",
		"Nicol Bolas, God-Pharaoh": "Bolas",
		"The Wandering Emperor":    "Emperor",
	} {
		id := effectstest.PushCatalogWalker(g, me.ID, name, "", 3)
		c, _ := g.LookupCardForEffect(id)
		subs := c.Effective().Subtypes
		if len(subs) != 1 || subs[0] != want {
			t.Errorf("%s: subtypes = %v, want [%s]", name, subs, want)
		}
	}
}

// The catalog declares loyalty abilities at all — the thing that was
// impossible before AbilityCost grew a Loyalty component.
func TestLoyaltyAbilitiesAreWired(t *testing.T) {
	for _, tc := range []struct {
		oracle string
		name   string
		want   []int
	}{
		{effectstest.TeferiTimeRavelerOracle, "Teferi, Time Raveler", []int{1, -3}},
		{effectstest.WanderingEmperorOracle, "The Wandering Emperor", []int{1, -1, -2}},
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
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	teferi := effectstest.PushCatalogWalker(g, me.ID, "Teferi, Time Raveler", effectstest.TeferiTimeRavelerOracle, 4)
	victim := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Victim")
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate -3: %v", err)
	}
	// CR 601/602: the cost is paid at announce, the effect waits.
	if got := effectstest.LoyaltyCount(g, teferi); got != 1 {
		t.Errorf("loyalty after -3 on a 4-loyalty Teferi: got %d, want 1", got)
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("the bounce happened at announce; it belongs at resolution")
	}
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := effectstest.PushCatalogWalker(g, me.ID, "Teferi, Time Raveler", effectstest.TeferiTimeRavelerOracle, 4)
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate -3 with no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand size %d -> %d, want +1 from the draw", handBefore, got)
	}
}

// CR 606.6 on a real card: Teferi at 2 loyalty cannot −3.
func TestTeferiCannotMinusThreeBelowCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := effectstest.PushCatalogWalker(g, me.ID, "Teferi, Time Raveler", effectstest.TeferiTimeRavelerOracle, 2)

	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{}); err != game.ErrInsufficientLoyalty {
		t.Errorf("-3 at 2 loyalty: got %v, want ErrInsufficientLoyalty", err)
	}
	// The +1 is still available: a refused activation doesn't spend
	// the turn's window.
	if err := g.ActivateCatalogAbility(me.ID, teferi, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1 after a refused -3: %v", err)
	}
	if got := effectstest.LoyaltyCount(g, teferi); got != 3 {
		t.Errorf("loyalty after +1: got %d, want 3", got)
	}
}

// CR 606.3 on a real card, across both of Teferi's abilities: one
// activation per planeswalker per turn, not one per ability.
func TestTeferiOneLoyaltyAbilityPerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := effectstest.PushCatalogWalker(g, me.ID, "Teferi, Time Raveler", effectstest.TeferiTimeRavelerOracle, 4)

	if err := g.ActivateCatalogAbility(me.ID, teferi, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{}); err != game.ErrLoyaltyAlreadyActivated {
		t.Errorf("-3 after +1 on the same turn: got %v, want ErrLoyaltyAlreadyActivated", err)
	}
}

// #1337: flash was missing from PrintedKeywords, which is real for
// an imported deck (Scryfall supplies the keyword there) but leaves
// a spawned or fixture Emperor without it.
func TestWanderingEmperorPrintsFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	emperor := effectstest.PushCatalogWalker(g, me.ID, "The Wandering Emperor", effectstest.WanderingEmperorOracle, 3)
	if !effectstest.HasEffectiveKeyword(t, g, emperor, "flash") {
		t.Error("The Wandering Emperor does not carry printed flash")
	}
}

// The Emperor's +1 puts a +1/+1 counter on up to one target creature
// and grants it first strike - two layers, one ability (#1208 fixed
// the card: S27 had the three abilities permuted and altered).
func TestWanderingEmperorPlusOneCountersAndFirstStrikes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	emperor := effectstest.PushCatalogWalker(g, me.ID, "The Wandering Emperor", effectstest.WanderingEmperorOracle, 3)
	bear := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Bear")

	if err := g.ActivateCatalogAbility(me.ID, emperor, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate +1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.LoyaltyCount(g, emperor); got != 4 {
		t.Errorf("loyalty after +1: got %d, want 4", got)
	}
	// A +1/+1 counter is read off the card (CurrentPower), not out of
	// the layer pass, so the counter itself is what this asserts.
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters after the +1: got %d, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the +1 did not grant first strike")
	}
}

// "Up to one target creature" is Min 0, so the +1 is activatable on
// an empty board and resolves doing nothing but the loyalty.
func TestWanderingEmperorPlusOneNeedsNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	emperor := effectstest.PushCatalogWalker(g, me.ID, "The Wandering Emperor", effectstest.WanderingEmperorOracle, 3)

	if err := g.ActivateCatalogAbility(me.ID, emperor, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate +1 with no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyCount(g, emperor); got != 4 {
		t.Errorf("loyalty after +1: got %d, want 4", got)
	}
}

// The Emperor's -1 makes the Samurai token; the loyalty cost is paid.
func TestWanderingEmperorMinusOneMakesASamurai(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	emperor := effectstest.PushCatalogWalker(g, me.ID, "The Wandering Emperor", effectstest.WanderingEmperorOracle, 3)

	if err := g.ActivateCatalogAbility(me.ID, emperor, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate -1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.LoyaltyCount(g, emperor); got != 2 {
		t.Errorf("loyalty after -1: got %d, want 2", got)
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

// The Emperor's -2 exiles a TAPPED creature - and only a tapped one
// (CR 115.4) - and gains 2 life.
func TestWanderingEmperorMinusTwoExilesTappedAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	emperor := effectstest.PushCatalogWalker(g, me.ID, "The Wandering Emperor", effectstest.WanderingEmperorOracle, 3)
	untapped := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Untapped")
	tapped := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Tapped")
	if err := g.TapCard(tapped, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	before := me.Life

	if err := g.ActivateCatalogAbility(me.ID, emperor, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: untapped}},
	}); err != game.ErrIllegalTarget {
		t.Errorf("-2 at an untapped creature: got %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, emperor, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tapped}},
	}); err != nil {
		t.Fatalf("-2 at a tapped creature: %v", err)
	}
	if got := effectstest.LoyaltyCount(g, emperor); got != 1 {
		t.Errorf("loyalty after -2: got %d, want 1", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature was not exiled")
	}
	// Exile, not destroy - the whole point of the ability.
	if !g.Exile.Contains(tapped) {
		t.Error("the tapped creature left the battlefield but was not exiled")
	}
	if got := me.Life - before; got != 2 {
		t.Errorf("life gained by the -2: got %d, want 2", got)
	}
}
