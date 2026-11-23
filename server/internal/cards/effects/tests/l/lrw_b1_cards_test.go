package testsl

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lrw_b1_cards_test.go — Lorwyn, batch 1 (#5593). Each test plays the real
// catalog entry through the engine and asserts what its text says.

const (
	lrwB1LignifyOracle             = "63625b26-3a19-4c3c-b18a-91247bb2a40c"
	lrwB1ElvishHarbingerOracle     = "28ddc23c-5b0b-4747-9a33-6845acf27bec"
	lrwB1JaceBelerenOracle         = "cc2b6e58-2d6f-495e-95f3-9496105c4cba"
	lrwB1LilianaVessOracle         = "a6c05941-2cfb-4dac-a7c0-ab808187eb7c"
	lrwB1AmoeboidChangelingOracle  = "68865793-daeb-4c55-80cb-12853f9f96c4"
	lrwB1GiltLeafPalaceOracle      = "85573a3d-2993-491a-8f8d-bbdb844fa84e"
	lrwB1DoranOracle               = "a46d96d9-8e59-4777-8208-7730b9e33240"
	lrwB1GarrukWildspeakerOracle   = "186b4def-4fff-4a2b-bcbf-5318b0ac5fa9"
	lrwB1ProwessOfTheFairOracle    = "31324235-8a2b-4ce4-b0d4-6d947489b930"
	lrwB1FlamekinHarbingerOracle   = "d6585e30-4ca0-4701-b274-b24f3508dd97"
	lrwB1BoggartShenanigansOracle  = "997d2f38-5a50-451e-bd48-e36dcb24d967"
	lrwB1FaerieHarbingerOracle     = "a171a01c-43f7-488a-870d-b1a9bcc501d1"
	lrwB1VividMarshOracle          = "20b32052-f66f-4eb8-b56e-00d531907f19"
	lrwB1MerrowCommerceOracle      = "a931b609-1486-447b-9479-61a45d64f2ef"
	lrwB1FamiliarsRuseOracle       = "3b6d773b-e2f7-4e74-9e77-f2e0d369a31e"
	lrwB1VividCragOracle           = "d8e2efe0-33a4-4303-9e83-ac42ea5df8cb"
	lrwB1SoulbrightFlamekinOracle  = "7ea8426a-3b66-4fd0-9863-2b68d379671f"
	lrwB1SeedguideAshOracle        = "7ff007cb-c994-4643-a6d4-c58e030514e1"
	lrwB1TimberProtectorOracle     = "9bc8a69b-6065-4ef7-a580-adb1e7c72ef4"
	lrwB1TreefolkHarbingerOracle   = "98c387bd-c9ba-464a-8380-839a5604a37f"
	lrwB1MerrowHarbingerOracle     = "035ee753-232f-4ba2-afbf-8d1eb162e6f8"
	lrwB1EyeblightsEndingOracle    = "9f3e1eef-31db-4e1a-b5a9-ac835f6a3eab"
	lrwB1AuntiesHovelOracle        = "245469ff-72b6-4846-8a82-a1d29f4d09bb"
	lrwB1MadAuntieOracle           = "b582b696-17fa-4eea-8124-293f9bd3a208"
	lrwB1AvianChangelingOracle     = "767c1f97-03de-474d-9112-e7521ba3b3f3"
	lrwB1BoggartHarbingerOracle    = "56508ae7-6013-4149-8935-96b8d5091baa"
	lrwB1HerbalPoulticeOracle      = "4a8e1f8e-3740-4ff3-b7fa-390ea278fb95"
	lrwB1SkeletalChangelingOracle  = "b9d388ca-1a63-4a11-bf33-9f7b87305a52"
	lrwB1ShimmeringGrottoOracle    = "bae49475-fe01-400b-8959-f0dde959577c"
	lrwB1WortOracle                = "6f477e2e-220e-48d9-9236-9564478e796a"
	lrwB1LeafGilderOracle          = "61324e37-4b79-4325-bf46-621b4270afe2"
	lrwB1AethersnipeOracle         = "3e0be003-4f42-4917-b85d-46baf2afd734"
	lrwB1AncientAmphitheaterOracle = "7211221d-d4d8-4bbe-9d2a-b82e005bfe8a"
	lrwB1SunriseSovereignOracle    = "c9abbb90-e3ae-4e7c-87ca-06ebebf65edb"
	lrwB1GiantHarbingerOracle      = "2445e58b-87ed-4ab2-8209-a5e1f566fba7"
	lrwB1WanderwineHubOracle       = "c3b46bd6-b3ef-452d-a916-995c44f1da07"
	lrwB1WizenedCennOracle         = "b5905ef3-6565-462f-a015-2bb5b5b09eff"
	lrwB1KithkinHarbingerOracle    = "a3d68848-e8cc-497c-951b-e7111f6d5142"
	lrwB1TarfireOracle             = "f07b821c-bbc1-4def-b009-6c03689478d0"
	lrwB1GaddockTeegOracle         = "98dd6d43-cc56-4656-9e0c-240d0679b404"
)

func TestLrwB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lrwB1LignifyOracle:             "Lignify",
		lrwB1ElvishHarbingerOracle:     "Elvish Harbinger",
		lrwB1JaceBelerenOracle:         "Jace Beleren",
		lrwB1LilianaVessOracle:         "Liliana Vess",
		lrwB1AmoeboidChangelingOracle:  "Amoeboid Changeling",
		lrwB1GiltLeafPalaceOracle:      "Gilt-Leaf Palace",
		lrwB1DoranOracle:               "Doran, the Siege Tower",
		lrwB1GarrukWildspeakerOracle:   "Garruk Wildspeaker",
		lrwB1ProwessOfTheFairOracle:    "Prowess of the Fair",
		lrwB1FlamekinHarbingerOracle:   "Flamekin Harbinger",
		lrwB1BoggartShenanigansOracle:  "Boggart Shenanigans",
		lrwB1FaerieHarbingerOracle:     "Faerie Harbinger",
		lrwB1VividMarshOracle:          "Vivid Marsh",
		lrwB1MerrowCommerceOracle:      "Merrow Commerce",
		lrwB1FamiliarsRuseOracle:       "Familiar's Ruse",
		lrwB1VividCragOracle:           "Vivid Crag",
		lrwB1SoulbrightFlamekinOracle:  "Soulbright Flamekin",
		lrwB1SeedguideAshOracle:        "Seedguide Ash",
		lrwB1TimberProtectorOracle:     "Timber Protector",
		lrwB1TreefolkHarbingerOracle:   "Treefolk Harbinger",
		lrwB1MerrowHarbingerOracle:     "Merrow Harbinger",
		lrwB1EyeblightsEndingOracle:    "Eyeblight's Ending",
		lrwB1AuntiesHovelOracle:        "Auntie's Hovel",
		lrwB1MadAuntieOracle:           "Mad Auntie",
		lrwB1AvianChangelingOracle:     "Avian Changeling",
		lrwB1BoggartHarbingerOracle:    "Boggart Harbinger",
		lrwB1HerbalPoulticeOracle:      "Herbal Poultice",
		lrwB1SkeletalChangelingOracle:  "Skeletal Changeling",
		lrwB1ShimmeringGrottoOracle:    "Shimmering Grotto",
		lrwB1WortOracle:                "Wort, Boggart Auntie",
		lrwB1LeafGilderOracle:          "Leaf Gilder",
		lrwB1AethersnipeOracle:         "Aethersnipe",
		lrwB1AncientAmphitheaterOracle: "Ancient Amphitheater",
		lrwB1SunriseSovereignOracle:    "Sunrise Sovereign",
		lrwB1GiantHarbingerOracle:      "Giant Harbinger",
		lrwB1WanderwineHubOracle:       "Wanderwine Hub",
		lrwB1WizenedCennOracle:         "Wizened Cenn",
		lrwB1KithkinHarbingerOracle:    "Kithkin Harbinger",
		lrwB1TarfireOracle:             "Tarfire",
		lrwB1GaddockTeegOracle:         "Gaddock Teeg",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// lrwB1HasType reports whether a battlefield permanent has a subtype after
// the layers.
func lrwB1HasType(t *testing.T, g *game.Game, id uuid.UUID, sub string) bool {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.HasSubtype(sub)
}

// --- Harbingers ---------------------------------------------------------------

func TestLrwB1HarbingersTutorTheirTypeToTheTop(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		match, other           string
		keywords               []string
	}{
		{"Elvish Harbinger", "Creature — Elf Druid", lrwB1ElvishHarbingerOracle, 1, 2, "Kindred Instant — Elf", "Creature — Goblin", nil},
		{"Flamekin Harbinger", "Creature — Elemental Shaman", lrwB1FlamekinHarbingerOracle, 1, 1, "Creature — Elemental", "Creature — Elf", nil},
		{"Faerie Harbinger", "Creature — Faerie Wizard", lrwB1FaerieHarbingerOracle, 2, 2, "Creature — Faerie Rogue", "Creature — Merfolk", []string{"flash", "flying"}},
		{"Treefolk Harbinger", "Creature — Treefolk Druid", lrwB1TreefolkHarbingerOracle, 0, 3, "Land — Forest", "Land — Island", nil},
		{"Merrow Harbinger", "Creature — Merfolk Wizard", lrwB1MerrowHarbingerOracle, 2, 3, "Creature — Merfolk Soldier", "Creature — Faerie", []string{"islandwalk"}},
		{"Boggart Harbinger", "Creature — Goblin Shaman", lrwB1BoggartHarbingerOracle, 2, 1, "Creature — Shapeshifter", "Creature — Giant", nil},
		{"Giant Harbinger", "Creature — Giant Shaman", lrwB1GiantHarbingerOracle, 3, 4, "Creature — Giant Warrior", "Creature — Goblin", nil},
		{"Kithkin Harbinger", "Creature — Kithkin Wizard", lrwB1KithkinHarbingerOracle, 1, 3, "Kindred Enchantment — Kithkin", "Creature — Elf", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			card := game.Card{Name: "Match", TypeLine: tc.match}
			if tc.match == "Creature — Shapeshifter" {
				card.Keywords = []string{game.KeywordChangeling}
			}
			match := effectstest.PushLibraryCardForTest(me, card)
			effectstest.PushLibraryCardForTest(me, game.Card{Name: "Other", TypeLine: tc.other})
			id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt after saying yes")
			}
			if effectstest.SearchOptionNamed(g, c, "Other") != uuid.Nil {
				t.Errorf("a %s card was offered", tc.other)
			}
			effectstest.AnswerSearchByID(t, g, me.ID, match)
			effectstest.PassPriorityAroundTable(t, g)
			top, err := me.Library.Top()
			if err != nil || top.InstanceID != match {
				t.Fatalf("top of library is %q, want the tutored card", top.Name)
			}
			if !top.IsKnownTo(opp.ID) {
				t.Error("the tutored card was not revealed")
			}
			for _, kw := range tc.keywords {
				if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

func TestLrwB1HarbingerDeclinedDoesNotSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Match", TypeLine: "Creature — Giant"})
	effectstest.CastHoldCreature(t, g, "Giant Harbinger", "Creature — Giant Shaman", lrwB1GiantHarbingerOracle, 3, 4)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("a declined Harbinger still searched")
	}
}

func TestLrwB1ElvishHarbingerTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Elvish Harbinger", "Creature — Elf Druid", lrwB1ElvishHarbingerOracle, 1, 2)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if me.ManaPool.CountColor("U") != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("pool %v, want {U}", me.ManaPool)
	}
}

