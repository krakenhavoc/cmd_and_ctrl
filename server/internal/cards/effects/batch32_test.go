package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch32_test.go — card-level coverage for the card-coverage
// roadmap's batch 32 (#395, `edhrec_rank` 3346–3446): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation or attack rather than by calling primitives.

const (
	b32GroveOfTheBurnwillowsOracle = "d33c3fbb-8306-4c2d-b0dd-88f12639da94"
	b32ReshapeTheEarthOracle       = "edc15ea8-d321-4884-bdcf-ae6198aab78b"
	b32VirulentEmissaryOracle      = "3d4bec90-7bbc-4385-a2b8-303c7a8d0a0f"
	b32SphinxsRevelationOracle     = "71d83fca-e40e-4d0e-956d-d0d6da9cc472"
	b32RockyTarPitOracle           = "8709b5b1-ef9e-45b2-bf4f-ef4c4d613dcd"
	b32BloodPactOracle             = "cfc5a284-7238-4c98-9d2f-7e5c3329be7b"
	b32QuestForRenewalOracle       = "cba4ab80-09e8-4868-a082-a9a3bade9571"
	b32TrailtrackerScoutOracle     = "d38646dc-d1ea-473c-aaf1-df8ba327459f"
	b32CloudshredderSliverOracle   = "f943c005-9b77-411a-b522-1182e22724e1"
	b32AngelicAccordOracle         = "aa95501f-bb59-494a-bcae-b74ca10ad57e"
	b32ErtaiResurrectedOracle      = "3d038f7c-95fa-4b71-8f74-b9b4dd45cde0"
	b32ObyraOracle                 = "4162e60f-9d70-4b8c-a838-df57b83c208d"
	b32TransmutationFontOracle     = "a68e57f3-f138-4c26-b8d5-55d1adec44a8"
	b32NeyaliOracle                = "4012b400-7dcd-43d6-8806-39a3cb743d8f"
	b32ImmaculateMagistrateOracle  = "e6e38bd4-e6dc-400b-8e08-956726842dc4"
	b32AxgardArmoryOracle          = "bce30fd0-ed1e-495d-9149-6a4c81c45c7b"
	b32TrygonPredatorOracle        = "c744b5f4-fbcf-48b8-9d60-5e9c6ac297e0"
	b32CausticCaterpillarOracle    = "45d35128-76e6-43f9-8d23-41f7506c3a71"
	b32HighlandForestOracle        = "35137378-6754-4bb1-a38e-5940890ccab1"
	b32DidntSayPleaseOracle        = "b90bc464-a95a-42e4-9d9a-4b0882eb57ba"
)

// b32Token pushes a creature token under controller, able to attack.
func b32Token(g *game.Game, controller uuid.UUID, name string, power, toughness int) uuid.UUID {
	return pushToken(g, controller, game.Card{
		Name: name, TypeLine: "Token Creature — Soldier", Power: power, Toughness: toughness,
	})
}

// b32PassUntilResolved passes priority until the given spell has
// left the stack, leaving whatever is below it and whatever it
// triggered alone.
func b32PassUntilResolved(t *testing.T, g *game.Game, spell uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && g.Stack.Contains(spell); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Stack.Contains(spell) {
		t.Fatal("the spell never resolved")
	}
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Highland
// Forest is a table row, so a transposed row is invisible until
// someone plays that exact card.
func TestBatch32CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b32GroveOfTheBurnwillowsOracle: "Grove of the Burnwillows",
		b32ReshapeTheEarthOracle:       "Reshape the Earth",
		b32VirulentEmissaryOracle:      "Virulent Emissary",
		b32SphinxsRevelationOracle:     "Sphinx's Revelation",
		b32RockyTarPitOracle:           "Rocky Tar Pit",
		b32BloodPactOracle:             "Blood Pact",
		b32QuestForRenewalOracle:       "Quest for Renewal",
		b32TrailtrackerScoutOracle:     "Trailtracker Scout",
		b32CloudshredderSliverOracle:   "Cloudshredder Sliver",
		b32AngelicAccordOracle:         "Angelic Accord",
		b32ErtaiResurrectedOracle:      "Ertai Resurrected",
		b32ObyraOracle:                 "Obyra, Dreaming Duelist",
		b32TransmutationFontOracle:     "Transmutation Font",
		b32NeyaliOracle:                "Neyali, Suns' Vanguard",
		b32ImmaculateMagistrateOracle:  "Immaculate Magistrate",
		b32AxgardArmoryOracle:          "Axgard Armory",
		b32TrygonPredatorOracle:        "Trygon Predator",
		b32CausticCaterpillarOracle:    "Caustic Caterpillar",
		b32HighlandForestOracle:        "Highland Forest",
		b32DidntSayPleaseOracle:        "Didn't Say Please",
	}
	if len(want) != 20 {
		t.Fatalf("the batch ships 20 cards, the table lists %d", len(want))
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
	}
}

