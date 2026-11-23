package testsg

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// --- Giant's Boulder ---------------------------------------------------

func TestHobA3GiantsBoulderScriesTwoWhenItEnters(t *testing.T) {
	hobA1ScryOnEntry(t, "Giant's Boulder", "Artifact", hobA3GiantsBoulderOracle, 2)
}

func TestHobA3GiantsBoulderPaysOneToTapForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	boulder := effectstest.PushCatalogPermanent(g, me.ID, "Giant's Boulder", "Artifact", hobA3GiantsBoulderOracle, false)

	if err := g.ActivateManaAbility(me.ID, boulder, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Fatal("the fixing tap needs {1} to pay; it ran with an empty pool")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, boulder, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want [R]: the {1} was spent, not produced", got)
	}
	if !effectstest.B16Tapped(t, g, boulder) {
		t.Error("the ability has a tap cost")
	}
}

func TestHobA3GiantsBoulderSacrificesForSevenToDestroyAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boulder := effectstest.PushCatalogPermanent(g, me.ID, "Giant's Boulder", "Artifact", hobA3GiantsBoulderOracle, false)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Ogre", 4, 4)
	effectstest.AdvanceToMain(t, g)

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	effectstest.B16Activate(t, g, me.ID, boulder, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Command Tower", "Land"),
		effectstest.SearchTestLand("Sol Ring", "Artifact"),
	)
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Old Thrush", "Creature — Bird", hobA3OldThrushOracle, 1, 2)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, life
}

func TestHobA3OldThrushGainsTwoAndPutsABasicLandOnTop(t *testing.T) {
	g, me, life := hobA3ThrushTable(t)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d: the two life is gained whatever the search does", me.Life, life+2)
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger did not offer a search")
	}
	for _, n := range []string{"Command Tower", "Sol Ring"} {
		if effectstest.SearchOptionNamed(g, c, n) != uuid.Nil {
			t.Errorf("%s is not a basic land card and must not be offered", n)
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Island")
	effectstest.PassPriorityAroundTable(t, g)
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
	before := effectstest.LibraryIDs(me)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger did not offer a search")
	}
	if err := effectstest.AnswerSearchApply(g, c, me.ID, me.ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SameIDList(effectstest.LibraryIDs(me), before) {
		t.Error("the library changed after a declined search")
	}
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
}

// --- Dwarven Warriors ------------------------------------------------------

func TestHobA3DwarvenWarriorsMakesAPowerTwoCreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	warriors := effectstest.PushCatalogPermanent(g, me.ID, "Dwarven Warriors", "Creature — Dwarf Warrior", hobA3DwarvenWarriorsOracle, false)
	small := effectstest.PushRestrictionBear(g, me, "Small Beast") // 2/2
	other := effectstest.PushRestrictionBear(g, me, "Other Beast")
	effectstest.AdvanceToMain(t, g)

	effectstest.B16Activate(t, g, me.ID, warriors, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(small)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, small, game.CantBeBlocked)
	effectstest.AssertRestrictions(t, g, other, 0)
	if !effectstest.B16Tapped(t, g, warriors) {
		t.Error("the ability has a tap cost")
	}
}

func TestHobA3DwarvenWarriorsRefusesAPowerThreeCreatureButMayTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	warriors := effectstest.PushCatalogPermanent(g, me.ID, "Dwarven Warriors", "Creature — Dwarf Warrior", hobA3DwarvenWarriorsOracle, false)
	big := effectstest.B12Creature(g, me.ID, "Big Beast", "Creature — Beast", 3, 3)
	effectstest.AdvanceToMain(t, g)

	err := g.ActivateCatalogAbility(me.ID, warriors, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(big)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a power-3 creature: err = %v, want ErrIllegalTarget", err)
	}
	effectstest.B16Activate(t, g, me.ID, warriors, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(warriors)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, warriors, game.CantBeBlocked)
}

// --- Fire of Orthanc -------------------------------------------------------

