package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
		effectstest.B32GroveOfTheBurnwillowsOracle: "Grove of the Burnwillows",
		effectstest.B32ReshapeTheEarthOracle:       "Reshape the Earth",
		effectstest.B32VirulentEmissaryOracle:      "Virulent Emissary",
		effectstest.B32SphinxsRevelationOracle:     "Sphinx's Revelation",
		effectstest.B32RockyTarPitOracle:           "Rocky Tar Pit",
		effectstest.B32BloodPactOracle:             "Blood Pact",
		effectstest.B32QuestForRenewalOracle:       "Quest for Renewal",
		effectstest.B32TrailtrackerScoutOracle:     "Trailtracker Scout",
		effectstest.B32CloudshredderSliverOracle:   "Cloudshredder Sliver",
		effectstest.B32AngelicAccordOracle:         "Angelic Accord",
		effectstest.B32ErtaiResurrectedOracle:      "Ertai Resurrected",
		effectstest.B32ObyraOracle:                 "Obyra, Dreaming Duelist",
		effectstest.B32TransmutationFontOracle:     "Transmutation Font",
		effectstest.B32NeyaliOracle:                "Neyali, Suns' Vanguard",
		effectstest.B32ImmaculateMagistrateOracle:  "Immaculate Magistrate",
		effectstest.B32AxgardArmoryOracle:          "Axgard Armory",
		effectstest.B32TrygonPredatorOracle:        "Trygon Predator",
		effectstest.B32CausticCaterpillarOracle:    "Caustic Caterpillar",
		effectstest.B32HighlandForestOracle:        "Highland Forest",
		effectstest.B32DidntSayPleaseOracle:        "Didn't Say Please",
	}
	if len(want) != 20 {
		t.Fatalf("the batch ships 20 cards, the table lists %d", len(want))
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
	}
}

// --- lands ----------------------------------------------------------

func TestB32GroveOfTheBurnwillowsColouredHalfGivesEachOpponentALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := effectstest.LifeOfOpponents(g)
	mine := me.Life
	grove := effectstest.SeedPermanentWithOracle(g, me.ID, "Grove of the Burnwillows", "Land", effectstest.B32GroveOfTheBurnwillowsOracle)

	if err := g.ActivateManaAbility(me.ID, grove, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the {C} half: %v", err)
	}
	if got := effectstest.LifeOfOpponents(g); got[0] != before[0] || got[1] != before[1] || got[2] != before[2] {
		t.Errorf("the {C} half gives nobody life: %v → %v", before, got)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(grove) })

	if err := g.ActivateManaAbility(me.ID, grove, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the coloured half: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	crypt := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Blood Crypt", TypeLine: "Land — Swamp Mountain"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pit := effectstest.PlayLandFromHand(t, g, "Rocky Tar Pit", effectstest.B32RockyTarPitOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, pit) {
		t.Fatal("Rocky Tar Pit enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(pit) })
	if err := g.ActivateCatalogAbility(me.ID, pit, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Swamp") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Blood Crypt") == uuid.Nil {
		t.Error("every Swamp or Mountain card is offered, nonbasics included")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest") != uuid.Nil {
		t.Error("a Forest is not offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, crypt)
	if !g.Battlefield.Contains(crypt) {
		t.Fatal("the chosen land is put onto the battlefield")
	}
	if effectstest.B16Tapped(t, g, crypt) {
		t.Error("a Mirage fetch puts the land in untapped")
	}
	if g.Battlefield.Contains(pit) {
		t.Error("the Tar Pit is sacrificed")
	}
}

func TestB32HighlandForestEntersTappedAndTapsForRedOrGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.B12PlayFromHand(t, g, "Highland Forest", "Snow Land — Mountain Forest", effectstest.B32HighlandForestOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, forest) {
		t.Fatal("Highland Forest enters tapped")
	}
	spec, _ := effects.Lookup(effectstest.B32HighlandForestOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{R|G}" {
		t.Errorf("mana abilities %+v, want one producing {R|G}", spec.ManaAbilities)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(forest) })
	if err := g.ActivateManaAbility(me.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("the land asks R or G, got %+v", pick)
	}
}

func TestB32AxgardArmoryTutorsAnAuraAndAnEquipmentThroughTwoPrompts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	armory := effectstest.PlayLandFromHand(t, g, "Axgard Armory", effectstest.B32AxgardArmoryOracle)
	if !effectstest.B16Tapped(t, g, armory) {
		t.Fatal("Axgard Armory enters tapped")
	}
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Filler", TypeLine: "Creature — Bear"},
		game.Card{Name: "Rancor", TypeLine: "Enchantment — Aura"},
		game.Card{Name: "Ethereal Armor", TypeLine: "Enchantment — Aura"},
		game.Card{Name: "Bonesplitter", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Bear Umbra", TypeLine: "Enchantment — Aura"},
	)
	rancor, bonesplitter := ids[1], ids[3]
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(armory) })
	effectstest.B06AddMana(me, "C", "R", "R", "W")
	if err := g.ActivateCatalogAbility(me.ID, armory, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(armory) {
		t.Error("the Armory is sacrificed")
	}

	// First prompt: the Auras alone, at most one.
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no Aura prompt")
	}
	if c.SearchMax != 1 || len(c.SearchCards) != 3 || effectstest.SearchOptionNamed(g, c, "Bonesplitter") != uuid.Nil {
		t.Errorf("the first prompt offers the three Auras, one at most: %d offered, max %d", len(c.SearchCards), c.SearchMax)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, rancor)
	if !me.Hand.Contains(rancor) {
		t.Fatal("the Aura goes to hand")
	}

	// Second prompt: the Equipment alone — a second Aura cannot be
	// taken.
	c = effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no Equipment prompt")
	}
	if len(c.SearchCards) != 1 || effectstest.SearchOptionNamed(g, c, "Bonesplitter") == uuid.Nil {
		t.Errorf("the second prompt offers the Equipment alone: %d offered", len(c.SearchCards))
	}
	effectstest.AnswerSearchByID(t, g, me.ID, bonesplitter)
	if !me.Hand.Contains(bonesplitter) {
		t.Fatal("the Equipment goes to hand")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("two prompts, no more")
	}
	if me.Library.Size() != 3 {
		t.Errorf("library %d, want the three cards left behind", me.Library.Size())
	}
}

