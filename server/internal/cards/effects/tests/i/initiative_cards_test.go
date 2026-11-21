package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- From the Catacombs ----------------------------------------------

// TestFromTheCatacombsReanimatesAnOpponentsCreatureAndTakesTheInitiative:
// any graveyard, under your control, with a corpse counter; you take the
// initiative and venture; and the creature is exiled when it would die.
func TestFromTheCatacombsReanimatesAnOpponentsCreatureAndTakesTheInitiative(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	dead := effectstest.PushGraveyardCardTyped(them, "Their Giant", "Creature — Giant")

	effectstest.CastCatalogSpell(t, g, "From the Catacombs", "Sorcery", effectstest.FromTheCatacombsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	effectstest.InitSettle(t, g)
	c := effectstest.FindBattlefieldCardForTest(g, dead)
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
	if room := effectstest.InitRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("room = %q, want secret-entrance", room)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(dead); err != nil {
			t.Errorf("destroy: %v", err)
		}
	})
	if z := effectstest.InitZone(g, dead); z != game.ZoneExile {
		t.Fatalf("the reanimated creature went to %q, want exile", z)
	}
}

// TestFromTheCatacombsEscapes: cast from the graveyard for its escape
// cost, exiling five other cards.
func TestFromTheCatacombsEscapes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.PushGraveyardCardTyped(me, "My Giant", "Creature — Giant")
	id := effectstest.SeedGraveyardCard(t, g, "From the Catacombs", "Sorcery", effectstest.FromTheCatacombsOracle)
	pay := effectstest.SeedGraveyardFodder(t, g, 5)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	for _, f := range pay {
		if z := effectstest.InitZone(g, f); z != game.ZoneExile {
			t.Fatalf("escape fodder in %q, want exile", z)
		}
	}
	effectstest.InitSettle(t, g)
	if effectstest.FindBattlefieldCardForTest(g, dead) == nil {
		t.Fatalf("the escaped spell did not reanimate")
	}
	if g.Initiative != me.ID {
		t.Fatalf("initiative = %v, want the caster", g.Initiative)
	}
	if z := effectstest.InitZone(g, id); z != game.ZoneGraveyard {
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
		{"Aarakocra Sneak", "Creature — Bird Rogue", effectstest.AarakocraSneakOracle, "flying"},
		{"Avenging Hunter", "Creature — Dragon Ranger", effectstest.AvengingHunterOracle, "trample"},
		{"Goliath Paladin", "Creature — Giant Knight", effectstest.GoliathPaladinOracle, "vigilance"},
		{"Underdark Explorer", "Creature — Lizard Warrior", effectstest.UnderdarkExplorerOracle, "menace"},
		{"Stirring Bard", "Creature — Dragon Bard", effectstest.StirringBardOracle, "defender"},
		{"Undermountain Adventurer", "Creature — Giant Warrior", effectstest.UndermountainAdventurerOracle, "vigilance"},
		{"Passageway Seer", "Creature — Tiefling Warlock", effectstest.PassagewaySeerOracle, "lifelink"},
		{"Vicious Battlerager", "Creature — Dwarf Barbarian", effectstest.ViciousBattleragerOracle, ""},
		{"Bloodboil Sorcerer", "Creature — Human Shaman Sorcerer", effectstest.BloodboilSorcererOracle, ""},
		{"Feywild Caretaker", "Creature — Orc Wizard", effectstest.FeywildCaretakerOracle, ""},
		{"White Plume Adventurer", "Creature — Orc Cleric", effectstest.WhitePlumeAdventurerOracle, ""},
		{"Trailblazer's Torch", "Artifact — Equipment", effectstest.TrailblazersTorchOracle, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.CastAndResolveCreature(t, g, tc.name, tc.typeLine, tc.oracle)
			effectstest.InitSettle(t, g)
			if g.Initiative != me.ID {
				t.Fatalf("initiative = %v, want the controller", g.Initiative)
			}
			if room := effectstest.InitRoom(g, me.ID); room != "secret-entrance" {
				t.Fatalf("room = %q, want secret-entrance", room)
			}
			if tc.keyword != "" && !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), tc.keyword) {
				t.Errorf("%s does not have %s", tc.name, tc.keyword)
			}
		})
	}
}

