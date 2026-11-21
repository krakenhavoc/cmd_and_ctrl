package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b8_cards_test.go — Tarkir: Dragonstorm leftovers, batch 8 (#4463,
// #4452). Each test plays the real catalog entry through the engine.

const (
	stdB8SibsigOracle     = "bbdb731d-c533-48ef-b82a-cf1a8ba36208"
	stdB8SidisiOracle     = "72074972-418d-4dce-83cf-ce8cf310fc42"
	stdB8WailOracle       = "fc0abdf3-3a88-4cb2-ba79-35fcaebbf7d2"
	stdB8TombguardOracle  = "2321b4fc-6b73-4d1a-84cc-13524adb0899"
	stdB8DracogenesisOrcl = "9b06f2bf-146d-4b30-b467-e16f1a942253"
	stdB8AdeptOracle      = "42bf9435-02e6-428f-a93c-fe4652deb4fc"
	stdB8SummonsOracle    = "1fd78bf7-43ba-4e09-834b-14bfc1fd9d0d"
	stdB8WarEffortOracle  = "d9bc6841-9f71-46af-ac1f-a01caee290c1"
	stdB8RelicOracle      = "61822bbe-0407-4425-b8d7-0263a8c521ec"
	stdB8RiteOracle       = "6512ddb8-0868-4d32-ab32-8a5250fc571b"
)

func TestStdB8EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB8SibsigOracle:     "The Sibsig Ceremony",
		stdB8SidisiOracle:     "Sidisi, Regent of the Mire",
		stdB8WailOracle:       "Wail of War",
		stdB8TombguardOracle:  "Yathan Tombguard",
		stdB8DracogenesisOrcl: "Dracogenesis",
		stdB8AdeptOracle:      "Equilibrium Adept",
		stdB8SummonsOracle:    "Reverberating Summons",
		stdB8WarEffortOracle:  "War Effort",
		stdB8RelicOracle:      "Dragonbroods' Relic",
		stdB8RiteOracle:       "Rite of Renewal",
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

// --- The Sibsig Ceremony -----------------------------------------------------

func TestStdB8SibsigCeremonyDiscountsCreatureSpellsAndReplacesWhatYouCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "The Sibsig Ceremony", "Legendary Enchantment", stdB8SibsigOracle, 0, 0)
	effectstest.ToMainForCost(t, g)
	// {2}{B}{B} costs {B}{B} with the discount: a pool of exactly {B}{B}
	// casts it strictly.
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Costly Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{B}{B}",
		Colors: []string{"B"}, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "B", "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a creature spell did not cost {2} less: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("the creature you cast survived its own trigger")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the destroyed creature is not in the graveyard")
	}
	druids := effectstest.SocB8Named(g, me.ID, "Zombie Druid")
	if len(druids) != 1 || druids[0].Power != 2 || druids[0].Toughness != 2 {
		t.Errorf("Zombie Druids %+v, want one 2/2", druids)
	}
}

func TestStdB8SibsigCeremonyIgnoresCreaturesThatWereNotCastByYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "The Sibsig Ceremony", "Legendary Enchantment", stdB8SibsigOracle, 0, 0)
	effectstest.ToMainForCost(t, g)
	// Put onto the battlefield, never cast.
	walked := effectstest.B12Creature(g, me.ID, "Put In Play", "Creature — Bear", 2, 2)
	// A token.
	g.WithWriteLock(func() {
		if err := (effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1}).Apply(effects.NewContext(g,
			&game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID})); err != nil {
			t.Fatalf("token: %v", err)
		}
	})
	// An opponent's creature spell.
	if !g.Battlefield.Contains(walked) {
		t.Fatal("the creature put in play was destroyed")
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.SocB8Named(g, me.ID, "Zombie Druid")) != 0 {
		t.Error("a creature that was not cast made a Zombie Druid")
	}
	_ = opp
}

func TestStdB8SibsigCeremonyDoesNotDiscountOtherSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "The Sibsig Ceremony", "Legendary Enchantment", stdB8SibsigOracle, 0, 0)
	effectstest.ToMainForCost(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Costly Charm", TypeLine: "Sorcery", ManaCost: "{2}{B}",
		Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a sorcery was discounted")
	}
}

