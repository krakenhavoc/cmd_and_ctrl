package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// znr_5508_b1_cards_test.go — Zendikar Rising batch 1 (#5508). Every test
// plays the real catalog entry through the engine.

const (
	znrB1CliffhavenSellSwordOracle = "5f0c8c60-96f1-4b26-a9f6-386e1cec517d"
	znrB1MurasaBruteOracle         = "5f84a140-577f-4a1e-a900-793664332fd6"
	znrB1LivingTempestOracle       = "d684553a-f225-49af-90eb-399f55ef65df"
	znrB1TajuruBlightbladeOracle   = "debbbdd8-c8e2-4384-bcac-dbdfb18b54fa"
	znrB1TajuruSnarecasterOracle   = "2061fe08-0c7c-4d79-a144-996c3012952c"
	znrB1CragplateBalothOracle     = "3f599e15-cbc5-4170-b6b4-efd16aa222ac"
	znrB1GhastlyGloomhunterOracle  = "70ab65ff-724f-4126-b472-97fa03ca3a34"
	znrB1SkyclaveSentinelOracle    = "3b7b166b-4329-4194-a9a7-f1fec7a92cec"
	znrB1KarganWarleaderOracle     = "63324e72-e580-456a-91be-766c9f07b7b3"
	znrB1TuktukRubblefortOracle    = "48dab180-4ffd-4b39-9f5a-6bdc017a1eae"
	znrB1MesaLynxOracle            = "211add65-b63d-43ad-8d69-b0598e003fef"
	znrB1ExpeditionHealerOracle    = "e8fbdf4b-ef64-4b82-b67f-ae5568772256"
	znrB1HagraConstrictorOracle    = "3282c7fb-b43d-4dd6-b5a3-cc7d93e01f19"
	znrB1AkoumHellhoundOracle      = "05074b68-6b31-4f9f-b3ec-bffe3e7eab47"
	znrB1SkyclaveGeopedeOracle     = "e00d601f-ac73-47a6-8920-779d71b482c9"
	znrB1BrushfireElementalOracle  = "b33a7ec6-b22b-41cf-9312-65add1529b51"
	znrB1ProwlingFelidarOracle     = "ed26f6bf-fcd6-4cf4-9b65-a492cb79f2df"
	znrB1CanyonJerboaOracle        = "254d4918-de40-4ac4-b542-f5c67a5e21eb"
	znrB1DreadwurmOracle           = "9feb7600-f844-4573-bcb7-d6b34473517b"
	znrB1MakindiOxOracle           = "e4f0fa12-542b-485f-a792-3a8aba933522"
	znrB1FearlessFledglingOracle   = "75ae9b87-7137-47a0-b3e1-8915ccd0edf1"
	znrB1JoragaVisionaryOracle     = "7854f475-5470-4f2b-99df-af9026895595"
	znrB1TurntimberAsceticOracle   = "90b16d52-4cf4-4788-a0d7-264d1bcd17f0"
	znrB1DauntlessSurvivorOracle   = "8371dee1-5f4a-410a-8b7c-e4110cdbe9bd"
	znrB1AngelheartProtectorOracle = "42aa599f-de2f-411a-8cf3-f5cfb40568b3"
	znrB1FarsightAdeptOracle       = "72b10cba-0f46-4b99-9634-e7173ac4c259"
	znrB1NimanaSkydancerOracle     = "27a66362-f2f5-4d46-9faf-a73752753304"
	znrB1KorCelebrantOracle        = "3f8f0ce9-d2ce-45cc-9fbf-0e2819b07a3e"
	znrB1ScorchRiderOracle         = "8165acd4-ac62-47f1-97dd-01860aee29ee"
	znrB1CunningGeysermageOracle   = "746768bc-9abc-495f-ad00-4b58497dce0d"
	znrB1KitesailClericOracle      = "0b9bb0d7-4139-46b6-a058-2f79be64aef2"
	znrB1TazeemRoilmageOracle      = "55a277b4-575f-463b-913c-7857acad20be"
	znrB1SeaGateBanneretOracle     = "ddf12758-b356-4ff8-b0ce-547083498335"
	znrB1TeeterpeakAmbusherOracle  = "fe769071-3f99-411b-878a-57a2200649a0"
	znrB1GuulDrazMucklordOracle    = "ad62cd58-fe6f-4844-b3dc-5ba21327fc43"
	znrB1ClericOfChillDepthsOracle = "99c13f8d-7f18-48ce-86da-c74ddb6c1205"
	znrB1DeliberateOracle          = "362fc643-ab76-4160-a73e-73397464db40"
	znrB1ThunderingRebukeOracle    = "bb4d9770-f230-4413-a7a4-bb1f1303661a"
	znrB1MoltenBlastOracle         = "5a680186-a84d-400f-9452-76f19bdfafa3"
	znrB1InordinateRageOracle      = "50402103-14ea-49b0-8e66-d2b7f7a69a41"
	znrB1FieldResearchOracle       = "bb377cf8-29ee-489e-86a9-93f4dc2ae02e"
	znrB1MightOfMurasaOracle       = "902eb71f-62b3-40cb-816b-f6b7e92e91e7"
	znrB1CinderclasmOracle         = "d7b49bee-2cb0-48e1-91c1-1e150a0da6ed"
	znrB1ArdentElectromancerOracle = "43827acb-70cf-44a3-8560-e9f7330c835e"
	znrB1MalakirBloodPriestOracle  = "3a3cb6fe-24ef-4c09-8ff8-f303fc472a8b"
	znrB1ShepherdOfHeroesOracle    = "6a732900-6529-4d5c-ab94-52bd0048d483"
	znrB1EmeriaCaptainOracle       = "151cafcb-a589-4681-a914-7313246a7626"
	znrB1ThunderingSparkmageOracle = "a786e495-ae3a-4e15-ab32-661d524e2b2d"
	znrB1KabiraOutriderOracle      = "69f9a2fc-4873-4d7f-8405-3fb9a2091de5"

	znrB1BearType = "Creature — Bear"
	znrB1Forest   = "Basic Land — Forest"
)

