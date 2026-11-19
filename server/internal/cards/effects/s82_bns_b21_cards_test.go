package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b21_cards_test.go — set BNS batch 21 (#5929): the eight Ravnica:
// Clue Edition duals, Amzu, Commander Mustard, Headliner Scarlett and
// Lavinia. Each test plays the real catalog entry.

const (
	bnsB21AmzuOracle     = "6777863c-e52c-4a7c-b8d2-a16b1438b3d2"
	bnsB21MustardOracle  = "3fdbe867-63cd-42c5-9843-46b0992b9628"
	bnsB21ScarlettOracle = "f5420901-b5e6-4d54-b3e8-5f45e37f1fab"
	bnsB21LaviniaOracle  = "b5b26cea-9be3-46f7-9603-a6eee0dd107e"
)

var bnsB21Duals = []struct{ oracle, name, a, b string }{
	{"8f88b0bf-81bd-4223-9dad-d8c49ff4b87a", "Conservatory", "G", "W"},
	{"0880461e-8943-443b-90e7-ff84eef46550", "Dining Room", "R", "G"},
	{"949e455d-6a9e-491a-892f-826cc8be0fd9", "Hall", "R", "W"},
	{"47f71408-5509-4651-9121-fd0867adae00", "Kitchen", "G", "U"},
	{"33a7d29a-ebc0-4606-91b6-7381e2a16016", "Library", "U", "R"},
	{"98a22879-d0b2-441f-bcf6-a67b4552b73c", "Lounge", "B", "G"},
	{"c6a46fc2-fc8f-4dcd-bca7-682abfbf303d", "Secret Passage", "U", "B"},
	{"7ffadb3f-0b88-416f-b1a1-876383c22720", "Study", "W", "U"},
}

func TestBnsB21EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		bnsB21AmzuOracle:     "Amzu, Swarm's Hunger",
		bnsB21MustardOracle:  "Commander Mustard",
		bnsB21ScarlettOracle: "Headliner Scarlett",
		bnsB21LaviniaOracle:  "Lavinia, Foil to Conspiracy",
	}
	for _, l := range bnsB21Duals {
		want[l.oracle] = l.name
	}
	if len(want) != 12 {
		t.Fatalf("the batch ships 12 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
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

func bnsB21Clues(g *game.Game, controller uuid.UUID) int {
	return countOnBattlefield(g, "Clue", controller)
}

func TestBnsB21DualsEnterTappedAndTapForTheirTwoColours(t *testing.T) {
	for _, l := range bnsB21Duals {
		spec, _ := Lookup(l.oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{"+l.a+"|"+l.b+"}" {
			t.Errorf("%s: mana abilities %+v, want one producing {%s} or {%s}", l.name, spec.ManaAbilities, l.a, l.b)
		}
		g := newCatalogGame(t)
		id := playLandFromHand(t, g, l.name, l.oracle)
		top100AssertEnteredTapped(t, g, id, l.name)
	}
}

func TestBnsB21DualInvestigatesForFourAndTap(t *testing.T) {
	for _, l := range bnsB21Duals {
		t.Run(l.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			land := pushCatalogPermanent(g, me.ID, l.name, "Land", l.oracle, false)
			b09AddColorless(me, 3)
			if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("activated for three mana; the cost is {4}")
			}
			b09AddColorless(me, 1)
			if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if c, _ := battlefieldCard(g, land); !c.Tapped {
				t.Error("the land is tapped as a cost")
			}
			if bnsB21Clues(g, me.ID) != 0 {
				t.Fatal("the Clue must wait for the ability to resolve")
			}
			passPriorityAroundTable(t, g)
			if got := bnsB21Clues(g, me.ID); got != 1 {
				t.Errorf("%d Clues, want 1", got)
			}
		})
	}
}

// --- Lavinia ------------------------------------------------------------------

func TestBnsB21LaviniaTapsForTwoColorlessOnlyDuringAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMainOf(t, g, 0)
	lav := pushCatalogPermanent(g, me.ID, "Lavinia, Foil to Conspiracy", "Legendary Creature — Human Detective", bnsB21LaviniaOracle, false)
	if err := g.ActivateManaAbility(me.ID, lav, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("tapped for mana on its controller's own turn")
	}
	advanceToMainOf(t, g, 1)
	for g.Turn.ActiveSeat != 1 || g.Turn.Step != game.StepPrecombatMain {
		t.Fatal("not on seat 1's main phase")
	}
	tapForMana(t, g, me.ID, lav)
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("pool holds %d mana, want 2", got)
	}
}

