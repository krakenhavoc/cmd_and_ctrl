package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardshu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/hu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b17_test.go — March of the Machine Sieges, batch 17 (#5477): both
// faces of ten double-faced cards, driven through the engine.

type momB17Face struct {
	key, name string
	keywords  []string
}

var momB17Faces = []momB17Face{
	{cardshu.InvasionOfDominariaOracleID, "Invasion of Dominaria", nil},
	{cardshu.InvasionOfDominariaOracleID + "#1", "Serra Faithkeeper", []string{"flying", "vigilance"}},
	{cardshu.InvasionOfEldraineOracleID, "Invasion of Eldraine", nil},
	{cardshu.InvasionOfEldraineOracleID + "#1", "Prickle Faeries", []string{"flying"}},
	{cardshu.InvasionOfErgamonOracleID, "Invasion of Ergamon", nil},
	{cardshu.InvasionOfErgamonOracleID + "#1", "Truga Cliffcharger", []string{"trample"}},
	{cardshu.InvasionOfKamigawaOracleID, "Invasion of Kamigawa", nil},
	{cardshu.InvasionOfKamigawaOracleID + "#1", "Rooftop Saboteurs", []string{"flying"}},
	{cardshu.InvasionOfMoagOracleID, "Invasion of Moag", nil},
	{cardshu.InvasionOfMoagOracleID + "#1", "Bloomwielder Dryads", nil},
	{cardshu.InvasionOfUlgrothaOracleID, "Invasion of Ulgrotha", nil},
	{cardshu.InvasionOfUlgrothaOracleID + "#1", "Grandmother Ravi Sengir", []string{"flying"}},
	{cardshu.InvasionOfXerexOracleID, "Invasion of Xerex", nil},
	{cardshu.InvasionOfXerexOracleID + "#1", "Vertex Paladin", []string{"flying"}},
	{cardshu.InvasionOfZendikarOracleID, "Invasion of Zendikar", nil},
	{cardshu.InvasionOfZendikarOracleID + "#1", "Awakened Skyclave", []string{"vigilance", "haste"}},
	{cardshu.InvasionOfAzgolOracleID, "Invasion of Azgol", nil},
	{cardshu.InvasionOfAzgolOracleID + "#1", "Ashen Reaper", []string{"menace"}},
	{cardshu.InvasionOfLorwynOracleID, "Invasion of Lorwyn", nil},
	{cardshu.InvasionOfLorwynOracleID + "#1", "Winnowing Forces", nil},
}

func TestMomB17EveryFaceIsRegisteredFull(t *testing.T) {
	for _, f := range momB17Faces {
		spec, ok := effects.Lookup(f.key)
		if !ok {
			t.Errorf("%s is not registered", f.name)
			continue
		}
		if spec.Name != f.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
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
	return effectstest.SiegeRow(oracleID, front, back, cost, defense, backType, power, toughness)
}

// momB17Enter puts the Siege onto the battlefield from the active seat's hand
// and names the next seat as its protector. The ETB trigger is left for the
// caller to answer.
func momB17Enter(t *testing.T, g *game.Game, row cards.Card) (owner, protector *game.Player, id uuid.UUID) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	owner = g.Seats[seat]
	protector = g.Seats[(seat+1)%len(g.Seats)]
	id = effectstest.ImportToHand(row, owner)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("move battle to battlefield: %v", err)
	}
	effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
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
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	life, hand := owner.Life, owner.Hand.Size()
	momB17Enter(t, g, momB17Row(cardshu.InvasionOfDominariaOracleID, "Invasion of Dominaria", "Serra Faithkeeper", "{2}{W}", "5", "Creature — Angel", "4", "4"))
	effectstest.PassPriorityAroundTable(t, g)
	if owner.Life != life+4 {
		t.Errorf("life = %d, want %d", owner.Life, life+4)
	}
	if owner.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d (the Siege left it, one card drawn)", owner.Hand.Size(), hand+1)
	}
}