// znrB1CastKicked casts a creature spell with real P/T from the active
// player's hand, paying its kicker (row 0), and settles the stack.
func znrB1CastKicked(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("CastSpell %s kicked: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// znrB1CastSpell casts a catalog instant or sorcery from the active
// player's hand, kicked or not, and settles the stack.
func znrB1CastSpell(t *testing.T, g *game.Game, name, typeLine, oracle string, kicked bool, targets []game.TargetRef) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := b13HandCard(active, name, typeLine, oracle)
	advanceToMain(t, g)
	params := game.CastSpellParams{Targets: targets}
	if kicked {
		params.OptionalCosts = []int{0}
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s (kicked=%v): %v", name, kicked, err)
	}
	passPriorityAroundTable(t, g)
}

// znrB1PlayForest has the active player play a Forest from a main phase.
func znrB1PlayForest(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToMain(t, g)
	b31PlayLand(t, g, "Forest", znrB1Forest, "")
	passPriorityAroundTable(t, g)
}

// --- every card is in the catalog ----------------------------------------

func TestZnrB1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		znrB1CliffhavenSellSwordOracle, znrB1MurasaBruteOracle, znrB1LivingTempestOracle, znrB1TajuruBlightbladeOracle,
		znrB1TajuruSnarecasterOracle, znrB1CragplateBalothOracle, znrB1GhastlyGloomhunterOracle, znrB1SkyclaveSentinelOracle,
		znrB1KarganWarleaderOracle, znrB1TuktukRubblefortOracle, znrB1MesaLynxOracle, znrB1ExpeditionHealerOracle,
		znrB1HagraConstrictorOracle, znrB1AkoumHellhoundOracle, znrB1SkyclaveGeopedeOracle, znrB1BrushfireElementalOracle,
		znrB1ProwlingFelidarOracle, znrB1CanyonJerboaOracle, znrB1DreadwurmOracle, znrB1MakindiOxOracle,
		znrB1FearlessFledglingOracle, znrB1JoragaVisionaryOracle, znrB1TurntimberAsceticOracle, znrB1DauntlessSurvivorOracle,
		znrB1AngelheartProtectorOracle, znrB1FarsightAdeptOracle, znrB1NimanaSkydancerOracle, znrB1KorCelebrantOracle,
		znrB1ScorchRiderOracle, znrB1CunningGeysermageOracle, znrB1KitesailClericOracle, znrB1TazeemRoilmageOracle,
		znrB1SeaGateBanneretOracle, znrB1TeeterpeakAmbusherOracle, znrB1GuulDrazMucklordOracle, znrB1ClericOfChillDepthsOracle,
		znrB1DeliberateOracle, znrB1ThunderingRebukeOracle, znrB1MoltenBlastOracle, znrB1InordinateRageOracle,
		znrB1FieldResearchOracle, znrB1MightOfMurasaOracle, znrB1CinderclasmOracle, znrB1ArdentElectromancerOracle,
		znrB1MalakirBloodPriestOracle, znrB1ShepherdOfHeroesOracle, znrB1EmeriaCaptainOracle, znrB1ThunderingSparkmageOracle,
		znrB1KabiraOutriderOracle,
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

func TestZnrB1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Living Tempest", "Creature — Elemental", znrB1LivingTempestOracle, []string{"flash", "flying"}},
		{"Tajuru Blightblade", "Creature — Elf Rogue", znrB1TajuruBlightbladeOracle, []string{"deathtouch"}},
		{"Tajuru Snarecaster", "Creature — Elf Rogue", znrB1TajuruSnarecasterOracle, []string{"reach"}},
		{"Cragplate Baloth", "Creature — Beast", znrB1CragplateBalothOracle, []string{"hexproof", "haste"}},
		{"Ghastly Gloomhunter", "Creature — Zombie Bat", znrB1GhastlyGloomhunterOracle, []string{"flying", "lifelink"}},
		{"Skyclave Sentinel", "Artifact Creature — Gargoyle", znrB1SkyclaveSentinelOracle, []string{"flying", "defender"}},
		{"Tuktuk Rubblefort", "Creature — Wall", znrB1TuktukRubblefortOracle, []string{"defender", "reach"}},
		{"Expedition Healer", "Creature — Kor Cleric", znrB1ExpeditionHealerOracle, []string{"vigilance"}},
		{"Skyclave Geopede", "Creature — Insect", znrB1SkyclaveGeopedeOracle, []string{"trample"}},
		{"Brushfire Elemental", "Creature — Elemental", znrB1BrushfireElementalOracle, []string{"haste"}},
		{"Prowling Felidar", "Creature — Cat Beast", znrB1ProwlingFelidarOracle, []string{"vigilance"}},
		{"Kitesail Cleric", "Creature — Kor Cleric", znrB1KitesailClericOracle, []string{"flying"}},
		{"Nimana Skydancer", "Creature — Human Rogue", znrB1NimanaSkydancerOracle, []string{"flash", "flying"}},
		{"Shepherd of Heroes", "Creature — Angel Cleric", znrB1ShepherdOfHeroesOracle, []string{"flying"}},
		{"Emeria Captain", "Creature — Angel Warrior", znrB1EmeriaCaptainOracle, []string{"flying", "vigilance"}},
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

// --- kicker entry counters -------------------------------------------------------

func TestZnrB1KickedCreaturesEnterWithCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness, want int
	}{
		{"Cragplate Baloth", "Creature — Beast", znrB1CragplateBalothOracle, 6, 6, 4},
		{"Ghastly Gloomhunter", "Creature — Zombie Bat", znrB1GhastlyGloomhunterOracle, 1, 1, 2},
		{"Skyclave Sentinel", "Artifact Creature — Gargoyle", znrB1SkyclaveSentinelOracle, 2, 3, 2},
	} {
		t.Run(tc.name+" kicked", func(t *testing.T) {
			g := newCatalogGame(t)
			id := znrB1CastKicked(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			if got := countersOn(g, id, game.CounterPlusOne); got != tc.want {
				t.Errorf("%d counters, want %d", got, tc.want)
			}
		})
		t.Run(tc.name+" unkicked", func(t *testing.T) {
			g := newCatalogGame(t)
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			if got := countersOn(g, id, game.CounterPlusOne); got != 0 {
				t.Errorf("%d counters unkicked, want 0", got)
			}
		})
	}
}

