package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_b13_cards_test.go — Aetherdrift (DFT) leftovers, batches 13 and 14
// (#4464, #4452). Each test plays the real catalog entry.

const (
	dftB13QuagFeastOracle = "78957246-4092-4f4a-80e3-142e24226962"
)

func dftB13Stock(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: p.ID, Controller: p.ID})
		}
	})
}

func dftB13EmptyGraveyard(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() { p.Graveyard.Cards = nil })
}

func TestDftB13QuagFeastDestroysOnlyWhenTheGraveyardIsBigEnough(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dftB13Stock(g, me, 5)
	dftB13EmptyGraveyard(g, me)
	small := dftSTCostly(g, opp.ID, "Small", "Creature — Bear", "{2}")
	big := dftSTCostly(g, opp.ID, "Big", "Creature — Bear", "{4}")
	// Two milled cards: mana value 2 dies (2 <= 2), mana value 4 does not.
	effectstest.CastCatalogSpell(t, g, "Quag Feast", "Sorcery", dftB13QuagFeastOracle, effectstest.DftABCardRef(big))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, big); !ok {
		t.Error("a mana value 4 creature was destroyed with two cards in the graveyard")
	}
	// Two cards in the graveyard already + the quag feast itself + two milled.
	effectstest.CastCatalogSpell(t, g, "Quag Feast", "Sorcery", dftB13QuagFeastOracle, effectstest.DftABCardRef(big))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, big); ok {
		t.Error("a mana value 4 creature survived with five cards in the graveyard")
	}
	_ = small
}

func TestDftB13QuagFeastTargetsAVehicleButNotALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dftB13Stock(g, me, 5)
	veh := dftSTCostly(g, opp.ID, "Cart", "Artifact — Vehicle", "{1}")
	land := effectstest.DftABPush(g, opp.ID, "Plains", "Basic Land — Plains", "", 0, 0)
	if err := effectstest.CastCatalogSpellErr(t, g, "Quag Feast", "Sorcery", dftB13QuagFeastOracle, effectstest.DftABCardRef(land)); err == nil {
		t.Error("Quag Feast targeted a land")
	}
	effectstest.CastCatalogSpell(t, g, "Quag Feast", "Sorcery", dftB13QuagFeastOracle, effectstest.DftABCardRef(veh))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, veh); ok {
		t.Error("the Vehicle survived")
	}
}

const (
	dftB13WickerfolkOracle = "730d02c7-2f8a-47d3-99f0-e4ada7f6255c"
	dftB13CaradoraOracle   = "a9d60d80-bcef-45e0-8ded-d70bd3c2780f"
)

func TestDftB13WickerfolkIsCastFromTheGraveyardForItsCostLifeAndASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	if !game.CardCastableFromZone(dftB13WickerfolkOracle, game.ZoneGraveyard) {
		t.Fatal("the graveyard is not open to Wickerfolk Indomitable")
	}
	if got := game.AlternativeCostsOfferedFromZone(dftB13WickerfolkOracle, game.ZoneHand); len(got) != 0 {
		t.Fatalf("the graveyard offer is shown from hand: %+v", got)
	}
	offers := game.AlternativeCostsOfferedFromZone(dftB13WickerfolkOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].ManaCost != "{3}{B}" || offers[0].Life != 2 {
		t.Fatalf("graveyard offers = %+v, want {3}{B}, 2 life and a sacrifice", offers)
	}
	wf := effectstest.GraveyardCardOf(g, me, "Wickerfolk Indomitable", "Artifact Creature — Scarecrow", "{3}{B}", dftB13WickerfolkOracle)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	land := effectstest.DftABPush(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	params := game.CastSpellParams{FromZone: "graveyard", AlternativeCost: offers[0].Key, Strict: true}

	// A land is neither an artifact nor a creature.
	effectstest.PayMana(t, g, me, "{C}{C}{C}{B}")
	params.AltCostIDs = []uuid.UUID{land}
	if err := g.CastSpell(me.ID, wf, params); err == nil {
		t.Fatal("sacrificed a land")
	}
	// Short of the mana: {B} is not {3}{B}.
	g.WithWriteLock(func() { me.ManaPool = nil })
	effectstest.PayMana(t, g, me, "{B}")
	params.AltCostIDs = []uuid.UUID{fodder}
	if err := g.CastSpell(me.ID, wf, params); err == nil {
		t.Fatal("cast from the graveyard without paying the mana cost")
	}
	if !g.Battlefield.Contains(fodder) || !effectstest.InGraveyard(me, wf) {
		t.Fatal("a refused cast paid part of its cost")
	}
	g.WithWriteLock(func() { me.ManaPool = nil })

	life := me.Life
	effectstest.PayMana(t, g, me, "{C}{C}{C}{B}")
	if err := g.CastSpell(me.ID, wf, params); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wf) {
		t.Fatal("Wickerfolk Indomitable did not enter")
	}
}

func TestDftB13WickerfolkFromTheHandCostsOnlyItsManaCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	card := game.Card{InstanceID: uuid.New(), Name: "Wickerfolk Indomitable", TypeLine: "Artifact Creature — Scarecrow",
		ManaCost: "{3}{B}", OracleID: dftB13WickerfolkOracle, Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	life := me.Life
	effectstest.PayMana(t, g, me, "{C}{C}{C}{B}")
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(card.InstanceID) || me.Life != life {
		t.Errorf("on battlefield %v, life %d (want %d)", g.Battlefield.Contains(card.InstanceID), me.Life, life)
	}
}

