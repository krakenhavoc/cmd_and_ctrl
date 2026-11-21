package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aer_5536_b1_cards_test.go — Aether Revolt, batch 1 (#5536). Every card
// is registered Full; every card with rules text is played through the
// engine.

const (
	aerBastionEnforcerOracle       = "87449f64-9500-4967-a073-cde19824e087"
	aerLathnuSailbackOracle        = "9c30147a-1131-402e-bdaf-9006b1022b83"
	aerPrizefighterConstructOracle = "6f4cfd69-c739-4398-93a5-e6238b7cd83b"
	aerAlleyStranglerOracle        = "fb9af562-7384-4575-93bb-34f1ada23735"
	aerGhirapurOspreyOracle        = "b4de9edd-4a5b-4858-8454-38381df8b9a2"
	aerGiftedAetherbornOracle      = "0c6a6587-2b3f-40d1-a5f2-06128da0d52e"
	aerAudaciousInfiltratorOracle  = "e01b3d52-297d-42e3-9de9-4f45c2359bcc"
	aerOutlandBoarOracle           = "fdb109f6-d185-4d1e-92b9-768e77562f55"
	aerNightMarketGuardOracle      = "3a7a27fe-60f0-4a21-a1c1-5fef8a528391"
	aerFrontlineRebelOracle        = "7471e445-4655-4336-9769-3debd197b77c"
	aerHinterlandDrakeOracle       = "f457f182-69d0-468d-9bf0-7170363bac26"
	aerConsulateDreadnoughtOracle  = "bc822c9a-633f-4830-9cb3-9af6a80622e4"
	aerIrontreadCrusherOracle      = "c951fca7-f333-4ca3-b5d6-83967d1d093d"
	aerUntetheredExpressOracle     = "652f049d-770e-4bd7-b65f-fe56e618366a"
	aerNightMarketAeronautOracle   = "e08157e6-a5a3-4b1e-80e0-32702288d8ed"
	aerGreenwheelLiberatorOracle   = "121a8717-4c5b-45a8-875a-2b802c1fe6f8"
	aerLifecraftCavalryOracle      = "13db0230-b57e-47fd-9025-580d06cc88ee"
	aerHiddenHerbalistsOracle      = "3289539e-51eb-4915-bd77-3ac362136f2a"
	aerDeadeyeHarpoonerOracle      = "959e1f72-4ba2-4c87-95d1-ca0265dc128c"
	aerVengefulRebelOracle         = "273bf1a8-42a9-4db5-bd83-06db4cb5e5df"
	aerSolemnRecruitOracle         = "cf747d6e-05ae-484d-9e72-1d0fb71fd4ff"
	aerDecommissionOracle          = "d001321d-b272-4f94-a333-7cfe8482da20"
	aerAugmentingAutomatonOracle   = "5ec0e58c-99b5-4ebf-8053-f59c30b86ef7"
	aerVerdantAutomatonOracle      = "d3af1c69-42ae-4d29-a9f1-72f862d26303"
	aerWatchfulAutomatonOracle     = "6f5e6799-cb0f-4403-9c2a-69abeac08b79"
	aerWelderAutomatonOracle       = "52eae443-fbe5-479f-a5a4-b5e4401d62d3"
	aerEmbraalGearSmasherOracle    = "01e81249-6c45-4fae-86ee-9ef5fbec99ec"
	aerRavenousIntruderOracle      = "06a94ea3-6cda-42fe-ba3b-40a0e9044f8e"
	aerAegisAutomatonOracle        = "727f4734-53cc-4279-ba51-dfa26c4ceb1d"
	aerImplCombustionOracle        = "f4bd26d5-4448-451f-9909-b6dea1551af7"
	aerImplExaminationOracle       = "0baf75b3-cc02-47cc-873a-8d5f17da0ec9"
	aerImplFerocityOracle          = "b92629c0-1619-4214-929f-4404992fabeb"
	aerImplImprovementOracle       = "a71b2ffb-6a62-4239-987f-ec82a358927f"
	aerImplMaliceOracle            = "16923d37-bac0-4e6a-bd02-f3aa9360deeb"
	aerPacificationArrayOracle     = "b2c14195-2747-490e-99ef-5b71048ff51a"
	aerPendulumOfPatternsOracle    = "992075d5-f413-4896-b75e-7bb2d589c50e"
	aerRenegadeMapOracle           = "f8506acb-3fc5-4fdf-8304-586843db6cfd"
	aerFiligreeCrawlerOracle       = "f74095aa-232f-432e-a899-86e9463ab5be"
	aerRidgescaleTuskerOracle      = "4eef2f24-6a5c-4bc7-b316-dd818dc2741a"
	aerDawnfeatherEagleOracle      = "e17db9d8-14b0-4f72-aacc-f51fae4597e2"
	aerSpirePatrolOracle           = "afacd2b1-1733-4b28-b834-71a31149fa00"
	aerThopterArrestOracle         = "9b9cc771-3988-41af-a49e-f3bb7a8e397c"
	aerRecklessRacerOracle         = "d9a09e7f-6bef-4e75-b5c7-cea20a69cf02"
	aerPreciseStrikeOracle         = "bfd2138b-1883-4900-a5f1-373884965c34"
	aerTakeIntoCustodyOracle       = "5e5f92b3-204b-42a3-b86f-8e2a1d263d1d"
	aerDaringDemolitionOracle      = "8c8a3962-42a5-4070-8719-00d7483c7dbe"
	aerInspiringRoarOracle         = "96cf2810-35b8-437e-b609-f9765604b8cf"
	aerHungryFlamesOracle          = "a2c00788-14e7-4ab5-8c71-2a176621a177"
	aerIceOverOracle               = "80f8a64e-b22e-413e-a143-f1943be3a574"
	aerWrangleOracle               = "6c7d8b6b-9890-4a94-85fc-f7545c4b1cf7"

	aerBearType      = "Creature — Bear"
	aerConstructType = "Artifact Creature — Construct"
	aerVehicleType   = "Artifact — Vehicle"
	aerArtifactType  = "Artifact"
)

// aerCastCreature puts a creature card with printed power and toughness
// into the active player's hand and casts it in a main phase, leaving it
// on the stack. The caller settles it (assistCardSettle).
func aerCastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// aerCrew activates a Vehicle's crew row (index 0) with the given crew.
func aerCrew(g *game.Game, me *game.Player, vehicle uuid.UUID, crew ...uuid.UUID) error {
	return g.ActivateCatalogAbility(me.ID, vehicle, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: crew})
}

