package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// slice_295c_removal_test.go — slice 295-c (Removal), the retriage
// slice off card-coverage batches 295/299/300/301/302/304/305/307.
// One file for the slice's twelve registered cards: the main
// behaviour and the illegal/refused case for each.

const (
	realityShiftOracle    = "70dc830e-d05b-4fc7-88dd-879e140b3fbf"
	abruptDecayOracle     = "1c747fe2-289e-492a-a846-aa77707e2dc3"
	voidRendOracle        = "713f16db-95ec-479e-a48c-7a69f7668d7f"
	cloudshiftOracle      = "6879f5ce-7a1b-4606-bad1-885779b0d456"
	essenceFluxOracle     = "64824ae5-efab-4b55-9d3c-b9c690bad857"
	decimateOracle        = "a4e5693f-12a0-451e-818d-d6efc7b4ed25"
	wearTearOracle        = "9842734c-1eac-4509-a731-4c22017ae586"
	requisitionRaidOracle = "09669283-eb14-46a2-b37a-51d7a52da891"
	lethalSchemeOracle    = "c4660b3c-a234-4a3e-83e7-e2fa9d556685"
	balefulMasteryOracle  = "adfcdadd-ddda-477b-8e72-0cae2430fb63"
	fieryConfluenceOracle = "3c22e031-4804-4c31-bd3c-c3f29d456b34"
	golgariCharmOracle    = "f0ab7166-9fcf-46e5-af2c-b7f628db4789"
)

// --- Reality Shift ---------------------------------------------------

func TestRealityShiftExilesAndTheControllerManifests(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	known := game.NewCard("Known Library Card", opp.ID)
	known.TypeLine = "Sorcery"
	top := known.InstanceID
	g.WithWriteLock(func() { opp.Library.PushTop(known) })

	castCatalogSpell(t, g, "Reality Shift", "Instant", realityShiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("the targeted creature should have been exiled")
	}
	if !g.Exile.Contains(bear) {
		t.Error("the targeted creature should be in exile")
	}
	mc := findCardAnywhere(t, g, top)
	if !mc.FaceDownIsPermanent() || mc.FaceDownKind != game.FaceDownManifested {
		t.Fatalf("the opponent's top card should have been manifested, got %+v", mc)
	}
	if mc.Controller != opp.ID {
		t.Errorf("manifester = %s, want the exiled creature's controller %s", mc.Controller, opp.ID)
	}
}

func TestRealityShiftCantTargetANoncreaturePermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)

	err := castCatalogSpellErr(t, g, "Reality Shift", "Instant", realityShiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Reality Shift at an artifact: got %v, want ErrIllegalTarget", err)
	}
}

// --- Abrupt Decay ------------------------------------------------------

func TestAbruptDecayDestroysACheapNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	cheap := pushCostedPermanentForTest(g, opp.ID, "Cheap Thing", "Creature — Bear", "{1}{G}")

	castCatalogSpell(t, g, "Abrupt Decay", "Instant", abruptDecayOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: cheap}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(cheap) {
		t.Error("a mana value 2 permanent should have been destroyed")
	}
}

func TestAbruptDecayCantTargetAManaValueFourPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pricey := pushCostedPermanentForTest(g, opp.ID, "Pricey Thing", "Creature — Giant", "{2}{R}{R}")

	err := castCatalogSpellErr(t, g, "Abrupt Decay", "Instant", abruptDecayOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: pricey}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Abrupt Decay at a mana value 4 permanent: got %v, want ErrIllegalTarget", err)
	}
}

// --- Void Rend -----------------------------------------------------

func TestVoidRendDestroysAnyNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pricey := pushCostedPermanentForTest(g, opp.ID, "Pricey Thing", "Creature — Giant", "{6}{R}{R}")

	castCatalogSpell(t, g, "Void Rend", "Instant", voidRendOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: pricey}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(pricey) {
		t.Error("Void Rend should destroy a nonland permanent regardless of mana value")
	}
}

func TestVoidRendCantTargetALand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	forest := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")

	err := castCatalogSpellErr(t, g, "Void Rend", "Instant", voidRendOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: forest}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Void Rend at a land: got %v, want ErrIllegalTarget", err)
	}
}

// --- Cloudshift ------------------------------------------------------

func TestCloudshiftReturnsUnderYourControlAsANewObject(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Cloudshift", "Instant", cloudshiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("the old instance should be gone; the creature returns as a new object")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "My Bear" && c.Controller == me.ID && c.InstanceID != bear {
			found = true
		}
	}
	if !found {
		t.Error("expected a new instance of the flickered creature under the caster's control")
	}
}

func TestCloudshiftCantTargetACreatureYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	err := castCatalogSpellErr(t, g, "Cloudshift", "Instant", cloudshiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Cloudshift at an opponent's creature: got %v, want ErrIllegalTarget", err)
	}
}

// --- Essence Flux ----------------------------------------------------

func TestEssenceFluxCountersASpiritOnReturn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spirit := b12Creature(g, me.ID, "My Spirit", "Creature — Spirit", 1, 1)

	castCatalogSpell(t, g, "Essence Flux", "Instant", essenceFluxOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spirit}})
	passPriorityAroundTable(t, g)

	var found *game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "My Spirit" && c.InstanceID != spirit {
			cc := c
			found = &cc
		}
	}
	if found == nil {
		t.Fatal("expected the Spirit to return as a new object")
	}
	if found.Counters["+1/+1"] != 1 {
		t.Errorf("+1/+1 counters = %d, want 1 for a returning Spirit", found.Counters["+1/+1"])
	}
}

func TestEssenceFluxDoesNotCounterANonSpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Essence Flux", "Instant", essenceFluxOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	var found *game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "My Bear" && c.InstanceID != bear {
			cc := c
			found = &cc
		}
	}
	if found == nil {
		t.Fatal("expected the Bear to return as a new object")
	}
	if found.Counters["+1/+1"] != 0 {
		t.Errorf("a non-Spirit should not get a counter, got %d", found.Counters["+1/+1"])
	}
}

// --- Decimate --------------------------------------------------------

func TestDecimateDestroysAllFourTypes(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	art := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	cre := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")

	castCatalogSpell(t, g, "Decimate", "Sorcery", decimateOracle,
		[]game.TargetRef{
			{Kind: game.TargetCard, ID: art},
			{Kind: game.TargetCard, ID: cre},
			{Kind: game.TargetCard, ID: ench},
			{Kind: game.TargetCard, ID: land},
		})
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{art, cre, ench, land} {
		if g.Battlefield.Contains(id) {
			t.Errorf("permanent %s should have been destroyed by Decimate", id)
		}
	}
}

func TestDecimateRequiresAllFourTargets(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	art := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	cre := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	// No land on the battlefield: the fourth clause has nothing to
	// offer, so the cast cannot be announced at all.

	err := castCatalogSpellErr(t, g, "Decimate", "Sorcery", decimateOracle,
		[]game.TargetRef{
			{Kind: game.TargetCard, ID: art},
			{Kind: game.TargetCard, ID: cre},
			{Kind: game.TargetCard, ID: ench},
		})
	if err == nil {
		t.Error("Decimate with no legal land target should be refused, not silently accepted")
	}
}

// --- Wear // Tear ------------------------------------------------------

func wearTearCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   wearTearOracle,
		Layout:     game.LayoutSplit,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Wear", TypeLine: "Instant", ManaCost: "{1}{R}"},
			{Name: "Tear", TypeLine: "Instant", ManaCost: "{W}"},
		},
	}
	c.SetFace(0)
	return c
}

func TestWearDestroysTargetArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	rock := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)

	c := wearTearCard(me.ID)
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("cast Wear: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("Wear should have destroyed the targeted artifact")
	}
}

func TestTearCannotBeCastYet(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")

	c := wearTearCard(me.ID)
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{
		Face:    1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}},
	})
	if err == nil {
		t.Error("casting face 1 (Tear) should be refused — split needs fusing, which isn't implemented")
	}
}

// --- Requisition Raid --------------------------------------------------

func TestRequisitionRaidAllThreeBullets(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	art := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castModal(t, g, "Requisition Raid", "Sorcery", requisitionRaidOracle,
		[]int{0, 1, 2},
		[]game.TargetRef{
			modeRef(game.TargetCard, art, 0, 0),
			modeRef(game.TargetCard, ench, 1, 0),
			modeRef(game.TargetPlayer, opp.ID, 2, 0),
		})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(art) {
		t.Error("the artifact bullet should have destroyed the artifact")
	}
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment bullet should have destroyed the enchantment")
	}
	bc, ok := g.LookupCardForEffect(bear)
	if !ok {
		t.Fatal("the bear should still be on the battlefield")
	}
	if bc.Counters["+1/+1"] != 1 {
		t.Errorf("+1/+1 counters on the opponent's bear = %d, want 1", bc.Counters["+1/+1"])
	}
}

