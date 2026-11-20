package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b16_test.go — engine-driven coverage for S82 set ONE, batch 16.

const (
	s82b16RotpriestOracle = "eafd7ba5-691d-4f8d-94bb-70a5612245a1"
	s82b16KinzuOracle     = "e4f4c41b-2533-497f-96bb-7083ef597268"
	s82b16AssaultOracle   = "95c0611a-b6f1-4cb8-8b47-73aa9c2a7d2c"
	s82b16DecreeOracle    = "92c43ee1-0f7d-4993-8d3a-689cc87d7480"
	s82b16GrafterOracle   = "ba922414-878f-4261-9bce-500c5c9e9506"
	s82b16RetrieverOracle = "d099be2e-1879-4bb0-a27d-a386308387c0"
	s82b16WurmquakeOracle = "0eeb9a62-1b36-42f5-8820-3331fa5ca3a6"
)

func TestS82B16VeneratedRotpriestPoisonsAnOpponentWhenYourCreatureIsTargetedBySpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	priest := effectstest.B12Push(g, me.ID, "Venerated Rotpriest", "Creature — Phyrexian Druid", s82b16RotpriestOracle, 1, 2)
	effectstest.AssertKeywords(t, g, priest, "toxic 1")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// An opponent's spell at my Bear: the trigger asks me for an opponent.
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
			t.Fatal("I was accepted as the target opponent")
		}
	}
	effectstest.PickPlayer(t, g, me.ID, other.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := other.Counters[game.CounterPoison]; got != 1 {
		t.Errorf("the chosen opponent has %d poison counters, want 1", got)
	}
	if got := opp.Counters[game.CounterPoison]; got != 0 {
		t.Errorf("the caster was not chosen: %d poison counters", got)
	}

	// The Rotpriest itself counts.
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: priest})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, other.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := other.Counters[game.CounterPoison]; got != 2 {
		t.Errorf("targeting the Rotpriest itself triggers it: %d poison counters, want 2", got)
	}

	// A creature an opponent controls does not.
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || other.Counters[game.CounterPoison] != 2 {
		t.Error("an opponent's creature being targeted does not trigger it")
	}
}

func TestS82B16CardsAreRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82b16RotpriestOracle: "Venerated Rotpriest",
		s82b16KinzuOracle:     "Kinzu of the Bleak Coven",
		s82b16AssaultOracle:   "Noxious Assault",
		s82b16DecreeOracle:    "Norn's Decree",
		s82b16GrafterOracle:   "Contaminant Grafter",
		s82b16RetrieverOracle: "Glissa's Retriever",
		s82b16WurmquakeOracle: "Wurmquake",
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

// --- Kinzu of the Bleak Coven ---------------------------------------------------

func TestS82B16KinzuPaysLifeExilesAndMakesAOneOneToxicCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Kinzu of the Bleak Coven", "Legendary Creature — Phyrexian Vampire", s82b16KinzuOracle, 5, 4)
	bear := effectstest.B12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", effectstest.SocB7GrizzlyBearsOracle, 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the dies trigger did not offer the 2 life")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want 2 paid", life, me.Life)
	}
	if effectstest.SocB7Zone(g, bear) != game.ZoneExile {
		t.Fatalf("the dead Bear is in %q, want exile", effectstest.SocB7Zone(g, bear))
	}
	toks := effectstest.S47Tokens(g, "Grizzly Bears")
	if len(toks) != 1 {
		t.Fatalf("Bear tokens = %d, want 1", len(toks))
	}
	if p, tg := effectstest.EffectivePTOf(t, g, toks[0].InstanceID); p != 1 || tg != 1 {
		t.Errorf("the copy is %d/%d, want 1/1", p, tg)
	}
	effectstest.AssertKeywords(t, g, toks[0].InstanceID, "toxic 1")
}

