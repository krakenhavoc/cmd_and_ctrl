package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// return_cost_test.go — #1213: AbilityCost.ReturnToHand, the "return a
// permanent you control to its owner's hand" cost component.
//
// The cards that asked for it are Quirion Ranger, Wirewood Symbiote,
// Master Transmuter and Meloku the Clouded Mirror; the catalog half is
// in cards/effects/return_cost_cards_test.go. What is pinned here is
// the ENGINE contract: paid at announce and in full, refused when the
// board cannot pay (CR 118.3), returned to the permanent's OWNER
// rather than to the activator, and settled without a prompt because
// a cost is one indivisible step (CR 601.2h / 602.2b).

// landCostSpec is "a land you control" as a cost predicate — the
// in-package stand-in for effects.ReturnAPermanentToHand's
// TargetPermanent.
func landCostSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "permanent",
		Label: "a land you control",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsLand()
		},
		Min: 1, Max: 1,
	}
}

func returnLandCost() AbilityCost {
	return AbilityCost{ReturnToHand: &ReturnToHandCost{
		Count:  1,
		Filter: landCostSpec(),
		Label:  "a land you control",
	}}
}

// pushReturnCostSource seats an artifact whose one ability costs
// `cost` and, on resolution, stamps a marker counter on itself so a
// test can see the effect ran.
func pushReturnCostSource(g *Game, owner *Player, cost AbilityCost) uuid.UUID {
	c := NewCard("Return Cost Source", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "return a land: mark",
		Cost:  cost,
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func pushReturnCostLand(g *Game, controller *Player, name string) uuid.UUID {
	c := NewCard(name, controller.ID)
	c.TypeLine = "Land"
	c.Controller = controller.ID
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// The headline: the named land is in its owner's hand at announce, the
// ability is on the stack, and it resolves afterwards.
func TestReturnCostPaysAtAnnounce(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushReturnCostSource(g, me, returnLandCost())
	forest := pushReturnCostLand(g, me, "Forest")
	island := pushReturnCostLand(g, me, "Island")

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{forest}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(forest) {
		t.Error("the named land is still on the battlefield — the cost is paid at announce")
	}
	if !me.Hand.Contains(forest) {
		t.Error("the named land did not reach its owner's hand")
	}
	if !g.Battlefield.Contains(island) {
		t.Error("the unnamed land was returned too")
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("StackMeta has %d items, want 1 (the cost is paid, the effect waits)", len(g.StackMeta))
	}
	if counterOf(g, src, "effect-ran") != 0 {
		t.Error("the effect ran at announce")
	}
	passBothForTest(g)
	if counterOf(g, src, "effect-ran") != 1 {
		t.Error("the ability did not resolve")
	}
}

// CR 109.5: "its OWNER's hand". A permanent you control but do not own
// goes back to the player who owns it, not to the activator.
func TestReturnCostGoesToTheOwnersHand(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushReturnCostSource(g, me, returnLandCost())
	// A land OWNED by the opponent that I control — a Mind Control
	// on a Maze of Ith, in effect.
	stolen := pushReturnCostLand(g, opp, "Their Land")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == stolen {
			g.Battlefield.Cards[i].Controller = me.ID
		}
	}

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{stolen}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Contains(stolen) {
		t.Error("the land went to the activator's hand, not its owner's")
	}
	if !opp.Hand.Contains(stolen) {
		t.Error("the land did not reach its owner's hand")
	}
}

// Every way the component is unpayable, and that none of them moves
// anything. CR 118.3: a cost is not partially payable.
func TestReturnCostRejectsWhatCannotPay(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushReturnCostSource(g, me, returnLandCost())
	mine := pushReturnCostLand(g, me, "My Land")
	theirs := pushReturnCostLand(g, opp, "Their Land")
	notALand := NewCard("Bear", me.ID)
	notALand.TypeLine = "Creature — Bear"
	notALand.Controller = me.ID
	g.Battlefield.PushTop(notALand)

	cases := []struct {
		name string
		ids  []uuid.UUID
		want error
	}{
		{"nothing named", nil, ErrInvalidParam},
		{"two named for a one-permanent clause", []uuid.UUID{mine, mine}, ErrInvalidParam},
		{"a card that does not exist", []uuid.UUID{uuid.New()}, ErrCardNotFound},
		{"an opponent's land", []uuid.UUID{theirs}, ErrCardCallerMismatch},
		{"a permanent the clause does not admit", []uuid.UUID{notALand.InstanceID}, ErrIllegalTarget},
	}
	for _, tc := range cases {
		err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ReturnIDs: tc.ids})
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(notALand.InstanceID) {
		t.Error("a refused activation moved a permanent")
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("StackMeta has %d items, want 0 — every activation was refused", len(g.StackMeta))
	}
}

// The board half of CR 118.3: with nothing the clause admits, the
// ability is not activatable at all, and the shared predicate says so
// before anything is announced. Same walk the view's picker and the
// legal enumerator read, so the three cannot disagree (#544).
func TestReturnCostIsUnpayableWithNoCandidates(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushReturnCostSource(g, me, returnLandCost())

	g.mu.RLock()
	opts := g.ReturnToHandOptionsForEffect(me.ID, src, returnLandCost().ReturnToHand)
	payable := g.ReturnToHandPayable(me.ID, src, returnLandCost().ReturnToHand)
	g.mu.RUnlock()
	if len(opts) != 0 {
		t.Errorf("options with no land of mine: %+v, want none", opts)
	}
	if payable {
		t.Error("ReturnToHandPayable is true with nothing to return")
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("activate with nothing to pay: err = %v, want ErrInvalidParam", err)
	}
}

// A permanent named for the cost does not have to be untapped —
// returning is not tapping. This is the one line tap_others_cost.go's
// validator has and this component's deliberately does not.
func TestReturnCostAcceptsATappedPermanent(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushReturnCostSource(g, me, returnLandCost())
	land := pushReturnCostLand(g, me, "Tapped Land")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == land {
			g.Battlefield.Cards[i].Tapped = true
		}
	}

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Hand.Contains(land) {
		t.Error("a tapped permanent could not pay the cost")
	}
}

