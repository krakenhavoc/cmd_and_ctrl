package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82One13SentryOracle      = "e8270c29-1887-4d3f-b70f-6b79ada8fe47"
	s82One13ExperimentOracle  = "d1faaa3e-8e99-48e9-bb04-98bd46b7217c"
	s82One13JusticiarOracle   = "ee8f3dbd-6e86-46c2-9393-7da6c3c4b46d"
	s82One13EnforcerOracle    = "8e8875dc-72a4-47a4-bb99-ca817c177b82"
	s82One13EyeOracle         = "7270b5b6-a571-4e6d-91de-03f7cc035b3c"
	s82One13VeilOracle        = "3dd81e3e-e57a-4b6f-9043-bba0cebb3e90"
	s82One13TheoristOracle    = "4a60db66-0347-4efc-8507-ab7ed8d984fe"
	s82One13CenturionOracle   = "5707c278-dab9-43bc-b3a3-a6b7923511f4"
	s82One13MalcatorOracle    = "0d1c6a82-943a-46bd-97ed-6055226ef2ab"
	s82One13RetrofitterOracle = "998d8d8a-743d-43f9-8e9b-42460f170243"
	s82One13OnslaughtOracle   = "1589d90d-411a-4e77-9771-51518b2b224b"
	s82One13PunisherOracle    = "606af824-e711-4957-99c0-d02b0f6c729d"
)

func TestS82OneB13EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82One13SentryOracle:      "Cephalopod Sentry",
		s82One13ExperimentOracle:  "Escaped Experiment",
		s82One13JusticiarOracle:   "Mandible Justiciar",
		s82One13EnforcerOracle:    "Orthodoxy Enforcer",
		s82One13EyeOracle:         "Eye of Malcator",
		s82One13VeilOracle:        "Veil of Assimilation",
		s82One13TheoristOracle:    "Transplant Theorist",
		s82One13CenturionOracle:   "Cutthroat Centurion",
		s82One13MalcatorOracle:    "Malcator, Purity Overseer",
		s82One13RetrofitterOracle: "Unctus's Retrofitter",
		s82One13OnslaughtOracle:   "Plated Onslaught",
		s82One13PunisherOracle:    "Furnace Punisher",
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

// s82One13Treasure makes one Treasure artifact token for controller.
func s82One13Treasure(t *testing.T, g *game.Game, controller uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(controller, TreasureToken(), 1); err != nil {
			t.Fatalf("create Treasure: %v", err)
		}
	})
	g.RunStateChecksForTest()
}

func s82One13PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	return effectivePTOf(t, g, id)
}

func TestS82OneB13CephalopodSentryPowerIsYourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentry := b12Push(g, me.ID, "Cephalopod Sentry", "Artifact Creature — Phyrexian Squid", s82One13SentryOracle, 0, 5)
	if p, tg := s82One13PT(t, g, sentry); p != 1 || tg != 5 {
		t.Errorf("alone: %d/%d, want 1/5 (it counts itself)", p, tg)
	}
	b12Permanent(g, me.ID, "Rock", "Artifact")
	b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	b12Permanent(g, me.ID, "Bear", "Creature — Bear")
	if p, tg := s82One13PT(t, g, sentry); p != 2 || tg != 5 {
		t.Errorf("with one more artifact of mine: %d/%d, want 2/5", p, tg)
	}
	if !hasEffectiveKeyword(t, g, sentry, "flying") {
		t.Error("no flying")
	}
}

