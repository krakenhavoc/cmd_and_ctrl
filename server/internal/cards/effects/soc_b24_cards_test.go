package effects

import (
	"testing"

	"github.com/google/uuid"

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

// socB24Perm puts a permanent with a printed mana cost on the battlefield.
func socB24Perm(g *game.Game, owner uuid.UUID, name, typeLine, cost string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
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

func socB24Named(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

// --- End of the Hunt -----------------------------------------------------------

func socB24EndOfTheHuntCard() game.Card {
	return game.Card{Name: "End of the Hunt", TypeLine: "Sorcery", OracleID: socB24EndOfTheHuntOracle, ManaCost: "{1}{B}"}
}

func TestSocB24EndOfTheHuntLetsTheOpponentExileAmongTheGreatestManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := socB24Perm(g, opp.ID, "Big Beast", "Creature — Beast", "{3}{G}{G}", 5, 5)
	walker := socB24Perm(g, opp.ID, "Tall Walker", "Legendary Planeswalker — Test", "{3}{U}{U}", 0, 0)
	addCountersForTest(g, walker, game.CounterLoyalty, 4)
	small := socB24Perm(g, opp.ID, "Small Bear", "Creature — Bear", "{1}{G}", 2, 2)
	rock := socB24Perm(g, opp.ID, "Big Rock", "Artifact", "{6}", 0, 0)
	mine := socB24Perm(g, me.ID, "Mine Giant", "Creature — Giant", "{7}", 7, 7)
	if _, err := msc17Cast(t, g, me, socB24EndOfTheHuntCard(), game.CastSpellParams{Targets: []game.TargetRef{msc17TargetPlayer(me.ID)}}); err == nil {
		t.Fatal("End of the Hunt targeted its own caster")
	}
	if _, err := msc17Cast(t, g, me, socB24EndOfTheHuntCard(), game.CastSpellParams{Targets: []game.TargetRef{msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster was asked to choose: the opponent chooses")
	}
	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the opponent was not asked")
	}
	if !hasID(pick.ChooseCards, big) || !hasID(pick.ChooseCards, walker) || len(pick.ChooseCards) != 2 {
		t.Fatalf("offered %v: only the creature and planeswalker tied at mana value 5 are legal (not the artifact, the smaller creature or the caster's)", pick.ChooseCards)
	}
	answerChooseCards(t, g, opp.ID, walker)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	king := socB24Perm(g, opp.ID, "King", "Creature — Giant", "{4}{R}", 5, 5)
	pawn := socB24Perm(g, opp.ID, "Pawn", "Creature — Soldier", "{1}", 1, 1)
	if _, err := msc17Cast(t, g, me, socB24EndOfTheHuntCard(), game.CastSpellParams{Targets: []game.TargetRef{msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil || len(pick.ChooseCards) != 1 || !hasID(pick.ChooseCards, king) {
		t.Fatalf("the pick %+v should offer exactly the one greatest permanent", pick)
	}
	answerChooseCards(t, g, opp.ID, king)
	if !g.Exile.Contains(king) || g.Battlefield.Contains(king) {
		t.Error("the greatest creature is not exiled")
	}
	if !g.Battlefield.Contains(pawn) {
		t.Error("the lesser creature left")
	}
}

func TestSocB24EndOfTheHuntAsksNothingOfAnOpponentWithNoCreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := socB24Perm(g, opp.ID, "Big Rock", "Artifact", "{6}", 0, 0)
	if _, err := msc17Cast(t, g, me, socB24EndOfTheHuntCard(), game.CastSpellParams{Targets: []game.TargetRef{msc17TargetPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, opp.ID) != nil || chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a choice was asked with nothing to exile")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a noncreature, nonplaneswalker permanent was exiled")
	}
}

// --- Ennis, Debate Moderator ---------------------------------------------------

func TestSocB24EnnisBlinksAnotherCreatureAndGrowsAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Ennis Bear", 2, 2)
	ennis := sbCastPicking(t, g, "Ennis, Debate Moderator", socB24EnnisOracle, bear)
	if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) {
		t.Fatal("the targeted creature was not exiled")
	}
	if n := counterCount(g, ennis, game.CounterPlusOne); n != 0 {
		t.Fatalf("Ennis has %d counters before the end step", n)
	}
	advanceTo(t, g, game.StepEnd)
	ringSettle(t, g)
	if n := socB24Named(g, "Ennis Bear"); n != 1 {
		t.Fatalf("%d Ennis Bear on the battlefield after the end step, want 1 (returned)", n)
	}
	if n := counterCount(g, ennis, game.CounterPlusOne); n != 1 {
		t.Errorf("Ennis has %d +1/+1 counters, want 1: a card was put into exile this turn", n)
	}
}

func TestSocB24EnnisCantTargetItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Ennis Bear", 2, 2)
	id := castCatalogSpell(t, g, "Ennis, Debate Moderator", "Legendary Creature — Human Cleric", socB24EnnisOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger asked for no target with another creature to blink")
	}
	if hasID(pick.PickTargetCards, id) || !hasID(pick.PickTargetCards, bear) {
		t.Errorf("targets offered %v: want the other creature and not Ennis", pick.PickTargetCards)
	}
}

func TestSocB24EnnisDoesNothingWhenNoCardWasExiledThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ennis := b12Push(g, me.ID, "Ennis, Debate Moderator", "Legendary Creature — Human Cleric", socB24EnnisOracle, 1, 1)
	advanceTo(t, g, game.StepEnd)
	ringSettle(t, g)
	if n := counterCount(g, ennis, game.CounterPlusOne); n != 0 {
		t.Errorf("Ennis has %d counters with nothing exiled this turn, want 0", n)
	}
}

func TestSocB24EnnisCountsAnyCardPutIntoExileByAnyone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ennis := b12Push(g, me.ID, "Ennis, Debate Moderator", "Legendary Creature — Human Cleric", socB24EnnisOracle, 1, 1)
	dead := socB11GraveyardCreature(opp, "Their Dead")
	g.WithWriteLock(func() {
		if err := g.ExileCardsKeyedForEffect([]uuid.UUID{dead}, game.KeyedTail{}); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	advanceTo(t, g, game.StepEnd)
	ringSettle(t, g)
	if n := counterCount(g, ennis, game.CounterPlusOne); n != 1 {
		t.Errorf("Ennis has %d counters after an opponent's graveyard card was exiled, want 1", n)
	}
}

func TestSocB24EnnisIgnoresAnExiledToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ennis := b12Push(g, me.ID, "Ennis, Debate Moderator", "Legendary Creature — Human Cleric", socB24EnnisOracle, 1, 1)
	token := b12Creature(g, me.ID, "Soldier", "Token Creature — Soldier", 1, 1)
	g.WithWriteLock(func() {
		if err := g.ExileCardsKeyedForEffect([]uuid.UUID{token}, game.KeyedTail{}); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	advanceTo(t, g, game.StepEnd)
	ringSettle(t, g)
	if n := counterCount(g, ennis, game.CounterPlusOne); n != 0 {
		t.Errorf("Ennis has %d counters after only a token was exiled, want 0: a token is not a card", n)
	}
}

func TestSocB24EnnisSeesACardExiledFromAGraveyardAsACost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ennis := b12Push(g, me.ID, "Ennis, Debate Moderator", "Legendary Creature — Human Cleric", socB24EnnisOracle, 1, 1)
	id := pushCatalogGraveyardCard(me, "Eternal Student", "Creature — Zombie Warlock", socB24StudentOracle, 4, 2)
	floatForTest(g, me, "CB")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Eternal Student: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	ringSettle(t, g)
	if n := counterCount(g, ennis, game.CounterPlusOne); n != 1 {
		t.Errorf("Ennis has %d counters after a card was exiled as a cost, want 1", n)
	}
}

// --- Environmental Scientist -----------------------------------------------------

func TestSocB24ScientistFetchesABasicLandToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Taiga", "Land — Mountain Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	castCatalogSpell(t, g, "Environmental Scientist", "Creature — Human Druid", socB24ScientistOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger asked for no search")
	}
	if searchOptionNamed(g, c, "Taiga") != uuid.Nil || searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a nonbasic land or a nonland card is offered")
	}
	handBefore := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Forest")
	if me.Hand.Size() != handBefore+1 || !b02bHandHasNamed(me, "Forest") {
		t.Errorf("hand %d -> %d, Forest in hand %v: want the basic land in hand", handBefore, me.Hand.Size(), b02bHandHasNamed(me, "Forest"))
	}
}

