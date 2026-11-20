package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m10_5585_b1_cards_test.go — Magic 2010, batch 1 (#5585). Every card is
// registered Full, and each one with rules text is played through the
// engine.

const (
	m10B1CentaurCourserOracle      = "2f5bf099-2e01-4e1c-9ebf-0ce0ac66939e"
	m10B1EliteVanguardOracle       = "b3b9a87d-cb95-435c-90b6-037406cab32e"
	m10B1RuneclawBearOracle        = "ec49dfcf-d16d-4621-af4b-4a6f09043221"
	m10B1SiegeMastodonOracle       = "b4c404d8-9f2d-4429-ac36-449ae319abb7"
	m10B1SilvercoatLionOracle      = "e3c0b11e-7340-4e0a-98b7-91b38439d4f9"
	m10B1WarpathGhoulOracle        = "17ae652c-b1c7-4dde-ae9e-50b9e286776c"
	m10B1ZombieGoliathOracle       = "eadd88b6-e75a-4482-8382-561718121772"
	m10B1ChildOfNightOracle        = "c650a7bc-e350-44a0-a698-d4a233d66156"
	m10B1DeadlyRecluseOracle       = "6b941291-1802-4ced-9869-39fa43825550"
	m10B1EmeraldOryxOracle         = "4c26c796-26ec-41f7-b9ca-58f4eb935c47"
	m10B1GriffinSentinelOracle     = "72df0b70-5c9f-44f3-96d5-81c95abc0065"
	m10B1KalonianBehemothOracle    = "68d0b5c0-df6a-4eea-8dea-20358270f916"
	m10B1KelinoreBatOracle         = "552dcf9f-ff4a-4057-a93e-62f5527050c9"
	m10B1MistLeopardOracle         = "49cade2f-92ab-48c8-aed3-d52ab3263506"
	m10B1StampedingRhinoOracle     = "8e4dc1fc-e412-4aa3-8793-5c4abc598521"
	m10B1StormfrontPegasusOracle   = "b35c1b30-c1dc-489f-bd79-ad6566fb9389"
	m10B1ViashinoSpearhunterOracle = "bf302113-7f2e-4ceb-a26f-1fd6e010be8a"
	m10B1ZephyrSpriteOracle        = "11f77158-0240-4882-94a4-6bd6b3d94bcc"
	m10B1BerserkersOracle          = "9e8a2d53-3145-421a-aad1-8f2b2f6eaec8"
	m10B1BlindingMageOracle        = "53a1faa1-ca67-4027-bee8-0f7c4584d62b"
	m10B1BorderlandRangerOracle    = "ffe25048-0bd4-49b5-aecb-2b1b57fb8d8b"
	m10B1CudgelTrollOracle         = "c2f86723-4765-4dac-a9f6-9cfa543613b2"
	m10B1DisentombOracle           = "e4ab986d-662c-4879-b6f1-6887a2ec4d7c"
	m10B1DivineVerdictOracle       = "3070f997-24f2-471d-af96-a2f3b588198e"
	m10B1FieryHellhoundOracle      = "ab71c8ba-c7ac-4ad2-a6db-29d903bbca78"
	m10B1GloriousChargeOracle      = "754d5009-7887-4293-b554-b3decddb33f7"
	m10B1JackalFamiliarOracle      = "e4d5d666-fd55-4c1a-89c8-2348288f6b30"
	m10B1SolemnOfferingOracle      = "aff51c27-fb61-41b3-a7ea-7092cd280c63"
	m10B1VampireAristocratOracle   = "c14739b0-5c19-4873-b027-314225a10f7b"
	m10B1AcolyteOfXathridOracle    = "62dca41b-58b0-45eb-ada7-568e80a2abec"
	m10B1AngelsMercyOracle         = "6b232bb7-d372-4174-a049-5f8d620810e6"
	m10B1BrambleCreeperOracle      = "33416979-2a3b-4184-818c-6a4e9ccb4580"
	m10B1CaptainOfTheWatchOracle   = "9f2af06d-8fb7-4276-bb92-3559a6d1fa18"
	m10B1DisorientOracle           = "33be346b-a232-4680-9b1b-a5eee57d2e75"
	m10B1DreadWarlockOracle        = "5791710d-c718-4f64-9c14-3668154716f1"
	m10B1GoblinArtilleryOracle     = "06db7444-0225-48c8-ab04-d69ff06e97b6"
	m10B1GreatSableStagOracle      = "d6dbb83d-3215-4446-8480-8f1bf8cf2938"
	m10B1HonorOfThePureOracle      = "eb4188bb-62df-4309-bde1-f66318fe2f05"
	m10B1IllusionaryServantOracle  = "28e99c5f-a78a-4323-b5e9-1b43303acaf6"
	m10B1LifelinkOracle            = "4fe8316a-9cff-43e9-a1d5-993a0c9daf3a"
	m10B1NaturesSpiralOracle       = "565922d6-8152-411f-8d4d-dc7c8523a49b"
	m10B1OakenformOracle           = "b254fefb-11d2-4369-ae5c-0c3bd48ae580"
	m10B1RegenerateOracle          = "4ce6c946-7a93-4a8d-a2c9-086cbd036e3c"
	m10B1RhoxPikemasterOracle      = "8866d33e-8159-4e10-8b11-914c12a0918f"
	m10B1SeismicStrikeOracle       = "3513d862-ef0d-4949-b33a-3fa9437cc444"
	m10B1TomeScourOracle           = "446a9d7e-9c1a-4e83-8342-2ed5acae2eed"
	m10B1UndeadSlayerOracle        = "08fc069e-7fdd-4726-aa08-a96bbd2bcaa3"
	m10B1VeteranArmorsmithOracle   = "fcb34e92-f991-401c-a494-c2919083bf9e"
	m10B1VeteranSwordsmithOracle   = "cc71b6c4-09d4-423b-a134-2f2f2fd7b533"
	m10B1WallOfFaithOracle         = "918d6cd4-76d5-4e21-b92b-1ecf35eb6cff"

	m10B1BearType    = "Creature — Bear"
	m10B1SoldierType = "Creature — Human Soldier"
	m10B1AuraType    = "Enchantment — Aura"
)

// m10B1GraveyardCard puts a card of the given type line into p's graveyard.
func m10B1GraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// m10B1CastNow casts a catalog spell from the active player's hand in the
// current step, without moving to a main phase.
func m10B1CastNow(g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) error {
	active := g.Seats[g.Turn.ActiveSeat]
	id := b13HandCard(active, name, typeLine, oracle)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets})
}

