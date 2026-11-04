package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b36_cards_test.go — Secrets of Strixhaven, thirty-sixth batch (#4456).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	socB36StandUpOracle      = "00e94b75-ad31-4e1f-b3be-eed7cd3e4d76"
	socB36SlothOracle        = "5d46e85f-4a04-48b1-afe9-3a47678041d4"
	socB36HonormancerOracle  = "3bcf54f8-fcd7-4ec0-b011-92b367411cdb"
	socB36HopesingerOracle   = "7244d04a-bb9c-4fbe-a346-4a3f45709da7"
	socB36DocentOracle       = "dbb1690a-f6f5-459c-9d7f-cb6541fb6688"
	socB36StressDreamOracle  = "52124954-b0b6-492c-8d03-fe54c220facc"
	socB36StrifeScholarOracl = "e85c6bd2-db15-4052-88c0-3ae25414f0ca"
	socB36SkycoachOracle     = "e286a191-39d4-4073-b844-11d2aba911b4"
	socB36FirstYearOracle    = "58b0c737-0a84-4f9a-b3b7-300c5de43874"
	socB36DromedaryOracle    = "697b14c6-fe6a-4880-9cf2-b6602a52ca0b"
)

func TestSocB36EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB36StandUpOracle:             "Stand Up for Yourself",
		socB36SlothOracle:               "Startled Relic Sloth",
		socB36HonormancerOracle:         "Stirring Honormancer",
		socB36HopesingerOracle:          "Stirring Hopesinger",
		socB36DocentOracle:              "Stone Docent",
		socB36StressDreamOracle:         "Stress Dream",
		socB36StrifeScholarOracl:        "Strife Scholar",
		socB36StrifeScholarOracl + "#1": "Awaken the Ages",
		socB36SkycoachOracle:            "Strixhaven Skycoach",
		socB36FirstYearOracle:           "Studious First-Year",
		socB36FirstYearOracle + "#1":    "Rampant Growth",
		socB36DromedaryOracle:           "Summoned Dromedary",
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

func socB36CardRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- Stand Up for Yourself ---------------------------------------------------

func TestSocB36StandUpForYourselfDestroysACreatureWithPowerThreeOrGreater(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	id := socB20HandCard(me, "Stand Up for Yourself", "Instant", "{2}{W}", socB36StandUpOracle)
	if err := socB20CastStrict(t, g, me, id, "WWW", socB36CardRef(ogre)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) || !opp.Graveyard.Contains(ogre) {
		t.Error("the power-3 creature was not destroyed")
	}
}

func TestSocB36StandUpForYourselfCannotTargetAPowerTwoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imp := b12Creature(g, opp.ID, "Imp", "Creature — Imp", 2, 9)
	id := socB20HandCard(me, "Stand Up for Yourself", "Instant", "{2}{W}", socB36StandUpOracle)
	if err := socB20CastStrict(t, g, me, id, "WWW", socB36CardRef(imp)); err == nil {
		t.Fatal("a power-2 creature was accepted as a target")
	}
	if !g.Battlefield.Contains(imp) {
		t.Error("the power-2 creature left the battlefield")
	}
}

// --- Startled Relic Sloth ----------------------------------------------------

func TestSocB36StartledRelicSlothHasTrampleAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sloth := pushCatalogPermanent(g, me.ID, "Startled Relic Sloth", "Creature — Sloth Beast", socB36SlothOracle, false)
	for _, kw := range []string{"trample", "lifelink"} {
		if !hasEffectiveKeyword(t, g, sloth, kw) {
			t.Errorf("the Sloth should have %s", kw)
		}
	}
}

func TestSocB36StartledRelicSlothExilesACardFromAGraveyardAtYourBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Startled Relic Sloth", "Creature — Sloth Beast", socB36SlothOracle, 4, 4)
	mineGy := socB11Graveyard(me, "My Dead", "Creature — Bear", "{1}{G}")
	theirGy := socB11Graveyard(opp, "Their Dead", "Creature — Bear", "{1}{G}")
	advanceTo(t, g, game.StepBeginCombat)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, theirGy)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirGy) || opp.Graveyard.Contains(theirGy) {
		t.Error("the chosen card is not exiled")
	}
	if !me.Graveyard.Contains(mineGy) {
		t.Error("an unchosen card left a graveyard")
	}
}

func TestSocB36StartledRelicSlothDoesNothingAtAnOpponentsBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Startled Relic Sloth", "Creature — Sloth Beast", socB36SlothOracle, 4, 4)
	mineGy := socB11Graveyard(me, "My Dead", "Creature — Bear", "{1}{G}")
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, opp.ID) != nil {
		t.Error("the Sloth triggered on my beginning of combat")
	}
	if !me.Graveyard.Contains(mineGy) {
		t.Error("the Sloth exiled a card on another player's turn")
	}
}

