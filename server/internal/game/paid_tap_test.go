package game

import (
	"testing"

	"github.com/google/uuid"
)

// paid_tap_test.go — #759: the payment record of a TapOthers cost
// (PaidCost.TappedOthers) and the CR 608.2h reader station resolves
// through (PaidTapPowerForEffect).
//
// Station puts "charge counters equal to the tapped creature's power"
// on its source (CR 702.184a), and the Edge of Eternities release
// notes pin WHEN that power is read: as the ability resolves, or as
// the creature last existed on the battlefield if it has left. So the
// record is an identity with a fallback number, and these tests pin
// the three readings — live, last-known, and "a new object under the
// same instance ID is not the creature that was tapped".

// pushPowerReadingSource seats an artifact whose one activated ability
// costs one ANOTHER untapped creature and, on resolution, records the
// power PaidTapPowerForEffect answered as a counter on itself — so a
// test can read what the resolution saw.
func pushPowerReadingSource(g *Game, owner *Player) uuid.UUID {
	c := NewCard("Power Reader", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "station-shaped: read power",
		Cost:  AbilityCost{TapOthers: tapOthersStationCost()},
		Effect: func(g *Game, item *StackItem) error {
			if len(item.Paid.TappedOthers) == 0 {
				return nil
			}
			return g.applyCounterLocked(item.SourceCardID, "read-power", g.PaidTapPowerForEffect(item.Paid.TappedOthers[0]))
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// onlyStackItem returns the one item on the stack, failing otherwise.
func onlyStackItem(t *testing.T, g *Game) *StackItem {
	t.Helper()
	if len(g.StackMeta) != 1 {
		t.Fatalf("StackMeta has %d items, want 1", len(g.StackMeta))
	}
	for _, it := range g.StackMeta {
		return it
	}
	return nil
}

// The payment writes down WHICH object was tapped and what its power
// was — after the item exists, because the taps are paid there.
func TestTapOthersPaymentIsRecordedOnTheStackItem(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushPowerReadingSource(g, me)
	bear := pushTapOthersPermanent(g, me, "Ready Bear", "Creature — Bear", false)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	item := onlyStackItem(t, g)
	got := item.Paid.TappedOthers
	if len(got) != 1 {
		t.Fatalf("TappedOthers = %+v, want one entry", got)
	}
	g.mu.RLock()
	epoch := findBattlefieldCard(g, bear).ObjectEpoch
	g.mu.RUnlock()
	want := PaidTap{ID: bear, Epoch: epoch, Power: 2}
	if got[0] != want {
		t.Errorf("TappedOthers[0] = %+v, want %+v", got[0], want)
	}
}

// CR 608.2h, the live half: the tapped creature is still on the
// battlefield, so a pump in response is what the resolution reads.
func TestPaidTapPowerIsReadAsTheAbilityResolves(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushPowerReadingSource(g, me)
	bear := pushTapOthersPermanent(g, me, "Ready Bear", "Creature — Bear", false)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := g.AddCounter(bear, "+1/+1", 3); err != nil {
		t.Fatalf("pump: %v", err)
	}
	passBothForTest(g)
	if got := counterOf(g, src, "read-power"); got != 5 {
		t.Errorf("resolution read power %d, want 5 — the 2/2 pumped to 5/5 in response", got)
	}
}

// CR 608.2h, the last-known half: the creature was pumped and then
// left, so the resolution reads the power it had AS IT LEFT — neither
// the power it was tapped with nor nothing.
func TestPaidTapPowerIsLastKnownWhenTheCreatureLeaves(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushPowerReadingSource(g, me)
	bear := pushTapOthersPermanent(g, me, "Ready Bear", "Creature — Bear", false)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := g.AddCounter(bear, "+1/+1", 2); err != nil {
		t.Fatalf("pump: %v", err)
	}
	if err := g.SacrificePermanent(me.ID, bear); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	item := onlyStackItem(t, g)
	if tap := item.Paid.TappedOthers[0]; !tap.Left || tap.Power != 4 {
		t.Errorf("record after the exit = %+v, want Left with the 4 power it left with", tap)
	}
	passBothForTest(g)
	if got := counterOf(g, src, "read-power"); got != 4 {
		t.Errorf("resolution read power %d, want 4 — its power as it last existed on the battlefield", got)
	}
}

// CR 400.7: the creature left and came back under the same instance
// ID. The permanent on the battlefield now is a NEW object, and the
// tapped creature is the one that left — so the resolution reads the
// frozen last-known number, not the new object's power.
func TestPaidTapPowerIgnoresANewObjectUnderTheSameID(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushPowerReadingSource(g, me)
	bear := pushTapOthersPermanent(g, me, "Ready Bear", "Creature — Bear", false)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	battlefield := ZoneRef{Kind: ZoneBattlefield}
	hand := ZoneRef{Kind: ZoneHand, Owner: me.ID}
	if err := g.MoveCardByID(battlefield, hand, bear); err != nil {
		t.Fatalf("bounce: %v", err)
	}
	if err := g.MoveCardByID(hand, battlefield, bear); err != nil {
		t.Fatalf("replay: %v", err)
	}
	if err := g.AddCounter(bear, "+1/+1", 5); err != nil {
		t.Fatalf("pump the new object: %v", err)
	}
	passBothForTest(g)
	if got := counterOf(g, src, "read-power"); got != 2 {
		t.Errorf("resolution read power %d, want 2 — the new object's 7 is not the tapped creature's", got)
	}
}

// The record is carried by undo: a clone's record is its own, so the
// exit hook rewriting one game's cannot reach the other's.
func TestPaidTapRecordIsDeepCopiedByClone(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushPowerReadingSource(g, me)
	bear := pushTapOthersPermanent(g, me, "Ready Bear", "Creature — Bear", false)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	clone := g.Clone()
	if err := g.SacrificePermanent(me.ID, bear); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	orig := onlyStackItem(t, g).Paid.TappedOthers[0]
	copied := onlyStackItem(t, clone).Paid.TappedOthers[0]
	if !orig.Left {
		t.Fatal("the original's record was not frozen by the exit")
	}
	if copied.Left || copied.ID != bear {
		t.Errorf("the clone's record = %+v — it shares the original's backing array", copied)
	}
}

// A CR 707.10 copy of the ability carries the SAME tapped creature,
// in a record of its own.
func TestAbilityCopyCarriesThePaidTaps(t *testing.T) {
	p := PaidCost{TappedOthers: []PaidTap{{ID: uuid.New(), Power: 3}}}
	c := copiedPaidCost(p)
	if len(c.TappedOthers) != 1 || c.TappedOthers[0] != p.TappedOthers[0] {
		t.Fatalf("copy's TappedOthers = %+v, want %+v", c.TappedOthers, p.TappedOthers)
	}
	c.TappedOthers[0].Left = true
	if p.TappedOthers[0].Left {
		t.Error("the copy's record aliases the original's")
	}
}