// --- every card is in the catalog --------------------------------------------

func TestM10B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		m10B1CentaurCourserOracle, m10B1EliteVanguardOracle, m10B1RuneclawBearOracle, m10B1SiegeMastodonOracle,
		m10B1SilvercoatLionOracle, m10B1WarpathGhoulOracle, m10B1ZombieGoliathOracle, m10B1ChildOfNightOracle,
		m10B1DeadlyRecluseOracle, m10B1EmeraldOryxOracle, m10B1GriffinSentinelOracle, m10B1KalonianBehemothOracle,
		m10B1KelinoreBatOracle, m10B1MistLeopardOracle, m10B1StampedingRhinoOracle, m10B1StormfrontPegasusOracle,
		m10B1ViashinoSpearhunterOracle, m10B1ZephyrSpriteOracle, m10B1BerserkersOracle, m10B1BlindingMageOracle,
		m10B1BorderlandRangerOracle, m10B1CudgelTrollOracle, m10B1DisentombOracle, m10B1DivineVerdictOracle,
		m10B1FieryHellhoundOracle, m10B1GloriousChargeOracle, m10B1JackalFamiliarOracle, m10B1SolemnOfferingOracle,
		m10B1VampireAristocratOracle, m10B1AcolyteOfXathridOracle, m10B1AngelsMercyOracle, m10B1BrambleCreeperOracle,
		m10B1CaptainOfTheWatchOracle, m10B1DisorientOracle, m10B1DreadWarlockOracle, m10B1GoblinArtilleryOracle,
		m10B1GreatSableStagOracle, m10B1HonorOfThePureOracle, m10B1IllusionaryServantOracle, m10B1LifelinkOracle,
		m10B1NaturesSpiralOracle, m10B1OakenformOracle, m10B1RegenerateOracle, m10B1RhoxPikemasterOracle,
		m10B1SeismicStrikeOracle, m10B1TomeScourOracle, m10B1UndeadSlayerOracle, m10B1VeteranArmorsmithOracle,
		m10B1VeteranSwordsmithOracle, m10B1WallOfFaithOracle,
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

