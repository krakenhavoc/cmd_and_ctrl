package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsa "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/a"
	cardsb "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/b"
	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	cardsmi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/mi"
	cardssh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/sh"
	cardsth "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/th"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB / HOC) cards, half B, batch 9. Every test drives
// the registered catalog row through the engine.

func hobB9Opponent(g *game.Game) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// hobB9Count is how many permanents of that name a player controls.
func hobB9Count(g *game.Game, who uuid.UUID, name string) int {
	return effectstest.HobB6BattlefieldCount(g, who, name)
}

func TestHobB9CardsAreRegisteredFull(t *testing.T) {
	for key, name := range map[string]string{
		cardssh.SmaugTheGreatCalamityOracleID:         "Smaug, the Great Calamity",
		cardssh.SmaugTheGreatCalamityOracleID + "#1":  "Spew Flame",
		cardsb.BeornReluctantHostOracleID:             "Beorn, Reluctant Host",
		cardsb.BeornReluctantHostOracleID + "#1":      "Till and Tend",
		cardskr.LakeTownMarinersOracleID:              "Lake-town Mariners",
		cardskr.LakeTownMarinersOracleID + "#1":       "Gone Fishing",
		cardsmi.MostDecrepitOldBirdOracleID:           "Most Decrepit Old Bird",
		cardsmi.MostDecrepitOldBirdOracleID + "#1":    "Speak Secrets",
		cardsth.ThranduilSindarinLiegeOracleID:        "Thranduil, Sindarin Liege",
		cardsth.ThranduilSindarinLiegeOracleID + "#1": "Silvan Rally",
		cardsa.AnUnexpectedPartyOracleID:              "An Unexpected Party",
		cardsa.AnUnexpectedPartyOracleID + "#1":       "At the Door",
		"011da9c5-aa8a-4fa0-b1f2-62b9f3760476":        "Belladonna Took",
		"008a11c1-d283-49fe-abd7-ff4fe8b1fe79":        "Rhovanion Rampager",
		"e3a665f9-6e51-4e0d-923b-e9552d5978a4":        "The Sackville-Bagginses",
		"bf06d873-08be-4a45-b8fb-fe1dfadc3a39":        "Chief of the Wilds",
		"9924c189-3c82-431a-b570-7805b2a406f0":        "Dáin Ironfoot",
		"94917b80-ac32-45ab-b4ee-157ade024bcd":        "Thorin, Mountain-king",
		"afcb21bb-aeb3-4024-9ce1-68ba7fb64b2d":        "Witch-king, Bringer of Ruin",
		hobB9InsideInformationOracle:                  "Inside Information",
	} {
		spec, ok := effects.Lookup(key)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Smaug, the Great Calamity // Spew Flame -------------------------------

var (
	hobB9SmaugFace = effectstest.HobA6Face("Smaug, the Great Calamity", "Legendary Creature — Dragon", "{5}{R}{R}", 5, 5, "R")
	hobB9SpewFlame = effectstest.HobA6Face("Spew Flame", "Sorcery — Adventure", "{4}{R}", 0, 0, "R")
)

func TestHobB9SpewFlameDealsFiveToACreature(t *testing.T) {
	for _, tc := range []struct {
		toughness int
		dies      bool
	}{{5, true}, {6, false}} {
		g := effectstest.NewCatalogGame(t)
		me, id := effectstest.HobA6HandAdventure(t, g, cardssh.SmaugTheGreatCalamityOracleID, hobB9SmaugFace, hobB9SpewFlame)
		victim := effectstest.HobB4Creature(g, hobB9Opponent(g).ID, "Giant", "Creature — Giant", 3, tc.toughness)

		effectstest.HobA6MustCast(t, g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(victim)})

		if g.Battlefield.Contains(victim) == tc.dies {
			t.Errorf("toughness %d: on battlefield = %v, want %v", tc.toughness, !tc.dies, !tc.dies)
		}
		effectstest.HobA6WentOnAnAdventure(t, g, id)
	}
}

func TestHobB9SpewFlameTargetsOnlyCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardssh.SmaugTheGreatCalamityOracleID, hobB9SmaugFace, hobB9SpewFlame)
	if err := effectstest.HobA6Cast(g, me, id, 1, []game.TargetRef{{Kind: game.TargetPlayer, ID: hobB9Opponent(g).ID}}); err == nil {
		t.Error("Spew Flame was cast at a player")
	}
}

func TestHobB9SmaugIsAFlyingFiveFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardssh.SmaugTheGreatCalamityOracleID, hobB9SmaugFace, hobB9SpewFlame)
	effectstest.HobA6MustCast(t, g, me, id, 0, nil)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Smaug did not resolve onto the battlefield")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 5 || tough != 5 {
		t.Errorf("Smaug is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Smaug has no flying")
	}
}

// --- Beorn, Reluctant Host // Till and Tend --------------------------------

var (
	hobB9BeornFace   = effectstest.HobA6Face("Beorn, Reluctant Host", "Legendary Creature — Human Bear Shapeshifter", "{4}{G}", 5, 5, "G")
	hobB9TillAndTend = effectstest.HobA6Face("Till and Tend", "Sorcery — Adventure", "{1}{G}", 0, 0, "G")
)

