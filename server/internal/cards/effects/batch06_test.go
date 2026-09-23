package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch06_test.go — card-level coverage for the card-coverage
// roadmap's batch 06 (#299, `edhrec_rank` 694–797): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or step. Helpers from
// batch01_test.go / batch04_test.go are reused by name; new ones are
// b06-prefixed.

const (
	b06IntoTheFloodMawOracle   = "8cb36a67-9206-4665-a03f-64f52ba559c4"
	b06GrimTutorOracle         = "e62f8d69-a559-4f13-a5c9-5fb750b4af2c"
	b06CastleArdenvaleOracle   = "f8f4fc60-725d-46d8-8e8f-e68e00d20589"
	b06WhipOfErebosOracle      = "53987a39-c18c-4c13-b1ea-fd1b2a369f9e"
	b06FireDiamondOracle       = "97b477d8-2e05-475e-8ed6-7d680cb21cd9"
	b06RiveteersOverlookOracle = "5548ff43-e5f6-4a63-8562-a2b1de06d6f5"
	b06RisingOfTheDayOracle    = "434a20f8-3f87-4004-9155-0f196fc2257e"
	b06VernalFenOracle         = "40544d12-0391-4a61-af95-9b8ec01ed8fc"
	b06WarstormSurgeOracle     = "42fb1a1c-ab3d-4cdc-a6ff-a591f7481583"
	b06ForbiddenOrchardOracle  = "cfd60d1f-9832-4408-b84e-0fd3018b015b"
	b06OranRiefOracle          = "e88027a6-24cc-4a8b-86db-734f26149ea8"
	b06ElvesOfDeepShadowOracle = "20347559-95a9-4689-bb79-c5bb3809b719"
	b06SpelunkingOracle        = "2962fe4c-bf48-454b-8a6b-0f8253352ae8"
	// #732's two fetch routes: a sorcery's search and an ETB
	// trigger's. Neither card is a batch-06 card; the IDs live here
	// because the Spelunking tests below are the only callers.
	b06CultivateOracle           = "8b755881-a72d-4e21-a369-d2924eb4585a"
	b06SolemnSimulacrumOracle    = "00c0543c-2a1f-4425-8283-4062d74a1637"
	b06ScouredBarrensOracle      = "d37f858e-03c8-4594-9b92-cd03699a1591"
	b06StartingTownOracle        = "d04e0975-f401-41b8-a9db-9bcf9cbbce66"
	b06IdyllicTutorOracle        = "57c9ff89-dd30-467b-bb3e-499eeea8cb94"
	b06BloodchiefAscensionOracle = "d5ea905b-4bb6-48d0-9082-c472703db550"
	b06ReturnToNatureOracle      = "777b8ec4-a783-4297-96b7-4f200d0eb734"
	b06DryadArborOracle          = "e996cd67-739c-40f4-b276-0042acf26c71"
	b06MossbornHydraOracle       = "84122d81-9634-4d3f-85a5-f6cf4303691a"
	b06EladamrisCallOracle       = "4acb6612-54e8-428d-acb6-c7259a5ad6a8"
	b06OrnithopterOracle         = "a3a98bc9-caa0-49b7-951c-fe4e4f54e4ba"
	b06SpringbloomDruidOracle    = "788a4896-3414-40bf-b391-d2efea2fb5f9"
	b06JungleHollowOracle        = "6de714e1-446d-4fb9-9e3d-bcd3ec6af9ca"
	b06TalrandOracle             = "ea1eb902-a23c-44ff-9169-19baf71de238"
	b06CastleLocthwainOracle     = "be811e70-aaaa-41f3-bf9e-5d3f9f719b49"
	b06MeteorGolemOracle         = "d9f11aa1-9219-42a8-85a9-a8f204160706"
	b06SkyDiamondOracle          = "2224b6e0-c5ff-45d0-84e3-83758c5fc99f"
	b06GildedGooseOracle         = "f2f09757-1931-47c0-a5f0-39280445489d"
)

// b06AddMana drops the listed colours straight into a player's pool
// so an activated ability's mana component can be paid.
func b06AddMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

// b06TriggerPromptsFor counts the yes/no trigger prompts addressed to
// chooser — the "exactly one trigger per card" canary Bloodchief
// Ascension needs.
func b06TriggerPromptsFor(g *game.Game, chooser uuid.UUID) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			n++
		}
	}
	return n
}

// b06AnswerAllTriggerPrompts answers every trigger prompt addressed
// to chooser the same way, then settles the stack.
func b06AnswerAllTriggerPrompts(t *testing.T, g *game.Game, chooser uuid.UUID, apply bool) {
	t.Helper()
	for i := 0; i < 16 && latestTriggerPrompt(g, chooser) != nil; i++ {
		answerLatestTriggerPrompt(t, g, chooser, apply)
	}
	passPriorityAroundTable(t, g)
}

// b06GreenCreature pushes a green 2/2 onto the battlefield through the
// real token path, so it carries an EventETB (Oran-Rief reads it).
func b06GreenCreature(t *testing.T, g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(controller, game.Card{
			Name: name, TypeLine: "Token Creature — Elf", Power: 2, Toughness: 2, Colors: []string{"G"},
		}, 1)
	})
	passPriorityAroundTable(t, g)
	return battlefieldIDNamed(g, controller, name)
}