// --- spells ---------------------------------------------------------

func TestB32ReshapeTheEarthPutsUpToTenLandsOntoTheBattlefieldTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cards := make([]game.Card, 0, 13)
	for i := 0; i < 12; i++ {
		cards = append(cards, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	}
	cards = append(cards, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	ids := effectstest.SeedSearchLibrary(me, cards...)
	effectstest.CastCatalogSpell(t, g, "Reshape the Earth", "Sorcery", effectstest.B32ReshapeTheEarthOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMax != 10 || len(c.SearchCards) != 12 {
		t.Errorf("twelve lands offered, up to ten taken: %d offered, max %d", len(c.SearchCards), c.SearchMax)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[:10]...)
	for _, id := range ids[:10] {
		if !g.Battlefield.Contains(id) {
			t.Fatal("every chosen land is put onto the battlefield")
		}
		if !effectstest.B16Tapped(t, g, id) {
			t.Error("the lands enter tapped")
		}
	}
	if me.Library.Size() != 3 {
		t.Errorf("library %d, want 3 (two Forests and the Bear)", me.Library.Size())
	}
}

func TestB32SphinxsRevelationGainsXAndDrawsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastXSpell(t, g, "Sphinx's Revelation", "Instant", effectstest.B32SphinxsRevelationOracle, "{X}{W}{U}{U}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d → %d, want +3", hand, me.Hand.Size())
	}
}

func TestB32BloodPactTargetPlayerDrawsTwoAndLosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, hand := opp.Life, opp.Hand.Size()
	mine := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Blood Pact", "Instant", effectstest.B32BloodPactOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	library, graveyard := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Didn't Say Please", "Instant", effectstest.B32DidntSayPleaseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	emissary := effectstest.CastCatalogSpell(t, g, "Virulent Emissary", "Creature — Elf Assassin", effectstest.B32VirulentEmissaryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	if !effectstest.HasEffectiveKeyword(t, g, emissary, "deathtouch") {
		t.Error("printed deathtouch did not reach the effective abilities")
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("another creature entering: life %d → %d, want +1", life, me.Life)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("two tokens entering: life %d → %d, want +3", life, me.Life)
	}
}

