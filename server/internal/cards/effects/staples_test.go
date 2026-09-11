package effects

// staples_test.go covers the Commander-staples pass: the most-played
// removal, ramp and mana rocks. Each test asserts the behaviour that
// makes the card worth playing rather than merely that it resolved —
// Nature's Lore's land arrives UNTAPPED, Beast Within's token goes to
// the VICTIM, Anguished Unmaking's life loss happens even on a fizzle.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	beastWithinOracle       = "7735eeba-693b-47e2-bd51-414379cf1016"
	generousGiftOracle      = "fae37e28-e137-4177-b973-fa8b4dd8f409"
	krosanGripOracle        = "3e39224c-72ce-4ecc-aa17-12c071ea1f3e"
	naturesClaimOracle      = "6d4e558e-9109-4918-a082-fdcbaffd516b"
	anguishedUnmakingOracle = "ad09b3c3-c8e7-481c-8c45-e7f234935117"
	blasphemousActOracle    = "7a2484a9-04fd-41a0-8224-610c1c07ed10"
	rampantGrowthOracle     = "8539f295-5d58-4436-a73a-b9277c4c7795"
	farseekOracle           = "495e52e6-4c2b-4574-9474-eadbdcc8b4ac"
	naturesLoreOracle       = "78826359-fe63-44ad-adc4-a17ffcd710e4"
	kodamasReachOracle      = "1593ea18-2f2f-4ab4-83fb-6ccc0bec8a90"
	woodElvesOracle         = "8973bd99-20f8-4867-90ef-50392147ee1b"
	thranDynamoOracle       = "a699c663-8131-4045-9265-a83e86609374"
	wornPowerstoneOracle    = "b166b670-febc-4821-855e-f8d465644c03"
	commandTowerOracle      = "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"
	mindStoneOracle         = "c97361b5-af16-4a7b-af85-a429dbaf4ad2"
	nightsWhisperOracle     = "7ffae8f8-3006-4969-a339-6d30678f87ea"
)

