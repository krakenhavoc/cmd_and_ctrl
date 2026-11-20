package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ptk_5622_cards_test.go — Portal Three Kingdoms, batch 1 (#5622). Every
// test plays the real catalog entry through the engine.

const (
	ptkShuEliteInfantryOracle        = "806b4f0c-f577-4d7f-b826-09e4479c314e"
	ptkShuFootSoldiersOracle         = "4d7ff9d5-dc06-4a70-a9c8-d6f88f02fd67"
	ptkStrawSoldiersOracle           = "8138a9d7-63cb-4d40-ac9e-1a9b10766113"
	ptkWuInfantryOracle              = "a5a39eaf-104d-4dfa-b370-5d3de0d158fe"
	ptkWeiInfantryOracle             = "41283a82-ce3f-41a6-995e-bc30d8ab5cf9"
	ptkBarbarianHordeOracle          = "eca9cb24-4dc8-48eb-b055-5fdf7c67bd43"
	ptkIndependentTroopsOracle       = "b9a8adf0-60f1-491c-8bbf-0b341fd23fab"
	ptkForestBearOracle              = "0994d596-0b15-4156-8e94-5b566dcad6cd"
	ptkMengHuosHordeOracle           = "27e6de72-68c4-4154-807f-5944fe97c92b"
	ptkSouthernElephantOracle        = "79de2319-078d-46db-9044-0088c623b73f"
	ptkTrainedJackalOracle           = "99545320-b6ad-4caf-b29a-47f637395ef7"
	ptkAlertShuInfantryOracle        = "0f019387-4684-493f-80c4-de6023096306"
	ptkShuCavalryOracle              = "d9180f63-c745-41a1-9f01-273e7ee13176"
	ptkShuEliteCompanionsOracle      = "edf304b8-523e-4233-8d77-1d45d33cbb86"
	ptkLuMengWuGeneralOracle         = "a1e2a966-6aad-4325-b644-3ced319bdd7d"
	ptkWuEliteCavalryOracle          = "7dc5d729-0553-46fb-aff7-2f0d9573837b"
	ptkWuLightCavalryOracle          = "c24da532-311b-4b83-82d2-7dc80409ba12"
	ptkWeiEliteCompanionsOracle      = "2db5e0cd-c991-4fda-9001-9e4f10e8a805"
	ptkWeiScoutOracle                = "f7fa054f-39cb-47d3-a55d-3f70ae164063"
	ptkWeiStrikeForceOracle          = "002ad5fa-a20c-4d30-beef-401d89868d67"
	ptkBarbarianGeneralOracle        = "d4f2e305-bec0-48c8-8040-f85618c71cc9"
	ptkLadyZhurongWarriorQueenOracle = "531e6ffb-28e0-4e8e-ad88-2f3a243789cf"
	ptkShuGeneralOracle              = "0fef6729-b044-4daf-834a-ac8c9711ad8a"
	ptkZhangFeiFierceWarriorOracle   = "f3f815b8-75c6-40e0-9716-872f0f5c9d46"
	ptkLuBuMasterAtArmsOracle        = "ff959f1f-0982-4761-a6b0-3044dff1d617"
	ptkMountainBanditOracle          = "e3638c54-e494-47d2-b614-94c9d9add85a"
	ptkRenegadeTroopsOracle          = "83a81308-4a15-49cd-b2f6-51431cca79e9"
	ptkTaoistHermitOracle            = "960b1251-12af-4766-8df1-bb98e7e967f5"
	ptkZodiacPigOracle               = "eb0a0515-7fa4-4319-9fc3-7a31c608ce9b"
	ptkZodiacRatOracle               = "5ea4d1d7-219b-4842-99ea-291842baca0d"
	ptkZodiacSnakeOracle             = "b97c4d9a-8ea0-40d0-839a-afcd785e807e"
	ptkZodiacOxOracle                = "9c9e71d9-a636-4574-8b7e-e6dd0f71c7c8"
	ptkZodiacDogOracle               = "4431ec11-9bd7-4750-8bc7-378beba06201"
	ptkZodiacGoatOracle              = "01800a0c-ad93-4e45-a904-38253a143638"
	ptkZodiacHorseOracle             = "f778bf66-e823-4b4c-8586-97e1d519efd8"
	ptkZodiacMonkeyOracle            = "47dd7ef4-7645-4878-ab91-261aa8fcf834"
	ptkZodiacRabbitOracle            = "048cc868-1657-448d-9f2c-97973e73f477"
	ptkZodiacTigerOracle             = "79aba9a7-b957-4139-8d50-f1c91fa17793"
	ptkZodiacRoosterOracle           = "7280b7c5-e4d0-4a20-8634-5e240e6b3ca0"
	ptkYoungWeiRecruitsOracle        = "cdada8f0-d001-485c-a245-5e1c1ab748ab"
	ptkYellowScarvesTroopsOracle     = "fcfa5671-a7b1-4170-9c72-f18a9eccfa31"
	ptkYellowScarvesCavalryOracle    = "8ff7e8e9-11b4-4a77-a423-70777f00482b"
	ptkYellowScarvesGeneralOracle    = "6fbade4d-502d-45d7-a79e-3fd9e3e7ccdf"
	ptkBrilliantPlanOracle           = "d83e1a42-11d1-412a-b66c-850c7a528777"
	ptkSpringOfEternalPeaceOracle    = "6fcdb870-a1b6-49fe-8b5c-cdde074a7f7d"
	ptkFireAmbushOracle              = "50463946-1ce3-4ff0-ad68-2fb87adbe2fd"
	ptkGhostlyVisitOracle            = "73b6694a-47d4-417a-af9b-470da14b5b0e"
	ptkWieldingTheGreenDragonOracle  = "22fa9020-783e-4292-857e-9d45a3458d71"
	ptkVirtuousChargeOracle          = "5315ed83-cb93-4522-b439-3784e7f31e37"
	ptkDesperateChargeOracle         = "26074e44-cf2b-459f-9a5d-7eed231ed2b4"
)

// --- every card is in the catalog ----------------------------------------

func TestPtk5622EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		ptkShuEliteInfantryOracle, ptkShuFootSoldiersOracle, ptkStrawSoldiersOracle,
		ptkWuInfantryOracle, ptkWeiInfantryOracle, ptkBarbarianHordeOracle,
		ptkIndependentTroopsOracle, ptkForestBearOracle, ptkMengHuosHordeOracle,
		ptkSouthernElephantOracle, ptkTrainedJackalOracle, ptkAlertShuInfantryOracle,
		ptkShuCavalryOracle, ptkShuEliteCompanionsOracle, ptkLuMengWuGeneralOracle,
		ptkWuEliteCavalryOracle, ptkWuLightCavalryOracle, ptkWeiEliteCompanionsOracle,
		ptkWeiScoutOracle, ptkWeiStrikeForceOracle, ptkBarbarianGeneralOracle,
		ptkLadyZhurongWarriorQueenOracle, ptkShuGeneralOracle, ptkZhangFeiFierceWarriorOracle,
		ptkLuBuMasterAtArmsOracle, ptkMountainBanditOracle, ptkRenegadeTroopsOracle,
		ptkTaoistHermitOracle, ptkZodiacPigOracle, ptkZodiacRatOracle,
		ptkZodiacSnakeOracle, ptkZodiacOxOracle, ptkZodiacDogOracle,
		ptkZodiacGoatOracle, ptkZodiacHorseOracle, ptkZodiacMonkeyOracle,
		ptkZodiacRabbitOracle, ptkZodiacTigerOracle, ptkZodiacRoosterOracle,
		ptkYoungWeiRecruitsOracle, ptkYellowScarvesTroopsOracle, ptkYellowScarvesCavalryOracle,
		ptkYellowScarvesGeneralOracle, ptkBrilliantPlanOracle, ptkSpringOfEternalPeaceOracle,
		ptkFireAmbushOracle, ptkGhostlyVisitOracle, ptkWieldingTheGreenDragonOracle,
		ptkVirtuousChargeOracle, ptkDesperateChargeOracle,
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

// --- vanilla bodies ----------------------------------------------------------

// A vanilla creature declares nothing: no keyword, no restriction, and it
// can attack and block like any creature.
func TestPtk5622VanillaCreaturesDeclareNothing(t *testing.T) {
	for _, oracle := range []string{
		ptkShuEliteInfantryOracle, ptkShuFootSoldiersOracle, ptkStrawSoldiersOracle,
		ptkWuInfantryOracle, ptkWeiInfantryOracle, ptkBarbarianHordeOracle,
		ptkIndependentTroopsOracle, ptkForestBearOracle, ptkMengHuosHordeOracle,
		ptkSouthernElephantOracle, ptkTrainedJackalOracle,
	} {
		spec, _ := Lookup(oracle)
		if len(spec.PrintedKeywords) != 0 || len(spec.Static) != 0 || len(spec.Triggered) != 0 ||
			len(spec.Activated) != 0 || spec.OnResolve != nil {
			t.Errorf("%s declares rules text it doesn't print", spec.Name)
		}
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bear := b12Push(g, opp.ID, "Forest Bear", "Creature — Bear", ptkForestBearOracle, 2, 2)
	if err := porBlock(t, g, attacker, opp.ID, bear); err != nil {
		t.Errorf("Forest Bear couldn't block: %v", err)
	}
}

// --- keyword bodies ----------------------------------------------------------

func TestPtk5622KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Alert Shu Infantry", "Creature — Human Soldier", ptkAlertShuInfantryOracle, []string{"vigilance"}},
		{"Shu Cavalry", "Creature — Human Soldier", ptkShuCavalryOracle, []string{"horsemanship"}},
		{"Shu Elite Companions", "Creature — Human Soldier", ptkShuEliteCompanionsOracle, []string{"horsemanship"}},
		{"Lu Meng, Wu General", "Legendary Creature — Human Soldier", ptkLuMengWuGeneralOracle, []string{"horsemanship"}},
		{"Wu Elite Cavalry", "Creature — Human Soldier", ptkWuEliteCavalryOracle, []string{"horsemanship"}},
		{"Wu Light Cavalry", "Creature — Human Soldier", ptkWuLightCavalryOracle, []string{"horsemanship"}},
		{"Wei Elite Companions", "Creature — Human Soldier", ptkWeiEliteCompanionsOracle, []string{"horsemanship"}},
		{"Wei Scout", "Creature — Human Soldier Scout", ptkWeiScoutOracle, []string{"horsemanship"}},
		{"Wei Strike Force", "Creature — Human Soldier", ptkWeiStrikeForceOracle, []string{"horsemanship"}},
		{"Barbarian General", "Creature — Human Barbarian Soldier", ptkBarbarianGeneralOracle, []string{"horsemanship"}},
		{"Lady Zhurong, Warrior Queen", "Legendary Creature — Human Soldier Warrior", ptkLadyZhurongWarriorQueenOracle, []string{"horsemanship"}},
		{"Shu General", "Creature — Human Soldier", ptkShuGeneralOracle, []string{"vigilance", "horsemanship"}},
		{"Zhang Fei, Fierce Warrior", "Legendary Creature — Human Soldier Warrior", ptkZhangFeiFierceWarriorOracle, []string{"vigilance", "horsemanship"}},
		{"Lu Bu, Master-at-Arms", "Legendary Creature — Human Soldier Warrior", ptkLuBuMasterAtArmsOracle, []string{"haste", "horsemanship"}},
		{"Mountain Bandit", "Creature — Human Soldier Rogue", ptkMountainBanditOracle, []string{"haste"}},
		{"Renegade Troops", "Creature — Human Soldier", ptkRenegadeTroopsOracle, []string{"haste"}},
		{"Taoist Hermit", "Creature — Human Mystic", ptkTaoistHermitOracle, []string{"hexproof"}},
		{"Zodiac Pig", "Creature — Boar", ptkZodiacPigOracle, []string{"swampwalk"}},
		{"Zodiac Rat", "Creature — Rat", ptkZodiacRatOracle, []string{"swampwalk"}},
		{"Zodiac Snake", "Creature — Snake", ptkZodiacSnakeOracle, []string{"swampwalk"}},
		{"Zodiac Ox", "Creature — Ox", ptkZodiacOxOracle, []string{"swampwalk"}},
		{"Zodiac Dog", "Creature — Dog", ptkZodiacDogOracle, []string{"mountainwalk"}},
		{"Zodiac Goat", "Creature — Goat", ptkZodiacGoatOracle, []string{"mountainwalk"}},
		{"Zodiac Horse", "Creature — Horse", ptkZodiacHorseOracle, []string{"islandwalk"}},
		{"Zodiac Monkey", "Creature — Monkey", ptkZodiacMonkeyOracle, []string{"forestwalk"}},
		{"Zodiac Rabbit", "Creature — Rabbit", ptkZodiacRabbitOracle, []string{"forestwalk"}},
		{"Zodiac Tiger", "Creature — Cat", ptkZodiacTigerOracle, []string{"forestwalk"}},
		{"Zodiac Rooster", "Creature — Bird", ptkZodiacRoosterOracle, []string{"plainswalk"}},
		{"Yellow Scarves Cavalry", "Creature — Human Soldier", ptkYellowScarvesCavalryOracle, []string{"horsemanship"}},
		{"Yellow Scarves General", "Creature — Human Soldier", ptkYellowScarvesGeneralOracle, []string{"horsemanship"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			for _, kw := range tc.keywords {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Horsemanship: a creature without it can't block, one with it can, and
// horsemanship doesn't stop its creature blocking an ordinary attacker.
func TestPtk5622HorsemanshipIsBlockedOnlyByHorsemanship(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cavalry := b12Push(g, me.ID, "Shu Cavalry", "Creature — Human Soldier", ptkShuCavalryOracle, 2, 2)
	bear := b12Push(g, opp.ID, "Forest Bear", "Creature — Bear", ptkForestBearOracle, 2, 2)
	if err := porBlock(t, g, cavalry, opp.ID, bear); err == nil {
		t.Error("a creature without horsemanship blocked Shu Cavalry")
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	cavalry = b12Push(g, me.ID, "Shu Cavalry", "Creature — Human Soldier", ptkShuCavalryOracle, 2, 2)
	light := b12Push(g, opp.ID, "Wu Light Cavalry", "Creature — Human Soldier", ptkWuLightCavalryOracle, 1, 2)
	if err := porBlock(t, g, cavalry, opp.ID, light); err != nil {
		t.Errorf("Wu Light Cavalry (horsemanship) couldn't block Shu Cavalry: %v", err)
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	scout := b12Push(g, opp.ID, "Wei Scout", "Creature — Human Soldier Scout", ptkWeiScoutOracle, 1, 1)
	if err := porBlock(t, g, attacker, opp.ID, scout); err != nil {
		t.Errorf("Wei Scout couldn't block a creature without horsemanship: %v", err)
	}
}

// Vigilance: Alert Shu Infantry and Shu General attack without tapping.
func TestPtk5622VigilanceAttacksWithoutTapping(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Alert Shu Infantry", ptkAlertShuInfantryOracle},
		{"Shu General", ptkShuGeneralOracle},
		{"Zhang Fei, Fierce Warrior", ptkZhangFeiFierceWarriorOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := b12Push(g, me.ID, tc.name, "Creature — Human Soldier", tc.oracle, 2, 2)
			advanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(id, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
			lockInAttacks(t, g)
			if alliTapped(t, g, id) {
				t.Errorf("%s tapped to attack", tc.name)
			}
		})
	}
}

// Haste: each attacks the turn it arrives; a creature without haste
// cast the same turn can't.
func TestPtk5622HasteAttacksTheTurnItArrives(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
	}{
		{"Mountain Bandit", "Creature — Human Soldier Rogue", ptkMountainBanditOracle, 1, 1},
		{"Renegade Troops", "Creature — Human Soldier", ptkRenegadeTroopsOracle, 4, 2},
		{"Lu Bu, Master-at-Arms", "Legendary Creature — Human Soldier Warrior", ptkLuBuMasterAtArmsOracle, 4, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			hasty := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			troops := castHoldCreature(t, g, "Independent Troops", "Creature — Human Soldier", ptkIndependentTroopsOracle, 2, 1)
			advanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(troops, opp.ID); err == nil {
				t.Error("a summoning-sick creature without haste attacked")
			}
			if err := g.DeclareAttacker(hasty, opp.ID); err != nil {
				t.Errorf("%s couldn't attack the turn it arrived: %v", tc.name, err)
			}
		})
	}
}

// Hexproof: an opponent's spell can't target Taoist Hermit; its
// controller's own spell can.
func TestPtk5622TaoistHermitCantBeTargetedByAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Push(g, opp.ID, "Taoist Hermit", "Creature — Human Mystic", ptkTaoistHermitOracle, 2, 2)
	if err := castCatalogSpellErr(t, g, "Fire Ambush", "Sorcery", ptkFireAmbushOracle, cardTarget(theirs)); err == nil {
		t.Error("an opponent's Fire Ambush targeted Taoist Hermit")
	}

	g = newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := b12Push(g, me.ID, "Taoist Hermit", "Creature — Human Mystic", ptkTaoistHermitOracle, 2, 2)
	castCatalogSpell(t, g, "Wielding the Green Dragon", "Sorcery", ptkWieldingTheGreenDragonOracle, cardTarget(mine))
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, mine); p != 6 || tt != 6 {
		t.Errorf("Taoist Hermit is %d/%d after its controller's pump, want 6/6", p, tt)
	}
}

// Landwalk: unblockable while the defender controls the land type,
// blockable when they don't.
func TestPtk5622LandwalkersCantBeBlockedWhileTheDefenderHasTheLand(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle, landType string
	}{
		{"Zodiac Pig", "Creature — Boar", ptkZodiacPigOracle, "Swamp"},
		{"Zodiac Rat", "Creature — Rat", ptkZodiacRatOracle, "Swamp"},
		{"Zodiac Snake", "Creature — Snake", ptkZodiacSnakeOracle, "Swamp"},
		{"Zodiac Ox", "Creature — Ox", ptkZodiacOxOracle, "Swamp"},
		{"Zodiac Dog", "Creature — Dog", ptkZodiacDogOracle, "Mountain"},
		{"Zodiac Goat", "Creature — Goat", ptkZodiacGoatOracle, "Mountain"},
		{"Zodiac Horse", "Creature — Horse", ptkZodiacHorseOracle, "Island"},
		{"Zodiac Monkey", "Creature — Monkey", ptkZodiacMonkeyOracle, "Forest"},
		{"Zodiac Rabbit", "Creature — Rabbit", ptkZodiacRabbitOracle, "Forest"},
		{"Zodiac Tiger", "Creature — Cat", ptkZodiacTigerOracle, "Forest"},
		{"Zodiac Rooster", "Creature — Bird", ptkZodiacRoosterOracle, "Plains"},
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

// "This creature can't block": the restriction is declared, and the block
// is refused.
func TestPtk5622CantBlock(t *testing.T) {
	cases := []struct {
		name, oracle string
	}{
		{"Young Wei Recruits", ptkYoungWeiRecruitsOracle},
		{"Yellow Scarves Troops", ptkYellowScarvesTroopsOracle},
		{"Yellow Scarves Cavalry", ptkYellowScarvesCavalryOracle},
		{"Yellow Scarves General", ptkYellowScarvesGeneralOracle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := b12Push(g, opp.ID, tc.name, "Creature — Human Soldier", tc.oracle, 2, 2)
			assertRestrictions(t, g, id, game.CantBlock)
			attacker := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			if err := porBlock(t, g, attacker, opp.ID, id); err == nil {
				t.Errorf("%s blocked", tc.name)
			}
		})
	}
}

// --- spells ------------------------------------------------------------------

func TestPtk5622BrilliantPlanDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Brilliant Plan", "Sorcery", ptkBrilliantPlanOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d -> %d, want +3", hand, me.Hand.Size())
	}
}

