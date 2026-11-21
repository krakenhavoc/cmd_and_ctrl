package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b24_cards_test.go — Secrets of Strixhaven, batch 24 (#4456). Each test
// plays the real catalog entry through the engine and asserts what the
// printed text says.

const (
	socB24EndOfTheHuntOracle   = "74b8182e-54be-444a-97b1-84fc5fd3c290"
	socB24EnnisOracle          = "97e1fcdf-04e8-40cf-b04a-8505d9b7cecc"
	socB24ScientistOracle      = "f30aa897-7027-47e7-ae58-73cbc448c1f0"
	socB24ScholarOracle        = "dd757328-0593-4fc7-a099-e82f5c0677ff"
	socB24StudentOracle        = "140cb5cf-8aa2-457c-9d65-5acefea00195"
	socB24TidecallerOracle     = "d32c9439-4c97-412b-83cc-6d9acb48be3c"
	socB24FiredancerOracle     = "f1812d69-c016-4bc3-b899-ed47f5e893e5"
	socB24FieldsOfStrifeOracle = "a32e08fa-bea4-4ba9-a126-9bf0a91f67e2"
	socB24FixWhatsBrokenOracle = "7f8f46c0-889a-4c9a-9bfa-2b43db608d55"
	socB24FlashbackOracle      = "02070488-9203-4304-9392-a111d20218c5"
)

func TestSocB24EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB24EndOfTheHuntOracle:   "End of the Hunt",
		socB24EnnisOracle:          "Ennis, Debate Moderator",
		socB24ScientistOracle:      "Environmental Scientist",
		socB24ScholarOracle:        "Essenceknit Scholar",
		socB24StudentOracle:        "Eternal Student",
		socB24TidecallerOracle:     "Exhibition Tidecaller",
		socB24FiredancerOracle:     "Expressive Firedancer",
		socB24FieldsOfStrifeOracle: "Fields of Strife",
		socB24FixWhatsBrokenOracle: "Fix What's Broken",
		socB24FlashbackOracle:      "Flashback",
	}
	for oracle, name := range want {
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

// socB24Graveyard puts a card with a mana cost into a graveyard.
func socB24Graveyard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// --- End of the Hunt -----------------------------------------------------------

func socB24EndOfTheHuntCard() game.Card {
	return game.Card{Name: "End of the Hunt", TypeLine: "Sorcery", OracleID: socB24EndOfTheHuntOracle, ManaCost: "{1}{B}"}
}

func TestSocB24EndOfTheHuntLetsTheOpponentExileAmongTheGreatestManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.SocB24Perm(g, opp.ID, "Big Beast", "Creature — Beast", "{3}{G}{G}", 5, 5)
	walker := effectstest.SocB24Perm(g, opp.ID, "Tall Walker", "Legendary Planeswalker — Test", "{3}{U}{U}", 0, 0)
	effectstest.AddCountersForTest(g, walker, game.CounterLoyalty, 4)
	small := effectstest.SocB24Perm(g, opp.ID, "Small Bear", "Creature — Bear", "{1}{G}", 2, 2)
	rock := effectstest.SocB24Perm(g, opp.ID, "Big Rock", "Artifact", "{6}", 0, 0)
	mine := effectstest.SocB24Perm(g, me.ID, "Mine Giant", "Creature — Giant", "{7}", 7, 7)
	if _, err := effectstest.Msc17Cast(t, g, me, socB24EndOfTheHuntCard(), game.CastSpellParams{Targets: []game.TargetRef{effectstest.Msc17TargetPlayer(me.ID)}}); err == nil {
		t.Fatal("End of the Hunt targeted its own caster")
	}
	if _, err := effectstest.Msc17Cast(t, g, me, socB24EndOfTheHuntCard(), game.CastSpellParams{Targets: []game.TargetRef{effectstest.Msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster was asked to choose: the opponent chooses")
	}
	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the opponent was not asked")
	}
	if !effectstest.HasID(pick.ChooseCards, big) || !effectstest.HasID(pick.ChooseCards, walker) || len(pick.ChooseCards) != 2 {
		t.Fatalf("offered %v: only the creature and planeswalker tied at mana value 5 are legal (not the artifact, the smaller creature or the caster's)", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, walker)
	if !g.Exile.Contains(walker) {
		t.Error("the chosen planeswalker is not exiled")
	}
	for name, id := range map[string]uuid.UUID{"the other tied creature": big, "the smaller creature": small, "the artifact": rock, "the caster's creature": mine} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s left the battlefield", name)
		}
	}
}

