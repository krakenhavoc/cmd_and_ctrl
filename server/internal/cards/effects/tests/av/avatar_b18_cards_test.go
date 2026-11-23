package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// avB18MillALesson mills a Lesson sorcery card off the top of p's
// library, so it reaches the graveyard through the engine's zone move.
func avB18MillALesson(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	g.WithWriteLock(func() {
		effectstest.SeedLibraryTop(p, "A Lesson", "Sorcery — Lesson", "{1}{U}")
		if err := g.MillNForEffect(p.ID, 1); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})
}

// avB18Allies counts the Ally tokens controller controls.
func avB18Allies(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == avB18AllyTokenName && effects.IsToken(c) && c.Controller == controller {
			n++
		}
	}
	return n
}

// --- The Lion-Turtle -------------------------------------------------------------

func TestAvatarB18LionTurtleGainsThreeLifeAndMakesAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := effectstest.LifeOf(g, me.ID)
	turtle := effectstest.CastCatalogSpell(t, g, "The Lion-Turtle", avB18LionTurtleTypeLine, avB18LionTurtleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+3 {
		t.Fatalf("life %d, want %d", got, life+3)
	}
	if !g.Battlefield.Contains(turtle) {
		t.Fatal("The Lion-Turtle did not resolve")
	}
	tapper := effectstest.B12Push(g, me.ID, "The Lion-Turtle", avB18LionTurtleTypeLine, avB18LionTurtleOracle, 3, 6)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, tapper, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("tap for {U}: %v", err)
	}
	effectstest.MkWantPool(t, me, map[string]int{"U": 1})
}

func TestAvatarB18LionTurtleCantAttackOrBlockUntilThreeLessons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	turtle := effectstest.B12Push(g, me.ID, "The Lion-Turtle", avB18LionTurtleTypeLine, avB18LionTurtleOracle, 3, 6)
	avB18MillALesson(t, g, me)
	avB18MillALesson(t, g, me)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(turtle, opp.ID); err == nil {
		t.Fatal("The Lion-Turtle attacked with two Lessons in the graveyard")
	}
	avB18MillALesson(t, g, me)
	if err := g.DeclareAttacker(turtle, opp.ID); err != nil {
		t.Fatalf("The Lion-Turtle could not attack with three Lessons: %v", err)
	}
	if c := effectstest.CardByID(g, turtle); c.Tapped {
		t.Error("vigilance: The Lion-Turtle tapped to attack")
	}
}

// --- Long Feng, Grand Secretariat --------------------------------------------------

func TestAvatarB18LongFengCountersOnYourCreatureOrLandDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	feng := effectstest.B12Push(g, me.ID, "Long Feng, Grand Secretariat", "Legendary Creature — Human Advisor", avB18LongFengOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.AvB18Destroy(t, g, bear)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, feng)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, feng, game.CounterPlusOne); got != 1 {
		t.Fatalf("after a creature died: %d counters, want 1", got)
	}

	effectstest.AvB18Destroy(t, g, land)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, feng)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, feng, game.CounterPlusOne); got != 2 {
		t.Fatalf("after a land died: %d counters, want 2", got)
	}

	// Neither an artifact of yours nor another player's creature.
	effectstest.AvB18Destroy(t, g, relic)
	effectstest.AvB18Destroy(t, g, theirs)
	if n := effectstest.TriggersOnStackFrom(g, feng); n != 0 || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatalf("an artifact or an opponent's creature triggered Long Feng (%d on the stack)", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, feng, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
}

func TestAvatarB18LongFengDoesNotTriggerOnItsOwnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	feng := effectstest.B12Push(g, me.ID, "Long Feng, Grand Secretariat", "Legendary Creature — Human Advisor", avB18LongFengOracle, 2, 3)
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AvB18Destroy(t, g, feng)
	if effectstest.LatestPickTarget(g, me.ID) != nil || g.Stack.Size() != 0 {
		t.Fatal("Long Feng triggered on its own death")
	}
}

// --- Messenger Hawk ----------------------------------------------------------------

