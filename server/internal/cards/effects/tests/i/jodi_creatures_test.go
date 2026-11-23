package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the Bringers ------------------------------------------------------

func TestBringersCanBeCastForTheFiveColourCost(t *testing.T) {
	for _, c := range []struct{ name, oracle, cost string }{
		{"Bringer of the Black Dawn", effectstest.BringerBlackDawnOracle, "{7}{B}{B}"},
		{"Bringer of the Blue Dawn", effectstest.BringerBlueDawnOracle, "{7}{U}{U}"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.PutInHand(me, game.Card{Name: c.name, TypeLine: "Creature — Bringer",
				OracleID: c.oracle, ManaCost: c.cost, Power: 5, Toughness: 5})
			effectstest.ToMain(t, g)

			if ac := game.AlternativeCostByKey(c.oracle, "bringer-wubrg"); ac == nil || ac.ManaCost != "{W}{U}{B}{R}{G}" {
				t.Fatalf("alternative cost = %+v", ac)
			}
			// Four colours are not enough.
			effectstest.B06AddMana(me, "W", "U", "B", "R")
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "bringer-wubrg"}); err == nil {
				t.Fatal("the alternative cost was accepted without {G}")
			}
			effectstest.B06AddMana(me, "G")
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "bringer-wubrg"}); err != nil {
				t.Fatalf("five-colour cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			card, ok := effectstest.BattlefieldCard(g, id)
			if !ok {
				t.Fatal("the Bringer did not enter")
			}
			if !game.HasKeyword(&card, "trample") {
				t.Error("a Bringer has trample")
			}
		})
	}
}

// Bringer of the Blue Dawn: the draw is a "may" asked as the trigger
// resolves.
func TestBringerOfTheBlueDawnMayDrawTwoAtUpkeep(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		effectstest.PushCatalogPermanent(g, opp.ID, "Bringer of the Blue Dawn", "Creature — Bringer", effectstest.BringerBlueDawnOracle, false)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
		before := opp.Hand.Size()
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
		if ask == nil {
			t.Fatal("the upkeep draw should ask")
		}
		if err := g.ResolveConfirm(ask.ID, opp.ID, accept); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		want := before
		if accept {
			want += 2
		}
		if got := opp.Hand.Size(); got != want {
			t.Errorf("accept=%v: hand %d, want %d", accept, got, want)
		}
	}
}

// Bringer of the Black Dawn: pay 2 life, then search and put the card
// on top. Declining costs nothing and finds nothing.
func TestBringerOfTheBlackDawnPaysLifeToTutorToTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Bringer of the Black Dawn", "Creature — Bringer", effectstest.BringerBlackDawnOracle, false)
	want := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: want, Name: "Sol Ring", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	for i := 0; i < 3; i++ {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	life := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("the upkeep tutor should ask")
	}
	if ask.LifeCost != 2 {
		t.Errorf("the prompt declares a 2 life cost, got %d", ask.LifeCost)
	}
	if err := g.ResolveConfirm(ask.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	c := effectstest.SearchChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no search prompt after paying")
	}
	if err := g.ResolveSearchLibrary(c.ID, opp.ID, []uuid.UUID{want}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if got := effectstest.LibraryIndex(opp, want); got != len(opp.Library.Cards)-1 {
		t.Errorf("the tutored card is at library index %d of %d, want the top", got, len(opp.Library.Cards))
	}
}

func TestBringerOfTheBlackDawnDecliningCostsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Bringer of the Black Dawn", "Creature — Bringer", effectstest.BringerBlackDawnOracle, false)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	life := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("the upkeep tutor should ask")
	}
	if err := g.ResolveConfirm(ask.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if opp.Life != life || effectstest.SearchChoiceFor(g, opp.ID) != nil {
		t.Errorf("declined: life %d (was %d), search prompt %v", opp.Life, life, effectstest.SearchChoiceFor(g, opp.ID) != nil)
	}
}

// A player who cannot pay (1 life) searches for nothing.
func TestBringerOfTheBlackDawnNeedsTheLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Bringer of the Black Dawn", "Creature — Bringer", effectstest.BringerBlackDawnOracle, false)
	opp.Life = 1
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm); ask != nil {
		if err := g.ResolveConfirm(ask.ID, opp.ID, true); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
	}
	if opp.Life != 1 || effectstest.SearchChoiceFor(g, opp.ID) != nil {
		t.Errorf("at 1 life nothing may be paid: life %d, search prompt %v", opp.Life, effectstest.SearchChoiceFor(g, opp.ID) != nil)
	}
}

// --- Damia -------------------------------------------------------------

func TestDamiaSkipsTheDrawStepAndDrawsUpToSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Damia, Sage of Stone", "Legendary Creature — Gorgon Wizard", effectstest.DamiaOracle, false)
	g.WithWriteLock(func() {
		for opp.Hand.Size() > 3 {
			c, _ := opp.Hand.PopTop()
			opp.Library.PushTop(c)
		}
	})
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != 7 {
		t.Fatalf("hand after the upkeep trigger = %d, want 7", got)
	}
	// The draw step is skipped: advancing to the main phase draws nothing.
	effectstest.ToMain(t, g)
	if got := opp.Hand.Size(); got != 7 {
		t.Errorf("hand after the (skipped) draw step = %d, want 7", got)
	}
}

func TestDamiaDoesNothingAtSevenOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Damia, Sage of Stone", "Legendary Creature — Gorgon Wizard", effectstest.DamiaOracle, false)
	for opp.Hand.Size() < 7 {
		effectstest.PutInHand(opp, game.Card{Name: "Filler", TypeLine: "Land"})
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if len(g.PendingTriggers) != 0 || g.Stack.Size() != 0 {
		t.Errorf("with seven cards in hand the ability must not trigger (stack %d, pending %d)",
			g.Stack.Size(), len(g.PendingTriggers))
	}
}

// --- Conflux -----------------------------------------------------------

func TestConfluxFindsOneCardOfEachColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	want := map[string]uuid.UUID{}
	for _, c := range []struct{ name, cost string }{
		{"White", "{W}"}, {"Blue", "{U}"}, {"Black", "{B}"}, {"Red", "{R}"}, {"Green", "{G}"},
	} {
		id := uuid.New()
		want[c.name] = id
		me.Library.PushTop(game.Card{InstanceID: id, Name: c.name + " Card", TypeLine: "Creature — Bear",
			ManaCost: c.cost, Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastCatalogSpell(t, g, "Conflux", "Sorcery", effectstest.ConfluxOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 5; i++ {
		c := effectstest.SearchChoiceFor(g, me.ID)
		if c == nil {
			break
		}
		if len(c.SearchCards) != 1 {
			t.Fatalf("search %d offers %d cards, want exactly the one of that colour", i, len(c.SearchCards))
		}
		if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{c.SearchCards[0]}); err != nil {
			t.Fatalf("ResolveSearchLibrary: %v", err)
		}
	}
	for name, id := range want {
		if !me.Hand.Contains(id) {
			t.Errorf("the %s card is not in hand", name)
		}
	}
	if got := me.Hand.Size(); got != hand+5 {
		t.Errorf("hand grew by %d, want 5", got-hand)
	}
}

// A gold card found for one colour is gone for the next: it cannot
// stand for two.
func TestConfluxOneGoldCardCannotBeTwoColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gold := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: gold, Name: "Gold Card", TypeLine: "Creature — Bear",
		ManaCost: "{W}{U}", Owner: me.ID, Controller: me.ID})
	blue := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: blue, Name: "Blue Card", TypeLine: "Creature — Bear",
		ManaCost: "{U}", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Conflux", "Sorcery", effectstest.ConfluxOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	// Each search has exactly one candidate, so it is taken without a
	// prompt. White takes the gold card (the first library card in
	// order is the blue one, which is not white), then blue must take
	// the blue card, because the gold one is already in hand.
	if !me.Hand.Contains(gold) {
		t.Error("the gold card is white, so the white search takes it")
	}
	if !me.Hand.Contains(blue) {
		t.Error("the blue search must find the blue card: the gold one cannot count twice")
	}
}

// With only the gold card to find, Conflux takes it ONCE.
func TestConfluxTakesALoneGoldCardOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gold := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: gold, Name: "Gold Card", TypeLine: "Creature — Bear",
		ManaCost: "{W}{U}", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Conflux", "Sorcery", effectstest.ConfluxOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if got := me.Hand.Size(); got != hand+1 || !me.Hand.Contains(gold) {
		t.Errorf("hand grew by %d (gold in hand %v), want exactly the one gold card", got-hand, me.Hand.Contains(gold))
	}
}

// --- Sheoldred ---------------------------------------------------------

func TestSheoldredReturnsACreatureAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sheoldred, Whispering One", "Legendary Creature — Phyrexian Praetor", effectstest.SheoldredWhisperingOracle, false)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityUntilChoice(t, g)
	// The target is the only creature card, so it is picked for us or
	// asked once.
	if c := effectstest.LatestChoiceOfKind(g, game.PendingChoicePickTarget); c != nil {
		effectstest.AnswerFirstPickTarget(t, g)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, dead); !ok {
		t.Error("Sheoldred should return the creature card from the graveyard to the battlefield")
	}
}

func TestSheoldredMakesAnOpponentSacrificeAtTheirUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sheoldred, Whispering One", "Legendary Creature — Phyrexian Praetor", effectstest.SheoldredWhisperingOracle, false)
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", false)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the opponent should be asked to sacrifice a creature at their upkeep")
	}
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("Sheoldred's controller is not asked")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, bear)
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("the chosen creature should have been sacrificed")
	}
}

func TestSheoldredDoesNotTouchYourOwnUpkeepCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Sheoldred, Whispering One", "Legendary Creature — Phyrexian Praetor", effectstest.SheoldredWhisperingOracle, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the edict is for opponents' upkeeps only")
	}
}