// --- every card is in the catalog ----------------------------------------

func TestAer5536B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		aerBastionEnforcerOracle, aerLathnuSailbackOracle, aerPrizefighterConstructOracle, aerAlleyStranglerOracle,
		aerGhirapurOspreyOracle, aerGiftedAetherbornOracle, aerAudaciousInfiltratorOracle, aerOutlandBoarOracle,
		aerNightMarketGuardOracle, aerFrontlineRebelOracle, aerHinterlandDrakeOracle, aerConsulateDreadnoughtOracle,
		aerIrontreadCrusherOracle, aerUntetheredExpressOracle, aerNightMarketAeronautOracle, aerGreenwheelLiberatorOracle,
		aerLifecraftCavalryOracle, aerHiddenHerbalistsOracle, aerDeadeyeHarpoonerOracle, aerVengefulRebelOracle,
		aerSolemnRecruitOracle, aerDecommissionOracle, aerAugmentingAutomatonOracle, aerVerdantAutomatonOracle,
		aerWatchfulAutomatonOracle, aerWelderAutomatonOracle, aerEmbraalGearSmasherOracle, aerRavenousIntruderOracle,
		aerAegisAutomatonOracle, aerImplCombustionOracle, aerImplExaminationOracle, aerImplFerocityOracle,
		aerImplImprovementOracle, aerImplMaliceOracle, aerPacificationArrayOracle, aerPendulumOfPatternsOracle,
		aerRenegadeMapOracle, aerFiligreeCrawlerOracle, aerRidgescaleTuskerOracle, aerDawnfeatherEagleOracle,
		aerSpirePatrolOracle, aerThopterArrestOracle, aerRecklessRacerOracle, aerPreciseStrikeOracle,
		aerTakeIntoCustodyOracle, aerDaringDemolitionOracle, aerInspiringRoarOracle, aerHungryFlamesOracle,
		aerIceOverOracle, aerWrangleOracle,
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

// --- printed keywords ------------------------------------------------------