func TestAvatarB18MessengerHawkMakesAClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	clues := len(avB5CluesB7(g, me.ID))
	effectstest.CastCatalogSpell(t, g, "Messenger Hawk", "Creature — Bird Scout", avB18HawkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(avB5CluesB7(g, me.ID)); got != clues+1 {
		t.Fatalf("%d Clues, want %d", got, clues+1)
	}
}

func TestAvatarB18MessengerHawkGetsTwoPowerAfterTwoDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "Draw A", "Draw B", "Draw C")
	hawk := effectstest.B12Push(g, me.ID, "Messenger Hawk", "Creature — Bird Scout", avB18HawkOracle, 1, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, hawk, "flying") {
		t.Error("Messenger Hawk has no flying")
	}
	for g.TurnTallyFor(me.ID).CardsDrawn < 1 {
		effectstest.AvB9Draw(t, g, me.ID, 1)
	}
	if g.TurnTallyFor(me.ID).CardsDrawn == 1 {
		if got := effectstest.EffectivePower(t, g, hawk); got != 1 {
			t.Fatalf("power %d with one card drawn this turn, want 1", got)
		}
		effectstest.AvB9Draw(t, g, me.ID, 1)
	}
	if got := effectstest.EffectivePower(t, g, hawk); got != 3 {
		t.Errorf("power %d after two draws this turn, want 3", got)
	}
	if got := effectstest.EffectiveToughness(t, g, hawk); got != 2 {
		t.Errorf("toughness %d, want 2", got)
	}
}

func TestAvatarB18MessengerHawkIsAOneTwoBeforeTwoDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hawk := effectstest.B12Push(g, me.ID, "Messenger Hawk", "Creature — Bird Scout", avB18HawkOracle, 1, 2)
	if g.TurnTallyFor(me.ID).CardsDrawn >= 2 {
		t.Skip("the fixture already drew two cards this turn")
	}
	if got := effectstest.EffectivePower(t, g, hawk); got != 1 {
		t.Errorf("power %d, want 1", got)
	}
}

// --- Platypus-Bear -----------------------------------------------------------------

func TestAvatarB18PlatypusBearMillsTwoOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	before := me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Platypus-Bear", "Creature — Platypus Bear", avB18PlatypusOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - before; got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

func TestAvatarB18PlatypusBearAttacksOnlyWithALessonInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Push(g, me.ID, "Platypus-Bear", "Creature — Platypus Bear", avB18PlatypusOracle, 2, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Fatal("a defender attacked with no Lesson in the graveyard")
	}
	avB18MillALesson(t, g, me)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("Platypus-Bear could not attack with a Lesson in the graveyard: %v", err)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "defender") {
		t.Error("the waiver removed defender")
	}
}

// --- Pretending Poxbearers ---------------------------------------------------------

func TestAvatarB18PoxbearersMakeAnAllyWhenTheyDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pox := effectstest.B12Push(g, me.ID, "Pretending Poxbearers", "Creature — Human Citizen Ally", avB18PoxbearersOracle, 2, 1)
	effectstest.AvB18Destroy(t, g, pox)
	effectstest.PassPriorityAroundTable(t, g)
	if got := avB18Allies(g, me.ID); got != 1 {
		t.Fatalf("%d Ally tokens, want 1", got)
	}
	ally := effectstest.BattlefieldIDsNamed(g, avB18AllyTokenName)[0]
	c := effectstest.AvB9Live(t, g, ally)
	if c.Power != 1 || c.Toughness != 1 || !c.IsCreature() || !c.HasSubtype("Ally") {
		t.Errorf("want a 1/1 Ally creature: %+v", c)
	}
}

// --- Professor Zei, Anthropologist -------------------------------------------------

