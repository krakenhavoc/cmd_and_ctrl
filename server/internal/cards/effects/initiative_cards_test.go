package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// initiative_cards_test.go pins the catalog cards built on the
// initiative (ADR 0203, #4096): one test per card, plus From the
// Catacombs' escape and its exile-instead clause.

const (
	fromTheCatacombsOracle        = "cc21b29f-2d1a-4769-9b65-f92bd2ca1cc8"
	aarakocraSneakOracle          = "c1882cb4-f69e-441e-8871-743ed636cad0"
	avengingHunterOracle          = "e0aa5e3d-17c0-4ac7-9d76-9ce579d1eb2b"
	goliathPaladinOracle          = "269a02ca-1425-4b32-81f3-440e1caadb17"
	underdarkExplorerOracle       = "7647682d-d491-4cb1-add7-b4d745994e62"
	stirringBardOracle            = "2583836d-7fd3-47b3-9b69-6814efe54a90"
	viciousBattleragerOracle      = "b8942c72-16ce-4f2f-875f-97a56e85a03e"
	passagewaySeerOracle          = "2d8ebb48-5f37-4b7e-b8ec-3296d9a039c7"
	feywildCaretakerOracle        = "f319b671-0388-48e8-ba98-04161a3e7f1a"
	imoenOracle                   = "280e05c5-7ab4-45bf-aa15-e594cec49ef3"
	safanaOracle                  = "afc1c443-93e3-4ef4-a404-1d9fc13e0b02"
	bloodboilSorcererOracle       = "acdcffa0-87e5-48a2-93d2-572febfc9d52"
	undermountainAdventurerOracle = "2e5e0abd-38a6-46ca-b922-556db52f1332"
	dungeoneersPackOracle         = "c66976d3-57f1-4c48-ada0-d74081f1699a"
	lootDisputeOracle             = "7c885698-69f2-44b5-abfe-9a2833822794"
	exploreTheUnderdarkOracle     = "13bad22a-6ba3-4197-8658-e47049a1cf1f"
	whitePlumeAdventurerOracle    = "51f091a7-9b0b-4362-8e9c-c174752f369b"
	trailblazersTorchOracle       = "1cdac498-3c94-48d6-b94c-1d98697370ea"
)

// initZone is the zone the card with this ID is in, "" when none.
func initZone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var k game.ZoneKind
	g.WithWriteLock(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			k = z.Kind
		}
	})
	return k
}

// initCompleteADungeon records a completed Undercity for `p`, as a
// state a test starts from.
func initCompleteADungeon(p *game.Player) {
	p.CompletedDungeons = append(p.CompletedDungeons, game.UndercityKey)
}

// --- From the Catacombs ----------------------------------------------

// TestFromTheCatacombsReanimatesAnOpponentsCreatureAndTakesTheInitiative:
// any graveyard, under your control, with a corpse counter; you take the
// initiative and venture; and the creature is exiled when it would die.
func TestFromTheCatacombsReanimatesAnOpponentsCreatureAndTakesTheInitiative(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	dead := pushGraveyardCardTyped(them, "Their Giant", "Creature — Giant")

	castCatalogSpell(t, g, "From the Catacombs", "Sorcery", fromTheCatacombsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	initSettle(t, g)
	c := findBattlefieldCardForTest(g, dead)
	if c == nil {
		t.Fatalf("the creature did not enter the battlefield")
	}
	if c.Controller != me.ID {
		t.Errorf("controller = %v, want the caster %v", c.Controller, me.ID)
	}
	if c.Counters["corpse"] != 1 {
		t.Errorf("corpse counters = %d, want 1", c.Counters["corpse"])
	}
	if g.Initiative != me.ID {
		t.Fatalf("initiative = %v, want the caster", g.Initiative)
	}
	if room := initRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("room = %q, want secret-entrance", room)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(dead); err != nil {
			t.Errorf("destroy: %v", err)
		}
	})
	if z := initZone(g, dead); z != game.ZoneExile {
		t.Fatalf("the reanimated creature went to %q, want exile", z)
	}
}

