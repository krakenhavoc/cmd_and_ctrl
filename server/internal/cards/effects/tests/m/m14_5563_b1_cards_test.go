package testsm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m14_5563_b1_cards_test.go — the first batch of Magic 2014 cards (#5563).
// Every test plays the real catalog entry through the engine.

const (
	m14B1KalonianTuskerOracle      = "df7f697e-6886-4897-a024-61ae225c1b34"
	m14B1MinotaurAbominationOracle = "017568ea-613f-4057-9a8d-61b087d6dc92"
	m14B1RegathanFirecatOracle     = "b3f609b5-fb4e-4e66-a70a-53bbe0fa2b9a"
	m14B1RumblingBalothOracle      = "3191c6ca-4d25-4ba3-bfe1-4aeab1295573"
	m14B1SliverConstructOracle     = "7d592d5b-95cf-4416-9ade-1c23a8ba01d1"
	m14B1UndeadMinotaurOracle      = "19b25d09-563f-4159-9ce0-c0343564bd3b"
	m14B1AccursedSpiritOracle      = "0e20afd2-507b-4878-a78c-f23b51ec071f"
	m14B1SeacoastDrakeOracle       = "2f4b5831-57a2-4778-90c1-0b8ec49afd46"
	m14B1DawnstrikePaladinOracle   = "c2bfb230-6949-4d5e-8cac-9d7dafdc5e9a"
	m14B1DeathgazeCockatriceOracle = "0fabb6e4-35cf-4e56-a27c-ff627d8e9b7a"
	m14B1BattleSliverOracle        = "d513e5df-31b0-4756-bf47-403460d0068e"
	m14B1MeganticSliverOracle      = "52e8ec03-23ed-43a0-bce2-ad0cd5c4d737"
	m14B1PredatorySliverOracle     = "6b589e70-9c1d-48dd-a597-479959e692bf"
	m14B1SteelformSliverOracle     = "ae58e71c-b617-404f-a5dc-5e6a487f2675"
	m14B1BlurSliverOracle          = "e0f05979-daba-43eb-b356-24284e7ca99c"
	m14B1GroundshakerSliverOracle  = "106e343f-c6f4-4155-ad01-63e482d4d38e"
	m14B1StrikingSliverOracle      = "6a15599d-8b00-4e0f-ac0c-d05f07c1e50d"
	m14B1SyphonSliverOracle        = "61f18560-8f26-4d7b-a6f8-ec388766123b"
	m14B1ThorncasterSliverOracle   = "70623348-7eda-4c68-89cd-2356b0896b79"
	m14B1ChargingGriffinOracle     = "1c99a5cc-d974-4a4b-a03b-a78bfe738ed7"
	m14B1MessengerDrakeOracle      = "4c62dce6-dc1f-4d2d-95dd-295810ada6a6"
	m14B1SoulmenderOracle          = "41b1d850-2fda-490b-a4f6-0d9bfb37103c"
	m14B1BloodBairnOracle          = "33bea2f0-c5aa-44b5-8714-5f50f7f2bb05"
	m14B1VampireWarlordOracle      = "393586d3-1f1a-4f3a-96b8-ed1a8fdf00c6"
	m14B1CorpseHaulerOracle        = "c2ba1c3b-08c7-4835-81f4-2f98843fb7b8"
	m14B1StonehornChanterOracle    = "22c5187b-e3db-4221-a173-e3b83a581fca"
	m14B1VialOfPoisonOracle        = "3d51077f-4ce5-475a-9e65-e9076d779825"
	m14B1ZephyrChargeOracle        = "576128ba-28aa-4ef9-9e72-8ca13f42c9b4"
	m14B1EncroachingWastesOracle   = "43144f06-079b-4515-a03a-01ea3e90d586"
	m14B1CyclopsTyrantOracle       = "40ac0cb0-bdf7-447d-9875-0eb3cc5d0adc"
	m14B1AcademyRaiderOracle       = "75131d75-0703-44d0-b503-35190be8e66f"
	m14B1LiturgyOfBloodOracle      = "aabea4a2-5107-4abd-a0c7-ac0024d98ae0"
	m14B1HiveStirringsOracle       = "c7dd0614-e11e-4b6b-a2cb-9904e457148f"
	m14B1HuntTheWeakOracle         = "37864d13-2aaa-477b-8453-a42f77ab4437"
	m14B1GlimpseTheFutureOracle    = "ff5a2bcf-efd0-4bdb-af52-21a93447c54c"
	m14B1MasterOfDiversionOracle   = "3402aa73-5e63-4158-b114-6f1d42ec8992"
	m14B1TrainedCondorOracle       = "99742ec7-7b67-4301-9307-3c2068046e33"
	m14B1ImposingSovereignOracle   = "4bba3663-7f3c-4589-88aa-cf0bfe36b171"
	m14B1FleshpulperGiantOracle    = "7cdeecb5-354a-4a6c-989f-d8c9554befb7"
	m14B1XathridNecromancerOracle  = "a6b38874-c8be-47f7-a656-ba49ec9e967f"
	m14B1StaffDeathOracle          = "c3fda16e-fd56-434e-91ac-561ae0482f43"
	m14B1StaffFlameOracle          = "fc71710a-8294-4370-89da-7ab1c2b38d19"
	m14B1StaffMindOracle           = "1466687f-cfe1-4e6e-ad8e-dca5e8e4c06d"
	m14B1StaffSunOracle            = "bd8bc927-e50a-4fab-9803-0602a841c7c2"
	m14B1StaffWildOracle           = "6bdff38d-0a52-42e5-86ea-edd7029b44f6"
	m14B1RingOfThreeWishesOracle   = "677c3616-295f-493a-8b60-1e22b0224cfa"
	m14B1DragonEggOracle           = "8f561498-66e7-4cbf-8cb9-e6fd4717993d"
	m14B1BarrageOracle             = "da7600f6-5d4c-46bf-936a-9e0f0d508e19"
	m14B1BlightcasterOracle        = "b40e2aae-3a1d-4bc2-bf67-e821ec800434"
	m14B1MaraudingMaulhornOracle   = "8f4f0cd2-9777-4ea1-944d-306347dc48a2"

	m14B1SliverType = "Creature — Sliver"
	m14B1BearType   = "Creature — Bear"
)

// --- every card is in the catalog ----------------------------------------

func TestM14B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		m14B1KalonianTuskerOracle, m14B1MinotaurAbominationOracle, m14B1RegathanFirecatOracle, m14B1RumblingBalothOracle,
		m14B1SliverConstructOracle, m14B1UndeadMinotaurOracle, m14B1AccursedSpiritOracle, m14B1SeacoastDrakeOracle,
		m14B1DawnstrikePaladinOracle, m14B1DeathgazeCockatriceOracle, m14B1BattleSliverOracle, m14B1MeganticSliverOracle,
		m14B1PredatorySliverOracle, m14B1SteelformSliverOracle, m14B1BlurSliverOracle, m14B1GroundshakerSliverOracle,
		m14B1StrikingSliverOracle, m14B1SyphonSliverOracle, m14B1ThorncasterSliverOracle, m14B1ChargingGriffinOracle,
		m14B1MessengerDrakeOracle, m14B1SoulmenderOracle, m14B1BloodBairnOracle, m14B1VampireWarlordOracle,
		m14B1CorpseHaulerOracle, m14B1StonehornChanterOracle, m14B1VialOfPoisonOracle, m14B1ZephyrChargeOracle,
		m14B1EncroachingWastesOracle, m14B1CyclopsTyrantOracle, m14B1AcademyRaiderOracle, m14B1LiturgyOfBloodOracle,
		m14B1HiveStirringsOracle, m14B1HuntTheWeakOracle, m14B1GlimpseTheFutureOracle, m14B1MasterOfDiversionOracle,
		m14B1TrainedCondorOracle, m14B1ImposingSovereignOracle, m14B1FleshpulperGiantOracle, m14B1XathridNecromancerOracle,
		m14B1StaffDeathOracle, m14B1StaffFlameOracle, m14B1StaffMindOracle, m14B1StaffSunOracle,
		m14B1StaffWildOracle, m14B1RingOfThreeWishesOracle, m14B1DragonEggOracle, m14B1BarrageOracle,
		m14B1BlightcasterOracle, m14B1MaraudingMaulhornOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- printed keywords -------------------------------------------------------

func TestM14B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Accursed Spirit", "Creature — Spirit", m14B1AccursedSpiritOracle, []string{"intimidate"}},
		{"Seacoast Drake", "Creature — Drake", m14B1SeacoastDrakeOracle, []string{"flying"}},
		{"Dawnstrike Paladin", "Creature — Human Knight", m14B1DawnstrikePaladinOracle, []string{"vigilance", "lifelink"}},
		{"Deathgaze Cockatrice", "Creature — Cockatrice", m14B1DeathgazeCockatriceOracle, []string{"flying", "deathtouch"}},
		{"Cyclops Tyrant", "Creature — Cyclops", m14B1CyclopsTyrantOracle, []string{"intimidate"}},
		{"Academy Raider", "Creature — Human Warrior", m14B1AcademyRaiderOracle, []string{"intimidate"}},
		{"Charging Griffin", "Creature — Griffin", m14B1ChargingGriffinOracle, []string{"flying"}},
		{"Messenger Drake", "Creature — Drake", m14B1MessengerDrakeOracle, []string{"flying"}},
		{"Trained Condor", "Creature — Bird", m14B1TrainedCondorOracle, []string{"flying"}},
		{"Dragon Egg", "Creature — Dragon Egg", m14B1DragonEggOracle, []string{"defender"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- the Slivers --------------------------------------------------------------

// Each anthem reaches every Sliver creature its controller controls, the
// lord included, and no opponent's Sliver or non-Sliver.
func TestM14B1SliverAnthems(t *testing.T) {
	for _, tc := range []struct {
		name, oracle       string
		power, toughness   int
		basePow, baseTough int
	}{
		{"Battle Sliver", m14B1BattleSliverOracle, 2, 0, 3, 3},
		{"Megantic Sliver", m14B1MeganticSliverOracle, 3, 3, 3, 3},
		{"Predatory Sliver", m14B1PredatorySliverOracle, 1, 1, 1, 1},
		{"Steelform Sliver", m14B1SteelformSliverOracle, 0, 1, 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			lord := effectstest.B12Push(g, me.ID, tc.name, m14B1SliverType, tc.oracle, tc.basePow, tc.baseTough)
			mine := effectstest.B12Creature(g, me.ID, "Drone", m14B1SliverType, 1, 1)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Drone", m14B1SliverType, 1, 1)
			bear := effectstest.B12Creature(g, me.ID, "Bear", m14B1BearType, 2, 2)
			if p, tough := effectstest.PtOf(t, g, lord); p != tc.basePow+tc.power || tough != tc.baseTough+tc.toughness {
				t.Errorf("the lord is %d/%d, want %d/%d", p, tough, tc.basePow+tc.power, tc.baseTough+tc.toughness)
			}
			if p, tough := effectstest.PtOf(t, g, mine); p != 1+tc.power || tough != 1+tc.toughness {
				t.Errorf("your Sliver is %d/%d, want %d/%d", p, tough, 1+tc.power, 1+tc.toughness)
			}
			if p, tough := effectstest.PtOf(t, g, theirs); p != 1 || tough != 1 {
				t.Errorf("an opponent's Sliver is %d/%d, want 1/1", p, tough)
			}
			if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
				t.Errorf("your non-Sliver is %d/%d, want 2/2", p, tough)
			}
		})
	}
}

func TestM14B1SliverKeywordGrants(t *testing.T) {
	for _, tc := range []struct{ name, oracle, keyword string }{
		{"Blur Sliver", m14B1BlurSliverOracle, "haste"},
		{"Groundshaker Sliver", m14B1GroundshakerSliverOracle, "trample"},
		{"Striking Sliver", m14B1StrikingSliverOracle, "first strike"},
		{"Syphon Sliver", m14B1SyphonSliverOracle, "lifelink"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			lord := effectstest.B12Push(g, me.ID, tc.name, m14B1SliverType, tc.oracle, 2, 2)
			mine := effectstest.B12Creature(g, me.ID, "Drone", m14B1SliverType, 1, 1)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Drone", m14B1SliverType, 1, 1)
			bear := effectstest.B12Creature(g, me.ID, "Bear", m14B1BearType, 2, 2)
			if !effectstest.HasEffectiveKeyword(t, g, lord, tc.keyword) {
				t.Errorf("the lord has no %s", tc.keyword)
			}
			if !effectstest.HasEffectiveKeyword(t, g, mine, tc.keyword) {
				t.Errorf("your Sliver has no %s", tc.keyword)
			}
			if effectstest.HasEffectiveKeyword(t, g, theirs, tc.keyword) {
				t.Errorf("an opponent's Sliver has %s", tc.keyword)
			}
			if effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("your non-Sliver has %s", tc.keyword)
			}
		})
	}
}