func TestZnrB1CragplateBalothCantBeCountered(t *testing.T) {
	spec, ok := Lookup(znrB1CragplateBalothOracle)
	if !ok || !spec.CantBeCountered {
		t.Error("Cragplate Baloth is not uncounterable")
	}
}

// The Sentinel attacks only while it has a +1/+1 counter.
func TestZnrB1SkyclaveSentinelAttacksOnlyWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bare := b12Push(g, me.ID, "Skyclave Sentinel", "Artifact Creature — Gargoyle", znrB1SkyclaveSentinelOracle, 2, 3)
	grown := apaPush(g, me.ID, me.ID, game.Card{
		Name: "Skyclave Sentinel", OracleID: znrB1SkyclaveSentinelOracle, TypeLine: "Artifact Creature — Gargoyle",
		Power: 2, Toughness: 3, Counters: map[string]int{game.CounterPlusOne: 1},
	})
	advanceToDeclareAttackersOf(t, g, seat)
	if err := g.DeclareAttacker(bare, opp.ID); err == nil {
		t.Error("a Sentinel with no counter attacked through defender")
	}
	if err := g.DeclareAttacker(grown, opp.ID); err != nil {
		t.Errorf("a Sentinel with a counter can't attack: %v", err)
	}
}

// --- statics ---------------------------------------------------------------------