// --- lands ----------------------------------------------------------

func TestB32GroveOfTheBurnwillowsColouredHalfGivesEachOpponentALife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := lifeOfOpponents(g)
	mine := me.Life
	grove := seedPermanentWithOracle(g, me.ID, "Grove of the Burnwillows", "Land", b32GroveOfTheBurnwillowsOracle)

	if err := g.ActivateManaAbility(me.ID, grove, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the {C} half: %v", err)
	}
	if got := lifeOfOpponents(g); got[0] != before[0] || got[1] != before[1] || got[2] != before[2] {
		t.Errorf("the {C} half gives nobody life: %v → %v", before, got)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(grove) })

	if err := g.ActivateManaAbility(me.ID, grove, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the coloured half: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("the coloured half asks R or G, got %+v", pick)
	}
	for i, p := range g.Seats[1:] {
		if p.Life != before[i]+1 {
			t.Errorf("opponent %d gained %d life, want 1", i+1, p.Life-before[i])
		}
	}
	if me.Life != mine {
		t.Errorf("the controller's own life moved: %d → %d", mine, me.Life)
	}
}

func TestB32RockyTarPitEntersTappedAndFetchesASwampOrMountainUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLibraryCardForTest(me, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	crypt := pushLibraryCardForTest(me, game.Card{Name: "Blood Crypt", TypeLine: "Land — Swamp Mountain"})
	pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pit := playLandFromHand(t, g, "Rocky Tar Pit", b32RockyTarPitOracle)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, pit) {
		t.Fatal("Rocky Tar Pit enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(pit) })
	if err := g.ActivateCatalogAbility(me.ID, pit, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Swamp") == uuid.Nil || searchOptionNamed(g, c, "Blood Crypt") == uuid.Nil {
		t.Error("every Swamp or Mountain card is offered, nonbasics included")
	}
	if searchOptionNamed(g, c, "Forest") != uuid.Nil {
		t.Error("a Forest is not offered")
	}
	answerSearchByID(t, g, me.ID, crypt)
	if !g.Battlefield.Contains(crypt) {
		t.Fatal("the chosen land is put onto the battlefield")
	}
	if b16Tapped(t, g, crypt) {
		t.Error("a Mirage fetch puts the land in untapped")
	}
	if g.Battlefield.Contains(pit) {
		t.Error("the Tar Pit is sacrificed")
	}
}

func TestB32HighlandForestEntersTappedAndTapsForRedOrGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := b12PlayFromHand(t, g, "Highland Forest", "Snow Land — Mountain Forest", b32HighlandForestOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, forest) {
		t.Fatal("Highland Forest enters tapped")
	}
	spec, _ := Lookup(b32HighlandForestOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{R|G}" {
		t.Errorf("mana abilities %+v, want one producing {R|G}", spec.ManaAbilities)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(forest) })
	if err := g.ActivateManaAbility(me.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("the land asks R or G, got %+v", pick)
	}
}

