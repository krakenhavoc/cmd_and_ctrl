package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// por_5630_cards_test.go — a batch of Portal cards (#5630). Every test
// plays the real catalog entry through the engine.

const (
	porArmoredPegasusOracle      = "f097a059-5505-4c3c-b879-7853ab6972ed"
	porArrogantVampireOracle     = "5edfe083-391b-41ae-ac4d-d648042cfa5e"
	porDesertDrakeOracle         = "ce8f4eb4-08b8-404b-9147-1e28c1b14a65"
	porDjinnOfTheLampOracle      = "72b42c63-fe4d-4823-9692-30fb5bab384a"
	porMoonSpriteOracle          = "1e41136b-afa7-45ee-8d04-d0c7514b5387"
	porSnappingDrakeOracle       = "e15060c3-3773-4548-8747-ff59dcf2b519"
	porSpottedGriffinOracle      = "4916773d-5ccb-48ff-8aa3-09771ae88e81"
	porStarlitAngelOracle        = "e66e3450-bddb-46ac-bc7e-c7732e258374"
	porWindDrakeOracle           = "d6ffdaf0-ac08-4de9-bbce-2eab2f86bcca"
	porRagingCougarOracle        = "695ebb35-55fd-4ed1-8cbe-e3fc1223115b"
	porRagingGoblinOracle        = "30997b43-fc13-41d3-8064-1ccc2cb6fd2b"
	porRagingMinotaurOracle      = "4cfd04b3-57bd-4860-8b47-57a0fd0a23fa"
	porArdentMilitiaOracle       = "23625877-b6db-480c-8885-a62b7d0457df"
	porKeenEyedArchersOracle     = "0ace32d6-7261-447c-9ee2-e03febaab91b"
	porWallOfGraniteOracle       = "8445094f-008b-491a-977c-e8582d5ab72c"
	porAnacondaOracle            = "3eff03f1-2c5f-4c59-b465-a8c4cd05e1ba"
	porBogRaidersOracle          = "60824fae-20ed-4122-82c9-e99a1b679c54"
	porBullHippoOracle           = "4d62a448-b6a5-43b1-a281-9e9361a5524a"
	porEliteCatWarriorOracle     = "9297e50f-f8a5-4793-b64d-a119404f94ce"
	porWillowDryadOracle         = "19d5eb29-2d35-44e7-afa1-3ba8d812ed41"
	porCravenGiantOracle         = "3ebd9c8d-c840-4c2e-ace6-352a68d1c20f"
	porCravenKnightOracle        = "27d138af-7c5d-4c7e-a310-63d5b03f039d"
	porHulkingGoblinOracle       = "1451b623-8214-4d81-9436-62dbe8b79539"
	porJungleLionOracle          = "45b42694-9ee8-44f3-9de3-aa2c0f1b6a13"
	porPhantomWarriorOracle      = "23745133-e5e2-4ce3-b94a-73d0d3d8a013"
	porCloudDragonOracle         = "3c46f309-69ae-43b1-adf6-bdb26599c1f4"
	porCloudPiratesOracle        = "d334aa85-3470-4f5d-9cbc-b88bf991a5af"
	porCloudSpiritOracle         = "9e1a6481-f460-4551-96e8-30b289f2cb92"
	porBeeStingOracle            = "f637d525-2f29-488a-9269-8e5aa377fbb7"
	porScorchingSpearOracle      = "9342fbb8-ab35-4895-946d-951ba6a2b067"
	porVolcanicHammerOracle      = "98fa5a06-0553-40fd-999c-bc31c9b3f4db"
	porLavaAxeOracle             = "387b6b07-a283-412d-94c3-f7f1dc76e858"
	porHandOfDeathOracle         = "dc45b2e3-272b-479b-8e3b-36eead606a3a"
	porRainOfTearsOracle         = "72cecab3-519e-4a23-9623-b423a5c5a251"
	porWintersGraspOracle        = "e9b8679d-52a9-4f0f-9365-f3e4b7a69805"
	porMonstrousGrowthOracle     = "35a05836-38d7-45c7-ac9a-996a682c2129"
	porHowlingFuryOracle         = "eda60752-d225-4fd0-9f0f-9b99e321b8fa"
	porSymbolOfUnsummoningOracle = "c44f1a81-269b-4f05-8ff2-e7ce19a93937"
	porMobilizeOracle            = "7416c8db-ec33-4d92-866c-b7fdc69c8b41"
	porVengeanceOracle           = "1d001145-5d14-43a9-bf3b-3ce5c20b2a46"
	porTouchOfBrillianceOracle   = "6365aba1-78d3-416c-89cd-9449578eedbf"
	porSacredNectarOracle        = "30870ee5-6ad7-48a9-983e-d3b018f2344f"
	porWarriorsChargeOracle      = "5cf28c04-76af-4eb8-9969-366dc8e04690"
	porSteadfastnessOracle       = "8f8dfc24-f466-4607-8991-acbcbb415db3"
	porVenerableMonkOracle       = "5c5cfa6d-857f-44a9-80f7-46f016ef71e4"
	porSpiritualGuardianOracle   = "d03860b4-c663-4230-9200-6b89fa849ae7"
	porOwlFamiliarOracle         = "099a5835-da6c-4e03-ad3e-aeb448897fed"
)

