package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b05TapOnBattlefield taps a permanent directly, for tests that need
// something to untap.
func b05TapOnBattlefield(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}

// b05CountTokensNamed counts tokens with the given name under a
// controller.
func b05CountTokensNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller && effects.IsToken(c) {
			n++
		}
	}
	return n
}

// --- registration --------------------------------------------------

func TestBatch05CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B05SawInHalfOracle:         "Saw in Half",
		effectstest.B05TalismanOfImpulseOracle: "Talisman of Impulse",
		effectstest.B05ArborElfOracle:          "Arbor Elf",
		effectstest.B05ScrawlingCrawlerOracle:  "Scrawling Crawler",
		effectstest.B05MinamoOracle:            "Minamo, School at Water's Edge",
		effectstest.B05DeathriteShamanOracle:   "Deathrite Shaman",
		effectstest.B05TributeOracle:           "Tribute to the World Tree",
		effectstest.B05VaultOfWhispersOracle:   "Vault of Whispers",
		effectstest.B05ElasIlKorOracle:         "Elas il-Kor, Sadistic Pilgrim",
		effectstest.B05StitchersSupplierOracle: "Stitcher's Supplier",
		effectstest.B05AftermathAnalystOracle:  "Aftermath Analyst",
		effectstest.B05PlaguecrafterOracle:     "Plaguecrafter",
		effectstest.B05OpenTheArmoryOracle:     "Open the Armory",
		effectstest.B05SavageLandsOracle:       "Savage Lands",
		effectstest.B05OphiomancerOracle:       "Ophiomancer",
		effectstest.B05KnightWhiteOrchidOracle: "Knight of the White Orchid",
		effectstest.B05MentorOfTheMeekOracle:   "Mentor of the Meek",
		effectstest.B05RuinousUltimatumOracle:  "Ruinous Ultimatum",
		effectstest.B05DramaticReversalOracle:  "Dramatic Reversal",
		effectstest.B05IntangibleVirtueOracle:  "Intangible Virtue",
		effectstest.B05HaywireMiteOracle:       "Haywire Mite",
		effectstest.B05CharcoalDiamondOracle:   "Charcoal Diamond",
		effectstest.B05DualcasterMageOracle:    "Dualcaster Mage",
	}
	if len(want) != 23 {
		t.Fatalf("the batch is 23 cards, the table lists %d", len(want))
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

// The two cycle rows, pinned by shape so a transposed row is caught.
func TestBatch05CycleRowsHaveTheirPrintedShape(t *testing.T) {
	talisman, ok := effects.Lookup(effectstest.B05TalismanOfImpulseOracle)
	if !ok || len(talisman.ManaAbilities) != 2 {
		t.Fatalf("Talisman of Impulse: want two mana abilities, got %+v", talisman.ManaAbilities)
	}
	if talisman.ManaAbilities[0].Produced != "{C}" || talisman.ManaAbilities[1].Produced != "{R|G}" || talisman.ManaAbilities[1].PainToYou != 1 {
		t.Errorf("Talisman of Impulse: painless {C} first, then {R|G} with the damage rider; got %+v", talisman.ManaAbilities)
	}
	savage, ok := effects.Lookup(effectstest.B05SavageLandsOracle)
	if !ok || len(savage.Replacements) != 1 || len(savage.ManaAbilities) != 1 || savage.ManaAbilities[0].Produced != "{B|R|G}" {
		t.Errorf("Savage Lands: want enters-tapped plus a {B|R|G} pipe, got %+v", savage)
	}
}

// --- lands and rocks -----------------------------------------------

func TestVaultOfWhispersTapsForBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vault := effectstest.SeedPermanentWithOracle(g, me.ID, "Vault of Whispers", "Artifact Land", effectstest.B05VaultOfWhispersOracle)
	if err := g.ActivateManaAbility(me.ID, vault, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
	card, _ := effectstest.BattlefieldCard(g, vault)
	if !card.IsArtifact() {
		t.Error("an Artifact Land must count as an artifact")
	}
}

func TestCharcoalDiamondEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastCatalogSpell(t, g, "Charcoal Diamond", "Artifact", effectstest.B05CharcoalDiamondOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("the Diamond did not resolve to the battlefield")
	}
	if !card.Tapped {
		t.Error("Charcoal Diamond must enter tapped")
	}
	if n := effectstest.TapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events — it should have ENTERED tapped, not been tapped after", n)
	}
}