func TestB32AxgardArmoryTutorsAnAuraAndAnEquipmentThroughTwoPrompts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	armory := playLandFromHand(t, g, "Axgard Armory", b32AxgardArmoryOracle)
	if !b16Tapped(t, g, armory) {
		t.Fatal("Axgard Armory enters tapped")
	}
	ids := seedSearchLibrary(me,
		game.Card{Name: "Filler", TypeLine: "Creature — Bear"},
		game.Card{Name: "Rancor", TypeLine: "Enchantment — Aura"},
		game.Card{Name: "Ethereal Armor", TypeLine: "Enchantment — Aura"},
		game.Card{Name: "Bonesplitter", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Bear Umbra", TypeLine: "Enchantment — Aura"},
	)
	rancor, bonesplitter := ids[1], ids[3]
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(armory) })
	b06AddMana(me, "C", "R", "R", "W")
	if err := g.ActivateCatalogAbility(me.ID, armory, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(armory) {
		t.Error("the Armory is sacrificed")
	}

	// First prompt: the Auras alone, at most one.
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no Aura prompt")
	}
	if c.SearchMax != 1 || len(c.SearchCards) != 3 || searchOptionNamed(g, c, "Bonesplitter") != uuid.Nil {
		t.Errorf("the first prompt offers the three Auras, one at most: %d offered, max %d", len(c.SearchCards), c.SearchMax)
	}
	answerSearchByID(t, g, me.ID, rancor)
	if !me.Hand.Contains(rancor) {
		t.Fatal("the Aura goes to hand")
	}

	// Second prompt: the Equipment alone — a second Aura cannot be
	// taken.
	c = searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no Equipment prompt")
	}
	if len(c.SearchCards) != 1 || searchOptionNamed(g, c, "Bonesplitter") == uuid.Nil {
		t.Errorf("the second prompt offers the Equipment alone: %d offered", len(c.SearchCards))
	}
	answerSearchByID(t, g, me.ID, bonesplitter)
	if !me.Hand.Contains(bonesplitter) {
		t.Fatal("the Equipment goes to hand")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("two prompts, no more")
	}
	if me.Library.Size() != 3 {
		t.Errorf("library %d, want the three cards left behind", me.Library.Size())
	}
}

// --- spells ---------------------------------------------------------

func TestB32ReshapeTheEarthPutsUpToTenLandsOntoTheBattlefieldTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cards := make([]game.Card, 0, 13)
	for i := 0; i < 12; i++ {
		cards = append(cards, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	}
	cards = append(cards, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	ids := seedSearchLibrary(me, cards...)
	castCatalogSpell(t, g, "Reshape the Earth", "Sorcery", b32ReshapeTheEarthOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMax != 10 || len(c.SearchCards) != 12 {
		t.Errorf("twelve lands offered, up to ten taken: %d offered, max %d", len(c.SearchCards), c.SearchMax)
	}
	answerSearchByID(t, g, me.ID, ids[:10]...)
	for _, id := range ids[:10] {
		if !g.Battlefield.Contains(id) {
			t.Fatal("every chosen land is put onto the battlefield")
		}
		if !b16Tapped(t, g, id) {
			t.Error("the lands enter tapped")
		}
	}
	if me.Library.Size() != 3 {
		t.Errorf("library %d, want 3 (two Forests and the Bear)", me.Library.Size())
	}
}

func TestB32SphinxsRevelationGainsXAndDrawsX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	castXSpell(t, g, "Sphinx's Revelation", "Instant", b32SphinxsRevelationOracle, "{X}{W}{U}{U}", 3, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d → %d, want +3", hand, me.Hand.Size())
	}
}

func TestB32BloodPactTargetPlayerDrawsTwoAndLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, hand := opp.Life, opp.Hand.Size()
	mine := me.Hand.Size()
	castCatalogSpell(t, g, "Blood Pact", "Instant", b32BloodPactOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+2 {
		t.Errorf("the target drew %d, want 2", opp.Hand.Size()-hand)
	}
	if opp.Life != life-2 {
		t.Errorf("the target's life %d → %d, want -2", life, opp.Life)
	}
	if me.Hand.Size() != mine {
		t.Error("the caster drew nothing")
	}
}

func TestB32DidntSayPleaseCountersAndTheControllerMillsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	library, graveyard := opp.Library.Size(), opp.Graveyard.Size()
	castCatalogSpell(t, g, "Didn't Say Please", "Instant", b32DidntSayPleaseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	if opp.Library.Size() != library-3 {
		t.Errorf("the Bolt's controller milled %d, want 3", library-opp.Library.Size())
	}
	if opp.Graveyard.Size() != graveyard+4 {
		t.Errorf("graveyard grew by %d, want 4 (the Bolt and three milled)", opp.Graveyard.Size()-graveyard)
	}
	if opp.AttemptedEmptyDraw {
		t.Error("a mill never loses the game")
	}
}

// --- creatures ------------------------------------------------------

func TestB32VirulentEmissaryGainsALifePerOtherCreatureEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	emissary := castCatalogSpell(t, g, "Virulent Emissary", "Creature — Elf Assassin", b32VirulentEmissaryOracle, nil)
	passPriorityAroundTable(t, g)
	life := me.Life
	if !hasEffectiveKeyword(t, g, emissary, "deathtouch") {
		t.Error("printed deathtouch did not reach the effective abilities")
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("another creature entering: life %d → %d, want +1", life, me.Life)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 2) })
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("two tokens entering: life %d → %d, want +3", life, me.Life)
	}
}