func TestB32ObyraDrainsForEachOtherFaerie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	obyra := effectstest.CastCatalogSpell(t, g, "Obyra, Dreaming Duelist", "Legendary Creature — Faerie Warrior", effectstest.B32ObyraOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	before := effectstest.LifeOfOpponents(g)
	if !effectstest.HasEffectiveKeyword(t, g, obyra, "flying") || !effectstest.HasEffectiveKeyword(t, g, obyra, "flash") {
		t.Error("printed flash and flying did not reach the effective abilities")
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOfOpponents(g); got[0] != before[0] {
		t.Fatal("a Bear is not a Faerie")
	}
	effectstest.CastCatalogSpell(t, g, "Spellstutter Sprite", "Creature — Faerie Wizard", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != before[i]-1 {
			t.Errorf("opponent %d lost %d, want 1", i+1, before[i]-p.Life)
		}
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.FaerieRogueToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != before[i]-2 {
			t.Errorf("a Faerie token counts: opponent %d lost %d, want 2", i+1, before[i]-p.Life)
		}
	}
}

func TestB32CloudshredderSliverGivesYourSliversFlyingAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shredder := effectstest.B12Push(g, me.ID, "Cloudshredder Sliver", "Creature — Sliver", effectstest.B32CloudshredderSliverOracle, 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "My Sliver", "Creature — Sliver", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Sliver", "Creature — Sliver", 2, 2)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, shredder, kw) {
			t.Errorf("the Cloudshredder grants itself %s", kw)
		}
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("another Sliver you control has %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("a non-Sliver has %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("an opponent's Sliver has %s — the clause says \"you control\"", kw)
		}
	}
}

func TestB32TrailtrackerScoutTapsForAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scout := effectstest.PushCatalogPermanent(g, me.ID, "Trailtracker Scout", "Creature — Raccoon Scout", effectstest.B32TrailtrackerScoutOracle, false)
	if err := g.ActivateManaAbility(me.ID, scout, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("with no commander the pick offers all five colours, got %+v", pick)
	}
	spec, _ := effects.Lookup(effectstest.B32TrailtrackerScoutOracle)
	// #3172: the expend 8 trigger landed; cav_b_expend_test.go drives it.
	if len(spec.Triggered) != 1 || spec.Completeness != effects.CompletenessFull {
		t.Error("Trailtracker Scout declares its expend 8 trigger and no caveat")
	}
}

func TestB32ImmaculateMagistrateCountsEveryElfYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	magistrate := effectstest.B12Push(g, me.ID, "Immaculate Magistrate", "Creature — Elf Shaman", effectstest.B32ImmaculateMagistrateOracle, 2, 2)
	effectstest.B12Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.B12Creature(g, me.ID, "Changeling", "Creature — Shapeshifter", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, magistrate, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 2 {
		t.Errorf("the Magistrate and the Elves are 2 Elves you control: %d counters", got)
	}
	if !effectstest.B16Tapped(t, g, magistrate) {
		t.Error("the tap is the cost")
	}
}

func TestB32CausticCaterpillarSacrificesToDestroyAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	caterpillar := effectstest.B12Push(g, me.ID, "Caustic Caterpillar", "Creature — Insect", effectstest.B32CausticCaterpillarOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, caterpillar, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	effectstest.B16Activate(t, g, me.ID, caterpillar, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(rock)})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
	if g.Battlefield.Contains(caterpillar) {
		t.Error("the Caterpillar is sacrificed")
	}
}

func TestB32TrygonPredatorOffersOnlyTheHitPlayersArtifactsAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim, other := g.Seats[0], g.Seats[1], g.Seats[2]
	predator := effectstest.B12Push(g, me.ID, "Trygon Predator", "Creature — Beast", effectstest.B32TrygonPredatorOracle, 2, 3)
	rock := effectstest.B12Permanent(g, victim.ID, "Their Rock", "Artifact")
	shrine := effectstest.B12Permanent(g, victim.ID, "Their Shrine", "Enchantment")
	bear := effectstest.B12Creature(g, victim.ID, "Their Bear", "Creature — Bear", 2, 2)
	otherRock := effectstest.B12Permanent(g, other.ID, "Other Rock", "Artifact")
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	if !effectstest.HasEffectiveKeyword(t, g, predator, "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}

	effectstest.AttackWith(t, g, victim.ID, predator)
	if victim.Life != 40-2 {
		t.Fatalf("the Predator connected for 2: life %d", victim.Life)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, rock) || !effectstest.HasID(p.PickTargetCards, shrine) {
		t.Error("the hit player's artifact and enchantment are offered")
	}
	if effectstest.HasID(p.PickTargetCards, bear) || effectstest.HasID(p.PickTargetCards, otherRock) || effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("a creature, another player's artifact and your own are not offered")
	}
	effectstest.PickCard(t, g, me.ID, shrine)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shrine) {
		t.Error("the chosen enchantment is destroyed")
	}
	if !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(otherRock) {
		t.Error("nothing else is destroyed")
	}
}