func TestS82B16KinzuDecliningPaysNothingAndKeepsTheCardInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Kinzu of the Bleak Coven", "Legendary Creature — Phyrexian Vampire", s82b16KinzuOracle, 5, 4)
	bear := effectstest.B12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", effectstest.SocB7GrizzlyBearsOracle, 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life || effectstest.SocB7Zone(g, bear) != game.ZoneGraveyard || len(effectstest.S47Tokens(g, "Grizzly Bears")) != 0 {
		t.Errorf("declined: life %d → %d, Bear in %q", life, me.Life, effectstest.SocB7Zone(g, bear))
	}
}

func TestS82B16KinzuIsNotAskedWithoutTwoLifeAndIgnoresOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kinzu := effectstest.B12Push(g, me.ID, "Kinzu of the Bleak Coven", "Legendary Creature — Phyrexian Vampire", s82b16KinzuOracle, 5, 4)
	theirs := effectstest.B12Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", effectstest.SocB7GrizzlyBearsOracle, 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("an opponent's creature dying triggered Kinzu")
	}
	g.WithWriteLock(func() { me.Life = 1 })
	bear := effectstest.B12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", effectstest.SocB7GrizzlyBearsOracle, 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) || len(effectstest.S47Tokens(g, "Grizzly Bears")) != 0 {
		t.Error("at 1 life the 2 life can't be paid, so nothing is offered or made")
	}
	effectstest.KillCreature(t, g, me.ID, kinzu)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("\"another\": Kinzu dying does not trigger itself")
	}
}

// --- Noxious Assault ----------------------------------------------------------------

func TestS82B16NoxiousAssaultPumpsYourCreaturesAndPoisonsEachBlockerOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	spare := effectstest.B12Creature(g, other.ID, "Other Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Noxious Assault", "Sorcery", s82b16AssaultOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePTOf(t, g, mine); p != 4 || tg != 4 {
		t.Errorf("my Bear is %d/%d, want 4/4", p, tg)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 3 {
		t.Errorf("an opponent's creature is pumped: power %d", p)
	}
	late := effectstest.B12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, late); p != 2 {
		t.Errorf("a creature that arrives later is not pumped (CR 611.2c): power %d", p)
	}
	effectstest.DeclareAttack(t, g, opp.ID, mine)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(theirs, mine); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Poison != 1 {
		t.Errorf("the blocker's controller has %d poison, want 1", opp.Poison)
	}
	if other.Poison != 0 || me.Poison != 0 {
		t.Errorf("only the blocker's controller is poisoned: %d / %d", other.Poison, me.Poison)
	}
	_ = spare
}

// --- Norn's Decree ----------------------------------------------------------------------

func TestS82B16NornsDecreePoisonsAnOpponentWhoseCreaturesHitYouOncePerStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Norn's Decree", "Enchantment", s82b16DecreeOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "Raider A", "Creature — Human", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Raider B", "Creature — Human", 2, 2)
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, me.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Poison != 1 {
		t.Errorf("two creatures dealt damage in one step: %d poison, want exactly 1", opp.Poison)
	}
}

func TestS82B16NornsDecreeDrawsForAnAttackerOnlyWhenAPoisonedPlayerIsAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Norn's Decree", "Enchantment", s82b16DecreeOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "My Raider", "Creature — Human", 2, 2)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("an unpoisoned defender: hand %d → %d", hand, me.Hand.Size())
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.B12Push(g2, opp2.ID, "Norn's Decree", "Enchantment", s82b16DecreeOracle, 0, 0)
	effectstest.AdvanceToMain(t, g2)
	if err := g2.SetPoison(opp2.ID, 1); err != nil {
		t.Fatal(err)
	}
	a2 := effectstest.B12Creature(g2, me2.ID, "My Raider", "Creature — Human", 2, 2)
	b2 := effectstest.B12Creature(g2, me2.ID, "My Other Raider", "Creature — Human", 2, 2)
	hand2 := me2.Hand.Size()
	effectstest.DeclareAttack(t, g2, opp2.ID, a2, b2)
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Hand.Size() != hand2+1 {
		t.Errorf("the attacking player (not the Decree's controller) draws one card: hand %d → %d", hand2, me2.Hand.Size())
	}
}

