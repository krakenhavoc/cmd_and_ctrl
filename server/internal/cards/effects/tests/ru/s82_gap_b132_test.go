package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b132_test.go — GAP batch 132 (S82 #5948), The Lord of the Rings:
// Tales of Middle-earth: Mordor Muster; Elven Farsight; Ent's Fury; Assault on
// Osgiliath; Hew the Entwood; Quarrel's End; Rally at the Hornburg; Swarming of
// Moria; Gandalf's Sanction; Arwen's Gift; Surrounded by Orcs; Treason of
// Isengard.

const (
	gapMusterOracle      = "492108b7-f831-4739-925f-c3a419859c52"
	gapFarsightOracle    = "04b35e95-2963-4a0a-bd1c-d516cda2dd37"
	gapEntsFuryOracle    = "6bc3c7fd-e371-467f-973c-4c9f0d1ba0bc"
	gapOsgiliathOracle   = "d2aa50d2-d22d-495e-afc9-23a51c8c8f32"
	gapHewOracle         = "8818bbf8-58ec-4070-a81a-c7ce248057ab"
	gapQuarrelsEndOracle = "fc46649d-9d84-4082-9d2d-fbb7ce91c6a4"
	gapRallyOracle       = "7d3f542a-de64-4de8-9de0-0e25004467f1"
	gapSwarmingOracle    = "c96402dc-1509-4136-bbfd-c86a85e3ee6f"
	gapSanctionOracle    = "6696d2e4-28ee-4ba7-8710-c1b735745270"
	gapArwensGiftOracle  = "13ac12db-d04b-4211-90df-886e1c597e8f"
	gapSurroundedOracle  = "65c5e120-7a3d-475f-9bb2-8ddb5062e4ba"
	gapTreasonOracle     = "b04ef0bd-5e96-45d4-afa8-1f3bc8544e1a"
)

func TestGapB132CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapMusterOracle:      "Mordor Muster",
		gapFarsightOracle:    "Elven Farsight",
		gapEntsFuryOracle:    "Ent's Fury",
		gapOsgiliathOracle:   "Assault on Osgiliath",
		gapHewOracle:         "Hew the Entwood",
		gapQuarrelsEndOracle: "Quarrel's End",
		gapRallyOracle:       "Rally at the Hornburg",
		gapSwarmingOracle:    "Swarming of Moria",
		gapSanctionOracle:    "Gandalf's Sanction",
		gapArwensGiftOracle:  "Arwen's Gift",
		gapSurroundedOracle:  "Surrounded by Orcs",
		gapTreasonOracle:     "Treason of Isengard",
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

// gapB132Sorcery casts a sorcery from the active seat's hand in a main phase,
// strictly, and returns the cast's error.
func gapB132Sorcery(t *testing.T, g *game.Game, name, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", OracleID: oracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	params.Strict = true
	return id, g.CastSpell(me.ID, id, params)
}

func gapB132Library(p *game.Player, cards ...game.Card) {
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
		p.Library.PushTop(c)
	}
}

func gapB132PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := effectstest.LayeredCard(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

// --- Mordor Muster / Swarming of Moria ----------------------------------------

func TestGapMordorMusterDrawsLosesALifeAndAmasses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	if _, err := gapB132Sorcery(t, g, "Mordor Muster", gapMusterOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
	if got := effectstest.CountersOn(g, effectstest.ArmyOf(t, g, me.ID).InstanceID, game.CounterPlusOne); got != 1 {
		t.Errorf("Army counters %d, want 1", got)
	}
}

func TestGapSwarmingOfMoriaMakesATreasureAndAmassesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if _, err := gapB132Sorcery(t, g, "Swarming of Moria", gapSwarmingOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldByName(g, "Treasure") != 1 {
		t.Errorf("Treasures = %d, want 1", effectstest.CountBattlefieldByName(g, "Treasure"))
	}
	if got := effectstest.CountersOn(g, effectstest.ArmyOf(t, g, me.ID).InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("Army counters %d, want 2", got)
	}
}

// --- Elven Farsight -----------------------------------------------------------

func gapB132Farsight(t *testing.T, top game.Card, reveal bool) (hand int, drew bool) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gapB132Library(me, top, game.Card{Name: "Second", TypeLine: "Sorcery"}, game.Card{Name: "Third", TypeLine: "Sorcery"}, game.Card{Name: "Fourth", TypeLine: "Sorcery"})
	if _, err := gapB132Sorcery(t, g, "Elven Farsight", gapFarsightOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	before := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, reveal)
	effectstest.PassPriorityAroundTable(t, g)
	return me.Hand.Size(), me.Hand.Size() > before
}

