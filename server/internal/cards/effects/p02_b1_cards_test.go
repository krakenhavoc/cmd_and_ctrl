package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// p02_b1_cards_test.go — Portal Second Age, batch 1 (#5625): thirteen
// vanilla bodies, twenty-two creatures whose whole text is keywords or a
// block rule, and eleven cards whose text an existing catalog card
// already prints. Each test plays the real catalog entry through the
// engine and asserts what its text says.

const (
	p02B1GoblinGliderOracle    = "72bce821-46f0-40eb-abdb-575d4305b8fd"
	p02B1GoblinRaiderOracle    = "7f11c830-fced-40db-8c02-b3c54f3e4372"
	p02B1OgreTaskmasterOracle  = "5604f04e-fd72-41bf-906f-eb89d4a3475d"
	p02B1TalasWarriorOracle    = "12beabf3-a84b-49c2-9462-8876a4ac74d4"
	p02B1BloodcurdlingOracle   = "3ee2060a-5152-4a53-8183-cbd642e4cc29"
	p02B1ChorusOfWoeOracle     = "e175a3c1-df6c-4aa7-a4c9-893b466ad9e0"
	p02B1RighteousChargeOracle = "252d9183-16eb-4345-8d15-56218b33abdd"
	p02B1ExtinguishOracle      = "b2c9f074-57ca-4709-976d-f432f632483f"
	p02B1FalseSummoningOracle  = "4f891c68-c959-4210-94e5-94a8e487d5ef"
	p02B1GoblinLoreOracle      = "d6fe945c-b561-4e35-8c8e-7dcaa2ec45ab"
	p02B1LurkingStalkerOracle  = "8c17f48b-467e-4953-85d0-221cef5b3bff"
	p02B1RavenousRatsOracle    = "2fa1bbfd-92b5-482c-b32d-4cdc286474c4"
	p02B1SalvageOracle         = "c01a090a-11e4-469e-aded-74203c06fef4"
	p02B1ScreechingDrakeOracle = "da2ac9fa-52c2-4896-a596-89f287406641"
	p02B1TempleAcolyteOracle   = "17e4085e-af44-4ba8-af05-ccfe4b35f96a"

	p02B1TreeMonkeyOracle       = "d7318464-e400-424f-ae6d-a9070c64df04"
	p02B1DakmorBatOracle        = "724d8247-33b1-42bf-a827-7b66e45504c1"
	p02B1AlabornGrenadierOracle = "69c4429e-a8a4-49c3-bb43-9dce4cc6b90a"
	p02B1ArmoredGriffinOracle   = "ad3d6003-66d3-486d-aa54-ddce1adb5ff1"
	p02B1BrimstoneDragonOracle  = "b20302d9-eb45-4895-9597-cf3af1abce56"
	p02B1OgreBerserkerOracle    = "9e6f76dc-b7db-49b6-a61e-3495b81e61d7"
	p02B1AngelicWallOracle      = "4502b24f-604b-4e36-9168-31c1a1ab4dab"
	p02B1GoblinPikerOracle      = "50608184-90d3-43d2-a221-deb186c78323"
)

// p02B1Vanilla are the batch's bodies with no rules text.
var p02B1Vanilla = map[string]string{
	"17969b71-9612-45ff-b789-1c033f832154": "Alaborn Trooper",
	"03ad624e-71f6-4f72-8ce0-7700e83458d9": "Barbtooth Wurm",
	"008290c2-499d-425b-9a6c-c4aeb47c33ff": "Dakmor Scorpion",
	"be77e3f8-76a3-4d03-a8be-ad4b0b72f929": "Goblin Cavaliers",
	p02B1GoblinPikerOracle:                 "Goblin Piker",
	"6ba7ab42-918e-472a-b364-06073cc354d2": "Golden Bear",
	"e62ec170-c588-41ff-8be8-b05a125693a1": "Norwood Ranger",
	"f7a96612-4fc5-4944-8b74-9eb6f24190b4": "Obsidian Giant",
	"31426e08-bfd1-44f1-8473-5e982cb8e841": "Ogre Warrior",
	"d7316cc6-a2c2-479c-80b5-53c86165e3d1": "Plated Wurm",
	"6469f61d-1b4c-4813-bb4a-99f6e6388d7b": "Talas Merchant",
	"51af8662-07e6-4771-8ff2-c54364987698": "Trokin High Guard",
	"4f598459-b776-4ac1-bc69-8fde21674451": "Volunteer Militia",
}