func TestB32ObyraDrainsForEachOtherFaerie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	obyra := castCatalogSpell(t, g, "Obyra, Dreaming Duelist", "Legendary Creature — Faerie Warrior", b32ObyraOracle, nil)
	passPriorityAroundTable(t, g)
	before := lifeOfOpponents(g)
	if !hasEffectiveKeyword(t, g, obyra, "flying") || !hasEffectiveKeyword(t, g, obyra, "flash") {
		t.Error("printed flash and flying did not reach the effective abilities")
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := lifeOfOpponents(g); got[0] != before[0] {
		t.Fatal("a Bear is not a Faerie")
	}
	castCatalogSpell(t, g, "Spellstutter Sprite", "Creature — Faerie Wizard", "", nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != before[i]-1 {
			t.Errorf("opponent %d lost %d, want 1", i+1, before[i]-p.Life)
		}
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, FaerieRogueToken(), 1) })
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != before[i]-2 {
			t.Errorf("a Faerie token counts: opponent %d lost %d, want 2", i+1, before[i]-p.Life)
		}
	}
}

func TestB32CloudshredderSliverGivesYourSliversFlyingAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shredder := b12Push(g, me.ID, "Cloudshredder Sliver", "Creature — Sliver", b32CloudshredderSliverOracle, 1, 1)
	mine := b12Creature(g, me.ID, "My Sliver", "Creature — Sliver", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Sliver", "Creature — Sliver", 2, 2)
	for _, kw := range []string{"flying", "haste"} {
		if !hasEffectiveKeyword(t, g, shredder, kw) {
			t.Errorf("the Cloudshredder grants itself %s", kw)
		}
		if !hasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("another Sliver you control has %s", kw)
		}
		if hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("a non-Sliver has %s", kw)
		}
		if hasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("an opponent's Sliver has %s — the clause says \"you control\"", kw)
		}
	}
}

func TestB32TrailtrackerScoutTapsForAnyColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	scout := pushCatalogPermanent(g, me.ID, "Trailtracker Scout", "Creature — Raccoon Scout", b32TrailtrackerScoutOracle, false)
	if err := g.ActivateManaAbility(me.ID, scout, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("with no commander the pick offers all five colours, got %+v", pick)
	}
	spec, _ := Lookup(b32TrailtrackerScoutOracle)
	if len(spec.Triggered) != 0 || spec.Completeness != CompletenessCaveats {
		t.Error("the expend trigger is a declared omission, not a half-built ability")
	}
}

func TestB32ImmaculateMagistrateCountsEveryElfYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	magistrate := b12Push(g, me.ID, "Immaculate Magistrate", "Creature — Elf Shaman", b32ImmaculateMagistrateOracle, 2, 2)
	b12Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	b12Creature(g, me.ID, "Changeling", "Creature — Shapeshifter", 1, 1)
	b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, magistrate, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)})
	if got := counterCount(g, bear, "+1/+1"); got != 2 {
		t.Errorf("the Magistrate and the Elves are 2 Elves you control: %d counters", got)
	}
	if !b16Tapped(t, g, magistrate) {
		t.Error("the tap is the cost")
	}
}

func TestB32CausticCaterpillarSacrificesToDestroyAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	caterpillar := b12Push(g, me.ID, "Caustic Caterpillar", "Creature — Insect", b32CausticCaterpillarOracle, 1, 1)
	advanceToMain(t, g)
	b06AddMana(me, "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, caterpillar, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	b16Activate(t, g, me.ID, caterpillar, 0, game.ActivateAbilityParams{Targets: b16TargetCard(rock)})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
	if g.Battlefield.Contains(caterpillar) {
		t.Error("the Caterpillar is sacrificed")
	}
}

