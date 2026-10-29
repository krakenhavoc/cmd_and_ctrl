package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hob_a10_cards_test.go — The Hobbit / The Hobbit Eternal, half A, tenth batch
// (#4454). Each test plays the real catalog entry and asserts what its text
// says.

const (
	hobA10BanquetOracle    = "2522a9dd-bcf3-4b84-abff-52b9e6dbde34"
	hobA10DancingOracle    = "5ba482e9-fbb0-4d9f-a3a9-414892bcdfed"
	hobA10GraspOracle      = "fcc9f3bc-0bc3-4189-a4d0-9698594b99a8"
	hobA10SmaugOracle      = "ce2fc562-7a71-4846-bfef-0a7c5f6c1ff3"
	hobA10RevelerOracle    = "11932191-4b19-49b1-bfe4-abb7b83b2e59"
	hobA10UncoverOracle    = "4c67359d-5590-4eb0-827e-b411967ee238"
	hobA10LordEaglesOracle = "3b38f314-d997-4c2a-b4c9-40dbe780993d"
	hobA10EaglesOracle     = "169a4a77-12ff-4997-9550-bdcd456aa2f7"
	hobA10BeornOracle      = "d9ed7252-11d1-432b-9101-ac08cd28826d"
)

func TestHobA10EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		hobA10BanquetOracle:    "Bag End Banquet",
		hobA10DancingOracle:    "Dancing from Dark to Dawn",
		hobA10GraspOracle:      "Enchanted River's Grasp",
		hobA10SmaugOracle:      "Smaug the Impenetrable",
		hobA10RevelerOracle:    "Silvan Reveler",
		hobA10UncoverOracle:    "Uncover the Moon-Letters",
		hobA10LordEaglesOracle: "The Lord of the Eagles",
		hobA10EaglesOracle:     "The Eagles Are Coming!",
		hobA10BeornOracle:      "Beorn's Hospitality",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// hobA10Tokens counts the tokens a player controls that have a subtype.
func hobA10Tokens(g *game.Game, controller uuid.UUID, subtype string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}

// --- Bag End Banquet ---------------------------------------------------------

func TestHobA10BagEndBanquetMakesThreeFoodAndTapsForOnePerFood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	banquet := castCatalogSpell(t, g, "Bag End Banquet", "Artifact", hobA10BanquetOracle, nil)
	passPriorityAroundTable(t, g)

	if got := hobA10Tokens(g, me.ID, "Food"); got != 3 {
		t.Fatalf("%d Food tokens, want 3", got)
	}
	b12Permanent(g, opp.ID, "Their Food", "Artifact — Food")
	if err := g.ActivateManaAbility(me.ID, banquet, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 3 || got[0] != "C" || got[1] != "C" || got[2] != "C" {
		t.Errorf("pool = %v, want [C C C] (my three Foods, not theirs)", got)
	}
	if !b16Tapped(t, g, banquet) {
		t.Error("the Banquet did not tap")
	}
}

func TestHobA10BagEndBanquetCountsEveryFoodYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	banquet := pushCatalogPermanent(g, me.ID, "Bag End Banquet", "Artifact", hobA10BanquetOracle, false)
	if err := g.ActivateManaAbility(me.ID, banquet, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 0 {
		t.Errorf("pool = %v, want nothing with no Food", got)
	}
	if !b16Tapped(t, g, banquet) {
		t.Error("the Banquet should still tap for no mana")
	}
}

// --- Dancing from Dark to Dawn -----------------------------------------------

func TestHobA10DancingPutsCountersEqualToTheCreatureSpellsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Dancing from Dark to Dawn", "Enchantment", hobA10DancingOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	assistCardCastX(t, g, "Test Giant", "Creature — Giant", "", "{2}{G}", []string{"G", "G", "G"}, 0, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)

	if n := hobA8Counters(t, g, bear); n != 3 {
		t.Errorf("a mana value 3 creature spell put %d counters, want 3", n)
	}
	if n := hobA8Counters(t, g, theirs); n != 0 {
		t.Errorf("an opponent's creature got %d counters", n)
	}
}

