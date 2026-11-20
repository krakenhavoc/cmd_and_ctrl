package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m13_5568_b1_cards_test.go — Magic 2013 batch 1 (#5568). Every test
// plays the real catalog entry through the engine.

const (
	m13B1AjanisSunstrikerOracle  = "31009c45-afba-45f4-a6ca-9dfdb9990e72"
	m13B1FaerieInvadersOracle    = "df7580fb-4c37-4c09-8785-e320cb124621"
	m13B1GuardianLionsOracle     = "9291f31d-6953-4fe1-9a88-b8dd09fde89c"
	m13B1PrimalHuntbeastOracle   = "44350127-f6b6-458f-8f2f-e57a78d83303"
	m13B1SentinelSpiderOracle    = "a37c9b13-dc45-4b55-a63d-5314abde7aab"
	m13B1SpikedBalothOracle      = "e873b9c2-1346-4f36-aa48-da58112792db"
	m13B1WarclampMastiffOracle   = "8bd301ea-32f0-4c3a-a215-c04b31fbdc7c"
	m13B1ServantOfNefaroxOracle  = "8740b9be-3ab4-4bd1-8a89-a59397c59e82"
	m13B1DuskmantleProwlerOracle = "9784ccf1-2238-417a-b6d1-cfc2e3dcae5f"
	m13B1KnightOfGloryOracle     = "10b8fd4d-1696-4b0d-b791-7bd3de504457"
	m13B1KnightOfInfamyOracle    = "e4b036ef-5319-4ccc-8777-7b1298119f90"
	m13B1RecklessBruteOracle     = "5ba6a472-b2f3-4d8e-838c-269aaa5eb2e3"
	m13B1ArcticAvenOracle        = "616cb1be-074a-46c2-b8f1-6b9f15b7bc23"
	m13B1CrimsonMuckwaderOracle  = "2b34f341-42bc-408a-85cc-b2d8797612b8"
	m13B1FlinthoofBoarOracle     = "acb58839-2b26-4724-9e9b-b49b8217b13f"
	m13B1HarborBanditOracle      = "18c65823-01ae-4314-a254-d89a1a362633"
	m13B1PrizedElephantOracle    = "4b1674b1-2176-493f-a463-936e715d51a2"
	m13B1AttendedKnightOracle    = "b3eb8f65-5eba-42a9-a1b4-4f37b14d03d9"
	m13B1BattleflightEagleOracle = "f36e316d-c0ad-414d-8cf0-b1c9ea779e03"
	m13B1BloodhunterBatOracle    = "7e3e09a2-b45c-4897-ba10-02628cca3f4c"
	m13B1BondBeetleOracle        = "9353f6dd-9a08-4e55-9b07-3e68d7d97ef4"
	m13B1ThundermawOracle        = "e958144e-01bc-46ff-8f56-ddf2eb098630"
	m13B1HealerOfThePrideOracle  = "69b9fa68-b409-4de1-9a40-5262386a0180"
	m13B1BloodReckoningOracle    = "bd5310df-2e6c-4584-9a5a-03d6d06b3c20"
	m13B1BattleJesterOracle      = "eba01ee1-55c4-48db-843b-fd6d48c32eda"
	m13B1RoaringPrimadoxOracle   = "2a098296-6a64-48ba-af40-5b68bae06bec"
	m13B1StaffOfNinOracle        = "48d9f488-4ebb-4fdb-a568-6930fc2f5416"
	m13B1ChronomatonOracle       = "610f25b0-770f-457a-b0eb-6deb832f71d4"
	m13B1RummagingGoblinOracle   = "2055eb91-ee36-4752-a6dc-581eaef335c8"
	m13B1WatercourserOracle      = "65dd4ae3-432c-4d57-ac14-4d827262ea0b"
	m13B1DutyBoundDeadOracle     = "44ee7def-6de3-473a-a0ae-6cac1b00dab4"
	m13B1FirewingPhoenixOracle   = "326d9b5c-ac04-45b5-9a8c-a9be7d8cca1b"
	m13B1WarFalconOracle         = "406bccb6-6602-4486-a556-e0b550650253"
	m13B1SublimeArchangelOracle  = "b5479c97-df19-49c7-b305-7a0336fd1cb3"
	m13B1MarkOfTheVampireOracle  = "af942d30-a191-4306-846d-6c26755ca3e6"
	m13B1CripplingBlightOracle   = "e365699a-bd77-473f-86b0-4e1ba0371cdc"
	m13B1TricksOfTheTradeOracle  = "9fd288d6-445e-4605-b206-aab43e300dde"
	m13B1CleaverRiotOracle       = "49f42896-90cd-4a05-bcca-12b19f780446"
	m13B1CowerInFearOracle       = "f1519fe3-af13-4491-9020-4a8d8f474713"
	m13B1CraterizeOracle         = "2825f7b0-21a9-44e2-b81f-542686981107"
	m13B1DownpourOracle          = "78eb5fb1-cb91-4300-9056-4225a416aec0"
	m13B1HydrosurgeOracle        = "4c2059d5-9a2e-4b89-a617-9125eb0cd570"
	m13B1MindSculptOracle        = "4d8592d7-2e36-4a7a-be9e-7c8f512d5f62"
	m13B1SearingSpearOracle      = "aafd44c1-74a9-4aa7-a4a2-67eb39a07478"
	m13B1SerpentsGiftOracle      = "b75fb789-6a78-4d23-b872-7627b478fb6f"
	m13B1ShowOfValorOracle       = "9645c1cb-2305-4bc0-89d0-a50815e91573"
	m13B1SmeltOracle             = "b5f4b732-099c-46e7-a03d-d8e4c78245b8"
	m13B1WildGuessOracle         = "a05ab59e-680a-4050-aecb-56ee69dcd13f"
	m13B1RangersPathOracle       = "259a023a-de55-4bab-9e24-01aebcf716eb"
	m13B1VileRebirthOracle       = "6845a62f-e0f1-448c-beeb-abac2dabd975"

	m13B1BearType   = "Creature — Bear"
	m13B1KnightType = "Creature — Human Knight"
	m13B1FirstStrk  = "first strike"
)

// m13B1Flyer puts a creature with flying on the battlefield under p.
func m13B1Flyer(g *game.Game, p uuid.UUID, name string, power, toughness int) uuid.UUID {
	return apaPush(g, p, p, game.Card{Name: name, TypeLine: "Creature — Bird", Power: power, Toughness: toughness, Keywords: []string{"flying"}})
}

// --- every card is in the catalog ----------------------------------------

func TestM13B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		m13B1AjanisSunstrikerOracle, m13B1FaerieInvadersOracle, m13B1GuardianLionsOracle, m13B1PrimalHuntbeastOracle,
		m13B1SentinelSpiderOracle, m13B1SpikedBalothOracle, m13B1WarclampMastiffOracle, m13B1ServantOfNefaroxOracle,
		m13B1DuskmantleProwlerOracle, m13B1KnightOfGloryOracle, m13B1KnightOfInfamyOracle, m13B1RecklessBruteOracle,
		m13B1ArcticAvenOracle, m13B1CrimsonMuckwaderOracle, m13B1FlinthoofBoarOracle, m13B1HarborBanditOracle,
		m13B1PrizedElephantOracle, m13B1AttendedKnightOracle, m13B1BattleflightEagleOracle, m13B1BloodhunterBatOracle,
		m13B1BondBeetleOracle, m13B1ThundermawOracle, m13B1HealerOfThePrideOracle, m13B1BloodReckoningOracle,
		m13B1BattleJesterOracle, m13B1RoaringPrimadoxOracle, m13B1StaffOfNinOracle, m13B1ChronomatonOracle,
		m13B1RummagingGoblinOracle, m13B1WatercourserOracle, m13B1DutyBoundDeadOracle, m13B1FirewingPhoenixOracle,
		m13B1WarFalconOracle, m13B1SublimeArchangelOracle, m13B1MarkOfTheVampireOracle, m13B1CripplingBlightOracle,
		m13B1TricksOfTheTradeOracle, m13B1CleaverRiotOracle, m13B1CowerInFearOracle, m13B1CraterizeOracle,
		m13B1DownpourOracle, m13B1HydrosurgeOracle, m13B1MindSculptOracle, m13B1SearingSpearOracle,
		m13B1SerpentsGiftOracle, m13B1ShowOfValorOracle, m13B1SmeltOracle, m13B1WildGuessOracle,
		m13B1RangersPathOracle, m13B1VileRebirthOracle,
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

// --- printed keywords -------------------------------------------------------

func TestM13B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Ajani's Sunstriker", "Creature — Cat Cleric", m13B1AjanisSunstrikerOracle, []string{"lifelink"}},
		{"Faerie Invaders", "Creature — Faerie Rogue", m13B1FaerieInvadersOracle, []string{"flash", "flying"}},
		{"Guardian Lions", "Creature — Cat", m13B1GuardianLionsOracle, []string{"vigilance"}},
		{"Primal Huntbeast", "Creature — Beast", m13B1PrimalHuntbeastOracle, []string{"hexproof"}},
		{"Sentinel Spider", "Creature — Spider", m13B1SentinelSpiderOracle, []string{"reach", "vigilance"}},
		{"Spiked Baloth", "Creature — Beast", m13B1SpikedBalothOracle, []string{"trample"}},
		{"Warclamp Mastiff", "Creature — Dog", m13B1WarclampMastiffOracle, []string{m13B1FirstStrk}},
		{"Servant of Nefarox", "Creature — Human Cleric", m13B1ServantOfNefaroxOracle, []string{game.KeywordExalted}},
		{"Duskmantle Prowler", "Creature — Vampire Rogue", m13B1DuskmantleProwlerOracle, []string{"haste", game.KeywordExalted}},
		{"Knight of Glory", m13B1KnightType, m13B1KnightOfGloryOracle, []string{"protection from black", game.KeywordExalted}},
		{"Knight of Infamy", m13B1KnightType, m13B1KnightOfInfamyOracle, []string{"protection from white", game.KeywordExalted}},
		{"Reckless Brute", "Creature — Ogre Warrior", m13B1RecklessBruteOracle, []string{"haste"}},
		{"Arctic Aven", "Creature — Bird Wizard", m13B1ArcticAvenOracle, []string{"flying"}},
		{"Attended Knight", m13B1KnightType, m13B1AttendedKnightOracle, []string{m13B1FirstStrk}},
		{"Battleflight Eagle", "Creature — Bird", m13B1BattleflightEagleOracle, []string{"flying"}},
		{"Bloodhunter Bat", "Creature — Bat", m13B1BloodhunterBatOracle, []string{"flying"}},
		{"Thundermaw Hellkite", "Creature — Dragon", m13B1ThundermawOracle, []string{"flying", "haste"}},
		{"Duty-Bound Dead", "Creature — Skeleton", m13B1DutyBoundDeadOracle, []string{game.KeywordExalted}},
		{"Firewing Phoenix", "Creature — Phoenix", m13B1FirewingPhoenixOracle, []string{"flying"}},
		{"War Falcon", "Creature — Bird", m13B1WarFalconOracle, []string{"flying"}},
		{"Sublime Archangel", "Creature — Angel", m13B1SublimeArchangelOracle, []string{"flying", game.KeywordExalted}},
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

// Protection from white (CR 702.16b): a white spell can't target Knight
// of Infamy, and can target Knight of Glory.
func TestM13B1KnightOfInfamyCantBeTargetedByWhite(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	infamy := b12Push(g, me.ID, "Knight of Infamy", m13B1KnightType, m13B1KnightOfInfamyOracle, 2, 1)
	hmlSetColors(g, infamy, "B")
	glory := b12Push(g, me.ID, "Knight of Glory", m13B1KnightType, m13B1KnightOfGloryOracle, 2, 1)
	hmlSetColors(g, glory, "W")
	white := func() uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Show of Valor", TypeLine: "Instant", ManaCost: "{1}{W}", Colors: []string{"W"},
			OracleID: m13B1ShowOfValorOracle, Owner: me.ID, Controller: me.ID,
		})
		return id
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})
	if err := g.CastSpell(me.ID, white(), game.CastSpellParams{Targets: cardTarget(infamy)}); err == nil {
		t.Error("a white spell targeted a creature with protection from white")
	}
	if err := g.CastSpell(me.ID, white(), game.CastSpellParams{Targets: cardTarget(glory)}); err != nil {
		t.Errorf("a white spell can't target Knight of Glory: %v", err)
	}
}