func TestHobA3FireOfOrthancDestroysAnArtifactOrALandButNotACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Sol Ring", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)

	if err := effectstest.CastCatalogSpellErr(t, g, "Fire of Orthanc", "Sorcery", hobA3FireOfOrthancOracle, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Fire of Orthanc targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Fire of Orthanc", "Sorcery", hobA3FireOfOrthancOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact should be destroyed")
	}
	effectstest.CastCatalogSpell(t, g, "Fire of Orthanc", "Sorcery", hobA3FireOfOrthancOracle, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !opp.Graveyard.Contains(land) {
		t.Error("the land should be destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was not a target and should be untouched")
	}
}

func TestHobA3FireOfOrthancStopsCreaturesWithoutFlyingFromBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := effectstest.PushSizedCreature(g, me.ID, "My Attacker", 3, 3)
	ground := effectstest.PushSizedCreature(g, opp.ID, "Their Bear", 2, 2)
	flyer := effectstest.PushSizedCreature(g, opp.ID, "Their Flyer", 1, 1, "flying")
	rock := effectstest.B12Permanent(g, opp.ID, "Sol Ring", "Artifact")

	effectstest.CastCatalogSpell(t, g, "Fire of Orthanc", "Sorcery", hobA3FireOfOrthancOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.PushSizedCreature(g, opp.ID, "Late Bear", 2, 2)

	effectstest.AssertRestrictions(t, g, ground, game.CantBlock)
	effectstest.AssertRestrictions(t, g, late, game.CantBlock)
	effectstest.AssertRestrictions(t, g, flyer, 0)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(ground, attacker); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("a creature without flying blocked: err = %v, want ErrIllegalBlock", err)
	}
	if err := g.DeclareBlocker(flyer, attacker); err != nil {
		t.Errorf("a creature with flying may still block: %v", err)
	}
}

// --- Goblin Fireleaper -----------------------------------------------------

func TestHobA3GoblinFireleaperPumpsForOneRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Fireleaper", "Creature — Goblin Warrior", hobA3GoblinFireleaperOID, false)
	effectstest.AdvanceToMain(t, g)
	if got := effects.B31CurrentPowerOf(g, id); got != 1 {
		t.Fatalf("base power = %d, want 1", got)
	}
	for i := 0; i < 2; i++ {
		effectstest.B06AddMana(me, "R", "R")
		effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effects.B31CurrentPowerOf(g, id); got != 3 {
		t.Errorf("power after two pumps = %d, want 3", got)
	}
}

func TestHobA3GoblinFireleaperDiesForItsPumpedPowerToAnOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Fireleaper", "Creature — Goblin Warrior", hobA3GoblinFireleaperOID, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Ogre", 5, 5)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Ogre", 5, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effects.B31CurrentPowerOf(g, id); got != 2 {
		t.Fatalf("pumped power = %d, want 2", got)
	}

	effectstest.EerieDestroy(t, g, id)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)
	if got := effectstest.DamageMarkedOn(g, victim); got != 2 {
		t.Errorf("damage on the target = %d, want the 2 power it died with", got)
	}
	if got := effectstest.DamageMarkedOn(g, mine); got != 0 {
		t.Errorf("damage on my own creature = %d, want 0", got)
	}
}

func TestHobA3GoblinFireleaperCanOnlyTargetACreatureAnOpponentControls(t *testing.T) {
	spec, _ := effects.Lookup(hobA3GoblinFireleaperOID)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	theirLand := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.RfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{mine: false, theirs: true, theirLand: false})
}

// --- Westfold Rider --------------------------------------------------------

func TestHobA3WestfoldRiderSacrificesToDestroyAnEnchantmentOrArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := effectstest.PushCatalogPermanent(g, me.ID, "Westfold Rider", "Creature — Human Knight", hobA3WestfoldRiderOracle, false)
	aura := effectstest.B12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	err := g.ActivateCatalogAbility(me.ID, rider, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a creature: err = %v, want ErrIllegalTarget", err)
	}
	effectstest.B16Activate(t, g, me.ID, rider, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(aura)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment survived")
	}
	if g.Battlefield.Contains(rider) || !me.Graveyard.Contains(rider) {
		t.Error("the Rider is the cost and should be in the graveyard")
	}
}