func TestZnrB1KarganWarleaderPumpsYourOtherWarriors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := b12Push(g, me.ID, "Kargan Warleader", "Creature — Human Warrior", znrB1KarganWarleaderOracle, 3, 3)
	warrior := b12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", znrB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Warrior", "Creature — Human Warrior", 2, 2)
	for _, tc := range []struct {
		id      uuid.UUID
		what    string
		p, tous int
	}{
		{lord, "the Warleader", 3, 3}, {warrior, "your Warrior", 3, 3}, {bear, "your Bear", 2, 2}, {theirs, "their Warrior", 2, 2},
	} {
		if p, tough := ptOf(t, g, tc.id); p != tc.p || tough != tc.tous {
			t.Errorf("%s is %d/%d, want %d/%d", tc.what, p, tough, tc.p, tc.tous)
		}
	}
}

func TestZnrB1TuktukRubblefortGivesYourCreaturesHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tuktuk Rubblefort", "Creature — Wall", znrB1TuktukRubblefortOracle, 0, 3)
	mine := b12Creature(g, me.ID, "Mine", znrB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", znrB1BearType, 2, 2)
	if !hasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("your creature has no haste")
	}
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
}

func TestZnrB1MesaLynxIsBiggerOnOtherTurns(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	lynx := b12Push(g, g.Seats[seat].ID, "Mesa Lynx", "Creature — Cat", znrB1MesaLynxOracle, 2, 1)
	advanceToMain(t, g)
	if p, tough := ptOf(t, g, lynx); p != 2 || tough != 1 {
		t.Errorf("on your turn the Lynx is %d/%d, want 2/1", p, tough)
	}
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if p, tough := ptOf(t, g, lynx); p != 2 || tough != 3 {
		t.Errorf("on another player's turn the Lynx is %d/%d, want 2/3", p, tough)
	}
}

func TestZnrB1ExpeditionHealerHasLifelinkWithAnotherCleric(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	healer := b12Push(g, me.ID, "Expedition Healer", "Creature — Kor Cleric", znrB1ExpeditionHealerOracle, 2, 2)
	if hasEffectiveKeyword(t, g, healer, "lifelink") {
		t.Error("the Healer has lifelink with no other Cleric")
	}
	b12Creature(g, me.ID, "Priest", "Creature — Human Cleric", 1, 1)
	if !hasEffectiveKeyword(t, g, healer, "lifelink") {
		t.Error("the Healer has no lifelink beside another Cleric")
	}
}

func TestZnrB1HagraConstrictorEntersWithCountersAndGrantsMenace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plain := b12Creature(g, me.ID, "Plain", znrB1BearType, 2, 2)
	grown := apaPush(g, me.ID, me.ID, game.Card{
		Name: "Grown", TypeLine: znrB1BearType, Power: 2, Toughness: 2,
		Counters: map[string]int{game.CounterPlusOne: 1},
	})
	snake := castHoldCreature(t, g, "Hagra Constrictor", "Creature — Snake", znrB1HagraConstrictorOracle, 0, 0)
	if !g.Battlefield.Contains(snake) {
		t.Fatal("the Constrictor did not survive entering")
	}
	if got := countersOn(g, snake, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, snake, "menace") || !hasEffectiveKeyword(t, g, grown, "menace") {
		t.Error("a creature you control with a +1/+1 counter has no menace")
	}
	if hasEffectiveKeyword(t, g, plain, "menace") {
		t.Error("a creature with no counter has menace")
	}
}

// --- landfall ----------------------------------------------------------------------

