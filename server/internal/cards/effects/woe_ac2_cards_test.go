package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_ac2_cards_test.go — Wilds of Eldraine, the second A–C batch (#4471).
// Each test plays the real catalog entry and asserts what its text says.

const (
	woeAC2AgathaOracle    = "f398d222-f8e7-4983-9a77-267adef7f551"
	woeAC2ChampionOracle  = "debdfab7-454f-4fec-9ee8-aa5b570ea7fb"
	woeAC2AlchemistOracle = aquaticAlchemistOracleID
	woeAC2ArchonOracle    = "180efbd4-0f77-41cc-8ae2-2e6ccf853d10"
	woeAC2GloryOracle     = "bc5631ae-494e-4987-8d17-57674bfbdb9f"
	woeAC2ChillOracle     = "be0dcad2-8e84-4e9f-b69c-837b40b8fb83"
	woeAC2RangerOracle    = "1189b5d5-ca8f-4b4e-a5f1-91122a485cf3"
	woeAC2FamiliarOracle  = brambleFamiliarOracleID
	woeAC2BreakOracle     = "b9d51e07-851f-4b7b-bffe-7c1efff0714c"
	woeAC2SellSwordOracle = callousSellSwordOracleID
	woeAC2CandyOracle     = "7c8e4f29-53e0-4242-8ef8-7a9476dce178"
	woeAC2HooliganOracle  = "c808892a-83d5-4b09-b0da-118b0d6fe58e"
	woeAC2CoopedOracle    = "3e8bcb44-b875-4fe4-b96f-97a20cbfec93"
	woeAC2BearType        = "Creature — Bear"
	woeAC2AuraType        = "Enchantment — Aura"
	woeAC2EnchantmentType = "Enchantment"
)

func TestWoeAC2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		woeAC2AgathaOracle:           "Agatha of the Vile Cauldron",
		woeAC2ChampionOracle:         "Agatha's Champion",
		woeAC2AlchemistOracle:        "Aquatic Alchemist",
		woeAC2AlchemistOracle + "#1": "Bubble Up",
		woeAC2ArchonOracle:           "Archon of the Wild Rose",
		woeAC2GloryOracle:            "Archon's Glory",
		woeAC2ChillOracle:            "Bitter Chill",
		woeAC2RangerOracle:           "Boundary Lands Ranger",
		woeAC2FamiliarOracle:         "Bramble Familiar",
		woeAC2FamiliarOracle + "#1":  "Fetch Quest",
		woeAC2BreakOracle:            "Break the Spell",
		woeAC2SellSwordOracle:        "Callous Sell-Sword",
		woeAC2SellSwordOracle + "#1": "Burn Together",
		woeAC2CandyOracle:            "Candy Trail",
		woeAC2HooliganOracle:         "Charging Hooligan",
		woeAC2CoopedOracle:           "Cooped Up",
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

// woeAC2CastCreature casts a catalog creature from the active player's hand
// with a printed size, in a main phase.
func woeAC2CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// woeAC2Aura puts a non-catalog Aura controlled by owner onto host.
func woeAC2Aura(t *testing.T, g *game.Game, owner, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := b12Permanent(g, owner, "Some Aura", woeAC2AuraType)
	woeB1Attach(t, g, aura, host)
	return aura
}

// --- Agatha of the Vile Cauldron ---------------------------------------------------

func TestWoeAC2AgathaDiscountsByHerPowerButLeavesOneMana(t *testing.T) {
	g, me, opp := spendTable(t)
	bear := b12Creature(g, me.ID, "Bear", woeAC2BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeAC2BearType, 2, 2)
	agatha := b12Push(g, me.ID, "Agatha of the Vile Cauldron", "Legendary Creature — Human Warlock", woeAC2AgathaOracle, 1, 1)
	for _, tc := range []struct{ cost, want string }{
		{"{2}{G}", "{1}{G}"},
		{"{1}", "{1}"},
		{"{R}{R}", "{R}{R}"},
	} {
		if got := lcPrice(t, g, me.ID, bear, tc.cost, 0); got != tc.want {
			t.Errorf("power 1: %s costs %s, want %s", tc.cost, got, tc.want)
		}
	}
	if got := lcPrice(t, g, opp.ID, theirs, "{3}", 0); got != "{3}" {
		t.Errorf("an opponent's creature's {3} costs %s, want {3}", got)
	}
	addCountersForTest(g, agatha, game.CounterPlusOne, 3)
	for _, tc := range []struct{ cost, want string }{
		{"{5}{G}", "{1}{G}"},
		{"{2}{G}", "{G}"},
		{"{4}", "{1}"},
	} {
		if got := lcPrice(t, g, me.ID, bear, tc.cost, 0); got != tc.want {
			t.Errorf("power 4: %s costs %s, want %s", tc.cost, got, tc.want)
		}
	}
}

func TestWoeAC2AgathaPumpsHerOtherCreatures(t *testing.T) {
	g, me, opp := spendTable(t)
	agatha := b12Push(g, me.ID, "Agatha of the Vile Cauldron", "Legendary Creature — Human Warlock", woeAC2AgathaOracle, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", woeAC2BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeAC2BearType, 2, 2)
	// Agatha's own ability is one of a creature you control: {4}{R}{G}
	// less one is {3}{R}{G}, five mana.
	floatForTest(g, me, "RGCCC")
	if err := g.ActivateCatalogAbility(me.ID, agatha, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate for five mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
	}
	assertKeywords(t, g, bear, "trample", "haste")
	if p, tough := effectivePTOf(t, g, agatha); p != 1 || tough != 1 {
		t.Errorf("Agatha is %d/%d, want 1/1 (other creatures only)", p, tough)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("the opponent's Bear has power %d, want 2", p)
	}
}

// --- Agatha's Champion -------------------------------------------------------------

func TestWoeAC2AgathasChampionBargainedFightsACreatureYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	token := woeILToken(g, me.ID)
	foe := b12Creature(g, opp.ID, "Foe", woeAC2BearType, 3, 3)
	champ := woeAC2CastCreature(t, g, "Agatha's Champion", "Creature — Human Knight", woeAC2ChampionOracle, 4, 4, game.CastSpellParams{
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) {
		t.Error("the 3/3 survived a fight with the 4/4")
	}
	if !g.Battlefield.Contains(champ) {
		t.Fatal("the Champion died")
	}
	assertKeywords(t, g, champ, "trample")
}

func TestWoeAC2AgathasChampionUnbargainedDoesNotFight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	foe := b12Creature(g, opp.ID, "Foe", woeAC2BearType, 3, 3)
	woeAC2CastCreature(t, g, "Agatha's Champion", "Creature — Human Knight", woeAC2ChampionOracle, 4, 4, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an unbargained Champion asked for a target")
	}
	if !g.Battlefield.Contains(foe) {
		t.Error("an unbargained Champion fought")
	}
}

// --- Aquatic Alchemist // Bubble Up -------------------------------------------------

func TestWoeAC2AquaticAlchemistPumpsOnlyOnTheFirstInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	alch := b12Push(g, me.ID, "Aquatic Alchemist", "Creature — Elemental", woeAC2AlchemistOracle, 1, 3)
	castCatalogSpell(t, g, "Plain Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, alch); p != 3 {
		t.Fatalf("after the first instant the Alchemist has power %d, want 3", p)
	}
	castCatalogSpell(t, g, "Second Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, alch); p != 3 {
		t.Errorf("after the second instant the Alchemist has power %d, want 3 (first each turn only)", p)
	}
}