func TestDftB13CaradoraAddsACounterToCreaturesAndVehiclesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.DftABPush(g, me.ID, "Caradora, Heart of Alacria", "Legendary Creature — Human Knight", dftB13CaradoraOracle, 4, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	cart := effectstest.DftABPush(g, me.ID, "Cart", "Artifact — Vehicle", "", 3, 3)
	rock := effectstest.DftABPush(g, me.ID, "Rock", "Artifact", "", 0, 0)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, c := range []struct {
		id   uuid.UUID
		want int
	}{{bear, 2}, {cart, 2}, {rock, 1}, {theirs, 1}} {
		if err := g.AddCounter(c.id, "+1/+1", 1); err != nil {
			t.Fatalf("AddCounter: %v", err)
		}
		if got := effectstest.CountersOn(g, c.id, game.CounterPlusOne); got != c.want {
			t.Errorf("counters on %v = %d, want %d", c.id, got, c.want)
		}
	}
}

func TestDftB13CaradoraSearchesForAMountOrVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	cart := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Cart", TypeLine: "Artifact — Vehicle"})
	effectstest.RfdCast(t, g, "Caradora, Heart of Alacria", "Legendary Creature — Human Knight", dftB13CaradoraOracle, 4, 2)
	var offered []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceSearchLibrary {
				offered = append(offered, c.SearchCards...)
			}
		}
	})
	if effectstest.HasID(offered, bear) {
		t.Error("a plain creature is offered")
	}
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(cart) {
		t.Error("the Vehicle was not searched into the hand")
	}
	if me.Hand.Contains(bear) {
		t.Error("a plain creature was found")
	}
}

const dftB13ChandraOracle = "30bb01c1-9253-4251-80e7-fb8236b3d858"

func dftB13Chandra(t *testing.T, g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	t.Helper()
	id := effectstest.PushCatalogPermanent(g, owner, "Chandra, Spark Hunter", "Legendary Planeswalker — Chandra", dftB13ChandraOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, loyalty) })
	return id
}

func TestDftB13ChandraZeroMakesAThreeTwoVehicleWithCrewOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chandra := dftB13Chandra(t, g, me.ID, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, chandra, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var veh uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Vehicle" {
			veh = c.InstanceID
		}
	}
	if veh == uuid.Nil {
		t.Fatal("no Vehicle token")
	}
	c, _ := effectstest.BattlefieldCard(g, veh)
	if !c.HasSubtype("Vehicle") || c.Power != 3 || c.Toughness != 2 || !c.IsArtifact() {
		t.Errorf("token is %s %d/%d", c.TypeLine, c.Power, c.Toughness)
	}
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	effectstest.DftABCrew(t, g, me, veh, 0, one)
}

func TestDftB13ChandraPlusTwoDrawsOnlyWhenYouSacrificeOrDiscard(t *testing.T) {
	for _, tc := range []struct {
		name   string
		option int
	}{{"sacrifice", 0}, {"discard", 1}, {"decline", -1}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := effectstest.EoTable(t)
			art := effectstest.EoArtifact(g, me.ID, "Bauble")
			land := effectstest.EoHand(me, "Forest", "Basic Land — Forest")
			chandra := dftB13Chandra(t, g, me.ID, 4)
			effectstest.B16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
			effectstest.PassPriorityAroundTable(t, g)
			hand := me.Hand.Size()
			switch tc.option {
			case 0:
				effectstest.EoPay(t, g, me.ID, 0, art)
			case 1:
				effectstest.EoPay(t, g, me.ID, 1, land)
			default:
				effectstest.AnswerPayUnless(t, g, me.ID, false)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want := hand
			if tc.option == 0 {
				want = hand + 1
			}
			if tc.option == 1 {
				want = hand // discarded one, drew one
			}
			if got := me.Hand.Size(); got != want {
				t.Errorf("hand %d, want %d", got, want)
			}
			if n := effectstest.CountersOn(g, chandra, game.CounterLoyalty); n != 6 {
				t.Errorf("loyalty %d, want 6", n)
			}
		})
	}
}

func TestDftB13ChandraAnimatesAVehicleWithHasteAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dftB13Chandra(t, g, me.ID, 4)
	cart := effectstest.DftABPush(g, me.ID, "Cart", "Artifact — Vehicle", "", 3, 3)
	theirs := effectstest.DftABPush(g, g.Seats[1].ID, "Their Cart", "Artifact — Vehicle", "", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	if effectstest.HasID(prompt.PickTargetCards, theirs) {
		t.Error("an opponent's Vehicle is offered")
	}
	effectstest.PickCard(t, g, me.ID, cart)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, cart)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Error("the Vehicle is not an artifact creature")
	}
	if !effectstest.HasEffectiveKeyword(t, g, cart, "haste") {
		t.Error("the Vehicle has no haste")
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if c, _ := effectstest.BattlefieldCard(g, cart); c.IsCreature() {
		t.Error("the Vehicle is still a creature after the turn")
	}
}

func TestDftB13ChandraEmblemDealsThreeWhenAnArtifactYouControlEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chandra := dftB13Chandra(t, g, me.ID, 7)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, chandra, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("one emblem: %d", me.Emblems.Size())
	}
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || opp.Life != life {
		t.Error("a creature that is not an artifact triggered the emblem")
	}
	effectstest.CastCatalogSpell(t, g, "Bauble", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("an artifact entered: %d -> %d, want 3 damage", life, opp.Life)
	}
}