func TestSocB24ScientistMayFailToFind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"))
	castCatalogSpell(t, g, "Environmental Scientist", "Creature — Human Druid", socB24ScientistOracle, nil)
	passPriorityAroundTable(t, g)
	handBefore := me.Hand.Size()
	answerSearchFailToFind(t, g, me.ID)
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d -> %d after declining to find", handBefore, me.Hand.Size())
	}
}

// --- Essenceknit Scholar -----------------------------------------------------------

func TestSocB24ScholarMakesAPestThatGainsLifeWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Essenceknit Scholar", "Creature — Dryad Warlock", socB24ScholarOracle, nil)
	passPriorityAroundTable(t, g)
	pestID := findBattlefieldByName(g, "Pest")
	if pestID == uuid.Nil {
		t.Fatal("no Pest was created")
	}
	pest := s58p6View(t, g, pestID)
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
	declareAttack(t, g, opp.ID, pestID)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d when the Pest attacked, want +1", life, me.Life)
	}
}

func TestSocB24ScholarsPestGainsNothingWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Essenceknit Scholar", "Creature — Dryad Warlock", socB24ScholarOracle, nil)
	passPriorityAroundTable(t, g)
	pestID := findBattlefieldByName(g, "Pest")
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pestID) })
	passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			b12Push(g, me.ID, "Essenceknit Scholar", "Creature — Dryad Warlock", socB24ScholarOracle, 3, 1)
			if tc.owner >= 0 {
				victim := b12Creature(g, g.Seats[tc.owner].ID, "Victim", "Creature — Bear", 2, 2)
				g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
				ringSettle(t, g)
			}
			advanceTo(t, g, game.StepEnd)
			hand := me.Hand.Size()
			ringSettle(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain {
		t.Fatalf("fixture starts in a main phase (%s): the test wants an instant-speed window", g.Turn.Step)
	}
	id := pushCatalogGraveyardCard(me, "Eternal Student", "Creature — Zombie Warlock", socB24StudentOracle, 4, 2)
	floatForTest(g, me, "RR")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with {R}{R}: it costs {1}{B}")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("the refused activation still moved the card")
	}
	floatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{1}{B}, exile this card from your graveyard (outside a main phase): %v", err)
	}
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("the card is not exiled as a cost")
	}
	passPriorityAroundTable(t, g)
	if n := socB24Named(g, "Inkling"); n != 2 {
		t.Fatalf("%d Inklings, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Inkling" {
			continue
		}
		v := s58p6View(t, g, c.InstanceID)
		if !v.IsToken() || v.CurrentPower() != 1 || v.CurrentToughness() != 1 || len(v.Colors) != 2 || !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
			t.Errorf("an Inkling is token %v %d/%d colors %v, want a 1/1 white and black flying token", v.IsToken(), v.CurrentPower(), v.CurrentToughness(), v.Colors)
		}
	}
}

func TestSocB24EternalStudentHasNoAbilityOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b12Push(g, me.ID, "Eternal Student", "Creature — Zombie Warlock", socB24StudentOracle, 4, 2)
	floatForTest(g, me, "CB")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the graveyard ability was activated from the battlefield")
	}
}

// --- Exhibition Tidecaller ---------------------------------------------------------

// socB24CastOpusSpell casts a plain sorcery for `cost`, paid with exactly
// `pool`, answers the Opus trigger's player target and settles the table.
func socB24CastOpusSpell(t *testing.T, g *game.Game, typeLine, cost, pool string, target uuid.UUID) {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := socB17Hand(me, "Test Spell", typeLine, cost, socB17PlainSorceryOracle)
	floatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", cost, err)
	}
	if target != uuid.Nil {
		if latestPickTarget(g, me.ID) == nil {
			t.Fatal("the Opus trigger asked for no target")
		}
		pickPlayer(t, g, me.ID, target)
	}
	for answerAnyTriggerOrderPrompt(t, g, me.ID) {
	}
	passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pushCatalogPermanent(g, me.ID, "Exhibition Tidecaller", "Creature — Djinn Wizard", socB24TidecallerOracle, false)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Exhibition Tidecaller", "Creature — Djinn Wizard", socB24TidecallerOracle, false)
	before := me.Library.Size()
	socB24CastOpusSpell(t, g, "Instant", "{1}{R}", "RR", me.ID)
	if got := before - me.Library.Size(); got != 3 {
		t.Errorf("my library lost %d cards, want 3", got)
	}
}

