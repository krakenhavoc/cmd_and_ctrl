package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// incubate_cards_a_test.go — the first slice of the remaining incubate
// cards (#3568, ADR 0185): one test per card.

const (
	assimilateEssenceOracle     = "a24d3d60-2bd7-40cb-8002-b092943f61fe"
	blightTitanOracle           = "f2c75577-c9be-48f3-8398-3020c79bd34e"
	blightedBurgeoningOracle    = "5e10fd35-9768-48e4-8cdb-95224ad33b9b"
	bloatedProcessorOracle      = "d1a26310-8c14-424b-8a7f-36793769c2a7"
	brimazBlightOracle          = "a6e7cbfc-6d5a-4a67-9dad-d512d3c97bda"
	chromeHostSeedsharkOracle   = "837c5d0e-0a33-4f9d-ae25-68a687bdf640"
	compleatedHuntmasterOracle  = "65b8bb9b-91f9-44e2-aa6e-e82e3760a4d9"
	corruptionOfTowashiOracle   = "242c39b4-5ef0-469a-a852-738f79d93bd6"
	essenceOfOrthodoxyOracle    = "51329e73-aebe-4f42-a4db-e092b1984a4b"
	furnaceGremlinOracle        = "bcbef58b-0c16-40f2-9b5c-14dfcc00f5e5"
	incubateACardsTestCreatures = "Creature — Phyrexian"
)

// pushIncubateACreature puts a catalog creature with the given printed
// size on the battlefield, ready to act.
func pushIncubateACreature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func TestAssimilateEssenceCountersAnUnpaidCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castFromSeat(t, g, them, "Assimilate Essence", "Instant", assimilateEssenceOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passUntilTaxed(t, g, me.ID)
	if !g.Stack.Contains(victim) {
		t.Fatal("the spell resolved while the {4} was unanswered")
	}
	answerPayUnless(t, g, me.ID, false)
	if g.Stack.Contains(victim) {
		t.Error("declining the {4} must counter the spell")
	}
	if n := len(incubatorsOf(g, them.ID)); n != 0 {
		t.Errorf("an unpaid counter made %d Incubators; the incubate is the paid branch", n)
	}
}

// castAssimilateAtTheirBear has seat 0 cast a creature spell and seat 1
// answer with Assimilate Essence, then passes until seat 0 is asked for
// the {4}. It returns the creature spell and the Assimilate Essence.
func castAssimilateAtTheirBear(t *testing.T, g *game.Game) (victim, assimilate uuid.UUID) {
	t.Helper()
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim = castCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	assimilate = castFromSeat(t, g, them, "Assimilate Essence", "Instant", assimilateEssenceOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passUntilTaxed(t, g, me.ID)
	return victim, assimilate
}

// Paying the {4} is the whole of "if they do": the spell resolves AND
// the caster incubates 2. The payer incubates nothing.
func TestAssimilateEssenceIncubatesTwoForTheCasterWhenTheTaxIsPaid(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	victim, _ := castAssimilateAtTheirBear(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if n := len(incubatorsOf(g, them.ID)); n != 0 {
		t.Fatalf("%d Incubators before the answer; the incubate waits for the payment", n)
	}
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("the spell is still on the stack after the payment")
	}
	if !battlefieldHas(g, victim) {
		t.Error("a paid spell must resolve, not be countered")
	}
	if got := theIncubator(t, g, them.ID).Counters["+1/+1"]; got != 2 {
		t.Errorf("the caster's Incubator has %d counters, want 2", got)
	}
	if n := len(incubatorsOf(g, me.ID)); n != 0 {
		t.Errorf("the payer got %d Incubators; \"you incubate 2\" is the caster", n)
	}
}

// "Pay" without the mana is a decline: countered, and no incubate.
func TestAssimilateEssenceTreatsAnUnaffordablePaymentAsADecline(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	victim, _ := castAssimilateAtTheirBear(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, victim) || g.Stack.Contains(victim) {
		t.Error("a payment that could not be made must still counter the spell")
	}
	if n := len(incubatorsOf(g, them.ID)); n != 0 {
		t.Errorf("an unaffordable payment made %d Incubators", n)
	}
}

func TestBlightTitanMillsTwoThenIncubatesForCreatureCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "")
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Milled One", TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Milled Two", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Blight Titan", "Creature — Phyrexian Giant", blightTitanOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 0 {
		t.Errorf("library = %d, want both cards milled", me.Library.Size())
	}
	// One creature card was already there and one was milled; the milled
	// sorcery does not count.
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 2 {
		t.Errorf("counters = %d, want 2 (creature cards in the graveyard after the mill)", got)
	}
}

func TestBlightedBurgeoningIncubatesTwoWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	castCatalogSpell(t, g, "Blighted Burgeoning", "Enchantment — Aura", blightedBurgeoningOracle, []game.TargetRef{pbCard(land)})
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 2 {
		t.Errorf("counters = %d, want 2", got)
	}
	spec, ok := Lookup(blightedBurgeoningOracle)
	if !ok || len(spec.ManaTriggers) != 1 {
		t.Fatalf("Blighted Burgeoning should carry its land mana trigger, got %+v", spec.ManaTriggers)
	}
}