// p02B1Keyworded are the batch's creatures whose whole text is printed
// keywords.
var p02B1Keyworded = map[string]struct {
	name, typeLine string
	keywords       []string
}{
	p02B1AlabornGrenadierOracle:            {"Alaborn Grenadier", "Creature — Human Soldier", []string{"vigilance"}},
	"61daa88d-fca4-4019-97f3-568346157ba6": {"Alaborn Musketeer", "Creature — Human Soldier", []string{"reach"}},
	"4e4e96aa-2f05-4f1d-96f8-6c42cd3be589": {"Norwood Archers", "Creature — Elf Archer", []string{"reach"}},
	p02B1TreeMonkeyOracle:                  {"Tree Monkey", "Creature — Monkey", []string{"reach"}},
	p02B1ArmoredGriffinOracle:              {"Armored Griffin", "Creature — Griffin", []string{"flying", "vigilance"}},
	p02B1BrimstoneDragonOracle:             {"Brimstone Dragon", "Creature — Dragon", []string{"flying", "haste"}},
	p02B1DakmorBatOracle:                   {"Dakmor Bat", "Creature — Bat", []string{"flying"}},
	"78962897-e2e3-4cb2-ae97-eca5dbe74e49": {"Moaning Spirit", "Creature — Spirit", []string{"flying"}},
	"8cb3f047-bd8b-4a34-8a2b-ced2cc0487eb": {"Talas Air Ship", "Creature — Human Pirate", []string{"flying"}},
	"71b7d5c4-c16c-4c6b-8e2e-e9b5cc128e1f": {"Talas Scout", "Creature — Human Pirate Scout", []string{"flying"}},
	"c643cfe1-5844-4eb1-b1f5-028382411773": {"Wild Griffin", "Creature — Griffin", []string{"flying"}},
	p02B1OgreBerserkerOracle:               {"Ogre Berserker", "Creature — Ogre Berserker", []string{"haste"}},
	p02B1AngelicWallOracle:                 {"Angelic Wall", "Creature — Wall", []string{"defender", "flying"}},
	"356e2df2-2b41-4d9d-8723-3e6a17a090e8": {"Goblin Mountaineer", "Creature — Goblin Scout", []string{"mountainwalk"}},
	"61652a2d-dfe6-4d4f-8fad-9de1960208cb": {"Lynx", "Creature — Cat", []string{"forestwalk"}},
	"568c4d02-501e-44ce-871c-adf9e4b9d8d6": {"River Bear", "Creature — Bear", []string{"islandwalk"}},
	"5c4cb33d-2a79-4934-839f-d34e9ca284f8": {"Raiding Nightstalker", "Creature — Nightstalker", []string{"swampwalk"}},
	"a6aa245c-b620-486c-864b-0ea4e9bc4048": {"Wild Ox", "Creature — Ox", []string{"swampwalk"}},
}

func TestP02B1EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{
		p02B1GoblinGliderOracle:    "Goblin Glider",
		p02B1GoblinRaiderOracle:    "Goblin Raider",
		p02B1OgreTaskmasterOracle:  "Ogre Taskmaster",
		p02B1TalasWarriorOracle:    "Talas Warrior",
		p02B1BloodcurdlingOracle:   "Bloodcurdling Scream",
		p02B1ChorusOfWoeOracle:     "Chorus of Woe",
		p02B1RighteousChargeOracle: "Righteous Charge",
		p02B1ExtinguishOracle:      "Extinguish",
		p02B1FalseSummoningOracle:  "False Summoning",
		p02B1GoblinLoreOracle:      "Goblin Lore",
		p02B1LurkingStalkerOracle:  "Lurking Nightstalker",
		p02B1RavenousRatsOracle:    "Ravenous Rats",
		p02B1SalvageOracle:         "Salvage",
		p02B1ScreechingDrakeOracle: "Screeching Drake",
		p02B1TempleAcolyteOracle:   "Temple Acolyte",
	}
	for oracle, name := range p02B1Vanilla {
		all[oracle] = name
	}
	for oracle, k := range p02B1Keyworded {
		all[oracle] = k.name
	}
	if len(all) != 46 {
		t.Fatalf("the batch lists %d cards, want 46", len(all))
	}
	for oracle, name := range all {
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

// --- bodies and keywords -----------------------------------------------------

func TestP02B1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for oracle, name := range p02B1Vanilla {
		spec, _ := Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
			t.Errorf("%s is a vanilla body and declares no abilities", name)
		}
	}
}

