package effects

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hob_a3_cards_test.go — The Hobbit / The Hobbit Eternal, half A, third
// batch (#4454): ten cards on seams that exist. Each test casts or plays the
// real catalog entry and asserts what the card's text says.

const (
	hobA3GiantsBoulderOracle   = "80433c2c-5196-42e0-b9fa-d7b12783f5fd"
	hobA3OldThrushOracle       = "c8259b0e-131f-44f0-86d4-7c4d711b328d"
	hobA3DwarvenWarriorsOracle = "cfc553cd-3b4c-47a9-bffb-e5790befb32c"
	hobA3FireOfOrthancOracle   = "407021e2-4072-4961-95be-f9309c341677"
	hobA3GoblinFireleaperOID   = "c5fa1631-8685-49e1-baad-86ef72a9499c"
	hobA3WestfoldRiderOracle   = "17d35b8f-2189-4587-9c57-425e93d883c6"
	hobA3ErrandRiderOracle     = "00749bad-dded-4fec-bf34-ff14cf561c26"
	hobA3FiendHunterOracle     = "cb9d557a-fc06-428c-8be6-7d28add33028"
	hobA3LandrovalOracle       = "467c9fe6-d1d3-4242-bc0c-e522b4f21d56"
	hobA3EagleOracle           = "cb4b4b63-682b-4439-9757-84cca7a27db2"
)

func TestHobA3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		hobA3GiantsBoulderOracle:   "Giant's Boulder",
		hobA3OldThrushOracle:       "Old Thrush",
		hobA3DwarvenWarriorsOracle: "Dwarven Warriors",
		hobA3FireOfOrthancOracle:   "Fire of Orthanc",
		hobA3GoblinFireleaperOID:   "Goblin Fireleaper",
		hobA3WestfoldRiderOracle:   "Westfold Rider",
		hobA3ErrandRiderOracle:     "Errand-Rider of Gondor",
		hobA3FiendHunterOracle:     "Fiend Hunter",
		hobA3LandrovalOracle:       "Landroval, Horizon Witness",
		hobA3EagleOracle:           "Eagle of the Great Shelf",
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

// --- Giant's Boulder ---------------------------------------------------

func TestHobA3GiantsBoulderScriesTwoWhenItEnters(t *testing.T) {
	hobA1ScryOnEntry(t, "Giant's Boulder", "Artifact", hobA3GiantsBoulderOracle, 2)
}

func TestHobA3GiantsBoulderPaysOneToTapForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	boulder := pushCatalogPermanent(g, me.ID, "Giant's Boulder", "Artifact", hobA3GiantsBoulderOracle, false)

	if err := g.ActivateManaAbility(me.ID, boulder, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Fatal("the fixing tap needs {1} to pay; it ran with an empty pool")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, boulder, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want [R]: the {1} was spent, not produced", got)
	}
	if !b16Tapped(t, g, boulder) {
		t.Error("the ability has a tap cost")
	}
}

func TestHobA3GiantsBoulderSacrificesForSevenToDestroyAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boulder := pushCatalogPermanent(g, me.ID, "Giant's Boulder", "Artifact", hobA3GiantsBoulderOracle, false)
	victim := pushVanillaCreature(g, opp.ID, "Their Ogre", 4, 4)
	advanceToMain(t, g)

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	b16Activate(t, g, me.ID, boulder, 0, game.ActivateAbilityParams{Targets: cardRefs(victim)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target survived")
	}
	if g.Battlefield.Contains(boulder) {
		t.Error("the Boulder was not sacrificed")
	}
}

// --- Old Thrush ----------------------------------------------------------

// hobA3ThrushTable seeds a library of two basics and two cards that are not,
// then casts the Thrush and resolves its trigger up to the search prompt.
func hobA3ThrushTable(t *testing.T) (*game.Game, *game.Player, int) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Command Tower", "Land"),
		searchTestLand("Sol Ring", "Artifact"),
	)
	life := me.Life
	castHoldCreature(t, g, "Old Thrush", "Creature — Bird", hobA3OldThrushOracle, 1, 2)
	passPriorityAroundTable(t, g)
	return g, me, life
}

