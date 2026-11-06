package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b28_cards_test.go — Avatar: The Last Airbender Eternal, batch 26
// part 28 (#5075): every card cast, triggered or activated through the
// engine.

const (
	avB28InsurgentOracle  = "5863c67b-a47c-4466-8bd1-210ab26e0d3f"
	avB28CustomerOracle   = "3faab84f-7055-4554-b948-181928d8d4e2"
	avB28KoalaSheepOracle = "2e0226eb-20a3-488c-8f2b-f99a5b4d0cf4"
	avB28PipsqueakOracle  = "08a5445b-e9e6-44c3-9bc4-aff1b91c18a7"
	avB28SokkaOracle      = "90b4d288-e950-4182-af62-516549a8712c"
	avB28SwordOracle      = "19963c90-e7b2-4935-adcd-d41f0faf401a"
	avB28TaleOracle       = "728157a3-fcd4-4295-85ff-1162dbb64f8a"
	avB28RoughBuddyOracle = "fc3ee37a-f676-44af-80d0-774ff271ce11"
	avB28ToucanOracle     = "39d283e0-9d9b-4c04-aa6a-66a9883af9a2"
	avB28ChakraOracle     = "12c57ebe-35a7-4cb3-936b-99c148a3cfed"
)

func TestAvatarB28EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB28InsurgentOracle:  "Inspired Insurgent",
		avB28CustomerOracle:   "Kindly Customer",
		avB28KoalaSheepOracle: "Koala-Sheep",
		avB28PipsqueakOracle:  "Pipsqueak, Rebel Strongarm",
		avB28SokkaOracle:      "Sokka, Swordmaster",
		avB28SwordOracle:      "Sokka's Sword Training",
		avB28TaleOracle:       "Tale of Momo",
		avB28RoughBuddyOracle: "That's Rough Buddy",
		avB28ToucanOracle:     "Toucan-Puffin",
		avB28ChakraOracle:     "Chakra Meditation",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// avB28PickIfAsked answers an open pick_target prompt for `chooser` with
// `ids` and resolves what follows.
func avB28PickIfAsked(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	if latestPickTarget(g, chooser) != nil {
		b17PickCards(t, g, chooser, ids...)
	}
	passPriorityAroundTable(t, g)
}

func avB28PlusOneCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return c.Counters[game.CounterPlusOne]
}

// --- Inspired Insurgent -------------------------------------------------------