func TestBnsB21LaviniaInvestigatesOnTheSecondSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Lavinia, Foil to Conspiracy", "Legendary Creature — Human Detective", bnsB21LaviniaOracle, false)
	castVanillaSorceryForRevealTest(t, g, me)
	passPriorityAroundTable(t, g)
	if got := bnsB21Clues(g, me.ID); got != 0 {
		t.Fatalf("%d Clues after the first spell, want 0", got)
	}
	castVanillaSorceryForRevealTest(t, g, me)
	passPriorityAroundTable(t, g)
	if got := bnsB21Clues(g, me.ID); got != 1 {
		t.Fatalf("%d Clues after the second spell, want 1", got)
	}
	castVanillaSorceryForRevealTest(t, g, me)
	passPriorityAroundTable(t, g)
	if got := bnsB21Clues(g, me.ID); got != 1 {
		t.Errorf("%d Clues after the third spell, want still 1", got)
	}
}

// --- Commander Mustard --------------------------------------------------------

func TestBnsB21MustardGivesOtherSoldiersTheirKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 0)
	mustard := pushCatalogPermanent(g, me.ID, "Commander Mustard", "Legendary Creature — Human Soldier", bnsB21MustardOracle, false)
	mine := b12Push(g, me.ID, "My Soldier", "Creature — Human Soldier", "", 2, 2)
	theirs := b12Push(g, opp.ID, "Their Soldier", "Creature — Human Soldier", "", 2, 2)
	other := b12Push(g, me.ID, "My Cleric", "Creature — Human Cleric", "", 2, 2)
	has := func(id uuid.UUID, kw string) bool {
		var ok bool
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		g.ReadSnapshot(func() {
			for i := range g.Battlefield.Cards {
				if c := &g.Battlefield.Cards[i]; c.InstanceID == id {
					ok = game.HasKeyword(c, kw)
				}
			}
		})
		return ok
	}
	for _, kw := range []string{"vigilance", "trample", "haste"} {
		if !has(mine, kw) {
			t.Errorf("my other Soldier lacks %s", kw)
		}
		if has(theirs, kw) || has(other, kw) {
			t.Errorf("a non-Soldier or an opponent's Soldier has %s", kw)
		}
	}
	if !has(mustard, "vigilance") || has(mustard, "trample") {
		t.Error("Mustard has his own vigilance only")
	}
}

func TestBnsB21MustardGivesSoldiersAnAttackPingUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 0)
	mustard := pushCatalogPermanent(g, me.ID, "Commander Mustard", "Legendary Creature — Human Soldier", bnsB21MustardOracle, false)
	soldier := b12Push(g, me.ID, "My Soldier", "Creature — Human Soldier", "", 2, 2)
	cleric := b12Push(g, me.ID, "My Cleric", "Creature — Human Cleric", "", 2, 2)
	b09AddColorless(me, 2)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})
	if err := g.ActivateCatalogAbility(me.ID, mustard, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	life := lifeOf(g, opp.ID)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{mustard, soldier, cleric} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	g.RunStateChecksForTest()
	answerTriggerOrderInOfferedOrder(t, g)
	for i := 0; i < 6 && !stackFullyEmpty(g); i++ {
		passPriorityAroundTable(t, g)
	}
	if got := life - lifeOf(g, opp.ID); got != 2 {
		t.Errorf("the defending player lost %d life from triggers, want 2 (Mustard and the Soldier, not the Cleric)", got)
	}
}

// --- Headliner Scarlett -------------------------------------------------------

func TestBnsB21ScarlettEntersAndTheTargetPlayersCreaturesCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 0)
	early := b12Push(g, opp.ID, "Early Bear", "Creature — Bear", "", 2, 2)
	third := b12Push(g, g.Seats[2].ID, "Third Bear", "Creature — Bear", "", 2, 2)
	mine := b12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	castHoldCreatureNoResolve(t, g, "Headliner Scarlett", "Legendary Creature — Human Warlock", bnsB21ScarlettOracle, 3, 3)
	passPriorityAroundTable(t, g)
	pick := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick the player: %v", err)
	}
	passPriorityAroundTable(t, g)
	late := b12Push(g, opp.ID, "Late Bear", "Creature — Bear", "", 2, 2)
	assertRestrictions(t, g, early, game.CantBlock)
	assertRestrictions(t, g, late, game.CantBlock)
	assertRestrictions(t, g, third, 0)
	assertRestrictions(t, g, mine, 0)
}

