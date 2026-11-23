package testsst

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stx_5502_b1_cards_test.go — the first batch of Strixhaven: School of
// Mages cards (#5502). Every test plays the real catalog entry through
// the engine.

const (
	stxAgelessGuardianOracle       = "468658bc-19d8-464a-ab8e-d36281680273"
	stxSpinedKarokOracle           = "7341a04a-b6d6-4106-b1c3-3483f9d37663"
	stxMolderingKarokOracle        = "5d95db27-53bb-4a20-9fb8-24998b03e767"
	stxNeedlethornDrakeOracle      = "6edacc06-9523-4ccf-a4c2-2ac4af386b8a"
	stxRelicSlothOracle            = "eb9b0655-3ca3-4d12-8665-54ef0c2c4de2"
	stxTwinscrollShamanOracle      = "a2f867cd-410d-4524-8660-38ad142f6b1f"
	stxWaterfallAerialistOracle    = "736026b2-84d4-4452-ae0d-5d1b0af0a400"
	stxOwlinShieldmageOracle       = "50c477c1-8457-40f6-bba9-1a4ba0862178"
	stxCleverLumimancerOracle      = "901eb916-f318-4a30-a43e-8e7f9eef358b"
	stxEagerFirstYearOracle        = "74788d52-dad7-4c10-8567-a2c86ed90a67"
	stxLoreholdPledgemageOracle    = "2c18d692-612c-4d43-b5dd-bd2bff37bbcb"
	stxQuandrixPledgemageOracle    = "8a39809e-219e-4f59-a233-ea28d05c30ef"
	stxWitherbloomApprenticeOracle = "696f554d-0485-48a5-9273-3f6fb7d16a5d"
	stxWitherbloomPledgemageOracle = "f10039ae-2695-4b0a-82b2-82ee79a992c7"
	stxLeoninLightscribeOracle     = "14079b33-cf68-4d16-962c-5c40518c4163"
	stxSilverquillApprenticeOracle = "5ab91144-5eb3-491b-a8be-7e1139d0d6c8"
	stxSpringmaneCervinOracle      = "4e964f31-c764-4af5-8c27-a46fd721fab4"
	stxProfessorOfZoomancyOracle   = "a22b8220-b41d-4b58-aa68-2c65ce37c517"
	stxPillardropRescuerOracle     = "5cf9002d-6c04-420b-8c3b-0a17f674bb09"
	stxBiblioplexAssistantOracle   = "3dadb654-6191-4ad3-8315-46f5613fbc01"
	stxBloodResearcherOracle       = "a465ba74-615a-4261-b4a0-fcb0819f10b3"
	stxSoothsayerAdeptOracle       = "29723d39-aaca-48fd-ba4a-14931e4483e9"
	stxOggyarBattleSeerOracle      = "645005dd-3438-4b88-bb19-561f9679a022"
	stxHallMonitorOracle           = "937a04c2-41be-4ff3-8992-6628a9ee9792"
	stxWormholeSerpentOracle       = "6eb21aca-116e-427b-9e75-8adaee0cc62a"
	stxSpecterOfTheFensOracle      = "048925e1-3bd6-4caa-90b9-e75e890032a5"
	stxCogworkArchivistOracle      = "60674ef7-f8e2-41e8-bab3-0100cfc9fbe9"
	stxUnwillingIngredientOracle   = "40d483fb-0dc7-413e-87f9-8d7cb7883fca"
	stxLeechFanaticOracle          = "17bcd92b-1826-4a36-9048-49e5e65a12fd"
	stxScurridColonyOracle         = "7dd68fd9-f5a7-49ee-aa36-93d646254335"
	stxLetterOfAcceptanceOracle    = "e0cb503e-9814-4280-9c29-4eff3379bc24"
	stxExcavatedWallOracle         = "16e48b95-f030-400c-a2b7-c3f3111cc165"
	stxBeamingDefianceOracle       = "f88eeb04-4b8e-4c76-a8c6-90887b94d518"
	stxChargeThroughOracle         = "7e24b4da-9728-43a0-8af4-f513a3e1104b"
	stxLashOfMaliceOracle          = "fc3d2a9c-08a5-4aea-9fb8-42466b9152a2"
	stxExpelOracle                 = "58776b25-794e-4a87-bc71-a77c025b9b03"
	stxHeatedDebateOracle          = "06c3e85c-17d8-44e1-a0f8-07ef3f1624cb"
	stxIntroToProphecyOracle       = "9ae1fded-2e77-498a-98e7-4a4908447175"
	stxSpiritSummoningOracle       = "a4825093-ea29-47d9-b52c-1c7735f1909e"
	stxInklingSummoningOracle      = "999a9310-d145-4c1a-8518-9d96602ca672"
	stxElementalSummoningOracle    = "f9e668f9-21af-4fd8-9a37-6505ab2234e3"
	stxPestSummoningOracle         = "f6e0cb77-4115-4948-8e1e-fdb774b1b254"
	stxMascotExhibitionOracle      = "a1ea2cba-f184-4a27-9971-3a56b0a0e390"
	stxStartFromScratchOracle      = "d8e255ec-7d66-43f6-b5f0-712f5957687f"
	stxTangletrapOracle            = "75effdde-0427-4e5a-bd6d-c3118f301520"
	stxSuddenBreakthroughOracle    = "f5576cd4-6b21-4b29-902c-9c1fbdd7b2be"
	stxExpandedAnatomyOracle       = "6bc5dfbb-98ec-4ff1-8ba2-e2594c53e6d9"
	stxEssenceInfusionOracle       = "60f412bb-27e9-4883-8862-96eae96dfee3"

	stxBearType   = "Creature — Bear"
	stxLessonType = "Sorcery — Lesson"
)