func TestHobA3WestfoldRiderDestroysAnArtifactToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := effectstest.PushCatalogPermanent(g, me.ID, "Westfold Rider", "Creature — Human Knight", hobA3WestfoldRiderOracle, false)
	rock := effectstest.B12Permanent(g, opp.ID, "Sol Ring", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, rider, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestHobA3WestfoldRiderActivatesOnlyAsASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rider := effectstest.PushCatalogPermanent(g, me.ID, "Westfold Rider", "Creature — Human Knight", hobA3WestfoldRiderOracle, false)
	aura := effectstest.B12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(aura)}); err == nil {
		t.Fatal("the ability ran outside a main phase")
	}
	if !g.Battlefield.Contains(rider) || !g.Battlefield.Contains(aura) {
		t.Error("a refused activation must not cost the Rider or touch the target")
	}
}

// --- Errand-Rider of Gondor --------------------------------------------------

func TestHobA3ErrandRiderDrawsThenTucksACardWithoutALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	keep := effectstest.HandCardForTest(me, "Keep", "Instant", "")
	hand := me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Errand-Rider of Gondor", "Creature — Human Soldier", hobA3ErrandRiderOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand = %d, want %d after the draw", got, hand+1)
	}
	ask := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ask == nil {
		t.Fatal("no prompt to put a card from your hand on the bottom")
	}
	if len(ask.ChooseCards) != me.Hand.Size() {
		t.Errorf("offered %d cards, want the whole hand (%d)", len(ask.ChooseCards), me.Hand.Size())
	}
	effectstest.AnswerChooseCards(t, g, me.ID, keep)
	if me.Hand.Contains(keep) {
		t.Error("the chosen card is still in hand")
	}
	if n := len(me.Library.Cards); n == 0 || me.Library.Cards[0].InstanceID != keep {
		t.Error("the chosen card is not on the bottom of the library")
	}
}

func TestHobA3ErrandRiderOnlyDrawsWhileYouControlALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.HandCardForTest(me, "Keep", "Instant", "")
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Aragorn", TypeLine: "Legendary Creature — Human Noble",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Errand-Rider of Gondor", "Creature — Human Soldier", hobA3ErrandRiderOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked to put a card on the bottom while a legendary creature is controlled")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d: the draw only", got, hand+1)
	}
}

// With nothing else in hand, the card just drawn is the only one on offer.
func TestHobA3ErrandRiderBottomsTheDrawnCardWhenItIsAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	drawn := effectstest.SeedLibrary(me, "Drawn")[0]
	effectstest.CastHoldCreature(t, g, "Errand-Rider of Gondor", "Creature — Human Soldier", hobA3ErrandRiderOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ask == nil || len(ask.ChooseCards) != 1 || ask.ChooseCards[0] != drawn {
		t.Fatalf("want a prompt offering only the drawn card, got %+v", ask)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, drawn)
	if me.Hand.Size() != 0 || me.Library.Cards[0].InstanceID != drawn {
		t.Error("the drawn card should be back on the bottom of the library")
	}
}

// --- Fiend Hunter -------------------------------------------------------------

func TestHobA3FiendHunterExilesACreatureAndReturnsItUnderItsOwnersControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// Stolen: mine to control, theirs to own.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Controller = me.ID })
	hunter := effectstest.CastHoldCreature(t, g, "Fiend Hunter", "Creature — Human Cleric", hobA3FiendHunterOracle, 1, 3)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Fatal("the creature is not exiled")
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hunter) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("after the Hunter left, %d Bears under its owner's control, want 1", n)
	}
	if n := effectstest.NamedOnBattlefield(g, "Their Bear", me.ID); n != 0 {
		t.Errorf("the Bear came back under %d of my controls, want its owner's", n)
	}
}