func TestAer5536B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Alley Strangler", "Creature — Aetherborn Rogue", aerAlleyStranglerOracle, []string{"menace"}},
		{"Ghirapur Osprey", "Creature — Bird", aerGhirapurOspreyOracle, []string{"flying"}},
		{"Gifted Aetherborn", "Creature — Aetherborn Vampire", aerGiftedAetherbornOracle, []string{"deathtouch", "lifelink"}},
		{"Hinterland Drake", "Creature — Drake", aerHinterlandDrakeOracle, []string{"flying"}},
		{"Lifecraft Cavalry", "Creature — Elf Warrior", aerLifecraftCavalryOracle, []string{"trample"}},
		{"Night Market Aeronaut", "Creature — Aetherborn Warrior", aerNightMarketAeronautOracle, []string{"flying"}},
		{"Solemn Recruit", "Creature — Dwarf Warrior", aerSolemnRecruitOracle, []string{"double strike"}},
		{"Reckless Racer", "Creature — Human Pilot", aerRecklessRacerOracle, []string{"first strike"}},
		{"Dawnfeather Eagle", "Creature — Bird", aerDawnfeatherEagleOracle, []string{"flying"}},
		{"Spire Patrol", "Creature — Human Soldier", aerSpirePatrolOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.AssertKeywords(t, g, id, tc.keywords...)
		})
	}
}

// --- combat statics ----------------------------------------------------------

func TestAer5536B1AudaciousInfiltratorEvadesArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rogue := effectstest.B12Push(g, me.ID, "Audacious Infiltrator", "Creature — Dwarf Rogue", aerAudaciousInfiltratorOracle, 3, 1)
	construct := effectstest.B12Creature(g, opp.ID, "Construct", aerConstructType, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", aerBearType, 2, 2)
	effectstest.BrAttack(t, g, rogue)
	if err := g.DeclareBlocker(construct, rogue); err == nil {
		t.Error("an artifact creature blocked the Infiltrator")
	}
	if err := g.DeclareBlocker(bear, rogue); err != nil {
		t.Errorf("a nonartifact creature can't block the Infiltrator: %v", err)
	}
}

func TestAer5536B1OutlandBoarEvadesSmallCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boar := effectstest.B12Push(g, me.ID, "Outland Boar", "Creature — Boar", aerOutlandBoarOracle, 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", aerBearType, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", aerBearType, 3, 3)
	effectstest.BrAttack(t, g, boar)
	if err := g.DeclareBlocker(small, boar); err == nil {
		t.Error("a creature with power 2 blocked the Boar")
	}
	if err := g.DeclareBlocker(big, boar); err != nil {
		t.Errorf("a creature with power 3 can't block the Boar: %v", err)
	}
}

func TestAer5536B1HinterlandDrakeCantBlockArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	construct := effectstest.B12Creature(g, me.ID, "Construct", aerConstructType, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", aerBearType, 2, 2)
	drake := effectstest.B12Push(g, opp.ID, "Hinterland Drake", "Creature — Drake", aerHinterlandDrakeOracle, 2, 3)
	effectstest.BrAttack(t, g, construct, bear)
	if err := g.DeclareBlocker(drake, construct); err == nil {
		t.Error("the Drake blocked an artifact creature")
	}
	if err := g.DeclareBlocker(drake, bear); err != nil {
		t.Errorf("the Drake can't block a nonartifact creature: %v", err)
	}
}

func TestAer5536B1NightMarketGuardBlocksTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	guard := effectstest.B12Push(g, g.Seats[1].ID, "Night Market Guard", aerConstructType, aerNightMarketGuardOracle, 3, 1)
	if got := effectstest.MbCapacity(t, g, guard); got != 2 {
		t.Errorf("block capacity %d, want 2", got)
	}
}

func TestAer5536B1FrontlineRebelMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	rebel := effectstest.B12Push(g, me.ID, "Frontline Rebel", "Creature — Human Warrior", aerFrontlineRebelOracle, 3, 3)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != rebel {
		t.Fatalf("pass with the Rebel home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(rebel, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Rebel attacking: %v", err)
	}
}

// --- Vehicles -------------------------------------------------------------------