// --- lands --------------------------------------------------------------------

func TestLrwB1RevealLands(t *testing.T) {
	effectstest.S12RevealLandCase(t, "Gilt-Leaf Palace", lrwB1GiltLeafPalaceOracle, "Kindred Instant — Elf", "Creature — Goblin")
	effectstest.S12RevealLandCase(t, "Auntie's Hovel", lrwB1AuntiesHovelOracle, "Creature — Goblin Shaman", "Creature — Elf")
	effectstest.S12RevealLandCase(t, "Ancient Amphitheater", lrwB1AncientAmphitheaterOracle, "Creature — Giant Warrior", "Creature — Kithkin")
	effectstest.S12RevealLandCase(t, "Wanderwine Hub", lrwB1WanderwineHubOracle, "Creature — Merfolk Wizard", "Creature — Giant")
}

func TestLrwB1VividLandsEnterTappedWithTwoCharge(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Vivid Marsh", lrwB1VividMarshOracle},
		{"Vivid Crag", lrwB1VividCragOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			land := effectstest.PlayLandFromHand(t, g, tc.name, tc.oracle)
			if tapped, _ := effectstest.BattlefieldCardTapped(g, land); !tapped {
				t.Error("entered untapped")
			}
			if n := effectstest.CountersOn(g, land, game.CounterCharge); n != 2 {
				t.Errorf("%d charge counters, want 2", n)
			}
		})
	}
}