func TestB32TrygonPredatorDeclinedDestroysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	predator := effectstest.B12Push(g, me.ID, "Trygon Predator", "Creature — Beast", effectstest.B32TrygonPredatorOracle, 2, 3)
	rock := effectstest.B12Permanent(g, victim.ID, "Their Rock", "Artifact")
	effectstest.AttackWith(t, g, victim.ID, predator)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("declining the trigger destroyed the artifact")
	}
}

// --- the enchantments -------------------------------------------------

func TestB32AngelicAccordMakesAnAngelAtAnyEndStepAfterFourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Angelic Accord", "Enchantment", effectstest.B32AngelicAccordOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Angel"); n != 0 {
		t.Errorf("three life is not four: %d Angels", n)
	}
	// An opponent's turn: two gains add up, a loss does not offset
	// them, and "each end step" includes theirs.
	effectstest.AdvanceToMainOf(t, g, 1)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -6)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2)
	})
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Angel"); n != 1 {
		t.Fatalf("four life gained on an opponent's turn: %d Angels, want 1", n)
	}
	angel := effectstest.BattlefieldIDNamed(g, me.ID, "Angel")
	if p, tt := effectstest.EffectivePower(t, g, angel), effectstest.EffectiveToughness(t, g, angel); p != 4 || tt != 4 {
		t.Errorf("the Angel is %d/%d, want 4/4", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, angel, "flying") {
		t.Error("the Angel flies")
	}
}

func TestB32QuestForRenewalCountsTapsAndUntapsDuringOtherPlayersUntapSteps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, far := g.Seats[0], g.Seats[1], g.Seats[3]
	quest := effectstest.B12Push(g, me.ID, "Quest for Renewal", "Enchantment", effectstest.B32QuestForRenewalOracle, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "Elf A", "Creature — Elf", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Elf B", "Creature — Elf", 1, 1)
	vigilant := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Watchful Elf", TypeLine: "Creature — Elf",
		Power: 1, Toughness: 1, Keywords: []string{"vigilance"}, Owner: me.ID, Controller: me.ID,
	})
	// Seat 3's creature: its own untap step is the last of the four,
	// so a tapped one stays tapped through seats 1 and 2's upkeeps.
	theirs := effectstest.B12Creature(g, far.ID, "Their Bear", "Creature — Bear", 2, 2)

	// A tap outside combat asks; an opponent's tap does not.
	if err := g.TapCard(a, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.TapCard(theirs, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("an opponent's creature tapping is not \"a creature you control\"")
	}
	if got := effectstest.CounterCount(g, quest, "quest"); got != 1 {
		t.Fatalf("one quest counter, got %d", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(a) })

	// Attacking taps; a vigilance attacker does not become tapped.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b, vigilant} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	effectstest.LockInAttacks(t, g)
	prompts := 0
	for effectstest.OwedMays(g, me.ID) > 0 {
		prompts++
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if prompts != 2 {
		t.Errorf("two attackers became tapped, the vigilant one did not: %d prompts", prompts)
	}
	if got := effectstest.CounterCount(g, quest, "quest"); got != 3 {
		t.Fatalf("three quest counters, got %d", got)
	}

	// Three counters: nothing untaps during the next player's untap
	// step.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, a) || !effectstest.B16Tapped(t, g, b) {
		t.Fatal("below four counters the creatures stay tapped")
	}

	// The fourth counter switches the untap on. #74: it is the untap
	// STEP, not the upkeep, and it is a turn-based action rather than
	// a trigger — so by the time the next player's upkeep is reached
	// the creatures are already upright, with nothing announced and
	// nothing to have responded to.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(quest, "quest", 1) })
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	if effectstest.B16Tapped(t, g, a) || effectstest.B16Tapped(t, g, b) {
		t.Error("with four counters every creature you control untaps during each other player's untap step")
	}
	if effectstest.TriggerOnStack(g, quest) != nil || len(g.PendingTriggers) != 0 {
		t.Error("the untap is a turn-based action — nothing goes on the stack")
	}
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("an opponent's creature is not yours to untap")
	}
}

// --- Ertai Resurrected ---------------------------------------------

