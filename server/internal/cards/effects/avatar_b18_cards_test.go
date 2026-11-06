package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b18_cards_test.go — Avatar: The Last Airbender, batches 18 and
// 19 (#4459): every card cast, triggered or activated through the
// engine.

const (
	avB18LionTurtleOracle   = "fff728a1-56f5-4e97-91e2-19aba5772909"
	avB18LongFengOracle     = "9bf9fe83-b0df-4ef8-ac54-c0d223d4e410"
	avB18HawkOracle         = "4029e884-8154-42b3-ae29-3cb98374a294"
	avB18PlatypusOracle     = "933e4e8a-01ea-4f58-8f5e-4f6c43f4e105"
	avB18PoxbearersOracle   = "fefef024-3a05-4221-bb74-419ee08716e5"
	avB18ZeiOracle          = "59a10335-72df-44a5-a514-5bac392cd282"
	avB18ScavengersOracle   = "0feacc7c-b93e-4322-bc8b-b3c19006128e"
	avB18SokkaOracle        = "427f8ef9-12b3-4fa0-b303-28153566b6dc"
	avB18SukiOracle         = "9223488e-5206-4d43-90ef-387189dcee16"
	avB18SunWarriorsOracle  = "513a2fca-fea4-434e-9b25-b0c0d7f322b2"
	avB18TollsOracle        = "5d1d9914-3ef2-45c5-b1cb-cb85d2f7394c"
	avB18IrohOracle         = "8a98ddac-7341-4118-b188-51d8d03afd1d"
	avB18WardenOracle       = "43b339ec-4d7d-41ac-a5ab-96c6a6de00e2"
	avB18MusiciansOracle    = "7166bf81-05c2-457d-a0bb-a175175b8e79"
	avB18WhiteLotusOracle   = "c55a4b25-ccc4-49de-9f98-a829c547a64d"
	avB18ZhaoOracle         = "d6b0198a-c1f2-4142-b5e2-8205adf14d0a"
	avB18BarrelsOracle      = "c6cd520c-106b-4ee5-bc02-6c5fbdbc78dd"
	avB18WarshipOracle      = "987aab63-a889-400f-a3bd-feaef978b407"
	avB18BattleFanOracle    = "849e54f9-7662-4186-859c-e4f52afbccac"
	avB18MeteorSwordOracle  = "181965fb-e685-4b1f-9bd6-3a33bed933b2"
	avB18AllyTokenName      = "Ally"
	avB18LionTurtleTypeLine = "Legendary Creature — Elder Cat Turtle"
)

func TestAvatarB18EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB18LionTurtleOracle:  "The Lion-Turtle",
		avB18LongFengOracle:    "Long Feng, Grand Secretariat",
		avB18HawkOracle:        "Messenger Hawk",
		avB18PlatypusOracle:    "Platypus-Bear",
		avB18PoxbearersOracle:  "Pretending Poxbearers",
		avB18ZeiOracle:         "Professor Zei, Anthropologist",
		avB18ScavengersOracle:  "Sandbender Scavengers",
		avB18SokkaOracle:       "Sokka, Lateral Strategist",
		avB18SukiOracle:        "Suki, Kyoshi Warrior",
		avB18SunWarriorsOracle: "Sun Warriors",
		avB18TollsOracle:       "Tolls of War",
		avB18IrohOracle:        "Uncle Iroh",
		avB18WardenOracle:      "Vindictive Warden",
		avB18MusiciansOracle:   "Wandering Musicians",
		avB18WhiteLotusOracle:  "White Lotus Reinforcements",
		avB18ZhaoOracle:        "Zhao, Ruthless Admiral",
		avB18BarrelsOracle:     "Barrels of Blasting Jelly",
		avB18WarshipOracle:     "Fire Nation Warship",
		avB18BattleFanOracle:   "Kyoshi Battle Fan",
		avB18MeteorSwordOracle: "Meteor Sword",
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

// --- fixtures ------------------------------------------------------------------

// avB18Destroy destroys a permanent and settles the state checks.
func avB18Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
}

// avB18MillALesson mills a Lesson sorcery card off the top of p's
// library, so it reaches the graveyard through the engine's zone move.
func avB18MillALesson(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	g.WithWriteLock(func() {
		seedLibraryTop(p, "A Lesson", "Sorcery — Lesson", "{1}{U}")
		if err := g.MillNForEffect(p.ID, 1); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})
}