func TestB32TrygonPredatorOffersOnlyTheHitPlayersArtifactsAndEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me, victim, other := g.Seats[0], g.Seats[1], g.Seats[2]
	predator := b12Push(g, me.ID, "Trygon Predator", "Creature — Beast", b32TrygonPredatorOracle, 2, 3)
	rock := b12Permanent(g, victim.ID, "Their Rock", "Artifact")
	shrine := b12Permanent(g, victim.ID, "Their Shrine", "Enchantment")
	bear := b12Creature(g, victim.ID, "Their Bear", "Creature — Bear", 2, 2)
	otherRock := b12Permanent(g, other.ID, "Other Rock", "Artifact")
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	if !hasEffectiveKeyword(t, g, predator, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}

	attackWith(t, g, victim.ID, predator)
	if victim.Life != 40-2 {
		t.Fatalf("the Predator connected for 2: life %d", victim.Life)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, rock) || !hasID(p.PickTargetCards, shrine) {
		t.Error("the hit player's artifact and enchantment are offered")
	}
	if hasID(p.PickTargetCards, bear) || hasID(p.PickTargetCards, otherRock) || hasID(p.PickTargetCards, mine) {
		t.Error("a creature, another player's artifact and your own are not offered")
	}
	pickCard(t, g, me.ID, shrine)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shrine) {
		t.Error("the chosen enchantment is destroyed")
	}
	if !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(otherRock) {
		t.Error("nothing else is destroyed")
	}
}

func TestB32TrygonPredatorDeclinedDestroysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	predator := b12Push(g, me.ID, "Trygon Predator", "Creature — Beast", b32TrygonPredatorOracle, 2, 3)
	rock := b12Permanent(g, victim.ID, "Their Rock", "Artifact")
	attackWith(t, g, victim.ID, predator)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("declining the trigger destroyed the artifact")
	}
}

// --- the enchantments -------------------------------------------------

func TestB32AngelicAccordMakesAnAngelAtAnyEndStepAfterFourLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Angelic Accord", "Enchantment", b32AngelicAccordOracle, 0, 0)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Angel"); n != 0 {
		t.Errorf("three life is not four: %d Angels", n)
	}
	// An opponent's turn: two gains add up, a loss does not offset
	// them, and "each end step" includes theirs.
	advanceToMainOf(t, g, 1)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -6)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2)
	})
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Angel"); n != 1 {
		t.Fatalf("four life gained on an opponent's turn: %d Angels, want 1", n)
	}
	angel := battlefieldIDNamed(g, me.ID, "Angel")
	if p, tt := effectivePower(t, g, angel), effectiveToughness(t, g, angel); p != 4 || tt != 4 {
		t.Errorf("the Angel is %d/%d, want 4/4", p, tt)
	}
	if !hasEffectiveKeyword(t, g, angel, "flying") {
		t.Error("the Angel flies")
	}
}

func TestB32QuestForRenewalCountsTapsAndUntapsDuringOtherPlayersUntapSteps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, far := g.Seats[0], g.Seats[1], g.Seats[3]
	quest := b12Push(g, me.ID, "Quest for Renewal", "Enchantment", b32QuestForRenewalOracle, 0, 0)
	a := b12Creature(g, me.ID, "Elf A", "Creature — Elf", 1, 1)
	b := b12Creature(g, me.ID, "Elf B", "Creature — Elf", 1, 1)
	vigilant := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Watchful Elf", TypeLine: "Creature — Elf",
		Power: 1, Toughness: 1, Keywords: []string{"vigilance"}, Owner: me.ID, Controller: me.ID,
	})
	// Seat 3's creature: its own untap step is the last of the four,
	// so a tapped one stays tapped through seats 1 and 2's upkeeps.
	theirs := b12Creature(g, far.ID, "Their Bear", "Creature — Bear", 2, 2)

	// A tap outside combat asks; an opponent's tap does not.
	if err := g.TapCard(a, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if err := g.TapCard(theirs, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("an opponent's creature tapping is not \"a creature you control\"")
	}
	if got := counterCount(g, quest, "quest"); got != 1 {
		t.Fatalf("one quest counter, got %d", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(a) })

	// Attacking taps; a vigilance attacker does not become tapped.
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b, vigilant} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	lockInAttacks(t, g)
	prompts := 0
	for latestTriggerPrompt(g, me.ID) != nil {
		prompts++
		answerLatestTriggerPrompt(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
	}
	if prompts != 2 {
		t.Errorf("two attackers became tapped, the vigilant one did not: %d prompts", prompts)
	}
	if got := counterCount(g, quest, "quest"); got != 3 {
		t.Fatalf("three quest counters, got %d", got)
	}

	// Three counters: nothing untaps during the next player's untap
	// step.
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, a) || !b16Tapped(t, g, b) {
		t.Fatal("below four counters the creatures stay tapped")
	}

	// The fourth counter switches the untap on. #74: it is the untap
	// STEP, not the upkeep, and it is a turn-based action rather than
	// a trigger — so by the time the next player's upkeep is reached
	// the creatures are already upright, with nothing announced and
	// nothing to have responded to.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(quest, "quest", 1) })
	advanceToUpkeepOf(t, g, 2)
	if b16Tapped(t, g, a) || b16Tapped(t, g, b) {
		t.Error("with four counters every creature you control untaps during each other player's untap step")
	}
	if triggerOnStack(g, quest) != nil || len(g.PendingTriggers) != 0 {
		t.Error("the untap is a turn-based action — nothing goes on the stack")
	}
	if !b16Tapped(t, g, theirs) {
		t.Error("an opponent's creature is not yours to untap")
	}
}

