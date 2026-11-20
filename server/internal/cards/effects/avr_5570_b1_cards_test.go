package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avr_5570_b1_cards_test.go — Avacyn Restored, batch 1 (#5570). Every
// non-vanilla card is played through the engine.

const (
	avrNettleSwineOracle         = "fe412468-b506-4ed8-8fff-3987ef3dc2ee"
	avrRagingPoltergeistOracle   = "15d460da-9110-4f33-a24f-8582a2bcd96f"
	avrRenegadeDemonOracle       = "5f444efc-3bc2-4ca8-986c-f804f9e1a95e"
	avrVorstclawOracle           = "aa944cbe-fdba-4001-97d5-c722fe744dcc"
	avrSeraphOfDawnOracle        = "38c588ae-7254-4fae-aa9a-03e2a5524492"
	avrThrabenValiantOracle      = "dee0b358-b9da-49b9-a47f-7907cde6d8ee"
	avrHauntedGuardianOracle     = "20b9bc53-7f6d-4616-a2df-5783ab577707"
	avrFarbogExplorerOracle      = "f474c1f8-43bb-4e56-8131-27c71b3193d2"
	avrButcherGhoulOracle        = "28caeed4-c175-469f-8d8e-f45bc0a51a1c"
	avrHoundOfGriselbrandOracle  = "17425a60-7554-4740-9455-976edd63c106"
	avrLatchSeekerOracle         = "4c1f62a8-d905-48aa-aba1-4bcb522a7ef6"
	avrScrapskinDrakeOracle      = "910e8269-6001-479a-ab2e-5a935569c33c"
	avrWanderingWolfOracle       = "1750cf0a-0709-4f02-90ea-572ff88a37a5"
	avrCathedralSanctifierOracle = "1050f029-60d4-4769-b39f-9100c949a822"
	avrGryffVanguardOracle       = "3bcdf832-4ce2-4043-b46a-0c5fd7019fcd"
	avrFerventCatharOracle       = "94bcae2f-86a6-4d38-b8e1-5ebf69dc7cbc"
	avrMistRavenOracle           = "12654457-3aae-4046-a976-8e41c3f78927"
	avrVoiceOfTheProvincesOracle = "81b15ed1-7069-4f8b-96b9-f6d67298afef"
	avrGoldnightRedeemerOracle   = "6fe35d1f-20b6-4fa8-a318-79e2d415701a"
	avrEmancipationAngelOracle   = "3834d3b0-9b66-465a-a8dc-22875a819fd9"
	avrPollutedDeadOracle        = "5ffca600-5286-4086-9efe-c398babeb6ae"
	avrRotcrownGhoulOracle       = "e9a420d4-a834-4a14-b874-1db96624d8ef"
	avrRiotRingleaderOracle      = "420d09b5-aced-4f71-914c-90d19f7476dc"
	avrHeirsOfStromkirkOracle    = "7149de06-d676-47bc-a50f-e1058d89c890"
	avrAlchemistsApprenticeOracl = "de085c51-2555-450e-a479-ddbb3bea9e8c"
	avrMadProphetOracle          = "91081181-aec6-47de-b8ef-241a2f4fe880"
	avrMarrowBatsOracle          = "f20d96cc-3949-4888-a909-c7d67c4d31c7"
	avrWolfirAvengerOracle       = "83deda94-66e8-4655-9cd5-6d97d809946d"
	avrMoorlandInquisitorOracle  = "76bbd280-dcbc-43b3-a8a1-603443b2ccaa"
	avrSearchlightGeistOracle    = "f9366a4a-54d0-4c65-920b-d8ddd475db3c"
	avrScaldingDevilOracle       = "3bb1536f-06cc-433f-aeeb-afd528f9af03"
	avrBannersRaisedOracle       = "b60d83ea-18bf-4c64-b4ee-981d54629362"
	avrNaturalEndOracle          = "05b5c0fb-84b2-4a32-bf66-179c4dccb3df"
	avrCursebreakOracle          = "ac5a13b1-0725-4978-8c93-cebd76be58ba"
	avrZealousStrikeOracle       = "d820ddf4-315d-470a-8d66-663a76a240c6"
	avrUncannySpeedOracle        = "85b0a2ea-dbdc-456e-9ce1-faa8c13f84d0"
	avrSnareTheSkiesOracle       = "7d011e22-335b-465f-9111-e95a72ebffe9"
	avrElgaudShieldmateOracle    = "a85f7e34-7783-474c-ac96-9b075b296b8f"
	avrGeistTrappersOracle       = "7116da4a-96c1-4183-b123-29c6b70ef543"
	avrHanweirLancerOracle       = "5fa9de04-6b70-4162-9ffb-6334277a4f30"
	avrLightningMaulerOracle     = "6c5664b1-fbb5-477e-80fd-cf2285014a10"
	avrNearheathPilgrimOracle    = "54c10a1d-360a-42cb-bb3b-c3ab57e1d7ea"
	avrNightshadePeddlerOracle   = "afd6089a-38cc-4a0f-853b-94d6be2d4470"
	avrPathbreakerWurmOracle     = "ab342d57-240a-4c56-aed6-ac3cdc51af8d"
	avrSilverbladePaladinOracle  = "ded0316f-5eed-4a80-91ad-ca236fd16f01"
	avrSpectralGateguardsOracle  = "334f7cd5-6c7b-4592-b564-8a0edac4a1b4"
	avrWingcrafterOracle         = "55d1f44f-8554-49e7-b36f-92682926e6d0"
	avrDruidsFamiliarOracle      = "6a5b1cd4-e597-4122-8fc7-4dce241d7c01"
	avrTrustedForcemageOracle    = "663ef3c9-f897-4875-b029-82ee6672db00"
	avrWolfirSilverheartOracle   = "dccf9497-cc7b-42df-9b46-f5ffea0c7474"

	avrBearType = "Creature — Bear"
)

// --- every card is in the catalog ----------------------------------------

func TestAvr5570B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		avrNettleSwineOracle, avrRagingPoltergeistOracle, avrRenegadeDemonOracle, avrVorstclawOracle,
		avrSeraphOfDawnOracle, avrThrabenValiantOracle, avrHauntedGuardianOracle, avrFarbogExplorerOracle,
		avrButcherGhoulOracle, avrHoundOfGriselbrandOracle, avrLatchSeekerOracle, avrScrapskinDrakeOracle,
		avrWanderingWolfOracle, avrCathedralSanctifierOracle, avrGryffVanguardOracle, avrFerventCatharOracle,
		avrMistRavenOracle, avrVoiceOfTheProvincesOracle, avrGoldnightRedeemerOracle, avrEmancipationAngelOracle,
		avrPollutedDeadOracle, avrRotcrownGhoulOracle, avrRiotRingleaderOracle, avrHeirsOfStromkirkOracle,
		avrAlchemistsApprenticeOracl, avrMadProphetOracle, avrMarrowBatsOracle, avrWolfirAvengerOracle,
		avrMoorlandInquisitorOracle, avrSearchlightGeistOracle, avrScaldingDevilOracle, avrBannersRaisedOracle,
		avrNaturalEndOracle, avrCursebreakOracle, avrZealousStrikeOracle, avrUncannySpeedOracle,
		avrSnareTheSkiesOracle, avrElgaudShieldmateOracle, avrGeistTrappersOracle, avrHanweirLancerOracle,
		avrLightningMaulerOracle, avrNearheathPilgrimOracle, avrNightshadePeddlerOracle, avrPathbreakerWurmOracle,
		avrSilverbladePaladinOracle, avrSpectralGateguardsOracle, avrWingcrafterOracle, avrDruidsFamiliarOracle,
		avrTrustedForcemageOracle, avrWolfirSilverheartOracle,
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

// --- printed keywords -----------------------------------------------------

func TestAvr5570B1PrintedKeywords(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Seraph of Dawn", "Creature — Angel", avrSeraphOfDawnOracle, []string{"flying", "lifelink"}},
		{"Thraben Valiant", "Creature — Human Soldier", avrThrabenValiantOracle, []string{"vigilance"}},
		{"Haunted Guardian", "Artifact Creature — Construct", avrHauntedGuardianOracle, []string{"defender", "first strike"}},
		{"Farbog Explorer", "Creature — Human Scout", avrFarbogExplorerOracle, []string{"swampwalk"}},
		{"Butcher Ghoul", "Creature — Zombie", avrButcherGhoulOracle, []string{game.KeywordUndying}},
		{"Hound of Griselbrand", "Creature — Elemental Dog", avrHoundOfGriselbrandOracle, []string{"double strike", game.KeywordUndying}},
		{"Scrapskin Drake", "Creature — Zombie Drake", avrScrapskinDrakeOracle, []string{"flying"}},
		{"Gryff Vanguard", "Creature — Human Knight", avrGryffVanguardOracle, []string{"flying"}},
		{"Fervent Cathar", "Creature — Human Knight", avrFerventCatharOracle, []string{"haste"}},
		{"Mist Raven", "Creature — Bird", avrMistRavenOracle, []string{"flying"}},
		{"Voice of the Provinces", "Creature — Angel", avrVoiceOfTheProvincesOracle, []string{"flying"}},
		{"Goldnight Redeemer", "Creature — Angel", avrGoldnightRedeemerOracle, []string{"flying"}},
		{"Emancipation Angel", "Creature — Angel", avrEmancipationAngelOracle, []string{"flying"}},
		{"Heirs of Stromkirk", "Creature — Vampire", avrHeirsOfStromkirkOracle, []string{"intimidate"}},
		{"Mad Prophet", "Creature — Human Shaman", avrMadProphetOracle, []string{"haste"}},
		{"Marrow Bats", "Creature — Bat Skeleton", avrMarrowBatsOracle, []string{"flying"}},
		{"Wolfir Avenger", "Creature — Wolf Warrior", avrWolfirAvengerOracle, []string{"flash"}},
		{"Searchlight Geist", "Creature — Spirit", avrSearchlightGeistOracle, []string{"flying"}},
		{"Elgaud Shieldmate", "Creature — Human Soldier", avrElgaudShieldmateOracle, []string{game.KeywordSoulbond}},
		{"Trusted Forcemage", "Creature — Human Shaman", avrTrustedForcemageOracle, []string{game.KeywordSoulbond}},
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

// Undying (CR 702.93a): it returns with a +1/+1 counter.
func TestAvr5570B1UndyingReturnsWithACounter(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Butcher Ghoul", "Creature — Zombie", avrButcherGhoulOracle},
		{"Hound of Griselbrand", "Creature — Elemental Dog", avrHoundOfGriselbrandOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			b27Kill(g, id)
			settleProwess(t, g)
			if !g.Battlefield.Contains(id) || countersOn(g, id, game.CounterPlusOne) != 1 {
				t.Fatalf("undying did not return %s with a +1/+1 counter", tc.name)
			}
		})
	}
}