func TestBloatedProcessorGrowsOnASacrificeAndIncubatesForItsLastKnownPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	proc := pushIncubateACreature(g, me.ID, "Bloated Processor", "Creature — Phyrexian", bloatedProcessorOracle, 3, 2)
	fodder := pushIncubateACreature(g, me.ID, "Fodder", incubateACardsTestCreatures, "", 1, 1)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, proc, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, fodder) {
		t.Error("the sacrificed Phyrexian is still on the battlefield")
	}
	if got := effectivePowerWithCounters(t, g, proc); got != 4 {
		t.Fatalf("power = %d, want 4 after the counter", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(proc) })
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 4 {
		t.Errorf("counters = %d, want 4 (its power when it died, counter included)", got)
	}
}

func TestBrimazBlightIncubatesForAPhyrexianCreatureSpellsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushIncubateACreature(g, me.ID, "Brimaz, Blight of Oreskos", "Legendary Creature — Phyrexian Cat", brimazBlightOracle, 3, 4)
	advanceToMain(t, g)
	bear := game.Card{InstanceID: uuid.New(), Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(bear)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(incubatorsOf(g, me.ID)); n != 0 {
		t.Fatalf("a non-Phyrexian, non-artifact creature spell incubated (%d Incubators)", n)
	}
	phy := game.Card{InstanceID: uuid.New(), Name: "Phyrexian Thing", TypeLine: "Creature — Phyrexian Horror", ManaCost: "{2}{B}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(phy)
	if err := g.CastSpell(me.ID, phy.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 3 {
		t.Errorf("counters = %d, want 3 (the spell's mana value)", got)
	}
}

func TestChromeHostSeedsharkIncubatesOnlyForNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shark := pushIncubateACreature(g, me.ID, "Chrome Host Seedshark", "Creature — Phyrexian Shark", chromeHostSeedsharkOracle, 2, 4)
	advanceToMain(t, g)
	if kw := effectiveAbilities(t, g, shark); len(kw) == 0 {
		t.Error("Chrome Host Seedshark should have flying")
	}
	bear := game.Card{InstanceID: uuid.New(), Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(bear)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(incubatorsOf(g, me.ID)); n != 0 {
		t.Fatalf("a creature spell incubated (%d)", n)
	}
	sorc := game.Card{InstanceID: uuid.New(), Name: "Plain Sorcery", TypeLine: "Sorcery", ManaCost: "{3}{U}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(sorc)
	if err := g.CastSpell(me.ID, sorc.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 4 {
		t.Errorf("counters = %d, want 4 (the sorcery's mana value)", got)
	}
}

func TestCompleatedHuntmasterSacrificesAnotherCreatureOrArtifactToIncubateThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hm := pushIncubateACreature(g, me.ID, "Compleated Huntmaster", "Creature — Phyrexian Elf Warrior", compleatedHuntmasterOracle, 2, 3)
	fodder := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	advanceToMain(t, g)
	// It may not sacrifice itself: "another".
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, hm, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{hm}}); err == nil {
		t.Fatal("the Huntmaster sacrificed itself to its own ability")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, hm, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, fodder) {
		t.Error("the sacrificed artifact is still on the battlefield")
	}
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 3 {
		t.Errorf("counters = %d, want 3", got)
	}
}

func TestCorruptionOfTowashiIncubatesFourAndDrawsOncePerTurnOnATransform(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Corruption of Towashi", "Enchantment", corruptionOfTowashiOracle, nil)
	passPriorityAroundTable(t, g)
	inc := theIncubator(t, g, me.ID)
	if inc.Counters["+1/+1"] != 4 {
		t.Fatalf("counters = %v, want four", inc.Counters)
	}
	handBefore := me.Hand.Size()
	transformIncubator(t, g, me, inc.InstanceID)
	towashiAnswer(t, g, me, true)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Fatalf("hand grew by %d on the first transform, want 1", got)
	}
	// A second permanent turning over the same turn draws nothing.
	second := pushIncubateACreature(g, me.ID, "Test DFC", "Creature — Human Werewolf", "", 1, 1)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventTransform, CardID: second, Actor: me.ID}) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand grew by %d after a second transform in one turn, want still 1", got)
	}
}

func TestEssenceOfOrthodoxyIncubatesTwoForItselfAndForAnotherPhyrexian(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Essence of Orthodoxy", "Creature — Phyrexian", essenceOfOrthodoxyOracle, nil)
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 2 {
		t.Fatalf("counters on entering = %d, want 2", got)
	}
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := len(incubatorsOf(g, me.ID)); n != 1 {
		t.Fatalf("a Bear entering made %d Incubators total, want still 1", n)
	}
	castCatalogSpell(t, g, "Other Phyrexian", "Creature — Phyrexian Horror", "", nil)
	passPriorityAroundTable(t, g)
	if n := len(incubatorsOf(g, me.ID)); n != 2 {
		t.Errorf("another Phyrexian entering gave %d Incubators total, want 2", n)
	}
}

func TestFurnaceGremlinIncubatesForItsPumpedPowerWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gremlin := pushIncubateACreature(g, me.ID, "Furnace Gremlin", "Creature — Phyrexian Gremlin", furnaceGremlinOracle, 1, 2)
	advanceToMain(t, g)
	for i := 0; i < 2; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "R"})
		if err := g.ActivateCatalogAbility(me.ID, gremlin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("pump %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
	}
	if got := effectivePowerWithCounters(t, g, gremlin); got != 3 {
		t.Fatalf("power = %d, want 3 after two pumps", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(gremlin) })
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 3 {
		t.Errorf("counters = %d, want 3 (its pumped power when it died)", got)
	}
}