// --- Contaminant Grafter ----------------------------------------------------------------

func TestS82B16ContaminantGrafterProliferatesOncePerDamageStepAndHasCorruptedEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	graft := effectstest.B12Push(g, me.ID, "Contaminant Grafter", "Creature — Phyrexian Druid", s82b16GrafterOracle, 5, 5)
	effectstest.AssertKeywords(t, g, graft, "trample", "toxic 1")
	extra := effectstest.B12Creature(g, me.ID, "Extra", "Creature — Human", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(graft, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(extra, other.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	effectstest.SettleAnsweringProliferate(t, g)
	// Toxic 1 gave the first poison; one proliferate (one trigger for both
	// players hit) adds a second to that player.
	if opp.Poison != 2 {
		t.Errorf("toxic then a single proliferate: %d poison, want 2", opp.Poison)
	}
}

func TestS82B16ContaminantGrafterCorruptedDrawsAndMayPutALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Contaminant Grafter", "Creature — Phyrexian Druid", s82b16GrafterOracle, 5, 5)
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := g.SetPoison(opp.ID, 2); err != nil {
		t.Fatal(err)
	}
	hand := me.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	_ = land
	if me.Hand.Size() != hand {
		t.Fatalf("two poison counters is not corrupted: hand %d → %d", hand, me.Hand.Size())
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.B12Push(g2, me2.ID, "Contaminant Grafter", "Creature — Phyrexian Druid", s82b16GrafterOracle, 5, 5)
	land2 := uuid.New()
	me2.Hand.PushTop(game.Card{InstanceID: land2, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me2.ID, Controller: me2.ID})
	if err := g2.SetPoison(opp2.ID, 3); err != nil {
		t.Fatal(err)
	}
	hand2 := me2.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, g2, 0)
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Hand.Size() != hand2+1 {
		t.Errorf("corrupted draws a card: hand %d → %d", hand2, me2.Hand.Size())
	}
	effectstest.AnswerChooseCards(t, g2, me2.ID, land2)
	if !g2.Battlefield.Contains(land2) {
		t.Error("the land chosen after the draw is not on the battlefield")
	}
}

// --- Glissa's Retriever ------------------------------------------------------------------

func TestS82B16GlissasRetrieverCannotBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	r := effectstest.B12Push(g, me.ID, "Glissa's Retriever", "Creature — Phyrexian Beast", s82b16RetrieverOracle, 6, 6)
	effectstest.AssertKeywords(t, g, r, "haste", "toxic 3")
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Human", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, r)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, r); err == nil {
		t.Error("a power 2 creature blocked it")
	}
	if err := g.DeclareBlocker(big, r); err != nil {
		t.Errorf("a power 3 creature could not block it: %v", err)
	}
}