// seedPermanentFor puts a permanent on the battlefield under owner's
// control and returns its instance ID.
func seedPermanentFor(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// countBattlefieldNamed counts battlefield cards named `name` under
// controller. Tokens have no stable ID to assert against, so the
// token-producing cards check by name + controller instead.
func countBattlefieldNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

// battlefieldCard finds a battlefield card by instance ID.
func battlefieldCard(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// --- Beast Within / Generous Gift --------------------------------

// The token goes to the permanent's controller, not the caster —
// that asymmetry is the entire design of the card.
func TestBeastWithinDestroysAndGivesVictimTheToken(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	landID := seedPermanentFor(g, victim.ID, "Ancient Tomb", "Land")

	castCatalogSpell(t, g, "Beast Within", "Instant", beastWithinOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: landID}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(landID) {
		t.Error("target permanent survived")
	}
	if !victim.Graveyard.Contains(landID) {
		t.Error("destroyed permanent did not reach its owner's graveyard")
	}
	if got := countBattlefieldNamed(g, victim.ID, "Beast"); got != 1 {
		t.Errorf("victim has %d Beast tokens, want 1", got)
	}
	if got := countBattlefieldNamed(g, caster.ID, "Beast"); got != 0 {
		t.Errorf("caster got %d Beast tokens, want 0 — the token is the drawback", got)
	}
}

// Beast Within can hit anything, including the caster's own board —
// worth pinning because an unrestricted TargetPermanent is easy to
// accidentally narrow later.
func TestBeastWithinCanTargetYourOwnPermanent(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	mineID := seedPermanentFor(g, caster.ID, "Sol Ring", "Artifact")

	castCatalogSpell(t, g, "Beast Within", "Instant", beastWithinOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mineID}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(mineID) {
		t.Error("own permanent survived")
	}
	if got := countBattlefieldNamed(g, caster.ID, "Beast"); got != 1 {
		t.Errorf("caster has %d Beast tokens, want 1", got)
	}
}

func TestGenerousGiftDestroysAndGivesElephant(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	creatureID := seedPermanentFor(g, victim.ID, "Blightsteel Colossus", "Creature — Golem")

	castCatalogSpell(t, g, "Generous Gift", "Instant", generousGiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: creatureID}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(creatureID) {
		t.Error("target survived")
	}
	if got := countBattlefieldNamed(g, victim.ID, "Elephant"); got != 1 {
		t.Errorf("victim has %d Elephant tokens, want 1", got)
	}
}

// --- Krosan Grip / Nature's Claim --------------------------------

func TestKrosanGripDestroysArtifactAndEnchantment(t *testing.T) {
	for _, tc := range []struct{ name, typeLine string }{
		{"Sol Ring", "Artifact"},
		{"Rhystic Study", "Enchantment"},
	} {
		t.Run(tc.typeLine, func(t *testing.T) {
			g := newCatalogGame(t)
			victim := g.Seats[1]
			id := seedPermanentFor(g, victim.ID, tc.name, tc.typeLine)

			castCatalogSpell(t, g, "Krosan Grip", "Instant", krosanGripOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
			passPriorityAroundTable(t, g)

			if g.Battlefield.Contains(id) {
				t.Errorf("%s survived Krosan Grip", tc.typeLine)
			}
		})
	}
}

// The 4 life goes to the permanent's controller — cracking your own
// artifact gains YOU the life, which is how the card is often used.
func TestNaturesClaimGivesLifeToThePermanentsController(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	before, casterBefore := victim.Life, caster.Life
	id := seedPermanentFor(g, victim.ID, "Smothering Tithe", "Enchantment")

	castCatalogSpell(t, g, "Nature's Claim", "Instant", naturesClaimOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(id) {
		t.Error("enchantment survived")
	}
	if got := victim.Life - before; got != 4 {
		t.Errorf("victim gained %d life, want 4", got)
	}
	if caster.Life != casterBefore {
		t.Errorf("caster life changed by %d, want 0", caster.Life-casterBefore)
	}
}

// --- Anguished Unmaking -----------------------------------------

func TestAnguishedUnmakingExilesAndCostsThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	before := caster.Life
	id := seedPermanentFor(g, victim.ID, "Avacyn, Angel of Hope", "Creature — Angel")

	castCatalogSpell(t, g, "Anguished Unmaking", "Instant", anguishedUnmakingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
	passPriorityAroundTable(t, g)

	// Exile, not destroy — that's why it answers indestructible.
	if !g.Exile.Contains(id) {
		t.Error("target was not exiled")
	}
	if victim.Graveyard.Contains(id) {
		t.Error("target reached the graveyard; it should be exiled")
	}
	if got := before - caster.Life; got != 3 {
		t.Errorf("caster lost %d life, want 3", got)
	}
}

// A spell with one target that has become illegal does not resolve at
// all (CR 608.2b), so the life loss does NOT happen either — the
// second sentence is part of the spell's resolution, not a cost.
// Pinned because it is tempting to read "You lose 3 life" as
// unconditional and implement it outside the target guard.
func TestAnguishedUnmakingFizzlesEntirelyWhenTargetLeaves(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	id := seedPermanentFor(g, victim.ID, "Avacyn, Angel of Hope", "Creature — Angel")

	castCatalogSpell(t, g, "Anguished Unmaking", "Instant", anguishedUnmakingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
	// The target leaves in response, before the spell resolves.
	if _, err := g.Battlefield.Remove(id); err != nil {
		t.Fatalf("remove target: %v", err)
	}
	lifeBefore := caster.Life
	passPriorityAroundTable(t, g)

	if caster.Life != lifeBefore {
		t.Errorf("caster lost %d life on a fizzled spell, want 0", lifeBefore-caster.Life)
	}
}

// --- Blasphemous Act --------------------------------------------

func TestBlasphemousActDestroysEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	mine := seedPermanentFor(g, g.Seats[0].ID, "Llanowar Elves", "Creature — Elf")
	theirs := seedPermanentFor(g, g.Seats[1].ID, "Grizzly Bears", "Creature — Bear")
	// A non-creature must survive — the wipe is creatures only.
	rock := seedPermanentFor(g, g.Seats[1].ID, "Sol Ring", "Artifact")

	castCatalogSpell(t, g, "Blasphemous Act", "Sorcery", blasphemousActOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs} {
		if g.Battlefield.Contains(id) {
			t.Error("a creature survived Blasphemous Act")
		}
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("Blasphemous Act destroyed a noncreature permanent")
	}
}

// --- Ramp -------------------------------------------------------

func TestRampantGrowthFetchesBasicTapped(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	forestID := pushLibraryCardForTest(caster, game.Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
	})

	castCatalogSpell(t, g, "Rampant Growth", "Sorcery", rampantGrowthOracle, nil)
	passPriorityAroundTable(t, g)

	fetched, ok := battlefieldCard(g, forestID)
	if !ok {
		t.Fatal("fetched basic not on the battlefield")
	}
	if !fetched.Tapped {
		t.Error("Rampant Growth's land must enter tapped")
	}
}

// Untapped is the whole reason Nature's Lore is premium over
// Rampant Growth, so this is the assertion that matters.
func TestNaturesLoreFetchesForestUntapped(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	forestID := pushLibraryCardForTest(caster, game.Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
	})

	castCatalogSpell(t, g, "Nature's Lore", "Sorcery", naturesLoreOracle, nil)
	passPriorityAroundTable(t, g)

	fetched, ok := battlefieldCard(g, forestID)
	if !ok {
		t.Fatal("fetched Forest not on the battlefield")
	}
	if fetched.Tapped {
		t.Error("Nature's Lore's land must enter UNTAPPED")
	}
}

// Nature's Lore takes any land with the Forest subtype, not just a
// basic — a dual counts, as in paper.
func TestNaturesLoreAcceptsNonBasicForest(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	bayouID := pushLibraryCardForTest(caster, game.Card{
		Name: "Bayou", TypeLine: "Land — Swamp Forest",
	})

	castCatalogSpell(t, g, "Nature's Lore", "Sorcery", naturesLoreOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bayouID) {
		t.Error("Nature's Lore should fetch a nonbasic Forest")
	}
}

// Farseek must skip a Forest — that restriction is the card.
func TestFarseekSkipsForest(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	// Island sits deeper than the Forest; SearchLibrary scans from the
	// bottom, so pushing the Island first makes it the earlier match
	// and proves the predicate rejects the Forest rather than merely
	// finding the Island first by luck.
	islandID := pushLibraryCardForTest(caster, game.Card{
		Name: "Island", TypeLine: "Basic Land — Island",
	})
	forestID := pushLibraryCardForTest(caster, game.Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
	})

	castCatalogSpell(t, g, "Farseek", "Sorcery", farseekOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(islandID) {
		t.Error("Farseek did not fetch the Island")
	}
	if g.Battlefield.Contains(forestID) {
		t.Error("Farseek fetched a Forest, which it may not")
	}
	fetched, _ := battlefieldCard(g, islandID)
	if !fetched.Tapped {
		t.Error("Farseek's land must enter tapped")
	}
}

func TestKodamasReachSplitsFieldAndHand(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	handBefore := caster.Hand.Size()
	f1 := pushLibraryCardForTest(caster, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pushLibraryCardForTest(caster, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	castCatalogSpell(t, g, "Kodama's Reach", "Sorcery — Arcane", kodamasReachOracle, nil)
	passPriorityAroundTable(t, g)
	// Two basics and one pick, so the battlefield half prompts. The
	// hand half is chained off it and finds one remaining Forest —
	// no decision, so no second prompt.
	answerSearchByID(t, g, caster.ID, f1)

	// One basic onto the battlefield tapped, one into hand. Hand is
	// +1 net: the Kodama's Reach card itself left for the graveyard.
	onField := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Controller == caster.ID {
			onField++
			if !c.Tapped {
				t.Error("battlefield half must enter tapped")
			}
		}
	}
	if onField != 1 {
		t.Errorf("%d Forests on the battlefield, want 1", onField)
	}
	if caster.Hand.Size() != handBefore+1 {
		t.Errorf("hand size %d, want %d (one fetched basic, spell gone)",
			caster.Hand.Size(), handBefore+1)
	}
}

// Wood Elves fetches on ETB, so it works off any battlefield entry,
// not just a cast.
func TestWoodElvesFetchesForestOnETB(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	forestID := pushLibraryCardForTest(caster, game.Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
	})

	castCatalogSpell(t, g, "Wood Elves", "Creature — Elf Scout", woodElvesOracle, nil)
	passPriorityAroundTable(t, g)

	fetched, ok := battlefieldCard(g, forestID)
	if !ok {
		t.Fatal("Wood Elves did not fetch a Forest")
	}
	if fetched.Tapped {
		t.Error("Wood Elves' Forest enters untapped")
	}
}

// --- Draw -------------------------------------------------------

func TestNightsWhisperDrawsTwoAndCostsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	handBefore, lifeBefore := caster.Hand.Size(), caster.Life

	castCatalogSpell(t, g, "Night's Whisper", "Sorcery", nightsWhisperOracle, nil)
	passPriorityAroundTable(t, g)

	// handBefore is sampled before castCatalogSpell seeds the spell
	// into hand, so the card's arrival and departure cancel: the delta
	// is purely the two cards drawn.
	if got := caster.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand delta %d, want 2 cards drawn", got)
	}
	if got := lifeBefore - caster.Life; got != 2 {
		t.Errorf("lost %d life, want 2", got)
	}
}