// porBlock attacks `defender` with `attacker` and tries to block it with
// `blocker`, returning the block declaration's verdict.
func porBlock(t *testing.T, g *game.Game, attacker, defender, blocker uuid.UUID) error {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	return g.DeclareBlocker(blocker, attacker)
}

// porFlyer seeds a non-catalog creature with printed flying.
func porFlyer(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: owner, Controller: owner, Keywords: []string{"flying"},
	})
}

// --- every card is in the catalog ----------------------------------------

func TestPor5630EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		porArmoredPegasusOracle, porArrogantVampireOracle, porDesertDrakeOracle,
		porDjinnOfTheLampOracle, porMoonSpriteOracle, porSnappingDrakeOracle,
		porSpottedGriffinOracle, porStarlitAngelOracle, porWindDrakeOracle,
		porRagingCougarOracle, porRagingGoblinOracle, porRagingMinotaurOracle,
		porArdentMilitiaOracle, porKeenEyedArchersOracle, porWallOfGraniteOracle,
		porAnacondaOracle, porBogRaidersOracle, porBullHippoOracle,
		porEliteCatWarriorOracle, porWillowDryadOracle, porCravenGiantOracle,
		porCravenKnightOracle, porHulkingGoblinOracle, porJungleLionOracle,
		porPhantomWarriorOracle, porCloudDragonOracle, porCloudPiratesOracle,
		porCloudSpiritOracle, porBeeStingOracle, porScorchingSpearOracle,
		porVolcanicHammerOracle, porLavaAxeOracle, porHandOfDeathOracle,
		porRainOfTearsOracle, porWintersGraspOracle, porMonstrousGrowthOracle,
		porHowlingFuryOracle, porSymbolOfUnsummoningOracle, porMobilizeOracle,
		porVengeanceOracle, porTouchOfBrillianceOracle, porSacredNectarOracle,
		porWarriorsChargeOracle, porSteadfastnessOracle, porVenerableMonkOracle,
		porSpiritualGuardianOracle, porOwlFamiliarOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- keyword bodies ----------------------------------------------------------