func TestHobB9TillAndTendGrantsOneExtraLandPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsb.BeornReluctantHostOracleID, hobB9BeornFace, hobB9TillAndTend)
	if got := g.ExtraLandDropsThisTurn[me.ID]; got != 0 {
		t.Fatalf("extra land plays before the spell = %d", got)
	}
	effectstest.HobA6MustCast(t, g, me, id, 1, nil)
	if got := g.ExtraLandDropsThisTurn[me.ID]; got != 1 {
		t.Errorf("extra land plays = %d, want 1", got)
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

func TestHobB9BeornIsATramplingFiveFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsb.BeornReluctantHostOracleID, hobB9BeornFace, hobB9TillAndTend)
	effectstest.HobA6MustCast(t, g, me, id, 0, nil)
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 5 || tough != 5 {
		t.Errorf("Beorn is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("Beorn has no trample")
	}
}

// --- Lake-town Mariners // Gone Fishing ------------------------------------

var (
	hobB9MarinersFace = effectstest.HobA6Face("Lake-town Mariners", "Creature — Human Citizen", "{4}{U}{U}", 6, 5, "U")
	hobB9GoneFishing  = effectstest.HobA6Face("Gone Fishing", "Instant — Adventure", "{3}{U}", 0, 0, "U")
)

func TestHobB9GoneFishingBlinksTwoOfYourCreaturesAndLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardskr.LakeTownMarinersOracleID, hobB9MarinersFace, hobB9GoneFishing)
	counters := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 2})
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")

	effectstest.HobA6MustCast(t, g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(counters), effectstest.HobA6Card(land)})

	if g.Battlefield.Contains(counters) || g.Battlefield.Contains(land) {
		t.Error("a blinked card kept its old identity: it must return as a new object (CR 400.7)")
	}
	if got := hobB9Count(g, me.ID, "Counter Bearer"); got != 1 {
		t.Fatalf("%d Counter Bearers after the blink, want 1", got)
	}
	if got := hobB9Count(g, me.ID, "Forest"); got != 1 {
		t.Errorf("%d Forests after the blink, want 1", got)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Counter Bearer" && c.Counters[game.CounterPlusOne] != 0 {
				t.Errorf("the returned creature has %d +1/+1 counters, want none", c.Counters[game.CounterPlusOne])
			}
		}
	})
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

func TestHobB9GoneFishingReturnsAStolenCreatureToItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardskr.LakeTownMarinersOracleID, hobB9MarinersFace, hobB9GoneFishing)
	opp := hobB9Opponent(g)
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Ogre", TypeLine: "Creature — Ogre",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID,
	})
	bear := effectstest.SeedCreature(g, "Bear", me.ID)

	effectstest.HobA6MustCast(t, g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(stolen), effectstest.HobA6Card(bear)})

	if got := hobB9Count(g, opp.ID, "Stolen Ogre"); got != 1 {
		t.Errorf("its owner controls %d Stolen Ogres after the blink, want 1", got)
	}
	if got := hobB9Count(g, me.ID, "Stolen Ogre"); got != 0 {
		t.Errorf("the thief still controls %d Stolen Ogres, want 0", got)
	}
}

func TestHobB9GoneFishingNeedsTwoTargetsOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardskr.LakeTownMarinersOracleID, hobB9MarinersFace, hobB9GoneFishing)
	mine := effectstest.SeedCreature(g, "Bear", me.ID)
	theirs := effectstest.SeedCreature(g, "Their Bear", hobB9Opponent(g).ID)
	if err := effectstest.HobA6Cast(g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(mine)}); err == nil {
		t.Error("cast with one target")
	}
	if err := effectstest.HobA6Cast(g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(mine), effectstest.HobA6Card(theirs)}); err == nil {
		t.Error("cast at a creature an opponent controls")
	}
}

func TestHobB9MarinersHaveVigilanceAndWardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mariners := effectstest.PushCatalogPermanent(g, me.ID, "Lake-town Mariners", "Creature — Human Citizen", cardskr.LakeTownMarinersOracleID, false)
	if !effectstest.HasEffectiveKeyword(t, g, mariners, "vigilance") {
		t.Error("the Mariners have no vigilance")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	shock := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: shock, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{Targets: []game.TargetRef{effectstest.HobA6Card(mariners)}}); err != nil {
		t.Fatalf("an opponent targets the Mariners: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("ward {2} did not ask the targeting opponent for payment")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(shock) {
		t.Error("declining the ward did not counter the spell")
	}
}

// --- Most Decrepit Old Bird // Speak Secrets --------------------------------

var (
	hobB9BirdFace     = effectstest.HobA6Face("Most Decrepit Old Bird", "Creature — Bird", "{U}", 1, 1, "U")
	hobB9SpeakSecrets = effectstest.HobA6Face("Speak Secrets", "Sorcery — Adventure", "{1}{U}", 0, 0, "U")
)