func TestZnrB1LandfallPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
	}{
		{"Akoum Hellhound", "Creature — Elemental Dog", znrB1AkoumHellhoundOracle, 0, 1},
		{"Skyclave Geopede", "Creature — Insect", znrB1SkyclaveGeopedeOracle, 3, 1},
		{"Brushfire Elemental", "Creature — Elemental", znrB1BrushfireElementalOracle, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[g.Turn.ActiveSeat].ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			znrB1PlayForest(t, g)
			if p, tough := ptOf(t, g, id); p != tc.p+2 || tough != tc.tough+2 {
				t.Errorf("after landfall %d/%d, want %d/%d", p, tough, tc.p+2, tc.tough+2)
			}
		})
	}
}

func TestZnrB1BrushfireElementalEvadesSmallBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elemental := b12Push(g, me.ID, "Brushfire Elemental", "Creature — Elemental", znrB1BrushfireElementalOracle, 1, 1)
	small := b12Creature(g, opp.ID, "Small", znrB1BearType, 2, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 3, 3)
	brAttack(t, g, elemental)
	if err := g.DeclareBlocker(small, elemental); err == nil {
		t.Error("a creature with power 2 blocked the Elemental")
	}
	if err := g.DeclareBlocker(big, elemental); err != nil {
		t.Errorf("a creature with power 3 can't block the Elemental: %v", err)
	}
}

func TestZnrB1ProwlingFelidarGetsACounterOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	cat := b12Push(g, g.Seats[g.Turn.ActiveSeat].ID, "Prowling Felidar", "Creature — Cat Beast", znrB1ProwlingFelidarOracle, 2, 3)
	znrB1PlayForest(t, g)
	if got := countersOn(g, cat, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after landfall, want 1", got)
	}
}

func TestZnrB1CanyonJerboaPumpsYourTeamOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	b12Push(g, me.ID, "Canyon Jerboa", "Creature — Mouse", znrB1CanyonJerboaOracle, 1, 2)
	mine := b12Creature(g, me.ID, "Mine", znrB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", znrB1BearType, 2, 2)
	znrB1PlayForest(t, g)
	if p, tough := ptOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestZnrB1DreadwurmGainsIndestructibleOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	wurm := b12Push(g, g.Seats[g.Turn.ActiveSeat].ID, "Dreadwurm", "Creature — Wurm Horror", znrB1DreadwurmOracle, 5, 4)
	if hasEffectiveKeyword(t, g, wurm, "indestructible") {
		t.Fatal("indestructible before any land entered")
	}
	znrB1PlayForest(t, g)
	if !hasEffectiveKeyword(t, g, wurm, "indestructible") {
		t.Error("no indestructible after landfall")
	}
}

func TestZnrB1MakindiOxTapsAnOpposingCreatureOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	b12Push(g, me.ID, "Makindi Ox", "Creature — Ox", znrB1MakindiOxOracle, 4, 4)
	theirs := b12Creature(g, opp.ID, "Theirs", znrB1BearType, 2, 2)
	advanceToMain(t, g)
	b31PlayLand(t, g, "Forest", znrB1Forest, "")
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if !hmlTapped(t, g, theirs) {
		t.Error("the opponent's creature is untapped")
	}
}

func TestZnrB1FearlessFledglingGrowsAndFliesOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	bird := b12Push(g, g.Seats[g.Turn.ActiveSeat].ID, "Fearless Fledgling", "Creature — Griffin", znrB1FearlessFledglingOracle, 1, 1)
	znrB1PlayForest(t, g)
	if got := countersOn(g, bird, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, bird, "flying") {
		t.Error("no flying after landfall")
	}
}

// --- enters triggers ---------------------------------------------------------------

func TestZnrB1JoragaVisionaryDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Joraga Visionary", "Creature — Elf Wizard", znrB1JoragaVisionaryOracle, 3, 2)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestZnrB1TurntimberAsceticGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	life := me.Life
	castHoldCreature(t, g, "Turntimber Ascetic", "Creature — Giant Cleric", znrB1TurntimberAsceticOracle, 5, 4)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestZnrB1DauntlessSurvivorPutsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	bear := b12Creature(g, me.ID, "Bear", znrB1BearType, 2, 2)
	castHoldCreature(t, g, "Dauntless Survivor", "Creature — Human Warrior", znrB1DauntlessSurvivorOracle, 1, 1)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
}

func TestZnrB1AngelheartProtectorGrantsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	bear := b12Creature(g, me.ID, "Bear", znrB1BearType, 2, 2)
	castHoldCreature(t, g, "Angelheart Protector", "Creature — Human Cleric", znrB1AngelheartProtectorOracle, 3, 2)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the target has no indestructible")
	}
}

