package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// war_5520_b1_cards_test.go — War of the Spark, batch 1 (#5520). Every
// card is registered Full, and every card with rules text is played
// through the engine.

const (
	warB1GoblinAssailantOracle    = "bc654150-b051-40bd-81f8-621abbb17e91"
	warB1IroncladKrovodOracle     = "bd1c9b5f-3b70-457c-8dd8-2420ce1c0d7c"
	warB1LazotepBehemothOracle    = "db145875-2a40-4fe3-b8ee-bf10ae8f7cd7"
	warB1NagaEternalOracle        = "9d6a10c6-5d39-49c6-b50d-8fc5b8ddcd77"
	warB1BanehoundOracle          = "63d846e2-4bef-4911-9508-273c75a4a040"
	warB1CharityExtractorOracle   = "f18a1a99-8363-4edf-9807-2e23fa007ce8"
	warB1KraulStingerOracle       = "8481d91e-0242-4edc-8f4a-ffa47cf2e097"
	warB1SunbladeAngelOracle      = "be37ef9d-955a-417e-bde8-bb9e0bdf3612"
	warB1WardscaleCrocodileOracle = "86f725f6-b48b-4dcc-9a02-e258a226e31c"
	warB1ArlinnsWolfOracle        = "fe4c0f02-c5fd-4cd5-9f80-0936e3875f1c"
	warB1PouncingLynxOracle       = "d8d92f97-d54c-4954-8324-b9c0a9b96ec2"
	warB1RagingKronchOracle       = "c05c2eae-1a9c-4ce7-9c76-22e40781c1a9"
	warB1TeyosLightshieldOracle   = "f1305d49-e543-49e9-9b36-e56f93f63bc3"
	warB1WallOfRunesOracle        = "35dfe1c7-48ab-436f-9a6e-44835ea01cf9"
	warB1PrismiteOracle           = "feb4fcec-6f3a-4643-9ffb-69d3f5953322"
	warB1LeylineProwlerOracle     = "8ca3852a-fa37-46ef-a74d-3c7487ae50fc"
	warB1CentaurNurturerOracle    = "f60e831b-1404-437a-a587-ed4f4b596084"
	warB1BulwarkGiantOracle       = "f606bd15-b380-447f-b147-72f811266593"
	warB1EliteGuardmageOracle     = "92dfeeb2-1117-422b-87ea-08a589f1134b"
	warB1GuildpactInformantOracle = "330f97ee-6bb2-4a4b-a8bc-411659dc4260"
	warB1MartyrForTheCauseOracle  = "b19cbb92-bdf9-4b34-bd16-db5a594f51b7"
	warB1TibaltsRagerOracle       = "0d7d8f4f-abf3-408b-a289-1dd6a3ebdbf6"
	warB1GoblinAssaultTeamOracle  = "d464b54c-3b02-4c4e-99ef-1ad75f336c65"
	warB1IronBullyOracle          = "7d78155c-c63f-48a3-86e5-12b80400d8de"
	warB1BloomHulkOracle          = "c99896a8-8b58-4946-8f2d-b4e2562f278a"
	warB1HuatlisRaptorOracle      = "999ece12-0f1f-4b20-a600-b7f1a584bbb4"
	warB1KiorasDambreakerOracle   = "795b903b-5260-4c06-b26f-a89cecdfa1f8"
	warB1ShriekdiverOracle        = "f51e1660-e1c8-4122-8c43-321c820a1fc1"
	warB1VampireOpportunistOracle = "6f525ccd-4776-492c-a1f9-8bce5740fb28"
	warB1AvenEternalOracle        = "ba7aad10-6897-4e85-a976-1bf9282de5de"
	warB1LazotepReaverOracle      = "85643f9b-49ea-434b-9f27-c1431064345e"
	warB1InvadingManticoreOracle  = "6e4bf4ee-317e-4c31-b121-6efbc6ae4fcf"
	warB1DreadhordeTwinsOracle    = "4fbb8e03-22f1-4171-88e9-877244e25613"
	warB1VizierOfTheScorpOracle   = "61897e1b-dea0-44c2-9e38-6e2627647e17"
	warB1HeraldOracle             = "9cdbe247-9dd3-4894-8e65-43d879c19321"
	warB1GrimInitiateOracle       = "96264dde-c0b6-4912-afc4-18757f938a4d"
	warB1RelentlessAdvanceOracle  = "f54573d1-9950-46c3-b8f1-cd6f081d62ad"
	warB1CallousDismissalOracle   = "9219b7d5-642d-4470-a7c4-f3e56bfae8c7"
	warB1CrushDissentOracle       = "3ece897d-5b52-4fca-9f97-d7e864134c0b"
	warB1HonorGodPharaohOracle    = "5ce7780d-8a32-44c2-b779-aae732f5b05d"
	warB1SarkhansCatharsisOracle  = "787935ab-26a6-4b4a-9c14-e106160e951d"
	warB1SolarBlazeOracle         = "424f3d53-47e7-4212-8267-53279ea8900a"
	warB1ContentiousPlanOracle    = "2fc6de19-9d04-422b-90bc-eee2f9f8c84c"
	warB1UnlikelyAidOracle        = "795e0ffb-fcd9-4cb9-ad18-8a0217229bfe"
	warB1SamutsSprintOracle       = "7d07821e-d34b-4441-a137-836ffce7127b"
	warB1TamiyosEpiphanyOracle    = "48db03c4-c847-4901-8ecb-f480ede94a37"
	warB1SamutOracle              = "80405d7a-f533-44bd-ac85-890476cf2b78"
	warB1AngrathOracle            = "de08f5d4-8868-4d0f-8c0e-d0d3a2105581"
	warB1TeyoOracle               = "cda93cf5-da51-4117-b461-f4448aa2d16c"
	warB1KioraOracle              = "1efe485e-6874-4f4c-b981-e10f41670865"

	warB1BearType = "Creature — Bear"
)

