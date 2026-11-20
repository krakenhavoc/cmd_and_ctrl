package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rav_5600_b4_cards_test.go — Ravnica: City of Guilds, batch 4 (#5600).
// Each test plays the real catalog entry through the engine and asserts
// what its text says.

const (
	ravB4FieryConclusionOracle = "91af3c81-615a-4b89-acf2-34c72001ebbd"
	ravB4GalvanicArcOracle     = "72a8620e-1ba2-40a8-9b22-11cc57864adf"
	ravB4GatherCourageOracle   = "ae033eb6-47ec-4b9e-9cdb-6d023aca07d1"
	ravB4GlassGolemOracle      = "24e48968-bfeb-4965-9e06-8b04397fcb53"
	ravB4GoliathSpiderOracle   = "17415789-159e-43cf-93f8-9926405ee8f1"
	ravB4GharialOracle         = "59483597-4d8c-4cdc-8bb5-226b1eef10f5"
	ravB4ForgelingOracle       = "c0d51f49-383b-40f1-9192-5925db1cdd1b"
	ravB4MossdogOracle         = "aedc3e69-3163-4b8d-97b2-0ad92f28a6dc"
	ravB4GuardianOracle        = "a7d5a9ca-27ec-4843-a9fd-0de0d67def39"
	ravB4GatekeeperOracle      = "861f0b1c-3b37-4d38-9232-e42565247076"
	ravB4MoldervineOracle      = "dd106f2d-5ede-43f0-928a-1ab392b90854"
	ravB4MoroiiOracle          = "418be12f-b40a-4018-86c1-98c9000a999a"
	ravB4RainOfEmbersOracle    = "041e2a01-3652-46c1-a9d1-d38ae7ee0551"
	ravB4RecollectOracle       = "5b38fafb-3999-46dc-928b-1677625c1943"
	ravB4ScionOracle           = "8aaa2217-012e-4e34-ba44-4986d7ee9ece"
	ravB4SagittarsOracle       = "027484b8-0df5-48b6-b69b-bb4e39c5ba81"
	ravB4SiegeWurmOracle       = "8a802eae-abfb-47fd-8458-1689055531ef"
	ravB4SkyknightOracle       = "61178a6c-70b7-447d-87ee-a8d9369c3d15"
	ravB4SparkmageOracle       = "69003cbc-8e5a-41dc-ac53-18f6e776217b"
	ravB4SunderingVitaeOracle  = "132c8a62-00bc-4b89-b775-099d8909f200"
	ravB4SpriteOracle          = "449142ff-7190-471b-99a5-f77095a827dd"
	ravB4TatteredDrakeOracle   = "1199a252-3c7c-4ae4-b063-dc132bfbea2f"
	ravB4UndercityShadeOracle  = "28f3d535-9503-4d35-aeed-d3e59259f760"
	ravB4FangtailOracle        = "5140cd2f-af30-4405-b5e4-996ef6349f59"
	ravB4SlasherOracle         = "63908256-a8a4-494f-b5b1-4b92c064af1d"
	ravB4KudzuOracle           = "62204fac-a367-475c-bfb9-80a87af2d2f1"
	ravB4VotaryOracle          = "83837ceb-7aae-4257-8545-a6cf34656e3b"
	ravB4WatchwolfOracle       = "a35c2e20-eb90-4132-b65f-be1fcb569819"
	ravB4SnitchesOracle        = "2ac4aca3-d5d6-4027-9f09-8b5f20e2438b"
	ravB4HuntedDragonOracle    = "39a4416f-9b40-432e-aeb6-cc7bf519b37a"
	ravB4HuntedLammasuOracle   = "a76667b5-2144-4395-a5c1-57cb3c29a99e"
	ravB4HuntedPhantasmOracle  = "ef4b334b-8407-4147-b434-602e59806906"
	ravB4HuntedTrollOracle     = "1f789fcf-3df6-45a6-a732-9f43e33718d6"
	ravB4SellSwordOracle       = "7549656b-359f-467e-90ab-6b21d93c26b9"
	ravB4InfectiousHostOracle  = "beb9c5c9-d4ab-423f-8937-91d308e09770"
	ravB4FlowOfIdeasOracle     = "dd397a32-06b0-4530-b075-412e2881aa53"
	ravB4TrumpeterOracle       = "6ad87598-5b57-496c-96ec-c04fe0616b6b"
	ravB4HammerfistOracle      = "ce916c92-350c-4546-882b-555dc450413d"
	ravB4GriffinOracle         = "beae83d2-20b8-47ef-83d9-25a7f90d1da2"
	ravB4ZealotOracle          = "4af23012-16eb-47be-b666-3e536d8f6c1b"
	ravB4RotwurmOracle         = "75df9e6a-a17c-4311-a4e7-ff0d75728337"
	ravB4SkysweeperOracle      = "f557697a-4c1b-44f6-a989-4879bbe26eb5"
	ravB4LoreBrokerOracle      = "e60a9e8d-0be4-4d08-a03a-af9f9e2be65b"
	ravB4ScarabOracle          = "9d954def-0045-4860-b154-cb64b6835fe0"
	ravB4HierarchOracle        = "2df404eb-fbfb-4902-b52b-71225b898dd4"
	ravB4GerminationOracle     = "3f9cf871-7df6-49d5-aab3-ad945571b707"
	ravB4ArmorerOracle         = "1b1214e9-1581-4008-b3d4-8b7af6b092fd"
	ravB4OathswornOracle       = "d3828907-1321-40ca-9531-30401e82a6b8"
	ravB4EmptyCatacombsOracle  = "49e0af9d-eb76-4ad6-b60a-430ae353728c"
	ravB4BearType              = "Creature — Bear"
	ravB4FlyingKeyword         = "flying"
	ravB4LifeFmt               = "life %d, want %d"
)

func TestRavB4EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ravB4FieryConclusionOracle: "Fiery Conclusion",
		ravB4GalvanicArcOracle:     "Galvanic Arc",
		ravB4GatherCourageOracle:   "Gather Courage",
		ravB4GlassGolemOracle:      "Glass Golem",
		ravB4GoliathSpiderOracle:   "Goliath Spider",
		ravB4GharialOracle:         "Grayscaled Gharial",
		ravB4ForgelingOracle:       "Greater Forgeling",
		ravB4MossdogOracle:         "Greater Mossdog",
		ravB4GuardianOracle:        "Guardian of Vitu-Ghazi",
		ravB4GatekeeperOracle:      "Loxodon Gatekeeper",
		ravB4MoldervineOracle:      "Moldervine Cloak",
		ravB4MoroiiOracle:          "Moroii",
		ravB4RainOfEmbersOracle:    "Rain of Embers",
		ravB4RecollectOracle:       "Recollect",
		ravB4ScionOracle:           "Scion of the Wild",
		ravB4SagittarsOracle:       "Selesnya Sagittars",
		ravB4SiegeWurmOracle:       "Siege Wurm",
		ravB4SkyknightOracle:       "Skyknight Legionnaire",
		ravB4SparkmageOracle:       "Sparkmage Apprentice",
		ravB4SunderingVitaeOracle:  "Sundering Vitae",
		ravB4SpriteOracle:          "Surveilling Sprite",
		ravB4TatteredDrakeOracle:   "Tattered Drake",
		ravB4UndercityShadeOracle:  "Undercity Shade",
		ravB4FangtailOracle:        "Viashino Fangtail",
		ravB4SlasherOracle:         "Viashino Slasher",
		ravB4KudzuOracle:           "Vinelasher Kudzu",
		ravB4VotaryOracle:          "Votary of the Conclave",
		ravB4WatchwolfOracle:       "Watchwolf",
		ravB4SnitchesOracle:        "Wizened Snitches",
		ravB4HuntedDragonOracle:    "Hunted Dragon",
		ravB4HuntedLammasuOracle:   "Hunted Lammasu",
		ravB4HuntedPhantasmOracle:  "Hunted Phantasm",
		ravB4HuntedTrollOracle:     "Hunted Troll",
		ravB4SellSwordOracle:       "Sell-Sword Brute",
		ravB4InfectiousHostOracle:  "Infectious Host",
		ravB4FlowOfIdeasOracle:     "Flow of Ideas",
		ravB4TrumpeterOracle:       "Thundersong Trumpeter",
		ravB4HammerfistOracle:      "Hammerfist Giant",
		ravB4GriffinOracle:         "Screeching Griffin",
		ravB4ZealotOracle:          "Flame-Kin Zealot",
		ravB4RotwurmOracle:         "Golgari Rotwurm",
		ravB4SkysweeperOracle:      "Elvish Skysweeper",
		ravB4LoreBrokerOracle:      "Lore Broker",
		ravB4ScarabOracle:          "Grave-Shell Scarab",
		ravB4HierarchOracle:        "Loxodon Hierarch",
		ravB4GerminationOracle:     "Golgari Germination",
		ravB4ArmorerOracle:         "Veteran Armorer",
		ravB4OathswornOracle:       "Oathsworn Giant",
		ravB4EmptyCatacombsOracle:  "Empty the Catacombs",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// ravB4Seats is seat 0 (the active seat at the start of a catalog game)
// and seat 1.
func ravB4Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[0], g.Seats[1]
}