// --- exalted ------------------------------------------------------------------

func TestM13B1ExaltedPumpsALoneAttacker(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Servant of Nefarox", "Creature — Human Cleric", m13B1ServantOfNefaroxOracle},
		{"Duskmantle Prowler", "Creature — Vampire Rogue", m13B1DuskmantleProwlerOracle},
		{"Knight of Glory", m13B1KnightType, m13B1KnightOfGloryOracle},
		{"Knight of Infamy", m13B1KnightType, m13B1KnightOfInfamyOracle},
		{"Duty-Bound Dead", "Creature — Skeleton", m13B1DutyBoundDeadOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 0, 4)
			bear := seedBear(g, me.ID)
			alaB1Attack(t, g, opp.ID, bear)
			if got := effectivePower(t, g, bear); got != 3 {
				t.Errorf("the lone attacker has power %d, want 3", got)
			}
		})
	}
}

// Sublime Archangel's own exalted plus one granted instance on each other
// creature you control, the attacker included: three instances here.
func TestM13B1SublimeArchangelGivesYourOtherCreaturesExalted(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Sublime Archangel", "Creature — Angel", m13B1SublimeArchangelOracle, 4, 3)
	bear := seedBear(g, me.ID)
	walker := b12Creature(g, me.ID, "Walker", m13B1BearType, 1, 1)
	theirs := b12Creature(g, opp.ID, "Theirs", m13B1BearType, 1, 1)
	if !hasAbility(effectiveAbilities(t, g, walker), game.KeywordExalted) {
		t.Fatal("another creature you control has no exalted")
	}
	if hasAbility(effectiveAbilities(t, g, theirs), game.KeywordExalted) {
		t.Fatal("an opponent's creature has exalted")
	}
	alaB1Attack(t, g, opp.ID, bear)
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("the lone attacker has power %d, want 5 (three exalted instances)", got)
	}
}