// avB18Allies counts the Ally tokens controller controls.
func avB18Allies(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == avB18AllyTokenName && IsToken(c) && c.Controller == controller {
			n++
		}
	}
	return n
}

// --- The Lion-Turtle -------------------------------------------------------------

func TestAvatarB18LionTurtleGainsThreeLifeAndMakesAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := lifeOf(g, me.ID)
	turtle := castCatalogSpell(t, g, "The Lion-Turtle", avB18LionTurtleTypeLine, avB18LionTurtleOracle, nil)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+3 {
		t.Fatalf("life %d, want %d", got, life+3)
	}
	if !g.Battlefield.Contains(turtle) {
		t.Fatal("The Lion-Turtle did not resolve")
	}
	tapper := b12Push(g, me.ID, "The Lion-Turtle", avB18LionTurtleTypeLine, avB18LionTurtleOracle, 3, 6)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, tapper, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("tap for {U}: %v", err)
	}
	mkWantPool(t, me, map[string]int{"U": 1})
}

func TestAvatarB18LionTurtleCantAttackOrBlockUntilThreeLessons(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	turtle := b12Push(g, me.ID, "The Lion-Turtle", avB18LionTurtleTypeLine, avB18LionTurtleOracle, 3, 6)
	avB18MillALesson(t, g, me)
	avB18MillALesson(t, g, me)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(turtle, opp.ID); err == nil {
		t.Fatal("The Lion-Turtle attacked with two Lessons in the graveyard")
	}
	avB18MillALesson(t, g, me)
	if err := g.DeclareAttacker(turtle, opp.ID); err != nil {
		t.Fatalf("The Lion-Turtle could not attack with three Lessons: %v", err)
	}
	if c := cardByID(g, turtle); c.Tapped {
		t.Error("vigilance: The Lion-Turtle tapped to attack")
	}
}

// --- Long Feng, Grand Secretariat --------------------------------------------------

func TestAvatarB18LongFengCountersOnYourCreatureOrLandDying(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	feng := b12Push(g, me.ID, "Long Feng, Grand Secretariat", "Legendary Creature — Human Advisor", avB18LongFengOracle, 2, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	avB18Destroy(t, g, bear)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, feng)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, feng, game.CounterPlusOne); got != 1 {
		t.Fatalf("after a creature died: %d counters, want 1", got)
	}

	avB18Destroy(t, g, land)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, feng)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, feng, game.CounterPlusOne); got != 2 {
		t.Fatalf("after a land died: %d counters, want 2", got)
	}

	// Neither an artifact of yours nor another player's creature.
	avB18Destroy(t, g, relic)
	avB18Destroy(t, g, theirs)
	if n := triggersOnStackFrom(g, feng); n != 0 || latestPickTarget(g, me.ID) != nil {
		t.Fatalf("an artifact or an opponent's creature triggered Long Feng (%d on the stack)", n)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, feng, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
}

func TestAvatarB18LongFengDoesNotTriggerOnItsOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	feng := b12Push(g, me.ID, "Long Feng, Grand Secretariat", "Legendary Creature — Human Advisor", avB18LongFengOracle, 2, 3)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	avB18Destroy(t, g, feng)
	if latestPickTarget(g, me.ID) != nil || g.Stack.Size() != 0 {
		t.Fatal("Long Feng triggered on its own death")
	}
}

// --- Messenger Hawk ----------------------------------------------------------------

func TestAvatarB18MessengerHawkMakesAClue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	clues := len(avB5CluesB7(g, me.ID))
	castCatalogSpell(t, g, "Messenger Hawk", "Creature — Bird Scout", avB18HawkOracle, nil)
	passPriorityAroundTable(t, g)
	if got := len(avB5CluesB7(g, me.ID)); got != clues+1 {
		t.Fatalf("%d Clues, want %d", got, clues+1)
	}
}

