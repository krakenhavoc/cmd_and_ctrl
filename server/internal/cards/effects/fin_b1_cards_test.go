package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b1_cards_test.go — Final Fantasy (FIN), batch 1 (#4462). Each test
// plays the real catalog entry and asserts what its text says.

const (
	finB1DionOracle        = dionOracleID
	finB1FatherOracle      = "87864421-505c-4bdb-b1e3-3e44e05f2beb"
	finB1GrahaOracle       = "0c4603be-d71a-4d33-b62c-04ead1987dbe"
	finB1CatchFishOracle   = sidequestCatchAFishOracleID
	finB1StiltzkinOracle   = "958c2a27-9fb1-42b4-a0ef-296584763a98"
	finB1KnightsOracle     = "73122880-d7f4-422f-a336-18773bcb7bfe"
	finB1VenatOracle       = venatOracleID
	finB1VendorOracle      = "bd0a7a86-a567-4145-b1d8-e0b72091db84"
	finB1EtherOracle       = "566533af-5e67-463f-ac33-dfcf1ec735c9"
	finB1IceFlanOracle     = "3b7120a6-700b-4c61-bea4-a835f0cc0bcb"
	finB1IceMagicOracle    = "88b0bd3b-d832-4059-b0d9-04630917ae94"
	finB1PixieOracle       = "c2c47e13-b581-41a8-a4af-542557a6249a"
	finB1JillOracle        = jillOracleID
	finB1QiqirnOracle      = "560cb970-1c1b-4ba7-8352-bc67b4ea7843"
	finB1EsperOracle       = "9d6fdbd4-ba51-4df1-b383-9b709b36a8d7"
	finB1CardCollectOracle = sidequestCardCollectionOracleID
	finB1CecilOracle       = cecilOracleID
	finB1GreatswordOracle  = "eda09ed4-7509-4d1e-b169-2b7c1c87e936"
	finB1FinalDaysOracle   = "f2b88031-bfb7-46c4-abdd-3b6b5f5acfa4"
	finB1HecteyesOracle    = "7a11b4ed-430d-4ef8-b0c1-6a022f12448c"
)

func TestFinB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finB1DionOracle:               "Dion, Bahamut's Dominant",
		finB1DionOracle + "#1":        "Bahamut, Warden of Light",
		finB1FatherOracle:             "From Father to Son",
		finB1GrahaOracle:              "G'raha Tia",
		finB1CatchFishOracle:          "Sidequest: Catch a Fish",
		finB1CatchFishOracle + "#1":   "Cooking Campsite",
		finB1StiltzkinOracle:          "Stiltzkin, Moogle Merchant",
		finB1KnightsOracle:            "Summon: Knights of Round",
		finB1VenatOracle:              "Venat, Heart of Hydaelyn",
		finB1VenatOracle + "#1":       "Hydaelyn, the Mothercrystal",
		finB1VendorOracle:             "Weapons Vendor",
		finB1EtherOracle:              "Ether",
		finB1IceFlanOracle:            "Ice Flan",
		finB1IceMagicOracle:           "Ice Magic",
		finB1PixieOracle:              "Il Mheg Pixie",
		finB1JillOracle:               "Jill, Shiva's Dominant",
		finB1JillOracle + "#1":        "Shiva, Warden of Ice",
		finB1QiqirnOracle:             "Qiqirn Merchant",
		finB1EsperOracle:              "Retrieve the Esper",
		finB1CardCollectOracle:        "Sidequest: Card Collection",
		finB1CardCollectOracle + "#1": "Magicked Card",
		finB1CecilOracle:              "Cecil, Dark Knight",
		finB1CecilOracle + "#1":       "Cecil, Redeemed Paladin",
		finB1GreatswordOracle:         "Dark Knight's Greatsword",
		finB1FinalDaysOracle:          "The Final Days",
		finB1HecteyesOracle:           "Hecteyes",
	} {
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

// --- fixtures ------------------------------------------------------------------

// finB1Opp is the seat after the active one.
func finB1Opp(g *game.Game) *game.Player { return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)] }

// finB1Cast puts a card in the active seat's hand, floats `n` of `color`
// and casts it strictly, in a main phase.
func finB1Cast(t *testing.T, g *game.Game, c game.Card, color string, n int, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	if err := socB19CastWithPool(g, me, c, color, n, params); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return c.InstanceID
}

// finB1Flashback puts the card in the active seat's graveyard, floats the
// pool and casts it for its flashback cost.
func finB1Flashback(t *testing.T, g *game.Game, c game.Card, color string, n int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	me.Graveyard.PushTop(c)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, color, n)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("flash back %s: %v", c.Name, err)
	}
	return c.InstanceID
}

// finB1Named lists what `owner` controls with that name.
func finB1Named(g *game.Game, owner uuid.UUID, name string) []game.Card {
	return socB19Named(g, owner, name)
}

// --- Il Mheg Pixie ---------------------------------------------------------------

func TestFinB1PixieFliesAndSurveilsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pixie := b12Push(g, me.ID, "Il Mheg Pixie", "Creature — Faerie", finB1PixieOracle, 2, 1)
	if !hasEffectiveKeyword(t, g, pixie, "flying") {
		t.Error("the Pixie has no flying")
	}
	declareAttack(t, g, opp.ID, pixie)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("attacking did not surveil 1")
	}
}

// --- Hecteyes ----------------------------------------------------------------------

func TestFinB1HecteyesMakesEachOpponentDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
			before[p.ID] = p.Hand.Size()
		}
	}
	mine := me.Hand.Size()
	tdCastFromHand(t, g, game.Card{Name: "Hecteyes", TypeLine: "Creature — Ooze Horror", OracleID: finB1HecteyesOracle, Power: 1, Toughness: 1}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	for _, p := range opps {
		discardFromHand(t, g, p.ID)
	}
	passPriorityAroundTable(t, g)
	for _, p := range opps {
		if got := p.Hand.Size(); got != before[p.ID]-1 {
			t.Errorf("%s has %d cards, want %d", p.Name, got, before[p.ID]-1)
		}
	}
	if me.Hand.Size() != mine {
		t.Errorf("you discarded: hand %d, want %d", me.Hand.Size(), mine)
	}
}