// Thorncaster Sliver grants its trigger: another Sliver of yours attacking
// deals the 1 damage.
func TestM14B1ThorncasterSliverGrantsTheAttackPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Thorncaster Sliver", m14B1SliverType, m14B1ThorncasterSliverOracle, 2, 2)
	drone := effectstest.B12Creature(g, me.ID, "Drone", m14B1SliverType, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m14B1BearType, 1, 3)
	effectstest.DeclareAttack(t, g, opp.ID, drone)
	effectstest.PickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, bear).DamageMarked; got != 1 {
		t.Errorf("the target has %d damage, want 1", got)
	}
}

// --- triggered abilities --------------------------------------------------------

func TestM14B1ChargingGriffinGrowsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	griffin := effectstest.B12Push(g, me.ID, "Charging Griffin", "Creature — Griffin", m14B1ChargingGriffinOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, griffin)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, griffin); p != 3 || tough != 3 {
		t.Errorf("the Griffin is %d/%d, want 3/3", p, tough)
	}
}

func TestM14B1MessengerDrakeDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	drake := effectstest.B12Push(g, me.ID, "Messenger Drake", "Creature — Drake", m14B1MessengerDrakeOracle, 3, 3)
	hand := me.Hand.Size()
	effectstest.B27Kill(g, drake)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestM14B1DragonEggHatchesADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	egg := effectstest.B12Push(g, me.ID, "Dragon Egg", "Creature — Dragon Egg", m14B1DragonEggOracle, 0, 2)
	effectstest.B27Kill(g, egg)
	effectstest.PassPriorityAroundTable(t, g)
	dragons := effectstest.TokensNamed(g, "Dragon")
	if len(dragons) != 1 {
		t.Fatalf("%d Dragon tokens, want 1", len(dragons))
	}
	if dragons[0].Controller != me.ID || !effectstest.HasEffectiveKeyword(t, g, dragons[0].InstanceID, "flying") {
		t.Error("the Dragon is not a flier under the Egg's controller")
	}
}

func TestM14B1AcademyRaiderRummagesOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := effectstest.B12Push(g, me.ID, "Academy Raider", "Creature — Human Warrior", m14B1AcademyRaiderOracle, 1, 1)
	effectstest.D2dSetHand(me, 2)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.DealCombatDamage(g, me.ID, raider, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt")
	}
	hand, grave := me.Hand.Size(), me.Graveyard.Size()
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.MonSettle(t, g)
	if me.Hand.Size() != hand || me.Graveyard.Size() != grave+1 {
		t.Errorf("hand %d graveyard %d, want %d and %d", me.Hand.Size(), me.Graveyard.Size(), hand, grave+1)
	}
}

func TestM14B1MasterOfDiversionTapsADefendersCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := effectstest.B12Push(g, me.ID, "Master of Diversion", "Creature — Human Scout", m14B1MasterOfDiversionOracle, 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", m14B1BearType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, master)
	effectstest.PickIfAsked(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, victim) {
		t.Error("the defending player's creature is untapped")
	}
}

func TestM14B1TrainedCondorGivesAnotherCreatureFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	condor := effectstest.B12Push(g, me.ID, "Trained Condor", "Creature — Bird", m14B1TrainedCondorOracle, 2, 1)
	ally := effectstest.B12Creature(g, me.ID, "Ally", m14B1BearType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, condor)
	effectstest.PickIfAsked(t, g, me.ID, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, ally, "flying") {
		t.Error("the other creature has no flying")
	}
}

