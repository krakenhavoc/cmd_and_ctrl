package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The cycle rows
// (Vernal Fen, the two Diamonds, the two gain lands) are loops over
// tables, so a transposed row is invisible until someone plays that
// exact card.
func TestBatch06CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B06IntoTheFloodMawOracle:     "Into the Flood Maw",
		effectstest.B06GrimTutorOracle:           "Grim Tutor",
		effectstest.B06CastleArdenvaleOracle:     "Castle Ardenvale",
		effectstest.B06WhipOfErebosOracle:        "Whip of Erebos",
		effectstest.B06FireDiamondOracle:         "Fire Diamond",
		effectstest.B06RiveteersOverlookOracle:   "Riveteers Overlook",
		effectstest.B06RisingOfTheDayOracle:      "Rising of the Day",
		effectstest.B06VernalFenOracle:           "Vernal Fen",
		effectstest.B06WarstormSurgeOracle:       "Warstorm Surge",
		effectstest.B06ForbiddenOrchardOracle:    "Forbidden Orchard",
		effectstest.B06OranRiefOracle:            "Oran-Rief, the Vastwood",
		effectstest.B06ElvesOfDeepShadowOracle:   "Elves of Deep Shadow",
		effectstest.B06SpelunkingOracle:          "Spelunking",
		effectstest.B06ScouredBarrensOracle:      "Scoured Barrens",
		effectstest.B06StartingTownOracle:        "Starting Town",
		effectstest.B06IdyllicTutorOracle:        "Idyllic Tutor",
		effectstest.B06BloodchiefAscensionOracle: "Bloodchief Ascension",
		effectstest.B06ReturnToNatureOracle:      "Return to Nature",
		effectstest.B06DryadArborOracle:          "Dryad Arbor",
		effectstest.B06MossbornHydraOracle:       "Mossborn Hydra",
		effectstest.B06EladamrisCallOracle:       "Eladamri's Call",
		effectstest.B06OrnithopterOracle:         "Ornithopter",
		effectstest.B06SpringbloomDruidOracle:    "Springbloom Druid",
		effectstest.B06JungleHollowOracle:        "Jungle Hollow",
		effectstest.B06TalrandOracle:             "Talrand, Sky Summoner",
		effectstest.B06CastleLocthwainOracle:     "Castle Locthwain",
		effectstest.B06MeteorGolemOracle:         "Meteor Golem",
		effectstest.B06SkyDiamondOracle:          "Sky Diamond",
		effectstest.B06GildedGooseOracle:         "Gilded Goose",
	}
	if len(want) != 29 {
		t.Fatalf("the batch registers 29 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness == effects.CompletenessUnreviewed && name != "Vernal Fen" {
			t.Errorf("%s ships without a Completeness declaration", name)
		}
	}
}

// --- lands ---------------------------------------------------------

func TestB06CastleArdenvaleEntersTappedWithoutAPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Castle Ardenvale", effectstest.B06CastleArdenvaleOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Castle Ardenvale")
}

func TestB06CastleArdenvaleEntersUntappedWithAPlainsTypedLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// A Godless Shrine is a Plains; a Wastes is not.
	effectstest.SeedLandOnBattlefield(g, me.ID, "Godless Shrine", "Land — Plains Swamp")
	id := effectstest.PlayLandFromHand(t, g, "Castle Ardenvale", effectstest.B06CastleArdenvaleOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, id, "Castle Ardenvale")
}

func TestB06CastleArdenvaleMakesAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	castle := effectstest.SeedPermanentWithOracle(g, me.ID, "Castle Ardenvale", "Land", effectstest.B06CastleArdenvaleOracle)
	effectstest.B06AddMana(me, "W", "W", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, castle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Human") != 1 {
		t.Error("the Castle should make one 1/1 Human")
	}
	if card, _ := effectstest.BattlefieldCard(g, castle); !card.Tapped {
		t.Error("the ability has a tap cost")
	}
}

func TestB06CastleLocthwainDrawsThenChargesTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	castle := effectstest.SeedPermanentWithOracle(g, me.ID, "Castle Locthwain", "Land", effectstest.B06CastleLocthwainOracle)
	effectstest.B06AddMana(me, "B", "B", "C")
	hand, life := me.Hand.Size(), me.Life
	if err := g.ActivateCatalogAbility(me.ID, castle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
	// The drawn card is counted: "draw, THEN lose life equal to the hand".
	if want := life - (hand + 1); me.Life != want {
		t.Errorf("life %d → %d, want %d (the hand after the draw)", life, me.Life, want)
	}
}

func TestB06DiamondsEnterTappedAndTapForTheirColour(t *testing.T) {
	for _, tc := range []struct{ name, oracle, color string }{
		{"Fire Diamond", effectstest.B06FireDiamondOracle, "R"},
		{"Sky Diamond", effectstest.B06SkyDiamondOracle, "U"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		rock := effectstest.CastCatalogSpell(t, g, tc.name, "Artifact", tc.oracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.Top100AssertEnteredTapped(t, g, rock, tc.name)

		fresh := effectstest.SeedPermanentWithOracle(g, me.ID, tc.name, "Artifact", tc.oracle)
		if err := g.ActivateManaAbility(me.ID, fresh, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: ActivateManaAbility: %v", tc.name, err)
		}
		if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != tc.color {
			t.Errorf("%s: pool %v, want [%s]", tc.name, got, tc.color)
		}
	}
}

func TestB06DryadArborTapsForGreenOnceItIsNotSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad", effectstest.B06DryadArborOracle, true)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a summoning-sick Dryad Arbor tapped for mana")
	}
	ready := effectstest.PushCatalogPermanent(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad", effectstest.B06DryadArborOracle, false)
	if err := g.ActivateManaAbility(me.ID, ready, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
}

func TestB06ElvesOfDeepShadowTapForBlackAndHurt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := effectstest.PushCatalogPermanent(g, me.ID, "Elves of Deep Shadow", "Creature — Elf Druid", effectstest.B06ElvesOfDeepShadowOracle, false)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
}

