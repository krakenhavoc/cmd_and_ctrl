package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMassCasualtySliceBIsRegistered(t *testing.T) {
	for oracle, want := range map[string]effects.Completeness{
		effectstest.McGraveResearcherOracle:        effects.CompletenessFull,
		effectstest.McGraveResearcherOracle + "#1": effects.CompletenessFull,
		effectstest.McBreachOracle:                 effects.CompletenessFull,
		effectstest.McAbdelOracle:                  effects.CompletenessFull,
		effectstest.McBeseechOracle:                effects.CompletenessFull,
		effectstest.McBroodlordOracle:              effects.CompletenessFull,
	} {
		spec, ok := effects.Lookup(oracle)
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
		effectstest.PushGraveyardCreature(g, p.ID, "Dead Bear", "{1}{G}")
	}
}

// mcPushResearcher casts a real preparation card and resolves it, so
// the permanent carries its Reanimate face.
func mcPushResearcher(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	card := effectstest.PreparationCard(me.ID, effectstest.McGraveResearcherOracle, nil,
		game.Face{Name: "Grave Researcher", TypeLine: "Creature — Troll Warlock", ManaCost: "{2}{B}", Colors: []string{"B"}, Power: 3, Toughness: 3},
		game.Face{Name: "Reanimate", TypeLine: "Sorcery", ManaCost: "{B}", Colors: []string{"B"}})
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandAndResolve(t, g, me, card)
	return card.InstanceID
}

func mcRunResearcherUpkeep(t *testing.T, g *game.Game, me *game.Player, toGraveyard bool) {
	t.Helper()
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	for i := 0; i < 8 && effectstest.SurveilChoiceFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	c := effectstest.SurveilChoiceFor(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	researcher := mcPushResearcher(t, g, me)
	mcGraveCreatures(t, g, me, 3)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	mcRunResearcherUpkeep(t, g, me, false)
	var prepared bool
	g.ReadSnapshot(func() { prepared = g.IsPreparedForEffect(researcher) })
	if !prepared {
		t.Error("three creature cards in the graveyard did not prepare it")
	}
}

func TestGraveResearcherCountsTheCreatureItSurveilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	researcher := mcPushResearcher(t, g, me)
	mcGraveCreatures(t, g, me, 2)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	mcRunResearcherUpkeep(t, g, me, false)
	var prepared bool
	g.ReadSnapshot(func() { prepared = g.IsPreparedForEffect(researcher) })
	if prepared {
		t.Error("two creature cards prepared it")
	}
}

func TestBreachTheMultiverseTakesOneFromEachGraveyardAndMakesPhyrexians(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	picks := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		picks[p.ID] = effectstest.PushGraveyardCreature(g, p.ID, "Grave Bear", "{1}{G}")
	}
	effectstest.CastCatalogSpell(t, g, "Breach the Multiverse", "Sorcery", effectstest.McBreachOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < len(g.Seats); i++ {
		c := effectstest.ChooseCardsChoiceFor(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	keep := effectstest.PushVanillaCreature(g, me.ID, "Bear C", 2, 2)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)

	abdel := effectstest.CastCatalogSpell(t, g, "Abdel Adrian, Gorion's Ward", "Legendary Creature — Human Warrior", effectstest.McAbdelOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
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
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("Soldiers = %d, want 2", n)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(abdel) })
	g.RunStateChecksForTest()
	for _, name := range []string{"Bear A", "Rock"} {
		if _, ok := effectstest.E2BattlefieldNamed(g, name); !ok {
			t.Errorf("%s did not return when Abdel left", name)
		}
	}
}

func mcBeseechTable(t *testing.T, mv string) (g *game.Game, me *game.Player, tutorTarget uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	tutorTarget = uuid.New()
	g.WithWriteLock(func() {
		effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: tutorTarget, Name: "Wanted Spell", TypeLine: "Sorcery",
			ManaCost: mv, Owner: me.ID, Controller: me.ID})
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	})
	return g, me, tutorTarget
}

func TestBeseechTheMirrorUnbargainedPutsTheCardIntoHand(t *testing.T) {
	g, me, wanted := mcBeseechTable(t, "{2}")
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Beseech the Mirror", "Sorcery", effectstest.McBeseechOracle, nil, nil, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, wanted)
	if !me.Hand.Contains(wanted) {
		t.Error("the found card is not in hand")
	}
	if effectstest.KeywordCastPromptFor(g, me.ID) != nil {
		t.Error("an unbargained Beseech offered a free cast")
	}
}