func TestAer5536B1VehiclesCrewAtTheirNumber(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		crew         int
	}{
		{"Consulate Dreadnought", aerConsulateDreadnoughtOracle, 6},
		{"Irontread Crusher", aerIrontreadCrusherOracle, 3},
		{"Untethered Express", aerUntetheredExpressOracle, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			vehicle := effectstest.B12Push(g, me.ID, tc.name, aerVehicleType, tc.oracle, 4, 4)
			short := effectstest.B12Creature(g, me.ID, "Short", aerBearType, tc.crew-1, 1)
			enough := effectstest.B12Creature(g, me.ID, "Enough", aerBearType, tc.crew, 1)
			effectstest.AdvanceToMain(t, g)
			if effectstest.IsCreatureNow(g, vehicle) {
				t.Fatal("the uncrewed Vehicle is a creature")
			}
			if err := aerCrew(g, me, vehicle, short); err == nil {
				t.Fatalf("Crew %d was paid with power %d", tc.crew, tc.crew-1)
			}
			if err := aerCrew(g, me, vehicle, enough); err != nil {
				t.Fatalf("Crew %d with power %d: %v", tc.crew, tc.crew, err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.IsCreatureNow(g, vehicle) {
				t.Error("the crewed Vehicle is not a creature")
			}
		})
	}
}

func TestAer5536B1UntetheredExpressGrowsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	express := effectstest.B12Push(g, me.ID, "Untethered Express", aerVehicleType, aerUntetheredExpressOracle, 4, 4)
	pilot := effectstest.B12Creature(g, me.ID, "Pilot", aerBearType, 1, 1)
	effectstest.AssertKeywords(t, g, express, "trample")
	effectstest.AdvanceToMain(t, g)
	if err := aerCrew(g, me, express, pilot); err != nil {
		t.Fatalf("Crew 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(express, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldPlusOnes(g, express); n != 1 {
		t.Errorf("%d +1/+1 counters after attacking, want 1", n)
	}
}

// --- revolt --------------------------------------------------------------------

func TestAer5536B1RevoltEntersWithCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Night Market Aeronaut", "Creature — Aetherborn Warrior", aerNightMarketAeronautOracle, 1},
		{"Greenwheel Liberator", "Creature — Elf Warrior", aerGreenwheelLiberatorOracle, 2},
		{"Lifecraft Cavalry", "Creature — Elf Warrior", aerLifecraftCavalryOracle, 2},
	} {
		for _, revolt := range []bool{false, true} {
			name := tc.name + " without revolt"
			if revolt {
				name = tc.name + " with revolt"
			}
			t.Run(name, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				if revolt {
					effectstest.RevoltFor(t, g, me.ID)
				}
				id := aerCastCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
				effectstest.AssistCardSettle(t, g, me.ID, nil, true)
				want := 0
				if revolt {
					want = tc.n
				}
				if got := effectstest.BattlefieldPlusOnes(g, id); got != want {
					t.Errorf("%d +1/+1 counters, want %d", got, want)
				}
			})
		}
	}
}

func TestAer5536B1HiddenHerbalistsAddsGreenWithRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aerCastCreature(t, g, "Hidden Herbalists", "Creature — Human Druid", aerHiddenHerbalistsOracle, 2, 2)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if n := me.ManaPool.Total(); n != 0 {
		t.Fatalf("no revolt: %d mana in the pool", n)
	}
	effectstest.RevoltFor(t, g, me.ID)
	aerCastCreature(t, g, "Hidden Herbalists", "Creature — Human Druid", aerHiddenHerbalistsOracle, 2, 2)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("revolt: %d mana in the pool, want 2", n)
	}
}

func TestAer5536B1DeadeyeHarpoonerDestroysATappedCreatureWithRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	effectstest.AssistCardTapped(t, g, victim)
	aerCastCreature(t, g, "Deadeye Harpooner", "Creature — Dwarf Warrior", aerDeadeyeHarpoonerOracle, 2, 2)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("no revolt: the tapped creature was destroyed")
	}
	effectstest.RevoltFor(t, g, me.ID)
	aerCastCreature(t, g, "Deadeye Harpooner", "Creature — Dwarf Warrior", aerDeadeyeHarpoonerOracle, 2, 2)
	pick := game.TargetRef{Kind: game.TargetCard, ID: victim}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
	if g.Battlefield.Contains(victim) {
		t.Error("revolt: the tapped creature survived")
	}
}