func TestB06ForbiddenOrchardGivesTheChosenOpponentASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[2]
	orchard := effectstest.SeedPermanentWithOracle(g, me.ID, "Forbidden Orchard", "Land", effectstest.B06ForbiddenOrchardOracle)
	if err := g.ActivateManaAbility(me.ID, orchard, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour means all five, got %+v", pick)
	}
	// #730: the colour pick gates the table, so it is answered before
	// the Orchard's own trigger is passed around. Which colour the
	// mana is does not matter here.
	effectstest.RiderAnswerManaPicks(t, g, me.ID, "G")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, opp.ID, "Spirit") != 1 {
		t.Error("the chosen opponent should get a 1/1 Spirit")
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Spirit") != 0 {
		t.Error("the Orchard's controller never gets the Spirit")
	}
}

func TestB06GainLandsEnterTappedGainOneAndTapForBoth(t *testing.T) {
	for _, tc := range []struct{ name, oracle, a, b string }{
		{"Scoured Barrens", effectstest.B06ScouredBarrensOracle, "W", "B"},
		{"Jungle Hollow", effectstest.B06JungleHollowOracle, "B", "G"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		before := me.Life
		id := effectstest.PlayLandFromHand(t, g, tc.name, tc.oracle)
		effectstest.Top100AssertEnteredTapped(t, g, id, tc.name)
		if effectstest.TriggerOnStack(g, id) == nil {
			t.Errorf("%s: the life is a trigger with a response window", tc.name)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != before+1 {
			t.Errorf("%s: life %d → %d, want +1", tc.name, before, me.Life)
		}
		spec, _ := effects.Lookup(tc.oracle)
		if want := "{" + tc.a + "|" + tc.b + "}"; len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, want)
		}
	}
}

func TestB06VernalFenIsABattleLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	tapped := effectstest.PlayLandFromHand(t, g, "Vernal Fen", effectstest.B06VernalFenOracle)
	effectstest.Top100AssertEnteredTapped(t, g, tapped, "Vernal Fen (one basic)")

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g2, me2.ID, "Swamp", "Basic Land — Swamp")
	effectstest.SeedLandOnBattlefield(g2, me2.ID, "Forest", "Basic Land — Forest")
	untapped := effectstest.PlayLandFromHand(t, g2, "Vernal Fen", effectstest.B06VernalFenOracle)
	effectstest.Top100AssertEnteredUntapped(t, g2, untapped, "Vernal Fen (two basics)")

	spec, _ := effects.Lookup(effectstest.B06VernalFenOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{B|G}" {
		t.Errorf("mana abilities %+v, want one producing {B|G}", spec.ManaAbilities)
	}
}

func TestB06OranRiefGrowsOnlyGreenCreaturesThatEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Oran-Rief, the Vastwood", "Land", effectstest.B06OranRiefOracle)
	old := effectstest.SeedCreature(g, "Old Bear", me.ID) // no ETB event: was here before the turn

	fresh := effectstest.B06GreenCreature(t, g, me.ID, "Fresh Elf")
	theirs := effectstest.B06GreenCreature(t, g, opp.ID, "Their Elf")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")

	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.CounterCount(g, fresh, "+1/+1") != 1 {
		t.Error("a green creature that entered this turn should get a counter")
	}
	if effectstest.CounterCount(g, theirs, "+1/+1") != 1 {
		t.Error("'each green creature' is every player's")
	}
	if effectstest.CounterCount(g, old, "+1/+1") != 0 {
		t.Error("a creature from an earlier turn must not grow")
	}
	if effectstest.CounterCount(g, goblin, "+1/+1") != 0 {
		t.Error("a red creature must not grow")
	}
}

func TestB06OranRiefForgetsLastTurnsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Oran-Rief, the Vastwood", "Land", effectstest.B06OranRiefOracle)
	elf := effectstest.B06GreenCreature(t, g, me.ID, "Elf")
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, elf, "+1/+1") != 0 {
		t.Error("'entered this turn' does not reach back to an earlier turn")
	}
}