// Defender: Haunted Guardian can't attack.
func TestAvr5570B1HauntedGuardianCantAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guardian := b12Push(g, me.ID, "Haunted Guardian", "Artifact Creature — Construct", avrHauntedGuardianOracle, 2, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(guardian, opp.ID); err == nil {
		t.Error("a creature with defender attacked")
	}
}

// --- combat statics ----------------------------------------------------------

func TestAvr5570B1LatchSeekerCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seeker := b12Push(g, me.ID, "Latch Seeker", "Creature — Spirit", avrLatchSeekerOracle, 3, 1)
	bear := b12Creature(g, opp.ID, "Bear", avrBearType, 2, 2)
	brAttack(t, g, seeker)
	if err := g.DeclareBlocker(bear, seeker); err == nil {
		t.Error("Latch Seeker was blocked")
	}
}

func TestAvr5570B1ScrapskinDrakeBlocksOnlyFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := b12Creature(g, me.ID, "Walker", avrBearType, 2, 2)
	flyer := apaPush(g, me.ID, me.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	drake := b12Push(g, opp.ID, "Scrapskin Drake", "Creature — Zombie Drake", avrScrapskinDrakeOracle, 2, 3)
	brAttack(t, g, walker, flyer)
	if err := g.DeclareBlocker(drake, walker); err == nil {
		t.Error("Scrapskin Drake blocked a creature without flying")
	}
	if err := g.DeclareBlocker(drake, flyer); err != nil {
		t.Errorf("Scrapskin Drake can't block a flyer: %v", err)
	}
}