// --- Arbor Elf / Minamo --------------------------------------------

func TestArborElfUntapsAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := effectstest.PushCatalogPermanent(g, me.ID, "Arbor Elf", "Creature — Elf Druid", effectstest.B05ArborElfOracle, false)
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	b05TapOnBattlefield(g, forest)

	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: forest}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if card, _ := effectstest.BattlefieldCard(g, forest); card.Tapped {
		t.Error("the Forest was not untapped")
	}
	if card, _ := effectstest.BattlefieldCard(g, elf); !card.Tapped {
		t.Error("the Elf has a tap cost and must be tapped")
	}
}

func TestArborElfRefusesANonForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := effectstest.PushCatalogPermanent(g, me.ID, "Arbor Elf", "Creature — Elf Druid", effectstest.B05ArborElfOracle, false)
	island := effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: island}},
	}); err == nil {
		t.Error("an Island was accepted for 'target Forest'")
	}
}

func TestMinamoUntapsALegendaryPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	minamo := effectstest.SeedPermanentWithOracle(g, me.ID, "Minamo, School at Water's Edge", "Legendary Land", effectstest.B05MinamoOracle)
	legend := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Commander", TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	b05TapOnBattlefield(g, legend)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})

	if err := g.ActivateCatalogAbility(me.ID, minamo, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a nonlegendary creature was accepted for 'target legendary permanent'")
	}
	if err := g.ActivateCatalogAbility(me.ID, minamo, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: legend}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if card, _ := effectstest.BattlefieldCard(g, legend); card.Tapped {
		t.Error("the legendary creature was not untapped")
	}
}

// --- Deathrite Shaman ----------------------------------------------

func TestDeathriteShamanExilesALandCardForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := effectstest.PushCatalogPermanent(g, me.ID, "Deathrite Shaman", "Creature — Elf Shaman", effectstest.B05DeathriteShamanOracle, false)
	land := effectstest.Batch01GraveyardCard(opp, "Forest", "Basic Land — Forest")

	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(land) {
		t.Error("the land card was not exiled from the opponent's graveyard")
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Errorf("want a five-colour pick for 'add one mana of any color', got %+v", pick)
	}
}

func TestDeathriteShamanDrainsOffAnInstantAndGainsOffACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := effectstest.PushCatalogPermanent(g, me.ID, "Deathrite Shaman", "Creature — Elf Shaman", effectstest.B05DeathriteShamanOracle, false)
	bolt := effectstest.Batch01GraveyardCard(opp, "Lightning Bolt", "Instant")
	bear := effectstest.Batch01GraveyardCard(opp, "Bear", "Creature — Bear")
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "G"})
	before := effectstest.LifeOfOpponents(g)
	myLife := me.Life

	if err := g.ActivateCatalogAbility(me.ID, shaman, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("activate {B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d -> %d, want -2", i+1, b, got)
		}
	}
	if !g.Exile.Contains(bolt) {
		t.Error("the instant was not exiled")
	}

	// Untap for the second activation — the test is about the effect,
	// not the once-per-turn tap.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(shaman) })
	if err := g.ActivateCatalogAbility(me.ID, shaman, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate {G}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+2 {
		t.Errorf("life %d -> %d, want +2", myLife, me.Life)
	}
}

func TestDeathriteShamanRefusesTheWrongCardType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shaman := effectstest.PushCatalogPermanent(g, me.ID, "Deathrite Shaman", "Creature — Elf Shaman", effectstest.B05DeathriteShamanOracle, false)
	bear := effectstest.PushGraveyardCardForTest(me, "Bear")
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a creature card was accepted for 'target land card'")
	}
}