func TestB06RiveteersOverlookSacrificesFetchesTappedAndGainsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Mountain", "Basic Land — Mountain"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
	)
	before := me.Life
	overlook := effectstest.PlayLandFromHand(t, g, "Riveteers Overlook", effectstest.B06RiveteersOverlookOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(overlook) || !me.Graveyard.Contains(overlook) {
		t.Fatal("the Overlook should have sacrificed itself")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the search should ask which basic to take")
	}
	if effectstest.SearchOptionNamed(g, c, "Plains") != uuid.Nil {
		t.Error("a Plains is not a Swamp, Mountain or Forest")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Mountain")
	mountain := effectstest.FindBattlefieldByName(g, "Mountain")
	if card, ok := effectstest.BattlefieldCard(g, mountain); !ok || !card.Tapped {
		t.Error("the fetched basic enters tapped")
	}
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
}

// TestB06RiveteersOverlookFetchesThroughAReflexiveTrigger is the
// #636 half: the sacrifice and the search are TWO stack items with a
// response window between them (CR 603.12), not one.
func TestB06RiveteersOverlookFetchesThroughAReflexiveTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
	)
	overlook := effectstest.PlayLandFromHand(t, g, "Riveteers Overlook", effectstest.B06RiveteersOverlookOracle)
	// One trip around the table resolves the entry trigger.
	for i := 0; i < len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !me.Graveyard.Contains(overlook) {
		t.Fatal("the entry trigger should have sacrificed the Overlook")
	}
	// The reflexive trigger is on the stack, from a source that is
	// already in a graveyard, and nothing has been searched yet.
	reflexive := effectstest.TriggerOnStack(g, overlook)
	if reflexive == nil {
		t.Fatal(`"when you do" should be a second trigger on the stack`)
	}
	if reflexive.Controller != me.ID {
		t.Errorf("reflexive trigger controlled by %v, want the sacrificing player", reflexive.Controller)
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("the search happened inside the entry trigger's resolution")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Error("the reflexive trigger should search when it resolves")
	}
	if spec, _ := effects.Lookup(effectstest.B06RiveteersOverlookOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("the reflexive trigger (#636) closed the card's only gap")
	}
}

func TestB06RiveteersOverlookDoesNothingIfItLeftBeforeResolving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
	)
	before := me.Life
	overlook := effectstest.PlayLandFromHand(t, g, "Riveteers Overlook", effectstest.B06RiveteersOverlookOracle)
	if effectstest.TriggerOnStack(g, overlook) == nil {
		t.Fatal("the ETB should be on the stack")
	}
	// Bounced in response: "when you do" never happens.
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(overlook) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("no sacrifice, no search")
	}
	if me.Life != before {
		t.Errorf("no sacrifice, no life: %d → %d", before, me.Life)
	}
	if !me.Hand.Contains(overlook) {
		t.Error("the bounced Overlook should be in hand")
	}
}

func TestB06StartingTownEntersUntappedEarlyAndTappedLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	early := effectstest.PlayLandFromHand(t, g, "Starting Town", effectstest.B06StartingTownOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, early, "Starting Town (turn 1)")

	g2 := effectstest.NewCatalogGame(t)
	g2.Seats[g2.Turn.ActiveSeat].TurnsBegun = 4
	late := effectstest.PlayLandFromHand(t, g2, "Starting Town", effectstest.B06StartingTownOracle)
	effectstest.Top100AssertEnteredTapped(t, g2, late, "Starting Town (turn 4)")
}

func TestB06StartingTownPutInDuringAnOpponentsTurnEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0]
	id := uuid.New()
	owner.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Starting Town",
		TypeLine:   "Land — Town",
		OracleID:   effectstest.B06StartingTownOracle,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	g.WithWriteLock(func() {
		g.Turn.Seq++
		g.Turn.ActiveSeat = 1
		g.Turn.PriorityHolder = 1
		if _, err := g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{Controller: owner.ID}); err != nil {
			t.Fatalf("put Starting Town onto the battlefield: %v", err)
		}
	})
	effectstest.Top100AssertEnteredTapped(t, g, id, "Starting Town on an opponent's turn")
}

func TestB06StartingTownColouredHalfCostsALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	town := effectstest.SeedPermanentWithOracle(g, me.ID, "Starting Town", "Land — Town", effectstest.B06StartingTownOracle)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, town, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the colourless half: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" || me.Life != before {
		t.Errorf("the {C} half is free: pool %v, life %d → %d", got, before, me.Life)
	}
	fresh := effectstest.SeedPermanentWithOracle(g, me.ID, "Starting Town", "Land — Town", effectstest.B06StartingTownOracle)
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the coloured half: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour means all five, got %+v", pick)
	}
}

// --- creatures and rocks -------------------------------------------

func TestB06GildedGooseMakesFoodAndEatsItForAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	goose := effectstest.CastCatalogSpell(t, g, "Gilded Goose", "Creature — Bird", effectstest.B06GildedGooseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Food") != 1 {
		t.Fatal("the ETB should make one Food")
	}
	if !effectstest.HasEffectiveKeyword(t, g, goose, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}

	// A fresh, unsick Goose for the two tap abilities.
	ready := effectstest.PushCatalogPermanent(g, me.ID, "Gilded Goose", "Creature — Bird", effectstest.B06GildedGooseOracle, false)
	effectstest.B06AddMana(me, "G", "C")
	if err := g.ActivateCatalogAbility(me.ID, ready, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the Food-maker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Food") != 2 {
		t.Fatal("{1}{G}, {T} should make a second Food")
	}

	eater := effectstest.PushCatalogPermanent(g, me.ID, "Gilded Goose", "Creature — Bird", effectstest.B06GildedGooseOracle, false)
	if err := g.ActivateManaAbility(me.ID, eater, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{eater}}); err == nil {
		t.Fatal("the Goose is not a Food and must not be able to eat itself")
	}
	food := effectstest.FindBattlefieldByName(g, "Food")
	if err := g.ActivateManaAbility(me.ID, eater, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{food}}); err != nil {
		t.Fatalf("eat a Food: %v", err)
	}
	if g.Battlefield.Contains(food) {
		t.Error("the Food was not sacrificed")
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour means all five, got %+v", pick)
	}
}

func TestB06OrnithopterFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thopter := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter",
		OracleID: effectstest.B06OrnithopterOracle, Power: 0, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if !effectstest.HasEffectiveKeyword(t, g, thopter, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
}

