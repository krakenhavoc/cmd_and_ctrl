package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch05_test.go — card-level coverage for the card-coverage
// roadmap's batch 05 (#298, `edhrec_rank` 578–693): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation or upkeep rather than by calling primitives.
// Every batch-specific name is b05-prefixed (batches 03, 04, 06 and 07
// are in flight beside this one).

const (
	b05SawInHalfOracle         = "eea18c55-8695-4ba1-9b38-3e7638692f5f"
	b05TalismanOfImpulseOracle = "f2ccc9e8-8e92-4f8c-8728-8c748630e0dd"
	b05ArborElfOracle          = "4567a528-75f0-4ea6-b927-3a500caf76ac"
	b05ScrawlingCrawlerOracle  = "1d60c80d-9a95-4051-bbf7-02ca51ee1be2"
	b05MinamoOracle            = "17784f90-89a1-47a5-83ef-ae60dfc30bd1"
	b05DeathriteShamanOracle   = "22f1a4a4-c423-4d1c-8775-0ed604a9fa51"
	b05TributeOracle           = "72deedab-7c17-4505-aeca-4bc8596d80a5"
	b05VaultOfWhispersOracle   = "09496421-74e4-466a-9546-56f2a0c8eef4"
	b05ElasIlKorOracle         = "a4d33e21-d3fe-4d5b-903c-b75e859d6f7f"
	b05StitchersSupplierOracle = "7fd61a18-6e4f-40c5-aa00-3d101ec1ec82"
	b05AftermathAnalystOracle  = "374e54d8-8e73-4268-9dde-c28c77bbbf32"
	b05PlaguecrafterOracle     = "2e3ee458-fa3f-4452-ab95-5a7fb5a0483b"
	b05OpenTheArmoryOracle     = "2ecf7771-8061-4710-ad2e-80b092ae0b4b"
	b05SavageLandsOracle       = "a3292406-3f49-42d6-a547-e43dd5797f84"
	b05OphiomancerOracle       = "55eca80c-dcd8-4c2f-aa0f-fb0aec7b80f7"
	b05KnightWhiteOrchidOracle = "7d0efc64-151f-45b4-8c0b-c32609f9d862"
	b05MentorOfTheMeekOracle   = "b9f4f96b-6e54-4fe6-8df7-623e0fc72409"
	b05RuinousUltimatumOracle  = "a6f38908-aa4f-4f99-a28e-85d11dab52e4"
	b05DramaticReversalOracle  = "da46904c-8fb8-44c2-b2ab-775a1cc12ec3"
	b05IntangibleVirtueOracle  = "d21c3c8f-d105-4ba9-bf69-e5f26f0f8ec5"
	b05HaywireMiteOracle       = "749d2994-44e7-40d3-8630-7bebed239e9e"
	b05CharcoalDiamondOracle   = "1386d111-a2a7-4df1-91d7-947664126989"
	b05DualcasterMageOracle    = "8eb7c0a5-6190-40de-b473-2d1daa3bbe28"
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
		if c.Name == name && c.Controller == controller && IsToken(c) {
			n++
		}
	}
	return n
}

// --- registration --------------------------------------------------