// stxB1Settle passes priority until the stack is empty, answering every
// pick_target prompt addressed to chooser with target.
func stxB1Settle(t *testing.T, g *game.Game, chooser uuid.UUID, target game.TargetRef) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, target); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				return
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("stack did not settle")
}

// stxB1Bolt casts Lightning Bolt from the active seat at an opponent
// and resolves it, with any trigger it causes. It returns the opponent.
func stxB1Bolt(t *testing.T, g *game.Game) *game.Player {
	t.Helper()
	opp := effectstest.Top100Opponent(g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	return opp
}

// stxB1Tapped seeds a tapped non-catalog creature.
func stxB1Tapped(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: stxBearType,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner, Tapped: true,
	})
}

// --- every card is in the catalog ----------------------------------------

func TestStx5502B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		stxAgelessGuardianOracle, stxSpinedKarokOracle, stxMolderingKarokOracle, stxNeedlethornDrakeOracle,
		stxRelicSlothOracle, stxTwinscrollShamanOracle, stxWaterfallAerialistOracle, stxOwlinShieldmageOracle,
		stxCleverLumimancerOracle, stxEagerFirstYearOracle, stxLoreholdPledgemageOracle, stxQuandrixPledgemageOracle,
		stxWitherbloomApprenticeOracle, stxWitherbloomPledgemageOracle, stxLeoninLightscribeOracle, stxSilverquillApprenticeOracle,
		stxSpringmaneCervinOracle, stxProfessorOfZoomancyOracle, stxPillardropRescuerOracle, stxBiblioplexAssistantOracle,
		stxBloodResearcherOracle, stxSoothsayerAdeptOracle, stxOggyarBattleSeerOracle, stxHallMonitorOracle,
		stxWormholeSerpentOracle, stxSpecterOfTheFensOracle, stxCogworkArchivistOracle, stxUnwillingIngredientOracle,
		stxLeechFanaticOracle, stxScurridColonyOracle, stxLetterOfAcceptanceOracle, stxExcavatedWallOracle,
		stxBeamingDefianceOracle, stxChargeThroughOracle, stxLashOfMaliceOracle, stxExpelOracle,
		stxHeatedDebateOracle, stxIntroToProphecyOracle, stxSpiritSummoningOracle, stxInklingSummoningOracle,
		stxElementalSummoningOracle, stxPestSummoningOracle, stxMascotExhibitionOracle, stxStartFromScratchOracle,
		stxTangletrapOracle, stxSuddenBreakthroughOracle, stxExpandedAnatomyOracle, stxEssenceInfusionOracle,
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

