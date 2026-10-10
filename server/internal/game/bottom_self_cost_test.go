package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// bottom_self_cost_test.go — #2726: AbilityCost.BottomSelf, "Put this
// creature on the bottom of its owner's library:" (Timestream
// Navigator). ReturnSelf's twin; the card rows live in
// cards/effects/timestream_navigator_test.go.

// seedBottomSelf puts a 3/3 creature with a put-this-on-the-bottom
// ability on the battlefield, owned by `owner` and controlled by
// `controller`.
func seedBottomSelf(g *Game, owner, controller uuid.UUID, effect func(*Game, *StackItem) error) uuid.UUID {
	c := NewCard("Navigator", owner)
	c.TypeLine = "Creature — Human"
	c.Power, c.Toughness = 3, 3
	c.Controller = controller
	c.SummonedThisTurn = false
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label:  "Put this creature on the bottom of its owner's library: remember it.",
		Cost:   AbilityCost{BottomSelf: true},
		Effect: effect,
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// Paid at announce (CR 602.2b): the permanent is the bottom card of its
// owner's library and the ability is on the stack before anyone can
// respond, and the exit is a real leave (EventLTB).
func TestABottomSelfCostPutsThePermanentOnTheBottomOfTheLibrary(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := seedBottomSelf(g, me.ID, me.ID, noopEffect)
	size := len(me.Library.Cards)
	ltb := countEventsOfKind(g, EventLTB)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(src) {
		t.Fatal("the source is still on the battlefield")
	}
	if len(me.Library.Cards) != size+1 || me.Library.Cards[0].InstanceID != src {
		t.Fatalf("library has %d cards, bottom %v; want %d with the source at the bottom", len(me.Library.Cards), me.Library.Cards[0].InstanceID, size+1)
	}
	if got := countEventsOfKind(g, EventLTB); got != ltb+1 {
		t.Errorf("EventLTB count = %d, want %d: leaves-the-battlefield triggers must see it go", got, ltb+1)
	}
	if item := onlyAbilityOnStack(g); item == nil || item.SourceCardID != src {
		t.Fatalf("ability on the stack = %+v, want one from the moved source", item)
	}
}

// "Its OWNER's library": a permanent gained from an opponent goes to
// that opponent's library, not the activator's.
func TestABottomSelfCostOnAStolenPermanentGoesToItsOwnersLibrary(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := seedBottomSelf(g, opp.ID, me.ID, noopEffect)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if opp.Library.Cards[0].InstanceID != src {
		t.Fatal("the stolen permanent is not at the bottom of its owner's library")
	}
	for _, c := range me.Library.Cards {
		if c.InstanceID == src {
			t.Fatal("the stolen permanent went to the activator's library")
		}
	}
}

// The ability resolves from the permanent's last-known information
// (CR 608.2h) although the library card is a new object (CR 400.7).
func TestABottomSelfCostLeavesTheEffectTheMovedPermanentsLKI(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	var seen PermanentInfo
	var seenOK bool
	src := seedBottomSelf(g, me.ID, me.ID, func(g *Game, item *StackItem) error {
		if ref, ok := g.SourceObjectForEffect(item); ok {
			seen, seenOK = g.PermanentForEffect(ref)
		}
		return nil
	})
	g.WithWriteLock(func() { findBattlefieldCard(g, src).Counters = map[string]int{CounterPlusOne: 2} })

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < len(g.Seats) && onlyAbilityOnStack(g) != nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if !seenOK || !seen.Left {
		t.Fatalf("at resolution the source reads %+v, %v; want the moved permanent's record", seen, seenOK)
	}
	if seen.Power != 5 || seen.Toughness != 5 {
		t.Errorf("last-known P/T = %d/%d, want 5/5", seen.Power, seen.Toughness)
	}
}

// Only a permanent can be put on the bottom; the source pays one
// component (CR 118.3).
func TestABottomSelfCostValidation(t *testing.T) {
	src := uuid.New()
	for _, z := range []ZoneKind{ZoneHand, ZoneGraveyard, ZoneExile, ZoneCommand} {
		if err := validateBottomSelfCostLocked(src, z, AbilityCost{BottomSelf: true}); !errors.Is(err, ErrActivationZoneNotAllowed) {
			t.Errorf("validate from %s: err = %v, want ErrActivationZoneNotAllowed", z, err)
		}
	}
	for _, cost := range []AbilityCost{
		{BottomSelf: true, SacrificeSelf: true}, {BottomSelf: true, ExileSelf: true}, {BottomSelf: true, ReturnSelf: true},
	} {
		if err := validateBottomSelfCostLocked(src, ZoneBattlefield, cost); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("validate %+v: err = %v, want ErrInvalidParam", cost, err)
		}
	}
	if err := validateBottomSelfCostLocked(src, ZoneBattlefield, AbilityCost{BottomSelf: true}, []uuid.UUID{src}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("source named by another component: err = %v, want ErrInvalidParam", err)
	}
	if AbilityNeedsPermanentSource(AbilityCost{BottomSelf: true}) == "" {
		t.Error("AbilityNeedsPermanentSource does not name the bottom-of-library cost")
	}
	if !AbilityAutoTapExclusions(src, AbilityCost{BottomSelf: true, Mana: "{1}"}, nil, nil, nil, nil)[src] {
		t.Error("the source is not excluded from the auto-tap plan")
	}
}