func TestSocB24EndOfTheHuntExilesTheOnlyGreatestAndSparesOnlyLesserBoards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	king := effectstest.SocB24Perm(g, opp.ID, "King", "Creature — Giant", "{4}{R}", 5, 5)
	pawn := effectstest.SocB24Perm(g, opp.ID, "Pawn", "Creature — Soldier", "{1}", 1, 1)
	if _, err := effectstest.Msc17Cast(t, g, me, socB24EndOfTheHuntCard(), game.CastSpellParams{Targets: []game.TargetRef{effectstest.Msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if pick == nil || len(pick.ChooseCards) != 1 || !effectstest.HasID(pick.ChooseCards, king) {
		t.Fatalf("the pick %+v should offer exactly the one greatest permanent", pick)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, king)
	if !g.Exile.Contains(king) || g.Battlefield.Contains(king) {
		t.Error("the greatest creature is not exiled")
	}
	if !g.Battlefield.Contains(pawn) {
		t.Error("the lesser creature left")
	}
}

func TestSocB24EndOfTheHuntAsksNothingOfAnOpponentWithNoCreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.SocB24Perm(g, opp.ID, "Big Rock", "Artifact", "{6}", 0, 0)
	if _, err := effectstest.Msc17Cast(t, g, me, socB24EndOfTheHuntCard(), game.CastSpellParams{Targets: []game.TargetRef{effectstest.Msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) != nil || effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a choice was asked with nothing to exile")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a noncreature, nonplaneswalker permanent was exiled")
	}
}

// --- Ennis, Debate Moderator ---------------------------------------------------

func TestSocB24EnnisBlinksAnotherCreatureAndGrowsAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Ennis Bear", 2, 2)
	ennis := effectstest.SbCastPicking(t, g, "Ennis, Debate Moderator", socB24EnnisOracle, bear)
	if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) {
		t.Fatal("the targeted creature was not exiled")
	}
	if n := effectstest.CounterCount(g, ennis, game.CounterPlusOne); n != 0 {
		t.Fatalf("Ennis has %d counters before the end step", n)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.RingSettle(t, g)
	if n := effectstest.SocB24Named(g, "Ennis Bear"); n != 1 {
		t.Fatalf("%d Ennis Bear on the battlefield after the end step, want 1 (returned)", n)
	}
	if n := effectstest.CounterCount(g, ennis, game.CounterPlusOne); n != 1 {
		t.Errorf("Ennis has %d +1/+1 counters, want 1: a card was put into exile this turn", n)
	}
}

func TestSocB24EnnisCantTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Ennis Bear", 2, 2)
	id := effectstest.CastCatalogSpell(t, g, "Ennis, Debate Moderator", "Legendary Creature — Human Cleric", socB24EnnisOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger asked for no target with another creature to blink")
	}
	if effectstest.HasID(pick.PickTargetCards, id) || !effectstest.HasID(pick.PickTargetCards, bear) {
		t.Errorf("targets offered %v: want the other creature and not Ennis", pick.PickTargetCards)
	}
}

func TestSocB24EnnisDoesNothingWhenNoCardWasExiledThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ennis := effectstest.B12Push(g, me.ID, "Ennis, Debate Moderator", "Legendary Creature — Human Cleric", socB24EnnisOracle, 1, 1)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.RingSettle(t, g)
	if n := effectstest.CounterCount(g, ennis, game.CounterPlusOne); n != 0 {
		t.Errorf("Ennis has %d counters with nothing exiled this turn, want 0", n)
	}
}