// --- Scrawling Crawler ---------------------------------------------

func TestScrawlingCrawlerUpkeepDrawsForAllAndDrainsOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Scrawling Crawler", "Artifact Creature — Phyrexian Construct", effectstest.B05ScrawlingCrawlerOracle, false)

	// Not on an opponent's upkeep.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if len(g.PendingTriggers)+len(g.StackMeta) != 0 {
		t.Fatal("the Crawler triggered on an opponent's upkeep")
	}
	// Each opponent's own draw step on the way round also drains
	// them — as printed — and AdvanceStep leaves those triggers on
	// the stack. Flush them at the last opponent's end step so the
	// snapshot below isolates the upkeep ability.
	effectstest.AdvanceToEndStepOf(t, g, 3)
	effectstest.PassPriorityAroundTable(t, g)
	// On the controller's.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	hands := make([]int, len(g.Seats))
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i], lives[i] = p.Hand.Size(), p.Life
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != hands[i]+1 {
			t.Errorf("seat %d drew %d, want 1", i, p.Hand.Size()-hands[i])
		}
		wantLife := lives[i]
		if i != 0 {
			wantLife--
		}
		if p.Life != wantLife {
			t.Errorf("seat %d life %d -> %d, want %d", i, lives[i], p.Life, wantLife)
		}
	}
}

// --- Tribute to the World Tree -------------------------------------

func TestTributeDrawsForBigCreaturesAndGrowsSmallOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Tribute to the World Tree", "Enchantment", effectstest.B05TributeOracle, false)
	before := me.Hand.Size()

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Beast"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("a 3/3 entering should draw one card, drew %d", me.Hand.Size()-before)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")
	card, _ := effectstest.BattlefieldCard(g, goblin)
	if card.Counters["+1/+1"] != 2 {
		t.Errorf("a 1/1 entering should get two +1/+1 counters, got %v", card.Counters)
	}
	if me.Hand.Size() != before+1 {
		t.Error("the small creature must not also draw")
	}
}

// --- Elas il-Kor ---------------------------------------------------

func TestElasIlKorGainsOnEntryAndDrainsOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elas := effectstest.PushCatalogPermanent(g, me.ID, "Elas il-Kor, Sadistic Pilgrim", "Legendary Creature — Phyrexian Kor Cleric", effectstest.B05ElasIlKorOracle, false)
	myLife := me.Life
	before := effectstest.LifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+1 {
		t.Errorf("another creature entering: life %d -> %d, want +1", myLife, me.Life)
	}

	goblin := effectstest.FindBattlefieldByName(g, "Goblin")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(goblin) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d after another creature died: %d -> %d, want -1", i+1, b, got)
		}
	}

	// Elas's own death is not "another creature".
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(elas) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d after Elas died: %d -> %d, want still -1", i+1, b, got)
		}
	}
}

// --- Stitcher's Supplier / Aftermath Analyst -----------------------

func TestStitchersSupplierMillsOnEntryAndOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	supplier := effectstest.CastCatalogSpell(t, g, "Stitcher's Supplier", "Creature — Zombie", effectstest.B05StitchersSupplierOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 3 {
		t.Fatalf("graveyard after entry = %d, want 3 milled", me.Graveyard.Size())
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(supplier) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 7 {
		t.Errorf("graveyard after death = %d, want 7 (3 + the Supplier + 3)", me.Graveyard.Size())
	}
}

func TestAftermathAnalystMillsThenReturnsEveryLandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	analyst := effectstest.CastCatalogSpell(t, g, "Aftermath Analyst", "Creature — Elf Detective", effectstest.B05AftermathAnalystOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 3 {
		t.Fatalf("graveyard after entry = %d, want 3 milled", me.Graveyard.Size())
	}
	forest := effectstest.Batch01GraveyardCard(me, "Forest", "Basic Land — Forest")
	island := effectstest.Batch01GraveyardCard(me, "Island", "Basic Land — Island")
	bear := effectstest.PushGraveyardCardForTest(me, "Bear")
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, analyst, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(analyst) {
		t.Error("the Analyst is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{forest, island} {
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Errorf("%s did not return", id)
			continue
		}
		if !card.Tapped {
			t.Errorf("%s returned untapped", id)
		}
	}
	if g.Battlefield.Contains(bear) {
		t.Error("a creature card came back — only land cards return")
	}
}

// --- Plaguecrafter -------------------------------------------------

func TestPlaguecrafterEdictsThoseWhoCanAndDiscardsThoseWhoCant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1 := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedCreature(g, "Their Bear", opp1.ID)
	// Seats 2 and 3 control nothing; they hold seven cards each.

	crafter := effectstest.CastCatalogSpell(t, g, "Plaguecrafter", "Creature — Human Shaman", effectstest.B05PlaguecrafterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// The controller's only creature is the Plaguecrafter and the
	// opponent's is the bear, so both picks are forced and collected
	// unasked (ADR 0175, #3664).
	if g.Battlefield.Contains(crafter) || g.Battlefield.Contains(bear) {
		t.Error("the Plaguecrafter's controller and the bear's owner must sacrifice")
	}
	// "Each player who can't discards a card" is the Mind Rot path:
	// its own prompt, addressed to that player over their own hand.
	if effectstest.DiscardOwed(g, g.Seats[2].ID) != 1 || effectstest.DiscardOwed(g, g.Seats[3].ID) != 1 {
		t.Error("players with no creature or planeswalker must be asked to discard one card")
	}
	if effectstest.DiscardOwed(g, me.ID) != 0 || effectstest.DiscardOwed(g, opp1.ID) != 0 || effectstest.SacrificeChoiceFor(g, g.Seats[2].ID) != nil {
		t.Error("a player gets exactly one of the two")
	}
}

// --- Open the Armory -----------------------------------------------

func TestOpenTheArmoryFindsAnAuraOrEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aura := effectstest.StapleLibraryCard(me, "Rancor", "Enchantment — Aura")
	effectstest.StapleLibraryCard(me, "Glorious Anthem", "Enchantment")

	effectstest.CastCatalogSpell(t, g, "Open the Armory", "Sorcery", effectstest.B05OpenTheArmoryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerSearchByID(t, g, me.ID, aura)
	}
	if !me.Hand.Contains(aura) {
		t.Error("the Aura did not reach the hand")
	}
}

// --- Ophiomancer ---------------------------------------------------

func TestOphiomancerMakesASnakeOnlyWhileYouHaveNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Ophiomancer", "Creature — Human Shaman", effectstest.B05OphiomancerOracle, false)

	effectstest.AdvanceToUpkeepOf(t, g, 1) // EACH upkeep — an opponent's counts
	effectstest.PassPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Snake"); n != 1 {
		t.Fatalf("Snakes after the first upkeep = %d, want 1", n)
	}
	snake := effectstest.FindBattlefieldByName(g, "Snake")
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, snake), "deathtouch") {
		t.Error("the Snake must have deathtouch")
	}

	effectstest.AdvanceToUpkeepOf(t, g, 2)
	if len(g.PendingTriggers)+len(g.StackMeta) != 0 {
		t.Fatal("triggered with a Snake already out — the intervening-if must stop it")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(snake) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Snake"); n != 1 {
		t.Errorf("Snakes after losing one and reaching another upkeep = %d, want 1", n)
	}
}

// --- Knight of the White Orchid ------------------------------------

func TestKnightOfTheWhiteOrchidFetchesAPlainsWhenBehind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest A", "Basic Land — Forest")
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest B", "Basic Land — Forest")
	plains := effectstest.StapleLibraryCard(me, "Plains", "Basic Land — Plains")

	effectstest.CastCatalogSpell(t, g, "Knight of the White Orchid", "Creature — Human Knight", effectstest.B05KnightWhiteOrchidOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerSearchByID(t, g, me.ID, plains)
	}
	card, ok := effectstest.BattlefieldCard(g, plains)
	if !ok {
		t.Fatal("the Plains was not fetched")
	}
	if card.Tapped {
		t.Error("Knight of the White Orchid's Plains enters UNTAPPED")
	}
}