func TestHobA3OldThrushGainsTwoAndPutsABasicLandOnTop(t *testing.T) {
	g, me, life := hobA3ThrushTable(t)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d: the two life is gained whatever the search does", me.Life, life+2)
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger did not offer a search")
	}
	for _, n := range []string{"Command Tower", "Sol Ring"} {
		if searchOptionNamed(g, c, n) != uuid.Nil {
			t.Errorf("%s is not a basic land card and must not be offered", n)
		}
	}
	answerSearchNamed(t, g, me.ID, "Island")
	passPriorityAroundTable(t, g)
	top := me.Library.Cards[len(me.Library.Cards)-1]
	if top.Name != "Island" {
		t.Fatalf("the top card is %q, want the Island put there after the shuffle", top.Name)
	}
	if me.Library.Size() != 4 {
		t.Errorf("library = %d, the card never left it", me.Library.Size())
	}
}

func TestHobA3OldThrushMayDeclineTheSearchAndKeepsTheLifeGain(t *testing.T) {
	g, me, life := hobA3ThrushTable(t)
	before := libraryIDs(me)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger did not offer a search")
	}
	if err := answerSearchApply(g, c, me.ID, me.ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !sameIDList(libraryIDs(me), before) {
		t.Error("the library changed after a declined search")
	}
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
}

// --- Dwarven Warriors ------------------------------------------------------

func TestHobA3DwarvenWarriorsMakesAPowerTwoCreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	warriors := pushCatalogPermanent(g, me.ID, "Dwarven Warriors", "Creature — Dwarf Warrior", hobA3DwarvenWarriorsOracle, false)
	small := pushRestrictionBear(g, me, "Small Beast") // 2/2
	other := pushRestrictionBear(g, me, "Other Beast")
	advanceToMain(t, g)

	b16Activate(t, g, me.ID, warriors, 0, game.ActivateAbilityParams{Targets: cardRefs(small)})
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, small, game.CantBeBlocked)
	assertRestrictions(t, g, other, 0)
	if !b16Tapped(t, g, warriors) {
		t.Error("the ability has a tap cost")
	}
}

func TestHobA3DwarvenWarriorsRefusesAPowerThreeCreatureButMayTargetItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	warriors := pushCatalogPermanent(g, me.ID, "Dwarven Warriors", "Creature — Dwarf Warrior", hobA3DwarvenWarriorsOracle, false)
	big := b12Creature(g, me.ID, "Big Beast", "Creature — Beast", 3, 3)
	advanceToMain(t, g)

	err := g.ActivateCatalogAbility(me.ID, warriors, 0, game.ActivateAbilityParams{Targets: cardRefs(big)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a power-3 creature: err = %v, want ErrIllegalTarget", err)
	}
	b16Activate(t, g, me.ID, warriors, 0, game.ActivateAbilityParams{Targets: cardRefs(warriors)})
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, warriors, game.CantBeBlocked)
}

// --- Fire of Orthanc -------------------------------------------------------

func TestHobA3FireOfOrthancDestroysAnArtifactOrALandButNotACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Sol Ring", "Artifact")
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)

	if err := castCatalogSpellErr(t, g, "Fire of Orthanc", "Sorcery", hobA3FireOfOrthancOracle, cardRefs(bear)); err == nil {
		t.Fatal("Fire of Orthanc targeted a creature")
	}
	castCatalogSpell(t, g, "Fire of Orthanc", "Sorcery", hobA3FireOfOrthancOracle, cardRefs(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact should be destroyed")
	}
	castCatalogSpell(t, g, "Fire of Orthanc", "Sorcery", hobA3FireOfOrthancOracle, cardRefs(land))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !opp.Graveyard.Contains(land) {
		t.Error("the land should be destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was not a target and should be untouched")
	}
}

func TestHobA3FireOfOrthancStopsCreaturesWithoutFlyingFromBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := pushSizedCreature(g, me.ID, "My Attacker", 3, 3)
	ground := pushSizedCreature(g, opp.ID, "Their Bear", 2, 2)
	flyer := pushSizedCreature(g, opp.ID, "Their Flyer", 1, 1, "flying")
	rock := b12Permanent(g, opp.ID, "Sol Ring", "Artifact")

	castCatalogSpell(t, g, "Fire of Orthanc", "Sorcery", hobA3FireOfOrthancOracle, cardRefs(rock))
	passPriorityAroundTable(t, g)
	late := pushSizedCreature(g, opp.ID, "Late Bear", 2, 2)

	assertRestrictions(t, g, ground, game.CantBlock)
	assertRestrictions(t, g, late, game.CantBlock)
	assertRestrictions(t, g, flyer, 0)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(ground, attacker); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("a creature without flying blocked: err = %v, want ErrIllegalBlock", err)
	}
	if err := g.DeclareBlocker(flyer, attacker); err != nil {
		t.Errorf("a creature with flying may still block: %v", err)
	}
}