// IDs for an ability with no return component are refused rather than
// ignored, exactly as an unexpected sacrifice_ids is: a client that
// sends them is confused about which ability it is firing.
func TestReturnCostRefusesIDsForAnAbilityWithoutOne(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushReturnCostSource(g, me, AbilityCost{})
	land := pushReturnCostLand(g, me, "Forest")

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{land}}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a refused activation returned the land anyway")
	}
}

// The payment is a real battlefield exit: EventLTB fires, so a
// leaves-the-battlefield watcher sees it — and because the cost is
// paid before the stack item is built, the trigger it queues resolves
// ABOVE the ability (CR 603.3b).
func TestReturnCostEmitsTheBattlefieldExit(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushReturnCostSource(g, me, returnLandCost())
	land := pushReturnCostLand(g, me, "Forest")

	before := len(g.Events)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	n := 0
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventLTB && ev.CardID == land {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d EventLTB for the returned land, want exactly one", n)
	}
}

// The source itself is a legal pick when the clause admits it —
// Master Transmuter returning herself — and the activation survives
// its own source leaving the battlefield.
func TestReturnCostMayReturnTheSource(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	// A LAND whose ability returns a land you control: the only land
	// on the board is the source, so the only legal payment is itself.
	c := NewCard("Self-Returning Land", me.ID)
	c.TypeLine = "Land"
	c.Controller = me.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "return a land: mark",
		Cost:  returnLandCost(),
		Effect: func(g *Game, item *StackItem) error {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 1)
		},
	}}
	g.Battlefield.PushTop(c)
	src := c.InstanceID
	life := me.Life

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{src}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(src) {
		t.Error("the source paid for its own ability and is still on the battlefield")
	}
	if !me.Hand.Contains(src) {
		t.Error("the source did not reach its owner's hand")
	}
	passBothForTest(g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want the ability to resolve even though its source left", life, me.Life)
	}
}