func TestHobB9OldBirdGrowsAtThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bird := effectstest.PushCatalogPermanent(g, me.ID, "Most Decrepit Old Bird", "Creature — Bird", cardsmi.MostDecrepitOldBirdOracleID, false)
	if !effectstest.HasEffectiveKeyword(t, g, bird, "flying") {
		t.Error("the Old Bird has no flying")
	}
	for i := 0; i < 6; i++ {
		effectstest.PushGraveyardCardForTest(me, "Fodder")
	}
	effectstest.SeedCreature(g, "Nudge", me.ID) // a battlefield change refreshes the layer cache
	if p, tough := effectstest.EffectivePTOf(t, g, bird); p != 1 || tough != 1 {
		t.Errorf("six cards in the graveyard: %d/%d, want 1/1", p, tough)
	}
	effectstest.PushGraveyardCardForTest(me, "Fodder")
	effectstest.SeedCreature(g, "Nudge", me.ID)
	if p, tough := effectstest.EffectivePTOf(t, g, bird); p != 2 || tough != 2 {
		t.Errorf("seven cards in the graveyard: %d/%d, want 2/2", p, tough)
	}
}

func TestHobB9SpeakSecretsMillsFourAndTakesAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsmi.MostDecrepitOldBirdOracleID, hobB9BirdFace, hobB9SpeakSecrets)
	below := effectstest.PlTop(me, "Below", "Creature — Bear", "{1}{G}")
	forest := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	bear := effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")
	divination := effectstest.PlTop(me, "Divination", "Sorcery", "{2}{U}")
	bolt := effectstest.PlTop(me, "Bolt", "Instant", "{R}")
	graveBefore := me.Graveyard.Size()

	effectstest.HobA6MustCast(t, g, me, id, 1, nil)

	if got := me.Graveyard.Size() - graveBefore; got != 4 {
		t.Fatalf("milled %d cards, want 4", got)
	}
	if !me.Library.Contains(below) {
		t.Error("the fifth card was milled")
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no instant or sorcery was offered from among the milled cards")
	}
	if !effectstest.HasID(pick.ChooseCards, bolt) || !effectstest.HasID(pick.ChooseCards, divination) || effectstest.HasID(pick.ChooseCards, bear) || effectstest.HasID(pick.ChooseCards, forest) {
		t.Errorf("offered %v, want just the instant and the sorcery", pick.ChooseCards)
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 1..1: the card is put into hand, not a \"may\"", pick.ChooseMin, pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bolt)
	if !effectstest.MiInHand(me, bolt) || effectstest.MiInHand(me, divination) {
		t.Error("exactly the chosen instant should reach the hand")
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

func TestHobB9SpeakSecretsWithNothingToTakeAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsmi.MostDecrepitOldBirdOracleID, hobB9BirdFace, hobB9SpeakSecrets)
	for i := 0; i < 4; i++ {
		effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")
	}
	effectstest.HobA6MustCast(t, g, me, id, 1, nil)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a prompt was raised with no instant or sorcery among the milled cards")
	}
}

// --- Thranduil, Sindarin Liege // Silvan Rally -----------------------------

var (
	hobB9ThranduilFace = effectstest.HobA6Face("Thranduil, Sindarin Liege", "Legendary Creature — Elf Noble", "{2}{G/U}{G/U}", 2, 3, "G", "U")
	hobB9SilvanRally   = effectstest.HobA6Face("Silvan Rally", "Sorcery — Adventure", "{1}{G/U}{G/U}", 0, 0, "G", "U")
)

func TestHobB9ThranduilPumpsOtherElvesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thranduil := effectstest.B12Push(g, me.ID, "Thranduil, Sindarin Liege", "Legendary Creature — Elf Noble", cardsth.ThranduilSindarinLiegeOracleID, 2, 3)
	elf := effectstest.B12Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Elf", "Creature — Elf Druid", "", 1, 1)
	for name, want := range map[string]struct {
		id   uuid.UUID
		p, t int
	}{
		"your other Elf":  {elf, 2, 2},
		"Thranduil":       {thranduil, 2, 3},
		"a Bear":          {bear, 2, 2},
		"an opposing Elf": {theirs, 1, 1},
	} {
		if p, tough := effectstest.EffectivePTOf(t, g, want.id); p != want.p || tough != want.t {
			t.Errorf("%s is %d/%d, want %d/%d", name, p, tough, want.p, want.t)
		}
	}
}

func TestHobB9ThranduilLandfallMakesAnElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Thranduil, Sindarin Liege", "Legendary Creature — Elf Noble", cardsth.ThranduilSindarinLiegeOracleID, 2, 3)
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := hobB9Count(g, me.ID, "Elf"); got != 1 {
		t.Fatalf("%d Elf tokens, want 1", got)
	}
	token := effectstest.FindBattlefieldByName(g, "Elf")
	if p, tough := effectstest.EffectivePTOf(t, g, token); p != 2 || tough != 2 {
		t.Errorf("the Elf token is %d/%d with Thranduil's anthem, want 2/2", p, tough)
	}
	if got := hobB9Count(g, opp.ID, "Elf"); got != 0 {
		t.Errorf("the opponent has %d Elf tokens", got)
	}
}