func TestRequisitionRaidCounterBulletAlone(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castModal(t, g, "Requisition Raid", "Sorcery", requisitionRaidOracle,
		[]int{2},
		[]game.TargetRef{modeRef(game.TargetPlayer, opp.ID, 0, 0)})
	passPriorityAroundTable(t, g)

	bc, ok := g.LookupCardForEffect(bear)
	if !ok || bc.Counters["+1/+1"] != 1 {
		t.Errorf("expected exactly one +1/+1 counter on the sole creature the named player controls")
	}
}

// --- Lethal Scheme -----------------------------------------------------

func TestLethalSchemeDestroysTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Lethal Scheme", "Instant", lethalSchemeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("Lethal Scheme should have destroyed the targeted creature")
	}
}

func TestLethalSchemeCantTargetANoncreatureNonPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)

	err := castCatalogSpellErr(t, g, "Lethal Scheme", "Instant", lethalSchemeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Lethal Scheme at an artifact: got %v, want ErrIllegalTarget", err)
	}
}

// --- Baleful Mastery ---------------------------------------------------

func TestBalefulMasteryPrintedCostExilesWithNoDraw(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	them := g.Seats[2]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	before := them.Hand.Size()

	castCatalogSpell(t, g, "Baleful Mastery", "Instant", balefulMasteryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("Baleful Mastery should have exiled the targeted creature")
	}
	if !g.Exile.Contains(bear) {
		t.Error("the targeted creature should be in exile, not the graveyard")
	}
	if them.Hand.Size() != before {
		t.Errorf("no opponent should draw when the printed cost was paid: hand %d -> %d", before, them.Hand.Size())
	}
}

func TestBalefulMasteryAltCostDrawsThenExiles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	beforeOpp, beforeThem := opp.Hand.Size(), g.Seats[2].Hand.Size()

	id, err := castWithTapParams(t, g, "Baleful Mastery", "Instant", "", balefulMasteryOracle,
		game.CastSpellParams{
			AlternativeCost: "baleful-mastery",
			Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		})
	if err != nil {
		t.Fatalf("cast Baleful Mastery for its alternative cost: %v", err)
	}
	_ = id
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("Baleful Mastery should have exiled the targeted creature")
	}
	afterOpp, afterThem := opp.Hand.Size(), g.Seats[2].Hand.Size()
	if afterOpp != beforeOpp+1 || afterThem != beforeThem {
		t.Errorf("the CHOSEN opponent should draw: opp %d -> %d, other %d -> %d",
			beforeOpp, afterOpp, beforeThem, afterThem)
	}
}

// --- Fiery Confluence --------------------------------------------------

func TestFieryConfluenceThreeSweepsKillsAThreeToughnessCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	tough := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)

	castModal(t, g, "Fiery Confluence", "Sorcery", fieryConfluenceOracle,
		[]int{0, 0, 0}, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(tough) {
		t.Error("three 1-damage sweeps should have killed the 3-toughness creature")
	}
}

func TestFieryConfluenceMixedModesDamagesOpponentAndDestroysArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	before := opp.Life

	castModal(t, g, "Fiery Confluence", "Sorcery", fieryConfluenceOracle,
		[]int{1, 1, 2},
		[]game.TargetRef{modeRef(game.TargetCard, rock, 2, 0)})
	passPriorityAroundTable(t, g)

	if opp.Life != before-4 {
		t.Errorf("two opponent-damage bullets should deal 4: life %d -> %d", before, opp.Life)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact bullet should have destroyed the rock")
	}
}

// --- Golgari Charm -----------------------------------------------------

func TestGolgariCharmAllCreaturesGetMinusOneMinusOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	oneToughness := b12Creature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1)
	mine := b12Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 1)

	castModal(t, g, "Golgari Charm", "Instant", golgariCharmOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(oneToughness) {
		t.Error("a 1/1 should die to -1/-1 until end of turn")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the -1/-1 is symmetric: it also kills the caster's own 1/1")
	}
}

func TestGolgariCharmDestroysTargetEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")

	castModal(t, g, "Golgari Charm", "Instant", golgariCharmOracle,
		[]int{1}, []game.TargetRef{modeRef(game.TargetCard, ench, 0, 0)})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(ench) {
		t.Error("the destroy-enchantment mode should have destroyed the target")
	}
}

func TestGolgariCharmRegeneratesYourCreaturesFromTheOtherMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := b12Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 1)

	castModal(t, g, "Golgari Charm", "Instant", golgariCharmOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })

	if !g.Battlefield.Contains(mine) {
		t.Error("a regenerated creature should survive a destroy effect")
	}
}
