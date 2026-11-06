package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b8_cards_test.go — Avatar: The Last Airbender batch 8 (#4459).

const (
	avB8CourtOfficialOracle = "ec84201d-757f-4a7a-b3a2-85ccd5ee81ae"
	avB8BlackSunOracle      = "dc857129-533c-41d1-8e8b-1c0443030d69"
	avB8EpicDownfallOracle  = "ed5d6aba-0b8d-48b3-a84c-618f45fb67b1"
	avB8DrillOracle         = "3d4b9d71-4275-4c18-a4dc-9aad234eb92c"
	avB8EngineerOracle      = "9a983666-376b-48a8-9f82-8d041d32ae93"
	avB8TrebuchetOracle     = "1f018a8c-6ceb-4d03-94a7-8713467c9cf5"
	avB8HuntersOracle       = "1941cf7a-ad53-4bce-beea-104951619a30"
	avB8VisionsOracle       = "16af0a22-833e-47c6-ac27-002702194595"
	avB8HeartlessActOracle  = "3641d572-8335-4804-88ad-edf4dc67a8e4"
	avB8HogMonkeyOracle     = "e11591ee-7074-488e-8d2f-f8c89e385a55"

	avB8DoomedTravelerOracle = "a30907c0-fbde-4fd3-a8c7-f304305fcea7"
)

func TestAvatarB8CardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		avB8CourtOfficialOracle: "Corrupt Court Official",
		avB8BlackSunOracle:      "Day of Black Sun",
		avB8EpicDownfallOracle:  "Epic Downfall",
		avB8DrillOracle:         "The Fire Nation Drill",
		avB8EngineerOracle:      "Fire Nation Engineer",
		avB8TrebuchetOracle:     "Fire Navy Trebuchet",
		avB8HuntersOracle:       "Foggy Swamp Hunters",
		avB8VisionsOracle:       "Foggy Swamp Visions",
		avB8HeartlessActOracle:  "Heartless Act",
		avB8HogMonkeyOracle:     "Hog-Monkey",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s is not Full", name)
		}
	}
}

// avB8Creature seeds a creature with a mana cost and printed keywords.
func avB8Creature(g *game.Game, owner uuid.UUID, name, oracle, manaCost string, power, toughness int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", OracleID: oracle, ManaCost: manaCost,
		Power: power, Toughness: toughness, Keywords: keywords, Owner: owner, Controller: owner,
	})
}

// avB8Cast puts a card with a mana cost in the active player's hand and
// casts it from a main phase with the given parameters.
func avB8Cast(t *testing.T, g *game.Game, name, typeLine, manaCost, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, name, typeLine, manaCost, oracle, params)
}

func avB8Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func avB8Named(g *game.Game, name string, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}

// --- Corrupt Court Official -----------------------------------------------------

func TestAvatarB8CorruptCourtOfficialMakesTargetOpponentDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Hand.Size()
	if before < 1 {
		t.Fatalf("test setup: opponent hand %d", before)
	}
	castCatalogSpell(t, g, "Corrupt Court Official", "Creature — Human Advisor", avB8CourtOfficialOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("the trigger targeted its own controller")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	if got := before - opp.Hand.Size(); got != 1 {
		t.Errorf("opponent discarded %d, want 1", got)
	}
}

// --- Day of Black Sun ------------------------------------------------------------

func TestAvatarB8DayOfBlackSunStripsThenDestroysUpToX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stubborn := avB8Creature(g, opp.ID, "Stubborn Wall", "", "{1}{W}", 0, 4, "indestructible")
	traveler := avB8Creature(g, me.ID, "Doomed Traveler", avB8DoomedTravelerOracle, "{W}", 1, 1)
	big := avB8Creature(g, opp.ID, "Big Bear", "", "{2}{G}", 3, 3, "indestructible")

	if _, err := avB8Cast(t, g, "Day of Black Sun", "Sorcery", "{X}{B}{B}", avB8BlackSunOracle,
		game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast Day of Black Sun X=2: %v", err)
	}
	passPriorityAroundTable(t, g)

	if onBattlefield(g, stubborn) {
		t.Error("an indestructible creature with mana value 2 survived: it should have lost indestructible first")
	}
	if onBattlefield(g, traveler) {
		t.Error("a mana value 1 creature survived")
	}
	if len(avB8Named(g, "Spirit", me.ID)) != 0 {
		t.Error("Doomed Traveler's dies trigger fired although it had lost all abilities")
	}
	if !onBattlefield(g, big) {
		t.Fatal("a mana value 3 creature was destroyed by X=2")
	}
	if !hasEffectiveKeyword(t, g, big, "indestructible") {
		t.Error("a creature outside the set lost its abilities")
	}
}