// --- printed keywords -----------------------------------------------------------

func TestM10B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Child of Night", "Creature — Vampire", m10B1ChildOfNightOracle, []string{"lifelink"}},
		{"Deadly Recluse", "Creature — Spider", m10B1DeadlyRecluseOracle, []string{"reach", "deathtouch"}},
		{"Emerald Oryx", "Creature — Antelope", m10B1EmeraldOryxOracle, []string{"forestwalk"}},
		{"Griffin Sentinel", "Creature — Griffin", m10B1GriffinSentinelOracle, []string{"flying", "vigilance"}},
		{"Kalonian Behemoth", "Creature — Beast", m10B1KalonianBehemothOracle, []string{"shroud"}},
		{"Kelinore Bat", "Creature — Bat", m10B1KelinoreBatOracle, []string{"flying"}},
		{"Mist Leopard", "Creature — Cat", m10B1MistLeopardOracle, []string{"shroud"}},
		{"Stampeding Rhino", "Creature — Rhino", m10B1StampedingRhinoOracle, []string{"trample"}},
		{"Stormfront Pegasus", "Creature — Pegasus", m10B1StormfrontPegasusOracle, []string{"flying"}},
		{"Viashino Spearhunter", "Creature — Lizard Warrior", m10B1ViashinoSpearhunterOracle, []string{"first strike"}},
		{"Zephyr Sprite", "Creature — Faerie", m10B1ZephyrSpriteOracle, []string{"flying"}},
		{"Great Sable Stag", "Creature — Elk", m10B1GreatSableStagOracle, []string{game.ProtectionFromColor("U"), game.ProtectionFromColor("B")}},
		{"Illusionary Servant", "Creature — Illusion", m10B1IllusionaryServantOracle, []string{"flying"}},
		{"Rhox Pikemaster", "Creature — Rhino Soldier", m10B1RhoxPikemasterOracle, []string{"first strike"}},
		{"Captain of the Watch", m10B1SoldierType, m10B1CaptainOfTheWatchOracle, []string{"vigilance"}},
		{"Wall of Faith", "Creature — Wall", m10B1WallOfFaithOracle, []string{"defender"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

func TestM10B1GreatSableStagCantBeCountered(t *testing.T) {
	spec, ok := Lookup(m10B1GreatSableStagOracle)
	if !ok || !spec.CantBeCountered {
		t.Error("Great Sable Stag can be countered")
	}
}

// --- combat statics ---------------------------------------------------------------

func TestM10B1BerserkersMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	berserkers := b12Push(g, me.ID, "Berserkers of Blood Ridge", "Creature — Human Berserker", m10B1BerserkersOracle, 4, 4)
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != berserkers {
		t.Fatalf("pass with the Berserkers home = %v, want a refusal naming them", err)
	}
	if err := g.DeclareAttacker(berserkers, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Berserkers attacking: %v", err)
	}
}

func TestM10B1JackalFamiliarCantAttackAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jackal := apaPush(g, me.ID, me.ID, apaCreature("Jackal Familiar", m10B1JackalFamiliarOracle, 2, 2))
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(jackal, opp.ID); err != nil {
		t.Fatalf("staging the Jackal: %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("passing with the Jackal alone: %v, want ErrAttackLimit", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("the companion was refused: %v", err)
	}
	lockInAttacks(t, g)
	if c := findBattlefieldCardForTest(g, jackal); c == nil || c.AttackingTarget == uuid.Nil {
		t.Error("the Jackal is not attacking beside the Bear")
	}
}

func TestM10B1JackalFamiliarCantBlockAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a1 := pushVanillaCreature(g, me.ID, "Raider One", 2, 2)
	a2 := pushVanillaCreature(g, me.ID, "Raider Two", 2, 2)
	jackal := apaPush(g, opp.ID, opp.ID, apaCreature("Jackal Familiar", m10B1JackalFamiliarOracle, 2, 2))
	bear := apaPush(g, opp.ID, opp.ID, apaCreature("Bear", "", 2, 2))
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a1, Target: opp.ID}, {Attacker: a2, Target: opp.ID}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	err := g.DeclareBlocker(jackal, a1)
	var re *game.BlockRefusedError
	if !errors.As(err, &re) || re.Reason != game.BlockReasonCantBlockAlone {
		t.Fatalf("blocking alone: %v, want cant_block_alone", err)
	}
	if err := g.DeclareBlocker(bear, a1); err != nil {
		t.Fatalf("the Bear block: %v", err)
	}
	if err := g.DeclareBlocker(jackal, a2); err != nil {
		t.Fatalf("the Jackal beside the Bear: %v", err)
	}
}