// --- Goblin Fireleaper -----------------------------------------------------

func TestHobA3GoblinFireleaperPumpsForOneRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Goblin Fireleaper", "Creature — Goblin Warrior", hobA3GoblinFireleaperOID, false)
	advanceToMain(t, g)
	if got := b31CurrentPowerOf(g, id); got != 1 {
		t.Fatalf("base power = %d, want 1", got)
	}
	for i := 0; i < 2; i++ {
		b06AddMana(me, "R", "R")
		b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
	}
	if got := b31CurrentPowerOf(g, id); got != 3 {
		t.Errorf("power after two pumps = %d, want 3", got)
	}
}

func TestHobA3GoblinFireleaperDiesForItsPumpedPowerToAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushCatalogPermanent(g, me.ID, "Goblin Fireleaper", "Creature — Goblin Warrior", hobA3GoblinFireleaperOID, false)
	mine := pushVanillaCreature(g, me.ID, "My Ogre", 5, 5)
	victim := pushVanillaCreature(g, opp.ID, "Their Ogre", 5, 5)
	advanceToMain(t, g)
	b06AddMana(me, "R", "R")
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := b31CurrentPowerOf(g, id); got != 2 {
		t.Fatalf("pumped power = %d, want 2", got)
	}

	eerieDestroy(t, g, id)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)
	if got := damageMarkedOn(g, victim); got != 2 {
		t.Errorf("damage on the target = %d, want the 2 power it died with", got)
	}
	if got := damageMarkedOn(g, mine); got != 0 {
		t.Errorf("damage on my own creature = %d, want 0", got)
	}
}

func TestHobA3GoblinFireleaperCanOnlyTargetACreatureAnOpponentControls(t *testing.T) {
	spec, _ := Lookup(hobA3GoblinFireleaperOID)
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	theirLand := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	rfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{mine: false, theirs: true, theirLand: false})
}

// --- Westfold Rider --------------------------------------------------------

func TestHobA3WestfoldRiderSacrificesToDestroyAnEnchantmentOrArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := pushCatalogPermanent(g, me.ID, "Westfold Rider", "Creature — Human Knight", hobA3WestfoldRiderOracle, false)
	aura := b12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	advanceToMain(t, g)

	err := g.ActivateCatalogAbility(me.ID, rider, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a creature: err = %v, want ErrIllegalTarget", err)
	}
	b16Activate(t, g, me.ID, rider, 0, game.ActivateAbilityParams{Targets: cardRefs(aura)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment survived")
	}
	if g.Battlefield.Contains(rider) || !me.Graveyard.Contains(rider) {
		t.Error("the Rider is the cost and should be in the graveyard")
	}
}

func TestHobA3WestfoldRiderDestroysAnArtifactToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := pushCatalogPermanent(g, me.ID, "Westfold Rider", "Creature — Human Knight", hobA3WestfoldRiderOracle, false)
	rock := b12Permanent(g, opp.ID, "Sol Ring", "Artifact")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, rider, 0, game.ActivateAbilityParams{Targets: cardRefs(rock)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestHobA3WestfoldRiderActivatesOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rider := pushCatalogPermanent(g, me.ID, "Westfold Rider", "Creature — Human Knight", hobA3WestfoldRiderOracle, false)
	aura := b12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, game.ActivateAbilityParams{Targets: cardRefs(aura)}); err == nil {
		t.Fatal("the ability ran outside a main phase")
	}
	if !g.Battlefield.Contains(rider) || !g.Battlefield.Contains(aura) {
		t.Error("a refused activation must not cost the Rider or touch the target")
	}
}

// --- Errand-Rider of Gondor --------------------------------------------------

func TestHobA3ErrandRiderDrawsThenTucksACardWithoutALegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Three")
	keep := handCardForTest(me, "Keep", "Instant", "")
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Errand-Rider of Gondor", "Creature — Human Soldier", hobA3ErrandRiderOracle, 3, 2)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand = %d, want %d after the draw", got, hand+1)
	}
	ask := chooseCardsChoiceFor(g, me.ID)
	if ask == nil {
		t.Fatal("no prompt to put a card from your hand on the bottom")
	}
	if len(ask.ChooseCards) != me.Hand.Size() {
		t.Errorf("offered %d cards, want the whole hand (%d)", len(ask.ChooseCards), me.Hand.Size())
	}
	answerChooseCards(t, g, me.ID, keep)
	if me.Hand.Contains(keep) {
		t.Error("the chosen card is still in hand")
	}
	if n := len(me.Library.Cards); n == 0 || me.Library.Cards[0].InstanceID != keep {
		t.Error("the chosen card is not on the bottom of the library")
	}
}