func TestB32ErtaiCountersASpellAndItsControllerDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	hand, life := opp.Hand.Size(), me.Life

	ertai := effectstest.CastCatalogSpell(t, g, "Ertai Resurrected", "Legendary Creature — Phyrexian Human Wizard", effectstest.B32ErtaiResurrectedOracle, nil)
	b32PassUntilResolved(t, g, ertai)
	if !g.Battlefield.Contains(ertai) {
		t.Fatal("Ertai flashes in above the Bolt")
	}
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil || c.ModeMin != 0 {
		t.Fatalf("a real 'choose up to one' prompt, declinable: %+v", c)
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, bolt) || effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("the counter mode offers the spell and not the creature")
	}
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
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

// The ability half of the old caveat: an opponent's triggered ability
// is a legal pick, goes away, and its controller draws.
func TestB32ErtaiCountersAnAbilityAndItsControllerDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	fired := false
	ability := effectstest.TaAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, &fired)
	hand := opp.Hand.Size()

	ertai := effectstest.CastCatalogSpell(t, g, "Ertai Resurrected", "Legendary Creature — Phyrexian Human Wizard", effectstest.B32ErtaiResurrectedOracle, nil)
	b32PassUntilResolved(t, g, ertai)
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, me.ID), me.ID, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); !effectstest.HasID(p.PickTargetCards, ability) {
		t.Fatal("the opponent's ability is offered")
	}
	effectstest.PickCard(t, g, me.ID, ability)
	effectstest.PassPriorityAroundTable(t, g)
	if fired {
		t.Error("the countered ability resolved")
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("its controller drew %d, want 1", opp.Hand.Size()-hand)
	}
	if !g.Battlefield.Contains(src) {
		t.Error("CR 701.6a: the source stays")
	}
	if spec, _ := effects.Lookup(effectstest.B32ErtaiResurrectedOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("both caveats are gone")
	}
}

func TestB32ErtaiDestroysACreatureAndItsControllerDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := opp.Hand.Size()
	ertai := effectstest.CastCatalogSpell(t, g, "Ertai Resurrected", "Legendary Creature — Phyrexian Human Wizard", effectstest.B32ErtaiResurrectedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, me.ID), me.ID, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); effectstest.HasID(p.PickTargetCards, ertai) {
		t.Error("\"another\": Ertai is not offered to himself")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the chosen creature is destroyed")
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("its controller drew %d, want 1", opp.Hand.Size()-hand)
	}
}

func TestB32ErtaiDeclinedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Ertai Resurrected", "Legendary Creature — Phyrexian Human Wizard", effectstest.B32ErtaiResurrectedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(c.ID, me.ID, nil); err != nil {
		t.Fatalf("choosing no mode: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) || opp.Hand.Size() != hand {
		t.Error("\"up to one\": nothing chosen, nothing happens")
	}
}

// --- Transmutation Font ----------------------------------------------

func TestB32TransmutationFontMakesTheChosenToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spec, _ := effects.Lookup(effectstest.B32TransmutationFontOracle)
	if len(spec.Activated) != 4 || spec.Completeness != effects.CompletenessFull {
		t.Fatalf("three token abilities and the tutor, Full; got %d abilities", len(spec.Activated))
	}
	font := effectstest.PushCatalogPermanent(g, me.ID, "Transmutation Font", "Artifact", effectstest.B32TransmutationFontOracle, false)
	effectstest.AdvanceToMain(t, g)
	for i, name := range []string{"Blood", "Clue", "Food"} {
		effectstest.B16Activate(t, g, me.ID, font, i, game.ActivateAbilityParams{})
		if n := effectstest.CountBattlefieldNamed(g, me.ID, name); n != 1 {
			t.Errorf("ability %d makes one %s: %d", i, name, n)
		}
		if err := g.ActivateCatalogAbility(me.ID, font, (i+1)%3, game.ActivateAbilityParams{}); err == nil {
			t.Fatal("the three abilities share the tap — one per untap")
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(font) })
	}
	blood := effectstest.BattlefieldIDNamed(g, me.ID, "Blood")
	if c, _ := effectstest.BattlefieldCard(g, blood); !effects.IsToken(c) || !c.IsArtifact() {
		t.Error("the Blood is an artifact token")
	}
	if len(game.ActivatedAbilitiesForCard(effects.BloodToken())) != 1 {
		t.Error("the Blood is the real token, with its own sacrifice ability")
	}
}

// --- Neyali, Suns' Vanguard ---------------------------------------------