func TestHobA10DancingReadsTheAnnouncedXOfTheSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Dancing from Dark to Dawn", "Enchantment", hobA10DancingOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	assistCardCastX(t, g, "Test Hydra", "Creature — Hydra", "", "{X}{G}", []string{"G", "G", "G"}, 2, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)

	if n := hobA8Counters(t, g, bear); n != 3 {
		t.Errorf("an {X}{G} spell cast with X=2 put %d counters, want 3", n)
	}
}

func TestHobA10DancingIgnoresNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Dancing from Dark to Dawn", "Enchantment", hobA10DancingOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	assistCardCastX(t, g, "Test Sorcery", "Sorcery", "", "{2}{G}", []string{"G", "G", "G"}, 0, nil)
	assistCardSettle(t, g, me.ID, nil, true)

	if n := hobA8Counters(t, g, bear); n != 0 {
		t.Errorf("a sorcery put %d counters, want 0", n)
	}
}

func TestHobA10DancingLandfallMakesABearForEachLandYouPlay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Dancing from Dark to Dawn", "Enchantment", hobA10DancingOracle, false)

	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := hobA10Tokens(g, me.ID, "Bear"); got != 1 {
		t.Fatalf("%d Bear tokens after a land, want 1", got)
	}
	var bear game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Bear") {
			bear = c
		}
	}
	if bear.CurrentPower() != 2 || bear.CurrentToughness() != 2 || len(bear.Colors) != 1 || bear.Colors[0] != "G" {
		t.Errorf("Bear token is %d/%d %v, want a 2/2 green Bear", bear.CurrentPower(), bear.CurrentToughness(), bear.Colors)
	}
	// An opponent's land is not yours.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: opp.ID, CardID: uuid.New()})
	})
	passPriorityAroundTable(t, g)
	if got := hobA10Tokens(g, me.ID, "Bear"); got != 1 {
		t.Errorf("%d Bear tokens after an opponent's event, want still 1", got)
	}
}

// --- Enchanted River's Grasp -------------------------------------------------

func TestHobA10GraspTapsTheCreatureRemovesItsCountersAndSilencesIt(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	flyer := pushSizedCreature(g, opp.ID, "Their Hawk", 2, 2, "flying")
	g.WithWriteLock(func() {
		c := findBattlefieldCardByID(g, flyer)
		c.Counters = map[string]int{game.CounterPlusOne: 2, "shield": 1}
	})

	castCatalogSpell(t, g, "Enchanted River's Grasp", "Enchantment — Aura", hobA10GraspOracle, s58Refs(flyer))
	passPriorityAroundTable(t, g)

	if !b16Tapped(t, g, flyer) {
		t.Error("the enchanted creature did not tap")
	}
	c := apaLive(g, flyer)
	if n := c.Counters[game.CounterPlusOne] + c.Counters["shield"]; n != 0 {
		t.Errorf("counters left on the creature: %v, want none", c.Counters)
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("creature is %d/%d, want its printed 2/2 with the counters gone", c.CurrentPower(), c.CurrentToughness())
	}
	if hasKeywordOnBattlefield(t, g, flyer, "flying") {
		t.Error("the enchanted creature should have lost flying")
	}
}

func TestHobA10GraspKeepsTheCreatureTappedThroughItsControllersUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Enchanted River's Grasp", "Enchantment — Aura", hobA10GraspOracle, s58Refs(victim))
	passPriorityAroundTable(t, g)

	advanceToPrecombatMainOf(t, g, 1)
	if !b16Tapped(t, g, victim) {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

func TestHobA10GraspCanOnlyEnchantACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	if err := castCatalogSpellErr(t, g, "Enchanted River's Grasp", "Enchantment — Aura", hobA10GraspOracle, s58Refs(land)); err == nil {
		t.Error("the Aura was cast on a land")
	}
}

// --- Smaug the Impenetrable --------------------------------------------------

func hobA10Smaug(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castHoldCreature(t, g, "Smaug the Impenetrable", "Legendary Creature — Dragon", hobA10SmaugOracle, 8, 7)
	passPriorityAroundTable(t, g)
	return id
}