// warB1TapForAnyColor taps id for its any-color mana ability, answering the
// color prompt with green, as Dawnhart Rejuvenator's test does.
func warB1TapForAnyColor(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) error {
	t.Helper()
	if err := g.ActivateManaAbility(p.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		return err
	}
	if pick := riderLatestManaPick(g, p.ID); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, p.ID, "G"); err != nil {
			t.Fatalf("color: %v", err)
		}
	}
	return nil
}

// warB1SettleWithPick settles the stack, answering one pick_target prompt
// with target if one opens.
func warB1SettleWithPick(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	passPriorityAroundTable(t, g)
	eoeABPickIfAsked(t, g, chooser, target)
	passPriorityAroundTable(t, g)
}

// --- every card is in the catalog ----------------------------------------

func TestWar5520B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		warB1GoblinAssailantOracle, warB1IroncladKrovodOracle, warB1LazotepBehemothOracle, warB1NagaEternalOracle,
		warB1BanehoundOracle, warB1CharityExtractorOracle, warB1KraulStingerOracle, warB1SunbladeAngelOracle,
		warB1WardscaleCrocodileOracle, warB1ArlinnsWolfOracle, warB1PouncingLynxOracle, warB1RagingKronchOracle,
		warB1TeyosLightshieldOracle, warB1WallOfRunesOracle, warB1PrismiteOracle, warB1LeylineProwlerOracle,
		warB1CentaurNurturerOracle, warB1BulwarkGiantOracle, warB1EliteGuardmageOracle, warB1GuildpactInformantOracle,
		warB1MartyrForTheCauseOracle, warB1TibaltsRagerOracle, warB1GoblinAssaultTeamOracle, warB1IronBullyOracle,
		warB1BloomHulkOracle, warB1HuatlisRaptorOracle, warB1KiorasDambreakerOracle, warB1ShriekdiverOracle,
		warB1VampireOpportunistOracle, warB1AvenEternalOracle, warB1LazotepReaverOracle, warB1InvadingManticoreOracle,
		warB1DreadhordeTwinsOracle, warB1VizierOfTheScorpOracle, warB1HeraldOracle, warB1GrimInitiateOracle,
		warB1RelentlessAdvanceOracle, warB1CallousDismissalOracle, warB1CrushDissentOracle, warB1HonorGodPharaohOracle,
		warB1SarkhansCatharsisOracle, warB1SolarBlazeOracle, warB1ContentiousPlanOracle, warB1UnlikelyAidOracle,
		warB1SamutsSprintOracle, warB1TamiyosEpiphanyOracle, warB1SamutOracle, warB1AngrathOracle,
		warB1TeyoOracle, warB1KioraOracle,
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