func TestStx5502B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Ageless Guardian", "Creature — Spirit Soldier", stxAgelessGuardianOracle, nil},
		{"Spined Karok", "Creature — Crocodile", stxSpinedKarokOracle, nil},
		{"Moldering Karok", "Creature — Zombie Crocodile", stxMolderingKarokOracle, []string{"trample", "lifelink"}},
		{"Needlethorn Drake", "Creature — Drake", stxNeedlethornDrakeOracle, []string{"flying", "deathtouch"}},
		{"Relic Sloth", "Creature — Sloth Beast", stxRelicSlothOracle, []string{"vigilance", "menace"}},
		{"Twinscroll Shaman", "Creature — Dwarf Shaman", stxTwinscrollShamanOracle, []string{"double strike"}},
		{"Waterfall Aerialist", "Creature — Djinn Wizard", stxWaterfallAerialistOracle, []string{"flying"}},
		{"Owlin Shieldmage", "Creature — Bird Warlock", stxOwlinShieldmageOracle, []string{"flying"}},
		{"Lorehold Pledgemage", "Creature — Kor Shaman", stxLoreholdPledgemageOracle, []string{"first strike"}},
		{"Pillardrop Rescuer", "Creature — Spirit Cleric", stxPillardropRescuerOracle, []string{"flying"}},
		{"Biblioplex Assistant", "Artifact Creature — Gargoyle", stxBiblioplexAssistantOracle, []string{"flying"}},
		{"Blood Researcher", "Creature — Vampire Druid", stxBloodResearcherOracle, []string{"menace"}},
		{"Oggyar Battle-Seer", "Creature — Ogre Shaman", stxOggyarBattleSeerOracle, []string{"haste"}},
		{"Hall Monitor", "Creature — Lizard Shaman", stxHallMonitorOracle, []string{"haste"}},
		{"Specter of the Fens", "Creature — Specter", stxSpecterOfTheFensOracle, []string{"flying"}},
		{"Cogwork Archivist", "Artifact Creature — Construct", stxCogworkArchivistOracle, []string{"reach"}},
		{"Unwilling Ingredient", "Creature — Frog", stxUnwillingIngredientOracle, []string{"menace"}},
		{"Scurrid Colony", "Creature — Squirrel", stxScurridColonyOracle, []string{"reach"}},
		{"Excavated Wall", "Artifact Creature — Wall", stxExcavatedWallOracle, []string{"defender"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s is not on the battlefield", tc.name)
			}
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- ward -------------------------------------------------------------------

// Waterfall Aerialist's ward {2}: an opponent's spell that targets it is
// countered unless they pay, and declining keeps the Aerialist.
func TestStx5502B1WaterfallAerialistWardCountersAnUnpaidSpell(t *testing.T) {
	g, me, opp, aerialist := effectstest.WardTable(t, "Waterfall Aerialist", "Creature — Djinn Wizard", stxWaterfallAerialistOracle)
	blade := effectstest.CastAtWardedCreature(t, g, opp, aerialist)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("the Aerialist's controller was asked to pay")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aerialist) {
		t.Error("a declined ward did not counter the spell")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
}

// Owlin Shieldmage's ward asks the caster for 3 life.
func TestStx5502B1OwlinShieldmageWardAsksForThreeLife(t *testing.T) {
	g, _, opp, owl := effectstest.WardTable(t, "Owlin Shieldmage", "Creature — Bird Warlock", stxOwlinShieldmageOracle)
	blade := effectstest.CastAtWardedCreature(t, g, opp, owl)
	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if prompt.LifeCost != 3 {
		t.Errorf("prompt LifeCost = %d, want 3", prompt.LifeCost)
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(owl) || !opp.Graveyard.Contains(blade) {
		t.Error("a declined ward did not counter the spell")
	}
}

// --- magecraft ----------------------------------------------------------------

func TestStx5502B1MagecraftPumpsThisCreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough, wantP, wantT int
	}{
		{"Clever Lumimancer", "Creature — Human Wizard", stxCleverLumimancerOracle, 0, 1, 2, 3},
		{"Eager First-Year", "Creature — Human Wizard", stxEagerFirstYearOracle, 2, 2, 3, 2},
		{"Lorehold Pledgemage", "Creature — Kor Shaman", stxLoreholdPledgemageOracle, 2, 2, 3, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := effectstest.Top100ActiveSeat(g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			stxB1Bolt(t, g)
			if p, tough := effectstest.PtOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%s is %d/%d after an instant, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestStx5502B1QuandrixPledgemageGetsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	id := effectstest.B12Push(g, me.ID, "Quandrix Pledgemage", "Creature — Merfolk Druid", stxQuandrixPledgemageOracle, 2, 2)
	stxB1Bolt(t, g)
	if got := effectstest.PlusOneCounters(g, id); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestStx5502B1WitherbloomApprenticeDrainsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.B12Push(g, me.ID, "Witherbloom Apprentice", "Creature — Human Druid", stxWitherbloomApprenticeOracle, 2, 2)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	opp := stxB1Bolt(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 1
		switch s.ID {
		case me.ID:
			want = lives[s.ID] + 1
		case opp.ID:
			want = lives[s.ID] - 4
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
}

func TestStx5502B1WitherbloomPledgemageGainsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.B12Push(g, me.ID, "Witherbloom Pledgemage", "Creature — Treefolk Warlock", stxWitherbloomPledgemageOracle, 5, 5)
	life := me.Life
	stxB1Bolt(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func TestStx5502B1LeoninLightscribePumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.B12Push(g, me.ID, "Leonin Lightscribe", "Creature — Cat Cleric", stxLeoninLightscribeOracle, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", stxBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", stxBearType, 2, 2)
	stxB1Bolt(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestStx5502B1SilverquillApprenticePumpsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.B12Push(g, me.ID, "Silverquill Apprentice", "Creature — Human Warlock", stxSilverquillApprenticeOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", stxBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	stxB1Settle(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the target is %d/%d, want 3/2", p, tough)
	}
}

// --- enters and life-gain triggers ----------------------------------------------

func TestStx5502B1SpringmaneCervinFeedsBloodResearcher(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	researcher := effectstest.B12Push(g, me.ID, "Blood Researcher", "Creature — Vampire Druid", stxBloodResearcherOracle, 2, 2)
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Springmane Cervin", "Creature — Elk", stxSpringmaneCervinOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if got := effectstest.PlusOneCounters(g, researcher); got != 1 {
		t.Errorf("Blood Researcher has %d +1/+1 counters, want 1", got)
	}
}

func TestStx5502B1ProfessorOfZoomancyMakesAPest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.CastHoldCreature(t, g, "Professor of Zoomancy", "Creature — Bear Druid", stxProfessorOfZoomancyOracle, 4, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Pest"); n != 1 {
		t.Errorf("%d Pests, want 1", n)
	}
}

func TestStx5502B1PillardropRescuerReturnsACheapCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.AdvanceToMain(t, g)
	dead := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", stxBearType, "", 2, 2)
	dear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dear, Name: "Dear", TypeLine: stxBearType, ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
	effectstest.CastHoldCreature(t, g, "Pillardrop Rescuer", "Creature — Spirit Cleric", stxPillardropRescuerOracle, 2, 2)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil && effectstest.HasID(pick.PickTargetCards, dear) {
		t.Error("a mana value 4 creature card was offered")
	}
	stxB1Settle(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dead})
	if !me.Hand.Contains(dead) || me.Graveyard.Contains(dead) {
		t.Error("the creature card did not return to your hand")
	}
	if !me.Graveyard.Contains(dear) {
		t.Error("the mana value 4 card left the graveyard")
	}
}

func TestStx5502B1BiblioplexAssistantPutsAnInstantOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.PushCatalogGraveyardCard(me, "Old Instant", "Instant", "", 0, 0)
	effectstest.CastHoldCreature(t, g, "Biblioplex Assistant", "Artifact Creature — Gargoyle", stxBiblioplexAssistantOracle, 2, 1)
	stxB1Settle(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: spell})
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != spell {
		t.Errorf("the top of the library is %q (%v), want the instant", top.Name, err)
	}
}

