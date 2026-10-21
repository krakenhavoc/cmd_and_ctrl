package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// guff_slice_a_test.go — the Guffster's Party deck request (#4184),
// slice a: Ajani Steadfast, Chandra, Legacy of Fire, Commodore Guff,
// Fog Bank, Gatewatch Beacon, Honor-Worn Shaku, Mobilized District,
// Oath of Gideon and Tezzeret the Seeker.

const (
	guffaAjaniOracle     = "90249526-1875-450a-882d-bf6f6bfd37a6"
	guffaChandraOracle   = "8dfa789d-ae1c-4f18-a55c-f49a363a5e33"
	guffaGuffOracle      = "7e5a7a1d-dcfb-4ab8-b773-f521cb4d3fa6"
	guffaFogBankOracle   = "3a2bb47d-228d-499e-a34b-9e10d99e9b2c"
	guffaBeaconOracle    = "2a31cd16-fb30-4bb0-b775-d5e8b9221dd4"
	guffaShakuOracle     = "babeeaf6-0fe4-491f-bbf5-63a0568b3d6e"
	guffaDistrictOracle  = "eb2094cf-b4be-4f52-8615-e179ef7c741d"
	guffaOathOracle      = "ea168acc-22a1-4ba5-affd-fb813b93535a"
	guffaTezzeretOracle  = "4aeafdf3-7828-4114-8c74-3bd46357caec"
	guffaSemestersOracle = "3aeba3c2-5517-4870-8802-29b0fa70ffae"
)

func TestGuffaSliceCompleteness(t *testing.T) {
	for _, oracle := range []string{guffaChandraOracle, guffaGuffOracle, guffaFogBankOracle, guffaBeaconOracle,
		guffaShakuOracle, guffaDistrictOracle, guffaOathOracle, guffaTezzeretOracle, guffaSemestersOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	// Ajani Steadfast is Full since #4192 (ajani_steadfast_emblem_test.go),
	// Semester's End since #4193 (semesters_end_test.go).
}

// guffaCastWalker casts a planeswalker with printed loyalty from the
// active player's hand and resolves it.
func guffaCastWalker(t *testing.T, g *game.Game, name string, loyalty int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: loyalty, Owner: active.ID, Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	t.Fatalf("%s did not reach the battlefield", name)
	return uuid.Nil
}

func guffaDamage(g *game.Game, id uuid.UUID) int {
	c, _ := battlefieldCard(g, id)
	return c.DamageMarked
}

// --- Fog Bank -------------------------------------------------------

func TestGuffaFogBankPreventsCombatDamageToAndByIt(t *testing.T) {
	g, me, opp := spendTable(t)
	fog := pushCatalogPermanent(g, me.ID, "Fog Bank", "Creature — Wall", guffaFogBankOracle, false)
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)

	if err := g.MarkCombatDamage(bear, fog, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := guffaDamage(g, fog); got != 0 {
		t.Errorf("Fog Bank took %d combat damage, want 0", got)
	}
	if err := g.MarkCombatDamage(fog, bear, 1); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := guffaDamage(g, bear); got != 0 {
		t.Errorf("Fog Bank dealt %d combat damage, want 0", got)
	}
	// Noncombat damage is not touched.
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(bear, fog, 1) })
	if got := guffaDamage(g, fog); got != 1 {
		t.Errorf("Fog Bank took %d noncombat damage, want 1", got)
	}
}

// --- Honor-Worn Shaku -----------------------------------------------

func TestGuffaShakuUntapsByTappingALegendaryPermanent(t *testing.T) {
	g, me, _ := spendTable(t)
	shaku := pushCatalogPermanent(g, me.ID, "Honor-Worn Shaku", "Artifact", guffaShakuOracle, false)
	plain := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	legend := pushCatalogPermanent(g, me.ID, "Legend", "Legendary Creature — Human", "", true)
	if err := g.ActivateManaAbility(me.ID, shaku, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Shaku for mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, shaku, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{plain}}); err == nil {
		t.Fatal("a nonlegendary creature paid the untap")
	}
	// A summoning-sick legend pays it: the cost is not {T}.
	b16Activate(t, g, me.ID, shaku, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{legend}})
	if !b16Tapped(t, g, legend) {
		t.Error("the legend did not tap for the cost")
	}
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, shaku) {
		t.Error("the Shaku is still tapped")
	}
}

