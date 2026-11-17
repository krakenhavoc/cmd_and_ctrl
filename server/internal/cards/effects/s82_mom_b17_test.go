package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b17_test.go — March of the Machine Sieges, batch 17 (#5477): both
// faces of ten double-faced cards, driven through the engine.

type momB17Face struct {
	key, name string
	keywords  []string
}

var momB17Faces = []momB17Face{
	{invasionOfDominariaOracleID, "Invasion of Dominaria", nil},
	{invasionOfDominariaOracleID + "#1", "Serra Faithkeeper", []string{"flying", "vigilance"}},
	{invasionOfEldraineOracleID, "Invasion of Eldraine", nil},
	{invasionOfEldraineOracleID + "#1", "Prickle Faeries", []string{"flying"}},
	{invasionOfErgamonOracleID, "Invasion of Ergamon", nil},
	{invasionOfErgamonOracleID + "#1", "Truga Cliffcharger", []string{"trample"}},
	{invasionOfKamigawaOracleID, "Invasion of Kamigawa", nil},
	{invasionOfKamigawaOracleID + "#1", "Rooftop Saboteurs", []string{"flying"}},
	{invasionOfMoagOracleID, "Invasion of Moag", nil},
	{invasionOfMoagOracleID + "#1", "Bloomwielder Dryads", nil},
	{invasionOfUlgrothaOracleID, "Invasion of Ulgrotha", nil},
	{invasionOfUlgrothaOracleID + "#1", "Grandmother Ravi Sengir", []string{"flying"}},
	{invasionOfXerexOracleID, "Invasion of Xerex", nil},
	{invasionOfXerexOracleID + "#1", "Vertex Paladin", []string{"flying"}},
	{invasionOfZendikarOracleID, "Invasion of Zendikar", nil},
	{invasionOfZendikarOracleID + "#1", "Awakened Skyclave", []string{"vigilance", "haste"}},
	{invasionOfAzgolOracleID, "Invasion of Azgol", nil},
	{invasionOfAzgolOracleID + "#1", "Ashen Reaper", []string{"menace"}},
	{invasionOfLorwynOracleID, "Invasion of Lorwyn", nil},
	{invasionOfLorwynOracleID + "#1", "Winnowing Forces", nil},
}

func TestMomB17EveryFaceIsRegisteredFull(t *testing.T) {
	for _, f := range momB17Faces {
		spec, ok := Lookup(f.key)
		if !ok {
			t.Errorf("%s is not registered", f.name)
			continue
		}
		if spec.Name != f.name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", f.name, spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(f.keywords) != len(spec.PrintedKeywords) {
			t.Errorf("%s: keywords %v, want %v", f.name, spec.PrintedKeywords, f.keywords)
			continue
		}
		for i, kw := range f.keywords {
			if spec.PrintedKeywords[i] != kw {
				t.Errorf("%s: keywords %v, want %v", f.name, spec.PrintedKeywords, f.keywords)
			}
		}
	}
}

// momB17Row is a Siege row with its two faces, the back in the front's colour.
func momB17Row(oracleID, front, back, cost, defense, backType, power, toughness string) cards.Card {
	return siegeRow(oracleID, front, back, cost, defense, backType, power, toughness)
}

// momB17Enter puts the Siege onto the battlefield from the active seat's hand
// and names the next seat as its protector. The ETB trigger is left for the
// caller to answer.
func momB17Enter(t *testing.T, g *game.Game, row cards.Card) (owner, protector *game.Player, id uuid.UUID) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	owner = g.Seats[seat]
	protector = g.Seats[(seat+1)%len(g.Seats)]
	id = importToHand(row, owner)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("move battle to battlefield: %v", err)
	}
	answerProtectorPrompt(t, g, owner.ID, protector.ID)
	return owner, protector, id
}

func momB17PickPlayer(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoicePickTarget {
			continue
		}
		if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		return
	}
	t.Fatal("no pick_target prompt was queued")
}

func momB17PickCard(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoicePickTarget {
			continue
		}
		if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		return
	}
	t.Fatal("no pick_target prompt was queued")
}

func momB17Counters(g *game.Game, id uuid.UUID, kind string) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Counters[kind]
		}
	}
	return -1
}

// --- Invasion of Dominaria // Serra Faithkeeper ---------------------------------

func TestInvasionOfDominariaGainsFourAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	life, hand := owner.Life, owner.Hand.Size()
	momB17Enter(t, g, momB17Row(invasionOfDominariaOracleID, "Invasion of Dominaria", "Serra Faithkeeper", "{2}{W}", "5", "Creature — Angel", "4", "4"))
	passPriorityAroundTable(t, g)
	if owner.Life != life+4 {
		t.Errorf("life = %d, want %d", owner.Life, life+4)
	}
	if owner.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d (the Siege left it, one card drawn)", owner.Hand.Size(), hand+1)
	}
}

func TestSerraFaithkeeperHasFlyingAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	id := pushCatalogPermanent(g, g.Seats[0].ID, "Serra Faithkeeper", "Creature — Angel", invasionOfDominariaOracleID+"#1", false)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

// --- Invasion of Eldraine // Prickle Faeries ------------------------------------

func TestInvasionOfEldraineMakesTheTargetOpponentDiscardTwo(t *testing.T) {
	g := newCatalogGame(t)
	_, _, _ = momB17Enter(t, g, momB17Row(invasionOfEldraineOracleID, "Invasion of Eldraine", "Prickle Faeries", "{3}{B}", "4", "Creature — Faerie", "2", "2"))
	victim := g.Seats[2]
	hand := victim.Hand.Size()
	momB17PickPlayer(t, g, victim.ID)
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, victim.ID); got != 2 {
		t.Fatalf("discard owed = %d, want 2", got)
	}
	answerDiscard(t, g, victim.ID, victim.Hand.Cards[0].InstanceID, victim.Hand.Cards[1].InstanceID)
	if victim.Hand.Size() != hand-2 {
		t.Errorf("hand = %d, want %d", victim.Hand.Size(), hand-2)
	}
	if g.Seats[1].Hand.Size() != 7 {
		t.Errorf("another opponent's hand changed to %d", g.Seats[1].Hand.Size())
	}
}

func TestPrickleFaeriesPingsAnOpponentWithTwoOrFewerCardsInHand(t *testing.T) {
	g := newCatalogGame(t)
	faeries := pushCatalogPermanent(g, g.Seats[0].ID, "Prickle Faeries", "Creature — Faerie", invasionOfEldraineOracleID+"#1", false)
	opp := g.Seats[1]
	opp.Hand.Cards = opp.Hand.Cards[:2]
	life := opp.Life
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, faeries) == nil {
		t.Fatal("no upkeep trigger with two cards in hand")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life = %d, want %d", opp.Life, life-2)
	}
}

func TestPrickleFaeriesIgnoresAnOpponentWithThreeCards(t *testing.T) {
	g := newCatalogGame(t)
	faeries := pushCatalogPermanent(g, g.Seats[0].ID, "Prickle Faeries", "Creature — Faerie", invasionOfEldraineOracleID+"#1", false)
	opp := g.Seats[1]
	opp.Hand.Cards = opp.Hand.Cards[:3]
	life := opp.Life
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, faeries) != nil {
		t.Fatal("triggered with three cards in hand")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life = %d, want %d", opp.Life, life)
	}
}

func TestPrickleFaeriesIgnoresItsControllersUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	faeries := pushCatalogPermanent(g, g.Seats[1].ID, "Prickle Faeries", "Creature — Faerie", invasionOfEldraineOracleID+"#1", false)
	g.Seats[1].Hand.Cards = g.Seats[1].Hand.Cards[:0]
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, faeries) != nil {
		t.Fatal("triggered on its controller's own upkeep")
	}
}

// --- Invasion of Ergamon // Truga Cliffcharger ----------------------------------

func ergamonRow() cards.Card {
	return momB17Row(invasionOfErgamonOracleID, "Invasion of Ergamon", "Truga Cliffcharger", "{R}{G}", "5", "Creature — Rhino", "3", "4")
}

func TestInvasionOfErgamonMakesATreasureAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	owner, _, _ := momB17Enter(t, g, ergamonRow())
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Treasure"); n != 1 {
		t.Fatalf("Treasures = %d, want 1", n)
	}
	hand := owner.Hand.Size()
	answerDiscard(t, g, owner.ID, owner.Hand.Cards[0].InstanceID)
	if owner.Hand.Size() != hand {
		t.Errorf("hand = %d, want %d (discarded one, drew one)", owner.Hand.Size(), hand)
	}
}