func TestM10B1DreadWarlockBlockedOnlyByBlack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warlock := b12Push(g, me.ID, "Dread Warlock", "Creature — Human Wizard Warlock", m10B1DreadWarlockOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", m10B1BearType, 2, 2)
	hmlSetColors(g, bear, "G")
	zombie := b12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	hmlSetColors(g, zombie, "B")
	brAttack(t, g, warlock)
	if err := g.DeclareBlocker(bear, warlock); err == nil {
		t.Error("a green creature blocked Dread Warlock")
	}
	if err := g.DeclareBlocker(zombie, warlock); err != nil {
		t.Errorf("a black creature can't block Dread Warlock: %v", err)
	}
}

// --- lords and anthems ------------------------------------------------------------

func TestM10B1SoldierLords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		dp, dt                 int
	}{
		{"Veteran Armorsmith", m10B1SoldierType, m10B1VeteranArmorsmithOracle, 0, 1},
		{"Veteran Swordsmith", m10B1SoldierType, m10B1VeteranSwordsmithOracle, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			lord := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			mine := b12Creature(g, me.ID, "My Soldier", m10B1SoldierType, 1, 1)
			bear := b12Creature(g, me.ID, "My Bear", m10B1BearType, 2, 2)
			theirs := b12Creature(g, opp.ID, "Their Soldier", m10B1SoldierType, 1, 1)
			if p, tough := ptOf(t, g, mine); p != 1+tc.dp || tough != 1+tc.dt {
				t.Errorf("your Soldier is %d/%d, want %d/%d", p, tough, 1+tc.dp, 1+tc.dt)
			}
			for _, id := range []uuid.UUID{lord, bear} {
				if p, tough := ptOf(t, g, id); p != 2 || tough != 2 {
					t.Errorf("a creature the lord does not name is %d/%d, want 2/2", p, tough)
				}
			}
			if p, tough := ptOf(t, g, theirs); p != 1 || tough != 1 {
				t.Errorf("an opponent's Soldier is %d/%d, want 1/1", p, tough)
			}
		})
	}
}

func TestM10B1RhoxPikemasterGivesYourOtherSoldiersFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Rhox Pikemaster", "Creature — Rhino Soldier", m10B1RhoxPikemasterOracle, 3, 3)
	mine := b12Creature(g, me.ID, "My Soldier", m10B1SoldierType, 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", m10B1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Soldier", m10B1SoldierType, 1, 1)
	if !hasEffectiveKeyword(t, g, mine, "first strike") {
		t.Error("your Soldier has no first strike")
	}
	if hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("a non-Soldier has first strike")
	}
	if hasEffectiveKeyword(t, g, theirs, "first strike") {
		t.Error("an opponent's Soldier has first strike")
	}
}