// b06ReplacementOrderFor returns the open CR 616 ordering prompt
// addressed to chooser, or nil.
func b06ReplacementOrderFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The cycle rows
// (Vernal Fen, the two Diamonds, the two gain lands) are loops over
// tables, so a transposed row is invisible until someone plays that
// exact card.
func TestBatch06CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b06IntoTheFloodMawOracle:     "Into the Flood Maw",
		b06GrimTutorOracle:           "Grim Tutor",
		b06CastleArdenvaleOracle:     "Castle Ardenvale",
		b06WhipOfErebosOracle:        "Whip of Erebos",
		b06FireDiamondOracle:         "Fire Diamond",
		b06RiveteersOverlookOracle:   "Riveteers Overlook",
		b06RisingOfTheDayOracle:      "Rising of the Day",
		b06VernalFenOracle:           "Vernal Fen",
		b06WarstormSurgeOracle:       "Warstorm Surge",
		b06ForbiddenOrchardOracle:    "Forbidden Orchard",
		b06OranRiefOracle:            "Oran-Rief, the Vastwood",
		b06ElvesOfDeepShadowOracle:   "Elves of Deep Shadow",
		b06SpelunkingOracle:          "Spelunking",
		b06ScouredBarrensOracle:      "Scoured Barrens",
		b06StartingTownOracle:        "Starting Town",
		b06IdyllicTutorOracle:        "Idyllic Tutor",
		b06BloodchiefAscensionOracle: "Bloodchief Ascension",
		b06ReturnToNatureOracle:      "Return to Nature",
		b06DryadArborOracle:          "Dryad Arbor",
		b06MossbornHydraOracle:       "Mossborn Hydra",
		b06EladamrisCallOracle:       "Eladamri's Call",
		b06OrnithopterOracle:         "Ornithopter",
		b06SpringbloomDruidOracle:    "Springbloom Druid",
		b06JungleHollowOracle:        "Jungle Hollow",
		b06TalrandOracle:             "Talrand, Sky Summoner",
		b06CastleLocthwainOracle:     "Castle Locthwain",
		b06MeteorGolemOracle:         "Meteor Golem",
		b06SkyDiamondOracle:          "Sky Diamond",
		b06GildedGooseOracle:         "Gilded Goose",
	}
	if len(want) != 29 {
		t.Fatalf("the batch registers 29 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness == CompletenessUnreviewed && name != "Vernal Fen" {
			t.Errorf("%s ships without a Completeness declaration", name)
		}
	}
}

// --- lands ---------------------------------------------------------

func TestB06CastleArdenvaleEntersTappedWithoutAPlains(t *testing.T) {
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Castle Ardenvale", b06CastleArdenvaleOracle)
	top100AssertEnteredTapped(t, g, id, "Castle Ardenvale")
}

func TestB06CastleArdenvaleEntersUntappedWithAPlainsTypedLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// A Godless Shrine is a Plains; a Wastes is not.
	seedLandOnBattlefield(g, me.ID, "Godless Shrine", "Land — Plains Swamp")
	id := playLandFromHand(t, g, "Castle Ardenvale", b06CastleArdenvaleOracle)
	top100AssertEnteredUntapped(t, g, id, "Castle Ardenvale")
}

func TestB06CastleArdenvaleMakesAHuman(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castle := seedPermanentWithOracle(g, me.ID, "Castle Ardenvale", "Land", b06CastleArdenvaleOracle)
	b06AddMana(me, "W", "W", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, castle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Human") != 1 {
		t.Error("the Castle should make one 1/1 Human")
	}
	if card, _ := battlefieldCard(g, castle); !card.Tapped {
		t.Error("the ability has a tap cost")
	}
}

func TestB06CastleLocthwainDrawsThenChargesTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castle := seedPermanentWithOracle(g, me.ID, "Castle Locthwain", "Land", b06CastleLocthwainOracle)
	b06AddMana(me, "B", "B", "C")
	hand, life := me.Hand.Size(), me.Life
	if err := g.ActivateCatalogAbility(me.ID, castle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
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
		{"Fire Diamond", b06FireDiamondOracle, "R"},
		{"Sky Diamond", b06SkyDiamondOracle, "U"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		rock := castCatalogSpell(t, g, tc.name, "Artifact", tc.oracle, nil)
		passPriorityAroundTable(t, g)
		top100AssertEnteredTapped(t, g, rock, tc.name)

		fresh := seedPermanentWithOracle(g, me.ID, tc.name, "Artifact", tc.oracle)
		if err := g.ActivateManaAbility(me.ID, fresh, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: ActivateManaAbility: %v", tc.name, err)
		}
		if got := batch01PoolColors(me); len(got) != 1 || got[0] != tc.color {
			t.Errorf("%s: pool %v, want [%s]", tc.name, got, tc.color)
		}
	}
}

func TestB06DryadArborTapsForGreenOnceItIsNotSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sick := pushCatalogPermanent(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad", b06DryadArborOracle, true)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a summoning-sick Dryad Arbor tapped for mana")
	}
	ready := pushCatalogPermanent(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad", b06DryadArborOracle, false)
	if err := g.ActivateManaAbility(me.ID, ready, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
}

func TestB06ElvesOfDeepShadowTapForBlackAndHurt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := pushCatalogPermanent(g, me.ID, "Elves of Deep Shadow", "Creature — Elf Druid", b06ElvesOfDeepShadowOracle, false)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
}