// --- Mobilized District ---------------------------------------------

func TestGuffaDistrictCostsLessForLegendaryCreaturesAndPlaneswalkers(t *testing.T) {
	g, me, _ := spendTable(t)
	district := pushCatalogPermanent(g, me.ID, "Mobilized District", "Land", guffaDistrictOracle, false)
	pushCatalogPermanent(g, me.ID, "Legend", "Legendary Creature — Human", "", false)
	pushWalkerForTest(g, me.ID, "Walker", "", 3)
	pushCatalogPermanent(g, me.ID, "Relic", "Legendary Artifact", "", false) // neither: no discount
	pushVanillaCreature(g, me.ID, "Bear", 2, 2)                              // not legendary: no discount

	apaMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, district, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{4} less two paid with one mana")
	}
	apaMana(me, "W")
	b16Activate(t, g, me.ID, district, 0, game.ActivateAbilityParams{Strict: true})
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool %d after paying {2}, want 0", me.ManaPool.Total())
	}
	passPriorityAroundTable(t, g)
	c := effectiveCardForTest(t, g, district)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Citizen") {
		t.Errorf("after animating: creature %v, land %v, Citizen %v", c.IsCreature(), c.IsLand(), c.HasSubtype("Citizen"))
	}
	if p, tough := effectivePower(t, g, district), effectiveToughness(t, g, district); p != 3 || tough != 3 {
		t.Errorf("P/T %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, district, "vigilance") {
		t.Error("no vigilance")
	}
}

// --- Gatewatch Beacon -----------------------------------------------

func TestGuffaBeaconMovesALoyaltyCounterOntoAnArrivingWalker(t *testing.T) {
	g, me, _ := spendTable(t)
	beacon := castCatalogSpell(t, g, "Gatewatch Beacon", "Artifact", guffaBeaconOracle, nil)
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, beacon); got != 3 {
		t.Fatalf("Beacon entered with %d loyalty counters, want 3", got)
	}
	walker := guffaCastWalker(t, g, "Walker", 3)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, walker); got != 4 {
		t.Errorf("walker loyalty %d, want 4", got)
	}
	if got := loyaltyOf(g, beacon); got != 2 {
		t.Errorf("Beacon loyalty %d, want 2", got)
	}

	// Declined: nothing moves.
	other := guffaCastWalker(t, g, "Other Walker", 2)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, other); got != 2 {
		t.Errorf("declined: walker loyalty %d, want 2", got)
	}
	if got := loyaltyOf(g, beacon); got != 2 {
		t.Errorf("declined: Beacon loyalty %d, want 2", got)
	}
}

// With no loyalty counters left the Beacon does not trigger (CR 603.4).
func TestGuffaBeaconWithNoCountersDoesNotTrigger(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Gatewatch Beacon", "Artifact", guffaBeaconOracle, false)
	walkerID := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: walkerID, Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 3, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, walkerID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for i := 0; i < 16 && countNamed(g, "Walker") == 0; i++ {
		passPriorityOnce(t, g)
	}
	g.RunStateChecksForTest()
	if len(g.StackMeta) != 0 {
		t.Fatalf("%d items on the stack after the walker resolved: an empty Beacon triggered", len(g.StackMeta))
	}
	if c := latestTriggerPrompt(g, me.ID); c != nil {
		t.Fatalf("an empty Beacon asked: %+v", c)
	}
}

// A Beacon emptied while its trigger waits does nothing and asks
// nothing: the intervening "if" is checked again (CR 603.4).
func TestGuffaBeaconEmptiedInResponseAsksNothing(t *testing.T) {
	g, me, _ := spendTable(t)
	beacon := castCatalogSpell(t, g, "Gatewatch Beacon", "Artifact", guffaBeaconOracle, nil)
	passPriorityAroundTable(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	walkerID := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: walkerID, Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 3, Owner: active.ID, Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, walkerID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// Resolve the walker only; its entry puts the Beacon's trigger on the stack.
	for i := 0; i < 16 && countNamed(g, "Walker") == 0; i++ {
		passPriorityOnce(t, g)
	}
	if countNamed(g, "Walker") != 1 || len(g.StackMeta) == 0 {
		t.Fatalf("walker on battlefield %d, stack %d: want the walker in play and the trigger waiting", countNamed(g, "Walker"), len(g.StackMeta))
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(beacon, game.CounterLoyalty, -3) })
	passPriorityAroundTable(t, g)
	if c := latestTriggerPrompt(g, me.ID); c != nil {
		t.Fatalf("an emptied Beacon asked: %+v", c)
	}
	var walker uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Walker" {
			walker = c.InstanceID
		}
	}
	if got := loyaltyOf(g, walker); got != 3 {
		t.Errorf("walker loyalty %d, want 3", got)
	}
}