// TestFromTheCatacombsEscapes: cast from the graveyard for its escape
// cost, exiling five other cards.
func TestFromTheCatacombsEscapes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := pushGraveyardCardTyped(me, "My Giant", "Creature — Giant")
	id := seedGraveyardCard(t, g, "From the Catacombs", "Sorcery", fromTheCatacombsOracle)
	pay := seedGraveyardFodder(t, g, 5)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	for _, f := range pay {
		if z := initZone(g, f); z != game.ZoneExile {
			t.Fatalf("escape fodder in %q, want exile", z)
		}
	}
	initSettle(t, g)
	if findBattlefieldCardForTest(g, dead) == nil {
		t.Fatalf("the escaped spell did not reanimate")
	}
	if g.Initiative != me.ID {
		t.Fatalf("initiative = %v, want the caster", g.Initiative)
	}
	if z := initZone(g, id); z != game.ZoneGraveyard {
		t.Errorf("the escaped sorcery is in %q, want the graveyard (escape exiles only its cost)", z)
	}
}

// --- the ETB family --------------------------------------------------

// TestInitiativeCreaturesTakeTheInitiativeOnEntry covers every creature
// whose text is a keyword and "When this creature enters, you take the
// initiative" — and the keyword.
func TestInitiativeCreaturesTakeTheInitiativeOnEntry(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Aarakocra Sneak", "Creature — Bird Rogue", aarakocraSneakOracle, "flying"},
		{"Avenging Hunter", "Creature — Dragon Ranger", avengingHunterOracle, "trample"},
		{"Goliath Paladin", "Creature — Giant Knight", goliathPaladinOracle, "vigilance"},
		{"Underdark Explorer", "Creature — Lizard Warrior", underdarkExplorerOracle, "menace"},
		{"Stirring Bard", "Creature — Dragon Bard", stirringBardOracle, "defender"},
		{"Undermountain Adventurer", "Creature — Giant Warrior", undermountainAdventurerOracle, "vigilance"},
		{"Passageway Seer", "Creature — Tiefling Warlock", passagewaySeerOracle, "lifelink"},
		{"Vicious Battlerager", "Creature — Dwarf Barbarian", viciousBattleragerOracle, ""},
		{"Bloodboil Sorcerer", "Creature — Human Shaman Sorcerer", bloodboilSorcererOracle, ""},
		{"Feywild Caretaker", "Creature — Orc Wizard", feywildCaretakerOracle, ""},
		{"White Plume Adventurer", "Creature — Orc Cleric", whitePlumeAdventurerOracle, ""},
		{"Trailblazer's Torch", "Artifact — Equipment", trailblazersTorchOracle, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := castAndResolveCreature(t, g, tc.name, tc.typeLine, tc.oracle)
			initSettle(t, g)
			if g.Initiative != me.ID {
				t.Fatalf("initiative = %v, want the controller", g.Initiative)
			}
			if room := initRoom(g, me.ID); room != "secret-entrance" {
				t.Fatalf("room = %q, want secret-entrance", room)
			}
			if tc.keyword != "" && !hasAbility(effectiveAbilities(t, g, id), tc.keyword) {
				t.Errorf("%s does not have %s", tc.name, tc.keyword)
			}
		})
	}
}

// --- Stirring Bard ---------------------------------------------------

func TestStirringBardGivesMenaceAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bard := pushPermanentForTest(g, me.ID, "Stirring Bard", stirringBardOracle, "Creature — Dragon Bard")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, bard, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	initSettle(t, g)
	got := effectiveAbilities(t, g, bear)
	if !hasAbility(got, "menace") || !hasAbility(got, "haste") {
		t.Fatalf("Bear has %v, want menace and haste", got)
	}
}

// --- Vicious Battlerager ---------------------------------------------