func TestLrwB1ShimmeringGrottoFiltersOneManaIntoAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	grotto := effectstest.PushCatalogPermanent(g, me.ID, "Shimmering Grotto", "Land", lrwB1ShimmeringGrottoOracle, false)
	if err := g.ActivateManaAbility(me.ID, grotto, 1, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("the filter was paid with an empty pool")
	}
	effectstest.ApaMana(me, "C")
	if err := g.ActivateManaAbility(me.ID, grotto, 1, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("filter: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.CountColor("B") != 1 {
		t.Errorf("pool %v, want {B}", me.ManaPool)
	}
}

func TestLrwB1LeafGilderTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Leaf Gilder", "Creature — Elf Druid", lrwB1LeafGilderOracle, 2, 1)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if me.ManaPool.CountColor("G") != 1 {
		t.Errorf("pool %v, want {G}", me.ManaPool)
	}
}

// --- planeswalkers ------------------------------------------------------------

func TestLrwB1JaceBeleren(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	jace := effectstest.PushWalkerForTest(g, me.ID, "Jace Beleren", lrwB1JaceBelerenOracle, 3)
	effectstest.AdvanceToMainOf(t, g, seat)
	hands := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		hands[p.ID] = p.Hand.Size()
	}
	effectstest.LrwB1Activate(t, g, me, jace, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Hand.Size() != hands[p.ID]+1 {
			t.Errorf("seat %s drew %d, want 1", p.Name, p.Hand.Size()-hands[p.ID])
		}
	}
	if effectstest.LoyaltyOf(g, jace) != 5 {
		t.Errorf("loyalty %d after +2, want 5", effectstest.LoyaltyOf(g, jace))
	}
	// Once a turn (CR 606.3).
	if err := g.ActivateCatalogAbility(me.ID, jace, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Error("a second loyalty ability was activated the same turn")
	}

	// −10 at ten loyalty: twenty cards from the target's library.
	g2 := effectstest.NewCatalogGame(t)
	seat2 := g2.Turn.ActiveSeat
	me2, opp2 := g2.Seats[seat2], g2.Seats[(seat2+1)%len(g2.Seats)]
	jace2 := effectstest.PushWalkerForTest(g2, me2.ID, "Jace Beleren", lrwB1JaceBelerenOracle, 10)
	effectstest.AdvanceToMainOf(t, g2, seat2)
	lib := opp2.Library.Size()
	gy := opp2.Graveyard.Size()
	effectstest.LrwB1Activate(t, g2, me2, jace2, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp2.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g2)
	milled := lib
	if milled > 20 {
		milled = 20
	}
	if opp2.Graveyard.Size() != gy+milled || opp2.Library.Size() != lib-milled {
		t.Errorf("milled %d, want %d", opp2.Graveyard.Size()-gy, milled)
	}
}