func TestZnrB1FarsightAdeptDrawsForYouAndTheOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	advanceToMain(t, g)
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	castHoldCreature(t, g, "Farsight Adept", "Creature — Kor Wizard", znrB1FarsightAdeptOracle, 3, 3)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != mine+1 {
		t.Errorf("your hand is %d, want %d", got, mine+1)
	}
	if got := opp.Hand.Size(); got != theirs+1 {
		t.Errorf("the opponent's hand is %d, want %d", got, theirs+1)
	}
}

func TestZnrB1NimanaSkydancerMillsTheOpponentTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := top100Opponent(g)
	library := opp.Library.Size()
	castHoldCreature(t, g, "Nimana Skydancer", "Creature — Human Rogue", znrB1NimanaSkydancerOracle, 2, 1)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := library - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

func TestZnrB1KorCelebrantGainsForItselfAndEachOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	life := me.Life
	castHoldCreature(t, g, "Kor Celebrant", "Creature — Kor Cleric", znrB1KorCelebrantOracle, 1, 4)
	if me.Life != life+1 {
		t.Fatalf("life %d after the Celebrant entered, want %d", me.Life, life+1)
	}
	castHoldCreature(t, g, "Bear", znrB1BearType, "", 2, 2)
	if me.Life != life+2 {
		t.Errorf("life %d after another creature entered, want %d", me.Life, life+2)
	}
}

// --- kicked enters triggers ---------------------------------------------------------

func TestZnrB1ScorchRiderHasHasteOnlyWhenKicked(t *testing.T) {
	g := newCatalogGame(t)
	kicked := znrB1CastKicked(t, g, "Scorch Rider", "Creature — Human Warrior", znrB1ScorchRiderOracle, 4, 3)
	if !hasEffectiveKeyword(t, g, kicked, "haste") {
		t.Error("a kicked Rider has no haste")
	}
	plain := castHoldCreature(t, g, "Scorch Rider", "Creature — Human Warrior", znrB1ScorchRiderOracle, 4, 3)
	if hasEffectiveKeyword(t, g, plain, "haste") {
		t.Error("an unkicked Rider has haste")
	}
}

func TestZnrB1CunningGeysermageBouncesWhenKicked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	theirs := b12Creature(g, opp.ID, "Theirs", znrB1BearType, 2, 2)
	castHoldCreature(t, g, "Cunning Geysermage", "Creature — Human Wizard", znrB1CunningGeysermageOracle, 3, 2)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("an unkicked Geysermage asked for a target")
	}
	znrB1CastKicked(t, g, "Cunning Geysermage", "Creature — Human Wizard", znrB1CunningGeysermageOracle, 3, 2)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the target did not return to its owner's hand")
	}
}

func TestZnrB1KitesailClericTapsTwoWhenKicked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	a := b12Creature(g, opp.ID, "A", znrB1BearType, 2, 2)
	b := b12Creature(g, opp.ID, "B", znrB1BearType, 2, 2)
	castHoldCreature(t, g, "Kitesail Cleric", "Creature — Kor Cleric", znrB1KitesailClericOracle, 1, 1)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("an unkicked Cleric asked for targets")
	}
	znrB1CastKicked(t, g, "Kitesail Cleric", "Creature — Kor Cleric", znrB1KitesailClericOracle, 1, 1)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the kicked tap")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(a, b)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hmlTapped(t, g, a) || !hmlTapped(t, g, b) {
		t.Error("both targets should be tapped")
	}
}

func TestZnrB1TazeemRoilmageReturnsASpellWhenKicked(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	spell := wthGraveyardCard(me, "Old Spell", "Instant")
	znrB1CastKicked(t, g, "Tazeem Roilmage", "Creature — Merfolk Wizard", znrB1TazeemRoilmageOracle, 2, 1)
	pickCard(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("the instant did not return to hand")
	}
}

// --- activated, dies and block triggers ----------------------------------------------