func TestPor5630KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keyword                string
	}{
		{"Armored Pegasus", "Creature — Pegasus", porArmoredPegasusOracle, "flying"},
		{"Arrogant Vampire", "Creature — Vampire", porArrogantVampireOracle, "flying"},
		{"Desert Drake", "Creature — Drake", porDesertDrakeOracle, "flying"},
		{"Djinn of the Lamp", "Creature — Djinn", porDjinnOfTheLampOracle, "flying"},
		{"Moon Sprite", "Creature — Faerie", porMoonSpriteOracle, "flying"},
		{"Snapping Drake", "Creature — Drake", porSnappingDrakeOracle, "flying"},
		{"Spotted Griffin", "Creature — Griffin", porSpottedGriffinOracle, "flying"},
		{"Starlit Angel", "Creature — Angel", porStarlitAngelOracle, "flying"},
		{"Wind Drake", "Creature — Drake", porWindDrakeOracle, "flying"},
		{"Cloud Dragon", "Creature — Illusion Dragon", porCloudDragonOracle, "flying"},
		{"Cloud Pirates", "Creature — Human Pirate", porCloudPiratesOracle, "flying"},
		{"Cloud Spirit", "Creature — Spirit", porCloudSpiritOracle, "flying"},
		{"Owl Familiar", "Creature — Bird", porOwlFamiliarOracle, "flying"},
		{"Raging Cougar", "Creature — Cat", porRagingCougarOracle, "haste"},
		{"Raging Goblin", "Creature — Goblin Berserker", porRagingGoblinOracle, "haste"},
		{"Raging Minotaur", "Creature — Minotaur Berserker", porRagingMinotaurOracle, "haste"},
		{"Ardent Militia", "Creature — Human Soldier", porArdentMilitiaOracle, "vigilance"},
		{"Keen-Eyed Archers", "Creature — Elf Archer", porKeenEyedArchersOracle, "reach"},
		{"Wall of Granite", "Creature — Wall", porWallOfGraniteOracle, "defender"},
		{"Anaconda", "Creature — Snake", porAnacondaOracle, "swampwalk"},
		{"Bog Raiders", "Creature — Zombie", porBogRaidersOracle, "swampwalk"},
		{"Bull Hippo", "Creature — Hippo", porBullHippoOracle, "islandwalk"},
		{"Elite Cat Warrior", "Creature — Cat Warrior", porEliteCatWarriorOracle, "forestwalk"},
		{"Willow Dryad", "Creature — Dryad", porWillowDryadOracle, "forestwalk"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

// Flying: a ground creature can't block it, a creature with reach can.
func TestPor5630WindDrakeIsBlockedOnlyByFlyingOrReach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drake := b12Push(g, me.ID, "Wind Drake", "Creature — Drake", porWindDrakeOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := porBlock(t, g, drake, opp.ID, bear); err == nil {
		t.Error("a ground creature blocked Wind Drake")
	}
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	drake = b12Push(g, me.ID, "Wind Drake", "Creature — Drake", porWindDrakeOracle, 2, 2)
	archers := b12Push(g, opp.ID, "Keen-Eyed Archers", "Creature — Elf Archer", porKeenEyedArchersOracle, 2, 2)
	if err := porBlock(t, g, drake, opp.ID, archers); err != nil {
		t.Errorf("Keen-Eyed Archers (reach) couldn't block Wind Drake: %v", err)
	}
}

// Haste: a Raging Goblin cast this turn attacks; a creature without
// haste cast this turn can't.
func TestPor5630RagingGoblinAttacksTheTurnItArrives(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	goblin := castHoldCreature(t, g, "Raging Goblin", "Creature — Goblin Berserker", porRagingGoblinOracle, 1, 1)
	knight := castHoldCreature(t, g, "Craven Knight", "Creature — Human Knight", porCravenKnightOracle, 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(knight, opp.ID); err == nil {
		t.Error("a summoning-sick creature without haste attacked")
	}
	if err := g.DeclareAttacker(goblin, opp.ID); err != nil {
		t.Errorf("Raging Goblin couldn't attack the turn it arrived: %v", err)
	}
}

func TestPor5630ArdentMilitiaAttacksWithoutTapping(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	militia := b12Push(g, me.ID, "Ardent Militia", "Creature — Human Soldier", porArdentMilitiaOracle, 2, 5)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(militia, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if alliTapped(t, g, militia) {
		t.Error("Ardent Militia tapped to attack")
	}
}

func TestPor5630WallOfGraniteCantAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := b12Push(g, me.ID, "Wall of Granite", "Creature — Wall", porWallOfGraniteOracle, 0, 7)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wall, opp.ID); err == nil {
		t.Error("Wall of Granite attacked")
	}
}

// Landwalk: unblockable while the defender controls the land type,
// blockable when they don't.
func TestPor5630LandwalkersCantBeBlockedWhileTheDefenderHasTheLand(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle, landType string
	}{
		{"Anaconda", "Creature — Snake", porAnacondaOracle, "Swamp"},
		{"Bog Raiders", "Creature — Zombie", porBogRaidersOracle, "Swamp"},
		{"Bull Hippo", "Creature — Hippo", porBullHippoOracle, "Island"},
		{"Elite Cat Warrior", "Creature — Cat Warrior", porEliteCatWarriorOracle, "Forest"},
		{"Willow Dryad", "Creature — Dryad", porWillowDryadOracle, "Forest"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, hasLand := range []bool{true, false} {
				g := newCatalogGame(t)
				me, opp := g.Seats[0], g.Seats[1]
				walker := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
				bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
				if hasLand {
					b12Permanent(g, opp.ID, tc.landType, "Basic Land — "+tc.landType)
				} else {
					b12Permanent(g, opp.ID, "Wastes", "Basic Land")
				}
				err := porBlock(t, g, walker, opp.ID, bear)
				if hasLand && err == nil {
					t.Errorf("%s was blocked while the defender controls a %s", tc.name, tc.landType)
				}
				if !hasLand && err != nil {
					t.Errorf("%s couldn't be blocked with no %s: %v", tc.name, tc.landType, err)
				}
			}
		})
	}
}