func TestBatch05CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b05SawInHalfOracle:         "Saw in Half",
		b05TalismanOfImpulseOracle: "Talisman of Impulse",
		b05ArborElfOracle:          "Arbor Elf",
		b05ScrawlingCrawlerOracle:  "Scrawling Crawler",
		b05MinamoOracle:            "Minamo, School at Water's Edge",
		b05DeathriteShamanOracle:   "Deathrite Shaman",
		b05TributeOracle:           "Tribute to the World Tree",
		b05VaultOfWhispersOracle:   "Vault of Whispers",
		b05ElasIlKorOracle:         "Elas il-Kor, Sadistic Pilgrim",
		b05StitchersSupplierOracle: "Stitcher's Supplier",
		b05AftermathAnalystOracle:  "Aftermath Analyst",
		b05PlaguecrafterOracle:     "Plaguecrafter",
		b05OpenTheArmoryOracle:     "Open the Armory",
		b05SavageLandsOracle:       "Savage Lands",
		b05OphiomancerOracle:       "Ophiomancer",
		b05KnightWhiteOrchidOracle: "Knight of the White Orchid",
		b05MentorOfTheMeekOracle:   "Mentor of the Meek",
		b05RuinousUltimatumOracle:  "Ruinous Ultimatum",
		b05DramaticReversalOracle:  "Dramatic Reversal",
		b05IntangibleVirtueOracle:  "Intangible Virtue",
		b05HaywireMiteOracle:       "Haywire Mite",
		b05CharcoalDiamondOracle:   "Charcoal Diamond",
		b05DualcasterMageOracle:    "Dualcaster Mage",
	}
	if len(want) != 23 {
		t.Fatalf("the batch is 23 cards, the table lists %d", len(want))
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

// The two cycle rows, pinned by shape so a transposed row is caught.
func TestBatch05CycleRowsHaveTheirPrintedShape(t *testing.T) {
	talisman, ok := Lookup(b05TalismanOfImpulseOracle)
	if !ok || len(talisman.ManaAbilities) != 2 {
		t.Fatalf("Talisman of Impulse: want two mana abilities, got %+v", talisman.ManaAbilities)
	}
	if talisman.ManaAbilities[0].Produced != "{C}" || talisman.ManaAbilities[1].Produced != "{R|G}" || talisman.ManaAbilities[1].Rider == nil {
		t.Errorf("Talisman of Impulse: painless {C} first, then {R|G} with the damage rider; got %+v", talisman.ManaAbilities)
	}
	savage, ok := Lookup(b05SavageLandsOracle)
	if !ok || len(savage.Replacements) != 1 || len(savage.ManaAbilities) != 1 || savage.ManaAbilities[0].Produced != "{B|R|G}" {
		t.Errorf("Savage Lands: want enters-tapped plus a {B|R|G} pipe, got %+v", savage)
	}
}

// --- lands and rocks -----------------------------------------------

func TestVaultOfWhispersTapsForBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vault := seedPermanentWithOracle(g, me.ID, "Vault of Whispers", "Artifact Land", b05VaultOfWhispersOracle)
	if err := g.ActivateManaAbility(me.ID, vault, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
	card, _ := battlefieldCard(g, vault)
	if !card.IsArtifact() {
		t.Error("an Artifact Land must count as an artifact")
	}
}

func TestCharcoalDiamondEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Charcoal Diamond", "Artifact", b05CharcoalDiamondOracle, nil)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the Diamond did not resolve to the battlefield")
	}
	if !card.Tapped {
		t.Error("Charcoal Diamond must enter tapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events — it should have ENTERED tapped, not been tapped after", n)
	}
}

// --- Arbor Elf / Minamo --------------------------------------------

func TestArborElfUntapsAForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := pushCatalogPermanent(g, me.ID, "Arbor Elf", "Creature — Elf Druid", b05ArborElfOracle, false)
	forest := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	b05TapOnBattlefield(g, forest)

	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: forest}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if card, _ := battlefieldCard(g, forest); card.Tapped {
		t.Error("the Forest was not untapped")
	}
	if card, _ := battlefieldCard(g, elf); !card.Tapped {
		t.Error("the Elf has a tap cost and must be tapped")
	}
}

func TestArborElfRefusesANonForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := pushCatalogPermanent(g, me.ID, "Arbor Elf", "Creature — Elf Druid", b05ArborElfOracle, false)
	island := seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: island}},
	}); err == nil {
		t.Error("an Island was accepted for 'target Forest'")
	}
}

func TestMinamoUntapsALegendaryPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	minamo := seedPermanentWithOracle(g, me.ID, "Minamo, School at Water's Edge", "Legendary Land", b05MinamoOracle)
	legend := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Commander", TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	bear := seedCreature(g, "Bear", me.ID)
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
	passPriorityAroundTable(t, g)
	if card, _ := battlefieldCard(g, legend); card.Tapped {
		t.Error("the legendary creature was not untapped")
	}
}