func TestGapElvenFarsightDrawsWhenARevealedCardIsACreature(t *testing.T) {
	if _, drew := gapB132Farsight(t, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear"}, true); !drew {
		t.Error("a revealed creature card did not draw")
	}
}

func TestGapElvenFarsightDrawsNothingForANonCreatureOrADeclinedReveal(t *testing.T) {
	if _, drew := gapB132Farsight(t, game.Card{Name: "Top Land", TypeLine: "Basic Land — Forest"}, true); drew {
		t.Error("a revealed land card drew")
	}
	if _, drew := gapB132Farsight(t, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear"}, false); drew {
		t.Error("a declined reveal drew")
	}
}

// --- Ent's Fury ---------------------------------------------------------------

func gapB132EntsFury(t *testing.T, power int) (counters int, p, tough int, theirsAlive bool) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Ent", "Creature — Treefolk", power, 6)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 5)
	if _, err := gapB132Sorcery(t, g, "Ent's Fury", gapEntsFuryOracle, game.CastSpellParams{Targets: effectstest.CardRefs(mine, theirs)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p, tough = gapB132PT(t, g, mine)
	return effectstest.CountersOn(g, mine, game.CounterPlusOne), p, tough, effectstest.B12ZoneOf(g, theirs) == game.ZoneBattlefield
}

func TestGapEntsFuryPutsACounterOnlyAtFourPowerThenPumpsAndFights(t *testing.T) {
	// Power 4: a counter, then +1/+1, so 6/8, and it fights the 2/5 and kills it.
	counters, p, tough, alive := gapB132EntsFury(t, 4)
	if counters != 1 || p != 6 || tough != 8 || alive {
		t.Errorf("power 4: counters %d, %d/%d, their creature alive %v; want 1, 6/8, dead", counters, p, tough, alive)
	}
	// Power 3: no counter, +1/+1 only (4/7), and its 4 damage kills a 2/5? No: 5 toughness survives.
	counters, p, tough, alive = gapB132EntsFury(t, 3)
	if counters != 0 || p != 4 || tough != 7 || !alive {
		t.Errorf("power 3: counters %d, %d/%d, their creature alive %v; want 0, 4/7, alive", counters, p, tough, alive)
	}
}

func TestGapEntsFuryRefusesTwoOfYourOwnCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	if _, err := gapB132Sorcery(t, g, "Ent's Fury", gapEntsFuryOracle, game.CastSpellParams{Targets: effectstest.CardRefs(a, b)}); err == nil {
		t.Error("a creature you control was the second target")
	}
}

// --- Assault on Osgiliath -----------------------------------------------------

func TestGapAssaultOnOsgiliathAmassesXThenGrantsGoblinsAndOrcs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	goblin := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Orc", "Creature — Orc", 1, 1)
	if _, err := gapB132Sorcery(t, g, "Assault on Osgiliath", gapOsgiliathOracle, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	army := effectstest.ArmyOf(t, g, me.ID)
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 3 {
		t.Fatalf("Army counters %d, want X = 3", got)
	}
	for _, id := range []uuid.UUID{goblin, army.InstanceID} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "double strike") || !effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
			t.Errorf("a Goblin or Orc of mine lacks double strike or haste")
		}
	}
	for _, id := range []uuid.UUID{bear, theirs} {
		if effectstest.HasKeywordOnBattlefield(t, g, id, "double strike") || effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
			t.Errorf("a creature that is not my Goblin or Orc got the keywords")
		}
	}
}

// --- Hew the Entwood ----------------------------------------------------------