func TestM10B1CaptainOfTheWatchMakesThreePumpedSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Soldier", m10B1SoldierType, 1, 1)
	captain := castHoldCreature(t, g, "Captain of the Watch", m10B1SoldierType, m10B1CaptainOfTheWatchOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(captain) {
		t.Fatal("Captain of the Watch did not resolve")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller != me.ID || !IsToken(c) || c.Name != "Soldier" {
			continue
		}
		n++
		if p, tough := ptOf(t, g, c.InstanceID); p != 2 || tough != 2 {
			t.Errorf("a Soldier token is %d/%d, want 2/2", p, tough)
		}
		if !hasEffectiveKeyword(t, g, c.InstanceID, "vigilance") {
			t.Error("a Soldier token has no vigilance")
		}
	}
	if n != 3 {
		t.Errorf("%d Soldier tokens, want 3", n)
	}
	if p, tough := ptOf(t, g, captain); p != 3 || tough != 3 {
		t.Errorf("the Captain is %d/%d, want 3/3", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("an opponent's Soldier is %d/%d, want 1/1", p, tough)
	}
}

func TestM10B1HonorOfThePurePumpsYourWhiteCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Honor of the Pure", "Enchantment", m10B1HonorOfThePureOracle, false)
	white := b12Creature(g, me.ID, "White", m10B1SoldierType, 2, 2)
	hmlSetColors(g, white, "W")
	green := b12Creature(g, me.ID, "Green", m10B1BearType, 2, 2)
	hmlSetColors(g, green, "G")
	theirs := b12Creature(g, opp.ID, "Their White", m10B1SoldierType, 2, 2)
	hmlSetColors(g, theirs, "W")
	if p, tough := ptOf(t, g, white); p != 3 || tough != 3 {
		t.Errorf("your white creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := ptOf(t, g, green); p != 2 || tough != 2 {
		t.Errorf("your green creature is %d/%d, want 2/2", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's white creature is %d/%d, want 2/2", p, tough)
	}
}

// --- triggered abilities ----------------------------------------------------------

func TestM10B1BorderlandRangerFetchesABasicLand(t *testing.T) {
	jmpB6FetchesABasicToHand(t, "Borderland Ranger", "Creature — Human Scout Ranger", m10B1BorderlandRangerOracle, 2, 2)
}

func TestM10B1BrambleCreeperPumpsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	id := b12Push(g, me.ID, "Bramble Creeper", "Creature — Elemental", m10B1BrambleCreeperOracle, 0, 3)
	declareAttack(t, g, them.ID, id)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, id); p != 5 || tough != 3 {
		t.Errorf("Bramble Creeper is %d/%d while attacking, want 5/3", p, tough)
	}
}

func TestM10B1IllusionaryServantIsSacrificedWhenTargeted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	servant := b12Push(g, me.ID, "Illusionary Servant", "Creature — Illusion", m10B1IllusionaryServantOracle, 3, 4)
	castCatalogSpell(t, g, "Disorient", "Instant", m10B1DisorientOracle, cardTarget(servant))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(servant) || !me.Graveyard.Contains(servant) {
		t.Error("the Servant was not sacrificed")
	}
}

// --- activated abilities -------------------------------------------------------------

func TestM10B1BlindingMageTapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mage := b12Push(g, me.ID, "Blinding Mage", "Creature — Human Wizard", m10B1BlindingMageOracle, 1, 2)
	bear := b12Creature(g, opp.ID, "Bear", m10B1BearType, 2, 2)
	apaMana(me, "W")
	apaActivate(t, g, me, mage, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !hmlTapped(t, g, bear) {
		t.Error("the target is untapped")
	}
	if !hmlTapped(t, g, mage) {
		t.Error("the Mage did not tap")
	}
}

func TestM10B1CudgelTrollRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	troll := b12Push(g, me.ID, "Cudgel Troll", "Creature — Troll", m10B1CudgelTrollOracle, 4, 3)
	apaMana(me, "G")
	apaActivate(t, g, me, troll, 0, game.ActivateAbilityParams{})
	if n := regenShieldsOn(t, g, troll); n != 1 {
		t.Fatalf("shields = %d, want 1", n)
	}
	destroyIt(t, g, troll)
	if !g.Battlefield.Contains(troll) {
		t.Error("the shield did not replace the destruction")
	}
}