func TestB06MossbornHydraEntersWithACounterAndDoublesOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hydra := effectstest.CastCatalogSpell(t, g, "Mossborn Hydra", "Creature — Elemental Hydra", effectstest.B06MossbornHydraOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hydra) {
		t.Fatal("a 0/0 that enters with a counter must survive the state check")
	}
	if effectstest.CounterCount(g, hydra, "+1/+1") != 1 {
		t.Fatalf("counters on entry = %d, want 1", effectstest.CounterCount(g, hydra, "+1/+1"))
	}
	if !effectstest.HasEffectiveKeyword(t, g, hydra, "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}

	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, hydra, "+1/+1") != 2 {
		t.Errorf("one landfall: %d counters, want 2", effectstest.CounterCount(g, hydra, "+1/+1"))
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Forest", TypeLine: "Token Land — Forest"}, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, hydra, "+1/+1") != 4 {
		t.Errorf("two landfalls: %d counters, want 4 (doubling, not +1)", effectstest.CounterCount(g, hydra, "+1/+1"))
	}
	// An opponent's land is not "a land you control".
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Forest", TypeLine: "Token Land — Forest"}, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, hydra, "+1/+1") != 4 {
		t.Error("an opponent's land triggered my landfall")
	}
}

func TestB06MeteorGolemDestroysAnOpponentsNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	mine := effectstest.SeedCreature(g, "My Bear", me.ID)

	effectstest.CastCatalogSpell(t, g, "Meteor Golem", "Artifact Creature — Golem", effectstest.B06MeteorGolemOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != rock {
		t.Fatalf("legal targets %v, want exactly the opponent's artifact (not their land, not my creature)", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the chosen permanent should be destroyed")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was never a legal target")
	}
}

func TestB06TalrandMakesADrakePerInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Talrand, Sky Summoner", "Legendary Creature — Merfolk Wizard", effectstest.B06TalrandOracle, false)

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	if effectstest.CountBattlefieldNamed(g, me.ID, "Drake") != 0 {
		t.Fatal("the Drake arrives when the trigger resolves, not at cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Drake") != 1 {
		t.Fatal("an instant should make one Drake")
	}
	drake := effectstest.FindBattlefieldByName(g, "Drake")
	if !effectstest.HasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("the Drake has flying")
	}

	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Drake") != 1 {
		t.Error("a creature spell is not an instant or sorcery")
	}
}

func TestB06WarstormSurgeShootsForTheNewcomersPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Warstorm Surge", "Enchantment", effectstest.B06WarstormSurgeOracle, false)
	before := opp.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Beast"), 1) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("a 3/3 entering should deal 3: %d → %d", before, opp.Life)
	}

	// An opponent's creature is not "you control".
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("3/3 green Beast"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's creature must not trigger the Surge")
	}
}

func TestB06RisingOfTheDayGrantsHasteAndPumpsLegends(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Rising of the Day", "Enchantment", effectstest.B06RisingOfTheDayOracle, false)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	legend := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Legend", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") || !effectstest.HasEffectiveKeyword(t, g, legend, "haste") {
		t.Error("creatures you control have haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature does not")
	}
	if p, tough := effectstest.EffectivePower(t, g, legend), effectstest.EffectiveToughness(t, g, legend); p != 3 || tough != 2 {
		t.Errorf("legendary creature = %d/%d, want 3/2", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a nonlegendary creature = %d power, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's legend = %d power, want 2", p)
	}
}

// --- Whip of Erebos ------------------------------------------------

func TestB06WhipOfErebosGrantsLifelinkToYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Whip of Erebos", "Legendary Enchantment Artifact", effectstest.B06WhipOfErebosOracle, false)
	mine := effectstest.SeedCreature(g, "Bear", me.ID)
	theirs := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "lifelink") {
		t.Error("creatures you control have lifelink")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("an opponent's creature does not")
	}
}

func TestB06WhipOfErebosReturnsWithHasteAndExilesAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	whip := effectstest.PushCatalogPermanent(g, me.ID, "Whip of Erebos", "Legendary Enchantment Artifact", effectstest.B06WhipOfErebosOracle, false)
	dead := effectstest.SeedGraveyardCreature(me, "Giant", "{4}{B}")
	effectstest.B06AddMana(me, "B", "B", "C", "C")

	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature should be back on the battlefield")
	}
	if !effectstest.HasEffectiveKeyword(t, g, dead, "haste") {
		t.Error("it gains haste")
	}
	if card, _ := effectstest.BattlefieldCard(g, whip); !card.Tapped {
		t.Error("the ability has a tap cost")
	}

	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dead) || !effectstest.ExileHas(g, dead) {
		t.Error("at the beginning of the next end step the creature is exiled")
	}
}

func TestB06WhipOfErebosExilesInsteadOfLettingItDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	whip := effectstest.PushCatalogPermanent(g, me.ID, "Whip of Erebos", "Legendary Enchantment Artifact", effectstest.B06WhipOfErebosOracle, false)
	dead := effectstest.SeedGraveyardCreature(me, "Giant", "{4}{B}")
	effectstest.B06AddMana(me, "B", "B", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Sacrificed to something: it goes to exile, not back to the yard
	// for a second whip.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(dead) })
	if me.Graveyard.Contains(dead) {
		t.Error("the whipped creature must not return to the graveyard")
	}
	if !effectstest.ExileHas(g, dead) {
		t.Error("it is exiled instead")
	}
}