// --- Ice Flan ----------------------------------------------------------------------

func TestFinB1IceFlanTapsAndStunsAnOpponentsCreatureOrArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finB1Opp(g)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	tdCastFromHand(t, g, game.Card{Name: "Ice Flan", TypeLine: "Creature — Elemental Ooze", OracleID: finB1IceFlanOracle, Power: 5, Toughness: 4}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, victim) || countersOn(g, victim, game.CounterStun) != 1 {
		t.Errorf("the target: tapped %v, %d stun counters; want tapped with one", tappedNow(g, victim), countersOn(g, victim, game.CounterStun))
	}
}

func TestFinB1IceFlanCannotTargetYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	tdCastFromHand(t, g, game.Card{Name: "Ice Flan", TypeLine: "Creature — Elemental Ooze", OracleID: finB1IceFlanOracle, Power: 5, Toughness: 4}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if tappedNow(g, mine) || countersOn(g, mine, game.CounterStun) != 0 {
		t.Error("your own creature was tapped or stunned")
	}
}

func TestFinB1IceFlanHasIslandcycling(t *testing.T) {
	spec, _ := Lookup(finB1IceFlanOracle)
	want := Typecycling("Islandcycling", "{2}", "an Island card", IsLandWithSubtype("Island")).Label
	if len(spec.Activated) != 1 || spec.Activated[0].Label != want {
		t.Errorf("activated rows %+v, want islandcycling {2}", spec.Activated)
	}
}

// --- Qiqirn Merchant ---------------------------------------------------------------

func TestFinB1QiqirnLootsForOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	merchant := pushCatalogPermanent(g, me.ID, "Qiqirn Merchant", "Creature — Beast Citizen", finB1QiqirnOracle, false)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, merchant, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d after a loot", me.Hand.Size(), hand)
	}
}

func TestFinB1QiqirnSacrificeCostsLessForEachTown(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	merchant := pushCatalogPermanent(g, me.ID, "Qiqirn Merchant", "Creature — Beast Citizen", finB1QiqirnOracle, false)
	for i := 0; i < 2; i++ {
		b12Permanent(g, me.ID, "Town", "Land — Town")
	}
	b12Permanent(g, finB1Opp(g).ID, "Their Town", "Land — Town")
	advanceToMain(t, g)
	// Seven less two Towns of yours is five: four mana is not enough.
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, merchant, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for four mana with two Towns, want it to cost five")
	}
	floatMana(t, g, me, "{C}")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, merchant, 1, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if onBattlefieldNow(g, merchant) {
		t.Error("the Merchant was not sacrificed")
	}
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d, want %d (three cards)", me.Hand.Size(), hand+3)
	}
}

// --- Retrieve the Esper ------------------------------------------------------------

func TestFinB1RetrieveTheEsperFromHandMakesAPlainRobotWarrior(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB1Cast(t, g, game.Card{Name: "Retrieve the Esper", TypeLine: "Sorcery", ManaCost: "{3}{U}", OracleID: finB1EsperOracle}, "U", 4, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	robots := finB1Named(g, me.ID, "Robot Warrior")
	if len(robots) != 1 {
		t.Fatalf("Robot Warriors %d, want 1", len(robots))
	}
	if p, tg := socB14PT(t, g, robots[0].InstanceID); p != 3 || tg != 3 {
		t.Errorf("the token is %d/%d, want 3/3", p, tg)
	}
	if !robots[0].IsArtifact() || !robots[0].IsCreature() {
		t.Errorf("the token is %q, want an artifact creature", robots[0].TypeLine)
	}
}

func TestFinB1RetrieveTheEsperFlashbackAddsTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := finB1Flashback(t, g, game.Card{Name: "Retrieve the Esper", TypeLine: "Sorcery", ManaCost: "{3}{U}", OracleID: finB1EsperOracle}, "U", 6)
	passPriorityAroundTable(t, g)
	robots := finB1Named(g, me.ID, "Robot Warrior")
	if len(robots) != 1 {
		t.Fatalf("Robot Warriors %d, want 1", len(robots))
	}
	if p, tg := socB14PT(t, g, robots[0].InstanceID); p != 5 || tg != 5 {
		t.Errorf("the token is %d/%d, want 5/5 with two +1/+1 counters", p, tg)
	}
	if me.Graveyard.Contains(id) {
		t.Error("the flashed-back spell returned to the graveyard instead of being exiled")
	}
}

// --- The Final Days ----------------------------------------------------------------

func TestFinB1TheFinalDaysMakesTwoTappedHorrors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB1Cast(t, g, game.Card{Name: "The Final Days", TypeLine: "Sorcery", ManaCost: "{2}{B}{B}", OracleID: finB1FinalDaysOracle}, "B", 4, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	horrors := finB1Named(g, me.ID, "Horror")
	if len(horrors) != 2 {
		t.Fatalf("Horrors %d, want 2", len(horrors))
	}
	for _, h := range horrors {
		if !h.Tapped {
			t.Error("a Horror entered untapped")
		}
		if p, tg := socB14PT(t, g, h.InstanceID); p != 2 || tg != 2 {
			t.Errorf("a Horror is %d/%d, want 2/2", p, tg)
		}
	}
}

func TestFinB1TheFinalDaysFlashbackMakesOneHorrorPerCreatureCardInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		pushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	}
	pushCatalogGraveyardCard(me, "Dead Spell", "Instant", "", 0, 0)
	finB1Flashback(t, g, game.Card{Name: "The Final Days", TypeLine: "Sorcery", ManaCost: "{2}{B}{B}", OracleID: finB1FinalDaysOracle}, "B", 6)
	passPriorityAroundTable(t, g)
	horrors := finB1Named(g, me.ID, "Horror")
	if len(horrors) != 3 {
		t.Fatalf("Horrors %d, want 3 (the creature cards in the graveyard)", len(horrors))
	}
	for _, h := range horrors {
		if !h.Tapped {
			t.Error("a Horror entered untapped")
		}
	}
}