func TestM10B1SelfPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
		p, tough, wantP, wantT       int
	}{
		{"Fiery Hellhound", "Creature — Elemental Dog", m10B1FieryHellhoundOracle, "R", 2, 2, 3, 2},
		{"Wall of Faith", "Creature — Wall", m10B1WallOfFaithOracle, "W", 0, 5, 0, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			apaMana(me, tc.mana)
			apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if p, tough := ptOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestM10B1VampireAristocratSacrificesForAPump(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	vampire := b12Push(g, me.ID, "Vampire Aristocrat", "Creature — Vampire Rogue Noble", m10B1VampireAristocratOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", m10B1BearType, 2, 2)
	apaActivate(t, g, me, vampire, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if p, tough := ptOf(t, g, vampire); p != 4 || tough != 4 {
		t.Errorf("the Aristocrat is %d/%d, want 4/4", p, tough)
	}
}

func TestM10B1AcolyteOfXathridDrainsAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	acolyte := b12Push(g, me.ID, "Acolyte of Xathrid", "Creature — Human Cleric", m10B1AcolyteOfXathridOracle, 0, 1)
	life := opp.Life
	apaMana(me, "C", "B")
	apaActivate(t, g, me, acolyte, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Errorf("the target is at %d, want %d", opp.Life, life-1)
	}
}

func TestM10B1GoblinArtilleryHitsTheTargetAndYou(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	artillery := b12Push(g, me.ID, "Goblin Artillery", "Creature — Goblin Warrior", m10B1GoblinArtilleryOracle, 1, 3)
	myLife, oppLife := me.Life, opp.Life
	apaActivate(t, g, me, artillery, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != oppLife-2 {
		t.Errorf("the target is at %d, want %d", opp.Life, oppLife-2)
	}
	if me.Life != myLife-3 {
		t.Errorf("you are at %d, want %d", me.Life, myLife-3)
	}
}

func TestM10B1UndeadSlayerExilesOnlyTheNamedTypes(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	slayer := b12Push(g, me.ID, "Undead Slayer", "Creature — Human Cleric", m10B1UndeadSlayerOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", m10B1BearType, 2, 2)
	vampire := b12Creature(g, opp.ID, "Vampire", "Creature — Vampire", 2, 2)
	apaMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, slayer, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err == nil {
		t.Fatal("Undead Slayer targeted a Bear")
	}
	apaActivate(t, g, me, slayer, 0, game.ActivateAbilityParams{Targets: cardTarget(vampire)})
	if g.Battlefield.Contains(vampire) || !g.Exile.Contains(vampire) {
		t.Error("the Vampire was not exiled")
	}
}

// --- the spells --------------------------------------------------------------------

func TestM10B1GraveyardReturnsToHand(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		good, bad              string
	}{
		{"Disentomb", "Sorcery", m10B1DisentombOracle, m10B1BearType, "Artifact"},
		{"Nature's Spiral", "Sorcery", m10B1NaturesSpiralOracle, "Enchantment", "Instant"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			good := m10B1GraveyardCard(me, "Good", tc.good)
			bad := m10B1GraveyardCard(me, "Bad", tc.bad)
			if err := castCatalogSpellErr(t, g, tc.name, tc.typeLine, tc.oracle, cardTarget(bad)); err == nil {
				t.Fatalf("%s targeted a %s card", tc.name, tc.bad)
			}
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, cardTarget(good))
			passPriorityAroundTable(t, g)
			if !me.Hand.Contains(good) || me.Graveyard.Contains(good) {
				t.Error("the card did not return to hand")
			}
		})
	}
}

func TestM10B1DivineVerdictDestroysOnlyACombatant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	raider := b12Creature(g, me.ID, "Raider", m10B1SoldierType, 2, 2)
	idle := b12Creature(g, me.ID, "Idle", m10B1SoldierType, 2, 2)
	brAttack(t, g, raider)
	if err := m10B1CastNow(g, "Divine Verdict", "Instant", m10B1DivineVerdictOracle, cardTarget(idle)); err == nil {
		t.Fatal("Divine Verdict targeted a creature that isn't in combat")
	}
	if err := m10B1CastNow(g, "Divine Verdict", "Instant", m10B1DivineVerdictOracle, cardTarget(raider)); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raider) {
		t.Error("the attacking creature survived")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("the idle creature was destroyed")
	}
}