// --- Stirring Bard ---------------------------------------------------

func TestStirringBardGivesMenaceAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bard := effectstest.PushPermanentForTest(g, me.ID, "Stirring Bard", effectstest.StirringBardOracle, "Creature — Dragon Bard")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, bard, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.InitSettle(t, g)
	got := effectstest.EffectiveAbilities(t, g, bear)
	if !effectstest.HasAbility(got, "menace") || !effectstest.HasAbility(got, "haste") {
		t.Fatalf("Bear has %v, want menace and haste", got)
	}
}

// --- Vicious Battlerager ---------------------------------------------

// Two blockers are two triggers: each blocker's controller loses 5.
func TestViciousBattleragerCostsEachBlockersControllerFiveLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	rager := effectstest.PushPermanentForTest(g, me.ID, "Vicious Battlerager", effectstest.ViciousBattleragerOracle, "Creature — Dwarf Barbarian")
	effectstest.FindBattlefieldCardForTest(g, rager).Power, effectstest.FindBattlefieldCardForTest(g, rager).Toughness = 1, 5
	a := effectstest.PushVanillaCreature(g, them.ID, "Wall A", 0, 4)
	b := effectstest.PushVanillaCreature(g, them.ID, "Wall B", 0, 4)
	before := them.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(rager, them.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(w, rager); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.InitSettle(t, g)
	if them.Life != before-10 {
		t.Fatalf("their life %d, want %d (5 per blocker)", them.Life, before-10)
	}
}

// --- end-step payoffs -------------------------------------------------