// --- Ice Magic ---------------------------------------------------------------------

func finB1IceMagic(t *testing.T, g *game.Game, mode int, color string, n int, target uuid.UUID) error {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	c := game.Card{Name: "Ice Magic", TypeLine: "Instant", ManaCost: "{1}{U}", OracleID: finB1IceMagicOracle,
		InstanceID: uuid.New(), Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(c)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, color, n)
	return g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, Modes: []int{mode}, Targets: cardRefs(target)})
}

func TestFinB1IceMagicBlizzardBouncesForNothingExtra(t *testing.T) {
	g := newCatalogGame(t)
	opp := finB1Opp(g)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := finB1IceMagic(t, g, 0, "U", 2, bear); err != nil {
		t.Fatalf("cast Blizzard for {1}{U}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefieldNow(g, bear) || !inHand(opp, bear) {
		t.Error("Blizzard did not return the creature to its owner's hand")
	}
}

func TestFinB1IceMagicBlizzaraNeedsTwoMoreAndTucksOnTopOrBottom(t *testing.T) {
	g := newCatalogGame(t)
	opp := finB1Opp(g)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := finB1IceMagic(t, g, 1, "U", 2, bear); err == nil {
		t.Fatal("Blizzara cast for the base cost alone")
	}
	if err := finB1IceMagic(t, g, 1, "U", 4, bear); err != nil {
		t.Fatalf("cast Blizzara for {3}{U}: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The OWNER (the opponent) chooses top or bottom, and the creature
	// moves when they have answered.
	c := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatalf("the owner was not asked top or bottom: %+v", g.PendingChoices)
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{bear}, nil); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if onBattlefieldNow(g, bear) {
		t.Fatal("the creature stayed on the battlefield")
	}
}

func TestFinB1IceMagicBlizzagaShufflesTheCreatureIntoItsOwnersLibrary(t *testing.T) {
	g := newCatalogGame(t)
	opp := finB1Opp(g)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	lib := opp.Library.Size()
	if err := finB1IceMagic(t, g, 2, "U", 8, bear); err != nil {
		t.Fatalf("cast Blizzaga for {5}{U}{1}{U}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefieldNow(g, bear) || opp.Library.Size() != lib+1 {
		t.Errorf("battlefield %v, library %d, want gone and %d", onBattlefieldNow(g, bear), opp.Library.Size(), lib+1)
	}
}

// --- From Father to Son ----------------------------------------------------------------

func finB1VehicleOnTop(me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: id, Name: "Test Vehicle", TypeLine: "Artifact — Vehicle", Owner: me.ID, Controller: me.ID})
	return id
}

func TestFinB1FromFatherToSonTutorsAVehicleToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := finB1VehicleOnTop(me)
	finB1Cast(t, g, game.Card{Name: "From Father to Son", TypeLine: "Sorcery", ManaCost: "{1}{W}", OracleID: finB1FatherOracle}, "W", 2, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Test Vehicle")
	passPriorityAroundTable(t, g)
	if !inHand(me, vehicle) || onBattlefieldNow(g, vehicle) {
		t.Error("the Vehicle did not go to your hand from a hand cast")
	}
}

func TestFinB1FromFatherToSonFlashbackPutsTheVehicleOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := finB1VehicleOnTop(me)
	id := finB1Flashback(t, g, game.Card{Name: "From Father to Son", TypeLine: "Sorcery", ManaCost: "{1}{W}", OracleID: finB1FatherOracle}, "W", 7)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Test Vehicle")
	passPriorityAroundTable(t, g)
	if !onBattlefieldNow(g, vehicle) || inHand(me, vehicle) {
		t.Error("the Vehicle did not enter the battlefield from a flashback cast")
	}
	if me.Graveyard.Contains(id) {
		t.Error("the flashed-back spell returned to the graveyard instead of being exiled")
	}
}

// --- G'raha Tia ------------------------------------------------------------------------

func TestFinB1GrahaDrawsOncePerTurnWhenOtherCreaturesAndArtifactsDie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	graha := b12Push(g, me.ID, "G'raha Tia", "Legendary Creature — Cat Archer", finB1GrahaOracle, 3, 5)
	if !hasEffectiveKeyword(t, g, graha, "reach") {
		t.Error("G'raha Tia has no reach")
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	land := b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	// A creature and an artifact dying together is one trigger.
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{bear, rock}) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("drew %d, want 1", got)
	}
	// Nothing else this turn draws: not a second death of yours, not a land
	// or an opponent's creature.
	bear2 := b12Creature(g, me.ID, "Bear 2", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{bear2, land, theirs}) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d in total, want 1 (once each turn)", got)
	}
}

func TestFinB1GrahaIgnoresLandsAndOpponentsPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "G'raha Tia", "Legendary Creature — Cat Archer", finB1GrahaOracle, 3, 5)
	land := b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{land, theirs}) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d, want 0", got)
	}
}

// --- Stiltzkin, Moogle Merchant ---------------------------------------------------------

func TestFinB1StiltzkinHandsOverAPermanentAndDrawsACard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stiltzkin := pushCatalogPermanent(g, me.ID, "Stiltzkin, Moogle Merchant", "Legendary Creature — Moogle", finB1StiltzkinOracle, false)
	if !hasEffectiveKeyword(t, g, stiltzkin, "lifelink") {
		t.Error("Stiltzkin has no lifelink")
	}
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	hand := me.Hand.Size()
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: rock}}
	b16Activate(t, g, me.ID, stiltzkin, 0, game.ActivateAbilityParams{Strict: true, Targets: targets})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, rock)
	if !ok || c.Controller != opp.ID {
		t.Errorf("the Rock is controlled by %v, want the opponent", c.Controller)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestFinB1StiltzkinCannotGiveAwayItselfOrToYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stiltzkin := pushCatalogPermanent(g, me.ID, "Stiltzkin, Moogle Merchant", "Legendary Creature — Moogle", finB1StiltzkinOracle, false)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	for name, targets := range map[string][]game.TargetRef{
		"itself":                {{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: stiltzkin}},
		"a permanent you don't": {{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: theirs}},
		"yourself as the gift":  {{Kind: game.TargetPlayer, ID: me.ID}, {Kind: game.TargetCard, ID: rock}},
	} {
		if err := g.ActivateCatalogAbility(me.ID, stiltzkin, 0, game.ActivateAbilityParams{Strict: true, Targets: targets}); err == nil {
			t.Errorf("activated with %s as a target", name)
		}
	}
}