func TestHobA10SmaugMakesATreasureForEachPointOfNoncombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	smaug := hobA10Smaug(t, g)
	for _, k := range []string{"flying", "indestructible", "haste"} {
		if !hasKeywordOnBattlefield(t, g, smaug, k) {
			t.Errorf("Smaug should have %s", k)
		}
	}

	before := countTreasuresControlledBy(g, me.ID)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.New(), smaug, 3); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID) - before; got != 3 {
		t.Errorf("3 noncombat damage made %d Treasures, want 3", got)
	}
	if got := countTreasuresControlledBy(g, opp.ID); got != 0 {
		t.Errorf("the damager's controller got %d Treasures", got)
	}

	// Lethal damage does not kill an indestructible creature, and still pays.
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.New(), smaug, 7); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, smaug); !ok {
		t.Fatal("Smaug died to damage")
	}
	if got := countTreasuresControlledBy(g, me.ID) - before; got != 10 {
		t.Errorf("3 and 7 noncombat damage made %d Treasures, want 10", got)
	}
}

func TestHobA10SmaugIgnoresCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	smaug := hobA10Smaug(t, g)
	before := countTreasuresControlledBy(g, me.ID)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: uuid.New(), Target: smaug, Amount: 4, Combat: true})
	})
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID) - before; got != 0 {
		t.Errorf("combat damage made %d Treasures, want 0", got)
	}
}

// --- Silvan Reveler ----------------------------------------------------------

func TestHobA10SilvanRevelerDiscardingALandPutsItOntoTheBattlefieldTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	land := seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	handBefore := me.Hand.Size()

	castHoldCreature(t, g, "Silvan Reveler", "Creature — Elf Citizen", hobA10RevelerOracle, 3, 2)
	// Hand: the Reveler left, the draw came in.
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Fatalf("hand = %d after the draw, want %d", got, handBefore+1)
	}
	answerDiscard(t, g, me.ID, land)
	passPriorityAroundTable(t, g)

	c, ok := battlefieldCard(g, land)
	if !ok {
		t.Fatal("the discarded land did not come onto the battlefield")
	}
	if !c.Tapped {
		t.Error("the land should enter tapped")
	}
	if me.Graveyard.Contains(land) {
		t.Error("the land is still in the graveyard")
	}
}

func TestHobA10SilvanRevelerDiscardingANonlandLeavesItInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	spell := seedLibraryTop(me, "Test Sorcery", "Sorcery", "{1}")

	castHoldCreature(t, g, "Silvan Reveler", "Creature — Elf Citizen", hobA10RevelerOracle, 3, 2)
	answerDiscard(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(spell) {
		t.Error("the discarded nonland card should be in the graveyard")
	}
	if _, ok := battlefieldCard(g, spell); ok {
		t.Error("a nonland card was put onto the battlefield")
	}
}

func hobA10RevelerInGraveyard(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return seedGraveyardCard(t, g, "Silvan Reveler", "Creature — Elf Citizen", hobA10RevelerOracle)
}

func TestHobA10SilvanRevelerLandfallFromTheGraveyardMayPayToReturnItToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	reveler := hobA10RevelerInGraveyard(t, g)

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	b06AddMana(me, "G", "G", "U")
	prompt := answerPayUnless(t, g, me.ID, true)
	if prompt.PayCost != "{1}{G}{U}" {
		t.Errorf("payment = %q, want {1}{G}{U}", prompt.PayCost)
	}
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(reveler) || me.Graveyard.Contains(reveler) {
		t.Errorf("paying did not return the Reveler to hand (hand=%t graveyard=%t)", me.Hand.Contains(reveler), me.Graveyard.Contains(reveler))
	}
	if _, ok := battlefieldCard(g, reveler); ok {
		t.Error("the Reveler went to the battlefield, not the hand")
	}
}

func TestHobA10SilvanRevelerLandfallMayBeDeclinedAndNeedsTheMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	reveler := hobA10RevelerInGraveyard(t, g)

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(reveler) {
		t.Fatal("declining moved the Reveler out of the graveyard")
	}

	// With no mana to spend, saying yes pays nothing and returns nothing.
	g.Seats[0].LandDropsPerTurn++
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		answerPayUnless(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
	}
	if !me.Graveyard.Contains(reveler) {
		t.Error("the Reveler came back without {1}{G}{U} being paid")
	}
}

func TestHobA10SilvanRevelerLandfallDoesNothingOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Silvan Reveler", "Creature — Elf Citizen", hobA10RevelerOracle, false)
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the Reveler asked about returning itself while on the battlefield")
	}
}