func TestInvasionOfErgamonDrawsNothingWhenYouDeclineToDiscard(t *testing.T) {
	g := newCatalogGame(t)
	owner, _, _ := momB17Enter(t, g, ergamonRow())
	passPriorityAroundTable(t, g)
	hand := owner.Hand.Size()
	answerDiscard(t, g, owner.ID)
	if owner.Hand.Size() != hand {
		t.Errorf("hand = %d, want %d", owner.Hand.Size(), hand)
	}
	if n := countBattlefieldByName(g, "Treasure"); n != 1 {
		t.Errorf("Treasures = %d, want 1", n)
	}
}

// Truga Cliffcharger, cast transformed from the defeated Siege: it discards a
// card and fetches a land or battle card, and only if a card was discarded.
func trugaSetup(t *testing.T) (g *game.Game, owner *game.Player, forest uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	owner = g.Seats[g.Turn.ActiveSeat]
	forest = uuid.New()
	owner.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner.ID, Controller: owner.ID})
	_, _, id := momB17Enter(t, g, ergamonRow())
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, owner.ID)
	defeatBattle(t, g, id, 5)
	toMainPhaseSameTurn(t, g)
	if err := g.CastSpell(owner.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast Truga Cliffcharger: %v", err)
	}
	passPriorityAroundTable(t, g)
	return g, owner, forest
}

func TestTrugaCliffchargerDiscardsToFetchALand(t *testing.T) {
	g, owner, forest := trugaSetup(t)
	answerDiscard(t, g, owner.ID, owner.Hand.Cards[0].InstanceID)
	search := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, owner.ID)
	if search == nil {
		t.Fatal("no search prompt after discarding")
	}
	if err := g.ResolveSearchLibrary(search.ID, owner.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("pick the land: %v", err)
	}
	g.SettleResolution()
	found := false
	for _, c := range owner.Hand.Cards {
		if c.InstanceID == forest {
			found = true
		}
	}
	if !found {
		t.Error("the Forest is not in hand")
	}
}

func TestTrugaCliffchargerFetchesNothingWithoutADiscard(t *testing.T) {
	g, owner, forest := trugaSetup(t)
	answerDiscard(t, g, owner.ID)
	passPriorityAroundTable(t, g)
	for _, c := range owner.Hand.Cards {
		if c.InstanceID == forest {
			t.Error("the Forest was fetched without a discard")
		}
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, owner.ID) != nil {
		t.Error("a search prompt opened though nothing was discarded")
	}
}

// --- Invasion of Kamigawa // Rooftop Saboteurs ----------------------------------

func TestInvasionOfKamigawaTapsAndStunsTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[2]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, g.Seats[g.Turn.ActiveSeat].ID, "Mine", "Creature — Bear", 2, 2)
	momB17Enter(t, g, momB17Row(invasionOfKamigawaOracleID, "Invasion of Kamigawa", "Rooftop Saboteurs", "{3}{U}", "4", "Creature — Moonfolk Ninja", "2", "3"))
	momB17PickCard(t, g, bear)
	passPriorityAroundTable(t, g)
	c := battlefieldCardFor(g, bear)
	if c == nil || !c.Tapped {
		t.Error("the target is not tapped")
	}
	if n := momB17Counters(g, bear, game.CounterStun); n != 1 {
		t.Errorf("stun counters = %d, want 1", n)
	}
	if n := momB17Counters(g, mine, game.CounterStun); n != 0 {
		t.Errorf("my own creature got %d stun counters", n)
	}
}

func TestRooftopSaboteursDrawOnCombatDamageToAPlayerOrBattle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	saboteurs := pushCatalogPermanent(g, me.ID, "Rooftop Saboteurs", "Creature — Moonfolk Ninja", invasionOfKamigawaOracleID+"#1", false)
	battle := pushCatalogPermanent(g, g.Seats[1].ID, "Some Siege", "Battle — Siege", "", false)
	if !hasKeywordOnBattlefield(t, g, saboteurs, "flying") {
		t.Error("no flying")
	}
	deal := func(target uuid.UUID, combat bool) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: saboteurs, Target: target, Amount: 2, Combat: combat})
		})
	}
	hand := me.Hand.Size()
	deal(g.Seats[1].ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("noncombat damage to a player triggered")
	}
	deal(g.Seats[1].ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatal("combat damage to a player did not draw")
	}
	deal(battle, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand = %d, want %d (a card for the player, one for the battle)", me.Hand.Size(), hand+2)
	}
}