// ravB4Flyer seeds a non-catalog creature with flying for `owner`.
func ravB4Flyer(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Bird", Power: power, Toughness: toughness, Keywords: []string{ravB4FlyingKeyword}})
}

// ravB4KillAndSettle destroys a permanent and runs the state checks, so
// its dies trigger is waiting.
func ravB4KillAndSettle(g *game.Game, id uuid.UUID) {
	b27Kill(g, id)
	g.RunStateChecksForTest()
}

// --- printed keywords -----------------------------------------------------------

func TestRavB4KeywordCreaturesHaveTheirKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Goliath Spider", "Creature — Spider", ravB4GoliathSpiderOracle, []string{"reach"}},
		{"Grayscaled Gharial", "Creature — Crocodile", ravB4GharialOracle, []string{"islandwalk"}},
		{"Skyknight Legionnaire", "Creature — Human Knight", ravB4SkyknightOracle, []string{ravB4FlyingKeyword, "haste"}},
		{"Moroii", "Creature — Vampire", ravB4MoroiiOracle, []string{ravB4FlyingKeyword}},
		{"Selesnya Sagittars", "Creature — Elf Archer", ravB4SagittarsOracle, []string{"reach"}},
		{"Surveilling Sprite", "Creature — Faerie Rogue", ravB4SpriteOracle, []string{ravB4FlyingKeyword}},
		{"Tattered Drake", "Creature — Zombie Drake", ravB4TatteredDrakeOracle, []string{ravB4FlyingKeyword}},
		{"Undercity Shade", "Creature — Shade", ravB4UndercityShadeOracle, []string{"fear"}},
		{"Wizened Snitches", "Creature — Faerie Rogue", ravB4SnitchesOracle, []string{ravB4FlyingKeyword}},
		{"Hunted Dragon", "Creature — Dragon", ravB4HuntedDragonOracle, []string{ravB4FlyingKeyword, "haste"}},
		{"Hunted Lammasu", "Creature — Lammasu", ravB4HuntedLammasuOracle, []string{ravB4FlyingKeyword}},
		{"Screeching Griffin", "Creature — Griffin", ravB4GriffinOracle, []string{ravB4FlyingKeyword}},
		{"Oathsworn Giant", "Creature — Giant Soldier", ravB4OathswornOracle, []string{"vigilance"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !p1g2HasKeyword(g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- convoke ----------------------------------------------------------------------

func TestRavB4SiegeWurmConvokesAndTramples(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB4Seats(g)
	taps := ravB1Tappers(g, me.ID, 7, "G")
	id, err := ravB1ConvokeCast(t, g, "Siege Wurm", "Creature — Wurm", "{5}{G}{G}", ravB4SiegeWurmOracle, 5, 5, taps)
	if err != nil {
		t.Fatalf("convoke cast with seven green creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Wurm never resolved")
	}
	if !p1g2HasKeyword(g, id, "trample") {
		t.Error("the Wurm has no trample")
	}
}

func TestRavB4GuardianOfVituGhaziConvokesAndHasVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB4Seats(g)
	taps := append(ravB1Tappers(g, me.ID, 7, "G"), ravB1Tappers(g, me.ID, 1, "W")...)
	id, err := ravB1ConvokeCast(t, g, "Guardian of Vitu-Ghazi", "Creature — Elemental", "{6}{G}{W}", ravB4GuardianOracle, 4, 7, taps)
	if err != nil {
		t.Fatalf("convoke cast with eight creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Guardian never resolved")
	}
	if !p1g2HasKeyword(g, id, "vigilance") {
		t.Error("the Guardian has no vigilance")
	}
}

func TestRavB4GatherCourageConvokesAPump(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB4Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", ravB4BearType, 2, 2)
	taps := ravB1Tappers(g, me.ID, 1, "G")
	if _, err := castWithTapParams(t, g, "Gather Courage", "Instant", "{G}", ravB4GatherCourageOracle,
		game.CastSpellParams{Strict: true, TapIDs: taps, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("convoke cast with one green creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
}

func TestRavB4SunderingVitaeConvokesADisenchant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	taps := ravB1Tappers(g, me.ID, 3, "G")
	if _, err := castWithTapParams(t, g, "Sundering Vitae", "Instant", "{2}{G}", ravB4SunderingVitaeOracle,
		game.CastSpellParams{Strict: true, TapIDs: taps, Targets: cardRefs(rock)}); err != nil {
		t.Fatalf("convoke cast with three green creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

// --- spells -----------------------------------------------------------------------

func TestRavB4FieryConclusionSacrificesAndDealsFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	target := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	victim := b12Creature(g, me.ID, "Victim", ravB4BearType, 2, 2)
	advanceToMain(t, g)
	id := b13HandCard(me, "Fiery Conclusion", "Instant", ravB4FieryConclusionOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(target)}); err == nil {
		t.Fatal("cast with no sacrifice was accepted")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(target), SacrificeIDs: []uuid.UUID{victim}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the victim was not sacrificed")
	}
	if g.Battlefield.Contains(target) {
		t.Error("a 5/5 survived 5 damage")
	}
}

func TestRavB4RainOfEmbersHitsEachCreatureAndEachPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	mine := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	castCatalogSpell(t, g, "Rain of Embers", "Sorcery", ravB4RainOfEmbersOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != lives[i]-1 {
			t.Errorf("seat %d: "+ravB4LifeFmt, i, p.Life, lives[i]-1)
		}
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("a 1/1 survived")
	}
	if got := findCardAnywhere(t, g, mine).DamageMarked; got != 1 {
		t.Errorf("the 3/3 has %d damage, want 1", got)
	}
}

func TestRavB4RecollectReturnsACardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB4Seats(g)
	old := wthGraveyardCard(me, "Old Spell", "Instant")
	castCatalogSpell(t, g, "Recollect", "Sorcery", ravB4RecollectOracle, cardRefs(old))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(old) || me.Graveyard.Contains(old) {
		t.Error("the card did not return to hand")
	}
}

func TestRavB4EmptyTheCatacombsReturnsEveryCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	mine := wthGraveyardCard(me, "My Bear", ravB4BearType)
	theirs := wthGraveyardCard(opp, "Their Bear", ravB4BearType)
	spell := wthGraveyardCard(opp, "Their Spell", "Instant")
	castCatalogSpell(t, g, "Empty the Catacombs", "Sorcery", ravB4EmptyCatacombsOracle, nil)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Error("your creature card stayed in your graveyard")
	}
	if !opp.Hand.Contains(theirs) {
		t.Error("an opponent's creature card stayed in their graveyard")
	}
	if !opp.Graveyard.Contains(spell) {
		t.Error("a noncreature card left the graveyard")
	}
}

func TestRavB4FlowOfIdeasDrawsOnePerIsland(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Flow of Ideas", "Sorcery", ravB4FlowOfIdeasOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d (two Islands you control)", got, hand+2)
	}
}