func TestSerraFaithkeeperHasFlyingAndVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Serra Faithkeeper", "Creature — Angel", cardshu.InvasionOfDominariaOracleID+"#1", false)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

// --- Invasion of Eldraine // Prickle Faeries ------------------------------------

func TestInvasionOfEldraineMakesTheTargetOpponentDiscardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, _, _ = momB17Enter(t, g, momB17Row(cardshu.InvasionOfEldraineOracleID, "Invasion of Eldraine", "Prickle Faeries", "{3}{B}", "4", "Creature — Faerie", "2", "2"))
	victim := g.Seats[2]
	hand := victim.Hand.Size()
	momB17PickPlayer(t, g, victim.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, victim.ID); got != 2 {
		t.Fatalf("discard owed = %d, want 2", got)
	}
	effectstest.AnswerDiscard(t, g, victim.ID, victim.Hand.Cards[0].InstanceID, victim.Hand.Cards[1].InstanceID)
	if victim.Hand.Size() != hand-2 {
		t.Errorf("hand = %d, want %d", victim.Hand.Size(), hand-2)
	}
	if g.Seats[1].Hand.Size() != 7 {
		t.Errorf("another opponent's hand changed to %d", g.Seats[1].Hand.Size())
	}
}

func TestPrickleFaeriesPingsAnOpponentWithTwoOrFewerCardsInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	faeries := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Prickle Faeries", "Creature — Faerie", cardshu.InvasionOfEldraineOracleID+"#1", false)
	opp := g.Seats[1]
	opp.Hand.Cards = opp.Hand.Cards[:2]
	life := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TriggerOnStack(g, faeries) == nil {
		t.Fatal("no upkeep trigger with two cards in hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life = %d, want %d", opp.Life, life-2)
	}
}

func TestPrickleFaeriesIgnoresAnOpponentWithThreeCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	faeries := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Prickle Faeries", "Creature — Faerie", cardshu.InvasionOfEldraineOracleID+"#1", false)
	opp := g.Seats[1]
	opp.Hand.Cards = opp.Hand.Cards[:3]
	life := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TriggerOnStack(g, faeries) != nil {
		t.Fatal("triggered with three cards in hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life = %d, want %d", opp.Life, life)
	}
}

func TestPrickleFaeriesIgnoresItsControllersUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	faeries := effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Prickle Faeries", "Creature — Faerie", cardshu.InvasionOfEldraineOracleID+"#1", false)
	g.Seats[1].Hand.Cards = g.Seats[1].Hand.Cards[:0]
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TriggerOnStack(g, faeries) != nil {
		t.Fatal("triggered on its controller's own upkeep")
	}
}

// --- Invasion of Ergamon // Truga Cliffcharger ----------------------------------

func ergamonRow() cards.Card {
	return momB17Row(cardshu.InvasionOfErgamonOracleID, "Invasion of Ergamon", "Truga Cliffcharger", "{R}{G}", "5", "Creature — Rhino", "3", "4")
}

func TestInvasionOfErgamonMakesATreasureAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, _, _ := momB17Enter(t, g, ergamonRow())
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Treasure"); n != 1 {
		t.Fatalf("Treasures = %d, want 1", n)
	}
	hand := owner.Hand.Size()
	effectstest.AnswerDiscard(t, g, owner.ID, owner.Hand.Cards[0].InstanceID)
	if owner.Hand.Size() != hand {
		t.Errorf("hand = %d, want %d (discarded one, drew one)", owner.Hand.Size(), hand)
	}
}

func TestInvasionOfErgamonDrawsNothingWhenYouDeclineToDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, _, _ := momB17Enter(t, g, ergamonRow())
	effectstest.PassPriorityAroundTable(t, g)
	hand := owner.Hand.Size()
	effectstest.AnswerDiscard(t, g, owner.ID)
	if owner.Hand.Size() != hand {
		t.Errorf("hand = %d, want %d", owner.Hand.Size(), hand)
	}
	if n := effectstest.CountBattlefieldByName(g, "Treasure"); n != 1 {
		t.Errorf("Treasures = %d, want 1", n)
	}
}