func TestKnightOfTheWhiteOrchidStaysQuietWhenNotBehind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.CastCatalogSpell(t, g, "Knight of the White Orchid", "Creature — Human Knight", effectstest.B05KnightWhiteOrchidOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
			t.Fatal("prompted although no opponent controls more lands — CR 603.4 says no trigger at all")
		}
	}
}

// --- Mentor of the Meek --------------------------------------------

func TestMentorOfTheMeekOffersToPayForSmallCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Mentor of the Meek", "Creature — Human Soldier", effectstest.B05MentorOfTheMeekOracle, false)
	before := me.Hand.Size()

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a 1/1 entering should ask the controller to pay {1}")
	}
	// The pool empties between steps (CR 106.4); fund the payment
	// when the prompt is answered, as Hashaton's test does.
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if me.Hand.Size() != before+1 {
		t.Errorf("paying should draw one card, drew %d", me.Hand.Size()-before)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Beast"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a 3/3 entering must not trigger 'power 2 or less'")
	}
}

// --- Ruinous Ultimatum / Dramatic Reversal -------------------------

func TestRuinousUltimatumDestroysOnlyOpponentsNonlands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedCreature(g, "My Bear", me.ID)
	theirs := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")

	effectstest.CastCatalogSpell(t, g, "Ruinous Ultimatum", "Sorcery", effectstest.B05RuinousUltimatumOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(rock) {
		t.Error("an opponent's nonland permanent survived")
	}
	if !g.Battlefield.Contains(land) || !g.Battlefield.Contains(mine) {
		t.Error("lands and the caster's own permanents must be untouched")
	}
}

func TestDramaticReversalUntapsNonlandsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	rock := effectstest.SeedPermanentFor(g, me.ID, "Rock", "Artifact")
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	for _, id := range []uuid.UUID{bear, rock, land} {
		b05TapOnBattlefield(g, id)
	}

	effectstest.CastCatalogSpell(t, g, "Dramatic Reversal", "Instant", effectstest.B05DramaticReversalOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{bear, rock} {
		if card, _ := effectstest.BattlefieldCard(g, id); card.Tapped {
			t.Errorf("%s stayed tapped", card.Name)
		}
	}
	if card, _ := effectstest.BattlefieldCard(g, land); !card.Tapped {
		t.Error("a land must stay tapped — the card says nonland")
	}
}

// --- Intangible Virtue ---------------------------------------------

func TestIntangibleVirtuePumpsCreatureTokensOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Intangible Virtue", TypeLine: "Enchantment",
		OracleID: effectstest.B05IntangibleVirtueOracle, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")

	if p, tough := effectstest.EffectivePower(t, g, goblin), effectstest.EffectiveToughness(t, g, goblin); p != 2 || tough != 2 {
		t.Errorf("Goblin token = %d/%d, want 2/2", p, tough)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, goblin), "vigilance") {
		t.Error("the token must have vigilance")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("nontoken Bear = %d power, want 2 (not pumped)", p)
	}
	if effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, bear), "vigilance") {
		t.Error("a nontoken creature must not gain vigilance")
	}
}

// --- Haywire Mite --------------------------------------------------

func TestHaywireMiteExilesAnArtifactAndGainsOnItsDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mite := effectstest.PushCatalogPermanent(g, me.ID, "Haywire Mite", "Artifact Creature — Insect", effectstest.B05HaywireMiteOracle, false)
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")
	golem := effectstest.SeedPermanentFor(g, opp.ID, "Golem", "Artifact Creature — Golem")
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	before := me.Life

	if err := g.ActivateCatalogAbility(me.ID, mite, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: golem}},
	}); err == nil {
		t.Error("an artifact creature was accepted for 'noncreature artifact'")
	}
	if err := g.ActivateCatalogAbility(me.ID, mite, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) {
		t.Error("the artifact was not exiled")
	}
	if me.Life != before+2 {
		t.Errorf("the Mite died as a cost: life %d -> %d, want +2", before, me.Life)
	}
}

