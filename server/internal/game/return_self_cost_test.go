package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// return_self_cost_test.go — #2028: AbilityCost.ReturnSelf, "Return this
// enchantment to its owner's hand:" (Gossamer Chains, Shigeki). The card
// rows live in cards/effects/return_this_cost_test.go; the CR 903.9b rows
// are in cost_commander_choice_test.go's table.

// returnSelfAbility is seedCostCard's edit for a permanent whose one
// ability costs "Return this to its owner's hand".
func returnSelfAbility(c *Card) {
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label:  "Return this to its owner's hand: nothing.",
		Cost:   AbilityCost{ReturnSelf: true},
		Effect: noopEffect,
	}}
}

// seedReturnSelf puts a 3/3 creature with a return-this ability on the
// battlefield, owned by `owner` and controlled by `controller`, with
// `effect` as the ability's effect.
func seedReturnSelf(g *Game, owner, controller uuid.UUID, effect func(*Game, *StackItem) error) uuid.UUID {
	c := NewCard("Bouncer", owner)
	c.TypeLine = "Creature — Shapeshifter"
	c.Power, c.Toughness = 3, 3
	c.Controller = controller
	c.SummonedThisTurn = false
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label:  "Return this creature to its owner's hand: remember it.",
		Cost:   AbilityCost{ReturnSelf: true},
		Effect: effect,
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// The cost is paid at announce: the permanent is in its owner's hand and
// the ability is on the stack before anyone can respond (CR 602.2b,
// CR 601.2h).
func TestAReturnThisCostPutsThePermanentInItsOwnersHand(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := seedReturnSelf(g, me.ID, me.ID, noopEffect)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Hand.Contains(src) {
		t.Fatal("the source is not in its owner's hand after the activation")
	}
	if g.Battlefield.Contains(src) {
		t.Fatal("the source is still on the battlefield")
	}
	if item := onlyAbilityOnStack(g); item == nil || item.SourceCardID != src {
		t.Fatalf("ability on the stack = %+v, want one from the returned source", item)
	}
}

// "Its OWNER's hand": a permanent gained from an opponent goes back to
// that opponent's hand, not the activator's.
func TestAReturnThisCostOnAStolenPermanentGoesToItsOwner(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := seedReturnSelf(g, opp.ID, me.ID, noopEffect)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !opp.Hand.Contains(src) {
		t.Fatal("the stolen permanent is not in its owner's hand")
	}
	if me.Hand.Contains(src) {
		t.Fatal("the stolen permanent went to the activator's hand")
	}
}

// The card in the hand is a new object (CR 400.7). The effect reads
// "this permanent" through the item's source object, stamped before the
// payment, and gets the permanent's last-known information (CR 608.2h).
func TestAReturnThisCostLeavesTheEffectTheReturnedPermanentsLKI(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	var seen PermanentInfo
	var seenOK bool
	src := seedReturnSelf(g, me.ID, me.ID, func(g *Game, item *StackItem) error {
		if ref, ok := g.SourceObjectForEffect(item); ok {
			seen, seenOK = g.PermanentForEffect(ref)
		}
		return nil
	})
	g.WithWriteLock(func() { findBattlefieldCard(g, src).Counters = map[string]int{CounterPlusOne: 2} })
	want := refOf(t, g, src)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	item := onlyAbilityOnStack(g)
	if item == nil || item.SourceObject != want {
		t.Fatalf("activated item = %+v, want SourceObject %+v (pre-cost)", item, want)
	}
	for i := 0; i < len(g.Seats) && onlyAbilityOnStack(g) != nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if !seenOK || !seen.Left {
		t.Fatalf("at resolution the source reads %+v, %v; want the returned permanent's record", seen, seenOK)
	}
	if seen.Power != 5 || seen.Toughness != 5 {
		t.Errorf("last-known P/T = %d/%d, want 5/5 (3/3 with two +1/+1 counters)", seen.Power, seen.Toughness)
	}
}

// Only a permanent can be returned: the ability does not function from a
// hand, so it cannot be activated there (CR 113.6), and the runtime check
// refuses any zone but the battlefield.
func TestAReturnThisCostNeedsThePermanentOnTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	c := NewCard("Bouncer", me.ID)
	c.TypeLine = "Creature — Shapeshifter"
	returnSelfAbility(&c)
	me.Hand.PushTop(c)

	err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{})
	if !errors.Is(err, ErrActivationZoneNotAllowed) {
		t.Fatalf("activate from hand: err = %v, want ErrActivationZoneNotAllowed", err)
	}
	if onlyAbilityOnStack(g) != nil {
		t.Fatal("an ability reached the stack")
	}
	for _, z := range []ZoneKind{ZoneHand, ZoneGraveyard, ZoneExile, ZoneCommand} {
		if err := validateReturnSelfCostLocked(c.InstanceID, z, AbilityCost{ReturnSelf: true}); !errors.Is(err, ErrActivationZoneNotAllowed) {
			t.Errorf("validate from %s: err = %v, want ErrActivationZoneNotAllowed", z, err)
		}
	}
	if why := AbilityNeedsPermanentSource(AbilityCost{ReturnSelf: true}); why == "" {
		t.Error("AbilityNeedsPermanentSource does not name the return-this cost")
	}
}

// The source pays one component (CR 118.3): naming it again to another
// component that moves a permanent is refused with nothing paid.
func TestAReturnThisSourceCannotPayAnotherComponent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cost   AbilityCost
		params func(src uuid.UUID) ActivateAbilityParams
	}{
		{"return a creature", AbilityCost{ReturnSelf: true, ReturnToHand: &ReturnToHandCost{Count: 1, Filter: creatureCostSpec(), Label: "a creature you control"}},
			func(src uuid.UUID) ActivateAbilityParams { return ActivateAbilityParams{ReturnIDs: []uuid.UUID{src}} }},
		{"sacrifice a creature", AbilityCost{ReturnSelf: true, SacrificeOther: creatureCostSpec()},
			func(src uuid.UUID) ActivateAbilityParams {
				return ActivateAbilityParams{SacrificeIDs: []uuid.UUID{src}}
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			src := seedReturnSelf(g, me.ID, me.ID, noopEffect)
			findBattlefieldCard(g, src).ActivatedAbilities[0].Cost = tc.cost

			if err := g.ActivateCatalogAbility(me.ID, src, 0, tc.params(src)); err == nil {
				t.Fatal("the source paid two components")
			}
			if !g.Battlefield.Contains(src) {
				t.Fatal("a refused activation moved the source")
			}
			if onlyAbilityOnStack(g) != nil {
				t.Fatal("a refused activation reached the stack")
			}
		})
	}
	if err := validateReturnSelfCostLocked(uuid.New(), ZoneBattlefield, AbilityCost{ReturnSelf: true, SacrificeSelf: true}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("return-this beside sacrifice-this: err = %v, want ErrInvalidParam", err)
	}
}

// The auto-tapper may not crack the source for the mana half of its own
// return-this ability: the return would then find nothing to pay with.
func TestAReturnThisSourceIsNotPlannedForMana(t *testing.T) {
	src := uuid.New()
	if !AbilityAutoTapExclusions(src, AbilityCost{ReturnSelf: true, Mana: "{1}"}, nil, nil, nil, nil)[src] {
		t.Error("the source of a return-this cost is not excluded from the auto-tap plan")
	}
}