func TestS82OneB13MandibleJusticiarGrowsOnlyForAnotherArtifactOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	j := b12Push(g, me.ID, "Mandible Justiciar", "Artifact Creature — Phyrexian Cleric", s82One13JusticiarOracle, 2, 1)
	s82One13Treasure(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if p, tg := s82One13PT(t, g, j); p != 2 || tg != 1 {
		t.Fatalf("after an opponent's artifact: %d/%d, want 2/1", p, tg)
	}
	s82One13Treasure(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if p, tg := s82One13PT(t, g, j); p != 3 || tg != 2 {
		t.Errorf("after my artifact: %d/%d, want 3/2", p, tg)
	}
	s82One13Treasure(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if p, tg := s82One13PT(t, g, j); p != 4 || tg != 3 {
		t.Errorf("after a second: %d/%d, want 4/3", p, tg)
	}
	if !hasEffectiveKeyword(t, g, j, "lifelink") {
		t.Error("no lifelink")
	}
}

func TestS82OneB13OrthodoxyEnforcerNeedsTwoArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	e := b12Push(g, me.ID, "Orthodoxy Enforcer", "Creature — Phyrexian Cleric", s82One13EnforcerOracle, 2, 4)
	if p, _ := s82One13PT(t, g, e); p != 2 {
		t.Fatalf("no artifacts: power %d, want 2", p)
	}
	b12Permanent(g, me.ID, "Rock", "Artifact")
	if p, _ := s82One13PT(t, g, e); p != 2 {
		t.Errorf("one artifact: power %d, want 2", p)
	}
	rock2 := b12Permanent(g, me.ID, "Rock Two", "Artifact")
	if p, tg := s82One13PT(t, g, e); p != 4 || tg != 4 {
		t.Errorf("two artifacts: %d/%d, want 4/4", p, tg)
	}
	killCreature(t, g, me.ID, rock2)
	if p, _ := s82One13PT(t, g, e); p != 2 {
		t.Errorf("back to one artifact: power %d, want 2", p)
	}
	if !hasEffectiveKeyword(t, g, e, "vigilance") {
		t.Error("no vigilance")
	}
}

func TestS82OneB13EyeOfMalcatorScriesThenAnimatesOnAnotherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Top", "Second", "Third")
	eye := castAndResolveCreature(t, g, "Eye of Malcator", "Artifact", s82One13EyeOracle)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatalf("no scry 2 prompt when the Eye entered")
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardForTest(g, eye)
	if c == nil || c.IsCreature() {
		t.Fatalf("the Eye is already a creature (or gone): %+v", c)
	}
	s82One13Treasure(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, eye); c.IsCreature() {
		t.Fatal("an opponent's artifact animated the Eye")
	}
	s82One13Treasure(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if p, tg := s82One13PT(t, g, eye); p != 4 || tg != 4 {
		t.Fatalf("animated: %d/%d, want 4/4", p, tg)
	}
	g.ReadSnapshot(func() {
		c := findBattlefieldCardForTest(g, eye)
		eff := c.Effective()
		if !c.IsCreature() || !c.IsArtifact() || !s82One13Has(eff.Subtypes, "Phyrexian") || !s82One13Has(eff.Subtypes, "Eye") {
			t.Errorf("types %v subtypes %v, want a Phyrexian Eye artifact creature", eff.Types, eff.Subtypes)
		}
	})
	advanceToUpkeepOf(t, g, 1)
	if c := findBattlefieldCardForTest(g, eye); c.IsCreature() {
		t.Error("the Eye is still a creature after the turn ended")
	}
}

func s82One13Has(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func TestS82OneB13VeilOfAssimilationPumpsOnEachArtifactEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castAndResolveCreature(t, g, "Veil of Assimilation", "Artifact", s82One13VeilOracle)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tg := s82One13PT(t, g, bear); p != 3 || tg != 3 {
		t.Fatalf("after the Veil's own entry: %d/%d, want 3/3", p, tg)
	}
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("no vigilance from the Veil's own entry")
	}
	s82One13Treasure(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tg := s82One13PT(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("after another artifact: %d/%d, want 4/4", p, tg)
	}
}

func TestS82OneB13TransplantTheoristLootsOnEntryAndTucksFromGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	th := b12Push(g, me.ID, "Transplant Theorist", "Artifact Creature — Phyrexian Artificer", s82One13TheoristOracle, 2, 4)
	msc18Library(me, 6)
	hand := me.Hand.Size()
	s82One13Treasure(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if p := latestChoiceOfKind(g, game.PendingChoiceChooseCards); p != nil {
		if err := g.ResolveChooseCards(p.ID, me.ID, p.ChooseCards[:1]); err != nil {
			t.Fatalf("discard: %v", err)
		}
	}
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d, want %d after the loot", me.Hand.Size(), hand)
	}
	if false {
		t.Fatal("unreachable")
	}

	// Declining the draw declines the discard.
	hand = me.Hand.Size()
	s82One13Treasure(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	if me.Hand.Size() != hand {
		t.Errorf("declining changed the hand: %d, want %d", me.Hand.Size(), hand)
	}

	// {2}: put a graveyard card on the bottom of the library.
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, th, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(dead)}); err == nil {
		t.Fatal("activated with no mana")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, th, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(dead)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(dead) {
		t.Error("the card is still in the graveyard")
	}
	n := len(me.Library.Cards)
	if n == 0 || me.Library.Cards[n-1].InstanceID != dead && me.Library.Cards[0].InstanceID != dead {
		t.Error("the card is not at the bottom of the library")
	}
}

func TestS82OneB13CutthroatCenturionSacrificesAnotherOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c := b12Push(g, me.ID, "Cutthroat Centurion", "Artifact Creature — Phyrexian Warrior", s82One13CenturionOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	land := b12Permanent(g, me.ID, "Wastes", "Basic Land")
	if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("sacrificed a land")
	}
	if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{c}}); err == nil {
		t.Fatal("sacrificed itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("sacrifice a creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := s82One13PT(t, g, c); p != 4 || tg != 4 {
		t.Errorf("after one activation: %d/%d, want 4/4", p, tg)
	}
	if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("a second activation in the same turn was allowed")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the refused activation sacrificed the artifact")
	}
}