func TestSocB24TidecallerMillsWhatIsLeftInAShortLibraryAndNobodyLoses(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Exhibition Tidecaller", "Creature — Djinn Wizard", socB24TidecallerOracle, false)
	seedSearchLibrary(opp, game.Card{Name: "A"}, game.Card{Name: "B"}, game.Card{Name: "C"}, game.Card{Name: "D"})
	grave := opp.Graveyard.Size()
	socB24CastOpusSpell(t, g, "Sorcery", "{4}{R}", "RRRRR", opp.ID)
	if opp.Library.Size() != 0 || opp.Graveyard.Size() != grave+4 {
		t.Errorf("library %d, graveyard +%d: want everything milled and nothing more", opp.Library.Size(), opp.Graveyard.Size()-grave)
	}
}

func TestSocB24TidecallerIgnoresCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Exhibition Tidecaller", "Creature — Djinn Wizard", socB24TidecallerOracle, false)
	before := opp.Library.Size()
	socB24CastOpusSpell(t, g, "Creature — Bear", "{4}{R}", "RRRRR", uuid.Nil)
	if opp.Library.Size() != before || latestPickTarget(g, me.ID) != nil {
		t.Error("a creature spell triggered Opus")
	}
}

// --- Expressive Firedancer -----------------------------------------------------------

func socB24Firedancer(g *game.Game, owner uuid.UUID) uuid.UUID {
	return socB17Perm(g, owner, "Expressive Firedancer", "Creature — Human Sorcerer", socB24FiredancerOracle, "{1}{R}", 2, 2)
}

func TestSocB24FiredancerGetsPlusOneOnAnyInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dancer := socB24Firedancer(g, me.ID)
	socB17CastSpellFor(t, g, "Instant", "{3}{R}", "RRRR")
	c := s58p6View(t, g, dancer)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Fatalf("Firedancer is %d/%d after a four-mana spell, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if hasEffectiveKeyword(t, g, dancer, "double strike") {
		t.Error("double strike with only four mana spent")
	}
	socB17CastSpellFor(t, g, "Sorcery", "{R}", "R")
	c = s58p6View(t, g, dancer)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("Firedancer is %d/%d after a second spell, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSocB24FiredancerGainsDoubleStrikeOnFiveOrMoreManaAndItLapses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dancer := socB24Firedancer(g, me.ID)
	socB17CastSpellFor(t, g, "Sorcery", "{4}{R}", "RRRRR")
	c := s58p6View(t, g, dancer)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Firedancer is %d/%d after a five-mana spell, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, dancer, "double strike") {
		t.Fatal("no double strike after five mana was spent")
	}
	advanceToMainOf(t, g, 1)
	advanceToMainOf(t, g, 0)
	c = s58p6View(t, g, dancer)
	if c.CurrentPower() != 2 || hasEffectiveKeyword(t, g, dancer, "double strike") {
		t.Errorf("Firedancer is %d/%d with double strike %v on my next turn: both are until end of turn", c.CurrentPower(), c.CurrentToughness(), hasEffectiveKeyword(t, g, dancer, "double strike"))
	}
}

func TestSocB24FiredancerIgnoresCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dancer := socB24Firedancer(g, me.ID)
	socB17CastSpellFor(t, g, "Creature — Bear", "{4}{R}", "RRRRR")
	if c := s58p6View(t, g, dancer); c.CurrentPower() != 2 || hasEffectiveKeyword(t, g, dancer, "double strike") {
		t.Errorf("a creature spell pumped the Firedancer (%d/%d)", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Fields of Strife ------------------------------------------------------------------

func TestSocB24FieldsOfStrifeEntersTappedMakesRedOrWhiteAndSurveilsForFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Next")
	id := playLandFromHand(t, g, "Fields of Strife", socB24FieldsOfStrifeOracle)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatalf("on battlefield %v, tapped %v; want it to enter tapped", ok, card.Tapped)
	}
	untapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("the land made {B}")
	}
	for _, colour := range []string{"R", "W"} {
		untapForTest(g, id)
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
			t.Fatalf("{T}: Add {%s}: %v", colour, err)
		}
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	}

	untapForTest(g, id)
	floatForTest(g, me, "CCRR")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("surveil activated without {W}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "CCRW")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{2}{R}{W}, {T}: surveil 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
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
	if card, _ := battlefieldCard(g, id); !card.Tapped {
		t.Error("the {T} cost did not tap the land")
	}
}