func TestAvatarB28InspiredInsurgentSacrificesToDestroyAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ins := b12Push(g, me.ID, "Inspired Insurgent", "Creature — Human Peasant Ally", avB28InsurgentOracle, 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ins, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(rock)}); err == nil {
		t.Fatal("activated without the {1}")
	}
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, ins, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err == nil {
		t.Fatal("a plain creature is not an artifact or enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, ins, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(rock)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(ins) {
		t.Fatal("the Insurgent was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

// --- Kindly Customer and Koala-Sheep ---------------------------------------------

func TestAvatarB28KindlyCustomerDrawsACardOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top Card")
	castCatalogSpell(t, g, "Kindly Customer", "Creature — Human Citizen", avB28CustomerOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestAvatarB28KoalaSheepGainsThreeLifeOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castCatalogSpell(t, g, "Koala-Sheep", "Creature — Bear Sheep", avB28KoalaSheepOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

// --- Pipsqueak, Rebel Strongarm -----------------------------------------------------

func TestAvatarB28PipsqueakCantAttackAloneWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pip := pushCatalogPermanent(g, me.ID, "Pipsqueak, Rebel Strongarm", "Legendary Creature — Human Rebel Ally", avB28PipsqueakOracle, false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(pip, opp.ID); err != nil {
		t.Fatalf("declare Pipsqueak: %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("Pipsqueak alone: %v, want ErrAttackLimit", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("declare Bear: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("with company: %v", err)
	}
}

func TestAvatarB28PipsqueakAttacksAloneWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pip := pushCatalogPermanent(g, me.ID, "Pipsqueak, Rebel Strongarm", "Legendary Creature — Human Rebel Ally", avB28PipsqueakOracle, false)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(pip, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(pip, opp.ID); err != nil {
		t.Fatalf("declare: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("Pipsqueak with a counter attacks alone: %v", err)
	}
	if c, _ := battlefieldCard(g, pip); c.AttackingTarget != opp.ID {
		t.Error("Pipsqueak is not attacking")
	}
}

// --- Sokka, Swordmaster ----------------------------------------------------------------

func TestAvatarB28SokkaDiscountsEquipmentByAllies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Sokka, Swordmaster", "Legendary Creature — Human Warrior Ally", avB28SokkaOracle, false)
	b12Creature(g, me.ID, "Ally Friend", "Creature — Human Ally", 1, 1)
	advanceToMain(t, g)
	blade := avB24Hand(g, me, game.Card{Name: "Blade", TypeLine: equipTypeLine, ManaCost: "{3}"})
	pricey := avB24Hand(g, me, game.Card{Name: "Pricey Blade", TypeLine: equipTypeLine, ManaCost: "{4}"})
	fillPool(me, 1)
	if err := g.CastSpell(me.ID, pricey, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a {4} Equipment cost only {1} with two Allies")
	}
	if err := g.CastSpell(me.ID, blade, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a {3} Equipment with two Allies for {1}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool %d after the discounted cast", me.ManaPool.Total())
	}
}

func TestAvatarB28SokkaAttachesAnEquipmentAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sokka := pushCatalogPermanent(g, me.ID, "Sokka, Swordmaster", "Legendary Creature — Human Warrior Ally", avB28SokkaOracle, false)
	axe := capGear(g, me.ID, uuid.Nil, "Axe", "{2}")
	theirs := capGear(g, opp.ID, uuid.Nil, "Their Axe", "{2}")
	advanceTo(t, g, game.StepBeginCombat)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("no target asked for")
	}
	b17PickCards(t, g, me.ID, axe)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, axe); host.ID != sokka {
		t.Errorf("the Axe is attached to %v, want Sokka", host)
	}
	if host := attachmentHostOf(t, g, theirs); host.Kind != "" {
		t.Errorf("the opponent's Axe moved to %v", host)
	}
}

// --- Sokka's Sword Training --------------------------------------------------------------

func TestAvatarB28SwordTrainingPumpsAndMakesAClue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Sokka's Sword Training", "Instant — Lesson", avB28SwordOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("bear %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if n := avB24Count(g, me.ID, "Clue"); n != 1 {
		t.Errorf("%d Clues, want 1", n)
	}
}

// --- Tale of Momo ---------------------------------------------------------------------

func TestAvatarB28TaleOfMomoCostsTwoLessAfterACreatureLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	advanceToMain(t, g)
	tale := avB24Hand(g, me, game.Card{Name: "Tale of Momo", TypeLine: "Sorcery", OracleID: avB28TaleOracle, ManaCost: "{2}{W}"})
	floatMana(t, g, me, "{W}")
	if err := g.CastSpell(me.ID, tale, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("discounted with nothing gone")
	}
	// A noncreature permanent leaving is not enough.
	avB9Sacrifice(t, g, rock)
	if err := g.CastSpell(me.ID, tale, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("discounted after only an artifact left")
	}
	avB9Sacrifice(t, g, fodder)
	if err := g.CastSpell(me.ID, tale, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Tale of Momo for {W} after a creature left: %v", err)
	}
}

func TestAvatarB28TaleOfMomoFindsAnAllyCreatureInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	lgKnowLibrary(me)
	ally := pushGraveyardCardWithTypeLineID(me, "Old Friend", "Creature — Human Ally")
	pushGraveyardCardWithTypeLineID(me, "Ally Lesson", "Sorcery — Lesson")
	pushGraveyardCardWithTypeLineID(me, "Plain Bear", "Creature — Bear")
	castCatalogSpell(t, g, "Tale of Momo", "Sorcery", avB28TaleOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	c := searchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ally {
		t.Fatalf("only the Ally creature card is a candidate, got %+v", c)
	}
	answerSearchByID(t, g, me.ID, ally)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ally) {
		t.Error("the Ally creature card is not in hand")
	}
}