// --- activated abilities ----------------------------------------------------------

func TestStx5502B1SoothsayerAdeptLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	adept := effectstest.B12Push(g, me.ID, "Soothsayer Adept", "Creature — Merfolk Wizard", stxSoothsayerAdeptOracle, 1, 3)
	hand := me.Hand.Size()
	effectstest.ApaMana(me, "C", "U")
	effectstest.ApaActivate(t, g, me, adept, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d before the discard, want %d", got, hand+1)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestStx5502B1OggyarBattleSeerScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	seer := effectstest.B12Push(g, me.ID, "Oggyar Battle-Seer", "Creature — Ogre Shaman", stxOggyarBattleSeerOracle, 3, 4)
	effectstest.ApaActivate(t, g, me, seer, 0, game.ActivateAbilityParams{})
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scrying %d cards, want 1", len(c.ScryCards))
	}
	if !effectstest.HmlTapped(t, g, seer) {
		t.Error("the Battle-Seer did not tap")
	}
}

func TestStx5502B1HallMonitorStopsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	monitor := effectstest.B12Push(g, me.ID, "Hall Monitor", "Creature — Lizard Shaman", stxHallMonitorOracle, 1, 1)
	raider := effectstest.B12Creature(g, me.ID, "Raider", stxBearType, 2, 2)
	stopped := effectstest.B12Creature(g, opp.ID, "Stopped", stxBearType, 2, 2)
	free := effectstest.B12Creature(g, opp.ID, "Free", stxBearType, 2, 2)
	effectstest.ApaMana(me, "C", "R")
	effectstest.ApaActivate(t, g, me, monitor, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(stopped)})
	effectstest.BrAttack(t, g, raider)
	if err := g.DeclareBlocker(stopped, raider); err == nil {
		t.Error("the targeted creature blocked")
	}
	if err := g.DeclareBlocker(free, raider); err != nil {
		t.Errorf("another creature can't block: %v", err)
	}
}