const (
	dftB13DredgerOracle = "3f21c1c6-7992-4036-ad3e-b599153e54fe"
	dftB13FangOracle    = "1a9bfe76-6c74-4393-b04e-461422a29ddb"
	dftB13RiseOracle    = "8eeb83b4-dca1-497e-a6b7-0485807aeef3"
)

// dftB13Vanilla puts a printed-vanilla creature card (its text was read and
// is empty) into a graveyard or library.
func dftB13Vanilla(p *game.Player, name string, graveyard bool) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: p.ID, Controller: p.ID, NoAbilitiesText: game.PrintedYes}
	if graveyard {
		p.Graveyard.PushTop(c)
	} else {
		p.Library.PushTop(c)
	}
	return c.InstanceID
}

func dftB13Keyworded(p *game.Player, name string) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Owner: p.ID, Controller: p.ID, Keywords: []string{"flying"}, NoAbilitiesText: game.PrintedNo}
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

func TestDftB13DredgersInsightGainsLifeOncePerBatchOfArtifactOrCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.DftABPush(g, me.ID, "Dredger's Insight", "Enchantment", dftB13DredgerOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	art := effectstest.DftABInGraveyard(me, "Gizmo", "Artifact")
	bear := effectstest.DftABInGraveyard(me, "Bear", "Creature — Bear")
	land := effectstest.DftABInGraveyard(me, "Forest", "Basic Land — Forest")
	life := me.Life
	if err := g.ReturnFromGraveyardForEffect(land, game.ZoneHand); err != nil {
		t.Fatalf("return land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Error("a land card leaving the graveyard gained life")
	}
	for _, id := range []uuid.UUID{art, bear} {
		if err := g.ReturnFromGraveyardForEffect(id, game.ZoneHand); err != nil {
			t.Fatalf("return: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d (1 for each separate departure)", me.Life, life+2)
	}
}

func TestDftB13DredgersInsightMillsFourAndTakesAnArtifactCreatureOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PlTop(me, "Deep", "Instant", "")
	effectstest.PlTop(me, "Spell", "Sorcery", "")
	gizmo := effectstest.PlTop(me, "Gizmo", "Artifact", "")
	effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.PlTop(me, "Ench", "Enchantment", "")
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.DftABCast(t, g, me, game.Card{Name: "Dredger's Insight", TypeLine: "Enchantment", ManaCost: "{1}{G}", OracleID: dftB13DredgerOracle})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("nothing was offered from among the milled cards")
	}
	if !effectstest.HasID(pick.ChooseCards, gizmo) || len(pick.ChooseCards) != 2 {
		t.Errorf("offered %d cards, want the artifact and the land", len(pick.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, gizmo)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(gizmo) {
		t.Error("the artifact is not in hand")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d: the card left the graveyard", me.Life, life+1)
	}
}

func TestDftB13FangDruidSummonerFindsAVanillaCreatureInEitherZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vanillaLib := dftB13Vanilla(me, "Library Bear", false)
	keyworded := dftB13Keyworded(me, "Flier")
	vanillaYard := dftB13Vanilla(me, "Graveyard Bear", true)
	effectstest.LgKnowLibrary(me)
	effectstest.CastCatalogSpell(t, g, "Fang-Druid Summoner", "Creature — Ape Druid", dftB13FangOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	dftB13BeginSearch(t, g, c)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	c = effectstest.SearchChoiceFor(g, me.ID)
	if effectstest.HasID(c.SearchCards, keyworded) {
		t.Error("a creature with flying is offered as having no abilities")
	}
	if effectstest.HasID(c.SearchCards, vanillaLib) {
		t.Error("the library was not chosen but is shown")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, vanillaYard)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(vanillaYard) {
		t.Error("the graveyard card did not go to hand")
	}
	if effectstest.LibraryKnownTo(me, me.ID) == 0 {
		t.Error("the library was shuffled although only the graveyard was searched")
	}
}

func TestDftB13FangDruidSummonerSearchingTheLibraryShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vanillaLib := dftB13Vanilla(me, "Library Bear", false)
	effectstest.LgKnowLibrary(me)
	effectstest.CastCatalogSpell(t, g, "Fang-Druid Summoner", "Creature — Ape Druid", dftB13FangOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	dftB13BeginSearch(t, g, effectstest.SearchChoiceFor(g, me.ID))
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary)
	effectstest.AnswerSearchByID(t, g, me.ID, vanillaLib)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(vanillaLib) {
		t.Fatal("the library card did not go to hand")
	}
	if effectstest.LibraryKnownTo(me, me.ID) != 0 {
		t.Error("the library was not shuffled")
	}
}

func TestDftB13RiseFromTheWreckReturnsOneCardPerClause(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.DftABInGraveyard(me, "Bear", "Creature — Bear")
	horse := effectstest.DftABInGraveyard(me, "Horse", "Creature — Mount")
	cart := effectstest.DftABInGraveyard(me, "Cart", "Artifact — Vehicle")
	vanilla := dftB13Vanilla(me, "Vanilla", true)
	keyworded := dftB13Keyworded(me, "Flier")
	other := effectstest.DftABInGraveyard(me, "Other", "Creature — Rat")
	if err := effectstest.CastCatalogSpellErr(t, g, "Rise from the Wreck", "Sorcery", dftB13RiseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: keyworded, Slot: 3}}); err == nil {
		t.Error("a creature with flying filled the creature-with-no-abilities clause")
	}
	effectstest.CastCatalogSpell(t, g, "Rise from the Wreck", "Sorcery", dftB13RiseOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Slot: 0},
		{Kind: game.TargetCard, ID: horse, Slot: 1},
		{Kind: game.TargetCard, ID: cart, Slot: 2},
		{Kind: game.TargetCard, ID: vanilla, Slot: 3},
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, horse, cart, vanilla} {
		if !me.Hand.Contains(id) {
			t.Errorf("%v was not returned", id)
		}
	}
	if me.Hand.Contains(other) || me.Hand.Contains(keyworded) {
		t.Error("an untargeted card was returned")
	}
}