func TestM14B1ImposingSovereignTapsOpponentsCreaturesOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Imposing Sovereign", "Creature — Human Noble", m14B1ImposingSovereignOracle, 2, 1)
	effectstest.AdvanceToMainOf(t, g, 1)
	theirs := effectstest.B20CastCreature(t, g, opp, "Their Bear", m14B1BearType, "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B31Tapped(t, g, theirs) {
		t.Error("an opponent's creature entered untapped")
	}
}

func TestM14B1FleshpulperGiantDestroysASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := effectstest.B12Creature(g, opp.ID, "Small", m14B1BearType, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", m14B1BearType, 3, 3)
	effectstest.CastHoldCreature(t, g, "Fleshpulper Giant", "Creature — Giant", m14B1FleshpulperGiantOracle, 4, 4)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, big) {
		t.Error("a creature with toughness 3 is offered")
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the toughness 2 creature survived")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the toughness 3 creature was destroyed")
	}
}

// The Necromancer's own death and a Human of yours dying each make a tapped
// Zombie; a non-Human and an opponent's Human do not.
func TestM14B1XathridNecromancerMakesZombiesForHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	necro := effectstest.B12Push(g, me.ID, "Xathrid Necromancer", "Creature — Human Wizard", m14B1XathridNecromancerOracle, 2, 2)
	human := effectstest.B12Creature(g, me.ID, "Villager", "Creature — Human Peasant", 1, 1)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Villager", "Creature — Human", 1, 1)
	for _, id := range []uuid.UUID{elf, theirs} {
		effectstest.B27Kill(g, id)
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.TokensNamedControlledBy(g, "Zombie", me.ID); n != 0 {
		t.Fatalf("%d Zombies after a non-Human and an opponent's Human died, want 0", n)
	}
	effectstest.B27Kill(g, human)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Zombie", me.ID); n != 1 {
		t.Fatalf("%d Zombies after your Human died, want 1", n)
	}
	effectstest.B27Kill(g, necro)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Zombie", me.ID); n != 2 {
		t.Fatalf("%d Zombies after the Necromancer died, want 2", n)
	}
	for _, z := range effectstest.TokensNamed(g, "Zombie") {
		if !z.Tapped {
			t.Error("a Zombie entered untapped")
		}
	}
}