// --- Sidisi, Regent of the Mire ----------------------------------------------

func stdB8SidisiBoard(t *testing.T) (g *game.Game, me *game.Player, sidisi uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	sidisi = effectstest.B12Push(g, me.ID, "Sidisi, Regent of the Mire", "Legendary Creature — Zombie Snake Warlock", stdB8SidisiOracle, 1, 3)
	effectstest.AdvanceToMain(t, g)
	return g, me, sidisi
}

func stdB8GraveyardCreature(me *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", ManaCost: cost,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	return id
}

func TestStdB8SidisiSacrificesAnXCreatureToReturnAnXPlusOneCreature(t *testing.T) {
	g, me, sidisi := stdB8SidisiBoard(t)
	fodder := effectstest.PushTypedCard(g, me.ID, "Two Drop", "Creature — Bear", "{1}{B}")
	three := stdB8GraveyardCreature(me, "Three Drop", "{2}{B}")
	if err := g.ActivateCatalogAbility(me.ID, sidisi, 0, game.ActivateAbilityParams{
		XValue: 2, SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.B16TargetCard(three),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) || !effectstest.B16Tapped(t, g, sidisi) {
		t.Error("the cost did not sacrifice the creature and tap Sidisi")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(three) {
		t.Error("the mana value 3 creature card did not return to the battlefield")
	}
}

func TestStdB8SidisiRefusesTheWrongManaValues(t *testing.T) {
	g, me, sidisi := stdB8SidisiBoard(t)
	fodder := effectstest.PushTypedCard(g, me.ID, "Two Drop", "Creature — Bear", "{1}{B}")
	two := stdB8GraveyardCreature(me, "Two Drop Again", "{1}{B}")
	three := stdB8GraveyardCreature(me, "Three Drop", "{2}{B}")
	four := stdB8GraveyardCreature(me, "Four Drop", "{3}{B}")
	for name, params := range map[string]game.ActivateAbilityParams{
		"a target with mana value X, not X plus 1":                   {XValue: 2, SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.B16TargetCard(two)},
		"a target with mana value X plus 2":                          {XValue: 2, SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.B16TargetCard(four)},
		"an announced X that is not the sacrificed creature's value": {XValue: 3, SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.B16TargetCard(four)},
		"Sidisi sacrificing herself":                                 {XValue: 0, SacrificeIDs: []uuid.UUID{sidisi}, Targets: effectstest.B16TargetCard(stdB8GraveyardCreature(me, "One Drop", "{B}"))},
	} {
		if err := g.ActivateCatalogAbility(me.ID, sidisi, 0, params); err == nil {
			t.Errorf("%s was accepted", name)
		}
		if !g.Battlefield.Contains(fodder) || effectstest.B16Tapped(t, g, sidisi) {
			t.Fatalf("%s: a refused activation paid its cost", name)
		}
	}
	_ = three
}

func TestStdB8SidisiANoncreatureCardIsNotATarget(t *testing.T) {
	g, me, sidisi := stdB8SidisiBoard(t)
	fodder := effectstest.PushTypedCard(g, me.ID, "Two Drop", "Creature — Bear", "{1}{B}")
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Three Mana Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{B}", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, sidisi, 0, game.ActivateAbilityParams{
		XValue: 2, SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.B16TargetCard(spell),
	}); err == nil {
		t.Error("a sorcery card was targeted")
	}
}

func TestStdB8SidisiATokenIsManaValueZeroAndXInTheCostIsZero(t *testing.T) {
	g, me, sidisi := stdB8SidisiBoard(t)
	g.WithWriteLock(func() {
		if err := (effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1}).Apply(effects.NewContext(g,
			&game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID})); err != nil {
			t.Fatalf("token: %v", err)
		}
	})
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" {
			token = c.InstanceID
		}
	}
	xCost := stdB8GraveyardCreature(me, "Hydra", "{X}{G}{G}") // mana value 2 in a graveyard (X is 0)
	one := stdB8GraveyardCreature(me, "One Drop", "{B}")
	if err := g.ActivateCatalogAbility(me.ID, sidisi, 0, game.ActivateAbilityParams{
		XValue: 0, SacrificeIDs: []uuid.UUID{token}, Targets: effectstest.B16TargetCard(xCost),
	}); err == nil {
		t.Error("a mana value 2 card was a legal target for X = 0")
	}
	if err := g.ActivateCatalogAbility(me.ID, sidisi, 0, game.ActivateAbilityParams{
		XValue: 0, SacrificeIDs: []uuid.UUID{token}, Targets: effectstest.B16TargetCard(one),
	}); err != nil {
		t.Fatalf("sacrificing a token for X = 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(one) {
		t.Error("the mana value 1 creature did not return")
	}
}

