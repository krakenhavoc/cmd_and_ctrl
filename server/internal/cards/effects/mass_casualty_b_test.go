package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mass_casualty_b_test.go — the Mass Casualty deck request (#3245),
// slice B: Grave Researcher, Breach the Multiverse, Abdel Adrian,
// Beseech the Mirror and Hoarding Broodlord.

const (
	mcGraveResearcherOracle = "6cb8b8c4-0674-4f14-9d89-010969fbb80e"
	mcBreachOracle          = "ebbd5b9c-07ff-4b42-81c2-ba09539f3a42"
	mcAbdelOracle           = "cab092f9-b7ff-43b9-935f-310869a4daf8"
	mcBeseechOracle         = "cbbe535b-12d6-4e6c-8b26-9220f06a1604"
	mcBroodlordOracle       = "18e856b5-aab7-4802-8cec-1d44dd95d9f7"
)

func TestMassCasualtySliceBIsRegistered(t *testing.T) {
	for oracle, want := range map[string]Completeness{
		mcGraveResearcherOracle:        CompletenessFull,
		mcGraveResearcherOracle + "#1": CompletenessFull,
		mcBreachOracle:                 CompletenessFull,
		mcAbdelOracle:                  CompletenessFull,
		mcBeseechOracle:                CompletenessFull,
		mcBroodlordOracle:              CompletenessCaveats,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != want {
			t.Errorf("%s: Completeness = %v, want %v", spec.Name, spec.Completeness, want)
		}
	}
}

func mcGraveCreatures(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		pushGraveyardCreature(g, p.ID, "Dead Bear", "{1}{G}")
	}
}

// mcPushResearcher casts a real preparation card and resolves it, so
// the permanent carries its Reanimate face.
func mcPushResearcher(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	card := preparationCard(me.ID, mcGraveResearcherOracle, nil,
		game.Face{Name: "Grave Researcher", TypeLine: "Creature — Troll Warlock", ManaCost: "{2}{B}", Colors: []string{"B"}, Power: 3, Toughness: 3},
		game.Face{Name: "Reanimate", TypeLine: "Sorcery", ManaCost: "{B}", Colors: []string{"B"}})
	advanceToMain(t, g)
	castFromHandAndResolve(t, g, me, card)
	return card.InstanceID
}

func mcRunResearcherUpkeep(t *testing.T, g *game.Game, me *game.Player, toGraveyard bool) {
	t.Helper()
	advanceToUpkeepOfSeat(t, g, 0)
	for i := 0; i < 8 && surveilChoiceFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt at the upkeep")
	}
	top, err := me.Library.Top()
	if err != nil {
		t.Fatal(err)
	}
	var gy, keep []uuid.UUID
	if toGraveyard {
		gy = []uuid.UUID{top.InstanceID}
	} else {
		keep = []uuid.UUID{top.InstanceID}
	}
	if err := g.ResolveSurveil(c.ID, me.ID, gy, keep); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
}

func TestGraveResearcherBecomesPreparedWithThreeCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	researcher := mcPushResearcher(t, g, me)
	mcGraveCreatures(t, g, me, 3)
	pushLibraryCardForTest(me, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	mcRunResearcherUpkeep(t, g, me, false)
	var prepared bool
	g.ReadSnapshot(func() { prepared = g.IsPreparedForEffect(researcher) })
	if !prepared {
		t.Error("three creature cards in the graveyard did not prepare it")
	}
}

func TestGraveResearcherCountsTheCreatureItSurveilled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	researcher := mcPushResearcher(t, g, me)
	mcGraveCreatures(t, g, me, 2)
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Library Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	})
	mcRunResearcherUpkeep(t, g, me, true)
	var prepared bool
	g.ReadSnapshot(func() { prepared = g.IsPreparedForEffect(researcher) })
	if !prepared {
		t.Error("the surveilled creature card was not counted toward the three")
	}
}

func TestGraveResearcherStaysUnpreparedWithTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	researcher := mcPushResearcher(t, g, me)
	mcGraveCreatures(t, g, me, 2)
	pushLibraryCardForTest(me, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	mcRunResearcherUpkeep(t, g, me, false)
	var prepared bool
	g.ReadSnapshot(func() { prepared = g.IsPreparedForEffect(researcher) })
	if prepared {
		t.Error("two creature cards prepared it")
	}
}

func TestBreachTheMultiverseTakesOneFromEachGraveyardAndMakesPhyrexians(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	picks := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		picks[p.ID] = pushGraveyardCreature(g, p.ID, "Grave Bear", "{1}{G}")
	}
	castCatalogSpell(t, g, "Breach the Multiverse", "Sorcery", mcBreachOracle, nil)
	passPriorityAroundTable(t, g)
	for i := 0; i < len(g.Seats); i++ {
		c := chooseCardsChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("pick %d: no choose-cards prompt for the caster", i)
		}
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{picks[c.FromPlayer]}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	for owner, id := range picks {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !g.Battlefield.Contains(id) {
			t.Errorf("the pick from %v is not on the battlefield", owner)
			continue
		}
		if c.Controller != me.ID {
			t.Errorf("the pick from %v is controlled by %v, want the caster", owner, c.Controller)
		}
	}
	for _, id := range append([]uuid.UUID{mine}, picks[me.ID]) {
		c, _ := g.LookupCardForEffect(id)
		if !c.HasSubtype("Phyrexian") {
			t.Errorf("%s is not a Phyrexian", c.Name)
		}
	}
	for _, p := range g.Seats {
		if p.Graveyard.Size() < 1 && p.Library.Size() > 0 {
			t.Errorf("seat %v milled nothing", p.ID)
		}
	}
}