// A spell of the Staff's color and a land of its type each gain 1 life; a
// spell of another color does not.
func TestM14B1StaffsGainLife(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, color, other, land string
	}{
		{"Staff of the Death Magus", m14B1StaffDeathOracle, "B", "W", "Swamp"},
		{"Staff of the Flame Magus", m14B1StaffFlameOracle, "R", "U", "Mountain"},
		{"Staff of the Mind Magus", m14B1StaffMindOracle, "U", "R", "Island"},
		{"Staff of the Sun Magus", m14B1StaffSunOracle, "W", "B", "Plains"},
		{"Staff of the Wild Magus", m14B1StaffWildOracle, "G", "B", "Forest"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Artifact", tc.oracle, false)
			life := me.Life
			effectstest.AvB1Cast(t, g, me, game.Card{Name: "Off Spell", TypeLine: "Sorcery", ManaCost: "{" + tc.other + "}", Colors: []string{tc.other}}, game.CastSpellParams{})
			effectstest.PassPriorityAroundTable(t, g)
			if me.Life != life {
				t.Fatalf("a spell of another color changed life to %d, want %d", me.Life, life)
			}
			effectstest.AvB1Cast(t, g, me, game.Card{Name: "On Spell", TypeLine: "Sorcery", ManaCost: "{" + tc.color + "}", Colors: []string{tc.color}}, game.CastSpellParams{})
			effectstest.PassPriorityAroundTable(t, g)
			if me.Life != life+1 {
				t.Fatalf("after a spell of its color life is %d, want %d", me.Life, life+1)
			}
			effectstest.B12PlayFromHand(t, g, tc.land, "Basic Land — "+tc.land, "", game.CastSpellParams{})
			effectstest.PassPriorityAroundTable(t, g)
			if me.Life != life+2 {
				t.Errorf("after a %s entered life is %d, want %d", tc.land, me.Life, life+2)
			}
		})
	}
}

// Blightcaster's trigger on an enchantment spell gives -2/-2.
func TestM14B1BlightcasterShrinksOnAnEnchantmentSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Blightcaster", "Creature — Human Wizard", m14B1BlightcasterOracle, 2, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m14B1BearType, 3, 3)
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Charm", TypeLine: "Enchantment"}, game.CastSpellParams{})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 1 || tough != 1 {
		t.Errorf("the Bear is %d/%d, want 1/1", p, tough)
	}
}

// --- activated abilities -------------------------------------------------------

func TestM14B1SoulmenderGainsOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cleric := effectstest.B12Push(g, me.ID, "Soulmender", "Creature — Human Cleric", m14B1SoulmenderOracle, 1, 1)
	life := me.Life
	effectstest.ApaActivate(t, g, me, cleric, 0, game.ActivateAbilityParams{})
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	if !effectstest.HmlTapped(t, g, cleric) {
		t.Error("Soulmender did not tap")
	}
}

func TestM14B1BloodBairnSacrificesToGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bairn := effectstest.B12Push(g, me.ID, "Blood Bairn", "Creature — Vampire", m14B1BloodBairnOracle, 2, 2)
	victim := effectstest.B12Creature(g, me.ID, "Victim", m14B1BearType, 1, 1)
	effectstest.ApaActivate(t, g, me, bairn, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{victim}})
	if g.Battlefield.Contains(victim) {
		t.Error("the other creature was not sacrificed")
	}
	if p, tough := effectstest.PtOf(t, g, bairn); p != 4 || tough != 4 {
		t.Errorf("Blood Bairn is %d/%d, want 4/4", p, tough)
	}
}

func TestM14B1VampireWarlordRegeneratesForASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	warlord := effectstest.B12Push(g, me.ID, "Vampire Warlord", "Creature — Vampire Warrior", m14B1VampireWarlordOracle, 4, 2)
	victim := effectstest.B12Creature(g, me.ID, "Victim", m14B1BearType, 1, 1)
	effectstest.ApaActivate(t, g, me, warlord, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{victim}})
	if g.Battlefield.Contains(victim) {
		t.Fatal("the other creature was not sacrificed")
	}
	effectstest.B27Kill(g, warlord)
	if !g.Battlefield.Contains(warlord) {
		t.Error("the shielded Warlord was destroyed")
	}
}

func TestM14B1CorpseHaulerReturnsACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hauler := effectstest.B12Push(g, me.ID, "Corpse Hauler", "Creature — Human Rogue", m14B1CorpseHaulerOracle, 2, 1)
	bear := effectstest.WthGraveyardCard(me, "Bear", m14B1BearType)
	rock := effectstest.WthGraveyardCard(me, "Rock", "Artifact")
	effectstest.ApaMana(me, "C", "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, hauler, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(rock)}); err == nil {
		t.Fatal("Corpse Hauler targeted an artifact card")
	}
	effectstest.ApaActivate(t, g, me, hauler, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !me.Hand.Contains(bear) {
		t.Error("the creature card did not return to hand")
	}
	if g.Battlefield.Contains(hauler) {
		t.Error("Corpse Hauler was not sacrificed")
	}
}

func TestM14B1StonehornChanterGainsVigilanceAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chanter := effectstest.B12Push(g, me.ID, "Stonehorn Chanter", "Creature — Rhino Cleric", m14B1StonehornChanterOracle, 4, 4)
	effectstest.ApaMana(me, "C", "C", "C", "C", "C", "W")
	effectstest.ApaActivate(t, g, me, chanter, 0, game.ActivateAbilityParams{})
	for _, kw := range []string{"vigilance", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, chanter, kw) {
			t.Errorf("the Chanter has no %s", kw)
		}
	}
}

func TestM14B1VialOfPoisonGrantsDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vial := effectstest.PushCatalogPermanent(g, me.ID, "Vial of Poison", "Artifact", m14B1VialOfPoisonOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m14B1BearType, 2, 2)
	effectstest.ApaMana(me, "C")
	effectstest.ApaActivate(t, g, me, vial, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("the Bear has no deathtouch")
	}
	if g.Battlefield.Contains(vial) {
		t.Error("the Vial was not sacrificed")
	}
}

func TestM14B1ZephyrChargeGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	charge := effectstest.PushCatalogPermanent(g, me.ID, "Zephyr Charge", "Enchantment", m14B1ZephyrChargeOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m14B1BearType, 2, 2)
	effectstest.ApaMana(me, "C", "U")
	effectstest.ApaActivate(t, g, me, charge, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear has no flying")
	}
}

func TestM14B1EncroachingWastesDestroysANonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wastes := effectstest.PushCatalogPermanent(g, me.ID, "Encroaching Wastes", "Land", m14B1EncroachingWastesOracle, false)
	basic := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	dual := effectstest.PushLandFor(g, opp.ID, "Their Dual", "Land")
	effectstest.ApaMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, wastes, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(basic)}); err == nil {
		t.Fatal("Encroaching Wastes targeted a basic land")
	}
	effectstest.ApaActivate(t, g, me, wastes, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(dual)})
	if g.Battlefield.Contains(dual) {
		t.Error("the nonbasic land survived")
	}
	if g.Battlefield.Contains(wastes) {
		t.Error("Encroaching Wastes was not sacrificed")
	}
}