// TestB06WhipOfErebosExilesInsteadOfBouncing pins the fix for the
// caveat #539's exit-primitive rework closed: BounceToHandForEffect
// now routes through routeCardToZoneLocked / executeZoneRouteLocked,
// the same CR 614 replacement window every other exit takes, so a
// whipped creature bounced to hand in response to the end-step exile
// is redirected to exile exactly like a sacrifice or a destroy.
func TestB06WhipOfErebosExilesInsteadOfBouncing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	whip := effectstest.PushCatalogPermanent(g, me.ID, "Whip of Erebos", "Legendary Enchantment Artifact", effectstest.B06WhipOfErebosOracle, false)
	dead := effectstest.SeedGraveyardCreature(me, "Giant", "{4}{B}")
	effectstest.B06AddMana(me, "B", "B", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Bounced in response to something: it goes to exile, not to hand
	// where it could be replayed for a second whip.
	if err := g.BounceToHandForEffect(dead); err != nil {
		t.Fatalf("BounceToHandForEffect: %v", err)
	}
	if me.Hand.Contains(dead) {
		t.Error("the whipped creature must not return to hand")
	}
	if !effectstest.ExileHas(g, dead) {
		t.Error("it is exiled instead")
	}
}

func TestB06WhipOfErebosIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	whip := effectstest.PushCatalogPermanent(g, me.ID, "Whip of Erebos", "Legendary Enchantment Artifact", effectstest.B06WhipOfErebosOracle, false)
	dead := effectstest.SeedGraveyardCreature(me, "Giant", "{4}{B}")
	effectstest.B06AddMana(me, "B", "B", "C", "C")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err == nil {
		t.Error("activated in combat: 'activate only as a sorcery'")
	}
}

// --- Spelunking ----------------------------------------------------

func TestB06SpelunkingDrawsOnEntryAndLetsYourTaplandsEnterUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	spelunking := effectstest.CastCatalogSpell(t, g, "Spelunking", "Enchantment", effectstest.B06SpelunkingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d → %d, want +1", before, me.Hand.Size())
	}

	// A gain land carries its own enters-tapped clause; both apply to
	// the same entry, so the land's controller orders them (CR 616.1)
	// and choosing Spelunking's last is what untaps it.
	land := effectstest.PlayLandFromHand(t, g, "Jungle Hollow", effectstest.B06JungleHollowOracle)
	order := effectstest.B06ReplacementOrderFor(g, me.ID)
	if order == nil {
		t.Fatal("two entry replacements applied but no CR 616 ordering prompt")
	}
	// The entering land's own clause is minted from a separate ID
	// base the meta lookup does not decode, so identify Spelunking's
	// by source and take the other as the land's.
	var landEff, spelunkingEff game.ReplacementEffectID
	for _, id := range order.ReplacementEffectIDs {
		if _, src := g.ReplacementOptionMetaForEffect(id); src == spelunking {
			spelunkingEff = id
		} else {
			landEff = id
		}
	}
	if len(order.ReplacementEffectIDs) != 2 || landEff == 0 || spelunkingEff == 0 {
		t.Fatalf("the prompt should list the land's clause and Spelunking's, got %v", order.ReplacementEffectIDs)
	}
	if err := g.ResolveReplacementOrder(order.ID, me.ID, []game.ReplacementEffectID{landEff, spelunkingEff}); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	effectstest.Top100AssertEnteredUntapped(t, g, land, "Jungle Hollow under Spelunking")
	effectstest.PassPriorityAroundTable(t, g)
}

func TestB06SpelunkingDoesNotUntapAnOpponentsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, opp.ID, "Spelunking", "Enchantment", effectstest.B06SpelunkingOracle, false)
	land := effectstest.PlayLandFromHand(t, g, "Scoured Barrens", effectstest.B06ScouredBarrensOracle)
	if effectstest.B06ReplacementOrderFor(g, g.Seats[g.Turn.ActiveSeat].ID) != nil {
		t.Fatal("an opponent's Spelunking must not apply to my land")
	}
	effectstest.Top100AssertEnteredTapped(t, g, land, "Scoured Barrens under an opponent's Spelunking")
	effectstest.PassPriorityAroundTable(t, g)
}

// #732: the caveat Spelunking shipped with said a land another effect
// put onto the battlefield TAPPED still entered tapped, because the
// fetching spell's clause was OR-ed in after the CR 614 pipeline had
// run. It is not any more — searchEnterBattlefieldLocked seeds it onto
// the entry event before the window opens (#478, #964) — so the
// replacement sees the tapped entry and clears it, exactly as it does
// for a land's own clause.
//
// Cultivate is the shape: "search your library for a basic land card,
// put it onto the battlefield TAPPED". The fetched Forest has no
// enters-tapped clause of its own, so Spelunking's is the only
// applicable replacement and there is no ordering prompt.
func TestB06SpelunkingUntapsALandFetchedTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Spelunking", "Enchantment", effectstest.B06SpelunkingOracle, false)

	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	effectstest.CastCatalogSpell(t, g, "Cultivate", "Sorcery", effectstest.B06CultivateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)

	fetched := effectstest.FindBattlefieldByName(g, "Forest")
	if fetched == uuid.Nil {
		t.Fatal("Cultivate put no Forest onto the battlefield")
	}
	effectstest.Top100AssertEnteredUntapped(t, g, fetched, "a Cultivate-fetched Forest under Spelunking")
	// It ENTERED untapped rather than entering tapped and untapping —
	// the distinction shocklands_test.go exists to make, and the one a
	// Tapped assertion alone cannot see.
	if n := effectstest.UntapEventsFor(g, fetched); n != 0 {
		t.Errorf("%d untap events on the fetched Forest — it entered tapped and was untapped after", n)
	}
	if effectstest.B06ReplacementOrderFor(g, me.ID) != nil {
		t.Error("a basic with no clause of its own must not raise a CR 616 ordering prompt")
	}
}