// --- Auras ------------------------------------------------------------------------

func TestRavB4GalvanicArcDealsThreeAndGrantsFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", ravB4BearType, 2, 2)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "Galvanic Arc", "Enchantment — Aura", ravB4GalvanicArcOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, ogre)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("a 3/3 survived 3 damage")
	}
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the enchanted creature has no first strike")
	}
}

func TestRavB4MoldervineCloakPumpsItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB4Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", ravB4BearType, 2, 2)
	castCatalogSpell(t, g, "Moldervine Cloak", "Enchantment — Aura", ravB4MoldervineOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
}

// --- dredge -----------------------------------------------------------------------

func TestRavB4DredgeReturnsTheCardInsteadOfDrawing(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Greater Mossdog", "Creature — Plant Dog", ravB4MossdogOracle, 3},
		{"Moldervine Cloak", "Enchantment — Aura", ravB4MoldervineOracle, 2},
		{"Grave-Shell Scarab", "Creature — Insect", ravB4ScarabOracle, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			p := g.Seats[1]
			id := gyCardWithOracle(p, tc.name, tc.typeLine, tc.oracle)
			lib, hand := p.Library.Size(), p.Hand.Size()
			if err := g.DrawCard(p.ID); err != nil {
				t.Fatalf("DrawCard: %v", err)
			}
			answerOptional(t, g, p, true)
			if !zoneHas(p.Hand, id) || zoneHas(p.Graveyard, id) {
				t.Fatal("the card did not return to the hand")
			}
			if got := p.Library.Size(); got != lib-tc.n {
				t.Errorf("library %d, want %d (milled %d, drew nothing)", got, lib-tc.n, tc.n)
			}
			if got := p.Hand.Size(); got != hand+1 {
				t.Errorf("hand %d, want %d", got, hand+1)
			}
		})
	}
}