// Truga Cliffcharger, cast transformed from the defeated Siege: it discards a
// card and fetches a land or battle card, and only if a card was discarded.
func trugaSetup(t *testing.T) (g *game.Game, owner *game.Player, forest uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	owner = g.Seats[g.Turn.ActiveSeat]
	forest = uuid.New()
	owner.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner.ID, Controller: owner.ID})
	_, _, id := momB17Enter(t, g, ergamonRow())
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, owner.ID)
	effectstest.DefeatBattle(t, g, id, 5)
	effectstest.ToMainPhaseSameTurn(t, g)
	if err := g.CastSpell(owner.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast Truga Cliffcharger: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, owner, forest
}

func TestTrugaCliffchargerDiscardsToFetchALand(t *testing.T) {
	g, owner, forest := trugaSetup(t)
	effectstest.AnswerDiscard(t, g, owner.ID, owner.Hand.Cards[0].InstanceID)
	search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, owner.ID)
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
	effectstest.AnswerDiscard(t, g, owner.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range owner.Hand.Cards {
		if c.InstanceID == forest {
			t.Error("the Forest was fetched without a discard")
		}
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, owner.ID) != nil {
		t.Error("a search prompt opened though nothing was discarded")
	}
}

// --- Invasion of Kamigawa // Rooftop Saboteurs ----------------------------------

func TestInvasionOfKamigawaTapsAndStunsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[2]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, g.Seats[g.Turn.ActiveSeat].ID, "Mine", "Creature — Bear", 2, 2)
	momB17Enter(t, g, momB17Row(cardshu.InvasionOfKamigawaOracleID, "Invasion of Kamigawa", "Rooftop Saboteurs", "{3}{U}", "4", "Creature — Moonfolk Ninja", "2", "3"))
	momB17PickCard(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.BattlefieldCardFor(g, bear)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	saboteurs := effectstest.PushCatalogPermanent(g, me.ID, "Rooftop Saboteurs", "Creature — Moonfolk Ninja", cardshu.InvasionOfKamigawaOracleID+"#1", false)
	battle := effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Some Siege", "Battle — Siege", "", false)
	if !effectstest.HasKeywordOnBattlefield(t, g, saboteurs, "flying") {
		t.Error("no flying")
	}
	deal := func(target uuid.UUID, combat bool) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: saboteurs, Target: target, Amount: 2, Combat: combat})
		})
	}
	hand := me.Hand.Size()
	deal(g.Seats[1].ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("noncombat damage to a player triggered")
	}
	deal(g.Seats[1].ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatal("combat damage to a player did not draw")
	}
	deal(battle, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand = %d, want %d (a card for the player, one for the battle)", me.Hand.Size(), hand+2)
	}
}

// --- Invasion of Moag // Bloomwielder Dryads ------------------------------------

func TestInvasionOfMoagPutsACounterOnEachCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	_, _, siege := momB17Enter(t, g, momB17Row(cardshu.InvasionOfMoagOracleID, "Invasion of Moag", "Bloomwielder Dryads", "{2}{G}{W}", "5", "Creature — Dryad", "3", "3"))
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Bloomwielder Dryads", "Creature — Dryad", cardshu.InvasionOfMoagOracleID+"#1", false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToEndStep(t, g)
	momB17PickCard(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := momB17Counters(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("counters on the target = %d, want 1", n)
	}
}

