package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Bone Saw: the plain pump-and-equip shape ----------------------

func TestBoneSawPumpsAndEquips(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	saw := effectstest.SeedEquipment(g, me.ID, "Bone Saw", effectstest.BoneSawOracle)

	effectstest.EquipTo(t, g, me.ID, saw, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3 (2 base + 1)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 2 {
		t.Errorf("toughness %d, want 2 — Bone Saw grants no toughness", got)
	}
}

// --- Cathar's Shield: a pump plus a keyword grant -------------------

func TestCatharsShieldPumpsAndGrantsVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	shield := effectstest.SeedEquipment(g, me.ID, "Cathar's Shield", effectstest.CatharsShieldOracle)

	effectstest.EquipTo(t, g, me.ID, shield, bear)

	if got := effectstest.EffectiveToughness(t, g, bear); got != 5 {
		t.Errorf("toughness %d, want 5 (2 base + 3)", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d, want 2 — Cathar's Shield grants no power", got)
	}
	if ab := effectstest.EffectiveAbilities(t, g, bear); !effectstest.ContainsString(ab, "vigilance") {
		t.Errorf("abilities %v missing vigilance", ab)
	}
}

// --- Hero's Blade: pump plus "attach itself to a legendary arrival" -

func TestHerosBladePumpsAndOptionallyAttachesToALegendaryArrival(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Hero's Blade", effectstest.HerosBladeOracle)

	effectstest.EquipTo(t, g, me.ID, blade, bear)
	if got := effectstest.EffectivePower(t, g, bear); got != 5 {
		t.Fatalf("power %d, want 5 (2 base + 3)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 4 {
		t.Fatalf("toughness %d, want 4 (2 base + 2)", got)
	}

	legend := effectstest.SeedLegendaryCreature(g, me.ID, "Commander")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: legend})
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, blade); host.Kind != game.TargetCard || host.ID != legend {
		t.Fatalf("Hero's Blade AttachedTo = %+v, want the legendary arrival %s", host, legend)
	}
	if got := effectstest.EffectivePower(t, g, legend); got != 5 {
		t.Errorf("power %d, want 5 (2 base + 3) after the Blade re-attached itself", got)
	}
}

// A non-legendary creature entering never prompts — the trigger's own
// AppliesTo, not the "you may" answer, is what filters it out.
func TestHerosBladeDoesNotOfferToAttachToAnOrdinaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEquipment(g, me.ID, "Hero's Blade", effectstest.HerosBladeOracle)

	ordinary := effectstest.SeedBear(g, me.ID)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: ordinary})
	})
	effectstest.PassPriorityAroundTable(t, g)

	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
			t.Fatalf("Hero's Blade offered to attach to a non-legendary creature")
		}
	}
}

// --- Dragonfire Blade: the pump and "hexproof from monocolored" ----
// (#3172 closed the last caveat; the targeting half is in
// hexproof_from_cards_test.go) ---------------------------------------

func TestDragonfireBladePumpsAndGrantsHexproofFromMonocoloredOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Dragonfire Blade", effectstest.DragonfireBladeOracle)

	effectstest.EquipTo(t, g, me.ID, blade, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4 (2 base + 2)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 4 {
		t.Errorf("toughness %d, want 4 (2 base + 2)", got)
	}
	ab := effectstest.EffectiveAbilities(t, g, bear)
	if !effectstest.ContainsString(ab, game.HexproofFromMonocolored) {
		t.Errorf("abilities %v lack %q", ab, game.HexproofFromMonocolored)
	}
	// #259: it must not leak out as plain hexproof, which would be a
	// strictly stronger card than what is printed.
	if effectstest.ContainsString(ab, "hexproof") {
		t.Errorf("abilities %v grant plain hexproof", ab)
	}
}

// --- Thran Power Suit: a count of attachments on the HOST, plus a --
// -- ward granted to the host ---------------------------------------