func TestPor5630CantBlockAndCantBeBlocked(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		want                   game.Restriction
	}{
		{"Craven Giant", "Creature — Giant", porCravenGiantOracle, game.CantBlock},
		{"Craven Knight", "Creature — Human Knight", porCravenKnightOracle, game.CantBlock},
		{"Hulking Goblin", "Creature — Goblin", porHulkingGoblinOracle, game.CantBlock},
		{"Jungle Lion", "Creature — Cat", porJungleLionOracle, game.CantBlock},
		{"Phantom Warrior", "Creature — Illusion Warrior", porPhantomWarriorOracle, game.CantBeBlocked},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			assertRestrictions(t, g, id, tc.want)
		})
	}
}

func TestPor5630JungleLionCantBlockAndPhantomWarriorCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	lion := b12Push(g, opp.ID, "Jungle Lion", "Creature — Cat", porJungleLionOracle, 2, 1)
	if err := porBlock(t, g, attacker, opp.ID, lion); err == nil {
		t.Error("Jungle Lion blocked")
	}
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	warrior := b12Push(g, me.ID, "Phantom Warrior", "Creature — Illusion Warrior", porPhantomWarriorOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := porBlock(t, g, warrior, opp.ID, bear); err == nil {
		t.Error("Phantom Warrior was blocked")
	}
}

// "Can block only creatures with flying": refused against a ground
// attacker, allowed against a flyer.
func TestPor5630CloudCreaturesBlockOnlyFlyers(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
	}{
		{"Cloud Dragon", "Creature — Illusion Dragon", porCloudDragonOracle},
		{"Cloud Pirates", "Creature — Human Pirate", porCloudPiratesOracle},
		{"Cloud Spirit", "Creature — Spirit", porCloudSpiritOracle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			cloud := b12Push(g, opp.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			if err := porBlock(t, g, bear, opp.ID, cloud); err == nil {
				t.Errorf("%s blocked a creature without flying", tc.name)
			}
			g = newCatalogGame(t)
			me, opp = g.Seats[0], g.Seats[1]
			bird := porFlyer(g, me.ID, "Bird")
			cloud = b12Push(g, opp.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			if err := porBlock(t, g, bird, opp.ID, cloud); err != nil {
				t.Errorf("%s couldn't block a flyer: %v", tc.name, err)
			}
		})
	}
}

// --- damage ----------------------------------------------------------------

func TestPor5630BurnHitsAPlayer(t *testing.T) {
	cases := []struct {
		name, oracle string
		amount       int
	}{
		{"Bee Sting", porBeeStingOracle, 2},
		{"Scorching Spear", porScorchingSpearOracle, 1},
		{"Volcanic Hammer", porVolcanicHammerOracle, 3},
		{"Lava Axe", porLavaAxeOracle, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			life := opp.Life
			castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, []game.TargetRef{playerTarget(opp.ID)})
			passPriorityAroundTable(t, g)
			if opp.Life != life-tc.amount {
				t.Errorf("life %d -> %d, want -%d", life, opp.Life, tc.amount)
			}
		})
	}
}

func TestPor5630BurnKillsACreature(t *testing.T) {
	cases := []struct {
		name, oracle string
		toughness    int
	}{
		{"Bee Sting", porBeeStingOracle, 2},
		{"Scorching Spear", porScorchingSpearOracle, 1},
		{"Volcanic Hammer", porVolcanicHammerOracle, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			victim := b12Creature(g, g.Seats[1].ID, "Victim", "Creature — Bear", 1, tc.toughness)
			castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, cardTarget(victim))
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(victim) {
				t.Errorf("%s left a %d-toughness creature alive", tc.name, tc.toughness)
			}
		})
	}
}

func TestPor5630LavaAxeCantTargetACreature(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Lava Axe", "Sorcery", porLavaAxeOracle, cardTarget(bear)); err == nil {
		t.Error("Lava Axe targeted a creature")
	}
}

// --- destroy ---------------------------------------------------------------

func TestPor5630HandOfDeathDestroysOnlyANonblackCreature(t *testing.T) {
	g := newCatalogGame(t)
	black := b12Creature(g, g.Seats[1].ID, "Zombie", "Creature — Zombie", 2, 2)
	alliSetColors(g, black, "B")
	if err := castCatalogSpellErr(t, g, "Hand of Death", "Sorcery", porHandOfDeathOracle, cardTarget(black)); err == nil {
		t.Error("Hand of Death targeted a black creature")
	}
	g = newCatalogGame(t)
	green := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	alliSetColors(g, green, "G")
	castCatalogSpell(t, g, "Hand of Death", "Sorcery", porHandOfDeathOracle, cardTarget(green))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the green creature survived Hand of Death")
	}
}