func TestBloomwielderDryadsAsksOpponentsToPayForWard(t *testing.T) {
	spec, _ := effects.Lookup(cardshu.InvasionOfMoagOracleID + "#1")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[2]
	life, mine := victim.Life, me.Life
	momB17Enter(t, g, momB17Row(cardshu.InvasionOfUlgrothaOracleID, "Invasion of Ulgrotha", "Grandmother Ravi Sengir", "{4}{B}", "5", "Legendary Creature — Human Wizard", "3", "3"))
	momB17PickPlayer(t, g, victim.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if victim.Life != life-3 {
		t.Errorf("victim life = %d, want %d", victim.Life, life-3)
	}
	if me.Life != mine+3 {
		t.Errorf("my life = %d, want %d", me.Life, mine+3)
	}
}

func TestInvasionOfUlgrothaCannotDamageItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, _, siege := momB17Enter(t, g, momB17Row(cardshu.InvasionOfUlgrothaOracleID, "Invasion of Ulgrotha", "Grandmother Ravi Sengir", "{4}{B}", "5", "Legendary Creature — Human Wizard", "3", "3"))
	momB17PickCard(t, g, siege)
	effectstest.PassPriorityAroundTable(t, g)
	if n := momB17Counters(g, siege, game.CounterDefense); n != 5 {
		t.Errorf("defense = %d, want the untouched 5", n)
	}
}

func TestGrandmotherRaviSengirGrowsWhenAnOpponentsCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ravi := effectstest.PushCatalogPermanent(g, me.ID, "Grandmother Ravi Sengir", "Legendary Creature — Human Wizard", cardshu.InvasionOfUlgrothaOracleID+"#1", false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 1, 1)
	life := me.Life
	kill := func(id uuid.UUID) {
		g.WithWriteLock(func() {
			if err := g.DealDamageToCreatureForEffect(uuid.Nil, id, 5); err != nil {
				t.Fatal(err)
			}
		})
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	hand := opp.Hand.Size()
	momB17Enter(t, g, momB17Row(cardshu.InvasionOfXerexOracleID, "Invasion of Xerex", "Vertex Paladin", "{2}{W}{U}", "4", "Creature — Angel Knight", "*", "*"))
	momB17PickCard(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature is still on the battlefield")
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d", opp.Hand.Size(), hand+1)
	}
}

func TestInvasionOfXerexMayTargetNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, _, siege := momB17Enter(t, g, momB17Row(cardshu.InvasionOfXerexOracleID, "Invasion of Xerex", "Vertex Paladin", "{2}{W}{U}", "4", "Creature — Angel Knight", "*", "*"))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(siege) {
		t.Error("the Siege left the battlefield")
	}
}

func TestVertexPaladinCountsTheCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	paladin := effectstest.PushCatalogPermanent(g, me.ID, "Vertex Paladin", "Creature — Angel Knight", cardshu.InvasionOfXerexOracleID+"#1", false)
	if p, tough := effectstest.PtOf(t, g, paladin); p != 1 || tough != 1 {
		t.Fatalf("alone: %d/%d, want 1/1", p, tough)
	}
	effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	if p, tough := effectstest.PtOf(t, g, paladin); p != 3 || tough != 3 {
		t.Errorf("with two more of mine: %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, paladin, "flying") {
		t.Error("no flying")
	}
}

// --- Invasion of Zendikar // Awakened Skyclave ----------------------------------

func TestInvasionOfZendikarFetchesTwoBasicLandsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	var forests []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		forests = append(forests, id)
		owner.Library.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner.ID, Controller: owner.ID})
	}
	owner.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Nonbasic", TypeLine: "Land", Owner: owner.ID, Controller: owner.ID})
	momB17Enter(t, g, momB17Row(cardshu.InvasionOfZendikarOracleID, "Invasion of Zendikar", "Awakened Skyclave", "{3}{G}", "3", "Creature — Elemental", "4", "4"))
	effectstest.PassPriorityAroundTable(t, g)
	search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, owner.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sky := effectstest.PushCatalogPermanent(g, me.ID, "Awakened Skyclave", "Creature — Elemental", cardshu.InvasionOfZendikarOracleID+"#1", false)
	c := effectstest.LayeredCard(t, g, sky)
	if !c.IsLand() || !c.IsCreature() {
		t.Errorf("%q is not a land creature", c.TypeLine)
	}
	for _, kw := range []string{"vigilance", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, sky, kw) {
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
	return momB17Row(cardshu.InvasionOfAzgolOracleID, "Invasion of Azgol", "Ashen Reaper", "{B}{R}", "4", "Creature — Zombie Elemental", "2", "1")
}