func TestM14B1RingOfThreeWishesTutorsForAWishCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ring := effectstest.CastCatalogSpell(t, g, "Ring of Three Wishes", "Artifact", m14B1RingOfThreeWishesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ring).Counters["wish"]; got != 3 {
		t.Fatalf("%d wish counters, want 3", got)
	}
	needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Needle", TypeLine: "Sorcery"})
	effectstest.ApaMana(me, "C", "C", "C", "C", "C")
	effectstest.ApaActivate(t, g, me, ring, 0, game.ActivateAbilityParams{})
	effectstest.AnswerSearchByID(t, g, me.ID, needle)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(needle) {
		t.Error("the searched card is not in hand")
	}
	if got := effectstest.FindCardAnywhere(t, g, ring).Counters["wish"]; got != 2 {
		t.Errorf("%d wish counters after one wish, want 2", got)
	}
}

func TestM14B1BarrageOfExpendablesSacrificesToPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	barrage := effectstest.PushCatalogPermanent(g, me.ID, "Barrage of Expendables", "Enchantment", m14B1BarrageOracle, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", m14B1BearType, 1, 1)
	life := opp.Life
	effectstest.ApaMana(me, "R")
	effectstest.ApaActivate(t, g, me, barrage, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{effectstest.PlayerTarget(opp.ID)},
		SacrificeIDs: []uuid.UUID{fodder},
	})
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
}

// --- combat statics -------------------------------------------------------------

func TestM14B1CyclopsTyrantCantBlockSmallCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tyrant := effectstest.B12Push(g, opp.ID, "Cyclops Tyrant", "Creature — Cyclops", m14B1CyclopsTyrantOracle, 3, 4)
	small := effectstest.B12Creature(g, me.ID, "Small", m14B1BearType, 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", m14B1BearType, 3, 3)
	effectstest.BrAttack(t, g, small, big)
	if err := g.DeclareBlocker(tyrant, small); err == nil {
		t.Error("the Tyrant blocked a creature with power 2")
	}
	if err := g.DeclareBlocker(tyrant, big); err != nil {
		t.Errorf("the Tyrant can't block a creature with power 3: %v", err)
	}
}

func TestM14B1MaraudingMaulhornMustAttackWithoutTheAdvocate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	maulhorn := effectstest.B12Push(g, me.ID, "Marauding Maulhorn", "Creature — Beast", m14B1MaraudingMaulhornOracle, 5, 3)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != maulhorn {
		t.Fatalf("pass with the Maulhorn home = %v, want a refusal naming it", err)
	}
}

func TestM14B1MaraudingMaulhornStaysHomeWithTheAdvocate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.B12Push(g, me.ID, "Marauding Maulhorn", "Creature — Beast", m14B1MaraudingMaulhornOracle, 5, 3)
	effectstest.B12Creature(g, me.ID, "Advocate of the Beast", "Creature — Elf Shaman", 2, 3)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	if err := g.PassPriority(); err != nil {
		t.Errorf("pass with Advocate of the Beast in play: %v", err)
	}
}

// --- the spells -------------------------------------------------------------------

func TestM14B1LiturgyOfBloodDestroysAndAddsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m14B1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Liturgy of Blood", "Sorcery", m14B1LiturgyOfBloodOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("%d mana in the pool, want 3", got)
	}
}

func TestM14B1HiveStirringsMakesTwoSlivers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.CastCatalogSpell(t, g, "Hive Stirrings", "Sorcery", m14B1HiveStirringsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Sliver", me.ID); n != 2 {
		t.Errorf("%d Sliver tokens, want 2", n)
	}
}

func TestM14B1HuntTheWeakCountersThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", m14B1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m14B1BearType, 1, 3)
	effectstest.CastCatalogSpell(t, g, "Hunt the Weak", "Sorcery", m14B1HuntTheWeakOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the fought creature survived 3 damage")
	}
	if p, tough := effectstest.PtOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
}

func TestM14B1GlimpseTheFutureTakesOneOfThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	top := effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.CastCatalogSpell(t, g, "Glimpse the Future", "Sorcery", m14B1GlimpseTheFutureOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top three", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top[1]) {
		t.Error("the chosen card is not in hand")
	}
	for _, id := range []uuid.UUID{top[0], top[2]} {
		if !me.Graveyard.Contains(id) {
			t.Error("a card not chosen is not in the graveyard")
		}
	}
}