func TestDftB13RiseFromTheWreckCanTargetNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.DftABInGraveyard(me, "Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Rise from the Wreck", "Sorcery", dftB13RiseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(bear) {
		t.Error("a card was returned without a target")
	}
}

const dftB13WebstrikeOracle = "48a1d55e-0b71-4724-ab2a-6d2c87e86b5b"

func TestDftB13WebstrikeEliteDestroysAnArtifactOrEnchantmentWithManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	two := dftSTCostly(g, opp.ID, "Two", "Artifact", "{2}")
	three := dftSTCostly(g, opp.ID, "Three", "Enchantment", "{3}")
	bear := dftSTCostly(g, opp.ID, "Bear", "Creature — Bear", "{2}")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.PushCatalogHandCard(me, "Webstrike Elite", "Creature — Insect Archer", dftB13WebstrikeOracle)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn"})
	effectstest.FillPool(me, 2)
	effectstest.FillPoolColored(me, "G", 2)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{XValue: 2}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt for the cycling trigger")
	}
	if !effectstest.HasID(prompt.PickTargetCards, two) || effectstest.HasID(prompt.PickTargetCards, three) || effectstest.HasID(prompt.PickTargetCards, bear) {
		t.Errorf("offered %v: want only the mana value 2 artifact", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, two)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, two); ok {
		t.Error("the artifact survived")
	}
	if _, ok := effectstest.BattlefieldCard(g, three); !ok {
		t.Error("the other enchantment was destroyed")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d (discarded one, drew one)", me.Hand.Size(), hand)
	}
}

func TestDftB13WebstrikeEliteCyclesWithNoTargetAtAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.PushCatalogHandCard(me, "Webstrike Elite", "Creature — Insect Archer", dftB13WebstrikeOracle)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn"})
	effectstest.FillPoolColored(me, "G", 2)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{XValue: 0}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || !me.Graveyard.Contains(id) {
		t.Errorf("hand %d (want %d), in graveyard %v", me.Hand.Size(), hand, me.Graveyard.Contains(id))
	}
}

const dftB13DrifterOracle = "b32649cb-bf5c-45fa-86f0-8bf4e8555f06"

func TestDftB13DuneDrifterReturnsAnArtifactOrCreatureWithManaValueXOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cheap := effectstest.GraveyardCardOf(g, me, "Cheap Bear", "Creature — Bear", "{2}", "")
	rock := effectstest.GraveyardCardOf(g, me, "Cheap Rock", "Artifact", "{1}", "")
	dear := effectstest.GraveyardCardOf(g, me, "Dear Bear", "Creature — Bear", "{3}", "")
	spell := effectstest.GraveyardCardOf(g, me, "Spell", "Sorcery", "{1}", "")
	effectstest.PayMana(t, g, me, "{C}{C}{W}{B}")
	drifter := effectstest.CastFromHandForTest(t, g, me, "Dune Drifter", "Artifact — Vehicle", "{X}{W}{B}", dftB13DrifterOracle,
		game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if !effectstest.HasID(pick.PickTargetCards, cheap) || !effectstest.HasID(pick.PickTargetCards, rock) {
		t.Errorf("offered %v, want the MV 2 creature and the MV 1 artifact", pick.PickTargetCards)
	}
	if effectstest.HasID(pick.PickTargetCards, dear) || effectstest.HasID(pick.PickTargetCards, spell) {
		t.Error("a card with mana value 3 or a sorcery is offered at X=2")
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) || !g.Battlefield.Contains(drifter) {
		t.Error("the creature did not return")
	}
	// Crew 2 is the engine's.
	one := effectstest.B12Creature(g, me.ID, "Two", "Creature — Rat", 2, 2)
	effectstest.DftABCrew(t, g, me, drifter, 0, one)
}

func TestDftB13DuneDrifterThatWasNotCastHasXZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	costed := effectstest.GraveyardCardOf(g, me, "Cheap Bear", "Creature — Bear", "{1}", "")
	free := effectstest.GraveyardCardOf(g, me, "Free Rock", "Artifact", "", "")
	drifter := effectstest.GraveyardCardOf(g, me, "Dune Drifter", "Artifact — Vehicle", "{X}{W}{B}", dftB13DrifterOracle)
	if err := g.ReturnFromGraveyardForEffect(drifter, game.ZoneBattlefield); err != nil {
		t.Fatalf("return: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, costed) || !effectstest.HasID(pick.PickTargetCards, free) {
		t.Errorf("offered %v, want only the mana value 0 artifact", pick.PickTargetCards)
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, free)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(free) {
		t.Error("the artifact did not return")
	}
}

const dftB13SwashbucklerOracle = "6a6cdb3e-c850-4b3c-bf4b-c53dd4eaf56e"

func dftB13Swashbuckler(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.DftABPush(g, owner, "Fearless Swashbuckler", "Creature — Fish Pirate", dftB13SwashbucklerOracle, 3, 3)
}