// --- Oath of Gideon -------------------------------------------------

func TestGuffaOathMakesTwoKorAlliesAndAddsALoyaltyCounter(t *testing.T) {
	g, me, _ := spendTable(t)
	castCatalogSpell(t, g, "Oath of Gideon", "Legendary Enchantment", guffaOathOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Kor Ally"); n != 2 {
		t.Errorf("%d Kor Ally tokens, want 2", n)
	}
	walker := guffaCastWalker(t, g, "Walker", 3)
	if got := loyaltyOf(g, walker); got != 4 {
		t.Errorf("walker entered with %d loyalty under the Oath, want 4", got)
	}
	_ = me
}

// A walker an effect puts onto the battlefield already holding loyalty
// counters skips the starting-loyalty stamp; the Oath adds to the
// entry's counters instead.
func TestGuffaOathAddsToAnEntryWithLoyaltyCounters(t *testing.T) {
	g, me, opp := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Oath of Gideon", "Legendary Enchantment", guffaOathOracle, false)
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 5, Owner: me.ID, Controller: me.ID})
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 5, Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		_ = g.PutOntoBattlefieldTogetherThenForEffect([]game.BatchEntry{{CardID: mine, From: game.ZoneGraveyard}},
			game.ZoneEntryOptions{Controller: me.ID, EntersWithCounters: map[string]int{game.CounterLoyalty: 2}}, nil)
		_ = g.PutOntoBattlefieldTogetherThenForEffect([]game.BatchEntry{{CardID: theirs, From: game.ZoneGraveyard}},
			game.ZoneEntryOptions{Controller: opp.ID, EntersWithCounters: map[string]int{game.CounterLoyalty: 2}}, nil)
	})
	for name, want := range map[string]int{"My Walker": 3, "Their Walker": 2} {
		var got int
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				got = c.Counters[game.CounterLoyalty]
			}
		}
		if got != want {
			t.Errorf("%s entered with %d loyalty, want %d", name, got, want)
		}
	}
}

// --- Ajani Steadfast ------------------------------------------------

func TestGuffaAjaniPlusOnePumpsAndGrantsKeywords(t *testing.T) {
	g, me, _ := spendTable(t)
	ajani := pushWalkerForTest(g, me.ID, "Ajani Steadfast", guffaAjaniOracle, 4)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	b16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{Targets: roonaCardRef(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("Bear %d/%d, want 3/3", p, tough)
	}
	for _, kw := range []string{"first strike", "vigilance", "lifelink"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("Bear lacks %s", kw)
		}
	}
	if got := loyaltyOf(g, ajani); got != 5 {
		t.Errorf("Ajani loyalty %d, want 5", got)
	}
}

func TestGuffaAjaniMinusTwoCountersCreaturesAndOtherWalkers(t *testing.T) {
	g, me, opp := spendTable(t)
	ajani := pushWalkerForTest(g, me.ID, "Ajani Steadfast", guffaAjaniOracle, 4)
	other := pushWalkerForTest(g, me.ID, "Other Walker", "", 3)
	theirWalker := pushWalkerForTest(g, opp.ID, "Their Walker", "", 3)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirBear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	b16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, ajani); got != 2 {
		t.Errorf("Ajani loyalty %d after −2, want 2 (no counter on himself)", got)
	}
	if got := loyaltyOf(g, other); got != 4 {
		t.Errorf("other walker loyalty %d, want 4", got)
	}
	if got := loyaltyOf(g, theirWalker); got != 3 {
		t.Errorf("opponent's walker loyalty %d, want 3", got)
	}
	if c, _ := battlefieldCard(g, bear); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("Bear has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
	}
	if c, _ := battlefieldCard(g, theirBear); c.Counters[game.CounterPlusOne] != 0 {
		t.Error("an opponent's creature got a counter")
	}
}