// --- Deathrite Shaman ----------------------------------------------

func TestDeathriteShamanExilesALandCardForMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := pushCatalogPermanent(g, me.ID, "Deathrite Shaman", "Creature — Elf Shaman", b05DeathriteShamanOracle, false)
	land := batch01GraveyardCard(opp, "Forest", "Basic Land — Forest")

	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(land) {
		t.Error("the land card was not exiled from the opponent's graveyard")
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Errorf("want a five-colour pick for 'add one mana of any color', got %+v", pick)
	}
}

func TestDeathriteShamanDrainsOffAnInstantAndGainsOffACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := pushCatalogPermanent(g, me.ID, "Deathrite Shaman", "Creature — Elf Shaman", b05DeathriteShamanOracle, false)
	bolt := batch01GraveyardCard(opp, "Lightning Bolt", "Instant")
	bear := batch01GraveyardCard(opp, "Bear", "Creature — Bear")
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "G"})
	before := lifeOfOpponents(g)
	myLife := me.Life

	if err := g.ActivateCatalogAbility(me.ID, shaman, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("activate {B}: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	passPriorityAroundTable(t, g)
	if me.Life != myLife+2 {
		t.Errorf("life %d -> %d, want +2", myLife, me.Life)
	}
}

func TestDeathriteShamanRefusesTheWrongCardType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shaman := pushCatalogPermanent(g, me.ID, "Deathrite Shaman", "Creature — Elf Shaman", b05DeathriteShamanOracle, false)
	bear := pushGraveyardCardForTest(me, "Bear")
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a creature card was accepted for 'target land card'")
	}
}

// --- Scrawling Crawler ---------------------------------------------

func TestScrawlingCrawlerUpkeepDrawsForAllAndDrainsOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Scrawling Crawler", "Artifact Creature — Phyrexian Construct", b05ScrawlingCrawlerOracle, false)

	// Not on an opponent's upkeep.
	advanceToUpkeepOf(t, g, 1)
	if len(g.PendingTriggers)+len(g.StackMeta) != 0 {
		t.Fatal("the Crawler triggered on an opponent's upkeep")
	}
	// Each opponent's own draw step on the way round also drains
	// them — as printed — and AdvanceStep leaves those triggers on
	// the stack. Flush them at the last opponent's end step so the
	// snapshot below isolates the upkeep ability.
	advanceToEndStepOf(t, g, 3)
	passPriorityAroundTable(t, g)
	// On the controller's.
	advanceToUpkeepOf(t, g, 0)
	hands := make([]int, len(g.Seats))
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i], lives[i] = p.Hand.Size(), p.Life
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Tribute to the World Tree", "Enchantment", b05TributeOracle, false)
	before := me.Hand.Size()

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 colorless Beast"), 1) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("a 3/3 entering should draw one card, drew %d", me.Hand.Size()-before)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)
	goblin := findBattlefieldByName(g, "Goblin")
	card, _ := battlefieldCard(g, goblin)
	if card.Counters["+1/+1"] != 2 {
		t.Errorf("a 1/1 entering should get two +1/+1 counters, got %v", card.Counters)
	}
	if me.Hand.Size() != before+1 {
		t.Error("the small creature must not also draw")
	}
}

// --- Elas il-Kor ---------------------------------------------------

func TestElasIlKorGainsOnEntryAndDrainsOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elas := pushCatalogPermanent(g, me.ID, "Elas il-Kor, Sadistic Pilgrim", "Legendary Creature — Phyrexian Kor Cleric", b05ElasIlKorOracle, false)
	myLife := me.Life
	before := lifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)
	if me.Life != myLife+1 {
		t.Errorf("another creature entering: life %d -> %d, want +1", myLife, me.Life)
	}

	goblin := findBattlefieldByName(g, "Goblin")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(goblin) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d after another creature died: %d -> %d, want -1", i+1, b, got)
		}
	}

	// Elas's own death is not "another creature".
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(elas) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d after Elas died: %d -> %d, want still -1", i+1, b, got)
		}
	}
}