// --- statics ----------------------------------------------------------------------

func TestRavB4LoxodonGatekeeperTapsWhatOpponentsBring(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := ravB4Seats(g)
	b12Push(g, opp.ID, "Loxodon Gatekeeper", "Creature — Elephant Soldier", ravB4GatekeeperOracle, 2, 3)
	land := playLandFromHand(t, g, "Forest", "")
	if !hmlTapped(t, g, land) {
		t.Error("an opponent's land entered untapped")
	}
	bear := castHoldCreature(t, g, "Bear", ravB4BearType, "", 2, 2)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear did not resolve")
	}
	if !hmlTapped(t, g, bear) {
		t.Error("an opponent's creature entered untapped")
	}
}

func TestRavB4LoxodonGatekeeperLeavesItsControllersOwnUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB4Seats(g)
	b12Push(g, me.ID, "Loxodon Gatekeeper", "Creature — Elephant Soldier", ravB4GatekeeperOracle, 2, 3)
	land := playLandFromHand(t, g, "Forest", "")
	if hmlTapped(t, g, land) {
		t.Error("its controller's own land entered tapped")
	}
}

func TestRavB4ScionOfTheWildCountsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	scion := b12Push(g, me.ID, "Scion of the Wild", "Creature — Avatar", ravB4ScionOracle, 0, 0)
	b12Creature(g, me.ID, "Bear", ravB4BearType, 2, 2)
	b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	b12Creature(g, opp.ID, "Their Bear", ravB4BearType, 2, 2)
	if p, tough := effectivePower(t, g, scion), effectiveToughness(t, g, scion); p != 3 || tough != 3 {
		t.Errorf("Scion is %d/%d, want 3/3", p, tough)
	}
}