// The second fetch route the issue named, and a different caller: an
// ETB trigger's search rather than a sorcery's.
func TestB06SpelunkingUntapsASolemnFetchedLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Spelunking", "Enchantment", effectstest.B06SpelunkingOracle, false)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})

	effectstest.CastCatalogSpell(t, g, "Solemn Simulacrum", "Artifact Creature — Golem", effectstest.B06SolemnSimulacrumOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)

	fetched := effectstest.FindBattlefieldByName(g, "Island")
	if fetched == uuid.Nil {
		t.Fatal("Solemn Simulacrum put no Island onto the battlefield")
	}
	effectstest.Top100AssertEnteredUntapped(t, g, fetched, "a Solemn-fetched Island under Spelunking")
	if n := effectstest.UntapEventsFor(g, fetched); n != 0 {
		t.Errorf("%d untap events on the fetched Island — it entered tapped and was untapped after", n)
	}
}

// The other half of the same seam, and the one that proves the entry
// site can PAUSE: a fetched land that carries its own enters-tapped
// clause has two applicable replacements, so CR 616.1 asks the
// controller to order them — from a search, which #478 made
// resumable. Answering finishes the fetch; nothing is stranded.
func TestB06SpelunkingOrdersAFetchedTaplandAndResumes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spelunking := effectstest.PushCatalogPermanent(g, me.ID, "Spelunking", "Enchantment", effectstest.B06SpelunkingOracle, false)
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		game.Card{Name: "Simic Guildgate", TypeLine: "Land — Gate", OracleID: effectstest.B16SimicGuildgateOracle},
	)

	effectstest.CastCatalogSpell(t, g, "Circuitous Route", "Sorcery", effectstest.B16CircuitousRouteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest", "Simic Guildgate")

	// The basic came in untapped with no prompt; the Guildgate is
	// still in flight behind the ordering question.
	effectstest.Top100AssertEnteredUntapped(t, g, effectstest.FindBattlefieldByName(g, "Forest"), "a fetched Forest under Spelunking")
	order := effectstest.B06ReplacementOrderFor(g, me.ID)
	if order == nil {
		t.Fatal("a fetched Guildgate under Spelunking raised no CR 616 ordering prompt")
	}
	var landEff, spelunkingEff game.ReplacementEffectID
	for _, id := range order.ReplacementEffectIDs {
		if _, src := g.ReplacementOptionMetaForEffect(id); src == spelunking {
			spelunkingEff = id
		} else {
			landEff = id
		}
	}
	if len(order.ReplacementEffectIDs) != 2 || landEff == 0 || spelunkingEff == 0 {
		t.Fatalf("the prompt should list the Guildgate's clause and Spelunking's, got %v", order.ReplacementEffectIDs)
	}
	// Spelunking's last is the untapping order (CR 616.1).
	if err := g.ResolveReplacementOrder(order.ID, me.ID, []game.ReplacementEffectID{landEff, spelunkingEff}); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}

	gate := effectstest.FindBattlefieldByName(g, "Simic Guildgate")
	if gate == uuid.Nil {
		t.Fatal("the fetched Guildgate was stranded by the ordering prompt")
	}
	effectstest.Top100AssertEnteredUntapped(t, g, gate, "a fetched Simic Guildgate under Spelunking")
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts still open after the order was answered", len(g.PendingChoices))
	}
}

// --- Springbloom Druid ---------------------------------------------

func TestB06SpringbloomDruidSacrificesALandThenFetchesTwoBasicsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Command Tower", "Land"),
	)
	effectstest.CastAndResolveCreature(t, g, "Springbloom Druid", "Creature — Elf Druid", effectstest.B06SpringbloomDruidOracle)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("#1026: the land is no longer a trigger target, so nothing is picked on announce")
	}
	effectstest.PassPriorityAroundTable(t, g)

	// The choice is made AT RESOLUTION now (#1026), and the search is
	// the sacrifice run's continuation — so the order is printed: the
	// land goes first, and the basics cannot be offered as the land to
	// sacrifice.
	sac := effectstest.SacrificeChoiceFor(g, me.ID)
	if sac == nil {
		t.Fatal("the ability should ask which land to sacrifice at resolution")
	}
	if len(sac.SacrificeOptions) != 1 || sac.SacrificeOptions[0] != forest {
		t.Errorf("sacrifice offered %v, want the one land %v", sac.SacrificeOptions, forest)
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("the search is printed AFTER the sacrifice and must not be queued beside it")
	}
	effectstest.AnswerSacrifice(t, g, me.ID, forest)

	if g.Battlefield.Contains(forest) || !me.Graveyard.Contains(forest) {
		t.Fatal("the chosen land should be sacrificed")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the search should ask which basics to take")
	}
	if effectstest.SearchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("a nonbasic is not offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Plains", "Island")
	for _, name := range []string{"Plains", "Island"} {
		id := effectstest.FindBattlefieldByName(g, name)
		if card, ok := effectstest.BattlefieldCard(g, id); !ok || !card.Tapped {
			t.Errorf("%s should be on the battlefield tapped", name)
		}
	}
}

func TestB06SpringbloomDruidDeclinedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
	)
	effectstest.CastAndResolveCreature(t, g, "Springbloom Druid", "Creature — Elf Druid", effectstest.B06SpringbloomDruidOracle)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Error("declining keeps the land")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("'if you do' — no sacrifice, no search")
	}
}

// TestB06SpringbloomDruidWithNoLandSearchesForNothing is #1026's "if
// you do", and the corner the trigger-target shape could not express:
// a controller with no land is asked the printed "you may" and
// searches for nothing, rather than having the whole trigger removed
// by CR 603.3d for want of a target the card does not print.
//
// The same path covers the response case the caveat was about — say
// yes, then have your only land removed before the ability resolves:
// the sacrifice prompt is never queued, the run's answer is empty, and
// "if you do" is false.
func TestB06SpringbloomDruidWithNoLandSearchesForNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Plains", "Basic Land — Plains"))
	effectstest.CastAndResolveCreature(t, g, "Springbloom Druid", "Creature — Elf Druid", effectstest.B06SpringbloomDruidOracle)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the ability targets nothing, so nothing is picked on announce")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("a seat with no land is asked nothing (CR 701.21a's 'if you can')")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("'if you do' — nothing was sacrificed, so there is no search")
	}
}

// TestB06SpringbloomDruidSacrificesALandPlayedInResponse is the other
// half of #1026, and the reason the choice moved: at trigger time the
// pick was frozen when the ability went on the stack, so a land that
// arrived in response was not a legal answer. At resolution it is.
func TestB06SpringbloomDruidSacrificesALandPlayedInResponse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Wastes", "Basic Land — Wastes"))
	effectstest.CastAndResolveCreature(t, g, "Springbloom Druid", "Creature — Elf Druid", effectstest.B06SpringbloomDruidOracle)
	// In response: a second land arrives. The old shape had already
	// locked its target in.
	late := effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	sac := effectstest.SacrificeChoiceFor(g, me.ID)
	if sac == nil {
		t.Fatal("the ability should ask which land to sacrifice at resolution")
	}
	if len(sac.SacrificeOptions) != 2 {
		t.Fatalf("both lands are legal answers at resolution, got %v", sac.SacrificeOptions)
	}
	effectstest.AnswerSacrifice(t, g, me.ID, late)
	if !g.Battlefield.Contains(first) || !me.Graveyard.Contains(late) {
		t.Error("the land chosen at resolution is the one sacrificed")
	}
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Error("'if you do' — the sacrifice happened, so the search follows")
	}
}

// --- Bloodchief Ascension ------------------------------------------

func TestB06BloodchiefAscensionQuestsWhenAnOpponentLostTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	asc := effectstest.PushCatalogPermanent(g, me.ID, "Bloodchief Ascension", "Enchantment", effectstest.B06BloodchiefAscensionOracle, false)

	// Damage counts ("Damage causes loss of life") — a Bolt is three.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	if effectstest.OwedMays(g, me.ID) == 0 {
		t.Fatal("an opponent lost 3 this turn: the end step should ask about the counter")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, asc, "quest") != 1 {
		t.Errorf("quest counters = %d, want 1", effectstest.CounterCount(g, asc, "quest"))
	}

	// Next turn: one life lost, by a drain, is not two.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Error("one life is not 'two or more'")
	}
	effectstest.PassPriorityAroundTable(t, g)

	// The turn after: two, spread as one damage and one drain, on the
	// SAME opponent — and "each end step" is any player's.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1)
		_ = g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 1)
	})
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	if effectstest.OwedMays(g, me.ID) == 0 {
		t.Fatal("1 + 1 on one opponent is two")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, asc, "quest") != 2 {
		t.Errorf("quest counters = %d, want 2", effectstest.CounterCount(g, asc, "quest"))
	}
}

func TestB06BloodchiefAscensionMyOwnLossDoesNotCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Bloodchief Ascension", "Enchantment", effectstest.B06BloodchiefAscensionOracle, false)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -5) })
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Error("my own life loss is not an opponent's")
	}
}

// b06QuestedAscension pushes a Bloodchief Ascension with three quest
// counters under `controller` — the online state.
func b06QuestedAscension(g *game.Game, controller uuid.UUID, counters int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bloodchief Ascension", TypeLine: "Enchantment",
		OracleID: effectstest.B06BloodchiefAscensionOracle, Owner: controller, Controller: controller,
		Counters: map[string]int{"quest": counters},
	})
}

func TestB06BloodchiefAscensionDrainsOncePerCardIntoAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b06QuestedAscension(g, me.ID, 3)
	myLife, oppLife := me.Life, opp.Life

	// A permanent dying — the battlefield-leave path.
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if n := effectstest.B06TriggerPromptsFor(g, me.ID); n != 1 {
		t.Fatalf("a creature dying: %d prompts, want exactly 1", n)
	}
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if opp.Life != oppLife-2 || me.Life != myLife+2 {
		t.Errorf("after the drain: opp %d → %d, me %d → %d; want -2 / +2", oppLife, opp.Life, myLife, me.Life)
	}

	// A discard.
	effectstest.HandCard(opp, "Card", "Sorcery")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	if n := effectstest.B06TriggerPromptsFor(g, me.ID); n != 1 {
		t.Fatalf("a discard: %d prompts, want exactly 1", n)
	}
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)

	// A mill of two is two cards, two triggers.
	effectstest.StapleLibraryCard(opp, "Top", "Instant")
	effectstest.StapleLibraryCard(opp, "Next", "Instant")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 2) })
	if n := effectstest.B06TriggerPromptsFor(g, me.ID); n != 2 {
		t.Fatalf("milling two: %d prompts, want exactly 2", n)
	}
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if opp.Life != oppLife-8 || me.Life != myLife+8 {
		t.Errorf("four cards in: opp %d → %d, me %d → %d; want -8 / +8", oppLife, opp.Life, myLife, me.Life)
	}

	// Declining loses nothing and gains nothing.
	effectstest.StapleLibraryCard(opp, "Declined", "Instant")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, false)
	if opp.Life != oppLife-8 || me.Life != myLife+8 {
		t.Error("'you may' — declining must change nothing")
	}
}