// --- Mana rocks -------------------------------------------------

func TestManaRockAbilityShapes(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, produced string
	}{
		{"Thran Dynamo", thranDynamoOracle, "{C}{C}{C}"},
		{"Worn Powerstone", wornPowerstoneOracle, "{C}{C}"},
		{"Mind Stone", mindStoneOracle, "{C}"},
		{"Command Tower", commandTowerOracle, "{W|U|B|R|G}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			abilities := game.ManaAbilitiesForCard(game.Card{OracleID: tc.oracle})
			if len(abilities) != 1 {
				t.Fatalf("%d mana abilities, want 1", len(abilities))
			}
			a := abilities[0]
			if !a.TapCost {
				t.Error("cost should be a tap")
			}
			if a.Produced != tc.produced {
				t.Errorf("produced %q, want %q", a.Produced, tc.produced)
			}
		})
	}
}

// Worn Powerstone's enters-tapped drawback is what makes it cost
// three rather than two.
func TestWornPowerstoneEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Worn Powerstone", "Artifact", wornPowerstoneOracle, nil)
	passPriorityAroundTable(t, g)

	stone, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Worn Powerstone not on the battlefield")
	}
	if !stone.Tapped {
		t.Error("Worn Powerstone must enter tapped")
	}
}

// Thran Dynamo has no ETB hook, so it must arrive untapped — the
// mirror of the Powerstone assertion above.
func TestThranDynamoEntersUntapped(t *testing.T) {
	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Thran Dynamo", "Artifact", thranDynamoOracle, nil)
	passPriorityAroundTable(t, g)

	dynamo, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Thran Dynamo not on the battlefield")
	}
	if dynamo.Tapped {
		t.Error("Thran Dynamo should enter untapped")
	}
}

// Mind Stone's sac-for-a-card ability is the first catalog use of a
// three-component cost: mana + tap + sacrifice-self.
func TestMindStoneSacrificeAbilityShape(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: mindStoneOracle})
	if len(abilities) != 1 {
		t.Fatalf("%d activated abilities, want 1", len(abilities))
	}
	cost := abilities[0].Cost
	if !cost.Tap {
		t.Error("cost should include a tap")
	}
	if !cost.SacrificeSelf {
		t.Error("cost should sacrifice the Stone itself")
	}
	if cost.Mana != "{1}" {
		t.Errorf("mana cost %q, want {1}", cost.Mana)
	}
	if abilities[0].Targets != nil {
		t.Error("drawing a card targets nothing")
	}
}