func TestRavB4SelesnyaSagittarsBlocksTwo(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := ravB4Seats(g)
	id := b12Push(g, opp.ID, "Selesnya Sagittars", "Creature — Elf Archer", ravB4SagittarsOracle, 2, 5)
	if got := mbCapacity(t, g, id); got != 2 {
		t.Errorf("capacity = %d, want 2", got)
	}
}

func TestRavB4WizenedSnitchesRevealsEveryLibrarysTop(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	other := g.Seats[2]
	if topVisibleTo(g, opp.ID, other.ID) {
		t.Fatal("a library's top card is visible with nothing on the battlefield")
	}
	b12Push(g, me.ID, "Wizened Snitches", "Creature — Faerie Rogue", ravB4SnitchesOracle, 1, 3)
	for _, owner := range g.Seats {
		for _, viewer := range g.Seats {
			if !topVisibleTo(g, owner.ID, viewer.ID) {
				t.Errorf("seat %s cannot see the top of %s's library", viewer.ID, owner.ID)
			}
		}
	}
}

func TestRavB4HuntedPhantasmCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	phantasm := b12Push(g, me.ID, "Hunted Phantasm", "Creature — Spirit", ravB4HuntedPhantasmOracle, 4, 6)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	brAttack(t, g, phantasm)
	if err := g.DeclareBlocker(wall, phantasm); err == nil {
		t.Error("the Phantasm was blocked")
	}
}

func TestRavB4VeteranArmorerShieldsYourOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	armorer := b12Push(g, me.ID, "Veteran Armorer", "Creature — Human Soldier", ravB4ArmorerOracle, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", ravB4BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", ravB4BearType, 2, 2)
	if got := effectiveToughness(t, g, mine); got != 3 {
		t.Errorf("your other creature has toughness %d, want 3", got)
	}
	if got := effectiveToughness(t, g, armorer); got != 2 {
		t.Errorf("the Armorer has toughness %d, want 2", got)
	}
	if got := effectiveToughness(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature has toughness %d, want 2", got)
	}
}