// --- Weapons Vendor -----------------------------------------------------------------------

func TestFinB1WeaponsVendorDrawsWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tdCastFromHand(t, g, game.Card{Name: "Weapons Vendor", TypeLine: "Creature — Human Artificer", OracleID: finB1VendorOracle, Power: 2, Toughness: 2}, game.CastSpellParams{})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func finB1VendorAtCombat(t *testing.T, withEquipment bool) (g *game.Game, sword, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Weapons Vendor", "Creature — Human Artificer", finB1VendorOracle, false)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if withEquipment {
		sword = b12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	}
	advanceTo(t, g, game.StepBeginCombat)
	b06AddMana(me, "G")
	passPriorityAroundTable(t, g)
	return g, sword, bear
}

func TestFinB1WeaponsVendorPaysOneToAttachAnEquipment(t *testing.T) {
	g, sword, bear := finB1VendorAtCombat(t, true)
	me := g.Seats[0]
	answerPayUnless(t, g, me.ID, true)
	pickCard(t, g, me.ID, sword)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, sword)
	if !ok || !c.IsAttachedTo(bear) {
		t.Errorf("the Sword is not attached to the Bear: %+v", c.AttachedTo)
	}
	if len(poolColors(me)) != 0 {
		t.Errorf("pool = %v after paying {1}", poolColors(me))
	}
}

func TestFinB1WeaponsVendorDecliningAttachesNothing(t *testing.T) {
	g, sword, bear := finB1VendorAtCombat(t, true)
	answerPayUnless(t, g, g.Seats[0].ID, false)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, sword); c.IsAttachedTo(bear) {
		t.Error("the Sword attached without paying")
	}
}

func TestFinB1WeaponsVendorDoesNotTriggerWithoutAnEquipment(t *testing.T) {
	g, _, _ := finB1VendorAtCombat(t, false)
	if hasPayUnlessFor(g, g.Seats[0].ID) {
		t.Error("the Vendor asked for {1} with no Equipment on the battlefield")
	}
}

// --- Ether ---------------------------------------------------------------------------------

func TestFinB1EtherAddsBlueExilesItselfAndCopiesTheNextInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ether := pushCatalogPermanent(g, me.ID, "Ether", "Artifact", finB1EtherOracle, false)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "U", 3)
	activateManaFor(t, g, me.ID, ether, 0, game.ManaAbilityParams{})
	if onBattlefieldNow(g, ether) || me.Graveyard.Contains(ether) {
		t.Fatal("Ether stayed on the battlefield or went to the graveyard; it should be exiled")
	}
	if got := me.ManaPool.Total(); got != 4 {
		t.Fatalf("pool has %d mana, want 4 (three floated plus Ether's {U})", got)
	}
	c := game.Card{Name: "Retrieve the Esper", TypeLine: "Sorcery", ManaCost: "{3}{U}", OracleID: finB1EsperOracle,
		InstanceID: uuid.New(), Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(finB1Named(g, me.ID, "Robot Warrior")); got != 2 {
		t.Errorf("Robot Warriors %d, want 2 (the spell and its copy)", got)
	}
	// Only the next one: a second spell this turn is not copied.
	c2 := game.Card{Name: "Retrieve the Esper", TypeLine: "Sorcery", ManaCost: "{3}{U}", OracleID: finB1EsperOracle,
		InstanceID: uuid.New(), Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(c2)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "U", 4)
	if err := g.CastSpell(me.ID, c2.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast second: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(finB1Named(g, me.ID, "Robot Warrior")); got != 3 {
		t.Errorf("Robot Warriors %d, want 3 (the second spell is not copied)", got)
	}
}

func TestFinB1EtherCopiesEvenIfTheManaGoesElsewhere(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ether := pushCatalogPermanent(g, me.ID, "Ether", "Artifact", finB1EtherOracle, false)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	activateManaFor(t, g, me.ID, ether, 0, game.ManaAbilityParams{})
	// The {U} is still floating; the spell is paid with other mana.
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "U", 4)
	c := game.Card{Name: "Retrieve the Esper", TypeLine: "Sorcery", ManaCost: "{3}{U}", OracleID: finB1EsperOracle,
		InstanceID: uuid.New(), Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(finB1Named(g, me.ID, "Robot Warrior")); got != 2 {
		t.Errorf("Robot Warriors %d, want 2", got)
	}
}

// --- double-faced fixtures -------------------------------------------------------------

// finB1DFC puts a transforming card onto the battlefield with `face` up and
// the given lore counters.
func finB1DFC(g *game.Game, owner uuid.UUID, oracle string, face, lore int, faces ...game.Face) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutTransform,
		Owner: owner, Controller: owner, Faces: faces,
	}
	c.SetFace(face)
	if lore > 0 {
		c.Counters = map[string]int{game.CounterLore: lore}
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

var (
	finB1DionFaces = []game.Face{
		{Name: "Dion, Bahamut's Dominant", TypeLine: "Legendary Creature — Human Noble Knight", ManaCost: "{3}{W}", Colors: []string{"W"}, Power: 3, Toughness: 3},
		{Name: "Bahamut, Warden of Light", TypeLine: "Legendary Enchantment Creature — Saga Dragon", Colors: []string{"W"}, Power: 5, Toughness: 5},
	}
	finB1JillFaces = []game.Face{
		{Name: "Jill, Shiva's Dominant", TypeLine: "Legendary Creature — Human Noble Warrior", ManaCost: "{2}{U}", Colors: []string{"U"}, Power: 2, Toughness: 2},
		{Name: "Shiva, Warden of Ice", TypeLine: "Legendary Enchantment Creature — Saga Elemental", Colors: []string{"U"}, Power: 4, Toughness: 5},
	}
	finB1VenatFaces = []game.Face{
		{Name: "Venat, Heart of Hydaelyn", TypeLine: "Legendary Creature — Elder Wizard", ManaCost: "{1}{W}{W}", Colors: []string{"W"}, Power: 3, Toughness: 3},
		{Name: "Hydaelyn, the Mothercrystal", TypeLine: "Legendary Creature — God", Colors: []string{"W"}, Power: 4, Toughness: 4},
	}
	finB1CecilFaces = []game.Face{
		{Name: "Cecil, Dark Knight", TypeLine: "Legendary Creature — Human Knight", ManaCost: "{B}", Colors: []string{"B"}, Power: 2, Toughness: 3},
		{Name: "Cecil, Redeemed Paladin", TypeLine: "Legendary Creature — Human Knight", Colors: []string{"W"}, Power: 4, Toughness: 4},
	}
	finB1CatchFishFaces = []game.Face{
		{Name: "Sidequest: Catch a Fish", TypeLine: "Enchantment", ManaCost: "{2}{W}", Colors: []string{"W"}},
		{Name: "Cooking Campsite", TypeLine: "Land", Colors: []string{"W"}},
	}
	finB1CardCollectFaces = []game.Face{
		{Name: "Sidequest: Card Collection", TypeLine: "Enchantment", ManaCost: "{3}{U}", Colors: []string{"U"}},
		{Name: "Magicked Card", TypeLine: "Artifact — Vehicle", Colors: []string{"U"}, Power: 4, Toughness: 4},
	}
)

// finB1Seat is the seat `n` places after the active one.
func finB1Seat(g *game.Game, n int) *game.Player { return g.Seats[(g.Turn.ActiveSeat+n)%len(g.Seats)] }

// finB1CardNamed finds a battlefield card by its current name.
func finB1CardNamed(g *game.Game, name string) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}