func TestLrwB1JaceBelerenMinusOneDrawsForTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	jace := effectstest.PushWalkerForTest(g, me.ID, "Jace Beleren", lrwB1JaceBelerenOracle, 3)
	effectstest.AdvanceToMainOf(t, g, seat)
	hand, mine := opp.Hand.Size(), me.Hand.Size()
	effectstest.LrwB1Activate(t, g, me, jace, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+1 || me.Hand.Size() != mine {
		t.Errorf("target drew %d, controller %d; want 1 and 0", opp.Hand.Size()-hand, me.Hand.Size()-mine)
	}
	if effectstest.LoyaltyOf(g, jace) != 2 {
		t.Errorf("loyalty %d, want 2", effectstest.LoyaltyOf(g, jace))
	}
}

func TestLrwB1LilianaVess(t *testing.T) {
	// −8: every creature card in every graveyard, under her controller.
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	liliana := effectstest.PushWalkerForTest(g, me.ID, "Liliana Vess", lrwB1LilianaVessOracle, 8)
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "A Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.LrwB1Activate(t, g, me, liliana, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{theirs, mine} {
		c := effectstest.FindBattlefieldCardByID(g, id)
		if c == nil {
			t.Fatalf("a creature card stayed in the graveyard")
		}
		if c.Controller != me.ID {
			t.Error("a returned creature is not under Liliana's controller")
		}
	}
	if !opp.Graveyard.Contains(spell) {
		t.Error("a noncreature card left the graveyard")
	}
}

func TestLrwB1LilianaVessPlusOneAndMinusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	liliana := effectstest.PushWalkerForTest(g, me.ID, "Liliana Vess", lrwB1LilianaVessOracle, 5)
	effectstest.AdvanceToMainOf(t, g, seat)
	hand := opp.Hand.Size()
	effectstest.LrwB1Activate(t, g, me, liliana, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Fatal("the controller was asked to discard")
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != hand-1 {
		t.Errorf("the target discarded %d, want 1", hand-opp.Hand.Size())
	}

	g2 := effectstest.NewCatalogGame(t)
	seat2 := g2.Turn.ActiveSeat
	me2, opp2 := g2.Seats[seat2], g2.Seats[(seat2+1)%len(g2.Seats)]
	liliana2 := effectstest.PushWalkerForTest(g2, me2.ID, "Liliana Vess", lrwB1LilianaVessOracle, 5)
	needle := effectstest.PushLibraryCardForTest(me2, game.Card{Name: "Anything", TypeLine: "Artifact"})
	effectstest.AdvanceToMainOf(t, g2, seat2)
	effectstest.LrwB1Activate(t, g2, me2, liliana2, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerSearchByID(t, g2, me2.ID, needle)
	effectstest.PassPriorityAroundTable(t, g2)
	top, err := me2.Library.Top()
	if err != nil || top.InstanceID != needle {
		t.Fatalf("top is %q, want the searched card", top.Name)
	}
	if top.IsKnownTo(opp2.ID) {
		t.Error("the −2 revealed the card; it doesn't say reveal")
	}
	if effectstest.LoyaltyOf(g2, liliana2) != 3 {
		t.Errorf("loyalty %d, want 3", effectstest.LoyaltyOf(g2, liliana2))
	}
}

func TestLrwB1GarrukWildspeaker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	garruk := effectstest.PushWalkerForTest(g, me.ID, "Garruk Wildspeaker", lrwB1GarrukWildspeakerOracle, 3)
	a := effectstest.B12Permanent(g, me.ID, "Forest A", "Basic Land — Forest")
	b := effectstest.B12Permanent(g, me.ID, "Forest B", "Basic Land — Forest")
	effectstest.AdvanceToMainOf(t, g, seat)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == a || id == b {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.LrwB1Activate(t, g, me, garruk, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(a, b)})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if tapped, _ := effectstest.BattlefieldCardTapped(g, id); tapped {
			t.Error("a target land is still tapped")
		}
	}
	if effectstest.LoyaltyOf(g, garruk) != 4 {
		t.Errorf("loyalty %d, want 4", effectstest.LoyaltyOf(g, garruk))
	}
}

func TestLrwB1GarrukMinusOneAndMinusFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	garruk := effectstest.PushWalkerForTest(g, me.ID, "Garruk Wildspeaker", lrwB1GarrukWildspeakerOracle, 3)
	effectstest.AdvanceToMainOf(t, g, seat)
	before := effectstest.CountNamed(g, "Beast")
	effectstest.LrwB1Activate(t, g, me, garruk, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountNamed(g, "Beast") != before+1 {
		t.Fatal("no Beast token")
	}

	g2 := effectstest.NewCatalogGame(t)
	seat2 := g2.Turn.ActiveSeat
	me2, opp2 := g2.Seats[seat2], g2.Seats[(seat2+1)%len(g2.Seats)]
	garruk2 := effectstest.PushWalkerForTest(g2, me2.ID, "Garruk Wildspeaker", lrwB1GarrukWildspeakerOracle, 4)
	mine := effectstest.B12Creature(g2, me2.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g2, opp2.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g2, seat2)
	effectstest.LrwB1Activate(t, g2, me2, garruk2, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.EffectivePower(t, g2, mine) != 5 || effectstest.EffectiveToughness(t, g2, mine) != 5 || !effectstest.HasEffectiveKeyword(t, g2, mine, "trample") {
		t.Error("your creature did not get +3/+3 and trample")
	}
	if effectstest.EffectivePower(t, g2, theirs) != 2 || effectstest.HasEffectiveKeyword(t, g2, theirs, "trample") {
		t.Error("an opponent's creature was pumped")
	}
	if g2.Battlefield.Contains(garruk2) {
		t.Error("Garruk survived −4 from four loyalty")
	}
}

// --- creatures and type changes -------------------------------------------------

func TestLrwB1LignifyMakesAVanilla04Treefolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	host := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flier", TypeLine: "Artifact Creature — Bird Construct",
		Power: 3, Toughness: 3, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Lignify", "Kindred Enchantment — Treefolk Aura", lrwB1LignifyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, host) != 0 || effectstest.EffectiveToughness(t, g, host) != 4 {
		t.Errorf("%d/%d, want 0/4", effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host))
	}
	if effectstest.HasEffectiveKeyword(t, g, host, "flying") {
		t.Error("the creature kept flying")
	}
	if !lrwB1HasType(t, g, host, "Treefolk") || lrwB1HasType(t, g, host, "Bird") {
		t.Errorf("subtypes %v, want Treefolk alone", effectstest.EffectiveSubtypes(t, g, host))
	}
	if !effectstest.LayeredCard(t, g, host).IsArtifact() {
		t.Error("the artifact stopped being an artifact")
	}
}