// --- combat statics -------------------------------------------------------------

func TestM13B1RecklessBruteMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	brute := b12Push(g, me.ID, "Reckless Brute", "Creature — Ogre Warrior", m13B1RecklessBruteOracle, 3, 1)
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != brute {
		t.Fatalf("pass with the Brute home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(brute, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Brute attacking: %v", err)
	}
}

func TestM13B1WarFalconAttacksOnlyWithAKnightOrSoldier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	falcon := b12Push(g, me.ID, "War Falcon", "Creature — Bird", m13B1WarFalconOracle, 2, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(falcon, opp.ID); err == nil {
		t.Fatal("War Falcon attacked with no Knight or Soldier")
	}
	b12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 1, 1)
	if err := g.DeclareAttacker(falcon, opp.ID); err != nil {
		t.Errorf("War Falcon can't attack beside a Soldier: %v", err)
	}
}

// --- land-type pumps ---------------------------------------------------------------

func TestM13B1LandTypePumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, land string
		power, toughness             int
	}{
		{"Arctic Aven", "Creature — Bird Wizard", m13B1ArcticAvenOracle, "Plains", 2, 1},
		{"Crimson Muckwader", "Creature — Lizard", m13B1CrimsonMuckwaderOracle, "Swamp", 2, 1},
		{"Flinthoof Boar", "Creature — Boar", m13B1FlinthoofBoarOracle, "Mountain", 2, 2},
		{"Harbor Bandit", "Creature — Human Rogue", m13B1HarborBanditOracle, "Island", 2, 2},
		{"Prized Elephant", "Creature — Elephant", m13B1PrizedElephantOracle, "Forest", 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			b12Permanent(g, opp.ID, tc.land, "Basic Land — "+tc.land)
			if p, tough := ptOf(t, g, id); p != tc.power || tough != tc.toughness {
				t.Fatalf("with only an opponent's %s it is %d/%d, want %d/%d", tc.land, p, tough, tc.power, tc.toughness)
			}
			b12Permanent(g, me.ID, tc.land, "Basic Land — "+tc.land)
			if p, tough := ptOf(t, g, id); p != tc.power+1 || tough != tc.toughness+1 {
				t.Errorf("with your %s it is %d/%d, want %d/%d", tc.land, p, tough, tc.power+1, tc.toughness+1)
			}
		})
	}
}