func TestZnrB1SeaGateBanneretPumpsYourTeam(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	banneret := b12Push(g, me.ID, "Sea Gate Banneret", "Creature — Kor Warrior", znrB1SeaGateBanneretOracle, 1, 2)
	mine := b12Creature(g, me.ID, "Mine", znrB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", znrB1BearType, 2, 2)
	apaMana(me, "W", "W", "W", "W", "W")
	apaActivate(t, g, me, banneret, 0, game.ActivateAbilityParams{})
	if p, tough := ptOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestZnrB1TeeterpeakAmbusherPumpsItself(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	goblin := b12Push(g, me.ID, "Teeterpeak Ambusher", "Creature — Goblin Warrior", znrB1TeeterpeakAmbusherOracle, 1, 3)
	apaMana(me, "R", "R", "R")
	apaActivate(t, g, me, goblin, 0, game.ActivateAbilityParams{})
	if p, tough := ptOf(t, g, goblin); p != 3 || tough != 3 {
		t.Errorf("the Ambusher is %d/%d, want 3/3", p, tough)
	}
}

func TestZnrB1GuulDrazMucklordPutsACounterWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	croc := b12Push(g, me.ID, "Guul Draz Mucklord", "Creature — Crocodile", znrB1GuulDrazMucklordOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", znrB1BearType, 2, 2)
	b27Kill(g, croc)
	g.RunStateChecksForTest()
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
}

func TestZnrB1ClericOfChillDepthsHoldsTheCreatureItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := hmlCreature(g, me.ID, "Attacker", "", 1, 5)
	cleric := hmlCreature(g, opp.ID, "Cleric of Chill Depths", znrB1ClericOfChillDepthsOracle, 1, 3)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(cleric, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, 0)
	if !tappedForTest(t, g, attacker) {
		t.Error("the blocked creature untapped during its controller's next untap step")
	}
}

// --- spells --------------------------------------------------------------------------

func TestZnrB1DeliberateScriesTwoThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Deliberate", "Instant", znrB1DeliberateOracle, nil)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scrying %d cards, want 2", len(c.ScryCards))
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestZnrB1ThunderingRebukeDealsFour(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 5)
	if err := castCatalogSpellErr(t, g, "Thundering Rebuke", "Sorcery", znrB1ThunderingRebukeOracle,
		[]game.TargetRef{playerTarget(opp.ID)}); err == nil {
		t.Fatal("Thundering Rebuke targeted a player")
	}
	znrB1CastSpell(t, g, "Thundering Rebuke", "Sorcery", znrB1ThunderingRebukeOracle, false, cardTarget(ogre))
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 4 {
		t.Errorf("the Ogre has %d damage, want 4", got)
	}
}

func TestZnrB1MoltenBlastEachMode(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 5)
	castModal(t, g, "Molten Blast", "Instant", znrB1MoltenBlastOracle, []int{0}, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 2 {
		t.Errorf("the Ogre has %d damage, want 2", got)
	}
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	castModal(t, g, "Molten Blast", "Instant", znrB1MoltenBlastOracle, []int{1}, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestZnrB1InordinateRagePumpsThenScries(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := b12Creature(g, me.ID, "Bear", znrB1BearType, 2, 2)
	znrB1CastSpell(t, g, "Inordinate Rage", "Instant", znrB1InordinateRageOracle, false, cardTarget(bear))
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 5/4", p, tough)
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Error("no scry 1 prompt")
	}
}

func TestZnrB1FieldResearchDrawsTwoOrThree(t *testing.T) {
	for _, tc := range []struct {
		kicked bool
		want   int
	}{{false, 2}, {true, 3}} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := top100ActiveSeat(g)
		hand := me.Hand.Size()
		znrB1CastSpell(t, g, "Field Research", "Sorcery", znrB1FieldResearchOracle, tc.kicked, nil)
		if got := me.Hand.Size() - hand; got != tc.want {
			t.Errorf("kicked=%v: drew %d, want %d", tc.kicked, got, tc.want)
		}
	}
}

func TestZnrB1MightOfMurasaPumpsThreeOrFive(t *testing.T) {
	for _, tc := range []struct {
		kicked bool
		want   int
	}{{false, 5}, {true, 7}} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := top100ActiveSeat(g)
		bear := b12Creature(g, me.ID, "Bear", znrB1BearType, 2, 2)
		znrB1CastSpell(t, g, "Might of Murasa", "Instant", znrB1MightOfMurasaOracle, tc.kicked, cardTarget(bear))
		if p, tough := ptOf(t, g, bear); p != tc.want || tough != tc.want {
			t.Errorf("kicked=%v: the Bear is %d/%d, want %d/%d", tc.kicked, p, tough, tc.want, tc.want)
		}
	}
}