func TestAvr5570B1WanderingWolfCantBeBlockedBySmallerCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf := b12Push(g, me.ID, "Wandering Wolf", "Creature — Wolf", avrWanderingWolfOracle, 2, 1)
	small := b12Creature(g, opp.ID, "Small", avrBearType, 1, 1)
	big := b12Creature(g, opp.ID, "Big", avrBearType, 2, 2)
	brAttack(t, g, wolf)
	if err := g.DeclareBlocker(small, wolf); err == nil {
		t.Error("a creature with lesser power blocked the Wolf")
	}
	if err := g.DeclareBlocker(big, wolf); err != nil {
		t.Errorf("a creature with equal power can't block the Wolf: %v", err)
	}
}

// --- enters triggers ------------------------------------------------------------

func TestAvr5570B1CathedralSanctifierGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castHoldCreature(t, g, "Cathedral Sanctifier", "Creature — Human Cleric", avrCathedralSanctifierOracle, 1, 1)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestAvr5570B1GryffVanguardDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Gryff Vanguard", "Creature — Human Knight", avrGryffVanguardOracle, 3, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestAvr5570B1VoiceOfTheProvincesMakesAHuman(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Voice of the Provinces", "Creature — Angel", avrVoiceOfTheProvincesOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Human", me.ID); n != 1 {
		t.Errorf("%d Human tokens, want 1", n)
	}
}