func TestB06ForbiddenOrchardGivesTheChosenOpponentASpirit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[2]
	orchard := seedPermanentWithOracle(g, me.ID, "Forbidden Orchard", "Land", b06ForbiddenOrchardOracle)
	if err := g.ActivateManaAbility(me.ID, orchard, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour means all five, got %+v", pick)
	}
	// #730: the colour pick gates the table, so it is answered before
	// the Orchard's own trigger is passed around. Which colour the
	// mana is does not matter here.
	riderAnswerManaPicks(t, g, me.ID, "G")
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, opp.ID, "Spirit") != 1 {
		t.Error("the chosen opponent should get a 1/1 Spirit")
	}
	if countBattlefieldNamed(g, me.ID, "Spirit") != 0 {
		t.Error("the Orchard's controller never gets the Spirit")
	}
}

func TestB06GainLandsEnterTappedGainOneAndTapForBoth(t *testing.T) {
	for _, tc := range []struct{ name, oracle, a, b string }{
		{"Scoured Barrens", b06ScouredBarrensOracle, "W", "B"},
		{"Jungle Hollow", b06JungleHollowOracle, "B", "G"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		before := me.Life
		id := playLandFromHand(t, g, tc.name, tc.oracle)
		top100AssertEnteredTapped(t, g, id, tc.name)
		if triggerOnStack(g, id) == nil {
			t.Errorf("%s: the life is a trigger with a response window", tc.name)
		}
		passPriorityAroundTable(t, g)
		if me.Life != before+1 {
			t.Errorf("%s: life %d → %d, want +1", tc.name, before, me.Life)
		}
		spec, _ := Lookup(tc.oracle)
		if want := "{" + tc.a + "|" + tc.b + "}"; len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, want)
		}
	}
}

func TestB06VernalFenIsABattleLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	tapped := playLandFromHand(t, g, "Vernal Fen", b06VernalFenOracle)
	top100AssertEnteredTapped(t, g, tapped, "Vernal Fen (one basic)")

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	seedLandOnBattlefield(g2, me2.ID, "Swamp", "Basic Land — Swamp")
	seedLandOnBattlefield(g2, me2.ID, "Forest", "Basic Land — Forest")
	untapped := playLandFromHand(t, g2, "Vernal Fen", b06VernalFenOracle)
	top100AssertEnteredUntapped(t, g2, untapped, "Vernal Fen (two basics)")

	spec, _ := Lookup(b06VernalFenOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{B|G}" {
		t.Errorf("mana abilities %+v, want one producing {B|G}", spec.ManaAbilities)
	}
}

func TestB06OranRiefGrowsOnlyGreenCreaturesThatEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := seedPermanentWithOracle(g, me.ID, "Oran-Rief, the Vastwood", "Land", b06OranRiefOracle)
	old := seedCreature(g, "Old Bear", me.ID) // no ETB event: was here before the turn

	fresh := b06GreenCreature(t, g, me.ID, "Fresh Elf")
	theirs := b06GreenCreature(t, g, opp.ID, "Their Elf")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)
	goblin := findBattlefieldByName(g, "Goblin")

	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if counterCount(g, fresh, "+1/+1") != 1 {
		t.Error("a green creature that entered this turn should get a counter")
	}
	if counterCount(g, theirs, "+1/+1") != 1 {
		t.Error("'each green creature' is every player's")
	}
	if counterCount(g, old, "+1/+1") != 0 {
		t.Error("a creature from an earlier turn must not grow")
	}
	if counterCount(g, goblin, "+1/+1") != 0 {
		t.Error("a red creature must not grow")
	}
}

func TestB06OranRiefForgetsLastTurnsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := seedPermanentWithOracle(g, me.ID, "Oran-Rief, the Vastwood", "Land", b06OranRiefOracle)
	elf := b06GreenCreature(t, g, me.ID, "Elf")
	advanceToNextSeatsTurn(t, g)
	advanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if counterCount(g, elf, "+1/+1") != 0 {
		t.Error("'entered this turn' does not reach back to an earlier turn")
	}
}

func TestB06RiveteersOverlookSacrificesFetchesTappedAndGainsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Swamp", "Basic Land — Swamp"),
		searchTestLand("Mountain", "Basic Land — Mountain"),
		searchTestLand("Forest", "Basic Land — Forest"),
	)
	before := me.Life
	overlook := playLandFromHand(t, g, "Riveteers Overlook", b06RiveteersOverlookOracle)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(overlook) || !me.Graveyard.Contains(overlook) {
		t.Fatal("the Overlook should have sacrificed itself")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the search should ask which basic to take")
	}
	if searchOptionNamed(g, c, "Plains") != uuid.Nil {
		t.Error("a Plains is not a Swamp, Mountain or Forest")
	}
	answerSearchNamed(t, g, me.ID, "Mountain")
	mountain := findBattlefieldByName(g, "Mountain")
	if card, ok := battlefieldCard(g, mountain); !ok || !card.Tapped {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		searchTestLand("Swamp", "Basic Land — Swamp"),
		searchTestLand("Forest", "Basic Land — Forest"),
	)
	overlook := playLandFromHand(t, g, "Riveteers Overlook", b06RiveteersOverlookOracle)
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
	reflexive := triggerOnStack(g, overlook)
	if reflexive == nil {
		t.Fatal(`"when you do" should be a second trigger on the stack`)
	}
	if reflexive.Controller != me.ID {
		t.Errorf("reflexive trigger controlled by %v, want the sacrificing player", reflexive.Controller)
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("the search happened inside the entry trigger's resolution")
	}
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) == nil {
		t.Error("the reflexive trigger should search when it resolves")
	}
	if spec, _ := Lookup(b06RiveteersOverlookOracle); spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("the reflexive trigger (#636) closed the card's only gap")
	}
}