func TestAvatarB18MessengerHawkGetsTwoPowerAfterTwoDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedLibrary(me, "Draw A", "Draw B", "Draw C")
	hawk := b12Push(g, me.ID, "Messenger Hawk", "Creature — Bird Scout", avB18HawkOracle, 1, 2)
	if !hasKeywordOnBattlefield(t, g, hawk, "flying") {
		t.Error("Messenger Hawk has no flying")
	}
	for g.TurnTallyFor(me.ID).CardsDrawn < 1 {
		avB9Draw(t, g, me.ID, 1)
	}
	if g.TurnTallyFor(me.ID).CardsDrawn == 1 {
		if got := effectivePower(t, g, hawk); got != 1 {
			t.Fatalf("power %d with one card drawn this turn, want 1", got)
		}
		avB9Draw(t, g, me.ID, 1)
	}
	if got := effectivePower(t, g, hawk); got != 3 {
		t.Errorf("power %d after two draws this turn, want 3", got)
	}
	if got := effectiveToughness(t, g, hawk); got != 2 {
		t.Errorf("toughness %d, want 2", got)
	}
}

func TestAvatarB18MessengerHawkIsAOneTwoBeforeTwoDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hawk := b12Push(g, me.ID, "Messenger Hawk", "Creature — Bird Scout", avB18HawkOracle, 1, 2)
	if g.TurnTallyFor(me.ID).CardsDrawn >= 2 {
		t.Skip("the fixture already drew two cards this turn")
	}
	if got := effectivePower(t, g, hawk); got != 1 {
		t.Errorf("power %d, want 1", got)
	}
}

// --- Platypus-Bear -----------------------------------------------------------------

func TestAvatarB18PlatypusBearMillsTwoOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	before := me.Graveyard.Size()
	castCatalogSpell(t, g, "Platypus-Bear", "Creature — Platypus Bear", avB18PlatypusOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - before; got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

func TestAvatarB18PlatypusBearAttacksOnlyWithALessonInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Push(g, me.ID, "Platypus-Bear", "Creature — Platypus Bear", avB18PlatypusOracle, 2, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Fatal("a defender attacked with no Lesson in the graveyard")
	}
	avB18MillALesson(t, g, me)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("Platypus-Bear could not attack with a Lesson in the graveyard: %v", err)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "defender") {
		t.Error("the waiver removed defender")
	}
}

// --- Pretending Poxbearers ---------------------------------------------------------

func TestAvatarB18PoxbearersMakeAnAllyWhenTheyDie(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pox := b12Push(g, me.ID, "Pretending Poxbearers", "Creature — Human Citizen Ally", avB18PoxbearersOracle, 2, 1)
	avB18Destroy(t, g, pox)
	passPriorityAroundTable(t, g)
	if got := avB18Allies(g, me.ID); got != 1 {
		t.Fatalf("%d Ally tokens, want 1", got)
	}
	ally := battlefieldIDsNamed(g, avB18AllyTokenName)[0]
	c := avB9Live(t, g, ally)
	if c.Power != 1 || c.Toughness != 1 || !c.IsCreature() || !c.HasSubtype("Ally") {
		t.Errorf("want a 1/1 Ally creature: %+v", c)
	}
}

// --- Professor Zei, Anthropologist -------------------------------------------------

func TestAvatarB18ZeiRummagesAndRebuysOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Next Card")
	zei := b12Push(g, me.ID, "Professor Zei, Anthropologist", "Legendary Creature — Human Advisor Ally", avB18ZeiOracle, 0, 3)
	hand := avB1Hand(me, game.Card{Name: "Spare", TypeLine: "Sorcery"})
	size := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, zei, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{hand}}); err != nil {
		t.Fatalf("rummage: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(hand) || me.Hand.Size() != size {
		t.Fatalf("discard one, draw one: hand %d (was %d), discarded in graveyard %v", me.Hand.Size(), size, me.Graveyard.Contains(hand))
	}

	// The rebuy: {1}, {T} — Zei is tapped from the rummage, so untap him.
	untapForTest(g, zei)
	bolt := pushGraveyardPermanent(me, "Old Bolt", "Instant", "{R}")
	bear := pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	fillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, zei, 1, game.ActivateAbilityParams{Strict: true, Targets: avB8Card(bear)}); err == nil {
		t.Fatal("a creature card was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, zei, 1, game.ActivateAbilityParams{Strict: true, Targets: avB8Card(bolt)}); err != nil {
		t.Fatalf("rebuy: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant did not return to hand")
	}
	if g.Battlefield.Contains(zei) {
		t.Error("Zei was not sacrificed")
	}
}