// --- printed keywords ------------------------------------------------------

func TestWar5520B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, oracle string
		keywords     []string
	}{
		{"Banehound", warB1BanehoundOracle, []string{"lifelink", "haste"}},
		{"Charity Extractor", warB1CharityExtractorOracle, []string{"lifelink"}},
		{"Kraul Stinger", warB1KraulStingerOracle, []string{"deathtouch"}},
		{"Sunblade Angel", warB1SunbladeAngelOracle, []string{"flying", "first strike", "vigilance", "lifelink"}},
		{"Wardscale Crocodile", warB1WardscaleCrocodileOracle, []string{"hexproof"}},
		{"Leyline Prowler", warB1LeylineProwlerOracle, []string{"deathtouch", "lifelink"}},
		{"Wall of Runes", warB1WallOfRunesOracle, []string{"defender"}},
		{"Elite Guardmage", warB1EliteGuardmageOracle, []string{"flying"}},
		{"Guildpact Informant", warB1GuildpactInformantOracle, []string{"flying"}},
		{"Goblin Assault Team", warB1GoblinAssaultTeamOracle, []string{"haste"}},
		{"Iron Bully", warB1IronBullyOracle, []string{"menace"}},
		{"Huatli's Raptor", warB1HuatlisRaptorOracle, []string{"vigilance"}},
		{"Shriekdiver", warB1ShriekdiverOracle, []string{"flying"}},
		{"Aven Eternal", warB1AvenEternalOracle, []string{"flying"}},
		{"Grim Initiate", warB1GrimInitiateOracle, []string{"first strike"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, "Creature — Test", tc.oracle, 1, 1)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// The four vanilla bodies resolve as plain creatures.
func TestWar5520B1VanillaCreaturesResolve(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Goblin Assailant", "Creature — Goblin Warrior", warB1GoblinAssailantOracle},
		{"Ironclad Krovod", "Creature — Beast", warB1IroncladKrovodOracle},
		{"Lazotep Behemoth", "Creature — Zombie Hippo", warB1LazotepBehemothOracle},
		{"Naga Eternal", "Creature — Zombie Snake", warB1NagaEternalOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if !g.Battlefield.Contains(id) {
				t.Errorf("%s did not resolve", tc.name)
			}
		})
	}
}

// --- combat statics ----------------------------------------------------------

func TestWar5520B1ArlinnsWolfEvadesSmallBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf := b12Push(g, me.ID, "Arlinn's Wolf", "Creature — Wolf", warB1ArlinnsWolfOracle, 3, 2)
	small := b12Creature(g, opp.ID, "Small", warB1BearType, 2, 2)
	big := b12Creature(g, opp.ID, "Big", warB1BearType, 3, 3)
	brAttack(t, g, wolf)
	if err := g.DeclareBlocker(small, wolf); err == nil {
		t.Error("a creature with power 2 blocked Arlinn's Wolf")
	}
	if err := g.DeclareBlocker(big, wolf); err != nil {
		t.Errorf("a creature with power 3 can't block Arlinn's Wolf: %v", err)
	}
}