// --- Dion, Bahamut's Dominant // Bahamut, Warden of Light -----------------------------------

func TestFinB1DionGivesItselfAndKnightsFlyingOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, next := g.Seats[g.Turn.ActiveSeat], finB1Seat(g, 1)
	dion := finB1DFC(g, me.ID, finB1DionOracle, 0, 0, finB1DionFaces...)
	knight := b12Creature(g, me.ID, "Knight", "Creature — Knight", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, next.ID, "Their Knight", "Creature — Knight", 2, 2)
	for id, want := range map[uuid.UUID]bool{dion: true, knight: true, bear: false, theirs: false} {
		if got := hasEffectiveKeyword(t, g, id, "flying"); got != want {
			t.Errorf("on your turn, flying of %s = %v, want %v", id, got, want)
		}
	}
	// On the next seat's turn, none of yours fly.
	nextSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	for g.Turn.ActiveSeat != nextSeat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for _, id := range []uuid.UUID{dion, knight} {
		if hasEffectiveKeyword(t, g, id, "flying") {
			t.Errorf("%s flies on an opponent's turn", id)
		}
	}
}

func TestFinB1DionMakesAKnightWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tdCastFromHand(t, g, game.Card{Name: "Dion, Bahamut's Dominant", TypeLine: "Legendary Creature — Human Noble Knight", OracleID: finB1DionOracle, Power: 3, Toughness: 3}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	knights := finB1Named(g, me.ID, "Knight")
	if len(knights) != 1 {
		t.Fatalf("Knights %d, want 1", len(knights))
	}
	if p, tg := socB14PT(t, g, knights[0].InstanceID); p != 2 || tg != 2 {
		t.Errorf("the Knight is %d/%d, want 2/2", p, tg)
	}
}

func TestFinB1DionTransformsIntoBahamutAtSorcerySpeedForSixMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dion := finB1DFC(g, me.ID, finB1DionOracle, 0, 0, finB1DionFaces...)
	advanceTo(t, g, game.StepBeginCombat)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "W", 6)
	if err := g.ActivateCatalogAbility(me.ID, dion, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed outside a main phase")
	}
	advanceToMain(t, g)
	for g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "W", 6)
	b16Activate(t, g, me.ID, dion, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	bahamut, ok := finB1CardNamed(g, "Bahamut, Warden of Light")
	if !ok {
		t.Fatal("Dion did not come back as Bahamut")
	}
	if bahamut.Counters[game.CounterLore] != 1 {
		t.Errorf("lore %d, want 1", bahamut.Counters[game.CounterLore])
	}
	if !hasEffectiveKeyword(t, g, bahamut.InstanceID, "flying") {
		t.Error("Bahamut has no flying")
	}
}

func TestFinB1BahamutWingsOfLightPumpsAndGrantsFlyingToOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finB1Opp(g)
	saga := finB1DFC(g, me.ID, finB1DionOracle, 1, 0, finB1DionFaces...)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if countersOn(g, bear, game.CounterPlusOne) != 1 || !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Errorf("Bear: %d counters, flying %v; want one counter and flying", countersOn(g, bear, game.CounterPlusOne), hasEffectiveKeyword(t, g, bear, "flying"))
	}
	if countersOn(g, saga, game.CounterPlusOne) != 0 {
		t.Error("Bahamut put a counter on itself")
	}
	if countersOn(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("an opponent's creature got a counter")
	}
}