// --- Saw in Half ---------------------------------------------------

func TestSawInHalfMakesTwoHalfSizeCopiesForTheVictim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Titan", TypeLine: "Creature — Giant",
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Saw in Half", "Instant", effectstest.B05SawInHalfOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: titan}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(titan) {
		t.Fatal("the creature survived")
	}
	if n := b05CountTokensNamed(g, opp.ID, "Titan"); n != 2 {
		t.Fatalf("victim controls %d Titan tokens, want 2", n)
	}
	if n := b05CountTokensNamed(g, me.ID, "Titan"); n != 0 {
		t.Error("the caster must not get the tokens")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Titan" && effects.IsToken(c) {
			if p, tough := effectstest.EffectivePower(t, g, c.InstanceID), effectstest.EffectiveToughness(t, g, c.InstanceID); p != 3 || tough != 3 {
				t.Errorf("token = %d/%d, want 3/3 (half of 5, rounded up)", p, tough)
			}
		}
	}
}

// --- Dualcaster Mage -----------------------------------------------

// Flash the Mage in on top of your own Bolt, pick the Bolt for the
// ETB, re-aim the copy at a second opponent. Both Bolts deal their
// 3, the Mage stays on the battlefield, and the copy is not a card
// (CR 707.10) so the graveyard holds the Bolt alone.
func TestDualcasterMageCopiesTheSpellBeneathIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID

	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	// Flash: cast with the Bolt still on the stack.
	mage := effectstest.CastCatalogSpell(t, g, "Dualcaster Mage", "Creature — Human Wizard", effectstest.B05DualcasterMageOracle, nil)

	// The Mage resolves first; its ETB trigger targets, so the
	// controller is asked which spell to copy before it goes on the
	// stack.
	for i := 0; i < 8 && effectstest.LatestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pick := effectstest.LatestPickTarget(g, me)
	if pick == nil {
		t.Fatal("the ETB should ask for the spell to copy")
	}
	if !effectstest.HasID(pick.PickTargetCards, bolt) {
		t.Fatalf("the Bolt should be a legal target: %v", pick.PickTargetCards)
	}
	if !g.Battlefield.Contains(mage) {
		t.Fatal("the Mage should be on the battlefield when its ETB asks")
	}
	effectstest.PickCard(t, g, me, bolt)

	// The trigger resolves and opens the CR 707.10c re-target prompt.
	for i := 0; i < 8 && effectstest.LatestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	retarget := effectstest.LatestPickTarget(g, me)
	if retarget == nil {
		t.Fatal("no re-target prompt for the copy")
	}
	if err := g.ResolvePickTarget(retarget.ID, me,
		game.TargetRef{Kind: game.TargetPlayer, ID: victimB}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.LifeOf(g, victimA); got != 37 {
		t.Errorf("original target life = %d, want 37", got)
	}
	if got := effectstest.LifeOf(g, victimB); got != 37 {
		t.Errorf("re-targeted copy life = %d, want 37", got)
	}
	if !g.Battlefield.Contains(mage) {
		t.Error("the Mage must stay on the battlefield")
	}
	if got := effectstest.GraveyardSize(g, me); got != 1 {
		t.Errorf("graveyard = %d cards, want 1 (the Bolt; a copy is not a card)", got)
	}
}

// With nothing to copy the trigger is removed without a prompt (CR
// 603.3d) — a Mage cast on an empty stack is just a 2/2.
func TestDualcasterMageWithNoSpellToCopyAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	effectstest.CastCatalogSpell(t, g, "Dualcaster Mage", "Creature — Human Wizard", effectstest.B05DualcasterMageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me) != nil {
		t.Error("no instant or sorcery on the stack — the trigger should have been removed, not prompted")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("no prompt of any kind should be pending, got %d", len(g.PendingChoices))
	}
}
