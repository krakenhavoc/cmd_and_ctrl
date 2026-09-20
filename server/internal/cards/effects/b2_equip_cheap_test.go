package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b2_equip_cheap_test.go covers the b2-equip-cheap batch (#1107):
// seven inexpensive Equipment. The shared helpers — advanceToMain,
// seedBear, seedEquipment, equipTo, effectivePower, effectiveToughness,
// effectiveAbilities, attachmentHostOf, assertRestrictions,
// dealCombatDamageToPlayer, advanceToNextSeatsTurn — all live in
// attachments_test.go and its siblings.

const (
	boneSawOracle         = "16e555f2-5aa8-4100-a036-eed48db0e84a"
	catharsShieldOracle   = "b270d091-ea4c-4dfd-9c50-48b45bbc396e"
	herosBladeOracle      = "e6bcd25d-39b6-4619-8a27-9f3d87f4a17b"
	dragonfireBladeOracle = "e809b847-b712-4558-95ea-9bb7356cde91"
	thranPowerSuitOracle  = "3952c288-5633-4e40-abe6-18940d079644"
	bilbosRingOracle      = "a7cc0f6b-6b17-4e76-b2a6-6a6ee2519b61"
	leylineAxeOracle      = "4f597675-0f6d-438c-990c-337171927a5e"
)

// --- Bone Saw: the plain pump-and-equip shape ----------------------

func TestBoneSawPumpsAndEquips(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	saw := seedEquipment(g, me.ID, "Bone Saw", boneSawOracle)

	equipTo(t, g, me.ID, saw, bear)

	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3 (2 base + 1)", got)
	}
	if got := effectiveToughness(t, g, bear); got != 2 {
		t.Errorf("toughness %d, want 2 — Bone Saw grants no toughness", got)
	}
}

// --- Cathar's Shield: a pump plus a keyword grant -------------------

func TestCatharsShieldPumpsAndGrantsVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	shield := seedEquipment(g, me.ID, "Cathar's Shield", catharsShieldOracle)

	equipTo(t, g, me.ID, shield, bear)

	if got := effectiveToughness(t, g, bear); got != 5 {
		t.Errorf("toughness %d, want 5 (2 base + 3)", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d, want 2 — Cathar's Shield grants no power", got)
	}
	if ab := effectiveAbilities(t, g, bear); !containsString(ab, "vigilance") {
		t.Errorf("abilities %v missing vigilance", ab)
	}
}

// --- Hero's Blade: pump plus "attach itself to a legendary arrival" -

func TestHerosBladePumpsAndOptionallyAttachesToALegendaryArrival(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Hero's Blade", herosBladeOracle)

	equipTo(t, g, me.ID, blade, bear)
	if got := effectivePower(t, g, bear); got != 5 {
		t.Fatalf("power %d, want 5 (2 base + 3)", got)
	}
	if got := effectiveToughness(t, g, bear); got != 4 {
		t.Fatalf("toughness %d, want 4 (2 base + 2)", got)
	}

	legend := seedLegendaryCreature(g, me.ID, "Commander")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: legend})
	})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, blade); host.Kind != game.TargetCard || host.ID != legend {
		t.Fatalf("Hero's Blade AttachedTo = %+v, want the legendary arrival %s", host, legend)
	}
	if got := effectivePower(t, g, legend); got != 5 {
		t.Errorf("power %d, want 5 (2 base + 3) after the Blade re-attached itself", got)
	}
}

// A non-legendary creature entering never prompts — the trigger's own
// AppliesTo, not the "you may" answer, is what filters it out.
func TestHerosBladeDoesNotOfferToAttachToAnOrdinaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedEquipment(g, me.ID, "Hero's Blade", herosBladeOracle)

	ordinary := seedBear(g, me.ID)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: ordinary})
	})
	passPriorityAroundTable(t, g)

	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
			t.Fatalf("Hero's Blade offered to attach to a non-legendary creature")
		}
	}
}

// --- Dragonfire Blade: the pump ships, the two undeclared clauses --
// -- do not --------------------------------------------------------

func TestDragonfireBladePumpsAndEquipsWithNoHexproofGranted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Dragonfire Blade", dragonfireBladeOracle)

	equipTo(t, g, me.ID, blade, bear)

	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4 (2 base + 2)", got)
	}
	if got := effectiveToughness(t, g, bear); got != 4 {
		t.Errorf("toughness %d, want 4 (2 base + 2)", got)
	}
	// #259: the undeclared "hexproof from monocolored" must not leak
	// out as plain hexproof, which would be a strictly stronger card
	// than what is printed.
	if ab := effectiveAbilities(t, g, bear); containsString(ab, "hexproof") {
		t.Errorf("abilities %v grant hexproof — the caveated clause must not leak in", ab)
	}
}

// --- Thran Power Suit: a count of attachments on the HOST, plus a --
// -- ward granted to the host ---------------------------------------