// Two blockers are two triggers: each blocker's controller loses 5.
func TestViciousBattleragerCostsEachBlockersControllerFiveLife(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	rager := pushPermanentForTest(g, me.ID, "Vicious Battlerager", viciousBattleragerOracle, "Creature — Dwarf Barbarian")
	findBattlefieldCardForTest(g, rager).Power, findBattlefieldCardForTest(g, rager).Toughness = 1, 5
	a := pushVanillaCreature(g, them.ID, "Wall A", 0, 4)
	b := pushVanillaCreature(g, them.ID, "Wall B", 0, 4)
	before := them.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(rager, them.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(w, rager); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
	initSettle(t, g)
	if them.Life != before-10 {
		t.Fatalf("their life %d, want %d (5 per blocker)", them.Life, before-10)
	}
}

// --- end-step payoffs -------------------------------------------------

// TestPassagewaySeerGrowsOnlyWithTheInitiative: the intervening "if".
func TestPassagewaySeerGrowsOnlyWithTheInitiative(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	seer := pushPermanentForTest(g, me.ID, "Passageway Seer", passagewaySeerOracle, "Creature — Tiefling Warlock")
	initVenture(t, g, them.ID)
	advanceToEndStepOf(t, g, 0)
	// CR 603.4: without the initiative the ability does not trigger at
	// all — nothing waits to go on the stack.
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == seer {
			t.Fatalf("the Seer triggered without the initiative")
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == seer {
			t.Fatalf("the Seer triggered without the initiative")
		}
	}
	initSettle(t, g)
	if n := findBattlefieldCardForTest(g, seer).Counters["+1/+1"]; n != 0 {
		t.Fatalf("grew %d without the initiative", n)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	seer2 := pushPermanentForTest(g2, me2.ID, "Passageway Seer", passagewaySeerOracle, "Creature — Tiefling Warlock")
	initVenture(t, g2, me2.ID)
	advanceToEndStepOf(t, g2, 0)
	initSettle(t, g2)
	if n := findBattlefieldCardForTest(g2, seer2).Counters["+1/+1"]; n != 1 {
		t.Fatalf("+1/+1 counters = %d with the initiative, want 1", n)
	}
	_ = me
}

func TestFeywildCaretakerMakesAFaerieDragonWithTheInitiative(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Feywild Caretaker", feywildCaretakerOracle, "Creature — Orc Wizard")
	initVenture(t, g, me.ID)
	advanceToEndStepOf(t, g, 0)
	initSettle(t, g)
	if n := monTokens(g, me.ID, "Faerie Dragon"); n != 1 {
		t.Fatalf("Faerie Dragons = %d, want 1", n)
	}
}

// TestImoenDrawsOneOrTwo: one card with the initiative, two once a
// dungeon has been completed.
func TestImoenDrawsOneOrTwo(t *testing.T) {
	for _, tc := range []struct {
		completed bool
		want      int
	}{{false, 1}, {true, 2}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		pushPermanentForTest(g, me.ID, "Imoen, Mystic Trickster", imoenOracle, "Legendary Creature — Human Rogue Wizard")
		if tc.completed {
			initCompleteADungeon(me)
		}
		initVenture(t, g, me.ID)
		advanceToEndStepOf(t, g, 0)
		before := me.Hand.Size()
		initSettle(t, g)
		if got := me.Hand.Size() - before; got != tc.want {
			t.Errorf("completed=%v: drew %d, want %d", tc.completed, got, tc.want)
		}
	}
}

func TestSafanaMakesOneTreasureOrThree(t *testing.T) {
	for _, tc := range []struct {
		completed bool
		want      int
	}{{false, 1}, {true, 3}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		pushPermanentForTest(g, me.ID, "Safana, Calimport Cutthroat", safanaOracle, "Legendary Creature — Human Rogue")
		if tc.completed {
			initCompleteADungeon(me)
		}
		initVenture(t, g, me.ID)
		advanceToEndStepOf(t, g, 0)
		initSettle(t, g)
		if got := monTokens(g, me.ID, "Treasure"); got != tc.want {
			t.Errorf("completed=%v: %d Treasures, want %d", tc.completed, got, tc.want)
		}
	}
}

// --- Bloodboil Sorcerer ----------------------------------------------

func TestBloodboilSorcererSacrificesToGoad(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	sorc := pushPermanentForTest(g, me.ID, "Bloodboil Sorcerer", bloodboilSorcererOracle, "Creature — Human Shaman Sorcerer")
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	theirs := pushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, sorc, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if z := initZone(g, fodder); z != game.ZoneGraveyard {
		t.Fatalf("the sacrifice went to %q", z)
	}
	initSettle(t, g)
	if !findBattlefieldCardForTest(g, theirs).IsGoadedBy(me.ID) {
		t.Fatalf("their creature is not goaded")
	}
}

// --- Undermountain Adventurer ----------------------------------------

func TestUndermountainAdventurerAddsTwoOrSixGreen(t *testing.T) {
	for _, tc := range []struct {
		completed bool
		want      int
	}{{false, 2}, {true, 6}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := pushPermanentForTest(g, me.ID, "Undermountain Adventurer", undermountainAdventurerOracle, "Creature — Giant Warrior")
		if tc.completed {
			initCompleteADungeon(me)
		}
		tapForMana(t, g, me.ID, id)
		if got := len(poolColors(me)); got != tc.want {
			t.Errorf("completed=%v: %d mana, want %d", tc.completed, got, tc.want)
		}
	}
}

// --- Dungeoneer's Pack -----------------------------------------------

func TestDungeoneersPackEntersTappedAndPaysOut(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pack := castAndResolveCreature(t, g, "Dungeoneer's Pack", "Artifact", dungeoneersPackOracle)
	if c := findBattlefieldCardForTest(g, pack); c == nil || !c.Tapped {
		t.Fatalf("the Pack did not enter tapped")
	}
	findBattlefieldCardForTest(g, pack).Tapped = false
	lifeBefore, handBefore := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, pack, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	initSettle(t, g)
	if g.Initiative != me.ID {
		t.Errorf("initiative = %v, want the activator", g.Initiative)
	}
	if me.Life != lifeBefore+3 {
		t.Errorf("life %d, want %d", me.Life, lifeBefore+3)
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), handBefore+1)
	}
	if monTokens(g, me.ID, "Treasure") != 1 {
		t.Errorf("no Treasure")
	}
	if initRoom(g, me.ID) != "secret-entrance" {
		t.Errorf("no venture")
	}
}