func TestAer5536B1VengefulRebelShrinksWithRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	ogre := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 4, 4)
	aerCastCreature(t, g, "Vengeful Rebel", "Creature — Aetherborn Warrior", aerVengefulRebelOracle, 3, 2)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if p, tough := effectstest.PtOf(t, g, ogre); p != 4 || tough != 4 {
		t.Fatalf("no revolt: the Ogre is %d/%d", p, tough)
	}
	effectstest.RevoltFor(t, g, me.ID)
	aerCastCreature(t, g, "Vengeful Rebel", "Creature — Aetherborn Warrior", aerVengefulRebelOracle, 3, 2)
	pick := game.TargetRef{Kind: game.TargetCard, ID: ogre}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
	if p, tough := effectstest.PtOf(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("revolt: the Ogre is %d/%d, want 1/1", p, tough)
	}
}

func TestAer5536B1SolemnRecruitGrowsAtEndStepWithRevolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	recruit := effectstest.B12Push(g, me.ID, "Solemn Recruit", "Creature — Dwarf Warrior", aerSolemnRecruitOracle, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if effectstest.TriggerOnStackFrom(g, "Solemn Recruit") {
		t.Fatal("triggered with no revolt")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	recruit2 := effectstest.B12Push(g2, me2.ID, "Solemn Recruit", "Creature — Dwarf Warrior", aerSolemnRecruitOracle, 2, 2)
	effectstest.RevoltFor(t, g2, me2.ID)
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	if n := effectstest.BattlefieldPlusOnes(g2, recruit2); n != 1 {
		t.Errorf("revolt: %d counters, want 1", n)
	}
	_ = recruit
}

func TestAer5536B1DecommissionGainsLifeOnlyWithRevolt(t *testing.T) {
	for _, revolt := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
		rock := effectstest.B12Permanent(g, opp.ID, "Rock", aerArtifactType)
		if revolt {
			effectstest.RevoltFor(t, g, me.ID)
		}
		life := me.Life
		effectstest.CastCatalogSpell(t, g, "Decommission", "Instant", aerDecommissionOracle, effectstest.CardTarget(rock))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) {
			t.Errorf("revolt %v: the artifact survived", revolt)
		}
		want := life
		if revolt {
			want = life + 3
		}
		if me.Life != want {
			t.Errorf("revolt %v: life %d, want %d", revolt, me.Life, want)
		}
	}
}

// --- activated abilities ---------------------------------------------------------

func TestAer5536B1AugmentingAutomatonPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	id := effectstest.B12Push(g, me.ID, "Augmenting Automaton", aerConstructType, aerAugmentingAutomatonOracle, 1, 1)
	effectstest.ApaMana(me, "B", "B")
	effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, id); p != 2 || tough != 2 {
		t.Errorf("the Automaton is %d/%d, want 2/2", p, tough)
	}
}

func TestAer5536B1VerdantAutomatonPutsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	id := effectstest.B12Push(g, me.ID, "Verdant Automaton", aerConstructType, aerVerdantAutomatonOracle, 1, 2)
	effectstest.ApaMana(me, "G", "G", "G", "G")
	effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if n := effectstest.BattlefieldPlusOnes(g, id); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
}

func TestAer5536B1WatchfulAutomatonScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	id := effectstest.B12Push(g, me.ID, "Watchful Automaton", aerConstructType, aerWatchfulAutomatonOracle, 2, 2)
	effectstest.ApaMana(me, "U", "U", "U")
	effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil || len(scry.ScryCards) != 1 {
		t.Fatalf("expected a scry 1 prompt, got %+v", scry)
	}
}

func TestAer5536B1WelderAutomatonPingsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	id := effectstest.B12Push(g, me.ID, "Welder Automaton", aerConstructType, aerWelderAutomatonOracle, 2, 1)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.ApaMana(me, "R", "R", "R", "R")
	effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	for _, s := range g.Seats {
		want := lives[s.ID] - 1
		if s.ID == me.ID {
			want = lives[s.ID]
		}
		if s.Life != want {
			t.Errorf("%s at %d, want %d", s.Name, s.Life, want)
		}
	}
}