func TestDftB13SwashbucklerLootsWhenAPirateAndAVehicleAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pirate := dftB13Swashbuckler(g, me.ID)
	cart := effectstest.DftABPush(g, me.ID, "Haunted Hellride", effectstest.DftABVehicle, dftHLHellrideOracle, 3, 3)
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.DftABCrew(t, g, me, cart, 0, one)
	for i := 0; i < 6; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn"})
	}
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, pirate, cart)
	g.RunStateChecksForTest()
	// Haunted Hellride's own "whenever you attack" asks for its target first.
	effectstest.PickCard(t, g, me.ID, pirate)
	effectstest.AnswerOpenOrderingIfAny(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		g.ReadSnapshot(func() {
			t.Logf("attackers %d, stack %d, pending triggers %d, choices %d", len(g.AttackersThisCombatForEffect()), len(g.StackMeta), len(g.PendingTriggers), len(g.PendingChoices))
			for _, ev := range g.Events {
				if ev.Kind == game.EventAttack {
					c, _ := g.LookupCardForEffect(ev.CardID)
					t.Logf("attack event batch %d actor %v card %s %q creature=%v pirate=%v vehicle=%v", ev.Batch, ev.Actor == me.ID, c.Name, c.TypeLine, c.IsCreature(), c.HasSubtypeWord("Pirate"), c.HasSubtypeWord("Vehicle"))
				}
			}
		})
		t.Fatalf("hand %d, want %d after drawing three", got, hand+3)
	}
	if pc := effectstest.DiscardChoiceFor(g, me.ID); pc == nil {
		t.Error("no discard prompt for the two cards")
	}
}

func TestDftB13SwashbucklerDoesNothingWithoutAVehicleAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pirate := dftB13Swashbuckler(g, me.ID)
	effectstest.DftABPush(g, me.ID, "Cart", "Artifact — Vehicle", "", 3, 3)
	for i := 0; i < 6; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn"})
	}
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, pirate)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d: no Vehicle attacked", got, hand)
	}
}

func TestDftB13SwashbucklerGivesVehiclesYouControlHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dftB13Swashbuckler(g, me.ID)
	mine := effectstest.DftABPush(g, me.ID, "Cart", "Artifact — Vehicle", "", 3, 3)
	theirs := effectstest.DftABPush(g, opp.ID, "Their Cart", "Artifact — Vehicle", "", 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("my Vehicle has no haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's Vehicle has haste")
	}
}

// dftB13BeginSearch says "yes, I search" to a "you may search" prompt, and
// stops before the zone question so a test can answer it itself.
func dftB13BeginSearch(t *testing.T, g *game.Game, c *game.PendingChoice) {
	t.Helper()
	if c != nil && c.SearchIsUndecided() {
		if err := g.BeginSearchLibrary(c.ID, c.Chooser); err != nil {
			t.Fatalf("BeginSearchLibrary: %v", err)
		}
	}
}

const dftB13LifecraftOracle = "caa69b51-1efd-4d0e-803b-914a83ccb836"

func TestDftB13LifecraftEngineMakesCrewedVehiclesTheChosenTypeAndPumpsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goblin := effectstest.PushTribalCreature(g, me.ID, "Mogg Fanatic", "Creature — Goblin", 1, 1)
	theirGoblin := effectstest.PushTribalCreature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	engine := effectstest.PushNamedTribePermanent(t, g, me.ID, "Lifecraft Engine", "Artifact — Vehicle", dftB13LifecraftOracle, "Goblin")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == engine {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 4, 4
			}
		}
		g.BumpLayerVersionForTest()
	})
	cart := effectstest.DftABPush(g, me.ID, "Haunted Hellride", effectstest.DftABVehicle, dftHLHellrideOracle, 3, 3)
	if got := effectstest.EffectivePower(t, g, goblin); got != 2 {
		t.Errorf("my Goblin has power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, theirGoblin); got != 1 {
		t.Errorf("an opponent's Goblin has power %d, want 1 (you control)", got)
	}
	if effectstest.HasAbility(effectstest.EffectiveSubtypes(t, g, cart), "Goblin") {
		t.Error("an uncrewed Vehicle is a Goblin")
	}
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.DftABCrew(t, g, me, cart, 0, one)
	subs := effectstest.EffectiveSubtypes(t, g, cart)
	if !effectstest.HasAbility(subs, "Goblin") || !effectstest.HasAbility(subs, "Vehicle") {
		t.Errorf("crewed Vehicle subtypes = %v, want Goblin beside Vehicle", subs)
	}
	if got := effectstest.EffectivePower(t, g, cart); got != 4 {
		t.Errorf("crewed Vehicle power %d, want 4 (3 and the Engine's +1/+1)", got)
	}
}

func TestDftB13LifecraftEngineDoesNotPumpItselfWhenCrewed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	engine := effectstest.PushNamedTribePermanent(t, g, me.ID, "Lifecraft Engine", "Artifact — Vehicle", dftB13LifecraftOracle, "Goblin")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == engine {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 4, 4
			}
		}
		g.BumpLayerVersionForTest()
	})
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.DftABCrew(t, g, me, engine, 0, big)
	if !effectstest.HasAbility(effectstest.EffectiveSubtypes(t, g, engine), "Goblin") {
		t.Error("the crewed Engine is not the chosen type")
	}
	if got := effectstest.EffectivePower(t, g, engine); got != 4 {
		t.Errorf("the Engine has power %d, want 4: it is not pumped by itself", got)
	}
}