// --- Loot Dispute -----------------------------------------------------

func TestLootDisputeEntersWithTheInitiativeAndATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Loot Dispute", "Enchantment", lootDisputeOracle)
	initSettle(t, g)
	if g.Initiative != me.ID || monTokens(g, me.ID, "Treasure") != 1 {
		t.Fatalf("initiative %v, Treasures %d", g.Initiative, monTokens(g, me.ID, "Treasure"))
	}
}

// Attacking the holder with two creatures is one Treasure.
func TestLootDisputeTreasureForAttackingTheInitiative(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Loot Dispute", lootDisputeOracle, "Enchantment")
	initVenture(t, g, them.ID)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, them.ID); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	initSettle(t, g)
	if n := monTokens(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Treasures = %d, want 1 for the attack", n)
	}
}

// Attacking a player without the initiative makes nothing.
func TestLootDisputeNothingForAttackingSomeoneElse(t *testing.T) {
	g := newCatalogGame(t)
	me, them, other := g.Seats[0], g.Seats[1], g.Seats[2]
	pushPermanentForTest(g, me.ID, "Loot Dispute", lootDisputeOracle, "Enchantment")
	initVenture(t, g, other.ID)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, them.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	initSettle(t, g)
	if n := monTokens(g, me.ID, "Treasure"); n != 0 {
		t.Fatalf("Treasures = %d, want 0", n)
	}
}