func TestWar5520B1PouncingLynxHasFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, other := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	mine := b12Push(g, me.ID, "Pouncing Lynx", "Creature — Cat", warB1PouncingLynxOracle, 2, 1)
	theirs := b12Push(g, other.ID, "Pouncing Lynx", "Creature — Cat", warB1PouncingLynxOracle, 2, 1)
	if !hasEffectiveKeyword(t, g, mine, "first strike") {
		t.Error("the active player's Lynx has no first strike")
	}
	if hasEffectiveKeyword(t, g, theirs, "first strike") {
		t.Error("a Lynx has first strike on another player's turn")
	}
}

func TestWar5520B1RagingKronchCantAttackAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kronch := b12Push(g, me.ID, "Raging Kronch", "Creature — Beast", warB1RagingKronchOracle, 4, 3)
	bear := b12Creature(g, me.ID, "Bear", warB1BearType, 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(kronch, opp.ID); err != nil {
		t.Fatalf("declaring the Kronch first was refused: %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("passing with the Kronch alone: %v, want ErrAttackLimit", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("the companion was refused: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("passing with two attackers: %v", err)
	}
}

// --- mana abilities -----------------------------------------------------------

func TestWar5520B1TapForAnyColor(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Leyline Prowler", "Creature — Nightmare Beast", warB1LeylineProwlerOracle},
		{"Centaur Nurturer", "Creature — Centaur Druid", warB1CentaurNurturerOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 3)
			if err := warB1TapForAnyColor(t, g, me, id); err != nil {
				t.Fatalf("mana ability: %v", err)
			}
			if me.ManaPool.Total() != 1 {
				t.Errorf("pool %d, want 1", me.ManaPool.Total())
			}
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
				t.Error("a tapped creature made mana again")
			}
		})
	}
}

func TestWar5520B1PrismiteFiltersTwoIntoAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Prismite", "Artifact Creature — Golem", warB1PrismiteOracle, 2, 1)
	samiGiveColorless(me, 2)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v after paying {2} for one red mana, want [R]", got)
	}
}

// --- enters triggers ------------------------------------------------------------

func TestWar5520B1EntersAndGainsLife(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		life, cards            int
	}{
		{"Centaur Nurturer", "Creature — Centaur Druid", warB1CentaurNurturerOracle, 3, 0},
		{"Bulwark Giant", "Creature — Giant Soldier", warB1BulwarkGiantOracle, 5, 0},
		{"Elite Guardmage", "Creature — Human Wizard", warB1EliteGuardmageOracle, 3, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			life, hand := me.Life, me.Hand.Size()
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 3)
			passPriorityAroundTable(t, g)
			if me.Life != life+tc.life {
				t.Errorf("life %d, want %d", me.Life, life+tc.life)
			}
			if got := me.Hand.Size(); got != hand+tc.cards {
				t.Errorf("hand %d, want %d", got, hand+tc.cards)
			}
		})
	}
}

func TestWar5520B1WallOfRunesScriesOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Wall of Runes", "Creature — Wall", warB1WallOfRunesOracle, 0, 4)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scrying %d cards, want 1", len(c.ScryCards))
	}
}

func TestWar5520B1EntersAndPutsACounterOnATarget(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Teyo's Lightshield", "Creature — Illusion", warB1TeyosLightshieldOracle},
		{"Iron Bully", "Artifact Creature — Golem", warB1IronBullyOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			ally := b12Creature(g, me.ID, "Ally", warB1BearType, 2, 2)
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 1, 1)
			warB1SettleWithPick(t, g, me.ID, ally)
			if got := countersOn(g, ally, game.CounterPlusOne); got != 1 {
				t.Errorf("ally counters = %d, want 1", got)
			}
		})
	}
}

// Teyo's Lightshield targets only a creature you control.
func TestWar5520B1TeyosLightshieldTargetsOnlyYourCreature(t *testing.T) {
	spec, ok := Lookup(warB1TeyosLightshieldOracle)
	if !ok || len(spec.Triggered) != 1 || spec.Triggered[0].Targets == nil {
		t.Fatal("Teyo's Lightshield has no targeted enters trigger")
	}
	if got := spec.Triggered[0].Targets.Label; got != "target creature you control" {
		t.Errorf("target label %q", got)
	}
}