// TestPassagewaySeerGrowsOnlyWithTheInitiative: the intervening "if".
func TestPassagewaySeerGrowsOnlyWithTheInitiative(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	seer := effectstest.PushPermanentForTest(g, me.ID, "Passageway Seer", effectstest.PassagewaySeerOracle, "Creature — Tiefling Warlock")
	effectstest.InitVenture(t, g, them.ID)
	effectstest.AdvanceToEndStepOf(t, g, 0)
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
	effectstest.InitSettle(t, g)
	if n := effectstest.FindBattlefieldCardForTest(g, seer).Counters["+1/+1"]; n != 0 {
		t.Fatalf("grew %d without the initiative", n)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	seer2 := effectstest.PushPermanentForTest(g2, me2.ID, "Passageway Seer", effectstest.PassagewaySeerOracle, "Creature — Tiefling Warlock")
	effectstest.InitVenture(t, g2, me2.ID)
	effectstest.AdvanceToEndStepOf(t, g2, 0)
	effectstest.InitSettle(t, g2)
	if n := effectstest.FindBattlefieldCardForTest(g2, seer2).Counters["+1/+1"]; n != 1 {
		t.Fatalf("+1/+1 counters = %d with the initiative, want 1", n)
	}
	_ = me
}

func TestFeywildCaretakerMakesAFaerieDragonWithTheInitiative(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Feywild Caretaker", effectstest.FeywildCaretakerOracle, "Creature — Orc Wizard")
	effectstest.InitVenture(t, g, me.ID)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.InitSettle(t, g)
	if n := effectstest.MonTokens(g, me.ID, "Faerie Dragon"); n != 1 {
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushPermanentForTest(g, me.ID, "Imoen, Mystic Trickster", effectstest.ImoenOracle, "Legendary Creature — Human Rogue Wizard")
		if tc.completed {
			effectstest.InitCompleteADungeon(me)
		}
		effectstest.InitVenture(t, g, me.ID)
		effectstest.AdvanceToEndStepOf(t, g, 0)
		before := me.Hand.Size()
		effectstest.InitSettle(t, g)
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushPermanentForTest(g, me.ID, "Safana, Calimport Cutthroat", effectstest.SafanaOracle, "Legendary Creature — Human Rogue")
		if tc.completed {
			effectstest.InitCompleteADungeon(me)
		}
		effectstest.InitVenture(t, g, me.ID)
		effectstest.AdvanceToEndStepOf(t, g, 0)
		effectstest.InitSettle(t, g)
		if got := effectstest.MonTokens(g, me.ID, "Treasure"); got != tc.want {
			t.Errorf("completed=%v: %d Treasures, want %d", tc.completed, got, tc.want)
		}
	}
}

// --- Bloodboil Sorcerer ----------------------------------------------

func TestBloodboilSorcererSacrificesToGoad(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	sorc := effectstest.PushPermanentForTest(g, me.ID, "Bloodboil Sorcerer", effectstest.BloodboilSorcererOracle, "Creature — Human Shaman Sorcerer")
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, sorc, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if z := effectstest.InitZone(g, fodder); z != game.ZoneGraveyard {
		t.Fatalf("the sacrifice went to %q", z)
	}
	effectstest.InitSettle(t, g)
	if !effectstest.FindBattlefieldCardForTest(g, theirs).IsGoadedBy(me.ID) {
		t.Fatalf("their creature is not goaded")
	}
}

// --- Undermountain Adventurer ----------------------------------------

func TestUndermountainAdventurerAddsTwoOrSixGreen(t *testing.T) {
	for _, tc := range []struct {
		completed bool
		want      int
	}{{false, 2}, {true, 6}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.PushPermanentForTest(g, me.ID, "Undermountain Adventurer", effectstest.UndermountainAdventurerOracle, "Creature — Giant Warrior")
		if tc.completed {
			effectstest.InitCompleteADungeon(me)
		}
		effectstest.TapForMana(t, g, me.ID, id)
		if got := len(effectstest.PoolColors(me)); got != tc.want {
			t.Errorf("completed=%v: %d mana, want %d", tc.completed, got, tc.want)
		}
	}
}

// --- Dungeoneer's Pack -----------------------------------------------

func TestDungeoneersPackEntersTappedAndPaysOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pack := effectstest.CastAndResolveCreature(t, g, "Dungeoneer's Pack", "Artifact", effectstest.DungeoneersPackOracle)
	if c := effectstest.FindBattlefieldCardForTest(g, pack); c == nil || !c.Tapped {
		t.Fatalf("the Pack did not enter tapped")
	}
	effectstest.FindBattlefieldCardForTest(g, pack).Tapped = false
	lifeBefore, handBefore := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, pack, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.InitSettle(t, g)
	if g.Initiative != me.ID {
		t.Errorf("initiative = %v, want the activator", g.Initiative)
	}
	if me.Life != lifeBefore+3 {
		t.Errorf("life %d, want %d", me.Life, lifeBefore+3)
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), handBefore+1)
	}
	if effectstest.MonTokens(g, me.ID, "Treasure") != 1 {
		t.Errorf("no Treasure")
	}
	if effectstest.InitRoom(g, me.ID) != "secret-entrance" {
		t.Errorf("no venture")
	}
}

// --- Loot Dispute -----------------------------------------------------

func TestLootDisputeEntersWithTheInitiativeAndATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastAndResolveCreature(t, g, "Loot Dispute", "Enchantment", effectstest.LootDisputeOracle)
	effectstest.InitSettle(t, g)
	if g.Initiative != me.ID || effectstest.MonTokens(g, me.ID, "Treasure") != 1 {
		t.Fatalf("initiative %v, Treasures %d", g.Initiative, effectstest.MonTokens(g, me.ID, "Treasure"))
	}
}

// Attacking the holder with two creatures is one Treasure.
func TestLootDisputeTreasureForAttackingTheInitiative(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Loot Dispute", effectstest.LootDisputeOracle, "Enchantment")
	effectstest.InitVenture(t, g, them.ID)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, them.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.InitSettle(t, g)
	if n := effectstest.MonTokens(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Treasures = %d, want 1 for the attack", n)
	}
}