func TestPtk5622SpringOfEternalPeaceGainsEight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castCatalogSpell(t, g, "Spring of Eternal Peace", "Sorcery", ptkSpringOfEternalPeaceOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+8 {
		t.Errorf("life %d -> %d, want +8", life, me.Life)
	}
}

func TestPtk5622FireAmbushHitsAPlayerOrACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	castCatalogSpell(t, g, "Fire Ambush", "Sorcery", ptkFireAmbushOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, opp.Life)
	}

	g = newCatalogGame(t)
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elephant := b12Push(g, opp.ID, "Southern Elephant", "Creature — Elephant", ptkSouthernElephantOracle, 3, 4)
	soldiers := b12Push(g, opp.ID, "Shu Foot Soldiers", "Creature — Human Soldier", ptkShuFootSoldiersOracle, 2, 3)
	castCatalogSpell(t, g, "Fire Ambush", "Sorcery", ptkFireAmbushOracle, cardTarget(soldiers))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(soldiers) {
		t.Error("Fire Ambush left a 3-toughness creature alive")
	}
	if !g.Battlefield.Contains(elephant) {
		t.Error("Fire Ambush hit a creature it didn't target")
	}
}

func TestPtk5622GhostlyVisitDestroysOnlyANonblackCreature(t *testing.T) {
	g := newCatalogGame(t)
	black := b12Push(g, g.Seats[1].ID, "Wei Infantry", "Creature — Human Soldier", ptkWeiInfantryOracle, 2, 1)
	alliSetColors(g, black, "B")
	if err := castCatalogSpellErr(t, g, "Ghostly Visit", "Sorcery", ptkGhostlyVisitOracle, cardTarget(black)); err == nil {
		t.Error("Ghostly Visit targeted a black creature")
	}

	g = newCatalogGame(t)
	green := b12Push(g, g.Seats[1].ID, "Forest Bear", "Creature — Bear", ptkForestBearOracle, 2, 2)
	alliSetColors(g, green, "G")
	castCatalogSpell(t, g, "Ghostly Visit", "Sorcery", ptkGhostlyVisitOracle, cardTarget(green))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the green creature survived Ghostly Visit")
	}
}

func TestPtk5622WieldingTheGreenDragonPumpsPlusFour(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Push(g, g.Seats[0].ID, "Forest Bear", "Creature — Bear", ptkForestBearOracle, 2, 2)
	castCatalogSpell(t, g, "Wielding the Green Dragon", "Sorcery", ptkWieldingTheGreenDragonOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 6 || tt != 6 {
		t.Errorf("Forest Bear is %d/%d, want 6/6", p, tt)
	}
}

func TestPtk5622TeamPumpsOnlyYourCreatures(t *testing.T) {
	cases := []struct {
		name, oracle string
		power, tough int
	}{
		{"Virtuous Charge", ptkVirtuousChargeOracle, 3, 3},
		{"Desperate Charge", ptkDesperateChargeOracle, 4, 2},
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
