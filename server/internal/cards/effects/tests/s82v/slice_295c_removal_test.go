package testss82v

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Reality Shift ---------------------------------------------------

func TestRealityShiftExilesAndTheControllerManifests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	known := game.NewCard("Known Library Card", opp.ID)
	known.TypeLine = "Sorcery"
	top := known.InstanceID
	g.WithWriteLock(func() { opp.Library.PushTop(known) })

	effectstest.CastCatalogSpell(t, g, "Reality Shift", "Instant", effectstest.RealityShiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("the targeted creature should have been exiled")
	}
	if !g.Exile.Contains(bear) {
		t.Error("the targeted creature should be in exile")
	}
	mc := effectstest.FindCardAnywhere(t, g, top)
	if !mc.FaceDownIsPermanent() || mc.FaceDownKind != game.FaceDownManifested {
		t.Fatalf("the opponent's top card should have been manifested, got %+v", mc)
	}
	if mc.Controller != opp.ID {
		t.Errorf("manifester = %s, want the exiled creature's controller %s", mc.Controller, opp.ID)
	}
}

func TestRealityShiftCantTargetANoncreaturePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)

	err := effectstest.CastCatalogSpellErr(t, g, "Reality Shift", "Instant", effectstest.RealityShiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Reality Shift at an artifact: got %v, want ErrIllegalTarget", err)
	}
}

// --- Abrupt Decay ------------------------------------------------------

func TestAbruptDecayDestroysACheapNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	cheap := effectstest.PushCostedPermanentForTest(g, opp.ID, "Cheap Thing", "Creature — Bear", "{1}{G}")

	effectstest.CastCatalogSpell(t, g, "Abrupt Decay", "Instant", effectstest.AbruptDecayOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: cheap}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(cheap) {
		t.Error("a mana value 2 permanent should have been destroyed")
	}
}

func TestAbruptDecayCantTargetAManaValueFourPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	pricey := effectstest.PushCostedPermanentForTest(g, opp.ID, "Pricey Thing", "Creature — Giant", "{2}{R}{R}")

	err := effectstest.CastCatalogSpellErr(t, g, "Abrupt Decay", "Instant", effectstest.AbruptDecayOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: pricey}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Abrupt Decay at a mana value 4 permanent: got %v, want ErrIllegalTarget", err)
	}
}

// --- Void Rend -----------------------------------------------------

func TestVoidRendDestroysAnyNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	pricey := effectstest.PushCostedPermanentForTest(g, opp.ID, "Pricey Thing", "Creature — Giant", "{6}{R}{R}")

	effectstest.CastCatalogSpell(t, g, "Void Rend", "Instant", effectstest.VoidRendOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: pricey}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(pricey) {
		t.Error("Void Rend should destroy a nonland permanent regardless of mana value")
	}
}

func TestVoidRendCantTargetALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	forest := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")

	err := effectstest.CastCatalogSpellErr(t, g, "Void Rend", "Instant", effectstest.VoidRendOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: forest}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Void Rend at a land: got %v, want ErrIllegalTarget", err)
	}
}

// --- Cloudshift ------------------------------------------------------

func TestCloudshiftReturnsUnderYourControlAsANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Cloudshift", "Instant", effectstest.CloudshiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	err := effectstest.CastCatalogSpellErr(t, g, "Cloudshift", "Instant", effectstest.CloudshiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Cloudshift at an opponent's creature: got %v, want ErrIllegalTarget", err)
	}
}

// --- Essence Flux ----------------------------------------------------

func TestEssenceFluxCountersASpiritOnReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spirit := effectstest.B12Creature(g, me.ID, "My Spirit", "Creature — Spirit", 1, 1)

	effectstest.CastCatalogSpell(t, g, "Essence Flux", "Instant", effectstest.EssenceFluxOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spirit}})
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Essence Flux", "Instant", effectstest.EssenceFluxOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	art := effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	cre := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")

	effectstest.CastCatalogSpell(t, g, "Decimate", "Sorcery", effectstest.DecimateOracle,
		[]game.TargetRef{
			{Kind: game.TargetCard, ID: art},
			{Kind: game.TargetCard, ID: cre},
			{Kind: game.TargetCard, ID: ench},
			{Kind: game.TargetCard, ID: land},
		})
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{art, cre, ench, land} {
		if g.Battlefield.Contains(id) {
			t.Errorf("permanent %s should have been destroyed by Decimate", id)
		}
	}
}

func TestDecimateRequiresAllFourTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	art := effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	cre := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	// No land on the battlefield: the fourth clause has nothing to
	// offer, so the cast cannot be announced at all.

	err := effectstest.CastCatalogSpellErr(t, g, "Decimate", "Sorcery", effectstest.DecimateOracle,
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
		OracleID:   effectstest.WearTearOracle,
		Layout:     game.LayoutSplit,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Wear", TypeLine: "Instant", ManaCost: "{1}{R}",
				OracleText: "Destroy target artifact.\nFuse (You may cast one or both halves of this card from your hand.)"},
			{Name: "Tear", TypeLine: "Instant", ManaCost: "{W}",
				OracleText: "Destroy target enchantment.\nFuse (You may cast one or both halves of this card from your hand.)"},
		},
	}
	c.SettleImported()
	return c
}

func TestWearDestroysTargetArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	rock := effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)

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
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("Wear should have destroyed the targeted artifact")
	}
}

// TestTearDestroysTargetEnchantment is CR 709.3 (ADR 0103): the right
// half of a split card is cast on its own, under its "#1" entry.
func TestTearDestroysTargetEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")

	c := wearTearCard(me.ID)
	me.Hand.PushTop(c)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{
		Face:    1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}},
	}); err != nil {
		t.Fatalf("cast Tear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("Tear did not destroy the enchantment")
	}
}