// --- Stitcher's Supplier / Aftermath Analyst -----------------------

func TestStitchersSupplierMillsOnEntryAndOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	supplier := castCatalogSpell(t, g, "Stitcher's Supplier", "Creature — Zombie", b05StitchersSupplierOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 3 {
		t.Fatalf("graveyard after entry = %d, want 3 milled", me.Graveyard.Size())
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(supplier) })
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 7 {
		t.Errorf("graveyard after death = %d, want 7 (3 + the Supplier + 3)", me.Graveyard.Size())
	}
}

func TestAftermathAnalystMillsThenReturnsEveryLandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	analyst := castCatalogSpell(t, g, "Aftermath Analyst", "Creature — Elf Detective", b05AftermathAnalystOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 3 {
		t.Fatalf("graveyard after entry = %d, want 3 milled", me.Graveyard.Size())
	}
	forest := batch01GraveyardCard(me, "Forest", "Basic Land — Forest")
	island := batch01GraveyardCard(me, "Island", "Basic Land — Island")
	bear := pushGraveyardCardForTest(me, "Bear")
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, analyst, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(analyst) {
		t.Error("the Analyst is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{forest, island} {
		card, ok := battlefieldCard(g, id)
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
	g := newCatalogGame(t)
	me, opp1 := g.Seats[0], g.Seats[1]
	seedCreature(g, "Their Bear", opp1.ID)
	// Seats 2 and 3 control nothing; they hold seven cards each.

	castCatalogSpell(t, g, "Plaguecrafter", "Creature — Human Shaman", b05PlaguecrafterOracle, nil)
	passPriorityAroundTable(t, g)

	sac := map[uuid.UUID]bool{}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSacrifice {
			sac[c.Chooser] = true
		}
	}
	// "Each player who can't discards a card" is the Mind Rot path:
	// its own prompt, addressed to that player over their own hand.
	if !sac[me.ID] || !sac[opp1.ID] {
		t.Errorf("the Plaguecrafter's controller and the bear's owner must be asked to sacrifice: %v", sac)
	}
	if discardOwed(g, g.Seats[2].ID) != 1 || discardOwed(g, g.Seats[3].ID) != 1 {
		t.Error("players with no creature or planeswalker must be asked to discard one card")
	}
	if discardOwed(g, me.ID) != 0 || discardOwed(g, opp1.ID) != 0 || sac[g.Seats[2].ID] {
		t.Error("a player gets exactly one of the two prompts")
	}
}

// --- Open the Armory -----------------------------------------------

func TestOpenTheArmoryFindsAnAuraOrEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aura := stapleLibraryCard(me, "Rancor", "Enchantment — Aura")
	stapleLibraryCard(me, "Glorious Anthem", "Enchantment")

	castCatalogSpell(t, g, "Open the Armory", "Sorcery", b05OpenTheArmoryOracle, nil)
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		answerSearchByID(t, g, me.ID, aura)
	}
	if !me.Hand.Contains(aura) {
		t.Error("the Aura did not reach the hand")
	}
}

// --- Ophiomancer ---------------------------------------------------

func TestOphiomancerMakesASnakeOnlyWhileYouHaveNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Ophiomancer", "Creature — Human Shaman", b05OphiomancerOracle, false)

	advanceToUpkeepOf(t, g, 1) // EACH upkeep — an opponent's counts
	passPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Snake"); n != 1 {
		t.Fatalf("Snakes after the first upkeep = %d, want 1", n)
	}
	snake := findBattlefieldByName(g, "Snake")
	if !eotHasAbility(effectiveAbilities(t, g, snake), "deathtouch") {
		t.Error("the Snake must have deathtouch")
	}

	advanceToUpkeepOf(t, g, 2)
	if len(g.PendingTriggers)+len(g.StackMeta) != 0 {
		t.Fatal("triggered with a Snake already out — the intervening-if must stop it")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(snake) })
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, 3)
	passPriorityAroundTable(t, g)
	if n := b05CountTokensNamed(g, me.ID, "Snake"); n != 1 {
		t.Errorf("Snakes after losing one and reaching another upkeep = %d, want 1", n)
	}
}