func TestStx5502B1WormholeSerpentMakesACreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	serpent := effectstest.B12Push(g, me.ID, "Wormhole Serpent", "Creature — Serpent", stxWormholeSerpentOracle, 3, 5)
	raider := effectstest.B12Creature(g, me.ID, "Raider", stxBearType, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", stxBearType, 2, 2)
	effectstest.ApaMana(me, "C", "C", "C", "U")
	effectstest.ApaActivate(t, g, me, serpent, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(raider)})
	effectstest.BrAttack(t, g, raider)
	if err := g.DeclareBlocker(blocker, raider); err == nil {
		t.Error("the unblockable creature was blocked")
	}
}

func TestStx5502B1SpecterOfTheFensDrainsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	specter := effectstest.B12Push(g, me.ID, "Specter of the Fens", "Creature — Specter", stxSpecterOfTheFensOracle, 2, 3)
	effectstest.ApaMana(me, "C", "C", "C", "C", "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, specter, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}); err == nil {
		t.Fatal("the Specter targeted its controller")
	}
	mine, theirs := me.Life, opp.Life
	effectstest.ApaActivate(t, g, me, specter, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if opp.Life != theirs-2 || me.Life != mine+2 {
		t.Errorf("lives %d / %d, want %d / %d", me.Life, opp.Life, mine+2, theirs-2)
	}
}

func TestStx5502B1CogworkArchivistBottomsAGraveyardCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	archivist := effectstest.B12Push(g, me.ID, "Cogwork Archivist", "Artifact Creature — Construct", stxCogworkArchivistOracle, 4, 5)
	card := effectstest.PushCatalogGraveyardCard(opp, "Their Card", "Sorcery", "", 0, 0)
	effectstest.ApaMana(me, "C", "C")
	effectstest.ApaActivate(t, g, me, archivist, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(card)})
	if opp.Graveyard.Contains(card) {
		t.Fatal("the card stayed in the graveyard")
	}
	bottom, err := opp.Library.Bottom()
	if err != nil || bottom.InstanceID != card {
		t.Errorf("the bottom of its owner's library is %q (%v), want the card", bottom.Name, err)
	}
}

func TestStx5502B1UnwillingIngredientDrawsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	life, hand := me.Life, me.Hand.Size()
	id := effectstest.PushCatalogGraveyardCard(me, "Unwilling Ingredient", "Creature — Frog", stxUnwillingIngredientOracle, 1, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d→%d life %d→%d, want one card and 1 life", hand, me.Hand.Size(), life, me.Life)
	}
	if me.Graveyard.Contains(id) {
		t.Error("the card stayed in the graveyard")
	}
}

