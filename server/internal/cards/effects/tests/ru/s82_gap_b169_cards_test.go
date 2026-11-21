package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 GAP batch 169 (Jumpstart 2022 enchantments, #5948).

const (
	gapB169HoldOracle    = "1095fcd9-8127-4345-9052-fccb350296d3"
	gapB169LayClaimOrcl  = "c7111365-b96b-4d6a-8c3c-73a0d8baf641"
	gapB169SlumberOracle = "a11d890e-38d4-467c-bc04-948979c142e0"
	gapB169SoTinyOracle  = "196948c7-7386-4b5f-ad2f-ec9e3aef26fe"
	gapB169WintersOracle = "9df37d14-89fa-4a28-98ee-426062b5740d"
	gapB169AerialOracle  = "290cd0b1-55d6-4b30-962d-075a6f75d562"
	gapB169BrightsOracle = "fa131186-d730-460b-b9e5-b97d89857e24"
	gapB169WorshipOracle = "ea6b1e8e-b6e7-4fce-bfb7-7adf66b9f240"
	gapB169MurderOracle  = "42ca56c5-a9c4-4a2c-ae05-bce47aa6e16c"
)

func TestGapB169CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB169HoldOracle: "Hold for Questioning", gapB169LayClaimOrcl: "Lay Claim",
		gapB169SlumberOracle: "Marit Lage's Slumber", gapB169SoTinyOracle: "So Tiny",
		gapB169WintersOracle: "Winter's Rest", gapB169AerialOracle: "Aerial Modification",
		gapB169BrightsOracle: "Caught in the Brights", gapB169WorshipOracle: "Forced Worship",
		gapB169MurderOracle: "Murder Investigation",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, name %q, completeness %v, caveats %v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGapB169HoldForQuestioningTapsInvestigatesAndLocksActivations(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Hold for Questioning", "Enchantment — Aura", gapB169HoldOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); !c.Tapped {
		t.Error("the enchanted creature was not tapped")
	}
	if clues := effectstest.MkmMP2Clues(g, me.ID); len(clues) != 1 {
		t.Errorf("the Aura's controller has %d Clues, want 1", len(clues))
	}
	effectstest.AssertRestrictions(t, g, bear, game.CantActivate|game.CantActivateMana)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if c, _ := effectstest.BattlefieldCard(g, bear); !c.Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

func TestGapB169HoldForQuestioningEnchantsAPlaneswalkerNotALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 3},
	})
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Hold for Questioning", "Enchantment — Aura", gapB169HoldOracle, effectstest.CardTarget(land)); err == nil {
		t.Error("the Aura may not enchant a land")
	}
	effectstest.CastCatalogSpell(t, g, "Hold for Questioning", "Enchantment — Aura", gapB169HoldOracle, effectstest.CardTarget(walker))
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, walker); !c.Tapped {
		t.Error("the enchanted planeswalker was not tapped")
	}
}

func TestGapB169LayClaimStealsAnyPermanentAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	effectstest.CastCatalogSpell(t, g, "Lay Claim", "Enchantment — Aura", gapB169LayClaimOrcl, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, land); got != me.ID {
		t.Errorf("the land's controller is %v, want the Aura's controller", got)
	}
	spec, _ := effects.Lookup(gapB169LayClaimOrcl)
	if len(spec.Activated) != 1 {
		t.Errorf("Lay Claim has %d activated abilities, want its cycling", len(spec.Activated))
	}
	g2 := effectstest.NewCatalogGame(t)
	id, p := effectstest.CycleFromHand(t, g2, "Lay Claim", "Enchantment — Aura", gapB169LayClaimOrcl, "{C}{C}")
	if !effectstest.InGraveyard(p, id) {
		t.Error("the cycled card is not in the graveyard")
	}
}

func TestGapB169SlumberScriesOnSnowEntriesAndMakesMaritLage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Marit Lage's Slumber", "Legendary Snow Enchantment", gapB169SlumberOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// Another snow permanent entering scries again; a nonsnow one does not.
	effectstest.B12PlayFromHand(t, g, "Snowy Hill", "Snow Land", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.B12PlayFromHand(t, g, "Plain Hill", "Land", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ScryChoiceFor(g, me.ID); c != nil {
		t.Fatal("a nonsnow permanent entering asked for a scry")
	}
	// Ten snow permanents (the Slumber, Snowy Hill and eight more): upkeep sacrifice.
	for i := 0; i < 8; i++ {
		effectstest.B12Permanent(g, me.ID, "Snow Rock", "Snow Land")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByName(g, "Marit Lage's Slumber") {
		t.Error("the Slumber is still on the battlefield")
	}
	if n := effectstest.B28TokensNamed(g, me.ID, "Marit Lage"); n != 1 {
		t.Errorf("%d Marit Lage tokens, want 1", n)
	}
}

func TestGapB169SlumberDoesNothingBelowTenSnowPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Marit Lage's Slumber", "Legendary Snow Enchantment", gapB169SlumberOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	for i := 0; i < 8; i++ {
		effectstest.B12Permanent(g, me.ID, "Snow Rock", "Snow Land")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Marit Lage's Slumber") || effectstest.B28TokensNamed(g, me.ID, "Marit Lage") != 0 {
		t.Error("nine snow permanents sacrificed the Slumber or made Marit Lage")
	}
}