func TestFinB1BahamutGigaflareDestroysThenReturnsFrontFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finB1Opp(g)
	finB1DFC(g, me.ID, finB1DionOracle, 1, 2, finB1DionFaces...)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	b04WaitForPick(t, g, me.ID)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if onBattlefieldNow(g, victim) {
		t.Error("Gigaflare did not destroy the target")
	}
	if _, ok := finB1CardNamed(g, "Bahamut, Warden of Light"); ok {
		t.Fatal("Bahamut stayed on the battlefield on its back face")
	}
	if _, ok := finB1CardNamed(g, "Dion, Bahamut's Dominant"); !ok {
		t.Fatal("Bahamut did not return front face up")
	}
	// Dion entered again, so it made another Knight.
	if got := len(finB1Named(g, me.ID, "Knight")); got != 1 {
		t.Errorf("Knights %d, want 1 from Dion's enters trigger", got)
	}
}

// --- Jill, Shiva's Dominant // Shiva, Warden of Ice ------------------------------------------

func TestFinB1JillReturnsUpToOneOtherNonlandPermanentWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finB1Opp(g)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tdCastFromHand(t, g, game.Card{Name: "Jill, Shiva's Dominant", TypeLine: "Legendary Creature — Human Noble Warrior", OracleID: finB1JillOracle, Power: 2, Toughness: 2}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if onBattlefieldNow(g, bear) || !inHand(opp, bear) {
		t.Error("Jill did not return the target to its owner's hand")
	}
}

func TestFinB1JillTransformsIntoShivaForFiveMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jill := finB1DFC(g, me.ID, finB1JillOracle, 0, 0, finB1JillFaces...)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "U", 5)
	b16Activate(t, g, me.ID, jill, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	shiva, ok := finB1CardNamed(g, "Shiva, Warden of Ice")
	if !ok || shiva.Counters[game.CounterLore] != 1 {
		t.Fatalf("Jill did not come back as Shiva with a lore counter: %+v", shiva)
	}
}

func TestFinB1ShivaMesmerizeMakesATargetCreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finB1Opp(g)
	finB1DFC(g, me.ID, finB1JillOracle, 1, 0, finB1JillFaces...)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	b04WaitForPick(t, g, me.ID)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, bear); err == nil {
		t.Error("an unblockable creature was blocked")
	}
}

func TestFinB1ShivaColdSnapTapsOpponentsLandsAndReturnsFrontFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finB1Opp(g)
	finB1DFC(g, me.ID, finB1JillOracle, 1, 2, finB1JillFaces...)
	theirLand := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	ourLand := b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, theirLand) {
		t.Error("an opponent's land is untapped")
	}
	if tappedNow(g, ourLand) {
		t.Error("your own land was tapped")
	}
	if _, ok := finB1CardNamed(g, "Jill, Shiva's Dominant"); !ok {
		t.Error("Shiva did not return front face up")
	}
}

// --- Summon: Knights of Round ----------------------------------------------------------------

func TestFinB1KnightsOfRoundMakesThreeKnightsPerChapterOneToFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := finS3CastSaga(t, g, "Summon: Knights of Round", "Enchantment Creature — Saga Knight", finB1KnightsOracle, "{6}{W}{W}", 3, 3)
	passPriorityAroundTable(t, g)
	if got := len(finB1Named(g, me.ID, "Knight")); got != 3 {
		t.Fatalf("Knights %d after chapter I, want 3", got)
	}
	if !hasEffectiveKeyword(t, g, saga, "indestructible") {
		t.Error("the Saga has no indestructible")
	}
}

func TestFinB1KnightsOfRoundChaptersTwoToFourEachMakeThreeMoreKnights(t *testing.T) {
	for chapter := 2; chapter <= 4; chapter++ {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		finS3Saga(g, me.ID, "Summon: Knights of Round", "Enchantment Creature — Saga Knight", finB1KnightsOracle, 3, 3, chapter-1)
		advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
		passPriorityAroundTable(t, g)
		if got := len(finB1Named(g, me.ID, "Knight")); got != 3 {
			t.Errorf("chapter %d: Knights %d, want 3", chapter, got)
		}
	}
}

func TestFinB1KnightsOfRoundChapterFiveBoostsOtherCreaturesAndMakesThemIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finB1Opp(g)
	saga := finS3Saga(g, me.ID, "Summon: Knights of Round", "Enchantment Creature — Saga Knight", finB1KnightsOracle, 3, 3, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if p, tg := socB14PT(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("Bear is %d/%d, want 4/4", p, tg)
	}
	if countersOn(g, bear, game.CounterIndestructible) != 1 || !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("Bear has no indestructible counter")
	}
	if p, tg := socB14PT(t, g, theirs); p != 2 || tg != 2 || countersOn(g, theirs, game.CounterIndestructible) != 0 {
		t.Errorf("an opponent's creature changed: %d/%d", p, tg)
	}
	if countersOn(g, saga, game.CounterIndestructible) != 0 {
		t.Error("the Saga gave itself a counter: it says other creatures")
	}
	if onBattlefieldNow(g, saga) {
		t.Error("the Saga was not sacrificed after chapter V")
	}
}

// --- Venat, Heart of Hydaelyn // Hydaelyn, the Mothercrystal -------------------------------------

func TestFinB1VenatDrawsOnlyOnceEachTurnForLegendarySpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB1DFC(g, me.ID, finB1VenatOracle, 0, 0, finB1VenatFaces...)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	for i := 0; i < 2; i++ {
		c := game.Card{Name: "Legend", TypeLine: "Legendary Creature — Test", InstanceID: uuid.New(), Owner: me.ID, Controller: me.ID, Power: 1, Toughness: 1}
		me.Hand.PushTop(c)
		if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	plain := game.Card{Name: "Plain", TypeLine: "Creature — Test", InstanceID: uuid.New(), Owner: me.ID, Controller: me.ID, Power: 1, Toughness: 1}
	me.Hand.PushTop(plain)
	if err := g.CastSpell(me.ID, plain.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Three cards left the hand (+3 pushed); one was drawn.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1 (one draw from three casts: two legendary, one not)", got)
	}
}