func TestHobA3FiendHunterMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Fiend Hunter", "Creature — Human Cleric", hobA3FiendHunterOracle, 1, 3)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, false)
	if !g.Battlefield.Contains(bear) {
		t.Error("a declined exile took the creature anyway")
	}
}

// The two abilities are separate: if the Hunter leaves before the entry
// trigger resolves, the leave trigger has nothing to return and the creature
// is then exiled for good.
func TestHobA3FiendHunterLeavingFirstExilesTheCreatureForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hunter := effectstest.CastHoldCreature(t, g, "Fiend Hunter", "Creature — Human Cleric", hobA3FiendHunterOracle, 1, 3)
	effectstest.PickCard(t, g, me.ID, bear)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hunter) })
	g.RunStateChecksForTest()
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Errorf("the creature should stay exiled: its return trigger resolved before the exile did (bf=%v exile=%v named=%d)",
			g.Battlefield.Contains(bear), g.Exile.Contains(bear), effectstest.NamedOnBattlefield(g, "Their Bear", opp.ID))
	}
}

func TestHobA3FiendHunterCannotTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hunter := effectstest.CastHoldCreature(t, g, "Fiend Hunter", "Creature — Human Cleric", hobA3FiendHunterOracle, 1, 3)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no pick_target prompt")
	}
	for _, id := range p.PickTargetCards {
		if id == hunter {
			t.Error("the Hunter offered itself as a target")
		}
	}
	if !effectstest.HasID(p.PickTargetCards, theirs) || !effectstest.HasID(p.PickTargetCards, mine) {
		t.Errorf("both other creatures are legal, got %v", p.PickTargetCards)
	}
}

// --- Landroval, Horizon Witness -------------------------------------------------

func hobA3LandrovalTable(t *testing.T) (g *game.Game, me, opp *game.Player, landroval uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	landroval = effectstest.PushCatalogPermanent(g, me.ID, "Landroval, Horizon Witness", "Legendary Creature — Bird Noble", hobA3LandrovalOracle, false)
	return
}

func TestHobA3LandrovalGivesAnAttackerFlyingWhenTwoCreaturesAttackAPlayer(t *testing.T) {
	g, me, opp, _ := hobA3LandrovalTable(t)
	a := effectstest.PushSizedCreature(g, me.ID, "Ground A", 2, 2)
	b := effectstest.PushSizedCreature(g, me.ID, "Ground B", 2, 2)
	flyer := effectstest.PushSizedCreature(g, me.ID, "Flyer", 2, 2, "flying")
	effectstest.DeclareAttack(t, g, opp.ID, a, b, flyer)

	effectstest.RfPermAExpectLegal(t, g, me.ID, hobA3LandrovalSpec(t).Triggered[0].Targets,
		map[uuid.UUID]bool{a: true, b: true, flyer: false})
	effectstest.PickCard(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, a, "flying") {
		t.Error("the chosen attacker should have flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, b, "flying") {
		t.Error("only the chosen attacker gains flying")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("three attackers at one player are one trigger, not one each")
	}
}

func TestHobA3LandrovalDoesNothingForASingleAttacker(t *testing.T) {
	g, me, opp, _ := hobA3LandrovalTable(t)
	a := effectstest.PushSizedCreature(g, me.ID, "Ground A", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("one attacker asked for a target")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, a, "flying") {
		t.Error("a lone attacker gained flying")
	}
}

// Two creatures attacking two different players is one attacker each.
func TestHobA3LandrovalNeedsTwoCreaturesAttackingTheSamePlayer(t *testing.T) {
	g, me, _, _ := hobA3LandrovalTable(t)
	a := effectstest.PushSizedCreature(g, me.ID, "Ground A", 2, 2)
	b := effectstest.PushSizedCreature(g, me.ID, "Ground B", 2, 2)
	first := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	second := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, first.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(b, second.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("one attacker per player asked for a target")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, a, "flying") || effectstest.HasEffectiveKeyword(t, g, b, "flying") {
		t.Error("an attacker gained flying")
	}
}

// Landroval counts among the two, and her own flying keeps her off the list.
func TestHobA3LandrovalCountsHerselfAmongTheAttackers(t *testing.T) {
	g, me, opp, landroval := hobA3LandrovalTable(t)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, landroval).Keywords = []string{"flying"} })
	a := effectstest.PushSizedCreature(g, me.ID, "Ground A", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, landroval, a)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: a}, true)
	if !effectstest.HasEffectiveKeyword(t, g, a, "flying") {
		t.Error("the other attacker should have flying")
	}
}