func TestWar5520B1EntersAndProliferates(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Bloom Hulk", "Creature — Plant Elemental", warB1BloomHulkOracle},
		{"Huatli's Raptor", "Creature — Dinosaur", warB1HuatlisRaptorOracle},
		{"Kiora's Dambreaker", "Creature — Leviathan", warB1KiorasDambreakerOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			grown := pushCounterCreature(g, me.ID, "Grown", game.CounterPlusOne, 1)
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 4, 4)
			settleAnsweringProliferate(t, g)
			if got := counterCount(g, grown, game.CounterPlusOne); got != 2 {
				t.Errorf("counters after proliferate = %d, want 2", got)
			}
		})
	}
}

// --- other triggers ---------------------------------------------------------------

func TestWar5520B1GuildpactInformantProliferatesOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	informant := b12Push(g, me.ID, "Guildpact Informant", "Creature — Faerie Rogue", warB1GuildpactInformantOracle, 1, 1)
	grown := pushCounterCreature(g, me.ID, "Grown", game.CounterPlusOne, 1)
	dealNoncombatDamage(g, me.ID, informant, opp.ID)
	settleAnsweringProliferate(t, g)
	if got := counterCount(g, grown, game.CounterPlusOne); got != 1 {
		t.Fatalf("noncombat damage proliferated: counters %d", got)
	}
	dealCombatDamage(g, me.ID, informant, opp.ID)
	settleAnsweringProliferate(t, g)
	if got := counterCount(g, grown, game.CounterPlusOne); got != 2 {
		t.Errorf("counters after combat damage = %d, want 2", got)
	}
}

func TestWar5520B1MartyrForTheCauseProliferatesWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	martyr := b12Push(g, me.ID, "Martyr for the Cause", "Creature — Human Soldier", warB1MartyrForTheCauseOracle, 2, 2)
	grown := pushCounterCreature(g, me.ID, "Grown", game.CounterPlusOne, 1)
	b27Kill(g, martyr)
	settleAnsweringProliferate(t, g)
	if got := counterCount(g, grown, game.CounterPlusOne); got != 2 {
		t.Errorf("counters after the Martyr died = %d, want 2", got)
	}
}

func TestWar5520B1TibaltsRagerPingsWhenItDiesAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rager := b12Push(g, me.ID, "Tibalt's Rager", "Creature — Devil", warB1TibaltsRagerOracle, 1, 2)
	b16Activate(t, g, me.ID, rager, 0, game.ActivateAbilityParams{})
	if got := apaLive(g, rager).CurrentPower(); got != 3 {
		t.Errorf("power %d after {1}{R}, want 3", got)
	}
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	b27Kill(g, rager)
	warB1SettleWithPick(t, g, me.ID, ogre)
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 1 {
		t.Errorf("the target has %d damage, want 1", got)
	}
}

func TestWar5520B1GoblinAssaultTeamCountersAnAllyWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	team := b12Push(g, me.ID, "Goblin Assault Team", "Creature — Goblin Warrior", warB1GoblinAssaultTeamOracle, 4, 1)
	ally := b12Creature(g, me.ID, "Ally", warB1BearType, 2, 2)
	b27Kill(g, team)
	warB1SettleWithPick(t, g, me.ID, ally)
	if got := countersOn(g, ally, game.CounterPlusOne); got != 1 {
		t.Errorf("ally counters = %d, want 1", got)
	}
}

// --- activated abilities ------------------------------------------------------------