// --- Stirring Honormancer ----------------------------------------------------

func socB36Honormancer() game.Card {
	return game.Card{Name: "Stirring Honormancer", TypeLine: "Creature — Rhino Bard", OracleID: socB36HonormancerOracle,
		Power: 4, Toughness: 5}
}

func TestSocB36StirringHonormancerLooksAtAsManyCardsAsYouHaveCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ids := seedLibrary(me, "A", "B", "C")
	msc17Play(t, g, me, socB36Honormancer())
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick was offered")
	}
	if len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("the look is %d cards, pick %d..%d; want 2 cards (the Honormancer and the Bear) and exactly one",
			len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, ids[1])
	g.SettleResolution()
	if !me.Hand.Contains(ids[1]) {
		t.Error("the chosen card is not in hand")
	}
	if !me.Graveyard.Contains(ids[0]) {
		t.Error("the rest of the look did not go to the graveyard")
	}
	if me.Hand.Contains(ids[2]) || me.Graveyard.Contains(ids[2]) {
		t.Error("a card beyond the top two was moved")
	}
}

func TestSocB36StirringHonormancerAloneTakesTheTopCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedLibrary(me, "A", "B")
	msc17Play(t, g, me, socB36Honormancer())
	g.SettleResolution()
	if !me.Hand.Contains(ids[0]) {
		t.Error("with one creature, the only card looked at was not put into hand")
	}
	if me.Graveyard.Contains(ids[1]) || me.Hand.Contains(ids[1]) {
		t.Error("the second card was touched with X = 1")
	}
}

// --- Stirring Hopesinger -----------------------------------------------------

func TestSocB36StirringHopesingerHasFlyingAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	singer := pushCatalogPermanent(g, me.ID, "Stirring Hopesinger", "Creature — Bird Bard", socB36HopesingerOracle, false)
	for _, kw := range []string{"flying", "lifelink"} {
		if !hasEffectiveKeyword(t, g, singer, kw) {
			t.Errorf("the Hopesinger should have %s", kw)
		}
	}
}

func TestSocB36StirringHopesingerPutsACounterOnEachCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	singer := b12Push(g, me.ID, "Stirring Hopesinger", "Creature — Bird Bard", socB36HopesingerOracle, 1, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	bystander := b12Creature(g, opp.ID, "Bystander", "Creature — Test", 1, 1)
	id := socB20HandCard(me, "Stand Up for Yourself", "Instant", "{2}{W}", socB36StandUpOracle)
	if err := socB20CastStrict(t, g, me, id, "WWW", socB36CardRef(ogre)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for name, c := range map[string]uuid.UUID{"the Hopesinger": singer, "the Bear": bear} {
		if got := fraHasCounters(t, g, c); got != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", name, got)
		}
	}
	if got := fraHasCounters(t, g, bystander); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

func TestSocB36StirringHopesingerIgnoresASpellThatTargetsNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	singer := b12Push(g, me.ID, "Stirring Hopesinger", "Creature — Bird Bard", socB36HopesingerOracle, 1, 3)
	seedLibrary(me, "A", "B")
	id := socB20HandCard(me, "Stress Dream", "Instant", "{3}{U}{R}", socB36StressDreamOracle)
	if err := socB20CastStrict(t, g, me, id, "URGGG", nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := fraHasCounters(t, g, singer); got != 0 {
		t.Errorf("the Hopesinger got %d counters from a spell with no creature target", got)
	}
}

// --- Stone Docent ------------------------------------------------------------

func socB36DocentInGraveyard(me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Stone Docent", TypeLine: "Creature — Spirit Chimera",
		OracleID: socB36DocentOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID})
	return id
}

