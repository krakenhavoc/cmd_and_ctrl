package testsl

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// loyalty_cost_add_test.go — #6148: loyalty ability cost modifications.
// Carth the Lion's "Planeswalkers' loyalty abilities you activate cost an
// additional [+1] to activate" combines with the printed loyalty cost (CR
// 606.5) and CR 606.6 judges the combined cost; Eidolon of Obstruction's
// "Loyalty abilities of planeswalkers your opponents control cost {1} more
// to activate" adds mana beside it (CR 606.4).

const (
	carthOracle     = "e0709843-0a91-4a21-9512-d7481707f3c9"
	eidolonOracle   = "15d38195-b661-42c1-82c9-1af2a3ba2ea5"
	testWalkerKey   = "test-6148-walker"
	testWalkerName  = "Test Walker"
	testWalkerTypes = "Legendary Planeswalker — Test"
)

// registerTestWalker registers a planeswalker with a +1, a 0, a −2 and a
// −3, each doing nothing.
func registerTestWalker(t *testing.T) {
	t.Helper()
	nothing := func(*game.Game, *game.StackItem) error { return nil }
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: testWalkerKey, Name: testWalkerName, Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{Label: "+1: Nothing.", Cost: effects.LoyaltyCost(1), Effect: nothing},
			{Label: "0: Nothing.", Cost: effects.LoyaltyCost(0), Effect: nothing},
			{Label: "−2: Nothing.", Cost: effects.LoyaltyCost(-2), Effect: nothing},
			{Label: "−3: Nothing.", Cost: effects.LoyaltyCost(-3), Effect: nothing},
		},
	})
}

func seedTestWalker(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: testWalkerName, TypeLine: testWalkerTypes, OracleID: testWalkerKey,
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
	return id
}

func loyaltyOn(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[game.CounterLoyalty]
			}
		}
	})
	return n
}

// CR 606.5: under Carth, +1 is +2, 0 is +1 and −2 is −1.
func TestCarthAddsOneLoyaltyToEachLoyaltyCost(t *testing.T) {
	registerTestWalker(t)
	for _, tc := range []struct {
		index int
		want  int
	}{{0, 6}, {1, 5}, {2, 3}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.PushCatalogPermanent(g, me.ID, "Carth the Lion", "Legendary Creature — Human Warrior", carthOracle, false)
		walker := seedTestWalker(g, me.ID, 4)
		if err := g.ActivateCatalogAbility(me.ID, walker, tc.index, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("ability %d: %v", tc.index, err)
		}
		if got := loyaltyOn(g, walker); got != tc.want {
			t.Errorf("ability %d from 4 loyalty under Carth: loyalty %d, want %d", tc.index, got, tc.want)
		}
	}
}

// CR 606.6 judges the combined cost: a −3 under Carth is a −2, so it is
// activatable at 2 loyalty; without Carth it is not.
func TestCarthMakesAMinusThreeActivatableAtTwoLoyalty(t *testing.T) {
	registerTestWalker(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bare := seedTestWalker(g, me.ID, 2)
	if err := g.ActivateCatalogAbility(me.ID, bare, 3, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrInsufficientLoyalty) {
		t.Fatalf("−3 at 2 loyalty without Carth: %v, want ErrInsufficientLoyalty", err)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Carth the Lion", "Legendary Creature — Human Warrior", carthOracle, false)
	// The enumerator offers it too, as the engine accepts it (#544).
	offered := map[int]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind != legal.KindActivate || m.Source != bare {
			continue
		}
		var p struct {
			Index int `json:"ability_index"`
		}
		if err := json.Unmarshal(m.Params, &p); err == nil {
			offered[p.Index] = true
		}
	}
	if !offered[3] || !offered[2] {
		t.Errorf("offered abilities %v at 2 loyalty under Carth, want the −2 and the −3", offered)
	}
	if err := g.ActivateCatalogAbility(me.ID, bare, 3, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−3 at 2 loyalty under Carth: %v", err)
	}
}

// "Loyalty abilities YOU activate": an opponent's Carth changes nothing,
// and the view shows the cost the controller would pay.
func TestCarthOnlyModifiesItsControllersLoyaltyAbilities(t *testing.T) {
	registerTestWalker(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, opp.ID, "Carth the Lion", "Legendary Creature — Human Warrior", carthOracle, false)
	mine := seedTestWalker(g, me.ID, 4)
	theirs := seedTestWalker(g, opp.ID, 4)

	for _, tc := range []struct {
		viewer, walker uuid.UUID
		want           int
	}{{me.ID, mine, 1}, {opp.ID, theirs, 2}} {
		row := effectstest.ActivatedRowOf(t, g, tc.viewer, tc.walker, 0)
		if row.LoyaltyCost == nil || *row.LoyaltyCost != tc.want {
			t.Errorf("the +1 shows loyalty cost %v, want %d", row.LoyaltyCost, tc.want)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	if got := loyaltyOn(g, mine); got != 5 {
		t.Errorf("my +1 under an opponent's Carth: loyalty %d, want 5", got)
	}
}

// Eidolon of Obstruction: an opponent's loyalty ability needs {1} beside
// its loyalty cost, and can't be activated without it; Eidolon's
// controller's own planeswalkers are untaxed.
func TestEidolonOfObstructionTaxesOpponentsLoyaltyAbilities(t *testing.T) {
	registerTestWalker(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, opp.ID, "Eidolon of Obstruction", "Enchantment Creature — Spirit", eidolonOracle, false)
	walker := seedTestWalker(g, me.ID, 4)

	if err := g.ActivateCatalogAbility(me.ID, walker, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("+1 with no mana under an opponent's Eidolon: %v, want ErrInsufficientMana", err)
	}
	if got := loyaltyOn(g, walker); got != 4 {
		t.Fatalf("a refused activation moved loyalty: %d", got)
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, walker, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1 with {1}: %v", err)
	}
	if got := loyaltyOn(g, walker); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}

	// The Eidolon's controller's own walker pays nothing extra.
	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g2, me2.ID, "Eidolon of Obstruction", "Enchantment Creature — Spirit", eidolonOracle, false)
	own := seedTestWalker(g2, me2.ID, 4)
	if err := g2.ActivateCatalogAbility(me2.ID, own, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1 under my own Eidolon: %v", err)
	}
}

// "Whenever Carth enters or a planeswalker you control dies": both look at
// the top seven for a planeswalker card.
func TestCarthLooksForAPlaneswalkerWhenAPlaneswalkerYouControlDies(t *testing.T) {
	registerTestWalker(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pw := effectstest.SeedLibraryTop(me, "Library Walker", "Legendary Planeswalker — Test", "")
	effectstest.PushCatalogPermanent(g, me.ID, "Carth the Lion", "Legendary Creature — Human Warrior", carthOracle, false)
	walker := seedTestWalker(g, me.ID, 2)
	// −2 at 2 loyalty is a −1 under Carth: use the −3 (a −2) to empty it.
	if err := g.ActivateCatalogAbility(me.ID, walker, 3, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no look-at-seven prompt after the planeswalker died")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != pw {
		t.Fatalf("offered %v, want the one planeswalker card", prompt.ChooseCards)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{pw}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(pw) {
		t.Error("the planeswalker card was not put into the hand")
	}
}