func TestZnrB1CinderclasmDealsOneOrTwoToEachCreature(t *testing.T) {
	for _, tc := range []struct {
		kicked bool
		dies   int // toughness that dies
	}{{false, 1}, {true, 2}} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me, opp := top100ActiveSeat(g), top100Opponent(g)
		small := b12Creature(g, opp.ID, "Small", znrB1BearType, 1, tc.dies)
		big := b12Creature(g, me.ID, "Big", znrB1BearType, 1, tc.dies+1)
		znrB1CastSpell(t, g, "Cinderclasm", "Instant", znrB1CinderclasmOracle, tc.kicked, nil)
		if g.Battlefield.Contains(small) {
			t.Errorf("kicked=%v: the 1/%d survived", tc.kicked, tc.dies)
		}
		if !g.Battlefield.Contains(big) {
			t.Errorf("kicked=%v: the 1/%d died", tc.kicked, tc.dies+1)
		}
	}
}

// --- party ---------------------------------------------------------------------------

// With the other three party types beside it, each party card counts
// four; alone, it counts itself.

func TestZnrB1ArdentElectromancerAddsRedPerPartyMember(t *testing.T) {
	for _, tc := range []struct {
		full bool
		want int
	}{{false, 1}, {true, 4}} {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		if tc.full {
			finB9FillParty(g, me.ID, "Wizard")
		}
		castHoldCreature(t, g, "Ardent Electromancer", "Creature — Human Wizard", znrB1ArdentElectromancerOracle, 3, 2)
		if got := me.ManaPool.Total(); got != tc.want {
			t.Errorf("full=%v: %d mana in the pool, want %d", tc.full, got, tc.want)
		}
	}
}

func TestZnrB1MalakirBloodPriestDrainsPerPartyMember(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	finB9FillParty(g, me.ID, "Cleric")
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	castHoldCreature(t, g, "Malakir Blood-Priest", "Creature — Vampire Cleric", znrB1MalakirBloodPriestOracle, 2, 1)
	for _, s := range g.Seats {
		want := lives[s.ID] - 4
		if s.ID == me.ID {
			want = lives[s.ID] + 4
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
}

func TestZnrB1ShepherdOfHeroesGainsTwoPerPartyMember(t *testing.T) {
	for _, tc := range []struct {
		full bool
		want int
	}{{false, 2}, {true, 8}} {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		if tc.full {
			finB9FillParty(g, me.ID, "Cleric")
		}
		life := me.Life
		castHoldCreature(t, g, "Shepherd of Heroes", "Creature — Angel Cleric", znrB1ShepherdOfHeroesOracle, 3, 4)
		if got := me.Life - life; got != tc.want {
			t.Errorf("full=%v: gained %d, want %d", tc.full, got, tc.want)
		}
	}
}

func TestZnrB1EmeriaCaptainCountersPerPartyMember(t *testing.T) {
	for _, tc := range []struct {
		full bool
		want int
	}{{false, 1}, {true, 4}} {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		if tc.full {
			finB9FillParty(g, me.ID, "Warrior")
		}
		id := castHoldCreature(t, g, "Emeria Captain", "Creature — Angel Warrior", znrB1EmeriaCaptainOracle, 1, 1)
		if got := countersOn(g, id, game.CounterPlusOne); got != tc.want {
			t.Errorf("full=%v: %d counters, want %d", tc.full, got, tc.want)
		}
	}
}

func TestZnrB1ThunderingSparkmageDealsPartyDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	finB9FillParty(g, me.ID, "Wizard")
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 6)
	castHoldCreature(t, g, "Thundering Sparkmage", "Creature — Human Wizard", znrB1ThunderingSparkmageOracle, 2, 2)
	answerPickTarget(t, g, ogre)
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 4 {
		t.Errorf("the Ogre has %d damage, want 4", got)
	}
}

func TestZnrB1KabiraOutriderPumpsPerPartyMember(t *testing.T) {
	for _, tc := range []struct {
		full bool
		want int
	}{{false, 3}, {true, 6}} {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		if tc.full {
			finB9FillParty(g, me.ID, "Warrior")
		}
		bear := b12Creature(g, me.ID, "Bear", znrB1BearType, 2, 2)
		castHoldCreature(t, g, "Kabira Outrider", "Creature — Human Warrior", znrB1KabiraOutriderOracle, 3, 3)
		answerPickTarget(t, g, bear)
		passPriorityAroundTable(t, g)
		if p, tough := ptOf(t, g, bear); p != tc.want || tough != tc.want {
			t.Errorf("full=%v: the Bear is %d/%d, want %d/%d", tc.full, p, tough, tc.want, tc.want)
		}
	}
}
