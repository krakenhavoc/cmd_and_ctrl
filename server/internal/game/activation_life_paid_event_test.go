package game

import "testing"

// activation_life_paid_event_test.go — #5662: EventActivateAbility says
// how much life the activation paid, so a trigger that watches the
// activation can ask "if life was paid to activate it" and "that much
// life" (Verrak, Warped Sengir) without finding the ability's stack
// item, which may be gone by the time it resolves.

func TestActivationEventCarriesTheLifePaid(t *testing.T) {
	cases := []struct {
		name      string
		cost      AbilityCost
		pool      []ManaToken
		phyrexian int
		want      int
	}{
		{name: "a printed life component", cost: AbilityCost{Life: 3}, want: 3},
		{name: "a Phyrexian symbol paid with life (CR 107.4f)", cost: AbilityCost{Mana: "{1}{G/P}"},
			pool: []ManaToken{{Color: "C"}}, phyrexian: 1, want: PhyrexianLifePerSymbol},
		{name: "both, one total", cost: AbilityCost{Mana: "{G/P}", Life: 1},
			phyrexian: 1, want: PhyrexianLifePerSymbol + 1},
		{name: "a Phyrexian symbol paid with mana", cost: AbilityCost{Mana: "{1}{G/P}"},
			pool: []ManaToken{{Color: "C"}, {Color: "G"}}, want: 0},
		{name: "no life in the cost", cost: AbilityCost{Mana: "{1}"},
			pool: []ManaToken{{Color: "C"}}, want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			advanceTo(t, g, StepPrecombatMain)
			id := phyrexianPodSource(g, me, tc.cost)
			me.ManaPool.AddMana(tc.pool...)
			life := me.Life
			if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{
				Strict: true, PhyrexianLife: tc.phyrexian,
			}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			evs := activationEventsOf(g, EventActivateAbility)
			if len(evs) != 1 {
				t.Fatalf("%d activation events, want 1", len(evs))
			}
			if evs[0].LifePaid != tc.want {
				t.Errorf("Event.LifePaid = %d, want %d", evs[0].LifePaid, tc.want)
			}
			if paid := lastAbilityPaid(t, g); paid.LifePaid != evs[0].LifePaid {
				t.Errorf("the event says %d and the item's PaidCost %d", evs[0].LifePaid, paid.LifePaid)
			}
			if got := life - me.Life; got != tc.want {
				t.Errorf("life went down by %d, want %d", got, tc.want)
			}
		})
	}
}