func TestAer5536B1EmbraalGearSmasherSacrificesAnArtifactToBurnEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	smasher := effectstest.B12Push(g, me.ID, "Embraal Gear-Smasher", "Creature — Human Warrior", aerEmbraalGearSmasherOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", aerBearType, 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", aerArtifactType)
	if err := g.ActivateCatalogAbility(me.ID, smasher, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a nonartifact creature paid the sacrifice")
	}
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.ApaActivate(t, g, me, smasher, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if !effectstest.HmlTapped(t, g, smasher) {
		t.Error("the Gear-Smasher did not tap")
	}
	for _, s := range g.Seats {
		want := lives[s.ID] - 2
		if s.ID == me.ID {
			want = lives[s.ID]
		}
		if s.Life != want {
			t.Errorf("%s at %d, want %d", s.Name, s.Life, want)
		}
	}
}

func TestAer5536B1RavenousIntruderEatsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	gremlin := effectstest.B12Push(g, me.ID, "Ravenous Intruder", "Creature — Gremlin", aerRavenousIntruderOracle, 1, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", aerArtifactType)
	effectstest.ApaActivate(t, g, me, gremlin, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if p, tough := effectstest.PtOf(t, g, gremlin); p != 3 || tough != 4 {
		t.Errorf("the Intruder is %d/%d, want 3/4", p, tough)
	}
}

func TestAer5536B1AegisAutomatonReturnsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	aegis := effectstest.B12Push(g, me.ID, "Aegis Automaton", aerConstructType, aerAegisAutomatonOracle, 0, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", aerBearType, 2, 2)
	effectstest.ApaMana(me, "W", "W", "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, aegis, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(aegis)}); err == nil {
		t.Fatal("the Automaton targeted itself")
	}
	effectstest.ApaActivate(t, g, me, aegis, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
}

// --- Implements ---------------------------------------------------------------------

func TestAer5536B1ImplementsSacrificeAndDraw(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, mana string
		draws              int
		target             func(g *game.Game, me, opp *game.Player) []game.TargetRef
		check              func(t *testing.T, g *game.Game, me, opp *game.Player, life, oppLife int, target []game.TargetRef)
	}{
		{"Implement of Combustion", aerImplCombustionOracle, "R", 1,
			func(_ *game.Game, _, opp *game.Player) []game.TargetRef {
				return []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
			},
			func(t *testing.T, _ *game.Game, _, opp *game.Player, _, oppLife int, _ []game.TargetRef) {
				if opp.Life != oppLife-1 {
					t.Errorf("the opponent is at %d, want %d", opp.Life, oppLife-1)
				}
			}},
		{"Implement of Examination", aerImplExaminationOracle, "U", 2, nil, nil},
		{"Implement of Ferocity", aerImplFerocityOracle, "G", 1,
			func(g *game.Game, me, _ *game.Player) []game.TargetRef {
				return effectstest.CardTarget(effectstest.B12Creature(g, me.ID, "Bear", aerBearType, 2, 2))
			},
			func(t *testing.T, g *game.Game, _, _ *game.Player, _, _ int, target []game.TargetRef) {
				if n := effectstest.BattlefieldPlusOnes(g, target[0].ID); n != 1 {
					t.Errorf("%d +1/+1 counters on the target, want 1", n)
				}
			}},
		{"Implement of Improvement", aerImplImprovementOracle, "W", 1, nil,
			func(t *testing.T, _ *game.Game, me, _ *game.Player, life, _ int, _ []game.TargetRef) {
				if me.Life != life+2 {
					t.Errorf("life %d, want %d", me.Life, life+2)
				}
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
			impl := effectstest.PushCatalogPermanent(g, me.ID, tc.name, aerArtifactType, tc.oracle, false)
			var targets []game.TargetRef
			if tc.target != nil {
				targets = tc.target(g, me, opp)
			}
			life, oppLife, hand := me.Life, opp.Life, me.Hand.Size()
			effectstest.ApaMana(me, tc.mana)
			effectstest.ApaActivate(t, g, me, impl, 0, game.ActivateAbilityParams{Targets: targets})
			if g.Battlefield.Contains(impl) || !me.Graveyard.Contains(impl) {
				t.Error("the Implement was not sacrificed")
			}
			if got := me.Hand.Size() - hand; got != tc.draws {
				t.Errorf("drew %d, want %d", got, tc.draws)
			}
			if tc.check != nil {
				tc.check(t, g, me, opp, life, oppLife, targets)
			}
		})
	}
}

func TestAer5536B1ImplementOfMaliceMakesAPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.D2dSetHand(opp, 3)
	impl := effectstest.PushCatalogPermanent(g, me.ID, "Implement of Malice", aerArtifactType, aerImplMaliceOracle, false)
	hand := me.Hand.Size()
	effectstest.ApaMana(me, "B")
	effectstest.ApaActivate(t, g, me, impl, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if !me.Graveyard.Contains(impl) {
		t.Error("the Implement was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Errorf("the target player owes %d discards, want 1", n)
	}
}

// A destroyed Implement draws too: the trigger is on the card, not the
// ability.
func TestAer5536B1ImplementDrawsWhenDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	impl := effectstest.PushCatalogPermanent(g, me.ID, "Implement of Improvement", aerArtifactType, aerImplImprovementOracle, false)
	hand := me.Hand.Size()
	effectstest.B27Kill(g, impl)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- other artifacts ------------------------------------------------------------------

func TestAer5536B1PacificationArrayTapsAnArtifactOrCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	array := effectstest.PushCatalogPermanent(g, me.ID, "Pacification Array", aerArtifactType, aerPacificationArrayOracle, false)
	land := effectstest.B12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", aerBearType, 2, 2)
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, array, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(land)}); err == nil {
		t.Fatal("the Array targeted a land")
	}
	effectstest.ApaActivate(t, g, me, array, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !effectstest.HmlTapped(t, g, bear) {
		t.Error("the creature is untapped")
	}
}