func TestLrwB1AmoeboidChangelingGivesAndTakesCreatureTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	amoeba := effectstest.B12Push(g, me.ID, "Amoeboid Changeling", "Creature — Shapeshifter", lrwB1AmoeboidChangelingOracle, 1, 1)
	other := effectstest.B12Push(g, me.ID, "Amoeboid Changeling", "Creature — Shapeshifter", lrwB1AmoeboidChangelingOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !lrwB1HasType(t, g, amoeba, "Goblin") {
		t.Fatal("a changeling is not every creature type")
	}
	effectstest.LrwB1Activate(t, g, me, amoeba, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !lrwB1HasType(t, g, bear, "Elf") || !lrwB1HasType(t, g, bear, "Bear") {
		t.Error("the Bear did not gain all creature types")
	}
	effectstest.LrwB1Activate(t, g, me, other, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if lrwB1HasType(t, g, bear, "Elf") || lrwB1HasType(t, g, bear, "Bear") {
		t.Error("the Bear kept a creature type after losing them all")
	}
}

func TestLrwB1ChangelingKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	avian := effectstest.B12Push(g, me.ID, "Avian Changeling", "Creature — Shapeshifter", lrwB1AvianChangelingOracle, 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, avian, "flying") || !lrwB1HasType(t, g, avian, "Kithkin") {
		t.Error("Avian Changeling is not a flying changeling")
	}
}

func TestLrwB1SkeletalChangelingRegeneratesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	skel := effectstest.B12Push(g, me.ID, "Skeletal Changeling", "Creature — Shapeshifter", lrwB1SkeletalChangelingOracle, 1, 1)
	effectstest.FloatForTest(g, me, "CB")
	if err := g.ActivateCatalogAbility(me.ID, skel, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	lrwB1DestroyAndExpectRegenerated(t, g, skel)
}

// lrwB1DestroyAndExpectRegenerated destroys a permanent and asserts its
// shield kept it on the battlefield, tapped.
func lrwB1DestroyAndExpectRegenerated(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var shields int
	g.ReadSnapshot(func() { shields = g.RegenerationShieldsOn(id) })
	if shields != 1 {
		t.Fatalf("%d regeneration shields, want 1", shields)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	if !g.Battlefield.Contains(id) {
		t.Fatal("the regenerated permanent was destroyed")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, id); !tapped {
		t.Error("regeneration did not tap it")
	}
}

func TestLrwB1HerbalPoulticeRegeneratesTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	poultice := effectstest.PushCatalogPermanent(g, me.ID, "Herbal Poultice", "Artifact", lrwB1HerbalPoulticeOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, poultice, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(poultice) {
		t.Error("the Poultice was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	lrwB1DestroyAndExpectRegenerated(t, g, bear)
}

func TestLrwB1MadAuntieLordsAndRegeneratesAnotherGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	auntie := effectstest.B12Push(g, me.ID, "Mad Auntie", "Creature — Goblin Shaman", lrwB1MadAuntieOracle, 2, 2)
	gob := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectstest.EffectivePower(t, g, gob) != 2 || effectstest.EffectivePower(t, g, theirs) != 1 || effectstest.EffectivePower(t, g, auntie) != 2 || effectstest.EffectivePower(t, g, bear) != 2 {
		t.Error("the lord pumped the wrong creatures")
	}
	if err := g.ActivateCatalogAbility(me.ID, auntie, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(auntie)}); err == nil {
		t.Error("Mad Auntie targeted herself")
	}
	if err := g.ActivateCatalogAbility(me.ID, auntie, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Error("a non-Goblin was accepted")
	}
	effectstest.LrwB1Activate(t, g, me, auntie, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	lrwB1DestroyAndExpectRegenerated(t, g, theirs)
}

func TestLrwB1TribalLords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cenn := effectstest.B12Push(g, me.ID, "Wizened Cenn", "Creature — Kithkin Cleric", lrwB1WizenedCennOracle, 2, 2)
	kith := effectstest.B12Creature(g, me.ID, "Kithkin", "Creature — Kithkin Soldier", 1, 1)
	theirKith := effectstest.B12Creature(g, opp.ID, "Their Kithkin", "Creature — Kithkin", 1, 1)
	sov := effectstest.B12Push(g, me.ID, "Sunrise Sovereign", "Creature — Giant Warrior", lrwB1SunriseSovereignOracle, 5, 5)
	giant := effectstest.B12Creature(g, me.ID, "Giant", "Creature — Giant", 3, 3)
	if effectstest.EffectivePower(t, g, kith) != 2 || effectstest.EffectivePower(t, g, theirKith) != 1 || effectstest.EffectivePower(t, g, cenn) != 2 {
		t.Error("Wizened Cenn pumped the wrong creatures")
	}
	if effectstest.EffectivePower(t, g, giant) != 5 || effectstest.EffectiveToughness(t, g, giant) != 5 || !effectstest.HasEffectiveKeyword(t, g, giant, "trample") {
		t.Error("the other Giant did not get +2/+2 and trample")
	}
	if effectstest.EffectivePower(t, g, sov) != 5 || effectstest.HasEffectiveKeyword(t, g, sov, "trample") {
		t.Error("Sunrise Sovereign pumped itself")
	}
}

func TestLrwB1TimberProtector(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	prot := effectstest.B12Push(g, me.ID, "Timber Protector", "Creature — Treefolk Warrior", lrwB1TimberProtectorOracle, 4, 6)
	tree := effectstest.B12Creature(g, me.ID, "Tree", "Creature — Treefolk", 2, 2)
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	kindred := effectstest.B12Permanent(g, me.ID, "Kindred Tree", "Kindred Enchantment — Treefolk")
	theirForest := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectstest.EffectivePower(t, g, tree) != 3 || effectstest.EffectivePower(t, g, prot) != 4 {
		t.Error("the anthem is wrong")
	}
	for id, want := range map[uuid.UUID]bool{tree: true, forest: true, kindred: true, theirForest: false, bear: false, prot: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "indestructible"); got != want {
			t.Errorf("%s indestructible = %v, want %v", id, got, want)
		}
	}
}

func TestLrwB1DoranMakesEveryCreatureDealDamageByToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	doran := effectstest.B12Push(g, me.ID, "Doran, the Siege Tower", "Legendary Creature — Treefolk Shaman", lrwB1DoranOracle, 0, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	for _, id := range []uuid.UUID{doran, theirs} {
		if !effectstest.LayeredCard(t, g, id).Effective().DamageByToughness {
			t.Error("a creature does not assign combat damage by toughness")
		}
	}
}

// --- triggers --------------------------------------------------------------------

func TestLrwB1ProwessOfTheFairMakesAnElfWhenANontokenElfDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Prowess of the Fair", "Kindred Enchantment — Elf", lrwB1ProwessOfTheFairOracle, false)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	theirElf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	before := effectstest.CountNamed(g, "Elf Warrior")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirElf) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("an opponent's Elf going to their graveyard triggered")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(elf) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountNamed(g, "Elf Warrior") != before+1 {
		t.Error("no Elf Warrior token")
	}
}