func TestWoeAC2BubbleUpPutsAnInstantOnTopOfYourLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	spell := pushGraveyardCardTyped(me, "Old Bolt", "Instant")
	card := woeB1Adventure(me, woeAC2AlchemistOracle,
		game.Face{Name: "Aquatic Alchemist", TypeLine: "Creature — Elemental", ManaCost: "{1}{U}", Colors: []string{"U"}},
		game.Face{Name: "Bubble Up", TypeLine: "Sorcery — Adventure", ManaCost: "{2}{U}", Colors: []string{"U"}})
	woeB1CastAdventure(t, g, me, card, cardRefs(spell))
	if top, err := me.Library.Top(); err != nil || top.InstanceID != spell {
		t.Error("the instant is not on top of your library")
	}
}

// --- Archon of the Wild Rose -------------------------------------------------------

func TestWoeAC2ArchonOfTheWildRoseMakesYourEnchantedCreaturesFourFourFliers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	archon := b12Push(g, me.ID, "Archon of the Wild Rose", "Creature — Archon", woeAC2ArchonOracle, 4, 4)
	mine := b12Creature(g, me.ID, "Mine", woeAC2BearType, 1, 1)
	theirAura := b12Creature(g, me.ID, "Theirs On Mine", woeAC2BearType, 1, 1)
	bare := b12Creature(g, me.ID, "Bare", woeAC2BearType, 1, 1)
	opposing := b12Creature(g, opp.ID, "Opposing", woeAC2BearType, 1, 1)
	woeAC2Aura(t, g, me.ID, mine)
	woeAC2Aura(t, g, opp.ID, theirAura)
	woeAC2Aura(t, g, me.ID, opposing)
	woeAC2Aura(t, g, me.ID, archon)

	if p, tough := effectivePTOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("my creature with my Aura is %d/%d, want 4/4", p, tough)
	}
	assertKeywords(t, g, mine, "flying")
	for _, id := range []uuid.UUID{theirAura, bare, opposing} {
		if p, tough := effectivePTOf(t, g, id); p != 1 || tough != 1 {
			t.Errorf("creature %s is %d/%d, want 1/1", id, p, tough)
		}
		if hasEffectiveKeyword(t, g, id, "flying") {
			t.Errorf("creature %s has flying", id)
		}
	}
}