func TestB06BloodchiefAscensionSeesSpellsResolvingAndBeingCountered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b06QuestedAscension(g, me.ID, 3)
	oppLife := opp.Life

	// An opponent's instant resolving lands in their graveyard.
	effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[2].ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B06TriggerPromptsFor(g, me.ID); n != 1 {
		t.Fatalf("a resolved spell: %d prompts, want exactly 1", n)
	}
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if opp.Life != oppLife-2 {
		t.Errorf("opp %d → %d, want -2", oppLife, opp.Life)
	}

	// A countered one lands there too.
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[2].ID}})
	effectstest.CastCatalogSpell(t, g, "Counterspell", "Instant", "cc187110-1148-4090-bbb8-e205694a39f5",
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	// Two cards reached graveyards: the Bolt (opponent's — counts)
	// and my Counterspell (mine — does not).
	if n := effectstest.B06TriggerPromptsFor(g, me.ID); n != 1 {
		t.Fatalf("a countered spell: %d prompts, want exactly 1 (my own spell does not count)", n)
	}
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if opp.Life != oppLife-4 {
		t.Errorf("opp %d → %d, want -4", oppLife, opp.Life)
	}
}

func TestB06BloodchiefAscensionIsOfflineBelowThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b06QuestedAscension(g, me.ID, 2)
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if n := effectstest.B06TriggerPromptsFor(g, me.ID); n != 0 {
		t.Errorf("two counters: %d prompts, want none", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- spells --------------------------------------------------------

func TestB06IntoTheFloodMawBouncesAnOpponentsCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	mine := effectstest.SeedCreature(g, "My Bear", me.ID)
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")
	legal := effectstest.LegalCards(g, me.ID, effectstest.B06IntoTheFloodMawOracle)
	if !legal[theirs] || legal[mine] || legal[rock] {
		t.Fatalf("legal targets should be exactly the opponent's creature: %v", legal)
	}
	effectstest.CastCatalogSpell(t, g, "Into the Flood Maw", "Instant", effectstest.B06IntoTheFloodMawOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the creature should be back in its owner's hand")
	}
}

func TestB06GrimTutorFindsAnyCardThenCostsThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	needle := effectstest.StapleLibraryCard(me, "Needle", "Sorcery")
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Grim Tutor", "Sorcery", effectstest.B06GrimTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("the tutor should ask which card to take")
	}
	if me.Life != before {
		t.Error("the life is lost after the search, not before")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, needle)
	if !me.Hand.Contains(needle) {
		t.Error("the chosen card did not reach the hand")
	}
	if me.Life != before-3 {
		t.Errorf("life %d → %d, want -3", before, me.Life)
	}
}

func TestB06TypedTutorsOfferOnlyTheirType(t *testing.T) {
	for _, tc := range []struct{ name, oracle, want, other string }{
		{"Idyllic Tutor", effectstest.B06IdyllicTutorOracle, "Enchantment", "Creature — Bear"},
		{"Eladamri's Call", effectstest.B06EladamrisCallOracle, "Creature — Bear", "Enchantment"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.SeedSearchLibrary(me,
			game.Card{Name: "Wanted A", TypeLine: tc.want},
			game.Card{Name: "Wanted B", TypeLine: tc.want},
			game.Card{Name: "Wrong", TypeLine: tc.other},
		)
		effectstest.CastCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.SearchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("%s: no search prompt", tc.name)
		}
		if effectstest.SearchOptionNamed(g, c, "Wrong") != uuid.Nil {
			t.Errorf("%s offered a %s", tc.name, tc.other)
		}
		effectstest.AnswerSearchNamed(t, g, me.ID, "Wanted A")
		if !effectstest.B02bHandHasNamed(me, "Wanted A") {
			t.Errorf("%s: the chosen card did not reach the hand", tc.name)
		}
	}
}

func TestB06ReturnToNatureThreeModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")
	effectstest.CastModal(t, g, "Return to Nature", "Instant", effectstest.B06ReturnToNatureOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("mode 0 destroys the artifact")
	}

	aura := effectstest.SeedPermanentFor(g, opp.ID, "Aura", "Enchantment")
	effectstest.CastModal(t, g, "Return to Nature", "Instant", effectstest.B06ReturnToNatureOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: aura}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("mode 1 destroys the enchantment")
	}

	// Mode 2 reads any graveyard — the caster's own included.
	buried := effectstest.PushGraveyardCardForTest(me, "Buried")
	effectstest.CastModal(t, g, "Return to Nature", "Instant", effectstest.B06ReturnToNatureOracle, []int{2},
		[]game.TargetRef{{Kind: game.TargetCard, ID: buried}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(buried) || !effectstest.ExileHas(g, buried) {
		t.Error("mode 2 exiles the card from the graveyard")
	}
}