func TestHobA3ErrandRiderOnlyDrawsWhileYouControlALegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Three")
	handCardForTest(me, "Keep", "Instant", "")
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Aragorn", TypeLine: "Legendary Creature — Human Noble",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Errand-Rider of Gondor", "Creature — Human Soldier", hobA3ErrandRiderOracle, 3, 2)
	passPriorityAroundTable(t, g)

	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked to put a card on the bottom while a legendary creature is controlled")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d: the draw only", got, hand+1)
	}
}

// With nothing else in hand, the card just drawn is the only one on offer.
func TestHobA3ErrandRiderBottomsTheDrawnCardWhenItIsAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	drawn := seedLibrary(me, "Drawn")[0]
	castHoldCreature(t, g, "Errand-Rider of Gondor", "Creature — Human Soldier", hobA3ErrandRiderOracle, 3, 2)
	passPriorityAroundTable(t, g)
	ask := chooseCardsChoiceFor(g, me.ID)
	if ask == nil || len(ask.ChooseCards) != 1 || ask.ChooseCards[0] != drawn {
		t.Fatalf("want a prompt offering only the drawn card, got %+v", ask)
	}
	answerChooseCards(t, g, me.ID, drawn)
	if me.Hand.Size() != 0 || me.Library.Cards[0].InstanceID != drawn {
		t.Error("the drawn card should be back on the bottom of the library")
	}
}

// --- Fiend Hunter -------------------------------------------------------------

func TestHobA3FiendHunterExilesACreatureAndReturnsItUnderItsOwnersControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// Stolen: mine to control, theirs to own.
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bear).Controller = me.ID })
	hunter := castHoldCreature(t, g, "Fiend Hunter", "Creature — Human Cleric", hobA3FiendHunterOracle, 1, 3)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Fatal("the creature is not exiled")
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hunter) })
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("after the Hunter left, %d Bears under its owner's control, want 1", n)
	}
	if n := namedOnBattlefield(g, "Their Bear", me.ID); n != 0 {
		t.Errorf("the Bear came back under %d of my controls, want its owner's", n)
	}
}

func TestHobA3FiendHunterMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Fiend Hunter", "Creature — Human Cleric", hobA3FiendHunterOracle, 1, 3)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, false)
	if !g.Battlefield.Contains(bear) {
		t.Error("a declined exile took the creature anyway")
	}
}

// The two abilities are separate: if the Hunter leaves before the entry
// trigger resolves, the leave trigger has nothing to return and the creature
// is then exiled for good.
func TestHobA3FiendHunterLeavingFirstExilesTheCreatureForGood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hunter := castHoldCreature(t, g, "Fiend Hunter", "Creature — Human Cleric", hobA3FiendHunterOracle, 1, 3)
	pickCard(t, g, me.ID, bear)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hunter) })
	g.RunStateChecksForTest()
	assistCardSettle(t, g, me.ID, nil, true)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Errorf("the creature should stay exiled: its return trigger resolved before the exile did (bf=%v exile=%v named=%d)",
			g.Battlefield.Contains(bear), g.Exile.Contains(bear), namedOnBattlefield(g, "Their Bear", opp.ID))
	}
}

func TestHobA3FiendHunterCannotTargetItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hunter := castHoldCreature(t, g, "Fiend Hunter", "Creature — Human Cleric", hobA3FiendHunterOracle, 1, 3)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no pick_target prompt")
	}
	for _, id := range p.PickTargetCards {
		if id == hunter {
			t.Error("the Hunter offered itself as a target")
		}
	}
	if !hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, mine) {
		t.Errorf("both other creatures are legal, got %v", p.PickTargetCards)
	}
}

// --- Landroval, Horizon Witness -------------------------------------------------

func hobA3LandrovalTable(t *testing.T) (g *game.Game, me, opp *game.Player, landroval uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	landroval = pushCatalogPermanent(g, me.ID, "Landroval, Horizon Witness", "Legendary Creature — Bird Noble", hobA3LandrovalOracle, false)
	return
}