func TestStx5502B1LetterOfAcceptanceCashesIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	letter := effectstest.PushCatalogPermanent(g, me.ID, "Letter of Acceptance", "Artifact", stxLetterOfAcceptanceOracle, false)
	hand := me.Hand.Size()
	effectstest.ApaMana(me, "C", "C")
	effectstest.ApaActivate(t, g, me, letter, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(letter) || !me.Graveyard.Contains(letter) {
		t.Error("the Letter was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestStx5502B1ExcavatedWallMillsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	wall := effectstest.B12Push(g, me.ID, "Excavated Wall", "Artifact Creature — Wall", stxExcavatedWallOracle, 0, 4)
	library, graveyard := me.Library.Size(), me.Graveyard.Size()
	effectstest.ApaMana(me, "C")
	effectstest.ApaActivate(t, g, me, wall, 0, game.ActivateAbilityParams{})
	if got := library - me.Library.Size(); got != 1 {
		t.Errorf("milled %d, want 1", got)
	}
	if got := me.Graveyard.Size() - graveyard; got != 1 {
		t.Errorf("%d cards reached the graveyard, want 1", got)
	}
}

// --- statics ---------------------------------------------------------------------

func TestStx5502B1LeechFanaticHasLifelinkOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	fanatic := effectstest.B12Push(g, g.Seats[seat].ID, "Leech Fanatic", "Creature — Human Warlock", stxLeechFanaticOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, fanatic, "lifelink") {
		t.Error("no lifelink during your turn")
	}
	effectstest.AdvanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if effectstest.HasEffectiveKeyword(t, g, fanatic, "lifelink") {
		t.Error("lifelink during another player's turn")
	}
}

func TestStx5502B1ScurridColonyGrowsWithEightLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	colony := effectstest.B12Push(g, me.ID, "Scurrid Colony", "Creature — Squirrel", stxScurridColonyOracle, 2, 2)
	for i := 0; i < 7; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	if p, tough := effectstest.PtOf(t, g, colony); p != 2 || tough != 2 {
		t.Errorf("with seven lands it is %d/%d, want 2/2", p, tough)
	}
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if p, tough := effectstest.PtOf(t, g, colony); p != 4 || tough != 4 {
		t.Errorf("with eight lands it is %d/%d, want 4/4", p, tough)
	}
}

// --- the spells -----------------------------------------------------------------

func TestStx5502B1BeamingDefianceProtectsYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", stxBearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Beaming Defiance", "Instant", stxBeamingDefianceOracle, effectstest.CardTarget(theirs)); err == nil {
		t.Fatal("Beaming Defiance targeted an opponent's creature")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", stxBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Beaming Defiance", "Instant", stxBeamingDefianceOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the Bear has no hexproof")
	}
}

func TestStx5502B1ChargeThroughGrantsTrampleAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", stxBearType, 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Charge Through", "Instant", stxChargeThroughOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear has no trample")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestStx5502B1LashOfMaliceIsPlusTwoMinusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	ogre := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Lash of Malice", "Instant", stxLashOfMaliceOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, ogre); p != 5 || tough != 1 {
		t.Errorf("the Ogre is %d/%d, want 5/1", p, tough)
	}
}

func TestStx5502B1ExpelExilesOnlyATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped", stxBearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Expel", "Instant", stxExpelOracle, effectstest.CardTarget(untapped)); err == nil {
		t.Fatal("Expel targeted an untapped creature")
	}
	tapped := stxB1Tapped(g, opp.ID, "Tapped")
	effectstest.CastCatalogSpell(t, g, "Expel", "Instant", stxExpelOracle, effectstest.CardTarget(tapped))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) || !g.Exile.Contains(tapped) {
		t.Error("the tapped creature was not exiled")
	}
}

func TestStx5502B1HeatedDebateDealsFourAndCantBeCountered(t *testing.T) {
	spec, ok := effects.Lookup(stxHeatedDebateOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatal("Heated Debate can be countered")
	}
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	giant := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Heated Debate", "Instant", stxHeatedDebateOracle, effectstest.CardTarget(giant))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, giant).DamageMarked; got != 4 {
		t.Errorf("the Giant has %d damage, want 4", got)
	}
}