// --- activated abilities -------------------------------------------------------------

func TestM13B1KeywordForMana(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana, keyword string
	}{
		{"Arctic Aven", "Creature — Bird Wizard", m13B1ArcticAvenOracle, "W", "lifelink"},
		{"Flinthoof Boar", "Creature — Boar", m13B1FlinthoofBoarOracle, "R", "haste"},
		{"Prized Elephant", "Creature — Elephant", m13B1PrizedElephantOracle, "G", "trample"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Fatalf("%s already has %s", tc.name, tc.keyword)
			}
			apaMana(me, tc.mana)
			apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s did not gain %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestM13B1HarborBanditCantBeBlockedForMana(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bandit := b12Push(g, me.ID, "Harbor Bandit", "Creature — Human Rogue", m13B1HarborBanditOracle, 2, 2)
	apaMana(me, "U", "U")
	apaActivate(t, g, me, bandit, 0, game.ActivateAbilityParams{})
	if restrictionsOf(t, g, bandit)&game.CantBeBlocked == 0 {
		t.Error("Harbor Bandit can still be blocked")
	}
}

func TestM13B1RegenerateForMana(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		mana                   []string
	}{
		{"Crimson Muckwader", "Creature — Lizard", m13B1CrimsonMuckwaderOracle, []string{"C", "C", "B"}},
		{"Duty-Bound Dead", "Creature — Skeleton", m13B1DutyBoundDeadOracle, []string{"C", "C", "C", "B"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			b27Kill(g, id)
			if !g.Battlefield.Contains(id) {
				t.Error("the shielded creature was destroyed")
			}
		})
	}
}