func TestAvatarB18ZeiCantRebuyOnAnotherPlayersTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]
	zei := b12Push(g, other.ID, "Professor Zei, Anthropologist", "Legendary Creature — Human Advisor Ally", avB18ZeiOracle, 0, 3)
	bolt := pushGraveyardPermanent(other, "Old Bolt", "Instant", "{R}")
	fillPoolColored(other, "C", 1)
	if err := g.ActivateCatalogAbility(other.ID, zei, 1, game.ActivateAbilityParams{Strict: true, Targets: avB8Card(bolt)}); err == nil {
		t.Fatal("Zei's rebuy was activated on another player's turn")
	}
}

// --- Sandbender Scavengers ---------------------------------------------------------

func TestAvatarB18ScavengersGrowOnSacrificesAndReanimateByPower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	scav := b12Push(g, me.ID, "Sandbender Scavengers", "Creature — Human Rogue", avB18ScavengersOracle, 1, 1)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	avB9Sacrifice(t, g, rock)
	if got := countersOn(g, scav, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after a sacrifice, want 1", got)
	}
	small := pushGraveyardPermanent(me, "Two Drop", "Creature — Bear", "{1}{G}")
	big := pushGraveyardPermanent(me, "Three Drop", "Creature — Bear", "{2}{G}")

	// Power 2 as it died: a mana value 3 card is not a legal target.
	avB18Destroy(t, g, scav)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	legal := latestPickTarget(g, me.ID).PickTargetCards
	if !hasID(legal, small) || hasID(legal, big) || hasID(legal, scav) {
		t.Fatalf("legal targets %v: want the mana value 2 card only", legal)
	}
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(scav) {
		t.Error("the Scavengers were not exiled")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the creature card did not return to the battlefield")
	}
}

func TestAvatarB18ScavengersIgnoreTheirOwnSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	scav := b12Push(g, me.ID, "Sandbender Scavengers", "Creature — Human Rogue", avB18ScavengersOracle, 1, 1)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(scav); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
	g.RunStateChecksForTest()
	// Only the dies trigger: no counter trigger for itself.
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(scav) {
		t.Error("declining the exile leaves the Scavengers in the graveyard")
	}
}

// --- Sokka, Lateral Strategist -----------------------------------------------------

func TestAvatarB18SokkaDrawsOnlyWhenAttackingWithAnother(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sokka := b12Push(g, me.ID, "Sokka, Lateral Strategist", "Legendary Creature — Human Warrior Ally", avB18SokkaOracle, 2, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "Drawn")
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, sokka, bear)
	if n := triggersOnStackFrom(g, sokka); n != 1 {
		t.Fatalf("%d Sokka triggers, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if cardByID(g, sokka).Tapped {
		t.Error("vigilance: Sokka tapped to attack")
	}
}

func TestAvatarB18SokkaAloneDoesNotDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sokka := b12Push(g, me.ID, "Sokka, Lateral Strategist", "Legendary Creature — Human Warrior Ally", avB18SokkaOracle, 2, 4)
	b12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, sokka)
	if n := triggersOnStackFrom(g, sokka); n != 0 {
		t.Fatalf("Sokka alone triggered %d times", n)
	}
}

func TestAvatarB18SokkaStayingHomeDoesNotDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sokka := b12Push(g, me.ID, "Sokka, Lateral Strategist", "Legendary Creature — Human Warrior Ally", avB18SokkaOracle, 2, 4)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a, b)
	if n := triggersOnStackFrom(g, sokka); n != 0 {
		t.Fatalf("Sokka triggered %d times without attacking", n)
	}
}

// --- Suki, Kyoshi Warrior ----------------------------------------------------------