// --- Tezzeret the Seeker --------------------------------------------

func TestGuffaTezzeretUntapsTwoArtifacts(t *testing.T) {
	g, me, _ := spendTable(t)
	tez := pushWalkerForTest(g, me.ID, "Tezzeret the Seeker", guffaTezzeretOracle, 4)
	a := pushCatalogPermanent(g, me.ID, "Rock A", "Artifact", "", false)
	b := pushCatalogPermanent(g, me.ID, "Rock B", "Artifact", "", false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{a, b, bear} {
			_ = g.TapTargetForEffect(id)
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, tez, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("+1 targeted a creature that is not an artifact")
	}
	b16Activate(t, g, me.ID, tez, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}})
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, a) || b16Tapped(t, g, b) {
		t.Error("an artifact is still tapped")
	}
}

func TestGuffaTezzeretMinusXFindsAnArtifactOfManaValueXOrLess(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	tez := pushWalkerForTest(g, me.ID, "Tezzeret the Seeker", guffaTezzeretOracle, 4)
	small := pushLibraryCardForTest(me, game.Card{Name: "Small Rock", TypeLine: "Artifact", ManaCost: "{2}"})
	big := pushLibraryCardForTest(me, game.Card{Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{3}"})
	bear := pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"})
	advanceToMainOf(t, g, seat)
	b16Activate(t, g, me.ID, tez, 1, game.ActivateAbilityParams{XValue: 2})
	if got := loyaltyOf(g, tez); got != 2 {
		t.Errorf("loyalty after −2 = %d, want 2", got)
	}
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, small); !ok {
		t.Error("the mana value 2 artifact was not put onto the battlefield")
	}
	for _, id := range []uuid.UUID{big, bear} {
		if _, wrong := battlefieldCard(g, id); wrong {
			t.Error("a card over the mana value, or not an artifact, came into play")
		}
	}
}

func TestGuffaTezzeretMinusFiveAnimatesYourArtifacts(t *testing.T) {
	g, me, opp := spendTable(t)
	tez := pushWalkerForTest(g, me.ID, "Tezzeret the Seeker", guffaTezzeretOracle, 5)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	b16Activate(t, g, me.ID, tez, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	c := effectiveCardForTest(t, g, rock)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Error("Rock is not an artifact creature")
	}
	if p, tough := effectivePower(t, g, rock), effectiveToughness(t, g, rock); p != 5 || tough != 5 {
		t.Errorf("Rock %d/%d, want 5/5", p, tough)
	}
	if effectiveCardForTest(t, g, theirs).IsCreature() {
		t.Error("an opponent's artifact was animated")
	}
	// One that arrives later is not (CR 611.2c).
	late := pushCatalogPermanent(g, me.ID, "Late Rock", "Artifact", "", false)
	if effectiveCardForTest(t, g, late).IsCreature() {
		t.Error("an artifact that arrived afterwards was animated")
	}
}

// --- Commodore Guff -------------------------------------------------

func TestGuffaGuffPlusOneMakesAWizardWhoseManaCastsOnlyPlaneswalkers(t *testing.T) {
	g, me, _ := spendTable(t)
	guff := pushWalkerForTest(g, me.ID, "Commodore Guff", guffaGuffOracle, 5)
	b16Activate(t, g, me.ID, guff, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	var wizard game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Wizard" {
			wizard = c
		}
	}
	if wizard.InstanceID == uuid.Nil {
		t.Fatal("no Wizard token")
	}
	if !wizard.IsToken() || wizard.Power != 1 || wizard.Toughness != 1 || len(wizard.Colors) != 1 || wizard.Colors[0] != "R" {
		t.Errorf("Wizard is %+v, want a 1/1 red token", wizard)
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == wizard.InstanceID {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	if err := g.ActivateManaAbility(me.ID, wizard.InstanceID, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Wizard for mana: %v", err)
	}
	bear := handSpell(me, "Red Creature", "Creature — Bear", "{R}")
	refusedForMana(t, g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}), "the Wizard's mana for a creature spell")
	walker := handSpell(me, "Red Walker", "Legendary Planeswalker — Test", "{R}")
	if err := g.CastSpell(me.ID, walker, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Wizard's {R} for a planeswalker spell: %v", err)
	}
}