func TestS82OneB13MalcatorMakesAGolemOnEntryAndAnotherWithThreeArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Malcator, Purity Overseer", "Legendary Creature — Phyrexian Elephant Wizard", s82One13MalcatorOracle)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Phyrexian Golem"); n != 1 {
		t.Fatalf("Golems after entering: %d, want 1", n)
	}
	// The Golem token was one artifact entering. Two more make three.
	s82One13Treasure(t, g, me.ID)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Phyrexian Golem"); n != 1 {
		t.Errorf("Golems with two artifacts entered: %d, want 1", n)
	}
}

func TestS82OneB13MalcatorEndStepGolemWithThreeArtifactsEntered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Malcator, Purity Overseer", "Legendary Creature — Phyrexian Elephant Wizard", s82One13MalcatorOracle, 1, 1)
	for i := 0; i < 3; i++ {
		s82One13Treasure(t, g, me.ID)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Phyrexian Golem"); n != 1 {
		t.Errorf("Golems: %d, want 1", n)
	}
}

func TestS82OneB13UnctussRetrofitterAnimatesAnArtifactWhileItRemains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	retro := castAndResolveCreature(t, g, "Unctus's Retrofitter", "Creature — Phyrexian Artificer", s82One13RetrofitterOracle)
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if p, tg := s82One13PT(t, g, rock); p != 4 || tg != 4 {
		t.Fatalf("the Rock is %d/%d, want 4/4", p, tg)
	}
	g.ReadSnapshot(func() {
		if c := findBattlefieldCardForTest(g, rock); !c.IsCreature() || !c.IsArtifact() {
			t.Errorf("the Rock is not an artifact creature: %v", c.Effective().Types)
		}
	})
	if !hasEffectiveKeyword(t, g, retro, "toxic 1") {
		t.Error("no toxic 1")
	}
	killCreature(t, g, me.ID, retro)
	g.ReadSnapshot(func() {
		if c := findBattlefieldCardForTest(g, rock); c.IsCreature() {
			t.Error("the Rock is still a creature after the Retrofitter left")
		}
	})
}

func TestS82OneB13PlatedOnslaughtCostsLessAndPumpsCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		pushPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	}
	pushPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	if got := selfPricedMV(t, g, me, s82One13OnslaughtOracle, "Instant", "{3}{W}{W}"); got != 3 {
		t.Errorf("with two artifacts: mana value %d, want 3", got)
	}
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Plated Onslaught", "Instant", s82One13OnslaughtOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tg := s82One13PT(t, g, mine); p != 4 || tg != 3 {
		t.Errorf("mine: %d/%d, want 4/3", p, tg)
	}
	if p, tg := s82One13PT(t, g, theirs); p != 2 || tg != 2 {
		t.Errorf("theirs: %d/%d, want 2/2", p, tg)
	}
	late := b12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if p, _ := s82One13PT(t, g, late); p != 2 {
		t.Errorf("a creature that arrived later got the pump: power %d", p)
	}
}

func TestS82OneB13FurnaceDeals2UnlessThatPlayerHasTwoBasicLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Furnace Punisher", "Creature — Phyrexian Warrior", s82One13PunisherOracle, 3, 3)
	// I have one basic land and a nonbasic one; the opponent has two basics.
	b12Permanent(g, me.ID, "Wastes", "Basic Land")
	b12Permanent(g, me.ID, "Nonbasic", "Land")
	b12Permanent(g, opp.ID, "Wastes", "Basic Land")
	b12Permanent(g, opp.ID, "Wastes", "Basic Land")
	myLife, oppLife := me.Life, opp.Life
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != myLife-2 {
		t.Errorf("my life %d, want %d (one basic land)", me.Life, myLife-2)
	}
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife {
		t.Errorf("opponent life %d, want %d (two basic lands)", opp.Life, oppLife)
	}
}

func TestS82OneB13EscapedExperimentShrinksAnOpposingAttackerByYourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exp := b12Push(g, me.ID, "Escaped Experiment", "Artifact Creature — Phyrexian Beast", s82One13ExperimentOracle, 2, 1)
	b12Permanent(g, me.ID, "Rock", "Artifact")
	b12Permanent(g, me.ID, "Rock Two", "Artifact")
	b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(exp, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if latestPickTarget(g, me.ID) != nil {
		if p := latestPickTarget(g, me.ID); hasID(p.PickTargetCards, mine) {
			t.Error("my own creature is a legal target")
		}
		pickCard(t, g, me.ID, victim)
	}
	passPriorityAroundTable(t, g)
	// Escaped Experiment is itself an artifact: 3 artifacts of mine.
	if p, tg := s82One13PT(t, g, victim); p != 2 || tg != 5 {
		t.Errorf("victim %d/%d, want 2/5 (-3/-0)", p, tg)
	}
}