func TestAvatarB18SukiCountsCreaturesAndMakesAnAttackingAlly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	suki := b12Push(g, me.ID, "Suki, Kyoshi Warrior", "Legendary Creature — Human Warrior Ally", avB18SukiOracle, 0, 4)
	b12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	if got := effectivePower(t, g, suki); got != 2 {
		t.Fatalf("power %d with two creatures, want 2", got)
	}
	declareAttack(t, g, opp.ID, suki)
	passAndSettleAttackSplit(t, g, me.ID)
	allies := battlefieldIDsNamed(g, avB18AllyTokenName)
	if len(allies) != 1 {
		t.Fatalf("%d Ally tokens, want 1", len(allies))
	}
	tok := cardByID(g, allies[0])
	if !tok.Tapped || tok.AttackingTarget != opp.ID {
		t.Errorf("want a tapped Ally attacking the defender: tapped %v, attacking %v", tok.Tapped, tok.AttackingTarget)
	}
	if got := effectivePower(t, g, suki); got != 3 {
		t.Errorf("power %d with three creatures, want 3", got)
	}
	if n := triggersOnStackFrom(g, suki); n != 0 {
		t.Error("the token was never declared, so Suki does not trigger again")
	}
}

// --- Sun Warriors ------------------------------------------------------------------

func TestAvatarB18SunWarriorsFirebendPerCreatureAndMakeAllies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sun := b12Push(g, me.ID, "Sun Warriors", "Creature — Human Warrior Ally", avB18SunWarriorsOracle, 3, 5)
	b12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Home Bear 2", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, sun)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 3})
	mkAdvanceTo(t, g, game.StepEndCombat)
	mkWantPool(t, me, map[string]int{"R": 3})
	mkAdvanceTo(t, g, game.StepPostcombatMain)
	mkWantPool(t, me, map[string]int{})

	fillPoolColored(me, "C", 4)
	if err := g.ActivateCatalogAbility(me.ID, sun, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the Ally ability was activated for four mana")
	}
	fillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, sun, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate {5}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := avB18Allies(g, me.ID); got != 1 {
		t.Errorf("%d Ally tokens, want 1", got)
	}
}

// --- Tolls of War ------------------------------------------------------------------

func TestAvatarB18TollsOfWarClueThenOneAllyPerTurnOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	clues := len(avB5CluesB7(g, me.ID))
	castCatalogSpell(t, g, "Tolls of War", "Enchantment", avB18TollsOracle, nil)
	passPriorityAroundTable(t, g)
	if got := len(avB5CluesB7(g, me.ID)); got != clues+1 {
		t.Fatalf("%d Clues, want %d", got, clues+1)
	}
	avB9Sacrifice(t, g, b12Permanent(g, me.ID, "Rock", "Artifact"))
	if got := avB18Allies(g, me.ID); got != 1 {
		t.Fatalf("%d Allies after the first sacrifice, want 1", got)
	}
	avB9Sacrifice(t, g, b12Permanent(g, me.ID, "Rock 2", "Artifact"))
	if got := avB18Allies(g, me.ID); got != 1 {
		t.Fatalf("%d Allies after a second sacrifice this turn, want 1 (once each turn)", got)
	}
}

func TestAvatarB18TollsOfWarIgnoresSacrificesOnOtherTurns(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]
	b12Push(g, other.ID, "Tolls of War", "Enchantment", avB18TollsOracle, 0, 0)
	avB9Sacrifice(t, g, b12Permanent(g, other.ID, "Rock", "Artifact"))
	if got := avB18Allies(g, other.ID); got != 0 {
		t.Errorf("%d Allies from a sacrifice on another player's turn, want 0", got)
	}
}

// --- Uncle Iroh --------------------------------------------------------------------

func TestAvatarB18UncleIrohDiscountsLessonSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Uncle Iroh", "Legendary Creature — Human Noble Ally", avB18IrohOracle, 4, 2)
	fillPoolColored(me, "R", 1)
	if _, err := castFromHandErr(g, me, "Plain Sorcery", "Sorcery", "{1}{R}", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a non-Lesson sorcery was discounted")
	}
	castFromHandForTest(t, g, me, "A Lesson", "Sorcery — Lesson", "{1}{R}", "", game.CastSpellParams{Strict: true})
}

func TestAvatarB18UncleIrohFirebendsOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	iroh := b12Push(g, me.ID, "Uncle Iroh", "Legendary Creature — Human Noble Ally", avB18IrohOracle, 4, 2)
	declareAttack(t, g, opp.ID, iroh)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 1})
}