func TestAvatarB18ZeiRummagesAndRebuysOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Next Card")
	zei := effectstest.B12Push(g, me.ID, "Professor Zei, Anthropologist", "Legendary Creature — Human Advisor Ally", avB18ZeiOracle, 0, 3)
	hand := effectstest.AvB1Hand(me, game.Card{Name: "Spare", TypeLine: "Sorcery"})
	size := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, zei, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{hand}}); err != nil {
		t.Fatalf("rummage: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(hand) || me.Hand.Size() != size {
		t.Fatalf("discard one, draw one: hand %d (was %d), discarded in graveyard %v", me.Hand.Size(), size, me.Graveyard.Contains(hand))
	}

	// The rebuy: {1}, {T} — Zei is tapped from the rummage, so untap him.
	effectstest.UntapForTest(g, zei)
	bolt := effectstest.PushGraveyardPermanent(me, "Old Bolt", "Instant", "{R}")
	bear := effectstest.PushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	effectstest.FillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, zei, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.AvB8Card(bear)}); err == nil {
		t.Fatal("a creature card was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, zei, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.AvB8Card(bolt)}); err != nil {
		t.Fatalf("rebuy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant did not return to hand")
	}
	if g.Battlefield.Contains(zei) {
		t.Error("Zei was not sacrificed")
	}
}

func TestAvatarB18ZeiCantRebuyOnAnotherPlayersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]
	zei := effectstest.B12Push(g, other.ID, "Professor Zei, Anthropologist", "Legendary Creature — Human Advisor Ally", avB18ZeiOracle, 0, 3)
	bolt := effectstest.PushGraveyardPermanent(other, "Old Bolt", "Instant", "{R}")
	effectstest.FillPoolColored(other, "C", 1)
	if err := g.ActivateCatalogAbility(other.ID, zei, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.AvB8Card(bolt)}); err == nil {
		t.Fatal("Zei's rebuy was activated on another player's turn")
	}
}

// --- Sandbender Scavengers ---------------------------------------------------------

func TestAvatarB18ScavengersGrowOnSacrificesAndReanimateByPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	scav := effectstest.B12Push(g, me.ID, "Sandbender Scavengers", "Creature — Human Rogue", avB18ScavengersOracle, 1, 1)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AvB9Sacrifice(t, g, rock)
	if got := effectstest.CountersOn(g, scav, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after a sacrifice, want 1", got)
	}
	small := effectstest.PushGraveyardPermanent(me, "Two Drop", "Creature — Bear", "{1}{G}")
	big := effectstest.PushGraveyardPermanent(me, "Three Drop", "Creature — Bear", "{2}{G}")

	// Power 2 as it died: a mana value 3 card is not a legal target.
	effectstest.AvB18Destroy(t, g, scav)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	legal := effectstest.LatestPickTarget(g, me.ID).PickTargetCards
	if !effectstest.HasID(legal, small) || effectstest.HasID(legal, big) || effectstest.HasID(legal, scav) {
		t.Fatalf("legal targets %v: want the mana value 2 card only", legal)
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(scav) {
		t.Error("the Scavengers were not exiled")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the creature card did not return to the battlefield")
	}
}

func TestAvatarB18ScavengersIgnoreTheirOwnSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	scav := effectstest.B12Push(g, me.ID, "Sandbender Scavengers", "Creature — Human Rogue", avB18ScavengersOracle, 1, 1)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(scav); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
	g.RunStateChecksForTest()
	// Only the dies trigger: no counter trigger for itself.
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(scav) {
		t.Error("declining the exile leaves the Scavengers in the graveyard")
	}
}

// --- Sokka, Lateral Strategist -----------------------------------------------------

func TestAvatarB18SokkaDrawsOnlyWhenAttackingWithAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sokka := effectstest.B12Push(g, me.ID, "Sokka, Lateral Strategist", "Legendary Creature — Human Warrior Ally", avB18SokkaOracle, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "Drawn")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, sokka, bear)
	if n := effectstest.TriggersOnStackFrom(g, sokka); n != 1 {
		t.Fatalf("%d Sokka triggers, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if effectstest.CardByID(g, sokka).Tapped {
		t.Error("vigilance: Sokka tapped to attack")
	}
}

func TestAvatarB18SokkaAloneDoesNotDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sokka := effectstest.B12Push(g, me.ID, "Sokka, Lateral Strategist", "Legendary Creature — Human Warrior Ally", avB18SokkaOracle, 2, 4)
	effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, sokka)
	if n := effectstest.TriggersOnStackFrom(g, sokka); n != 0 {
		t.Fatalf("Sokka alone triggered %d times", n)
	}
}