func TestThranPowerSuitCountsAurasAndEquipmentAttachedToTheHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	suit := effectstest.SeedEquipment(g, me.ID, "Thran Power Suit", effectstest.ThranPowerSuitOracle)

	effectstest.EquipTo(t, g, me.ID, suit, bear)
	// The Suit counts itself: one Equipment attached to the bear.
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Fatalf("power %d, want 3 (2 base + 1 for the Suit counting itself)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 3 {
		t.Fatalf("toughness %d, want 3", got)
	}

	saw := effectstest.SeedEquipment(g, me.ID, "Bone Saw", effectstest.BoneSawOracle)
	effectstest.EquipTo(t, g, me.ID, saw, bear)
	// Now two Equipment are attached: the count grows to 2 (+2/+2 from
	// the Suit) on top of Bone Saw's own flat +1/+0 — the two statics
	// are read fresh on the same recompute, with no bookkeeping.
	if got := effectstest.EffectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5 (2 base + 2 count + 1 Bone Saw)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 4 {
		t.Errorf("toughness %d, want 4 (2 base + 2 count)", got)
	}
}

func TestThranPowerSuitWardsTheEquippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	suit := effectstest.SeedEquipment(g, me.ID, "Thran Power Suit", effectstest.ThranPowerSuitOracle)
	effectstest.EquipTo(t, g, me.ID, suit, bear)

	effectstest.CastAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the opponent's removal spell")
	}
}

// --- Bilbo's Ring: a condition on the static, a real "attacks -----
// -- alone" trigger, and a creature-type-restricted equip -----------

func TestBilbosRingGrantsHexproofAndCantBeBlockedOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	ring := effectstest.SeedEquipment(g, me.ID, "Bilbo's Ring", effectstest.BilbosRingOracle)
	effectstest.EquipTo(t, g, me.ID, ring, bear)

	if ab := effectstest.EffectiveAbilities(t, g, bear); !effectstest.ContainsString(ab, "hexproof") {
		t.Errorf("abilities %v missing hexproof during the controller's own turn", ab)
	}
	effectstest.AssertRestrictions(t, g, bear, game.CantBeBlocked)

	effectstest.AdvanceToNextSeatsTurn(t, g)
	if ab := effectstest.EffectiveAbilities(t, g, bear); effectstest.ContainsString(ab, "hexproof") {
		t.Errorf("hexproof survived outside the controller's turn: %v", ab)
	}
	effectstest.AssertRestrictions(t, g, bear, 0)
}

func TestBilbosRingAttacksAloneDrawsAndLosesLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	ring := effectstest.SeedEquipment(g, me.ID, "Bilbo's Ring", effectstest.BilbosRingOracle)
	effectstest.EquipTo(t, g, me.ID, ring, bear)

	handBefore := me.Hand.Size()
	lifeBefore := me.Life

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d, want %d — attacking alone should draw a card", got, handBefore+1)
	}
	if got := me.Life; got != lifeBefore-1 {
		t.Errorf("life %d, want %d — attacking alone should cost 1 life", got, lifeBefore-1)
	}
}

func TestBilbosRingDoesNotTriggerWhenAttackingWithAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	other := effectstest.SeedBear(g, me.ID)
	ring := effectstest.SeedEquipment(g, me.ID, "Bilbo's Ring", effectstest.BilbosRingOracle)
	effectstest.EquipTo(t, g, me.ID, ring, bear)

	handBefore := me.Hand.Size()
	lifeBefore := me.Life

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{bear, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand %d, want %d — attacking alongside another creature must not trigger", got, handBefore)
	}
	if got := me.Life; got != lifeBefore {
		t.Errorf("life %d, want %d — attacking alongside another creature must not trigger", got, lifeBefore)
	}
}

func TestBilbosRingEquipHalflingRefusesANonHalfling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	halfling := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bandobras Took", TypeLine: "Legendary Creature — Halfling Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	ring := effectstest.SeedEquipment(g, me.ID, "Bilbo's Ring", effectstest.BilbosRingOracle)

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
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, ring); host.Kind != game.TargetCard || host.ID != halfling {
		t.Fatalf("Bilbo's Ring AttachedTo = %+v, want the Halfling %s", host, halfling)
	}
}

// --- Leyline Axe: pump plus two keyword grants ----------------------

func TestLeylineAxePumpsAndGrantsDoubleStrikeAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	axe := effectstest.SeedEquipment(g, me.ID, "Leyline Axe", effectstest.LeylineAxeOracle)

	effectstest.EquipTo(t, g, me.ID, axe, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3 (2 base + 1)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 3 {
		t.Errorf("toughness %d, want 3 (2 base + 1)", got)
	}
	ab := effectstest.EffectiveAbilities(t, g, bear)
	for _, want := range []string{"double strike", "trample"} {
		if !effectstest.ContainsString(ab, want) {
			t.Errorf("abilities %v missing %q", ab, want)
		}
	}
}