func TestM13B1WatercourserTradesToughnessForPower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	id := b12Push(g, me.ID, "Watercourser", "Creature — Elemental", m13B1WatercourserOracle, 2, 3)
	apaMana(me, "U")
	apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if p, tough := ptOf(t, g, id); p != 3 || tough != 2 {
		t.Errorf("Watercourser is %d/%d, want 3/2", p, tough)
	}
}

func TestM13B1ChronomatonGrowsItself(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	id := b12Push(g, me.ID, "Chronomaton", "Artifact Creature — Golem", m13B1ChronomatonOracle, 1, 1)
	apaMana(me, "C")
	apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !hmlTapped(t, g, id) {
		t.Error("Chronomaton did not tap")
	}
}

func TestM13B1RummagingGoblinLootsAndNeedsACardToDiscard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	goblin := b12Push(g, me.ID, "Rummaging Goblin", "Creature — Goblin Rogue", m13B1RummagingGoblinOracle, 1, 1)
	me.Hand.Cards = me.Hand.Cards[:0]
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no card to discard")
	}
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discard cost was not paid")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 1 {
		t.Errorf("hand is %d, want the one drawn card", me.Hand.Size())
	}
}

func TestM13B1FirewingPhoenixReturnsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	phoenix := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: phoenix, Name: "Firewing Phoenix", TypeLine: "Creature — Phoenix",
		OracleID: m13B1FirewingPhoenixOracle, Owner: me.ID, Controller: me.ID})
	apaMana(me, "C", "R", "R", "R")
	b16Activate(t, g, me.ID, phoenix, 0, game.ActivateAbilityParams{})
	if !me.Hand.Contains(phoenix) {
		t.Error("Firewing Phoenix did not return to hand")
	}
}