func TestLrwB1BoggartShenanigansPingsWhenAnotherGoblinDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Boggart Shenanigans", "Kindred Enchantment — Goblin", lrwB1BoggartShenanigansOracle, false)
	gob := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(gob) })
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
}

func TestLrwB1MerrowCommerceUntapsYourMerfolkAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.PushCatalogPermanent(g, me.ID, "Merrow Commerce", "Kindred Enchantment — Merfolk", lrwB1MerrowCommerceOracle, false)
	merfolk := effectstest.B12Creature(g, me.ID, "Merfolk", "Creature — Merfolk", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == merfolk || id == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	for i := 0; i < 40 && !(g.Turn.Step == game.StepEnd && g.Turn.ActiveSeat == seat); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, merfolk); tapped {
		t.Error("the Merfolk is still tapped")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, bear); !tapped {
		t.Error("a non-Merfolk was untapped")
	}
}

func TestLrwB1WortReturnsAGoblinCardAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	wort := effectstest.B12Push(g, me.ID, "Wort, Boggart Auntie", "Legendary Creature — Goblin Shaman", lrwB1WortOracle, 3, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, wort, "fear") {
		t.Error("Wort has no fear")
	}
	card := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Tarfire", TypeLine: "Kindred Instant — Goblin", Owner: me.ID, Controller: me.ID})
	// Walk to this seat's next upkeep.
	for i := 0; i < 80; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat {
			break
		}
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, card)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) {
		t.Error("the Goblin card did not return to hand")
	}
}

func TestLrwB1SeedguideAshFetchesForestsTappedWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ash := effectstest.B12Push(g, me.ID, "Seedguide Ash", "Creature — Treefolk Druid", lrwB1SeedguideAshOracle, 4, 4)
	f1 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	f2 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ash) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Island") != uuid.Nil {
		t.Error("an Island was offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, f1, f2)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{f1, f2} {
		if tapped, on := effectstest.BattlefieldCardTapped(g, id); !on || !tapped {
			t.Error("a Forest is not on the battlefield tapped")
		}
	}
}

func TestLrwB1AethersnipeBouncesANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := effectstest.B12Permanent(g, opp.ID, "Land", "Land")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	snipe := effectstest.CastHoldCreature(t, g, "Aethersnipe", "Creature — Elemental", lrwB1AethersnipeOracle, 4, 4)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(snipe) {
		t.Error("a hard-cast Aethersnipe left the battlefield")
	}
}

