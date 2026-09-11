package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loyalty_view_test.go — the wire half of #329 / #334. The client
// cannot grey a loyalty row it can't see the cost of, and it cannot
// enforce CR 606.5 against state the server keeps to itself.

func loyaltyPtr(n int) *int { return &n }

// seatLoyaltyWalker puts a planeswalker with one +1 and one −3
// ability on the battlefield under the first seat.
func seatLoyaltyWalker(g *game.Game, loyalty int) (uuid.UUID, uuid.UUID) {
	owner := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Walker",
		TypeLine:   "Legendary Planeswalker — Test",
		// Non-empty so stampActivatedAbilities doesn't skip it;
		// intrinsic abilities win over the catalog lookup.
		OracleID:   "00000000-0000-0000-0000-0000000000aa",
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{game.CounterLoyalty: loyalty},
		ActivatedAbilities: []game.ActivatedAbilityShape{
			{Label: "+1: do a thing", Cost: game.AbilityCost{Loyalty: loyaltyPtr(1)}},
			{
				Label:   "−3: bounce up to one",
				Cost:    game.AbilityCost{Loyalty: loyaltyPtr(-3)},
				Targets: &game.TargetSpec{Mode: "permanent", Label: "up to one target permanent", Zones: []game.ZoneKind{game.ZoneBattlefield}, Min: 0, Max: 1},
			},
		},
	})
	return id, owner
}

func walkerView(t *testing.T, g *game.Game, id uuid.UUID) CardView {
	t.Helper()
	v := ViewOfGame(g)
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("planeswalker %s missing from the battlefield view", id)
	return CardView{}
}

func TestActivatedAbilityViewCarriesLoyaltyCost(t *testing.T) {
	g := buildActiveGame(t)
	id, _ := seatLoyaltyWalker(g, 4)
	c := walkerView(t, g, id)

	if len(c.ActivatedAbilities) != 2 {
		t.Fatalf("got %d abilities on the wire, want 2", len(c.ActivatedAbilities))
	}
	plus, minus := c.ActivatedAbilities[0], c.ActivatedAbilities[1]
	if plus.LoyaltyCost == nil || *plus.LoyaltyCost != 1 {
		t.Errorf("+1 loyalty_cost: got %v, want 1", plus.LoyaltyCost)
	}
	if minus.LoyaltyCost == nil || *minus.LoyaltyCost != -3 {
		t.Errorf("−3 loyalty_cost: got %v, want -3", minus.LoyaltyCost)
	}
	// CR 606.5 rides the loyalty component, so a card doesn't have
	// to remember to declare SorcerySpeed — but the client greys on
	// the flag, so the view stamps it.
	if !plus.SorcerySpeed || !minus.SorcerySpeed {
		t.Error("a loyalty ability should surface as sorcery-speed")
	}
}

// "Up to one target" only works if the count reaches the client:
// the picker needs Min 0 to let the player confirm with nothing
// selected. abilityLegalTargets never copied Min / Max across, so
// every ability clause shipped as 0 / 0 — "unbounded".
func TestAbilityLegalTargetsCarryTheCount(t *testing.T) {
	g := buildActiveGame(t)
	id, _ := seatLoyaltyWalker(g, 4)
	c := walkerView(t, g, id)

	lt := c.ActivatedAbilities[1].LegalTargets
	if lt == nil {
		t.Fatal("the −3 clause has no legal-target view")
	}
	if lt.Min != 0 || lt.Max != 1 {
		t.Errorf("up-to-one clause count: got %d..%d, want 0..1", lt.Min, lt.Max)
	}
}

// CR 606.5's once-per-turn flag was server-only, which is why
// canActivateLoyalty had to take the client's guess as an argument.
func TestCardViewReportsLoyaltyActivatedThisTurn(t *testing.T) {
	g := buildActiveGame(t)
	id, owner := seatLoyaltyWalker(g, 4)

	if walkerView(t, g, id).LoyaltyActivated {
		t.Error("loyalty_activated set before anything was activated")
	}
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(owner, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if !walkerView(t, g, id).LoyaltyActivated {
		t.Error("loyalty_activated not reported after an activation")
	}
}