func TestRavB4OathswornGiantGivesOthersToughnessAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	giant := b12Push(g, me.ID, "Oathsworn Giant", "Creature — Giant Soldier", ravB4OathswornOracle, 3, 4)
	mine := b12Creature(g, me.ID, "My Bear", ravB4BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", ravB4BearType, 2, 2)
	if got := effectiveToughness(t, g, mine); got != 4 {
		t.Errorf("your other creature has toughness %d, want 4", got)
	}
	if !hasEffectiveKeyword(t, g, mine, "vigilance") {
		t.Error("your other creature has no vigilance")
	}
	if got := effectiveToughness(t, g, giant); got != 4 {
		t.Errorf("the Giant has toughness %d, want 4", got)
	}
	if hasEffectiveKeyword(t, g, theirs, "vigilance") || effectiveToughness(t, g, theirs) != 2 {
		t.Error("an opponent's creature was affected")
	}
}

// --- triggered abilities ----------------------------------------------------------

func TestRavB4MoroiiLosesOneEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB4Seats(g)
	b12Push(g, me.ID, "Moroii", "Creature — Vampire", ravB4MoroiiOracle, 4, 4)
	walkToUpkeepOfSeat0(t, g)
	life := me.Life
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf(ravB4LifeFmt, me.Life, life-1)
	}
}

func TestRavB4VinelasherKudzuGrowsOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB4Seats(g)
	kudzu := pushCatalogPermanent(g, me.ID, "Vinelasher Kudzu", "Creature — Plant", ravB4KudzuOracle, false)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, kudzu); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestRavB4SparkmageApprenticePingsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	life := opp.Life
	castHoldCreature(t, g, "Sparkmage Apprentice", "Creature — Human Wizard", ravB4SparkmageOracle, 1, 1)
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf(ravB4LifeFmt, opp.Life, life-1)
	}
}

func TestRavB4FlameKinZealotPumpsAndHastesYourTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", ravB4BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", ravB4BearType, 2, 2)
	zealot := castHoldCreature(t, g, "Flame-Kin Zealot", "Creature — Elemental Berserker", ravB4ZealotOracle, 2, 2)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, zealot} {
		if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 3 || tough != 3 {
			t.Errorf("your creature is %d/%d, want 3/3", p, tough)
		}
		if !hasEffectiveKeyword(t, g, id, "haste") {
			t.Error("your creature has no haste")
		}
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

func TestRavB4HuntedCycleGivesTheTargetOpponentTokens(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		token                  string
		n, tokenP, tokenT      int
		tokenKeyword           string
	}{
		{"Hunted Dragon", "Creature — Dragon", ravB4HuntedDragonOracle, 6, 6, "Knight", 3, 2, 2, "first strike"},
		{"Hunted Lammasu", "Creature — Lammasu", ravB4HuntedLammasuOracle, 5, 5, "Horror", 1, 4, 4, ""},
		{"Hunted Phantasm", "Creature — Spirit", ravB4HuntedPhantasmOracle, 4, 6, "Goblin", 5, 1, 1, ""},
		{"Hunted Troll", "Creature — Troll Warrior", ravB4HuntedTrollOracle, 8, 4, "Faerie", 4, 1, 1, ravB4FlyingKeyword},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := ravB4Seats(g)
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			passPriorityAroundTable(t, g)
			pickPlayer(t, g, me.ID, opp.ID)
			passPriorityAroundTable(t, g)
			tokens := battlefieldIDsNamed(g, tc.token)
			if len(tokens) != tc.n {
				t.Fatalf("%d %s tokens, want %d", len(tokens), tc.token, tc.n)
			}
			for _, id := range tokens {
				if got := controllerOf(t, g, id); got != opp.ID {
					t.Errorf("a %s is controlled by %v, want the target opponent", tc.token, got)
				}
				if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != tc.tokenP || tough != tc.tokenT {
					t.Errorf("a %s is %d/%d, want %d/%d", tc.token, p, tough, tc.tokenP, tc.tokenT)
				}
				if tc.tokenKeyword != "" && !hasEffectiveKeyword(t, g, id, tc.tokenKeyword) {
					t.Errorf("a %s has no %s", tc.token, tc.tokenKeyword)
				}
			}
		})
	}
}

func TestRavB4SellSwordBruteDealsTwoToYouWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB4Seats(g)
	brute := b12Push(g, me.ID, "Sell-Sword Brute", "Creature — Human Mercenary", ravB4SellSwordOracle, 2, 2)
	life := me.Life
	ravB4KillAndSettle(g, brute)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf(ravB4LifeFmt, me.Life, life-2)
	}
}

func TestRavB4InfectiousHostDrainsTheTargetPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ravB4Seats(g)
	host := b12Push(g, me.ID, "Infectious Host", "Creature — Zombie", ravB4InfectiousHostOracle, 1, 1)
	life := opp.Life
	ravB4KillAndSettle(g, host)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf(ravB4LifeFmt, opp.Life, life-2)
	}
}

func TestRavB4SurveillingSpriteMayDrawWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB4Seats(g)
	sprite := b12Push(g, me.ID, "Surveilling Sprite", "Creature — Faerie Rogue", ravB4SpriteOracle, 1, 1)
	hand := me.Hand.Size()
	ravB4KillAndSettle(g, sprite)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestRavB4GolgariGerminationMakesASaprolingForANontokenDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB4Seats(g)
	pushCatalogPermanent(g, me.ID, "Golgari Germination", "Enchantment", ravB4GerminationOracle, false)
	token := b12Creature(g, me.ID, "Spirit", "Token Creature — Spirit", 1, 1)
	ravB4KillAndSettle(g, token)
	passPriorityAroundTable(t, g)
	if n := b27CountNamed(g, me.ID, "Saproling"); n != 0 {
		t.Fatalf("a token's death made %d Saprolings", n)
	}
	bear := b12Creature(g, me.ID, "My Bear", ravB4BearType, 2, 2)
	ravB4KillAndSettle(g, bear)
	passPriorityAroundTable(t, g)
	if n := b27CountNamed(g, me.ID, "Saproling"); n != 1 {
		t.Errorf("%d Saprolings, want 1", n)
	}
}

func TestRavB4LoxodonHierarchGainsFourAndRegeneratesYourTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ravB4Seats(g)
	life := me.Life
	hierarch := castHoldCreature(t, g, "Loxodon Hierarch", "Creature — Elephant Cleric", ravB4HierarchOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf(ravB4LifeFmt, me.Life, life+4)
	}
	bear := b12Creature(g, me.ID, "My Bear", ravB4BearType, 2, 2)
	apaMana(me, "G", "W")
	apaActivate(t, g, me, hierarch, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(hierarch) {
		t.Error("the Hierarch was not sacrificed")
	}
	b27Kill(g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the shielded Bear was destroyed")
	}
}

// --- activated abilities ----------------------------------------------------------

func TestRavB4RegenerationRows(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		row                    int
		mana                   []string
	}{
		{"Tattered Drake", "Creature — Zombie Drake", ravB4TatteredDrakeOracle, 0, []string{"B"}},
		{"Votary of the Conclave", "Creature — Human Soldier", ravB4VotaryOracle, 0, []string{"C", "C", "G"}},
		{"Hunted Troll", "Creature — Troll Warrior", ravB4HuntedTrollOracle, 0, []string{"G"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me, _ := ravB4Seats(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, id, tc.row, game.ActivateAbilityParams{})
			b27Kill(g, id)
			if !g.Battlefield.Contains(id) {
				t.Error("the shielded creature was destroyed")
			}
		})
	}
}