func TestAvatarB18SokkaStayingHomeDoesNotDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sokka := effectstest.B12Push(g, me.ID, "Sokka, Lateral Strategist", "Legendary Creature — Human Warrior Ally", avB18SokkaOracle, 2, 4)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	if n := effectstest.TriggersOnStackFrom(g, sokka); n != 0 {
		t.Fatalf("Sokka triggered %d times without attacking", n)
	}
}

// --- Suki, Kyoshi Warrior ----------------------------------------------------------

func TestAvatarB18SukiCountsCreaturesAndMakesAnAttackingAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	suki := effectstest.B12Push(g, me.ID, "Suki, Kyoshi Warrior", "Legendary Creature — Human Warrior Ally", avB18SukiOracle, 0, 4)
	effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	if got := effectstest.EffectivePower(t, g, suki); got != 2 {
		t.Fatalf("power %d with two creatures, want 2", got)
	}
	effectstest.DeclareAttack(t, g, opp.ID, suki)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	allies := effectstest.BattlefieldIDsNamed(g, avB18AllyTokenName)
	if len(allies) != 1 {
		t.Fatalf("%d Ally tokens, want 1", len(allies))
	}
	tok := effectstest.CardByID(g, allies[0])
	if !tok.Tapped || tok.AttackingTarget != opp.ID {
		t.Errorf("want a tapped Ally attacking the defender: tapped %v, attacking %v", tok.Tapped, tok.AttackingTarget)
	}
	if got := effectstest.EffectivePower(t, g, suki); got != 3 {
		t.Errorf("power %d with three creatures, want 3", got)
	}
	if n := effectstest.TriggersOnStackFrom(g, suki); n != 0 {
		t.Error("the token was never declared, so Suki does not trigger again")
	}
}

// --- Sun Warriors ------------------------------------------------------------------

func TestAvatarB18SunWarriorsFirebendPerCreatureAndMakeAllies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sun := effectstest.B12Push(g, me.ID, "Sun Warriors", "Creature — Human Warrior Ally", avB18SunWarriorsOracle, 3, 5)
	effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Home Bear 2", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, sun)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 3})
	effectstest.MkAdvanceTo(t, g, game.StepEndCombat)
	effectstest.MkWantPool(t, me, map[string]int{"R": 3})
	effectstest.MkAdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.MkWantPool(t, me, map[string]int{})

	effectstest.FillPoolColored(me, "C", 4)
	if err := g.ActivateCatalogAbility(me.ID, sun, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the Ally ability was activated for four mana")
	}
	effectstest.FillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, sun, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate {5}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := avB18Allies(g, me.ID); got != 1 {
		t.Errorf("%d Ally tokens, want 1", got)
	}
}

// --- Tolls of War ------------------------------------------------------------------

func TestAvatarB18TollsOfWarClueThenOneAllyPerTurnOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	clues := len(avB5CluesB7(g, me.ID))
	effectstest.CastCatalogSpell(t, g, "Tolls of War", "Enchantment", avB18TollsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(avB5CluesB7(g, me.ID)); got != clues+1 {
		t.Fatalf("%d Clues, want %d", got, clues+1)
	}
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, me.ID, "Rock", "Artifact"))
	if got := avB18Allies(g, me.ID); got != 1 {
		t.Fatalf("%d Allies after the first sacrifice, want 1", got)
	}
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, me.ID, "Rock 2", "Artifact"))
	if got := avB18Allies(g, me.ID); got != 1 {
		t.Fatalf("%d Allies after a second sacrifice this turn, want 1 (once each turn)", got)
	}
}

func TestAvatarB18TollsOfWarIgnoresSacrificesOnOtherTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.B12Push(g, other.ID, "Tolls of War", "Enchantment", avB18TollsOracle, 0, 0)
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, other.ID, "Rock", "Artifact"))
	if got := avB18Allies(g, other.ID); got != 0 {
		t.Errorf("%d Allies from a sacrifice on another player's turn, want 0", got)
	}
}

// --- Uncle Iroh --------------------------------------------------------------------