// --- Invasion of Moag // Bloomwielder Dryads ------------------------------------

func TestInvasionOfMoagPutsACounterOnEachCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	_, _, siege := momB17Enter(t, g, momB17Row(invasionOfMoagOracleID, "Invasion of Moag", "Bloomwielder Dryads", "{2}{G}{W}", "5", "Creature — Dryad", "3", "3"))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := momB17Counters(g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("my creature has %d counters, want 1", n)
		}
	}
	if n := momB17Counters(g, theirs, game.CounterPlusOne); n != 0 {
		t.Errorf("their creature has %d counters, want 0", n)
	}
	if n := momB17Counters(g, siege, game.CounterPlusOne); n > 0 {
		t.Errorf("the Siege, not a creature, got %d counters", n)
	}
}

func TestBloomwielderDryadsPutsACounterOnTargetCreatureAtYourEnd(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Bloomwielder Dryads", "Creature — Dryad", invasionOfMoagOracleID+"#1", false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToEndStep(t, g)
	momB17PickCard(t, g, bear)
	passPriorityAroundTable(t, g)
	if n := momB17Counters(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("counters on the target = %d, want 1", n)
	}
}

func TestBloomwielderDryadsAsksOpponentsToPayForWard(t *testing.T) {
	spec, _ := Lookup(invasionOfMoagOracleID + "#1")
	found := false
	for _, tr := range spec.Triggered {
		if tr.Key == "Bloomwielder Dryads — ward {2}" {
			found = true
		}
	}
	if !found {
		t.Error("no ward {2} trigger")
	}
}

// --- Invasion of Ulgrotha // Grandmother Ravi Sengir ----------------------------

func TestInvasionOfUlgrothaDamagesAnyOtherTargetAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[2]
	life, mine := victim.Life, me.Life
	momB17Enter(t, g, momB17Row(invasionOfUlgrothaOracleID, "Invasion of Ulgrotha", "Grandmother Ravi Sengir", "{4}{B}", "5", "Legendary Creature — Human Wizard", "3", "3"))
	momB17PickPlayer(t, g, victim.ID)
	passPriorityAroundTable(t, g)
	if victim.Life != life-3 {
		t.Errorf("victim life = %d, want %d", victim.Life, life-3)
	}
	if me.Life != mine+3 {
		t.Errorf("my life = %d, want %d", me.Life, mine+3)
	}
}

func TestInvasionOfUlgrothaCannotDamageItself(t *testing.T) {
	g := newCatalogGame(t)
	_, _, siege := momB17Enter(t, g, momB17Row(invasionOfUlgrothaOracleID, "Invasion of Ulgrotha", "Grandmother Ravi Sengir", "{4}{B}", "5", "Legendary Creature — Human Wizard", "3", "3"))
	momB17PickCard(t, g, siege)
	passPriorityAroundTable(t, g)
	if n := momB17Counters(g, siege, game.CounterDefense); n != 5 {
		t.Errorf("defense = %d, want the untouched 5", n)
	}
}

func TestGrandmotherRaviSengirGrowsWhenAnOpponentsCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ravi := pushCatalogPermanent(g, me.ID, "Grandmother Ravi Sengir", "Legendary Creature — Human Wizard", invasionOfUlgrothaOracleID+"#1", false)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 1, 1)
	life := me.Life
	kill := func(id uuid.UUID) {
		g.WithWriteLock(func() {
			if err := g.DealDamageToCreatureForEffect(uuid.Nil, id, 5); err != nil {
				t.Fatal(err)
			}
		})
		g.RunStateChecksForTest()
		passPriorityAroundTable(t, g)
	}
	kill(mine)
	if n := momB17Counters(g, ravi, game.CounterPlusOne); n != 0 || me.Life != life {
		t.Fatalf("my own creature dying gave %d counters and %d life", n, me.Life-life)
	}
	kill(theirs)
	if n := momB17Counters(g, ravi, game.CounterPlusOne); n != 1 {
		t.Errorf("counters = %d, want 1", n)
	}
	if me.Life != life+1 {
		t.Errorf("life = %d, want %d", me.Life, life+1)
	}
}