func TestSocB36StoneDocentGainsLifeAndSurveilsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	docent := socB36DocentInGraveyard(me)
	ids := seedLibrary(me, "A", "B")
	rfCrAAddMana(t, g, me, "{W}")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, docent, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(docent) {
		t.Fatal("the Docent was not exiled as a cost")
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	ask := surveilChoiceFor(g, me.ID)
	if ask == nil {
		t.Fatal("no surveil 1 was asked")
	}
	if len(ask.ScryCards) != 1 || ask.ScryCards[0] != ids[0] {
		t.Fatalf("the surveil looks at %v, want just the top card", ask.ScryCards)
	}
	if err := g.ResolveSurveil(ask.ID, me.ID, []uuid.UUID{ids[0]}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(ids[0]) {
		t.Error("the surveilled card did not go to the graveyard")
	}
}

func TestSocB36StoneDocentIsSorcerySpeedAndGraveyardOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	docent := socB36DocentInGraveyard(me)
	rfCrAAddMana(t, g, me, "{W}")
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, docent, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	if !me.Graveyard.Contains(docent) {
		t.Error("a refused activation moved the Docent")
	}
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	toMain(t, g2)
	live := pushCatalogPermanent(g2, me2.ID, "Stone Docent", "Creature — Spirit Chimera", socB36DocentOracle, false)
	rfCrAAddMana(t, g2, me2, "{W}")
	if err := g2.ActivateCatalogAbility(me2.ID, live, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated from the battlefield")
	}
}

// --- Stress Dream ------------------------------------------------------------

func socB36CastStressDream(t *testing.T, g *game.Game, me *game.Player, targets []game.TargetRef) {
	t.Helper()
	id := socB20HandCard(me, "Stress Dream", "Instant", "{3}{U}{R}", socB36StressDreamOracle)
	if err := socB20CastStrict(t, g, me, id, "URGGG", targets); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestSocB36StressDreamDealsFiveAndThenLooksAtTheTopTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	ids := seedLibrary(me, "A", "B", "C")
	socB36CastStressDream(t, g, me, socB36CardRef(ogre))
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick was offered")
	}
	if len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("the look is %d cards, pick %d..%d; want 2 and exactly one", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, ids[1])
	g.SettleResolution()
	if !opp.Graveyard.Contains(ogre) {
		t.Error("5 damage did not kill the 5/5")
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the chosen card is not in hand")
	}
	if me.Hand.Contains(ids[0]) || me.Graveyard.Contains(ids[0]) {
		t.Error("the other card left the library other than to the bottom")
	}
	if got := me.Library.Cards[0].InstanceID; got != ids[0] {
		t.Errorf("the bottom of the library is %v, want the card not taken", got)
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[2] {
		t.Errorf("the top of the library is %v, want the third card", top)
	}
}

func TestSocB36StressDreamWithNoTargetStillLooksAndDamageIsNotDestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedLibrary(me, "A", "B", "C")
	socB36CastStressDream(t, g, me, nil)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick was offered without a target")
	}
	answerChooseCards(t, g, me.ID, ids[0])
	g.SettleResolution()
	if !me.Hand.Contains(ids[0]) {
		t.Error("the chosen card is not in hand")
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	wall := b12Creature(g2, opp2.ID, "Wall", "Creature — Wall", 0, 6)
	seedLibrary(me2, "A", "B")
	socB36CastStressDream(t, g2, me2, socB36CardRef(wall))
	if !g2.Battlefield.Contains(wall) {
		t.Error("5 damage killed a 0/6")
	}
}

// --- Strife Scholar // Awaken the Ages ---------------------------------------

func socB36StrifeScholar(owner uuid.UUID) game.Card {
	return preparationCard(owner, socB36StrifeScholarOracl, nil,
		game.Face{Name: "Strife Scholar", TypeLine: "Creature — Orc Sorcerer", ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 3, Toughness: 2},
		game.Face{Name: "Awaken the Ages", TypeLine: "Sorcery", ManaCost: "{5}{R}", Colors: []string{"R"}})
}

func TestSocB36StrifeScholarEntersPreparedAndAwakenTheAgesMakesTwoSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	c := socB36StrifeScholar(me.ID)
	castFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Strife Scholar did not enter prepared")
	}
	fraBCastCopy(t, g, me, "Awaken the Ages")
	if got := b16CountNamed(g, "Spirit"); got != 2 {
		t.Fatalf("%d Spirit tokens, want 2", got)
	}
	for _, card := range g.Battlefield.Cards {
		if card.Name != "Spirit" {
			continue
		}
		if card.Controller != me.ID || card.Power != 2 || card.Toughness != 2 {
			t.Errorf("Spirit token is %d/%d under %v, want a 2/2 under me", card.Power, card.Toughness, card.Controller)
		}
		if len(card.Colors) != 2 {
			t.Errorf("Spirit token colours %v, want red and white", card.Colors)
		}
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

func TestSocB36StrifeScholarWardsWithTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scholar := pushCatalogPermanent(g, me.ID, "Strife Scholar", "Creature — Orc Sorcerer", socB36StrifeScholarOracl, false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, g, opp, scholar)
	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the opponent who targeted the Scholar")
	}
	if prompt.LifeCost != 2 {
		t.Errorf("the ward charges %d life, want 2", prompt.LifeCost)
	}
}