// A vanilla body on the battlefield is its printed card: it attacks and
// blocks like any creature.
func TestP02B1GoblinPikerIsAPlainCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	piker := b12Push(g, me.ID, "Goblin Piker", "Creature — Goblin Warrior", p02B1GoblinPikerOracle, 2, 1)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := porBlock(t, g, piker, opp.ID, bear); err != nil {
		t.Errorf("a vanilla Goblin Piker couldn't be blocked: %v", err)
	}
	if p, tough := ptOf(t, g, piker); p != 2 || tough != 1 {
		t.Errorf("Goblin Piker is %d/%d, want 2/1", p, tough)
	}
}

func TestP02B1KeywordedCreaturesHaveTheirKeywords(t *testing.T) {
	for oracle, k := range p02B1Keyworded {
		t.Run(k.name, func(t *testing.T) {
			spec, _ := Lookup(oracle)
			if !slices.Equal(spec.PrintedKeywords, k.keywords) {
				t.Errorf("printed keywords %v, want %v", spec.PrintedKeywords, k.keywords)
			}
			if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
				t.Errorf("%s declares more than its keywords", k.name)
			}
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, k.name, k.typeLine, oracle, 1, 1)
			for _, kw := range k.keywords {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s on the battlefield", k.name, kw)
				}
			}
		})
	}
}

// Flying: a ground creature can't block Dakmor Bat; Tree Monkey (reach)
// can (CR 702.9b).
func TestP02B1DakmorBatIsBlockedOnlyByFlyingOrReach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bat := b12Push(g, me.ID, "Dakmor Bat", "Creature — Bat", p02B1DakmorBatOracle, 1, 1)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := porBlock(t, g, bat, opp.ID, bear); err == nil {
		t.Error("a ground creature blocked Dakmor Bat")
	}
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	bat = b12Push(g, me.ID, "Dakmor Bat", "Creature — Bat", p02B1DakmorBatOracle, 1, 1)
	monkey := b12Push(g, opp.ID, "Tree Monkey", "Creature — Monkey", p02B1TreeMonkeyOracle, 1, 1)
	if err := porBlock(t, g, bat, opp.ID, monkey); err != nil {
		t.Errorf("Tree Monkey (reach) couldn't block Dakmor Bat: %v", err)
	}
}

// Haste: Brimstone Dragon and Ogre Berserker attack the turn they are
// cast (CR 702.10b).
func TestP02B1HasteCreaturesAttackTheTurnTheyArrive(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
	}{
		{"Brimstone Dragon", "Creature — Dragon", p02B1BrimstoneDragonOracle, 6, 6},
		{"Ogre Berserker", "Creature — Ogre Berserker", p02B1OgreBerserkerOracle, 4, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			advanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(id, opp.ID); err != nil {
				t.Errorf("%s couldn't attack the turn it arrived: %v", tc.name, err)
			}
		})
	}
}

// Vigilance: attacking doesn't tap it (CR 702.20b).
func TestP02B1VigilantCreaturesAttackWithoutTapping(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Alaborn Grenadier", "Creature — Human Soldier", p02B1AlabornGrenadierOracle},
		{"Armored Griffin", "Creature — Griffin", p02B1ArmoredGriffinOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			declareAttack(t, g, opp.ID, id)
			if alliTapped(t, g, id) {
				t.Errorf("%s tapped to attack", tc.name)
			}
		})
	}
}

// Defender: Angelic Wall can't attack (CR 702.3b) but blocks a flyer.
func TestP02B1AngelicWallCantAttackAndBlocksAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := b12Push(g, me.ID, "Angelic Wall", "Creature — Wall", p02B1AngelicWallOracle, 0, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wall, opp.ID); err == nil {
		t.Error("Angelic Wall attacked")
	}
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	bird := porFlyer(g, me.ID, "Bird")
	wall = b12Push(g, opp.ID, "Angelic Wall", "Creature — Wall", p02B1AngelicWallOracle, 0, 4)
	if err := porBlock(t, g, bird, opp.ID, wall); err != nil {
		t.Errorf("Angelic Wall couldn't block a flyer: %v", err)
	}
}