func TestSocB24EnnisCountsAnyCardPutIntoExileByAnyone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ennis := effectstest.B12Push(g, me.ID, "Ennis, Debate Moderator", "Legendary Creature — Human Cleric", socB24EnnisOracle, 1, 1)
	dead := effectstest.SocB11GraveyardCreature(opp, "Their Dead")
	g.WithWriteLock(func() {
		if err := g.ExileCardsKeyedForEffect([]uuid.UUID{dead}, game.KeyedTail{}); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.RingSettle(t, g)
	if n := effectstest.CounterCount(g, ennis, game.CounterPlusOne); n != 1 {
		t.Errorf("Ennis has %d counters after an opponent's graveyard card was exiled, want 1", n)
	}
}

func TestSocB24EnnisIgnoresAnExiledToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ennis := effectstest.B12Push(g, me.ID, "Ennis, Debate Moderator", "Legendary Creature — Human Cleric", socB24EnnisOracle, 1, 1)
	token := effectstest.B12Creature(g, me.ID, "Soldier", "Token Creature — Soldier", 1, 1)
	g.WithWriteLock(func() {
		if err := g.ExileCardsKeyedForEffect([]uuid.UUID{token}, game.KeyedTail{}); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.RingSettle(t, g)
	if n := effectstest.CounterCount(g, ennis, game.CounterPlusOne); n != 0 {
		t.Errorf("Ennis has %d counters after only a token was exiled, want 0: a token is not a card", n)
	}
}

func TestSocB24EnnisSeesACardExiledFromAGraveyardAsACost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ennis := effectstest.B12Push(g, me.ID, "Ennis, Debate Moderator", "Legendary Creature — Human Cleric", socB24EnnisOracle, 1, 1)
	id := effectstest.PushCatalogGraveyardCard(me, "Eternal Student", "Creature — Zombie Warlock", socB24StudentOracle, 4, 2)
	effectstest.FloatForTest(g, me, "CB")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Eternal Student: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.RingSettle(t, g)
	if n := effectstest.CounterCount(g, ennis, game.CounterPlusOne); n != 1 {
		t.Errorf("Ennis has %d counters after a card was exiled as a cost, want 1", n)
	}
}

// --- Environmental Scientist -----------------------------------------------------

func TestSocB24ScientistFetchesABasicLandToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Taiga", "Land — Mountain Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.CastCatalogSpell(t, g, "Environmental Scientist", "Creature — Human Druid", socB24ScientistOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger asked for no search")
	}
	if effectstest.SearchOptionNamed(g, c, "Taiga") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a nonbasic land or a nonland card is offered")
	}
	handBefore := me.Hand.Size()
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	if me.Hand.Size() != handBefore+1 || !effectstest.B02bHandHasNamed(me, "Forest") {
		t.Errorf("hand %d -> %d, Forest in hand %v: want the basic land in hand", handBefore, me.Hand.Size(), effectstest.B02bHandHasNamed(me, "Forest"))
	}
}

func TestSocB24ScientistMayFailToFind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))
	effectstest.CastCatalogSpell(t, g, "Environmental Scientist", "Creature — Human Druid", socB24ScientistOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	handBefore := me.Hand.Size()
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d -> %d after declining to find", handBefore, me.Hand.Size())
	}
}

// --- Essenceknit Scholar -----------------------------------------------------------

func TestSocB24ScholarMakesAPestThatGainsLifeWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Essenceknit Scholar", "Creature — Dryad Warlock", socB24ScholarOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pestID := effectstest.FindBattlefieldByName(g, "Pest")
	if pestID == uuid.Nil {
		t.Fatal("no Pest was created")
	}
	pest := effectstest.S58p6View(t, g, pestID)
	if !pest.IsToken() || pest.CurrentPower() != 1 || pest.CurrentToughness() != 1 || len(pest.Colors) != 2 {
		t.Fatalf("the Pest is token %v %d/%d colors %v, want a 1/1 black and green token", pest.IsToken(), pest.CurrentPower(), pest.CurrentToughness(), pest.Colors)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == pestID {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, pestID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d when the Pest attacked, want +1", life, me.Life)
	}
}

func TestSocB24ScholarsPestGainsNothingWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Essenceknit Scholar", "Creature — Dryad Warlock", socB24ScholarOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pestID := effectstest.FindBattlefieldByName(g, "Pest")
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pestID) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d -> %d when the Pest died: this Pest gains life when it attacks, not dies", life, me.Life)
	}
}

func TestSocB24ScholarDrawsAtTheEndStepOnlyIfACreatureDiedUnderYourControl(t *testing.T) {
	for _, tc := range []struct {
		name  string
		owner int
		draws bool
	}{
		{"nothing died", -1, false},
		{"my creature died", 0, true},
		{"an opponent's creature died", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.B12Push(g, me.ID, "Essenceknit Scholar", "Creature — Dryad Warlock", socB24ScholarOracle, 3, 1)
			if tc.owner >= 0 {
				victim := effectstest.B12Creature(g, g.Seats[tc.owner].ID, "Victim", "Creature — Bear", 2, 2)
				g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
				effectstest.RingSettle(t, g)
			}
			effectstest.AdvanceTo(t, g, game.StepEnd)
			hand := me.Hand.Size()
			effectstest.RingSettle(t, g)
			want := hand
			if tc.draws {
				want++
			}
			if me.Hand.Size() != want {
				t.Errorf("hand %d -> %d at the end step, want %d", hand, me.Hand.Size(), want)
			}
		})
	}
}