func TestM13B1StaffOfNinPingsAndDrawsEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	staff := pushCatalogPermanent(g, me.ID, "Staff of Nin", "Artifact", m13B1StaffOfNinOracle, false)
	walkToUpkeepOfSeat0(t, g)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand is %d after the upkeep trigger, want %d", got, hand+1)
	}
	life := opp.Life
	apaActivate(t, g, me, staff, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
}

// --- triggered abilities ---------------------------------------------------------

func TestM13B1AttendedKnightMakesASoldier(t *testing.T) {
	g := newCatalogGame(t)
	castHoldCreature(t, g, "Attended Knight", m13B1KnightType, m13B1AttendedKnightOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if got := onBattlefieldNamed(g, "Soldier"); got != 1 {
		t.Errorf("%d Soldier tokens, want 1", got)
	}
}

func TestM13B1BattleflightEaglePumpsAndGrantsFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", m13B1BearType, 2, 2)
	castHoldCreature(t, g, "Battleflight Eagle", "Creature — Bird", m13B1BattleflightEagleOracle, 2, 2)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear has no flying")
	}
}

func TestM13B1BondBeetlePutsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", m13B1BearType, 2, 2)
	castHoldCreature(t, g, "Bond Beetle", "Creature — Insect", m13B1BondBeetleOracle, 0, 1)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", got)
	}
}

func TestM13B1BloodhunterBatDrainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := top100Opponent(g)
	life, oppLife := me.Life, opp.Life
	castHoldCreature(t, g, "Bloodhunter Bat", "Creature — Bat", m13B1BloodhunterBatOracle, 2, 2)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("the target is at %d, want %d", opp.Life, oppLife-2)
	}
	if me.Life != life+2 {
		t.Errorf("you are at %d, want %d", me.Life, life+2)
	}
}

// Thundermaw Hellkite hits only the flyers an opponent controls, and taps
// the survivors.
func TestM13B1ThundermawHellkiteGroundsOpposingFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := top100Opponent(g)
	small := m13B1Flyer(g, opp.ID, "Small Flyer", 1, 1)
	big := m13B1Flyer(g, opp.ID, "Big Flyer", 3, 3)
	walker := b12Creature(g, opp.ID, "Walker", m13B1BearType, 2, 2)
	mine := m13B1Flyer(g, me.ID, "My Flyer", 1, 1)
	castHoldCreature(t, g, "Thundermaw Hellkite", "Creature — Dragon", m13B1ThundermawOracle, 5, 5)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("the 1/1 flyer survived 1 damage")
	}
	if got := findCardAnywhere(t, g, big).DamageMarked; got != 1 {
		t.Errorf("the 3/3 flyer has %d damage, want 1", got)
	}
	if !hmlTapped(t, g, big) {
		t.Error("the surviving flyer is untapped")
	}
	if got := findCardAnywhere(t, g, walker).DamageMarked; got != 0 || hmlTapped(t, g, walker) {
		t.Error("a creature without flying was hit or tapped")
	}
	if !g.Battlefield.Contains(mine) || hmlTapped(t, g, mine) {
		t.Error("your own flyer was hit or tapped")
	}
}

func TestM13B1HealerOfThePrideGainsTwoPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Healer of the Pride", "Creature — Cat Cleric", m13B1HealerOfThePrideOracle, 2, 3)
	life := me.Life
	castHoldCreature(t, g, "Bear", m13B1BearType, "", 2, 2)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestM13B1BloodReckoningDrainsEachAttackerAtYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Blood Reckoning", "Enchantment", m13B1BloodReckoningOracle, false)
	a := b12Creature(g, opp.ID, "Attacker A", m13B1BearType, 2, 2)
	b := b12Creature(g, opp.ID, "Attacker B", m13B1BearType, 2, 2)
	c := b12Creature(g, opp.ID, "Attacker C", m13B1BearType, 2, 2)
	life, myLife, thirdLife := opp.Life, me.Life, third.Life
	advanceToDeclareAttackersOf(t, g, 1)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, me.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	if err := g.DeclareAttacker(c, third.ID); err != nil {
		t.Fatalf("DeclareAttacker at a third player: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("the attackers' controller lost %d life, want 2", got)
	}
	if me.Life != myLife || third.Life != thirdLife {
		t.Error("a player other than the attackers' controller changed life")
	}
}