// --- Archon's Glory ----------------------------------------------------------------

func TestWoeAC2ArchonsGloryPumpsAndGrantsOnlyWhenBargained(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "Bear", woeAC2BearType, 2, 2)
	castCatalogSpell(t, g, "Archon's Glory", "Instant", woeAC2GloryOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	if hasEffectiveKeyword(t, g, bear, "flying") || hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("an unbargained Glory granted flying or lifelink")
	}

	token := woeILToken(g, me.ID)
	elk := b12Creature(g, me.ID, "Elk", woeAC2BearType, 2, 2)
	p1g2Cast(t, g, "Archon's Glory", "Instant", woeAC2GloryOracle, "", game.CastSpellParams{
		Targets: cardRefs(elk), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, elk); p != 4 || tough != 4 {
		t.Errorf("the Elk is %d/%d, want 4/4", p, tough)
	}
	assertKeywords(t, g, elk, "flying", "lifelink")
}

// --- Bitter Chill ------------------------------------------------------------------

func TestWoeAC2BitterChillTapsAndPaysToScryAndDrawWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	foe := b12Creature(g, opp.ID, "Foe", woeAC2BearType, 2, 2)
	seedLibrary(me, "A", "B", "C")
	chill := castCatalogSpell(t, g, "Bitter Chill", woeAC2AuraType, woeAC2ChillOracle, cardRefs(foe))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(chill) {
		t.Fatal("Bitter Chill did not resolve")
	}
	if !isTapped(g, foe) {
		t.Fatal("the enchanted creature was not tapped")
	}

	hand := me.Hand.Size()
	woeB1Destroy(t, g, chill)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no {1} offer when Bitter Chill went to the graveyard")
	}
	fillPoolColored(me, "U", 1)
	answerPayUnless(t, g, me.ID, true)
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d after paying, want one draw", hand, got)
	}
}

func TestWoeAC2BitterChillDeclinedDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	foe := b12Creature(g, opp.ID, "Foe", woeAC2BearType, 2, 2)
	chill := castCatalogSpell(t, g, "Bitter Chill", woeAC2AuraType, woeAC2ChillOracle, cardRefs(foe))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	woeB1Destroy(t, g, chill)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil || me.Hand.Size() != hand {
		t.Errorf("declining scried or drew: hand %d -> %d", hand, me.Hand.Size())
	}
}

// --- Boundary Lands Ranger ---------------------------------------------------------

func TestWoeAC2BoundaryLandsRangerRummagesWithAPowerFourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	b12Push(g, me.ID, "Boundary Lands Ranger", "Creature — Human Ranger", woeAC2RangerOracle, 2, 2)
	b12Creature(g, me.ID, "Giant", "Creature — Giant", 4, 4)
	pushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Instant"})
	fodder := handCard(me, "Fodder", "Instant")
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard offer at the beginning of combat")
	}
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(fodder) || me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: one discarded and one drawn should leave it the same", hand, me.Hand.Size())
	}
}

func TestWoeAC2BoundaryLandsRangerDoesNothingWithoutPowerFour(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	b12Push(g, me.ID, "Boundary Lands Ranger", "Creature — Human Ranger", woeAC2RangerOracle, 2, 2)
	b12Creature(g, me.ID, "Bear", woeAC2BearType, 3, 3)
	handCard(me, "Fodder", "Instant")
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("a discard was offered with no creature of power 4 or greater")
	}
}