const (
	dftB13WinterOracle     = "343833c1-19fd-4e8f-b82e-0f20903caaff"
	dftB13PathmakerOracle  = "f8103865-5086-4f16-a35a-a56acdab1547"
	dftB13RoverOracle      = "0394846e-bc27-4d67-988e-b08a389690de"
	dftB13SitaOracle       = "3b9bc41a-a43f-4be7-9452-f61f5ad5b482"
	dftB13RiptideOracle    = "f3838ff1-5403-4f3f-8531-36029c13fcfc"
	dftB13RedshiftOracle   = "582c7e54-d64c-4ffb-ae5f-d2752b3ed10b"
	dftB13GetawayOracle    = "c844cc98-b8b9-447d-92ee-64f4a4906e7c"
	dftB13WinterArtifactTL = "Artifact Creature — Construct"
)

func TestDftB13WinterAndTheArtifactsYouControlHaveWardPayTwoLife(t *testing.T) {
	g, me, opp, _ := effectstest.WardTable(t, "Winter, Cursed Rider", "Legendary Creature — Human Warlock", dftB13WinterOracle)
	robot := effectstest.PushCatalogPermanent(g, me.ID, "Robot", dftB13WinterArtifactTL, "", false)
	for _, victim := range []uuid.UUID{effectstest.FindBattlefieldByName(g, "Winter, Cursed Rider"), robot} {
		blade := effectstest.CastAtWardedCreature(t, g, opp, victim)
		life := opp.Life
		prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
		if prompt == nil {
			t.Fatalf("no ward prompt for %v", victim)
		}
		if prompt.LifeCost != 2 {
			t.Errorf("ward asks %d life, want 2", prompt.LifeCost)
		}
		if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
			t.Fatalf("decline: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(victim) || !opp.Graveyard.Contains(blade) || opp.Life != life {
			t.Error("declining the ward must counter the spell and cost nothing")
		}
	}
}

func TestDftB13WinterDoesNotWardAnOpponentsArtifact(t *testing.T) {
	g, me, _, _ := effectstest.WardTable(t, "Winter, Cursed Rider", "Legendary Creature — Human Warlock", dftB13WinterOracle)
	theirs := effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Their Robot", dftB13WinterArtifactTL, "", false)
	effectstest.CastAtWardedCreature(t, g, me, theirs)
	if prompt := effectstest.PassUntilConfirmFor(t, g, me.ID); prompt != nil {
		t.Error("an opponent's artifact has ward")
	}
}