func TestPor5630LandDestruction(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Rain of Tears", porRainOfTearsOracle},
		{"Winter's Grasp", porWintersGraspOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
			if err := castCatalogSpellErr(t, g, tc.name, "Sorcery", tc.oracle, cardTarget(bear)); err == nil {
				t.Errorf("%s targeted a creature", tc.name)
			}
			g = newCatalogGame(t)
			land := b12Permanent(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")
			castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, cardTarget(land))
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(land) {
				t.Errorf("the land survived %s", tc.name)
			}
		})
	}
}

func TestPor5630VengeanceDestroysOnlyATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	untapped := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Vengeance", "Sorcery", porVengeanceOracle, cardTarget(untapped)); err == nil {
		t.Error("Vengeance targeted an untapped creature")
	}
	g = newCatalogGame(t)
	tapped := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	b02bSetTapped(g, tapped, true)
	castCatalogSpell(t, g, "Vengeance", "Sorcery", porVengeanceOracle, cardTarget(tapped))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived Vengeance")
	}
}

// --- pumps -----------------------------------------------------------------

func TestPor5630TargetPumps(t *testing.T) {
	cases := []struct {
		name, oracle string
		power, tough int
	}{
		{"Monstrous Growth", porMonstrousGrowthOracle, 6, 6},
		{"Howling Fury", porHowlingFuryOracle, 6, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
			castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if p, tt := ptOf(t, g, bear); p != tc.power || tt != tc.tough {
				t.Errorf("bear is %d/%d, want %d/%d", p, tt, tc.power, tc.tough)
			}
		})
	}
}

func TestPor5630TeamPumpsOnlyYourCreatures(t *testing.T) {
	cases := []struct {
		name, oracle string
		power, tough int
	}{
		{"Warrior's Charge", porWarriorsChargeOracle, 3, 3},
		{"Steadfastness", porSteadfastnessOracle, 2, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
			passPriorityAroundTable(t, g)
			if p, tt := ptOf(t, g, mine); p != tc.power || tt != tc.tough {
				t.Errorf("my bear is %d/%d, want %d/%d", p, tt, tc.power, tc.tough)
			}
			if p, tt := ptOf(t, g, theirs); p != 2 || tt != 2 {
				t.Errorf("their bear is %d/%d, want 2/2", p, tt)
			}
		})
	}
}

// --- other spells ------------------------------------------------------------

func TestPor5630SymbolOfUnsummoningBouncesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Symbol of Unsummoning", "Sorcery", porSymbolOfUnsummoningOracle, cardTarget(bear))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestPor5630MobilizeUntapsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b02bSetTapped(g, mine, true)
	b02bSetTapped(g, theirs, true)
	castCatalogSpell(t, g, "Mobilize", "Sorcery", porMobilizeOracle, nil)
	passPriorityAroundTable(t, g)
	if alliTapped(t, g, mine) {
		t.Error("your creature is still tapped")
	}
	if !alliTapped(t, g, theirs) {
		t.Error("an opponent's creature untapped")
	}
}

func TestPor5630TouchOfBrillianceDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Touch of Brilliance", "Sorcery", porTouchOfBrillianceOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want +2", hand, me.Hand.Size())
	}
}

func TestPor5630SacredNectarGainsFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castCatalogSpell(t, g, "Sacred Nectar", "Sorcery", porSacredNectarOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, me.Life)
	}
}

// --- enters triggers ---------------------------------------------------------

func TestPor5630EntersGainLife(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		power, tough, life     int
	}{
		{"Venerable Monk", "Creature — Human Monk Cleric", porVenerableMonkOracle, 2, 2, 2},
		{"Spiritual Guardian", "Creature — Spirit", porSpiritualGuardianOracle, 3, 4, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			life := me.Life
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			if me.Life != life+tc.life {
				t.Errorf("life %d -> %d, want +%d", life, me.Life, tc.life)
			}
		})
	}
}

func TestPor5630OwlFamiliarLootsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	yard := me.Graveyard.Size()
	owl := castHoldCreature(t, g, "Owl Familiar", "Creature — Bird", porOwlFamiliarOracle, 1, 1)
	discardFromHand(t, g, me.ID)
	g.SettleResolution()
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(owl) {
		t.Fatal("Owl Familiar is not on the battlefield")
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d (draw one, discard one)", got, before)
	}
	if got := me.Graveyard.Size(); got != yard+1 {
		t.Errorf("graveyard %d, want %d", got, yard+1)
	}
}