func TestHobA3LandrovalGivesAnAttackerFlyingWhenTwoCreaturesAttackAPlayer(t *testing.T) {
	g, me, opp, _ := hobA3LandrovalTable(t)
	a := pushSizedCreature(g, me.ID, "Ground A", 2, 2)
	b := pushSizedCreature(g, me.ID, "Ground B", 2, 2)
	flyer := pushSizedCreature(g, me.ID, "Flyer", 2, 2, "flying")
	declareAttack(t, g, opp.ID, a, b, flyer)

	rfPermAExpectLegal(t, g, me.ID, hobA3LandrovalSpec(t).Triggered[0].Targets,
		map[uuid.UUID]bool{a: true, b: true, flyer: false})
	pickCard(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, a, "flying") {
		t.Error("the chosen attacker should have flying")
	}
	if hasEffectiveKeyword(t, g, b, "flying") {
		t.Error("only the chosen attacker gains flying")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("three attackers at one player are one trigger, not one each")
	}
}

func TestHobA3LandrovalDoesNothingForASingleAttacker(t *testing.T) {
	g, me, opp, _ := hobA3LandrovalTable(t)
	a := pushSizedCreature(g, me.ID, "Ground A", 2, 2)
	declareAttack(t, g, opp.ID, a)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("one attacker asked for a target")
	}
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, a, "flying") {
		t.Error("a lone attacker gained flying")
	}
}

// Two creatures attacking two different players is one attacker each.
func TestHobA3LandrovalNeedsTwoCreaturesAttackingTheSamePlayer(t *testing.T) {
	g, me, _, _ := hobA3LandrovalTable(t)
	a := pushSizedCreature(g, me.ID, "Ground A", 2, 2)
	b := pushSizedCreature(g, me.ID, "Ground B", 2, 2)
	first := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	second := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, first.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(b, second.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("one attacker per player asked for a target")
	}
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, a, "flying") || hasEffectiveKeyword(t, g, b, "flying") {
		t.Error("an attacker gained flying")
	}
}

// Landroval counts among the two, and her own flying keeps her off the list.
func TestHobA3LandrovalCountsHerselfAmongTheAttackers(t *testing.T) {
	g, me, opp, landroval := hobA3LandrovalTable(t)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, landroval).Keywords = []string{"flying"} })
	a := pushSizedCreature(g, me.ID, "Ground A", 2, 2)
	declareAttack(t, g, opp.ID, landroval, a)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: a}, true)
	if !hasEffectiveKeyword(t, g, a, "flying") {
		t.Error("the other attacker should have flying")
	}
}

func hobA3LandrovalSpec(t *testing.T) Spec {
	t.Helper()
	spec, ok := Lookup(hobA3LandrovalOracle)
	if !ok {
		t.Fatal("Landroval is not registered")
	}
	return spec
}

// --- Eagle of the Great Shelf -----------------------------------------------------

func TestHobA3EagleGetsPlusOnePlusOneForEachOtherCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	eagle := pushCatalogPermanent(g, me.ID, "Eagle of the Great Shelf", "Creature — Bird Soldier", hobA3EagleOracle, false)
	b12Creature(g, me.ID, "Friend A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Friend B", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Friend C", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	base := gaLayered(t, g, eagle)

	declareAttack(t, g, opp.ID, eagle)
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, eagle)
	if c.CurrentPower() != base.CurrentPower()+3 || c.CurrentToughness() != base.CurrentToughness()+3 {
		t.Errorf("got %d/%d, want %d/%d: +1/+1 for each of my three other creatures",
			c.CurrentPower(), c.CurrentToughness(), base.CurrentPower()+3, base.CurrentToughness()+3)
	}
}

func TestHobA3EagleAloneGetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	eagle := pushCatalogPermanent(g, me.ID, "Eagle of the Great Shelf", "Creature — Bird Soldier", hobA3EagleOracle, false)
	base := gaLayered(t, g, eagle)
	declareAttack(t, g, opp.ID, eagle)
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, eagle)
	if c.CurrentPower() != base.CurrentPower() || c.CurrentToughness() != base.CurrentToughness() {
		t.Errorf("got %d/%d, want the unchanged %d/%d", c.CurrentPower(), c.CurrentToughness(), base.CurrentPower(), base.CurrentToughness())
	}
}

func TestHobA3EagleFliesAndDoesNotPumpWhenItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castHoldCreature(t, g, "Eagle of the Great Shelf", "Creature — Bird Soldier", hobA3EagleOracle, 2, 5)
	if !effectiveAbilitiesContain(t, g, id, "flying") {
		t.Error("the Eagle should have flying")
	}
	b12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	c := gaLayered(t, g, id)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 5 {
		t.Errorf("got %d/%d, want 2/5: the pump is an attack trigger, not a static", c.CurrentPower(), c.CurrentToughness())
	}
}