func TestB32NeyaliGivesAttackingTokensDoubleStrikeAndExilesTheTopCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neyali := effectstest.B12Push(g, me.ID, "Neyali, Suns' Vanguard", "Legendary Creature — Human Rebel", effectstest.B32NeyaliOracle, 3, 3)
	a := effectstest.B32Token(g, me.ID, "Soldier A", 1, 1)
	b := effectstest.B32Token(g, me.ID, "Soldier B", 1, 1)
	home := effectstest.B32Token(g, me.ID, "Stay-at-home", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// seedSearchLibrary lists top-first.
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Top Card", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	top, deep := ids[0], ids[1]

	effectstest.DeclareAttack(t, g, opp.ID, a, b, bear)
	if n := effectstest.TriggersOnStackFrom(g, neyali) + len(g.PendingTriggers); n != 1 {
		t.Fatalf("\"one or more\": %d triggers for two tokens", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "double strike") {
			t.Error("an attacking token has double strike")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, home, "double strike") || effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("a token that stayed home and a nontoken attacker do not")
	}
	if !g.Exile.Contains(top) || !me.Library.Contains(deep) {
		t.Fatal("the top card is exiled")
	}
	perm := effectstest.ExiledPermission(g, top)
	if perm.Player != me.ID || perm.CastOnly || !effectstest.PermissionLive(g, perm, me.ID) {
		t.Errorf("%+v: the controller may PLAY it this turn", perm)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if opp.Life != 40-2-2-2 {
		t.Errorf("two double-striking 1/1s and a 2/2: life %d, want 34", opp.Life)
	}
}

// "During any turn you attacked with a token, you may play that card"
// (#3192): the exiled card is playable only during a turn its holder
// attacked with a token, on a LATER such turn too, and it does not need
// Neyali, or her trigger, to still be there.
func TestB32NeyaliCardIsPlayableOnlyDuringTurnsATokenAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neyali := effectstest.B12Push(g, me.ID, "Neyali, Suns' Vanguard", "Legendary Creature — Human Rebel", effectstest.B32NeyaliOracle, 3, 3)
	token := effectstest.B32Token(g, me.ID, "Soldier", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "First", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 1", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 2", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 3", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 4", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 5", TypeLine: "Sorcery"},
	)
	first := ids[0]
	// Something else exiled with a grant of its own is never Neyali's.
	other := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: other, Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		_, _ = g.ExileTopWithPermissionForEffect(opp.ID, me.ID, 1, game.CastPermission{})
	})

	// Turn 1: a token attacks a player; the top card is exiled and is
	// playable now.
	effectstest.DeclareAttack(t, g, opp.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(first) {
		t.Fatal("the first attack exiles the top card")
	}
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, first), me.ID) {
		t.Fatal("playable the turn a token attacked")
	}

	// Turn 2 (mine again): nothing has attacked yet, so the card is
	// DORMANT: still exiled, still granted, not playable.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	perm := effectstest.ExiledPermission(g, first)
	if perm.Player != me.ID || perm.Condition == nil {
		t.Fatalf("the grant did not persist across the turn: %+v", perm)
	}
	if effectstest.PermissionLive(g, perm, me.ID) {
		t.Fatal("playable on a turn before any token attacked")
	}
	// A nontoken attacker does not wake it.
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, first), me.ID) {
		t.Fatal("playable on a turn only a nontoken attacked")
	}

	// Turn 3: Neyali is gone, so no trigger resolves, but a token
	// attacking is enough: the card is playable on this LATER turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(neyali) })
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, first), me.ID) {
		t.Fatal("playable before the token attacked")
	}
	effectstest.DeclareAttack(t, g, opp.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, first), me.ID) {
		t.Error("a token attacked on a later turn and the card is still shut")
	}
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, other), me.ID) {
		t.Error("a card something else exiled became playable")
	}
}

func TestB32NeyaliTokensAttackingOnlyAPlaneswalkerExileNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Neyali, Suns' Vanguard", "Legendary Creature — Human Rebel", effectstest.B32NeyaliOracle, 3, 3)
	token := effectstest.B32Token(g, me.ID, "Soldier", 1, 1)
	walker := effectstest.PushWalkerForTest(g, opp.ID, "Their Walker", "", 3)
	library := me.Library.Size()
	effectstest.DeclareAttack(t, g, walker, token)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, token, "double strike") {
		t.Error("the double strike does not care what the token attacks")
	}
	if me.Library.Size() != library {
		t.Error("\"attack a player\": a planeswalker attack exiles nothing")
	}
}