func TestGapHewTheEntwoodPutsChosenCardsInInOrderAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		lands = append(lands, effectstest.PushPermanentForTest(g, me.ID, "Mountain", "", "Basic Land — Mountain"))
	}
	keep := effectstest.PushPermanentForTest(g, me.ID, "Kept Land", "", "Basic Land — Mountain")
	_ = keep
	gapB132Library(me,
		game.Card{Name: "Rock", TypeLine: "Artifact"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Unrevealed", TypeLine: "Sorcery"},
	)
	if _, err := gapB132Sorcery(t, g, "Hew the Entwood", gapHewOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, lands...)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range lands {
		if effectstest.B12ZoneOf(g, id) != game.ZoneGraveyard {
			t.Fatalf("a sacrificed land is in %s", effectstest.B12ZoneOf(g, id))
		}
	}
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to choose among the revealed cards")
	}
	var rock, forest, bear uuid.UUID
	for _, id := range pick.ChooseCards {
		c, _ := g.LookupCardForEffect(id)
		switch c.Name {
		case "Rock":
			rock = id
		case "Forest":
			forest = id
		case "Bear":
			bear = id
		}
	}
	if len(pick.ChooseCards) != 3 || rock == uuid.Nil || forest == uuid.Nil || bear == uuid.Nil {
		t.Fatalf("the three revealed cards are not the candidates: %v", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err == nil {
		t.Fatal("a creature card was chosen")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, rock, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, rock); !ok || c.Tapped {
		t.Errorf("the artifact card: on the battlefield %v, tapped %v; want it untapped on the battlefield", ok, ok && c.Tapped)
	}
	if c, ok := effectstest.BattlefieldCard(g, forest); !ok || !c.Tapped {
		t.Errorf("the land card: on the battlefield %v, tapped %v; want it tapped on the battlefield", ok, ok && !c.Tapped)
	}
	if effectstest.B12ZoneOf(g, bear) != game.ZoneLibrary {
		t.Fatalf("the unchosen revealed card is in %s, want the library", effectstest.B12ZoneOf(g, bear))
	}
	// The unchosen card is on the bottom; the unrevealed one is still on top.
	if me.Library.Cards[0].InstanceID != bear {
		t.Error("the unchosen card is not at the bottom of the library")
	}
	if top, _ := me.Library.Top(); top.Name != "Unrevealed" {
		t.Errorf("the top of the library is %q, want the unrevealed card", top.Name)
	}
}

func TestGapHewTheEntwoodWithNoLandsSacrificedRevealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushPermanentForTest(g, me.ID, "Mountain", "", "Basic Land — Mountain")
	gapB132Library(me, game.Card{Name: "Rock", TypeLine: "Artifact"})
	if _, err := gapB132Sorcery(t, g, "Hew the Entwood", gapHewOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("cards were revealed with no land sacrificed")
	}
	if effectstest.B12ZoneOf(g, land) != game.ZoneBattlefield {
		t.Errorf("the land is in %s", effectstest.B12ZoneOf(g, land))
	}
}

// --- Quarrel's End ------------------------------------------------------------

func TestGapQuarrelsEndDiscardsDrawsTwoAndMakesAHumanSoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	hand := me.Hand.Size()
	_, paid := effectstest.CastWithDiscard(t, g, "Quarrel's End", gapQuarrelsEndOracle, "Sorcery", 1)
	if !me.Graveyard.Contains(paid[0]) {
		t.Fatal("the discard was not paid at announcement")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want the discard and the spell out and two cards drawn", hand, me.Hand.Size())
	}
	var token *game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Human Soldier" && c.IsToken() {
			cc := c
			token = &cc
		}
	}
	if token == nil || token.Power != 1 || token.Toughness != 1 || len(token.Colors) != 1 || token.Colors[0] != "W" {
		t.Errorf("token = %+v, want a 1/1 white Human Soldier", token)
	}
}

func TestGapQuarrelsEndCannotBeCastWithNothingToDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Quarrel's End", TypeLine: "Sorcery", OracleID: gapQuarrelsEndOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("cast with no card to discard")
	}
}

// --- Rally at the Hornburg ----------------------------------------------------

func TestGapRallyAtTheHornburgMakesTwoSoldiersAndGivesHumansHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human Peasant", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if _, err := gapB132Sorcery(t, g, "Rally at the Hornburg", gapRallyOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	soldiers := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Human Soldier" && c.IsToken() {
			soldiers++
			if !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "haste") {
				t.Error("a new Soldier token lacks haste")
			}
		}
	}
	if soldiers != 2 {
		t.Errorf("Soldier tokens = %d, want 2", soldiers)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, human, "haste") {
		t.Error("an existing Human lacks haste")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("a non-Human gained haste")
	}
}

// --- Gandalf's Sanction -------------------------------------------------------