func TestStdB8SidisiIsSorcerySpeed(t *testing.T) {
	g, me, sidisi := stdB8SidisiBoard(t)
	fodder := effectstest.PushTypedCard(g, me.ID, "Two Drop", "Creature — Bear", "{1}{B}")
	three := stdB8GraveyardCreature(me, "Three Drop", "{2}{B}")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, sidisi, 0, game.ActivateAbilityParams{
		XValue: 2, SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.B16TargetCard(three),
	}); err == nil {
		t.Error("activated in combat")
	}
}

// --- Wail of War -------------------------------------------------------------

func TestStdB8WailOfWarShrinksTheOpponentsCreaturesUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	runt := effectstest.B12Creature(g, opp.ID, "Runt", "Creature — Bear", 1, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	bystander := effectstest.B12Creature(g, other.ID, "Bystander", "Creature — Bear", 1, 1)
	if err := effectstest.RfCastModal(t, g, "Wail of War", stdB8WailOracle, []int{0}, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(runt) {
		t.Error("the 1/1 survived -1/-1")
	}
	if p, tough := effectstest.MkmMPPT(t, g, big); p != 2 || tough != 2 {
		t.Errorf("the 3/3 is %d/%d, want 2/2", p, tough)
	}
	for _, id := range []uuid.UUID{mine, bystander} {
		if !g.Battlefield.Contains(id) {
			t.Error("a creature the targeted opponent does not control was shrunk")
		}
	}
}

func TestStdB8WailOfWarReturnsUpToTwoCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := stdB8GraveyardCreature(me, "A", "{1}")
	b := stdB8GraveyardCreature(me, "B", "{2}")
	c := stdB8GraveyardCreature(me, "C", "{3}")
	if err := effectstest.RfCastModal(t, g, "Wail of War", stdB8WailOracle, []int{1}, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) || me.Hand.Contains(c) {
		t.Error("exactly the two targeted creature cards should return to the hand")
	}
}

func TestStdB8WailOfWarCanChooseNoCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := effectstest.RfCastModal(t, g, "Wail of War", stdB8WailOracle, []int{1}, nil); err != nil {
		t.Fatalf("up to two means none is legal: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Yathan Tombguard --------------------------------------------------------

func TestStdB8YathanTombguardDrawsAndLosesLifeForACreatureWithACounterOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Yathan Tombguard", "Creature — Human Warrior", stdB8TombguardOracle, 2, 3)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	counter := effectstest.B12Creature(g, me.ID, "Countered", "Creature — Bear", 2, 2)
	stun := effectstest.B12Creature(g, me.ID, "Stunned Elsewhere", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID})
		if err := (effects.AddCounter{Target: counter, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			t.Fatalf("counter: %v", err)
		}
		if err := (effects.AddCounter{Target: stun, Kind: "charge", N: 1}).Apply(ctx); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	hand, life := me.Hand.Size(), me.Life
	effectstest.DeclareAttack(t, g, opp.ID, plain, counter, stun)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2 (the two creatures with counters)", hand, got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, me.Life)
	}
}

func TestStdB8YathanTombguardIgnoresCreaturesYouDoNotControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Yathan Tombguard", "Creature — Human Warrior", stdB8TombguardOracle, 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Countered", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := (effects.AddCounter{Target: theirs, Kind: game.CounterPlusOne, N: 1}).Apply(effects.NewContext(g,
			&game.StackItem{Kind: game.StackItemSpell, Controller: opp.ID, Owner: opp.ID})); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	hand := me.Hand.Size()
	g.Turn.ActiveSeat = 1
	effectstest.DeclareAttack(t, g, me.ID, theirs)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: an opponent's creature triggered it", hand, me.Hand.Size())
	}
}