func TestAvr5570B1GoldnightRedeemerGainsTwoPerOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Creature(g, me.ID, "Bear", avrBearType, 2, 2)
	b12Creature(g, me.ID, "Second Bear", avrBearType, 2, 2)
	b12Creature(g, opp.ID, "Their Bear", avrBearType, 2, 2)
	life := me.Life
	castHoldCreature(t, g, "Goldnight Redeemer", "Creature — Angel", avrGoldnightRedeemerOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestAvr5570B1MistRavenBouncesATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", avrBearType, 2, 2)
	castHoldCreature(t, g, "Mist Raven", "Creature — Bird", avrMistRavenOracle, 2, 2)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the target creature did not return to its owner's hand")
	}
}

func TestAvr5570B1FerventCatharStopsATargetFromBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	atk := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 4)
	other := pushVanillaCreature(g, opp.ID, "Other Wall", 0, 4)
	castHoldCreature(t, g, "Fervent Cathar", "Creature — Human Knight", avrFerventCatharOracle, 2, 1)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, wall)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, atk); err == nil {
		t.Error("the targeted creature blocked")
	}
	if err := g.DeclareBlocker(other, atk); err != nil {
		t.Errorf("another creature can't block: %v", err)
	}
}

// The return is chosen as the trigger resolves, among every permanent
// its controller controls, the Angel itself included.
func TestAvr5570B1EmancipationAngelReturnsAPermanentYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	angel := castHoldCreature(t, g, "Emancipation Angel", "Creature — Angel", avrEmancipationAngelOracle, 3, 3)
	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("no return pick")
	}
	if !hasID(p.ChooseCards, rock) || !hasID(p.ChooseCards, angel) {
		t.Error("a noncreature permanent and the Angel itself are both offered")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(rock) || !me.Hand.Contains(rock) {
		t.Error("the artifact did not return to hand")
	}
}

// --- dies triggers -----------------------------------------------------------------

func TestAvr5570B1PollutedDeadDestroysATargetLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := b12Push(g, me.ID, "Polluted Dead", "Creature — Zombie", avrPollutedDeadOracle, 3, 3)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	b27Kill(g, dead)
	g.RunStateChecksForTest()
	answerPickTargetIfOpen(t, g, land)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the target land survived")
	}
}

func TestAvr5570B1RotcrownGhoulMillsFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := b12Push(g, me.ID, "Rotcrown Ghoul", "Creature — Zombie", avrRotcrownGhoulOracle, 3, 3)
	before := opp.Graveyard.Size()
	b27Kill(g, ghoul)
	g.RunStateChecksForTest()
	answerPickTargetPlayerIfOpen(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - before; got != 5 {
		t.Errorf("the target player milled %d, want 5", got)
	}
}

// --- attack and damage triggers ---------------------------------------------------------

func TestAvr5570B1RiotRingleaderPumpsYourHumans(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leader := b12Push(g, me.ID, "Riot Ringleader", "Creature — Human Warrior", avrRiotRingleaderOracle, 2, 2)
	human := b12Creature(g, me.ID, "Home Human", "Creature — Human Soldier", 1, 1)
	bear := b12Creature(g, me.ID, "Home Bear", avrBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Human", "Creature — Human Soldier", 1, 1)
	declareAttack(t, g, opp.ID, leader)
	passPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
	}{{leader, 3}, {human, 2}, {bear, 2}, {theirs, 1}} {
		if p, _ := ptOf(t, g, tc.id); p != tc.want {
			t.Errorf("power %d, want %d", p, tc.want)
		}
	}
}

func TestAvr5570B1HeirsOfStromkirkGrowsOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heirs := b12Push(g, me.ID, "Heirs of Stromkirk", "Creature — Vampire", avrHeirsOfStromkirkOracle, 2, 2)
	dealCombatDamage(g, me.ID, heirs, opp.ID)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, heirs); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// --- activated abilities ----------------------------------------------------------------