func TestHobB9SilvanRallyTakesUpToTwoLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsth.ThranduilSindarinLiegeOracleID, hobB9ThranduilFace, hobB9SilvanRally)
	bear := effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")
	lands := []uuid.UUID{
		effectstest.PlTop(me, "Forest", "Basic Land — Forest", ""),
		effectstest.PlTop(me, "Island", "Basic Land — Island", ""),
		effectstest.PlTop(me, "Swamp", "Basic Land — Swamp", ""),
	}
	effectstest.HobA6MustCast(t, g, me, id, 1, nil)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land was offered from among the milled cards")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Errorf("bounds %d..%d, want 0..2", pick.ChooseMin, pick.ChooseMax)
	}
	if effectstest.HasID(pick.ChooseCards, bear) || len(pick.ChooseCards) != 3 {
		t.Errorf("offered %v, want exactly the three lands", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, lands[0], lands[1])
	if !effectstest.MiInHand(me, lands[0]) || !effectstest.MiInHand(me, lands[1]) || effectstest.MiInHand(me, lands[2]) {
		t.Error("exactly the two chosen lands should reach the hand")
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

// --- An Unexpected Party // At the Door ------------------------------------

var (
	hobB9PartyFace = effectstest.HobA6Face("An Unexpected Party", "Enchantment", "{2}{W}{W}", 0, 0, "W")
	hobB9AtTheDoor = effectstest.HobA6Face("At the Door", "Sorcery — Adventure", "{X}{2}{W}", 0, 0, "W")
)

func TestHobB9UnexpectedPartyPumpsTheChosenTypeYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := hobB9Opponent(g)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsa.AnUnexpectedPartyOracleID, hobB9PartyFace, hobB9AtTheDoor)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)

	effectstest.HobA6MustCast(t, g, me, id, 0, nil)
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	effectstest.PassPriorityAroundTable(t, g)

	for name, want := range map[string]struct {
		id   uuid.UUID
		p, t int
	}{
		"your Elf":        {elf, 3, 3},
		"a Bear":          {bear, 2, 2},
		"an opposing Elf": {theirs, 1, 1},
	} {
		if p, tough := effectstest.EffectivePTOf(t, g, want.id); p != want.p || tough != want.t {
			t.Errorf("%s is %d/%d, want %d/%d", name, p, tough, want.p, want.t)
		}
	}
}

func TestHobB9AtTheDoorMakesXDwarves(t *testing.T) {
	for _, x := range []int{0, 3} {
		g := effectstest.NewCatalogGame(t)
		me, id := effectstest.HobA6HandAdventure(t, g, cardsa.AnUnexpectedPartyOracleID, hobB9PartyFace, hobB9AtTheDoor)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1, XValue: x}); err != nil {
			t.Fatalf("cast At the Door for X = %d: %v", x, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := hobB9Count(g, me.ID, "Dwarf"); got != x {
			t.Errorf("X = %d made %d Dwarves", x, got)
		}
		if x > 0 {
			dwarf := effectstest.FindBattlefieldByName(g, "Dwarf")
			if p, tough := effectstest.EffectivePTOf(t, g, dwarf); p != 2 || tough != 2 {
				t.Errorf("a Dwarf is %d/%d, want 2/2", p, tough)
			}
			effectstest.HobA6WentOnAnAdventure(t, g, id)
		}
	}
}

// --- Belladonna Took -------------------------------------------------------

func TestHobB9BelladonnaTookEscalatesWithEachToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	took := effectstest.HobPush(t, g, me.ID, "Belladonna Took", "Legendary Creature — Halfling Citizen", 2, 2)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	token := func() {
		g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	life, hand := me.Life, me.Hand.Size()

	token()
	if me.Life != life+1 || me.Hand.Size() != hand || effectstest.HobB7Counters(t, g, bear) != 0 {
		t.Errorf("first token: life %d -> %d, hand %d -> %d, bear counters %d; want +1 life only", life, me.Life, hand, me.Hand.Size(), effectstest.HobB7Counters(t, g, bear))
	}
	token()
	if me.Life != life+1 || me.Hand.Size() != hand+1 || effectstest.HobB7Counters(t, g, bear) != 0 {
		t.Errorf("second token: life %d -> %d, hand %d -> %d, bear counters %d; want a card only", life, me.Life, hand, me.Hand.Size(), effectstest.HobB7Counters(t, g, bear))
	}
	token()
	if me.Life != life+1 || me.Hand.Size() != hand+1 || effectstest.HobB7Counters(t, g, bear) != 1 || effectstest.HobB7Counters(t, g, took) != 1 {
		t.Errorf("third token: bear counters %d, Belladonna counters %d; want one on each creature", effectstest.HobB7Counters(t, g, bear), effectstest.HobB7Counters(t, g, took))
	}
	token()
	if me.Life != life+1 || me.Hand.Size() != hand+1 || effectstest.HobB7Counters(t, g, bear) != 1 {
		t.Error("the fourth token should do nothing")
	}
	// An opponent's token is not yours.
	before := opp.Life
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 || opp.Life != before || me.Hand.Size() != hand+1 {
		t.Error("an opponent's token triggered Belladonna Took")
	}
}

// --- Rhovanion Rampager ----------------------------------------------------

func hobB9RampagerAttack(t *testing.T, g *game.Game, me *game.Player, fodder uuid.UUID) uuid.UUID {
	t.Helper()
	rampager := effectstest.HobPush(t, g, me.ID, "Rhovanion Rampager", "Creature — Wolf", 3, 2)
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, rampager)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !effectstest.HasID(prompt.ChooseCards, fodder) || effectstest.HasID(prompt.ChooseCards, rampager) {
		t.Fatalf("the sacrifice is asked as the trigger resolves, over another creature, with a floor of zero: %+v", prompt)
	}
	return rampager
}

