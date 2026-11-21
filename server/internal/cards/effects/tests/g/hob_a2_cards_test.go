package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hob_a2_cards_test.go — The Hobbit / The Hobbit Eternal, half A, second
// batch (#4454): ten cards whose seams all exist. Each test casts or
// plays the real catalog entry and asserts what the card's text says.

const (
	hobA2QuarrelOracle         = "db05b114-2eb5-46bc-81e7-7e036d8eaab3"
	hobA2TrollNegotiationsOrcl = "6e275711-eb76-4924-ba38-c38a36d9d82a"
	hobA2GuardianOracle        = "c3fb2829-64f2-4795-b815-edcc5bbbc360"
	hobA2PathmakerOracle       = "f1b36dde-9cde-46c8-8d5a-9ad93feba41f"
	hobA2BattleScarredOracle   = "3e765710-e567-4af0-b5f5-7c182649b757"
	hobA2LookoutOracle         = "ed839fae-3429-4ac5-997c-286d8e8cf05f"
	hobA2GreyDogOracle         = "6f83da19-fd89-44ec-88f3-0c3fddfbd1b2"
	hobA2ThrorsMapOracle       = "925cacdc-6d09-4fbd-9258-9f0539da180c"
	hobA2GiftOfStrandsOracle   = "7a7789e5-f2ac-49b2-943e-7d49e4727fb6"
	hobA2SpatulaOracle         = "3cc333a1-c854-4b7c-8002-910e48222371"
)

func TestHobA2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		hobA2QuarrelOracle:         "Quarrel",
		hobA2TrollNegotiationsOrcl: "Troll Negotiations",
		hobA2GuardianOracle:        "Guardian of the Halls",
		hobA2PathmakerOracle:       "Mirkwood Pathmaker",
		hobA2BattleScarredOracle:   "Battle-Scarred Goblin",
		hobA2LookoutOracle:         "Lothlórien Lookout",
		hobA2GreyDogOracle:         "Long-Bodied Grey Dog",
		hobA2ThrorsMapOracle:       "Thrór's Map",
		hobA2GiftOfStrandsOracle:   "Gift of Strands",
		hobA2SpatulaOracle:         "Well-Worn Spatula",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// Quarrel: the creature you control deals damage equal to its power and
// takes none back; the second target must be an opponent's creature.
func TestHobA2QuarrelIsOneSided(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Beast", "Creature — Beast", 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Quarrel", "Instant", hobA2QuarrelOracle, effectstest.S58Refs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("4 damage should kill the Bear")
	}
	if effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Error("the Beast deals the damage and takes none")
	}
}

func TestHobA2QuarrelRefusesYourOwnCreatureAsTheVictim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Beast", "Creature — Beast", 4, 4)
	other := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Quarrel", "Instant", hobA2QuarrelOracle, effectstest.S58Refs(mine, other)); err != game.ErrIllegalTarget {
		t.Errorf("your own creature as the victim: err = %v, want ErrIllegalTarget", err)
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Quarrel", "Instant", hobA2QuarrelOracle, effectstest.S58Refs(theirs, theirs)); err != game.ErrIllegalTarget {
		t.Errorf("their creature as the damage source: err = %v, want ErrIllegalTarget", err)
	}
}

// Troll Negotiations: two +1/+1 counters first, then the fight with both
// powers read after them. A 2/2 becomes 4/4, kills the 3/3 and survives
// its 3 damage.
func TestHobA2TrollNegotiationsCountersThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)

	effectstest.CastCatalogSpell(t, g, "Troll Negotiations", "Sorcery", hobA2TrollNegotiationsOrcl, effectstest.S58Refs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) {
		t.Error("the pumped 4/4 should kill the 3/3")
	}
	c, ok := effectstest.BattlefieldCard(g, mine)
	if !ok {
		t.Fatal("the 4/4 died to 3 damage")
	}
	if got := c.Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("creature is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if effectstest.DamageMarkedOn(g, mine) != 3 {
		t.Errorf("damage on the fighter = %d, want 3 (a fight is two-sided)", effectstest.DamageMarkedOn(g, mine))
	}
}

func TestHobA2TrollNegotiationsRefusesYourOwnCreatureAsTheSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Troll Negotiations", "Sorcery", hobA2TrollNegotiationsOrcl, effectstest.S58Refs(a, b)); err != game.ErrIllegalTarget {
		t.Errorf("err = %v, want ErrIllegalTarget", err)
	}
}