func TestWar5520B1ShriekdiverGainsHaste(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	diver := b12Push(g, me.ID, "Shriekdiver", "Creature — Zombie Bird Warrior", warB1ShriekdiverOracle, 2, 1)
	if hasEffectiveKeyword(t, g, diver, "haste") {
		t.Fatal("haste before the ability")
	}
	b16Activate(t, g, me.ID, diver, 0, game.ActivateAbilityParams{})
	if !hasEffectiveKeyword(t, g, diver, "haste") {
		t.Error("no haste after {1}")
	}
}

func TestWar5520B1VampireOpportunistDrainsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0] // lifeOfOpponents reads seats 1..3
	vamp := b12Push(g, me.ID, "Vampire Opportunist", "Creature — Vampire", warB1VampireOpportunistOracle, 2, 1)
	life := me.Life
	opps := lifeOfOpponents(g)
	b16Activate(t, g, me.ID, vamp, 0, game.ActivateAbilityParams{})
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	for i, l := range lifeOfOpponents(g) {
		if l != opps[i]-2 {
			t.Errorf("opponent %d life %d, want %d", i, l, opps[i]-2)
		}
	}
}

// --- amass -----------------------------------------------------------------------

func TestWar5520B1EntersAndAmasses(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
		grant                  string
	}{
		{"Aven Eternal", "Creature — Zombie Bird Warrior", warB1AvenEternalOracle, 1, ""},
		{"Lazotep Reaver", "Creature — Zombie Beast", warB1LazotepReaverOracle, 1, ""},
		{"Invading Manticore", "Creature — Zombie Manticore", warB1InvadingManticoreOracle, 2, ""},
		{"Dreadhorde Twins", "Creature — Zombie Jackal Warrior", warB1DreadhordeTwinsOracle, 2, "trample"},
		{"Vizier of the Scorpion", "Creature — Zombie Wizard", warB1VizierOfTheScorpOracle, 1, "deathtouch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 1, 1)
			passPriorityAroundTable(t, g)
			army := armyOf(t, g, me.ID)
			if got := army.Counters[game.CounterPlusOne]; got != tc.n {
				t.Errorf("Army counters = %d, want %d", got, tc.n)
			}
			if !army.HasSubtype("Zombie") {
				t.Error("the Army is not a Zombie")
			}
			if tc.grant != "" && !game.HasKeyword(army, tc.grant) {
				t.Errorf("the Zombie Army token has no %s", tc.grant)
			}
		})
	}
}

// The grants reach Zombie tokens only, not a nontoken Zombie.
func TestWar5520B1ZombieTokenGrantsSkipNontokenZombies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Dreadhorde Twins", "Creature — Zombie Jackal Warrior", warB1DreadhordeTwinsOracle, 2, 2)
	b12Push(g, me.ID, "Vizier of the Scorpion", "Creature — Zombie Wizard", warB1VizierOfTheScorpOracle, 1, 1)
	zombie := b12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
	if hasEffectiveKeyword(t, g, zombie, "trample") || hasEffectiveKeyword(t, g, zombie, "deathtouch") {
		t.Error("a nontoken Zombie got a Zombie-token grant")
	}
}

func TestWar5520B1DiesAndAmasses(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Herald of the Dreadhorde", "Creature — Zombie Warrior", warB1HeraldOracle, 2},
		{"Grim Initiate", "Creature — Zombie Warrior", warB1GrimInitiateOracle, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			hobDestroy(t, g, id)
			if got := hobArmyCounters(t, g, me.ID); got != tc.n {
				t.Errorf("Army counters = %d, want %d", got, tc.n)
			}
		})
	}
}

func TestWar5520B1RelentlessAdvanceAmassesThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Relentless Advance", "Sorcery", warB1RelentlessAdvanceOracle, nil)
	passPriorityAroundTable(t, g)
	if got := hobArmyCounters(t, g, me.ID); got != 3 {
		t.Errorf("Army counters = %d, want 3", got)
	}
}