// Landwalk: unblockable while the defender controls the land type,
// blockable when they don't (CR 702.14c).
func TestP02B1LandwalkersCantBeBlockedWhileTheDefenderHasTheLand(t *testing.T) {
	for _, tc := range []struct{ name, landType string }{
		{"Goblin Mountaineer", "Mountain"},
		{"Lynx", "Forest"},
		{"River Bear", "Island"},
		{"Raiding Nightstalker", "Swamp"},
		{"Wild Ox", "Swamp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var oracle, typeLine string
			for o, k := range p02B1Keyworded {
				if k.name == tc.name {
					oracle, typeLine = o, k.typeLine
				}
			}
			for _, hasLand := range []bool{true, false} {
				g := newCatalogGame(t)
				me, opp := g.Seats[0], g.Seats[1]
				walker := b12Push(g, me.ID, tc.name, typeLine, oracle, 2, 2)
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

// --- block rules -------------------------------------------------------------

func TestP02B1CantBlockAndCantBeBlocked(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		want                   game.Restriction
	}{
		{"Goblin Glider", "Creature — Goblin", p02B1GoblinGliderOracle, game.CantBlock},
		{"Goblin Raider", "Creature — Goblin Warrior", p02B1GoblinRaiderOracle, game.CantBlock},
		{"Ogre Taskmaster", "Creature — Ogre", p02B1OgreTaskmasterOracle, game.CantBlock},
		{"Talas Warrior", "Creature — Human Pirate Warrior", p02B1TalasWarriorOracle, game.CantBeBlocked},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			assertRestrictions(t, g, id, tc.want)
		})
	}
}

func TestP02B1GoblinGliderFliesButCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	glider := b12Push(g, me.ID, "Goblin Glider", "Creature — Goblin", p02B1GoblinGliderOracle, 1, 1)
	if !hasEffectiveKeyword(t, g, glider, "flying") {
		t.Error("Goblin Glider has no flying")
	}
	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := porFlyer(g, me.ID, "Bird")
	glider = b12Push(g, opp.ID, "Goblin Glider", "Creature — Goblin", p02B1GoblinGliderOracle, 1, 1)
	if err := porBlock(t, g, attacker, opp.ID, glider); err == nil {
		t.Error("Goblin Glider blocked")
	}
}

func TestP02B1OgreTaskmasterCantBlockAndTalasWarriorCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ogre := b12Push(g, opp.ID, "Ogre Taskmaster", "Creature — Ogre", p02B1OgreTaskmasterOracle, 4, 3)
	if err := porBlock(t, g, attacker, opp.ID, ogre); err == nil {
		t.Error("Ogre Taskmaster blocked")
	}
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	warrior := b12Push(g, me.ID, "Talas Warrior", "Creature — Human Pirate Warrior", p02B1TalasWarriorOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := porBlock(t, g, warrior, opp.ID, bear); err == nil {
		t.Error("Talas Warrior was blocked")
	}
}

// --- spells ------------------------------------------------------------------

func TestP02B1BloodcurdlingScreamPumpsPowerByX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castFromHand(t, g, game.Card{Name: "Bloodcurdling Scream", TypeLine: "Sorcery", OracleID: p02B1BloodcurdlingOracle, ManaCost: "{X}{B}"},
		game.CastSpellParams{XValue: 3, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 5/2", p, tough)
	}
}

func TestP02B1TeamPumpsOnlyYourCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
	}{
		{"Chorus of Woe", p02B1ChorusOfWoeOracle, 3, 2},
		{"Righteous Charge", p02B1RighteousChargeOracle, 4, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := seatsForCombat(g)
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

// Extinguish counters only a sorcery spell, False Summoning only a
// creature spell.
func TestP02B1CounterspellsTakeOnlyTheirKindOfSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	piker := castCatalogSpell(t, g, "Goblin Piker", "Creature — Goblin Warrior", p02B1GoblinPikerOracle, nil)
	extinguish := pushCatalogHandCard(opp, "Extinguish", "Instant", p02B1ExtinguishOracle)
	if err := g.CastSpell(opp.ID, extinguish, game.CastSpellParams{Targets: cardTarget(piker)}); err == nil {
		t.Fatal("Extinguish accepted a creature spell as its target sorcery spell")
	}
	summoning := pushCatalogHandCard(opp, "False Summoning", "Instant", p02B1FalseSummoningOracle)
	if err := g.CastSpell(opp.ID, summoning, game.CastSpellParams{Targets: cardTarget(piker)}); err != nil {
		t.Fatalf("cast False Summoning at a creature spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(piker) || !me.Graveyard.Contains(piker) {
		t.Error("the countered creature spell is not in its owner's graveyard")
	}

	chorus := castCatalogSpell(t, g, "Chorus of Woe", "Sorcery", p02B1ChorusOfWoeOracle, nil)
	summoning2 := pushCatalogHandCard(opp, "False Summoning", "Instant", p02B1FalseSummoningOracle)
	if err := g.CastSpell(opp.ID, summoning2, game.CastSpellParams{Targets: cardTarget(chorus)}); err == nil {
		t.Fatal("False Summoning accepted a sorcery as its target creature spell")
	}
	if err := g.CastSpell(opp.ID, extinguish, game.CastSpellParams{Targets: cardTarget(chorus)}); err != nil {
		t.Fatalf("cast Extinguish at a sorcery: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(chorus) {
		t.Error("the countered sorcery is not in its owner's graveyard")
	}
	if p, tough := ptOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the countered Chorus of Woe still pumped: the Bear is %d/%d", p, tough)
	}
}

func TestP02B1GoblinLoreDrawsFourAndDiscardsThreeAtRandom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 6; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	hand, yard := me.Hand.Size(), me.Graveyard.Size()
	castCatalogSpell(t, g, "Goblin Lore", "Sorcery", p02B1GoblinLoreOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1 (four drawn, three discarded)", hand, got)
	}
	if got := me.Graveyard.Size(); got != yard+3+1 {
		t.Errorf("graveyard %d → %d, want +3 discarded and the spell", yard, got)
	}
}

func TestP02B1SalvagePutsYourCardOnTopAndRefusesAnotherGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushGraveyardCardTyped(me, "My Card", "Sorcery")
	theirs := pushGraveyardCardTyped(opp, "Their Card", "Sorcery")
	if err := castCatalogSpellErr(t, g, "Salvage", "Sorcery", p02B1SalvageOracle, cardRefs(theirs)); err == nil {
		t.Fatal("Salvage targeted a card in another player's graveyard")
	}
	castCatalogSpell(t, g, "Salvage", "Sorcery", p02B1SalvageOracle, cardRefs(mine))
	passPriorityAroundTable(t, g)
	if top, err := me.Library.Top(); err != nil || top.InstanceID != mine {
		t.Error("the card is not on top of the library")
	}
}

// --- creatures with a trigger --------------------------------------------------

func TestP02B1LurkingNightstalkerGetsPlusTwoWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	stalker := b12Push(g, me.ID, "Lurking Nightstalker", "Creature — Nightstalker", p02B1LurkingStalkerOracle, 1, 1)
	if p, tough := ptOf(t, g, stalker); p != 1 || tough != 1 {
		t.Fatalf("Lurking Nightstalker is %d/%d before attacking, want 1/1", p, tough)
	}
	declareAttack(t, g, them.ID, stalker)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, stalker); p != 3 || tough != 1 {
		t.Errorf("Lurking Nightstalker is %d/%d while attacking, want 3/1", p, tough)
	}
}

func TestP02B1RavenousRatsMakesTargetOpponentDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Hand.Size()
	if before < 1 {
		t.Fatalf("test setup: opponent hand %d", before)
	}
	castHoldCreature(t, g, "Ravenous Rats", "Creature — Rat", p02B1RavenousRatsOracle, 1, 1)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("the trigger targeted its own controller")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	if got := before - opp.Hand.Size(); got != 1 {
		t.Errorf("opponent discarded %d, want 1", got)
	}
}

func TestP02B1ScreechingDrakeLootsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	yard := me.Graveyard.Size()
	drake := castHoldCreature(t, g, "Screeching Drake", "Creature — Drake", p02B1ScreechingDrakeOracle, 2, 2)
	discardFromHand(t, g, me.ID)
	g.SettleResolution()
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(drake) {
		t.Fatal("Screeching Drake is not on the battlefield")
	}
	if !hasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("Screeching Drake has no flying")
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d (draw one, discard one)", got, before)
	}
	if got := me.Graveyard.Size(); got != yard+1 {
		t.Errorf("graveyard %d, want %d", got, yard+1)
	}
}

func TestP02B1TempleAcolyteGainsThreeOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castHoldCreature(t, g, "Temple Acolyte", "Creature — Human Cleric", p02B1TempleAcolyteOracle, 1, 3)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}