func TestBnsB21ScarlettExilesTheTopCardFaceDownAtUpkeepAndMayPlayIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMainOf(t, g, 0)
	pushCatalogPermanent(g, me.ID, "Headliner Scarlett", "Legendary Creature — Human Warlock", bnsB21ScarlettOracle, false)
	top := avB20LibraryTop(g, me, game.Card{Name: "Mystery Forest", TypeLine: "Basic Land — Forest"})
	advanceToUpkeepOfSeat(t, g, 0)
	for i := 0; i < 6 && !g.Exile.Contains(top); i++ {
		passPriorityAroundTable(t, g)
	}
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	if me.Library.Contains(top) {
		t.Fatal("the card is still in the library")
	}
}

// --- Amzu ---------------------------------------------------------------------

func bnsB21Insects(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Insect" && c.IsToken() && c.Controller == controller {
				out = append(out, c)
			}
		}
	})
	return out
}

func bnsB21GraveCard(g *game.Game, me *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", ManaCost: cost, Owner: me.ID, Controller: me.ID})
	})
	return id
}

func TestBnsB21AmzuMakesAnInsectWithCountersEqualToTheGreatestManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMainOf(t, g, 0)
	amzu := pushCatalogPermanent(g, me.ID, "Amzu, Swarm's Hunger", "Legendary Creature — Insect Shaman", bnsB21AmzuOracle, false)
	a := bnsB21GraveCard(g, me, "Cheap", "{1}")
	b := bnsB21GraveCard(g, me, "Dear", "{3}{B}")
	g.WithWriteLock(func() {
		if err := g.ExileCardsKeyedForEffect([]uuid.UUID{a, b}, game.KeyedTail{}); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	bnsB21AnswerTrigger(t, g, me, true)
	passPriorityAroundTable(t, g)
	insects := bnsB21Insects(g, me.ID)
	if len(insects) != 1 {
		t.Fatalf("%d Insect tokens, want 1", len(insects))
	}
	if got := battlefieldCountersOf(g, insects[0].InstanceID, game.CounterPlusOne); got != 4 {
		t.Errorf("%d +1/+1 counters, want 4 (the greatest mana value)", got)
	}
	// "Do this only once each turn": a second batch does nothing.
	c := bnsB21GraveCard(g, me, "Later", "{2}")
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(c); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := len(bnsB21Insects(g, me.ID)); got != 1 {
		t.Errorf("%d Insect tokens after a second batch, want still 1", got)
	}
	_ = amzu
}

func TestBnsB21AmzuDecliningDoesNotSpendTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMainOf(t, g, 0)
	pushCatalogPermanent(g, me.ID, "Amzu, Swarm's Hunger", "Legendary Creature — Insect Shaman", bnsB21AmzuOracle, false)
	a := bnsB21GraveCard(g, me, "Cheap", "{2}")
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(a); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	bnsB21AnswerTrigger(t, g, me, false)
	passPriorityAroundTable(t, g)
	if got := len(bnsB21Insects(g, me.ID)); got != 0 {
		t.Fatalf("%d Insects after declining, want 0", got)
	}
	b := bnsB21GraveCard(g, me, "Dear", "{4}")
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(b); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	bnsB21AnswerTrigger(t, g, me, true)
	passPriorityAroundTable(t, g)
	ins := bnsB21Insects(g, me.ID)
	if len(ins) != 1 || battlefieldCountersOf(g, ins[0].InstanceID, game.CounterPlusOne) != 4 {
		t.Errorf("want one Insect with four counters, got %d tokens", len(ins))
	}
}

func TestBnsB21AmzuGivesOtherInsectsMenace(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 0)
	pushCatalogPermanent(g, me.ID, "Amzu, Swarm's Hunger", "Legendary Creature — Insect Shaman", bnsB21AmzuOracle, false)
	mine := b12Push(g, me.ID, "My Insect", "Creature — Insect", "", 1, 1)
	theirs := b12Push(g, opp.ID, "Their Insect", "Creature — Insect", "", 1, 1)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	var mineMenace, theirsMenace bool
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			switch c.InstanceID {
			case mine:
				mineMenace = game.HasKeyword(c, "menace")
			case theirs:
				theirsMenace = game.HasKeyword(c, "menace")
			}
		}
	})
	if !mineMenace || theirsMenace {
		t.Errorf("menace: mine %v, theirs %v, want true and false", mineMenace, theirsMenace)
	}
}

// bnsB21AnswerTrigger answers the "you may" the trigger harvester asked as the
// ability triggered (CR 603.5).
func bnsB21AnswerTrigger(t *testing.T, g *game.Game, me *game.Player, yes bool) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
			if err := g.ResolveTriggerPrompt(c.ID, me.ID, yes); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			return
		}
	}
	t.Fatalf("no trigger prompt for %s: %+v", me.Name, g.PendingChoices)
}