// --- Eternal Student -----------------------------------------------------------------

func TestSocB24EternalStudentMakesTwoFlyingInklingsFromTheGraveyardAtInstantSpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain {
		t.Fatalf("fixture starts in a main phase (%s): the test wants an instant-speed window", g.Turn.Step)
	}
	id := effectstest.PushCatalogGraveyardCard(me, "Eternal Student", "Creature — Zombie Warlock", socB24StudentOracle, 4, 2)
	effectstest.FloatForTest(g, me, "RR")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with {R}{R}: it costs {1}{B}")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("the refused activation still moved the card")
	}
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{1}{B}, exile this card from your graveyard (outside a main phase): %v", err)
	}
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("the card is not exiled as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.SocB24Named(g, "Inkling"); n != 2 {
		t.Fatalf("%d Inklings, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Inkling" {
			continue
		}
		v := effectstest.S58p6View(t, g, c.InstanceID)
		if !v.IsToken() || v.CurrentPower() != 1 || v.CurrentToughness() != 1 || len(v.Colors) != 2 || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
			t.Errorf("an Inkling is token %v %d/%d colors %v, want a 1/1 white and black flying token", v.IsToken(), v.CurrentPower(), v.CurrentToughness(), v.Colors)
		}
	}
}

func TestSocB24EternalStudentHasNoAbilityOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Eternal Student", "Creature — Zombie Warlock", socB24StudentOracle, 4, 2)
	effectstest.FloatForTest(g, me, "CB")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the graveyard ability was activated from the battlefield")
	}
}

// --- Exhibition Tidecaller ---------------------------------------------------------

// socB24CastOpusSpell casts a plain sorcery for `cost`, paid with exactly
// `pool`, answers the Opus trigger's player target and settles the table.
func socB24CastOpusSpell(t *testing.T, g *game.Game, typeLine, cost, pool string, target uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.SocB17Hand(me, "Test Spell", typeLine, cost, effectstest.SocB17PlainSorceryOracle)
	effectstest.FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", cost, err)
	}
	if target != uuid.Nil {
		if effectstest.LatestPickTarget(g, me.ID) == nil {
			t.Fatal("the Opus trigger asked for no target")
		}
		effectstest.PickPlayer(t, g, me.ID, target)
	}
	for effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID) {
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB24TidecallerMillsThreeAndTenOnFiveOrMoreMana(t *testing.T) {
	for _, tc := range []struct {
		name, cost, pool string
		mills            int
	}{
		{"one mana", "{R}", "R", 3},
		{"four mana", "{3}{R}", "RRRR", 3},
		{"five mana", "{4}{R}", "RRRRR", 10},
		{"six mana", "{5}{R}", "RRRRRR", 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.PushCatalogPermanent(g, me.ID, "Exhibition Tidecaller", "Creature — Djinn Wizard", socB24TidecallerOracle, false)
			before, grave := opp.Library.Size(), opp.Graveyard.Size()
			socB24CastOpusSpell(t, g, "Sorcery", tc.cost, tc.pool, opp.ID)
			if got := before - opp.Library.Size(); got != tc.mills {
				t.Errorf("the library lost %d cards, want %d", got, tc.mills)
			}
			if got := opp.Graveyard.Size() - grave; got != tc.mills {
				t.Errorf("the graveyard gained %d cards, want %d", got, tc.mills)
			}
		})
	}
}

func TestSocB24TidecallerCanMillItsOwnController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Exhibition Tidecaller", "Creature — Djinn Wizard", socB24TidecallerOracle, false)
	before := me.Library.Size()
	socB24CastOpusSpell(t, g, "Instant", "{1}{R}", "RR", me.ID)
	if got := before - me.Library.Size(); got != 3 {
		t.Errorf("my library lost %d cards, want 3", got)
	}
}

func TestSocB24TidecallerMillsWhatIsLeftInAShortLibraryAndNobodyLoses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Exhibition Tidecaller", "Creature — Djinn Wizard", socB24TidecallerOracle, false)
	effectstest.SeedSearchLibrary(opp, game.Card{Name: "A"}, game.Card{Name: "B"}, game.Card{Name: "C"}, game.Card{Name: "D"})
	grave := opp.Graveyard.Size()
	socB24CastOpusSpell(t, g, "Sorcery", "{4}{R}", "RRRRR", opp.ID)
	if opp.Library.Size() != 0 || opp.Graveyard.Size() != grave+4 {
		t.Errorf("library %d, graveyard +%d: want everything milled and nothing more", opp.Library.Size(), opp.Graveyard.Size()-grave)
	}
}