func TestAer5536B1PendulumOfPatternsGainsThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	life := me.Life
	pendulum := effectstest.CastCatalogSpell(t, g, "Pendulum of Patterns", aerArtifactType, aerPendulumOfPatternsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Fatalf("life %d after it entered, want %d", me.Life, life+3)
	}
	hand := me.Hand.Size()
	effectstest.ApaMana(me, "C", "C", "C", "C", "C")
	effectstest.ApaActivate(t, g, me, pendulum, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(pendulum) {
		t.Error("the Pendulum was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestAer5536B1RenegadeMapEntersTappedAndSacrificesToSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	mapID := effectstest.CastCatalogSpell(t, g, "Renegade Map", aerArtifactType, aerRenegadeMapOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, ok := effectstest.BattlefieldCardTapped(g, mapID); !ok || !tapped {
		t.Fatal("the Map did not enter tapped")
	}
	effectstest.FindBattlefieldCardForTest(g, mapID).Tapped = false
	effectstest.ApaActivate(t, g, me, mapID, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(mapID) || !me.Graveyard.Contains(mapID) {
		t.Error("the Map was not sacrificed")
	}
}

// --- triggered abilities -------------------------------------------------------------

func TestAer5536B1FiligreeCrawlerLeavesAThopter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	crawler := effectstest.B12Push(g, g.Seats[0].ID, "Filigree Crawler", "Artifact Creature — Insect", aerFiligreeCrawlerOracle, 2, 2)
	effectstest.B27Kill(g, crawler)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Thopter"); n != 1 {
		t.Errorf("%d Thopters, want 1", n)
	}
}

func TestAer5536B1RidgescaleTuskerCountersEachOtherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", aerBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", aerBearType, 2, 2)
	tusker := aerCastCreature(t, g, "Ridgescale Tusker", "Creature — Pangolin Beast", aerRidgescaleTuskerOracle, 5, 5)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if n := effectstest.BattlefieldPlusOnes(g, mine); n != 1 {
		t.Errorf("your other creature has %d counters, want 1", n)
	}
	if n := effectstest.BattlefieldPlusOnes(g, tusker); n != 0 {
		t.Errorf("the Tusker has %d counters, want 0", n)
	}
	if n := effectstest.BattlefieldPlusOnes(g, theirs); n != 0 {
		t.Errorf("an opponent's creature has %d counters, want 0", n)
	}
}

func TestAer5536B1DawnfeatherEagleRalliesYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", aerBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", aerBearType, 2, 2)
	aerCastCreature(t, g, "Dawnfeather Eagle", "Creature — Bird", aerDawnfeatherEagleOracle, 3, 3)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if p, tough := effectstest.PtOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "vigilance") {
		t.Error("your creature has no vigilance")
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestAer5536B1SpirePatrolTapsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", aerBearType, 2, 2)
	aerCastCreature(t, g, "Spire Patrol", "Creature — Human Soldier", aerSpirePatrolOracle, 3, 2)
	pick := game.TargetRef{Kind: game.TargetCard, ID: bear}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
	if !effectstest.HmlTapped(t, g, bear) {
		t.Error("the opponent's creature is untapped")
	}
}