func TestHobA2GuardianOfTheHallsHasTrampleAndGrowsByThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Guardian of the Halls", "Creature — Elf Soldier", hobA2GuardianOracle, false)
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "trample") {
		t.Error("Guardian of the Halls should have trample")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G", "G", "C", "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, id)
	if got := c.Counters[game.CounterPlusOne]; got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("creature is %d/%d, want 4/4 (1/1 test body plus three counters)", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestHobA2MirkwoodPathmakerCountsYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Mirkwood Pathmaker", "Creature — Elf Ranger", hobA2PathmakerOracle, false)
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if c := effectstest.GaLayered(t, g, id); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("with three lands got %d/%d, want 3/3 (an opponent's land does not count)", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if c := effectstest.GaLayered(t, g, id); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("with four lands got %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
}

// Battle-Scarred Goblin: when it becomes blocked it deals 1 damage to
// each creature blocking it, once however many block.
func TestHobA2BattleScarredGoblinDamagesEachBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	goblin := effectstest.PushCatalogPermanent(g, me.ID, "Battle-Scarred Goblin", "Creature — Goblin Warrior", hobA2BattleScarredOracle, false)
	a := effectstest.PushVanillaCreature(g, them.ID, "Wall A", 0, 5)
	b := effectstest.PushVanillaCreature(g, them.ID, "Wall B", 0, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(goblin, them.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(w, goblin); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
	for _, w := range []uuid.UUID{a, b} {
		if got := effectstest.FindBattlefieldCardForTest(g, w).DamageMarked; got != 1 {
			t.Errorf("blocker damage = %d, want 1", got)
		}
	}
}

// With nobody blocking, the trigger never fires.
func TestHobA2BattleScarredGoblinUnblockedDealsNothingExtra(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	goblin := effectstest.PushCatalogPermanent(g, me.ID, "Battle-Scarred Goblin", "Creature — Goblin Warrior", hobA2BattleScarredOracle, false)
	bystander := effectstest.PushVanillaCreature(g, them.ID, "Bystander", 0, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(goblin, them.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
	if got := effectstest.FindBattlefieldCardForTest(g, bystander).DamageMarked; got != 0 {
		t.Errorf("a creature that did not block took %d damage", got)
	}
}

func TestHobA2LothlorienLookoutScriesOneWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	lookout := effectstest.PushCatalogPermanent(g, me.ID, "Lothlórien Lookout", "Creature — Elf Scout", hobA2LookoutOracle, false)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(lookout, them.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry prompt from the attack trigger: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("scry 1 offered %d cards", len(c.ScryCards))
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
}

func TestHobA2LongBodiedGreyDogMakesATappedTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CastCatalogSpell(t, g, "Long-Bodied Grey Dog", "Creature — Dog", hobA2GreyDogOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Dog should have resolved")
	}
	for _, kw := range []string{"flash", "reach"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, id, kw) {
			t.Errorf("missing %s", kw)
		}
	}
	treasures, tapped := 0, 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == me.ID {
			treasures++
			if c.Tapped {
				tapped++
			}
		}
	}
	if treasures != 1 || tapped != 1 {
		t.Errorf("Treasures = %d (tapped %d), want one, tapped", treasures, tapped)
	}
}

// Thrór's Map: the search offers only basic lands and puts the pick into
// your hand.
func TestHobA2ThrorsMapFetchesABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Command Tower", TypeLine: "Land"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	id := effectstest.CastCatalogSpell(t, g, "Thrór's Map", "Legendary Artifact", hobA2ThrorsMapOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Thrór's Map should have resolved")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != forest {
		t.Fatalf("want a search offering only the basic Forest, got %+v", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(forest) {
		t.Error("the basic land did not go to hand")
	}
}

func TestHobA2ThrorsMapLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	id := effectstest.PushCatalogPermanent(g, me.ID, "Thrór's Map", "Legendary Artifact", hobA2ThrorsMapOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1 before the discard", got-hand)
	}
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", effectstest.DiscardOwed(g, me.ID))
	}
	if !effectstest.B20Tapped(t, g, id) {
		t.Error("the loot has a tap cost")
	}
}

func TestHobA2GiftOfStrandsPumpsAndScriesTwo(t *testing.T) {
	spec, _ := effects.Lookup(hobA2GiftOfStrandsOracle)
	hasFlash := false
	for _, kw := range spec.PrintedKeywords {
		hasFlash = hasFlash || kw == "flash"
	}
	if !hasFlash {
		t.Error("Gift of Strands has flash")
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	target := effectstest.PushVanillaCreature(g, me.ID, "Recipient", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Gift of Strands", "Enchantment — Aura", hobA2GiftOfStrandsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura should have resolved")
	}
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("want a scry 2 prompt, got %+v", c)
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if got := effectstest.GaLayered(t, g, target); got.CurrentPower() != 5 || got.CurrentToughness() != 5 {
		t.Errorf("enchanted creature is %d/%d, want 5/5", got.CurrentPower(), got.CurrentToughness())
	}
}

func TestHobA2WellWornSpatulaGainsLifeAndEquips(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := me.Life
	id := effectstest.CastCatalogSpell(t, g, "Well-Worn Spatula", "Artifact — Equipment", hobA2SpatulaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Spatula should have resolved")
	}
	if me.Life != before+2 {
		t.Errorf("life = %d, want %d", me.Life, before+2)
	}
	effectstest.B06AddMana(me, "C")
	effectstest.EquipTo(t, g, me.ID, id, bear)
	if c := effectstest.GaLayered(t, g, bear); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("equipped creature is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
}