// --- That's Rough Buddy -----------------------------------------------------------------

func TestAvatarB28RoughBuddyPutsOneCounterAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top Card")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "That's Rough Buddy", "Instant — Lesson", avB28RoughBuddyOracle, cardRefs(bear))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if n := avB28PlusOneCounters(t, g, bear); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestAvatarB28RoughBuddyPutsTwoCountersAfterACreatureLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top Card")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	avB9Sacrifice(t, g, fodder)
	castCatalogSpell(t, g, "That's Rough Buddy", "Instant — Lesson", avB28RoughBuddyOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if n := avB28PlusOneCounters(t, g, bear); n != 2 {
		t.Errorf("%d counters, want 2", n)
	}
}

func TestAvatarB28RoughBuddyIgnoresAnOpponentsCreatureLeaving(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Top Card")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	avB9Sacrifice(t, g, theirs)
	castCatalogSpell(t, g, "That's Rough Buddy", "Instant — Lesson", avB28RoughBuddyOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if n := avB28PlusOneCounters(t, g, bear); n != 1 {
		t.Errorf("%d counters, want 1: the creature that left was not yours", n)
	}
}

// --- Toucan-Puffin -------------------------------------------------------------------------

func TestAvatarB28ToucanPuffinPumpsACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Toucan-Puffin", "Creature — Bird", avB28ToucanOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target asked for")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(theirs)); err == nil {
		t.Fatal("an opponent's creature was accepted")
	}
	avB28PickIfAsked(t, g, me.ID, bear)
	c, _ := battlefieldCard(g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Errorf("bear %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Chakra Meditation ---------------------------------------------------------------------

func TestAvatarB28ChakraMeditationReturnsAnInstantOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bolt := pushGraveyardCardWithTypeLineID(me, "Old Bolt", "Instant")
	bear := pushGraveyardCardWithTypeLineID(me, "Old Bear", "Creature — Bear")
	castCatalogSpell(t, g, "Chakra Meditation", "Enchantment", avB28ChakraOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target asked for")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(bear)); err == nil {
		t.Fatal("a creature card was accepted")
	}
	avB28PickIfAsked(t, g, me.ID, bolt)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant is not back in hand")
	}
}

func avB28ChakraTable(t *testing.T, lessons int) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Drawn One", "Drawn Two")
	for i := 0; i < lessons; i++ {
		pushGraveyardCardWithTypeLineID(me, "Old Lesson", "Sorcery — Lesson")
	}
	pushCatalogPermanent(g, me.ID, "Chakra Meditation", "Enchantment", avB28ChakraOracle, false)
	advanceToMain(t, g)
	return g, me
}

func TestAvatarB28ChakraMeditationLootsOnAnInstant(t *testing.T) {
	g, me := avB28ChakraTable(t, 2)
	keep := avB24Hand(g, me, game.Card{Name: "Keeper", TypeLine: "Creature — Bear"})
	shock := avB24Hand(g, me, game.Card{Name: "Quick Spell", TypeLine: "Instant"})
	if err := g.CastSpell(me.ID, shock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, keep)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(keep) {
		t.Error("no discard with only two Lessons in the graveyard")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d (draw one, discard one)", me.Hand.Size(), hand)
	}
}

func TestAvatarB28ChakraMeditationSkipsTheDiscardWithThreeLessons(t *testing.T) {
	g, me := avB28ChakraTable(t, 3)
	avB24Hand(g, me, game.Card{Name: "Keeper", TypeLine: "Creature — Bear"})
	shock := avB24Hand(g, me, game.Card{Name: "Quick Spell", TypeLine: "Instant"})
	if err := g.CastSpell(me.ID, shock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDiscardFromHand {
			t.Fatal("a discard was asked for with three Lessons in the graveyard")
		}
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestAvatarB28ChakraMeditationIgnoresCreatureSpells(t *testing.T) {
	g, me := avB28ChakraTable(t, 0)
	bear := avB24Hand(g, me, game.Card{Name: "Hand Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: a creature spell drew", me.Hand.Size(), hand)
	}
}