func TestDftB13WinterExhaustExilesXArtifactsAndShrinksEveryOtherNonartifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	winter := effectstest.DftABPush(g, me.ID, "Winter, Cursed Rider", "Legendary Creature — Human Warlock", dftB13WinterOracle, 3, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	robot := effectstest.B12Creature(g, me.ID, "Robot", dftB13WinterArtifactTL, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	one := effectstest.DftABInGraveyard(me, "Gizmo", "Artifact")
	two := effectstest.DftABInGraveyard(me, "Gadget", "Artifact Creature — Construct")
	notArtifact := effectstest.DftABInGraveyard(me, "Corpse", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{U}{B}")
	// A card that is not an artifact can't pay.
	if err := g.ActivateCatalogAbility(me.ID, winter, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{notArtifact}, Strict: true}); err == nil {
		t.Fatal("exiled a creature card that is not an artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, winter, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{one, two}, Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(one) || me.Graveyard.Contains(two) || !me.Graveyard.Contains(notArtifact) {
		t.Error("the two artifact cards were not the ones exiled")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 1 || tough != 1 {
		t.Errorf("my Bear is %d/%d, want 1/1 (-2/-2)", p, tough)
	}
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("the opponent's 2/2 survived -2/-2")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, robot); p != 2 || tough != 2 {
		t.Errorf("the artifact creature is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, winter); p != 3 || tough != 2 {
		t.Errorf("Winter is %d/%d, want 3/2 (\"each other\")", p, tough)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{U}{B}")
	if err := g.ActivateCatalogAbility(me.ID, winter, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

func TestDftB13GuidelightPathmakerPutsACheapArtifactOntoTheBattlefieldAndKeepsAnExpensiveOneInHand(t *testing.T) {
	for _, tc := range []struct {
		name          string
		cost          string
		onBattlefield bool
	}{{"mana value 2", "{2}", true}, {"mana value 3", "{3}", false}, {"mana value 0", "", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			rock := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: tc.cost})
			spell := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{1}"})
			effectstest.RfdCast(t, g, "Guidelight Pathmaker", "Artifact — Vehicle", dftB13PathmakerOracle, 6, 5)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			if effectstest.HasID(c.SearchCards, spell) {
				t.Error("a sorcery is offered")
			}
			effectstest.BeginSearchIfAsked(t, g, c)
			effectstest.AnswerSearchByID(t, g, me.ID, rock)
			effectstest.PassPriorityAroundTable(t, g)
			if got := g.Battlefield.Contains(rock); got != tc.onBattlefield {
				t.Errorf("on the battlefield %v, want %v", got, tc.onBattlefield)
			}
			if got := me.Hand.Contains(rock); got == tc.onBattlefield {
				t.Errorf("in hand %v, want %v", got, !tc.onBattlefield)
			}
		})
	}
}

func TestDftB13GuidelightPathmakerMayDeclineAndThenDoesNotShuffle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	effectstest.LgKnowLibrary(me)
	effectstest.RfdCast(t, g, "Guidelight Pathmaker", "Artifact — Vehicle", dftB13PathmakerOracle, 6, 5)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchIsUndecided() {
		if err := g.DeclineSearchLibrary(c.ID, c.Chooser); err != nil {
			t.Fatalf("decline: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LibraryKnownTo(me, me.ID) == 0 {
		t.Error("the library was shuffled although the search was declined")
	}
}

func TestDftB13RoverBladesGrantsDoubleStrikeAndAnimatedItCannotEquip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	blades := effectstest.DftABPush(g, me.ID, "Rover Blades", "Artifact — Equipment Vehicle", dftB13RoverOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	crew := effectstest.B12Creature(g, me.ID, "Crew", "Creature — Rat", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, blades, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, Strict: true}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Fatal("the equipped creature has no double strike")
	}
	effectstest.DftABCrew(t, g, me, blades, 1, crew)
	if !effectstest.HasEffectiveKeyword(t, g, blades, "double strike") {
		t.Error("the crewed Vehicle has no double strike of its own")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("a creature Equipment stayed attached")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	_ = g.ActivateCatalogAbility(me.ID, blades, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("a crewed Rover Blades equipped a creature")
	}
}

func TestDftB13SitaVarmaExhaustAddsXCountersAndMaySetEveryOtherCreatureToHerPower(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		sita := effectstest.DftABPush(g, me.ID, "Sita Varma, Masked Racer", "Legendary Creature — Human Rogue", dftB13SitaOracle, 2, 3)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
		theirs := effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Bear", "Creature — Bear", 1, 1)
		effectstest.AdvanceToMain(t, g)
		effectstest.FloatMana(t, g, me, "{C}{C}{G}{G}{U}")
		if err := g.ActivateCatalogAbility(me.ID, sita, 0, game.ActivateAbilityParams{XValue: 2, Strict: true}); err != nil {
			t.Fatalf("exhaust: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.ConfirmChoiceFor(g, me.ID)
		if prompt == nil {
			t.Fatal("no \"you may\" prompt")
		}
		if got := effectstest.CountersOn(g, sita, game.CounterPlusOne); got != 2 {
			t.Errorf("%d counters, want 2 before the question", got)
		}
		if err := g.ResolveConfirm(prompt.ID, me.ID, accept); err != nil {
			t.Fatalf("answer: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := 1
		if accept {
			want = 4 // Sita's power after two counters
		}
		if p, tough := effectstest.EffectivePTOf(t, g, bear); p != want || tough != want {
			t.Errorf("accept=%v: my Bear is %d/%d, want %d/%d", accept, p, tough, want, want)
		}
		if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 1 {
			t.Errorf("an opponent's Bear has power %d, want 1", p)
		}
		if p, tough := effectstest.PtOf(t, g, sita); p != 4 || tough != 5 {
			t.Errorf("Sita is %d/%d, want 4/5", p, tough)
		}
	}
}

func TestDftB13RiptideGearhulkTucksOnePermanentOfEachOpponentThirdFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a, b := g.Seats[1], g.Seats[2]
	rockA := effectstest.PushCatalogPermanent(g, a.ID, "Rock A", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, a.ID, "Plains", "Basic Land — Plains", "", false)
	rockB := effectstest.PushCatalogPermanent(g, b.ID, "Rock B", "Artifact", "", false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)
	for _, p := range []*game.Player{a, b} {
		for i := 0; i < 5; i++ {
			effectstest.PushLibraryCardForTest(p, game.Card{Name: "Filler"})
		}
	}
	effectstest.RfdCast(t, g, "Riptide Gearhulk", "Artifact Creature — Construct", dftB13RiptideOracle, 2, 5)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil && effectstest.HasID(pick.PickTargetCards, mine) {
		t.Error("my own permanent is offered")
	}
	// One prompt per opponent's clause, in turn order.
	effectstest.PickCard(t, g, me.ID, rockA)
	effectstest.PickCard(t, g, me.ID, rockB)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rockA) || g.Battlefield.Contains(rockB) {
		t.Fatal("a permanent stayed on the battlefield")
	}
	if got := effectstest.LibraryTopIDs(a, 3)[2]; got != rockA {
		t.Error("the first opponent's permanent is not third from the top of its owner's library")
	}
	if got := effectstest.LibraryTopIDs(b, 3)[2]; got != rockB {
		t.Error("the second opponent's permanent is not third from the top of its owner's library")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my permanent was tucked")
	}
}

func TestDftB13RedshiftAddsManaEqualToItsPowerForAbilitiesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	red := effectstest.DftABPush(g, me.ID, "Redshift, Rocketeer Chief", "Legendary Creature — Goblin Pilot", dftB13RedshiftOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(red, game.CounterPlusOne, 1) })
	if err := g.ActivateManaAbility(me.ID, red, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	color := effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	got := 0
	for _, tok := range me.ManaPool.Units() {
		if tok.Color == color {
			got++
		}
	}
	if got != 3 || me.ManaPool.Total() != 3 {
		t.Fatalf("pool has %d mana of %s (total %d), want 3 of one colour (Redshift's power with a counter)", got, color, me.ManaPool.Total())
	}
	cost, _ := game.ParseCost("{" + color + "}")
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{" + color + "}"}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("the mana paid for a spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("the mana refused an activated ability")
	}
}

func TestDftB13RedshiftExhaustPutsAnyNumberOfPermanentCardsFromHandOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	red := effectstest.DftABPush(g, me.ID, "Redshift, Rocketeer Chief", "Legendary Creature — Goblin Pilot", dftB13RedshiftOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	rock := effectstest.PutInHand(me, game.Card{Name: "Rock", TypeLine: "Artifact"})
	bear := effectstest.PutInHand(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	forest := effectstest.PutInHand(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	spell := effectstest.PutInHand(me, game.Card{Name: "Spell", TypeLine: "Sorcery"})
	effectstest.FillPool(me, 10)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, red, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if effectstest.HasID(pick.ChooseCards, spell) || len(pick.ChooseCards) != 3 {
		t.Errorf("offered %d cards, want the three permanent cards", len(pick.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, rock, bear, forest)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{rock, bear, forest} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%v did not enter", id)
		}
	}
	if !me.Hand.Contains(spell) {
		t.Error("the sorcery left the hand")
	}
}

func TestDftB13ExplosiveGetawayExilesOneAndReturnsItAtTheEndStepAfterFourDamageToEachCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saved := effectstest.B12Creature(g, me.ID, "Saved Giant", "Creature — Giant", 3, 3)
	hit := effectstest.B12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 5, 5)
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Explosive Getaway", "Sorcery", dftB13GetawayOracle, effectstest.DftABCardRef(saved))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(saved) {
		t.Fatal("the target is still on the battlefield")
	}
	if _, ok := effectstest.BattlefieldCard(g, small); ok {
		t.Error("a 2/2 survived 4 damage")
	}
	if c, ok := effectstest.BattlefieldCard(g, hit); !ok || c.DamageMarked != 4 {
		t.Errorf("the 5/5 took %v damage", c.DamageMarked)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Saved Giant") == uuid.Nil {
		t.Error("the exiled creature did not return at the beginning of the next end step")
	}
}

func TestDftB13ExplosiveGetawayWithNoTargetStillSweeps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Explosive Getaway", "Sorcery", dftB13GetawayOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, mine); ok {
		t.Error("my own 2/2 survived the 4 damage")
	}
}