// --- Epic Downfall ---------------------------------------------------------------

func TestAvatarB8EpicDownfallExilesOnlyManaValueThreeOrMore(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := avB8Creature(g, opp.ID, "Small Bear", "", "{1}{G}", 2, 2)
	big := avB8Creature(g, opp.ID, "Big Bear", "", "{2}{G}", 3, 3)

	if _, err := avB8Cast(t, g, "Epic Downfall", "Sorcery", "{1}{B}", avB8EpicDownfallOracle,
		game.CastSpellParams{Strict: true, Targets: avB8Card(small)}); err == nil {
		t.Fatal("Epic Downfall targeted a mana value 2 creature")
	}
	if _, err := avB8Cast(t, g, "Epic Downfall", "Sorcery", "{1}{B}", avB8EpicDownfallOracle,
		game.CastSpellParams{Targets: avB8Card(big)}); err != nil {
		t.Fatalf("Epic Downfall on a mana value 3 creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, big) {
		t.Fatal("the target was not removed")
	}
	if !g.Exile.Contains(big) {
		t.Error("the target was not exiled")
	}
}

// --- The Fire Nation Drill -------------------------------------------------------

func TestAvatarB8FireNationDrillTapsToDestroyPowerFourOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := avB8Creature(g, opp.ID, "Victim", "", "{3}", 4, 4)
	brute := avB8Creature(g, opp.ID, "Brute", "", "{4}", 5, 5)

	drill := castCatalogSpell(t, g, "The Fire Nation Drill", "Legendary Artifact — Vehicle", avB8DrillOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !tapCostTapped(g, drill) {
		t.Fatal("the Drill was not tapped")
	}
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: brute}); err == nil {
		t.Fatal("the reflexive trigger targeted a power 5 creature")
	}
	pickTriggerTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, victim) {
		t.Error("the power 4 creature was not destroyed")
	}
	if !onBattlefield(g, brute) {
		t.Error("the power 5 creature was destroyed")
	}
	if !hasEffectiveKeyword(t, g, drill, "trample") {
		t.Error("the Drill has no trample")
	}
}

func TestAvatarB8FireNationDrillDeclinedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := avB8Creature(g, opp.ID, "Victim", "", "{3}", 2, 2)
	drill := castCatalogSpell(t, g, "The Fire Nation Drill", "Legendary Artifact — Vehicle", avB8DrillOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if tapCostTapped(g, drill) {
		t.Error("a declined Drill was tapped")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("declining still created the reflexive trigger")
	}
	if !onBattlefield(g, victim) {
		t.Error("declining destroyed a creature")
	}
}

func TestAvatarB8FireNationDrillStripsOpponentsHexproofAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drill := pushCatalogPermanent(g, me.ID, "The Fire Nation Drill", "Legendary Artifact — Vehicle", avB8DrillOracle, false)
	theirs := avB8Creature(g, opp.ID, "Warded God", "", "{3}", 3, 3, "hexproof", "indestructible")
	mine := avB8Creature(g, me.ID, "My God", "", "{3}", 3, 3, "indestructible")
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ActivateCatalogAbility(me.ID, drill, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate {1}: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"hexproof", "indestructible"} {
		if hasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("the opponent's creature kept %s", kw)
		}
	}
	if !hasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("the Drill's controller's own creature lost indestructible")
	}
	if err := g.ActivateCatalogAbility(me.ID, drill, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the {1} ability was activated with an empty pool")
	}
}

// --- Fire Nation Engineer ----------------------------------------------------------

func TestAvatarB8FireNationEngineerRaidCountersAVehicleAfterAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engineer := pushCatalogPermanent(g, me.ID, "Fire Nation Engineer", "Creature — Human Artificer", avB8EngineerOracle, false)
	vehicle := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cart", TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: engineer}); err == nil {
		t.Fatal("the Engineer targeted itself")
	}
	pickTriggerTarget(t, g, me.ID, vehicle)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, vehicle, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the Vehicle: %d, want 1", got)
	}
}

func TestAvatarB8FireNationEngineerWithoutAnAttackDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Fire Nation Engineer", "Creature — Human Artificer", avB8EngineerOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || triggersOnStackFrom(g, bear) != 0 {
		t.Fatal("raid triggered without an attack")
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("a counter was placed without an attack: %d", got)
	}
}

// --- Fire Navy Trebuchet -----------------------------------------------------------