// --- Bramble Familiar // Fetch Quest ------------------------------------------------

func TestWoeAC2BrambleFamiliarDiscardsToReturnToHand(t *testing.T) {
	g, me, _ := spendTable(t)
	fam := b12Push(g, me.ID, "Bramble Familiar", "Creature — Elemental Raccoon", woeAC2FamiliarOracle, 2, 2)
	fodder := handCard(me, "Fodder", "Instant")
	floatForTest(g, me, "GC")
	if err := g.ActivateCatalogAbility(me.ID, fam, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Contains(fodder) {
		t.Fatal("the discard was not paid")
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(fam) {
		t.Error("Bramble Familiar did not return to hand")
	}
}

func TestWoeAC2FetchQuestPutsAMilledPermanentCardOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	var milled []uuid.UUID
	for _, c := range []struct{ name, typeLine string }{
		{"Instant A", "Instant"}, {"Sorcery B", "Sorcery"}, {"Instant C", "Instant"},
		{"Bear D", woeAC2BearType}, {"Instant E", "Instant"}, {"Glade F", "Land"}, {"Instant G", "Instant"},
	} {
		milled = append(milled, plTop(me, c.name, c.typeLine, ""))
	}
	bear, land := milled[3], milled[5]
	card := woeB1Adventure(me, woeAC2FamiliarOracle,
		game.Face{Name: "Bramble Familiar", TypeLine: "Creature — Elemental Raccoon", ManaCost: "{1}{G}", Colors: []string{"G"}},
		game.Face{Name: "Fetch Quest", TypeLine: "Sorcery — Adventure", ManaCost: "{5}{G}{G}", Colors: []string{"G"}})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Fetch Quest: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the milled cards")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[bear] || !offered[land] || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the creature and the land, exactly one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature card is not on the battlefield")
	}
	if !me.Graveyard.Contains(land) {
		t.Error("the land card left the graveyard")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("the Adventure card is not in exile after it resolved")
	}
}

// --- Break the Spell ---------------------------------------------------------------

func TestWoeAC2BreakTheSpellDrawsForYoursOrAToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	theirs := b12Permanent(g, opp.ID, "Their Enchantment", woeAC2EnchantmentType)
	mine := b12Permanent(g, me.ID, "My Enchantment", woeAC2EnchantmentType)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Token", TypeLine: "Token Enchantment",
		Owner: opp.ID, Controller: opp.ID,
	})

	for _, tc := range []struct {
		id    uuid.UUID
		draws int
		what  string
	}{
		{theirs, 0, "an opponent's nontoken enchantment"},
		{mine, 1, "my enchantment"},
		{token, 1, "an opponent's token enchantment"},
	} {
		hand := me.Hand.Size()
		castCatalogSpell(t, g, "Break the Spell", "Instant", woeAC2BreakOracle, cardRefs(tc.id))
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(tc.id) {
			t.Errorf("%s was not destroyed", tc.what)
		}
		if got := me.Hand.Size() - hand; got != tc.draws {
			t.Errorf("destroying %s drew %d, want %d", tc.what, got, tc.draws)
		}
	}
}

func TestWoeAC2BreakTheSpellDrawsNothingWhenNothingIsDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sturdy Enchantment", TypeLine: woeAC2EnchantmentType,
		Keywords: []string{"indestructible"}, Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Break the Spell", "Instant", woeAC2BreakOracle, cardRefs(mine))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || me.Hand.Size() != hand {
		t.Errorf("an indestructible enchantment: on battlefield %v, hand %d -> %d", g.Battlefield.Contains(mine), hand, me.Hand.Size())
	}
}

// --- Callous Sell-Sword // Burn Together --------------------------------------------

func TestWoeAC2CallousSellSwordEntersWithACounterPerDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	woeB1Destroy(t, g, b12Creature(g, me.ID, "Mine A", woeAC2BearType, 1, 1))
	woeB1Destroy(t, g, b12Creature(g, me.ID, "Mine B", woeAC2BearType, 1, 1))
	woeB1Destroy(t, g, b12Creature(g, opp.ID, "Theirs", woeAC2BearType, 1, 1))
	card := woeB1Adventure(me, woeAC2SellSwordOracle,
		game.Face{Name: "Callous Sell-Sword", TypeLine: "Creature — Human Soldier", ManaCost: "{1}{B}", Colors: []string{"B"}, Power: 2, Toughness: 2},
		game.Face{Name: "Burn Together", TypeLine: "Sorcery — Adventure", ManaCost: "{R}", Colors: []string{"R"}})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(card.InstanceID) {
		t.Fatal("Callous Sell-Sword did not resolve")
	}
	if n := plusOneCounters(g, card.InstanceID); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2 (two of my creatures died)", n)
	}
}