func TestB06RiveteersOverlookDoesNothingIfItLeftBeforeResolving(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		searchTestLand("Swamp", "Basic Land — Swamp"),
		searchTestLand("Forest", "Basic Land — Forest"),
	)
	before := me.Life
	overlook := playLandFromHand(t, g, "Riveteers Overlook", b06RiveteersOverlookOracle)
	if triggerOnStack(g, overlook) == nil {
		t.Fatal("the ETB should be on the stack")
	}
	// Bounced in response: "when you do" never happens.
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(overlook) })
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
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
	g := newCatalogGame(t)
	early := playLandFromHand(t, g, "Starting Town", b06StartingTownOracle)
	top100AssertEnteredUntapped(t, g, early, "Starting Town (turn 1)")

	g2 := newCatalogGame(t)
	g2.Turn.Number = 4
	late := playLandFromHand(t, g2, "Starting Town", b06StartingTownOracle)
	top100AssertEnteredTapped(t, g2, late, "Starting Town (turn 4)")
}

func TestB06StartingTownColouredHalfCostsALife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	town := seedPermanentWithOracle(g, me.ID, "Starting Town", "Land — Town", b06StartingTownOracle)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, town, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the colourless half: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" || me.Life != before {
		t.Errorf("the {C} half is free: pool %v, life %d → %d", got, before, me.Life)
	}
	fresh := seedPermanentWithOracle(g, me.ID, "Starting Town", "Land — Town", b06StartingTownOracle)
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the coloured half: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour means all five, got %+v", pick)
	}
}

// --- creatures and rocks -------------------------------------------

func TestB06GildedGooseMakesFoodAndEatsItForAnyColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	goose := castCatalogSpell(t, g, "Gilded Goose", "Creature — Bird", b06GildedGooseOracle, nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Food") != 1 {
		t.Fatal("the ETB should make one Food")
	}
	if !hasEffectiveKeyword(t, g, goose, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}

	// A fresh, unsick Goose for the two tap abilities.
	ready := pushCatalogPermanent(g, me.ID, "Gilded Goose", "Creature — Bird", b06GildedGooseOracle, false)
	b06AddMana(me, "G", "C")
	if err := g.ActivateCatalogAbility(me.ID, ready, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the Food-maker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Food") != 2 {
		t.Fatal("{1}{G}, {T} should make a second Food")
	}

	eater := pushCatalogPermanent(g, me.ID, "Gilded Goose", "Creature — Bird", b06GildedGooseOracle, false)
	if err := g.ActivateManaAbility(me.ID, eater, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{eater}}); err == nil {
		t.Fatal("the Goose is not a Food and must not be able to eat itself")
	}
	food := findBattlefieldByName(g, "Food")
	if err := g.ActivateManaAbility(me.ID, eater, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{food}}); err != nil {
		t.Fatalf("eat a Food: %v", err)
	}
	if g.Battlefield.Contains(food) {
		t.Error("the Food was not sacrificed")
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour means all five, got %+v", pick)
	}
}

func TestB06OrnithopterFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thopter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter",
		OracleID: b06OrnithopterOracle, Power: 0, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if !hasEffectiveKeyword(t, g, thopter, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
}

func TestB06MossbornHydraEntersWithACounterAndDoublesOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hydra := castCatalogSpell(t, g, "Mossborn Hydra", "Creature — Elemental Hydra", b06MossbornHydraOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hydra) {
		t.Fatal("a 0/0 that enters with a counter must survive the state check")
	}
	if counterCount(g, hydra, "+1/+1") != 1 {
		t.Fatalf("counters on entry = %d, want 1", counterCount(g, hydra, "+1/+1"))
	}
	if !hasEffectiveKeyword(t, g, hydra, "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if counterCount(g, hydra, "+1/+1") != 2 {
		t.Errorf("one landfall: %d counters, want 2", counterCount(g, hydra, "+1/+1"))
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Forest", TypeLine: "Token Land — Forest"}, 1)
	})
	passPriorityAroundTable(t, g)
	if counterCount(g, hydra, "+1/+1") != 4 {
		t.Errorf("two landfalls: %d counters, want 4 (doubling, not +1)", counterCount(g, hydra, "+1/+1"))
	}
	// An opponent's land is not "a land you control".
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Forest", TypeLine: "Token Land — Forest"}, 1)
	})
	passPriorityAroundTable(t, g)
	if counterCount(g, hydra, "+1/+1") != 4 {
		t.Error("an opponent's land triggered my landfall")
	}
}

func TestB06MeteorGolemDestroysAnOpponentsNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")
	seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	mine := seedCreature(g, "My Bear", me.ID)

	castCatalogSpell(t, g, "Meteor Golem", "Artifact Creature — Golem", b06MeteorGolemOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != rock {
		t.Fatalf("legal targets %v, want exactly the opponent's artifact (not their land, not my creature)", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the chosen permanent should be destroyed")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was never a legal target")
	}
}

func TestB06TalrandMakesADrakePerInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Talrand, Sky Summoner", "Legendary Creature — Merfolk Wizard", b06TalrandOracle, false)

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	if countBattlefieldNamed(g, me.ID, "Drake") != 0 {
		t.Fatal("the Drake arrives when the trigger resolves, not at cast")
	}
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Drake") != 1 {
		t.Fatal("an instant should make one Drake")
	}
	drake := findBattlefieldByName(g, "Drake")
	if !hasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("the Drake has flying")
	}

	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Drake") != 1 {
		t.Error("a creature spell is not an instant or sorcery")
	}
}

func TestB06WarstormSurgeShootsForTheNewcomersPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Warstorm Surge", "Enchantment", b06WarstormSurgeOracle, false)
	before := opp.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 green Beast"), 1) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("a 3/3 entering should deal 3: %d → %d", before, opp.Life)
	}

	// An opponent's creature is not "you control".
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("3/3 green Beast"), 1) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's creature must not trigger the Surge")
	}
}

func TestB06RisingOfTheDayGrantsHasteAndPumpsLegends(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Rising of the Day", "Enchantment", b06RisingOfTheDayOracle, false)
	bear := seedCreature(g, "Bear", me.ID)
	legend := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Legend", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if !hasEffectiveKeyword(t, g, bear, "haste") || !hasEffectiveKeyword(t, g, legend, "haste") {
		t.Error("creatures you control have haste")
	}
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature does not")
	}
	if p, tough := effectivePower(t, g, legend), effectiveToughness(t, g, legend); p != 3 || tough != 2 {
		t.Errorf("legendary creature = %d/%d, want 3/2", p, tough)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a nonlegendary creature = %d power, want 2", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's legend = %d power, want 2", p)
	}
}

// --- Whip of Erebos ------------------------------------------------

func TestB06WhipOfErebosGrantsLifelinkToYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Whip of Erebos", "Legendary Enchantment Artifact", b06WhipOfErebosOracle, false)
	mine := seedCreature(g, "Bear", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)
	if !hasEffectiveKeyword(t, g, mine, "lifelink") {
		t.Error("creatures you control have lifelink")
	}
	if hasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("an opponent's creature does not")
	}
}

func TestB06WhipOfErebosReturnsWithHasteAndExilesAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	whip := pushCatalogPermanent(g, me.ID, "Whip of Erebos", "Legendary Enchantment Artifact", b06WhipOfErebosOracle, false)
	dead := seedGraveyardCreature(me, "Giant", "{4}{B}")
	b06AddMana(me, "B", "B", "C", "C")

	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature should be back on the battlefield")
	}
	if !hasEffectiveKeyword(t, g, dead, "haste") {
		t.Error("it gains haste")
	}
	if card, _ := battlefieldCard(g, whip); !card.Tapped {
		t.Error("the ability has a tap cost")
	}

	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dead) || !exileHas(g, dead) {
		t.Error("at the beginning of the next end step the creature is exiled")
	}
}

func TestB06WhipOfErebosExilesInsteadOfLettingItDie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	whip := pushCatalogPermanent(g, me.ID, "Whip of Erebos", "Legendary Enchantment Artifact", b06WhipOfErebosOracle, false)
	dead := seedGraveyardCreature(me, "Giant", "{4}{B}")
	b06AddMana(me, "B", "B", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Sacrificed to something: it goes to exile, not back to the yard
	// for a second whip.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(dead) })
	if me.Graveyard.Contains(dead) {
		t.Error("the whipped creature must not return to the graveyard")
	}
	if !exileHas(g, dead) {
		t.Error("it is exiled instead")
	}
}

func TestB06WhipOfErebosIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	whip := pushCatalogPermanent(g, me.ID, "Whip of Erebos", "Legendary Enchantment Artifact", b06WhipOfErebosOracle, false)
	dead := seedGraveyardCreature(me, "Giant", "{4}{B}")
	b06AddMana(me, "B", "B", "C", "C")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err == nil {
		t.Error("activated in combat: 'activate only as a sorcery'")
	}
}

// --- Spelunking ----------------------------------------------------

func TestB06SpelunkingDrawsOnEntryAndLetsYourTaplandsEnterUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	spelunking := castCatalogSpell(t, g, "Spelunking", "Enchantment", b06SpelunkingOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d → %d, want +1", before, me.Hand.Size())
	}

	// A gain land carries its own enters-tapped clause; both apply to
	// the same entry, so the land's controller orders them (CR 616.1)
	// and choosing Spelunking's last is what untaps it.
	land := playLandFromHand(t, g, "Jungle Hollow", b06JungleHollowOracle)
	order := b06ReplacementOrderFor(g, me.ID)
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
	top100AssertEnteredUntapped(t, g, land, "Jungle Hollow under Spelunking")
	passPriorityAroundTable(t, g)
}