func TestM13B1GoblinBattleJesterRedSpellStopsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Goblin Battle Jester", "Creature — Goblin", m13B1BattleJesterOracle, 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", m13B1BearType, 2, 2)
	floatForTest(g, me, "R")
	castFromHandForTest(t, g, me, "Spell", "Instant", "{R}", "", strict)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, blocker)
	passPriorityAroundTable(t, g)
	if restrictionsOf(t, g, blocker)&game.CantBlock == 0 {
		t.Error("the target can still block")
	}
}

func TestM13B1RoaringPrimadoxReturnsACreatureEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Roaring Primadox", "Creature — Beast", m13B1RoaringPrimadoxOracle, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", m13B1BearType, 2, 2)
	walkToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Roaring Primadox queued no return pick")
	}
	if p.ChooseMin != 1 {
		t.Error("the return is mandatory")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the Bear did not return to hand")
	}
}

// --- auras -------------------------------------------------------------------

func TestM13B1Auras(t *testing.T) {
	for _, tc := range []struct {
		name, oracle     string
		power, toughness int
		keyword          string
		restriction      game.Restriction
	}{
		{"Mark of the Vampire", m13B1MarkOfTheVampireOracle, 5, 5, "lifelink", 0},
		{"Crippling Blight", m13B1CripplingBlightOracle, 2, 2, "", game.CantBlock},
		{"Tricks of the Trade", m13B1TricksOfTheTradeOracle, 5, 3, "", game.CantBeBlocked},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := b12Creature(g, me.ID, "Bear", m13B1BearType, 3, 3)
			castCatalogSpell(t, g, tc.name, auraTypeLine, tc.oracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, bear); p != tc.power || tough != tc.toughness {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.power, tc.toughness)
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
			if tc.restriction != 0 && restrictionsOf(t, g, bear)&tc.restriction == 0 {
				t.Errorf("the Bear lacks restriction %v", tc.restriction)
			}
		})
	}
}

// --- the spells ----------------------------------------------------------------

func TestM13B1CleaverRiotGivesYourCreaturesDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := b12Creature(g, me.ID, "Mine", m13B1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", m13B1BearType, 2, 2)
	castCatalogSpell(t, g, "Cleaver Riot", "Sorcery", m13B1CleaverRiotOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, mine, "double strike") {
		t.Error("your creature has no double strike")
	}
	if hasEffectiveKeyword(t, g, theirs, "double strike") {
		t.Error("an opponent's creature has double strike")
	}
}

func TestM13B1CowerInFearShrinksOnlyOpponents(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := b12Creature(g, me.ID, "Mine", m13B1BearType, 2, 2)
	small := b12Creature(g, opp.ID, "Small", m13B1BearType, 1, 1)
	big := b12Creature(g, opp.ID, "Big", m13B1BearType, 3, 3)
	castCatalogSpell(t, g, "Cower in Fear", "Instant", m13B1CowerInFearOracle, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("the opponent's 1/1 survived")
	}
	if p, tough := ptOf(t, g, big); p != 2 || tough != 2 {
		t.Errorf("the opponent's 3/3 is %d/%d, want 2/2", p, tough)
	}
	if p, tough := ptOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("your creature is %d/%d, want 2/2", p, tough)
	}
}

func TestM13B1DestroyTargetPermanentOfAType(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, good, bad string
	}{
		{"Craterize", "Sorcery", m13B1CraterizeOracle, "Basic Land — Mountain", "Artifact"},
		{"Smelt", "Instant", m13B1SmeltOracle, "Artifact", "Basic Land — Mountain"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			opp := top100Opponent(g)
			good := b12Permanent(g, opp.ID, "Good", tc.good)
			bad := b12Permanent(g, opp.ID, "Bad", tc.bad)
			if err := castCatalogSpellErr(t, g, tc.name, tc.typeLine, tc.oracle, cardTarget(bad)); err == nil {
				t.Fatalf("%s targeted a %s", tc.name, tc.bad)
			}
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, cardTarget(good))
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(good) {
				t.Error("the target survived")
			}
		})
	}
}