// TestWearTearFusedDestroysBoth is CR 702.102: cast fused from hand, for
// {1}{R}{W}, Wear's target then Tear's, both destroyed.
func TestWearTearFusedDestroysBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	rock := effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")

	c := wearTearCard(me.ID)
	me.Hand.PushTop(c)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{
		Fuse: true,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: rock, Slot: 0},
			{Kind: game.TargetCard, ID: ench, Slot: 1},
		},
	}); err != nil {
		t.Fatalf("cast Wear // Tear fused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(ench) {
		t.Errorf("fused Wear // Tear: artifact gone %v, enchantment gone %v — want both",
			!g.Battlefield.Contains(rock), !g.Battlefield.Contains(ench))
	}
	// A pair in the wrong slots is refused: Tear's clause cannot take
	// the artifact.
	c2 := wearTearCard(me.ID)
	me.Hand.PushTop(c2)
	rock2 := effectstest.B12Push(g, opp.ID, "Their Other Rock", "Artifact", "", 0, 0)
	ench2 := effectstest.B12Permanent(g, opp.ID, "Their Other Aura", "Enchantment")
	if err := g.CastSpell(me.ID, c2.InstanceID, game.CastSpellParams{
		Fuse: true,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: ench2, Slot: 0},
			{Kind: game.TargetCard, ID: rock2, Slot: 1},
		},
	}); err == nil {
		t.Error("a fused cast with the targets in the wrong clauses was accepted")
	}
}

// --- Requisition Raid --------------------------------------------------

func TestRequisitionRaidAllThreeBullets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	art := effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.CastModal(t, g, "Requisition Raid", "Sorcery", effectstest.RequisitionRaidOracle,
		[]int{0, 1, 2},
		[]game.TargetRef{
			effectstest.ModeRef(game.TargetCard, art, 0, 0),
			effectstest.ModeRef(game.TargetCard, ench, 1, 0),
			effectstest.ModeRef(game.TargetPlayer, opp.ID, 2, 0),
		})
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.CastModal(t, g, "Requisition Raid", "Sorcery", effectstest.RequisitionRaidOracle,
		[]int{2},
		[]game.TargetRef{effectstest.ModeRef(game.TargetPlayer, opp.ID, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)

	bc, ok := g.LookupCardForEffect(bear)
	if !ok || bc.Counters["+1/+1"] != 1 {
		t.Errorf("expected exactly one +1/+1 counter on the sole creature the named player controls")
	}
}

// --- Lethal Scheme -----------------------------------------------------

func TestLethalSchemeDestroysTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Lethal Scheme", "Instant", effectstest.LethalSchemeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("Lethal Scheme should have destroyed the targeted creature")
	}
}

func TestLethalSchemeCantTargetANoncreatureNonPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)

	err := effectstest.CastCatalogSpellErr(t, g, "Lethal Scheme", "Instant", effectstest.LethalSchemeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Lethal Scheme at an artifact: got %v, want ErrIllegalTarget", err)
	}
}

// --- Baleful Mastery ---------------------------------------------------

func TestBalefulMasteryPrintedCostExilesWithNoDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	them := g.Seats[2]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	before := them.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Baleful Mastery", "Instant", effectstest.BalefulMasteryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	beforeOpp, beforeThem := opp.Hand.Size(), g.Seats[2].Hand.Size()

	id, err := effectstest.CastWithTapParams(t, g, "Baleful Mastery", "Instant", "", effectstest.BalefulMasteryOracle,
		game.CastSpellParams{
			AlternativeCost: "baleful-mastery",
			Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		})
	if err != nil {
		t.Fatalf("cast Baleful Mastery for its alternative cost: %v", err)
	}
	_ = id
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	tough := effectstest.B12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)

	effectstest.CastModal(t, g, "Fiery Confluence", "Sorcery", effectstest.FieryConfluenceOracle,
		[]int{0, 0, 0}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(tough) {
		t.Error("three 1-damage sweeps should have killed the 3-toughness creature")
	}
}

func TestFieryConfluenceMixedModesDamagesOpponentAndDestroysArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	before := opp.Life

	effectstest.CastModal(t, g, "Fiery Confluence", "Sorcery", effectstest.FieryConfluenceOracle,
		[]int{1, 1, 2},
		[]game.TargetRef{effectstest.ModeRef(game.TargetCard, rock, 2, 0)})
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != before-4 {
		t.Errorf("two opponent-damage bullets should deal 4: life %d -> %d", before, opp.Life)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact bullet should have destroyed the rock")
	}
}

// --- Golgari Charm -----------------------------------------------------

func TestGolgariCharmAllCreaturesGetMinusOneMinusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	oneToughness := effectstest.B12Creature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 1)

	effectstest.CastModal(t, g, "Golgari Charm", "Instant", effectstest.GolgariCharmOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(oneToughness) {
		t.Error("a 1/1 should die to -1/-1 until end of turn")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the -1/-1 is symmetric: it also kills the caster's own 1/1")
	}
}

func TestGolgariCharmDestroysTargetEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")

	effectstest.CastModal(t, g, "Golgari Charm", "Instant", effectstest.GolgariCharmOracle,
		[]int{1}, []game.TargetRef{effectstest.ModeRef(game.TargetCard, ench, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(ench) {
		t.Error("the destroy-enchantment mode should have destroyed the target")
	}
}

func TestGolgariCharmRegeneratesYourCreaturesFromTheOtherMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.B12Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 1)

	effectstest.CastModal(t, g, "Golgari Charm", "Instant", effectstest.GolgariCharmOracle, []int{2}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })

	if !g.Battlefield.Contains(mine) {
		t.Error("a regenerated creature should survive a destroy effect")
	}
}