// --- Vindictive Warden -------------------------------------------------------------

func TestAvatarB18VindictiveWardenFirebendsAndPingsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warden := b12Push(g, me.ID, "Vindictive Warden", "Creature — Human Soldier", avB18WardenOracle, 2, 3)
	if !hasKeywordOnBattlefield(t, g, warden, "menace") {
		t.Error("no menace")
	}
	declareAttack(t, g, opp.ID, warden)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 1})
	mkAdvanceTo(t, g, game.StepPostcombatMain)
	before := lifeOfOpponents(g)
	fillPoolColored(me, "C", 2)
	if err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ping was activated for two mana")
	}
	fillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate {3}: %v", err)
	}
	passPriorityAroundTable(t, g)
	after := lifeOfOpponents(g)
	for i := range before {
		if after[i] != before[i]-1 {
			t.Errorf("opponent %d life %d, want %d", i, after[i], before[i]-1)
		}
	}
}

// --- Wandering Musicians -----------------------------------------------------------

func TestAvatarB18WanderingMusiciansPumpTheTeamOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mus := b12Push(g, me.ID, "Wandering Musicians", "Creature — Human Bard Ally", avB18MusiciansOracle, 2, 5)
	home := b12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, mus)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, mus); got != 3 {
		t.Errorf("Musicians power %d, want 3", got)
	}
	if got := effectivePower(t, g, home); got != 3 {
		t.Errorf("your other creature power %d, want 3", got)
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature power %d, want 2", got)
	}
}

// --- White Lotus Reinforcements ----------------------------------------------------

func TestAvatarB18WhiteLotusPumpsOtherAlliesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lotus := b12Push(g, me.ID, "White Lotus Reinforcements", "Creature — Human Soldier Ally", avB18WhiteLotusOracle, 2, 3)
	ally := b12Creature(g, me.ID, "My Ally", "Creature — Human Ally", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirAlly := b12Creature(g, opp.ID, "Their Ally", "Creature — Human Ally", 1, 1)
	if p, tg := effectivePower(t, g, ally), effectiveToughness(t, g, ally); p != 2 || tg != 2 {
		t.Errorf("your Ally %d/%d, want 2/2", p, tg)
	}
	if p := effectivePower(t, g, lotus); p != 2 {
		t.Errorf("White Lotus pumps itself: power %d", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Ally got pumped: power %d", p)
	}
	if p := effectivePower(t, g, theirAlly); p != 1 {
		t.Errorf("an opponent's Ally got pumped: power %d", p)
	}
	if !hasKeywordOnBattlefield(t, g, lotus, "vigilance") {
		t.Error("no vigilance")
	}
}

// --- Zhao, Ruthless Admiral --------------------------------------------------------

func TestAvatarB18ZhaoFirebendsTwoAndPumpsOnSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zhao := b12Push(g, me.ID, "Zhao, Ruthless Admiral", "Legendary Creature — Human Soldier", avB18ZhaoOracle, 3, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, zhao)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 2})

	avB9Sacrifice(t, g, b12Permanent(g, me.ID, "Rock", "Artifact"))
	if got := effectivePower(t, g, zhao); got != 4 {
		t.Errorf("Zhao power %d after a sacrifice, want 4", got)
	}
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("bear power %d after a sacrifice, want 3", got)
	}
}

func TestAvatarB18ZhaoIgnoresItsOwnSacrificeAndOthersSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	zhao := b12Push(g, me.ID, "Zhao, Ruthless Admiral", "Legendary Creature — Human Soldier", avB18ZhaoOracle, 3, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	avB9Sacrifice(t, g, b12Permanent(g, opp.ID, "Their Rock", "Artifact"))
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("an opponent's sacrifice pumped: power %d", got)
	}
	avB9Sacrifice(t, g, zhao)
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("Zhao's own sacrifice pumped: power %d", got)
	}
}

// --- Barrels of Blasting Jelly -----------------------------------------------------