func TestHobB9RampagerSacrificeBecomesCountersEqualToItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// A 2/2 with a +1/+1 counter: its power as it last stood is 3.
	fodder := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	rampager := hobB9RampagerAttack(t, g, me, fodder)

	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(fodder) {
		t.Error("the chosen creature was not sacrificed")
	}
	if got := effectstest.HobB7Counters(t, g, rampager); got != 3 {
		t.Errorf("the Rampager has %d +1/+1 counters, want 3 (the fodder's power with its counter)", got)
	}
}

func TestHobB9RampagerDeclinedSacrificesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.HobB4Creature(g, me.ID, "Fodder", "Creature — Ogre", 3, 3)
	rampager := hobB9RampagerAttack(t, g, me, fodder)

	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(fodder) || effectstest.HobB7Counters(t, g, rampager) != 0 {
		t.Error("declining the sacrifice still cost a creature or gave counters")
	}
}

func TestHobB9RampagerDiesAndAmassesGoblinsByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.HobB4Creature(g, me.ID, "Fodder", "Creature — Ogre", 3, 3)
	rampager := hobB9RampagerAttack(t, g, me, fodder)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobB7Counters(t, g, rampager); got != 3 {
		t.Fatalf("the Rampager has %d counters, want 3", got)
	}

	effectstest.HobDestroy(t, g, rampager)

	if got := effectstest.HobArmyCounters(t, g, me.ID); got != 6 {
		t.Errorf("the Goblin Army has %d counters, want 6: the Rampager's power as it last stood (3 base + 3 counters)", got)
	}
}

// --- The Sackville-Bagginses -----------------------------------------------

func hobB9SackvilleEnter(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return hobCast(t, g, "The Sackville-Bagginses", "Legendary Creature — Halfling Citizen", 2, 2)
}

func TestHobB9SackvilleSacrificeDrawsAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	hand := me.Hand.Size()
	lives := effectstest.LifeOfOpponents(g)

	hobB9SackvilleEnter(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !effectstest.HasID(prompt.ChooseCards, bear) {
		t.Fatalf("the sacrifice is a \"may\" over another creature or artifact: %+v", prompt)
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
	if got := hobB9Count(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
	for i, l := range lives {
		if g.Seats[i+1].Life != l {
			t.Error("a nontoken sacrifice drained an opponent")
		}
	}
}

func TestHobB9SackvilleCanSacrificeAnArtifactOrDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.PushTypedCard(g, me.ID, "Relic", "Artifact", "")
	hand := me.Hand.Size()
	hobB9SackvilleEnter(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || !effectstest.HasID(prompt.ChooseCards, relic) {
		t.Fatalf("an artifact is a legal sacrifice: %+v", prompt)
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(relic) || me.Hand.Size() != hand || hobB9Count(g, me.ID, "Treasure") != 0 {
		t.Error("declining still sacrificed, drew or made a Treasure")
	}
}

func TestHobB9SackvilleSacrificingATokenDrainsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	treasure := effectstest.FindBattlefieldByName(g, "Treasure")
	hobB9SackvilleEnter(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, treasure)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	before := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("the targeted opponent's life %d -> %d, want -1", before, opp.Life)
	}
	if got := hobB9Count(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want the new one", got)
	}
}

func TestHobB9SackvilleDrainsOnlyForYourOwnTokenSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.HobPush(t, g, me.ID, "The Sackville-Bagginses", "Legendary Creature — Halfling Citizen", 2, 2)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 1)
	})
	mine, theirs := effectstest.FindBattlefieldByName(g, "Treasure"), uuid.Nil
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Treasure" && c.Controller == opp.ID {
				theirs = c.InstanceID
			}
		}
	})
	if mine == theirs {
		mine = uuid.Nil
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Treasure" && c.Controller == me.ID {
					mine = c.InstanceID
				}
			}
		})
	}
	before := opp.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || opp.Life != before {
		t.Error("an opponent's token sacrifice triggered The Sackville-Bagginses")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d -> %d after your own token sacrifice, want -1", before, opp.Life)
	}
}

// --- Chief of the Wilds ----------------------------------------------------

const hobB9ChiefType = "Legendary Creature — Wolf"

func TestHobB9ChiefHasMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	chief := effectstest.HobPush(t, g, g.Seats[0].ID, "Chief of the Wilds", hobB9ChiefType, 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, chief, "menace") {
		t.Error("the Chief has no menace")
	}
}