func TestFinB1VenatHerosSunderingExilesANonlandPermanentAndTransforms(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finB1Opp(g)
	venat := finB1DFC(g, me.ID, finB1VenatOracle, 0, 0, finB1VenatFaces...)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "W", 7)
	if err := g.ActivateCatalogAbility(me.ID, venat, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(land)}); err == nil {
		t.Fatal("exiled a land")
	}
	b16Activate(t, g, me.ID, venat, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if onBattlefieldNow(g, bear) || opp.Graveyard.Contains(bear) {
		t.Error("the target was not exiled")
	}
	if _, ok := finB1CardNamed(g, "Hydaelyn, the Mothercrystal"); !ok {
		t.Error("Venat did not transform")
	}
}

func TestFinB1HydaelynBlessingOfLightCountersProtectsAndDrawsForALegend(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hyd := finB1DFC(g, me.ID, finB1VenatOracle, 1, 0, finB1VenatFaces...)
	if !hasEffectiveKeyword(t, g, hyd, "indestructible") {
		t.Error("Hydaelyn is not indestructible")
	}
	legend := b12Creature(g, me.ID, "Legend", "Legendary Creature — Test", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	hand := me.Hand.Size()
	answerPickTarget(t, g, legend)
	passPriorityAroundTable(t, g)
	if countersOn(g, legend, game.CounterPlusOne) != 1 || !hasEffectiveKeyword(t, g, legend, "indestructible") {
		t.Error("the legend got no counter or no indestructible")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1 for a legendary creature", got)
	}
}

func TestFinB1HydaelynDrawsNothingForANonlegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB1DFC(g, me.ID, finB1VenatOracle, 1, 0, finB1VenatFaces...)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	hand := me.Hand.Size()
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if countersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Error("the Bear got no counter")
	}
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d, want 0", got)
	}
}

// --- Sidequest: Catch a Fish // Cooking Campsite ------------------------------------------------

// finB1AtUpkeep puts the Sidequest under the next seat and walks to that
// seat's upkeep, with `top` on top of its library.
func finB1AtUpkeep(t *testing.T, g *game.Game, top game.Card) (p *game.Player, quest, topID uuid.UUID) {
	t.Helper()
	p = finB1Seat(g, 1)
	quest = finB1DFC(g, p.ID, finB1CatchFishOracle, 0, 0, finB1CatchFishFaces...)
	top.InstanceID = uuid.New()
	top.Owner, top.Controller = p.ID, p.ID
	p.Library.PushTop(top)
	advanceTo(t, g, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	return p, quest, top.InstanceID
}

func TestFinB1CatchAFishTakesACreatureMakesFoodAndTransforms(t *testing.T) {
	g := newCatalogGame(t)
	p, quest, top := finB1AtUpkeep(t, g, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	answerConfirmFor(t, g, p, true)
	passPriorityAroundTable(t, g)
	if !inHand(p, top) {
		t.Fatal("the creature card did not go to your hand")
	}
	if got := len(finB1Named(g, p.ID, "Food")); got != 1 {
		t.Errorf("Food tokens %d, want 1", got)
	}
	if c, ok := battlefieldCard(g, quest); !ok || c.Name != "Cooking Campsite" || !c.IsLand() {
		t.Errorf("the enchantment is %q, want Cooking Campsite", c.Name)
	}
}

func TestFinB1CatchAFishAlsoTakesAnArtifactCard(t *testing.T) {
	g := newCatalogGame(t)
	p, _, top := finB1AtUpkeep(t, g, game.Card{Name: "Top Rock", TypeLine: "Artifact"})
	answerConfirmFor(t, g, p, true)
	passPriorityAroundTable(t, g)
	if !inHand(p, top) {
		t.Error("the artifact card did not go to your hand")
	}
}

func TestFinB1CatchAFishDecliningLeavesTheCardAndTheEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	p, quest, top := finB1AtUpkeep(t, g, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	answerConfirmFor(t, g, p, false)
	passPriorityAroundTable(t, g)
	if inHand(p, top) || len(finB1Named(g, p.ID, "Food")) != 0 {
		t.Error("declining still took the card or made Food")
	}
	if c, _ := battlefieldCard(g, quest); c.Name != "Sidequest: Catch a Fish" {
		t.Errorf("the enchantment transformed: %q", c.Name)
	}
}

func TestFinB1CatchAFishIgnoresAOtherCard(t *testing.T) {
	g := newCatalogGame(t)
	p, quest, top := finB1AtUpkeep(t, g, game.Card{Name: "Top Spell", TypeLine: "Sorcery"})
	if confirmChoiceFor(g, p.ID) != nil {
		t.Error("asked about a card that is neither an artifact nor a creature")
	}
	if inHand(p, top) {
		t.Error("took a sorcery")
	}
	if c, _ := battlefieldCard(g, quest); c.Name != "Sidequest: Catch a Fish" {
		t.Errorf("the enchantment transformed: %q", c.Name)
	}
}

func TestFinB1CookingCampsiteTapsForWhiteAndSacrificesAnArtifactForCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	camp := finB1DFC(g, me.ID, finB1CatchFishOracle, 1, 0, finB1CatchFishFaces...)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	activateManaFor(t, g, me.ID, camp, 0, game.ManaAbilityParams{})
	if got := poolColors(me); len(got) != 1 || got[0] != "W" {
		t.Fatalf("pool %v, want one {W}", got)
	}
	// Tapped for mana, so it can't pay the second ability's {T} now.
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	camp2 := finB1DFC(g, me.ID, finB1CatchFishOracle, 1, 0, finB1CatchFishFaces...)
	fillPoolColored(me, "W", 3)
	b16Activate(t, g, me.ID, camp2, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}})
	passPriorityAroundTable(t, g)
	if countersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Errorf("Bear has %d counters, want 1", countersOn(g, bear, game.CounterPlusOne))
	}
	if onBattlefieldNow(g, rock) {
		t.Error("the artifact was not sacrificed")
	}
}

// --- Sidequest: Card Collection // Magicked Card ---------------------------------------------------