// --- Ertai Resurrected ---------------------------------------------

func TestB32ErtaiCountersASpellAndItsControllerDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	hand, life := opp.Hand.Size(), me.Life

	ertai := castCatalogSpell(t, g, "Ertai Resurrected", "Legendary Creature — Phyrexian Human Wizard", b32ErtaiResurrectedOracle, nil)
	b32PassUntilResolved(t, g, ertai)
	if !g.Battlefield.Contains(ertai) {
		t.Fatal("Ertai flashes in above the Bolt")
	}
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, bolt) || !hasID(p.PickTargetCards, theirs) || p.PickTargetMin != 0 {
		t.Error("the spell on the stack and the creature are both offered, and declining is allowed")
	}
	if hasID(p.PickTargetCards, ertai) {
		t.Error("\"another\": Ertai is not offered to himself")
	}
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	if me.Life != life {
		t.Error("a countered Bolt deals no damage")
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("the Bolt's controller drew %d, want 1", opp.Hand.Size()-hand)
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the counter mode destroys nothing")
	}
}

func TestB32ErtaiDestroysACreatureAndItsControllerDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := opp.Hand.Size()
	castCatalogSpell(t, g, "Ertai Resurrected", "Legendary Creature — Phyrexian Human Wizard", b32ErtaiResurrectedOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the chosen creature is destroyed")
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("its controller drew %d, want 1", opp.Hand.Size()-hand)
	}
}

func TestB32ErtaiDeclinedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := opp.Hand.Size()
	castCatalogSpell(t, g, "Ertai Resurrected", "Legendary Creature — Phyrexian Human Wizard", b32ErtaiResurrectedOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("picking no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) || opp.Hand.Size() != hand {
		t.Error("\"up to one\": nothing chosen, nothing happens")
	}
}

// --- Transmutation Font ----------------------------------------------

func TestB32TransmutationFontMakesTheChosenToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spec, _ := Lookup(b32TransmutationFontOracle)
	if len(spec.Activated) != 3 || spec.Completeness != CompletenessCaveats {
		t.Fatalf("three token abilities and a declared tutor omission; got %d abilities", len(spec.Activated))
	}
	font := pushCatalogPermanent(g, me.ID, "Transmutation Font", "Artifact", b32TransmutationFontOracle, false)
	advanceToMain(t, g)
	for i, name := range []string{"Blood", "Clue", "Food"} {
		b16Activate(t, g, me.ID, font, i, game.ActivateAbilityParams{})
		if n := countBattlefieldNamed(g, me.ID, name); n != 1 {
			t.Errorf("ability %d makes one %s: %d", i, name, n)
		}
		if err := g.ActivateCatalogAbility(me.ID, font, (i+1)%3, game.ActivateAbilityParams{}); err == nil {
			t.Fatal("the three abilities share the tap — one per untap")
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(font) })
	}
	blood := battlefieldIDNamed(g, me.ID, "Blood")
	if c, _ := battlefieldCard(g, blood); !IsToken(c) || !c.IsArtifact() {
		t.Error("the Blood is an artifact token")
	}
	if len(game.ActivatedAbilitiesForCard(BloodToken())) != 1 {
		t.Error("the Blood is the real token, with its own sacrifice ability")
	}
}

// --- Neyali, Suns' Vanguard ---------------------------------------------