// --- Strixhaven Skycoach -----------------------------------------------------

func socB36Skycoach() game.Card {
	return game.Card{Name: "Strixhaven Skycoach", TypeLine: "Artifact — Vehicle", OracleID: socB36SkycoachOracle,
		Power: 3, Toughness: 2}
}

func socB36Forest(p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: p.ID, Controller: p.ID})
	return id
}

func TestSocB36StrixhavenSkycoachMaySearchForABasicLandWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := socB36Forest(me)
	msc17Play(t, g, me, socB36Skycoach())
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if !me.Hand.Contains(forest) {
		t.Fatal("the basic land was not put into hand")
	}
	if me.Library.Contains(forest) {
		t.Error("the land is still in the library")
	}
}

func TestSocB36StrixhavenSkycoachSearchCanBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := socB36Forest(me)
	msc17Play(t, g, me, socB36Skycoach())
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if me.Hand.Contains(forest) || !me.Library.Contains(forest) {
		t.Error("the land moved although the search was declined")
	}
}

func TestSocB36StrixhavenSkycoachFliesAndCrewsWithPowerTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	coach := b12Push(g, me.ID, "Strixhaven Skycoach", "Artifact — Vehicle", socB36SkycoachOracle, 3, 2)
	if !hasEffectiveKeyword(t, g, coach, "flying") {
		t.Error("the Skycoach should have flying")
	}
	small := pushCrewerForTest(g, me.ID, "Small Crewer", 1)
	big := pushCrewerForTest(g, me.ID, "Big Crewer", 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, coach, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Fatal("crew 2 was paid with power 1")
	}
	if hasEffectiveType(t, g, coach, "Creature") {
		t.Fatal("the Skycoach is a creature before it is crewed")
	}
	if err := g.ActivateCatalogAbility(me.ID, coach, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{big}}); err != nil {
		t.Fatalf("crew 2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveType(t, g, coach, "Creature") {
		t.Error("the crewed Skycoach is not a creature")
	}
}

// --- Studious First-Year // Rampant Growth -----------------------------------

func TestSocB36StudiousFirstYearEntersPreparedAndRampantGrowthRamps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	forest := socB36Forest(me)
	c := preparationCard(me.ID, socB36FirstYearOracle, nil,
		game.Face{Name: "Studious First-Year", TypeLine: "Creature — Bear Wizard", ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 1},
		game.Face{Name: "Rampant Growth", TypeLine: "Sorcery", ManaCost: "{1}{G}", Colors: []string{"G"}})
	castFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Studious First-Year did not enter prepared")
	}
	fraBCastCopy(t, g, me, "Rampant Growth")
	takeSearchPrompts(t, g)
	card, ok := battlefieldCard(g, forest)
	if !ok {
		t.Fatal("the basic land is not on the battlefield")
	}
	if !card.Tapped {
		t.Error("the land entered untapped")
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

// --- Summoned Dromedary ------------------------------------------------------

func socB36DromedaryInGraveyard(me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Summoned Dromedary", TypeLine: "Creature — Spirit Camel",
		OracleID: socB36DromedaryOracle, Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID})
	return id
}

func TestSocB36SummonedDromedaryHasVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	camel := pushCatalogPermanent(g, me.ID, "Summoned Dromedary", "Creature — Spirit Camel", socB36DromedaryOracle, false)
	if !hasEffectiveKeyword(t, g, camel, "vigilance") {
		t.Error("the Dromedary should have vigilance")
	}
}

func TestSocB36SummonedDromedaryReturnsToHandFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	camel := socB36DromedaryInGraveyard(me)
	rfCrAAddMana(t, g, me, "{C}{W}")
	b16Activate(t, g, me.ID, camel, 0, game.ActivateAbilityParams{})
	if !me.Hand.Contains(camel) || me.Graveyard.Contains(camel) {
		t.Fatal("the Dromedary did not return to hand")
	}
}

func TestSocB36SummonedDromedaryRefusesOutsideAMainPhaseOrOffTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	camel := socB36DromedaryInGraveyard(me)
	rfCrAAddMana(t, g, me, "{C}{W}")
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, camel, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	if !me.Graveyard.Contains(camel) {
		t.Error("a refused activation moved the Dromedary")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	toMain(t, g2)
	live := pushCatalogPermanent(g2, me2.ID, "Summoned Dromedary", "Creature — Spirit Camel", socB36DromedaryOracle, false)
	rfCrAAddMana(t, g2, me2, "{C}{W}")
	if err := g2.ActivateCatalogAbility(me2.ID, live, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated from the battlefield")
	}
}