func TestAvatarB18BarrelsFilterOnceEachTurnAndBlast(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	barrels := b12Push(g, me.ID, "Barrels of Blasting Jelly", "Artifact", avB18BarrelsOracle, 0, 0)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "C", 2)
	if err := g.ActivateManaAbility(me.ID, barrels, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("{1}: add {G}: %v", err)
	}
	mkWantPool(t, me, map[string]int{"C": 1, "G": 1})
	if err := g.ActivateManaAbility(me.ID, barrels, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("a second activation this turn was allowed")
	}

	victim := b12Creature(g, opp.ID, "Their Big", "Creature — Giant", 5, 5)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "C", 5)
	if err := g.ActivateCatalogAbility(me.ID, barrels, 0, game.ActivateAbilityParams{Strict: true, Targets: avB8Card(victim)}); err != nil {
		t.Fatalf("blast: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(victim) {
		t.Error("5 damage did not kill the 5/5")
	}
	if g.Battlefield.Contains(barrels) {
		t.Error("the Barrels were not sacrificed")
	}
}

// --- Fire Nation Warship -----------------------------------------------------------

func TestAvatarB18WarshipLeavesAClueOnlyWhenItDiesCrewed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	clues := len(avB5CluesB7(g, me.ID))

	// Uncrewed, it is not a creature: destroying it is not dying.
	idle := b12Push(g, me.ID, "Fire Nation Warship", "Artifact — Vehicle", avB18WarshipOracle, 4, 4)
	avB18Destroy(t, g, idle)
	passPriorityAroundTable(t, g)
	if got := len(avB5CluesB7(g, me.ID)); got != clues {
		t.Fatalf("an uncrewed Warship made a Clue (%d, want %d)", got, clues)
	}

	ship := b12Push(g, me.ID, "Fire Nation Warship", "Artifact — Vehicle", avB18WarshipOracle, 4, 4)
	small := b12Creature(g, me.ID, "Deckhand", "Creature — Human", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Fatal("power 1 crewed Crew 2")
	}
	crew := b12Creature(g, me.ID, "Sailor", "Creature — Human", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := avB9Live(t, g, ship)
	if !c.IsCreature() || !game.HasKeyword(&c, "reach") {
		t.Fatalf("a crewed Warship is a creature with reach: %+v", c)
	}
	avB18Destroy(t, g, ship)
	passPriorityAroundTable(t, g)
	if got := len(avB5CluesB7(g, me.ID)); got != clues+1 {
		t.Errorf("%d Clues after the crewed Warship died, want %d", got, clues+1)
	}
}

// --- Kyoshi Battle Fan -------------------------------------------------------------

func TestAvatarB18BattleFanMakesAnAllyAndEquipsIt(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	fan := castCatalogSpell(t, g, "Kyoshi Battle Fan", "Artifact — Equipment", avB18BattleFanOracle, nil)
	passPriorityAroundTable(t, g)
	allies := battlefieldIDsNamed(g, avB18AllyTokenName)
	if len(allies) != 1 {
		t.Fatalf("%d Ally tokens, want 1", len(allies))
	}
	ally := allies[0]
	if att := cardByID(g, fan).AttachedTo; att.Kind != game.TargetCard || att.ID != ally {
		t.Fatalf("the Fan is attached to %+v, want the Ally", att)
	}
	if p, tg := effectivePower(t, g, ally), effectiveToughness(t, g, ally); p != 2 || tg != 1 {
		t.Errorf("equipped Ally %d/%d, want 2/1", p, tg)
	}

	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	fillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, fan, 0, game.ActivateAbilityParams{Strict: true, Targets: avB8Card(bear)}); err == nil {
		t.Fatal("equip {2} was paid with one mana")
	}
	fillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, fan, 0, game.ActivateAbilityParams{Strict: true, Targets: avB8Card(bear)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("equipped bear power %d, want 3", got)
	}
}

// --- Meteor Sword ------------------------------------------------------------------

func TestAvatarB18MeteorSwordDestroysAPermanentAndPumpsThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	relic := b12Permanent(g, opp.ID, "Their Relic", "Enchantment")
	sword := castCatalogSpell(t, g, "Meteor Sword", "Artifact — Equipment", avB18MeteorSwordOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Fatal("the enchantment was not destroyed")
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	fillPoolColored(me, "C", 3)
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{Strict: true, Targets: avB8Card(bear)}); err != nil {
		t.Fatalf("equip {3}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tg != 5 {
		t.Errorf("equipped bear %d/%d, want 5/5", p, tg)
	}
}