func TestB06SpelunkingDoesNotUntapAnOpponentsLands(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, opp.ID, "Spelunking", "Enchantment", b06SpelunkingOracle, false)
	land := playLandFromHand(t, g, "Scoured Barrens", b06ScouredBarrensOracle)
	if b06ReplacementOrderFor(g, g.Seats[g.Turn.ActiveSeat].ID) != nil {
		t.Fatal("an opponent's Spelunking must not apply to my land")
	}
	top100AssertEnteredTapped(t, g, land, "Scoured Barrens under an opponent's Spelunking")
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Spelunking", "Enchantment", b06SpelunkingOracle, false)

	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	castCatalogSpell(t, g, "Cultivate", "Sorcery", b06CultivateOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, forest)

	fetched := findBattlefieldByName(g, "Forest")
	if fetched == uuid.Nil {
		t.Fatal("Cultivate put no Forest onto the battlefield")
	}
	top100AssertEnteredUntapped(t, g, fetched, "a Cultivate-fetched Forest under Spelunking")
	// It ENTERED untapped rather than entering tapped and untapping —
	// the distinction shocklands_test.go exists to make, and the one a
	// Tapped assertion alone cannot see.
	if n := untapEventsFor(g, fetched); n != 0 {
		t.Errorf("%d untap events on the fetched Forest — it entered tapped and was untapped after", n)
	}
	if b06ReplacementOrderFor(g, me.ID) != nil {
		t.Error("a basic with no clause of its own must not raise a CR 616 ordering prompt")
	}
}

// The second fetch route the issue named, and a different caller: an
// ETB trigger's search rather than a sorcery's.
func TestB06SpelunkingUntapsASolemnFetchedLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Spelunking", "Enchantment", b06SpelunkingOracle, false)
	pushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})

	castCatalogSpell(t, g, "Solemn Simulacrum", "Artifact Creature — Golem", b06SolemnSimulacrumOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	fetched := findBattlefieldByName(g, "Island")
	if fetched == uuid.Nil {
		t.Fatal("Solemn Simulacrum put no Island onto the battlefield")
	}
	top100AssertEnteredUntapped(t, g, fetched, "a Solemn-fetched Island under Spelunking")
	if n := untapEventsFor(g, fetched); n != 0 {
		t.Errorf("%d untap events on the fetched Island — it entered tapped and was untapped after", n)
	}
}

// The other half of the same seam, and the one that proves the entry
// site can PAUSE: a fetched land that carries its own enters-tapped
// clause has two applicable replacements, so CR 616.1 asks the
// controller to order them — from a search, which #478 made
// resumable. Answering finishes the fetch; nothing is stranded.
func TestB06SpelunkingOrdersAFetchedTaplandAndResumes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spelunking := pushCatalogPermanent(g, me.ID, "Spelunking", "Enchantment", b06SpelunkingOracle, false)
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Plains", "Basic Land — Plains"),
		game.Card{Name: "Simic Guildgate", TypeLine: "Land — Gate", OracleID: b16SimicGuildgateOracle},
	)

	castCatalogSpell(t, g, "Circuitous Route", "Sorcery", b16CircuitousRouteOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Forest", "Simic Guildgate")

	// The basic came in untapped with no prompt; the Guildgate is
	// still in flight behind the ordering question.
	top100AssertEnteredUntapped(t, g, findBattlefieldByName(g, "Forest"), "a fetched Forest under Spelunking")
	order := b06ReplacementOrderFor(g, me.ID)
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

	gate := findBattlefieldByName(g, "Simic Guildgate")
	if gate == uuid.Nil {
		t.Fatal("the fetched Guildgate was stranded by the ordering prompt")
	}
	top100AssertEnteredUntapped(t, g, gate, "a fetched Simic Guildgate under Spelunking")
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts still open after the order was answered", len(g.PendingChoices))
	}
}

// --- Springbloom Druid ---------------------------------------------

func TestB06SpringbloomDruidSacrificesALandThenFetchesTwoBasicsTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Swamp", "Basic Land — Swamp"),
		searchTestLand("Command Tower", "Land"),
	)
	castAndResolveCreature(t, g, "Springbloom Druid", "Creature — Elf Druid", b06SpringbloomDruidOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("#1026: the land is no longer a trigger target, so nothing is picked on announce")
	}
	passPriorityAroundTable(t, g)

	// The choice is made AT RESOLUTION now (#1026), and the search is
	// the sacrifice run's continuation — so the order is printed: the
	// land goes first, and the basics cannot be offered as the land to
	// sacrifice.
	sac := sacrificeChoiceFor(g, me.ID)
	if sac == nil {
		t.Fatal("the ability should ask which land to sacrifice at resolution")
	}
	if len(sac.SacrificeOptions) != 1 || sac.SacrificeOptions[0] != forest {
		t.Errorf("sacrifice offered %v, want the one land %v", sac.SacrificeOptions, forest)
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("the search is printed AFTER the sacrifice and must not be queued beside it")
	}
	answerSacrifice(t, g, me.ID, forest)

	if g.Battlefield.Contains(forest) || !me.Graveyard.Contains(forest) {
		t.Fatal("the chosen land should be sacrificed")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the search should ask which basics to take")
	}
	if searchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("a nonbasic is not offered")
	}
	answerSearchNamed(t, g, me.ID, "Plains", "Island")
	for _, name := range []string{"Plains", "Island"} {
		id := findBattlefieldByName(g, name)
		if card, ok := battlefieldCard(g, id); !ok || !card.Tapped {
			t.Errorf("%s should be on the battlefield tapped", name)
		}
	}
}

func TestB06SpringbloomDruidDeclinedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Swamp", "Basic Land — Swamp"),
	)
	castAndResolveCreature(t, g, "Springbloom Druid", "Creature — Elf Druid", b06SpringbloomDruidOracle)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Error("declining keeps the land")
	}
	if searchChoiceFor(g, me.ID) != nil {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me, searchTestLand("Plains", "Basic Land — Plains"))
	castAndResolveCreature(t, g, "Springbloom Druid", "Creature — Elf Druid", b06SpringbloomDruidOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the ability targets nothing, so nothing is picked on announce")
	}
	passPriorityAroundTable(t, g)

	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("a seat with no land is asked nothing (CR 701.17a's 'if you can')")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("'if you do' — nothing was sacrificed, so there is no search")
	}
}

// TestB06SpringbloomDruidSacrificesALandPlayedInResponse is the other
// half of #1026, and the reason the choice moved: at trigger time the
// pick was frozen when the ability went on the stack, so a land that
// arrived in response was not a legal answer. At resolution it is.
func TestB06SpringbloomDruidSacrificesALandPlayedInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Swamp", "Basic Land — Swamp"),
		searchTestLand("Wastes", "Basic Land — Wastes"))
	castAndResolveCreature(t, g, "Springbloom Druid", "Creature — Elf Druid", b06SpringbloomDruidOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)

	// In response: a second land arrives. The old shape had already
	// locked its target in.
	late := seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	passPriorityAroundTable(t, g)

	sac := sacrificeChoiceFor(g, me.ID)
	if sac == nil {
		t.Fatal("the ability should ask which land to sacrifice at resolution")
	}
	if len(sac.SacrificeOptions) != 2 {
		t.Fatalf("both lands are legal answers at resolution, got %v", sac.SacrificeOptions)
	}
	answerSacrifice(t, g, me.ID, late)
	if !g.Battlefield.Contains(first) || !me.Graveyard.Contains(late) {
		t.Error("the land chosen at resolution is the one sacrificed")
	}
	if searchChoiceFor(g, me.ID) == nil {
		t.Error("'if you do' — the sacrifice happened, so the search follows")
	}
}

// --- Bloodchief Ascension ------------------------------------------

func TestB06BloodchiefAscensionQuestsWhenAnOpponentLostTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	asc := pushCatalogPermanent(g, me.ID, "Bloodchief Ascension", "Enchantment", b06BloodchiefAscensionOracle, false)

	// Damage counts ("Damage causes loss of life") — a Bolt is three.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	if latestTriggerPrompt(g, me.ID) == nil {
		t.Fatal("an opponent lost 3 this turn: the end step should ask about the counter")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if counterCount(g, asc, "quest") != 1 {
		t.Errorf("quest counters = %d, want 1", counterCount(g, asc, "quest"))
	}

	// Next turn: one life lost, by a drain, is not two.
	advanceToNextSeatsTurn(t, g)
	advanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("one life is not 'two or more'")
	}
	passPriorityAroundTable(t, g)

	// The turn after: two, spread as one damage and one drain, on the
	// SAME opponent — and "each end step" is any player's.
	advanceToNextSeatsTurn(t, g)
	advanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1)
		_ = g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 1)
	})
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	if latestTriggerPrompt(g, me.ID) == nil {
		t.Fatal("1 + 1 on one opponent is two")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if counterCount(g, asc, "quest") != 2 {
		t.Errorf("quest counters = %d, want 2", counterCount(g, asc, "quest"))
	}
}

func TestB06BloodchiefAscensionMyOwnLossDoesNotCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Bloodchief Ascension", "Enchantment", b06BloodchiefAscensionOracle, false)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -5) })
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("my own life loss is not an opponent's")
	}
}

// b06QuestedAscension pushes a Bloodchief Ascension with three quest
// counters under `controller` — the online state.
func b06QuestedAscension(g *game.Game, controller uuid.UUID, counters int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bloodchief Ascension", TypeLine: "Enchantment",
		OracleID: b06BloodchiefAscensionOracle, Owner: controller, Controller: controller,
		Counters: map[string]int{"quest": counters},
	})
}

func TestB06BloodchiefAscensionDrainsOncePerCardIntoAnOpponentsGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b06QuestedAscension(g, me.ID, 3)
	myLife, oppLife := me.Life, opp.Life

	// A permanent dying — the battlefield-leave path.
	bear := seedCreature(g, "Bear", opp.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if n := b06TriggerPromptsFor(g, me.ID); n != 1 {
		t.Fatalf("a creature dying: %d prompts, want exactly 1", n)
	}
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if opp.Life != oppLife-2 || me.Life != myLife+2 {
		t.Errorf("after the drain: opp %d → %d, me %d → %d; want -2 / +2", oppLife, opp.Life, myLife, me.Life)
	}

	// A discard.
	handCard(opp, "Card", "Sorcery")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	if n := b06TriggerPromptsFor(g, me.ID); n != 1 {
		t.Fatalf("a discard: %d prompts, want exactly 1", n)
	}
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)

	// A mill of two is two cards, two triggers.
	stapleLibraryCard(opp, "Top", "Instant")
	stapleLibraryCard(opp, "Next", "Instant")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 2) })
	if n := b06TriggerPromptsFor(g, me.ID); n != 2 {
		t.Fatalf("milling two: %d prompts, want exactly 2", n)
	}
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if opp.Life != oppLife-8 || me.Life != myLife+8 {
		t.Errorf("four cards in: opp %d → %d, me %d → %d; want -8 / +8", oppLife, opp.Life, myLife, me.Life)
	}

	// Declining loses nothing and gains nothing.
	stapleLibraryCard(opp, "Declined", "Instant")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	b06AnswerAllTriggerPrompts(t, g, me.ID, false)
	if opp.Life != oppLife-8 || me.Life != myLife+8 {
		t.Error("'you may' — declining must change nothing")
	}
}