func TestAvr5570B1AlchemistsApprenticeSacrificesToDraw(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	apprentice := b12Push(g, me.ID, "Alchemist's Apprentice", "Creature — Human Wizard", avrAlchemistsApprenticeOracl, 1, 1)
	hand := me.Hand.Size()
	apaActivate(t, g, me, apprentice, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(apprentice) {
		t.Error("the Apprentice was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestAvr5570B1MadProphetRummages(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	prophet := b12Push(g, me.ID, "Mad Prophet", "Creature — Human Shaman", avrMadProphetOracle, 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	d2dSetHand(me, 2)
	pitched := me.Hand.Cards[0].InstanceID
	lib := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, prophet, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitched}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitched) {
		t.Error("the discarded card is not in the graveyard")
	}
	if me.Hand.Size() != 2 || me.Library.Size() != lib-1 {
		t.Errorf("hand %d, library %d -> %d: a card should have been drawn", me.Hand.Size(), lib, me.Library.Size())
	}
	if !b13Tapped(t, g, prophet) {
		t.Error("the Prophet is not tapped")
	}
}

func TestAvr5570B1MarrowBatsRegeneratesForFourLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bats := b12Push(g, me.ID, "Marrow Bats", "Creature — Bat Skeleton", avrMarrowBatsOracle, 4, 1)
	life := me.Life
	apaActivate(t, g, me, bats, 0, game.ActivateAbilityParams{})
	if me.Life != life-4 {
		t.Errorf("life %d after paying, want %d", me.Life, life-4)
	}
	b27Kill(g, bats)
	if !g.Battlefield.Contains(bats) {
		t.Error("the shielded Bats were destroyed")
	}
}

func TestAvr5570B1WolfirAvengerRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	wolf := b12Push(g, me.ID, "Wolfir Avenger", "Creature — Wolf Warrior", avrWolfirAvengerOracle, 3, 3)
	apaMana(me, "G", "G")
	apaActivate(t, g, me, wolf, 0, game.ActivateAbilityParams{})
	b27Kill(g, wolf)
	if !g.Battlefield.Contains(wolf) {
		t.Error("the shielded Avenger was destroyed")
	}
}

func TestAvr5570B1SelfKeywordRows(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
		mana                            []string
	}{
		{"Moorland Inquisitor", "Creature — Human Soldier", avrMoorlandInquisitorOracle, "first strike", []string{"W", "W", "W"}},
		{"Searchlight Geist", "Creature — Spirit", avrSearchlightGeistOracle, "deathtouch", []string{"B", "B", "B", "B"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Fatalf("%s has %s before activating", tc.name, tc.keyword)
			}
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s did not gain %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestAvr5570B1ScaldingDevilPingsAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	devil := b12Push(g, me.ID, "Scalding Devil", "Creature — Devil", avrScaldingDevilOracle, 1, 1)
	life := opp.Life
	apaMana(me, "R", "R", "R")
	apaActivate(t, g, me, devil, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Errorf("the target player is at %d, want %d", opp.Life, life-1)
	}
	if hmlTapped(t, g, devil) {
		t.Error("the Devil's ability has no tap in its cost")
	}
}

// --- the spells -------------------------------------------------------------------------

func TestAvr5570B1BannersRaisedPumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := b12Creature(g, me.ID, "Mine", avrBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", avrBearType, 2, 2)
	castCatalogSpell(t, g, "Banners Raised", "Instant", avrBannersRaisedOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 3 || tough != 2 {
		t.Errorf("your creature is %d/%d, want 3/2", p, tough)
	}
	if p, _ := ptOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

func TestAvr5570B1DestroyAndGainLife(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, bad string
		life              int
	}{
		{"Natural End", avrNaturalEndOracle, "Creature — Bear", 3},
		{"Cursebreak", avrCursebreakOracle, "Artifact", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me, opp := top100ActiveSeat(g), top100Opponent(g)
			bad := b12Permanent(g, opp.ID, "Wrong Kind", tc.bad)
			if err := castCatalogSpellErr(t, g, tc.name, "Instant", tc.oracle, cardTarget(bad)); err == nil {
				t.Fatalf("%s targeted a %s", tc.name, tc.bad)
			}
			ench := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
			life := me.Life
			castCatalogSpell(t, g, tc.name, "Instant", tc.oracle, cardTarget(ench))
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(ench) {
				t.Error("the enchantment survived")
			}
			if me.Life != life+tc.life {
				t.Errorf("life %d, want %d", me.Life, life+tc.life)
			}
		})
	}
}

func TestAvr5570B1TargetedPumps(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, keyword string
		p, tough              int
	}{
		{"Zealous Strike", avrZealousStrikeOracle, "first strike", 4, 4},
		{"Uncanny Speed", avrUncannySpeedOracle, "haste", 5, 2},
		{"Snare the Skies", avrSnareTheSkiesOracle, "reach", 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			bear := b12Creature(g, me.ID, "Bear", avrBearType, 2, 2)
			castCatalogSpell(t, g, tc.name, "Instant", tc.oracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, bear); p != tc.p || tough != tc.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
			if !hasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
		})
	}
}