func TestBeseechTheMirrorBargainedOffersAFreeCast(t *testing.T) {
	g, me, wanted := mcBeseechTable(t, "{2}{B}{B}")
	token := effectstest.PushTokens(g, me.ID, effects.RedGoblinToken(), 1)[0]
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Beseech the Mirror", "Sorcery", effectstest.McBeseechOracle, nil, []int{0}, []uuid.UUID{token}); err != nil {
		t.Fatalf("bargained cast: %v", err)
	}
	if g.Battlefield.Contains(token) {
		t.Fatal("the token was not sacrificed to bargain")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, wanted)
	if !g.Exile.Contains(wanted) {
		t.Fatal("the found card is not in exile awaiting the cast")
	}
	effectstest.AnswerMayCastPrompt(t, g, me.ID, false)
	if !me.Hand.Contains(wanted) {
		t.Error("declining the free cast did not put the card into hand")
	}
}

// The bargained offer is made inside Beseech's resolution (CR 608.2g,
// #4250): the card carries a {0} grant capped at 4 under a resolution_cast
// prompt, and casting it puts it on the stack before Beseech finishes.
func TestBeseechTheMirrorAcceptedGrantLapsesToHand(t *testing.T) {
	g, me, wanted := mcBeseechTable(t, "{1}{B}")
	token := effectstest.PushTokens(g, me.ID, effects.RedGoblinToken(), 1)[0]
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Beseech the Mirror", "Sorcery", effectstest.McBeseechOracle, nil, []int{0}, []uuid.UUID{token}); err != nil {
		t.Fatalf("bargained cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, wanted)
	var grant *game.CastPermission
	for i := range me.CastPermissions {
		if me.CastPermissions[i].LapseOnPass == game.LapseToHand {
			grant = &me.CastPermissions[i]
		}
	}
	if grant == nil {
		t.Fatal("the offer granted no free cast")
	}
	if grant.Cost != "{0}" || grant.MaxSpellManaValue == nil || *grant.MaxSpellManaValue != 4 {
		t.Errorf("grant = cost %q cap %v, want {0} capped at 4", grant.Cost, grant.MaxSpellManaValue)
	}
	c := effectstest.KeywordCastPromptFor(g, me.ID)
	if c == nil || c.MayCastKeyword != game.MayCastKeywordFreeOrHand || !g.ResolutionPaused() {
		t.Fatalf("prompt = %+v, want a free_or_hand resolution_cast holding Beseech's resolution", c)
	}
	effectstest.AnswerMayCastPrompt(t, g, me.ID, true)
	if !g.Stack.Contains(wanted) || g.ResolutionPaused() {
		t.Error("the free cast is not on the stack, or Beseech is still resolving")
	}
}

// A table paused on the offer is a restore point, and the restored
// decline puts the card into the hand.
func TestBeseechTheMirrorOfferSurvivesARestoreMidPrompt(t *testing.T) {
	g, me, wanted := mcBeseechTable(t, "{1}{B}")
	token := effectstest.PushTokens(g, me.ID, effects.RedGoblinToken(), 1)[0]
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Beseech the Mirror", "Sorcery", effectstest.McBeseechOracle, nil, []int{0}, []uuid.UUID{token}); err != nil {
		t.Fatalf("bargained cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, wanted)
	restored := effectstest.RestoreRoundTrip(t, g, true)
	rme := restored.PlayerByIDForEffect(me.ID)
	if rc := effectstest.KeywordCastPromptFor(restored, rme.ID); rc == nil || effectstest.KeywordCastCard(rc) != wanted {
		t.Fatalf("the restored table lost the offer: %+v", rc)
	}
	effectstest.AnswerMayCastPrompt(t, restored, rme.ID, false)
	if !rme.Hand.Contains(wanted) {
		t.Error("the restored decline did not put the card into hand")
	}
	if restored.ResolutionPaused() {
		t.Error("the restored resolution did not finish")
	}
}

func TestBeseechTheMirrorBargainedOverFourGoesToHand(t *testing.T) {
	g, me, wanted := mcBeseechTable(t, "{3}{B}{B}")
	token := effectstest.PushTokens(g, me.ID, effects.RedGoblinToken(), 1)[0]
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Beseech the Mirror", "Sorcery", effectstest.McBeseechOracle, nil, []int{0}, []uuid.UUID{token}); err != nil {
		t.Fatalf("bargained cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, wanted)
	if !me.Hand.Contains(wanted) {
		t.Error("a five-mana card was not put into hand")
	}
}

func TestBeseechTheMirrorRefusesACreatureForBargain(t *testing.T) {
	g, me, _ := mcBeseechTable(t, "{2}")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Beseech the Mirror", "Sorcery", effectstest.McBeseechOracle, nil, []int{0}, []uuid.UUID{bear}); err == nil {
		t.Error("a nontoken creature paid bargain")
	}
}

func TestHoardingBroodlordExilesAFaceDownCardYouMayPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wanted := uuid.New()
	g.WithWriteLock(func() {
		effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: wanted, Name: "Wanted", TypeLine: "Creature — Bear",
			ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})
	})
	effectstest.CastCatalogSpell(t, g, "Hoarding Broodlord", "Creature — Dragon", effectstest.McBroodlordOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, wanted)
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