// --- Uncover the Moon-Letters ------------------------------------------------

func hobA10Uncover(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Uncover the Moon-Letters", "Enchantment", hobA10UncoverOracle, false)
}

func TestHobA10UncoverDrawsXThenDiscardsTwoForTheManaSpent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hobA10Uncover(g, me.ID)

	assistCardCastX(t, g, "Test Sorcery", "Sorcery", "", "{2}{U}", []string{"U", "U", "U"}, 0, nil)
	handAfterCast := me.Hand.Size()
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if got := me.Hand.Size(); got != handAfterCast+3 {
		t.Fatalf("hand = %d after drawing, want %d (three mana was spent)", got, handAfterCast+3)
	}
	if got := discardOwed(g, me.ID); got != 2 {
		t.Fatalf("owed a %d-card discard, want 2", got)
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handAfterCast+1 {
		t.Errorf("hand = %d, want %d: drew three, discarded two", got, handAfterCast+1)
	}
}

func TestHobA10UncoverDecliningTheDrawDeclinesTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hobA10Uncover(g, me.ID)

	assistCardCastX(t, g, "Test Instant", "Instant", "", "{1}{U}", []string{"U", "U"}, 0, nil)
	handAfterCast := me.Hand.Size()
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handAfterCast {
		t.Errorf("hand = %d, want %d after declining", got, handAfterCast)
	}
	if discardOwed(g, me.ID) != 0 {
		t.Error("a declined draw left a discard prompt open")
	}
}

func TestHobA10UncoverIgnoresCreatureSpellsAndOpponentsSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hobA10Uncover(g, me.ID)

	assistCardCastX(t, g, "Test Giant", "Creature — Giant", "", "{2}{G}", []string{"G", "G", "G"}, 0, nil)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("a creature spell offered the draw")
	}

	// Another seat's noncreature spell is not "you cast".
	g2 := newCatalogGame(t)
	mine := g2.Seats[0]
	advanceToMain(t, g2)
	hobA10Uncover(g2, g2.Seats[1].ID)
	assistCardCastX(t, g2, "Test Sorcery", "Sorcery", "", "{2}{U}", []string{"U", "U", "U"}, 0, nil)
	passPriorityAroundTable(t, g2)
	if latestConfirmFor(g2, mine.ID) != nil || latestConfirmFor(g2, g2.Seats[1].ID) != nil {
		t.Error("an Uncover the Moon-Letters I do not control offered a draw for my spell")
	}
}

// --- The Lord of the Eagles --------------------------------------------------

func hobA10LordPrice(t *testing.T, g *game.Game, me *game.Player) int {
	t.Helper()
	return selfPricedMV(t, g, me, hobA10LordEaglesOracle, "Legendary Creature — Bird Noble", "{7}{U}{U}")
}

func TestHobA10LordOfTheEaglesCostsLessByTheTotalPowerOfYourFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	if mv := hobA10LordPrice(t, g, me); mv != 9 {
		t.Fatalf("with no flyers the Lord costs %d, want 9", mv)
	}
	pushSizedCreature(g, me.ID, "Hawk", 3, 3, "flying")
	pushSizedCreature(g, me.ID, "Roc", 2, 4, "flying")
	pushSizedCreature(g, me.ID, "Ogre", 5, 5)
	pushSizedCreature(g, opp.ID, "Their Dragon", 6, 6, "flying")
	if mv := hobA10LordPrice(t, g, me); mv != 4 {
		t.Errorf("with 3 + 2 power of flyers the Lord costs %d, want 4 (the Ogre and their Dragon do not count)", mv)
	}
}

func TestHobA10LordOfTheEaglesNeverCostsLessThanItsColoredMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushSizedCreature(g, me.ID, "Great Eagle", 20, 20, "flying")
	if mv := hobA10LordPrice(t, g, me); mv != 2 {
		t.Errorf("with 20 power of flyers the Lord costs %d, want 2 ({U}{U})", mv)
	}
}

func TestHobA10LordOfTheEaglesCountsFlyersByTheirCurrentPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hawk := pushSizedCreature(g, me.ID, "Hawk", 1, 1, "flying")
	g.WithWriteLock(func() {
		findBattlefieldCardByID(g, hawk).Counters = map[string]int{game.CounterPlusOne: 2}
	})
	if mv := hobA10LordPrice(t, g, me); mv != 6 {
		t.Errorf("a 1/1 flyer with two +1/+1 counters made the Lord cost %d, want 6", mv)
	}
}