func TestWar5520B1CallousDismissalBouncesAndAmasses(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Theirs", warB1BearType, 2, 2)
	castCatalogSpell(t, g, "Callous Dismissal", "Sorcery", warB1CallousDismissalOracle, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("the target did not return to its owner's hand")
	}
	if got := hobArmyCounters(t, g, me.ID); got != 1 {
		t.Errorf("Army counters = %d, want 1", got)
	}
}

func TestWar5520B1CallousDismissalRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := castCatalogSpellErr(t, g, "Callous Dismissal", "Sorcery", warB1CallousDismissalOracle, cardTarget(land)); err == nil {
		t.Error("Callous Dismissal targeted a land")
	}
}

func TestWar5520B1CrushDissentTaxesThenAmasses(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Bear", warB1BearType, "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castFromSeat(t, g, them, "Crush Dissent", "Instant", warB1CrushDissentOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passUntilTaxed(t, g, me.ID)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) || g.Battlefield.Contains(victim) {
		t.Error("the declined spell was not countered")
	}
	if got := hobArmyCounters(t, g, them.ID); got != 2 {
		t.Errorf("Army counters = %d, want 2", got)
	}
}

func TestWar5520B1HonorTheGodPharaohDiscardsDrawsAndAmasses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	before := me.Library.Size()
	_, paid := castWithDiscard(t, g, "Honor the God-Pharaoh", warB1HonorGodPharaohOracle, "Sorcery", 1)
	if !me.Graveyard.Contains(paid[0]) {
		t.Fatal("the discard was not paid at announce")
	}
	passPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if got := hobArmyCounters(t, g, me.ID); got != 1 {
		t.Errorf("Army counters = %d, want 1", got)
	}
}

// --- other spells -----------------------------------------------------------------

func TestWar5520B1SarkhansCatharsisDealsFiveToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	castCatalogSpell(t, g, "Sarkhan's Catharsis", "Instant", warB1SarkhansCatharsisOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("life %d, want %d", opp.Life, life-5)
	}
}

func TestWar5520B1SolarBlazeHasEachCreatureHitItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	glass := b12Creature(g, me.ID, "Glass", warB1BearType, 3, 3)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 2, 5)
	castCatalogSpell(t, g, "Solar Blaze", "Sorcery", warB1SolarBlazeOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(glass) {
		t.Error("the 3/3 survived three damage")
	}
	if got := damageOn(t, g, wall); got != 2 {
		t.Errorf("the 2/5 has %d damage, want 2", got)
	}
}

func TestWar5520B1ContentiousPlanProliferatesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	grown := pushCounterCreature(g, me.ID, "Grown", game.CounterPlusOne, 1)
	advanceToMain(t, g)
	before := me.Library.Size()
	castCatalogSpell(t, g, "Contentious Plan", "Sorcery", warB1ContentiousPlanOracle, nil)
	settleAnsweringProliferate(t, g)
	if got := counterCount(g, grown, game.CounterPlusOne); got != 2 {
		t.Errorf("counters %d, want 2", got)
	}
	if got := before - me.Library.Size(); got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestWar5520B1UnlikelyAidPumpsAndProtects(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", warB1BearType, 2, 2)
	castCatalogSpell(t, g, "Unlikely Aid", "Instant", warB1UnlikelyAidOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if got := apaLive(g, bear).CurrentPower(); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("no indestructible")
	}
}