// --- soulbond -----------------------------------------------------------------------------

// Each soulbond card pairs with a Bear through its own enters trigger;
// both then have the keyword, and an unpaired creature does not.
func TestAvr5570B1SoulbondKeywordGrants(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
		power, toughness                int
	}{
		{"Elgaud Shieldmate", "Creature — Human Soldier", avrElgaudShieldmateOracle, "hexproof", 2, 3},
		{"Geist Trappers", "Creature — Human Warrior", avrGeistTrappersOracle, "reach", 3, 5},
		{"Hanweir Lancer", "Creature — Human Knight", avrHanweirLancerOracle, "first strike", 2, 2},
		{"Lightning Mauler", "Creature — Human Berserker", avrLightningMaulerOracle, "haste", 2, 1},
		{"Nearheath Pilgrim", "Creature — Human Cleric", avrNearheathPilgrimOracle, "lifelink", 2, 1},
		{"Nightshade Peddler", "Creature — Human Druid", avrNightshadePeddlerOracle, "deathtouch", 1, 1},
		{"Pathbreaker Wurm", "Creature — Wurm", avrPathbreakerWurmOracle, "trample", 6, 4},
		{"Silverblade Paladin", "Creature — Human Knight", avrSilverbladePaladinOracle, "double strike", 2, 2},
		{"Spectral Gateguards", "Creature — Spirit Soldier", avrSpectralGateguardsOracle, "vigilance", 2, 5},
		{"Wingcrafter", "Creature — Human Wizard", avrWingcrafterOracle, "flying", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceTo(t, g, game.StepPrecombatMain)
			bear := b12Creature(g, me.ID, "Bear", avrBearType, 2, 2)
			if hasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Fatalf("the Bear has %s before the pair", tc.keyword)
			}
			id := sbondPairCard(t, g, me, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, bear)
			for _, c := range []uuid.UUID{id, bear} {
				if !hasEffectiveKeyword(t, g, c, tc.keyword) {
					t.Errorf("a paired creature has no %s", tc.keyword)
				}
			}
			loner := b12Creature(g, me.ID, "Loner", avrBearType, 2, 2)
			if hasEffectiveKeyword(t, g, loner, tc.keyword) {
				t.Errorf("an unpaired creature has %s", tc.keyword)
			}
		})
	}
}

func TestAvr5570B1SoulbondPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness, n    int
	}{
		{"Druid's Familiar", "Creature — Bear", avrDruidsFamiliarOracle, 2, 2, 2},
		{"Trusted Forcemage", "Creature — Human Shaman", avrTrustedForcemageOracle, 2, 2, 1},
		{"Wolfir Silverheart", "Creature — Wolf Warrior", avrWolfirSilverheartOracle, 4, 4, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceTo(t, g, game.StepPrecombatMain)
			bear := b12Creature(g, me.ID, "Bear", avrBearType, 2, 2)
			id := sbondPairCard(t, g, me, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, bear)
			if p, tough := ptOf(t, g, bear); p != 2+tc.n || tough != 2+tc.n {
				t.Errorf("the paired Bear is %d/%d, want %d/%d", p, tough, 2+tc.n, 2+tc.n)
			}
			if p, tough := ptOf(t, g, id); p != tc.power+tc.n || tough != tc.toughness+tc.n {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.power+tc.n, tc.toughness+tc.n)
			}
			loner := b12Creature(g, me.ID, "Loner", avrBearType, 2, 2)
			if p, tough := ptOf(t, g, loner); p != 2 || tough != 2 {
				t.Errorf("an unpaired creature is %d/%d, want 2/2", p, tough)
			}
		})
	}
}