func TestM10B1GloriousChargePumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := b12Creature(g, me.ID, "Mine", m10B1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", m10B1BearType, 2, 2)
	castCatalogSpell(t, g, "Glorious Charge", "Instant", m10B1GloriousChargeOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestM10B1SolemnOfferingDestroysAndGainsFour(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	bear := b12Creature(g, opp.ID, "Bear", m10B1BearType, 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Solemn Offering", "Sorcery", m10B1SolemnOfferingOracle, cardTarget(bear)); err == nil {
		t.Fatal("Solemn Offering targeted a creature")
	}
	life := me.Life
	castCatalogSpell(t, g, "Solemn Offering", "Sorcery", m10B1SolemnOfferingOracle, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestM10B1AngelsMercyGainsSeven(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	life := me.Life
	castCatalogSpell(t, g, "Angel's Mercy", "Instant", m10B1AngelsMercyOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+7 {
		t.Errorf("life %d, want %d", me.Life, life+7)
	}
}

func TestM10B1DisorientShrinksPowerUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 8, 8)
	castCatalogSpell(t, g, "Disorient", "Instant", m10B1DisorientOracle, cardTarget(giant))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, giant); p != 1 || tough != 8 {
		t.Errorf("the Giant is %d/%d, want 1/8", p, tough)
	}
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p, tough := ptOf(t, g, giant); p != 8 || tough != 8 {
		t.Errorf("the Giant is %d/%d next turn, want 8/8", p, tough)
	}
}

func TestM10B1RegenerateShieldsACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := b12Creature(g, me.ID, "Bear", m10B1BearType, 2, 2)
	castCatalogSpell(t, g, "Regenerate", "Instant", m10B1RegenerateOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if n := regenShieldsOn(t, g, bear); n != 1 {
		t.Fatalf("shields = %d, want 1", n)
	}
	destroyIt(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the shield did not replace the destruction")
	}
}

func TestM10B1SeismicStrikeCountsYourMountains(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	castCatalogSpell(t, g, "Seismic Strike", "Instant", m10B1SeismicStrikeOracle, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 2 {
		t.Errorf("the target has %d damage, want 2", got)
	}
}

func TestM10B1TomeScourMillsFive(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	library, graveyard := target.Library.Size(), target.Graveyard.Size()
	castCatalogSpell(t, g, "Tome Scour", "Sorcery", m10B1TomeScourOracle, []game.TargetRef{playerTarget(target.ID)})
	passPriorityAroundTable(t, g)
	if got := library - target.Library.Size(); got != 5 {
		t.Errorf("milled %d, want 5", got)
	}
	if got := target.Graveyard.Size() - graveyard; got != 5 {
		t.Errorf("%d cards reached the graveyard, want 5", got)
	}
}

// --- the Auras ---------------------------------------------------------------------

func TestM10B1OakenformPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", m10B1BearType, 2, 2)
	castCatalogSpell(t, g, "Oakenform", m10B1AuraType, m10B1OakenformOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the enchanted creature is %d/%d, want 5/5", p, tough)
	}
}

func TestM10B1LifelinkGrantsLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", m10B1BearType, 2, 2)
	other := b12Creature(g, me.ID, "Other", m10B1BearType, 2, 2)
	castCatalogSpell(t, g, "Lifelink", m10B1AuraType, m10B1LifelinkOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the enchanted creature has no lifelink")
	}
	if hasEffectiveKeyword(t, g, other, "lifelink") {
		t.Error("a creature the Aura is not on has lifelink")
	}
}