// #4864: Landroval counts only creatures whose attack target is the
// player. One at the player and one at their planeswalker, or two at a
// battle they protect, are not two creatures attacking a player.
func TestHobA3LandrovalCountsOnlyAttackersAtThePlayer(t *testing.T) {
	for _, tc := range hornAttackCases {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp, _ := hobA3LandrovalTable(t)
			other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
			walker, battle := hornAttackTargets(g, opp.ID, other.ID)
			at := tc.at(opp.ID, walker, battle)
			ids := [2]uuid.UUID{effectstest.PushSizedCreature(g, me.ID, "Ground A", 2, 2), effectstest.PushSizedCreature(g, me.ID, "Ground B", 2, 2)}
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			for i, id := range ids {
				if err := g.DeclareAttacker(id, at[i]); err != nil {
					t.Fatalf("DeclareAttacker %d: %v", i, err)
				}
			}
			effectstest.LockInAttacks(t, g)
			if got := effectstest.LatestPickTarget(g, me.ID) != nil; got != tc.want {
				t.Errorf("Landroval triggered = %v, want %v", got, tc.want)
			}
		})
	}
}

func hobA3LandrovalSpec(t *testing.T) effects.Spec {
	t.Helper()
	spec, ok := effects.Lookup(hobA3LandrovalOracle)
	if !ok {
		t.Fatal("Landroval is not registered")
	}
	return spec
}

// --- Eagle of the Great Shelf -----------------------------------------------------

func TestHobA3EagleGetsPlusOnePlusOneForEachOtherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	eagle := effectstest.PushCatalogPermanent(g, me.ID, "Eagle of the Great Shelf", "Creature — Bird Soldier", hobA3EagleOracle, false)
	effectstest.B12Creature(g, me.ID, "Friend A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Friend B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Friend C", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	base := effectstest.GaLayered(t, g, eagle)

	effectstest.DeclareAttack(t, g, opp.ID, eagle)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, eagle)
	if c.CurrentPower() != base.CurrentPower()+3 || c.CurrentToughness() != base.CurrentToughness()+3 {
		t.Errorf("got %d/%d, want %d/%d: +1/+1 for each of my three other creatures",
			c.CurrentPower(), c.CurrentToughness(), base.CurrentPower()+3, base.CurrentToughness()+3)
	}
}

func TestHobA3EagleAloneGetsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	eagle := effectstest.PushCatalogPermanent(g, me.ID, "Eagle of the Great Shelf", "Creature — Bird Soldier", hobA3EagleOracle, false)
	base := effectstest.GaLayered(t, g, eagle)
	effectstest.DeclareAttack(t, g, opp.ID, eagle)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, eagle)
	if c.CurrentPower() != base.CurrentPower() || c.CurrentToughness() != base.CurrentToughness() {
		t.Errorf("got %d/%d, want the unchanged %d/%d", c.CurrentPower(), c.CurrentToughness(), base.CurrentPower(), base.CurrentToughness())
	}
}

func TestHobA3EagleFliesAndDoesNotPumpWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CastHoldCreature(t, g, "Eagle of the Great Shelf", "Creature — Bird Soldier", hobA3EagleOracle, 2, 5)
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "flying") {
		t.Error("the Eagle should have flying")
	}
	effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	c := effectstest.GaLayered(t, g, id)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 5 {
		t.Errorf("got %d/%d, want 2/5: the pump is an attack trigger, not a static", c.CurrentPower(), c.CurrentToughness())
	}
}