func TestWar5520B1SamutsSprintPumpsHastesAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", warB1BearType, 2, 2)
	castCatalogSpell(t, g, "Samut's Sprint", "Instant", warB1SamutsSprintOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	live := apaLive(g, bear)
	if live.CurrentPower() != 4 || live.CurrentToughness() != 3 {
		t.Errorf("%d/%d, want 4/3", live.CurrentPower(), live.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("no haste")
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("no scry 1 after the pump")
	}
}

func TestWar5520B1TamiyosEpiphanyScriesFourThenDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	before := me.Library.Size()
	castCatalogSpell(t, g, "Tamiyo's Epiphany", "Sorcery", warB1TamiyosEpiphanyOracle, nil)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 4 {
		t.Errorf("scrying %d cards, want 4", len(c.ScryCards))
	}
	if got := before - me.Library.Size(); got != 0 {
		t.Errorf("drew %d before the scry was answered", got)
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

// --- planeswalkers ---------------------------------------------------------------

func TestWar5520B1SamutGivesHasteAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	samut := pushWalkerForTest(g, me.ID, "Samut, Tyrant Smasher", warB1SamutOracle, 5)
	mine := b12Creature(g, me.ID, "Mine", warB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", warB1BearType, 2, 2)
	if !hasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("your creature has no haste")
	}
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
	b16Activate(t, g, me.ID, samut, 0, game.ActivateAbilityParams{Targets: cardTarget(mine)})
	live := apaLive(g, mine)
	if live.CurrentPower() != 4 || live.CurrentToughness() != 3 {
		t.Errorf("%d/%d after −1, want 4/3", live.CurrentPower(), live.CurrentToughness())
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("no scry 1 after the pump")
	}
	if got := loyaltyOf(g, samut); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

func TestWar5520B1AngrathGivesMenaceAndAmasses(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	angrath := pushWalkerForTest(g, me.ID, "Angrath, Captain of Chaos", warB1AngrathOracle, 5)
	mine := b12Creature(g, me.ID, "Mine", warB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", warB1BearType, 2, 2)
	if !hasEffectiveKeyword(t, g, mine, "menace") {
		t.Error("your creature has no menace")
	}
	if hasEffectiveKeyword(t, g, theirs, "menace") {
		t.Error("an opponent's creature has menace")
	}
	b16Activate(t, g, me.ID, angrath, 0, game.ActivateAbilityParams{})
	if got := hobArmyCounters(t, g, me.ID); got != 2 {
		t.Errorf("Army counters = %d, want 2", got)
	}
	if got := loyaltyOf(g, angrath); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

func TestWar5520B1TeyoGivesYouHexproofAndMakesAWall(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	teyo := pushWalkerForTest(g, me.ID, "Teyo, the Shieldmage", warB1TeyoOracle, 5)
	if !hasPlayerAbility(playerAbilities(g, me), KeywordHexproof) {
		t.Error("Teyo's controller has no hexproof")
	}
	if hasPlayerAbility(playerAbilities(g, opp), KeywordHexproof) {
		t.Error("an opponent has hexproof from your Teyo")
	}
	b16Activate(t, g, me.ID, teyo, 0, game.ActivateAbilityParams{})
	var wall *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Wall" && c.Controller == me.ID && c.IsToken() {
			wall = c
		}
	}
	if wall == nil {
		t.Fatal("no Wall token")
	}
	if wall.Power != 0 || wall.Toughness != 3 || !hasEffectiveKeyword(t, g, wall.InstanceID, "defender") {
		t.Errorf("Wall %d/%d defender=%v, want a 0/3 with defender", wall.Power, wall.Toughness,
			hasEffectiveKeyword(t, g, wall.InstanceID, "defender"))
	}
	if got := loyaltyOf(g, teyo); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

func TestWar5520B1KioraDrawsForBigCreaturesAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	kiora := pushWalkerForTest(g, me.ID, "Kiora, Behemoth Beckoner", warB1KioraOracle, 7)
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Goblin Assailant", "Creature — Goblin Warrior", warB1GoblinAssailantOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("a 2-power creature drew: hand %d, want %d", got, hand)
	}
	castHoldCreature(t, g, "Lazotep Behemoth", "Creature — Zombie Hippo", warB1LazotepBehemothOracle, 5, 4)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("a 5-power creature: hand %d, want %d", got, hand+1)
	}
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	tapForTest(t, g, land)
	b16Activate(t, g, me.ID, kiora, 0, game.ActivateAbilityParams{Targets: cardTarget(land)})
	if hmlTapped(t, g, land) {
		t.Error("the −1 did not untap the land")
	}
	if got := loyaltyOf(g, kiora); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}