func TestRavB4SelfPumpRows(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		mana                   []string
		wantP, wantT           int
	}{
		{"Undercity Shade", "Creature — Shade", ravB4UndercityShadeOracle, 1, 1, []string{"B"}, 2, 2},
		{"Greater Forgeling", "Creature — Elemental", ravB4ForgelingOracle, 3, 4, []string{"C", "R"}, 6, 1},
		{"Viashino Slasher", "Creature — Lizard Warrior", ravB4SlasherOracle, 1, 2, []string{"R"}, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me, _ := ravB4Seats(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestRavB4ViashinoFangtailPingsAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := ravB4Seats(g)
	fangtail := b12Push(g, me.ID, "Viashino Fangtail", "Creature — Lizard Warrior", ravB4FangtailOracle, 3, 3)
	life := opp.Life
	apaActivate(t, g, me, fangtail, 0, game.ActivateAbilityParams{Targets: playerRefs(opp.ID)})
	if opp.Life != life-1 {
		t.Errorf(ravB4LifeFmt, opp.Life, life-1)
	}
	if !hmlTapped(t, g, fangtail) {
		t.Error("the Fangtail did not tap")
	}
}

func TestRavB4ThundersongTrumpeterStopsAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := ravB4Seats(g)
	trumpeter := b12Push(g, me.ID, "Thundersong Trumpeter", "Creature — Human Soldier", ravB4TrumpeterOracle, 2, 1)
	bear := b12Creature(g, me.ID, "My Bear", ravB4BearType, 2, 2)
	apaActivate(t, g, me, trumpeter, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Error("the target creature attacked")
	}
}

func TestRavB4ScreechingGriffinKeepsATargetFromBlockingIt(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := ravB4Seats(g)
	griffin := b12Push(g, me.ID, "Screeching Griffin", "Creature — Griffin", ravB4GriffinOracle, 2, 2)
	held := ravB4Flyer(g, opp.ID, "Held Bird", 1, 1)
	free := ravB4Flyer(g, opp.ID, "Free Bird", 1, 1)
	apaMana(me, "R")
	apaActivate(t, g, me, griffin, 0, game.ActivateAbilityParams{Targets: cardTarget(held)})
	brAttack(t, g, griffin)
	if err := g.DeclareBlocker(held, griffin); err == nil {
		t.Error("the target creature blocked the Griffin")
	}
	if err := g.DeclareBlocker(free, griffin); err != nil {
		t.Errorf("another flyer can't block the Griffin: %v", err)
	}
}

func TestRavB4HammerfistGiantHitsGroundCreaturesAndPlayers(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := ravB4Seats(g)
	giant := b12Push(g, me.ID, "Hammerfist Giant", "Creature — Giant Warrior", ravB4HammerfistOracle, 5, 4)
	bear := b12Creature(g, opp.ID, "Their Bear", ravB4BearType, 2, 2)
	bird := ravB4Flyer(g, opp.ID, "Their Bird", 1, 1)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	apaActivate(t, g, me, giant, 0, game.ActivateAbilityParams{})
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("a creature without flying survived 4 damage")
	}
	if g.Battlefield.Contains(giant) {
		t.Error("the Giant did not deal 4 damage to itself")
	}
	if !g.Battlefield.Contains(bird) {
		t.Error("a flyer was dealt damage")
	}
	for i, p := range g.Seats {
		if p.Life != lives[i]-4 {
			t.Errorf("seat %d: "+ravB4LifeFmt, i, p.Life, lives[i]-4)
		}
	}
}

func TestRavB4GolgariRotwurmSacrificesForALifeLoss(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := ravB4Seats(g)
	wurm := b12Push(g, me.ID, "Golgari Rotwurm", "Creature — Zombie Wurm", ravB4RotwurmOracle, 5, 4)
	bear := b12Creature(g, me.ID, "My Bear", ravB4BearType, 2, 2)
	life := opp.Life
	apaMana(me, "B")
	apaActivate(t, g, me, wurm, 0, game.ActivateAbilityParams{Targets: playerRefs(opp.ID), SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if opp.Life != life-1 {
		t.Errorf(ravB4LifeFmt, opp.Life, life-1)
	}
}

func TestRavB4ElvishSkysweeperDestroysOnlyAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := ravB4Seats(g)
	sweeper := b12Push(g, me.ID, "Elvish Skysweeper", "Creature — Elf Warrior", ravB4SkysweeperOracle, 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", ravB4BearType, 2, 2)
	ground := b12Creature(g, opp.ID, "Their Bear", ravB4BearType, 2, 2)
	bird := ravB4Flyer(g, opp.ID, "Their Bird", 3, 3)
	apaMana(me, "C", "C", "C", "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, sweeper, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(ground), SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("the Skysweeper targeted a creature without flying")
	}
	apaActivate(t, g, me, sweeper, 0, game.ActivateAbilityParams{Targets: cardTarget(bird), SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bird) {
		t.Error("the flyer survived")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
}

func TestRavB4LoreBrokerLootsTheTable(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, _ := ravB4Seats(g)
	broker := b12Push(g, me.ID, "Lore Broker", "Creature — Human Rogue", ravB4LoreBrokerOracle, 1, 2)
	hands := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i] = p.Hand.Size()
	}
	apaActivate(t, g, me, broker, 0, game.ActivateAbilityParams{})
	for i, p := range g.Seats {
		if got := p.Hand.Size(); got != hands[i]+1 {
			t.Errorf("seat %d hand %d before discarding, want %d", i, got, hands[i]+1)
		}
		if n := discardOwed(g, p.ID); n != 1 {
			t.Errorf("seat %d owes %d discards, want 1", i, n)
		}
	}
}

func TestRavB4GraveShellScarabSacrificesToDraw(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, _ := ravB4Seats(g)
	scarab := b12Push(g, me.ID, "Grave-Shell Scarab", "Creature — Insect", ravB4ScarabOracle, 4, 4)
	hand := me.Hand.Size()
	apaMana(me, "C")
	apaActivate(t, g, me, scarab, 0, game.ActivateAbilityParams{})
	if !me.Graveyard.Contains(scarab) {
		t.Error("the Scarab was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}