func TestAbdelAdrianExilesUntilHeLeavesAndMakesSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	keep := pushVanillaCreature(g, me.ID, "Bear C", 2, 2)
	land := pushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)

	abdel := castCatalogSpell(t, g, "Abdel Adrian, Gorion's Ward", "Legendary Creature — Human Warrior", mcAbdelOracle, nil)
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no choose-cards prompt")
	}
	for _, id := range c.ChooseCards {
		if id == land || id == abdel {
			t.Error("a land or Abdel himself was offered")
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Fatal("the picks were not exiled")
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("an unpicked permanent was exiled")
	}
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("Soldiers = %d, want 2", n)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(abdel) })
	g.RunStateChecksForTest()
	for _, name := range []string{"Bear A", "Rock"} {
		if _, ok := e2BattlefieldNamed(g, name); !ok {
			t.Errorf("%s did not return when Abdel left", name)
		}
	}
}

func mcBeseechTable(t *testing.T, mv string) (g *game.Game, me *game.Player, tutorTarget uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	tutorTarget = uuid.New()
	g.WithWriteLock(func() {
		pushLibraryCardForTest(me, game.Card{InstanceID: tutorTarget, Name: "Wanted Spell", TypeLine: "Sorcery",
			ManaCost: mv, Owner: me.ID, Controller: me.ID})
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	})
	return g, me, tutorTarget
}

func TestBeseechTheMirrorUnbargainedPutsTheCardIntoHand(t *testing.T) {
	g, me, wanted := mcBeseechTable(t, "{2}")
	if _, err := castWithOptionalCosts(t, g, "Beseech the Mirror", "Sorcery", mcBeseechOracle, nil, nil, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, wanted)
	if !me.Hand.Contains(wanted) {
		t.Error("the found card is not in hand")
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMayCast {
			t.Error("an unbargained Beseech offered a free cast")
		}
	}
}

func TestBeseechTheMirrorBargainedOffersAFreeCast(t *testing.T) {
	g, me, wanted := mcBeseechTable(t, "{2}{B}{B}")
	token := pushTokens(g, me.ID, RedGoblinToken(), 1)[0]
	if _, err := castWithOptionalCosts(t, g, "Beseech the Mirror", "Sorcery", mcBeseechOracle, nil, []int{0}, []uuid.UUID{token}); err != nil {
		t.Fatalf("bargained cast: %v", err)
	}
	if g.Battlefield.Contains(token) {
		t.Fatal("the token was not sacrificed to bargain")
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, wanted)
	if !g.Exile.Contains(wanted) {
		t.Fatal("the found card is not in exile awaiting the cast")
	}
	answerMayCastPrompt(t, g, me.ID, false)
	if !me.Hand.Contains(wanted) {
		t.Error("declining the free cast did not put the card into hand")
	}
}

func TestBeseechTheMirrorAcceptedGrantLapsesToHand(t *testing.T) {
	g, me, wanted := mcBeseechTable(t, "{1}{B}")
	token := pushTokens(g, me.ID, RedGoblinToken(), 1)[0]
	if _, err := castWithOptionalCosts(t, g, "Beseech the Mirror", "Sorcery", mcBeseechOracle, nil, []int{0}, []uuid.UUID{token}); err != nil {
		t.Fatalf("bargained cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, wanted)
	answerMayCastPrompt(t, g, me.ID, true)
	var grant *game.CastPermission
	for i := range me.CastPermissions {
		if me.CastPermissions[i].LapseOnPass == game.LapseToHand {
			grant = &me.CastPermissions[i]
		}
	}
	if grant == nil {
		t.Fatal("accepting granted no pass-closed free cast")
	}
	if grant.Cost != "{0}" || grant.MaxSpellManaValue == nil || *grant.MaxSpellManaValue != 4 {
		t.Errorf("grant = cost %q cap %v, want {0} capped at 4", grant.Cost, grant.MaxSpellManaValue)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(wanted) {
		t.Error("passing without casting did not put the card into hand")
	}
}

func TestBeseechTheMirrorBargainedOverFourGoesToHand(t *testing.T) {
	g, me, wanted := mcBeseechTable(t, "{3}{B}{B}")
	token := pushTokens(g, me.ID, RedGoblinToken(), 1)[0]
	if _, err := castWithOptionalCosts(t, g, "Beseech the Mirror", "Sorcery", mcBeseechOracle, nil, []int{0}, []uuid.UUID{token}); err != nil {
		t.Fatalf("bargained cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, wanted)
	if !me.Hand.Contains(wanted) {
		t.Error("a five-mana card was not put into hand")
	}
}

func TestBeseechTheMirrorRefusesACreatureForBargain(t *testing.T) {
	g, me, _ := mcBeseechTable(t, "{2}")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := castWithOptionalCosts(t, g, "Beseech the Mirror", "Sorcery", mcBeseechOracle, nil, []int{0}, []uuid.UUID{bear}); err == nil {
		t.Error("a nontoken creature paid bargain")
	}
}

func TestHoardingBroodlordExilesAFaceDownCardYouMayPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wanted := uuid.New()
	g.WithWriteLock(func() {
		pushLibraryCardForTest(me, game.Card{InstanceID: wanted, Name: "Wanted", TypeLine: "Creature — Bear",
			ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})
	})
	castCatalogSpell(t, g, "Hoarding Broodlord", "Creature — Dragon", mcBroodlordOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, wanted)
	c, ok := g.LookupCardForEffect(wanted)
	if !ok || !g.Exile.Contains(wanted) {
		t.Fatal("the found card is not in exile")
	}
	if !c.FaceDown {
		t.Error("the card was exiled face up")
	}
	granted := false
	for _, perm := range me.CastPermissions {
		if perm.NamesCard(c) {
			granted = true
		}
	}
	if !granted {
		t.Error("no permission to play the exiled card")
	}
}