func TestSocB24TidecallerIgnoresCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Exhibition Tidecaller", "Creature — Djinn Wizard", socB24TidecallerOracle, false)
	before := opp.Library.Size()
	socB24CastOpusSpell(t, g, "Creature — Bear", "{4}{R}", "RRRRR", uuid.Nil)
	if opp.Library.Size() != before || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a creature spell triggered Opus")
	}
}

// --- Expressive Firedancer -----------------------------------------------------------

func socB24Firedancer(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.SocB17Perm(g, owner, "Expressive Firedancer", "Creature — Human Sorcerer", socB24FiredancerOracle, "{1}{R}", 2, 2)
}

func TestSocB24FiredancerGetsPlusOneOnAnyInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dancer := socB24Firedancer(g, me.ID)
	effectstest.SocB17CastSpellFor(t, g, "Instant", "{3}{R}", "RRRR")
	c := effectstest.S58p6View(t, g, dancer)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Fatalf("Firedancer is %d/%d after a four-mana spell, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if effectstest.HasEffectiveKeyword(t, g, dancer, "double strike") {
		t.Error("double strike with only four mana spent")
	}
	effectstest.SocB17CastSpellFor(t, g, "Sorcery", "{R}", "R")
	c = effectstest.S58p6View(t, g, dancer)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("Firedancer is %d/%d after a second spell, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSocB24FiredancerGainsDoubleStrikeOnFiveOrMoreManaAndItLapses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dancer := socB24Firedancer(g, me.ID)
	effectstest.SocB17CastSpellFor(t, g, "Sorcery", "{4}{R}", "RRRRR")
	c := effectstest.S58p6View(t, g, dancer)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Firedancer is %d/%d after a five-mana spell, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, dancer, "double strike") {
		t.Fatal("no double strike after five mana was spent")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	c = effectstest.S58p6View(t, g, dancer)
	if c.CurrentPower() != 2 || effectstest.HasEffectiveKeyword(t, g, dancer, "double strike") {
		t.Errorf("Firedancer is %d/%d with double strike %v on my next turn: both are until end of turn", c.CurrentPower(), c.CurrentToughness(), effectstest.HasEffectiveKeyword(t, g, dancer, "double strike"))
	}
}

func TestSocB24FiredancerIgnoresCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dancer := socB24Firedancer(g, me.ID)
	effectstest.SocB17CastSpellFor(t, g, "Creature — Bear", "{4}{R}", "RRRRR")
	if c := effectstest.S58p6View(t, g, dancer); c.CurrentPower() != 2 || effectstest.HasEffectiveKeyword(t, g, dancer, "double strike") {
		t.Errorf("a creature spell pumped the Firedancer (%d/%d)", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Fields of Strife ------------------------------------------------------------------

func TestSocB24FieldsOfStrifeEntersTappedMakesRedOrWhiteAndSurveilsForFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top", "Next")
	id := effectstest.PlayLandFromHand(t, g, "Fields of Strife", socB24FieldsOfStrifeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatalf("on battlefield %v, tapped %v; want it to enter tapped", ok, card.Tapped)
	}
	effectstest.UntapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("the land made {B}")
	}
	for _, colour := range []string{"R", "W"} {
		effectstest.UntapForTest(g, id)
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
			t.Fatalf("{T}: Add {%s}: %v", colour, err)
		}
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	}

	effectstest.UntapForTest(g, id)
	effectstest.FloatForTest(g, me, "CCRR")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("surveil activated without {W}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "CCRW")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{2}{R}{W}, {T}: surveil 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
	top := c.ScryCards[0]
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not go to the graveyard")
	}
	if card, _ := effectstest.BattlefieldCard(g, id); !card.Tapped {
		t.Error("the {T} cost did not tap the land")
	}
}

// --- Fix What's Broken ------------------------------------------------------------------