// --- Fix What's Broken ------------------------------------------------------------------

func TestSocB24FixWhatsBrokenPaysXLifeAndReturnsEveryArtifactAndCreatureOfThatManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creature := socB24Graveyard(me, "Two Creature", "Creature — Bear", "{1}{G}")
	artifact := socB24Graveyard(me, "Two Artifact", "Artifact", "{2}")
	both := socB24Graveyard(me, "Two Both", "Artifact Creature — Golem", "{2}")
	bigger := socB24Graveyard(me, "Three Creature", "Creature — Bear", "{2}{G}")
	spell := socB24Graveyard(me, "Two Instant", "Instant", "{1}{U}")
	land := socB24Graveyard(me, "Dead Land", "Land", "")
	theirs := socB24Graveyard(opp, "Their Two", "Creature — Bear", "{1}{G}")
	life := me.Life
	castWipe(t, g, "Fix What's Broken", "Sorcery", socB24FixWhatsBrokenOracle, game.CastSpellParams{XValue: 2})
	if me.Life != life-2 {
		t.Fatalf("life %d -> %d at cast, want -2 (X life is a cast cost)", life, me.Life)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	free := socB24Graveyard(me, "Free Artifact", "Artifact", "")
	creature := socB24Graveyard(me, "One Creature", "Creature — Bear", "{G}")
	land := socB24Graveyard(me, "Dead Land", "Land", "")
	life := me.Life
	castWipe(t, g, "Fix What's Broken", "Sorcery", socB24FixWhatsBrokenOracle, game.CastSpellParams{XValue: 0})
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Life = 3
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fix What's Broken", TypeLine: "Sorcery", OracleID: socB24FixWhatsBrokenOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 5}); err == nil {
		t.Error("paid 5 life with 3")
	}
	if me.Life != 3 {
		t.Errorf("life %d after the refused cast", me.Life)
	}
}

// --- Flashback ----------------------------------------------------------------------------

func TestSocB24FlashbackGivesAnInstantFlashbackAtItsManaCostAndExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	spell := socB24Graveyard(me, "Filler Bolt", "Instant", "{1}{R}")
	castCatalogSpell(t, g, "Flashback", "Instant", socB24FlashbackOracle, cardRefs(spell))
	passPriorityAroundTable(t, g)
	perm := grantedPermissionOn(g, me.ID, spell, game.ZoneGraveyard)
	if !perm.Granted() || perm.AltCostKey != "flashback" || !perm.ExileOnResolution || perm.Cost != "" {
		t.Fatalf("granted permission %+v: want flashback at the card's own mana cost, exiling on resolution", perm)
	}
	floatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err == nil {
		t.Fatal("cast the {1}{R} card for flashback with only {R}")
	}
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("flashback cast for {1}{R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("the flashed-back card is not exiled")
	}
}

func TestSocB24FlashbackLapsesAtTheEndOfTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	spell := socB24Graveyard(me, "Filler Bolt", "Sorcery", "{R}")
	castCatalogSpell(t, g, "Flashback", "Instant", socB24FlashbackOracle, cardRefs(spell))
	passPriorityAroundTable(t, g)
	perm := grantedPermissionOn(g, me.ID, spell, game.ZoneGraveyard)
	if !perm.Granted() || !permissionLive(g, perm, me.ID) {
		t.Fatal("no live flashback on a sorcery card")
	}
	later := g.Clone()
	later.Seats[later.Turn.ActiveSeat].TurnsBegun++
	if permissionLive(later, perm, me.ID) {
		t.Error("the grant outlived the turn it was made on")
	}
}

func TestSocB24FlashbackTargetsOnlyYourInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	creature := socB24Graveyard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	theirs := socB24Graveyard(opp, "Their Bolt", "Instant", "{R}")
	for name, id := range map[string]uuid.UUID{"a creature card": creature, "an opponent's instant": theirs} {
		if err := castCatalogSpellErr(t, g, "Flashback", "Instant", socB24FlashbackOracle, cardRefs(id)); err != game.ErrIllegalTarget {
			t.Errorf("Flashback targeting %s: err = %v, want ErrIllegalTarget", name, err)
		}
	}
}