func TestFinB1CardCollectionDrawsThreeThenDiscardsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	tdCastFromHand(t, g, game.Card{Name: "Sidequest: Card Collection", TypeLine: "Enchantment", OracleID: finB1CardCollectOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Fatalf("hand grew by %d before the discard, want 3 (draw three)", got)
	}
	if c := discardChoiceFor(g, me.ID); c == nil || c.ChooseMax != 2 {
		t.Fatalf("no discard-two prompt: %+v", c)
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1", got)
	}
}

func finB1CardCollectionAtEndStep(t *testing.T, graveyard int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	quest := finB1DFC(g, me.ID, finB1CardCollectOracle, 0, 0, finB1CardCollectFaces...)
	for i := 0; i < graveyard; i++ {
		pushCatalogGraveyardCard(me, "Junk", "Sorcery", "", 0, 0)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	return g, quest
}

func TestFinB1CardCollectionTransformsAtYourEndStepWithEightCardsInYourGraveyard(t *testing.T) {
	g, quest := finB1CardCollectionAtEndStep(t, 8)
	c, ok := battlefieldCard(g, quest)
	if !ok || c.Name != "Magicked Card" {
		t.Fatalf("the enchantment is %q, want Magicked Card", c.Name)
	}
	if !hasEffectiveKeyword(t, g, quest, "flying") {
		t.Error("Magicked Card has no flying")
	}
	spec, _ := Lookup(finB1CardCollectOracle + "#1")
	if len(spec.Activated) != 1 || spec.Activated[0].Label != "Crew 1" {
		t.Errorf("Magicked Card's abilities: %+v", spec.Activated)
	}
}

func TestFinB1CardCollectionDoesNothingWithSevenCardsInYourGraveyard(t *testing.T) {
	g, quest := finB1CardCollectionAtEndStep(t, 7)
	if c, _ := battlefieldCard(g, quest); c.Name != "Sidequest: Card Collection" {
		t.Errorf("the enchantment transformed with seven cards: %q", c.Name)
	}
}

// --- Cecil, Dark Knight // Cecil, Redeemed Paladin ----------------------------------------------------

func TestFinB1CecilDarknessMakesYouLoseLifeEqualToTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cecil := finB1DFC(g, me.ID, finB1CecilOracle, 0, 0, finB1CecilFaces...)
	if !hasEffectiveKeyword(t, g, cecil, "deathtouch") {
		t.Error("Cecil has no deathtouch")
	}
	advanceToMain(t, g)
	life := me.Life
	dealCombatDamageToPlayer(g, cecil, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	if c, _ := battlefieldCard(g, cecil); c.Name != "Cecil, Dark Knight" {
		t.Errorf("Cecil transformed above half life: %q", c.Name)
	}
}

func TestFinB1CecilUntapsAndTransformsAtHalfYourStartingLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cecil := finB1DFC(g, me.ID, finB1CecilOracle, 0, 0, finB1CecilFaces...)
	advanceToMain(t, g)
	finS3Tap(g, cecil)
	g.WithWriteLock(func() { me.Life = game.StartingLife/2 + 2 })
	dealCombatDamageToPlayer(g, cecil, opp.ID, 2)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, cecil)
	if !ok || c.Name != "Cecil, Redeemed Paladin" {
		t.Fatalf("Cecil is %q at %d life, want the Paladin", c.Name, me.Life)
	}
	if c.Tapped {
		t.Error("Cecil stayed tapped")
	}
	if !hasEffectiveKeyword(t, g, cecil, "lifelink") {
		t.Error("the Paladin has no lifelink")
	}
	// A second damage event after the flip does not turn it back.
	dealCombatDamageToPlayer(g, cecil, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, cecil); c.Name != "Cecil, Redeemed Paladin" {
		t.Errorf("Cecil flipped back: %q", c.Name)
	}
}

func TestFinB1CecilProtectGivesOtherAttackersIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cecil := finB1DFC(g, me.ID, finB1CecilOracle, 1, 0, finB1CecilFaces...)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, cecil, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the other attacker is not indestructible")
	}
	if hasEffectiveKeyword(t, g, idle, "indestructible") || hasEffectiveKeyword(t, g, cecil, "indestructible") {
		t.Error("a non-attacker or Cecil herself is indestructible")
	}
}

// --- Dark Knight's Greatsword ---------------------------------------------------------------------------

func TestFinB1GreatswordMakesAHeroAttachesAndGivesPlusThreeAndKnight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sword := tdCastFromHand(t, g, game.Card{Name: "Dark Knight's Greatsword", TypeLine: "Artifact — Equipment", OracleID: finB1GreatswordOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	heroes := finB1Named(g, me.ID, "Hero")
	if len(heroes) != 1 {
		t.Fatalf("Heroes %d, want 1", len(heroes))
	}
	hero := heroes[0].InstanceID
	if c, _ := battlefieldCard(g, sword); !c.IsAttachedTo(hero) {
		t.Fatal("the Greatsword is not attached to the Hero")
	}
	if p, tg := socB14PT(t, g, hero); p != 4 || tg != 1 {
		t.Errorf("the Hero is %d/%d, want 4/1", p, tg)
	}
	if v := s58p6View(t, g, hero); !v.HasSubtype("Knight") || !v.HasSubtype("Hero") {
		t.Errorf("the Hero's type line is %q, want a Knight in addition to Hero", v.TypeLine)
	}
}

func TestFinB1GreatswordEquipsForThreeLifeOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sword := b12Permanent(g, me.ID, "Dark Knight's Greatsword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == sword {
				g.Battlefield.Cards[i].OracleID = finB1GreatswordOracle
			}
		}
	})
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	life := me.Life
	// The equip ability is the first activated row (job select is a trigger).
	b16Activate(t, g, me.ID, sword, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(a)})
	passPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
	if c, _ := battlefieldCard(g, sword); !c.IsAttachedTo(a) {
		t.Fatal("the Greatsword is not on A")
	}
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(b)}); err == nil {
		t.Error("equipped twice in one turn")
	}
	if me.Life != life-3 {
		t.Errorf("the refused activation still cost life: %d", me.Life)
	}
}