// --- Dracogenesis ------------------------------------------------------------

func TestStdB8DracogenesisCastsDragonSpellsForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Dracogenesis", "Enchantment", stdB8DracogenesisOrcl, 0, 0)
	effectstest.ToMainForCost(t, g)
	dragon := effectstest.HandCardFull(me, "Big Dragon", "Creature — Dragon", "{4}{R}{R}", "", []string{"R"})
	bear := effectstest.HandCardFull(me, "Big Bear", "Creature — Bear", "{4}{R}{R}", "", []string{"R"})
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, dragon, game.ZoneHand), effects.GrantedAltCostFreeDragonSpells) == nil {
		t.Fatal("no free offer for a Dragon spell")
	}
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, bear, game.ZoneHand), effects.GrantedAltCostFreeDragonSpells) != nil {
		t.Error("a non-Dragon spell was offered free")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostFreeDragonSpells}); err == nil {
		t.Error("a Bear was cast for free")
	}
	if err := g.CastSpell(me.ID, dragon, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostFreeDragonSpells}); err != nil {
		t.Fatalf("the Dragon was not cast for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) {
		t.Error("the Dragon did not resolve")
	}
}

func TestStdB8DracogenesisIsOnlyForTheControllersSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Dracogenesis", "Enchantment", stdB8DracogenesisOrcl, 0, 0)
	effectstest.ToMainForCost(t, g)
	theirs := effectstest.HandCardFull(opp, "Their Dragon", "Creature — Dragon", "{4}{R}{R}", "", []string{"R"})
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, opp.ID, theirs, game.ZoneHand), effects.GrantedAltCostFreeDragonSpells) != nil {
		t.Error("an opponent's Dragon was offered free")
	}
}

// --- Equilibrium Adept -------------------------------------------------------

func TestStdB8EquilibriumAdeptExilesTheTopCardPlayableUntilTheEndOfYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Exiled Bolt", "L1", "L2", "L3", "L4", "L5", "L6")
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Equilibrium Adept", TypeLine: "Creature — Dog Monk", OracleID: stdB8AdeptOracle, Power: 2, Toughness: 4})
	if !effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Fatal("the exiled card is not playable by me")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if !effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the window closed before the end of my next turn")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	if !effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the window closed during my next turn")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the card was still playable after my next turn ended")
	}
}

func TestStdB8EquilibriumAdeptFlurryGivesDoubleStrikeOnTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	adept := effectstest.B12Push(g, me.ID, "Equilibrium Adept", "Creature — Dog Monk", stdB8AdeptOracle, 2, 4)
	effectstest.ToMainForCost(t, g)
	cast := func(name string) {
		id := effectstest.HandCardFull(me, name, "Instant", "", "", []string{"R"})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	cast("First")
	if effectstest.HasEffectiveKeyword(t, g, adept, "double strike") {
		t.Fatal("double strike after the first spell")
	}
	cast("Second")
	if !effectstest.HasEffectiveKeyword(t, g, adept, "double strike") {
		t.Fatal("no double strike after the second spell")
	}
	cast("Third")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, adept, "double strike") {
		t.Error("double strike outlasted the turn")
	}
}

// --- Reverberating Summons ---------------------------------------------------

func stdB8CastFree(t *testing.T, g *game.Game, me *game.Player, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		id := effectstest.HandCardFull(me, "Cantrip", "Instant", "", "", []string{"R"})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
}

func TestStdB8ReverberatingSummonsBecomesAMonkAtCombatAfterTwoSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	summons := effectstest.B12Push(g, me.ID, "Reverberating Summons", "Enchantment", stdB8SummonsOracle, 0, 0)
	effectstest.ToMainForCost(t, g)
	stdB8CastFree(t, g, me, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.TmtB1Card(t, g, summons)
	if !c.IsCreature() || !c.HasSubtype("Monk") || !c.IsEnchantment() || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Fatalf("Summons is %q %d/%d, want a 3/3 Monk enchantment creature", c.TypeLine, c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, summons, "haste") {
		t.Error("the Summons has no haste")
	}
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.TmtB1Card(t, g, summons).IsCreature() {
		t.Error("the Summons is still a creature after the turn ended")
	}
}

func TestStdB8ReverberatingSummonsNeedsTwoSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	summons := effectstest.B12Push(g, me.ID, "Reverberating Summons", "Enchantment", stdB8SummonsOracle, 0, 0)
	effectstest.ToMainForCost(t, g)
	stdB8CastFree(t, g, me, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Card(t, g, summons).IsCreature() {
		t.Error("one spell animated the Summons")
	}
}

func TestStdB8ReverberatingSummonsDiscardsYourHandAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	summons := effectstest.B12Push(g, me.ID, "Reverberating Summons", "Enchantment", stdB8SummonsOracle, 0, 0)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.ToMainForCost(t, g)
	effectstest.FillPoolColored(me, "R", 2)
	if err := g.ActivateCatalogAbility(me.ID, summons, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Size() != 0 || g.Battlefield.Contains(summons) {
		t.Errorf("hand %d, on battlefield %v: the cost was not paid", me.Hand.Size(), g.Battlefield.Contains(summons))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 2 {
		t.Errorf("hand %d, want 2 drawn", me.Hand.Size())
	}
}

// --- War Effort --------------------------------------------------------------

func TestStdB8WarEffortGivesPlusOneAndAWarriorEachAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "War Effort", "Enchantment", stdB8WarEffortOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if p, _ := effectstest.MkmMPPT(t, g, bear); p != 3 {
		t.Errorf("the Bear is %d power, want 3", p)
	}
	if p, _ := effectstest.MkmMPPT(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature is %d power, want 2", p)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	warriors := effectstest.SocB8Named(g, me.ID, "Warrior")
	if len(warriors) != 1 {
		t.Fatalf("%d Warriors, want 1 (one trigger per attack)", len(warriors))
	}
	w := warriors[0]
	if !w.Tapped || w.AttackingTarget == uuid.Nil || w.CurrentPower() != 2 || w.Power != 1 {
		t.Errorf("Warrior tapped=%v attacking=%v power=%d, want tapped, attacking, 1/1 plus the anthem", w.Tapped, w.AttackingTarget, w.CurrentPower())
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.SocB8Named(g, me.ID, "Warrior")); got != 0 {
		t.Errorf("%d Warriors after the end step, want it sacrificed", got)
	}
}

func TestStdB8WarEffortMakesOneWarriorForAWideAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "War Effort", "Enchantment", stdB8WarEffortOracle, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b, c)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	if got := len(effectstest.SocB8Named(g, me.ID, "Warrior")); got != 1 {
		t.Errorf("%d Warriors for a three-creature attack, want 1", got)
	}
}

// --- Dragonbroods' Relic -----------------------------------------------------