// --- Knight of the White Orchid ------------------------------------

func TestKnightOfTheWhiteOrchidFetchesAPlainsWhenBehind(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLandOnBattlefield(g, opp.ID, "Forest A", "Basic Land — Forest")
	seedLandOnBattlefield(g, opp.ID, "Forest B", "Basic Land — Forest")
	plains := stapleLibraryCard(me, "Plains", "Basic Land — Plains")

	castCatalogSpell(t, g, "Knight of the White Orchid", "Creature — Human Knight", b05KnightWhiteOrchidOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		answerSearchByID(t, g, me.ID, plains)
	}
	card, ok := battlefieldCard(g, plains)
	if !ok {
		t.Fatal("the Plains was not fetched")
	}
	if card.Tapped {
		t.Error("Knight of the White Orchid's Plains enters UNTAPPED")
	}
}

func TestKnightOfTheWhiteOrchidStaysQuietWhenNotBehind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	castCatalogSpell(t, g, "Knight of the White Orchid", "Creature — Human Knight", b05KnightWhiteOrchidOracle, nil)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
			t.Fatal("prompted although no opponent controls more lands — CR 603.4 says no trigger at all")
		}
	}
}

// --- Mentor of the Meek --------------------------------------------

func TestMentorOfTheMeekOffersToPayForSmallCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Mentor of the Meek", "Creature — Human Soldier", b05MentorOfTheMeekOracle, false)
	before := me.Hand.Size()

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("a 1/1 entering should ask the controller to pay {1}")
	}
	// The pool empties between steps (CR 106.4); fund the payment
	// when the prompt is answered, as Hashaton's test does.
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	answerPayUnless(t, g, me.ID, true)
	if me.Hand.Size() != before+1 {
		t.Errorf("paying should draw one card, drew %d", me.Hand.Size()-before)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 colorless Beast"), 1) })
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("a 3/3 entering must not trigger 'power 2 or less'")
	}
}

// --- Ruinous Ultimatum / Dramatic Reversal -------------------------

func TestRuinousUltimatumDestroysOnlyOpponentsNonlands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := seedCreature(g, "My Bear", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")
	land := seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")

	castCatalogSpell(t, g, "Ruinous Ultimatum", "Sorcery", b05RuinousUltimatumOracle, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(rock) {
		t.Error("an opponent's nonland permanent survived")
	}
	if !g.Battlefield.Contains(land) || !g.Battlefield.Contains(mine) {
		t.Error("lands and the caster's own permanents must be untouched")
	}
}

func TestDramaticReversalUntapsNonlandsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedCreature(g, "Bear", me.ID)
	rock := seedPermanentFor(g, me.ID, "Rock", "Artifact")
	land := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	for _, id := range []uuid.UUID{bear, rock, land} {
		b05TapOnBattlefield(g, id)
	}

	castCatalogSpell(t, g, "Dramatic Reversal", "Instant", b05DramaticReversalOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{bear, rock} {
		if card, _ := battlefieldCard(g, id); card.Tapped {
			t.Errorf("%s stayed tapped", card.Name)
		}
	}
	if card, _ := battlefieldCard(g, land); !card.Tapped {
		t.Error("a land must stay tapped — the card says nonland")
	}
}

// --- Intangible Virtue ---------------------------------------------

func TestIntangibleVirtuePumpsCreatureTokensOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Intangible Virtue", TypeLine: "Enchantment",
		OracleID: b05IntangibleVirtueOracle, Owner: me.ID, Controller: me.ID,
	})
	bear := seedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	goblin := findBattlefieldByName(g, "Goblin")

	if p, tough := effectivePower(t, g, goblin), effectiveToughness(t, g, goblin); p != 2 || tough != 2 {
		t.Errorf("Goblin token = %d/%d, want 2/2", p, tough)
	}
	if !eotHasAbility(effectiveAbilities(t, g, goblin), "vigilance") {
		t.Error("the token must have vigilance")
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("nontoken Bear = %d power, want 2 (not pumped)", p)
	}
	if eotHasAbility(effectiveAbilities(t, g, bear), "vigilance") {
		t.Error("a nontoken creature must not gain vigilance")
	}
}