const dftB13MimeoplasmOracle = "2b8c2882-dbca-466a-a0af-dd1d65edfe7c"

func TestDftB13MimeoplasmExilesUpToXCreatureCardsForThreeCountersEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dftB13EmptyGraveyard(g, me)
	a := effectstest.DftABInGraveyard(me, "Bear", "Creature — Bear")
	b := effectstest.DftABInGraveyard(me, "Rat", "Creature — Rat")
	c := effectstest.DftABInGraveyard(me, "Ox", "Creature — Ox")
	rock := effectstest.DftABInGraveyard(me, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	ooze := effectstest.CastFromHandForTest(t, g, me, "Mimeoplasm, Revered One", "Legendary Creature — Ooze", "{X}{B}{G}{U}", dftB13MimeoplasmOracle,
		game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the creature cards to exile")
	}
	if effectstest.HasID(pick.ChooseCards, rock) || len(pick.ChooseCards) != 3 || pick.ChooseMax != 2 {
		t.Errorf("offered %d cards, up to %d: want the three creature cards, up to X=2", len(pick.ChooseCards), pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, ooze); !ok {
		t.Fatal("the 0/0 died before its counters arrived")
	}
	if got := effectstest.CountersOn(g, ooze, game.CounterPlusOne); got != 6 {
		t.Errorf("%d +1/+1 counters, want 6 (three for each of two cards)", got)
	}
	if p, tough := effectstest.PtOf(t, g, ooze); p != 6 || tough != 6 {
		t.Errorf("Mimeoplasm is %d/%d, want 6/6", p, tough)
	}
	if me.Graveyard.Contains(a) || me.Graveyard.Contains(b) || !me.Graveyard.Contains(c) {
		t.Error("the wrong cards were exiled")
	}
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Error("the named cards are not in exile")
	}
}

func TestDftB13MimeoplasmBecomesACopyOfACardExiledWithItExceptZeroZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dftB13EmptyGraveyard(g, me)
	bear := effectstest.DftABInGraveyard(me, "Grizzly Bears", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	ooze := effectstest.CastFromHandForTest(t, g, me, "Mimeoplasm, Revered One", "Legendary Creature — Ooze", "{X}{B}{G}{U}", dftB13MimeoplasmOracle,
		game.CastSpellParams{XValue: 1})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	// A creature card not exiled with it is not a legal target.
	other := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: other, Name: "Other", TypeLine: "Creature — Wolf", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}, Strict: true}); err == nil {
		t.Fatal("copied a card that was not exiled with it")
	}
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, Strict: true}); err != nil {
		t.Fatalf("copy ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, ooze)
	if !ok {
		t.Fatal("Mimeoplasm left the battlefield")
	}
	if c.Name != "Grizzly Bears" || !c.HasSubtype("Bear") {
		t.Errorf("Mimeoplasm is %q %q, want a Bear named Grizzly Bears", c.Name, c.TypeLine)
	}
	if p, tough := effectstest.PtOf(t, g, ooze); p != 3 || tough != 3 {
		t.Errorf("the copy is %d/%d, want 3/3 (0/0 with three counters)", p, tough)
	}
	// "has this ability": it can copy again.
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, Strict: true}); err != nil {
		t.Errorf("the copy lost the ability: %v", err)
	}
}

func TestDftB13MimeoplasmThatWasNotCastExilesNothingAndDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.DftABInGraveyard(me, "Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Mimeoplasm, Revered One", TypeLine: "Legendary Creature — Ooze",
			ManaCost: "{X}{B}{G}{U}", OracleID: dftB13MimeoplasmOracle, PrintedPTKnown: true, Owner: me.ID, Controller: me.ID})
	})
	if err := g.ReturnFromGraveyardForEffect(id, game.ZoneBattlefield); err != nil {
		t.Fatalf("return: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked which cards to exile with X = 0")
	}
	g.RunStateChecksForTest()
	if _, ok := effectstest.BattlefieldCard(g, id); ok {
		t.Error("a 0/0 with no counters survived")
	}
}