func TestLrwB1AethersnipeEvokedIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	id := effectstest.CastWithAltCost(t, g, "Aethersnipe", "Creature — Elemental", lrwB1AethersnipeOracle, "evoke")
	for i := 0; i < 40; i++ {
		if p := effectstest.LatestPickTarget(g, caster.ID); p != nil {
			effectstest.AnswerPickTarget(t, g, rock)
			continue
		}
		if c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceTriggerOrder); c != nil {
			if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
				t.Fatalf("trigger order: %v", err)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) && len(g.PendingChoices) == 0 {
			break
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Battlefield.Contains(id) {
		t.Error("an evoked Aethersnipe stayed on the battlefield")
	}
	if !opp.Hand.Contains(rock) {
		t.Error("the evoked Aethersnipe's bounce did not happen")
	}
}

// --- spells and activations -------------------------------------------------------

func TestLrwB1EyeblightsEndingDestroysANonElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elf := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Eyeblight's Ending", "Kindred Instant — Elf", lrwB1EyeblightsEndingOracle, effectstest.CardRefs(elf)); err == nil {
		t.Fatal("an Elf was accepted as the target")
	}
	effectstest.CastCatalogSpell(t, g, "Eyeblight's Ending", "Kindred Instant — Elf", lrwB1EyeblightsEndingOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the non-Elf survived")
	}
}

func TestLrwB1TarfireDealsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Tarfire", "Kindred Instant — Goblin", lrwB1TarfireOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
}

func TestLrwB1FamiliarsRuseReturnsACreatureAndCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	opponent := g.Seats[1]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	mine := effectstest.B12Creature(g, caster.ID, "Mine", "Creature — Bear", 2, 2)
	boltID := uuid.New()
	opponent.Hand.PushTop(game.Card{InstanceID: boltID, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: "4457ed35-7c10-48c8-9776-456485fdf070", Owner: opponent.ID, Controller: opponent.ID})
	if err := g.CastSpell(opponent.ID, boltID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}},
	}); err != nil {
		t.Fatalf("Bolt: %v", err)
	}
	ruse := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: ruse, Name: "Familiar's Ruse", TypeLine: "Instant",
		OracleID: lrwB1FamiliarsRuseOracle, Owner: caster.ID, Controller: caster.ID})
	if err := g.CastSpell(caster.ID, ruse, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: boltID}},
	}); err == nil {
		t.Fatal("cast without returning a creature")
	}
	if err := g.CastSpell(caster.ID, ruse, game.CastSpellParams{
		Targets:   []game.TargetRef{{Kind: game.TargetCard, ID: boltID}},
		ReturnIDs: []uuid.UUID{mine},
	}); err != nil {
		t.Fatalf("Familiar's Ruse: %v", err)
	}
	if !caster.Hand.Contains(mine) {
		t.Error("the creature was not returned as the cost")
	}
	life := caster.Life
	effectstest.PassPriorityAroundTable(t, g)
	if !opponent.Graveyard.Contains(boltID) || caster.Life != life {
		t.Error("the Bolt was not countered")
	}
}

func TestLrwB1SoulbrightFlamekinOffersEightRedOnTheThirdResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	flamekin := effectstest.B12Push(g, me.ID, "Soulbright Flamekin", "Creature — Elemental Shaman", lrwB1SoulbrightFlamekinOracle, 2, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 1; i <= 3; i++ {
		effectstest.FloatForTest(g, me, "CC")
		if err := g.ActivateCatalogAbility(me.ID, flamekin, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if i < 3 && effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
			t.Fatalf("resolution %d offered the mana", i)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the target did not gain trample")
	}
	c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if c == nil {
		t.Fatal("the third resolution did not offer the mana")
	}
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if me.ManaPool.CountColor("R") != 8 {
		t.Errorf("pool holds %d red, want 8", me.ManaPool.CountColor("R"))
	}
}

func TestLrwB1GaddockTeegStopsBigAndXNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Gaddock Teeg", "Legendary Creature — Kithkin Advisor", lrwB1GaddockTeegOracle, 2, 2)
	for _, tc := range []struct {
		name, typeLine, cost string
		refused              bool
	}{
		{"Big Sorcery", "Sorcery", "{3}{U}", true},
		{"X Sorcery", "Sorcery", "{X}{R}", true},
		{"Small Sorcery", "Sorcery", "{2}{U}", false},
		{"Big Creature", "Creature — Bear", "{5}{G}", false},
	} {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: tc.name, TypeLine: tc.typeLine, ManaCost: tc.cost, Owner: me.ID, Controller: me.ID})
		err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 1})
		if refused := err != nil; refused != tc.refused {
			t.Errorf("%s: cast err = %v, want refused %v", tc.name, err, tc.refused)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
}