func TestAvatarB18UncleIrohDiscountsLessonSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Uncle Iroh", "Legendary Creature — Human Noble Ally", avB18IrohOracle, 4, 2)
	effectstest.FillPoolColored(me, "R", 1)
	if _, err := effectstest.CastFromHandErr(g, me, "Plain Sorcery", "Sorcery", "{1}{R}", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a non-Lesson sorcery was discounted")
	}
	effectstest.CastFromHandForTest(t, g, me, "A Lesson", "Sorcery — Lesson", "{1}{R}", "", game.CastSpellParams{Strict: true})
}

func TestAvatarB18UncleIrohFirebendsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	iroh := effectstest.B12Push(g, me.ID, "Uncle Iroh", "Legendary Creature — Human Noble Ally", avB18IrohOracle, 4, 2)
	effectstest.DeclareAttack(t, g, opp.ID, iroh)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 1})
}

// --- Vindictive Warden -------------------------------------------------------------

func TestAvatarB18VindictiveWardenFirebendsAndPingsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warden := effectstest.B12Push(g, me.ID, "Vindictive Warden", "Creature — Human Soldier", avB18WardenOracle, 2, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, warden, "menace") {
		t.Error("no menace")
	}
	effectstest.DeclareAttack(t, g, opp.ID, warden)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 1})
	effectstest.MkAdvanceTo(t, g, game.StepPostcombatMain)
	before := effectstest.LifeOfOpponents(g)
	effectstest.FillPoolColored(me, "C", 2)
	if err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ping was activated for two mana")
	}
	effectstest.FillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	after := effectstest.LifeOfOpponents(g)
	for i := range before {
		if after[i] != before[i]-1 {
			t.Errorf("opponent %d life %d, want %d", i, after[i], before[i]-1)
		}
	}
}

// --- Wandering Musicians -----------------------------------------------------------

func TestAvatarB18WanderingMusiciansPumpTheTeamOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mus := effectstest.B12Push(g, me.ID, "Wandering Musicians", "Creature — Human Bard Ally", avB18MusiciansOracle, 2, 5)
	home := effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, mus)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, mus); got != 3 {
		t.Errorf("Musicians power %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, home); got != 3 {
		t.Errorf("your other creature power %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature power %d, want 2", got)
	}
}

// --- White Lotus Reinforcements ----------------------------------------------------

func TestAvatarB18WhiteLotusPumpsOtherAlliesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lotus := effectstest.B12Push(g, me.ID, "White Lotus Reinforcements", "Creature — Human Soldier Ally", avB18WhiteLotusOracle, 2, 3)
	ally := effectstest.B12Creature(g, me.ID, "My Ally", "Creature — Human Ally", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirAlly := effectstest.B12Creature(g, opp.ID, "Their Ally", "Creature — Human Ally", 1, 1)
	if p, tg := effectstest.EffectivePower(t, g, ally), effectstest.EffectiveToughness(t, g, ally); p != 2 || tg != 2 {
		t.Errorf("your Ally %d/%d, want 2/2", p, tg)
	}
	if p := effectstest.EffectivePower(t, g, lotus); p != 2 {
		t.Errorf("White Lotus pumps itself: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Ally got pumped: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, theirAlly); p != 1 {
		t.Errorf("an opponent's Ally got pumped: power %d", p)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, lotus, "vigilance") {
		t.Error("no vigilance")
	}
}

// --- Zhao, Ruthless Admiral --------------------------------------------------------

func TestAvatarB18ZhaoFirebendsTwoAndPumpsOnSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zhao := effectstest.B12Push(g, me.ID, "Zhao, Ruthless Admiral", "Legendary Creature — Human Soldier", avB18ZhaoOracle, 3, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, zhao)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 2})

	effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, me.ID, "Rock", "Artifact"))
	if got := effectstest.EffectivePower(t, g, zhao); got != 4 {
		t.Errorf("Zhao power %d after a sacrifice, want 4", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("bear power %d after a sacrifice, want 3", got)
	}
}

func TestAvatarB18ZhaoIgnoresItsOwnSacrificeAndOthersSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	zhao := effectstest.B12Push(g, me.ID, "Zhao, Ruthless Admiral", "Legendary Creature — Human Soldier", avB18ZhaoOracle, 3, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact"))
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("an opponent's sacrifice pumped: power %d", got)
	}
	effectstest.AvB9Sacrifice(t, g, zhao)
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("Zhao's own sacrifice pumped: power %d", got)
	}
}