func TestHobB9ChiefGrowsForEveryOtherWolfYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chief := effectstest.HobPush(t, g, me.ID, "Chief of the Wilds", hobB9ChiefType, 4, 4)

	effectstest.HobEnter(t, g, me.ID, game.Card{Name: "Wolf", TypeLine: "Creature — Wolf", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobB7Counters(t, g, chief); got != 2 {
		t.Fatalf("another Wolf entered: %d counters, want 2 (and not doubled)", got)
	}
	effectstest.HobEnter(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.HobEnter(t, g, opp.ID, game.Card{Name: "Their Wolf", TypeLine: "Creature — Wolf", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobB7Counters(t, g, chief); got != 2 {
		t.Errorf("a Bear or an opposing Wolf entered: %d counters, want still 2", got)
	}
}

func TestHobB9ChiefDoublesTheTriggersOfOtherWolvesOnly(t *testing.T) {
	for _, tc := range []struct {
		name  string
		chief bool
		want  int
	}{{"without the Chief", false, 3}, {"with the Chief", true, 6}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			if tc.chief {
				effectstest.HobPush(t, g, me.ID, "Chief of the Wilds", hobB9ChiefType, 4, 4)
			}
			wolf := effectstest.HobPush(t, g, me.ID, "Rhovanion Rampager", "Creature — Wolf", 3, 2)
			effectstest.HobDestroy(t, g, wolf)
			if got := effectstest.HobArmyCounters(t, g, me.ID); got != tc.want {
				t.Errorf("a Wolf's dies trigger amassed %d, want %d", got, tc.want)
			}
		})
	}
}

func TestHobB9ChiefDoesNotDoubleOtherCreaturesTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.HobPush(t, g, me.ID, "Chief of the Wilds", hobB9ChiefType, 4, 4)
	goblin := effectstest.HobPush(t, g, me.ID, "Fearsome Goblin Pair", "Creature — Goblin Soldier", 1, 1)
	effectstest.HobDestroy(t, g, goblin)
	if got := effectstest.HobArmyCounters(t, g, me.ID); got != 4 {
		t.Errorf("a Goblin's dies trigger amassed %d, want 4, once", got)
	}
}

func TestHobB9ChiefDoublesABattlesTriggers(t *testing.T) {
	for _, tc := range []struct {
		name  string
		chief bool
		want  int
	}{{"without the Chief", false, 2}, {"with the Chief", true, 4}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			seat := g.Turn.ActiveSeat
			owner := g.Seats[seat]
			protector := g.Seats[(seat+1)%len(g.Seats)]
			if tc.chief {
				effectstest.HobPush(t, g, owner.ID, "Chief of the Wilds", hobB9ChiefType, 4, 4)
			}
			id := effectstest.ImportToHand(effectstest.NewPhyrexiaRow(), owner)
			effectstest.ToMainPhaseSameTurn(t, g)
			if err := g.CastSpell(owner.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
				t.Fatalf("cast the Siege for X = 2: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if got := hobB9Count(g, owner.ID, "Knight"); got != tc.want {
				t.Errorf("the Battle's enters trigger made %d Knights, want %d", got, tc.want)
			}
		})
	}
}

// hobB9Damage is the damage marked on a battlefield creature.
func hobB9Damage(g *game.Game, id uuid.UUID) int {
	c, _ := effectstest.BattlefieldCard(g, id)
	return c.DamageMarked
}

// hobB9Attach attaches an Equipment to a creature.
func hobB9Attach(t *testing.T, g *game.Game, equipment, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(equipment, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
}

// --- Dáin Ironfoot ---------------------------------------------------------

func TestHobB9DainCreatesAnAxeAndAttachesItToATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedCreature(g, "Bear", me.ID)

	hobCast(t, g, "Dáin Ironfoot", "Legendary Creature — Dwarf Warrior", 1, 4)
	axe := effectstest.FindBattlefieldByName(g, "Axe")
	if axe == uuid.Nil {
		t.Fatal("no Axe was created")
	}
	if c, _ := effectstest.BattlefieldCard(g, axe); !c.IsToken() || !c.HasSubtype("Equipment") || !c.IsArtifact() || len(c.Colors) != 0 {
		t.Errorf("the Axe is %q colours %v: want a colorless Equipment artifact token", c.TypeLine, c.Colors)
	}
	if host := effectstest.AttachmentHostOf(t, g, axe); host.ID != uuid.Nil {
		t.Fatal("the Axe was attached before the reflexive trigger chose a creature")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, axe); host.ID != bear {
		t.Errorf("the Axe is on %s, want the Bear", host.ID)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 3 {
		t.Errorf("the equipped Bear has power %d, want 3", p)
	}
}

func TestHobB9DainGivesEquippedAttackersDoubleStrike(t *testing.T) {
	for _, dainEquipped := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dain := effectstest.HobPush(t, g, me.ID, "Dáin Ironfoot", "Legendary Creature — Dwarf Warrior", 1, 4)
		equipped := effectstest.HobB4Creature(g, me.ID, "Equipped Ally", "Creature — Dwarf", 2, 2)
		plain := effectstest.HobB4Creature(g, me.ID, "Plain Ally", "Creature — Dwarf", 2, 2)
		resting := effectstest.HobB4Creature(g, me.ID, "Resting Ally", "Creature — Dwarf", 2, 2)
		hobB9Attach(t, g, effectstest.SeedEquipment(g, me.ID, "Blade", ""), equipped)
		hobB9Attach(t, g, effectstest.SeedEquipment(g, me.ID, "Spare Blade", ""), resting)
		if dainEquipped {
			hobB9Attach(t, g, effectstest.SeedEquipment(g, me.ID, "Dain's Blade", ""), dain)
		}

		effectstest.DeclareAttack(t, g, opp.ID, dain, equipped, plain)
		effectstest.PassPriorityAroundTable(t, g)

		for name, want := range map[string]struct {
			id uuid.UUID
			ds bool
		}{
			"an equipped attacker":     {equipped, true},
			"an unequipped attacker":   {plain, false},
			"an equipped non-attacker": {resting, false},
			"Dáin":                     {dain, dainEquipped},
		} {
			if got := effectstest.HasEffectiveKeyword(t, g, want.id, "double strike"); got != want.ds {
				t.Errorf("Dáin equipped %v: %s has double strike = %v, want %v", dainEquipped, name, got, want.ds)
			}
		}
	}
}