// Loud Ruckus: completing the Undercity makes a 5/5 Dragon.
func TestLootDisputeLoudRuckusOnCompletingADungeon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Loot Dispute", lootDisputeOracle, "Enchantment")
	initVenture(t, g, me.ID)
	initVenture(t, g, me.ID, "lost-well")
	initVenture(t, g, me.ID, "stash")
	initVenture(t, g, me.ID)
	if monTokens(g, me.ID, "Dragon") != 0 {
		t.Fatalf("a Dragon before the dungeon was completed")
	}
	initVenture(t, g, me.ID)
	if n := monTokens(g, me.ID, "Dragon"); n != 1 {
		t.Fatalf("Dragons = %d after completing the Undercity, want 1", n)
	}
}

// --- Explore the Underdark -------------------------------------------

func TestExploreTheUnderdarkFetchesTappedLandsThenTakesTheInitiative(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	gate := pushLibraryCardForTest(me, game.Card{Name: "Azorius Guildgate", TypeLine: "Land — Gate"})
	castCatalogSpell(t, g, "Explore the Underdark", "Sorcery", exploreTheUnderdarkOracle, nil)
	for i := 0; i < 8 && initPending(g, game.PendingChoiceSearchLibrary, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := initPending(g, game.PendingChoiceSearchLibrary, me.ID)
	if c == nil {
		t.Fatalf("no search prompt")
	}
	if g.Initiative == me.ID {
		t.Fatalf("took the initiative before the search")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{forest, gate}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	initSettle(t, g)
	for _, id := range []uuid.UUID{forest, gate} {
		l := findBattlefieldCardForTest(g, id)
		if l == nil || !l.Tapped {
			t.Errorf("land %v not on the battlefield tapped", id)
		}
	}
	if g.Initiative != me.ID {
		t.Fatalf("initiative = %v, want the caster", g.Initiative)
	}
}

// --- White Plume Adventurer ------------------------------------------

func TestWhitePlumeUntapsOneOrAllOnAnOpponentsUpkeep(t *testing.T) {
	for _, tc := range []struct {
		completed bool
		want      int
	}{{false, 1}, {true, 2}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		plume := pushPermanentForTest(g, me.ID, "White Plume Adventurer", whitePlumeAdventurerOracle, "Creature — Orc Cleric")
		a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
		if tc.completed {
			initCompleteADungeon(me)
		}
		advanceToUpkeepOf(t, g, 1)
		for _, id := range []uuid.UUID{plume, a} {
			findBattlefieldCardForTest(g, id).Tapped = true
		}
		// Settle, answering the one-creature choice with the Bear.
		for i := 0; i < 16; i++ {
			if c := initPending(g, game.PendingChoiceOwnPermanents, me.ID); c != nil {
				if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{a}); err != nil {
					t.Fatalf("choose: %v", err)
				}
				continue
			}
			if c := initPending(g, game.PendingChoiceChooseCards, me.ID); c != nil {
				if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a}); err != nil {
					t.Fatalf("choose: %v", err)
				}
				continue
			}
			if stackFullyEmpty(g) {
				break
			}
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		untapped := 0
		for _, id := range []uuid.UUID{plume, a} {
			if !findBattlefieldCardForTest(g, id).Tapped {
				untapped++
			}
		}
		if untapped != tc.want {
			t.Errorf("completed=%v: %d untapped, want %d", tc.completed, untapped, tc.want)
		}
	}
}

// --- Trailblazer's Torch ---------------------------------------------

func TestTrailblazersTorchBurnsEachBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	torch := pushPermanentForTest(g, me.ID, "Trailblazer's Torch", trailblazersTorchOracle, "Artifact — Equipment")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, torch, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	initSettle(t, g)
	a := pushVanillaCreature(g, them.ID, "Wall A", 0, 5)
	b := pushVanillaCreature(g, them.ID, "Wall B", 0, 5)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, them.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(w, bear); err != nil {
			t.Fatal(err)
		}
	}
	lockInBlocks(t, g)
	initSettle(t, g)
	for _, w := range []uuid.UUID{a, b} {
		if got := findBattlefieldCardForTest(g, w).DamageMarked; got != 2 {
			t.Errorf("blocker damage = %d, want 2", got)
		}
	}
}