func TestHobA10LordOfTheEaglesHasFlashAndFlying(t *testing.T) {
	spec, ok := Lookup(hobA10LordEaglesOracle)
	if !ok {
		t.Fatal("not registered")
	}
	got := map[string]bool{}
	for _, k := range spec.PrintedKeywords {
		got[k] = true
	}
	if !got["flash"] || !got["flying"] {
		t.Errorf("printed keywords = %v, want flash and flying", spec.PrintedKeywords)
	}
}

// --- The Eagles Are Coming! --------------------------------------------------

func hobA10CastEagles(t *testing.T, g *game.Game, targets []game.TargetRef, kicked bool) error {
	t.Helper()
	params := game.CastSpellParams{Targets: targets}
	if kicked {
		params.OptionalCosts = []int{0}
	}
	_, err := hobA4Cast(t, g, "The Eagles Are Coming!", "Instant", hobA10EaglesOracle, params)
	return err
}

// hobA10AdvanceToUpkeep walks to the next upkeep and lets its triggers resolve.
func hobA10AdvanceToUpkeep(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 40; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep {
			passPriorityAroundTable(t, g)
			return
		}
	}
	t.Fatal("never reached an upkeep")
}

func TestHobA10EaglesAreComingReturnsOneCreatureAndMakesABirdAtTheNextUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)

	if err := hobA10CastEagles(t, g, s58Refs(bear), false); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Fatal("the chosen creature did not return to hand")
	}
	if _, ok := battlefieldCard(g, other); !ok {
		t.Error("the other creature left the battlefield")
	}
	if got := hobA10Tokens(g, me.ID, "Bird"); got != 0 {
		t.Fatalf("%d Birds before the next upkeep, want 0", got)
	}

	hobA10AdvanceToUpkeep(t, g)
	if got := hobA10Tokens(g, me.ID, "Bird"); got != 1 {
		t.Fatalf("%d Birds at the next upkeep, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Bird") {
			if c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !c.HasSubtype("Soldier") || !game.HasKeyword(&c, "flying") {
				t.Errorf("Bird token is %d/%d, soldier=%t, flying=%t; want a 4/4 flying Bird Soldier",
					c.CurrentPower(), c.CurrentToughness(), c.HasSubtype("Soldier"), game.HasKeyword(&c, "flying"))
			}
			if len(c.Colors) != 1 || c.Colors[0] != "W" {
				t.Errorf("Bird colours = %v, want white", c.Colors)
			}
		}
	}
	hobA10AdvanceToUpkeep(t, g)
	if got := hobA10Tokens(g, me.ID, "Bird"); got != 1 {
		t.Errorf("%d Birds after a second upkeep, want still 1 (the trigger is once)", got)
	}
}

func TestHobA10EaglesAreComingKickedReturnsAnyNumberAndMakesABirdForEach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	c := b12Creature(g, me.ID, "Cat", "Creature — Cat", 2, 2)

	if err := hobA10CastEagles(t, g, s58Refs(a, b, c), true); err != nil {
		t.Fatalf("CastSpell kicked: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !me.Hand.Contains(id) {
			t.Errorf("creature %s did not return to hand", id)
		}
	}
	hobA10AdvanceToUpkeep(t, g)
	if got := hobA10Tokens(g, me.ID, "Bird"); got != 3 {
		t.Errorf("%d Birds for three returned creatures, want 3", got)
	}
}

func TestHobA10EaglesAreComingKickedMayChooseNoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if err := hobA10CastEagles(t, g, nil, true); err != nil {
		t.Fatalf("a kicked cast with no targets was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, bear); !ok {
		t.Error("an untargeted spell returned a creature")
	}
	hobA10AdvanceToUpkeep(t, g)
	if got := hobA10Tokens(g, me.ID, "Bird"); got != 0 {
		t.Errorf("%d Birds for no creatures, want 0", got)
	}
}