func TestGuffaGuffMinusThreeDrawsAndBurnsByPlaneswalkerCount(t *testing.T) {
	g, me, _ := spendTable(t)
	guff := pushWalkerForTest(g, me.ID, "Commodore Guff", guffaGuffOracle, 5)
	pushWalkerForTest(g, me.ID, "Other Walker", "", 3)
	for i := 0; i < 3; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Card", TypeLine: "Instant"})
	}
	hand := me.Hand.Size()
	life := lifeOfOpponents(g)
	b16Activate(t, g, me.ID, guff, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	for i, l := range lifeOfOpponents(g) {
		if life[i]-l != 2 {
			t.Errorf("opponent %d lost %d life, want 2", i, life[i]-l)
		}
	}
}

func TestGuffaGuffEndStepPutsALoyaltyCounterOnAnotherWalker(t *testing.T) {
	g, me, _ := spendTable(t)
	guff := pushWalkerForTest(g, me.ID, "Commodore Guff", guffaGuffOracle, 5)
	other := pushWalkerForTest(g, me.ID, "Other Walker", "", 3)
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, guff) {
			t.Error("Guff is offered as his own target")
		}
		pickCard(t, g, me.ID, other)
	}
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, other); got != 4 {
		t.Errorf("other walker loyalty %d, want 4", got)
	}
	if got := loyaltyOf(g, guff); got != 5 {
		t.Errorf("Guff loyalty %d, want 5", got)
	}
}

// --- Chandra, Legacy of Fire ----------------------------------------

func TestGuffaChandraPlusOneAddsRedPerWalker(t *testing.T) {
	g, me, _ := spendTable(t)
	chandra := pushWalkerForTest(g, me.ID, "Chandra, Legacy of Fire", guffaChandraOracle, 3)
	pushWalkerForTest(g, me.ID, "Other Walker", "", 3)
	pushWalkerForTest(g, me.ID, "Third Walker", "", 3)
	b16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("pool %d, want {R}{R}{R}", got)
	}
}

func TestGuffaChandraEndStepBurnsEachOpponent(t *testing.T) {
	g, me, _ := spendTable(t)
	pushWalkerForTest(g, me.ID, "Chandra, Legacy of Fire", guffaChandraOracle, 3)
	pushWalkerForTest(g, me.ID, "Other Walker", "", 3)
	life := lifeOfOpponents(g)
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	for i, l := range lifeOfOpponents(g) {
		if life[i]-l != 2 {
			t.Errorf("opponent %d lost %d life, want 2", i, life[i]-l)
		}
	}
}

func TestGuffaChandraZeroRemovesCountersAndExilesThatMany(t *testing.T) {
	g, me, opp := spendTable(t)
	chandra := pushWalkerForTest(g, me.ID, "Chandra, Legacy of Fire", guffaChandraOracle, 3)
	other := pushWalkerForTest(g, me.ID, "Other Walker", "", 3)
	theirs := pushWalkerForTest(g, opp.ID, "Their Walker", "", 3)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Card", TypeLine: "Instant"})
	}
	library := me.Library.Size()
	b16Activate(t, g, me.ID, chandra, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no choose-cards prompt: %+v", g.PendingChoices)
	}
	for _, id := range []uuid.UUID{chandra, other} {
		if !hasID(c.ChooseCards, id) {
			t.Errorf("%s is not offered", id)
		}
	}
	for _, id := range []uuid.UUID{theirs, bear} {
		if hasID(c.ChooseCards, id) {
			t.Errorf("%s is offered (an opponent's, or no loyalty counter)", id)
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{chandra, other}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := loyaltyOf(g, chandra); got != 2 {
		t.Errorf("Chandra loyalty %d, want 2", got)
	}
	if got := loyaltyOf(g, other); got != 2 {
		t.Errorf("other walker loyalty %d, want 2", got)
	}
	if got := library - me.Library.Size(); got != 2 {
		t.Errorf("exiled %d from the library, want 2", got)
	}
}