func TestB32NeyaliGivesAttackingTokensDoubleStrikeAndExilesTheTopCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neyali := b12Push(g, me.ID, "Neyali, Suns' Vanguard", "Legendary Creature — Human Rebel", b32NeyaliOracle, 3, 3)
	a := b32Token(g, me.ID, "Soldier A", 1, 1)
	b := b32Token(g, me.ID, "Soldier B", 1, 1)
	home := b32Token(g, me.ID, "Stay-at-home", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// seedSearchLibrary lists top-first.
	ids := seedSearchLibrary(me,
		game.Card{Name: "Top Card", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	top, deep := ids[0], ids[1]

	declareAttack(t, g, opp.ID, a, b, bear)
	if n := triggersOnStackFrom(g, neyali) + len(g.PendingTriggers); n != 1 {
		t.Fatalf("\"one or more\": %d triggers for two tokens", n)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !hasEffectiveKeyword(t, g, id, "double strike") {
			t.Error("an attacking token has double strike")
		}
	}
	if hasEffectiveKeyword(t, g, home, "double strike") || hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("a token that stayed home and a nontoken attacker do not")
	}
	if !g.Exile.Contains(top) || !me.Library.Contains(deep) {
		t.Fatal("the top card is exiled")
	}
	perm := exiledPermission(g, top)
	if perm.Player != me.ID || perm.CastOnly || !perm.Active(me.ID, g.Turn.Number) {
		t.Errorf("%+v: the controller may PLAY it this turn", perm)
	}
	advanceTo(t, g, game.StepCombatDamage)
	if opp.Life != 40-2-2-2 {
		t.Errorf("two double-striking 1/1s and a 2/2: life %d, want 34", opp.Life)
	}
}

func TestB32NeyaliRegrantsEarlierExiledCardsOnALaterTokenAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Neyali, Suns' Vanguard", "Legendary Creature — Human Rebel", b32NeyaliOracle, 3, 3)
	token := b32Token(g, me.ID, "Soldier", 1, 1)
	// seedSearchLibrary lists top-first; the draw between the two
	// attacks comes off the filler below the first card.
	ids := seedSearchLibrary(me,
		game.Card{Name: "First", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 1", TypeLine: "Sorcery"},
		game.Card{Name: "Second", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 2", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 3", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 4", TypeLine: "Sorcery"},
	)
	first, second := ids[0], ids[2]
	// Something else exiled with a grant of its own — never Neyali's
	// to re-grant.
	other := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: other, Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		_, _ = g.ExileTopWithPermissionForEffect(opp.ID, me.ID, 1, game.CastPermission{})
	})
	if !g.Exile.Contains(other) {
		t.Fatal("fixture: the other card is in exile")
	}

	declareAttack(t, g, opp.ID, token)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(first) {
		t.Fatal("the first attack exiles the top card")
	}
	// The grant lapses with the turn.
	advanceToMainOf(t, g, 1)
	if exiledPermission(g, first).Active(me.ID, g.Turn.Number) {
		t.Fatal("the grant ends with the controller's turn")
	}
	// The next token attack exiles another card AND re-grants the
	// first one for this turn; the other card is left alone.
	advanceToMainOf(t, g, 0)
	declareAttack(t, g, opp.ID, token)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(second) {
		t.Fatal("the second attack exiles the next card")
	}
	for _, id := range []uuid.UUID{first, second} {
		if !exiledPermission(g, id).Active(me.ID, g.Turn.Number) {
			t.Errorf("a card Neyali exiled may be played on a turn a token attacked")
		}
	}
	if exiledPermission(g, other).Active(me.ID, g.Turn.Number) {
		t.Error("a card something else exiled is not re-granted")
	}
}

func TestB32NeyaliTokensAttackingOnlyAPlaneswalkerExileNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Neyali, Suns' Vanguard", "Legendary Creature — Human Rebel", b32NeyaliOracle, 3, 3)
	token := b32Token(g, me.ID, "Soldier", 1, 1)
	walker := pushWalkerForTest(g, opp.ID, "Their Walker", "", 3)
	library := me.Library.Size()
	declareAttack(t, g, walker, token)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, token, "double strike") {
		t.Error("the double strike does not care what the token attacks")
	}
	if me.Library.Size() != library {
		t.Error("\"attack a player\": a planeswalker attack exiles nothing")
	}
}