func TestHobA10EaglesAreComingUnkickedTakesExactlyOneCreatureYouOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	if err := hobA10CastEagles(t, g, s58Refs(a, b), false); err == nil {
		t.Error("an unkicked cast named two creatures")
	}
	if err := hobA10CastEagles(t, g, nil, false); err == nil {
		t.Error("an unkicked cast named no creature")
	}
	if err := hobA10CastEagles(t, g, s58Refs(theirs), false); err == nil {
		t.Error("an unkicked cast named a creature I do not own")
	}
	if err := hobA10CastEagles(t, g, s58Refs(theirs), true); err == nil {
		t.Error("a kicked cast named a creature I do not own")
	}
}

func TestHobA10EaglesAreComingCanChooseACreatureYouOwnButDoNotControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	stolen := b12Creature(g, opp.ID, "Stolen Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		findBattlefieldCardByID(g, stolen).Owner = me.ID
	})

	if err := hobA10CastEagles(t, g, s58Refs(stolen), false); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(stolen) {
		t.Error("the creature I own did not return to my hand")
	}
	hobA10AdvanceToUpkeep(t, g)
	if got := hobA10Tokens(g, me.ID, "Bird"); got != 1 {
		t.Errorf("%d Birds, want 1 under my control", got)
	}
}

// A token you own does move to your hand (and then ceases to exist as a
// state-based action), so it is a creature "returned to your hand this way".
func TestHobA10EaglesAreComingCountsATokenItReturns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, TokenCard("2/2 green Bear"), 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Controller == me.ID {
			token = c.InstanceID
		}
	}
	if err := hobA10CastEagles(t, g, s58Refs(token), false); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, token); ok {
		t.Fatal("the token is still on the battlefield")
	}
	hobA10AdvanceToUpkeep(t, g)
	if got := hobA10Tokens(g, me.ID, "Bird"); got != 1 {
		t.Errorf("%d Birds for a returned token, want 1", got)
	}
}

// --- Beorn's Hospitality -----------------------------------------------------

func hobA10Beorn(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Beorn's Hospitality", "Enchantment", hobA10BeornOracle, false)
}

func TestHobA10BeornLandfallPutsACounterOnTargetCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hobA10Beorn(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if n := hobA8Counters(t, g, bear); n != 1 {
		t.Errorf("landfall put %d counters on the chosen creature, want 1", n)
	}
	if n := hobA8Counters(t, g, theirs); n != 0 {
		t.Errorf("an opponent's creature got %d counters", n)
	}
}

func TestHobA10BeornBecomesABearCreatureWithPowerAndToughnessEqualToYourLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	beorn := hobA10Beorn(g, me.ID)
	for i := 0; i < 3; i++ {
		b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")

	if apaLive(g, beorn).IsCreature() {
		t.Fatal("the enchantment is a creature before it is activated")
	}
	floatMana(t, g, me, "{G}{G}{G}{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, beorn, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	c := apaLive(g, beorn)
	if !c.IsCreature() || !c.HasSubtype("Bear") {
		t.Fatalf("after activating it is creature=%t Bear=%t, want a Bear creature", c.IsCreature(), c.HasSubtype("Bear"))
	}
	if !c.IsEnchantment() {
		t.Error("it should still be an enchantment (in addition to its other types)")
	}
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("it is %d/%d with three lands of mine and one of theirs, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}

	// Live: another land grows it.
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	c = apaLive(g, beorn)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("with a fourth land it is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestHobA10BeornTheChangeDoesNotEndAndCountersStackOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	beorn := hobA10Beorn(g, me.ID)
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	floatMana(t, g, me, "{G}{G}{G}{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, beorn, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	advanceToNextSeatsTurn(t, g)
	advanceToPrecombatMainOf(t, g, 0)
	c := apaLive(g, beorn)
	if !c.IsCreature() || c.CurrentPower() != 2 {
		t.Errorf("a turn later it is creature=%t power=%d, want a 2-power creature (the effect does not end)", c.IsCreature(), c.CurrentPower())
	}

	// A +1/+1 counter lands on top of the set base (layer 7d after 7b).
	g.WithWriteLock(func() {
		findBattlefieldCardByID(g, beorn).Counters = map[string]int{game.CounterPlusOne: 1}
	})
	if c = apaLive(g, beorn); c.CurrentPower() != 3 {
		t.Errorf("with a +1/+1 counter its power is %d, want 3", c.CurrentPower())
	}
}