// --- Thorin, Mountain-king -------------------------------------------------

// hobB9CastThorin casts Thorin and answers his two target prompts: the
// Equipment (any number), then the creature.
func hobB9CastThorin(t *testing.T, g *game.Game, me uuid.UUID, equipment []uuid.UUID, creature uuid.UUID) {
	t.Helper()
	effectstest.CastHoldCreature(t, g, "Thorin, Mountain-king", "Legendary Creature — Dwarf Noble", effectstest.HobOracle(t, "Thorin, Mountain-king"), 3, 4)
	first := effectstest.LatestPickTarget(g, me)
	if first == nil {
		t.Fatal("no prompt for the Equipment")
	}
	var picks []game.TargetRef
	for _, id := range equipment {
		picks = append(picks, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.ResolvePickTargets(first.ID, me, picks); err != nil {
		t.Fatalf("pick the Equipment: %v", err)
	}
	second := effectstest.LatestPickTarget(g, me)
	if second == nil {
		t.Fatal("no prompt for the creature")
	}
	if err := g.ResolvePickTargets(second.ID, me, []game.TargetRef{{Kind: game.TargetCard, ID: creature, Slot: 1}}); err != nil {
		t.Fatalf("pick the creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestHobB9ThorinAttachesAnyNumberOfEquipmentAndTheCreatureDealsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blade := effectstest.SeedEquipment(g, me.ID, "Blade", "")
	spear := effectstest.SeedEquipment(g, me.ID, "Spear", "")
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	victim := effectstest.HobB4Creature(g, opp.ID, "Troll", "Creature — Troll", 4, 5)
	// A Sword the Bear already wears does not "become attached" again.
	hobB9Attach(t, g, spear, bear)

	hobB9CastThorin(t, g, me.ID, []uuid.UUID{blade, spear}, bear)

	for name, id := range map[string]uuid.UUID{"Blade": blade, "Spear": spear} {
		if host := effectstest.AttachmentHostOf(t, g, id); host.ID != bear {
			t.Errorf("the %s is on %s, want the Bear", name, host.ID)
		}
	}
	// One Equipment moved, so the reflexive trigger is waiting for its target.
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the reflexive trigger did not ask for its target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: victim}}); err != nil {
		t.Fatalf("pick the Troll: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The Bear (2/2) wears no power-giving Equipment here, so it deals 2.
	if got := hobB9Damage(g, victim); got != 2 {
		t.Errorf("the Troll has %d damage, want 2 (the Bear's power)", got)
	}
}

func TestHobB9ThorinWithNoEquipmentMovedRaisesNoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Blade", "")
	hobB9Attach(t, g, blade, bear)
	victim := effectstest.HobB4Creature(g, opp.ID, "Troll", "Creature — Troll", 4, 5)

	hobB9CastThorin(t, g, me.ID, []uuid.UUID{blade}, bear)

	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an Equipment already on the creature raised the reflexive trigger")
	}
	if hobB9Damage(g, victim) != 0 {
		t.Error("damage was dealt although nothing became attached")
	}
}

func TestHobB9ThorinWithNoEquipmentChosenAttachesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Blade", "")

	hobB9CastThorin(t, g, me.ID, nil, bear)

	if host := effectstest.AttachmentHostOf(t, g, blade); host.ID != uuid.Nil {
		t.Error("an Equipment that was not chosen was attached")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("choosing no Equipment still raised the reflexive trigger")
	}
}

// --- Inside Information ----------------------------------------------------

const hobB9InsideInformationOracle = "66961bcb-f588-47a7-b3f3-12f4ac674477"

// hobB9InsideInformation seeds the opponent's library (a Wolf, a Shock, a
// Forest, then two more) and casts the spell for X at the opponent.
func hobB9InsideInformation(t *testing.T, g *game.Game, x int) (opp *game.Player, ids []uuid.UUID) {
	t.Helper()
	opp = g.Seats[1]
	ids = effectstest.SetLibraryTopFirst(opp,
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf", ManaCost: "{2}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deeper", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Deepest", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
	)
	effectstest.CastXSpell(t, g, "Inside Information", "Sorcery", hobB9InsideInformationOracle, "{X}{B}{B}", x,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	return opp, ids
}

func hobB9PlayFromExile(g *game.Game, me *game.Player, id uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", AlternativeCost: "inside_information"})
}

func TestHobB9InsideInformationExilesXCardsFromTheOpponentsLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, ids := hobB9InsideInformation(t, g, 3)
	for i, id := range ids {
		if want := i < 3; g.Exile.Contains(id) != want {
			t.Errorf("card %d in exile = %v, want %v", i, !want, want)
		}
	}
	if !opp.Library.Contains(ids[3]) || !opp.Library.Contains(ids[4]) {
		t.Error("cards below the top three left the library")
	}
}

func TestHobB9InsideInformationAtXZeroExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, ids := hobB9InsideInformation(t, g, 0)
	for _, id := range ids {
		if g.Exile.Contains(id) || !opp.Library.Contains(id) {
			t.Fatal("X = 0 exiled a card")
		}
	}
}

func TestHobB9InsideInformationCastsForLifeEqualToManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, ids := hobB9InsideInformation(t, g, 3)
	wolf, shock := ids[0], ids[1]

	if err := g.CastSpell(me.ID, wolf, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("a card exiled this way was cast for its printed mana cost")
	}
	before := me.Life
	if err := hobB9PlayFromExile(g, me, wolf); err != nil {
		t.Fatalf("cast the Wolf for life: %v", err)
	}
	if got := before - me.Life; got != 3 {
		t.Errorf("life paid for the Wolf = %d, want 3 (its mana value) and no mana", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, wolf); !ok || c.Controller != me.ID {
		t.Error("the Wolf should resolve onto the battlefield under the caster's control")
	}
	before = me.Life
	if err := hobB9PlayFromExile(g, me, shock); err != nil {
		t.Fatalf("cast the Shock for life: %v", err)
	}
	if got := before - me.Life; got != 1 {
		t.Errorf("life paid for the Shock = %d, want 1", got)
	}
}

func TestHobB9InsideInformationCannotBeClaimedWithoutTheLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, ids := hobB9InsideInformation(t, g, 3)
	me.Life = 2
	if err := hobB9PlayFromExile(g, me, ids[0]); err == nil {
		t.Error("a mana value 3 spell was cast with 2 life")
	}
	if me.Life != 2 {
		t.Errorf("a refused cast still cost life: %d", me.Life)
	}
}

func TestHobB9InsideInformationPlaysAnExiledLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, ids := hobB9InsideInformation(t, g, 3)
	forest := ids[2]
	before := me.Life
	if err := g.CastSpell(me.ID, forest, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the exiled Forest: %v", err)
	}
	if c, ok := effectstest.BattlefieldCard(g, forest); !ok || c.Controller != me.ID {
		t.Error("the Forest should be on the battlefield under the player's control")
	}
	if me.Life != before {
		t.Errorf("playing a land cost %d life, want none", before-me.Life)
	}
}

func TestHobB9InsideInformationEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, ids := hobB9InsideInformation(t, g, 3)
	seq := g.Turn.Seq
	for i := 0; i < 60 && g.Turn.Seq == seq; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Seq == seq {
		t.Fatal("the turn never ended")
	}
	if err := hobB9PlayFromExile(g, me, ids[0]); err == nil {
		t.Error("a card was played from exile after the turn ended")
	}
	if err := g.CastSpell(me.ID, ids[2], game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("a land was played from exile after the turn ended")
	}
}

// --- Witch-king, Bringer of Ruin -------------------------------------------

func hobB9WitchKingAttack(t *testing.T, g *game.Game) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	witch := effectstest.HobPush(t, g, me.ID, "Witch-king, Bringer of Ruin", "Legendary Creature — Wraith Noble", 5, 3)
	if !effectstest.HasEffectiveKeyword(t, g, witch, "flying") {
		t.Error("the Witch-king has no flying")
	}
	effectstest.DeclareAttack(t, g, opp.ID, witch)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestHobB9WitchKingOffersOnlyTheLeastPowerCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	small1 := effectstest.HobB4Creature(g, opp.ID, "Small One", "Creature — Goblin", 1, 3)
	small2 := effectstest.HobB4Creature(g, opp.ID, "Small Two", "Creature — Goblin", 1, 1)
	big := effectstest.HobB4Creature(g, opp.ID, "Big", "Creature — Giant", 3, 3)

	hobB9WitchKingAttack(t, g)

	prompt := effectstest.SacrificeChoiceFor(g, opp.ID)
	if prompt == nil {
		t.Fatal("the defending player was not asked to sacrifice")
	}
	if !effectstest.HasID(prompt.SacrificeOptions, small1) || !effectstest.HasID(prompt.SacrificeOptions, small2) || effectstest.HasID(prompt.SacrificeOptions, big) {
		t.Errorf("offered %v, want just the two power-1 creatures", prompt.SacrificeOptions)
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, small2)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small2) || !g.Battlefield.Contains(small1) || !g.Battlefield.Contains(big) {
		t.Error("exactly the chosen creature should be sacrificed")
	}
}

func TestHobB9WitchKingTakesTheUniqueLeastPowerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	weak := effectstest.HobB4Creature(g, opp.ID, "Weak", "Creature — Goblin", 0, 4)
	effectstest.HobB4Creature(g, opp.ID, "Strong", "Creature — Giant", 2, 2)
	effectstest.HobB4Creature(g, opp.ID, "Strong Too", "Creature — Giant", 2, 2)

	hobB9WitchKingAttack(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, weak)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(weak) {
		t.Error("the unique least-power creature survived")
	}
}

func TestHobB9WitchKingAsksNobodyWhenTheDefenderHasNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	hobB9WitchKingAttack(t, g)
	if effectstest.SacrificeChoiceFor(g, g.Seats[1].ID) != nil {
		t.Error("a defender with no creatures was asked to sacrifice")
	}
}