func TestS82B16GlissasRetrieverExilesItselfAndReturnsOneCardPerCorruptedOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	r := effectstest.B12Push(g, me.ID, "Glissa's Retriever", "Creature — Phyrexian Beast", s82b16RetrieverOracle, 6, 6)
	if err := g.SetPoison(opp.ID, 3); err != nil {
		t.Fatal(err)
	}
	if err := g.SetPoison(other.ID, 2); err != nil {
		t.Fatal(err)
	}
	a := effectstest.PushGraveyardCardForTest(me, "Card A")
	b := effectstest.PushGraveyardCardForTest(me, "Card B")
	effectstest.KillCreature(t, g, me.ID, r)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SocB7Zone(g, r) != game.ZoneExile {
		t.Fatalf("the Retriever is in %q, want exile", effectstest.SocB7Zone(g, r))
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	if len(p.PickTargetCards) == 0 || effectstest.HasID(p.PickTargetCards, r) {
		t.Errorf("candidates %v: the exiled Retriever can't be chosen", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, a)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("one corrupted opponent allows one target, not two")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || me.Hand.Contains(b) {
		t.Errorf("card A in hand %v, card B in hand %v", me.Hand.Contains(a), me.Hand.Contains(b))
	}
}

func TestS82B16GlissasRetrieverDoesNotRaiseTheReturnWithoutACorruptedOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	r := effectstest.B12Push(g, me.ID, "Glissa's Retriever", "Creature — Phyrexian Beast", s82b16RetrieverOracle, 6, 6)
	a := effectstest.PushGraveyardCardForTest(me, "Card A")
	effectstest.KillCreature(t, g, me.ID, r)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SocB7Zone(g, r) != game.ZoneExile {
		t.Errorf("the Retriever is exiled whatever the poison: %q", effectstest.SocB7Zone(g, r))
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil || me.Hand.Contains(a) {
		t.Error("X is 0: nothing can be returned")
	}
}

// --- Wurmquake ------------------------------------------------------------------------------

func s82b16Wurms(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Wurm" && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}

func s82b16CastWurmquake(t *testing.T, g *game.Game, me *game.Player, from game.ZoneKind) {
	t.Helper()
	id := uuid.New()
	card := game.Card{InstanceID: id, Name: "Wurmquake", TypeLine: "Sorcery", ManaCost: "{4}{G}{G}",
		OracleID: s82b16WurmquakeOracle, Owner: me.ID, Controller: me.ID}
	params := game.CastSpellParams{Strict: true}
	mana := "GGGGGG"
	if from == game.ZoneGraveyard {
		me.Graveyard.PushTop(card)
		params.FromZone, params.AlternativeCost = string(game.ZoneGraveyard), "flashback"
		mana = "GGGGGGGGGG"
	} else {
		me.Hand.PushTop(card)
	}
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, mana)
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast Wurmquake from %s: %v", from, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
}

func TestS82B16WurmquakeMakesOneSixSixWurmAndOneMorePerCorruptedOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s82b16CastWurmquake(t, g, me, game.ZoneHand)
	wurms := s82b16Wurms(g, me.ID)
	if len(wurms) != 1 {
		t.Fatalf("%d Wurms, want 1 with nobody corrupted", len(wurms))
	}
	if p, tg := effectstest.EffectivePTOf(t, g, wurms[0].InstanceID); p != 6 || tg != 6 {
		t.Errorf("the Wurm is %d/%d, want 6/6", p, tg)
	}
	effectstest.AssertKeywords(t, g, wurms[0].InstanceID, "trample", "toxic 1")

	g2 := effectstest.NewCatalogGame(t)
	me2, a, b, c := g2.Seats[0], g2.Seats[1], g2.Seats[2], g2.Seats[3]
	for _, p := range []*game.Player{a, b} {
		if err := g2.SetPoison(p.ID, 3); err != nil {
			t.Fatal(err)
		}
	}
	if err := g2.SetPoison(c.ID, 2); err != nil {
		t.Fatal(err)
	}
	s82b16CastWurmquake(t, g2, me2, game.ZoneHand)
	if got := len(s82b16Wurms(g2, me2.ID)); got != 3 {
		t.Errorf("%d Wurms, want 1 + 2 corrupted opponents", got)
	}
}

func TestS82B16WurmquakeFlashbackMakesTenTenWurms(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s82b16CastWurmquake(t, g, me, game.ZoneGraveyard)
	wurms := s82b16Wurms(g, me.ID)
	if len(wurms) != 1 {
		t.Fatalf("%d Wurms, want 1", len(wurms))
	}
	if p, tg := effectstest.EffectivePTOf(t, g, wurms[0].InstanceID); p != 10 || tg != 10 {
		t.Errorf("the flashback Wurm is %d/%d, want 10/10 ({8}{G}{G} spent)", p, tg)
	}
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Wurmquake" {
			t.Error("the flashed-back Wurmquake stays in the graveyard instead of being exiled")
		}
	}
}