func TestSocB24FixWhatsBrokenPaysXLifeAndReturnsEveryArtifactAndCreatureOfThatManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creature := socB24Graveyard(me, "Two Creature", "Creature — Bear", "{1}{G}")
	artifact := socB24Graveyard(me, "Two Artifact", "Artifact", "{2}")
	both := socB24Graveyard(me, "Two Both", "Artifact Creature — Golem", "{2}")
	bigger := socB24Graveyard(me, "Three Creature", "Creature — Bear", "{2}{G}")
	spell := socB24Graveyard(me, "Two Instant", "Instant", "{1}{U}")
	land := socB24Graveyard(me, "Dead Land", "Land", "")
	theirs := socB24Graveyard(opp, "Their Two", "Creature — Bear", "{1}{G}")
	life := me.Life
	effectstest.CastWipe(t, g, "Fix What's Broken", "Sorcery", socB24FixWhatsBrokenOracle, game.CastSpellParams{XValue: 2})
	if me.Life != life-2 {
		t.Fatalf("life %d -> %d at cast, want -2 (X life is a cast cost)", life, me.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"the creature": creature, "the artifact": artifact, "the artifact creature": both} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s with mana value 2 did not return", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"the mana value 3 creature": bigger, "the instant": spell, "the land": land} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("%s left the graveyard", name)
		}
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("a card from an opponent's graveyard returned")
	}
}

func TestSocB24FixWhatsBrokenWithXZeroReturnsOnlyManaValueZeroPermanentCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	free := socB24Graveyard(me, "Free Artifact", "Artifact", "")
	creature := socB24Graveyard(me, "One Creature", "Creature — Bear", "{G}")
	land := socB24Graveyard(me, "Dead Land", "Land", "")
	life := me.Life
	effectstest.CastWipe(t, g, "Fix What's Broken", "Sorcery", socB24FixWhatsBrokenOracle, game.CastSpellParams{XValue: 0})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d -> %d with X = 0", life, me.Life)
	}
	if !g.Battlefield.Contains(free) {
		t.Error("the mana value 0 artifact did not return")
	}
	if !me.Graveyard.Contains(creature) || !me.Graveyard.Contains(land) {
		t.Error("a mana value 1 creature or a land returned on X = 0")
	}
}

func TestSocB24FixWhatsBrokenCantPayMoreLifeThanYouHave(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Life = 3
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fix What's Broken", TypeLine: "Sorcery", OracleID: socB24FixWhatsBrokenOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 5}); err == nil {
		t.Error("paid 5 life with 3")
	}
	if me.Life != 3 {
		t.Errorf("life %d after the refused cast", me.Life)
	}
}

// --- Flashback ----------------------------------------------------------------------------

func TestSocB24FlashbackGivesAnInstantFlashbackAtItsManaCostAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	spell := socB24Graveyard(me, "Filler Bolt", "Instant", "{1}{R}")
	effectstest.CastCatalogSpell(t, g, "Flashback", "Instant", socB24FlashbackOracle, effectstest.CardRefs(spell))
	effectstest.PassPriorityAroundTable(t, g)
	perm := effectstest.GrantedPermissionOn(g, me.ID, spell, game.ZoneGraveyard)
	if !perm.Granted() || perm.AltCostKey != "flashback" || !perm.ExileOnResolution || perm.Cost != "" {
		t.Fatalf("granted permission %+v: want flashback at the card's own mana cost, exiling on resolution", perm)
	}
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err == nil {
		t.Fatal("cast the {1}{R} card for flashback with only {R}")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("flashback cast for {1}{R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("the flashed-back card is not exiled")
	}
}

func TestSocB24FlashbackLapsesAtTheEndOfTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	spell := socB24Graveyard(me, "Filler Bolt", "Sorcery", "{R}")
	effectstest.CastCatalogSpell(t, g, "Flashback", "Instant", socB24FlashbackOracle, effectstest.CardRefs(spell))
	effectstest.PassPriorityAroundTable(t, g)
	perm := effectstest.GrantedPermissionOn(g, me.ID, spell, game.ZoneGraveyard)
	if !perm.Granted() || !effectstest.PermissionLive(g, perm, me.ID) {
		t.Fatal("no live flashback on a sorcery card")
	}
	later := g.Clone()
	later.Seats[later.Turn.ActiveSeat].TurnsBegun++
	if effectstest.PermissionLive(later, perm, me.ID) {
		t.Error("the grant outlived the turn it was made on")
	}
}

func TestSocB24FlashbackTargetsOnlyYourInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	creature := socB24Graveyard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	theirs := socB24Graveyard(opp, "Their Bolt", "Instant", "{R}")
	for name, id := range map[string]uuid.UUID{"a creature card": creature, "an opponent's instant": theirs} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Flashback", "Instant", socB24FlashbackOracle, effectstest.CardRefs(id)); err != game.ErrIllegalTarget {
			t.Errorf("Flashback targeting %s: err = %v, want ErrIllegalTarget", name, err)
		}
	}
}