func gapB132Sanction(t *testing.T, spells int, toughness int, marked int) (creatureSurvives bool, controllerLife int) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < spells; i++ {
		kind := "Instant"
		if i%2 == 1 {
			kind = "Sorcery"
		}
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spent", TypeLine: kind, Owner: me.ID, Controller: me.ID})
	}
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Not a spell", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Giant", 2, toughness)
	if marked > 0 {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == theirs {
					g.Battlefield.Cards[i].DamageMarked = marked
				}
			}
		})
	}
	if _, err := gapB132Sorcery(t, g, "Gandalf's Sanction", gapSanctionOracle, game.CastSpellParams{Targets: effectstest.CardRefs(theirs)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return effectstest.B12ZoneOf(g, theirs) == game.ZoneBattlefield, opp.Life
}

func TestGapGandalfsSanctionSendsExcessDamageToTheCreaturesController(t *testing.T) {
	// 5 instants and sorceries against a 3-toughness creature: 3 lethal, 2 excess.
	survives, life := gapB132Sanction(t, 5, 3, 0)
	if survives || life != 40-2 {
		t.Errorf("5 spells vs toughness 3: survives %v, life %d; want dead and 38", survives, life)
	}
	// Damage already marked counts: 1 needed, so 4 excess.
	survives, life = gapB132Sanction(t, 5, 3, 2)
	if survives || life != 40-4 {
		t.Errorf("5 spells vs toughness 3 with 2 marked: survives %v, life %d; want dead and 36", survives, life)
	}
	// Exactly lethal: no excess.
	survives, life = gapB132Sanction(t, 3, 3, 0)
	if survives || life != 40 {
		t.Errorf("3 spells vs toughness 3: survives %v, life %d; want dead and 40", survives, life)
	}
	// Not enough: all of it to the creature, which survives.
	survives, life = gapB132Sanction(t, 2, 5, 0)
	if !survives || life != 40 {
		t.Errorf("2 spells vs toughness 5: survives %v, life %d; want alive and 40", survives, life)
	}
}

func TestGapGandalfsSanctionWithNoInstantsOrSorceriesDealsNothing(t *testing.T) {
	survives, life := gapB132Sanction(t, 0, 1, 0)
	if !survives || life != 40 {
		t.Errorf("no spells: survives %v, life %d; want alive and 40", survives, life)
	}
}

// --- Arwen's Gift -------------------------------------------------------------

func TestGapArwensGiftCostsOneLessWithTwoLegendaryCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Arwen's Gift", TypeLine: "Sorcery", OracleID: gapArwensGiftOracle,
		ManaCost: "{3}{U}", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	price := func() string {
		card, _ := g.LookupCardForEffect(id)
		p, err := g.PriceCast(me.ID, card, game.CastSpellParams{})
		if err != nil {
			t.Fatalf("PriceSpell: %v", err)
		}
		return p.Total.String()
	}
	if got := price(); got != "{3}{U}" {
		t.Errorf("with no legendary creature: %s, want {3}{U}", got)
	}
	effectstest.B12Creature(g, me.ID, "Legend One", "Legendary Creature — Human", 1, 1)
	if got := price(); got != "{3}{U}" {
		t.Errorf("with one legendary creature: %s, want {3}{U}", got)
	}
	effectstest.B12Creature(g, me.ID, "Legend Two", "Legendary Creature — Human", 1, 1)
	if got := price(); got != "{2}{U}" {
		t.Errorf("with two legendary creatures: %s, want {2}{U}", got)
	}
}

func TestGapArwensGiftScriesTwoThenDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gapB132Library(me, game.Card{Name: "One", TypeLine: "Sorcery"}, game.Card{Name: "Two", TypeLine: "Sorcery"}, game.Card{Name: "Three", TypeLine: "Sorcery"})
	if _, err := gapB132Sorcery(t, g, "Arwen's Gift", gapArwensGiftOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want two cards drawn after the scry", hand, me.Hand.Size())
	}
}

// --- Surrounded by Orcs -------------------------------------------------------

func TestGapSurroundedByOrcsMillsXWhereXIsTheArmysPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	// An Army with 2 counters already: amass 3 makes it 5/5, so X = 5.
	army := effectstest.B12Creature(g, me.ID, "Orc Army", "Token Creature — Orc Army", 2, 2)
	before := opp.Library.Size()
	if _, err := gapB132Sorcery(t, g, "Surrounded by Orcs", gapSurroundedOracle, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, army, game.CounterPlusOne); got != 3 {
		t.Errorf("Army counters %d, want 3", got)
	}
	if got := before - opp.Library.Size(); got != 5 {
		t.Errorf("the opponent milled %d, want 5 (the Army's power)", got)
	}
}

// --- Treason of Isengard ------------------------------------------------------

func TestGapTreasonOfIsengardPutsAnInstantOrSorceryOnTopAndAmasses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if _, err := gapB132Sorcery(t, g, "Treason of Isengard", gapTreasonOracle, game.CastSpellParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("a creature card was a legal target")
	}
	if _, err := gapB132Sorcery(t, g, "Treason of Isengard", gapTreasonOracle, game.CastSpellParams{Targets: effectstest.CardRefs(spell)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if top, _ := me.Library.Top(); top.InstanceID != spell {
		t.Error("the instant card is not on top of the library")
	}
	if got := effectstest.CountersOn(g, effectstest.ArmyOf(t, g, me.ID).InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("Army counters %d, want 2", got)
	}
}

func TestGapTreasonOfIsengardResolvesWithNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if _, err := gapB132Sorcery(t, g, "Treason of Isengard", gapTreasonOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("a zero-target cast was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, effectstest.ArmyOf(t, g, me.ID).InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("Army counters %d, want 2", got)
	}
}