// --- Haywire Mite --------------------------------------------------

func TestHaywireMiteExilesAnArtifactAndGainsOnItsDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mite := pushCatalogPermanent(g, me.ID, "Haywire Mite", "Artifact Creature — Insect", b05HaywireMiteOracle, false)
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")
	golem := seedPermanentFor(g, opp.ID, "Golem", "Artifact Creature — Golem")
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
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) {
		t.Error("the artifact was not exiled")
	}
	if me.Life != before+2 {
		t.Errorf("the Mite died as a cost: life %d -> %d, want +2", before, me.Life)
	}
}

// --- Saw in Half ---------------------------------------------------

func TestSawInHalfMakesTwoHalfSizeCopiesForTheVictim(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Titan", TypeLine: "Creature — Giant",
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Saw in Half", "Instant", b05SawInHalfOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: titan}})
	passPriorityAroundTable(t, g)

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
		if c.Name == "Titan" && IsToken(c) {
			if p, tough := effectivePower(t, g, c.InstanceID), effectiveToughness(t, g, c.InstanceID); p != 3 || tough != 3 {
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
	g := newCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	// Flash: cast with the Bolt still on the stack.
	mage := castCatalogSpell(t, g, "Dualcaster Mage", "Creature — Human Wizard", b05DualcasterMageOracle, nil)

	// The Mage resolves first; its ETB trigger targets, so the
	// controller is asked which spell to copy before it goes on the
	// stack.
	for i := 0; i < 8 && latestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pick := latestPickTarget(g, me)
	if pick == nil {
		t.Fatal("the ETB should ask for the spell to copy")
	}
	if !hasID(pick.PickTargetCards, bolt) {
		t.Fatalf("the Bolt should be a legal target: %v", pick.PickTargetCards)
	}
	if !g.Battlefield.Contains(mage) {
		t.Fatal("the Mage should be on the battlefield when its ETB asks")
	}
	pickCard(t, g, me, bolt)

	// The trigger resolves and opens the CR 707.10c re-target prompt.
	for i := 0; i < 8 && latestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	retarget := latestPickTarget(g, me)
	if retarget == nil {
		t.Fatal("no re-target prompt for the copy")
	}
	if err := g.ResolvePickTarget(retarget.ID, me,
		game.TargetRef{Kind: game.TargetPlayer, ID: victimB}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := lifeOf(g, victimA); got != 37 {
		t.Errorf("original target life = %d, want 37", got)
	}
	if got := lifeOf(g, victimB); got != 37 {
		t.Errorf("re-targeted copy life = %d, want 37", got)
	}
	if !g.Battlefield.Contains(mage) {
		t.Error("the Mage must stay on the battlefield")
	}
	if got := graveyardSize(g, me); got != 1 {
		t.Errorf("graveyard = %d cards, want 1 (the Bolt; a copy is not a card)", got)
	}
}

// With nothing to copy the trigger is removed without a prompt (CR
// 603.3d) — a Mage cast on an empty stack is just a 2/2.
func TestDualcasterMageWithNoSpellToCopyAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	castCatalogSpell(t, g, "Dualcaster Mage", "Creature — Human Wizard", b05DualcasterMageOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me) != nil {
		t.Error("no instant or sorcery on the stack — the trigger should have been removed, not prompted")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("no prompt of any kind should be pending, got %d", len(g.PendingChoices))
	}
}