// --- Invasion of Xerex // Vertex Paladin ----------------------------------------

func TestInvasionOfXerexReturnsTheTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	hand := opp.Hand.Size()
	momB17Enter(t, g, momB17Row(invasionOfXerexOracleID, "Invasion of Xerex", "Vertex Paladin", "{2}{W}{U}", "4", "Creature — Angel Knight", "*", "*"))
	momB17PickCard(t, g, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature is still on the battlefield")
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d", opp.Hand.Size(), hand+1)
	}
}

func TestInvasionOfXerexMayTargetNothing(t *testing.T) {
	g := newCatalogGame(t)
	_, _, siege := momB17Enter(t, g, momB17Row(invasionOfXerexOracleID, "Invasion of Xerex", "Vertex Paladin", "{2}{W}{U}", "4", "Creature — Angel Knight", "*", "*"))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(siege) {
		t.Error("the Siege left the battlefield")
	}
}

func TestVertexPaladinCountsTheCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paladin := pushCatalogPermanent(g, me.ID, "Vertex Paladin", "Creature — Angel Knight", invasionOfXerexOracleID+"#1", false)
	if p, tough := ptOf(t, g, paladin); p != 1 || tough != 1 {
		t.Fatalf("alone: %d/%d, want 1/1", p, tough)
	}
	b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	if p, tough := ptOf(t, g, paladin); p != 3 || tough != 3 {
		t.Errorf("with two more of mine: %d/%d, want 3/3", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, paladin, "flying") {
		t.Error("no flying")
	}
}

// --- Invasion of Zendikar // Awakened Skyclave ----------------------------------

func TestInvasionOfZendikarFetchesTwoBasicLandsTapped(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	var forests []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		forests = append(forests, id)
		owner.Library.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner.ID, Controller: owner.ID})
	}
	owner.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Nonbasic", TypeLine: "Land", Owner: owner.ID, Controller: owner.ID})
	momB17Enter(t, g, momB17Row(invasionOfZendikarOracleID, "Invasion of Zendikar", "Awakened Skyclave", "{3}{G}", "3", "Creature — Elemental", "4", "4"))
	passPriorityAroundTable(t, g)
	search := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, owner.ID)
	if search == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(search.ID, owner.ID, forests[:2]); err != nil {
		t.Fatalf("pick lands: %v", err)
	}
	g.SettleResolution()
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" {
			n++
			if !c.Tapped {
				t.Error("a fetched land entered untapped")
			}
		}
		if c.Name == "Nonbasic" {
			t.Error("fetched a nonbasic land")
		}
	}
	if n != 2 {
		t.Errorf("Forests on the battlefield = %d, want 2", n)
	}
}

func TestAwakenedSkyclaveIsALandCreatureThatTapsForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sky := pushCatalogPermanent(g, me.ID, "Awakened Skyclave", "Creature — Elemental", invasionOfZendikarOracleID+"#1", false)
	c := layeredCard(t, g, sky)
	if !c.IsLand() || !c.IsCreature() {
		t.Errorf("%q is not a land creature", c.TypeLine)
	}
	for _, kw := range []string{"vigilance", "haste"} {
		if !hasKeywordOnBattlefield(t, g, sky, kw) {
			t.Errorf("no %s", kw)
		}
	}
	if err := g.ActivateManaAbility(me.ID, sky, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool = %d, want 1", me.ManaPool.Total())
	}
}

// --- Invasion of Azgol // Ashen Reaper ------------------------------------------

func azgolRow() cards.Card {
	return momB17Row(invasionOfAzgolOracleID, "Invasion of Azgol", "Ashen Reaper", "{B}{R}", "4", "Creature — Zombie Elemental", "2", "1")
}