// --- Barrels of Blasting Jelly -----------------------------------------------------

func TestAvatarB18BarrelsFilterOnceEachTurnAndBlast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	barrels := effectstest.B12Push(g, me.ID, "Barrels of Blasting Jelly", "Artifact", avB18BarrelsOracle, 0, 0)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "C", 2)
	if err := g.ActivateManaAbility(me.ID, barrels, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("{1}: add {G}: %v", err)
	}
	effectstest.MkWantPool(t, me, map[string]int{"C": 1, "G": 1})
	if err := g.ActivateManaAbility(me.ID, barrels, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("a second activation this turn was allowed")
	}

	victim := effectstest.B12Creature(g, opp.ID, "Their Big", "Creature — Giant", 5, 5)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "C", 5)
	if err := g.ActivateCatalogAbility(me.ID, barrels, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.AvB8Card(victim)}); err != nil {
		t.Fatalf("blast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	clues := len(avB5CluesB7(g, me.ID))

	// Uncrewed, it is not a creature: destroying it is not dying.
	idle := effectstest.B12Push(g, me.ID, "Fire Nation Warship", "Artifact — Vehicle", avB18WarshipOracle, 4, 4)
	effectstest.AvB18Destroy(t, g, idle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(avB5CluesB7(g, me.ID)); got != clues {
		t.Fatalf("an uncrewed Warship made a Clue (%d, want %d)", got, clues)
	}

	ship := effectstest.B12Push(g, me.ID, "Fire Nation Warship", "Artifact — Vehicle", avB18WarshipOracle, 4, 4)
	small := effectstest.B12Creature(g, me.ID, "Deckhand", "Creature — Human", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Fatal("power 1 crewed Crew 2")
	}
	crew := effectstest.B12Creature(g, me.ID, "Sailor", "Creature — Human", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.AvB9Live(t, g, ship)
	if !c.IsCreature() || !game.HasKeyword(&c, "reach") {
		t.Fatalf("a crewed Warship is a creature with reach: %+v", c)
	}
	effectstest.AvB18Destroy(t, g, ship)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(avB5CluesB7(g, me.ID)); got != clues+1 {
		t.Errorf("%d Clues after the crewed Warship died, want %d", got, clues+1)
	}
}

// --- Kyoshi Battle Fan -------------------------------------------------------------

func TestAvatarB18BattleFanMakesAnAllyAndEquipsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	fan := effectstest.CastCatalogSpell(t, g, "Kyoshi Battle Fan", "Artifact — Equipment", avB18BattleFanOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	allies := effectstest.BattlefieldIDsNamed(g, avB18AllyTokenName)
	if len(allies) != 1 {
		t.Fatalf("%d Ally tokens, want 1", len(allies))
	}
	ally := allies[0]
	if att := effectstest.CardByID(g, fan).AttachedTo; att.Kind != game.TargetCard || att.ID != ally {
		t.Fatalf("the Fan is attached to %+v, want the Ally", att)
	}
	if p, tg := effectstest.EffectivePower(t, g, ally), effectstest.EffectiveToughness(t, g, ally); p != 2 || tg != 1 {
		t.Errorf("equipped Ally %d/%d, want 2/1", p, tg)
	}

	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.FillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, fan, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.AvB8Card(bear)}); err == nil {
		t.Fatal("equip {2} was paid with one mana")
	}
	effectstest.FillPoolColored(me, "C", 1)
	if err := g.ActivateCatalogAbility(me.ID, fan, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.AvB8Card(bear)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("equipped bear power %d, want 3", got)
	}
}

// --- Meteor Sword ------------------------------------------------------------------

func TestAvatarB18MeteorSwordDestroysAPermanentAndPumpsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	relic := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Enchantment")
	sword := effectstest.CastCatalogSpell(t, g, "Meteor Sword", "Artifact — Equipment", avB18MeteorSwordOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Fatal("the enchantment was not destroyed")
	}
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.FillPoolColored(me, "C", 3)
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.AvB8Card(bear)}); err != nil {
		t.Fatalf("equip {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tg != 5 {
		t.Errorf("equipped bear %d/%d, want 5/5", p, tg)
	}
}