func TestAvatarB8FireNavyTrebuchetHurlsABoulderThatIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	treb := pushCatalogPermanent(g, me.ID, "Fire Navy Trebuchet", "Artifact Creature — Wall", avB8TrebuchetOracle, false)
	for _, kw := range []string{"defender", "reach"} {
		if !hasEffectiveKeyword(t, g, treb, kw) {
			t.Errorf("Fire Navy Trebuchet has no %s", kw)
		}
	}
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	for c := latestChoiceOfKind(g, game.PendingChoicePayAmount); c != nil; c = latestChoiceOfKind(g, game.PendingChoicePayAmount) {
		if err := g.ResolvePayAmount(c.ID, me.ID, c.PayAmount.Max); err != nil {
			t.Fatalf("place the Boulder: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	boulders := avB8Named(g, "Ballistic Boulder", me.ID)
	if len(boulders) != 1 {
		t.Fatalf("a two-creature attack made %d Boulders, want 1", len(boulders))
	}
	bo := boulders[0]
	if !bo.Tapped || bo.AttackingTarget != opp.ID {
		t.Errorf("the Boulder is tapped=%v attacking %v, want tapped and attacking the defender", bo.Tapped, bo.AttackingTarget)
	}
	if !IsToken(bo) || !hasEffectiveKeyword(t, g, bo.InstanceID, "flying") {
		t.Error("the Boulder is not a flying token")
	}
	if bo.Power != 2 || bo.Toughness != 1 {
		t.Errorf("the Boulder is %d/%d, want 2/1", bo.Power, bo.Toughness)
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, bo.InstanceID) {
		t.Error("the Boulder survived the end step")
	}
}

// --- Foggy Swamp Hunters -----------------------------------------------------------

func TestAvatarB8FoggySwampHuntersGainLifelinkAndMenaceOnTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hunters := pushCatalogPermanent(g, me.ID, "Foggy Swamp Hunters", "Creature — Human Ranger Ally", avB8HuntersOracle, false)
	for g.TurnTallyFor(me.ID).CardsDrawn < 1 {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatal(err)
		}
	}
	if g.TurnTallyFor(me.ID).CardsDrawn != 1 {
		t.Fatalf("test setup: %d cards drawn", g.TurnTallyFor(me.ID).CardsDrawn)
	}
	for _, kw := range []string{"lifelink", "menace"} {
		if hasEffectiveKeyword(t, g, hunters, kw) {
			t.Errorf("Hunters have %s after one draw", kw)
		}
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	for _, kw := range []string{"lifelink", "menace"} {
		if !hasEffectiveKeyword(t, g, hunters, kw) {
			t.Errorf("Hunters lack %s after the second draw", kw)
		}
	}
}

// --- Foggy Swamp Visions -----------------------------------------------------------

func TestAvatarB8FoggySwampVisionsCopiesExiledCreaturesUntilYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	traveler := seedGraveyardCreature(opp, "Doomed Traveler", "{W}")
	for i := range opp.Graveyard.Cards {
		if opp.Graveyard.Cards[i].InstanceID == traveler {
			opp.Graveyard.Cards[i].OracleID = avB8DoomedTravelerOracle
			opp.Graveyard.Cards[i].TypeLine = "Creature — Human Soldier"
		}
	}
	giant := seedGraveyardCreature(me, "Zombie Giant", "{4}{B}")
	helper := pushVanillaCreature(g, me.ID, "Helper", 1, 1)
	targets := append(avB8Card(traveler), avB8Card(giant)...)

	// {1}{B}{B} plus waterbend {2}: one tap pays {1} of it, the pool the rest.
	if _, err := avB8Cast(t, g, "Foggy Swamp Visions", "Sorcery", "{1}{B}{B}", avB8VisionsOracle, game.CastSpellParams{
		Strict: true, XValue: 1, Targets: targets,
	}); err == nil {
		t.Fatal("X=1 took two targets")
	}
	for _, c := range []string{"B", "B", "C", "C"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	if _, err := avB8Cast(t, g, "Foggy Swamp Visions", "Sorcery", "{1}{B}{B}", avB8VisionsOracle, game.CastSpellParams{
		Strict: true, XValue: 2, TapIDs: []uuid.UUID{helper}, Targets: targets,
	}); err != nil {
		t.Fatalf("cast with waterbend {2} (one tap, one mana): %v", err)
	}
	if !tapCostTapped(g, helper) {
		t.Error("the waterbend tap did not tap the helper")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{traveler, giant} {
		if !g.Exile.Contains(id) {
			t.Errorf("a target card was not exiled")
		}
	}
	travelers := avB8Named(g, "Doomed Traveler", me.ID)
	giants := avB8Named(g, "Zombie Giant", me.ID)
	if len(travelers) != 1 || len(giants) != 1 {
		t.Fatalf("copies: %d Travelers and %d Giants under the caster's control, want one of each", len(travelers), len(giants))
	}
	if !IsToken(travelers[0]) || !IsToken(giants[0]) {
		t.Fatal("the copies are not tokens")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, travelers[0].InstanceID) || onBattlefield(g, giants[0].InstanceID) {
		t.Fatal("the copies survived the caster's end step")
	}
	// Each copy keeps the copied card's abilities: the Traveler token's
	// dies trigger makes its Spirit as it is sacrificed.
	if len(avB8Named(g, "Spirit", me.ID)) != 1 {
		t.Error("the Doomed Traveler copy's dies trigger did not make a Spirit")
	}
}

// --- Heartless Act -----------------------------------------------------------------

func TestAvatarB8HeartlessActDestroysOnlyACreatureWithNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	clean := avB8Creature(g, opp.ID, "Clean Bear", "", "{1}{G}", 2, 2)
	marked := avB8Creature(g, opp.ID, "Marked Bear", "", "{1}{G}", 2, 2)
	if err := g.AddCounter(marked, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := avB8Cast(t, g, "Heartless Act", "Instant", "{1}{B}", avB8HeartlessActOracle, game.CastSpellParams{
		Strict: true, Modes: []int{0}, Targets: avB8Card(marked),
	}); err == nil {
		t.Fatal("the destroy mode targeted a creature with a counter")
	}
	if _, err := avB8Cast(t, g, "Heartless Act", "Instant", "{1}{B}", avB8HeartlessActOracle, game.CastSpellParams{
		Modes: []int{0}, Targets: avB8Card(clean),
	}); err != nil {
		t.Fatalf("destroy mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, clean) {
		t.Error("the creature with no counters survived")
	}
}

func TestAvatarB8HeartlessActRemovesUpToThreeCountersOfYourChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := avB8Creature(g, opp.ID, "Pumped Bear", "", "{1}{G}", 2, 2)
	if err := g.AddCounter(big, game.CounterPlusOne, 3); err != nil {
		t.Fatal(err)
	}
	if err := g.AddCounter(big, "shield", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := avB8Cast(t, g, "Heartless Act", "Instant", "{1}{B}", avB8HeartlessActOracle, game.CastSpellParams{
		Modes: []int{1}, Targets: avB8Card(big),
	}); err != nil {
		t.Fatalf("remove mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Kinds are asked in sorted order: +1/+1 first, then shield.
	first := numberPrompt(t, g, me.ID)
	if first.PayAmount.Max != 3 {
		t.Fatalf("first question allows %d, want 3", first.PayAmount.Max)
	}
	if err := g.ResolvePayAmount(first.ID, me.ID, 2); err != nil {
		t.Fatal(err)
	}
	second := numberPrompt(t, g, me.ID)
	if second.PayAmount.Max != 1 {
		t.Fatalf("second question allows %d, want the 1 left of three", second.PayAmount.Max)
	}
	if err := g.ResolvePayAmount(second.ID, me.ID, 1); err != nil {
		t.Fatal(err)
	}
	if got := countersOn(g, big, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters left: %d, want 1", got)
	}
	if got := countersOn(g, big, "shield"); got != 0 {
		t.Errorf("shield counters left: %d, want 0", got)
	}
	if !onBattlefield(g, big) {
		t.Error("the remove mode destroyed the creature")
	}
}

// --- Hog-Monkey -------------------------------------------------------------------

func TestAvatarB8HogMonkeyGivesMenaceToACounteredCreatureAndExhaustsOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hog := pushCatalogPermanent(g, me.ID, "Hog-Monkey", "Creature — Boar Monkey", avB8HogMonkeyOracle, false)
	plain := pushVanillaCreature(g, me.ID, "Plain Bear", 2, 2)
	advanceToMain(t, g)

	for i := 0; i < 5; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	if err := g.ActivateCatalogAbility(me.ID, hog, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Exhaust — {5}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, hog, game.CounterPlusOne); got != 2 {
		t.Fatalf("+1/+1 counters on Hog-Monkey: %d, want 2", got)
	}
	for i := 0; i < 5; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	if err := g.ActivateCatalogAbility(me.ID, hog, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: plain}); err == nil {
		t.Fatal("the trigger targeted a creature with no +1/+1 counter")
	}
	pickTriggerTarget(t, g, me.ID, hog)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, hog, "menace") {
		t.Error("the targeted creature did not gain menace")
	}
}

func TestAvatarB8HogMonkeyWithNoCounteredCreatureAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hog := pushCatalogPermanent(g, me.ID, "Hog-Monkey", "Creature — Boar Monkey", avB8HogMonkeyOracle, false)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the trigger asked for a target with no legal one")
	}
	if hasEffectiveKeyword(t, g, hog, "menace") {
		t.Error("Hog-Monkey gained menace with no counter")
	}
}