func TestAer5536B1ThopterArrestExilesAnOpponentsArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", aerArtifactType)
	effectstest.CastCatalogSpell(t, g, "Thopter Arrest", "Enchantment", aerThopterArrestOracle, nil)
	pick := game.TargetRef{Kind: game.TargetCard, ID: rock}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
	if g.Battlefield.Contains(rock) {
		t.Error("the opponent's artifact is still on the battlefield")
	}
}

func TestAer5536B1RecklessRacerLootsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B")
	fodder := effectstest.HandCard(me, "Fodder", "Instant")
	racer := effectstest.B12Push(g, me.ID, "Reckless Racer", "Creature — Human Pilot", aerRecklessRacerOracle, 2, 3)
	effectstest.DeclareAttack(t, g, opp.ID, racer)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt when the Racer became tapped")
	}
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, fodder)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the discarded card is not in the graveyard")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want one discarded and one drawn", hand, me.Hand.Size())
	}
}

// --- spells ------------------------------------------------------------------------------

func TestAer5536B1PreciseStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", aerBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Precise Strike", "Instant", aerPreciseStrikeOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the Bear has no first strike")
	}
}

func TestAer5536B1TakeIntoCustodyTapsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", aerBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Take into Custody", "Instant", aerTakeIntoCustodyOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, bear) {
		t.Error("the creature is untapped")
	}
}

func TestAer5536B1DaringDemolitionDestroysAVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", aerArtifactType)
	if err := effectstest.CastCatalogSpellErr(t, g, "Daring Demolition", "Sorcery", aerDaringDemolitionOracle, effectstest.CardTarget(rock)); err == nil {
		t.Fatal("Daring Demolition targeted a noncreature, non-Vehicle artifact")
	}
	vehicle := effectstest.B12Push(g, opp.ID, "Cart", aerVehicleType, "", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Daring Demolition", "Sorcery", aerDaringDemolitionOracle, effectstest.CardTarget(vehicle))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vehicle) {
		t.Error("the uncrewed Vehicle survived")
	}
}

func TestAer5536B1InspiringRoarCountersYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	a := effectstest.B12Creature(g, me.ID, "A", aerBearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", aerBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", aerBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Inspiring Roar", "Sorcery", aerInspiringRoarOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := effectstest.BattlefieldPlusOnes(g, id); n != 1 {
			t.Errorf("your creature has %d counters, want 1", n)
		}
	}
	if n := effectstest.BattlefieldPlusOnes(g, theirs); n != 0 {
		t.Errorf("an opponent's creature has %d counters, want 0", n)
	}
}

func TestAer5536B1HungryFlamesSplitsItsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", aerBearType, 4, 4)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Hungry Flames", "Instant", aerHungryFlamesOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ogre}, effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 3 {
		t.Errorf("the creature has %d damage, want 3", got)
	}
	if opp.Life != life-2 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-2)
	}
}

func TestAer5536B1IceOverKeepsAnArtifactTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := g.Seats[1]
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", aerArtifactType, "", false)
	effectstest.Bns44Aura(t, g, "Ice Over", aerIceOverOracle, rock)
	effectstest.FindBattlefieldCardForTest(g, rock).Tapped = true
	effectstest.EndTurn(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceTo(t, g, game.StepDraw)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, rock); !tapped {
		t.Error("the enchanted artifact untapped in its controller's untap step")
	}
}

func TestAer5536B1WrangleStealsASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	big := effectstest.B12Creature(g, opp.ID, "Big", aerBearType, 5, 5)
	if err := effectstest.CastCatalogSpellErr(t, g, "Wrangle", "Sorcery", aerWrangleOracle, effectstest.CardTarget(big)); err == nil {
		t.Fatal("Wrangle targeted a creature with power 5")
	}
	bear := effectstest.B12Creature(g, opp.ID, "Bear", aerBearType, 4, 4)
	effectstest.AssistCardTapped(t, g, bear)
	effectstest.CastCatalogSpell(t, g, "Wrangle", "Sorcery", aerWrangleOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindCardAnywhere(t, g, bear); c.Controller != me.ID {
		t.Error("you don't control the creature")
	}
	if effectstest.HmlTapped(t, g, bear) {
		t.Error("the creature was not untapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the creature has no haste")
	}
}