func TestB06BloodchiefAscensionSeesSpellsResolvingAndBeingCountered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b06QuestedAscension(g, me.ID, 3)
	oppLife := opp.Life

	// An opponent's instant resolving lands in their graveyard.
	batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[2].ID}})
	passPriorityAroundTable(t, g)
	if n := b06TriggerPromptsFor(g, me.ID); n != 1 {
		t.Fatalf("a resolved spell: %d prompts, want exactly 1", n)
	}
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if opp.Life != oppLife-2 {
		t.Errorf("opp %d → %d, want -2", oppLife, opp.Life)
	}

	// A countered one lands there too.
	spell := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[2].ID}})
	castCatalogSpell(t, g, "Counterspell", "Instant", "cc187110-1148-4090-bbb8-e205694a39f5",
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	// Two cards reached graveyards: the Bolt (opponent's — counts)
	// and my Counterspell (mine — does not).
	if n := b06TriggerPromptsFor(g, me.ID); n != 1 {
		t.Fatalf("a countered spell: %d prompts, want exactly 1 (my own spell does not count)", n)
	}
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if opp.Life != oppLife-4 {
		t.Errorf("opp %d → %d, want -4", oppLife, opp.Life)
	}
}

func TestB06BloodchiefAscensionIsOfflineBelowThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b06QuestedAscension(g, me.ID, 2)
	bear := seedCreature(g, "Bear", opp.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if n := b06TriggerPromptsFor(g, me.ID); n != 0 {
		t.Errorf("two counters: %d prompts, want none", n)
	}
	passPriorityAroundTable(t, g)
}

// --- spells --------------------------------------------------------

func TestB06IntoTheFloodMawBouncesAnOpponentsCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := seedCreature(g, "Their Bear", opp.ID)
	mine := seedCreature(g, "My Bear", me.ID)
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")
	legal := legalCards(g, me.ID, b06IntoTheFloodMawOracle)
	if !legal[theirs] || legal[mine] || legal[rock] {
		t.Fatalf("legal targets should be exactly the opponent's creature: %v", legal)
	}
	castCatalogSpell(t, g, "Into the Flood Maw", "Instant", b06IntoTheFloodMawOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the creature should be back in its owner's hand")
	}
}

func TestB06GrimTutorFindsAnyCardThenCostsThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	needle := stapleLibraryCard(me, "Needle", "Sorcery")
	before := me.Life
	castCatalogSpell(t, g, "Grim Tutor", "Sorcery", b06GrimTutorOracle, nil)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("the tutor should ask which card to take")
	}
	if me.Life != before {
		t.Error("the life is lost after the search, not before")
	}
	answerSearchByID(t, g, me.ID, needle)
	if !me.Hand.Contains(needle) {
		t.Error("the chosen card did not reach the hand")
	}
	if me.Life != before-3 {
		t.Errorf("life %d → %d, want -3", before, me.Life)
	}
}

func TestB06TypedTutorsOfferOnlyTheirType(t *testing.T) {
	for _, tc := range []struct{ name, oracle, want, other string }{
		{"Idyllic Tutor", b06IdyllicTutorOracle, "Enchantment", "Creature — Bear"},
		{"Eladamri's Call", b06EladamrisCallOracle, "Creature — Bear", "Enchantment"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		seedSearchLibrary(me,
			game.Card{Name: "Wanted A", TypeLine: tc.want},
			game.Card{Name: "Wanted B", TypeLine: tc.want},
			game.Card{Name: "Wrong", TypeLine: tc.other},
		)
		castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
		passPriorityAroundTable(t, g)
		c := searchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("%s: no search prompt", tc.name)
		}
		if searchOptionNamed(g, c, "Wrong") != uuid.Nil {
			t.Errorf("%s offered a %s", tc.name, tc.other)
		}
		answerSearchNamed(t, g, me.ID, "Wanted A")
		if !b02bHandHasNamed(me, "Wanted A") {
			t.Errorf("%s: the chosen card did not reach the hand", tc.name)
		}
	}
}

func TestB06ReturnToNatureThreeModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")
	castModal(t, g, "Return to Nature", "Instant", b06ReturnToNatureOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("mode 0 destroys the artifact")
	}

	aura := seedPermanentFor(g, opp.ID, "Aura", "Enchantment")
	castModal(t, g, "Return to Nature", "Instant", b06ReturnToNatureOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: aura}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("mode 1 destroys the enchantment")
	}

	// Mode 2 reads any graveyard — the caster's own included.
	buried := pushGraveyardCardForTest(me, "Buried")
	castModal(t, g, "Return to Nature", "Instant", b06ReturnToNatureOracle, []int{2},
		[]game.TargetRef{{Kind: game.TargetCard, ID: buried}})
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(buried) || !exileHas(g, buried) {
		t.Error("mode 2 exiles the card from the graveyard")
	}
}