func TestStx5502B1IntroductionToProphecyScriesThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Introduction to Prophecy", stxLessonType, stxIntroToProphecyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scrying %d cards, want 2", len(c.ScryCards))
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("drew before the scry was answered: hand %d, want %d", got, hand)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestStx5502B1LessonsMakeTheirTokens(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		want         map[string]int
	}{
		{"Spirit Summoning", stxSpiritSummoningOracle, map[string]int{"Spirit": 1}},
		{"Inkling Summoning", stxInklingSummoningOracle, map[string]int{"Inkling": 1}},
		{"Elemental Summoning", stxElementalSummoningOracle, map[string]int{"Elemental": 1}},
		{"Pest Summoning", stxPestSummoningOracle, map[string]int{"Pest": 2}},
		{"Mascot Exhibition", stxMascotExhibitionOracle, map[string]int{"Inkling": 1, "Spirit": 1, "Elemental": 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			effectstest.CastCatalogSpell(t, g, tc.name, stxLessonType, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			for name, n := range tc.want {
				if got := effectstest.CountBattlefieldNamed(g, me.ID, name); got != n {
					t.Errorf("%d %s tokens, want %d", got, name, n)
				}
			}
		})
	}
}

func TestStx5502B1LessonTokenBodies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Mascot Exhibition", stxLessonType, stxMascotExhibitionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		name       string
		power, tgh int
		keyword    string
	}{
		{"Inkling", 2, 1, "flying"},
		{"Spirit", 3, 2, ""},
		{"Elemental", 4, 4, ""},
	} {
		tokens := effectstest.TokensNamed(g, tc.name)
		if len(tokens) != 1 {
			t.Errorf("%d %s tokens, want 1", len(tokens), tc.name)
			continue
		}
		id := tokens[0].InstanceID
		if p, tough := effectstest.PtOf(t, g, id); p != tc.power || tough != tc.tgh {
			t.Errorf("the %s is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.tgh)
		}
		if tc.keyword != "" && !effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
			t.Errorf("the %s has no %s", tc.name, tc.keyword)
		}
	}
}

func TestStx5502B1StartFromScratchEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	life := opp.Life
	effectstest.CastModal(t, g, "Start from Scratch", stxLessonType, stxStartFromScratchOracle, []int{0}, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-1)
	}
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastModal(t, g, "Start from Scratch", stxLessonType, stxStartFromScratchOracle, []int{1}, effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestStx5502B1TangletrapEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	walker := effectstest.B12Creature(g, opp.ID, "Walker", stxBearType, 6, 6)
	if _, err := effectstest.CastModalErr(t, g, "Tangletrap", "Instant", stxTangletrapOracle, []int{0}, effectstest.CardTarget(walker)); err == nil {
		t.Fatal("the flying bullet targeted a creature without flying")
	}
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 6, Toughness: 6, Keywords: []string{"flying"}})
	effectstest.CastModal(t, g, "Tangletrap", "Instant", stxTangletrapOracle, []int{0}, effectstest.CardTarget(flyer))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, flyer).DamageMarked; got != 5 {
		t.Errorf("the flyer has %d damage, want 5", got)
	}
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastModal(t, g, "Tangletrap", "Instant", stxTangletrapOracle, []int{1}, effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestStx5502B1SuddenBreakthroughPumpsAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", stxBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Sudden Breakthrough", "Instant", stxSuddenBreakthroughOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the Bear has no first strike")
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
}

func TestStx5502B1TwoCountersAndAKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Expanded Anatomy", stxLessonType, stxExpandedAnatomyOracle, "vigilance"},
		{"Essence Infusion", "Sorcery", stxEssenceInfusionOracle, "lifelink"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			bear := effectstest.B12Creature(g, me.ID, "Bear", stxBearType, 2, 2)
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(bear))
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.PlusOneCounters(g, bear); got != 2 {
				t.Errorf("%d +1/+1 counters, want 2", got)
			}
			if !effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
		})
	}
}