func TestStdB8DragonbroodsRelicTapsACreatureForManaOfAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.B12Push(g, me.ID, "Dragonbroods' Relic", "Artifact", stdB8RelicOracle, 0, 0)
	fresh := effectstest.PushVanillaCreature(g, me.ID, "Fresh Bear", 2, 2)
	if err := g.ActivateManaAbility(me.ID, relic, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the mana ability worked without tapping a creature")
	}
	if err := g.ActivateManaAbility(me.ID, relic, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{fresh}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if !effectstest.B20Tapped(t, g, relic) || !effectstest.B20Tapped(t, g, fresh) {
		t.Error("both the artifact and the creature should be tapped")
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("color choice = %+v, want five colors", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
}

func TestStdB8DragonbroodsRelicMakesAnAllColorsReliquaryDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	relic := effectstest.B12Push(g, me.ID, "Dragonbroods' Relic", "Artifact", stdB8RelicOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		effectstest.FillPoolColored(me, c, 1)
	}
	effectstest.FillPoolColored(me, "C", 3)
	oppLife, life := opp.Life, me.Life
	if err := g.ActivateCatalogAbility(me.ID, relic, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(relic) {
		t.Error("the Relic was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pickTarget := effectstest.LatestPickTarget(g, me.ID)
	if pickTarget == nil {
		t.Fatal("the token's enters trigger asked for no target")
	}
	if err := g.ResolvePickTarget(pickTarget.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	dragons := effectstest.SocB8Named(g, me.ID, "Reliquary Dragon")
	if len(dragons) != 1 {
		t.Fatalf("%d Reliquary Dragons, want 1", len(dragons))
	}
	d := dragons[0]
	if d.Power != 4 || d.Toughness != 4 || len(d.Colors) != 5 || !d.HasSubtype("Dragon") {
		t.Errorf("the token is %d/%d %v %q, want a 4/4 five-color Dragon", d.Power, d.Toughness, d.Colors, d.TypeLine)
	}
	for _, kw := range []string{"flying", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, d.InstanceID, kw) {
			t.Errorf("the token has no %s", kw)
		}
	}
	if opp.Life != oppLife-3 {
		t.Errorf("the target took %d, want 3", oppLife-opp.Life)
	}
	if me.Life != life+3 {
		t.Errorf("lifelink gained %d, want 3 (the token dealt the damage)", me.Life-life)
	}
}

func TestStdB8DragonbroodsRelicTokenIsSorcerySpeedAndNeedsAllFiveColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := effectstest.B12Push(g, me.ID, "Dragonbroods' Relic", "Artifact", stdB8RelicOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "C", 8)
	if err := g.ActivateCatalogAbility(me.ID, relic, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the Dragon was made without the five colors")
	}
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		effectstest.FillPoolColored(me, c, 1)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, relic, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated in combat")
	}
}

// --- Rite of Renewal ---------------------------------------------------------

func TestStdB8RiteOfRenewalReturnsPermanentsShufflesAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	creature := stdB8GraveyardCreature(me, "Dead Bear", "{1}")
	relic := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: relic, Name: "Old Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	sorcery := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: sorcery, Name: "Old Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	var theirs []uuid.UUID
	for _, n := range []string{"T1", "T2", "T3"} {
		id := uuid.New()
		opp.Graveyard.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
		theirs = append(theirs, id)
	}
	lib := opp.Library.Size()
	id := effectstest.CastCatalogSpell(t, g, "Rite of Renewal", "Sorcery", stdB8RiteOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: creature, Slot: 0}, {Kind: game.TargetCard, ID: relic, Slot: 0},
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 1},
		{Kind: game.TargetCard, ID: theirs[0], Slot: 2}, {Kind: game.TargetCard, ID: theirs[1], Slot: 2},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(creature) || !me.Hand.Contains(relic) || me.Hand.Contains(sorcery) {
		t.Error("the two permanent cards, and only they, should return to the hand")
	}
	if opp.Library.Size() != lib+2 || opp.Graveyard.Contains(theirs[0]) || opp.Graveyard.Contains(theirs[1]) || !opp.Graveyard.Contains(theirs[2]) {
		t.Errorf("opponent library %d (want %d), the picked cards should be in it and the third still in the graveyard", opp.Library.Size(), lib+2)
	}
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Errorf("exiled=%v in graveyard=%v: Rite of Renewal exiles itself", g.Exile.Contains(id), me.Graveyard.Contains(id))
	}
}

func TestStdB8RiteOfRenewalWorksWithNoCardsTargetedAndOnYourself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "Mine", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	id := effectstest.CastCatalogSpell(t, g, "Rite of Renewal", "Sorcery", stdB8RiteOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 1},
		{Kind: game.TargetCard, ID: mine, Slot: 2},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(mine) {
		t.Error("the card did not leave the graveyard")
	}
	if !g.Exile.Contains(id) {
		t.Error("Rite of Renewal did not exile itself")
	}
}

func TestStdB8RiteOfRenewalCardsMustBeInTheTargetPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	elsewhere := uuid.New()
	other.Graveyard.PushTop(game.Card{InstanceID: elsewhere, Name: "Elsewhere", TypeLine: "Creature — Bear", Owner: other.ID, Controller: other.ID})
	if err := effectstest.CastCatalogSpellErr(t, g, "Rite of Renewal", "Sorcery", stdB8RiteOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 1},
		{Kind: game.TargetCard, ID: elsewhere, Slot: 2},
	}); err == nil {
		t.Error("a card from another player's graveyard was accepted")
	}
	_ = me
}