func TestWoeAC2BurnTogetherThrowsYourCreatureThenSacrificesIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	thrower := b12Creature(g, me.ID, "Thrower", woeAC2BearType, 3, 3)
	life := lifeOf(g, opp.ID)
	card := woeB1Adventure(me, woeAC2SellSwordOracle,
		game.Face{Name: "Callous Sell-Sword", TypeLine: "Creature — Human Soldier", ManaCost: "{1}{B}", Colors: []string{"B"}},
		game.Face{Name: "Burn Together", TypeLine: "Sorcery — Adventure", ManaCost: "{R}", Colors: []string{"R"}})
	woeB1CastAdventure(t, g, me, card, []game.TargetRef{
		{Kind: game.TargetCard, ID: thrower, Slot: 0},
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 1},
	})
	if got := lifeOf(g, opp.ID); got != life-3 {
		t.Errorf("opponent at %d, want %d", got, life-3)
	}
	if g.Battlefield.Contains(thrower) {
		t.Error("the thrower was not sacrificed")
	}
}

// --- Candy Trail -------------------------------------------------------------------

func TestWoeAC2CandyTrailScriesTwoThenSacrificesForLifeAndACard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C", "D")
	trail := castCatalogSpell(t, g, "Candy Trail", "Artifact — Food Clue", woeAC2CandyOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if c := scryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %v, want one looking at two cards", c)
	}
	answerScryKeepAll(t, g, me.ID)

	advanceTo(t, g, game.StepPostcombatMain)
	life, hand := lifeOf(g, me.ID), me.Hand.Size()
	fillPoolColored(me, "C", 2)
	if err := g.ActivateCatalogAbility(me.ID, trail, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(trail) {
		t.Error("Candy Trail was not sacrificed")
	}
	if got := lifeOf(g, me.ID); got != life+3 {
		t.Errorf("life %d -> %d, want +3", life, got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one draw", hand, got)
	}
}

// --- Charging Hooligan -------------------------------------------------------------

func TestWoeAC2ChargingHooliganPumpsPerAttackerAndTramplesWithARat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	hool := b12Push(g, me.ID, "Charging Hooligan", "Creature — Human Peasant", woeAC2HooliganOracle, 3, 3)
	rat := b12Creature(g, me.ID, "Rat", "Creature — Rat", 1, 1)
	b12Creature(g, me.ID, "Stays Home", woeAC2BearType, 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{hool, rat} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, hool); p != 5 || tough != 3 {
		t.Errorf("the Hooligan is %d/%d, want 5/3 (+1/+0 for each of two attackers)", p, tough)
	}
	assertKeywords(t, g, hool, "trample")
}

func TestWoeAC2ChargingHooliganAloneHasNoTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	hool := b12Push(g, me.ID, "Charging Hooligan", "Creature — Human Peasant", woeAC2HooliganOracle, 3, 3)
	lciB1Attack(t, g, hool, opp.ID)
	if p, _ := effectivePTOf(t, g, hool); p != 4 {
		t.Errorf("the Hooligan has power %d, want 4 (it counts itself)", p)
	}
	if hasEffectiveKeyword(t, g, hool, "trample") {
		t.Error("the Hooligan has trample with no Rat attacking")
	}
}

// --- Cooped Up ---------------------------------------------------------------------

func TestWoeAC2CoopedUpPacifiesThenExilesTheCreature(t *testing.T) {
	g, me, opp := spendTable(t)
	foe := b12Creature(g, opp.ID, "Foe", woeAC2BearType, 3, 3)
	aura := pushCatalogPermanent(g, me.ID, "Cooped Up", woeAC2AuraType, woeAC2CoopedOracle, false)
	woeB1Attach(t, g, aura, foe)
	assertRestrictions(t, g, foe, game.CantAttackOrBlock)
	floatForTest(g, me, "WCC")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(foe) {
		t.Error("the enchanted creature was not exiled")
	}
}