func TestInvasionOfAzgolMakesThemSacrificeAndLoseALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[2]
	keep := effectstest.B12Creature(g, victim.ID, "Keep", "Creature — Bear", 2, 2)
	lose := effectstest.B12Creature(g, victim.ID, "Lose", "Creature — Bear", 2, 2)
	life := victim.Life
	momB17Enter(t, g, azgolRow())
	momB17PickPlayer(t, g, victim.ID)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, victim.ID)
	if c == nil {
		t.Fatal("the victim was not asked what to sacrifice")
	}
	if err := g.ResolveChooseCards(c.ID, victim.ID, []uuid.UUID{lose}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lose) || !g.Battlefield.Contains(keep) {
		t.Error("the wrong creature was sacrificed")
	}
	if victim.Life != life-1 {
		t.Errorf("life = %d, want %d", victim.Life, life-1)
	}
}

func TestInvasionOfAzgolStillCostsALifeWithNothingToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[2]
	life := victim.Life
	momB17Enter(t, g, azgolRow())
	momB17PickPlayer(t, g, victim.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if victim.Life != life-1 {
		t.Errorf("life = %d, want %d", victim.Life, life-1)
	}
}

func TestAshenReaperGrowsOnlyIfAPermanentWentToAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	reaper := effectstest.PushCatalogPermanent(g, me.ID, "Ashen Reaper", "Creature — Zombie Elemental", cardshu.InvasionOfAzgolOracleID+"#1", false)
	effectstest.AdvanceToEndStep(t, g)
	if effectstest.TriggerOnStack(g, reaper) != nil {
		t.Fatal("triggered with nothing in a graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := momB17Counters(g, reaper, game.CounterPlusOne); n != 0 {
		t.Fatalf("counters = %d, want 0", n)
	}
}

func TestAshenReaperGrowsAfterAPermanentDiesThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	reaper := effectstest.PushCatalogPermanent(g, me.ID, "Ashen Reaper", "Creature — Zombie Elemental", cardshu.InvasionOfAzgolOracleID+"#1", false)
	victim := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, victim, 5); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := momB17Counters(g, reaper, game.CounterPlusOne); n != 1 {
		t.Errorf("counters = %d, want 1", n)
	}
}

// --- Invasion of Lorwyn // Winnowing Forces -------------------------------------

func lorwynRow() cards.Card {
	return momB17Row(cardshu.InvasionOfLorwynOracleID, "Invasion of Lorwyn", "Winnowing Forces", "{4}{B}{G}", "5", "Creature — Elf Warrior", "*", "*")
}

func lorwynLands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Permanent(g, owner, "Forest", "Basic Land — Forest")
	}
}

func TestInvasionOfLorwynDestroysANonElfWithPowerAtMostYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lorwynLands(g, me.ID, 3)
	small := effectstest.B12Creature(g, g.Seats[1].ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, g.Seats[1].ID, "Big", "Creature — Bear", 4, 4)
	elf := effectstest.B12Creature(g, g.Seats[1].ID, "Elf", "Creature — Elf Warrior", 1, 1)
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
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the legal target survived")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(elf) {
		t.Error("an unrelated creature died")
	}
}

func TestInvasionOfLorwynRefusesATargetWithTooMuchPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lorwynLands(g, me.ID, 3)
	big := effectstest.B12Creature(g, g.Seats[1].ID, "Big", "Creature — Bear", 4, 4)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forces := effectstest.PushCatalogPermanent(g, me.ID, "Winnowing Forces", "Creature — Elf Warrior", cardshu.InvasionOfLorwynOracleID+"#1", false)
	if p, tough := effectstest.PtOf(t, g, forces); p != 0 || tough != 0 {
		t.Fatalf("with no lands: %d/%d", p, tough)
	}
	lorwynLands(g, me.ID, 4)
	lorwynLands(g, g.Seats[1].ID, 2)
	if p, tough := effectstest.PtOf(t, g, forces); p != 4 || tough != 4 {
		t.Errorf("with four lands of mine: %d/%d, want 4/4", p, tough)
	}
}