func TestGapB169SoTinyShrinksPowerMoreWithAFullGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 7, 7)
	spec, _ := effects.Lookup(gapB169SoTinyOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("So Tiny keywords = %v, want flash", spec.PrintedKeywords)
	}
	effectstest.CastCatalogSpell(t, g, "So Tiny", "Enchantment — Aura", gapB169SoTinyOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 7 {
		t.Errorf("with a small graveyard the creature is %d/%d, want 5/7", p, tough)
	}
	effectstest.AcGraveyardCards(opp, 7)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest(); g.RecomputeLayersIfStaleLocked() })
	if p, tough := effectstest.PtOf(t, g, bear); p != 1 || tough != 7 {
		t.Errorf("with seven cards in its controller's graveyard the creature is %d/%d, want 1/7", p, tough)
	}
}

func TestGapB169WintersRestHoldsOnlyWhileAnotherSnowPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Winter's Rest", "Snow Enchantment — Aura", gapB169WintersOracle, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Fatal("the enters trigger did not tap the enchanted creature")
	}
	// The Aura is itself snow, which does not count as "another".
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if c, _ := effectstest.BattlefieldCard(g, host); c.Tapped {
		t.Error("without another snow permanent the creature stayed tapped")
	}
}

func TestGapB169WintersRestKeepsItTappedWithAnotherSnowPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "Snow Rock", "Snow Land")
	effectstest.CastCatalogSpell(t, g, "Winter's Rest", "Snow Enchantment — Aura", gapB169WintersOracle, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Error("the creature untapped despite another snow permanent")
	}
}

func TestGapB169AerialModificationMakesAVehicleAFlyingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	van := effectstest.B12Creature(g, me.ID, "Van", "Artifact — Vehicle", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Aerial Modification", "Enchantment — Aura", gapB169AerialOracle, effectstest.CardTarget(van))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, van); !c.IsCreature() {
		t.Error("the enchanted Vehicle is not a creature")
	}
	if p, tough := effectstest.PtOf(t, g, van); p != 5 || tough != 5 {
		t.Errorf("the Vehicle is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, van, "flying") {
		t.Error("the Vehicle does not have flying")
	}
}

func TestGapB169AerialModificationOnACreatureOnlyPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Aerial Modification", "Enchantment — Aura", gapB169AerialOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the creature is %d/%d, want 4/4", p, tough)
	}
	land := effectstest.B12Permanent(g, g.Seats[0].ID, "Land", "Land")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Aerial Modification", "Enchantment — Aura", gapB169AerialOracle, effectstest.CardTarget(land)); err == nil {
		t.Error("the Aura may not enchant a land")
	}
}

func TestGapB169CaughtInTheBrightsExilesOnAVehicleAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	van := effectstest.B12Creature(g, me.ID, "Van", "Artifact Creature — Vehicle", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Caught in the Brights", "Enchantment — Aura", gapB169BrightsOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, victim, game.CantAttackOrBlock)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(van, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, victim) {
		t.Error("the enchanted creature was not exiled")
	}
	if z := effectstest.ZoneOf(g, victim); z != game.ZoneExile {
		t.Errorf("the enchanted creature is in %v, want exile", z)
	}
}

func TestGapB169CaughtInTheBrightsIgnoresANonVehicleAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Caught in the Brights", "Enchantment — Aura", gapB169BrightsOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, victim) {
		t.Error("a non-Vehicle attack exiled the enchanted creature")
	}
}

func TestGapB169ForcedWorshipStopsAttacksNotBlocksAndReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Forced Worship", "Enchantment — Aura", gapB169WorshipOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, bear, game.CantAttack)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{W}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOf(g, aura); z != game.ZoneHand {
		t.Errorf("the Aura is in %v, want its owner's hand", z)
	}
	effectstest.AssertRestrictions(t, g, bear, 0)
}

func TestGapB169MurderInvestigationMakesSoldiersEqualToItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Murder Investigation", "Enchantment — Aura", gapB169MurderOracle, effectstest.CardTarget(theirs)); err == nil {
		t.Error("the Aura may only enchant a creature you control")
	}
	mine := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Murder Investigation", "Enchantment — Aura", gapB169MurderOracle, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(theirs, mine, 4); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B28TokensNamed(g, me.ID, "Soldier"); n != 4 {
		t.Errorf("%d Soldier tokens, want 4 (its power)", n)
	}
}