// Attacking a player without the initiative makes nothing.
func TestLootDisputeNothingForAttackingSomeoneElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushPermanentForTest(g, me.ID, "Loot Dispute", effectstest.LootDisputeOracle, "Enchantment")
	effectstest.InitVenture(t, g, other.ID)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, them.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.InitSettle(t, g)
	if n := effectstest.MonTokens(g, me.ID, "Treasure"); n != 0 {
		t.Fatalf("Treasures = %d, want 0", n)
	}
}

// Loud Ruckus: completing the Undercity makes a 5/5 Dragon.
func TestLootDisputeLoudRuckusOnCompletingADungeon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Loot Dispute", effectstest.LootDisputeOracle, "Enchantment")
	effectstest.InitVenture(t, g, me.ID)
	effectstest.InitVenture(t, g, me.ID, "lost-well")
	effectstest.InitVenture(t, g, me.ID, "stash")
	effectstest.InitVenture(t, g, me.ID)
	if effectstest.MonTokens(g, me.ID, "Dragon") != 0 {
		t.Fatalf("a Dragon before the dungeon was completed")
	}
	effectstest.InitVenture(t, g, me.ID)
	if n := effectstest.MonTokens(g, me.ID, "Dragon"); n != 1 {
		t.Fatalf("Dragons = %d after completing the Undercity, want 1", n)
	}
}

// --- Explore the Underdark -------------------------------------------

func TestExploreTheUnderdarkFetchesTappedLandsThenTakesTheInitiative(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	gate := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Azorius Guildgate", TypeLine: "Land — Gate"})
	effectstest.CastCatalogSpell(t, g, "Explore the Underdark", "Sorcery", effectstest.ExploreTheUnderdarkOracle, nil)
	for i := 0; i < 8 && effectstest.InitPending(g, game.PendingChoiceSearchLibrary, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := effectstest.InitPending(g, game.PendingChoiceSearchLibrary, me.ID)
	if c == nil {
		t.Fatalf("no search prompt")
	}
	if g.Initiative == me.ID {
		t.Fatalf("took the initiative before the search")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{forest, gate}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	effectstest.InitSettle(t, g)
	for _, id := range []uuid.UUID{forest, gate} {
		l := effectstest.FindBattlefieldCardForTest(g, id)
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		plume := effectstest.PushPermanentForTest(g, me.ID, "White Plume Adventurer", effectstest.WhitePlumeAdventurerOracle, "Creature — Orc Cleric")
		a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
		if tc.completed {
			effectstest.InitCompleteADungeon(me)
		}
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		for _, id := range []uuid.UUID{plume, a} {
			effectstest.FindBattlefieldCardForTest(g, id).Tapped = true
		}
		// Settle, answering the one-creature choice with the Bear.
		for i := 0; i < 16; i++ {
			if c := effectstest.InitPending(g, game.PendingChoiceOwnPermanents, me.ID); c != nil {
				if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{a}); err != nil {
					t.Fatalf("choose: %v", err)
				}
				continue
			}
			if c := effectstest.InitPending(g, game.PendingChoiceChooseCards, me.ID); c != nil {
				if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a}); err != nil {
					t.Fatalf("choose: %v", err)
				}
				continue
			}
			if effectstest.StackFullyEmpty(g) {
				break
			}
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		untapped := 0
		for _, id := range []uuid.UUID{plume, a} {
			if !effectstest.FindBattlefieldCardForTest(g, id).Tapped {
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
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	torch := effectstest.PushPermanentForTest(g, me.ID, "Trailblazer's Torch", effectstest.TrailblazersTorchOracle, "Artifact — Equipment")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, torch, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.InitSettle(t, g)
	a := effectstest.PushVanillaCreature(g, them.ID, "Wall A", 0, 5)
	b := effectstest.PushVanillaCreature(g, them.ID, "Wall B", 0, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, them.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(w, bear); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
	for _, w := range []uuid.UUID{a, b} {
		if got := effectstest.FindBattlefieldCardForTest(g, w).DamageMarked; got != 2 {
			t.Errorf("blocker damage = %d, want 2", got)
		}
	}
}