func TestThranPowerSuitCountsAurasAndEquipmentAttachedToTheHost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	suit := seedEquipment(g, me.ID, "Thran Power Suit", thranPowerSuitOracle)

	equipTo(t, g, me.ID, suit, bear)
	// The Suit counts itself: one Equipment attached to the bear.
	if got := effectivePower(t, g, bear); got != 3 {
		t.Fatalf("power %d, want 3 (2 base + 1 for the Suit counting itself)", got)
	}
	if got := effectiveToughness(t, g, bear); got != 3 {
		t.Fatalf("toughness %d, want 3", got)
	}

	saw := seedEquipment(g, me.ID, "Bone Saw", boneSawOracle)
	equipTo(t, g, me.ID, saw, bear)
	// Now two Equipment are attached: the count grows to 2 (+2/+2 from
	// the Suit) on top of Bone Saw's own flat +1/+0 — the two statics
	// are read fresh on the same recompute, with no bookkeeping.
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5 (2 base + 2 count + 1 Bone Saw)", got)
	}
	if got := effectiveToughness(t, g, bear); got != 4 {
		t.Errorf("toughness %d, want 4 (2 base + 2 count)", got)
	}
}

func TestThranPowerSuitWardsTheEquippedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	suit := seedEquipment(g, me.ID, "Thran Power Suit", thranPowerSuitOracle)
	equipTo(t, g, me.ID, suit, bear)

	castAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the opponent's removal spell")
	}
}

// --- Bilbo's Ring: a condition on the static, a real "attacks -----
// -- alone" trigger, and a creature-type-restricted equip -----------

func TestBilbosRingGrantsHexproofAndCantBeBlockedOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	ring := seedEquipment(g, me.ID, "Bilbo's Ring", bilbosRingOracle)
	equipTo(t, g, me.ID, ring, bear)

	if ab := effectiveAbilities(t, g, bear); !containsString(ab, "hexproof") {
		t.Errorf("abilities %v missing hexproof during the controller's own turn", ab)
	}
	assertRestrictions(t, g, bear, game.CantBeBlocked)

	advanceToNextSeatsTurn(t, g)
	if ab := effectiveAbilities(t, g, bear); containsString(ab, "hexproof") {
		t.Errorf("hexproof survived outside the controller's turn: %v", ab)
	}
	assertRestrictions(t, g, bear, 0)
}

func TestBilbosRingAttacksAloneDrawsAndLosesLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	ring := seedEquipment(g, me.ID, "Bilbo's Ring", bilbosRingOracle)
	equipTo(t, g, me.ID, ring, bear)

	handBefore := me.Hand.Size()
	lifeBefore := me.Life

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d, want %d — attacking alone should draw a card", got, handBefore+1)
	}
	if got := me.Life; got != lifeBefore-1 {
		t.Errorf("life %d, want %d — attacking alone should cost 1 life", got, lifeBefore-1)
	}
}

func TestBilbosRingDoesNotTriggerWhenAttackingWithAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	other := seedBear(g, me.ID)
	ring := seedEquipment(g, me.ID, "Bilbo's Ring", bilbosRingOracle)
	equipTo(t, g, me.ID, ring, bear)

	handBefore := me.Hand.Size()
	lifeBefore := me.Life

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{bear, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand %d, want %d — attacking alongside another creature must not trigger", got, handBefore)
	}
	if got := me.Life; got != lifeBefore {
		t.Errorf("life %d, want %d — attacking alongside another creature must not trigger", got, lifeBefore)
	}
}

func TestBilbosRingEquipHalflingRefusesANonHalfling(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	halfling := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bandobras Took", TypeLine: "Legendary Creature — Halfling Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	ring := seedEquipment(g, me.ID, "Bilbo's Ring", bilbosRingOracle)

	err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if err == nil {
		t.Fatal("Equip Halfling {1} accepted a non-Halfling creature")
	}

	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: halfling}},
	}); err != nil {
		t.Fatalf("Equip Halfling {1} refused a legal Halfling target: %v", err)
	}
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, ring); host.Kind != game.TargetCard || host.ID != halfling {
		t.Fatalf("Bilbo's Ring AttachedTo = %+v, want the Halfling %s", host, halfling)
	}
}

// --- Leyline Axe: pump plus two keyword grants ----------------------

func TestLeylineAxePumpsAndGrantsDoubleStrikeAndTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	axe := seedEquipment(g, me.ID, "Leyline Axe", leylineAxeOracle)

	equipTo(t, g, me.ID, axe, bear)

	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3 (2 base + 1)", got)
	}
	if got := effectiveToughness(t, g, bear); got != 3 {
		t.Errorf("toughness %d, want 3 (2 base + 1)", got)
	}
	ab := effectiveAbilities(t, g, bear)
	for _, want := range []string{"double strike", "trample"} {
		if !containsString(ab, want) {
			t.Errorf("abilities %v missing %q", ab, want)
		}
	}
}