func TestM13B1DownpourTapsUpToThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	var ids []uuid.UUID
	for i := 0; i < 3; i++ {
		ids = append(ids, b12Creature(g, opp.ID, "Target", m13B1BearType, 2, 2))
	}
	castCatalogSpell(t, g, "Downpour", "Instant", m13B1DownpourOracle, cardTargets(ids...))
	passPriorityAroundTable(t, g)
	for i, id := range ids {
		if !hmlTapped(t, g, id) {
			t.Errorf("target %d is untapped", i)
		}
	}
}

func TestM13B1TargetPumps(t *testing.T) {
	for _, tc := range []struct {
		name, oracle     string
		power, toughness int
		keyword          string
	}{
		{"Hydrosurge", m13B1HydrosurgeOracle, 0, 5, ""},
		{"Show of Valor", m13B1ShowOfValorOracle, 7, 9, ""},
		{"Serpent's Gift", m13B1SerpentsGiftOracle, 5, 5, "deathtouch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			bear := b12Creature(g, me.ID, "Bear", m13B1BearType, 5, 5)
			castCatalogSpell(t, g, tc.name, "Instant", tc.oracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, bear); p != tc.power || tough != tc.toughness {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.power, tc.toughness)
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
		})
	}
}

func TestM13B1SearingSpearDealsThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	life := opp.Life
	castCatalogSpell(t, g, "Searing Spear", "Instant", m13B1SearingSpearOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-3)
	}
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	castCatalogSpell(t, g, "Searing Spear", "Instant", m13B1SearingSpearOracle, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 3 {
		t.Errorf("the creature has %d damage, want 3", got)
	}
}

func TestM13B1MindSculptMillsAnOpponentSeven(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	if err := castCatalogSpellErr(t, g, "Mind Sculpt", "Sorcery", m13B1MindSculptOracle, []game.TargetRef{playerTarget(me.ID)}); err == nil {
		t.Fatal("Mind Sculpt targeted its caster")
	}
	library := opp.Library.Size()
	castCatalogSpell(t, g, "Mind Sculpt", "Sorcery", m13B1MindSculptOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := library - opp.Library.Size(); got != 7 {
		t.Errorf("milled %d, want 7", got)
	}
}

func TestM13B1WildGuessDrawsTwoAfterTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	before := me.Library.Size()
	_, paid := castWithDiscard(t, g, "Wild Guess", m13B1WildGuessOracle, "Sorcery", 1)
	if !me.Graveyard.Contains(paid[0]) {
		t.Fatal("the discard was not paid at announce")
	}
	passPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

func TestM13B1RangersPathFetchesTwoForestsTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedBasicsInLibrary(me, 2)
	castCatalogSpell(t, g, "Ranger's Path", "Sorcery", m13B1RangersPathOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids...)
	for i, id := range ids {
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Errorf("Forest %d was not fetched", i)
			continue
		}
		if !c.Tapped {
			t.Errorf("Forest %d entered untapped", i)
		}
	}
}

func TestM13B1VileRebirthExilesAndMakesAZombie(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	rock := wthGraveyardCard(opp, "Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Vile Rebirth", "Instant", m13B1VileRebirthOracle, cardTarget(rock)); err == nil {
		t.Fatal("Vile Rebirth targeted a noncreature card")
	}
	corpse := wthGraveyardCard(opp, "Corpse", m13B1BearType)
	castCatalogSpell(t, g, "Vile Rebirth", "Instant", m13B1VileRebirthOracle, cardTarget(corpse))
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(corpse) {
		t.Error("the creature card is still in the graveyard")
	}
	if got := onBattlefieldNamed(g, "Zombie"); got != 1 {
		t.Errorf("%d Zombie tokens, want 1", got)
	}
}