func TestInvasionOfAzgolMakesThemSacrificeAndLoseALife(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[2]
	keep := b12Creature(g, victim.ID, "Keep", "Creature — Bear", 2, 2)
	lose := b12Creature(g, victim.ID, "Lose", "Creature — Bear", 2, 2)
	life := victim.Life
	momB17Enter(t, g, azgolRow())
	momB17PickPlayer(t, g, victim.ID)
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, victim.ID)
	if c == nil {
		t.Fatal("the victim was not asked what to sacrifice")
	}
	if err := g.ResolveChooseCards(c.ID, victim.ID, []uuid.UUID{lose}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lose) || !g.Battlefield.Contains(keep) {
		t.Error("the wrong creature was sacrificed")
	}
	if victim.Life != life-1 {
		t.Errorf("life = %d, want %d", victim.Life, life-1)
	}
}

func TestInvasionOfAzgolStillCostsALifeWithNothingToSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[2]
	life := victim.Life
	momB17Enter(t, g, azgolRow())
	momB17PickPlayer(t, g, victim.ID)
	passPriorityAroundTable(t, g)
	if victim.Life != life-1 {
		t.Errorf("life = %d, want %d", victim.Life, life-1)
	}
}

func TestAshenReaperGrowsOnlyIfAPermanentWentToAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	reaper := pushCatalogPermanent(g, me.ID, "Ashen Reaper", "Creature — Zombie Elemental", invasionOfAzgolOracleID+"#1", false)
	advanceToEndStep(t, g)
	if triggerOnStack(g, reaper) != nil {
		t.Fatal("triggered with nothing in a graveyard")
	}
	passPriorityAroundTable(t, g)
	if n := momB17Counters(g, reaper, game.CounterPlusOne); n != 0 {
		t.Fatalf("counters = %d, want 0", n)
	}
}

func TestAshenReaperGrowsAfterAPermanentDiesThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	reaper := pushCatalogPermanent(g, me.ID, "Ashen Reaper", "Creature — Zombie Elemental", invasionOfAzgolOracleID+"#1", false)
	victim := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, victim, 5); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n := momB17Counters(g, reaper, game.CounterPlusOne); n != 1 {
		t.Errorf("counters = %d, want 1", n)
	}
}

// --- Invasion of Lorwyn // Winnowing Forces -------------------------------------

func lorwynRow() cards.Card {
	return momB17Row(invasionOfLorwynOracleID, "Invasion of Lorwyn", "Winnowing Forces", "{4}{B}{G}", "5", "Creature — Elf Warrior", "*", "*")
}

func lorwynLands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		b12Permanent(g, owner, "Forest", "Basic Land — Forest")
	}
}

func TestInvasionOfLorwynDestroysANonElfWithPowerAtMostYourLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lorwynLands(g, me.ID, 3)
	small := b12Creature(g, g.Seats[1].ID, "Small", "Creature — Bear", 3, 3)
	big := b12Creature(g, g.Seats[1].ID, "Big", "Creature — Bear", 4, 4)
	elf := b12Creature(g, g.Seats[1].ID, "Elf", "Creature — Elf Warrior", 1, 1)
	momB17Enter(t, g, lorwynRow())
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			for _, id := range []uuid.UUID{big, elf} {
				for _, off := range c.PickTargetCards {
					if off == id {
						t.Errorf("offered an illegal target %s", id)
					}
				}
			}
		}
	}
	momB17PickCard(t, g, small)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the legal target survived")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(elf) {
		t.Error("an unrelated creature died")
	}
}

func TestInvasionOfLorwynRefusesATargetWithTooMuchPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lorwynLands(g, me.ID, 3)
	big := b12Creature(g, g.Seats[1].ID, "Big", "Creature — Bear", 4, 4)
	momB17Enter(t, g, lorwynRow())
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
				t.Error("a creature with power 4 was accepted with three lands")
			}
		}
	}
}

func TestWinnowingForcesCountsYourLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forces := pushCatalogPermanent(g, me.ID, "Winnowing Forces", "Creature — Elf Warrior", invasionOfLorwynOracleID+"#1", false)
	if p, tough := ptOf(t, g, forces); p != 0 || tough != 0 {
		t.Fatalf("with no lands: %d/%d", p, tough)
	}
	lorwynLands(g, me.ID, 4)
	lorwynLands(g, g.Seats[1].ID, 2)
	if p, tough := ptOf(t, g, forces); p != 4 || tough != 4 {
		t.Errorf("with four lands of mine: %d/%d, want 4/4", p, tough)
	}
}
